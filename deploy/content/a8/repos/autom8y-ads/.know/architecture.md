---
domain: architecture
generated_at: "2026-03-27T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "1cfde11"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

**Language**: Python 3.12
**Framework**: FastAPI + Pydantic v2 + uvicorn
**Package**: `autom8_ads` (installed from `src/autom8_ads/`)
**Entry point**: `src/autom8_ads/app.py` -> `create_app()` -> FastAPI lifespan

---

## Package Structure

The source tree is rooted at `src/autom8_ads/` with 14 sub-packages:

| Package | Files | One-line Purpose |
|---|---|---|
| `api/` | 27 | FastAPI routers -- HTTP boundary layer |
| `cache/` | 2 | In-process TTL cache for active campaign trees |
| `cleanup/` | 6 | Cleanup pipeline -- score, decide, evict stale ad objects |
| `clients/` | 8 | Data, Asana, ad_optimizations, asset, intelligence clients (protocol + impls) |
| `events/` | 2 | In-process event bus with fire-and-forget dispatch |
| `guards/` | 8 | Pre-mutation safety checks -- budget, targeting, status, lifecycle enrichment |
| `intelligence/` | 6 | Vertical policy engine, mutation pattern analysis, creative bridge |
| `launch/` | 3 | Launch orchestration (idempotency, context building, service) |
| `lifecycle/` | 7 | Campaign search, locking, budget reconciliation, factory, strategies |
| `models/` | 15 | Pydantic domain models -- all shared data contracts |
| `platforms/` | 5 | AdPlatform protocol + Meta adapter + translator |
| `reconciliation/` | 6 | Business record reconciliation (discovery, resolution, monitoring) |
| `routing/` | 2 | Account router -- maps platform+spend to Meta account ID |
| Root | 4 | `app.py`, `config.py`, `dependencies.py`, `errors.py` |

**Hub packages** (imported by many others):
- `models/` -- consumed by every package
- `errors.py` -- consumed by api, cleanup, launch, lifecycle, platforms, routing
- `platforms/protocol.py` -- consumed by cleanup, launch, lifecycle
- `intelligence/models.py` -- consumed by guards, api, intelligence sub-modules

**Leaf packages** (import few siblings):
- `cache/` -- only imports `models/cache.py` and `platforms/`
- `routing/` -- only imports `errors.py` and `models/enums.py`
- `events/` -- only imports `models/events.py`
- `reconciliation/` -- imports `clients/`, `models/`, `errors.py`

**File counts**: 118 Python source files total.

---

## Layer Boundaries

The codebase uses a strict 4-layer model:

```
+--------------------------------------------------------+
|  HTTP Layer: api/  (routes, request parsing, DI wiring)|
|  Entry: create_app() in app.py                         |
+-----------------------+--------------------------------+
                        | FastAPI Depends()
+-----------------------v--------------------------------+
|  Service Layer: launch/, cleanup/, lifecycle/          |
|  LaunchService, CleanupService, AdFactory, V2Strategy  |
+-----------------------+--------------------------------+
                        | Protocol types
+-----------------------v--------------------------------+
|  Intelligence Layer: intelligence/, guards/            |
|  VerticalPolicyEngine, MutationPatternAnalyzer,       |
|  CreativePerformanceBridge, LifecycleEnrichmentGuard  |
+-----------------------+--------------------------------+
                        | Protocol types
+-----------------------v--------------------------------+
|  Infrastructure Layer: platforms/, clients/, cache/    |
|  MetaPlatformAdapter, DataCampaignClient, TreeCache,  |
|  IntelligenceDataClient, DataAssetClient              |
+--------------------------------------------------------+
```

