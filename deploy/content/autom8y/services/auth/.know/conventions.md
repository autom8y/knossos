---
domain: conventions
generated_at: "2026-03-16T00:02:18Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

### Philosophy

The codebase follows a **two-tier error hierarchy**: domain-layer exceptions propagate up to the route layer, where they are caught and converted to `HTTPException`. The domain layer never speaks HTTP; the route layer never catches raw database or crypto exceptions directly.

### Custom Exception Hierarchy

Every subsystem defines a base exception that inherits from `Exception`, followed by typed subclasses. The base exception carries both a human `message` and a machine-readable `code` string. This pattern is used uniformly:

```
Exception
  -> CharterException(message, code="CHARTER_ERROR")         services/auth/src/charter/exceptions.py
       -> ValidationError(code="VALIDATION_ERROR")
       -> QueryError(code="QUERY_ERROR")
       -> ConflictError(code="CONFLICT")
       -> CrossBusinessError(code="CROSS_BUSINESS_VIOLATION")
       -> PrivilegeEscalationError(code="PRIVILEGE_ESCALATION")
       -> CircuitBreakerOpen(code="CIRCUIT_BREAKER_OPEN")
  -> CredentialVaultError(message, code="CREDENTIAL_ERROR")   src/services/credential_vault.py
       -> CredentialNotFoundError(code="CREDENTIAL_NOT_FOUND")
       -> ProviderNotFoundError(code="PROVIDER_NOT_FOUND")
       -> CredentialExpiredError(code="CREDENTIAL_EXPIRED")
       -> CredentialRevokedError(code="CREDENTIAL_REVOKED")
  -> OAuthStateError                                          src/services/oauth_state.py
       -> InvalidStateError, StateExpiredError, StateConsumedError
  -> EncryptionError                                          src/services/encryption_service.py
       -> KMSError, DecryptionError
  -> OAuthAdapterError                                        src/services/oauth_adapters/base.py
       -> TokenExchangeError, TokenRefreshError
  -> JWTError                                                 src/auth/jwt_handler.py
       -> TokenExpiredError, InvalidTokenError
  -> APIKeyError                                              src/auth/api_key_handler.py
  -> RevocationError                                          src/redis_client.py
  -> RateLimitError                                           src/redis_client.py
```

Each subclass has a fixed `code` string set at instantiation, not at the call site.

### Error Creation Pattern

Service-layer exceptions are raised with a descriptive message and the code is pre-wired in the subclass `__init__`:

```python
# From src/charter/exceptions.py
class ConflictError(CharterException):
    def __init__(self, message: str):
        super().__init__(message, code="CONFLICT")
```

At the call site, only the human message is passed: `raise ConflictError("Role already exists")`.

### Route Boundary: Exception-to-HTTPException Translation

Routes catch typed domain exceptions and re-raise as `HTTPException`. The dominant pattern in the charter routes (`src/routes/charter.py`) is explicit per-exception-type mapping:

```python
except ValidationError as e:
    raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))
except ConflictError as e:
    raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=str(e))
except QueryError as e:
    raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, ...)
```

Routes in `src/routes/auth.py`, `src/routes/rbac.py`, and `src/routes/credentials.py` use a mixed pattern: they raise `HTTPException` directly inline (no intermediate domain exception) for simpler validation failures, but catch domain exceptions for service operations. Total inline `HTTPException` raises: 168 occurrences across 11 route files.

### HTTPException Detail Structure

Two conventions coexist. The preferred convention (used in auth routes, dependencies):

```python
raise HTTPException(
    status_code=status.HTTP_409_CONFLICT,
    detail={"error": "email_exists", "message": "Email already registered"},
)
```

The fallback convention (charter routes, simpler cases):

```python
raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))
```

The `{"error": "...", "message": "..."}` dict form dominates in auth routes (75 occurrences across 8 files). Charter routes use `detail=str(e)` (9 occurrences in that file alone).

### Error Propagation in Services

