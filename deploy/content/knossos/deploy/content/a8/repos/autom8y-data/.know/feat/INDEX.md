---
domain: feat/index
generated_at: "2026-03-18T21:15:00Z"
expires_after: "30d"
source_scope:
  - "./src/**/*.py"
  - "./docs/**/*.md"
  - "./.know/*.md"
  - "./alembic/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "8c9dfeb"
confidence: 0.91
format_version: "1.0"
---

# Feature Census

> 28 features identified across 7 categories. 22 recommended for GENERATE, 6 recommended for SKIP.

---

## analytics-engine

| Field | Value |
|-------|-------|
| Name | Analytics Engine Facade |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.98 |

**Source Evidence**:
- `src/autom8_data/analytics/engine.py`: 2,377-line facade class; primary async entry point `await engine.get(business=..., metrics=..., period=...)`; referenced by ADR-001, ADR-007, ADR-011
- `README.md`: Described as the primary query API; shown in Quick Start example
- `docs/INDEX.md`: TDD-0001 through TDD-0002 design this layer
- `.know/design-constraints.md`: TENSION-001 documents ongoing god-class decomposition

**Rationale**: GENERATE — 1+ decision records (ADR-0001 through ADR-0012), 10+ implementation files (engine.py plus all execution/initialization sub-modules), user-facing interface surface exists (`GET /analytics/query`), and multiple packages depend on `AnalyticsEngine`. This is the computational heart of the platform.

---

## query-pipeline

| Field | Value |
|-------|-------|
| Name | SQL Query Compilation Pipeline |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `src/autom8_data/analytics/core/query/`: 10 files — `builder.py`, `planner.py`, `executor.py`, `sql_generator.py`, `fact_resolver.py`, `filter_merge.py`, `filters.py`, `dimension_resolver.py`, `window_metric_sql.py`, `options.py`
- `docs/INDEX.md`: ADR-0001 (Split-Query Strategy), ADR-0012 (Dimension Resolution Strategy), ADR-TEMPORAL-006 (Temporal Join Aggregation Modes)
- `.know/scar-tissue.md`: SCAR-001 (Cartesian product prevention), SCAR-002 (non-deterministic raw grain ordering)

**Rationale**: GENERATE — 10 implementation files, 3 decision records directly addressing this feature, and user-facing surface via `GET /analytics/query`. The split-query strategy and Cartesian prevention are hallmark design decisions that require thorough knowledge documentation.

---

## join-graph

| Field | Value |
|-------|-------|
| Name | Join Graph and Path Optimization |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `src/autom8_data/analytics/core/joins/`: 3 files — `graph.py`, `optimizer.py`, `canonical_paths.py`
- `.know/architecture.md`: Describes `JoinGraph` as NetworkX DiGraph with Dijkstra's algorithm; bidirectional edges; cycle and pathological depth protection
- `docs/INDEX.md`: ADR-0002 (NetworkX for Join Graph Traversal)

**Rationale**: GENERATE — 1 decision record (ADR-0002), cross-cutting concern imported by query pipeline and engine, user-facing impact (incorrect joins surface as silent result corruption). The join graph is an essential conceptual model for understanding multi-table analytics correctness.

---

## metric-registry

| Field | Value |
|-------|-------|
| Name | Metric Registry and Composite Metric System |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `src/autom8_data/analytics/core/registry/`: 4 files — `metric_registry.py`, `composite.py`, `dependency.py`, `discovery.py`
- `src/autom8_data/analytics/core/metrics/`: `library.py`, `decorators.py`
- `docs/INDEX.md`: ADR-0004 (Formula Abstraction for Composite Metrics), ADR-0010 (Metric Contract for Deprecation Lifecycle)
- `.know/architecture.md`: `FORMULA_REGISTRY` dict, `DivisionFormula`, `SumFormula`, `WeightedAverageFormula`, 10+ formula types

**Rationale**: GENERATE — 2 decision records, 6 implementation files, user-facing API surface (metric names appear in all query endpoints), multiple packages depend on `MetricRegistry`. The formula hierarchy and YAML-driven composition is a distinctive architectural pattern.

---

## materialization-pipeline

| Field | Value |
|-------|-------|
| Name | MySQL-to-Parquet Materialization Pipeline |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.98 |

