package users

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/gofiber/fiber/v2"
)

type UsersHandlers interface {
	ListUsers(c *fiber.Ctx, params testerv1.ListUsersParams) error
	CreateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx, id int32) error
	GetUser(c *fiber.Ctx, id int32) error
	UpdateUser(c *fiber.Ctx, id int32) error
}
