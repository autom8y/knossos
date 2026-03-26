---
domain: architecture
generated_at: "2026-03-18T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "51f5e8d"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The project is a single Python package `autom8_data` installed from `src/autom8_data/`. The package description is "4-layer database platform for autom8 ecosystem." It uses optional extras (`analytics`, `api`, `semantic`, `grpc`) to gate layers.

**Top-level package layout:**

| Package | Purpose | Key Exported Types | Relationship |
|---|---|---|---|
| `autom8_data` root | Package init; auto-configures logging; re-exports `core` and `analytics` | `__version__`, `core`, `analytics` | Hub: everything imports through here |
| `autom8_data.core` | Layer 1 — universal data access: SQLModel domain models, repository pattern, filters, mappers | `Repository`, domain model classes, `get_module_logger` | Leaf for domain imports; hub for analytics |
| `autom8_data.core.models` | SQLModel ORM entities (Ad, Business, Lead, Campaign, Appointment, etc.) | `Business`, `Lead`, `Campaign`, `Ad`, `Adset`, `Appointment`, `Payment`, `Vertical`, `discover_models` | Leaf; imported by services, repositories, API layer |
| `autom8_data.core.repositories` | Typed repository implementations over SQLModel (lead, business, messages, etc.) | `LeadRepository`, `BusinessRepository`, etc. | Imports `core.models` |
| `autom8_data.core.filters` | Filter DSL: builder, introspection, types | `FilterBuilder`, filter type definitions | Imported by API layer |
| `autom8_data.core.config` | `AppSettings` (pydantic-settings), `DuckDBMode`, `get_settings` | `AppSettings`, `DuckDBMode` | Leaf; imported everywhere |
| `autom8_data.analytics` | Layer 2 — analytics semantic layer | `AnalyticsEngine` | Hub |
| `autom8_data.analytics.core` | Central analytics infrastructure package; re-exports everything from sub-packages | All analytics types | Hub — re-exports ~90 symbols |
| `autom8_data.analytics.core.models` | Semantic domain models: `Dimension`, `Metric`, `QueryableMetric`, `CompositeMetric`, `JoinDefinition`, `WindowMetric`, protocols | Core dataclass types | Leaf; referenced by all analytics packages |
| `autom8_data.analytics.core.infra` | Infrastructure: exceptions, cache, connection, logging utilities, materializer, scheduler, backend abstraction, Prometheus metrics | `QueryCache`, `ConnectionRouter`, `MaterializationJob`, `AnalyticsBackend`, `QueryConnection` | Leaf toward `models`; hub for query/registry layers |
| `autom8_data.analytics.core.infra.scheduler` | APScheduler-based materialization scheduler with circuit breaker, retry, distributed lock | `MaterializationScheduler`, `CircuitBreakerScheduler` | Imports `infra.materializer` |
| `autom8_data.analytics.core.registry` | `MetricRegistry` — stores dimensions, metrics, joins; composite metric support with `Formula` hierarchy; auto-discovery | `MetricRegistry`, `CompositeMetric`, `DependencyResolver`, `FORMULA_REGISTRY` | Imports `models`, `infra`; imported by engine |
| `autom8_data.analytics.core.query` | Query pipeline: `QueryBuilder` (fluent), `QueryPlanner`, `QueryExecutor`, `SQLGenerator`, `FilterParser`, `FactTableResolver`, `QueryPlan` | `QueryBuilder`, `QueryPlanner`, `QueryExecutor`, `SQLGenerator` | Imports `registry`, `infra`; called by engine |
| `autom8_data.analytics.core.joins` | `JoinGraph` (NetworkX DiGraph), `JoinPathOptimizer`, `CanonicalJoinPath` catalog, `CANONICAL_PATHS` | `JoinGraph`, `JoinPathOptimizer`, canonical path constants | Imports `models`, `infra` |
| `autom8_data.analytics.core.dimensions` | `DimensionManager`, `DimensionCache`, `TimeDimensionRegistry`, `TimePeriodParser`, time intelligence | `DimensionManager`, `TimePeriod`, `PeriodType` | Imports `models` |
| `autom8_data.analytics.core.metrics` | `MetricsLibrary`, `MetricAvailability`, `@metric`/`@dimension` decorators, rolling window definitions | `MetricsLibrary`, decorator functions, window join builders | Imports `models`, `infra`; registered into `MetricRegistry` |
| `autom8_data.analytics.core.output` | `QueryResult`, `ResultProcessor`, `DisplayFormatter`, discovery API types (`DimensionInfo`, `MetricInfo`, `PeriodInfo`) | `QueryResult`, `ResultProcessor`, `DisplayFormatter` | Imports `models`, `infra` |
| `autom8_data.analytics.core.request` | `AnalyticsRequest` (Pydantic), `PhoneVerticalScope`, `RequestTranslator`, `TimeResolver` | `AnalyticsRequest`, `TimeResolver`, `RequestTranslator` | Imports `models`, `infra` |
| `autom8_data.analytics.core.execution` | `QueryOrchestrator`, `CompositeCalculator`, `RollingAggregator` — extracted from engine god-class | Execution classes | Imports `query`, `registry`, `infra` |
| `autom8_data.analytics.core.domain` | `operational_mode` — OperationalMode enum and hydration | `get_current_mode`, `hydrate` | Leaf |
| `autom8_data.analytics.insights` | `InsightDefinition` model, `InsightRegistry`, `PostProcessorRegistry`, post-processor chain | `InsightDefinition`, `InsightRegistry`, `ComputedFieldSpec` | Imports `analytics.core` |
| `autom8_data.analytics.insights.processors` | Named post-processors: `reconciliation`, `coverage`, `sms_cost` | Processor classes | Imports `insights` models |
| `autom8_data.analytics.primitives` | Statistical computation primitives by domain: anomaly, correlation, efficiency, health, optimization, pacing, peer, creative | `SelfDeviationCalculator`, correlation, efficiency primitives | Imports `analytics.core` |
| `autom8_data.analytics.routes` | FastAPI routers for analytics endpoints: `/analytics/query`, `/analytics/insights`, `/analytics/schema`, `/analytics/health`, `/admin`, `/data-service`, `/intelligence` | Router objects | Imports `engine`, `insights` |
| `autom8_data.analytics.services` | Analytics service layer: `BatchInsightExecutor`, cache services, health score, frame type mapper | `BatchInsightExecutor`, service classes | Imports `engine`, `api.models` |
| `autom8_data.analytics.clients` | Outbound HTTP clients for analytics: `section_timeline` | Client classes | Leaf |
| `autom8_data.analytics.entity_source` | Entity source protocol + filter for scoping analytics queries | `EntitySourceProtocol` | Leaf |
| `autom8_data.analytics.fixtures` | Test fixture builder for mock analytics data | `MockDataFixture` | Leaf (test/init support) |
| `autom8_data.api` | Layer 4 — FastAPI REST + gRPC API surface | `FastAPI app factory`, `create_app` | Hub |
| `autom8_data.api.main` | `create_app()` factory: lifespan management, middleware, route registration, gRPC server startup | `create_app` | Hub for API |
| `autom8_data.api.routes` | ~30 CRUD and analytics endpoint routers | Router objects per entity | Imports `services`, `schemas`, `core.models` |
| `autom8_data.api.services` | API-layer services: insights, gid-map, batch executors, lead/appointment detail | Service classes | Imports `analytics.engine`, `core.repositories` |
| `autom8_data.api.schemas` | Pydantic request/response schemas for API contracts | Schema classes | Leaf |
| `autom8_data.api.data_service_models` | Data service client models (~25 entity models) | `PhoneVerticalPair`, per-entity Pydantic models | Leaf |
| `autom8_data.api.auth` | JWT validation, JWKS, auth middleware | `AuthExtractMiddleware`, `warmup_auth` | Leaf |
| `autom8_data.grpc` | gRPC dual-protocol CRUD server | `GRPCServer` | Imports `proto`, `core.repositories` |
| `autom8_data.grpc.handlers` | gRPC service handlers: business, lead, address, appointment, payment, vertical, health | Handler classes | Imports `core.repositories`, `grpc.adapters` |
| `autom8_data.grpc.adapters` | Proto-to-domain model adapters | Adapter classes | Imports `core.models`, `proto` |
| `autom8_data.proto` | Generated protobuf Python stubs (do not hand-edit) | Generated types | Leaf; mypy errors suppressed |
| `autom8_data.services` | Layer 1 entity service classes (~25 services: ad, campaign, lead, business, etc.) | Service classes | Imports `core.repositories`, `core.models` |
| `autom8_data.utils` | Utilities: `phone_normalizer` | Helper functions | Leaf |
| `autom8_data.clients` | Outbound HTTP client wrappers (top-level) | Client classes | Leaf |