**Import direction** (observed from import statements):
- `api/` imports `launch/`, `cleanup/`, `dependencies.py`, `models/`, `platforms/`, `intelligence/`
- `dependencies.py` imports `launch/`, `cleanup/`, `lifecycle/`, `clients/`, `routing/`, `platforms/`
- `launch/service.py` imports `clients/`, `models/`, `errors.py`, `launch/context.py`
- `lifecycle/strategies/v2_meta.py` imports `platforms/protocol.py`, `models/`, `errors.py`
- `cleanup/service.py` imports `cleanup/strategy.py` (Protocol), `clients/`, `events/`, `cache/`
- `platforms/meta/adapter.py` imports `models/`, `errors.py` -- does NOT import `launch/` or `cleanup/`
- `clients/data.py` imports from `autom8y_interop.data` (external SDK), `models/`
- `intelligence/vertical_policy.py` imports `intelligence/models.py`, `clients/intelligence_data.py` (via TYPE_CHECKING)
- `intelligence/mutation_patterns.py` imports `intelligence/config_normalizer.py`, `intelligence/models.py`, `clients/intelligence_data.py` (via TYPE_CHECKING)
- `intelligence/creative_bridge.py` imports `intelligence/models.py`, `clients/asset_data.py`
- `guards/lifecycle_enrichment.py` imports `guards/protocol.py`, `intelligence/models.py`, `intelligence/mutation_patterns.py`, `intelligence/vertical_policy.py` (via TYPE_CHECKING)

**Boundary enforcement patterns**:
- Protocols (`AdPlatform`, `LaunchStrategy`, `CleanupStrategy`, `EventBus`, `CleanupDataProtocol`) are defined at layer boundaries. Service layer talks to infrastructure via Protocol types, never concrete classes.
- TYPE_CHECKING guards are used extensively to prevent runtime circular imports. Example: `launch/service.py` uses `if TYPE_CHECKING: from autom8_ads.events.bus import EventBus`.
- `app.py` is the single wiring point -- it imports singletons and stores them on `app.state`. Routes access singletons via `dependencies.py` FastAPI Depends functions, not direct imports.
- The `models/` package is truly shared and has no imports from any sibling package.

---

## Entry Points and API Surface

### Application Entry Point

`src/autom8_ads/app.py` -> `create_app() -> FastAPI`

**Startup sequence** (in `lifespan()` async context manager):
1. `AdsConfig()` -- load config from env vars
2. Fail-fast guard: stub data client disallowed in non-LOCAL envs
3. `_init_auth(config)` -> `AuthClient` (JWKS pre-fetch)
4. `_init_meta_client(config)` -> `(MetaAdsClient | None, MetaPlatformAdapter)`
5. `_build_default_account_router(config)` -> `AccountRouter`
6. `LaunchIdempotencyCache(ttl_seconds=...)` -> idempotency store
7. `_init_campaign_lock(config, _Env)` -> `CampaignLock | NullCampaignLock`
8. `_init_data_clients(config)` -> `(DataCampaignProtocol, CleanupDataProtocol)`
9. `AsanaServiceClient | StubAsanaServiceClient`
10. `AdOptimizationsProtocol` client (stub or HTTP)
11. `DataReadProtocol` + `DataInsightProtocol` clients (via interop SDK resolvers)
12. `AsanaOfferClient` + `AsanaIntakeClient` (optional, for offer lifecycle)
13. `DataAssetProtocol` client (stub or HTTP)
14. `_init_intelligence_data_client(config)` -> `IntelligenceDataProtocol` (stub or HTTP)
15. `MutationPatternAnalyzer(intelligence_client=...)` -> mutation pattern analysis
16. `VerticalPolicyEngine(intelligence_client=..., insight_client=...)` -> vertical policy engine
17. `CreativePerformanceBridge(asset_client=..., insight_client=...)` -> creative performance bridge
18. `_init_event_bus()` -> `InProcessEventBus` with `EventLogSubscriber`
19. `_init_tree_cache(config, meta_client, adapter)` -> `TreeCache` + background refresh tasks

All singletons stored on `app.state.*`.

### HTTP Endpoints

