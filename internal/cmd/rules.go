package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

const defaultAccountID = "8942bfb6-6b18-402d-96cc-9738db35cf0f"

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage risk management guardrails",
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active risk rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		accountID, _ := cmd.Flags().GetString("account")

		rules, err := c.ListRiskRules(cmd.Context(), accountID)
		if err != nil {
			return err
		}

		if len(rules) == 0 {
			fmt.Println("No risk rules configured.")
			fmt.Println("Set one: tradekit rules set max-daily-loss 3.0")
			return nil
		}

		for _, r := range rules {
			status := "ON"
			if !r.Enabled {
				status = "OFF"
			}
			fmt.Printf("  [%s] %-25s %s  params=%v\n", status, r.RuleType, r.Scope, r.Params)
		}
		return nil
	},
}

var rulesSetCmd = &cobra.Command{
	Use:   "set <rule-type> <value>",
	Short: "Set a risk rule (max-daily-loss, max-exposure, max-per-symbol, trading-hours, max-daily-trades)",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		accountID, _ := cmd.Flags().GetString("account")

		ruleType := args[0]
		params := map[string]interface{}{}

		switch ruleType {
		case "max-daily-loss":
			params["maxPct"] = parseFloat(args[1])
			ruleType = "max_daily_loss"
		case "max-exposure":
			params["maxPct"] = parseFloat(args[1])
			ruleType = "max_exposure"
		case "max-per-symbol":
			params["maxPct"] = parseFloat(args[1])
			ruleType = "max_per_symbol"
		case "max-daily-trades":
			params["maxTrades"] = parseInt(args[1])
			ruleType = "max_daily_trades"
		case "trading-hours":
			if len(args) < 3 {
				return fmt.Errorf("usage: tradekit rules set trading-hours 09:00 17:30")
			}
			params["start"] = args[1]
			params["end"] = args[2]
			params["timezone"] = "America/Sao_Paulo"
			ruleType = "trading_hours"
		case "max-concurrent":
			params["maxPositions"] = parseInt(args[1])
			ruleType = "max_concurrent_positions"
		case "min-rr":
			params["minRatio"] = parseFloat(args[1])
			ruleType = "min_rr_ratio"
		default:
			return fmt.Errorf("unknown rule: %s\nAvailable: max-daily-loss, max-exposure, max-per-symbol, max-daily-trades, trading-hours, max-concurrent, min-rr", ruleType)
		}

		rule, err := c.CreateRiskRule(cmd.Context(), types.CreateRiskRuleRequest{
			TradingAccountID: accountID,
			Scope:            "account",
			RuleType:         ruleType,
			Params:           params,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Rule created: %s\n", rule.RuleType)
		fmt.Printf("  ID: %s\n", rule.ID)
		fmt.Printf("  Params: %v\n", rule.Params)
		fmt.Printf("  Enabled: %v\n", rule.Enabled)
		return nil
	},
}

var rulesDeleteCmd = &cobra.Command{
	Use:   "delete <rule-id>",
	Short: "Delete a risk rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		if err := c.DeleteRiskRule(cmd.Context(), args[0]); err != nil {
			return err
		}
		fmt.Println("Rule deleted.")
		return nil
	},
}

var rulesViolationsCmd = &cobra.Command{
	Use:   "violations",
	Short: "Show recent rule violations (blocked signals)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		violations, err := c.ListRiskViolations(cmd.Context())
		if err != nil {
			return err
		}

		if len(violations) == 0 {
			fmt.Println("No violations recorded.")
			return nil
		}

		for _, v := range violations {
			ts := ""
			if len(v.CreatedAt) >= 16 {
				ts = v.CreatedAt[:16]
			}
			fmt.Printf("  %s  [%s]  %s  %s\n", ts, v.RuleType, v.Symbol, v.Description)
		}
		return nil
	},
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func init() {
	rulesListCmd.Flags().String("account", defaultAccountID, "Trading account ID")
	rulesSetCmd.Flags().String("account", defaultAccountID, "Trading account ID")

	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesSetCmd)
	rulesCmd.AddCommand(rulesDeleteCmd)
	rulesCmd.AddCommand(rulesViolationsCmd)

	rootCmd.AddCommand(rulesCmd)
}