Services propagate exceptions upward; they do not swallow them. The notable exception to this rule is audit logging (`src/utils/audit.py`): `log_event()` catches `Exception` and returns `None` rather than propagating — audit failures must never block business logic. Redis/token operations follow a **fail-open policy**: errors are logged at `warning` level, not raised.

### Logging at Error Boundaries

All error handling is paired with structured logger calls. The module-level logger is always:

```python
logger = get_logger(__name__)
```

Present in 30 of 30 observed service/route/utility files. Structured logging calls use keyword arguments, not f-strings:

```python
logger.error("database_initialization_failed", error=str(e))
logger.warning("token_tracking_failed", user_id=user_id, jti=jti, error=str(e))
```

### LBYL vs EAFP

The codebase is EAFP (Easier to Ask Forgiveness than Permission) at service boundaries and LBYL (Look Before You Leap) at business logic. Database existence checks (`db.exec(select(X).where(...)).first()`) are LBYL guard clauses at the top of route handlers. Service-layer crypto and I/O operations use try/except around the entire operation.

## File Organization

### Top-Level Source Layout

All production source code lives under `services/auth/src/` (installed as the `src` package per `pyproject.toml`). The package structure:

```
src/
  auth/          # JWT, API key, password utilities + FastAPI dependency
  charter/       # MOTHBALLED RBAC module (frozen, not registered)
  db/            # Engine, session factory, seeds
  mailer/        # Email service
  middleware/    # FastAPI middlewares
  models/        # SQLModel table definitions (one model per file)
  observability/ # Logging wrapper
  routes/        # FastAPI routers (one route group per file)
  schemas/       # Pydantic request/response schemas
  services/      # Business logic services
    oauth_adapters/  # Adapter sub-package for OAuth providers
  utils/         # Cross-cutting utilities (audit, brute-force, rate-limit)
  workers/       # Background workers
  config.py      # Pydantic Settings singleton
  health_models.py
  jwks.py        # JWKS/JWK key management
  main.py        # FastAPI app factory + lifespan
  redis_client.py # Redis client + RevocationClient
```

### Per-Package File Responsibilities

**`src/models/`** — One file per SQLModel table. Each file contains exactly one or two tightly related `SQLModel` table classes. The `__init__.py` is a barrel re-export file that imports all models and lists them in `__all__`. The shared column type `TIMESTAMPTZ` lives in `src/models/types.py`.

**`src/schemas/`** — Pydantic `BaseModel` request/response shapes separated from SQLModel tables. Route-specific schemas that are small in scope may also live inline in the route file (e.g., `src/routes/credentials.py` defines `StoreCredentialRequest` locally; `src/routes/internal.py` defines `ServiceTokenRequest` locally). The central schemas live in `src/schemas/auth.py` (auth routes), `src/schemas/admin.py`, `src/schemas/identifier.py`.

**`src/routes/`** — One router per domain area. Each file creates `router = APIRouter(prefix="...", tags=["..."])` at module level. The `__init__.py` is a barrel that imports active routers and gates mothballed routers with commented-out imports and MOTHBALLED comment blocks.

**`src/auth/`** — The auth utilities package. Barrel re-export in `__init__.py`. Note: this package avoids importing from its own `__init__.py` internally (circular import risk documented in the `__init__.py` docstring).

**`src/services/`** — Business logic. Factory functions follow the `get_{service_name}_service(db)` pattern. The `__init__.py` re-exports service classes and their exceptions for external consumers.

**`src/utils/`** — Cross-cutting tools that are not service-layer. Three files: `audit.py` (audit log writing), `brute_force.py` (in-memory login protection), `rate_limit.py` (Redis-backed rate limit for API keys).

### Config and Settings

All configuration lives in `src/config.py` as a single `Settings` class inheriting `Autom8yBaseSettings`. Settings are accessed via `get_settings()` (an `@lru_cache`-decorated factory function). Settings are instantiated at module level in files that need them:

```python
settings = get_settings()
```

This is the canonical pattern — no dependency injection for settings.

### MOTHBALLED Convention

Frozen subsystems are preserved in-place but annotated with a docstring banner:

```python
"""
================================================================================
MOTHBALLED - DO NOT USE
================================================================================
...
```

