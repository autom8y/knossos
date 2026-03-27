---
domain: conventions
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

# Codebase Conventions

## Error Handling Style

### Two Separate Error Hierarchies

The codebase maintains two parallel, domain-specific error hierarchies. These are NOT ad-hoc — every service layer and semantic layer exception descends from one of these roots.

**Service Layer (`src/autom8_data/services/base.py`)**

Root: `ServiceError(Exception)` with dual gRPC+HTTP status mappings baked into the class:

```
ServiceError
├── NotFoundError          grpc=NOT_FOUND, http=404, code="RESOURCE_NOT_FOUND"
├── AlreadyExistsError     grpc=ALREADY_EXISTS, http=409, code="RESOURCE_ALREADY_EXISTS"
├── FKValidationError      grpc=FAILED_PRECONDITION, http=400, code="FK_VALIDATION_FAILED"
├── ImmutableFieldError    grpc=INVALID_ARGUMENT, http=400, code="IMMUTABLE_FIELD"
├── ValidationError        grpc=INVALID_ARGUMENT, http=400, code="VALIDATION_ERROR"
├── DatabaseUnavailableError grpc=UNAVAILABLE, http=503, code="DATABASE_UNAVAILABLE"
├── InvalidFilterError     grpc=INVALID_ARGUMENT, http=400
└── InvalidPeriodError
```

Every `ServiceError` subclass carries: `grpc_status` (class var), `http_status` (class var), `code` (class var), `.message`, `.field` (optional). The `to_error_detail()` method returns the canonical HTTP error dict.

**Analytics/Semantic Layer (`src/autom8_data/analytics/core/infra/exceptions.py`)**

Root: `SemanticError(Exception)` for all analytics pipeline errors:

```
SemanticError
├── JoinPathError                  # Cannot find join path; includes available_tables
├── JoinPathTooLongError           # Exceeds max depth; includes path_tables
├── DimensionNotFoundError         # Fuzzy-matched suggestions via rapidfuzz
├── MetricNotFoundError            # Fuzzy-matched suggestions via rapidfuzz
├── JoinNotFoundError
├── SchemaValidationError          # Aggregates multiple ValidationError dataclasses
├── GrainIncompatibleError         # Metric unreachable from dimension via join graph
├── DimensionInconsistencyError    # Same dimension resolves differently across split queries
├── ParquetRequiredError           # Insight requires LOCAL backend; ATTACH fallback unavailable
├── QueryTimeoutError              # DuckDB query exceeded hard timeout
├── CartesianRiskError             # Fact-to-fact join would inflate metrics (SCAR-001)
├── InvalidColumnError             # Column not found in result (order_by)
└── LookbackExceededError          # start_date exceeds Parquet lookback window (LOCAL mode)
```

Sub-hierarchies extend `SemanticError`:
- `src/autom8_data/analytics/core/registry/dependency.py`: `DependencyResolutionError > CircularDependencyError, DependencyDepthError, MissingDependencyError`
- `src/autom8_data/analytics/core/infra/materializer.py`: `MaterializationError > ConfigurationError, DataQualityError, SyncError`
- `src/autom8_data/analytics/core/infra/scheduler/_exceptions.py`: `SchedulerError > SchedulerNotAvailableError, SchedulerAlreadyRunningError, SchedulerNotRunningError, CircuitOpenError`
- `src/autom8_data/analytics/core/joins/graph.py`: `CycleDetectedError(SemanticError)`

**Fuzzy-match suggestions**: `DimensionNotFoundError` and `MetricNotFoundError` override `__str__()` to call `rapidfuzz`-based `get_suggestions()` (imported lazily from `validation.py`). This pattern avoids circular imports and is project-specific — new "not found" errors for user-facing types should follow this pattern.

**Auth Layer (`src/autom8_data/api/auth/exceptions.py`)**

Separate hierarchy not inheriting from either root: `AuthenticationError`, `AuthorizationError`, `KeyConfigurationError`, `AuthServiceUnavailableError`.

### Error Creation Pattern

Standard Python `raise ExactErrorType(structured_args)` — no sentinel errors, no error codes as integers. All exceptions carry structured fields (from_table, to_table, reason, available_tables, etc.) built at construction time. Message text is formatted in `__init__` and passed to `super().__init__(message)`.

### Error Propagation

