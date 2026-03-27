---
domain: architecture
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "39376b6"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

## Package Structure

The repository is a standalone Python service named `autom8-scheduling`, packaged under `src/autom8_scheduling/` using the `src` layout. The package was extracted from `autom8y-data` per migration ticket `TDD-SCHED-EXTRACT-001`. The build backend is `hatchling`.

**Top-level layout:**

```
autom8y-scheduling/
├── src/
│   └── autom8_scheduling/          # Main package (31 Python files)
│       ├── __init__.py             # Package entry, version = "0.1.0"
│       ├── app.py                  # Application factory (create_app)
│       ├── config.py               # SchedulingSettings (pydantic-settings)
│       ├── health.py               # Health router (/health, /ready, /health/deps)
│       ├── api/
│       │   ├── __init__.py
│       │   └── routes/
│       │       ├── __init__.py     # All scheduling HTTP endpoints (1,019 lines)
│       │       └── deps.py         # FastAPI dependencies (get_session, get_gcal_client)
│       ├── models/
│       │   ├── __init__.py         # Re-export layer (Address, Appointment, Business, etc.)
│       │   ├── _base.py            # SQLModel base, validators, ApiField stub
│       │   ├── scheduling.py       # Owned domain models (Appointment, Address, Hours, Employee, BusinessOffer)
│       │   ├── schemas.py          # Pydantic request schemas (BookingRequestBody, etc.)
│       │   └── shared.py           # Read-only copies: Business (table=chiropractors), Lead
│       ├── scheduling/
│       │   ├── __init__.py
│       │   ├── booking.py          # BookingEngine (book, cancel, reschedule)
│       │   ├── constants.py        # SLOT_BLOCKING_STATUSES, CANCELLABLE_STATUSES
│       │   ├── engine.py           # AvailabilityEngine + _parse_appointment_datetime
│       │   ├── gcal_overlay.py     # GCalOverlay (freebusy subtraction + in-memory cache)
│       │   ├── gcal_sync.py        # GCal write-path dispatchers (book/cancel/reschedule)
│       │   ├── reminder.py         # ReminderEngine (get_eligible, mark_sent)
│       │   ├── validation.py       # check_scheduling_readiness
│       │   ├── write_ops.py        # validate_status_transition, emit_scheduling_audit_event
│       │   └── notifications/
│       │       ├── __init__.py
│       │       ├── config.py       # NotificationConfig (SendGrid + Twilio env vars)
│       │       ├── dispatch.py     # dispatch_booking_notifications, dispatch_cancel_notifications
│       │       ├── email_service.py # NotificationEmailService (SendGrid)
│       │       ├── schemas.py      # SmsResult, EmailResult, NotificationType, NotificationChannel
│       │       ├── sms_service.py  # NotificationSmsService (Twilio)
│       │       └── templates.py    # SMS and email template functions
│       └── services/
│           ├── __init__.py
│           ├── base.py             # BaseService[T], ServiceError hierarchy
│           └── appointment.py      # AppointmentService (CRUD)
├── tests/                          # 9 test files
│   ├── conftest.py
│   ├── test_health.py
│   ├── test_proxy.py
│   ├── test_sprint2_instrumentation.py
│   ├── test_write_ops.py
│   └── golden_traces/
│       ├── conftest.py
│       ├── serializer.py
│       ├── span_tree.py
│       └── test_golden_traces.py
├── Dockerfile
├── justfile
├── pyproject.toml
└── secretspec.toml
```

**Organizational principle:** Hybrid layer + domain. Top-level splits `api/`, `models/`, `scheduling/`, `services/` (layers). The `scheduling/` package groups all domain logic (booking, availability, GCal, reminders, notifications) as sub-modules.

**File counts:** 31 source `.py` files, 9 test `.py` files.

## Dependency Graph

### Internal Module Dependencies

