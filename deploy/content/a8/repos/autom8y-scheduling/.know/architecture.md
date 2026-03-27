---
domain: architecture
generated_at: "2026-03-27T19:56:20Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4557333"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

**Language**: Python 3.12. Build system: `hatchling`. Package manager: `uv`. Installed as `autom8-scheduling`, source at `src/autom8_scheduling/`.

**Directory layout** (47 source files total):

```
src/autom8_scheduling/
├── __init__.py                          # Package root (empty)
├── app.py                               # Application factory (create_app)
├── config.py                            # SchedulingSettings (pydantic-settings)
├── health.py                            # Health/readiness router (standalone)
├── api/
│   └── routes/
│       ├── __init__.py
│       ├── deps.py                      # FastAPI DI: get_session, get_gcal_client
│       ├── handlers.py                  # Shared handler utilities + gate checks
│       ├── metrics.py                   # Prometheus metric definitions
│       ├── internal.py                  # internal_router (/internal/v1)
│       ├── businesses.py                # businesses_router (/api/v1/businesses)
│       └── appointments/
│           ├── __init__.py              # appointments_router; re-exports for test patching
│           ├── availability.py          # GET /availability
│           ├── booking.py               # POST /book /cancel /cancel-next /reschedule /confirm
│           └── queries.py              # GET / (list), GET /{id} (detail)
├── models/
│   ├── __init__.py                      # Re-export layer for all models
│   ├── _base.py                         # SQLModel base imports, ApiField stub, validators
│   ├── envelopes.py                     # SchedulingResponse[T], SchedulingErrorResponse
│   ├── responses.py                     # Per-endpoint response data models
│   ├── schemas.py                       # Request body schemas (BookingRequestBody, etc.)
│   ├── scheduling.py                    # Domain ORM models (Appointment, Address, Hours, Employee, BusinessOffer)
│   └── shared.py                        # Read-only cross-domain models (Business, Lead)
├── scheduling/
│   ├── __init__.py
│   ├── booking.py                       # BookingEngine (book/cancel/reschedule)
│   ├── booking_helpers.py               # Helper functions for BookingEngine
│   ├── constants.py                     # TERMINAL_STATUSES, CANCELLABLE_STATUSES taxonomy
│   ├── engine.py                        # AvailabilityEngine (slot computation)
│   ├── exceptions.py                    # Exception hierarchy (SchedulingEngineError subtypes)
│   ├── gcal_overlay.py                  # GCalOverlay (post-availability busy-period filter)
│   ├── gcal_sync.py                     # dispatch_gcal_booking_sync / cancel / reschedule
│   ├── ghl_sync.py                      # dispatch_ghl_booking_sync / cancel / reschedule (stub)
│   ├── offer_resolution.py              # resolve_business_offer, SchedulingConfig dataclass
│   ├── reminder.py                      # ReminderEngine
│   ├── results.py                       # Frozen result dataclasses (BookingResult, CancelResult, etc.)
│   ├── validation.py                    # check_scheduling_readiness, compute_enable_prerequisites
│   ├── write_ops.py                     # validate_status_transition, emit_scheduling_audit_event
│   └── notifications/
│       ├── __init__.py
│       ├── config.py                    # Notification configuration
│       ├── dispatch.py                  # dispatch_booking_notifications, dispatch_cancel_notifications
│       ├── email_service.py             # NotificationEmailService
│       ├── schemas.py                   # EmailResult, SmsResult
│       ├── sms_service.py              # NotificationSmsService
│       └── templates.py                # format_*_sms, render_*_email
└── services/
    ├── __init__.py
    ├── base.py                          # BaseService[T], ServiceError hierarchy
    └── appointment.py                   # AppointmentService (CRUD layer for Appointment)
```

**Package purposes summary**:

