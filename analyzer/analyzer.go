package analyzer

import (
	"math"
	"sort"
	"stock/common/types"
	"time"
)

type Analyzer struct {
	trades      []types.Trade
	initialCash float64
}

func NewAnalyzer(trades []types.Trade, initialCash float64) *Analyzer {
	return &Analyzer{
		trades:      trades,
		initialCash: initialCash,
	}
}

// 计算总收益率
func (a *Analyzer) TotalReturn(finalValue float64) float64 {
	return (finalValue - a.initialCash) / a.initialCash
}

// 计算年化收益率
func (a *Analyzer) AnnualizedReturn(finalValue float64, duration time.Duration) float64 {
	totalReturn := a.TotalReturn(finalValue)
	years := duration.Hours() / (24 * 365)
	return math.Pow(1+totalReturn, 1/years) - 1
}

// 计算最大回撤
func (a *Analyzer) MaxDrawdown(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	peak := values[0]
	maxDrawdown := 0.0

	for _, value := range values {
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

// 计算夏普比率
func (a *Analyzer) SharpeRatio(returns []float64, riskFreeRate float64) float64 {
	meanReturn := a.Mean(returns)
	stdDev := a.StandardDeviation(returns)
	return (meanReturn - riskFreeRate) / stdDev
}

// 计算平均收益率
func (a *Analyzer) Mean(returns []float64) float64 {
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	return sum / float64(len(returns))
}

// 计算标准差
func (a *Analyzer) StandardDeviation(returns []float64) float64 {
	mean := a.Mean(returns)
	variance := 0.0
	for _, r := range returns {
		diff := r - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))
	return math.Sqrt(variance)
}

// 计算胜率
func (a *Analyzer) WinRate() float64 {
	winCount := 0
	for _, trade := range a.trades {
		if trade.Type == types.ActionSell && trade.Price > 0 {
			winCount++
		}
	}
	return float64(winCount) / float64(len(a.trades))
}

// 计算平均盈利/亏损
func (a *Analyzer) AverageProfitLoss() (float64, float64) {
	var totalProfit, totalLoss float64
	var profitCount, lossCount int

	// 按交易对计算盈亏
	for i := 0; i < len(a.trades)-1; i += 2 {
		buy := a.trades[i]
		sell := a.trades[i+1]

		if buy.Type == types.ActionBuy && sell.Type == types.ActionSell {
			profit := sell.Price - buy.Price
			if profit > 0 {
				totalProfit += profit
				profitCount++
			} else {
				totalLoss += math.Abs(profit)
				lossCount++
			}
		}
	}

	avgProfit := 0.0
	if profitCount > 0 {
		avgProfit = totalProfit / float64(profitCount)
	}

	avgLoss := 0.0
	if lossCount > 0 {
		avgLoss = totalLoss / float64(lossCount)
	}

	return avgProfit, avgLoss
}

// 计算盈亏比
func (a *Analyzer) ProfitLossRatio() float64 {
	avgProfit, avgLoss := a.AverageProfitLoss()
	if avgLoss == 0 {
		return 0
	}
	return avgProfit / avgLoss
}

// 计算年波动率
func (a *Analyzer) AnnualVolatility(returns []float64) float64 {
	stdDev := a.StandardDeviation(returns)
	return stdDev * math.Sqrt(252) // 252个交易日
}

// 计算索提诺比率
func (a *Analyzer) SortinoRatio(returns []float64, targetReturn float64) float64 {
	downsideReturns := make([]float64, 0)
	for _, r := range returns {
		if r < targetReturn {
			downsideReturns = append(downsideReturns, r)
		}
	}
	downsideDev := a.StandardDeviation(downsideReturns)
	if downsideDev == 0 {
		return 0
	}
	return (a.Mean(returns) - targetReturn) / downsideDev
}

// 计算卡尔玛比率
func (a *Analyzer) CalmarRatio(finalValue float64, maxDrawdown float64, duration time.Duration) float64 {
	if maxDrawdown == 0 {
		return 0
	}
	annualizedReturn := a.AnnualizedReturn(finalValue, duration)
	return annualizedReturn / maxDrawdown
}

// 计算回撤持续时间
func (a *Analyzer) DrawdownDuration(values []float64) time.Duration {
	if len(values) == 0 {
		return 0
	}

	var maxDuration time.Duration
	var currentDuration time.Duration
	peak := values[0]

	for i := 1; i < len(values); i++ {
		if values[i] > peak {
			peak = values[i]
			currentDuration = 0
		} else {
			currentDuration += time.Duration(1) * time.Hour * 24 // 假设每天一个数据点
		}
		if currentDuration > maxDuration {
			maxDuration = currentDuration
		}
	}

	return maxDuration
}

// 计算风险价值 (VaR)
func (a *Analyzer) ValueAtRisk(returns []float64, confidenceLevel float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// 对收益率排序
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sort.Float64s(sortedReturns)

	// 计算VaR
	index := int(math.Floor(float64(len(sortedReturns)) * (1 - confidenceLevel)))
	if index >= len(sortedReturns) {
		index = len(sortedReturns) - 1
	}
	return sortedReturns[index]
}
