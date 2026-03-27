---
domain: architecture
generated_at: "2026-03-01T12:42:56Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "762ed0e"
confidence: 0.95
format_version: "1.0"
---

# Codebase Architecture

**Language**: Python 3.11+
**Project**: `autom8y-ads` (package `autom8_ads`)
**Description**: Ad lifecycle management service for the autom8y platform. Manages ad creation, status updates, and performance insights via a FastAPI HTTP service.
**Build system**: Hatchling, dependency management via `uv`
**Key dependencies**: FastAPI, Pydantic v2, pydantic-settings, uvicorn, autom8y-{config,http,log,meta,telemetry} SDK suite

---

## Package Structure

The source root is `src/autom8_ads/`. 8 sub-packages, 1 entry-level module `app.py` and 3 top-level modules.

| Package / Module | File count | Purpose |
|---|---|---|
| `src/autom8_ads/` (root) | 5 files | App factory, config, DI, errors, version |
| `src/autom8_ads/api/` | 6 files | FastAPI routers (HTTP endpoints) |
| `src/autom8_ads/clients/` | 3 files | External service client protocols + stubs |
| `src/autom8_ads/events/` | 3 files | In-process event bus + domain event subscribers |
| `src/autom8_ads/launch/` | 4 files | Launch orchestration (service, context builder, idempotency) |
| `src/autom8_ads/lifecycle/` | 7 files | Campaign lifecycle logic: factory, strategies, budget, matcher, search, lock |
| `src/autom8_ads/models/` | 13 files | Pydantic domain models and enums |
| `src/autom8_ads/platforms/` | 6 files | Platform protocol, types, translator, Meta adapter + params |
| `src/autom8_ads/routing/` | 3 files | Account routing config and router |

**Hub packages** (imported by many):
- `autom8_ads.models` -- imported by virtually every other package; defines shared types
- `autom8_ads.errors` -- imported by `lifecycle`, `platforms`, `launch`, `routing`
- `autom8_ads.platforms.protocol` -- imported by `lifecycle` and `launch`

**Leaf packages** (import few siblings):
- `autom8_ads.routing` -- depends only on `models.enums` and `errors`
- `autom8_ads.events` -- depends only on `models.events` and `autom8y_log`
- `autom8_ads.models` sub-modules -- mostly depend on `models.base` and `models.enums`

**Detailed package contents:**

`src/autom8_ads/__init__.py` -- exposes `__version__`

`src/autom8_ads/app.py` -- `create_app() -> FastAPI`, `lifespan()` context manager, `RequestIDMiddleware`, `_register_exception_handlers()`

`src/autom8_ads/config.py` -- `AdsConfig(Autom8yBaseSettings)` with `ADS_` env prefix; fields: service_name, autom8y_env, asana/data service URLs, launch_timeout_seconds, idempotency_ttl_seconds, account_routing_config, meta credentials, auth_disabled

`src/autom8_ads/dependencies.py` -- FastAPI DI functions: `get_config`, `get_translator`, `get_event_bus`, `get_platform_adapter`, `get_launch_service`, `verify_jwt`, `_build_default_account_router`; type alias exports: `LaunchServiceDep`, `AuthDep`, `TranslatorDep`

`src/autom8_ads/errors.py` -- Error hierarchy: `AdsError` (base, 500) -> `AdsValidationError` (422), `AdsPlatformError` (502) -> `AdsTransientError` (503), `AdsNotFoundError` (404), `AdsBudgetError` (422); plus `AdsConfigError` (500), `AdsExternalServiceError` (503)

`src/autom8_ads/models/` -- 13 files:
- `base.py`: `AdsModel(BaseModel)` -- frozen, extra="ignore", from_attributes=True
- `enums.py`: `Platform`, `AdObjectType`, `DomainStatus`, `CampaignObjective`, `CreativeType`, `AssetType`, `LaunchTrigger` (all `StrEnum`)
- `launch.py`: `LaunchIntent`, `PlatformExtensions`, `LaunchResult`, `LaunchRequest`, `LaunchResponse`
- `campaign.py`: `Campaign`, `AdHierarchy`
- `ad.py`, `ad_group.py`, `creative.py`: individual ad object models
- `budget.py`: `Budget`, `BudgetLevel`, `BudgetPeriod`
- `schedule.py`: `Schedule`
- `targeting.py`: `TargetingSpec`
- `events.py`: `DomainEvent`, `CampaignCreated`, `AdGroupCreated`, `AdCreated`, `StatusChanged`, `BudgetAdjusted`
- `insights.py`: `InsightRow`, `InsightsResponse`
- `name_encoding.py`: `NameEncoding[T]`, `CampaignNameFields`, `AdGroupNameFields`, `AdNameFields`, plus singletons `CAMPAIGN_NAME`, `AD_GROUP_NAME`, `AD_NAME`
- `responses.py`: `CampaignSummary`, `PaginatedResponse`, `StatusUpdateRequest`, `StatusUpdateResponse`
- `search.py`: `CampaignSearchFilter`, `CampaignSearchResult`, `CampaignWriteRequest`

