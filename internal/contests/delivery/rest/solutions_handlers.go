package rest

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	"io"
)

const (
	maxSolutionSize int64 = 10 * 1024 * 1024
)

func (h *Handlers) CreateSolution(c *fiber.Ctx, params testerv1.CreateSolutionParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher, models.RoleStudent:
		s, err := c.FormFile("solution")
		if err != nil {
			return err
		}

		if s.Size == 0 || s.Size > maxSolutionSize {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		f, err := s.Open()
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		id, err := h.contestsUC.CreateSolution(ctx, &models.SolutionCreation{
			UserId:   session.UserId,
			TaskId:   params.TaskId,
			Language: params.Language,
			Solution: string(b),
		})
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(testerv1.CreateSolutionResponse{
			Id: id,
		})
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) GetSolution(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		solution, err := h.contestsUC.GetSolution(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(testerv1.GetSolutionResponse{Solution: SolutionDTO(*solution)})
	case models.RoleStudent:
		_, err := h.contestsUC.GetParticipantId3(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		solution, err := h.contestsUC.GetSolution(ctx, id)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(testerv1.GetSolutionResponse{Solution: SolutionDTO(*solution)})
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) ListSolutions(c *fiber.Ctx, params testerv1.ListSolutionsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	filter := models.SolutionsFilter{
		ContestId:     params.ContestId,
		Page:          params.Page,
		PageSize:      params.PageSize,
		ParticipantId: params.ParticipantId,
		TaskId:        params.TaskId,
		Language:      params.Language,
		Order:         params.Order,
		State:         params.State,
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		solutionsList, err := h.contestsUC.ListSolutions(ctx, filter)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(ListSolutionsResponseDTO(solutionsList))
	case models.RoleStudent:
		if params.ContestId == nil {
			return c.SendStatus(pkg.ToREST(pkg.NoPermission))
		}

		participantId, err := h.contestsUC.GetParticipantId(ctx, *params.ContestId, session.UserId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(pkg.NoPermission))
		}

		// Student cannot view other users' solutions
		if params.ParticipantId != nil && *params.ParticipantId != participantId {
			return c.SendStatus(pkg.ToREST(pkg.NoPermission))
		}

		filter.ParticipantId = &participantId
		solutionsList, err := h.contestsUC.ListSolutions(ctx, filter)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(ListSolutionsResponseDTO(solutionsList))
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}