The `src/routes/__init__.py` gates mothballed routers via commented-out imports with inline ADR citations. This is the authoritative guard.

### `__init__.py` Export Pattern

Every package has an `__init__.py` with an explicit `__all__` list. Barrel re-exports are the norm. The rule: if something is exported from a package `__init__.py`, it belongs in `__all__`.

## Domain-Specific Idioms

### 1. Structured Logging via `get_logger(__name__)`

Every module that logs instantiates a module-level logger:

```python
from src.observability.logger import get_logger
logger = get_logger(__name__)
```

This wrapper binds `request_id` from context automatically (set by middleware). Consumers never use Python's `logging` module directly. Log calls use keyword arguments for structured fields, not positional f-strings. The log event name is a `snake_case` string as the first positional argument:

```python
logger.info("auth_service_starting", service=settings.SERVICE_NAME, environment=settings.autom8y_env.value)
logger.error("database_initialization_failed", error=str(e))
logger.warning("business_scope_violation", path=path, token_business_id=..., ...)
```

### 2. `AuditEventType` Constants Class

Audit events are identified by dot-separated strings (`"user.login_success"`, `"api_key.created"`). These strings are collected as class attributes on `AuditEventType` in `src/utils/audit.py`:

```python
class AuditEventType:
    USER_LOGIN_SUCCESS = "user.login_success"
    API_KEY_CREATED = "api_key.created"
```

All audit event logging must use these constants, never raw strings.

### 3. Fail-Open Pattern for Redis

Redis-dependent operations (token tracking, revocation checks, rate limiting) follow a fail-open policy. The import is guarded:

```python
try:
    import redis.asyncio as redis
    REDIS_AVAILABLE = True
except ImportError:
    redis = None
    REDIS_AVAILABLE = False
```

Operations catch `Exception` broadly and log a `warning` rather than propagating. This ensures Redis outages degrade gracefully without blocking auth.

### 4. `get_{name}_service()` Factory Pattern

Services expose an async factory function that accepts a `db` session and returns a service instance. This is the canonical dependency injection hook for FastAPI:

```python
async def get_encryption_service() -> EncryptionService:  # src/services/encryption_service.py
async def get_credential_vault_service(...)               # src/services/credential_vault.py
async def get_api_key_service(db: Session | AsyncSession) # src/services/api_key_service.py
```

These are registered with `Depends(...)` in route handlers.

### 5. `TIMESTAMPTZ` Custom Type

All datetime columns use the module-level constant from `src/models/types.py`:

```python
sa_type=TIMESTAMPTZ  # = DateTime(timezone=True)
```

Never use `DateTime()` directly in model field definitions. All datetimes use `UTC`: `datetime.now(UTC)`.

### 6. ADR and PRD Citation Convention

Code comments cite design decisions using a stable reference pattern: `ADR-NNNN`, `ADR-DOMAIN-KEYWORD`, `TDD-NAME`, `PRD-NAME`. These appear in inline comments, docstrings, and MOTHBALLED banners. 84 occurrences across 29 files. They serve as audit trail pointers, not live links.

### 7. Circuit Breaker Pattern (Charter module)

The `CircuitBreaker` class in `src/charter/circuit_breaker.py` wraps the Charter service client. A global singleton instance is constructed at module load (`_charter_circuit_breaker = CircuitBreaker(...)`). It raises `CircuitBreakerOpen` after failure threshold, caught at the route boundary and returned as HTTP 503.

### 8. Business-Scoped Multi-Tenancy Everywhere

`business_id` is a first-class field on all domain entities (User, APIKey, RefreshToken, etc.) and is required in JWT payloads. The middleware `src/middleware/business_scope.py` enforces that the `business_id` in the URL or body matches the JWT claim. Every service operation that modifies data validates `business_id` before writing.

### 9. `ClassVar` for Immutable Configuration

Configuration fields that must not be overridden by environment variables are declared as `ClassVar`:

```python
JWT_ALGORITHM: ClassVar[str] = "RS256"
```

