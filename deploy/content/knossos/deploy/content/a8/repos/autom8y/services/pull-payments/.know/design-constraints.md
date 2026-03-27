---
domain: design-constraints
generated_at: "2026-03-16T00:04:02Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

**Service**: pull-payments
**Source tree**: `services/pull-payments/src/pull_payments/`
**Supporting files**: `scripts/`, `tests/`, `pyproject.toml`

---

## Tension Catalog Completeness

### TENSION-001: Dual PaymentRecord Model (Type: naming mismatch / duplication)

**Location**: `src/pull_payments/models.py:109` and `scripts/dry_run.py:52`

Two distinct `PaymentRecord` dataclasses exist in the same repository, serving different purposes. The service's canonical model (`models.py`) is a full domain type with `to_dict()`. The script's model (`dry_run.py`) is a lightweight display type with `vertical_name` and `business_name` display fields. Neither imports the other; they duplicate field definitions.

- `models.py:PaymentRecord` — canonical, fully typed, has `to_dict()`
- `dry_run.py:PaymentRecord` — script-only, adds `vertical_name`/`business_name`, uses `str` for `created` (not `datetime`)

**Historical reason**: dry_run.py was written as a standalone debugging tool before the canonical model was fully stabilized. It was never updated to import from `models.py`.

**Ideal resolution**: dry_run.py should import and extend `models.PaymentRecord` or use it directly, not define a duplicate.

**Cost**: Low maintenance risk today (script is not in the critical path), but will diverge if canonical model fields change.

---

### TENSION-002: Dual chunk_date_range Functions (Type: duplication / naming mismatch)

**Location**: `src/pull_payments/orchestrator.py:64` (`chunk_date_range`) and `orchestrator.py:84` (`chunk_date_range_explicit`), also `scripts/dry_run.py:94` (third copy)

The orchestrator defines two chunking helpers with different signatures:
- `chunk_date_range(days_back, chunk_size=15)` — computes start/end from `days_back` relative to now
- `chunk_date_range_explicit(start, end, chunk_size=15)` — takes explicit datetimes

`dry_run.py:94` implements a third, independent `chunk_date_range` that duplicates the first variant exactly. Both the module and script compute `now - days_back` independently.

**Historical reason**: The explicit variant was added later when `sync_payments_for_range` was introduced for backfill support. The script predates that refactor and was not updated.

**Ideal resolution**: Export both chunking functions from `orchestrator.py`; have `dry_run.py` import `chunk_date_range` from the module instead of redefining it.

**Cost**: Medium. The drift means bug fixes to one copy are not propagated to others.

---

### TENSION-003: StagedBatch as Non-Dataclass Class (Type: under-engineering / structural inconsistency)

**Location**: `src/pull_payments/staging.py:63`

`StagedBatch` is a manual `__slots__` class rather than a dataclass or Pydantic model. All other value objects in the service (`StagedBatchMetadata`, `ReplayResult`, `SyncResult`, `BatchResult`, `Business`, `PaymentRecord`) are dataclasses. The comment at line 59 says "re-exported from models for backward compatibility" but the class is defined fresh in `staging.py` — it is not actually in `models.py`.

**Historical reason**: The comment is misleading; `StagedBatch` was likely written inline to avoid circular import issues between `staging.py` and `models.py`. The `__slots__` approach was a quick choice rather than a deliberate design.

**Ideal resolution**: Convert to `@dataclass(slots=True)` and move to `models.py` alongside other domain types, or confirm the circular import concern and add an explanatory comment.

**Cost**: Low risk but creates conceptual inconsistency. Misleading comment risks future confusion.

---

### TENSION-004: settings_override Fixture Patches Three Import Sites (Type: layering violation / over-engineering in tests)

**Location**: `tests/conftest.py:43-48`

The `settings_override` fixture patches `get_settings` at three explicit module paths: `pull_payments.config`, `pull_payments.orchestrator`, and `pull_payments.clients.data_service`. This means any new module that calls `get_settings()` directly will silently bypass the test override.

**Historical reason**: `lru_cache` on `get_settings()` means patching the function at the source doesn't propagate — callers have already captured the cached value at import time. Each call site must be individually patched.

