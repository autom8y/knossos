---
domain: scar-tissue
generated_at: "2026-03-23T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "4febf1f"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

**Project**: autom8y-meta (Python SDK)
**Language**: Python 3.12+
**Observation date**: 2026-03-23

## Failure Catalog Completeness

The failure catalog is assembled from two evidence streams: git commit history scoped to `sdks/python/autom8y-meta/`, and cross-cutting monorepo commits with markers or explicit autom8y-meta references. No inline code markers (SCAR-, DEF-, HACK, FIXME, CRITICAL, WORKAROUND) were found in any autom8y-meta source file — the project uses commit message tags and a TDD-referenced marker scheme instead.

### SCAR-001 — Docker COPY --link overlay failure

**What failed**: Docker builds using `COPY --link --from=uv /uv /uvx /bin/` failed with a `/bin/sh` overlay error in BuildKit when the source stage is not a base image layer. The `--link` flag is only safe for base-image copies.

**When**: Commit `eb77ac4` (Docker fix) then formalized as SCAR-001 in commit `251a5f9` (2026-02-27).

**Fix commits**:
- `eb77ac4` — `fix(docker): remove --link from COPY --from=uv to fix BuildKit /bin/sh overlay`
- `5f52b71` — `fix(auth-mysql-sync): remove --link from COPY --from=uv to fix BuildKit /bin/sh overlay`
- `251a5f9` — `fix(templates): harden scaffold Dockerfiles against SCAR-001 and supply-chain risks`

**Fix location**: Scaffold templates at `scripts/templates/Dockerfile.ecs-fargate` and `scripts/templates/Dockerfile.lambda-scheduled` (monorepo root). All service Dockerfiles that used `--link` on non-base-image COPY stages.

**Marker**: SCAR-001 (explicit in commit `251a5f9` message).

### SCAR-002 — Pydantic forward-reference resolution failure with TYPE_CHECKING + future annotations

**What failed**: `datetime` was imported under `TYPE_CHECKING` guard in four model files (`campaign.py`, `ad.py`, `ad_set.py`, `lead_form.py`). Combined with `from __future__ import annotations` (which delays all annotation evaluation), Pydantic v2's `model_validate()` could not resolve the forward reference at runtime, causing `PydanticUserError`.

**When**: Commit `6073ff3` (2026-03-02), version bump to 0.2.1.

**Fix commit**: `6073ff3` — `fix(autom8y-meta): move datetime to runtime imports for Pydantic compat`

**Fix locations**:
- `src/autom8y_meta/models/campaign.py`
- `src/autom8y_meta/models/ad.py`
- `src/autom8y_meta/models/ad_set.py`
- `src/autom8y_meta/models/lead_form.py`

**Pattern**: Moved `from datetime import datetime` out of `if TYPE_CHECKING:` block to top-level runtime imports in all four files.

### SCAR-003 — CursorPaginator return type mismatch blocking callers

**What failed**: `get_account_campaigns`, `get_campaign_ad_sets`, and `get_ad_set_ads` on `MetaAdsClient` were typed as `-> AsyncIterator[T]` instead of `-> CursorPaginator[T]`. Callers attempting to call `fetch_one_page()` (a `CursorPaginator`-only method) received mypy type errors and could not use the paginator's stateless single-page API.

**When**: Commit `2a0e278` (2026-03-01).

**Fix commit**: `2a0e278` — `fix(meta): use CursorPaginator return type for hierarchy methods`

**Fix location**: `src/autom8y_meta/client.py` (three method signatures, lines ~302-342).

### SCAR-004 — AUTOM8Y_ENV env_prefix shadowing: env var resolved as {PREFIX}AUTOM8Y_ENV

**What failed**: Child `Autom8yBaseSettings` classes with a custom `env_prefix` (e.g., `META_`) looked for `META_AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`. ECS task definitions only set `AUTOM8Y_ENV`, so `autom8y_env` defaulted to `LOCAL`, triggering production URL guards that rejected non-production URLs.

**When**: Commit `1367461` (2026-03-08), autom8y-config v1.2.1.

**Fix commit**: `1367461` — `fix(config): AUTOM8Y_ENV now read regardless of child class env_prefix`

**Fix location**: `sdks/python/autom8y-config/src/autom8y_config/base_settings.py` — added `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` to the `autom8y_env` field. `MetaConfig` inherits this fix via `Autom8yBaseSettings`.

