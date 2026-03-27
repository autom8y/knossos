---
domain: scar-tissue
generated_at: "2026-03-16T00:04:02Z"
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

**Service:** `pull-payments`
**Source tree:** `services/pull-payments/`
**Git commits analyzed:** 74 total, 30 scar-relevant (fix/bug/hotfix/revert)

---

## Failure Catalog Completeness

All scars identified from `git log` filtered for fix/bug/regression/revert/hotfix keywords, cross-referenced against code comments and test names. No SCAR-NNN, BUG-, WORKAROUND, or HACK markers found in source — this codebase uses conventional commit fix: prefixes and inline comments as its primary scar-marking system.

### SCAR-001: Stripe API v2025 Removed `invoice` Field from Charges

**Commit:** `5b81aeb` — 2026-02-02
**Symptom:** Invoice-backed refunds were categorized as "other" instead of by product type. Root cause: Stripe API v2025-12-15.clover removed the `invoice` field from the charge object, so `_resolve_invoice_id()` always returned `None`.
**Fix:** Implemented invoice resolution via the `invoice_payments` API — lookup by `payment_intent_id` instead of reading `charge.invoice`. SDK bumped from `autom8y-stripe 1.0.1` to `1.1.0`.
**Location:** `src/pull_payments/orchestrator.py` — refund sync path

### SCAR-002: Non-Succeeded Refunds Incorrectly Processed

**Commit:** `5b81aeb` — 2026-02-02
**Symptom:** Refunds with `status != "succeeded"` (pending, failed, canceled) were passed through the sync pipeline and written as financial records.
**Fix:** Added `if refund.status != "succeeded": continue` guard at line 403 of `orchestrator.py`.
**Location:** `src/pull_payments/orchestrator.py:403`

### SCAR-003: Inclusive Upper Date Bound Caused Cross-Month Duplicate Records

**Commit:** `2fbc1a9` — 2026-02-02
**Symptom:** Invoice, refund, and charge handlers used `lte` (inclusive) for the date range upper bound. Records created at midnight month boundaries appeared in two consecutive sync windows and survived the delete step of the earlier month, producing duplicate payment records during backfills.
**Fix:** Changed all handlers to use `lt` (exclusive). `autom8y-stripe` bumped to `1.1.1`.
**Location:** `autom8y-stripe` SDK (not pull-payments source directly), but fix required in `pyproject.toml` pin update.

### SCAR-004: `refunds_skipped` Not Tracked — Silent Duplicate Drop

**Commit:** `bf58f43` — 2026-02-02
**Symptom:** `sync_refunds_for_range` only tracked `success_count` from batch results, silently dropping `failure_count`. Duplicate refunds produced warning logs but incremented no counter, making `refunds_skipped` in `SyncResult` only reflect payment duplicates.
**Fix:** Added `refunds_skipped: int = 0` field to `SyncResult`. Return signature changed from 3-tuple to 4-tuple `(processed, written, skipped, not_found)`.
**Location:** `src/pull_payments/models.py:17` and `orchestrator.py` sync_refunds_for_range return tuple

### SCAR-005: Business Lookup Error Handling Wrong — 471 False ERROR Logs/Month

**Commit:** `8220079` — 2026-01-30
**Symptom:** Data service returns `200 + data=None` for not-found businesses, but code was handling `404` as not-found. `400` validation errors were logged at ERROR level. Produced ~471 false ERROR logs per month and a dead `404` branch.
**Fix:** Handle `200 + data=None` as not-found. Log `400` at DEBUG as `business_lookup_rejected`. Keep `500+` as ERROR. Remove dead `404` branch. Rename log event `unmatched_payment_write` → `business_not_found_for_invoice` at INFO.
**Location:** `src/pull_payments/clients/data_service.py` — `get_business_by_stripe_id()`

### SCAR-006: `BatchResultItem.error` Non-Optional — 100 ValidationErrors on Backfill

