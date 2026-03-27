---
domain: test-coverage
generated_at: "2026-03-27T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "1cfde11"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Uncovered Source Modules (No Direct or Indirect Tests)

The following source modules have **no dedicated test file** and limited indirect coverage:

**High-criticality uncovered modules:**

1. `src/autom8_ads/app.py` -- The application factory (`create_app()`) and lifespan manager (`lifespan()`) are exercised only indirectly. The `lifespan()` function contains critical wiring logic including the H-01 stub-guard validation (`ADS_USE_STUB_DATA_CLIENT must be false in non-LOCAL environments`), `RequestIDMiddleware`, and per-client shutdown sequencing. No dedicated test exercises these paths in isolation. The `_init_campaign_lock()` branch logic (real DynamoDB vs `NullCampaignLock` with warning) has no direct test.

2. `src/autom8_ads/dependencies.py` -- The `get_cleanup_service()`, `get_launch_service()`, and `verify_jwt()` dependency functions are tested indirectly through the API endpoint tests. The error path in `get_cleanup_service()` (when `cleanup_data_client is None` and `use_stub_data_client is False`) has no direct test.

**Low-criticality model gaps (data-only structs, exercised transitively):**

- `src/autom8_ads/models/ad.py` -- `Ad` model used across platform translator and test fixtures but no dedicated tests.
- `src/autom8_ads/models/ad_group.py` -- Same pattern.
- `src/autom8_ads/models/base.py` -- `AdsModel` (frozen Pydantic base), verified through all model tests.
- `src/autom8_ads/models/creative.py` -- No dedicated tests.
- `src/autom8_ads/models/targeting.py` -- `TargetingSpec` used in `conftest.py` fixtures and throughout launch tests but no dedicated test file.
- `src/autom8_ads/models/cache.py` -- `CachedTree` exercised in `test_tree_cache.py` and `test_data_sync.py`.
- `src/autom8_ads/models/search.py` -- `CampaignSearchResult` exercised in `test_campaign_search.py` and `test_campaign_search_activation.py`.

### Protocol-Only Modules

- `src/autom8_ads/cleanup/strategy.py` -- Protocol definition only. Protocol conformance is verified in `test_scoring.py` via `ScoringProtocol`.
- `src/autom8_ads/lifecycle/strategies/base.py` -- Launch strategy protocol. Conformance tested in `test_strategy.py`.

### Integration Test Coverage Gap

Only one integration test file exists: `tests/integration/test_meta_smoke.py`, gated behind `@pytest.mark.integration`. This test requires real Meta API credentials (`ADS_META_ACCESS_TOKEN`) and is **excluded from the default pytest run** (`addopts = "-m 'not integration'"`). The default test suite does not exercise real Meta API calls.

### Coverage Threshold

`pyproject.toml` sets `fail_under = 85` for `pytest-cov`. No recent coverage report is present in the repository, so exact line coverage is unverifiable from static analysis.

### Project Alchemy Test Coverage (Sprint 3-7 additions)

New test files added during the Project Alchemy initiative:

| Test File | Covers |
|---|---|
| `tests/guards/test_protocol.py` | Guard Protocol, GuardResult, GuardChain |
| `tests/guards/test_budget_guard.py` | BudgetGuard |
| `tests/guards/test_targeting_guard.py` | TargetingGuard |
| `tests/guards/test_status_transition_guard.py` | StatusTransitionGuard |
| `tests/guards/test_offer_lifecycle_guard.py` | OfferLifecycleGuard |
| `tests/test_intelligence_data_client.py` | IntelligenceDataProtocol + Stub + HTTP triple |
| `tests/test_mutation_patterns_wiring.py` | MutationPatternAnalyzer DI wiring |
| `tests/test_vertical_policy_wiring.py` | VerticalPolicyEngine DI wiring |
| `tests/test_feature_activation.py` | Feature activation flags and gating |
| `tests/test_sprint5.py` | Sprint 5 override + staging validation + rep visibility |
| `tests/unit/test_offer_lifecycle_mapping.py` | 21-section SECTION_ACTION_MAP correctness |
| `tests/test_creative_bridge.py` | CreativePerformanceBridge |
| `tests/test_intelligence.py` | Intelligence API endpoints |
| `tests/test_asset_client.py` | DataAssetProtocol |
| `tests/test_reconciliation.py` | Reconciliation pipeline |

Estimated passing tests: ~1761 (as of Sprint 7 validation).

### Prioritized Gap List

| Priority | Module | Gap Type | Risk |
|---|---|---|---|
| High | `app.py` / `dependencies.py` | Lifespan/DI wiring logic | H-01 guard and shutdown sequences untested directly |
| Medium | `dependencies.py` `get_cleanup_service()` error path | Error branch | Silent mis-wiring possible |
| Medium | `intelligence/config_normalizer.py` | Unit coverage | 3-era normalization logic only tested transitively |
| Low | `models/ad.py`, `models/ad_group.py`, `models/creative.py` | Data struct | Simple Pydantic models; low regression risk |
| Low | `models/targeting.py` | Data struct | Used transitively throughout launch tests |
| Low | `guards/creative_enrichment.py`, `guards/lifecycle_enrichment.py` | Integration | Enrichment paths tested via wiring tests, not isolated unit tests |
| Info | Integration suite | Real API path | Gated by credentials, expected gap |

