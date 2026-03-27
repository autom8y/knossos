---
domain: architecture
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

# Codebase Architecture

## Package Structure

The auth service is a Python 3.12 FastAPI microservice rooted at `services/auth/src/`. The `src/` directory contains flat-level modules and sub-packages. There is no `app/` directory — the service uses `src/` as the source root (configured via `pyproject.toml` `pythonpath = ["src"]`).

**Top-level source layout** (`services/auth/src/`):

| Path | Purpose | File count |
|---|---|---|
| `src/main.py` | FastAPI app entry point, middleware registration, router inclusion | 1 |
| `src/config.py` | Pydantic Settings class, all env var config | 1 |
| `src/jwks.py` | RSA key loading and JWKS generation | 1 |
| `src/redis_client.py` | Redis clients: token revocation + rate limiting | 1 |
| `src/health_models.py` | Health probe response types | 1 |
| `src/auth/` | Authentication primitives: JWT, API key, password, FastAPI dep | 4 |
| `src/charter/` | Mothballed RBAC sub-system (ADR-CHARTER-001) | 6 |
| `src/db/` | Database engine, session factory, seeds | 3 (+2 seeds) |
| `src/mailer/` | SendGrid email delivery | 3 |
| `src/middleware/` | HTTP middleware: rate limit, security headers, business scope | 3 |
| `src/models/` | SQLModel ORM table definitions | 14 |
| `src/observability/` | Structured logging wrapper (autom8y-log) | 1 |
| `src/routes/` | FastAPI route handlers | 9 (6 active, 3 mothballed) |
| `src/schemas/` | Pydantic request/response schemas | 3 |
| `src/services/` | Business logic services | 6 (+2 oauth_adapters) |
| `src/utils/` | Cross-cutting utilities: audit, brute force, rate limit | 3 |
| `src/workers/` | Mothballed background worker (credential refresh) | 1 |

**Active routes** (included in `main.py`): `auth`, `rbac`, `api_keys`, `well_known`, `internal`, `admin`

**Mothballed routes** (present on disk, excluded via comments in `main.py`): `credentials`, `oauth`, `charter` — each with `MOTHBALLED` docstrings and references to ADR-VAULT-001 or ADR-CHARTER-001.

**Hub packages** (imported by many siblings): `src/models/` (imported by routes, services, workers, seeds), `src/auth/` (imported by routes, services), `src/config.py` (imported by nearly every module).

**Leaf packages** (import nothing internal): `src/observability/logger.py`, `src/health_models.py`, `src/schemas/`.

**Key files with full paths**:
- `services/auth/src/main.py`
- `services/auth/src/config.py`
- `services/auth/src/models/__init__.py`
- `services/auth/src/auth/__init__.py`

## Layer Boundaries

The codebase has three functional layers. Import direction flows downward; no upward imports are present.

**Layer 1: API Surface** (routes)
- `services/auth/src/routes/*.py`
- Imports from: `src/auth`, `src/models`, `src/schemas`, `src/db/database`, `src/services/*`, `src/utils/*`, `src/observability/logger`
- Does not import from other route modules

**Layer 2: Core Logic** (services, auth primitives)
- `services/auth/src/services/*.py` — business logic (api_key_service, identifier, credential_vault, encryption_service, oauth_state, token_lookup)
- `services/auth/src/auth/*.py` — crypto primitives (jwt_handler, api_key_handler, password)
- Imports from: `src/models`, `src/config`, `src/observability/logger`, `src/redis_client`
- Services do NOT import from routes

**Layer 3: Infrastructure** (models, db, config, redis, observability)
- `services/auth/src/models/*.py` — SQLModel ORM tables; imports only from `src/models/types`
- `services/auth/src/db/database.py` — engine and session factory; imports from `src/config`
- `services/auth/src/config.py` — settings; imports from `autom8y_config` SDK only
- `services/auth/src/redis_client.py` — Redis clients; imports from `src/config`, `src/observability/logger`
- `services/auth/src/observability/logger.py` — wraps `autom8y_log` SDK; no internal imports

**Middleware** sits outside this layering. It imports from `src/observability/logger` only (not from services or routes).

