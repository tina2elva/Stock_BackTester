package common

import (
	"fmt"
	"stock/common/types"
)

type ConsoleLogger struct{}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (l *ConsoleLogger) LogData(data *types.DataPoint) {
	// 禁用数据日志
}

func (l *ConsoleLogger) LogTrade(trade types.Trade) {
	// 交易日志格式
	action := "买入"
	if trade.Type == types.ActionSell {
		action = "卖出"
	}

	totalAmount := trade.Quantity * trade.Price
	netAmount := totalAmount - trade.Fee

	fmt.Printf("[交易] %s %s %.2f股 @ %.2f元,交易总额: %.2f元, 手续费: %.2f元, 净交易额: %.2f元\n",
		trade.Timestamp.Format("2006-01-02 15:04:05"),
		action,
		trade.Quantity,
		trade.Price,
		totalAmount,
		trade.Fee,
		netAmount)
}

func (l *ConsoleLogger) LogEnd(portfolio types.Portfolio) {
	// 结束日志格式
	fmt.Printf("[END] Portfolio Value: %.2f, Cash: %.2f, Positions: %v\n",
		portfolio.GetValue(),
		portfolio.GetCash(),
		portfolio.GetPositions())
}
