---
domain: architecture
generated_at: "2026-03-23T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4febf1f"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The `autom8y-meta` package is a Python library (`src/autom8y_meta/`) organized into three sub-packages and five top-level modules. All source lives under `src/autom8y_meta/`.

**Language**: Python 3.12+, async-first (asyncio / httpx), Pydantic v2.

**Top-level module inventory** (5 modules):

| Module | File | Purpose |
|--------|------|---------|
| `__init__` | `src/autom8y_meta/__init__.py` | Public API surface; re-exports all user-facing symbols from sub-packages |
| `client` | `src/autom8y_meta/client.py` | `MetaAdsClient` — primary entry point; async context manager; delegates all operations to handlers |
| `config` | `src/autom8y_meta/config.py` | `MetaConfig` (pydantic-settings from `META_*` env vars) and `MetaAccountConfig` (per-account routing) |
| `auth` | `src/autom8y_meta/auth.py` | `AppSecretProofGenerator` — HMAC-SHA256 `app_secret_proof` generation and caching |
| `errors` | `src/autom8y_meta/errors.py` | 5+1 exception hierarchy (`MetaError` base + 5 typed subclasses) |
| `pagination` | `src/autom8y_meta/pagination.py` | `CursorPaginator[T]` (AsyncIterator) and `PageResult[T]` dataclass |
| `rate_limiter` | `src/autom8y_meta/rate_limiter.py` | `MetaRateLimiter` — TokenBucket + asyncio.Semaphore composition |

**`handlers/` sub-package** (10 files — hub package, imports from `models/`, `auth`, `errors`, `pagination`, `rate_limiter`):

| Handler | File | Operations |
|---------|------|-----------|
| `BaseHandler` | `src/autom8y_meta/handlers/base.py` | Auth param merging, rate limiting, retry loop, error classification |
| `CampaignHandler` | `src/autom8y_meta/handlers/campaigns.py` | Campaign CRUD, paginated list, nested tree traversal, search |
| `AdSetHandler` | `src/autom8y_meta/handlers/ad_sets.py` | Ad set CRUD, paginated list by campaign |
| `AdHandler` | `src/autom8y_meta/handlers/ads.py` | Ad CRUD, paginated list by ad set |
| `CreativeHandler` | `src/autom8y_meta/handlers/creatives.py` | Image/video upload, creative CRUD |
| `InsightsHandler` | `src/autom8y_meta/handlers/insights.py` | Sync insights fetch, async polling with report run lifecycle |
| `LeadFormHandler` | `src/autom8y_meta/handlers/lead_forms.py` | Lead form CRUD (page-scoped, not account-scoped) |
| `PageHandler` | `src/autom8y_meta/handlers/pages.py` | Instagram accounts, page assets, ad previews |
| `TokenHandler` | `src/autom8y_meta/handlers/tokens.py` | Token extension (short-lived -> long-lived), page access tokens |
| `ConversionHandler` | `src/autom8y_meta/handlers/conversions.py` | CAPI event dispatch to `/datasets/{id}/events` |

**`models/` sub-package** (10 files — leaf package, imports only from `models/base` and `models/enums`):

| Model file | Key types | Purpose |
|------------|-----------|---------|
| `src/autom8y_meta/models/base.py` | `MetaModel` | Pydantic base for all response DTOs (frozen, extra=ignore) |
| `src/autom8y_meta/models/enums.py` | `CampaignStatus`, `AdSetStatus`, `AdStatus`, `ObjectiveType`, `AssetType`, `PreviewFormat` | StrEnum constants for Meta API values |
| `src/autom8y_meta/models/campaign.py` | `Campaign`, `CampaignWithChildren`, `CampaignCreateParams`, `CampaignUpdateParams` | Campaign DTOs + mutation params |
| `src/autom8y_meta/models/ad_set.py` | `AdSet`, `AdSetWithAds`, `AdSetCreateParams`, `AdSetUpdateParams` | Ad set DTOs + mutation params |
| `src/autom8y_meta/models/ad.py` | `Ad`, `AdCreateParams`, `AdUpdateParams` | Ad DTOs + mutation params |
| `src/autom8y_meta/models/creative.py` | `Creative`, `CreativeSpec`, `AssetUploadResult` | Creative DTOs + upload result |
| `src/autom8y_meta/models/insights.py` | `InsightsRow`, `ReportRunStatus` | Insights row + async report run polling state |
| `src/autom8y_meta/models/conversion.py` | `ConversionEventPayload`, `ConversionResponse` | CAPI event + response |
| `src/autom8y_meta/models/lead_form.py` | `LeadForm`, `LeadFormCreateParams` | Lead form DTO + create params |
| `src/autom8y_meta/models/page.py` | `Page`, `InstagramAccount`, `PageAsset` | Page-related DTOs |
| `src/autom8y_meta/models/account.py` | `AdAccount` | Ad account DTO (internal, not re-exported via `__init__`) |

