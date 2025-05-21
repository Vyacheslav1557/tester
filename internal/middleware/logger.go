package middleware

import (
	"errors"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"net/http"
)

type errorResponse struct {
	Err string `json:"error"`
	Msg string `json:"message"`
}

// ErrorHandlerMiddleware handles errors, maps them to HTTP status codes and logs them
func ErrorHandlerMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		statusCode := c.Response().StatusCode()

		var cErr *pkg.CustomError
		if errors.As(err, &cErr) {
			statusCode = pkg.ToREST(err)
		}

		resp := errorResponse{
			Err: http.StatusText(statusCode),
			Msg: "",
		}

		var fErr *fiber.Error
		if errors.As(err, &fErr) {
			statusCode = fErr.Code
			resp.Err = http.StatusText(statusCode)
			resp.Msg = fErr.Message
		}

		logFields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", statusCode),
			zap.String("error", err.Error()),
		}

		if cErr != nil {
			resp.Msg = cErr.Message

			logFields = append(logFields,
				zap.NamedError("basic", cErr.Basic),
				zap.NamedError("cause", cErr.Cause),
				zap.String("operation", cErr.Op),
				zap.String("message", cErr.Message),
			)
		}

		switch statusCode {
		case http.StatusInternalServerError:
			logger.Error("Internal server error", logFields...)
		case http.StatusBadRequest:
			logger.Warn("Bad request", logFields...)
		case http.StatusUnauthorized, http.StatusForbidden:
			logger.Info("Authentication/Authorization error", logFields...)
		case http.StatusNotFound:
			logger.Info("Resource not found", logFields...)
		default:
			logger.Error("Unhandled error", logFields...)
		}

		return c.Status(statusCode).JSON(resp)
	}
}
