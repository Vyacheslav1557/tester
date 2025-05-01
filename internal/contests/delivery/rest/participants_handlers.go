package rest

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) CreateParticipant(c *fiber.Ctx, params testerv1.CreateParticipantParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		id, err := h.contestsUC.CreateParticipant(ctx, params.ContestId, params.UserId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.JSON(testerv1.CreateParticipantResponse{
			Id: id,
		})
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) UpdateParticipant(c *fiber.Ctx, params testerv1.UpdateParticipantParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		var req testerv1.UpdateParticipantRequest
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		err = h.contestsUC.UpdateParticipant(ctx, params.ParticipantId, models.ParticipantUpdate{
			Name: req.Name,
		})
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}

func (h *Handlers) DeleteParticipant(c *fiber.Ctx, params testerv1.DeleteParticipantParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.contestsUC.DeleteParticipant(c.Context(), params.ParticipantId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))

	}
}

func (h *Handlers) ListParticipants(c *fiber.Ctx, params testerv1.ListParticipantsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		participantsList, err := h.contestsUC.ListParticipants(c.Context(), models.ParticipantsFilter{
			Page:      params.Page,
			PageSize:  params.PageSize,
			ContestId: params.ContestId,
		})
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		resp := testerv1.ListParticipantsResponse{
			Participants: make([]testerv1.ParticipantsListItem, len(participantsList.Participants)),
			Pagination:   PaginationDTO(participantsList.Pagination),
		}

		for i, participant := range participantsList.Participants {
			resp.Participants[i] = ParticipantsListItemDTO(*participant)
		}

		return c.JSON(resp)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}
