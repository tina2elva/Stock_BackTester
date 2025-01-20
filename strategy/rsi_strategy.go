package strategy

import (
	"log"
	"stock/common"
	"stock/indicators"
)

type RSIStrategy struct {
	period     int
	overbought float64
	oversold   float64
	dataBuffer []common.Bar
}

func NewRSIStrategy(period int, overbought, oversold float64) *RSIStrategy {
	// 优化RSI参数
	return &RSIStrategy{
		period:     9,  // 缩短周期以提高灵敏度
		overbought: 65, // 降低超买阈值
		oversold:   35, // 提高超卖阈值
	}
}

func (s *RSIStrategy) Name() string {
	return "RSI Strategy"
}

func (s *RSIStrategy) Run(data []common.Bar) []common.Signal {
	signals := make([]common.Signal, len(data))

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

		// 生成交易信号，增加过滤条件
		if rsi < s.oversold && rsi > 30 { // 在超卖区域但不过度
			signals[i] = common.Signal{
				Action: common.ActionBuy,
				Price:  data[i].Close,
				Time:   data[i].Time,
				Qty:    2, // 增加交易单位
			}
		} else if rsi > s.overbought && rsi < 70 { // 在超买区域但不过度
			signals[i] = common.Signal{
				Action: common.ActionSell,
				Price:  data[i].Close,
				Time:   data[i].Time,
				Qty:    2, // 增加交易单位
			}
		}
	}

	return signals
}

func (s *RSIStrategy) OnData(data *common.DataPoint, portfolio common.Portfolio) {
	// 添加新数据点到缓冲区
	s.dataBuffer = append(s.dataBuffer, common.Bar{
		Time:   data.Timestamp.Unix(),
		Open:   data.Open,
		High:   data.High,
		Low:    data.Low,
		Close:  data.Close,
		Volume: data.Volume,
	})

	// 需要至少period+1个bar来计算RSI
	if len(s.dataBuffer) < s.period+1 {
		return
	}

	// 计算RSI值
	rsiValues, err := indicators.RSI(s.dataBuffer, s.period)
	if err != nil {
		return
	}
	if len(rsiValues) == 0 {
		return
	}

	// 使用最新的RSI值
	currentRSI := rsiValues[len(rsiValues)-1]

	// 生成交易信号
	if currentRSI < s.oversold {
		quantity := 1.0 // 默认交易1单位
		portfolio.Buy(data.Timestamp, data.Close, quantity)
		log.Printf("[交易日志] 策略: %s | 买入 | 时间: %s | 价格: %.2f | 数量: %.2f | 可用资金: %.2f | 持仓: %.2f",
			s.Name(),
			data.Timestamp.Format("2006-01-02 15:04:05"),
			data.Close,
			quantity,
			portfolio.AvailableCash(),
			portfolio.PositionSize("asset"))
	} else if currentRSI > s.overbought {
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

func (s *RSIStrategy) OnEnd(portfolio common.Portfolio) {
	// 关闭所有仓位
	if closer, ok := portfolio.(interface{ CloseAllPositions() }); ok {
		closer.CloseAllPositions()
	}
}

// Calculate returns RSI indicator values
func (s *RSIStrategy) Calculate(candles []common.Candle) map[string][]float64 {
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
