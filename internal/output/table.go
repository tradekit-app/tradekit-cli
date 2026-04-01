package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
)

type TableFormatter struct {
	Color bool
}

func (f *TableFormatter) Format(w io.Writer, data any) error {
	switch v := data.(type) {
	case []types.Trade:
		return f.formatTrades(w, v)
	case *types.Trade:
		return f.formatTradeDetail(w, v)
	case *types.Quote:
		return f.formatQuote(w, v)
	case []types.SearchResult:
		return f.formatSearchResults(w, v)
	case *types.HistoricalData:
		return f.formatHistory(w, v)
	case *types.User:
		return f.formatUser(w, v)
	case *types.TradeStats:
		return f.formatTradeStats(w, v)
	case []types.APIKey:
		return f.formatAPIKeys(w, v)
	case *types.CreateAPIKeyResponse:
		return f.formatNewAPIKey(w, v)
	case *types.QuotesResponse:
		return f.formatQuotes(w, v)
	case *types.TodayResponse:
		return f.formatToday(w, v)
	case *types.DashboardResponse:
		return f.formatDashboard(w, v)
	case *types.PortfolioResponse:
		return f.formatPortfolio(w, v)
	case map[string]any:
		return f.formatKeyValue(w, v)
	default:
		// Fallback to JSON for unknown types
		jf := &JSONFormatter{Pretty: true}
		return jf.Format(w, data)
	}
}

func (f *TableFormatter) newTable(w io.Writer) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)
	if f.Color {
		t.Style().Color = table.ColorOptionsDefault
	}
	return t
}

func (f *TableFormatter) formatTrades(w io.Writer, trades []types.Trade) error {
	if len(trades) == 0 {
		fmt.Fprintln(w, "No trades found.")
		return nil
	}

	t := f.newTable(w)
	t.AppendHeader(table.Row{"Symbol", "Direction", "Entry Price", "Exit Price", "P&L", "Status", "Date"})

	for _, trade := range trades {
		pnl := "-"
		if trade.NetPnl != nil {
			pnl = *trade.NetPnl
		}
		exitPrice := "-"
		if trade.ExitPrice != nil {
			exitPrice = *trade.ExitPrice
		}
		t.AppendRow(table.Row{
			trade.Symbol,
			strings.ToUpper(trade.Direction),
			trade.EntryPrice,
			exitPrice,
			f.colorPnl(pnl),
			trade.Status,
			trade.EntryDate.Format("2006-01-02"),
		})
	}

	t.Render()
	return nil
}

func (f *TableFormatter) formatTradeDetail(w io.Writer, trade *types.Trade) error {
	t := f.newTable(w)
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight, AlignHeader: text.AlignRight},
	})

	rows := []table.Row{
		{"ID", trade.ID},
		{"Symbol", trade.Symbol},
		{"Asset Type", trade.AssetType},
		{"Direction", strings.ToUpper(trade.Direction)},
		{"Status", trade.Status},
		{"Entry Date", trade.EntryDate.Format("2006-01-02 15:04")},
		{"Entry Price", trade.EntryPrice},
		{"Entry Qty", trade.EntryQuantity},
	}

	if trade.ExitDate != nil {
		rows = append(rows, table.Row{"Exit Date", trade.ExitDate.Format("2006-01-02 15:04")})
	}
	if trade.ExitPrice != nil {
		rows = append(rows, table.Row{"Exit Price", *trade.ExitPrice})
	}
	if trade.StopLoss != nil {
		rows = append(rows, table.Row{"Stop Loss", *trade.StopLoss})
	}
	if trade.TakeProfit != nil {
		rows = append(rows, table.Row{"Take Profit", *trade.TakeProfit})
	}
	if trade.NetPnl != nil {
		rows = append(rows, table.Row{"Net P&L", f.colorPnl(*trade.NetPnl)})
	}
	if trade.PnlPercentage != nil {
		rows = append(rows, table.Row{"P&L %", *trade.PnlPercentage + "%"})
	}
	if trade.Setup != "" {
		rows = append(rows, table.Row{"Setup", trade.Setup})
	}
	if trade.Strategy != "" {
		rows = append(rows, table.Row{"Strategy", trade.Strategy})
	}
	if trade.Notes != "" {
		rows = append(rows, table.Row{"Notes", trade.Notes})
	}

	for _, row := range rows {
		t.AppendRow(row)
	}
	t.Render()
	return nil
}

