---
domain: architecture
generated_at: "2026-03-16T20:00:00Z"
expires_after: "7d"
source_scope:
  - "./*/src/**/*.py"
  - "./*/tests/**/*.py"
  - "./*/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The codebase is a Python multi-service monorepo under `services/`. Each service is an independent Python package with its own `pyproject.toml`, `src/`, and `tests/` directories. All services target Python 3.12 (except `ads` which targets 3.11) and use `hatchling` as the build backend.

### Service Inventory

| Service Directory | Package Name | Purpose | Runtime Type | File Count (src) |
|---|---|---|---|---|
| `account-status-recon/` | `account_status_recon` | Three-way reconciliation across billing, campaign, and contract data; posts Slack report to `#account-health` | AWS Lambda (scheduled) | 13 |
| `ads/` | `autom8_ads` | Ad launch engine — accepts `OfferPayload` from autom8_asana and creates Meta ad campaigns | FastAPI HTTP service | 30+ |
| `auth/` | `auth` (src layout, no package name) | JWT authentication and authorization microservice; issues RS256 tokens, manages users, RBAC, API keys | FastAPI HTTP service | 40+ |
| `auth-mysql-sync/` | (src layout) | Sync projection of NHC MySQL entities (employees, chiropractors) to Auth PostgreSQL | AWS Lambda (scheduled, 4h) |  15 |
| `devconsole/` | `autom8_devconsole` | Developer console for real-time OTel span visualization; receives OTLP traces, renders NiceGUI UI | FastAPI + NiceGUI (localhost dev tool) | 20 |
| `pull-payments/` | `pull_payments` | Stripe payment sync service; pulls invoices and writes to data service | AWS Lambda (scheduled) | 11 |
| `reconcile-ads/` | `reconcile_ads` | Ads cross-check reconciliation — Asana vs Meta alignment; posts Slack report | AWS Lambda (scheduled) | 13 |
| `reconcile-spend/` | `reconcile_spend` | Spend vs collected anomaly detection; posts Slack report to configured channel | AWS Lambda (scheduled) | 17 |
| `slack-alert/` | `slack_alert` | SNS-triggered Lambda that posts CloudWatch alarm notifications to Slack | AWS Lambda (event-driven via SNS) | 2 |
| `sms-performance-report/` | `sms_performance_report` | Daily SMS concierge performance report fetched from data service; posts to Slack | AWS Lambda (scheduled) | 9 |
| `_template/` | (scaffold) | Service template for new satellite services | N/A | 0 |

### Internal Module Structure Per Service

**account-status-recon** (`services/account-status-recon/src/account_status_recon/`):
- `__main__.py` — local CLI invocation with `--csv`/`--json` output modes
- `handler.py` — AWS Lambda entry point (`lambda_handler`), calls `run_reconciliation()`
- `orchestrator.py` — 7-step reconciliation pipeline (fetch → readiness → join → rules → report → metrics → event)
- `fetcher.py` — parallel async fetch from three data sources (billing, campaigns, offers)
- `joiner.py` — three-way join on `(office_phone, vertical)`
- `rules.py` — verdict rules across 5 axes (billing, budget, campaign, three-way, contract)
- `readiness.py` — data freshness gate (PASS/WARN/FAIL)
- `models.py` — domain types: `SourcePresence`, `BillingData`, `CampaignData`, `ContractData`, `AccountRecord`, `AccountFinding`, `ReconciliationResult`
- `report.py` — Slack Block Kit report builder
- `metrics.py` — CloudWatch metric emission
- `config.py` — `Settings(LambdaServiceSettingsMixin, Autom8yBaseSettings)` with `@lru_cache get_settings()`
- `errors.py` — service-specific exception types

