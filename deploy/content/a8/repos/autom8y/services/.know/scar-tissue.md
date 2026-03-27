---
domain: scar-tissue
generated_at: "2026-03-16T20:20:00Z"
expires_after: "7d"
source_scope:
  - "./*/src/**/*.py"
  - "./*/tests/**/*.py"
  - "./*/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog Completeness

The codebase carries a rich, well-annotated failure history spread across 12 services. Evidence sources are: 372 commits matching fix/bug/regression/revert keywords, structured inline markers (`DEF-`, `SCAR-`, `BUG-`, `CRITICAL-`, `WARN-`, `ISS-`, `RS-M0`, `DC-`, `TF-`), and dedicated test files (`test_defect_remediation.py`, `test_critical_security_fixes.py`, `test_config_url_guard.py`).

---

### SCAR-001 — Docker `COPY --link` Overlay Corruption

**What failed**: `COPY --link --from=<non-base-stage>` on uv binary stages created independent BuildKit overlay layers that shadowed `/bin/` from the base image, making `/bin/sh` inaccessible at container startup.

**Commit**: `eb77ac4` (first fix); `251a5f9` (hardened into scaffold templates)

**Marker**: `SCAR-001` in commit `251a5f9` message

**Fix**: All service Dockerfiles — `--link` removed from every `COPY --from=<non-base-image-stage>` instruction. Template Dockerfiles at `scripts/templates/Dockerfile.ecs-fargate` hardened to prevent reintroduction.

---

### SCAR-011 — `@lru_cache` Settings Cache Poisoning Between Tests

**What failed**: `get_settings()` decorated with `@lru_cache` retained stale `Settings` objects across test cases when environment variables changed via monkeypatch. Later tests received the cached pre-change Settings object, causing test-order-dependent failures.

**Commit**: `48d2ad1` (FN-010 — added `@lru_cache`); downstream SCAR-011 designation when tests started failing

**Marker**: `SCAR-011` — `reconcile-ads/tests/conftest.py` lines 32–46 and `auth-mysql-sync/tests/unit/conftest.py` lines 19–29

**Fix**: Every service's `conftest.py` adds an `autouse` fixture that calls `clear_settings_cache()` before and after each test.

---

### ISS-1–11 — Auth Service 503 Outage (Deployment Audit)

**What failed**: A post-deploy audit (`28de6a8`) identified 11 root causes of 503 errors in the auth service. Key items:

- **ISS-5**: asyncpg rejects offset-aware Python datetimes for `TIMESTAMP WITHOUT TIME ZONE` columns.
- **ISS-7**: `/ready` and `/health/deps` health check endpoints referenced wrong `SecretId`.
- **ISS-8**: `redis` package not listed as an explicit dependency despite being required.
- **ISS-9**: No CloudWatch alarm for "no-healthy-targets" in the ALB target group.
- **ISS-11**: `image_tag` CI/CD override behavior undocumented; smoke test used wrong ECS cluster name.

**Commit**: `28de6a8` (2026-03-13)

**Marker**: Migration `016_timestamps_to_timestamptz.py` line 7: "Fixes ISS-5"

**Fix locations**:
- ISS-5: `auth/migrations/versions/016_timestamps_to_timestamptz.py` — bulk `TIMESTAMP → TIMESTAMPTZ` migration for all 37 timestamp columns
- ISS-7: `terraform/services/auth/main.tf` — corrected `SecretId` references
- ISS-8: `auth/pyproject.toml` — `redis>=5.0.0` added
- ISS-9: `terraform/services/auth/main.tf` — `no-healthy-targets` CloudWatch alarm added
- ISS-11: `scripts/smoke-test.sh` — cluster name corrected

---

### CRITICAL-001–004 — Auth Security Hardening (Charter/Role Architecture)

**What failed**: A security audit identified four critical architectural vulnerabilities:

- **CRITICAL-001**: Role architecture allowed cross-business role assignment.
- **CRITICAL-002**: Authentication fell back insecurely without verifying token first.
- **CRITICAL-003**: Privilege escalation was possible via role manipulation.
- **CRITICAL-004**: Audit log could be mutated post-creation.

