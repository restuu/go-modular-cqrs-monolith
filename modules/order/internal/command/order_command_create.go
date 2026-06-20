package command

import (
	"context"

	"go-modular-cqrs-monolith/modules/order/api/dto"
)

func (c *Command) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (err error) {
	return nil
}
