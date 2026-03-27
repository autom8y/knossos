---
domain: scar-tissue
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

# Codebase Scar Tissue

## Failure Catalog Completeness

This section catalogs identified failures, regressions, and bugs across the auth service git history (78 fix-tagged commits), code markers, and test file naming evidence.

### SCAR-001: Login Endpoint Timeout via Request Body Consumption

**What failed**: The `rate_limit_login` middleware called `await request.body()` to extract the user's email for per-email rate limiting. This consumed the ASGI request body stream, causing the downstream login endpoint to hang indefinitely waiting for body data that was already consumed.

**When**: Commit `e638524` (2025-12-27), tagged `CRITICAL`.

**How fixed**: Changed the rate-limiting key from email address to client IP address. The `request.body()` call was removed entirely. Brute-force per-user protection remains in the login endpoint itself via `BruteForceProtection`.

**Marker**: In-source comment at `services/auth/src/middleware/rate_limit.py` lines 121–129: `# Use client IP for rate limiting (don't consume request body!)` / `# Reading request.body() would consume the stream and break the login endpoint.`

### SCAR-002: NameError in Service-Token Endpoint on Successful Exchange

**What failed**: The `exchange_service_token` endpoint in `routes/internal.py` referenced a variable `scope` in a logger call, but `scope` was only defined inside a nested helper `_create_service_token()`. Any successful service-token exchange returned HTTP 500.

**When**: Commit `495ae13`.

**How fixed**: Defined the `scope` variable in the outer function before the logger call in `services/auth/src/routes/internal.py` line ~375.

**Marker**: Regression test class `TestServiceTokenScopeRegression` in `services/auth/tests/test_internal_routes.py` line 599, with docstring: `"Regression: successful key exchange must not 500 due to scope NameError."`

### SCAR-003: Double-Revocation Bug in Service Key Deletion

**What failed**: The `revoke_service_key` endpoint queried `APIKey` without filtering `revoked_at IS NULL`. A second revocation attempt would find the already-revoked key (with a matching name) and return HTTP 409 "already revoked" — preventing a create → revoke → create → revoke cycle from working.

**When**: Commit `7f9b102`.

**How fixed**: Added `APIKey.revoked_at.is_(None)` filter to both UUID and name lookup paths in `services/auth/src/routes/admin.py` lines 295, 400, 478, 601, 610. Changed the 409 "already revoked" response to 404 "not found", consistent with soft-delete semantics.

**Marker**: `revoked_at.is_(None)` guard appears at five separate lookup sites in `services/auth/src/routes/admin.py`.

### SCAR-004: Phantom Import in `get_audit_logs` Route

**What failed**: `routes/auth.py` lazy-imported `extract_token_from_header` from `src.routes.rbac`, a module that does not export this function. The `get_audit_logs` code path would raise `ImportError` at runtime.

**When**: Commit `ce6f967`.

**How fixed**: Import redirected to the correct source `src.auth.jwt_handler` in `services/auth/src/routes/auth.py` line ~1172.

**Marker**: Commit message text: "The phantom import would raise ImportError if the get_audit_logs code path was executed."

### SCAR-005: `email/` Package Shadowing Python stdlib `email` Module

**What failed**: The `src/email/` package shadowed Python's built-in `email` module when `PYTHONPATH` included `services/auth/src/`. This caused `importlib.metadata` (used by `autom8y-config`) to fail with a circular import at startup, preventing the service from starting.

**When**: Commit `26f40f9`.

**How fixed**: Renamed `src/email/` to `src/mailer/` with a 4-file blast radius (directory rename, `routes/auth.py` lazy import site, `tests/test_email_service.py` imports and patch targets). Spike doc written: `docs/spikes/SPIKE-AUTH-EMAIL-STDLIB-SHADOW.md`.

**Marker**: Package is now `src/mailer/` throughout. `routes/auth.py` line 986: `from src.mailer import get_email_service`.

### SCAR-006: Circular Import in `auth` Package (`dependencies.py` → `src.auth`)

**What failed**: `dependencies.py` imported from `src.auth` (the barrel package `__init__.py`) while `src.auth` was still initializing its own imports. This caused `ImportError` on partial module load at startup.

