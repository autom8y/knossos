---
domain: scar-tissue
generated_at: "2026-03-01T12:42:56Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "762ed0e"
confidence: 0.88
format_version: "1.0"
---

# Codebase Scar Tissue

## Failure Catalog

This catalog covers all identifiable past bugs, regressions, and failure events extracted from 66 commits of git history (2026-02-13 through 2026-02-26), plus inline SCAR markers and defensive code patterns in the current source tree. Fifteen fix-tagged commits were identified.

---

### SCAR-001: MetaAdsClient Constructor Mismatch (API Contract Violation)

**Commit**: `147cee3` -- 2026-02-13 -- `fix(ads): fix MetaAdsClient constructor call and narrow exception catch [RF-014]`

**What failed**: `MetaAdsClient` was constructed with `access_token=` as a direct kwarg. The actual SDK constructor requires a `MetaConfig` object passed as `config=`. Silent in local dev (stub fallback). Second bug: `except (ImportError, Exception)` swallowed configuration errors.

**Fix location**: `src/autom8_ads/app.py` lines 76-91

**Current defensive state**: `except ImportError` only. `MetaConfig` constructed explicitly.

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
- `src/autom8_ads/platforms/meta/constants.py` lines 9-13 (`META_BUDGET_CONFLICT_SUBCODE = 1885621`)

**Regression tests**:
- `tests/platforms/test_meta_params.py` line 50: `test_daily_budget_not_in_output` -- annotated `"""SCAR-003"""`
- `tests/lifecycle/test_strategy.py` line 302: `test_ad_set_params_exclude_daily_budget` -- annotated `"""SCAR-003"""`

---

### SCAR-004: Pydantic `TYPE_CHECKING`-Guarded Imports Cause Runtime Failure (Integration Failure)

**Commit**: `3b31899` -- 2026-02-25 -- `fix(models): move TYPE_CHECKING imports to runtime for Pydantic compatibility`

**What failed**: `from __future__ import annotations` + `if TYPE_CHECKING:` guards on imports used as Pydantic field types caused `PydanticUserError` at class construction time.

**Affected files**: `api/launch.py`, `api/campaigns.py`, `api/status.py`, `api/insights.py`, `models/ad.py`, `models/ad_group.py`

**Fix locations**: All affected files now carry `# noqa: TC001 - needed at runtime for FastAPI Depends()`. `pyproject.toml` configures `runtime-evaluated-base-classes` for `AdsModel`.

---

### SCAR-005: CI SHA Sourcing -- Stale SHA in Satellite Dispatch (CI/CD Failure)

**Commit**: `33617b8` -- 2026-02-17

**What failed**: `workflow_run` trigger uses outer workflow SHA, not triggering run's head commit.

**Fix**: `.github/workflows/satellite-dispatch.yml` uses `github.event.workflow_run.head_sha || github.sha`.

---

### SCAR-006: Docker BuildKit `--link` + `--chown` Incompatibility (Build System)

**Commit**: `30e9a49` -- 2026-02-25

**What failed**: `COPY --link --from=builder --chown=appuser:appuser` invalid in BuildKit.

**Fix**: `Dockerfile` lines 36-37, `--link` removed from COPY instructions with `--chown`.

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

### SCAR-011: LaunchRequest Accepts Unknown Fields (API Contract Violation)

**Commit**: `a5aca4e` -- 2026-02-15 (QA adversarial testing)

**What failed**: `LaunchRequest` had no `extra="forbid"`, silently ignoring unknown fields.

**Fix**: `src/autom8_ads/models/launch.py` line 78: `model_config = ConfigDict(extra="forbid")`.

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
- `tests/conftest.py` lines 10-37: session-scoped autouse fixture strips both env vars
- `src/autom8_ads/config.py` line 34: `validation_alias=AliasChoices("AUTOM8Y_ENV", "ADS_ENVIRONMENT")`

---

### SCAR-014: Coverage Report Missing `__main__` Exclusion (Toolchain)

**Commit**: `4fd9fd8` -- 2026-02-26

**Fix**: `pyproject.toml` `[tool.coverage.report]` `exclude_lines` updated.

---

### SCAR-015: `mypy --strict` Errors in Meta Adapter (Tooling Compliance)

**Commit**: `a5aca4e` -- 2026-02-15

**What failed**: 14 `mypy --strict` errors from incorrect `# type: ignore` suppression codes.

**Fix**: `src/autom8_ads/platforms/meta/adapter.py` stubs use `# type: ignore[no-redef]` only.

---

## Category Coverage

