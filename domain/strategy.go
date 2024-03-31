package domain

import "time"

/**
 * 价值平均定投策略增强版（增加上浮比例）
 *
 * 当前持有价值=当前基金持有份额*当前基金价格
 * 定投目标价值=本次定投结束后所持有基金的总价值
 * 上期目标价值=上一期计算出来的定投目标价值
 *
 * 定投目标价值=上期目标价值+首次定投金额*(1+上浮比例)
 * 说明:首次定投时“上期目标价值 ”为0
 *
 * 本期定投金额=定投目标价值-当前持有价值
 */
func CalculateConform(stock *Stock, orders []StockOrder, date time.Time) TradeDetail {
	var tradeDetail TradeDetail

	return tradeDetail
}
