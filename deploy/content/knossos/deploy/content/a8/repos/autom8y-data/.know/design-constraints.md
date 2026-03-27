---
domain: design-constraints
generated_at: "2026-03-25T02:05:48Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "b8da042"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog Completeness

### TENSION-001: api/services Shim Layer for Migrated Analytics Services

**Type:** Naming mismatch / module relocation drift

**Location:** `src/autom8_data/api/services/` (7 backward-compat shim files)

**Description:** Services were migrated from `api.services.*` to `analytics.services.*` but the original module paths are kept alive as shim re-exports (`from analytics.services.X import * # noqa: F401, F403`). Tracked as HF-7.

**Files:** `src/autom8_data/api/services/insights_service.py`, `src/autom8_data/api/services/batch_insight_executor.py`, `src/autom8_data/api/services/batch_cache.py`, `src/autom8_data/api/services/frame_type_mapper.py`, `src/autom8_data/api/services/health_score_service.py`, `src/autom8_data/api/services/unified_batch_executor.py`, `src/autom8_data/api/services/cache.py`

**Historical reason:** Services were extracted from `api.services` to `analytics.services` to fix layering — analytics shouldn't be a subdirectory of the HTTP API. But callers outside the repo still use the `api.services.*` paths.

**Ideal resolution:** Delete shims after all external callers are updated to import from `analytics.services.*`.

**Resolution cost:** Medium — requires audit of all importers across the org.

---

### TENSION-002: analytics.services Imports From api.* (Layer Inversion)

**Type:** Layer boundary violation

**Location:** `src/autom8_data/analytics/services/insights_service.py` (lines 39, 47, 66), `src/autom8_data/analytics/services/batch_insight_executor.py` (lines 19-20), `src/autom8_data/analytics/services/health_score_service.py` (lines 25-26), `src/autom8_data/analytics/services/batch_cache.py` (line 21), `src/autom8_data/analytics/services/unified_batch_executor.py` (lines 29-30)

**Description:** `analytics.services.*` (semantic layer) imports `api.data_service_models`, `api.metrics`, `api.models`, `api.models_batch`, `api.models_health`, and `api.services.period_translator`. This inverts the intended layer dependency (Core -> Domain -> Application -> API).

**Historical reason:** InsightsService grew organically to serve the DataService API. Models like `InsightsRequest` / `InsightsResponse` live in `api.data_service_models` because they were defined there first.

**Ideal resolution:** Promote shared models to `analytics.core.models` or a dedicated `autom8_data.shared_models` package.

**Resolution cost:** High — affects proto contracts, gRPC adapters, and external callers.

**Critical note:** The import ORDER in `insights_service.py` is load-order-sensitive. `# ruff: noqa: I001` and `# isort: skip_file` guard this. Auto-sort tools must NOT reorder these imports.

---

### TENSION-003: Connection Abstraction Proliferation

**Type:** Premature abstraction / connection type proliferation

**Location:** `src/autom8_data/analytics/core/infra/`

**Description:** Three distinct connection types coexist: `IbisConnectionAdapter` (wraps Ibis BaseBackend for MySQL ATTACH), `LocalDuckDBConnection` (reads Parquet), and `DuckDBPooledConnectionAdapter` (pool member). All implement the `QueryableConnection` protocol but have divergent capabilities. SQL generation must account for which backend is active (CTE-bypass logic for `IbisConnectionAdapter`, ibis-avoidance for WITH-clauses).

**Historical reason:** Connection routing (LOCAL/ATTACH/HYBRID) was added iteratively as the Parquet materialization path matured.

**Ideal resolution:** Unify around a single `QueryableConnection` interface with no adapter-specific code paths in query generation.

**Resolution cost:** High — SQL generator and orchestrator have mode-specific branches throughout.

---

### TENSION-004: Duplicated Budget CTE SQL

**Type:** Under-abstraction — duplicated SQL in initialization

**Location:** `src/autom8_data/analytics/initialization.py` (lines 150-188)

