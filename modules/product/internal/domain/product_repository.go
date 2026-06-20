package domain

import "context"

type ProductRepository interface {
	FindProductByModelNumber(ctx context.Context, modelNumber string) (Product, error)
}