**Source Evidence**:
- `src/autom8_data/analytics/core/infra/materializer.py`: Core file; atomic symlink swap, row-count validation, SHA256 checksums, 96-version retention
- `src/autom8_data/analytics/core/infra/scheduler/`: 8 files — `_base.py`, `_benchmark.py`, `_circuit_breaker.py`, `_materialization.py`, `_registry.py`, `_retry.py`, `_exceptions.py`, plus `__init__.py`
- `docs/INDEX.md`: ADR-0017 through ADR-0024 (8 ADRs dedicated to materialization)
- `docs/operations/materialization/`: `README.md`, `rollout-checklist.md`
- `.know/scar-tissue.md`: SCAR-003 (backtick quoting incident causing total sync failure)

**Rationale**: GENERATE — 8 dedicated decision records, 10+ implementation files, production-critical path that delivers the 551x performance improvement. The atomic swap and circuit breaker patterns are non-obvious and essential to understand.

---

## connection-routing

| Field | Value |
|-------|-------|
| Name | Connection Routing (LOCAL / ATTACH / HYBRID) |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.96 |

**Source Evidence**:
- `src/autom8_data/analytics/core/infra/connection_router.py`: `ConnectionRouter`, `ConsumerCircuitBreaker`; three DuckDB modes
- `src/autom8_data/analytics/core/infra/parquet_connection.py`: `LocalDuckDBConnection` (8.7ms path)
- `src/autom8_data/analytics/core/infra/backend.py`: `IbisConnectionAdapter`, `AnalyticsBackend`, `QueryableConnection` protocols
- `.know/architecture.md`: Documents 551x speedup on Parquet path vs 4,821ms MySQL ATTACH path
- `docs/INDEX.md`: ADR-0017 (DuckDB Materialization Strategy)

**Rationale**: GENERATE — Cross-cutting concern touching infra and query layers, user-facing performance impact, decision record (ADR-0017). The HYBRID mode with Parquet-prefer and ATTACH fallback is a subtle routing contract.

---

## insights-layer

| Field | Value |
|-------|-------|
| Name | Named Insights and Post-Processor Chain |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.96 |

**Source Evidence**:
- `src/autom8_data/analytics/insights/`: `models.py`, `registry.py`, `post_processor.py`, `computed_fields.py`
- `src/autom8_data/analytics/insights/processors/`: `reconciliation.py`, `coverage.py`, `sms_cost.py`
- `src/autom8_data/analytics/insight_executor.py`: `InsightExecutor`
- `docs/INDEX.md`: ADR-0011 (Insights Layer), ADR-RCWD-002 and ADR-RCWD-003 referenced in `window_aggregation.py`
- `src/autom8_data/analytics/routes/insights.py`: `POST /analytics/insights/{name}` user-facing endpoint

**Rationale**: GENERATE — 2 decision records, 8 implementation files, user-facing endpoint surface (`GET /analytics/insights`). Named insights are a higher-order abstraction above raw metric queries — the post-processor chain (reconciliation, coverage, SMS cost) is domain-specific and non-obvious.

---

## dimension-management

| Field | Value |
|-------|-------|
| Name | Dimension Management and Time Intelligence |
| Category | Core Analytics Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.92 |

**Source Evidence**:
- `src/autom8_data/analytics/core/dimensions/`: 7 files — `manager.py`, `time_dimensions.py`, `time_intelligence.py`, `cache.py`, `computed.py`, `overrides.py`, `scope.py`
- `docs/INDEX.md`: ADR-0005 (Trailing vs Last Period Semantics), ADR-0012 (Dimension Resolution Strategy), ADR-0007 (Attribution Date Source)
- `.know/test-coverage.md`: Notes thin coverage — `time_intelligence.py`, `cache.py`, `manager.py`, `overrides.py` have 0 direct test file references

**Rationale**: GENERATE — 3 decision records, 7 implementation files, cross-cutting dependency (imported by engine, request pipeline). The distinction between trailing and last period semantics is subtle and frequently misunderstood.

---

## batch-insight-execution

