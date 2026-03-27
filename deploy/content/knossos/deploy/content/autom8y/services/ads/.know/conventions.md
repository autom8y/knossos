---
domain: conventions
generated_at: "2026-03-16T00:14:42Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

**Language**: Python 3.11+
**Framework**: FastAPI + Pydantic v2 + pydantic-settings
**Build**: hatchling via `pyproject.toml`
**Lint/format**: ruff (line-length 100, rules E/F/I/UP/B/SIM/RUF); mypy strict mode

---

## Error Handling Style

### Error Creation: Typed Hierarchy with ClassVar Codes

All errors inherit from a single base class `AdsError(Exception)` defined in `services/ads/src/autom8_ads/errors.py`. The pattern is:

- Each error subclass declares `code: ClassVar[str]` (SCREAMING_SNAKE_CASE, always prefixed `ADS_` except `LAUNCH_IN_PROGRESS`) and `http_status: ClassVar[int]`.
- The base `__init__` accepts `message: str, **context: Any` and stores both as instance attributes. Subclasses with known context fields override `__init__` to use keyword-only arguments and pass context kwargs explicitly to `super().__init__`.
- `to_dict()` on the base class serializes `code`, `message`, `http_status`, and all context kwargs into a flat dict for logging and API responses.

```
AdsError (base)          code="ADS_ERROR",            http_status=500
  AdsValidationError     code="ADS_VALIDATION_ERROR", http_status=422
  AdsPlatformError       code="ADS_PLATFORM_ERROR",   http_status=502
  AdsTransientError      code="ADS_TRANSIENT_ERROR",  http_status=503
  AdsBudgetError         code="ADS_BUDGET_ERROR",     http_status=422
  AdsConfigError         code="ADS_CONFIG_ERROR",     http_status=500
  LaunchInProgressError  code="LAUNCH_IN_PROGRESS",   http_status=409
```

### Error Wrapping and Context Enrichment

No `%w`-style wrapping -- this is Python. Context is passed as kwargs to the error constructor and stored in `self.context`. Example from `LaunchInProgressError`:

```python
super().__init__(
    f"Launch already in progress for offer={offer_id}, platform={platform}",
    offer_id=offer_id,
    platform=platform,
)
```

For unexpected exceptions (`Exception as e`), `str(e)` is used to capture the message and `getattr(e, "context", None)` extracts structured context from `AdsError` subclasses transparently.

### Error Propagation

Two patterns coexist:

1. **Selective re-raise** (`launch/service.py`): A broad `except Exception as e` block at the pipeline boundary caches the error result first, then re-raises `LaunchInProgressError` and `AdsValidationError` unconditionally; all others are re-raised after logging. This keeps the cache consistent regardless of exception type.

2. **Return-not-raise** (`lifecycle/strategies/v2_meta.py`): The `V2MetaLaunchStrategy.execute()` catches all exceptions and returns a `LaunchResult(success=False, ...)` instead of raising. Partial entity creation state (campaign_id, ad_set_id, etc.) is preserved in the result. This is the "soft failure" pattern for the innermost layer.

Non-critical failures use silent degradation: `_try_persist()` and URL builder failures are caught and logged as warnings, never propagated.

### Error Handling at Boundaries

**HTTP boundary** (`api/launch.py`):
- `LaunchInProgressError` -> `HTTPException(409)` with structured detail dict
- `AdsValidationError` -> `HTTPException(422)` with `e.to_dict()`
- `AdsError` (other) -> `HTTPException(500)` with `e.to_dict()` + `logger.error`
- All conversions use `raise HTTPException(...) from e` to preserve cause chain

**Routing boundary** (`routing/router.py`):
- Raises `AdsConfigError` when no account matched -- a configuration-class error, not a platform error

**URL builder boundary** (`launch/service.py`):
- URL failures log `logger.warning("url_builder_failed", ...)` and silently omit the URL from the response (graceful degradation pattern)

### Logging Convention

Structured logging via `logging.getLogger(__name__)`. Log messages use **snake_case event strings** as the message (e.g., `"launch_complete"`, `"idempotency_cache_hit"`, `"url_builder_failed"`). All contextual data goes in the `extra={}` dict -- never interpolated into the message string. This enables log aggregation by event type.

---

## File Organization

### Package Layout