| Method | Path | Router file | Purpose |
|---|---|---|---|
| `GET` | `/health` | `api/health.py` | Liveness probe (no I/O) |
| `GET` | `/ready` | `api/health.py` | Readiness probe (checks adapter, data client, asana) |
| `GET` | `/health/deps` | `api/health.py` | Dependency probe with tree cache staleness |
| `GET` | `/health/cleanup` | `api/health.py` | Cleanup subsystem health probe |
| `POST` | `/api/v1/offers/{offer_id}/launch` | `api/launch.py` | Launch ad hierarchy for an offer |
| `GET` | `/api/v1/accounts/{account_id}/campaigns` | `api/campaigns.py` | List campaigns with cursor pagination |
| `GET` | `/api/v1/campaigns/{campaign_id}` | `api/campaigns.py` | Fetch single campaign |
| `GET` | `/api/v1/campaigns/{campaign_id}/hierarchy` | `api/campaigns.py` | Campaign + ad groups + ads (cap 1000) |
| `GET` | `/api/v1/accounts/{account_id}/campaigns/active/tree` | `api/campaigns.py` | Active campaign tree (served from cache) |
| `POST` | `/api/v1/cleanup` | `api/cleanup.py` | Run cleanup pipeline (dry_run optional) |
| `GET` | `/api/v1/cleanup/history` | `api/cleanup.py` | Paginated cleanup run history |
| `GET` | `/api/v1/campaigns/{campaign_id}/insights` | `api/insights.py` | Campaign performance insights |
| `PATCH` | `/api/v1/campaigns/{campaign_id}/status` | `api/status.py` | Update campaign status |
| `GET` | `/api/v1/accounts/{account_id}` | `api/accounts.py` | Account details |
| `GET` | `/api/v1/ad-groups/{ad_group_id}` | `api/ad_groups.py` | Fetch single ad group |
| `GET` | `/api/v1/ad-groups/{ad_group_id}/ads` | `api/ad_groups.py` | List ads in ad group |
| `PATCH` | `/api/v1/ad-groups/{ad_group_id}` | `api/ad_groups.py` | Update ad group fields |
| `GET` | `/api/v1/ad-groups/{ad_group_id}/children` | `api/ad_groups.py` | List ad group children |
| `GET` | `/api/v1/ads/{ad_id}` | `api/ads.py` | Fetch single ad |
| `PATCH` | `/api/v1/ads/{ad_id}` | `api/ads.py` | Update ad fields |
| `GET` | `/api/v1/ads/{ad_id}/children` | `api/ads.py` | List ad children |
| `POST` | `/api/v1/assets/upload` | `api/assets.py` | Upload creative asset |
| `POST` | `/api/v1/creatives` | `api/creatives.py` | Create creative |
| `GET` | `/api/v1/creatives/{creative_id}` | `api/creatives.py` | Fetch creative |
| `GET` | `/api/v1/creatives/{creative_id}/assets` | `api/creatives.py` | List creative assets |
| `PATCH` | `/api/v1/campaigns/{campaign_id}` | `api/campaigns_patch.py` | Update campaign fields |
| `PATCH` | `/api/v1/ad-groups/{ad_group_id}` | `api/ad_groups_patch.py` | Update ad group fields (patch router) |
| `PATCH` | `/api/v1/ads/{ad_id}` | `api/ads_patch.py` | Update ad fields (patch router) |
| `POST` | `/api/v1/bulk` | `api/bulk.py` | Bulk operations across objects |
| `GET` | `/api/v1/intelligence/verticals/{vertical_key}` | `api/intelligence.py` | Get vertical intelligence data |
| `GET` | `/api/v1/intelligence/mutation-patterns/{campaign_id}` | `api/intelligence.py` | Get mutation patterns for campaign |
| `GET` | `/api/v1/intelligence/creative-performance/{creative_id}` | `api/intelligence.py` | Get creative performance intelligence |
| `GET` | `/api/v1/policies/verticals` | `api/policies.py` | List all vertical policies |
| `GET` | `/api/v1/policies/verticals/{vertical_key}` | `api/policies.py` | Get specific vertical policy |
| `PUT` | `/api/v1/policies/verticals/{vertical_key}/override` | `api/policy_override.py` | Override vertical policy thresholds |
| `POST` | `/api/v1/intelligence/mutation-analysis` | `api/mutation_patterns.py` | Run mutation pattern analysis |
| `POST` | `/api/v1/intelligence/rep-performance` | `api/mutation_patterns.py` | Rep performance (reserved, 403) |
| `POST` | `/api/v1/creative-performance` | `api/creative_performance.py` | Creative performance analysis |
| `POST` | `/api/v1/staging/validate-metadata-guid` | `api/staging_validation.py` | Validate metadata GUID for staging |
| `POST` | `/api/v1/reconciliation/discover` | `api/reconciliation.py` | Discover reconciliation candidates |
| `GET` | `/api/v1/reconciliation/status` | `api/reconciliation.py` | Reconciliation pipeline status |