**Description:** `business_offers_budget` and `payments_budget` CTEs are structurally identical SQL. Kept as separate physical tables to avoid CTE alias collision in split-query execution. A `_validate_budget_spec_consistency` call at startup enforces consistency.

**Ideal resolution:** Parameterize the SQL into a factory function; generate both tables from one source of truth.

**Resolution cost:** Low — contained to `initialization.py`.

---

### TENSION-005: AnalyticsEngine God-Class (2,828 lines)

**Type:** God-class / decomposition in progress

**Location:** `src/autom8_data/analytics/engine.py` (2,828 lines)

**Description:** `AnalyticsEngine` remains 2,828 lines after RD-003 decomposition. Extracted submodules exist (`insight_executor.py`, `initialization.py`, `execution.py`) but the class retains delegation stubs for backward test-patching compatibility (SM-007 deferred).

**Historical reason:** Accumulated query logic, connection management, cache management, and metric resolution over multiple feature phases. Decomposition (RD-003) is underway but gated on test suite migration (SM-007).

**Ideal resolution:** Complete SM-007: migrate test suite off private method patching, then delete delegation stubs and fully extract remaining engine responsibilities.

**Resolution cost:** High — test suite changes required before stubs can be removed.

---

### TENSION-006: analytics/routes Depends on api/* (Layer Violation)

**Type:** Layering violation

**Location:** `src/autom8_data/analytics/routes/` and `src/autom8_data/api/routes/`

**Description:** `analytics.routes.*` imports from `api.auth`, `api.dependencies`, `api.models`, `api.rate_limit` — analytics is Layer 2, API is Layer 4. The routes in `analytics.routes` duplicate the structure of `api.routes` without a clear distinction.

**Ideal resolution:** Move analytics HTTP routes to `api.routes.analytics.*` or establish that `analytics.routes` is explicitly allowed to import from `api.auth`.

**Resolution cost:** Medium — structural reorganization, no logic change.

---

### TENSION-007: Dual Schema System (schema_factory vs manual schemas)

**Type:** Missing abstraction — dual schema system

**Location:** `src/autom8_data/api/schemas/` (appointment.py, business.py, lead.py, payment.py, address.py)

**Description:** Five CRUD schema files carry the comment "Dual schema system consolidation deferred to separate 10x-dev initiative." Both `api.schemas.*` and `api.data_service_models.*` serve schema roles for overlapping entities. The schema_factory pattern (ORM model -> ApiField metadata -> derived schemas) is the intended canonical approach but is not universally applied.

**Ideal resolution:** Consolidate to schema_factory for all CRUD entities; eliminate `api.schemas.*` vs `api.data_service_models.*` split.

**Resolution cost:** Medium — methodical migration, low complexity per entity.

---

### TENSION-008: Appointment Datetime Fields Stored as VARCHAR

**Type:** Naming mismatch — appointment datetime fields stored as VARCHAR

**Location:** `src/autom8_data/core/models/_scheduling.py` (lines 81-89, INV-005)

**Description:** `Appointment.start_datetime` and `end_datetime` are typed as `str` in Python because MySQL stores them as `VARCHAR(40)`. Any code consuming these fields must manually parse ISO 8601 strings.

**Ideal resolution:** MySQL column migration to `DATETIME`, Python type changed to `datetime`.

**Resolution cost:** High — requires MySQL schema migration coordinated across all services reading the table.

---

## Trade-off Documentation

### What was chosen and why

**1. Physical tables over VIEWs for budget CTEs (initialization.py line 133)**
CTEs were materialized as `CREATE TABLE AS` instead of `CREATE VIEW AS` to bypass a DuckDB 1.4.4 query planner crash (NULL pointer dereference in ARM64 environments). This is load-bearing jank — reverting to VIEWs breaks on the current DuckDB version.

**2. Separate DuckDB connections per pool slot (connection_pool.py)**
DuckDB does not support concurrent queries on the same connection. The pool creates N fully independent connections each with their own MySQL ATTACH state. Trade-off: higher startup cost and memory for true isolation.

