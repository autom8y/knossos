---
domain: architecture
generated_at: "2026-03-25T02:05:48Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "b8da042"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

**Language**: Python 3.12. Build system: Hatchling (`pyproject.toml`). Package manager: uv with workspace support. The installable package is `autom8-data` (wheel: `src/autom8_data`).

**Self-described 4-layer platform**: "4-layer database platform for autom8 ecosystem" — re-exported from `src/autom8_data/__init__.py` lines 1-51.

### Root-Level Sub-packages (7 primary, 2 generated/infra)

| Package | Path | Purpose | Layer |
|---|---|---|---|
| `core` | `src/autom8_data/core/` | Layer 1: Universal data access — SQLModel ORM, Repository pattern, config, logging | L1 |
| `analytics` | `src/autom8_data/analytics/` | Layer 2: Analytics semantic layer — AnalyticsEngine, metrics, dimensions, query pipeline | L2 |
| `api` | `src/autom8_data/api/` | Layer 4: FastAPI REST + gRPC server, routes, auth middleware | L4 (optional) |
| `grpc` | `src/autom8_data/grpc/` | gRPC server infra (adapters + handlers), runs alongside FastAPI on port 50051 | L4 (optional) |
| `services` | `src/autom8_data/services/` | Layer 1.5: Domain services for ad-platform operations (ad, appointment, lead, etc.) | L1.5 |
| `clients` | `src/autom8_data/clients/` | Top-level client re-export stub | utility |
| `utils` | `src/autom8_data/utils/` | Shared utilities: `phone_normalizer.py` | leaf |
| `proto` | `src/autom8_data/proto/autom8/data/v1/` | Generated protobuf/betterproto code — do not hand-edit | generated |

**Layer 3 (Semantic Analytics / AI)**: Referenced in `src/autom8_data/__init__.py` as `semantic_analytics` optional import — not present as a directory in the current codebase. Installed via `openai` + `instructor` extras.

### `core/` Sub-packages (L1 — 5 sub-packages, 11+ files)

| Sub-package/file | Purpose |
|---|---|
| `core/models/` | SQLModel ORM models partitioned by domain: `_base.py` (shared imports), `_scheduling.py` (Appointment, Address, Hours, Employee, BusinessOffer), `_advertising.py`, `_communications.py`, `_platform.py` |
| `core/repositories/` | Domain-specific repository classes: `lead.py`, `appointments.py`, `business.py`, `messages.py`, `dimension_enrichment.py` |
| `core/schemas/` | Pydantic schemas: `_ad_persistence.py` |
| `core/filters/` | SQL filter helpers |
| `core/config.py` | `AppSettings`, `DatabaseSettings`, `AnalyticsSettings`, `DuckDBMode`, `SQLLogLevel` — pydantic-settings with `DB_*` and `ANALYTICS_*` env prefixes |
| `core/repository.py` | Generic `Repository[T]` base class (SQLModel Session, CRUD pattern) |
| `core/schema_meta.py` | `ApiField`, `FieldRole` — SQLModel field metadata for API schema derivation |
| `core/logging.py` | `get_module_logger`, `configure`, `enable_sql_debug` — wraps autom8y-log SDK |
| `core/clock.py` | `utc_now` utility |
| `core/validators.py` | `validate_e164_phone`, `validate_non_empty`, etc. |
| `core/seed_constants.py` | Seed data constants |
| `core/mappers.py` | Shared mapper utilities |
| `core/correlation.py` | Core correlation helpers |
| `core/factories.py` | Session/engine factories |

**Hub package**: `core/models/_base.py` — imported by every domain model file via star-import.

### `analytics/` Sub-packages (L2 — 8 sub-packages)