**Hub vs leaf classification**:
- **Hub** (many imports): `handlers/base.py` (imports `errors`, `auth`, `rate_limiter`, `httpx`); `client.py` (imports all 9 handlers + `config`, `auth`, `rate_limiter`, `errors`)
- **Intermediate** (import `models/`, `pagination`): All specific handler files
- **Leaf** (no internal imports): `models/base.py`, `models/enums.py`, `auth.py`, `errors.py`, `pagination.py`, `rate_limiter.py`

## Layer Boundaries

The package has four distinct layers. Import direction is strictly top-down — no upward imports observed in any file.

```
Layer 0 (entry): src/autom8y_meta/__init__.py
                 ↓ re-exports
Layer 1 (facade): src/autom8y_meta/client.py
                 ↓ imports 9 handler classes
Layer 2 (handlers): src/autom8y_meta/handlers/*.py
                 ↓ imports models, pagination, rate_limiter, auth, errors
Layer 3 (infrastructure): src/autom8y_meta/config.py
                           src/autom8y_meta/auth.py
                           src/autom8y_meta/rate_limiter.py
                           src/autom8y_meta/pagination.py
                           src/autom8y_meta/errors.py
                           src/autom8y_meta/models/*.py
```

**Observed import graph** (summary):

- `client.py` imports: all 9 handlers, `config`, `auth`, `rate_limiter`, `errors`
- `handlers/base.py` imports: `auth`, `rate_limiter`, `errors` (and `autom8y_http.ExponentialBackoffRetry`)
- `handlers/campaigns.py` imports: `handlers/base`, `models/campaign`, `models/ad_set`, `models/ad`, `pagination`
- `handlers/ad_sets.py` imports: `handlers/base`, `models/ad_set`, `pagination`
- `handlers/ads.py` imports: `handlers/base`, `models/ad`, `pagination`
- `handlers/insights.py` imports: `handlers/base`, `models/insights`, `config` (for poll settings)
- `handlers/tokens.py` imports: `handlers/base`, `config`
- `handlers/conversions.py` imports: `handlers/base`, `models/conversion`
- `handlers/creatives.py` imports: `handlers/base`, `models/creative`, `errors`
- `handlers/lead_forms.py` imports: `handlers/base`, `models/lead_form`
- `handlers/pages.py` imports: `handlers/base`, `models/enums`, `models/page`
- `models/campaign.py` imports: `models/base`, `models/enums`, `models/ad_set`
- `models/ad_set.py` imports: `models/base`, `models/ad`
- `models/ad.py` imports: `models/base`
- `models/creative.py` imports: `models/base`
- `models/page.py` imports: `models/base`
- `models/insights.py` imports: `models/base`
- `models/conversion.py` imports: `models/base`
- `models/account.py` imports: `models/base`

**Model dependency chain in `models/`**: `ad.py` <- `ad_set.py` <- `campaign.py` (one-directional; the tree traversal types build upward).

**Circular dependency avoidance**: `TYPE_CHECKING` guards are used in `client.py` (imports model types only for type hints) and `handlers/base.py` (imports `httpx`, `auth`, `rate_limiter` only for type hints). All runtime imports flow downward.

**External dependencies**: `autom8y_http` (provides `TokenBucketRateLimiter`, `RateLimiterConfig`, `ExponentialBackoffRetry`, `RetryConfig`); `autom8y_config` (provides `Autom8yBaseSettings`); `autom8y_log` (provides `get_logger`); `httpx` (async HTTP); `pydantic` / `pydantic-settings`.

## Entry Points and API Surface

**No CLI entry point.** This is a library package — there is no `__main__.py` and no console script defined in `pyproject.toml`. The user-facing API surface is entirely programmatic.

