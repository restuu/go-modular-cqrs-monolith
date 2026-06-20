package router

import (
	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/httpx/middleware"
)

func registerInternal(app *fiber.App, deps Deps) {
	g := app.Group("/internal", middleware.InternalSecret(deps.InternalSecret))
	g.Get("/metrics", deps.MetricsHandler)
}
