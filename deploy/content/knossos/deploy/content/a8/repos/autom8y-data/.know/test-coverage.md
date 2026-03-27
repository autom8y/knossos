---
domain: test-coverage
generated_at: "2026-03-25T02:05:48Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "b8da042"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Critical Untested Packages

**gRPC layer** is the most acute gap. The gRPC surface has 6 handler files and 7 adapter files; only 2 handlers and 1 adapter base class have tests.

- `src/autom8_data/grpc/handlers/` — 8 files (address, appointment, business, health, lead, payment, vertical); tested: `health`, `lead` only
- `src/autom8_data/grpc/adapters/` — 8 files; tested: `base` only; address, appointment, business, lead, payment, vertical adapters have no tests

**Core repositories** — 6 files, 2 covered:
- `src/autom8_data/core/repositories/` — `business` and `messages` tested; `address`, `appointments`, `dimension_enrichment`, `lead` have no tests

**Service layer gaps** — 35 service files, ~15 have direct tests. Untested service modules:
- `assets_ad_creatives`, `asset_vertical`, `platform_asset`, `question`, `review`, `split_test_config`, `neighborhood`, `employee` have 0 test files directly referencing them
- `ad_persistence_helpers`, `ad_persistence_responses` — referenced by 1 test file each but no dedicated test

**API auth** — `src/autom8_data/api/auth/` has 6 files (`dependencies.py`, `exceptions.py`, `identity_middleware.py`, `jwt.py`, `middleware.py`); only `tests/api/test_auth.py` covers them at a high level

**Analytics layer** — well-covered overall, with one gap:
- `src/autom8_data/analytics/clients/section_timeline.py` — 0 test files

### Excluded Test Directories (not run in normal CI)

These directories exist but are excluded via `pyproject.toml` `addopts`:
- `tests/golden_traces/` — excluded pending `autom8y_telemetry.testing.span_tree` (not yet in registry)
- `tests/property/` — excluded; contains `test_span_tree_invariants.py` using `hypothesis`
- `tests/spikes/` — excluded
- `tests/integration/test_phantom_trace_e2e.py` — excluded

### Prioritized Gap List

1. **HIGH** — gRPC adapters: `address`, `appointment`, `business`, `lead`, `payment`, `vertical` (protocol contract, no tests)
2. **HIGH** — Core repositories: `appointments`, `lead`, `dimension_enrichment` (critical business logic, ORM-heavy, 0 tests)
3. **MEDIUM** — Service layer: `assets_ad_creatives`, `asset_vertical`, `employee`, `neighborhood`, `platform_asset`, `question`, `review`, `split_test_config`, `vertical`
4. **MEDIUM** — API auth internals: `jwt.py`, `identity_middleware.py`, `dependencies.py` (security code, integration-level only)
5. **LOW** — `analytics/clients/section_timeline.py` (single utility file)

### Well-Covered Critical Paths

The analytics query pipeline is comprehensively tested:
- Join graph + Cartesian prevention: `tests/analytics/test_cartesian_product_prevention.py`, `test_discovery_join_graph.py`, `test_sql_join_generation.py`
- Rolling window engine: `test_rolling_aggregator.py`, `test_rolling_window_boundaries.py`, `test_multi_period_rolling.py`, `test_count_distinct_rolling.py`
- Composite metrics: `tests/analytics/core/metrics/test_composite_golden.py`
- Scheduler/materializer: `tests/analytics/test_scheduler.py`, `test_materializer.py`, `test_materializer_h3.py`, `test_materializer_h4.py`, `test_materializer_ri.py`

---

## Testing Conventions

### Test Function Naming

- Files: `test_{module_name}.py` (one test file per source module is the pattern, though not universal)
- Classes: `class Test{Feature}:` — dominant pattern (1,764 test classes vs 282 top-level functions)
- Functions: `def test_{description}` or `async def test_{description}`
- `asyncio_mode = "auto"` is set — async tests do not need `@pytest.mark.asyncio`

### Fixture Patterns

**Conftest hierarchy** (12 conftest files):
```
tests/conftest.py                                  # Root: data policy reset, loguru bridge
tests/analytics/conftest.py                        # Registry factories, DataFrame factories, time period fixtures
tests/analytics/core/metrics/conftest.py           # production_registry (module-scoped, expensive ORM import)
tests/analytics/primitives/conftest.py             # Primitives-specific fixtures
tests/api/conftest.py                              # module_client (TestClient), auth/rate-limit disable
tests/api/services/conftest.py                     # Service-layer fixtures
tests/dev_data/conftest.py                         # Dev data fixtures
tests/golden_master/conftest.py                    # Golden master anonymized loader
tests/golden_traces/conftest.py                    # (excluded)
tests/grpc/conftest.py                             # gRPC fixtures
tests/performance/conftest.py                      # Performance baselines
tests/property/conftest.py                         # (excluded)
```

**Key fixture patterns**:
- Registry fixture: `empty_registry`, `mock_registry`, `production_registry` (module-scoped)
- DataFrame factory: `sample_dataframe_factory(rows, start_date, include_nulls, columns)` returns Polars DataFrames
- API client: `module_client` — module-scoped `TestClient` (avoids 1.8s `create_app()` overhead per test)
- Data policy: `reset_data_policy_after_test` (autouse) and `permissive_data_policy` (2010 min date)
- CRUD tests skip when `DATABASE_URL` not configured (CI compatibility)

**Fixture scope patterns**:
- `autouse=True` for cleanup/isolation (data policy reset, loguru bridge, config cache clear, auth disable)
- `scope="module"` for expensive setup (production_registry, module_client)
- `scope="session"` for shared engine/client in performance tests
- Factory pattern: fixtures returning callables (`sample_dataframe_factory`, `mock_registry_factory`)

