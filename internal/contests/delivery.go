package contests

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/gofiber/fiber/v2"
)

type ContestsHandlers interface {
	ListContests(c *fiber.Ctx, params testerv1.ListContestsParams) error
	CreateContest(c *fiber.Ctx) error
	DeleteContest(c *fiber.Ctx, id int32) error
	GetContest(c *fiber.Ctx, id int32) error
	UpdateContest(c *fiber.Ctx, id int32) error
	DeleteParticipant(c *fiber.Ctx, params testerv1.DeleteParticipantParams) error
	ListParticipants(c *fiber.Ctx, params testerv1.ListParticipantsParams) error
	UpdateParticipant(c *fiber.Ctx, params testerv1.UpdateParticipantParams) error
	CreateParticipant(c *fiber.Ctx, params testerv1.CreateParticipantParams) error
	ListSolutions(c *fiber.Ctx, params testerv1.ListSolutionsParams) error
	CreateSolution(c *fiber.Ctx, params testerv1.CreateSolutionParams) error
	GetSolution(c *fiber.Ctx, id int32) error
	DeleteTask(c *fiber.Ctx, id int32) error
	CreateTask(c *fiber.Ctx, params testerv1.CreateTaskParams) error
	GetMonitor(c *fiber.Ctx, params testerv1.GetMonitorParams) error
	GetTask(c *fiber.Ctx, id int32) error
}