**Affected directly**: `MetaConfig` (META_ prefix), `AuthSettings` (AUTH__ prefix), and any future `Autom8yBaseSettings` subclass with a non-empty `env_prefix`.

### SCAR-005 — pytest --import-mode default causing cross-package import collisions

**What failed**: pytest's default import mode caused module path collisions when the monorepo workspace included multiple packages. Tests from one SDK would shadow or interfere with another's modules during the test collection phase.

**When**: Commit `299aded` (monorepo root).

**Fix commit**: `299aded` — `fix(autom8y): add --import-mode=importlib to root pytest config`

**Fix location**: Root `pyproject.toml` (monorepo) — added `addopts = "--import-mode=importlib"` to `[tool.pytest.ini_options]`. The autom8y-meta `pyproject.toml` inherits from the workspace root.

### SCAR-006 — mypy strict errors from TYPE_CHECKING-only imports in test files

**What failed**: Multiple SDK test suites (including autom8y-meta-adjacent SDKs) accumulated `mypy --strict` failures because imports used in test bodies were placed under `TYPE_CHECKING` for ruff compliance, then referenced at runtime. CI blocked SDK publishes.

**When**: Commits `b2a0c53`, `8a6000f`, `28e1aae` (monorepo-wide).

**Fix commits**:
- `b2a0c53` — `fix(sdk): resolve mypy strict errors in reconciliation, telemetry, and devx-types`
- `8a6000f` — `refactor(sdk): fix manual type errors in test suites for mypy strict`
- `28e1aae` — `refactor(sdk): fix all src/ mypy strict errors and extract shared LoggerProtocol`

**Fix location (autom8y-meta-scoped)**: `tests/` — test files now use `if TYPE_CHECKING:` only for type-narrowing imports, not for runtime dependencies.

### SCAR-007 — DEF-LOW-001-004: assert guards and inefficient queries

**What failed**: A hygiene audit identified assert statements used as runtime guards (which are stripped in optimized mode) and `len(db.exec().all())` patterns that fetch entire result sets to count rows. Tracked as DEF-LOW-001 through DEF-LOW-004.

**When**: Commit `fbdb80d` (2026-02-10).

**Fix commit**: `fbdb80d` — `feat(sdk): expand test coverage and fix small issues (DEF-LOW-001, 002, 003, 004)`

**Fix locations**: Auth service and autom8y-config SDK (not autom8y-meta source directly). The autom8y-meta `client.py` uses `assert` statements post-`_ensure_initialized()` explicitly as mypy type-narrowing hints (documented in the `_ensure_initialized` docstring), not as runtime guards — this is a deliberate exception to the DEF-LOW pattern.

### SCAR-008 — SEC-001: deprecated datetime.utcnow() producing naive datetimes

**What failed**: 220 instances of `datetime.utcnow()` and `datetime.utcfromtimestamp()` across the monorepo produced timezone-naive datetimes. Python 3.12 deprecated `utcnow()`; mixing naive and aware datetimes caused comparison errors (notably in auth service password reset and token expiry). Tagged SEC-001.

**When**: Commit `a7f086c` (2026-02-10).

**Fix commit**: `a7f086c` — `fix(security): replace deprecated datetime.utcnow() with timezone-aware alternative (SEC-001)`

**Fix locations**: autom8y-auth SDK, auth service, pull-payments service. autom8y-meta was not affected (no datetime comparison logic in its source).

## Category Coverage

| Category | Scars |
|---|---|
| Integration failure (API contract / type mismatch) | SCAR-002 (Pydantic forward-ref), SCAR-003 (return type mismatch) |
| Config drift (env var resolution) | SCAR-004 (env_prefix shadowing AUTOM8Y_ENV) |
| Build / CI infrastructure failure | SCAR-001 (Docker --link), SCAR-005 (pytest import mode), SCAR-006 (mypy strict) |
| Security / deprecation | SCAR-008 (datetime.utcnow deprecated) |
| Code quality / hygiene | SCAR-007 (DEF-LOW: assert guards, N+1 count queries) |

Categories observed but not found in autom8y-meta scope:
- **Data corruption**: No evidence of data mutation bugs in this SDK (read/write path is HTTP to external API).
- **Race condition**: Rate limiter uses asyncio.Semaphore + token bucket; no recorded race condition scar.
- **Schema evolution**: No ORM migrations in this SDK; schema is Pydantic model-only.
- **Performance cliff**: No recorded performance regression in autom8y-meta; rate limiting was specified up-front.

5 distinct categories are represented across 8 scars.

