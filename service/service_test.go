package service

import (
	"dingtou/config"
	"dingtou/domain"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func setup(t *testing.T) {
	// 从本地读取环境变量
	err := godotenv.Load()
	if err != nil {
		t.Errorf("Error loading .env file error = %v", err)
	}

	// 数据库初始化
	dsn := os.Getenv("DB_DSN")
	config.InitDatabase(dsn)
}

func TestStockService_Query(t *testing.T) {
	setup(t)

	// 测试
	var stockService StockService
	stocks, err := stockService.Query("weibo_2685310785")

	// var stocks []domain.Stock
	// stocks, err = domain.GetOwnerStocks("weibo_2685310785")
	if err == nil {
		var lastStock domain.Stock
		for _, stock := range stocks {
			lastStock = stock
			orders, _ := stock.GetStockOrders()
			t.Logf("Code:%s,Name:%s,orders:%v", stock.Code, stock.Name, len(orders))

			for _, order := range orders {
				snapshot := order.GetSnapshot()
				t.Logf("orderId:%s,orders:%v,currentTargetValue:%v", order.OutId, snapshot.BuyOrderOutIds, snapshot.TradeCfg.Attributes.CurrentTargetValue)
			}
		}

		pricePull := domain.BuildPricePull(&lastStock)
		price := pricePull.CurrentPrice()
		t.Logf("pricePull:%T,value:%v", price, price)

		d, _ := time.Parse("2006-01-02", "2024-03-21")
		stockPriceSlice := pricePull.ListPrice(d, 5)
		t.Logf("stockPriceSlice:%v", stockPriceSlice)
	} else {
		t.Errorf("StockService.Query() error = %v", err)
	}

}

func TestTradeService_Conform(t *testing.T) {
	setup(t)

	// 测试
	var tradeService TradeService
	orders, err := tradeService.Conform("weibo_2685310785")

	if err == nil {
		for _, order := range orders {
			t.Logf("orders:%v", order)
		}

	} else {
		t.Errorf("StockService.Query() error = %v", err)
	}
}