All endpoints except `/health` and `/ready` require `Authorization: Bearer <JWT>` validated by `verify_jwt` dependency.

### Key Exported Interfaces

- `AdPlatform` (Protocol, `platforms/protocol.py`) -- 14-method contract: get/create/update/delete objects, get children, pagination, search, insights, conversion events, asset upload, creative creation. Implemented by `MetaPlatformAdapter`.
- `LaunchStrategy` (Protocol, `lifecycle/strategies/base.py`) -- `execute(intent, platform, extensions) -> LaunchResult`. Implemented by `V2MetaLaunchStrategy`.
- `CleanupStrategy` (Protocol, `cleanup/strategy.py`) -- `execute(intent, platform, data_client, *, tree_cache) -> CleanupResult`. Implemented by `MetaCleanupPipeline`.
- `EventBus` (Protocol, `events/bus.py`) -- `publish(event)`, `subscribe(type, handler)`. Implemented by `InProcessEventBus`.
- `CleanupDataProtocol` (Protocol, `clients/data.py`) -- 4 PV lifecycle state query methods + `record_cleanup_result`.
- `DataCampaignProtocol` (from `autom8y_interop.data`, re-exported in `clients/data.py`) -- campaign CRUD + search.
- `AsanaServiceProtocol` (`clients/asana.py`) -- `update_custom_fields(task_gid, ...)`.
- `IntelligenceDataProtocol` (Protocol, `clients/intelligence_data.py`) -- `get_mutation_outcome`, `get_mutation_outcomes_batch`, `get_mutation_trajectory`. Stub + HTTP triple for NEW-01/NEW-02 cross-service endpoints.
- `AdOptimizationsProtocol` (Protocol, `clients/ad_optimizations.py`) -- `record_optimization`, `get_latest_optimization`. Mutation config snapshot read/write.
- `DataAssetProtocol` (Protocol, `clients/asset_data.py`) -- `get_asset`, `get_assets_for_creative`, `get_asset_verticals`. Creative asset pipeline access.
- `Guard` (Protocol, `guards/protocol.py`) -- `check(object_type, context) -> GuardResult`. Pre-mutation safety check contract. Implemented by `BudgetGuard`, `TargetingGuard`, `StatusTransitionGuard`, `OfferLifecycleGuard`, `CreativeEnrichmentGuard`.
- `VerticalPolicyEngine` (`intelligence/vertical_policy.py`) -- Per-vertical threshold lookup with 4-tier confidence model. Normalizes 54 raw verticals to 15 canonical keys.
- `MutationPatternAnalyzer` (`intelligence/mutation_patterns.py`) -- Mutation history normalization, targeting correlation, budget ramp trajectory analysis.
- `CreativePerformanceBridge` (`intelligence/creative_bridge.py`) -- Creative swap detection, fatigue assessment, 5-level degradation hierarchy (FULL -> ASSET_TYPE -> VERTICAL -> ECOSYSTEM -> INSUFFICIENT).

---

## Key Abstractions