| Package | Purpose | Hub/Leaf |
|---|---|---|
| `app` | Application factory, lifespan, middleware, router mounting | Hub (imports nearly everything) |
| `config` | Settings singleton (`SchedulingSettings`) via pydantic-settings | Leaf |
| `health` | Health/readiness/deps endpoints (health_router) | Leaf |
| `api/routes/deps` | FastAPI dependency injectors | Leaf (imported by all route modules) |
| `api/routes/handlers` | Shared handler utilities, gating checks, exception mapping | Hub (route files) |
| `api/routes/metrics` | Prometheus metric definitions | Leaf |
| `api/routes/appointments` | appointments_router + re-export package | Hub (route modules) |
| `api/routes/appointments/*` | Focused handler files (availability, booking, queries) | Leaf |
| `api/routes/businesses` | businesses_router | Leaf |
| `api/routes/internal` | internal_router (hidden from OpenAPI) | Leaf |
| `models/_base` | SQLModel base, ApiField stub, validators | Leaf |
| `models/envelopes` | `SchedulingResponse[T]`, `SchedulingErrorResponse` | Leaf (imported by all routes) |
| `models/responses` | Per-endpoint response data models | Leaf |
| `models/schemas` | Request body schemas | Leaf |
| `models/scheduling` | Domain ORM models (Appointment, Address, Hours, Employee, BusinessOffer) | Leaf |
| `models/shared` | Read-only cross-domain ORM models (Business, Lead) | Leaf |
| `scheduling/engine` | AvailabilityEngine (pure slot computation) | Hub (models, constants) |
| `scheduling/booking` | BookingEngine (book/cancel/reschedule) | Hub (models, exceptions, results) |
| `scheduling/offer_resolution` | `resolve_business_offer`, `SchedulingConfig` dataclass | Leaf |
| `scheduling/constants` | Status taxonomies (TERMINAL_STATUSES, etc.) | Leaf |
| `scheduling/exceptions` | Exception hierarchy | Leaf |
| `scheduling/results` | Frozen result dataclasses | Leaf |
| `scheduling/write_ops` | Status validation, audit logging, notification dispatch | Leaf |
| `scheduling/gcal_sync` | Google Calendar event sync | Leaf |
| `scheduling/ghl_sync` | GoHighLevel calendar event sync (stub, GHL_AVAILABLE=False) | Leaf |
| `scheduling/gcal_overlay` | GCal busy-period overlay for availability results | Leaf |
| `scheduling/reminder` | ReminderEngine (reminder-eligible queries) | Leaf |
| `scheduling/validation` | Readiness checks, prerequisite computation | Leaf |
| `scheduling/notifications/*` | SMS + email dispatch pipeline | Leaf |
| `services/base` | BaseService[T], ServiceError hierarchy | Leaf |
| `services/appointment` | AppointmentService (admin CRUD) | Leaf |

---

## Layer Boundaries

**Layer model (top-to-bottom import direction)**:

```
┌────────────────────────────────────────────────────────────┐
│  ENTRY / APPLICATION LAYER                                 │
│  app.py, config.py, health.py                              │
└───────────────────────────┬────────────────────────────────┘
                            │ imports
┌───────────────────────────▼────────────────────────────────┐
│  API SURFACE LAYER (routers / handlers / deps)             │
│  api/routes/{appointments, businesses, internal, handlers, │
│  deps, metrics}                                            │
└───────────────────────────┬────────────────────────────────┘
                            │ imports
┌───────────────────────────▼────────────────────────────────┐
│  DOMAIN ENGINE LAYER (scheduling)                          │
│  scheduling/{engine, booking, booking_helpers,             │
│   offer_resolution, validation, write_ops, results,        │
│   exceptions, constants, gcal_sync, ghl_sync,              │
│   gcal_overlay, reminder, notifications/*}                 │
└───────────────────────────┬────────────────────────────────┘
                            │ imports
┌───────────────────────────▼────────────────────────────────┐
│  DATA / MODEL LAYER                                        │
│  models/{scheduling, shared, _base, envelopes,             │
│   responses, schemas}                                      │
└───────────────────────────┬────────────────────────────────┘
                            │ reads via AsyncSession
┌───────────────────────────▼────────────────────────────────┐
│  INFRASTRUCTURE (MySQL via aiomysql + SQLAlchemy async)    │
│  External: GCal (autom8y-gcal SDK), GHL (stub)             │
│  Notifications: Twilio (SMS), SMTP (email)                 │
└────────────────────────────────────────────────────────────┘
```

