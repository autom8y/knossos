---
domain: test-coverage
generated_at: "2026-03-16T15:40:19Z"
expires_after: "7d"
source_scope:
  - "./python/**/*.py"
  - "./python/**/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Packages With Strong Coverage (10+ test files, near-complete source mapping)

**autom8y-auth** (19 test files, 18 non-testing source modules): All public surface covered. `_detection.py` and `_compat.py` are internal modules covered indirectly through `test_middleware.py`, `test_resilience.py`, and `test_integration.py`. No meaningful gap.

**autom8y-http** (20 test files, 16 non-testing source modules): Exceptional coverage with dedicated files for circuit breaker, rate limiter, retry, timeout, redaction, sync client, concurrency, escape hatch, reexports, and adversarial edge cases. Resilience subdirectory adds `test_client.py`, `test_client_circuit_groups.py`, `test_config.py`, `test_registry.py`, `test_qa_adversarial_ws6.py`. No meaningful gap.

**autom8y-reconciliation** (12 test files, 12 non-testing source modules): Full coverage including QA adversarial subdirectory for gate, correlator, verdict, and general QA scenarios.

**autom8y-sendgrid** (13 test files, 8 non-testing source modules): Full coverage plus `qa/` adversarial subdirectory for client, models, stats, suppressions, and templates.

**autom8y-telemetry** (32 test files, ~22 non-testing source modules): Exceptional coverage with conventions, fastapi, and benchmarks subdirectories. Covers conventions check, contract compliance, plugin contract, span testing, AI, genai, scheduling, SMS, gcal decorator, BSP config, and more.

**autom8y-log** (18 test files, ~13 non-testing source modules): Strong coverage including backends subdirectory (loguru, structlog, output parity), chaos tests, integration tests (http retry, additional processors), and unit tests for middleware, factory, contextvars, sensitive field filter, positional args compat, stdlib adapter.

**autom8y-interop** (14 test files, 14 non-testing source modules): All subdomains covered: ads (clients, errors), asana (clients, errors), data (clients, errors, helpers, lifecycle, models, protocols, stubs). Note: `test_scheduling.py` and `test_scheduling_chaos.py` test scheduling types that live in `data/models.py` (embedded), not a separate source file.

**autom8y-cache** (18 test files, 19 non-testing source modules): Strong coverage with backends, integration, and unit subdirectories. Covers redis, S3, tiered, memory backends. `test_completeness.py` and `test_testing_subpackage.py` act as coverage gap sentinels.

**autom8y-meta** (21 test files, 27 source modules): Good coverage of handlers (all 9 handlers covered) and models (ad_set_models, campaign_models, enums). **Gap**: `models/account.py`, `models/ad.py`, `models/base.py`, `models/conversion.py`, `models/creative.py`, `models/insights.py`, `models/lead_form.py`, and `models/page.py` have no dedicated test files. These are covered only indirectly through handler tests. File: `python/autom8y-meta/tests/test_coverage_gaps.py` acknowledges this and explicitly documents known gaps.

### Packages With Moderate Gaps

**autom8y-stripe** (9 test files, 20 non-testing source modules): Significant gap.

- Covered: `client.py`, `config.py`, `errors.py`, `rate_limiter.py`, `utils.py`, `handlers/charge.py`, `handlers/invoice.py`, `handlers/subscription.py`, `handlers/refund.py`, `categorization/products.py`, `categorization/verticals.py`.
- **Not covered by dedicated tests**: `models/base.py`, `models/charge.py`, `models/customer.py`, `models/invoice.py`, `models/product.py`, `models/refund.py`, `models/subscription.py`, `models/vertical.py`, `handlers/base.py`. The models directory has 8 source files with zero dedicated test files. `test_client.py` and `test_verticals.py` import two model types (`Vertical`, `VerticalMatch`) incidentally.
- Path: `python/autom8y-stripe/src/autom8y_stripe/models/`

**autom8y-slack** (3 test files, 6 non-testing source modules):

- Covered: `client.py`, `config.py`, `formatter.py`.
- **Not covered**: No dedicated test file for `errors.py`, `models.py`, nor the `testing/` subpackage behavior. Slack has no integration or adversarial tests.
- Path: `python/autom8y-slack/src/autom8y_slack/`

**autom8y-gcal** (9 test files, 11 non-testing source modules):

