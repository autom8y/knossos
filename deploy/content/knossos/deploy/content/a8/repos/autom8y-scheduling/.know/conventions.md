---
domain: conventions
generated_at: "2026-03-27T19:56:20Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4557333"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

### Two Parallel Exception Hierarchies

This codebase uses two distinct exception hierarchies that coexist and serve different concerns.

**Hierarchy 1: Domain engine exceptions** (`src/autom8_scheduling/scheduling/exceptions.py`)

Base class `SchedulingEngineError(Exception)` with structured subclasses that carry context as attributes:

```
SchedulingEngineError
├── BookingConflict(conflict_reason, office_phone, start_datetime, end_datetime, conflicting_ids)
│   └── RescheduleConflict(old_appointment_id, ...)  # extends BookingConflict
├── AppointmentNotFound(appointment_id, office_phone, lead_phone, reason)
├── AppointmentNotCancellable(appointment_id, current_status, office_phone, lead_phone)
└── AppointmentAlreadyCancelled(appointment_id, office_phone, lead_phone)

SchedulingGateError(Exception)  # separate -- gate-level rejections
    (status_code, code, message, details)
```

Each exception stores structured data as instance attributes (not just a string message). The `super().__init__()` call receives a human-readable summary but the rich data travels via attributes.

**Hierarchy 2: Service layer exceptions** (`src/autom8_scheduling/services/base.py`)

Base class `ServiceError(Exception)` with class-level `http_status` and `code` constants:

```
ServiceError(message, field=None)
├── NotFoundError          [http_status=404, code="RESOURCE_NOT_FOUND"]
├── FKValidationError      [http_status=400, code="FK_VALIDATION_FAILED"]
├── ImmutableFieldError    [http_status=400, code="IMMUTABLE_FIELD"]
└── ValidationError        [http_status=400, code="VALIDATION_ERROR"]
```

**Hierarchy 3: Write-path validation** (`src/autom8_scheduling/scheduling/write_ops.py`)

`SchedulingValidationError(Exception)` carries `.message` attribute, maps to HTTP 422.

### Error Creation Pattern

Exceptions are raised directly with structured context — no wrapping in a helper function:

```python
raise BookingConflict(
    conflict_reason="slot_taken",
    office_phone=office_phone,
    start_datetime=start_datetime,
    end_datetime=end_datetime,
)
```

No `raise X from Y` chaining is used within domain logic. Exception chaining (`from exc`) appears only at infrastructure boundaries (DB, GCal) where the original exception should be preserved for debugging:

```python
# handlers.py line 152
raise HTTPException(status_code=500, detail="Internal error...") from exc
```

### Error Propagation Style

Domain functions raise exceptions and do not return error codes. The `SchedulingEngineError` subtypes propagate up from engine methods to route handlers. Route handlers catch specific types in priority order:

1. Catch `RescheduleConflict` before `BookingConflict` (subclass-first pattern)
2. Catch `BookingConflict` with sub-dispatch on `.conflict_reason`
3. Catch `AppointmentNotFound`, `AppointmentNotCancellable`, `AppointmentAlreadyCancelled`
4. Fall through to 500 for unknown `SchedulingEngineError` subtypes

### Error Handling at API Boundaries

All errors at the API boundary are converted to a consistent `SchedulingErrorResponse` envelope via named helper functions in `handlers.py`:

- `handle_engine_exception(exc, request) -> JSONResponse` — maps `SchedulingEngineError` subtypes
- `handle_validation_error(exc, request) -> JSONResponse` — maps `SchedulingValidationError`
- `handle_scheduling_gate_error(request, exc) -> JSONResponse` — registered as FastAPI global handler for `SchedulingGateError`
- `error_response(request, status_code, code, message, ...) -> JSONResponse` — the raw builder, used by all handlers

No `HTTPException` with plain string `detail` is used for domain errors. The two places where bare `HTTPException` appears (`check_global_scheduling`, DB failure paths) are known inconsistencies documented in the codebase.

### Error Code System

All errors carry a SCREAMING_SNAKE_CASE `code` string. Examples:
- `SLOT_CONFLICT`
- `APPOINTMENT_NOT_FOUND`
- `APPOINTMENT_NOT_CANCELLABLE`
- `SCHEDULING_DISABLED_FOR_BUSINESS`
- `INVALID_STATUS_TRANSITION`
- `INTERNAL_SERVER_ERROR`

These codes are stable across versions per ADR-scheduling-engine-exceptions and are intended for programmatic handling by AI agents.

### Logging at Error Sites

