---
domain: scar-tissue
generated_at: "2026-03-16T15:40:19Z"
expires_after: "7d"
source_scope:
  - "./python/**/*.py"
  - "./python/**/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog

This catalog documents past bugs, regressions, and production failures extracted from git history and inline code markers. Each entry includes the commit hash, affected package, failure description, and resolution.

### SCAR-003 — Circuit Breaker Label Cardinality (Prometheus Metric Explosion)

**Package**: `autom8y-http`
**Commit**: Surfaced in resilience module design; referenced in `python/autom8y-http/src/autom8y_http/circuit_breaker.py:312` and `python/autom8y-http/src/autom8y_http/resilience/registry.py:45`
**What happened**: Circuit breaker state transitions were logged with full URL paths as label values. URLs are unbounded (include IDs, query strings), causing Prometheus label cardinality explosion — each unique URL creates a new metric time-series, exhausting memory.
**Fix**: The `circuit_group` field (bounded, alphanumeric, 1-128 chars, validated by `_GROUP_NAME_PATTERN`) replaced raw URLs as the observability label. `CircuitBreakerRegistry` enforces naming rules at registration time (`max_groups=50`, hard cap `=200`).
**Defensive code location**: `python/autom8y-http/src/autom8y_http/resilience/registry.py:44-49`

### SCAR-011 — Settings Singleton Cache Between Tests (Test Isolation Failure)

**Package**: Cross-cutting (autom8y-config pattern; consumers include autom8y-sendgrid, autom8y-gcal, autom8y-sms-test)
**Commit**: `48d2ad1` (`fix(hygiene): FN-010 - cache get_settings() with @lru_cache`)
**What happened**: `get_settings()` was introduced with `@lru_cache()` to create a singleton Settings instance. A side effect: tests that mutate environment variables (`monkeypatch.setenv`) ran against a cached settings object from a previous test, not the modified environment. Config state bled between tests.
**Fix**: Every consumer `conftest.py` must call `config_module._config = None` (or the relevant `cache_clear()` method) in an `autouse` fixture. The pattern is documented in `autom8y-sms-test`'s `CONFTEST_TEMPLATE`.
**Defensive code locations**:
- `python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py:6-8,83,104-105`
- `python/autom8y-sendgrid/tests/conftest.py:15`
- `python/autom8y-gcal/tests/conftest.py:16-22`

### SCAR-014 — structlog `cache_logger_on_first_use` Test State Leak

**Package**: Cross-cutting (autom8y-log, autom8y-http, autom8y-gcal, autom8y-sendgrid, autom8y-sms-test)
**Commit**: Referenced in `autom8y-log/tests/conftest.py:23` and `autom8y-log/tests/conftest.py:40`
**What happened**: `structlog` caches the logging pipeline after the first logger invocation when `cache_logger_on_first_use=True`. Production code sets this flag for performance. In tests, a logger configured in one test function's scope persists into subsequent tests, causing assertions on log output to fail or observe the wrong processor chain.
**Fix**: Test `conftest.py` files call `structlog.configure(cache_logger_on_first_use=False)` in an `autouse` fixture before each test. Production code retains `True`; test harness overrides it.
**Defensive code locations**:
- `python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py:112-113`
- `python/autom8y-log/tests/conftest.py:23-40`
- `python/autom8y-gcal/tests/conftest.py:26-35`

### SCAR-AUTH-001 — Logger Positional Args TypeError → Production 500s

**Package**: `autom8y-auth`
**Commit**: `71b58ca` (`fix(auth): remove positional args from logger calls — fix production 500s`)
**Date**: 2026-02-12
**What happened**: `autom8y-auth` v1.0.0 used stdlib `%s` positional formatting in logger calls (e.g., `logger.info("Validating token %s", token_id)`). The `LoguruToProtocolAdapter` (from `autom8y-log`) only accepts `(event: str, **kwargs)`. Passing positional args caused `TypeError` on every JWT validation call, meaning every authenticated reconciliation request returned HTTP 500.
**Fix**: All logger calls converted to keyword-only: `logger.info("validating_token", token_id=token_id)`. The `LoggerProtocol` signature (`def info(self, event: str, *args: Any, **kwargs: Any)`) accepts `*args` for stdlib compat but the `LoguruToProtocolAdapter` implementation discards them.
**Fix location**: `python/autom8y-auth/src/autom8y_auth/` (all logger calls)

