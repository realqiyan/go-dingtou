package domain

import (
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
	Snapshot        string    `json:"snapshot" gorm:"snapshot"`                   // 交易快照
}

// TableName 表名称
func (*StockOrder) TableName() string {
	return "stock_order"
}