Most modules use `logging.getLogger(__name__)` directly. The app factory uses `autom8y_log.get_logger(__name__)`. This is a known inconsistency — `pyproject.toml` bans `loguru` but the SDK inconsistency is tolerated. Structured log events use snake_case event names and pass context via `extra={}` dict:

```python
logger.error(
    "scheduling_gate_db_error",
    extra={"office_phone": office_phone, "error": str(exc)},
)
```

Infrastructure-layer code (app.py via `autom8y_log`) passes kwargs directly:
```python
logger.error("gcal_client_failed", error=str(e), error_type=type(e).__name__)
```

---

## File Organization

### Package Structure

```
src/autom8_scheduling/
├── __init__.py           (empty)
├── app.py                (application factory: create_app(), lifespan, middleware)
├── config.py             (SchedulingSettings, get_settings() singleton)
├── health.py             (health_router, health check functions)
├── api/
│   ├── __init__.py
│   └── routes/
│       ├── __init__.py
│       ├── deps.py       (FastAPI dependency providers: get_session, get_gcal_client, get_current_user)
│       ├── handlers.py   (shared handler utilities: exception mapping, response builders, gating checks)
│       ├── businesses.py (businesses_router — monolithic single-file router)
│       ├── internal.py   (internal_router — hidden from OpenAPI)
│       └── appointments/ (sub-package: router split across 3 files)
│           ├── __init__.py  (defines appointments_router, re-exports for patching)
│           ├── availability.py
│           ├── booking.py
│           └── queries.py
├── models/
│   ├── __init__.py       (re-export layer: Address, Appointment, Business, BusinessOffer, Employee, Hours, Lead)
│   ├── _base.py          (SQLModel imports, ApiField stub, FieldRole, validators)
│   ├── envelopes.py      (SchedulingResponse[T], SchedulingErrorResponse, ResponseMeta, ErrorDetail)
│   ├── responses.py      (typed Pydantic response data models per endpoint)
│   ├── schemas.py        (Pydantic request schemas, shared annotated types, AppointmentStatusValue)
│   ├── scheduling.py     (ORM models: Appointment, Address, Hours, Employee, BusinessOffer)
│   └── shared.py         (READ-ONLY ORM models: Business, Lead)
├── scheduling/
│   ├── __init__.py
│   ├── booking.py        (BookingEngine class — write path: book, cancel, reschedule)
│   ├── booking_helpers.py (idempotency key parsing, slot resolution helpers)
│   ├── constants.py      (TERMINAL_STATUSES, KNOWN_STATUSES, CANCELLABLE_STATUSES, is_slot_blocking())
│   ├── engine.py         (AvailabilityEngine class — read path: slot generation)
│   ├── exceptions.py     (SchedulingEngineError hierarchy, SchedulingGateError)
│   ├── gcal_overlay.py   (GCal busy-time overlay for availability computation)
│   ├── gcal_sync.py      (GCal event create/update/delete sync)
│   ├── ghl_sync.py       (GoHighLevel calendar sync)
│   ├── offer_resolution.py (resolve_business_offer(), SchedulingConfig dataclass)
│   ├── reminder.py       (reminder-eligible appointment queries)
│   ├── results.py        (frozen dataclasses: BookingResult, CancelResult, RescheduleResult, etc.)
│   ├── validation.py     (business hours validation, scheduling prerequisites)
│   ├── write_ops.py      (validate_status_transition, emit_scheduling_audit_event, dispatch_scheduling_notifications)
│   └── notifications/
│       ├── __init__.py
│       ├── config.py
│       ├── dispatch.py
│       ├── email_service.py
│       ├── schemas.py
│       ├── sms_service.py
│       └── templates.py
└── services/
    ├── __init__.py
    ├── base.py           (ServiceError hierarchy, BaseService with FieldMask)
    └── appointment.py    (AppointmentService — admin CRUD on appointments)
```

### File Naming Conventions

- Route files are noun-plural: `appointments/`, `businesses.py`, `internal.py`
- Sub-module split follows concern, not entity: `booking.py`, `availability.py`, `queries.py` (not appointment-write, appointment-read)
- `_base.py` prefix indicates a shared import/stub layer (private to the package)
- `handlers.py` is the shared utility file for route layers
- `deps.py` is always the FastAPI dependency provider file

### What Goes Where

