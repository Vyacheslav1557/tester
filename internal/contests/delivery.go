package contests

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/gofiber/fiber/v2"
)

type ContestsHandlers interface {
	CreateParticipant(c *fiber.Ctx, contestId int32, params testerv1.CreateParticipantParams) error
	ListParticipants(c *fiber.Ctx, contestId int32, params testerv1.ListParticipantsParams) error
	DeleteParticipant(c *fiber.Ctx, contestId int32, params testerv1.DeleteParticipantParams) error

	CreateContestProblem(c *fiber.Ctx, contestId int32, params testerv1.CreateContestProblemParams) error
	GetContestProblem(c *fiber.Ctx, contestId int32, problemId int32) error
	DeleteContestProblem(c *fiber.Ctx, contestId int32, problemId int32) error

	CreateContest(c *fiber.Ctx, params testerv1.CreateContestParams) error
	GetContest(c *fiber.Ctx, id int32) error
	ListContests(c *fiber.Ctx, params testerv1.ListContestsParams) error
	UpdateContest(c *fiber.Ctx, id int32) error
	DeleteContest(c *fiber.Ctx, id int32) error

	GetMonitor(c *fiber.Ctx, contestId int32) error
}
