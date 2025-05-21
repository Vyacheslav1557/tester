package middleware

import (
	"errors"
	"fmt"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/sessions"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v4"
	"strings"
)

func AuthMiddleware(jwtSecret string, sessionsUC sessions.UseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var authHeader string

		if websocket.IsWebSocketUpgrade(c) {
			authHeader = c.Query("token", "")
		} else {
			authHeader = c.Get("Authorization", "")
		}

		if authHeader == "" {
			return c.Next()
		}

		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || strings.ToLower(authParts[0]) != "bearer" {
			return c.Next()
		}

		parsedToken, err := jwt.ParseWithClaims(authParts[1], &models.JWT{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		})
		if err != nil {
			return c.Next()
		}

		token, ok := parsedToken.Claims.(*models.JWT)
		if !ok {
			return c.Next()
		}

		err = token.Valid()
		if err != nil {
			return c.Next()
		}

		ctx := c.Context()

		session, err := sessionsUC.ReadSession(ctx, token.SessionId)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return c.Next()
			}

			return err
		}

		c.Locals("session", session)
		return c.Next()
	}
}