| Concern | File |
|---------|------|
| Application factory, lifespan, middleware | `app.py` |
| Settings, env vars | `config.py` |
| FastAPI dependencies (`Depends(...)`) | `api/routes/deps.py` |
| Exception mapping, response builders | `api/routes/handlers.py` |
| Domain ORM models (scheduling tables) | `models/scheduling.py` |
| Read-only foreign ORM models | `models/shared.py` |
| API request schemas (body models) | `models/schemas.py` |
| API response data models | `models/responses.py` |
| Response envelope wrappers | `models/envelopes.py` |
| Domain constants + frozensets | `scheduling/constants.py` |
| Domain exceptions | `scheduling/exceptions.py` |
| Domain success result types | `scheduling/results.py` |
| Cross-cutting write operations | `scheduling/write_ops.py` |

### `__init__.py` Export Patterns

`models/__init__.py` is a flat re-export layer — all domain ORM models re-exported with explicit `__all__`. Route packages (`appointments/__init__.py`) re-export names for test patching compatibility, documented inline. `api/__init__.py` is empty. `scheduling/__init__.py` is empty.

### Constants and Configuration

Constants live in `constants.py` as module-level `frozenset[str]` with inline documentation. Configuration is in `config.py` as a `pydantic_settings.BaseSettings` subclass with `@lru_cache` singleton factory. No `CONSTANTS.py` or `settings.py` naming — the convention is descriptive file names.

---

## Domain-Specific Idioms

### Two-Tier Success/Error Return Pattern

The scheduling domain separates success outcomes (typed frozen dataclasses) from failure outcomes (typed exceptions). This is explicit per ADR-scheduling-engine-exceptions:

- **Successes**: `BookingResult`, `IdempotentBookingResult`, `CancelResult`, `IdempotentCancelResult`, `RescheduleResult`, `IdempotentRescheduleResult` — all in `results.py`
- **Failures**: `SchedulingEngineError` subtypes — in `exceptions.py`

Idempotent replays return a distinct type (`Idempotent*`) rather than a bool flag on a shared type. The caller uses `isinstance()` to distinguish fresh vs. replay, then maps to HTTP 201 or 200 accordingly.

### Deny-List Status Strategy

The slot-blocking strategy is a deny-list, not an allow-list. `is_slot_blocking(status)` returns `True` for any status NOT in `TERMINAL_STATUSES`. This is a deliberate defensive pattern documented extensively in `constants.py` to defend against unknown statuses from 27 upstream source systems. `SLOT_BLOCKING_STATUSES` frozenset is marked deprecated — the deny-list is the canonical approach.

Unknown statuses trigger a structured warning log but are treated as slot-blocking (fail safe):
```python
_logger.warning("unknown_appointment_status_blocking_by_default", extra={"status": status})
```

### SchedulingConfig Dataclass as Config Resolution Layer

All scheduling configuration reads go through `SchedulingConfig` (frozen dataclass in `offer_resolution.py`). It is never instantiated directly in business logic — always via two class methods:
- `SchedulingConfig.from_business_offer(offer)` — extracts from a resolved `BusinessOffer`
- `SchedulingConfig.not_configured()` — sentinel for "no offer found" (all-disabled defaults)

This pattern hides the inverted `disabled` -> `scheduling_enabled` semantic and normalizes all nullable fields to defaults in one place.

### `ApiField` Compatibility Stub

`models/_base.py` provides `ApiField()` as a thin wrapper around `sqlmodel.Field()` that silently discards `roles`, `validator`, and `api_alias` kwargs. This exists because the ORM models were copied from `autom8y-data` and retain `ApiField` calls, but the scheduling service does not run the autom8y-data schema factory. The stub allows copy-paste model code to work without modification.

The `# noqa: F405` comment appears on every model field line because `models/scheduling.py` and `models/shared.py` do `from autom8_scheduling.models._base import *` and static linters flag the star-import names as undefined.

### Structured Logging Event Name Convention

Log events use snake_case event name strings as the first positional argument — not interpolated messages:

```python
logger.info("scheduling_gate_rejected", extra={"reason": "global_disabled"})
logger.error("scheduling_gate_db_error", extra={"office_phone": ..., "error": ...})
```

The exception is `app.py` which uses the `autom8y_log` SDK and passes kwargs directly instead of `extra={}`.

### Idempotency Key Pattern

All write operations (book, cancel, reschedule) take an `idempotency_key` string. The engine checks for an existing appointment with that key before executing the operation, returning the existing result as an `Idempotent*` result type if found. The key is stored in the appointment's `config` JSON column.

### GCal Shadow Mode

The `gcal_shadow_mode` flag on `SchedulingConfig` controls whether GCal sync writes real events or only logs what it would do. This is a standard feature-flag pattern implemented as a bool on the config dataclass — no separate shadow-mode abstraction layer.

