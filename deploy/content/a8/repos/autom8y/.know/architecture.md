---
domain: architecture
generated_at: "2026-03-25T12:13:17Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "3fe30a4"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "3b84414216e3b39a03383be96b726206e9a9b54735258543352582ae8d4b0239"
---

# Codebase Architecture

**Language**: Python (primary). The repository is a Python monorepo managed with `uv` workspaces. Two JavaScript/TypeScript sites (`sites/docs`, `sites/landing`) exist but are outside the platform core. Terraform modules for infrastructure are also present.

**Workspace Root**: `/Users/tomtenuta/Code/a8/repos/autom8y/pyproject.toml`

## Package Structure

The repository is organized into three tiers:

### Tier 1: SDK Packages (`sdks/python/`)

20 independently-versioned Python packages published to a private AWS CodeArtifact index (`autom8y-python`). All follow `src/` layout with `autom8y_*` package names.

| SDK Package | Python Package | Purpose | Internal Dependencies |
|-------------|---------------|---------|----------------------|
| `autom8y-config` | `autom8y_config` | Pydantic Settings base class with SSM/SecretsManager secret resolution, `Autom8yBaseSettings`, `LambdaServiceSettingsMixin`, `SecretResolver` | None (foundation) |
| `autom8y-log` | `autom8y_log` | Structlog-based structured logging factory (`get_logger`, `configure_logging`, `LogConfig`, OTel trace ID injection) | `autom8y-config`, `autom8y-core` |
| `autom8y-core` | `autom8y_core` | Authenticated HTTP client with Bearer token injection (`Client`, `TokenManager`), `DataServiceClient` for the data plane, canonical transport errors | None (uses `httpx`, `pydantic`) |
| `autom8y-http` | `autom8y_http` | Batteries-included HTTP transport (`Autom8yHttpClient`) with rate limiting, retry/backoff, circuit breaker, concurrency controller, header redaction, OTel instrumentation | `autom8y-log` |
| `autom8y-auth` | `autom8y_auth` | JWT validation SDK: JWKS fetching, claims parsing, FastAPI dependencies/middleware, token manager, circuit breaker | `autom8y-core`, `autom8y-http` |
| `autom8y-telemetry` | `autom8y_telemetry` | OpenTelemetry initialization (`init_telemetry`, `TelemetryConfig`), domain decorators (`trace_gcal`, `trace_sms`, `trace_reconciliation`, etc.), FastAPI/Lambda instrument wrappers, typed semantic convention constants | `autom8y-config`, `autom8y-log` |
| `autom8y-cache` | `autom8y_cache` | Multi-backend caching (`InMemoryCacheProvider`, `RedisCacheProvider`, `S3CacheProvider`, `TieredCacheProvider`), versioned `CacheEntry`, completeness tracking, hierarchy tracking, batch operations | `autom8y-core` |
| `autom8y-events` | `autom8y_events` | Fire-and-forget domain event publishing to AWS EventBridge (`EventPublisher`, `DomainEvent`) | `autom8y-config` |
| `autom8y-interop` | `autom8y_interop` | Canonical typed inter-service clients: `data` (7 protocol-partitioned clients), `ads` (campaign tree), `asana` (offer query). Includes stubs and lifecycle factory functions for test-switching | `autom8y-core`, `autom8y-http` |
| `autom8y-meta` | `autom8y_meta` | Async Meta Graph API v24.0 client: campaigns, ad sets, ads, creatives, lead forms, CAPI, pagination, rate limiting | `autom8y-http`, `autom8y-config`, `autom8y-log` |
| `autom8y-stripe` | `autom8y_stripe` | Async Stripe client with rate limiting, payment categorization, typed handlers | `autom8y-http`, `autom8y-config` |
| `autom8y-slack` | `autom8y_slack` | Async Slack client with Block Kit formatting, CloudWatch alarm integration | `autom8y-config` |
| `autom8y-ai` | `autom8y_ai` | Protocol-first AI client SDK with Anthropic adapter (`adapters/anthropic.py`) | `autom8y-http`, `autom8y-log` |
| `autom8y-sendgrid` | `autom8y_sendgrid` | Async SendGrid SDK: typed email methods, suppressions, template management | `autom8y-http`, `autom8y-config` |
| `autom8y-gcal` | `autom8y_gcal` | Google Calendar SDK with DWD (domain-wide delegation) auth, FreeBusy chunking, Events CRUD, resource channels | `autom8y-http`, `autom8y-config` |
| `autom8y-calendly` | `autom8y_calendly` | Calendly API v2 client with webhook verification, payload normalization | `autom8y-http`, `autom8y-config` |
| `autom8y-google` | `autom8y_google` | Google Knowledge Graph enrichment (SerpAPI backend), fuzzy matching, two-level cache, circuit breaker | `autom8y-http`, `autom8y-config`, `autom8y-cache`, `autom8y-ai` |
| `autom8y-reconciliation` | `autom8y_reconciliation` | Composable reconciliation primitives: `ReadinessGate`, `Correlator`, `UnifiedVerdict`, `ReconciliationReportBuilder`, `ReconciliationMetrics` | `autom8y-config`, `autom8y-devx-types`, `autom8y-log`, `autom8y-telemetry` |
| `autom8y-saga` | `autom8y_saga` | Saga orchestration SDK: typed models (`SagaContext`, `SagaDefinition`), `SagaStepProtocol`, `CompensationRegistry`, Step Function outputs | None (external only) |
| `autom8y-devx-types` | `autom8_devx_types` | Zero-dependency canonical type definitions for dev-x plugin contracts (`_narrative.py`, `_span.py`, `_types.py`) | None (intentionally dependency-free) |

