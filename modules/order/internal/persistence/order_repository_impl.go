package persistence

import (
	"context"
	"database/sql"

	"go-modular-cqrs-monolith/modules/order/internal/domain"
)

type OrderRepositoryImpl struct {
	db *sql.DB
}

func NewOrderRepositoryImpl(db *sql.DB) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{
		db: db,
	}
}

func (o *OrderRepositoryImpl) FindOrderById(ctx context.Context, id int) (res domain.Order, err error) {
	return res, nil
}

func (o *OrderRepositoryImpl) FindOrderItemsByOrderId(ctx context.Context, id int) (res []domain.OrderItem, err error) {
	return res, nil
}