```
src/autom8_ads/
  __init__.py          # Package marker + __version__ only
  app.py               # FastAPI app factory + lifespan
  config.py            # AdsConfig (pydantic-settings BaseSettings)
  dependencies.py      # FastAPI Depends() injectors
  errors.py            # Full error hierarchy
  api/
    __init__.py        # Empty
    health.py          # GET /health router
    launch.py          # POST /api/v1/launches, DELETE /api/v1/launches/... router
  clients/
    __init__.py        # Empty
    data.py            # StubDataServiceClient
  launch/
    __init__.py        # Empty
    idempotency.py     # LaunchIdempotencyCache + LaunchCacheEntry
    mapper.py          # OfferPayloadMapper
    service.py         # LaunchService (main orchestrator)
  lifecycle/
    __init__.py        # Empty
    factory.py         # AdFactory (thin delegator)
    strategies/
      __init__.py      # Empty
      base.py          # LaunchStrategy Protocol
      v2_meta.py       # V2MetaLaunchStrategy (only implementation)
  models/
    __init__.py        # Empty
    base.py            # AdsModel (frozen Pydantic BaseModel)
    enums.py           # All StrEnum definitions
    launch.py          # LaunchContext, LaunchResult
    offer.py           # OfferPayload, LaunchResponse
    targeting.py       # TargetingSpec
  platforms/
    __init__.py        # Empty
    protocol.py        # AdPlatform Protocol, DataServiceProtocol
  routing/
    __init__.py        # Empty
    config.py          # AccountRule, AccountRouterConfig
    router.py          # AccountRouter
  urls/
    __init__.py        # Empty
    meta.py            # MetaPlatformConfig, MetaUrlBuilder
```

### File Organization Rules

- **One concern per file, named for that concern**: `errors.py` holds all errors; `config.py` holds config; `enums.py` holds all enums.
- **All subpackage `__init__.py` files are empty** -- no re-exports, no barrel files. Callers import directly from the module (e.g., `from autom8_ads.models.enums import Platform`).
- **Root `__init__.py`** only declares `__version__`.
- **`app.py`** is the application factory; it wires singletons in `lifespan()` and calls `include_router()` for each API router.
- **`dependencies.py`** is the sole location for `Depends()` injectors. All injectors read from `request.app.state` and use `cast()` for type safety.
- **Protocols live in `platforms/protocol.py`** -- the abstract interface layer is a dedicated module, not mixed into implementations.
- **Config models live in `routing/config.py`** alongside their consumer (`routing/router.py`) rather than in `models/`. The distinction: `models/` holds domain objects; `routing/` holds both the config schema and the router that uses it.
- **Tests mirror source structure**: `tests/api/`, `tests/launch/`, `tests/models/` etc. match `src/autom8_ads/api/`, `src/autom8_ads/launch/`, etc.

### Generated/Special Files

No generated files. No `py.typed` marker present (mypy strict is configured via `pyproject.toml` directly).

---

## Domain-Specific Idioms

### 1. The `AdsModel` Frozen Base

All domain value objects inherit from `AdsModel` (defined in `models/base.py`), not directly from `pydantic.BaseModel`. `AdsModel` is always `frozen=True, extra="ignore", from_attributes=True`. This is the project's immutable value object pattern.

**Exception**: `OfferPayload` and `LaunchResponse` (in `models/offer.py`) inherit directly from `pydantic.BaseModel` with `frozen=True, extra="forbid"` -- the inbound/outbound wire models are more strict (`extra="forbid"`) vs. domain objects (`extra="ignore"`).

### 2. `StrEnum` for All Enumerations

All enums use `StrEnum` (Python 3.11 stdlib), not `str, Enum`. Values are lowercase strings (e.g., `Platform.META = "meta"`). This means enum values serialize to strings without special handling and can be compared directly to string literals.

### 3. Protocol-Based Adapters

External integrations (ad platforms, data service) are expressed as `typing.Protocol` classes in `platforms/protocol.py`. `AdPlatform` is `@runtime_checkable`. The concrete `V2MetaLaunchStrategy` and `LaunchService` depend on protocols, not concrete adapters. This allows `AsyncMock()` in tests without a real platform client.

### 4. app.state Singleton Pattern

All singletons are created in `app.py`'s `lifespan()` and stored on `app.state`. FastAPI's `Depends()` injectors in `dependencies.py` retrieve them via `request.app.state`. This is the project's DI pattern -- no DI framework, no global singletons.

### 5. model_copy for Immutable Updates

Frozen Pydantic models are updated using `.model_copy(update={...})`:

```python
ctx = ctx.model_copy(update={"account_id": account_id})
```

This is the only sanctioned way to produce a modified copy of an immutable domain object.

### 6. Idempotency Cache Key Format

Cache keys are `f"{offer_id}:{platform.value}"` -- a colon-separated composite of Asana GID and lowercase platform string. This is a project-specific pattern: the composite key is constructed at point of use, not stored on a domain object.

### 7. Keyword-Only Arguments for URL Builder Methods

All `MetaUrlBuilder` public methods use keyword-only arguments (`*, account_id, office_phone, ...`). This prevents positional argument errors for calls with many string parameters. The `MetaUrlBuilder` class itself accepts `MetaPlatformConfig` (a thin config wrapper), not raw strings.

### 8. ADR References in Code