```
app.py
  → config.py  (SchedulingSettings, get_settings)
  → health.py  (health_router)
  → api/routes/__init__.py  (scheduling_router)
  → [autom8y_auth, autom8y_telemetry, autom8y_log, autom8y_config]  (SDKs)

api/routes/__init__.py  (scheduling_router)
  → api/routes/deps.py  (get_session, get_gcal_client)
  → models/__init__.py  (Address, Appointment, Business, Employee)
  → models/schemas.py  (request body types)
  → scheduling/booking.py  (BookingEngine)
  → scheduling/engine.py  (AvailabilityEngine, _parse_appointment_datetime)
  → scheduling/gcal_overlay.py  (GCalOverlay)
  → scheduling/gcal_sync.py  (dispatch_gcal_*)
  → scheduling/reminder.py  (ReminderEngine)
  → scheduling/validation.py  (check_scheduling_readiness)
  → [autom8y_core.errors, prometheus_client]

api/routes/deps.py
  → [autom8y_auth]  (BaseClaims, require_auth)

models/__init__.py
  → models/scheduling.py  (Address, Appointment, BusinessOffer, Employee, Hours)
  → models/shared.py  (Business, Lead)

models/scheduling.py
  → models/_base.py  (* import)

models/shared.py
  → models/_base.py  (* import)

scheduling/engine.py
  → models/__init__.py  (Address, Appointment, Business, Employee, Hours)
  → scheduling/constants.py  (SLOT_BLOCKING_STATUSES)
  → [autom8y_core.errors, autom8y_telemetry.conventions, opentelemetry, prometheus_client]

scheduling/booking.py
  → models/__init__.py  (Appointment, Business)
  → scheduling/constants.py  (CANCELLABLE_STATUSES, SLOT_BLOCKING_STATUSES)
  → scheduling/engine.py  (_parse_appointment_datetime)
  → [autom8y_telemetry.conventions, autom8y_telemetry.effects, opentelemetry, prometheus_client]

scheduling/gcal_overlay.py
  → models/__init__.py  (BusinessOffer)
  → [autom8y_telemetry.conventions, opentelemetry, autom8y_gcal (optional)]

scheduling/gcal_sync.py
  → models/__init__.py  (Appointment, BusinessOffer)
  → [autom8y_telemetry.conventions, autom8y_telemetry.effects, opentelemetry, autom8y_gcal (optional)]

scheduling/reminder.py
  → models/__init__.py  (Address, Appointment, Business, Employee)
  → scheduling/constants.py  (SLOT_BLOCKING_STATUSES)
  → scheduling/engine.py  (_parse_appointment_datetime)

scheduling/validation.py
  → models/__init__.py  (Address, Business, Hours)

scheduling/write_ops.py
  (no internal imports — pure domain logic)

scheduling/notifications/dispatch.py
  → models/__init__.py  (Appointment)
  → scheduling/notifications/email_service.py
  → scheduling/notifications/sms_service.py
  → scheduling/notifications/schemas.py
  → scheduling/notifications/templates.py
  → [autom8y_telemetry.conventions, autom8y_telemetry.effects, opentelemetry]

scheduling/notifications/email_service.py
  → scheduling/notifications/config.py
  → scheduling/notifications/schemas.py
  → [autom8y_http]

scheduling/notifications/sms_service.py
  → scheduling/notifications/config.py
  → scheduling/notifications/schemas.py
  → [autom8y_http]

services/appointment.py
  → models/__init__.py  (Appointment, Business, Lead)
  → models/_base.py  (E164_PHONE_PATTERN, utc_now)
  → scheduling/write_ops.py  (emit_scheduling_audit_event, validate_status_transition)
  → services/base.py  (BaseService, error hierarchy)
```

**Dependency direction:** Strictly one-way. `api/` imports from `scheduling/` and `models/`, `scheduling/` imports from `models/`, `services/` imports from both `scheduling/` and `models/`. No circular dependencies observed.

### External Dependencies (from `pyproject.toml`)

**Runtime:**
- `fastapi>=0.109.0` — web framework
- `uvicorn[standard]>=0.27.0` — ASGI server
- `pydantic>=2.0.0`, `pydantic-settings>=2.0.0` — data validation + config
- `sqlalchemy>=2.0.23` — async ORM
- `sqlmodel>=0.0.14` — SQLAlchemy + Pydantic model fusion
- `aiomysql>=0.2.0` — async MySQL driver
- `prometheus-client>=0.20.0` — Prometheus metrics
- `opentelemetry-instrumentation-sqlalchemy>=0.41b0` — OTel SQLAlchemy instrumentation
- `autom8y-config>=1.2.1` — base settings (Autom8yBaseSettings)
- `autom8y-log>=0.5.6` — structured logging (get_logger)
- `autom8y-core>=1.1.1` — domain errors (TimezoneNotConfiguredError)
- `autom8y-auth[observability]>=1.1.0` — JWT middleware (JWTAuthMiddleware, require_auth)
- `autom8y-telemetry[fastapi,otlp]>=0.6.0` — OTel instrumentation + conventions
- `autom8y-http>=0.3.0` — HTTP client (Autom8yHttpClient, used for Twilio/SendGrid)

