// Package router builds the Fiber app and registers all HTTP routes.
// Each route group has its own file in this package; this file orchestrates them.
package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"go-modular-cqrs-monolith/platform/httpx/middleware"
	"go-modular-cqrs-monolith/platform/httpx/response"
)

// New constructs the Fiber app with all routes registered.
func New(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return response.InternalServerError(c)
		},
	})

	// registerHealth is called before any middleware so that liveness probes
	// bypass request logging and Prometheus counters (noise reduction).
	// recover.New() is still the outermost wrapper so panics are always caught.
	app.Use(recover.New())
	registerHealth(app, deps.HealthProbes)
	setupCors(app)

	// Middleware order is load-bearing:
	// RequestID must precede RequestLogger so the ID is in Locals when we log.
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(deps.Logger))
	app.Use(middleware.RequestMetrics(deps.Recorder))

	authLimiter := newAuthLimiter()
	publicLimiter := newPublicLimiter()

	registerInternal(app, deps)

	// Self-registering modules mount their own routes here.
	// Both limiters are passed so modules can choose the appropriate one.
	for _, reg := range deps.Modules {
		reg(app, publicLimiter, authLimiter, deps.AuthMiddleware)
	}

	return app
}
