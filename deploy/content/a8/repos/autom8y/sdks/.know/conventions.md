---
domain: conventions
generated_at: "2026-03-16T15:40:19Z"
expires_after: "7d"
source_scope:
  - "./python/**/*.py"
  - "./python/**/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

### Exception Class Architecture

Every package defines a dedicated `errors.py` module at the package root. Each package's exception hierarchy follows a strict three-layer pattern:

**Layer 1: Package root base exception**
Inherits from `Exception` directly. All other exceptions in the package inherit from this base. Naming convention: `{Domain}Error` (e.g., `AuthError`, `TransportError`, `CacheError`, `HttpPrimitiveError`, `SlackError`, `GCalError`).

**Layer 2: Semantic mid-tier grouping** (present in larger packages)
Groups exceptions by retry semantics or domain boundary:
- `autom8y-auth`: `TransientAuthError` / `PermanentAuthError` (retry vs. no-retry grouping)
- `autom8y-core`: `TokenAcquisitionError` / `DataServiceError` (transport domain grouping)
- `autom8y-interop/_common`: `ServiceError` / `ServiceUnavailableError` / `ServiceTimeoutError`

**Layer 3: Concrete leaf exceptions**
Specific error conditions. Named `{Domain}{Condition}Error` (e.g., `InvalidTokenError`, `CircuitBreakerOpenError`, `CacheSerializationError`).

### ClassVar Attributes on All Exceptions

Every exception class carries two `ClassVar` attributes:
- `code: ClassVar[str]` — machine-readable SCREAMING_SNAKE_CASE string (e.g., `"INVALID_TOKEN"`, `"CIRCUIT_BREAKER_OPEN"`)
- `http_status: ClassVar[int]` — HTTP status code the error maps to

