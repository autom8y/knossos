---
domain: architecture
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

# Codebase Architecture

## Package Structure

The `autom8_ads` service is a Python 3.11+ FastAPI application packaged under `src/autom8_ads/`. There are 8 top-level packages and 33 source files.

**Package inventory** (all paths relative to `src/autom8_ads/`):

| Package | Files | Purpose | Classification |
|---------|-------|---------|----------------|
| `api/` | `health.py`, `launch.py` | HTTP route handlers (FastAPI routers) | Surface layer |
| `clients/` | `data.py` | External service clients | Infrastructure leaf |
| `config.py` | (module file) | `AdsConfig` Pydantic-settings model, reads env vars | Leaf |
| `dependencies.py` | (module file) | FastAPI `Depends()` accessors that extract singletons from `app.state` | Glue layer |
| `errors.py` | (module file) | Exception hierarchy with HTTP status codes | Shared leaf |
| `launch/` | `idempotency.py`, `mapper.py`, `service.py` | Core orchestration: idempotency, payload mapping, the 6-step pipeline | Hub (imports most sibling packages) |
| `lifecycle/` | `factory.py`, `strategies/base.py`, `strategies/v2_meta.py` | Ad creation strategies and the strategy-pattern factory | Domain logic leaf |
| `models/` | `base.py`, `enums.py`, `launch.py`, `offer.py`, `targeting.py` | Pydantic domain models and enums | Shared leaf (imported by nearly everything) |
| `platforms/` | `protocol.py` | `AdPlatform` and `DataServiceProtocol` Protocols — contracts without implementations | Contract leaf |
| `routing/` | `config.py`, `router.py` | Account routing rules and the router that evaluates them | Domain logic leaf |
| `urls/` | `meta.py` | Meta Ads Manager deep-link URL builder | Infrastructure leaf |
| `app.py` | (module file) | Application factory (`create_app`) and `lifespan` context manager | Root hub |

**Hub/leaf classification**:
- **Hub packages** (import many siblings): `launch/service.py` imports `config`, `errors`, `launch/idempotency`, `launch/mapper`, `lifecycle/factory`, `lifecycle/strategies/v2_meta`, `models/enums`, `models/launch`, `models/offer`, `platforms/protocol`, `routing/router`, `urls/meta`. `app.py` similarly imports most packages for wiring. These two are the hubs.
- **Leaf packages** (imported by others, import little): `models/`, `errors.py`, `config.py`, `platforms/protocol.py`, `routing/config.py`, `urls/meta.py`, `lifecycle/strategies/v2_meta.py`.

**Source files**: `src/autom8_ads/` — `services/ads/src/autom8_ads/`

---

## Layer Boundaries

The import graph forms a clear 3-layer model with no observed circular dependencies:

```
HTTP Surface (api/)
       |
       v
Orchestration (launch/service.py, app.py)
       |
       v
Domain Logic (lifecycle/, routing/, urls/, launch/idempotency, launch/mapper)
       |
       v
Shared Contracts (models/, platforms/protocol.py, errors.py, config.py)
```

**Detailed import relationships**:

- `api/launch.py` imports: `dependencies`, `errors`, `launch/idempotency`, `launch/service`, `models/enums`, `models/offer` — it stays in the surface layer and delegates all business logic downward.
- `api/health.py` imports: only `fastapi`. Pure leaf.
- `dependencies.py` imports: `launch/idempotency`, `launch/service` — thin glue between FastAPI and singletons.
- `launch/service.py` (the central hub) imports: `config`, all `errors`, `launch/idempotency`, `launch/mapper`, `lifecycle/factory`, `lifecycle/strategies/v2_meta`, `models/enums`, `models/launch`, `models/offer`, `platforms/protocol`, `routing/router`, `urls/meta`.
- `launch/mapper.py` imports: `models/enums`, `models/launch`, `models/offer`, `models/targeting` — pure transformation.
- `lifecycle/factory.py` imports: `lifecycle/strategies/base`, `models/launch`, `platforms/protocol` — delegates to strategy.
- `lifecycle/strategies/v2_meta.py` imports: `models/launch`, `platforms/protocol` — calls platform adapter methods.
- `lifecycle/strategies/base.py` imports: `models/launch`, `platforms/protocol` — protocol definition only.
- `routing/router.py` imports: `errors`, `models/enums`, `routing/config`.
- `urls/meta.py` imports: stdlib only (`logging`, `datetime`, `urllib.parse`). Pure function leaf.
- `models/*` import only each other within models (`launch.py` imports `base`, `enums`, `targeting`; `offer.py` imports `launch`). `models/base.py` imports only `pydantic`. `models/enums.py` imports only stdlib.
- `platforms/protocol.py` imports: `models/launch`. Contract only.
- `clients/data.py` imports: stdlib only. The stub implementation.

