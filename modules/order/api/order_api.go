package api

import (
	"database/sql"

	"go-modular-cqrs-monolith/modules/order/internal/command"
	"go-modular-cqrs-monolith/modules/order/internal/persistence"
	"go-modular-cqrs-monolith/modules/order/internal/port"
	"go-modular-cqrs-monolith/modules/order/internal/query"
	"go-modular-cqrs-monolith/modules/order/internal/transport"

	"github.com/gofiber/fiber/v3"
)

type RouterDeps struct {
	App            *fiber.App
	AuthMiddleware fiber.Handler
	AuthLimiter    fiber.Handler
	PublicLimiter  fiber.Handler
}

type ServiceDeps struct {
	DB             *sql.DB
	ProductService port.ProductService
}

type Service struct {
	*command.Command
	*query.Query
}

func NewService(deps ServiceDeps) *Service {
	orderRepo := persistence.NewOrderRepositoryImpl(deps.DB)

	return &Service{
		Command: command.NewCommand(),
		Query: query.NewQuery(
			orderRepo,
			deps.ProductService,
		),
	}
}

func (s *Service) RegisterRouter(deps RouterDeps) {
	handler := transport.NewOrderHTTPHandler(s.Command, s.Query)
	middlewares := transport.Middlewares{
		Auth:    deps.AuthLimiter,
		Limiter: deps.AuthMiddleware,
	}

	transport.Mount(deps.App, handler, middlewares)
}
