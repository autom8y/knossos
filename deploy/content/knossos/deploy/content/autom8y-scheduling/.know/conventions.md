---
domain: conventions
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "39376b6"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## File Organization

The codebase follows a clear layered package structure under `src/autom8_scheduling/`. Files are organized by functional responsibility, not by class count — multiple closely related classes may coexist in a single file, while domain boundary files receive their own module.

**Package layout:**

```
src/autom8_scheduling/
├── __init__.py              # package version only
├── app.py                   # FastAPI factory (create_app)
├── config.py                # SchedulingSettings + get_settings()
├── health.py                # health/readiness/deps endpoints + models
├── models/
│   ├── __init__.py          # re-export barrel: imports from scheduling.py + shared.py
│   ├── _base.py             # shared SQLModel imports, ApiField stub, validators
│   ├── scheduling.py        # 5 ORM models: Appointment, Address, Hours, Employee, BusinessOffer
│   ├── shared.py            # 2 cross-domain ORM models: Business, Lead
│   └── schemas.py           # 6 Pydantic request bodies
├── services/
│   ├── __init__.py          # 1-line docstring only
│   ├── base.py              # ServiceError hierarchy + BaseService (5 classes)
│   └── appointment.py       # AppointmentService
├── scheduling/
│   ├── __init__.py          # docstring listing sub-modules
│   ├── engine.py            # AvailabilityEngine
│   ├── booking.py           # BookingEngine
│   ├── reminder.py          # ReminderEngine
│   ├── gcal_overlay.py      # GCalOverlay + cache internals
│   ├── gcal_sync.py         # fire-and-forget GCal write path
│   ├── write_ops.py         # AppointmentStatus, state machine, audit logging
│   ├── validation.py        # check_scheduling_readiness
│   ├── constants.py         # SLOT_BLOCKING_STATUSES, CANCELLABLE_STATUSES
│   └── notifications/
│       ├── __init__.py
│       ├── dispatch.py      # orchestrator functions
│       ├── schemas.py       # NotificationType, NotificationChannel, SmsResult, EmailResult
│       ├── config.py        # NotificationConfig + lru_cache singleton
│       ├── templates.py     # format_* / render_* pure functions
│       ├── sms_service.py   # NotificationSmsService
│       └── email_service.py # NotificationEmailService
└── api/
    ├── __init__.py
    └── routes/
        ├── __init__.py      # all route handlers + scheduling_router (monolith file)
        └── deps.py          # FastAPI dependency providers
```

**File naming patterns:**
- Snake_case throughout
- Private/internal helpers prefixed with underscore (e.g., `_base.py`, `_CacheEntry`, `_FreeBusyCache`)
- Domain engine classes get their own file (`engine.py`, `booking.py`, `reminder.py`)
- Constants and schemas get dedicated files (`constants.py`, `schemas.py`)

**`__init__.py` patterns:**
- Top-level `__init__.py`: version export only, minimal `__all__`
- `models/__init__.py`: explicit re-export barrel with `__all__`, collects from sub-modules
- `scheduling/__init__.py`: docstring listing sub-modules, no re-exports
- `api/__init__.py`, `services/__init__.py`: 1-line docstring, no re-exports
- `notifications/__init__.py`: docstring only
- `api/routes/__init__.py`: all route handlers live here (router-as-module pattern — entire routes layer is one file, not split by resource)

**Notable pattern:** `models/scheduling.py` and `models/shared.py` use a wildcard import (`from autom8_scheduling.models._base import *  # noqa: F403`) to pull SQLModel primitives into scope. This is explicitly tolerated by `noqa: F403` and documented in `_base.py`'s `__all__`.

## Naming Conventions

**Modules:** `snake_case`, descriptive of functional role — `write_ops`, `gcal_overlay`, `gcal_sync`, `booking`, `reminder`, `constants`. Sub-packages follow domain vocabulary (`notifications`, `scheduling`, `services`).