**Package counts:** 36 distinct sub-packages under `src/autom8_data/`, approximately 220 Python source files (excluding `__pycache__`).

**Hub packages** (many incoming imports): `autom8_data.core`, `autom8_data.analytics.core`, `autom8_data.analytics.engine`, `autom8_data.analytics.core.infra`

**Leaf packages** (no or few internal imports): `autom8_data.proto`, `autom8_data.core.config`, `autom8_data.analytics.core.models`, `autom8_data.core.models`, `autom8_data.api.schemas`

---

## Layer Boundaries

The package `__init__.py` explicitly documents the four-layer architecture:

```
Layer 1: Core data access       (autom8_data.core)
Layer 2: Analytics              (autom8_data.analytics)
Layer 3: Semantic Analytics     (autom8_data.semantic_analytics) [optional, not present in src]
Layer 4: REST API               (autom8_data.api)
```

**Import direction (traced from source files):**

```
autom8_data.api
    └── imports autom8_data.analytics.engine (AnalyticsEngine)
    └── imports autom8_data.core.config (AppSettings)
    └── imports autom8_data.core.models (SQLModel entities)
    └── imports autom8_data.core.repositories (typed repos)
    └── imports autom8_data.services (entity services)
    └── imports autom8_data.grpc.server (GRPCServer)

autom8_data.analytics.engine
    └── imports autom8_data.analytics.core (MetricRegistry, QueryBuilder, etc.)
    └── imports autom8_data.analytics.insights (InsightRegistry)
    └── imports autom8_data.analytics.initialization (component factories)
    └── imports autom8_data.core.config (AppSettings)
    └── imports autom8_data.core.logging (get_module_logger)
    └── imports autom8y_telemetry, autom8y_config (external SDKs)

autom8_data.analytics.core
    └── imports autom8_data.analytics.core.models (leaf: no internal deps)
    └── imports autom8_data.analytics.core.infra (imports models, core.config, core.logging)
    └── imports autom8_data.analytics.core.registry (imports models, infra)
    └── imports autom8_data.analytics.core.query (imports registry, infra, models)
    └── imports autom8_data.analytics.core.joins (imports models, infra)
    └── imports autom8_data.analytics.core.dimensions (imports models)
    └── imports autom8_data.analytics.core.output (imports models, infra)

autom8_data.services
    └── imports autom8_data.core.repositories
    └── imports autom8_data.core.models

autom8_data.grpc
    └── imports autom8_data.proto (generated stubs)
    └── imports autom8_data.core.repositories
    └── imports autom8_data.core.models
```

