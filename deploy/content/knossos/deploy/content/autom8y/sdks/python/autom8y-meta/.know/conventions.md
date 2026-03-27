---
domain: conventions
generated_at: "2026-03-23T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4febf1f"
confidence: 0.93
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

**Language**: Python 3.12+ (async-only SDK)
**Package**: `autom8y-meta` — Async Meta Graph API client built on Pydantic v2, httpx, and the internal `autom8y-http`/`autom8y-log`/`autom8y-config` platform libraries.

## Error Handling Style

### Error Creation: Custom Class Hierarchy, Not Ad-hoc Exceptions

All errors in this project derive from a single base class `MetaError` defined in `src/autom8y_meta/errors.py`. The hierarchy is:

```
MetaError (base)
├── MetaAPIError         — code="META_API_ERROR",    http_status=502
├── MetaRateLimitError   — code="META_RATE_LIMIT",   http_status=429
├── MetaNotFoundError    — code="META_NOT_FOUND",    http_status=404
├── MetaBudgetError      — code="META_BUDGET_ERROR", http_status=422
└── MetaConfigError      — code="META_CONFIG_ERROR", http_status=500
```

Every exception class carries two `ClassVar` attributes: `code` (a string like `"META_NOT_FOUND"`) and `http_status` (int). This is described in the module docstring as "5+1 error hierarchy per TDD Section 8". `MetaError.__init__` accepts `message: str` plus `**context: Any`, storing context as a `dict` accessible via `to_dict()`.

Native Python `ValueError` and `RuntimeError` are used for programming contract violations (invalid arguments, uninitialized client) — these are NOT wrapped in `MetaError`.

### Error Wrapping: Cause Chaining on Boundary Transitions

The project uses Python's `raise X from e` cause-chaining exclusively at the client initialization boundary:

```python
# client.py line 91
raise MetaConfigError(f"Failed to load MetaConfig from environment: {e}") from e
```

Internal errors raised from `BaseHandler._classify_error` are constructed directly and `raise`d without cause-chaining. Transient/rate-limit errors are caught in the retry loop, logged, and re-raised after retry exhaustion.

### Error Propagation: Three-Phase Pattern

All API calls go through `BaseHandler._request` (`src/autom8y_meta/handlers/base.py`), which enforces this sequence:

1. **Attempt loop** (max 3 retries): catches `MetaRateLimitError` and transient `MetaAPIError`, calls `await self._retry.wait(attempt)`, continues
2. **Non-retryable errors**: raised immediately from `_handle_response` -> `_classify_error`
3. **Exhaustion**: after all retries, re-raises the `last_error`

The `_classify_error` method dispatches to the appropriate subclass based on HTTP status code and Meta's `error.code`/`error_subcode` fields. Classification logic is fully contained in `BaseHandler` — handler subclasses do not implement their own error dispatch.

### Error Categorization: Numeric Error Codes

`_classify_error` maps numeric signals to exception types:
- HTTP 429 or `code` in range 80000-80099 -> `MetaRateLimitError`
- `code=100, subcode=33` -> `MetaNotFoundError`
- `subcode in {2446149, 1885650}` -> `MetaBudgetError`
- `code in {190, 10}` -> `MetaConfigError`
- All other non-200 -> `MetaAPIError`

`MetaBudgetError` additionally parses the minimum budget value from the error message text using a module-level private regex `_BUDGET_PATTERN = re.compile(r"\$(\d+(?:\.\d+)?)")` and the private function `_parse_minimum_budget`. This private helper is imported directly by `BaseHandler` — the leading underscore signals it is internal API.

### Error Handling at Boundaries: Structured Logging, Not String Formatting

Structured logging uses `autom8y_log.get_logger(__name__)`. Log calls use keyword arguments for context, never f-string interpolation in log calls:

```python
logger.warning("api_rate_limited", path=path, attempt=attempt, retry_after=e.retry_after)
logger.warning("api_error", path=path, error_code=e.error_code, is_transient=e.is_transient, attempt=attempt)
```

There are no user-facing error messages. All error context surfaces via `to_dict()` for serialization to API responses or structured logs.

### Error Checking Convention: Guard + Raise for Invalid State

Client readiness uses a dedicated `_ensure_initialized` guard method that raises `RuntimeError` (not a `MetaError`) when the async context manager has not been entered:

```python
# client.py lines 161-171
def _ensure_initialized(self) -> None:
    if self._http is None or self._campaign_handler is None:
        raise RuntimeError("MetaAdsClient not initialized. Use 'async with MetaAdsClient() as client:'")
```