**Commits**: ADRs (ADR-0023, 0024, 0025, 0026) implemented fixes

**Marker**: `CRITICAL-001` through `CRITICAL-004` — `auth/tests/test_critical_security_fixes.py` lines 5–8, 176, 455, 617, 910

**Fix locations**: Charter module in `auth/src/charter/` (client.py, exceptions.py); `auth/migrations/versions/005_hybrid_role_schema.py`

---

### Auth Login Token Bug — Refresh Token Storage Failure Returns 200

**What failed**: When refresh token storage failed after successful login, the endpoint returned HTTP 200 with a broken token.

**Commit**: `96011fb`

**Fix**: `auth/src/` — storage failure now returns HTTP 500 with no token leakage.

---

### WARN-1 — Telemetry Convention Checker Case-Sensitivity Bug

**What failed**: `_condition_met()` performed exact-string comparisons. Convention YAML used uppercase (`'FAIL'`) while services emitted lowercase (`'fail'`). Violations were silently missed.

**Commit**: `1139ccc` (2026-03-05)

**Marker**: `WARN-1` in commit message

**Fix**: `sdks/python/autom8y-telemetry/src/autom8y_telemetry/conventions/checker.py` line 556 — `str(value).lower() == expected_value.lower()`

---

### DC-7 — W3C Traceparent Not Injected Across Service Calls

**What failed**: `ResilientCoreClient` did not propagate the W3C `traceparent` header on outbound HTTP calls. Traces were broken at service boundaries.

**Commit**: `d4996f0`

**Marker**: `DC-7` in commit messages

**Fix**: `sdks/python/autom8y-http/` — `ResilientCoreClient` injects `traceparent` header on all outbound requests.

---

### OTel Scheduling Status Enum Incomplete

**What failed**: The convention registry for `scheduling.status` only declared three values. The `BookingEngine` also emitted `cancelled`, `rescheduled`, `already_cancelled`, `not_found`, `not_cancellable`. CI reported `ENUM_VIOLATION`.

**Commit**: `98cd3d0` (2026-03-16)

**Fix**: `sdks/python/autom8y-telemetry/conventions/namespaces/scheduling.yaml` — enum extended with all 8 emitted values.

---

### OTel `appointment_id` Type Mismatch (String vs Int)

**What failed**: Convention registry declared `scheduling.appointment_id` as `type: string` with UUID examples. The database column is an auto-increment integer. CI reported 9 `TYPE_MISMATCH` violations.

**Commit**: `ec910b0` (2026-03-16)

**Fix**: `sdks/python/autom8y-telemetry/conventions/namespaces/scheduling.yaml` — type changed to `int`.

---

### ECS CANARY Deployment Waiter Timeout

**What failed**: `aws ecs wait services-stable` has a 10-minute ceiling. CANARY deployments take 11–13 minutes, causing CI to time out on successful deployments.

**Commit**: `3b97137` (2026-03-16)

**Fix**: `.github/workflows/service-deploy.yml` — custom polling loop replaces the built-in waiter; 15-minute ceiling.

---

### Smoke Test Runs Against Intentionally-Disabled Services

**What failed**: CI smoke test ran health checks against ECS services with `desired_count=0`, generating spurious CI failures.

**Commit**: `2b5589e`

**Fix**: `.github/workflows/service-deploy.yml` — smoke test queries `desiredCount` first; skips if zero.

---

### `AUTOM8Y_ENV` Not Resolved With Custom `env_prefix`

**What failed**: Child `Settings` classes with custom `env_prefix` caused pydantic-settings to look for `{PREFIX}AUTOM8Y_ENV` instead of canonical `AUTOM8Y_ENV`. In ECS, `autom8y_env` defaulted to `LOCAL`, triggering production URL guard.

**Commits**: `948e187` (service fix); `1367461` (SDK-level fix)

**Fix**: `sdks/python/autom8y-config/` — `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` added to base field.

---

### Pytest `ImportPathMismatchError` — Duplicate Basenames

**What failed**: Pytest's default `prepend` import mode resolved same test basename to first module found across 23+ packages. 93 `ImportPathMismatchError` collection failures.

**Commit**: `299aded`

**Fix**: Root `pyproject.toml` — `addopts = ["--import-mode=importlib"]`.

