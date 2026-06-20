package port

import (
	"context"

	productdto "go-modular-cqrs-monolith/modules/product/api/dto"
)

// ProductCatalog is the inbound capability interface declared by the order module.
// Query and Command handlers depend on this shape — not on the product module's API directly.
type ProductCatalog interface {
	GetByModelNumber(ctx context.Context, modelNumber string) (res productdto.ProductDetailResponse, err error)
	FindByModelNumbers(ctx context.Context, modelNumbers []string) (res []productdto.ProductListItemResponse, err error)
}

// ProductAPI is the adaptee interface that mirrors product/api.Service's public surface.
// product/*Service satisfies this structurally; main.go wires it as ServiceDeps.ProductService.
// When the product module adds new methods, extend this interface here — not ProductCatalog.
type ProductAPI interface {
	GetByModelNumber(ctx context.Context, modelNumber string) (res productdto.ProductDetailResponse, err error)
	Search(ctx context.Context, req productdto.ProductSearchRequest) (res productdto.ProductSearchResponse, err error)
}
