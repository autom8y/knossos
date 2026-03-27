---
domain: design-constraints
generated_at: "2026-03-18T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "51f5e8d"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog Completeness

### TENSION-001: AnalyticsEngine God-Class Decomposition In Progress

**Files**: `src/autom8_data/analytics/engine.py` (2,377 lines, ~40 methods on `AnalyticsEngine`)

`AnalyticsEngine` was a single god-class responsible for: request translation, query building, connection management, cache management, rolling aggregation, composite calculation, insight execution, and result formatting. A decomposition (labelled "RD-003") has been ongoing — multiple modules carry the comment "Extracted from engine.py as part of the god-class decomposition (Phase 1)":

- `analytics/core/execution/orchestrator.py` — `QueryOrchestrator`
- `analytics/core/execution/rolling.py` — `RollingAggregator`
- `analytics/core/execution/composite.py` — `CompositeCalculator`
- `analytics/core/request/translator.py` — `RequestTranslator`
- `analytics/core/request/time_resolver.py` — time resolution
- `analytics/core/output/processor.py` — output formatting

Despite the extraction, `engine.py` remains 2,377 lines. The extracted classes each define a `Protocol` in `analytics/core/models/protocols.py` (Phase 1 decomposition protocols: `QueryOrchestratorProtocol`, `RollingAggregatorProtocol`, `CompositeCalculatorProtocol`, `RequestTranslatorProtocol`, `TimeResolverProtocol`), but **none of those protocols are imported or referenced outside** `protocols.py` itself — `engine.py` continues to hold the orchestration logic inline, and the protocols currently serve only as documentation artifacts.

**Current state**: Partial decomposition. Extracted helpers exist but engine.py still orchestrates them directly without going through the protocol interfaces.

**Ideal state**: Each phase-1 protocol has exactly one implementation wired through dependency injection in `AnalyticsEngine.__init__`.

**Why current state persists**: Decomposition was halted mid-phase; extracting modules reduces size of collaborating code but not the engine itself.

---

### TENSION-002: Dual Batch Insight Execution Paths

**Files**:
- `src/autom8_data/analytics/services/batch_insight_executor.py` (505 lines, `BatchInsightExecutor`)
- `src/autom8_data/analytics/services/unified_batch_executor.py` (764 lines, `UnifiedBatchExecutor`)

Two distinct batch execution paths exist side-by-side, both mounted at different API endpoints in the same route file:

```python
# insights.py line 292
executor = BatchInsightExecutor(engine=engine)   # /insights/{name}/execute/batch

# insights.py line 342
executor = UnifiedBatchExecutor(...)              # /insights/{name}/batch
```

The `BatchInsightExecutor` is tagged "legacy" in the metrics registry (`# Batch Insight Metrics (legacy)` at `api/metrics/__init__.py:60`). `UnifiedBatchExecutor` is the intended successor (TDD-unified-batch-analytics).

**Current state**: Both paths alive in production; legacy path not removed.

**Ideal state**: Single batch execution path via `UnifiedBatchExecutor`; legacy endpoint removed or aliased.

**Why persists**: Backward-compatible migration — callers of the legacy endpoint exist outside this codebase.

---

### TENSION-003: Dual Translator Naming Collision

**Files**:
- `src/autom8_data/analytics/core/infra/translator.py` — `SemanticTranslator` (business terms → technical columns)
- `src/autom8_data/analytics/core/request/translator.py` — `RequestTranslator` (wraps SemanticTranslator, applies defaults)

Two files named `translator.py` at different layers. The `RequestTranslator` wraps `SemanticTranslator` but both are "translators." Any newcomer encountering one must verify which `translator` is in use. The dependency comment in `request/translator.py` explicitly states: "This module may import from infra/translator" — meaning the dependency is known but not structurally enforced.

**Current state**: Two translators with overlapping names; one wraps the other.

**Ideal state**: Either merge into a single translation service or rename to clarify the layering (`InfraSemanticTranslator` vs `RequestNormalizationTranslator`).

---

### TENSION-004: analytics/routes Lives Outside api/routes But Imports From It

**Files**: All of `src/autom8_data/analytics/routes/`

The analytics routes package (`analytics/routes/`) imports from `autom8_data.api.*` (JWT auth, dependencies, rate limiting, response models). Simultaneously, `api/routes/__init__.py` imports the analytics routers and re-exports them as if they were `api/routes` modules. This means:

- `analytics/routes/` depends on `api/` (upward dependency from domain layer to API layer)
- `api/routes/__init__.py` depends on `analytics/routes/` (downward dependency)

