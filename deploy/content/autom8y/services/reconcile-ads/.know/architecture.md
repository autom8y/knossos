---
domain: architecture
generated_at: "2026-03-16T00:02:18Z"
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

The service is a single-package Python application named `reconcile_ads`, living under `src/reconcile_ads/`. There are no sub-packages. The package contains 11 source modules plus `__init__.py` and `__main__.py`.

**Package root**: `src/reconcile_ads/`

| Module | Purpose | Key Exports |
|--------|---------|-------------|
| `__init__.py` | Package identity, version metadata | `__version__` |
| `__main__.py` | Local CLI entry point (CSV/JSON export) | `main()` |
| `config.py` | Pydantic-settings configuration | `Settings`, `get_settings()`, `clear_settings_cache()` |
| `models.py` | All domain data models and type aliases | `OfferRecord`, `AdRecord`, `AdGroupRecord`, `CampaignRecord`, `CampaignIndex`, `AdSetIndex`, `JoinResult`, `GhostCampaign`, `DecodeFailure`, `BudgetResult`, `Finding`, `FinancialSummary`, `ReconciliationResult`, `AdsSourceMeta`, `AsanaSourceMeta` |
| `errors.py` | Error hierarchy | `ReconcileAdsError`, `FetchError`, `JoinError`, `ReportError` |
| `fetcher.py` | Upstream data fetch (parallel asyncio) | `fetch_all()`, `fetch_ads_tree()`, `fetch_asana_offers()`, `AdsServiceUnavailableError`, `AsanaServiceUnavailableError` |
| `joiner.py` | Two-level join engine, index building | `build_campaign_index()`, `build_ad_set_index()`, `execute_join()`, `detect_ghosts()`, `dedup_weekly_ad_spend()` |
| `rules.py` | Pure verdict rule functions | `apply_all_rules()`, `rule_status_alignment()`, `rule_budget_alignment()`, `rule_delivery_health()` |
| `readiness.py` | Pipeline readiness gate | `check_pipeline_readiness()` |
| `report.py` | Slack Block Kit report builder | `build_report()`, `build_all_clear_report()`, `build_degraded_alert()`, `build_truncation_alert()`, `build_stale_data_alert()` |
| `metrics.py` | CloudWatch metric emission | `emit_reconciliation_metrics()` |
| `handler.py` | AWS Lambda entry point | `lambda_handler()` |
| `orchestrator.py` | Full pipeline coordination | `run_reconciliation()` |

**Import topology observations:**
- `models.py` is the true leaf package — imported by all other modules, imports only from `autom8y_reconciliation.verdict` (external SDK)
- `config.py` is the second-most-imported module, used by `fetcher.py`, `handler.py`, and `orchestrator.py`
- `orchestrator.py` is the hub — it imports from all other internal modules: `fetcher`, `joiner`, `models`, `readiness`, `report`, `rules`
- `errors.py` is defined but **not imported by `orchestrator.py`** — the orchestrator catches service-specific errors from `fetcher.py` directly (`AdsServiceUnavailableError`, `AsanaServiceUnavailableError` which duplicate `FetchError`)
- `joiner.py` inlines campaign/ad-group name decode logic rather than importing from `autom8y-ads` (canonical `NameEncoding` lives in a separate repo not available as a Python dependency)

## Layer Boundaries

The service has three distinct layers:

### Layer 1 — Invocation Surface (Lambda / CLI)
- `src/reconcile_ads/handler.py` — AWS Lambda entry point, decorated with `@instrument_lambda`
- `src/reconcile_ads/__main__.py` — local CLI (`python -m reconcile_ads`) for manual runs with `--csv` or `--json` output

Both invoke `run_reconciliation()` from `orchestrator.py`. Neither contains business logic.

### Layer 2 — Orchestration
- `src/reconcile_ads/orchestrator.py` — `run_reconciliation()` coordinates the full 10-step pipeline. This is the only module that touches all other functional modules.

### Layer 3 — Functional Modules (called by orchestrator only)
- `src/reconcile_ads/fetcher.py` — I/O, external HTTP (reads from autom8y-ads, autom8y-asana via interop SDK)
- `src/reconcile_ads/readiness.py` — pure evaluation, delegates to `autom8y_reconciliation.ReadinessGate`
- `src/reconcile_ads/joiner.py` — pure data transformation (index building, join, ghost detection)
- `src/reconcile_ads/rules.py` — pure verdict functions (no I/O, documented as side-effect-free)
- `src/reconcile_ads/report.py` — Slack Block Kit builder, delegates to `autom8y_reconciliation.report.ReconciliationReportBuilder`
- `src/reconcile_ads/metrics.py` — CloudWatch emission, delegates to `autom8y_reconciliation.metrics.ReconciliationMetrics`

