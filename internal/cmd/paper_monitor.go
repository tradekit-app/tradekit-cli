package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/client"
)

// Paper-monitor lifecycle commands (services/strategy PR-7 endpoints).
// All bind under `tradekit strategy <verb>` so the CLI tree stays small.
// Output defaults to a one-line digest; `-o json` dumps the full
// response for piping to jq.

var strategyPromotePaperCmd = &cobra.Command{
	Use:   "promote-paper <strategy-id>",
	Short: "Move a validated strategy into paper-monitoring",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		body := &client.PromoteToPaperBody{}
		if v, _ := cmd.Flags().GetString("trading-account"); v != "" {
			body.TradingAccountID = v
		}
		if v, _ := cmd.Flags().GetFloat64("initial-equity"); v > 0 {
			body.InitialEquity = &v
		}
		if path, _ := cmd.Flags().GetString("criteria-file"); path != "" {
			// Caller-supplied JSON wraps both criteria sets:
			//   { "activationCriteria": {...}, "killCriteria": {...} }
			data, err := readFile(path)
			if err != nil {
				return fmt.Errorf("reading criteria file: %w", err)
			}
			var wrapper struct {
				Activation json.RawMessage `json:"activationCriteria"`
				Kill       json.RawMessage `json:"killCriteria"`
			}
			if err := json.Unmarshal(data, &wrapper); err != nil {
				return fmt.Errorf("parsing criteria file: %w", err)
			}
			body.ActivationCriteria = wrapper.Activation
			body.KillCriteria = wrapper.Kill
		}
		resp, err := c.PromoteToPaper(cmd.Context(), args[0], body)
		if err != nil {
			return err
		}
		return printResponse(cmd, resp, fmt.Sprintf("Promoted to paper: %s (lifecycle_stage=%v, equity=%v)",
			args[0], resp["lifecycleStage"], resp["state"]))
	},
}

var strategyPromoteLiveCmd = &cobra.Command{
	Use:   "promote-live <strategy-id>",
	Short: "Promote a paper-stage strategy to live (real signal flow)",
	Long: `Caller is expected to have verified activation criteria first
(via tradekit strategy ticks <id> + manual evaluation OR via the day-30
review /schedule routine). This command only enforces the lifecycle
invariant — it does NOT re-evaluate criteria server-side.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		resp, err := c.PromoteToLive(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printResponse(cmd, resp, fmt.Sprintf("Promoted to LIVE: %s — signals will route to MT5 starting next scan",
			args[0]))
	},
}

var strategyKillCmd = &cobra.Command{
	Use:   "kill <strategy-id> --reason \"...\"",
	Short: "Terminate a paper monitor (paper → killed)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reason, _ := cmd.Flags().GetString("reason")
		if reason == "" {
			return fmt.Errorf("--reason is required: provide a one-line audit string")
		}
		c := getClient(cmd)
		resp, err := c.KillPaper(cmd.Context(), args[0], reason)
		if err != nil {
			return err
		}
		return printResponse(cmd, resp, fmt.Sprintf("KILLED: %s — reason=%q", args[0], reason))
	},
}

var strategyTickCmd = &cobra.Command{
	Use:   "tick <strategy-id>",
	Short: "Run today's tick on a paper or live strategy",
	Long: `Fetches today's bars, computes the strategy's per-tick metrics
(z-score for pair_trade, etc.), persists a strategy_ticks row, and
when the strategy is in lifecycle_stage='live' AND the action is
ENTER/EXIT/STOP_LOSS, also emits a strategy_signals row that routes
to MT5.

For paper strategies, no signal is emitted — only the tick log row.

--asOfDate is reserved for v2 PIT replay; v1 rejects values != today.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asOfDate, _ := cmd.Flags().GetString("asOfDate")
		c := getClient(cmd)
		resp, err := c.RunTick(cmd.Context(), args[0], asOfDate)
		if err != nil {
			return err
		}
		summary := fmt.Sprintf("Tick: action=%v paperPnL=%v",
			resp["action"], resp["paperPnL"])
		return printResponse(cmd, resp, summary)
	},
}

var strategyTicksCmd = &cobra.Command{
	Use:   "ticks <strategy-id>",
	Short: "List recent ticks (newest first)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		c := getClient(cmd)
		ticks, err := c.ListTicks(cmd.Context(), args[0], limit)
		if err != nil {
			return err
		}
		if outFmt, _ := cmd.Flags().GetString("output"); outFmt == "json" {
			b, _ := json.MarshalIndent(ticks, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(ticks) == 0 {
			fmt.Println("No ticks recorded yet.")
			return nil
		}
		fmt.Printf("%-12s %-20s %-10s %s\n", "DATE", "ACTION", "PAPER_PNL", "Z_SCORE")
		for _, t := range ticks {
			z := ""
			if inputs, ok := t["inputs"].(map[string]any); ok {
				if zv, ok := inputs["zScore"]; ok {
					z = fmt.Sprintf("%v", zv)
				}
			}
			fmt.Printf("%-12s %-20v %-10v %s\n",
				truncate(fmt.Sprintf("%v", t["tickDate"]), 10),
				t["action"], t["paperPnL"], z)
		}
		return nil
	},
}

var strategyStateCmd = &cobra.Command{
	Use:   "state <strategy-id>",
	Short: "Show the current in-flight state (position + equity)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		resp, err := c.GetState(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		summary := "FLAT"
		if pos, ok := resp["position"].(map[string]any); ok && pos != nil {
			summary = fmt.Sprintf("%v from %v (entryZ=%v)",
				pos["direction"], pos["entryDate"], pos["entryZ"])
		}
		return printResponse(cmd, resp,
			fmt.Sprintf("State: %s — equity=%v lastTickDate=%v",
				summary, resp["equity"], resp["lastTickDate"]))
	},
}

// printResponse honours the --output flag. JSON dumps verbatim; default
// prints the human-readable summary line.
func printResponse(cmd *cobra.Command, resp map[string]any, humanLine string) error {
	if outFmt, _ := cmd.Flags().GetString("output"); outFmt == "json" {
		b, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Println(humanLine)
	return nil
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func init() {
	strategyPromotePaperCmd.Flags().String("trading-account", "",
		"trading account UUID; defaults to strategy's tradingAccountId, then 100000 BRL fallback")
	strategyPromotePaperCmd.Flags().Float64("initial-equity", 0,
		"override paper equity (BRL). When 0, equity is read from trading_accounts.initial_balance")
	strategyPromotePaperCmd.Flags().String("criteria-file", "",
		"path to JSON file with {activationCriteria, killCriteria} (kind-specific shape)")

	strategyKillCmd.Flags().String("reason", "",
		"one-line audit string explaining the kill (REQUIRED)")
	_ = strategyKillCmd.MarkFlagRequired("reason")

	strategyTickCmd.Flags().String("asOfDate", "",
		"YYYY-MM-DD; v1 only honors today (PIT replay deferred to v2)")

	strategyTicksCmd.Flags().Int("limit", 30, "max number of ticks to fetch")

	strategyCmd.AddCommand(strategyPromotePaperCmd)
	strategyCmd.AddCommand(strategyPromoteLiveCmd)
	strategyCmd.AddCommand(strategyKillCmd)
	strategyCmd.AddCommand(strategyTickCmd)
	strategyCmd.AddCommand(strategyTicksCmd)
	strategyCmd.AddCommand(strategyStateCmd)
}
