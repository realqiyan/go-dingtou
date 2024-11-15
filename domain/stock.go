package domain

import (
	"dingtou/config"
	"dingtou/util"
	"encoding/json"
	"fmt"
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
	SellProfitRatio             string `json:"sellProfitRatio"`
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
	ID            uint64    `json:"id,string" gorm:"id"`
	Code          string    `json:"code" gorm:"code"`          // 股票基金编码
	Type          string    `json:"type" gorm:"type"`          // 股票/基金
	Market        string    `json:"market" gorm:"market"`      // 市场：沪、深、港、美、基
	Owner         string    `json:"owner" gorm:"owner"`        // 归属人
	TradeCfg      string    `json:"-" gorm:"trade_cfg"`        // 交易配置：例如交易费用等
	TotalFee      float64   `json:"totalFee" gorm:"total_fee"` // 总投入金额
	Amount        float64   `json:"amount" gorm:"amount"`      // 持有份额
	LastTradeTime time.Time `json:"lastTradeTime" gorm:"last_trade_time"`
	TradeStatus   string    `json:"tradeStatus" gorm:"trade_status"` // 当前状态：结算中，结算完毕
	Name          string    `json:"name" gorm:"name"`                // 显示名
	Category      string    `json:"category" gorm:"category"`        // 分类 例如：大盘、小盘、价值、行业、香港、债券、货币等
	SubCategory   string    `json:"subCategory" gorm:"sub_category"` // 子分类 例如：300指数、500指数、养老、医药、传媒等
	Status        int64     `json:"status" gorm:"status"`            // 状态 0失效 1有效

	TradeCfgStruct *TradeCfg `json:"tradeCfg" gorm:"-"` // 状态 0失效 1有效
}

// StockPrice
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

// GetOwnerStocks 获取owner的证券
func GetOwnerStocks(owner string) ([]Stock, error) {
	var stocks []Stock
	result := config.DB.Where("owner = ?", owner).Find(&stocks)
	return stocks, result.Error
}

// 获取所以订单 （周定投一年52条记录，直接取全部订单）
func (s *Stock) GetStockOrders() ([]StockOrder, error) {
	var orders []StockOrder
	result := config.DB.Where("stock_id = ?", s.ID).Find(&orders)
	return orders, result.Error
}

// 获取待处理订单
func (s *Stock) GetStockWaitProcessOrders() ([]StockOrder, error) {
	var orders []StockOrder
	result := config.DB.Where("stock_id = ? and trade_status = ?", s.ID, util.PROCESSING).Find(&orders)
	return orders, result.Error
}

// 创建Stock
func (s *Stock) Create() error {
	result := config.DB.Create(s)
	return result.Error
}

// 更新Stock
func (s *Stock) Update() error {
	result := config.DB.Save(s)
	return result.Error
}

// TableName 表名称
func (*Stock) TableName() string {
	return "stock"
}

// GetTradeCfg 获取交易配置
func (s *Stock) GetTradeCfg() *TradeCfg {
	if s.TradeCfgStruct != nil {
		return s.TradeCfgStruct
	}
	var tradeCfg TradeCfg
	err := json.Unmarshal([]byte(s.TradeCfg), &tradeCfg)
	if err != nil {
		panic(err)
	}
	s.TradeCfgStruct = &tradeCfg
	return s.TradeCfgStruct
}

// 生成订单
func (s *Stock) Conform(order *StockOrder) error {
	now := time.Now()
	order.Stock = s
	order.Code = s.Code
	order.StockId = s.ID
	order.TradeStatus = util.PROCESSING
	s.TradeStatus = util.PROCESSING
	order.CreateTime = now

	historyOrders, _ := s.GetStockOrders()
	tradeDetail := CalculateConform(s, historyOrders, now)

	order.TradeAmount = tradeDetail.TradeAmount
	order.TradeFee = tradeDetail.TradeFee
	order.TradeServiceFee = tradeDetail.TradeServiceFee
	order.TradeTime = now

	var tradeType string
	if tradeDetail.TradeFee >= 0 {
		tradeType = util.BUY
	} else {
		tradeType = util.SELL
	}
	order.Type = tradeType
	order.OutId = buildOutId(tradeType, now, s)

	order.Dependencies = tradeDetail.SellOrders

	// 交易快照
	snapshot := make(map[string]string)
	tradeCfgByte, _ := json.Marshal(s.TradeCfgStruct)
	snapshot["tradeCfg"] = string(tradeCfgByte)

	var outIds []string
	for _, order := range tradeDetail.SellOrders {
		outIds = append(outIds, order.OutId)
	}
	buyOrderOutIdsByte, _ := json.Marshal(outIds)
	snapshot["buyOrderOutIds"] = string(buyOrderOutIdsByte)

	snapshotByte, _ := json.Marshal(snapshot)
	order.Snapshot = string(snapshotByte)

	return nil
}

// 结算订单
func (s *Stock) Settlement(order *StockOrder) error {
	order.Stock = s
	tradeDetail, _ := CalculateSettlement(order)
	order.TradeFee = tradeDetail.TradeFee
	order.TradeAmount = tradeDetail.TradeAmount
	order.TradeServiceFee = tradeDetail.TradeServiceFee

	amount := util.FloatAdd(s.Amount, order.TradeAmount)
	totalTradeFee := util.FloatAdd(s.TotalFee, order.TradeFee)

	s.Amount = amount
	s.TotalFee = totalTradeFee
	order.TradeStatus = util.DONE
	s.TradeStatus = util.DONE
	s.LastTradeTime = order.TradeTime
	return nil
}

func buildOutId(prefix string, now time.Time, s *Stock) string {
	str := now.Format("20060102")
	return fmt.Sprintf("%s_%s_%s%s_%s", s.Owner, prefix, s.Market, s.Code, str)
}