**Circular dependency note**: `src/auth/__init__.py` previously caused a circular import with `src/auth/dependencies.py`. This was resolved by having `dependencies.py` import directly from `src/auth/jwt_handler` rather than from the `src.auth` package barrel. This pattern is documented in the `__init__.py` docstring at line 9–10.

**Mothballed code**: `src/charter/`, `src/services/credential_vault.py`, `src/services/encryption_service.py`, `src/services/oauth_adapters/`, `src/workers/token_refresh_worker.py` are present on disk but not wired into `main.py`. Their models ARE still loaded by `src/models/__init__.py` (required for Alembic migrations).

## Entry Points and API Surface

**Application entry point**: `services/auth/src/main.py` — creates the FastAPI `app` instance with `lifespan` context manager, registers middleware, includes routers.

**Startup sequence** (via `lifespan`):
1. `settings.validate_required_secrets()` — asserts `DATABASE_URL`, `JWT_PRIVATE_KEY`, and `JWT_ALGORITHM=RS256`
2. `init_db()` — `SQLModel.metadata.create_all(engine)` for all registered tables
3. On shutdown: closes revocation client, rate limit client, identifier Redis client

**HTTP endpoints** (all active):

| Method | Path | Router | Purpose |
|---|---|---|---|
| GET | `/health` | main.py inline | Liveness probe — always 200, no I/O |
| GET | `/ready` | main.py inline | Readiness probe — checks DB + Secrets Manager |
| GET | `/health/deps` | main.py inline | Detailed deps — DB + Secrets Manager + Redis |
| GET | `/.well-known/jwks.json` | `routes/well_known` | JWKS public key for JWT validation |
| POST | `/auth/register` | `routes/auth` | User registration |
| POST | `/auth/login` | `routes/auth` | Login — returns access + refresh tokens |
| POST | `/auth/refresh` | `routes/auth` | Refresh access token |
| POST | `/auth/logout` | `routes/auth` | Logout — revokes refresh token |
| GET | `/auth/me` | `routes/auth` | Current user profile |
| GET | `/auth/business-lookup` | `routes/auth` | Look up business by email domain |
| POST | `/auth/password-reset` | `routes/auth` | Initiate password reset email |
| POST | `/auth/confirm-password-reset` | `routes/auth` | Confirm password reset with token |
| POST | `/auth/api-keys` | `routes/api_keys` | Create API key |
| GET | `/auth/api-keys` | `routes/api_keys` | List API keys |
| DELETE | `/auth/api-keys/{key_id}` | `routes/api_keys` | Revoke API key |
| POST | `/auth/api-keys/validate` | `routes/api_keys` | Validate API key (for consuming services) |
| POST | `/auth/roles` | `routes/rbac` | Create role (admin:roles permission required) |
| GET | `/auth/roles` | `routes/rbac` | List roles |
| GET | `/auth/roles/{role_id}` | `routes/rbac` | Get role |
| PUT | `/auth/roles/{role_id}` | `routes/rbac` | Update role |
| DELETE | `/auth/roles/{role_id}` | `routes/rbac` | Delete role |
| POST | `/auth/permissions` | `routes/rbac` | Create permission |
| GET | `/auth/permissions` | `routes/rbac` | List permissions |
| POST | `/auth/roles/{role_id}/permissions/{permission_id}` | `routes/rbac` | Assign permission to role |
| DELETE | `/auth/roles/{role_id}/permissions/{permission_id}` | `routes/rbac` | Remove permission from role |
| POST | `/auth/users/{user_id}/roles/{role_id}` | `routes/rbac` | Assign role to user |
| DELETE | `/auth/users/{user_id}/roles/{role_id}` | `routes/rbac` | Remove role from user |
| POST | `/internal/service-token` | `routes/internal` | Exchange service API key for service JWT |
| POST | `/internal/identifiers/resolve` | `routes/internal` | Batch identifier resolution (auth UUID <-> external ID) |
| POST | `/internal/revoke/user/{user_id}` | `routes/internal` | Revoke all tokens for user |
| POST | `/internal/revoke/token/{jti}` | `routes/internal` | Revoke specific token by JTI |
| GET | `/internal/revoke/status/{jti}` | `routes/internal` | Check token revocation status |
| POST | `/internal/admin/service-keys` | `routes/admin` | Create service API key |
| GET | `/internal/admin/service-keys` | `routes/admin` | List service API keys |
| POST | `/internal/admin/service-keys/{id}/rotate` | `routes/admin` | Rotate service key |
| DELETE | `/internal/admin/service-keys/{id}` | `routes/admin` | Revoke service key |

