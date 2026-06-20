package metrics

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns a Fiber handler that serves the Prometheus text exposition format
// for the given registry.
func Handler(reg *prometheus.Registry) fiber.Handler {
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	return adaptor.HTTPHandler(h)
}
