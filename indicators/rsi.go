package indicators

import (
	"errors"
	"stock/common"
)

// RSI 计算相对强弱指数
func RSI(data []common.Bar, period int) ([]float64, error) {
	if len(data) < period+1 {
		return nil, errors.New("not enough data points to calculate RSI")
	}

	gains := make([]float64, len(data))
	losses := make([]float64, len(data))

	// 计算每日价格变化
	for i := 1; i < len(data); i++ {
		change := data[i].Close - data[i-1].Close
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	// 计算平均收益和平均损失
	avgGain := make([]float64, len(data))
	avgLoss := make([]float64, len(data))

	// 计算第一个平均收益/损失
	var sumGain, sumLoss float64
	for i := 1; i <= period; i++ {
		sumGain += gains[i]
		sumLoss += losses[i]
	}
	avgGain[period] = sumGain / float64(period)
	avgLoss[period] = sumLoss / float64(period)

	// 计算后续的平滑平均收益/损失
	for i := period + 1; i < len(data); i++ {
		avgGain[i] = (avgGain[i-1]*(float64(period)-1) + gains[i]) / float64(period)
		avgLoss[i] = (avgLoss[i-1]*(float64(period)-1) + losses[i]) / float64(period)
	}

	// 计算RSI
	rsi := make([]float64, len(data))
	for i := period; i < len(data); i++ {
		if avgLoss[i] == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain[i] / avgLoss[i]
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi[period:], nil
}