### SCAR-AUTH-002 — Dev Mode Bypass Always Returns UserClaims

**Package**: `autom8y-auth`
**Commit**: `382cc6b` (`fix(auth): fix dev mode bypass for service token validation`)
**Date**: 2026-03-02
**What happened**: `autom8y-auth` v1.1.0 `_dev_bypass_claims()` always returned `UserClaims`. When `validate_service_token()` was called (S2S auth), it received `UserClaims` and raised `InvalidTokenTypeError`. S2S authentication was completely broken in dev mode even with `AUTH_DEV_MODE=true`.
**Fix**: Added `_dev_bypass_service_claims()` returning `ServiceClaims(scope="*")`. Both `validate_service_token()` and `validate_user_token()` guard with dev mode short-circuits independently.
**Fix location**: `python/autom8y-auth/src/autom8y_auth/token_manager.py`

### SCAR-AUTH-003 — `datetime.utcnow()` Timezone-Naive Tokens (SEC-001)

**Package**: `autom8y-auth` (and auth service, pull-payments)
**Commit**: `a7f086c` (`fix(security): replace deprecated datetime.utcnow() with timezone-aware alternative`)
**Date**: 2026-02-10
**What happened**: 220 instances of `datetime.utcnow()` across auth service, pull-payments, and autom8y-auth SDK produced timezone-naive datetimes. Python 3.12 deprecates this. Downstream code comparing naive datetimes against timezone-aware values raises `TypeError` at runtime.
**Fix**: Replaced with `datetime.now(timezone.utc)` and `datetime.fromtimestamp(ts, tz=timezone.utc)` throughout.
**Fix locations**: `python/autom8y-auth/src/autom8y_auth/credentials.py:103` shows the current pattern (`datetime.now(UTC)`).

### SCAR-AUTH-004 — Refresh Token Storage Silent 200 on Failure

**Package**: `autom8y-auth` (auth service)
**Commit**: `96011fb` (`fix(auth): close 5xx observability gaps and fix login token bug`)
**Date**: 2026-03-06
**What happened**: When refresh token storage failed, the endpoint returned HTTP 200 with a broken (unstored) token. Clients had no signal that the token was invalid until they attempted to use it, causing confusing downstream failures.
**Fix**: Storage failure path now returns HTTP 500 so callers receive an honest failure and can retry.

### SCAR-CONFIG-001 — env_prefix Shadows AUTOM8Y_ENV

**Package**: `autom8y-config`
**Commit**: `1367461` (`fix(config): AUTOM8Y_ENV now read regardless of child class env_prefix`)
**Date**: 2026-03-08
**What happened**: Child classes with a custom `env_prefix` (e.g., `AUTH__`, `STRIPE_`, `ADS_`) caused pydantic-settings to look for `{PREFIX}AUTOM8Y_ENV` (e.g., `AUTH__AUTOM8Y_ENV`) instead of the canonical `AUTOM8Y_ENV`. ECS task definitions only set `AUTOM8Y_ENV`. Result: `autom8y_env` defaulted to `LOCAL`, triggering the production URL guard and blocking service startup.
**Fix**: `autom8y_env` field now uses `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")`, bypassing pydantic-settings' prefix logic for this specific field.
**Fix location**: `python/autom8y-config/src/autom8y_config/base_settings.py:87-93`

### SCAR-TELEMETRY-001 — Lambda BatchSpanProcessor Spans Lost on Freeze