| Field | Value |
|-------|-------|
| Name | Batch Insight Execution (Dual Path) |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `src/autom8_data/analytics/services/batch_insight_executor.py`: 505 lines, `BatchInsightExecutor`
- `src/autom8_data/analytics/services/unified_batch_executor.py`: 764 lines, `UnifiedBatchExecutor`
- `.know/design-constraints.md`: TENSION-002 documents two paths mounted at different endpoints
- `src/autom8_data/analytics/routes/insights.py`: Both paths exposed at `POST /insights/{name}/execute/batch` and `POST /insights/{name}/batch`

**Rationale**: GENERATE — User-facing endpoint surface, multiple module files, and a documented design tension (TENSION-002). The dual-path nature is a known technical debt risk that knowledge documentation should capture.

---

## market-intelligence

| Field | Value |
|-------|-------|
| Name | Market Intelligence Analytics (Efficiency, Anomaly, Creative, Optimization, Peer) |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.93 |

**Source Evidence**:
- `src/autom8_data/analytics/primitives/`: 11 sub-packages — `anomaly/` (3 files), `correlation/` (3 files), `creative/` (5 files), `efficiency/` (4 files), `health/` (5 files), `optimization/` (4 files), `pacing/` (3 files), `peer/` (4 files), `shared/` (3 files), `config/` (2 files), plus `entity_keys.py`
- `src/autom8_data/analytics/routes/intelligence.py`: `POST /api/v1/intelligence/market-efficiency`, `/anomaly-detection`, `/creative-intelligence`, `/budget-optimization`, `/peer-ranking`
- `.know/test-coverage.md`: Full primitives test coverage confirmed

**Rationale**: GENERATE — 37+ implementation files across sub-packages, 5 user-facing API endpoints under `/intelligence`, multiple statistical domains. Market intelligence is a distinct product surface distinguishing this platform from a simple CRUD API.

---

## analytics-health-scoring

| Field | Value |
|-------|-------|
| Name | Analytics Health Score Service |
| Category | Core Analytics Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `src/autom8_data/analytics/services/health_score_service.py`: Orchestrates entity resolution, metric fetching, benchmark lookup, and score computation
- `src/autom8_data/analytics/primitives/health/`: `ahi.py`, `chi.py`, `formulas.py`, `models.py`, `score_engine.py`
- `src/autom8_data/analytics/routes/analytics_health.py`: User-facing endpoint
- `src/autom8_data/analytics/primitives/shared/benchmark.py`: `BenchmarkStore` for industry benchmarks

**Rationale**: GENERATE — 6+ implementation files, user-facing endpoint, cross-cutting dependency on benchmark store and health primitives. AHI/CHI formulas are domain-specific business logic.

---

## rest-api-layer

| Field | Value |
|-------|-------|
| Name | FastAPI REST API Layer |
| Category | User-Facing Interface |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.97 |

**Source Evidence**:
- `src/autom8_data/api/main.py`: `create_app()` factory; lifespan management; middleware chain; 30+ router registrations
- `src/autom8_data/api/routes/`: ~40 route files (30 CRUD entities + analytics routes)
- `docs/INDEX.md`: PRD-0008 (API Excellence), ADR-0033–ADR-0037 (Discovery API, Error Suggestion, QueryResult Output, SQL Capture, Runbook Format)

**Rationale**: GENERATE — 5 decision records, 40+ implementation files, user-facing surface with ~30 CRUD entity endpoints plus analytics endpoints. The layered middleware chain (CORS, auth, rate limiting, request correlation) is non-trivial.

---

## grpc-api-layer

| Field | Value |
|-------|-------|
| Name | gRPC Dual-Protocol CRUD Server |
| Category | User-Facing Interface |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.93 |

**Source Evidence**:
- `src/autom8_data/grpc/server.py`: gRPC server on port 50051
- `src/autom8_data/grpc/handlers/`: 7 handlers — `address.py`, `appointment.py`, `business.py`, `health.py`, `lead.py`, `payment.py`, `vertical.py`
- `src/autom8_data/grpc/adapters/`: Proto-to-domain adapters
- `src/autom8_data/proto/`: Generated protobuf stubs with 6 service definitions

**Rationale**: GENERATE — 15+ implementation files, user-facing service surface (AddressServiceHandler, BusinessServiceHandler, LeadServiceHandler, etc.), optional extra dependency gate (`grpc`). The dual-protocol REST+gRPC design is a key architectural commitment.

---

## jwt-authentication

