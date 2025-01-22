package backtest

import (
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
}

func NewBacktest(startDate time.Time, endDate time.Time, initialCash float64, dataSource datasource.DataSource, broker broker.Broker, logger types.Logger) *Backtest {
	return &Backtest{
		startDate:   startDate,
		endDate:     endDate,
		initialCash: initialCash,
		dataSource:  dataSource,
		broker:      broker,
		logger:      logger,
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

	data, err := b.dataSource.GetData("", b.startDate, b.endDate)
	if err != nil {
		return nil, err
	}

	for index, strategy := range b.strategies {
		for _, d := range data {
			err := strategy.OnData(d, b.portfolios[index])
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
		err := strategy.OnEnd(b.portfolios[index])
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
		}
	}

	return &BacktestResult{
		StartDate:   b.startDate,
		EndDate:     b.endDate,
		InitialCash: b.initialCash,
		Results:     results,
	}, nil
}

func (b *Backtest) AnalyzeTrades() (float64, float64, float64) {
	var totalProfit float64
	var totalLoss float64
	var totalTrades int

	for _, portfolio := range b.portfolios {
		for _, trade := range portfolio.Transactions() {
			totalTrades++
			if trade.Type == types.ActionBuy {
				totalProfit += trade.Price * trade.Quantity
			} else if trade.Type == types.ActionSell {
				totalLoss += trade.Price * trade.Quantity
			}
		}
	}

	return totalProfit, totalLoss, float64(totalTrades)
}

func (b *Backtest) Results() *BacktestResult {
	results := make([]StrategyResult, len(b.strategies))
	for i := range b.strategies {
		results[i] = StrategyResult{
			Strategy:    b.strategies[i],
			Portfolio:   b.portfolios[i],
			FinalValue:  b.portfolios[i].GetValue(),
			Trades:      b.portfolios[i].Transactions(),
			EquityCurve: []float64{},
		}
	}

	return &BacktestResult{
		StartDate:   b.startDate,
		EndDate:     b.endDate,
		InitialCash: b.initialCash,
		Results:     results,
	}
}
