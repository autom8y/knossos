---
domain: scar-tissue
generated_at: "2026-03-16T00:14:42Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

**Service**: `autom8_ads` (Python/FastAPI)
**Root**: `services/ads/`
**Observation Date**: 2026-03-16
**Git History Span**: 7 commits (full project history visible)

---

## Failure Catalog

### SCAR-001: `_try_persist` Type Annotation Mismatch (LaunchResult vs LaunchContext)

**What failed**: The `_try_persist` method in `LaunchService` had its `ctx` parameter annotated as `LaunchResult`, but the method is called with a `LaunchContext` object from Step 5 of the pipeline. These are distinct model types. The annotation was wrong from initial implementation.

**When**: Commit `e864245` (2026-02-15)

**How it was fixed**: Import `LaunchContext` and change the `ctx` parameter annotation from `LaunchResult` to `LaunchContext`. The runtime behavior was not broken (Python duck-typing), but mypy and static analysis would flag misuse.

**Marker today**: The corrected annotation is at `services/ads/src/autom8_ads/launch/service.py` lines 237-240. No inline comment marker survives.

**Category**: Type System / Static Analysis

**Fix location**: `services/ads/src/autom8_ads/launch/service.py:237-258`

**Defensive pattern spawned**: The test class `TestTryPersistTypeBug` in `tests/qa/test_service_adversarial.py:78-102` explicitly names and documents this defect. The test does not prevent regression (it only asserts `result.success is True`), but serves as a living memory of the defect.

**Regression test**: `tests/qa/test_service_adversarial.py::TestTryPersistTypeBug::test_try_persist_parameter_type` -- weak test, does not guard the annotation.

---

### SCAR-002: `AUTOM8Y_ENV` Shadowed by Child Class `env_prefix`

**What failed**: Any `BaseSettings` subclass that declares a custom `env_prefix` caused pydantic-settings to look for `{PREFIX}AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`. In ECS task definitions where only `AUTOM8Y_ENV` is injected (no prefix), this caused `autom8y_env` to silently default to `LOCAL`, which triggered a production URL guard and routed traffic incorrectly.

**When**: Commit `1367461` (2026-03-08)

**How it was fixed**: Added `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` to the `autom8y_env` base field in `autom8y-config` SDK, bypassing the child prefix. Also fixed the `ads` service `pyproject.toml` workspace source conflict.

**Marker today**: `services/ads/pyproject.toml` now uses `autom8y-config = { workspace = true }` (line 71). The `AdsConfig` class at `services/ads/src/autom8_ads/config.py` now uses `env_prefix=""` (line 18).

**Category**: Configuration / Environment Variables / ECS Deployment

**Fix location**: `services/ads/pyproject.toml:71`, `services/ads/src/autom8_ads/config.py:17-21`

**Defensive pattern spawned**: The fix in the base SDK means all child settings classes inherit correct `AUTOM8Y_ENV` reading regardless of prefix. The `AdsConfig` env_prefix was ultimately removed entirely (SCAR-003), making this a belt-and-suspenders defense.

**Regression test**: No dedicated regression test visible in the ads service test suite for this failure mode.

---

### SCAR-003: `ADS_` Env Prefix Caused Env Var Naming Drift

**What failed**: `AdsConfig` declared `env_prefix="ADS_"`, requiring all env vars to be named `ADS_DATA_SERVICE_URL`, etc. The fleet standard moved to unprefixed canonical names. Terraform and ECS task definitions used the canonical names without prefix, causing misconfiguration.

**When**: Commit `958493f` (2026-03-15) -- removes `ADS_` prefix; Commit `35f29a1` (2026-03-14) -- renames `data_service_url` to `autom8y_data_url`

**How it was fixed**: Set `env_prefix=""` in `SettingsConfigDict`. Renamed field `data_service_url` -> `autom8y_data_url`.

**Marker today**: `services/ads/src/autom8_ads/config.py:17-21` (`env_prefix=""`), line 28 (`autom8y_data_url`).

**Category**: Configuration / Environment Variables / IaC-Code Drift

**Fix location**: `services/ads/src/autom8_ads/config.py:17-18, 28`

