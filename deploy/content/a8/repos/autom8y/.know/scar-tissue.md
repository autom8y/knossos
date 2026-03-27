---
domain: scar-tissue
generated_at: "2026-03-25T12:13:17Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "3fe30a4"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "c560748d333bca99de6e38cb459c5fb85f941f00f921de0c4a384e48c55b4d7f"
---

# Codebase Scar Tissue

**Project**: autom8y (Python monorepo)
**Observation Date**: 2026-03-25
**Primary Language**: Python (3.11/3.12) -- services, SDKs, Lambda functions
**Secondary Languages**: HCL (Terraform IaC), YAML (CI workflows, OTel conventions)
**Source Scope**: `services/`, `sdks/python/`, `terraform/`, `.github/workflows/`, `scripts/`
**Git History Depth**: Full history (~421 fix-tagged commits scanned)

---

## Failure Catalog

### Group 1: Auth Service Production Outages

**SCAR-001 -- Auth 503 Outage (ISS-1 through ISS-11)**

The auth service went into a 503 outage with 11 root causes catalogued:
- ISS-5: `asyncpg` rejects offset-aware Python datetimes against `TIMESTAMP WITHOUT TIME ZONE` schema columns
- ISS-7: Health check used wrong `SecretId` (`"auth-service-secrets"` instead of `"autom8y/auth/jwt-private-key"`)
- ISS-8: `redis` package missing from `pyproject.toml`
- ISS-9: No CloudWatch alarm for empty target group -- silent deployment failures

Fix: `services/auth/migrations/versions/016_timestamps_to_timestamptz.py`, `services/auth/src/main.py` lines 302/351, `services/auth/pyproject.toml`, `terraform/services/auth/main.tf`

**Auth SCAR-001 -- Login timeout via body consumption** (commit `e638524`): Rate limit middleware consumed ASGI body stream, causing downstream hang. Fix: `services/auth/src/middleware/rate_limit.py` lines 121-129.

**Auth SCAR-007 -- Connection pool defeat** (commit `38673aa`): New engine per request. Fix: module-level singletons in `services/auth/src/db/database.py` lines 35-38.

### Group 2: Auth Security Hardening Scars

**Auth SCAR-012 -- User enumeration timing attack** (HIGH-001, commit `23e9bfc`): Non-existent user logins returned in ~5ms vs ~100ms. Fix: `_DUMMY_HASH` at `services/auth/src/routes/auth.py` lines 65-70.

**Auth SCAR-013 -- Refresh token business_id bypass** (HIGH-002): Token refresh queried by `token_hash` only, allowing cross-tenant access. Fix: `services/auth/src/routes/auth.py` line 721.

**Auth SCAR-014 -- Password reset O(n) DoS** (HIGH-003): Full table scan for reset tokens. Fix: `token_prefix` column + index in migration `015`.

**Auth SCAR-019 -- Logout no-op (re-hash bug)** (commit `ae2e2d2`): Logout re-hashed token with non-deterministic Argon2id. Fix: `find_token_by_value()` in `services/auth/src/services/token_lookup.py`.

**Auth SCAR-023 -- JWT algorithm override via env var** (commit `65eb859`): `JWT_ALGORITHM` was settable via env. Fix: `ClassVar[str] = "RS256"` in `services/auth/src/config.py` line 39.

**Auth SCAR-026 -- CRITICAL-001 through CRITICAL-004** (ADRs 0023-0026): Missing role scoping by business_id, no JWT key rotation fallback, privilege escalation risk, mutable audit logs. Hardened with 1300-line regression suite at `services/auth/tests/test_critical_security_fixes.py`.

### Group 3: Configuration / Environment Variable Drift

**SCAR-012** (commit `1367461`): Child config classes with custom `env_prefix` searched for `{PREFIX}AUTOM8Y_ENV` instead of `AUTOM8Y_ENV`. Fix: `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` in `sdks/python/autom8y-config/src/autom8y_config/base_settings.py` lines 87-93.

