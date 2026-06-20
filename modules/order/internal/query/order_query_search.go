package query

import (
	"context"

	"go-modular-cqrs-monolith/modules/order/api/dto"
)

func (q *Query) Search(ctx context.Context, req *dto.OrderSearchRequest) (res dto.OrderSearchResponse, err error) {
	return res, nil
}
