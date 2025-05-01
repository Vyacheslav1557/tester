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

	jwtSecret string
}

func NewHandlers(problemsUC problems.UseCase, contestsUC contests.UseCase, jwtSecret string) *Handlers {
	return &Handlers{
		problemsUC: problemsUC,
		contestsUC: contestsUC,

		jwtSecret: jwtSecret,
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

func (h *Handlers) CreateContest(c *fiber.Ctx) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		id, err := h.contestsUC.CreateContest(ctx, "Название контеста")
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(&testerv1.CreateContestResponse{
			Id: id,
		})
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) GetContest(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		contest, err := h.contestsUC.GetContest(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		tasks, err := h.contestsUC.GetTasks(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		solutions := make([]*models.SolutionsListItem, 0)
		participantId, err := h.contestsUC.GetParticipantId(ctx, contest.Id, session.UserId)
		if err == nil { // Admin or Teacher may not participate in contest
			solutions, err = h.contestsUC.GetBestSolutions(ctx, id, participantId)
			if err != nil {
				return c.SendStatus(pkg.ToREST(err))
			}
		}

		return c.JSON(GetContestResponseDTO(contest, tasks, solutions))
	case models.RoleStudent:
		contest, err := h.contestsUC.GetContest(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		tasks, err := h.contestsUC.GetTasks(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		participantId, err := h.contestsUC.GetParticipantId(ctx, contest.Id, session.UserId)
		solutions, err := h.contestsUC.GetBestSolutions(c.Context(), id, participantId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(GetContestResponseDTO(contest, tasks, solutions))
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) UpdateContest(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
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
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) DeleteContest(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.DeleteContest(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) ListContests(c *fiber.Ctx, params testerv1.ListContestsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	filter := models.ContestsFilter{
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		contestsList, err := h.contestsUC.ListContests(ctx, filter)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(ListContestsResponseDTO(contestsList))
	case models.RoleStudent:
		filter.UserId = &session.UserId
		contestsList, err := h.contestsUC.ListContests(ctx, filter)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(ListContestsResponseDTO(contestsList))
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}
