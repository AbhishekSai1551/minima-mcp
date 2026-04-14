# Minima MCP Server

A containerized Go MCP (Model Context Protocol) server for Minima blockchain integration, built with Secure SDLC principles including enforced API access boundaries, input validation, rate-limiting, and SOC 2-aligned structured audit logging.

## Architecture

```
┌──────────────────┐     ┌───────────────────┐     ┌──────────────────┐
│   MCP Client     │────▶│   Minima MCP      │────▶│   Minima Node    │
│  (Claude/Cursor) │     │   Server :3001    │     │   :9004/:9005    │
└──────────────────┘     │                   │     └──────────────────┘
                         │ ┌───────────────┐ │
                         │ │ Rate Limiter  │ │     ┌──────────────────┐
                         │ ├───────────────┤ │────▶│  Tenzro Network  │
                         │ │ Auth Layer    │ │     │  (via tenzro-cli)│
                         │ ├───────────────┤ │     └──────────────────┘
                         │ │ Input Valid.  │ │
                         │ ├───────────────┤ │     ┌──────────────────┐
                         │ │ Audit Logger  │ │────▶│  Audit Logs     │
                         │ └───────────────┘ │     │  (SOC 2 trail)  │
                         └───────────────────┘     └──────────────────┘
```

## MCP Tools (25 tools across 6 categories)

### Public Node Tools
| Tool | Description |
|------|-------------|
| `minima_get_status` | Node status, sync state, block height, peers |
| `minima_get_block` | Block information by height |
| `minima_get_transaction` | Transaction details by ID |
| `minima_get_network_info` | Network peers and connection status |

### Tools for wallets
| Tool | Auth | Description |
|------|------|-------------|
| `minima_get_balance` | Public | Wallet balance (confirmed/unconfirmed) |
| `minima_send_transaction` | Auth | Send tokens to an address |
| `minima_get_transactions` | Auth | Transaction history with pagination |
| `minima_get_coins` | Auth | List all coins/tokens in wallet |

### Tools for smart contracts
| Tool | Auth | Description |
|------|------|-------------|
| `minima_deploy_contract` | Auth | Deploy a smart contract script |
| `minima_execute_contract` | Auth | Execute a contract function |
| `minima_list_contracts` | Auth | List all deployed contracts |
| `minima_get_contract` | Auth | Query contract state |

### Tokens tools
| Tool | Auth | Description |
|------|------|-------------|
| `minima_create_token` | Auth | Create a custom token |
| `minima_list_tokens` | Public | List all tokens |
| `minima_transfer_token` | Auth | Transfer a token |
| `minima_get_token_info` | Public | Token details by ID |

### Keys & Vault tools
| Tool | Auth | Description |
|------|------|-------------|
| `minima_generate_key` | Auth | Generate a new key pair |
| `minima_list_keys` | Auth | List vault public keys |
| `minima_sign_message` | Auth | Sign a message with a private key |
| `minima_verify_signature` | Public | Verify a cryptographic signature |
| `minima_get_vault` | Auth | Vault information |

### MiniDAPPs tools
| Tool | Auth | Description |
|------|------|-------------|
| `minima_list_minidapps` | Auth | List installed MiniDAPPs |
| `minima_install_minidapp` | Auth | Install a MiniDAPP |
| `minima_uninstall_minidapp` | Auth | Uninstall a MiniDAPP |
| `minima_get_minidapp_info` | Public | MiniDAPP details |
| `minima_run_minidapp` | Auth | Execute a MiniDAPP command |

## Quick Start

### Build & Run

```bash
# Build
go build -o bin/minima-mcp ./cmd/minima-mcp

# Run with defaults
./bin/minima-mcp

# Run with custom Minima node
MINIMA_DATA_HOST=192.168.1.100 MINIMA_DATA_PORT=9004 ./bin/minima-mcp
```

### Docker

```bash
# Full stack (Minima node + MCP server)
docker compose up -d

# Build only the MCP server
docker build -t minima-mcp .
docker run -p 3001:3001 \
  -e MINIMA_DATA_HOST=host.docker.internal \
  -e MINIMA_DATA_PORT=9004 \
  minima-mcp
```

### Tenzro CLI Integration

Use `tenzro-cli` for onboarding and identity management:

```bash
# Join Tenzro network (provisions DID + wallet + onboarding key)
tenzro-cli join --name "minima-mcp-operator" --rpc https://rpc.tenzro.network

# Register this MCP server as a Tenzro tool
tenzro-cli tool register \
  --name "minima-mcp" \
  --description "Minima blockchain MCP server" \
  --endpoint /mcp \
  --tool-type mcp \
  --category blockchain \
  --version 1.0.0 \
  --creator-did did:tenzro:human:your-uuid

# Bridge tokens between Tenzro and Minima
tenzro-cli bridge quote --from-chain tenzro --to-chain minima --token TNZO --amount 100
```

## Connect MCP Clients

### Claude Desktop

```json
{
  "mcpServers": {
    "minima": {
      "type": "http",
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:3001/mcp"]
    }
  }
}
```

### Claude Code

```bash
claude mcp add minima http://localhost:3001/mcp
```

## Security

### API Access Boundaries (Tiered Auth)
- **Public tools** — no auth required (`minima_get_status`, `minima_get_balance`, etc.)
- **Authenticated tools** — require `Authorization: Bearer <token>` header
- **Auth modes**: `tiered` (default), `false` (all public), `full` (all require auth)

### Input Validation
- Address format validation (0x/Mx prefixes, hex format, length checks)
- Amount validation (numeric, non-negative, max 1e18)
- Script injection prevention (dangerous patterns blocked)
- Contract ID, function name, token name format enforcement
- Private/public key length validation

### Rate Limiting
- Token bucket algorithm per-client IP
- Configurable requests/second and burst size
- Automatic cleanup of stale buckets
- `Retry-After` header on 429 responses

### Structured Audit Logging (SOC 2)
- JSON-structured events with timestamps, actor, action, parameters
- Automatic field redaction (private keys, passwords, secrets → `[REDACTED]`)
- Daily log rotation with configurable retention
- Event types: `tool_call`, `tool_result`, `auth`, `rate_limit`, `error`, `system`
- Immutable append-only log files
- 90-day default retention (configurable)

## Configuration

See `.env.example` for all configuration options.

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `3001` | MCP server listen port |
| `MINIMA_DATA_HOST` | `localhost` | Minima data endpoint host |
| `MINIMA_DATA_PORT` | `9004` | Minima data endpoint port |
| `MINIMA_MINIDAPP_PORT` | `9005` | MiniDAPP endpoint port |
| `RATE_LIMIT_RPS` | `10` | Requests per second per IP |
| `RATE_LIMIT_BURST` | `20` | Burst capacity |
| `AUTH_MODE` | `tiered` | Auth mode: tiered/false/full |
| `AUDIT_ENABLED` | `true` | Enable structured audit logging |
| `AUDIT_MAX_AGE_DAYS` | `90` | Log retention days |

## Testing

```bash
go test ./... -v
```

## Project Structure

```
minima-mcp/
├── cmd/
│   ├── minima-mcp/          # Server entry point
│   └── tenzro-cli/          # Tenzro CLI integration wrapper
├── internal/
│   ├── audit/               # Structured audit logging (SOC 2)
│   ├── config/              # Configuration management
│   ├── mcp/                 # MCP server + tool definitions
│   ├── minima/              # Minima blockchain client
│   └── ratelimit/           # Token bucket rate limiter
├── Dockerfile
├── docker-compose.yml
└── .env.example
```