**Boundary-enforcement patterns**: Python `Protocol` (structural subtyping via `typing.Protocol`) is used in `platforms/protocol.py` to define `AdPlatform` and `DataServiceProtocol`. This decouples the launch pipeline from any concrete platform SDK. `AdPlatform` is `@runtime_checkable`, allowing `isinstance` checks on injected adapters.

No circular dependencies observed. The `models/` package is a pure leaf — every other package imports from it, but it imports from nothing internal except itself.

---

## Entry Points and API Surface

**Service entry point**: The application is a FastAPI ASGI service. Entry is `src/autom8_ads/app.py` via `create_app()`. No `__main__.py` exists in the source tree; the service is invoked via `uvicorn autom8_ads.app:create_app --factory` (standard FastAPI pattern).

**Application startup** (traced from `create_app`):
1. `create_app()` constructs a `FastAPI` instance with `lifespan=lifespan`.
2. On startup, the `lifespan` async context manager in `app.py` fires:
   - Instantiates `AdsConfig()` (reads env vars).
   - Creates `OfferPayloadMapper`, `MetaUrlBuilder`, `LaunchIdempotencyCache`, `AccountRouter`, `StubDataServiceClient`.
   - If `platform_adapter` was pre-injected via `app.state`, creates `LaunchService` (otherwise deferred — test injection path).
   - Stores all singletons on `app.state`.
3. `create_app()` includes two routers: `health_router` and `launch_router`.

**HTTP API surface** (2 routers, 3 endpoints):

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| `GET` | `/health` | `api/health.py:health` | Health check; returns `{"status": "ok", "service": "autom8_ads"}` |
| `POST` | `/api/v1/launches` | `api/launch.py:launch_ads` | Launch ads for a pre-resolved `OfferPayload`; returns `LaunchResponse` |
| `DELETE` | `/api/v1/launches/{offer_id}/{platform}` | `api/launch.py:clear_launch_cache` | Clear idempotency cache entry for retry |

**Dependency injection**: `dependencies.py` exposes `get_launch_service(request)` and `get_idempotency_cache(request)` as FastAPI `Depends()` callables. They cast from `app.state` using `typing.cast`.

**Key exported interfaces** (Protocols consumed by other packages):
- `AdPlatform` (`platforms/protocol.py`): 4 async methods (`create_campaign`, `create_ad_set`, `create_creative`, `create_ad`). No concrete implementation ships in this service; the adapter is injected externally (or via test mocks).
- `DataServiceProtocol` (`platforms/protocol.py`): 4 async methods for write-back to `autom8_data`. Currently satisfied by `StubDataServiceClient` in `clients/data.py`.
- `LaunchStrategy` (`lifecycle/strategies/base.py`): Protocol with a single `execute(platform, ctx) -> LaunchResult` method.

---

## Key Abstractions

The 10 most central types in the codebase, ranked by usage frequency and centrality:

**1. `LaunchContext`** (`src/autom8_ads/models/launch.py`)
The internal domain model for an ad launch. Created by `OfferPayloadMapper.to_launch_context()` from an inbound `OfferPayload`. Contains all fields needed by the platform adapter: `offer_id`, `platform`, `account_id`, `objective`, `daily_budget_cents`, `targeting: TargetingSpec`, `page_id`, `asset_ids`, `business_name`, `vertical_key`, `task_gid`. Inherits from `AdsModel` (frozen, immutable). Passed into `AdFactory.launch()` and all platform adapter methods. Notable: explicitly has no `vertical` field (stakeholder decision; only `vertical_key` for name encoding).

**2. `OfferPayload`** (`src/autom8_ads/models/offer.py`)
The inbound API request model. All field inheritance (from Asana parent tasks) is resolved by the caller (`autom8_asana`) before sending. Has strict validators: E.164 phone, platform must be `"meta"`, `algo_version` must be 2, `asset_ids` non-empty. Uses `extra="forbid"` (strict schema). Maps to `LaunchContext` via `OfferPayloadMapper`.