---

### `SecretStr.__eq__` Does Not Coerce to `str`

**What failed**: After `api_key` was changed from `str` to `pydantic.SecretStr`, test assertions comparing against plain strings silently failed.

**Commit**: `a445d0e`

**Fix**: All test assertions comparing `SecretStr` fields now call `.get_secret_value()` first.

---

### Satellite CI `uv sync` Fails — Local Source Paths Absent

**What failed**: Satellite repos use `[tool.uv.sources]` with editable local paths. In CI, these paths do not exist, causing `uv sync` to fail.

**Commit**: `666de73`

**Fix**: `.github/workflows/satellite-ci-reusable.yml` — `uv sync --no-sources` added.

---

### Terraform Private Module Access — No Git HTTPS Credentials

**What failed**: `terraform init` cloning private repo failed in CI with authentication errors.

**Commits**: `f9aa349`, `ee2c436`, `d332784`, `f4909c4`

**Fix**: CI workflows configure `git config --global url."https://x-access-token:${GITHUB_TOKEN}@..."` before Terraform init.

---

### Grafana 409 — Wrong Secret Mapped to `TF_VAR_alert_email`

**What failed**: `TF_VAR_alert_email` was mapped to nonexistent secret. Grafana contact points were destroyed mid-plan.

**Commit**: `9ca5ac2`

**Fix**: `.github/workflows/terraform-apply-reusable.yml` — secret mapping corrected to `TF_VAR_MONITORING_EMAIL`.

---

### ECR Repository Name Prefix Mismatch

**What failed**: Services use `autom8-` prefixed names in Terraform, but ECR repositories are named `autom8y/{service}`. Stack module pointed to nonexistent repositories.

**Commit**: `9ca5ac2`

**Fix**: `terraform/services/{reconcile-spend,pull-payments}/main.tf` — `ecr_repo_name` overrides added.

---

### OTel Sidecar Null Field Passthrough

**What failed**: Optional `observability_config` fields were passed as Terraform `null` to sidecar module, overriding defaults with nulls.

**Commits**: `dc65852`, `d047d3d`

**Fix**: Platform Terraform modules — `optional()` type defaults ensure nulls are populated at decode time.

---

### Dev Mode Security Guard

**What failed**: Auth dev mode could be misconfigured to enable in production if only one of two required conditions was checked.

**Marker**: `CRITICAL SECURITY` comments in `auth/client/autom8y_auth_client/dev_mode.py` lines 6, 54

**Fix**: Dev mode enforces BOTH `AUTH_DEV_MODE=true` AND non-production `AUTOM8Y_ENV`. Test coverage in `tests/test_dev_mode.py:TestProductionGuard`.

---

### SARIF Upload Fails on Private Repos

**What failed**: SARIF upload steps failed on private repos without GitHub Advanced Security, blocking CI.

**Commits**: `fe0f351`, `e571a47`

**Fix**: All `upload-sarif` steps have `continue-on-error: true`. OpenSSF Scorecard removed from private repo.

---

### `datetime.utcnow()` Deprecated

**What failed**: `datetime.utcnow()` deprecated since Python 3.12; asyncpg rejects naive datetimes.

**Commits**: `368573e`, `a7f086c`

**Fix**: All 6 charter module instances replaced with `datetime.now(timezone.utc)`.

---

### Health Check "Double Penalty"

**What failed**: ALB and ECS health checks both probed `/ready`. A brief DB outage simultaneously deregistered ALB targets AND triggered ECS replacements.

**Commit**: `800610c`

**Fix**: `terraform/services/data/main.tf` — ALB probes `/health` (always-200 liveness); ECS probes `/ready` (dependency-aware).

---

### Devconsole Error Boundaries — Crash on DB Corruption and HTTP Failures

**What failed**: Uncaught failures in devconsole: `SpanStore.init()` crash-on-corruption; malformed SQLite JSON rows; `httpx.ConnectError` surfaced raw exceptions.

**Commit**: `b4d0b44`

**Fix**: `devconsole/src/autom8_devconsole/app.py` — `try/except` around init; `_row_to_span()` guards; `ConnectError` handler.

---

