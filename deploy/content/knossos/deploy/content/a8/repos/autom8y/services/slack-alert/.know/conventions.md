---
domain: conventions
generated_at: "2026-03-16T20:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

This service uses a **per-record isolation with deferred invocation failure** pattern.

### Error Creation

Errors are raised as standard Python `RuntimeError` with an f-string message. There are no custom exception classes or exception hierarchies in this service.

```
# services/slack-alert/src/slack_alert/handler.py, line 85
raise RuntimeError(f"Failed to deliver {result['errors']} of {result['processed']} alerts")
```

### Error Propagation: Per-Record Isolation

Processing is structured as a counter-aggregate pattern: all records are attempted regardless of individual failures. Each record is wrapped in a bare `except Exception` block that increments an error counter rather than aborting the loop. This ensures partial delivery (some records succeed, some fail) is always attempted.

```python
# handler.py, lines 57-59
except Exception:
    log.exception("failed_to_post_alert")
    errors += 1
```

### Error Reporting at Lambda Boundary

After all records are processed, the invocation is failed if any errors occurred. This is a deliberate choice documented in an inline comment: the raise is required so CloudWatch Lambda Error metrics are populated, enabling SLO burn-rate alerting. Without it, Slack delivery failures (e.g. `invalid_auth`, `channel_not_found`) would be invisible to the 99.9% SLO.

```python
# handler.py, lines 84-85
if result["errors"] > 0:
    raise RuntimeError(f"Failed to deliver {result['errors']} of {result['processed']} alerts")
```

### Logging at Error Sites

Exceptions are logged via `log.exception(...)` with a snake_case event string argument. The `log.exception` call captures the full traceback automatically. No manual traceback extraction. Structured keyword arguments are added to success log calls (`log.info`) but NOT at error sites (the `exception` call carries the exception implicitly).

### No Error Codes or Categorization

No error code system, no custom exception hierarchy, no error wrapping (e.g., `raise X from Y`). Errors are logged and counted; only the aggregate count surfaces to the Lambda runtime.

### Error Handling at Config Boundary

`SlackConfig()` is called with `# type: ignore[call-arg]` comment because the `bot_token` field is resolved at runtime via an `_ARN` env var by `Autom8yBaseSettings`. No explicit error handling wraps the config construction — if config fails, the Lambda invocation fails with an unhandled exception.

## File Organization

### Package Structure

The service follows the `src/` layout convention:

```
services/slack-alert/
├── src/
│   └── slack_alert/
│       ├── __init__.py          # Package docstring only
│       └── handler.py           # All Lambda logic
├── tests/
│   ├── __init__.py
│   └── test_handler.py
├── pyproject.toml
├── secretspec.toml              # Env var documentation (not runtime)
├── Dockerfile
├── Justfile
└── README.md
```

Configured in `pyproject.toml`:
```toml
[tool.hatch.build.targets.wheel]
packages = ["src/slack_alert"]
```

### Single-File Handler

All Lambda logic lives in one file: `src/slack_alert/handler.py`. There is no sub-module structure within the package. The handler module has:
- A module-level docstring describing the service architecture
- Module-level logging setup (`configure_logging()`, `log = get_logger(__name__)`) at import time
- A private async helper (`_process_event`) prefixed with `_`
- A public Lambda entry point (`lambda_handler`) decorated with `@instrument_lambda`

### `__init__.py` Convention

The package `__init__.py` contains only a one-line docstring. It does not re-export symbols. No `__all__` defined.

### `secretspec.toml`

A service-specific file that documents the complete env var surface. It is explicitly documentation-only (no runtime behavior). References a named ADR (`ADR-ENV-NAMING-CONVENTION`) for naming-tier rationale.

### Justfile Pattern

The Justfile is a thin router that imports shared primitives from the monorepo (`../../just/`, `../../devops/`). Service-specific logic lives in `just/` subdirectory (`just/_globals.just`, `just/ci.just`). Pattern: service Justfiles do not define targets directly — they compose via imports.

### Test Layout

Tests mirror the src layout but without the `slack_alert/` subdirectory. A single test file `tests/test_handler.py` covers the handler module. No conftest.py, no fixtures directory.

## Domain-Specific Idioms

### Async Lambda Bridge

The Lambda handler (`lambda_handler`) is synchronous (required by Lambda runtime) but delegates to an async inner function (`_process_event`) via `asyncio.run(...)`. This is the project idiom for async code in Lambda: use a sync entry point with `asyncio.run()` to bridge to async execution.

### `@instrument_lambda` Decorator

The Lambda entry point is decorated with `@instrument_lambda` from `autom8y_telemetry.aws`. This is an SDK-provided decorator applied to the outermost Lambda handler. The private processing logic (`_process_event`) is NOT decorated — only the public entry point is instrumented.

