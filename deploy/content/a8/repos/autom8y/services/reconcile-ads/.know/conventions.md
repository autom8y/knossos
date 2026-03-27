---
domain: conventions
generated_at: "2026-03-16T00:02:18Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.93
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

### Custom Exception Hierarchy

The service defines a formal exception hierarchy in `src/reconcile_ads/errors.py`. All service errors inherit from a single base class:

```
ReconcileAdsError(Exception)
  ├── FetchError       (http_status=503, transient)
  ├── JoinError        (http_status=500, permanent)
  └── ReportError      (http_status=503, transient)
```

Every exception class carries two `ClassVar` fields:
- `code: ClassVar[str]` — a SCREAMING_SNAKE_CASE string code (e.g., `"FETCH_ERROR"`)
- `http_status: ClassVar[int]` — maps the error to an HTTP status code

`FetchError.__init__` accepts structured args (`source_name`, `method`, `time_remaining`) and formats its own message string. Other errors rely on default `Exception` message passing.

### Boundary-Level Errors in Fetcher

`src/reconcile_ads/fetcher.py` defines two additional error classes — `AdsServiceUnavailableError` and `AsanaServiceUnavailableError` — that are NOT in `errors.py`. These are fetcher-local errors that mirror the interop layer's errors. They follow the same `(service_name, method, time_remaining)` constructor pattern as `FetchError`.

The fetcher uses a three-clause `try/except` pattern at service boundaries:
1. Catch the interop-layer-specific error, re-raise as the local service error (`# noqa: B904`)
2. Re-raise the local error if it was already the right type (avoids double-wrapping)
3. Catch `Exception` as a fallback and raise the local unavailable error

### Error Propagation in Orchestrator

In `src/reconcile_ads/orchestrator.py`, errors from the fetcher are caught with explicit `except AdsServiceUnavailableError` / `except AsanaServiceUnavailableError` blocks. On catch, the orchestrator:
- Sets `result.degraded = True` and `result.degraded_source = exc.service_name`
- Posts a Slack alert via `_safe_slack_post`
- Returns early (does not re-raise — degraded run is a controlled exit)

Critical non-blocking operations (Slack posting, metric emission, EventBridge publishing) each wrap their logic in `try/except Exception` with `log.exception` or `log.debug`, never propagating. This is the "fail-open / best-effort" pattern, documented inline (e.g., `# E15: Lambda should still return success`).

In `handler.py`, uncaught exceptions from `run_reconciliation` are caught, logged with `log.exception`, and re-raised to trigger Lambda retry/DLQ.

### Logging at Boundaries

Structured logging uses `log.warning(...)` for degraded conditions and `log.exception(...)` for unexpected failures. Log messages use snake_case event names as the first positional arg (e.g., `"ads_service_unavailable"`, `"slack_post_failed"`). Contextual fields are passed as keyword arguments.

### Error Code Comments

Several edge cases are annotated with inline codes like `# E10:`, `# E13:`, `# E15:`, etc. These are PRD edge-case references, not a runtime error code system.

---

## File Organization

### Package Layout

```
src/reconcile_ads/
  __init__.py       — version export only (importlib.metadata pattern)
  __main__.py       — local CLI entry point (argparse, CSV/JSON export)
  config.py         — Settings class + get_settings() + clear_settings_cache()
  errors.py         — Service error hierarchy (custom exceptions)
  models.py         — All Pydantic models, dataclasses, and type aliases
  fetcher.py        — Upstream data fetching (two async functions + fetch_all)
  joiner.py         — Two-level join engine (index building + join execution)
  readiness.py      — Pipeline readiness gate (SDK delegation)
  rules.py          — Pure verdict rule functions (no I/O)
  report.py         — Slack Block Kit report builder (SDK delegation)
  metrics.py        — CloudWatch metric emission (SDK delegation)
  handler.py        — Lambda entry point (@instrument_lambda decorator)
  orchestrator.py   — Main pipeline orchestration (async, calls all other modules)
```

### File Responsibility Pattern

Each file has a focused, single-domain responsibility documented in its module docstring. The docstring for each file states what it does in 1-3 sentences. No file mixes concerns: models never import from handler; rules never import from fetcher.

### Where Things Live

- **All Pydantic models, dataclasses, and type aliases**: `models.py` (single file, subdivided by `# ---` section banners)
- **Settings**: `config.py` — one `Settings` class, one `@lru_cache` accessor `get_settings()`, one `clear_settings_cache()` for test isolation
- **Constants**: Module-level in the file where used (e.g., `SEPARATOR`, `KNOWN_VERTICALS`, `_ACTIVE_CLASSIFICATIONS` in `joiner.py`; `MAX_BLOCKS`, `_REPORT_CONFIG` in `report.py`; `NAMESPACE` in `metrics.py`)
- **Private helpers**: Underscore-prefixed module-level functions in the file that uses them (e.g., `_decode_campaign_name`, `_build_finding_from_offer`, `_safe_slack_post`)
- **`__init__.py`**: Only exports package version. No business logic re-exports.