**File counts (approximate)**: Each SDK has 5-20 source files. `autom8y-cache` and `autom8y-interop` are the most complex (~25+ files each). `autom8y-devx-types` and `autom8y-events` are the smallest (~3-4 files each).

**Hub packages** (imported by many SDKs and services): `autom8y-config` (imported by 15/17 SDKs + all services), `autom8y-log` (imported by 12/17), `autom8y-http` (imported by 8/17).

**Leaf packages** (import no autom8y siblings): `autom8y-config`, `autom8y-core`, `autom8y-devx-types`, `autom8y-saga`, `autom8y-sms-test`.

### Tier 2: Service Packages (`services/`)

13 deployable service packages. Each has its own `pyproject.toml` and imports from the SDK tier. Services are classified by deployment archetype.

| Service | Python Package | Description | Archetype |
|---------|---------------|-------------|-----------|
| `auth` | `src/auth/` (flat, not src-layout) | JWT authentication, RBAC, API key management -- FastAPI long-running ECS service with RDS database | `ecs-fargate-rds` |
| `ads` | `autom8_ads` | Ad launch engine -- FastAPI ECS service; manages Meta campaign creation and routing | `ecs-fargate-stateless` |
| `account-status-recon` | `account_status_recon` | Three-way reconciliation (billing vs campaign vs contract) -- scheduled Lambda | `lambda-scheduled` |
| `pull-payments` | `pull_payments` | Stripe payment sync -- scheduled Lambda | `lambda-scheduled` |
| `reconcile-ads` | `reconcile_ads` | Asana vs Meta ads alignment reconciliation -- scheduled Lambda | `lambda-scheduled` |
| `reconcile-spend` | `reconcile_spend` | Spend vs collected anomaly detection -- scheduled Lambda | `lambda-scheduled` |
| `slack-alert` | `slack_alert` | SNS-triggered CloudWatch alarm Slack notifications -- event-driven Lambda | `lambda-event-driven` |
| `auth-mysql-sync` | `src/` (flat) | NHC MySQL -> Auth PostgreSQL projection sync -- scheduled Lambda | `lambda-scheduled` |
| `sms-performance-report` | `sms_performance_report` | Daily SMS concierge cost/booking metrics Slack summary -- scheduled Lambda | `lambda-scheduled` |
| `contente-onboarding` | `contente_onboarding` | Provisioning saga Lambda handlers (Step Functions state machine steps) | `lambda-event-driven` |
| `validate-business` | `validate_business` | Pre-saga business onboarding validation gate | `lambda-event-driven` |
| `calendly-intake` | `calendly_intake` | Calendly webhook intake pipeline with Google enrichment | `lambda-event-driven` |
| `devconsole` | `autom8_devconsole` | Real-time OTel span visualization developer tool | (local tool) |

**Satellite services** (separate repos, referenced but not contained here): `ads`, `data`, `asana`, `scheduling`, `sms` -- these have `satellite.repo` entries in `services.yaml`.

