package domain

import "time"

type StockPrice struct {
	/**
	 * 股票基金
	 */
	Stock *Stock

	/**
	 * 日期
	 */
	Date time.Time

	/**
	 * 价格
	 */
	Price float64

	/**
	 * 复权价格
	 */
	RehabPrice float64
}