- Services raise typed exceptions; callers do not catch-all and re-raise (propagate directly)
- Analytics engine propagates `SemanticError` up to API boundary
- API boundary maps errors to HTTP via registered exception handlers (`src/autom8_data/api/errors.py`, `register_exception_handlers()`)

### Error Handling at API Boundary

`src/autom8_data/api/errors.py` contains a centralized `register_exception_handlers()` function that registers per-type FastAPI exception handlers. This is the canonical pattern for adding new error types to the API surface. Each handler:
1. Extracts `request_id` from `request.state`
2. Calls `record_api_error(code, status)` for Prometheus metrics
3. Logs with `extra={"event_type": ..., "request_id": ...}` (structured, not string-formatted)
4. Returns `JSONResponse(status_code, content=ErrorResponse(...).model_dump())`

The catch-all `generic_error_handler` uses `logger.exception(...)` (preserves traceback in logs) and returns generic 500 — never exposes stack traces to clients.

### Error Checking Convention

No inline error suppression (`try/except: pass`). 476 `raise` sites across 120 files; 881 `try/except` blocks across 124 files. This ratio indicates substantial structured exception use without over-suppression.

### Logging

Two logging entry points:
- `src/autom8_data/core/logging.py` — `get_module_logger(__name__)` returns a `_LazyLogger` (deferred SDK init). Used in 146 files across all packages.
- `src/autom8_data/analytics/core/infra/logging.py` — `get_json_logger(__name__)` used only in analytics-internal code (2 files). Provides `QueryContext` dataclass for structured query logging.

The `_LazyLogger` wraps `autom8y_log.get_logger()` and supports `.bind()` for adding contextual fields. The standard pattern at module level:

```python
from autom8_data.core.logging import get_module_logger
logger = get_module_logger(__name__)
```

Log calls use **keyword arguments for structured data** (not string formatting), consistent with structured logging: `logger.info("event_name", extra={"key": "value"})`. F-strings are permitted per ruff config (G004 suppressed), but the SDK-structured style (`key=value`) is preferred in application code.

Direct `loguru` or `structlog` imports are **banned** (enforced via `ruff` `TID251`). Use `autom8y_log` SDK only.

---

## File Organization

### Top-Level Package Structure

```
src/autom8_data/
├── __init__.py
├── analytics/          # Semantic analytics layer (DuckDB, Polars, metrics/dimensions/insights)
│   ├── core/           # Engine internals: models, query, execution, infra, joins, registry, metrics
│   ├── insights/       # InsightDefinition library, registry, post-processors
│   ├── primitives/     # Domain-specific analytics modules (health, pacing, anomaly, etc.)
│   ├── services/       # Analytics service layer (cache, batch execution)
│   ├── routes/         # FastAPI analytics routes
│   ├── clients/        # External analytics API clients
│   ├── fixtures/       # Test fixture factories for analytics models
│   └── entity_source/  # Protocol and filter for entity data source
├── api/                # FastAPI CRUD API layer
│   ├── routes/         # HTTP route handlers (one file per resource)
│   ├── data_service_models/ # Pydantic models for data service (prefixed with _)
│   ├── schemas/        # Input/output Pydantic schemas (per-resource)
│   ├── auth/           # Authentication/authorization
│   ├── services/       # API service layer (gid_map, period translation, etc.)
│   ├── clients/        # API-layer external clients
│   └── metrics/        # Prometheus metrics for API layer
├── core/               # Shared infrastructure (config, models, repositories, filters, logging)
│   ├── models/         # SQLModel ORM models (grouped by domain: _advertising, _scheduling, etc.)
│   ├── repositories/   # Database repositories (one file per entity or concern)
│   ├── filters/        # Filter DSL (types, builder, introspection)
│   └── schemas/        # Core-layer Pydantic schemas
├── services/           # Entity service layer (one file per entity, e.g., lead.py, appointment.py)
├── grpc/               # gRPC protocol layer
│   ├── handlers/       # gRPC service handlers (one per entity)
│   └── adapters/       # gRPC <-> SQLModel adapters (one per entity)
├── proto/              # Generated protobuf code (DO NOT edit manually)
├── utils/              # Cross-cutting utilities (e.g., phone_normalizer)
└── clients/            # Top-level external client adapters
```

### File Naming Conventions

