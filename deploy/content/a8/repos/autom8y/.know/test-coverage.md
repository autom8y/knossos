---
domain: test-coverage
generated_at: "2026-03-25T12:13:17Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "3fe30a4"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "ee0ce3dc0fda663f9d14c72bd112af1325e753ccadb47a217b6e98bdba62b7ba"
---

# Codebase Test Coverage

## Coverage Gaps

### SDK Packages -- Coverage Assessment

21 SDK packages under `sdks/python/`. All have at least one test file. Coverage depth varies significantly:

**Well-covered SDKs** (test files >= source files):
- `autom8y-telemetry`: src=38, tests=46 -- most tested SDK; includes conventions/, fastapi/, benchmarks/
- `autom8y-log`: src=22, tests=25 -- unit/, integration/, chaos/ subdirectories
- `autom8y-sendgrid`: src=15, tests=16 -- includes qa/ adversarial
- `autom8y-saga`: src=11, tests=11 -- full parity
- `autom8y-reconciliation`: src=15, tests=14 -- near-parity
- `autom8y-http`: src=24, tests=23 -- circuit breaker, retry, rate limiter, resilience/
- `autom8y-interop`: src=27, tests=24 -- scheduling chaos, backward compat

**Moderate coverage** (50-90% ratio):
- `autom8y-auth`: src=26, tests=21
- `autom8y-cache`: src=36, tests=22 (61%) -- integration/ subdir present
- `autom8y-google`: src=18, tests=15 -- qa/circuit_breaker, degradation
- `autom8y-gcal`: src=17, tests=12 -- qa/ adversarial present
- `autom8y-meta`: src=30, tests=25 -- near-parity
- `autom8y-calendly`: src=12, tests=9

**Weak coverage** (< 50% ratio):
- `autom8y-stripe`: src=27, tests=11 -- **critical gap**: 7 model files have no individual test module
- `autom8y-core`: src=14, tests=7 -- `base_client.py`, `protocols.py` have no dedicated test files
- `autom8y-slack`: src=9, tests=4 -- `models.py` and error paths lightly tested
- `autom8y-events`: src=3, tests=1 -- "POC scope" happy path only; EventBridge error paths untested
- `autom8y-devx-types`: src=5, tests=1 -- partial coverage

### Service Coverage Assessment

13 active services. All have test files.

**Well-covered services**:
- `auth`: src=125, tests=45 -- most-tested service; Hypothesis fuzz, integration/
- `reconcile-spend`: src=23, tests=23 -- full parity + golden traces (4 JSON snapshots); Hypothesis property tests
- `contente-onboarding`: src=18, tests=19 -- saga integration/ with happy path, compensation, idempotency, adversarial

**Moderate coverage**:
- `calendly-intake`: src=34, tests=25 -- e2e/ subdirectory; gap in `app.py`, `clients/factory.py`
- `devconsole`: src=21, tests=19 -- near-parity
- `ads`: src=43, tests=14 -- **significant gap**: `lifecycle/` subdirectory has ZERO test coverage
- `pull-payments`: src=17, tests=14
- `reconcile-ads`: src=15, tests=9

**Light coverage**:
- `auth-mysql-sync`: src=20, tests=7 -- unit/ and integration/ split; requires live DB for integration
- `account-status-recon`: src=16, tests=10 -- `handler.py`, `readiness.py`, `metrics.py` untested
- `sms-performance-report`: src=9, tests=8 -- reasonable
- `validate-business`: src=5, tests=3 -- handler and rules covered
- `slack-alert`: src=3, tests=1 -- handler smoke test only

### Critical Untested Paths

1. **HIGH** -- `autom8y-events` error paths: "happy path only (POC scope)"; EventBridge failures untested
2. **HIGH** -- `ads` lifecycle strategies: `lifecycle/strategies/v2_meta.py` has zero coverage
3. **MEDIUM** -- `autom8y-stripe` models: 7 model files with no test coverage (financial data paths)
4. **MEDIUM** -- `autom8y-core` base_client.py: foundational abstract client untested independently
5. **LOW** -- `account-status-recon` metrics.py: observability blind spot

### Coverage Infrastructure

- Coverage configured per-package in `pyproject.toml` via `[tool.coverage.run]` with `branch = true`
- `[tool.coverage.report]` excludes `pragma: no cover`, `TYPE_CHECKING`, `@abstractmethod`, `raise NotImplementedError`
- `pytest-cov>=4.0` in workspace dev dependencies
- No centralized coverage dashboard or CI gate enforcing minimum thresholds

---

## Testing Conventions

### Test Function Naming

Dominant pattern: class-based tests with `async def test_{description}` methods.
- Class-based (`class Test`): 1,398 instances in SDKs, 794 in services
- Standalone functions (`def test_`): 120 in SDKs, 65 in services

Class names: `TestSubjectBehavior` or `TestSubject`. Methods: `test_{action}_{condition}`.

### Async Test Patterns

All packages use `asyncio_mode = "auto"` in `[tool.pytest.ini_options]`. No `@pytest.mark.asyncio` decorators needed. Tests freely mix `async def test_*` and `def test_*`.

### Assertion Patterns

