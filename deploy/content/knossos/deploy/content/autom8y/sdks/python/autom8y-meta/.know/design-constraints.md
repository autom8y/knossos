---
domain: design-constraints
generated_at: "2026-03-23T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4febf1f"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

**Project**: autom8y-meta
**Language**: Python 3.12 (async-only, pydantic v2, httpx)
**Observation date**: 2026-03-23

## Tension Catalog Completeness

### TENSION-001: Frozen Model Duplication — Campaign/AdSet WithChildren Pattern

**Location**: `src/autom8y_meta/models/campaign.py` (lines 32-55), `src/autom8y_meta/models/ad_set.py` (lines 34-61)

`CampaignWithChildren` duplicates all 12 fields of `Campaign` verbatim. `AdSetWithAds` duplicates all 14 fields of `AdSet` verbatim. Neither inherits from the base type. The comment in `ad_set.py` (line 40-43) explicitly names the constraint:

> "All fields from AdSet are duplicated (not inherited) because AdSet is frozen and does not have an ads field. We avoid subclassing to prevent Pydantic frozen model mutation issues."

The structural tension: Pydantic v2 frozen models cannot be extended to add fields without workarounds. The chosen resolution is field duplication, which means that any addition to `Campaign` or `AdSet` requires a parallel update to their `WithChildren`/`WithAds` counterparts. The docstring acknowledges this but the risk of drift is live.

**Impact**: Two active maintenance synchronization points. A developer adding a field to `Campaign` will get no compiler warning if they forget to update `CampaignWithChildren`.

### TENSION-002: act_ Prefix Normalization Scattered Across Handlers

**Location**: 8 call sites across 4 handler files:
- `src/autom8y_meta/handlers/campaigns.py` (lines 33, 117, 243)
- `src/autom8y_meta/handlers/ad_sets.py` (line 22)
- `src/autom8y_meta/handlers/ads.py` (line 22)
- `src/autom8y_meta/handlers/creatives.py` (lines 19, 42, 58)

The Meta API requires `act_{id}` format for ad account endpoints. Each handler independently calls `account_id.removeprefix("act_")` to strip an existing prefix before re-adding it. There is no central normalization utility. If Meta changes the prefix convention or a new handler is added, each call site must be updated independently.

### TENSION-003: `object_type` Parameter Accepted but Not Routed for Insights

**Location**: `src/autom8y_meta/handlers/insights.py` (lines 39, 63), `src/autom8y_meta/client.py` (lines 419, 441)

`InsightsHandler.get_insights()` and `get_insights_async()` both accept an `object_type: str` parameter that is never used in the method body. The parameter is accepted in the public API surface, forwarded from `client.py`, but silently dropped. Any caller passing meaningful `object_type` values gets no differentiated behavior and no error.

**Impact**: Dead parameter on a public API; a caller cannot distinguish insight level by this means. The `level` parameter serves the actual routing purpose.

### TENSION-004: `_ensure_initialized()` + `assert` Dual-Pattern for Type Narrowing

**Location**: `src/autom8y_meta/client.py` — 33 `assert self._X is not None` statements throughout.

The client uses `_ensure_initialized()` as a runtime guard (raises `RuntimeError` if `__aenter__` was not called), then immediately follows with `assert self._handler is not None` on every method before using the handler. The `_ensure_initialized()` docstring explains this (line 164-167): the asserts exist only for mypy type narrowing, not for additional runtime safety.

This pattern works correctly but creates cognitive dissonance: a developer reading `assert self._campaign_handler is not None` might believe it provides protection, when in fact `_ensure_initialized()` already covers the failure case and the assert is a mypy workaround.

### TENSION-005: `search_objects` Calls `CampaignHandler._matches` Across Handler Boundaries

**Location**: `src/autom8y_meta/client.py` (lines 405, 410)