**3. `LaunchResponse`** (`src/autom8_ads/models/offer.py`)
The outbound API response model returned to `autom8_asana` for Asana write-back. Flat structure combining `LaunchResult` fields plus two pre-built URLs (`ad_account_url`, `live_ads_url`). Constructed via `LaunchResponse.from_launch_result(result, ...)` factory classmethod. Also stored in the idempotency cache.

**4. `LaunchService`** (`src/autom8_ads/launch/service.py`)
The central orchestrator. 6-step pipeline: idempotency check -> account routing -> payload mapping -> ad creation via `AdFactory` -> data persistence -> URL construction and response assembly. Accepts all dependencies via constructor injection. The only component that sees the full picture.

**5. `AdPlatform`** (`src/autom8_ads/platforms/protocol.py`)
The `@runtime_checkable` Protocol that decouples the launch pipeline from Meta's SDK. Four methods map directly to the V2 Meta launch sequence. No concrete implementation ships in this repo — the adapter comes from outside (likely `autom8y-auth` or a platform-specific package).

**6. `AdsModel`** (`src/autom8_ads/models/base.py`)
Base Pydantic model for all internal domain objects. Configured `frozen=True` (immutable instances), `extra="ignore"` (forward compatibility with platform responses), `from_attributes=True`. `LaunchContext`, `LaunchResult`, `TargetingSpec`, and routing models all inherit from it.

**7. `LaunchIdempotencyCache`** (`src/autom8_ads/launch/idempotency.py`)
In-memory TTL cache keyed by `"{offer_id}:{platform}"`. Split TTL design (ADR-ADS-007): 5-minute TTL for `in_progress` entries, 24-hour TTL for `completed`/`failed` entries. Expiry is lazy (evict-on-read). Single-instance design (no distributed locking). Stores `LaunchResponse` objects (not `LaunchResult`) so cached responses include pre-built URLs.

**8. `V2MetaLaunchStrategy`** (`src/autom8_ads/lifecycle/strategies/v2_meta.py`)
The only launch strategy (ADR-ADS-002). Implements the 5-step Meta ad creation sequence: Campaign -> AdSet -> Creative -> Ad -> Result. On partial failure, returns a `LaunchResult` with partial IDs and `status="partial"` rather than raising. Identifies failure step via `_failure_step()` for logging.

**9. `AccountRouter`** (`src/autom8_ads/routing/router.py`)
Evaluates `AccountRule` list in order, matching on `platform` and `weekly_spend_cents` range. Falls back to the `is_default` account for the platform. Raises `AdsConfigError` if no match and no default. Currently seeded with a single hardcoded production rule in `app.py` (`_create_default_router_config()`).

**10. `AdsError` hierarchy** (`src/autom8_ads/errors.py`)
All exceptions carry `code: ClassVar[str]` and `http_status: ClassVar[int]`. Subtypes: `AdsValidationError` (422), `AdsPlatformError` (502), `AdsTransientError` (503), `AdsBudgetError` (422), `AdsConfigError` (500), `LaunchInProgressError` (409). `to_dict()` serializes for API error responses.

**Design patterns observed**:
- **Strategy pattern**: `LaunchStrategy` Protocol + `AdFactory` factory + `V2MetaLaunchStrategy` concrete strategy. The factory is closed for extension but the strategy is swappable.
- **Protocol-based dependency injection**: `AdPlatform` and `DataServiceProtocol` are Python `Protocol` types; concrete adapters are injected externally, making the service testable without a real Meta SDK.
- **Singleton-via-app.state**: All service-layer singletons are created once in `lifespan()` and stored on `app.state`. Route handlers retrieve them through `dependencies.py` callables.
- **Frozen immutable models**: `AdsModel` and most Pydantic models are `frozen=True`. Mutation is done via `model_copy(update={...})` (e.g., injecting the routed `account_id` into `LaunchContext`).
- **Stub-first data client**: `StubDataServiceClient` satisfies `DataServiceProtocol` and logs writes without persisting. Real implementation is planned for "Move 4" (per source comments).
- **Graceful degradation on data writes**: Step 5 of `LaunchService` only calls `_try_persist()` when `data_writes_enabled=True` (env var, default `False`). Failures in `_try_persist` are swallowed and logged as warnings.

---

## Data Flow

**Primary pipeline: HTTP request -> ad launch -> response**

