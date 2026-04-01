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
| `tradekit auth login` | Log in to TradeKit |
| `tradekit auth logout` | Log out |
| `tradekit auth status` | Show current user info |
| `tradekit auth apikey create` | Create an API key |
| `tradekit trade list` | List trades (with filters) |
| `tradekit trade get <id>` | Trade details |
| `tradekit trade positions` | Open positions |
| `tradekit trade stats` | Trading statistics |
| `tradekit market quote <symbol>` | Real-time quote |
| `tradekit market search <query>` | Symbol search |
| `tradekit market history <symbol>` | Historical OHLCV |
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
