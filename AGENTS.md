# AGENTS.md

## Build & Run Commands
- Build: `go build -o bin/minima-mcp ./cmd/minima-mcp`
- Run: `./bin/minima-mcp` or `go run ./cmd/minima-mcp`
- Docker: `docker compose up -d`

## Test & Lint Commands
- Test all: `go test ./... -v`
- Test single package: `go test ./internal/minima/... -v`
- Vet: `go vet ./...`
- Build check: `go build ./...`

## Architecture
- MCP server at `/mcp` endpoint (JSON-RPC 2.0 over HTTP)
- Minima node connection via JSON-RPC to `localhost:9004` (data) and `localhost:9005` (MiniDAPPs)
- 25 tools across 6 categories: Node, Wallet, Contracts, Tokens, Keys, MiniDAPPs
- Tiered auth: public tools (no auth), authenticated tools (require Bearer token)
- Rate limiting: token bucket per IP
- Audit logging: JSON-structured, daily rotation, 90-day retention, field redaction

## Code Style
- Go standard formatting (`gofmt`)
- No comments unless explicitly requested
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Package-by-feature layout under `internal/`

## Key Configuration
- `AUTH_MODE`: tiered (default), false, full
- `MINIMA_DATA_HOST`/`MINIMA_DATA_PORT`: Minima node data endpoint
- `AUDIT_ENABLED`: enable SOC 2 audit trail