### Core Types

| Type | File | Purpose | Consumers |
|---|---|---|---|
| `AdsModel` | `models/base.py` | Frozen Pydantic base (immutable, `extra="ignore"`) | All model classes |
| `AdsError` (hierarchy) | `errors.py` | 7-class error hierarchy with HTTP status codes; base for all domain exceptions | All packages |
| `Platform` / `AdObjectType` / `DomainStatus` | `models/enums.py` | StrEnum taxonomy for platforms, object types, statuses | All packages |
| `PlatformAdObject` | `platforms/types.py` | Minimal wrapper: `id`, `type`, `status`, `name`, `raw` dict -- returned by all AdPlatform reads | `platforms/`, `lifecycle/`, `cleanup/` |
| `LaunchIntent` | `models/launch.py` | Validated launch specification: offer_id, account, platform, budget, targeting, asset_ids | `launch/`, `lifecycle/` |
| `LaunchResult` | `models/launch.py` | Result of a launch: success bool, campaign/adset/ad/creative IDs, warnings | `launch/`, `api/` |
| `DomainEvent` | `models/events.py` | Base for 10 event types: CampaignCreated, CleanupStarted, ObjectEvicted, TreeDriftDetected, etc. | `events/`, `launch/`, `cleanup/`, `api/` |
| `CachedTree` | `models/cache.py` | Snapshot of active campaign tree: data, account_id, built_at, counts | `cache/`, `api/`, `platforms/` |
| `CleanupIntent` | `models/cleanup.py` | Cleanup job specification: account_id, platform, dry_run, scope, scoring_config | `cleanup/`, `api/` |
| `AdsConfig` | `config.py` | Pydantic-settings config with ADS_ prefix; includes Meta credentials, env, service URLs, stub flags | `app.py`, `dependencies.py`, `api/` |

### Design Patterns

1. **Protocol-based dependency inversion**: All cross-layer contracts are Python `typing.Protocol` classes (runtime_checkable). `AdPlatform`, `LaunchStrategy`, `CleanupStrategy`, `EventBus`, `DataCampaignProtocol`, `CleanupDataProtocol`, `AsanaServiceProtocol`, `CampaignMatcher` are all Protocols. This enables stub injection in tests without subclassing.

2. **Singleton-on-app-state DI**: Singletons (adapter, data client, event bus, tree cache, campaign lock) are created once in `lifespan()` and stored on `app.state`. FastAPI Depends functions read from `request.app.state` -- `get_launch_service()`, `get_cleanup_service()`, `get_platform_adapter()` etc. Services are assembled per-request from singleton components.

3. **Stub/real client toggle**: Every external dependency has a stub counterpart. `use_stub_data_client=True` (default LOCAL) -> `StubDataCampaignClient` + `StubCleanupDataClient`. `use_stub_asana_client=True` -> `StubAsanaServiceClient`. Real clients are gated by `AdsConfig` flags with fail-fast guards in non-LOCAL environments.

4. **Strategy pattern for launch**: `AdFactory` delegates entirely to `LaunchStrategy.execute()`. The only strategy is `V2MetaLaunchStrategy` (ADR-ADS-002 mandates no V1 legacy code). BudgetReconciler and CampaignSearchService are optional injected collaborators.

5. **5-phase cleanup pipeline**: `MetaCleanupPipeline` implements discover -> score -> decide -> execute -> report. Single tree snapshot per run (D-PIPE-2). Semaphore concurrency control (D-PIPE-7, `asyncio.Semaphore`). CampaignSearchService reuse check in execute phase (D-PIPE-3).

6. **Name encoding convention**: Ad object names encode structured metadata as bullet-delimited fields. `models/name_encoding.py` defines `CampaignNameFields`, `AdGroupNameFields`, `AdNameFields` and their encode/decode functions. The cleanup scoring and search-before-create both decode names to extract PV keys and objective info.

