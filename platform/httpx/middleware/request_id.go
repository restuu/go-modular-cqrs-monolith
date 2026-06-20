package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"go-modular-cqrs-monolith/platform/logging"
)

// RequestID reads the inbound X-Request-ID header. If absent or not a valid
// UUID it generates a fresh one. The ID is stored in Fiber locals, injected
// into the request context, and echoed in the X-Request-ID response header.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		var id uuid.UUID

		if raw := c.Get("X-Request-ID"); raw != "" {
			parsed, err := uuid.Parse(raw)
			if err == nil {
				id = parsed
			}
		}

		if id == uuid.Nil {
			id = uuid.New()
		}

		c.Locals(LocalsRequestID, id)
		c.SetContext(logging.WithRequestID(c.Context(), id))
		c.Set("X-Request-ID", id.String())

		return c.Next()
	}
}
