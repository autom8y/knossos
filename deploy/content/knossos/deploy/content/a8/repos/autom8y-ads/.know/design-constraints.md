---
domain: design-constraints
generated_at: "2026-03-25T12:12:38Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "1cfde11"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: V2MetaLaunchStrategy Naming Mismatch
- **Type**: Naming mismatch / semantic gap
- **Location**: `src/autom8_ads/lifecycle/strategies/v2_meta.py`
- **Description**: Named `V2MetaLaunchStrategy` but "This is the ONLY strategy. There is no V1 strategy" (line 42). The V2 prefix is a historical artifact from a superseded legacy system.
- **Ideal resolution**: Rename to `MetaLaunchStrategy` during WS-2 Phase 2b when a second platform adapter arrives.
- **Resolution cost**: LOW (rename + 2 import sites). Should coincide with platform-agnostic extraction.

### TENSION-002: Lifecycle Layer Imports Platform-Specific Code
- **Type**: Layering violation / semantic coupling
- **Location**: `src/autom8_ads/lifecycle/strategies/v2_meta.py` lines 19-25; `src/autom8_ads/platforms/meta/params.py`
- **Description**: `V2MetaLaunchStrategy` imports from `platforms/meta/constants.py` and `platforms/meta/params.py`. The lifecycle layer reaches into platform-specific code. Additionally `build_ad_set_params` takes `LaunchIntent` (domain type), creating bidirectional dependency.
- **Ideal resolution**: Param functions should accept primitive arguments or platform-specific ACL input types.
- **Resolution cost**: MEDIUM. Full resolution is WS-2 Phase 2b.

### TENSION-003: `optimization_goal` is a Raw String
- **Type**: Under-engineering / premature flattening
- **Location**: `src/autom8_ads/models/ad_group.py` line 29
- **Description**: `optimization_goal: str | None` stores platform-specific strings directly. No domain enum exists.
- **Ideal resolution**: Define `OptimizationGoal` domain enum; store platform-specific string in `extensions` dict.
- **Resolution cost**: MEDIUM.

### TENSION-004: AdGroup vs AdSet Naming Collision
- **Type**: Naming mismatch / terminology collision
- **Location**: `src/autom8_ads/models/ad_group.py`; `src/autom8_ads/platforms/meta/adapter.py` line 171
- **Description**: Domain uses `AdGroup`/`AD_GROUP`; Meta API uses `adset`. `clients/data.py` uses `AdSetWriteRequest`/`AdSetWriteResponse` (Meta terminology).
- **Ideal resolution**: Canonicalize on `AdGroup`/`ad_group` across all models and client interfaces.
- **Resolution cost**: LOW-MEDIUM.

### TENSION-005: Dual Discovery Paths in CleanupPipeline
- **Type**: Duplication / parallel code paths
- **Location**: `src/autom8_ads/cleanup/pipeline.py` -- `_discover()` vs `_discover_from_adapter()`
- **Description**: Two discovery methods exist: `_discover()` (uses `tree_cache`, active-only) and `_discover_from_adapter()` (uses `adapter.get_cleanup_tree()`, all statuses). Both contain identical PV-state accumulation logic (Gates 3/5). The adapter path was added for SCAR-020, but the tree_cache path was retained as fallback. This duplication creates maintenance burden -- any change to PV gates must be applied to both methods.
- **Ideal resolution**: Extract shared PV-state logic into a helper; remove tree_cache fallback once adapter injection is guaranteed.
- **Resolution cost**: LOW (refactor, no behavior change).

### TENSION-006: Health Contract Models are Copy-Paste
- **Type**: Under-engineering / missing abstraction
- **Location**: `src/autom8_ads/api/health_models.py` lines 3-4
- **Description**: Health models explicitly marked as "Copy this module into each satellite's codebase" rather than a shared package.
- **Ideal resolution**: Extract to `autom8y_health` SDK package.
- **Resolution cost**: MEDIUM.