```
autom8_asana HTTP POST /api/v1/launches
    |
    v
api/launch.py: launch_ads()
    |  validates body as OfferPayload (Pydantic, field_validators)
    |  extracts trace_id from x-trace-id header
    v
launch/service.py: LaunchService.launch()
    |
    +- Step 1: LaunchIdempotencyCache.get_or_set_in_progress(offer_id, platform)
    |     key = "{offer_id}:{platform}"
    |     MISS -> marks in_progress, continues
    |     HIT (in_progress) -> raises LaunchInProgressError -> 409
    |     HIT (completed/failed) -> returns cached LaunchResponse
    |
    +- Step 2: AccountRouter.route(platform, weekly_spend_cents)
    |     evaluates AccountRule list in-order
    |     falls back to is_default rule for platform
    |
    +- Step 3: OfferPayloadMapper.to_launch_context(payload)
    |     computes daily_budget_cents = weekly_ad_spend_cents // 7 if not provided
    |     wraps payload.targeting dict as TargetingSpec(raw=...)
    |     returns LaunchContext (frozen, immutable)
    |     injects routed account_id via model_copy()
    |
    +- Step 4: AdFactory.launch(ctx) -> V2MetaLaunchStrategy.execute(platform, ctx)
    |     platform.create_campaign(ctx) -> campaign_id (str)
    |     platform.create_ad_set(ctx, campaign_id) -> ad_group_id (str)
    |     platform.create_creative(ctx) -> creative_id (str)
    |     platform.create_ad(ctx, ad_group_id, creative_id) -> ad_id (str)
    |     returns LaunchResult
    |     on exception: returns partial LaunchResult (status="partial"/"failed")
    |
    +- Step 5: _try_persist() (only if data_writes_enabled=True and result.success)
    |     StubDataServiceClient.record_campaign({...}) -- currently logs only
    |     exceptions swallowed (non-blocking)
    |
    +- Step 6: _build_response(result, account_id, payload)
          if result.campaign_id: MetaUrlBuilder.ad_account_url(...)
          if result.ad_id: MetaUrlBuilder.live_ads_url(...)
          LaunchResponse.from_launch_result(result, ad_account_url, live_ads_url)
          LaunchIdempotencyCache.complete(offer_id, platform, response)
          returns LaunchResponse
    |
    v
HTTP 200 LaunchResponse JSON -> autom8_asana for Asana write-back
```

**Secondary pipeline: cache clear -> retry enablement**

```
DELETE /api/v1/launches/{offer_id}/{platform}
    |
    v
api/launch.py: clear_launch_cache()
    |  Platform(platform) validation -- 422 on unknown platform
    v
LaunchIdempotencyCache.clear(offer_id, Platform)
    |  pops key from _cache dict
    v
HTTP 204 No Content
```

**Configuration flow** (environment variables -> service config):

```
Environment variables (no prefix, fleet standard)
    -> AdsConfig() [pydantic-settings reads on lifespan startup]
    |  AUTOM8Y_DATA_URL -> autom8y_data_url
    |  LAUNCH_TIMEOUT_SECONDS -> launch_timeout_seconds
    |  IDEMPOTENCY_COMPLETED_TTL_SECONDS -> idempotency_completed_ttl_seconds
    |  IDEMPOTENCY_IN_PROGRESS_TTL_SECONDS -> idempotency_in_progress_ttl_seconds
    |  DATA_WRITES_ENABLED -> data_writes_enabled (default False)
    |  META_BUSINESS_ID -> meta_business_id (default "596194394092925")
    v
app.py lifespan() -- distributes config values to:
    LaunchIdempotencyCache(completed_ttl_seconds, in_progress_ttl_seconds)
    MetaUrlBuilder(MetaPlatformConfig(business_id=config.meta_business_id))
    LaunchService(config=config, ...)
```

---

## Knowledge Gaps

- **No concrete `AdPlatform` implementation in this service**: The `AdPlatform` Protocol is defined but not implemented here. The Meta SDK adapter lives in an external package. Tests inject mock adapters.
- **`account_routing_config` env var is present in `AdsConfig` but unused**: `AdsConfig.account_routing_config` (str, default "") is declared but `app.py` ignores it entirely, using `_create_default_router_config()` with a hardcoded production account ID instead. The intended file-based routing config loading is not implemented.
- **No `__main__.py`**: The service has no explicit `python -m autom8_ads` entry point. Invocation is via `uvicorn` directly.
- **`StubDataServiceClient` is the only `DataServiceProtocol` implementation**: There is no real HTTP client to `autom8_data`. Move 4 work is deferred.
- **Tests not read**: Test files were enumerated but not read. The `tests/conftest.py` fixture patterns and mock adapter setup are not documented here.
