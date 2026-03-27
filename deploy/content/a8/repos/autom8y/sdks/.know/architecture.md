---
domain: architecture
generated_at: "2026-03-16T15:40:19Z"
expires_after: "7d"
source_scope:
  - "./python/**/*.py"
  - "./python/**/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The SDK monorepo lives at `/Users/tomtenuta/Code/a8/repos/autom8y/sdks/python/` and contains 18 independently-published packages. All packages target Python >=3.12, use `uv_build` (except `autom8y-devx-types` and `autom8y-sms-test` which use `hatchling`), and follow strict mypy + ruff configurations.

### Package Inventory

| Package | Python module | Version | Purpose | File count (src) |
|---|---|---|---|---|
| `autom8y-config` | `autom8y_config` | 1.2.1 | Configuration base class with secret resolution (SSM/Secrets Manager/env://) | 9 |
| `autom8y-log` | `autom8y_log` | 0.5.6 | Structured logging via structlog/loguru with protocol DI | 16 |
| `autom8y-core` | `autom8y_core` | 1.2.0 | Authenticated HTTP client with token management and data service client | 11 |
| `autom8y-http` | `autom8y_http` | 0.5.0 | HTTP transport primitives: rate limiter, retry, circuit breaker, resilience | 22 |
| `autom8y-telemetry` | `autom8y_telemetry` | 0.6.1 | OpenTelemetry tracing, domain decorators, convention registry, AWS/FastAPI integrations | 32 |
| `autom8y-auth` | `autom8y_auth` | 1.1.1 | JWT validation with JWKS caching, FastAPI middleware, server-side token verification | 21 |
| `autom8y-cache` | `autom8y_cache` | 0.4.0 | Tiered caching (Redis/S3/memory) with schema versioning and completeness tracking | 37 |
| `autom8y-events` | `autom8y_events` | 0.1.0 | Fire-and-forget domain event publishing to AWS EventBridge | 3 |
| `autom8y-gcal` | `autom8y_gcal` | 0.1.0 | Google Calendar API v3 SDK with DWD auth, FreeBusy, Events CRUD, Watch channels | 16 |
| `autom8y-meta` | `autom8y_meta` | 0.2.1 | Meta (Facebook) Graph API client with pagination and multiple handlers | 30 |
| `autom8y-sendgrid` | `autom8y_sendgrid` | 0.1.0 | SendGrid async client with templates, suppressions, stats | 13 |
| `autom8y-slack` | `autom8y_slack` | 0.2.0 | Async Slack client with Block Kit formatting | 8 |
| `autom8y-stripe` | `autom8y_stripe` | 1.3.1 | Async Stripe client with rate limiting and vertical categorization | 22 |
| `autom8y-reconciliation` | `autom8y_reconciliation` | 1.1.0 | Reconciliation pipeline primitives: gate, correlator, verdict, report, metrics | 15 |
| `autom8y-interop` | `autom8y_interop` | 1.2.0 | Inter-service typed clients for data service, Asana, and ads | 14 |
| `autom8y-ai` | `autom8y_ai` | 1.3.0 | Protocol-first AI client SDK with Anthropic adapter | 10 |
| `autom8y-devx-types` | `autom8_devx_types` | 1.1.0 | Zero-dependency canonical types for devx plugin contracts (note: underscore, not hyphen) | 5 |
| `autom8y-sms-test` | `autom8y_sms_test` | 0.1.0 | Unit-tier test infrastructure for Twilio SMS integration | 6 |

**Hub packages** (imported by many): `autom8y-config`, `autom8y-log`, `autom8y-http`, `autom8y-core`.

**Leaf packages** (import hub packages, used by services): `autom8y-gcal`, `autom8y-meta`, `autom8y-sendgrid`, `autom8y-slack`, `autom8y-stripe`.

**Notable naming anomaly**: `autom8y-devx-types` uses module name `autom8_devx_types` (single-underscore `autom8` prefix, not `autom8y`), as declared in its `pyproject.toml` `[tool.hatch.build.targets.wheel] packages = ["src/autom8_devx_types"]`.

### Internal Package Layout Convention

Every package follows `src/{module_name}/` layout with:
- `__init__.py` — public API surface with full `__all__`
- `protocols.py` — `@runtime_checkable Protocol` definitions
- `config.py` or `_settings.py` — Pydantic settings subclass
- `errors.py` — exception hierarchy
- `testing/` — pytest fixtures, factories, stubs (exposed via `pytest11` entry point)
- Optional `adapters/`, `backends/`, `handlers/`, `resilience/` subdirectories for implementations

## Layer Boundaries

The SDK dependency graph forms a clear four-tier model:

### Tier 0: Zero-dependency types
- **`autom8y-devx-types`** (`autom8_devx_types`) — zero runtime dependencies by design. Publishes `NarrativeRuleSet`, `ParsedSpan`, `Rule`, `Predicate`, `Renderer` used by the devx plugin ecosystem. All other packages that need these types import from here to guarantee `isinstance()` coherence.

### Tier 1: Foundation infrastructure (no internal cross-deps)
- **`autom8y-config`** — depends only on `pydantic` + `pydantic-settings`. The base settings class. All other packages depend on it.
- **`autom8y-log`** — depends on `autom8y-config`, `autom8y-core`, `structlog`. Logging factory.

**Important**: `autom8y-core` and `autom8y-http` are sibling tier-1 packages that depend on each other at the optional-extras level:
- `autom8y-core` has `autom8y-http` as a transitive dependency (via services that compose both)
- `autom8y-http` lists `autom8y-core` as an optional `[core]` extra, used by `ResilientCoreClient` in `autom8y_http.resilience.client`

### Tier 2: Transport layer
- **`autom8y-core`** — depends on `httpx` + `pydantic`. Provides `Client` (authenticated, token-injecting), `BaseClient` (abstract typed service base), `TokenManager`, `DataServiceClient`. Contains legacy aliases (`Autom8yClient = Client`) for backward compat.
- **`autom8y-http`** — depends on `autom8y-log` + `httpx`. Provides `Autom8yHttpClient` (opinionated async client with rate limiting/retry/circuit breaker), `SyncHttpClient`, all policy primitives.

### Tier 3: Observability and caching
- **`autom8y-telemetry`** — depends on `autom8y-config`, `autom8y-log`, `opentelemetry-api/sdk`. Provides `init_telemetry()`, domain-specific trace decorators (`trace_gcal`, `trace_sms`, `trace_reconciliation`, `trace_scheduling`), convention registry with `convention-check` CLI.
- **`autom8y-auth`** — depends on `autom8y-core`, `autom8y-http`, `PyJWT[crypto]`. Server-side JWT validation, JWKS caching, FastAPI middleware. Optional observability via `autom8y-log` + `autom8y-telemetry`.
- **`autom8y-cache`** — depends on `autom8y-core`. Provides tiered cache (Redis/S3/memory) with schema versioning, completeness tracking, hierarchy tracking.
- **`autom8y-events`** — depends on `autom8y-config` only (foundation-tier). EventBridge publisher; `boto3` is optional (present at Lambda runtime).

### Tier 4: Domain SDKs (external API wrappers)
These packages all depend on `autom8y-http` + `autom8y-config`, and use `Autom8yHttpClient` as their transport:
- **`autom8y-gcal`** — adds `google-auth` + `requests` for DWD authentication
- **`autom8y-meta`** — wraps Meta Graph API v21
- **`autom8y-sendgrid`** — wraps SendGrid v3 API
- **`autom8y-slack`** — wraps Slack Web API (uses raw `httpx`, not `autom8y-http`)
- **`autom8y-stripe`** — wraps `stripe` SDK v14 with autom8y rate limiting on top
- **`autom8y-interop`** — depends on `autom8y-core` + `autom8y-http`; typed clients for internal autom8y services (data, Asana, ads)

### Tier 5: Cross-cutting domain logic
- **`autom8y-reconciliation`** — depends on `autom8y-config`, `autom8y-devx-types`, `autom8y-log`, `autom8y-telemetry`. Reconciliation pipeline primitives extracted from service implementations.
- **`autom8y-ai`** — depends on `autom8y-http`, `autom8y-log`. Protocol-first AI client with optional Anthropic adapter.

### Test-only packages
- **`autom8y-sms-test`** — depends only on `httpx`. Unit-tier test infrastructure for Twilio SMS, installed via pytest plugin.

### Import direction rule
```
services → tier-4 domain SDKs → tier-2/3 foundation → tier-1 config/log → tier-0 types
```
Cross-tier imports always flow downward. No tier-1 package imports from tier-3 or tier-4.

## Entry Points and API Surface

### Public consumption patterns

**1. Authenticated HTTP to internal services (`autom8y-core`)**
```python
from autom8y_core import Client, DataServiceClient, PhoneVerticalPair
client = Client.from_env()                      # reads SERVICE_API_KEY, AUTOM8Y_AUTH_URL
response = client.get("https://data.api.autom8y.io/api/v1/health")
```
`Client.from_env()` is the canonical factory. `DataServiceClient.from_env()` wraps `Client` for typed data service calls.

**2. Resilient HTTP to external APIs (`autom8y-http`)**
```python
from autom8y_http import Autom8yHttpClient, HttpClientConfig
config = HttpClientConfig(base_url="https://api.example.com", timeout=30.0)
async with Autom8yHttpClient(config, logger=logger) as client:
    response = await client.get("/users/me")
    async with client.raw() as raw:           # escape hatch: bypass policies
        async with raw.stream("GET", "/large.csv") as resp: ...
```

**3. Configuration (`autom8y-config`)**
```python
from autom8y_config import Autom8yBaseSettings
class MySettings(Autom8yBaseSettings):
    database_url: str
    api_key: SecretStr
    model_config = SettingsConfigDict(env_prefix="MY_SERVICE_")
settings = MySettings()           # loads env vars, resolves ssm:// and secretsmanager:// URIs
settings.to_safe_dict()           # safe for logging (secrets redacted)
```

**4. Structured logging (`autom8y-log`)**
```python
from autom8y_log import configure_logging, get_logger, LogConfig
configure_logging(LogConfig(level="INFO", format="json"))
logger = get_logger(__name__)
logger.info("event_name", key="value")
logger.bind(request_id="abc").info("scoped_event")
```

**5. Telemetry (`autom8y-telemetry`)**
```python
from autom8y_telemetry import init_telemetry, TelemetryContext
init_telemetry("my-service", logger=logger)
# Domain-specific decorators:
from autom8y_telemetry import trace_gcal, trace_reconciliation, trace_sms, trace_scheduling
```

**6. JWT validation (`autom8y-auth`)**
```python
from autom8y_auth import AuthClient, AuthSettings, JWTAuthMiddleware, require_auth
# FastAPI: app.add_middleware(JWTAuthMiddleware, exclude_paths=["/health"])
# Dependency: claims: BaseClaims = Depends(require_auth)
```

**7. Google Calendar (`autom8y-gcal`)**
```python
from autom8y_gcal import GCalClient, GCalConfig
async with GCalClient(GCalConfig()) as client:
    result = await client.free_busy.query(calendar_ids=[...], time_min=..., time_max=...)
    events = await client.events.list(calendar_id=..., time_min=...)
```

**8. Reconciliation (`autom8y-reconciliation`)**
```python
from autom8y_reconciliation import ReadinessGate, Correlator, UnifiedVerdict, ReconciliationReportBuilder
```

### Pytest plugin entry points
14 of 18 packages register pytest plugins via `[project.entry-points."pytest11"]`. This makes test fixtures available to consumers without explicit imports:
- `autom8y_ai`, `autom8y_auth`, `autom8y_cache`, `autom8y_config`, `autom8y_core`, `autom8y_gcal`, `autom8y_http`, `autom8y_log`, `autom8y_reconciliation`, `autom8y_sendgrid`, `autom8y_slack`, `autom8y_sms_test`, `autom8y_stripe`, `autom8y_telemetry`

### CLI entry points
- `autom8y-telemetry` exports `convention-check` → `autom8y_telemetry.conventions.checker:main`

## Key Abstractions

### 1. `LoggerProtocol` (defined in `autom8y-core`)
Located at `python/autom8y-core/src/autom8y_core/protocols.py`.

`@runtime_checkable` Protocol with `debug/info/warning/error/exception(event, **kwargs)` and `bind(**kwargs) -> Self`. Every SDK that accepts a logger uses `LoggerProtocol` from `autom8y_core` for DI, avoiding a hard dependency on `autom8y-log`. This is the cross-cutting DI pattern for observability.

### 2. `HttpClientProtocol` + `Autom8yHttpClient` (in `autom8y-http`)
`HttpClientProtocol` at `python/autom8y-http/src/autom8y_http/protocols.py` defines the async contract: `get/post/put/delete` + `raw()` escape hatch. `Autom8yHttpClient` is the concrete implementation composing:
- `TokenBucketRateLimiter` (token bucket algorithm)
- `ExponentialBackoffRetry` (with Retry-After header support)
- `CircuitBreaker` (CLOSED/OPEN/HALF_OPEN state machine)
- `ConcurrencyController` (semaphore-based)

### 3. `Autom8yBaseSettings` (in `autom8y-config`)
At `python/autom8y-config/src/autom8y_config/base_settings.py`. Extends pydantic `BaseSettings` with URI-based secret resolution (`ssm://`, `secretsmanager://`, `env://`), `to_safe_dict()` redaction, and `Autom8yEnvironment` guard that prevents production URLs in local/test environments.

### 4. `CacheProvider` protocol (in `autom8y-cache`)
The `CacheProvider` protocol at `python/autom8y-cache/src/autom8y_cache/protocols/cache.py` defines versioned cache operations. Concrete implementations: `InMemoryCacheProvider`, `RedisCacheProvider`, `S3CacheProvider`, `NullCacheProvider`, `TieredCacheProvider`. Factory `create_cache_provider()` selects backend from environment. Schema versioning uses `SchemaVersion` and `SchemaVersionProvider` protocol for cache migration without data loss.

### 5. `BaseClient` / `Client` pattern (in `autom8y-core`)
`Client` at `python/autom8y-core/src/autom8y_core/client.py` is a sync+async authenticated HTTP client (Bearer token injection via `TokenManager`). `BaseClient` at `python/autom8y-core/src/autom8y_core/base_client.py` is the abstract base for typed service clients: subclasses implement typed methods using `self._http` (a `Client` instance).

### 6. Domain trace decorators (in `autom8y-telemetry`)
`trace_gcal`, `trace_sms`, `trace_reconciliation`, `trace_scheduling`, `trace_chat`, `trace_decision` — decorator functions in `autom8y_telemetry` that wrap async functions with OpenTelemetry spans and populate domain-specific semantic convention attributes.

### 7. Protocol-first design pattern
All domain SDKs expose a `*Protocol` type (e.g., `GCalClientProtocol`, `HttpClientProtocol`, `TelemetryContextProtocol`). These are `@runtime_checkable` to support both type-safe DI and `isinstance()` guards. The pattern: Protocol → Concrete implementation → Testing stub/mock.

### 8. devx plugin system (`autom8y-devx-types` + `autom8y-telemetry`)
`autom8_devx_types` defines `NarrativeRuleSet`, `NarrativeFragment`, `Rule`, `Predicate`, `Renderer` with zero runtime dependencies. `autom8y-telemetry` discovers plugins via `[project.entry-points."autom8y_devx.narrative_rules"]` (implemented in `autom8y-reconciliation`: `reconciliation = "autom8y_reconciliation.narrative.rules:get_rules"`). The convention registry in `autom8y_telemetry.conventions` provides a YAML-driven span attribute registry with `convention-check` CLI.

## Data Flow

### Configuration flow
```
Environment vars / AWS SSM / Secrets Manager
    → autom8y_config.SecretResolver (resolves ssm://, secretsmanager://, env:// URIs)
    → Autom8yBaseSettings subclass (service-specific settings class)
    → SDK constructors (e.g., GCalConfig(), StripeConfig(), HttpClientConfig())
```
The `AUTOM8Y_ENV` env var drives the environment guard; `SERVICE_API_KEY` is the canonical inter-service auth key; `AUTOM8Y_DATA_URL` is the canonical data service URL.

### Authenticated request flow (internal services via `autom8y-core`)
```
Client.from_env()
    → TokenManager (reads SERVICE_API_KEY, AUTOM8Y_AUTH_URL)
    → POST /token → JWT Bearer token (cached, auto-refreshed)
    → requests carry Authorization: Bearer <token>
    → DataServiceClient: typed methods → JSON → Pydantic models
```

### External API request flow (via `autom8y-http`)
```
Autom8yHttpClient(config, logger)
    → RateLimiter.acquire() (token bucket)
    → ConcurrencyController.acquire() (semaphore)
    → CircuitBreaker.check() (fail-fast if OPEN)
    → httpx.AsyncClient.request() (with OTel trace context headers)
    → CircuitBreaker.record_success/failure()
    → Pydantic model parsing at domain SDK layer
```

### Telemetry flow
```
init_telemetry(service_name)
    → TracerProvider configured with BatchSpanProcessor
    → OTLP exporter (gRPC or HTTP, configured via OTEL_EXPORTER_OTLP_ENDPOINT)
    → domain decorators (trace_gcal, trace_sms, etc.) wrap async functions
    → W3C Trace Context propagated via TelemetryContext.inject(headers) on outbound calls
    → TelemetryContext.extract(headers) on inbound calls
    → trace_id/span_id injected into structlog via add_otel_trace_ids processor
```

### Caching flow
```
create_cache_provider()           # reads AUTOM8Y_CACHE_ENABLED, REDIS_URL, S3_BUCKET
    → TieredCacheProvider (hot: Redis, cold: S3)
    → CacheEntry (versioned, typed by EntryType)
    → CompletenessLevel tracking (PARTIAL vs FULL)
    → SchemaVersion validation on get/set (migration via SchemaVersionProvider)
```

### Event publishing flow (Lambda context)
```
EventPublisher(bus_name="autom8y-events")
    → boto3 EventBridge client (runtime-provided in Lambda, optional dep)
    → DomainEvent(source=..., detail_type=..., detail={...})
    → fire-and-forget (no ack, no retry)
```

## Knowledge Gaps

1. **No workspace-level `pyproject.toml`**: There is no `pyproject.toml` at `python/` or above — each package is managed independently. The `dist/` directory contains pre-built wheels. Workspace tool configuration (ruff, mypy) is at package level; a root `ruff.toml` is referenced by several packages via `extend = "../../../ruff.toml"` but was not read.

2. **`autom8y-auth` full export list not read**: The `__init__.py` was read partially (50 lines). Full export surface of `AuthClient`, `TokenManager`, claims hierarchy was not fully inventoried.

3. **`autom8y-meta` handlers** — 11 handler files for ad sets, ads, campaigns, conversions, creatives, insights, lead forms, pages, tokens were not individually read. The overall pattern is consistent with other domain SDKs but granular handler method signatures are not documented here.

4. **`autom8y-reconciliation` narrative plugin** — The `autom8y_devx.narrative_rules` entry point (`reconciliation = "autom8y_reconciliation.narrative.rules:get_rules"`) and how `NarrativeRuleSet`/`get_rules()` compose was not fully read.

5. **Root `ruff.toml`** — referenced by packages via `extend = "../../../ruff.toml"`. Location and shared lint rules not confirmed.