**Import discipline observations**:
- `scheduling/` packages import from `models/` but never from `api/`.
- `api/routes/` packages import from both `scheduling/` and `models/`.
- `models/` packages import only from `models/_base` (no upward imports).
- `services/` imports from `models/` and `scheduling/write_ops` only. `AppointmentService` is not currently wired to any router — it exists but is not exposed via API endpoints.
- `scheduling/gcal_sync` and `ghl_sync` use `TYPE_CHECKING` guards to avoid circular imports with `offer_resolution.SchedulingConfig`.
- No circular dependencies were observed across the import graph.

**Cross-domain read-only boundary**: `models/shared.py` contains `Business` (table=`chiropractors`) and `Lead` (table=`leads`) SQLModel models marked READ-ONLY with explicit comments: "Do not write to this table from the scheduling service." These tables are owned by `autom8y-data`.

---

## Entry Points and API Surface

**Application factory**: `src/autom8_scheduling/app.py::create_app()` — called by uvicorn or test harness. Returns a configured `FastAPI` application.

**Startup sequence** (`lifespan`):
1. Phase 1: MySQL async engine pool (`aiomysql` via SQLAlchemy). Stored on `app.state.db_engine`.
2. Phase 1.5: SQLAlchemy OpenTelemetry instrumentation.
3. Phase 2: GCal client initialization (conditional on `SCHEDULING_GCAL_ENABLED`). Stored on `app.state.gcal_client`.
4. Phase 3: APScheduler placeholder (not yet implemented — noted as FR-25).

**Middleware stack** (inner-to-outer, applied in declaration order):
1. Error handling middleware — injects `x-request-id`, catches unhandled 500s.
2. CORSMiddleware — allows `autom8y.io`, `app.autom8y.io`, `localhost:3000` (local only).
3. `JWTAuthMiddleware` (autom8y-auth SDK) — excludes `/health`, `/ready`, `/health/deps`, `/docs`, `/openapi.json`.

**Router mount points**:
```
health_router    → /health, /ready, /health/deps   (no prefix)
appointments_router → /api/v1/appointments           (prefix)
businesses_router   → /api/v1/businesses             (prefix)
internal_router     → /internal/v1                   (prefix, include_in_schema=False)
```

**Complete endpoint inventory**:

| Method | Path | Handler | Summary |
|---|---|---|---|
| GET | `/health` | `health_check` | Liveness probe (no I/O, always 200) |
| GET | `/ready` | `readiness_check` | Readiness probe (MySQL check) |
| GET | `/health/deps` | `deps_check` | Dependency probe (MySQL + GCal) |
| GET | `/api/v1/appointments/availability` | `check_availability` | Check available appointment slots |
| GET | `/api/v1/appointments` | `get_appointments` | List appointments (paginated) |
| GET | `/api/v1/appointments/{appointment_id}` | `get_appointment_detail` | Get appointment detail |
| POST | `/api/v1/appointments/book` | `book_appointment` | Book a new appointment |
| POST | `/api/v1/appointments/cancel` | `cancel_appointment` | Cancel a specific appointment |
| POST | `/api/v1/appointments/cancel-next` | `cancel_next_appointment` | Cancel earliest upcoming appointment |
| POST | `/api/v1/appointments/reschedule` | `reschedule_appointment` | Reschedule appointment |
| POST | `/api/v1/appointments/confirm` | `confirm_appointment` | Confirm appointment (status transition) |
| GET | `/api/v1/businesses/{phone}/config` | `get_business_config` | Get business scheduling config |
| PATCH | `/api/v1/businesses/{phone}/config` | `update_scheduling_config` | Update business scheduling config |
| GET | `/api/v1/businesses/{phone}/readiness` | `check_readiness` | Check scheduling readiness |
| GET | `/api/v1/businesses/{phone}/hours` | `get_business_hours` | Get business operating hours |
| GET | `/internal/v1/reminders/eligible` | `get_reminder_eligible` | Get reminder-eligible appointments |
| POST | `/internal/v1/reminders/mark-sent` | `mark_reminder_sent` | Mark reminder as sent |
| GET | `/internal/v1/gcal/unsynced` | `get_unsynced_appointments` | Get unsynced GCal appointments |
| POST | `/internal/v1/gcal/reconcile` | `reconcile_gcal_events` | Backfill GCal events |