### Frozen Request Models

All Pydantic request body models use `model_config = ConfigDict(frozen=True, extra="forbid")`. This makes request objects immutable after parsing and rejects unknown fields, preventing data from silently being ignored.

---

## Naming Patterns

### Class Names

| Pattern | Examples |
|---------|---------|
| `*Engine` — stateful domain classes with injected session | `BookingEngine`, `AvailabilityEngine` |
| `*Error` — domain exceptions | `SchedulingEngineError`, `SchedulingGateError`, `SchedulingValidationError`, `ServiceError`, `NotFoundError` |
| `*Result` — frozen success dataclasses | `BookingResult`, `CancelResult`, `RescheduleResult`, `IdempotentBookingResult` |
| `*Config` — frozen configuration dataclasses | `SchedulingConfig` |
| `*Settings` — pydantic-settings classes | `SchedulingSettings` |
| `*Data` — Pydantic response payload models | `AvailabilityData`, `BookingResponseData`, `AppointmentListData`, `BusinessConfigData` |
| `*Response` / `*ErrorResponse` — envelope wrappers | `SchedulingResponse`, `SchedulingErrorResponse` |
| `*Request*` / `*RequestBody` — request input models | `BookingRequestBody`, `CancelRequestBody`, `ConfirmRequest`, `CancelNextRequest` |
| `*Service` — service-layer CRUD classes | `AppointmentService` |

### Function Names

- Route dependencies: `get_*` — `get_session`, `get_gcal_client`, `get_current_user`
- Response builders: `success_response`, `error_response`, `build_meta`
- Gate checks: `check_*` — `check_global_scheduling`, `check_business_scheduling`
- Exception handlers: `handle_*_exception`, `handle_*_error` — `handle_engine_exception`, `handle_validation_error`, `handle_scheduling_gate_error`
- Domain resolution: `resolve_*` — `resolve_business_offer`
- Domain queries: `is_*` — `is_slot_blocking`
- Utility/factory constructors: `get_settings`, `create_app`

### Variable Names

- Logger: `logger = logging.getLogger(__name__)` (all modules except `constants.py` which uses `_logger` to indicate private use)
- Settings singleton: `settings`
- Engine instances: `engine` (DB), `gcal_client`, `booking_engine`
- Session variables: `session`, `result`, `stmt`

### Constants

All constant frozensets use SCREAMING_SNAKE_CASE: `TERMINAL_STATUSES`, `KNOWN_STATUSES`, `CANCELLABLE_STATUSES`, `SLOT_BLOCKING_STATUSES`. String codes in error responses use SCREAMING_SNAKE_CASE: `SLOT_CONFLICT`, `BUSINESS_NOT_FOUND`, `INTERNAL_SERVER_ERROR`.

### File Naming

- Python files: `snake_case.py`
- Route sub-packages use plural nouns: `appointments/`
- Private/internal helpers prefix: `_base.py` (star-import backing module)
- No `utils.py` or `helpers.py` at the package root — helpers are co-located with the module that owns them (`booking_helpers.py` alongside `booking.py`)

### Package Naming

The top-level package is `autom8_scheduling` (underscore, not hyphen). The pyproject name is `autom8-scheduling` (hyphen). The SDK packages it depends on follow `autom8y_*` (note the `y` suffix distinguishing internal SDKs from the service itself).

### Annotation Patterns

Shared annotated types (PhoneNumber, ISODatetime, ISODate) are defined as module-level `Annotated[str, Field(...)]` aliases in `models/schemas.py` and imported throughout. This avoids repeating the E.164 regex and field description on every model that uses a phone number.

---

## Knowledge Gaps

1. **Booking engine internal logic**: `scheduling/booking.py` and `scheduling/engine.py` were not read in full. The internal slot generation algorithm and idempotency key storage mechanism are not fully documented here.

2. **Notifications sub-package**: `scheduling/notifications/` (6 files) was not read. SMS and email dispatch implementation details, template formats, and dispatch conditions are undocumented.

3. **GHL sync internals**: `scheduling/ghl_sync.py` was partially observed (log calls only). The GoHighLevel calendar sync protocol and error handling specifics are not documented.

4. **Test conventions**: The `tests/` directory was discovered but no test files were read. Test fixture patterns, mock patching conventions, and async test setup are not captured here.

5. **Health check implementation**: `health.py` was not read in full — only the error handling at lines 120-143 was observed. The dependency probe structure is undocumented.