`MetaAdsClient.search_objects()` reaches into `CampaignHandler._matches` — a private static method on a different class — to perform name matching for adsets and ads. `_matches` is a pure utility function (case-insensitive substring match) but is coupled to `CampaignHandler` by location. If `CampaignHandler` is refactored, this cross-layer call becomes a hidden breakage point.

### TENSION-006: `_semaphore._value` Private Attribute Access

**Location**: `src/autom8y_meta/rate_limiter.py` (line 77)

`get_stats()` reads `asyncio.Semaphore._value` — a CPython implementation detail not part of the public `asyncio.Semaphore` API. The comment is absent. This works on CPython 3.12 but is not guaranteed across Python implementations or future CPython versions.

### TENSION-007: Insights Pagination in `InsightsHandler` vs `CursorPaginator` Abstraction

**Location**: `src/autom8y_meta/handlers/insights.py` (lines 109-116)

`get_insights_async()` manually implements cursor pagination (while loop on `paging.next`, cursor extraction) while the rest of the codebase uses `CursorPaginator`. This creates two pagination implementations. The insights pagination is functionally equivalent but bypasses the shared abstraction. If cursor key names change or pagination logic needs updating, two places require updates.

### TENSION-008: Token Extension Leaks Credentials in Query Parameters

**Location**: `src/autom8y_meta/handlers/tokens.py` (lines 44-49)

`extend_token()` sends `client_secret` as a query parameter via `params=`. The Meta OAuth token exchange endpoint requires this, but query parameters appear in server access logs and HTTP infrastructure logs. The `access_token` is also sent as a query parameter by `AppSecretProofGenerator.get_auth_params()` on every request. This is the Graph API's specified authentication mechanism and cannot be changed unilaterally, but it is a structural property that agents handling logging, tracing, or request inspection must account for.

## Trade-off Documentation

### TENSION-001: Why Field Duplication Persists

**Current state**: `CampaignWithChildren` (12 duplicated fields) and `AdSetWithAds` (14 duplicated fields) maintain all base model fields inline.

**Ideal state**: Single `Campaign` model extended by composition or optional fields.

**Why current persists**: Pydantic v2 frozen model inheritance has a known limitation — you cannot subclass a frozen model and add fields to produce a new frozen model without triggering Pydantic validation conflicts. The comment at `ad_set.py:40-43` documents this. The alternatives (making the base model non-frozen or using `model_copy(update=...)`) each trade type safety for ergonomics. The duplication was the least-surprising resolution.

**External constraint**: Pydantic v2 behavior. If Pydantic adds native "frozen extension" support, this can be resolved.

### TENSION-002: Why act_ Normalization Is Not Centralized

**Current state**: 8 scattered `removeprefix("act_")` call sites in handlers.

**Ideal state**: A single `normalize_account_id(account_id: str) -> str` utility.

**Why current persists**: No evidence of attempted centralization. The pattern is currently mechanical and low-risk (all handlers handle it the same way), but growth of new handlers increases the risk of inconsistency. No ADR found.

### TENSION-003: Why `object_type` Is Accepted but Unused

**Current state**: Parameter present on public interface, silently ignored.

**Ideal state**: Either remove the parameter or use it to differentiate endpoint behavior.

**Why current persists**: This appears to be a forward-compatibility stub from initial API design. Meta's insights endpoint URL is `/{object_id}/insights` regardless of object type — the type is implicit in the object ID. The `object_type` may have been added anticipating future differentiation. No ADR found.

### TENSION-006: Why `_semaphore._value` Is Used

**Current state**: Private attribute access for concurrency stats.

**Ideal state**: Use a separate counter maintained alongside the semaphore.

**Why current persists**: CPython's `asyncio.Semaphore` does not expose a public `available_slots` property. The `_value` attribute is stable in practice on CPython 3.12/3.13 and is the simplest path. A counter could be maintained independently but would require locking to stay consistent under concurrency.

### TENSION-007: Why Insights Pagination Bypasses CursorPaginator