### Tier 3: Supporting Directories

- `services/_template/` -- Service scaffold template (excluded from workspace)
- `tools/ecosystem-observer/` -- Internal tooling (not a deployed service)
- `sites/docs/`, `sites/landing/` -- JavaScript/TypeScript web properties
- `observability/` -- Grafana/Prometheus/Loki/Tempo configuration files
- `terraform/` -- Infrastructure-as-code for all services

---

## Layer Boundaries

The codebase implements a clear dependency DAG with no circular imports across layers:

```
SERVICES (consumers)
    imports from: autom8y-interop, autom8y-config, autom8y-log, autom8y-telemetry,
                  autom8y-http, autom8y-core, autom8y-events, autom8y-reconciliation,
                  autom8y-saga, autom8y-stripe, autom8y-slack, autom8y-gcal, autom8y-meta

SDK DOMAIN CLIENTS (leaf clients)
    autom8y-interop: autom8y-core, autom8y-http
    autom8y-meta: autom8y-http, autom8y-config, autom8y-log
    autom8y-stripe: autom8y-http, autom8y-config
    autom8y-gcal: autom8y-http, autom8y-config
    autom8y-slack: autom8y-config
    autom8y-sendgrid: autom8y-http, autom8y-config
    autom8y-calendly: autom8y-http, autom8y-config
    autom8y-google: autom8y-http, autom8y-config, autom8y-cache, autom8y-ai
    autom8y-ai: autom8y-http, autom8y-log
    autom8y-saga: (external only -- no autom8y deps)

SDK CROSS-CUTTING (middleware tier)
    autom8y-telemetry: autom8y-config, autom8y-log
    autom8y-reconciliation: autom8y-config, autom8y-devx-types, autom8y-log, autom8y-telemetry
    autom8y-cache: autom8y-core
    autom8y-events: autom8y-config
    autom8y-auth: autom8y-core, autom8y-http

SDK TRANSPORT (mid-level)
    autom8y-http: autom8y-log
    autom8y-core: (external only: httpx, pydantic)
    autom8y-log: autom8y-config, autom8y-core

SDK FOUNDATION (no autom8y deps)
    autom8y-config: (external only: pydantic-settings, boto3)
    autom8y-devx-types: (zero deps)
    autom8y-saga: (external only)
    autom8y-sms-test: (pytest infrastructure only)
```

**Import direction rule**: Foundation -> Transport -> Cross-cutting -> Domain clients -> Services. Never reversed.

**Boundary enforcement mechanism**: uv workspace `constraint-dependencies` in root `pyproject.toml` enforces minimum version floors for every SDK. Services declare SDK deps explicitly in their own `pyproject.toml`. No `*` version ranges.

**Import enforcement**: Ruff TID251 (`[lint.flake8-tidy-imports.banned-api]`) in `ruff.toml` bans direct `httpx`, `structlog`, and `loguru` imports in services -- they must go through `autom8y_http.Autom8yHttpClient` and `autom8y_log.get_logger()`.

**Potential layer concern**: `autom8y-log` imports `autom8y-core` (a mid-level transport package). The dependency may be for `LoggerProtocol` type re-export. Worth watching as a possible layer inversion.

---

## Entry Points and API Surface

### ECS FastAPI Services