`src/autom8_ads/platforms/` -- 6 files:
- `protocol.py`: `AdPlatform(Protocol)` -- 13 methods (get/create/update/delete object, get_children, search_objects, get_insights, send_conversion, upload_asset, create_creative)
- `types.py`: `PlatformAdObject`, `PlatformAssetRef`
- `translator.py`: `MetaTranslator` -- `to_campaign()`, `to_ad_group()`, `to_ad()`
- `meta/__init__.py`: re-exports from `meta/adapter.py`
- `meta/adapter.py`: `MetaPlatformAdapter` (singleton, implements AdPlatform), plus SDK stub classes when `autom8y_meta` not installed
- `meta/params.py`: pure functions `build_campaign_params`, `build_ad_set_params`, `build_ad_params`, `build_creative_spec`
- `meta/constants.py`: `META_OBJECTIVE_MAP`, `META_DAILY_BUDGET_PARAM`, `META_BUDGET_CONFLICT_SUBCODE`

`src/autom8_ads/lifecycle/` -- 7 files:
- `factory.py`: `AdFactory` -- wires platform + strategy, calls `strategy.execute()`
- `strategies/base.py`: `LaunchStrategy(Protocol)` -- `execute(intent, platform, extensions) -> LaunchResult`
- `strategies/v2_meta.py`: `V2MetaLaunchStrategy` -- 5-step pipeline (validate -> search -> campaign -> adset -> ad+creative)
- `budget.py`: `BudgetReconciler`, `BudgetFixResult`
- `campaign_matcher.py`: `CampaignMatcher(Protocol)`, `DefaultCampaignMatcher`, 6 composable `CampaignFilter` functions
- `campaign_search.py`: `CampaignSearchService` -- calls data client + platform + matcher
- `campaign_lock.py`: campaign-level locking (WIP)

`src/autom8_ads/launch/` -- 4 files:
- `service.py`: `LaunchService` -- 7-step orchestrator (idempotency -> routing -> context -> factory -> persist -> asana -> event bus)
- `context.py`: `build_launch_intent()`, `build_platform_extensions()`
- `idempotency.py`: `LaunchIdempotencyCache`, `LaunchCacheEntry` -- in-memory TTL dict keyed by `offer_id:platform`

`src/autom8_ads/clients/` -- 3 files:
- `data.py`: `DataServiceProtocol` + `StubDataServiceClient` -- 7-method protocol (record_campaign, record_ad_set, record_ad, record_creative, get_campaign_hierarchy, search_campaigns, import_campaign)
- `asana.py`: `AsanaServiceProtocol` + `StubAsanaServiceClient` -- 2-method protocol (get_task, update_custom_fields)

`src/autom8_ads/events/` -- 3 files:
- `bus.py`: `EventBus(Protocol)`, `InProcessEventBus` -- asyncio.create_task fire-and-forget dispatch
- `subscribers.py`: `EventLogSubscriber` -- logs all domain events

`src/autom8_ads/routing/` -- 3 files:
- `config.py`: `AccountRule`, `AccountRouterConfig`
- `router.py`: `AccountRouter` -- rule-based routing by platform + weekly_spend_cents

`src/autom8_ads/api/` -- 6 files:
- `health.py`: `/health`, `/ready`, `/health/deps` endpoints
- `health_models.py`: `HealthStatus`, `CheckResult`, response builder functions
- `launch.py`: `POST /api/v1/offers/{offer_id}/launch`
- `campaigns.py`: `GET /api/v1/accounts/{account_id}/campaigns`, `GET /api/v1/campaigns/{campaign_id}`, `GET /api/v1/campaigns/{campaign_id}/hierarchy`
- `insights.py`: `GET /api/v1/campaigns/{campaign_id}/insights`
- `status.py`: `PATCH /api/v1/campaigns/{campaign_id}/status`

---

## Layer Boundaries