## Fix-Location Mapping

| Scar | File(s) | Function / Area | Verified Exists |
|---|---|---|---|
| SCAR-001 | `scripts/templates/Dockerfile.ecs-fargate`, `scripts/templates/Dockerfile.lambda-scheduled` | COPY stage directives | Yes (monorepo root) |
| SCAR-002 | `src/autom8y_meta/models/campaign.py` | Module-level imports | Yes |
| SCAR-002 | `src/autom8y_meta/models/ad.py` | Module-level imports | Yes |
| SCAR-002 | `src/autom8y_meta/models/ad_set.py` | Module-level imports | Yes |
| SCAR-002 | `src/autom8y_meta/models/lead_form.py` | Module-level imports | Yes |
| SCAR-003 | `src/autom8y_meta/client.py` | `get_account_campaigns`, `get_campaign_ad_sets`, `get_ad_set_ads` signatures (~lines 302-342) | Yes |
| SCAR-004 | `sdks/python/autom8y-config/src/autom8y_config/base_settings.py` | `autom8y_env` field `validation_alias` | Yes (monorepo) |
| SCAR-005 | Monorepo root `pyproject.toml` | `[tool.pytest.ini_options]` `addopts` | Yes (monorepo) |
| SCAR-006 | `tests/` | Multiple test files, TYPE_CHECKING imports | Yes |
| SCAR-007 | `sdks/python/autom8y-auth/`, `sdks/python/autom8y-config/` | Auth routes, config audit log | Yes (monorepo) |
| SCAR-008 | `sdks/python/autom8y-auth/src/autom8y_auth/credentials.py`, auth service, pull-payments | Credential expiry, token datetime comparisons | Yes (monorepo) |

Compound fixes: SCAR-001 required changes to both scaffold templates and all derived service Dockerfiles. SCAR-004 required ecosystem-wide version floor bumps in the root `pyproject.toml` `constraint-dependencies`.

## Defensive Pattern Documentation

### SCAR-001 → Defensive pattern: `--link`-free COPY + hash-verified two-step install

The scaffold templates now enforce:
1. No `--link` on `COPY --from=<non-base-stage>` directives.
2. Two-step install: `uv pip compile --generate-hashes` -> `uv pip install --require-hashes` (supply-chain integrity + reproducibility).
3. Self-install uses `--no-deps` to avoid re-resolving the dependency graph.

This pattern is the canonical scaffold going forward. New services generated from templates inherit it. No dedicated regression test — scaffold templates are not unit-tested.

### SCAR-002 → Defensive pattern: runtime datetime imports in Pydantic models

All four model files now import `datetime` at the module level (not under `TYPE_CHECKING`). The `from __future__ import annotations` flag is retained for other forward-reference benefits. The rule is: any type used in a Pydantic model field annotation must be imported at runtime, regardless of `TYPE_CHECKING` guard.

No explicit regression test for this exact failure mode. The existing model instantiation tests (e.g., `tests/models/test_campaign_models.py`) would catch a regression because they call `model_validate()` with actual datetime values.

### SCAR-003 → Defensive pattern: CursorPaginator return type on all paginated hierarchy methods

`client.py` now correctly types `get_account_campaigns`, `get_campaign_ad_sets`, and `get_ad_set_ads` as `-> CursorPaginator[T]` (imported under `TYPE_CHECKING`). The `_ensure_initialized()` pattern with a subsequent `assert` (documented as mypy-narrowing-only) is the canonical guard for all delegating methods.

The `CursorPaginator.fetch_one_page()` method is tested in `tests/test_pagination.py`.

### SCAR-004 → Defensive pattern: AliasChoices for cross-prefix env vars

`autom8y-config` `base_settings.py` now uses `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` on the `autom8y_env` field. This is the canonical pattern for any base field that must be read from a canonical env var name regardless of the child class's `env_prefix`. The autom8y-meta `MetaConfig` inherits this protection.

No autom8y-meta-local regression test. The fix is in autom8y-config v1.2.1; autom8y-meta's `pyproject.toml` pins `autom8y-config>=1.2.1` via the workspace constraint floor.

### SCAR-005 → Defensive pattern: --import-mode=importlib in root pytest config

The root `pyproject.toml` enforces `--import-mode=importlib` globally. This prevents cross-package name collisions in the workspace. autom8y-meta tests inherit this setting.

### SCAR-006 → Defensive pattern: mypy strict mode gated in CI

