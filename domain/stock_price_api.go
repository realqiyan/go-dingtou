package domain

import (
	"time"
)

// 价格拉取策略
type PriceAPI interface {

	/**
	 * 获取当前金额
	 *
	 * @return 当前金额
	 */
	GetCurrentPrice() float64

	/**
	 * 价格列表
	 *
	 * @param date  当前日期
	 * @param x     交易日数量
	 * @return 价格列表
	 */
	ListPrice(date time.Time, x int16) []StockPrice

	/**
	 * 获取结算金额
	 *
	 * @param date  交易时间
	 * @return 结算金额
	 */
	GetSettlementPrice(date time.Time) float64

	/**
	 * 获取指数估值比例
	 *
	 * @return 指数估值比例
	 */
	GetIndexValuationRatio() float64
}
