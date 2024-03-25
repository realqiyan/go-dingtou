package main

import (
	"log"
	"os"

	dmodel "dingtou/model"

	"github.com/joho/godotenv"

	"encoding/json"
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
	dmodel.InitDatabase(dsn)

	// 测试
	var stocks []dmodel.Stock
	stocks, err = dmodel.GetOwnerStocks("weibo_2685310785")
	if err == nil {
		for _, stock := range stocks {
			bytes, _ := json.MarshalIndent(stock.GetTradeCfg(), "", "\t")
			log.Printf("Code:%s,Name:%s,TradeCfg:%s", stock.Code, stock.Name, string(bytes))
		}
	}

}
