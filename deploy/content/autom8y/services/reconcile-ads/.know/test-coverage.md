---
domain: test-coverage
generated_at: "2026-03-16T00:02:18Z"
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

# Codebase Test Coverage

## Coverage Gaps

### Source Modules Without Test Files

The service has 12 source files. Coverage by module:

| Source Module | Test File | Coverage Status |
|---|---|---|
| `src/reconcile_ads/handler.py` | `tests/test_handler.py` | Covered |
| `src/reconcile_ads/orchestrator.py` | `tests/test_orchestrator.py`, `tests/test_instrumentation.py` | Covered (highest coverage in suite) |
| `src/reconcile_ads/joiner.py` | `tests/test_joiner.py`, `tests/test_adversarial.py` | Covered (most heavily tested) |
| `src/reconcile_ads/rules.py` | `tests/test_rules.py`, `tests/test_adversarial.py` | Covered |
| `src/reconcile_ads/readiness.py` | `tests/test_readiness.py`, `tests/test_adversarial.py` | Covered |
| `src/reconcile_ads/report.py` | `tests/test_report.py`, `tests/test_adversarial.py` | Covered |
| `src/reconcile_ads/metrics.py` | `tests/test_metrics.py` | Covered |
| `src/reconcile_ads/models.py` | No dedicated test file | No direct test (tested indirectly through all other modules) |
| `src/reconcile_ads/config.py` | No dedicated test file | No direct test (settings cache clear tested via conftest autouse) |
| `src/reconcile_ads/errors.py` | No dedicated test file | No direct test (exercised indirectly via fetcher error paths) |
| `src/reconcile_ads/fetcher.py` | No dedicated test file | **Gap: significant** |
| `src/reconcile_ads/__main__.py` | No dedicated test file | **Gap: significant** |

### Critical Path Coverage Assessment

**Well-covered critical paths:**
- Join engine (index building, two-level join, ghost detection, dedup) — 40 tests in `test_joiner.py` plus 65 adversarial
- Verdict rules (status, budget, delivery, apply_all_rules) — 28 tests in `test_rules.py`
- Pipeline readiness gate (staleness tiers, truncation detection) — 17 tests in `test_readiness.py`
- Handler lifecycle (success, failure, metric isolation, outcome routing) — 6 tests in `test_handler.py`
- Orchestrator pipeline flow (degraded paths, stale abort, Slack failure, truncation abort) — 8 tests in `test_orchestrator.py`
- OTel instrumentation (7 convention spans, child span parent-child, side effect events) — 17 tests in `test_instrumentation.py`

**Under-covered critical paths:**

1. **`fetcher.py`** — No test file. `fetch_ads_tree`, `fetch_asana_offers`, and `fetch_all` contain the interop client lifecycle, circuit-breaker error translation, and `asyncio.gather` failure handling. The orchestrator tests mock `fetch_all` at the call site, so the internal error translation logic (e.g., `InteropAdsError` → `AdsServiceUnavailableError` mapping, `_safe_int` coercion, truncation flag extraction from responses) is untested. This is a **high-criticality gap** — the fetcher is the service boundary with two external systems.

2. **`__main__.py`** — No test file. The local invocation path (`--csv`, `--json` modes, `_finding_to_row` serialization) is entirely untested. Lower operational criticality (this is dev-tooling), but the `_finding_to_row` function has complex conditional logic for verdict serialization that would benefit from tests.

3. **`models.py`** — No direct tests. The Pydantic models are exercised by all other tests, so functional behavior is indirectly covered. However, field validation edge cases and computed properties (e.g., `CampaignMatch.meta_weekly_spend`, `ReconciliationResult.to_dict()`) are not tested in isolation.

4. **`errors.py`** — No direct tests. The error hierarchy (4 classes) is simple and thoroughly exercised through orchestrator/fetcher paths, but the `http_status` and `code` ClassVar attributes are never asserted.

5. **`config.py`** — No direct Settings validation tests. The settings cache clear is covered by the autouse conftest fixture, but `AliasChoices` resolution (e.g., `RECONCILE_ADS_SERVICE_KEY` → `SERVICE_API_KEY` fallback), field validators, and `to_safe_dict()` redaction behavior are not explicitly tested.

### Prioritized Gap List

