---
domain: test-coverage
generated_at: "2026-03-23T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4febf1f"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

**Project**: autom8y-meta — Async Meta Graph API client SDK
**Language**: Python 3.12+
**Test runner**: `pytest` with `pytest-asyncio` (`asyncio_mode = "auto"`)
**Coverage tool**: `pytest --cov` (`pytest-cov`, branch coverage configured)
**Test data**: No `fixtures/` or `testdata/` directories — all test data is inline or constructed via fixtures in `conftest.py`

## Coverage Gaps

### Untested Source Modules

The following source modules have no dedicated test file and no direct import in the test suite:

| Source Module | Lines (est.) | Risk Level | Notes |
|---|---|---|---|
| `src/autom8y_meta/models/account.py` | ~18 | Low | `AdAccount` model — only 7 fields, all optional except `id`. Not imported in any test file. |
| `src/autom8y_meta/models/base.py` | ~22 | Low | `MetaModel` base class. Exercised indirectly through every model test; no dedicated structural test. |
| `src/autom8y_meta/handlers/base.py` | ~200 (est.) | **Critical** | `BaseHandler._request` retry logic, `_classify_error`, `_handle_response`, `_parse_retry_after`, `_parse_usage_header` — **these ARE covered** via `test_errors.py`, `test_coverage_gaps.py`, and `test_qa_adversarial.py`. Not a gap despite having no dedicated file. |

### Model Layer Gaps

Three model source files have no dedicated test file in `tests/models/`:

- `src/autom8y_meta/models/ad.py` — `Ad`, `AdCreateParams`, `AdUpdateParams` tested only via handler/client delegation tests and `test_coverage_gaps.py`, not a standalone model test.
- `src/autom8y_meta/models/conversion.py` — `ConversionEventPayload` tested in `test_coverage_gaps.py` and `test_qa_adversarial.py` but no dedicated model file.
- `src/autom8y_meta/models/creative.py` — `CreativeSpec`, `AssetUploadResult` tested via handler tests and `test_coverage_gaps.py`, not isolated.
- `src/autom8y_meta/models/insights.py` — `InsightsRow`, `ReportRunStatus` imported in `test_qa_adversarial.py` but no dedicated model tests.
- `src/autom8y_meta/models/lead_form.py` — tested in handler and coverage-gap files.
- `src/autom8y_meta/models/page.py` — tested via page handler tests only.
- `src/autom8y_meta/models/account.py` — `AdAccount` model has **zero test coverage** anywhere in the test suite.

### Negative / Edge-Case Blind Spots

- `AdAccount` model: zero coverage — fields, validation, and `MetaModel` config behavior entirely untested.
- Insights async polling failure path (timeout/job-failed): covered in `test_insights.py` (`test_get_insights_async_timeout`, `test_get_insights_async_job_failed`).
- Config env-var loading from environment: `test_coverage_gaps.py::TestClientConfigFallback.test_no_config_raises_config_error` covers the failure path but no test exercises successful env-var loading.
- `handlers/base.py` `_request` — HTTP network errors (connection timeout, DNS failure) not tested; only API-level error responses are tested.
- Concurrent rate limiter behavior: `test_rate_limiter.py` tests sequential calls but no concurrent stress paths.

### Prioritized Gap List

1. **High**: `AdAccount` model (`models/account.py`) — zero test coverage for a response type that any handler returning account data would use.
2. **Medium**: Env-var-based config loading success path — only the failure path is tested.
3. **Medium**: `handlers/base.py` HTTP transport-level errors (connection reset, timeout at socket level).
4. **Low**: Dedicated model tests for `ad.py`, `creative.py`, `insights.py`, `lead_form.py`, `page.py` — currently tested indirectly.

## Testing Conventions

### Test Function Naming

All tests follow `test_{behavior_description}` naming with explicit docstrings on most test methods. Examples:
- `test_retry_on_rate_limit` — describes the stimulus and expected behavior
- `test_context_manager` — describes the interface being tested
- `test_classify_rate_limit_boundary_80000` — boundary case encoded in name

### Test Class Patterns

Tests are organized into classes using `class Test{Subject}:` pattern. No bare test functions at module level — every test belongs to a class. Examples:
- `TestMetaRateLimiter`, `TestAppSecretProofGenerator`, `TestErrorHierarchy`
- `TestClientLifecycle`, `TestCampaignCRUD`, `TestAdSetCRUD`
- `TestBaseHandlerRetry`, `TestClientDelegationGaps`, `TestModelOptionalFields`

The `test_coverage_gaps.py` and `test_qa_adversarial.py` files serve explicit gap-closing roles and document their purpose in module-level docstrings.

### Assertion Patterns

All assertions use bare `assert` statements (no assertion library). `pytest.raises()` used for exception testing. No `pytest.approx`. Examples:
- `assert result == {"data": "ok"}`
- `assert isinstance(error, MetaRateLimitError)`
- `with pytest.raises(MetaAPIError, match="Bad request"):`

### Async Test Pattern

