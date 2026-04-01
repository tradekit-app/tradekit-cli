package types

import "time"

type Quote struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	Exchange         string `json:"exchange"`
	Currency         string `json:"currency"`
	Price            string `json:"price"`
	Change           string `json:"change"`
	ChangePercent    string `json:"changePercent"`
	Open             string `json:"open"`
	High             string `json:"high"`
	Low              string `json:"low"`
	PreviousClose    string `json:"previousClose"`
	Volume           int64  `json:"volume"`
	MarketCap        int64  `json:"marketCap"`
	FiftyTwoWeekHigh string `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow  string `json:"fiftyTwoWeekLow"`
	Timestamp        string `json:"timestamp"`
}

type OHLCV struct {
	Date     time.Time `json:"date"`
	Open     string    `json:"open"`
	High     string    `json:"high"`
	Low      string    `json:"low"`
	Close    string    `json:"close"`
	AdjClose string    `json:"adjClose"`
	Volume   int64     `json:"volume"`
}

type HistoricalData struct {
	Symbol   string  `json:"symbol"`
	Currency string  `json:"currency"`
	Data     []OHLCV `json:"data"`
}

type QuotesResponse struct {
	Quotes []Quote             `json:"quotes"`
	Errors []QuoteError        `json:"errors,omitempty"`
}

type QuoteError struct {
	Symbol  string `json:"symbol"`
	Message string `json:"message"`
}

type SearchResult struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	Type     string `json:"type"`
}