**3. Import order sensitivity in insights_service.py (HF-7)**
The circular dependency between `analytics.services.insights_service` and `api.services.__init__` was resolved by fixing the import order rather than restructuring modules. `# ruff: noqa: I001` and `# isort: skip_file` prevent auto-formatters from breaking the initialization sequence.

**4. Three-tier dimension resolution (ADR-022)**
Dimension filter overrides use a three-tier lookup (metric-level override -> table-level uniqueness check -> enrichment view fallback) instead of trusting `dim.table` directly. Documented at `query/planner.py:889`: "NOTE: We ALWAYS use three-tier resolution instead of trusting dim.table."

**5. Parquet HYBRID mode (ADR-017)**
The HYBRID routing mode prefers Parquet (551x faster than MySQL ATTACH) with fallback to ATTACH when Parquet data is stale. This creates schema divergence: enriched columns must be computed at both materialization time (Parquet) and connection time (enrichment_views.py). Trade-off: dual implementation of the same enrichment logic.

---

## Abstraction Gap Mapping

### Missing Abstractions

**MISSING-001: Enrichment logic duplicated across materializer.py and enrichment_views.py**

The same business_phone resolution logic, objective_id mapping, and targeting_config JSON extraction exists in two places: `src/autom8_data/analytics/core/infra/materializer.py` (Polars-based, runs at materialization time) and `src/autom8_data/analytics/core/infra/enrichment_views.py` (SQL-based, runs at connection time for the ATTACH path).

**MISSING-002: SQL escaping implemented in multiple places**

`_escape_sql_string_value` exists on `AnalyticsEngine` (engine.py:1160), `_escape_sql_value` exists in `computed_fields.py` (line 43), and IN-clause escaping in `filters.py` (lines 214-218) each implement `str.replace("'", "''")` independently. No shared `sql_escape()` utility.

**MISSING-003: Budget CTE SQL structurally duplicated**

`business_offers_budget` and `payments_budget` tables are created with identical SQL (`src/autom8_data/analytics/initialization.py`:150-188). Startup validates consistency via `_validate_budget_spec_consistency`.

**MISSING-004: normalization.py excluded from shared/__init__.py due to cycle**

`src/autom8_data/analytics/primitives/shared/__init__.py` explicitly does NOT import `normalization.py` to avoid a circular import cycle through `health/models.py`.

### Premature Abstractions

**PREMATURE-001: FrameTypeMapper**

`src/autom8_data/analytics/services/frame_type_mapper.py` provides a frozen-dataclass mapping from `frame_type` to `InsightDefinition` name. It is a hardcoded lookup table dressed as a configurable abstraction.

**PREMATURE-002: AnalyticsBackend Protocol**

`src/autom8_data/analytics/core/infra/backend.py` defines `AnalyticsBackend`, `QueryableConnection`, `ConnectionProvider`, `IbisConnectionAdapter`, `DuckDBPooledConnectionAdapter`, `InMemoryBackend`, and `ProductionBackend` all in one module. The `InMemoryBackend` and `ProductionBackend` serve only to distinguish test from production init paths.

---

## Load-Bearing Code Identification

### LOAD-001: Import order in analytics/services/insights_service.py

**File:** `src/autom8_data/analytics/services/insights_service.py` (lines 1-66)

**Why load-bearing:** The file carries `# ruff: noqa: I001` and `# isort: skip_file`. Any tool that auto-sorts imports will break the module initialization order, causing `ImportError` at startup. Do NOT allow isort, ruff, or pre-commit hooks to reorder these imports.

---

### LOAD-002: IbisConnectionAdapter CTE bypass

**File:** `src/autom8_data/analytics/core/infra/backend.py` (lines 265-270)

