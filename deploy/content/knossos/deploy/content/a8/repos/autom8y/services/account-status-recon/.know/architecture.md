---
domain: architecture
generated_at: "2026-03-16T14:32:40Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

**Language**: Python 3.12. Single flat package: `src/account_status_recon/`. Build system: Hatchling. Dependency manager: uv. Runtime: AWS Lambda (container image).

**Module inventory** (11 modules under `src/account_status_recon/`):

| Module | Purpose | Key Types Exported | Role |
|--------|---------|-------------------|------|
| `__init__.py` | Package marker (one-liner) | — | Leaf |
| `__main__.py` | Local CLI runner (`python -m account_status_recon`) | `main()` | Entry point (dev) |
| `config.py` | Settings via pydantic-settings + autom8y_config | `Settings`, `get_settings()`, `clear_settings_cache()` | Leaf (no intra-package imports) |
| `errors.py` | Error hierarchy | `AccountStatusReconError`, `FetchError`, `JoinError`, `ReportError` | Leaf |
| `models.py` | All domain value objects and result types | `SourcePresence`, `BillingData`, `CampaignData`, `ContractData`, `AccountRecord`, `AccountFinding`, `FetchResult`, `ReconciliationResult` | Hub (imported by virtually every other module) |
| `fetcher.py` | Parallel data fetch from 3 upstream APIs | `fetch_all()`, `fetch_billing()`, `fetch_campaigns()`, `fetch_offers()` | Imports `models`, `config` |
| `joiner.py` | Three-way join on `(office_phone, vertical)` | `three_way_join()` | Imports `models`; uses `autom8y_reconciliation.correlator` |
| `rules.py` | Pure verdict logic across 5 axes | `apply_all_rules()`, `rule_status()`, `rule_budget()`, `rule_delivery()`, `rule_billing()`, `rule_three_way()` | Imports `models` via TYPE_CHECKING only |
| `readiness.py` | Pipeline readiness gate | `evaluate_readiness()` | Imports `config`, `models` via TYPE_CHECKING |
| `report.py` | Slack Block Kit report builder | `build_slack_report()`, `render_account_finding()` | Imports `models` via TYPE_CHECKING |
| `metrics.py` | CloudWatch metric emission | `emit_metrics()` | Imports `models` via TYPE_CHECKING |
| `handler.py` | Lambda entry point | `lambda_handler()` | Hub — imports orchestrator, config, models, metrics |
| `orchestrator.py` | Main pipeline coordinator | `run_reconciliation()` | Hub — imports fetcher, joiner, rules, readiness, report, config, models, errors |

**Hub vs leaf classification:**
- **Hub (consumed by many)**: `models.py` is the leaf-dependency anchor — every other module depends on it (directly or via TYPE_CHECKING).
- **Hub (consumes many)**: `orchestrator.py` coordinates all pipeline steps; `handler.py` wires the Lambda entry point.
- **Leaf (pure logic, no intra-package imports at runtime)**: `rules.py`, `report.py`, `metrics.py`, `readiness.py` use `TYPE_CHECKING` guards to avoid circular imports.
- **True leaf**: `config.py`, `errors.py`, `models.py` have no intra-package imports.

---

## Layer Boundaries

The package uses a clear three-layer model despite being a single flat package:

```
Lambda / CLI surface
    handler.py         (Lambda entry — @instrument_lambda, asyncio.run)
    __main__.py        (Local CLI — argparse, asyncio.run)
           |
           v
    orchestrator.py    (Pipeline coordinator — all steps sequenced here)
           |
           v
Domain / pipeline modules (one responsibility each)
    fetcher.py         (I/O: fetch from 3 upstream APIs)
    joiner.py          (pure: three-way join via SDK Correlator)
    rules.py           (pure: 5-axis verdict logic)
    readiness.py       (pure: gate evaluation via SDK ReadinessGate)
    report.py          (pure: Slack Block Kit construction)
    metrics.py         (I/O: CloudWatch emission via SDK)
           |
           v
Foundation (no intra-package imports)
    models.py          (value objects, result types — frozen dataclasses)
    config.py          (Settings — pydantic-settings, lru_cache singleton)
    errors.py          (error hierarchy — ClassVar code + http_status)
```

