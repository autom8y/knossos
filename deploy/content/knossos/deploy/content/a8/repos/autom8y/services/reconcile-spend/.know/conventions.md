---
domain: conventions
generated_at: "2026-03-09T00:04:44Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "8e41207"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---
# Codebase Conventions

## Error Handling Style

### Pattern Overview

This codebase uses **no custom exception types** within the service itself. All error handling relies on built-in exceptions and SDK-defined exceptions imported from platform libraries. The one project-defined exception is `DataServiceUnavailableError`, which comes from `autom8y_interop.data`, not from within `reconcile_spend`.

### Error Creation

Errors are not raised within service code in the normal business logic path. Instead, the codebase uses early-return / `return None` patterns extensively (see `rules.py`). All five rule functions return `Anomaly | None` and return `None` to indicate non-detection rather than raising.

Exception handling for date parsing in `rule_stale_account` is the only internal `try/except` that catches specific exception types:

```python
# src/reconcile_spend/rules.py:185-206
try:
    latest = datetime.fromisoformat(record.latest_payment)
    ...
except (ValueError, TypeError):
    log.warning("stale_rule_date_parse_failed", ...)
```

### Error Propagation at the Lambda Boundary

The handler (`handler.py`) uses a two-tier exception containment pattern:

**Tier 1 — Best-effort side effects (never block primary flow):**
All non-critical operations (metric emission, event publishing) are wrapped in `try/except Exception` with a `log.debug(...)` call only. These exceptions are swallowed silently at the debug level:

```python
# src/reconcile_spend/handler.py:84-87
try:
    emit_reconciliation_metrics(metrics, result, near_miss_data, outcome=outcome)
except Exception:
    log.debug("metric_emission_failed_in_handler", exc_info=True)
```

**Tier 2 — Fatal errors (re-raise to trigger Lambda retry/DLQ):**
The outer handler `try/except` catches all `Exception`, attempts best-effort metric emission, logs with `log.exception(...)`, then **re-raises**:

```python
# src/reconcile_spend/handler.py:121-143
except Exception as e:
    # ... best-effort failure metric emission ...
    log.exception("reconciliation_failed", error=str(e))
    raise
```

### Error Handling in Client Code

The Asana resolve client (`clients/asana_resolve.py`) uses the **empty-return-on-error** pattern throughout — all failure paths return `{}` rather than raising, enabling graceful degradation:

```python
# src/reconcile_spend/clients/asana_resolve.py:38-42
try:
    response = await http.post_async(url, json=body, timeout=timeout)
except Exception:
    log.warning("asana_resolve_failed", url=url, error="request_exception")
    return {}
```

`ValidationError` from Pydantic is caught specifically (not as a bare `Exception`) to enable richer diagnostic logging:

```python
# src/reconcile_spend/clients/asana_resolve.py:52-62
except ValidationError as e:
    log.warning(
        "asana_resolve_contract_validation_failed",
        ...
        response_keys=list(raw_body.keys()) if isinstance(raw_body, dict) else None,
    )
    return {}
except Exception:
    log.warning("asana_resolve_parse_error", url=url)
    return {}
```

### Metrics Module Pattern

`metrics.py` wraps its entire body in `try/except Exception` internally and logs at `debug` level. Callers wrap metric invocations in additional `try/except` at the call site — double-guarding against metric emission failure:

```python
# src/reconcile_spend/metrics.py:154-155
except Exception:
    log.debug("metric_emission_failed", exc_info=True)
```

### SDK Circuit Breaker Exception

`DataServiceUnavailableError` (from `autom8y_interop.data`) is the one non-generic exception caught by name in the main flow. It carries structured attributes (`service_name`, `method`, `time_remaining`) and is handled by posting a Slack degradation alert and returning early — not by re-raising:

```python
# src/reconcile_spend/orchestrator.py:320-348
except DataServiceUnavailableError as exc:
    log.warning("data_service_unavailable", service=exc.service_name, ...)
    result.circuit_breaker_state = "open"
    # ... post Slack alert ...
    return result, _empty_near_miss_data()
```

### Logging Style

All logging uses **structlog-style keyword arguments** (no f-string messages). Log messages are `snake_case` event strings. The logger is always created as a module-level `log = get_logger(__name__)`.

```python
log.info("reconciliation_started", period_days=period_days)
log.warning("missing_office_phone")
log.exception("reconciliation_failed", error=str(e))
```

