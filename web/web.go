package web

import (
	"dingtou/domain"
	"dingtou/service"
	"dingtou/util"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var tradeService service.TradeService

func InitWeb(port int) {

	router := gin.Default()

	router.StaticFS("/static", http.Dir("./web/static"))

	router.GET("/", func(c *gin.Context) {
		//url重定向（坑）
		c.Redirect(http.StatusMovedPermanently, "/static/index.html") //301 永久移动
	})

	trade := router.Group("/trade")
	{
		trade.GET("/conform", tradeConform)
		trade.POST("/buy", tradeBuy)
	}

	router.Run(fmt.Sprintf("0.0.0.0:%d", port))
}

func tradeConform(c *gin.Context) {
	owner := os.Getenv("DEFAULT_OWNER")
	orders, err := tradeService.Conform(owner)
	if err != nil {
		c.JSON(500, err.Error())
	}
	c.JSON(200, orders)
}

func tradeBuy(c *gin.Context) {
	owner := os.Getenv("DEFAULT_OWNER")
	formOutIds := c.PostForm("outIds")
	var outIds []string
	json.Unmarshal([]byte(formOutIds), &outIds)

	formOrders := c.PostForm("orders")
	var orders []domain.StockOrder
	json.Unmarshal([]byte(formOrders), &orders)

	var buyOrders []domain.StockOrder
	for _, order := range orders {
		if util.StringInSlice(outIds, order.OutId) && owner == order.Stock.Owner {
			buyOrder, _ := tradeService.Buy(order)
			buyOrders = append(buyOrders, buyOrder)
		}
	}
	c.JSON(200, buyOrders)

}
