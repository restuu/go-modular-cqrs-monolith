package query

import (
	"context"
	"log/slog"

	"go-modular-cqrs-monolith/modules/product/api/dto"
	"go-modular-cqrs-monolith/platform/logging/attr"
)

func (q *Query) GetByModelNumber(ctx context.Context, modelNumber string) (res dto.ProductDetailResponse, err error) {
	product, err := q.repo.FindProductByModelNumber(ctx, modelNumber)
	if err != nil {
		slog.ErrorContext(ctx, "Query.GetByModelNumber() - repo.FindProductByModelNumber() failed",
			attr.Err(err))
		return res, err
	}

	res = dto.ProductDetailResponse{
		ModelNumber: product.ModelNumber,
		Name:        product.Name,
		Price:       product.Price,
	}

	return res, nil
}
