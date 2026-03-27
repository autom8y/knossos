---
domain: design-constraints
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "39376b6"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Structural Tensions

### Tension 1: Migrated Code with Relaxed Type Enforcement

The entire scheduling domain was migrated from `autom8y-data` under task `TDD-SCHED-EXTRACT-001`. This migration introduced a structural tension: the code was written against `autom8y-data`'s strict SQLAlchemy typing patterns, but the scheduling service cannot import `autom8y-data` at runtime. The resolution was to suppress mypy errors across almost all application modules rather than refactor types.

**Evidence**: `pyproject.toml` lines 125-135 — `ignore_errors = true` for `autom8_scheduling.api.routes`, `autom8_scheduling.services.*`, `autom8_scheduling.scheduling.*`, `autom8_scheduling.models.*`. This means `mypy --strict` is not enforced on any production code path. Type safety exists only in `config.py`, `app.py`, and `health.py`.

### Tension 2: ApiField Stub That Silently Discards Metadata

`ApiField` in `autom8y-data` is a rich metadata wrapper that drives schema derivation. In the scheduling service, it is stubbed as a no-op wrapper around `sqlmodel.Field` that accepts and drops `roles`, `validator`, and `api_alias`.

**Evidence**: `src/autom8_scheduling/models/_base.py` lines 70-83. The `validator` parameter is accepted but never called. Field-level validators annotated on `Appointment.office_phone`, `Appointment.phone`, etc., are silently ignored at the ORM layer. Service-layer code (`AppointmentService`) manually calls `_validate_phone_e164` to compensate, but `BookingEngine` does not.

### Tension 3: Dual Write-Path Architecture (BookingEngine vs AppointmentService)

The codebase has two overlapping paths for appointment writes: `BookingEngine` (booking.py) handles atomic check-and-book, and `AppointmentService` (appointment.py) handles CRUD operations. The `write_ops.py` module (`ADR-write-path-unification`) was added to standardize behavior across both, but the unification is incomplete: `dispatch_scheduling_notifications` in `write_ops.py` is a no-op for `source="booking_engine"`.

**Evidence**: `src/autom8_scheduling/scheduling/write_ops.py` lines 208-211.

### Tension 4: Idempotency Key Lookup Scans Full Table

Idempotency is implemented by storing keys in `appointments.config` (a JSON column) and doing a full scan filtered only by `source="sms-ai"` and `config IS NOT NULL`. There is no index on the JSON path, and no database-level uniqueness constraint.

**Evidence**: `src/autom8_scheduling/scheduling/booking.py` lines 685-711. Tolerable at current scale but will degrade with appointment volume.

### Tension 5: In-Memory FreeBusy Cache Is Process-Local

The GCal overlay cache (`_FreeBusyCache`) is a module-level singleton. In a multi-instance deployment, each container has its own cache. Cache invalidation cannot be coordinated across instances.

**Evidence**: `src/autom8_scheduling/scheduling/gcal_overlay.py` lines 104-110. TTL is 300 seconds (`CACHE_TTL_EVENTUAL`).

## Frozen Areas

### Frozen Area 1: `appointments` Table Schema

The table schema for `appointments` mirrors a production MySQL database (`nhc_production`) shared with the legacy `autom8y-data` codebase. Columns cannot be renamed, retyped, or dropped without coordinating with multiple consumers.

**Evidence**: `src/autom8_scheduling/models/scheduling.py` lines 10-86. `appointment_id` maps to `sa_column_kwargs={"name": "id"}`. `appt_type` maps to `sa_column_kwargs={"name": "type"}`. The `guid` column is documented as "Dead UUID column (100% NULL, not a real FK)" — it cannot be dropped without DB migration.

### Frozen Area 2: `chiropractors` Table (Business)

The `Business` model maps to `chiropractors` and is explicitly annotated READ-ONLY except for four scheduling configuration fields. The table is owned by `autom8y-data`.

**Evidence**: `src/autom8_scheduling/models/shared.py` lines 1-8. Scheduling-owned writable columns: `scheduling_enabled`, `buffer_minutes`, `gcal_enabled`, `gcal_shadow_mode`.

### Frozen Area 3: `leads` Table

The `Lead` model is explicitly READ-ONLY with no allowed writes from the scheduling service.

### Frozen Area 4: `office_phone` as Primary Business Key

Throughout the codebase, `office_phone` in E.164 format is the implicit "natural key" used to join across all scheduling tables. It is marked `FieldRole.IMMUTABLE` and `FieldRole.FK` on every model that carries it. Changing a business's phone number would require rewriting all related rows across all tables.

### Frozen Area 5: `appointments.config` JSON as State Bag

