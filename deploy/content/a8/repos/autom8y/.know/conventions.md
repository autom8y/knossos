---
domain: conventions
generated_at: "2026-03-25T12:13:17Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "3fe30a4"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "ee0ce3dc0fda663f9d14c72bd112af1325e753ccadb47a217b6e98bdba62b7ba"
---

# Codebase Conventions

**Language**: Python 3.11+ (async-first, typed codebase)
**Package layout**: `src/` layout throughout -- `src/{package_name}/` under each `pyproject.toml`. Applies uniformly to all SDKs under `sdks/python/` and services under `services/`.

## Error Handling Style

### Exception Hierarchy Pattern

Every SDK package defines a dedicated `errors.py` at the package root. All 19 `errors.py` files in the SDK layer follow a single structural pattern:

```python
class {Package}Error(Exception):
    code: ClassVar[str] = "{PACKAGE}_ERROR"
    http_status: ClassVar[int] = 500

    def __init__(self, message: str, **context: Any) -> None:
        self.message = message
        self.context = context
        super().__init__(message)

    def to_dict(self) -> dict[str, Any]: ...  # present in newer packages

class {Package}APIError({Package}Error):
    code: ClassVar[str] = "{PACKAGE}_API_ERROR"
    http_status: ClassVar[int] = 502
```

Evidence: `sdks/python/autom8y-gcal/src/autom8y_gcal/errors.py` (lines 15-151), `sdks/python/autom8y-http/src/autom8y_http/errors.py` (lines 12-143), `sdks/python/autom8y-core/src/autom8y_core/errors.py`.

**Key variant**: `ClassVar[str] = "SCREAMING_SNAKE"` codes are mandatory on every subclass. `http_status: ClassVar[int]` is present on newer packages (gcal, interop, core) but absent on older ones (http primitive errors, config). `to_dict()` serialization method is present on newer SDK hierarchies.

**Programming contract violations** (`ValueError`, `RuntimeError`) are NOT wrapped in domain errors. They are raised directly for incorrect usage (uninitialized client, invalid arguments).

### `ClassVar` Import Pattern

`ClassVar` is always imported under `TYPE_CHECKING` guard in SDK error files to avoid runtime overhead:

```python
from __future__ import annotations
from typing import TYPE_CHECKING
if TYPE_CHECKING:
    from typing import ClassVar
```

Evidence: `sdks/python/autom8y-core/src/autom8y_core/errors.py`, `sdks/python/autom8y-gcal/src/autom8y_gcal/errors.py`, `sdks/python/autom8y-auth/src/autom8y_auth/errors.py`.

### Error Wrapping: `raise X from e` at Boundary Transitions

Cause-chaining (`raise NewError(...) from e`) is used at SDK-internal boundary transitions -- specifically when catching low-level transport exceptions (httpx, google-auth, network errors) and converting them to domain errors. 31 occurrences across 13 files.

### Error Propagation: Three-Phase Handler Pattern

All SDK clients follow this internal propagation pattern within the HTTP call path:
1. Attempt the call via the shared `_request()` method
2. Classify the HTTP status code into the appropriate domain error subclass
3. Re-raise with cause-chain (`from exc`)

### Error Handling at Boundaries: Structured Keyword Logging

At service boundaries, errors are logged with structured keyword arguments:

```python
logger.info("handler_invoked", days_back=days_back)
logger.warning("api_rate_limited", path=path, attempt=attempt, retry_after=e.retry_after)
```

Legacy f-string usage exists in `sdks/python/autom8y-auth/src/autom8y_auth/middleware.py` and `_jwks.py` -- the newer canonical style is `logger.{level}("event_name", key=value)`.

### Inter-Service Error Hierarchy (autom8y-interop)

Cross-service errors extend `ServiceError` / `ServiceUnavailableError` from `sdks/python/autom8y-interop/src/autom8y_interop/_common/errors.py`. These carry:
- `service_name: str` -- machine-readable service identifier
- `error_category: ErrorCategory` (StrEnum: `circuit_open`, `client_error`, `server_error`, `transport`, `unknown`)
- `time_remaining: float` -- seconds until circuit breaker retry

### Error Location Convention

| Location | Pattern |
|---|---|
| SDK errors | `src/{package}/errors.py` (dedicated file) |
| Service-internal errors | Either co-located in module or `errors.py` |
| Auth service | Multiple `exceptions.py` files per subdomain (`charter/exceptions.py`, `auth/exceptions.py`) |
| Cross-SDK errors | `autom8y-core` `errors.py` imported by services for transport errors |

### Private Internal Exceptions

Internal implementation exceptions prefixed with underscore:

```python
class _RetryableRateLimitError(TokenAcquisitionError):
    """Internal: 429 with a valid Retry-After header..."""
```

Evidence: `sdks/python/autom8y-core/src/autom8y_core/errors.py` line 141.

---

## File Organization

### SDK Package Layout

All 21 SDKs follow an identical layout under `sdks/python/autom8y-{name}/`:

```
src/autom8y_{name}/
  __init__.py        # Public API surface (re-exports, __version__)
  config.py          # Settings class (subclass of Autom8yBaseSettings)
  errors.py          # Exception hierarchy
  protocols.py       # Protocol definitions
  client.py          # Primary client class (async)
  models.py          # Pydantic models
  testing/           # Test fixtures and stubs
    __init__.py
    fixtures.py
    stubs.py / factories.py
  py.typed           # PEP 561 marker
```

### Private File Convention: `_` Prefix

Internal implementation files prefixed with `_`:
- `_circuit_breaker.py`, `_compat.py`, `_constants.py`, `_mixin_base.py`, `_observability.py`
- `_secrets.py`, `_dynamo.py` (contente-onboarding internals)

### Resource-Based Subpackaging: `resource_{name}.py` Pattern

For SDK clients with multiple API resource groups (GCal, SendGrid, Stripe), capabilities are split into `resource_{name}.py` files as `Mixin` classes, then composed via multiple inheritance:

```
resource_calendars.py   -> CalendarsMixin
resource_events.py      -> EventsMixin
client.py               -> GCalClient(FreeBusyMixin, EventsMixin, ChannelsMixin, CalendarsMixin)
```

Evidence: `sdks/python/autom8y-gcal/src/autom8y_gcal/client.py`, `sdks/python/autom8y-sendgrid/src/autom8y_sendgrid/`.

### Testing Subpackage: Mandatory and Structured

Every SDK package exports a `testing/` subpackage (28 instances found). Structure: `factories.py`, `fixtures.py`, `stubs.py` or `mocks.py`, `transports.py`.

### Service Handler Organization

**Lambda services** use flat-file organization: `handler.py`, `config.py`, `models.py`, `orchestrator.py`, `clients/`.

**Saga-based services** (contente-onboarding): each compensable step lives in its own file: `create_business.py`, `compensate_business.py`, `create_gcal.py`, `compensate_gcal.py`, etc.

### Config Singleton Pattern

Every service and SDK has a `config.py` with:

```python
from functools import lru_cache

@lru_cache
def get_settings() -> Settings:
    return Settings()
```

Evidence: 97 occurrences of `@lru_cache|get_settings()` across all config files.

### Import Boundary Enforcement (ruff TID)

Ruff enforces SDK-only imports via `[lint.flake8-tidy-imports.banned-api]` in `ruff.toml`. Services cannot directly import `loguru`, `structlog`, or raw `httpx`.

### Mothballed Code Pattern

Dead code preserved with `# MOTHBALLED:` comment blocks including feature name, ADR reference, and "DO NOT re-enable without..." guard. Do not clean up these blocks.

---

## Domain-Specific Idioms

### SDK-Only Import Layer

Hard enforced convention: services **must** use `autom8y_log.get_logger(__name__)` (never `loguru` or `structlog` directly), and `autom8y_http.Autom8yHttpClient` (never raw `httpx`).

### Structured Log Event Names

All log event names are `snake_case` string literals as the first positional argument. Additional context as keyword arguments:

```python
logger.info("auth_service_starting", service=settings.SERVICE_NAME, environment=env)
```

212 occurrences across 39 service files.

### Pydantic Settings with `AliasChoices` for Backward Compat

```python
autom8y_env: Autom8yEnvironment = Field(
    default=Autom8yEnvironment.LOCAL,
    validation_alias=AliasChoices("AUTOM8Y_ENV", "SERVICE_ENV"),
)
```

### Protocol Files for Dependency Injection

SDKs define `protocols.py` with `typing.Protocol` classes. Names follow `{Domain}Protocol` or `{Domain}ClientProtocol`.

### Async Context Manager Protocol for All Clients

Every SDK client that holds resources must be used as `async with`:

```python
async with GCalClient(config) as client:
    result = await client.free_busy.query(...)
```

Client methods call `_ensure_initialized()` first. 299 `async with` occurrences across 49 SDK files.

### `_MixinBase` Conditional Typing Pattern

For mypy strict compliance in mixin-based clients:

```python
if TYPE_CHECKING:
    class _HostClassInternals: ...
    _MixinBase = _HostClassInternals
else:
    _MixinBase = object
```

### Module-Level Client Singleton for Lambda Reuse

Lambda handlers initialize expensive clients at module level for warm start reuse.

