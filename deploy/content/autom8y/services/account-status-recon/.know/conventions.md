---
domain: conventions
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

# Codebase Conventions

**Primary Language**: Python 3.12
**Build system**: Hatchling, `uv` as runtime
**Source root**: `src/account_status_recon/`
**Static analysis**: Ruff (inherits `../../ruff.toml`), mypy strict

---

## Error Handling Style

The service uses a **three-layer custom exception hierarchy** rooted at `AccountStatusReconError`. All domain errors live in `src/account_status_recon/errors.py`.

### Error Creation Pattern

Every error class carries two `ClassVar` annotations on the class body — `code: ClassVar[str]` (a screaming-snake-case machine-readable tag) and `http_status: ClassVar[int]`. This is explicitly labelled as "the established service error hierarchy pattern" in the module docstring with a pointer to `.know/conventions.md`, signalling it is a platform-wide convention, not local invention.

```
AccountStatusReconError (base, code="ACCOUNT_STATUS_RECON_ERROR", http_status=500)
├── FetchError            (code="FETCH_ERROR",  http_status=503)
├── JoinError             (code="JOIN_ERROR",   http_status=500)
└── ReportError           (code="REPORT_ERROR", http_status=503)
```

Transient errors (`FetchError`, `ReportError`) accept structured keyword arguments (`source_name`, `method`, `time_remaining`) and compose a human-readable message in `__init__` before calling `super().__init__(message)`. Permanent errors (`JoinError`) use no custom `__init__` — they rely on the base message.

### Error Wrapping Conventions

Chaining is done with `raise SomeError(...) from exc` (see `src/account_status_recon/orchestrator.py:381-384`). No custom wrap helpers exist. The `from exc` clause is used wherever the original exception is available.

### Error Propagation Style

There are three distinct propagation tiers observed in this service:

1. **Swallow and convert to result** — the fetcher layer (`src/account_status_recon/fetcher.py`) catches every `Exception` inside each `fetch_*` coroutine and returns `FetchResult(source_name=..., error=exc)`. The caller inspects `.success` to distinguish success from failure. Errors never escape the fetcher boundary.

2. **Swallow silently, log at debug** — best-effort operations (metric emission, EventBridge publish, metrics in handler) catch `Exception` and `log.debug(..., exc_info=True)`. These are explicitly annotated "best-effort: exceptions logged, never propagated" in docstrings (`src/account_status_recon/metrics.py:26`, `src/account_status_recon/orchestrator.py:343`).

3. **Log and re-raise** — the Lambda handler's outermost `try/except` (`src/account_status_recon/handler.py:69-76`) catches `Exception`, calls `log.exception(...)`, then `raise` — preserving the original traceback for Lambda's error reporting.

### Error Handling at Boundaries

- **Slack posting**: `_safe_slack_post()` in `src/account_status_recon/orchestrator.py:347-384` catches exceptions, records a `FAILED` side-effect telemetry event, then raises `ReportError(...) from exc`.
- **Lambda entry point**: no swallowing; `log.exception("reconciliation_failed", error=str(e))` then re-raise.
- **Inline `if err`**: not used. Pattern is `if not fr.success:` on `FetchResult`, avoiding direct exception inspection.

### Error Checking Conventions

- Guard conditions use early `return` or `return None` (e.g., `src/account_status_recon/rules.py:58-59`, `292-295`).
- `assert` is used solely for type narrowing after exhaustive conditional checks (e.g., `src/account_status_recon/rules.py:76-77`); not for runtime validation.
- No must-style panic helpers; no `typing.Never` or `typing.assert_never`.

---

## File Organization

### Module-per-Concern Layout

The single package `account_status_recon` in `src/account_status_recon/` is organized strictly one-module-per-concern:

| File | Contents |
|------|----------|
| `__init__.py` | One-line package docstring; no exports |
| `__main__.py` | CLI entry point (`python -m account_status_recon`) |
| `config.py` | `Settings` (pydantic-settings) + `get_settings()` + `clear_settings_cache()` |
| `models.py` | All domain dataclasses (value objects and result containers) |
| `errors.py` | Full exception hierarchy |
| `fetcher.py` | Parallel data fetch layer (3 fetch functions + `fetch_all`) |
| `joiner.py` | Three-way join logic + key-extraction functions + parse helpers |
| `readiness.py` | Pipeline readiness evaluation |
| `rules.py` | Verdict rules (pure functions, no I/O) |
| `report.py` | Slack report builder |
| `metrics.py` | CloudWatch metric emission |
| `orchestrator.py` | Main pipeline coordinator |
| `handler.py` | Lambda handler entry point |

There is no `utils.py`, no `helpers.py`, no catch-all file. Every concern has an explicit, named home.

### Internal Function Prefix Convention

Private functions that are not part of a module's public surface use a single-underscore prefix: `_aggregate_result_counts`, `_publish_complete_event`, `_safe_slack_post`, `_build_all_failed_alert`, `_build_readiness_abort_alert` in `orchestrator.py`; `_parse_billing_row`, `_parse_campaign_items`, `_parse_contract_row` in `joiner.py`; `_aggregate_verdict_counts` in `metrics.py`; `_severity_emoji` in `report.py`.

