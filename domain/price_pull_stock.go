package domain

import (
	"encoding/json"
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

type sinaStockPrice struct {
	Day    string `json:"day"`
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Close  string `json:"close"`
	Volume string `json:"volume"`
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
	now := time.Now()
	between := int16(now.Sub(date).Abs().Hours() / 24)
	datalen := between + x

	//日K:https://quotes.sina.cn/cn/api/json_v2.php/CN_MarketDataService.getKLineData?symbol=sz000002&scale=240&ma=no&datalen=30
	var pullUrl = "https://quotes.sina.cn/cn/api/json_v2.php/CN_MarketDataService.getKLineData?symbol=%s%s&scale=240&ma=no&datalen=%d"
	var pullTargetUrl = fmt.Sprintf(pullUrl, s.Stock.Market, s.Stock.Code, datalen)
	log.Printf("%s", pullTargetUrl)
	resp, err := http.Get(pullTargetUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var sinaStockPriceArr []sinaStockPrice
	_ = json.Unmarshal(bodyByte, &sinaStockPriceArr)
	log.Printf("%s", sinaStockPriceArr)

	//前复权:https://finance.sina.com.cn/realstock/company/sz000002/qfq.js
	// adjustApiUrl := fmt.Sprintf("https://finance.sina.com.cn/realstock/company/%s%s/qfq.js", s.Stock.Market, s.Stock.Code)

	panic("unimplemented")
}