func (f *TableFormatter) formatQuote(w io.Writer, q *types.Quote) error {
	change := fmt.Sprintf("%s (%s%%)", q.Change, q.ChangePercent)
	if f.Color {
		if strings.HasPrefix(q.Change, "-") {
			change = "\033[31m" + change + "\033[0m"
		} else {
			change = "\033[32m+" + change + "\033[0m"
		}
	}

	t := f.newTable(w)
	t.AppendRow(table.Row{"Symbol", q.Symbol})
	if q.Name != "" && q.Name != q.Symbol {
		t.AppendRow(table.Row{"Name", q.Name})
	}
	t.AppendRow(table.Row{"Price", q.Price + " " + q.Currency})
	t.AppendRow(table.Row{"Change", change})
	if q.Open != "0" && q.Open != "" {
		t.AppendRow(table.Row{"Open", q.Open})
	}
	if q.High != "0" && q.High != "" {
		t.AppendRow(table.Row{"High", q.High})
	}
	if q.Low != "0" && q.Low != "" {
		t.AppendRow(table.Row{"Low", q.Low})
	}
	t.AppendRow(table.Row{"Prev Close", q.PreviousClose})
	if q.Volume > 0 {
		t.AppendRow(table.Row{"Volume", formatNumber(q.Volume)})
	}
	if q.MarketCap > 0 {
		t.AppendRow(table.Row{"Market Cap", formatLargeNumber(q.MarketCap)})
	}
	if q.FiftyTwoWeekHigh != "0" && q.FiftyTwoWeekHigh != "" {
		t.AppendRow(table.Row{"52W High", q.FiftyTwoWeekHigh})
	}
	if q.FiftyTwoWeekLow != "0" && q.FiftyTwoWeekLow != "" {
		t.AppendRow(table.Row{"52W Low", q.FiftyTwoWeekLow})
	}

	t.Render()
	return nil
}

func (f *TableFormatter) formatSearchResults(w io.Writer, results []types.SearchResult) error {
	if len(results) == 0 {
		fmt.Fprintln(w, "No results found.")
		return nil
	}

	t := f.newTable(w)
	t.AppendHeader(table.Row{"Symbol", "Name", "Exchange", "Type"})
	for _, r := range results {
		t.AppendRow(table.Row{r.Symbol, r.Name, r.Exchange, r.Type})
	}
	t.Render()
	return nil
}

func (f *TableFormatter) formatHistory(w io.Writer, h *types.HistoricalData) error {
	if len(h.Data) == 0 {
		fmt.Fprintln(w, "No historical data found.")
		return nil
	}

	fmt.Fprintf(w, "%s (%s)\n\n", h.Symbol, h.Currency)

	t := f.newTable(w)
	t.AppendHeader(table.Row{"Date", "Open", "High", "Low", "Close", "Volume"})

	for _, bar := range h.Data {
		t.AppendRow(table.Row{
			bar.Date.Format("2006-01-02"),
			bar.Open,
			bar.High,
			bar.Low,
			bar.Close,
			formatNumber(bar.Volume),
		})
	}
	t.Render()
	return nil
}

func (f *TableFormatter) formatUser(w io.Writer, u *types.User) error {
	t := f.newTable(w)
	t.AppendRow(table.Row{"Name", u.Name})
	t.AppendRow(table.Row{"Email", u.Email})
	t.AppendRow(table.Row{"Plan", u.SubscriptionPlan})
	t.AppendRow(table.Row{"Email Verified", u.EmailVerified})
	t.AppendRow(table.Row{"2FA Enabled", u.TwoFactorEnabled})
	t.AppendRow(table.Row{"Member Since", u.CreatedAt.Format("2006-01-02")})
	t.Render()
	return nil
}

