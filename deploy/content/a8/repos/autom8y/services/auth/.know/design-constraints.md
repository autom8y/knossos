---
domain: design-constraints
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

# Codebase Design Constraints

## Tension Catalog Completeness

### TENSION-001: Dual Role System (Active vs Mothballed Charter)

The codebase contains two parallel, incompatible RBAC implementations.

**Active system**: `src/models/role.py`, `src/models/permission.py`, `src/routes/rbac.py` — tables `roles`, `permissions`, `role_permissions`, `user_roles`. Roles are always business-scoped (`business_id` is non-nullable on the `roles` table).

**Mothballed Charter system**: `src/charter/` (models, client, audit, circuit_breaker, schemas) + `src/routes/charter.py` — uses a separate `charter_roles`, `charter_permissions`, `charter_role_permissions`, `charter_user_roles` table family. Migration 005 (`services/auth/migrations/versions/005_hybrid_role_schema.py`) introduces system-wide (null `business_id`) roles in the charter schema per ADR-0024. This hybrid architecture exists only in Charter and is not reflected in the active `roles` table.

Both systems define `Role`, `Permission`, and `UserRole` concepts. JWT payloads embed `roles` and `permissions` from the active RBAC tables. Charter's role-to-permission resolution lives in `src/charter/client.py` but its routes are commented out in `src/main.py` (line 242-247).

**Impact**: Any future activation of Charter requires a migration strategy from the active RBAC system, which has no system-role concept.

### TENSION-002: Refresh Token Identity Crisis — JWT Verified as API Key

`src/services/token_lookup.py` (line 61) and `src/routes/auth.py` (lines 548, 736, 1068) treat refresh tokens as API keys for hashing/verification purposes: `hash_api_key(refresh_token)` and `verify_api_key(stored.token_hash, raw_token)`. The function names reference "api_key" but operate on JWT refresh tokens. The data model (`src/models/refresh_token.py`) uses field name `token_hash` with description "Argon2id hashed refresh token."

This is not merely a naming mismatch — `verify_api_key` from `src/auth/api_key_handler.py` is shared infrastructure. Changing the API key hashing parameters would silently break refresh token verification.

### TENSION-003: Redis Used for Three Independent Purposes via One URL

`src/redis_client.py` hosts `RevocationClient` and `RateLimitClient` as distinct classes, both sharing `REDIS_URL` from config. `src/services/identifier.py` instantiates a third Redis connection (`_identifier_redis`) independently, also from `REDIS_URL`, using a different key namespace (`cache:identifier:*`). There is no shared connection pool — three separate connection pools compete for the same Redis instance. Each has its own lifecycle management (`close_revocation_client`, `close_rate_limit_client`, `close_identifier_redis` in `src/main.py` lines 87-108).

### TENSION-004: Sync/Async DB Session Dual Track (ADR-ASYNC-DB-001 in Progress)

`src/db/database.py` maintains both `_sync_engine` and `_async_engine` module-level instances. `src/services/api_key_service.py` (lines 43, 124, 175) explicitly supports `Session | AsyncSession` with `isinstance` branching. The internal route `src/routes/internal.py` (line 218) uses `AsyncSession` while most routes use sync `Session` via `get_db`. This is documented as a "hybrid migration period" per ADR-ASYNC-DB-001 but no terminal state is defined in the code.

### TENSION-005: Mothballed Credential Vault — Dead Code With Live Database Schema

15 files bear the `MOTHBALLED - DO NOT USE` header (per ADR-VAULT-001): `src/services/credential_vault.py`, `src/services/encryption_service.py`, `src/services/oauth_state.py`, `src/services/oauth_adapters/` (3 files), `src/models/credential_provider.py`, `src/models/encryption_key.py`, `src/models/external_credential.py`, `src/models/credential_access_log.py`, `src/models/oauth_state.py`, `src/routes/credentials.py`, `src/routes/oauth.py`, `src/workers/token_refresh_worker.py`, `src/db/seeds/seed_providers.py`, `src/workers/__init__.py`.

The models are still imported in `src/models/__init__.py` (lines 13-16) because Alembic migrations depend on them (stated in `services/auth/src/services/credential_vault.py` line 15). This means mothballed model classes are in the active SQLModel metadata and will be included in `init_db()`.

### TENSION-006: `business_scope_validation` Middleware Defined But Not Registered