### Section Banners in models.py

`models.py` uses `# ---------------------------------------------------------------------------` separator comments with descriptive labels to divide logical groups:
- `# Upstream data models`
- `# Index type aliases`
- `# Join result types`
- `# Verdict and finding types`
- `# Financial summary`
- `# Reconciliation result`
- `# Source metadata (for readiness gate)`

### Internal Package Usage

`TYPE_CHECKING` guards are used consistently for type-only imports that would cause circular dependencies or unnecessary import overhead at runtime (e.g., in `fetcher.py` and `readiness.py` importing from `models`). The pattern is always:

```python
from __future__ import annotations
...
if TYPE_CHECKING:
    from reconcile_ads.models import AdsSourceMeta
```

### Re-export Pattern

`readiness.py` and `models.py` use `__all__` for re-exports to maintain backward compatibility as the SDK integration evolved. `readiness.py` re-exports `ReadinessStatus` from the SDK. `models.py` re-exports verdict types from `autom8y_reconciliation`.

### Generated / Vendor Code

No generated files. Name-encoding decode logic from `autom8y-ads` (separate repo) is manually inlined in `joiner.py` with a comment explaining why (cannot import as dependency).

---

## Domain-Specific Idioms

### Bullet-Encoded Names

Meta campaign and ad group names carry structured data encoded with a bullet separator (`"\u2022"`). The joiner module replicates the decode logic from `autom8y-ads` since that repo is not an importable dependency. `CampaignNameFields` and `AdGroupNameFields` are `NamedTuple`s. Missing fields are padded with empty strings (error code `E16`).

### Two-Level Join

The core algorithm is a two-level join against two in-memory dict indexes:
- **Join A**: `(office_phone, vertical)` → list of `CampaignRecord` (campaign-level)
- **Join B**: `(office_phone, offer_id)` → list of `AdGroupRecord` (ad-set-level)

Index type aliases are defined in `models.py`: `CampaignIndex = dict[tuple[str, str], list[CampaignRecord]]` and `AdSetIndex = dict[tuple[str, str], list[AdGroupRecord]]`. Keys are always `tuple[str, str]`.

### Ghost Campaign Detection

Ghost campaigns are Meta campaigns with no matching Asana offer. They are detected by set-difference: all campaign index keys minus the set of keys that participated in a successful join.

### Deduped Budget

`weekly_ad_spend` cascades from the Asana Unit level and is shared across sibling offers in the same `(phone, vertical)` group. The join deduplicates it using `dedup_weekly_ad_spend()` (first-seen wins, `E17`).

### SDK Delegation Pattern

Several modules are thin wrappers around SDK classes. `readiness.py` delegates entirely to `ReadinessGate`. `metrics.py` delegates to `ReconciliationMetrics`. `report.py` delegates to `ReconciliationReportBuilder`. Service-specific logic (custom metrics, finding renderers) is passed as arguments to SDK entry points.

### Best-Effort / Fail-Open Operations

Non-blocking side effects (Slack, EventBridge, CloudWatch metrics) use a consistent pattern: `try/except Exception` with debug or exception logging, never propagating. Annotated with `E15` (Slack) or similar edge-case codes. The `_safe_slack_post` helper in `orchestrator.py` encapsulates the Slack fail-open pattern.

### Edge Case Codes

Inline comments like `# E3:`, `# E10:`, `# E13:` reference PRD Appendix items. They are documentation annotations, not a runtime code system. They appear at the point of the edge-case handling decision.

### Verdict Types from SDK

Verdict enums (`StatusVerdict`, `BudgetVerdict`, `DeliveryVerdict`, `UnifiedVerdict`, `VerdictAxis`) come from `autom8y_reconciliation.verdict`. The service never defines its own verdict types. `UnifiedVerdict.create(axis, verdict)` is the factory pattern for populating the `verdicts` dict in `Finding`.

### `Finding.severity` Dual-Path

`Finding.severity` uses the SDK `severity_key()` when the unified `verdicts` dict is populated, but falls back to a legacy `severity_map` keyed on `StatusVerdict` values. This dual-path is a migration artifact explicitly labeled "backward compatibility during migration."

### OpenTelemetry Span Discipline

The orchestrator wraps each pipeline step in `with _tracer.start_as_current_span(...)` and sets span attributes using constants from `autom8y_telemetry.conventions`. The tracer is a module-level `_tracer = trace.get_tracer("autom8y.reconciliation")`. Side effects (Slack, EventBridge, CloudWatch) are recorded with `record_side_effect(span, system=..., operation=..., target=..., payload=..., status=...)`.

---

## Naming Patterns

### File Names

- Files named after their single responsibility: `fetcher.py`, `joiner.py`, `rules.py`, `report.py`, `metrics.py`, `readiness.py`
- Standard entry-point names: `handler.py` (Lambda), `__main__.py` (CLI), `config.py`, `models.py`, `errors.py`
- Package name uses underscores: `reconcile_ads` (Python import name), hyphens for distribution name: `reconcile-ads`

