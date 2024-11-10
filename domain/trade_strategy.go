package domain

import (
	"dingtou/util"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"
)

/**
 * const SMA_STRATEGY_DEFAULT_VALUE = "4|10,30,60,120|1.5,1.25,1,0"
 * sma策略 配置说明：均线数量n|均线1,均线2,均线n|高于1条均线购买比例,高于2条均线购买比例,高于n条均线购买比例 例如：4|10,30,60,120|1.5,1.25,1,0
 * 注意：自由落体趋势会暂停购买（低于所有均线不买入）
 */
type SmaStrategy struct {
	LineSize  uint8   //0-255
	LineLevel []uint8 //0-255
	BuyRatio  []float32
}

var smaStrategyConfig SmaStrategy = SmaStrategy{
	LineSize:  4,
	LineLevel: []uint8{10, 30, 60, 120},
	BuyRatio:  []float32{1.5, 1.25, 1, 1},
}

/**
 * 卖出时，当前价格涨幅少于比例不卖出 15%
 */
const sellProfitRatioDefaultVal = "0.15"

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
func CalculateConform(stock *Stock, orders []StockOrder, tradeDate time.Time) TradeDetail {

	layout := "20060102"

	tradeCfg := stock.GetTradeCfg()
	attributes := &tradeCfg.Attributes

	// 已经投入部分的目标价值
	currentTargetValue, _ := strconv.ParseFloat(attributes.CurrentTargetValue, 64)

	// 冗余记录上期目标价值
	attributes.PreTargetValue = fmt.Sprintf("%f", currentTargetValue)

	// 当天已经买过就跳过计算
	for _, order := range orders {
		if order.TradeTime.Format(layout) == tradeDate.Format(layout) {
			return TradeDetail{targetValue: currentTargetValue}
		}
	}

	// 当前价格
	currentPrice := stock.GetCurrentPrice()

	// 冗余记录实时交易价格
	attributes.CurrentTradePrice = fmt.Sprintf("%f", currentPrice)

	//
	// 计算步长 - 均线策略smaStrategyConfig
	increment := calculateIncrement(stock, tradeCfg, attributes, currentPrice, smaStrategyConfig)

	// 冗余记录步长
	attributes.CurrentIncrement = fmt.Sprintf("%f", increment)

	// 目标价值=上期目标价值+increment
	targetValue := util.FloatAdd(currentTargetValue, increment)

	// 买入上限
	if attributes.MaxTargetValue != "" {
		maxTargetValue, _ := strconv.ParseFloat(attributes.MaxTargetValue, 64)
		if targetValue > maxTargetValue {
			log.Printf("stock:%s,maxTargetValue:%f,targetValue:%f", stock.Code, maxTargetValue, targetValue)
			targetValue = maxTargetValue
		}
	}

	// 冗余记录本次目标价值
	attributes.CurrentTargetValue = fmt.Sprintf("%f", targetValue)

	// 当前总份额
	amount := stock.Amount
	// 当前总价值
	currentValue := util.FloatMul(amount, currentPrice)

	// 交易金额
	tradeFee := util.FloatSub(targetValue, currentValue)

	// 最大单次交易金额处理 PER_MAX_TRADE_PRICE
	if attributes.PerMaxTradePrice != "" {
		perMaxTradePrice, _ := strconv.ParseFloat(attributes.PerMaxTradePrice, 64)
		if tradeFee > perMaxTradePrice {
			// log.Printf("stock:%s,perMaxTradePrice:%f,tradeFee:%f", stock.Code, perMaxTradePrice, tradeFee)
			tradeFee = perMaxTradePrice
		}
	}

	// sellProfitRatio卖出的盈利比例
	if attributes.SellProfitRatio == "" {
		attributes.SellProfitRatio = sellProfitRatioDefaultVal
	}
	sellProfitRatio, _ := strconv.ParseFloat(attributes.SellProfitRatio, 64)

	//最总交易金额（tradeFee）如果是买入则继续计算，如果是卖出，就去匹配历史交易订单。
	if tradeFee >= 0 {
		return buy(tradeCfg, targetValue, tradeFee, currentPrice)
	} else {
		return sell(tradeCfg, orders, targetValue, tradeFee, currentPrice, sellProfitRatio)
	}

}

