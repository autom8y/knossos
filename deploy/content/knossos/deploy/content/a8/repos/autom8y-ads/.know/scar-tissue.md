---
domain: scar-tissue
generated_at: "2026-03-27T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "1cfde11"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog

This catalog covers all identifiable past bugs, regressions, and failure events extracted from 162 commits of git history (2026-02-13 through 2026-03-24), inline SCAR/DEF markers, and defensive code patterns in the current source tree. Twenty-six fix-tagged commits were identified.

---

### SCAR-001: MetaAdsClient Constructor Mismatch (API Contract Violation)

**Commit**: `147cee3` -- 2026-02-13 -- `fix(ads): fix MetaAdsClient constructor call and narrow exception catch [RF-014]`

**What failed**: `MetaAdsClient` was constructed with `access_token=` as a direct kwarg. The actual SDK constructor requires a `MetaConfig` object passed as `config=`. Silent in local dev (stub fallback). Second bug: `except (ImportError, Exception)` swallowed configuration errors.

**Fix location**: `src/autom8_ads/app.py` lines 93-111 (refactored into `_init_meta_client()` helper)

**Current defensive state**: `except ImportError` only. `MetaConfig` constructed explicitly with named credential args. Real vs. stub path separated into `_init_meta_client()` helper.

---

### SCAR-002: DynamoDB Blocking the Async Event Loop (Race Condition)

**Commit**: `bfb0079` -- 2026-02-15 -- `fix(ads): use asyncio.to_thread for DynamoDB lock calls` (QA defect D-003)

**What failed**: `CampaignLock.acquire()` and `release()` called synchronous boto3 `put_item()`/`delete_item()` directly from async coroutines, blocking the entire event loop.

**Fix location**: `src/autom8_ads/lifecycle/campaign_lock.py` lines 66-79 and 98-106

**Current defensive state**: Both calls wrapped in `asyncio.to_thread()`. `NullCampaignLock` exists for test/local environments.

---

### SCAR-003: Meta API Error 1885621 -- `daily_budget` at Ad Set Level (API Contract Violation)

**Commit**: `c453a4b` -- 2026-02-20 -- `fix(meta): move budget to campaign level, add advantage_audience`

**What failed**: Ad set creation payload included `daily_budget`. Meta API rejects with error 1885621 (budget field conflict). Budget must live at campaign level only. Also, `advantage_audience` missing from `targeting_automation`.

**Fix locations**:
- `src/autom8_ads/platforms/meta/params.py` lines 44-46 (INVARIANT docstring)
- `src/autom8_ads/platforms/meta/constants.py` lines 9-13 (`META_BUDGET_CONFLICT_SUBCODE = 1885621`, `META_DAILY_BUDGET_PARAM`)

**Regression tests**:
- `tests/platforms/test_meta_params.py` line 50: `test_daily_budget_not_in_output` -- annotated `"""SCAR-003"""`
- `tests/lifecycle/test_strategy.py` line 302: `test_ad_set_params_exclude_daily_budget` -- annotated `"""SCAR-003"""`

---

### SCAR-004: Pydantic `TYPE_CHECKING`-Guarded Imports Cause Runtime Failure (Integration Failure)

**Commit**: `3b31899` -- 2026-02-25 -- `fix(models): move TYPE_CHECKING imports to runtime for Pydantic compatibility`

**What failed**: `from __future__ import annotations` + `if TYPE_CHECKING:` guards on imports used as Pydantic field types caused `PydanticUserError` at class construction time.

**Affected files**: `api/launch.py`, `api/campaigns.py`, `api/status.py`, `api/insights.py`, `models/ad.py`, `models/ad_group.py`

**Fix locations**: All affected files carry `# noqa: TC001 - needed at runtime for FastAPI Depends()`. `pyproject.toml` configures `runtime-evaluated-base-classes` for `AdsModel`.

---

### SCAR-005: CI SHA Sourcing -- Stale SHA in Satellite Dispatch (CI/CD Failure)

**Commit**: `33617b8` -- 2026-02-17

**What failed**: `workflow_run` trigger uses outer workflow SHA, not triggering run's head commit.