### Mock Patterns

- `unittest.mock.MagicMock` / `AsyncMock` — used in 159 test files (dominant pattern)
- `unittest.mock.patch` / `@patch` — used in 169 test files
- Service layer tests: mock the DB session with `AsyncMock`, inject via FastAPI dependency override
- API tests pattern: `get_db_session` dependency overridden with async mock, `TestClient` from `fastapi.testclient`

### Assertion Patterns

- `assert` statements (standard pytest idiom)
- `pytest.raises` used in 128 files for negative path testing
- Syrupy snapshot testing (`snapshot` fixture) used in 21 files, primarily in `tests/golden_master/`
- SQL generation tests assert on string fragments of generated SQL

### Async Pattern

- 1,297 `async def test_` functions, 282 top-level sync `def test_` functions
- `asyncio_mode = "auto"` — no decorator needed
- `pytest-asyncio` + `pytest-timeout` enforced

### Parametrize Usage

- `@pytest.mark.parametrize` used in 49 files — moderate adoption
- `hypothesis` used in 2 files (excluded from default run)

### Dev Data / Factory Pattern

- `autom8-dev-data` package (`packages/autom8-dev-data/`) provides `polyfactory`-based model factories
- Referenced by 7 test files directly; the `fixtures/` source package acts as a bridge
- Golden master tests use an anonymized loader (`tests/golden_master/anonymized_loader.py`)

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

---

## Test Structure Summary

### Distribution by Directory

| Directory | Test Files | Focus |
|-----------|-----------|-------|
| `tests/analytics/` | 181 | Analytics engine, query pipeline, metrics, rolling windows |
| `tests/api/` | 104 | REST CRUD endpoints, auth, rate limiting, response contracts |
| `tests/services/` | 15 | Service-layer unit tests (mocked DB) |
| `tests/core/` | 14 | ORM types, config, factories, repositories |
| `tests/primitives/` | 10 | Health scoring, peer analytics (standalone) |
| `tests/golden_master/` | 8 | Snapshot regression tests for query output |
| `tests/unit/` | 7 | Pure unit tests (transport config, formula taxonomy) |
| `tests/integration/` | 7 | Full-stack: DuckDB, temporal joins, rolling joins |
| `tests/performance/` | 5 | SLA benchmarks (API, engine, fixture scaling) |
| `tests/grpc/` | 4 | gRPC server + 2 handlers |
| `tests/dev_data/` | 3 | Dev data anonymization |
| `tests/enrichment/` | 2 | Enrichment view parity |
| `tests/security/` | 1 | Filter injection |
| `tests/utils/` | 1 | Phone normalizer |

**Total test files**: 372 `test_*.py` files

### Most Heavily Tested Areas

1. **Analytics query pipeline** (`tests/analytics/`) — 181 test files covering SQL generation, rolling windows, composite metrics, materializer, dimensions, filters, discovery API, circuit breaker, scheduler, and insight executor
2. **API layer** (`tests/api/`) — 104 test files covering all CRUD endpoints, auth, rate limiting, batch operations, middleware, CORS, error handlers, reconciliation contracts
3. **Analytics primitives** — 84+ test files across anomaly detection, correlation, creative analysis, efficiency, health scoring, optimization, pacing, peer comparison

### Integration vs Unit Tests

- **Unit tests** (`tests/unit/`): 7 files in dedicated `unit/` directory. Additional unit-style tests exist throughout `tests/analytics/` (using mock registries and in-memory DuckDB with no external deps)
- **Integration tests** (`tests/integration/`): 7 files, gated on `DATABASE_URL`. Cover business API, cursor pool concurrency, data service API, DuckDB ibis queries, rolling window joins, temporal joins
- **Golden master tests** (`tests/golden_master/`): 8 files using syrupy snapshots + seeded DuckDB. The most integration-like tests without requiring MySQL
- **Performance/SLA tests** (`tests/performance/`): 5 files measuring p50/p95/p99 latency with CSV baseline exports. Use in-memory DuckDB (AUTOM8Y_ENV=mock)

### Test Invocation Commands

```bash
# Default (full suite minus excluded dirs)
uv run pytest
# or
just test

# With coverage
uv run pytest --cov=src/autom8_data --cov-report=html

# Smoke only (fast feedback)
uv run pytest -m smoke

# Exclude slow tests
uv run pytest -m "not slow"

# Single directory
uv run pytest tests/analytics/

# Parallel
uv run pytest -n auto
```

### Database Skip Pattern

Two conftest files implement identical `pytest_collection_modifyitems` hooks:
- `tests/api/conftest.py` — skips CRUD test modules when `DATABASE_URL` not set
- `tests/grpc/conftest.py` — skips `@pytest.mark.integration` gRPC tests when `DATABASE_URL` not set

This allows full CI runs without a database while CRUD tests run locally.

### Performance Baseline Infrastructure

`tests/performance/baselines/` contains ~250+ CSV baseline files (regression tracking), actively updated through March 2026. This is a living performance regression suite.

---

## Knowledge Gaps

1. **Actual test execution / branch coverage percentages** — the coverage configuration exists (`pyproject.toml`) but no `.coverage` file was found. Actual measured coverage % per module is unknown without running `pytest --cov`.
2. **`tests/api/services/conftest.py` content** — not read in detail; fixtures it provides to the 14 API services tests are undocumented here.
3. **Indirect coverage depth** — many source modules are exercised only through higher-level integration tests. The depth of branch-level coverage is unknown without instrumented runs.
4. **`tests/semantic_analytics/` status** — contains only `__init__.py`, unclear if placeholder or intentionally empty.
5. **`analytics/core/request/`** — has no dedicated test dir; may be covered indirectly by engine tests.
