---
domain: design-constraints
generated_at: "2026-03-16T00:14:42Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Platform Protocol Exists but Has No Concrete Implementation

**Location**: `services/ads/src/autom8_ads/platforms/protocol.py`

The `AdPlatform` protocol defines four async methods (`create_campaign`, `create_ad_set`, `create_creative`, `create_ad`) intended to abstract Meta, TikTok, Google, etc. However, **no concrete implementation exists in the codebase**. The docstring reads "MetaPlatformAdapter implements this for Meta. Future adapters for TikTok, Google, etc. would also implement this." but `MetaPlatformAdapter` is absent. All tests inject `AsyncMock` as the platform adapter. The application lifespan code (`app.py:76`) uses `getattr(app.state, "platform_adapter", None)` and only creates `LaunchService` when a platform adapter is already present — meaning the service cannot actually launch ads in production without an externally injected adapter.

**Type**: Missing abstraction / load-bearing gap
**Risk**: Service cannot function in production without an externally-supplied platform adapter that is not in this repository.

---

### TENSION-002: In-Memory Idempotency Cache Is Single-Instance-Only

**Location**: `services/ads/src/autom8_ads/launch/idempotency.py:36`

The `LaunchIdempotencyCache` is a plain Python dict. The class comment explicitly documents: "Single-instance deployment for Move 3. No distributed locking needed." This cache provides no distributed coordination. If the service is horizontally scaled, two concurrent requests for the same `offer_id` on different instances will both proceed, violating the idempotency guarantee.

**Type**: Structural tension — capability constrained to single-instance topology
**Trade-off**: Simplicity now vs. correctness at scale. Deferred to a future "Move."

---

### TENSION-003: Data Persistence Layer Is a Non-Functional Stub

**Location**: `services/ads/src/autom8_ads/clients/data.py`

`StubDataServiceClient` implements `DataServiceProtocol` but performs no actual persistence — it only logs at `WARNING` level. The module docstring states: "Real implementation deferred to Move 4." The config field `data_writes_enabled` defaults to `False`. Even when set to `True`, the stub client is wired in `app.py`. Any ad launch "recorded" to autom8_data during "Move 3" is silently dropped.

**Type**: Premature abstraction (protocol defined) + missing implementation
**Trade-off**: Protocol contract is locked in; implementation is intentionally absent.

---

### TENSION-004: Account Routing Config Is Hardcoded in Application Lifespan

**Location**: `services/ads/src/autom8_ads/app.py:28-41`

The `_create_default_router_config()` function hardcodes a single production account ID (`act_890095768862663`) and business ID (`596194394092925`). The `AdsConfig.account_routing_config` field (`config.py:38`) exists and defaults to `""` — implying future loading from a file or env — but `app.py` never reads this field. Account routing cannot be reconfigured without a code change.

**Type**: Under-engineering — config field exists but is ignored; hardcoded production credential in application code
**Risk**: Environment-specific IDs embedded in source. Config field is dead weight.

---

### TENSION-005: LaunchService Instantiation Is Conditional on Platform Adapter Presence

**Location**: `services/ads/src/autom8_ads/app.py:76-89`

`LaunchService` is only created during lifespan if `platform_adapter` is already in `app.state` at startup time (`getattr(app.state, "platform_adapter", None)`). This means the `POST /api/v1/launches` endpoint would raise `AttributeError` on `request.app.state.launch_service` in production — because `launch_service` is never set when `platform_adapter` is absent. The `get_launch_service` dependency (`dependencies.py:17`) uses a plain `cast()` with no guard.

**Type**: Structural gap — application can start and accept traffic with a broken state
**Risk**: Silent runtime failure, not a startup failure.

---

### TENSION-006: V2-Only Strategy Enforcement Is Split Across Two Layers

**Location**:
- `services/ads/src/autom8_ads/models/offer.py:96-99` (field validator)
- `services/ads/src/autom8_ads/lifecycle/strategies/base.py:14` (protocol docstring)
- `services/ads/src/autom8_ads/launch/service.py:115` (hardcoded instantiation)

ADR-ADS-002 mandates V2-only. This is enforced three ways: (1) `OfferPayload` rejects `algo_version != 2` at model validation, (2) `LaunchService.launch()` always instantiates `V2MetaLaunchStrategy()` unconditionally, (3) the `LaunchStrategy` protocol docstring says "V2MetaLaunchStrategy is the only implementation." The strategy pattern (Protocol + AdFactory) is therefore purely anticipatory complexity — there is no routing logic, no strategy selection, no version dispatch. `AdFactory` adds one layer of indirection for a single code path.