**Defensive pattern spawned**: The docstring now reads "All settings are read from unprefixed environment variables (fleet standard)." This comment serves as a guard against re-introducing a prefix.

**Regression test**: No dedicated test. The env var naming is validated at startup.

---

### SCAR-004: `request.app.state` Attributes Untyped (Starlette Any Leak)

**What failed**: `request.app.state` is typed as `Any` in Starlette. The dependency injectors were returning `request.app.state.launch_service` without a cast, causing mypy to propagate `Any` into the typed return signatures, silently defeating type checking on all downstream uses.

**When**: Commit `3df641d` (2026-03-12), cross-service mypy remediation

**How it was fixed**: Added `cast(LaunchService, ...)` and `cast(LaunchIdempotencyCache, ...)` in `dependencies.py`.

**Marker today**: `services/ads/src/autom8_ads/dependencies.py:8, 18, 23` -- `cast()` calls.

**Category**: Type System / Static Analysis

**Fix location**: `services/ads/src/autom8_ads/dependencies.py:16-23`

**Defensive pattern spawned**: `cast()` on every `app.state` attribute access is now the pattern for this codebase.

**Regression test**: No dedicated test; validated by mypy returning 0 errors.

---

### SCAR-005: `V2MetaLaunchStrategy` Catches All Exceptions -- Domain Error Type Erasure

**What failed / Design flaw documented**: `V2MetaLaunchStrategy.execute()` wraps the entire 5-step pipeline in a bare `except Exception as e:` block and returns a `LaunchResult(success=False, ...)` rather than propagating. This means `AdsValidationError` and other typed domain exceptions are caught, converted to a string error message, and returned as a failed result. The `LaunchService` exception handler that re-raises `AdsValidationError` (which would map to HTTP 422) never sees these errors -- they arrive as a `LaunchResult` with `success=False`, resulting in HTTP 200 with `{"success": false}`.

**When**: Design choice present since the initial `feat(ads-api)` commit (`bc2e1ee`). First documented in the QA adversarial test suite.

**Marker today**: QA test `tests/qa/test_service_adversarial.py:105-135` documents this with "FINDING (RISK)" and "Severity: MEDIUM" inline comments. The strategy code is at `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py:87`.

**Category**: Error Handling / API Contract

**Fix location**: `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py:87-107` and `services/ads/src/autom8_ads/launch/service.py:150`

**Defensive pattern spawned**: Isolated URL builder and data client failures behind their own `try/except` blocks in `launch/service.py`.

**Regression test**: `tests/qa/test_service_adversarial.py::TestErrorClassification::test_classify_ads_validation_error` -- documents the behavior.

---

### SCAR-006: Cache Key Collision Risk with Colon-Containing `offer_id`

**What failed / Risk documented**: The idempotency cache key is `f"{offer_id}:{platform.value}"`. An `offer_id` containing a colon produces a key identical to a different `offer_id`+platform combination.

**When**: Present since initial implementation. Documented in QA adversarial suite.

**Severity**: LOW -- offer IDs are Asana task GIDs (numeric strings). Colon injection is not expected.

**Marker today**: `tests/qa/test_idempotency_adversarial.py:26-43`

**Category**: Cache / Key Construction

**Fix location**: `services/ads/src/autom8_ads/launch/idempotency.py:58`

**Defensive pattern spawned**: None added.

**Regression test**: `tests/qa/test_idempotency_adversarial.py::TestCacheKeyManipulation::test_offer_id_with_colon_collides_with_key_format`

---

### SCAR-007: Cache Entry with `completed_at=None` Never Expires

**What failed / Risk documented**: The `_is_expired` method checks `if entry.completed_at is not None` before computing elapsed time for non-in-progress entries. An entry with `status="completed"` but `completed_at=None` will never be evicted, creating a permanent memory leak for that key.

**When**: Present since initial implementation. Documented in QA adversarial suite.

**Severity**: LOW -- requires manual cache construction or future code regression.

**Marker today**: `tests/qa/test_idempotency_adversarial.py:167-190`

**Category**: Cache / Memory / Eviction Logic

**Fix location**: `services/ads/src/autom8_ads/launch/idempotency.py:108-110`

**Defensive pattern spawned**: None added; only documented.

