package service

import (
	"dingtou/domain"
	"encoding/json"
	"log"
	"sort"
	"sync"
	"time"
)

type StockService struct {
}

/**
 * 查询证券
 */
func (s StockService) Query(owner string) ([]domain.Stock, error) {
	return domain.GetOwnerStocks(owner)
}

type TradeService struct {
}

/**
 * 计算股票基金购买金额
 */
func (t TradeService) Conform(owner string) ([]domain.StockOrder, error) {
	stocks, err := domain.GetOwnerStocks(owner)
	if err != nil {
		log.Printf("GetOwnerStocks(%s) error:%v", owner, err)
		return nil, err
	}
	size := len(stocks)
	orders := make([]domain.StockOrder, 0)

	// async
	var wg sync.WaitGroup
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func(stock *domain.Stock) {
			var order domain.StockOrder
			stock.Conform(&order)
			orders = append(orders, order)
			wg.Done()
		}(&stocks[i])
	}
	wg.Wait()

	// 按照StockId从小到大排序
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].StockId < orders[j].StockId
	})

	return orders, nil

}

/**
 * 结算股票基金
 */
func (t TradeService) Settlement(owner string) ([]domain.StockOrder, error) {
	stocks, err := domain.GetOwnerStocks(owner)
	if err != nil {
		log.Printf("GetOwnerStocks(%s) error:%v", owner, err)
		return nil, err
	}
	size := len(stocks)
	orders := make([]domain.StockOrder, 0)
	// async
	var wg sync.WaitGroup
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func(stock *domain.Stock) {
			waitProcessOrders, _ := stock.GetStockWaitProcessOrders()
			for _, order := range waitProcessOrders {
				stock.Settlement(&order)
				order.Update()
				orders = append(orders, order)
			}
			stock.Update()
			wg.Done()
		}(&stocks[i])
	}
	wg.Wait()

	return orders, nil

}

/**
 * 购买股票基金
 */
func (t TradeService) Buy(order domain.StockOrder) (domain.StockOrder, error) {
	stock := order.Stock

	tradeCfgByte, _ := json.Marshal(stock.TradeCfgStruct)
	stock.TradeCfg = string(tradeCfgByte)

	// 创建订单
	err := order.Create()
	if err == nil {
		// 更新stock
		stock.LastTradeTime = time.Now()
		err = stock.Update()
	}

	return order, err
}
