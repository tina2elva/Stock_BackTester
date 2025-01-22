package broker

import (
	"fmt"
	"time"

	"stock/common"
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
	CalculateTradeCost(action common.Action, price float64, quantity float64) float64
	// 获取日志记录器
	Logger() common.Logger
}

type SimulatedBroker struct {
	feeRate float64
	logger  common.Logger
	account *types.Account
	orders  map[string]*types.Order
}

func NewSimulatedBroker(feeRate float64, logger common.Logger, initialCash float64) *SimulatedBroker {
	return &SimulatedBroker{
		feeRate: feeRate,
		logger:  logger,
		account: &types.Account{
			Cash:    initialCash,
			Equity:  initialCash,
			Balance: initialCash,
		},
		orders: make(map[string]*types.Order),
	}
}

func (b *SimulatedBroker) Logger() common.Logger {
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

	// 模拟订单执行
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

func (b *SimulatedBroker) CalculateTradeCost(action common.Action, price float64, quantity float64) float64 {
	return price * quantity * b.feeRate
}

func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}