| Sub-package | Purpose |
|---|---|
| `analytics/core/` | The semantic analytics engine (infra, models, registry, query, joins, dimensions, metrics, output, domain, execution, request) |
| `analytics/engine.py` | `AnalyticsEngine` — primary public class for all analytics queries |
| `analytics/initialization.py` | One-time engine initialization (mock + production backends) |
| `analytics/execution.py` | Re-exports from `analytics/core/execution/` |
| `analytics/insights/` | `InsightDefinition`, `InsightRegistry`, `ComputedFieldSpec` — named query specifications |
| `analytics/primitives/` | Reusable analytics primitives: pacing, health, peer benchmarking, correlation, anomaly, efficiency, creative intelligence, optimization |
| `analytics/services/` | `InsightsService`, `InsightsCache`, `FrameTypeMapper` — service orchestration |
| `analytics/routes/` | FastAPI routers for analytics endpoints: query, insights, intelligence, schema, admin, analytics_health, data_service |
| `analytics/clients/` | `AsanaSectionTimelineClient` for Asana section timeline integration |
| `analytics/entity_source/` | `EntitySourceProvider` protocol for active account filtering |
| `analytics/fixtures/` | `MockDataFixture` — deterministic test data seeding |
| `analytics/vertical_cache.py` | `VerticalCache` |
| `analytics/presets.py` | `BusinessPresets` |

### `analytics/core/` Deep Sub-packages (11 sub-packages)

| Sub-package | Layer rule | Purpose |
|---|---|---|
| `core/infra/` | No internal imports | Foundation: exceptions, logging, caching, connections, materialization, metrics (Prometheus), polars utilities, enrichment views, scheduler |
| `core/models/` | Import from `infra/` only | Core data structures: `Dimension`, `Metric`, `JoinDefinition`, `Protocols`, constraints, `WindowMetric` |
| `core/registry/` | Import from `models/`, `infra/` only | `MetricRegistry`, `CompositeMetric`, `DependencyResolver`, dimension auto-discovery |
| `core/joins/` | Import from `models/`, `infra/` only | `JoinGraph` (NetworkX), `JoinPathOptimizer`, canonical join paths |
| `core/dimensions/` | Import from `models/`, `infra/` only | `DimensionManager`, `DimensionCache`, `DimensionScope`, time intelligence |
| `core/metrics/` | Import from `models/`, `registry/`, `infra/` | `MetricsLibrary`, `MetricAvailability`, `@metric`/`@dimension` decorators |
| `core/output/` | Import from `models/`, `infra/`, `registry/` | `QueryResult`, `DisplayFormatter`, `ResultProcessor`, discovery API types |
| `core/query/` | — | `QueryBuilder`, `QueryPlanner`, `QueryExecutor`, `FilterParser`, `SQLGenerator`, `FactTableResolver` |
| `core/execution/` | Import from `models/`, `infra/`, `registry/` | `QueryOrchestrator`, `RollingAggregator`, `CompositeCalculator`, `RawGrainStrategy` |
| `core/request/` | Import from `models/`, `infra/` | `AnalyticsRequest`, `PreparedRequest`, `RequestTranslator`, `TimeResolver` |
| `core/domain/` | — | `OperationalMode` FSM (NORMAL/CLOSE_PERIOD/INCIDENT/MAINTENANCE) |
| `core/aggregation.py` | — | `PeriodBucket`, `aggregate_by_time_bucket` |

### `api/` Sub-packages (L4)

| Sub-package | Purpose |
|---|---|
| `api/main.py` | FastAPI application factory with lifespan (engine, pool, scheduler, gRPC, auth) |
| `api/routes/` | ~35 routers: CRUD for every domain entity + delegated analytics routes |
| `api/auth/` | JWT auth: `verify_jwt`, `JWTClaims`, `warmup_auth`, `AuthExtractMiddleware` |
| `api/services/` | `PeriodTranslator`, `InsightsService`, `GidMapService`, `HealthScoreService`, `BatchInsightExecutor` |
| `api/schemas/` | Pydantic request/response schemas |
| `api/clients/` | Typed client wrappers |
| `api/data_service_models/` | DataServiceClient API models |
| `api/metrics/` | API-layer Prometheus metrics |
| `api/middleware.py` | `RequestLoggingMiddleware`, `RequestTimeoutMiddleware`, `DeprecationHeaderMiddleware` |
| `api/errors.py` | Centralized exception handler registration |
| `api/rate_limit.py` | SlowAPI rate limiter setup |
| `api/dependencies.py` | FastAPI dependency factories: `get_metric_registry`, `get_query_connection`, `get_query_builder_factory` |