**Classes:**
- ORM models: `PascalCase` nouns matching the domain entity — `Appointment`, `Business`, `Address`, `Employee`, `Hours`, `BusinessOffer`, `Lead`
- Engine/service classes: `PascalCase` + `Engine` or `Service` suffix — `AvailabilityEngine`, `BookingEngine`, `ReminderEngine`, `BaseService`, `AppointmentService`, `NotificationSmsService`, `NotificationEmailService`
- Pydantic request bodies: `PascalCase` + `Body` suffix — `BookingRequestBody`, `CancelRequestBody`, `ConfigUpdateBody`
- Pydantic settings: `PascalCase` + `Settings` suffix — `SchedulingSettings`, `NotificationConfig` (outlier — no suffix)
- Internal/private classes: underscore-prefixed — `_CacheEntry`, `_FreeBusyCache`
- StrEnum classes: `PascalCase` nouns — `AppointmentStatus`, `NotificationType`, `NotificationChannel`, `HealthStatus`
- Data containers: `PascalCase` + `Result` suffix — `SmsResult`, `EmailResult`, `CheckResult`, `HealthResponse`, `EmailContent`
- Error types: `PascalCase` + `Error` — `ServiceError`, `NotFoundError`, `FKValidationError`, `ValidationError`, `SchedulingValidationError`

**Functions:**
- `snake_case`, verbs or verb phrases — `compute_availability`, `book`, `cancel`, `reschedule`, `check_scheduling_readiness`, `dispatch_booking_notifications`
- Private helpers: underscore-prefixed — `_check_global_scheduling`, `_check_business_scheduling`, `_get_business_for_notifications`, `_initialize_mysql_pool`, `_hash_pii`
- Singleton factories: `get_{resource}` pattern — `get_settings()`, `get_notification_config()`, `get_session()`, `get_gcal_client()`
- Template/formatting functions: `format_{context}_{type}` or `render_{context}_{type}` — `format_booking_confirmation_sms`, `render_booking_email`
- Dispatch functions: `dispatch_{resource}_{action}` — `dispatch_booking_notifications`, `dispatch_gcal_booking_sync`

**Variables and constants:**
- Variables: `snake_case`
- Module-level constants: `UPPER_SNAKE_CASE` — `SLOT_BLOCKING_STATUSES`, `CANCELLABLE_STATUSES`, `SCHEDULING_ENABLED_ENV`, `TWILIO_API_BASE`, `CACHE_TTL_EVENTUAL`
- Private module-level singletons: underscore-prefixed lowercase — `_tracer`, `_notification_tracer`

**Domain-specific conventions:**
- `office_phone` (not `business_phone`) as the canonical business identifier throughout
- `lead_phone` (not `patient_phone`) as the canonical patient identifier in API layer; templates use `patient_phone` internally
- `ApiField` is a function (not a class) with `N802` noqa to allow PascalCase by convention inherited from autom8y-data
- Exception: `NotificationConfig` uses `UPPER_CASE` field names for env-var-sourced Pydantic fields (direct env var name mapping), while `SchedulingSettings` uses lowercase with prefix

## Error Handling Style

**Service layer:** Custom exception hierarchy rooted at `ServiceError(Exception)` in `src/autom8_scheduling/services/base.py`. Each subclass carries `http_status` and `code` class attributes and a `to_error_detail()` method:

```
ServiceError
├── NotFoundError       (http_status=404, code="RESOURCE_NOT_FOUND")
├── FKValidationError   (http_status=400, code="FK_VALIDATION_FAILED")
├── ImmutableFieldError (http_status=400, code="IMMUTABLE_FIELD")
└── ValidationError     (http_status=400, code="VALIDATION_ERROR")
```

These are raised in `AppointmentService` and caught by route handlers to produce HTTP responses.

**Scheduling domain layer:** A separate `SchedulingValidationError(Exception)` in `write_ops.py` for state machine violations. This is a domain error, not a service error. Route handlers are expected to catch it and return HTTP 400/422.

**API layer:** Domain exceptions are translated at the route boundary into `HTTPException` with appropriate status codes. The pattern is:
```python
try:
    result = await engine.book(...)
except DomainException as exc:
    raise HTTPException(status_code=4xx, detail=...) from exc
```

**Exception chaining:** `raise X from exc` is used consistently for exception wrapping in route handlers.

**Broad except + logger.exception:** Post-commit fire-and-forget operations (notification dispatch, GCal sync) use `except Exception: logger.exception(...)` with a structured log message and `return None` or fallback. This isolates failures in optional side effects from the main write path.