**auth** (`services/auth/src/main.py`)
- Entry: `app = FastAPI(lifespan=lifespan)` with lifespan, `uvicorn.run("src.main:app")` when `__main__`
- Routers mounted: `auth`, `rbac`, `api_keys`, `well_known`, `internal`, `admin`, `oauth_cc`
- HTTP API surface:
  - `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `POST /auth/refresh`
  - `GET /auth/me`, `POST /auth/password-reset`, `GET /auth/audit-logs`
  - `POST /api-keys`, `GET /api-keys`, `DELETE /api-keys/{id}`, `POST /api-keys/rotate`
  - `POST /roles`, `GET /roles`, `GET /roles/{id}`, `PUT /roles/{id}`, `DELETE /roles/{id}`
  - `POST /permissions`, `GET /permissions`, `DELETE /permissions/{id}`
  - `POST /rbac/assign`, `GET /.well-known/jwks.json`
  - `POST /internal/service-token`, `POST /internal/identifiers/resolve`
  - `POST /internal/revoke/user/{user_id}`, `POST /internal/revoke/token/{jti}`, `GET /internal/revoke/status/{jti}`
  - `GET /health` (liveness), `GET /ready` (readiness), `GET /health/deps` (dependency probe)
  - MOTHBALLED: credential vault routes (ADR-VAULT-001), charter routes (ADR-CHARTER-001), oauth routes

**ads** (`services/ads/src/autom8_ads/app.py`)
- Entry: `create_app()` factory; `lifespan` initializes singletons
- HTTP API surface: `POST /api/v1/launches`, `DELETE /api/v1/launches/{offer_id}/{platform}`, `GET /health`

### Lambda Services

| Service | Handler path | Trigger | Entry function |
|---------|-------------|---------|---------------|
| auth-mysql-sync | `services/auth-mysql-sync/src/main.py` | EventBridge scheduled (4h) | `lambda_handler(event, context)` |
| reconcile-ads | `services/reconcile-ads/src/reconcile_ads/handler.py` | EventBridge scheduled | `lambda_handler(event, context)` |
| reconcile-spend | `services/reconcile-spend/src/reconcile_spend/handler.py` | EventBridge scheduled | `lambda_handler(event, context)` |
| account-status-recon | `services/account-status-recon/src/account_status_recon/handler.py` | EventBridge scheduled | `lambda_handler(event, context)` |
| slack-alert | `services/slack-alert/src/slack_alert/handler.py` | SNS (CloudWatch alarm) | `handler(event, context)` |
| pull-payments | `services/pull-payments/src/pull_payments/handler.py` | EventBridge scheduled | `lambda_handler` |
| sms-performance-report | `services/sms-performance-report/src/sms_performance_report/handler.py` | EventBridge scheduled | `lambda_handler` |

**Lambda Saga Handlers** (`contente-onboarding`): Multiple handlers in one package, one per Step Functions state machine step. Files: `create_business.py`, `create_gcal.py`, `notify_sms.py`, `writeback_company_id.py`, `update_calendar_id.py`, `compensate_business.py`, `compensate_gcal.py`, `validate_business_saga.py`.

### SDK Exported Interfaces

Key exported types by SDK (see `__init__.py` for each):

- **autom8y-config**: `Autom8yBaseSettings`, `SecretResolver`, `Autom8yEnvironment`, `LambdaServiceSettingsMixin`
- **autom8y-log**: `get_logger`, `configure_logging`, `LogContextMiddleware`, `LoggerProtocol`
- **autom8y-http**: `Autom8yHttpClient`, `SyncHttpClient`, `HttpClientConfig`, `CircuitBreaker`, `ExponentialBackoffRetry`, `TokenBucketRateLimiter` + protocol interfaces
- **autom8y-core**: `Client`, `BaseClient`, `TokenManager`, `DataServiceClient`, `PhoneVerticalPair`, transport error hierarchy
- **autom8y-auth**: `AuthClient`, `AuthConfig`, `JWTAuthMiddleware`, `require_auth`, `BaseClaims`, `UserClaims`, `ServiceClaims`
- **autom8y-telemetry**: `init_telemetry`, `instrument_app`, `instrument_lambda`, `TelemetryContext`, `emit_business_metric`
- **autom8y-gcal**: `GCalClient`, `GCalConfig`, `GoogleCalendarAuthProvider`, `FreeBusyResponse`, `Event`, `WatchChannel`
- **autom8y-reconciliation**: `ReadinessGate`, `Correlator`, `UnifiedVerdict`, `ReconciliationReportBuilder`, `ReconciliationMetrics`
- **autom8y-events**: `EventPublisher`, `DomainEvent`
- **autom8y-interop**: sub-packages `data`, `asana`, `ads` with typed client stubs
- **autom8y-saga**: `SagaContext`, `SagaStepProtocol`, `CompensationRegistry`

### Deployment API (services.yaml)

`services.yaml` is the platform's single source of truth for service metadata. Schema keys: `archetype`, `deploy_method`, `satellite.repo`, `control.enabled`. CI entry: `.github/workflows/satellite-receiver.yml`. SDK CI: `.github/workflows/sdk-ci.yml`. SDK publish: `.github/workflows/sdk-publish-v2.yml` -> AWS CodeArtifact.

---

## Key Abstractions

### 1. `Autom8yBaseSettings` (`autom8y_config.base_settings`)

The universal configuration base class. Every service and SDK config class inherits it. Provides:
- Automatic resolution of `ssm://`, `secretsmanager://`, `env://` URI schemes in field values
- `to_safe_dict()` for log-safe serialization (redacts `SecretStr` fields)
- Guard rails preventing production URLs (`*.autom8y.io`) in local/test environments
- Environment normalization via `_ENV_ALIAS_MAP` (`dev` -> `local`, `prod` -> `production`)
- `lru_cache` factory pattern (`get_settings()`) used across all services for singleton config

