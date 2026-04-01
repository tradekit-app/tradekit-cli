package client

import (
	"context"
	"net/http"

	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

func (c *Client) Login(ctx context.Context, req types.LoginRequest) (*types.LoginResponse, error) {
	resp, err := Do[types.LoginResponse](ctx, c, http.MethodPost, "/api/auth/login", req, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetMe(ctx context.Context) (*types.User, error) {
	resp, err := Do[types.User](ctx, c, http.MethodGet, "/api/auth/me", nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) Logout(ctx context.Context) error {
	_, err := Do[any](ctx, c, http.MethodPost, "/api/auth/logout", nil, nil)
	return err
}

func (c *Client) CreateAPIKey(ctx context.Context, req types.CreateAPIKeyRequest) (*types.CreateAPIKeyResponse, error) {
	resp, err := Do[types.CreateAPIKeyResponse](ctx, c, http.MethodPost, "/api/auth/api-keys", req, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) ListAPIKeys(ctx context.Context) ([]types.APIKey, error) {
	resp, err := Do[[]types.APIKey](ctx, c, http.MethodGet, "/api/auth/api-keys", nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := Do[any](ctx, c, http.MethodDelete, "/api/auth/api-keys/"+id, nil, nil)
	return err
}
