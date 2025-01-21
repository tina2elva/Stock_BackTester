package common

import (
	"fmt"
	"time"
)

// Ensure time package is used for timestamp formatting
var _ = time.Time{}

type ConsoleLogger struct{}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (l *ConsoleLogger) LogData(data *DataPoint) {
	// 禁用数据日志
}

func (l *ConsoleLogger) LogTrade(trade Trade) {
	// 交易日志格式
	action := "买入"
	if trade.Type == ActionSell {
		action = "卖出"
	}

	totalAmount := trade.Quantity * trade.Price
	netAmount := totalAmount - trade.Fee

	fmt.Printf("[交易] %s %s %.2f股 @ %.2f元\n",
		trade.Timestamp.Format("2006-01-02 15:04:05"),
		action,
		trade.Quantity,
		trade.Price)

	fmt.Printf("    交易总额: %.2f元, 手续费: %.2f元, 净交易额: %.2f元\n",
		totalAmount,
		trade.Fee,
		netAmount)
}

func (l *ConsoleLogger) LogEnd(portfolio Portfolio) {
	// 结束日志格式
	fmt.Printf("[END] Portfolio Value: %.2f, Cash: %.2f, Positions: %v\n",
		portfolio.GetValue(),
		portfolio.GetCash(),
		portfolio.GetPositions())
}