func (f *TableFormatter) formatTradeStats(w io.Writer, s *types.TradeStats) error {
	t := f.newTable(w)
	t.AppendRow(table.Row{"Total Trades", s.TotalTrades})
	t.AppendRow(table.Row{"Winning", s.WinningTrades})
	t.AppendRow(table.Row{"Losing", s.LosingTrades})
	t.AppendRow(table.Row{"Net P&L", f.colorPnl(fmt.Sprintf("%.2f", s.NetPnl))})
	t.AppendRow(table.Row{"Gross P&L", fmt.Sprintf("%.2f", s.GrossPnl)})
	t.AppendRow(table.Row{"Volume", fmt.Sprintf("%.2f", s.TotalVolume)})
	t.Render()
	return nil
}

func (f *TableFormatter) formatAPIKeys(w io.Writer, keys []types.APIKey) error {
	if len(keys) == 0 {
		fmt.Fprintln(w, "No API keys found.")
		return nil
	}

	t := f.newTable(w)
	t.AppendHeader(table.Row{"ID", "Name", "Prefix", "Scopes", "Last Used", "Created"})
	for _, k := range keys {
		lastUsed := "-"
		if k.LastUsed != nil {
			lastUsed = k.LastUsed.Format("2006-01-02")
		}
		t.AppendRow(table.Row{
			k.ID[:8] + "...",
			k.Name,
			k.Prefix + "...",
			strings.Join(k.Scopes, ", "),
			lastUsed,
			k.CreatedAt.Format("2006-01-02"),
		})
	}
	t.Render()
	return nil
}

func (f *TableFormatter) formatNewAPIKey(w io.Writer, resp *types.CreateAPIKeyResponse) error {
	fmt.Fprintln(w, "API key created successfully!")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Key: %s\n", resp.RawKey)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  Save this key — it won't be shown again.")
	fmt.Fprintf(w, "  To use it: tradekit config set api_key %s\n", resp.RawKey)
	return nil
}

func (f *TableFormatter) formatKeyValue(w io.Writer, m map[string]any) error {
	t := f.newTable(w)
	t.AppendHeader(table.Row{"Key", "Value"})
	for k, v := range m {
		t.AppendRow(table.Row{k, fmt.Sprintf("%v", v)})
	}
	t.Render()
	return nil
}

