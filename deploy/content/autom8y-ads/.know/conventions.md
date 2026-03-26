---
domain: conventions
generated_at: "2026-03-01T12:42:56Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "762ed0e"
confidence: 0.92
format_version: "1.0"
---

# Codebase Conventions

## Error Handling Style

### Error Hierarchy

All domain errors descend from a single base class `AdsError` defined in `src/autom8_ads/errors.py`. The hierarchy is:

```
AdsError (base, HTTP 500, code="ADS_ERROR")
├── AdsValidationError (HTTP 422, code="ADS_VALIDATION_ERROR")
├── AdsConfigError (HTTP 500, code="ADS_CONFIG_ERROR")
├── AdsExternalServiceError (HTTP 503, code="ADS_EXTERNAL_SERVICE_ERROR")
└── AdsPlatformError (HTTP 502, code="ADS_PLATFORM_ERROR")
    ├── AdsTransientError (HTTP 503, code="ADS_TRANSIENT_ERROR")
    ├── AdsNotFoundError (HTTP 404, code="ADS_NOT_FOUND")
    └── AdsBudgetError (HTTP 422, code="ADS_BUDGET_ERROR")
```

### Error Creation Pattern

Every error class carries two `ClassVar` fields -- `code` and `http_status` -- which are the machine-readable error code and the HTTP status it maps to. Constructors accept a `message: str` plus `**context: Any` kwargs that become structured metadata on the instance.

Specialist constructors use keyword-only arguments for required context:

```python
# src/autom8_ads/errors.py lines 38-41
def __init__(self, *, field: str, reason: str, **context: Any) -> None:
    super().__init__(f"{field}: {reason}", field=field, reason=reason, **context)
```

Platform errors accept `original: Exception` as the wrapped upstream exception:

```python
# src/autom8_ads/errors.py lines 50-58
def __init__(self, platform: Platform, original: Exception, **context: Any) -> None:
    self.platform = platform
    self.original = original
    super().__init__(str(original), platform=platform.value, **context)
```

### Error Serialization

`AdsError.to_dict()` produces a flat dict with `code`, `message`, `http_status`, plus all `**context` keys. This is the canonical serialization shape used by exception handlers.

### Error Propagation at HTTP Boundary

`app.py` registers two FastAPI exception handlers. `AdsTransientError` gets its own handler that adds `Retry-After` header. `AdsError` is the catch-all. Both return:

```json
{
  "error": "<error.code>",
  "message": "<error.message>",
  "detail": { /* error.to_dict() */ }
}
```

### Error Wrapping at Platform Boundary

The adapter layer (`src/autom8_ads/platforms/meta/adapter.py`, `_map_error` method) translates SDK-specific exceptions into `AdsError` subtypes via explicit `isinstance` dispatch. The original exception is always passed as `original` and re-raised with `from e`.

### Non-Fatal Side-Effect Pattern

Non-fatal side-effect calls (`_try_persist`, `_try_asana_writeback`, `_try_emit_campaign_created`) catch bare `Exception`, log a warning, append to a `warnings: list[str]`, and continue. This is the standard pattern for external service calls that should not block the primary flow.

### Logging Conventions with Errors

All logging uses `autom8y_log.get_logger(__name__)`. Module-level singleton. Log calls always pass structured data via `extra={}`, never via f-string interpolation in the message string. Error messages are snake_case string literals used as the event key (e.g., `"budget_auto_adjusted"`, `"event_handler_error"`).

---

## File Organization

### Package Layout

