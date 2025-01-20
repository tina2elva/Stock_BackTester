package strategy

import (
	"log"
	"stock/common"
	"time"
)

type SimpleStrategy struct {
	bought     bool
	buyPrice   float64
	macdFast   int
	macdSlow   int
	macdSignal int
}

func NewSimpleStrategy() *SimpleStrategy {
	return &SimpleStrategy{
		bought:     false,
		macdFast:   12, // 默认快速EMA周期
		macdSlow:   26, // 默认慢速EMA周期
		macdSignal: 9,  // 默认信号线周期
	}
}

func (s *SimpleStrategy) OnData(data *common.DataPoint, portfolio common.Portfolio) {
	// 调试日志
	// 获取指标值
	ma5, hasMA5 := data.Indicators["MA5"]
	macd, hasMACD := data.Indicators["MACD"]
	signal, hasSignal := data.Indicators["Signal"]
	histogram, hasHistogram := data.Indicators["MACDHistogram"]

	// 调试日志
	log.Printf("[%s] Close: %.2f, MA5: %.2f, MACD: %.2f, Signal: %.2f, Bought: %v",
		data.Timestamp.Format("2006-01-02"),
		data.Close,
		ma5,
		macd,
		signal,
		s.bought)

	if !s.bought && hasMA5 && hasMACD && hasSignal && hasHistogram {
		// 买入条件：收盘价高于MA5的98%且MACD上穿信号线
		if data.Close > ma5*0.98 && macd > signal && histogram > 0 {
			log.Printf("Buy signal: Close (%.2f) > MA5 (%.2f), MACD (%.2f) > Signal (%.2f)",
				data.Close, ma5, macd, signal)
			portfolio.Buy(data.Timestamp, data.Close, 100)
			s.bought = true
			s.buyPrice = data.Close
		}
	} else if s.bought && hasMACD && hasSignal && hasHistogram {
		// 调整止损/止盈条件
		currentReturn := (data.Close - s.buyPrice) / s.buyPrice

		// 止损条件：下跌5%时卖出
		if currentReturn < -0.05 {
			log.Printf("Stop loss triggered: %.2f%%", currentReturn*100)
			portfolio.Sell(data.Timestamp, data.Close, 100)
			s.bought = false
		}
		// 止盈条件：上涨10%时卖出
		if currentReturn > 0.10 {
			log.Printf("Take profit triggered: %.2f%%", currentReturn*100)
			portfolio.Sell(data.Timestamp, data.Close, 100)
			s.bought = false
		}
		// MACD下穿信号线时卖出
		if macd < signal && histogram < 0 {
			log.Printf("MACD cross down: MACD (%.2f) < Signal (%.2f)", macd, signal)
			portfolio.Sell(data.Timestamp, data.Close, 100)
			s.bought = false
		}
	}
}

func (s *SimpleStrategy) OnEnd(portfolio common.Portfolio) {
	// Close any open positions
	if s.bought {
		portfolio.Sell(time.Now(), 0, 1)
	}
}
