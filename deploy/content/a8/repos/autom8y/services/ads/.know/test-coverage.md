---
domain: test-coverage
generated_at: "2026-03-16T00:14:42Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

**Service**: `autom8-ads` -- Ad launch engine for the autom8y platform
**Language**: Python 3.11, pytest + pytest-asyncio
**Test invocation**: `pytest` (configured in `pyproject.toml`, testpaths=`["tests"]`, pythonpath=`["src"]`)
**Total tests collected**: 204

---

## Coverage Gaps

### Untested source modules

The following source modules have no dedicated test file:

| Module | Path | Criticality | Notes |
|--------|------|-------------|-------|
| `AdFactory` | `services/ads/src/autom8_ads/lifecycle/factory.py` | Medium | Thin orchestrator; tested indirectly through `LaunchService` integration tests via mock platform, but no unit tests for the factory itself |
| `V2MetaLaunchStrategy` | `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py` | High | Five-step pipeline is the core execution path; tested indirectly through `LaunchService` and QA adversarial tests, but `_failure_step` static method has no direct unit test; `tests/lifecycle/` directory exists with only `__init__.py` |
| `StubDataServiceClient` | `services/ads/src/autom8_ads/clients/data.py` | Low | Stub logging shim; `tests/clients/` directory exists with only `__init__.py`; intentionally omitted as this is a stub pending Move 4 |
| `dependencies.py` | `services/ads/src/autom8_ads/dependencies.py` | Low | FastAPI DI wiring; two trivial cast functions; coverage via endpoint tests |
| `app.py` | `services/ads/src/autom8_ads/app.py` | Medium | Lifespan logic (singleton init chain) is not directly tested; endpoint tests bypass lifespan by directly setting `app.state`; `_create_default_router_config()` is untested |
| `platforms/protocol.py` | `services/ads/src/autom8_ads/platforms/protocol.py` | Low | Protocol definition; `tests/platforms/` exists with only `__init__.py`; no real platform adapter implementation exists yet |
| `models/base.py` | `services/ads/src/autom8_ads/models/base.py` | Low | Base model configuration; no direct tests |
| `models/targeting.py` | `services/ads/src/autom8_ads/models/targeting.py` | Medium | `TargetingSpec` model is used by mapper; tested only indirectly through mapper tests |
| `models/enums.py` | `services/ads/src/autom8_ads/models/enums.py` | Low | `Platform`, `CampaignObjective` enums; tested implicitly in all test files |
| `routing/config.py` | `services/ads/src/autom8_ads/routing/config.py` | Low | `AccountRouterConfig`/`AccountRule` pydantic models; tested through router tests |

### Blind spots and negative tests

**Documented findings from QA adversarial suite** (already captured in `tests/qa/`):

1. **Empty string `offer_id` accepted** -- `test_empty_string_offer_id_accepted` documents that empty `offer_id` creates a vacuous cache key `":meta"`. No validation at model level.
2. **Cross-field budget validation missing** -- `test_daily_budget_exceeds_weekly` documents no validation that `daily_budget_cents <= weekly_ad_spend_cents`. A caller can set `daily_budget_cents=100000, weekly_ad_spend_cents=1000`.
3. **Zero daily budget from integer division** -- `test_weekly_spend_one_cent_daily_calc` documents that `weekly_ad_spend_cents=1` produces `daily_budget_cents=0` via integer division, which would fail at Meta API level.
4. **`trigger` field accepts arbitrary strings** -- `test_trigger_accepts_arbitrary_value` documents `trigger` has no enum constraint despite PRD specifying `asana | manual | api`.
5. **`LaunchResponse.status` not enum-validated** -- `test_status_no_enum_validation` documents any string passes.
6. **DEFECT-001: `%1D` injection in `vertical_key`** -- `test_vertical_key_with_percent_encoding_attack` detects that `%1D` in `vertical_key` is treated as a filter separator. Severity: LOW (caller is trusted).
7. **DEFECT-002: `_try_persist` parameter type annotation mismatch** -- `TestTryPersistTypeBug` documents that `_try_persist` first parameter is annotated `LaunchResult` but called with `LaunchContext`.
8. **Cache `completed_at=None` leak** -- `test_completed_entry_with_none_completed_at` documents an entry with `status=completed` but `completed_at=None` never expires.
9. **`AdsValidationError` from platform masked as HTTP 200** -- `TestErrorClassification.test_classify_ads_validation_error` documents `V2MetaLaunchStrategy` catches `AdsValidationError` and returns `success=False` with HTTP 200 instead of propagating to the 422 handler.
10. **`test_no_asana_imports_in_source` uses a stale hardcoded worktree path** -- In `tests/qa/test_contract_verification.py` lines 229-252, a hardcoded worktree path points to a worktree that no longer exists. This test will silently pass (finding no imports) when the directory does not exist.