**When**: Commit `3f12ef9`.

**How fixed**: `dependencies.py` now imports directly from `src.auth.jwt_handler` (not from the barrel). The barrel `__init__.py` documents the pattern with an explicit comment explaining the cycle was broken.

**Marker**: `services/auth/src/auth/__init__.py` lines 7–11: barrel re-export documentation comment explaining the circular dependency fix.

### SCAR-007: Connection Pool Defeat — New Engine Per Request

**What failed**: `get_db()` created a new `SQLAlchemy` engine on every request, defeating connection pooling entirely. Under concurrent load this caused database connection flooding, 10–11 second request delays, and 401 errors instead of proper 429 rate-limit responses.

**When**: Commit `38673aa`.

**How fixed**: Engines are now created once at module level in `services/auth/src/db/database.py` lines 35–38 and reused across all requests. Pool settings: `pool_size=10`, `max_overflow=20`, `pool_timeout=30`, `pool_pre_ping=True`.

**Marker**: `services/auth/src/db/database.py` lines 11–14: module docstring section "Connection Pooling: Engines are created ONCE at module level and reused across all requests. This is critical for performance under concurrent load."

### SCAR-008: Missing Admin Router Registration (All `/internal/admin/*` Routes 404)

**What failed**: `admin.py` module was created with a full router but never registered in `main.py`. All `/internal/admin/*` endpoints returned 404 in production.

**When**: Commit `e0640ea`.

**How fixed**: Added `admin` to the import list in `services/auth/src/main.py` line 15 router registration.

**Marker**: Commit message: "The admin.py module existed but its router was never registered in main.py, causing all /internal/admin/* endpoints to return 404."

### SCAR-009: Missing Admin Schemas Module (ImportError at Startup)

**What failed**: `routes/admin.py` imported `CreateServiceKeyRequest`, `CreateServiceKeyResponse`, `ServiceKeyInfo`, `RotateServiceKeyResponse`, `AdminAPIError` from `src.schemas.admin`, but the file `src/schemas/admin.py` was never committed, causing `ImportError` at startup.

**When**: Commit `a423b2e`.

**How fixed**: `services/auth/src/schemas/admin.py` was created.

### SCAR-010: `get_settings()` Re-Parsing Env Vars on Every Call

**What failed**: `get_settings()` created a new `Settings()` instance on every call, re-parsing all environment variables each time. This was inefficient and caused inconsistent behavior in tests that mutated environment state.

**When**: Commit `48d2ad1` (hygiene finding FN-010).

**How fixed**: Added `@lru_cache` decorator to `get_settings()` in `services/auth/src/config.py` line 174. Note: tests that modify environment must call `get_settings.cache_clear()`.

### SCAR-011: Database Session Leak from Manual Session Creation in `list_roles`

**What failed**: `routes/charter.py` `list_roles` endpoint created a session manually via `next(get_db())` instead of using FastAPI dependency injection. FastAPI could not manage the session lifecycle, creating a session leak risk.

**When**: Commit `48d2ad1` (hygiene finding FN-011).

**How fixed**: `list_roles` now uses `db: Session = Depends(get_db)` parameter in `services/auth/src/routes/charter.py`.

### SCAR-012: User Enumeration via Timing Attack on Login (HIGH-001)

**What failed**: Login attempts for non-existent users returned in ~5ms (no password hash comparison), while attempts with wrong passwords took ~100ms (Argon2id). This timing difference allowed attackers to enumerate valid email addresses.

**When**: Commit `23e9bfc` (security audit finding HIGH-001).

**How fixed**: A `_DUMMY_HASH` is computed once at module load in `services/auth/src/routes/auth.py` line 70. When a user is not found, `verify_password(_DUMMY_HASH, request.password)` is still called, normalizing response time to ~100ms regardless.

**Marker**: `services/auth/src/routes/auth.py` lines 65–70 comment block. Regression tests: `TestTimingAttacks` class in `services/auth/tests/test_security.py` line 259.

### SCAR-013: Refresh Token Hash Validation Bypass via `business_id` Claim Tampering (HIGH-002)

