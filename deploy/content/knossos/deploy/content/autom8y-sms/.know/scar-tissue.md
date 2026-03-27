---
domain: scar-tissue
generated_at: "2026-03-25T12:09:30Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "9934462"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

**Primary Language**: Python (pyproject.toml present, `src/autom8_sms/` package, pytest test suite)

**Project**: autom8y-sms -- AWS Lambda scheduled SMS conversation service using Claude for AI-generated responses, Twilio for delivery, and autom8y-interop for data access.

---

## Failure Catalog Completeness

This section catalogs all identified scars from git commit history, code markers, and named defect entries. The project uses several marker conventions: `DEF-N`, `SCAR-N`, `WP-*`, `SD-*`, `DEFECT-N`, `SEC-*`, and `REQ-*`.

---

### SCAR-011: Config Singleton Pollution Between Tests

**What failed**: The `get_config()` function in `src/autom8_sms/config.py` uses a module-level `_config: Config | None = None` singleton. Tests that mutated or loaded config in one test contaminated subsequent tests, causing failures that depended on test execution order.

**When**: First evidenced via the `SCAR-011` comment in conftest.py.

**How it was fixed**: An `autouse` fixture in `tests/conftest.py` (lines 41-56) explicitly resets `config_module._config = None` between every test.

**Marker today**: `SCAR-011` comment in `tests/conftest.py` line 45 and 52.

---

### SCAR-014: structlog Logger Caching Across Tests

**What failed**: structlog's `cache_logger_on_first_use` feature caused logger instances to be frozen after first use. When tests changed logging configuration, previously-cached loggers continued using the old configuration, producing false test results and making log-assertion tests flaky.

**When**: Discovered alongside or immediately after SCAR-011 (same `autouse` fixture).

**How it was fixed**: The same `_clean_sms_env` autouse fixture at `tests/conftest.py` line 55-56 calls `structlog.configure(cache_logger_on_first_use=False)` before every test.

**Marker today**: `SCAR-014` comment in `tests/conftest.py` lines 46 and 55.

---

### SCAR-022/DEF-009: uv export flag conflict in CI Docker build

**What failed**: `uv export --frozen` and `--no-sources` are mutually exclusive flags. CI Docker build broke when both were specified.

**When**: Commit `7360746` -- "fix(ci): replace --frozen with --no-sources in uv export (DEF-009/SCAR-022)"

**How it was fixed**: Corrected the uv export flags in CI configuration.

**Marker today**: Git log only (CI workflow files outside src scope).

---

### DEF-2: Unguarded model_validate on Malformed API Responses

**What failed**: `get_business()` in the data service client called `BusinessResponse.model_validate(response.json())` without a try/except, causing an unhandled `ValidationError` to propagate when the API returned a malformed or unexpected response shape.

**When**: Surfaced during the interop migration (commit `f0dfaa5`).

**How it was fixed**: Wrapped in a try/except `ValidationError` block in `src/autom8_sms/clients/data_service.py` lines 288-291, logging `contract_validation_failed` and returning `None` on failure.

**Marker today**: `DEF-2` in `tests/test_defect_remediation.py` module docstring and class names.

---

### DEF-3: Hard Key Access on next_appointment Dict

**What failed**: Context parsing in `get_conversation_context()` used `next_appt["start_datetime"]` (hard key access) on a dict that could exist without that key, causing `KeyError` crashes.

**When**: Same era as DEF-2 (interop migration, commit `f0dfaa5`).

**How it was fixed**: Changed to `next_appt.get("start_datetime")` at `src/autom8_sms/clients/data_service.py` line 231.

**Marker today**: `DEF-3` in `tests/test_defect_remediation.py`.

---

### DEFECT-01: Claude Could Book Without Prior Availability Check

**What failed**: The `validate_business_hours()` function originally allowed bookings to proceed when no `AvailabilityResponse` existed. Claude could call `book_appointment` without first calling `check_availability`, bypassing business hours validation entirely.

**When**: Identified during the SD-03/04/05 security review.

