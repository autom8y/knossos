---
domain: architecture
generated_at: "2026-03-16T00:04:02Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The `pull-payments` service is a Python 3.12 Lambda function that syncs Stripe invoice and refund data into the autom8y platform. It lives at `services/pull-payments/` and is built with `hatchling`. The single installable package is `pull_payments`, located under `src/pull_payments/`.

**Directory layout:**

```
services/pull-payments/
├── pyproject.toml              — build config, deps, tool config
├── secretspec.toml             — env var inventory (documentation only)
├── README.md
├── Dockerfile
├── Justfile
├── src/
│   └── pull_payments/          — top-level package
│       ├── __init__.py         — version declaration (0.1.0)
│       ├── handler.py          — Lambda entry point
│       ├── config.py           — settings (pydantic-settings)
│       ├── orchestrator.py     — sync logic (invoices + refunds)
│       ├── models.py           — domain dataclasses
│       ├── staging.py          — S3/tmp batch staging
│       ├── replay.py           — staged batch replay
│       ├── metrics.py          — CloudWatch metric emission
│       └── clients/            — sub-package: external API clients
│           ├── __init__.py     — re-exports DataServiceClient, BatchResult, BusinessInfo
│           ├── data_service.py — autom8y-data HTTP client
│           └── models.py       — Pydantic API response models
└── tests/                      — 13 test files (pytest, asyncio_mode=auto)
```

**Module inventory:**

| Module | Purpose | Key types | Role |
|--------|---------|-----------|------|
| `handler` | Lambda entry point; wires event → orchestrator → response | `lambda_handler` (decorated with `@instrument_lambda`) | Surface (leaf: imported by nothing) |
| `config` | Typed settings via pydantic-settings; `lru_cache` singleton | `Settings`, `get_settings()`, `clear_settings_cache()` | Shared leaf imported by all modules |
| `orchestrator` | Core sync logic: chunk dates, fetch Stripe, resolve businesses, build payment dicts, batch write | `sync_payments`, `sync_payments_for_range`, `sync_refunds_for_range`, `categorize_refund`, `chunk_date_range`, `chunk_date_range_explicit` | Hub (imports config, models, clients, metrics) |
| `models` | Domain dataclasses | `SyncResult`, `ReplayResult`, `RefundAttribution`, `BatchResult`, `Business`, `PaymentRecord`, `StagedBatchMetadata` | Pure leaf (no internal imports) |
| `staging` | S3-backed batch staging with /tmp fallback | `stage_batch`, `list_staged_keys`, `load_staged_batch`, `delete_staged_key`, `move_to_dead_letter`, `move_to_corrupt`, `migrate_tmp_to_s3`, `StagedBatch` | Mid-layer (imports config, models) |
| `replay` | Time-budgeted S3 batch replay with circuit breaker coordination | `replay_staged_batches` | Mid-layer (imports models, staging; TYPE_CHECKING: clients.data_service) |
| `metrics` | CloudWatch replay metric emission | `emit_replay_metrics` | Leaf (imports models; calls autom8y_telemetry) |
| `clients/data_service` | `DataServiceClient`: business lookup + batch payment write + count + delete | `DataServiceClient` | Hub within clients (imports config, staging, clients/models) |
| `clients/models` | Pydantic parsing models for HTTP responses | `BusinessLookupResponse`, `BatchCreateResponse`, `CountResponse`, `DeleteResponse` | Pure leaf |

**Hub / leaf classification:**
- **Hub**: `orchestrator` (imports config, models, clients, metrics, autom8y_stripe, autom8y_interop), `clients/data_service` (imports config, staging, clients/models, autom8y_http, autom8y_interop)
- **Pure leaf (no internal imports)**: `models`, `clients/models`
- **Shared configuration leaf**: `config` (imported by handler, orchestrator, staging, replay, clients/data_service)

---

## Layer Boundaries

The codebase exhibits a clear three-layer architecture. Imports always flow downward:

```
Layer 1: Lambda Surface
  handler.py
      |
      | imports sync_payments, get_settings
      v
Layer 2: Orchestration / Business Logic
  orchestrator.py
      |
      | imports DataServiceClient, get_settings, models, metrics
      |           (uses autom8y_stripe externally)
      v
Layer 3: Infrastructure
  clients/data_service.py   staging.py   replay.py   metrics.py
      |                        |              |            |
      | all import config      |              |            |
      | clients/data_service   |              |            |
      |  imports staging       |              |            |
      | replay imports staging |              |            |
      v
Layer 0: Pure Data / Config
  models.py    config.py    clients/models.py
```

**Import direction rules observed:**
- `handler` → `orchestrator`, `config` only. Never imports infrastructure directly.
- `orchestrator` → `clients/data_service`, `config`, `models`, `metrics`. Never imports `staging` or `replay` directly (those are encapsulated inside `DataServiceClient.__aenter__`).
- `clients/data_service` → `staging` (for failure path). `replay` is imported lazily inside `__aenter__` to break the circular dependency that would arise from `replay` → `clients/data_service` → `replay`.
- `replay` → `staging`, `models`. References `DataServiceClient` only under `TYPE_CHECKING` (not at runtime) to prevent the circular import.
- `staging` → `config`, `models`. No upward imports.
- `models` → nothing (stdlib only).
- `clients/models` → nothing (pydantic only).

**Circular dependency avoidance pattern:**
- `replay.py` imports `DataServiceClient` only for type annotation under `if TYPE_CHECKING:`, while `clients/data_service.py` imports `replay_staged_batches` lazily inside `__aenter__`. This avoids the `replay` ↔ `data_service` circular dependency at import time (`src/pull_payments/replay.py:30–31`, `src/pull_payments/clients/data_service.py:69`).

**Boundary enforcement pattern:**
- No cross-layer shortcutting found. The handler never reaches into `staging` or `clients` directly. The orchestrator never touches `replay` directly.

---

## Entry Points and API Surface

### Lambda Entry Point

**File:** `src/pull_payments/handler.py`

```python
@instrument_lambda
def lambda_handler(event: dict[str, Any] | None = None, context: Any = None) -> dict[str, Any]:
```

- Triggered by AWS EventBridge on schedule: `cron(15 0,12,23 * * ? *)` (3x daily at 00:15, 12:15, 23:15 UTC).
- Reads optional `days_back` override from the event dict; defaults to `settings.default_days_back` (3 days).
- Calls `asyncio.run(sync_payments(days_back))` — bridges sync Lambda runtime to async orchestrator.
- On success: emits dead-man's-switch metric `Autom8y/PullPayments/SuccessTimestamp` and returns `statusCode: 200` with full `SyncResult` fields serialized to JSON.
- On failure: logs exception and re-raises to trigger Lambda retry / DLQ.

### Public Orchestration Functions

**File:** `src/pull_payments/orchestrator.py`

| Function | Signature | Purpose |
|----------|-----------|---------|
| `sync_payments` | `async (days_back: int \| None) -> SyncResult` | Main entry: computes date range from `days_back`, delegates to `sync_payments_for_range` |
| `sync_payments_for_range` | `async (start, end, chunk_size_days, dry_run) -> SyncResult` | Full sync over explicit range; manages StripeClient + DataServiceClient context managers |
| `sync_refunds_for_range` | `async (stripe, data, start, end, chunk_size, dry_run) -> tuple[int,int,int,int]` | Refund pass; called from within `sync_payments_for_range` |
| `categorize_refund` | `(refund_id, refund_amount, invoice, charge, vertical_id, categorize_fn) -> RefundAttribution` | Pure function: 4-case refund attribution |
| `chunk_date_range` | `(days_back, chunk_size) -> Iterator[tuple[datetime,datetime]]` | Relative range chunker |
| `chunk_date_range_explicit` | `(start, end, chunk_size) -> Iterator[tuple[datetime,datetime]]` | Absolute range chunker |

