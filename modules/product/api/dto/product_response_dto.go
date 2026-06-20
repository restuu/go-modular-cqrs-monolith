package dto

import "github.com/shopspring/decimal"

type ProductDetailResponse struct {
	ModelNumber string
	Name        string
	Price       decimal.Decimal
}