- Covered: `auth.py`, `channels.py` (via `test_channels.py`), `client.py`, `config.py`, `errors.py`, `events.py`, `freebusy.py`, `models.py`, plus `qa/test_qa_adversarial.py`.
- **Not covered**: `_constants.py`, `_mixin_base.py`, `protocols.py` have no dedicated test files. `_mixin_base.py` is tested indirectly through resource method tests.
- Path: `python/autom8y-gcal/src/autom8y_gcal/`

**autom8y-core** (6 test files, 8 non-testing source modules):

- Covered: `base_client.py`, `client.py`, `config.py`, `errors.py`, `token_manager.py`, `clients/data_service.py`.
- **Not covered**: `protocols.py`, `models/data_service.py` (if models/ data_service is distinct from clients/ data_service). The `test_data_service.py` file exists but may exercise only the `clients/data_service.py` path.
- Path: `python/autom8y-core/src/autom8y_core/`

### Packages With Thin Coverage

**autom8y-events** (1 test file, 2 source modules):

- `test_publisher.py` is marked "Happy path tests only (POC scope)" in its docstring — explicitly limited.
- **Not covered**: Error paths, EventBridge failure handling, malformed events. `event.py` has no test file.
- Path: `python/autom8y-events/tests/test_publisher.py`

**autom8y-devx-types** (1 test file, 4 source modules):

- `test_types.py` covers the main types.
- **Not covered**: `_narrative.py`, `_span.py`, `_version.py` have no dedicated test files.
- Path: `python/autom8y-devx-types/tests/test_types.py`

### Critical Paths Coverage Assessment

| Critical Path | Covered | Notes |
|---------------|---------|-------|
| JWT validation (autom8y-auth) | Yes | `test_claims.py`, `test_token_manager.py`, `test_jwks_concurrency.py` |
| HTTP circuit breaker (autom8y-http) | Yes | `test_circuit_breaker.py`, `resilience/test_client_circuit_groups.py` |
| HTTP retry (autom8y-http) | Yes | `test_retry.py`, `test_qa_adversarial_050.py` |
| Cache tiered resolution (autom8y-cache) | Yes | `test_tiered.py`, `integration/test_tiered_integration.py` |
| OTel span conventions (autom8y-telemetry) | Yes | `test_conventions.py`, `test_contract_compliance.py`, `test_convention_check.py` |
| GCal scheduling operations | Partial | `test_events.py`, `test_freebusy.py` — `_mixin_base.py` indirectly covered |
| Stripe models deserialization | Weak | Models imported incidentally; no structural validation tests |
| EventBridge publish (autom8y-events) | Partial | Happy path only, explicit POC scope annotation |
| Reconciliation verdict/gate | Yes | Dedicated tests + adversarial suite |
| Log backend parity | Yes | `test_backends/test_output_parity.py` |

## Testing Conventions

### Test File Naming

All test files follow the `test_{module}.py` convention (no `*_test.py` style observed). File names generally match source module names: `client.py` → `test_client.py`, `errors.py` → `test_errors.py`, `rate_limiter.py` → `test_rate_limiter.py`.

### Test Class Structure

Tests use pytest class-based organization. Classes are named `class Test{Concept}`:

```python
class TestAutom8yHttpClientInit:
    """Tests for client initialization."""
```

Not universal — some files use standalone `def test_*()` functions without a class wrapper (e.g., `test_publisher.py` in autom8y-events, many tests in autom8y-auth).

### Async Test Pattern

All packages declare `asyncio_mode = "auto"` in `[tool.pytest.ini_options]`, meaning async test functions run without requiring `@pytest.mark.asyncio`. Some older or explicitly annotated tests still use `@pytest.mark.asyncio` explicitly (385 occurrences across 40 files), but the decorator is redundant given the global mode setting.

### Fixture Patterns

Two layers of fixtures exist:

1. **pytest11 entry point plugins**: 14 packages register testing fixtures as pip-installable pytest plugins via `[project.entry-points."pytest11"]`. When a downstream package depends on `autom8y-http`, its fixtures are auto-available. Packages with pytest11 plugins: `autom8y-ai`, `autom8y-auth`, `autom8y-cache`, `autom8y-config`, `autom8y-core`, `autom8y-gcal`, `autom8y-http`, `autom8y-log`, `autom8y-reconciliation`, `autom8y-sendgrid`, `autom8y-slack`, `autom8y-sms-test`, `autom8y-stripe`, `autom8y-telemetry`.