**SCAR-022** (commit `666de732`): Satellite repos `[tool.uv.sources]` local paths don't exist in CI. Fix: `uv sync --no-sources` in `.github/workflows/satellite-ci-reusable.yml`.

**SCAR-015** (commit `73c579a`): `%` in DATABASE_URL interpreted as configparser interpolation. Fix: `.replace("%", "%%")` in `services/auth/migrations/env.py` line 20.

### Group 4: Test Isolation Failures

**SCAR-011**: Settings singleton leaked state between tests. Fix: `cache_clear()` autouse fixture. Template at `sdks/python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py` lines 101-112.

**SCAR-014**: structlog cached logger configuration. Fix: `configure(cache_logger_on_first_use=False)` in `sdks/python/autom8y-log/tests/conftest.py` line 40.

**SCAR-PYTEST-001** (commit `03f81758`): pytest-asyncio 1.0 breaking API. Fix: `pytest-asyncio>=1.2,<2.0` pin across all pyproject.toml files.

**SCAR-PYTEST-002** (commit `299aded`): 93 `ImportPathMismatchError` from shared test file basenames. Fix: `--import-mode=importlib` in root `pyproject.toml`.

### Group 5: Observability / Telemetry Failures

**SCAR-TELEMETRY-001** (commit `dac2866`): Lambda `BatchSpanProcessor` dropped spans on freeze. Fix: `_force_flush_provider()` in `try/finally` at `sdks/python/autom8y-telemetry/src/autom8y_telemetry/aws/lambda_instrument.py` lines 96-101.

**SCAR-TELEMETRY-002** (commit `dac2866`): `instrument_app()` never called `FastAPIInstrumentor.instrument_app()`. No per-request spans despite config. Fix in `sdks/python/autom8y-telemetry/src/autom8y_telemetry/fastapi/instrument.py`.

**SCAR-TELEMETRY-003 / SCAR-010** (commit `14b415a`): OTel Python SDK auto-appends `/v1/traces` only from env var, not constructor kwarg. Traces silently dropped. Fix in `sdks/python/autom8y-telemetry/src/autom8y_telemetry/init.py` lines 204-208.

**SCAR-TELEMETRY-005** (commit `1139ccc`): Convention checker case-sensitive comparison. Fix: `.lower()` in `sdks/python/autom8y-telemetry/src/autom8y_telemetry/conventions/checker.py`.

**SCAR-004** (commit `12cfd801`): DMS alert rule used `"max"` reducer. Fix: `"count"` in `terraform/modules/platform/primitives/grafana-service-alerts/`.

**SCAR-HTTP-003** (commit `d4996f0`): Missing W3C `traceparent` header injection. Fix in `sdks/python/autom8y-http/src/autom8y_http/resilience/client.py` lines 137-138.

### Group 6: CI / CD / Deployment Failures

**SCAR-028** (commit `2968cfa`): `aws ecs update-service` without `--force-new-deployment` left stale code running. Fix: `.github/workflows/service-deploy.yml` line 196.

**SCAR-008** (commit `f1fb1818`): SDK publish version detection missed multi-commit pushes. Fix: CodeArtifact registry query in `.github/workflows/sdk-publish-v2.yml`.

**SCAR-024** (commit `9689c796`): ECS health check `startPeriod` 60s too short for ADOT sidecar. Fix: `startPeriod: 120` in `.github/workflows/service-deploy.yml`.

**SCAR-005**: autom8y-config 69 mypy errors blocked 5 release candidates. Fix in `sdks/python/autom8y-config/src/autom8y_config/base_settings.py`.

**SCAR-018** (commit `f837f7b7`): Terraform 1.6.0 crashed on sensitive values. Fix: `terraform_version: 1.9.0` in all CI workflows.

### Group 7: Dependency / Package Management Failures