| Field | Value |
|-------|-------|
| Name | JWT Authentication and RBAC |
| Category | User-Facing Interface |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `src/autom8_data/api/auth/`: 5 files — `jwt.py`, `middleware.py`, `identity_middleware.py`, `dependencies.py`, `exceptions.py`
- `docs/INDEX.md`: PRD-AUTH-001, PRD-AUTH-002, ADR-AUTH-001 through ADR-AUTH-013 (13 auth-specific decisions), PRD-API-KEY-REMOVAL
- `src/autom8_data/api/main.py`: `_initialize_auth()`, `warmup_auth()`, `AuthExtractMiddleware`, `AuthIdentityMiddleware`

**Rationale**: GENERATE — 13 decision records spanning auth design, key format, RBAC granularity, and audit log architecture. The migration from API keys to JWT (PRD-API-KEY-REMOVAL) is a completed initiative with significant history. Auth cross-cuts every API endpoint.

---

## rate-limiting

| Field | Value |
|-------|-------|
| Name | Role-Aware API Rate Limiting |
| Category | User-Facing Interface |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `src/autom8_data/api/rate_limit.py`: `slowapi` integration; 5 rate limit tiers (admin 5/min, heavy analytics 10/min, export 10/min, health 120/min, diagnostic 6/min); `RateLimitIdentity` for per-identity tracking
- `src/autom8_data/api/main.py`: `SlowAPIMiddleware` registration; conditional on `settings.rate_limit.enabled`
- `docs/INDEX.md`: PRD-0010 (Pre-Launch Hardening Sprint, requirement C2)
- `docs/runbooks/RATE_LIMIT_OPERATIONS_RUNBOOK.md`: Operations runbook exists

**Rationale**: GENERATE — Decision record (PRD-0010 C2), operations runbook, user-facing security surface. Role-aware per-identity rate limiting (not per-IP) is a non-obvious implementation detail.

---

## data-service-api

| Field | Value |
|-------|-------|
| Name | DataServiceClient API Contract |
| Category | User-Facing Interface |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `src/autom8_data/analytics/routes/data_service.py`: `POST /api/v1/data-service/insights`, `/gid-map`, `GET /appointments`, `GET /leads`
- `src/autom8_data/api/data_service_models/`: 30 entity model files representing DataServiceClient contract
- `src/autom8_data/api/services/`: `batch_cache.py`, `batch_insight_executor.py`, `gid_map_service.py`, `frame_type_mapper.py`, `period_translator.py`
- `.know/architecture.md`: Notes cross-layer violation (Layer 4 → Layer 2 import in `batch_insight_executor.py`)

**Rationale**: GENERATE — 30+ model files, 4 dedicated endpoints, documented cross-layer architectural tension. The DataServiceClient contract bridges the analytics platform to external consumers with HTTP 207 Multi-Status partial failure support.

---

## crud-entity-layer

| Field | Value |
|-------|-------|
| Name | CRUD Entity Services and Repository Layer |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.93 |

**Source Evidence**:
- `src/autom8_data/services/`: 28 entity service files (ad, adset, appointment, business, campaign, employee, lead, payment, etc.)
- `src/autom8_data/core/repositories/`: 5 repository files (address, appointments, business, lead, messages)
- `src/autom8_data/core/repository.py`: Generic `Repository[T]` base class
- `src/autom8_data/core/models/`: 5 model files organizing domain entities across advertising, communications, platform, scheduling

**Rationale**: GENERATE — 30+ implementation files, user-facing surface via REST and gRPC CRUD endpoints, `Repository[T]` is the foundational abstraction for all data access. The entity model structure (advertising, communications, scheduling domains) warrants documentation.

---

## orm-domain-models

| Field | Value |
|-------|-------|
| Name | SQLModel ORM Domain Entities |
| Category | Core Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `src/autom8_data/core/models/`: 5 model group files — `_advertising.py`, `_base.py`, `_communications.py`, `_platform.py`, `_scheduling.py`
- `alembic/versions/`: 4 migration files documenting schema evolution
- `.ledge/decisions/DIV-007-employee-enabled-null-semantics.md`: Domain-specific decision on NULL semantics

