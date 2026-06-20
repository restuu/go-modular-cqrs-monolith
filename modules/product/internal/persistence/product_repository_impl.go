package persistence

import (
	"context"
	"database/sql"

	"go-modular-cqrs-monolith/modules/product/internal/domain"
)

type ProductRepositoryImpl struct {
	db *sql.DB
}

func NewProductRepositoryImpl(db *sql.DB) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{
		db: db,
	}
}

func (p *ProductRepositoryImpl) FindProductByModelNumber(ctx context.Context, modelNumber string) (
	res domain.Product,
	err error,
) {
	return res, nil
}
