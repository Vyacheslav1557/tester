package solutions

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/gofiber/fiber/v2"
)

type SolutionsHandlers interface {
	CreateSolution(c *fiber.Ctx, params testerv1.CreateSolutionParams) error
	GetSolution(c *fiber.Ctx, id int32) error
	ListSolutions(c *fiber.Ctx, params testerv1.ListSolutionsParams) error
}