**Regression test**: `tests/qa/test_idempotency_adversarial.py::TestCacheCompletionEdgeCases::test_completed_entry_with_none_completed_at`

---

### SCAR-008: Daily Budget Underflow from Integer Division (`weekly // 7`)

**What failed / Risk documented**: When `daily_budget_cents` is not explicitly provided, the mapper computes it as `payload.weekly_ad_spend_cents // 7`. If `weekly_ad_spend_cents < 7` (e.g., 1 cent), integer division produces `0`, which would be rejected by Meta API.

**When**: Present since initial implementation. Documented in QA adversarial suite.

**Severity**: Implicit MEDIUM -- zero daily budget passes local validation but fails at Meta API level.

**Marker today**: `tests/qa/test_adversarial_payload.py:121-132`

**Category**: Business Logic / Budget Calculation

**Fix location**: `services/ads/src/autom8_ads/launch/mapper.py:28-29`

**Defensive pattern spawned**: None added.

**Regression test**: `tests/qa/test_adversarial_payload.py::TestBudgetEdgeCases::test_weekly_spend_one_cent_daily_calc`

---

### SCAR-009: pyproject.toml Workspace Source Conflict (Path vs Workspace Refs)

**What failed**: The `ads` service `pyproject.toml` mixed `{ path = "../../sdks/...", editable = true }` workspace sources with `{ workspace = true }` references, causing dependency resolution failures.

**When**: Commit `1367461` (2026-03-08) -- fixed as part of the `AUTOM8Y_ENV` config fix

**How it was fixed**: All SDK sources changed from path references to `{ workspace = true }`.

**Marker today**: `services/ads/pyproject.toml:70-73`

**Category**: Build / Dependency Management

**Fix location**: `services/ads/pyproject.toml:70-73`

**Defensive pattern spawned**: The workspace-wide `uv.lock` enforces consistent source resolution.

**Regression test**: CI validates workspace dependency resolution on every push.

---

## Category Coverage

| Category | SCAR IDs | Count |
|---|---|---|
| Type System / Static Analysis | SCAR-001, SCAR-004 | 2 |
| Configuration / Environment Variables / IaC-Code Drift | SCAR-002, SCAR-003 | 2 |
| Error Handling / API Contract | SCAR-005 | 1 |
| Cache / Key Construction | SCAR-006 | 1 |
| Cache / Memory / Eviction Logic | SCAR-007 | 1 |
| Business Logic / Budget Calculation | SCAR-008 | 1 |
| Build / Dependency Management | SCAR-009 | 1 |

**Categories present**: 7 distinct failure categories across 9 scars.