### `services/` (Domain Services, 30+ files)

Dense domain service layer. Each file is a named domain service (e.g. `ad.py`, `appointment.py`, `lead.py`, `payment.py`, `business.py`, `campaign.py`, `review.py`). These use SQLAlchemy async sessions and call core repositories. Also contains: `ad_persistence.py` (decomposed into `ad_persistence_helpers.py`, `ad_persistence_responses.py`).

### `grpc/` Sub-packages

| Sub-package | Purpose |
|---|---|
| `grpc/handlers/` | betterproto `ServiceBase` extensions: `lead.py`, `address.py`, `appointment.py`, `business.py`, `payment.py`, `vertical.py`, `health.py` |
| `grpc/adapters/` | `BaseAdapter` + per-entity adapters: translate proto <-> domain |
| `grpc/server.py` | `GRPCServer`, `create_grpc_server` — lifecycle managed by `api/main.py` |

### `packages/` Workspace Member

`packages/autom8-dev-data/` — development data fixtures package, installed via `testing` extra.

---

## Layer Boundaries

The codebase has an explicitly enforced four-layer model with documented dependency rules in each package's `__init__.py`:

```
Layer 4 (API / gRPC): api/, grpc/
    |
    v
Layer 2 (Analytics Semantic Layer): analytics/
    |
    v (analytics/core sub-packages have their own DAG)
Layer 1 (Core Data Access): core/
    |
    v
Layer 0 (External): MySQL (via SQLAlchemy/asyncmy), DuckDB, Redis, autom8y SDKs
```

### Import Direction (Layer Enforcement)

**Layer 4 imports Layer 2 and Layer 1:**
- `src/autom8_data/api/main.py` imports `AnalyticsEngine`, `ConnectionPool`, `AppSettings` from layers 2 and 1
- `src/autom8_data/api/dependencies.py` imports `MetricRegistry`, `QueryBuilder`, `QueryConnection` from `analytics.core`
- `src/autom8_data/grpc/` imports from `core/` repositories and `grpc/adapters/`

**Layer 2 imports Layer 1:**
- `src/autom8_data/analytics/engine.py` imports `AppSettings`, `get_settings` from `autom8_data.core.config`
- `src/autom8_data/analytics/core/infra/materializer.py` imports `get_module_logger` from `autom8_data.core.logging`
- `src/autom8_data/analytics/core/infra/connection_pool.py` imports `get_settings` from `autom8_data.core.config`

**Layer 1 does NOT import from Layer 2 or Layer 4** (verified: `src/autom8_data/core/__init__.py` only imports from `core/` sub-modules)

### `analytics/core/` Internal DAG (Strictly Enforced)

The dependency rule is declared in each sub-package `__init__.py` docstring:

```
infra/    (no internal imports — leaf package)
  ^
  |
models/   (imports infra/ only)
  ^
  |
registry/ (imports models/, infra/ only)
joins/    (imports models/, infra/ only)
dimensions/ (imports models/, infra/ only)
  ^
  |
metrics/  (imports models/, registry/, infra/)
output/   (imports models/, infra/, registry/)
execution/ (imports models/, infra/, registry/)
request/  (imports models/, infra/)
  ^
  |
query/    (imports all of the above)
```

**infra/ is the leaf (no internal imports)** — confirmed in `src/autom8_data/analytics/core/infra/__init__.py` docstring: "DEPENDENCY RULE: This package must NOT import from any other core/ package."

**models/ is a hub** — consumed by registry/, joins/, dimensions/, metrics/, output/, execution/, request/, query/.

**Known circular dependency workaround**: Three modules in `pyproject.toml` `[tool.mypy.overrides]` are declared with `ignore_missing_imports = true` due to circular import friction: `analytics.core.joins.join_graph`, `analytics.core.metrics_availability`, `analytics.registry`. These are worked around with `TYPE_CHECKING` guards and interface packages.

### Services Layer Position

