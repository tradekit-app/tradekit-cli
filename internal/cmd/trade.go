package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/client"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Manage trades",
}

var tradeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trades",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		status, _ := cmd.Flags().GetString("status")
		symbol, _ := cmd.Flags().GetString("symbol")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		account, _ := cmd.Flags().GetString("account")

		resp, err := c.ListTrades(cmd.Context(), client.ListTradesOptions{
			Page:    page,
			PerPage: perPage,
			Status:  status,
			Symbol:  symbol,
			From:    from,
			To:      to,
			Account: account,
		})
		if err != nil {
			return err
		}

		if err := printResult(cmd, resp.Data); err != nil {
			return err
		}

		if resp.Meta != nil && resp.Meta.TotalPages > 1 {
			fmt.Printf("\nPage %d of %d (total: %d)\n", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total)
		}
		return nil
	},
}

var tradeGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get trade details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		trade, err := c.GetTrade(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, trade)
	},
}

var tradePositionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "Show open positions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		positions, err := c.GetPositions(cmd.Context())
		if err != nil {
			return err
		}
		return printResult(cmd, positions)
	},
}

var tradeStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show trade statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		stats, err := c.GetTradeStats(cmd.Context())
		if err != nil {
			return err
		}
		return printResult(cmd, stats)
	},
}

var tradeExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all trades (no pagination)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		status, _ := cmd.Flags().GetString("status")
		symbol, _ := cmd.Flags().GetString("symbol")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		account, _ := cmd.Flags().GetString("account")

		trades, err := c.ExportTrades(cmd.Context(), client.ListTradesOptions{
			Status:  status,
			Symbol:  symbol,
			From:    from,
			To:      to,
			Account: account,
		})
		if err != nil {
			return err
		}
		return printResult(cmd, trades)
	},
}

var tradeTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's trades and stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		resp, err := c.GetTodayTrades(cmd.Context())
		if err != nil {
			return err
		}
		return printResult(cmd, resp)
	},
}

var tradeAddCmd = &cobra.Command{
	Use:   "add <symbol> <long|short> <price> <quantity>",
	Short: "Quick-add a trade",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		accountID, _ := cmd.Flags().GetString("account")
		stopLoss, _ := cmd.Flags().GetString("stop-loss")
		takeProfit, _ := cmd.Flags().GetString("take-profit")
		notes, _ := cmd.Flags().GetString("notes")

		req := types.QuickTradeRequest{
			Symbol:    args[0],
			Direction: args[1],
			Price:     args[2],
			Quantity:  args[3],
		}
		if accountID != "" {
			req.AccountID = &accountID
		}
		if stopLoss != "" {
			req.StopLoss = &stopLoss
		}
		if takeProfit != "" {
			req.TakeProfit = &takeProfit
		}
		if notes != "" {
			req.Notes = &notes
		}

		trade, err := c.QuickCreateTrade(cmd.Context(), req)
		if err != nil {
			return err
		}
		fmt.Printf("Trade created: %s %s %s @ %s\n", trade.Symbol, trade.Direction, trade.EntryQuantity, trade.EntryPrice)
		return nil
	},
}

func init() {
	tradeListCmd.Flags().Int("page", 1, "Page number")
	tradeListCmd.Flags().Int("per-page", 20, "Results per page")
	tradeListCmd.Flags().String("status", "", "Filter by status: open, closed, cancelled")
	tradeListCmd.Flags().StringP("symbol", "s", "", "Filter by symbol")
	tradeListCmd.Flags().String("from", "", "From date (YYYY-MM-DD)")
	tradeListCmd.Flags().String("to", "", "To date (YYYY-MM-DD)")
	tradeListCmd.Flags().StringP("account", "a", "", "Trading account ID")

	tradeExportCmd.Flags().String("status", "", "Filter by status")
	tradeExportCmd.Flags().StringP("symbol", "s", "", "Filter by symbol")
	tradeExportCmd.Flags().String("from", "", "From date (YYYY-MM-DD)")
	tradeExportCmd.Flags().String("to", "", "To date (YYYY-MM-DD)")
	tradeExportCmd.Flags().StringP("account", "a", "", "Trading account ID")

	tradeAddCmd.Flags().StringP("account", "a", "", "Trading account ID (uses default if omitted)")
	tradeAddCmd.Flags().String("stop-loss", "", "Stop loss price")
	tradeAddCmd.Flags().String("take-profit", "", "Take profit price")
	tradeAddCmd.Flags().String("notes", "", "Trade notes")

	tradeCmd.AddCommand(tradeListCmd)
	tradeCmd.AddCommand(tradeGetCmd)
	tradeCmd.AddCommand(tradePositionsCmd)
	tradeCmd.AddCommand(tradeStatsCmd)
	tradeCmd.AddCommand(tradeExportCmd)
	tradeCmd.AddCommand(tradeTodayCmd)
	tradeCmd.AddCommand(tradeAddCmd)
	tradeCmd.AddCommand(tradePortfolioCmd)
}

var tradePortfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Show positions with current market prices",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		portfolio, err := c.GetPortfolio(cmd.Context())
		if err != nil {
			return err
		}
		return printResult(cmd, portfolio)
	},
}