```
src/autom8_ads/
├── __init__.py          # Version only (__version__ = "0.1.0")
├── app.py               # FastAPI app factory + lifespan + exception handlers
├── config.py            # AdsConfig (pydantic-settings, ADS_ prefix)
├── dependencies.py      # FastAPI DI functions + type aliases
├── errors.py            # Full error hierarchy
│
├── api/                 # HTTP route handlers (one file per tag group)
│   ├── campaigns.py     # Campaign read endpoints
│   ├── health.py        # /health, /ready, /health/deps
│   ├── health_models.py # Health check response models
│   ├── insights.py      # Insights endpoint
│   ├── launch.py        # POST /offers/{id}/launch
│   └── status.py        # Status update endpoint
│
├── clients/             # External service clients (protocol + stub)
│   ├── asana.py         # AsanaServiceProtocol + StubAsanaServiceClient
│   └── data.py          # DataServiceProtocol + StubDataServiceClient
│
├── events/              # Domain event bus
│   ├── bus.py           # EventBus protocol + InProcessEventBus
│   └── subscribers.py   # EventLogSubscriber
│
├── launch/              # Launch pipeline orchestration
│   ├── context.py       # build_launch_intent(), build_platform_extensions()
│   ├── idempotency.py   # LaunchIdempotencyCache
│   └── service.py       # LaunchService (central orchestrator)
│
├── lifecycle/           # Ad object lifecycle management
│   ├── budget.py        # BudgetReconciler + BudgetFixResult
│   ├── campaign_lock.py # CampaignLock (DynamoDB) + NullCampaignLock
│   ├── campaign_matcher.py  # CampaignMatcher protocol + DefaultCampaignMatcher
│   ├── campaign_search.py   # CampaignSearchService
│   ├── factory.py       # AdFactory (strategy dispatch)
│   └── strategies/
│       ├── base.py      # LaunchStrategy protocol
│       └── v2_meta.py   # V2MetaLaunchStrategy (only concrete impl)
│
├── models/              # Domain model definitions (one model-group per file)
│   ├── base.py          # AdsModel (frozen Pydantic BaseModel)
│   ├── enums.py         # All StrEnum definitions
│   ├── events.py        # Domain event models
│   ├── launch.py        # LaunchIntent, LaunchRequest, LaunchResult, LaunchResponse
│   ├── name_encoding.py # NameEncoding[T] generic + field tuples
│   ├── responses.py     # HTTP response wrapper models
│   ├── search.py        # Campaign search types
│   └── (ad, ad_group, budget, campaign, creative, insights, schedule, targeting)
│
├── platforms/           # Platform adapter layer
│   ├── protocol.py      # AdPlatform protocol (runtime_checkable)
│   ├── translator.py    # MetaTranslator
│   ├── types.py         # PlatformAdObject, PlatformAssetRef
│   └── meta/
│       ├── adapter.py   # MetaPlatformAdapter + SDK stub fallbacks
│       ├── constants.py # Meta API subcodes, param names, objective mappings
│       └── params.py    # Pure build_*_params() functions
│
└── routing/             # Account routing
    ├── config.py        # AccountRule, AccountRouterConfig
    └── router.py        # AccountRouter
```

### File Boundary Rules

Each `clients/` file contains: the `Protocol` (runtime_checkable), any request/response models specific to the stub, and the `Stub*` implementation. Protocol and stub are co-located.

`api/` files contain: one `APIRouter` instance, endpoint functions, and no business logic. Business logic goes in `launch/` or `lifecycle/`.

`models/` files group by domain concept, not by I/O direction. All enums live in `models/enums.py`. `models/base.py` contains only `AdsModel`.

`platforms/meta/` separates: adapter (protocol implementation), constants (magic numbers), and params (pure param-builder functions).

`dependencies.py` is the FastAPI DI registry. All `Annotated[T, Depends(...)]` type aliases live here.

---

## Domain-Specific Idioms

### 1. NameEncoding -- Structured Metadata in Ad Name Strings

Ad object names on Meta are used as denormalized storage for structured metadata. `NameEncoding[T]` generic in `src/autom8_ads/models/name_encoding.py` encodes/decodes `NamedTuple` fields to/from bullet-separated strings (U+2022). Module-level encoding instances are singletons.

### 2. Protocol + Stub Pattern

External dependencies use: `@runtime_checkable Protocol`, `Stub*` implementation (logs + returns mock data), and (eventually) real implementation. Appears in: `clients/asana.py`, `clients/data.py`, `events/bus.py`, `lifecycle/campaign_matcher.py`, `lifecycle/campaign_lock.py`.

### 3. Lazy SDK Import Guard

SDK dependencies that may not be installed use `try/except ImportError` with `_HAS_META_SDK` boolean flag. Local stub classes replace real ones when SDK is absent. Pattern in `src/autom8_ads/platforms/meta/adapter.py`.

### 4. `from __future__ import annotations` + `TYPE_CHECKING` Guard

