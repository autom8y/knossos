---
domain: scar-tissue
generated_at: "2026-03-27T19:56:20Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4557333"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog

**15 scars identified** across git history, code markers, and test docstrings. The existing `.know/scar-tissue.md` (generated at commit `443da37`) contains 10 numbered scars. This observation adds 3 previously undocumented scars (ST-3, ST-4, and the "Datetime VARCHAR parsing" failure), resolves a SCAR-005 numbering conflict, and confirms all 10 documented scars are still present.

### SCAR-001: TOCTOU Race Condition (Booking Overlap)

**Category**: Integration failure / race condition

**What failed**: Single-query overlap check was susceptible to two concurrent bookings both passing a non-locking availability check, then both inserting overlapping appointments.

**When**: Pre-migration; embedded in monorepo scheduling service before extraction.

**Fix location**: `src/autom8_scheduling/scheduling/booking.py` — `check_overlap_nonlocking()` + `select_overlapping()` (with `with_for_update`). After WS-2 refactoring, the SELECT FOR UPDATE implementation moved to `src/autom8_scheduling/scheduling/booking_helpers.py`.

**Evidence**: `tests/test_booking_engine.py` line 4: `ST-1: TOCTOU race condition defense`. `booking_helpers.py` lines 8, 96, 128 reference ST-1.

---

### SCAR-002: JSON Config Field Defensive Parsing

**Category**: Schema evolution / integration failure

**What failed**: `appointment.config` is a JSON column, but aiomysql returns it as a Python `str` rather than a deserialized `dict` in some driver configurations. Code that assumed `config` was always a `dict` crashed with `AttributeError`.

**Fix location**: `src/autom8_scheduling/scheduling/booking_helpers.py` — consolidated `find_by_config_key()` (line 55: "Preserves ST-2 defensive JSON parsing"). Also inline in `BookingEngine.cancel()` and `BookingEngine.reschedule()` write paths in `src/autom8_scheduling/scheduling/booking.py`.

**Defensive pattern**:
```python
config = appointment.config or {}
if isinstance(config, str):
    try:
        config = json.loads(config)
    except (json.JSONDecodeError, TypeError):
        config = {}
```

**Regression test**: `tests/test_booking_engine.py` line 5: `ST-2: JSON config field defensive parsing`. `tests/test_offer_resolution.py` line 924: `test_st2_json_config_defensive_parsing`.

---

### SCAR-003: Phantom Columns on Business Model (Data Access Realignment)

**Category**: Schema evolution / config drift

**What failed**: The `Business` ORM model had four columns that do not exist in the production database: `scheduling_enabled`, `buffer_minutes`, `gcal_enabled`, `gcal_shadow_mode`. All calls to these columns silently returned Python defaults (False, 0), causing: scheduling gate always rejected (403), buffer always 0, GCal overlay never applied, PATCH /config crashed with `OperationalError`.

**Fix location**:
- `src/autom8_scheduling/models/shared.py` — four phantom columns removed
- `src/autom8_scheduling/models/scheduling.py` — `BusinessOffer` gained `buffer_minutes` and `gcal_shadow_mode` (migration 010 columns)
- `src/autom8_scheduling/scheduling/offer_resolution.py` — new module resolving config from `business_offers`
- `src/autom8_scheduling/api/routes/handlers.py`, `businesses.py`, `scheduling/validation.py` — rewired to `SchedulingConfig`

**Defensive pattern**: `SchedulingConfig` dataclass is the single authoritative interface. `SchedulingConfig.from_business_offer(offer)` handles inverted `disabled` semantics. `SchedulingConfig.not_configured()` is the safe default.

**Regression test**: `tests/test_offer_resolution.py` — full suite including inverted semantics, resolution chain, not-configured case.

---

### SCAR-004: aiomysql TINYINT Returns int 0, Not Python False

**Category**: Integration failure / type system

**What failed**: Code used `offer.disabled is False` (identity check) which failed for aiomysql-returned TINYINT values (integer 0), causing disabled offers to incorrectly appear enabled.

**Fix location**: `src/autom8_scheduling/scheduling/offer_resolution.py` line 98:
```python
scheduling_enabled=offer.disabled == False,  # noqa: E712 — aiomysql returns int 0 for TINYINT
```