**What failed**: Refresh token validation queried by `token_hash` only, without verifying the `business_id` claim. An attacker with a valid token could tamper the `business_id` claim to access another tenant.

**When**: Commit `23e9bfc` (security audit finding HIGH-002).

**How fixed**: Query now finds candidate tokens by `user_id`/`business_id` from the JWT claims, then cryptographically verifies the token hash. Fix at `services/auth/src/routes/auth.py` line 721: `# SECURITY FIX (HIGH-002): Cryptographically verify token hash`.

### SCAR-014: Password Reset Token O(n) Table Scan DoS Vector (HIGH-003)

**What failed**: Password reset token lookup iterated the full `password_reset_tokens` table with an O(n) scan, hashing each row's token until a match was found. This was a DoS vector at scale.

**When**: Commit `23e9bfc` (HIGH-003). Migration `015` added in commit `555486f`.

**How fixed**: Added `token_prefix` column (first 12 chars of plaintext token) to `password_reset_tokens` table with an index. Reset verification now does O(1) prefix lookup in `services/auth/src/routes/auth.py` line 1043. Model field annotated: `services/auth/src/models/password_reset_token.py` line 43.

**Marker**: Migration `015_add_token_prefix_to_password_reset.py`.

### SCAR-015: CORS Wildcard Origins Permitting Any Subdomain

**What failed**: CORS was configured with wildcard domains (`https://*.autom8y.io`), which in some browsers allows more origins than intended and is a security misconfiguration.

**When**: Commit `65eb859` (security audit finding MEDIUM-002).

**How fixed**: CORS origins replaced with an explicit allowlist in `services/auth/src/config.py` lines 73–82: `autom8y.io`, `app.autom8y.io`, `auth.autom8y.io`, `api.autom8y.io`, `docs.autom8y.io`, plus `localhost:3000` in local only.

### SCAR-016: `datetime.utcnow()` Timezone-Naive Datetime Returning Incorrect Values (SEC-001)

**What failed**: 220 instances of `datetime.utcnow()` and `datetime.utcfromtimestamp()` throughout the service and SDK were deprecated and timezone-naive, producing incorrect comparisons with timezone-aware PostgreSQL `TIMESTAMPTZ` columns.

**When**: Commit `a7f086c` (SEC-001).

**How fixed**: All callsites replaced with `datetime.now(timezone.utc)` and `datetime.fromtimestamp(ts, tz=timezone.utc)` throughout `services/auth/src/`.

### SCAR-017: TIMESTAMP vs. TIMESTAMPTZ Mismatch Causing `asyncpg` TypeErrors (ISS-5)

**What failed**: Postgres schema columns were `TIMESTAMP WITHOUT TIME ZONE`. The codebase correctly used `datetime.now(UTC)` (offset-aware). `asyncpg` rejected offset-aware Python datetimes for `TIMESTAMP WITHOUT TIME ZONE` columns, causing runtime TypeErrors on async DB operations.

**When**: Commit `28de6a8` (ISS-5), migration added 2026-03-13.

**How fixed**: Migration `016_timestamps_to_timestamptz.py` converts all `TIMESTAMP` columns to `TIMESTAMP WITH TIME ZONE` across all tables.

### SCAR-018: Redis Dependency Missing, Causing Rate Limiting and Token Revocation Failures (ISS-8)

**What failed**: The `redis` package was not declared as a dependency in `pyproject.toml`, so rate limiting and token revocation code failed silently or with ImportError in production deployments.

**When**: Commit `28de6a8` (ISS-8).

**How fixed**: `redis>=5.0.0` added to `services/auth/pyproject.toml`. Conditional import pattern in `services/auth/src/redis_client.py` lines 19–25 provides graceful degradation if Redis is still unavailable.

### SCAR-019: Logout Token Lookup Using Re-Hash Instead of Stored Hash (C-001)

**What failed**: The logout endpoint re-hashed the incoming refresh token with Argon2id to find it in the database. Because Argon2id is salted (non-deterministic), the re-computed hash never matched the stored hash, making logout effectively a no-op — tokens could not be invalidated.

**When**: Commit `ae2e2d2` (Sprint R1, finding C-001).