**Package**: `autom8y-telemetry`
**Commit**: `dac2866` (`fix(telemetry): add force_flush for Lambda + FastAPI span instrumentation`)
**Date**: 2026-03-06
**What happened**: Lambda freezes execution immediately after handler return. `BatchSpanProcessor` buffers spans on a 5-second timer interval. Result: spans created during Lambda invocations were never exported to Tempo.
**Fix**: `_force_flush_provider()` called in a `try/finally` block in `instrument_lambda`'s wrapper. Uses `hasattr(provider, "force_flush")` guard for safety with no-op `ProxyTracerProvider`.
**Fix location**: `python/autom8y-telemetry/src/autom8y_telemetry/aws/lambda_instrument.py:96-101, 111-119`

### SCAR-TELEMETRY-002 — FastAPI `instrument_app()` Never Called FastAPIInstrumentor

**Package**: `autom8y-telemetry`
**Commit**: `dac2866` (same as SCAR-TELEMETRY-001)
**Date**: 2026-03-06
**What happened**: `instrument_app()` initialized `TracerProvider` and the OTLP exporter but never actually called `FastAPIInstrumentor.instrument_app()`. No per-request spans were created for any ECS/FastAPI service despite full OTLP configuration.
**Fix**: Added missing `FastAPIInstrumentor.instrument_app(app, tracer_provider=provider)` call in `instrument_app()`.
**Fix location**: `python/autom8y-telemetry/src/autom8y_telemetry/fastapi/instrument.py`

### SCAR-TELEMETRY-003 — OTLP HTTP Exporter Missing `/v1/traces` Suffix (SCAR-SRE-016)

**Package**: `autom8y-telemetry`
**Commit**: `14b415a` (`fix(telemetry): append /v1/traces to HTTP OTLP exporter endpoint`)
**Date**: 2026-03-06
**What happened**: The OpenTelemetry Python SDK auto-appends `/v1/traces` only when reading `OTEL_EXPORTER_OTLP_ENDPOINT` from an environment variable, not when the endpoint is passed as a constructor kwarg. `autom8y-telemetry` passed it as a kwarg. Lambda services (using HTTP protocol) silently exported to the wrong URL path, dropping all traces.
**Fix**: `_create_exporter()` explicitly appends `/v1/traces`: `endpoint = config.endpoint.rstrip("/") + "/v1/traces"`.
**Fix location**: `python/autom8y-telemetry/src/autom8y_telemetry/init.py:204-208`

### SCAR-TELEMETRY-004 — Silent OTLP ImportError Swallowed (SCAR-SRE-011)

**Package**: `autom8y-telemetry`
**Commit**: `f3bc481` (`fix(telemetry): add otlp extra to 4 services, surface silent export failures`)
**Date**: 2026-03-06
**What happened**: `_create_exporter()` caught `ImportError` when `opentelemetry-exporter-otlp-proto-http` was missing, but only logged via the caller-provided `logger` parameter. `instrument_lambda()` called `init_telemetry()` without passing a logger. The `ImportError` was silently swallowed.
**Fix**: `_create_exporter()` now always logs via `stdlib logging.getLogger(__name__)` as a fallback regardless of whether a caller-provided logger is present.
**Fix location**: `python/autom8y-telemetry/src/autom8y_telemetry/init.py:179, 193-197, 213-219`

### SCAR-TELEMETRY-005 — Convention Checker Case-Sensitive `_condition_met` (WARN-1)

**Package**: `autom8y-telemetry`
**Commit**: `1139ccc` (`fix(autom8y-telemetry): case-insensitive _condition_met comparison`)
**Date**: 2026-03-16
**What happened**: The convention checker's `_condition_met()` performed exact-match string comparison for conditional requirement values. Bundled YAML conventions used uppercase (`FAIL`), but services emitted lowercase (`fail`). Conditional requirements silently passed when they should have failed.
**Fix**: Added `.lower()` normalization before comparison. 6 regression tests in `tests/conventions/test_condition_met.py`.
**Fix location**: `python/autom8y-telemetry/src/autom8y_telemetry/conventions/checker.py`