**Dependency rules documented in source (from module docstrings):**
- `src/autom8_data/analytics/core/infra/backend.py` line 20: "This module must NOT import from other core/ packages"
- `src/autom8_data/analytics/core/execution/orchestrator.py` line 11: "This module may import from query/, registry/, infra/"
- `src/autom8_data/analytics/core/registry/metric_registry.py` line 13: "This package may import from models/, infra/, and output/"
- `src/autom8_data/analytics/core/registry/composite.py` line 11: "This package may import from models/ and infra/ only"

**Cross-layer violation noted in pyproject.toml:** `autom8_data.analytics.services.batch_insight_executor` imports from `autom8_data.api.data_service_models._base` (Layer 4 → Layer 2 direction inversion). This is the only documented cross-layer irregularity found.

**mypy circular import exemptions** (from `pyproject.toml` lines 183-187):
- `autom8_data.analytics.core.joins.join_graph`
- `autom8_data.analytics.core.metrics_availability`
- `autom8_data.analytics.registry`

These modules have circular imports that required mypy suppression.

**Import enforcement:** The `pyproject.toml` includes `import-linter>=2.11` in dev dependencies, suggesting layer boundary enforcement via lint rules, though the `.importlinter` configuration file was not observed in scope.

---

## Entry Points and API Surface

**No CLI entry point.** This is a library/service package, not a CLI tool. The `pyproject.toml` does not define `[project.scripts]`. There is no `__main__.py`.

**FastAPI application entry point:** `src/autom8_data/api/main.py` — `create_app()` function.

