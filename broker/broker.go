package broker

import (
	"stock/common"
	"time"
)

type Broker interface {
	// 执行买入操作
	Buy(timestamp time.Time, price float64, quantity float64) error
	// 执行卖出操作
	Sell(timestamp time.Time, price float64, quantity float64) error
	// 计算交易成本
	CalculateTradeCost(action common.Action, price float64, quantity float64) float64
}

type SimulatedBroker struct {
	feeRate float64
}

func NewSimulatedBroker(feeRate float64) *SimulatedBroker {
	return &SimulatedBroker{
		feeRate: feeRate,
	}
}

func (b *SimulatedBroker) Buy(timestamp time.Time, price float64, quantity float64) error {
	return nil
}

func (b *SimulatedBroker) Sell(timestamp time.Time, price float64, quantity float64) error {
	return nil
}

func (b *SimulatedBroker) CalculateTradeCost(action common.Action, price float64, quantity float64) float64 {
	return price * quantity * b.feeRate
}
