package service

type StockService interface {

	/**
	 * 查询证券
	 */
	Query(owner string) []model.Stock
}

type TradeService interface {

	/**
	 * 计算股票基金购买金额
	 */
	conform(owner string) []model.Order

	/**
	 * 购买股票基金
	 */
	buy(order model.Order) model.Order
}
