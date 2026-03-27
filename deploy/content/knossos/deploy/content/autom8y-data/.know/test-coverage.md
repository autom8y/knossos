---
domain: test-coverage
generated_at: "2026-03-18T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "51f5e8d"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Packages Without Test Files (Critical Assessment)

**Covered packages** (have dedicated test files or subdirectories):

| Source Package | Test Coverage |
|---|---|
| `analytics/` engine + execution | `tests/analytics/test_analytics_engine.py`, `test_engine_*.py` (lifecycle, batch, cache, errors, happy path) |
| `analytics/core/query/` | `tests/analytics/core/query/` — 5 test files (filters, executor timeout, SQL generator temporal, merge key coercion, count mode temporal). Builder, planner, dimension_resolver, fact_resolver, filter_merge, options, window_metric_sql covered indirectly via integration/analytics tests |
| `analytics/core/metrics/` | `tests/analytics/core/metrics/` — 8 test files including golden tests and production_registry fixture |
| `analytics/core/infra/` | `tests/analytics/core/infra/` — 4 files (connection_read_replica, datetime_utils_sql, metrics_timeout, query_cache); plus `tests/unit/analytics/core/infra/` — 2 files (cache_metrics, ttl_policy) |
| `analytics/core/execution/` | Covered indirectly: `test_composite_metrics.py`, `test_composite_after_rolling.py`, `test_rolling_aggregator.py`, `test_rolling_window_boundaries.py`, `test_count_distinct_rolling.py`, `test_multi_fact_rolling.py`; integration: `tests/integration/test_rolling_window_joins_integration.py` |
| `analytics/primitives/` | `tests/analytics/primitives/` — full subdirectory coverage (anomaly, config, correlation, creative, efficiency, health, optimization, pacing, peer, shared) |
| `api/routes/` | `tests/api/routes/` — 7 test files (analytics_health, business, data_service, detail_endpoints, gid_mappings, health, intelligence, messages, messages_export, messages_ws3, phone_utils) |
| `api/services/` | `tests/api/services/` — 14 test files |
| `core/` | `tests/core/` — 14 test files (config, validators, ORM types, repositories, factories) |
| `grpc/` | `tests/grpc/` — 4 test files (server, adapters/base, handlers/health_handler, handlers/lead_handler) |
| `services/` | `tests/services/` — 13 test files covering all service modules |

**Uncovered or critically thin packages:**

1. **`analytics/core/dimensions/` — thin coverage.** 7 source files (cache.py, computed.py, manager.py, overrides.py, scope.py, time_dimensions.py, time_intelligence.py). Only one dedicated test file: `tests/analytics/core/dimensions/test_dimension_type_inference.py`. The `time_intelligence.py`, `cache.py`, `manager.py`, `overrides.py` modules have 0 direct test file references. Dimension behavior is exercised indirectly by hundreds of analytics tests, but isolated unit tests for dimension cache and override logic are absent.

2. **`analytics/core/joins/` — no dedicated test directory.** 3 source files (canonical_paths.py, graph.py, optimizer.py). Join behavior covered only through integration and behavioral tests: `test_cartesian_product_prevention.py`, `test_discovery_join_graph.py`, `test_offer_join_paths.py`, `test_sql_join_generation.py`, `test_asset_joins.py`. No unit tests for `graph.py` or `optimizer.py` in isolation.

3. **`analytics/core/infra/audit_trail.py`** — 0 test file references found. No test exercises audit trail logging in isolation.

4. **`analytics/core/infra/pool_metrics.py`** — 0 test file references found.

5. **`analytics/fixtures/`** (builder.py, config.py, schema.py) — only `test_fixture_scaling.py` (performance) exists; no unit tests for fixture builder logic.

6. **`analytics/routes/`** — 7 route modules exist at `src/autom8_data/analytics/routes/`. Tests exist at `tests/api/routes/` but these map to `api/routes/`, not `analytics/routes/`. The analytics-layer routes (admin.py, analytics_health.py, data_service.py, insights.py, intelligence.py, query.py, schema.py) appear to be the same routes tested via the API test suite, but no dedicated `tests/analytics/routes/` directory exists.

