---
domain: test-coverage
generated_at: "2026-03-01T12:42:56Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "762ed0e"
confidence: 0.88
format_version: "1.0"
---

# Codebase Test Coverage

## Coverage Gaps

### Source Modules Without Direct Test Files

**High Criticality (no dedicated test file):**
- `src/autom8_ads/app.py` -- Application factory and lifespan. `create_app()` exercised indirectly through all API tests, but lifespan startup/teardown (Meta SDK init, state wiring) has no dedicated unit test.
- `src/autom8_ads/dependencies.py` -- DI wiring. `get_config`, `get_platform_adapter`, `get_launch_service`, `verify_jwt` exercised only via override-pattern in API tests. Direct unit tests for dependency resolution (e.g., missing `app.state` attributes) absent.
- `src/autom8_ads/lifecycle/strategies/base.py` -- Strategy base class. No dedicated test; relied upon via `V2MetaLaunchStrategy`.

**Medium Criticality (untested or under-tested):**
- `src/autom8_ads/models/creative.py` -- `AssetRef`, `Creative` models. No test file.
- `src/autom8_ads/models/targeting.py` -- `TargetingSpec`. Used in fixtures but no validation test.
- `src/autom8_ads/models/ad.py`, `models/ad_group.py`, `models/campaign.py` -- No dedicated tests (separate from generic `test_models.py`).
- `src/autom8_ads/routing/config.py` -- Exercised via router tests but not tested for invalid configurations directly.

**Low Criticality:**
- `src/autom8_ads/models/base.py` -- Base Pydantic model. Behavior covered implicitly.

### Critical Path Coverage Assessment

| Critical Path | Covered | Notes |
|---|---|---|
| POST /api/v1/offers/{id}/launch (auth) | Yes | `tests/api/test_launch_endpoint.py` |
| POST /api/v1/offers/{id}/launch (success) | Yes | `tests/api/test_launch_endpoint.py`, `tests/api/test_e2e_launch.py` |
| LaunchService orchestration | Yes | `tests/launch/test_service.py` |
| V2MetaLaunchStrategy 5-step pipeline | Yes | `tests/lifecycle/test_strategy.py` |
| Budget reconciliation (guardrails) | Yes | `tests/lifecycle/test_budget_reconciler.py`, `tests/lifecycle/test_budget_campaign_integration.py` |
| Campaign search / smart matching | Yes | `tests/lifecycle/test_campaign_search.py` |
| Event bus dispatch | Yes | `tests/events/test_bus.py` |
| App lifespan / startup | No | No dedicated test |
| verify_jwt dependency (edge cases) | Partial | Tested via endpoint assertions |

### Test Blind Spots

**Error paths not covered:**
- `app.py` lifespan failure (Meta SDK fails to initialize): no test
- `dependencies.py` `get_platform_adapter` when `app.state.platform_adapter` is absent
- Concurrent duplicate launch requests (race condition in `LaunchIdempotencyCache`)

**Boundary conditions missing:**
- `TargetingSpec` with `age_min > age_max`
- `LaunchIntent` with `daily_budget_cents` at int boundary values
- `AccountRouterConfig` with duplicate `is_default=True` rules

**Integration-level gaps:**
- `tests/integration/test_meta_smoke.py` is skipped by default (`-m 'not integration'`)
- No test for Asana writeback path (stub silently no-ops)
- No test for `DataServiceProtocol` real implementation

### Prioritized Gap List

1. **App lifespan tests** -- high risk: startup failures silent in test suite
2. **`dependencies.py` direct unit tests** -- medium risk: DI wiring bugs only caught at integration level
3. **`TargetingSpec` validation** -- medium risk: bad targeting reaches platform with no domain rejection
4. **`verify_jwt` internals** -- medium risk: security-critical, only pass/fail tested
5. **Concurrent idempotency** -- lower risk: asyncio concurrency not exercised

---

## Testing Conventions

### Test Function / Method Naming

All test functions follow `test_{verb}_{condition}_{expected_outcome}`. Examples:
- `test_full_pipeline_success`
- `test_duplicate_request_returns_cached`
- `test_campaign_created_adset_fails`
- `test_meta_archived_campaigns_excluded`

Tests organized into classes: `Test{ComponentUnderTest}{Scenario}`. Examples:
- `TestLaunchEndpointAuth`, `TestLaunchEndpointValidation`, `TestLaunchEndpointSuccess`
- `TestV2MetaLaunchStrategySuccess`, `TestV2MetaLaunchStrategyPartialFailure`
- `TestBudgetReconcilerGuardrailPass`, `TestBudgetReconcilerGuardrailReject`

Every test class and function has a one-line docstring.

### Fixture / Factory Patterns

**conftest.py fixtures** (`tests/conftest.py`):
- `_isolate_env_from_host` -- session-scoped autouse fixture stripping `AUTOM8Y_ENV` and `ADS_ENVIRONMENT`
- `sample_targeting` -- minimal `TargetingSpec`
- `sample_launch_intent` -- complete `LaunchIntent`
- `sample_launch_result` -- success `LaunchResult`
- `sample_account_router_config` -- multi-rule router config

