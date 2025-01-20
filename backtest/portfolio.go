package backtest

import (
	"time"

	"stock/common"
)

type portfolioImpl struct {
	cash         float64
	initialCash  float64
	positions    map[string]float64
	tradeHistory []common.Trade
}

func NewPortfolio(initialCash float64) *portfolioImpl {
	return &portfolioImpl{
		cash:         initialCash,
		initialCash:  initialCash,
		positions:    make(map[string]float64),
		tradeHistory: make([]common.Trade, 0),
	}
}

func (p *portfolioImpl) GetTradeHistory() []common.Trade {
	return p.tradeHistory
}

func (p *portfolioImpl) GetInitialValue() float64 {
	return p.initialCash
}

func (p *portfolioImpl) Buy(timestamp time.Time, price float64, quantity float64) {
	totalCost := price * quantity
	if totalCost > p.cash {
		return
	}
	p.cash -= totalCost
	p.positions["asset"] += quantity
}

func (p *portfolioImpl) Sell(timestamp time.Time, price float64, quantity float64) {
	if quantity > p.positions["asset"] {
		return
	}
	p.cash += price * quantity
	p.positions["asset"] -= quantity
}

func (p *portfolioImpl) GetPositions() map[string]float64 {
	return p.positions
}

func (p *portfolioImpl) GetCash() float64 {
	return p.cash
}

func (p *portfolioImpl) GetValue() float64 {
	value := p.cash
	for _, quantity := range p.positions {
		value += quantity * 100 // Assuming price is 100 for simplicity
	}
	return value
}
