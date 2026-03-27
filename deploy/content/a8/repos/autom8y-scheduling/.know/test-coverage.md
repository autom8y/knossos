---
domain: test-coverage
generated_at: "2026-03-27T19:56:20Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4557333"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Source Modules Without Corresponding Test Files

The source tree contains 46 Python files (excluding `__init__.py` and non-code). Mapping each to test coverage:

**Tested (direct coverage):**
- `scheduling/engine.py` — `test_availability_engine.py`, `test_per_offer_duration.py`
- `scheduling/booking.py` — `test_booking_engine.py`, `golden_traces/test_golden_traces.py`
- `scheduling/booking_helpers.py` — indirectly via `test_booking_engine.py` (helpers are patched, not unit-tested directly)
- `scheduling/write_ops.py` — `test_write_ops.py`
- `scheduling/constants.py` — `test_status_taxonomy.py`
- `scheduling/offer_resolution.py` — `test_offer_resolution.py`, `test_per_offer_duration.py`
- `scheduling/ghl_sync.py` — `test_ghl_sync.py`
- `scheduling/validation.py` — `test_per_offer_duration.py` (via `check_scheduling_readiness`)
- `scheduling/results.py` — `test_booking_engine.py`
- `scheduling/exceptions.py` — `test_booking_engine.py`, `test_routes_appointments.py`
- `models/envelopes.py` — `test_model_layer.py`
- `models/schemas.py` — `test_model_layer.py`
- `models/responses.py` — `test_model_layer.py`
- `api/routes/businesses.py` — `test_routes_businesses.py`
- `api/routes/appointments/availability.py` — `test_routes_appointments.py`
- `api/routes/appointments/booking.py` — `test_routes_appointments.py`
- `api/routes/appointments/queries.py` — `test_routes_appointments.py`
- `health.py` — `test_health.py`
- `app.py` — via `conftest.py` fixture
- `config.py` — via `conftest.py` fixture (settings cache_clear pattern)

**Untested (no test file, no indirect coverage observed):**
- `scheduling/notifications/sms_service.py` — no test file
- `scheduling/notifications/email_service.py` — no test file
- `scheduling/notifications/dispatch.py` — partially patched in golden_traces (tracer patched), but dispatch logic itself is not exercised in any test
- `scheduling/notifications/templates.py` — no test file
- `scheduling/notifications/schemas.py` — no test file
- `scheduling/notifications/config.py` — no test file
- `scheduling/reminder.py` — no test file
- `scheduling/gcal_overlay.py` — patched/mocked in instrumentation tests but no direct unit tests
- `scheduling/gcal_sync.py` — patched at tracer level only; sync logic untested
- `api/routes/handlers.py` — no test file
- `api/routes/internal.py` — no test file
- `api/routes/metrics.py` — no test file
- `api/routes/deps.py` — indirectly exercised via conftest dependency override; not directly tested
- `services/appointment.py` — no test file
- `services/base.py` — no test file
- `models/scheduling.py` — no test file (SQLModel ORM definitions)
- `models/shared.py` — no test file
- `models/_base.py` — no test file
- `scheduling/gcal_overlay.py` — no direct test; used in golden_traces but tracer is patched

### Coverage of Critical Paths

**API Handlers:** The `api/routes/appointments/` sub-package is covered via `test_routes_appointments.py` and `test_routes_businesses.py` at the HTTP layer. `api/routes/handlers.py` and `api/routes/internal.py` have no coverage.

**Core Business Logic:** `booking.py`, `engine.py`, `write_ops.py`, `offer_resolution.py`, and `constants.py` are all covered. `booking_helpers.py` is patched in tests but not directly unit-tested — its functions (`find_by_config_key`, `resolve_business`, `check_overlap_nonlocking`, `select_overlapping`) are always mocked, meaning real query logic is never exercised.

**Data Access:** The session layer is always mocked via `AsyncMock`. No test exercises real SQL. `services/appointment.py` — which wraps the appointment query service — has zero test coverage.

**Notifications:** The entire `scheduling/notifications/` sub-package (6 files: `dispatch.py`, `sms_service.py`, `email_service.py`, `templates.py`, `schemas.py`, `config.py`) is unguarded. The `dispatch.py` `_notification_tracer` is patched in golden_traces, but no test exercises actual dispatch logic or SMS/email rendering.

**Reminder pipeline:** `scheduling/reminder.py` has no test file and no indirect test coverage.

**GCal sync:** `scheduling/gcal_sync.py` and `scheduling/gcal_overlay.py` are instrumented and patched for OTel tests, but actual sync operations are never exercised in tests.

### Negative Test Coverage

Negative tests are present for critical paths: `BookingConflict`, `AppointmentNotFound`, `AppointmentAlreadyCancelled`, `AppointmentNotCancellable`, `RescheduleConflict`, `SchedulingGateError`, `TimezoneNotConfiguredError`, and HTTP 404/400/503 responses. The business routes (`test_routes_businesses.py`) include refusal and prereq-not-met paths.

### Prioritized Gap List