### Structured Logging with Event Strings

Logging uses `autom8y_log` (`configure_logging()` + `get_logger(__name__)`). Log calls use snake_case event string as the first positional argument, followed by structured keyword arguments:

```python
log.info("handler_invoked", record_count=len(event.get("Records", [])))
log.info("alert_posted", alarm_name=alarm.get("AlarmName"), state=alarm.get("NewStateValue"))
log.exception("failed_to_post_alert")
```

No f-strings or string interpolation in log calls. All contextual data is passed as kwargs.

### SNS Record Access Pattern

SNS records are accessed defensively with `.get()` chains and defaults:
```python
record.get("Sns", {}).get("Message", "{}")
```
The default for `Message` is `"{}"` (a valid empty JSON object) so `json.loads()` succeeds even on malformed records. This prevents parse errors from crashing the entire batch.

### `Autom8yBaseSettings` ARN Auto-Resolution

Secrets are not passed explicitly. `SlackConfig()` is instantiated with no arguments. The `bot_token` is resolved transparently by the `Autom8yBaseSettings` base class via an `_ARN` env var convention (`SLACK_BOT_TOKEN_ARN`). The `# type: ignore[call-arg]` comment documents this at the call site.

### Pydantic Block Kit Serialization

Pydantic models are serialized for the Slack API using `.model_dump(exclude_none=True)` — the `exclude_none=True` kwarg is the project idiom to avoid sending null fields to Slack's Block Kit API.

### `from __future__ import annotations`

All source files use `from __future__ import annotations` for deferred annotation evaluation (PEP 563). This is mandatory in conjunction with `ruff`'s `flake8-type-checking` rules, which move type-only imports into `TYPE_CHECKING` blocks.

## Naming Patterns

### Module Names

Snake_case. Package name matches service name with hyphens replaced: `slack-alert` -> `slack_alert`.

### Function Names

- Public Lambda entry point: `lambda_handler` (conventional Lambda handler name)
- Private helpers: `_process_event` (leading underscore for internal helpers not intended for external use)

### Variable Names

Snake_case throughout. Descriptive names that reflect the domain:
- `alarm`, `sns_message`, `message`, `channel` — domain-specific
- `processed`, `errors` — aggregate counters
- `record` — iteration variable for SNS records
- `client`, `config` — SDK handles named by role not type

### Log Event Strings

Snake_case string literals as first positional argument: `"handler_invoked"`, `"alert_posted"`, `"failed_to_post_alert"`. Pattern is past-tense verb phrase or noun phrase describing what occurred.

### Type Annotations

- `dict[str, Any]` (lowercase generic, Python 3.12+) — not `Dict[str, Any]`
- `Any` from `typing` for Lambda context and event payloads
- Union with `None` uses `X | None` syntax (not `Optional[X]`)
- Return types annotated on all functions

### Config / Settings Class Naming

Settings classes from external SDKs keep their upstream names (`SlackConfig`). The service has no local Settings class (documented in `secretspec.toml`).

### Test Function Names

`test_` prefix + descriptive snake_case phrase:
```python
def test_handler_module_exists():
```
No class grouping — flat test functions.

### Constant Names

Module-level constants use ALL_CAPS (not observed in Python source, but `SERVICE_NAME := "slack-alert"` in Justfile follows this convention).

### Env Var Naming Tiers

Documented in `secretspec.toml` (references `ADR-ENV-NAMING-CONVENTION`):
- Tier 1: `AUTOM8Y_*` — ecosystem-global
- Tier 2: Bare vendor names (e.g., `SLACK_*`) — third-party SDK constraint
- Tier 4: Service-scoped

## Knowledge Gaps

The following areas could not be fully observed due to the minimal size of this service:

1. **Multi-file handler conventions**: With only one handler file, it is unknown whether larger Lambda services split by concern (e.g., `models.py`, `clients.py`, `utils.py`) or stay in a single file.
2. **Pydantic Settings pattern**: This service explicitly has no local Settings class. The Settings pattern (field definitions, validators, `model_config`) present in other services is unobservable here.
3. **Error wrapping (`raise X from Y`)**: Not used in this service; unknown whether this pattern appears in more complex services.
4. **Conftest and fixture conventions**: No `conftest.py` exists. Fixture patterns are not observable from this service alone.
5. **Integration test patterns**: Only a smoke test is present. No mocking, no async test patterns, no `respx` usage despite `respx` being in dev dependencies.
6. **`TYPE_CHECKING` import blocks**: Referenced by ruff config (`flake8-type-checking` rules) but not observed in practice (the service imports are small enough that no TYPE_CHECKING block is needed).
