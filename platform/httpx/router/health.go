package router

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/httpx/response"
)

const healthCheckTimeout = 3 * time.Second

// registerHealth registers the liveness/readiness probe route.
// Intentionally registered before the request logger and metrics middleware so
// high-frequency scrapes from Docker and orchestrators do not pollute request
// logs or inflate Prometheus counters.
func registerHealth(app *fiber.App, probes map[string]func(ctx context.Context) error) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), healthCheckTimeout)
		defer cancel()

		checks := make(fiber.Map, len(probes))
		healthy := true
		for name, probe := range probes {
			if err := probe(ctx); err != nil {
				checks[name] = fmt.Sprintf("error: %s", err.Error())
				healthy = false
			} else {
				checks[name] = "ok"
			}
		}

		status := "ok"
		if !healthy {
			status = "degraded"
		}
		data := fiber.Map{"status": status, "checks": checks}

		if !healthy {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{"success": false, "data": data, "error": nil, "meta": nil},
			)
		}
		return response.OK(c, data)
	})
}
