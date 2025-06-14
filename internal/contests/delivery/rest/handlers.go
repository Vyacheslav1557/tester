package rest

import (
	"context"
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/contests"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	problemsUC problems.UseCase
	contestsUC contests.UseCase
}

func NewHandlers(problemsUC problems.UseCase, contestsUC contests.UseCase) *Handlers {
	return &Handlers{
		problemsUC: problemsUC,
		contestsUC: contestsUC,
	}
}

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

func (h *Handlers) CreateContest(c *fiber.Ctx, params testerv1.CreateContestParams) error {
	const op = "ContestsHandlers.CreateContest"

	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		if params.Title == "" {
			return pkg.Wrap(pkg.ErrBadInput, nil, op, "empty title")
		}

		id, err := h.contestsUC.CreateContest(ctx, params.Title)
		if err != nil {
			return err
		}

		return c.JSON(&testerv1.CreationResponse{Id: id})
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) GetContest(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		contest, err := h.contestsUC.GetContest(ctx, id)
		if err != nil {
			return err
		}

		ps, err := h.contestsUC.GetContestProblems(ctx, id)
		if err != nil {
			return err
		}

		return c.JSON(GetContestResponseDTO(contest, ps))
	case models.RoleStudent:
		isParticipant, err := h.contestsUC.IsParticipant(ctx, id, session.UserId)
		if err != nil {
			return err
		}

		if !isParticipant {
			return pkg.NoPermission
		}

		contest, err := h.contestsUC.GetContest(ctx, id)
		if err != nil {
			return err
		}

		ps, err := h.contestsUC.GetContestProblems(ctx, id)
		if err != nil {
			return err
		}

		return c.JSON(GetContestResponseDTO(contest, ps))
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) UpdateContest(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		var req testerv1.UpdateContestRequest
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		err = h.contestsUC.UpdateContest(ctx, id, models.ContestUpdate{
			Title: req.Title,
		})
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) DeleteContest(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.DeleteContest(ctx, id)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) ListContests(c *fiber.Ctx, params testerv1.ListContestsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	filter := models.ContestsFilter{
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		contestsList, err := h.contestsUC.ListContests(ctx, filter)
		if err != nil {
			return err
		}

		return c.JSON(ListContestsResponseDTO(contestsList))
	case models.RoleStudent:
		filter.UserId = &session.UserId
		contestsList, err := h.contestsUC.ListContests(ctx, filter)
		if err != nil {
			return err
		}

		return c.JSON(ListContestsResponseDTO(contestsList))
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) CreateContestProblem(c *fiber.Ctx, contestId int32, params testerv1.CreateContestProblemParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.CreateContestProblem(ctx, contestId, params.ProblemId)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) GetContestProblem(c *fiber.Ctx, contestId int32, problemId int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		p, err := h.contestsUC.GetContestProblem(ctx, contestId, problemId)
		if err != nil {
			return err
		}

		return c.JSON(GetContestProblemResponseDTO(p))
	case models.RoleStudent:
		isParticipant, err := h.contestsUC.IsParticipant(ctx, contestId, session.UserId)
		if err != nil {
			return err
		}

		if !isParticipant {
			return pkg.NoPermission
		}

		p, err := h.contestsUC.GetContestProblem(ctx, contestId, problemId)
		if err != nil {
			return err
		}

		return c.JSON(GetContestProblemResponseDTO(p))
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) DeleteContestProblem(c *fiber.Ctx, contestId int32, problemId int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.DeleteContestProblem(c.Context(), contestId, problemId)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) CreateParticipant(c *fiber.Ctx, contestId int32, params testerv1.CreateParticipantParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.CreateParticipant(ctx, contestId, params.UserId)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) DeleteParticipant(c *fiber.Ctx, contestId int32, params testerv1.DeleteParticipantParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.DeleteParticipant(ctx, contestId, params.UserId)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) ListParticipants(c *fiber.Ctx, contestId int32, params testerv1.ListParticipantsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		participantsList, err := h.contestsUC.ListParticipants(ctx, models.ParticipantsFilter{
			Page:      params.Page,
			PageSize:  params.PageSize,
			ContestId: contestId,
		})
		if err != nil {
			return err
		}

		resp := testerv1.ListUsersResponse{
			Users:      make([]testerv1.User, len(participantsList.Users)),
			Pagination: PaginationDTO(participantsList.Pagination),
		}

		for i, user := range participantsList.Users {
			resp.Users[i] = UserDTO(*user)
		}

		return c.JSON(resp)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) GetMonitor(c *fiber.Ctx, contestId int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher, models.RoleStudent:
		monitor, err := h.contestsUC.GetMonitor(ctx, contestId)
		if err != nil {
			return err
		}
		return c.JSON(GetMonitorResponseDTO(monitor))
	default:
		return pkg.NoPermission
	}
}