### SCAR-TELEMETRY-006 — OTel Convention Registry appointment_id Type Mismatch

**Package**: `autom8y-telemetry`
**Commits**: `ec910b0` (type fix), `98cd3d0` (enum extension)
**Date**: 2026-03-16
**What happened**: The scheduling namespace YAML declared `appointment_id` as `type: string` with UUID examples. The database column is an auto-increment integer. This caused 9 `TYPE_MISMATCH` violations in CI convention checks. Separately, `scheduling.status` enum only included `booked`, `idempotent_success`, `conflict` — missing `cancelled`, `rescheduled`, `already_cancelled`, `not_found`, `not_cancellable` emitted by the BookingEngine.
**Fix**: Updated both source YAML and bundled `_data` copy for both issues.
**Fix locations**:
- `python/autom8y-telemetry/conventions/namespaces/scheduling.yaml`
- `python/autom8y-telemetry/src/autom8y_telemetry/conventions/_data/namespaces/scheduling.yaml`

### SCAR-HTTP-001 — Missing `pydantic-settings` Dep in `autom8y-http[core]` (RS-A01)

**Package**: `autom8y-http`
**Commit**: `86a276d` (`fix(sdks): add missing pydantic-settings dep, fix phantom autom8y-http dep`)
**Date**: 2026-02-19
**What happened**: `autom8y-http[core]` had an unconditional `from pydantic_settings import BaseSettings, SettingsConfigDict` in `resilience/config.py` but `pydantic-settings` was not listed as a dependency. Consumers installing `autom8y-http[core]` without separately declaring `pydantic-settings` got `ImportError` at import time.
**Fix**: Added `pydantic-settings>=2.0.0` to `autom8y-http`'s optional `[core]` extras.

### SCAR-HTTP-002 — `autom8y-slack` Phantom Dependency on `autom8y-http` (RS-A02/A03)

**Package**: `autom8y-slack`
**Commit**: `86a276d` (same as HTTP-001)
**Date**: 2026-02-19
**What happened**: `autom8y-slack` declared `autom8y-http>=0.3.0` as a dependency, but zero files in `autom8y-slack/src/` imported `autom8y_http`. The phantom dependency inflated the transitive closure without benefit.
**Fix**: Replaced with `httpx>=0.25` direct dependency.

### SCAR-HTTP-003 — W3C Traceparent Not Propagated Cross-Service (DC-7)

**Package**: `autom8y-http`
**Commit**: `d4996f0` (`fix(autom8y-http): inject W3C traceparent via ResilientCoreClient`)
**Date**: 2026-03-15
**What happened**: `ResilientCoreClient` made outbound HTTP requests without injecting the W3C `traceparent` header. Distributed traces broke at service boundaries.
**Fix**: `_inject_trace_context()` added to `ResilientCoreClient`, calling `opentelemetry.propagate.inject()` into request headers before the first attempt. Injection is gated on OTel availability and silently no-ops on any exception.
**Fix location**: `python/autom8y-http/src/autom8y_http/resilience/client.py:137-138`

### SCAR-GCAL-001 — Missing `requests` Dep for `google-auth` Transport

**Package**: `autom8y-gcal`
**Commit**: `b82324b` (`fix(gcal): add requests as explicit dependency for google-auth transport`)
**Date**: 2026-03-12
**What happened**: `google-auth` transport layer requires the `requests` library for HTTP-based credential refreshes. It was not declared in `autom8y-gcal`'s explicit dependencies.
**Fix**: Added `requests` to `autom8y-gcal/pyproject.toml` explicit deps.

### SCAR-GCAL-002 — camelCase Field Access Broken Under mypy Strict (aab3499)