### Layer 4 — Shared Foundations (imported by functional modules)
- `src/reconcile_ads/models.py` — all Pydantic models and type aliases; imported by `fetcher`, `joiner`, `rules`, `readiness`, `report`, `metrics`, `orchestrator`
- `src/reconcile_ads/config.py` — settings singleton; imported by `fetcher`, `handler`, `orchestrator`
- `src/reconcile_ads/errors.py` — error hierarchy; defined but partially superseded by inline errors in `fetcher.py`

**Import direction rule**: Layer 1 → Layer 2 → Layer 3 → Layer 4. Functional modules do NOT import each other (no lateral dependencies within Layer 3). Data flows through the orchestrator as explicit function arguments.

## Entry Points and API Surface

### Production Entry Point
`src/reconcile_ads/handler.py` — `lambda_handler(event, context)`

- Decorated with `@instrument_lambda` from `autom8y_telemetry.aws` (wraps the function with OTel tracing)
- Event source: AWS EventBridge scheduled rule (event dict is largely ignored, schedule-driven)
- Returns: `{"statusCode": 200, "body": JSON}` on success; re-raises on failure to trigger Lambda retry/DLQ
- Calls `asyncio.run(run_reconciliation())` — bridges sync Lambda handler to async pipeline

### Local Development Entry Point
`src/reconcile_ads/__main__.py` — invoked via `uv run python -m reconcile_ads`

CLI flags:
- `--csv DIR` — write findings CSV + result JSON + metadata to directory
- `--json` — print result JSON to stdout
- (no flags) — pretty-print JSON to stdout

Uses `ReconciliationCsvExporter` from `autom8y_reconciliation.csv_export` for structured output.

### Core API Function
`src/reconcile_ads/orchestrator.py` — `async def run_reconciliation() -> ReconciliationResult`

This is the only function that external callers (handler and `__main__`) invoke. It takes no arguments; configuration is loaded internally via `get_settings()`.

### Configuration Surface
`src/reconcile_ads/config.py` — `get_settings() -> Settings` (LRU-cached singleton)

`Settings` inherits from `LambdaServiceSettingsMixin` and `Autom8yBaseSettings`. Key fields:
- `service_api_key: SecretStr` — S2S JWT auth (R-0: `RECONCILE_ADS_SERVICE_KEY`, R-1 fallback: `SERVICE_API_KEY`)
- `meta_account_id: str` — Meta ad account ID (required)
- `ads_service_url`, `asana_service_url` — upstream service endpoints
- `budget_drift_threshold_pct: float` (default 5.0%), `budget_mismatch_threshold_pct: float` (default 20.0%)
- `slack_channel: str` (default `#ads-health`), `slack_max_blocks: int` (default 50)
- `max_staleness_seconds: int` (default 1800)
- `max_campaigns: int` (default 100, max 500)
- `include_transitional: bool` (default False)

`clear_settings_cache()` resets both the LRU cache and `Autom8yBaseSettings.reset_resolver()` — used in tests.

## Key Abstractions

### Data Models (`src/reconcile_ads/models.py`)

**Upstream data models** (raw API responses, flattened):
- `OfferRecord(BaseModel)` — Asana offer row: `gid`, `name`, `section`, `classification`, `office`, `office_phone`, `vertical`, `offer_id`, `weekly_ad_spend`, `platforms`
- `CampaignRecord(BaseModel)` — Meta campaign with decoded name fields (`decoded_phone`, `decoded_vertical_key`, `decoded_business_name`, `decoded_controls_budget`, etc.)
- `AdGroupRecord(BaseModel)` — Meta ad set with decoded `offer_id` and `phone`, contains child `ads: list[AdRecord]`
- `AdRecord(BaseModel)` — Meta ad leaf node

**Index type aliases** (join keys → records):
- `CampaignIndex = dict[tuple[str, str], list[CampaignRecord]]` — key: `(office_phone, vertical_key)`
- `AdSetIndex = dict[tuple[str, str], list[AdGroupRecord]]` — key: `(office_phone, offer_id)`

**Join result types**:
- `JoinResult(BaseModel)` — one Asana offer matched against both indexes; holds `offer`, `campaign_match`, `ad_set_match`, `join_a_key`, `join_b_key`
- `CampaignMatch(BaseModel)` — matched campaigns with `total_daily_budget_cents` and computed `meta_weekly_spend` property
- `AdSetMatch(BaseModel)` — matched ad groups with counts
- `GhostCampaign(BaseModel)` — Meta campaign key with no Asana match; has `daily_budget_dollars` property