**Type**: Over-engineering — strategy pattern applied to a single concrete implementation
**Trade-off**: Architectural flexibility for anticipated future strategies vs. current complexity cost.

---

### TENSION-007: Targeting Is Simultaneously Typed and Opaque

**Location**: `services/ads/src/autom8_ads/models/targeting.py` and `services/ads/src/autom8_ads/models/offer.py:44`

`TargetingSpec` has typed fields (`geo_locations`, `radius_km`, `age_min`, etc.) but also a `raw: dict[str, Any] | None` passthrough. `OfferPayload.targeting` is typed as `dict[str, Any]` (fully opaque). `OfferPayloadMapper` constructs `TargetingSpec(raw=payload.targeting)` — passing the entire dict as `raw`, ignoring the typed fields entirely. The typed fields on `TargetingSpec` are never populated by the mapper; they exist but serve no purpose in the current pipeline.

**Type**: Premature abstraction — typed fields defined but bypassed in the only production code path
**Risk**: The typed fields create a false impression of validation. Targeting content passes through unvalidated.

---

### TENSION-008: Platform Naming Mismatch (Meta vs. Facebook)

**Location**: `services/ads/src/autom8_ads/urls/meta.py:36-37`

Two distinct base URLs are used:
- `ADSETS_BASE = "https://business.facebook.com/adsmanager/manage/adsets"` — uses `facebook.com` domain
- `ADS_BASE = "https://adsmanager.facebook.com/adsmanager/manage/ads"` — uses `adsmanager.facebook.com` subdomain

These are documented as "Matches legacy AdAccount.link property" and "Matches legacy ActiveAdsURL.facebook_ads_url()". The two different domains are a legacy artifact from different eras of Meta's Ads Manager URL structure, not a deliberate architectural choice. They must not be normalized to a single domain without validating against current Meta Ads Manager routing.

**Type**: Naming mismatch / legacy artifact
**Risk**: Normalizing either URL to match the other would break deep links.

---

### TENSION-009: `LaunchContext` Carries Asana-Derived Fields

**Location**: `services/ads/src/autom8_ads/models/launch.py:47-49`

`LaunchContext` has `task_gid: str = ""` and `vertical_key: str = ""` with empty-string defaults. The docstring explicitly notes these are "Name encoding source fields" sourced from Asana offer context. The service is designed to be "decoupled from Asana" (`launch.py:13`) but carries Asana-specific vocabulary (`task_gid` is an Asana GID) in its core domain model. Neither field is used in the launch pipeline itself — they are passed through for URL encoding only.

**Type**: Naming mismatch — Asana vocabulary in a service explicitly designed to be Asana-free
**Trade-off**: Functional decoupling achieved; conceptual naming coupling persists.

---

### TENSION-010: `status` and `error_type` Fields Are Unvalidated Strings

**Location**: `services/ads/src/autom8_ads/models/offer.py:115`, `129`

`LaunchResponse.status` and `error_type` are typed as `str` with no enum validator. The PRD/TDD specifies that `status` shall be one of `completed`, `failed`, `partial` and `error_type` shall be one of `validation`, `platform`, `transient`, `budget`. The QA adversarial tests (`test_contract_verification.py:158-168`, `210-217`) document this explicitly as intentional — "values are set by LaunchService/V2MetaLaunchStrategy, not by external callers." However, this creates a documentation-enforcement gap; the contract is PRD-level, not model-level.

**Type**: Under-engineering — semantic contract not enforced at the model boundary
**Documented finding**: Acknowledged in `tests/qa/test_contract_verification.py`.

---

## Trade-off Documentation

### Trade-off A: Stub Data Client over Real Implementation (Deferred to Move 4)

**Decision**: Use `StubDataServiceClient` with `data_writes_enabled=False` default.
**What was chosen**: Fire-and-forget data persistence stub that logs but does not write.
**What was rejected**: Building autom8_data client integration in the same sprint.
**Why tension persists**: "Move 3" is the current state. Move 4 (real data client) is unscheduled. The stub is designed to be swapped without interface changes.
**ADR reference**: Implied by `TDD-ADS-CORE` docstring in `service.py:3-6`.
**Load-bearing implication**: `data_writes_enabled=False` must remain the default until a real client is wired. Setting it to `True` with the stub produces silent data loss.