**Error logging pattern:** Structured log events as machine-parseable snake_case strings, followed by `extra={...}` dict:
```python
logger.error("scheduling_gate_db_error", extra={"office_phone": ..., "error": str(exc)})
logger.exception("post_commit_dispatch_error", extra={"appointment_id": ..., "flow": "booking"})
```

## Type Annotation Patterns

**Annotation coverage:** Type annotations are present on all function signatures across the codebase. The `pyproject.toml` sets `mypy strict = true`, though domain modules have `ignore_errors = true`.

**Return types:** All public functions annotate return types. Route handlers use `dict[str, Any]`, `dict[str, Any] | JSONResponse`, or `list[dict[str, Any]]`. Engine methods return `dict[str, Any]`.

**Union syntax:** Python 3.10+ `X | Y` union syntax used throughout (e.g., `int | None`, `dict[str, Any] | None`). `Optional[X]` is not used.

**`Annotated[...]` for FastAPI query params:** Used consistently for route handler query parameters:
```python
office_phone: Annotated[str, Query(description="Business phone (E.164)")]
employee_id: Annotated[int | None, Query(description="Employee filter")] = None
```

**`typing` module usage:**
- `Any` — heavily used in engine return types and service interfaces
- `Generic[T]` + `TypeVar("T")` — used in `BaseService` for the generic entity type
- `TYPE_CHECKING` block — used in `gcal_overlay.py` and `gcal_sync.py` to defer imports
- `collections.abc.AsyncGenerator` — used in `conftest.py` and `deps.py` inside `TYPE_CHECKING` blocks
- No `Protocol` usage found
- No `TypeAlias` found

**`from __future__ import annotations`:** Used selectively in `models/_base.py`, `health.py`, `gcal_overlay.py`, `gcal_sync.py`, and test files. Not universally applied.

**Pydantic `ConfigDict`:** All Pydantic models use `model_config = ConfigDict(frozen=True, extra="forbid")` for immutable, strict request body parsing (Pydantic v2 pattern).

**mypy relaxation:** All domain modules (`api.routes`, `services.*`, `scheduling.*`, `models.*`) have `ignore_errors = true` in `pyproject.toml` due to raw SQL patterns migrated from autom8y-data. Strict checking is only active for `app.py`, `config.py`, `health.py`, and `__init__.py`.

## Import Organization

**Import ordering:** Ruff `isort` is configured with `known-first-party` covering both `autom8_scheduling` (local) and all `autom8y_*` SDK packages. Three logical import groups:

1. **Standard library** — `import asyncio`, `import logging`, `from datetime import ...`, `from typing import ...`
2. **Third-party** — FastAPI, SQLAlchemy, Pydantic, Prometheus, OpenTelemetry, `httpx`
3. **First-party** — both `autom8_scheduling.*` (local) and `autom8y_*` SDKs (treated as first-party)

**Absolute vs relative imports:** Mixed strategy:
- Top-level package and cross-package imports use absolute imports (`from autom8_scheduling.models import ...`)
- Intra-package imports within `notifications/` use relative imports (`from .config import get_notification_config`, `from .schemas import SmsResult`)
- No parent-relative (`from ..`) imports found

**`__all__` usage:** Selective — only 6 of 31 source files define `__all__`:
- `autom8_scheduling/__init__.py` — version only
- `models/__init__.py` — re-export barrel (explicit)
- `models/_base.py` — wildcard-export source
- `models/schemas.py` — explicit public API
- `api/routes/__init__.py` — exports `scheduling_router`
- `api/routes/deps.py` — exports dependency functions

**Lazy / deferred imports:** Conditional imports in two patterns:
1. Optional dependency guard — `try: import autom8y_gcal; GCAL_AVAILABLE = True except ImportError: GCAL_AVAILABLE = False`
2. Post-commit deferred imports inside route handlers (avoids circular imports)

## Documentation Style

**Docstring format:** No formal Google/NumPy/Sphinx docstring format enforced. Docstrings are terse — single-sentence imperative descriptions. No `Args:`, `Returns:`, or `Raises:` sections.