7. **Non-fatal side effects**: Asana writeback, data persistence, event emission, and cleanup result recording are all wrapped in try/except and log warnings on failure. Failures in side effects never fail the primary operation. Pattern: `await self._try_*(...)`.

8. **Guard chain with enrichment**: `guards/protocol.py` defines `Guard` Protocol and `GuardChain` which evaluates guards in order, short-circuiting on first deny. `LifecycleEnrichmentGuard` wraps structural guards and enriches responses with intelligence data (vertical policy + mutation patterns) via 2-second timeout. On timeout/error, returns structural-only result with reduced confidence (0.5).

9. **Intelligence layer degradation**: Intelligence components (`VerticalPolicyEngine`, `MutationPatternAnalyzer`, `CreativePerformanceBridge`) are non-blocking and degrade gracefully. `CreativePerformanceBridge` uses a 5-level degradation hierarchy: FULL -> ASSET_TYPE -> VERTICAL -> ECOSYSTEM -> INSUFFICIENT. All intelligence endpoints return useful results even with partial data.

10. **21-section offer lifecycle mapping**: `guards/lifecycle_enrichment.py` defines `SECTION_ACTION_MAP` mapping all 21 active Asana offer lifecycle sections to structured `SectionAction` guidance (optimization intent, allowed mutation types, parameter direction, guard behavior). Top 5 sections by volume have data-derived mappings; remaining 16 use conservative defaults.

---

## Data Flow

### Launch Pipeline

```
POST /api/v1/offers/{offer_id}/launch
  -> api/launch.py -> LaunchService.launch(offer_id, request)
    1. IdempotencyCache.get_or_set_in_progress(offer_id, platform)
       -> return cached result if present (duplicate suppression, TTL=300s)
    2. AccountRouter.route(platform, weekly_spend_cents)
       -> evaluate AccountRule list -> return account_id string
    3. build_launch_intent() + build_platform_extensions()
       -> LaunchIntent + PlatformExtensions (from models/launch.py)
    4. AdFactory.launch(intent, extensions)
       -> V2MetaLaunchStrategy.execute(intent, platform, extensions)
         a. Validate (budget, spend, algo_version, asset_ids)
         b. CampaignSearchService.find_or_create_campaign() [optional]
            -> DataCampaignProtocol.search_campaigns() (autom8_data HTTP)
            -> AdPlatform.search_objects() (Meta API)
            -> if match: return campaign_id (reuse path)
         c. AdPlatform.create_object(CAMPAIGN) -> platform_id
         d. AdPlatform.create_object(AD_GROUP) -> platform_id
            [budget error: BudgetReconciler.handle_budget_error() + retry]
         e. AdPlatform.create_creative() -> platform_id
         f. AdPlatform.create_object(AD) -> platform_id
         g. Return LaunchResult
    5. _try_persist(): DataCampaignProtocol.record_campaign/ad_set/ad/creative
       (non-fatal -- each entity independently, warns on failure)
    6. _try_asana_writeback(): AsanaServiceProtocol.update_custom_fields()
       (non-fatal)
    7. _try_emit_campaign_created(): EventBus.publish(CampaignCreated)
       (non-fatal, asyncio.create_task)
    8. IdempotencyCache.complete(offer_id, platform, result)
  <- LaunchResponse{status, result, trace_id}
```

### Campaigns Read Pipeline (Tree Cache Path)

```
GET /api/v1/accounts/{account_id}/campaigns/active/tree
  -> api/campaigns.py -> get_active_campaign_tree()
    1. TreeCache.get(account_id) -> CachedTree | None
    2. Cache miss: TreeCache.get_or_fetch()
       -> asyncio.Event deduplication for cold-start concurrency
       -> MetaPlatformAdapter.get_active_tree(account_id, translator)
         -> MetaAdsClient.get_account_campaigns_tree() [single SDK call]
         -> Filter ACTIVE-only campaigns/adsets/ads
         -> Return CachedTree{data: List[CampaignTree], built_at, counts}
    3. Cache hit: slice by max_campaigns, set X-Cache-Age header
  <- JSONResponse{items: CampaignTree[], total_campaigns, total_active_campaigns}

Background refresh: periodic_tree_refresh() asyncio task
  -> every cache_refresh_interval_seconds (default 900s = 15min)
  -> stale-while-revalidate: keeps serving old cache on refresh failure
```