### Class Names

- Pydantic models: `PascalCase` noun phrases — `OfferRecord`, `AdRecord`, `CampaignRecord`, `AdGroupRecord`, `JoinResult`, `Finding`, `ReconciliationResult`
- Source metadata dataclasses: `AdsSourceMeta`, `AsanaSourceMeta`
- Error classes: `ReconcileAdsError`, `FetchError`, `JoinError`, `ReportError` (always suffixed with `Error`)
- Settings class: simply `Settings` (one per service)
- NamedTuples: `CampaignNameFields`, `AdGroupNameFields`

### Function Names

- Public async functions: `fetch_ads_tree`, `fetch_asana_offers`, `fetch_all`, `run_reconciliation`
- Public sync functions: `build_campaign_index`, `build_ad_set_index`, `execute_join`, `detect_ghosts`, `dedup_weekly_ad_spend`, `apply_all_rules`, `check_pipeline_readiness`
- Rule functions: `rule_status_alignment`, `rule_budget_alignment`, `rule_delivery_health` (always prefixed `rule_`)
- Private helpers: `_decode_campaign_name`, `_decode_ad_group_name`, `_build_summary_lines`, `_render_finding_block`, `_safe_slack_post`, `_publish_complete_event` (always prefixed `_`)
- Factory/accessor: `get_settings()` (singleton accessor pattern), `clear_settings_cache()`

### Variable Names

- Type aliases: `CampaignIndex`, `AdSetIndex` (PascalCase, described in models.py)
- Constants: `SCREAMING_SNAKE_CASE` — `SEPARATOR`, `KNOWN_VERTICALS`, `NAMESPACE`, `MAX_BLOCKS`, `FINDING_CSV_COLUMNS`
- Private constants: `_ACTIVE_CLASSIFICATIONS`, `_TRANSITIONAL_CLASSIFICATIONS`, `_INACTIVE_CLASSIFICATIONS`, `_REPORT_CONFIG`, `_tracer`
- Loop variables mirror domain terms: `phone`, `vertical`, `offer`, `campaign`, `key`, `findings`

### Logging

Logger always named from module: `log = get_logger(__name__)`. Log event names are lowercase `snake_case` strings as the first positional arg: `"reconciliation_started"`, `"ads_tree_fetched"`, `"offer_skipped_null_phone"`.

### Span Names

OTel spans use dot-notation namespaced to the tracer: `"reconciliation.fetch"`, `"reconciliation.readiness_gate"`, `"reconciliation.join"`, `"reconciliation.verdicts"`, `"reconciliation.verdicts.status_rule"`, `"reconciliation.report"`, `"reconciliation.metrics"`, `"reconciliation.event_publish"`.

### Test Fixtures

Factory functions in `conftest.py` use `make_` prefix: `make_offer()`, `make_campaign_raw_name()`, `make_tree()`, `make_default_campaign_tree()`. Pytest fixtures use descriptive nouns: `default_offer`, `default_tree`, `matched_pair`. The autouse settings fixture uses a leading underscore: `_clear_settings`.

### Type Annotations

All source files start with `from __future__ import annotations`. Strict mypy is configured. Type aliases use `PascalCase`. `ClassVar` fields use `TYPE_CHECKING` guard for import. `SecretStr` is used for all secret fields in settings. `tuple[str, str]` (not `Tuple`) is the canonical join-key type.

---

## Knowledge Gaps

1. **`report.py` `_render_finding_block`**: The function is defined but not imported into the SDK builder call. It is passed as a `renderer` callback to `ReportSection`. The mechanism by which the SDK calls this callback is not visible (lives in `autom8y_reconciliation.report`). Behavior of the SDK's `build_report` and block-limit enforcement is opaque without reading the SDK.

2. **Interop layer error classes**: `AdsServiceUnavailableError` and `AsanaServiceUnavailableError` are imported from `autom8y_interop.ads.errors` and `autom8y_interop.asana.errors`. Their exact shape is not visible. The fetcher wraps them, but the original error attributes (e.g., whether they always have `service_name`, `method`, `time_remaining`) must be assumed from usage.

3. **`autom8y_reconciliation` SDK surface**: This service relies heavily on SDK primitives (`ReadinessGate`, `ReconciliationMetrics`, `ReconciliationReportBuilder`, `UnifiedVerdict`, `severity_key`). The conventions for extending SDK behavior are not documented in this service.

4. **`__main__.py` `_finding_to_row`**: This function contains a complex inline ternary for enum serialization (line 53-55) that is visibly awkward. Whether this is a known debt or an oversight is unclear.

5. **`fetcher.py` imports inside function body**: `autom8y_core`, `autom8y_http`, and `autom8y_interop` are imported inside the function body of `fetch_ads_tree` and `fetch_asana_offers`, with `ImportError` caught and converted to `ServiceUnavailableError`. Whether this lazy-import pattern is a project convention or a workaround for optional dependencies is not explained by any comment.