SDK publish CI (`sdk-publish-v2.yml`) runs `mypy --strict` as a hard gate. Test files use `if TYPE_CHECKING:` only for true type-annotation imports (not for values used at runtime). The `conftest.py` provides shared fixtures (http client, rate limiter, proof generator) to prevent ad-hoc runtime imports in individual test files.

### SCAR-007 → Defensive pattern: RuntimeError replaces assert for runtime guards

`_ensure_initialized()` in `client.py` uses an explicit `if ... raise RuntimeError(...)` (not an assert) as the actual guard. Post-guard `assert` statements are typed as mypy-narrowing-only (documented in the method's docstring). This is the documented exception: `assert` in autom8y-meta client code means "mypy narrowing only, not a runtime guard."

### SCAR-008 → Defensive pattern: timezone-aware datetimes via datetime.now(timezone.utc)

Not directly in autom8y-meta source (no datetime arithmetic in this SDK). The convention is enforced monorepo-wide via ruff's `DTZ` rules and the SEC-001 sweep. autom8y-meta has no datetime mutation code, but any future additions must use `datetime.now(timezone.utc)`.

## Agent-Relevance Tagging

| Scar | Relevant Agents | Why |
|---|---|---|
| SCAR-001 (Docker --link) | `principal-engineer`, `architect` | Any new service Dockerfile or scaffold template modification must not reintroduce `--link` on non-base-image COPY stages. The two-step hash-verified install is required. |
| SCAR-002 (Pydantic forward-ref) | `principal-engineer` | Any new Pydantic model in autom8y-meta that uses `from __future__ import annotations` must import all field types at runtime, not under `TYPE_CHECKING`. This applies to all datetime, Enum, and custom type annotations used in model fields. |
| SCAR-003 (CursorPaginator return type) | `principal-engineer`, `qa-adversary` | New hierarchy traversal methods on `MetaAdsClient` must return `CursorPaginator[T]`, not `AsyncIterator[T]`. The qa-adversary should test `fetch_one_page()` access on all paginated endpoints. |
| SCAR-004 (env_prefix shadowing) | `principal-engineer`, `architect` | Any new `Autom8yBaseSettings` subclass with a non-empty `env_prefix` must not assume canonical env vars (like `AUTOM8Y_ENV`) will be picked up automatically. Use `validation_alias` for shared cross-prefix fields. |
| SCAR-005 (pytest import mode) | `principal-engineer` | If adding a new pytest plugin or modifying test discovery config, preserve `--import-mode=importlib`. Do not remove or override it in per-package `pyproject.toml`. |
| SCAR-006 (mypy strict) | `principal-engineer`, `qa-adversary` | All SDK source files must pass `mypy --strict`. Test files use `TYPE_CHECKING` only for annotation-only imports. Any new test fixtures should be added to `conftest.py`, not inlined as runtime imports inside test methods. |
| SCAR-007 (assert guards) | `principal-engineer` | In autom8y-meta, `assert` after `_ensure_initialized()` is intentional mypy narrowing. In all other contexts (new service code, new handlers), use `if condition: raise RuntimeError(...)` instead of bare `assert`. |
| SCAR-008 (datetime.utcnow) | `principal-engineer` | Use `datetime.now(timezone.utc)` everywhere. Ruff's DTZ rules will catch violations in CI, but the agent should not introduce them in new code. |

Cross-agent scars (relevant to both implementation and review): SCAR-002, SCAR-004.

## Knowledge Gaps

1. **No inline SCAR-/DEF- markers in autom8y-meta source**: The project uses commit message tags for cross-cutting issues. SCAR-001 is the only scar explicitly labeled in a commit message scope for autom8y-meta. Future scars may be harder to surface without a code-level annotation convention.

2. **No autom8y-meta-specific regression test for SCAR-002**: The Pydantic forward-reference failure is guarded only implicitly by existing model instantiation tests.

3. **SCAR-004 fix lives in autom8y-config, not autom8y-meta**: The env_prefix shadowing bug is fixed upstream. The protection depends on the workspace constraint floor (`autom8y-config>=1.2.1`).

4. **Rate limiter `_requests_waiting` counter is not thread-safe**: The `MetaRateLimiter` uses a plain integer for `_requests_waiting` with non-atomic increment/decrement. In asyncio this is safe (single-threaded event loop), but the code has no comment documenting this assumption. If the SDK were used with `asyncio.run_in_executor` or a thread pool, this would be a race condition. This is a potential-but-unconfirmed scar.