---

### Trade-off B: V2-Only Strategy with Strategy Pattern (ADR-ADS-002)

**Decision**: Enforce V2-only via field validator and hardcoded instantiation, but wrap in Strategy pattern.
**What was chosen**: V2MetaLaunchStrategy as the sole concrete strategy, Protocol + AdFactory scaffolding preserved.
**What was rejected**: Inline launch logic in `LaunchService`, or a strategy registry/dispatcher.
**Why tension persists**: The strategy pattern adds `AdFactory` → `LaunchStrategy` indirection for zero current benefit. The benefit is anticipated: adding V3 strategy or a TikTok strategy without changing `LaunchService`.
**ADR reference**: `ADR-ADS-002`, referenced in `models/offer.py:99`, `lifecycle/strategies/base.py:14`, `lifecycle/strategies/v2_meta.py:1-22`.

---

### Trade-off C: In-Memory Idempotency Cache Scoped to Single Instance (ADR-ADS-007)

**Decision**: Plain dict cache with TTL eviction, no distributed coordination.
**What was chosen**: Simplicity — in-memory dict is fast, zero dependencies, sufficient for Move 3 single-instance deployment.
**What was rejected**: Redis/Memcached distributed cache, database-backed idempotency.
**Why tension persists**: "Move 3" is single-instance. Horizontal scaling will require rearchitecting this component. The class comment marks the constraint explicitly.
**ADR reference**: `ADR-ADS-007`, referenced in `launch/idempotency.py:3`.

---

### Trade-off D: Field Inheritance Resolved Upstream (Decoupling Asana)

**Decision**: `autom8_ads` receives a fully pre-resolved `OfferPayload`. No parent task field resolution.
**What was chosen**: All field inheritance (office_phone from Business, vertical_key from Unit, weekly_ad_spend from Unit) is resolved by `autom8_asana` before calling this service.
**What was rejected**: Ad service fetching parent Asana task data itself.
**Why tension persists**: `LaunchContext` still carries `task_gid` and the `LaunchResponse` comment says "for Asana write-back" — functional decoupling achieved but conceptual coupling remains in naming.

---

### Trade-off E: Targeting Spec as Typed Wrapper over Opaque Dict

**Decision**: Define `TargetingSpec` with typed fields, but pass the whole dict as `raw`.
**What was chosen**: `TargetingSpec(raw=payload.targeting)` — typed struct exists but the typed fields are never populated.
**What was rejected**: Strict targeting schema validation with rejection of unknown fields.
**Why tension persists**: Meta's targeting API accepts an open-ended dict. Strict typing would require constantly updating the model as Meta's API evolves. The typed fields are aspirational documentation, not enforcement.

---

## Abstraction Gap Mapping

### Missing Abstractions

**MISSING-001: No Concrete `AdPlatform` Implementation**
- Location: `services/ads/src/autom8_ads/platforms/` (only `protocol.py` exists)
- Effect: The entire `platforms/` package has one file defining the interface with zero implementations. The production binary cannot make API calls without external injection.
- Duplicated logic at risk: None — but any attempt to add real Meta API calls requires creating `MetaPlatformAdapter` implementing `AdPlatform`.

**MISSING-002: No Account Routing Config File Loading**
- Location: `services/ads/src/autom8_ads/config.py:38`, `services/ads/src/autom8_ads/app.py:28-41`
- Effect: `account_routing_config` field exists but is never read. `_create_default_router_config()` is the only path. Account rules cannot be dynamically configured.

**MISSING-003: No Retry / Circuit-Breaker Around Platform API Calls**
- Location: `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py:36-107`
- Effect: `V2MetaLaunchStrategy` catches exceptions and returns partial results but has no retry logic. A transient Meta API error on `create_ad_set` produces a failed/partial launch with no automatic retry. Upstream caller must retry via the `DELETE /api/v1/launches/{offer_id}/{platform}` cache-clear endpoint.

### Premature Abstractions