All files use `from __future__ import annotations`. Import-only-for-type references live in `if TYPE_CHECKING:` blocks. Enforced by ruff `TCH` rule set. `pyproject.toml` configures `runtime-evaluated-base-classes` for Pydantic compatibility.

### 5. Structured Logging with `autom8y_log`

Direct use of `loguru`, `structlog`, or `logging` is banned by ruff lint (`flake8-tidy-imports.banned-api`). All logging uses `autom8y_log.get_logger(__name__)`. Log messages are snake_case string literals; context in `extra={}`.

### 6. `warnings: list[str]` Accumulator

Non-fatal side-effect calls accumulate short string codes into a `warnings` list surfaced in `LaunchResult.warnings`. Example strings: `"data_persistence_skipped: ..."`, `"asana_writeback_skipped: ..."`.

### 7. App State as Service Locator

FastAPI's `app.state` is the singleton registry. `lifespan()` initializes singletons; DI functions in `dependencies.py` extract them from `request.app.state`.

### 8. Metric Logging via Structured Log Events

Metrics emitted as structured log events: `logger.info("metric_counter", extra={"metric": "budget_auto_fix_total", "value": 1})`.

### 9. `NullObject` / `Null*` for Optional Infrastructure

Optional infrastructure gets a `Null*` no-op variant (e.g., `NullCampaignLock`).

### 10. Budget Amounts Always in Cents

All monetary amounts in integer cents. The `_cents` suffix is mandatory on all money fields.

---

## Naming Patterns

### Classes

- Domain errors: `Ads{Concept}Error` (e.g., `AdsValidationError`, `AdsBudgetError`)
- Platform adapter: `{Platform}PlatformAdapter` (e.g., `MetaPlatformAdapter`)
- Protocol contracts: bare noun (e.g., `AdPlatform`, `LaunchStrategy`, `EventBus`)
- Stub implementations: `Stub{ServiceName}Client` (e.g., `StubAsanaServiceClient`)
- Null implementations: `Null{ClassName}` (e.g., `NullCampaignLock`)
- Config: `{Service}Config` (e.g., `AdsConfig`, `AccountRouterConfig`)
- Domain models: noun phrases (e.g., `LaunchIntent`, `LaunchResult`, `PlatformExtensions`)
- Strategy implementations: `V{version}{Platform}{Concept}Strategy` (e.g., `V2MetaLaunchStrategy`)
- Enums: `StrEnum` subclasses, PascalCase (e.g., `Platform`, `DomainStatus`)
- NamedTuples: `{Concept}NameFields` (e.g., `CampaignNameFields`)

### Functions

- Builder functions: `build_{thing}(...)` (e.g., `build_campaign_params`)
- Filter functions: `filter_{criterion}(...)` (e.g., `filter_algo_version`)
- Private helpers: underscore prefix (e.g., `_map_error`, `_try_persist`)
- FastAPI DI functions: `get_{thing}(request)` (e.g., `get_config`, `get_event_bus`)
- Parse utilities: `_parse_{type}(value)` (e.g., `_parse_dt`)

### Variables

- Module-level logger: always `logger = get_logger(__name__)`
- Module-level constants: `UPPER_SNAKE_CASE` (e.g., `CAMPAIGN_NAME`, `META_BUDGET_CONFLICT_SUBCODE`)
- Private constants: `_UPPER_SNAKE_CASE` (e.g., `_DEFAULT_WEEKLY_SPEND_CENTS`)
- Type aliases for DI: `{Concept}Dep = Annotated[T, Depends(...)]` (e.g., `LaunchServiceDep`)

### Files

- One class (or one functional grouping) per file, snake_case filenames
- Protocol + implementation in same file for client protocols
- Protocol-only in `base.py` or `protocol.py` for platform/strategy protocols

### Deviation from PEP 8

No observed deviations. Line length is 100 (set in `pyproject.toml`). All formatting enforced by ruff.

---

## Knowledge Gaps

- `lifecycle/campaign_search.py` interface observed indirectly; full implementation not read
- Several model files (`ad.py`, `ad_group.py`, `campaign.py`, `budget.py`, etc.) not read in full but usage patterns are clear
- Test organization and fixture patterns documented in `.know/test-coverage.md`
- `autom8y_log`, `autom8y_config`, `autom8y_http`, `autom8y_telemetry` SDK internals not locally available; only usage patterns visible