**Optional (gcal extra):**
- `autom8y-gcal>=0.1.0` — Google Calendar SDK (GCalClient, EventCreate, EventDateTime)

**Dev:**
- `pytest`, `pytest-asyncio`, `pytest-cov`, `httpx`, `mypy`, `ruff`, `respx`, `autom8y-telemetry[conventions]`

All `autom8y-*` packages resolve from a private AWS CodeArtifact index at `autom8y-696318035277.d.codeartifact.us-east-1.amazonaws.com/pypi/autom8y-python/simple/`.

## Layer Architecture

The service follows a 4-layer architecture:

| Layer | Packages | Responsibility |
|---|---|---|
| **Presentation / API** | `api/routes/`, `health.py` | HTTP endpoints, request parsing, response formatting, Prometheus metrics, auth middleware wiring |
| **Domain / Business Logic** | `scheduling/` | Availability computation, atomic booking, GCal sync, reminders, notification dispatch, status state machine |
| **Data Models** | `models/` | SQLModel ORM entities, Pydantic request schemas, validation utilities |
| **Infrastructure** | `services/`, `config.py`, `app.py` | CRUD service base, application factory, lifespan management, DB pool, GCal client lifecycle |

**Cross-cutting concerns:**
- `app.py` wires all layers at startup (lifespan: MySQL pool → GCal client → yield)
- `scheduling/write_ops.py` provides shared domain operations used by both `scheduling/booking.py` (domain) and `services/appointment.py` (infrastructure service)
- Observability (OTel spans, Prometheus histograms/counters) is scattered across both domain and presentation layers, which is intentional per `OBS-SCHED-001`

**Data flow between layers:**
```
HTTP Request
  → Middleware (JWT auth, CORS, error handling)  [app.py]
  → Route handler  [api/routes/__init__.py]
  → Dependency injection (get_session → AsyncSession)  [api/routes/deps.py]
  → Engine/Validator  [scheduling/booking.py, scheduling/engine.py]
  → SQLAlchemy queries  [models/ ORM entities]
  → MySQL (nhc_production)
  → Response dict  →  JSON response
  → Post-commit fire-and-forget: notifications (Twilio, SendGrid) + GCal sync
```

## Entry Points and Public API

### HTTP Entry Point
The ASGI app is created by `create_app()` in `src/autom8_scheduling/app.py` and served by uvicorn:
```
uvicorn autom8_scheduling.app:create_app --factory --host 0.0.0.0 --port 8000
```
Local dev port is 5140 (per `justfile`).

### Health Endpoints (`health_router` — no auth)

| Endpoint | Method | Description | Call Chain |
|---|---|---|---|
| `GET /health` | Liveness | No I/O, always 200 | `health_check()` → `liveness_response()` |
| `GET /ready` | Readiness | MySQL SELECT 1 | `readiness_check(request)` → `_check_mysql(engine)` → SQLAlchemy |
| `GET /health/deps` | Dependency probe | MySQL + GCal | `deps_check(request)` → `_check_mysql()` + `_check_gcal()` |

### Scheduling Endpoints (`scheduling_router`, prefix `/api/v1/scheduling`, JWT auth required)

| Endpoint | Method | Description | Key Call Chain |
|---|---|---|---|
| `GET /availability` | Query | Compute available slots | `check_availability` → `_check_business_scheduling` → `AvailabilityEngine.compute_availability` → `[GCalOverlay.filter_availability]` |
| `POST /book` | Mutation | Atomic check + book | `book_appointment` → `BookingEngine.book` → `session.commit()` → `dispatch_booking_notifications` + `dispatch_gcal_booking_sync` |
| `POST /cancel` | Mutation | Soft-delete | `cancel_appointment` → `BookingEngine.cancel` → `session.commit()` → `dispatch_cancel_notifications` + `dispatch_gcal_cancel_sync` |
| `POST /reschedule` | Mutation | Cancel old + book new | `reschedule_appointment` → `BookingEngine.reschedule` → `session.commit()` → notifications + `dispatch_gcal_reschedule_sync` |
| `GET /appointments` | Query | List active appointments for lead | `get_appointments` → SQLAlchemy JOIN (Appointment + Employee) |
| `GET /readiness` | Query | Business scheduling prereqs | `check_readiness` → `check_scheduling_readiness` → queries Business, Address, Hours |
| `PATCH /config` | Mutation | Per-business scheduling config | `update_scheduling_config` → Business update → `session.commit()` |
| `GET /reminders/eligible` | Query | Appointments needing SMS reminders | `get_reminder_eligible` → `ReminderEngine.get_eligible` → 4-table JOIN |
| `POST /reminders/mark-sent` | Mutation | Mark reminder sent in config JSON | `mark_reminder_sent` → `ReminderEngine.mark_sent` → Appointment.config update |
| `GET /gcal/unsynced` | Query | Appointments missing GCal event_id | `get_unsynced_appointments` → Appointment WHERE event_id IS NULL |
| `POST /gcal/reconcile` | Mutation | Backfill GCal events | `reconcile_gcal_events` → loop → `dispatch_gcal_booking_sync` |

