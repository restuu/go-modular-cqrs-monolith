package port

import (
	"context"

	productdto "go-modular-cqrs-monolith/modules/product/api/dto"
)

type ProductService interface {
	GetByModelNumber(ctx context.Context, modelNumber string) (res productdto.ProductDetailResponse, err error)
}