**Import direction** (strictly top-down):
- `handler.py` → `orchestrator.py` → `fetcher/joiner/rules/readiness/report.py` → `models.py`
- No circular imports. TYPE_CHECKING guards in `rules.py`, `report.py`, `metrics.py`, `readiness.py` prevent runtime circular imports while preserving type hints.
- `orchestrator.py` is the only module that imports across all pipeline steps simultaneously.

**Boundary enforcement pattern:**
```python
# Used in rules.py, report.py, metrics.py, readiness.py:
if TYPE_CHECKING:
    from account_status_recon.models import AccountFinding, AccountRecord, ...
```
This pattern decouples runtime imports while preserving type safety.

**External SDK layer** (autom8y-* packages, not in-repo):
- `autom8y_reconciliation` — Correlator, ReadinessGate, ReconciliationMetrics, ReconciliationReportBuilder, verdict types
- `autom8y_interop` — DataInsightClient, AdsCampaignTreeClient, AsanaOfferClient
- `autom8y_core` / `autom8y_http` — Client, ResilientCoreClient
- `autom8y_slack` — SlackClient
- `autom8y_telemetry` — `@instrument_lambda`, `record_side_effect`, OTEL spans
- `autom8y_events` — DomainEvent, EventPublisher
- `autom8y_log` — `configure_logging()`, `get_logger()`
- `autom8y_config` — Autom8yBaseSettings, LambdaServiceSettingsMixin

---

## Entry Points and API Surface

### Production entry point

**`src/account_status_recon/handler.py`** — Lambda handler registered as:
```
CMD ["account_status_recon.handler.lambda_handler"]
```

Invocation path:
1. EventBridge scheduled trigger fires the Lambda
2. `lambda_handler(event, context)` — decorated with `@instrument_lambda` (auto-configures OTEL)
3. `asyncio.run(run_reconciliation())` — runs async pipeline
4. On success: emits CloudWatch metrics, returns `{"statusCode": 200, "body": {...}}`
5. On failure: emits failure metric, re-raises (Lambda failure)

### Development / local entry point

**`src/account_status_recon/__main__.py`** — invoked via:
```
uv run python -m account_status_recon [--csv DIR] [--json]
```

Options:
- `--json`: Print full `ReconciliationResult.to_dict()` to stdout
- `--csv DIR`: Write findings CSV + result JSON + metadata to `DIR/`
- (no flags): Pretty-print result JSON to stdout

Both entry points call the same `run_reconciliation()` function in `orchestrator.py`.

### No HTTP API surface

This is a Lambda function, not an HTTP service. There are no routes, no REST endpoints. The only external interface is the Lambda invocation boundary (EventBridge → Lambda).

### Key function contracts

| Function | Location | Signature | Notes |
|----------|----------|-----------|-------|
| `lambda_handler` | `handler.py` | `(event: dict \| None, context: Any) -> dict` | `@instrument_lambda`, sync wrapper |
| `run_reconciliation` | `orchestrator.py` | `() -> ReconciliationResult` (async) | Primary pipeline |
| `fetch_all` | `fetcher.py` | `(settings: Settings) -> dict[str, FetchResult]` (async) | Never raises |
| `three_way_join` | `joiner.py` | `(billing_rows, campaign_items, contract_rows) -> list[AccountRecord]` | Pure |
| `apply_all_rules` | `rules.py` | `(records, **thresholds) -> list[AccountFinding]` | Pure |
| `evaluate_readiness` | `readiness.py` | `(fetch_results, settings) -> ReadinessResult` | Pure |
| `build_slack_report` | `report.py` | `(findings, result, warning_banner) -> list[dict]` | Pure |
| `emit_metrics` | `metrics.py` | `(result, outcome) -> None` | Best-effort, never raises |
| `get_settings` | `config.py` | `() -> Settings` | `@lru_cache` singleton |

---

## Key Abstractions

### Core domain types (all in `src/account_status_recon/models.py`)

**`AccountRecord`** (frozen dataclass) — The unified join result for a single `(office_phone, vertical)` account. Contains optional `BillingData`, `CampaignData`, `ContractData`, and required `SourcePresence`. This is the pivot type that rules.py operates on.

