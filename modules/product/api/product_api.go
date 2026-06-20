package api

import (
	"database/sql"

	"go-modular-cqrs-monolith/modules/product/internal/persistence"
	"go-modular-cqrs-monolith/modules/product/internal/query"
)

type ServiceDeps struct {
	DB *sql.DB
}

type Service struct {
	*query.Query
}

func NewService(deps ServiceDeps) *Service {
	productRepo := persistence.NewProductRepositoryImpl(deps.DB)

	return &Service{
		Query: query.NewQuery(productRepo),
	}
}
