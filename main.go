package main

import (
	"fmt"
	"log"
	"os"
	"stock/analyzer"
	"stock/backtest"
	"stock/broker"
	"stock/datasource"
	"stock/strategy"
	"stock/visualization"
	"time"
)

func main() {
	// 初始化数据源
	ds := datasource.NewCSVDataSource("data/cmb.csv")

	// 获取招商银行2020-2022年历史数据
	startDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)
	data, err := ds.GetData("600036.SH", startDate, endDate)
	if err != nil {
		log.Fatalf("获取数据失败: %v", err)
	}

	// 初始化多个策略
	strategies := []strategy.Strategy{
		strategy.NewMACDStrategy(12, 26, 9),
		strategy.NewRSIStrategy(14, 30, 70),
	}

	// 初始化费用配置
	feeConfig := backtest.DefaultFeeConfig()

	// 初始资金
	initialCash := 100000.0

	// 初始化broker
	broker := broker.NewSimulatedBroker(feeConfig.Commission)

	// 初始化回测引擎
	bt := backtest.NewBacktest(data, initialCash, feeConfig, broker)
	for _, strategy := range strategies {
		bt.AddStrategy(strategy)
	}

	// 运行回测
	bt.Run()

	// 获取回测结果
	results := bt.Results()

	// 初始化analyzer
	analyzer := analyzer.NewAnalyzer(results.Trades, initialCash)

	// 计算关键指标
	finalValue := results.FinalValue
	duration := endDate.Sub(startDate)
	totalReturn := analyzer.TotalReturn(finalValue)
	annualizedReturn := analyzer.AnnualizedReturn(finalValue, duration)
	maxDrawdown := analyzer.MaxDrawdown(results.EquityCurve)
	winRate := analyzer.WinRate()
	avgProfit, avgLoss := analyzer.AverageProfitLoss()
	profitLossRatio := analyzer.ProfitLossRatio()

	// 输出回测结果
	fmt.Println("\n回测结果:")
	fmt.Printf("初始资金: %.2f\n", initialCash)
	fmt.Printf("最终资产: %.2f\n", finalValue)
	fmt.Printf("总收益率: %.2f%%\n", totalReturn*100)
	fmt.Printf("年化收益率: %.2f%%\n", annualizedReturn*100)
	fmt.Printf("最大回撤: %.2f%%\n", maxDrawdown*100)
	fmt.Printf("交易次数: %d\n", len(results.Trades))
	fmt.Printf("胜率: %.2f%%\n", winRate*100)
	fmt.Printf("平均盈利: %.2f\n", avgProfit)
	fmt.Printf("平均亏损: %.2f\n", avgLoss)
	fmt.Printf("盈亏比: %.2f\n", profitLossRatio)

	// 可视化结果
	chart := visualization.NewChart("招商银行 MACD策略回测")
	err = chart.PlotCandlestick(data, "cmb_candlestick.png")
	if err != nil {
		log.Fatalf("生成图表失败: %v", err)
	}

	// 保存交易记录
	file, err := os.Create("cmb_trades.csv")
	if err != nil {
		log.Fatalf("创建交易记录文件失败: %v", err)
	}
	defer file.Close()

	file.WriteString("时间,类型,价格,数量,费用\n")
	for _, trade := range results.Trades {
		file.WriteString(fmt.Sprintf("%s,%s,%.2f,%.2f,%.2f\n",
			trade.Timestamp.Format("2006-01-02"),
			trade.Type,
			trade.Price,
			trade.Quantity,
			trade.Fee))
	}

	fmt.Println("\n回测完成，结果已保存到 cmb_candlestick.png 和 cmb_trades.csv")
}
