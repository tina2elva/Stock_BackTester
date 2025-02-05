package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"stock/analyzer"
	"stock/backtest"
	"stock/broker"
	"stock/common"
	"stock/common/types"
	"stock/datasource"
	"stock/strategy"
	"stock/visualization"
)

func main() {
	// 测试通达信数据源
	tdxDs := datasource.NewTDXDataSource("data/sh600036.day")
	startDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)
	tdxData, err := tdxDs.GetData("600036.SH", datasource.PeriodTypeDay, startDate, endDate)
	if err != nil {
		log.Fatalf("获取通达信数据失败: %v", err)
	}
	fmt.Printf("成功读取 %d 条通达信数据\n", len(tdxData))

	// 使用CSV数据源
	ds := datasource.NewCSVDataSource("data/cmb.csv")

	// 获取招商银行2020-2022年历史数据
	data, err := ds.GetData("600036.SH", datasource.PeriodTypeDay, startDate, endDate)
	if err != nil {
		log.Fatalf("获取数据失败: %v", err)
	}

	// 初始化多个策略
	strategies := []strategy.Strategy{
		strategy.NewMACDStrategy(12, 26, 9, []int{5, 10, 20}),
	}

	// 初始化费用配置
	feeConfig := backtest.DefaultFeeConfig

	// 初始资金
	initialCash := 100000.0

	// 初始化broker
	logger := common.NewConsoleLogger()
	broker := broker.NewSimulatedBroker(
		broker.NewFixedFeeCalculator(feeConfig.Commission),
		types.Logger(logger),
		initialCash,
	)

	// 初始化回测引擎
	bt := backtest.NewBacktest(startDate, endDate, initialCash, tdxDs, broker, logger, []string{"600036.SH"})
	for _, strategy := range strategies {
		bt.AddStrategy(strategy)
	}

	// 运行回测
	results, _ := bt.Run()

	// 获取回测结果
	if len(results.Results) == 0 {
		log.Fatal("没有回测结果")
	}

	// 遍历所有策略结果
	for i, result := range results.Results {
		strategyName := strings.TrimPrefix(fmt.Sprintf("%T", strategies[i]), "*strategy.")
		strategyName = strings.TrimSuffix(strategyName, "Strategy")
		strategyName = strings.ToUpper(strategyName)
		fmt.Printf("\n策略 %d (%s) 回测结果:\n", i+1, strategyName)

		// 初始化analyzer
		analyzer := analyzer.NewAnalyzer(result.Trades, initialCash)
		// 计算关键指标
		finalValue := result.FinalValue
		duration := endDate.Sub(startDate)
		totalReturn := analyzer.TotalReturn(finalValue)
		annualizedReturn := analyzer.AnnualizedReturn(finalValue, duration)
		maxDrawdown := result.MaxDrawdown
		winRate := analyzer.WinRate()
		avgProfit, avgLoss := analyzer.AverageProfitLoss()
		profitLossRatio := analyzer.ProfitLossRatio()

		// 输出回测结果
		fmt.Printf("初始资金: %.2f\n", initialCash)
		fmt.Printf("最终资产: %.2f\n", finalValue)
		fmt.Printf("总收益率: %.2f%%\n", totalReturn*100)
		fmt.Printf("年化收益率: %.2f%%\n", annualizedReturn*100)
		fmt.Printf("最大回撤: %.2f%%\n", maxDrawdown*100)
		fmt.Printf("交易次数: %d\n", len(result.Trades))
		fmt.Printf("胜率: %.2f%%\n", winRate*100)
		fmt.Printf("平均盈利: %.2f\n", avgProfit)
		fmt.Printf("平均亏损: %.2f\n", avgLoss)
		fmt.Printf("盈亏比: %.2f\n", profitLossRatio)

		// 计算新增指标
		returns := result.Returns
		annualVolatility := analyzer.AnnualVolatility(returns)
		sortinoRatio := analyzer.SortinoRatio(returns, 0)
		calmarRatio := analyzer.CalmarRatio(finalValue, maxDrawdown, duration)
		drawdownDuration := analyzer.DrawdownDuration(result.Values)
		var95 := analyzer.ValueAtRisk(returns, 0.95)

		// 显示新增指标
		fmt.Printf("年波动率: %.2f%%\n", annualVolatility*100)
		fmt.Printf("索提诺比率: %.2f\n", sortinoRatio)
		fmt.Printf("卡尔玛比率: %.2f\n", calmarRatio)
		fmt.Printf("最大回撤持续时间: %s\n", drawdownDuration)
		fmt.Printf("95%%置信度VaR: %.2f%%\n", var95*100)

		// 将DataPoint转换为Candle
		candles := make([]types.Candle, len(data))
		for i, dp := range data {
			candles[i] = types.Candle{
				Timestamp:  dp.Timestamp,
				Open:       dp.Open,
				High:       dp.High,
				Low:        dp.Low,
				Close:      dp.Close,
				Volume:     dp.Volume,
				Indicators: make(map[string]interface{}),
			}
		}

		// 计算并填充指标数据
		for _, strategy := range strategies {
			indicatorValues := strategy.Calculate(candles)
			for indicatorName, values := range indicatorValues {
				for i := range candles {
					candles[i].Indicators[indicatorName] = values[i]
				}
			}
		}

		// 按策略名称分组交易数据
		tradesMap := map[string][]types.Trade{
			strategyName: result.Trades,
		}

		// 可视化结果
		chart := visualization.NewChart(fmt.Sprintf("招商银行 %s 策略回测", strategyName))
		chartFile := fmt.Sprintf("cmb_%s_candlestick.html", strategyName)
		err = chart.PlotCandlestick(candles, tradesMap, chartFile)
		if err != nil {
			log.Fatalf("生成图表失败: %v", err)
		}

		// 保存交易记录
		tradeFile := fmt.Sprintf("cmb_%s_trades.csv", strategyName)
		file, err := os.Create(tradeFile)
		if err != nil {
			log.Fatalf("创建交易记录文件失败: %v", err)
		}
		defer file.Close()

		file.WriteString("时间,类型,价格,数量,费用\n")
		for _, trade := range result.Trades {
			file.WriteString(fmt.Sprintf("%s,%s,%.2f,%.2f,%.2f\n",
				trade.Timestamp.Format("2006-01-02"),
				trade.Type,
				trade.Price,
				trade.Quantity,
				trade.Fee))
		}

		fmt.Printf("\n策略 %d (%s) 回测结果已保存到 %s 和 %s\n",
			i+1, strategyName, chartFile, tradeFile)
	}

	fmt.Println("\n所有策略回测完成")
}
