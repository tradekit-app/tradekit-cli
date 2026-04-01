package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

func (c *Client) GetQuotes(ctx context.Context, symbols []string) (*types.QuotesResponse, error) {
	params := url.Values{}
	params.Set("symbols", strings.Join(symbols, ","))
	resp, err := Do[types.QuotesResponse](ctx, c, http.MethodGet, "/api/market-data/quotes", nil, params)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetQuote(ctx context.Context, symbol string) (*types.Quote, error) {
	resp, err := Do[types.Quote](ctx, c, http.MethodGet, "/api/market-data/quote/"+symbol, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) SearchSymbols(ctx context.Context, query string, limit int) ([]types.SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	resp, err := Do[[]types.SearchResult](ctx, c, http.MethodGet, "/api/market-data/search", nil, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type HistoryOptions struct {
	Interval string // 1m, 5m, 15m, 30m, 1h, 1d, 1wk, 1mo
	Range    string // 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, max
}

func (c *Client) GetHistory(ctx context.Context, symbol string, opts HistoryOptions) (*types.HistoricalData, error) {
	params := url.Values{}
	if opts.Interval != "" {
		params.Set("interval", opts.Interval)
	}
	if opts.Range != "" {
		params.Set("range", opts.Range)
	}
	resp, err := Do[types.HistoricalData](ctx, c, http.MethodGet, "/api/market-data/history/"+symbol, nil, params)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
