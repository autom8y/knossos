---
domain: scar-tissue
generated_at: "2026-03-16T00:02:18Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog Completeness

### SCAR-001: ResilientCoreClient Construction API Mismatch

**What failed**: The fetcher initially constructed `ResilientCoreClient` using invalid kwargs (`service_api_key`, `auth_url`) passed directly to the constructor. The actual SDK API requires a three-step chain: `Config() -> Client.from_config() -> ResilientCoreClient.wrap()`.

**When**: Commit `44ce81f` (2026-03-02), discovered during local dry run via Lambda RIE.

**How fixed**: Rewrote client construction in `fetch_ads_tree` and `fetch_asana_offers` to use the correct three-call chain.

**Fix location**: `src/reconcile_ads/fetcher.py`, lines 83-86 and 192-195.

---

### SCAR-002: InteropAdsError Attribute Name Wrong (`service` vs `service_name`)

**What failed**: Error translation code in `fetcher.py` accessed `exc.service` but the base `ServiceError` class stores the value as `.service_name`. Produced `AttributeError` on circuit-open paths.

**When**: Commit `44ce81f` (2026-03-02), same dry run as SCAR-001.

**How fixed**: Corrected attribute access from `exc.service` to `exc.service_name` in the `except InteropAdsError` clause.

**Fix location**: `src/reconcile_ads/fetcher.py`, lines 132-137 (ads) and 282-287 (asana).

---

### SCAR-003: `classification` Included in `select_fields` Causing 422

**What failed**: `classification` was listed in the `select_fields` array passed to `RowsRequest`. The autom8y-asana API uses `extra="forbid"` on `RowsRequest`, so `classification` (which is a top-level filter parameter, not a DataFrame column) caused a `422 UNKNOWN_FIELD` response on every offer fetch — breaking all reconciliation runs.

**When**: Commit `f98c5c6` (2026-03-02).

**How fixed**: Removed `"classification"` from the `select_fields` list. Classification is now injected client-side after the fetch (see defensive pattern DEF-001 below).

**Fix location**: `src/reconcile_ads/fetcher.py`, lines 179-189 (current `select_fields` list), lines 216-219 (client-side injection).

---

### SCAR-004: `AUTOM8Y_ENV` Not Resolved by Child Config Classes

**What failed**: Child `Settings` classes with custom `env_prefix` (e.g., `ADS_`) looked for `ADS_AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`. Lambda task definitions only set `AUTOM8Y_ENV`, so `autom8y_env` always defaulted to `LOCAL` in production, triggering the production URL guard incorrectly.

**When**: Commit `1367461` (ecosystem-wide fix). The `reconcile-ads` `Settings` class uses `env_prefix=""` so was less directly affected, but depends on the fixed SDK.

**How fixed**: Added `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` to the base field in `autom8y-config`. The `reconcile-ads` `Settings` class now carries an explicit `validation_alias=AliasChoices("AUTOM8Y_ENV")` on the field (line 117 of `config.py`).

**Fix location**: `src/reconcile_ads/config.py`, line 115-118.

---

### SCAR-005: `pytest-asyncio` Breaking API Change Blocked CI

**What failed**: `pytest-asyncio` 1.0 introduced a breaking API change causing `'Package' object has no attribute 'obj'` during test collection. This blocked the SDK Publish CI workflow, preventing all service deployments including `reconcile-ads`.

**When**: Commit `03f8175` (ecosystem-wide fix).

**How fixed**: Pinned `pytest-asyncio>=1.2,<2.0` across all 20 `pyproject.toml` files in the monorepo.

**Fix location**: `pyproject.toml` (dev dependency group).

---

### SCAR-006: Mypy Type Errors — `cast()` and `asyncio.gather` Result Narrowing

**What failed**: mypy reported type errors in `fetcher.py` because (a) the ternary branch producing `response_dict` used `cast(dict[str, Any], response)` instead of `cast("dict[str, Any]", response)`, and (b) `asyncio.gather` result handling used indexing without `isinstance` guard, preventing mypy from narrowing away exception types.

**When**: Commit `3df641d` (multi-service fix).

**How fixed**: Quoted the type expression in `cast()` (ruff TC006 compliance). Changed `asyncio.gather` result handling to use `isinstance(result, Exception)` guards.

**Fix location**: `src/reconcile_ads/fetcher.py`, line 97 (`cast("dict[str, Any]", response)`), lines 324-336 (isinstance guards in `fetch_all`).

---

### SCAR-007: Ruff TC006 — Quoted Type in `typing.cast`

**What failed**: Ruff lint rule TC006 flagged unquoted type expressions in `typing.cast()` calls used with `TYPE_CHECKING`-guarded imports at runtime. Blocked CI linting step.

**When**: Commit `2224a49`.

**How fixed**: Quoted the type string in `cast()`.

**Fix location**: `src/reconcile_ads/fetcher.py`, line 97.

---

### SCAR-008: Settings Cache Poisoning Between Tests (SCAR-011 in code)