This prevents a stray `JWT_ALGORITHM=HS256` env var from downgrading the signing algorithm.

## Naming Patterns

### Classes

- SQLModel tables: `PascalCase`, named as singular nouns (`User`, `Business`, `APIKey`, `RefreshToken`). Table name is set explicitly via `__tablename__` using `snake_case_plural`.
- Pydantic schemas: `PascalCase` with `Request`/`Response` suffix (`RegisterRequest`, `TokenResponse`, `ErrorResponse`). Schemas that are purely internal views have no suffix (`UserResponse` exception: all user-facing response schemas end in `Response`).
- Exception classes: `PascalCase`, end in `Error` or `Exception` (`CredentialVaultError`, `TokenExpiredError`). The `code` string uses `UPPER_SNAKE_CASE`.
- Service classes: `PascalCase + Service` suffix (`APIKeyService`, `CredentialVaultService`, `IdentifierService`, `EmailService`). The one worker class: `TokenRefreshWorker`.

### Functions and Methods

- PEP 8 `snake_case` throughout.
- FastAPI dependency factories: `get_{noun}` pattern (`get_db`, `get_settings`, `get_current_user`, `get_charter_client`).
- Service factory functions: `get_{name}_service` (e.g., `get_api_key_service`, `get_encryption_service`).
- Private helpers: single leading underscore (`_get_signing_key`, `_sync_engine`, `_track_token_async`, `_extract_business_id_from_request`).
- Lifecycle/startup functions in `main.py`: descriptive verbs (`init_db`, `close_rate_limit_client`).

### Variables and Constants

- Module-level constants: `UPPER_SNAKE_CASE` (`REVOKED_TOKEN_KEY`, `CACHE_TTL_SECONDS`, `UNSCOPED_ENDPOINTS`).
- Module-level singleton instances: `snake_case` with leading underscore (`_sync_engine`, `_signing_key`, `_charter_circuit_breaker`).
- JWT payload fields: `snake_case` (`business_id`, `token_type`, `jti`).
- Log event names: `lower_snake_case` string literals used as first positional arg (`"auth_service_starting"`, `"business_scope_violation"`).

### Files and Modules

- Source files: `snake_case.py`.
- Route files: named after the domain they handle (`auth.py`, `rbac.py`, `api_keys.py`, `internal.py`).
- Model files: named after the singular entity (`user.py`, `api_key.py`, `refresh_token.py`).
- Service files: named after the service (`credential_vault.py`, `encryption_service.py`, `token_lookup.py`).
- Utility files: named after function (`audit.py`, `brute_force.py`, `rate_limit.py`).
- Exception files: always named `exceptions.py`, co-located with their subsystem package.

### Acronym Conventions

- `JWT`, `API`, `JWKS`, `RS256`, `KMS`, `RBAC`, `CORS`, `TTL` — all-caps in comments and class/variable names when used as standalone acronyms.
- In identifiers they collapse to their lowercase form when combined: `jwt_handler.py`, `api_key.py`, `jwks.py`.
- `ADR` prefix for Architecture Decision Records; `TDD` prefix for Technical Design Documents; `PRD` prefix for Product Requirements Documents — used in code comments as reference tokens.

## Knowledge Gaps

1. **`src/redis_client.py` full structure** — Only the first 60 lines were read. The `RateLimitClient` class at line 330 and `RateLimitError` were discovered via grep but not fully read. Rate-limit key patterns and full client API are partially unknown.
2. **`src/services/oauth_adapters/asana.py` full error handling** — Observed outer structure and 4 except clauses; full retry/backoff logic not read.
3. **`src/workers/token_refresh_worker.py`** — Outer structure observed; worker lifecycle and error recovery policy not fully documented.
4. **Exact inline schema location policy** — Both `src/schemas/auth.py` and inline definitions in route files (credentials, internal) are used. The boundary rule for when to define inline vs. in `schemas/` is not explicitly documented in code comments.
5. **`src/auth/api_key_handler.py` full API** — The `GeneratedAPIKey` named tuple and full key generation pipeline not read; only class/function names were observed.