The `config` JSON column is used by multiple subsystems. Keys stored in it (`idempotency_key`, `cancel_idempotency_key`, `reschedule_idempotency_key`, `reminder_sent_at`, `confirmation_sms_sent_at`, `cancellation_sms_sent_at`, etc.) are the only cross-subsystem coordination mechanism. These key names are frozen.

### Frozen Area 6: API Route Prefix `/api/v1/scheduling`

The route prefix is wired in `_mount_routers` and is the public API contract consumed by the main `autom8y` monorepo.

### Frozen Area 7: Health Endpoint Paths

The JWT auth middleware explicitly excludes `/health`, `/ready`, `/health/deps`, `/docs`, `/openapi.json` from auth. These paths are frozen.

### Frozen Area 8: `SLOT_BLOCKING_STATUSES` and `CANCELLABLE_STATUSES`

These `frozenset` constants were derived from a production audit of 149,226 appointments with 16 distinct status values. Changing them changes which slots appear available — a user-visible behavior change.

**Evidence**: `src/autom8_scheduling/scheduling/constants.py` lines 1-26.

## Version and Compatibility Constraints

### Python Version

- **Minimum Python**: 3.12 (enforced in `pyproject.toml` `requires-python = ">=3.12"`, `[tool.ruff] target-version = "py312"`, `[tool.mypy] python_version = "3.12"`)
- **Docker base image**: `python:3.12-slim`

### Core Framework Versions

| Dependency | Constraint | Reason |
|---|---|---|
| `fastapi>=0.109.0` | Lower bound | Security and lifespan API compatibility |
| `uvicorn[standard]>=0.27.0` | Lower bound | Async worker stability |
| `pydantic>=2.0.0` | Lower bound | Pydantic v2 API |
| `pydantic-settings>=2.0.0` | Lower bound | Pydantic v2 settings API |
| `sqlalchemy>=2.0.23` | Lower bound | Async 2.0 API |
| `sqlmodel>=0.0.14` | Lower bound | Pydantic v2 compatibility |
| `aiomysql>=0.2.0` | Lower bound | Async MySQL driver |
| `prometheus-client>=0.20.0` | Lower bound | Histogram label API |

### autom8y Internal SDK Constraints

All internal packages sourced from AWS CodeArtifact:

| Package | Minimum Version |
|---|---|
| `autom8y-config>=1.2.1` | Provides `Autom8yBaseSettings`, `Autom8yEnvironment` |
| `autom8y-log>=0.5.6` | Provides `get_logger` |
| `autom8y-core>=1.1.1` | Provides `TimezoneNotConfiguredError` |
| `autom8y-auth[observability]>=1.1.0` | JWT middleware |
| `autom8y-telemetry[fastapi,otlp]>=0.6.0` | OTel instrumentation |
| `autom8y-http>=0.3.0` | HTTP client for Twilio/SendGrid |
| `autom8y-gcal>=0.1.0` | Optional GCal SDK |

### Other Constraints

- **uv lock and DEF-009/SCAR-022**: `--frozen` is mutually exclusive with `--no-sources` in `uv>=0.15.4`. Dockerfile uses `--no-sources` only.
- **pytest-asyncio**: `>=1.2,<2.0` upper-bound cap.
- **Coverage threshold**: 20% (intentionally low for migration state).
- **Env variable prefix**: `SCHEDULING_` per `ADR-ENV-NAMING-CONVENTION Decision 11`.

## Invariants and Contracts

### Invariant 1: GCal Failure Never Rolls Back a MySQL Booking

**Evidence**: `src/autom8_scheduling/scheduling/gcal_sync.py` lines 7-8:
```python
# INVARIANT: GCal failure NEVER rolls back a MySQL booking.
# INVARIANT: appointments.event_id is the ONLY binding column.
```

### Invariant 2: GCal Overlay Is Subtractive Only

**Evidence**: `src/autom8_scheduling/scheduling/gcal_overlay.py` lines 3-4 and line 119: `INVARIANT: Can only remove slots.`

### Invariant 3: Status State Machine (VALID_TRANSITIONS)

Terminal statuses (`completed`, `cancelled`, `no_show`) have no outbound transitions. `validate_status_transition(current, new)` raises `SchedulingValidationError` on invalid transitions.

### Invariant 4: Date Range Lookforward Hard Limit

Availability queries cannot span more than 14 days (`_MAX_LOOKFORWARD_DAYS = 14`).

### Invariant 5: E.164 Phone Format

Phone numbers must match `^\+[1-9]\d{6,14}$`. Enforced explicitly in `AppointmentService._validate_phone_e164`.

### Invariant 6: Appointments Are Soft-Deleted

Cancellations set `status="cancelled"` and write metadata to `config`. Physical deletes only via `AppointmentService.delete()` (admin-only).

