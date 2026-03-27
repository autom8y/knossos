---
domain: test-coverage
generated_at: "2026-03-16T00:02:18Z"
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

## Coverage Gaps

### Modules Without Test Files

The following source modules have no dedicated test file:

| Module | Criticality | Notes |
|--------|-------------|-------|
| `src/middleware/security_headers.py` | Medium | No test exercises security header injection on responses |
| `src/middleware/business_scope.py` | High | No direct test; covered obliquely through integration flows but not isolated |
| `src/middleware/rate_limit.py` | High | Exercised indirectly in `test_business_lookup.py` and `test_api_key_hardening.py`, never as the subject |
| `src/utils/brute_force.py` | High | No dedicated test; brute force behavior is partially tested via `test_security.py` rate limit tests |
| `src/observability/logger.py` | Low | SDK-level logging tested in `tests/integration/test_sdk_consumption.py` |
| `src/db/database.py` | Medium | No direct tests; exercised as a dependency through all integration flows |
| `src/config.py` | Medium | Imported by most test modules but never tested in isolation |
| `src/db/seeds/seed_providers.py` | Low | No tests |
| `src/db/seeds/seed_dev_keys.py` | Low | `test_seed_dev_keys.py` covers this |

### Critical Path Coverage

**Well-covered critical paths:**

- **Password auth flow** (`/auth/login`, `/auth/register`, `/auth/me`, `/auth/logout`): covered in `tests/test_integration_flows.py`, `tests/test_auth_ux_integration.py`, `tests/test_security.py`
- **JWT issuance and verification**: dedicated `tests/test_auth_utils.py` (37 test functions), `tests/test_rs256_jwks.py` (25 test functions), `tests/test_token_lifecycle.py` (12 functions)
- **Token revocation** (ADR-0017): `tests/test_revocation.py` with 28 test functions including fail-open Redis behavior
- **API key lifecycle**: `tests/test_api_keys.py`, `tests/test_api_key_hardening.py` (26 functions), `tests/test_admin_api.py` (16 functions)
- **RBAC**: `tests/test_rbac.py` (15 functions), `tests/test_rbac_lifecycle.py` (10 functions), `tests/test_critical_security_fixes.py` (privilege escalation and hybrid role system)
- **Password reset flow**: `tests/test_password_reset.py` (10 functions) — covers valid, expired, invalid, and already-used tokens
- **Multi-tenant isolation**: `tests/test_multitenant_isolation.py` (11 functions), `tests/test_security.py` cross-tenant tests
- **Health endpoints**: `tests/test_health_contract.py` (36 functions) — three-tier health contract
- **Internal routes**: `tests/test_internal_routes.py` (20 functions)

**Under-covered areas:**

1. **`src/middleware/security_headers.py`** — No test verifies that response headers (e.g., `X-Frame-Options`, `Strict-Transport-Security`, `X-Content-Type-Options`) are injected correctly. This is a security-relevant gap.
2. **`src/middleware/business_scope.py`** — Tested only as a side-effect in integration flows; middleware behavior on malformed or missing business headers is not directly tested.
3. **`src/utils/brute_force.py`** — Brute force lockout logic has no isolated test. Rate limiting behavior is tested via `test_business_lookup.py` (lines 166-188) and `test_security.py` (line 111), but the in-memory `BruteForceGuard` object itself is not covered.
4. **OAuth flow (`src/routes/oauth.py` and adapters)** — `tests/test_oauth_flow.py` is **entirely skipped** at module level (`pytestmark = pytest.mark.skip(reason="Vault feature mothballed - see ADR-VAULT-001")`). 5 additional test methods within the file have individual skip annotations. The OAuth callback and authorization redirect paths are untested.
5. **`src/services/credential_vault.py`** — `tests/test_credential_vault.py` has 26 test functions but `pytestmark = pytest.mark.skip(reason="Vault feature mothballed - see ADR-VAULT-001")`. All vault tests are skipped.
6. **`src/services/encryption_service.py`** — `tests/test_encryption_service.py` is **entirely skipped** (`pytestmark = pytest.mark.skip(reason="Vault feature mothballed - see ADR-VAULT-001")`).
7. **`tests/test_credential_access.py`** — Also **entirely skipped** (ADR-VAULT-001).
8. **Token refresh worker error paths** — `tests/test_token_refresh_worker.py` (17 functions) tests the worker, but success-path emphasis; concurrent refresh collision scenarios are not tested.
9. **Charter module** — Multiple test methods in `tests/test_critical_security_fixes.py` are skipped individually (`@pytest.mark.skip(reason="Charter module mothballed per ADR-CHARTER-001")`). The charter audit trigger tests use `pytest.skip()` at runtime when not on live PostgreSQL. `tests/test_charter_integration.py` and `tests/test_charter_versioning.py` appear to be active (not globally skipped).