### No Barrel `__init__.py` Exports

`__init__.py` carries only a one-line docstring. No symbols are re-exported from the package root. Consumers must import from the specific submodule.

### Module-Level Singletons

Each module that uses logging declares `log = get_logger(__name__)` at module top-scope (not inside functions). The OTel tracer `_tracer = trace.get_tracer("autom8y.reconciliation")` is likewise a private module-level singleton (underscore-prefixed) in `handler.py` and `orchestrator.py`.

### `from __future__ import annotations`

Present in every source module (absent only in `handler.py` and `__init__.py`). Its purpose is to enable PEP 563 string-based annotation evaluation — required for forward references, and enforced by the `TCH` ruff rule (type-checking imports behind `TYPE_CHECKING` guards).

### `TYPE_CHECKING` Guard Pattern

Heavy use of `if TYPE_CHECKING:` blocks to keep imports that are only needed for type annotations out of the runtime import graph:

- `errors.py` guards `ClassVar` import (used only for annotations)
- `fetcher.py` guards `Settings` import (function argument type only)
- `readiness.py` guards `Settings` and `FetchResult`
- `rules.py` guards all four model types
- `report.py` guards `AccountFinding`
- `metrics.py` guards `ReconciliationResult`

### Test Layout

```
tests/
  __init__.py
  conftest.py            — shared factories and fixtures
  test_rules.py
  test_joiner.py
  test_orchestrator.py
  test_fetcher.py
  test_instrumentation.py
  qa/
    __init__.py
    test_edge_cases_adversarial.py
    test_qa_adversary.py
```

Tests mirror the `src/` module structure: one test file per source module. The `qa/` subdirectory holds adversarial and edge-case tests separated from unit tests.

---

## Domain-Specific Idioms

### FetchResult Envelope (Fail-Open Fetch)

Every data fetch returns `FetchResult` — a frozen dataclass carrying `source_name`, `data`, `meta`, `error`, and a computed `.success` property. This envelope pattern prevents fetch errors from propagating as exceptions; the orchestrator inspects `.success` to implement graceful degradation. This is a project-specific pattern extending the `autom8y_reconciliation` SDK.

```python
@dataclass(frozen=True)
class FetchResult:
    source_name: str
    data: Any = None
    meta: Any = None
    error: Exception | None = None

    @property
    def success(self) -> bool:
        return self.error is None and self.data is not None
```

### `SourcePresence` as 3-bit State Carrier

`SourcePresence(billing: bool, campaign: bool, contract: bool)` encodes the 2^3-1=7 possible source states for a given account key. The class docstring explicitly enumerates all 7 states. Downstream code always checks `record.presence.billing`, `record.presence.campaign`, etc. rather than checking for `None` on the data fields directly.

### Frozen Dataclasses for Value Objects, Mutable Dataclasses for Result Containers

The convention is crisp:
- **Frozen**: `SourcePresence`, `BillingData`, `CampaignData`, `ContractData`, `AccountRecord`, `FetchResult` — immutable after construction; used as keys or identity-stable values.
- **Mutable**: `AccountFinding`, `ReconciliationResult` — containers built up incrementally during pipeline execution.

### `@lru_cache` Settings Singleton

`get_settings()` in `src/account_status_recon/config.py` is decorated `@lru_cache`. A paired `clear_settings_cache()` function exists exclusively for tests to reset the cache and the underlying resolver. This is the canonical pattern for settings access — never instantiate `Settings()` directly outside tests.

### `asyncio.gather(return_exceptions=True)` for Fan-Out

All three data fetches run concurrently via `asyncio.gather(..., return_exceptions=True)` in `fetch_all()`. Unexpected exceptions are caught and converted to `FetchResult` errors; the orchestrator always receives a dict of exactly three entries regardless of failures.

### Verdict Rules as Pure Functions

All rule functions (`rule_status`, `rule_budget`, `rule_delivery`, `rule_billing`, `rule_three_way`) are pure: no I/O, no side effects, no logging (except one `log.warning` guard in `rule_three_way` for zero-expected-spend). They receive data objects and return `UnifiedVerdict | None` (or `list[UnifiedVerdict]` for billing). The module docstring flags this explicitly: "pure functions, no I/O."

### Inline Imports to Break Circular Dependencies

`from autom8y_reconciliation.verdict import Severity` and `from account_status_recon.rules import rule_three_way` appear inside function bodies in `orchestrator.py:289, 301`. This is a deliberate pattern to avoid circular imports at module load time, not a code quality shortcut.

### Section-Comment Delimiters

Long modules use `# ===...===` banner comments to separate logical sections:

```python
# =============================================================================
# STATUS axis rules (mirrors reconcile-ads/rules.py:rule_status_alignment)
# =============================================================================
```

This convention appears in `rules.py`, `config.py`, `pyproject.toml`, and `conftest.py`. Comments also include cross-references to sibling services (e.g., "Pattern: reconcile-spend/orchestrator.py:299-315").

### FR-/EC-/ADR-/NFR- Reference Tags in Comments