**Underscore-prefixed files (`_name.py`)**: Internal/private modules not intended for direct public import. Extensively used in:
- `api/data_service_models/_lead.py`, `_business.py`, etc. — one Pydantic model set per entity, prefixed with `_`, re-exported from `__init__.py`
- `analytics/core/infra/scheduler/_base.py`, `_exceptions.py`, `_retry.py`, `_benchmark.py`, `_registry.py`, `_materialization.py`, `_circuit_breaker.py` — internal scheduler components

**`_crud.py` suffix**: Route handler files follow `{plural_entity}_crud.py` pattern (e.g., `leads_crud.py`, `appointments_crud.py`, `ad_accounts_crud.py`). Non-CRUD routes use descriptive names (`admin.py`, `health.py`, `analytics_health.py`).

**`library.py`**: Large, flat registration-style modules. Examples: `analytics/core/metrics/library.py` (metric definitions), `analytics/insights/library.py` (insight definitions).

**`models.py` / `types.py`**: Pure model/type definitions. Used within `core/` and per-subpackage in analytics. Note: multiple files within the same package may exist (e.g., `api/models.py`, `api/models_batch.py`, `api/models_health.py`), distinguished by semantic suffixes.

**`constants.py`**: Module-level constants. Found at `analytics/core/infra/constants.py` and `analytics/core/constants.py`.

**`validators.py`**: Standalone validation helpers (e.g., `core/validators.py`, `services/validators.py`).

**`decorators.py`**: Decorator definitions isolated into their own file (e.g., `analytics/core/metrics/decorators.py`).

### Per-Package File Organization

Within `analytics/core/`, each sub-package has a clear concern:
- `models/` — dataclasses, enums; no business logic
- `infra/` — connection pools, schedulers, materialization, logging, caching
- `query/` — SQL generation, planning, execution, filtering
- `execution/` — query orchestration strategies (rolling, composite, raw grain)
- `registry/` — metric registry, dependency resolution, composite metric discovery
- `joins/` — join graph, join optimizer, canonical join paths
- `dimensions/` — dimension manager, time dimensions, scoping, caching
- `output/` — QueryResult, formatters, post-processing

Within `services/` (top-level): one file per entity. No sub-packages. Business logic lives here; repositories live in `core/repositories/`.

### `__init__.py` Exports

240 of 420 source files define `__all__`. The pattern is consistent: `__init__.py` files at package boundaries aggregate and re-export public APIs. Internal modules use `__all__` to mark their public surface. The `api/data_service_models/__init__.py` re-exports all `_entity.py` models publicly.

### Generated Code

`src/autom8_data/proto/` contains generated protobuf code. The `pyproject.toml` [tool.mypy.overrides] suppresses all type errors for `autom8_data.proto.*`. The scheduler subdirectory (`analytics/core/infra/scheduler/`) uses the `_` prefix for all internal files, re-exporting only the public API through `__init__.py`.

### DEPENDENCY RULE Comments

Layer boundaries are enforced by inline comments at the module level:

```
DEPENDENCY RULE: This package must NOT import from any other core/ package.   # infra/
DEPENDENCY RULE: This package may ONLY import from infra/.                     # models/
DEPENDENCY RULE: This package may import from models/ and infra/ only.         # registry/
DEPENDENCY RULE: This module may import from registry/, models/.               # execution/
```

These appear in module docstrings and `__init__.py` files across `analytics/core/`. They document — but do not enforce — layer ordering.

---

## Domain-Specific Idioms

### Registry Pattern

The codebase uses typed registry classes rather than global dicts. Each domain has a named `*Registry` class that owns storage and retrieval:
- `MetricRegistry` — stores Dimensions, Metrics, JoinDefinitions, CompositeMetrics
- `InsightRegistry` — stores InsightDefinitions
- `MaterializationRegistry` — discovers and stores table materialization configs
- `PostProcessorRegistry` — stores post-processor callables
- `TimeDimensionRegistry` — stores time dimension definitions

Registries expose `register_*()` methods (e.g., `register_dimension`, `register_metric`, `register_insight`) and `get_*()` retrieval methods. They raise specific `*NotFoundError` exceptions with available options when lookups fail.

### Metric Registration Pattern

Metrics are registered into a central `MetricRegistry` via three tiers:
1. **Auto-discovery** (80%): Standard SQLModel columns auto-registered from ORM models
2. **YAML** (15%): Simple aggregations with metadata in YAML config files
3. **Decorators** (5%): Complex computed logic via `@metric` and `@dimension` decorators (`analytics/core/metrics/decorators.py`)

