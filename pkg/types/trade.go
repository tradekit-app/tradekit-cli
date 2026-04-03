package types

import "time"

type Trade struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	TradingAccountID string    `json:"tradingAccountId"`
	Symbol           string    `json:"symbol"`
	AssetType        string    `json:"assetType"`
	Direction        string    `json:"direction"`
	EntryDate        time.Time `json:"entryDate"`
	EntryPrice       string    `json:"entryPrice"`
	EntryQuantity    string    `json:"entryQuantity"`
	ExitDate         *time.Time `json:"exitDate,omitempty"`
	ExitPrice        *string   `json:"exitPrice,omitempty"`
	ExitQuantity     *string   `json:"exitQuantity,omitempty"`
	StopLoss         *string   `json:"stopLoss,omitempty"`
	TakeProfit       *string   `json:"takeProfit,omitempty"`
	Commission       *string   `json:"commission,omitempty"`
	Swap             *string   `json:"swap,omitempty"`
	OtherCosts       *string   `json:"otherCosts,omitempty"`
	GrossPnl         *string   `json:"grossPnl,omitempty"`
	NetPnl           *string   `json:"netPnl,omitempty"`
	PnlPercentage    *string   `json:"pnlPercentage,omitempty"`
	RiskRewardRatio  *string   `json:"riskRewardRatio,omitempty"`
	Setup            string    `json:"setup,omitempty"`
	Timeframe        string    `json:"timeframe,omitempty"`
	Strategy         string    `json:"strategy,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	LessonsLearned   string    `json:"lessonsLearned,omitempty"`
	EmotionBefore    string    `json:"emotionBefore,omitempty"`
	EmotionAfter     string    `json:"emotionAfter,omitempty"`
	ConfidenceLevel  int       `json:"confidenceLevel,omitempty"`
	Status           string    `json:"status"`
	Tags             []Tag     `json:"tags,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type TradingAccount struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	Name           string    `json:"name"`
	Broker         string    `json:"broker"`
	AccountNumber  string    `json:"accountNumber,omitempty"`
	IsDemo         bool      `json:"isDemo"`
	IsActive       bool      `json:"isActive"`
	InitialBalance string    `json:"initialBalance"`
	Currency       string    `json:"currency"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type AccountState struct {
	AvailableFunds   string                `json:"availableFunds"`
	TotalDeposits    string                `json:"totalDeposits"`
	TotalWithdrawals string                `json:"totalWithdrawals"`
	RealizedPnl      string                `json:"realizedPnl"`
	CapitalInUse     string                `json:"capitalInUse"`
	OpenPositions    []OpenPositionSummary `json:"openPositions"`
}

type OpenPositionSummary struct {
	Symbol    string `json:"symbol"`
	Direction string `json:"direction"`
	Quantity  string `json:"quantity"`
	AvgPrice  string `json:"avgPrice"`
}

type Tag struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	IsSystem bool   `json:"isSystem"`
}

type TradeStats struct {
	TotalTrades   int64   `json:"totalTrades"`
	WinningTrades int64   `json:"winningTrades"`
	LosingTrades  int64   `json:"losingTrades"`
	GrossPnl      float64 `json:"grossPnl"`
	NetPnl        float64 `json:"netPnl"`
	TotalVolume   float64 `json:"totalVolume"`
	WinRate       string  `json:"winRate"`
	TotalPnl      string  `json:"totalPnl"`
	AverageWin    string  `json:"averageWin"`
	AverageLoss   string  `json:"averageLoss"`
	ProfitFactor  string  `json:"profitFactor"`
	MaxDrawdown   string  `json:"maxDrawdown"`
	SharpeRatio   string  `json:"sharpeRatio"`
	Expectancy    string  `json:"expectancy"`
}

type CreateSignalRequest struct {
	Symbol      string  `json:"symbol"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Price       string  `json:"price"`
	Quantity    *string `json:"quantity,omitempty"`
	ScheduledAt *string `json:"scheduledAt,omitempty"`
}

