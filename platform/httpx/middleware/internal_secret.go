package middleware

import (
	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/httpx/response"
)

// InternalSecret guards /internal/* endpoints with a shared header (X-Internal-Secret).
// When secret is empty (dev default), all requests pass through.
func InternalSecret(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if secret != "" && c.Get("X-Internal-Secret") != secret {
			return response.Unauthorized(c)
		}
		return c.Next()
	}
}
