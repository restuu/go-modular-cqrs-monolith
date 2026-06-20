package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/metrics"
)

// RequestMetrics returns a Fiber middleware that records each request in the Prometheus
// recorder (counter + latency histogram). Must be placed after RequestLogger in the chain
// so the route pattern is resolved before we read it.
func RequestMetrics(rec *metrics.Recorder) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		route := c.Route().Path
		// Fiber's default 404 handler registers at "/". When no route matches,
		// Fiber falls through to it regardless of the actual request path.
		// Label these as "unknown" to prevent cardinality explosion from arbitrary paths.
		if route == "" || (route == "/" && c.Path() != "/") {
			route = "unknown"
		}

		rec.ObserveRequest(c.Method(), route, c.Response().StatusCode(), time.Since(start))
		return err
	}
}
