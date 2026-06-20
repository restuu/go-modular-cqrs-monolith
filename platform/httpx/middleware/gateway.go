package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"go-modular-cqrs-monolith/platform/httpx/response"
)

// GatewayClaims consumes caller identity already parsed and forwarded by an
// upstream API gateway (e.g. Kong) as plain headers, instead of verifying a
// JWT itself. It writes to the exact same Locals keys as Auth, so every
// downstream consumer (GetUserID, AdminOnly, and every domain module's
// handlers) works unchanged regardless of which middleware ran.
//
// Security contract: this middleware trusts the headers outright. The
// gateway MUST be the only ingress to this service and MUST strip any
// client-supplied copies of userIDHeader/roleHeader before forwarding the
// request, or callers could forge their own identity.
func GatewayClaims(userIDHeader, roleHeader string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := uuid.Parse(c.Get(userIDHeader))
		if err != nil {
			return response.Unauthorized(c)
		}

		c.Locals(LocalsUserID, userID)
		c.Locals(LocalsRole, c.Get(roleHeader))
		return c.Next()
	}
}