**PREMATURE-001: `AdFactory` with Single Strategy**
- Location: `services/ads/src/autom8_ads/lifecycle/factory.py`
- Effect: `AdFactory` is a four-line class that calls `self._strategy.execute(self._platform, ctx)`. It adds an indirection layer with no current value. The factory pattern is correct if strategy selection or strategy chaining is planned, but today it is a wrapper for a single hardcoded strategy instantiation.

**PREMATURE-002: `TargetingSpec` Typed Fields**
- Location: `services/ads/src/autom8_ads/models/targeting.py:18-23`
- Effect: `geo_locations`, `radius_km`, `age_min`, `age_max`, `genders`, `languages` are defined but the `OfferPayloadMapper` bypasses them entirely by using `TargetingSpec(raw=payload.targeting)`. These fields are never read downstream.

**PREMATURE-003: `DataServiceProtocol` with Four Methods**
- Location: `services/ads/src/autom8_ads/platforms/protocol.py:54-71`
- Effect: `DataServiceProtocol` defines `record_campaign`, `record_ad_set`, `record_ad`, `record_creative`. Only `record_campaign` is ever called in the pipeline (`service.py:246`). The other three methods are defined in the protocol and stubbed but have no call sites.

---

## Load-Bearing Code Identification

### LOAD-001: `LaunchIdempotencyCache.get_or_set_in_progress()` — Atomic Check-and-Set

**Location**: `services/ads/src/autom8_ads/launch/idempotency.py:48-68`

This method is the idempotency gate. It atomically checks for an existing cache entry and, if absent, inserts an `in_progress` marker. The key format is `f"{offer_id}:{platform.value}"`. **Do not change the key format** without migrating all in-flight cache entries — any change creates a window where duplicate launches proceed. The eviction is called on every `get_or_set_in_progress()`, not on a timer; the implicit assumption is that the cache does not grow unboundedly between requests.

**Refactor risk**: Key format change breaks idempotency. Threading model assumes synchronous Python (GIL). Any move to async eviction or external cache requires a full rewrite.

---

### LOAD-002: `V2MetaLaunchStrategy.execute()` — Partial Result Semantics

**Location**: `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py:87-107`

When any step fails after campaign creation, the strategy catches the exception and returns `LaunchResult(success=False, status="partial", campaign_id=campaign_id, ...)`. This partial result propagates through `LaunchService._build_response()`, which uses it to build `ad_account_url` (if `campaign_id` is set) but not `live_ads_url` (only if `ad_id` is set). **Do not change partial result detection logic** (`status = "partial" if campaign_id else "failed"`) without updating `LaunchService._build_response()` and the downstream Asana write-back logic simultaneously.

**Refactor risk**: Changing `status` semantics breaks the URL building logic and any downstream consumer that branches on `status == "partial"`.

---

### LOAD-003: `MetaUrlBuilder` Filter Encoding — Proprietary `%1E`/`%1D` Separators

**Location**: `services/ads/src/autom8_ads/urls/meta.py:89`, `133`

The URL filter encoding uses `%1E` as field/operator/value separator and `%1D` as filter separator within the `filter_set` parameter. These are Meta's proprietary ASCII control character codes (record separator / unit separator). The column sets (`ADSETS_COLUMNS`, `ADS_COLUMNS`) are also fixed strings matching Meta's expected parameter names. **Do not change these values** without testing against a live Meta Ads Manager instance — they replicate legacy behavior from `AdAccount.link` and `ActiveAdsURL.facebook_ads_url()`.

**Refactor risk**: Changing separator characters breaks all generated deep links. The two base URLs (`business.facebook.com` vs. `adsmanager.facebook.com`) must remain distinct — they are not interchangeable.

---

### LOAD-004: `OfferPayload` Field Set — PRD-Locked Contract

**Location**: `services/ads/src/autom8_ads/models/offer.py:19-100`

`OfferPayload` has `extra="forbid"` and is tested against `PRD Section 6.1` in `tests/qa/test_contract_verification.py:71-100`. The field set is exactly as specified in the PRD. Adding or removing fields is a **breaking API contract change** requiring PRD update, downstream `autom8_asana` update, and test update. The `algo_version` field validator enforcing `v == 2` is load-bearing enforcement of ADR-ADS-002 — removing it would allow V1 payloads to reach a V2-only strategy.

**Refactor risk**: Any field addition requires `extra="forbid"` to be temporarily relaxed to prevent existing callers from breaking. The `algo_version` validator must not be removed without replacing ADR-ADS-002 enforcement elsewhere.

