package strategy

import (
	"fmt"
	"math"
	"stock/common/types"
	"stock/indicators"
	"stock/portfolio"
	"time"
)

type MACDStrategy struct {
	Periods         []int
	fastPeriod      int
	slowPeriod      int
	signalPeriod    int
	prevMACD        float64
	prevSignal      float64
	dataBuffer      []types.Bar
	multiPeriodData map[int][]types.Bar // 存储不同周期的数据
}

func NewMACDStrategy(fast, slow, signal int, periods []int) *MACDStrategy {
	return &MACDStrategy{
		fastPeriod:   fast,
		slowPeriod:   slow,
		signalPeriod: signal,
		Periods:      periods,
		dataBuffer:   make([]types.Bar, 0),
		prevMACD:     math.NaN(),
		prevSignal:   math.NaN(),
	}
}

func convertToCommonBars(bars []types.Bar) []types.Bar {
	commonBars := make([]types.Bar, len(bars))
	for i, bar := range bars {
		commonBars[i] = types.Bar{
			Time:   bar.Time,
			Open:   bar.Open,
			High:   bar.High,
			Low:    bar.Low,
			Close:  bar.Close,
			Volume: bar.Volume,
		}
	}
	return commonBars
}

func (s *MACDStrategy) Name() string {
	return "MACD Strategy"
}

func (s *MACDStrategy) Run(data []types.Bar) []types.Signal {
	signals := make([]types.Signal, len(data))

	// Need at least slowPeriod + signalPeriod bars to calculate MACD
	if len(data) < s.slowPeriod+s.signalPeriod {
		return signals
	}

	// Calculate MACD values
	macdValues, err := indicators.MACD(data, s.fastPeriod, s.slowPeriod, s.signalPeriod)
	if err != nil {
		return signals
	}

	// Skip first slowPeriod + signalPeriod points to allow indicators to stabilize
	start := s.slowPeriod + s.signalPeriod
	if start < 1 {
		start = 1
	}

	for i := start; i < len(data); i++ {
		macd := macdValues[i].MACD
		signal := macdValues[i].Signal

		// Generate buy/sell signals based on MACD crossover
		// Only generate signals when we have valid MACD and Signal values
		if !math.IsNaN(macd) && !math.IsNaN(signal) {
			if !math.IsNaN(s.prevMACD) && !math.IsNaN(s.prevSignal) {
				if s.prevMACD < s.prevSignal && macd > signal {
					signals[i] = types.Signal{
						Action: types.ActionBuy,
						Price:  data[i].Close,
						Time:   data[i].Time,
						Qty:    1, // 默认交易1单位
					}
				} else if s.prevMACD > s.prevSignal && macd < signal {
					signals[i] = types.Signal{
						Action: types.ActionSell,
						Price:  data[i].Close,
						Time:   data[i].Time,
						Qty:    1, // 默认交易1单位
					}
				}
			}
			s.prevMACD = macd
			s.prevSignal = signal
		}
	}

	return signals
}

// OnStart initializes the strategy
func (s *MACDStrategy) OnStart(portfolio *portfolio.Portfolio) error {
	s.dataBuffer = make([]types.Bar, 0)
	return nil
}

// OnData handles new market data
func (s *MACDStrategy) OnData(data []*types.DataPoint, portfolio *portfolio.Portfolio) error {
	// Process each stock's data point
	for _, dp := range data {
		// Add new data point to buffer
		s.dataBuffer = append(s.dataBuffer, types.Bar{
			Time:   dp.Timestamp.Unix(),
			Open:   dp.Open,
			High:   dp.High,
			Low:    dp.Low,
			Close:  dp.Close,
			Volume: dp.Volume,
		})

		// Need at least slowPeriod + signalPeriod bars to calculate MACD
		if len(s.dataBuffer) < s.slowPeriod+s.signalPeriod {
			continue
		}

		// Calculate MACD values
		macdValues, err := indicators.MACD(s.dataBuffer, s.fastPeriod, s.slowPeriod, s.signalPeriod)
		if err != nil {
			return err
		}
		if len(macdValues) == 0 {
			continue
		}

		// Use the last two MACD values to detect crossovers
		if len(macdValues) < 2 {
			continue
		}

		prevMACD := macdValues[len(macdValues)-2].MACD
		prevSignal := macdValues[len(macdValues)-2].Signal
		currentMACD := macdValues[len(macdValues)-1].MACD
		currentSignal := macdValues[len(macdValues)-1].Signal

		// Generate buy/sell signals based on MACD crossover
		if len(s.dataBuffer) == 0 {
			continue
		}
		latestBar := s.dataBuffer[len(s.dataBuffer)-1]
		if prevMACD < prevSignal && currentMACD > currentSignal {
			quantity := 1.0 // default quantity
			portfolio.Buy(dp.Symbol, time.Unix(latestBar.Time, 0), latestBar.Close, quantity)
		} else if prevMACD > prevSignal && currentMACD < currentSignal {
			quantity := 1.0 // 默认交易1单位
			portfolio.Sell(dp.Symbol, time.Unix(latestBar.Time, 0), latestBar.Close, quantity)
		}
	}
	return nil
}

