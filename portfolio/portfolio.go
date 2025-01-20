package portfolio

import (
	"stock/common"
	"time"
)

type Portfolio struct {
	cash          float64
	initialCash   float64
	positions     map[string]float64
	trades        []common.Trade
	positionSizes map[string]float64
	feeCalculator common.FeeCalculator
}

func NewPortfolio(initialCash float64) *Portfolio {
	return &Portfolio{
		cash:          initialCash,
		initialCash:   initialCash,
		positions:     make(map[string]float64),
		trades:        make([]common.Trade, 0),
		positionSizes: make(map[string]float64),
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

func (p *Portfolio) SetFeeCalculator(calculator common.FeeCalculator) {
	p.feeCalculator = calculator
}

func (p *Portfolio) CalculateTradeCost(action common.Action, price float64, quantity float64) float64 {
	if p.feeCalculator == nil {
		return 0
	}
	return p.feeCalculator.CalculateFee(action, price, quantity)
}

func (p *Portfolio) Buy(timestamp time.Time, price float64, quantity float64) {
	// 计算实际价格（考虑滑点）
	actualPrice := price
	if p.feeCalculator != nil {
		actualPrice = p.feeCalculator.GetActualPrice(common.ActionBuy, price)
	}

	// 计算总成本（包括交易费用）
	totalCost := actualPrice * quantity
	if p.feeCalculator != nil {
		totalCost += p.feeCalculator.CalculateFee(common.ActionBuy, actualPrice, quantity)
	}

	if p.cash >= totalCost {
		// 扣除总成本
		p.cash -= totalCost
		// 更新持仓
		p.positions["asset"] += quantity
		p.positionSizes["asset"] += actualPrice * quantity
		// 记录交易
		p.trades = append(p.trades, common.Trade{
			Timestamp: timestamp,
			Price:     actualPrice,
			Quantity:  quantity,
			Type:      common.ActionBuy,
			Fee:       p.feeCalculator.CalculateFee(common.ActionBuy, actualPrice, quantity),
		})
	}
}

func (p *Portfolio) Sell(timestamp time.Time, price float64, quantity float64) {
	if p.positions["asset"] >= quantity {
		// 计算实际价格（考虑滑点）
		actualPrice := price
		if p.feeCalculator != nil {
			actualPrice = p.feeCalculator.GetActualPrice(common.ActionSell, price)
		}

		// 计算总收入（扣除交易费用）
		totalCost := actualPrice * quantity
		if p.feeCalculator != nil {
			totalCost -= p.feeCalculator.CalculateFee(common.ActionSell, actualPrice, quantity)
		}

		// 更新现金
		p.cash += totalCost
		// 更新持仓
		p.positions["asset"] -= quantity
		p.positionSizes["asset"] -= actualPrice * quantity
		// 记录交易
		p.trades = append(p.trades, common.Trade{
			Timestamp: timestamp,
			Price:     actualPrice,
			Quantity:  quantity,
			Type:      common.ActionSell,
			Fee:       p.feeCalculator.CalculateFee(common.ActionSell, actualPrice, quantity),
		})
	}
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
