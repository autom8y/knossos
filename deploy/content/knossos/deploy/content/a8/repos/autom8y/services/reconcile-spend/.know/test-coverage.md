---
domain: test-coverage
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
# Codebase Test Coverage

## Coverage Gaps

### Modules and their test coverage

| Source Module | Test Coverage | Notes |
|---|---|---|
| `src/reconcile_spend/rules.py` | Comprehensive | `test_rules.py`, `test_rules_properties.py` (property-based), `test_adversarial_data.py` |
| `src/reconcile_spend/orchestrator.py` | Comprehensive | `test_orchestrator.py`, `test_enrichment.py`, `test_parsing_properties.py` |
| `src/reconcile_spend/handler.py` | Comprehensive | `test_handler.py` — 13 tests covering success, error, metrics, period_days dispatch |
| `src/reconcile_spend/config.py` | Comprehensive | `test_config.py`, `test_config_url_guard.py` |
| `src/reconcile_spend/readiness.py` | Comprehensive | `test_readiness.py`, `test_readiness_properties.py` |
| `src/reconcile_spend/report.py` | Comprehensive | `test_report.py`, `test_enrichment.py` (link rendering T12-T13) |
| `src/reconcile_spend/metrics.py` | Comprehensive | `test_metrics.py` — EMF blobs, per-category dimensions, pipeline readiness metric |
| `src/reconcile_spend/clients/data_service.py` | Good | `test_data_service.py`, `test_data_service_circuit_breaker.py` |
| `src/reconcile_spend/clients/models.py` | Comprehensive | `test_client_models.py`, `test_contract_regression.py` |
| `src/reconcile_spend/clients/asana_resolve.py` | Good | `test_enrichment.py` (T5-T8), `test_defect_remediation.py` (DEF-1) |
| `src/reconcile_spend/models.py` | Partial | Tested transitively through all modules. No dedicated unit test for `AnomalyCategory.label`, `severity`, `ReconciliationResult.degraded`, `ReconciliationResult.to_dict()`. |
| `src/reconcile_spend/stubs.py` | None | `stubs.py` contains `ThreeWayComparison` and `AsanaReconciliation` stub classes. Explicitly flagged as NOT imported from the active flow. No tests exist. Intentional — stubs are dead code placeholders. |

### Prioritized Gap List

1. **`stubs.py`** — Zero tests. Low criticality: file is intentionally inactive stub code for future 3-way reconciliation. No production execution path.
2. **`models.py` — domain model properties** — `AnomalyCategory.label` and `AnomalyCategory.severity` properties, `ReconciliationResult.degraded` boolean property, and `ReconciliationResult.to_dict()` are tested only transitively (via handler, orchestrator, and report tests). No isolation tests for these properties exist.
3. **`handler.py` — `_get_event_publisher()`** — The `_EVENTS_AVAILABLE` / `DomainEvent` / `EventPublisher` branch (the "POC 2" domain event publishing path) is not tested. This path requires the `autom8y_events` package to be present. Tests mock `run_reconciliation` before reaching the publisher call, so the publisher invocation path is not exercised.
4. **`orchestrator.py` — `_build_circuit_open_alert()`** — The circuit-breaker alert Slack blocks builder is not directly tested (unlike `_build_stale_data_alert()` which has `test_stale_alert_blocks_structure`). The circuit-breaker path is tested indirectly but the block structure is not asserted.
5. **`orchestrator.py` — `enrich_anomalies()` with no vertical** — The `seen`/`criteria` deduplication path when `record.vertical is None` is not explicitly tested.
6. **No coverage measurement infrastructure** — `pyproject.toml` includes `pytest-cov>=4.0` as a dev dependency but no `--cov` flag appears in `addopts`. No `.coveragerc` exists. Coverage measurement must be invoked manually.

### Negative Test Coverage