---

### LOAD-005: `LaunchService._try_persist()` — Platform Context Access Pattern

**Location**: `services/ads/src/autom8_ads/launch/service.py:247`

The persist method contains `ctx.platform if hasattr(ctx, "platform") else "meta"`. This `hasattr` check is defensive code against a type the system already knows has a `platform` attribute (`LaunchContext` always has it). The `hasattr` guard is a code smell but is harmless; **removing it** or changing it to a direct attribute access is safe, but any future use of `_try_persist` with a different context type would silently default to `"meta"`.

**Refactor risk**: Low — but signals a historical moment of uncertainty about the context type.

---

## Evolution Constraint Documentation

### Safe Areas (can be modified without coordination)

- `services/ads/src/autom8_ads/clients/data.py` — Stub implementation; replace with real client when Move 4 begins. Interface (`DataServiceProtocol`) is locked, implementation is not.
- `services/ads/src/autom8_ads/models/targeting.py` typed fields — Currently bypassed. Populating them requires only `OfferPayloadMapper` changes.
- `services/ads/src/autom8_ads/routing/config.py` — `AccountRule` and `AccountRouterConfig` models can be extended without breaking existing rules.
- `services/ads/src/autom8_ads/api/health.py` — Standalone health endpoint with no dependencies.
- `services/ads/src/autom8_ads/errors.py` — Error hierarchy can be extended by adding new subclasses. Existing codes are stable.

### Coordinated Areas (require multi-file changes)

- **Adding a new platform** (e.g., TikTok): Requires adding `Platform.TIKTOK` usage (enum already has `TIKTOK`), a new concrete `AdPlatform` implementation, potentially a new strategy, and `OfferPayload.validate_platform` to accept `"tiktok"`. Currently `"tiktok"` is rejected by the validator (`allowed = {"meta"}`).
- **Replacing in-memory cache with distributed cache**: Requires changes to `LaunchIdempotencyCache`, `app.py` lifespan wiring, and potentially `LaunchService` to handle distributed lock semantics differently.
- **Populating `TargetingSpec` typed fields**: Requires `OfferPayloadMapper.to_launch_context()` changes and downstream strategy consumption in `V2MetaLaunchStrategy`.
- **Wiring real `MetaPlatformAdapter`**: Requires creating `services/ads/src/autom8_ads/platforms/meta.py`, updating `app.py` lifespan to instantiate it, and adding Meta SDK to `pyproject.toml` dependencies.

### Migration Areas (deferred work with explicit markers)

- **`StubDataServiceClient` → real data client** (Move 4): All four protocol methods (`record_campaign`, `record_ad_set`, `record_ad`, `record_creative`) must be implemented. Only `record_campaign` has a call site currently.
- **Account routing from config file**: `AdsConfig.account_routing_config` field is a placeholder. `_create_default_router_config()` in `app.py` must be replaced with a file/env loader.

### Frozen Areas (do not modify without explicit approval)

- **`OfferPayload` field set**: PRD-locked. Requires PRD update + downstream service coordination.
- **Idempotency cache key format** (`f"{offer_id}:{platform.value}"`): Changes break in-flight request deduplication.
- **`MetaUrlBuilder` filter encoding** (`%1E`/`%1D` separators, column strings, base URLs): Matches legacy Meta deep-link format. Changes break all generated URLs.
- **`algo_version` validator** enforcing `v == 2`: ADR-ADS-002 enforcement. Cannot be removed until a V3 strategy is ready and a migration plan exists.

---

## Risk Zone Mapping

### RISK-001: Production Service Cannot Start Without External Platform Adapter Injection

**Location**: `services/ads/src/autom8_ads/app.py:76-89`
**Severity**: Critical
**Cross-reference**: TENSION-001, TENSION-005

The `LaunchService` is only constructed when `app.state.platform_adapter` is already set before lifespan executes. In production deployments, this would need to be set before `create_app()` is called, or lifespan fails silently to wire the service. The `get_launch_service` dependency will raise `AttributeError` on the first real launch request. **No startup health check verifies that `launch_service` is present in `app.state`.**

**Missing defense**: A startup validation that asserts `app.state.launch_service` exists after lifespan, or a meaningful error at the `get_launch_service` dependency layer.

---

### RISK-002: Zero-Budget Launch Reaches Platform API