**Why load-bearing:** CTE queries (starting with `WITH`) bypass the Ibis `sql()` -> `.to_polars()` path and go directly to `raw_sql().pl()`. This avoids duplicate CTE name errors caused by Ibis wrapping queries in additional WITH clauses. Do NOT remove or merge the `WITH`-check without testing against ATTACH-mode CTE queries.

---

### LOAD-003: Physical table creation order in initialization.py

**File:** `src/autom8_data/analytics/initialization.py` (lines 130-190, 793-820, 887-940)

**Why load-bearing:** Dimension tables and budget tables must be created BEFORE fact tables and joins are registered (per ADR-022). Enrichment VIEWs order also matters. Do NOT reorder `_build_budget_ctes()` or enrichment view creation within `initialize_local_mode()`.

---

### LOAD-004: Three-tier dimension resolution in query/planner.py

**File:** `src/autom8_data/analytics/core/query/planner.py` (lines 889-937)

**Why load-bearing:** The three-tier resolution replaces trusting `dim.table` for all filter/dimension lookups. Bypassing this reintroduces Cartesian product fan-out bugs. Do NOT short-circuit tier resolution by using `dim.table` directly.

---

### LOAD-005: execute_to_polars_safe asyncio lock

**File:** `src/autom8_data/analytics/core/infra/backend.py` (lines 272-282)

**Why load-bearing:** DuckDB rejects concurrent queries on the same connection with NULL pointer dereference. The asyncio lock serializes access during split-query parallel execution (SPIKE-duckdb-concurrency-crash). Do NOT remove the `_query_lock` from `IbisConnectionAdapter`.

---

### LOAD-006: Operational mode FSM module-level singleton

**File:** `src/autom8_data/analytics/core/domain/operational_mode.py` (lines 204-210)

**Why load-bearing:** Mode state is stored in module-level globals. The `hydrate()` function must be called during FastAPI lifespan startup. Do NOT attempt to make `OperationalMode` state instance-level without coordinating with all consumers.

---

### LOAD-007: AnalyticsEngine delegation stubs (SM-007 deferred)

**File:** `src/autom8_data/analytics/engine.py` (lines 2305-2644)

**Why load-bearing:** Methods like `_execute_query`, `_apply_rolling_aggregation`, `_is_composite_metric`, `_resolve_date_range` are delegation stubs retained for test patching compatibility. Do NOT delete delegation stubs until SM-007 (test migration) is complete.

---

## Evolution Constraint Documentation

### Changeability Ratings

| Component | Rating | Reason |
|-----------|--------|--------|
| `ConnectionRouter._determine_backend()` | coordinated | Multi-callers; circuit breaker, ATTACH fallback |
| `FactTableResolver` / `QueryPlan` | coordinated | Flows through builder, executor, planner |
| `MaterializationRegistry` | coordinated | Single source of truth for table config |
| `_parse_single_filter` | frozen | SQL injection boundary |
| `engine.get()` kwargs tunnel | migration | InsightExecutor and InsightsService depend on it |
| `analytics/routes/*.py` | coordinated | Bidirectional coupling with `api/` |
| `BatchInsightExecutor` endpoint | migration | External callers; legacy metric instruments active |
| `analytics/core/models/protocols.py` | safe | Not currently wired, safe to remove or wire |
| `analytics/core/metrics/library.py` | coordinated | Single file containing all metric defs |
| `analytics/core/infra/materializer.py` | coordinated | Error types, manifest models, sync job |
| `InsightDefinition` schema | frozen | `frozen=True` Pydantic models; field changes require updating all constructors in `insights/library.py` |
| `api.services.*` shim layer | migration | External callers depend on these import paths |
| Budget CTE physical tables | frozen (DuckDB 1.4.4) | Cannot revert to VIEWs on current DuckDB version |
| `start_datetime` VARCHAR columns | migration | MySQL schema change required before Python model update |

### Deprecated / In-Progress Migrations