Also in `src/autom8_scheduling/scheduling/engine.py` line 298:
```python
(Employee.enabled == True) | (Employee.enabled == None),  # noqa: E711, E712
```

**Comment marker**: `# noqa: E712 — aiomysql returns int 0 for TINYINT`

**Regression test**: `tests/test_offer_resolution.py` — `test_disabled_integer_zero_means_scheduling_enabled`, `test_disabled_integer_one_means_scheduling_disabled`.

---

### SCAR-005 (Datetime VARCHAR Parsing) — NUMBERING CONFLICT NOTE

**Category**: Schema evolution / integration failure

**What failed**: MySQL stores `appointment.start_datetime` and `appointment.end_datetime` as VARCHAR columns. aiomysql returns them as Python strings in two possible formats: ISO 8601 (`2026-04-06T10:00:00`) or legacy MySQL format (`2026-04-06 10:00:00`). Code that passed these directly to `datetime.fromisoformat()` failed on the legacy format.

**Fix location**: `src/autom8_scheduling/scheduling/engine.py` lines 368-385 — `_parse_appointment_datetime()` function. Used in `src/autom8_scheduling/scheduling/booking_helpers.py` (lines 99-100, 114-115, 131-132, 151-152, 188), `src/autom8_scheduling/scheduling/reminder.py` (lines 34-35, 76), and `src/autom8_scheduling/api/routes/internal.py` (lines 65-66).

**Defensive pattern**: Two-attempt parse: `datetime.fromisoformat()` first, then `datetime.strptime(dt_str, "%Y-%m-%d %H:%M:%S")`. Both paths add UTC timezone if tzinfo is None.

**Regression test**: `tests/test_availability_engine.py` class `TestAppointmentDatetimeParsing` (line 614): `test_iso_format_with_timezone`, `test_iso_format_without_timezone_assumes_utc`, `test_legacy_mysql_format`, `test_unparseable_datetime_raises_value_error`.

**Numbering conflict**: The test file `tests/test_availability_engine.py` (line 5 and 615) calls this SCAR-005. The existing `.know/scar-tissue.md` labels SCAR-005 as "OTel Span appointment_id Missing on Idempotent Path." These are two distinct scars. The OTel scar should be renumbered or this one should carry a distinct identifier. This observation uses SCAR-005 for the datetime parsing failure (matching test evidence) and documents the OTel scar below.

---

### SCAR-005-OTel: OTel Span appointment_id Missing on Idempotent Path

**Category**: Integration failure / observability

**What failed**: The booking span did not set `scheduling.appointment_id` on the idempotent path (when an existing appointment was found via idempotency_key). The span attribute was only set on fresh bookings, creating telemetry gaps.

**Fix location**: `src/autom8_scheduling/scheduling/booking.py` — idempotency success path. Commits: `c442e9e`, `f84b316`.

**Defensive pattern**: Always set `SCHEDULING_APPOINTMENT_ID` on spans before returning, including idempotent paths.

**Regression test**: Referenced in `tests/test_booking_engine.py` module docstring but not a discrete test class.

---

### SCAR-006: env_prefix Causing Production Env Var Resolution Failure

**Category**: Config drift

**What failed**: `SchedulingSettings` had `env_prefix=""` (empty), conflicting with infrastructure-provided environment variables using the `SCHEDULING_` prefix. All settings fell back to defaults in production.

**Fix location**: `src/autom8_scheduling/config.py`. Commit: `947dcf1`.

**Defensive pattern**: All env vars follow `SCHEDULING_{FIELD_NAME}`. ADR reference: Decision 11 of the env naming convention ADR.

---

### SCAR-007: GCalConfig Call-Arg Type Mismatch

**Category**: Type system

**What failed**: `GCalConfig(impersonation_target=..., timeout=...)` produces a mypy `call-arg` error because `autom8y-gcal` lacks typed stubs. Code is functionally correct but mypy strict mode rejects it.

**Fix location**: `src/autom8_scheduling/app.py` line 72: `# type: ignore[call-arg]`. Commits: `c9c49a0` (initial suppress), `39376b6` (restore after accidental removal — SCAR repeated due to the type ignore being stripped).

