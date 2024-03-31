package domain

import (
	"dingtou/util"
	"fmt"
	"log"
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
	BuyRatio:  []float32{1.5, 1.25, 1, 0},
}

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
	currentPrice := stock.CurrentPrice()

	// 冗余记录实时交易价格
	attributes.CurrentTradePrice = fmt.Sprintf("%f", currentPrice)

	//
	// 计算步长 - 均线策略smaStrategyConfig
	increment := calculateIncrement(stock, &tradeCfg, currentPrice, smaStrategyConfig)

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
			log.Printf("stock:%s,perMaxTradePrice:%f,tradeFee:%f", stock.Code, perMaxTradePrice, tradeFee)
			tradeFee = perMaxTradePrice
		}
	}

	//最总交易金额（tradeFee）如果是买入则继续计算，如果是卖出，就去匹配历史交易订单。
	if tradeFee >= 0 {
		return buy(&tradeCfg, targetValue, tradeFee, currentPrice)
	} else {
		return sell(&tradeCfg, orders, targetValue, tradeFee, currentPrice)
	}

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

func sell(tradeCfg *TradeCfg, orders []StockOrder, targetValue, tradeFee, currentPrice float64) TradeDetail {

	selledOrderOutIds := make([]string, 32) //已经卖出的订单outId
	canSellOrders := make([]StockOrder, 32) //可以卖出的订单

	// 需要过滤已经卖出的订单
	for _, order := range orders {
		if order.Type == "sell" {
			selledOrderOutIds = append(selledOrderOutIds, order.GetSnapshot().BuyOrderOutIds...)
		}
	}
	log.Printf("selledOrderOutIds:%v", selledOrderOutIds)

	//TODO 找出可以卖的订单
	for _, order := range orders {
		if order.Type == "buy" && !contains(selledOrderOutIds, order.OutId) {
			canSellOrders = append(canSellOrders, order)
		}
	}
	log.Printf("canSellOrders:%v", canSellOrders)

	// // 需要过滤已经卖出的订单
	// List<String> orderOutIds = stockOrders.stream()
	// .filter(order -> TradeType.SELL.equals(order.getType()))
	// .map(Order::getSnapshot)
	// .filter(Objects::nonNull)
	// .map(snapshot -> snapshot.getOrDefault(OrderSnapshotKeys.BUY_ORDER_OUT_IDS, null))
	// .filter(Objects::nonNull)
	// .flatMap(outIds -> JSON.parseArray(outIds, String.class).stream())
	// .collect(Collectors.toList());

	// // 找出可以卖的订单
	// BigDecimal sellTotalFee = tradeFee.abs();
	// List<Order> orderList = stockOrders.stream()
	// .filter(order -> TradeType.BUY.equals(order.getType()))
	// .filter(order -> !orderOutIds.contains(order.getOutId()))
	// .filter(order -> order.getStatus().equals(TradeStatus.DONE))
	// .filter(order -> order.getTradeFee().compareTo(BigDecimal.ZERO) > 0)
	// .filter(order -> order.getTradeAmount().multiply(currentPrice).compareTo(sellTotalFee) <= 0)
	// .peek(order -> {
	// 	BigDecimal currentProfitFee = order.getTradeAmount().multiply(currentPrice)
	// 			.subtract(order.getTradeFee())
	// 			.subtract(order.getTradeServiceFee());
	// 	order.setCurrentProfitFee(currentProfitFee);
	// 	BigDecimal currentProfitRatio = currentProfitFee.divide(order.getTradeFee(), 2, RoundingMode.HALF_UP);
	// 	order.setCurrentProfitRatio(currentProfitRatio);
	// })
	// .filter(order -> order.getCurrentProfitRatio().compareTo(BigDecimal.valueOf(SELL_PROFIT_RATIO)) > 0)
	// .sorted(Comparator.comparing(Order::getTradeFee))
	// .collect(Collectors.toList());

	// // 按照金额从小到大排序
	// BigDecimal sellTotalAmount = sellTotalFee.divide(currentPrice, 2, RoundingMode.HALF_UP);
	// List<Order> sellOrders = new ArrayList<>();
	// BigDecimal sellAmount = new BigDecimal(0);
	// for (Order order : orderList) {
	// BigDecimal tradeAmount = order.getTradeAmount();
	// if (tradeAmount.compareTo(sellTotalAmount) > 0 || sellAmount.add(tradeAmount).compareTo(sellTotalAmount) > 0) {
	// break;
	// }
	// sellOrders.add(order);
	// sellAmount = sellAmount.add(tradeAmount);
	// }

	// sellAmount = BigDecimal.ZERO.subtract(sellAmount);
	// BigDecimal sellFee = sellAmount.multiply(currentPrice);

	// // 手续费
	// BigDecimal tradeServiceFee = getTradeServiceFee(sellFee, stock.getTradeCfg());
	// return new TradeDetail(targetValue, sellFee, sellAmount, tradeServiceFee, sellOrders);

	return TradeDetail{}
}

func calculateIncrement(stock *Stock, tradeCfg *TradeCfg, currentPrice float64, smaStrategyConfig SmaStrategy) float64 {

	// 默认步长
	increment := tradeCfg.Increment

	// TODO unimplemented

	// 基于估值水位调整步长
	// increment = calculateIncrementByValuationRatio(tradeCfg, increment)

	// // 注意：股价不是前复权时统计会有问题
	// Date now = new Date();

	// // 计算均线平均价格
	// // 均线&价格
	// final Map<Integer, BigDecimal> average = new HashMap<>(smaStrategyPair.getLeft().size());
	// for (Integer averageVal : smaStrategyPair.getLeft()) {
	//     List<StockPrice> stockPrices = priceManager.getPrices(stock, now, averageVal);
	//     if (null == stockPrices || stockPrices.isEmpty()) {
	//         return increment;
	//     }
	//     BigDecimal totalPrice = BigDecimal.ZERO;
	//     for (StockPrice stockPrice : stockPrices) {
	//         // 没有复权至就直接返回
	//         if (null == stockPrice.getRehabPrice()) {
	//             return increment;
	//         }
	//         totalPrice = totalPrice.add(stockPrice.getRehabPrice());
	//     }
	//     average.put(averageVal, totalPrice.divide(BigDecimal.valueOf(stockPrices.size()), 4, RoundingMode.HALF_UP));
	// }

	// // 比较现价超过均线数量来决定浮动比例
	// // 不用均线定比例的原因：下跌趋势过程中120均价>60均价>30均价>现价，股价突然上升，120均价>现价>60均价>30均价，这时现价低于120均线，大于60均价和30均价，使用120均线不合适。
	// int overNum = -1;
	// for (Integer averageVal : smaStrategyPair.getLeft()) {
	//     BigDecimal linePrice = average.get(averageVal);
	//     if (null != currentPrice && currentPrice.doubleValue() > linePrice.doubleValue()) {
	//         overNum++;
	//     }
	// }
	// //低于所有均线就进入下跌通道了，暂停买入。
	// BigDecimal buyRatio = overNum == -1 ? BigDecimal.ZERO : smaStrategyPair.getRight().get(overNum);
	// log.info("stock:{},overNum-1:{},multiplyVal:{}", stock.getCode(), overNum, buyRatio);
	// increment = increment.multiply(buyRatio);
	// return increment;

	return increment
}
