package service

import "dingtou/domain"

type StockService interface {

	/**
	 * 查询证券
	 */
	Query(owner string) []domain.Stock
}

type TradeService interface {

	/**
	 * 计算股票基金购买金额
	 */
	conform(owner string) []domain.StockOrder

	/**
	 * 购买股票基金
	 */
	buy(order domain.StockOrder) domain.StockOrder
}
