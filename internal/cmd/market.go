package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/client"
)

var marketCmd = &cobra.Command{
	Use:   "market",
	Short: "Market data (public — no auth required)",
}

var marketQuoteCmd = &cobra.Command{
	Use:   "quote <symbol> [symbol2] [symbol3] ...",
	Short: "Get real-time quote for one or more symbols",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)

		if len(args) == 1 {
			quote, err := c.GetQuote(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printResult(cmd, quote)
		}

		// Multiple symbols — use batch endpoint
		resp, err := c.GetQuotes(cmd.Context(), args)
		if err != nil {
			return err
		}
		return printResult(cmd, resp)
	},
}

var marketSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for symbols",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		limit, _ := cmd.Flags().GetInt("limit")
		results, err := c.SearchSymbols(cmd.Context(), args[0], limit)
		if err != nil {
			return err
		}
		return printResult(cmd, results)
	},
}

var marketHistoryCmd = &cobra.Command{
	Use:   "history <symbol>",
	Short: "Get historical OHLCV data",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		interval, _ := cmd.Flags().GetString("interval")
		period, _ := cmd.Flags().GetString("range")

		data, err := c.GetHistory(cmd.Context(), args[0], client.HistoryOptions{
			Interval: interval,
			Range:    period,
		})
		if err != nil {
			return err
		}
		return printResult(cmd, data)
	},
}

func init() {
	marketSearchCmd.Flags().IntP("limit", "n", 10, "Max results")

	marketHistoryCmd.Flags().StringP("interval", "i", "1d", "Interval: 1m, 5m, 15m, 30m, 1h, 1d, 1wk, 1mo")
	marketHistoryCmd.Flags().StringP("range", "r", "3mo", "Range: 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, max")

	marketCmd.AddCommand(marketQuoteCmd)
	marketCmd.AddCommand(marketSearchCmd)
	marketCmd.AddCommand(marketHistoryCmd)
}
