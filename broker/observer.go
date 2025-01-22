package broker

import (
	"stock/common/types"
	"sync"
)

// DefaultObserver 默认观测器实现
type DefaultObserver struct {
	mu     sync.Mutex
	trades []*types.Trade
	orders []*types.Order
}

// NewDefaultObserver 创建新的默认观测器
func NewDefaultObserver() *DefaultObserver {
	return &DefaultObserver{
		trades: make([]*types.Trade, 0),
		orders: make([]*types.Order, 0),
	}
}

// OnOrder 处理新订单
func (o *DefaultObserver) OnOrder(order *types.Order) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.orders = append(o.orders, order)
}

// OnTrade 处理交易完成
func (o *DefaultObserver) OnTrade(trade *types.Trade) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.trades = append(o.trades, trade)
}

// GetTrades 获取所有交易记录
func (o *DefaultObserver) GetTrades() []*types.Trade {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.trades
}

// GetOrders 获取所有订单记录
func (o *DefaultObserver) GetOrders() []*types.Order {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.orders
}

// Clear 清除所有记录
func (o *DefaultObserver) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.trades = make([]*types.Trade, 0)
	o.orders = make([]*types.Order, 0)
}