Negative tests are well-represented. Examples:
- `test_not_flagged_below_min_spend`, `test_not_flagged_when_invoices_exist` — Rule 1 negative cases
- `test_not_flagged_at_threshold` (R3, R4) — threshold boundary as non-firing case
- `test_handles_invalid_date_gracefully` — Rule 5 malformed date
- `ADV-P01` through `ADV-P09` — adversarial parsing boundary tests
- `TestDef4NonFiniteVariancePct` — non-finite guard tests
- `test_dev_env_with_production_data_url_raises` — production URL guard rejections

### Critical Path Coverage

| Critical Path | Coverage Status |
|---|---|
| Lambda entry -> `run_reconciliation` | Tested via `test_handler.py` (mocked orchestrator) |
| `run_reconciliation` end-to-end flow | Tested via `test_orchestrator.py` (mocked HTTP clients) |
| Data parsing edge cases (E4, E5, E9, E10) | Fully tested: `test_orchestrator.py`, `test_parsing_properties.py`, `test_defect_remediation.py` |
| All 5 anomaly detection rules | Fully tested: positive, negative, boundary, property-based |
| Budget-aware overbilled/underbilled path | Tested: `test_rules.py` budget-aware tests, property P7a/P7b |
| Circuit-breaker open path | Tested: `test_orchestrator.py` + `test_data_service_circuit_breaker.py` |
| Freshness gate PASS/WARN/FAIL | Fully tested: `test_readiness.py`, `test_readiness_properties.py`, `TestFreshnessGate` |
| Asana enrichment full path | Tested: `test_enrichment.py` T9-T11 |
| Slack report truncation (50-block limit) | Tested: `TestReportTruncation` |
| Metrics emission (EMF) | Tested: `test_metrics.py` (all named metrics, dimensions, per-category blobs) |

---

## Testing Conventions

### Test Function Naming

- Test classes use `Test{Feature}` pattern: `TestRuleAdsRunningNoPayment`, `TestRunReconciliation`, `TestFreshnessGate`
- Test methods use `test_{what_and_expected_outcome}` descriptive naming: `test_flagged_when_spending_no_collected_no_invoices`, `test_boundary_exact_min_spend`, `test_not_flagged_when_invoices_exist`
- Property tests use `test_{property_invariant}` naming: `test_no_rule_raises_for_valid_input`, `test_at_most_one_of_r1_r4_per_record`
- ADV tests use ADV-ID naming in docstrings: `ADV-P01` through `ADV-I07`
- Contract tests use CR-ID naming: `CR-RR-01`, `CR-AR-01`
- Defect regression tests use DEF-ID naming: `DEF-1`, `DEF-4`

### Fixture and Parameterize Patterns

- **conftest.py fixtures** — Module-level shared fixtures for `ClientRecord` instances (`record_ads_no_payment`, `record_paying_no_ads`, `record_overbilled`, `record_underbilled`, `record_stale`, `record_clean`, `record_zero_zero`), settings (`mock_settings`), and anomaly collections (`sample_anomalies`, `sample_enriched_anomalies`, `mock_insight_response`)
- **Local `_mock_env` fixtures** — Repeated pattern across `TestRunReconciliation`, `TestCircuitBreakerBehavior`, `TestCreateDataClient`, `TestFreshnessGate`: local `autouse=True` fixture sets env vars via `monkeypatch` and calls `clear_settings_cache()` before and after
- **`@pytest.mark.parametrize`** — Used in `test_config_url_guard.py` to loop over `["development", "local", "test"]` environment names
- **Hypothesis `@given` + `@settings`** — Consistent pattern: `@settings(max_examples=200, deadline=None)` on all property tests

### Assertion Patterns

- `assert result is not None` / `assert result is None` for rule return values
- `assert result.category == AnomalyCategory.X` for category checking
- `assert "string" in result.description` for description content
- `assert len(records) == N` for count assertions
- Mock verification: `mock.assert_awaited_once()`, `mock.assert_called_once_with(...)`
- Pytest raises: `with pytest.raises(ErrorType, match="pattern")`
- For Slack blocks: `assert blocks[N]["type"] == "header"`, `assert "text" in blocks[N]["text"]["text"]`
- Powertools EMF assertions: parse stdout JSON, find `_aws.CloudWatchMetrics`, assert dimension values

