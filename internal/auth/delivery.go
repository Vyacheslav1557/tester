package auth

import (
	"github.com/gofiber/fiber/v2"
)

type AuthHandlers interface {
	ListSessions(c *fiber.Ctx) error
	Terminate(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
}
