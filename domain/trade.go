package domain

import (
	"encoding/json"
	"time"
)

// StockOrder
type StockOrder struct {
	ID              int64     `json:"id" gorm:"id"`
	StockId         int64     `json:"stock_id" gorm:"stock_id"`
	Code            string    `json:"code" gorm:"code"`
	CreateTime      time.Time `json:"create_time" gorm:"create_time"`
	Type            string    `json:"type" gorm:"type"` // buy:买 sell:卖 bc:补偿
	OutId           string    `json:"out_id" gorm:"out_id"`
	TradeTime       time.Time `json:"trade_time" gorm:"trade_time"`               // 交易日期
	TradeFee        float64   `json:"trade_fee" gorm:"trade_fee"`                 // 交易金额
	TradeAmount     float64   `json:"trade_amount" gorm:"trade_amount"`           // 交易数量
	TradeServiceFee float64   `json:"trade_service_fee" gorm:"trade_service_fee"` // 交易服务费
	TradeStatus     string    `json:"trade_status" gorm:"trade_status"`           // 0:进行中 1:结算完成
	Snapshot        []byte    `json:"snapshot" gorm:"snapshot"`                   // 交易快照

	CurrentProfitFee   float64 //当前盈亏金额
	CurrentProfitRatio float64 //当前盈亏比例
}

// TradeDetail
type TradeDetail struct {
	/**
	 * 目标金额
	 */
	targetValue float64
	/**
	 * 交易金额
	 */
	TradeFee float64
	/**
	 * 交易份额
	 */
	TradeAmount float64
	/**
	 * 交易手续费
	 */
	TradeServiceFee float64

	/**
	 * 卖出的订单
	 */
	SellOrders []StockOrder
}

// 订单快照
type OrderSnapshot struct {
	TradeCfg       TradeCfg `json:"tradeCfg"`       //交易配置快照
	BuyOrderOutIds []string `json:"buyOrderOutIds"` //卖出依赖的订单
}

// 获取订单快照
func (o *StockOrder) GetSnapshot() OrderSnapshot {
	var orderSnapshot OrderSnapshot
	if len(o.Snapshot) == 0 {
		return orderSnapshot
	}

	var snapshotMap map[string]string = make(map[string]string)
	err := json.Unmarshal(o.Snapshot, &snapshotMap)
	if err != nil {
		panic(err)
	}

	// 交易快照
	tradeCfgStr, exist := snapshotMap["tradeCfg"]
	if exist {
		var tradeCfg TradeCfg
		err := json.Unmarshal([]byte(tradeCfgStr), &tradeCfg)
		if err != nil {
			panic(err)
		}
		orderSnapshot.TradeCfg = tradeCfg
	}

	// 依赖的订单唯一ID
	buyOrderOutIdsStr, exist := snapshotMap["buyOrderOutIds"]
	if exist {
		var buyOrderOutIds []string
		err := json.Unmarshal([]byte(buyOrderOutIdsStr), &buyOrderOutIds)
		if err != nil {
			panic(err)
		}
		orderSnapshot.BuyOrderOutIds = buyOrderOutIds
	}

	return orderSnapshot
}

// TableName 表名称
func (*StockOrder) TableName() string {
	return "stock_order"
}