// OnEnd handles backtest completion
func (s *MACDStrategy) OnEnd(portfolio *portfolio.Portfolio, symbol string) error {
	// Close all positions
	if closer, ok := interface{}(portfolio).(interface{ CloseAllPositions() }); ok {
		closer.CloseAllPositions()
	}
	return nil
}

// Calculate returns MACD indicator values
func (s *MACDStrategy) Calculate(candles []types.Candle) map[string][]float64 {
	// Convert candles to bars
	bars := make([]types.Bar, len(candles))
	for i, c := range candles {
		bars[i] = types.Bar{
			Time:   c.Timestamp.Unix(),
			Open:   c.Open,
			High:   c.High,
			Low:    c.Low,
			Close:  c.Close,
			Volume: c.Volume,
		}
	}

	// Initialize result map
	result := make(map[string][]float64)
	result["MACD"] = make([]float64, len(candles))
	result["Signal"] = make([]float64, len(candles))
	result["Histogram"] = make([]float64, len(candles))

	// Calculate base MACD values
	macdValues, err := indicators.MACD(bars, s.fastPeriod, s.slowPeriod, s.signalPeriod)
	if err != nil {
		return nil
	}

	// Fill base MACD values
	for i, v := range macdValues {
		result["MACD"][i] = v.MACD
		result["Signal"][i] = v.Signal
		result["Histogram"][i] = v.Histogram
	}

	// Calculate multi-period MACD values
	for _, period := range s.Periods {
		if period <= 0 {
			continue
		}

		// Resample data to target period
		resampledBars := s.resampleBars(bars, period)
		if len(resampledBars) == 0 {
			continue
		}

		// Calculate MACD for resampled data
		resampledMACD, err := indicators.MACD(resampledBars, s.fastPeriod, s.slowPeriod, s.signalPeriod)
		if err != nil {
			continue
		}

		// Map resampled MACD values back to original timeframe
		mappedMACD := s.mapResampledValues(resampledMACD, len(bars), period)

		// Add to result with period suffix
		result[fmt.Sprintf("MACD_%d", period)] = mappedMACD["MACD"]
		result[fmt.Sprintf("Signal_%d", period)] = mappedMACD["Signal"]
		result[fmt.Sprintf("Histogram_%d", period)] = mappedMACD["Histogram"]
	}

	return result
}

// resampleBars resamples bars to target period
func (s *MACDStrategy) resampleBars(bars []types.Bar, period int) []types.Bar {
	if len(bars) == 0 || period <= 0 {
		return nil
	}

	resampled := make([]types.Bar, 0)
	currentBar := types.Bar{
		Time:   bars[0].Time,
		Open:   bars[0].Open,
		High:   bars[0].High,
		Low:    bars[0].Low,
		Close:  bars[0].Close,
		Volume: bars[0].Volume,
	}

	for i := 1; i < len(bars); i++ {
		if i%period == 0 {
			resampled = append(resampled, currentBar)
			currentBar = types.Bar{
				Time:   bars[i].Time,
				Open:   bars[i].Open,
				High:   bars[i].High,
				Low:    bars[i].Low,
				Close:  bars[i].Close,
				Volume: bars[i].Volume,
			}
		} else {
			currentBar.High = math.Max(currentBar.High, bars[i].High)
			currentBar.Low = math.Min(currentBar.Low, bars[i].Low)
			currentBar.Close = bars[i].Close
			currentBar.Volume += bars[i].Volume
		}
	}

	// Add last bar
	if len(resampled) == 0 || resampled[len(resampled)-1].Time != currentBar.Time {
		resampled = append(resampled, currentBar)
	}

	return resampled
}

// mapResampledValues maps resampled indicator values back to original timeframe
func (s *MACDStrategy) mapResampledValues(resampledMACD []types.MACDValue, originalLength, period int) map[string][]float64 {
	result := make(map[string][]float64)
	result["MACD"] = make([]float64, originalLength)
	result["Signal"] = make([]float64, originalLength)
	result["Histogram"] = make([]float64, originalLength)

	resampledIndex := 0
	for i := 0; i < originalLength; i++ {
		if i > 0 && i%period == 0 {
			resampledIndex++
		}
		if resampledIndex >= len(resampledMACD) {
			continue
		}

		result["MACD"][i] = resampledMACD[resampledIndex].MACD
		result["Signal"][i] = resampledMACD[resampledIndex].Signal
		result["Histogram"][i] = resampledMACD[resampledIndex].Histogram
	}

	return result
}