### Prioritized gap list

| Priority | Gap | File(s) |
|----------|-----|---------|
| High | `V2MetaLaunchStrategy` -- no direct unit tests for step sequencing or `_failure_step` logic | `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py` |
| High | Stale hardcoded worktree path in Asana dependency freedom test | `services/ads/tests/qa/test_contract_verification.py:229` |
| Medium | `app.py` lifespan and `_create_default_router_config` untested | `services/ads/src/autom8_ads/app.py` |
| Medium | `TargetingSpec` model has no direct validation tests | `services/ads/src/autom8_ads/models/targeting.py` |
| Medium | Zero daily budget from 1-cent weekly spend not rejected at model layer | Model + mapper layers |
| Low | `StubDataServiceClient` untested (intentional stub) | `services/ads/src/autom8_ads/clients/data.py` |
| Low | No test for `TIKTOK` platform routing or strategy (only META path exists) | Multiple |

---

## Testing Conventions

### Test function naming

Tests follow class-based grouping with descriptive method names:

- Classes: `TestClassName` pattern, e.g., `TestOfferPayload`, `TestLaunchIdempotencyCache`, `TestLaunchServiceHappyPath`
- Methods: `test_<what>_<expected_outcome>` pattern, e.g., `test_empty_string_offer_id_accepted`, `test_validation_error_returns_422`, `test_platform_error_cached_as_failed`
- Every test function has a one-line docstring explaining what it tests

### Assertion patterns

- Direct `assert` statements (no assertion library like `assertpy`)
- Pydantic `ValidationError` tested via `pytest.raises(ValidationError, match=r"pattern")`
- Custom errors tested via `pytest.raises(SomeError)` with attribute inspection: `exc_info.value.offer_id`
- HTTP status codes via `response.status_code == N`
- JSON body via `response.json()["field"]`
- Boolean outcomes: `assert x is True/False` (not `assert x == True`)

### Helper patterns

- `_make_payload(**overrides)` factory functions defined locally per test module (duplicated across `test_launch_endpoint.py`, `test_service.py`, and QA modules -- not shared from conftest)
- `_valid_payload()` functions return dict (pre-Pydantic) for HTTP test clients
- `_make_service(mock_platform, ...)` factory functions assemble full dependency graphs inline

### conftest.py patterns (`services/ads/tests/conftest.py`)

Provides 8 shared fixtures:
- `ads_config` -- `AdsConfig` with `data_writes_enabled=False`
- `account_router` -- `AccountRouter` with one default Meta rule
- `url_builder` -- `MetaUrlBuilder` with production `business_id`
- `mapper` -- `OfferPayloadMapper()`
- `idempotency_cache` -- `LaunchIdempotencyCache` with short TTLs
- `mock_platform` -- `AsyncMock` with pre-set return values (`camp_123`, `adset_456`, etc.)
- `sample_offer_payload` -- fully populated `OfferPayload`
- `sample_launch_result` / `sample_launch_response` -- completed success objects

Fixtures are NOT class-scoped; all default to function scope.

### Async test patterns

- `asyncio_mode = "auto"` in `pyproject.toml` -- no `@pytest.mark.asyncio` decorator needed
- Async test methods are declared `async def test_...` directly
- Async tests live inside sync test classes (pytest-asyncio auto-detects them)
- `asyncio_default_fixture_loop_scope = "function"` -- fresh loop per test

### No fixtures for

- No `@pytest.fixture(scope="module")` or `scope="session"` -- all function-scoped
- No pytest marks (no `@pytest.mark.integration`, `@pytest.mark.slow`, `@pytest.mark.skip` except one `pytest.skip()` inline call in `test_url_builder_adversarial.py`)
- No `pytest.ini` or `setup.cfg` -- configuration is entirely in `pyproject.toml`

### Test skip/environment patterns

One inline skip in `test_url_builder_adversarial.py`:
```python
pytest.skip("DEFECT-001: %1D in vertical_key creates filter injection. Severity: LOW ...")
```
This is conditional -- only skips when the defect is triggered. No environment-gated tests.

### Coverage tooling