### TENSION-007: Domain Model Split Between `models/search.py` and `clients/data.py`
- **Type**: Naming mismatch / split location
- **Location**: `src/autom8_ads/models/search.py` vs `src/autom8_ads/clients/data.py`
- **Description**: `CampaignWriteRequest` and `CampaignSearchResult` in `models/search.py`, but `AdSetWriteRequest`, `AdSetWriteResponse`, etc. remain in `clients/data.py`.
- **Ideal resolution**: Move all data-service contract models to `models/`.
- **Resolution cost**: LOW.

### TENSION-008: Hardcoded Meta-Specific Values in Param Builders
- **Type**: Missing abstraction / hardcoded behavior
- **Location**: `src/autom8_ads/platforms/meta/params.py` lines 66-68
- **Description**: `"optimization_goal": "LEAD_GENERATION"`, `"billing_event": "IMPRESSIONS"`, `"bid_amount": intent.daily_budget_cents` are string literals not in `constants.py`. The `params: dict[str, Any]` interface provides zero type safety.
- **Resolution cost**: LOW for constant extraction; HIGH for typed param models.

### TENSION-009: Idempotency `clear()` Affects All In-Flight Launches
- **Type**: Under-engineering / missing guard
- **Location**: `src/autom8_ads/launch/service.py` lines 145-163
- **Description**: On error, `LaunchIdempotencyCache.clear()` removes ALL entries, not just the failed offer. A single failed launch invalidates idempotency for concurrent launches.
- **Ideal resolution**: Add targeted `remove(offer_id, platform)` method.
- **Resolution cost**: LOW.

### TENSION-010: Interop Placeholder Subclassing Pattern
- **Type**: Naming mismatch / type-system friction
- **Location**: `src/autom8_ads/clients/data.py`
- **Description**: Models from `autom8y_interop` are imported with `_Interop*` prefixed aliases, then subclassed locally with typed fields. The `model_dump()` calls before crossing the interop boundary create runtime overhead and `# type: ignore[override]` suppressions accumulate. This is a Phase 2 interop migration artifact.
- **Ideal resolution**: Phase 3 interop migration to fully typed SDK models (removes subclassing).
- **Resolution cost**: MEDIUM (requires interop SDK changes).

---

## Trade-off Documentation

### TRADE-001: Stub-First Integration Pattern
External services use stubs as default in LOCAL. Real clients activated by config flags. Cost: all persistence is no-op in LOCAL. Benefit: local dev and tests run without external infra.

### TRADE-002: In-Process Idempotency Cache
In-memory dict with TTL. No distributed state. Multi-instance deployment allows duplicate launches. Chosen because: single-instance deployment is the current target. When horizontal scaling arrives, TENSION-009 must be fixed first.

### TRADE-003: PaginatedResponse is Not Generic
Concrete `items: list[CampaignSummary]` instead of `Generic[T]`. Pydantic generics complexity cited. Adding a second list endpoint requires a new concrete type.

### TRADE-004: AdPlatform Protocol Uses `dict[str, Any]` for Params
Untyped dicts at protocol boundary. Zero compile-time safety. Runtime errors only surface during actual API calls. Chosen because: Meta API params change frequently and typed models create high maintenance burden.

### TRADE-005: `asyncio.sleep(2.0)` in Budget Retry
Hardcoded 2-second delay before retry after budget fix. Not configurable. Adds latency to every budget error recovery.

### TRADE-006: CampaignSearchService Conditionally Wired
`CampaignSearchService` is now wired when `CampaignLock` is available (non-LOCAL environments with DynamoDB). In LOCAL, search-before-create is skipped. This means duplicate campaigns can be created in local dev.

### TRADE-007: Single-Worker Tree Cache Assumption
`TreeCache` in `src/autom8_ads/cache/tree_cache.py` line 22 notes "Assumes single uvicorn worker." Multi-worker deployment would require shared-memory or Redis-backed cache. The background refresh `asyncio.Task` only runs in one worker.

---

## Abstraction Gap Mapping

### GAP-001: `optimization_goal` Lacks Domain Enum
`src/autom8_ads/models/ad_group.py` line 29. Raw string from Meta API. No canonical enum. Same as TENSION-003.

### GAP-002: No `account_id` in Campaign Persistence
`src/autom8_ads/launch/service.py` line 208: `account_id=""` passed despite being available in `LaunchIntent`.

