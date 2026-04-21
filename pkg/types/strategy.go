package types

type Strategy struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags"`
	Symbols  []string `json:"symbols"`
	IsActive bool     `json:"isActive"`
}

type StrategyLivePerformance struct {
	TotalTrades   int     `json:"totalTrades"`
	WinRate       float64 `json:"winRate"`
	TotalPnl      float64 `json:"totalPnl"`
	ProfitFactor  float64 `json:"profitFactor"`
}

type BacktestResult struct {
	ID               string  `json:"id"`
	TotalReturn      string  `json:"totalReturn"`
	WinRate          string  `json:"winRate"`
	ProfitFactor     string  `json:"profitFactor"`
	MaxDrawdown      string  `json:"maxDrawdown"`
	SharpeRatio      string  `json:"sharpeRatio"`
	EntrySignals    int     `json:"entrySignals"`
	SimulatedTrades []any   `json:"simulatedTrades"`
	AnnualizedReturn string `json:"annualizedReturn"`
}

type BacktestsResponse struct {
	Backtests []BacktestResult `json:"backtests"`
}

type LeaderboardEntry struct {
	Rank         int     `json:"rank"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Source       string  `json:"source"` // "live" or "backtest"
	Trades       int     `json:"trades"`
	WinRate      float64 `json:"winRate"` // 0..1
	TotalPnl     float64 `json:"totalPnl"`
	ProfitFactor float64 `json:"profitFactor"`
}