**Categories NOT observed** (absent from this service's history): Race conditions, data corruption, security vulnerabilities, network retry logic failures, database migration errors, auth token handling bugs.

---

## Fix-Location Mapping

| SCAR | Fix File | Lines |
|---|---|---|
| SCAR-001 | `services/ads/src/autom8_ads/launch/service.py` | 237-240 |
| SCAR-002 | `services/ads/pyproject.toml` (SDK version pin) | 16 |
| SCAR-003 | `services/ads/src/autom8_ads/config.py` | 17-18, 28 |
| SCAR-004 | `services/ads/src/autom8_ads/dependencies.py` | 8, 18, 23 |
| SCAR-005 | `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py` | 87-107 |
| SCAR-006 | `services/ads/src/autom8_ads/launch/idempotency.py` | 58 |
| SCAR-007 | `services/ads/src/autom8_ads/launch/idempotency.py` | 108-110 |
| SCAR-008 | `services/ads/src/autom8_ads/launch/mapper.py` | 28-29 |
| SCAR-009 | `services/ads/pyproject.toml` | 70-73 |

---

## Defensive Pattern Documentation

### Pattern DP-001: `cast()` on `app.state` Attributes
Born from SCAR-004. Every singleton retrieved from `request.app.state` is wrapped in `cast(ConcreteType, request.app.state.attribute)`. Location: `services/ads/src/autom8_ads/dependencies.py:18, 23`.

### Pattern DP-002: Graceful Degradation on Non-Critical Path Failures
Born from SCAR-005's acceptance. Non-critical paths (URL builder, data client write) are wrapped in isolated `try/except` blocks that log a warning and set the output field to `None` rather than propagating. Locations: `services/ads/src/autom8_ads/launch/service.py:175-207` (URL builder), `services/ads/src/autom8_ads/launch/service.py:237-258` (`_try_persist`).

### Pattern DP-003: Idempotency Cache -- Split TTL Design
Born from the design decision to prevent duplicate launches and safely expire stale in-progress markers. In-progress entries use a 5-minute TTL; completed/failed entries use a 24-hour TTL. Location: `services/ads/src/autom8_ads/launch/idempotency.py`. ADR reference: `ADR-ADS-007` (referenced in docstring but ADR file not located in this repo).

### Pattern DP-004: `env_prefix=""` Fleet Standard Comment Guard
Born from SCAR-003. The `AdsConfig` docstring reads "All settings are read from unprefixed environment variables (fleet standard)" -- a prose guard against re-introduction of a prefix. Location: `services/ads/src/autom8_ads/config.py:3`.

### Pattern DP-005: Partial Result Tracking in Strategy
Born from SCAR-005 design. `V2MetaLaunchStrategy` tracks each entity ID independently before attempting the next step. On failure, whatever was created is returned in the `LaunchResult` so the caller can perform cleanup or build partial URLs. Location: `services/ads/src/autom8_ads/lifecycle/strategies/v2_meta.py:31-34, 87-107`.

### Pattern DP-006: `algo_version` Validator Enforcing V2-Only Strategy
Born from ADR-ADS-002. `OfferPayload` rejects any payload with `algo_version != 2`, preventing accidental re-introduction of a V1 code path. Location: `services/ads/src/autom8_ads/models/offer.py:95-99`.

### Pattern DP-007: QA Adversarial Test Suite as Living Scar Registry
The `tests/qa/` directory contains six adversarial test files whose inline comments explicitly name findings, severities, and failure modes. These serve as a living catalog of known risks.

---

## Agent-Relevance Tagging

| SCAR | Agent Roles That Need This | Why |
|---|---|---|
| SCAR-001 (`_try_persist` annotation) | principal-engineer, qa-adversary | Must verify type annotations match actual call sites when modifying pipeline methods |
| SCAR-002 (`AUTOM8Y_ENV` shadowing) | principal-engineer, architect | Must not re-introduce env_prefix on BaseSettings subclasses -- causes silent env var resolution failure in ECS |
| SCAR-003 (`ADS_` prefix drift) | principal-engineer, architect | Must use unprefixed env var names; IaC and code must agree on canonical names |
| SCAR-004 (`app.state` Any leak) | principal-engineer | Must always `cast()` when reading from `app.state` to preserve type safety |
| SCAR-005 (domain error type erasure) | architect, qa-adversary, principal-engineer | Must understand that strategy-layer exceptions become `LaunchResult(success=False)` and will NOT trigger HTTP 4xx responses |
| SCAR-006 (cache key colon collision) | qa-adversary, principal-engineer | Must validate `offer_id` does not contain `:` if cache key format changes |
| SCAR-007 (cache `completed_at=None` leak) | qa-adversary, principal-engineer | Must ensure `complete()` always sets `completed_at` to prevent permanent cache entries |
| SCAR-008 (daily budget underflow) | qa-adversary, principal-engineer | Must add minimum daily budget validation before platform adapter call to prevent zero-budget API errors |
| SCAR-009 (pyproject.toml source conflict) | principal-engineer | Must use `{ workspace = true }` for SDK dependencies, not `{ path = ... }` |

---

## Knowledge Gaps

1. **SCAR-005 has no fix plan**: The domain error type erasure is documented as "MEDIUM" severity but has no associated fix commit or ADR.
2. **ADR-ADS-007 not located**: The idempotency split-TTL design references `ADR-ADS-007` in the docstring, but no ADR file was found.
3. **No regression test for SCAR-002**: The `AUTOM8Y_ENV` child-prefix shadowing bug has no test in the ads service.
4. **SCAR-008 minimum budget guard is absent**: No `min_daily_budget_cents` validation exists.
5. **`data_writes_enabled` defaults to `False`**: Whether this is intentional ("Move 3" stage) or a forgotten flag is not documented.
6. **Git history is shallow (7 commits)**: Pre-extraction failure history is not recoverable from this log.
