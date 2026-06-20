package router

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// newAuthLimiter rate-limits auth endpoints to deter brute force (10 req/min/IP).
func newAuthLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
	})
}

// newPublicLimiter caps public API at 100 req/min/IP.
func newPublicLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
	})
}
