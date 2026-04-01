# Contributing to TradeKit CLI

Thanks for your interest in contributing! This CLI is a pure API client — you don't need access to the TradeKit backend to contribute.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/tradekit-cli.git
   cd tradekit-cli
   ```
3. Install Go 1.23+
4. Build:
   ```bash
   make build
   ```
5. Run tests:
   ```bash
   make test
   ```

## Project Structure

```
cmd/tradekit/          Entry point
internal/
  cmd/                 Cobra command definitions
  client/              HTTP client (one file per API domain)
  config/              Config management (~/.tradekit/)
  auth/                Credential storage
  output/              Output formatters (table, JSON, CSV)
pkg/types/             API response types (exported)
```

## Adding a New Command

1. Add API types to `pkg/types/` if needed
2. Add client method to the appropriate file in `internal/client/`
3. Add table formatting in `internal/output/table.go`
4. Create the cobra command in `internal/cmd/`
5. Register the command in the parent's `init()` function

## Code Style

- Run `make lint` before submitting
- Follow existing patterns — commands are noun-verb (`tradekit trade list`)
- Keep client methods simple: one HTTP call per method
- Table output should be concise and scannable

## Pull Requests

- Keep PRs focused on a single feature or fix
- Include a clear description of what changed and why
- Add tests for new client methods when possible

## API Reference

The CLI talks to the TradeKit API gateway. Public endpoints (market data, screening) don't require authentication. Most other endpoints require a JWT token or API key.

API base URL: `https://api.tradekit.com.br`