**Commit:** `1fbed59` — 2026-03-03
**Symptom:** Data service returns `null` for `error` on successful batch items. `BatchResultItem.error` was declared non-optional (`error: BatchErrorDetail`), causing 100 `ValidationError`s per backfill run. Code then attempted `item.error.code` unconditionally, which would have also raised `AttributeError` for `None`.
**Fix:** Changed to `error: BatchErrorDetail | None = None`. Added `if item.error:` guard before logging.
**Location:** `src/pull_payments/clients/models.py:32`

### SCAR-007: `BusinessInfo` and `BatchResult` Were Empty Placeholder Models — 8 mypy Errors Blocking CI

**Commit:** `8920edd` — 2026-03-03
**Symptom:** `BusinessInfo` and `BatchResult` in `autom8y-interop` were empty placeholder models (`extra="allow"` only, no declared fields). This produced 8 mypy errors in pull-payments blocking CI deployment.
**Fix:** Graduated both to declared-field models. `DataServiceClient` return type updated to `BusinessInfo`. `replay.py` TYPE_CHECKING import updated.
**Location:** `autom8y-interop` SDK (fix) + `src/pull_payments/clients/data_service.py` (consumer update)

### SCAR-008: `Client.from_env()` Bypassed Lambda Secrets Extension — Every Lambda Invocation Failed

**Commit:** `0734e2b` — 2026-02-18
**Symptom:** `Client.from_env()` reads `SERVICE_API_KEY` directly from `os.environ`. In Lambda, the secrets extension sets `SERVICE_API_KEY_ARN` (not the raw value) and resolves it via HTTP. `from_env()` bypassed that path entirely, causing both services to fail on every scheduled Lambda invocation since deployment.
**Fix:** Construct `autom8y_core.Config` from the already-resolved `settings.service_api_key` and use `Client.from_config()` instead.
**Location:** `src/pull_payments/clients/data_service.py` — `DataServiceClient.__init__()`

### SCAR-009: `AUTOM8Y_ENV` Not Read When Child Class Uses Custom `env_prefix`

**Commit:** `1367461` — 2026-03-08
**Symptom:** Child `Settings` classes with custom `env_prefix` looked for `{PREFIX}AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`. In ECS task definitions where only `AUTOM8Y_ENV` is set, `autom8y_env` defaulted to `LOCAL`, triggering the production URL guard and blocking the service.
**Fix:** Added `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` to the base field so pydantic-settings always reads the canonical env var. `autom8y-config` bumped to `v1.2.1`.
**Location:** `src/pull_payments/config.py:105-108`

### SCAR-010: Proportional Refund Split Emitted Zero-Amount Records

**Commit:** `cdbd929` — 2026-02-19 (RS-M01)
**Symptom:** Integer floor division in `categorize_refund()` Case 3 (proportional split) produced `share=0` for small line items. These zero-amount payment records were written to the financial batch — semantically empty data with correct totals.
**Fix:** Added `if share == 0: continue` guard after share computation at `orchestrator.py:303`.
**Location:** `src/pull_payments/orchestrator.py:302-304`

### SCAR-011: Docker `COPY --link` for uv Shadowed `/bin/sh` — Container Build Failures

**Commit:** `eb77ac4` — 2026-02-26
**Symptom:** `COPY --link --from=uv /uv /uvx /bin/` creates independent layers that shadow `/bin/` from the base image, making `/bin/sh` inaccessible. Container builds failed at runtime.
**Fix:** Remove `--link` from the uv `COPY` instruction only. Secrets extension `COPY --link` retained (different destination path).
**Location:** `Dockerfile:54`

### SCAR-012: CodeArtifact Index URL Not Passed to `pip install` — Private Packages Unresolvable

**Commit:** `48e37d4` — 2026-02-26
**Symptom:** `uv pip install --require-hashes` could not resolve private `autom8y-*` packages without the CodeArtifact index URL. Applied to all services.
**Fix:** Added conditional `EXTRA_INDEX_URL` injection in Dockerfile `RUN` step.
**Location:** `Dockerfile:81-88`