1. **CRITICAL:** `scheduling/notifications/` sub-package (6 files) — zero unit tests on SMS/email send paths, template rendering, dispatch conditional logic
2. **HIGH:** `scheduling/booking_helpers.py` — real query logic is always mocked; no test exercises the actual SQLAlchemy query construction
3. **HIGH:** `services/appointment.py` — appointment query service; no test file
4. **HIGH:** `scheduling/reminder.py` — reminder eligibility and dispatch logic; no test file
5. **MEDIUM:** `api/routes/handlers.py`, `api/routes/internal.py` — internal endpoints untested
6. **MEDIUM:** `scheduling/gcal_sync.py` / `scheduling/gcal_overlay.py` — sync and overlay logic exercised only at the tracer level
7. **LOW:** `models/scheduling.py`, `models/shared.py`, `models/_base.py` — ORM definitions, low priority for unit tests but schema validation not tested

### Coverage Threshold Note

The CI pipeline runs with `coverage_threshold: 20` — an extremely low bar. This confirms that large untested areas are known and accepted (or deferred).

---

## Testing Conventions

### Test Function Naming

All tests use **class-based organization** (`class TestXxx`) with `snake_case` method names that follow the pattern `test_{scenario}_{expected_outcome}`. Examples:
- `test_book_returns_booking_result`
- `test_nonlocking_overlap_raises_conflict`
- `test_happy_path_returns_availability`
- `test_enable_refused_prerequisites_not_met`

The naming convention is consistent across all 15 test files.

### Assertion Patterns

Three distinct assertion patterns are used:

1. **pytest plain assert** — used throughout for attribute and type checks:
   ```python
   assert isinstance(result, BookingResult)
   assert result.appointment_id == SYNTHETIC_APPT_ID
   ```

2. **`pytest.raises` context manager** — used for all exception paths with `exc_info` attribute inspection:
   ```python
   with pytest.raises(BookingConflict) as exc_info:
       ...
   assert exc_info.value.conflict_reason == "slot_taken"
   ```

3. **Custom envelope helpers** — `tests/helpers.py` provides `assert_success()` and `assert_error()` for HTTP response envelope unwrapping. All route tests use these exclusively.

No custom matchers or third-party assertion libraries (e.g., assertpy, hamcrest) are used. `respx` is used in `test_proxy.py` for HTTP mock assertion.

### Test Helper and Fixture Patterns

**Factory functions (module-level):** Test data is built via module-level helpers prefixed with `_make_`:
- `_make_mock_appointment()` in `test_booking_engine.py`
- `_make_mock_offer()` in `test_offer_resolution.py`, `test_per_offer_duration.py`, `test_ghl_sync.py`
- `_make_scheduling_config()` in `test_routes_appointments.py`
- `_make_hours_row()`, `_make_appointment_row()`, `_scalar_one_or_none_result()`, `_all_result()` in `test_availability_engine.py`

**Engine stub helpers:** `test_booking_engine.py` uses `_stub_engine_for_booking()`, `_stub_engine_for_cancel()`, `_stub_engine_for_reschedule()` — helper functions that call `patch(...).start()` and return the patch list.

**`MagicMock` / `AsyncMock` pattern:** The database session is always an `AsyncMock` with `side_effect` lists for sequential `execute()` calls. This is the universal data access mock pattern across all non-route tests.

### conftest.py Fixtures and Scope

Two conftest files exist:

**`tests/conftest.py`** (function-scoped by default):
- `app` — FastAPI instance with manually set `app.state`, no lifespan
- `client` — `httpx.AsyncClient` bound to bare app
- `scheduling_app` — Generator fixture that sets `AUTH_DEV_MODE=true`, overrides `get_session` dependency, yields `(app, mock_session)`
- `scheduling_client` — `AsyncClient` bound to auth-bypassed app with `AUTH_HEADERS` preset; yields `(client, mock_session)`

**`tests/golden_traces/conftest.py`** (function-scoped):
- `otel_provider` — `TracerProvider` + `InMemorySpanExporter`; patches three module-level tracer singletons
- `mock_business` — `MagicMock` with `guid`
- `mock_booking_engine` — `BookingEngine` with stubbed internal helpers and `session.add()` capture
- `golden_snapshot` — factory fixture returning a compare callable for JSON snapshot assertion
- `snapshot_update` — boolean for `--snapshot-update` flag

**Notable:** `scheduling_app` is a `Generator` (not `AsyncGenerator`) fixture, using `os.environ` mutation with teardown. The `get_settings.cache_clear()` call is present in setup and teardown to avoid cached settings leakage between tests.

### Test Data Directories and Fixture Files

- `tests/golden_traces/snapshots/booking_happy_path.json` — the only golden snapshot file (OTel span tree)
- Constants are defined at module level in each test file (e.g., `OFFICE_PHONE`, `START_DT`, `SYNTHETIC_APPT_ID`)
- No external fixture files (YAML, CSV, SQL) are used

### Parametrize Patterns

`@pytest.mark.parametrize` is used in 2 files:
- `test_model_layer.py` — `TestPhoneValidation` parametrizes 8 request models against invalid phone input
- `test_status_taxonomy.py` — parametrizes over status sets (3 uses)