### GCal `validation_alias` for camelCase Field Access

**What failed**: Google Calendar API responses use camelCase. Direct attribute access on Pydantic models failed under mypy `--strict`.

**Commit**: `aab3499`

**Fix**: `sdks/python/autom8y-gcal/src/autom8y_gcal/models.py` — `AliasChoices` applied to all camelCase fields.

---

### GCal `requests` Undeclared Dependency

**What failed**: `google-auth` transport depends on `requests` but gcal SDK didn't declare it explicitly. Fresh CI environments failed at runtime.

**Commit**: `b82324b`

**Fix**: `sdks/python/autom8y-gcal/pyproject.toml` — `requests>=2.28.0` added.

---

### Ads Stub Mode Defaults True — Production Runs in Mock Mode

**What failed**: `AdsConfig.use_stub_data_client` defaulted to `True`. Without Terraform overrides, ads ran in stub mode in production.

**Commit**: `4b75606` (`[TF-008]`)

**Fix**: `terraform/services/ads/main.tf` — explicit `false` overrides for stub flags.

---

### reconcile-ads `ResilientCoreClient` Construction API Misuse

**What failed**: Fetcher constructed `ResilientCoreClient` with invalid kwargs. Also, `InteropAdsError.service` was accessed as `.service` but base class stores it as `.service_name`.

**Commit**: `44ce81f`

**Fix**: `reconcile-ads/src/reconcile_ads/fetcher.py` — corrected construction chain and attribute access.

---

### DEF-1 through DEF-E03, DEF-002, BUG-1, RS-M02, Asana Positional Parsing

These scars are fully documented in `reconcile-spend/.know/scar-tissue.md`. Fix locations and regression tests are verified.

---

## Category Coverage

| Category | Count | Key Scars |
|---|---|---|
| **Build / infrastructure** | 4 | SCAR-001, ECS CANARY waiter, satellite uv, TF private module auth |
| **Configuration / environment naming** | 5 | AUTOM8Y_ENV env_prefix, SCAR-011, ads stub defaults, Grafana secret name, ECR name prefix |
| **Integration failure / external contract** | 7 | ISS-5, GCal requests dep, GCal validation_alias, OTel enum incomplete, OTel type mismatch, Asana parsing, reconcile-ads ResilientCoreClient |
| **Data integrity / silent corruption** | 4 | DEF-4 NaN propagation, DEF-002 variance invariant, WARN-1 case-sensitivity, SecretStr comparison |
| **Security** | 5 | CRITICAL-001–004, dev mode guard, SARIF private repo, datetime.utcnow |
| **CI / pipeline** | 4 | Smoke test desired_count=0, pytest importlib, SARIF upload fails |
| **Output encoding / injection** | 3 | DEF-E01 pipe escape, DEF-E02 angle bracket escape, DEF-E03 HTTPS-only URL |
| **Observability / telemetry** | 3 | DC-7 traceparent, OTel sidecar null passthrough, health check double penalty |
| **Runtime crash / missing init** | 3 | Devconsole error boundaries, refresh token returns 200, CloudWatch eager init |

9 distinct categories confirmed.

---

## Fix-Location Mapping