### Invariant 7: Atomic Check-and-Book via SELECT FOR UPDATE

The final conflict check before inserting uses `SELECT FOR UPDATE` to prevent TOCTOU races. Referenced as `ADR-002`.

### Invariant 8: List Pagination Hard Cap

`AppointmentService.list()` enforces `limit = min(max(1, limit), 100)`.

### Invariant 9: Idempotency via JSON Config Keys

All three booking operations store idempotency keys in `appointments.config`. On repeat calls, the engine returns the stored result without re-executing.

### Invariant 10: Settings Singleton via lru_cache

`get_settings()` is decorated with `@lru_cache` — settings are loaded once per process and cached indefinitely.

### Invariant 11: Cancellation Ownership Verification

`BookingEngine.cancel()` verifies both `office_phone` and `lead_phone` match before cancelling.

### Invariant 12: Request IDs

Every HTTP request receives an `x-request-id` header value (or new UUID4). Included in all 500 error responses.

### Invariant 13: Immutable Appointment Fields

`appointment_id`, `guid`, `office_phone`, `phone`, `created`, `chiropractor_guid` cannot be updated through `AppointmentService.update()`.

## Risk Zones

### Risk Zone 1: Booking Conflict Detection Under Concurrent Load

The booking flow uses a two-phase conflict check with full in-memory appointment scans. Neither check uses SQL range predicates to narrow the scan.

**Evidence**: `src/autom8_scheduling/scheduling/booking.py` lines 719-770.

### Risk Zone 2: Datetime Parsing of VARCHAR Timestamps

Appointment datetimes stored as `VARCHAR`. Malformed datetimes in existing production records are silently skipped during availability computation, potentially making slots appear available that should be blocked.

### Risk Zone 3: GCal Sync Using Separate Session Commits

`_store_event_id` and `_clear_event_id` call `session.commit()` independently from the booking transaction. If the GCal API call succeeds but the `event_id` commit fails, there is no recovery path.

### Risk Zone 4: Notification Deduplication via JSON Config

SMS and email notifications are deduplicated by checking timestamps in `appointments.config`. This deduplication is not atomic: concurrent notification attempts could both pass the check before either writes.

### Risk Zone 5: Async Session Creation Per Request

`get_session` creates a new `async_sessionmaker` factory on every request rather than reusing one created at startup.

### Risk Zone 6: GCal Shadow Mode Default

`gcal_shadow_mode` defaults to `True`. If accidentally disabled in production without validation, GCal's busy periods would immediately remove slots.

### Risk Zone 7: pool_recycle=3600 with MySQL

Connection killed between ping and use in a single request can still cause a transient error despite `pool_pre_ping=True`.

## External Constraints

### MySQL Schema Ownership (nhc_production)

Tables `chiropractors`, `leads` are owned by the legacy `autom8y-data` system. Schema changes must be coordinated.

### Google Calendar API Rate Limits

5-minute in-memory cache is the primary rate limit defense. No retry-with-backoff on FreeBusy calls.

### Twilio API

Hardcoded API version `/2010-04-01/`. Timeout fixed at 10 seconds.

### SendGrid API

Hardcoded endpoint `https://api.sendgrid.com/v3/mail/send`. 2 retry attempts for 5xx.

### AWS CodeArtifact for Internal Packages

Docker build requires `EXTRA_INDEX_URL` as a build argument. Cannot build without AWS credentials.

### autom8y-auth JWT Contract

JWT validation logic, token expiry, and claim structure owned by `autom8y-auth` SDK.

### autom8y-telemetry OTel Span Attribute Names

Span attribute constants imported at top level from `autom8y_telemetry.conventions`. If the SDK renames a constant, the scheduling service will fail to compile.

### Satellite Deploy Protocol

Post-test CI dispatches `satellite-deploy` event to `autom8y/autom8y`. Renaming the service or repository requires updating the receiving monorepo.

### IANA Timezone in addresses Table

`addresses.timezone` must contain a valid IANA timezone string. The scheduling service cannot override or default this value.

## Knowledge Gaps

1. **`autom8y-gcal` SDK interface**: The exact API surface (`GCalConfig` constructor args, `free_busy.query` response schema) is not documented in this repository. There is a `type: ignore[call-arg]` on the `GCalConfig` constructor call.
2. **Migration history**: No migration files in the observed scope. Models reflect a snapshot of the production schema.
3. **Pool sizing rationale**: `pool_size=3`, `max_overflow=5` referenced as `ADR-POOL-001 defaults`. The ADR itself is not in this repository.
4. **`autom8y-config Autom8yBaseSettings` behavior**: `is_local`, `autom8y_env`, and secret resolution are inherited from the external package.