**Package**: `autom8y-gcal`
**Commit**: `aab3499` (`fix(gcal): use validation_alias for mypy-safe camelCase field access`)
**What happened**: Google Calendar API returns camelCase fields (`dateTime`, `timeZone`, `htmlLink`, `nextPageToken`, `nextSyncToken`, `resourceId`, `resourceUri`). Direct camelCase attribute access was not mypy-safe under `--strict`.
**Fix**: All camelCase fields use `validation_alias=AliasChoices("snake_case", "camelCase")` so both the API response (camelCase) and Python code (snake_case) work correctly.
**Fix location**: `python/autom8y-gcal/src/autom8y_gcal/models.py:62,66,140,161,164,185,187`

### SCAR-AI-001 — `SecretStr.__eq__` Does Not Coerce to `str`

**Package**: `autom8y-ai`
**Commit**: `a445d0e` (`fix(ai): use get_secret_value() for SecretStr api_key comparisons`)
**Date**: 2026-03-11
**What happened**: `api_key` changed from `str` to `pydantic.SecretStr` in `autom8y-ai` v1.3.0. Test assertions compared `adapter._config.api_key == "sk-ant-..."` directly. `SecretStr.__eq__` does not coerce to `str`, so all such comparisons returned `False`, causing CI test failures.
**Fix**: All test assertions use `adapter._config.api_key.get_secret_value() == "sk-ant-..."`.

### SCAR-INTEROP-001 — Scheduling Error Hierarchy Split (Dual Type Identity)

**Package**: `autom8y-interop`, `autom8y-core`
**Commits**: `3a7f57c` (restore), `95ab426` (consolidate)
**What happened**: `SchedulingError`, `SlotConflictError`, and `TimezoneNotConfiguredError` were independently defined in both `autom8y-core` (extending `TransportError`) and `autom8y-interop` (extending `ServiceError`). The two hierarchies had incompatible base classes, causing `isinstance` checks to fail.
**Fix**: Interop retains local `SchedulingError` definitions extending `ServiceError`. `autom8y-interop/src/autom8y_interop/data/errors.py` re-exports from `autom8y_core.errors`.
**Fix location**: `python/autom8y-interop/src/autom8y_interop/data/errors.py:14-27`

### SCAR-PYTEST-001 — `pytest-asyncio` 1.0 Breaking API (`Package.obj`)

**Package**: All 18+ SDK packages (monorepo-wide)
**Commit**: `03f8175` (`fix(sdks): pin pytest-asyncio>=1.2,<2.0 to fix Package.obj breakage`)
**Date**: 2026-03-06
**What happened**: `pytest-asyncio` 1.0 introduced breaking API changes. `'Package' object has no attribute 'obj'` blocked the SDK Publish CI workflow entirely.
**Fix**: Pin `pytest-asyncio>=1.2,<2.0` across all 20 `pyproject.toml` files.

### SCAR-PYTEST-002 — Monorepo Duplicate Test Basenames Cause ImportPathMismatchError

**Package**: Root monorepo (affects all 23+ packages)
**Commit**: `299aded` (`fix(autom8y): add --import-mode=importlib to root pytest config`)
**Date**: 2026-03-16
**What happened**: 23+ packages each have `test_config.py`, `test_client.py`, etc. Pytest's default `prepend` import mode caused 93 `ImportPathMismatchError` collection failures.
**Fix**: `--import-mode=importlib` added to root `pyproject.toml:79`.

### SCAR-TELEMETRY-007 — sys.path Hacks in Test Files

**Package**: `autom8y-telemetry`
**Commit**: `574f870` (`refactor(telemetry): remove sys.path hacks from test files [RF-005]`)
**Date**: 2026-03-15
**What happened**: Telemetry test files used `sys.path.insert(0, ...)` to import from the `conventions/` directory rather than through the installed package. Broke under `--import-mode=importlib`.
**Fix**: Replaced with clean imports from `autom8y_telemetry.conventions.checker` and `autom8y_telemetry.conventions.registry`.

### SCAR-AUTH-FASTAPI-001 — FastAPI Imports Eagerly Required in Non-FastAPI Consumers

