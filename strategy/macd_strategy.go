package strategy

import (
	"log"
	"stock/common"
	"stock/indicators"
)

type MACDStrategy struct {
	fastPeriod   int
	slowPeriod   int
	signalPeriod int
	prevMACD     float64
	prevSignal   float64
	dataBuffer   []common.Bar
}

func NewMACDStrategy(fast, slow, signal int) *MACDStrategy {
	return &MACDStrategy{
		fastPeriod:   fast,
		slowPeriod:   slow,
		signalPeriod: signal,
	}
}

func (s *MACDStrategy) Name() string {
	return "MACD Strategy"
}

func (s *MACDStrategy) Run(data []common.Bar) []common.Signal {
	signals := make([]common.Signal, len(data))

	// Calculate MACD values
	macdValues, err := indicators.MACD(data, s.fastPeriod, s.slowPeriod, s.signalPeriod)
	if err != nil {
		return signals
	}

	// Skip first signalPeriod points to allow indicators to stabilize
	start := s.signalPeriod
	if start < 1 {
		start = 1
	}

	for i := start; i < len(data); i++ {
		macd := macdValues[i].MACD
		signal := macdValues[i].Signal

		// Generate buy/sell signals based on MACD crossover
		if s.prevMACD < s.prevSignal && macd > signal {
			signals[i] = common.Signal{
				Action: common.ActionBuy,
				Price:  data[i].Close,
				Time:   data[i].Time,
				Qty:    1, // 默认交易1单位
			}
		} else if s.prevMACD > s.prevSignal && macd < signal {
			signals[i] = common.Signal{
				Action: common.ActionSell,
				Price:  data[i].Close,
				Time:   data[i].Time,
				Qty:    1, // 默认交易1单位
			}
		}

		s.prevMACD = macd
		s.prevSignal = signal
	}

	return signals
}

// OnData handles new market data
func (s *MACDStrategy) OnData(data *common.DataPoint, portfolio common.Portfolio) {
	// Add new data point to buffer
	s.dataBuffer = append(s.dataBuffer, common.Bar{
		Time:   data.Timestamp.Unix(),
		Open:   data.Open,
		High:   data.High,
		Low:    data.Low,
		Close:  data.Close,
		Volume: data.Volume,
	})

	// Need at least slowPeriod + signalPeriod bars to calculate MACD
	if len(s.dataBuffer) < s.slowPeriod+s.signalPeriod {
		return
	}

	// Calculate MACD values
	macdValues, err := indicators.MACD(s.dataBuffer, s.fastPeriod, s.slowPeriod, s.signalPeriod)
	if err != nil {
		return
	}
	if len(macdValues) == 0 {
		return
	}

	// Use the last two MACD values to detect crossovers
	if len(macdValues) < 2 {
		return
	}

	prevMACD := macdValues[len(macdValues)-2].MACD
	prevSignal := macdValues[len(macdValues)-2].Signal
	currentMACD := macdValues[len(macdValues)-1].MACD
	currentSignal := macdValues[len(macdValues)-1].Signal

	// Generate buy/sell signals based on MACD crossover
	if prevMACD < prevSignal && currentMACD > currentSignal {
		quantity := 1.0 // 默认交易1单位
		portfolio.Buy(data.Timestamp, data.Close, quantity)
		log.Printf("[交易日志] 策略: %s | 买入 | 时间: %s | 价格: %.2f | 数量: %.2f | 可用资金: %.2f | 持仓: %.2f",
			s.Name(),
			data.Timestamp.Format("2006-01-02 15:04:05"),
			data.Close,
			quantity,
			portfolio.AvailableCash(),
			portfolio.PositionSize("asset"))
	} else if prevMACD > prevSignal && currentMACD < currentSignal {
		quantity := 1.0 // 默认交易1单位
		portfolio.Sell(data.Timestamp, data.Close, quantity)
		log.Printf("[交易日志] 策略: %s | 卖出 | 时间: %s | 价格: %.2f | 数量: %.2f | 可用资金: %.2f | 持仓: %.2f",
			s.Name(),
			data.Timestamp.Format("2006-01-02 15:04:05"),
			data.Close,
			quantity,
			portfolio.AvailableCash(),
			portfolio.PositionSize("asset"))
	}
}

// OnEnd handles backtest completion
func (s *MACDStrategy) OnEnd(portfolio common.Portfolio) {
	// Close all positions
	if closer, ok := portfolio.(interface{ CloseAllPositions() }); ok {
		closer.CloseAllPositions()
	}
}

// Calculate returns MACD indicator values
func (s *MACDStrategy) Calculate(candles []common.Candle) map[string][]float64 {
	// Convert candles to bars
	bars := make([]common.Bar, len(candles))
	for i, c := range candles {
		bars[i] = common.Bar{
			Time:   c.Timestamp.Unix(),
			Open:   c.Open,
			High:   c.High,
			Low:    c.Low,
			Close:  c.Close,
			Volume: c.Volume,
		}
	}

	// Calculate MACD values
	macdValues, err := indicators.MACD(bars, s.fastPeriod, s.slowPeriod, s.signalPeriod)
	if err != nil {
		return nil
	}

	// Prepare result
	result := make(map[string][]float64)
	result["MACD"] = make([]float64, len(candles))
	result["Signal"] = make([]float64, len(candles))
	result["Histogram"] = make([]float64, len(candles))

	// Fill result arrays
	for i, v := range macdValues {
		result["MACD"][i] = v.MACD
		result["Signal"][i] = v.Signal
		result["Histogram"][i] = v.Histogram
	}

	return result
}
