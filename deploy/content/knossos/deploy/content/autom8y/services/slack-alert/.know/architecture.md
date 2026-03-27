---
domain: architecture
generated_at: "2026-03-16T20:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.97
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The `slack-alert` service is a minimal Python Lambda. It contains exactly one importable package with one module file.

**Directory layout:**

```
services/slack-alert/
  src/
    slack_alert/
      __init__.py          -- package marker + docstring
      handler.py           -- all application logic (91 lines)
  tests/
    __init__.py
    test_handler.py        -- smoke test (9 lines)
  pyproject.toml           -- project manifest (hatchling build, uv deps)
  Dockerfile               -- multi-stage Lambda container image
  Justfile                 -- CI recipes (thin router into just/ modules)
  just/
    _globals.just          -- service-level constants (APP, VERSION, IMAGE)
    ci.just                -- ci, ci-setup, ci-lint, ci-test recipes
```

**Module inventory:**

| Module | File | Purpose | Exported names |
|--------|------|---------|----------------|
| `slack_alert` | `src/slack_alert/__init__.py` | Package marker | none (docstring only) |
| `slack_alert.handler` | `src/slack_alert/handler.py` | Lambda entry point + SNS processing | `lambda_handler`, `_process_event` |

There are no hub modules and no leaf modules — the service is a single-module microservice. `handler.py` is the entire application surface.

## Layer Boundaries

This service has no internal layering. It is a single-file Lambda that calls external SDKs directly.

**Dependency graph (imports in `handler.py`):**

```
handler.py
  -> autom8y_slack          (external SDK: SlackClient, SlackConfig, format_cloudwatch_alarm)
  -> autom8y_log            (external SDK: configure_logging, get_logger)
  -> autom8y_telemetry.aws  (external SDK: instrument_lambda decorator)
  -> stdlib: asyncio, json, os
```

**Layer model:**

| Layer | Contents | Notes |
|-------|----------|-------|
| Entry surface | `lambda_handler` in `handler.py` | Decorated with `@instrument_lambda`; called by Lambda runtime |
| Processing | `_process_event` in `handler.py` | `async` coroutine; iterates SNS records, calls SDK |
| Infrastructure (external) | `autom8y_slack`, `autom8y_log`, `autom8y_telemetry` | All platform SDKs, not in this repo |

There is no internal `core/` or `utils/` layer. The service delegates all formatting, delivery, and observability to platform SDKs.

**Import direction:** `lambda_handler` -> `_process_event` -> SDK calls. No circular imports possible (single file).

## Entry Points and API Surface

**Lambda entry point:** `slack_alert.handler.lambda_handler`

Declared as the container CMD in `services/slack-alert/Dockerfile` line 92:
```
CMD ["slack_alert.handler.lambda_handler"]
```

**Invocation contract:**

```python
def lambda_handler(
    event: dict[str, Any] | None = None,
    context: Any = None,
) -> dict[str, Any]:
```

- `event`: Standard Lambda event dict. Expected shape: `{"Records": [{"Sns": {"Message": "<JSON string>"}}]}`
- `context`: Standard Lambda context object (unused)
- Returns: `{"statusCode": 200, "body": "<JSON with processed/errors counts>"}` on success
- Raises: `RuntimeError` if any SNS record fails Slack delivery (surfaces to CloudWatch Lambda Error metric)

**Trigger:** AWS SNS subscription. SNS delivers CloudWatch Alarm state-change notifications.

**Instrumentation decorator:** `@instrument_lambda` from `autom8y_telemetry.aws` wraps `lambda_handler`. This adds OpenTelemetry tracing around the invocation.

**No HTTP surface.** This is not an API Gateway Lambda. It has no routes, no HTTP handlers, no web framework.

**Environment variables consumed:**

| Env var | Default | Purpose |
|---------|---------|---------|
| `SLACK_CHANNEL` | `#platform-alerts` | Slack channel to post to |
| `SLACK_BOT_TOKEN_ARN` | (required) | AWS SSM/Secrets Manager ARN; resolved transparently by `SlackConfig` at runtime |

Secrets are not injected at deploy time. The AWS Parameters and Secrets Lambda Extension (installed in the container image at `/opt/extensions/`) resolves ARNs at runtime via `localhost:2773`.

## Key Abstractions

**`lambda_handler` (sync entry point):**
Bridges the synchronous Lambda runtime interface to the async `_process_event` coroutine using `asyncio.run()`. Handles the "fail the invocation if any record errored" policy.

**`_process_event` (async processor):**
Per-record processor. Iterates over `event["Records"]`, parses each SNS `Message` field as JSON (a CloudWatch Alarm payload), formats it, and posts to Slack. Uses try/except per record to ensure all records are attempted before surfacing errors.

**`SlackConfig` (from `autom8y_slack`):**
Settings object instantiated with no arguments. Resolves `SLACK_BOT_TOKEN` via `_ARN` env var auto-resolution pattern from `autom8y_config`/`autom8y_slack`. The `# type: ignore[call-arg]` annotation on line 31 of `handler.py` documents that mypy cannot see the ARN-resolution magic.

**`format_cloudwatch_alarm` (from `autom8y_slack`):**
Formatter function that converts a raw CloudWatch alarm dict into a Block Kit `SlackMessage` (with `.text` and `.blocks` attributes). The handler then calls `client.post_message(channel=..., text=..., blocks=[...])`.

**`@instrument_lambda` (from `autom8y_telemetry.aws`):**
Decorator that wraps Lambda handlers with OpenTelemetry instrumentation. Applied at module load time.

**Error handling policy (load-bearing):**
Per-record isolation via try/except inside `_process_event` ensures all records are attempted. After all records are processed, the invocation is deliberately failed (`raise RuntimeError(...)`) if any record errored. This surfaces Slack delivery failures to CloudWatch Lambda Error metrics, enabling SLO burn-rate alerting. Without this raise, delivery failures would be invisible. This is documented in an inline comment at `handler.py` lines 79-83.

## Data Flow

```
SNS topic
  |
  | (CloudWatch Alarm state-change notification as JSON string in Sns.Message)
  v
lambda_handler(event, context)
  |
  | asyncio.run()
  v
_process_event(event)
  |
  | for each record in event["Records"]:
  |   record["Sns"]["Message"] -> json.loads() -> alarm dict
  |   format_cloudwatch_alarm(alarm) -> SlackMessage(text, blocks)
  |   SlackClient.post_message(channel, text, blocks)
  |
  v
Slack Web API (chat.postMessage)
  |
  v
Returns {"processed": N, "errors": M}
  |
  | if errors > 0: raise RuntimeError (-> CloudWatch Lambda Error metric)
  | else: return {"statusCode": 200, "body": JSON}
  v
Lambda runtime (success or error recorded in CloudWatch)
```

**Secret resolution path (sidecar, not in flow above):**
At cold start, `SlackConfig()` calls out to `localhost:2773` (AWS Parameters and Secrets Lambda Extension) to resolve `SLACK_BOT_TOKEN_ARN` -> actual bot token. This happens before the first `async with SlackClient(config=config)`.

## Knowledge Gaps

- The `autom8y_slack`, `autom8y_log`, and `autom8y_telemetry` SDK internals are not in this service's source tree. Their behavior (retry logic, rate limiting, Block Kit schema) is documented in those SDKs' own `.know/` files if they exist.
- No Terraform IaC for this Lambda is in scope for this audit (would be in a separate `terraform/` directory at the repo root).
- The error behavior of `@instrument_lambda` when `lambda_handler` raises is not documented in this service.