**Application startup trace:**
```
create_app()                              # api/main.py
  ├── lifespan(app)                       # async context manager
  │   ├── hydrate_operational_mode()      # analytics/core/domain/operational_mode.py
  │   ├── AnalyticsEngine(auto_initialize=True)  # analytics/engine.py
  │   │   └── _ensure_initialized()
  │   │       ├── init_mock_connection() or init_production_connection()
  │   │       │   # analytics/initialization.py
  │   │       ├── MetricRegistry()         # register all metrics/dimensions
  │   │       ├── ConnectionRouter(mode)   # HYBRID/LOCAL/ATTACH
  │   │       └── DimensionManager(registry)
  │   ├── MaterializationScheduler.start() # optional, if enabled
  │   └── GRPCServer.start()               # optional, grpc/server.py
  ├── register middleware                   # CORS, auth, rate limit, request logging
  ├── register exception handlers           # api/errors.py
  └── include_router(all routers)           # ~30+ routers
```

**REST API route surface (from `api/routes/__init__.py` and `analytics/routes/`):**

Analytics routes (prefix `/analytics` or similar):
- `GET /analytics/query` — raw metric query endpoint
- `GET /analytics/insights` — insight discovery and execution
- `GET /analytics/schema` — schema discovery (metrics and dimensions)
- `GET /analytics/health` — analytics subsystem health check
- `GET /analytics/data-service` — DataServiceClient insights endpoint
- `POST /analytics/intelligence` — AI-powered analytics
- `GET|POST /admin` — admin auth

CRUD routes (one router per entity):
- addresses, ads, ad_accounts, ad_creatives, ad_insights, ad_optimizations, ad_platforms, ad_questions, adsets, appointments, asset_verticals, assets_ad_creatives, assets, business_offers, businesses, campaigns, employees, gid_mappings, hours, leads, messages, neighborhoods, offers, payments, platform_assets, questions, reviews, split_test_configs, verticals

**gRPC service surface (from `grpc/server.py`, `grpc/handlers/`):**
- `AddressServiceHandler`
- `AppointmentServiceHandler`
- `BusinessServiceHandler`
- `HealthServiceHandler`
- `LeadServiceHandler`
- `PaymentServiceHandler`
- `VerticalServiceHandler`

gRPC listens on port 50051; FastAPI listens on port 8080 (standard uvicorn default).

**Key exported interfaces:**
- `AnalyticsEngine` — primary analytics API; consumed by `api/main.py`, `analytics/services/`, `analytics/routes/`
- `Repository[T]` — generic CRUD base; consumed by all entity repositories
- `MetricRegistry` — metric/dimension/join storage; consumed by engine, query layer
- `QueryBuilder` — fluent query construction; consumed by `QueryOrchestrator`
- `InsightDefinition` — named query spec; consumed by `InsightExecutor`, `InsightRegistry`

---

## Key Abstractions

**1. `AnalyticsEngine` (`src/autom8_data/analytics/engine.py`)**
The facade for all analytics operations. Instance-based (not singleton) with lazy `_ensure_initialized()`. Supports `async with` context manager pattern. Primary method: `await engine.get(business=..., metrics=..., period=...)`.

**2. `QueryableMetric` (`src/autom8_data/analytics/core/models/metric.py`)**
Protocol/dataclass for a metric that can be queried. Has `name`, `sql_expression`, `aggregation`, vertical availability scope. Subclassed by `Metric` (raw) and `CompositeMetric` (derived).

**3. `CompositeMetric` (`src/autom8_data/analytics/core/registry/composite.py`)**
A metric whose value is computed from other metric values using a `Formula`. Formula hierarchy: `DivisionFormula`, `SumFormula`, `AverageFormula`, `WeightedAverageFormula`, `PercentageFormula`, and ~10 more. All formulas are serializable (no lambdas) to support YAML config.

**4. `MetricRegistry` (`src/autom8_data/analytics/core/registry/metric_registry.py`)**
Dict-based central store for `Dimension`, `Metric`, `JoinDefinition`. Supports composite metric dependency resolution. Has `get_metric()`, `get_dimension()`, `get_join()`, `register_*()` methods.

**5. `JoinGraph` (`src/autom8_data/analytics/core/joins/graph.py`)**
NetworkX `DiGraph` storing FK relationships. Uses Dijkstra's algorithm for shortest join path discovery. Bidirectional edges (FK direction + reverse analytics direction). Protects against cycles and pathological join depths.

