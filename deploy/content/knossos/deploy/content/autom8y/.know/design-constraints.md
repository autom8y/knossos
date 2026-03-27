---
domain: design-constraints
generated_at: "2026-03-25T12:13:17Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "3fe30a4"
confidence: 0.87
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "3b84414216e3b39a03383be96b726206e9a9b54735258543352582ae8d4b0239"
---

# Codebase Design Constraints

**Language:** Python monorepo. Services under `services/`, SDKs under `sdks/python/`, infrastructure under `terraform/`. 21+ SDK packages and 13+ application services. The `scheduling` service was removed; `contente-onboarding`, `calendly-intake`, `validate-business`, and `devconsole` services are new additions.

---

## Tension Catalog

### TENSION-P001: Auth Service Uses Bare `src` Package With `from src.` Imports

**Type:** Naming mismatch / layering quirk
**Location:** `services/auth/src/main.py`, all auth routes and services
**Evidence:** 207 `from src.` imports across auth service files. `services/auth/pyproject.toml`: `packages = ["src"]`.
**Historical reason:** Auth scaffolded before `autom8_` naming convention.
**Ideal resolution:** Rename root package to `autom8_auth`; update all imports.
**Resolution cost:** High. 207+ import sites, Docker entrypoints, Alembic compatibility.

---

### TENSION-P002: Mothballed Code Still On-Disk (Credential Vault, Charter, OAuth)

**Type:** Zombie abstraction / dead code with Alembic constraint
**Locations:** `services/auth/src/services/credential_vault.py`, `services/auth/src/models/external_credential.py`, `services/auth/src/routes/charter.py`, `services/auth/src/routes/credentials.py`, `services/auth/src/routes/oauth.py`
**Historical reason:** ADR-VAULT-001 mothballed Credential Vault. ADR-CHARTER-001 mothballed Charter.
**Resolution cost:** High -- requires Alembic migration dropping tables.

---

### TENSION-P003: Charter Client Has UUID Type Safety Gap

**Type:** Over-engineering residue / type coercion debt
**Location:** `services/auth/src/charter/client.py`
**Evidence:** 38 `str(user_id)` coercions plus `# type: ignore` comments.
**Resolution cost:** Medium -- Charter is mothballed, so blocked.

---

### TENSION-P004: Ads Service LaunchService Is Not Wired in Production

**Type:** Under-engineering / incomplete wiring
**Location:** `services/ads/src/autom8_ads/app.py:74-89`
**Evidence:** `LaunchService` created only if `platform_adapter` is set. No concrete adapter injected.
**Historical reason:** Meta platform adapter deferred to Move 4.

---

### TENSION-P005: StubDataServiceClient in Production (Write Noop)

**Type:** Under-engineering / stub in prod path
**Location:** `services/ads/src/autom8_ads/clients/data.py`
**Evidence:** `StubDataServiceClient` logs warnings but persists nothing.
**Resolution cost:** Medium -- protocol defined in `autom8y-interop`.

---

### TENSION-P006: Auth Service Has Three Independent Redis Connection Pools

**Type:** Dual-system pattern / resource multiplication
**Location:** `services/auth/src/redis_client.py`, `services/auth/src/services/identifier.py`
**Evidence:** Three global Redis clients: `_revocation_client`, `_rate_limit_client`, `_identifier_redis`. All independent fail-open logic.
**Resolution cost:** Medium -- all 3 fail-open paths must be preserved.

---

### TENSION-P007: Auth Service Uses Both Sync and Async DB Sessions

**Type:** Dual-system pattern / in-progress migration
**Location:** `services/auth/src/db/database.py`, `services/auth/src/services/api_key_service.py`
**Evidence:** `APIKeyService.__init__` accepts `Session | AsyncSession`. ADR-ASYNC-DB-001 migration ongoing.
**Resolution cost:** Medium.

---

### TENSION-P008: Package Naming Mismatch -- `autom8-ads` vs `autom8y-*`

**Type:** Naming mismatch
**Location:** `services/ads/pyproject.toml`
**Evidence:** `name = "autom8-ads"` (no 'y'). All SDKs follow `autom8y-*` convention.
**Resolution cost:** High -- 40+ import sites.

---

### TENSION-P008b: Devconsole Naming Mismatch -- `autom8-devconsole` vs `autom8y-*`

**Type:** Naming mismatch
**Location:** `services/devconsole/pyproject.toml`
**Evidence:** `name = "autom8-devconsole"`. Internal package `autom8_devconsole`.
**Resolution cost:** Low -- localhost-only developer tool.

---

### TENSION-P009: GUID_NAMESPACE Is a Frozen Cryptographic Constant