**Finding types**:
- `Finding(BaseModel)` — single reconciliation finding; holds three verdict axes (`status_verdict`, `budget_verdict`, `delivery_verdict`) plus unified `verdicts: dict[VerdictAxis, UnifiedVerdict]`; `severity` property uses SDK `severity_key()` for sort ordering
- `FinancialSummary(BaseModel)` — aggregate financial impact (ghost daily budget, missing weekly spend, drift total variance, hollow daily budget)
- `ReconciliationResult(BaseModel)` — full run result returned by `run_reconciliation()`; has `to_dict()` for Lambda JSON response

**Source metadata** (for readiness gate):
- `AdsSourceMeta` (frozen dataclass) — `fetch_timestamp`, `campaign_count`, `has_more_campaigns`, `total_active_campaigns`
- `AsanaSourceMeta` (frozen dataclass) — `fetch_timestamp`, `data_age_seconds`, `staleness_ratio`, `total_count`, `active_returned_count`, `active_total_available`

### Verdict Types (from `autom8y_reconciliation` SDK)
Imported and re-exported by `models.py`:
- `StatusVerdict` enum: `ALIGNED`, `MISSING_CAMPAIGN`, `GHOST_CAMPAIGN`, `DELIVERY_GAP`, `TRANSITIONAL`
- `BudgetVerdict` enum: `MATCHED`, `DRIFT`, `MISMATCH`, `BUDGET_UNAVAILABLE`
- `DeliveryVerdict` enum: `HEALTHY`, `HOLLOW`, `BARREN`
- `VerdictAxis` enum: `STATUS`, `BUDGET`, `DELIVERY`
- `UnifiedVerdict` — wrapper used in the `verdicts` dict on `Finding`

### Name Encoding (`src/reconcile_ads/joiner.py`)
Campaign and ad-group names carry structured data encoded with bullet (`\u2022`) as separator. The joiner inlines decode logic (canonical implementation is in `autom8y-ads` service, not importable):
- `CampaignNameFields(NamedTuple)` — 7 fields: `office_phone`, `business_name`, `vertical_key`, `objective_type`, `optimized_for`, `algo_version`, `controls_budget`
- `AdGroupNameFields(NamedTuple)` — 10 fields: `offer_id`, `office_phone`, `targeting_desc`, `optimization_goal`, `gender`, `age_range`, `language`, `is_dynamic`, `asset_ids`, `question_ids`
- Missing fields are padded with empty strings (E16 guard)

### Error Hierarchy (`src/reconcile_ads/errors.py`)
- `ReconcileAdsError(Exception)` — base; `code: ClassVar[str]`, `http_status: ClassVar[int]`
- `FetchError(ReconcileAdsError)` — `code="FETCH_ERROR"`, `http_status=503`
- `JoinError(ReconcileAdsError)` — `code="JOIN_ERROR"`, `http_status=500`
- `ReportError(ReconcileAdsError)` — `code="REPORT_ERROR"`, `http_status=503`

Note: `fetcher.py` uses its own parallel hierarchy (`AdsServiceUnavailableError`, `AsanaServiceUnavailableError`) that does NOT inherit from `ReconcileAdsError`. These are the errors the orchestrator actually catches. `errors.py` is defined but not used by the orchestrator pipeline.

### SDK Integrations
The service depends heavily on internal SDKs:
- `autom8y_reconciliation` — `ReadinessGate`, `ReconciliationReportBuilder`, `ReconciliationMetrics`, verdict types, `ReconciliationCsvExporter`
- `autom8y_interop` — `AdsCampaignTreeClient`, `AsanaOfferClient` (imported lazily inside `fetcher.py` functions)
- `autom8y_telemetry` — `@instrument_lambda`, `record_side_effect()`
- `autom8y_slack` — `SlackClient` (async context manager)
- `autom8y_events` — `EventPublisher`, `DomainEvent`
- `autom8y_log` — `get_logger()`, `configure_logging()`
- `autom8y_config` — `Autom8yBaseSettings`, `LambdaServiceSettingsMixin`

## Data Flow

### High-Level Pipeline (10 steps, defined in `orchestrator.py` docstring)

