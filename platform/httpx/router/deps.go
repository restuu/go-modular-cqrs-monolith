package router

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"go-modular-cqrs-monolith/platform/metrics"
)

// RouteRegistrar is implemented by modules that self-register their HTTP routes.
// The router calls each registrar after constructing the rate-limiters so
// modules can apply them without re-creating the limiters themselves.
// publicLimiter caps general API traffic; authLimiter is stricter for auth routes;
// authMW is the identity strategy chosen once in main.go (local JWT or gateway claims).
type RouteRegistrar func(app *fiber.App, publicLimiter fiber.Handler, authLimiter fiber.Handler, authMW fiber.Handler)

type Deps struct {
	Logger *slog.Logger

	// AuthMiddleware establishes caller identity for protected routes. It is
	// chosen once in main.go based on AUTH_MODE (verify a local JWT, or trust
	// claims forwarded by an upstream gateway) and injected into every module
	// unchanged — modules never know which strategy is active.
	AuthMiddleware fiber.Handler
	InternalSecret string

	Recorder       *metrics.Recorder
	MetricsHandler fiber.Handler

	// HealthProbes are called by GET /healthz with a short deadline.
	// Keys are dependency names (e.g. "postgres", "redis"); values return nil if healthy.
	// A nil map is valid — the endpoint still serves but reports no dependency checks.
	HealthProbes map[string]func(ctx context.Context) error

	// Modules contains self-registering route modules (e.g. bookmarkapi.Service.RegisterRoutes).
	// Each registrar is called after the standard routes with the shared limiters.
	Modules []RouteRegistrar
}