**ads** (`services/ads/src/autom8_ads/`):
- `app.py` — FastAPI app factory with lifespan (creates singletons: `LaunchService`, `AccountRouter`, `OfferPayloadMapper`, `LaunchIdempotencyCache`, `MetaUrlBuilder`, `StubDataServiceClient`)
- `api/health.py` — `GET /health`
- `api/launch.py` — `POST /api/v1/launches`, `DELETE /api/v1/launches/{offer_id}/{platform}`
- `launch/service.py` — `LaunchService`: 6-step ad launch orchestration
- `launch/mapper.py` — `OfferPayloadMapper`: maps `OfferPayload` -> `LaunchContext`
- `launch/idempotency.py` — `LaunchIdempotencyCache`: in-memory dedup by `(offer_id, platform)`
- `lifecycle/factory.py` — `AdFactory`: wraps strategy execution
- `lifecycle/strategies/v2_meta.py` — `V2MetaLaunchStrategy`: Meta-specific launch strategy (only strategy in use)
- `lifecycle/strategies/base.py` — abstract base strategy
- `platforms/protocol.py` — `AdPlatform`, `DataServiceProtocol` protocols
- `models/offer.py` — `OfferPayload`, `LaunchResponse` (Pydantic v2)
- `models/launch.py` — `LaunchContext`, `LaunchResult`
- `models/enums.py` — `Platform` enum (currently: `meta` only)
- `models/targeting.py`, `models/base.py` — targeting structures
- `routing/config.py` — `AccountRouterConfig`, `AccountRule`
- `routing/router.py` — `AccountRouter`: routes to ad account by platform and spend
- `urls/meta.py` — `MetaUrlBuilder`, `MetaPlatformConfig`: constructs Meta ad manager URLs
- `clients/data.py` — `StubDataServiceClient` (placeholder for data service writes)
- `config.py` — `AdsConfig`
- `errors.py` — `AdsError`, `AdsValidationError`, `AdsBudgetError`, `AdsTransientError`, `AdsPlatformError`, `LaunchInProgressError`

**auth** (`services/auth/src/`):
- `main.py` — FastAPI app instance, middleware wiring (CORS, security headers, rate limiting, error handling), router registration, `/health`, `/ready`, `/health/deps` endpoints
- `config.py` — `Settings(Autom8yBaseSettings)` with RS256 JWT config, Redis config, CORS origins, DB URL
- `auth/jwt_handler.py` — JWT signing and validation
- `auth/api_key_handler.py` — API key verification
- `auth/password.py` — password hashing (argon2/bcrypt)
- `auth/dependencies.py` — FastAPI dependency providers
- `jwks.py` — JWKS utilities for `/.well-known/jwks.json`
- `middleware/business_scope.py`, `rate_limit.py`, `security_headers.py` — HTTP middleware
- `models/` — SQLModel ORM types: `User`, `Role`, `Permission`, `RefreshToken`, `ExternalCredential`, `EncryptionKey`, `OAuthState`
- `routes/auth.py` — auth endpoints
- `routes/rbac.py` — RBAC management
- `routes/api_keys.py` — API key management
- `routes/well_known.py` — JWKS endpoint
- `routes/internal.py` — internal service endpoints
- `routes/admin.py` — admin endpoints
- `routes/charter.py`, `routes/credentials.py`, `routes/oauth.py` — MOTHBALLED (see ADR-VAULT-001, ADR-CHARTER-001)
- `mailer/` — email sending via SendGrid

**auth-mysql-sync** (`services/auth-mysql-sync/src/`):
- `main.py` — Lambda handler + `run_local_sync()` for development
- `sync/orchestrator.py` — `SyncOrchestrator`: 6-step full sync algorithm; also `execute_incremental_sync()`
- `sync/mysql_reader.py` — `MySQLReader`: aiomysql-based queries
- `sync/auth_writer.py` — `AuthWriter`: asyncpg-based upserts to PostgreSQL
- `sync/transformer.py` — `transform_chiropractor_to_business()`, `transform_employee_to_user()`
- `sync/guid_converter.py` — `normalize_chiropractor_guid()`: GUID normalization for ext-id mapping
- `config.py` — `get_settings()` with MySQL + PostgreSQL connection params
- `observability/logger.py` — custom logging setup with `sync_id` correlation
- `observability/metrics.py` — `MetricsEmitter`, `SyncMetrics`, `emit_success_timestamp()`
- `portover/cli.py` — `typer`-based CLI for one-time data portover operations
- `portover/handler.py` — portover logic