**Current state**: Manual while loop in `get_insights_async()` at lines 109-116.

**Ideal state**: Wrap the final results pagination in `CursorPaginator`.

**Why current persists**: `get_insights_async()` is a two-phase operation (create job, poll, then fetch). The pagination occurs at the third phase. Wrapping it into `CursorPaginator` would require restructuring the polling phase, which increases complexity. The inline implementation was simpler for a method that returns `list[InsightsRow]` rather than a paginator. The trade-off is duplicated cursor-extraction logic.

## Abstraction Gap Mapping

### GAP-001: Missing Account ID Normalization Utility

**Type**: Missing abstraction (duplicated logic that should be extracted)

**Location**: Handlers: `campaigns.py` (3 sites), `ad_sets.py` (1), `ads.py` (1), `creatives.py` (3)

8 occurrences of `acct = account_id.removeprefix("act_")` followed by `f"/act_{acct}/..."`. A utility function `_act_path(account_id: str, resource: str) -> str` would be a natural extraction. Currently, adding a new handler that needs an account-scoped endpoint must independently implement this normalization.

### GAP-002: Missing Enum Coverage for AdSet/Ad Status Fields

**Type**: Incomplete abstraction (premature untyping)

**Location**: `src/autom8y_meta/models/ad_set.py` (line 19: `status: str | None`), `src/autom8y_meta/models/ad.py` (line 17: `status: str | None`)

`Campaign` uses `CampaignStatus` enum for its `status` field. `AdSet` and `Ad` use raw `str`. `AdSetStatus` and `AdStatus` enums exist in `src/autom8y_meta/models/enums.py` but are not applied to the response models. The `Params` models (`AdSetCreateParams`) also use `str` defaults (`status: str = "PAUSED"`). The enums exist but are not wired into the response models or all create params.

**Maintenance burden**: Callers must know valid string values rather than using the enum, and the response models lose validation.

### GAP-003: `_matches` Static Method Belongs to Utility Layer, Not CampaignHandler

**Type**: Abstraction in wrong location (orphaned utility function)

**Location**: `src/autom8y_meta/handlers/campaigns.py` (lines 259-264), called from `client.py` (lines 405, 410) as `CampaignHandler._matches`

A pure name-matching utility (`_matches(name, substrings, match_all)`) lives as a private static method on `CampaignHandler`. It is then called externally by `client.py` for matching on `AdSet` and `Ad` names. This is a module-level utility function that was placed inside `CampaignHandler` during initial implementation and was never extracted.

### GAP-004: `object_type` as Zombie Parameter

**Type**: Zombie abstraction (parameter that never materialized)

**Location**: `insights.py` lines 39, 63; `client.py` lines 419, 441

The `object_type` parameter on both insights methods accepts a value and does nothing with it. The public API surface implies differentiated behavior by type, but the implementation is uniform regardless of type.

### GAP-005: No Search Abstraction for AdSet/Ad

**Type**: Missing abstraction

**Location**: `src/autom8y_meta/client.py` (lines 400-411)

`search_objects()` handles campaigns via `CampaignHandler.search()` (a proper method) but handles adsets and ads with inline iteration and `CampaignHandler._matches` calls directly in `client.py`. The pattern is inconsistent: campaign search is delegated to a handler, while adset/ad search is implemented inline in the client.

## Load-Bearing Code Identification

### LOAD-001: `BaseHandler._request()` — Central Request Lifecycle

**Location**: `src/autom8y_meta/handlers/base.py` (lines 60-110)

All 9 handler subclasses call `self._request()` for every API call. This method owns: rate limit acquisition, auth parameter injection, retry logic (3 attempts, exponential backoff), response parsing, and error classification. Every API call in the SDK passes through this path.

**What a naive fix would break**: Changing retry semantics, parameter merging order, or rate limiter acquisition point affects all handlers simultaneously. Changing the `MetaRateLimitError` exception handling (line 84-94) would affect whether rate-limit errors trigger retry or bubble up.

