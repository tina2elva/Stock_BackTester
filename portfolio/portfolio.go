package portfolio

import (
	"stock/broker"
	"stock/common"
	"stock/strategy"
	"time"
)

type Portfolio struct {
	cash            float64
	initialCash     float64
	positions       map[string]float64
	trades          []common.Trade
	positionSizes   map[string]float64
	broker          broker.Broker
	currentStrategy strategy.Strategy
}

func NewPortfolio(initialCash float64, broker broker.Broker) *Portfolio {
	return &Portfolio{
		cash:          initialCash,
		initialCash:   initialCash,
		positions:     make(map[string]float64),
		trades:        make([]common.Trade, 0),
		positionSizes: make(map[string]float64),
		broker:        broker,
	}
}

func (p *Portfolio) SetCurrentStrategy(strategy strategy.Strategy) {
	p.currentStrategy = strategy
}

func (p *Portfolio) Balance() float64 {
	return p.cash
}

func (p *Portfolio) GetCash() float64 {
	return p.cash
}

func (p *Portfolio) GetInitialValue() float64 {
	return p.initialCash
}

func (p *Portfolio) AvailableCash() float64 {
	return p.cash
}

func (p *Portfolio) PositionSize(symbol string) float64 {
	return p.positionSizes[symbol]
}

func (p *Portfolio) Transactions() []common.Trade {
	return p.trades
}

func (p *Portfolio) Buy(timestamp time.Time, price float64, quantity float64) error {
	order, err := p.broker.CreateOrder(common.ActionBuy, price, quantity)
	if err != nil {
		return err
	}

	if order.Status == broker.OrderFilled {
		cost := price * quantity
		fee := p.broker.CalculateTradeCost(common.ActionBuy, price, quantity)
		totalCost := cost + fee

		if p.cash >= totalCost {
			p.cash -= totalCost
			p.positions["asset"] += quantity
			p.positionSizes["asset"] += quantity
			trade := common.Trade{
				Timestamp: timestamp,
				Price:     price,
				Quantity:  quantity,
				Type:      common.ActionBuy,
				Fee:       fee,
				Strategy:  p.currentStrategy.Name(),
				OrderID:   order.ID,
			}
			p.trades = append(p.trades, trade)

			// Log trade if logger is configured
			if p.broker.Logger() != nil {
				p.broker.Logger().LogTrade(trade)
			}
		}
	}
	return nil
}

func (p *Portfolio) Sell(timestamp time.Time, price float64, quantity float64) error {
	if p.positions["asset"] < quantity {
		return common.ErrInsufficientPosition
	}

	order, err := p.broker.CreateOrder(common.ActionSell, price, quantity)
	if err != nil {
		return err
	}

	if order.Status == broker.OrderFilled {
		proceeds := price * quantity
		fee := p.broker.CalculateTradeCost(common.ActionSell, price, quantity)
		totalProceeds := proceeds - fee

		p.cash += totalProceeds
		p.positions["asset"] -= quantity
		p.positionSizes["asset"] -= quantity
		trade := common.Trade{
			Timestamp: timestamp,
			Price:     price,
			Quantity:  quantity,
			Type:      common.ActionSell,
			Fee:       fee,
			Strategy:  p.currentStrategy.Name(),
			OrderID:   order.ID,
		}
		p.trades = append(p.trades, trade)

		// Log trade if logger is configured
		if p.broker.Logger() != nil {
			p.broker.Logger().LogTrade(trade)
		}
	}
	return nil
}

func (p *Portfolio) GetPositions() map[string]float64 {
	return p.positions
}

func (p *Portfolio) Positions() map[string]float64 {
	return p.positions
}

func (p *Portfolio) Trades() []common.Trade {
	return p.trades
}

func (p *Portfolio) GetTrades() []common.Trade {
	return p.trades
}

func (p *Portfolio) GetValue() float64 {
	// TODO: Implement actual price lookup
	currentPrice := 100.0 // Placeholder value
	positionValue := 0.0
	for _, qty := range p.positions {
		positionValue += qty * currentPrice
	}
	return p.cash + positionValue
}