`asyncio_mode = "auto"` is set globally — async test methods require no `@pytest.mark.asyncio` decorator. All async tests are plain `async def test_*` methods inside classes.

### HTTP Mocking Pattern

All HTTP interactions use `respx` (not `unittest.mock` or `httpx.MockTransport`). The `@respx.mock` decorator is applied at the method level. No module-level `respx.mock` context. Routes are set up per-test with `respx.get(url).mock(return_value=...)` or `route.side_effect = [...]` for multi-call sequences.

### Fixture Patterns

Central fixtures live in `tests/conftest.py`. Six fixtures defined:
- `meta_config` — `MetaConfig` with test secrets
- `meta_accounts` — `list[MetaAccountConfig]` with one entry
- `meta_client` — full `MetaAdsClient` async context manager (async fixture yielding client)
- `proof_generator` — `AppSecretProofGenerator` built from `meta_config`
- `rate_limiter` — `MetaRateLimiter` with high rate limit for test speed
- `http_client` — raw `httpx.AsyncClient` for direct handler testing

Some test files define local fixtures (`@pytest.fixture` in the test class or at module scope) for handler-specific setup (e.g., `InsightsHandler` in `test_insights.py`, `BaseHandler` in `test_errors.py` and `test_qa_adversarial.py`).

### Test Data Management

No fixture files, no `testdata/` directory, no factory libraries (no `polyfactory`). All test data is constructed inline in test methods or via conftest fixtures. JSON payloads are literal dicts in `respx.mock` setup.

Builder helpers appear in `test_campaigns_tree.py` only: `_make_ad()`, `_make_adset()`, `_make_campaign()` module-level functions produce raw Meta API response dicts for hierarchy tests.

### Skip Patterns

No `pytest.skip`, `@pytest.mark.skip`, or `skipIf` usage anywhere in the test suite. All 21 test files run unconditionally.

### Integration Test Signals

No `@pytest.mark.integration` markers, no separate integration test files, no `*.integration.test.*` pattern. The entire suite is unit-level with mocked HTTP.

### Coverage Exclusions

Configured in `pyproject.toml`:
```
[tool.coverage.report]
exclude_lines = ["pragma: no cover", "if TYPE_CHECKING:", "@abstractmethod", "raise NotImplementedError"]
```
Branch coverage is enabled (`branch = true`).

## Test Structure Summary

### Distribution

| Location | Files | Test Classes | Approximate Test Count |
|---|---|---|---|
| `tests/` (root) | 7 | 23 | ~163 |
| `tests/handlers/` | 10 | 14 | ~97 |
| `tests/models/` | 3 | 8 | ~24 |
| `tests/` (conftest) | 1 | — | 6 fixtures |
| **Total** | **21** | **45** | **~305** |

### Most Heavily Tested Areas

1. **`test_qa_adversarial.py`** — 98 test methods; deepest file by count. Covers error classification boundary conditions, adversarial model validation, concurrent rate limiter behavior, pagination edge cases, and malformed response handling.
2. **`tests/handlers/test_campaigns_tree.py`** — 31 test methods organized into 5 classes. Covers the full `get_account_campaigns_tree()` nested expansion feature including 5-page pagination simulation with 112 campaigns.
3. **`tests/test_client.py`** — 29 test methods across 10 classes covering the full `MetaAdsClient` public API surface.
4. **`tests/test_errors.py`** — 26 test methods for error hierarchy, classification, and parsing.
5. **`tests/test_coverage_gaps.py`** — 24 test methods explicitly targeting retry logic, delegation gaps, optional model fields, effective_status parameters, and rate limiter exception paths.

### Test Package Naming

No `tests/` `__init__.py` at root. `tests/handlers/__init__.py` and `tests/models/__init__.py` are not present. Test collection relies on `testpaths = ["tests"]` in `pyproject.toml`.

### Unit vs Integration

100% unit tests with mocked HTTP via `respx`. No integration tests against live Meta Graph API endpoints. The test suite is self-contained with no external service dependencies.

### How Tests Are Run

```
pytest                          # all tests
pytest --cov src/autom8y_meta   # with coverage (branch mode)
pytest tests/handlers/          # handler tests only
pytest tests/models/            # model tests only
```

`asyncio_mode = "auto"` means no explicit async runner invocation needed. `pythonpath = ["src"]` means `autom8y_meta` is importable without install.

## Knowledge Gaps

- **Actual branch coverage percentage unknown**: No coverage run was executed. The `pyproject.toml` configures `pytest-cov` with branch coverage, but the numeric coverage figure is not available without running `pytest --cov`.
- **`AdAccount` model test presence**: Confirmed absent via grep; risk is low given simple model structure.
- **Transport-level error behavior**: Unknown whether `BaseHandler._request` handles `httpx.ConnectError`, `httpx.ReadTimeout`, etc. — no tests for these paths exist.
- **Env-var config success path**: Testing that `MetaConfig` loads correctly from `META_APP_ID` / `META_APP_SECRET` / `META_ACCESS_TOKEN` env vars is not covered.