/**
 * CalculateSettlement
 */
func CalculateSettlement(order *StockOrder) (TradeDetail, error) {
	var tradeDetail TradeDetail
	tradeFee := order.TradeFee
	tradeServiceFee := order.TradeServiceFee
	tradeAmount := order.TradeAmount

	if order.Stock.Type == "fund" {
		settlementPrice := order.Stock.GetSettlementPrice(order.TradeTime)
		if settlementPrice <= 0 {
			return tradeDetail, fmt.Errorf("%s settlementPrice is nil", order.Code)
		}
		var realTradeFee float64
		if order.Type == util.BUY {
			realTradeFee = util.FloatSub(tradeFee, tradeServiceFee)
		} else if order.Type == util.SELL {
			realTradeFee = tradeFee
		}
		tradeAmount = util.FloatDiv(realTradeFee, settlementPrice)
	}

	tradeDetail.TradeFee = tradeFee
	tradeDetail.TradeAmount = tradeAmount
	tradeDetail.TradeServiceFee = tradeServiceFee

	return tradeDetail, nil
}

func buy(tradeCfg *TradeCfg, targetValue, tradeFee, currentPrice float64) TradeDetail {
	// 计算交易份额
	tradeAmount := util.FloatDiv(tradeFee, currentPrice)
	// 最小交易份额（一手）
	minTradeAmount := tradeCfg.MinTradeAmount

	// 不足一手的部分放弃（简化计算）
	batchSize := int(util.FloatDiv(tradeAmount, minTradeAmount))

	//购买数量
	tradeAmount = util.FloatMul(float64(batchSize), minTradeAmount)

	//购买金额
	tradeFee = util.FloatMul(currentPrice, tradeAmount)

	// 手续费
	tradeServiceFee := util.FloatMul(tradeFee, tradeCfg.ServiceFeeRate)
	if tradeServiceFee < tradeCfg.MinServiceFee {
		tradeServiceFee = tradeCfg.MinServiceFee
	}

	return TradeDetail{
		targetValue:     targetValue,
		TradeFee:        tradeFee,
		TradeAmount:     tradeAmount,
		TradeServiceFee: tradeServiceFee,
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func sell(tradeCfg *TradeCfg, orders []StockOrder, targetValue, tradeFee, currentPrice float64, sellProfitRatio float64) TradeDetail {

	var selledOrderOutIds []string //已经卖出的订单outId
	var canSellOrders []StockOrder //可以卖出的订单

	// 需要过滤已经卖出的订单
	for _, order := range orders {
		if order.Type == util.SELL {
			selledOrderOutIds = append(selledOrderOutIds, order.GetSnapshot().BuyOrderOutIds...)
		}
	}
	log.Printf("selledOrderOutIds:%v", selledOrderOutIds)

	// 找出可以卖的订单
	sellTotalFee := util.FloatAbs(tradeFee)
	for _, order := range orders {
		orderCurrentPrice := util.FloatMul(order.TradeAmount, currentPrice)
		if order.Type == util.BUY && !contains(selledOrderOutIds, order.OutId) && order.TradeStatus == "done" && order.TradeFee > 0 && orderCurrentPrice < sellTotalFee {
			// 当前盈利
			currentProfitFee := util.FloatSub(orderCurrentPrice, order.TradeFee)
			currentProfitFee = util.FloatSub(currentProfitFee, order.TradeServiceFee)

			// 当前盈利比例
			currentProfitRatio := util.FloatDiv(currentProfitFee, order.TradeFee)

			// 大于sellProfitRatio才卖
			if currentProfitRatio >= sellProfitRatio {
				order.CurrentProfitFee = currentProfitFee
				order.CurrentProfitRatio = currentProfitRatio
				canSellOrders = append(canSellOrders, order)
			}
		}
	}
	//log.Printf("canSellOrders:%v", canSellOrders)

	sellTotalAmount := util.FloatDiv(sellTotalFee, currentPrice)

	var sellOrders []StockOrder
	var sellAmount float64 = 0

	// 按照金额从小到大排序
	sort.Slice(canSellOrders, func(i, j int) bool {
		return canSellOrders[i].TradeFee < canSellOrders[j].TradeFee
	})

	for _, order := range canSellOrders {
		tradeAmount := order.TradeAmount
		if tradeAmount > sellTotalAmount || util.FloatAdd(sellAmount, tradeAmount) > sellTotalAmount {
			break
		}
		sellOrders = append(sellOrders, order)
		sellAmount = util.FloatAdd(sellAmount, tradeAmount)
	}

	log.Printf("sellOrders:%v", sellOrders)

	sellAmount = -sellAmount
	sellFee := util.FloatMul(sellAmount, currentPrice)

	// 手续费
	tradeServiceFee := util.FloatMul(-sellFee, tradeCfg.ServiceFeeRate)
	if tradeServiceFee < tradeCfg.MinServiceFee {
		tradeServiceFee = tradeCfg.MinServiceFee
	}

	return TradeDetail{
		targetValue:     targetValue,
		TradeFee:        sellFee,
		TradeAmount:     sellAmount,
		TradeServiceFee: tradeServiceFee,
		SellOrders:      sellOrders,
	}
}

func maxInSlice(nums []uint8) uint8 {
	max := nums[0]
	for _, value := range nums {
		if value > max {
			max = value
		}
	}
	return max
}

func calculateIncrement(stock *Stock, tradeCfg *TradeCfg, attributes *Attributes, currentPrice float64, smaStrategyConfig SmaStrategy) float64 {

	// 默认步长
	increment := tradeCfg.Increment

	// 基于估值水位调整步长
	increment = calculateIncrementByValuationRatio(stock, attributes, increment)

	// 注意：股价不是前复权时统计会有问题
	now := time.Now()

	// 计算均线平均价格
	// 均线&价格
	average := make(map[uint8]float64)

	// 通过获取最大的日期间隔
	maxDay := maxInSlice(smaStrategyConfig.LineLevel)
	maxStockPriceSlice := stock.ListPrice(now, int16(maxDay))
	// log.Printf("maxStockPriceSlice:%v", maxStockPriceSlice)

	for _, v := range smaStrategyConfig.LineLevel {
		totalPrice := 0.0
		// stockPriceSlice := stock.ListPrice(now, int16(v))
		stockPriceSlice := maxStockPriceSlice[:v]
		for _, price := range stockPriceSlice {
			if price.RehabPrice <= 0 {
				return increment
			}
			totalPrice = util.FloatAdd(totalPrice, price.RehabPrice)
		}
		average[v] = util.FloatDiv(totalPrice, float64(len(stockPriceSlice)))
	}

	// 比较现价超过均线数量来决定浮动比例
	// 不用均线定比例的原因：下跌趋势过程中120均价>60均价>30均价>现价，股价突然上升，120均价>现价>60均价>30均价，这时现价低于120均线，大于60均价和30均价，使用120均线不合适。
	overNum := -1
	for _, v := range smaStrategyConfig.LineLevel {
		linePrice := average[v]
		if currentPrice > linePrice {
			overNum++
		}
	}
	//低于所有均线就进入下跌通道了，暂停买入。
	var buyRatio float64 = 0
	if overNum > -1 {
		buyRatio = float64(smaStrategyConfig.BuyRatio[overNum])
	}
	log.Printf("stock:%s,overNum-1:%v,multiplyVal:%v", stock.Code, overNum, buyRatio)
	increment = util.FloatMul(increment, buyRatio)

	return increment
}

func calculateIncrementByValuationRatio(stock *Stock, attributes *Attributes, increment float64) float64 {
	// 跟踪的指数估值

	indexValuationRatio := stock.GetIndexValuationRatio()

	if indexValuationRatio > 0 {

		// 冗余记录当前指数估值
		attributes.CurrentTargetIndexValuation = fmt.Sprintf("%f", indexValuationRatio)

		// 估值水位75%～100% 0倍
		// 估值水位50%～75%  0.5倍
		// 估值水位25%～50%  1倍
		// 估值水位 0%～25%  1.5倍
		if indexValuationRatio > 0.75 {
			increment = 0
		} else if indexValuationRatio > 0.5 {
			increment = util.FloatMul(increment, 0.5)
		} else if indexValuationRatio > 0.25 {
			increment = util.FloatMul(increment, 1.0)
		} else if indexValuationRatio >= 0.0 {
			increment = util.FloatMul(increment, 1.5)
		}

	}
	return increment
}
