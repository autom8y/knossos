---
domain: test-coverage
generated_at: "2026-03-25T12:09:30Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "9934462"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Packages and Modules Without Test Coverage

The coverage configuration in `pyproject.toml` explicitly omits two areas from measurement:

```
[tool.coverage.run]
omit = [
    "src/autom8_sms/debug.py",
    "src/autom8_sms/console/*",
]
```

Despite that omission, `test_console_fork_replay.py` and `test_otel_emission.py` exercise the console package in practice. The omission reflects intentional scoping, not an oversight.

**Source modules with NO direct test file:**

| Module | Criticality | Notes |
|--------|-------------|-------|
| `src/autom8_sms/handlers/appointment_reminder.py` | High | Lambda entry point for reminder handler. Only tested indirectly through `tests/integration/test_reminder_e2e.py`. The handler's `lambda_handler` and `instrument_lambda` decoration path have no unit-level test. |
| `src/autom8_sms/debug.py` | Low | Dev-only CLI debug tool. Explicitly omitted from coverage. Zero tests. Acceptable. |
| `src/autom8_sms/console/_instrument.py` | Medium | OTel instrumentation for console session. Covered only indirectly by `test_otel_emission.py`. |
| `src/autom8_sms/console/_session.py` | Medium | Console session management. Not directly unit-tested; `test_console_fork_replay.py` hits adjacent modules. |
| `src/autom8_sms/console/_stubs.py` | Low | Test-time stubs -- no production runtime risk. |
| `src/autom8_sms/console/__main__.py` | Low | CLI entry point. No unit test. Typical for entry points. |
| `src/autom8_sms/console/_renderer.py` | Low | Partially tested via `test_console_fork_replay.py` (TerminalRenderer). |

**Source modules with good unit test coverage:**

| Module | Test File(s) |
|--------|-------------|
| `src/autom8_sms/clients/twilio.py` | `test_twilio_client.py` (30+ tests), `test_orchestrator_mock_transport.py` |
| `src/autom8_sms/clients/data_service.py` | `test_data_service.py`, `test_circuit_breaker.py`, `test_defect_remediation.py` |
| `src/autom8_sms/clients/models.py` | `test_envelope_models.py` |
| `src/autom8_sms/config.py` | `test_config_migration.py` |
| `src/autom8_sms/models/conversation.py` | `test_models.py`, `test_ws3_instrumentation.py` |
| `src/autom8_sms/models/twilio.py` | `test_twilio_client.py` |
| `src/autom8_sms/handlers/client_lead.py` | `test_telemetry.py`, `test_correlation_id.py`, `test_pilot_config.py` |
| `src/autom8_sms/services/orchestrator.py` | `test_direction_change.py`, `test_tool_dispatch.py`, `test_pilot_gate.py`, `test_sdr_prompt_v2.py`, `test_scheduled_prompt.py`, `test_ws3_instrumentation.py`, integration tests |
| `src/autom8_sms/services/reminder/orchestrator.py` | `test_reminder.py` (extensive) |
| `src/autom8_sms/services/reminder/templates.py` | `test_reminder.py` |
| `src/autom8_sms/services/reminder/config.py` | `test_reminder.py` |
| `src/autom8_sms/tools/validation.py` | `test_tool_validation.py`, `test_adversarial_tool_use.py` |
| `src/autom8_sms/tools/formatting.py` | `test_tool_formatting.py` |
| `src/autom8_sms/tools/models.py` | `test_tool_dispatch.py` |
| `src/autom8_sms/tools/definitions.py` | `test_tool_dispatch.py`, `test_pilot_gate.py` |
| `src/autom8_sms/prompts/sdr_prompt.py` | `test_sdr_prompt_v2.py`, `test_ws3_instrumentation.py` |
| `src/autom8_sms/prompts/scheduled_prompt.py` | `test_scheduled_prompt.py`, `test_ws3_instrumentation.py` |
| `src/autom8_sms/narrative/rules.py` | `test_narrative_plugin.py` (23 test functions) |
| `src/autom8_sms/console/_fork.py` | `test_console_fork_replay.py` |
| `src/autom8_sms/console/_replay.py` | `test_console_fork_replay.py` |
| `src/autom8_sms/console/_models.py` | `test_otel_emission.py`, `test_console_fork_replay.py` |
| `src/autom8_sms/console/_otel.py` | `test_otel_emission.py` |

### Critical Path Coverage Assessment

