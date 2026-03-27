---
domain: test-coverage
generated_at: "2026-03-16T20:09:00Z"
expires_after: "7d"
source_scope:
  - "./*/src/**/*.py"
  - "./*/tests/**/*.py"
  - "./*/pyproject.toml"
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

### Services Without Test Files

None. All 10 active services have test directories with test files. The `_template` service has a `tests/` directory with only a `.gitkeep` sentinel — expected for a scaffold.

### Per-Service Gap Analysis

**account-status-recon** (7 test files, 160 test functions)

Source modules: `handler.py`, `config.py`, `errors.py`, `fetcher.py`, `joiner.py`, `metrics.py`, `models.py`, `orchestrator.py`, `readiness.py`, `report.py`, `rules.py`

Tested: `fetcher`, `instrumentation` (via test_instrumentation.py), `joiner`, `orchestrator`, `rules`, `qa/`
Not directly tested: `handler.py`, `config.py`, `errors.py`, `metrics.py`, `models.py`, `readiness.py`, `report.py`

Gap severity: Moderate. `handler.py` (Lambda entry point), `readiness.py`, and `report.py` are tested in sibling services (reconcile-ads has test_handler, test_readiness, test_report) — this service has not yet added those files. `config.py` and `models.py` are commonly tested only indirectly.

**ads** (14 test files, 204 test functions)

Source modules: `api/health.py`, `api/launch.py`, `app.py`, `clients/data.py`, `config.py`, `dependencies.py`, `errors.py`, `launch/idempotency.py`, `launch/mapper.py`, `launch/service.py`, `lifecycle/factory.py`, `lifecycle/strategies/base.py`, `lifecycle/strategies/v2_meta.py`, `models/*`, `platforms/protocol.py`, `routing/config.py`, `routing/router.py`, `urls/meta.py`

Tested: `api/launch.py`, `launch/idempotency.py`, `launch/mapper.py`, `launch/service.py`, `models/offer.py`, `routing/router.py`, `urls/meta.py`, `errors.py`, plus extensive `qa/` adversarial suite
Not directly tested: `api/health.py`, `app.py`, `clients/data.py`, `config.py`, `dependencies.py`, `lifecycle/factory.py`, `lifecycle/strategies/*.py`, `platforms/protocol.py` (tests/lifecycle/, tests/clients/, tests/platforms/ directories exist but are empty — only `__init__.py`)

Gap severity: Significant. The `lifecycle/` module (factory and strategies) controls how ads are constructed for specific platforms and has zero test files. `clients/data.py` is also untested.

**auth** (44 test files, 704 test functions — largest service)

9 test files are fully skipped at module level (pytestmark = pytest.mark.skip):
- `auth/tests/test_oauth_state.py` — Vault mothballed (ADR-VAULT-001)
- `auth/tests/test_encryption_service.py` — Vault mothballed
- `auth/tests/test_token_refresh_worker.py` — Vault mothballed
- `auth/tests/test_oauth_flow.py` — Vault mothballed
- `auth/tests/test_credential_access.py` — Vault mothballed
- `auth/tests/test_credential_vault.py` — Vault mothballed
- `auth/tests/test_charter_client.py` — Charter mothballed (ADR-CHARTER-001)
- `auth/tests/test_charter_integration.py` — Charter mothballed
- `auth/tests/test_charter_versioning.py` — Charter mothballed

Plus 5 individual methods in `test_critical_security_fixes.py` are skipped for Charter.

Active coverage is strong: `api_keys`, `admin_api`, `audit_logging`, `auth_utils`, `business_lookup`, `email_service`, `health_contract`, `integration_flows`, `internal_routes`, `multitenant_isolation`, `password_reset`, `rbac`, `rbac_lifecycle`, `refresh_token_storage`, `revocation`, `rs256_jwks`, `security`, `seed_dev_keys`, `service_key_manager`, `token_lifecycle`

Not directly tested: `middleware/business_scope.py`, `middleware/rate_limit.py`, `middleware/security_headers.py`, `services/identifier.py`, `services/token_lookup.py`, `utils/brute_force.py`, `utils/rate_limit.py`, `routes/charter.py` (mothballed), `routes/credentials.py` (partially skipped)

**auth-mysql-sync** (7 test files, 116 test functions)