**Ideal resolution**: Use `monkeypatch.setenv` + `clear_settings_cache()` (which the `mock_settings` fixture already does correctly) and avoid module-level `lru_cache` caching, OR make modules call `get_settings()` lazily inside function bodies only (not at module scope). The `mock_settings` fixture is the correct pattern; `settings_override` is a workaround.

**Cost**: Medium. Any new module calling `get_settings()` at import time will escape the test harness silently.

---

### TENSION-005: `businesses_not_found` Accumulates Even When Business Is Not Required (Type: logic tension / unclear semantics)

**Location**: `src/pull_payments/orchestrator.py:635` (invoice sync), `orchestrator.py:432` (refund sync)

When a business lookup returns `None`, the service logs `businesses_not_found` and increments the counter — but then **continues processing** the invoice/refund with `office_phone=None` and `vertical_id=1` (default). The counter name implies a skip, but the behavior is degraded write, not skip. The handler response exposes this counter at line 88 in `handler.py`.

**Historical reason**: The original design allowed writes with null `office_phone` to capture revenue data even for unmatched customers. The counter was added for observability but its name suggests skips.

**Ideal resolution**: Rename to `businesses_unmatched` or add a separate `payments_written_without_business` counter to clarify that unmatched does not mean skipped.

**Cost**: Low code impact but causes operational confusion when reading metrics.

---

### TENSION-006: Dual `BatchResult` Export from `data_service.py` (Type: naming mismatch / re-export confusion)

**Location**: `src/pull_payments/clients/data_service.py:34`

```python
__all__ = ["BatchResult", "BusinessInfo", "DataServiceClient"]
```

`BatchResult` and `BusinessInfo` are imported from `autom8y_interop.data.models` and re-exported from this module. However, `pull_payments/models.py:90` also defines its own `BatchResult` dataclass. These are two different `BatchResult` types — one from the shared interop SDK, one local. If a caller imports `BatchResult` from `data_service`, they get the interop version. If they import from `models`, they get the local version.

**Historical reason**: The interop SDK's `BatchResult` was adopted to satisfy the `DataPaymentProtocol` interface. The local `BatchResult` in `models.py` appears to be a legacy type or a parallel definition that was not cleaned up.

**Ideal resolution**: Audit all usages; if the local `BatchResult` is redundant, delete it and use the interop version. If it serves a different purpose (e.g., includes `errors: list`), document the distinction.

**Cost**: Medium risk. Import errors are possible if a developer confuses the two.

---

## Trade-off Documentation

### TENSION-001 Trade-off: Script Independence vs. DRY

The `dry_run.py` script's duplicate `PaymentRecord` is a deliberate trade-off for script portability. Scripts in `scripts/` are often run in isolation or without the full dependency graph. The current approach sacrifices DRY for operational convenience. The trade-off persists because scripts are low-priority refactor targets.

### TENSION-002 Trade-off: Explicit Range API vs. Relative API

`chunk_date_range` (relative) and `chunk_date_range_explicit` (absolute) coexist because the Lambda entry point uses relative ranges (`days_back`) while the backfill script needs explicit month boundaries. Both are useful. The only cost is the orphaned third copy in `dry_run.py`, which should be removed.

### TENSION-003 Trade-off: `StagedBatch` Inline Class

The `__slots__` pattern was chosen to avoid importing `StagedBatchMetadata` from `models.py` into `staging.py` at the time of writing — except `staging.py` already imports `StagedBatchMetadata` (line 22 of `staging.py`). The circular import concern, if it existed, no longer applies. This is a trade-off that was never revisited.

### TENSION-004 Trade-off: `lru_cache` on `get_settings()`

Caching settings is correct for Lambda: the interpreter is long-lived between invocations and settings never change. The cost is test friction — callers must be patched individually. The `clear_settings_cache()` escape hatch and `mock_settings` fixture manage this adequately for the current three-module scope, but will not scale cleanly.

### TENSION-005 Trade-off: Write-With-Null vs. Skip-Unmatched

Writing payment records with `office_phone=None` was a deliberate product decision: Stripe revenue should be captured even when the business lookup fails (e.g., a test customer or a cancelled account). The trade-off is accepting data quality degradation (null phone) in exchange for complete revenue capture.

