package types

import "time"

type LoginRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	TwoFactorCode string `json:"twoFactorCode,omitempty"`
}

type LoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         User      `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	EmailVerified    bool      `json:"emailVerified"`
	TwoFactorEnabled bool     `json:"twoFactorEnabled"`
	SubscriptionPlan string   `json:"subscriptionPlan"`
	IsAdmin          bool      `json:"isAdmin"`
	Timezone         string    `json:"timezone"`
	Locale           string    `json:"locale"`
	CreatedAt        time.Time `json:"createdAt"`
}

type APIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	LastUsed  *time.Time `json:"lastUsed,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type CreateAPIKeyRequest struct {
	Name      string    `json:"name"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type CreateAPIKeyResponse struct {
	APIKey APIKey `json:"apiKey"`
	RawKey string `json:"rawKey"`
}
