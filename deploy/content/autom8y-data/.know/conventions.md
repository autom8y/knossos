---
domain: conventions
generated_at: "2026-03-18T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "51f5e8d"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

### Error Creation Patterns

The project maintains two distinct error hierarchies depending on the layer:

**Analytics semantic layer (analytics/core/infra/exceptions.py):**

`SemanticError(Exception)` is the base class. All semantic layer exceptions inherit from it. Subclasses are rich dataclasses: they carry typed attributes and produce actionable messages including fuzzy-matched suggestions via `rapidfuzz`. The pattern for creating these exceptions is:

```python
class DimensionNotFoundError(SemanticError):
    def __init__(self, name: str, available: list[str]):
        self.name = name
        self.available = available
        super().__init__(f"Dimension '{name}' not found")

    def __str__(self) -> str:
        from .validation import get_suggestions
        base = f"Dimension '{self.name}' not found"
        if self.available:
            suggestions = get_suggestions(self.name, self.available)
            if suggestions:
                return f"{base}. Did you mean: {', '.join(suggestions)}?"
        return base
```

The `__str__` override pattern (not `__init__` message) is used when suggestions require lazy computation. Exceptions store structured attributes (`from_table`, `to_table`, `path_length`, `max_depth`, etc.) as instance variables for programmatic consumption.

Error types in the semantic layer include: `SemanticError`, `JoinPathError`, `JoinPathTooLongError`, `DimensionNotFoundError`, `MetricNotFoundError`, `JoinNotFoundError`, `SchemaValidationError`, `GrainIncompatibleError`, `DimensionInconsistencyError`, `ParquetRequiredError`, `QueryTimeoutError`, `CartesianRiskError`, `InvalidColumnError`, `MaterializationError`, `ConfigurationError`, `DataQualityError`, `SyncError`, `CycleDetectedError`, `ParquetConnectionError`, `DependencyResolutionError`.

**CRUD service layer (services/base.py):**

`ServiceError(Exception)` is the base class. Subclasses carry class-level `grpc_status`, `http_status`, and `code` attributes — enabling protocol-agnostic error handling. Constructor signature is `(message: str, field: str | None = None)`. Subclasses expose `to_error_detail()` for HTTP serialization.

```python
class NotFoundError(ServiceError):
    grpc_status = Status.NOT_FOUND
    http_status = 404
    code = "RESOURCE_NOT_FOUND"
```

Error types in the service layer: `ServiceError`, `NotFoundError`, `AlreadyExistsError`, `FKValidationError`, `ImmutableFieldError`, `ValidationError`, `DatabaseUnavailableError`, `InvalidFilterError`, `InvalidPeriodError`.

**Scheduler layer (analytics/core/infra/scheduler/_exceptions.py):**

`SchedulerError(Exception)` is its own base. Subclasses: `SchedulerNotAvailableError`, `SchedulerAlreadyRunningError`, `SchedulerNotRunningError`, `CircuitOpenError`.

**Auth layer (api/auth/exceptions.py):**

Auth exceptions (`AuthenticationError`, `AuthorizationError`, `KeyConfigurationError`, `AuthServiceUnavailableError`) do NOT inherit from `ServiceError`. They have their own standalone hierarchy.

### Error Wrapping

The `db_error_boundary()` async context manager in `services/base.py` wraps raw SQLAlchemy exceptions into `ServiceError` subclasses at the boundary. It is used by route handlers, not inside service methods. Key mappings:
- `OperationalError` → `DatabaseUnavailableError`
- `IntegrityError` (errno 1062) → `AlreadyExistsError`
- `IntegrityError` (errno 1452) → `FKValidationError`
- `DataError` → `ValidationError`

The sentinel helper `is_mysql_duplicate_key_error(exc)` detects MySQL errno 1062 by inspecting `exc.orig.args[0]`.

The `_safe_db_message(exc)` helper strips connection URLs from error messages before logging.