**How fixed**: Logout now calls `find_token_by_value()` helper in `services/auth/src/services/token_lookup.py`, which uses cryptographic comparison rather than re-hashing. See `services/auth/src/routes/auth.py` lines 635–660.

### SCAR-020: `refresh` Permissions Returning SQLAlchemy `Row` Objects Instead of Strings (H-001)

**What failed**: Refresh endpoint permissions were extracted from SQLAlchemy `Row` objects rather than plain Python strings, causing downstream permission checks to compare `Row("read")` against `"read"` — always False.

**When**: Commit `ae2e2d2` (Sprint R1, finding H-001).

**How fixed**: Added explicit string extraction via `[str(row[0]) for row in result]` pattern. Already visible in `services/auth/src/routes/auth.py` line 97 in `_get_user_permissions()`.

### SCAR-021: `audit.log_event` Using `commit()` Instead of `flush()`, Breaking Transaction Boundaries (H-002)

**What failed**: `log_event()` called `session.commit()` inside the middle of request transactions, prematurely committing partial state and creating inconsistent transaction boundaries across `auth.py` and `rbac.py`.

**When**: Commit `ae2e2d2` (Sprint R1, finding H-002).

**How fixed**: `log_event()` in `services/auth/src/utils/audit.py` line 124 now uses `session.flush()`. Commit boundaries were normalized across `routes/auth.py` and `routes/rbac.py`.

### SCAR-022: Redis Connections Not Closed on Shutdown (H-003)

**What failed**: Redis client connections were never closed during service shutdown, causing connection leaks and potential Redis server exhaustion over time.

**When**: Commit `ae2e2d2` (Sprint R1, finding H-003).

**How fixed**: `close_rate_limit_client()` and `close_revocation_client()` are called during FastAPI lifespan shutdown in `services/auth/src/main.py` lines 73–83.

### SCAR-023: `JWT_ALGORITHM` Overridable via Environment Variable

**What failed**: `JWT_ALGORITHM` was a regular Pydantic settings field, meaning a stray `JWT_ALGORITHM=HS256` environment variable could silently downgrade the service from RS256 to a symmetric algorithm, breaking security guarantees.

**When**: Commit `65eb859` (security audit finding MEDIUM-003).

**How fixed**: `JWT_ALGORITHM` declared as `ClassVar[str] = "RS256"` in `services/auth/src/config.py` line 39, making it invisible to Pydantic settings resolution. ADR-0016 referenced.

### SCAR-024: `AUTOM8Y_ENV` Not Read by Child Settings Classes with Custom `env_prefix`

**What failed**: Child classes of `Autom8yBaseSettings` with a custom `env_prefix` (e.g., `AUTH__`) looked for `{PREFIX}AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`, causing `autom8y_env` to default to `LOCAL` in ECS task definitions where only `AUTOM8Y_ENV` is set. This triggered a production URL guard incorrectly.

**When**: Commit `1367461`.

**How fixed**: `AliasChoices("AUTOM8Y_ENV", "SERVICE_ENV")` added to the `autom8y_env` field in `services/auth/src/config.py` line 30, ensuring pydantic-settings reads the canonical env var regardless of prefix.

### SCAR-025: Refresh Token Storage Failure Returning HTTP 200 with Broken Token

**What failed**: If `RefreshToken` storage failed (e.g., DB write error) during login, the service returned HTTP 200 with an access token but no valid stored refresh token. The user would get a deferred 401 on their first token refresh attempt with no indication of the original failure.

**When**: Commit `96011fb`.

**How fixed**: Storage failure now raises `HTTPException(500)` in `services/auth/src/routes/auth.py` lines 580–588, with a clear message "An internal error occurred during login. Please try again." Regression tests: `services/auth/tests/test_refresh_token_storage.py`.

### SCAR-026: 4 CRITICAL Security Audit Findings (CRITICAL-001 through CRITICAL-004)

**What failed**: A security audit identified four critical architectural flaws:
- **CRITICAL-001**: Hybrid role architecture missing — roles not scoped to business_id
- **CRITICAL-002**: Verify-then-fallback missing — old JWT keys not accepted during rotation
- **CRITICAL-003**: Privilege escalation prevention missing
- **CRITICAL-004**: Audit log mutability — logs could be altered post-creation