### DataServiceClient API

**File:** `src/pull_payments/clients/data_service.py`

| Method | HTTP endpoint | Purpose |
|--------|--------------|---------|
| `get_business_by_stripe_id(stripe_id)` | `GET /api/v1/businesses/lookup?stripe_id=...` | Resolve Stripe customer to autom8y business |
| `batch_create_payments(payments, *, skip_staging)` | `POST /api/v1/payments/batch` with `write_mode=insert` | Insert-only batch write; stages on failure |
| `count_payments(filters)` | `POST /api/v1/payments/count` | Filter-based payment count |
| `delete_payments(filters, confirm)` | `POST /api/v1/payments/bulk-delete` | Filter-based bulk delete (used for null-guid cleanup) |

On entry (`__aenter__`), `DataServiceClient` triggers `replay_staged_batches` to drain previously staged failures before starting the current sync.

### Exported Types (clients package)

`src/pull_payments/clients/__init__.py` re-exports: `DataServiceClient`, `BatchResult`, `BusinessInfo`.

---

## Key Abstractions

### 1. `Settings` (`src/pull_payments/config.py`)
- **Package:** `pull_payments.config`
- **Inherits from:** `LambdaServiceSettingsMixin` → `Autom8yBaseSettings` → `pydantic_settings.BaseSettings`
- **Purpose:** Typed, cached configuration singleton. All secret fields use `SecretStr` for automatic redaction.
- **Key fields:** `autom8y_data_url`, `service_api_key` (with `AliasChoices` for R-0/R-1 key pattern), `default_days_back`, `chunk_size_days`, `batch_size`, `staging_s3_bucket`, `replay_time_budget_seconds`, `max_replay_attempts`, `autom8y_env`.
- **Access pattern:** All modules call `get_settings()` (lru_cache); tests call `clear_settings_cache()` between cases.

### 2. `SyncResult` (`src/pull_payments/models.py`)
- **Package:** `pull_payments.models`
- **Type:** `@dataclass`
- **Purpose:** Accumulator for a full sync run. Fields: `invoices_processed`, `payments_written`, `payments_skipped`, `businesses_not_found`, `refunds_processed`, `refunds_written`, `refunds_skipped`, `guid_cleanup_count`, `duration_seconds`, `circuit_breaker_state`, `replay_result: ReplayResult | None`.
- **Usage:** Created in `orchestrator.sync_payments_for_range`, mutated in-place, returned to handler, serialized to Lambda response JSON.

### 3. `ReplayResult` (`src/pull_payments/models.py`)
- **Package:** `pull_payments.models`
- **Purpose:** Result of one replay cycle; fields: `replayed_count`, `failed_count`, `dead_lettered_count`, `corrupt_count`, `skipped_count`, `migrated_from_tmp`, `elapsed_seconds`, `remaining_staged`.
- **Usage:** Created by `replay.replay_staged_batches`, stored on `DataServiceClient._last_replay_result`, propagated to `SyncResult.replay_result`, reported in handler response and CloudWatch metrics.

### 4. `RefundAttribution` (`src/pull_payments/models.py`)
- **Package:** `pull_payments.models`
- **Purpose:** Output of `categorize_refund`. Contains `records: list[dict]` (1..N payment dicts with negative amounts) and `method: str` (`"exact"` | `"partial"` | `"proportional"` | `"direct_charge"`).

### 5. `DataServiceClient` (`src/pull_payments/clients/data_service.py`)
- **Package:** `pull_payments.clients.data_service`
- **Inherits from:** `autom8y_http.resilience.BaseDataServiceClient`
- **Purpose:** Async context manager wrapping all autom8y-data API calls. Injects W3C trace headers. Triggers replay on entry. Has circuit breaker state (`circuit_breaker_state`) surfaced from base class.
- **Design pattern:** async context manager (`__aenter__` / `__aexit__`); all external usage is `async with DataServiceClient() as data:`.