**6. `QueryResult` (`src/autom8_data/analytics/core/output/result.py`)**
Rich result wrapper: `QueryResult.df` (Polars DataFrame) + execution metadata (`duration_ms`, `cache_hit`, `sql`). Conversion methods: `to_dict()`, `to_pandas()`, `to_json()`, `to_csv()`, `pivot_on()`. Replaces previous `Union[Dict, DataFrame]` return type.

**7. `InsightDefinition` (`src/autom8_data/analytics/insights/models.py`)**
Pydantic frozen model specifying a named query: metrics list, dimensions, computed fields (`ComputedFieldSpec`), post-processors (`PostProcessorRef`), freshness config. Pre-registered in `InsightRegistry`.

**8. `ConnectionRouter` (`src/autom8_data/analytics/core/infra/connection_router.py`)**
Routes queries to DuckDB LOCAL (Parquet), ATTACH (MySQL), or HYBRID modes. Three modes defined in `DuckDBMode` enum. Provides `async with router.get_connection() as conn` pattern. Enables 551x speedup on Parquet path.

**9. `MaterializationJob` (`src/autom8_data/analytics/core/infra/materializer.py`)**
Orchestrates MySQL → Parquet sync. Features: atomic symlink swap, row-count validation (1% tolerance), SHA256 checksums, 96-version retention, `asyncio.Lock` overlap protection.

**10. `Repository[T]` (`src/autom8_data/core/repository.py`)**
Generic SQLModel-based CRUD base class. Methods: `get(id)`, `create(obj)`, `update(obj)`, `delete(id)`, `list(skip, limit)`. Extended by specific entity repositories in `core/repositories/`.

**Design patterns identified:**
- **Protocol-based dependency inversion**: `AnalyticsBackend`, `FilterParserProtocol`, `SQLGeneratorProtocol`, `JoinGraphProtocol`, `CacheProtocol`, `RegistryProtocol` — all in `analytics/core/models/protocols.py`. Allows swapping implementations (test vs prod) without concrete coupling.
- **Sentinel object pattern**: `_PERIOD_NOT_SET = object()` in `insight_executor.py` and analogous patterns in `orchestrator.py` — distinguishes "not provided" from `None`.
- **Formula registry pattern**: `FORMULA_REGISTRY` dict maps string names to `Formula` subclasses for YAML-driven composite metric configuration.
- **God-class decomposition (RD-003)**: `AnalyticsEngine` was decomposed; execution logic extracted to `execution/orchestrator.py`, `execution/composite.py`, `execution/rolling.py`, initialization to `initialization.py`, insight execution to `insight_executor.py`.
- **Optional dependency gating**: `try: import semantic_analytics / except ImportError: pass` in root `__init__.py`; APScheduler and Redis imports guarded by try/except in scheduler.

---

## Data Flow

### Flow 1: Analytics Query (primary path)

```
Caller calls:
  await engine.get(business="+1...", metrics=["cps", "conversions"], period="last_quarter")
    |
    v
AnalyticsEngine._ensure_initialized()     # One-time setup
    ├── MetricRegistry.register_*()        # Load metrics library
    ├── ConnectionRouter(mode=HYBRID)      # Connection routing setup
    └── DimensionManager(registry)         # Dimension index
    |
    v
AnalyticsRequest (Pydantic validation)    # analytics/core/request/models.py
    |
    v
RequestTranslator.translate()             # analytics/core/request/translator.py
    → resolve dimensions, time period, scope filters
    |
    v
QueryOrchestrator.execute()               # analytics/core/execution/orchestrator.py
    |
    v
QueryBuilder("leads", registry, conn)     # analytics/core/query/builder.py (fluent API)
    .dimensions(...)
    .metrics(...)
    .filters(...)
    ├── FactTableResolver → QueryPlan      # splits multi-fact queries to prevent Cartesian
    ├── QueryPlanner                       # join graph traversal
    ├── SQLGenerator                       # generates SQL
    └── QueryExecutor.execute()            # async DuckDB execution
        ├── ConnectionRouter.get_connection()  # picks PARQUET/ATTACH/HYBRID
        │   ├── LocalDuckDBConnection (Parquet) — 8.7ms path
        │   └── IbisConnectionAdapter (MySQL ATTACH) — 4,821ms path
        ├── ThreadPoolExecutor (DuckDB blocking ops)
        ├── QueryCache.get/set             # Redis or memory cache
        └── returns Polars DataFrame
    |
    v
ResultProcessor.process()                 # analytics/core/output/processor.py
    → CompositeCalculator for CompositeMetrics
    → RollingAggregator for rolling window metrics
    |
    v
QueryResult(df=polars_df, duration_ms=..., sql=..., cache_hit=...)
```

