package api

import (
	"database/sql"

	"go-modular-cqrs-monolith/modules/order/internal/adapter"
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
	ProductService port.ProductAPI
}

type Service struct {
	*command.Command
	*query.Query
}

func NewService(deps ServiceDeps) *Service {
	orderRepo := persistence.NewOrderRepositoryImpl(deps.DB)
	orderProductAdapter := adapter.NewOrderProductAdapter(deps.ProductService)

	return &Service{
		Command: command.NewCommand(),
		Query: query.NewQuery(
			orderRepo,
			orderProductAdapter,
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