### Cleanup Pipeline

```
POST /api/v1/cleanup
  -> api/cleanup.py -> CleanupService.cleanup(intent, platform)
    1. CampaignLock.acquire(f"campaign_lock:{account_id}:cleanup", owner_id)
       -> raises AdsValidationError if lock held
    2. EventBus.publish(CleanupStarted)
    3. MetaCleanupPipeline.execute(intent, platform, cleanup_data_client)
       Phase 1 (discover): MetaPlatformAdapter.get_cleanup_tree()
         -> no status filter: ALL campaigns/adsets/ads
         -> CleanupDataProtocol: get_inactive_pvs, get_activating_pvs,
           get_awaiting_rep_update_pvs, get_active_pvs
         -> Build skip_cleanup_pvs set (Gate 5 protection)
       Phase 2 (score): DefaultScoringStrategy per candidate
         -> ScoringConfig.weights applied to status, age, PV membership
         -> composite_score float
       Phase 3 (decide): EVICT | RETAIN | SKIP per candidate
         -> Threshold comparison, hard gates (skip_pvs, object_type filter)
         -> CampaignSearchService.find_or_create_campaign() for reuse check (D-PIPE-3)
       Phase 4 (execute): for each EVICT candidate:
         -> dry_run=False: AdPlatform.delete_object()
         -> Semaphore concurrency control (D-PIPE-7)
         -> EventBus.publish(ObjectEvicted | ObjectRetained) per object
       Phase 5 (report): CleanupDataProtocol.record_cleanup_result()
         -> aggregated CleanupResult{evaluated, evicted, retained, errors}
    4. CleanupHistory.record(result, account_id, run_id) [in-memory ring buffer]
    5. CampaignLock.release()
  <- CleanupResponse{status, result, trace_id, account_id}
```

### Configuration Loading

```
App startup: AdsConfig() [pydantic-settings]
  -> env prefix: ADS_
  -> canonical aliases: AUTOM8Y_ENV (Tier 1), AUTOM8Y_ASANA_URL / AUTOM8Y_DATA_URL (Tier 3)
  -> auth toggle: AUTH_DEV_MODE (Tier 2)
  -> secrets: ADS_META_ACCESS_TOKEN, ADS_META_APP_ID, ADS_META_APP_SECRET
  -> multi-account: ADS_META_ACCOUNT_IDS (comma-separated string -> list[str])
  -> constraints: auth_disabled only allowed in LOCAL; meta_account_ids required in non-LOCAL
```

### Event Bus Data Flow

```
InProcessEventBus.publish(event)
  -> asyncio.create_task(_safe_dispatch(handler, event)) per subscriber
  -> errors isolated: logged, never propagated

Subscriptions wired in _init_event_bus():
  EventLogSubscriber -> CampaignCreated, AdGroupCreated, AdCreated,
    BudgetAdjusted, StatusChanged, CleanupStarted, CleanupCompleted,
    ObjectEvicted, ObjectRetained, TreeDriftDetected
```

### Intelligence Data Flow (Project Alchemy)