`src/autom8_data/services/` sits between L1 (core) and L4 (api). Domain services use SQLAlchemy async sessions. The `api/routes/` call into `api/services/` which delegate to `analytics/services/` or directly to `services/` domain services.

### Boundary-Enforcement Pattern

Each `analytics/core/` sub-package declares its dependency rule in the `__init__.py` docstring. Mypy + ruff enforce import hygiene. `import-linter` is in `[dependency-groups]` for architectural import validation.

---

## Entry Points and API Surface

### Entry Point

There is no CLI entry point (no `__main__.py` or `[project.scripts]` in `pyproject.toml`). The package is a **library** consumed by other services. The FastAPI application is the primary runtime surface.

**FastAPI app factory**: `src/autom8_data/api/main.py` — `create_app()` function returns a `FastAPI` instance.

**gRPC server**: `src/autom8_data/grpc/server.py` — `create_grpc_server()` returns `GRPCServer`, started as a side-car in the FastAPI lifespan on port 50051.

### FastAPI Lifespan Initialization Order

From `src/autom8_data/api/main.py` `lifespan()`:
1. JWT auth warmup (if `AUTOM8Y_AUTH_ENABLED=true`)
2. `AnalyticsEngine` initialization (`app.state.analytics.engine`)
3. `ConnectionPool` initialization (`app.state.analytics.connection_pool`)
4. `ThreadPoolExecutor` for DuckDB queries (`app.state.analytics.query_executor_pool`)
5. `MaterializationScheduler` (optional, if `materialization.enabled=True`)
6. `VerticalBenchmarkMaterializer` (optional)
7. `GRPCServer` (optional, if `grpc.enabled=True`)

### API Routes (from `src/autom8_data/api/routes/__init__.py` and `src/autom8_data/analytics/routes/__init__.py`)

**CRUD routers** (from `api/routes/`):
- `addresses_crud_router`, `ad_accounts_crud_router`, `ad_creatives_crud_router`, `ad_insights_crud_router`, `ad_optimizations_crud_router`, `ad_persistence_router`, `ad_platforms_crud_router`, `ad_questions_crud_router`, `ads_crud_router`, `adsets_crud_router`, `appointments_crud_router`, `asset_verticals_crud_router`, `assets_ad_creatives_crud_router`, `assets_crud_router`, `business_offers_crud_router`, `businesses_crud_router`, `campaigns_crud_router`, `employees_crud_router`, `gid_mappings_router`, `health_router`, `hours_crud_router`, `leads_crud_router`, `messages_router`, `neighborhoods_crud_router`, `offers_crud_router`, `payments_crud_router`, `platform_assets_crud_router`, `questions_crud_router`, `reviews_crud_router`, `split_test_configs_crud_router`, `verticals_crud_router`

**Analytics routers** (from `analytics/routes/`):
- `query_router` — analytics query execution
- `insights_router` — insight discovery and batch execution
- `intelligence_router` — market intelligence analysis
- `schema_router` — metric/dimension/period discovery (TDD-0007)
- `analytics_health_router` — health score computation
- `data_service_router` — DataServiceClient API (insights, gid-map)
- `admin_router` — admin/materialization management
- `cleanup_pvs_router`, `cleanup_results_router` — cleanup operations

### Key Exported Interfaces (Cross-Package Contracts)

| Interface | Package | Consumers |
|---|---|---|
| `Repository[T]` | `core.repository` | Domain repositories, tests |
| `AppSettings` | `core.config` | All layers (get_settings() singleton) |
| `AnalyticsEngine` | `analytics.engine` | `api/main.py`, `api/dependencies.py`, tests |
| `MetricRegistry` | `analytics.core.registry` | `QueryBuilder`, `AnalyticsEngine`, API dependencies |
| `QueryBuilder` | `analytics.core.query` | `AnalyticsEngine`, API dependencies |
| `QueryResult` | `analytics.core.output` | API routes, tests |
| `ConnectionPool` | `analytics.core.infra` | `api/main.py` lifespan |
| `MaterializationJob` | `analytics.core.infra.materializer` | `MaterializationScheduler` |
| `InsightDefinition` | `analytics.insights` | `InsightRegistry`, route handlers |
| `AnalyticsBackend` | `analytics.core.infra.backend` | `AnalyticsEngine`, tests |

