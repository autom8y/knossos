---
domain: conventions
generated_at: "2026-03-27T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "1cfde11"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

> Language: Python 3.12. Framework: FastAPI + Pydantic v2. Package: `autom8_ads` under `src/autom8_ads/`.

---

## Error Handling Style

### Error Hierarchy

All errors inherit from a single base class defined in `src/autom8_ads/errors.py`. The hierarchy is:

```
AdsError (base)
+-- AdsValidationError          -- domain validation failures (HTTP 422)
+-- AdsConfigError              -- configuration/auth errors (HTTP 500)
+-- AdsExternalServiceError     -- external service communication failures (HTTP 503)
|   +-- CampaignLockError       -- DynamoDB lock failures (defined in campaign_lock.py)
+-- AdsPlatformError            -- platform API errors (HTTP 502)
    +-- AdsTransientError       -- retryable errors: rate limit, timeout, 5xx (HTTP 503)
    +-- AdsNotFoundError        -- ad object not found on platform (HTTP 404)
    +-- AdsBudgetError          -- budget constraint violations (HTTP 422)
```

Each error class carries:
- `code: ClassVar[str]` -- string error code (e.g. `"ADS_TRANSIENT_ERROR"`)
- `http_status: ClassVar[int]` -- mapped HTTP status
- `to_dict()` -- serializes to `{code, message, http_status, **context}`

### Error Creation Pattern

Always use domain error classes from `autom8_ads.errors`, never raw `Exception`. Examples:

- **Validation errors**: `raise AdsValidationError(field="daily_budget_cents", reason="Budget must be positive")` -- keyword-only `field` + `reason` args
- **Config errors**: `raise AdsConfigError("ADS_META_ACCOUNT_IDS must be set ...")` -- positional message string
- **Platform errors**: `raise AdsPlatformError(Platform.META, original_exception, ...)` -- platform enum + original exception as positional args
- **Transient errors**: `raise AdsTransientError(Platform.META, original, retry_after=30.0)` -- adds optional `retry_after` kwarg
- **External service errors**: subclass pattern used in `CampaignLockError` -- `super().__init__(service="campaign_lock", reason=reason)`

Bare `ValueError` is used in one place: inside `model_post_init` in `AdsConfig` (config validation), using the `msg = "..."` / `raise ValueError(msg)` split-line style (to satisfy linters).

`RuntimeError` appears once in `src/autom8_ads/cache/tree_cache.py` (line 88) for cold-start fetch failure -- a minor exception to the rule.

### Error Wrapping Convention

Exception chaining via `from e` is used consistently when re-raising:

```python
raise CampaignLockError(reason=f"Failed to create DynamoDB client: {e}") from e
raise HTTPException(...) from e   # in dependencies.py auth boundary
```

### Error Propagation: Domain vs. HTTP Boundary

- **Inside domain layer** (`lifecycle/`, `cleanup/`, `platforms/`): raise `AdsError` subclasses directly.
- **At the HTTP boundary** (`app.py`): two registered exception handlers map domain errors to JSON responses:
  - `AdsTransientError` handler: adds `Retry-After` header when `retry_after` is set
  - `AdsError` handler: converts `{error, message, detail}` JSON shape
- **At the auth dependency boundary** (`dependencies.py`): external SDK errors (`TokenExpiredError`, `CircuitOpenError`, `AuthError`) are caught and re-raised as `HTTPException` with structured `detail` dicts containing `{"error": "SCREAMING_SNAKE", "message": "..."}`.

### Logging at Error Boundaries

Pattern: log with a snake_case event key as the first positional argument, then `extra={}` dict. Level selection:

| Level | When used |
|-------|-----------|
| `logger.info()` | Token expiry, expected circuit events |
| `logger.warning()` | Degraded state, non-fatal failures |
| `logger.error()` | Misconfiguration that is caught and converted |
| `logger.exception()` | Unexpected exceptions in background loops (e.g. `tree_sync.py`) |
| `logger.debug()` | Stub operations, lock release failures (non-fatal) |

### Logger Instantiation

Every module that logs declares:
```python
from autom8y_log import get_logger
logger = get_logger(__name__)
```
This is enforced by a ruff banned-api rule prohibiting direct `loguru` or `structlog` imports. 28 of 60 source files declare a logger (modules that only define models do not).

---

## File Organization

### Source Directory Structure