### 2. `Autom8yHttpClient` (`autom8y_http.client`)

The standard HTTP transport primitive for all outbound calls. Combines:
- `TokenBucketRateLimiter`: Token bucket algorithm, thread-safe via asyncio.Lock
- `ExponentialBackoffRetry`: Configurable base/max delay with jitter and Retry-After header support
- `CircuitBreaker`: Three-state FSM (CLOSED -> OPEN -> HALF_OPEN)
- `ConcurrencyController`: Semaphore-based concurrent request limiting
- `raw()` escape hatch: Direct `httpx.AsyncClient` access for streaming/custom flows
- OTel trace context propagation via `InstrumentedTransport`
- `redact_headers()` for secure logging

### 3. Protocol-Stub Pattern (`autom8y_interop.data`)

The interop package defines capability protocols (e.g., `DataReadProtocol`, `DataInsightProtocol`) as Python `Protocol` classes. Each protocol has:
- A real HTTP client (`DataReadClient`) backed by `Autom8yHttpClient`
- A deterministic in-memory stub (`StubDataReadClient`) for testing
- A lifecycle factory (`resolve_read_client(use_stub=True, ...)`) that switches between them

### 4. `SagaContext` + `SagaStepProtocol` (`autom8y_saga`)

The saga orchestration contracts for Step Functions workflows. `SagaContext` carries state across steps. `SagaStepProtocol` defines the step interface. `CompensationRegistry` maps step types to rollback handlers. Used in `contente-onboarding` for multi-step business provisioning with forward/compensate semantics.

### 5. `@instrument_lambda` + Domain Decorators (`autom8y_telemetry`)

OTel instrumentation entry points. `@instrument_lambda` wraps a Lambda handler with trace initialization. Domain decorators (`@trace_sms`, `@trace_reconciliation`, `@trace_gcal`, `@trace_scheduling`) instrument specific business operations with typed span attributes drawn from the telemetry conventions subpackage.

The `autom8y_telemetry.conventions` subpackage is auto-generated from a YAML registry manifest. It exposes typed `Final[str]` attribute key constants (e.g., `RECONCILIATION_METRICS_EMITTED`, `PROVISIONING_SAGA_ID`). This ensures span attribute naming is codegen-enforced, not ad-hoc strings.

### 6. `CacheEntry` + `TieredCacheProvider` (`autom8y_cache`)

Versioned cache entry with metadata (`key`, `data`, `entry_type`, `version`). `TieredCacheProvider` coordinates a hot tier (Redis) and cold tier (S3). `CompletenessTracker` tracks whether a cached record is `PARTIAL` or `FULL`, enabling transparent fetch-on-miss upgrades. Primarily used in `autom8y-google` for enrichment caching.

### 7. `ReadinessGate` + `UnifiedVerdict` (`autom8y_reconciliation`)

Composable reconciliation primitives extracted from `reconcile-ads` and `reconcile-spend`. `ReadinessGate` validates data freshness and completeness before a reconciliation run. `Correlator` performs cross-source joins. `UnifiedVerdict` applies severity rules. `ReconciliationReportBuilder` formats Slack `mrkdwn` output.

### Design Patterns in Use

- **Protocol-first design**: `protocols.py` in nearly every SDK defines structural types
- **`from_env()` factory**: Found on `Client`, `TokenManager`, and SDK config classes
- **`lru_cache` settings singleton**: `@lru_cache` on `get_settings()` in every service
- **Lazy import guarding**: `contextlib.suppress(ImportError)` for optional FastAPI/AWS extras
- **`secretsmanager://` URI prefix**: DSL for secret references resolved by `SecretResolver`

---

## Data Flow

### 1. Service Configuration Flow