**devconsole** (`services/devconsole/src/autom8_devconsole/`):
- `__main__.py` — entry point: initializes OTel, creates FastAPI app, calls `ui.run_with()`, starts uvicorn
- `app.py` — `create_app()` factory; lifespan manages `SpanBuffer`, `SpanStore`, `TempoClient`, HTTP client; defines `POST /v1/traces` and `GET /health`; registers NiceGUI pages (`/` and `/session/{session_id}`)
- `span_buffer.py` — `SpanBuffer`: in-memory ring buffer for `ParsedSpan` objects
- `otlp_receiver.py` — OTLP protobuf HTTP receiver
- `persistence.py` — `SpanStore`: SQLite WAL-mode persistence
- `tempo_client.py` — `TempoClient`: queries Grafana Tempo for historical traces
- `config.py` — `DevconsoleSettings`
- `ui/` — NiceGUI UI components: `ConversationLens`, `DecisionLens`, `PerformanceLens`, `InfrastructureLens`, `SessionTreePanel`, `SideEffectPanel`, `MutationSummaryPanel`, `PayloadDiffPanel`, `theme.py`, `keyboard_shortcuts.py`

**pull-payments** (`services/pull-payments/src/pull_payments/`):
- `handler.py` — Lambda entry point (`lambda_handler`), calls `sync_payments(days_back)`
- `orchestrator.py` — payment sync pipeline: paginated Stripe invoice fetch → business lookup → vertical extraction → data service writes; also manages circuit breaker state
- `clients/data_service.py` — `DataServiceClient`: writes payment records to autom8y-data
- `clients/models.py` — client-layer types
- `staging.py` — S3-based staging for replay/recovery of failed writes
- `replay.py` — replay logic for staged (failed) payment records
- `models.py` — `SyncResult`, `RefundAttribution`
- `metrics.py` — metric emission including dead-man's-switch
- `config.py` — `Settings(LambdaServiceSettingsMixin, Autom8yBaseSettings)`

**reconcile-ads** (`services/reconcile-ads/src/reconcile_ads/`): Mirrors `account-status-recon` structure. Key difference: joins Asana offer data against Meta campaign data rather than a three-source billing join.

**reconcile-spend** (`services/reconcile-spend/src/reconcile_spend/`): Mirrors `account-status-recon` structure with a `clients/asana_resolve.py` for Asana URL enrichment and `stubs.py` for test stubs. Has `clients/data_service.py` with circuit breaker pattern.

**slack-alert** (`services/slack-alert/src/slack_alert/`):
- `handler.py` — only source file; Lambda entry point for SNS → Slack pipeline

**sms-performance-report** (`services/sms-performance-report/src/sms_performance_report/`):
- `handler.py` — Lambda entry point, calls `run_sms_report()`
- `orchestrator.py` — fetch SMS insight from data service, build Slack report
- `clients/data_service.py` — `DataInsightClient` usage
- `report.py`, `readiness.py`, `config.py`

---

## Layer Boundaries

### Two Service Archetypes

**Lambda workers** (account-status-recon, auth-mysql-sync, pull-payments, reconcile-ads, reconcile-spend, slack-alert, sms-performance-report): Flat invocation model. No HTTP server. Lambda handler → orchestrator → domain modules.

**FastAPI services** (ads, auth, devconsole): HTTP request → router → service/dependency → domain modules.

### Import Graph: Lambda Workers

```
handler.py (entry point)
  ↓ imports
orchestrator.py (hub — imports all domain modules)
  ↓ imports
fetcher.py / rules.py / joiner.py / readiness.py (domain leaves)
  ↓ imports
models.py (leaf — imported by all, imports only external SDKs)
config.py (leaf — imported by all, imports only autom8y_config)
errors.py (leaf — no internal imports)
metrics.py (leaf — imports models, autom8y_telemetry)
report.py (leaf — imports models, autom8y_slack)
```

The orchestrator is the hub in every Lambda worker service. Domain modules (`fetcher`, `joiner`, `rules`, `readiness`, `report`) are leaves that do not import each other.

### Import Graph: ads (FastAPI)

```
app.py (hub — wires all singletons)
  ↓ imports
api/launch.py, api/health.py (route modules — import service layer)
  ↓ imports
launch/service.py (LaunchService — hub of business logic)
  ↓ imports
launch/mapper.py, launch/idempotency.py (leaves)
lifecycle/factory.py → lifecycle/strategies/v2_meta.py (strategy chain)
routing/router.py ← routing/config.py
urls/meta.py ← urls/... (URL construction)
platforms/protocol.py (protocol — only imported, not importing)
models/ (leaf — imported by all)
errors.py, config.py (leaves)
```

### Import Graph: auth (FastAPI)