```
src/autom8_ads/
+-- __init__.py          -- version export
+-- app.py               -- FastAPI application factory, lifespan, exception handlers
+-- config.py            -- AdsConfig (pydantic-settings, ADS_ prefix)
+-- dependencies.py      -- FastAPI DI dependencies: auth, service factories
+-- errors.py            -- Full error hierarchy (single file)
+-- api/                 -- FastAPI routers: one file per resource
|   +-- campaigns.py     -- GET /campaigns routes
|   +-- cleanup.py       -- POST /cleanup routes
|   +-- health.py        -- GET /health
|   +-- health_models.py -- Health response models (separate from router)
|   +-- insights.py      -- GET /insights routes
|   +-- launch.py        -- POST /launch routes
|   +-- status.py        -- GET /status routes
+-- models/              -- Pure Pydantic domain models (no business logic)
|   +-- base.py          -- AdsModel (frozen, extra="ignore", from_attributes)
|   +-- enums.py         -- All StrEnum definitions (single file)
|   +-- ad.py, ad_group.py, campaign.py, creative.py  -- per-object models
|   +-- budget.py, schedule.py, targeting.py          -- sub-models
|   +-- launch.py        -- LaunchIntent, LaunchRequest, LaunchResponse, LaunchResult
|   +-- cleanup.py       -- CleanupIntent, CleanupResult models
|   +-- events.py        -- Domain event types
|   +-- search.py        -- Search filter/result models
|   +-- scoring.py       -- ScoringConfig and scoring result models
|   +-- cache.py         -- CachedTree, CachedObject models
|   +-- responses.py     -- Composite read-model responses
|   +-- insights.py      -- InsightRow, InsightsResponse
|   +-- name_encoding.py -- NameEncoding[T] + CampaignNameFields etc.
+-- platforms/           -- Platform adapter layer
|   +-- protocol.py      -- AdPlatform Protocol (runtime_checkable)
|   +-- types.py         -- PlatformAdObject, ChildrenPage, PlatformAssetRef
|   +-- translator.py    -- MetaTranslator (SDK -> domain model)
|   +-- meta/            -- Meta platform implementation
|       +-- adapter.py   -- MetaPlatformAdapter (implements AdPlatform)
|       +-- params.py    -- build_campaign_params(), build_ad_set_params() etc.
|       +-- constants.py -- META_BUDGET_CONFLICT_SUBCODE, META_DAILY_BUDGET_PARAM etc.
|       +-- stubs.py     -- StubMetaAdsClient for tests/local dev
|       +-- __init__.py  -- Public re-exports
+-- lifecycle/           -- Ad launch lifecycle domain logic
|   +-- strategies/
|   |   +-- base.py      -- LaunchStrategy Protocol
|   |   +-- v2_meta.py   -- V2MetaLaunchStrategy (the only strategy)
|   +-- factory.py       -- AdFactory (composes strategy + platform)
|   +-- budget.py        -- BudgetReconciler
|   +-- campaign_lock.py -- CampaignLock (DynamoDB) + NullCampaignLock
|   +-- campaign_matcher.py -- Campaign reuse matching
|   +-- campaign_search.py  -- CampaignSearchService
|   +-- campaign_budget.py  -- (inferred)
+-- launch/              -- Launch service and support
|   +-- service.py       -- LaunchService (orchestrator)
|   +-- context.py       -- LaunchContext (builds intent from Asana task)
|   +-- idempotency.py   -- LaunchIdempotencyCache (in-memory TTL)
+-- cleanup/             -- Ad cleanup pipeline
|   +-- service.py       -- CleanupService (orchestrator)
|   +-- pipeline.py      -- CleanupPipeline (scoring + action logic)
|   +-- scoring.py       -- AdObjectScorer
|   +-- strategy.py      -- CleanupStrategy Protocol
|   +-- history.py       -- CleanupHistory
|   +-- tree_sync.py     -- TreeSyncService (background drift detection)
+-- guards/              -- Pre-mutation safety checks
|   +-- protocol.py      -- Guard Protocol + GuardResult + GuardChain
|   +-- budget.py        -- BudgetGuard
|   +-- targeting.py     -- TargetingGuard
|   +-- status_transition.py -- StatusTransitionGuard
|   +-- offer_lifecycle.py   -- OfferLifecycleGuard
|   +-- creative_enrichment.py -- CreativeEnrichmentGuard
|   +-- lifecycle_enrichment.py -- LifecycleEnrichmentGuard (21-section mapping)
+-- intelligence/        -- Data-driven optimization intelligence
|   +-- models.py        -- Domain models: policies, patterns, confidence, recommendations
|   +-- vertical_policy.py  -- VerticalPolicyEngine (54->15 canonical verticals, 4-tier confidence)
|   +-- mutation_patterns.py -- MutationPatternAnalyzer (targeting, budget ramp, CPS impact)
|   +-- config_normalizer.py -- Config JSON normalization across 3 schema eras
|   +-- creative_bridge.py   -- CreativePerformanceBridge (swap detection, fatigue, 5-level degradation)
+-- reconciliation/      -- Business record reconciliation
|   +-- models.py        -- Reconciliation domain models
|   +-- discovery.py     -- Candidate discovery
|   +-- resolver.py      -- BusinessResolver
|   +-- monitor.py       -- Pipeline monitoring
|   +-- service.py       -- ReconciliationService orchestrator
+-- clients/             -- External service clients
|   +-- asana.py         -- AsanaServiceProtocol + StubAsanaServiceClient
|   +-- asana_http.py    -- Real HTTP AsanaServiceClient
|   +-- data.py          -- DataCampaignProtocol re-exports + CleanupDataProtocol
|   +-- data_http.py     -- HttpDataServiceClient
|   +-- ad_optimizations.py  -- AdOptimizationsProtocol + Stub + HTTP (single file)
|   +-- asset_data.py    -- DataAssetProtocol + Stub + HTTP (single file)
|   +-- intelligence_data.py -- IntelligenceDataProtocol + Stub + HTTP (single file)
+-- cache/
|   +-- tree_cache.py    -- TreeCache + periodic_tree_refresh()
+-- events/
|   +-- bus.py           -- InProcessEventBus
|   +-- subscribers.py   -- EventLogSubscriber
+-- routing/
    +-- config.py        -- AccountRoutingConfig
    +-- router.py        -- AccountRouter
```