### SCAR-013: Lambda Layer Binary Named `bootstrap` Not `aws-parameters-and-secrets-lambda-extension`

**Commit:** `c8f7b52` — 2026-02-17
**Symptom:** Dockerfile `COPY` path expected `extensions/aws-parameters-and-secrets-lambda-extension` but Lambda Layer v12 ZIP contains `extensions/bootstrap`. Caused secrets extension to never load.
**Location:** `Dockerfile` — secrets extension COPY step

### SCAR-014: Lambda Layer ARN Contained Wrong Account ID in CI

**Commit:** `531df6e` — 2026-02-17
**Symptom:** CI/CD workflow had hardcoded wrong AWS account ID in the Lambda Layer ARN for the secrets extension. Extension download failed in CI.

### SCAR-015: `assert` Statements in Production Code — Disabled by Python Optimization

**Commit:** `d2b0b5c` — 2026-02-10 (DEF-MOD-ASSERT)
**Symptom:** `autom8y-stripe` client used `assert` for runtime invariants. `except Exception` in pull-payments orchestrator was too broad, masking non-Stripe errors.
**Fix:** Replaced 21 asserts with `RuntimeError`. Narrowed `except Exception` to `except (StripeError, ConnectionError, TimeoutError)`.
**Location:** `src/pull_payments/orchestrator.py:335-340` and `486-491`

### SCAR-016: `datetime.utcnow()` Used Throughout — Timezone-Naive Datetimes (SEC-001)

**Commit:** `a7f086c` — 2026-02-10
**Symptom:** Deprecated `datetime.utcnow()` used across pull-payments scripts. Python 3.12 deprecation warning; also semantically incorrect for timezone-aware operations.
**Fix:** Replaced with `datetime.now(UTC)` throughout scripts. Added ruff `SIM117` per-file-ignore for test files.
**Location:** `scripts/dry_run.py` and related scripts

### SCAR-017: W3C `traceparent` Not Injected into Outbound HTTP — Broken Distributed Traces (DC-7)

**Commit:** `d4996f0` — 2026-03-15
**Symptom:** `ResilientCoreClient` did not inject W3C `traceparent`/`tracestate` headers into outbound requests. Cross-service spans (SMS → Scheduling boundary) were not correlated in the dev console.
**Fix:** Added `_inject_trace_context()` to `ResilientCoreClient` in `autom8y-http`. Gated on OTel availability, no-ops silently on exception.
**Location:** `autom8y-http` SDK (fix). Pull-payments consumer via `DataServiceClient._trace_headers()` at `src/pull_payments/clients/data_service.py:56-64`

### SCAR-018: `pytest-asyncio` Breaking Change in v1.0 — `'Package' object has no attribute 'obj'`

**Commit:** `03f8175` — 2026-03-06
**Symptom:** `pytest-asyncio` 1.0 introduced breaking API changes causing `'Package' object has no attribute 'obj'` during test collection. Blocked the SDK Publish CI workflow, preventing all service deployments.
**Fix:** Pinned `pytest-asyncio>=1.2,<2.0` across all pyproject.toml files.
**Location:** `pyproject.toml` dev dependencies

### SCAR-019: ADOT Metrics Duplicate Timestamps from ECS Task Replicas

**Commit:** `835e57e` — 2026-02-12
**Symptom:** Multiple ECS task replicas emitted metrics with identical labels, causing AMP 400 errors from duplicate timestamps.
**Fix:** Added `instance` label using `HOSTNAME` env var to ADOT Collector scrape config.
**Location:** ADOT/Terraform config (not pull-payments source directly)

---

## Category Coverage

Failure modes classified across 7 distinct categories:

