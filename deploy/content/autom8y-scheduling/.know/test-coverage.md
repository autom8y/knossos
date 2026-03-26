---
domain: test-coverage
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

# Codebase Test Coverage

## Test Organization

The test suite lives under `tests/` with a two-level structure: a top-level flat directory for general tests and a `golden_traces/` subdirectory for snapshot-based OTel tests.

**Directory layout:**

```
tests/
  __init__.py
  conftest.py                        # top-level app + async HTTP client fixtures
  test_health.py                     # health/readiness endpoint tests
  test_proxy.py                      # scheduling reverse proxy contract tests
  test_sprint2_instrumentation.py    # OTel span attribute correctness tests
  test_write_ops.py                  # write path business logic unit tests
  golden_traces/
    __init__.py
    conftest.py                      # OTel provider fixture + golden snapshot factory
    serializer.py                    # JSON serialization helpers (not a test file)
    span_tree.py                     # span normalization utilities (not a test file)
    test_golden_traces.py            # golden snapshot regression test
    snapshots/
      booking_happy_path.json        # sole golden snapshot file
```

**Naming conventions:** All test files follow `test_*.py`. `python_files = "test_*.py"` set in `pyproject.toml`.

**Test style:** Class-based organization throughout. Every test file uses `class Test...` grouping with `async def test_...` methods. No standalone functions at module scope.

**Counts:** 5 test files, 21 test classes, 69 test functions, 2 conftest.py files, 2 utility modules in `golden_traces/`.

The structure does NOT mirror the source package layout — tests are organized by concern (health, proxy contract, instrumentation sprint, write-ops logic, golden traces).

## Test Infrastructure

**Test runner:** pytest, configured in `pyproject.toml`:

```
asyncio_mode = "auto"
asyncio_default_fixture_loop_scope = "function"
python_files = "test_*.py"
testpaths = ["tests"]
addopts = ["--import-mode=importlib", "--tb=short", "-v"]
```

`asyncio_mode = "auto"` means all `async def test_...` functions run as coroutines without explicit `@pytest.mark.asyncio`.

**Dev dependencies:**

| Package | Version | Role |
|---|---|---|
| `pytest` | >=7.4.0 | Test runner |
| `pytest-asyncio` | >=1.2,<2.0 | Async test support |
| `pytest-cov` | >=4.1.0 | Coverage plugin |
| `httpx` | >=0.25 | Async HTTP client for ASGI testing |
| `mypy` | >=1.7.0 | Static type checking |
| `ruff` | >=0.1.6 | Linter/formatter |
| `respx` | >=0.21.0 | httpx mock/intercept library |
| `autom8y-telemetry[conventions]` | latest | OTel testing conventions |

No factory library (e.g. `factory_boy`, `polyfactory`). Test objects constructed via `MagicMock()` and hand-coded attribute assignment.

**Coverage configuration:**

- `source = ["src/autom8_scheduling"]`, `branch = true`
- Excluded: `pragma: no cover`, `if TYPE_CHECKING:`, `@abstractmethod`, `raise NotImplementedError`, `if __name__ == .__main__.:`

**Coverage threshold:** 20% (set in CI at `.github/workflows/test.yml` line 23). Intentionally low for migration state.

**CI workflow:** Triggers on push/PR to `main`. Delegates to reusable `satellite-ci-reusable.yml`. Integration tests only on push to main, not PRs. Convention check gated on `sprint2_instrumentation` test filter.

## Coverage Patterns

### Tested source modules

| Source Module | Test File(s) | Testing Approach |
|---|---|---|
| `scheduling/booking.py` | `test_sprint2_instrumentation.py`, `golden_traces/test_golden_traces.py` | Unit (mocked session + stubbed helpers), OTel span verification, golden snapshot |
| `scheduling/engine.py` | `test_sprint2_instrumentation.py` | Unit (mocked session), OTel span attribute verification |
| `scheduling/gcal_overlay.py` | `test_sprint2_instrumentation.py` | Unit (mocked gcal client), OTel span attribute verification |
| `scheduling/gcal_sync.py` | `test_sprint2_instrumentation.py` | Unit with `patch` on internal helpers |
| `scheduling/notifications/dispatch.py` | `test_sprint2_instrumentation.py` | Unit with `patch` on internal helpers and service classes |
| `scheduling/write_ops.py` | `test_write_ops.py` | Pure unit tests (no mocks needed) |
| `health.py` | `test_health.py` | Integration (ASGI client against live FastAPI app with mocked state) |
| `app.py` | `tests/conftest.py` (fixture) | Implicit — used to construct test application |

### Proxy contract (behavioral testing)

`test_proxy.py` tests the reverse proxy logic from `autom8y-data`, which is NOT in this repository. The test file replicates the proxy handler inline.

### Summary: 8 of 25 source modules have direct test coverage (32%)

## Test Gaps

Gaps ordered by risk:

### High-risk gaps (untested business logic)

**`services/appointment.py` — `AppointmentService`**: Status update logic calling `validate_status_transition`, `emit_scheduling_audit_event`, and `dispatch_scheduling_notifications` is untested. Primary business logic path.