`register_*()` functions in `analytics/core/metrics/library.py` are top-level module functions (not methods), called during initialization with a `MetricRegistry` argument.

### `__table_type__` Discovery Pattern

SQLModel classes annotate themselves with `__table_type__ = "fact"` or `__table_type__ = "dimension"` (string literals, not enum values). The `MaterializationRegistry.discover_tables()` method reads this attribute from SQLModel metadata to auto-classify tables.

### `db_error_boundary` + `transactional` Context Manager Pattern

Service operations never commit. Callers wrap service calls with:
```python
async with transactional(session):
    result = await service.create(...)
```
`transactional` composes `db_error_boundary()` and `session.begin()`. Services never call `session.commit()` directly.

### `BatchWriteResult` Pattern

Batch mutations return a `BatchWriteResult[T]` dataclass containing `results: list[T | BatchItemError]`, `success_count`, `failure_count`. Each failed item is represented as a `BatchItemError` with `index`, `code`, `message`, `field`. This pattern allows partial batch success.

### Fuzzy-Matched Suggestions

Semantic layer errors (`DimensionNotFoundError`, `MetricNotFoundError`) use `rapidfuzz` to generate typo-correction suggestions. This happens in `validation.get_suggestions()`, called lazily from `__str__`. Errors that include suggestions always format them as: `"Did you mean: {opt1}, {opt2}?"`.

### Cartesian Prevention Pattern

The join graph enforces that fact-table-to-fact-table joins are forbidden in single queries. `CartesianRiskError` (SCAR-001) is raised when detected. The `FactResolver` handles split-query execution to prevent row multiplication.

### Polars as the Internal DataFrame

`polars` is the canonical DataFrame type for all internal analytics operations. `QueryResult.df` is always a `polars.DataFrame`. The `.to_pandas()` conversion is available only at the output boundary. Never pass `pandas` DataFrames internally.

### `QueryResult` Envelope Type

`src/autom8_data/analytics/core/output/result.py::QueryResult` is the standard return type from `AnalyticsEngine.get()`. It wraps a Polars DataFrame with execution metadata (duration_ms, cache_hit, sql, schema). All analytics routes and services return or unpack `QueryResult` — never raw DataFrames.

### CompositeMetric Formula Pattern

`src/autom8_data/analytics/core/registry/composite.py` defines `CompositeMetric(MetricBase)` with a `Formula` abstract base class. Formulas are serializable (no lambdas) to support YAML config. Registered in `FORMULA_REGISTRY` dict for YAML deserialization via `formula_from_dict()`.

### Services Never Commit

`src/autom8_data/services/base.py` docstring: "Services NEVER commit transactions (caller owns transaction boundary). Services use flush() via repositories for constraint validation."

### Google AIP-134 FieldMask

`BaseService` in `services/base.py` implements FieldMask validation for partial updates, following Google AIP-134. This is a project-specific adoption of the FieldMask pattern applied uniformly to all CRUD operations.

### Optional Import Pattern (Scheduler)

Optional dependencies are guarded with try/except at the `__init__.py` level:
```python
try:
    from .scheduler import CircuitBreaker, ...
    _SCHEDULER_AVAILABLE = True
except ImportError:
    _SCHEDULER_AVAILABLE = False
```

### `_LazyLogger` Pattern

The `get_module_logger(__name__)` function returns a `_LazyLogger` proxy that defers creation of the underlying `autom8y_log` logger until first use.

### SDK-Only Imports

Direct imports of `loguru`, `httpx`, and `structlog` are **banned** via ruff `TID251`. All code must use the `autom8y_*` SDK wrappers.

### StrEnum Preference

Enums throughout the codebase prefer `StrEnum` over `Enum` for string-valued members. `IntEnum` is used only where integer comparison is needed. `Enum` is used for non-string, non-integer types.

### ADR / TDD / FR Reference Comments

Inline traceability comments link code to decision records: `ADR-014`, `TDD-CRUD-API-001`, `FR-OPS-001`, `WP-TEL-DATA`. Documentation-only markers, not executable code.

---

## Naming Patterns

### Type Naming

