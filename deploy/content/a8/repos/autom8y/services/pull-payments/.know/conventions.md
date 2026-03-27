---
domain: conventions
generated_at: "2026-03-16T00:04:02Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

**Project**: `pull-payments` — Stripe payment sync Lambda service (Python 3.12)
**Source root**: `services/pull-payments`

---

## Error Handling Style

### Pattern: Catch-Log-Return, never Raise at boundaries

The dominant error handling pattern is **catch at service call boundaries, log with structured event name, return a safe fallback value**. Exceptions are almost never re-raised at the client/integration layer; they are absorbed and the caller receives `None`, `0`, or an empty `BatchResult`. The sole exception is `handler.py`, which deliberately re-raises to trigger Lambda retry and DLQ.

### Error creation

Errors are never created in this codebase. No custom exception classes exist. Error types consumed are from the SDK and stdlib: `StripeAPIError`, `StripeNotFoundError`, `StripeError`, `CircuitBreakerOpenError`, `json.JSONDecodeError`, `KeyError`, `OSError`.

### Exception specificity pattern

Handlers always distinguish the SDK-specific error (`CircuitBreakerOpenError`) from the general case (`Exception`). The pattern is:

```python
except CircuitBreakerOpenError as exc:
    log.warning("circuit_breaker_open", service="autom8y-data", method=..., time_remaining=exc.time_remaining)
    ...
    return safe_fallback
except Exception as exc:
    log.error("event_name", service="autom8y-data", error=str(exc))
    ...
    return safe_fallback
```

Files: `src/pull_payments/clients/data_service.py` (lines 100–133, 210–231, 273–282, 315–324)

### Stripe error handling in orchestrator

In orchestrator.py the pattern narrows further — the first `except` handles network-like Stripe errors, the second handles API errors:

```python
except (StripeAPIError, StripeNotFoundError) as e:
    log.warning("charge_lookup_failed", refund_id=..., error=str(e))
    continue
```

```python
except (StripeError, ConnectionError, TimeoutError) as e:
    log.warning("subscription_fetch_failed", subscription_id=..., error=str(e))
```

These appear at lines 414–421 and 454–459 in `src/pull_payments/orchestrator.py`.

### Structured log event names

All `log.*()` calls use a **snake_case event-name string as positional first argument**, followed by keyword context fields. There are no format strings or f-strings in log calls. Examples:

```python
log.info("handler_invoked", days_back=days_back)
log.warning("circuit_breaker_open", service="autom8y-data", method="get_business_by_stripe_id", time_remaining=exc.time_remaining, stripe_id=stripe_id)
log.error("staged_file_corrupt", key=key, error=str(exc))
```

Logger is always obtained at module level via `log = get_logger(__name__)` from `autom8y_log`.

### Boundary re-raise

Only `handler.py` re-raises after logging, explicitly to trigger Lambda retry/DLQ:

```python
except Exception as e:
    log.exception("sync_failed", error=str(e))
    raise
```

### Staging as error-recovery idiom

When `batch_create_payments` fails (circuit breaker or exception), the pattern is: **stage to S3, return a `BatchResult(success_count=0, failure_count=len(payments))`**. This is project-specific graceful degradation, not standard Python error handling:

```python
if not skip_staging:
    stage_batch(payments)
return BatchResult(success_count=0, failure_count=len(payments))
```

File: `src/pull_payments/clients/data_service.py` lines 217–231.

---

## File Organization

### src/ layout

All production source lives under `src/pull_payments/`. Tests are under `tests/`. Scripts (dev utilities) under `scripts/`. This is the standard `src/` layout with hatchling build backend.

```
src/pull_payments/
  __init__.py          # version string only
  handler.py           # Lambda entry point (@instrument_lambda)
  orchestrator.py      # core sync logic (chunk, fetch, categorize, batch-write)
  staging.py           # S3 batch staging + /tmp fallback
  replay.py            # time-budgeted replay of staged batches
  models.py            # all dataclasses (domain models)
  config.py            # Settings + get_settings() + clear_settings_cache()
  metrics.py           # service-specific CloudWatch metric emission
  clients/
    __init__.py        # re-exports DataServiceClient, BatchResult, BusinessInfo
    data_service.py    # DataServiceClient (extends BaseDataServiceClient)
    models.py          # Pydantic response-parsing models (NOT domain models)
```

### One concern per file

Each file has a single, clearly stated concern declared in its module docstring. The mapping is:
- `handler.py` — Lambda entry + instrument decorator
- `orchestrator.py` — sync orchestration (chunking, invoice loop, refund loop)
- `staging.py` — S3 staging mechanics
- `replay.py` — replay loop algorithm
- `models.py` — dataclass definitions only
- `config.py` — settings class + cached accessor + cache-clear
- `metrics.py` — CloudWatch emission for replay metrics
- `clients/models.py` — Pydantic parse-boundary models only (NOT domain models; comment in docstring clarifies this explicitly)
- `clients/data_service.py` — HTTP client implementation