The layer model documented in `engine.py` header states "Core -> Domain -> Application -> API," yet the analytics routes bypass this by directly coupling to `api.auth`, `api.rate_limit`, and `api.models`.

**Current state**: Bidirectional coupling between `analytics/` and `api/`.

**Ideal state**: Analytics routes should be moved entirely into `api/routes/` with shared dependencies injected, not imported directly from `api/`.

---

### TENSION-005: **kwargs Hidden Parameter Tunnel Through engine.get()

**File**: `src/autom8_data/analytics/engine.py` (lines 731-750)

`AnalyticsEngine.get()` accepts `**kwargs` and pops internal control parameters from it:

```python
insight_cache_ttl: int = kwargs.pop("_insight_cache_ttl", 300)
freshness_sla_minutes: int | None = kwargs.pop("_freshness_sla_minutes", None)
grouping_dimensions: list[str] | None = kwargs.pop("_grouping_dimensions", None)
# also: _skip_grain_validation
```

These are effectively private API parameters tunneled through the public interface using underscore-prefix conventions in dict keys. The parameters originate in `InsightExecutor` and `InsightsService`. This breaks type safety: callers can pass unsupported kwargs silently.

**Current state**: Engine uses kwargs sentinel pop for insight-context overrides.

**Ideal state**: Explicit `AnalyticsRequest` model with typed optional fields for all control parameters.

---

### TENSION-006: `execution.py` Package Naming Mismatch

**File**: `src/autom8_data/analytics/execution.py`

This file is named `execution.py` but its actual responsibility is **connection routing** (`RoutingConnectionProvider`, `DirectConnectionProvider`, `execute_sql_with_pool`, `execute_sql_with_router`). It was "Extracted from engine.py as part of RD-003 engine decomposition." The execution subpackage at `analytics/core/execution/` handles actual query execution. The naming collision requires mental disambiguation on every import.

**Current state**: `execution.py` = connection provider; `core/execution/` = query orchestration.

**Ideal state**: Rename `execution.py` to `connection_providers.py` or merge into `core/infra/`.

---

### TENSION-007: `analytics/core/metrics/library.py` God-File (3,957 lines)

**File**: `src/autom8_data/analytics/core/metrics/library.py`

A single 3,957-line module contains all metric definitions for all verticals and contexts. Every new metric registration requires editing this file. No class structure separates metric domains. There is a separate `analytics/insights/library.py` (1,199 lines) for insight definitions — a structural parallel that works well — but no equivalent decomposition for the metrics library.

**Current state**: Monolithic metric registry in one file.

**Ideal state**: Per-vertical or per-domain metric modules, registered via a discovery mechanism.

---

### TENSION-008: `materializer.py` God-File (3,038 lines)

**File**: `src/autom8_data/analytics/core/infra/materializer.py`

The largest file in `infra/` at 3,038 lines. Contains: error types, manifest models (`TableManifest`, `SyncManifest`), sync configuration (`TableSyncConfig`, `RIJoinPair`, `CoverageColumn`), and the `MaterializationJob` class (the full sync orchestrator). The models defined here (e.g., `SyncManifest`) are referenced from `scheduler/_materialization.py` and the admin routes — tight coupling across concerns.

---

## Trade-off Documentation

### TENSION-001 Trade-off

**Current state justification**: The partial decomposition was done incrementally to avoid big-bang rewrites. Each extracted module can be tested independently. The engine's 2,377 lines are smaller than the pre-decomposition size.

**Why ideal state hasn't been reached**: Protocol-based wiring requires changing call sites inside engine.py. This is a coordinated multi-file change (engine, initialization, all protocols) that likely requires a dedicated work stream.

**Attempted before**: The "Phase 1" extraction is documented. No evidence of a Phase 2 ADR or TDD, suggesting the effort stalled.

**ADR reference**: `TDD-RD-003-engine-decomposition`, `ADR-RD-003-B` (referenced in engine.py and execution.py headers but not present in `.ledge/`).

---

### TENSION-002 Trade-off

**Why legacy path persists**: `BatchInsightExecutor` serves a `/insights/{name}/execute/batch` endpoint with a `BatchInsightExecuteRequest` model. External clients depend on this schema. Migration requires coordinated client-side changes.

**External constraint**: The `api/metrics/__init__.py` explicitly labels the old metrics as "(legacy)" which suggests awareness but no migration timeline.

---

### TENSION-004 Trade-off

