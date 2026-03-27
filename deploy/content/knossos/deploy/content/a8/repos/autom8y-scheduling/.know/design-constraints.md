---
domain: design-constraints
generated_at: "2026-03-27T19:56:20Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4557333"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

---

## Tension Catalog

### TENSION-001: `chiropractors` Table Name vs `Business` Model Name

**Type**: Naming mismatch
**Location**: `src/autom8_scheduling/models/shared.py:13-76`
**Historical reason**: The production database uses `chiropractors` as the table name (legacy vertical-specific naming from before the platform generalized). The Python model was renamed to `Business` to match the platform's domain language.
**Current state**: `class Business(SQLModel): __tablename__ = "chiropractors"`. Every ORM reference to `Business` queries `chiropractors` transparently. Foreign keys in all other models (`Address`, `Hours`, `BusinessOffer`, `Appointment`) reference `chiropractors.office_phone` and `chiropractors.guid` explicitly.
**Ideal resolution**: Database table renamed to `businesses`. Requires a migration in autom8y-data and coordination across all services reading this table.
**Resolution cost**: High (cross-service, requires autom8y-data owner coordination).
**Changeability**: Migration (breaking rename across services).

---

### TENSION-002: `appointments.id` Stored as `appointment_id` in ORM

**Type**: Naming mismatch
**Location**: `src/autom8_scheduling/models/scheduling.py:22-28`
**Historical reason**: The database column is `id` but the domain concept and all business logic references use `appointment_id`. The ORM alias (`sa_column_kwargs={"name": "id"}`) bridges the gap.
**Current state**: `appointment_id: int | None = ApiField(sa_column_kwargs={"name": "id"})`. Queries filter on `Appointment.appointment_id == appointment_id` which translates to `id = ?`.
**Ideal resolution**: Database column renamed to `appointment_id`.
**Resolution cost**: Medium (migration + multi-service coordination for the appointments table).
**Changeability**: Migration.

---

### TENSION-003: Dead `guid` Column on `Appointment`

**Type**: Schema evolution debt
**Location**: `src/autom8_scheduling/models/scheduling.py:31-35`
**Historical reason**: `guid` was intended as a UUID foreign key for the `chiropractors` table. 100% of production rows have NULL for this column. The real FK is `chiropractor_guid` → `chiropractors.guid`.
**Current state**: Both `guid` (100% NULL, dead) and `chiropractor_guid` (real FK) exist. The comment `"Dead UUID column (100% NULL, not a real FK)"` marks it.
**Ideal resolution**: Remove `guid` column from appointments table. Update ORM model.
**Resolution cost**: Medium (migration, must not break existing select queries).
**Changeability**: Migration.

---

### TENSION-004: `street_num_suffiX` Column Name Typo in Database