type Signal struct {
	ID               string  `json:"id"`
	Symbol           string  `json:"symbol"`
	Type             string  `json:"type"`
	RuleName         string  `json:"ruleName"`
	Description      string  `json:"description"`
	Price            string  `json:"price"`
	Status           string  `json:"status"`
	ScheduledAt      string  `json:"scheduledAt,omitempty"`
	ExpiresAt        string  `json:"expiresAt,omitempty"`
	ExecutionTicket  *int64  `json:"executionTicket,omitempty"`
	ExecutionPrice   string  `json:"executionPrice,omitempty"`
	ExecutionVolume  string  `json:"executionVolume,omitempty"`
	ExecutionSuccess *bool   `json:"executionSuccess,omitempty"`
	ExecutionError   string  `json:"executionError,omitempty"`
	ExecutedAt       string  `json:"executedAt,omitempty"`
	CreatedAt        string  `json:"createdAt"`
}

type QuickTradeRequest struct {
	Symbol     string  `json:"symbol"`
	Direction  string  `json:"direction"`
	Price      string  `json:"price"`
	Quantity   string  `json:"quantity"`
	AccountID  *string `json:"accountId,omitempty"`
	StopLoss   *string `json:"stopLoss,omitempty"`
	TakeProfit *string `json:"takeProfit,omitempty"`
	Notes      *string `json:"notes,omitempty"`
}

type TodayResponse struct {
	Trades      []Trade     `json:"trades"`
	Stats       TradeStats  `json:"stats"`
	TotalTrades int64       `json:"totalTrades"`
	Date        string      `json:"date"`
}

type Position struct {
	Symbol        string  `json:"symbol"`
	AssetType     string  `json:"assetType"`
	Direction     string  `json:"direction"`
	TradeCount    int64   `json:"tradeCount"`
	TotalQuantity float64 `json:"totalQuantity"`
	AvgEntryPrice float64 `json:"avgEntryPrice"`
	TotalCosts    float64 `json:"totalCosts"`
}

type CreateRiskRuleRequest struct {
	TradingAccountID string         `json:"tradingAccountId"`
	StrategyID       *string        `json:"strategyId,omitempty"`
	Scope            string         `json:"scope"`
	RuleType         string         `json:"ruleType"`
	Params           map[string]any `json:"params"`
}

type RiskRule struct {
	ID               string         `json:"id"`
	TradingAccountID string         `json:"tradingAccountId"`
	Scope            string         `json:"scope"`
	RuleType         string         `json:"ruleType"`
	Params           map[string]any `json:"params"`
	Enabled          bool           `json:"enabled"`
	CreatedAt        string         `json:"createdAt"`
}

type RiskViolation struct {
	ID          string `json:"id"`
	RuleType    string `json:"ruleType"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

type MT5AccountInfo struct {
	Balance    float64 `json:"balance"`
	Equity     float64 `json:"equity"`
	Margin     float64 `json:"margin"`
	FreeMargin float64 `json:"freeMargin"`
	Leverage   int     `json:"leverage"`
	Currency   string  `json:"currency"`
}

type MT5Position struct {
	Ticket     int64   `json:"ticket"`
	Symbol     string  `json:"symbol"`
	Type       string  `json:"type"`
	Volume     float64 `json:"volume"`
	OpenPrice  float64 `json:"openPrice"`
	StopLoss   float64 `json:"stopLoss"`
	TakeProfit float64 `json:"takeProfit"`
	Profit     float64 `json:"profit"`
}

type MT5AccountResponse struct {
	Account   *MT5AccountInfo `json:"account"`
	Positions []MT5Position   `json:"positions"`
}

type DashboardResponse struct {
	MonthStats    TradeStats `json:"monthStats"`
	TodayStats    TradeStats `json:"todayStats"`
	OpenPositions []Position `json:"openPositions"`
	PositionCount int        `json:"positionCount"`
	RecentTrades  []Trade    `json:"recentTrades"`
}

type EnrichedPosition struct {
	Symbol        string  `json:"symbol"`
	Direction     string  `json:"direction"`
	TotalQuantity float64 `json:"totalQuantity"`
	AvgEntryPrice float64 `json:"avgEntryPrice"`
	CurrentPrice  string  `json:"currentPrice,omitempty"`
}

type PortfolioResponse struct {
	Positions []EnrichedPosition `json:"positions"`
}