---

## Testing Conventions

### Test Function and Class Naming

- Test functions: `test_{behavior_under_test}` -- e.g., `test_local_match_returns_platform_id`, `test_both_searches_empty_returns_none`. All follow `snake_case` with descriptive behavior names.
- Test classes: `Test{Subject}{Scenario}` -- e.g., `TestCampaignSearchLocalMatch`, `TestCampaignSearchMetaFallback`, `TestCampaignSearchLocking`. Classes group related scenarios. Not all test files use classes; many use top-level functions directly (e.g., `test_scoring.py`, `test_config.py`).
- Spec references embedded in docstrings: e.g., `"""SC-004: autom8_data match found"""`, `"""E-12: All Meta campaigns have incompatible names"""`. Requirement codes (SC-NNN, E-NN, FR-N) are systematically embedded in test docstrings.

### Assertion Patterns

Assertion-related calls across test files use three patterns:

1. **Direct `assert`** -- dominant pattern. Used for value equality (`assert result == "local_platform_id"`), boolean checks (`assert result is None`), and type assertions.
2. **`pytest.raises()`** -- used for exception path testing across model validation, error handling, and config guard tests.
3. **Mock call assertions** -- `assert_called_once()`, `assert_not_called()`, `assert_called_once_with()`, `assert_awaited_once_with()` -- heavily used in lifecycle and cleanup tests to verify integration contracts (e.g., `platform.search_objects.assert_not_called()`).

### Fixture and Factory Patterns

**Top-level `tests/conftest.py`** provides:
- `_isolate_env_from_host` -- session-scoped autouse fixture stripping `AUTOM8Y_ENV` to prevent host env leakage.
- `auth_keys` (session-scoped) -- RSA keypair for JWT signing.
- `auth_client_with_mock_transport` -- `AuthClient` wired to mock JWKS transport.
- `make_signed_token` -- factory fixture returning a JWT-creation function.
- `sample_targeting`, `sample_launch_intent`, `sample_launch_result`, `sample_account_router_config` -- domain model fixtures.

**`tests/_factories.py`** provides module-level factory functions (not fixtures):
- `make_intent(**overrides)` -- builds `LaunchIntent` with defaults + overrides.
- `make_extensions(**overrides)` -- builds `PlatformExtensions`.
- `make_platform_object(obj_id, obj_type, ...)` -- builds `PlatformAdObject`.
- `make_mock_platform()` -- returns a `MagicMock` with `AsyncMock` methods pre-configured.

Factories are imported directly into test files via `from tests._factories import make_intent`.

Per-file local helpers are common (e.g., `_make_candidate()`, `_make_compatible_campaign_name()`) for test-specific data construction.

**`tests/integration/conftest.py`** provides session-scoped fixtures tied to live Meta credentials: `real_meta_client`, `meta_adapter`, `cleanup_tracker`.

### Mock and Stub Patterns

- `AsyncMock` and `MagicMock` from `unittest.mock` -- used in 37 of 79 test files.
- `respx` -- HTTP-level mocking library listed in dev dependencies; used sparingly (primary usage is `AsyncMock` for protocol mocks rather than HTTP-level interception).
- `monkeypatch` -- pytest fixture used in config and URL guard tests.
- `StubDataCampaignClient`, `StubCleanupDataClient`, `StubAsanaServiceClient`, `StubAdOptimizationsClient`, `StubDataAssetClient`, `StubIntelligenceDataClient` -- domain stubs from source code, reused in tests without mocking framework.

### Skip Patterns

Only one skip pattern: `pytest.skip()` in `tests/integration/conftest.py` at line 28 -- skips entire integration session when `ADS_META_ACCESS_TOKEN` is absent. No `@pytest.mark.skip` or `@pytest.mark.xfail` annotations found in the test suite.

### Property-Based Testing

`hypothesis` is installed as a dev dependency. Used in exactly 1 test file: `tests/models/test_name_encoding.py`. `@given` decorators test `NameEncoding` roundtrip properties. No other files use hypothesis.

### Test Data Directories

No `fixtures/` or `testdata/` directories found. All test data is constructed programmatically via factory functions and local helper functions.

### Async Testing

`pytest-asyncio` with `asyncio_mode = "auto"` and `asyncio_default_fixture_loop_scope = "session"` -- async tests are first-class citizens. `@pytest.mark.asyncio` is still used explicitly on some tests (found in `test_campaign_search.py`), which is redundant under `asyncio_mode = "auto"` but not incorrect.

---

## Test Structure Summary

### Overall Distribution