**Fix**: `.github/workflows/satellite-dispatch.yml` uses `github.event.workflow_run.head_sha || github.sha`.

---

### SCAR-006: Docker BuildKit `--link` + `--chown` Incompatibility (Build System)

**Commit**: `30e9a49` -- 2026-02-25

**What failed**: `COPY --link --from=builder --chown=appuser:appuser` invalid in BuildKit.

**Fix**: `Dockerfile` lines 41-42, `--link` removed from `COPY` with `--chown`. (`--link` on `COPY --from=builder` without `--chown` is retained.)

---

### SCAR-007: Docker Compose v2 Build Context Resolution (Dev Environment)

**Commit**: `9d846dc` -- 2026-02-22

**What failed**: `context: .` resolved to monorepo root under Compose v2 + BuildKit with `-f`.

**Fix**: `docker-compose.override.yml` uses `context: ${AUTOM8Y_ADS_DIR:-.}`.

---

### SCAR-008: `uv pip compile` Silently Failing for Private Registry Deps (Build System)

**Commit**: `7807f9f` -- 2026-02-22

**What failed**: `uv pip compile` silently dropped `autom8y-*` packages (not on PyPI). `UV_NO_CONFIG=1` missing.

**Fix**: `uv export --frozen | grep -v autom8[y-]` for public dep extraction. `UV_NO_CONFIG=1` on install calls.

---

### SCAR-009: Missing Production Dependency -- `autom8y-http` (Dependency Failure)

**Commit**: `81b1a4f` -- 2026-02-23

**What failed**: `autom8y-http` missing from production deps despite `TID251` banning direct `httpx` imports.

**Fix**: `pyproject.toml` -- `autom8y-http[otel]>=0.5.0` added to `dependencies`.

---

### SCAR-010: `autom8y-config` Version Constraint Stale in Lockfile (Dependency)

**Commit**: `ad18c18` -- 2026-02-22

**What failed**: `uv.lock` pinned 0.3.0 after constraint bumped to `>=0.4.0`.

**Fix**: Lockfile regenerated.

---

### SCAR-011: LaunchRequest / StatusUpdateRequest Accept Unknown Fields (API Contract Violation)

**Commits**: `a5aca4e` -- 2026-02-15 (LaunchRequest); `b86ec9c` -- 2026-03-12 (StatusUpdateRequest regression fix)

**What failed**: `LaunchRequest` had no `extra="forbid"`, silently ignoring unknown fields. The DP-007 pattern was not applied to `StatusUpdateRequest` when that model was introduced.

**Fix locations**:
- `src/autom8_ads/models/launch.py` line 78: `model_config = ConfigDict(extra="forbid")`
- `src/autom8_ads/models/responses.py` line 72: `model_config = ConfigDict(extra="forbid")` on `StatusUpdateRequest`

**Note**: The SCAR-011 regression (StatusUpdateRequest introduced without `extra="forbid"`) confirms that DP-007 is not enforced by tooling -- only by convention.

---

### SCAR-012: Ruff `target-version` Mismatch (Toolchain Drift)

**Commit**: `fb95013` -- 2026-02-25

**What failed**: Ruff `target-version` set to `py312` while `requires-python` was `>=3.11`.

**Fix**: `pyproject.toml` `[tool.ruff]` aligned to `py311`.

---

### SCAR-013: `ADS_ENVIRONMENT` Env Variable -- Env Isolation Test Failure (Config Drift)

**Commits**: `4589a09`, `deafd3b`, `a93fc3b` -- 2026-02-22

**What failed**: Migration from `ADS_ENVIRONMENT` to `AUTOM8Y_ENV` was incomplete. Host shell's `direnv` exported `AUTOM8Y_ENV=local` leaked into tests.

**Fix locations**:
- `tests/conftest.py` lines 36-45: session-scoped autouse fixture strips `AUTOM8Y_ENV`
- `src/autom8_ads/config.py` lines 55-57: `validation_alias=AliasChoices("AUTOM8Y_ENV")` (note: `ADS_ENVIRONMENT` alias removed after clean-break refactor `6ad4564`)

