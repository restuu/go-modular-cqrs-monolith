package query

import (
	"go-modular-cqrs-monolith/modules/order/internal/domain"
	"go-modular-cqrs-monolith/modules/order/internal/port"
)

type Query struct {
	orderRepo  domain.OrderRepository
	productSvc port.ProductCatalog
}

func NewQuery(
	orderRepo domain.OrderRepository,
	productSvc port.ProductCatalog,
) *Query {
	return &Query{
		orderRepo:  orderRepo,
		productSvc: productSvc,
	}
}
