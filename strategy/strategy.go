package strategy

import "stock/common"

type Strategy interface {
	OnData(data *common.DataPoint, portfolio common.Portfolio)
	OnEnd(portfolio common.Portfolio)
	Calculate(candles []common.Candle) map[string][]float64
	Name() string
}