### 6. `StagedBatch` / staging functions (`src/pull_payments/staging.py`)
- **Package:** `pull_payments.staging`
- **Purpose:** Durable failure buffer. Failed payment batches are written to S3 `active/` prefix as JSON objects (key format: `active/{timestamp}_{uuid8}.json`). Falls back to `/tmp/staged_batches/` if S3 is unconfigured or fails. `StagedBatch` is a slot class: `key`, `payments`, `metadata`.
- **S3 prefixes:** `active/` (pending retry), `dead-letter/` (exhausted after `max_replay_attempts`), `corrupt/` (unparseable).

### 7. `StagedBatchMetadata` (`src/pull_payments/models.py`)
- **Package:** `pull_payments.models`
- **Purpose:** Tracks attempt count, staged time, invocation ID, source, and original key for each staged batch. Enables dead-letter escalation after `max_replay_attempts`.

### 8. Pydantic response models (`src/pull_payments/clients/models.py`)
- **Package:** `pull_payments.clients.models`
- **Purpose:** Parsing boundary only. `BusinessLookupResponse`, `BatchCreateResponse` (200/207), `CountResponse`, `DeleteResponse`. Callers receive domain dataclasses (`BusinessInfo`, `BatchResult`), not Pydantic models.

### Design Patterns Observed
- **Dead-man's switch**: `emit_success_timestamp(NAMESPACE)` in handler; Grafana alerts when metric is absent or stale >16h.
- **Circular import avoidance via lazy import + TYPE_CHECKING guard**: `replay.py` / `clients/data_service.py` pair.
- **Dry-run mode**: `sync_payments_for_range` and `sync_refunds_for_range` accept `dry_run=True`; runs full Stripe pipeline but skips writes.
- **Partial 207 re-staging**: On HTTP 207, failed records (non-duplicate errors) are extracted by index and re-staged as a new smaller batch (`ADR-WS2-004`).
- **Circuit breaker coordination**: Replay respects `circuit_breaker_state` from `BaseDataServiceClient`; halts on `open`, allows one probe on `half_open`.

---

## Data Flow

### Primary Path: Invoice Sync

```
EventBridge trigger
  → lambda_handler (handler.py)
  → asyncio.run(sync_payments(days_back))
  → sync_payments_for_range(start, end, chunk_size)
      → [replay phase on DataServiceClient.__aenter__]
          → migrate_tmp_to_s3() — drain legacy /tmp files
          → list_staged_keys() — enumerate S3 active/ prefix
          → for each batch: batch_create_payments(skip_staging=True)
          → on success: delete_staged_key
          → on exhaustion: move_to_dead_letter
      → StripeClient.get_invoices(chunk_start, chunk_end)  [async generator]
          → [filter: status == "paid" only]
          → DataServiceClient.get_business_by_stripe_id(invoice.customer_id)
              → GET /api/v1/businesses/lookup?stripe_id=...
              → returns BusinessInfo | None
          → StripeClient.get_subscription(invoice.subscription_id)
              → extract_vertical(sub.description) → vertical_id
          → for each line_item in invoice.line_items:
              → stripe.categorize_product(line_item.description) → ProductMatch
              → build event_id, billing_reason
              → append payment dict to batch
              → when batch >= batch_size (100):
                  → DataServiceClient.batch_create_payments(batch)
                      → POST /api/v1/payments/batch {payments, write_mode:"insert"}
                      → on HTTP 200/207: return BatchResult
                      → on circuit open or exception: stage_batch(payments)
      → sync_refunds_for_range(stripe, data, start, end, chunk_size)
          → StripeClient.get_refunds(chunk_start, chunk_end)  [async generator]
              → [filter: status == "succeeded" only]
              → stripe.get_charge(refund.charge_id) → Charge
              → DataServiceClient.get_business_by_stripe_id(charge.customer_id)
              → categorize_refund(refund_id, amount, invoice, charge, vertical_id, categorize_fn)
                  → returns RefundAttribution (1..N records with negative amounts)
              → batch + flush same as invoice path
      → _cleanup_null_guid(data, start, end)
          → DataServiceClient.delete_payments(filters, confirm=True)
              → POST /api/v1/payments/bulk-delete
      → emit_replay_metrics(data.last_replay_result)
          → emit_business_metrics(namespace, metrics)
              → CloudWatch PutMetricData
  → emit_success_timestamp(NAMESPACE) [dead-man's switch]
  → return {statusCode: 200, body: JSON(SyncResult)}
```