func (f *TableFormatter) colorPnl(pnl string) string {
	if !f.Color || pnl == "-" || pnl == "" {
		return pnl
	}
	if strings.HasPrefix(pnl, "-") {
		return "\033[31m" + pnl + "\033[0m"
	}
	return "\033[32m" + pnl + "\033[0m"
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func (f *TableFormatter) formatQuotes(w io.Writer, resp *types.QuotesResponse) error {
	if len(resp.Quotes) == 0 {
		fmt.Fprintln(w, "No quotes found.")
		return nil
	}

	t := f.newTable(w)
	t.AppendHeader(table.Row{"Symbol", "Price", "Change", "Change %", "Volume"})

	for _, q := range resp.Quotes {
		change := q.Change
		if f.Color {
			if strings.HasPrefix(q.Change, "-") {
				change = "\033[31m" + change + "\033[0m"
			} else {
				change = "\033[32m+" + change + "\033[0m"
			}
		}
		t.AppendRow(table.Row{
			q.Symbol,
			q.Price + " " + q.Currency,
			change,
			q.ChangePercent + "%",
			formatNumber(q.Volume),
		})
	}
	t.Render()

	if len(resp.Errors) > 0 {
		fmt.Fprintf(w, "\nFailed: ")
		for i, e := range resp.Errors {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprintf(w, "%s (%s)", e.Symbol, e.Message)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func (f *TableFormatter) formatToday(w io.Writer, resp *types.TodayResponse) error {
	fmt.Fprintf(w, "Today: %s\n\n", resp.Date)

	// Stats summary
	t := f.newTable(w)
	t.AppendRow(table.Row{"Trades", resp.TotalTrades})
	t.AppendRow(table.Row{"Winning", resp.Stats.WinningTrades})
	t.AppendRow(table.Row{"Losing", resp.Stats.LosingTrades})
	t.AppendRow(table.Row{"Net P&L", f.colorPnl(fmt.Sprintf("%.2f", resp.Stats.NetPnl))})
	t.Render()

	if len(resp.Trades) > 0 {
		fmt.Fprintln(w)
		return f.formatTrades(w, resp.Trades)
	}
	return nil
}

func (f *TableFormatter) formatDashboard(w io.Writer, d *types.DashboardResponse) error {
	fmt.Fprintln(w, "Dashboard (last 30 days)")
	fmt.Fprintln(w)

	t := f.newTable(w)
	t.AppendRow(table.Row{"Total Trades", d.MonthStats.TotalTrades})
	t.AppendRow(table.Row{"Win Rate", d.MonthStats.WinRate})
	t.AppendRow(table.Row{"Month P&L", f.colorPnl(fmt.Sprintf("%.2f", d.MonthStats.NetPnl))})
	t.AppendRow(table.Row{"Today P&L", f.colorPnl(fmt.Sprintf("%.2f", d.TodayStats.NetPnl))})
	t.AppendRow(table.Row{"Open Positions", d.PositionCount})
	t.AppendRow(table.Row{"Profit Factor", d.MonthStats.ProfitFactor})
	t.Render()

	if len(d.OpenPositions) > 0 {
		fmt.Fprintln(w, "\nOpen Positions")
		pt := f.newTable(w)
		pt.AppendHeader(table.Row{"Symbol", "Direction", "Qty", "Avg Price"})
		for _, p := range d.OpenPositions {
			pt.AppendRow(table.Row{
				p.Symbol,
				strings.ToUpper(p.Direction),
				fmt.Sprintf("%.0f", p.TotalQuantity),
				fmt.Sprintf("%.2f", p.AvgEntryPrice),
			})
		}
		pt.Render()
	}

	if len(d.RecentTrades) > 0 {
		fmt.Fprintln(w, "\nRecent Trades")
		return f.formatTrades(w, d.RecentTrades)
	}
	return nil
}

func (f *TableFormatter) formatPortfolio(w io.Writer, p *types.PortfolioResponse) error {
	if len(p.Positions) == 0 {
		fmt.Fprintln(w, "No open positions.")
		return nil
	}

	t := f.newTable(w)
	t.AppendHeader(table.Row{"Symbol", "Direction", "Qty", "Avg Entry", "Current", "Unreal. P&L"})

	for _, pos := range p.Positions {
		currentPrice := "-"
		pnl := "-"
		if pos.CurrentPrice != "" {
			currentPrice = pos.CurrentPrice
			// Try to calculate unrealized P&L
			if cp, err := strconv.ParseFloat(pos.CurrentPrice, 64); err == nil {
				diff := cp - pos.AvgEntryPrice
				if strings.ToLower(pos.Direction) == "short" {
					diff = -diff
				}
				unrealized := diff * pos.TotalQuantity
				pnl = f.colorPnl(fmt.Sprintf("%.2f", unrealized))
			}
		}
		t.AppendRow(table.Row{
			pos.Symbol,
			strings.ToUpper(pos.Direction),
			fmt.Sprintf("%.0f", pos.TotalQuantity),
			fmt.Sprintf("%.2f", pos.AvgEntryPrice),
			currentPrice,
			pnl,
		})
	}
	t.Render()
	return nil
}

func formatLargeNumber(n int64) string {
	switch {
	case n >= 1_000_000_000_000:
		return fmt.Sprintf("%.1fT", float64(n)/1_000_000_000_000)
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

