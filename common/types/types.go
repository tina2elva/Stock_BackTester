package types

import (
	"errors"
	"time"
)

// OrderType 定义订单类型
type OrderType int

const (
	OrderTypeBuy OrderType = iota
	OrderTypeSell
	OrderTypeMarket
	OrderTypeLimit
	OrderTypeStop
)

// OrderStatus 定义订单状态
type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota
	OrderStatusPending
	OrderStatusFilled
	OrderStatusCanceled
	OrderStatusRejected
)

// Action 交易动作
type Action int

const (
	ActionBuy Action = iota
	ActionSell
	ActionHold
)

// Broker 定义经纪人接口
type Broker interface {
	ExecuteOrder(order *Order) error
	GetAccount() *Account
	CancelOrder(orderID string) error
}

// Order 定义订单结构
type Order struct {
	ID         string
	StrategyID string
	Symbol     string
	Quantity   float64
	Price      float64
	Type       OrderType
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Order方法扩展
func (o *Order) CanExecute() bool {
	return o.Status == OrderStatusNew || o.Status == OrderStatusFilled
}

func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusNew
}

func (o *Order) SetStatus(status OrderStatus) error {
	// 验证状态转换
	switch status {
	case OrderStatusFilled:
		if !o.CanExecute() {
			return ErrInvalidOrderState
		}
	case OrderStatusCanceled:
		if !o.CanCancel() {
			return ErrOrderCannotBeCanceled
		}
	case OrderStatusRejected:
		// 任何状态都可以被拒绝
	default:
		return ErrInvalidOrderState
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// Account 账户信息
type Account struct {
	Cash    float64
	Equity  float64
	Margin  float64
	Balance float64
}

// Bar K线数据
type Bar struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Candle 蜡烛图数据
type Candle struct {
	Timestamp  time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     float64
	Indicators map[string]interface{}
}

// DataPoint 数据点
type DataPoint struct {
	Timestamp  time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	Indicators map[string]float64
}

// Trade 交易记录
type Trade struct {
	ID        string
	Timestamp time.Time
	Price     float64
	Quantity  float64
	Type      Action
	Fee       float64
	Strategy  string
	OrderID   string
}

// MACDValue MACD指标值
type MACDValue struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// RSIValue RSI指标值
type RSIValue struct {
	Fast   float64
	Medium float64
	Slow   float64
}

// Logger 日志接口
type Logger interface {
	LogData(data *DataPoint)
	LogTrade(trade Trade)
	LogEnd(portfolio Portfolio)
}

// FeeConfig 费用配置
type FeeConfig struct {
	StampDuty  float64
	Commission float64
	Fee        float64
	Slippage   float64
	MinLotSize int
}

// FeeCalculator 费用计算接口
type FeeCalculator interface {
	CalculateFee(action Action, price float64, quantity float64) float64
	GetActualPrice(action Action, price float64) float64
}

// Portfolio 投资组合接口
type Portfolio interface {
	GetCash() float64
	GetValue() float64
	GetInitialValue() float64
	AvailableCash() float64
	PositionSize(symbol string) float64
	Transactions() []Trade
	GetPositions() map[string]float64
	GetTrades() []Trade
	Buy(timestamp time.Time, price float64, quantity float64) error
	Sell(timestamp time.Time, price float64, quantity float64) error
}

// Signal 交易信号
type Signal struct {
	Action Action
	Price  float64
	Time   int64
	Qty    int
}

// 定义错误类型
var (
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrInsufficientPosition  = errors.New("insufficient position")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderCannotBeCanceled = errors.New("order cannot be canceled")
	ErrInvalidOrderState     = errors.New("invalid order state transition")
	ErrInvalidQuantity       = errors.New("invalid quantity")
)