**Lambda handler -- client_lead (`handlers/client_lead.py`):**
- Tested: Tracer initialization, `_tracer` module-level, `_telemetry_ctx`, `_add_conversation_id` processor, `conversation_id_var` ContextVar, `lambda_handler` span creation, pilot phone parsing (`_get_pilot_phones`).
- NOT tested: `process_pending_conversations()` function body, full polling-loop error paths, async error handling for data service failures in the main loop, `asyncio.run()` interaction with the Lambda runtime.

**Lambda handler -- appointment_reminder (`handlers/appointment_reminder.py`):**
- Tested: Indirectly via `tests/integration/test_reminder_e2e.py` (orchestrator-level).
- NOT tested: `lambda_handler` itself, `instrument_lambda` decorator interaction, `emit_success_timestamp` call path, cold-start logging config for this handler.

**ConversationOrchestrator (`services/orchestrator.py`):**
- Well tested: `_generate_response()`, `_resolve_tools()`, dispatch loop, token accumulation, iteration limit, calendar-link fallback, pilot gating, tool routing by lead status, error handling for `DataServiceUnavailableError`.
- NOT tested: `process_conversation()` full async error paths (data-service outage when fetching context, Twilio failure modes when logging), `from_env()` factory, `close()` teardown.

**Console package:**
- `_fork.py`, `_replay.py`, `_models.py`, `_otel.py`, `_renderer.py` have direct unit tests via `test_console_fork_replay.py` and `test_otel_emission.py`.
- `_session.py`, `_instrument.py`, `__main__.py` are untested but low-production-risk (dev tool).

### Prioritized Coverage Gap List (by risk)

1. **HIGH -- `process_pending_conversations()` in `handlers/client_lead.py`**: Core polling loop; outage here silently breaks all inbound SMS handling. Tests exist for supporting components but not the loop itself.
2. **HIGH -- `handlers/appointment_reminder.lambda_handler`**: The reminder Lambda entry point is not directly tested; CI cannot detect a regression in the `@instrument_lambda`-decorated handler.
3. **MEDIUM -- Error paths in `ConversationOrchestrator.process_conversation()`**: Success path is well-covered by integration tests; failure paths (data-service down when fetching context, Twilio failure) are not systematically tested.
4. **MEDIUM -- `console/_session.py`**: Interactive session management (console product); growing surface area without direct tests.
5. **LOW -- `debug.py`**: Dev tool only. No production risk.

### Coverage Infrastructure

`pyproject.toml` declares `pytest-cov>=4.1.0` in dev dependencies and configures branch coverage:

```toml
[tool.coverage.run]
source = ["src/autom8_sms"]
branch = true
```

A `.coverage` file is present in the repo root, confirming `pytest --cov` has been run. No CI artifact surfaces the exact coverage percentage. The `just test-cov` command produces an HTML report locally.

**Negative test coverage:** Present and consistent. The following test files explicitly test error-path and rejection behavior:
- `test_twilio_client.py`: 4xx, 5xx, timeout, connect error cases.
- `test_tool_validation.py`: past-date rejection, inverted ranges, malformed formats, business-hours mismatches.
- `test_adversarial_tool_use.py`: prompt injection bypass attempts, security validation bypass scenarios.
- `test_data_service.py`: `get_business` 404, null data, malformed-response degradation.
- `test_defect_remediation.py`: Regression tests for DEF-2 (malformed model validation) and DEF-3 (key access guard).
- `test_pilot_config.py`: Malformed phone rejection, all-malformed returns None.

---

## Testing Conventions

### Test Function Naming

All test functions follow one of two patterns:

1. **Class-based grouping** (dominant pattern): `class Test{Area}:` containing `def test_{specific_behavior}`. Examples:
   - `class TestTwilioSendSmsErrors: def test_4xx_raises_twilio_error`
   - `class TestReminderOrchestratorMarkFailure: def test_mark_failure_still_counts_as_sent`
   - `class TestDispatchLoopIterationLimit: def test_max_iterations_fallback`
2. **Module-level functions** (minority): `async def test_{behavior}`. Used in `test_direction_change.py`.

Class names follow `Test{Subject}{Scenario}` or `Test{Subject}` patterns. Method names follow `test_{action}_{condition_or_expected_result}` or simply `test_{condition}`.

### Test Patterns

**Async tests:** `pytest-asyncio` with `asyncio_mode = "auto"` (set in `pyproject.toml`). All async tests use `async def test_...()` without needing `@pytest.mark.asyncio` (though it appears on some older tests). The `async def` pattern is the standard.

**Parametrized tests:** `@pytest.mark.parametrize` used in `test_narrative_plugin.py` (4 uses). The codebase otherwise uses repeated explicit test methods instead of parametrize. This is a consistent convention across most test files.