`src/middleware/business_scope.py` defines `business_scope_validation` middleware but it is never registered in `src/main.py`. Only `SecurityHeadersMiddleware`, `CORSMiddleware`, and `rate_limit_login` are active middleware. The business scope isolation the middleware provides (cross-tenant access prevention) is absent at the ASGI layer.

### TENSION-007: Charter `is_admin` Bug — Known Pre-Activation Blocker

`src/routes/charter.py` (line 80-84) contains a self-documented DRIFT: `User` model has no `is_admin` attribute. The code uses `getattr(current_user, "is_admin", False)` with a TODO and a note that this is a known bug in `src/main.py` (line 245-246). Since Charter is mothballed, this is latent but blocks reactivation.

### TENSION-008: `verify_token` Derives Public Key From Private Key at Verification Time

`src/auth/jwt_handler.py` (lines 296-312): when no `public_key_pem` is provided, `verify_token` loads the private key and derives the public key each call. There is no caching of the public key equivalent to the `_signing_key` cache (line 26). This is a hot-path performance concern for every authenticated request.

## Trade-off Documentation

### TENSION-001 Trade-off

- **Current state**: Active RBAC (`roles`/`permissions` tables) runs; Charter (`charter_*` tables) is mothballed with all routes commented out.
- **Ideal state**: One unified RBAC system. Charter was designed to supersede the active RBAC with system roles and a circuit-breaker-backed client.
- **Why current state persists**: Charter routes were disabled per ADR-CHARTER-001 (referenced but ADR file not present in `docs/decisions/`). The known `is_admin` bug (TENSION-007) was a pre-existing blocker. The cost of live migration between two role schemas while preserving tenant isolation has not been absorbed.
- **ADR links**: ADR-CHARTER-001, ADR-0024 — referenced in code but files not found in `services/auth/docs/`.

### TENSION-002 Trade-off

- **Current state**: Refresh tokens are stored as Argon2id hashes and verified using the API key verification function.
- **Ideal state**: Separate `hash_refresh_token` / `verify_refresh_token` functions that make the intent explicit.
- **Why current state persists**: Argon2id parameters are the same; reusing the function avoids duplication. There is a comment in `src/auth/password.py` (lines 11-13) noting the password hasher instance is the canonical one for all hashing operations. The functional outcome is identical; only semantics and coupling are at risk.

### TENSION-003 Trade-off

- **Current state**: Three Redis connection pools for one Redis instance.
- **Ideal state**: Single shared connection pool with namespaced key prefixes per concern.
- **Why current state persists**: Each concern (revocation, rate limiting, identifier caching) was added incrementally with its own connection management. Unifying requires coordinating lifecycle across three unrelated subsystems. The fail-open policy per ADR-0017 means Redis failures are non-blocking, reducing urgency.

### TENSION-004 Trade-off

- **Current state**: Sync and async DB session factories coexist; `APIKeyService` has `isinstance` branching.
- **Ideal state**: Fully async DB layer.
- **Why current state persists**: Migration is in progress per ADR-ASYNC-DB-001. Internal service-token endpoint was migrated first (`src/routes/internal.py` uses `AsyncSession`). Most route handlers remain on sync `Session`.
- **External constraint**: Blocking migration is SQLModel's incomplete async support; some `db.exec(select(...))` patterns have no clean async equivalent.

### TENSION-005 Trade-off

- **Current state**: 15 mothballed files remain in the codebase; models are imported in `__init__.py`.
- **Ideal state**: Models stripped to a migration-only shim or moved to a `legacy/` namespace.
- **Why current state persists**: Alembic requires model classes to be importable at migration time. Removing imports from `src/models/__init__.py` would break `alembic revision --autogenerate`. Full removal requires stripping the Alembic dependency on those models too.

### TENSION-006 Trade-off

- **Current state**: `business_scope_validation` exists but is unregistered.
- **Ideal state**: Either registered in `src/main.py` or explicitly removed.
- **Why current state persists**: Unknown. No ADR or comment explains the omission. The middleware is implemented and tested (implied by its existence in the codebase) but was never wired.

### TENSION-007 Trade-off

- **Current state**: Bug documented in `src/main.py` comment; Charter remains disabled.
- **Why current state persists**: Charter is mothballed. The fix is a one-line change (add an `is_admin` property to `User` or replace the check) but is blocked on the Charter reactivation decision.

## Abstraction Gap Mapping

### Gap-001: No Unified Token Verification Abstraction

