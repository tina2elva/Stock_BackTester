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
	log.Printf("[%s] Close: %.2f, MA5: %.2f, MACD: %.2f, Signal: %.2f, Bought: %v",
		data.Timestamp.Format("2006-01-02"),
		data.Close,
		data.MA5,
		data.MACD,
		data.Signal,
		s.bought)

	if !s.bought {
		// 买入条件：收盘价高于MA5的98%且MACD上穿信号线
		if data.Close > data.MA5*0.98 && data.MACD > data.Signal && data.MACDHistogram > 0 {
			log.Printf("Buy signal: Close (%.2f) > MA5 (%.2f), MACD (%.2f) > Signal (%.2f)",
				data.Close, data.MA5, data.MACD, data.Signal)
			portfolio.Buy(data.Timestamp, data.Close, 100)
			s.bought = true
			s.buyPrice = data.Close
		}
	} else if s.bought {
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
		if data.MACD < data.Signal && data.MACDHistogram < 0 {
			log.Printf("MACD cross down: MACD (%.2f) < Signal (%.2f)", data.MACD, data.Signal)
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