### S3 Staging vs. /tmp Fallback

The staging system uses S3 as primary and `/tmp` as fallback. This is documented in `staging.py:5-6` (FR-12). The trade-off: S3 provides cross-invocation durability; `/tmp` is ephemeral per Lambda instance. The `migrate_tmp_to_s3()` function handles migration of legacy `/tmp` files. The constraint: if S3 is unconfigured (`staging_s3_bucket = ""`), the entire replay system is disabled — `list_staged_keys()` returns `[]` immediately when bucket is empty.

### Circuit Breaker State as String

`circuit_breaker_state` is typed as `str` in `SyncResult` and defaults to `"unknown"`. The actual circuit breaker object lives in `BaseDataServiceClient` (SDK). The local service propagates it as a raw string rather than an enum. This was a deliberate choice to avoid coupling to the SDK's internal enum type.

---

## Abstraction Gap Mapping

### Missing Abstraction: Batch Flush Logic (Duplication)

**Location**: `orchestrator.py` — four separate batch flush blocks

The pattern of "accumulate a batch list, flush when `len(batch) >= settings.batch_size`, then flush remaining after the loop" appears four times:
1. Invoice sync inner loop (line ~698)
2. Invoice sync outer flush (line ~719)
3. Refund sync inner loop (line ~502)
4. Refund sync outer flush (line ~523)

Each block is 12-15 lines and nearly identical. The `dry_run` branching doubles the conditional logic. A `BatchFlusher` or async context manager abstraction would remove this duplication. No such abstraction exists.

**Maintenance burden**: Changes to batch behavior (e.g., adding telemetry, changing `skip_staging` logic) must be made in all four locations.

---

### Zombie Abstraction: Legacy `/tmp` Functions

**Location**: `staging.py:297-331`

Three functions — `list_staged_files()`, `load_staged_file(path)`, `remove_staged_file(path)` — are explicitly labeled "Legacy function retained for backward compatibility with existing tests." These functions wrap `/tmp`-based behavior that has been superseded by the S3-backed system. They serve no production code path; their only consumers are tests.

These are zombie abstractions: they exist to preserve test compatibility, not production behavior. They add surface area for confusion (callers may use the legacy API thinking it is current).

---

### Premature Abstraction: `DataPaymentProtocol` in Orchestrator

**Location**: `orchestrator.py:28` — imports `DataPaymentProtocol` from `autom8y_interop.data`

The `_cleanup_null_guid` function is typed against `DataPaymentProtocol`, but `DataServiceClient` is the only concrete implementation in this service. The protocol adds an indirection layer that is not exercised by any alternate implementation. This is a premature abstraction that exists because `DataPaymentProtocol` is the SDK-defined interface — not because the service has multiple write-path implementations.

**Evidence**: `sync_payments_for_range` at line 603 uses `DataServiceClient()` directly (not the protocol), so the protocol typing is not even consistently applied within the same file.

---

### Missing Abstraction: Vertical Resolution Logic (Duplication)

**Location**: `orchestrator.py:640-652` (invoice sync) and `orchestrator.py:444-466` (refund sync)

The logic to resolve a vertical from a subscription — "if subscription_id, fetch subscription, extract vertical from description, fall back to business default, fall back to 1" — is implemented twice with nearly identical error handling. A `_resolve_vertical(stripe, invoice_or_charge, business)` helper function is absent.

---

## Load-Bearing Code Identification

### LOAD-001: `stage_batch()` in `staging.py`

**File**: `src/pull_payments/staging.py:84`

**Callers**: `data_service.py` (3 call sites in `batch_create_payments`), `replay.py` (2 call sites)

This function is the sole write path for failed batches. If it raises an unhandled exception, payment data is lost with no fallback. The current implementation catches all exceptions from S3 and falls back to `/tmp`, so it is defensively written. However:

- The `/tmp` fallback path (`_stage_to_tmp`) does NOT call `record_side_effect` on the span inside the except block of `stage_batch` — the span attribute for `backend` is set but the side-effect is recorded only in the non-exception S3 path and the explicit `bucket is empty` path. There is a subtle observability gap when `/tmp` fallback fires due to S3 exception.
- If `/tmp` is also unavailable (disk full on Lambda), the `OSError` from `path.write_text` is not caught and will propagate, causing the calling function to fail.

