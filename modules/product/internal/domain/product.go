package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Product struct {
	InternalID  int64
	ModelNumber string
	Name        string
	Price       decimal.Decimal
	CreatedAt   time.Time
}
