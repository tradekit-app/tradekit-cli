package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/client"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Send trading signals to MT5",
}

var signalBuyCmd = &cobra.Command{
	Use:   "buy <symbol> <price>",
	Short: "Send a BUY signal to MT5",
	Long:  "Creates an entry signal that flows through the pipeline to your MT5 EA for execution.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendSignal(cmd, args[0], "entry", "long", args[1])
	},
}

var signalSellCmd = &cobra.Command{
	Use:   "sell <symbol> <price>",
	Short: "Send a SELL signal to MT5",
	Long:  "Creates an entry signal (short) that flows through the pipeline to your MT5 EA for execution.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendSignal(cmd, args[0], "entry", "short", args[1])
	},
}

var signalCloseCmd = &cobra.Command{
	Use:   "close <symbol> [price]",
	Short: "Send a CLOSE/exit signal to MT5",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		price := "0"
		if len(args) > 1 {
			price = args[1]
		}
		return sendSignal(cmd, args[0], "exit", "", price)
	},
}

var signalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent signals (all statuses)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		pendingOnly, _ := cmd.Flags().GetBool("pending")

		if pendingOnly {
			signals, err := c.ListPendingSignals(cmd.Context())
			if err != nil {
				return err
			}
			if len(signals) == 0 {
				fmt.Println("No pending signals.")
				return nil
			}
			printSignals(signals)
			return nil
		}

		status, _ := cmd.Flags().GetString("status")
		symbol, _ := cmd.Flags().GetString("symbol")
		limit, _ := cmd.Flags().GetInt("limit")

		signals, err := c.ListAllSignals(cmd.Context(), client.ListSignalsOptions{
			Status: status,
			Symbol: symbol,
			Limit:  limit,
		})
		if err != nil {
			return err
		}
		if len(signals) == 0 {
			fmt.Println("No signals found.")
			return nil
		}
		printSignals(signals)
		return nil
	},
}

var signalStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Show detailed status of a signal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		signal, err := c.GetSignal(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Signal: %s\n", signal.ID)
		fmt.Printf("  Symbol:      %s\n", signal.Symbol)
		fmt.Printf("  Type:        %s\n", signal.Type)
		fmt.Printf("  Price:       %s\n", signal.Price)
		fmt.Printf("  Status:      %s\n", signal.Status)
		fmt.Printf("  Description: %s\n", signal.Description)
		fmt.Printf("  Created:     %s\n", signal.CreatedAt[:19])

		if signal.ScheduledAt != "" {
			fmt.Printf("  Scheduled:   %s\n", signal.ScheduledAt[:19])
		}
		if signal.ExpiresAt != "" && len(signal.ExpiresAt) >= 19 {
			fmt.Printf("  Expires:     %s\n", signal.ExpiresAt[:19])
		}

		if signal.ExecutionSuccess != nil {
			fmt.Println()
			fmt.Println("  Execution Results:")
			if *signal.ExecutionSuccess {
				fmt.Println("    Status:    SUCCESS")
			} else {
				fmt.Println("    Status:    FAILED")
			}
			if signal.ExecutionTicket != nil {
				fmt.Printf("    Ticket:    %d\n", *signal.ExecutionTicket)
			}
			if signal.ExecutionPrice != "" {
				fmt.Printf("    Fill:      %s\n", signal.ExecutionPrice)
			}
			if signal.ExecutionVolume != "" {
				fmt.Printf("    Volume:    %s\n", signal.ExecutionVolume)
			}
			if signal.ExecutionError != "" {
				fmt.Printf("    Error:     %s\n", signal.ExecutionError)
			}
			if signal.ExecutedAt != "" && len(signal.ExecutedAt) >= 19 {
				fmt.Printf("    Executed:  %s\n", signal.ExecutedAt[:19])
			}
		}

		return nil
	},
}

func printSignals(signals []types.Signal) {
	for _, s := range signals {
		extra := ""
		if s.Status == "scheduled" && s.ScheduledAt != "" && len(s.ScheduledAt) >= 16 {
			extra = "  scheduled: " + s.ScheduledAt[:16]
		}
		if s.ExecutionSuccess != nil {
			if *s.ExecutionSuccess {
				ticket := int64(0)
				if s.ExecutionTicket != nil {
					ticket = *s.ExecutionTicket
				}
				extra = fmt.Sprintf("  executed: ticket=%d price=%s", ticket, s.ExecutionPrice)
			} else {
				extra = "  FAILED: " + s.ExecutionError
			}
		}
		expiresStr := ""
		if len(s.ExpiresAt) >= 16 {
			expiresStr = s.ExpiresAt[:16]
		}
		idStr := s.ID
		if len(idStr) > 8 {
			idStr = idStr[:8] + "..."
		}
		fmt.Printf("  %s  %-8s  %-5s  @ %-8s  [%-12s]  %s%s\n",
			idStr, s.Symbol, s.Type, s.Price, s.Status, expiresStr, extra)
	}
}

