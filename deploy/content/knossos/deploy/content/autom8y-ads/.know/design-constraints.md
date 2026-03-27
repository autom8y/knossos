---
domain: design-constraints
generated_at: "2026-03-01T12:42:56Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "762ed0e"
confidence: 0.88
format_version: "1.0"
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

### TENSION-005: CampaignLock is Over-Engineered for Current Use
- **Type**: Over-engineering / premature abstraction
- **Location**: `src/autom8_ads/lifecycle/campaign_lock.py`; `src/autom8_ads/lifecycle/campaign_search.py`
- **Description**: DynamoDB-backed distributed lock exists but `CampaignSearchService` is not wired by default. `NullCampaignLock` exists because the real lock cannot be used in most environments.
- **Resolution cost**: LOW for clarification; MEDIUM for full integration.

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

### TENSION-010: `data_writes_enabled` Config Dead Field
- **Type**: Missing abstraction / config dead field
- **Location**: `src/autom8_ads/config.py` line 72
- **Description**: `data_writes_enabled: bool` defined but never referenced anywhere in source code.
- **Ideal resolution**: Implement gating logic or remove the unused field.
- **Resolution cost**: LOW.

---

## Trade-off Documentation

### TRADE-001: Stub-First Integration Pattern
External services use stubs as default. Real clients not yet available. Cost: all persistence is no-op in production.

### TRADE-002: In-Process Idempotency Cache
In-memory dict with TTL. No distributed state. Multi-instance deployment allows duplicate launches.

### TRADE-003: PaginatedResponse is Not Generic
Concrete `items: list[CampaignSummary]` instead of `Generic[T]`. Pydantic generics complexity cited. Adding a second list endpoint requires a new concrete type.

### TRADE-004: AdPlatform Protocol Uses `dict[str, Any]` for Params
Untyped dicts at protocol boundary. Zero compile-time safety. Runtime errors only surface during actual API calls.

### TRADE-005: `asyncio.sleep(2.0)` in Budget Retry
Hardcoded 2-second delay before retry after budget fix. Not configurable. Adds latency to every budget error recovery.

### TRADE-006: CampaignSearchService Not Wired by Default
`budget_reconciler` and `campaign_search` both default to `None`. Infrastructure exists; wiring does not.

---

## Abstraction Gap Mapping

### GAP-001: `optimization_goal` Lacks Domain Enum
`src/autom8_ads/models/ad_group.py` line 29. Raw string from Meta API. No canonical enum.

### GAP-002: No `account_id` in Campaign Persistence
`src/autom8_ads/launch/service.py` line 208: `account_id=""` passed despite being available in `LaunchIntent`.

### GAP-003: `PaginatedResponse` Locked to `CampaignSummary`
`src/autom8_ads/models/responses.py` lines 26-35. Each new list endpoint requires a new concrete paginated type.

### GAP-004: `CreativeSpec` is Not a Domain Model
`src/autom8_ads/platforms/meta/params.py` lines 99-110. `build_creative_spec()` returns a bare dict. Creative content has no domain representation.

### GAP-005: Budget Default Value Duplicated in 3 Files
`35000` (cents = $350/week) appears in: `launch/service.py` line 24, `launch/context.py` line 80, `lifecycle/strategies/v2_meta.py` line 298. Not shared via a single constant.

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
**Must not**: Reorder, add, or remove fields. Do not change `SEPARATOR` (U+2022).

### LBC-005: `build_ad_set_params` INVARIANT -- No `daily_budget`
**Location**: `src/autom8_ads/platforms/meta/params.py` lines 44-46
**Why**: SCAR-003. Including `daily_budget` triggers Meta error 1885621.
**Must not**: Add `daily_budget` to ad set params under any circumstances.

### LBC-006: `LaunchIdempotencyCache._make_key()` -- Composite Key Format
**Location**: `src/autom8_ads/launch/idempotency.py` line 76
**Why**: Key format `"{offer_id}:{platform.value}"` is the idempotency contract.
**Must not**: Change key format without coordinated migration.

---

## Evolution Constraints

### EC-001: Adding a Second Platform (TikTok)
Requires: new adapter, new strategy, factory/service platform selection, account routing config.
Frozen: `AdPlatform` protocol must remain stable. `V2MetaLaunchStrategy` name and direct import in `dependencies.py`.

### EC-002: Replacing Stub Clients
Swap-safe: both protocols use `@runtime_checkable`. Swap point: `app.py` lines 116-117.
Constraint: `data_writes_enabled` does not gate writes. Real client schema must match protocols exactly.

### EC-003: CampaignSearchService Activation
Requires DynamoDB table provisioning, config fields, wiring in `get_launch_service()`.

### EC-004: Name Encoding Schema is Frozen
Field order, field names, separator character frozen for all deployed campaigns.

### EC-005: Horizontal Scaling Requires Idempotency Migration
`LaunchIdempotencyCache` is in-process. TENSION-009 `clear()` bug must be fixed before distributed cache.

### EC-006: JWT Validation is Structural Only
`verify_jwt()` checks 3-part dot structure only. No signature, issuer, or expiry validation. Documented as "swap-ready for real autom8y-auth integration."

---

## Risk Zones

### RISK-001: No Auth Signature Validation in Production
**Location**: `src/autom8_ads/dependencies.py` lines 145-167
**Severity**: HIGH. Any token matching `x.y.z` pattern passes. Known placeholder.

### RISK-002: Budget Reconciler Not Activated by Default
**Location**: `src/autom8_ads/dependencies.py` line 63
**Severity**: MEDIUM. Budget-too-low errors cause launch failures without auto-adjustment.

### RISK-003: `CampaignSearchService._try_import()` Silently Fails
**Severity**: LOW-MEDIUM. System degrades gracefully but import failures cause compounding search inefficiency.

### RISK-004: Idempotency `clear()` Affects All In-Flight Launches
**Location**: `src/autom8_ads/launch/service.py` lines 145-146, 162-163
**Severity**: MEDIUM. Under concurrent load, single failure clears all idempotency protection.

### RISK-005: Health Check Reports Dependencies as Healthy Based Only on None-Check
**Location**: `src/autom8_ads/api/health.py` lines 45-73
**Severity**: LOW. No connectivity probing; broken connections show as `"ok"`.

### RISK-006: `asyncio.to_thread()` in CampaignLock Has No Timeout
**Location**: `src/autom8_ads/lifecycle/campaign_lock.py` lines 66-78
**Severity**: LOW (currently not activated). Would be MEDIUM when CampaignSearchService is wired.

---

## Knowledge Gaps

1. `PlatformAdObject.raw` internal schema from real `autom8y_meta` SDK is unknown
2. `autom8y_config.Autom8yBaseSettings` base class behavior not visible in this codebase
3. `autom8y_telemetry.instrument_app()` security layer additions unknown
4. `autom8y_meta.MetaConfig` full contract not visible
5. Whether `data_writes_enabled` was intended to gate `CampaignSearchService` activation is unclear
