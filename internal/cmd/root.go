package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	accountKey   contextKey = "account"
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

		// Resolve --account (name or id) into an ID via /trading-accounts
		accountRef := ""
		if v := cmd.Flag("account"); v != nil && v.Value.String() != "" {
			accountRef = v.Value.String()
		}
		accountID := ""
		if accountRef != "" {
			resolved, err := resolveAccountID(cmd.Context(), c, accountRef)
			if err != nil {
				return err
			}
			accountID = resolved
		}

		// Store in context
		ctx := cmd.Context()
		ctx = context.WithValue(ctx, clientKey, c)
		ctx = context.WithValue(ctx, formatterKey, formatter)
		ctx = context.WithValue(ctx, accountKey, accountID)
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
	rootCmd.PersistentFlags().String("account", "", "Scope to a trading account (name or id)")

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

// getAccountID returns the resolved trading-account ID for the current command,
// or "" if --account was not supplied.
func getAccountID(cmd *cobra.Command) string {
	if v := cmd.Context().Value(accountKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// resolveAccountID accepts an account name or UUID and returns the UUID.
// Exact-id match wins; otherwise does a case-insensitive name match against
// the /trading-accounts list. Returns an error when nothing matches or more
// than one name matches.
func resolveAccountID(ctx context.Context, c *client.Client, ref string) (string, error) {
	// Looks like a UUID — pass through
	if len(ref) == 36 && ref[8] == '-' {
		return ref, nil
	}
	accounts, err := c.ListTradingAccounts(ctx)
	if err != nil {
		return "", fmt.Errorf("listing accounts to resolve --account %q: %w", ref, err)
	}
	var matches []string
	for _, a := range accounts {
		if a.Name == ref || a.ID == ref {
			return a.ID, nil
		}
		if strings.EqualFold(a.Name, ref) {
			matches = append(matches, a.ID)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no trading account matches %q", ref)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("account %q is ambiguous — pass the id instead", ref)
	}
}