### Global Feature Flag
`SCHEDULING_ENABLED` env var — checked at the start of every mutating and query endpoint via `_check_global_scheduling()`. Returns HTTP 503 when disabled.

## Data Flow

### Primary Request Flow (Booking)

```
POST /api/v1/scheduling/book (BookingRequestBody JSON)
  1. JWTAuthMiddleware validates Bearer token (autom8y-auth SDK)
  2. book_appointment() extracts body: {office_phone, lead_phone, start/end_datetime, idempotency_key}
  3. _check_global_scheduling() — reads SCHEDULING_ENABLED env var
  4. _check_business_scheduling(office_phone, session) — SELECT chiropractors WHERE office_phone = ?
  5. BookingEngine.book():
     a. _find_by_idempotency_key() — full table scan on appointments.config JSON
     b. _resolve_business() — SELECT chiropractors
     c. _check_overlap_nonlocking() — SELECT appointments WHERE status IN slot_blocking
     d. _select_overlapping() — SELECT ... WITH FOR UPDATE (TOCTOU defense)
     e. INSERT INTO appointments (status="scheduled", config={idempotency_key, booking_source, booked_at})
     f. session.flush()
  6. Return result dict {status, appointment_id, start/end_datetime}
  7. session.commit()
  8. asyncio.gather([dispatch_booking_notifications, [dispatch_gcal_booking_sync]])
     — fire-and-forget, exceptions absorbed
```

### Availability Computation Flow

```
GET /api/v1/scheduling/availability?office_phone=...&start_date=...&end_date=...
  1. AvailabilityEngine.compute_availability():
     a. Validate date range (max 14 days)
     b. SELECT addresses.timezone
     c. SELECT chiropractors.appt_duration
     d. [Optional] SELECT employees.id (employee filter)
     e. SELECT hours (type="office")
     f. SELECT appointments (status IN slot_blocking) — scan for overlap detection
  2. Generate candidate slots from hours blocks
  3. Filter slots with overlap detection (UTC conversion)
  4. [Optional] GCalOverlay.filter_availability():
     a. SELECT business_offers.master_cal_id
     b. _fetch_freebusy() → in-memory cache check → autom8y_gcal.free_busy.query()
     c. _subtract_busy_periods() — removes overlapping slots (subtractive only)
  5. Return {office_phone, timezone, appt_duration_minutes, days: [{date, day_name, slots}]}
```

### Notification Dispatch Flow (Post-Commit)

```
After session.commit() on book/cancel/reschedule:
  asyncio.gather([
    dispatch_booking_notifications() or dispatch_cancel_notifications():
      - Check appointments.config for existing sent timestamp (idempotency)
      - SMS: NotificationSmsService.send() → Autom8yHttpClient → POST api.twilio.com/Messages.json
      - Email: NotificationEmailService.send() → Autom8yHttpClient → POST api.sendgrid.com/v3/mail/send
      - Write sent timestamps back to appointments.config JSON
    [dispatch_gcal_booking_sync() / dispatch_gcal_cancel_sync() / dispatch_gcal_reschedule_sync()]:
      - Resolve calendar_id from business_offers.master_cal_id
      - gcal_client.events.insert/delete() via autom8y_gcal SDK
      - Write/clear appointments.event_id
  ])
```

### Key Data Models