### Error Propagation

- **Service layer**: Services raise typed `ServiceError` subclasses directly. They do NOT commit or handle transactions. Route handlers own the transaction boundary via `async with transactional(session)`.
- **Semantic layer**: Errors propagate immediately to the caller. No error aggregation except within `SchemaValidationError`, which collects a list of `ValidationError` dataclass instances for batch reporting.
- **Boundary translation**: `src/autom8_data/api/errors.py` registers a comprehensive set of `FastAPI` exception handlers that convert every domain exception type to a structured JSON response (format: `{"error": {"code": ..., "message": ..., "request_id": ...}}`). Stack traces are never exposed to clients.

### Error at Boundaries

The `register_exception_handlers()` function in `src/autom8_data/api/errors.py` registers handlers from most-specific to least-specific. The catch-all `generic_error_handler` logs the exception with `logger.exception(...)` and returns a generic 500.

Logging of errors follows a severity convention: authentication failures → `logger.info`; resource exhaustion / timeouts → `logger.warning`; unhandled exceptions → `logger.exception`.

### Error Code System

Both layers use machine-readable string codes:
- Semantic layer: `VALIDATION_ERROR`, `METRIC_NOT_FOUND`, `DIMENSION_NOT_FOUND`, `INSIGHT_NOT_FOUND`, `REQUIRED_FILTER_MISSING`, `INVALID_SORT_COLUMN`, `SERVICE_BUSY`, `QUERY_TIMEOUT`, `DATA_STALE`
- Service layer: `RESOURCE_NOT_FOUND`, `RESOURCE_ALREADY_EXISTS`, `FK_VALIDATION_FAILED`, `IMMUTABLE_FIELD`, `VALIDATION_ERROR`, `DATABASE_UNAVAILABLE`, `INTERNAL_ERROR`

Codes are SCREAMING_SNAKE_CASE strings, not integers.

---

## File Organization

### Top-Level Package Structure

`src/autom8_data/` contains seven top-level packages: `analytics/`, `api/`, `clients/`, `core/`, `grpc/`, `proto/`, `services/`, `utils/`.

```
autom8_data/
  analytics/      # Analytics semantic layer + routes
  api/            # FastAPI HTTP API (routes, schemas, auth, models)
  clients/        # Internal gRPC/HTTP client wrappers
  core/           # Cross-cutting: config, logging, models, repositories
  grpc/           # gRPC handlers and adapters (outbound proto/grpc layer)
  proto/          # Generated protobuf stubs
  services/       # Business logic services (CRUD entity operations)
  utils/          # Utilities (e.g., phone normalization)
```

### Sub-Package Conventions

**`analytics/core/` contains functionally-named sub-packages:**
- `infra/` — foundation utilities: exceptions, logging, caching, connections, registry, materialization (NO internal deps)
- `models/` — dataclass definitions: Dimension, Metric, JoinDefinition, etc. (may import from `infra/` only)
- `registry/` — MetricRegistry, CompositeMetric, dependency resolution (imports from `models/`, `infra/`)
- `query/` — query builder, SQL generator, fact resolver, filter handling
- `execution/` — composite, rolling, and orchestrator for multi-metric execution
- `dimensions/` — dimension manager, time dimensions, computed fields, overrides
- `joins/` — join graph, optimizer, canonical paths
- `output/` — result types, post-processor, formatter
- `request/` — API request model translation
- `metrics/` — availability and library definitions
- `domain/` — operational mode enum
- `constants.py` — top-level constants for the analytics core

**`analytics/primitives/` — domain-specific analytic algorithms:**
Organized by domain: `anomaly/`, `config/`, `correlation/`, `creative/`, `efficiency/`, `health/`, `optimization/`, `pacing/`, `peer/`, `shared/`.