**When**: Commits `1b1306b` (P0), `65eb859` (P2), `a4f7fac` (P3). ADRs: 0023, 0024, 0025, 0026.

**How fixed**: Full security hardening campaign. Regression tests: `services/auth/tests/test_critical_security_fixes.py` (1300+ lines covering all four findings).

## Category Coverage

| Category | Scars |
|----------|-------|
| **Integration failure / Import error** | SCAR-004 (phantom import), SCAR-005 (stdlib shadow), SCAR-006 (circular import), SCAR-009 (missing module) |
| **Performance / Connection management** | SCAR-007 (connection pool defeat), SCAR-010 (settings re-parse), SCAR-022 (Redis leak) |
| **Security — timing attack** | SCAR-012 (user enumeration via timing) |
| **Security — auth bypass / token tampering** | SCAR-013 (refresh token business_id bypass), SCAR-019 (logout no-op re-hash), SCAR-020 (Row object vs string) |
| **Security — DoS** | SCAR-014 (O(n) password reset scan), SCAR-001 (body consumption timeout) |
| **Security — config/hardening** | SCAR-015 (CORS wildcard), SCAR-023 (JWT_ALGORITHM override), SCAR-024 (env_prefix canonical name) |
| **Schema evolution** | SCAR-017 (TIMESTAMP→TIMESTAMPTZ), SCAR-014 (token_prefix column add) |
| **Dependency / infra** | SCAR-018 (Redis missing dep), SCAR-016 (utcnow() deprecation) |
| **Route registration** | SCAR-008 (admin router not registered) |
| **Transaction boundary** | SCAR-021 (audit log commit→flush) |
| **Silent wrong behavior** | SCAR-002 (NameError on success), SCAR-003 (double-revocation), SCAR-011 (session leak), SCAR-025 (200 with broken token) |
| **Architectural security** | SCAR-026 (CRITICAL-001 through 004) |

**Distinct categories identified**: 12

**Categories searched but not found in this codebase**: data corruption (non-datetime), race condition (async lock issues), message queue failures, cache invalidation bugs.

## Fix-Location Mapping

| Scar | Fix File(s) | Function / Location |
|------|-------------|---------------------|
| SCAR-001 | `services/auth/src/middleware/rate_limit.py` | `rate_limit_login()` lines 121–135 |
| SCAR-002 | `services/auth/src/routes/internal.py` | `exchange_service_token()` ~line 375 |
| SCAR-003 | `services/auth/src/routes/admin.py` | `revoke_service_key()` and related — lines 295, 400, 478, 601, 610 |
| SCAR-004 | `services/auth/src/routes/auth.py` | `get_audit_logs()` line ~1172 |
| SCAR-005 | `services/auth/src/mailer/` (directory rename from `email/`) | Package-level; import site at `routes/auth.py` line 986 |
| SCAR-006 | `services/auth/src/auth/__init__.py`, `services/auth/src/auth/dependencies.py` | Barrel re-export, `dependencies.py` direct import |
| SCAR-007 | `services/auth/src/db/database.py` | Module-level globals `_sync_engine`, `_async_engine`; `get_sync_engine()`, `get_async_engine()` |
| SCAR-008 | `services/auth/src/main.py` | Router registration line 15 |
| SCAR-009 | `services/auth/src/schemas/admin.py` | New file (created) |
| SCAR-010 | `services/auth/src/config.py` | `get_settings()` line 174, `@lru_cache` decorator |
| SCAR-011 | `services/auth/src/routes/charter.py` | `list_roles()` — `Depends(get_db)` parameter |
| SCAR-012 | `services/auth/src/routes/auth.py` | `_DUMMY_HASH` lines 65–70; `verify_password(_DUMMY_HASH, ...)` line 419 |
| SCAR-013 | `services/auth/src/routes/auth.py` | Refresh endpoint line 721 |
| SCAR-014 | `services/auth/src/routes/auth.py`, `services/auth/src/models/password_reset_token.py`, `services/auth/migrations/versions/015_add_token_prefix_to_password_reset.py` | Lines 955, 965, 1043 in auth.py |
| SCAR-015 | `services/auth/src/config.py` | `CORS_ORIGINS` property lines 73–82 |
| SCAR-016 | All `*.py` files in `services/auth/src/` (220 callsites) | Ecosystem-wide |
| SCAR-017 | `services/auth/migrations/versions/016_timestamps_to_timestamptz.py` | Migration file |
| SCAR-018 | `services/auth/pyproject.toml`, `services/auth/src/redis_client.py` | Dependency declaration; conditional import lines 19–25 |
| SCAR-019 | `services/auth/src/routes/auth.py`, `services/auth/src/services/token_lookup.py` | Logout function lines 635–660 |
| SCAR-020 | `services/auth/src/routes/auth.py` | `_get_user_permissions()` line 97 |
| SCAR-021 | `services/auth/src/utils/audit.py` | `log_event()` line 124 |
| SCAR-022 | `services/auth/src/main.py` | Lifespan shutdown block lines 73–83 |
| SCAR-023 | `services/auth/src/config.py` | `JWT_ALGORITHM: ClassVar[str] = "RS256"` line 39 |
| SCAR-024 | `services/auth/src/config.py` | `autom8y_env` field `AliasChoices` line 30 |
| SCAR-025 | `services/auth/src/routes/auth.py` | Login endpoint lines 572–588 |
| SCAR-026 | `services/auth/src/charter/` (full module), `services/auth/migrations/versions/005_hybrid_role_schema.py`, `006_audit_immutability_trigger.py` | Multiple files |

