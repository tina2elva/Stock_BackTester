package portfolio

import (
	"stock/broker"
	"stock/common/types"
	"stock/orders"
	"time"
)

type Portfolio struct {
	cash           float64
	initialCash    float64
	positions      map[string]float64 // 各股票持仓数量
	positionPrices map[string]float64 // 各股票持仓成本价
	trades         []types.Trade
	positionSizes  map[string]float64
	broker         broker.Broker
	orderManager   *orders.OrderManager
}

func NewPortfolio(initialCash float64, broker broker.Broker, orderManager *orders.OrderManager) *Portfolio {
	return &Portfolio{
		cash:           initialCash,
		initialCash:    initialCash,
		positions:      make(map[string]float64),
		positionPrices: make(map[string]float64),
		trades:         make([]types.Trade, 0),
		positionSizes:  make(map[string]float64),
		broker:         broker,
		orderManager:   orderManager,
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

func (p *Portfolio) Transactions() []types.Trade {
	return p.trades
}

func (p *Portfolio) Buy(symbol string, timestamp time.Time, price float64, quantity float64) error {
	// 通过OrderManager创建订单
	order, err := p.orderManager.CreateOrder("manual", symbol, quantity, types.OrderTypeBuy)
	if err != nil {
		return err
	}

	// 执行订单
	err = p.orderManager.ExecuteOrder(order.ID)
	if err != nil {
		return err
	}

	// 获取更新后的订单状态
	order, err = p.orderManager.GetOrder(order.ID)
	if err != nil {
		return err
	}

	if order.Status == types.OrderStatusFilled {
		cost := price * quantity
		fee := p.broker.CalculateTradeCost(types.ActionBuy, price, quantity)
		totalCost := cost + fee

		if p.cash >= totalCost {
			p.cash -= totalCost
			p.positions[symbol] += quantity
			p.positionPrices[symbol] = price
			p.positionSizes[symbol] += quantity
			trade := types.Trade{
				Timestamp: timestamp,
				Symbol:    symbol,
				Price:     price,
				Quantity:  quantity,
				Type:      types.ActionBuy,
				Fee:       fee,
				Strategy:  "manual",
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

func (p *Portfolio) Sell(symbol string, timestamp time.Time, price float64, quantity float64) error {
	if p.positions[symbol] < quantity {
		return types.ErrInsufficientPosition
	}

	// 通过OrderManager创建订单
	order, err := p.orderManager.CreateOrder("manual", symbol, quantity, types.OrderTypeSell)
	if err != nil {
		return err
	}

	// 执行订单
	err = p.orderManager.ExecuteOrder(order.ID)
	if err != nil {
		return err
	}

	// 获取更新后的订单状态
	order, err = p.orderManager.GetOrder(order.ID)
	if err != nil {
		return err
	}

	if order.Status == types.OrderStatusFilled {
		proceeds := price * quantity
		fee := p.broker.CalculateTradeCost(types.ActionSell, price, quantity)
		totalProceeds := proceeds - fee

		p.cash += totalProceeds
		p.positions[symbol] -= quantity
		p.positionSizes[symbol] -= quantity
		trade := types.Trade{
			Timestamp: timestamp,
			Symbol:    symbol,
			Price:     price,
			Quantity:  quantity,
			Type:      types.ActionSell,
			Fee:       fee,
			Strategy:  "manual",
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

func (p *Portfolio) Trades() []types.Trade {
	return p.trades
}

func (p *Portfolio) GetTrades() []types.Trade {
	return p.trades
}

func (p *Portfolio) GetValue() float64 {
	positionValue := 0.0
	for symbol, qty := range p.positions {
		// 使用持仓成本价计算价值
		positionValue += qty * p.positionPrices[symbol]
	}
	return p.cash + positionValue
}

// GetSymbolValue 获取指定股票持仓价值
func (p *Portfolio) GetSymbolValue(symbol string) float64 {
	if qty, ok := p.positions[symbol]; ok {
		return qty * p.positionPrices[symbol]
	}
	return 0
}

// GetSymbolPosition 获取指定股票持仓信息
func (p *Portfolio) GetSymbolPosition(symbol string) (float64, float64) {
	qty := p.positions[symbol]
	price := p.positionPrices[symbol]
	return qty, price
}
