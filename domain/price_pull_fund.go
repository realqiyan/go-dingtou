package domain

import (
	"time"
)

// 价格拉取策略
type FundPricePull struct {
	Stock *Stock
}

// CurrentPrice implements PricePull.
func (f FundPricePull) CurrentPrice() float64 {
	panic("unimplemented")
}

// GetSettlementPrice implements PricePull.
func (f FundPricePull) GetSettlementPrice(date time.Time) float64 {
	panic("unimplemented")
}

// ListPrice implements PricePull.
func (f FundPricePull) ListPrice(date time.Time, x int16) []StockPrice {
	panic("unimplemented")
}