```
src/main.py (app instance, middleware, router registration)
  ↓ includes
src/routes/{auth,rbac,api_keys,well_known,internal,admin}.py
  ↓ imports
src/auth/{jwt_handler,api_key_handler,password,dependencies}.py
src/models/{user,role,permission,...}.py (SQLModel ORM, leaf)
src/middleware/{business_scope,rate_limit,security_headers}.py (HTTP middleware)
src/utils/{audit,brute_force,rate_limit}.py (utility leaves)
src/mailer/{service,templates,schemas}.py (email subsystem)
src/config.py, src/jwks.py (configuration leaves)
```

### Shared SDK Layer (Cross-Cutting)

All services import from the `autom8y-*` SDK family (resolved via uv workspace):

| SDK | Purpose | Consumers |
|---|---|---|
| `autom8y-config` (`Autom8yBaseSettings`, `LambdaServiceSettingsMixin`) | Settings management, SSM/Secrets Manager resolution | All services |
| `autom8y-log` (`get_logger`, `configure_logging`) | Structured logging via structlog | All services |
| `autom8y-telemetry` (`instrument_lambda`, `instrument_app`, `record_side_effect`) | OTel tracing, Lambda instrumentation, CloudWatch metrics | All except devconsole (which receives telemetry) |
| `autom8y-http` (`Autom8yHttpClient`) | HTTP client with retry, circuit breaker | account-status-recon, devconsole, reconcile-spend, sms-performance-report |
| `autom8y-interop` (`DataInsightClient`, `DataPaymentProtocol`) | Canonical inter-service client contracts | account-status-recon, pull-payments, reconcile-spend, sms-performance-report |
| `autom8y-slack` (`SlackClient`, `format_cloudwatch_alarm`) | Slack Web API Block Kit posting | account-status-recon, reconcile-ads, reconcile-spend, slack-alert, sms-performance-report |
| `autom8y-events` (`EventPublisher`, `DomainEvent`) | EventBridge publishing | account-status-recon, reconcile-ads, reconcile-spend |
| `autom8y-reconciliation` (`ReadinessStatus`, `ReconciliationCsvExporter`) | Reconciliation SDK primitives | account-status-recon, reconcile-ads, reconcile-spend |
| `autom8y-stripe` (`StripeClient`, `Invoice`, `Charge`) | Stripe SDK with pagination | pull-payments |
| `autom8y-meta` | Meta Ads API client | ads |
| `autom8y-auth` | Auth client (JWT validation for S2S) | ads |

### Dependency Direction Rule

Dependencies flow inward: external SDKs → service-specific domain modules → orchestrator → handler. No inter-service Python imports — services communicate only via HTTP (`autom8y-http`) or shared SDK contracts (`autom8y-interop`).

---

## Entry Points and API Surface

### Lambda Entry Points

| Service | Handler Function | Trigger | Event Schema |
|---|---|---|---|
| `account-status-recon` | `handler.lambda_handler` in `account_status_recon/handler.py` | EventBridge scheduled | Empty `{}` |
| `auth-mysql-sync` | `handler` in `auth-mysql-sync/src/main.py` | EventBridge scheduled (4h) | EventBridge event |
| `pull-payments` | `lambda_handler` in `pull_payments/handler.py` | EventBridge scheduled | Optional `{"days_back": N}` override |
| `reconcile-ads` | implied `lambda_handler` in `reconcile_ads/handler.py` | EventBridge scheduled | Empty `{}` |
| `reconcile-spend` | `lambda_handler` in `reconcile_spend/handler.py` | EventBridge scheduled | Empty `{}` |
| `slack-alert` | `lambda_handler` in `slack_alert/handler.py` | SNS (CloudWatch alarm notifications) | SNS Records with JSON-encoded alarm |
| `sms-performance-report` | `lambda_handler` in `sms_performance_report/handler.py` | EventBridge scheduled | Empty `{}` |

All Lambda handlers are decorated with `@instrument_lambda` from `autom8y-telemetry` for automatic OTel instrumentation.

### Local CLI Entry Points

Lambda workers with `__main__.py` (account-status-recon, reconcile-ads, reconcile-spend) support:
- `uv run python -m {package}` — prints JSON to stdout
- `uv run python -m {package} --csv output/` — writes CSV and JSON to disk
- `uv run python -m {package} --json` — prints JSON to stdout

auth-mysql-sync provides a separate portover CLI via `auth-mysql-portover` console script (`portover/cli.py` using typer).

### FastAPI HTTP Surface

