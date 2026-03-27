---
domain: architecture
generated_at: "2026-03-09T00:04:44Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "8e41207"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---
# Codebase Architecture

## Package Structure

The service is a Python 3.12 AWS Lambda deployed as a container image. The source root is `src/reconcile_spend/`. Two packages exist:

**`reconcile_spend` (root package)** — 10 modules, 1 `__init__.py`

| Module | File | Purpose | Key Exported Types |
|---|---|---|---|
| `__init__.py` | `src/reconcile_spend/__init__.py` | Package version declaration | `__version__` |
| `handler` | `src/reconcile_spend/handler.py` | AWS Lambda entry point; bridges Lambda runtime to async orchestrator | `lambda_handler` |
| `orchestrator` | `src/reconcile_spend/orchestrator.py` | Main async reconciliation pipeline (Steps 1-5) | `run_reconciliation`, `parse_client_records`, `enrich_anomalies`, `compute_near_miss_data` |
| `config` | `src/reconcile_spend/config.py` | Pydantic-settings configuration; cached singleton | `Settings`, `get_settings`, `clear_settings_cache` |
| `models` | `src/reconcile_spend/models.py` | Domain dataclasses and enums (pure data, no IO) | `ClientRecord`, `Anomaly`, `EnrichedAnomaly`, `ReconciliationResult`, `AnomalyCategory` |
| `rules` | `src/reconcile_spend/rules.py` | Five pure anomaly detection rule functions + dispatcher | `detect_anomalies`, `rule_ads_running_no_payment`, `rule_paying_no_ads`, `rule_overbilled`, `rule_underbilled`, `rule_stale_account` |
| `report` | `src/reconcile_spend/report.py` | Slack Block Kit report builder; enforces 50-block limit | `build_report`, `build_all_clear_report` |
| `metrics` | `src/reconcile_spend/metrics.py` | CloudWatch EMF metric emission (zero-latency stdout) | `emit_reconciliation_metrics`, `emit_dms_success_timestamp` |
| `readiness` | `src/reconcile_spend/readiness.py` | Pipeline freshness gate (PASS/WARN/FAIL three-tier) | `check_pipeline_readiness`, `ReadinessResult`, `ReadinessStatus` |
| `stubs` | `src/reconcile_spend/stubs.py` | Architectural stub for future 3-way reconciliation path; not imported in active flow | `ThreeWayComparison`, `AsanaReconciliation` |

**`reconcile_spend.clients` (sub-package)** — 3 modules

| Module | File | Purpose | Key Exported Types |
|---|---|---|---|
| `__init__.py` | `src/reconcile_spend/clients/__init__.py` | Namespace declaration only | — |
| `data_service` | `src/reconcile_spend/clients/data_service.py` | Factory for `DataInsightClient` from `autom8y-interop` | `create_data_client` |
| `models` | `src/reconcile_spend/clients/models.py` | Pydantic HTTP-boundary response models (contract layer) | `ReconciliationRow`, `InsightResponse`, `DataQuality`, `InsightMetadata`, `AsanaResolveItem`, `AsanaResolveResponse` |
| `asana_resolve` | `src/reconcile_spend/clients/asana_resolve.py` | Async batch Asana Unit GID resolver | `batch_resolve_units` |

**Hub vs leaf classification:**
- `orchestrator` is the **hub**: imports from `clients/data_service`, `clients/models`, `config`, `models`, `readiness`, `report`, and `rules`
- `handler` is the **entry hub**: imports from `config`, `metrics`, `models`, and `orchestrator`
- `models` and `readiness` are **leaf** packages: imported by others, import no siblings
- `rules` is a **leaf**: imports only from `models`
- `report` is a **leaf**: imports only from `models`
- `metrics` is a **near-leaf**: imports `models` via `TYPE_CHECKING` only (runtime-free)
- `stubs` is an **orphan**: not imported anywhere in active code

---

## Layer Boundaries

The codebase has three clear layers:

```
Layer 1: Lambda Surface
  handler.py
       |
       v
Layer 2: Application Core
  orchestrator.py
  ┌─────┬─────┬────────┬──────────┬─────────┬─────────┐
  rules  report metrics readiness  config    clients/
                                             └── data_service
                                             └── asana_resolve
       |
       v
Layer 3: Data / Domain
  models.py          (domain dataclasses)
  clients/models.py  (Pydantic HTTP-boundary contracts)
  config.py          (settings singleton)
  stubs.py           (dead/future code, no runtime imports)
```

**Import direction (enforced unidirectionally):**

- `handler` → `orchestrator`, `config`, `metrics`, `models`
- `orchestrator` → `clients/data_service`, `clients/models`, `config`, `models`, `readiness`, `report`, `rules`
- `clients/data_service` → `config` (lazy import inside function to avoid circular)
- `clients/asana_resolve` → `clients/models`
- `rules` → `models`
- `report` → `models`
- `metrics` → `models` (TYPE_CHECKING only, no runtime import)
- `readiness` → nothing (no intra-package imports)
- `models` → nothing (no intra-package imports)
- `config` → nothing from this package (only SDK base classes)