### Two-tier model split

There are **two** distinct model files with different purposes:
1. `src/pull_payments/models.py` — `@dataclass` domain models (`SyncResult`, `ReplayResult`, `BatchResult`, `Business`, `PaymentRecord`, etc.)
2. `src/pull_payments/clients/models.py` — `pydantic.BaseModel` HTTP response schemas (`BusinessLookupResponse`, `BatchCreateResponse`, etc.)

The docstring in `clients/models.py` states explicitly: "These models validate JSON at the parsing boundary. They are NOT domain models."

### Module-level logger and tracer

Every module that does I/O or computation declares at module level:

```python
log = get_logger(__name__)
_tracer = trace.get_tracer("autom8y.payments")
```

Private globals use leading underscore (`_tracer`, `_s3_client`).

### Lazy-init singleton for boto3

`staging.py` uses a module-level `_s3_client: Any = None` with a getter `_get_s3_client()` that imports boto3 inside the function. This is documented as "ADR-WS2-003" and avoids cold-start overhead.

### TYPE_CHECKING guard

All imports used only for type hints are placed under `if TYPE_CHECKING:` to avoid circular imports. Pattern:

```python
if TYPE_CHECKING:
    from pull_payments.models import SyncResult
```

Files: `handler.py` line 18, `data_service.py` line 31, `replay.py` line 29.

### `from __future__ import annotations`

Used in files that have forward references to break cycles: `config.py`, `replay.py`, `data_service.py`.

### `__all__` at the boundary

`clients/__init__.py` and `clients/data_service.py` both declare `__all__` to control the public surface of the `clients` package.

### `scripts/` is development tooling

The `scripts/` directory contains ad-hoc dev/ops scripts (`backfill.py`, `dry_run.py`, `inspect_refunds.py`, `test_single_write.py`). These are NOT part of the deployed package (`[tool.hatch.build.targets.wheel]` lists only `src/pull_payments`).

---

## Domain-Specific Idioms

### OTel tracing: every public function wraps in a span

Every substantive function (async or sync) wraps its body in a `with _tracer.start_as_current_span(...)` context manager with span attributes set at entry and result metrics set at exit. The tracer name is always `"autom8y.payments"`. Span names follow `pull_payments.{module}.{operation}` dot-notation, e.g.:

- `pull_payments.period.extract`
- `pull_payments.refund.categorize`
- `pull_payments.replay.run`
- `pull_payments.batch.stage`

`span.set_attribute()` is called with `pull_payments.*` prefixed keys. `span.add_event()` is used for branch outcomes (e.g., `"refund.case.direct_charge"`, `"replay.batch.corrupt"`).

### `record_side_effect()` for every external mutation

Every write or delete to an external system (S3, data service) must call `record_side_effect(span, system=..., operation=..., target=..., payload={...})` after the call. This is a telemetry SDK convention for tracking side effects in the OTel span. Systems used: `"s3"`, `"data_service"`, `"cloudwatch"`, `"filesystem"`.

### Chunked date-range iteration

Sync is always performed in date chunks, not full ranges. `chunk_date_range()` and `chunk_date_range_explicit()` yield `(start, end)` tuples. The orchestrator always loops `for chunk_start, chunk_end in chunk_date_range_explicit(start, end, chunk_size):`.

### Batch-and-flush pattern

Payment records are accumulated in a `batch: list[dict[str, Any]] = []` and flushed when `len(batch) >= settings.batch_size`. After the loop, a final flush handles the tail. This pattern appears in both `sync_payments_for_range` and `sync_refunds_for_range`.

### `get_settings()` + `@lru_cache`

Settings are accessed through `get_settings()` (cached via `@lru_cache`), never instantiated directly outside `config.py`. Tests clear the cache via `clear_settings_cache()` and `get_settings.cache_clear()`. The settings singleton is passed into functions as `settings = get_settings()` at function start, not stored on `self`.

### S3 staging key format

Staged batches use the key format `{prefix}/{timestamp}_{uuid8}.json`. Prefixes: `active/` (pending), `dead-letter/` (exhausted), `corrupt/` (unparseable). The `/tmp/staged_batches/` path is legacy and migrated to S3 on replay via `migrate_tmp_to_s3()`.

### `dry_run` parameter pattern

Both `sync_payments_for_range` and `sync_refunds_for_range` accept `dry_run: bool = False`. When `dry_run=True`, all Stripe API calls proceed (fetching invoices, resolving businesses) but no writes are issued. The count fields still reflect what would have been written.

### `DataPaymentProtocol` for type checking

Orchestrator uses `DataPaymentProtocol` from `autom8y_interop.data` as the type annotation for the `data` parameter, enabling protocol-based injection for testing.

### `NAMESPACE = "Autom8y/PullPayments"`