---

### SCAR-014: Coverage Report Missing `__main__` Exclusion (Toolchain)

**Commit**: `4fd9fd8` -- 2026-02-26

**Fix**: `pyproject.toml` `[tool.coverage.report]` `exclude_lines` updated.

---

### SCAR-015: `mypy --strict` Errors in Meta Adapter (Tooling Compliance)

**Commit**: `a5aca4e` -- 2026-02-15

**What failed**: 14 `mypy --strict` errors from incorrect `# type: ignore` suppression codes.

**Fix**: `src/autom8_ads/platforms/meta/adapter.py` stubs use `# type: ignore[no-redef]` only. Stub infrastructure later extracted to `src/autom8_ads/platforms/meta/stubs.py` (commit `052d780`).

---

### SCAR-016: AuthSettings env_prefix Causes prod URL Guard Crash (Config / Auth Integration)

**Commit**: `f3d3099` -- 2026-03-08 -- `fix(auth): pass autom8y_env to AuthSettings to prevent prod URL guard crash`

**What failed**: `AuthSettings` uses `env_prefix="AUTH__"`, so it reads `AUTH__AUTOM8Y_ENV`, not `AUTOM8Y_ENV`. Without explicit forwarding, `AuthSettings` defaulted to `LOCAL`, causing the production URL guard to reject the default `jwks_url` at service startup in non-LOCAL environments.

**Fix location**: `src/autom8_ads/app.py` line 68: `AuthSettings(dev_mode=config.auth_disabled, autom8y_env=config.autom8y_env)` -- explicitly forwards the env value from `AdsConfig`.

---

### SCAR-017: `python-jose` vs `pyjwt` Import Incompatibility in Tests (Dependency Drift)

**Commit**: `90a5095` -- 2026-03-12 -- `fix(test): replace jose JWT import with pyjwt`

**What failed**: `tests/test_jwt_verification.py` imported `from jose import jwt as jose_jwt` for HS256 algorithm confusion attack test. The autom8y-auth SDK now uses `pyjwt`, making `python-jose` an unneeded dev dep that caused import failure.

**Fix location**: `tests/test_jwt_verification.py` -- replaced `from jose import jwt as jose_jwt` with `import jwt`; `jwt.encode()` call updated accordingly.

---

### SCAR-018: GitHub Actions SARIF Upload Breaks on Private Repos (CI/CD)

**Commits**: `4b2dac8`, `5440c6c` -- 2026-03-11

**What failed**: GitHub Advanced Security (GHAS) SARIF upload step unconditionally fails for private repositories that don't have GHAS enabled. Caused CI to block on a non-essential security scan step.

**Fix**: `.github/workflows/*.yml` -- added `if: github.event.repository.private == false` conditional on SARIF upload; added `continue-on-error: true` as backstop.

---

### SCAR-019: GitHub App Token Missing for Integration Tests (CI/CD)

**Commits**: `a8e207e`, `8e14520` -- 2026-03-11

**What failed**: Integration tests required a GitHub App token with specific scopes. `GITHUB_TOKEN` (default Actions token) lacked those scopes. Tests had to be disabled, then re-enabled after GitHub App token was wired.

**Fix location**: `.github/workflows/ci.yml` -- integration tests gated on GitHub App token presence via `secrets.GH_APP_TOKEN`.

---

### SCAR-020: CleanupPipeline Used Active-Only Tree Cache (Logic / Over-Retention Bug)

**Commit**: `eee6786` -- 2026-03-24 -- `fix(cleanup): full-tree fetch + ad-group-level is_dynamic_creative detection`

**What failed (Bug 1 -- CRITICAL)**: `MetaCleanupPipeline` used `tree_cache` (filtered to `effective_status=ACTIVE`) to discover cleanup candidates. Objects that should be cleaned up (inactive, archived, orphaned) were invisible to the pipeline. Result: 100% RETAIN decisions in dry-run -- zero evictions.