---

## File Organization

### Package Structure

```
src/
  reconcile_spend/
    __init__.py          — package version export only
    handler.py           — Lambda entry point; @instrument_lambda + @metrics decorators
    orchestrator.py      — main async reconciliation flow; pure orchestration logic
    models.py            — domain dataclasses and IntEnum/StrEnum types
    config.py            — pydantic-settings Settings class + get_settings() factory
    rules.py             — pure rule functions; detect_anomalies() dispatcher
    report.py            — Slack Block Kit builder; pure rendering logic
    metrics.py           — CloudWatch EMF emission; delegates to autom8y_telemetry
    readiness.py         — pipeline readiness gate; pure function + frozen dataclass
    stubs.py             — unimplemented future classes (3-way reconciliation)
    clients/
      __init__.py        — minimal docstring only
      data_service.py    — DataInsightClient factory function
      models.py          — Pydantic response models for API boundaries
      asana_resolve.py   — async batch resolve function
```

### File-Content Responsibilities

| File | Responsibility |
|------|---------------|
| `handler.py` | Lambda entry point, decorator application, async dispatch via `asyncio.run()` |
| `orchestrator.py` | Full orchestration flow: fetch -> parse -> detect -> enrich -> report -> post |
| `models.py` | All domain dataclasses (`ClientRecord`, `Anomaly`, `EnrichedAnomaly`, `ReconciliationResult`) and domain enums |
| `config.py` | `Settings` class (pydantic-settings) and `get_settings()` cached factory |
| `rules.py` | Five pure rule functions + `detect_anomalies()` dispatcher |
| `report.py` | All Slack Block Kit block construction; rendering helpers prefixed with `_` |
| `metrics.py` | All CloudWatch EMF emission; `emit_reconciliation_metrics()` and `emit_dms_success_timestamp()` |
| `readiness.py` | Pipeline readiness gate; `check_pipeline_readiness()` pure function |
| `stubs.py` | Architectural placeholder classes; not imported in active flow |
| `clients/data_service.py` | Single factory function `create_data_client()` |
| `clients/models.py` | Pydantic `BaseModel` response contracts for HTTP APIs |
| `clients/asana_resolve.py` | Single async function `batch_resolve_units()` |

### Separation Philosophy

**Strict separation of parsing from domain logic.** `clients/models.py` contains Pydantic models used only at the HTTP boundary. They are explicitly **not** domain models. The orchestrator transforms `ReconciliationRow` (Pydantic) into `ClientRecord` (dataclass) via `parse_client_records()` before passing to domain logic.

**Pure functions preferred over classes.** Rule functions, the readiness check, and the report builder are all module-level pure functions — not class methods. Classes are reserved for data containers (dataclasses) and SDK integration (Settings inheriting from pydantic-settings).

**Private helpers use underscore prefix.** Module-level helpers not intended for external use are prefixed with `_`:
- `report.py`: `_format_data_quality`, `_format_freshness`, `_sanitize_mrkdwn_label`, `_count_by_category`, `_render_anomaly_sections`
- `orchestrator.py`: `_empty_near_miss_data`, `_build_stale_data_alert`, `_build_circuit_open_alert`

### Module-Level Initialization Pattern

Module-level state is kept minimal and explicit:
- `log = get_logger(__name__)` — always module-level
- `tracer = get_tracer(__name__)` — module-level in orchestrator
- `metrics = create_metrics("ReconcileSpend")` — module-level in handler
- `_event_publisher: "EventPublisher | None" = None` — lazy singleton with getter

### `__init__.py` Usage

`src/reconcile_spend/__init__.py` exports only `__version__`. It does **not** re-export any domain symbols. `clients/__init__.py` contains only a docstring. No public re-export surface is maintained.

### Constants

Module-level constants in `report.py`:
```python
MAX_BLOCKS = 50
RESERVED_BLOCKS = 10
```

Namespace constant in `metrics.py`:
```python
NAMESPACE = "Autom8y/ReconcileSpend"
```

### Configuration Factory Pattern

Config uses `@lru_cache` on the factory function, not class-level caching. `clear_settings_cache()` is provided for test teardown:

```python
# src/reconcile_spend/config.py:137-146
@lru_cache
def get_settings() -> Settings:
    return Settings()

def clear_settings_cache() -> None:
    get_settings.cache_clear()
    Autom8yBaseSettings.reset_resolver()
```

