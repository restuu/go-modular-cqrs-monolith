package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go-modular-cqrs-monolith/platform/httpx/response"
)

const (
	LocalsUserID    = "user_id"
	LocalsRole      = "user_role"
	LocalsRequestID = "request_id"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Auth(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c)
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Unauthorized(c)
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return response.Unauthorized(c)
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return response.Unauthorized(c)
		}

		c.Locals(LocalsUserID, userID)
		c.Locals(LocalsRole, claims.Role)
		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals(LocalsRole).(string)
		if !ok || role != "admin" {
			return response.Forbidden(c)
		}
		return c.Next()
	}
}

func GetUserID(c fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(LocalsUserID).(uuid.UUID)
	return id, ok
}