**Coverage:**
- All 31 source files have module-level docstrings
- All public classes have 1-line class docstrings
- Public methods on engine/service classes have 1-line docstrings
- Route handler functions all have single-sentence docstrings

**Module docstrings** follow a consistent pattern: first line is a one-sentence summary, followed by blank line and detail when needed:
```python
"""Scheduling Service - FastAPI Application Factory.

create_app() factory with 3-phase lifespan:
  Phase 1: MySQL connection pool
  Phase 2: GCal client (conditional)
  Phase 3: APScheduler placeholder (future)
"""
```

**Inline comments:** Used liberally for section demarcation:
- Wide separator bars in dense files: `# ---------------------------------------------------------------------------`
- Section headers with `# =====` in `services/base.py`
- Inline context comments on fields: `# Dead UUID column - 100% NULL in production. NOT a real FK.`
- ADR/ticket references in comments: `# ADR-002: SELECT FOR UPDATE within transaction`, `# OBS-SCHED-001 REC-01 through REC-03`

## Domain Idioms

**1. Engine pattern:** Domain logic is encapsulated in `*Engine` classes that take `AsyncSession` at construction and expose async methods. They are instantiated per-request by route handlers. No state beyond the session:
```python
class AvailabilityEngine:
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def compute_availability(self, office_phone: str, ...) -> dict[str, Any]: ...
```

**2. Keyword-only arguments:** All public dispatch functions and engine methods use `*` to enforce keyword-only parameters:
```python
async def dispatch_booking_notifications(
    *,
    booking_result: dict[str, Any],
    session: AsyncSession,
    business_name: str,
    ...
) -> None:
```

**3. `lru_cache` singleton factory:** Configuration objects are lazily created and cached via `@lru_cache` on a `get_*` factory function:
```python
@lru_cache
def get_settings() -> SchedulingSettings:
    return SchedulingSettings()
```

**4. Fire-and-forget with `asyncio.gather(return_exceptions=True)`:** Post-commit side effects (notifications, GCal sync) are wrapped in `asyncio.gather(*coros, return_exceptions=True)` inside a `try/except Exception: logger.exception(...)` block. This pattern appears in every mutation route.

**5. Prometheus metrics at module level:** Metric objects declared as module-level constants following `scheduling_{operation}_{metric_type}` naming:
```python
scheduling_booking_outcomes = Counter("scheduling_booking_outcomes_total", ...)
scheduling_availability_duration = Histogram("scheduling_availability_duration_seconds", ...)
```

**6. OpenTelemetry span wrapping:** Domain operations instrumented with `_tracer.start_as_current_span(...)` using `with` context manager, injecting semantic attribute constants from `autom8y_telemetry.conventions`.

**7. `office_phone` as the universal tenant key:** `office_phone` (E.164 format) is the business identity key threaded through all engine calls, route parameters, notification dispatch, and database queries.

**8. Soft-delete pattern:** Cancellation implemented as status mutation (`status="cancelled"`) plus metadata written to the `config` JSON column.

**9. Graceful degradation for optional integrations:** GCal is conditionally available. Both `gcal_overlay.py` and `gcal_sync.py` check `GCAL_AVAILABLE` before executing GCal-dependent logic.

**10. `StrEnum` for status/type enumerations:** `StrEnum` (Python 3.11+) used instead of `Enum` for domain value sets, making enum members directly comparable to strings.

**11. ADR/ticket annotation in module docstrings and comments:** Implementation decisions traced back to ADR/ticket identifiers throughout.

## Knowledge Gaps

- **`NotificationConfig` field casing**: Unlike `SchedulingSettings` (lowercase with prefix), `NotificationConfig` uses `SCREAMING_SNAKE_CASE` with no prefix. The reason (migrated verbatim from autom8y-data) is not documented.
- **`api/routes/__init__.py` size**: The entire scheduling API router (~1000 lines, 12 endpoints) lives in a single `__init__.py`. No comment explains why it was not split.
- **`_parse_appointment_datetime` cross-module reference**: `booking.py` imports a private function from `engine.py`. The leading underscore signals it is not part of the public API, but it is used externally.
- **SQLAlchemy session transaction boundaries**: `flush()` vs `commit()` usage is not uniformly documented. Engine classes use `flush()`; `gcal_sync.py` helpers call `session.commit()` directly.
