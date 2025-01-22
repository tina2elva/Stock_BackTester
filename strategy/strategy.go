package strategy

import (
	"stock/common/types"
	"stock/portfolio"
)

type Strategy interface {
	OnStart(portfolio *portfolio.Portfolio) error
	OnData(data []*types.DataPoint, portfolio *portfolio.Portfolio) error
	OnEnd(portfolio *portfolio.Portfolio, symbol string) error
	Calculate(candles []types.Candle) map[string][]float64
	Name() string
}
