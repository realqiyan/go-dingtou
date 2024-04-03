package domain

import (
	"dingtou/config"
	"encoding/json"
	"time"
)

// StockOrder
type StockOrder struct {
	ID              uint64    `json:"id,string" gorm:"id"`
	StockId         uint64    `json:"stockId,string" gorm:"stock_id"`
	Code            string    `json:"code" gorm:"code"`
	CreateTime      time.Time `json:"createTime" gorm:"create_time"`
	Type            string    `json:"type" gorm:"type"` // buy:买 sell:卖 bc:补偿
	OutId           string    `json:"outId" gorm:"out_id"`
	TradeTime       time.Time `json:"tradeTime" gorm:"trade_time"`              // 交易日期
	TradeFee        float64   `json:"tradeFee" gorm:"trade_fee"`                // 交易金额
	TradeAmount     float64   `json:"tradeAmount" gorm:"trade_amount"`          // 交易数量
	TradeServiceFee float64   `json:"tradeServiceFee" gorm:"trade_service_fee"` // 交易服务费
	TradeStatus     string    `json:"tradeStatus" gorm:"trade_status"`          // 0:进行中 1:结算完成
	Snapshot        string    `json:"snapshot" gorm:"snapshot"`                 // 交易快照

	Stock              *Stock       `json:"stock" gorm:"-"`              //Stock
	CurrentProfitFee   float64      `json:"currentProfitFee" gorm:"-"`   //当前盈亏金额
	CurrentProfitRatio float64      `json:"currentProfitRatio" gorm:"-"` //当前盈亏比例
	Dependencies       []StockOrder `json:"dependencies" gorm:"-"`       //依赖的订单
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

// 新增 StockOrder
func (s *StockOrder) Create() error {
	result := config.DB.Create(s)
	return result.Error
}

// 更新 StockOrder
func (s *StockOrder) Update() error {
	result := config.DB.Save(s)
	return result.Error
}

// TableName 表名称
func (*StockOrder) TableName() string {
	return "stock_order"
}

// 获取订单快照
func (o *StockOrder) GetSnapshot() OrderSnapshot {
	var orderSnapshot OrderSnapshot
	if len(o.Snapshot) == 0 {
		return orderSnapshot
	}

	var snapshotMap map[string]string = make(map[string]string)
	err := json.Unmarshal([]byte(o.Snapshot), &snapshotMap)
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