---

## Domain-Specific Idioms

### Edge Case Codes (E-prefixed identifiers)

The codebase uses a labeling convention of `E{N}` to reference known edge cases across code, comments, and docstrings. These are project-internal defect/edge-case tracking codes, not Python error codes:

| Code | Meaning |
|------|---------|
| E4 | Skip rows with NULL/empty `office_phone` |
| E5 | Skip rows where both spend and collected are zero |
| E9 | Treat negative collected as 0 |
| E10 | Skip rows with non-finite (NaN/Inf) values |
| DEF-4 | Defect reference for NaN silently suppressing anomaly detection |

These codes appear in docstrings, inline comments, and ADR-style comments. They are not defined as constants.

### Rule Numbering Convention

The five anomaly detection rules are named `rule_{name}` and referred to as R1-R5 throughout comments and docstrings. They are individually defined pure functions plus a `detect_anomalies()` dispatcher — not a class hierarchy.

### Budget-Aware Path

A recurring two-path pattern in `rules.py`: budget-aware path (using `expected_collection` / `expected_variance`) takes precedence over the naive variance path (using raw `variance_pct`). This is explicitly labeled with comments:

```python
# Budget-aware path: compare collected vs expected_collection
if record.expected_collection is not None and record.expected_collection > 0 ...:
    ...
    return None  # <- explicit: do not fall through to naive path

# Fallback: naive collected-vs-spend variance
if record.variance_pct is not None and record.variance_pct > threshold_pct:
    ...
```

### Open-World Accounting Posture

A recurring design philosophy labeled in comments: surface gaps rather than hide them. Staleness data causing a missing `staleness_seconds` field is treated as PASS (assume fresh) rather than FAIL. Missing optional fields degrade gracefully.

### POC/PROTOTYPE Shortcut Comments

Prototype-quality code is labeled with `# POC {N}:` and `# PROTOTYPE SHORTCUT:` prefixes. Production-intended replacements are noted inline. Example in `handler.py` for the `autom8y_events` integration.

### Async Pattern

The service is primarily async (`async with`, `await`), driven at the Lambda boundary via `asyncio.run(run_reconciliation(...))`. Client factories return async context managers. The handler itself is synchronous (Lambda requirement) and bridges via `asyncio.run()`.

### `to_safe_dict()` Pattern

`Settings` (via `Autom8yBaseSettings`) exposes `to_safe_dict()` for logging configuration without leaking `SecretStr` values. Referenced in docstring but not called within this service's code directly.

### Pydantic v2 Model Patterns

- `model_validate(dict)` for construction from dicts (not `parse_obj`)
- `ConfigDict(extra="ignore")` — lenient validation; upstream schema evolution does not break parsing
- All `ReconciliationRow` fields are `Optional` with `None` defaults (lenient contract)
- `SecretStr` for secrets; `.get_secret_value()` at call site only

### Inline Import Pattern

One instance of a deferred import inside a function body to avoid circular dependency or lazy-load a heavy module:

```python
# src/reconcile_spend/orchestrator.py:209
from reconcile_spend.clients.asana_resolve import batch_resolve_units
```

Also in `orchestrator.py:309`:
```python
from datetime import timedelta
```

---

## Naming Patterns

### Type / Class Names

| Pattern | Example |
|---------|---------|
| PascalCase for all types | `ClientRecord`, `Anomaly`, `EnrichedAnomaly`, `ReconciliationResult` |
| Noun-noun composition | `ReconciliationRow`, `ReconciliationResult`, `ReadinessResult`, `AsanaResolveItem` |
| Action+Target for "result of operation" types | `ReconciliationResult`, `ReadinessResult` |
| Domain noun for value object types | `ClientRecord`, `Anomaly`, `EnrichedAnomaly` |
| `Status` suffix for StrEnum/IntEnum status enumerations | `ReadinessStatus`, `AnomalyCategory` |

### Function / Method Names

