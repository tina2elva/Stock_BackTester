package backtest

import (
	"math"
	"stock/broker"
	"stock/common"
	"stock/portfolio"
	"stock/strategy"
	"sync"
)

type Backtest struct {
	data       []*common.DataPoint
	strategies []strategy.Strategy
	portfolios []*portfolio.Portfolio
	feeConfig  common.FeeConfig
}

func (b *Backtest) AddStrategy(strategy strategy.Strategy) {
	b.strategies = append(b.strategies, strategy)
	// 为每个策略创建一个对应的portfolio
	broker := broker.NewSimulatedBroker(0.0003) // 使用默认佣金率
	initialCash := 100000.0                     // 默认初始资金
	b.portfolios = append(b.portfolios, portfolio.NewPortfolio(initialCash, broker))
}

type PerformanceMetrics struct {
	TotalReturn  float64
	MaxDrawdown  float64
	WinRate      float64
	SharpeRatio  float64
	SortinoRatio float64
	Volatility   float64
	ProfitFactor float64
	NumTrades    int
}

type BacktestResults struct {
	FinalValue  float64
	Positions   map[string]float64
	Cash        float64
	Trades      []common.Trade
	Metrics     PerformanceMetrics
	EquityCurve []float64
}

func NewBacktest(data []*common.DataPoint, initialCash float64, feeConfig common.FeeConfig, broker broker.Broker) *Backtest {
	return &Backtest{
		data:       data,
		feeConfig:  feeConfig,
		strategies: []strategy.Strategy{},
		portfolios: []*portfolio.Portfolio{},
	}
}

// DefaultFeeConfig 返回默认的费用配置
func DefaultFeeConfig() common.FeeConfig {
	return common.FeeConfig{
		StampDuty:  0.001,  // 印花税 0.1%
		Commission: 0.0003, // 佣金 0.03%
		Fee:        5.0,    // 固定手续费 5元
		Slippage:   0.0005, // 滑点 0.05%
		MinLotSize: 1,      // 最小交易单位 1手
	}
}

func (b *Backtest) Run() {
	var wg sync.WaitGroup
	wg.Add(len(b.strategies))

	for i := range b.strategies {
		go func(index int) {
			defer wg.Done()
			for _, dataPoint := range b.data {
				// 在交易前设置当前策略名称
				b.portfolios[index].SetCurrentStrategy(b.strategies[index])
				b.strategies[index].OnData(dataPoint, b.portfolios[index])
			}
			b.portfolios[index].SetCurrentStrategy(b.strategies[index])
			b.strategies[index].OnEnd(b.portfolios[index])
		}(i)
	}
	wg.Wait()
}

func (b *Backtest) Results() []*BacktestResults {
	var results []*BacktestResults

	for i := range b.strategies {
		result := &BacktestResults{
			FinalValue:  b.portfolios[i].Balance(),
			Positions:   b.portfolios[i].GetPositions(),
			Cash:        b.portfolios[i].AvailableCash(),
			Trades:      b.portfolios[i].Transactions(),
			EquityCurve: b.calculateEquityCurve(b.portfolios[i]),
		}
		result.Metrics = b.calculateMetrics(result, b.portfolios[i])
		results = append(results, result)
	}
	return results
}

func (b *Backtest) calculateEquityCurve(p *portfolio.Portfolio) []float64 {
	var equityCurve []float64
	initialValue := p.AvailableCash()
	currentValue := initialValue

	for _, trade := range p.Transactions() {
		tradeValue := trade.Price * trade.Quantity
		if trade.Type == common.ActionBuy {
			currentValue -= tradeValue
		} else if trade.Type == common.ActionSell {
			currentValue += tradeValue
		}
		equityCurve = append(equityCurve, currentValue)
	}
	return equityCurve
}

func (b *Backtest) calculateMetrics(results *BacktestResults, p *portfolio.Portfolio) PerformanceMetrics {
	metrics := PerformanceMetrics{
		NumTrades: len(results.Trades),
	}

	if len(results.Trades) == 0 {
		return metrics
	}

	// Calculate returns and drawdowns
	var returns []float64
	var equityCurve []float64
	initialValue := p.GetInitialValue()
	currentValue := initialValue

	for _, trade := range results.Trades {
		tradeValue := trade.Price * trade.Quantity
		if trade.Type == common.ActionBuy {
			currentValue -= tradeValue
		} else if trade.Type == common.ActionSell {
			currentValue += tradeValue
		}
		equityCurve = append(equityCurve, currentValue)
		returns = append(returns, (currentValue-initialValue)/initialValue)
	}

	// Calculate metrics
	metrics.TotalReturn = (results.FinalValue - initialValue) / initialValue
	metrics.MaxDrawdown = calculateMaxDrawdown(equityCurve)
	metrics.WinRate = calculateWinRate(results.Trades)
	metrics.SharpeRatio = calculateSharpeRatio(returns)
	metrics.SortinoRatio = calculateSortinoRatio(returns)
	metrics.Volatility = calculateVolatility(returns)
	metrics.ProfitFactor = calculateProfitFactor(results.Trades)

	return metrics
}

func calculateMaxDrawdown(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0.0
	}

	peak := equityCurve[0]
	maxDrawdown := 0.0

	for _, value := range equityCurve {
		if value > peak {
			peak = value
		}
		drawdown := (peak - value) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown
}

func calculateWinRate(trades []common.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	wins := 0
	for _, trade := range trades {
		if trade.Type == common.ActionSell && trade.Price > 0 {
			wins++
		}
	}
	return float64(wins) / float64(len(trades))
}

func calculateSharpeRatio(returns []float64) float64 {
	if len(returns) == 0 {
		return 0.0
	}

	meanReturn := calculateMean(returns)
	stdDev := calculateStdDev(returns)

	// Assuming risk-free rate of 0 for simplicity
	return meanReturn / stdDev
}

func calculateSortinoRatio(returns []float64) float64 {
	if len(returns) == 0 {
		return 0.0
	}

	meanReturn := calculateMean(returns)
	downsideDev := calculateDownsideDeviation(returns)

	return meanReturn / downsideDev
}

func calculateVolatility(returns []float64) float64 {
	return calculateStdDev(returns)
}

func calculateProfitFactor(trades []common.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	grossProfit := 0.0
	grossLoss := 0.0

	for _, trade := range trades {
		if trade.Type == common.ActionSell {
			profit := trade.Price * trade.Quantity
			if profit > 0 {
				grossProfit += profit
			} else {
				grossLoss += -profit
			}
		}
	}

	if grossLoss == 0 {
		return 0.0
	}
	return grossProfit / grossLoss
}

// Helper functions
func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	mean := calculateMean(values)
	variance := 0.0

	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func calculateDownsideDeviation(returns []float64) float64 {
	if len(returns) == 0 {
		return 0.0
	}

	// Using 0 as the minimum acceptable return
	mar := 0.0
	sumSquares := 0.0

	for _, r := range returns {
		if r < mar {
			diff := r - mar
			sumSquares += diff * diff
		}
	}

	return math.Sqrt(sumSquares / float64(len(returns)))
}