**`api/data_service_models/` — underscore-prefixed private model files:**
Each domain entity gets its own `_entity.py` file. This is a deliberate namespacing pattern: the underscore prefix signals internal/private module status even though all are collected by the package `__init__.py`. Examples: `_lead.py`, `_appointment.py`, `_payment.py`, `_base.py` (shared imports for all model files).

**`core/models/` — domain models:**
Entity models split by domain category with underscore-prefixed filenames: `_base.py` (shared imports), `_advertising.py`, `_communications.py`, `_platform.py`, `_scheduling.py`. All SQLModel subclasses.

### What Lives Where

- **Constants**: Dedicated `constants.py` files exist at the sub-package level where needed (e.g., `analytics/core/constants.py`, `analytics/core/infra/constants.py`). Module-level private constants use `_SCREAMING_SNAKE_CASE`.
- **Exceptions**: Centralized in `exceptions.py` files within each major subsystem (`analytics/core/infra/exceptions.py`, `analytics/core/infra/scheduler/_exceptions.py`, `api/auth/exceptions.py`).
- **Logging**: Module-level logger created at top of each module: `logger = get_module_logger(__name__)`.
- **Protocols/Interfaces**: Protocol classes live in `protocols.py` files within the relevant package (e.g., `analytics/core/models/protocols.py`).
- **`__init__.py` as re-export hubs**: Sub-packages use their `__init__.py` to re-export public API with explicit `__all__` lists. Consumers import from the package, not from individual modules.
- **Scheduler internals**: Files prefixed with `_` (e.g., `_base.py`, `_circuit_breaker.py`, `_retry.py`) inside `infra/scheduler/` — the underscore prefix denotes internal-only files not exposed directly by the package.

### DEPENDENCY RULE Comments

Layer boundaries are enforced by inline comments at the module level:

```
DEPENDENCY RULE: This package must NOT import from any other core/ package.   # infra/
DEPENDENCY RULE: This package may ONLY import from infra/.                     # models/
DEPENDENCY RULE: This package may import from models/ and infra/ only.         # registry/
DEPENDENCY RULE: This module may import from registry/, models/.               # execution/
```

These appear in module docstrings and `__init__.py` files across `analytics/core/`. They document — but do not enforce — layer ordering. Present in approximately 33 files.

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

### `__table_type__` Discovery Pattern

SQLModel classes annotate themselves with `__table_type__ = "fact"` or `__table_type__ = "dimension"` (string literals, not enum values). The `MaterializationRegistry.discover_tables()` method reads this attribute from SQLModel metadata to auto-classify tables. This is a project-specific pattern that enables declarative table registration without explicit configuration.

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

### Optional Import Pattern (Scheduler)

Optional dependencies are guarded with try/except at the `__init__.py` level:
```python
try:
    from .scheduler import CircuitBreaker, ...
    _SCHEDULER_AVAILABLE = True
except ImportError:
    _SCHEDULER_AVAILABLE = False
```
Names are then conditionally added to `__all__`. This pattern is used when a subsystem requires an optional dependency (e.g., `apscheduler`).

### `_LazyLogger` Pattern

The `get_module_logger(__name__)` function returns a `_LazyLogger` proxy that defers creation of the underlying `autom8y_log` logger until first use. This prevents auto-configure warnings when modules define module-level loggers at import time. All modules use: `logger = get_module_logger(__name__)`.

### ADR / TDD / FR Reference Comments

Inline traceability comments appear in module docstrings and inline, linking code to decision records:
- `ADR-014`, `ADR-0052`, `ADR-0061` — Architecture Decision Records
- `TDD-CRUD-API-001`, `TDD-AUTH-001` — Test-Driven Design documents
- `FR-OPS-001`, `WP-TEL-DATA` — Functional requirements and work packages

These are documentation-only markers, not executable code.

### StrEnum Preference

Enums throughout the codebase prefer `StrEnum` over `Enum` for string-valued members. This allows direct string comparison without `.value`. `IntEnum` is used only where integer comparison is needed (e.g., `JoinPenalty`, `ServingStatus`). `Enum` is used for non-string, non-integer types (e.g., `MetricType`, `TranslationContext`).

