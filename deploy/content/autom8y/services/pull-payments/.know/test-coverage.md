---
domain: test-coverage
generated_at: "2026-03-16T00:04:02Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

## Coverage Gaps

### Source modules vs. test coverage mapping

| Source Module | Test File(s) | Coverage Status |
|---|---|---|
| `handler.py` | None | **Not directly tested** |
| `config.py` | `test_config_security.py`, `test_config_extension.py`, `test_config_url_guard.py` | Well covered |
| `orchestrator.py` | `test_orchestrator.py`, `test_instrumentation_orchestrator.py` | Well covered |
| `models.py` | Used as fixtures, tested implicitly | No dedicated test file |
| `metrics.py` | No dedicated test file | **Not directly tested** |
| `staging.py` | `test_staging.py`, `test_instrumentation_staging.py`, `test_data_service_circuit_breaker.py` | Well covered |
| `replay.py` | `test_replay.py`, `test_replay_integration.py`, `test_instrumentation_orchestrator.py` | Well covered |
| `clients/data_service.py` | `test_data_service.py`, `test_data_service_circuit_breaker.py`, `test_instrumentation_data_service.py` | Well covered |
| `clients/models.py` | `test_client_models.py` | Well covered |

### Untested modules

**`handler.py`** (Lambda entry point, critical):
- `lambda_handler` exercises `asyncio.run(sync_payments(...))`, `emit_success_timestamp`, and `record_side_effect`. No test exercises this function at all. Any regression in the handler's response-building logic, exception re-raise behavior, or field mapping from `SyncResult` to the JSON body is invisible to CI.
- **Criticality: High** — this is the actual Lambda entry point; regressions here would cause silent deploy failures.

**`metrics.py`** (`emit_replay_metrics`):
- No test file covers `emit_replay_metrics`. The function calls `emit_business_metrics` (SDK) and `record_side_effect` (telemetry). SDK errors during emission are swallowed at the SDK level, but the span instrumentation path and metric list correctness are untested.
- **Criticality: Medium** — Grafana alerts depend on these metrics being emitted correctly; no test validates the metric payload shape.

**`models.py`** (data classes):
- `SyncResult.to_dict()`, `ReplayResult.to_dict()`, and `PaymentRecord.to_dict()` have no dedicated tests. These are exercised indirectly through orchestrator tests but no assertion verifies their serialization contracts.
- **Criticality: Low-Medium** — used in handler JSON body; a typo or missing field would go undetected until production.

### Weak coverage areas

**`orchestrator.py` — `sync_payments_for_range` dry-run path**: The `dry_run=True` branch exists (with counting-without-writing logic) but the unit tests focus on the live write path. Dry-run counting correctness for both invoices and refunds is not explicitly asserted.

**`staging.py` — `_move_key` on S3 exception**: `_move_key` has a `except Exception` catch that logs and swallows the error, but no test verifies this fallback behavior when the copy/delete fails mid-operation.

**`staging.py` — `migrate_tmp_to_s3` on partial failure**: The function loops over files and catches exceptions per-file. The case where one file fails (S3 error) while others succeed is not tested.

**`orchestrator.py` — `sync_refunds_for_range` charge lookup failure / invoice lookup failure**: The `except (StripeAPIError, StripeNotFoundError)` branches for charge and invoice fetching exist but no test asserts the `continue` (skip refund) behavior under these error conditions.

**Negative tests**: Tests for refund categorization `categorize_refund` cover all four cases (direct charge, exact, partial, proportional). However, edge cases like zero-amount line items in the proportional path, or an invoice with a single line item where the refund exceeds all line items, are untested.

### Prioritized gap list

1. `handler.py` — add at minimum a unit test for the happy-path response shape and the exception-re-raise behavior
2. `metrics.py` — add a test asserting `emit_replay_metrics` calls `emit_business_metrics` with the correct metric payload
3. `models.py` serialization — add tests for `to_dict()` on `SyncResult`, `ReplayResult`, and `PaymentRecord`
4. `staging.py` `_move_key` error path — add a test that S3 `copy_object` failure is absorbed gracefully
5. Refund sync error paths — add tests for `sync_refunds_for_range` when charge or invoice lookup fails