### Configuration / Environment Variable Handling

Resolved once at first `get_settings()` call (lru_cache), then reused. Key variables (from `secretspec.toml`):

| Env Var | Settings field | Purpose |
|---------|---------------|---------|
| `AUTOM8Y_ENV` | `autom8y_env` | Deployment environment |
| `AUTOM8Y_DATA_URL` | `autom8y_data_url` | Data service base URL |
| `AUTOM8Y_AUTH_URL` | `auth_base_url` (inherited) | Auth service URL |
| `PULL_PAYMENTS_SERVICE_KEY` | `service_api_key` (R-0) | Service-specific API key |
| `SERVICE_API_KEY` | `service_api_key` (R-1 fallback) | Platform fallback key |
| `STRIPE_API_KEY` | consumed by `autom8y-stripe` SDK | Stripe credentials |
| `STAGING_S3_BUCKET` | `staging_s3_bucket` | S3 bucket for failed batch durability |
| `REPLAY_TIME_BUDGET_SECONDS` | `replay_time_budget_seconds` | Max replay time per invocation |
| `MAX_REPLAY_ATTEMPTS` | `max_replay_attempts` | Dead-letter threshold |
| `DEFAULT_DAYS_BACK` | `default_days_back` | Look-back window |
| `BATCH_SIZE` | `batch_size` | Records per write batch |

In Lambda prod: secrets resolved via AWS Parameters and Secrets Lambda Extension (HTTP to localhost:2773). Locally: direct env var reads. `LambdaServiceSettingsMixin` provides the extension resolution behavior.

### External Service Interactions

| System | SDK / library | Interaction |
|--------|--------------|-------------|
| Stripe | `autom8y-stripe` (`StripeClient`) | Paginated async generators: `get_invoices`, `get_refunds`, `get_charge`, `get_subscription`; product categorization; vertical extraction |
| autom8y-data | `DataServiceClient` (this service) over `autom8y-http` | Business lookup, batch payment create, count, bulk delete |
| AWS CloudWatch | `autom8y-telemetry.aws` | Dead-man's switch metric; replay metrics (`StagedBatchCount`, `ReplayedCount`, etc.) |
| AWS S3 | `boto3` (lazy) | Staged batch storage: `active/`, `dead-letter/`, `corrupt/` prefixes |
| AWS Lambda runtime | `autom8y-telemetry.aws.instrument_lambda` | Decorator wrapping; OTLP span export |

## Knowledge Gaps

- The `Justfile` and `scripts/` directory were not read; operational commands and script content are not documented here.
- `docker-compose.override.yml` and `Dockerfile` were not read; container build specifics are not captured.
- The `autom8y-stripe` SDK internals (`StripeClient`, `categorize_product`, `extract_vertical`, `build_event_id`, `build_billing_reason`) were not read as they live outside this service's source.
- The `BaseDataServiceClient` circuit breaker implementation (in `autom8y-http`) is not documented; only its interface (`circuit_breaker_state`, `_ensure_http()`) is observable here.
- The README's listed endpoint URLs (`/api/v1/business/stripe/{stripe_id}`, `/api/v1/payments/crud/batch`) appear stale relative to the actual client code (`/api/v1/businesses/lookup`, `/api/v1/payments/batch`). These README docs were not audited against the live service.