**Defensive pattern**: The `# type: ignore[call-arg]` must remain until `autom8y-gcal` exports typed stubs. SCAR-007 recurred once when the suppression was accidentally removed (commit `39376b6`).

---

### SCAR-008: mypy Strict Mode Cannot Apply to Migrated Scheduling Code

**Category**: Type system

**What failed**: Migrated `scheduling/` domain code uses SQLAlchemy 2.x raw patterns from the pre-typed monorepo era. Applying mypy strict mode produced hundreds of errors for functionally correct patterns.

**Fix location**: `pyproject.toml` — `[[tool.mypy.overrides]]` for `autom8_scheduling.scheduling.*` and `autom8_scheduling.api.routes.*` with `ignore_errors = true`. Commits: `9d163d1`, `66da42b`, `a19b432`.

**Defensive pattern**: Strict island expanding incrementally — new modules must NOT join the `ignore_errors = true` block. Current strict island: `scheduling.exceptions`, `scheduling.results`, `scheduling.constants`, `scheduling.write_ops`.

---

### SCAR-009: autom8y-telemetry Path Override Revert

**Category**: Integration failure

**What failed**: A local path override for `autom8y-telemetry` in `pyproject.toml` was accidentally left in place, causing CI to fail when the path was not available. Commit: `1cc7c87`.

**Fix location**: `pyproject.toml` `[tool.uv.sources]` section.

**Defensive pattern**: All autom8y SDK dependencies must point to the `autom8y` CodeArtifact index in CI. Never commit local path overrides without a CI guard.

---

### SCAR-010: Golden Traces Module Missing for CI Collection

**Category**: Integration failure / test infrastructure

**What failed**: `tests/golden_traces/` lacked `__init__.py` and proper module structure, causing pytest collection to fail in CI while passing locally. Commit: `f55ce70`.

**Fix location**: `tests/golden_traces/__init__.py`, `tests/golden_traces/conftest.py`, `tests/golden_traces/serializer.py`, `tests/golden_traces/span_tree.py`.

**Defensive pattern**: All test subdirectories must have `__init__.py` for `--import-mode=importlib` pytest collection.

---

### SCAR-012: `from __future__ import annotations` Breaks Runtime Type Inspection

**Category**: Type system / integration failure

**What failed**: Adding `from __future__ import annotations` to modules that use runtime type inspection (pytest fixtures with `AsyncGenerator`, SQLModel field validators, frozen dataclasses with `isinstance()`) caused `NameError` at runtime — annotations became lazy strings instead of resolved types.

**Fix location**: The absence of the import IS the fix. Comment marker placed at top of affected files.

**Defensive pattern**: Comment `# Note: from __future__ import annotations intentionally omitted (SCAR-012).` present in:
- `tests/golden_traces/conftest.py` (line 10)
- `tests/golden_traces/test_golden_traces.py` (line 9)
- `tests/golden_traces/span_tree.py` (line 10)
- `tests/golden_traces/serializer.py` (line 10)
- `tests/test_booking_engine.py` (line 16)
- `tests/test_offer_resolution.py` (line 13)
- `tests/test_availability_engine.py` (line 7)
- `tests/test_ghl_sync.py` (line 11)
- `tests/test_status_taxonomy.py` (line 15)
- `tests/test_per_offer_duration.py` (line 15)
- `src/autom8_scheduling/scheduling/booking_helpers.py` (line 12)

---

### SCAR-022 / DEF-009: uv sync --frozen vs --no-sources in CI

**Category**: Config drift / integration failure

**What failed**: CI pipeline used `uv sync --frozen` which failed when the private `autom8y` package index was unavailable or the lock file referenced paths that changed. Commit: `63621a4`.

**Fix location**: `.github/workflows/` CI configuration.

**Comment marker**: Commit message: `fix(ci): replace --frozen with --no-sources in uv sync (DEF-009/SCAR-022)`.

**Defensive pattern**: Use `uv sync --no-sources` in CI where the private index may not be accessible.

---

### ST-3: GCal Fire-and-Forget Invariant

**Category**: Integration failure / data integrity