**Notable boundary patterns:**
- `clients/data_service.py` uses a deferred `from reconcile_spend.config import get_settings` inside the factory function body to avoid a potential circular import at module load time.
- `orchestrator.py` defers `from reconcile_spend.clients.asana_resolve import batch_resolve_units` inside the `enrich_anomalies` function body.
- No circular dependencies exist in the active code path.
- `stubs.py` has zero inbound imports and is not part of any active import graph.

---

## Entry Points and API Surface

**Lambda Entry Point:**

`reconcile_spend.handler.lambda_handler` — declared as the container CMD in `Dockerfile` line 103: `CMD ["reconcile_spend.handler.lambda_handler"]`

**Invocation signature:**
```python
@instrument_lambda
@metrics.log_metrics(capture_cold_start_metric=True)
def lambda_handler(
    event: dict[str, Any] | None = None,
    context: Any = None,
) -> dict[str, Any]:
```

**Lambda event inputs:**
- `event.period_days` (optional int): overrides configured `reconciliation_period_days`
- Triggered on a schedule (EventBridge); no HTTP API surface

**Lambda return contract:**
```json
{"statusCode": 200, "body": "{\"success\": true, \"accounts_analyzed\": N, ...}"}
```
On unhandled exception, re-raises to trigger Lambda retry/DLQ.

**Key exported interfaces between packages:**

| Interface | Package | Consumers |
|---|---|---|
| `get_settings() -> Settings` | `config` | `handler`, `orchestrator`, `clients/data_service` |
| `run_reconciliation(period_days) -> tuple[ReconciliationResult, dict]` | `orchestrator` | `handler` |
| `detect_anomalies(records, *, thresholds) -> list[Anomaly]` | `rules` | `orchestrator` |
| `build_report(anomalies, ...) -> list[dict]` | `report` | `orchestrator` |
| `check_pipeline_readiness(staleness_seconds, max_staleness_seconds) -> ReadinessResult` | `readiness` | `orchestrator` |
| `create_data_client() -> DataInsightClient` | `clients/data_service` | `orchestrator` |
| `batch_resolve_units(http, criteria, base_url, timeout) -> dict[tuple, str]` | `clients/asana_resolve` | `orchestrator` |
| `emit_reconciliation_metrics(metrics, result, near_miss_data, outcome)` | `metrics` | `handler` |
| `emit_dms_success_timestamp()` | `metrics` | `handler` |

**External service integrations (all via platform SDK):**
- `autom8y-data` insight API: `POST /api/v1/insights/reconciliation/execute` (via `DataInsightClient`)
- `autom8y-asana` resolve API: `POST /v1/resolve/unit` (via `asana_resolve.batch_resolve_units`)
- Slack: Block Kit message posting via `autom8y-slack.SlackClient.send_blocks`
- CloudWatch EMF: stdout emission via `autom8y-telemetry`
- EventBridge bus (optional): `ReconciliationComplete` domain event via `autom8y-events` (try/except optional SDK import)
- AWS Parameters and Secrets Lambda Extension: secrets resolved at runtime via localhost:2773

---

## Key Abstractions

**1. `AnomalyCategory` (`models.py`)** — `IntEnum` with 5 values (`ADS_RUNNING_NO_PAYMENT=1` through `STALE_ACCOUNT=5`). Carries `.label` and `.severity` properties. Used as the primary classification axis throughout the detection->reporting pipeline. Severity ordering (`high/high/medium/medium/low`) drives Slack report section ordering.

**2. `ClientRecord` (`models.py`)** — Frozen dataclass. Represents a single client's reconciliation metrics after parsing and edge-case normalization. 18 fields. The core unit of computation in `rules.py`. Key field: `variance_pct: float | None` (None when original computation invalid after negative-collected adjustment).

**3. `ReconciliationRow` (`clients/models.py`)** — Pydantic `BaseModel` (HTTP boundary). All 18 fields optional with `None` defaults. Intentionally lenient (`extra="ignore"`) to survive upstream schema drift. Maps 1:1 to `ClientRecord` but is the untrusted/unvalidated form.

**4. `Anomaly` (`models.py`)** — Frozen dataclass. Output of a single rule application. Carries `category: AnomalyCategory`, financial amounts, and a human-readable `description` string. Sorted by `(category, office_phone)` before reporting.

**5. `EnrichedAnomaly` (`models.py`)** — Frozen dataclass wrapping `Anomaly` with three optional URL fields: `stripe_dashboard_url`, `invoice_url`, `asana_unit_url`. Produced by `enrich_anomalies()` in the orchestrator.

**6. `ReconciliationResult` (`models.py`)** — Mutable dataclass. Accumulates run state across orchestrator steps. Returned to `handler` and serialized as the Lambda response body. `degraded` property computes to True when `circuit_breaker_state == "open"` or `skipped_stale`.

