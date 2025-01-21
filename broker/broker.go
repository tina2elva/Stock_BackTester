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
	// 获取日志记录器
	Logger() common.Logger
}

type SimulatedBroker struct {
	feeRate float64
	logger  common.Logger
}

func NewSimulatedBroker(feeRate float64, logger common.Logger) *SimulatedBroker {
	return &SimulatedBroker{
		feeRate: feeRate,
		logger:  logger,
	}
}

func (b *SimulatedBroker) Logger() common.Logger {
	return b.logger
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