2. **Per-package `conftest.py`**: Every tested package has a `conftest.py` at `tests/conftest.py`. Some are thin (autom8y-telemetry: 8 lines — delegates to pytest11 plugin). Some are substantial (autom8y-auth: 120+ lines with session-scoped RSA keypair, token factories, JWKS document builder).

Pattern for rich conftest: session-scoped key material generated once, per-test fixtures built on top:

```python
_SESSION_KEYPAIR = RSAKeyPairFixture()

@pytest.fixture(scope="session")
def rsa_keypair() -> tuple[rsa.RSAPrivateKey, rsa.RSAPublicKey]:
    return _SESSION_KEYPAIR.private_key, _SESSION_KEYPAIR.public_key
```

### Testing Subpackages

Most packages expose a `testing/` subpackage at `src/{pkg}/testing/`. This contains reusable factories, stubs, mocks, and settings for downstream consumers. The `conftest.py` in the package's own test suite imports from its own `testing/` subpackage — keeping production logic separate from test helpers.

### Mock Strategies

- **unittest.mock** (`AsyncMock`, `MagicMock`, `patch`): Used in 88 test files — the dominant mock strategy.
- **httpx MockTransport / pytest-httpx (`respx`)**: Used in autom8y-http, autom8y-sms-test, and autom8y-sendgrid for HTTP-level transport mocking.
- **moto** (`mock_aws`): Used in autom8y-events for EventBridge stubbing.
- **monkeypatch**: Used for environment variable injection (AWS credentials, env settings).

### Pytest Markers Used

- `@pytest.mark.asyncio`: Explicit async marker (redundant with `asyncio_mode=auto` but present in 40 files).
- `@pytest.mark.parametrize`: Used in interop (protocol enumeration), stripe (categorization), log, and others — 58 occurrences across 14 files.
- `@pytest.mark.integration`: Registered in autom8y-http `conftest.py` but used inconsistently — integration tests in `autom8y-cache` and `autom8y-log` use directory-based separation (`tests/integration/`) not the marker.
- `@pytest.mark.skip`, `@pytest.mark.xfail`: Rare, spotted in chaos tests.

### Adversarial / QA Test Subdirectory Pattern

Several packages have a `tests/qa/` subdirectory for adversarial tests generated by the `qa-adversary` agent:

- `autom8y-gcal/tests/qa/test_qa_adversarial.py`
- `autom8y-reconciliation/tests/qa/` (4 files: correlator, gate, verdict, general)
- `autom8y-sendgrid/tests/qa/` (5 files)

Some adversarial tests are in the root tests directory with `test_adversarial.py` naming (autom8y-http, autom8y-log).

### Test Data / Fixtures Directories

- `autom8y-slack/tests/fixtures/`: Slack webhook payload fixtures.
- `autom8y-stripe/tests/fixtures/`: JSON Stripe webhook payloads (`charge.json`, `customer.json`, `invoice.json`, `refund.json`, `subscription.json`).
- `autom8y-stripe/src/autom8y_stripe/categorization/data/`: CSV/JSON static data files used in categorization tests.

### Coverage Configuration

All packages with `[tool.pytest.ini_options]` also configure `[tool.coverage.run]` with `branch = True` and explicit `source = ["src/{package}"]`. This means branch coverage is the standard, not just line coverage.

## Test Structure Summary

### Package Coverage Matrix

| Package | Test Files | Src Files (non-testing) | Has conftest | Has pytest11 | Has integration tests | Has adversarial tests |
|---------|-----------|------------------------|-------------|-------------|---------------------|---------------------|
| autom8y-ai | 7 | 7 | Yes | Yes | No | No |
| autom8y-auth | 19 | 18 | Yes | Yes | Yes (test_integration.py) | No |
| autom8y-cache | 18 | 19 | Yes | Yes | Yes (tests/integration/) | No |
| autom8y-config | 7 | 8 | Yes | Yes | No | No |
| autom8y-core | 6 | 8 | Yes | Yes | No | No |
| autom8y-devx-types | 1 | 4 | No | No | No | No |
| autom8y-events | 1 | 2 | No | No | No | No |
| autom8y-gcal | 9 | 11 | Yes | Yes | No | Yes (tests/qa/) |
| autom8y-http | 20 | 16 | Yes | Yes | No | Yes (test_adversarial.py) |
| autom8y-interop | 14 | 14 | No | No | No | No |
| autom8y-log | 18 | 13 | Yes | Yes | Yes (tests/integration/) | Yes (test_adversarial.py) |
| autom8y-meta | 21 | 27 | Yes | No | No | Yes (test_qa_adversarial.py) |
| autom8y-reconciliation | 12 | 12 | Yes | Yes | No | Yes (tests/qa/) |
| autom8y-sendgrid | 13 | 8 | Yes | Yes | No | Yes (tests/qa/) |
| autom8y-slack | 3 | 6 | Yes | Yes | No | No |
| autom8y-sms-test | 6 | 6 | Yes | Yes | No | Yes (test_adversarial.py) |
| autom8y-stripe | 9 | 20 | Yes | Yes | No | No |
| autom8y-telemetry | 32 | 22 | Yes | Yes | No | No |