**What failed**: GCal write failures were previously not isolated from MySQL booking outcomes. A GCal failure could roll back or block a MySQL booking commit, creating false negatives for the customer.

**Fix location**: `src/autom8_scheduling/scheduling/gcal_sync.py` lines 1-4 — module-level INVARIANT docstring:
```
INVARIANT: GCal failure NEVER rolls back a MySQL booking.
INVARIANT: appointments.event_id is the ONLY binding column.
```
Fire-and-forget wrapper pattern wraps all GCal calls in `try/except Exception`.

**Regression test**: `tests/test_offer_resolution.py` line 951: `test_st3_gcal_fire_and_forget` — asserts `"INVARIANT: GCal failure NEVER rolls back"` is present in source. `tests/test_booking_engine.py` line 6: `ST-3: GCal fire-and-forget (verified via no-touch assertion)`.

---

### ST-4: GCal Overlay Subtractive-Only Invariant

**Category**: Integration failure / data integrity

**What failed**: The GCal overlay (which subtracts Google Calendar busy periods from MySQL-derived availability slots) was at risk of adding slots rather than only removing them if the overlay logic were modified incorrectly. MySQL is the authoritative source of truth; GCal can only reduce available slots.

**Fix location**: `src/autom8_scheduling/scheduling/gcal_overlay.py` lines 1-3 — module-level INVARIANT docstring:
```
INVARIANT: Subtractive only. Cannot add slots. MySQL is source of truth.
```
Also enforced at line 119 within `_apply_busy_periods()`: `"""Subtract busy periods from candidate slots. INVARIANT: Can only remove slots."""`

**Regression test**: `tests/test_offer_resolution.py` line 961: `test_st4_gcal_subtractive_only` — asserts `"INVARIANT: Subtractive only"` is present in source.

---

## Category Coverage

6 distinct failure mode categories observed:

| Category | Scars |
|----------|-------|
| Integration failure (general) | SCAR-001, SCAR-002, SCAR-004, SCAR-005-OTel, SCAR-009, SCAR-010, ST-3, ST-4 |
| Schema evolution | SCAR-002, SCAR-003, SCAR-005 |
| Config drift | SCAR-003, SCAR-006, SCAR-022/DEF-009 |
| Type system | SCAR-004, SCAR-007, SCAR-008, SCAR-012 |
| Observability gap | SCAR-005-OTel |
| Data integrity / invariant violation | ST-3, ST-4 |

**Searched but not found**:
- **Data corruption**: No observed cases of corruption in appointment records.
- **Security**: No security-specific scars; auth is delegated to `autom8y-auth` SDK.
- **Performance cliff**: Awareness exists (REC-04 scan size histogram metric in engine.py), but no confirmed cliff events documented.
- **Idempotency failure**: ST-5 (idempotency via JSON config keys) represents a preventive pattern rather than a post-failure fix — no confirmed production idempotency failure recorded.

---

## Fix-Location Mapping

| Scar | Primary Fix File | Function / Section | File Exists |
|------|-----------------|---------------------|-------------|
| SCAR-001 | `src/autom8_scheduling/scheduling/booking_helpers.py` | `check_overlap_nonlocking()`, `select_overlapping()` | Yes |
| SCAR-002 | `src/autom8_scheduling/scheduling/booking_helpers.py` | `find_by_config_key()` (line 55); also `booking.py` cancel/reschedule | Yes |
| SCAR-003 | `src/autom8_scheduling/models/shared.py`, `scheduling.py`, `offer_resolution.py`, `api/routes/handlers.py` | Multiple — structural refactor | Yes |
| SCAR-004 | `src/autom8_scheduling/scheduling/offer_resolution.py:98`, `engine.py:298` | `SchedulingConfig.from_business_offer()`, `_query_available_employees()` | Yes |
| SCAR-005 | `src/autom8_scheduling/scheduling/engine.py:368-385` | `_parse_appointment_datetime()` | Yes |
| SCAR-005-OTel | `src/autom8_scheduling/scheduling/booking.py` | Idempotency success path, span attribute set | Yes |
| SCAR-006 | `src/autom8_scheduling/config.py` | `SchedulingSettings.model_config` | Yes |
| SCAR-007 | `src/autom8_scheduling/app.py:72` | `GCalConfig` initialization | Yes |
| SCAR-008 | `pyproject.toml` | `[[tool.mypy.overrides]]` block | Yes |
| SCAR-009 | `pyproject.toml` | `[tool.uv.sources]` | Yes |
| SCAR-010 | `tests/golden_traces/__init__.py` | Module init (absent = the bug) | Yes |
| SCAR-012 | 11 files (see catalog) | Module header (absence of import = the fix) | Yes (all 11) |
| SCAR-022/DEF-009 | `.github/workflows/` CI config | `uv sync` command | Yes |
| ST-3 | `src/autom8_scheduling/scheduling/gcal_sync.py:1-4` | Module docstring INVARIANT + try/except wrapping | Yes |
| ST-4 | `src/autom8_scheduling/scheduling/gcal_overlay.py:1-3,119` | Module docstring INVARIANT + `_apply_busy_periods()` | Yes |