**`api/routes/__init__.py` — 11 route handlers**: `check_availability`, `book_appointment`, `cancel_appointment`, `reschedule_appointment`, `get_appointments`, `update_scheduling_config`, `get_reminder_eligible`, `mark_reminder_sent`, `get_unsynced_appointments`, `reconcile_gcal_events`, `check_readiness` — none tested via ASGI client.

**`scheduling/reminder.py` — `ReminderEngine`**: Reminder eligibility logic untested.

**`scheduling/validation.py` — `check_scheduling_readiness`**: Readiness check untested.

### Medium-risk gaps

**`scheduling/notifications/templates.py`**: Template functions with format string logic untested.

**`scheduling/notifications/sms_service.py` and `email_service.py`**: Only exercised via `MagicMock` patches. Real service logic including `_hash_pii()` has no direct tests.

**`scheduling/gcal_sync.py` — cancel and reschedule paths**: Only `dispatch_gcal_booking_sync` is tested. `dispatch_gcal_cancel_sync` and `dispatch_gcal_reschedule_sync` untested.

### Lower-risk gaps

**`scheduling/constants.py` — `normalize_status()`**: No direct unit test.

**`models/` subpackage**: All SQLModel/Pydantic model definitions untested.

**`config.py`**: Application settings loading untested.

**`services/base.py`**: `BaseService` and exception hierarchy untested.

## Fixture and Mock Patterns

### conftest.py hierarchy

**Top-level** (`tests/conftest.py`):
- `app` fixture: Creates `FastAPI` via `create_app()`, sets `app.state.db_engine = None` and `app.state.gcal_client = None` to bypass lifespan.
- `client` fixture: `httpx.AsyncClient` with `httpx.ASGITransport(app=app)`.

**Golden traces** (`tests/golden_traces/conftest.py`):
- `otel_provider` fixture: Fresh `TracerProvider` + `InMemorySpanExporter` per test; patches three module-level `_tracer` singletons via direct monkeypatching.
- `mock_booking_engine` fixture: `BookingEngine` with `AsyncMock` session; four internal helpers stubbed via direct attribute assignment.
- `golden_snapshot` fixture: Factory returning `_compare(tree, name)` callable. Reads/writes JSON snapshots. Supports `--snapshot-update` flag.

### Mock strategy

Dominant pattern: **direct method stubbing via attribute assignment** on class instance:
```python
engine._find_by_idempotency_key = fake_find_idempotency
engine._resolve_business = fake_resolve_business
```

`mock.patch` as context manager used in `test_sprint2_instrumentation.py` for patching internal functions.

`respx.mock` used exclusively in `test_proxy.py` for httpx-level HTTP interception.

`AsyncMock()` for database session mocks. `MagicMock()` for non-async objects.

### No factory libraries

No `factory_boy`, `polyfactory`, or similar. All test data is literal dict construction or `MagicMock()` with attribute assignment.

### Shared constants (golden traces)

Deterministic test constants at module scope for snapshot stability: `TEST_OFFICE_PHONE`, `TEST_LEAD_PHONE`, `TEST_START_DATETIME`, `TEST_END_DATETIME`, `TEST_IDEMPOTENCY_KEY`, `SYNTHETIC_APPOINTMENT_ID`.

## Test Quality Signals

### Positive signals

**Golden file / snapshot testing:** One snapshot at `tests/golden_traces/snapshots/booking_happy_path.json`. `NormalizedSpan` type is frozen/hashable. Volatile keys stripped before snapshotting.

**OTel span attribute correctness tests:** `test_sprint2_instrumentation.py` explicitly asserts typed constant keys are present AND old bare string keys are NOT present. Regression guard against key regressions.

**Side-effect event verification:** Tests verify `side_effect` span events carry correct system/operation/target attributes.

**Test class docstrings:** All test classes and methods have docstrings. Improves legibility.

**Integration label:** `@pytest.mark.integration` on `TestSchedulingGoldenTraces`. Aligns with CI's `run_integration` flag.

**Convention check gating:** CI runs convention check filtered to `sprint2_instrumentation` tests.

### Weaknesses / anti-patterns

**No parameterized tests:** Zero `@pytest.mark.parametrize` usage across all 69 test functions. Transition validation uses explicit loops instead.

**No property-based tests:** No `hypothesis` usage.

**Boilerplate duplication in sprint2 tests:** Each test manually re-imports, saves tracer, patches, and restores in `try/finally`. Pattern already solved by `golden_traces/conftest.py`'s `otel_provider` fixture.

**Very low coverage threshold:** CI coverage threshold 20%. Does not enforce meaningful coverage discipline.

**No `@pytest.mark.slow`:** No test speed segmentation exists.

**No integration tests for API routes:** The 11 route handlers have no ASGI-level integration tests despite the conftest `client` fixture being available.

**`test_proxy.py` tests a non-local behavioral replica:** The proxy behavior lives in `autom8y-data`, not this repository. Drift possible without detection.

## Knowledge Gaps

- Coverage percentage for measured modules not known without running `pytest --cov`; only the 20% threshold confirmed.
- No CI run output available to inspect passing/failing tests.
- `autom8y-telemetry[conventions]` convention check behavior is external; exact gate criteria not observable.
- `scheduling/validation.py`, `scheduling/reminder.py`, `services/appointment.py`, and all `models/` files were not read in detail; coverage gaps inferred from absence of test imports.