**Type:** Frozen cryptographic constant
**Location:** `services/auth-mysql-sync/src/sync/guid_converter.py:20`
**Evidence:** `GUID_NAMESPACE: Final[uuid.UUID] = uuid.UUID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")` seeds deterministic UUID v5.
**Resolution cost:** Permanently frozen.

---

### TENSION-P010: Reconcile-Spend Has Dead 3-Way Reconciliation Stubs

**Type:** Missing abstraction / architectural stub
**Location:** `services/reconcile-spend/src/reconcile_spend/stubs.py`
**Evidence:** `ThreeWayComparison` and `AsanaExpectedSpendEnricher` defined but never called.

---

### TENSION-P011: `autom8y-auth` _compat.py Maintains Deprecated Re-Exports Past Deadline

**Type:** Backward-compatibility bridge / overdue migration
**Location:** `sdks/python/autom8y-auth/src/autom8y_auth/_compat.py`
**Evidence:** Deprecation states "removed in v1.0.0" -- current version is v1.1.1.
**Resolution cost:** Low -- delete `_compat.py`. Risk: consumers using deprecated imports break.

---

### TENSION-P012: TokenManager Duplicated in autom8y-auth and autom8y-core

**Type:** Dual-system pattern / duplication
**Locations:** `sdks/python/autom8y-auth/src/autom8y_auth/token_manager.py`, `sdks/python/autom8y-core/src/autom8y_core/token_manager.py`
**Resolution cost:** Medium -- verify callers, deprecate auth version.

---

### TENSION-P013: mypy Excluded from pre-commit for `services/auth/.*`

**Type:** Quality gate gap
**Location:** `.pre-commit-config.yaml:107`
**Resolution cost:** Medium -- migrate auth to pyproject.toml mypy config.

---

### TENSION-P014: ADR-SEC-GATE-POLICY Governance Artifact

**Evidence:** `.ledge/decisions/ADR-SEC-GATE-POLICY.md` confirmed present. Partially resolved.

---

### TENSION-P015-P020: Scheduling Service Tensions -- OBSOLETE

**Status:** Scheduling service removed. All tensions P015-P020 (notification dispatch stub, idempotency scan, GCal fire-and-forget, APScheduler placeholder, dual gcal_enabled flags) are stale. Corresponding LOAD-P010, LOAD-P011, RISK-P010-P012 also obsolete.

---

### TENSION-P021: SagaContext Schema Coupled to SFN ASL State Machine

**Type:** Dual-system coupling / schema stability constraint
**Location:** `sdks/python/autom8y-saga/src/autom8y_saga/models.py:29` (`SagaContext`)
**Evidence:** `extra="forbid"`. If ASL passes additional keys or SagaContext gains new required fields without ASL update, all handlers fail at `SagaContext.model_validate(event)`.
**Historical reason:** Intentional per ADR-contente-onboarding-saga to catch typos.
**Resolution cost:** Low to document; coordinated to change.

---

### TENSION-P022: DynamoDB `_dynamo.py` Validates at Module Import Time

**Type:** Startup ordering constraint
**Location:** `services/contente-onboarding/src/contente_onboarding/_dynamo.py:21-24`
**Evidence:** Missing `SAGA_CONTEXT_TABLE_NAME` kills Lambda before handler invocation.
**Resolution cost:** Low code change; high correctness risk if done naively.

---

### TENSION-P023: New Services Use Plain Naming (Not `autom8y-*` or `autom8-*`)

**Type:** Naming mismatch / ecosystem inconsistency
**Location:** `services/contente-onboarding/pyproject.toml`, `services/validate-business/pyproject.toml`, `services/calendly-intake/pyproject.toml`
**Evidence:** Package names: `contente-onboarding`, `validate-business`, `calendly-intake`. Third naming convention.
**Resolution cost:** Low -- internal-only packages.

---

### TENSION-P024: autom8y-devx-types Uses `hatchling` While All Others Use `uv_build`

**Type:** Toolchain inconsistency
**Location:** `sdks/python/autom8y-devx-types/pyproject.toml`
**Resolution cost:** Low -- trivial migration.

---

### TENSION-P025: Calendly-Intake Redis Idempotency Is Non-Authoritative (Fail-Open)

**Type:** Idempotency gap / best-effort deduplication
**Location:** `services/calendly-intake/src/calendly_intake/app.py:40-52`
**Evidence:** Redis fail-open means duplicate webhooks may process during Redis outages.
**Resolution cost:** Low to document; medium to change.

---

## Trade-off Documentation