Token verification occurs in three locations with slightly divergent patterns:
1. `src/auth/dependencies.py` (`get_current_user`) — calls `verify_token` then checks `token_type == "access"`
2. `src/middleware/business_scope.py` (`business_scope_validation`) — calls `verify_token` directly without `token_type` check
3. `src/routes/internal.py` — calls `verify_token` inline for service token verification

No shared `verify_access_token()` wrapper consolidates the `token_type` check + revocation check. The Redis revocation check (`is_token_revoked`) exists in `RevocationClient` but is not called from `get_current_user` in `src/auth/dependencies.py` — meaning revoked tokens can still pass the dependency unless the specific endpoint checks separately.

**Maintenance burden**: Adding a revocation check to authentication requires modifying 3+ locations.

### Gap-002: No Permission Object for Authorization

`src/routes/rbac.py` defines `require_admin` as a FastAPI dependency that checks `"admin:roles" in permissions`. Permissions are plain strings in the JWT payload. There is no typed `Permission` object used at the route layer — the `Permission` SQLModel in `src/models/permission.py` is used only for CRUD, never for authorization decisions.

**Risk**: Permission string values are hardcoded across route files with no central registry of valid permission strings.

### Gap-003: In-Memory Brute Force Protection Does Not Survive Restarts

`src/utils/brute_force.py` uses `defaultdict(list)` in-memory with module-level globals. Failed attempt state is lost on process restart or ECS task replacement. Redis is available for persistence but brute force protection does not use it.

**Maintenance burden**: Horizontally scaled deployments (multiple ECS tasks) have independent brute force counters — an attacker can distribute attempts across instances.

### Gap-004: Business Scope Extraction Logic Is An Inline Heuristic

`src/middleware/business_scope.py` `_extract_business_id_from_request` uses URL path position heuristics (e.g., "if path_parts[0] == 'businesses' ..."). If new route prefixes are added that don't match existing patterns, cross-tenant access becomes undetectable by this layer. There is no abstract "scoped request" concept.

## Load-Bearing Code Identification

### LBC-001: `src/auth/password.py` — `_hasher` Instance

`services/auth/src/auth/password.py` defines the `_hasher` Argon2id instance (lines 18-21). The comment (lines 11-13) explicitly documents: "Imported by api_key_handler.py to ensure consistent security parameters." This instance is the canonical hasher for **all** Argon2id operations in the service: password hashing, API key hashing, and (via `hash_api_key`) refresh token hashing.

- **What depends on it**: `hash_password`, `verify_password` (auth flow), `hash_api_key`, `verify_api_key` (API key and refresh token flows)
- **Naive fix danger**: Changing Argon2id parameters would invalidate ALL existing stored hashes across passwords, API keys, and refresh tokens. No migration path exists without re-hashing all secrets.
- **Rating**: FROZEN — any change requires coordinated hash migration.

### LBC-002: `src/auth/jwt_handler.py` — `JWT_ALGORITHM = ClassVar["RS256"]`

`services/auth/src/config.py` (lines 37-39): `JWT_ALGORITHM` is a `ClassVar[str]` set to `"RS256"`. This is documented: "Keep this as a ClassVar so a stray env var like JWT_ALGORITHM=HS256 can't silently override runtime behavior." `validate_required_secrets` (line 166-170) explicitly raises if this ever deviates.

- **What depends on it**: All JWT signing (`create_access_token`, `create_refresh_token`) and verification (`verify_token`). All consuming services using the auth client library (`client/autom8y_auth_client/`) rely on RS256 token structure.
- **Naive fix danger**: Switching to HS256 would invalidate all issued tokens and break every service using `autom8y_auth_client`.
- **Rating**: FROZEN — algorithm changes require coordinated rotation across all dependent services.

### LBC-003: `src/models/__init__.py` — Mothballed Model Imports

`services/auth/src/models/__init__.py` imports all mothballed Credential Vault models (lines 13-16). This file is imported by Alembic migrations.

- **What depends on it**: Alembic `env.py` (`services/auth/migrations/env.py`) imports models to build SQLAlchemy metadata for migration generation.
- **Naive fix danger**: Removing mothballed imports from `__init__.py` breaks `alembic revision --autogenerate`.
- **Rating**: COORDINATED — removing requires simultaneously updating Alembic env and migration history.

### LBC-004: `src/redis_client.py` — Global Singleton Pattern

`_revocation_client`, `_rate_limit_client` are module-level globals in `src/redis_client.py`. They are initialized lazily on first call and closed explicitly in `src/main.py` lifespan (lines 87-108).

