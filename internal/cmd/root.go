package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tradekit-dev/tradekit-cli/internal/auth"
	"github.com/tradekit-dev/tradekit-cli/internal/client"
	"github.com/tradekit-dev/tradekit-cli/internal/config"
	"github.com/tradekit-dev/tradekit-cli/internal/output"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type contextKey string

const (
	clientKey    contextKey = "client"
	formatterKey contextKey = "formatter"
)

var rootCmd = &cobra.Command{
	Use:   "tradekit",
	Short: "TradeKit CLI — trading journal from your terminal",
	Long: `TradeKit CLI is an open-source command-line client for the TradeKit
trading journal platform (tradekit.com.br).

Manage trades, check market data, run backtests, and analyze
your trading performance — all from the terminal.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override from flags
		if v := cmd.Flag("base-url"); v != nil && v.Changed {
			cfg.BaseURL = v.Value.String()
		}

		// Setup auth store
		store := auth.NewStore(config.Dir())

		// Check for API key from flag or env
		if apiKey := viper.GetString("api_key"); apiKey != "" {
			creds, _ := store.Load()
			creds.APIKey = apiKey
			_ = store.Save(creds)
		}

		// Create client
		c := client.New(cfg.BaseURL, store, version)

		// Determine output format
		outputFormat := output.Format(cfg.Output)
		if v := cmd.Flag("output"); v != nil && v.Changed {
			outputFormat = output.Format(v.Value.String())
		}
		formatter := output.New(outputFormat)

		// Store in context
		ctx := cmd.Context()
		ctx = context.WithValue(ctx, clientKey, c)
		ctx = context.WithValue(ctx, formatterKey, formatter)
		cmd.SetContext(ctx)

		return nil
	},
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: table, json, csv")
	rootCmd.PersistentFlags().String("base-url", "", "API base URL (default: https://api.tradekit.com.br)")
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication")

	_ = viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(tradeCmd)
	rootCmd.AddCommand(marketCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func getClient(cmd *cobra.Command) *client.Client {
	return cmd.Context().Value(clientKey).(*client.Client)
}

func getFormatter(cmd *cobra.Command) output.Formatter {
	return cmd.Context().Value(formatterKey).(output.Formatter)
}

func printResult(cmd *cobra.Command, data any) error {
	return getFormatter(cmd).Format(os.Stdout, data)
}
