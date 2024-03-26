package domain

import (
	"encoding/json"
	"time"
)

// 交易配置
type TradeCfg struct {
	TradeStrategy  string     `json:"tradeStrategy"`
	Increment      float64    `json:"increment"`
	ServiceFeeRate float64    `json:"serviceFeeRate"`
	MinServiceFee  float64    `json:"minServiceFee"`
	MinTradeAmount float64    `json:"minTradeAmount"`
	Attributes     Attributes `json:"attributes"`
}

// 交易属性
type Attributes struct {
	PerMaxTradePrice            string `json:"perMaxTradePrice"`
	CurrentTargetValue          string `json:"currentTargetValue"`
	PreTargetValue              string `json:"preTargetValue"`
	MaxTargetValue              string `json:"maxTargetValue"`
	CurrentTradePrice           string `json:"currentTradePrice"`
	TargetIndexCode             string `json:"targetIndexCode"`
	CurrentTargetIndexValuation string `json:"currentTargetIndexValuation"`
	CurrentIncrement            string `json:"currentIncrement"`
}

// Stock
type Stock struct {
	ID            int64     `json:"id" gorm:"id"`
	Code          string    `json:"code" gorm:"code"`           // 股票基金编码
	Type          string    `json:"type" gorm:"type"`           // 股票/基金
	Market        string    `json:"market" gorm:"market"`       // 市场：沪、深、港、美、基
	Owner         string    `json:"owner" gorm:"owner"`         // 归属人
	TradeCfg      string    `json:"trade_cfg" gorm:"trade_cfg"` // 交易配置：例如交易费用等
	TotalFee      float64   `json:"total_fee" gorm:"total_fee"` // 总投入金额
	Amount        float64   `json:"amount" gorm:"amount"`       // 持有份额
	LastTradeTime time.Time `json:"last_trade_time" gorm:"last_trade_time"`
	TradeStatus   string    `json:"trade_status" gorm:"trade_status"` // 当前状态：结算中，结算完毕
	Name          string    `json:"name" gorm:"name"`                 // 显示名
	Category      string    `json:"category" gorm:"category"`         // 分类 例如：大盘、小盘、价值、行业、香港、债券、货币等
	SubCategory   string    `json:"sub_category" gorm:"sub_category"` // 子分类 例如：300指数、500指数、养老、医药、传媒等
	Status        int64     `json:"status" gorm:"status"`             // 状态 0失效 1有效
}

// TableName 表名称
func (*Stock) TableName() string {
	return "stock"
}

// GetTradeCfg 获取交易配置
func (s *Stock) GetTradeCfg() TradeCfg {
	var tradeCfg TradeCfg
	err := json.Unmarshal([]byte(s.TradeCfg), &tradeCfg)
	if err != nil {
		panic(err)
	}
	return tradeCfg
}

// GetOwnerStocks 获取owner的证券
func GetOwnerStocks(owner string) ([]Stock, error) {
	var stocks []Stock
	result := DB.Where("owner = ?", owner).Find(&stocks)
	return stocks, result.Error
}

// 获取所以订单 （周定投一年52条记录，直接取全部订单）
func (s *Stock) GetStockOrders() ([]StockOrder, error) {
	var orders []StockOrder
	result := DB.Where("stock_id = ?", s.ID).Find(&orders)
	return orders, result.Error
}
