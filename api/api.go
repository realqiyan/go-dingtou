package api

import "dingtou/domain"

type StockAPI interface {

	/**
	 * 查询证券
	 */
	Query(owner string) ([]domain.Stock, error)
}

type TradeAPI interface {

	/**
	 * 计算股票基金购买金额
	 */
	conform(owner string) ([]domain.StockOrder, error)

	/**
	 * 购买股票基金
	 */
	buy(order domain.StockOrder) (domain.StockOrder, error)
}