**Package**: `autom8y-auth`
**Commit**: `8419771` (`fix(sdk): Make FastAPI imports lazy to avoid requiring optional deps`)
**Date**: 2025-12-18
**What happened**: `autom8y-auth` v0.x eagerly imported FastAPI-specific exports at module load time. Non-web consumers got `ImportError` on any import from `autom8y_auth`.
**Fix**: FastAPI-specific exports are lazy-loaded; attempting to use them without FastAPI installed raises a descriptive `ImportError` rather than failing at import time.

## Category Coverage

### Observability / Tracing Infrastructure (4 scars)
SCAR-TELEMETRY-001, SCAR-TELEMETRY-002, SCAR-TELEMETRY-003, SCAR-TELEMETRY-004. All four were root causes of zero traces reaching Tempo despite full OTLP configuration. This is the densest scar cluster.

### Test Isolation / State Leakage (3 scars)
SCAR-011, SCAR-014, SCAR-PYTEST-001. Failures caused tests to observe state from previous tests or to fail to collect entirely.

### Integration / Configuration Failure (3 scars)
SCAR-CONFIG-001, SCAR-GCAL-001, SCAR-HTTP-001. Undeclared or misconfigured integration points that worked locally but failed in deployment.

### Security / Type System (2 scars)
SCAR-AUTH-003, SCAR-AI-001.

### Production API Failures (2 scars)
SCAR-AUTH-001, SCAR-AUTH-004. Both caused user-visible production breakage.

### Dependency / Phantom Dep Management (2 scars)
SCAR-HTTP-002, SCAR-GCAL-002. Undeclared or incorrect dependency declarations.

### Error Hierarchy / Type Identity (1 scar)
SCAR-INTEROP-001. Premature consolidation that split type identity across packages.

### Observability / Metric Cardinality (1 scar)
SCAR-003. Unbounded Prometheus label values.

### Build / CI Infrastructure (2 scars)
SCAR-PYTEST-001, SCAR-PYTEST-002. Both blocked CI entirely.

### Auth / Authorization Logic (1 scar)
SCAR-AUTH-002. Local dev workflow broken, S2S auth impossible.

### Categories searched but not observed:
- **Data corruption**: Not observed (no evidence of corrupted persistent data)
- **Race condition**: Not observed (though SCAR-011/SCAR-014 are state-leakage, not races)
- **Schema evolution**: Not observed

## Fix-Location Mapping

| Scar ID | Fix File(s) |
|---------|-------------|
| SCAR-003 | `python/autom8y-http/src/autom8y_http/resilience/registry.py:44-49`, `circuit_breaker.py:312` |
| SCAR-011 | `python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py:104-105`, `python/autom8y-sendgrid/tests/conftest.py:15`, `python/autom8y-gcal/tests/conftest.py:16-22` |
| SCAR-014 | `python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py:112-113`, `python/autom8y-log/tests/conftest.py:23-40`, `python/autom8y-gcal/tests/conftest.py:26-35` |
| SCAR-AUTH-001 | `python/autom8y-auth/src/autom8y_auth/` (all logger call sites) |
| SCAR-AUTH-002 | `python/autom8y-auth/src/autom8y_auth/token_manager.py` |
| SCAR-AUTH-003 | `python/autom8y-auth/src/autom8y_auth/credentials.py:103` |
| SCAR-AUTH-004 | `python/autom8y-auth/src/autom8y_auth/` (login endpoint) |
| SCAR-CONFIG-001 | `python/autom8y-config/src/autom8y_config/base_settings.py:87-93` |
| SCAR-TELEMETRY-001 | `python/autom8y-telemetry/src/autom8y_telemetry/aws/lambda_instrument.py:96-101, 111-119` |
| SCAR-TELEMETRY-002 | `python/autom8y-telemetry/src/autom8y_telemetry/fastapi/instrument.py` |
| SCAR-TELEMETRY-003 | `python/autom8y-telemetry/src/autom8y_telemetry/init.py:204-208` |
| SCAR-TELEMETRY-004 | `python/autom8y-telemetry/src/autom8y_telemetry/init.py:179, 193-197, 213-219` |
| SCAR-TELEMETRY-005 | `python/autom8y-telemetry/src/autom8y_telemetry/conventions/checker.py` |
| SCAR-TELEMETRY-006 | `python/autom8y-telemetry/conventions/namespaces/scheduling.yaml`, `src/autom8y_telemetry/conventions/_data/namespaces/scheduling.yaml` |
| SCAR-HTTP-001 | `python/autom8y-http/pyproject.toml` |
| SCAR-HTTP-002 | `python/autom8y-slack/pyproject.toml` |
| SCAR-HTTP-003 | `python/autom8y-http/src/autom8y_http/resilience/client.py:137-138` |
| SCAR-GCAL-001 | `python/autom8y-gcal/pyproject.toml` |
| SCAR-GCAL-002 | `python/autom8y-gcal/src/autom8y_gcal/models.py:62,66,140,161,164,185,187` |
| SCAR-AI-001 | `python/autom8y-ai/tests/test_adapter.py`, `tests/test_fixtures.py` |
| SCAR-INTEROP-001 | `python/autom8y-interop/src/autom8y_interop/data/errors.py:14-27` |
| SCAR-AUTH-FASTAPI-001 | `python/autom8y-auth/src/autom8y_auth/__init__.py` |
| SCAR-PYTEST-001 | All 18+ `pyproject.toml` files |
| SCAR-PYTEST-002 | Root `pyproject.toml:79` |

