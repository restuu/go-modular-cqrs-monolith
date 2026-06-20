package adapter

import (
	"context"
	"fmt"

	"go-modular-cqrs-monolith/modules/order/internal/port"
	productdto "go-modular-cqrs-monolith/modules/product/api/dto"
)

// OrderProductAdapter translates order's ProductCatalog interface into product's ProductAPI shape.
// It exists because the two shapes diverge: order wants FindByModelNumbers([]string) while product
// exposes Search(ProductSearchRequest). All translation logic lives here; neither side is changed.
type OrderProductAdapter struct {
	productAPI port.ProductAPI
}

func NewOrderProductAdapter(productAPI port.ProductAPI) *OrderProductAdapter {
	return &OrderProductAdapter{
		productAPI: productAPI,
	}
}

func (a *OrderProductAdapter) GetByModelNumber(ctx context.Context, modelNumber string) (
	res productdto.ProductDetailResponse,
	err error,
) {
	return a.productAPI.GetByModelNumber(ctx, modelNumber)
}

// FindByModelNumbers satisfies port.ProductCatalog by mapping to a single Search page.
// Assumption: one record per model number; page size matches input length exactly.
// An empty slice returns an empty result without calling Search.
func (a *OrderProductAdapter) FindByModelNumbers(
	ctx context.Context,
	modelNumbers []string,
) ([]productdto.ProductListItemResponse, error) {
	if len(modelNumbers) == 0 {
		return nil, nil
	}

	req := productdto.ProductSearchRequest{
		ModelNumbers: modelNumbers,
		Page:         1,
		Size:         len(modelNumbers),
	}

	searchRes, err := a.productAPI.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("find products by model numbers: %w", err)
	}

	return searchRes.Items, nil
}