**How it was fixed**: Changed the guard at `src/autom8_sms/tools/validation.py` lines 254-264 to return `ValidationResult(valid=False)` -- rejecting the booking and telling Claude to call `check_availability` first.

**Marker today**: `DEFECT-01` comment in `validation.py` line 254.

---

### DEFECT-02: No Far-Future Cap on Booking Datetime

**What failed**: `validate_temporal_bounds_booking()` only checked that the booking datetime was not in the past. Claude could hallucinate a booking far into the future.

**When**: Same SD-05 security review sprint as DEFECT-01.

**How it was fixed**: Added a `MAX_LOOKFORWARD_DAYS` (14 days) cap check at `src/autom8_sms/tools/validation.py` lines 196-205.

**Marker today**: `DEFECT-02` comment in `validation.py` line 196.

---

### DEFECT-03: end_datetime Not Validated (OPEN)

**What failed**: `end_datetime` field in booking tool calls is not validated -- known open issue.

**When**: Documented in adversarial test suite.

**How it was fixed**: Still open; documented in `tests/test_adversarial_tool_use.py` line 452.

**Marker today**: Comment in adversarial test file.

---

### Scar: Direction Type Too Narrow (Pydantic Literal -> str)

**What failed**: `MessageRecord.direction` was typed as `Literal["inbound", "outbound"]`. The API returns `"outbound-scheduled"`, `"outbound-claude"`, and potentially other variants. Pydantic raised `ValidationError` on any non-literal direction value.

**When**: Commit `e8b7a53` -- "fix: relax message direction type to match API."

**How it was fixed**: Changed `direction` to `str` in `src/autom8_sms/models/conversation.py` line 24.

**Marker today**: Inline docstring on the field: `"Message direction (inbound, outbound, outbound-scheduled, etc.)"` -- the "etc." is the scar.

---

### Scar: Inbound Message Double-Logging

**What failed**: The orchestrator was calling the data service to log inbound messages on receipt, but inbound messages are already logged by the core server via the Twilio webhook. This caused duplicate records.

**When**: Commit `43c1ee0` -- "fix: remove inbound message logging (handled server-side)."

**How it was fixed**: Removed the inbound logging call from `src/autom8_sms/services/orchestrator.py`. Module-level docstring documents this.

**Marker today**: Module-level docstring note in `orchestrator.py`.

---

### Scar: Hardcoded America/Denver Timezone Fallback

**What failed**: `_get_business_timezone()` in the orchestrator fell back to `"America/Denver"` when `address.timezone` was `None`. This silently masked data quality issues and produced wrong availability windows for any business outside Mountain Time.

**When**: Commit `49c7ada` -- "fix(orchestrator): remove hardcoded America/Denver timezone fallback."

**How it was fixed**: Now raises `ValueError` when timezone is not configured at `src/autom8_sms/services/orchestrator.py`. The dispatch loop's `except Exception` handler catches this and returns a graceful degradation response.

**Marker today**: The raised `ValueError` path itself is the defensive fix.

---

### Scar: CI SHA Sourcing -- Deploying Untested Commits

**What failed**: The satellite dispatch workflow used the wrong commit SHA source, resulting in potential deployment of untested commits.

**When**: Commit `39fe7dd` -- "fix(ci): use correct SHA in satellite dispatch to prevent deploying untested commits."

**How it was fixed**: Changed the SHA sourcing logic in `.github/workflows/satellite-dispatch.yml`.

**Marker today**: No inline code marker; the commit message documents the failure mode.

---

### Scar: Config Singleton Pollution -- Env Var Standardization Migration

**What failed**: The service config used multiple competing env var name schemes simultaneously (`SMS_SERVICE_*`, `SMS_*`, `AUTOM8Y_*`), with `AliasChoices` bridging them. Tests that set one scheme's vars didn't see values when the singleton loaded using a different scheme.

**When**: Commit `625584d` -- "refactor(config): clean-break env var standardization."

**How it was fixed**: All legacy aliases and `_warn_legacy_env_vars` validator removed. A single canonical scheme established. Test env cleanup in `conftest.py` `_clean_sms_env` strips all three prefix families.

