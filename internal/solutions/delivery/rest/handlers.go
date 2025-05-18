package rest

import (
	"context"
	"errors"
	"fmt"
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/contests"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"github.com/Vyacheslav1557/tester/internal/runner"
	"github.com/Vyacheslav1557/tester/internal/runner/usecase/tester"
	"github.com/Vyacheslav1557/tester/internal/solutions"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	"io"
	"unicode/utf8"
)

type Handlers struct {
	solutionsUC solutions.UseCase
	problemsUC  problems.UseCase
	runnerUC    runner.UseCase
	contestsUC  contests.UseCase
}

func NewHandlers(solutionsUC solutions.UseCase, runnerUC runner.UseCase, problemsUC problems.UseCase, contestsUC contests.UseCase) *Handlers {
	return &Handlers{
		solutionsUC: solutionsUC,
		runnerUC:    runnerUC,
		problemsUC:  problemsUC,
		contestsUC:  contestsUC,
	}
}

const (
	maxSolutionSize int64 = 10 * 1024 * 1024 // 10 MB
)

const (
	sessionKey = "session"
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
		Penalty:   20, // TODO: get penalty from the fucking contest
	})
	if err != nil {
		return err
	}

	problem, err := h.problemsUC.GetProblemById(ctx, params.ProblemId)
	if err != nil {
		return err
	}

	// FIXME: check if the solutions builds
	// if there are no tests, just accept the solution
	if problem.Meta.Count == 0 {
		err := h.solutionsUC.UpdateSolution(ctx, id, &models.SolutionUpdate{
			State:      models.Accepted,
			Score:      100,
			TimeStat:   0,
			MemoryStat: 0,
		})
		if err != nil {
			return err
		}

		return nil
	}

	zipPath, err := h.problemsUC.DownloadTestsArchive(ctx, problem.Id)
	if err != nil {
		return err
	}

	packet := Packet{
		problemId:   problem.Id,
		solution:    []byte(solution),
		updatedAt:   problem.UpdatedAt.Unix(),
		language:    langName,
		zipPath:     zipPath,
		timeLimit:   int64(problem.TimeLimit),
		memoryLimit: int64(problem.MemoryLimit),
		meta:        &problem.Meta,
	}

	go func() {
		ch := h.runnerUC.Test(ctx, packet)

		solutionUpdate := models.SolutionUpdate{
			State:      models.Saved,
			Score:      0,
			TimeStat:   0,
			MemoryStat: 0,
		}

		testsPassed := 0
		testsPassedExpected := problem.Meta.Count

		for msg := range ch {
			// handle msg details here (when WS is implemented)

			if msg.Err != nil {
				var stErr *tester.StateErr
				if errors.As(msg.Err, &stErr) {
					solutionUpdate.State = stErr.State
					break
				}

				fmt.Println("something really bad happened here")
				break
			}

			if msg.Metrics != nil {
				testsPassed++

				// doing this way we get the max over all tests
				solutionUpdate.MemoryStat = max(
					solutionUpdate.MemoryStat,
					int32(msg.Metrics.MaximumResidentSetSize),
				)
				solutionUpdate.TimeStat = max(
					solutionUpdate.TimeStat,
					int32(msg.Metrics.ElapsedTime.Milliseconds()),
				)
			}
		}

		if testsPassed != testsPassedExpected && solutionUpdate.State == models.Saved {
			fmt.Println("something bad had happened")
			return
		}

		if testsPassed == testsPassedExpected {
			solutionUpdate.Score = 100
			solutionUpdate.State = models.Accepted
		}

		if err := h.solutionsUC.UpdateSolution(ctx, id, &solutionUpdate); err != nil {
			fmt.Println(err)
		}
	}()

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

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		solutionsList, err := h.solutionsUC.ListSolutions(ctx, filter)
		if err != nil {
			return err
		}

		return c.JSON(ListSolutionsResponseDTO(solutionsList))
	case models.RoleStudent:
		if params.ContestId == nil {
			return pkg.Wrap(pkg.ErrBadInput, nil, "", "contest id is required")
		}
		if params.UserId != nil && *params.UserId != session.UserId {
			return pkg.Wrap(pkg.NoPermission, nil, "", "cannot list solutions for another user")
		}

		filter.UserId = &session.UserId

		solutionsList, err := h.solutionsUC.ListSolutions(ctx, filter)
		if err != nil {
			return err
		}

		return c.JSON(ListSolutionsResponseDTO(solutionsList))
	default:
		return pkg.NoPermission
	}
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

func ListSolutionsResponseDTO(solutionsList *models.SolutionsList) *testerv1.ListSolutionsResponse {
	resp := testerv1.ListSolutionsResponse{
		Solutions:  make([]testerv1.SolutionsListItem, len(solutionsList.Solutions)),
		Pagination: PaginationDTO(solutionsList.Pagination),
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

type Packet struct {
	solution    []byte
	problemId   int32
	updatedAt   int64
	language    models.LanguageName
	zipPath     string
	timeLimit   int64
	memoryLimit int64
	meta        *models.Meta
}

func (p Packet) Solution() []byte {
	return p.solution
}

func (p Packet) UniquePacketName() string {
	return fmt.Sprintf("%d_%d", p.problemId, p.updatedAt)
}

func (p Packet) Lang() models.LanguageName {
	return p.language
}

func (p Packet) ZipPath() string {
	return p.zipPath
}

func (p Packet) TL() int64 {
	return p.timeLimit
}

func (p Packet) ML() int64 {
	return p.memoryLimit
}

func (p Packet) Meta() *models.Meta {
	return p.meta
}
