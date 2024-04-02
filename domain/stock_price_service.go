package domain

import (
	"dingtou/util"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type indexValuation struct {
	Data struct {
		IndexCode    string  `json:"index_code"`
		Name         string  `json:"name"`
		Pe           float64 `json:"pe"`
		Pb           float64 `json:"pb"`
		PePercentile float64 `json:"pe_percentile"`
		PbPercentile float64 `json:"pb_percentile"`
		Roe          float64 `json:"roe"`
		Yeild        float64 `json:"yeild"`
		EvaType      string  `json:"eva_type"`
		EvaTypeInt   int     `json:"eva_type_int"`
		BondYeild    float64 `json:"bond_yeild"`
		Peg          float64 `json:"peg"`
		Date         string  `json:"date"`
	} `json:"data"`
	ResultCode int `json:"result_code"`
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
	Total int                  `json:"total"`
	Data  stockAdjustItemSlice `json:"data"`
}

type stockAdjustItemSlice []originalStockAdjustItem

// Len is the number of elements in the collection.
func (s stockAdjustItemSlice) Len() int {
	return len(s)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (s stockAdjustItemSlice) Less(i, j int) bool {
	// 近期排前 历史排后
	return s[i].AdjustDate.After(s[j].AdjustDate)
}

// Swap swaps the elements with indexes i and j.
func (s stockAdjustItemSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
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

// GetCurrentPrice implements PriceAPI.
func (s *Stock) GetCurrentPrice() float64 {
	var url = "http://qt.gtimg.cn/q=%s%s"
	var target_url = fmt.Sprintf(url, s.Market, s.Code)
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

// GetSettlementPrice implements PriceAPI.
func (s *Stock) GetSettlementPrice(date time.Time) float64 {
	stockPrices := s.ListPrice(date, 1)
	if len(stockPrices) >= 1 {
		return stockPrices[0].Price
	}
	return 0
}

// ListPrice implements PriceAPI.
func (s *Stock) ListPrice(date time.Time, x int16) []StockPrice {
	now := time.Now()
	between := int16(now.Sub(date).Abs().Hours() / 24)
	datalen := between + x

	//日K:https://quotes.sina.cn/cn/api/json_v2.php/CN_MarketDataService.getKLineData?symbol=sz000002&scale=240&ma=no&datalen=30
	priceUrlTemplate := "https://quotes.sina.cn/cn/api/json_v2.php/CN_MarketDataService.getKLineData?symbol=%s%s&scale=240&ma=no&datalen=%d"
	pullPriceUrl := fmt.Sprintf(priceUrlTemplate, s.Market, s.Code, datalen)
	bodyByte := getContent(pullPriceUrl)
	var stockPriceArr []originalStockPrice
	_ = json.Unmarshal(bodyByte, &stockPriceArr)
	log.Printf("%s", stockPriceArr)

	layout := "2006-01-02"
	size := len(stockPriceArr)
	stockPriceSlice := make([]StockPrice, size)
	for i := size - 1; i >= 0; i-- {
		var stockPrice StockPrice
		stockPrice.Stock = s
		d, _ := time.Parse(layout, stockPriceArr[i].Day)
		stockPrice.Date = d
		f, _ := strconv.ParseFloat(stockPriceArr[i].Close, 64)
		stockPrice.Price = f
		stockPriceSlice[i] = stockPrice
	}

	//前复权:https://finance.sina.com.cn/realstock/company/sz000002/qfq.js
	adjustUrlTemplate := "https://finance.sina.com.cn/realstock/company/%s%s/qfq.js"
	pullAdjustUrl := fmt.Sprintf(adjustUrlTemplate, s.Market, s.Code)
	stockAdjust := getContent(pullAdjustUrl)
	log.Printf("%s", stockAdjust)

	reg, _ := regexp.Compile("({.*})")
	if reg.Match(stockAdjust) {
		var adjustResult originalStockAdjustResult
		_ = json.Unmarshal(reg.FindAll(stockAdjust, 1)[0], &adjustResult)

		for i := len(adjustResult.Data) - 1; i >= 0; i-- {
			//log.Printf("adjustDataStr:%s,AdjustValStr:%s", adjustData.AdjustDateStr, adjustData.AdjustValStr)
			d, _ := time.Parse(layout, adjustResult.Data[i].AdjustDateStr)
			adjustResult.Data[i].AdjustDate = d

			f, _ := strconv.ParseFloat(adjustResult.Data[i].AdjustValStr, 64)
			adjustResult.Data[i].AdjustVal = f
			//log.Printf("adjustData:%v,AdjustVal:%v", adjustResult.Data[i].AdjustDate, adjustResult.Data[i].AdjustVal)
		}

		// 近期排前 历史排后
		sort.Sort(adjustResult.Data)
		log.Printf("after sort adjustData:%v", adjustResult.Data)

		// 计算复权价格
		for i := size - 1; i >= 0; i-- {
			calcRehabPrice(&stockPriceSlice[i], adjustResult.Data)
		}

	}

	return stockPriceSlice

}

// GetIndexValuationRatio implements PriceAPI.
func (s *Stock) GetIndexValuationRatio() float64 {
	targetIndexCode := s.GetTradeCfg().Attributes.TargetIndexCode
	if targetIndexCode != "" {
		template := "https://danjuanapp.com/djapi/index_eva/detail/%s"
		url := fmt.Sprintf(template, targetIndexCode)
		log.Printf("GetIndexValuationRatio url:%s", url)
		bodyByte := getContent(url)

		var indexValuation indexValuation
		_ = json.Unmarshal(bodyByte, &indexValuation)
		log.Printf("indexValuation:%v", indexValuation)
		if indexValuation.ResultCode == 0 {
			return indexValuation.Data.PePercentile
		}
	}
	return 0
}

func calcRehabPrice(stockPrice *StockPrice, stockAdjustItemSlice stockAdjustItemSlice) {
	for _, adjust := range stockAdjustItemSlice {
		if stockPrice.Date.After(adjust.AdjustDate) || stockPrice.Date.Equal(adjust.AdjustDate) {
			stockPrice.RehabPrice = util.FloatDiv(stockPrice.Price, adjust.AdjustVal)
			return
		}
	}
}

func getContent(url string) []byte {
	log.Printf("%s", url)

	req, _ := http.NewRequest("GET", "https://forward.myworker.win/"+url, nil)
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"123\", \"Not:A-Brand\";v=\"8\", \"Chromium\";v=\"123\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")

	resp, err := http.DefaultClient.Do(req)
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