**Middleware stack** (applied in `main.py`, outer → inner):
1. `SecurityHeadersMiddleware` — sets HSTS, CSP, X-Frame-Options, etc.
2. `CORSMiddleware` — explicit allowed origins, no wildcard
3. `rate_limit_login` — in-memory IP-based rate limit on `POST /auth/login`
4. `error_handling_middleware` — catches unhandled exceptions, adds `request_id` to all requests

**Key exported interfaces** (used as FastAPI dependencies):
- `get_current_user` (`src/auth/dependencies.py`) — extracts and validates JWT Bearer token; returns payload dict with `sub`, `business_id`, `email`, `roles`, `permissions`
- `get_db()` (`src/db/database.py`) — yields synchronous `sqlmodel.Session`
- `get_async_db()` (`src/db/database.py`) — yields async `AsyncSession`

## Key Abstractions

**1. `Settings` / `get_settings()`**
- File: `services/auth/src/config.py`
- Extends `autom8y_config.Autom8yBaseSettings` (external SDK)
- Singleton via `@lru_cache` on `get_settings()`
- `JWT_ALGORITHM` is a `ClassVar[str] = "RS256"` — cannot be overridden by env var (ADR-0016)
- `REDIS_URL` is a computed property assembled from component fields (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`)
- `validate_required_secrets()` called on startup to fail-fast if misconfigured

**2. `User` and `UserBusiness` (SQLModel)**
- File: `services/auth/src/models/user.py`
- `User` maps to `users` table; `UserBusiness` is the many-to-many join to `businesses`
- Users hold Argon2id `password_hash`, optional `external_user_id` (maps to NHC `employees.dashboard_uuid`)
- Multi-tenancy: a user can belong to multiple businesses via `UserBusiness`

**3. `APIKey` (SQLModel)**
- File: `services/auth/src/models/api_key.py`
- `key_hash`: Argon2id hash (plaintext never stored)
- `key_prefix`: first 12 chars of key, indexed — enables O(1) prefix lookup (TDD-AUTH-HARDENING-001)
- `is_service_account`: distinguishes user keys from service-to-service keys
- `allow_multi_tenant`: gate for service tokens without a `business_id`

**4. `RevocationClient` and `RateLimitClient`**
- File: `services/auth/src/redis_client.py`
- Both implement fail-open policy: if Redis is unavailable, operations succeed (logged as warning)
- `RevocationClient`: manages `revoked:token:{jti}` and `active:user:{user_id}:tokens` keys
- `RateLimitClient`: sliding window via atomic INCR + EXPIRE on `ratelimit:apikey:{prefix}:{endpoint}`
- Both exposed as lazily-initialized global singletons via `get_revocation_client()` / `get_rate_limit_client()`

**5. `APIKeyService`**
- File: `services/auth/src/services/api_key_service.py`
- Supports both sync `Session` and async `AsyncSession` during migration period (ADR-ASYNC-DB-001)
- O(1) primary path: prefix-indexed DB lookup then Argon2id hash verification
- O(n) fallback: full scan for legacy keys without `key_prefix`

**6. `IdentifierService`**
- File: `services/auth/src/services/identifier.py`
- Bidirectional mapping: `auth_uuid` ↔ `external_business_id` / `external_user_id`
- Redis-cached (5-minute TTL, key: `cache:identifier:{entity_type}:{id_type}:{identifier}`)
- `guid_migrations` table lookup for numeric NHC GUIDs (FR-IDENTIFIER-007)

**7. JWT token model**
- Tokens are RS256-signed with `kid` header for JWKS rotation support
- Access token payload: `sub`, `iss`, `iat`, `exp`, `jti`, `business_id`, `business_name`, `email`, `roles`, `permissions`, `token_type="access"`
- Service token payload: additionally has `service_name`, `scope` (`"multi-tenant"` or `"single-tenant"`)
- Refresh tokens are stored in the `refresh_tokens` table; access tokens are tracked in Redis for revocation

**Design patterns observed**:
- Barrel re-export: `src/auth/__init__.py` re-exports from 4 sub-modules with explicit `__all__`
- Factory functions: `get_api_key_service(db)`, `get_identifier_service(db)` — not classes directly
- Fail-open: Redis errors silently allow tokens through (explicit ADR-0017)
- `ClassVar` for unoverridable config (`JWT_ALGORITHM`)

## Data Flow

**Primary path: User login → JWT issuance**

1. HTTP `POST /auth/login` (JSON body: `email`, `password`) arrives at `routes/auth.py`
2. `rate_limit_login` middleware checks IP-based in-memory counter before the request reaches the handler
3. Handler queries `users` table for `email` match
4. `verify_password()` (`src/auth/password.py`) — Argon2id verification
5. Handler queries `UserBusiness`, `UserRole`, `RolePermission`, `Permission` to build roles/permissions lists
6. `create_access_token_with_tracking()` (`src/auth/jwt_handler.py`) — RS256-signs token, tracks JTI in Redis via `RevocationClient.track_token()`
7. `create_refresh_token()` — signs refresh token; stored in `refresh_tokens` table
8. Response: `{access_token, refresh_token, token_type: "Bearer"}`

**Service-to-service path: API key → service JWT**

1. `POST /internal/service-token` (JSON body: `service_name`, optional `business_id`; `X-API-Key` header)
2. `check_rate_limit()` — Redis-backed sliding window per key prefix
3. `APIKeyService.find_by_prefix_and_verify()` — prefix index lookup + Argon2id verify
4. Checks `is_service_account=True`, `allow_multi_tenant` if `business_id=None`
5. `_create_service_token()` — RS256-signs service JWT with 30-minute TTL
6. Response: `{access_token, token_type: "Bearer", expires_in: 1800}`

**Configuration flow**:
- `Settings` (`pydantic-settings`) reads from env vars at process start
- `JWT_PRIVATE_KEY` → PEM string → `load_private_key()` → `RSAPrivateKey` → cached in `_signing_key` global (`jwt_handler.py`)
- `REDIS_HOST` components → `REDIS_URL` property → `RevocationClient`, `RateLimitClient`, `IdentifierService` all connect lazily on first use
- `DATABASE_URL` → `get_sync_engine()` / `get_async_engine()` — engines created once at module load, pooled (size=10, max_overflow=20)

**External service interactions**:
- PostgreSQL: via SQLAlchemy/SQLModel (sync for most routes, async for `internal/service-token`)
- Redis: for token revocation tracking, API key rate limiting, identifier caching (all fail-open)
- AWS Secrets Manager: optional, checked in readiness probe (`GET /ready`)
- SendGrid: via `src/mailer/service.py` (httpx direct — no SDK wrapper)
- `autom8y-telemetry` SDK: auto-instruments FastAPI metrics, SQLAlchemy, Redis for OpenTelemetry

## Knowledge Gaps

1. **`src/routes/admin.py` full endpoint logic**: Only the first 80 lines were read. The exact handlers for `rotate` and `delete` service key endpoints are undocumented here.
2. **Alembic migrations**: `migrations/` directory was not enumerated. The migration history and exact schema state at each version are not captured.
3. **`src/charter/` internals**: The mothballed Charter module (`client.py`, `audit.py`, `circuit_breaker.py`) was not read. Its internal design is not documented, as it is inactive per ADR-CHARTER-001.
4. **`src/services/credential_vault.py` internals**: Mothballed per ADR-VAULT-001; only the header was read to confirm status.
5. **`src/middleware/business_scope.py`**: Present on disk but not imported in `main.py`; its logic was not read.
6. **`client/` directory**: An `auth client CLI` exists (referenced in `pyproject.toml` ruff ignores) but was not enumerated or read.
7. **`runbooks/` and `scripts/`**: Operational tooling referenced in ruff ignores; not read.
