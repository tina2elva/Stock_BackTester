package common

import (
	"errors"
	"time"
)

type Bar struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type Candle struct {
	Timestamp  time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     float64
	Indicators map[string]interface{}
}

type DataPoint struct {
	Timestamp  time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	Indicators map[string]float64
}

type OrderStatus int

const (
	OrderPending OrderStatus = iota
	OrderFilled
	OrderCancelled
	OrderRejected
)

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

type Action int

const (
	ActionBuy Action = iota
	ActionSell
	ActionHold
)

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInsufficientPosition = errors.New("insufficient position")
)

type Signal struct {
	Action Action
	Price  float64
	Time   int64
	Qty    int
}

type MACDValue struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

type RSIValue struct {
	Fast   float64
	Medium float64
	Slow   float64
}

type Logger interface {
	LogData(data *DataPoint)
	LogTrade(trade Trade)
	LogEnd(portfolio Portfolio)
}

type FeeConfig struct {
	StampDuty  float64
	Commission float64
	Fee        float64
	Slippage   float64
	MinLotSize int
}

type FeeCalculator interface {
	CalculateFee(action Action, price float64, quantity float64) float64
	GetActualPrice(action Action, price float64) float64
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
	Buy(timestamp time.Time, price float64, quantity float64) error
	Sell(timestamp time.Time, price float64, quantity float64) error
}