### File Naming Rules

- **One concern per file**: `errors.py` for all errors, `enums.py` for all enums, `base.py` for the base model.
- **Protocol/interface files**: named after the abstract concept (`protocol.py`, `base.py`).
- **Service implementations**: `service.py` in the relevant package.
- **HTTP clients**: `{name}_http.py` for real HTTP; `{name}.py` for protocol + stub.
- **Constants**: `constants.py` within the platform sub-package.
- **Params builders**: `params.py` for pure data-building functions.

### `__init__.py` Usage

`__init__.py` files are present throughout but most are empty (boundary markers). The `platforms/meta/__init__.py` provides a controlled re-export surface. `clients/data.py` uses `__all__` explicitly to declare its public API.

### `from __future__ import annotations`

All source files (118 as of Project Alchemy) include `from __future__ import annotations` as the first line. This is a universal project convention enabling PEP 563 postponed evaluation of annotations throughout.

### `if TYPE_CHECKING:` Blocks

Many files use `if TYPE_CHECKING:` guards. These are used when:
- An import would create a circular dependency
- The type is only needed for annotation (not runtime), especially for Protocol arguments and return types

### Generated Code

No generated code detected. The `models/` directory is hand-authored Pydantic models.

---

## Domain-Specific Idioms

### 1. Protocol + Stub + Real HTTP Triple

A recurring pattern for every external service dependency:

```
{name}.py         -- defines Protocol + StubXxx (no HTTP, deterministic returns)
{name}_http.py    -- defines real HttpXxx (uses autom8y-http SDK)
```

This allows local development and tests to run without external services. The `StubDataCampaignClient` and `StubCleanupDataClient` follow this pattern. `NullCampaignLock` in `campaign_lock.py` is the same pattern applied to the DynamoDB lock.

Additional Protocol + Stub + HTTP triples added during Project Alchemy:

| Protocol | Stub | HTTP | File |
|---|---|---|---|
| `AdOptimizationsProtocol` | `StubAdOptimizationsClient` | `HttpAdOptimizationsClient` | `clients/ad_optimizations.py` |
| `DataAssetProtocol` | `StubDataAssetClient` | `HttpDataAssetClient` | `clients/asset_data.py` |
| `IntelligenceDataProtocol` | `StubIntelligenceDataClient` | `HttpIntelligenceDataClient` | `clients/intelligence_data.py` |

All three follow the identical file structure: Protocol definition + Stub class (deterministic returns) + HTTP class (uses `ResilientCoreClient`) in a single file. This differs from the original pattern (`data.py` + `data_http.py` two-file split) -- the single-file variant was adopted for these clients because their protocol surface is smaller.