### Configuration Interface

Environment variables (from `src/autom8_data/core/config.py`):
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE`, `DB_READ_REPLICA_HOST`
- `ANALYTICS_SQL_LOG_LEVEL`, `ANALYTICS_SQL_LOG_FILE`, `ANALYTICS_QUERY_TIMEOUT_SECONDS`
- `AUTOM8Y_AUTH_ENABLED`, `AUTH__JWKS_URL`
- `AUTOM8Y_ENVIRONMENT` (from autom8y-config SDK)

---

## Key Abstractions

### 1. `Dimension` (dataclass, `analytics/core/models/dimension.py`)

A GROUP-BY attribute for SQL analytics. Fields: `name`, `table`, `column`, `dimension_type` (`DimensionType`: TIME, GEOGRAPHIC, CATEGORICAL, IDENTIFIER), `nullable`, `null_handling`, `auto_order`, `fallback_source` (COALESCE tuple), `default_value`. Used by `MetricRegistry`, `QueryBuilder`, `DimensionManager`, `QueryPlanner`.

### 2. `Metric` (dataclass, `analytics/core/models/metric.py`)

An aggregatable measure with business rules for vertical availability. Key fields: `MetricContract` (versioning, deprecation), `MetricAggregation`, `MetricVerticalScope`. Subtype `QueryableMetric` extends with SQL column reference. Subtype `WindowMetric` adds window function specs.

### 3. `CompositeMetric` (dataclass, `analytics/core/registry/composite.py`)

A metric calculated from other metrics using a `Formula`. Formulas: `DivisionFormula`, `SumFormula`, `AverageFormula`, `WeightedAverageFormula`, `PercentageFormula`, `CostDivisionFormula`, `InverseDivisionFormula`, `DateDifferenceFormula`, `OffsetFormula`. Powers derived metrics like CPS, CTR, conversion rates.

### 4. `MetricRegistry` (class, `analytics/core/registry/metric_registry.py`)

Central registry for dimensions, metrics, and joins. Consumers register `Dimension`, `Metric`, `JoinDefinition` objects. Provides lookup, dependency resolution, join graph construction. Singleton pattern via `get_metric_registry()` in `api/dependencies.py`.

### 5. `JoinGraph` (class, `analytics/core/joins/graph.py`)

NetworkX-based directed graph of FK relationships. Nodes = tables, edges = join conditions. Used by `JoinPathOptimizer` to discover minimum-cost join paths. Canonical paths defined in `analytics/core/joins/canonical_paths.py` (e.g., `CAMPAIGN_PATH`, `APPOINTMENT_CAMPAIGN_PATH`).

### 6. `AnalyticsEngine` (class, `analytics/engine.py`)

The primary public API for analytics queries. Instance-based with lazy initialization (`auto_initialize` flag). Three-layer caching strategy (memory -> Redis -> disk). Async-first. Accepts business-language queries (`business="+1..."`, `vertical="dental"`, `metrics=["cps"]`, `period="last_quarter"`). Delegates to `QueryOrchestrator` which uses `QueryBuilder` + `QueryExecutor`.

### 7. `MaterializationJob` (class, `analytics/core/infra/materializer.py`)

Orchestrates MySQL -> Parquet sync with atomic swap via symlink. Provides 551x performance improvement over MySQL ATTACH. Safety: asyncio.Lock, row count validation (1% tolerance), SHA256 checksums, 96-version retention. Table metadata: `__table_type__` (fact/dimension), `__date_column__`, `__materialization__` (lookback_days) declared on SQLModel classes.

### 8. `ConnectionRouter` (class, `analytics/core/infra/connection_router.py`)

Routes queries to LOCAL (Parquet), ATTACH (MySQL via DuckDB ATTACH), or HYBRID (prefer Parquet, fallback to ATTACH if stale). Mode controlled by `DuckDBMode` enum in `AppSettings`.

### 9. `InsightDefinition` (Pydantic model, `analytics/insights/`)

Named query specification: `name`, `display_name`, `description`, `metrics`, `dimensions`, `computed_fields` (`ComputedFieldSpec`), `freshness_config`, `required_filters`. Stored in `InsightRegistry`. Executed by `InsightsService`.

### 10. `ApiField` + `FieldRole` (factory, `core/schema_meta.py`)

SQLModel field decorator that carries API metadata: `api_alias`, `roles` (AUTO, IMMUTABLE, FK), `validator`. Used by schema factory to derive Create/Update/Record Pydantic schemas from SQLModel source of truth.

### Design Patterns Observed

| Pattern | Where | Description |
|---|---|---|
| Protocol-based abstraction | `analytics/core/models/protocols.py` | `AnalyticsBackend`, `CacheProtocol`, `RegistryProtocol` — enables test injection without ABC inheritance |
| Canonical join paths | `analytics/core/joins/canonical_paths.py` | Pre-defined join paths (e.g., `CAMPAIGN_PATH`) to prevent accidental Cartesian products |
| Star-import aggregation | `core/models/_base.py` | All domain models star-import shared SQLModel/SQLAlchemy imports from `_base.py` |
| Enrichment views | `analytics/core/infra/enrichment_views.py` | DuckDB VIEWs mirror Parquet post-processing so both backends (LOCAL and ATTACH) see enriched columns |
| Atomic Parquet swap | `analytics/core/infra/materializer.py` | Symlink-based atomic rename for zero-downtime materialization updates |
| Operational mode FSM | `analytics/core/domain/operational_mode.py` | Four-mode FSM (NORMAL/CLOSE_PERIOD/INCIDENT/MAINTENANCE) persisted in Redis, governs staleness thresholds |
| Three-layer cache | `analytics/core/infra/cache.py` | Memory -> Redis -> disk for query results and dimension resolution |
| Dependency DAG resolution | `analytics/core/registry/dependency.py` | `DependencyResolver` validates no circular composite metric dependencies, topological sort for resolution order |

---

## Data Flow

### Pipeline 1: MySQL -> Parquet Materialization

```
MySQL (source)
  -> QueryConnection.from_env() (autom8_data.core.config.DatabaseSettings.mysql_url)
  -> MaterializationJob.sync_all_tables() (analytics/core/infra/materializer.py)
    -> Per-table: query MySQL with lookback window (fact tables) or full (dimension tables)
    -> Apply post-processing: enrichment columns (business_phone, vertical, campaign_objective, adset targeting, etc.)
    -> Write versioned Parquet snapshot (atomic symlink swap)
    -> Update SyncManifest + TableManifest (JSON metadata alongside Parquet files)
  -> Storage: configurable directory (PARQUET_DIR env or AppSettings)