1. **`fetcher.py`** — High priority. Interop error translation, `_safe_int`, truncation extraction, and `asyncio.gather` dual-failure behavior are all untested production code paths.
2. **`config.py`** — Medium priority. `AliasChoices` fallback and `SecretStr` redaction behavior are load-bearing for multi-environment correctness.
3. **`__main__.py`** — Low priority. Dev tooling; `_finding_to_row` serialization logic is the only actionable target.
4. **`models.py`** — Low priority. Indirectly covered; computed properties could be unit-tested cheaply.
5. **`errors.py`** — Very low priority. Trivial hierarchy; fully exercised transitively.

---

## Testing Conventions

### Test Function Naming

All test functions follow `test_{behavior_under_test}` with descriptive snake_case names. Class-grouped tests use `Test{ModuleName}` or `Test{FunctionName}` class names:

- `TestBuildCampaignIndex`, `TestExecuteJoin`, `TestDetectGhosts` in `test_joiner.py`
- `TestRuleStatusAlignment`, `TestRuleBudgetAlignment`, `TestRuleDeliveryHealth` in `test_rules.py`
- `TestCheckPipelineReadiness`, `TestCompletenessCheck` in `test_readiness.py`
- `TestFetchSpan`, `TestReadinessGateSpan`, `TestJoinSpan` in `test_instrumentation.py`

Function names follow a `test_{condition}_{expected_result}` pattern, e.g., `test_active_not_matched_missing`, `test_drift_between_thresholds`, `test_asana_very_stale_fail`. All test functions have docstrings (single-sentence behavior statement).

### Assertion Patterns

- **Standard `assert` statements** — no custom matchers or `unittest.TestCase` assertion methods. All assertions use raw Python `assert`.
- **`pytest.raises`** used in `test_handler.py` for exception-path assertions.
- **Mock call assertions**: `mock.assert_called_once()`, `mock.assert_called_once_with()` used in handler and orchestrator tests.
- **Attribute pattern**: Tests assert specific object attribute values (`result.degraded is True`, `result.pipeline_readiness == "pass"`) rather than asserting on string representations.
- **Count assertions**: Used throughout joiner tests (`assert len(index) == 1`, `assert len(failures) == 0`).
- **OTel span assertions**: `test_instrumentation.py` uses `dict(span.attributes or {})` pattern to safely access span attributes and assert on specific keys.

### Test Fixture Patterns

**conftest.py** at `tests/conftest.py` provides:
- `autouse=True` fixture `_clear_settings()` — clears `@lru_cache` on `get_settings()` before and after every test. Referenced as SCAR-011 (settings cache poisoning).
- Named fixtures: `default_offer`, `default_tree`, `matched_pair` — rarely used directly (most tests use factory functions).
- Factory functions (not fixtures): `make_offer()`, `make_tree()`, `make_default_campaign_tree()`, `make_campaign_raw_name()`, `make_ad_group_raw_name()` — these are module-level functions with keyword-argument defaults. Tests import them directly from `conftest` by name.

**Local fixtures**: `test_orchestrator.py` defines a local `mock_settings()` fixture. `test_instrumentation.py` uses `convention_span_exporter` and `convention_checker` — provided by the `autom8y-telemetry[conventions]` pytest plugin (registered via `pytest11`).

**Fixture scope**: All fixtures and factory functions use `function` scope (default). The `asyncio_default_fixture_loop_scope = "function"` in pyproject.toml makes the async event loop function-scoped.

### Test Skip/Mark Patterns

- `@pytest.mark.asyncio` used on all async test methods (17 occurrences across `test_orchestrator.py`, `test_instrumentation.py`, `test_adversarial.py`).
- No `@pytest.mark.skip` or `@pytest.mark.xfail` anywhere in the test suite — zero skipped or expected-failure tests.
- No custom marks registered.

### Test Data / Fixtures Directories

No `tests/fixtures/` or `tests/data/` directory exists. All test data is synthetic in-memory construction via conftest factory functions. The conftest module docstring explicitly states: "All test data uses synthetic/fabricated values per NFR-9. No real phone numbers or business names." The bullet separator `\u2022` (bullet) used in campaign/ad-group name encoding is defined as `SEP = "\u2022"` in conftest and imported by test files.

---