**Rationale**: GENERATE — Domain model groupings, schema migration history, and a dedicated decision record on null semantics. The entity taxonomy (Ad, Business, Lead, Campaign, Appointment, Payment) maps directly to the business domain.

---

## schema-migration

| Field | Value |
|-------|-------|
| Name | Database Schema Migration (Alembic) |
| Category | Infrastructure |
| Complexity | LOW |
| Recommendation | **GENERATE** |
| Confidence | 0.82 |

**Source Evidence**:
- `alembic/versions/`: 4 migrations — `001_baseline_after_unique_constraints.py`, `002_add_attribution_method_to_payments.py`, `003_add_messages_phone_indexes.py`, `004_add_sms_instrumentation_columns.py`
- `alembic.ini`: Alembic configuration
- `docs/INDEX.md`: ADR-0021 (Schema Evolution Policy)

**Rationale**: GENERATE — 1 decision record (ADR-0021 Schema Evolution Policy), migration history with 4 versions tracking attribution method changes and SMS instrumentation columns. Documents the database schema evolution contract.

---

## configuration-system

| Field | Value |
|-------|-------|
| Name | Environment Configuration System (AppSettings) |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.95 |

**Source Evidence**:
- `src/autom8_data/core/config.py`: `AppSettings` (pydantic-settings), `DuckDBMode`, `get_settings`, `Autom8yEnvironment`
- `docs/INDEX.md`: ADR-0025 (Factory Method Configuration Pattern), PRD-0005 (Configuration Strategy)
- `pyproject.toml`: 5 optional extras gates (analytics, api, semantic, grpc, all) controlled by config

**Rationale**: GENERATE — 1 decision record, configuration gates all four layers and optional extras. `AUTOM8Y_ENV`, `DuckDBMode`, `AUTOM8Y_AUTH_ENABLED`, `MATERIALIZATION_SCHEDULER_ENABLED`, and `GRPC_ENABLED` are all operational levers with production consequences.

---

## deployment-infrastructure

| Field | Value |
|-------|-------|
| Name | Cloud Deployment Infrastructure (Fargate / Terraform) |
| Category | Infrastructure |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.85 |

**Source Evidence**:
- `docs/INDEX.md`: ADR-0026 through ADR-0032 (7 ADRs on VPC, NAT, storage, secrets, Terraform state, container image, reusable Fargate module); PRD-0006 (Deployment Infrastructure)
- `docs/INDEX.md`: `deployment-summary-20251209.md` staging deployment
- `Dockerfile`, `Dockerfile.dev`: Container build definitions
- `docker-compose.override.yml`: Local dev override

**Rationale**: GENERATE — 7 decision records, PRD-0006, Dockerfile evidence. The Fargate VPC and Terraform state management decisions are architectural commitments requiring documentation.

---

## query-caching

| Field | Value |
|-------|-------|
| Name | Multi-Tier Query Caching (Memory + Redis) |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.92 |

**Source Evidence**:
- `src/autom8_data/analytics/core/infra/cache.py`: `QueryCache`, `Cache` protocol
- `src/autom8_data/analytics/core/infra/ttl_policy.py`: TTL policy engine
- `src/autom8_data/analytics/core/infra/cache_metrics.py`: `CacheStatsProvider` protocol
- `.know/architecture.md`: "Three-layer caching strategy: memory → Redis → disk"; cache hit achieves ~0.044ms average
- `src/autom8_data/analytics/services/batch_cache.py`, `cache.py`: Batch cache services
- `src/autom8_data/api/services/batch_cache.py`: API-layer cache service

**Rationale**: GENERATE — Cross-cutting concern imported by engine and API layers, Redis distributed lock for materialization, TTL policy controls freshness guarantees. Three-tier cache behavior (L1 memory, L2 Redis, L3 disk) is user-observable.

---

## observability

| Field | Value |
|-------|-------|
| Name | Prometheus Metrics and OpenTelemetry Observability |
| Category | Infrastructure |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `src/autom8_data/analytics/core/infra/metrics.py`: Prometheus metrics for analytics operations
- `src/autom8_data/analytics/core/infra/pool_metrics.py`: Connection pool metrics
- `src/autom8_data/api/metrics/mysql_pool_metrics.py`: MySQL pool Prometheus metrics
- `pyproject.toml`: `autom8y-telemetry[fastapi,otlp]`, `opentelemetry-instrumentation-httpx`, `prometheus-client>=0.23.1`
- `docs/runbooks/PERFORMANCE_TESTING_ROADMAP.md`: Performance instrumentation strategy

