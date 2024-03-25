package service

import dmodel "dingtou/model"

type StockService interface {

	/**
	 * 查询证券
	 */
	Query(owner string) []dmodel.Stock
}

type TradeService interface {

	/**
	 * 计算股票基金购买金额
	 */
	conform(owner string) []dmodel.StockOrder

	/**
	 * 购买股票基金
	 */
	buy(order dmodel.StockOrder) dmodel.StockOrder
}
