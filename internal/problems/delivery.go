package problems

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/gofiber/fiber/v2"
)

type ProblemsHandlers interface {
	ListProblems(c *fiber.Ctx, params testerv1.ListProblemsParams) error
	CreateProblem(c *fiber.Ctx) error
	DeleteProblem(c *fiber.Ctx, id int32) error
	GetProblem(c *fiber.Ctx, id int32) error
	UpdateProblem(c *fiber.Ctx, id int32) error
	UploadProblem(c *fiber.Ctx, id int32) error
}