**Marker today**: The `_clean_sms_env` fixture's explicit prefix list at `tests/conftest.py` lines 48-50.

---

### Scar: Docker Build Non-Reproducibility (lock file / --frozen flag conflicts)

**What failed**: A sequence of CI failures across five commits shows a history of `uv` flag incompatibilities in the Dockerfile.

**When**: Commits `336a58a`, `c71a7bc`, `932f6bb`, `a8c7817`, `e87831b`, `10eebda` -- all in rapid succession.

**How it was fixed**: `test_dockerfile.py` (WP-LOCK) now asserts the exact `uv export` flags.

**Marker today**: `tests/test_dockerfile.py` -- `WP-LOCK` tests exhaustively guard the Dockerfile pattern.

---

### Scar: datetime.utcnow() Deprecation

**What failed**: Usage of `datetime.utcnow()` (deprecated since Python 3.12) produced deprecation warnings that failed CI assertions.

**When**: Commits `d5a9232`, `a8de029`, `6f2dc15`, `2b35933`.

**How it was fixed**: All usages replaced with `datetime.now(UTC)`. Guarded by ruff rule UP017.

**Marker today**: `from datetime import UTC` is now the import pattern throughout the codebase.

---

### Scar: Pydantic Forward Reference -- datetime behind TYPE_CHECKING

**What failed**: `datetime` hidden behind `TYPE_CHECKING` guard -- Pydantic v2 `model_rebuild` fails to resolve `datetime` annotation at class construction.

**When**: Commit `248f527` -- "fix(models): resolve Pydantic forward reference failures for datetime."

**How it was fixed**: `datetime` imported unconditionally with `# noqa: TC003` suppression.

**Marker today**: `src/autom8_sms/console/_models.py` line 11; `src/autom8_sms/models/conversation.py` line 3.

---

### Scar: WP-TEL-SMS -- Trace Context Not Propagated in Data Service Requests

**What failed**: W3C `traceparent` headers are not injected by the interop `DataMessageClient` in its outbound HTTP requests.

**When**: Documented at the time of interop migration as a deliberate known gap.

**How it was fixed**: Not fixed at the application layer -- documented as a known limitation. Reliance on infrastructure-level propagation (ADOT sidecar).

**Marker today**: `WP-TEL-SMS` comment in `data_service.py` module docstring.

---

### Scar: Pilot Mode Hardcoded Phones / PILOT_MODE_ENABLED Flag

**What failed**: The original pilot filter used a hardcoded `PILOT_OFFICE_PHONES` constant and a separate `PILOT_MODE_ENABLED` boolean flag in module scope. This made pilot configuration a code change rather than a config change.

**When**: Commits `a46a5b1`, `951a5ef` (Phase 4 pilot hardening).

**How it was fixed**: Replaced with `_get_pilot_phones()` function that reads env var at cold start. `tests/test_pilot_config.py` asserts old constants do not exist.

**Marker today**: `tests/test_pilot_config.py` test `test_old_constants_removed` is the regression guard.

---

### Scar: CRITICAL Prompt Guardrails -- Miscommunication with Scheduled Leads

**What failed**: The SDR Claude prompt did not have explicit guardrails preventing it from sending scheduling links to already-scheduled leads.

**When**: Evidenced by the `CRITICAL BEHAVIOR UPDATE` marker in the SDR prompt -- a production-discovered failure.

**How it was fixed**: Explicit `CRITICAL BEHAVIOR UPDATE` section added to `src/autom8_sms/prompts/sdr_prompt.py` lines 54-57. `CRITICAL INSTRUCTION` at line 149.

**Marker today**: `CRITICAL BEHAVIOR UPDATE` at `sdr_prompt.py` line 54; `CRITICAL INSTRUCTION` at line 149; `CRITICAL:` at `scheduled_prompt.py` line 58.

---

### Scar: genai Shim -- autom8y_telemetry.genai Module Missing

**What failed**: `autom8y_telemetry.genai` submodule does not exist in v0.5.2 -- `ImportError` in console module.

