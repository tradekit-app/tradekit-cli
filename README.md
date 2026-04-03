# TradeKit CLI

The official command-line client for [TradeKit](https://tradekit.com.br) — a trading journal platform for Brazilian traders.

Manage trades, check market data, run backtests, and analyze your trading performance from the terminal.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install tradekit-dev/tap/tradekit
```

### Binary Download

Download the latest release from the [releases page](https://github.com/tradekit-dev/tradekit-cli/releases).

### From Source

```bash
go install github.com/tradekit-dev/tradekit-cli/cmd/tradekit@latest
```

## Quick Start

```bash
# Check a stock quote (no login required)
tradekit market quote PETR4

# Search for symbols
tradekit market search "petrobras"

# Log in to your TradeKit account
tradekit auth login

# View your trades
tradekit trade list

# Check open positions
tradekit trade positions

# View trade statistics
tradekit trade stats
```

## Commands

| Command | Description |
|---------|-------------|
| **Auth** | |
| `tradekit auth login` | Log in to TradeKit |
| `tradekit auth logout` | Log out |
| `tradekit auth status` | Show current user info |
| `tradekit auth apikey create` | Create an API key |
| **Trading** | |
| `tradekit trade list` | List trades (with filters) |
| `tradekit trade get <id>` | Trade details |
| `tradekit trade positions` | Open positions |
| `tradekit trade portfolio` | Positions with live market prices |
| `tradekit trade export` | Export all trades (no pagination) |
| `tradekit trade today` | Today's trades and stats |
| `tradekit trade add <symbol> <dir> <price> <qty>` | Quick-add a trade |
| `tradekit dashboard` | Full trading dashboard summary |
| **Signals (MT5)** | |
| `tradekit signal buy <symbol> <price>` | Send buy signal to MT5 |
| `tradekit signal sell <symbol> <price>` | Send sell signal to MT5 |
| `tradekit signal close <symbol>` | Send close/exit signal |
| `tradekit signal buy VALE3 83 --tomorrow` | Schedule for market open |
| `tradekit signal buy PETR4 47 -q 500` | Buy with specific quantity |
| `tradekit signal list` | List all signals (all statuses) |
| `tradekit signal status <id>` | Detailed execution results |
| **Risk Rules** | |
| `tradekit rules set max-daily-loss 2.0` | Set risk guardrail |
| `tradekit rules set max-exposure 60` | Limit total exposure |
| `tradekit rules set trading-hours 09:00 17:30` | Trading window |
| `tradekit rules list` | View active rules |
| `tradekit rules violations` | View blocked signals |
| **Market Data** | |
| `tradekit market quote <symbol> [symbol2...]` | Real-time quote(s) |
| `tradekit market search <query>` | Symbol search |
| `tradekit market history <symbol>` | Historical OHLCV |
| **MT5** | |
| `tradekit mt5 account` | MT5 balance, equity, positions |
| **Config** | |
| `tradekit config set <key> <value>` | Set config value |
| `tradekit version` | Version info |

## Output Formats

```bash
# Default: formatted table
tradekit trade list

# JSON (for scripting)
tradekit trade list -o json

# CSV (for spreadsheets)
tradekit trade list -o csv

# Pipe to jq
tradekit trade list -o json | jq '.[].symbol'
```

## Configuration

Config is stored in `~/.tradekit/config.yaml`:

```yaml
base_url: "https://api.tradekit.com.br"
output: "table"
default_account: ""
color: true
```

Set values with:
```bash
tradekit config set output json
tradekit config set default_account <uuid>
```

### Environment Variables

All config can be set via environment variables with the `TRADEKIT_` prefix:

```bash
export TRADEKIT_API_KEY=tk_live_...
export TRADEKIT_BASE_URL=https://api.tradekit.com.br
```

## Authentication

### Interactive Login

```bash
tradekit auth login
```

Tokens are stored securely in `~/.tradekit/credentials` (file mode 0600). Access tokens auto-refresh when expired.

### API Key (non-interactive)

```bash
# Create a key via the CLI
tradekit auth apikey create --name "my-script" --scopes read,write

# Or set one directly
tradekit config set api_key tk_live_...

# Or via environment
export TRADEKIT_API_KEY=tk_live_...
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