The codebase has a clear 4-layer architecture with strict import direction (outer -> inner):

```
API Layer (api/)
    |
    v
Orchestration Layer (launch/, dependencies.py)
    |
    v
Domain Layer (lifecycle/, routing/, events/, models/)
    |
    v
Infrastructure Layer (platforms/, clients/)
```

**Layer 1 -- API Layer** (`src/autom8_ads/api/`):
- FastAPI routers that receive HTTP requests and return HTTP responses
- Depends on: `dependencies.py` for DI, `models.launch` / `models.responses` / `models.insights` for request/response types
- Does NOT import from `lifecycle/` or `platforms/` directly; accesses them via injected dependencies
- Exception: `api/status.py` and `api/campaigns.py` call `adapter.get_object()` directly (thin platform access for read operations)

**Layer 2 -- Orchestration Layer** (`launch/`, `dependencies.py`, `app.py`):
- `LaunchService` is the primary orchestrator, composing: idempotency, routing, context building, factory execution, data persistence, Asana writeback, event emission
- `dependencies.py` is the DI wiring layer: assembles `LaunchService`, `AdFactory`, `V2MetaLaunchStrategy`, `AccountRouter` per-request from app.state singletons
- `app.py` creates all singletons in `lifespan()` and stores them on `app.state`

**Layer 3 -- Domain Layer**:
- `lifecycle/`: Campaign creation logic, strategy pattern, budget handling, campaign reuse matching
- `routing/`: Account selection rules
- `events/`: In-process pub-sub for domain events
- `models/`: Pydantic data types shared across all layers

**Layer 4 -- Infrastructure Layer**:
- `platforms/`: `AdPlatform` protocol + `MetaPlatformAdapter` (wraps `autom8y_meta` SDK). This is the only code that talks to the Meta Ads API.
- `clients/`: Protocol stubs for `autom8_data` and `autom8_asana` services (currently stub-only)

**Import direction rules observed:**
- `api/` imports `launch/`, `models/`, `dependencies.py`, `platforms/translator.py`
- `launch/` imports `lifecycle/`, `models/`, `clients/`, `platforms/protocol.py`, `routing/`, `events/`
- `lifecycle/` imports `models/`, `errors/`, `platforms/protocol.py`, `platforms/meta/constants.py`
- `platforms/` imports `models/`, `errors/`
- `routing/` imports `models/enums`, `errors`
- `events/` imports `models/events`, `autom8y_log`
- `models/` imports only from `models/base`, `models/enums`, standard library

**Protocol boundaries** enforce decoupling:
- `AdPlatform(Protocol)` at `src/autom8_ads/platforms/protocol.py`
- `LaunchStrategy(Protocol)` at `src/autom8_ads/lifecycle/strategies/base.py`
- `DataServiceProtocol` and `AsanaServiceProtocol` in `clients/`
- `EventBus(Protocol)` in `events/bus.py`

---

## Entry Points and API Surface

### Application Entry Point

The FastAPI app is created by `create_app()` in `src/autom8_ads/app.py`. Served via uvicorn pointing to `autom8_ads.app:create_app()`.

**Startup sequence** (`lifespan()` in `app.py`):
1. `AdsConfig()` instantiated from environment
2. `MetaAdsClient` (real or stub) created if `ADS_META_ACCESS_TOKEN` set
3. `MetaPlatformAdapter(meta_client)` stored on `app.state.platform_adapter`
4. `AccountRouter` built from config stored on `app.state.account_router`
5. `LaunchIdempotencyCache` instantiated on `app.state.idempotency_cache`
6. `StubDataServiceClient` and `StubAsanaServiceClient` on `app.state`
7. `InProcessEventBus` wired with `EventLogSubscriber` for 5 event types on `app.state.event_bus`

### HTTP API Surface

**Health endpoints** (no auth):
- `GET /health` -- liveness probe
- `GET /ready` -- readiness probe
- `GET /health/deps` -- dependency probe

**Launch endpoint** (auth required):
- `POST /api/v1/offers/{offer_id}/launch` -> `LaunchRequest` body -> `LaunchResponse`

**Campaign read endpoints** (auth required):
- `GET /api/v1/accounts/{account_id}/campaigns` -- paginated campaign list
- `GET /api/v1/campaigns/{campaign_id}` -- single campaign
- `GET /api/v1/campaigns/{campaign_id}/hierarchy` -- campaign + ad_groups + ads

**Insights endpoint** (auth required):
- `GET /api/v1/campaigns/{campaign_id}/insights?date_start=&date_stop=&fields=`