```
EventBridge trigger
       |
handler.lambda_handler()  [handler.py]
       |
asyncio.run(run_reconciliation())  [orchestrator.py]
       |
Step 1: fetch_all(settings)  [fetcher.py]
  ├── fetch_ads_tree()  → (ads_tree: dict, ads_meta: AdsSourceMeta)
  └── fetch_asana_offers()  → (offers_data: list[dict], asana_meta: AsanaSourceMeta)
       [parallel via asyncio.gather, error converts to AdsServiceUnavailableError
        or AsanaServiceUnavailableError → degraded alert to Slack → return early]
       |
Step 2: check_pipeline_readiness(ads_meta, asana_meta, max_staleness_seconds)  [readiness.py]
  → ReadinessResult [PASS | WARN | FAIL]
  [FAIL → build_truncation_alert or build_stale_data_alert → Slack → return early]
       |
Step 3: parse offers
  → [OfferRecord.model_validate(row) for row in offers_data]
       |
Step 4: build indexes  [joiner.py]
  ├── build_campaign_index(ads_tree)  → (CampaignIndex, list[DecodeFailure])
  └── build_ad_set_index(ads_tree)   → (AdSetIndex, list[DecodeFailure])
       |
Step 4b: log_multi_campaign_keys(campaign_index)  [joiner.py]
       |
Step 5: execute_join(offers, campaign_index, ad_set_index)  [joiner.py]
  → list[JoinResult]
  detect_ghosts(campaign_index, matched_keys)  → list[GhostCampaign]
  dedup_weekly_ad_spend(offers)  → dict[(phone, vertical), float]
       |
Step 6: apply_all_rules(join_results, ghost_campaigns, deduped_budgets, ...)  [rules.py]
  → (list[Finding], FinancialSummary)
  [sorted by Finding.severity, tie-broken by status_verdict + office_phone]
       |
Step 7: build_report or build_all_clear_report  [report.py]
  → list[dict] (Slack Block Kit blocks, capped at 50)
  → SlackClient.send_blocks()  [best-effort, failure logged but not raised]
       |
Step 8: emit_reconciliation_metrics(result, outcome)  [metrics.py]
  → ReconciliationMetrics.emit_core() + emit_custom() + emit_dms_timestamp()
  → CloudWatch namespace: Autom8y/Reconciliation, dimension: Service=reconcile-ads
       |
Step 9: _publish_complete_event(result)  [orchestrator.py]
  → EventPublisher().publish(DomainEvent(detail_type="CampaignAlignmentComplete"))
  [best-effort, failure debug-logged]
       |
return ReconciliationResult
       |
handler.lambda_handler() → emit_reconciliation_metrics (separately, for failure metric)
                         → return {"statusCode": 200, "body": JSON}
```

### Configuration Flow
Environment variables → `Settings` (Pydantic-settings, env_prefix="") → `get_settings()` (LRU-cached) → consumed by `orchestrator.py`, `fetcher.py`, `handler.py`

Secret resolution in Lambda: AWS Parameters and Secrets Lambda Extension at localhost:2773 (when `_ARN` env vars are present); falls back to direct env var reading for local development.

### Two-Level Join Logic
The join engine creates two indexes from Meta's campaign tree response:
- **Join A** (campaign level): key = `(office_phone, vertical)` from decoded campaign names
- **Join B** (ad-set level): key = `(office_phone, offer_id)` from decoded ad-group names

Each Asana offer is matched against both indexes. Unmatched Meta campaigns become `GhostCampaign` entries.

### Verdict Determination
Three independent rule functions evaluate each `JoinResult`:
1. `rule_status_alignment(offer, campaign_match)` → `StatusVerdict` (ALIGNED / MISSING_CAMPAIGN / GHOST_CAMPAIGN / TRANSITIONAL)
2. `rule_budget_alignment(offer, campaign_match, deduped_budget)` → `BudgetResult` with `BudgetVerdict` (MATCHED / DRIFT / MISMATCH / BUDGET_UNAVAILABLE)
3. `rule_delivery_health(campaign, ad_groups)` → `DeliveryVerdict` (HEALTHY / HOLLOW / BARREN)

Budget rule uses deduped weekly_ad_spend (first-wins per `(phone, vertical)` key — E17) to avoid double-counting when multiple offers share the same unit budget.

## Knowledge Gaps

- The `Justfile` and `docker-compose.override.yml` were not read; they may document local development workflows and Docker configuration.
- The `autom8y_interop` client interface (method signatures, response schemas for `AdsCampaignTreeClient` and `AsanaOfferClient`) is not documented here as those packages are external to this service.
- The EventBridge schedule frequency is not documented in-service (lives in Terraform/IaC).
- The relationship between `errors.py` hierarchy and `fetcher.py`'s inline error classes appears to be an unresolved tension — `errors.py` defines `FetchError` but `orchestrator.py` catches `AdsServiceUnavailableError` / `AsanaServiceUnavailableError` from `fetcher.py` directly.
- ADR-RCA-001 (referenced in `joiner.py` docstring) was not read; it documents the two-level join design rationale.