**Location**: `services/ads/src/autom8_ads/launch/mapper.py:28-29`
**Severity**: Medium
**Cross-reference**: TENSION-007

When `weekly_ad_spend_cents=1` and `daily_budget_cents` is not provided, the mapper computes `1 // 7 = 0`. This zero daily budget is placed in `LaunchContext.daily_budget_cents` and passed to the platform API. Meta's API will reject a zero budget, but `autom8_ads` has no pre-flight guard. This is documented as a "FINDING" in `tests/qa/test_adversarial_payload.py:124-133`.

**Missing defense**: A `daily_budget_cents >= 1` assertion in `OfferPayloadMapper.to_launch_context()` or a `AdsBudgetError` guard.

---

### RISK-003: `offer_id` Accepts Empty Strings and Null Bytes as Cache Keys

**Location**: `services/ads/src/autom8_ads/models/offer.py:30`, `services/ads/src/autom8_ads/launch/idempotency.py:58`
**Severity**: Low
**Cross-reference**: TENSION-010

`offer_id` has no format constraints. Empty string, null bytes, and path traversal strings are accepted. The idempotency cache key becomes `":{platform.value}"` for empty `offer_id`, colliding across all offers with empty IDs. Documented in `tests/qa/test_adversarial_payload.py:36-65`.

**Missing defense**: Non-empty string validator on `OfferPayload.offer_id`.

---

### RISK-004: `data_writes_enabled=True` with Stub Client Silently Drops Data

**Location**: `services/ads/src/autom8_ads/app.py:70-71`, `services/ads/src/autom8_ads/config.py:41`
**Severity**: Medium
**Cross-reference**: TENSION-003

If `DATA_WRITES_ENABLED=true` is set in environment but the real data client is not wired (i.e., Move 4 hasn't shipped), the stub client will log `WARNING: stub_data_write` for every launch. This is a silent data loss scenario. No error is raised, no alarm is triggered beyond a log line.

**Missing defense**: A startup check that raises `AdsConfigError` if `data_writes_enabled=True` and `data_client` is an instance of `StubDataServiceClient`.

---

### RISK-005: `status` and `error_type` Contract Not Enforced at Model Boundary

**Location**: `services/ads/src/autom8_ads/models/offer.py:115`, `129`
**Severity**: Low
**Cross-reference**: TENSION-010

Any string is accepted for `status` and `error_type`. The constraint is only documented in PRD and tests, not enforced by the model. A bug in `LaunchService` or `V2MetaLaunchStrategy` producing an unexpected status value would propagate silently to downstream consumers.

**Missing defense**: `Literal["completed", "failed", "partial"]` type for `status`; `Literal["validation", "platform", "transient", "budget"]` for `error_type`.

---

### RISK-006: `asset_ids` Accepts Lists of Empty Strings

**Location**: `services/ads/src/autom8_ads/models/offer.py:74-78`
**Severity**: Low

The `validate_asset_ids_non_empty` validator ensures the list is non-empty but does not validate individual elements. `["", ""]` is a valid `asset_ids` value. These empty strings would be passed to the Meta API as asset references.

**Missing defense**: Item-level validator ensuring each `asset_id` is a non-empty string.

---

## Knowledge Gaps

- **ADR documents**: References to `ADR-ADS-002` and `ADR-ADS-007` in source code comments. No ADR files were found in `services/ads/`. The actual ADR content is unknown — only the decision outcomes are observable from code.
- **`MetaPlatformAdapter`**: Referenced in `platforms/protocol.py` docstring but not present anywhere in this repository. May exist in a separate service, a private package, or may not exist yet.
- **Terraform / IaC configuration**: The env vars (`META_BUSINESS_ID`, `ACCOUNT_ROUTING_CONFIG`, `DATA_WRITES_ENABLED`, etc.) are not validated here — their injection mechanism (Lambda secrets, SSM, etc.) is outside this codebase scope.
- **`Move 3` / `Move 4` milestone definitions**: Referenced in multiple docstrings but not defined in this codebase. These appear to be project-level development milestones.
- **Asana `task_gid` usage downstream**: `LaunchContext.task_gid` is populated but the `V2MetaLaunchStrategy` never reads it. If it is used for name encoding (implied by "Name encoding source fields" comment), that logic is either absent or in a non-existent `MetaPlatformAdapter`.