## Defensive Patterns

### Pattern 1: Bounded Circuit Breaker Group Names (from SCAR-003)
`_GROUP_NAME_PATTERN = re.compile(r"^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$")` enforced at registration time. `max_groups=50`, hard cap `=200`. Circuit state logged with `circuit_group` (never URLs or IDs).

### Pattern 2: Settings Cache Clear Autouse Fixture (from SCAR-011)
Every consumer `conftest.py`: `config_module._config = None` or `.cache_clear()` in `@pytest.fixture(autouse=True)`.

### Pattern 3: structlog Cache Disable in Tests (from SCAR-014)
`structlog.configure(cache_logger_on_first_use=False)` in autouse test fixture.

### Pattern 4: LoggerProtocol Keyword-Only Calls (from SCAR-AUTH-001)
All logger calls use `logger.info("event_name", key=value)` — never positional `%s` formatting.

### Pattern 5: `validation_alias=AliasChoices` for Env-Prefix Bypass (from SCAR-CONFIG-001)
Fields that must be read from a canonical env var regardless of child `env_prefix` use `AliasChoices("field_name", "CANONICAL_ENV_VAR")`.

### Pattern 6: Lambda Force-Flush in `try/finally` (from SCAR-TELEMETRY-001)
`_force_flush_provider()` in `finally` block of every Lambda handler wrapper.

### Pattern 7: Explicit `/v1/traces` Suffix for HTTP OTLP Constructor (from SCAR-TELEMETRY-003)
`endpoint = config.endpoint.rstrip("/") + "/v1/traces"` — always explicit when passing endpoint as constructor kwarg.

### Pattern 8: Stdlib Fallback Logger for ImportError (from SCAR-TELEMETRY-004)
`_fallback = _logging.getLogger(__name__)` inside `_create_exporter()`. Always logs to stdlib even when caller-provided logger is `None`.

### Pattern 9: `pytest.importorskip` for Optional Dependencies in Tests
`autom8y_core = pytest.importorskip("autom8y_core", reason="...")` at module level.

### Pattern 10: `get_secret_value()` for SecretStr Comparisons (from SCAR-AI-001)
Never compare `SecretStr` with `==` against a plain `str`. Always call `.get_secret_value()` first.

### Pattern 11: `--import-mode=importlib` in Root Pytest Config (from SCAR-PYTEST-002)
Monorepo with duplicate test basenames across packages requires `importlib` mode. Never remove this setting.