func GetContestResponseDTO(contest *models.Contest, problems []*models.ContestProblemsListItem) *testerv1.GetContestResponse {
	resp := testerv1.GetContestResponse{
		Contest:  ContestDTO(*contest),
		Problems: make([]testerv1.ContestProblemListItem, len(problems)),
	}

	for i, task := range problems {
		resp.Problems[i] = ContestProblemsListItemDTO(*task)
	}

	return &resp
}

func ListContestsResponseDTO(contestsList *models.ContestsList) *testerv1.ListContestsResponse {
	resp := testerv1.ListContestsResponse{
		Contests:   make([]testerv1.Contest, len(contestsList.Contests)),
		Pagination: PaginationDTO(contestsList.Pagination),
	}

	for i, contest := range contestsList.Contests {
		resp.Contests[i] = ContestDTO(*contest)
	}

	return &resp
}

func GetContestProblemResponseDTO(p *models.ContestProblem) *testerv1.GetContestProblemResponse {
	resp := testerv1.GetContestProblemResponse{
		Problem: testerv1.ContestProblem{
			ProblemId:   p.ProblemId,
			Title:       p.Title,
			MemoryLimit: p.MemoryLimit,
			TimeLimit:   p.TimeLimit,

			Position: p.Position,

			LegendHtml:       p.LegendHtml,
			InputFormatHtml:  p.InputFormatHtml,
			OutputFormatHtml: p.OutputFormatHtml,
			NotesHtml:        p.NotesHtml,
			ScoringHtml:      p.ScoringHtml,

			//Meta:             MetaDTO(p.Meta),
			//Samples:          SamplesDTO(p.Samples),

			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
	}

	return &resp
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func ContestDTO(c models.Contest) testerv1.Contest {
	return testerv1.Contest{
		Id:        c.Id,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ContestProblemsListItemDTO(t models.ContestProblemsListItem) testerv1.ContestProblemListItem {
	return testerv1.ContestProblemListItem{
		ProblemId:   t.ProblemId,
		Position:    t.Position,
		Title:       t.Title,
		MemoryLimit: t.MemoryLimit,
		TimeLimit:   t.TimeLimit,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func UserDTO(u models.User) testerv1.User {
	return testerv1.User{
		Id:        u.Id,
		Username:  u.Username,
		Role:      int32(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func GetMonitorResponseDTO(m *models.Monitor) testerv1.GetMonitorResponse {
	resp := testerv1.GetMonitorResponse{
		Participants: make([]testerv1.ParticipantsStat, len(m.Participants)),
		Summary:      make([]testerv1.ProblemStatSummary, len(m.Summary)),
	}

	ProblemAttemptsDTO := func(p *models.ProblemAttempts) testerv1.ProblemAttempts {
		return testerv1.ProblemAttempts{
			ProblemId:      p.ProblemId,
			Position:       p.Position,
			State:          stateP(p.State),
			FailedAttempts: p.FAttempts,
		}
	}

	ParticipantsStatDTO := func(p models.ParticipantsStat) testerv1.ParticipantsStat {
		s := testerv1.ParticipantsStat{
			UserId:   p.UserId,
			Username: p.Username,
			Solved:   p.Solved,
			Penalty:  p.Penalty,
			Attempts: make([]testerv1.ProblemAttempts, len(p.Attempts)),
		}

		for i, attempt := range p.Attempts {
			s.Attempts[i] = ProblemAttemptsDTO(attempt)
		}

		return s
	}

	ProblemStatSummaryDTO := func(p models.ProblemStatSummary) testerv1.ProblemStatSummary {
		return testerv1.ProblemStatSummary{
			ProblemId: p.ProblemId,
			Position:  p.Position,
			SAttempts: p.SAttempts,
			FAttempts: p.UnsAttempts,
			TAttempts: p.TAttempts,
		}
	}

	for i, user := range m.Participants {
		resp.Participants[i] = ParticipantsStatDTO(*user)
	}

	for i, summary := range m.Summary {
		resp.Summary[i] = ProblemStatSummaryDTO(*summary)
	}

	return resp
}

func stateP(s *models.State) *int32 {
	if s == nil {
		return nil
	}
	return (*int32)(s)
}