| Category | Scars | Notes |
|----------|-------|-------|
| **Stripe API Contract Drift** | SCAR-001, SCAR-002, SCAR-003 | API version upgrade removed field; status filtering; boundary semantics |
| **Data Model / Type System** | SCAR-006, SCAR-007, SCAR-010 | Optional fields, placeholder models, integer arithmetic edge cases |
| **Infrastructure / Secrets** | SCAR-008, SCAR-009, SCAR-013, SCAR-014 | Lambda extension bypass, env prefix bug, binary name mismatch, ARN |
| **Observability / Logging** | SCAR-004, SCAR-005, SCAR-017, SCAR-019 | Silent counter drop, false ERROR logs, missing trace propagation, metric label collision |
| **Build / CI / Toolchain** | SCAR-011, SCAR-012, SCAR-018 | Docker layer shadowing, private registry, pytest version pin |
| **Security / Hygiene** | SCAR-015, SCAR-016 | Assert in production, timezone-naive datetimes |
| **Developer Experience** | (SCAR-003 also) | Backfill correctness — dev tooling misleading |

**Categories searched but not found in this service:** data-race/concurrency bugs, OOM, memory leaks, SQL/database bugs (service has no direct DB access).

---

## Fix-Location Mapping

| Scar | Primary File | Function / Line | Secondary Location |
|------|-------------|-----------------|-------------------|
| SCAR-001 | `orchestrator.py` | `sync_refunds_for_range()` — invoice resolution | `autom8y-stripe` SDK |
| SCAR-002 | `orchestrator.py:403` | `sync_refunds_for_range()` — status guard | — |
| SCAR-003 | `autom8y-stripe` SDK | invoice/refund/charge handlers | `pyproject.toml` pin |
| SCAR-004 | `models.py:17`, `orchestrator.py` | `SyncResult.refunds_skipped`; return tuple | — |
| SCAR-005 | `clients/data_service.py` | `get_business_by_stripe_id()` | `orchestrator.py` log event rename |
| SCAR-006 | `clients/models.py:32` | `BatchResultItem.error` field | `clients/data_service.py` `if item.error:` guard |
| SCAR-007 | `autom8y-interop` SDK | `BusinessInfo`, `BatchResult` models | `clients/data_service.py` return type |
| SCAR-008 | `clients/data_service.py` | `DataServiceClient.__init__()` | `autom8y-core` SDK |
| SCAR-009 | `config.py:105-108` | `Settings.autom8y_env` field alias | `autom8y-config` SDK v1.2.1 |
| SCAR-010 | `orchestrator.py:302-304` | `categorize_refund()` Case 3 loop | — |
| SCAR-011 | `Dockerfile:54` | `COPY --from=uv` layer | — |
| SCAR-012 | `Dockerfile:81-88` | `RUN uv pip install` step | — |
| SCAR-013 | `Dockerfile` | secrets extension `COPY` path | — |
| SCAR-014 | CI workflow | Lambda Layer ARN | — |
| SCAR-015 | `orchestrator.py:335-340`, `486-491` | `except (StripeError, ConnectionError, TimeoutError)` | `autom8y-stripe` SDK |
| SCAR-016 | `scripts/dry_run.py` | `chunk_date_range()` | `scripts/inspect_refunds.py` |
| SCAR-017 | `clients/data_service.py:56-64` | `DataServiceClient._trace_headers()` | `autom8y-http` SDK |
| SCAR-018 | `pyproject.toml` | `[tool.uv.dev-dependencies]` pytest-asyncio pin | — |
| SCAR-019 | ADOT/Terraform | scrape config | — |

All paths verified to exist in current tree. SCAR-003, SCAR-007, SCAR-019 fixes live in SDK/infra repos but are consumed here via pinned versions.

---

## Defensive Pattern Documentation

### Active Guards in Source

**SCAR-002 guard — refund status filter**
- File: `src/pull_payments/orchestrator.py:402-409`
- Pattern: `if refund.status != "succeeded": continue` with DEBUG log
- Regression test: `test_non_succeeded_refunds_skipped` in `tests/test_orchestrator.py:545`