### 2. StrEnum for All Domain Enumerations

All enumerations use `StrEnum` (not `IntEnum`, not `Enum`). This means enum values serialize as lowercase strings in JSON automatically. Core enums live in `src/autom8_ads/models/enums.py` (single file). Exception: intelligence-domain enums (`ConfidenceLevel`, `MutationType`, `MutationEra`, `DegradationLevel`, `FatigueStatus`, `CreativeFormat`, `PracticeMaturity`) are defined in `src/autom8_ads/intelligence/models.py` to keep the intelligence package self-contained.

### 3. NameEncoding[T] -- Structured Metadata in Platform Name Strings

A project-specific generic class in `src/autom8_ads/models/name_encoding.py` encodes structured fields into platform ad object name strings using a bullet separator (U+2022). Three concrete schemas: `CampaignNameFields`, `AdGroupNameFields`, `AdNameFields` -- all `NamedTuple` subclasses. Typed singleton instances (`CAMPAIGN_NAME`, `AD_GROUP_NAME`, `AD_NAME`) are used at encode/decode call sites.

This is a legacy pattern that continues in V2 because platform name fields serve as denormalized storage.

### 4. AdsModel as Universal Base

All Pydantic domain models inherit `AdsModel` (not `BaseModel` directly). `AdsModel` enforces:
- `frozen=True` -- immutability
- `extra="ignore"` -- forward compatibility with new platform fields
- `from_attributes=True` -- ORM-mode compatibility

Exception: `LaunchRequest` overrides with `extra="forbid"` (strict API input). `LaunchCacheEntry` uses `BaseModel` directly (not `AdsModel`) -- a minor inconsistency.

### 5. AdPlatform Protocol with `runtime_checkable`

The `AdPlatform` Protocol in `src/autom8_ads/platforms/protocol.py` is `@runtime_checkable`. This allows `isinstance()` checks in tests without importing concrete adapters.

### 6. LaunchStrategy / CleanupStrategy / AdPlatform Protocols

The project uses `Protocol` (not ABC) for all extension points:
- `LaunchStrategy` -- `execute(intent, platform, extensions) -> LaunchResult`
- `CleanupStrategy` -- in `cleanup/strategy.py`
- `AdPlatform` -- 14-method platform contract

Composition via constructor injection (no `__new__` chains, no registry).

### 7. Module-Level Constants for SDK Mapping

Platform-specific constant lookup tables live in `constants.py` files. Example: `DOMAIN_TO_META_STATUS: dict[DomainStatus, str]` in `adapter.py` maps domain status enums to Meta string literals.

### 8. Interop Boundary Pattern

Models from `autom8y_interop` are imported with underscore-prefixed aliases (`_InteropCampaignWriteResponse`) and then subclassed locally to add typed fields. The local subclass name drops the leading underscore and removes the `Interop` prefix. Local instances must call `model_dump()` before crossing the interop boundary (documented in `clients/data.py` module docstring).

### 9. Deferred Error Imports

`AdsConfigError` and related errors are imported inside functions/methods when they would create circular imports at module level. Pattern:

```python
from autom8_ads.errors import AdsConfigError
raise AdsConfigError("...")
```

This appears in `config.py`, `app.py`, and `dependencies.py`.

### 10. `app.state` for Singleton DI

All application-level singletons are attached to `app.state` during lifespan startup. FastAPI `Depends()` factories in `dependencies.py` extract them via `request.app.state`. This is the project's DI container.

### 11. Guard Protocol + GuardChain (Project Alchemy)

Pre-mutation safety checks follow the `Guard` Protocol (`guards/protocol.py`). Each guard implements `check(object_type, context) -> GuardResult`. The `GuardChain` evaluates guards in order and short-circuits on first deny. Guards are safety nets: timeout -> allow with confidence=0.5. `LifecycleEnrichmentGuard` is a post-structural enrichment layer that queries the intelligence infrastructure (vertical policy + mutation patterns) with a 2-second timeout.

### 12. Confidence Tier Model (Project Alchemy)

Intelligence components use a shared 4-tier confidence model (`ConfidenceLevel` StrEnum in `intelligence/models.py`): HIGH, MODERATE, LOW, INSUFFICIENT. Confidence is determined by sample size thresholds (e.g., HIGH: >= 100 businesses AND >= 10,000 leads). All intelligence responses carry confidence metadata.

