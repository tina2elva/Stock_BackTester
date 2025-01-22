package broker

import (
	"fmt"
	"stock/common"
	"time"
)

// 订单状态类型
type OrderStatus int

const (
	OrderPending         OrderStatus = iota // 订单待处理
	OrderPartiallyFilled                    // 订单部分成交
	OrderFilled                             // 订单完全成交
	OrderCancelled                          // 订单已取消
)

// 订单结构
type Order struct {
	ID        string
	Timestamp time.Time
	Action    common.Action
	Price     float64
	Quantity  float64
	Filled    float64
	Status    OrderStatus
	Fee       float64
}

type Broker interface {
	// 创建新订单
	CreateOrder(action common.Action, price float64, quantity float64) (*Order, error)
	// 取消订单
	CancelOrder(orderID string) error
	// 获取订单状态
	GetOrderStatus(orderID string) (*Order, error)
	// 获取所有订单
	GetOrders() ([]*Order, error)
	// 计算交易成本
	CalculateTradeCost(action common.Action, price float64, quantity float64) float64
	// 获取日志记录器
	Logger() common.Logger
}

type SimulatedBroker struct {
	feeRate float64
	logger  common.Logger
	orders  map[string]*Order
}

func NewSimulatedBroker(feeRate float64, logger common.Logger) *SimulatedBroker {
	return &SimulatedBroker{
		feeRate: feeRate,
		logger:  logger,
		orders:  make(map[string]*Order),
	}
}

func (b *SimulatedBroker) Logger() common.Logger {
	return b.logger
}

func (b *SimulatedBroker) CreateOrder(action common.Action, price float64, quantity float64) (*Order, error) {
	order := &Order{
		ID:        generateOrderID(),
		Timestamp: time.Now(),
		Action:    action,
		Price:     price,
		Quantity:  quantity,
		Filled:    0,
		Status:    OrderPending,
		Fee:       b.CalculateTradeCost(action, price, quantity),
	}

	// 模拟订单执行
	if quantity > 0 {
		order.Filled = quantity
		order.Status = OrderFilled
	}

	b.orders[order.ID] = order
	return order, nil
}

func (b *SimulatedBroker) CancelOrder(orderID string) error {
	if order, exists := b.orders[orderID]; exists {
		order.Status = OrderCancelled
		return nil
	}
	return fmt.Errorf("order not found")
}

func (b *SimulatedBroker) GetOrderStatus(orderID string) (*Order, error) {
	if order, exists := b.orders[orderID]; exists {
		return order, nil
	}
	return nil, fmt.Errorf("order not found")
}

func (b *SimulatedBroker) GetOrders() ([]*Order, error) {
	orders := make([]*Order, 0, len(b.orders))
	for _, order := range b.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (b *SimulatedBroker) CalculateTradeCost(action common.Action, price float64, quantity float64) float64 {
	return price * quantity * b.feeRate
}

func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}
