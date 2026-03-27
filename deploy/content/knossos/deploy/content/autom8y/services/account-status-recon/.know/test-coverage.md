---
domain: test-coverage
generated_at: "2026-03-16T14:32:40Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.9
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Covered Critical Paths

The following source modules have corresponding test files:

- `src/account_status_recon/rules.py` — covered by `tests/test_rules.py` (all 5 rule functions + `apply_all_rules`), `tests/qa/test_edge_cases_adversarial.py` (EC-1 through EC-20), and `tests/qa/test_qa_adversary.py`
- `src/account_status_recon/joiner.py` — covered by `tests/test_joiner.py` (all 7 presence states, dedup, aggregation, deterministic ordering) and `tests/qa/test_qa_adversary.py` (key extraction edge cases, whitespace variants)
- `src/account_status_recon/orchestrator.py` — covered by `tests/test_orchestrator.py` (happy path, degraded mode, all-sources-fail abort, GAP-1 multi-axis scenario) and `tests/test_instrumentation.py` (7 telemetry span tests)
- `src/account_status_recon/fetcher.py` — covered by `tests/test_fetcher.py` (FetchResult structure, fetch_all success/partial/full failure, exception handling)
- `src/account_status_recon/readiness.py` — covered indirectly via `tests/qa/test_edge_cases_adversarial.py` (EC-13, EC-14, EC-15: truncation, staleness warn/fail) and `tests/qa/test_qa_adversary.py` (mixed pass/warn/fail readiness)
- `src/account_status_recon/report.py` — covered by `tests/qa/test_edge_cases_adversarial.py` (EC-17 mrkdwn escaping) and `tests/qa/test_qa_adversary.py` (None business_name, all financial fields, all-clear variant)
- `src/account_status_recon/models.py` — covered indirectly throughout all test files via factory functions in `tests/conftest.py`
- `src/account_status_recon/handler.py` — covered by `tests/test_instrumentation.py` (TC-7: metrics span via `lambda_handler`)
- `src/account_status_recon/errors.py` — covered in `tests/test_instrumentation.py` (ReportError raised on Slack failure)

### Untested Source Modules

- `src/account_status_recon/config.py` — No dedicated test file. Settings construction and validation are tested indirectly via `MagicMock` settings objects in orchestrator and fetcher tests. No test exercises actual `pydantic-settings` validation, secret loading, or env-var parsing. **This is the highest-risk untested area.**
- `src/account_status_recon/metrics.py` — No dedicated test file. The `ReconciliationMetrics` class is exercised only via mock in `tests/test_instrumentation.py` (TC-7). Actual CloudWatch `put_metric_data` call paths and metric field mappings are not tested.
- `src/account_status_recon/__main__.py` — No test. Entry-point wrapper; low criticality.

### Test Blind Spots

**Error paths**:
- `fetch_billing`, `fetch_campaigns`, `fetch_offers` functions in `src/account_status_recon/fetcher.py` are individually tested only via `FetchResult` boundary type assertions (`tests/test_fetcher.py` lines 36-51). The actual HTTP client calls inside these functions (lazy imports, `autom8y_http`, auth token exchange) are never exercised even with mocks — the tests bypass the functions entirely.
- Config loading failure paths (missing env vars, invalid secret format) are not tested.

**Negative tests present**: Rules tests include explicit negative tests — e.g., `test_billing_nan_guard`, `test_billing_inf_guard`, `test_budget_unavailable_zero`, `test_status_both_absent`, `test_three_way_zero_expected`. NaN/Inf guards are verified to return `[]` or `None`.

**Coverage measurement infrastructure**: `pyproject.toml` configures `pytest-cov` (line 113). The `just test-cov` recipe runs `pytest --cov=src --cov-report=term-missing --cov-report=html tests/`. No CI coverage gate or numeric threshold is enforced in the observed configuration.

### Prioritized Gap List

1. **`src/account_status_recon/config.py`** (HIGH RISK): Real settings validation is never tested. A misconfigured secret format or renamed env var would only surface at Lambda startup in production.
2. **`src/account_status_recon/metrics.py`** (MEDIUM RISK): CloudWatch metric field names and values are untested. A rename in `ReconciliationMetrics` would not be caught.
3. **`src/account_status_recon/fetcher.py` — HTTP internals** (MEDIUM RISK): The actual `fetch_billing`, `fetch_campaigns`, `fetch_offers` HTTP call paths are untested. Transport-layer changes would not be caught.

---

## Testing Conventions

### Test Function and Class Naming

Tests are organized exclusively with `class`-per-subject and `def test_*` methods. No module-level bare test functions are used. Class naming conventions:

- One class per rule function: `TestRuleStatus`, `TestRuleBudget`, `TestRuleDelivery`, `TestRuleBilling`, `TestRuleThreeWay`, `TestApplyAllRules` (in `tests/test_rules.py`)
- One class per concept: `TestThreeWayJoin` (in `tests/test_joiner.py`), `TestFetchBilling`, `TestFetchAll` (in `tests/test_fetcher.py`)
- EC-numbered classes for PRD edge cases: `TestEC1AllSourcesAllMatch` through `TestEC20SlackFailureAfterVerdicts` (in `tests/qa/test_edge_cases_adversarial.py`)
- Orchestrator classes by scenario: `TestOrchestratorHappyPath`, `TestOrchestratorDegradedMode`, `TestOrchestratorGAP1` (in `tests/test_orchestrator.py`)
- Telemetry classes by span convention ID: `TestFetchSpan`, `TestReadinessGateSpan`, `TestJoinSpan`, `TestVerdictsSpan`, `TestReportSpan`, `TestEventPublishSpan`, `TestMetricsSpan`, `TestConventionCompliance` (in `tests/test_instrumentation.py`)

Method names use `test_{descriptive_condition}` pattern. Docstrings on test methods describe the scenario and expected outcome (e.g., `"""Active contract + campaign present -> ALIGNED."""`).

### Assertion Patterns

No third-party assertion library (e.g., pytest-check, Hamcrest) is used. All assertions use stdlib `assert` statements:

```python
assert result is not None
assert result.axis == VerdictAxis.STATUS
assert result.value == StatusVerdict.ALIGNED
assert result.severity == Severity.CRITICAL
assert len(findings) == 0
assert any(v.value == BillingVerdict.ADS_RUNNING_NO_PAYMENT for v in result)
```

`pytest.raises` is used for expected exception scenarios:
```python
with pytest.raises(ReportError):
    await run_reconciliation()
```

### Test Helper and Factory Patterns

All synthetic data construction is centralized in `tests/conftest.py` as plain factory functions (not fixtures, except for 4 `@pytest.fixture`-decorated wrappers):

- `make_billing(**kwargs) -> BillingData` — builds `BillingData` with sensible defaults
- `make_billing_row(**kwargs) -> dict` — builds raw billing insight row dict
- `make_campaign(**kwargs) -> CampaignData` — builds `CampaignData`
- `make_campaign_item(**kwargs) -> dict` — builds raw campaign tree item dict
- `make_contract(**kwargs) -> ContractData` — builds `ContractData`
- `make_contract_row(**kwargs) -> dict` — builds raw Asana offer row dict
- `make_account_record(**kwargs) -> AccountRecord` — builds `AccountRecord` with auto-computed `SourcePresence`

All factories accept keyword overrides; defaults are chosen to produce "healthy/normal" values (e.g., `spend=100.0, collected=95.0, variance=-5.0`). Tests override only the fields relevant to the case under test.

Fixtures provided: `billing_data`, `campaign_data`, `contract_data`, `full_account_record`. These are used only in simple test cases; most tests call factory functions directly.

Mock helpers are defined locally within test files rather than in conftest (e.g., `_mock_settings()` in `tests/test_orchestrator.py` and `tests/test_instrumentation.py`, `_make_fetch_results()` duplicated in both files).

### Skip and Marker Patterns

`@pytest.mark.asyncio` is used in `tests/test_instrumentation.py` for async test methods. The `asyncio_mode = "auto"` configuration in `pyproject.toml` (line 61) makes this marker optional for most tests — async methods in `tests/test_orchestrator.py` and `tests/test_fetcher.py` do not use the decorator and rely on auto-mode.

No `t.Skip` / `pytest.skip` patterns are observed in any test file. No build-tag integration test markers (e.g., `@pytest.mark.integration`) are in use despite the `test-int` just recipe filtering on `-m "integration"`.

### Test Environment Management

No `tempfile`/`t.TempDir()` or filesystem-based test setup is used (this is a pure-logic service with no file I/O). Test isolation is via `unittest.mock.patch` context managers and `AsyncMock`.

The instrumentation tests use a special autouse fixture to reset module-level `_tracer` objects:
```python
@pytest.fixture(autouse=True)
def _reset_otel_provider(convention_span_exporter: InMemorySpanExporter) -> None:
```