Tested: `sync/auth_writer.py`, `sync/mysql_reader.py`, `sync/orchestrator.py` (integration); `config.py`, `sync/guid_converter.py`, `sync/transformer.py` (unit); SDK integration via `test_sdk_integration.py`

Not directly tested: `main.py`, `observability/logger.py`, `observability/metrics.py`, `portover/cli.py`, `portover/handler.py`

Integration tests require live databases and are gated by `@pytest.mark.integration`.

**devconsole** (19 test files, 428 test functions)

Tested: all major UI modules (conversation_lens, decision_lens, duration_badges, error_boundaries, infrastructure_lens, keyboard_shortcuts, multi_service, mutation_summary, payload_diff, performance_lens, session_tree, session_bookmark, theme, tree_performance); core infrastructure (otlp_receiver, persistence, span_buffer, tempo_client, smoke_helpers)

Not directly tested: `app.py`, `config.py`, `ui/side_effect.py`, `ui/conversation.py`

Gap severity: Low. Most modules have dedicated test files.

**pull-payments** (13 test files, 168 test functions)

Tested: `clients/models.py`, `config.py` (multiple angles including url_guard, security, extension), `clients/data_service.py`, `data_service_circuit_breaker`, `orchestrator.py`, `replay.py`, `staging.py`, instrumentation for data_service/orchestrator/staging

Not directly tested: `handler.py`, `metrics.py`, `models.py`

**reconcile-ads** (9 test files, 194 test functions)

Tested: `handler.py`, `instrumentation`, `joiner.py`, `metrics.py`, `orchestrator.py`, `readiness.py`, `report.py`, `rules.py`, adversarial suite

Not directly tested: `config.py`, `errors.py`, `fetcher.py`, `models.py`

Gap severity: Low. Core business logic paths (orchestrator, rules, readiness, handler, joiner) are all covered.

**reconcile-spend** (21 test files, 407 test functions — most comprehensive service)

Tested: `clients/models.py`, `config.py`, `config_url_guard`, `clients/data_service.py`, `data_service_circuit_breaker`, `defect_remediation`, `enrichment`, `handler.py`, `instrumentation`, `metrics.py`, `orchestrator.py`, `parsing_properties` (hypothesis), `readiness.py`, `readiness_properties` (hypothesis), `report.py`, `rules.py`, `rules_properties` (hypothesis), `three_way_rules`, `three_way_adversarial`, `adversarial_data`, `contract_regression`

Not directly tested: `clients/asana_resolve.py`, `errors.py`, `stubs.py`

This service is the coverage exemplar.

**slack-alert** (1 test file, 1 test function)

Tested: Single smoke test — `test_handler_module_exists()` verifies `lambda_handler` is importable.

Gap severity: High. The entire service is covered by one module-existence assertion. No behavioral tests for the actual Slack notification logic.

**sms-performance-report** (6 test files, 43 test functions)

Tested: `clients/models.py`, `config.py`, `handler.py`, `orchestrator.py`, `readiness.py`, `report.py`

Not directly tested: `clients/data_service.py`

Gap severity: Low. Strong module alignment.

### Prioritized Gap List