**SCAR-006** (commit `395e8196`): Grafana provider v3.x phantom diff -> 409. Fix: provider pin `~> 4.0` in `grafana-service-alerts/versions.tf`.

**SCAR-HTTP-001** (commit `86a276d`): `pydantic_settings` used at import time but undeclared. Fix: added to `[core]` extras.

**SCAR-GCAL-001** (commit `b82324b`): `google-auth` transport requires undeclared `requests`. Fix: added to `autom8y-gcal/pyproject.toml`.

**SCAR-AUTH-FASTAPI-001** (commit `8419771`): Eager FastAPI import crashed non-web consumers. Fix: lazy-loading in `__init__.py`.

### Group 8: Data Integrity / Business Logic Failures

**DEF-1 (reconcile-spend)** (commit `c893811`): Non-200 data service responses unhandled. Fix in `services/reconcile-spend/src/reconcile_spend/clients/data_service.py`.

**DEF-4 (reconcile-spend)** (commit `c893811`): NaN/Inf guard missed variance fields. Fix in `services/reconcile-spend/src/reconcile_spend/orchestrator.py` lines 76-90.

**DEF-E01/E02 (reconcile-spend)** (commit `5be1729`): Slack mrkdwn link corruption from special characters. Fix: HTML entity escaping in `services/reconcile-spend/src/reconcile_spend/report.py`.

### Group 9: Contract / Schema Drift

**SCAR-019**: `autom8y-interop` models evolved without satellite detection. Fix: JSON Schema golden files via `scripts/contract_schemas.py`.

**SCAR-026 / GCal camelCase** (commit `aab3499`): Google Calendar API camelCase not mypy-safe. Fix: `validation_alias=AliasChoices()` in `sdks/python/autom8y-gcal/src/autom8y_gcal/models.py`.

**SCAR-INTEROP-001**: `SchedulingError` hierarchy split across packages breaking `isinstance`. Fix: explicit re-export in `sdks/python/autom8y-interop/src/autom8y_interop/data/errors.py` lines 14-27.

---

## Category Coverage

| Category | Count | Key Scars |
|----------|-------|-----------|
| Integration failure | ~12 | SCAR-001, DEF-1, SCAR-INTEROP-001 |
| Security / auth bypass | 7 | auth SCAR-012, -013, -014, -019, -023, -026 |
| Config drift / env var mismatch | 15+ | SCAR-012, SCAR-CONFIG-001, TF-001-TF-010 |
| Test isolation | 5 | SCAR-011, SCAR-014, SCAR-PYTEST-001, -002 |
| Observability / tracing | 9 | SCAR-TELEMETRY-001 through -006, SCAR-004 |
| CI / CD / deployment | 7 | SCAR-028, SCAR-008, SCAR-024 |
| Dependency management | 6 | SCAR-007, SCAR-HTTP-001, SCAR-GCAL-001 |
| Data integrity | 6 | DEF-4, DEF-E01/E02, DEF-E03 |
| Schema evolution | 3 | SCAR-013, SCAR-017, SCAR-TELEMETRY-006 |
| Performance / resource mgmt | 4 | auth SCAR-007, SCAR-003 |
| Type safety / API contract | 5 | SCAR-021, SCAR-026 |

**Distinct failure mode categories**: 11

**Not found**: race conditions in production async code, message queue failures, cache invalidation bugs.

---

## Fix-Location Mapping