```
AWS SSM Parameter Store / Secrets Manager
    -> (URI: ssm:// or secretsmanager://)
Lambda Parameters & Secrets Extension (sidecar)
    -> populates environment variables
Autom8yBaseSettings.model_post_init / _resolve_secret_uris validator
    -> resolves URI-valued fields
Settings instance (lru_cache singleton)
    -> Service business logic
```

For Lambda services using `LambdaServiceSettingsMixin`, the `_resolve_extension_secrets` validator runs first (before `_resolve_secret_uris`) to fetch secrets from the Lambda extension's HTTP API.

### 2. Outbound HTTP Request Flow

```
Service calls domain client method (e.g., DataReadClient.get_business(...))
    -> Autom8yHttpClient: RateLimiter.acquire()
    -> CircuitBreaker: check state (CLOSED/HALF_OPEN pass, OPEN raises CircuitBreakerOpenError)
    -> ExponentialBackoffRetry: attempt 1..N
    -> InstrumentedTransport: inject W3C trace context headers (traceparent, tracestate)
    -> httpx.AsyncClient: execute HTTP request
    -> Response deserialization via Pydantic model
```

### 3. Lambda Reconciliation Pipeline Flow

```
EventBridge scheduler -> Lambda invocation
    -> @instrument_lambda: init_telemetry(), create root span
    -> asyncio.run(run_reconciliation())
    -> ReadinessGate.check(): validates data staleness thresholds
    -> Parallel data fetch via interop clients
    -> Correlator.join(): cross-source alignment
    -> UnifiedVerdict.evaluate(): severity classification
    -> ReconciliationReportBuilder.build(): format Slack mrkdwn
    -> autom8y_slack.client: post report to Slack channel
    -> EventPublisher.publish(): emit DomainEvent to EventBridge
    -> emit_metrics(): CloudWatch PutMetricData
```

### 4. Saga Provisioning Flow

```
Step Functions state machine -> Lambda invocation per step
    -> step handler (e.g., create_business.py, create_gcal.py)
    -> SagaContext deserialized from Step Functions input
    -> SagaStepProtocol.execute(context) -> StepResult
    -> (on success) Step Functions advances to next state
    -> (on failure) CompensationRegistry.compensate() -> rollback handler
```

### 5. SDK Publish Pipeline

```
Developer: just sdk-bump autom8y-{name} patch|minor|major
    -> Git push -> GitHub Actions sdk-publish-v2.yml
    -> uv build -> wheel + sdist
    -> AWS OIDC authentication to CodeArtifact
    -> uv publish -> CodeArtifact registry
    -> Workspace constraint-dependencies enforces minimum floor versions
```

### 6. Inter-Service Data Flow

```
Service A (caller)
    -> autom8y-core Client.from_env()
       -> reads SERVICE_API_KEY from env
       -> POST /internal/service-token -> JWT access token
    -> Service B HTTP endpoint
       -> autom8y-auth JWTAuthMiddleware validates token
       -> ServiceClaims injected into route handler
```

---

## Knowledge Gaps

1. **Satellite services** (`autom8y-data`, `autom8y-ads`, `autom8y-asana`, `autom8y-scheduling`, `autom8y-sms`): These 5 services live in separate repositories. Their internal architecture is not visible in this monorepo. `autom8y-interop` protocols define the contract boundary from this side.

2. **`auth-mysql-sync` service internal structure**: The `src/` layout is flat (not the modern `src/{pkg}` layout). The `portover/` and `sync/` subdirectory internals were not read in depth.

3. **`autom8y-cache` protocol hierarchy**: The `protocols/`, `backends/`, `completeness/`, `hierarchy/`, `metrics/` subdirectories contain substantial implementation not read in detail.

4. **`autom8y-log` backend adapters**: The `adapters/` and `backends/` subdirectories were not read. Structlog processor chain and backend selection were inferred from docstrings.

5. **`devconsole` service**: Functions as a local dev tool for OTel span visualization. Not deployed via `services.yaml`. Internal implementation not read.

6. **`autom8y-telemetry` convention code generation**: The `conventions/_data/` YAML files and `checker.py` / `registry.py` mechanism were not read in detail.

7. **Canary infrastructure**: The land notes mention "ECS canary infrastructure foundation" but no canary-specific Terraform or application code was observed in scope.