**Why it persists**: The `analytics/` subdomain grew its own routing as the analytics surface expanded. Routing was co-located with logic for development speed. The bidirectional coupling was tolerated because the API layer exists in the same codebase.

**Risk**: Any change to `api.auth.jwt`, `api.rate_limit`, or `api.models` requires auditing `analytics/routes/` as a dependent. This is documented in ADR-FSH-001 and ADR-FSH-002 which reference filter translation consolidation — both in `.ledge/decisions/`.

---

### TENSION-005 Trade-off

**Why it exists**: `engine.get()` is the public API surface for both direct callers and insight-context callers. Adding explicit typed parameters for every insight-context override would have required a breaking API change or method overloading. The kwargs tunnel was chosen as a non-breaking approach.

**External constraint**: The engine public API is used by callers outside this file; changing the signature requires coordinating all call sites.

---

## Abstraction Gap Mapping

### Gap 1: No `AnalyticsRequest` Model as First-Class Entry Point

The `engine.get()` signature accepts 12+ positional/keyword parameters plus `**kwargs`. A corresponding `AnalyticsRequest` model exists in `analytics/core/request/models.py`, but it is an internal model created midway through the pipeline — not a public-facing contract at the `engine.get()` boundary. Callers construct the request parameter by parameter rather than constructing a request object.

**Duplicated pattern**: Both `InsightDefinition.validate_grouping_dimensions()` (in `insights/models.py`) and `AnalyticsRequest.validate_grouping_dimensions_subset()` (in `core/request/models.py` line 355) implement the same validation logic. The comment at line 358 explicitly notes: "Mirrors InsightDefinition.validate_grouping_dimensions but applied [at request level]."

---

### Gap 2: Phase 1 Protocols Without Implementations (Zombie Protocol Layer)

**File**: `src/autom8_data/analytics/core/models/protocols.py`

Five protocols defined for the Phase 1 engine decomposition (`QueryOrchestratorProtocol`, `RollingAggregatorProtocol`, `CompositeCalculatorProtocol`, `RequestTranslatorProtocol`, `TimeResolverProtocol`) have zero usages outside `protocols.py` itself. The concrete implementations (e.g., `QueryOrchestrator`, `RollingAggregator`) exist but are not typed as conforming to their protocols at any injection point.

**Maintenance burden**: The protocols are maintained in sync with the classes they were designed to describe, but provide no runtime or static typing benefit since they are not referenced.

---

### Gap 3: Duplicate Batch Metric Tracking

**File**: `src/autom8_data/api/metrics/__init__.py`

Two parallel sets of Prometheus metrics exist:
- Legacy batch metrics: `BATCH_DURATION_HISTOGRAM`, `BATCH_SIZE_HISTOGRAM`
- Unified batch metrics: `BATCH_INSIGHT_DURATION`, `BATCH_INSIGHT_SIZE`, `BATCH_INSIGHT_CACHE_HITS`, `BATCH_INSIGHT_CACHE_MISSES`

Both sets are registered at startup. The legacy set is no longer connected to a live code path that populates it (since `UnifiedBatchExecutor` populates the new metrics), creating dead telemetry.

---

### Gap 4: Over-Engineered `analytics/core/infra/` Dependency Boundary

**Files**: All of `src/autom8_data/analytics/core/infra/`

The `infra/` package declares itself a zero-dependency foundation layer ("must NOT import from other core/ packages"). However, `infra/validation.py` imports from `analytics/core/models/` (lines 31, 182, 339), all guarded by `TYPE_CHECKING`. This means the rule holds at runtime but is violated at the type-checking layer. The rule as documented cannot be enforced by importlinter or similar tools without special TYPE_CHECKING handling.

Similarly, `analytics/core/metrics/library.py` imports from `analytics/core/query/window_metric_sql` (line 1181), which the rule in `metrics/__init__.py` says "must NOT import from query/." This is a live (not TYPE_CHECKING guarded) rule violation.

---

## Load-Bearing Code Identification

### LOAD-001: `ConnectionRouter` in `analytics/core/infra/connection_router.py`

**File**: `src/autom8_data/analytics/core/infra/connection_router.py` (1,366 lines)

Every analytics query path goes through `ConnectionRouter._determine_backend()` to select between Parquet and MySQL ATTACH. The router implements the consumer-side circuit breaker (`ConsumerCircuitBreaker`), fallback reason tracking (`FallbackReason`), per-table freshness checks, and stale parquet resolution. It is imported in:
- `api/main.py` (startup, shutdown)
- `api/dependencies.py`
- `analytics/engine.py`
- `analytics/insight_executor.py`

