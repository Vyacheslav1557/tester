package rest

import (
	"context"
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	problemsUC problems.UseCase

	jwtSecret string
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

func NewHandlers(problemsUC problems.UseCase) *Handlers {
	return &Handlers{
		problemsUC: problemsUC,
	}
}

func (h *Handlers) ListProblems(c *fiber.Ctx, params testerv1.ListProblemsParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		problemsList, err := h.problemsUC.ListProblems(c.Context(), models.ProblemsFilter{
			Page:     params.Page,
			PageSize: params.PageSize,
			Title:    params.Title,
			Order:    params.Order,
		})
		if err != nil {
			return err
		}

		resp := testerv1.ListProblemsResponse{
			Problems:   make([]testerv1.ProblemsListItem, len(problemsList.Problems)),
			Pagination: PaginationDTO(problemsList.Pagination),
		}

		for i, problem := range problemsList.Problems {
			resp.Problems[i] = ProblemsListItemDTO(*problem)
		}
		return c.JSON(resp)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) CreateProblem(c *fiber.Ctx, params testerv1.CreateProblemParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		if params.Title == "" {
			return pkg.Wrap(pkg.ErrBadInput, nil, "CreateProblem", "empty title")
		}

		id, err := h.problemsUC.CreateProblem(c.Context(), params.Title)
		if err != nil {
			return err
		}

		return c.JSON(testerv1.CreationResponse{
			Id: id,
		})
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) DeleteProblem(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		err := h.problemsUC.DeleteProblem(c.Context(), id)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) GetProblem(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		problem, err := h.problemsUC.GetProblemById(c.Context(), id)
		if err != nil {
			return err
		}

		return c.JSON(
			testerv1.GetProblemResponse{Problem: *ProblemDTO(problem)},
		)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) UpdateProblem(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		var req testerv1.UpdateProblemRequest
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		err = h.problemsUC.UpdateProblem(c.Context(), id, &models.ProblemUpdate{
			Title:       req.Title,
			MemoryLimit: req.MemoryLimit,
			TimeLimit:   req.TimeLimit,

			Legend:       req.Legend,
			InputFormat:  req.InputFormat,
			OutputFormat: req.OutputFormat,
			Notes:        req.Notes,
			Scoring:      req.Scoring,
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) UploadProblem(c *fiber.Ctx, id int32) error {
	const op = "ProblemsHandlers.UploadProblem"

	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		a, err := c.FormFile("archive")
		if err != nil {
			return pkg.Wrap(pkg.ErrBadInput, err, op, "no archive uploaded")
		}

		if a.Size == 0 || a.Size > 1024*1024*1024 {
			return pkg.Wrap(pkg.ErrBadInput, err, op, "invalid archive size")
		}

		f, err := a.Open()
		if err != nil {
			return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to open archive")
		}
		defer f.Close()

		err = h.problemsUC.UploadProblem(ctx, id, f, a.Size)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func ProblemsListItemDTO(p models.ProblemsListItem) testerv1.ProblemsListItem {
	return testerv1.ProblemsListItem{
		Id:          p.Id,
		Title:       p.Title,
		MemoryLimit: p.MemoryLimit,
		TimeLimit:   p.TimeLimit,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ProblemDTO(p *models.Problem) *testerv1.Problem {
	return &testerv1.Problem{
		Id:          p.Id,
		Title:       p.Title,
		TimeLimit:   p.TimeLimit,
		MemoryLimit: p.MemoryLimit,

		Legend:       p.Legend,
		InputFormat:  p.InputFormat,
		OutputFormat: p.OutputFormat,
		Notes:        p.Notes,
		Scoring:      p.Scoring,

		LegendHtml:       p.LegendHtml,
		InputFormatHtml:  p.InputFormatHtml,
		OutputFormatHtml: p.OutputFormatHtml,
		NotesHtml:        p.NotesHtml,
		ScoringHtml:      p.ScoringHtml,

		//Meta:    MetaDTO(p.Meta),
		//Samples: SamplesDTO(p.Samples),

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

//func MetaDTO(m models.Meta) testerv1.Meta {
//	return testerv1.Meta{
//		Author: m.Author,
//	}
//}