Every public method on `MetaAdsClient` calls `_ensure_initialized()` first, then uses `assert handler is not None` for mypy type narrowing (not for runtime checking).

## File Organization

### Package Root: `src/autom8y_meta/`

The `src/` layout is used (PEP 517 + `uv_build`). Source root: `src/autom8y_meta/`.

### Top-Level File Responsibility Map

| File | Contents |
|------|----------|
| `__init__.py` | Public API surface — imports and re-exports everything from submodules. Defines `__version__` via `importlib.metadata`. Contains usage example in module docstring. |
| `errors.py` | Complete error class hierarchy. One file for all exceptions. Also contains `_parse_minimum_budget` private helper. |
| `client.py` | `MetaAdsClient` facade — the only class consumers directly instantiate. Delegates all operations to handler instances. |
| `config.py` | `MetaConfig` (pydantic-settings) and `MetaAccountConfig` (plain BaseModel). Both in one file. |
| `auth.py` | `AppSecretProofGenerator` — single class, single concern. |
| `pagination.py` | `CursorPaginator` (async iterator) and `PageResult` (frozen dataclass). |
| `rate_limiter.py` | `MetaRateLimiter` — wraps `autom8y_http` token bucket + asyncio semaphore. |

### Subpackage: `handlers/`

One handler class per file, named after the resource (plural noun): `campaigns.py`, `ad_sets.py`, `ads.py`, `creatives.py`, `insights.py`, `lead_forms.py`, `pages.py`, `tokens.py`, `conversions.py`. Each inherits from `BaseHandler` in `base.py`. The `__init__.py` re-exports all handler classes with an explicit `__all__`.

Module-level functions that don't belong on a class are placed at file scope (e.g., `_parse_campaign_with_children` in `campaigns.py`). These are prefixed with `_` to indicate internal scope.

### Subpackage: `models/`

One model file per resource (singular noun): `campaign.py`, `ad_set.py`, `ad.py`, `creative.py`, `conversion.py`, `insights.py`, `lead_form.py`, `page.py`, `account.py`. Plus:
- `base.py` — `MetaModel` base class
- `enums.py` — all `StrEnum` types together (not one file per enum)

Each model file contains: response models, `WithChildren`/`WithAds` variants where needed, and `CreateParams`/`UpdateParams` classes.

### Import Pattern: TYPE_CHECKING Guard Consistently Applied

7 of 30 source files use `if TYPE_CHECKING:` for type-only imports. This is used when: (a) the import would be circular at runtime, or (b) the type is only needed for annotation strings (with `from __future__ import annotations` active).

All 27 non-`__init__` source files carry `from __future__ import annotations` as the first non-docstring import line.

### No Entry Point

There is no `__main__.py` or `app.py`. This is a library, not a service. Entry point is `MetaAdsClient` imported from the package root.

### Test Mirror Structure

Tests in `tests/` mirror the source structure: `tests/handlers/test_{resource}.py` and `tests/models/test_{resource}_models.py`. Root-level test files test top-level modules (`test_client.py`, `test_errors.py`, `test_config.py`, etc.). Shared fixtures live in `tests/conftest.py`.

## Domain-Specific Idioms

### 1. `to_api_params()` — Params Model Serialization Convention

All `CreateParams` and `UpdateParams` models implement a `to_api_params() -> dict[str, Any]` method. This method:
- Builds a `params: dict[str, Any] = {...}` with required fields
- Only includes optional fields when not `None` (explicit optional inclusion pattern)
- Converts budget integers to strings (Meta API requires string representation of cent values)
- Converts `StrEnum` values via `.value` for wire format

This is present in 6 files: `campaign.py`, `ad_set.py`, `ad.py`, `lead_form.py`, `creative.py`, `conversion.py`.

### 2. Response Models vs. Params Models — Different Base Classes

Response models (things returned from Meta API) inherit from `MetaModel`:
- `frozen=True, extra="ignore", from_attributes=True`

Params models (things sent to Meta API) inherit from `BaseModel` with different config:
- `frozen=True, extra="forbid"` — extra fields are forbidden, not ignored

This distinction is project-specific and not documented in Pydantic's defaults. `extra="forbid"` on params catches developer mistakes when constructing API calls.

### 3. `WithChildren`/`WithAds` Pattern — No Inheritance from Parent Models

When a model needs embedded children (e.g., `CampaignWithChildren`, `AdSetWithAds`), the project duplicates all parent fields rather than inheriting. The rationale is explicitly documented in docstrings:

> "All fields from `AdSet` are duplicated (not inherited) because `AdSet` is frozen and does not have an `ads` field. We avoid subclassing to prevent Pydantic frozen model mutation issues."