7. **`clients/` package** — `src/autom8_data/clients/` contains only `__init__.py`. The `tests/clients/` directory contains only `__init__.py`. No client code to test.

8. **Empty stub directories** — `tests/cli/`, `tests/semantic_analytics/`, `tests/enrichment/` each contain only `__init__.py`, indicating planned but not implemented test suites for corresponding source functionality.

9. **`tests/analytics/test_analytics_engine.py` — 8 tests skipped.** Reasons: "Test needs update for new engine architecture." Tests at lines 158, 181, 210, 225, 252, 272, 301, 335 are all marked `@pytest.mark.skip`. This represents an active coverage gap in AnalyticsEngine unit test coverage.

### Critical Path Coverage Assessment

| Critical Path | Coverage Status | Key Test Files |
|---|---|---|
| Analytics engine `.get()` | Strong — multiple happy path + error tests | `test_engine_get_happy_path.py`, `test_engine_get_errors.py`, `test_engine_batch.py`, `tests/golden_master/test_analytics_golden.py` |
| Query compilation pipeline | Moderate — SQL generation well covered; planner/resolver only via integration | `test_sql_generation.py`, `test_sql_join_generation.py`, `test_cartesian_product_prevention.py` |
| Rolling window aggregation | Strong | `test_rolling_aggregator.py`, `test_count_distinct_rolling.py`, `test_multi_fact_rolling.py`, `test_rolling_window_boundaries.py`, integration test |
| Materializer | Strong — 4 test files (base, h3, h4, ri) | `test_materializer.py`, `test_materializer_h3.py`, `test_materializer_h4.py`, `test_materializer_ri.py` |
| Composite metrics | Strong | `test_composite_metrics.py`, `test_composite_after_rolling.py`, `analytics/core/metrics/test_composite_golden.py` |
| Cartesian product prevention | Good — regression coverage | `test_cartesian_product_prevention.py`, `test_dimension_adaptation_join_safety.py` |
| Circuit breaker | Good | `test_circuit_breaker.py`, `test_consumer_circuit_breaker.py` |
| API routes | Strong — CRUD + services + routes suites | `tests/api/` (97 test files, ~2,206 test functions including class-based) |
| gRPC handlers | Moderate — unit tests present, integration gated on DATABASE_URL | `tests/grpc/handlers/`, `tests/grpc/adapters/` |

### Negative Test / Error Path Coverage

- 574 `pytest.raises` usages across test suite — error path testing is present.
- 144 test functions with names containing error/exception/fail/invalid/empty/null/negative (case-insensitive match).
- `tests/analytics/test_calls_adversarial.py`, `tests/analytics/test_fsh_adversarial.py`, `tests/analytics/test_window_metrics_adversarial.py`, `tests/api/services/test_question_level_stats_adversarial.py` — explicit adversarial test files exist.
- **Gap**: `analytics/core/infra/audit_trail.py` and `analytics/core/infra/pool_metrics.py` have no error path coverage.
- **Gap**: 8 skipped tests in `test_analytics_engine.py` represent untested error/edge paths for engine lifecycle and stats APIs.

### Prioritized Gap List

1. **HIGH**: `tests/analytics/test_analytics_engine.py` — 8 skipped tests ("Test needs update for new engine architecture"). Engine architecture changed but tests not updated.
2. **HIGH**: `analytics/core/dimensions/` — no unit tests for dimension cache, manager, overrides, scope logic. Only type inference tested in isolation.
3. **MEDIUM**: `analytics/core/joins/graph.py` and `optimizer.py` — no unit tests; only behavioral/regression coverage via higher-level tests.
4. **MEDIUM**: `analytics/core/infra/audit_trail.py` — zero test coverage.
5. **MEDIUM**: `analytics/core/infra/pool_metrics.py` — zero test coverage.
6. **LOW**: `analytics/fixtures/builder.py` — no unit tests; only scaling test exists.
7. **LOW**: `tests/cli/`, `tests/semantic_analytics/`, `tests/enrichment/` — stub directories with no content.

