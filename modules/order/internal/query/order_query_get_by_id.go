package query

import (
	"context"

	"go-modular-cqrs-monolith/modules/order/api/dto"
)

func (q *Query) GetById(ctx context.Context, orderID string) (res dto.OrderDetailResponse, err error) {
	return res, nil
}