## Agent-Relevance Tags

| Scar ID | Relevant Agent Roles | Why |
|---------|---------------------|-----|
| SCAR-003 | principal-engineer, qa-adversary | Any new observable metric label must use bounded group names |
| SCAR-011 | principal-engineer, qa-adversary | Every new package `conftest.py` must include settings cache clear autouse fixture |
| SCAR-014 | principal-engineer, qa-adversary | Every new package `conftest.py` must include structlog cache disable autouse fixture |
| SCAR-AUTH-001 | principal-engineer | Never use `%s` positional format with `LoggerProtocol`; always keyword args |
| SCAR-AUTH-002 | principal-engineer, qa-adversary | Dev mode bypass logic must handle both UserClaims and ServiceClaims independently |
| SCAR-AUTH-003 | principal-engineer | Use `datetime.now(timezone.utc)`, never `datetime.utcnow()` |
| SCAR-AUTH-004 | principal-engineer, qa-adversary | Storage failures must return honest HTTP errors, never silent 200 |
| SCAR-CONFIG-001 | architect, principal-engineer | Child config classes with `env_prefix` must use `AliasChoices` for canonical env vars |
| SCAR-TELEMETRY-001 | principal-engineer | All Lambda handlers must use `instrument_lambda` decorator |
| SCAR-TELEMETRY-002 | principal-engineer | `instrument_app()` must be called for FastAPI |
| SCAR-TELEMETRY-003 | principal-engineer | Use `autom8y-telemetry >= 0.5.2`; do not manually construct OTLP HTTP endpoints |
| SCAR-TELEMETRY-004 | principal-engineer | `instrument_lambda` must pass a logger to `init_telemetry` |
| SCAR-TELEMETRY-005 | principal-engineer | Convention YAML values are case-sensitive at definition time; checker normalizes at check time |
| SCAR-TELEMETRY-006 | principal-engineer | OTel convention registry and bundled `_data/` copy must be updated together |
| SCAR-HTTP-001 | principal-engineer | Optional dep extras must declare all their transitive requirements |
| SCAR-HTTP-002 | architect, principal-engineer | Phantom deps must not be declared; verify with import analysis before adding |
| SCAR-HTTP-003 | principal-engineer | All inter-service HTTP clients must use `ResilientCoreClient` (handles traceparent injection) |
| SCAR-GCAL-001 | principal-engineer | All transitive google-auth transport deps must be explicit in pyproject.toml |
| SCAR-GCAL-002 | principal-engineer | Google API response fields need `AliasChoices` for camelCase + snake_case dual access |
| SCAR-AI-001 | principal-engineer, qa-adversary | All `SecretStr` comparisons must call `.get_secret_value()` |
| SCAR-INTEROP-001 | architect, principal-engineer | Error hierarchies spanning packages must preserve `service_name` for cross-service callers |
| SCAR-AUTH-FASTAPI-001 | principal-engineer | Optional framework deps must be lazy-imported with helpful error messages |
| SCAR-PYTEST-001 | principal-engineer | Do not upgrade `pytest-asyncio` past `<2.0` without reviewing API changes |
| SCAR-PYTEST-002 | principal-engineer | Never remove `--import-mode=importlib` from root pytest config |

## Knowledge Gaps

1. **SCAR-SRE-001, SCAR-SRE-009, SCAR-SRE-010, SCAR-SRE-017** are referenced in runbooks outside this SDK monorepo scope. Their SDK-side defensive patterns (if any) were not found within `sdks/python/`.

2. **SCAR-003 and SCAR-014 origin commits** were not found — these scars are referenced widely in conftest comments but no `fix(SCAR-NNN)` commits were identified.

3. **Scars in `autom8y-cache`, `autom8y-events`, `autom8y-stripe`, `autom8y-meta`, `autom8y-reconciliation`** had no named SCAR identifiers found via git history or inline comments.

4. **Data corruption, race condition, schema evolution** categories searched but not observed in this SDK monorepo.
