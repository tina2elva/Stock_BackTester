package strategy

import (
	"stock/common/types"
	"stock/indicators"
	"stock/portfolio"
	"time"
)

type RSIStrategy struct {
	period      int
	overbought  float64
	oversold    float64
	dataBuffer  map[string][]types.Bar         // 按symbol存储数据
	multiPeriod map[int]map[string][]types.Bar // 多周期数据
	logger      types.Logger
}

func NewRSIStrategy(period int, overbought, oversold float64, logger types.Logger) *RSIStrategy {
	return &RSIStrategy{
		period:      period,
		overbought:  overbought,
		oversold:    oversold,
		dataBuffer:  make(map[string][]types.Bar),
		multiPeriod: make(map[int]map[string][]types.Bar),
		logger:      logger,
	}
}

func (s *RSIStrategy) Name() string {
	return "RSI Strategy"
}

func (s *RSIStrategy) Run(data []types.Bar) []types.Signal {
	signals := make([]types.Signal, len(data))

	// 计算RSI值
	rsiValues, err := indicators.RSI(data, s.period)
	if err != nil {
		return signals
	}

	// 跳过前period个点以让指标稳定
	start := s.period
	if start < 1 {
		start = 1
	}

	for i := start; i < len(data); i++ {
		rsi := rsiValues[i]

		// 生成交易信号
		if rsi < s.oversold {
			signals[i] = types.Signal{
				Action: types.ActionBuy,
				Price:  data[i].Close,
				Time:   data[i].Time,
				Qty:    1,
			}
		} else if rsi > s.overbought {
			signals[i] = types.Signal{
				Action: types.ActionSell,
				Price:  data[i].Close,
				Time:   data[i].Time,
				Qty:    1,
			}
		}
	}

	return signals
}

func (s *RSIStrategy) OnStart(portfolio *portfolio.Portfolio) error {
	s.dataBuffer = make(map[string][]types.Bar)
	return nil
}

func (s *RSIStrategy) OnData(data []*types.DataPoint, portfolio *portfolio.Portfolio) error {
	// 处理每个股票的数据点
	for _, dp := range data {
		// 初始化symbol的数据缓冲区
		if _, exists := s.dataBuffer[dp.Symbol]; !exists {
			s.dataBuffer[dp.Symbol] = make([]types.Bar, 0)
		}

		// 添加新数据点到缓冲区
		s.dataBuffer[dp.Symbol] = append(s.dataBuffer[dp.Symbol], types.Bar{
			Time:   dp.Timestamp.Unix(),
			Open:   dp.Open,
			High:   dp.High,
			Low:    dp.Low,
			Close:  dp.Close,
			Volume: dp.Volume,
		})

		// 需要至少period+1个bar来计算RSI
		if len(s.dataBuffer[dp.Symbol]) < s.period+1 {
			continue
		}

		// 计算RSI值
		rsiValues, err := indicators.RSI(s.dataBuffer[dp.Symbol], s.period)
		if err != nil {
			return err
		}
		if len(rsiValues) == 0 {
			continue
		}

		// 使用最新的RSI值
		currentRSI := rsiValues[len(rsiValues)-1]

		// 生成交易信号
		if currentRSI < s.oversold {
			quantity := 1.0
			portfolio.Buy(dp.Symbol, time.Unix(s.dataBuffer[dp.Symbol][len(s.dataBuffer[dp.Symbol])-1].Time, 0), dp.Close, quantity)
		} else if currentRSI > s.overbought {
			quantity := 1.0
			portfolio.Sell(dp.Symbol, time.Unix(s.dataBuffer[dp.Symbol][len(s.dataBuffer[dp.Symbol])-1].Time, 0), dp.Close, quantity)
		}
	}
	return nil
}

func (s *RSIStrategy) OnEnd(portfolio *portfolio.Portfolio, symbol string) error {
	// 关闭所有仓位
	if closer, ok := interface{}(portfolio).(interface{ CloseAllPositions() }); ok {
		closer.CloseAllPositions()
	}
	return nil
}

func (s *RSIStrategy) Calculate(candles []types.Candle) map[string][]float64 {
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

	// Calculate RSI values
	rsiValues, err := indicators.RSI(bars, s.period)
	if err != nil {
		return nil
	}

	// Prepare result
	result := make(map[string][]float64)
	result["RSI"] = make([]float64, len(candles))

	// Fill result array
	for i, v := range rsiValues {
		result["RSI"][i] = v
	}

	return result
}