### 4. Handler Instantiation: Deferred to `__aenter__`

All handler instances and the `httpx.AsyncClient` are initialized to `None` in `__init__` and only assigned in `__aenter__`. This prevents resource leaks from constructing clients without context manager usage. The `_ensure_initialized` guard + `assert handler is not None` double-check pattern is consistent across all 40+ methods in `client.py`.

### 5. `act_` Prefix Normalization

Account IDs in Meta's API require the `act_` prefix, but users may pass IDs with or without it. The project normalizes this using `account_id.removeprefix("act_")` and then prepends `act_` in URL construction:

```python
acct = account_id.removeprefix("act_")
return await self._request("POST", f"/act_{acct}/campaigns", ...)
```

This appears in `CampaignHandler`, `AdSetHandler`, and `AdHandler`.

### 6. Nested Expansion Truncation Warning

When Meta's nested field expansion returns paginated inner results (more data exists than the `inner_limit`), the project logs a `nested_expansion_truncated` warning rather than silently dropping data or raising an error. This is annotated with `# FR-15:` comments referencing a requirements document.

### 7. StrEnum for All Enumerated API Values

All Meta API string constants that have a finite enumerated set use Python 3.11+ `StrEnum`. This allows `status == CampaignStatus.ACTIVE` and `status == "ACTIVE"` to both work, since `StrEnum` members are strings. The enums are all uppercase screaming-snake matching Meta's wire format.

## Naming Patterns

### Type and Class Naming

| Pattern | Examples | Rule |
|---------|----------|------|
| `Meta{Resource}` | `MetaAdsClient`, `MetaConfig`, `MetaRateLimiter`, `MetaModel`, `MetaError` | All project-level classes prefixed with `Meta` |
| `{Resource}Handler` | `CampaignHandler`, `AdSetHandler`, `InsightsHandler` | Internal handler classes (not exposed as public API) |
| `{Resource}CreateParams` | `CampaignCreateParams`, `AdSetCreateParams` | Request body models for creation |
| `{Resource}UpdateParams` | `CampaignUpdateParams`, `AdSetUpdateParams` | Request body models for updates |
| `{Resource}WithChildren` | `CampaignWithChildren`, `AdSetWithAds` | Expanded hierarchy models |

The `Meta` prefix is applied to all public-facing types. Handler classes do not have the `Meta` prefix because they are implementation details not exposed in `__all__`.

### File Naming

- Source files: `snake_case.py` with resource names matching their handler/model name (e.g., `ad_sets.py` for `AdSetHandler`, `campaigns.py` for `CampaignHandler`)
- Test files: `test_{module_name}.py` pattern strictly followed
- No `.pyi` stub files (the package has `py.typed` marker implied by `"Typing :: Typed"` classifier)

### Variable and Attribute Naming

- Private attributes: always prefixed with `_` (e.g., `self._http`, `self._rate_limiter`, `self._config`)
- Module-level private functions: leading underscore (e.g., `_parse_minimum_budget`, `_parse_campaign_with_children`)
- Type variables: single uppercase letter (`T`) as is conventional
- Async context managers: `limit()` (not `acquire()`, not `throttle()`)
- Fixtures in tests: descriptive noun phrases (`meta_config`, `meta_client`, `proof_generator`)

### Acronym Conventions

- `fb_trace_id` — Facebook trace ID uses `fb_` prefix (not `facebook_`)
- `cpc`, `cpm`, `ctr` — all lowercase for API metric field names
- `CAPI` used in comments for Conversions API
- `acct` used as local variable for normalized account ID (after stripping `act_` prefix)

### Package Naming Inconsistency

One notable inconsistency: the Python package is `autom8y_meta` (underscore) but the distribution name is `autom8y-meta` (hyphen). This follows Python packaging convention and is not a defect, but agents should use `autom8y_meta` in import statements and `autom8y-meta` in `pyproject.toml` / dependency specs.

## Knowledge Gaps

- The `autom8y_config.Autom8yBaseSettings` base class behavior is not observable from this repository — what additional behavior it adds to `pydantic_settings.BaseSettings` is undocumented here.
- The `autom8y_http.ExponentialBackoffRetry`, `RetryConfig`, `TokenBucketRateLimiter`, and `RateLimiterConfig` types are imported from workspace siblings — their contracts are assumed but not verified.
- The full set of Meta Graph API error codes beyond those classified in `BaseHandler._classify_error` is not captured.
- `handlers/insights.py` and `handlers/tokens.py` use `TYPE_CHECKING` imports for `MetaConfig` — suggests constructor signatures differ from the base `BaseHandler.__init__`; not fully read.
