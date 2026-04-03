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

type ListSignalsOptions struct {
	Status string
	Symbol string
	Limit  int
}

func (c *Client) ListAllSignals(ctx context.Context, opts ListSignalsOptions) ([]types.Signal, error) {
	params := url.Values{}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	if opts.Symbol != "" {
		params.Set("symbol", opts.Symbol)
	}
	limit := 20
	if opts.Limit > 0 {
		limit = opts.Limit
	}
	params.Set("per_page", fmt.Sprintf("%d", limit))

	resp, err := Do[[]types.Signal](ctx, c, http.MethodGet, "/api/signals", nil, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListRiskRules(ctx context.Context, accountID string) ([]types.RiskRule, error) {
	params := url.Values{}
	if accountID != "" {
		params.Set("account_id", accountID)
	}
	resp, err := Do[[]types.RiskRule](ctx, c, http.MethodGet, "/api/risk-rules", nil, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateRiskRule(ctx context.Context, req types.CreateRiskRuleRequest) (*types.RiskRule, error) {
	resp, err := Do[types.RiskRule](ctx, c, http.MethodPost, "/api/risk-rules", req, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteRiskRule(ctx context.Context, id string) error {
	_, err := Do[any](ctx, c, http.MethodDelete, "/api/risk-rules/"+id, nil, nil)
	return err
}

func (c *Client) ListRiskViolations(ctx context.Context) ([]types.RiskViolation, error) {
	resp, err := Do[[]types.RiskViolation](ctx, c, http.MethodGet, "/api/risk-violations", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetMT5Account(ctx context.Context, connectionID string) (*types.MT5AccountResponse, error) {
	params := url.Values{}
	params.Set("connectionId", connectionID)
	resp, err := Do[types.MT5AccountResponse](ctx, c, http.MethodGet, "/api/mt5/account", nil, params)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetSignal(ctx context.Context, id string) (*types.Signal, error) {
	resp, err := Do[types.Signal](ctx, c, http.MethodGet, "/api/signals/"+id, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetDashboard(ctx context.Context) (*types.DashboardResponse, error) {
	resp, err := Do[types.DashboardResponse](ctx, c, http.MethodGet, "/api/trades/dashboard", nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