All 26 fix file paths verified to exist in the repository.

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Location | Regression Test |
|------|-------------------|----------|-----------------|
| SCAR-001 | IP-based rate limiting; explicit comment forbidding `request.body()` in middleware | `middleware/rate_limit.py` lines 127–135 | Implicit via `TestRateLimiting` in `tests/test_security.py` |
| SCAR-002 | `TestServiceTokenScopeRegression` regression test class | `tests/test_internal_routes.py` line 599 | Yes — `test_service_token_does_not_500_on_success` |
| SCAR-003 | `revoked_at.is_(None)` filter on all key lookup queries | `routes/admin.py` at 5 sites | `TestListServiceKeys.test_list_service_keys_excludes_revoked` in `tests/test_service_key_manager.py` |
| SCAR-004 | Import directly from canonical module (`src.auth.jwt_handler`) | `routes/auth.py` line ~1172 | No dedicated regression test found |
| SCAR-005 | Package renamed to `src/mailer/`; spike doc written | `src/mailer/`, `docs/spikes/SPIKE-AUTH-EMAIL-STDLIB-SHADOW.md` | `tests/test_email_service.py` (import verification) |
| SCAR-006 | Barrel `__init__.py` documents the broken cycle; `dependencies.py` imports directly | `src/auth/__init__.py` lines 7–11 | Startup-level; no dedicated test |
| SCAR-007 | Module-level singleton engines; `pool_pre_ping=True`; docstring warns "ONCE at module level" | `db/database.py` lines 11–14, 34–38 | No dedicated regression test found |
| SCAR-008 | Router added to import list in `main.py` | `main.py` line 15 | `tests/test_internal_routes.py` exercise admin routes |
| SCAR-009 | `schemas/admin.py` exists with `NOTE: api_key is only returned at creation time` | `schemas/admin.py` lines 39, 61 | `tests/test_service_key_manager.py` |
| SCAR-010 | `@lru_cache` on `get_settings()`; comment warning about `cache_clear()` in tests | `config.py` line 174 | No dedicated regression test; note in commit message |
| SCAR-011 | FastAPI `Depends(get_db)` pattern enforced; manual session creation removed | `routes/charter.py` | `tests/test_charter_client.py` |
| SCAR-012 | `_DUMMY_HASH` module-level constant; `verify_password()` always called; `BUSINESS_LOOKUP_MIN_RESPONSE_TIME_MS=100` | `routes/auth.py` lines 65–70, 235 | `TestTimingAttacks` in `tests/test_security.py` lines 259–320; `test_verify_password_timing_safe` in `tests/test_auth_utils.py` line 105 |
| SCAR-013 | Cryptographic token hash verification before accepting refresh | `routes/auth.py` line 721 | `TestTokenTampering.test_swapped_refresh_and_access_tokens` in `tests/test_security.py` line 91 |
| SCAR-014 | `token_prefix` column + index; O(1) prefix lookup; `HIGH-003 FIX` inline comment | `routes/auth.py` lines 955, 1043; `models/password_reset_token.py` line 43 | `tests/test_password_reset.py` |
| SCAR-015 | Explicit CORS allowlist (no wildcards); `is_local` guard for localhost | `config.py` lines 70–82 | `TestHTTPSecurity.test_cors_headers_present` in `tests/test_security.py` line 501 |
| SCAR-016 | `datetime.now(UTC)` used universally; ruff rule enforced | Ecosystem-wide | `TestTimingAttacks` (indirect); CI lint enforcement |
| SCAR-017 | Migration 016 converts all tables to `TIMESTAMPTZ`; `TIMESTAMPTZ` constants in ORM models | `migrations/versions/016_timestamps_to_timestamptz.py` | No dedicated regression test |
| SCAR-018 | Conditional Redis import with `REDIS_AVAILABLE` flag; fail-open policy | `redis_client.py` lines 18–25, 44–50 | `TestRateLimitClient.test_check_returns_allowed_when_redis_unavailable` in `tests/test_api_key_hardening.py` line 213 |
| SCAR-019 | `find_token_by_value()` helper using cryptographic comparison | `services/token_lookup.py`; `routes/auth.py` lines 635–660 | `tests/test_revocation.py` |
| SCAR-020 | `[str(row[0]) for row in result]` explicit string extraction | `routes/auth.py` line 97 | No dedicated regression test found |
| SCAR-021 | `session.flush()` in `log_event()`; commit boundaries explicitly managed per-route | `utils/audit.py` line 124 | `tests/test_audit_logging.py` |
| SCAR-022 | `close_revocation_client()` in FastAPI lifespan shutdown | `main.py` lines 73–83 | No dedicated regression test found |
| SCAR-023 | `JWT_ALGORITHM: ClassVar[str] = "RS256"` — pydantic-invisible field; `validate_required_secrets()` runtime check | `config.py` lines 37–39, 166–171 | `tests/test_rs256_jwks.py` |
| SCAR-024 | `AliasChoices("AUTOM8Y_ENV", "SERVICE_ENV")` on `autom8y_env` field | `config.py` line 30 | No dedicated regression test in auth service |
| SCAR-025 | `HTTPException(500)` on refresh token storage failure; comment explains "deferred 401" risk | `routes/auth.py` lines 580–588 | `tests/test_refresh_token_storage.py` — `test_login_returns_500_when_refresh_token_storage_fails` line 25 |
| SCAR-026 | Full Charter security hardening (ADRs 0023–0026); audit trigger in migration 006 | `charter/` module; `migrations/versions/006_audit_immutability_trigger.py` | `tests/test_critical_security_fixes.py` (1300+ lines) |

