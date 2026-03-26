---
last_verified: 2026-03-26
---

# CLI Reference: serve

> Start an HTTP server for Slack webhook event processing (Clew initiative).

`ari serve` runs the Clew reasoning service — an HTTP server that processes Slack webhook events using a knowledge retrieval and Claude generation pipeline.

**Family**: serve
**Commands**: 1
**Priority**: HIGH

---

## Commands

### ari serve

Start the Clew HTTP webhook server.

**Synopsis**:
```bash
ari serve [flags]
```

**Description**:
Starts an HTTP server for Slack webhook event processing. The server provides:

- Slack webhook signature verification (HMAC-SHA256)
- Reasoning pipeline: knowledge retrieval → trust scoring → Claude generation
- Health endpoints: `/health` (liveness), `/ready` (readiness)
- Graceful shutdown on SIGTERM/SIGINT
- Request ID propagation and structured logging
- OpenTelemetry tracing (optional, via `OTEL_EXPORTER_OTLP_ENDPOINT`)
- Concurrency limiting for pipeline queries

**Configuration hierarchy** (highest priority wins):
1. CLI flags (`--port`, `--slack-signing-secret`, etc.)
2. Process environment variables (`SLACK_SIGNING_SECRET`, `PORT`, etc.)
3. Org env file (`$XDG_DATA_HOME/knossos/orgs/{org}/serve.env`)
4. Hardcoded defaults (port=8080, log_level=INFO, max_concurrent=10)

**Required secrets** (must be provided via any tier):

| Secret | Environment Variable | Description |
|--------|---------------------|-------------|
| Slack signing secret | `SLACK_SIGNING_SECRET` | Slack app signing secret for webhook verification |
| Slack bot token | `SLACK_BOT_TOKEN` | Slack bot OAuth token |
| Anthropic API key | `ANTHROPIC_API_KEY` | Claude API key for reasoning pipeline |

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--org` | string | active org | Organization name (env: `KNOSSOS_ORG`) |
| `--port` | int | 8080 | Server port (env: `PORT`) |
| `--slack-signing-secret` | string | - | Slack signing secret (env: `SLACK_SIGNING_SECRET`) |
| `--slack-bot-token` | string | - | Slack bot token (env: `SLACK_BOT_TOKEN`) |
| `--env-file` | string | org default | Path to env file |
| `--max-concurrent` | int | 10 | Max concurrent pipeline queries (env: `MAX_CONCURRENT`) |
| `--drain-timeout` | duration | 30s | Graceful shutdown drain timeout |

**Optional environment variables**:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `LOG_LEVEL` | INFO | Log level: DEBUG, INFO, WARN, ERROR |
| `MAX_CONCURRENT` | 10 | Max concurrent pipeline queries |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (empty) | OTLP collector endpoint; empty disables tracing |

**Examples**:
```bash
# Start server for the autom8y org (reads serve.env from org directory)
ari serve --org autom8y

# Override the default port
ari serve --org autom8y --port 3000

# Load from a custom env file
ari serve --env-file /path/to/custom.env

# Provide secrets directly via environment
SLACK_SIGNING_SECRET=xxx SLACK_BOT_TOKEN=xoxb-xxx ari serve
```

**Health endpoints**:
```bash
# Liveness probe
curl http://localhost:8080/health

# Readiness probe
curl http://localhost:8080/ready
```

---

## Configuration File

The org env file at `$XDG_DATA_HOME/knossos/orgs/{org}/serve.env` stores per-org configuration:

```env
SLACK_SIGNING_SECRET=your_signing_secret
SLACK_BOT_TOKEN=xoxb-your-bot-token
ANTHROPIC_API_KEY=sk-ant-your-key
PORT=8080
LOG_LEVEL=INFO
MAX_CONCURRENT=10
```

Use `ari org init <org>` to scaffold the org directory before configuring this file.

---

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--channel` | string | `all` | Target channel: claude, gemini, or all |
| `--config` | string | `$XDG_CONFIG_HOME/knossos/config.yaml` | Config file path |
| `-o, --output` | string | `text` | Output format: text, json, yaml |
| `-p, --project-dir` | string | auto-discovered | Project root directory |
| `-v, --verbose` | bool | false | Enable verbose output (JSON lines to stderr) |

---

## See Also

- [`ari org init`](cli-org.md#ari-org-init) — Bootstrap an org directory
- [`ari registry sync`](cli-registry.md#ari-registry-sync) — Sync knowledge domain catalog
- [Architecture Map: serve subsystem](../../reference/architecture-map.md) — Package and data flow details