**Factory module** (`tests/_factories.py`):
- `make_intent(**overrides)` -- builds `LaunchIntent` with keyword overrides
- `make_extensions(**overrides)` -- builds `PlatformExtensions`
- `make_platform_object(obj_id, obj_type, status, name)` -- builds `PlatformAdObject`
- `default_create_side_effect(object_type, *, account_id, params)` -- mock side effect
- `make_mock_platform()` -- full `MagicMock(spec=AdFactory)` with pre-wired methods

**Integration conftest** (`tests/integration/conftest.py`):
- `meta_config` -- session-scoped, skips if token absent
- `real_meta_client` -- async context manager for real `MetaAdsClient`
- `cleanup_tracker` -- tracks created objects for teardown in reverse order

### Assertion Patterns

**Direct assert** (dominant):
```python
assert result.success is True
assert response.status_code == 200
```

**`pytest.raises`** for error paths:
```python
with pytest.raises(AdsValidationError) as exc_info:
    await strategy.execute(ctx, platform)
assert exc_info.value.field == "daily_budget_cents"
```

**`capsys.readouterr()`** for structured log assertions.

**Mock call assertions**: `assert_called_once`, `assert_called_once_with`, `assert_not_called`.

**`asyncio.sleep(0.01)`** in event bus tests to yield control before asserting.

### pytest.mark Usage

- `@pytest.mark.asyncio` -- on every async test; `asyncio_mode = "auto"` in `pyproject.toml`
- `@pytest.mark.integration` -- real API smoke tests; excluded from default run
- No `@pytest.mark.parametrize` observed

### Test Helper Patterns

**Local factory functions** prefixed with `_make_` within individual test files for file-specific setup.

**Dependency override pattern** for FastAPI testing:
```python
app.dependency_overrides[verify_jwt] = _skip_auth
app.dependency_overrides[get_launch_service] = lambda: mock_service
```

**`TestClient` (sync)** for most API tests; **`httpx.AsyncClient` with `ASGITransport`** for full async E2E.

---

## Test Structure Summary

### Overall Distribution

| Source Package | Source Files | Has Tests | Test Location |
|---|---|---|---|
| `api/` | 7 files | Yes (all routers) | `tests/api/` (6 test files) |
| `clients/` | 2 files | Yes | `tests/clients/` (2 test files) |
| `events/` | 2 files | Yes | `tests/events/` (2 test files) |
| `launch/` | 3 files | Yes | `tests/launch/` (3 test files) |
| `lifecycle/` | 7 files | Yes (most) | `tests/lifecycle/` (7 test files) |
| `models/` | 13 files | Partial | `tests/models/` (8 test files) |
| `platforms/` | 5 files | Yes | `tests/platforms/` (4 test files) |
| `routing/` | 2 files | Yes | `tests/routing/` (1 test file) |
| `config.py` | 1 file | Yes | `tests/test_config.py`, `test_config_url_guard.py` |
| `errors.py` | 1 file | Yes | `tests/test_errors.py` |
| `app.py` | 1 file | Indirect only | No dedicated test |
| `dependencies.py` | 1 file | Indirect only | No dedicated test |

**Summary**: ~45 source files; ~35 have direct or indirect test coverage. 2 high-criticality gaps (`app.py`, `dependencies.py`).

### Most Heavily Tested Areas

1. **Lifecycle / Strategy** -- 7 test files, 100+ test cases. V2MetaLaunchStrategy has thorough coverage.
2. **API endpoints** -- all 5 routers covered: auth, validation, error mapping, success tests.
3. **Error hierarchy** -- all 7 error types exhaustively tested.
4. **Event bus** -- 6 scenarios including slow-handler, failing-handler, multi-subscriber.

### Test File Organization

Tests mirror source structure: `src/autom8_ads/{package}/{module}.py` -> `tests/{package}/test_{module}.py`

Exceptions:
- `tests/_factories.py` -- shared factory module
- `tests/test_qa_adversarial.py` -- cross-cutting adversarial tests
- `tests/lifecycle/test_budget_campaign_integration.py` -- integration spanning multiple packages

### How Tests Are Run

**Command**: `pytest -m 'not integration'`

**Configuration** (`pyproject.toml`):
- `asyncio_mode = "auto"`
- `asyncio_default_fixture_loop_scope = "session"`
- `testpaths = ["tests"]`
- `pythonpath = ["src"]`

**Coverage**: Source `src/autom8_ads`, branch coverage enabled, failure threshold 85%.

**Markers**: `integration` -- excluded from default runs; requires `ADS_META_ACCESS_TOKEN`.

---

## Knowledge Gaps

1. `test_qa_adversarial.py` (55KB) was partially read; complete adversarial scenarios not fully enumerated
2. `tests/models/test_models.py` and `test_responses.py` contents inferred from naming
3. Actual coverage percentage not measured; 85% is the configured minimum
4. `tests/platforms/test_translator.py` and `test_meta_params.py` existence inferred but not fully read
5. `tests/lifecycle/test_campaign_matcher.py` and `test_campaign_lock.py` assumed to exist