**ads** (`autom8_ads`):
- `GET /health` — liveness probe; returns `{"status": "ok", "service": "autom8_ads"}`
- `POST /api/v1/launches` — accepts `OfferPayload` JSON body; returns `LaunchResponse`; idempotent on `(offer_id, platform)`; returns 409 if launch in progress, 422 for validation errors
- `DELETE /api/v1/launches/{offer_id}/{platform}` — clears idempotency cache to enable retry

**auth** (`auth/src/main.py`):
- `GET /health` — liveness probe (no dependencies checked)
- `GET /ready` — readiness probe (checks DB + optional AWS Secrets Manager)
- `GET /health/deps` — detailed dependency check (DB, Secrets Manager, Redis)
- `POST /auth/login` (rate-limited) — password login, issues JWT
- `POST /auth/refresh` — refresh token exchange
- `POST /auth/logout` — token revocation
- Routes in `routes/auth.py` — authentication flows
- Routes in `routes/rbac.py` — role and permission management
- Routes in `routes/api_keys.py` — API key lifecycle
- `GET /.well-known/jwks.json` (`routes/well_known.py`) — RS256 public key set
- Routes in `routes/internal.py` — internal S2S endpoints
- Routes in `routes/admin.py` — admin operations
- MOTHBALLED: `routes/credentials.py`, `routes/oauth.py`, `routes/charter.py`

**devconsole** (`autom8_devconsole`):
- `POST /v1/traces` — OTLP HTTP protobuf receiver (receives spans from instrumented services)
- `GET /health` — health check with buffer size and persistence status
- `GET /` — NiceGUI UI (main console view)
- `GET /session/{session_id}` — bookmarked session view
- `GET /openapi` — OpenAPI documentation

### Middleware (auth service)

```
Request → SecurityHeadersMiddleware → CORSMiddleware → rate_limit_login (login only)
        → error_handling_middleware → route handlers
```

---

## Key Abstractions

### Configuration Pattern

All services use `Autom8yBaseSettings` (from `autom8y-config`) as the base settings class. Lambda services additionally mix in `LambdaServiceSettingsMixin` for SSM/Secrets Manager ARN auto-resolution.

```python
# Canonical pattern in every service:
class Settings(LambdaServiceSettingsMixin, Autom8yBaseSettings):
    _SERVICE_KEY_ALIAS = "SERVICE_API_KEY"
    service_api_key: SecretStr = Field(validation_alias=AliasChoices(...))
    autom8y_data_url: str = Field(default="https://data.api.autom8y.io")

@lru_cache
def get_settings() -> Settings:
    return Settings()
```

Key pattern: `SecretStr` for all sensitive values; `lru_cache` for singleton access; `AliasChoices` for canonical `SERVICE_API_KEY` plus legacy aliases.

### Orchestrator Pattern

Every Lambda worker service has a `run_reconciliation()` or `sync_payments()` async function in `orchestrator.py` that:
1. Loads settings from `get_settings()`
2. Runs the domain pipeline using `asyncio.run()` from the synchronous Lambda handler
3. Uses `autom8y_telemetry` spans for each major step
4. Posts results to Slack
5. Emits metrics to CloudWatch
6. (For reconciliation services) publishes `DomainEvent` to EventBridge

### Pydantic v2 Models

All data models use Pydantic v2 (`BaseModel` or `BaseSettings`). Domain value objects use either `@dataclass(frozen=True)` (account-status-recon, reconcile-spend) or `BaseModel` with `ConfigDict(frozen=True)` (ads).

**Key types:**
- `OfferPayload` / `LaunchResponse` in `ads/src/autom8_ads/models/offer.py` — ad launch request/response
- `AccountRecord` / `AccountFinding` / `ReconciliationResult` in `account-status-recon/src/account_status_recon/models.py` — three-way reconciliation domain model
- `ClientRecord` / `Anomaly` / `EnrichedAnomaly` in `reconcile-spend/src/reconcile_spend/models.py` — spend anomaly model
- `SyncSummary` in `auth-mysql-sync/src/sync/orchestrator.py` — sync run summary
- `SyncResult` in `pull-payments/src/pull_payments/models.py` — payment sync result
- `ParsedSpan` in `devconsole/src/autom8_devconsole/span_buffer.py` — OTel span representation

### Protocol Pattern (ads)

`ads` uses Python `Protocol` classes in `platforms/protocol.py`:
- `AdPlatform` — interface for Meta/other ad platforms
- `DataServiceProtocol` — interface for data service writes

