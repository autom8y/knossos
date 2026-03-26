---
domain: scar-tissue
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

# Codebase Scar Tissue

## Defensive Patterns

### TOCTOU Race Condition Defense (booking.py)

The booking engine implements a two-phase overlap check to defend against time-of-check/time-of-use (TOCTOU) race conditions.

**Pattern**: `src/autom8_scheduling/scheduling/booking.py`, lines 116-162

- Phase 1 (Step 2): `_check_overlap_nonlocking()` — fast, non-locking scan. Early-exit if conflict found.
- Phase 2 (Step 3): `_select_overlapping()` — `SELECT FOR UPDATE`. Acquires row locks before INSERT.

Documented in `ADR-002`. The two-step design: the expensive lock is only acquired if the cheaper check passes.

### JSON Config Field Defensive Parsing (booking.py)

Three idempotency lookup helpers perform defensive JSON deserialization on `config` values:

```python
if isinstance(config, str):
    try:
        config = json.loads(config)
    except (json.JSONDecodeError, TypeError):
        continue
```

**Locations**: `booking.py` lines 703-708, 832-836, 862-866

Guards against known database inconsistency: the `config` column can be returned as a raw string by older SQLAlchemy/MySQL driver versions.

### Appointment Datetime Multi-Format Parser (engine.py)

`_parse_appointment_datetime()` attempts two parse formats:

**Location**: `src/autom8_scheduling/scheduling/engine.py`, lines 353-370

1. `datetime.fromisoformat()` (ISO 8601)
2. `datetime.strptime(dt_str, "%Y-%m-%d %H:%M:%S")` (MySQL-style)

The fallback documents a known data quality issue: appointment datetimes in the existing database were stored in non-ISO format. Unparseable rows are skipped with `logger.warning("unparseable_appointment_datetime")`.

### Scan Size Histogram (engine.py, booking.py)

Two `Histogram` metrics registered for observability of full-table scan sizes:

- `scheduling_availability_scan_size` in `engine.py` lines 34-39
- `scheduling_booking_scan_size` in `booking.py` lines 37-42

Comment: `# REC-04: Appointment scan size histogram`. Documents that idempotency lookup performs a full table scan.

### GCal Fire-and-Forget Exception Swallowing (gcal_sync.py)

All three GCal sync operations wrap their entire body in `try/except Exception` and log without re-raising.

Module-level invariant at line 3: `INVARIANT: GCal failure NEVER rolls back a MySQL booking.`

### GCal Subtractive-Only Invariant (gcal_overlay.py)

`_subtract_busy_periods()` carries: `INVARIANT: Can only remove slots.`

`filter_availability()` returns the original unfiltered result when `shadow_mode=True`, preventing any behavioral impact until shadow mode is explicitly disabled per business.

### Pool Pre-Ping and Pool Recycle (app.py)

`pool_pre_ping=True` guards against connections silently severed by MySQL's `wait_timeout`. `pool_recycle=3600` recycles connections at 1-hour intervals.

### Liveness vs. Readiness Separation (health.py)

`/health` performs zero I/O. `/ready` performs `SELECT 1` MySQL probe. `/health/deps` probes both MySQL and GCal. `_check_gcal()` returns `DEGRADED` (not `UNAVAILABLE`) if the GCal client is None.

### GCal FreeBusy In-Memory Cache (gcal_overlay.py)

`_FreeBusyCache` implements TTL expiry and LRU-style eviction at max capacity (1000 entries). Cache TTL constants documented as `ADR-GCAL-OV-003`: `CACHE_TTL_EVENTUAL = 300`, `CACHE_TTL_STRICT = 0`.

### Graceful Degradation on Import Failures

Five distinct optional dependency guards in `app.py`, `gcal_overlay.py`, and `gcal_sync.py`. All return `None` or no-op rather than crashing the service.

## Workarounds and Hacks

### ApiField Stub (models/_base.py)

`src/autom8_scheduling/models/_base.py`, lines 48-83

`ApiField` accepts `roles`, `validator`, and `api_alias` kwargs and silently discards them. Deliberate compatibility shim: models migrated from `autom8y-data` which has a full `ApiField` implementation.

### READ-ONLY Model Copies (models/shared.py)