### 13. Canonical Vertical Normalization (Project Alchemy)

`VerticalPolicyEngine` normalizes 54 raw vertical entries to 15 canonical keys via `CANONICAL_VERTICAL_MAP`. Niche verticals fall back to ecosystem defaults. This pattern ensures consistent vertical identification across all intelligence queries.

---

## Naming Patterns

### Package and Module Names

- `snake_case` throughout, as Python standard
- Package names are concern-oriented nouns: `lifecycle`, `cleanup`, `platforms`, `launch`, `cache`, `events`, `routing`, `clients`
- Sub-packages use platform names: `platforms/meta/`
- `models/` contains only data, never logic

### Class Naming

| Pattern | Example |
|---------|---------|
| Domain models | `Campaign`, `AdGroup`, `Ad`, `Creative` -- plain nouns |
| Domain errors | `AdsError`, `AdsValidationError`, `AdsPlatformError` -- `Ads` prefix + PascalCase |
| Config | `AdsConfig` -- `Ads` prefix |
| Services | `LaunchService`, `CleanupService` -- `Service` suffix |
| Strategies | `V2MetaLaunchStrategy` -- version prefix + platform + `Strategy` suffix |
| Protocols | `AdPlatform`, `LaunchStrategy`, `CleanupStrategy` -- no `I` prefix (not Java-style) |
| Adapters | `MetaPlatformAdapter` -- platform + `Adapter` suffix |
| Factories | `AdFactory` -- domain noun + `Factory` |
| Caches | `TreeCache`, `LaunchIdempotencyCache` -- noun + `Cache` |
| Stubs | `StubAsanaServiceClient`, `StubDataCampaignClient` -- `Stub` prefix |
| Null objects | `NullCampaignLock` -- `Null` prefix |
| Encoders | `NameEncoding[T]` -- noun + generic |
| Intelligence engines | `VerticalPolicyEngine`, `MutationPatternAnalyzer` -- domain noun + `Engine`/`Analyzer` |
| Bridges | `CreativePerformanceBridge` -- cross-domain noun + `Bridge` |
| Guards | `BudgetGuard`, `TargetingGuard`, `LifecycleEnrichmentGuard` -- domain noun + `Guard` |

### Field Naming

- Monetary fields use `_cents` suffix: `daily_budget_cents`, `lifetime_budget_cents`, `weekly_ad_spend_cents`, `minimum_budget_cents`
- Platform IDs: `platform_id` (the vendor-assigned ID), `id` (internal DB ID)
- Raw platform data: `raw` dict field on models that carry unstructured SDK responses
- Timestamps: `created_at`, `updated_at`, `platform_created_at` -- `_at` suffix, `datetime` type
- Boolean flags: `auth_disabled`, `use_stub_data_client`, `controls_budget`, `is_dynamic` -- no `is_` prefix by default, except where it reads naturally

### Logging Event Keys

Structured log event keys are `snake_case` strings as the first positional argument: `"jwt_expired"`, `"meta_client_configured"`, `"tree_drift_detected"`, `"autom8_ads starting"`. The convention is predominantly `snake_case_nouns` but a few use natural English phrases (`"Starting ad launch"`). Structured data goes in `extra={}`.

### Private Members

Private instance attributes: `_field_name` (single underscore). No name mangling (`__`) outside of `__context` in `model_post_init` (Pydantic convention).

### Type Annotation Conventions

- `from __future__ import annotations` universal (all 118 source files) -- all annotation strings are lazy
- Type-only imports under `if TYPE_CHECKING:` (27 files)
- Union types use `X | Y` (Python 3.10+ syntax, enforced by `target-version = "py312"`)
- `list[str]` / `dict[str, Any]` (lowercase generics, not `List[str]`)
- `str | None` preferred over `Optional[str]`

---

## Knowledge Gaps

1. `routing/config.py` contents -- The `AccountRoutingConfig` model and routing rule structure were not directly inspected; the routing logic was observed only through `dependencies.py` callsites.
2. `cleanup/strategy.py` and `cleanup/history.py` -- The `CleanupStrategy` Protocol body and `CleanupHistory` class were not read; their role is inferred from the pattern.
3. `platforms/translator.py` -- `MetaTranslator` class was referenced but not read.
4. `clients/asana_http.py` full content -- only the logger declaration was observed.
5. Test conventions -- intentionally out of scope per criteria note (see `test-coverage` domain).