| Test Package | Test Files | Source Files Covered |
|---|---|---|
| `tests/api/` | 20 | `api/` (27 source files) + `app.py`, `dependencies.py` indirectly |
| `tests/cleanup/` | 11 | `cleanup/` (6 source files) |
| `tests/lifecycle/` | 8 | `lifecycle/` (7 source files) |
| `tests/models/` | 11 | `models/` (16 source files, ~10 with dedicated tests) |
| `tests/platforms/` | 8 | `platforms/` (7 source files) |
| `tests/clients/` | 7 | `clients/` (8 source files) |
| `tests/launch/` | 4 | `launch/` (3 source files) |
| `tests/cache/` | 2 | `cache/` (1 source file) |
| `tests/events/` | 2 | `events/` (2 source files) |
| `tests/guards/` | 5 | `guards/` (8 source files) |
| `tests/routing/` | 1 | `routing/` (2 source files) |
| `tests/unit/` | 1 | `guards/lifecycle_enrichment.py` (offer lifecycle mapping) |
| `tests/integration/` | 1 | Real Meta API smoke (gated) |
| `tests/` (root) | 17 | `config.py`, `errors.py`, intelligence, reconciliation, feature activation, sprint5 |
| **Total** | **98** | **~95 of 118 source files directly exercised** |

### Most Heavily Tested Areas

1. **`cleanup/`** -- 11 test files, including adversarial tests (`test_qa_adversarial.py`, `test_sprint6_adversarial.py`), contract tests, hardening tests (`test_h01_stub_guard.py`, `test_h02_multi_account.py`), and protocol tests (`test_data_client_contract.py`, `test_error_classification_contract.py`). Cleanup is the most thoroughly tested domain.

2. **`lifecycle/`** -- 8 test files covering search-before-create, budget reconciliation, campaign locking, and integration scenarios. Includes `test_budget_campaign_integration.py` for cross-subsystem behavior.

3. **`models/`** -- 10 test files. Most model files have dedicated tests. Key untested model files are data-container only.

4. **`api/`** -- 9 test files. Full endpoint coverage. Uses `httpx.AsyncClient` in test mode with FastAPI `TestClient`.

5. **`guards/`** -- 5 test files covering budget, targeting, status transition, offer lifecycle, and protocol conformance. Guard chain and enrichment exercised in integration paths.

6. **`intelligence/`** -- Covered by root-level test files: `test_intelligence_data_client.py`, `test_mutation_patterns_wiring.py`, `test_vertical_policy_wiring.py`, `test_creative_bridge.py`, `test_feature_activation.py`, `test_sprint5.py`, `test_intelligence.py`. The `test_offer_lifecycle_mapping.py` in `tests/unit/` covers the 21-section mapping.

### Least Tested Areas

1. **`app.py`** -- Application factory and lifespan. No dedicated tests. Covered indirectly.
2. **`dependencies.py`** -- DI functions. No dedicated tests. Covered indirectly.
3. **`models/` data containers** -- `ad.py`, `ad_group.py`, `creative.py`, `targeting.py` lack dedicated tests.
4. **`intelligence/config_normalizer.py`** -- 3-era config JSON normalization. Covered indirectly via mutation patterns tests.
5. **`reconciliation/`** -- Covered by `test_reconciliation.py` at root level; individual modules lack dedicated unit tests.

### Test Invocation

Default run (excludes integration):
```
uv run pytest tests/ -v --tb=short
```

With coverage:
```
uv run pytest tests/ -v --tb=short --cov=src/autom8_ads --cov-report=term-missing
```

Integration tests (requires real Meta credentials):
```
uv run pytest tests/ -m integration
```

Coverage threshold enforced: 85% (`fail_under = 85` in `pyproject.toml`).

### Test Organization Principles

- **Package mirroring**: `tests/` mirrors `src/autom8_ads/` package structure exactly.
- **Spec-annotated tests**: Requirement codes embedded in docstrings (SC-NNN, E-NN, FR-N) trace tests to specifications.
- **Adversarial test files**: `test_qa_adversarial.py` (both in `tests/` root and `tests/cleanup/`) and `test_sprint6_adversarial.py` represent end-to-end adversarial scenarios added during QA phases.
- **Contract tests**: `test_data_client_contract.py`, `test_error_classification_contract.py`, `test_data_campaign_client_contract.py` verify interface contracts rather than implementation details.
- **Hardening tests**: `test_h01_stub_guard.py`, `test_h02_multi_account.py` named with sprint-hardening prefixes.

---

## Knowledge Gaps

- Exact line/branch coverage percentage is unknown without running `pytest --cov`. Static analysis estimates ~65/79 source files are directly exercised; branch coverage within those files is not determinable statically.
- `lifecycle/budget.py` import coverage: `test_budget.py` exists in `tests/models/` (covers `models/budget.py`), but `lifecycle/budget.py` is covered via `test_budget_reconciler.py` -- this mapping requires confirmation.
- Whether the 85% coverage threshold currently passes is not verifiable without running the test suite.