`Business` and `Lead` are duplicated ORM models pointing at the same tables (`chiropractors`, `leads`). Structural workaround for the monorepo-to-satellite extraction.

### `from __future__ import annotations` Omission for SCAR-012

All four golden traces modules carry: `Note: from __future__ import annotations intentionally omitted (SCAR-012).`

Documents a known incompatibility between `from __future__ import annotations` and runtime annotation evaluation.

### `noqa: E711, E712` on SQLAlchemy Boolean/None Comparison

`src/autom8_scheduling/scheduling/engine.py`, line 286

```python
(Employee.enabled == True) | (Employee.enabled == None),  # noqa: E711, E712
```

SQLAlchemy requires `==` operator for correct SQL generation. This is a SQLAlchemy API limitation workaround.

### GCal Shadow Mode Feature Flag

`gcal_shadow_mode: bool = Field(default=True)` — deployed in shadow mode by default. An explicit "ship it but don't use it" workaround for deploying new integration code safely.

### `type: ignore[call-arg]` on GCalConfig (app.py)

`src/autom8_scheduling/app.py`, line 73. The `GCalConfig` constructor requires `GOOGLE_SA_KEY_JSON` resolved from env internally. This suppression has been restored twice (commits `c9c49a0` and `39376b6`) after being accidentally removed.

## Bug Fix Archaeology

### Production Env Var Resolution Bug (commits 947dcf1, 449884e)

**Hotspot**: `src/autom8_scheduling/config.py`

1. **947dcf1**: `SchedulingSettings` had `env_prefix="SCHEDULING_"` causing pydantic-settings to look for `SCHEDULING_AUTOM8Y_ENV` instead of `AUTOM8Y_ENV`. In production, `autom8y_env` defaulted to `LOCAL`, causing localhost MySQL connection.

2. **449884e**: Re-introduced `env_prefix="SCHEDULING_"` properly scoped with explicit component fields per ADR-ENV-NAMING-CONVENTION Decision 11.

### OTel Span Missing Required Attribute (commits f84b316, c442e9e)

**Hotspot**: `src/autom8_scheduling/scheduling/booking.py`

Two sequential fixes for missing `scheduling.appointment_id` attribute on spans. OTel convention enforcement via CI gates caught attribute gaps that tests had not covered.

### uv `--frozen` / `--no-sources` Mutual Exclusivity (commit 63621a4)

**Reference**: `DEF-009/SCAR-022`. uv >= 0.15.4 made `--frozen` and `--no-sources` mutually exclusive.

### autom8y-telemetry Path Override Revert (commit 1cc7c87)

A `pyproject.toml` override for local monorepo path had to be reverted when CI pulled from CodeArtifact. Pattern: monorepo-to-satellite extraction creates path vs. registry resolution tension.

### mypy Relaxation Sequence (commits 9d163d1, 66da42b, a19b432, c9c49a0)

Four sequential mypy suppression commits documenting the migration of untyped code from the monorepo.

## Failure Catalog

### Booking Conflict Failures

| Outcome | `conflict_reason` | Trigger |
|---|---|---|
| `status: "conflict"` | `"business_not_found"` | `office_phone` has no matching Business row |
| `status: "conflict"` | `"slot_taken"` | Non-locking pre-check finds overlap |
| `status: "conflict"` | `"slot_taken"` | Locking SELECT FOR UPDATE finds overlap |
| `status: "idempotent_success"` | n/a | Prior booking found via idempotency key |

### Idempotency Lookup Linear Scan

All three idempotency lookups perform full table scans filtered by `config IS NOT NULL` and `source = "sms-ai"`, then do in-Python key matching. The `scheduling_booking_scan_size` histogram (REC-04) was added to monitor this.

### GCal Sync Failure Modes

Silent failures caught and logged (no re-raise):
- `calendar_id` is `None`: logs `gcal_sync_skipped` with `reason: "no_calendar_id"`
- `event_id` is null: logs `gcal_sync_skipped` with `reason: "no_event_id"`
- GCal API call throws: logs `gcal_sync_error`
- `_store_event_id()` finds no appointment: logs `gcal_sync_store_event_id_failed`

### Timezone Not Configured

`AvailabilityEngine` raises `TimezoneNotConfiguredError` when `addresses.timezone` is NULL.

### Availability Date Range Limits

