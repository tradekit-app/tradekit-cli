package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

type strategyListData struct {
	Data []types.Strategy `json:"data"`
}

func (c *Client) ListStrategies(ctx context.Context, perPage int) ([]types.Strategy, error) {
	params := url.Values{}
	if perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}
	// The API wraps the list in {success,data:[...]} — Do[T] already unwraps
	// one level into resp.Data, so we use []Strategy directly.
	resp, err := Do[[]types.Strategy](ctx, c, http.MethodGet, "/api/strategies", nil, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetStrategyLivePerformance(ctx context.Context, id string, accountID string) (*types.StrategyLivePerformance, error) {
	var params url.Values
	if accountID != "" {
		params = url.Values{"tradingAccountId": []string{accountID}}
	}
	resp, err := Do[types.StrategyLivePerformance](ctx, c, http.MethodGet,
		"/api/trades/strategy/"+id+"/performance", nil, params)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetBacktests(ctx context.Context, id string) ([]types.BacktestResult, error) {
	resp, err := Do[types.BacktestsResponse](ctx, c, http.MethodGet,
		"/api/strategies/"+id+"/backtests", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data.Backtests, nil
}

// TradingAccount mirrors the API shape for an account.
type TradingAccount struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Broker   string `json:"broker"`
	IsDemo   bool   `json:"isDemo"`
	IsActive bool   `json:"isActive"`
}

type tradingAccountsResponse struct {
	Accounts []TradingAccount `json:"accounts"`
}

func (c *Client) ListTradingAccounts(ctx context.Context) ([]TradingAccount, error) {
	resp, err := Do[tradingAccountsResponse](ctx, c, http.MethodGet, "/api/trading-accounts", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data.Accounts, nil
}