**Load-bearing status**: Undocumented in code. The method docstring says "Execute an authenticated, rate-limited, retriable request" but does not flag it as the sole request pathway.

### LOAD-002: `AppSecretProofGenerator.get_auth_params()` — Authentication Bypass Gate

**Location**: `src/autom8y_meta/auth.py` (lines 43-54)

This method is called on every request via `BaseHandler._request()` (line 70). It contains a conditional bypass: if `app_secret` is empty, `appsecret_proof` is omitted from the params. This is the mechanism that allows apps without "Require App Secret" enabled to authenticate with token-only. Removing or changing this condition would break authentication for one of the two app configurations.

### LOAD-003: `_parse_campaign_with_children()` — Nested Expansion Parser

**Location**: `src/autom8y_meta/handlers/campaigns.py` (lines 267-333)

This free function parses Meta's nested field expansion response format. It handles the unusual Meta API convention where embedded edges return a `{"data": [...], "paging": {...}}` sub-dict rather than a flat list. It also implements the FR-15 truncation warning (lines 298-305, 315-323). Any change to the response key names or the dict reconstruction pattern affects the entire tree traversal feature.

### LOAD-004: `MetaRateLimiter.limit()` Context Manager — Concurrency Accounting

**Location**: `src/autom8y_meta/rate_limiter.py` (lines 57-73)

The `_requests_waiting` counter is incremented on entry and decremented in two places: on successful semaphore acquisition and in the `except BaseException` handler. The ordering between the semaphore acquisition and counter adjustment is load-bearing: the counter must decrement before `yield` in the success path and in the exception path before `raise`.

## Evolution Constraint Documentation

### Changeability Ratings

| Area | Changeability | Evidence |
|---|---|---|
| `BaseHandler._request()` | **coordinated** | All 9 handlers depend on it; retry, auth, and rate limit semantics are unified here |
| `AppSecretProofGenerator` | **coordinated** | Called on every request; dual-mode auth (proof vs. token-only) must remain intact |
| `MetaModel` base config | **coordinated** | 10+ response models inherit `frozen=True, extra="ignore"`; changing these affects all downstream consumers |
| `CursorPaginator` | **coordinated** | Used by campaigns, ad_sets, ads, and creatives handlers; interface changes require handler updates |
| `MetaConfig` field names + env prefix | **migration** | Callers use env var `META_*` names; any rename requires a migration window |
| `act_` prefix handling (8 sites) | **coordinated** | No central utility; all 8 sites must change together if the normalization logic changes |
| `_parse_campaign_with_children()` | **frozen** | Coupled to Meta API's specific nested expansion response shape; can only change when Meta changes the format |
| `InsightsRow` field types (`str | None`) | **safe** | Additive field changes allowed; removing or narrowing fields would break consumers |
| Handler classes (campaigns, ads, etc.) | **safe** | Can add methods; existing method signatures are part of public API through `client.py` |
| `CampaignWithChildren` / `AdSetWithAds` field lists | **coordinated** | Must stay in sync with `Campaign` / `AdSet` fields (TENSION-001) |
| `ObjectiveType` enum legacy values | **migration** | `LEAD_GENERATION` and `CONVERSIONS` are documented as "v24.0 still accepts these"; they exist for backward compatibility |

### Deprecated / In-Progress Markers

No `warnings.warn(DeprecationWarning)` calls found. No `# Deprecated:` comments found. The `ObjectiveType` enum comments note legacy values but do not formally deprecate them.

### External Dependency Constraints