### `@instrument_lambda` Decorator

All Lambda handlers use `@instrument_lambda` from `autom8y_telemetry.aws` for OTel trace injection.

### `StrEnum` for All API/Domain Constants

All enumerated string constants use Python 3.11+ `StrEnum`. 24 StrEnum classes in the SDK layer.

### `configure_logging()` + `get_logger(__name__)` at Module Level

175 occurrences of `from autom8y_log import|get_logger` across 52 files.

### `from __future__ import annotations` -- Universal First Import

Every non-`__init__` source file uses this as the first import line (PEP 563).

### `if TYPE_CHECKING:` for Import-Only Types

Used throughout the codebase to avoid circular imports and heavy runtime overhead.

### Keyword-Only Arguments on Complex Methods

Methods with 2+ parameters use `*,` to prevent positional mismatches.

### ADR Annotations

Architectural decisions cited inline: `# Auth uses a fail-open policy for Redis (ADR-0017)`.

---

## Naming Patterns

### Package/Module Names

| Context | Pattern | Example |
|---|---|---|
| SDK packages | `autom8y-{name}` (kebab) / `autom8y_{name}` (snake) | `autom8y-gcal` / `autom8y_gcal` |
| Internal service packages | `autom8_{name}` (not `autom8y_`) | `autom8_ads`, `autom8_devconsole` |
| Lambda/worker services | `{noun}-{noun}` (kebab) / `{noun}_{noun}` (snake) | `reconcile-ads` / `reconcile_ads` |
| New services | Plain name, no prefix | `contente_onboarding`, `calendly_intake` |

**Important inconsistency**: SDKs use `autom8y_` prefix; service packages use `autom8_` (missing the `y`). This is intentional product naming -- do not "fix" it.

### Class Names

| Pattern | Convention | Example |
|---|---|---|
| Error classes | `{Domain}Error`, `{Specific}Error` | `TransportError`, `SlotConflictError` |
| Settings classes | `Settings` (always just `Settings` per service) | `class Settings(Autom8yBaseSettings)` |
| Client classes | `{Domain}Client`, `Autom8y{Domain}Client` | `GCalClient`, `Autom8yHttpClient` |
| Protocol classes | `{Domain}Protocol`, `{Domain}ClientProtocol` | `GCalClientProtocol` |
| Pydantic models | `{Noun}`, `{Noun}Response`, `{Noun}Create` | `Event`, `EventCreate` |
| Mixin base | `_{Domain}MixinBase` | `_GCalMixinBase` |

### File Names

- Dedicated error files: `errors.py` (SDKs) or `exceptions.py` (auth service internal modules)
- Private/internal module prefix: `_` (single underscore)
- Test files: `test_{module}.py` (prefix, not suffix). QA: `test_qa_adversarial.py`
- Handler entrypoints: `handler.py`

### Variable/Attribute Naming

- Logger: `logger = get_logger(__name__)` (services) or `log = get_logger(__name__)` (SDKs)
- Settings: always via `get_settings()` function
- Constants: `UPPER_SNAKE_CASE`

### Acronym Conventions

- `GCal` (Google Calendar), `JWKS`, `JWT`, `DWD`, `SA`, `SSM`, `ARN`

### Auth Key Naming

Every service S2S API key follows a single derivation rule. DB key name: `{service}-service` (kebab-case). SM path: `autom8y/auth/service-api-keys/{db-key-name}`. Lambda env var: `SERVICE_API_KEY`. ECS env var: `AUTH_SERVICE_API_KEY`.

### Env Var Naming

- Canonical environment: `AUTOM8Y_ENV`
- Service API key: `SERVICE_API_KEY`
- Data service URL: `AUTOM8Y_DATA_URL`
- Config prefix: `{PACKAGE}_` (e.g., `CALENDLY_`, `STRIPE_`, `LOG_`)

---

## Knowledge Gaps

1. **Async vs sync conventions**: No explicit rule document about when sync clients are acceptable. `autom8y-http` has both `client.py` (async) and `sync_client.py`.

2. **Migration/Alembic conventions**: Auth service has `services/auth/migrations/` but naming and versioning conventions were not deeply observed.

3. **Service-layer error handling depth**: Auth service uses a different layout pattern (flat `src/` without `src/{package}/` nesting) -- its error handling patterns were not fully audited.

4. **`autom8y-saga` step protocol conventions**: Saga step and compensation result types were not fully read.

5. **`autom8y-telemetry` conventions system**: YAML-driven telemetry attribute registry observed but not deeply read.

6. **Prometheus metrics naming**: Metrics follow `{service}_{noun}_{verb}_{unit}` pattern but formal naming spec not located.