| Model | Table | Ownership | Purpose |
|---|---|---|---|
| `Appointment` | `appointments` | Owned | Core booking record. Primary key: `appointment_id` (int). `config` JSON blob stores idempotency keys, cancellation metadata, notification timestamps. `event_id` stores GCal event ID. |
| `Address` | `addresses` | Owned | Business location. `timezone` (IANA string) is critical — drives all datetime conversions. |
| `Hours` | `hours` | Owned | Operating hours per day. `day` = ISO 8601 int (1=Mon). `available_times` = JSON string of explicit slot times or null for open/close-based generation. |
| `Employee` | `employees` | Owned | Staff member. `calendar_id`, `enabled` flag for filtering. |
| `BusinessOffer` | `business_offers` | Owned | Contains `master_cal_id` — the Google Calendar ID resolved for GCal overlay/sync. |
| `Business` | `chiropractors` | Read-only copy | Per-business config: `scheduling_enabled`, `buffer_minutes`, `gcal_enabled`, `gcal_shadow_mode`. |
| `Lead` | `leads` | Read-only copy | Lead entity. Used only for FK validation in `AppointmentService`. |

**Datetime representation:** Appointment `start_datetime` and `end_datetime` are VARCHAR columns, not TIMESTAMP. They are parsed via `_parse_appointment_datetime()` which handles ISO 8601 (with or without tzinfo) and `%Y-%m-%d %H:%M:%S` format.

## Configuration and Environment

### Configuration Classes

**`SchedulingSettings`** (`src/autom8_scheduling/config.py`)
- Inherits from `Autom8yBaseSettings` (autom8y-config SDK)
- `env_prefix = "SCHEDULING_"`
- Singleton via `@lru_cache` on `get_settings()`

| Field | Env Var | Default | Notes |
|---|---|---|---|
| `db_host` | `SCHEDULING_DB_HOST` | `localhost` | MySQL hostname |
| `db_port` | `SCHEDULING_DB_PORT` | `3306` | MySQL port |
| `db_username` | `SCHEDULING_DB_USERNAME` | `root` | |
| `db_password` | `SCHEDULING_DB_PASSWORD` | `root` | Required in prod |
| `db_name` | `SCHEDULING_DB_NAME` | `nhc_production` | |
| `database_url` | `SCHEDULING_DATABASE_URL` | `None` | Overrides component fields if set |
| `pool_size` | `SCHEDULING_POOL_SIZE` | `3` | ADR-POOL-001 |
| `max_overflow` | `SCHEDULING_MAX_OVERFLOW` | `5` | ADR-POOL-001 |
| `gcal_enabled` | `SCHEDULING_GCAL_ENABLED` | `False` | Gates GCal client init |
| `gcal_shadow_mode` | `SCHEDULING_GCAL_SHADOW_MODE` | `True` | When true: GCal overlay observes but does not filter |
| `gcal_impersonation_email` | `SCHEDULING_GCAL_IMPERSONATION_EMAIL` | `support@...` | GCal DWD impersonation target |
| `gcal_timeout` | `SCHEDULING_GCAL_TIMEOUT` | `30.0` | GCal API timeout |
| `cors_origins` | `SCHEDULING_CORS_ORIGINS` | `[autom8y.io, app.autom8y.io]` | Localhost added automatically in local env |
| `service_name` | `SCHEDULING_SERVICE_NAME` | `scheduling` | |
| `service_version` | `SCHEDULING_SERVICE_VERSION` | `0.1.0` | |

`AUTOM8Y_ENV` is resolved by `Autom8yBaseSettings` — values: `local`, `staging`, `production`, `test`.

**`NotificationConfig`** (`src/autom8_scheduling/scheduling/notifications/config.py`)
- Does NOT use `SCHEDULING_` prefix
- Singleton via `@lru_cache`

| Field | Env Var | Secret |
|---|---|---|
| `SENDGRID_API_KEY` | `SENDGRID_API_KEY` | Yes |
| `NOTIFICATION_EMAIL_FROM_ADDRESS` | `NOTIFICATION_EMAIL_FROM_ADDRESS` | No |
| `TWILIO_ACCOUNT_SID` | `TWILIO_ACCOUNT_SID` | No |
| `TWILIO_AUTH_TOKEN` | `TWILIO_AUTH_TOKEN` | Yes |

### Feature Flags