**Status update endpoint** (auth required):
- `PATCH /api/v1/campaigns/{campaign_id}/status` -> `StatusUpdateRequest` -> `StatusUpdateResponse`
  - Emits `StatusChanged` domain event on success

### Auth

`verify_jwt` in `dependencies.py` -- structural JWT validation only (3 dot-separated parts). Auth can be disabled via `ADS_AUTH_DISABLED=true` (LOCAL environment only, guarded by `model_post_init`).

### Middleware

`RequestIDMiddleware`: reads/generates `X-Request-ID` header.
`autom8y_telemetry.instrument_app()`: `MetricsMiddleware` + `/metrics` endpoint + tracing (degrades gracefully if SDK absent).

---

## Key Abstractions

### 1. `AdsModel` -- Base Pydantic Model (`src/autom8_ads/models/base.py`)
All domain models inherit. `frozen=True`, `extra="ignore"`, `from_attributes=True`.

### 2. `AdPlatform(Protocol)` -- Platform Abstraction (`src/autom8_ads/platforms/protocol.py`)
13-method `runtime_checkable` protocol. Single seam between business logic and platform SDKs.

### 3. `LaunchStrategy(Protocol)` (`src/autom8_ads/lifecycle/strategies/base.py`)
One implementation: `V2MetaLaunchStrategy`. 5-step pipeline.

### 4. `LaunchIntent` + `PlatformExtensions` (`src/autom8_ads/models/launch.py`)
Command objects feeding the entire launch pipeline.

### 5. `PlatformAdObject` (`src/autom8_ads/platforms/types.py`)
Universal ad object wrapper with `raw` dict for full SDK response.

### 6. `NameEncoding[T]` (`src/autom8_ads/models/name_encoding.py`)
Bidirectional name serialization using bullet (U+2022) separators. Legacy denormalized storage pattern.

### 7. `EventBus(Protocol)` / `InProcessEventBus` (`src/autom8_ads/events/bus.py`)
Fire-and-forget domain events via `asyncio.create_task()`. 5 event types.

### 8. `BudgetReconciler` (`src/autom8_ads/lifecycle/budget.py`)
Budget error recovery for Meta subcodes 1885621 (conflict) and budget-too-low.

### 9. Error Hierarchy (`src/autom8_ads/errors.py`)
`AdsError` base with typed subtypes mapped to HTTP status codes.

### 10. `AccountRouter` (`src/autom8_ads/routing/router.py`)
Rule-based account routing by platform + weekly_spend_cents range.

---

## Data Flow

### Primary Flow: Ad Launch

```
HTTP POST /api/v1/offers/{offer_id}/launch
  -> LaunchService.launch()
    Step 1: Idempotency check (offer_id:platform key)
    Step 2: Account routing (platform + spend -> account_id)
    Step 3: Build LaunchIntent + PlatformExtensions
    Step 4: AdFactory -> V2MetaLaunchStrategy.execute()
      4.1: Validate intent + extensions
      4.2: (optional) Campaign search/reuse
      4.3: Create campaign (CAMPAIGN_NAME.encode -> build_campaign_params -> platform.create_object)
      4.4: Create ad set (AD_GROUP_NAME.encode -> build_ad_set_params -> platform.create_object)
      4.5: Create creative + ad (build_creative_spec -> platform.create_creative, build_ad_params -> platform.create_object)
    Step 5: Data persistence [non-fatal stub]
    Step 6: Asana writeback [non-fatal stub]
    Step 6b: CampaignCreated event [fire-and-forget]
    Step 7: Cache completed result
  -> LaunchResponse
```

### Campaign Read Flow

```
GET /campaigns/{id} -> adapter.get_object() -> MetaTranslator.to_campaign() -> Campaign JSON
```

### Status Update Flow

```
PATCH /campaigns/{id}/status -> adapter.get_object() (previous) -> adapter.update_status() -> event_bus.publish(StatusChanged) -> StatusUpdateResponse
```

---

## Knowledge Gaps

1. `lifecycle/campaign_lock.py` -- infrastructure exists but not wired into default launch path
2. `lifecycle/campaign_search.py` -- `CampaignSearchService` exists but defaults to `None` in DI assembly
3. Deployment topology (Dockerfile, docker-compose) not inspected
4. `docs/architecture/` and `docs/spikes/` may contain ADRs not captured here
5. No persistence layer -- service is stateless; idempotency cache is in-memory; both service clients are stubs