Pure `pytest` assertion style: native Python `assert` statements. `pytest.raises()` context manager used extensively: 148 SDK files, 70 service files.

### Mock and Stub Patterns

Three distinct strategies:
1. **`respx`** -- for mocking `httpx` HTTP calls
2. **`unittest.mock` (MagicMock, AsyncMock, patch)** -- general mocking
3. **`moto`** -- for mocking AWS services (EventBridge, DynamoDB, SSM)

### Testing Subpackage Pattern

18 of 21 SDK packages expose a `testing/` subpackage with: `factories.py`, `fixtures.py`, `stubs.py` or `mocks.py`.

### conftest.py Patterns

Common patterns:
- `pytest_configure(config)` -- registers custom markers
- `anyio_backend` fixture returning `"asyncio"`
- `autouse=True` reset fixtures
- AWS credential mocking via `monkeypatch.setenv`

### Test Skip Patterns

`pytest.mark.integration` is the primary skip marker. Integration tests isolated in `tests/integration/` subdirectories. No widespread `pytest.mark.skip` or `pytest.mark.xfail`.

### Advanced Test Patterns

- **Adversarial tests** (`test_qa_adversarial_*.py`): present in autom8y-http, autom8y-log, autom8y-sms-test, autom8y-telemetry, autom8y-sendgrid, ads, validate-business, account-status-recon, contente-onboarding, autom8y-gcal, autom8y-saga
- **Contract/compliance tests**: `test_contract_compliance.py`, `test_plugin_contract.py` in autom8y-telemetry
- **Chaos tests**: `autom8y-log/tests/chaos/` and `autom8y-interop/tests/test_scheduling_chaos.py`
- **Property-based tests** (Hypothesis): `reconcile-spend` (3 files + strategies.py); `auth` (`test_openapi_fuzz.py`)
- **Golden trace tests**: `reconcile-spend/tests/golden_traces/` with 4 JSON snapshot scenarios

### Test Data / Fixtures

Python tests use:
- Factory classes (polyfactory pattern in `testing/` subpackages)
- `conftest.py` fixtures
- JSON snapshot files in `reconcile-spend/tests/golden_traces/snapshots/`
- `moto` AWS mocks for Lambda service tests

---

## Test Structure Summary

### Overall Distribution

| Component | Packages | Source Files | Test Files | Ratio |
|-----------|----------|-------------|------------|-------|
| SDKs (`sdks/python/`) | 21 | ~426 | ~263 | 0.62 |
| Services (`services/`) | 13 | ~353 | ~212 | 0.60 |
| Tools | 1 | ~10 | ~3 | 0.30 |

Total: approximately 479 test files targeting ~779 source files.

### Most Heavily Tested Areas

1. **`autom8y-telemetry`** (46 test files, 38 src) -- conventions enforcement, AWS EMF, FastAPI integration, benchmarks
2. **`services/auth`** (45 test files, 125 src) -- OAuth flows, RBAC, tokens, encryption, multitenant isolation, OpenAPI fuzz
3. **`services/reconcile-spend`** (29+ test files, 23 src) -- golden trace snapshot testing; Hypothesis property tests
4. **`autom8y-log`** (25 test files, 22 src) -- multi-tier with unit/, integration/, chaos/
5. **`autom8y-interop`** (24 test files, 27 src) -- cross-service client contracts, scheduling chaos

### Integration vs Unit Test Separation

Separation is **structural**:
- Integration tests live in `tests/integration/` requiring external infrastructure
- Unit tests at `tests/` root or `tests/unit/`
- `pytest.mark.integration` marker for documentation and selective execution
- Services with explicit separation: `auth-mysql-sync`, `auth`, `contente-onboarding`, `autom8y-cache`, `autom8y-log`

### Test Invocation Commands

```bash
# Per-SDK (from sdk-ci.yml)
uv run --package <sdk-name> pytest tests/ -v --tb=short

# Per-service
uv run pytest tests/ -v --tb=short

# Integration tests only
uv run pytest tests/ -m "integration" -v --timeout=<seconds>

# With coverage (satellite-ci-reusable.yml, optional)
uv run pytest tests/ -v --cov=<package> --cov-report=xml --cov-fail-under=80
```

Coverage enforcement (`--cov-fail-under=80`) available via `satellite-ci-reusable.yml` but requires opt-in.

### CI Integration

Tests invoked via satellite-dispatch pipeline (`just sdk-deploy` includes pytest gate). Per-service CI uses satellite-dispatch/satellite-receiver pattern. No centralized test coverage reporting.

---

## Knowledge Gaps

1. **Actual branch coverage percentages unknown** -- branch coverage configured but no reports collected; percentages cannot be stated without running the test suite.

2. **`autom8y-meta` test content** -- 21 test files exist but were not individually examined; testing subpackage structure is absent.

3. **CI coverage gating** -- unclear whether any CI stage enforces minimum coverage threshold; `sdk-deploy` runs pytest with `-x` but no `--cov` flag visible.

4. **`reconcile-ads` and `pull-payments` test depth** -- test files exist but contents not examined.

5. **`tools/ecosystem-observer` CLI and reporters untested** -- `cli.py`, `reporters/json_reporter.py`, `reporters/table_reporter.py` have no test files.