---

## Naming Patterns

### File Naming

- **Singular nouns for single-responsibility modules**: `aggregation.py`, `executor.py`, `builder.py`, `planner.py`, `optimizer.py`
- **Plural nouns for collections/libraries**: `metrics/library.py`, `insights/library.py`, `dimensions/`
- **Verb-noun for action-focused modules**: `fact_resolver.py`, `filter_merge.py`, `dimension_resolver.py`, `sql_generator.py`, `window_metric_sql.py`
- **Noun-noun for compound concepts**: `connection_pool.py`, `connection_router.py`, `time_dimensions.py`, `cache_metrics.py`
- **Underscore-prefix for internal/private modules**: `_base.py`, `_circuit_breaker.py`, `_exceptions.py`, `_retry.py` within `scheduler/`; all files in `api/data_service_models/` and `core/models/`

### Class Naming

- **`*Error` suffix** for all exceptions (never `*Exception` except where inheriting from Python built-ins)
- **`*Registry` suffix** for registry classes
- **`*Service` suffix** for business logic service classes
- **`*Adapter` suffix** for gRPC adapter classes
- **`*Handler` suffix** for gRPC handler classes
- **`*Config` suffix** for configuration dataclasses
- **`*Spec` suffix** for declarative specification dataclasses
- **`Base*` prefix** for abstract base classes (`BaseService`, `BaseAdapter`, `BaseScheduler`)
- **`*Builder` suffix** for query builder classes
- **`*Resolver` suffix** for resolver classes

### Variable and Constant Naming

- Module-level private constants: `_SCREAMING_SNAKE_CASE` (e.g., `_INVENTORY_CACHE_TTL_SECONDS`, `_NON_DIGIT_PATTERN`, `_SCHEDULER_AVAILABLE`)
- Public constants: `SCREAMING_SNAKE_CASE` (e.g., `MYSQL_DUPLICATE_KEY_ERRNO`, `PARQUET_ONLY_COLUMNS`)
- Module-level loggers: always named `logger` (lowercase): `logger = get_module_logger(__name__)`

### Module/Package Naming

- Packages: lowercase, snake_case, singular where conceptually singular (`core/`, `infra/`, `registry/`), plural where they contain collections of things (`models/`, `routes/`, `services/`, `metrics/`)
- The `primitives/` sub-packages use domain nouns: `anomaly/`, `correlation/`, `creative/`, `health/`, `optimization/`, `pacing/`, `peer/`, `shared/`

### API Route File Naming

API routes follow `{entity_plural}_crud.py` (e.g., `leads_crud.py`, `businesses_crud.py`, `payments_crud.py`). Special-purpose routes do not use the `_crud` suffix (e.g., `health.py`, `admin.py`, `messages.py`, `gid_mappings.py`).

### Deviations from PEP 8

- Nested or deferred imports inside function bodies or `__str__` methods are intentional (circular import avoidance pattern), not violations.
- `# type: ignore[...]` comments appear where SQLAlchemy or protobuf typing gaps exist — this is expected and documented.

---

## Knowledge Gaps

- **gRPC adapter error handling**: The `grpc/adapters/` layer uses `GRPCError` from `grpclib` but the mapping from `ServiceError` → `GRPCError` status codes was not fully traced. The `BaseAdapter` class was not read in detail.
- **`api/routes/factory.py` pattern**: A router factory pattern was observed but not read in depth. The exact mechanism for generating CRUD routes programmatically is undocumented here.
- **`core/schema_meta.py` `ApiField`/`FieldRole` pattern**: These are imported in `core/models/_base.py` and used across all SQLModel entities but were not read. Their role in FieldMask validation or schema generation is a gap.
- **`analytics/fixtures/` structure**: Fixture builder and config patterns were not examined.