**What a naive fix would break**: Changing the `_determine_backend()` return type or the `RouterBackend` enum values would break all routing logic. Changing the circuit breaker threshold logic would affect production availability decisions. The fallback to ATTACH is the production safety net for stale parquet — removing it without replacing it would cause query failures when parquet materializer is behind.

**Changeability rating**: `coordinated` — changes require simultaneous updates to engine initialization, admin reset endpoints, and metrics gauges.

---

### LOAD-002: `FactTableResolver` + `QueryPlan` in `analytics/core/query/fact_resolver.py`

**File**: `src/autom8_data/analytics/core/query/fact_resolver.py`

`FactTableResolver` resolves which DuckDB fact tables to JOIN for a given set of metrics. `QueryPlan` is the output — the canonical data structure that flows from resolver → `QueryBuilder` → `QueryExecutor`. Both are imported in:
- `analytics/core/query/builder.py` (line 63)
- `analytics/core/query/executor.py` (line 50, TYPE_CHECKING)
- `analytics/core/query/planner.py` (line 30, TYPE_CHECKING)

**What a naive fix would break**: Changing the `QueryPlan` fields or the resolver's table-selection logic would break: the builder's split-merge behavior, the executor's connection routing per table, and the planner's filter reachability analysis. These are tightly coupled across 3 files in the same package.

**Changeability rating**: `coordinated`.

---

### LOAD-003: `MaterializationRegistry` in `analytics/core/infra/registry.py`

**File**: `src/autom8_data/analytics/core/infra/registry.py` (594 lines)

`MaterializationRegistry` is the single source of truth for which tables are registered for materialization, their sync states, and config. It is initialized at engine startup and consulted by:
- `MaterializationJob` (materializer)
- `ConnectionRouter` (freshness checks)
- Admin endpoints (status reporting)
- `scheduler/_materialization.py`

**Changeability rating**: `coordinated` — all table registration code and freshness thresholds flow through here.

---

### LOAD-004: `_parse_single_filter` in `analytics/core/query/filters.py`

**File**: `src/autom8_data/analytics/core/query/filters.py` (line 229)

This function is the SQL injection boundary for all filter inputs. It uses single-quote escaping (`str.replace("'", "''")`) to produce SQL literal values. All user-supplied filter values traverse this function before entering SQL.

**What a naive fix would break**: Any refactoring that removes the quote escaping or adds a code path that bypasses it would introduce SQL injection. The function is called from `parse_filters`, `parse_filters_with_column_map`, and `_parse_in_clause`.

**Changeability rating**: `frozen` — changes require security review.

---

### LOAD-005: `engine.get()` **kwargs sentinel tunnel

**File**: `src/autom8_data/analytics/engine.py` (lines 731-750)

The kwargs pop pattern for `_insight_cache_ttl`, `_freshness_sla_minutes`, `_grouping_dimensions`, `_skip_grain_validation` is load-bearing in the sense that `InsightExecutor` and `InsightsService` depend on passing these through `**kwargs`. Changing `engine.get()` to refuse unknown kwargs would break insight execution silently (the pops would simply not find their keys, falling back to defaults).

**Changeability rating**: `migration` — removing this requires first adding explicit parameters to `engine.get()`, then updating all callers.

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
| `analytics/core/models/protocols.py` (Phase 1 protocols) | safe | Not currently wired, safe to remove or wire |
| `analytics/core/metrics/library.py` | coordinated | Single file containing all metric defs |
| `analytics/core/infra/materializer.py` | coordinated | Error types, manifest models, sync job |

### Deprecated / In-Progress Migrations

- **DuckDB ATTACH mode**: Marked as "legacy" in `core/config.py`. HYBRID mode (Parquet-first, ATTACH fallback) is the current production path. ATTACH-only mode is nominally available but the circuit breaker protects against relying on it.
- **Legacy batch execution** (`BatchInsightExecutor`): Marked legacy in metrics. `UnifiedBatchExecutor` is the successor but the legacy endpoint remains live.
- **engine.py god-class decomposition (RD-003)**: Phase 1 extractions done; protocols defined but not wired. No Phase 2 TDD found in `.ledge/specs/`.
- **gid_map_service TODO**: `api/services/gid_map_service.py:87` has `TODO: Replace with real autom8_asana integration per TDD-AUTOM8-DATA-API-001`. Referenced TDD not present in `.ledge/`.
- **data_service TODO**: `analytics/routes/data_service.py:121` has `TODO: Replace with real local-storage-backed client when Entity` (truncated) — incomplete migration.