Parametrization is used for validation boundary testing, not for business logic variations.

### Async Test Pattern

All async tests use `pytest-asyncio` with `asyncio_mode = "auto"` (configured in `pyproject.toml`). Tests are either undecorated (relying on auto-mode) or decorated with `@pytest.mark.asyncio`. The conftest uses `AsyncGenerator` for async fixtures.

---

## Test Structure Summary

### Overall Test Distribution

| Test File | Focus Layer | Test Count (approx) |
|-----------|-------------|---------------------|
| `test_booking_engine.py` | Unit — BookingEngine | ~25 tests |
| `test_availability_engine.py` | Unit — AvailabilityEngine | ~15 tests |
| `test_model_layer.py` | Unit — Pydantic models | ~25 tests |
| `test_write_ops.py` | Unit — write_ops.py | ~15 tests |
| `test_status_taxonomy.py` | Unit — constants.py | ~12 tests |
| `test_offer_resolution.py` | Unit — offer_resolution.py | ~15 tests |
| `test_ghl_sync.py` | Unit — ghl_sync.py | ~8 tests |
| `test_per_offer_duration.py` | Unit — duration resolution chain | ~7 tests |
| `test_routes_appointments.py` | Integration — HTTP routes | ~15 tests |
| `test_routes_businesses.py` | Integration — HTTP routes | ~8 tests |
| `test_health.py` | Integration — health endpoints | ~5 tests |
| `test_proxy.py` | Integration — proxy contract | ~5 tests |
| `test_sprint2_instrumentation.py` | Integration — OTel spans | ~8 tests |
| `golden_traces/test_golden_traces.py` | Snapshot — OTel booking trace | ~1 test |

### Most Heavily Tested Areas

1. **BookingEngine** — booking, cancel, reschedule paths; idempotency variants; TOCTOU defense; result type distinctions; rollback verification
2. **Pydantic model layer** — 20 response models checked for field descriptions; phone validation; request constraints; OpenAPI schema generation
3. **AvailabilityEngine** — slot generation, buffer application, date range clamping, overlap detection, timezone resolution, employee filtering, datetime parsing (SCAR-005)
4. **Status taxonomy / write_ops** — enum completeness, transition validation, terminal state enforcement

### Test Organization (Unit vs Integration vs Snapshot)

- **Unit tests** (~70% of test volume): Directly instantiate domain classes (`BookingEngine`, `AvailabilityEngine`, model classes, `SchedulingConfig`) with mocked sessions. All use `AsyncMock` for the DB layer.
- **Integration tests** (~25% of test volume): Use the full FastAPI app via `httpx.AsyncClient` with auth bypass and mocked session. Verify HTTP contract: status codes, envelope structure, error codes.
- **Snapshot tests** (~5%): One golden trace snapshot (`tests/golden_traces/snapshots/booking_happy_path.json`) captures the OTel span tree from the full booking path.

### Test Runner Configuration

From `pyproject.toml [tool.pytest.ini_options]`:
- `asyncio_mode = "auto"` — all coroutines are auto-wrapped as async tests
- `asyncio_default_fixture_loop_scope = "function"` — one event loop per test
- `testpaths = ["tests"]`
- `--import-mode=importlib`
- `-v` verbose
- `--tb=short`

Coverage is configured (`[tool.coverage.run]`) with `branch = true` and `source = ["src/autom8_scheduling"]`.

### CI Test Configuration

From `.github/workflows/test.yml`:
- Delegated to `autom8y/autom8y-workflows/.github/workflows/satellite-ci-reusable.yml@main` (reusable workflow, external)
- `coverage_threshold: 20` — very low bar; explicitly accepted
- `convention_check: true` with filter `sprint2_instrumentation` — only the instrumentation tests are checked for OTel convention compliance
- Integration tests run only on push to main (`run_integration: ${{ github.event_name == 'push' }}`)
- Second job: OpenAPI spec drift check (`scripts/generate_openapi.py --check`)

### No Test Data or Database Fixtures

No SQL fixture files, seed scripts, or database migrations are run during tests. All data access is mocked at the SQLAlchemy `session.execute()` level.

---

## Knowledge Gaps

- The full content of `test_offer_resolution.py` was partially read (token limit at line 60 of ~400); test count estimate is based on module structure pattern. The resolution chain cases are described in the docstring.
- The full content of `test_routes_appointments.py` was partially read (exceeded token limit); route test count is estimated from class structure visible in first 80 lines.
- `test_sprint2_instrumentation.py` was partially read (first 40 lines only); OTel attribute assertions and span hierarchy coverage are not fully documented.
- `test_write_ops.py` was partially read (first 60 lines); full coverage of `emit_scheduling_audit_event` and `dispatch_scheduling_notifications` tests is not verified.
- `tests/golden_traces/test_golden_traces.py` was not read; the number of snapshot scenarios beyond `booking_happy_path` is unknown.
- The reusable CI workflow (`satellite-ci-reusable.yml`) is external and not available for reading; integration test mechanics beyond what is configured here are opaque.
- No actual coverage report is available; the 20% threshold confirms gaps exist but exact module-level coverage percentages are unknown.