- **DuckDB ATTACH mode**: Marked as "legacy" in `core/config.py`. HYBRID mode is the current production path.
- **Legacy batch execution** (`BatchInsightExecutor`): Marked legacy in metrics. `UnifiedBatchExecutor` is the successor but the legacy endpoint remains live.
- **engine.py god-class decomposition (RD-003)**: Phase 1 extractions done; protocols defined but not wired. No Phase 2 TDD found.
- **gid_map_service TODO**: `api/services/gid_map_service.py:87` has `TODO: Replace with real autom8_asana integration`.
- **data_service TODO**: `analytics/routes/data_service.py:121` has incomplete migration TODO.

### External Dependency Constraints

- `autom8y_telemetry` (`trace_computation`) — imported in `engine.py` and `insight_executor.py`
- `autom8y_config` (`Autom8yEnvironment`) — imported in `engine.py`. Environment enum values used in routing decisions
- MySQL ATTACH: Production HYBRID mode depends on MySQL being reachable for fallback

---

## Risk Zone Mapping

### RISK-001: SQL Filter Values Bypass Parameterization

**File:** `src/autom8_data/analytics/core/query/filters.py` (lines 214-277), `src/autom8_data/analytics/insights/computed_fields.py` (line 43), `src/autom8_data/analytics/engine.py` (line 1160)

Filter values are SQL-literal-injected via string formatting. Escaping uses single-quote doubling (`replace("'", "''")`). Three separate implementations exist with no centralized guard. New call sites that add string interpolation without escaping bypass the pattern.

---

### RISK-002: Lead List Endpoint Guard at Service Layer Only

**File:** `src/autom8_data/services/lead.py` (lines 460-484)

`LeadService.list()` requires at least one of `office_phone` or `phone` to prevent returning all 300K+ leads. The guard exists in the service layer but not at the HTTP route layer. If a route refactoring bypasses the service, the guard is missing.

---

### RISK-003: DuckDB Concurrency Crash If connection_provider Passed to Inner Orchestrator

**File:** `src/autom8_data/analytics/engine.py` (lines 2343-2359)

The constraint is enforced by passing `connection_provider=None` explicitly. No assertion or type-level enforcement prevents accidentally passing the provider. The only guard is the inline comment.

---

### RISK-004: OperationalMode Redis Write Failure Silently Swallowed

**File:** `src/autom8_data/analytics/core/domain/operational_mode.py` (lines 448-453)

`set_mode()` attempts Redis persistence and sets `persisted = False` on failure. The transition still succeeds (in-memory state is updated). No alerting or circuit-breaker is wired to the `persisted=False` case.

---

### RISK-005: ATTACH Mode Cannot Set access_mode = READ_ONLY

**File:** `src/autom8_data/analytics/core/infra/connection_pool.py` (lines 315-321)

DuckDB rejects `SET access_mode = 'READ_ONLY'` when ATTACH is active. Defense-in-depth relies on: SELECT-only MySQL user, route allowlisting, PII masking middleware, and WAF SQLi rules. No database-level read-only guarantee for ATTACH connections.

---

### RISK-006: Shared Metric List Across Three Insight Definitions

**File:** `src/autom8_data/analytics/insights/library.py` (lines 23-26)

`_ADSET_AD_BASE_METRICS` list shared across `adset_level_stats`, `ad_level_stats`, and `demographic_stats`. No static analysis or test enforces that the three insights have identical metrics. A future edit that appends a metric silently changes all three.

---

## Knowledge Gaps

1. **External callers of `api.services.*` shims**: Cannot enumerate shim deletion risk without cross-org import audit.
2. **Full ADR corpus not inspected**: ADR references (ADR-001 through ADR-WS4-5) are referenced in code comments but actual ADR files were not located in the repository.
3. **Test suite patching inventory (SM-007)**: Which tests patch `AnalyticsEngine` private methods is unknown without inspecting the test directory.
4. **SPIKE-duckdb-concurrency-crash document**: Referenced in engine.py but not found in the repo.
5. **Proto / gRPC layer constraints**: `proto/` and `grpc/` packages were not deeply inspected.