**Compound fixes**: SCAR-003 (data access realignment) is the most complex — spans 4+ files and required a new module (`offer_resolution.py`). SCAR-012 spans 11 files. SCAR-004 appears in 2 separate files with distinct fix sites.

---

## Defensive Patterns

| Scar | Defensive Pattern | Regression Test Present |
|------|------------------|------------------------|
| SCAR-001 | Two-phase overlap check (non-locking fast-fail + SELECT FOR UPDATE re-check) | Yes — `TestTOCTOUDefense` in `test_booking_engine.py`, `test_st1_toctou_two_phase_overlap_check` in `test_offer_resolution.py` |
| SCAR-002 | `isinstance(config, str)` guard before `.get()` / `json.loads()` fallback | Yes — `tests/test_booking_engine.py` ST-2, `test_st2_json_config_defensive_parsing` |
| SCAR-003 | `SchedulingConfig` dataclass as single authoritative interface; `not_configured()` safe default | Yes — full `tests/test_offer_resolution.py` suite |
| SCAR-004 | `== False` (equality) not `is False` (identity) for aiomysql TINYINT; `# noqa: E712` suppression | Yes — `test_disabled_integer_zero_means_scheduling_enabled` |
| SCAR-005 | `_parse_appointment_datetime()` two-attempt parse (ISO then legacy MySQL format) | Yes — `TestAppointmentDatetimeParsing` class in `test_availability_engine.py` |
| SCAR-005-OTel | Set `SCHEDULING_APPOINTMENT_ID` on all span return paths including idempotent | No discrete test class observed |
| SCAR-006 | `env_prefix="SCHEDULING_"` enforced in `SchedulingSettings` | No explicit regression test observed |
| SCAR-007 | `# type: ignore[call-arg]` at `GCalConfig` init; must not be removed | No explicit regression test; CI mypy gate guards it |
| SCAR-008 | `ignore_errors = true` in pyproject.toml for migrated modules; strict island expansion | No explicit regression test; mypy CI gate guards it |
| SCAR-009 | Use CodeArtifact registry source in `[tool.uv.sources]`; never commit path overrides | No explicit regression test; CI guards it |
| SCAR-010 | All test subdirectories must have `__init__.py` | No explicit regression test; CI collection guards it |
| SCAR-012 | Comment `# Note: from __future__ import annotations intentionally omitted (SCAR-012).` | Yes — `TestScarTissuePreservation.test_st2_json_config_defensive_parsing` indirectly verifies; `test_offer_resolution.py:13` notes the omission |
| SCAR-022/DEF-009 | `uv sync --no-sources` in CI pipelines | No explicit regression test; CI gate guards it |
| ST-3 | `try/except Exception` wrapping all GCal write calls; module-level INVARIANT comment | Yes — `test_st3_gcal_fire_and_forget` |
| ST-4 | Subtractive-only overlay logic; module-level INVARIANT comment | Yes — `test_st4_gcal_subtractive_only` |

**Scars with no dedicated regression test** (guarded only by CI gates): SCAR-005-OTel, SCAR-006, SCAR-007, SCAR-008, SCAR-009, SCAR-010, SCAR-022/DEF-009.