The CloudWatch namespace is a module-level constant declared in both `handler.py` and `metrics.py`. The full path of metric emission is: `emit_success_timestamp(NAMESPACE)` for the dead-man's-switch and `emit_business_metrics(namespace=NAMESPACE, metrics=[...])` for replay metrics.

---

## Naming Patterns

### Classes

- **Domain dataclasses**: PascalCase noun phrases — `SyncResult`, `ReplayResult`, `BatchResult`, `RefundAttribution`, `StagedBatchMetadata`, `Business`, `PaymentRecord`, `StagedBatch`
- **Pydantic parse-boundary models**: PascalCase + `Response`/`Data` suffix — `BusinessLookupResponse`, `BusinessData`, `BatchCreateResponse`, `BatchErrorDetail`, `BatchResultItem`, `CountResponse`, `DeleteResponse`
- **Settings class**: plain `Settings` (no suffix) — extends `LambdaServiceSettingsMixin, Autom8yBaseSettings`
- **Client class**: `DataServiceClient` (extends `BaseDataServiceClient`)

### Functions

- **Public async functions**: verb-noun — `sync_payments`, `sync_payments_for_range`, `sync_refunds_for_range`, `replay_staged_batches`
- **Private functions**: leading underscore + verb-noun — `_extract_period`, `_cleanup_null_guid`, `_get_s3_client`, `_get_bucket`, `_build_key`, `_stage_to_tmp`, `_move_key`, `_is_full_success`, `_has_non_duplicate_failures`, `_extract_failed_records`
- **Pure functions**: verb-noun without `async` — `categorize_refund`, `chunk_date_range`, `chunk_date_range_explicit`

### Variables

- **Module-level singletons**: leading underscore — `_s3_client`, `_tracer`
- **Loop accumulator**: `batch` (not `batches`, `results`, `items`)
- **Counter variables**: `{noun}_{verb}ed` — `refunds_processed`, `refunds_written`, `payments_skipped`, `businesses_not_found`, `replayed_count`, `dead_lettered_count`
- **Settings local**: always named `settings` — `settings = get_settings()`
- **Result accumulator**: always named `result` — `result = SyncResult()`, `result = ReplayResult()`

### File naming

All source files are `snake_case.py`. File names map directly to their primary concern: `handler.py`, `orchestrator.py`, `staging.py`, `replay.py`, `models.py`, `config.py`, `metrics.py`.

### Log event names

All log event names are `snake_case` string literals. They follow `{noun}_{verb}` or `{verb}_{noun}` patterns. Examples: `"handler_invoked"`, `"sync_failed"`, `"batch_staged_to_s3"`, `"staged_replay_completed"`, `"business_not_found_for_invoice"`, `"circuit_breaker_open"`.

### OTel span names

Span names use dot-notation: `pull_payments.{module_noun}.{operation}`. Examples: `pull_payments.period.extract`, `pull_payments.refund.categorize`, `pull_payments.replay.run`, `pull_payments.batch.stage`, `pull_payments.invoice_sync.run`, `pull_payments.refund_sync.run`.

### OTel attribute keys

Span attributes use dot-notation with `pull_payments.` prefix: `pull_payments.refund.id`, `pull_payments.refund.amount_cents`, `pull_payments.replay.replayed_count`, `pull_payments.staging.backend`, `pull_payments.lookup.stripe_id`.

### Environment variable naming

Settings fields use `snake_case` names. The canonical env var is the uppercased field name. Service-specific alias is declared via `AliasChoices`:

```python
service_api_key: SecretStr = Field(
    validation_alias=AliasChoices("PULL_PAYMENTS_SERVICE_KEY", "SERVICE_API_KEY"),
)
```

`AUTOM8Y_ENV` is read directly (bypasses `env_prefix`), documented as "ADR-ENV-NAMING-CONVENTION Decision 1".

### Acronym conventions

`guid` is lowercase in identifiers (`guid_cleanup_count`, `null_guid`). Stripe IDs (`cus_xxx`, `re_xxx`, `sub_xxx`) appear only in string literals and comments, not in identifier names.

---

## Knowledge Gaps

1. **`scripts/` internals not read**: `backfill.py`, `dry_run.py`, `inspect_refunds.py`, `test_single_write.py` were not examined. They may contain additional one-off idioms or patterns not reflected above.
2. **`autom8y_http.resilience.BaseDataServiceClient` internals unknown**: The base class provides `_ensure_http()`, `circuit_breaker_state`, `__aenter__`/`__aexit__` lifecycle. Its implementation is in an external SDK not in this repo. The conventions above describe only what is observable in this service.
3. **`autom8y_config.LambdaServiceSettingsMixin` contract**: Provides `auth_base_url`, `service_api_key_value`, `_SERVICE_KEY_ALIAS` hook. Full MRO behavior is not observable without the SDK source.
4. **Ruff rules from `../../ruff.toml`**: The root-level ruff config is extended but not examined. Some lint rules may impose additional naming or style conventions not captured above.
