package domain

import "github.com/shopspring/decimal"

type Order struct {
	ID int
}

type OrderItem struct {
	ID                  int
	OrderID             int
	ProductModelNumber  string
	ProductSellingPrice decimal.Decimal
	ProductPrice        decimal.Decimal
	Quantity            int
}