Allows injection of mock adapters in tests without subclassing.

### Idempotency Pattern (ads)

`LaunchIdempotencyCache` in `ads/src/autom8_ads/launch/idempotency.py` provides in-memory idempotency on `(offer_id, platform)`. On receipt of a duplicate request:
- Status `in_progress` → raises `LaunchInProgressError` → HTTP 409
- Status `completed` → returns cached `LaunchResponse`

### Readiness Gate Pattern (reconciliation services)

All reconciliation services use a three-tier data freshness gate in `readiness.py`:
- `ReadinessStatus.PASS` — proceed normally
- `ReadinessStatus.WARN` — proceed with degraded-signal banner in report
- `ReadinessStatus.FAIL` — abort with Slack alert, no anomaly detection

### Side Effect Recording Pattern

All Lambda and FastAPI services use `record_side_effect(span, system=..., operation=..., target=..., payload=..., status=...)` from `autom8y-telemetry` to record external mutations (Slack posts, EventBridge publishes, CloudWatch metric writes) as OTel span events.

### Dead-Man's Switch

Both `auth-mysql-sync` and `pull-payments` call `emit_success_timestamp()` after successful completion. A Grafana alert fires if this metric is absent or stale beyond a threshold (8h for auth-mysql-sync, 16h for pull-payments).

---

## Data Flow

### Lambda Worker: Reconciliation Pattern

```
EventBridge (scheduled) → lambda_handler
  → asyncio.run(run_reconciliation())
    → get_settings()                          [env vars / SSM]
    → fetch_all(settings)                     [HTTP → autom8y-data, autom8y-ads, autom8y-asana]
    → evaluate_readiness(sources, settings)   [data freshness check]
    → three_way_join(billing, campaigns, offers) [domain join logic]
    → apply_all_rules(records, thresholds)    [verdict rules → AccountFinding list]
    → build_slack_report(findings, result)    [Block Kit JSON]
    → SlackClient.send_blocks(channel, blocks) [→ Slack Web API]
    → emit_metrics(result)                    [→ CloudWatch]
    → EventPublisher.publish(DomainEvent)     [→ AWS EventBridge]
  → return {"statusCode": 200, "body": result.to_dict()}
```

**Configuration sources** for reconciliation services:
- `SERVICE_API_KEY` / `ACCOUNT_STATUS_RECON_SERVICE_KEY` (SSM or env)
- `META_ACCOUNT_ID`, `AUTOM8Y_DATA_URL`, `AUTOM8Y_ADS_URL`, `AUTOM8Y_ASANA_URL`
- `SLACK_CHANNEL`, threshold percentages, staleness limits

### Lambda Worker: pull-payments

```
EventBridge (scheduled) → lambda_handler
  → asyncio.run(sync_payments(days_back))
    → DataServiceClient.get_business_by_stripe_id()    [→ autom8y-data HTTP]
    → StripeClient.list_invoices(days_back)             [→ Stripe API (paginated)]
    → extract vertical from subscription description    [local transform]
    → DataServiceClient.batch_write_payments()          [→ autom8y-data HTTP]
    → [on failure] stage to S3, replay from staging    [S3 for durability]
    → emit_success_timestamp(NAMESPACE)                [→ CloudWatch dead-man]
```

### Lambda Worker: auth-mysql-sync

```
EventBridge (4h schedule) → handler
  → asyncio.run(_run_sync(settings, sync_id))
    → MySQLReader.connect()             [aiomysql → NHC MySQL]
    → AuthWriter.connect()              [asyncpg → Auth PostgreSQL]
    → SyncOrchestrator.execute_full_sync()
        → MySQLReader.get_sync_eligible_employees()   [MySQL query]
        → MySQLReader.get_chiropractors_by_ids()      [MySQL query]
        → transform_chiropractor_to_business()        [local transform]
        → AuthWriter.upsert_businesses()              [PostgreSQL upsert]
        → transform_employee_to_user()                [local transform]
        → AuthWriter.upsert_users()                   [PostgreSQL upsert]
        → AuthWriter.upsert_memberships()             [PostgreSQL upsert]
        → AuthWriter.deactivate_stale_memberships()   [PostgreSQL update]
    → MetricsEmitter.emit_sync_metrics()             [→ CloudWatch]
    → emit_success_timestamp()                        [→ CloudWatch dead-man]
```

### FastAPI Service: ads

