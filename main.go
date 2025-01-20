package main

import (
	"fmt"
	"log"
	"os"
	"stock/backtest"
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

	// 初始化MACD策略
	// 初始化费用配置
	feeConfig := backtest.DefaultFeeConfig()

	// 初始化RSI策略
	rsiStrategy := strategy.NewRSIStrategy(14, 30, 70)
	bt := backtest.NewBacktest(data, rsiStrategy, 100000, feeConfig) // 初始资金10万

	// 运行回测
	bt.Run()

	// 获取回测结果
	results := bt.Results()

	// 输出回测结果
	fmt.Println("回测结果:")
	fmt.Printf("初始资金: %.2f\n", 100000.0)
	fmt.Printf("最终资产: %.2f\n", results.FinalValue)
	fmt.Printf("收益率: %.2f%%\n", (results.FinalValue-100000)/100000*100)
	fmt.Printf("最大回撤: %.2f%%\n", results.Metrics.MaxDrawdown*100)
	fmt.Printf("交易次数: %d\n", results.Metrics.NumTrades)
	fmt.Printf("胜率: %.2f%%\n", results.Metrics.WinRate*100)

	// 可视化结果
	chart := visualization.NewChart("招商银行 RSI策略回测")
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

	file.WriteString("时间,类型,价格,数量\n")
	for _, trade := range results.Trades {
		file.WriteString(fmt.Sprintf("%s,%s,%.2f,%.2f\n",
			trade.Timestamp.Format("2006-01-02"),
			trade.Type,
			trade.Price,
			trade.Quantity))
	}

	fmt.Println("回测完成，结果已保存到 cmb_candlestick.png 和 cmb_trades.csv")
}