---

## Testing Conventions

### Naming Conventions

- **Test files**: `test_{subject}.py` (e.g., `test_sql_generation.py`, `test_composite_metrics.py`)
- **Test functions**: `test_{action_or_scenario}()` — snake_case descriptive names (e.g., `test_leads_metric_count_aggregation`, `test_empty_data_handling`)
- **Test classes**: `Test{Subject}` — used in `tests/unit/` subdirectory and `tests/api/` CRUD test classes; e.g., `TestResilientCoreClientInit`, `TestMetricTypeEnumProperties`
- **Adversarial test files**: named `test_{subject}_adversarial.py`
- **Golden master test files**: live in `tests/golden_master/`

### Fixture Patterns

**Conftest hierarchy** (8 conftest.py files):

| File | Scope | Key Fixtures |
|---|---|---|
| `tests/conftest.py` | Global | `reset_data_policy_after_test` (autouse), `permissive_data_policy`, `log_reset_state` (autouse override), `_loguru_caplog_propagation` (session) |
| `tests/analytics/conftest.py` | Analytics suite | `empty_registry`, `mock_registry`, `registry_with_deprecated_metrics`, `mock_registry_factory`, `sample_dataframe_factory`, `empty_dataframe`, `dataframe_with_nulls`, `trailing_7_days`, `trailing_30_days`, `month_boundary_period`, `year_boundary_period` |
| `tests/analytics/core/metrics/conftest.py` | Metrics tests | `production_registry` (module-scoped) — full production registry with all 8 registration functions called |
| `tests/analytics/primitives/conftest.py` | Primitives suite | `reset_primitives_config` (autouse) — clears config cache and env vars |
| `tests/api/conftest.py` | API suite | `module_client` (module-scoped TestClient), `disable_auth_for_api_tests` (autouse), `disable_rate_limiting_for_api_tests` (autouse), `pytest_collection_modifyitems` (skips CRUD tests without DATABASE_URL) |
| `tests/api/services/conftest.py` | API services | (not read; assumed service-layer fixtures) |
| `tests/golden_master/conftest.py` | Golden master | Seeded DuckDB, `AnalyticsEngine` instance, dtype normalization utilities, deterministic date fixtures, syrupy `SnapshotAssertion` via `JSONSnapshotExtension` |
| `tests/grpc/conftest.py` | gRPC suite | `mock_session`, `mock_session_factory`, `pytest_collection_modifyitems` (skips integration without DATABASE_URL) |
| `tests/performance/conftest.py` | Performance suite | `baseline_results` (session), `performance_client` (session TestClient, AUTOM8Y_ENV=mock), `analytics_engine` (session, pytest_asyncio), `save_baseline_csv` (autouse, exports CSV after all tests) |

**Fixture scope patterns**:
- `autouse=True` for cleanup/isolation (data policy reset, loguru bridge, config cache clear, auth disable)
- `scope="module"` for expensive setup (production_registry, module_client)
- `scope="session"` for shared engine/client in performance tests
- Factory pattern: fixtures returning callables (`sample_dataframe_factory`, `mock_registry_factory`)

### Assertion Patterns

- **pytest native assertions**: standard `assert` statements throughout
- **`pytest.raises`**: 574 usages — primary exception testing pattern
- **syrupy snapshot assertions**: `snapshot == result` pattern in golden master tests (`tests/golden_master/`); snapshots stored in `tests/golden_master/__snapshots__/*.ambr`
- **Mock assertions**: 502 usages of `assert_called`, `assert_awaited`, `called_once`, `call_count`, `call_args` patterns
- **No `self.assertEqual` / unittest-style assertions** observed in function-based tests; class-based tests in `tests/unit/` use `assert` not `self.assert*`

### Mock Patterns

