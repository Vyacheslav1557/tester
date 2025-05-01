package rest

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) GetMonitor(c *fiber.Ctx, params testerv1.GetMonitorParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher, models.RoleStudent:
		contest, err := h.contestsUC.GetContest(ctx, params.ContestId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		monitor, err := h.contestsUC.GetMonitor(ctx, params.ContestId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		tasks, err := h.contestsUC.GetTasks(ctx, params.ContestId)
		if err != nil {
			return c.SendStatus(pkg.ToREST(err))
		}

		resp := testerv1.GetMonitorResponse{
			Contest:           ContestDTO(*contest),
			Tasks:             make([]testerv1.TasksListItem, len(tasks)),
			Participants:      make([]testerv1.ParticipantsStat, len(monitor.Participants)),
			SummaryPerProblem: make([]testerv1.ProblemStatSummary, len(monitor.Summary)),
		}

		for i, participant := range monitor.Participants {
			resp.Participants[i] = testerv1.ParticipantsStat{
				Id:             participant.Id,
				Name:           participant.Name,
				PenaltyInTotal: participant.PenaltyInTotal,
				Solutions:      make([]testerv1.SolutionsListItem, len(participant.Solutions)),
				SolvedInTotal:  participant.SolvedInTotal,
			}

			for j, solution := range participant.Solutions {
				resp.Participants[i].Solutions[j] = SolutionsListItemDTO(*solution)
			}
		}

		for i, problem := range monitor.Summary {
			resp.SummaryPerProblem[i] = testerv1.ProblemStatSummary{
				Id:      problem.Id,
				Success: problem.Success,
				Total:   problem.Total,
			}
		}

		for i, task := range tasks {
			resp.Tasks[i] = TasksListItemDTO(*task)
		}
		return c.JSON(resp)
	default:
		return c.SendStatus(pkg.ToREST(pkg.NoPermission))
	}
}