### Test Helper Patterns

- `tests/helpers.py` — `to_rows(dicts)` converts raw dict lists to `ReconciliationRow` instances; used pervasively across all test files requiring parsed input
- `tests/strategies.py` — Three Hypothesis strategy factories: `client_record_strategy()`, `budget_aware_record_strategy()`, `reconciliation_row_strategy()`; both with explicit `allow_nan` and `allow_infinity` flags to simulate pre/post-parsing contracts
- `_make_near_miss(**overrides)` helper — locally defined in `test_metrics.py` with defaults, accepting keyword overrides
- `_make_record(variance_pct=..., **kwargs)` helpers — locally defined in `TestComputeNearMissData` and `TestEnrichAnomalies`
- `MockLambdaContext` dataclass — defined in `test_handler.py` to satisfy Powertools `capture_cold_start_metric=True` requirement

### Skip Patterns

No `@pytest.mark.skip` or `@pytest.mark.xfail` decorators found anywhere in the test suite.

### Test Environment Management

- **`_isolate_env` autouse fixture** in `conftest.py` — Clears `AUTOM8Y_ENV` and `ENVIRONMENT` shell-inherited env vars via `monkeypatch.delenv` before every test to prevent direnv leakage
- **`clear_settings_cache()`** — Explicitly called before and after tests that create `Settings` objects; prevents `@lru_cache` stale state between tests
- **Powertools cold-start reset** in `test_handler.py` — `metrics_base.is_cold_start = True` reset between tests; `m.clear_metrics()` and `m.clear_default_dimensions()` to prevent shared state accumulation
- **`_clear_powertools_metrics` autouse fixture** in `test_metrics.py` — Same Powertools reset pattern

### Test Data Directories

No `fixtures/` or `testdata/` directories exist. All test data is defined inline via dicts in conftest.py and individual test files. The `to_rows()` helper in `helpers.py` converts dicts to typed models.

---

## Test Structure Summary

### Overall Distribution

| Layer | Test File(s) | Test Count (approx.) | Coverage Focus |
|---|---|---|---|
| Rules | `test_rules.py`, `test_rules_properties.py` | ~40 + 7 property | All 5 rules, mutual exclusion, boundary, asymmetric thresholds |
| Orchestrator | `test_orchestrator.py` | ~30 | parse_client_records, compute_near_miss_data, run_reconciliation, freshness gate |
| Handler | `test_handler.py` | ~13 | Lambda entry, metrics dispatch, error propagation, period override |
| Config | `test_config.py`, `test_config_url_guard.py` | ~20 | Defaults, secrets, caching, URL guard |
| Readiness | `test_readiness.py`, `test_readiness_properties.py` | ~15 + 1 property | Three-tier PASS/WARN/FAIL, message content |
| Report | `test_report.py` | ~40 | Blocks structure, truncation, data quality, freshness, all-clear |
| Metrics | `test_metrics.py` | ~20 | EMF metric names/values, dimensions, pipeline readiness metric |
| Enrichment | `test_enrichment.py` | ~25 | Stripe fields, EnrichedAnomaly, batch_resolve_units, enrich_anomalies, link rendering |
| Client Models | `test_client_models.py` | ~15 | Pydantic validation for DataQuality, InsightResponse |
| Contract Regression | `test_contract_regression.py` | ~20 | ReconciliationRow and AsanaResolveResponse contract testing |
| Adversarial | `test_adversarial_data.py` | ~35 | ADV-P, ADV-R, ADV-N, ADV-F, ADV-I boundary probing |
| Defect Regression | `test_defect_remediation.py` | ~10 | DEF-1 malformed responses, DEF-4 non-finite variance_pct |
| Data Service | `test_data_service.py`, `test_data_service_circuit_breaker.py` | ~15 | Client factory, HTTP protocol, circuit breaker tripping |
| Parsing Properties | `test_parsing_properties.py` | 1 property (200 examples) | E4/E5/E9/E10 post-conditions |