```
HTTP POST /api/v1/launches (from autom8_asana)
  → launch_ads(body: OfferPayload)
    → LaunchService.launch(payload)
      → LaunchIdempotencyCache.get_or_set_in_progress()  [in-memory]
      → AccountRouter.route(platform, spend)              [config-based routing]
      → OfferPayloadMapper.to_launch_context(payload)     [local transform]
      → AdFactory.launch(ctx)
        → V2MetaLaunchStrategy.execute()                  [→ Meta Ads API]
      → StubDataServiceClient.record_campaign()           [stub, no-op]
      → MetaUrlBuilder.ad_account_url(), live_ads_url()  [URL construction]
      → LaunchIdempotencyCache.complete()                 [cache result]
  → return LaunchResponse (campaign IDs + URLs for Asana write-back)
```

**External integrations** for ads:
- Meta Ads API via `autom8y-meta` SDK
- Data service (write-back, currently stub)
- Caller: `autom8_asana` (not in this repo)

### FastAPI Service: auth

```
HTTP POST /auth/login
  → rate_limit_login middleware (Redis-backed)
  → auth route handler
    → User lookup → PostgreSQL (sqlmodel/asyncpg)
    → password verify (argon2/bcrypt)
    → JWT signing (RS256 via PyJWT + private key)
    → RefreshToken create → PostgreSQL
    → [optional] audit log → PostgreSQL
  → return {access_token, refresh_token}

GET /.well-known/jwks.json
  → generate_jwks(settings.jwt_private_key, settings.JWT_KEY_ID)
  → return JWKS (RS256 public key set)
```

**External integrations** for auth:
- PostgreSQL (`DATABASE_URL` via asyncpg/psycopg2)
- Redis (`REDIS_HOST`) for token revocation (ADR-0017) and rate limiting (fail-open)
- AWS Secrets Manager (optional, `AWS_SECRETS_MANAGER_ENABLED`)
- SendGrid (`SENDGRID_API_KEY`) for transactional email (Phase 4)

### FastAPI Service: devconsole

```
OTel-instrumented service → POST /v1/traces (protobuf)
  → otlp_receiver parses protobuf → ParsedSpan objects
  → SpanBuffer.add(span)          [in-memory ring buffer]
  → [background] SpanStore.store_spans() [SQLite WAL]

Browser → GET / or /session/{id}
  → NiceGUI renders five-lens UI (Conversation/Decision/Perf/Infra + tree)
  → ui.timer(0.5) live_tick() pulls from SpanBuffer

Operator → [optional] TempoClient queries Grafana Tempo for historical traces
```

**Configuration sources** for devconsole:
- `SCHEDULING_BASE_URL` — scheduling service for conversation driver
- `TEMPO_BASE_URL` — Grafana Tempo for historical replay
- `PERSISTENCE_ENABLED`, `PERSISTENCE_DB_PATH` — SQLite persistence
- `NICEGUI_PORT`, `OTLP_RECEIVER_PORT`, `SPAN_BUFFER_SIZE`

---

## Knowledge Gaps

1. **Auth service route details**: The route modules (`routes/auth.py`, `routes/rbac.py`, `routes/api_keys.py`, `routes/internal.py`, `routes/admin.py`) were not individually read. Specific endpoint paths and request/response schemas for auth are undocumented.
2. **auth service db layer**: Files in `auth/src/db/`, `auth/src/services/`, `auth/src/redis_client.py`, `auth/src/observability/` were not read. SQLModel schema details and Redis integration specifics are partially inferred from config.
3. **Reconcile-ads orchestrator**: `reconcile-ads/src/reconcile_ads/orchestrator.py` and `rules.py` were not individually read; inferred to mirror `account-status-recon` based on identical module names and pyproject.toml description.
4. **SMS performance report orchestrator**: `sms-performance-report/src/sms_performance_report/orchestrator.py` was not read; pipeline inferred from handler and client structure.
5. **Pull-payments staging/replay**: `staging.py` and `replay.py` contents not read; S3 staging for failed writes inferred from handler response body fields and pyproject test deps.
6. **_template**: No Python source files present; scaffold purpose only.
7. **Shared uv workspace root**: `services/uv.lock` and root `pyproject.toml` not examined; workspace layout and SDK version pinning not documented.
8. **Terraform / infra**: Deployment topology (ECS vs Lambda, VPC, IAM) not documented here (out of scope for Python architecture observation).