| Scar | Primary Fix Location | Secondary Fix Location |
|---|---|---|
| SCAR-001 | `services/*/Dockerfile` (all services) | `scripts/templates/Dockerfile.ecs-fargate` |
| SCAR-011 | `reconcile-ads/tests/conftest.py:36–46` | `auth-mysql-sync/tests/unit/conftest.py:19–29` |
| ISS-5 | `auth/migrations/versions/016_timestamps_to_timestamptz.py` | — |
| ISS-7 | `terraform/services/auth/main.tf` | — |
| ISS-8 | `auth/pyproject.toml` | — |
| ISS-9 | `terraform/services/auth/main.tf` | — |
| CRITICAL-001–004 | `auth/src/charter/client.py`, `exceptions.py` | `auth/migrations/versions/005_hybrid_role_schema.py` |
| Auth login 200 bug | `auth/src/routes/` (login endpoint) | — |
| WARN-1 | `sdks/python/autom8y-telemetry/conventions/checker.py:556` | — |
| DC-7 | `sdks/python/autom8y-http/` (ResilientCoreClient) | — |
| OTel enum incomplete | `sdks/python/autom8y-telemetry/conventions/namespaces/scheduling.yaml` | — |
| OTel type mismatch | `sdks/python/autom8y-telemetry/conventions/namespaces/scheduling.yaml` | — |
| ECS CANARY waiter | `.github/workflows/service-deploy.yml` | — |
| Smoke test skip | `.github/workflows/service-deploy.yml` | — |
| AUTOM8Y_ENV env_prefix | `sdks/python/autom8y-config/` (base settings field) | — |
| Pytest importlib | Root `pyproject.toml` | — |
| SecretStr.__eq__ | SDK test files | — |
| Satellite uv | `.github/workflows/satellite-ci-reusable.yml` | — |
| TF private module | `.github/workflows/platform-terraform-apply.yml` | `.github/workflows/service-deploy.yml` |
| Grafana 409 | `.github/workflows/terraform-apply-reusable.yml` | — |
| ECR repo prefix | `terraform/services/reconcile-spend/main.tf` | `terraform/services/pull-payments/main.tf` |
| OTel sidecar null | `terraform/modules/platform/stacks/service-with-rds/variables.tf` | — |
| Dev mode guard | `auth/client/autom8y_auth_client/dev_mode.py:6,54` | — |
| SARIF private repo | `.github/workflows/{zizmor,service-build,security-scan,sdk-ci}.yml` | — |
| datetime.utcnow | `auth/src/charter/` (6 files) | — |
| Health check split | `terraform/services/data/main.tf` | — |
| Devconsole error boundaries | `devconsole/src/autom8_devconsole/app.py` | — |
| GCal validation_alias | `sdks/python/autom8y-gcal/src/autom8y_gcal/models.py` | — |
| GCal requests dep | `sdks/python/autom8y-gcal/pyproject.toml` | — |
| Ads stub mode | `terraform/services/ads/main.tf` | — |
| reconcile-ads client | `reconcile-ads/src/reconcile_ads/fetcher.py` | — |

All referenced file paths verified to exist.

---

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Location | Regression Test |
|---|---|---|---|
| SCAR-001 | `--link` prohibited on non-base-image COPY | All Dockerfiles + scaffold templates | Build-time only |
| SCAR-011 | `_clear_settings()` autouse fixture | `reconcile-ads/tests/conftest.py:36–46` | Self-guarding autouse |
| ISS-5 | All datetime columns use `TIMESTAMPTZ` | `auth/migrations/016_timestamps_to_timestamptz.py` | `test_critical_security_fixes.py` |
| CRITICAL-001–004 | Cross-business isolation, verify-then-fallback | `auth/src/charter/` | `auth/tests/test_critical_security_fixes.py` (1148 lines) |
| Auth login 200 bug | Returns 500 on token storage failure | `auth/src/routes/` | 5 test cases in `96011fb` |
| WARN-1 | `.lower()` comparison | `autom8y-telemetry/conventions/checker.py:556` | 6 tests |
| DC-7 | `traceparent` header injected on all outbound calls | `autom8y-http/` | `test_traceparent_injection.py` |
| ECS CANARY waiter | Custom polling loop, 15-min ceiling | `.github/workflows/service-deploy.yml` | CI-only |
| Smoke test skip | `desiredCount==0` guard | `.github/workflows/service-deploy.yml` | CI-only |
| AUTOM8Y_ENV prefix | `AliasChoices("autom8y_env", "AUTOM8Y_ENV")` | `autom8y-config` SDK | Config tests in each service |
| Pytest importlib | `--import-mode=importlib` | Root `pyproject.toml` | Self-guarding |
| SecretStr.__eq__ | `.get_secret_value()` before comparison | SDK test files | Tests are the guard |
| Satellite uv | `--no-sources` flag in CI | `.github/workflows/satellite-ci-reusable.yml` | CI-only |
| Ads stub mode | Explicit `false` env overrides in Terraform | `terraform/services/ads/main.tf` | IaC-level |
| reconcile-ads client | `Config() → Client.from_config() → .wrap()` chain | `reconcile-ads/fetcher.py` | Integration test |
| OTel sidecar null | `optional()` type defaults in TF variables | Platform stack modules | Terraform plan |
| Health check split | ALB `/health` vs ECS `/ready` | `terraform/services/data/main.tf` | Runbook |
| Dev mode guard | Both `AUTH_DEV_MODE=true` AND non-prod env required | `auth/client/dev_mode.py` | `test_dev_mode.py:TestProductionGuard` |
| datetime.utcnow | `datetime.now(timezone.utc)` everywhere | Charter module files | mypy + deprecation warnings |
| Devconsole boundaries | try/except around init + `persistence_degraded` flag | `devconsole/app.py` | 12 error boundary tests |