| Pattern | Example |
|---------|---------|
| `snake_case` universally | `detect_anomalies`, `parse_client_records`, `check_pipeline_readiness` |
| `rule_{noun}` for individual rules | `rule_ads_running_no_payment`, `rule_paying_no_ads`, `rule_overbilled` |
| `build_{noun}` for constructors returning collections | `build_report`, `build_all_clear_report` |
| `emit_{noun}` for side-effect metric functions | `emit_reconciliation_metrics`, `emit_dms_success_timestamp` |
| `create_{noun}` for factory functions | `create_data_client` |
| `get_{noun}` for cached accessor functions | `get_settings` |
| `_build_{noun}` for private Slack block builders | `_build_stale_data_alert`, `_build_circuit_open_alert` |
| `_format_{noun}` for private formatting helpers | `_format_data_quality`, `_format_freshness` |
| `_render_{noun}` for private rendering helpers | `_render_anomaly_sections` |
| `_count_{noun}` / `_sanitize_{noun}` for private utilities | `_count_by_category`, `_sanitize_mrkdwn_label` |
| `batch_{verb}_{noun}` for bulk async operations | `batch_resolve_units` |

### Variable Names

| Pattern | Example |
|---------|---------|
| `snake_case` universally | `period_days`, `variance_pct`, `overbilled_threshold_pct` |
| `_pct` suffix for percentage floats | `variance_pct`, `overbilled_threshold_pct`, `coverage_pct` |
| `_seconds` suffix for time durations in seconds | `staleness_seconds`, `duration_seconds`, `max_staleness_seconds` |
| `_minutes` suffix for time durations in minutes | `staleness_minutes`, `threshold_minutes` |
| `_days` suffix for time durations in days | `period_days`, `stale_days_threshold` |
| `_count` suffix for integer tallies | `anomaly_count`, `unreconciled_count` |
| `_url` suffix for URL strings | `stripe_dashboard_url`, `invoice_url`, `asana_unit_url` |
| `_blocks` for Slack Block Kit lists | `blocks`, `stale_blocks` |
| `_client` for HTTP client instances | `data_client`, `slack_client` |
| `raw_` prefix for pre-validated data | `raw_data`, `raw_row_count`, `raw_body` |
| `enriched_` prefix for augmented objects | `enriched_anomalies` |

### Module / File Names

All `snake_case`, noun-focused:
- `handler.py`, `orchestrator.py`, `models.py`, `config.py`, `rules.py`, `report.py`, `metrics.py`, `readiness.py`, `stubs.py`
- Client subpackage: `data_service.py`, `models.py`, `asana_resolve.py`

### Constants

`SCREAMING_SNAKE_CASE` for module-level constants:
- `MAX_BLOCKS`, `RESERVED_BLOCKS` — in `report.py`
- `NAMESPACE` — in `metrics.py`
- `_EVENTS_AVAILABLE`, `_LEAKED_ENV_VARS` — `SCREAMING_SNAKE_CASE` with `_` prefix for private module-level constants

### Metric Names

CloudWatch metric names use `PascalCase` (AWS convention), not `snake_case`:
- `AccountsAnalyzed`, `AnomalyCount`, `RunDuration`, `AllClear`, `RunSuccess`

### Enum Label Properties

`AnomalyCategory` enum uses `.label` (human-readable string) and `.severity` as computed properties, not `str(enum)` or `.name`:

```python
cat.label  # -> "Ads Running, No Payment"
cat.severity  # -> "high"
```

---

## Knowledge Gaps

1. **`autom8y_config.Autom8yBaseSettings` internals**: The `LambdaServiceSettingsMixin`, `ExtensionSecret`, and `reset_resolver()` conventions are inherited from platform SDK. Their full behavior (especially `reset_resolver()`) is not visible from this service.

2. **`autom8y_interop.data` contract**: The `DataInsightClient`, `DataServiceUnavailableError`, `PERIOD_MAP`, and `resolve_insight_client()` signatures are imported but not visible in this repo. Circuit breaker state machine details are opaque.

3. **ruff.toml base config**: `pyproject.toml` extends `../../ruff.toml`. The inherited rule set is not visible here, making it impossible to know which Ruff rules are base-enabled vs. locally extended.

4. **`autom8y_log.get_logger` structlog configuration**: The structured logging format (JSON vs. text, field ordering, timestamp format) is configured in the platform package, not here. Only the invocation style (`configure_logging()` at module load; `get_logger(__name__)` per module) is observable.

5. **`stubs.py` lifecycle**: The file is explicitly documented as non-imported. There is no mechanism visible in the codebase that would prevent it from being imported accidentally.