**When**: Commit `bbbfe26`.

**How it was fixed**: Local `_genai_attrs.py` shim created in console package.

**Marker today**: `src/autom8_sms/console/_genai_attrs.py` with explicit TODO for removal.

---

### Scar: DC-1 devx-types Not Published to CodeArtifact

**What failed**: `autom8y-devx-types` not published to CodeArtifact -- CI resolution fails with `--no-sources`.

**When**: Commits `c681233`, `f48f592`.

**How it was fixed**: Moved to optional dep; `pytest.importorskip` fallback in tests.

**Marker today**: `tests/test_narrative_plugin.py` line 38.

---

### Scar: 48-hour Guard Duplication (SM-008)

**What failed**: 48-hour reschedule logic duplicated in two places -- risk of drift.

**When**: Commit `a7a4cda`.

**How it was fixed**: Extracted to `_hours_until_appointment()` + `RESCHEDULE_HOURS_THRESHOLD` constant.

**Marker today**: `src/autom8_sms/services/orchestrator.py` line 123.

---

### Scar: mypy method-assign Stale Ignore

**What failed**: `type: ignore[assignment]` used for method reassignment -- stale error code, mypy strict treats unused ignores as errors.

**When**: Commit `25efd0c`.

**How it was fixed**: Changed to `type: ignore[method-assign]`.

**Marker today**: `src/autom8_sms/console/_instrument.py` lines 201, 222.

---

## Category Coverage

Each scar classified by failure mode category:

| Category | Scars | Count |
|----------|-------|-------|
| **Test Isolation / State Pollution** | SCAR-011, SCAR-014 | 2 |
| **AI Tool Validation / Security** | DEFECT-01, DEFECT-02, DEFECT-03 (open) | 3 |
| **Data Deserialization / Type Safety** | DEF-2, DEF-3, Pydantic Forward Ref, Direction Literal | 4 |
| **CI / Build Pipeline** | SCAR-022/DEF-009, CI SHA sourcing, AUTOM8Y_ prefix, Docker flags | 4 |
| **Config Drift** | Config env naming, Pilot hardcoded constants | 2 |
| **Silent Data Quality Masking** | Hardcoded timezone | 1 |
| **Dependency Resolution** | DC-1 devx-types, genai shim | 2 |
| **Tooling / Static Analysis** | datetime.utcnow(), mypy method-assign | 2 |
| **Integration Failure** | Inbound double-logging, WP-TEL-SMS trace gap | 2 |
| **AI Prompt Safety** | CRITICAL prompt guardrails | 1 |

**7 distinct primary categories represented** -- well above the minimum 3.

Categories searched but not found: race conditions, memory leaks, database migration failures, network timeouts.

---

## Fix-Location Mapping