**What failed**: The `@lru_cache` on `get_settings()` retained stale `Settings` objects across tests when environment variables changed. Tests that mutated env vars would pollute the settings singleton for subsequent tests, producing incorrect behavior or false passes.

**When**: Recorded as SCAR-011 in the codebase; materialized during test development.

**How fixed**: Autouse fixture `_clear_settings` in `conftest.py` calls `clear_settings_cache()` before and after each test.

**Fix location**: `tests/conftest.py`, lines 32-46.

---

### SCAR-009: Fetcher Error Double-Wrap Risk (Bug 2 Regression Guard)

**What failed**: When the fetcher's own `AdsServiceUnavailableError`/`AsanaServiceUnavailableError` is raised inside the inner `try` block (e.g., by a nested call), the generic `except Exception` clause at the bottom would catch it and wrap it in a new error instance, losing circuit-breaker context (`time_remaining`, `service_name`, `method`).

**When**: Adversarial test added in commit `0fd47c4`. Labeled "Bug 2 regression guard" in `test_adversarial.py` line 1252.

**How fixed**: Added explicit re-raise guards (`except AdsServiceUnavailableError: raise` and `except AsanaServiceUnavailableError: raise`) before the generic `except Exception` fallback.

**Fix location**: `src/reconcile_ads/fetcher.py`, lines 138-139 (ads) and 288-289 (asana).

---

### SCAR-010: Env Var Canonical Name Drift (DEF-005 through DEF-009)

**What failed**: Multiple services used legacy env var names (`AUTH_BASE_URL`, `DATA_SERVICE_URL`, `ADS_ENVIRONMENT`) instead of canonical Tier 3 names (`AUTOM8Y_AUTH_URL`, `AUTOM8Y_DATA_URL`, `AUTOM8Y_ENV`). In Lambda, only the canonical names are injected, causing config fields to fall back to wrong defaults.

**When**: Commits `35f29a1` and `5f94481` (stakeholder-validated corrections).

**How fixed**: All SDK clients and service configs updated to read canonical env var names. Legacy `_LEGACY_ENV_MAP` bridge in `autom8y-auth` removed entirely.

**Fix location** (reconcile-ads): No direct source change — the fix lives in `autom8y-config` SDK and `autom8y-core` SDK. `reconcile-ads` was a downstream consumer that inherited the fix via SDK version bumps.

---

## Category Coverage

| Category | Scars | Examples |
|----------|-------|---------|
| **API contract violation** | SCAR-001, SCAR-002, SCAR-003 | Wrong constructor kwargs, wrong error attribute name, forbidden field in request |
| **Type system / linting** | SCAR-006, SCAR-007 | mypy cast narrowing, ruff TC006 quoted type |
| **Configuration / env resolution** | SCAR-004, SCAR-010 | `AUTOM8Y_ENV` prefix mismatch, legacy env var names in SDK clients |
| **Test infrastructure** | SCAR-005, SCAR-008 | pytest-asyncio breaking API, settings cache poisoning |
| **Defensive error handling** | SCAR-009 | Error double-wrap suppression |

Five distinct failure categories covered. Categories span the full surface from API boundary through config resolution to CI infrastructure.

---

## Fix-Location Mapping

| Scar | File | Lines |
|------|------|-------|
| SCAR-001 | `src/reconcile_ads/fetcher.py` | 83-86, 192-195 |
| SCAR-002 | `src/reconcile_ads/fetcher.py` | 132-137, 282-287 |
| SCAR-003 | `src/reconcile_ads/fetcher.py` | 179-189 (select_fields), 216-219 (client-side inject) |
| SCAR-004 | `src/reconcile_ads/config.py` | 115-118 |
| SCAR-005 | `pyproject.toml` | dev dependency group |
| SCAR-006 | `src/reconcile_ads/fetcher.py` | 97, 324-336 |
| SCAR-007 | `src/reconcile_ads/fetcher.py` | 97 |
| SCAR-008 | `tests/conftest.py` | 32-46 |
| SCAR-009 | `src/reconcile_ads/fetcher.py` | 138-139, 288-289 |
| SCAR-010 | SDK-layer (`autom8y-config`, `autom8y-core`); inherited via version bumps | N/A in service source |

---

## Defensive Pattern Documentation

### DEF-001: Client-Side `classification` Injection

Born from SCAR-003. After the fix removed `classification` from `select_fields`, the fetcher now makes two separate `query_rows` calls (one with `classification="active"`, one with `classification="activating"`) and injects the classification string into each row dict client-side. This avoids any API contract violation while preserving the data that downstream rules need.

Location: `src/reconcile_ads/fetcher.py`, lines 215-219.

```python
for row in active_result.data:
    row["classification"] = "active"
for row in activating_result.data:
    row["classification"] = "activating"
```

---

### DEF-002: Error Re-Raise Guards Before Generic Catch

Born from SCAR-009. Both `fetch_ads_tree` and `fetch_asana_offers` have a three-layer exception structure: (1) interop-specific error translated, (2) service's own error re-raised explicitly, (3) generic fallback. The explicit re-raise at layer 2 prevents the generic catch from wrapping already-typed errors.