---

## Agent-Relevance Tags

| Scar | Agent Responsibility Area | Why the Agent Needs This |
|------|--------------------------|--------------------------|
| SCAR-001 | principal-engineer, qa-adversary | Any modification to booking overlap logic must preserve both phases of the TOCTOU check. Single-query replacement silently reintroduces the race. |
| SCAR-002 | principal-engineer | Any code reading `appointment.config` must apply the `isinstance(config, str)` guard. The aiomysql driver behavior is permanent. |
| SCAR-003 | principal-engineer, architect | No agent should add columns back to the `Business` model. Config lives on `business_offers` via `SchedulingConfig`. |
| SCAR-004 | principal-engineer, qa-adversary | Any agent writing queries against TINYINT columns from aiomysql must use `== False` not `is False`. |
| SCAR-005 | principal-engineer | Any new code consuming `start_datetime` or `end_datetime` from appointment rows must call `_parse_appointment_datetime()`, not `datetime.fromisoformat()` directly. |
| SCAR-005-OTel | observability-engineer | Any new span around booking operations must set appointment_id on all return paths, not just the fresh booking path. |
| SCAR-006 | platform-engineer, principal-engineer | Any agent modifying `config.py` or adding new settings must maintain `SCHEDULING_` prefix. Never set `env_prefix=""`. |
| SCAR-007 | principal-engineer, platform-engineer | Do not remove `# type: ignore[call-arg]` at `app.py:72` until `autom8y-gcal` exports typed stubs. |
| SCAR-008 | principal-engineer | New modules must NOT join the `ignore_errors = true` override block. Strict island must only grow, not shrink. |
| SCAR-009 | platform-engineer | Before committing `pyproject.toml`, verify `[tool.uv.sources]` has no local path overrides for autom8y packages. |
| SCAR-010 | qa-adversary, principal-engineer | New test subdirectories require `__init__.py`. |
| SCAR-012 | principal-engineer, qa-adversary | Do not add `from __future__ import annotations` to files marked with the SCAR-012 comment. Check before modifying file headers. |
| SCAR-022/DEF-009 | platform-engineer | Use `uv sync --no-sources` in CI. Do not revert to `--frozen`. |
| ST-3 | principal-engineer, architect | GCal write operations must be wrapped in `try/except Exception`. GCal failure must never roll back a MySQL booking. |
| ST-4 | principal-engineer, qa-adversary | GCal overlay logic must only remove slots. Any change to `_apply_busy_periods()` in `gcal_overlay.py` must verify the INVARIANT holds. |

---

## Knowledge Gaps

1. **SCAR-005 numbering conflict**: The existing `.know/scar-tissue.md` assigns SCAR-005 to the OTel span idempotency scar, but `tests/test_availability_engine.py` (lines 5 and 615) assigns SCAR-005 to the Datetime VARCHAR parsing scar. These are two distinct scars with the same number. A principal-engineer should resolve the numbering (suggest: OTel scar becomes SCAR-011, or a new slot number beyond 022).

2. **SCAR-005-OTel lacks a discrete regression test**: The OTel idempotent path fix (commits `c442e9e`, `f84b316`) has no `TestScarTissuePreservation`-style test asserting the span attribute is set on idempotent returns.

3. **SCAR-006, SCAR-009, SCAR-010, SCAR-022 have no regression tests**: These are config drift and CI infrastructure scars guarded only by CI execution. A re-introduction could pass locally before being caught. The `TestScarTissuePreservation` pattern (as applied to ST-1 through ST-4) should be extended to cover these.

4. **SCAR gap in numbering**: SCAR numbers observed are 001-010, 012, 022. SCARs 011, 013-021 are unaccounted for. These may come from the monorepo history (pre-extraction) and are outside this repo's git history.

5. **ST-5 (idempotency via JSON config keys) is named in code markers** (`booking_helpers.py` line 10, `test_booking_engine.py` line 7) but does not have a corresponding SCAR-NNN number in the `.know/scar-tissue.md`. It represents a design scar (three duplicate methods consolidated via GAP-001) with a regression test (`test_st5_idempotency_via_json_config_keys` in `test_offer_resolution.py`).
