package portfolio

import (
	"stock/broker"
	"stock/common"
	"time"
)

type Portfolio struct {
	cash          float64
	initialCash   float64
	positions     map[string]float64
	trades        []common.Trade
	positionSizes map[string]float64
	broker        broker.Broker
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
	err := p.broker.Buy(timestamp, price, quantity)
	if err != nil {
		return err
	}

	cost := price * quantity
	fee := p.broker.CalculateTradeCost(common.ActionBuy, price, quantity)
	totalCost := cost + fee

	if p.cash >= totalCost {
		p.cash -= totalCost
		p.positions["asset"] += quantity
		p.positionSizes["asset"] += price * quantity
		p.trades = append(p.trades, common.Trade{
			Timestamp: timestamp,
			Price:     price,
			Quantity:  quantity,
			Type:      common.ActionBuy,
			Fee:       fee,
		})
	}
	return nil
}

func (p *Portfolio) Sell(timestamp time.Time, price float64, quantity float64) error {
	if p.positions["asset"] < quantity {
		return common.ErrInsufficientPosition
	}

	err := p.broker.Sell(timestamp, price, quantity)
	if err != nil {
		return err
	}

	proceeds := price * quantity
	fee := p.broker.CalculateTradeCost(common.ActionSell, price, quantity)
	totalProceeds := proceeds - fee

	p.cash += totalProceeds
	p.positions["asset"] -= quantity
	p.positionSizes["asset"] -= price * quantity
	p.trades = append(p.trades, common.Trade{
		Timestamp: timestamp,
		Price:     price,
		Quantity:  quantity,
		Type:      common.ActionSell,
		Fee:       fee,
	})
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
