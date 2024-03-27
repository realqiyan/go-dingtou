package domain

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 价格拉取策略
type StockPricePull struct {
	Stock *Stock
}

// CurrentPrice implements PricePull.
func (s StockPricePull) CurrentPrice() float64 {
	var url = "http://qt.gtimg.cn/q=%s%s"
	var target_url = fmt.Sprintf(url, s.Stock.Market, s.Stock.Code)
	log.Printf("%s", target_url)
	resp, err := http.Get(target_url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	body := string(bodyByte)
	price := strings.Split(body, "~")[3]

	log.Printf("%s", price)
	ret, err := strconv.ParseFloat(price, 64)
	if err != nil {
		panic(err)
	}
	return ret
}

// GetSettlementPrice implements PricePull.
func (s StockPricePull) GetSettlementPrice(date time.Time) float64 {
	panic("unimplemented")
}

// ListPrice implements PricePull.
func (s StockPricePull) ListPrice(date time.Time, x int16) []StockPrice {
	panic("unimplemented")
}