```

**Merge point**: Enrichment categories A-E applied during materialization:
- Category A: `calls.business_phone`, `messages.business_phone` via chiropractor dual-path join
- Category B: `adsets.target_*` extracted from `targeting_config` JSON
- Category C: `campaigns.campaign_objective` from `OBJECTIVE_ID_TO_NAME` mapping
- Category D: `payments.office_phone` COALESCE NULL resolution
- Category E: `vertical` column on all 5 fact tables

### Pipeline 2: Analytics Query Execution

```
HTTP request (POST /query or via AnalyticsEngine.get())
  -> RequestTranslator (analytics/core/request/translator.py) — business terms -> technical terms
  -> TimeResolver — period string -> date range
  -> ConnectionRouter.get_connection() (analytics/core/infra/connection_router.py)
    -> DuckDBMode.LOCAL -> LocalDuckDBConnection (Parquet)
    -> DuckDBMode.ATTACH -> IbisConnectionAdapter (MySQL ATTACH)
    -> DuckDBMode.HYBRID -> Parquet if fresh (staleness < threshold), else ATTACH
  -> QueryOrchestrator (analytics/core/execution/orchestrator.py)
    -> QueryBuilder -> QueryPlanner -> SQLGenerator
      -> JoinGraph.find_path() -> JoinPathOptimizer (canonical paths preferred)
      -> FilterParser (filter DSL -> SQL WHERE)
      -> FactTableResolver (multi-fact queries)
    -> QueryExecutor.execute()
      -> DuckDB or MySQL query execution
      -> CompositeCalculator: resolve composite metrics from base results
      -> RollingAggregator: trailing period window aggregation
  -> ResultProcessor -> QueryResult (Polars DataFrame + metadata)
  -> DisplayFormatter (optional) -> formatted output