**Key exported interfaces between packages**:
- `appointments/__init__.py` re-exports `BookingEngine`, `AvailabilityEngine`, handler utilities, and `validate_status_transition` at the package level specifically to preserve `mock.patch` targets for tests.
- `offer_resolution.SchedulingConfig` (frozen dataclass) is the canonical config contract between the gate logic and all engine calls.

---

## Key Abstractions

The 8 most important types in this codebase, by centrality:

**1. `SchedulingResponse[T]`** — `src/autom8_scheduling/models/envelopes.py`
Generic Pydantic model. Every successful API response is wrapped in this envelope. `T` is the domain-specific data payload. Contains `data: T`, `meta: ResponseMeta`, and optional `pagination: PaginationMeta`. Consumed by all route handlers via `success_response()` in `handlers.py`.

**2. `SchedulingConfig`** — `src/autom8_scheduling/scheduling/offer_resolution.py`
Frozen dataclass. Single source of truth for per-business scheduling configuration resolved from `BusinessOffer`. Fields: `scheduling_enabled`, `buffer_minutes`, `gcal_enabled`, `gcal_shadow_mode`, `head_employee_id`, `master_cal_id`, `offer_id`, `business_offer_guid`, `duration_minutes`, `ghl_enabled`, `ghl_calendar_id`. Constructed via `SchedulingConfig.from_business_offer(offer)` or `SchedulingConfig.not_configured()`. Passed into `BookingEngine.book()`, `engine.compute_availability()`, and all GCal/GHL sync dispatchers.

**3. `BookingEngine`** — `src/autom8_scheduling/scheduling/booking.py`
Session-scoped class. Primary write path for appointment state changes. Methods: `book(...)`, `cancel(...)`, `reschedule(...)`. Returns typed result dataclasses (`BookingResult`, `CancelResult`, `RescheduleResult`, or their idempotent variants). Raises `SchedulingEngineError` subtypes on failure. Enforces idempotency via `idempotency_key` lookups.

**4. `AvailabilityEngine`** — `src/autom8_scheduling/scheduling/engine.py`
Session-scoped class. Pure computation — no writes, no side effects. Method: `compute_availability(office_phone, start_date, end_date, employee_id, buffer_minutes, offer_duration_minutes)`. Returns dict with `days`, `timezone`, `appt_duration_minutes`. Uses three-tier duration resolution: offer-level > business-level > default (30 min). Max lookforward window: 14 days.

**5. `SchedulingEngineError` hierarchy** — `src/autom8_scheduling/scheduling/exceptions.py`
Exception class hierarchy. Base: `SchedulingEngineError`. Subtypes: `BookingConflict` (409 or 404), `AppointmentNotFound` (404), `AppointmentNotCancellable` (409), `AppointmentAlreadyCancelled` (200 idempotent), `RescheduleConflict` (409, extends `BookingConflict`). Separate: `SchedulingGateError` (403, registered as FastAPI exception handler). HTTP mapping enforced in `handlers.handle_engine_exception()`.

