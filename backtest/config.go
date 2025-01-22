package backtest

import (
	"time"

	"stock/common/types"
	"stock/datasource"
	"stock/strategy"
)

// Config 回测配置
type Config struct {
	// 数据源配置
	DataSource datasource.DataSource
	Symbol     string
	StartDate  time.Time
	EndDate    time.Time

	// 策略配置
	Strategies []strategy.Strategy

	// 资金配置
	InitialCash float64

	// 费用配置
	Commission  float64
	StampDuty   float64
	TransferFee float64
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	return &Config{
		Commission:  0.0003,  // 佣金：万分之三
		StampDuty:   0.001,   // 印花税：千分之一
		TransferFee: 0.00002, // 过户费：万分之0.2
	}
}

// DefaultFeeConfig 默认费用配置
var DefaultFeeConfig = Config{
	Commission:  0.0003,  // 佣金：万分之三
	StampDuty:   0.001,   // 印花税：千分之一
	TransferFee: 0.00002, // 过户费：万分之0.2
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.DataSource == nil {
		return types.ErrInvalidDataSource
	}
	if c.Symbol == "" {
		return types.ErrInvalidSymbol
	}
	if c.StartDate.IsZero() || c.EndDate.IsZero() {
		return types.ErrInvalidDateRange
	}
	if c.InitialCash <= 0 {
		return types.ErrInvalidInitialCash
	}
	if len(c.Strategies) == 0 {
		return types.ErrNoStrategy
	}
	return nil
}