**What failed (Bug 2 -- HIGH)**: `is_dynamic_creative` was checked via a broken name heuristic at campaign level. The flag must be read from `ad_group.extensions` (populated by `MetaTranslator` from the Meta API's native `is_dynamic_creative` field on ad sets).

**Fix locations**:
- `src/autom8_ads/cleanup/pipeline.py`: Added `adapter` and `translator` constructor params; `_discover_from_adapter()` path uses `get_cleanup_tree()` (all statuses); assert guard for `tree_cache is not None` in else branch
- `src/autom8_ads/platforms/meta/adapter.py`: Added `get_cleanup_tree()` public method fetching all objects regardless of status

---

### SCAR-021: CleanupPipeline DI Missing Adapter+Translator Injection (Wiring Bug)

**Commit**: `677504d` -- 2026-03-24 -- `fix(cleanup): wire adapter+translator into pipeline DI for full-tree fetch`

**What failed**: `MetaCleanupPipeline.__init__` accepted `adapter` and `translator` parameters (added in SCAR-020 fix) but `dependencies.py` was not passing them. Pipeline silently fell back to the active-only `tree_cache` path, making the SCAR-020 fix ineffective in production.

**Fix location**: `src/autom8_ads/dependencies.py` `get_cleanup_service()` function -- now extracts `platform_adapter` from `request.app.state` and instantiates `MetaTranslator()` before passing both to `MetaCleanupPipeline`.

---

### SCAR-022: `uv sync --frozen` Incompatible with `--no-sources` in uv >=0.15.4 (Build System)

**Commit**: `a69a311` -- 2026-03-24 -- `fix(ci): replace --frozen with --no-sources in uv sync (DEF-009/SCAR-022)`

**What failed**: `uv sync --frozen --no-sources` became mutually exclusive in uv >=0.15.4. `--frozen` (prevent lockfile writes) and `--no-sources` (resolve from registry instead of monorepo path deps) cannot be used together. CI builds started failing with a uv argument error.

**Fix location**: `Dockerfile` line 28: `uv sync --no-sources --no-dev` (removed `--frozen`). Comment documents the DEF-009 defect number.

---

### SCAR-023: HttpDataServiceClient Used Unauthenticated HTTP Client (Auth / API Security)

**Commit**: `8222edf` -- 2026-03-24 -- `fix(cleanup): replace unauthenticated HTTP client with ResilientCoreClient`

**What failed**: `HttpDataServiceClient` was constructed with `Autom8yHttpClient` (raw HTTP, no auth header). The autom8y-data PV endpoints are protected and returned 401. PV protection Gates 3 and 5 were blind -- data returned as empty sets -- leading to over-eviction in the cleanup pipeline.

**Fix location**: `src/autom8_ads/app.py` lines 173-186 (lifespan DI) -- `HttpDataServiceClient` now constructed with `ResilientCoreClient` (wraps `autom8y_core.Client` with retry + circuit breaker + `Authorization: Bearer` injection from `SERVICE_API_KEY`).

---

### SCAR-024: PV Endpoint URLs Wrong -- `/offers/pvs/` vs `/pvs/` (API Contract Violation)

**Commit**: `a064d2c` -- 2026-03-23 -- `fix(cleanup): correct PV endpoint URLs and register cleanup domain events`

**What failed (F-01)**: `HttpDataServiceClient` sent PV queries to `/api/v1/offers/pvs/{status}` but autom8y-data routes are at `/api/v1/pvs/{status}`. All PV queries returned 404.

**What failed (F-02)**: Five cleanup domain events (`CleanupStarted`, `CleanupCompleted`, `ObjectEvicted`, `ObjectRetained`, `TreeDriftDetected`) were defined but never registered in `app.py` lifespan `EventBus` subscriptions.

**Fix locations**:
- `src/autom8_ads/clients/data_http.py` line 184: `_fetch_pv_set()` uses `/api/v1/pvs/{status}`
- `src/autom8_ads/app.py` `lifespan()` -- five cleanup event types added to `event_bus.subscribe()` loop

---

### SCAR-025: SERVICE_API_KEY Not Forwarded to `resolve_campaign_client` (Auth / Config)

**Commit**: `64a7fae` -- 2026-03-23 -- `fix(ads): pass SERVICE_API_KEY explicitly to resolve_campaign_client`

**What failed**: The interop SDK's `resolve_campaign_client` reads `service_key` from `env_prefix + SERVICE_API_KEY`, not the bare `SERVICE_API_KEY` env var. Without explicit passing, the campaign client authenticated with an empty key, causing 401/403 errors from the data service.

**Fix location**: `src/autom8_ads/app.py` lifespan -- `resolve_campaign_client(service_key=os.environ.get("SERVICE_API_KEY", ""))` now explicitly forwards the key.

---

## Category Coverage

| Category | Scars |
|---|---|
| **Integration / API Contract Violation** | SCAR-001, SCAR-003, SCAR-004, SCAR-011, SCAR-024 |
| **Race Condition / Async Correctness** | SCAR-002 |
| **CI/CD and Build System** | SCAR-005, SCAR-006, SCAR-007, SCAR-008, SCAR-018, SCAR-019, SCAR-022 |
| **Dependency Management** | SCAR-009, SCAR-010, SCAR-017 |
| **Config / Environment Drift** | SCAR-012, SCAR-013, SCAR-014, SCAR-016 |
| **Tooling Compliance** | SCAR-015 |
| **Auth / Security** | SCAR-023, SCAR-025 |
| **Logic / Over-Retention (Pipeline)** | SCAR-020, SCAR-021 |

Categories searched but not found: data corruption, security breach, performance cliff.

---

## Fix-Location Mapping

| Scar | Fix File(s) | Key Lines / Artifact |
|------|------------|----------------------|
| SCAR-001 | `src/autom8_ads/app.py` | `_init_meta_client()` helper, lines 85-134 |
| SCAR-002 | `src/autom8_ads/lifecycle/campaign_lock.py` | 66-79, 98-106 |
| SCAR-003 | `src/autom8_ads/platforms/meta/params.py` | 44-61 |
| SCAR-003 | `src/autom8_ads/platforms/meta/constants.py` | 9-13 |
| SCAR-004 | `src/autom8_ads/api/launch.py` | line 12 (`noqa: TC001`) |
| SCAR-004 | `src/autom8_ads/api/campaigns.py` | line 14 |
| SCAR-004 | `src/autom8_ads/api/status.py` | line 13 |
| SCAR-004 | `src/autom8_ads/api/insights.py` | line 5 |
| SCAR-004 | `pyproject.toml` | `runtime-evaluated-base-classes` |
| SCAR-005 | `.github/workflows/satellite-dispatch.yml` | 45-46, 52-53 |
| SCAR-006 | `Dockerfile` | lines 41-42 |
| SCAR-007 | `docker-compose.override.yml` | `context:` line |
| SCAR-008 | `Dockerfile.dev` | build script section |
| SCAR-009 | `pyproject.toml` | `dependencies` list |
| SCAR-010 | `uv.lock` | version stanza |
| SCAR-011 | `src/autom8_ads/models/launch.py` | 78 |
| SCAR-011 | `src/autom8_ads/models/responses.py` | 72 (`StatusUpdateRequest`) |
| SCAR-012 | `pyproject.toml` | `[tool.ruff]` |
| SCAR-013 | `tests/conftest.py` | 36-45 |
| SCAR-013 | `src/autom8_ads/config.py` | 55-57 |
| SCAR-014 | `pyproject.toml` | `exclude_lines` |
| SCAR-015 | `src/autom8_ads/platforms/meta/adapter.py` | stub import block |
| SCAR-015 | `src/autom8_ads/platforms/meta/stubs.py` | entire file |
| SCAR-016 | `src/autom8_ads/app.py` | line 68 |
| SCAR-017 | `tests/test_jwt_verification.py` | line 320 |
| SCAR-018 | `.github/workflows/` | SARIF upload conditionals |
| SCAR-019 | `.github/workflows/ci.yml` | integration test gate |
| SCAR-020 | `src/autom8_ads/cleanup/pipeline.py` | constructor, lines 100-115, 156-157 |
| SCAR-020 | `src/autom8_ads/platforms/meta/adapter.py` | `get_cleanup_tree()` method |
| SCAR-021 | `src/autom8_ads/dependencies.py` | `get_cleanup_service()` function |
| SCAR-022 | `Dockerfile` | line 28 |
| SCAR-023 | `src/autom8_ads/app.py` | lifespan lines 173-186 |
| SCAR-024 | `src/autom8_ads/clients/data_http.py` | line 184 |
| SCAR-024 | `src/autom8_ads/app.py` | event bus registration |
| SCAR-025 | `src/autom8_ads/app.py` | `resolve_campaign_client()` call |

---

## Defensive Patterns

### DP-001: Import-Safe Stub Adapter Pattern (SCAR-001, SCAR-004, SCAR-015)
`src/autom8_ads/platforms/meta/adapter.py` lines 65-88: `try/except ImportError` with `_HAS_META_SDK` sentinel and stub classes imported from `src/autom8_ads/platforms/meta/stubs.py`. The stub infrastructure was extracted from inline definitions in commit `052d780` (refactor `[RF-WS2-001]`).

### DP-002: SCAR-003 Structural Invariant + Named Constant (SCAR-003)
`build_ad_set_params()` docstring INVARIANT. `META_DAILY_BUDGET_PARAM` and `META_BUDGET_CONFLICT_SUBCODE = 1885621` constants in `constants.py` lines 9-13.

### DP-003: Regression Test Annotations for Named Scars (SCAR-003)
Two tests annotated with `"""SCAR-003"""` in docstrings: `tests/platforms/test_meta_params.py:50` and `tests/lifecycle/test_strategy.py:302`.

### DP-004: `asyncio.to_thread()` for Sync AWS SDK Calls (SCAR-002)
`campaign_lock.py` lines 66-79, 98-106. `NullCampaignLock` for test environments.

### DP-005: Environment Isolation Fixture (SCAR-013)
`tests/conftest.py` lines 36-45: session-scoped autouse fixture strips `AUTOM8Y_ENV` from host environment. The `ADS_ENVIRONMENT` alias was removed in the clean-break refactor, so only the canonical var is isolated.

### DP-006: `noqa: TC001` Runtime Import Comments (SCAR-004)
All FastAPI `Depends()`/`Query()` type annotations carry explicit `# noqa: TC00x` with explanatory comments.

### DP-007: `extra="forbid"` on Inbound Request Models (SCAR-011)
`LaunchRequest` and `StatusUpdateRequest` use `ConfigDict(extra="forbid")`. Pattern must be applied to all future inbound API request models -- not enforced by tooling.

### DP-008: `AUTOM8Y_ENV` Explicit Forwarding to Sub-Settings (SCAR-016)
`app.py` `_init_auth()`: `AuthSettings(autom8y_env=config.autom8y_env)`. Pattern: any SDK settings class with its own `env_prefix` must have `autom8y_env` forwarded explicitly from `AdsConfig`.

### DP-009: `xfail` Tests as Known-Gap Sentinels
`tests/integration/test_meta_smoke.py` lines 124, 313, 343: three smoke tests marked `xfail` with detailed reason strings documenting known API prerequisites.

### DP-010: `ResilientCoreClient` for All Authenticated Service Calls (SCAR-023, SCAR-025)
`app.py` lifespan: `HttpDataServiceClient` constructed with `ResilientCoreClient` (retry + circuit breaker + auth injection). `resolve_campaign_client` receives `service_key` explicitly. Pattern: all outbound service clients must use `autom8y-http` resilience wrapper, not raw `Autom8yHttpClient`.

### DP-011: `get_cleanup_tree()` Adapter Path for Full-Status Traversal (SCAR-020)
`MetaPlatformAdapter.get_cleanup_tree()` fetches all objects regardless of `effective_status`. Cleanup pipeline uses this adapter path when `adapter` is injected; falls back to `tree_cache` only when adapter is unavailable. The assert guard at `cleanup/pipeline.py` line 157 documents the invariant.

---

## Agent-Relevance Tags

| Scar | Relevant Roles | Why |
|------|---------------|-----|
| SCAR-001 | principal-engineer, architect | SDK integration; `MetaConfig` construction required |
| SCAR-002 | principal-engineer | Any new sync-SDK calls in async context must use `asyncio.to_thread()` |
| SCAR-003 | principal-engineer, qa-adversary | Meta API structural rule; `daily_budget` must NOT appear in ad set params |
| SCAR-004 | principal-engineer | Never move field types or `Depends()` types under `TYPE_CHECKING` |
| SCAR-005 | principal-engineer (CI) | `workflow_run` trigger must use `github.event.workflow_run.head_sha` |
| SCAR-006 | principal-engineer (infra) | Docker `COPY --link` and `--chown` cannot be combined |
| SCAR-007 | principal-engineer (infra) | Compose overrides must use absolute path variables for build context |
| SCAR-008 | principal-engineer (infra) | Use `uv export --frozen` not `uv pip compile` for public dep extraction |
| SCAR-009 | principal-engineer, architect | All HTTP calls must use `autom8y-http`; raw `httpx` banned by TID251 |
| SCAR-010 | principal-engineer | After bumping SDK constraints, always regenerate `uv.lock` |
| SCAR-011 | principal-engineer, qa-adversary | All inbound API request models must have `extra="forbid"` |
| SCAR-012 | principal-engineer | Ruff `target-version` must match `requires-python` |
| SCAR-013 | principal-engineer, qa-adversary | Tests must not depend on host `AUTOM8Y_ENV`; use conftest isolation |
| SCAR-014 | principal-engineer | Coverage `exclude_lines` must include `if __name__ == "__main__":` |
| SCAR-015 | principal-engineer | Stub `type: ignore` suppressions must use valid mypy error codes |
| SCAR-016 | principal-engineer, architect | SDK settings with own `env_prefix` require explicit `autom8y_env=` forwarding |
| SCAR-017 | principal-engineer | Test JWT libs must match the auth SDK's runtime library (`pyjwt`) |
| SCAR-018 | principal-engineer (CI) | SARIF uploads require GHAS-enabled repo; guard with `continue-on-error` |
| SCAR-019 | principal-engineer (CI) | Integration test workflows require GitHub App token, not default `GITHUB_TOKEN` |
| SCAR-020 | principal-engineer, qa-adversary | Cleanup discovery must traverse all statuses; never use active-only cache |
| SCAR-021 | principal-engineer | DI wiring for pipeline must include both `adapter` and `translator` |
| SCAR-022 | principal-engineer (infra) | `uv sync --frozen` and `--no-sources` are mutually exclusive in uv >=0.15.4 |
| SCAR-023 | principal-engineer, architect | All service HTTP clients must use `ResilientCoreClient` (auth injection) |
| SCAR-024 | principal-engineer | autom8y-data PV routes are `/api/v1/pvs/{status}`, not `/api/v1/offers/pvs/` |
| SCAR-025 | principal-engineer | `resolve_campaign_client` requires explicit `service_key=` from env |

---

## Knowledge Gaps

1. SCAR-001 through SCAR-002 numbering assigned chronologically by prior theoros; no SCAR registry document found in repo
2. QA defect D-003 (referenced in SCAR-002 commit) has no tracking document in the repository
3. RF-001 through RF-005 refactoring tags absent from git history
4. DEF-009 (referenced in SCAR-022 commit) implies a separate defect registry; no `DEF-` tracking file found in repo
5. SCAR-016 through SCAR-025 assigned by this theoros; numbering not validated against any external registry
6. Incident `ads-meta-api-error-100-post-redeploy` referenced in SCAR-016 commit message has no post-mortem document in `.ledge/`
7. SCAR-020 commit message identifies three bugs (Bug 1, Bug 2, and a third implied bug); Bug 3 details not clearly documented in commit body
8. `ADS_ENVIRONMENT` alias removal timeline: the prior document shows `AliasChoices("AUTOM8Y_ENV", "ADS_ENVIRONMENT")` but the current `config.py` shows only `"AUTOM8Y_ENV"` -- the removal occurred in commit `6ad4564` (clean-break env standardization)