### Negative / Error Path Tests

Negative test coverage is strong in the active test suite:
- HTTP 401/403/404/409/422 responses are consistently tested (`test_admin_api.py`, `test_rbac.py`, `test_password_reset.py`)
- Token tampering scenarios: `test_security.py:TestTokenTampering`
- SQL injection resilience: `test_security.py:TestInputSanitization`
- Timing attack resistance: `test_security.py:TestTimingAttacks`
- Redis unavailability (fail-open): `test_revocation.py:TestRevocationClient`

**Blind spots**: No negative tests for malformed JWT `kid` values, oversized payloads, or concurrent login burst behavior. OpenAPI fuzz testing (`test_openapi_fuzz.py`) provides 1 Schemathesis-generated conformance test as a trial addition (noted as trial per TDD-SCHEMATHESIS-AUTH-TRIAL).

### Coverage Measurement Infrastructure

`pytest-cov>=4.0` is declared in `[dependency-groups] dev`. The `just test-cov` recipe runs:
```
pytest --cov=src --cov-report=term-missing tests/
```
No `.coveragerc` or `[tool.coverage]` block exists in `pyproject.toml`. No coverage threshold enforcement is configured. No CI coverage gate is documented in the test recipes.

### Prioritized Gap List

1. **HIGH**: `src/middleware/security_headers.py` — security-relevant, zero coverage
2. **HIGH**: `src/utils/brute_force.py` — auth hardening logic, no isolated test
3. **HIGH**: OAuth flow tests permanently disabled (ADR-VAULT-001) — the routes file exists, no tests run
4. **MEDIUM**: `src/middleware/business_scope.py` — no isolated middleware test
5. **MEDIUM**: Vault/encryption suite (3 entire test files skipped) — mothballed per ADR-VAULT-001 but source code is present
6. **MEDIUM**: Coverage threshold not enforced — no gate prevents regression
7. **LOW**: `src/db/seeds/seed_providers.py` — no tests

## Testing Conventions

### Test Function Naming

All test functions use the `test_{action}_{condition}_{expected_result}` pattern, e.g.:
- `test_missing_api_key_returns_401`
- `test_create_role_duplicate_name`
- `test_confirm_password_reset_invalid_token`
- `test_password_validation_timing_constant`

Test classes use `Test{SubjectOrBehavior}` naming (e.g., `TestPasswordHashing`, `TestTokenTampering`, `TestRevocationClient`). 33 test files use class-based organization; a few (e.g., `test_openapi_fuzz.py`) have module-level test functions.

### Fixture Patterns

**`conftest.py`** (`services/auth/tests/conftest.py`) defines the core fixture pyramid:

- `session_fixture` — in-memory SQLite via SQLModel `StaticPool`; function scope
- `client_fixture` — `TestClient` wrapping the FastAPI app with `dependency_overrides` for `get_db` and `get_async_db`
- `business` — seeds a `Business` row via the session
- `user` — seeds a `User` + `UserBusiness` row, depends on `business`
- `admin_role` — seeds a `Role`, depends on `business`
- `read_permission` — seeds a `Permission`
- `_clean_env` — `autouse=True`; uses `monkeypatch` for environment variable isolation between tests

