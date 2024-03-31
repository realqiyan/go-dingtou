package service

import (
	"dingtou/domain"
	"log"
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
	orders := make([]domain.StockOrder, size)

	for i := 0; i < size; i++ {
		var stock *domain.Stock = &stocks[i]
		order, _ := stock.Conform()
		orders[i] = order
	}

	return orders, nil

}

/**
 * 购买股票基金
 */
func (t TradeService) Buy(order domain.StockOrder) (domain.StockOrder, error) {
	panic("unimplemented")
}