**Scars with no dedicated regression test**: SCAR-004, SCAR-006, SCAR-007, SCAR-010, SCAR-017, SCAR-020, SCAR-022, SCAR-024 (8 of 26).

## Agent-Relevance Tagging

| Scar | Responsibility Area | Why Agents Need This |
|------|--------------------|-----------------------|
| SCAR-001 | Middleware / Auth | Any change to login middleware must not call `request.body()`; the comment is a load-bearing guard |
| SCAR-002 | API / Internal routes | When extending `exchange_service_token`, scope variables must be defined before logger calls |
| SCAR-003 | API / Admin routes | All key/token lookup queries must include `revoked_at.is_(None)` filter — soft-delete semantics throughout |
| SCAR-004 | API / Auth routes | Lazy imports must be verified against actual module exports before committing |
| SCAR-005 | Architecture | New packages under `src/` must not shadow stdlib module names; use project-specific names |
| SCAR-006 | Architecture | The `src.auth` barrel is the canonical import path; submodules may not import back from the barrel |
| SCAR-007 | Database | Engines must not be created inside request handlers; always use `get_sync_engine()` / `get_async_engine()` |
| SCAR-008 | API / Router wiring | New route modules must be registered in `main.py`; existence of the module is not sufficient |
| SCAR-009 | API / Schemas | Schema modules must be committed alongside the routes that import them |
| SCAR-010 | Config | `get_settings()` is a singleton; tests that mutate env must call `get_settings.cache_clear()` |
| SCAR-011 | Database / DI | Never create DB sessions manually; always use `Depends(get_db)` for FastAPI routes |
| SCAR-012 | Auth / Security | User-not-found path must always call `verify_password(_DUMMY_HASH, ...)` — removing this breaks timing safety |
| SCAR-013 | Auth / Security | Refresh token validation must verify both token hash AND business_id claim |
| SCAR-014 | Database / Security | Password reset lookup uses `token_prefix` for O(1); never revert to full table scan |
| SCAR-015 | Security / Config | CORS origins must remain explicit; no wildcards |
| SCAR-016 | All | Use `datetime.now(UTC)` universally; `datetime.utcnow()` is banned and caught by ruff |
| SCAR-017 | Database | All new datetime columns must use `TIMESTAMPTZ`; `TIMESTAMP` columns cause asyncpg errors |
| SCAR-018 | Infrastructure | Redis is a declared dependency (`redis>=5.0.0`); fail-open behavior is intentional |
| SCAR-019 | Auth | Logout/revocation flows must use `find_token_by_value()`, not re-hashing |
| SCAR-020 | Database | SQLAlchemy `Row` objects must be explicitly cast to `str` when used as permission strings |
| SCAR-021 | Database / Audit | `log_event()` uses `flush()` not `commit()`; commit boundaries are per-route responsibility |
| SCAR-022 | Infrastructure | Redis clients must be closed in lifespan shutdown; see `close_revocation_client()` pattern |
| SCAR-023 | Security / Config | `JWT_ALGORITHM` is a `ClassVar` — cannot be overridden by env vars; this is intentional |
| SCAR-024 | Config | `autom8y_env` uses `AliasChoices` — always reads `AUTOM8Y_ENV` regardless of `env_prefix` |
| SCAR-025 | Auth | Refresh token storage failure must surface as 500, not silently succeed with 200 |
| SCAR-026 | Security / Charter | Charter RBAC is mothballed — `DO NOT re-enable without reviewing ADR-CHARTER-001`; CRITICAL fixes documented in ADRs 0023–0026 |