**Fixtures:** Defined at multiple levels:
- Global autouse: `_clean_sms_env` in `tests/conftest.py` -- clears env vars and resets config singleton before every test. References `SCAR-011` and `SCAR-014` in comments.
- Module-level fixtures: `@pytest.fixture` inside test classes (e.g., `client` fixtures in `test_data_service.py`).
- Cross-module fixtures: `twilio_client`, `mock_twilio` in `tests/conftest.py`; `mock_claude`, `mock_data_client` in `tests/integration/conftest.py`.
- SDK-provided fixtures: Auto-discovered via pytest11 entry points from `autom8y-log[testing]`, `autom8y-http[testing]`, `autom8y-core[testing]`, `autom8y-config[testing]`, `autom8y-sms-test[fixtures]`.

**Mocking patterns:**
- `AsyncMock` and `MagicMock` from `unittest.mock` -- standard pattern for all external clients.
- `MockAIClient` from `autom8y_ai.testing` -- specialized mock for Claude responses, supports `responses=[...]` sequence mode.
- `MockTwilioTransport` from `autom8y_sms_test` -- transport-level HTTP interception (ADR-009 pattern). Installed via `install_mock_transport()`.
- `StubDataSchedulingClient` from `autom8y_interop.data.stubs` -- pre-built stub for scheduling data client.
- `monkeypatch.setenv` / `monkeypatch.delenv` for environment variable control (preferred over `patch.dict(os.environ, ...)`).

**Assertion patterns:**
- Direct attribute access: `assert result.success is True`.
- `pytest.raises(ExceptionType)` for error paths, often with `exc_info.value.code` inspection.
- Mock call assertions: `.assert_called_once()`, `.assert_called_once_with(...)`, `.call_count`.
- String content assertions: `assert "booked" in result.response_text.lower()`.
- Helper assertions: `assert_sms_sent(mock_twilio, to=..., body_contains=...)` from `autom8y_sms_test`.

**Test skip patterns:** `@pytest.mark.skip(reason="C-N aspirational: ...")` in `test_wsc_span_completeness.py` documents future implementation gaps as structured test stubs. Markers defined in `pyproject.toml` (`integration`, `db_integration`, `db_mysql`, `e2e`, `slow`) but the `integration` tests in `tests/integration/` do not use `@pytest.mark.integration`. Integration separation is structural (subdirectory), not marker-based.

### Test Data Patterns

**Factory hierarchy:**
- `tests/helpers.py`: Base domain model factories -- `make_lead()`, `make_business()`, `make_address()`, `make_context()`. Returns domain model instances with sensible defaults. Documented as stable; other modules compose it.
- `tests/factories/scheduling.py`: Scheduling-specific factories -- `make_hours()`, `make_appointment()`, `make_employee()`, `make_business_offer()`, `make_weekly_hours()`, `make_booking_flow_responses()`, `make_scheduling_context()`. Composes `helpers.make_context()`.
- `tests/integration/helpers.py`: Integration-level helpers -- `make_test_config()`, `make_test_availability()`, `make_test_data_client()`.

**No `testdata/` or `fixtures/` directories.** All test data is programmatic (factory functions), not file-based.

**Test environment management:** The `_clean_sms_env` autouse fixture clears all `SMS_SERVICE_`, `TWILIO_`, `AUTOM8Y_`, `DATA_SERVICE_`, `AUTH_BASE_` prefixed env vars and resets the config singleton (`config_module._config = None`) before every test. This is the universal isolation mechanism.

### SCAR References in Tests

Defensive patterns born from past bugs are annotated directly in test code:
- `SCAR-011`: Config singleton reset between tests (enforced by autouse fixture).
- `SCAR-014`: Structlog logger cache disabled (`cache_logger_on_first_use=False`).
- `DEF-01`, `DEF-2`, `DEF-3`: Regression tests named after defect IDs in `test_defect_remediation.py` and `test_tool_validation.py`.

---

## Test Structure Summary

### Distribution

**Total test files:** 30+ (including helpers and conftest files)
**Unit test files:** 26+ (in `tests/` root)
**Integration test files:** 4 (in `tests/integration/`)
**Helper/fixture files:** 6 (`conftest.py` x2, `helpers.py` x2, `factories/scheduling.py`, `integration/helpers.py`)

**Breakdown by domain:**