Many test files define their own local fixtures (e.g., `test_multitenant_isolation.py` adds `other_business` and `other_user`; `test_api_key_hardening.py` adds `test_session`, `test_business`, `test_user` with independent SQLite instances).

The `autom8y-auth[testing]` SDK (`RSAKeyPairFixture` and token factories) is available and noted in conftest.py for new tests.

### Assertion Patterns

All tests use native `assert` statements (pytest-style). No `unittest.TestCase` assertions (`assertEqual`, `assertIn`) are used. Common patterns:

- `assert response.status_code == 200`
- `assert "access_token" in response.json()`
- `assert response.json()["detail"] == "..."` for error messages
- `assert hashed.startswith("$argon2id$")` for crypto output format checks

### Mock Patterns

`unittest.mock` is used throughout: `MagicMock`, `AsyncMock`, `patch`. 13 test files use mocking, totaling 384 mock-related occurrences. Common mock targets:
- Redis clients (`mock_redis = AsyncMock()`, injected via direct attribute assignment on the service)
- KMS client in encryption tests (`mock_kms_client`)
- Email service HTTP client (`httpx`)
- Rate limiter (`src.utils.rate_limit.get_rate_limit_client`)

No `pytest-mock` (`mocker` fixture) usage — all mocking uses stdlib `unittest.mock`.

### Test Skip Patterns

Two skip mechanisms are used:

1. **Module-level `pytestmark = pytest.mark.skip(...)`** — disables entire file:
   - `tests/test_encryption_service.py` (ADR-VAULT-001)
   - `tests/test_credential_access.py` (ADR-VAULT-001)
   - `tests/test_oauth_flow.py` (ADR-VAULT-001)
   - `tests/test_credential_vault.py` (ADR-VAULT-001)

2. **Function-level `@pytest.mark.skip(...)`** — disables individual tests:
   - `tests/test_critical_security_fixes.py` — 5 charter-related test functions (ADR-CHARTER-001)
   - `tests/test_revocation.py:230` — 1 test (noted future work)
   - `tests/test_business_lookup.py:211` — 1 test (TestClient async limitation)
   - `tests/test_oauth_flow.py` — 5+ additional individual skips beyond module-level

3. **Runtime `pytest.skip()`** — used inside test body for PostgreSQL-dependent trigger tests (lines 1113, 1125, 1137 of `test_critical_security_fixes.py`)

### Test Data Management

- **In-memory SQLite**: Primary test database. Each test gets a fresh `session` via `StaticPool`. No external database required for unit/integration tests.
- **Static RSA key**: A 2048-bit test private key defined as `TEST_RSA_PRIVATE_KEY` constant in conftest.py. Imported directly by 4 test files.
- **Test UUIDs**: Some tests use hardcoded IDs (`"test-business-1"`, `"test-user-1"`); newer tests use `uuid4()` for uniqueness.
- **No test data directories** — no fixtures from JSON/YAML files; all test data seeded via fixtures.

### Test Environment Management

Environment variables are set in conftest.py using `os.environ.setdefault()` before any `src.*` imports. Pattern is documented with an explanation of why `setdefault` is used over direct assignment (CI env var precedence). Several test files replicate env var setup locally (`test_health_contract.py`, `tests/integration/test_sdk_consumption.py`).

The `integration` pytest marker gates tests that require external infra (DB, Redis). Marker is declared in `pyproject.toml`. No `postgres_required` marker is formally declared but is used in `test_critical_security_fixes.py`.

## Test Structure Summary

### Overall Distribution

| Location | Test Files | Test Functions (approx) |
|----------|------------|-------------------------|
| `tests/` (main) | 34 files | ~688 functions |
| `tests/integration/` | 1 file | 16 functions |
| `client/tests/` | 9 files | 227 functions |
| **Total** | **44 files** | **~931 functions** |

