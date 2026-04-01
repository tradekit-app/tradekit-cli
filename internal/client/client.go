package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tradekit-dev/tradekit-cli/internal/auth"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

var (
	ErrUnauthorized = fmt.Errorf("unauthorized — run: tradekit auth login")
	ErrForbidden    = fmt.Errorf("forbidden — this feature may require a higher plan")
	ErrNotFound     = fmt.Errorf("not found")
	ErrRateLimited  = fmt.Errorf("rate limited — try again shortly")
	ErrServer       = fmt.Errorf("server error — try again later")
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthStore  *auth.Store
	UserAgent  string
}

func New(baseURL string, store *auth.Store, version string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		AuthStore: store,
		UserAgent: "tradekit-cli/" + version,
	}
}

func Do[T any](ctx context.Context, c *Client, method, path string, body any, params url.Values) (*types.Response[T], error) {
	resp, err := c.doRequest(ctx, method, path, body, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle 401 with token refresh
	if resp.StatusCode == http.StatusUnauthorized && c.AuthStore.HasRefreshToken() {
		if refreshErr := c.refreshTokens(ctx); refreshErr == nil {
			resp.Body.Close()
			resp, err = c.doRequest(ctx, method, path, body, params)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-JSON errors (e.g., 502/503 from load balancer)
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") && resp.StatusCode >= 400 {
		return nil, mapStatusError(resp.StatusCode, string(data))
	}

	var apiResp types.Response[T]
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success && apiResp.Error != nil {
		return &apiResp, mapAPIError(resp.StatusCode, apiResp.Error)
	}

	return &apiResp, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, params url.Values) (*http.Response, error) {
	u := c.BaseURL + path
	if params != nil {
		u += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Inject auth
	if token := c.getAuthToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Retry on 429 with backoff
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt == maxRetries {
				return nil, ErrRateLimited
			}
			wait := time.Duration(attempt+1) * 2 * time.Second
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if secs, err := strconv.Atoi(retryAfter); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			time.Sleep(wait)
			continue
		}

		// Retry on 5xx
		if resp.StatusCode >= 500 && attempt < maxRetries {
			resp.Body.Close()
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("request failed after retries")
}

func (c *Client) getAuthToken() string {
	if c.AuthStore == nil {
		return ""
	}
	return c.AuthStore.GetToken()
}

func (c *Client) refreshTokens(ctx context.Context) error {
	creds, err := c.AuthStore.Load()
	if err != nil || creds.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	refreshReq := types.RefreshRequest{RefreshToken: creds.RefreshToken}
	body, _ := json.Marshal(refreshReq)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed with status %d", resp.StatusCode)
	}

	var apiResp types.Response[types.RefreshResponse]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}

	if !apiResp.Success {
		return fmt.Errorf("refresh failed")
	}

	creds.AccessToken = apiResp.Data.AccessToken
	creds.RefreshToken = apiResp.Data.RefreshToken
	creds.ExpiresAt = apiResp.Data.ExpiresAt
	return c.AuthStore.Save(creds)
}

func mapAPIError(status int, errBody *types.ErrorBody) error {
	msg := errBody.Message
	if msg == "" {
		msg = errBody.Code
	}

	switch {
	case status == http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrUnauthorized, msg)
	case status == http.StatusForbidden:
		if strings.Contains(msg, "API key authentication") {
			return fmt.Errorf("this feature requires API key authentication\n  Create one: tradekit auth apikey create")
		}
		if strings.Contains(msg, "subscription plan") {
			return fmt.Errorf("this feature requires a Pro plan\n  Upgrade: https://tradekit.com.br/pricing")
		}
		return fmt.Errorf("%w: %s", ErrForbidden, msg)
	case status == http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, msg)
	case status == http.StatusTooManyRequests:
		return ErrRateLimited
	case status == http.StatusBadRequest:
		if len(errBody.Details) > 0 {
			var parts []string
			for field, detail := range errBody.Details {
				parts = append(parts, fmt.Sprintf("  %s: %s", field, detail))
			}
			return fmt.Errorf("validation error:\n%s", strings.Join(parts, "\n"))
		}
		return fmt.Errorf("bad request: %s", msg)
	case status >= 500:
		return fmt.Errorf("%w: %s", ErrServer, msg)
	default:
		return fmt.Errorf("API error (%d): %s", status, msg)
	}
}

func mapStatusError(status int, body string) error {
	switch {
	case status == http.StatusUnauthorized:
		return ErrUnauthorized
	case status == http.StatusForbidden:
		return ErrForbidden
	case status == http.StatusTooManyRequests:
		return ErrRateLimited
	case status >= 500:
		return ErrServer
	default:
		return fmt.Errorf("HTTP %d: %s", status, body)
	}
}