| Scar | Primary Fix File(s) |
|------|---------------------|
| SCAR-001 | `services/auth/migrations/versions/016_timestamps_to_timestamptz.py`, `services/auth/src/main.py`, `services/auth/pyproject.toml`, `terraform/services/auth/main.tf` |
| SCAR-002 | `services/auth/src/db/seeds/seed_dev_keys.py`, `docker/dev/Dockerfile.ecs-3.11` |
| SCAR-003 | `sdks/python/autom8y-http/src/autom8y_http/resilience/registry.py` lines 44-46 |
| SCAR-004 | `terraform/modules/platform/primitives/grafana-service-alerts/` |
| SCAR-005 | `sdks/python/autom8y-config/src/autom8y_config/base_settings.py` |
| SCAR-006 | `terraform/modules/platform/primitives/grafana-service-alerts/versions.tf` |
| SCAR-007 | `pyproject.toml` + all 20 SDK `pyproject.toml` files |
| SCAR-008 | `.github/workflows/sdk-publish-v2.yml` |
| SCAR-009 | `terraform/services/grafana/alerting_trace_pipeline.tf` lines 18-19 |
| SCAR-011 | `sdks/python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py` lines 101-112 |
| SCAR-012 | `sdks/python/autom8y-config/src/autom8y_config/base_settings.py` lines 87-93 |
| SCAR-013 | `services/auth/migrations/versions/016_timestamps_to_timestamptz.py` |
| SCAR-014 | `sdks/python/autom8y-log/tests/conftest.py` line 40 |
| SCAR-015 | `services/auth/migrations/env.py` line 20 |
| SCAR-019 | `scripts/contract_schemas.py` |
| SCAR-022 | `.github/workflows/satellite-ci-reusable.yml` |
| SCAR-024 | `.github/workflows/service-deploy.yml` |
| SCAR-025 | `sdks/python/autom8y-auth/src/autom8y_auth/_observability.py` |
| SCAR-028 | `.github/workflows/service-deploy.yml` line 196 |
| Auth SCAR-012 | `services/auth/src/routes/auth.py` lines 65-70 |
| Auth SCAR-013 | `services/auth/src/routes/auth.py` line 721 |
| Auth SCAR-019 | `services/auth/src/services/token_lookup.py` |
| Auth SCAR-023 | `services/auth/src/config.py` line 39 |
| Auth SCAR-026 | `services/auth/tests/test_critical_security_fixes.py` (1300+ lines) |

---

## Defensive Pattern Documentation

| Pattern | Spawned By | Location |
|---------|-----------|----------|
| IP-based rate limiting; never consume body in middleware | auth SCAR-001 | `services/auth/src/middleware/rate_limit.py:121-135` |
| `_DUMMY_HASH` + constant-time verify on user-not-found | auth SCAR-012 | `services/auth/src/routes/auth.py:65-70` |
| Token hash AND business_id verification on refresh | auth SCAR-013 | `services/auth/src/routes/auth.py:721` |
| `token_prefix` column + index for O(1) reset lookup | auth SCAR-014 | migration `015` |
| `find_token_by_value()` for logout (no re-hash) | auth SCAR-019 | `services/auth/src/services/token_lookup.py` |
| `JWT_ALGORITHM: ClassVar` -- non-overridable | auth SCAR-023 | `services/auth/src/config.py:39` |
| Full security hardening (CRITICAL-001-004) | auth SCAR-026 | `services/auth/tests/test_critical_security_fixes.py` |
| Bounded circuit breaker group names (regex) | SCAR-003 | `sdks/python/autom8y-http/src/autom8y_http/resilience/registry.py:44-49` |
| Settings cache clear autouse fixture | SCAR-011 | `sdks/python/autom8y-sms-test/src/autom8y_sms_test/fixtures.py:83-112` |
| `structlog.configure(cache_logger_on_first_use=False)` | SCAR-014 | `sdks/python/autom8y-log/tests/conftest.py:40` |
| `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` | SCAR-012 | `sdks/python/autom8y-config/src/autom8y_config/base_settings.py:87-93` |
| `uv sync --no-sources` for satellite CI | SCAR-022 | `.github/workflows/satellite-ci-reusable.yml` |
| Contract schema golden files | SCAR-019 | `scripts/contract_schemas.py` |
| `--force-new-deployment` on ECS update | SCAR-028 | `.github/workflows/service-deploy.yml:196` |
| `_force_flush_provider()` in Lambda try/finally | SCAR-TELEMETRY-001 | `sdks/python/autom8y-telemetry/src/autom8y_telemetry/aws/lambda_instrument.py:96-101` |
| Explicit `/v1/traces` suffix for OTLP endpoint | SCAR-TELEMETRY-003 | `sdks/python/autom8y-telemetry/src/autom8y_telemetry/init.py:204-208` |
| Fail-open Redis with `REDIS_AVAILABLE` flag | SCAR-027 | `services/auth/src/redis_client.py:18-25` |
| NaN/Inf guard on numeric fields | DEF-4 | `services/reconcile-spend/src/reconcile_spend/orchestrator.py:76-90` |
| HTML entity escape for Slack mrkdwn | DEF-E01/E02 | `services/reconcile-spend/src/reconcile_spend/report.py` |
| `--import-mode=importlib` in root pytest config | SCAR-PYTEST-002 | `pyproject.toml:79` |
| Module-level singleton DB engines with pool_pre_ping | auth SCAR-007 | `services/auth/src/db/database.py:11-14` |