---

## Testing Conventions

### Test function naming

All test functions follow `test_<behavior_under_test>` with descriptive names that encode the scenario:
- `test_stages_to_s3_with_metadata`
- `test_fallback_to_tmp_on_s3_error`
- `test_dead_letter_at_max_attempts`
- `test_half_open_allows_one_probe`

### Test class naming

Tests are organized into classes using the `Test<Unit><Scenario>` or `Test<Class>` pattern:
- `TestStageBatchS3`, `TestListStagedKeys`, `TestLoadStagedBatch` — per-function grouping
- `TestReplayHappyPath`, `TestReplayTimeBudget`, `TestReplayCircuitBreakerOpen` — per-scenario grouping
- `TestCircuitBreakerBehavior`, `TestDataServiceClient` — per-component grouping

### Assertion patterns

- Direct attribute assertions: `assert result.replayed_count == 1`
- Mock call assertions: `mock_delete.assert_called_once_with("active/1000_abcd.json")`
- `isinstance` checks: `assert isinstance(result, BatchResult)`
- JSON deserialization followed by key assertions: `data = json.loads(...); assert data["payments"] == payments`
- No `assert_called_once()` without argument verification where the argument matters

### Fixture patterns

**conftest.py** defines shared fixtures:
- `mock_settings` — sets all env vars via `monkeypatch.setenv`, creates a `Settings()` instance, clears cache before/after
- `settings_override` — patches `get_settings` at all import locations (config, orchestrator, data_service)
- Domain object mocks: `mock_invoice`, `mock_subscription`, `mock_vertical_match`, `mock_product_match`, `mock_stripe_client`, `mock_http_client`
- Response mocks: `mock_business_response`, `mock_batch_success_response`, `mock_batch_partial_response`

Local fixtures also appear within test classes:
- `mock_data_client` in `test_replay.py` — creates an `AsyncMock` with `circuit_breaker_state = "closed"`
- `_mock_env` and `_redirect_staged_dir` in `test_data_service_circuit_breaker.py` — autouse fixtures for environment isolation

### Test data management

- No `fixtures/` or `testdata/` directories — all test data is constructed inline using `MagicMock` / `AsyncMock`, JSON literals, or `pytest.MonkeyPatch`
- S3 content is represented as `json.dumps(...).encode()` in `MagicMock(read=lambda: body_content)` patterns
- `tmp_path` is used via `monkeypatch.setattr(staging, "STAGED_DIR", tmp_path / "staged_batches")` to redirect filesystem writes

### Skip and mark patterns

- No `@pytest.mark.integration` on any test (the Justfile defines `-m "not integration"` for unit and `-m "integration"` for integration runs, but no test actually carries the marker — the integration test file is named `test_replay_integration.py` but no marker is applied)
- No `@pytest.mark.skip` or `@pytest.mark.xfail` found in any test file
- `asyncio_mode = "auto"` in `pyproject.toml` — async tests do not need `@pytest.mark.asyncio` decoration (some tests still carry it as legacy)

### Test environment management

- Settings isolation via `clear_settings_cache()` called in fixture setup and teardown
- `monkeypatch.setenv` used for all environment variable injection (never `os.environ` direct mutation)
- `monkeypatch.setattr` used for module-level globals (e.g., `staging._s3_client`, `staging.STAGED_DIR`)
- `patch.object(staging, "_get_bucket", ...)` used to isolate bucket configuration

### Instrumentation test pattern

Three dedicated instrumentation test files verify OpenTelemetry span emission:
- `test_instrumentation_staging.py`
- `test_instrumentation_orchestrator.py`
- `test_instrumentation_data_service.py`

These use a `convention_span_exporter` fixture (provided by `autom8y-telemetry[conventions]`) that captures spans in memory for assertion. Tests verify span names, event names, and `record_side_effect` calls.

---

## Test Structure Summary

### Overall distribution

