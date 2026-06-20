package query

import "go-modular-cqrs-monolith/modules/product/internal/domain"

type Query struct {
	repo domain.ProductRepository
}

func NewQuery(repo domain.ProductRepository) *Query {
	return &Query{
		repo: repo,
	}
}