- `unittest.mock`: `MagicMock`, `AsyncMock`, `patch` — 4,508 mock-related imports/usages across 289 files
- `AsyncMock` used heavily for async service/engine mocking
- `patch()` decorator and context manager both used
- `MagicMock` used for sync dependencies (registries, session factories)

### Async Test Patterns

- **`asyncio_mode = "auto"`** in `pyproject.toml` — all `async def test_*` functions automatically treated as async tests (no `@pytest.mark.asyncio` required per-test)
- 91 async test functions (using `async def test_`)
- `pytest_asyncio.fixture` used in performance conftest for session-scoped async fixtures
- 1,032 occurrences of `pytest.mark.asyncio` (these are legacy marks; auto mode makes them redundant but harmless)

### Pytest Marks

Registered in `pyproject.toml`:

| Mark | Count | Usage |
|---|---|---|
| `asyncio` | 1,032 | Legacy per-test async marker (auto mode renders redundant) |
| `parametrize` | 82 | Parameterized test cases |
| `smoke` | 26 | Fast-feedback critical path tests (target: <30s total) |
| `skip` | 23 | Conditional skips (database not configured, architecture gaps) |
| `benchmark` | 10 | Performance benchmark tests |
| `slow` | 8 | Slow tests (>5s) |
| `xfail` | 8 | Expected failures with documented reasons |
| `integration` | 2 | Tests requiring external services |
| `skipif` | 9 | Conditional skips (e.g., Prometheus not installed) |

**Custom marks registered**: `benchmark`, `slow`, `integration`, `smoke`, `smoke_live` (live MySQL).

### Test Data Patterns

- **In-memory DuckDB**: primary test data store; seeded with deterministic data in golden master and performance suites
- **Performance baselines CSV**: `tests/performance/baselines/baseline_YYYYMMDD_HHMMSS.csv` — multiple historical baselines present (oldest: 2025-11-26)
- **Snapshot files**: `tests/golden_master/__snapshots__/` — 6 `.ambr` files (syrupy JSON snapshots)
- **No separate `test_data/` or `fixtures/` data directories** — test data constructed inline or via factory fixtures
- **`TEST_PHONE = "+19259998806"`** — sentinel test business phone used across analytics and performance tests

### Database Skip Pattern

Two conftest files implement identical `pytest_collection_modifyitems` hooks:
- `tests/api/conftest.py` — skips 14 CRUD test modules when `DATABASE_URL` not set
- `tests/grpc/conftest.py` — skips `@pytest.mark.integration` gRPC tests when `DATABASE_URL` not set

This allows full CI runs without a database while CRUD tests run locally.

---

## Test Structure Summary

### Distribution Overview

| Directory | Test Files | Test Functions (incl. class methods) | Notes |
|---|---|---|---|
| `tests/analytics/` | 162 | ~3,342 | Largest suite; analytics engine, query, metrics |
| `tests/api/` | 97 | ~2,206 | CRUD + routes + services |
| `tests/primitives/` | 10 | ~201 | Health + peer primitives |
| `tests/core/` | 14 | ~212 | ORM, config, repositories |
| `tests/golden_master/` | 8 | ~112 | Snapshot regression tests |
| `tests/services/` | 12 | ~308 | Service-layer unit tests |
| `tests/integration/` | 6 | ~121 | End-to-end DB-dependent tests |
| `tests/unit/` | 7 | ~108 | Pure unit tests (transport, aggregation taxonomy) |
| `tests/grpc/` | 4 | ~48 | gRPC handler + adapter tests |
| `tests/performance/` | 5 | ~27 | SLA benchmarks |
| `tests/utils/` | 1 | ~29 | Phone normalizer |
| `tests/` (top-level) | 5 | ~130 | Alembic, cache metrics, engine integration |
| **TOTAL** | **331** | **~6,848** | Across 88 test directories |

### Most Heavily Tested Areas

1. **Analytics engine and query pipeline** (`tests/analytics/`) — 162 test files covering SQL generation, rolling windows, composite metrics, materializer, dimensions, filters, discovery API, circuit breaker, scheduler, and insight executor. This is the domain core and reflects the complexity of the analytics semantic layer.