Decision references like `(ADR-ADS-002)` appear in docstrings and comments to anchor code constraints to design decisions. These are not enforced at runtime -- they are documentation anchors for agents and developers.

### 9. Filter Encoding: Meta's Proprietary Format

`MetaUrlBuilder` uses `%1E` as field/operator/value separator and `%1D` as filter separator within Meta's URL filter scheme. The bullet character `\u2022` prefixes vertical_key in campaign name filters. These are not standard URL encoding -- they are Meta-platform-specific.

### 10. Graceful Degradation Pattern

Non-critical steps in the pipeline (URL construction, data persistence) are wrapped in their own `try/except Exception`, log a warning, and allow the main operation to succeed. The function `_try_persist()` exemplifies this: it is `async`, called with `await`, but its failures are swallowed silently.

---

## Naming Patterns

### Type Names

- **Error classes**: `Ads{Concern}Error` pattern (`AdsValidationError`, `AdsPlatformError`, `AdsBudgetError`), except `LaunchInProgressError` which is named for the condition, not the error category.
- **Config/settings classes**: `{Domain}Config` pattern (`AdsConfig`, `AccountRouterConfig`, `MetaPlatformConfig`).
- **Model classes**: Named for what they represent, not their base class: `OfferPayload`, `LaunchContext`, `LaunchResult`, `LaunchResponse`, `TargetingSpec`, `AccountRule`.
- **Protocol classes**: Named for the capability, not "Interface" or "Abstract": `AdPlatform`, `DataServiceProtocol`, `LaunchStrategy`.
- **Implementation classes**: Named for what they do: `LaunchService`, `AccountRouter`, `OfferPayloadMapper`, `MetaUrlBuilder`, `LaunchIdempotencyCache`, `AdFactory`, `V2MetaLaunchStrategy`.

### Variable and Attribute Names

- Private instance attributes use single underscore prefix: `self._config`, `self._platform`, `self._router`, `self._cache`.
- Local variables in pipeline steps are short and domain-meaningful: `ctx` for `LaunchContext`, `platform` for `Platform` enum value, `result` for `LaunchResult`.
- Dictionary keys in `extra={}` log dicts use snake_case matching the field names they correspond to (e.g., `"offer_id"`, `"campaign_id"`).

### Module/Package Names

- Package: `autom8_ads` (underscore, not hyphen; project name is `autom8-ads` in pyproject but the Python package is `autom8_ads`).
- Subpackages use noun plural or functional nouns: `models`, `api`, `clients`, `launch`, `lifecycle`, `platforms`, `routing`, `urls`.
- File names are singular and descriptive: `service.py`, `mapper.py`, `idempotency.py`, `factory.py`, `protocol.py`, `router.py`.

### Constant Naming

- Module-level regex constants: `SCREAMING_SNAKE_CASE` (e.g., `E164_PATTERN`).
- Class-level URL constants: `SCREAMING_SNAKE_CASE` class attributes (e.g., `MetaUrlBuilder.ADSETS_BASE`, `MetaUrlBuilder.ADS_COLUMNS`).
- Error codes: `SCREAMING_SNAKE_CASE` strings, domain-prefixed (`ADS_VALIDATION_ERROR`).

### Function Naming

- Private helper methods: `_snake_case` with single leading underscore (`_build_response`, `_classify_error`, `_try_persist`, `_evict_expired`, `_is_expired`).
- Static helpers: also `_snake_case` with `@staticmethod` (`_failure_step`, `_campaign_name_contains`).
- FastAPI route handler functions: snake_case, verb-first (`launch_ads`, `clear_launch_cache`, `health`).
- `from_*` classmethod factory pattern: `LaunchResponse.from_launch_result(...)`.

### Deviations from Standard Python

- `from __future__ import annotations` is present in every file -- this enables postponed evaluation of annotations for forward references, consistent across all modules.
- No `Optional[T]` -- all files use `T | None` (Python 3.10+ union syntax, enabled by `from __future__ import annotations`).
- `ClassVar[str]` and `ClassVar[int]` on Pydantic model/exception class bodies use `TYPE_CHECKING` guard in `errors.py` to avoid runtime import overhead.

---

## Knowledge Gaps

- No concrete `AdPlatform` implementation is present in this service (the Meta SDK adapter lives elsewhere). The protocol contract is documented but the implementation conventions of the real adapter cannot be observed here.
- `autom8y-config`, `autom8y-log`, `autom8y-auth` SDK conventions are not observable from this service alone.
- The `routing/config.py` pattern (YAML/JSON config loading) is stubbed -- `account_routing_config` env var exists in `AdsConfig` but the loading code is not yet implemented; only the in-memory default is wired.
- No `Makefile`, `Dockerfile`, or CI configuration is present in this service directory -- deployment and build conventions cannot be documented from source alone.
