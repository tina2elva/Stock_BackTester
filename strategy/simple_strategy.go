package strategy

import (
	"stock/common/types"
	"time"
)

type SimpleStrategy struct {
	bought     bool
	buyPrice   float64
	macdFast   int
	macdSlow   int
	macdSignal int
	logger     types.Logger
}

func (s *SimpleStrategy) Name() string {
	return "简单策略"
}

func NewSimpleStrategy(logger types.Logger) *SimpleStrategy {
	return &SimpleStrategy{
		bought:     false,
		macdFast:   12, // 默认快速EMA周期
		macdSlow:   26, // 默认慢速EMA周期
		macdSignal: 9,  // 默认信号线周期
		logger:     logger,
	}
}

func (s *SimpleStrategy) OnData(data *types.DataPoint, portfolio types.Portfolio) {
	// 记录数据
	if s.logger != nil {
		s.logger.LogData(data)
	}

	// 获取指标值
	ma5, hasMA5 := data.Indicators["MA5"]
	macd, hasMACD := data.Indicators["MACD"]
	signal, hasSignal := data.Indicators["Signal"]
	histogram, hasHistogram := data.Indicators["MACDHistogram"]

	if !s.bought && hasMA5 && hasMACD && hasSignal && hasHistogram {
		// 买入条件：收盘价高于MA5的98%且MACD上穿信号线
		if data.Close > ma5*0.98 && macd > signal && histogram > 0 {
			portfolio.Buy(data.Timestamp, data.Close, 100)
			s.bought = true
			s.buyPrice = data.Close
			if s.logger != nil {
				s.logger.LogTrade(types.Trade{
					Timestamp: data.Timestamp,
					Price:     data.Close,
					Quantity:  100,
					Type:      types.ActionBuy,
				})
			}
		}
	} else if s.bought && hasMACD && hasSignal && hasHistogram {
		// 调整止损/止盈条件
		currentReturn := (data.Close - s.buyPrice) / s.buyPrice

		// 止损条件：下跌5%时卖出
		if currentReturn < -0.05 {
			portfolio.Sell(data.Timestamp, data.Close, 100)
			s.bought = false
			if s.logger != nil {
				s.logger.LogTrade(types.Trade{
					Timestamp: data.Timestamp,
					Price:     data.Close,
					Quantity:  100,
					Type:      types.ActionSell,
				})
			}
		}
		// 止盈条件：上涨10%时卖出
		if currentReturn > 0.10 {
			portfolio.Sell(data.Timestamp, data.Close, 100)
			s.bought = false
			if s.logger != nil {
				s.logger.LogTrade(types.Trade{
					Timestamp: data.Timestamp,
					Price:     data.Close,
					Quantity:  100,
					Type:      types.ActionSell,
				})
			}
		}
		// MACD下穿信号线时卖出
		if macd < signal && histogram < 0 {
			portfolio.Sell(data.Timestamp, data.Close, 100)
			s.bought = false
			if s.logger != nil {
				s.logger.LogTrade(types.Trade{
					Timestamp: data.Timestamp,
					Price:     data.Close,
					Quantity:  100,
					Type:      types.ActionSell,
				})
			}
		}
	}
}

func (s *SimpleStrategy) OnEnd(portfolio types.Portfolio) {
	// 记录结束状态
	if s.logger != nil {
		s.logger.LogEnd(portfolio)
	}

	// Close any open positions
	if s.bought {
		portfolio.Sell(time.Now(), 0, 1)
	}
}