**Primary entry point**: `MetaAdsClient` from `src/autom8y_meta/client.py`.

**Initialization path**:
1. Caller constructs `MetaConfig` (from env vars or explicit args) and optional `list[MetaAccountConfig]`.
2. `MetaAdsClient(config, accounts=...)` stores config, builds `AppSecretProofGenerator`. All handlers are `None`.
3. `async with MetaAdsClient(...) as client:` triggers `__aenter__`: creates `httpx.AsyncClient`, `MetaRateLimiter`, and all 9 handler instances atomically. Returns `self`.
4. `__aexit__` closes the HTTP client and clears the proof cache.

**Public methods on `MetaAdsClient`** (complete list):

| Method | Return type | Description |
|--------|-------------|-------------|
| `get_account(account_id)` | `MetaAccountConfig` | Look up per-account config |
| `create_campaign(account_id, *, params)` | `Campaign` | Create campaign |
| `get_campaign(campaign_id, *, fields)` | `Campaign` | Fetch single campaign |
| `update_campaign(campaign_id, *, params)` | `Campaign` | Update campaign |
| `delete_ad_object(object_id)` | `None` | Delete any ad object |
| `create_ad_set(account_id, *, params)` | `AdSet` | Create ad set |
| `get_ad_set(ad_set_id, *, fields)` | `AdSet` | Fetch single ad set |
| `update_ad_set(ad_set_id, *, params)` | `AdSet` | Update ad set |
| `create_ad(account_id, *, params)` | `Ad` | Create ad |
| `get_ad(ad_id, *, fields)` | `Ad` | Fetch single ad |
| `update_ad(ad_id, *, params)` | `Ad` | Update ad |
| `get_account_campaigns(account_id, ...)` | `CursorPaginator[Campaign]` | Async iterator over campaigns |
| `get_campaign_ad_sets(campaign_id, ...)` | `CursorPaginator[AdSet]` | Async iterator over ad sets |
| `get_ad_set_ads(ad_set_id, ...)` | `CursorPaginator[Ad]` | Async iterator over ads |
| `get_account_campaigns_tree(account_id, ...)` | `CursorPaginator[CampaignWithChildren]` | Nested hierarchy fetch |
| `search_objects(account_id, object_type, name_substrings, ...)` | `list[Campaign \| AdSet \| Ad]` | Name substring search |
| `get_insights(object_id, object_type, ...)` | `list[InsightsRow]` | Sync insights (small query) |
| `get_insights_async(object_id, object_type, ...)` | `list[InsightsRow]` | Async polling insights (large query) |
| `send_conversion_event(dataset_id, event, ...)` | `ConversionResponse` | CAPI event |
| `upload_image(account_id, image_url)` | `AssetUploadResult` | Upload image via URL |
| `upload_video(account_id, video_url)` | `AssetUploadResult` | Upload video via URL |
| `create_creative(account_id, *, spec)` | `Creative` | Create ad creative |
| `get_lead_forms(page_id)` | `list[LeadForm]` | List lead forms for page |
| `get_lead_form(form_id)` | `LeadForm` | Fetch single lead form |
| `create_lead_form(page_id, *, payload)` | `LeadForm` | Create lead form |
| `archive_lead_form(form_id)` | `None` | Archive lead form |
| `get_page_instagram_accounts(page_id)` | `list[InstagramAccount]` | Instagram accounts for page |
| `get_page_assets(page_id, asset_type)` | `list[PageAsset]` | Page images/videos |
| `get_ad_preview(ad_id, *, ad_format)` | `str` | Ad preview as HTML iframe |
| `get_page_access_token(page_id)` | `str` | Get page access token |
| `extend_token(short_lived_token)` | `str` | Exchange for long-lived token |
| `get_rate_limiter_stats()` | `dict[str, Any]` | Rate limiter observability |

**Key exported interfaces** from `src/autom8y_meta/__init__.py`: All public types listed above are re-exported. `AdAccount`, `Page`, `models/lead_form`, and `models/account` module types are available via `models.__init__` but `AdAccount` and `Page` are not re-exported in the top-level `__all__`.

