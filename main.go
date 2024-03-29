package main

import (
	"log"
	"os"
	"time"

	"dingtou/config"
	"dingtou/domain"

	"github.com/joho/godotenv"
)

func main() {
	log.Printf("start dingtou app.")

	// 从本地读取环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 数据库初始化
	dsn := os.Getenv("DB_DSN")
	log.Printf("DB_DSN:%v", dsn)
	config.InitDatabase(dsn)

	// 测试
	var stocks []domain.Stock
	stocks, err = domain.GetOwnerStocks("weibo_2685310785")
	if err == nil {
		var lastStock domain.Stock
		for _, stock := range stocks {
			lastStock = stock
			orders, _ := stock.GetStockOrders()
			log.Printf("Code:%s,Name:%s,orders:%v", stock.Code, stock.Name, len(orders))

			for _, order := range orders {
				snapshot := order.GetSnapshot()
				log.Printf("orderId:%s,orders:%v,currentTargetValue:%v", order.OutId, snapshot.BuyOrderOutIds, snapshot.TradeCfg.Attributes.CurrentTargetValue)
			}
		}

		pricePull := domain.BuildPricePull(&lastStock)
		price := pricePull.CurrentPrice()
		log.Printf("pricePull:%T,value:%v", price, price)

		t, _ := time.Parse("2006-01-02", "2024-03-21")
		log.Printf("Parse:%v", t)
		pricePull.ListPrice(t, 5)
	}

}
