package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var strategyCmd = &cobra.Command{
	Use:   "strategy",
	Short: "Strategy commands",
}

var strategyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your strategies",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		perPage, _ := cmd.Flags().GetInt("per-page")
		strategies, err := c.ListStrategies(cmd.Context(), perPage)
		if err != nil {
			return err
		}
		if len(strategies) == 0 {
			fmt.Println("No strategies found.")
			return nil
		}
		// lifecycle_stage is the operational truth (only paper/live actually run).
		// Status is the legacy enum and is kept in JSON output but not in the
		// human-readable table — too many users mistook "status=active" for
		// "this will trade tomorrow" when it could equally mean "validated but
		// not promoted yet".
		stageGlyph := map[string]string{
			"live":       "▶",
			"paper":      "◐",
			"validated":  "·",
			"backtested": "·",
			"draft":      "·",
			"killed":     "✗",
			"archived":   "·",
		}
		for _, s := range strategies {
			ls := s.LifecycleStage
			if ls == "" {
				ls = "?"
			}
			g := stageGlyph[ls]
			if g == "" {
				g = "·"
			}
			kind := s.Kind
			if kind == "" {
				kind = "—"
			}
			fmt.Printf("%s %s  %-40s  stage=%-10s kind=%-22s symbols=%v\n",
				g, s.ID, truncate(s.Name, 40), ls, kind, s.Symbols)
		}
		fmt.Println()
		fmt.Println("Legend: ▶ live (scanned + emit signals)  ◐ paper (scanned, no signal)  · validated/draft (not scanned)  ✗ killed")
		return nil
	},
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

var strategyRevalidateAllCmd = &cobra.Command{
	Use:   "revalidate-all",
	Short: "Re-run /validate on all active strategies and diff verdicts",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		strategies, err := c.ListStrategies(cmd.Context(), 100)
		if err != nil {
			return err
		}

		// Filter active.
		type row struct {
			ID, Name, Old, New string
			OldRobust          float64
			NewRobust          float64
			Err                string
		}
		var rows []row
		active := 0
		for _, s := range strategies {
			if s.Status != "active" {
				continue
			}
			active++
		}
		fmt.Printf("Re-validating %d active strategies (this may take a while)…\n\n", active)

		i := 0
		for _, s := range strategies {
			if s.Status != "active" {
				continue
			}
			i++
			r := row{ID: s.ID, Name: s.Name}

			old, _ := c.GetValidationSummary(cmd.Context(), s.ID)
			r.Old = strOf(old, "verdict")
			r.OldRobust = numOf(old, "robustnessScore")

			fmt.Printf("[%d/%d] %s — %s … ", i, active, s.ID[:8], truncate(s.Name, 35))
			fresh, err := c.RunValidationPipeline(cmd.Context(), s.ID)
			if err != nil {
				r.Err = err.Error()
				fmt.Printf("ERROR: %s\n", r.Err)
				rows = append(rows, r)
				continue
			}
			r.New = strOf(fresh, "verdict")
			r.NewRobust = numOf(fresh, "robustnessScore")
			arrow := "→"
			marker := "  "
			if r.Old != "" && r.Old != r.New {
				marker = "⚠ "
			}
			fmt.Printf("%s%s %s %s (robust %.0f→%.0f)\n", marker, displayVerdict(r.Old), arrow, displayVerdict(r.New), r.OldRobust, r.NewRobust)
			rows = append(rows, r)
		}

		// Summary
		fmt.Println()
		fmt.Println("=== Verdict diff summary ===")
		shifts := 0
		for _, r := range rows {
			if r.Err != "" {
				continue
			}
			if r.Old != "" && r.Old != r.New {
				shifts++
				fmt.Printf("  %s  %-35s  %s → %s  (robust %.0f → %.0f)\n",
					r.ID[:8], truncate(r.Name, 35), displayVerdict(r.Old), displayVerdict(r.New), r.OldRobust, r.NewRobust)
			}
		}
		if shifts == 0 {
			fmt.Println("  No verdict changes.")
		}
		errs := 0
		for _, r := range rows {
			if r.Err != "" {
				errs++
				fmt.Printf("  %s  %-35s  ERROR: %s\n", r.ID[:8], truncate(r.Name, 35), r.Err)
			}
		}
		fmt.Printf("\n%d strategies validated, %d verdict shifts, %d errors\n", len(rows)-errs, shifts, errs)
		return nil
	},
}

func strOf(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

func numOf(m map[string]any, k string) float64 {
	if m == nil {
		return 0
	}
	if v, ok := m[k].(float64); ok {
		return v
	}
	return 0
}

func displayVerdict(v string) string {
	if v == "" {
		return "(none)"
	}
	return v
}

var strategyRulesCmd = &cobra.Command{
	Use:   "rules <id>",
	Short: "Show a strategy's entry/exit rules",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		raw, err := c.GetStrategyRaw(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Name: %v\n", raw["name"])
		fmt.Printf("Symbols: %v  Direction: %v  Timeframe: %v\n", raw["symbols"], raw["direction"], raw["timeframe"])
		rules, _ := raw["rules"].([]any)
		fmt.Printf("Rules (%d):\n", len(rules))
		for i, r := range rules {
			rm, _ := r.(map[string]any)
			conds, _ := rm["conditions"].([]any)
			fmt.Printf("  [%d] %s (%v) %v %d conds\n", i, rm["name"], rm["type"], rm["logicOp"], len(conds))
			for j, c := range conds {
				cm, _ := c.(map[string]any)
				left, _ := cm["left"].(map[string]any)
				right, _ := cm["right"].(map[string]any)
				fmt.Printf("       %d: left=%v(p=%v,f=%v) %v right=%v(p=%v,v=%v)\n",
					j,
					left["indicator"], left["period"], left["field"],
					cm["comparator"],
					right["indicator"], right["period"], right["value"])
			}
		}
		return nil
	},
}

func init() {
	strategyListCmd.Flags().Int("per-page", 50, "results per page")
	strategyCmd.AddCommand(strategyListCmd)
	strategyCmd.AddCommand(strategyRulesCmd)
	strategyCmd.AddCommand(strategyRevalidateAllCmd)
	rootCmd.AddCommand(strategyCmd)
}