Location: `src/reconcile_ads/fetcher.py`, lines 132-145 and 282-295.

---

### DEF-003: `_safe_int()` for MagicMock / Unexpected Upstream Types

Born from adversarial testing experience. The helper `_safe_int()` coerces values from upstream responses to `int | None`, guarding against `MagicMock` objects or unexpected float types leaking into metadata. Its docstring explicitly names this protection.

Location: `src/reconcile_ads/fetcher.py`, lines 22-31.

---

### DEF-004: `clear_settings_cache` Autouse Fixture

Born from SCAR-008 (SCAR-011 in test code). `conftest.py` has an autouse fixture that calls `clear_settings_cache()` before and after every test, preventing lru_cache poisoning. The `clear_settings_cache()` function also calls `Autom8yBaseSettings.reset_resolver()` to clear any SDK-level state.

Location: `tests/conftest.py`, lines 36-46; `src/reconcile_ads/config.py`, lines 129-132.

---

### DEF-005: Unknown Classification Treated as Inactive (Fail-Safe)

Born from SCAR-003 context: when `classification` injection fails or returns an unexpected value, `rule_status_alignment` in `rules.py` uses `(offer.classification or "").strip().lower()` for safe normalization, then explicitly logs `unknown_classification` and falls through to the inactive/ghost-detection path.

Location: `src/reconcile_ads/rules.py`, lines 70, 88-95.

---

### DEF-006: Campaign/Ad Group Name Decode Field Padding

Born from production encoding assumptions. Both `_decode_campaign_name` and `_decode_ad_group_name` pad the decoded `parts` list with empty strings when fewer fields than expected are present, rather than raising. This prevents `IndexError` on malformed or legacy campaign names.

Location: `src/reconcile_ads/joiner.py`, lines 77-79, 89-91.

---

### DEF-007: `asyncio.gather(return_exceptions=True)` with `isinstance` Guards

Born from SCAR-006. `fetch_all` uses `return_exceptions=True` so that one upstream source failing does not cancel the other task's gather. Results are then checked with `isinstance(result, Exception)` before unpacking. This allows the orchestrator to surface which source failed independently.

Location: `src/reconcile_ads/fetcher.py`, lines 322-336.

---

## Agent-Relevance Tagging

| Scar | Relevant Agents / Areas |
|------|------------------------|
| SCAR-001 (ResilientCoreClient API) | **principal-engineer**: Any code that constructs interop clients must use `Config -> Client.from_config -> ResilientCoreClient.wrap()` chain. Never pass `service_api_key` or `auth_url` directly to `ResilientCoreClient`. |
| SCAR-002 (`.service` vs `.service_name`) | **principal-engineer**, **hallucination-hunter**: Always use `.service_name` on `ServiceError` subclasses from `autom8y-interop`. |
| SCAR-003 (`classification` in `select_fields`) | **principal-engineer**: `RowsRequest` uses `extra="forbid"`. Filter parameters (`classification`) are NOT selectable columns. |
| SCAR-004 (env_prefix + `AUTOM8Y_ENV`) | **principal-engineer**, **qa-adversary**: Any `Settings` subclass with a non-empty `env_prefix` must add `validation_alias=AliasChoices("AUTOM8Y_ENV")` explicitly or inherit from the fixed base class. |
| SCAR-005 (pytest-asyncio pin) | **principal-engineer**: Do not unpin `pytest-asyncio` above `<2.0`. The 1.x -> 2.x boundary has breaking changes. |
| SCAR-006/007 (mypy/ruff typing) | **principal-engineer**: Type expressions in `cast()` must be quoted strings when used with `TYPE_CHECKING`-gated imports. |
| SCAR-008 (settings cache poisoning) | **qa-adversary**: All test suites for services using `@lru_cache` on settings must include the `_clear_settings` autouse pattern. |
| SCAR-009 (error double-wrap) | **principal-engineer**: Exception handler chains must re-raise the service's own error types before the generic `except Exception` fallback. |
| SCAR-010 (env var canonical names) | **principal-engineer**: Use `AUTOM8Y_AUTH_URL`, `AUTOM8Y_DATA_URL`, `AUTOM8Y_ENV` as canonical names. `AUTH_BASE_URL`, `DATA_SERVICE_URL` are dead. |

---

## Knowledge Gaps

1. **No inline `SCAR-` markers in source code**: The only in-source `SCAR-` reference is `SCAR-011` in `conftest.py`. All other scars were reconstructed from git commit messages. The scar numbering system is not consistently applied across source files.

2. **SCAR-010 fix location is SDK-layer only**: The env var canonicalization fix lives in `autom8y-config` and `autom8y-core` SDKs (not in this repo). The precise lines in those SDKs are not captured here.

3. **Truncation detection feature (commit `2cdcff3`)**: The commit adding truncation detection logic is documented in DEF-001 context but not labeled as a distinct scar.

4. **No test names explicitly labeled as regression guards** except the two identified (`test_missing_classification_treated_as_ghost`, `TestFetcherErrorReraiseAdversarial`). Broader `test_adversarial.py` coverage intent is not labeled per-scar.
