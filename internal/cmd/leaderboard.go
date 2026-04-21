package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/client"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

const liveTradesThreshold = 30

var leaderboardCmd = &cobra.Command{
	Use:   "leaderboard",
	Short: "Rank strategies by performance (live when mature, backtest otherwise)",
	Long: `Shows each strategy with the most reliable numbers available:
live stats when there are enough trades (` + strconv.Itoa(liveTradesThreshold) + `+), otherwise
the latest backtest. Matches the web app leaderboard logic.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		ctx := cmd.Context()

		strategies, err := c.ListStrategies(ctx, 100)
		if err != nil {
			return fmt.Errorf("listing strategies: %w", err)
		}

		entries := buildLeaderboard(ctx, c, strategies)

		format, _ := cmd.Flags().GetString("output")
		if format == "json" {
			return printResult(cmd, entries)
		}

		return printLeaderboardTable(entries)
	},
}

func init() {
	rootCmd.AddCommand(leaderboardCmd)
}

func categoryFromTags(tags []string) string {
	set := map[string]bool{}
	for _, t := range tags {
		set[t] = true
	}
	switch {
	case set["pair-trading"]:
		return "pair"
	case set["options"]:
		return "options"
	case set["ml"]:
		return "ml"
	case set["quant"]:
		return "quant"
	default:
		return "rules"
	}
}

func buildLeaderboard(ctx context.Context, c *client.Client, strategies []types.Strategy) []types.LeaderboardEntry {
	type slot struct {
		strategy types.Strategy
		live     *types.StrategyLivePerformance
		bt       *types.BacktestResult
	}

	slots := make([]slot, len(strategies))
	var wg sync.WaitGroup
	for i := range strategies {
		slots[i].strategy = strategies[i]
		wg.Add(1)
		go func(i int, s types.Strategy) {
			defer wg.Done()
			if live, err := c.GetStrategyLivePerformance(ctx, s.ID); err == nil {
				slots[i].live = live
			}
			if bts, err := c.GetBacktests(ctx, s.ID); err == nil && len(bts) > 0 {
				slots[i].bt = &bts[0]
			}
		}(i, strategies[i])
	}
	wg.Wait()

	entries := make([]types.LeaderboardEntry, 0, len(slots))
	for _, sl := range slots {
		e := types.LeaderboardEntry{
			Name:     sl.strategy.Name,
			Category: categoryFromTags(sl.strategy.Tags),
			Source:   "none",
		}

		switch {
		case sl.live != nil && sl.live.TotalTrades >= liveTradesThreshold:
			e.Source = "live"
			e.Trades = sl.live.TotalTrades
			e.WinRate = sl.live.WinRate
			e.TotalPnl = sl.live.TotalPnl
			e.ProfitFactor = sl.live.ProfitFactor
		case sl.bt != nil:
			trades := sl.bt.EntrySignals
			if trades == 0 {
				trades = len(sl.bt.SimulatedTrades)
			}
			if trades > 0 {
				e.Source = "backtest"
				e.Trades = trades
				wr := parseFloat(sl.bt.WinRate)
				if wr > 1 {
					wr /= 100
				}
				e.WinRate = wr
				e.TotalPnl = parseFloat(sl.bt.TotalReturn)
				e.ProfitFactor = parseFloat(sl.bt.ProfitFactor)
			}
		}

		entries = append(entries, e)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if (a.Source == "none") != (b.Source == "none") {
			return a.Source != "none"
		}
		return a.TotalPnl > b.TotalPnl
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries
}

func printLeaderboardTable(entries []types.LeaderboardEntry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tStrategy\tCat\tSrc\tTrd\tWin%\tPnL/Ret\tPF")
	fmt.Fprintln(w, "---\t---\t---\t---\t---\t---\t---\t---")
	for _, e := range entries {
		if e.Source == "none" {
			fmt.Fprintf(w, "%d\t%s\t%s\t-\t-\t-\t-\t-\n",
				e.Rank, trunc(e.Name, 40), e.Category)
			continue
		}
		pf := "-"
		if e.ProfitFactor > 0 {
			pf = fmt.Sprintf("%.2f", e.ProfitFactor)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%.1f%%\t%.2f\t%s\n",
			e.Rank, trunc(e.Name, 40), e.Category, e.Source,
			e.Trades, e.WinRate*100, e.TotalPnl, pf)
	}
	return w.Flush()
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