| File | Test Functions | Test Classes | Focus |
|---|---|---|---|
| `test_orchestrator.py` | 33 | ~12 | orchestration logic, refund categorization, chunk_date_range |
| `test_staging.py` | 26 | 8 | S3 staging, /tmp fallback, legacy functions |
| `test_replay.py` | 23 | 12 | replay algorithm, helper functions |
| `test_data_service_circuit_breaker.py` | 16 | 1 | CB open/close, trip-and-degrade, state machine |
| `test_instrumentation_orchestrator.py` | 13 | 7 | span emission in orchestrator & replay |
| `test_replay_integration.py` | 12 | 11 | real moto S3 integration for replay module |
| `test_data_service.py` | 10 | 1 | HTTP methods, response parsing, context manager |
| `test_config_security.py` | 10 | 1 | secret redaction, caching, alias resolution |
| `test_client_models.py` | 9 | ~3 | Pydantic response model validation |
| `test_config_url_guard.py` | 5 | 1 | dev/prod URL guard validation |
| `test_config_extension.py` | 5 | 1 | Lambda extension secret resolution |
| `test_instrumentation_staging.py` | 3 | 3 | span/side-effect emission in staging |
| `test_instrumentation_data_service.py` | 3 | 3 | span/side-effect emission in data service |
| **Total** | **168** | **67** | |

### Most heavily tested areas

1. **Staging module** — 29 tests across `test_staging.py` + `test_instrumentation_staging.py` + `test_data_service_circuit_breaker.py`. The S3 key format, metadata envelope, /tmp fallback, and move operations are all exercised.

2. **Replay algorithm** — 35 tests across `test_replay.py` + `test_replay_integration.py` + `test_instrumentation_orchestrator.py`. Every branch of the replay loop (time budget, circuit breaker states, dead-letter, partial 207, all-duplicate, corrupt, empty) has a dedicated test class.

3. **Orchestrator / refund categorization** — 33 tests in `test_orchestrator.py`. `categorize_refund` has tests for all 4 cases (direct charge, exact match, partial, proportional). `chunk_date_range` has arithmetic coverage.

### Unit vs integration test distinction

The Justfile differentiates unit (`-m "not integration"`) from integration (`-m "integration"`) runs, but **no test carries `@pytest.mark.integration`**. `test_replay_integration.py` uses `moto[s3]` to create a real in-memory S3 bucket, making it effectively an integration test for the replay+staging pipeline, but it runs in the default `just test` invocation alongside unit tests.

### Test organization patterns

- Tests are grouped by component (one test file per source file), with instrumentation variants as separate files
- Test classes group related scenarios for the same function: e.g., `TestStageBatchS3` vs `TestStageBatchTmpFallback`
- Helper functions are extracted per-file (e.g., `_make_s3_mock`, `_make_charge`, `_add_empty_refunds`) rather than put in conftest.py
- No pytest plugins beyond `pytest-asyncio` and `pytest-cov` are used

### How tests are run

```
just test         # All tests: uv run pytest tests/ -v
just test-unit    # -m "not integration" (no marker applied, runs all)
just test-int     # -m "integration" (no marker applied, runs nothing)
just test-cov     # --cov=src --cov-report=term-missing --cov-report=html
```

Coverage reporting uses `pytest-cov` with `--cov-report=html` producing an `htmlcov/` directory.

### conftest.py patterns

`tests/conftest.py`:
- One conftest, no nested conftests
- Provides: settings fixtures (`mock_settings`, `settings_override`), domain object mocks (`mock_invoice`, `mock_subscription`, `mock_stripe_client`, etc.), and response payload fixtures (`mock_business_response`, `mock_batch_success_response`, `mock_batch_partial_response`)
- Cache hygiene: `clear_settings_cache()` called before and after `mock_settings` to prevent test pollution from `lru_cache`

---

## Knowledge Gaps

- Coverage percentages: `pytest-cov` was not executed during this audit; the gap list is based on structural analysis (test file presence vs. source module), not measured line coverage numbers
- `scripts/test_single_write.py` was not examined in detail — it appears to be a developer script, not a pytest test, and is excluded from the `tests/` testpath
- The `autom8y-telemetry[conventions]` `convention_span_exporter` fixture internals were not traced; instrumentation test assumptions about what that fixture captures are based on the test code, not the SDK source