**`SourcePresence`** (frozen dataclass) — Three boolean flags (`billing`, `campaign`, `contract`) encoding which of 7 possible presence states an account is in. Properties: `source_count`, `label`, `full`.

**`AccountFinding`** (mutable dataclass) — An account with at least one actionable verdict. Carries `verdicts: dict[VerdictAxis, UnifiedVerdict]` plus display data (financial figures, links). `max_severity` property computes worst severity via `severity_key()`.

**`FetchResult`** (frozen dataclass) — Result of fetching one data source. Either `data + meta` (success) or `error` (failure). `.success` property distinguishes. Orchestrator collects 3 of these; failed ones become degraded sources.

**`ReconciliationResult`** (mutable dataclass) — Full run outcome: counters, findings list, coverage stats, readiness status. Serialized via `.to_dict()` for Lambda response and EventBridge event.

### Verdict system (from `autom8y_reconciliation` SDK)

The 5 verdict axes and their types:

| Axis | Enum | Values |
|------|------|--------|
| `STATUS` | `StatusVerdict` | ALIGNED, GHOST_CAMPAIGN, MISSING_CAMPAIGN, TRANSITIONAL |
| `BUDGET` | `BudgetVerdict` | MATCHED, DRIFT, MISMATCH, BUDGET_UNAVAILABLE |
| `DELIVERY` | `DeliveryVerdict` | HEALTHY, BARREN, HOLLOW |
| `BILLING` | `BillingVerdict` | ADS_RUNNING_NO_PAYMENT, PAYING_NO_ADS, OVERBILLED, UNDERBILLED, STALE_ACCOUNT |
| `THREE_WAY` | `ThreeWayVerdict` | MATCHED, ACTUAL_VS_EXPECTED, BILLED_VS_EXPECTED, ALL_DIVERGENT |

All wrapped in `UnifiedVerdict` (via `UnifiedVerdict.create(axis, value)`). The SDK provides `severity_key()` for cross-axis severity comparison.

### Settings pattern

`Settings` inherits from two mixins:
- `LambdaServiceSettingsMixin` — provides `service_api_key_value`, `auth_base_url`; handles Lambda extension secret resolution
- `Autom8yBaseSettings` — pydantic-settings base with SecretStr support

`get_settings()` is an `@lru_cache` singleton. `clear_settings_cache()` resets it (used in tests).

### Error hierarchy

```
AccountStatusReconError (base, code="ACCOUNT_STATUS_RECON_ERROR", http_status=500)
  FetchError (code="FETCH_ERROR", http_status=503)
  JoinError (code="JOIN_ERROR", http_status=500)
  ReportError (code="REPORT_ERROR", http_status=503)
```

Note: `FetchError` is defined but the fetcher never raises it — it catches all exceptions and returns `FetchResult(error=exc)`. `ReportError` is raised by `_safe_slack_post` in orchestrator.py on Slack post failure.

### SDK Correlator pattern (in `joiner.py`)

`Correlator.build_index(rows, key_fn)` — builds `{CompositeKey: list[row]}` index.
`Correlator.dedup(rows, key_fn)` — first-wins deduplication by composite key.
`CompositeKey = tuple[str, str]` — the `(office_phone, vertical)` join key.

Campaign name encoding: `"<phone>•<business_name>•<vertical>"` — bullet separator `\u2022` is used to decode `(phone, vertical)` from the campaign's `raw_name`.

---

## Data Flow

### Primary pipeline (triggered by EventBridge on schedule)

