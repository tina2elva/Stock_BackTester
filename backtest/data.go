package backtest

import (
	"stock/common/types"
	"stock/indicators"
	"stock/portfolio"
	"time"
)

type DataHandler interface {
	OnData(data *types.DataPoint, portfolio *portfolio.Portfolio)
	OnEnd(portfolio *portfolio.Portfolio)
}

// PreprocessData 预处理数据，计算技术指标
func PreprocessData(data []types.Bar) []*types.DataPoint {
	if len(data) < 26 { // 需要至少26个数据点计算MACD
		return nil
	}

	// 计算技术指标
	ma5 := indicators.SMA(data, 5)
	macdValues, err := indicators.MACD(data, 12, 26, 9)
	if err != nil {
		return nil
	}

	// 对齐指标长度
	start := len(data) - len(ma5)
	if len(macdValues) < len(ma5) {
		start = len(data) - len(macdValues)
	}

	// 创建数据点
	points := make([]*types.DataPoint, len(data)-start)
	for i := start; i < len(data); i++ {
		points[i-start] = &types.DataPoint{
			Timestamp: time.Unix(data[i].Time, 0),
			Open:      data[i].Open,
			High:      data[i].High,
			Low:       data[i].Low,
			Close:     data[i].Close,
			Volume:    data[i].Volume,
			Indicators: map[string]float64{
				"MA5":           ma5[i-start],
				"MACD":          macdValues[i-start].MACD,
				"Signal":        macdValues[i-start].Signal,
				"MACDHistogram": macdValues[i-start].Histogram,
			},
		}
	}

	return points
}