| Tension | Current State | Why It Persists |
|---------|--------------|----------------|
| P001 | `from src.` imports (207 sites) | Alembic + Docker coupling |
| P002 | Mothballed code on disk | Alembic migration history references models |
| P004/P005 | Stub client, no adapter | Blocked on Move 4 |
| P006 | 3 Redis connection pools | Features implemented independently |
| P007 | Hybrid sync/async DB | In-progress migration (ADR-ASYNC-DB-001) |
| P008 | `autom8-ads` naming | Pre-convention; 40+ import sites |
| P009 | GUID_NAMESPACE frozen | Deterministic UUID v5 |
| P011 | _compat.py past deadline (v1.1.1) | No consumer audit |
| P012 | Duplicate TokenManager | Core factoring incomplete |
| P021 | SagaContext extra=forbid + ASL | Intentional typo guard |
| P025 | Redis fail-open idempotency | Availability over guarantee |

### External Constraints
- **NHC MySQL schema:** Field names dictated by external system
- **Meta Ads API:** Protocol shape externally constrained
- **AWS CodeArtifact:** Hardcoded URL in workspace pyproject.toml
- **AWS Step Functions ASL:** SagaContext schema coupled to SFN state machine definition
- **GCal DWD Scope:** Two-tenant architecture (thenaturalhealthcompany.org + contenteapp.com)

---

## Abstraction Gap Mapping

### Missing Abstractions
- **Shared Redis connection pool** (auth) -- 3 independent pools with identical config
- **Real DataServiceClient** (ads) -- protocol defined, only stub exists
- **Real MetaPlatformAdapter** (ads) -- protocol defined, no implementation
- **Formal ADR for auth `src` rename** -- 207 imports, no decision record
- **SagaContext schema versioning** -- no `schema_version` field; adding fields requires coordinated change
- **Idempotency guarantee for calendly-intake** -- Redis fail-open means no durable deduplication

### Premature Abstractions
- **`AdFactory`** (ads) -- 8-line wrapper with single caller
- **`DataSchedulingProtocol`** -- protocol without reachable implementation in this repo
- **`CompensationRegistry`** (autom8y-saga) -- design-time registry; SFN ASL is actual controller
- **`SagaDefinition` and `SagaStepDefinition`** -- defined but not used by current services

### Duplication
- **TokenManager** -- identical in `autom8y-auth` and `autom8y-core`
- **`_compat.py` pattern** -- repeated in `autom8y-auth` and `autom8y-log`
- **Secret resolution via extension** -- `services/contente-onboarding/src/contente_onboarding/_secrets.py` duplicates `LambdaServiceSettingsMixin`
- **Client lazy-init globals** -- each of 7 contente-onboarding handlers has own `_client = None` pattern

---

## Load-Bearing Code Identification

### LOAD-P001: GUID_NAMESPACE (FROZEN)
**Location:** `services/auth-mysql-sync/src/sync/guid_converter.py:20`
**Naive fix failure:** Namespace change silently diverges identity mapping for all existing records.

### LOAD-P002: Autom8yBaseSettings._guard_production_urls()
**Location:** `sdks/python/autom8y-config/src/autom8y_config/base_settings.py:157-200`
**Naive fix failure:** Removing guard allows LOCAL/TEST to hit production.

### LOAD-P003: Auth JWT_ALGORITHM ClassVar
**Location:** `services/auth/src/config.py:39`
**Naive fix failure:** Instance variable allows algorithm confusion attacks.

### LOAD-P004: Auth DB Engine Module-Level Singleton
**Location:** `services/auth/src/db/database.py:34-38`
**Naive fix failure:** Per-request engine defeats connection pooling.

### LOAD-P005: Incremental Sync Deactivation Exclusion
**Location:** `services/auth-mysql-sync/src/sync/orchestrator.py:443`
**Naive fix failure:** Adding deactivation to incremental sync incorrectly deactivates memberships.

### LOAD-P006: Audit Log Immutability Triggers
**Location:** `services/auth/migrations/versions/006_audit_immutability_trigger.py:55`
**Naive fix failure:** Dropping trigger allows audit log mutation.

### LOAD-P007: Fail-Open Policy for Redis (ADR-0017)
**Location:** `services/auth/src/redis_client.py:265-295`
**Naive fix failure:** Fail-closed causes total auth outage during Redis issues.

### LOAD-P008: Autom8yBaseSettings._resolve_secret_uris()
**Location:** `sdks/python/autom8y-config/src/autom8y_config/base_settings.py:231-254`
**Naive fix failure:** Removing `mode="before"` breaks all secret resolution.

### LOAD-P009: Parliament Pre-commit Hook (SEC-04)
**Location:** `.pre-commit-config.yaml:124-139`

### LOAD-P012: SagaContext Schema (extra="forbid")
**Location:** `sdks/python/autom8y-saga/src/autom8y_saga/models.py:29`
**Naive fix failure:** Adding new required field without ASL update fails all handlers.

### LOAD-P013: DynamoDB Write-Ahead Module-Level Validation
**Location:** `services/contente-onboarding/src/contente_onboarding/_dynamo.py:21-24`
**Naive fix failure:** Missing env var kills Lambda container at import time (invisible in standard metrics).

---

