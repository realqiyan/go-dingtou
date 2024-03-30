package domain

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 价格拉取策略
type StockPricePull struct {
	Stock *Stock
}

type originalStockPrice struct {
	Day    string `json:"day"`
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Close  string `json:"close"`
	Volume string `json:"volume"`
}

type originalStockAdjustResult struct {
	Total int                       `json:"total"`
	Data  []originalStockAdjustItem `json:"data"`
}
type originalStockAdjustItem struct {
	/**
	 * 复权日期
	 */
	AdjustDateStr string `json:"d"`
	AdjustDate    time.Time
	/**
	 * 复权比例
	 */
	AdjustValStr string `json:"f"`
	AdjustVal    float64
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
	priceUrlTemplate := "https://quotes.sina.cn/cn/api/json_v2.php/CN_MarketDataService.getKLineData?symbol=%s%s&scale=240&ma=no&datalen=%d"
	pullPriceUrl := fmt.Sprintf(priceUrlTemplate, s.Stock.Market, s.Stock.Code, datalen)
	bodyByte := getContent(pullPriceUrl)
	var stockPriceArr []originalStockPrice
	_ = json.Unmarshal(bodyByte, &stockPriceArr)
	log.Printf("%s", stockPriceArr)

	//前复权:https://finance.sina.com.cn/realstock/company/sz000002/qfq.js
	adjustUrlTemplate := "https://finance.sina.com.cn/realstock/company/%s%s/qfq.js"
	pullAdjustUrl := fmt.Sprintf(adjustUrlTemplate, s.Stock.Market, s.Stock.Code)
	stockAdjust := getContent(pullAdjustUrl)
	log.Printf("%s", stockAdjust)

	reg, _ := regexp.Compile("({.*})")
	if reg.Match(stockAdjust) {
		var adjustResult originalStockAdjustResult
		_ = json.Unmarshal(reg.FindAll(stockAdjust, 1)[0], &adjustResult)

		layout := "2006-01-02"
		for _, adjustData := range adjustResult.Data {
			log.Printf("adjustData:%v,AdjustVal:%v", adjustData.AdjustDateStr, adjustData.AdjustValStr)
			d, _ := time.Parse(layout, adjustData.AdjustDateStr)
			adjustData.AdjustDate = d

			f, _ := strconv.ParseFloat(adjustData.AdjustValStr, 64)
			adjustData.AdjustVal = f
			log.Printf("adjustData:%v,AdjustVal:%v", adjustData.AdjustDate, adjustData.AdjustVal)
		}

		// var stockAdjustArr []originalStockAdjust
		// data, _ := json.Marshal(adjustMap["data"])

		// log.Printf("data:%v", string(data))

		// _ = json.Unmarshal(data, &stockAdjustArr)

		// log.Printf("%v", stockAdjustArr)
	}

	panic("unimplemented")
}

func getContent(url string) []byte {
	log.Printf("%s", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return bodyByte
}