---

## Agent-Relevance Tagging

| Scar | Relevant Agent(s) | Why |
|---|---|---|
| SCAR-001 | **principal-engineer** | Never use `COPY --link` for non-base-image stages in Dockerfiles |
| SCAR-011 | **principal-engineer**, **qa-adversary** | Any `@lru_cache` settings requires `clear_settings_cache()` autouse fixture |
| ISS-5 / TIMESTAMPTZ | **principal-engineer**, **architect** | All ORM datetime columns must use `TIMESTAMPTZ`; never naive `DateTime()` |
| CRITICAL-001–004 | **architect**, **principal-engineer**, **qa-adversary** | Auth role assignments must enforce business isolation; audit logs are immutable |
| Auth login 200 bug | **principal-engineer** | Side-effecting operations in auth flows must return honest error codes on failure |
| WARN-1 | **principal-engineer** | Use `.lower()` comparisons in telemetry checkers |
| DC-7 | **principal-engineer** | All S2S HTTP calls must go through `ResilientCoreClient` (preserves traceparent) |
| OTel enum/type | **principal-engineer** | When new span outcomes are added, update convention YAML atomically |
| ECS CANARY waiter | **principal-engineer** | Do not use `aws ecs wait services-stable` for CANARY deployments |
| AUTOM8Y_ENV env_prefix | **architect** | Settings with custom `env_prefix` must use `AliasChoices` for global env vars |
| Pytest importlib | **principal-engineer**, **qa-adversary** | Do not remove `--import-mode=importlib`; basenames are non-unique |
| SecretStr.__eq__ | **principal-engineer**, **qa-adversary** | Always call `.get_secret_value()` before comparing `SecretStr` |
| Satellite uv | **principal-engineer** | Satellite CI must use `uv sync --no-sources` |
| TF private module | **principal-engineer** | TF workflows fetching private modules need git HTTPS credentials |
| Grafana/ECR naming | **principal-engineer**, **architect** | Verify TF naming assumptions against actual resource names |
| Ads stub mode | **principal-engineer** | Never rely on code defaults for production-critical flags; override in Terraform |
| reconcile-ads client | **principal-engineer** | `ResilientCoreClient` requires `Config() → Client.from_config() → .wrap()` chain |
| Health check split | **architect** | New services must split ALB `/health` vs ECS `/ready` to avoid double-penalty |
| Dev mode guard | **principal-engineer**, **qa-adversary** | Dev-only feature flags must require BOTH flag AND non-production env |
| datetime.utcnow | **principal-engineer** | Never use `datetime.utcnow()`; always `datetime.now(timezone.utc)` |
| Devconsole boundaries | **principal-engineer** | Dev tools must degrade gracefully; wrap all init/DB ops in try/except |

---

## Knowledge Gaps

1. **DEF-E04** is not evidenced in source or tests despite DEF-E01, E02, E03, and E05 existing.
2. **ISS-1 through ISS-4, ISS-6, ISS-10** from the auth 503 outage audit were not individually traced.
3. **DC-1, DC-8, F-3, F-4, SMS-TEST** appear in phase-closure commit `0432d7f` as resolved advisories but not traced in detail — SDK-level issues.
4. **DEF-MED-001–003, DEF-LOW-001–004** appear in SDK quality commits but not traced in detail.
5. **EC-10 reference** in `reconcile-spend/metrics.py` refers to an error code category not defined in scope.
6. **No service-level scar markers** found in `auth-mysql-sync`, `account-status-recon`, `sms-performance-report`, or `slack-alert` source code beyond cross-cutting scars.
