package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const defaultConnectionID = "a9d16ae9-45c5-44e1-8e62-1c300fb7aa7f"

var mt5Cmd = &cobra.Command{
	Use:   "mt5",
	Short: "MetaTrader 5 account data",
}

var mt5AccountCmd = &cobra.Command{
	Use:   "account",
	Short: "Show MT5 account balance, equity, and positions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		connID, _ := cmd.Flags().GetString("connection")

		data, err := c.GetMT5Account(cmd.Context(), connID)
		if err != nil {
			return err
		}

		if data.Account == nil {
			fmt.Println("No account data available. Is the EA running?")
			return nil
		}

		acc := data.Account
		fmt.Printf("MT5 Account (%s)\n\n", acc.Currency)
		fmt.Printf("  Balance:     %.2f\n", acc.Balance)
		fmt.Printf("  Equity:      %.2f\n", acc.Equity)
		fmt.Printf("  Margin:      %.2f\n", acc.Margin)
		fmt.Printf("  Free Margin: %.2f\n", acc.FreeMargin)
		fmt.Printf("  Leverage:    1:%d\n", acc.Leverage)

		if len(data.Positions) > 0 {
			fmt.Printf("\n  Open Positions (%d):\n", len(data.Positions))
			for _, p := range data.Positions {
				profitColor := ""
				resetColor := ""
				if p.Profit >= 0 {
					profitColor = "\033[32m"
				} else {
					profitColor = "\033[31m"
				}
				resetColor = "\033[0m"
				fmt.Printf("    #%-8d %-6s %-4s  vol=%.2f  open=%.2f  %s%+.2f%s\n",
					p.Ticket, p.Symbol, p.Type, p.Volume, p.OpenPrice,
					profitColor, p.Profit, resetColor)
			}
		} else {
			fmt.Println("\n  No open positions.")
		}

		return nil
	},
}

func init() {
	mt5AccountCmd.Flags().String("connection", defaultConnectionID, "MT5 connection ID")

	mt5Cmd.AddCommand(mt5AccountCmd)
	rootCmd.AddCommand(mt5Cmd)
}