**Naive-fix failure mode**: Removing the S3 fallback would silently lose data on S3 errors. Removing `/tmp` fallback entirely would cause unhandled `OSError` on disk-full.

**Safe-refactor requirement**: Any change to `stage_batch` must preserve both the S3 path and the `/tmp` fallback, and any new fallback must handle `OSError`.

---

### LOAD-002: `_extract_failed_records()` in `replay.py`

**File**: `src/pull_payments/replay.py:377`

This function is called during partial 207 re-staging. If `batch_result.errors` is `None`, it returns the entire original payment list (`return original_payments`). This means a 207 with no error detail causes the entire batch to be re-staged as "failed" — potentially causing data duplication on retry.

**Dependents**: `replay_staged_batches()` at line 253 — this is the only caller, but it is the critical re-staging decision point.

**Naive-fix failure mode**: Returning an empty list when errors is None would silently drop data. The current behavior (return all) is the safer default but risks duplicates.

**Safe-refactor requirement**: Any change must ensure the total payment count after re-staging equals the actual failure count. Changing the behavior requires coordinating with the dedup logic in the data service (`duplicate_event_id` error code).

---

### LOAD-003: `_is_full_success()` in `replay.py`

**File**: `src/pull_payments/replay.py:356`

Controls whether a staged batch is deleted after replay. The fallback branch (line 364-365):
```python
# success_count > 0 with no error details -> treat as success (existing behavior)
return batch_result.success_count > 0
```
This means a batch with `success_count=3, failure_count=2, errors=None` is treated as full success and the original batch is deleted. Partial failures without error details are swallowed.

**Dependents**: All replay execution paths.

**Safe-refactor requirement**: Changing this logic requires understanding whether the data service API can return `failure_count > 0` with no `errors` in the response body. The comment "existing behavior" suggests this was an intentional trade-off, not an oversight.

---

### LOAD-004: `autom8y_env` AliasChoices in `config.py`

**File**: `src/pull_payments/config.py:105-108`

```python
autom8y_env: Autom8yEnvironment = Field(
    default=Autom8yEnvironment.LOCAL,
    validation_alias=AliasChoices("AUTOM8Y_ENV"),
)
```

The comment "bypasses env_prefix per ADR-ENV-NAMING-CONVENTION Decision 1" is load-bearing documentation. The `model_config` has `env_prefix=""`, so all fields read directly from env var names without prefix. If `env_prefix` were changed, `autom8y_env` would break (would look for `{prefix}AUTOM8Y_ENV`). This field participates in the SDK's URL guard logic (`Autom8yBaseSettings`).

**Safe-refactor requirement**: Any change to `env_prefix` in `SettingsConfigDict` requires auditing all field names and validation aliases simultaneously.

---

### LOAD-005: `service_api_key` AliasChoices in `config.py`

**File**: `src/pull_payments/config.py:54-57`

```python
service_api_key: SecretStr = Field(
    validation_alias=AliasChoices("PULL_PAYMENTS_SERVICE_KEY", "SERVICE_API_KEY"),
    ...
)
```

The dual alias `(PULL_PAYMENTS_SERVICE_KEY, SERVICE_API_KEY)` is the R-1 fallback pattern. `PULL_PAYMENTS_SERVICE_KEY` is the service-specific canonical name; `SERVICE_API_KEY` is the generic fallback. Removing either alias without coordinating with Terraform IaC would cause the Lambda to fail to start.

---

## Evolution Constraint Documentation

### Area: S3 Staging Infrastructure

**Changeability rating**: COORDINATED (requires IaC change + code change)

**Evidence**: `staging_s3_bucket` env var (config.py:84). If empty, staging is silently disabled (no error, no warning beyond log). This means adding S3 staging in production requires:
1. Provisioning the S3 bucket in Terraform
2. Setting `STAGING_S3_BUCKET` in the Lambda environment
3. Verifying `migrate_tmp_to_s3()` handles any existing `/tmp` files

The staging bucket is not validated at startup — an empty string is accepted silently. There is no guard that logs a warning when staging is disabled in production.

