package domain

import "context"

type OrderRepository interface {
	FindOrderById(ctx context.Context, id int) (Order, error)
	FindOrderItemsByOrderId(ctx context.Context, id int) ([]OrderItem, error)
}