**Total**: ~300+ discrete test cases across 18 test files (not counting Hypothesis examples which run 200x per property test).

### Most Heavily Tested Areas

1. **Anomaly detection rules** (`rules.py`) — Comprehensive: every rule has positive, negative, boundary, and property-based coverage. Budget-aware path has explicit isolation tests.
2. **Data parsing** (`orchestrator.parse_client_records`) — Covered by unit tests, property tests (P4), adversarial tests, and defect regression tests.
3. **Slack report building** (`report.py`) — Highly tested: block structure, truncation, data quality display, freshness display, mrkdwn escaping, link rendering.
4. **Pipeline readiness gate** (`readiness.py`) — Both explicit test class and property test covering tier partitioning exhaustively.

### Integration vs Unit Test Patterns

- **Unit tests** (dominant): Direct function calls with locally-constructed inputs. No HTTP. Examples: `test_rules.py`, `test_readiness.py`, `test_report.py`, `test_config.py`.
- **Integration tests with mocks** (orchestrator/handler level): `run_reconciliation` is tested with `AsyncMock` for HTTP clients and `SlackClient`, patched at the module level. This verifies the coordination logic without network calls.
- **Property-based tests** (`hypothesis`): 6 property test functions across 3 files (`test_rules_properties.py`, `test_parsing_properties.py`, `test_readiness_properties.py`), each running 200 examples. These verify invariants rather than specific values.
- **Contract regression tests** (`test_contract_regression.py`): A distinct test category focused purely on Pydantic model validation behavior — not business logic.
- **Defect regression tests** (`test_defect_remediation.py`): Named after specific defect IDs (DEF-1, DEF-4). Pattern for tracking that fixed bugs stay fixed.
- **No `@pytest.mark.integration`** markers — there are no integration markers. Test categorization is by file name convention rather than markers.

### conftest.py Patterns

Single `conftest.py` at `tests/conftest.py`. Contains:
- `_isolate_env` — autouse fixture, clears leaked env vars
- `mock_settings` — yields a configured `Settings` instance; calls `clear_settings_cache()` on setup and teardown
- 6 `ClientRecord` fixtures for each anomaly scenario
- `mock_insight_response` — 8 rows with mixed anomaly types plus edge cases (null phone, zero-zero, E10 extra fields)
- `sample_anomalies` and `sample_enriched_anomalies` for report tests

### How Tests are Run

- Runner: `pytest` (configured in `pyproject.toml`)
- `asyncio_mode = "auto"` — all `async def` tests run automatically without explicit `@pytest.mark.asyncio`
- `testpaths = ["tests"]`
- `pythonpath = ["src", "tests"]` — both source package and the `tests/` directory (for `helpers` and `strategies` imports)
- `--import-mode=importlib`
- Coverage: `pytest-cov` installed but not in default `addopts`; must be invoked with `--cov=reconcile_spend`

---

## Knowledge Gaps

1. **Actual coverage percentage unknown** — No coverage run data exists (no `.coverage` file, no configured `addopts --cov`). Coverage measurement must be executed manually to determine line/branch percentages.
2. **`test_adversarial_data.py` full ADV-I tests** — The adversarial integration tests (`ADV-I01` through `ADV-I07`) were partially visible. The integration-level adversarial tests exist but exact IDs and scenarios were not fully read.
3. **`autom8y_events` SDK availability in CI** — `handler.py` has a `try/except ImportError` branch for `autom8y_events`. Tests never install this package; the event publishing path is never exercised. Whether CI includes this package is unknown.
4. **Hypothesis database** — No `.hypothesis/` directory state observed. Shrunk examples from prior runs may not be persisted in CI.