```

**Cache intercept**: `QueryExecutor` checks three-layer cache (memory L1 -> Redis L2) before executing SQL. Cache key is deterministic hash of (metrics, dimensions, filters, period).

### Pipeline 3: FastAPI Request -> Response (CRUD)

```
HTTP request -> FastAPI routing
  -> Auth middleware (AuthExtractMiddleware extracts bearer token)
  -> verify_jwt / require_auth (JWKS validation via autom8y-auth SDK)
  -> Route handler (api/routes/*.py)
    -> get_db() dependency -> SQLAlchemy AsyncSession
    -> api/services/ or services/ domain service -> async CRUD
    -> core/repositories/ -> SQLModel query
  -> Pydantic response serialization
```

### Pipeline 4: gRPC Request -> Response

```
gRPC client request (port 50051)
  -> GRPCServer (grpc/server.py) -> betterproto handler (grpc/handlers/*.py)
  -> Handler creates adapter with fresh async SQLAlchemy session
  -> grpc/adapters/*.py -> core/repositories/ or services/
  -> proto <-> domain translation in adapter
  -> gRPC response
```

### Pipeline 5: Materialization Scheduling

```
APScheduler timer (15-minute clock-aligned intervals)
  -> MaterializationScheduler / RegistryDrivenScheduler (analytics/core/infra/scheduler/)
  -> Redis distributed lock (prevents multi-worker double-sync)
  -> MaterializationJob.sync_all_tables() (see Pipeline 1)
  -> Circuit breaker (autom8y-http SDK) with exponential backoff retry
  -> Prometheus metrics update (analytics/core/infra/metrics.py)
  -> OperationalMode FSM check (analytics/core/domain/operational_mode.py)
    -> MAINTENANCE mode: skip freshness check entirely
    -> CLOSE_PERIOD mode: tighter staleness thresholds
```

### Configuration Merge Points

1. **Environment variables** -> `pydantic-settings` -> `AppSettings` (composite: `DatabaseSettings`, `AnalyticsSettings`, materialization config from `autom8y-config` base)
2. **`DuckDBMode`** (`ANALYTICS_DUCKDB_MODE` env var) -> `ConnectionRouter` selects backend at query time
3. **`OperationalMode`** (Redis-persisted FSM) -> overrides staleness thresholds and sync intervals
4. **`__materialization__` dict on SQLModel class** -> `MaterializationRegistry` discovers per-table lookback_days and sync frequency

---

## Knowledge Gaps

1. **`analytics/execution.py` (top-level file)**: Only seen as import target, not read in detail. Likely re-exports from `analytics/core/execution/`.
2. **`analytics/presets.py` and `analytics/vertical_cache.py`**: Not read individually — purpose inferred from exports (`BusinessPresets`, `VerticalCache`).
3. **`api/schema_factory.py`**: Not read — inferred to implement the `ApiField`/`FieldRole` -> schema derivation pipeline documented in `core/schema_meta.py`.
4. **`packages/autom8-dev-data/`**: Workspace member not explored. Contains dev fixtures per `pyproject.toml` testing extra.
5. **`proto/autom8/data/v1/`**: Generated protobuf code — not read (mypy suppresses errors on this package entirely).
6. **`analytics/primitives/` internals**: Only `__init__.py` read. The nine primitive sub-packages (anomaly, correlation, creative, efficiency, health, optimization, pacing, peer, shared) have their models documented via exports but implementation not traced.
7. **`alembic/`**: Schema migrations not explored — present at project root.
8. **`tests/`**: Not in scope for architecture observation.
9. **`config/`**: Root-level config directory not explored (likely contains environment-specific YAML or dotenv files).
