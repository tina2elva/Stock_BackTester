package orders

import (
	"errors"
	"time"

	"stock/common/types"
)

var (
	ErrInvalidOrderState = errors.New("invalid order state transition")
)

// OrderManager 订单管理器
type OrderManager struct {
	orders map[string]*types.Order
	broker types.Broker
}

// NewOrderManager 创建新的订单管理器
func NewOrderManager(broker types.Broker) *OrderManager {
	return &OrderManager{
		orders: make(map[string]*types.Order),
		broker: broker,
	}
}

// CreateOrder 创建新订单
func (om *OrderManager) CreateOrder(strategyID, symbol string, quantity float64, orderType types.OrderType) (*types.Order, error) {
	if quantity <= 0 {
		return nil, types.ErrInvalidQuantity
	}

	order := &types.Order{
		ID:         generateOrderID(),
		StrategyID: strategyID,
		Symbol:     symbol,
		Quantity:   quantity,
		Type:       orderType,
		Status:     types.OrderStatusNew,
		CreatedAt:  time.Now(),
	}

	om.orders[order.ID] = order
	return order, nil
}

// ExecuteOrder 执行订单
func (om *OrderManager) ExecuteOrder(orderID string) error {
	order, err := om.validateOrder(orderID)
	if err != nil {
		return err
	}

	// 验证订单状态转换
	if !CanExecute(order) {
		return ErrInvalidOrderState
	}

	// 调用broker执行订单
	if err := om.broker.ExecuteOrder(order); err != nil {
		if err := SetOrderStatus(order, types.OrderStatusRejected); err != nil {
			return err
		}
		return err
	}

	return SetOrderStatus(order, types.OrderStatusFilled)
}

// CancelOrder 取消订单
func (om *OrderManager) CancelOrder(orderID string) error {
	order, err := om.validateOrder(orderID)
	if err != nil {
		return err
	}

	// 验证订单状态转换
	if !CanCancel(order) {
		return types.ErrOrderCannotBeCanceled
	}

	return SetOrderStatus(order, types.OrderStatusCanceled)
}

// GetOrder 获取订单详情
func (om *OrderManager) GetOrder(orderID string) (*types.Order, error) {
	return om.validateOrder(orderID)
}

// validateOrder 验证订单有效性
func (om *OrderManager) validateOrder(orderID string) (*types.Order, error) {
	order, exists := om.orders[orderID]
	if !exists {
		return nil, types.ErrOrderNotFound
	}
	return order, nil
}

// generateOrderID 生成唯一订单ID
func generateOrderID() string {
	return "order_" + time.Now().Format("20060102150405")
}

// CanExecute 判断订单是否可以执行
func CanExecute(o *types.Order) bool {
	return o.Status == types.OrderStatusNew || o.Status == types.OrderStatusFilled
}

// CanCancel 判断订单是否可以取消
func CanCancel(o *types.Order) bool {
	return o.Status == types.OrderStatusNew
}

// SetOrderStatus 设置订单状态
func SetOrderStatus(o *types.Order, status types.OrderStatus) error {
	// 验证状态转换
	switch status {
	case types.OrderStatusFilled:
		if !CanExecute(o) {
			return ErrInvalidOrderState
		}
	case types.OrderStatusCanceled:
		if !CanCancel(o) {
			return types.ErrOrderCannotBeCanceled
		}
	case types.OrderStatusRejected:
		// 任何状态都可以被拒绝
	default:
		return ErrInvalidOrderState
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}