**Environment variables** (from `MetaConfig`, prefix `META_`):
- Required: `META_APP_ID`, `META_APP_SECRET`, `META_ACCESS_TOKEN`
- Optional: `META_API_VERSION` (default `v24.0`), `META_BASE_URL`, `META_TIMEOUT` (45s), `META_RATE_PER_MINUTE` (200), `META_MAX_CONCURRENCY` (50), `META_INSIGHTS_POLL_INTERVAL` (15s), `META_INSIGHTS_MAX_WAIT` (3600s)

## Key Abstractions

**1. `MetaAdsClient`** (`src/autom8y_meta/client.py`)
Facade over all 9 domain handlers. Enforces async-context-manager lifecycle. `_ensure_initialized()` guards all public methods and raises `RuntimeError` if called outside `async with`. The `assert` statements following `_ensure_initialized()` are mypy narrowing hints only — documented inline.

**2. `MetaModel`** (`src/autom8y_meta/models/base.py`)
Pydantic `BaseModel` subclass with `frozen=True`, `extra="ignore"`, `from_attributes=True`. All API response DTOs inherit from it. `extra="ignore"` allows Meta responses to contain undocumented fields without raising validation errors.

**3. `BaseHandler`** (`src/autom8y_meta/handlers/base.py`)
Shared HTTP lifecycle for all 9 handler classes. Central locus of: auth param injection (`_proof.get_auth_params()`), rate limiting (`async with self._rate_limiter.limit()`), retry loop (3 attempts, exponential backoff), error classification (`_classify_error`), and response parsing. All concrete handlers inherit from `BaseHandler` and call `self._request(method, path, ...)`.

**4. `CursorPaginator[T]`** (`src/autom8y_meta/pagination.py`)
`AsyncIterator[T]` wrapping Meta's cursor-based pagination. Constructed with a `fetch_page` callable and a `parse_item` callable — producers pass closures capturing the necessary URL and params. Also exposes `fetch_one_page(cursor)` for stateless single-page access (returns `PageResult[T]`). Used by `CampaignHandler`, `AdSetHandler`, `AdHandler`.

**5. `AppSecretProofGenerator`** (`src/autom8y_meta/auth.py`)
Computes `hmac_sha256(access_token, app_secret)` once per token and caches it. `get_auth_params()` returns `{"access_token": ..., "appsecret_proof": ...}` which is merged into every request's query params. Cleared on `__aexit__`.

**6. `MetaRateLimiter`** (`src/autom8y_meta/rate_limiter.py`)
Composes `autom8y_http.TokenBucketRateLimiter` (rate: N req/minute) with `asyncio.Semaphore` (concurrency: M parallel requests). Exposed via `async with limiter.limit()` context manager. Provides `get_stats()` for observability.

**7. Error hierarchy** (`src/autom8y_meta/errors.py`)
`MetaError` (base) -> 5 typed subclasses:
- `MetaAPIError` (HTTP 502, Meta error response with `error_code`, `is_transient`, `fb_trace_id`)
- `MetaRateLimitError` (HTTP 429, `retry_after`, `usage_header`)
- `MetaNotFoundError` (HTTP 404, code=100 subcode=33)
- `MetaBudgetError` (HTTP 422, subcodes 2446149/1885650, parses `minimum_budget_cents` from message)
- `MetaConfigError` (HTTP 500, codes 190/10 — expired/revoked tokens)

All carry `.code` (string), `.http_status` (int), and `.to_dict()` for structured logging.

**8. `*Params` / `to_api_params()` pattern** (all `*CreateParams`, `*UpdateParams` models)
Create/update param models use `BaseModel` with `extra="forbid"` (not `MetaModel`) and define `to_api_params() -> dict[str, Any]` to serialize to the exact field names the Meta Graph API expects. Budget values (int cents) are serialized as strings. None fields are excluded from the output dict.

**9. `CampaignWithChildren` / `AdSetWithAds`** (tree traversal models)
Flat duplication of `Campaign`/`AdSet` fields plus an embedded list of children (`adsets: list[AdSetWithAds]`, `ads: list[Ad]`). Explicitly documented not to inherit from the base frozen models to avoid Pydantic frozen mutation issues. Populated by `_parse_campaign_with_children()` in `handlers/campaigns.py`.