### Flow 2: Materialization Pipeline

```
MaterializationScheduler (APScheduler, every 15 min)
    └── MaterializationJob.sync_all_tables()   # analytics/core/infra/materializer.py
        ├── Redis distributed lock (multi-worker guard)
        ├── Circuit breaker check               # scheduler/_circuit_breaker.py
        ├── For each table:
        │   ├── MySQL ATTACH query (DuckDB ATTACH)
        │   ├── Polars DataFrame export
        │   ├── _apply_post_processing()         # adds derived columns (business_phone, targeting fields)
        │   ├── Write Parquet to versioned dir
        │   ├── Row count validation (1% tolerance)
        │   └── SHA256 checksum
        ├── Atomic symlink swap (zero-downtime)
        └── SyncManifest written
```

### Flow 3: Insight Execution (named query)

```
POST /analytics/insights/{insight_name}
    |
    v
InsightExecutor.execute(insight_name, period, filters)   # analytics/insight_executor.py
    ├── InsightRegistry.get(insight_name) → InsightDefinition
    ├── engine.get(metrics=insight.metrics, ...)         # standard analytics query
    ├── ComputedFieldSpec queries (separate SQL to avoid Cartesian)
    ├── DataFrame merge of main result + computed fields
    └── PostProcessorRegistry.run_chain(processors, df)  # analytics/insights/post_processor.py
        └── e.g., ReconciliationProcessor, CoverageProcessor, SmsCostProcessor
```

### Flow 4: CRUD API (Layer 1 path)

```
HTTP request → FastAPI router (api/routes/{entity}_crud.py)
    → Depends(get_session) → AsyncSession (SQLAlchemy asyncmy/MySQL)
    → EntityService (services/{entity}.py)
        → EntityRepository (core/repositories/{entity}.py)
            → BaseRepository[T] (core/repository.py)
                → SQLModel/SQLAlchemy query
                → MySQL
    → Pydantic schema response
```

### Flow 5: gRPC CRUD

```
gRPC client → grpclib (port 50051)
    → Handler (grpc/handlers/{entity}.py)
    → Adapter (grpc/adapters/{entity}.py)  # proto ↔ domain model translation
    → Repository (core/repositories/{entity}.py)
    → MySQL (via AsyncSession)
    → Adapter.to_proto()
    → Proto response
```

### Configuration cascade

`AppSettings` (pydantic-settings) reads from environment variables with prefix `AUTOM8Y_`. Key config points:
- `AUTOM8Y_ENV` → `Autom8yEnvironment` (LOCAL/MOCK/STAGING/PROD)
- `DuckDBMode` (LOCAL/ATTACH/HYBRID) → governs `ConnectionRouter` behavior
- `AUTOM8Y_AUTH_ENABLED` → gates JWT validation at API startup
- `MATERIALIZATION_SCHEDULER_ENABLED` → gates scheduler lifecycle
- `GRPC_ENABLED` → gates gRPC server startup

## Knowledge Gaps

1. **`autom8_data.clients` (top-level)** — only a directory with no Python files visible in the listing; purpose unknown. May be empty or contain only `__init__.py`.

2. **`.importlinter` configuration** — `pyproject.toml` declares `import-linter` in dev deps but no `.importlinter` config file was found in scope; layer enforcement contract may not be active.

3. **`autom8_data.analytics.core.metrics.library`** — the actual metric definitions file. Not read. This is the business logic heart of the metrics layer and would require separate deep-dive for full documentation.

4. **`autom8_data.analytics.core.joins.canonical_paths`** — canonical join path definitions. Not read fully; only `__init__.py` exports observed.

5. **`autom8_data.analytics.primitives`** sub-packages (anomaly, correlation, efficiency, health, optimization, pacing, peer, creative) — only `self_deviation.py` sampled. Full primitive catalog undocumented.

6. **`autom8_data.api.data_service_models`** — ~25 entity models for the data service client. Not individually read; role inferred from file names.

7. **Alembic migration schema** — present at `/alembic/` but out of scope for source architecture.

8. **`semantic_analytics` layer (Layer 3)** — referenced in `__init__.py` as optional but not present under `src/`. May live in a separate package or not yet implemented.