**6. `Appointment`** — `src/autom8_scheduling/models/scheduling.py`
SQLModel ORM class. Table: `appointments`. Key fields: `appointment_id` (PK, alias `id`), `office_phone` (FK to `chiropractors`), `phone` (lead phone, join key), `start_datetime`/`end_datetime` (VARCHAR, ISO 8601 strings), `status` (nullable VARCHAR), `event_id` (GCal binding), `ghl_event_id` (GHL binding), `config` (JSON blob, stores notification timestamps).

**7. `TERMINAL_STATUSES`** — `src/autom8_scheduling/scheduling/constants.py`
`frozenset[str]`. The deny-list strategy for slot blocking: any status NOT in this set blocks a slot. Derived from a production audit of 151,028 appointments across 16 statuses from 27 source systems. Value: `{"patient", "cancelled", "no_show", "no-show", "paid", "finished", "system"}`. Consumed by `AvailabilityEngine`, `BookingEngine`, `ReminderEngine`, and query handlers.

**8. `VALID_TRANSITIONS`** — `src/autom8_scheduling/scheduling/write_ops.py`
`dict[str, frozenset[str]]`. Status machine definition. Terminal statuses (completed, cancelled, no_show, etc.) are absent from this dict — any missing key = terminal. Enforced via `validate_status_transition(current, new)`. Used in the `confirm` endpoint and `AppointmentService`.

**Design patterns observed**:
- **Envelope pattern**: All API responses wrapped in `SchedulingResponse[T]` or `SchedulingErrorResponse`. Never raw dicts or bare Pydantic models.
- **Fire-and-forget side effects**: Post-commit notifications (SMS, email, GCal, GHL) dispatched via `asyncio.gather(*coros, return_exceptions=True)`. Failure never rolls back the booking.
- **Idempotency via key lookup**: `booking.py` checks `idempotency_key` before inserting; returns `IdempotentBookingResult` on replay.
- **Frozen dataclass results**: `BookingResult`, `CancelResult`, `RescheduleResult` and their `Idempotent*` variants are frozen dataclasses, not Pydantic models. `isinstance()` checks distinguish fresh vs replay at the handler layer.
- **Re-export for test patching**: `appointments/__init__.py` re-exports all patchable names so `mock.patch("autom8_scheduling.api.routes.appointments.BookingEngine")` works correctly even after handler split.
- **Global kill switch**: `SCHEDULING_ENABLED` env var checked in `check_global_scheduling()`. Any value of `"false"`, `"0"`, or `"no"` disables all scheduling endpoints with HTTP 503.

---

## Data Flow

**Primary booking flow**:
```
HTTP POST /api/v1/appointments/book
  → JWTAuthMiddleware (validate JWT)
  → error_handling_middleware (inject request_id)
  → book_appointment() handler (booking.py)
      → check_global_scheduling() — env var kill switch
      → check_business_scheduling(office_phone, session) — gate check
          → resolve_business_offer(session, office_phone) → SchedulingConfig
      → BookingEngine(session).book(...)
          → Idempotency key check (SELECT appointments WHERE idempotency_key=...)
          → Conflict check (SELECT appointments WHERE slot overlaps)
          → INSERT appointment
          → Returns BookingResult | IdempotentBookingResult | raises SchedulingEngineError
      → session.commit()
      → asyncio.gather(
            dispatch_booking_notifications(...),   # SMS + email
            dispatch_gcal_booking_sync(...),       # GCal (if gcal_enabled)
            dispatch_ghl_booking_sync(...),        # GHL (currently no-op, GHL_AVAILABLE=False)
         return_exceptions=True)
      → success_response(BookingResponseData, request, 201)
```