The `convention_span_exporter` fixture is provided by `autom8y-telemetry[conventions]` via a `pytest11` plugin entry point (not defined in this repo's test files).

### Test Data Patterns

No `testdata/` directory or fixture files (JSON, YAML, golden files) are used. All test data is constructed inline via factory functions. Phone numbers used in tests are consistently `+15551234567` (single account) or `+155500000NN` (multi-account scenarios). EC test class names are referenced in docstrings to PRD requirements (e.g., `"""EC-11: NaN spend -> empty list."""`).

---

## Test Structure Summary

### Distribution

**Source modules**: 11 substantive Python modules in `src/account_status_recon/` (excluding `__init__.py`, `__main__.py`, `__pycache__`).

**Test files**: 6 test files across 2 directories:
- `tests/test_rules.py` — 5 test classes, ~40 test methods; covers all 5 rule functions
- `tests/test_joiner.py` — 1 class, 10 test methods; covers `three_way_join`
- `tests/test_orchestrator.py` — 3 classes, 7 async test methods; covers `run_reconciliation` pipeline
- `tests/test_fetcher.py` — 2 classes, 6 async test methods; covers `fetch_all` and `FetchResult`
- `tests/test_instrumentation.py` — 8 classes, 13 async/sync test methods; covers all 7 OTEL convention spans
- `tests/qa/test_edge_cases_adversarial.py` — 20 EC-numbered classes, ~22 test methods; covers all PRD edge cases
- `tests/qa/test_qa_adversary.py` — 17 adversarial classes, ~40 test methods; covers gaps not addressed by main test suite

Ratio: 9 of 11 substantive source modules have some test coverage (82%). 2 modules (`config.py`, `metrics.py`) are only exercised via mocks.

### Most Heavily Tested Areas

- **Rules engine** (`rules.py`): Covered by three separate test files — `test_rules.py`, `test_edge_cases_adversarial.py`, and `test_qa_adversary.py`. Threshold boundary conditions, NaN/Inf guards, classification whitespace normalization, and multi-anomaly billing overlap all have dedicated tests.
- **Three-way join** (`joiner.py`): All 7 presence states (FR-3) verified in a single test. Key extraction edge cases (malformed bullets, whitespace, missing keys) covered in `test_qa_adversary.py`. Campaign budget aggregation and first-wins dedup both tested.
- **Telemetry convention compliance** (`orchestrator.py`, `handler.py`): 7 distinct span conventions (TC-1 through TC-7) each have 2-3 test methods verifying attribute names, types, and side-effect event structure.

### Package Naming Pattern

All test files use `package foo_test` style via the `tests/` directory layout with `from conftest import ...` (flat import via `pythonpath = ["src", "tests"]` in pyproject.toml). No `from account_status_recon_test import ...` pattern — tests import source modules directly. This means tests are **external by file location** but use direct imports (not the `package foo_test` distinction seen in Go). No test uses `from __future__ import annotations` to create package-internal tests.

### Integration vs. Unit Test Distinction

No formal `@pytest.mark.integration` markers are applied to any test despite the `just test-int` recipe existing. All tests execute with `just test` or `pytest tests/`. Instrumentation tests in `tests/test_instrumentation.py` are the closest to integration tests — they exercise the full async orchestrator pipeline with only fetch, Slack, and EventBridge mocked out.

The `tests/qa/` subdirectory is not formally gated; it runs as part of the default test suite.

### Test Invocation Command

```
uv run pytest tests/ -v
```

Or via just:
```
just test        # all tests
just test-unit   # unit only (excludes @pytest.mark.integration, though none exist currently)
just test-cov    # tests with coverage report (pytest --cov=src --cov-report=term-missing)
```

The `PYTEST` variable resolves to `uv run pytest` (defined in shared `_globals.just`). Test discovery: `python_files = "test_*.py"`, `testpaths = ["tests"]`.

---

## Knowledge Gaps

1. **`src/account_status_recon/config.py` internals**: The actual field names, validators, and secret loading logic were not read. Coverage gap analysis was inferred from the absence of a `test_config.py` file and the presence of `MagicMock` settings in all test helpers.
2. **`src/account_status_recon/metrics.py` internals**: The `ReconciliationMetrics` class fields and `emit_core`/`emit_custom`/`emit_dms_timestamp` method signatures were not inspected. The mock in TC-7 reveals these method names but not their CloudWatch dimension/metric-name mappings.
3. **CI coverage enforcement**: The CI configuration (`just/ci.just`) was not read. It is unknown whether a coverage threshold gate is enforced in CI beyond what is visible in `pyproject.toml`.
4. **`autom8y-reconciliation` SDK internals**: Verdict types, `Correlator`, and `evaluate_readiness` are imported from the SDK package. SDK source was not inspected; test behavior is inferred from test assertions.