**7. `Settings` (`config.py`)** — `LambdaServiceSettingsMixin + Autom8yBaseSettings` subclass. Resolved via `@lru_cache` singleton `get_settings()`. Contains asymmetric thresholds (overbilled 10%, underbilled 25%), two staleness settings with different units (`max_staleness_minutes` for anomaly gate, `max_staleness_seconds` for pipeline readiness — intentionally separate per inline comment).

**8. `ReadinessResult` / `ReadinessStatus` (`readiness.py`)** — Frozen dataclass + `StrEnum`. Encapsulates the three-tier freshness gate decision. `ReadinessStatus` values: `PASS`, `WARN`, `FAIL`.

**9. Detect-rules pattern (`rules.py`)** — Five pure functions each with signature `(ClientRecord, threshold) -> Anomaly | None`. `detect_anomalies()` runs all five per record and aggregates. Rules are not mutually exclusive but have documented overlap exclusions: R1 (collected==0) and R2/R3/R4 (collected>0) are mutually exclusive; only R5 (stale payment) is independent.

**10. Two-tier model split** — The codebase deliberately separates Pydantic HTTP-boundary models (`clients/models.py`) from pure-Python domain dataclasses (`models.py`). The boundary is the `parse_client_records()` function in `orchestrator.py`, which transforms `list[ReconciliationRow]` into `list[ClientRecord]`.

---

## Data Flow

**Primary happy-path pipeline:**

```
EventBridge (scheduled)
  -> lambda_handler (handler.py)
     -> asyncio.run(run_reconciliation(period_days))
        -> create_data_client() [autom8y-interop DataInsightClient]
        -> data_client.get_insight("reconciliation", period=...)
             POST /api/v1/insights/reconciliation/execute
             Returns: InsightResult {data: list[dict], data_quality: dict, metadata: dict}
        -> [ReconciliationRow.model_validate(row) for row in data]  # Pydantic boundary
        -> check_pipeline_readiness(staleness_seconds, max_staleness_seconds)
             PASS -> continue
             WARN -> continue (log warning)
             FAIL -> post stale-data Slack alert, return early
        -> parse_client_records(raw_data)  # ReconciliationRow -> ClientRecord
             E4: skip null office_phone
             E5: skip zero/zero accounts
             E9: clamp negative collected to 0
             E10: skip non-finite (NaN/Inf) values
        -> detect_anomalies(records, thresholds)  # rules.py, pure functions
             5 rules applied per record -> list[Anomaly]
        -> compute_near_miss_data(records, ...)  # statistics for metrics
        -> enrich_anomalies(anomalies, records, http, settings)
             batch_resolve_units(http, criteria, asana_url)
               POST /v1/resolve/unit -> AsanaResolveResponse
             Construct stripe_dashboard_url, invoice_url, asana_unit_url
        -> build_report(enriched_anomalies, ...)  # report.py -> list[Block Kit dicts]
        -> slack_client.send_blocks(channel, blocks, text)
        -> return (ReconciliationResult, near_miss_data)
     -> emit_reconciliation_metrics(metrics, result, near_miss_data, outcome)  # best-effort
     -> emit_dms_success_timestamp()  # dead-man's-switch
     -> (optional) publisher.publish(DomainEvent("ReconciliationComplete"))  # fire-and-forget
     -> return {"statusCode": 200, "body": json.dumps({...})}
```

**Degraded paths:**

1. **Circuit breaker open**: `DataServiceUnavailableError` caught in orchestrator -> post circuit-open Slack alert -> return `ReconciliationResult(circuit_breaker_state="open")` with empty near-miss data
2. **Extreme data staleness**: `ReadinessStatus.FAIL` -> post stale-data Slack alert -> `result.skipped_stale = True` -> return early
3. **Metric emission failure**: wrapped in try/except; never blocks Slack report
4. **Asana resolve failure**: returns empty dict; anomalies render without Asana links
5. **Event publish failure**: wrapped in try/except; fire-and-forget

**Configuration merge point:**

`get_settings()` (lru_cache singleton) is the single merge point for all configuration. Environment variables are resolved at first call, with secrets optionally delegated to AWS Parameters and Secrets Lambda Extension at `localhost:2773`.

---

## Knowledge Gaps

- The `docs/` directory at `services/reconcile-spend/docs/` was not read; it may contain ADRs or design notes not reflected here.
- The `just/` directory and `Justfile` were not inspected; they likely define dev/build commands but are not part of the runtime architecture.
- The `docker-compose.override.yml` was not read; it may define local development integration configuration.
- `autom8y-interop`, `autom8y-slack`, `autom8y-telemetry`, `autom8y-config`, and `autom8y-http` SDK internals are not documented here (they are external platform packages).
- The `tests/conftest.py` and test helper files were not read; fixture patterns and test doubles are not captured in this document.
