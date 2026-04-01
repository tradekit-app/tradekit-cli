package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Credentials struct {
	AccessToken  string    `yaml:"access_token"`
	RefreshToken string    `yaml:"refresh_token"`
	ExpiresAt    time.Time `yaml:"expires_at"`
	UserID       string    `yaml:"user_id"`
	Email        string    `yaml:"email"`
	Plan         string    `yaml:"plan"`
	APIKey       string    `yaml:"api_key"`
}

type Store struct {
	dir  string
	path string
}

func NewStore(dir string) *Store {
	return &Store{
		dir:  dir,
		path: filepath.Join(dir, "credentials"),
	}
}

func (s *Store) Load() (*Credentials, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}
	return &creds, nil
}

func (s *Store) Save(creds *Credentials) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}
	return nil
}

func (s *Store) Clear() error {
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}
	return nil
}

func (s *Store) GetToken() string {
	creds, err := s.Load()
	if err != nil {
		return ""
	}
	if creds.APIKey != "" {
		return creds.APIKey
	}
	if creds.AccessToken != "" {
		return creds.AccessToken
	}
	return ""
}

func (s *Store) IsLoggedIn() bool {
	creds, err := s.Load()
	if err != nil {
		return false
	}
	return creds.AccessToken != "" || creds.APIKey != ""
}

func (s *Store) IsExpired() bool {
	creds, err := s.Load()
	if err != nil {
		return true
	}
	if creds.APIKey != "" {
		return false
	}
	return time.Now().After(creds.ExpiresAt)
}

func (s *Store) HasRefreshToken() bool {
	creds, err := s.Load()
	if err != nil {
		return false
	}
	return creds.RefreshToken != ""
}