| Flag | Mechanism | Effect |
|---|---|---|
| `SCHEDULING_ENABLED` | env var (os.getenv) | Global kill switch for all scheduling endpoints |
| `scheduling_enabled` | `Business.scheduling_enabled` DB column | Per-business gating (HTTP 403 if false) |
| `SCHEDULING_GCAL_ENABLED` | `SchedulingSettings.gcal_enabled` | Gates GCal client initialization in lifespan |
| `gcal_enabled` | `Business.gcal_enabled` DB column | Per-business GCal overlay/sync activation |
| `gcal_shadow_mode` | `Business.gcal_shadow_mode` DB column | When true, GCal overlay runs but does not remove slots |

### Secrets Management

Per `secretspec.toml`: secrets are injected as env vars. Production secrets (DB password, DATABASE_URL, Google SA key JSON) come from AWS Secrets Manager. `AUTH_DEV_MODE` disables JWT auth in local environment only.

### OpenAPI Documentation

Docs (`/docs`) are disabled in `AUTOM8Y_ENV=production`. ReDoc is always disabled.

## Infrastructure Integration

### MySQL (Primary Data Store)

- **Driver:** `aiomysql` (async)
- **ORM:** SQLAlchemy 2.x async + SQLModel
- **Connection:** `AsyncEngine` created in lifespan Phase 1 via `create_async_engine()`
- **Connection string:** `mysql+aiomysql://user:pass@host:port/nhc_production`
- **Pool config:** `pool_size=3`, `max_overflow=5`, `pool_pre_ping=True`, `pool_recycle=3600`
- **Session lifecycle:** Per-request `AsyncSession` via `async_sessionmaker`, yielded from `get_session()` dependency
- **Instrumentation:** `SQLAlchemyInstrumentor` attached to `engine.sync_engine` in lifespan Phase 1.5

**Tables accessed:**

| Table | SQLModel Class | Access Pattern |
|---|---|---|
| `appointments` | `Appointment` | Read + Write (INSERT, UPDATE status/config/event_id) |
| `addresses` | `Address` | Read-only (timezone resolution) |
| `hours` | `Hours` | Read-only (availability computation) |
| `employees` | `Employee` | Read-only (name resolution, enabled check) |
| `business_offers` | `BusinessOffer` | Read-only (GCal calendar_id resolution) |
| `chiropractors` | `Business` | Read-mostly (scheduling_enabled, buffer_minutes config write via PATCH /config) |
| `leads` | `Lead` | Read-only (FK validation in AppointmentService) |

### Google Calendar API

- **SDK:** `autom8y-gcal` (optional dependency, `[gcal]` extra)
- **Client type:** `GCalClient` from `autom8y_gcal.client`
- **Config:** `GCalConfig(impersonation_target=..., timeout=...)` from `autom8y_gcal.config`
- **Auth:** Service Account + Domain-Wide Delegation (DWD). Credentials from `GOOGLE_SA_KEY_JSON` (SSM) or `GOOGLE_APPLICATION_CREDENTIALS` file path
- **Lifecycle:** Client initialized in lifespan Phase 2 via `client.__aenter__()`, torn down via `__aexit__()`; stored on `app.state.gcal_client`
- **Operations used:**
  - `gcal_client.free_busy.query(calendar_ids, time_min, time_max)` — availability overlay
  - `gcal_client.events.insert(calendar_id, EventCreate)` — booking sync
  - `gcal_client.events.delete(calendar_id, event_id)` — cancel sync
- **Graceful degradation:** `ImportError` and runtime errors both result in `None` client; all call sites check for `None` before invoking

**Calendar ID resolution:** `BusinessOffer.master_cal_id` column in `business_offers` table.

**In-memory FreeBusy cache:** `_FreeBusyCache` (process-local, not shared across workers), 5-minute TTL, max 1000 entries, LRU-style eviction. Cache key: `gcal:freebusy:{calendar_id}:{sha256(time_min:time_max)[:12]}`.

### Twilio (SMS Notifications)

- **Integration pattern:** Direct HTTP REST via `autom8y_http.Autom8yHttpClient`
- **Endpoint:** `POST https://api.twilio.com/2010-04-01/Accounts/{ACCOUNT_SID}/Messages.json`
- **Auth:** HTTP Basic (account_sid, auth_token)
- **Module:** `scheduling/notifications/sms_service.py`
- **Failure mode:** Errors absorbed; `SmsResult.success=False` returned. Timestamps written to `appointments.config` prevent duplicate sends.

### SendGrid (Email Notifications)

