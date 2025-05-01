package rest

import (
	"context"
	"encoding/base64"
	"github.com/Vyacheslav1557/tester/internal/auth"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

type Handlers struct {
	authUC    auth.UseCase
	jwtSecret string
}

func NewHandlers(authUC auth.UseCase, jwtSecret string) *Handlers {
	return &Handlers{
		authUC:    authUC,
		jwtSecret: jwtSecret,
	}
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

func (h *Handlers) ListSessions(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *Handlers) Terminate(c *fiber.Ctx) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	err = h.authUC.Terminate(ctx, session.UserId)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization", "")
	if authHeader == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	username, pwd, err := parseBasicAuth(authHeader)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	credentials := &models.Credentials{
		Username: strings.ToLower(username),
		Password: pwd,
	}
	device := &models.Device{
		Ip:       c.IP(),
		UseAgent: c.Get("User-Agent", ""),
	}

	ctx := c.Context()

	session, err := h.authUC.Login(ctx, credentials, device)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, models.JWT{
		SessionId: session.Id,
		UserId:    session.UserId,
		Role:      session.Role,
		IssuedAt:  time.Now().Unix(),
	})

	token, err := claims.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Authorization", "Bearer "+token)

	return c.SendStatus(fiber.StatusOK)
}

func (h *Handlers) Logout(c *fiber.Ctx) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	err = h.authUC.Logout(c.Context(), session.Id)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *Handlers) Refresh(c *fiber.Ctx) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	err = h.authUC.Refresh(c.Context(), session.Id)
	if err != nil {
		return c.SendStatus(pkg.ToREST(err))
	}

	return c.SendStatus(fiber.StatusOK)
}

func parseBasicAuth(header string) (string, string, error) {
	const (
		op  = "parseBasicAuth"
		msg = "invalid auth header"
	)

	authParts := strings.Split(header, " ")
	if len(authParts) != 2 || strings.ToLower(authParts[0]) != "basic" {
		return "", "", pkg.Wrap(pkg.ErrUnauthenticated, nil, op, msg)
	}

	decodedAuth, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return "", "", pkg.Wrap(pkg.ErrUnauthenticated, nil, op, msg)
	}

	authParts = strings.Split(string(decodedAuth), ":")
	if len(authParts) != 2 {
		return "", "", pkg.Wrap(pkg.ErrUnauthenticated, nil, op, msg)
	}

	return authParts[0], authParts[1], nil
}