```
EventBridge scheduled trigger
    → lambda_handler (handler.py)
    → asyncio.run(run_reconciliation())

    Step 1: fetch_all(settings)                    [fetcher.py]
        asyncio.gather(fetch_billing, fetch_campaigns, fetch_offers)
        Each fetcher: Settings → SDK client → API call → FetchResult
        Sources:
          billing   → DataInsightClient.get_insight("reconciliation", period="7d")
          campaigns → AdsCampaignTreeClient.get_active_tree(account_id, max_campaigns)
          offers    → AsanaOfferClient.query_rows("offer", classification=active/activating)
        Returns: dict[str, FetchResult] (always 3 entries, each may be success/error)

    Step 2: evaluate_readiness(successful_sources, settings)   [readiness.py]
        Builds SourceMetadata per present source with staleness + completeness checks
        SDK ReadinessGate.evaluate(sources) → ReadinessResult
        worst-wins: if any source FAIL → abort with Slack alert
        if WARN → continue with warning_banner

    Step 3: three_way_join(billing_rows, campaign_items, contract_rows)  [joiner.py]
        Phase 1: Build indexes per source via Correlator.build_index()
        Phase 2: Compute key union (all unique (office_phone, vertical) keys)
        Phase 3: Iterate sorted keys → AccountRecord per key
        Returns: list[AccountRecord] (one per unique account key)

    Step 4: apply_all_rules(records, **thresholds)  [rules.py]
        For each AccountRecord, evaluate applicable axes:
          STATUS axis   → needs campaign OR contract
          BUDGET axis   → needs campaign AND contract
          DELIVERY axis → needs campaign only
          BILLING axis  → needs billing only (may produce multiple verdicts)
          THREE_WAY axis → needs billing AND contract
        Filter: only accounts with actionable verdicts become AccountFinding
        Sort: most severe first, then alpha by office_phone
        Returns: list[AccountFinding]

    Step 5: build_slack_report(findings, result, warning_banner)  [report.py]
        Groups findings by severity tier (CRITICAL → HIGH → MEDIUM → LOW)
        SDK ReconciliationReportBuilder → Slack Block Kit blocks
        50-block limit (FR-21)
        All-clear variant if zero findings (FR-28)
        SlackClient.send_blocks(channel, blocks, text)

    Step 6: emit_metrics(result, outcome)  [metrics.py, via handler.py]
        SDK ReconciliationMetrics.emit_core() → core CW metrics
        SDK ReconciliationMetrics.emit_custom() → source coverage + verdict counts
        emit_dms_timestamp() → LastSuccessTimestamp dead-man's-switch

    Step 7: _publish_complete_event(result)  [orchestrator.py]
        EventPublisher().publish(DomainEvent("AccountStatusComplete", ...))
        Best-effort, exceptions swallowed
```

### Configuration resolution path

```
Environment variables / AWS SSM (via Lambda extension at localhost:2773)
    → pydantic-settings validation in Settings.__init__()
    → AliasChoices: ACCOUNT_STATUS_RECON_SERVICE_KEY then SERVICE_API_KEY
    → @lru_cache get_settings() singleton
    → passed as `settings` arg to each fetcher
```

Secret resolution: The Lambda extension (AWS Parameters and Secrets Lambda Extension v12) intercepts env var lookups and resolves SSM Parameter Store paths at runtime. No deploy-time secret injection needed.

### Failure modes and degraded paths

- **Single source fails**: Fetch exception caught → `FetchResult(error=exc)` → `degraded_sources` list populated → continues with remaining sources → warning_banner in report
- **All sources fail (EC-6)**: No successful sources → immediate Slack alert + return (no join/rules/report)
- **Readiness gate FAIL (FR-15)**: Staleness or completeness check fails → Slack abort alert + return
- **Slack post fails**: `ReportError` raised → Lambda fails (findings logged to CloudWatch via EC-20 pattern)
- **Metrics emission fails**: Exception swallowed, logged at DEBUG level
- **EventBridge publish fails**: Exception swallowed, logged at DEBUG level

### Merge points for thresholds

All verdict thresholds originate from `Settings` (env vars → pydantic validation), pass through `orchestrator.run_reconciliation()` as keyword arguments to `apply_all_rules()`, and are forwarded per-rule to `rule_budget()`, `rule_three_way()`, `rule_billing()`.

---

## Knowledge Gaps

- The `just/` subdirectory structure (service-local Justfile modules: `dev.just`, `test.just`, `fmt.just`, `docker.just`, `ci.just`, `lambda.just`) was not read — CI/CD commands are not documented in detail. This is out of scope for architecture but relevant for operations.
- The `autom8y_reconciliation` SDK internals (Correlator, ReadinessGate, ReconciliationReportBuilder, verdict types, metrics) were not read (external package). The interface contracts are documented from usage, not source.
- No Terraform / IaC files are present in this directory — infrastructure is co-located in a separate devops directory referenced via shared Justfile imports.
- Tests were not read for this architecture document (covered by `test-coverage` domain).