2. **API layer** (`tests/api/`) — 97 test files covering all CRUD endpoints, auth, rate limiting, batch operations, middleware, CORS, error handlers, gRPC/REST parity, reconciliation contracts.

3. **Analytics primitives** (`tests/analytics/primitives/`) — 84 test files across anomaly detection, correlation, creative analysis, efficiency, health scoring, optimization, pacing, peer comparison, and shared bucketing/normalization.

### Integration vs Unit Tests

- **Unit tests** (`tests/unit/`): 7 files in dedicated `unit/` directory. Additional unit-style tests exist throughout `tests/analytics/` (using mock registries and in-memory DuckDB with no external deps).
- **Integration tests** (`tests/integration/`): 6 files, gated on `DATABASE_URL`. Cover business API, cursor pool concurrency, data service API, DuckDB ibis queries, rolling window joins, temporal joins.
- **Golden master tests** (`tests/golden_master/`): 8 files using syrupy snapshots + seeded DuckDB. These are the most integration-like tests without requiring MySQL.
- **Performance/SLA tests** (`tests/performance/`): 5 files measuring p50/p95/p99 latency with CSV baseline exports. Use in-memory DuckDB (AUTOM8Y_ENV=mock).
- **CRUD tests** (`tests/api/test_*_crud.py`): ~14 files, conditionally skipped without DATABASE_URL.

### Test Invocation

**Run all tests:**
```
pytest tests/
```

**Run without slow tests:**
```
pytest -m "not slow" tests/
```

**Run smoke tests only:**
```
pytest -m smoke tests/
```

**Run excluding database-dependent tests:**
```
pytest tests/  # DATABASE_URL not set → CRUD and integration tests auto-skipped
```

**Run performance benchmarks:**
```
pytest tests/performance/
```

**Run with coverage:**
```
pytest --cov=src/autom8_data --cov-report=html tests/
```

Coverage config: branch coverage, `src/autom8_data` as source, missing lines shown, excludes `pragma: no cover`, `TYPE_CHECKING` blocks, abstractmethods.

### Conftest.py Hierarchy Summary

```
tests/conftest.py                           <- global data policy + loguru bridge
tests/analytics/conftest.py                 <- registry + dataframe + time period fixtures
tests/analytics/core/metrics/conftest.py    <- production_registry (module-scoped)
tests/analytics/primitives/conftest.py      <- primitives config isolation
tests/api/conftest.py                       <- TestClient + auth/rate-limit disable + DB skip
tests/api/services/conftest.py              <- service-layer fixtures
tests/golden_master/conftest.py             <- seeded DuckDB + engine + syrupy
tests/grpc/conftest.py                      <- mock session + DB skip for integration
tests/performance/conftest.py               <- session-scoped engine + CSV baseline export
```

---

## Knowledge Gaps

1. **Actual test execution / branch coverage percentages** — the coverage configuration exists (`pyproject.toml`) but no `.coverage` file or `htmlcov/` directory was found in the repository root. Actual measured coverage % per module is unknown without running `pytest --cov`.

2. **`tests/api/services/conftest.py` content** — not read; fixtures it provides to the 14 API services tests are undocumented here.

3. **`tests/analytics/core/domain/test_operational_mode_persistence.py`** — listed in test files; the `tests/analytics/core/domain/` directory has no other test files despite the source `core/domain/` package existing. Whether `operational_mode_persistence` is the only domain-level concern tested is unclear.

4. **Indirect coverage depth** — many source modules (e.g., `analytics/core/execution/orchestrator.py`, `analytics/core/dimensions/time_intelligence.py`) are exercised only through higher-level integration tests. The depth of branch-level coverage within these modules is unknown without instrumented runs.

5. **`tests/analytics/test_analytics_engine.py` skip rationale depth** — 8 tests are skipped referencing "new engine architecture." The specific architecture change that invalidated them is not documented in the test file beyond the skip reason strings.