- **What depends on it**: `src/auth/jwt_handler.py` (`_track_token_async`), `src/routes/internal.py` (revocation endpoints), `src/middleware/rate_limit.py` (rate limiting).
- **Naive fix danger**: Converting to a DI-injected pattern (passing as dependency) requires changing all call sites. The `_track_token_async` function in `jwt_handler.py` calls `get_revocation_client()` inside an async background tracking step — it cannot receive a DI-injected client without refactoring.
- **Rating**: COORDINATED — restructuring the global pattern requires cross-file changes.

### LBC-005: `src/jwks.py` — Public Key Derivation From Private Key

`verify_token` in `src/auth/jwt_handler.py` (lines 303-312) imports `get_public_key_from_private` and `load_private_key` from `src/jwks.py`. Every token verification in the default (no `public_key_pem`) path loads and parses the PEM private key to derive the public key. Only the signing key (`_signing_key`) is cached; the verification path has no equivalent cache.

- **What depends on it**: Every authenticated route via `get_current_user`, business scope middleware, and internal token exchange.
- **Naive fix danger**: Adding a cached public key bypasses the PEM parsing but requires thread-safe initialization at module level.
- **Rating**: SAFE to add caching, but the hot path is currently unoptimized.

## Evolution Constraint Documentation

### Area: JWT Token Algorithm

**Changeability**: FROZEN

RS256 is enforced as `ClassVar` in config. Changing requires: rotating all issued tokens, updating JWKS endpoint, coordinating all consuming services. See LBC-002.

### Area: Argon2id Parameters (password/API key/refresh token hashing)

**Changeability**: MIGRATION

Changing parameters invalidates all stored hashes. Would require a re-hash migration for all `User.password_hash`, `APIKey.key_hash`, `RefreshToken.token_hash`, `PasswordResetToken.token_hash` rows simultaneously. No migration tooling exists for this.

### Area: Mothballed Charter Module

**Changeability**: FROZEN (until ADR-CHARTER-001 activation criteria are met)

Charter cannot be reactivated without:
1. Fixing `is_admin` bug in `src/routes/charter.py` (line 80-84)
2. Deciding migration strategy from active RBAC to Charter RBAC
3. Deciding whether to deprecate `roles`/`permissions`/`user_roles` tables or run both systems

ADR-CHARTER-001 activation criteria are referenced but the ADR file is not present in `services/auth/docs/`.

### Area: Mothballed Credential Vault

**Changeability**: FROZEN (until ADR-VAULT-001 activation criteria are met)

15 mothballed files. Reactivation requires: reviewing ADR-VAULT-001 activation triggers, re-wiring routes in `src/main.py`, and re-enabling seed_providers. The credential vault config fields (`CREDENTIAL_KMS_KEY_ARN`, OAuth credentials) remain in `src/config.py` and are loaded even when vault is mothballed.

### Area: Sync-to-Async DB Migration (ADR-ASYNC-DB-001)

**Changeability**: COORDINATED

In-progress migration. `APIKeyService` dual-mode is a shim. Routes should migrate to `get_async_db` incrementally. When complete, `_sync_engine`, `get_db`, and `sessionmaker` factory can be removed from `src/db/database.py`.

### Area: Redis Connection Architecture

**Changeability**: COORDINATED

Three independent connection pools. Refactoring to a shared pool requires coordinating `src/redis_client.py`, `src/services/identifier.py`, and lifecycle management in `src/main.py`.

### Deprecated Markers

- `src/workers/token_refresh_worker.py` — MOTHBALLED
- `src/services/oauth_adapters/` — MOTHBALLED
- `src/routes/credentials.py`, `src/routes/oauth.py` — MOTHBALLED
- `src/middleware/business_scope.py` — Implemented but unregistered

### External Dependency Constraints

- **`autom8y_config.Autom8yBaseSettings`**: Auth `Settings` inherits from this SDK base class. The `_guard_production_urls` override (lines 127-135 of `src/config.py`) explicitly no-ops a base class behavior — changes to the base class guard behavior require re-evaluating this override.
- **`autom8y_telemetry`**: `instrument_app` is called at import time in `src/main.py`. If the SDK's `InstrumentationConfig` interface changes, startup fails.
- **`autom8y_auth_client` (in `services/auth/client/`)**: Consumes the JWKS endpoint and JWT payload structure. Any change to JWT claims or algorithm is a breaking change for this client.

## Risk Zone Mapping

### RISK-001: Revocation Check Not Called in Primary Auth Dependency