- **Meta Graph API v24.0 pinned** (`MetaConfig.api_version = "v24.0"`): Upgrading the API version may change field names, endpoint URLs, or nested expansion syntax.
- **Pydantic v2**: The `frozen=True` duplication tension (TENSION-001) is directly caused by Pydantic v2 behavior.
- **autom8y-http primitives** (`ExponentialBackoffRetry`, `TokenBucketRateLimiter`, `RateLimiterConfig`): `BaseHandler` and `MetaRateLimiter` are built on these workspace packages.
- **Python 3.12+ minimum**: `StrEnum` and `str.removeprefix()` require Python 3.9+; `requires-python = ">=3.12"` is the declared constraint.

## Risk Zone Mapping

### RISK-001: `search_objects` — Incorrect Account ID Passed to Edge Endpoints

**Location**: `src/autom8y_meta/client.py` (lines 402-411)

When `object_type == "adset"`, `search_objects()` calls `self._ad_set_handler.get_campaign_ad_sets(account_id)` — passing an `account_id` where the method signature expects a `campaign_id`. The endpoint for `get_campaign_ad_sets` resolves to `/{campaign_id}/adsets`. If an account ID is passed, the Meta API will return an error or empty result. The input is not validated before the call.

**Cross-reference**: TENSION-005, GAP-005

### RISK-002: `InsightsHandler.get_insights_async()` — Unbounded Polling Without Jitter

**Location**: `src/autom8y_meta/handlers/insights.py` (lines 87-103)

The polling loop uses `asyncio.sleep(self._poll_interval)` with a fixed interval and no jitter. In a multi-tenant scenario with many simultaneous insight report jobs, all polling tasks would fire at the same wall-clock intervals, creating synchronized bursts against the Meta status endpoint.

### RISK-003: `_semaphore._value` — CPython Private API Dependency

**Location**: `src/autom8y_meta/rate_limiter.py` (line 77)

`asyncio.Semaphore._value` is not in the public Python asyncio API. A Python version upgrade could silently break `get_stats()`. No fallback or try/except guards this access.

**Cross-reference**: TENSION-006

### RISK-004: `access_token` in Query Parameters — Logging Risk

**Location**: `src/autom8y_meta/auth.py` (line 50), `src/autom8y_meta/handlers/tokens.py` (lines 47-49)

All requests send `access_token` as a query parameter. `extend_token()` additionally sends `client_secret`. Any HTTP request log, proxy log, or structured trace that logs the full request URL would capture these credentials. The SDK itself does not leak (logger logs `path` not full URLs), but infrastructure logging layers above the SDK could.

### RISK-005: `CampaignHandler._matches` Called from Client — Visibility Coupling

**Location**: `src/autom8y_meta/client.py` (lines 405, 410)

`CampaignHandler._matches` is a name-prefixed private method called cross-module. If `CampaignHandler` is renamed or the `_matches` method is made non-static, `search_objects()` will break silently at runtime (AttributeError) rather than at import time.

### RISK-006: Missing Validation for `insights_poll_interval` and `insights_max_wait` in Polling Loop

**Location**: `src/autom8y_meta/handlers/insights.py` (lines 75-103)

`max_wait` can be passed directly to `get_insights_async()` overriding the config value. If a caller passes `max_wait=0`, the while condition `elapsed < wait` is immediately false and the loop body never executes, raising `MetaError(... timed out after 0s)` without ever checking job status. The config-level `gt=0` validator does not apply to the method-level override.

## Knowledge Gaps

- **`handlers/ads.py`, `handlers/ad_sets.py`, `handlers/pages.py`, `handlers/lead_forms.py`**: Not fully read. Patterns are predictable from other handlers but edge-case behaviors are not individually documented.
- **`models/creative.py`, `models/page.py`, `models/lead_form.py`**: Not read. Field structure and to_api_params patterns not individually verified.
- **`autom8y-http`, `autom8y-config`, `autom8y-log` workspace packages**: Behavior relied upon by this SDK is documented by their own packages. Constraints imposed by those packages are not audited here.
- **Test suite not read**: Tests exist (`tests/` — 25 files confirmed) but were not examined. Test-visible constraints are not documented.
