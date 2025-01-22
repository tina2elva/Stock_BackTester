package indicators

import (
	"errors"
	"math"
	"stock/common"
)

// MACD calculates Moving Average Convergence Divergence
func MACD(bars []common.Bar, fastPeriod, slowPeriod, signalPeriod int) ([]common.MACDValue, error) {
	if len(bars) < slowPeriod {
		return nil, errors.New("not enough data points")
	}

	// Calculate fast and slow EMAs
	fastEMA := EMA(bars, fastPeriod)
	slowEMA := EMA(bars, slowPeriod)

	if fastEMA == nil || slowEMA == nil {
		return nil, errors.New("failed to calculate EMA")
	}

	// Calculate MACD line
	macdLine := make([]float64, len(bars))
	for i := slowPeriod - 1; i < len(bars); i++ {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}

	// Calculate Signal line (EMA of MACD line)
	signalLine := EMAForMACD(macdLine[slowPeriod-1:], signalPeriod)
	if signalLine == nil {
		return nil, errors.New("failed to calculate signal line")
	}

	// Pad signalLine with NaN values to match original length
	fullSignalLine := make([]float64, len(bars))
	for i := 0; i < slowPeriod-1; i++ {
		fullSignalLine[i] = math.NaN()
	}
	copy(fullSignalLine[slowPeriod-1:], signalLine)
	signalLine = fullSignalLine

	// Calculate MACD histogram
	histogram := make([]float64, len(bars))
	for i := range histogram {
		histogram[i] = macdLine[i] - signalLine[i]
	}

	// Prepare result
	result := make([]common.MACDValue, len(bars))
	for i := range result {
		result[i] = common.MACDValue{
			MACD:      macdLine[i],
			Signal:    signalLine[i],
			Histogram: histogram[i],
		}
	}

	return result, nil
}

// EMA calculates Exponential Moving Average
func EMA(bars []common.Bar, period int) []float64 {
	if len(bars) < period {
		return nil
	}

	ema := make([]float64, len(bars))
	k := 2.0 / float64(period+1)

	// Calculate first EMA as SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += bars[i].Close
	}
	ema[period-1] = sum / float64(period)

	// Calculate remaining EMAs
	for i := period; i < len(bars); i++ {
		ema[i] = bars[i].Close*k + ema[i-1]*(1-k)
	}

	return ema
}

// SMA calculates Simple Moving Average
func SMA(bars []common.Bar, period int) []float64 {
	if len(bars) < period {
		return nil
	}

	sma := make([]float64, len(bars))

	for i := period - 1; i < len(bars); i++ {
		var sum float64
		for j := 0; j < period; j++ {
			sum += bars[i-j].Close
		}
		sma[i] = sum / float64(period)
	}

	return sma
}

// EMAForMACD calculates EMA for MACD signal line
func EMAForMACD(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}

	ema := make([]float64, len(values))
	k := 2.0 / float64(period+1)

	// Calculate first EMA as SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	ema[period-1] = sum / float64(period)

	// Calculate remaining EMAs
	for i := period; i < len(values); i++ {
		ema[i] = values[i]*k + ema[i-1]*(1-k)
	}

	return ema
}
