package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Paper-monitor / lifecycle endpoints (services/strategy PR-7).
// Responses are decoded as map[string]any rather than typed structs so
// the CLI can print them without forcing the types package to track every
// field on the platform side. The CLI's `--output json` path passes the
// raw map through unchanged, which is the most useful shape for piping
// to jq.

// PromoteToPaper calls POST /api/strategies/{id}/promote-paper. Body is
// optional — if all fields zero, server applies kind-aware defaults.
type PromoteToPaperBody struct {
	TradingAccountID   string `json:"tradingAccountId,omitempty"`
	InitialEquity      *float64 `json:"initialEquity,omitempty"`
	ActivationCriteria any    `json:"activationCriteria,omitempty"`
	KillCriteria       any    `json:"killCriteria,omitempty"`
}

func (c *Client) PromoteToPaper(ctx context.Context, id string, body *PromoteToPaperBody) (map[string]any, error) {
	resp, err := Do[map[string]any](ctx, c, http.MethodPost,
		"/api/strategies/"+id+"/promote-paper", body, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// PromoteToLive calls POST /api/strategies/{id}/promote-live.
func (c *Client) PromoteToLive(ctx context.Context, id string) (map[string]any, error) {
	resp, err := Do[map[string]any](ctx, c, http.MethodPost,
		"/api/strategies/"+id+"/promote-live", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// KillPaper calls POST /api/strategies/{id}/kill with a reason in body.
func (c *Client) KillPaper(ctx context.Context, id, reason string) (map[string]any, error) {
	body := map[string]string{"reason": reason}
	resp, err := Do[map[string]any](ctx, c, http.MethodPost,
		"/api/strategies/"+id+"/kill", body, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// RunTick calls POST /api/strategies/{id}/tick.
// asOfDate is "YYYY-MM-DD" or "" for today; v1 only honors today.
func (c *Client) RunTick(ctx context.Context, id, asOfDate string) (map[string]any, error) {
	var body any
	if asOfDate != "" {
		body = map[string]string{"asOfDate": asOfDate}
	}
	resp, err := Do[map[string]any](ctx, c, http.MethodPost,
		"/api/strategies/"+id+"/tick", body, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ListTicks calls GET /api/strategies/{id}/ticks?limit=N.
func (c *Client) ListTicks(ctx context.Context, id string, limit int) ([]map[string]any, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	resp, err := Do[[]map[string]any](ctx, c, http.MethodGet,
		"/api/strategies/"+id+"/ticks", nil, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetState calls GET /api/strategies/{id}/state.
func (c *Client) GetState(ctx context.Context, id string) (map[string]any, error) {
	resp, err := Do[map[string]any](ctx, c, http.MethodGet,
		"/api/strategies/"+id+"/state", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
