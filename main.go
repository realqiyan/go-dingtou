package main

import (
	"log"
	"os"

	"dingtou/config"

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

}