**Design patterns**:
- **Async context manager lifecycle**: Client resources (HTTP client, handlers) created on `__aenter__`, released on `__aexit__`. Pattern enforced across all methods via `_ensure_initialized()`.
- **Closure-based paginators**: `CursorPaginator` is instantiated with a `fetch_page` closure that captures the endpoint URL and params, keeping the paginator generic.
- **Read-after-write**: All create/update operations (campaign, ad set, ad, creative, lead form) follow a POST-then-GET pattern: POST returns only the new object ID, then GET fetches the full object. This avoids partial-response surprises from Meta.
- **`act_` prefix normalization**: All account-scoped handlers strip and re-apply the `act_` prefix via `account_id.removeprefix("act_")`. This lets callers pass IDs with or without the prefix.

## Data Flow

### Primary request path

```
Caller code
  -> MetaAdsClient.{method}(...)
    -> self._ensure_initialized()
    -> Handler.{operation}(...)
      -> BaseHandler._request(method, path, *, params, json_body, data)
        -> Merge auth params: {**params, **proof.get_auth_params()}
        -> for attempt in range(3):
            -> async with rate_limiter.limit():   # Semaphore + TokenBucket
                -> httpx.AsyncClient.request(method, path, params=..., json=..., data=...)
                -> _handle_response(response)
                  -> if 200: return response.json()
                  -> else: raise _classify_error(status_code, body, headers)
            -> on MetaRateLimitError: wait (retry.wait), continue
            -> on MetaAPIError (transient): wait, continue
            -> on MetaAPIError (non-transient): raise immediately
```

### Pagination path

```
Caller: async for item in client.get_account_campaigns(account_id):
  -> CampaignHandler.get_account_campaigns(account_id, ...) returns CursorPaginator
  -> CursorPaginator.__anext__()
    -> if buffer empty: _fetch_next_page()
      -> fetch_page(cursor)  [closure over account_id, fields, limit]
        -> BaseHandler._request("GET", f"/act_{acct}/campaigns", params=...)
      -> parse each item: Campaign.model_validate(item)
      -> extract paging.cursors.after; if paging.next exists: advance cursor, else: mark exhausted
    -> return buffer.popleft()
```

### Async insights polling path

```
client.get_insights_async(object_id, object_type, ...)
  -> InsightsHandler.get_insights_async(...)
    -> POST /{object_id}/insights -> report_run_id
    -> poll loop (every poll_interval seconds, up to max_wait):
        -> GET /{report_run_id}
        -> ReportRunStatus.model_validate(status)
        -> if "Job Completed": break
        -> if "Job Failed": raise MetaAPIError
        -> else: asyncio.sleep(poll_interval)
    -> GET /{report_run_id}/insights (+ follow cursor pagination)
    -> return list[InsightsRow]
```

### Configuration loading path

```
MetaConfig()  [triggered by MetaAdsClient.__init__ if no config passed]
  -> autom8y_config.Autom8yBaseSettings (pydantic-settings)
  -> reads environment vars with META_ prefix
  -> validates: app_id, app_secret, access_token (required)
  -> validates: api_version starts with "v"
  -> derives: graph_url = f"{base_url}/{api_version}"
```

### Error classification path (inside `BaseHandler._classify_error`)

```
HTTP status + response body + headers
  -> 429 or error_code in [80000-80099] -> MetaRateLimitError (parse Retry-After, x-business-use-case-usage)
  -> error_code=100, subcode=33          -> MetaNotFoundError
  -> subcode in {2446149, 1885650}       -> MetaBudgetError (parse minimum budget from message string)
  -> error_code in {190, 10}             -> MetaConfigError
  -> else                                -> MetaAPIError (with is_transient flag from Meta response)
```

## Knowledge Gaps

- `src/autom8y_meta/models/lead_form.py` was not read directly (file exists, handler `LeadFormHandler` imports `LeadForm` and `LeadFormCreateParams` from it — structure is consistent with all other model files).
- `autom8y_http`, `autom8y_config`, `autom8y_log` are workspace dependencies. Their internal APIs (`TokenBucketRateLimiter`, `RateLimiterConfig`, `ExponentialBackoffRetry`, `RetryConfig`, `Autom8yBaseSettings`, `get_logger`) are referenced but not observed in this audit scope.
- `AdAccount` is exported by `models/__init__` but not re-exported in the top-level `__init__.__all__`. Whether this is intentional (internal use only) or an oversight is not determinable from source alone.
