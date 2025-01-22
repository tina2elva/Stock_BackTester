package broker

import (
	"fmt"
	"time"

	"stock/common/types"
)

type Broker interface {
	// 创建新订单
	CreateOrder(strategyID string, symbol string, quantity float64, orderType types.OrderType) (*types.Order, error)
	// 执行订单
	ExecuteOrder(order *types.Order) error
	// 取消订单
	CancelOrder(orderID string) error
	// 获取订单状态
	GetOrderStatus(orderID string) (*types.Order, error)
	// 获取所有订单
	GetOrders() ([]*types.Order, error)
	// 获取账户信息
	GetAccount() *types.Account
	// 计算交易成本
	CalculateTradeCost(action types.Action, price float64, quantity float64) float64
	// 获取日志记录器
	Logger() types.Logger
	// 获取单个仓位
	GetPosition(symbol string) (*types.Position, error)
	// 获取所有仓位
	GetPositions() (map[string]*types.Position, error)
	// 更新仓位
	UpdatePosition(symbol string, price float64, quantity float64, action types.Action) error
}

type SimulatedBroker struct {
	feeRate   float64
	logger    types.Logger
	account   *types.Account
	orders    map[string]*types.Order
	positions map[string]*types.Position
}

func NewSimulatedBroker(feeRate float64, logger types.Logger, initialCash float64) *SimulatedBroker {
	return &SimulatedBroker{
		feeRate: feeRate,
		logger:  logger,
		account: &types.Account{
			Cash:      initialCash,
			Equity:    initialCash,
			Balance:   initialCash,
			Positions: make(map[string]*types.Position),
		},
		orders:    make(map[string]*types.Order),
		positions: make(map[string]*types.Position),
	}
}

func (b *SimulatedBroker) GetPosition(symbol string) (*types.Position, error) {
	if pos, exists := b.positions[symbol]; exists {
		return pos, nil
	}
	return nil, types.ErrOrderNotFound
}

func (b *SimulatedBroker) GetPositions() (map[string]*types.Position, error) {
	return b.positions, nil
}

func (b *SimulatedBroker) UpdatePosition(symbol string, price float64, quantity float64, action types.Action) error {
	pos, exists := b.positions[symbol]
	if !exists {
		pos = types.NewPosition(symbol)
		b.positions[symbol] = pos
	}

	pos.Update(price, quantity, action)

	// 更新账户信息
	b.account.Positions[symbol] = pos
	b.account.Equity = b.account.Cash
	for _, p := range b.positions {
		b.account.Equity += p.MarketValue
	}

	return nil
}

func (b *SimulatedBroker) Logger() types.Logger {
	return b.logger
}

func (b *SimulatedBroker) CreateOrder(strategyID string, symbol string, quantity float64, orderType types.OrderType) (*types.Order, error) {
	order := &types.Order{
		ID:         generateOrderID(),
		StrategyID: strategyID,
		Symbol:     symbol,
		Quantity:   quantity,
		Type:       orderType,
		Status:     types.OrderStatusNew,
		CreatedAt:  time.Now(),
	}

	b.orders[order.ID] = order
	return order, nil
}

func (b *SimulatedBroker) ExecuteOrder(order *types.Order) error {
	if order.Status != types.OrderStatusNew {
		return types.ErrOrderCannotBeCanceled
	}

	// 计算交易成本
	cost := b.CalculateTradeCost(types.ActionBuy, order.Price, order.Quantity)
	if b.account.Cash < cost {
		return types.ErrInsufficientFunds
	}

	// 更新账户现金
	b.account.Cash -= cost
	b.account.Balance -= cost

	// 更新仓位
	var action types.Action
	switch order.Type {
	case types.OrderTypeBuy:
		action = types.ActionBuy
	case types.OrderTypeSell:
		action = types.ActionSell
	default:
		return types.ErrInvalidOrderState
	}

	err := b.UpdatePosition(order.Symbol, order.Price, order.Quantity, action)
	if err != nil {
		return err
	}

	// 更新订单状态
	order.Status = types.OrderStatusFilled
	order.UpdatedAt = time.Now()
	return nil
}

func (b *SimulatedBroker) CancelOrder(orderID string) error {
	order, exists := b.orders[orderID]
	if !exists {
		return types.ErrOrderNotFound
	}

	if order.Status != types.OrderStatusNew {
		return types.ErrOrderCannotBeCanceled
	}

	order.Status = types.OrderStatusCanceled
	order.UpdatedAt = time.Now()
	return nil
}

func (b *SimulatedBroker) GetOrderStatus(orderID string) (*types.Order, error) {
	if order, exists := b.orders[orderID]; exists {
		return order, nil
	}
	return nil, types.ErrOrderNotFound
}

func (b *SimulatedBroker) GetOrders() ([]*types.Order, error) {
	orders := make([]*types.Order, 0, len(b.orders))
	for _, order := range b.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (b *SimulatedBroker) GetAccount() *types.Account {
	return b.account
}

func (b *SimulatedBroker) CalculateTradeCost(action types.Action, price float64, quantity float64) float64 {
	return price * quantity * b.feeRate
}

func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}