---

### Area: Replay Algorithm

**Changeability rating**: COORDINATED (algorithm has documented spec reference)

**Evidence**: `replay.py:1` — "See TDD-write-path-resilience.md for the full algorithm specification." The algorithm is specified externally. Changes to circuit breaker state handling, time budget logic, or dead-letter thresholds must be reconciled with that spec. The `max_replay_attempts` and `replay_time_budget_seconds` are configurable via settings.

---

### Area: Service Auth Key Aliases

**Changeability rating**: MIGRATION-REQUIRED

**Evidence**: `config.py:54-57` — dual alias. Removing `SERVICE_API_KEY` as a fallback requires coordinating with all Terraform environment variable definitions to ensure `PULL_PAYMENTS_SERVICE_KEY` is set everywhere before the alias is dropped.

---

### Area: Legacy `/tmp` Functions

**Changeability rating**: SAFE (test-only dependency)

**Evidence**: `staging.py:297-331` — three legacy functions. These are not called in production code paths. They can be removed after updating the tests that depend on them (`tests/test_staging.py` and others). No Terraform or Lambda environment change required.

---

### Area: Batch Payload Shape (dict, not dataclass)

**Changeability rating**: COORDINATED (interface with data service API)

**Evidence**: `orchestrator.py:663-695` — payment records are built as raw `dict[str, Any]` and sent to `data.batch_create_payments(batch)`. The `PaymentRecord` dataclass in `models.py` exists but is NOT used in the actual write path (the orchestrator builds dicts inline). Changing the payload shape requires coordinating with the `autom8y_data` API schema.

The `PaymentRecord.to_dict()` method at `models.py:151` is a dead code path in the current orchestrator — the orchestrator builds its own dicts.

---

### Area: `asyncio.run()` in Lambda Handler

**Changeability rating**: FROZEN (Lambda runtime constraint)

**Evidence**: `handler.py:50` — `asyncio.run(sync_payments(days_back))`. AWS Lambda does not provide a persistent event loop between invocations. `asyncio.run()` creates a new event loop per invocation. This is intentional and correct for Lambda. Any refactor to use a persistent loop would require Lambda container reuse awareness and would break the isolation model.

---

### Deprecated Markers and Compatibility Shims

- `staging.py:28-29`: `STAGED_DIR = Path("/tmp/staged_batches")` and `CORRUPT_DIR = STAGED_DIR / "corrupt"` — legacy constants retained for backward compat. Marked as "Legacy /tmp paths -- retained for migration and fallback".
- `staging.py:297`: `list_staged_files()` — explicit "Legacy function" docstring.
- `staging.py:305`: `load_staged_file(path)` — explicit "Legacy function retained for backward compatibility with existing tests."
- `staging.py:322`: `remove_staged_file(path)` — explicit "Legacy function retained for backward compatibility with existing tests."
- `staging.py:59-60`: Comment "Dataclass re-exported from models for backward compatibility" — inaccurate (the class is not in models.py).

---

## Risk Zone Mapping

### RISK-001: Silent Data Loss When Both S3 and /tmp Fail

**Location**: `src/pull_payments/staging.py:136-148` (`stage_batch` except block)

**Type**: Silent failure — `OSError` from `/tmp` write is not caught

**Evidence**: `_stage_to_tmp()` at line 339 calls `path.write_text(...)` without exception handling. If Lambda's `/tmp` volume is full (disk quota exhausted), this raises `OSError`, which propagates uncaught through `stage_batch()` and into `batch_create_payments()`, which also does not catch `OSError`. The exception would reach the orchestrator's batch flush block, which only catches the `Exception` type for the `data.batch_create_payments()` call — not for the staging write.

**Recommended guard**: Wrap `_stage_to_tmp()` with `try/except OSError` and log a critical event when staging completely fails. Accept the data loss at that point but record it durably (e.g., a CloudWatch metric increment).

**Cross-reference**: LOAD-001

---

### RISK-002: Unguarded Input in `delete_payments` / `count_payments`

**Location**: `src/pull_payments/clients/data_service.py:257-292` and `294-340`

**Type**: Unvalidated filter dict passed to external API