| Category | Scars |
|---|---|
| **Integration / API Contract Violation** | SCAR-001, SCAR-003, SCAR-004, SCAR-011 |
| **Race Condition / Async Correctness** | SCAR-002 |
| **CI/CD and Build System** | SCAR-005, SCAR-006, SCAR-007, SCAR-008 |
| **Dependency Management** | SCAR-009, SCAR-010 |
| **Config / Environment Drift** | SCAR-012, SCAR-013, SCAR-014 |
| **Tooling Compliance** | SCAR-015 |

Categories searched but not found: data corruption, security breach, performance cliff.

---

## Fix-Location Mapping

| Scar | Fix File(s) | Key Lines |
|------|------------|-----------|
| SCAR-001 | `src/autom8_ads/app.py` | 75-91 |
| SCAR-002 | `src/autom8_ads/lifecycle/campaign_lock.py` | 66-79, 98-106 |
| SCAR-003 | `src/autom8_ads/platforms/meta/params.py` | 44-61 |
| SCAR-003 | `src/autom8_ads/platforms/meta/constants.py` | 9-13 |
| SCAR-004 | `src/autom8_ads/api/launch.py` | 12 |
| SCAR-004 | `src/autom8_ads/api/campaigns.py` | 14 |
| SCAR-004 | `src/autom8_ads/api/status.py` | 13 |
| SCAR-004 | `src/autom8_ads/api/insights.py` | 5 |
| SCAR-004 | `pyproject.toml` | `runtime-evaluated-base-classes` |
| SCAR-005 | `.github/workflows/satellite-dispatch.yml` | 45-46, 52-53 |
| SCAR-006 | `Dockerfile` | 36-37 |
| SCAR-007 | `docker-compose.override.yml` | `context:` line |
| SCAR-008 | `Dockerfile.dev` | build script section |
| SCAR-009 | `pyproject.toml` | `dependencies` list |
| SCAR-010 | `uv.lock` | version stanza |
| SCAR-011 | `src/autom8_ads/models/launch.py` | 78 |
| SCAR-012 | `pyproject.toml` | `[tool.ruff]` |
| SCAR-013 | `tests/conftest.py` | 10-37 |
| SCAR-013 | `src/autom8_ads/config.py` | 34 |
| SCAR-014 | `pyproject.toml` | `exclude_lines` |
| SCAR-015 | `src/autom8_ads/platforms/meta/adapter.py` | 76-142 |

---

## Defensive Patterns

### DP-001: Import-Safe Stub Adapter Pattern (SCAR-001, SCAR-004, SCAR-015)
`src/autom8_ads/platforms/meta/adapter.py` lines 50-142: `try/except ImportError` with `_HAS_META_SDK` sentinel and local stub classes.

### DP-002: SCAR-003 Structural Invariant + Named Constant (SCAR-003)
`build_ad_set_params()` docstring INVARIANT. `META_DAILY_BUDGET_PARAM` and `META_BUDGET_CONFLICT_SUBCODE = 1885621` constants.

### DP-003: Regression Test Annotations for Named Scars (SCAR-003)
Two tests annotated with `"""SCAR-003"""` in docstrings: `tests/platforms/test_meta_params.py:50` and `tests/lifecycle/test_strategy.py:302`.

### DP-004: `asyncio.to_thread()` for Sync AWS SDK Calls (SCAR-002)
`campaign_lock.py` lines 66-79, 98-106. `NullCampaignLock` for test environments.

### DP-005: Environment Isolation Fixture (SCAR-013)
`tests/conftest.py` lines 10-37: session-scoped autouse fixture strips env vars.

### DP-006: `noqa: TC001` Runtime Import Comments (SCAR-004)
All FastAPI `Depends()`/`Query()` type annotations carry explicit `# noqa: TC00x` with explanatory comments.

### DP-007: `extra="forbid"` on Inbound Request Models (SCAR-011)
`LaunchRequest` uses `ConfigDict(extra="forbid")`.

### DP-008: `AUTOM8Y_ENV`/`ADS_ENVIRONMENT` Backward-Compat Alias (SCAR-013)
`config.py` line 34: `validation_alias=AliasChoices("AUTOM8Y_ENV", "ADS_ENVIRONMENT")`.

### DP-009: `xfail` Tests as Known-Gap Sentinels
`tests/integration/test_meta_smoke.py` lines 124, 313, 343: Three smoke tests marked `xfail` with detailed reason strings documenting known API prerequisites.

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

---

## Knowledge Gaps

1. SCAR-001 and SCAR-002 numbering was assigned by theoros chronologically; no SCAR registry document was found
2. QA defect D-003 (referenced in SCAR-002 commit) has no tracking document in the repository
3. RF-001 through RF-005 refactoring tags are absent from git history
4. SM-001 tag scope (service mesh/manifest refactor) origin unknown
5. Direnv venv activation failure (SCAR-013 precursor) is implied but not documented