**Scars with no dedicated regression test**: SCAR-004, SCAR-006, SCAR-009, SCAR-010, SCAR-012, auth SCAR-006, auth SCAR-007.

---

## Agent-Relevance Tagging

| Scar Group | Relevant Agents | Why |
|-----------|----------------|-----|
| Auth security (SCAR-012-014, -019, -023, -026) | qa-adversary, principal-engineer | Any login/refresh/logout change must preserve timing safety and hash verification |
| Config/env drift (SCAR-012, SCAR-CONFIG-001) | architect, principal-engineer | New config classes must use AliasChoices for canonical env vars |
| Test isolation (SCAR-011, -014, SCAR-PYTEST-001, -002) | principal-engineer, qa-adversary | Every new conftest.py must include settings cache clear AND structlog disable |
| Observability (SCAR-TELEMETRY-001 through -006) | principal-engineer | Always `instrument_lambda`; never OTLP endpoint without `/v1/traces` |
| CI/CD deployment (SCAR-028, -008, -024) | platform-engineer, release-executor | `--force-new-deployment` non-negotiable; `startPeriod: 120` |
| Dependency management (SCAR-HTTP-001, SCAR-GCAL-001) | principal-engineer | All optional deps must declare transitive requirements |
| Data integrity (DEF-1, DEF-4, DEF-E01-E03) | qa-adversary, principal-engineer | NaN/Inf guard before variance; Slack output must HTML-escape |
| Contract drift (SCAR-019, SCAR-INTEROP-001) | architect, principal-engineer | Interop model changes require `scripts/contract_schemas.py --check` |
| Provisioning / saga (DynamoDB sentinel) | principal-engineer, qa-adversary | Saga steps must be idempotent via DynamoDB sentinel |

---

## Knowledge Gaps

1. **Lambda architecture revert root cause undocumented**: Commit sequence `4745049 -> a4f9e39 (revert) -> 9e15f05` not captured in named scar entry.

2. **Auth SCAR-026 ADR details**: CRITICAL-001 through CRITICAL-004 cataloged at summary level only; ADR documents not read.

3. **Credential Vault and Charter mothball lineage**: Original failure/decision not captured. ADR-VAULT-001 and ADR-CHARTER-001 exist but were not read.

4. **Security hardening campaign sessions 4-5**: Campaign was "3/5 sessions complete" at last observation. Sessions 4-5 outcomes undocumented.

5. **SRE SCAR numbering beyond SCAR-SRE-009 and SCAR-SRE-016**: Other SRE-scoped scars not located in current monorepo scope.

6. **Services without scar docs**: `services/contente-onboarding/`, `services/reconcile-ads/`, `services/calendly-intake/`, `services/pull-payments/` lack primary `.know/scar-tissue.md` files.

7. **S2S auth seed gap (SCAR-002)**: Marked fixed but working state across all services not independently confirmed.