## Evolution Constraint Documentation

| Area | Rating | Rationale |
|------|--------|-----------|
| `GUID_NAMESPACE` | **Frozen** | Deterministic UUID; changing corrupts identity |
| `JWT_ALGORITHM` ClassVar | **Frozen** | Security -- must remain ClassVar |
| `charter_audit_logs` trigger | **Frozen** | ADR-0026 / CRITICAL-004 |
| Mothballed model files | **Frozen** | Alembic history dependency |
| ADR-SEC-GATE-POLICY | **Frozen** | CONSTITUTIONAL -- CRITICAL=BLOCK |
| Incremental sync deactivation | **Frozen** | Design choice preventing lockouts |
| `SagaContext` field schema | **Frozen** | Coupled to SFN ASL |
| `SAGA_CONTEXT_TABLE_NAME` env var | **Frozen** | Module-level validation at cold start |
| Auth `from src.` imports | **Migration** | 207 sites; coordinated rename |
| ADR-ASYNC-DB-001 | **Migration** | In-progress |
| StubDataServiceClient | **Migration** | Move 4 pending |
| autom8y-auth _compat.py | **Migration** | Deadline passed (v1.1.1) |
| TokenManager duplication | **Migration** | Core factoring incomplete |
| Redis 3-pool split | **Coordinated** | Fail-open paths must be preserved |
| `Autom8yBaseSettings` | **Coordinated** | All services extend it |
| Parliament pre-commit | **Coordinated** | SEC-04; security review required |
| mypy auth exclusion | **Migration** | Requires pyproject.toml migration |
| autom8y-devx-types `hatchling` | **Migration** | Trivial align with uv_build |

---

## Risk Zone Mapping

### RISK-P001: Ads LaunchService Not Wired in Production
`services/ads/src/autom8_ads/app.py:76-89` -- `AttributeError` on production requests.

### RISK-P002: Charter Routes Contain Bug Preventing Admin Access
`services/auth/src/routes/charter.py:79-85` -- `is_admin` does not exist on User model.

### RISK-P003: Incremental Sync Has No Deactivation Boundary Guard
`services/auth-mysql-sync/src/sync/orchestrator.py:443` -- Only comment prevents adding deactivation.

### RISK-P004: Three Redis Clients Have Independent Fail-Open Logic
`services/auth/src/redis_client.py` -- Incident response must check 3 log streams.

### RISK-P005: model_post_init Override Silently Drops Production URL Guard
`sdks/python/autom8y-config/src/autom8y_config/base_settings.py:157-163`

### RISK-P006: Auth Service Imports structlog Directly
`services/auth/src/observability/logger.py:16-17` -- Bypasses `autom8y-log`.

### RISK-P007: Parliament Pre-commit Hook Has Silent Bypass
`.pre-commit-config.yaml:136` -- Skips when not installed. Local only.

### RISK-P008: Semgrep Security Rules File Is Empty
`.semgrep-security.yml` -- `rules: []`. Registry packs used via CLI flag.

### RISK-P009: autom8y-config Is DAG Root for All SDK Releases
Config SDK imported by 15/17 SDKs and all services.

### RISK-P013: contente-onboarding `_secrets.py` Bypasses LambdaServiceSettingsMixin
`services/contente-onboarding/src/contente_onboarding/_secrets.py` -- Custom secret resolution parallel to autom8y-config.

### RISK-P014: Non-Idempotent GCal Calendar Creation in Saga
`services/contente-onboarding/src/contente_onboarding/create_gcal.py` -- DynamoDB sentinel guard required for idempotency.

### RISK-P015: Devconsole References Removed Scheduling Service
`services/devconsole/src/autom8_devconsole/config.py:44` -- `SCHEDULING_BASE_URL: str = "http://localhost:8001"`. Service no longer exists.

---

## Knowledge Gaps

1. **KG-1:** ADR-VAULT-001 and ADR-CHARTER-001 content not found in `.ledge/decisions/`.
2. **KG-2:** "Move" terminology (Move 3, Move 4) in ads service -- no definition document found.
3. **KG-3:** `autom8y-data` service source not in this repository.
4. **KG-4:** `autom8y-asana` service source not in this repository.
5. **KG-5:** `guid_migrations` table schema not cataloged.
6. **KG-6:** ADR-SEC-GATE-POLICY file present; content not audited this cycle.
7. **KG-7:** autom8y-auth v1.0.0 removal obligations -- `_compat.py` still at v1.1.1.
8. **KG-8:** ADR-ASYNC-DB-001 completion criteria and timeline not documented.
9. **KG-9:** AWS SFN ASL definition for contente-onboarding not in this repo.
10. **KG-10:** `autom8y-data` modular monolith decomposition status unclear.
11. **KG-11:** Calendly-intake deduplication behavior under Redis outage unverified.