**Evidence**: Both `count_payments` and `delete_payments` accept `filters: dict[str, Any]` and pass them directly to the data service API with no local validation. The backfill script (`scripts/backfill.py:82-88`) constructs these filter dicts inline. If the filter structure is malformed (wrong `op` values, missing `logic` key), the error surfaces as an HTTP 400/500 from the data service, not a local validation error.

**Recommended guard**: Add a lightweight Pydantic model for `FilterGroup` / `FilterCondition` at the `DataServiceClient` boundary to catch malformed filter construction before the HTTP call.

---

### RISK-003: Stripe Refund `businesses_not_found` Continues Processing

**Location**: `src/pull_payments/orchestrator.py:432-433`

**Type**: Silent fallback — business not found does not skip refund

**Evidence**: In the refund sync path, when a business is not found, the code increments the counter and continues. `vertical_id` defaults to `(business.default_vertical_id if business else None) or 1` (line 438). `office_phone` is set to `business.office_phone if business else None` in the payment dict (line 490). A refund record IS written with null office_phone and vertical_id=1.

Compare to the invoice sync path (line 627-635): identical pattern. The behavior is consistent, but if the intent is to skip unmatched refunds (as the `businesses_not_found` counter name suggests), the guard is absent.

**Cross-reference**: TENSION-005

---

### RISK-004: `load_staged_batch` Sends All Exceptions to `corrupt/`

**Location**: `src/pull_payments/staging.py:185-196`

**Type**: Overly broad exception classification

**Evidence**: The general `except Exception` branch in `load_staged_batch` calls `move_to_corrupt(key)` for any S3 API error (e.g., access denied, throttling, network timeout). A transient S3 error causes the batch to be permanently moved to `corrupt/`, which is designed for unrecoverable data corruption. A throttled S3 read is not corruption — but the batch will be classified as corrupt and never retried.

The comment `# BUGFIX-01: mirror JSON error path to prevent infinite retry` explains why the exception was added, but the classification is too broad.

**Recommended guard**: Distinguish between S3 API errors (transient — should return `None` without moving to corrupt/) and JSON parse errors (permanent — should move to corrupt/). Only `json.JSONDecodeError` and `KeyError` on the parsed structure represent true corruption.

---

### RISK-005: `conftest.py` `mock_settings` Does Not Set `STAGING_S3_BUCKET`

**Location**: `tests/conftest.py:11-27`

**Type**: Missing test coverage for S3 staging path

**Evidence**: `mock_settings` does not set `STAGING_S3_BUCKET`. The default in `config.py:84` is `""`, which means all test invocations run with S3 staging disabled. The staging-related tests (`test_staging.py`, `test_replay.py`) likely mock S3 directly, but any test that goes through `batch_create_payments` on failure will silently fall back to `/tmp` rather than exercising the S3 path.

**Recommended guard**: Add a `staging_s3_fixture` that sets `STAGING_S3_BUCKET` to a moto-mocked bucket, and ensure at least one test validates S3 staging behavior end-to-end.

---

## Knowledge Gaps

1. **`autom8y_interop.DataPaymentProtocol`** — The full interface contract is in the SDK, not visible in this service. It is unclear which protocol methods are called in production vs. the subset used by `_cleanup_null_guid`.

2. **`autom8y_http.resilience.BaseDataServiceClient`** — The circuit breaker state machine (`open`, `half_open`, `closed`) and the `circuit_breaker_state` property are implemented in the SDK. The behavior of `half_open` probing is consumed in `replay.py` but not defined here.

3. **`ADR-WS2-003`, `ADR-WS2-004`, `ADR-ENV-NAMING-CONVENTION`** — Referenced in code comments but the ADR documents are not in this service's tree. Their content is assumed from context.

4. **`TDD-write-path-resilience.md`** — Referenced in `replay.py:8` as the spec for the replay algorithm. File location unknown; not found in this service's tree.

5. **`BUGFIX-01`** label in `staging.py:186` — No associated issue tracker link or description of the original infinite-retry bug that motivated the fix.

6. **`PaymentRecord.to_dict()` callers** — The method exists in `models.py` but the orchestrator does not use it. Whether any external caller (script, test) uses it is not confirmed from the source tree.
