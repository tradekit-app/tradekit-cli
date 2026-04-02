package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

type ListTradesOptions struct {
	Page    int
	PerPage int
	Status  string
	Symbol  string
	From    string
	To      string
	Account string
}

func (c *Client) ListTrades(ctx context.Context, opts ListTradesOptions) (*types.Response[[]types.Trade], error) {
	params := url.Values{}
	if opts.Page > 0 {
		params.Set("page", fmt.Sprintf("%d", opts.Page))
	}
	if opts.PerPage > 0 {
		params.Set("per_page", fmt.Sprintf("%d", opts.PerPage))
	}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	if opts.Symbol != "" {
		params.Set("symbol", opts.Symbol)
	}
	if opts.From != "" {
		params.Set("from", opts.From)
	}
	if opts.To != "" {
		params.Set("to", opts.To)
	}
	if opts.Account != "" {
		params.Set("trading_account_id", opts.Account)
	}
	return Do[[]types.Trade](ctx, c, http.MethodGet, "/api/trades/", nil, params)
}

func (c *Client) GetTrade(ctx context.Context, id string) (*types.Trade, error) {
	resp, err := Do[types.Trade](ctx, c, http.MethodGet, "/api/trades/"+id, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetPositions(ctx context.Context) ([]types.Trade, error) {
	resp, err := Do[[]types.Trade](ctx, c, http.MethodGet, "/api/trades/positions", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetTradeStats(ctx context.Context) (*types.TradeStats, error) {
	resp, err := Do[types.TradeStats](ctx, c, http.MethodGet, "/api/trades/stats", nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) ExportTrades(ctx context.Context, opts ListTradesOptions) ([]types.Trade, error) {
	params := url.Values{}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	if opts.Symbol != "" {
		params.Set("symbol", opts.Symbol)
	}
	if opts.From != "" {
		params.Set("from", opts.From)
	}
	if opts.To != "" {
		params.Set("to", opts.To)
	}
	if opts.Account != "" {
		params.Set("account_id", opts.Account)
	}
	resp, err := Do[[]types.Trade](ctx, c, http.MethodGet, "/api/trades/export", nil, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetTodayTrades(ctx context.Context) (*types.TodayResponse, error) {
	resp, err := Do[types.TodayResponse](ctx, c, http.MethodGet, "/api/trades/today", nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) QuickCreateTrade(ctx context.Context, req types.QuickTradeRequest) (*types.Trade, error) {
	resp, err := Do[types.Trade](ctx, c, http.MethodPost, "/api/trades/quick", req, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetPortfolio(ctx context.Context) (*types.PortfolioResponse, error) {
	// Get open positions
	dashResp, err := c.GetDashboard(ctx)
	if err != nil {
		return nil, err
	}

	if len(dashResp.OpenPositions) == 0 {
		return &types.PortfolioResponse{Positions: nil}, nil
	}

	// Get quotes for all position symbols
	var symbols []string
	for _, p := range dashResp.OpenPositions {
		symbols = append(symbols, p.Symbol)
	}

	quotes := make(map[string]*types.Quote)
	// Try batch first, fall back to individual
	quotesResp, err := c.GetQuotes(ctx, symbols)
	if err == nil && quotesResp != nil {
		for i := range quotesResp.Quotes {
			quotes[quotesResp.Quotes[i].Symbol] = &quotesResp.Quotes[i]
		}
	} else {
		// Fallback: fetch individually
		for _, sym := range symbols {
			q, qErr := c.GetQuote(ctx, sym)
			if qErr == nil {
				quotes[sym] = q
			}
		}
	}

	// Enrich positions
	var enriched []types.EnrichedPosition
	for _, p := range dashResp.OpenPositions {
		ep := types.EnrichedPosition{
			Symbol:        p.Symbol,
			Direction:     p.Direction,
			TotalQuantity: p.TotalQuantity,
			AvgEntryPrice: p.AvgEntryPrice,
		}
		if q, ok := quotes[p.Symbol]; ok {
			ep.CurrentPrice = q.Price
			// Calculate unrealized P&L: (current - entry) * qty * direction_multiplier
			// Leave the calculation to the formatter since prices are strings
		}
		enriched = append(enriched, ep)
	}

	return &types.PortfolioResponse{Positions: enriched}, nil
}

func (c *Client) CreateSignal(ctx context.Context, req types.CreateSignalRequest) (*types.Signal, error) {
	resp, err := Do[types.Signal](ctx, c, http.MethodPost, "/api/signals", req, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) ListPendingSignals(ctx context.Context) ([]types.Signal, error) {
	type pendingResp struct {
		Signals []types.Signal `json:"signals"`
	}
	resp, err := Do[pendingResp](ctx, c, http.MethodGet, "/api/signals/pending", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data.Signals, nil
}

func (c *Client) GetDashboard(ctx context.Context) (*types.DashboardResponse, error) {
	resp, err := Do[types.DashboardResponse](ctx, c, http.MethodGet, "/api/trades/dashboard", nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
