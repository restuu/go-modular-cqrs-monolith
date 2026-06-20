package dto

import "github.com/shopspring/decimal"

type ProductDetailResponse struct {
	ModelNumber string
	Name        string
	Price       decimal.Decimal
}

type ProductSearchResponse struct {
	Items []ProductListItemResponse
	Total int
}

type ProductListItemResponse struct {
	ModelNumber string
}
