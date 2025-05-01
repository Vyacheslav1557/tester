package rest

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) CreateTask(c *fiber.Ctx, params testerv1.CreateTaskParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		id, err := h.contestsUC.CreateTask(ctx, params.ContestId, params.ProblemId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(testerv1.CreateTaskResponse{
			Id: id,
		})
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) GetTask(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		contest, err := h.contestsUC.GetContest(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		tasks, err := h.contestsUC.GetTasks(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		t, err := h.contestsUC.GetTask(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(GetTaskResponseDTO(contest, tasks, t))
	case models.RoleStudent:
		_, err = h.contestsUC.GetParticipantId2(ctx, id, session.UserId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(pkg.NoPermission))
		}

		contest, err := h.contestsUC.GetContest(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		tasks, err := h.contestsUC.GetTasks(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		t, err := h.contestsUC.GetTask(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(GetTaskResponseDTO(contest, tasks, t))
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) DeleteTask(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.DeleteTask(c.Context(), id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}
