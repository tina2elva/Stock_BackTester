package common

import "time"

type Bar struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type DataPoint struct {
	Timestamp     time.Time
	Open          float64
	High          float64
	Low           float64
	Close         float64
	Volume        float64
	MA5           float64 // 5日移动平均线
	MACD          float64
	Signal        float64
	MACDHistogram float64
}

type Trade struct {
	ID        string
	Timestamp time.Time
	Price     float64
	Quantity  float64
	Type      Action
	Fee       float64
}

type Action int

const (
	ActionBuy Action = iota
	ActionSell
	ActionHold
)

type Signal struct {
	Action Action
	Price  float64
	Time   int64
	Qty    int // 交易数量（股数）
}

type MACDValue struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// 交易费用配置
type FeeConfig struct {
	StampDuty  float64 // 印花税
	Commission float64 // 佣金
	Fee        float64 // 固定手续费
	Slippage   float64 // 滑点
	MinLotSize int     // 最小交易单位
}

// 交易费用计算器接口
type FeeCalculator interface {
	CalculateFee(action Action, price float64, quantity float64) float64
	GetActualPrice(action Action, price float64) float64
}

// 默认交易费用计算器
type DefaultFeeCalculator struct {
	Config FeeConfig
}

func (c *DefaultFeeCalculator) CalculateFee(action Action, price float64, quantity float64) float64 {
	total := price * quantity
	switch action {
	case ActionBuy:
		return total*(c.Config.StampDuty+c.Config.Commission) + c.Config.Fee
	case ActionSell:
		return total*c.Config.Commission + c.Config.Fee
	default:
		return 0
	}
}

func (c *DefaultFeeCalculator) GetActualPrice(action Action, price float64) float64 {
	switch action {
	case ActionBuy:
		return price * (1 + c.Config.Slippage)
	case ActionSell:
		return price * (1 - c.Config.Slippage)
	default:
		return price
	}
}

type Portfolio interface {
	GetCash() float64
	GetValue() float64
	GetInitialValue() float64
	AvailableCash() float64
	PositionSize(symbol string) float64
	Transactions() []Trade
	GetPositions() map[string]float64
	GetTrades() []Trade
	Buy(timestamp time.Time, price float64, quantity float64)
	Sell(timestamp time.Time, price float64, quantity float64)
	SetFeeCalculator(calculator FeeCalculator)
	CalculateTradeCost(action Action, price float64, quantity float64) float64
}