Total: 216 test files across 18 packages (all have at least 1 test file).

### Test Directory Organization Patterns

Three structural patterns observed:

1. **Flat**: All tests in `tests/` root (autom8y-ai, autom8y-config, autom8y-events).
2. **Subdirectory by concern**: `tests/integration/`, `tests/backends/`, `tests/resilience/`, `tests/unit/`, `tests/chaos/`, `tests/fastapi/`, `tests/benchmarks/`, `tests/conventions/`, `tests/qa/`, `tests/clients/`, `tests/handlers/`, `tests/models/`, `tests/categorization/`.
3. **Mixed**: Root-level tests plus one or more subdirectories.

### Test Runner Commands

Per-package test execution uses `uv run pytest` from the package directory. No monorepo-level test runner script exists. Each package is run independently:

```
cd python/autom8y-http && uv run pytest
cd python/autom8y-auth && uv run pytest
```

Standard pytest options apply: `-v` for verbose, `--cov` for coverage, `-x` to stop on first failure.

The `devx-types` package does not have `[tool.pytest.ini_options]` in its `pyproject.toml`, making it the only package without explicit pytest configuration.

### Integration vs Unit Distinction

The monorepo uses directory-based rather than marker-based separation for integration tests:

- **Unit tests**: `tests/` root level (mocked dependencies, fast).
- **Integration tests**: `tests/integration/` subdirectory (external service dependencies).

Packages with real integration test directories:
- `autom8y-cache/tests/integration/`: Tests against Redis (`test_redis_integration.py`) and S3 (`test_s3_integration.py`) — require live infrastructure.
- `autom8y-log/tests/integration/`: Tests HTTP retry behavior (`test_http_retry_integration.py`), additional processors (`test_additional_processors.py`).
- `autom8y-auth/tests/test_integration.py`: End-to-end token acquisition and validation.

The `@pytest.mark.integration` marker is registered only in `autom8y-http/tests/conftest.py` but the integration tests in cache and log use directory structure instead. There is no consistent mechanism to run only integration tests across the monorepo.

### Special Package Notes

- **autom8y-sms-test**: This is itself a testing library (provides `MockTwilioTransport` and SMS assertion helpers). Its own tests validate that the testing library works correctly — tests-of-tests.
- **autom8y-telemetry**: Has benchmarks at `tests/benchmarks/` using functional test patterns to validate OTel span operation performance.
- **autom8y-devx-types**: Missing pytest configuration entirely; runs by discovery alone.
- **autom8y-meta**: Has `test_coverage_gaps.py` — a deliberate test file that documents known untested models, serving as a formal coverage gap acknowledgement.

## Knowledge Gaps

1. **Actual test counts per function**: Grep-based discovery cannot report the number of `def test_*` functions within class-based test files without reading every file. The per-file counts are approximate for class-based tests.
2. **Branch coverage percentages**: No `.coverage` artifacts or coverage report files are in scope. Actual coverage percentages are unknown — only structural coverage (which source modules have test files) was assessed.
3. **Interop scheduling source**: `test_scheduling.py` and `test_scheduling_chaos.py` exist in autom8y-interop but no `scheduling.py` source module was found. Scheduling types appear to be embedded in `data/models.py`. The exact boundary was not confirmed.
4. **Whether integration tests pass in CI**: Whether the Redis and S3 integration tests are skipped in CI or require infrastructure is not documented in the test files themselves.
5. **`autom8y-meta/models/` indirect coverage depth**: Handler tests import model classes but the depth of model-layer validation through handler tests was not measured.