**Platform-wide scars** (affect any new service or module): SCAR-005, SCAR-006, SCAR-007, SCAR-011, SCAR-016, SCAR-017.

**Historical / already-defended scars** (marker still present, low active risk): SCAR-004, SCAR-008, SCAR-009.

## Knowledge Gaps

1. **SCAR-026 sub-findings detail**: CRITICAL-001 through CRITICAL-004 are cataloged at the level of ADR numbers (0023–0026) and test class headers, but the exact before/after behavior of each finding was not read from the ADR documents themselves (not observed in scope).

2. **Missing regression tests**: 8 of 26 scars have no dedicated regression test (SCAR-004, -006, -007, -010, -017, -020, -022, -024). The risk of regression recurrence for these is undocumented.

3. **Sprint R1 finding C-001 (SCAR-019) — `token_lookup.py` not read**: The `find_token_by_value()` helper implementation was referenced but not read directly; the fix location is confirmed from `routes/auth.py` import and commit message, but the helper's internal logic was not observed.

4. **Dev mode production guard (client SDK)**: `services/auth/client/autom8y_auth_client/dev_mode.py` contains a CRITICAL security guard (`AUTH_DEV_MODE=true` must not activate in production). This is scoped to the SDK client subdirectory and was observed but not deeply cataloged as a separate scar — it may represent a past failure that necessitated the guard.

5. **Credential Vault and Charter mothball lineage**: Both features are mothballed with `DO NOT re-enable` guards, but the original failure or design decision that led to mothballing was not captured. ADR-VAULT-001 and ADR-CHARTER-001 exist but were not read.