- **Integration pattern:** Direct HTTP REST via `autom8y_http.Autom8yHttpClient`
- **Endpoint:** `POST https://api.sendgrid.com/v3/mail/send`
- **Auth:** Bearer token (SENDGRID_API_KEY)
- **Module:** `scheduling/notifications/email_service.py`
- **Retry:** 2 attempts for 5xx and timeout errors, 1-second delay between attempts
- **Destination:** Business email address (business alert to provider, not patient-facing for email)

### OpenTelemetry (Distributed Tracing)

- **SDK:** `autom8y-telemetry[fastapi,otlp]`
- **Initialization:** `instrument_app(app, InstrumentationConfig(service_name="scheduling"))` in `create_app()`
- **Tracers used:** `autom8y.scheduling` (booking, availability, notifications), `autom8y.gcal` (GCal operations)
- **Span names:** `scheduling.slots.compute`, `scheduling.booking.confirm`, `scheduling.booking.cancel`, `scheduling.booking.reschedule`, `gcal.freebusy.query`, `gcal.events.insert`, `gcal.events.delete`, `scheduling.notification.send`
- **Custom attributes:** `BUSINESS_ID`, `SCHEDULING_APPOINTMENT_ID`, `SCHEDULING_STATUS`, `SCHEDULING_SLOTS_AVAILABLE`, `SCHEDULING_BUFFER_MINUTES`, `GCAL_CALENDAR_ID`, `GCAL_EVENT_ID`, etc. (from `autom8y_telemetry.conventions`)
- **Side effects:** `record_side_effect(span, system, operation, target, payload)` from `autom8y_telemetry.effects` marks all DB and external writes in spans

### Prometheus (Metrics)

- **SDK:** `prometheus-client`
- **Metrics defined in `api/routes/__init__.py`:**
  - `scheduling_booking_outcomes_total` — Counter by outcome
  - `scheduling_availability_duration_seconds` — Histogram by status
  - `scheduling_booking_duration_seconds` — Histogram by outcome
  - `scheduling_cancel_outcomes_total` — Counter
  - `scheduling_cancel_duration_seconds` — Histogram
  - `scheduling_reschedule_outcomes_total` — Counter
  - `scheduling_reschedule_duration_seconds` — Histogram
  - `scheduling_gcal_overlay_duration_seconds` — Histogram by status + cache_hit
- **Metrics defined in `scheduling/engine.py`:**
  - `scheduling_availability_scan_size` — Histogram (appointment scan count)
- **Metrics defined in `scheduling/booking.py`:**
  - `scheduling_booking_scan_size` — Histogram by operation (idempotency, overlap_precheck, overlap_locked)

### GitHub Actions (CI)

- **Workflow:** `test.yml` delegates to shared reusable workflow `autom8y/autom8y-workflows/.github/workflows/satellite-ci-reusable.yml@main`
- **CI scope:** mypy (src/autom8_scheduling), pytest coverage threshold 20%, OTel convention check (`sprint2_instrumentation` filter), integration tests on push to main
- **Additional workflows:** `dependency-review.yml`, `gitleaks.yml`, `trufflehog-scan.yml` (secret scanning), `zizmor.yml` (Actions security), `satellite-dispatch.yml`

## Knowledge Gaps

1. **`src/autom8_scheduling/api/__init__.py`** — not read; expected to be empty or a simple re-export. Low impact.
2. **`src/autom8_scheduling/services/__init__.py`** — not read; expected to be empty. Low impact.
3. **`autom8y-gcal` SDK internals** — `GCalClient`, `GCalConfig`, `EventCreate`, `EventDateTime` are used but the SDK source is not in scope. The integration contract is documented from call sites.
4. **`autom8y_config.Autom8yBaseSettings`** — the `is_local` property used in `SchedulingSettings.cors_origins_with_local` is inherited from the SDK; exact behavior undocumented here.
5. **All CI workflow details beyond `test.yml`** — the `satellite-dispatch.yml`, `zizmor.yml`, `gitleaks.yml`, and `trufflehog-scan.yml` workflows were not read. Their content is ancillary to the core architecture.
6. **`tests/test_proxy.py` and `tests/golden_traces/`** — test files were not read individually. Their architectural significance is limited to validating the integration layer.
7. **`renovate.json`** — dependency automation config, not read. Not architecturally significant.
8. **`devbox.json`** — local dev environment toolchain, not read.
9. **APScheduler (Phase 3)** — commented placeholder in `app.py` lifespan for GCal reconciliation cron. Not implemented.