Inline comments and docstrings reference spec requirements by ID:
- `FR-N` — functional requirement
- `EC-N` — edge case
- `NFR-N` — non-functional requirement
- `ADR-ASR-N` — architecture decision record

These tags appear throughout every module (e.g., `# EC-11: Guard against NaN/Inf`, `# FR-9: Zero or None expected_spend -> skip`, `NFR-3: All settings with defaults and validation`). This is the project's traceability convention.

### Test Factory Functions (not Fixtures) as Primary Test Data Builders

`conftest.py` provides both free-standing factory functions (`make_billing`, `make_campaign`, `make_contract`, `make_account_record`, `make_billing_row`, `make_campaign_item`, `make_contract_row`) and thin `@pytest.fixture` wrappers that call the factories with default arguments. The pattern enables both fixture injection and direct parametrized calls in the same test file.

---

## Naming Patterns

### Module Names: Snake-Case, One-Concern Nouns

`config`, `models`, `errors`, `fetcher`, `joiner`, `readiness`, `rules`, `report`, `metrics`, `orchestrator`, `handler`. All singular nouns (no plurals, no verbs). The package name itself is `account_status_recon` (snake-case of the service directory `account-status-recon`).

### Class Names: PascalCase Domain Nouns

- **Data models**: `BillingData`, `CampaignData`, `ContractData`, `AccountRecord`, `AccountFinding`, `SourcePresence`, `FetchResult`, `ReconciliationResult`
- **Settings**: `Settings` (not `ServiceSettings` or `AccountStatusReconSettings`)
- **Error classes**: `AccountStatusReconError`, `FetchError`, `JoinError`, `ReportError`

No `*Options`, `*Config` (except `ReportConfig` from the SDK), no `*Builder` suffix — those come from the SDK layer (`ReconciliationReportBuilder`, `ReportConfig`, `ReportSection`).

### Function Names: Snake-Case Verb Phrases

- Entry point: `lambda_handler`
- Pipeline function: `run_reconciliation`
- Fetch functions: `fetch_billing`, `fetch_campaigns`, `fetch_offers`, `fetch_all`
- Join: `three_way_join`
- Rule functions: `rule_status`, `rule_budget`, `rule_delivery`, `rule_billing`, `rule_three_way`, `apply_all_rules`
- Report builder: `build_slack_report`, `render_account_finding`
- Metrics: `emit_metrics`
- Readiness: `evaluate_readiness`

Public functions use `verb_noun` or `verb_adjective_noun`. Private helpers are prefixed with `_` (e.g., `_parse_billing_row`, `_aggregate_result_counts`).

### Key-Extraction Functions: `{source}_key_fn`

Three functions in `joiner.py` follow the `{source}_key_fn` naming: `billing_key_fn`, `campaign_key_fn`, `contract_key_fn`. These are not private (no underscore) because they are passed as callbacks to `Correlator.build_index`.

### Constant Names: SCREAMING_SNAKE_CASE, Module-Level, Prefixed by Role

- Module-private classification sets: `_ACTIVE_CLASSIFICATIONS`, `_TRANSITIONAL_CLASSIFICATIONS` (underscore prefix)
- CSV column list: `FINDING_CSV_COLUMNS` (no underscore)

### Acronym Conventions

- `ID` not `Id` (consistent with Python conventions; observed in `meta_account_id`, `offer_id`, `customer_id`)
- `URL` not `Url` (`asana_offer_url`, `autom8y_data_url`, `autom8y_ads_url`, `autom8y_asana_url`)
- `pct` not `percent` for percentage variables (`variance_pct`, `budget_drift_threshold_pct`, `actual_vs_expected_pct`)

### Settings Field Names: Lowercase with Env Var Alignment

All `Settings` fields are lowercase with underscores and correspond directly to environment variable names (case-insensitive via `case_sensitive=False`). Where the env var name deviates from the field name, `AliasChoices` is used (e.g., `service_api_key` with alias `["ACCOUNT_STATUS_RECON_SERVICE_KEY", "SERVICE_API_KEY"]`). The private `_SERVICE_KEY_ALIAS` ClassVar holds the service-specific alias.

### Telemetry Attribute Names: `account_status_recon.{noun}` Span Attributes

Custom OTel span attributes in `orchestrator.py` use the `account_status_recon.` prefix for service-specific attributes (e.g., `"account_status_recon.sources_succeeded"`, `"account_status_recon.degraded_sources"`). SDK-provided constants from `autom8y_telemetry.conventions` are used for cross-service attributes (e.g., `RECONCILIATION_SOURCES_COUNT`).

---

## Knowledge Gaps

- **Root-level `ruff.toml` conventions** are captured but the banned-api rule enforcement against specific patterns (e.g., no direct `httpx`, no `loguru`) could not be fully cross-checked against the source files without running ruff.
- **Test naming patterns** (`test_rules.py`, `test_joiner.py`, etc.) and fixture depth are observed but test-coverage specifics are out of scope for this domain (see `test-coverage.md`).
- **`secretspec.toml`** was not read — it governs secret resolution but is infrastructure config rather than code convention.
- **`Justfile`** targets were not read — they may encode additional developer workflow conventions.
