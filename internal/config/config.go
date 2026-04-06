package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultBaseURL = "https://gateway-202557329647.southamerica-east1.run.app"
	DefaultOutput  = "table"
)

type Config struct {
	BaseURL        string `mapstructure:"base_url"`
	Output         string `mapstructure:"output"`
	DefaultAccount string `mapstructure:"default_account"`
	Color          bool   `mapstructure:"color"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tradekit")
}

func FilePath() string {
	return filepath.Join(Dir(), "config.yaml")
}

func Load() (*Config, error) {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	viper.SetDefault("base_url", DefaultBaseURL)
	viper.SetDefault("output", DefaultOutput)
	viper.SetDefault("default_account", "")
	viper.SetDefault("color", true)

	viper.SetEnvPrefix("TRADEKIT")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Set(key, value string) error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	viper.Set(key, value)
	return viper.WriteConfigAs(FilePath())
}

func Get(key string) string {
	return viper.GetString(key)
}

func AllSettings() map[string]any {
	return viper.AllSettings()
}
