package backtest

import (
	"sort"
	"stock/broker"
	"stock/common/types"
	"stock/datasource"
	"stock/orders"
	"stock/portfolio"
	"stock/strategy"
	"time"
)

type Backtest struct {
	startDate   time.Time
	endDate     time.Time
	initialCash float64
	dataSource  datasource.DataSource
	strategies  []strategy.Strategy
	portfolios  []*portfolio.Portfolio
	broker      broker.Broker
	logger      types.Logger
	symbols     []string
}

type BacktestResult struct {
	StartDate   time.Time
	EndDate     time.Time
	InitialCash float64
	Results     []StrategyResult
}

type StrategyResult struct {
	Strategy    strategy.Strategy
	Portfolio   *portfolio.Portfolio
	FinalValue  float64
	Trades      []types.Trade
	EquityCurve []float64
	MaxDrawdown float64
}

func NewBacktest(startDate time.Time, endDate time.Time, initialCash float64, dataSource datasource.DataSource, broker broker.Broker, logger types.Logger, symbols []string) *Backtest {
	return &Backtest{
		startDate:   startDate,
		endDate:     endDate,
		initialCash: initialCash,
		dataSource:  dataSource,
		broker:      broker,
		logger:      logger,
		symbols:     symbols,
	}
}

func (b *Backtest) AddStrategy(strategy strategy.Strategy) {
	b.strategies = append(b.strategies, strategy)
	portfolio := portfolio.NewPortfolio(b.initialCash, b.broker, orders.NewOrderManager(b.broker))
	b.portfolios = append(b.portfolios, portfolio)
}

func (b *Backtest) Run() (*BacktestResult, error) {
	if len(b.strategies) == 0 {
		return nil, types.ErrNoStrategy
	}

	// Initialize strategies
	for index, strategy := range b.strategies {
		err := strategy.OnStart(b.portfolios[index])
		if err != nil {
			return nil, err
		}
	}

	// Main backtest loop
	equityCurves := make([][]float64, len(b.strategies))

	// Get data for all symbols
	allData := make([]*types.DataPoint, 0)
	for _, symbol := range b.symbols {
		data, err := b.dataSource.GetData(symbol, datasource.PeriodTypeDay, b.startDate, b.endDate)
		if err != nil {
			return nil, err
		}
		allData = append(allData, data...)
	}

	for index, strategy := range b.strategies {
		// Group data points by timestamp
		dataByTimestamp := make(map[time.Time][]*types.DataPoint)
		for _, d := range allData {
			dataByTimestamp[d.Timestamp] = append(dataByTimestamp[d.Timestamp], d)
		}

		// Sort timestamps and process data points in order
		sortedTimestamps := make([]time.Time, 0, len(dataByTimestamp))
		for timestamp := range dataByTimestamp {
			sortedTimestamps = append(sortedTimestamps, timestamp)
		}
		sort.Slice(sortedTimestamps, func(i, j int) bool {
			return sortedTimestamps[i].Before(sortedTimestamps[j])
		})

		for _, timestamp := range sortedTimestamps {
			dataPoints := dataByTimestamp[timestamp]
			err := strategy.OnData(dataPoints, b.portfolios[index])
			if err != nil {
				return nil, err
			}
			// Record daily portfolio value
			equityCurves[index] = append(equityCurves[index], b.portfolios[index].GetValue())
		}
	}

	// Finalize strategies
	for index, strategy := range b.strategies {
		b.logger.LogEnd(b.portfolios[index])
		err := strategy.OnEnd(b.portfolios[index], b.symbols[0])
		if err != nil {
			return nil, err
		}
	}

	// Calculate results
	results := make([]StrategyResult, len(b.strategies))
	for i := range b.strategies {
		results[i] = StrategyResult{
			Strategy:    b.strategies[i],
			Portfolio:   b.portfolios[i],
			FinalValue:  b.portfolios[i].GetValue(),
			Trades:      b.portfolios[i].Transactions(),
			EquityCurve: equityCurves[i],
			MaxDrawdown: calculateMaxDrawdown(equityCurves[i]),
		}
	}

	return &BacktestResult{
		StartDate:   b.startDate,
		EndDate:     b.endDate,
		InitialCash: b.initialCash,
		Results:     results,
	}, nil
}

func calculateMaxDrawdown(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0
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

func (b *Backtest) AnalyzeTrades() map[string]types.TradeAnalysis {
	analysis := make(map[string]types.TradeAnalysis)

	for _, portfolio := range b.portfolios {
		for _, trade := range portfolio.Transactions() {
			symbolAnalysis := analysis[trade.Symbol]
			symbolAnalysis.TotalTrades++

			if trade.Type == types.ActionBuy {
				symbolAnalysis.TotalBuy += trade.Price * trade.Quantity
			} else if trade.Type == types.ActionSell {
				symbolAnalysis.TotalSell += trade.Price * trade.Quantity
			}

			analysis[trade.Symbol] = symbolAnalysis
		}
	}

	return analysis
}