**Rationale**: GENERATE — Multiple implementation files, cross-cutting telemetry dependency, OTEL trace propagation on outbound calls. Observability configuration determines production visibility.

---

## analytics-schema-discovery

| Field | Value |
|-------|-------|
| Name | Analytics Schema Discovery API |
| Category | User-Facing Interface |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.90 |

**Source Evidence**:
- `src/autom8_data/analytics/discovery.py`: `DiscoveryAPI` class; `list_metrics()`, `list_dimensions()`, `list_periods()`
- `src/autom8_data/analytics/routes/schema.py`: `GET /analytics/schema` endpoint
- `src/autom8_data/analytics/core/output/info.py`: `DimensionInfo`, `MetricInfo`, `PeriodInfo` discovery types
- `docs/INDEX.md`: ADR-0033 (Discovery API Design)

**Rationale**: GENERATE — 1 decision record (ADR-0033), user-facing endpoint (`GET /analytics/schema`), multi-file implementation. Schema discovery is an API-excellence feature enabling client-side introspection.

---

## filter-system

| Field | Value |
|-------|-------|
| Name | Analytics Filter DSL |
| Category | Core Analytics Platform |
| Complexity | MEDIUM |
| Recommendation | **GENERATE** |
| Confidence | 0.88 |

**Source Evidence**:
- `src/autom8_data/core/filters/`: `builder.py`, `introspection.py`, `types.py`
- `src/autom8_data/analytics/core/query/filters.py`: Filter parsing in query pipeline
- `src/autom8_data/analytics/core/query/filter_merge.py`: Filter merge logic
- `.ledge/decisions/ADR-FSH-001-filter-translation-consolidation.md`: Dedicated filter ADR
- `.ledge/decisions/ADR-FSH-002-required-filters-validation-symmetry.md`: Filter validation ADR

**Rationale**: GENERATE — 2 decision records, 5 implementation files, cross-cutting from core to query pipeline. The filter DSL with required-filter validation symmetry is a correctness constraint.

---

## rolling-window-aggregation

| Field | Value |
|-------|-------|
| Name | Rolling Window Metric Aggregation |
| Category | Core Analytics Platform |
| Complexity | HIGH |
| Recommendation | **GENERATE** |
| Confidence | 0.94 |

**Source Evidence**:
- `src/autom8_data/analytics/core/execution/rolling.py`: `RollingAggregator`
- `src/autom8_data/analytics/core/models/window_metric.py`: `WindowMetric` model
- `src/autom8_data/analytics/window_aggregation.py`: Raw grain promotion helpers; references ADR-RCWD-002, ADR-RCWD-003
- `docs/INDEX.md`: ADR-0008 (Backward-Anchored Rolling Windows), ADR-TEMPORAL-006 (Temporal Join Aggregation Modes), PRD-TEMPORAL-EXTENSIONS, ADR-0009 (Raw Grain Aggregation for COUNT_DISTINCT)
- `.know/scar-tissue.md`: SCAR-002 (non-deterministic raw grain ordering in rolling queries)

**Rationale**: GENERATE — 4 decision records, 4 implementation files, known scar (SCAR-002). Backward-anchored semantics and COUNT_DISTINCT raw grain promotion are non-obvious correctness requirements.

---

## section-timeline-client

| Field | Value |
|-------|-------|
| Name | Section Timeline External Client |
| Category | Tooling |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.75 |

**Source Evidence**:
- `src/autom8_data/analytics/clients/section_timeline.py`: `AsanaSectionTimelineClient`, `SectionTimelineClient` protocol, `SectionTimelineData`
- Used by `reconciliation.py` processor; imported optionally

**Rationale**: SKIP — Fewer than 5 implementation files, no decision records, internal dependency of reconciliation post-processor. The Asana integration is an external dependency wrapper, not a platform-defining feature.

---

## phone-normalizer

| Field | Value |
|-------|-------|
| Name | Phone Number Normalization Utility |
| Category | Tooling |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.85 |