```
GET /api/v1/intelligence/verticals/{vertical_key}
  -> api/intelligence.py -> VerticalPolicyEngine.get_policy(vertical_key)
    1. Normalize vertical_key via CANONICAL_VERTICAL_MAP (54 -> 15 canonical)
    2. Check in-memory policy cache -> hit: return cached VerticalPolicy
    3. Cache miss: query insight_client for vertical performance data
    4. Compute thresholds (budget, CPS, CPL, booking rate, cadence)
    5. Assign confidence tier based on sample size
    6. Cache policy, return VerticalPolicy
  <- VerticalPolicy{budget, cps, cpl, booking_rate, cadence, confidence, metadata}

POST /api/v1/intelligence/mutation-analysis
  -> api/mutation_patterns.py -> MutationPatternAnalyzer.analyze(request)
    1. Fetch mutation history from IntelligenceDataProtocol
    2. Normalize config JSON via ConfigNormalizer (3 schema eras)
    3. Classify mutation types (targeting, budget, creative, status)
    4. Compute pre/post attribution windows (14d pre, 3d settle, 7d post)
    5. Detect confounding mutations, apply confidence penalties
    6. Return targeting correlations, budget ramp, CPS impact
  <- MutationPatternResponse{correlations, ramp_result, cps_impact, confidence}

POST /api/v1/creative-performance
  -> api/creative_performance.py -> CreativePerformanceBridge.analyze(request)
    1. Resolve creative asset identity via DataAssetProtocol
    2. Detect creative swaps in mutation history
    3. Correlate pre/post swap performance
    4. Compute composite confidence score
    5. Assess creative fatigue (frequency, staleness, performance decay)
    6. Degrade through hierarchy: FULL -> ASSET_TYPE -> VERTICAL -> ECOSYSTEM -> INSUFFICIENT
  <- CreativePerformanceResult{swap_records, fatigue, recommendations, confidence}
```

### Guard Enrichment Data Flow

```
Guard chain evaluate (pre-mutation):
  -> GuardChain.evaluate(object_type, context)
    1. Structural guards (budget, targeting, status, offer lifecycle) -> GuardResult
    2. If structural allow -> LifecycleEnrichmentGuard.enrich(guard_result, context)
       a. Parallel queries (asyncio.gather, 2s timeout):
          - VerticalPolicyEngine.get_policy(vertical_key)
          - MutationPatternAnalyzer.query_recent(campaign_id)
       b. Build data-informed Recommendation (confidence, guidance, section_action)
       c. Return EnrichedGuardResponse{structural_result, recommendations}
    3. Timeout/error: return structural-only result, confidence=0.5
  <- EnrichedGuardResponse | GuardResult
```

---

## Knowledge Gaps

1. `src/autom8_ads/cleanup/scoring.py` and `src/autom8_ads/cleanup/strategy.py` -- `DefaultScoringStrategy` implementation details and `CleanupStrategy` protocol signature not fully read (limited by turn budget). The 5-phase pipeline structure is documented from the pipeline.py header.
2. `src/autom8_ads/lifecycle/budget.py` and `src/autom8_ads/lifecycle/campaign_lock.py` -- `BudgetReconciler` and `CampaignLock` / `NullCampaignLock` internals not read. Their interfaces are understood from consumers (`V2MetaLaunchStrategy`, `app.py`).
3. `src/autom8_ads/lifecycle/campaign_matcher.py` -- `DefaultCampaignMatcher` logic not read. Interface understood from `CampaignSearchService` constructor.
4. `src/autom8_ads/platforms/meta/params.py` and `src/autom8_ads/platforms/translator.py` -- `build_campaign_params`, `build_ad_set_params`, `MetaTranslator.to_campaign/to_ad_group/to_ad` not read in detail. Their roles are clear from consumers.
5. `src/autom8_ads/cleanup/tree_sync.py` -- `TreeSyncService` (emits `TreeDriftDetected`) not fully read; its role is inferred from the event model.
6. No `__main__.py` or `Dockerfile` entry point traced -- the application is launched via uvicorn targeting `autom8_ads.app:create_app()` or similar, but the exact uvicorn command was not verified from source.
7. `src/autom8_ads/reconciliation/` -- Full reconciliation pipeline (discovery, resolution, monitoring) internals not fully read. Roles inferred from file names and API surface.
8. `src/autom8_ads/intelligence/models.py` -- Full model inventory not enumerated. Contains ~30+ domain models for vertical policies, mutation patterns, creative performance, confidence scoring, and recommendations.