| Scar | Fix Location | Function/Context | File Exists |
|------|-------------|------------------|-------------|
| SCAR-011 | `tests/conftest.py:52-53` | `_clean_sms_env` fixture | Yes |
| SCAR-014 | `tests/conftest.py:55-56` | `_clean_sms_env` fixture | Yes |
| DEF-2 | `src/autom8_sms/clients/data_service.py:287-291` | `get_business()` | Yes |
| DEF-3 | `src/autom8_sms/clients/data_service.py:231` | `get_conversation_context()` | Yes |
| DEFECT-01 | `src/autom8_sms/tools/validation.py:252-264` | `validate_business_hours()` | Yes |
| DEFECT-02 | `src/autom8_sms/tools/validation.py:196-205` | `validate_temporal_bounds_booking()` | Yes |
| DEFECT-03 | OPEN -- no fix applied | `validate_temporal_bounds_booking()` lacks end_datetime check | N/A |
| Direction Literal->str | `src/autom8_sms/models/conversation.py:24` | `MessageRecord.direction` | Yes |
| Inbound double-logging | `src/autom8_sms/services/orchestrator.py` (removed) | `process_conversation()` | Yes |
| Hardcoded timezone | `src/autom8_sms/services/orchestrator.py` | `_get_business_timezone()` | Yes |
| CI SHA sourcing | `.github/workflows/satellite-dispatch.yml` | Workflow dispatch step | Not audited (outside src scope) |
| Config env naming | `src/autom8_sms/config.py:42-53` | `SmsServiceConfig` | Yes |
| Docker flag conflicts | `Dockerfile` | `uv export` step | Not audited (outside src scope) |
| datetime.utcnow() | Codebase-wide -- `client_lead.py`, `orchestrator.py`, `scheduled_prompt.py`, `data_service.py` | Various | Yes |
| WP-TEL-SMS trace gap | `src/autom8_sms/clients/data_service.py:11-13` | Module-level documented gap | Yes |
| Pilot hardcoded constants | `src/autom8_sms/handlers/client_lead.py:69-91` | `_get_pilot_phones()` | Yes |
| CRITICAL prompt guardrails | `src/autom8_sms/prompts/sdr_prompt.py:54-57, 149` | SDR system prompt string | Yes |
| Pydantic Forward Ref | `src/autom8_sms/models/conversation.py:3`; `src/autom8_sms/console/_models.py:11` | Module-level import | Yes |
| genai Shim | `src/autom8_sms/console/_genai_attrs.py` | New shim module | Yes |
| DC-1 devx-types | `tests/test_narrative_plugin.py:38`; `pyproject.toml` optional deps | `pytest.importorskip` | Yes |
| 48h Guard | `src/autom8_sms/services/orchestrator.py:123` | `RESCHEDULE_HOURS_THRESHOLD` constant | Yes |
| mypy method-assign | `src/autom8_sms/console/_instrument.py:201,222` | `type: ignore[method-assign]` | Yes |

All Python source fix locations verified to exist. CI/CD and Dockerfile paths were not individually verified (outside `src/` scope).

---

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Regression Test |
|------|------------------|-----------------|
| SCAR-011 | `autouse` fixture resets `config_module._config = None` before each test | `tests/conftest.py` (autouse -- all tests benefit) |
| SCAR-014 | `autouse` fixture calls `structlog.configure(cache_logger_on_first_use=False)` | `tests/conftest.py` (autouse -- all tests benefit) |
| DEF-2 | try/except `ValidationError` on `model_validate` in `get_business()`, log + return None | `tests/test_defect_remediation.py::TestDef2MalformedResponseDegradation` |
| DEF-3 | `.get("start_datetime")` instead of `["start_datetime"]` in context parsing | `tests/test_defect_remediation.py::TestDef3SafeKeyAccess` (3 test cases) |
| DEFECT-01 | Hard reject in `validate_business_hours()` when `availability_response is None` | `tests/test_adversarial_tool_use.py` (adversarial suite) |
| DEFECT-02 | `MAX_LOOKFORWARD_DAYS` cap in `validate_temporal_bounds_booking()`, hard reject | `tests/test_adversarial_tool_use.py` (adversarial suite) |
| DEFECT-03 | OPEN -- no defensive pattern yet | Documented in adversarial test as known gap |
| Direction Literal->str | `direction: str` field type; outbound direction uses literal `"outbound-claude"` value | `tests/test_direction_change.py` (explicit negative test) |
| Inbound double-logging | Removed call; module docstring explains server-side logging responsibility | `tests/test_orchestrator_mock_transport.py` (flow integration) |
| Hardcoded timezone | `ValueError` raised; dispatch loop degrades gracefully via `except Exception` | Orchestrator mock transport tests |
| CI SHA sourcing | SHA-pinned workflow dispatch | No unit test (CI workflow behavior) |
| Config env naming | Single canonical env scheme; `_clean_sms_env` strips all legacy prefixes | `tests/test_config_migration.py` |
| Docker flag conflicts | `tests/test_dockerfile.py` asserts exact `uv export` flags | `tests/test_dockerfile.py` (WP-LOCK -- 9 assertions) |
| datetime.utcnow() | `from datetime import UTC`; UP017 ruff rule enforced | Ruff lint enforces at CI time |
| WP-TEL-SMS trace gap | Documented as known gap; infrastructure-level propagation is the strategy | `tests/test_telemetry.py` (verifies tracer exists, not propagation) |
| Pilot hardcoded constants | `_get_pilot_phones()` with E.164 validation; `test_old_constants_removed` assertion | `tests/test_pilot_config.py` |
| CRITICAL prompt guardrails | Explicit CRITICAL sections in SDR prompt string | `tests/test_sdr_prompt_v2.py` (prompt content tests) |
| Pydantic Forward Ref | Runtime import with `# noqa: TC003` suppression | Model instantiation tests implicitly guard |
| genai Shim | Local `_genai_attrs.py` module | `tests/test_otel_emission.py` |
| DC-1 devx-types | `pytest.importorskip` fallback | Import skip is the guard |
| 48h Guard | `RESCHEDULE_HOURS_THRESHOLD` constant + `_hours_until_appointment()` | Scheduling e2e tests |
| mypy method-assign | Updated ignore code | mypy strict mode enforces at CI |