**Source Evidence**:
- `src/autom8_data/utils/phone_normalizer.py`: Single file utility
- `src/autom8_data/api/routes/_phone_utils.py`: API-layer phone helpers

**Rationale**: SKIP — Pure utility, fewer than 5 implementation files, no decision records, no cross-cutting concerns. E164 phone normalization is a helper function, not a feature.

---

## vertical-cache

| Field | Value |
|-------|-------|
| Name | Vertical Resolution Cache |
| Category | Tooling |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.78 |

**Source Evidence**:
- `src/autom8_data/analytics/vertical_cache.py`: `VerticalCache` ergonomic wrapper delegating to `DimensionCache`
- Used by analytics routes and services

**Rationale**: SKIP — Single file, no decision records, ergonomic wrapper over `DimensionCache` (part of dimension-management). Its behavior is fully captured by the dimension-management feature.

---

## analytics-test-fixtures

| Field | Value |
|-------|-------|
| Name | Analytics Test Fixture Builder |
| Category | Tooling |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.80 |

**Source Evidence**:
- `src/autom8_data/analytics/fixtures/`: `__init__.py`, `schema.py`
- `.know/test-coverage.md`: Only `test_fixture_scaling.py` exists; no unit tests for fixture builder logic

**Rationale**: SKIP — Pure test support utility, fewer than 5 implementation files, no decision records, no user-facing surface.

---

## payment-backfill

| Field | Value |
|-------|-------|
| Name | Payment Office Phone Backfill Job |
| Category | Tooling |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.82 |

**Source Evidence**:
- `src/autom8_data/analytics/core/infra/backfill_job.py`: Automated backfill for `payments.office_phone IS NULL` via chiropractor customer_id match; dry-run by default; max 100 records per run

**Rationale**: SKIP — Single file, no decision records, narrow operational scope (payment phone resolution), internal maintenance job rather than a platform feature. Fewer than 5 implementation files and no cross-cutting concerns beyond SQLAlchemy session.

---

## business-query-presets

| Field | Value |
|-------|-------|
| Name | Business Analytics Presets |
| Category | Core Analytics Platform |
| Complexity | LOW |
| Recommendation | **SKIP** |
| Confidence | 0.78 |

**Source Evidence**:
- `src/autom8_data/analytics/presets.py`: `BusinessPresets` class; delegates to `engine.get()` for common query patterns (business summary, performance breakdowns)

**Rationale**: SKIP — Single file, no decision records, delegates entirely to `engine.get()`. Pre-built patterns are convenience wrappers rather than a distinct feature with independent conceptual model. Covered under analytics-engine knowledge.

---

## Census Gaps

1. **Semantic Analytics Layer (Layer 3)**: Referenced in `README.md` and `__init__.py` as an optional import (`try: import semantic_analytics`), and listed as a pyproject.toml extra (`semantic = ["openai>=1.0.0", "instructor>=0.2.0"]`). No source files found under `src/autom8_data/semantic_analytics/`. Whether this is planned but not implemented or lives in a separate repository could not be determined.

2. **Admin Route Feature Boundary**: `src/autom8_data/analytics/routes/admin.py` exposes scheduler management (`POST /analytics/admin/sync`, etc.) but it overlaps with both the materialization-pipeline and rest-api-layer features. Not classified as a separate feature due to the overlap, but the admin auth allowlist has its own security documentation (`docs/security/COMPLIANCE-admin-materialization-allowlist.md`, `PENTEST-admin-materialization-allowlist.md`, `REVIEW-admin-materialization-allowlist.md`).

3. **Reconciliation Audit Trail**: `src/autom8_data/analytics/core/infra/audit_trail.py` is a standalone append-only JSONL reconciliation audit log. It could be its own feature (reconciliation observability) but was classified as part of the insights-layer and observability features due to its support role.

4. **GID Mapping System**: `src/autom8_data/api/services/gid_map_service.py` and `src/autom8_data/api/routes/gid_mappings.py` implement a global ID mapping lookup. Its boundaries with the data-service-api feature are blurry — GID maps are consumed as part of the DataServiceClient contract.

5. **Period Index Feature (PRD-0007)**: Referenced in `docs/INDEX.md` as PRD-0007 (Period Index) with test plan `TP-0007-period-index.md`, but no implementation source files were found under the expected location. May be planned or partially implemented.