- **Exception classes**: Always end in `Error` (never `Exception`), e.g., `NotFoundError`, `CartesianRiskError`, `LookbackExceededError`
- **Enum classes**: Name describes the enumerated concept, not `*Enum` suffix. `OperationalMode`, `MetricAggregation`, `DimensionType`
- **Pydantic models**: Descriptive noun phrase, no `*Model` suffix. `InsightDefinition`, `ComputedFieldSpec`, `QueryResult`
- **Protocol classes**: Suffixed `Protocol` (e.g., `DimensionManagerProtocol`, `EntitySourceProtocol`)
- **Registry classes**: Suffixed `Registry` (e.g., `MetricRegistry`, `InsightRegistry`)
- **Manager classes**: Suffixed `Manager` (e.g., `DimensionManager`, `BatchCacheManager`)

### Function Naming

- **Factory functions**: `get_*()` (dependency injection providers), `create_*()` (explicit construction), `build_*()` (incremental construction)
- **Registration functions**: `register_*()` (add to registry), `discover_*()` (scan and register)
- **Private module-level functions**: `_snake_case()` prefix underscore

### File Naming

- **Route files**: `{plural_entity}_crud.py` (CRUD) or descriptive name
- **Internal/private files**: `_module_name.py` (underscore prefix; scheduler submodules, data_service_models)
- **Test files**: `test_*.py` (pytest convention)
- **Model files**: `_entity.py` for data_service_models subpackage, `models.py` for general use

### Variable and Constant Naming

- Module-level private constants: `_SCREAMING_SNAKE_CASE` (e.g., `_INVENTORY_CACHE_TTL_SECONDS`, `_NON_DIGIT_PATTERN`, `_SCHEDULER_AVAILABLE`)
- Public constants: `SCREAMING_SNAKE_CASE` (e.g., `MYSQL_DUPLICATE_KEY_ERRNO`, `PARQUET_ONLY_COLUMNS`)
- Module-level loggers: always named `logger` (lowercase): `logger = get_module_logger(__name__)`

### Module/Package Naming

- Packages: lowercase, snake_case, singular where conceptually singular (`core/`, `infra/`, `registry/`), plural where they contain collections (`models/`, `routes/`, `services/`, `metrics/`)
- The `primitives/` sub-packages use domain nouns: `anomaly/`, `correlation/`, `creative/`, `health/`, `optimization/`, `pacing/`, `peer/`, `shared/`

### API Route File Naming

API routes follow `{entity_plural}_crud.py` (e.g., `leads_crud.py`, `businesses_crud.py`, `payments_crud.py`). Special-purpose routes do not use the `_crud` suffix (e.g., `health.py`, `admin.py`).

### Star Imports and `__init__.py` Re-exports

Star imports (`from .module import *`) are an established convention (F403/F405 suppressed globally). `__init__.py` files use `__all__` to control what is re-exported. New packages MUST define `__all__` in both the module and the re-exporting `__init__.py`.

### Type Annotations

The project uses Python 3.10+ union syntax (`X | Y` not `Union[X, Y]`, `X | None` not `Optional[X]`) — enforced by `ruff UP` rules and `python_version = "3.12"` mypy target. `TYPE_CHECKING` guard blocks are used in 88 files to avoid circular import at runtime while preserving type information.

### Mypy Strict Mode

`[tool.mypy] strict = true`. Overrides suppress errors only for specific modules with justified reasons (SQLAlchemy descriptor patterns, generated proto code, circular imports).

### Deviations from PEP 8

- Nested or deferred imports inside function bodies or `__str__` methods are intentional (circular import avoidance pattern), not violations.
- `# type: ignore[...]` comments appear where SQLAlchemy or protobuf typing gaps exist — this is expected and documented.

---

## Knowledge Gaps

- **gRPC adapter error handling**: The `grpc/adapters/` layer uses `GRPCError` from `grpclib` but the mapping from `ServiceError` -> `GRPCError` status codes was not fully traced.
- **YAML metric configuration format**: The exact YAML schema for tier-2 metric registration (`primitives/config/`) was observed at surface level but not read in detail.
- **`core/filters/` DSL internals**: `FilterGroup`, `OrderSpec`, `build_where_clauses` were observed as imports but the full filter DSL implementation was not traced.
- **Analytics engine initialization sequence**: `analytics/initialization.py` was observed but not fully read — the startup order for registry population and connection pool initialization is not fully documented here.