- `pytest-cov>=4.0` in dev dependencies
- `[tool.coverage.run]` configured with `source = ["src/autom8_ads"]`, `branch = true`
- Exclusions: `pragma: no cover`, `if TYPE_CHECKING:`, `@abstractmethod`, `raise NotImplementedError`
- Coverage report via `pytest --cov` (not run as part of `pytest` default command)

---

## Test Structure Summary

### Overall test distribution

**204 total tests** across 15 test files:

| File | Test Count (approx) | Focus |
|------|--------------------|----|
| `tests/api/test_launch_endpoint.py` | 11 | HTTP endpoint integration (POST, DELETE, health) |
| `tests/launch/test_idempotency.py` | 10 | `LaunchIdempotencyCache` unit tests |
| `tests/launch/test_mapper.py` | 12 | `OfferPayloadMapper.to_launch_context()` unit tests |
| `tests/launch/test_service.py` | 9 | `LaunchService` pipeline integration tests |
| `tests/models/test_offer.py` | 18 | `OfferPayload` + `LaunchResponse` model validation |
| `tests/routing/test_router.py` | 5 | `AccountRouter` routing logic |
| `tests/test_errors.py` | 7 | `AdsError` hierarchy unit tests |
| `tests/urls/test_meta_url_builder.py` | 25 | `MetaUrlBuilder` URL construction |
| `tests/qa/test_adversarial_payload.py` | ~25 | Adversarial `OfferPayload` boundary/injection tests |
| `tests/qa/test_contract_verification.py` | ~12 | Contract verification against PRD schema |
| `tests/qa/test_endpoint_adversarial.py` | ~20 | Adversarial endpoint tests (path traversal, malformed JSON) |
| `tests/qa/test_idempotency_adversarial.py` | ~18 | Adversarial cache edge cases, TTL, eviction |
| `tests/qa/test_service_adversarial.py` | ~22 | Adversarial service: error classification, logging, URL builder failure |
| `tests/qa/test_url_builder_adversarial.py` | ~20 | Adversarial URL builder injection, encoding, column verification |

### Most heavily tested areas

1. **`MetaUrlBuilder`** -- Most exhaustive. Both a unit test file (`test_meta_url_builder.py`, 25 tests) and a full adversarial file (`test_url_builder_adversarial.py`, ~20 tests) covering URL structure, filter encoding, injection, column constants, and legacy regression.
2. **`LaunchIdempotencyCache`** -- Two files: `test_idempotency.py` (core semantics) and `test_idempotency_adversarial.py` (TTL boundaries, cache key collisions, eviction). Combined ~28 tests.
3. **`OfferPayload` model** -- Two files: `test_offer.py` (validation rules) and `test_adversarial_payload.py` (boundary + injection). Combined ~43 tests.
4. **HTTP endpoints** -- Two files: `test_launch_endpoint.py` (happy path + idempotency) and `test_endpoint_adversarial.py` (security, malformed input, status code semantics). Combined ~31 tests.

### Integration vs. unit test distinction

There are no explicit integration markers. Classification by structure:

| Type | Files | Pattern |
|------|-------|---------|
| Unit | `test_idempotency.py`, `test_mapper.py`, `test_offer.py`, `test_errors.py`, `test_router.py`, `test_meta_url_builder.py` | Tests a single class with direct instantiation, no HTTP |
| Integration | `test_launch_endpoint.py`, `test_service.py` | Tests multiple collaborators wired together; endpoint tests use FastAPI `TestClient` |
| QA / Adversarial | `tests/qa/*.py` | Mix of unit and integration; explicitly adversarial orientation per module docstrings |

All tests run in a single `pytest` invocation with no separation mechanism. No `--integration` flag or marker filter is documented.

### Test invocation command

```
cd services/ads
pytest
```

With coverage:
```
pytest --cov=src/autom8_ads --cov-report=term-missing
```

### Test configuration

Fully in `services/ads/pyproject.toml`:
```toml
[tool.pytest.ini_options]
asyncio_mode = "auto"
asyncio_default_fixture_loop_scope = "function"
testpaths = ["tests"]
pythonpath = ["src"]
```

No `conftest.py` at the repo root -- only at `tests/` level.

---

## Knowledge Gaps

1. **Actual coverage percentage unknown** -- No coverage run output was captured. The `branch = true` config and 204 tests suggest solid coverage of core paths, but actual line/branch figures are not documented here.
2. **No CI/CD test execution log** -- Unknown if tests currently pass clean or have any known failures in CI. The hardcoded worktree path in `test_contract_verification.py` is a suspected silent false-pass.