### GAP-003: `PaginatedResponse` Locked to `CampaignSummary`
`src/autom8_ads/models/responses.py` lines 26-35. Each new list endpoint requires a new concrete paginated type.

### GAP-004: Budget Default Value Duplicated in 3 Files
`35000` (cents = $350/week) appears in: `launch/service.py` line 24, `launch/context.py` line 80, `lifecycle/strategies/v2_meta.py` line 298. Not shared via a single constant.

### GAP-005: PV-State Accumulation Duplicated in Pipeline
`src/autom8_ads/cleanup/pipeline.py` -- identical PV-state fetch + set construction in `_discover()` and `_discover_from_adapter()`. Maintenance burden: any change to Gate 3/5 logic must be applied twice.

### GAP-006: TreeSyncService Not Wired in Default Path
`src/autom8_ads/cleanup/tree_sync.py` defines `TreeSyncService` and emits `TreeDriftDetected` events, but it is not instantiated in `app.py` lifespan. The drift detection feature exists in code but is dormant.

---

## Load-Bearing Code

### LBC-001: `V2MetaLaunchStrategy.execute()` -- Sole Production Launch Path
**Location**: `src/autom8_ads/lifecycle/strategies/v2_meta.py` lines 65-224
**Why**: Only `LaunchStrategy` implementation. No fallback, no alternative. Step ordering (CAMPAIGN -> AD_GROUP -> AD) is a Meta API requirement.
**Must not**: Inline-refactor or split without validating full integration test suite. Highest blast radius in the codebase.

### LBC-002: `AdPlatform` Protocol
**Location**: `src/autom8_ads/platforms/protocol.py`
**Why**: 6+ files depend on it. `@runtime_checkable`. Method signature changes cascade to all adapters and callers.
**Must not**: Add/remove methods without updating `MetaPlatformAdapter` and all test mocks.

### LBC-003: `models/enums.py` -- 43+ Import Sites
**Why**: Shared vocabulary for entire system. Renaming enum values is a breaking change.
**Must not**: Rename existing enum values or change string representations. Adding new variants is safe.

### LBC-004: `NameEncoding` + `CampaignNameFields` -- Encoding Contract
**Location**: `src/autom8_ads/models/name_encoding.py`
**Why**: Field order is the encoding schema. Existing campaigns become undecodeable if schema changes.
**Must not**: Reorder, add, or remove fields. Do not change separator (U+2022).

### LBC-005: `build_ad_set_params` INVARIANT -- No `daily_budget`
**Location**: `src/autom8_ads/platforms/meta/params.py` lines 44-46
**Why**: SCAR-003. Including `daily_budget` triggers Meta error 1885621.
**Must not**: Add `daily_budget` to ad set params under any circumstances.

### LBC-006: `LaunchIdempotencyCache._make_key()` -- Composite Key Format
**Location**: `src/autom8_ads/launch/idempotency.py` line 76
**Why**: Key format `"{offer_id}:{platform.value}"` is the idempotency contract.
**Must not**: Change key format without coordinated migration.

### LBC-007: `CampaignLock` Conditional Wiring
**Location**: `src/autom8_ads/app.py` `_init_campaign_lock()` and `src/autom8_ads/dependencies.py` `get_launch_service()`
**Why**: `CampaignSearchService` is only injected when `campaign_lock` is not `NullCampaignLock`. Removing the lock condition silently enables search-before-create without distributed locking protection.
**Must not**: Enable `CampaignSearchService` without a real `CampaignLock`.

### LBC-008: H-01 Stub Guard in Lifespan
**Location**: `src/autom8_ads/app.py` lifespan, lines 50-58
**Why**: Prevents `StubDataCampaignClient` from running in non-LOCAL environments. Bypassing this guard allows stub no-op persistence in production.
**Must not**: Remove or weaken without replacing with real client validation.

---

## Evolution Constraints

### EC-001: Adding a Second Platform (TikTok)
Requires: new adapter implementing `AdPlatform`, new strategy implementing `LaunchStrategy`, factory/service platform selection, account routing config expansion.
Frozen: `AdPlatform` protocol must remain stable. `V2MetaLaunchStrategy` name and direct import in `dependencies.py`.