`from typing import ClassVar` and `if TYPE_CHECKING: from typing import ClassVar` are both used (the latter pattern appears in `autom8y-core` and `autom8y-auth`'s errors.py).

### Constructor Pattern

The base exception in each package accepts `message: str | None = None` and falls back to the class docstring:
```python
def __init__(self, message: str | None = None) -> None:
    self.message = message or self.__class__.__doc__ or "Error"
    super().__init__(self.message)
```
Leaf exceptions store domain-specific attributes as instance attrs before calling `super().__init__()`. Examples:
- `python/autom8y-core/src/autom8y_core/errors.py`: `SlotConflictError.__init__` stores `office_phone`, `start_datetime`, `conflict_reason` then calls `super().__init__(f"slot conflict for...")`
- `python/autom8y-auth/src/autom8y_auth/errors.py`: `PermissionDeniedError` stores `self.required`
- `python/autom8y-core/src/autom8y_core/errors.py`: `RetryExhaustedError` stores `self.attempts`, `self.last_error`

### Exception Chaining

`raise X from e` is the standard pattern for wrapping lower-level exceptions. This preserves `__cause__` and is tested explicitly (see `python/autom8y-sendgrid/tests/test_errors.py:157` and `python/autom8y-gcal/tests/test_errors.py:108`).

Do NOT use bare `raise X(str(e))` without `from e` — chaining is required.

Evidence: 19 instances of `raise ... from e` across `python/autom8y-core`, `python/autom8y-auth`, `python/autom8y-config`, `python/autom8y-ai`, `python/autom8y-cache`, `python/autom8y-meta`.

### Private Internal Exceptions

Internal-only exceptions are named with a leading underscore: `_RetryableRateLimitError` (`python/autom8y-core/src/autom8y_core/errors.py:141`), `_TransientFetchError` (`python/autom8y-auth/src/autom8y_auth/_jwks.py:38`). These are never exported from `__init__.py`.

### Logging at Error Boundaries

The codebase uses `autom8y-log` as the canonical logging package. The logging call signature is **structlog-style**: the first argument is an event name (snake_case string), followed by keyword arguments for structured fields.

```python
logger.warning("retry_waiting", attempt=1, delay_seconds=0.5)
logger.error("jwks_fetch_failed", url=url, status=status)
logger.info("service_started", version="1.0.0")
```

Module-level logger instantiation: `logger = get_logger(__name__)` (module-scope) or `_logger = get_logger(__name__)` (private, convention in `autom8y-auth`).

`get_logger` is imported from `autom8y_log`:
```python
from autom8y_log import get_logger
logger = get_logger(__name__)
```

22 source files use `get_logger(__name__)` at module level.

`exc_info=True` is used in warning calls where exception info should be captured without re-raising:
```python
logger.warning("Cold-tier set failed", exc_info=True)
```
(5 occurrences in `python/autom8y-cache/src/autom8y_cache/tiered.py`)

`logger.exception(...)` is used only when inside an `except` block and you want traceback capture with ERROR level (equivalent to `exc_info=True` at ERROR level). It is NOT used with `exc_info=True` redundantly.

### Packages Without Dedicated errors.py

- `autom8y-log`: No errors.py (uses stdlib exceptions + Python's built-in error types)
- `autom8y-telemetry`: No errors.py (telemetry init failures are logged, not raised)
- `autom8y-events`: No errors.py (simple publisher, no domain errors)
- `autom8y-devx-types`: No errors.py (type definitions only)

## File Organization

### src Layout

All packages use the `src/` layout. Structure: `python/{package-name}/src/{package_module}/`. Package names use hyphens (`autom8y-auth`), module names use underscores (`autom8y_auth`).

One exception: `autom8y-devx-types` uses module name `autom8_devx_types` (no trailing `y`).

### Standard File Boundaries

Every package follows a consistent set of top-level files:

| File | Purpose |
|------|---------|
| `__init__.py` | Public API surface — explicit `__all__`, docstring with usage examples |
| `errors.py` | Exception hierarchy for the package |
| `config.py` | Settings/configuration class (Pydantic-based) |
| `protocols.py` | Protocol definitions for type-safe interfaces |
| `testing/` | Sub-package for testing utilities, fixtures, factories |

**errors.py location variants**: Most packages have `errors.py` at the module root. `autom8y-cache` has `_errors.py` (private) at root and re-exports via `errors/__init__.py`. `autom8y-interop` has per-subdomain errors: `_common/errors.py`, `ads/errors.py`, `asana/errors.py`, `data/errors.py`.

### `__init__.py` Pattern

Every package `__init__.py` follows this structure:
1. Module docstring (multi-paragraph, includes: description, list of core components, code example)
2. `__version__` via `importlib.metadata.version()`
3. All public imports
4. Explicit `__all__` list (alphabetically sorted within logical groups)
5. Conditional imports for optional dependencies guarded with `try/except ImportError`

Example of optional backend pattern from `python/autom8y-cache/src/autom8y_cache/__init__.py`:
```python
try:
    from .backends.redis import RedisCacheProvider, RedisConfig
    __all__.extend(["RedisCacheProvider", "RedisConfig"])
except ImportError:
    pass
```

### Private vs. Public Module Convention

Files prefixed with `_` (single underscore) are internal implementation details not for direct import by consumers:
- `_batch.py`, `_errors.py`, `_settings.py` in `autom8y-cache`
- `_circuit_breaker.py`, `_jwks.py`, `_observability.py`, `_detection.py`, `_compat.py` in `autom8y-auth`
- `_retry_utils.py`, `_timeout.py` in `autom8y-http`
- `_mixin_base.py` in `autom8y-gcal` and `autom8y-sendgrid`

### Testing Sub-Package Structure

Every non-trivial package ships a `testing/` sub-package inside `src/` (co-located with the source, not alongside tests). This sub-package is registered as a `pytest11` entry point so fixtures are auto-discovered.

Files within `testing/`:
- `fixtures.py` — `@pytest.fixture` definitions, prefixed with `{package_name}_` to prevent collision
- `factories.py` — Plain functions and classes for constructing test objects (no pytest dependency)
- `mocks.py` — Mock implementations of protocols
- `stubs.py` — Stub implementations for integration boundaries
- `transports.py` — Mock HTTP transports
- `backends.py` — Test backend implementations

All fixture names are prefixed with the package short name: `auth_`, `cache_`, `core_`, `gcal_`, `sendgrid_`, etc. This is an explicit convention documented in fixture file docstrings.

### Submodule Patterns

When a concept grows beyond a single file, it gets promoted to a submodule directory with `__init__.py` acting as the re-export facade. Examples:
- `autom8y-cache/src/autom8y_cache/backends/` — multiple cache backend implementations
- `autom8y-cache/src/autom8y_cache/protocols/` — protocol definitions
- `autom8y-http/src/autom8y_http/resilience/` — resilience patterns
- `autom8y-auth/src/autom8y_auth/clients/` — auth client implementations
- `autom8y-interop/src/autom8y_interop/ads/`, `/asana/`, `/data/` — per-integration submodules

## Domain-Specific Idioms

### `from_env()` Classmethods

Every client class and config class that needs environment variable loading implements `@classmethod def from_env(cls) -> T`. This is the canonical way to instantiate configured clients in production code. It reads from environment variables via Pydantic Settings.

16 instances across: `TokenManager.from_env()`, `Client.from_env()`, `DataServiceClient.from_env()`, `AuthAdminClient.from_env()`, `RedisCacheProvider.from_env()`, `S3CacheProvider.from_env()`, `TieredCacheProvider.from_env()`, `CacheSettings.from_env()`, etc.

### `Autom8yBaseSettings` — Base for All Service Configuration

All service configuration classes inherit from `autom8y_config.Autom8yBaseSettings` (not directly from `pydantic_settings.BaseSettings`). This base class provides:
- Automatic secret URI resolution (`ssm://`, `secretsmanager://`, `env://`)
- Production URL guard (raises `ValueError` if production URLs used in LOCAL/TEST envs)
- `to_safe_dict()` — redacts `SecretStr` fields for safe logging
- `autom8y_env: Autom8yEnvironment` field with `AUTOM8Y_ENV` env var binding

Config classes use `SettingsConfigDict` with `env_prefix`:
```python
model_config = SettingsConfigDict(env_prefix="GCAL_", extra="ignore")
```

The `extra="ignore"` pattern is standard — unexpected env vars are silently dropped, not errors.

### `LoggerProtocol` — Dependency-Injected Logging

Components accept loggers via `logger: Any = None` or `logger: LoggerProtocol | None = None` constructor parameters rather than importing a module-level logger. The `autom8y_log.ensure_protocol()` function wraps any duck-typed logger into a protocol-conformant wrapper:

```python
from autom8y_log import ensure_protocol
self._logger: LoggerProtocol | None = ensure_protocol(logger) if logger else None
```

This pattern is used in `autom8y-http` (`ExponentialBackoffRetry`, `CircuitBreaker`, `Autom8yHttpClient`) and `autom8y-cache` (`HierarchyAwareResolver`).

Module-level loggers (used in `autom8y-meta`, `autom8y-auth`, `autom8y-telemetry`) are `get_logger(__name__)` — not configurable via DI, appropriate for implementation modules rather than library components.

### Protocol-First Interface Design

Every cross-boundary interface is expressed as a `typing.Protocol`. Protocol names end in `Protocol`: `CacheProvider`, `LoggerProtocol`, `HttpClientProtocol`, `GCalClientProtocol`, `AIClientProtocol`, `HierarchyResolverProtocol`, etc.

8 packages have dedicated `protocols.py` files. Private sub-protocols (internal structuring only) use leading underscore: `_FreeBusyProtocol`, `_EventsProtocol` in `autom8y-gcal`.

### Exception Hierarchy Documentation in Module Docstring

Error hierarchies are documented as ASCII trees in the module docstring of `errors.py`:
```python
"""
Error Hierarchy (v1.0):
    AuthError
        TransientAuthError          -- may resolve on retry
            JWKSFetchError
            CircuitOpenError
        PermanentAuthError          -- will NOT resolve on retry
            ...
"""
```
This is established in `python/autom8y-auth/src/autom8y_auth/errors.py:7`.

### `is_transient` Property on Base Exceptions

Auth exceptions implement `err.is_transient` as a property that returns `isinstance(self, TransientAuthError)`. This enables retry logic to check `err.is_transient` without importing the mid-tier class directly.

### `ErrorCodes` Mapping Class for Legacy Compatibility

When a new SDK replaces a legacy system (e.g., autom8y-db), a companion `ErrorCodes` class maps new exception types to old string codes. See `python/autom8y-auth/src/autom8y_auth/errors.py:288`.

### Versioned `__version__` in Every Package

```python
from importlib.metadata import PackageNotFoundError, version
try:
    __version__ = version("autom8y-log")
except PackageNotFoundError:
    __version__ = "0.0.0+dev"
```
This is in every `__init__.py`. The fallback `"0.0.0+dev"` is the standard sentinel for editable installs without a built wheel.

### `from __future__ import annotations` Usage

Present in `errors.py` files universally (enables `ClassVar` usage under `TYPE_CHECKING` guard). Not present as a blanket rule across all files — used selectively where forward references or TYPE_CHECKING guards require it.

### `if TYPE_CHECKING:` Guard for ClassVar Imports

```python
from __future__ import annotations
from typing import TYPE_CHECKING
if TYPE_CHECKING:
    from typing import ClassVar
```
This pattern appears in `python/autom8y-core/src/autom8y_core/errors.py` and `python/autom8y-auth/src/autom8y_auth/errors.py`. At runtime, `ClassVar` is only needed for Pydantic/type checker tooling, not for execution.

## Naming Patterns

### Package Names

- Distribution: kebab-case with `autom8y-` prefix: `autom8y-auth`, `autom8y-cache`, `autom8y-http`
- Import module: underscore version: `autom8y_auth`, `autom8y_cache`, `autom8y_http`
- Exception: `autom8y-devx-types` uses `autom8_devx_types` as module name (missing `y`)

### Error Class Naming

- Base: `{Domain}Error` — e.g., `AuthError`, `CacheError`, `GCalError`
- Concrete: `{Domain}{Condition}Error` — e.g., `GCalRateLimitError`, `CacheSerializationError`, `SendGridBounceError`
- Private: `_{Name}Error` — e.g., `_RetryableRateLimitError`, `_TransientFetchError`

### Configuration Classes

- Service configs that read from environment: `{Domain}Config` inheriting `Autom8yBaseSettings` (e.g., `GCalConfig`, `SlackConfig`, `SendGridConfig`, `StripeConfig`, `MetaConfig`, `AuthSettings`)
- Internal config (no env vars): `{Domain}Config` as plain dataclass or `BaseModel`-free class (e.g., `TieredConfig`, `DataServiceConfig`, `ReportConfig`)
- HTTP primitive configs that DO inherit `BaseSettings` directly: `RateLimiterConfig`, `RetryConfig`, `CircuitBreakerConfig`, `HttpClientConfig` (these use `env_prefix="RATE_LIMIT_"` etc.)

The naming distinction between `{Domain}Config` and `{Domain}Settings` is not enforced consistently — `AuthSettings` and `CacheSettings` use the `Settings` suffix while most use `Config`.

### Protocol Naming

`{Role}Protocol` — e.g., `CacheProvider`, `HttpClientProtocol`, `RateLimiterProtocol`, `RetryPolicyProtocol`, `HierarchyResolverProtocol`. Note: `CacheProvider` is the exception — it omits `Protocol` suffix (it was established before the naming convention solidified).

### Error Code Strings

`SCREAMING_SNAKE_CASE` strings. Examples: `"AUTH_ERROR"`, `"CIRCUIT_BREAKER_OPEN"`, `"SLOT_CONFLICT"`, `"TIMEZONE_NOT_CONFIGURED"`. These match the class name in SCREAMING_SNAKE_CASE without the word "Error" at the end.

### Testing Fixture Prefix

All `@pytest.fixture` names in `testing/fixtures.py` are prefixed with the SDK short name:
- `auth_*`: auth SDK fixtures
- `cache_*`: cache SDK fixtures
- `core_*`: core SDK fixtures
- `gcal_*`: GCal SDK fixtures

### Factory Function Naming

Factory functions in `testing/factories.py` use `create_*` prefix (not `make_*` or `build_*`):
- `create_test_token()`
- `create_user_claims()`
- `create_service_claims()`
- `create_jwks_document()`

### Env Variable Naming

Platform-wide vars use `AUTOM8Y_` prefix: `AUTOM8Y_ENV`, `AUTOM8Y_CACHE_ENABLED`, `AUTOM8Y_DATA_URL`. Service-specific API keys use `SERVICE_API_KEY` (generic canonical name). Package-specific vars use domain prefix: `GCAL_`, `RATE_LIMIT_`, `RETRY_`.

### File Naming Conventions

- Implementation modules: `snake_case.py` (e.g., `circuit_breaker.py`, `rate_limiter.py`, `token_manager.py`)
- Private implementation: `_snake_case.py` (e.g., `_circuit_breaker.py`, `_retry_utils.py`)
- Config: always `config.py`
- Errors: always `errors.py` (occasionally `_errors.py` when kept private)
- Protocols: always `protocols.py`
- Client class: `client.py` or `{name}_client.py` (e.g., `base_client.py`, `sync_client.py`)

## Knowledge Gaps

1. **autom8y-interop full structure**: Only partial observation — `ads/`, `asana/`, `data/` submodules observed, but `_common/` conventions and whether `autom8y-events` shares the same error pattern was not fully confirmed.

2. **autom8y-reconciliation narrative sub-package**: The `narrative/` sub-package inside `autom8y-reconciliation` was not explored — unknown whether it introduces additional idioms.

3. **autom8y-telemetry conventions submodule**: `conventions/` and `conventions/_data/` sub-packages were not read — possibly contain span naming conventions that constitute domain-specific idioms.

4. **autom8y-devx-types**: Minimal package with `_span.py`, `_types.py`, `_version.py`, `_narrative.py`. Contents not fully read — likely type aliases/stubs only.

5. **No observed use of `dataclasses`**: All configuration is Pydantic-based. Whether `@dataclass` appears anywhere in the codebase was not checked.

6. **`from __future__ import annotations` frequency**: Confirmed in `errors.py` files — not confirmed as universal across all module files.