**Availability flow**:
```
HTTP GET /api/v1/appointments/availability?office_phone=...&start_date=...&end_date=...
  → check_availability() handler (availability.py)
      → check_global_scheduling()
      → check_business_scheduling() → SchedulingConfig (buffer_minutes, duration_minutes, gcal_enabled)
      → AvailabilityEngine(session).compute_availability(...)
          → resolve_timezone (SELECT addresses.timezone)
          → resolve_appt_duration (3-tier: offer > business > default 30min)
          → get_hours_by_day (SELECT hours WHERE type='office')
          → get_scheduled_appointments (SELECT appointments WHERE NOT TERMINAL)
          → generate candidate slots, subtract conflicts and buffer
          → Returns dict with days/slots
      → GCalOverlay.filter_availability(...) [if gcal_enabled and gcal_client]
          → Fetch GCal busy periods for calendar_id
          → Subtract busy periods from available slots
      → success_response(AvailabilityData, request, 200)
```

**Configuration sources**:
- Application config: `SCHEDULING_*` env vars → `SchedulingSettings` (`config.py`), cached via `@lru_cache`.
- Per-business scheduling config: resolved from `business_offers` table via `resolve_business_offer()`. `disabled` field is inverted: `disabled=False` → `scheduling_enabled=True`.
- GCal enabled: `business_offers.master_cal_id IS NOT NULL AND non-empty`.
- GHL enabled: `business_offers.ghl_calendar_id IS NOT NULL AND non-empty`.
- Notification credentials: `TWILIO_*` env vars (checked at runtime in `NotificationSmsService.is_configured()`). SMTP credentials for email via `NotificationEmailService.is_configured()`.

**External service integrations**:
- **MySQL** (aiomysql + SQLAlchemy async): Primary datastore. Async connection pool configured in lifespan. Session injected per-request via `get_session` dependency.
- **Google Calendar** (`autom8y-gcal` SDK, optional): Initialized as singleton on `app.state.gcal_client`. Used for availability overlay and event sync. Gracefully degrades if SDK not installed or config missing.
- **GoHighLevel** (no SDK yet): `ghl_sync.py` exists with dispatch function stubs, but `GHL_AVAILABLE = False` — all calls silently no-op. Schema columns (`ghl_event_id`, `ghl_calendar_id`) are present in database.
- **Twilio** (via `NotificationSmsService`): SMS notifications. Checked for configuration at dispatch time, not at startup.
- **SMTP email** (via `NotificationEmailService`): Email notifications. Same pattern.
- **OpenTelemetry** (`autom8y-telemetry[fastapi,otlp]`): Traces via `instrument_app`, SQLAlchemy spans, custom `scheduling.*` spans in engines and notification dispatch. Side effects recorded via `record_side_effect()`.
- **Prometheus** (`prometheus-client`): Counters and histograms defined in `metrics.py`. Exposed via standard Prometheus endpoint.

---

## Knowledge Gaps

1. **`src/autom8_scheduling/scheduling/booking.py` and `booking_helpers.py`**: These files were identified but not read in full. `BookingEngine` internals (idempotency key storage schema, conflict check SQL, employee assignment logic) are documented at the interface level only.

2. **`src/autom8_scheduling/scheduling/gcal_overlay.py`**: GCalOverlay's caching mechanism (`last_cache_hit` attribute) and the GCal busy period fetch logic were not read in detail.

3. **`src/autom8_scheduling/scheduling/reminder.py`**: `ReminderEngine.get_eligible()` and `mark_sent()` implementation details not read.

4. **`src/autom8_scheduling/scheduling/validation.py`**: `check_scheduling_readiness()` and `compute_enable_prerequisites()` internals not read beyond their call signatures.

5. **`src/autom8_scheduling/services/appointment.py`** (beyond line 60): `AppointmentService.update()` and `partial_update()` implementation details not confirmed. Notably, `AppointmentService` is not wired to any router — it exists but is not reachable via HTTP.

6. **`src/autom8_scheduling/models/responses.py` and `schemas.py`**: Not read — field-level schema details for all response/request types are undocumented.

7. **`tests/`**: Test file structure not explored. No information on test organization, fixture patterns, or coverage targets.

8. **`Dockerfile` and `justfile`**: Deployment and operational scripts not read.
