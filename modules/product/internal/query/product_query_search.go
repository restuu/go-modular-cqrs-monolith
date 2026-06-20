package query

import (
	"context"

	"go-modular-cqrs-monolith/modules/product/api/dto"
)

func (q *Query) Search(ctx context.Context, req dto.ProductSearchRequest) (res dto.ProductSearchResponse, err error) {
	return res, nil
}