Notable gap: The WP-TEL-SMS trace propagation gap has no regression test guarding the gap itself. DEFECT-03 is open with no fix.

---

## Agent-Relevance Tagging

| Scar | Relevant Agent Roles | Why |
|------|---------------------|-----|
| SCAR-011, SCAR-014 | Any agent writing tests | Must understand the autouse fixture that resets config and structlog between tests |
| DEF-2, DEF-3 | Any agent modifying data service client | Always wrap `model_validate` in try/except; use `.get()` not `[]` for optional API response fields |
| DEFECT-01, DEFECT-02 | Any agent modifying tool dispatch or validation | SD-04/SD-05 are security controls; do not relax validation guards |
| DEFECT-03 | Any agent working on booking validation | Known open defect -- end_datetime validation is missing |
| Direction Literal->str | Any agent modifying domain models | API response field types must not use `Literal` for fields the API may extend |
| Inbound double-logging | Any agent modifying message logging | Server-side handles inbound logging; Lambda handles outbound only |
| Hardcoded timezone | Any agent modifying scheduling logic | Never hardcode fallback timezones; fail loudly on missing config |
| Config env naming | Any agent modifying config | Follow canonical `AUTOM8Y_*` / `SMS_*` / `TWILIO_*` scheme |
| Docker flag conflicts | Any agent modifying Dockerfile | `uv export --frozen --no-hashes --no-dev` is the tested form |
| datetime.utcnow() | Any agent writing date/time code | Always use `datetime.now(UTC)` with `from datetime import UTC` |
| WP-TEL-SMS trace gap | Any agent working on observability | Trace propagation in data service calls is an open gap |
| Pilot hardcoded constants | Any agent modifying feature flags | All feature flags must be env-var-driven, never hardcoded |
| CRITICAL prompt guardrails | Any agent modifying prompts | Must not remove the CRITICAL guardrail sections |
| Pydantic Forward Ref | Any agent using TYPE_CHECKING with Pydantic models | datetime must be imported unconditionally for Pydantic v2 |

---

## Knowledge Gaps

1. **No SCAR-001 through SCAR-010 entries located.** The numbering jumps to SCAR-011 and SCAR-014, suggesting earlier scars were either pre-marking convention or in a different repo.

2. **No SCAR-012, SCAR-013, SCAR-015 through SCAR-021 entries located.** These numbers may belong to the broader autom8y monorepo scar registry.

3. **DEFECT-03 is open** -- end_datetime validation gap is documented but unfixed. No defensive pattern or regression test guards it.

4. **CI/CD and Dockerfile scars have no unit-level regression coverage** beyond what `test_dockerfile.py` checks at the text level.

5. **WP-TEL-SMS trace propagation gap has no regression guard.** A future engineer adding trace injection to the data service client has no test to verify or prevent regression.

6. **DEF-1 missing.** DEF-2 and DEF-3 exist but DEF-1 is not documented. Either it was fixed without a dedicated test or it belongs to a different module.

7. **The `type: ignore[union-attr]` suppressions on `scheduling_client` calls** in `orchestrator.py` represent an unresolved type-narrowing gap that has no named scar or defensive test.