- Date range exceeding 14 days: `ValueError`
- `end_date < start_date`: `ValueError`
- `start_date < today`: silently adjusted to `today`
- `end_date < today`: returns empty availability

### Missing/Invalid Appointment Duration

If `appt_duration` is `None` or `<= 0`: falls back to `_DEFAULT_APPT_DURATION = 30`. Logged at INFO.

### Status Terminal State Transitions

Any status not in `VALID_TRANSITIONS` is treated as terminal. Unknown statuses fail closed (tested explicitly: `test_unknown_status_is_treated_as_terminal`).

### Database Unavailable at Startup

`_initialize_mysql_pool()` catches `RuntimeError` and returns `None`. App starts in degraded state.

## Type Safety Escapes

### `# type: ignore[attr-defined]` — Appointment.created ORM Column

`src/autom8_scheduling/services/appointment.py`, line 237. `Appointment.created.desc()` method not visible to mypy through SQLModel stubs.

### `# type: ignore[call-arg]` — GCalConfig Constructor

`src/autom8_scheduling/app.py`, line 73. Restored twice after accidental removal.

### `# type: ignore[union-attr]` — BusinessOffer.master_cal_id

`src/autom8_scheduling/scheduling/gcal_sync.py` line 377, `gcal_overlay.py` line 267. Known SQLAlchemy + mypy limitation with `.is_not()` on nullable columns.

### `# type: ignore[arg-type]` — Appointment.appointment_id in WHERE

`src/autom8_scheduling/scheduling/gcal_sync.py` lines 406 and 441. mypy infers broader type from dict access.

### `# type: ignore[attr-defined]` in Tests

`tests/test_write_ops.py` lines 214-217, 234. `logging.LogRecord` does not have typed attributes for `extra` dict fields.

### Pervasive `Any` Typing

Used for external optional dependency interfaces (`gcal_client: Any`) and JSON-shaped data bags (`dict[str, Any]`). Concentrated in: `api/routes/__init__.py`, `scheduling/booking.py`, `scheduling/engine.py`, `scheduling/gcal_sync.py`, `scheduling/gcal_overlay.py`.

## Test-Encoded Knowledge

### Idempotency Is Step 0 (golden_traces/conftest.py)

The `mock_booking_engine` fixture explicitly stubs `_find_by_idempotency_key` to return `None`. Encodes the knowledge that idempotency lookup is the first step of every booking flow.

### Transition Target Sets Must Be Frozenset (test_write_ops.py)

Explicit test that `VALID_TRANSITIONS` values are `frozenset`, not `set`. Encodes that mutable transition sets would allow accidental runtime modification.

### Unknown Status Treated as Terminal (test_write_ops.py)

Documents the deliberate design decision that unknown statuses fail closed (treated as terminal).

### rescheduled and reschedule Symmetric Transitions (test_write_ops.py)

Documents production data reality: two overlapping status strings exist and both must have identical transition rules.

### `has_9_canonical_statuses` — Production Status Count (test_write_ops.py)

Encodes: there are more status strings in the database (16) than the canonical enum exposes (9).

### OTel Convention Compliance (test_sprint2_instrumentation.py)

The entire test suite was added to verify OTel spans carry correct typed attribute keys. Exists because two `REQUIRED_MISSING` violations were caught by CI convention checks.

### SCAR-012 — `from __future__ import annotations` Incompatibility

Four golden_traces modules carry `SCAR-012` reference tag. Prevents the fix from being reintroduced by contributors who see it as a missing best practice.

## Knowledge Gaps

1. **SCAR-012 content not visible**: The SCAR-012 catalog entry is referenced but its artifact is not present in this repo.
2. **DEF-009 and SCAR-022 content not visible**: Referenced in commit `63621a4`.
3. **TDD-SCHED-EXTRACT-001 migration document not present**: Multiple files reference this extraction ticket.
4. **ADR-002 full text not located**: Referenced in `booking.py` module docstring. May reside in the monorepo.
5. **No tests for availability engine edge cases**: Defensive patterns like `_parse_appointment_datetime()` with MySQL-style format, `_generate_block_candidates()` with malformed JSON, and `_FreeBusyCache` TTL/eviction have no test coverage.
6. **`REC-04` origin not documented**: The `# REC-04: Appointment scan size histogram` comment references a recommendation catalog but no document explains it.