| Domain | Files | Notable tests |
|--------|-------|---------------|
| Clients (Twilio, Data, Models) | 4 | `test_twilio_client.py` (30+ tests), `test_data_service.py` (14 tests), `test_circuit_breaker.py`, `test_envelope_models.py` |
| Services -- Orchestrator | 6 | `test_tool_dispatch.py` (25+ tests), `test_direction_change.py`, `test_pilot_gate.py`, `test_orchestrator_mock_transport.py`, `test_sdr_prompt_v2.py`, `test_scheduled_prompt.py` |
| Services -- Reminder | 1 | `test_reminder.py` (30+ tests across templates, orchestrator, config) |
| Tools | 3 | `test_tool_validation.py`, `test_tool_formatting.py`, `test_adversarial_tool_use.py` |
| Models | 2 | `test_models.py`, `test_ws3_instrumentation.py` |
| Config | 2 | `test_pilot_config.py`, `test_config_migration.py` |
| Handlers | 2 | `test_telemetry.py`, `test_correlation_id.py` |
| Console | 2 | `test_console_fork_replay.py`, `test_otel_emission.py` |
| Narrative | 1 | `test_narrative_plugin.py` (23 tests) |
| Infrastructure | 2 | `test_dockerfile.py`, `test_defect_remediation.py` |
| Integration | 4 | `test_booking_flow.py`, `test_scheduling_e2e.py`, `test_reminder_e2e.py`, `test_tool_use_flow.py` |

### Heavily Tested Areas

1. **ConversationOrchestrator / tool dispatch loop**: 6 unit test files and 4 integration test files all exercise the orchestrator. The dispatch loop (FR-4), iteration limit (FR-7), token accumulation (FR-8), pilot gating, and tool routing by lead status (WS2-B) are extensively covered.
2. **TwilioClient**: 30+ tests covering success, error (4xx/5xx/timeout/connect), response parsing, request capture, magic numbers, and sequence mode.
3. **ReminderOrchestrator + templates**: 30+ tests covering all batch processing scenarios -- empty, single, multiple, mixed, dry-run, mark failure, Twilio failure, opt-out error, context manager.
4. **Tool validation (SD-03/04/05)**: Temporal validation, business hours, booking context allowlist, and adversarial bypass attempts all tested.
5. **SmsDataClient**: Composite context parsing, 409 idempotency, `get_business`, `get_recent_messages`, delegation, backward-compat alias.

### Integration vs Unit Distribution

Integration tests (in `tests/integration/`) are structural -- they are not marked with `@pytest.mark.integration`. They require the same mock infrastructure as unit tests (no live services). They differ in scope: they compose multiple components (MockAIClient + MockTwilioTransport + mock data client) to validate the full `process_conversation()` or `ReminderOrchestrator.run()` path.

No live-service tests (L3 `db_integration`, L4 `e2e`) are present in the repo. The markers are defined but unused.

### Test Invocation

**Standard run:**
```
uv run pytest
```

**With coverage:**
```
uv run pytest --cov=autom8_sms --cov-report=html
```
or
```
just test-cov
```

**Single test:**
```
uv run pytest tests/test_tool_dispatch.py
```

**Coverage source:** `src/autom8_sms`, branch coverage enabled, `src/autom8_sms/debug.py` and `src/autom8_sms/console/*` omitted.

### Test Package Naming Patterns

- Test files: `test_{module_name}.py` or `test_{feature_or_work_package}.py` (e.g., `test_ws3_instrumentation.py`, `test_defect_remediation.py`, `test_pilot_gate.py`).
- No mirror structure: all unit tests live flat in `tests/`, not in a mirrored `tests/autom8_sms/` tree.
- Integration tests in `tests/integration/` use `test_{flow_name}.py` naming.
- `tests/pythonpath = ["src", "tests"]` set in `pyproject.toml` enables `from helpers import ...` and `from factories.scheduling import ...` without package prefix.

---

## Knowledge Gaps

1. **Actual coverage percentage**: The `.coverage` binary file exists but was not parsed. The exact line/branch coverage percentage is not captured in this document. Running `just test-cov` would surface the HTML report.
2. **`tests/integration/test_reminder_e2e.py` and `test_tool_use_flow.py` content**: These files were not fully read. Their specific test scenarios and assertion patterns are not documented here.
3. **`test_adversarial_tool_use.py` full content**: Only partially read. The adversarial test scenarios beyond security validation constants are not fully cataloged.
4. **`src/autom8_sms/console/_session.py` and `_instrument.py` surface area**: These modules were not read; their test gap cannot be precisely quantified.
5. **Whether `tests/integration/` tests run in the default `pytest` invocation**: They appear to (no `@pytest.mark.integration` filter applied), but the CI configuration was not reviewed to confirm.