**SCAR-005 guard — business lookup 200+null vs 400 vs 500+**
- File: `src/pull_payments/clients/data_service.py` — `get_business_by_stripe_id()`
- Pattern: explicit status branch tree; `return None` on `200+data=None`; DEBUG on `400`; ERROR on `500+`
- Regression tests: `test_get_business_by_stripe_id_not_found`, `test_get_business_by_stripe_id_bad_request`, `test_get_business_by_stripe_id_server_error` in `tests/test_data_service.py`

**SCAR-006 guard — optional error field**
- File: `src/pull_payments/clients/models.py:32`
- Pattern: `error: BatchErrorDetail | None = None` + `if item.error:` before attribute access
- Regression test: `test_optional_fields_default_to_none` in `tests/test_client_models.py:56`; `test_accepts_minimal_success` at line 88

**SCAR-010 guard — zero-amount proportional share skip**
- File: `src/pull_payments/orchestrator.py:302-304`
- Pattern: `if share == 0: continue` after floor division
- Regression tests: `test_proportional_skips_zero_amount_records` (line 1125), `test_proportional_all_small_items_collapse_to_largest` in `tests/test_orchestrator.py`

**SCAR-015 guard — narrowed exception catch**
- File: `src/pull_payments/orchestrator.py:335-340`, `486-491`
- Pattern: `except (StripeError, ConnectionError, TimeoutError)` instead of bare `except Exception`

**SCAR-016 guard — timezone-aware datetimes**
- Files: `scripts/dry_run.py`, `scripts/inspect_refunds.py`
- Pattern: `datetime.now(UTC)` / `datetime.fromtimestamp(ts, tz=UTC)`
- ADR reference: ruff `UP` rule enforcement via pyproject.toml

**SCAR-009 guard — env prefix bypass for AUTOM8Y_ENV**
- File: `src/pull_payments/config.py:105-108`
- Pattern: `validation_alias=AliasChoices("AUTOM8Y_ENV")` on `autom8y_env` field
- Comment at line 104: "Reads AUTOM8Y_ENV directly (bypasses env_prefix per ADR-ENV-NAMING-CONVENTION Decision 1)"

**SCAR-008 guard — secrets-extension-aware client construction**
- File: `src/pull_payments/clients/data_service.py`
- Pattern: `Client.from_config(Config(api_key=settings.service_api_key.get_secret_value()))` — uses resolved settings, never `Client.from_env()`

**SCAR-011 guard — Dockerfile layer ordering**
- File: `Dockerfile:54`
- Pattern: secrets extension COPY retains `--link` (safe, different destination); uv COPY drops `--link`

**Staging legacy /tmp fallback (born from data-loss risk, not a named scar)**
- File: `src/pull_payments/staging.py`
- Pattern: `# Legacy /tmp paths -- retained for migration and fallback` at line 28; `migrate_tmp_to_s3()` at line 244; `list_tmp_staged_files()` at line 292
- Replay phase auto-migrates `/tmp` files to S3 on each invocation

### Scars Without Dedicated Defensive Pattern in Source

- SCAR-003 (date boundary `lte`→`lt`): fix lives in SDK, no local guard possible
- SCAR-004 (refunds_skipped): fixed by adding field + 4-tuple return; no runtime guard needed
- SCAR-017 (trace propagation): defensive guard in `autom8y-http` SDK (no-op on exception); local `_trace_headers()` method isolates concern
- SCAR-018 (pytest-asyncio): version pin in `pyproject.toml`; no runtime guard
- SCAR-019 (ADOT replica labels): fix in infra config; no source guard

### Config URL Guard (Related Defensive Pattern)

- File: `tests/test_config_url_guard.py`
- Tests: `test_dev_env_with_production_data_url_raises`, `test_dev_env_with_production_auth_url_raises`
- Guards against: accidentally hitting production URLs from a LOCAL/DEV environment (escalated by SCAR-009 when env was defaulting to LOCAL)

---

## Agent-Relevance Tagging