1. **slack-alert behavioral coverage** — only a smoke import test exists; Slack formatting, error handling, and alert routing are untested
2. **ads lifecycle/ module** — `lifecycle/factory.py` and `lifecycle/strategies/v2_meta.py` have empty test directories with no test files
3. **auth mothballed test suite** — 9 fully-skipped test files for Vault and Charter features; if/when features revive, tests need re-enabling
4. **account-status-recon missing modules** — `handler.py`, `readiness.py`, `report.py` have no tests (sibling services have these covered)
5. **auth middleware** — `business_scope.py`, `security_headers.py`, `brute_force.py` have no direct test coverage
6. **auth-mysql-sync portover/** — `portover/cli.py` and `portover/handler.py` untested
7. **ads clients/data.py** — external data client untested
8. **pull-payments handler.py** — Lambda entry point untested
9. **reconcile-spend clients/asana_resolve.py** — Asana resolution logic untested

---

## Testing Conventions

### Test Function Naming

Dominant pattern: `test_{noun}_{condition}` or `test_{noun}_{action}` — e.g., `test_full_pipeline_happy_path`, `test_readiness_gate_span_fail_path_sets_staleness`, `test_handler_module_exists`.

Two structural patterns coexist:
- **Class-based grouping** (dominant): `class TestFetchSpan:` containing `def test_*` methods. Counts: 648 test classes, 2,164 class-based methods vs 65 bare top-level test functions. Classes are named for the feature/component being tested.
- **Bare functions** (minor): Used in simple or legacy test files like `slack-alert/tests/test_handler.py` and some `reconcile-spend` tests.

### Fixture Patterns

**conftest.py usage:** 9 of 10 services have at least one `conftest.py`. `devconsole` has no `conftest.py` — fixtures are defined locally within individual test files.

**Key fixture locations:**
- `reconcile-spend/tests/conftest.py` — domain objects (ClientRecord variants), settings, and API response mocks
- `auth/tests/conftest.py` — RSA key pair, SQLite in-memory engine, FastAPI TestClient, business/user/role/permission objects
- `pull-payments/tests/conftest.py` — settings, Stripe mock hierarchy (invoice, subscription, line items, async generator)
- `ads/tests/conftest.py` — config objects, router, url_builder, mapper, idempotency cache, mock platform
- `sms-performance-report/tests/conftest.py` — settings, SMSPerformanceRow instances
- `auth-mysql-sync/tests/unit/conftest.py` — autouse env var setup via monkeypatch, settings cache clearing (references SCAR-011)
- `auth-mysql-sync/tests/integration/conftest.py` — integration-specific fixtures

**Autouse fixtures:** Pattern used broadly for environment isolation:
- `_isolate_env` (reconcile-spend) — clears leaked env vars before every test via `monkeypatch.delenv`
- `_clean_env` (auth) — environment snapshot/restore via monkeypatch autouse
- `_clear_settings` (auth-mysql-sync unit) — clears `@lru_cache` on `get_settings()` before/after each test
- `_sdk_env_vars` (auth-mysql-sync unit) — sets required env vars via monkeypatch autouse

**Settings cache pattern:** Services using pydantic settings with `@lru_cache` call `clear_settings_cache()` in fixture setup and teardown. This is consistent across reconcile-spend, pull-payments, sms-performance-report, and auth-mysql-sync.

**Mock patterns:**
- `unittest.mock.AsyncMock` and `MagicMock` (not pytest-mock) — present in 59 test files (1,887 occurrences)
- `unittest.mock.patch` — used for multi-location patching (e.g., pull-payments patching `get_settings` at 3 import sites simultaneously)
- `AsyncMock` for async clients: Stripe client, HTTP client, data service
- `MagicMock` for sync objects with attribute access (invoice, subscription objects)

### Assertion Patterns

- Standard `assert` statements throughout (2,429 total test functions, each typically containing multiple assert calls)
- `pytest.raises` context manager used in 50 files (191 occurrences) for error path testing
- pytest raises is the primary negative-test mechanism

### Test Skip/Mark Patterns

- `@pytest.mark.integration` — used in `auth-mysql-sync` and `auth` to gate tests requiring live databases/services
- `pytestmark = pytest.mark.skip(reason="...")` — module-level skip with ADR reference (auth service, 9 files). Skip reasons always cite a specific ADR (ADR-VAULT-001, ADR-CHARTER-001)
- `@pytest.mark.asyncio` — explicit per-method mark used in services that mix async and sync tests within one file
- `@pytest.mark.usefixtures("_mock_env")` — class-level fixture attachment pattern in reconcile-spend instrumentation tests
- `@pytest.mark.parametrize` — used in 15 files (31 occurrences) for data-driven tests

### Test Data Management

- **Inline fixtures** in conftest.py returning typed domain objects (ClientRecord, OfferPayload, SMSPerformanceRow, etc.)
- **Helper module** pattern: `reconcile-spend/tests/helpers.py` provides `to_rows()` conversion utility
- **Hypothesis strategies module**: `reconcile-spend/tests/strategies.py` — reusable property-based test strategies (`client_record_strategy()`, `budget_aware_record_strategy()`, `reconciliation_row_strategy()`)
- No database fixture files (`.sql`, `.json`) found — test data is constructed programmatically in fixtures

### Property-Based Testing

Used in two services:
- `reconcile-spend` — 3 property test files: `test_rules_properties.py`, `test_readiness_properties.py`, `test_parsing_properties.py` using Hypothesis with `max_examples=200`
- `auth` — `test_openapi_fuzz.py` uses Hypothesis for OpenAPI schema fuzzing

### Adversarial/QA Test Pattern

Services with explicit `qa/` subdirectories or `*_adversarial.py` files:
- `ads/tests/qa/` — 6 adversarial test files (payload, contract, endpoint, idempotency, service, url_builder)
- `account-status-recon/tests/qa/` — 2 adversarial files (edge_cases, qa_adversary)
- `reconcile-ads/tests/test_adversarial.py` — inline adversarial test file
- `reconcile-spend/tests/test_adversarial_data.py`, `test_three_way_adversarial.py` — inline adversarial files

---

## Test Structure Summary

### Distribution Per Service

| Service | Test Files | Test Functions | Notes |
|---------|-----------|----------------|-------|
| auth | 44 | 704 | Largest; includes 9 fully-skipped files (mothballed features) |
| reconcile-spend | 21 | 407 | Most comprehensive; property-based tests, adversarial suite, helpers/strategies modules |
| devconsole | 19 | 428 | UI-focused; no conftest.py; inline fixtures |
| ads | 14 | 204 | Deep qa/ adversarial suite; lifecycle/ module gap |
| pull-payments | 13 | 168 | Circuit breaker, instrumentation, replay, config security coverage |
| reconcile-ads | 9 | 194 | Well-structured; instrumentation tests prominent |
| auth-mysql-sync | 7 | 116 | Explicit unit/integration split with DB marker gating |
| account-status-recon | 7 | 160 | Strong instrumentation coverage; missing handler/readiness/report |
| sms-performance-report | 6 | 43 | Good module alignment; data_service untested |
| slack-alert | 1 | 1 | Single smoke import test only |
| **Total** | **142** | **2,429** | |

### Most Heavily Tested Areas

1. **auth** — JWT lifecycle, RBAC, API keys, token operations, multitenant isolation, refresh token storage, security hardening, RS256/JWKS
2. **reconcile-spend** — reconciliation rules (unit + property-based), three-way rules, instrumentation spans, circuit breaker, adversarial data
3. **devconsole** — all UI lens components (session tree, conversation lens, decision lens, infrastructure lens), span buffer, OTLP receiver

### Integration vs Unit Test Distinction

Services formally separate integration tests:
- **auth-mysql-sync**: `tests/unit/` and `tests/integration/` subdirectories; integration tests gated with `@pytest.mark.integration`. Integration tests require live MySQL and PostgreSQL instances.
- **auth**: `tests/integration/` subdirectory; multiple tests at top level also use `pytestmark = pytest.mark.integration`.

All other services do not use the unit/integration directory split — tests are functionally grouped (e.g., `tests/api/`, `tests/launch/`, `tests/qa/`) or flat.

Most tests use mocked external dependencies (AsyncMock, MagicMock, monkeypatched settings) and are effectively unit/component tests.

### Test Run Commands

All services use `just` as the task runner. Template defines:

```
just test          # uv run pytest tests/ -v
just test-unit     # uv run pytest tests/unit
just test-int      # uv run pytest tests/integration
just test-cov      # uv run pytest --cov=src --cov-report=term-missing tests/
just test-watch    # uv run ptw -- tests/
```

### Pytest Configuration Baseline (All Services)

All services set `asyncio_mode = "auto"` and `asyncio_default_fixture_loop_scope = "function"`. All use `--import-mode=importlib`, `python_files = "test_*.py"`, `testpaths = ["tests"]`. No service uses `pytest-mock` — the standard library `unittest.mock` is preferred. No service enforces a minimum coverage percentage via `fail_under`.

---

## Knowledge Gaps

1. **Actual line coverage percentages** — No coverage reports were run; all gap analysis is based on file-level presence/absence, not line-level execution data
2. **devconsole `app.py` / `config.py` test coverage** — No conftest.py and no dedicated test files found; indirect coverage through other tests is unknown
3. **auth `middleware/` actual test coverage** — `test_security.py` and `test_auth_ux_integration.py` may exercise middleware indirectly via FastAPI test client, but direct middleware unit tests are absent
4. **reconcile-ads `fetcher.py` coverage** — `account-status-recon` has a `test_fetcher.py`; reconcile-ads does not, though the fetcher may be exercised through orchestrator tests
5. **CI test execution configuration** — No `.github/workflows/` found under `services/`; the satellite pipeline mechanics were not inspected
6. **Integration test execution environment** — Which integration tests pass in CI vs are skipped was not determined
