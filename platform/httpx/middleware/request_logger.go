package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestLogger returns a Fiber middleware that emits one structured JSON log
// line per request after the handler chain completes.
// Fields logged: method, path, status, duration_ms, request_id, ip, user_agent.
// Level: Info for 2xx/3xx, Warn for 4xx, Error for 5xx.
//
// Must be placed after RequestID in the middleware chain so request_id is
// available in Locals.
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := c.Response().StatusCode()

		var requestID string
		if id, ok := c.Locals(LocalsRequestID).(uuid.UUID); ok {
			requestID = id.String()
		}

		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
			"ip", c.IP(),
			"user_agent", string(c.Request().Header.UserAgent()),
		}

		switch {
		case status >= 500:
			logger.Error("request completed", attrs...)
		case status >= 400:
			logger.Warn("request completed", attrs...)
		default:
			logger.Info("request completed", attrs...)
		}

		return err
	}
}