### EC-002: Replacing Stub Clients
Swap-safe: both protocols use `@runtime_checkable`. Swap point: `app.py` lifespan.
Constraint: Real client schema must match protocols exactly. `StubDataCampaignClient` still available for LOCAL.

### EC-003: CampaignSearchService Activation (RESOLVED)
Previously unwired. Now conditionally wired when `CampaignLock` is real (non-LOCAL). Full activation requires DynamoDB table provisioning.

### EC-004: Name Encoding Schema is Frozen
Field order, field names, separator character frozen for all deployed campaigns. Adding new fields requires append-only schema evolution.

### EC-005: Horizontal Scaling Requires Idempotency Migration
`LaunchIdempotencyCache` is in-process. TENSION-009 `clear()` bug must be fixed before distributed cache. `TreeCache` also requires migration (TRADE-007).

### EC-006: JWT Validation is Real SDK (RESOLVED)
Previously structural-only (3-part dot check). Now uses `autom8y-auth` SDK with RS256/JWKS validation. The `AuthClient` is initialized in lifespan with JWKS pre-fetch.

### EC-007: `AdPlatform` Protocol Method Count Growth
Protocol currently has 14 methods. `get_cleanup_tree()` was added for SCAR-020. Each new method requires implementation in `MetaPlatformAdapter` and updates to all test mocks (`make_mock_platform()`). Growth beyond ~20 methods should trigger protocol splitting.

---

## Risk Zones

### RISK-001: Budget Reconciler Not Activated by Default
**Location**: `src/autom8_ads/dependencies.py` line 63
**Severity**: MEDIUM. Budget-too-low errors cause launch failures without auto-adjustment. `BudgetReconciler` exists in `src/autom8_ads/lifecycle/budget.py` but is never injected.

### RISK-002: `CampaignSearchService._try_import()` Silently Fails
**Severity**: LOW-MEDIUM. System degrades gracefully but import failures cause compounding search inefficiency.

### RISK-003: Idempotency `clear()` Affects All In-Flight Launches
**Location**: `src/autom8_ads/launch/service.py` lines 145-146, 162-163
**Severity**: MEDIUM. Under concurrent load, single failure clears all idempotency protection. Same as TENSION-009.

### RISK-004: Health Check Reports Dependencies as Healthy Based Only on None-Check
**Location**: `src/autom8_ads/api/health.py` lines 45-73
**Severity**: LOW. No connectivity probing; broken connections show as `"ok"`.

### RISK-005: `asyncio.to_thread()` in CampaignLock Has No Timeout
**Location**: `src/autom8_ads/lifecycle/campaign_lock.py` lines 66-78
**Severity**: MEDIUM (CampaignSearchService is now wired, so this lock is called on every launch in non-LOCAL envs). A hung DynamoDB call blocks the launch indefinitely.

### RISK-006: TreeSyncService is Dormant
**Location**: `src/autom8_ads/cleanup/tree_sync.py`
**Severity**: LOW. Tree drift detection events are defined and subscribed but the service that emits them (`TreeSyncService`) is not instantiated. If cache drift occurs, it will not be detected until manual inspection.

### RISK-007: Placeholder Meta Account ID in Config Default
**Location**: `src/autom8_ads/config.py`
**Severity**: LOW (guarded by `model_post_init` in non-LOCAL). In LOCAL, the placeholder `"act_PLACEHOLDER"` is used, which may produce confusing errors if accidentally sent to real Meta API.

---

## Knowledge Gaps

1. `PlatformAdObject.raw` internal schema from real `autom8y_meta` SDK is unknown
2. `autom8y_config.Autom8yBaseSettings` base class behavior not visible in this codebase
3. `autom8y_telemetry.instrument_app()` security layer additions unknown
4. `autom8y_meta.MetaConfig` full contract not visible
5. Whether `TreeSyncService` is intentionally dormant or a missed wiring is unclear
6. `autom8y_interop` Phase 3 migration timeline is not documented in any ADR