### External Dependency Constraints

- `autom8y_telemetry` (`trace_computation`) — imported in `engine.py` and `insight_executor.py`. Removing or renaming would require coordinated change with this external package.
- `autom8y_config` (`Autom8yEnvironment`) — imported in `engine.py`. Environment enum values used in routing decisions.
- MySQL ATTACH: Production HYBRID mode depends on MySQL being reachable for fallback. Removing ATTACH support is a migration, not a deletion.

---

## Risk Zone Mapping

### RISK-001: Filter Values Bypass Parameterization

**File**: `src/autom8_data/analytics/core/query/filters.py` (lines 218-277)

Filter values are SQL-literal-injected via string formatting (`f"{qualified_column} {operator} {formatted_value}"`). Escaping uses single-quote doubling (`replace("'", "''")`) which is correct for string literals but relies on the calling code correctly classifying values as "string" vs "numeric." The classification happens at lines 218-225 based on operator parsing — not on column type. A filter value that looks numeric (e.g., `">0.5"`) bypasses quoting; if a string value somehow passes the numeric check, it would be unquoted.

**Risk**: Partial injection protection; type classification is by convention not schema enforcement.

---

### RISK-002: `computed_fields.py` Bare Exception Silencing

**File**: `src/autom8_data/analytics/insights/computed_fields.py` (lines 713-714)

```python
except Exception:
    pass
```

A bare `except Exception: pass` in the computed field processor. When a computed field calculation fails, the failure is silently swallowed. The output metric will be absent or carry a stale/default value. This is unguarded — there is no logging, no counter increment, no metric that fires to indicate a silent computation failure.

**Risk**: Computed field failures produce incorrect results without any observability signal.

---

### RISK-003: `enrichment_views.py` Continues Past Failures

**File**: `src/autom8_data/analytics/core/infra/enrichment_views.py` (lines 438-443)

View creation failures log via `logger.exception` but the loop continues. If an enrichment view fails to create, queries that depend on that view will fail at query time with a "table not found" error, not at initialization time. The failure source is an initialization-time view creation error but the symptom appears at query time.

**Risk**: Query-time failures with opaque error messages when initialization silently fails.

---

### RISK-004: `ConnectionRouter` ATTACH Probe Swallows All Exceptions

**File**: `src/autom8_data/analytics/core/infra/connection_router.py` (lines 833-836, 862-866)

Two `except Exception` blocks in the ATTACH availability probe and circuit-breaker suppression check log warnings and continue. A router in HYBRID mode will treat an unresolvable probe failure identically to "ATTACH unavailable" — falling back to Parquet. This is correct behavior but the exception source is never re-raised or stored, making it invisible in the structured logs unless the log aggregation picks up the warning.

**Risk**: ATTACH availability determination can fail silently; circuit breaker suppression check failure is non-fatal but untracked.

---

### RISK-005: `analytics/services/insights_service.py` Validates Phone Format Late

**File**: `src/autom8_data/analytics/services/insights_service.py` (line 384, `_validate_phone_formats`)

Phone format validation is a private method called during service execution, not at the FastAPI model level. Malformed phone numbers reach the service layer before rejection. The `api/routes/data_service.py` endpoint passes phones through without schema-level format enforcement. Error messages are generated internally at service level.

**Risk**: Invalid phone numbers pass route-layer validation; format errors surface as service-layer exceptions rather than 422 responses.

---

## Knowledge Gaps

1. **TDD-RD-003-engine-decomposition and ADR-RD-003-B** are referenced in source comments (`engine.py`, `execution.py`, `core/execution/__init__.py`) but not present in `.ledge/`. Unable to verify the intended Phase 2 scope.

2. **TDD-AUTOM8-DATA-API-001** referenced in `api/services/gid_map_service.py` — not present in `.ledge/specs/`. Cannot determine migration timeline or scope.

3. **`analytics/core/metrics/library.py` metric count** — the file is 3,957 lines with no visible class structure. The count of registered metrics and their vertical breakdown was not catalogued in this audit.

4. **`materializer.py` scheduler integration** — the relationship between `MaterializationJob` and `scheduler/_materialization.py` was not fully traced. The scheduler package's internal decomposition was observed but not fully read.

5. **Import-linter / enforcement** — no `importlinter` or `tach` configuration was found in the project root. The documented `DEPENDENCY RULE` comments are convention-only with no automated enforcement.