| Scar | Primary Agents | Rationale |
|------|---------------|-----------|
| SCAR-001 | **principal-engineer**, **qa-adversary** | Any Stripe integration must use `invoice_payments` API for invoice resolution, not `charge.invoice`. Must be known before touching refund pipeline. |
| SCAR-002 | **principal-engineer** | Refund status filter is not obvious from API docs; must explicitly guard non-succeeded statuses. |
| SCAR-003 | **principal-engineer**, **qa-adversary** | Date range boundary semantics: always use `lt` (exclusive) for upper bounds in Stripe pagination. Backfill tests must span month boundaries. |
| SCAR-004 | **principal-engineer** | Counter completeness: sync functions must track all outcome paths (skipped/failed), not just success. |
| SCAR-005 | **principal-engineer**, **qa-adversary** | Data service protocol: not-found is `200+null`, not `404`. New business lookup code must handle this contract. |
| SCAR-006 | **principal-engineer** | Pydantic model discipline: API response fields that can be `null` must be typed `Optional`. Validate against real API responses before declaring types. |
| SCAR-007 | **principal-engineer** | Interop model discipline: placeholder models with `extra="allow"` block mypy and hide runtime errors. Graduate to declared fields before using in typed return values. |
| SCAR-008 | **principal-engineer**, **architect** | Lambda secrets: never use `Client.from_env()` in Lambda — always go through resolved `Settings`. Platform-wide pattern. |
| SCAR-009 | **architect**, **principal-engineer** | Config inheritance: `env_prefix` on child classes breaks canonical env var reading. `AUTOM8Y_ENV` must always use `AliasChoices` bypass. |
| SCAR-010 | **qa-adversary**, **principal-engineer** | Arithmetic edge cases in financial logic: floor division on small amounts can produce 0; must skip rather than write zero-amount records. |
| SCAR-011, SCAR-012 | **principal-engineer** | Docker/BuildKit patterns: `COPY --link` is unsafe for `/bin/` destinations; CodeArtifact requires explicit index URL. |
| SCAR-013, SCAR-014 | **principal-engineer** | Infra: Lambda extension binary name and ARN are deployment-specific; must be verified against actual ZIP/account. |
| SCAR-015 | **principal-engineer**, **qa-adversary** | Python hygiene: `assert` is stripped in optimized mode; `except Exception` is too broad — these are platform-wide standards. |
| SCAR-016 | **principal-engineer** | Python hygiene: `datetime.utcnow()` is deprecated and semantically wrong for timezone-aware pipelines. Platform-wide standard. |
| SCAR-017 | **architect**, **principal-engineer** | Observability: W3C trace context must be explicitly injected into outbound HTTP calls; it does not propagate automatically through `autom8y-core` clients. |
| SCAR-018 | **principal-engineer** | Dependency management: major version bumps in pytest plugins can break test collection silently; pin upper bounds on test tooling. |
| SCAR-019 | **architect** | Infra/observability: ECS task replicas require unique metric labels; AMP rejects duplicate timestamps. Historical only — fix in ADOT config. |

---

## Knowledge Gaps

1. **No SCAR-NNN naming system in source.** The codebase uses conventional commit `fix:` prefixes and ADR references but no inline `# SCAR-NNN` markers. Scars can only be discovered via `git log` — not by grepping source.

2. **SCAR-003, SCAR-007, SCAR-019 fix locations are outside this service tree.** Their defensive effects are felt here only via SDK version pins. A future regression would not be caught by auditing only this service's source.

3. **No investigation of the `autom8y-stripe` SDK scar history** (SCAR-001, SCAR-002 root causes live there). SDK-level failures that surface in pull-payments are not fully documented from the SDK side.

4. **Staging/replay `/tmp` fallback has no identified originating failure.** The pattern exists as a defensive design decision (FR-12) rather than a documented failure. The migration path (`migrate_tmp_to_s3`) implies prior deployments wrote to `/tmp` before S3 staging existed — those records are the implicit scar.