Note: The 4 mothballed test files (`test_encryption_service.py`, `test_credential_access.py`, `test_oauth_flow.py`, `test_credential_vault.py`) account for approximately 89 test functions that are entirely skipped at runtime.

### Most Heavily Tested Areas

1. **`service_key_manager` CLI tool** — `test_service_key_manager.py` with 79 test functions (largest single test file)
2. **Health contract models** — `test_health_contract.py` with 36 test functions across 10 test classes
3. **Auth utils (JWT/password/API key)** — `test_auth_utils.py` with 37 test functions
4. **Charter client** — `test_charter_client.py` with 38 test functions
5. **OAuth state machine** — `test_oauth_state.py` with 23 test functions
6. **Token revocation** — `test_revocation.py` with 28 test functions
7. **Security/attack scenarios** — `test_security.py` with 21 test functions, `test_critical_security_fixes.py` with 28 active functions

### Test Module Organization Patterns

Tests are organized thematically, not mirroring the source tree exactly. There is no `tests/unit/` subdirectory — all non-integration tests live flat in `tests/`. The `tests/integration/` subdirectory contains only one file (`test_sdk_consumption.py`), which tests SDK import and consumption patterns.

The `integration` marker is used inconsistently: some tests marked `integration` run against SQLite in-memory (e.g., `test_multitenant_isolation.py`, `test_rbac.py`), not live PostgreSQL. This means the marker conflates "requires any db" with "requires live infra."

### Integration vs Unit Test Distribution

No hard boundary exists between unit and integration tests. The categorization inferred from markers:

- **Marked `integration`**: `test_rs256_jwks.py`, `test_multitenant_isolation.py`, `test_revocation.py`, `test_load.py`, `test_rbac.py`, `test_token_lifecycle.py`, `test_integration_flows.py`, `test_health_contract.py` (36 tests), `test_business_lookup.py` (no marker), `tests/integration/test_sdk_consumption.py`
- **Unmarked (unit-style)**: `test_auth_utils.py`, `test_auth_ux_integration.py`, `test_password_reset.py`, `test_admin_api.py`, `test_api_key_hardening.py`, `test_audit_logging.py`, `test_service_key_manager.py`, `test_charter_client.py`, `test_seed_dev_keys.py`

### Test Runner and Invocation

**Runner**: `pytest` with asyncio mode `auto` and `asyncio_default_fixture_loop_scope = "function"`.

**Primary invocations** (from `just/test.just`):
- `just test` — all tests: `pytest tests/ -v`
- `just test-int` — integration only: `pytest tests/integration`
- `just test-cov` — with coverage: `pytest --cov=src --cov-report=term-missing tests/`

**Key pytest configuration** (from `pyproject.toml`):
- `testpaths = ["tests"]` — only main `tests/` directory, not `client/tests/`
- `pythonpath = ["src"]` — `src` on sys.path for bare imports
- `--import-mode=importlib`
- `markers = ["integration: requires external infrastructure (DB, Redis, external packages)"]`

The `client/tests/` (9 files, 227 functions) are **not in the default pytest run**. They require explicit invocation.

## Knowledge Gaps

1. **Client SDK test runner**: How `client/tests/` is invoked (separate `pyproject.toml`? separate CI step?) is not visible from the auth service root alone.
2. **CI coverage reporting**: Whether coverage reports are uploaded to a coverage service (Codecov, etc.) is not visible from `just/ci.just` without reading the CI config.
3. **`postgres_required` marker**: Used in `test_critical_security_fixes.py` but not declared in `pyproject.toml`. Whether this is an intentional undeclared marker or an omission is unclear.
4. **Load directory**: `tests/load/` appears in the directory listing (`ls`) but contains `k6_load_test.js`, not Python. k6 is the external load testing tool; how it integrates with CI is not documented in this scope.
5. **ADR-VAULT-001 / ADR-CHARTER-001**: These ADRs govern the mothballed test skips. The decision documents themselves (`services/auth/.ledge/decisions/`) were not read — the reasoning behind mothballing (and whether tests should be removed vs kept-skipped) is a knowledge gap.