**Type**: Naming mismatch (external)
**Location**: `src/autom8_scheduling/models/scheduling.py:135-139`
**Historical reason**: A typo in the original schema migration created `street_num_suffiX` (capital X) in the `addresses` table. This is now load-bearing: changing the name requires a data migration.
**Current state**: `sa_column_kwargs={"name": "street_num_suffiX"}` with a comment documenting the typo.
**Ideal resolution**: Column renamed in DB with a migration.
**Resolution cost**: Low (isolated rename, no business logic depends on this field's name).
**Changeability**: Migration. The field is ORM-only and not used in scheduling logic.

---

### TENSION-005: JSON Idempotency Keys in `config` BLOB vs Dedicated Column

**Type**: Under-engineering / schema evolution debt
**Location**: `src/autom8_scheduling/scheduling/booking_helpers.py:39-80` (consolidated `find_by_config_key()`)
**Historical reason**: Idempotency keys (`idempotency_key`, `cancel_idempotency_key`, `reschedule_idempotency_key`) are stored in the `config` JSON column rather than dedicated indexed columns. This was migrated from the monorepo where the JSON approach was expedient.
**Current state**: Idempotency checks scan all `source=="sms-ai"` appointments filtered by `office_phone`, deserializing each `config` blob to find the matching key. GAP-001 was resolved by consolidating three duplicate methods into a single `find_by_config_key()` in `booking_helpers.py`. A Prometheus histogram (`scheduling_booking_scan_size`) monitors scan sizes.
**Performance implication**: At low appointment volumes this is acceptable. At high volumes (1000+ appointments per business), the scan cost grows linearly.
**Ideal resolution**: Add `idempotency_key`, `cancel_idempotency_key`, `reschedule_idempotency_key` as indexed columns on `appointments`.
**Resolution cost**: Medium (3 migrations + ORM update + booking.py rewrite).
**Changeability**: Coordinated (multi-file change to booking_helpers.py + migrations). ADR required before changing.
**Related**: SCAR-002 (config JSON defensive parsing), RISK-001.

---

### TENSION-006: mypy Strict Mode Island vs Migrated Code Relaxation

**Type**: Dual-system / type system debt
**Location**: `pyproject.toml` mypy configuration (lines 127-148)
**Historical reason**: Scheduling domain code was migrated from `autom8y-data` (a monorepo) with raw SQLAlchemy 2.x patterns that predate strict typing. Applying strict mypy to the migrated code produced hundreds of errors for patterns that are functionally correct.
**Current state**: Two tiers of mypy strictness. Strict island: `scheduling.exceptions`, `scheduling.results`, `scheduling.constants`, `scheduling.write_ops`, `scheduling.offer_resolution`. Relaxed (`ignore_errors = true`): all `api.routes.*`, `scheduling.*` (except strict island), `services.*`, `models.*`.
**Note**: `offer_resolution.py` was added to the strict island in a B-grade remediation sprint (SCAR-008 resolution). The island now covers 5 modules.
**Ideal resolution**: Incrementally expand the strict island as modules are modernized.
**Resolution cost**: Ongoing (per-module effort, low per-module cost).
**Changeability**: Safe (add new modules to strict island one at a time). Never regress a module from strict to relaxed.

---

### TENSION-007: `Business.appt_duration` as Business-Level Fallback in 3-Tier Duration Chain

**Type**: Split responsibility / naming tension
**Location**: `src/autom8_scheduling/scheduling/engine.py` (`_resolve_appt_duration`), `src/autom8_scheduling/scheduling/offer_resolution.py` (`SchedulingConfig.duration_minutes`)
**Historical reason**: `appt_duration` is a real column on `chiropractors` (not a phantom). Per ADR-scheduling-per-offer-wiring, `SchedulingConfig` now carries `duration_minutes` (per-offer) but `Business.appt_duration` remains as the second tier for the 51% of businesses with no per-offer duration set.
**Current state**: 3-tier duration resolution: (1) `SchedulingConfig.duration_minutes` from `business_offers.duration`, (2) `Business.appt_duration` from `chiropractors.appt_duration`, (3) `_DEFAULT_APPT_DURATION = 30` minutes. The engine's `_resolve_appt_duration()` still queries `chiropractors.appt_duration` as tier 2.
**Ideal resolution**: Over time, populate per-offer duration for all offers, making tier 2 obsolete. Not a code change — an operational data population task.
**Changeability**: Safe (current 3-tier chain is backward-compatible; tier 2 just becomes unused over time).

---

### TENSION-008: `Business.offer_id` Used as Fallback in Offer Resolution Chain

**Type**: Layering / cross-table dependency
**Location**: `src/autom8_scheduling/scheduling/offer_resolution.py:49-59`
**Historical reason**: The fallback resolution path uses `chiropractors.offer_id` to find a default business_offer when no non-disabled offer exists. This creates a dependency path: `Business` (shared/read-only) is joined in the fallback query.
**Current state**: Fallback query JOINs `business_offers` on `chiropractors.office_phone` filtered by `chiropractors.offer_id`. Necessary for 750 businesses with all-disabled offers.
**Ideal resolution**: No tension — this is an intentional design decision from ADR-scheduling-offer-resolution.
**Changeability**: Frozen (changing offer resolution requires a new ADR).

---

### TENSION-009: No DB-Level Constraint Enforcing scheduling_enabled Semantics

**Type**: Missing abstraction / risk zone
**Location**: `src/autom8_scheduling/scheduling/offer_resolution.py:98`
**Historical reason**: The inverted semantics (`disabled=False` → scheduling enabled) are a historical database convention from the monorepo era. aiomysql returns integer 0 for MySQL TINYINT columns, making `offer.disabled == False` (equality) load-bearing (not `is False`).
**Current state**: All enforcement is in `SchedulingConfig.from_business_offer()`. If code bypasses `SchedulingConfig` and reads `disabled` directly, the semantics can be silently inverted.
**Ideal resolution**: Convention documentation + test coverage (in place). See LBC-002.
**Changeability**: Safe (only touches offer_resolution.py).

---

### TENSION-010: Dual Calendar Systems (GCal + GHL) with Separate Binding Columns

**Type**: Dual-system pattern
**Location**: `src/autom8_scheduling/scheduling/gcal_sync.py`, `src/autom8_scheduling/scheduling/ghl_sync.py`, `src/autom8_scheduling/models/scheduling.py:84-85`
**Historical reason**: The service originally wrote only to GCal. ADR-scheduling-calendar-integration added GHL calendar write (for 84.2% of offers that have both `ghl_calendar_id` and `master_cal_id`). Each calendar system requires its own binding column because a single appointment can have events in both simultaneously.
**Current state**: `appointments.event_id` binds GCal events. `appointments.ghl_event_id` binds GHL events. Two structurally symmetric modules: `gcal_sync.py` (457 lines, production) and `ghl_sync.py` (430 lines, scaffolded with `GHL_AVAILABLE = False` pending client library). Both use fire-and-forget via `asyncio.gather(return_exceptions=True)`.
**Key constraint**: GHL sync currently always skips gracefully because `GHL_AVAILABLE = False` — the GHL client library does not exist yet. When the library ships, this is the activation point.
**Ideal resolution**: When the GHL client library ships, flip the conditional import in `ghl_sync.py` to mirror `gcal_sync.py`.
**Changeability**: Coordinated (requires GHL client library + GHL reconcile endpoint + schema migration for `ghl_event_id`). ADR-scheduling-calendar-integration defines the activation plan.

---

### TENSION-011: GCal Calendar ID Resolution (Per-Business) vs GHL Calendar ID Resolution (Per-Offer)

**Type**: Asymmetric resolution logic
**Location**: `src/autom8_scheduling/scheduling/gcal_sync.py` (`_resolve_calendar_id`), `src/autom8_scheduling/scheduling/ghl_sync.py` (uses `config.ghl_calendar_id`)
**Historical reason**: GCal resolves calendar ID via `office_phone` → first available `BusinessOffer.master_cal_id` (any offer for the business). GHL must resolve per-offer because 234 business-calendar combinations share GHL calendars across offers. Per ADR-scheduling-calendar-integration, `ghl_calendar_id` comes directly from the resolved `SchedulingConfig`.
**Current state**: The two sync modules have different calendar ID resolution paths. `gcal_sync.py` runs a DB query (`_resolve_calendar_id`). `ghl_sync.py` uses `config.ghl_calendar_id` already in memory. GHL is cheaper at dispatch time (no extra query).
**Ideal resolution**: No change needed. The asymmetry is intentional and documented.
**Changeability**: Coordinated (resolving calendar IDs from SchedulingConfig for GCal too would require refactoring gcal_sync.py dispatch signatures).

---

### TENSION-012: `dispatch_scheduling_notifications()` in `write_ops.py` is a No-Op for BookingEngine in v1

**Type**: Premature abstraction / stub
**Location**: `src/autom8_scheduling/scheduling/write_ops.py:149-223`
**Historical reason**: The write-path unification ADR defined `dispatch_scheduling_notifications()` as a shared function for both `BookingEngine` and `AppointmentService`. In v1, `BookingEngine` route handlers dispatch notifications directly; the function short-circuits (`if source == "booking_engine": return`).
**Current state**: The function is a partial implementation. `AppointmentService` status changes call it for `cancelled`/`scheduled` transitions. `BookingEngine` dispatches notifications via its own route-level code path. The function's docstring acknowledges this: "BookingEngine continues to use its existing notification dispatch in route handlers for now."
**Ideal resolution**: Phase 3 unification where BookingEngine also routes through this function. Requires refactoring the route-level asyncio.gather block.
**Changeability**: Coordinated (requires touching booking.py route handlers and write_ops.py together).

---

### TENSION-013: `gcal_sync.py` Accepts `booking_result: dict[str, Any]` (Pre-typed Interface)

**Type**: Layering violation / interface debt
**Location**: `src/autom8_scheduling/scheduling/gcal_sync.py:46-56`, `src/autom8_scheduling/scheduling/ghl_sync.py:43-56`, `src/autom8_scheduling/scheduling/notifications/dispatch.py:113-116`
**Historical reason**: `gcal_sync.py` was migrated from autom8y-data before the typed result classes (`BookingResult`, `CancelResult`, etc.) existed. Its dispatch functions accept `booking_result: dict[str, Any]` rather than the typed result objects defined in `results.py`.
**Current state**: 3 dispatch functions in `gcal_sync.py` and 3 in `ghl_sync.py` use dict key access (`booking_result["appointment_id"]`, `booking_result["start_datetime"]`). The typed `BookingResult` dataclasses exist but the sync modules don't use them. `notifications/dispatch.py` also uses dict key access at lines 113-116.
**Risk**: Dict key access is not type-checked (these files are in the `ignore_errors = true` zone). A mismatched key produces a runtime `KeyError`.
**Ideal resolution**: Update dispatch signatures to accept typed result objects from `results.py`. Gate on mypy strict island expansion.
**Changeability**: Coordinated (changing dispatch signatures requires touching all 3 call sites in booking.py routes, plus the sync modules).

---

## Trade-off Documentation

### Trade-off 1: Fire-and-Forget Calendar Sync vs Transactional Consistency

**ADR**: ADR-scheduling-calendar-integration
**Current state**: MySQL booking is committed before any calendar dispatch fires. Calendar sync runs in `asyncio.gather(return_exceptions=True)` post-commit. GCal and GHL failures are logged but never propagate to the caller.
**Why current persists**: GCal fire-and-forget has a 99.8-100% success rate in production. The invariant "GCal/GHL failure NEVER rolls back a MySQL booking" is the key design decision — patient bookings are guaranteed at the cost of occasional calendar drift. Calendar drift is recoverable; booking rollbacks are not.
**Trade-off**: Occasional calendar drift (silent failure mode) in exchange for booking durability. The alternative (transactional calendar sync) would require distributed transactions or two-phase commit — not viable for external APIs.
**Tensions linked**: TENSION-010.

---

### Trade-off 2: Allow-List → Deny-List Status Inversion

**ADR**: ADR-scheduling-status-taxonomy
**Current state**: `SLOT_BLOCKING_STATUSES` (allow-list, 6 entries) is deprecated. `TERMINAL_STATUSES` (deny-list, 7 entries) is authoritative. Any status not in the terminal set blocks a slot by default.
**Why current persists**: Production audit found 3 statuses (`none`, `called`, `contacted`) invisible to the allow-list, creating 947 double-booking-vulnerable slots across 38 active businesses. The deny-list eliminates the class of failure.
**Trade-off**: Possible false positives (over-blocking) for unknown statuses in exchange for zero false negatives (double-bookings). Triple-write pattern adds up to 3x appointment row multiplier for 79 businesses. `SLOT_BLOCKING_STATUSES` retained as deprecated constant for backward compatibility.
**Tensions linked**: None.

---

### Trade-off 3: SchedulingConfig as Single Config Interface vs Multi-Source Reality

**ADR**: ADR-scheduling-offer-resolution, ADR-scheduling-per-offer-wiring
**Current state**: `SchedulingConfig` (frozen dataclass) centralizes all scheduling dimensions: `scheduling_enabled`, `buffer_minutes`, `gcal_enabled`, `gcal_shadow_mode`, `head_employee_id`, `master_cal_id`, `offer_id`, `business_offer_guid`, `duration_minutes`, `ghl_enabled`, `ghl_calendar_id`. All populated from `business_offers` at request time. `Business.appt_duration` remains as fallback tier 2 for duration (not on `SchedulingConfig`).
**Why current persists**: The prior phantom-column design (`chiropractors.scheduling_enabled`) read config that didn't exist, returning Python defaults silently. The offer resolution redesign fixes this by reading from `business_offers` where the data actually lives.
**External constraint**: The 4 phantom columns (`scheduling_enabled`, `buffer_minutes`, `gcal_enabled`, `gcal_shadow_mode`) do not exist on `chiropractors`. They exist on `business_offers` via migrations 009/010. No code change can make phantom columns exist without a migration.
**Tensions linked**: TENSION-007.

---

### Trade-off 4: mypy Strict Island with Broad Suppression Default

**ADR**: No ADR — migration decision. Documented in `pyproject.toml` comment and `.know/scar-tissue.md` SCAR-008.
**Current state**: `ignore_errors = true` for `api.routes.*`, `scheduling.*`, `services.*`, `models.*`. 5 modules in the strict island override this for new/refactored code.
**Why current persists**: Migrated code from autom8y-data uses SQLAlchemy 2.x raw patterns that generate hundreds of false mypy errors. The strict island is an incremental remediation strategy.
**Constraint**: Any new module added to `autom8_scheduling.scheduling.*` inherits `ignore_errors = true` unless explicitly added to the strict island. Convention (SCAR-008) mandates adding new non-migrated modules to the island.
**Tensions linked**: TENSION-006.

---

### Trade-off 5: Envelope Pattern vs Health Endpoint Exclusion

**ADR**: ADR-scheduling-envelope-pattern
**Current state**: All agent-facing endpoints use `SchedulingResponse[T]` / `SchedulingErrorResponse` envelopes. Health endpoints (`/health`, `/ready`, `/health/deps`) retain `HealthResponse` Pydantic model — no scheduling envelope.
**Why current persists**: Health endpoints serve infrastructure (load balancers, Kubernetes probes), not AI agents. The `HealthResponse` model follows a platform-wide health contract shared with `autom8y-auth` and other services. Wrapping health in the scheduling envelope would break that contract.
**Tension**: The codebase has two response shape systems coexisting in the same service.

---

## Abstraction Gap Mapping

### GAP-001: Idempotency Key Scan (RESOLVED)

**Prior state**: Three identical `_find_by_*_idempotency_key` methods in `booking.py` sharing the same structure (25 lines each, 3 locations = 75 total lines of duplication).
**Current state**: Consolidated into single `find_by_config_key(session, key_name, key_value, ...)` in `src/autom8_scheduling/scheduling/booking_helpers.py:39-80`. Duplication eliminated.
**Status**: Resolved. No maintenance burden.

---

### GAP-002: `check_scheduling_readiness()` Partially Duplicates Enable-Gate Queries (RESOLVED)

**Prior state**: `check_scheduling_readiness()` in `validation.py` and the PATCH `/config` handler in `businesses.py` both computed timezone and hours prerequisites independently.
**Current state**: `compute_enable_prerequisites()` extracted to `src/autom8_scheduling/scheduling/validation.py:23-48`. Both `check_scheduling_readiness()` and the PATCH handler call the shared function.
**Status**: Resolved. No maintenance burden.

---

### GAP-003: `gcal_sync.py` and `ghl_sync.py` Accept `dict[str, Any]` Instead of Typed Results

**Location**: `src/autom8_scheduling/scheduling/gcal_sync.py`, `src/autom8_scheduling/scheduling/ghl_sync.py`, `src/autom8_scheduling/scheduling/notifications/dispatch.py:113-116`
**Current state**: Dispatch functions accept `booking_result: dict[str, Any]` and use key access (`booking_result["appointment_id"]`). Typed result classes (`BookingResult`, `CancelResult`, `RescheduleResult`) exist in `src/autom8_scheduling/scheduling/results.py` but are not used by the sync modules.
**Maintenance burden**: Dict key mismatches produce runtime `KeyError`. Not caught by mypy (files are in `ignore_errors = true` zone). Any change to result field names must be manually tracked to 3 call sites in gcal_sync.py, 3 in ghl_sync.py, and 3 in notifications/dispatch.py.
**Recommended abstraction**: Update dispatch signatures to accept typed result objects. Expansion of mypy strict island for these modules should accompany the change.
**Related**: TENSION-013.

---

### GAP-004: `dispatch_scheduling_notifications()` is an Incomplete Abstraction

**Location**: `src/autom8_scheduling/scheduling/write_ops.py:149-223`
**Current state**: The function is called by `AppointmentService` but is a no-op for `BookingEngine` (which dispatches notifications directly in route handlers). The function's design assumes it will eventually unify both paths, but v1 only unifies the AppointmentService path.
**Maintenance burden**: Two parallel notification dispatch patterns exist: (1) `asyncio.gather()` in booking route handlers, (2) `dispatch_scheduling_notifications()` in write_ops.py. When notification behavior needs to change, it may need to change in both locations.
**Recommended abstraction**: Phase 3 unification — route handlers delegate to `dispatch_scheduling_notifications()` with `source="booking_engine"`.
**Related**: TENSION-012.

---

### GAP-005: Zombie `ApiField` and `FieldRole` Stubs

**Location**: `src/autom8_scheduling/models/_base.py:58-83`
**Current state**: `ApiField` discards `roles`, `validator`, and `api_alias` parameters silently. `FieldRole` constants exist as documentation markers only. Neither is used by any active code path — they exist to prevent import errors in ORM model definitions that use these parameters.
**Maintenance burden**: Models carry vestigial `roles={FieldRole.IMMUTABLE}`, `validator=validate_e164_phone` parameters that do nothing at runtime. E.164 validation is re-implemented in `services/appointment.py:41-47` independently (duplicated validator logic).
**Recommended abstraction**: Either wire the validators (making `ApiField` actually call `validator` functions) or remove the parameters from model definitions and delete `FieldRole`. LBC-003 documents why naive wiring is dangerous.
**Related**: LBC-003.

---

## Load-Bearing Code

### LBC-001: Two-Phase Overlap Check Sequence in `BookingEngine`

**Location**: `src/autom8_scheduling/scheduling/booking_helpers.py` (`check_overlap_nonlocking`, `select_overlapping`), called from `src/autom8_scheduling/scheduling/booking.py`
**Dependents**: `BookingEngine.book()`, `BookingEngine.reschedule()`.
**What it does**: `check_overlap_nonlocking()` performs a fast non-locking conflict scan. `select_overlapping()` performs a locking query with SELECT FOR UPDATE. Both must be called in this order.
**Naive fix failure mode**: Removing `select_overlapping` eliminates the atomicity guarantee — two concurrent bookings can both pass the precheck and both insert conflicting appointments under load. This is the TOCTOU race condition this pattern defends against (SCAR-001).
**Safe refactor**: Any refactor must preserve (1) non-locking phase first, (2) locking phase with SELECT FOR UPDATE, (3) both phases filter by `TERMINAL_STATUSES` deny-list logic.
**Status**: Documented in test (`TestTOCTOUDefense`) and module docstring.

---

### LBC-002: `SchedulingConfig.from_business_offer()` Inverted Semantics Handling

**Location**: `src/autom8_scheduling/scheduling/offer_resolution.py:98`
**Code**: `scheduling_enabled=offer.disabled == False`
**What it does**: The equality comparison (not identity `is`) is load-bearing. aiomysql returns integer `0` for MySQL TINYINT columns, so `offer.disabled is False` would fail for enabled offers (disabled=0 is truthy, not `False`).
**Dependents**: All scheduling gate checks, config reads, PATCH writes.
**Naive fix failure mode**: Changing `== False` to `is False` silently breaks scheduling for all businesses whose `disabled` was returned as integer 0 by aiomysql. The gate would always return 403.
**Status**: Documented with `# noqa: E712` and regression test.

---

### LBC-003: `ApiField` Stub Silently Discards Validators

**Location**: `src/autom8_scheduling/models/_base.py:70-83`
**What it does**: Accepts `validator`, `roles`, `api_alias` kwargs and discards them — forwards only SQLModel/SQLAlchemy kwargs to `Field()`. This allows ORM models to carry API metadata parameters without runtime cost.
**Dependents**: All ORM model classes in `src/autom8_scheduling/models/scheduling.py` and `src/autom8_scheduling/models/shared.py`.
**Naive fix failure mode**: Making `ApiField` actually call `validator` functions would cause all ORM reads to run validators on every object instantiation, including DB reads where the data is already valid. This would add validation overhead to every query and could fail on legacy NULL values.
**Status**: Documented with comment in `_base.py`.

---

### LBC-004: `asyncio.gather(return_exceptions=True)` Block in Booking Routes

**Location**: `src/autom8_scheduling/api/routes/appointments/booking.py` (post-commit gather block)
**What it does**: After the MySQL appointment commit, all fire-and-forget side effects (GCal sync, GHL sync, notifications) are dispatched in a single `asyncio.gather(*coros, return_exceptions=True)`. The `return_exceptions=True` flag ensures one failure does not cancel other coroutines.
**Dependents**: All 3 booking routes (POST /book, POST /cancel, POST /reschedule). GCal sync, GHL sync, and notification dispatch are all gathered here.
**Naive fix failure mode**: Changing to `asyncio.gather(*coros)` (without `return_exceptions=True`) would cause a single sync failure to propagate as an exception and cancel remaining coroutines, violating the fire-and-forget invariants for all calendar systems.
**Safe change rule**: Any new fire-and-forget side effect must be added to this gather block, not called separately. Never remove `return_exceptions=True`.

---

### LBC-005: `TERMINAL_STATUSES` Deny-List Is the Slot-Blocking Authority

**Location**: `src/autom8_scheduling/scheduling/constants.py:43-53`
**What it does**: Any status NOT in `TERMINAL_STATUSES` blocks a time slot. This is the deny-list strategy adopted in ADR-scheduling-status-taxonomy. `SLOT_BLOCKING_STATUSES` (deprecated allow-list) is retained for reference but is not authoritative.
**Dependents**: `AvailabilityEngine` (SQLAlchemy query filter), `BookingEngine` overlap checks, `is_slot_blocking()` function.
**Naive fix failure mode**: Adding a status to `TERMINAL_STATUSES` releases all appointments with that status as available time slots. Incorrect additions (e.g., adding "pending") would make 8,711 pending appointments invisible to the engine.
**Safe change rule**: Only add statuses confirmed as terminal (appointment definitively over) with production data evidence. Any removal from `TERMINAL_STATUSES` requires the status to be re-validated as slot-blocking.

---

## Evolution Constraints

| Area | Changeability | Evidence |
|------|--------------|---------|
| `scheduling/exceptions.py` | Frozen (API contract) | ADR-scheduling-engine-exceptions. HTTP status code mapping is stable per C5 constraint. |
| `models/envelopes.py` | Frozen (API contract) | ADR-scheduling-envelope-pattern. Envelope shape is stable. |
| `scheduling/gcal_overlay.py` | Frozen (subtraction invariant) | INVARIANT: subtractive only. Cannot add slots. MySQL is source of truth. ADR explicitly marks as UNCHANGED. |
| `scheduling/gcal_sync.py` | Frozen for GHL activation sprint | ADR-scheduling-calendar-integration: "gcal_sync.py is unchanged." GHL work mirrors it. |
| `TERMINAL_STATUSES` in `constants.py` | Migration (requires ADR + production data evidence) | Changes affect what appointments block slots — double-booking risk on incorrect changes. |
| `CANCELLABLE_STATUSES` in `constants.py` | Coordinated (requires design review) | Changes affect what the service's own cancel/reschedule accepts. |
| Offer resolution algorithm in `offer_resolution.py` | Frozen (requires new ADR) | ADR-scheduling-offer-resolution defines the 2-tier resolution algorithm. |
| `ghl_sync.py` conditional import block | Coordinated (GHL activation sprint) | `GHL_AVAILABLE = False` is the activation gate. Changing requires GHL client library + reconcile endpoint. |
| `LBC-002` inverted semantics comparison | Frozen (correctness invariant) | `offer.disabled == False` must not become `is False`. aiomysql TINYINT constraint. |
| `LBC-004` gather block with `return_exceptions=True` | Safe (additive only) | Can add new coroutines to the gather. Must not remove `return_exceptions=True`. |
| `models/responses.py` | Safe (additive only) | Can add fields. Removing fields breaks API clients. |
| `models/schemas.py` | Safe (additive only) | New optional request fields are safe. |
| `api/routes/handlers.py` helper functions | Safe (internal) | Only called by route files in the same package. |
| `scheduling/notifications/` | Safe | No external dependencies on notification internals. |
| `.know/` files | Safe | Knowledge files, no runtime dependency. |

### In-Progress Evolutions

| Evolution | State | Location |
|-----------|-------|---------|
| mypy strict island expansion | In progress | `pyproject.toml` mypy overrides. Currently 5 modules. New non-migrated modules must be added per SCAR-008. |
| GHL calendar client library | Not implemented | `ghl_sync.py:39` — `GHL_AVAILABLE = False`. Awaiting `autom8y_ghl` library. |
| GHL reconcile endpoint | Deferred | Scoped out per capstone review H-03. Builds on `internal.py` GCal backfill pattern. |
| APScheduler for GCal/GHL reconciliation | Not implemented | `app.py` Phase 3 comment. |
| Per-offer routing via lead attribution chain | Not implemented | ADR-scheduling-offer-resolution mentions full attribution chain as future path. |
| Employee scheduling rules | Not implemented | `employee_ids_in`, `employee_ids_ex` on `BusinessOffer` are present but not consumed by the scheduling engine. |
| `dispatch_scheduling_notifications()` unification | Partial | v1 no-op for BookingEngine; Phase 3 target is full unification. |
| Typed dispatch interfaces for sync modules | Not started | `gcal_sync.py` and `ghl_sync.py` use `dict[str, Any]` for booking results (GAP-003). |

---

## Risk Zone Mapping

### RISK-001: Idempotency Scan O(N) Growth

**Location**: `src/autom8_scheduling/scheduling/booking_helpers.py:39-80` (`find_by_config_key`)
**TENSION link**: TENSION-005
**Missing guard**: No index on config JSON content. Scan size grows linearly with sms-ai appointments per `office_phone`.
**Evidence of awareness**: `scheduling_booking_scan_size` Prometheus histogram (REC-04) monitors scan size per operation. No documented alert threshold.
**Validation absent**: Config JSON deserialization in the scan loop (ST-2 defensive parsing) adds CPU overhead proportional to appointment count.
**Recommended guard**: Add indexed columns for idempotency keys (TENSION-005). Until then, monitor histogram p99 per business and add alert at scan size > 100.

---

### RISK-002: `check_business_scheduling()` Swallows DB Errors as 500

**Location**: `src/autom8_scheduling/api/routes/handlers.py:137-152`
**TENSION link**: None
**Missing guard**: DB errors during business lookup are caught generically and returned as 500. The error is logged but the client receives no actionable detail. Both the business lookup and offer resolution use this pattern.
**Evidence**: `except Exception as exc: logger.error("scheduling_gate_db_error", ...)`. Generic catch.
**Recommended guard**: Distinguish transient DB errors (5xx retry) from permanent errors (4xx). Circuit-breaker pattern would help at scale.

---

### RISK-003: No Guard on `resolve_business_offer()` Fallback Query When `Business.offer_id` is NULL

**Location**: `src/autom8_scheduling/scheduling/offer_resolution.py:49-59`
**TENSION link**: TENSION-008
**Missing guard**: The fallback query JOINs `chiropractors` where `Business.offer_id == BusinessOffer.offer_id`. If `Business.offer_id` is NULL for a business, the JOIN condition `NULL == offer_id` evaluates to NULL (correct SQL behavior). The query runs unnecessarily for businesses with `offer_id = NULL` on `chiropractors`.
**Evidence**: Fallback only executes when primary returns None, which is a natural guard. Risk is low but worth noting for performance profiling at scale.
**Recommended guard**: Explicit check for `business.offer_id is not None` before running fallback query (minor optimization, no correctness impact).

---

### RISK-004: `ghl_sync.py` Dict Key Access with No Type Safety

**Location**: `src/autom8_scheduling/scheduling/ghl_sync.py:56-80`, `src/autom8_scheduling/scheduling/gcal_sync.py:57-80`
**TENSION link**: TENSION-013, GAP-003
**Missing guard**: `booking_result["appointment_id"]` and related key accesses are not type-checked. If the dict keys change (e.g., the result structure is refactored), runtime `KeyError` would surface only at calendar sync time, post-commit, in the fire-and-forget block. The error would be swallowed by the outer `try/except Exception` guard and logged but never surface to the caller.
**Evidence of gap**: Both modules are in the `ignore_errors = true` mypy zone. The typed result classes in `results.py` are not used here.
**Recommended guard**: Update dispatch signatures to typed result objects (GAP-003 resolution). Expand mypy strict island to cover sync modules.

---

### RISK-005: `GHL_AVAILABLE = False` Is a Hardcoded Constant, Not a Conditional Import

**Location**: `src/autom8_scheduling/scheduling/ghl_sync.py:39`
**TENSION link**: TENSION-010
**Missing guard**: Unlike `gcal_sync.py` which uses a `try/except ImportError` to set `GCAL_AVAILABLE`, `ghl_sync.py` hardcodes `GHL_AVAILABLE = False`. If a `autom8y_ghl` library is installed but the conditional import block is not added, GHL sync will remain silently disabled even though the library is available.
**Evidence**: Line 39 comment: "No GHL client library exists yet." with the full conditional import pattern shown in a comment block.
**Recommended guard**: When activating GHL, replace the hardcoded `False` with the conditional import pattern (shown in the file's comment block). The ADR-scheduling-calendar-integration specifies exactly this.

---

## Knowledge Gaps

1. **GHL reconcile endpoint**: No implementation exists yet (deferred per H-03). When the GHL client library ships, a reconcile endpoint analogous to `internal.py` GCal backfill will be needed. The design is clear but the code does not exist.

2. **`appointments.ghl_event_id` column migration status**: ADR-scheduling-calendar-integration requires adding `ghl_event_id` to the `appointments` table. The ORM model field exists (`scheduling.py:85`), but whether the migration has been applied to production is not observable from source code alone.

3. **`business_offers.duration` column migration status**: ADR-scheduling-per-offer-wiring adds `duration` to `business_offers`. The ORM field exists (`scheduling.py:239-241`), but migration application status is not verifiable from source.

4. **Triple-write pipeline source identity**: The bulk import pipeline producing `none`/`called`/`contacted` statuses has a NULL source. The source system responsible is not identified in the codebase.

5. **`Employee.calendar_id` usage**: 18% of employees have `calendar_id` values per ADR-scheduling-per-offer-wiring. The field exists on the `Employee` model (`scheduling.py:202`) but is not consumed by the scheduling engine in current code. The path from employee calendar ID to scheduling logic is not documented.

6. **`offer_resolution.py` strict island**: As of this observation, the file comment does not confirm strict island membership. The pyproject.toml mypy override block should be verified to include `autom8_scheduling.scheduling.offer_resolution` in the strict list.