**Location**: `services/auth/src/auth/dependencies.py` — `get_current_user`

`get_current_user` verifies the JWT signature and expiry but does NOT check Redis revocation (`RevocationClient.is_token_revoked`). The revocation infrastructure exists and is used in `src/routes/internal.py` for explicit revocation endpoints. However, the main authentication dependency does not check if the token's `jti` is in the blocklist.

**Risk**: A token revoked via the `/internal/revoke/token/{jti}` or `/internal/revoke/user/{user_id}` endpoints continues to authenticate successfully until natural expiry (up to 15 minutes).

**Input path lacking validation**: All protected routes using `Depends(get_current_user)`.

### RISK-002: Business Scope Middleware Not Registered

**Location**: `services/auth/src/middleware/business_scope.py`

`business_scope_validation` is implemented but absent from the ASGI middleware stack in `src/main.py`. Cross-tenant request validation is not enforced at the middleware layer.

**Risk**: An authenticated user with a valid token for `business_id=A` can send requests containing `business_id=B` in the URL or body without middleware interception. Individual route handlers may or may not enforce this check independently.

### RISK-003: In-Memory Brute Force State Not Shared Across Instances

**Location**: `services/auth/src/utils/brute_force.py`

Module-level global `_brute_force` and `_password_reset_limiter` instances hold all failed attempt state in process memory.

**Risk**: In a horizontally scaled deployment (multiple ECS tasks), an attacker can distribute 4 attempts per instance and never trigger a lockout, since attempt counts are never shared.

### RISK-004: `guid_migrations` Query Assumes External Table Existence

**Location**: `services/auth/src/services/identifier.py` (lines 302-322)

`_resolve_numeric_guids` executes raw SQL against `guid_migrations` table via `self._db.connection()`. This table is not in the auth service's Alembic migration history — it is an external cross-service table expected to exist. The fallback (lines 316-338) catches the exception and silently returns empty results.

**Risk**: If `guid_migrations` does not exist or is inaccessible, numeric external ID resolution silently returns not-found for all inputs with no alerting.

### RISK-005: Credential Vault Config Present but Vault is Mothballed

**Location**: `services/auth/src/config.py` (lines 110-121)

`CREDENTIAL_KMS_KEY_ARN`, `OAUTH_CALLBACK_BASE_URL`, `OAUTH_DEFAULT_REDIRECT_URI`, `OAUTH_ASANA_CLIENT_ID`, `OAUTH_ASANA_CLIENT_SECRET` are loaded from environment even though Credential Vault is mothballed. If these are accidentally injected in production (e.g., via stale IaC), there is no runtime warning.

**Risk**: Low direct risk, but stale credential config pollutes the `Settings` object and may cause confusion during incident response or secret rotation.

### RISK-006: `expires_at` Timezone Assumption in `APIKeyService`

**Location**: `services/auth/src/services/api_key_service.py` (lines 233-237)

`check_key_expiry` contains: `# If expires_at is naive, assume it's UTC`. This defensive coding suggests some `APIKey.expires_at` values in the database may be stored as timezone-naive timestamps.

**Risk**: If a timezone-naive `expires_at` is incorrectly offset, an expired key could appear valid or a valid key could appear expired.

## Knowledge Gaps

1. **ADR files not found**: ADR-VAULT-001, ADR-CHARTER-001, ADR-0017, ADR-0016, ADR-0024, ADR-ASYNC-DB-001, ADR-AUTH-UX-003, ADR-AUTH-UX-004, ADR-ENV-NAMING-CONVENTION are all referenced in source code but no `docs/decisions/` directory exists in `services/auth/`. It is unclear whether these documents exist elsewhere in the monorepo.

2. **`business_scope_validation` registration intent**: No comment, ADR reference, or TODO explains why the middleware is implemented but unregistered. Could be intentional (per-route enforcement instead) or an omission.

3. **Charter reactivation status**: No documentation clarifies whether Charter is a future roadmap item or permanently abandoned. The mothball comment instructs to review "activation triggers in ADR-CHARTER-001" but that document is unavailable.

4. **`guid_migrations` table provenance**: The table is queried by `src/services/identifier.py` but is not managed by auth's Alembic migrations. No documentation identifies which service owns this table or its schema.

5. **Revocation check integration design intent**: Whether `get_current_user` intentionally skips the revocation check (relying on token TTL) or whether this is an omission is not documented. ADR-0017 addresses the fail-open policy for Redis but does not clarify whether all auth paths should check revocation.