## Test Structure Summary

### Overall Distribution

- **Test files**: 9 test files + 1 conftest
- **Total test functions**: 194 across all test files
- **Test classes**: 22 `Test*` classes grouping related tests

| Test File | Test Count | Module Tested | Type |
|---|---|---|---|
| `tests/test_adversarial.py` | 65 | joiner, rules, readiness, report | adversarial/boundary |
| `tests/test_joiner.py` | 40 | joiner.py | unit |
| `tests/test_rules.py` | 28 | rules.py | unit |
| `tests/test_instrumentation.py` | 17 | orchestrator, handler, rules | integration/OTel |
| `tests/test_readiness.py` | 17 | readiness.py | unit |
| `tests/test_report.py` | 8 | report.py | unit |
| `tests/test_orchestrator.py` | 8 | orchestrator.py | integration |
| `tests/test_handler.py` | 6 | handler.py | unit |
| `tests/test_metrics.py` | 5 | metrics.py | unit |

### Most Heavily Tested Areas

1. **Joiner** (`joiner.py`) — 105 tests (40 unit + 65 adversarial). The two-level join engine and edge-case decode behavior is the most exhaustively tested module. Every E-numbered edge case from the PRD is referenced in test docstrings (E1-E18).
2. **Rules** (`rules.py`) — Approximately 90 test cases (28 rules unit + 28 adversarial rules + orchestrator integration). The decision matrix for status alignment (PRD Appendix C) is tested cell-by-cell.
3. **Readiness** (`readiness.py`) — 17 unit tests covering all staleness tier boundaries (exact threshold, 2x threshold, PASS/WARN/FAIL) plus truncation detection for both sources.

### Integration vs Unit Test Patterns

**Unit tests** (`test_joiner.py`, `test_rules.py`, `test_readiness.py`, `test_report.py`, `test_metrics.py`): Pure function calls with no async, no patching of upstream services. Inputs constructed via factory functions.

**Integration tests** (`test_orchestrator.py`, `test_handler.py`): Mock external dependencies (`fetch_all`, `SlackClient`, `asyncio.run`) at the module boundary using `unittest.mock.patch`. Exercise the full pipeline control flow.

**OTel instrumentation tests** (`test_instrumentation.py`): Hybrid — mock fetch/Slack but use real OTel SDK with `InMemorySpanExporter`. Verify span names, attributes, events, and parent-child relationships for 7 convention spans. A TC-9 convention compliance gate calls `assert_convention_compliant()` from the `autom8y-telemetry` testing fixtures.

**Adversarial tests** (`test_adversarial.py`): Boundary/edge-case layer targeting joiner decode robustness (1-field name, unicode in business names, all-caps bullets), rules extreme inputs (budget formula edge cases, delivery health with mixed ad groups), and report truncation.

### Test Invocation Command

From `pyproject.toml` `[tool.pytest.ini_options]`:

```
pytest --import-mode=importlib --tb=short -v
```

`testpaths = ["tests"]`, `asyncio_mode = "auto"` (pytest-asyncio auto mode). Python path includes both `src` and `tests` (allows `from conftest import ...` without package prefix). Dev dependencies include `pytest>=7.0`, `pytest-asyncio>=1.2,<2.0`, `pytest-cov>=4.0`.

---

## Knowledge Gaps

1. **Actual runtime coverage percentages** — No `pytest-cov` report was executed; line/branch coverage numbers for each module are not available. The gap analysis above is structural (file-level) rather than line-level.
2. **Adversarial test class inventory** — `test_adversarial.py` was 66KB and only partially read; the full set of adversarial class names and their targets was not fully enumerated. The count (65 tests) is reliable from Grep.
3. **`autom8y-reconciliation` SDK test coverage** — Tests mock SDK primitives (`ReconciliationMetrics`, `ReconciliationReportBuilder`, `ReadinessGate`) at the class/method level. The SDK's internal behavior is not within scope but the mocking strategy means SDK regressions would pass service tests.
4. **Interop client test coverage** — `fetch_ads_tree` and `fetch_asana_offers` are mocked entirely in orchestrator tests. The actual HTTP + circuit-breaker behavior of `AdsCampaignTreeClient` and `AsanaOfferClient` is exercised only in the interop SDK's own tests, not here.
