package rest

import (
	"context"
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/contests"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"github.com/Vyacheslav1557/tester/internal/solutions"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"unicode/utf8"
)

type Handlers struct {
	solutionsUC solutions.UseCase
	problemsUC  problems.UseCase
	contestsUC  contests.UseCase
}

func NewHandlers(
	solutionsUC solutions.UseCase,
	problemsUC problems.UseCase,
	contestsUC contests.UseCase,
) *Handlers {
	handlers := &Handlers{
		solutionsUC: solutionsUC,
		problemsUC:  problemsUC,
		contestsUC:  contestsUC,
	}

	return handlers
}

const (
	maxSolutionSize int64 = 10 * 1024 * 1024 // 10 MB
	sessionKey            = "session"
)

func sessionFromCtx(ctx context.Context) (*models.Session, error) {
	const op = "sessionFromCtx"

	session, ok := ctx.Value(sessionKey).(*models.Session)
	if !ok {
		return nil, pkg.Wrap(pkg.ErrUnauthenticated, nil, op, "")
	}

	return session, nil
}

func (h *Handlers) CreateSolution(c *fiber.Ctx, params testerv1.CreateSolutionParams) error {
	const op = "SolutionsHandlers.CreateSolution"

	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		break
	case models.RoleStudent:
		isParticipant, err := h.contestsUC.IsParticipant(ctx, params.ContestId, session.UserId)
		if err != nil {
			return err
		}

		if !isParticipant {
			return pkg.NoPermission
		}

		break
	default:
		return pkg.NoPermission
	}

	s, err := c.FormFile("solution")
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to get solution")
	}

	if s.Size == 0 || s.Size > maxSolutionSize {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid solution size")
	}

	f, err := s.Open()
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to open solution")
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to read solution")
	}

	if len(b) == 0 {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "solution is empty")
	}

	solution := string(b)

	if !utf8.ValidString(solution) {
		return pkg.Wrap(pkg.ErrBadInput, err, op, "solution is not valid utf-8")
	}

	// check if language is valid
	langName := models.LanguageName(params.Language)
	if err := langName.Valid(); err != nil {
		return err
	}

	id, err := h.solutionsUC.CreateSolution(ctx, &models.SolutionCreation{
		UserId:    session.UserId,
		ProblemId: params.ProblemId,
		ContestId: params.ContestId,
		Language:  langName,
		Solution:  solution,
		Penalty:   20, // TODO: get penalty from contest
	})
	if err != nil {
		return err
	}

	return c.JSON(testerv1.CreationResponse{Id: id})
}

func (h *Handlers) GetSolution(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		solution, err := h.solutionsUC.GetSolution(ctx, id)
		if err != nil {
			return err
		}

		return c.JSON(testerv1.GetSolutionResponse{Solution: SolutionDTO(*solution)})
	case models.RoleStudent:
		/*
			Probably this is not a good idea
			to check if the solution belongs to the user
			after getting it

			But it is simple and ok for now
		*/
		solution, err := h.solutionsUC.GetSolution(ctx, id)

		// check if the solution belongs to the user
		if err == nil && solution.UserId != session.UserId {
			return pkg.NoPermission
		}

		if err != nil {
			return err
		}

		return c.JSON(testerv1.GetSolutionResponse{Solution: SolutionDTO(*solution)})
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) ListSolutions(c *fiber.Ctx, params testerv1.ListSolutionsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	filter := ListSolutionsParamsDTO(params)

	var solutionsList *models.SolutionsList

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		if params.ContestId == nil {
			return pkg.Wrap(pkg.ErrBadInput, nil, "", "contest id is required")
		}

		solutionsList, err = h.solutionsUC.ListSolutions(ctx, filter)
		if err != nil {
			return err
		}

		at, err := NewJWT(session.UserId, *params.ContestId, session.Role)
		if err != nil {
			return err
		}

		return c.JSON(ListSolutionsResponseDTO(solutionsList, at))
	case models.RoleStudent:
		if params.ContestId == nil {
			return pkg.Wrap(pkg.ErrBadInput, nil, "", "contest id is required")
		}
		if params.UserId == nil || *params.UserId != session.UserId {
			return pkg.Wrap(pkg.NoPermission, nil, "", "cannot list solutions for another user")
		}

		solutionsList, err = h.solutionsUC.ListSolutions(ctx, filter)
		if err != nil {
			return err
		}

		at, err := NewJWT(session.UserId, *params.ContestId, session.Role)
		if err != nil {
			return err
		}

		return c.JSON(ListSolutionsResponseDTO(solutionsList, at))
	default:
		return pkg.NoPermission
	}
}

type CustomClaims struct {
	UserId    int32       `json:"UserId"`
	ContestId int32       `json:"ContestId"`
	Role      models.Role `json:"Role"`
	jwt.RegisteredClaims
}

func NewJWT(userId int32, contestId int32, role models.Role) (string, error) {
	claims := CustomClaims{
		UserId:    userId,
		ContestId: contestId,
		Role:      role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key"))
}

func ListSolutionsParamsDTO(params testerv1.ListSolutionsParams) models.SolutionsFilter {
	var langName *models.LanguageName = nil
	if params.Language != nil {
		t := models.LanguageName(*params.Language)
		langName = &t
	}

	var state *models.State = nil
	if params.State != nil {
		t := models.State(*params.State)
		state = &t
	}

	return models.SolutionsFilter{
		ContestId: params.ContestId,
		Page:      params.Page,
		PageSize:  params.PageSize,
		UserId:    params.UserId,
		ProblemId: params.ProblemId,
		Language:  langName,
		Order:     params.Order,
		State:     state,
	}
}

func ListSolutionsResponseDTO(solutionsList *models.SolutionsList, at string) *testerv1.ListSolutionsResponse {
	resp := testerv1.ListSolutionsResponse{
		Solutions:   make([]testerv1.SolutionsListItem, len(solutionsList.Solutions)),
		Pagination:  PaginationDTO(solutionsList.Pagination),
		AccessToken: at,
	}

	for i, solution := range solutionsList.Solutions {
		resp.Solutions[i] = SolutionsListItemDTO(*solution)
	}

	return &resp
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func SolutionsListItemDTO(s models.SolutionsListItem) testerv1.SolutionsListItem {
	return testerv1.SolutionsListItem{
		Id: s.Id,

		UserId:   s.UserId,
		Username: s.Username,

		State:      int32(s.State),
		Score:      s.Score,
		Penalty:    s.Penalty,
		TimeStat:   s.TimeStat,
		MemoryStat: s.MemoryStat,
		Language:   int32(s.Language),

		ProblemId:    s.ProblemId,
		ProblemTitle: s.ProblemTitle,

		Position: s.Position,

		ContestId:    s.ContestId,
		ContestTitle: s.ContestTitle,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func SolutionDTO(s models.Solution) testerv1.Solution {
	return testerv1.Solution{
		Id: s.Id,

		UserId:   s.UserId,
		Username: s.Username,

		Solution: s.Solution,

		State:      int32(s.State),
		Score:      s.Score,
		Penalty:    s.Penalty,
		TimeStat:   s.TimeStat,
		MemoryStat: s.MemoryStat,
		Language:   int32(s.Language),

		ProblemId:    s.ProblemId,
		ProblemTitle: s.ProblemTitle,

		Position: s.Position,

		ContestId:    s.ContestId,
		ContestTitle: s.ContestTitle,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
