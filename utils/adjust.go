package utils

import (
	"fmt"
)

// EventType 复权事件类型
type EventType string

const (
	EventDividend EventType = "dividend" // 现金分红
	EventBonus    EventType = "bonus"    // 送股
	EventSplit    EventType = "split"    // 拆股
	EventRights   EventType = "rights"   // 配股
)

// StockPrice 股票价格数据
type StockPrice struct {
	Date   string  // 日期
	Open   float64 // 开盘价
	Close  float64 // 收盘价
	High   float64 // 最高价
	Low    float64 // 最低价
	Volume int64   // 成交量
}

// Event 复权事件
type Event struct {
	Type  EventType // 事件类型
	Value float64   // 事件值（分红金额、送股比例、拆股比例等）
	Price float64   // 配股价格（仅用于配股事件）
}

// AdjustPrice 复权价格计算
func AdjustPrice(prices []StockPrice, eventMap map[string][]Event, isForward bool) []StockPrice {
	// 复权因子
	adjustFactor := 1.0
	adjustedPrices := make([]StockPrice, len(prices))

	for i := len(prices) - 1; i >= 0; i-- {
		price := prices[i]

		// 如果当前日期有复权事件
		if events, ok := eventMap[price.Date]; ok {
			for _, event := range events {
				switch event.Type {
				case EventDividend:
					// 现金分红：调整价格
					if isForward {
						adjustFactor *= (1 - event.Value/price.Close)
					} else {
						adjustFactor /= (1 - event.Value/price.Close)
					}
				case EventBonus:
					// 送股：调整价格
					if isForward {
						adjustFactor *= (1 + event.Value)
					} else {
						adjustFactor /= (1 + event.Value)
					}
				case EventSplit:
					// 拆股：调整价格
					if isForward {
						adjustFactor *= event.Value
					} else {
						adjustFactor /= event.Value
					}
				case EventRights:
					// 配股：调整价格
					rightsRatio := event.Value
					rightsPrice := event.Price
					if isForward {
						adjustFactor *= (price.Close + rightsRatio*rightsPrice) / (price.Close * (1 + rightsRatio))
					} else {
						adjustFactor /= (price.Close + rightsRatio*rightsPrice) / (price.Close * (1 + rightsRatio))
					}
				}
			}
		}

		// 复权后的价格
		adjustedPrices[i] = StockPrice{
			Date:   price.Date,
			Open:   price.Open * adjustFactor,
			Close:  price.Close * adjustFactor,
			High:   price.High * adjustFactor,
			Low:    price.Low * adjustFactor,
			Volume: price.Volume,
		}
	}

	return adjustedPrices
}

func main() {
	// 示例数据
	prices := []StockPrice{
		{Date: "2023-01-01", Open: 100, Close: 105, High: 110, Low: 95, Volume: 1000},
		{Date: "2023-02-01", Open: 105, Close: 110, High: 115, Low: 100, Volume: 1200},
		{Date: "2023-03-01", Open: 110, Close: 115, High: 120, Low: 105, Volume: 1300},
		{Date: "2023-04-01", Open: 115, Close: 120, High: 125, Low: 110, Volume: 1400},
	}

	// 事件数据（同一日期可能包含多个事件）
	eventMap := map[string][]Event{
		"2023-03-01": {
			{Type: EventDividend, Value: 5}, // 现金分红 5 元
			{Type: EventBonus, Value: 0.1},  // 送股 10%
		},
		"2023-02-01": {
			{Type: EventSplit, Value: 2}, // 拆股 1:2
		},
		"2023-04-01": {
			{Type: EventRights, Value: 0.5, Price: 50}, // 配股 10 配 5，配股价 50 元
		},
	}

	// 前复权
	forwardAdjusted := AdjustPrice(prices, eventMap, true)
	fmt.Println("前复权价格:")
	for _, price := range forwardAdjusted {
		fmt.Printf("Date: %s, Open: %.2f, Close: %.2f, High: %.2f, Low: %.2f, Volume: %d\n",
			price.Date, price.Open, price.Close, price.High, price.Low, price.Volume)
	}

	// 后复权
	backwardAdjusted := AdjustPrice(prices, eventMap, false)
	fmt.Println("后复权价格:")
	for _, price := range backwardAdjusted {
		fmt.Printf("Date: %s, Open: %.2f, Close: %.2f, High: %.2f, Low: %.2f, Volume: %d\n",
			price.Date, price.Open, price.Close, price.High, price.Low, price.Volume)
	}
}