func sendSignal(cmd *cobra.Command, symbol, signalType, direction, price string) error {
	c := getClient(cmd)

	symbol = strings.ToUpper(symbol)

	desc := fmt.Sprintf("CLI %s %s @ %s", strings.ToUpper(direction), symbol, price)
	if signalType == "exit" {
		desc = fmt.Sprintf("CLI CLOSE %s", symbol)
	}

	notes, _ := cmd.Flags().GetString("notes")
	if notes != "" {
		desc = notes
	}

	req := types.CreateSignalRequest{
		Symbol:      symbol,
		Type:        signalType,
		Description: desc,
		Price:       price,
	}

	if qty, _ := cmd.Flags().GetString("quantity"); qty != "" {
		req.Quantity = &qty
	}

	// Parse scheduling flags
	scheduledAt, schedErr := parseScheduledAt(cmd)
	if schedErr != nil {
		return fmt.Errorf("invalid schedule: %w", schedErr)
	}
	req.ScheduledAt = scheduledAt

	signal, err := c.CreateSignal(cmd.Context(), req)
	if err != nil {
		return err
	}

	action := "BUY"
	if direction == "short" {
		action = "SELL"
	}
	if signalType == "exit" {
		action = "CLOSE"
	}

	if signal.Status == "scheduled" && signal.ScheduledAt != "" {
		fmt.Printf("Signal scheduled: %s %s @ %s\n", action, symbol, price)
		fmt.Printf("  Scheduled for: %s\n", signal.ScheduledAt[:19])
	} else {
		fmt.Printf("Signal sent: %s %s @ %s\n", action, symbol, price)
	}
	fmt.Printf("  ID: %s\n", signal.ID)
	fmt.Printf("  Status: %s\n", signal.Status)
	fmt.Printf("  Expires: %s\n", signal.ExpiresAt)
	fmt.Println("\nSignal will be delivered to your MT5 EA via the pipeline.")

	return nil
}

func parseScheduledAt(cmd *cobra.Command) (*string, error) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*60*60)
	}

	if at, _ := cmd.Flags().GetString("at"); at != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", at, loc)
		if err != nil {
			return nil, fmt.Errorf("use format: 2006-01-02 15:04")
		}
		s := t.Format(time.RFC3339)
		return &s, nil
	}
	if tomorrow, _ := cmd.Flags().GetBool("tomorrow"); tomorrow {
		t := time.Now().In(loc).AddDate(0, 0, 1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, loc)
		s := t.Format(time.RFC3339)
		return &s, nil
	}
	if delay, _ := cmd.Flags().GetString("delay"); delay != "" {
		d, err := time.ParseDuration(delay)
		if err != nil {
			return nil, fmt.Errorf("use format: 30m, 2h, etc")
		}
		t := time.Now().Add(d)
		s := t.Format(time.RFC3339)
		return &s, nil
	}
	return nil, nil
}

func init() {
	signalBuyCmd.Flags().String("notes", "", "Signal description/notes")
	signalBuyCmd.Flags().StringP("quantity", "q", "", "Number of shares (e.g., 500)")
	signalBuyCmd.Flags().String("at", "", "Schedule for specific time (e.g., '2026-04-02 09:00')")
	signalBuyCmd.Flags().Bool("tomorrow", false, "Schedule for tomorrow at market open (09:00 BRT)")
	signalBuyCmd.Flags().String("delay", "", "Delay signal by duration (e.g., 30m, 2h)")

	signalSellCmd.Flags().String("notes", "", "Signal description/notes")
	signalSellCmd.Flags().StringP("quantity", "q", "", "Number of shares")
	signalSellCmd.Flags().String("at", "", "Schedule for specific time")
	signalSellCmd.Flags().Bool("tomorrow", false, "Schedule for tomorrow at market open")
	signalSellCmd.Flags().String("delay", "", "Delay signal by duration")

	signalCloseCmd.Flags().String("notes", "", "Signal description/notes")

	signalListCmd.Flags().Bool("pending", false, "Show only pending/scheduled signals")
	signalListCmd.Flags().String("status", "", "Filter by status: pending, scheduled, acted_on, expired, dismissed")
	signalListCmd.Flags().StringP("symbol", "s", "", "Filter by symbol")
	signalListCmd.Flags().IntP("limit", "n", 20, "Max results")

	signalCmd.AddCommand(signalBuyCmd)
	signalCmd.AddCommand(signalSellCmd)
	signalCmd.AddCommand(signalCloseCmd)
	signalCmd.AddCommand(signalListCmd)
	signalCmd.AddCommand(signalStatusCmd)

	rootCmd.AddCommand(signalCmd)
}
