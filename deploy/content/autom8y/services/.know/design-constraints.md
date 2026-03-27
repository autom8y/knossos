---
domain: design-constraints
generated_at: "2026-03-16T20:09:00Z"
expires_after: "7d"
source_scope:
  - "./*/src/**/*.py"
  - "./*/tests/**/*.py"
  - "./*/pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog Completeness

This section catalogs all structural tensions observed across all 10 services. Tensions from existing per-service `.know/design-constraints.md` files are synthesized and cross-referenced; new tensions for uncovered services are documented here.

### TENSION-001: Duplicate Name Encoding — reconcile-ads / autom8y-ads Schema Coupling

**Type**: Naming mismatch / phantom contract
**Location**: `reconcile-ads/src/reconcile_ads/joiner.py` lines 34-91; `account-status-recon/src/account_status_recon/joiner.py` lines 44-53

The bullet-separator campaign name encoding (`\u2022`-delimited: `phone•offer_id•vertical•...`) is the join key used by both `reconcile-ads` and `account-status-recon`. This encoding lives in `autom8y-ads` but is replicated inline without a contract test. Both services and the `autom8y-ads` service must stay synchronized on separator character (U+2022), field positions, and field count — but there is no automated cross-service assertion.

---

### TENSION-002: Dual Verdict System (Legacy Scalars + SDK Unified Verdicts) — reconcile-ads

**Type**: Half-completed migration
**Location**: `reconcile-ads/src/reconcile_ads/models.py` lines 198-250

`Finding` carries both legacy scalar verdicts (`status_verdict`, `budget_verdict`, `delivery_verdict`) and the new unified SDK `verdicts: dict[VerdictAxis, UnifiedVerdict]`. The `severity` property uses SDK path when `verdicts` is populated, falling back to legacy map. The report builder reads legacy scalar fields. Both paths are simultaneously active.

---

### TENSION-003: Dual Staleness Threshold System — reconcile-spend

**Type**: Naming mismatch + orphaned config field
**Location**: `reconcile-spend/src/reconcile_spend/config.py` lines 91-113

`max_staleness_minutes` (default 45, no consumer) and `max_staleness_seconds` (default 1800, feeds `readiness.py`) coexist. The Slack alert displays `max_staleness_seconds // 60 = 30 min` as the threshold, while `max_staleness_minutes = 45` is silently ignored by all code.

---

### TENSION-004: Protocol Downcast via `assert isinstance` — reconcile-spend

**Type**: Layering violation
**Location**: `reconcile-spend/src/reconcile_spend/clients/data_service.py` lines 21-29

`resolve_insight_client()` returns `DataInsightProtocol`; immediately asserted to concrete `DataInsightClient` to access `.circuit_breaker_state` and `.http`. Python `assert` is compiled away under `-O` optimization.

---

### TENSION-005: Dual RBAC Systems (Active + Mothballed Charter) — auth

**Type**: Dual system / dead code with live schema impact
**Location**: `auth/src/charter/` (all files); `auth/src/routes/charter.py`

Two parallel, incompatible RBAC systems. Active system: `roles`/`permissions`/`user_roles` tables (always business-scoped). Charter system: `charter_*` table family with system-wide null-`business_id` roles. Charter routes are commented out in `src/main.py:242-247`.

---

### TENSION-006: Refresh Token Verified via API Key Function — auth

**Type**: Naming mismatch / shared infrastructure coupling
**Location**: `auth/src/services/token_lookup.py:61`; `auth/src/routes/auth.py` lines 548, 736, 1068

`hash_api_key(refresh_token)` and `verify_api_key(stored.token_hash, raw_token)` are used for refresh token operations. Changing API key hashing parameters would silently break refresh token verification.

---

### TENSION-007: Three Separate Redis Connection Pools for One Instance — auth

**Type**: Under-engineering / incremental coupling
**Location**: `auth/src/redis_client.py`; `auth/src/services/identifier.py`

`RevocationClient`, `RateLimitClient`, and `_identifier_redis` each maintain independent connection pools to the same Redis URL.

---

### TENSION-008: Sync/Async DB Session Dual Track — auth

**Type**: In-progress migration
**Location**: `auth/src/db/database.py`; `auth/src/services/api_key_service.py` lines 43, 124, 175

`_sync_engine` and `_async_engine` coexist. `APIKeyService` has `isinstance(session, AsyncSession)` branching. Blocked by SQLModel incomplete async support.

---

### TENSION-009: Mothballed Credential Vault — 15 Dead Files with Live Schema — auth

**Type**: Dead code with live infrastructure dependency
**Location**: `auth/src/services/credential_vault.py` and 14 other files (per ADR-VAULT-001)

All 15 files are MOTHBALLED per their headers. Models remain imported in `src/models/__init__.py:13-16` because Alembic requires them.

---

### TENSION-010: `business_scope_validation` Middleware Implemented But Unregistered — auth

**Type**: Missing security enforcement
**Location**: `auth/src/middleware/business_scope.py`

Middleware is implemented and tested but absent from the ASGI middleware stack in `src/main.py`. Cross-tenant access validation is not enforced at the middleware layer.

---

### TENSION-011: Platform Protocol Exists Without Concrete Implementation — ads

**Type**: Missing abstraction / structural gap
**Location**: `ads/src/autom8_ads/platforms/protocol.py`

`AdPlatform` protocol defines four async methods. No concrete `MetaPlatformAdapter` exists in the repository.

---

### TENSION-012: Stub Data Persistence Layer — ads

**Type**: Premature abstraction + intentional deferral
**Location**: `ads/src/autom8_ads/clients/data.py`

`StubDataServiceClient` performs no persistence — only logs at `WARNING`. `data_writes_enabled` defaults to `False`. Real implementation deferred to "Move 4" (unscheduled). Setting `DATA_WRITES_ENABLED=true` with the stub causes silent data loss.

---

### TENSION-013: V2-Only Strategy via Over-Engineered Strategy Pattern — ads

**Type**: Over-engineering
**Location**: `ads/src/autom8_ads/lifecycle/`

ADR-ADS-002 mandates V2-only. The `LaunchStrategy` Protocol + `AdFactory` indirection layer has no routing logic and no strategy selection. It adds complexity for a single concrete implementation path.

---

### TENSION-014: `stubs.py` Zombie Architectural Documentation — reconcile-spend

**Type**: Intentionally dead code preserved as documentation
**Location**: `reconcile-spend/src/reconcile_spend/stubs.py`

`ThreeWayComparison` and `AsanaReconciliation` classes are never imported. Preserved as reference for ADR-reconciliation-architecture-evolution.md.

---

### TENSION-015: `src.` Package Prefix Import Pattern — auth-mysql-sync

**Type**: Non-standard package layout / fragile import paths
**Location**: All files in `auth-mysql-sync/src/`

Unlike all other services (which use a named package under `src/`), `auth-mysql-sync` uses `from src.config import ...`. The package name IS `src`. This creates a fragile import path that only works when the working directory is `services/auth-mysql-sync/`.

---

### TENSION-016: `provision_business` GUID Lookup — Full Table Scan Workaround

**Type**: Under-engineering / performance risk
**Location**: `auth-mysql-sync/src/portover/handler.py` lines 226-256

`MySQLReader` has no direct GUID lookup method. The workaround: fetch ALL sync-eligible employees, extract all office phones, fetch ALL chiropractors by those phones, then iterate to find the target.

---

### TENSION-017: Financial Summary Fields Never Populated — account-status-recon

**Type**: Zombie stub / scaffolded but unimplemented
**Location**: `account-status-recon/src/account_status_recon/models.py` lines 247-250

`total_ghost_daily_budget`, `total_underbilled_variance`, `total_contract_drift` are always `0.0`.

---

### TENSION-018: `stale_days_threshold` and `three_way_severe_threshold_pct` Never Used — account-status-recon

**Type**: Interface / implementation mismatch
**Location**: `account-status-recon/src/account_status_recon/rules.py`; `account-status-recon/src/account_status_recon/config.py`

`stale_days_threshold` parameter is threaded through the call chain but Rule 5 uses a binary `== 0` check. `three_way_severe_threshold_pct` is in config and function signature but `rule_three_way` only uses one threshold.

---

### TENSION-019: Inline Rule Reimplementation Across Three Services

**Type**: Cross-service duplication
**Location**: `account-status-recon/src/account_status_recon/rules.py` (entire file)

STATUS, BUDGET, DELIVERY, BILLING, THREE_WAY rules are inline reimplementations of logic in `reconcile-ads/rules.py` and `reconcile-spend/rules.py`. ADR-ASR-003 explicitly chose this over SDK extraction.

---

### TENSION-020: Third-Party SDK SLACK_ Prefix vs. AUTOM8Y_ Ecosystem Convention — slack-alert

**Type**: Naming mismatch / external constraint
**Location**: `slack-alert/src/slack_alert/handler.py:31`

`SlackConfig` uses `env_prefix="SLACK_"`, making production env var `SLACK_BOT_TOKEN`. Ecosystem convention is `AUTOM8Y_*`.

---

### TENSION-021: Dual `PaymentRecord` and Dual `BatchResult` Types — pull-payments

**Type**: Naming mismatch / duplication
**Location**: `pull-payments/src/pull_payments/models.py`; `pull-payments/scripts/dry_run.py:52`; `pull-payments/src/pull_payments/clients/data_service.py:34`

`PaymentRecord` exists in both `models.py` (canonical) and `dry_run.py` (script-only with extra display fields). `BatchResult` exists both locally and as an interop SDK re-export.

---

### TENSION-022: S3 Staging + /tmp Fallback Dual Backend — pull-payments

**Type**: Two-backend system / migration in progress
**Location**: `pull-payments/src/pull_payments/staging.py`

S3 is primary; `/tmp` is fallback. Three legacy `/tmp` functions preserved for test backward compatibility. The staging bucket is not validated at startup.

---

### TENSION-023: Service-Specific API Key Alias + Generic Fallback (Fleet-Wide Pattern)

**Type**: Dual-alias env var pattern
**Location**: All lambda services config files

Every lambda service uses `AliasChoices("SERVICE_SPECIFIC_KEY", "SERVICE_API_KEY")`. Removing either alias without coordinating Terraform changes causes Lambda startup failure.

---

### TENSION-024: `GUID_NAMESPACE` is Immutable Post-Deployment — auth-mysql-sync

**Type**: Frozen constant / migration constraint
**Location**: `auth-mysql-sync/src/sync/guid_converter.py:20`

`GUID_NAMESPACE = uuid.UUID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")` is explicitly documented: "This namespace MUST NOT change after initial deployment." Changing it would break all existing external ID references.

---

### TENSION-025: devconsole Has No Production Infrastructure

**Type**: Developer-only service with no Lambda/ECS deployment path
**Location**: `devconsole/src/autom8_devconsole/`

The devconsole is a NiceGUI-based developer observability tool. No `handler.py`, no Lambda entry point, no IaC. Fleet-wide deployment assumptions do not apply.

---

## Trade-off Documentation

### Trade-off for TENSION-001 (Name Encoding Drift)
**Chosen**: Inline decode in each service using the same bullet-separator pattern.
**Rejected**: Extract to a shared `autom8y-naming` package.
**Why persists**: Zero cross-service import friction. Shared package requires versioned release cycle.

### Trade-off for TENSION-005 (Dual RBAC)
**Chosen**: Active RBAC runs; Charter mothballed.
**Rejected**: Unified RBAC from the start.
**Why persists**: Charter reactivation requires migration strategy from active system + fixing `is_admin` bug.

### Trade-off for TENSION-008 (Sync/Async DB)
**Chosen**: Incremental async migration, `APIKeyService` dual-mode shim.
**Why persists**: SQLModel incomplete async support.

### Trade-off for TENSION-009 (Mothballed Vault)
**Chosen**: Keep mothballed models imported in `__init__.py`.
**Rejected**: Remove from codebase.
**Why persists**: Alembic requires model classes importable for `autogenerate`.

### Trade-off for TENSION-011 (Missing Platform Adapter)
**Chosen**: Protocol defined, injection deferred to caller.
**Why persists**: Move 3 constraint — `MetaPlatformAdapter` not yet built.

### Trade-off for TENSION-012 (Stub Data Client)
**Chosen**: `StubDataServiceClient` with `data_writes_enabled=False`.
**Why persists**: Move 4 unscheduled. `DATA_WRITES_ENABLED=true` with stub causes silent data loss.

### Trade-off for TENSION-019 (Cross-Service Rule Duplication)
**Chosen**: Inline reimplementation using SDK verdict types.
**Rejected**: Import rule functions from sibling services (not Python packages).
**ADR**: ADR-ASR-003.

### Trade-off for TENSION-022 (Dual Staging Backend)
**Chosen**: S3 primary with `/tmp` fallback (FR-12).
**Why persists**: Migration in progress.

### Trade-off for TENSION-024 (Immutable GUID Namespace)
**Chosen**: Fixed UUID v5 namespace, never changes post-deployment.
**Why persists**: Deterministic UUID v5 generation is the design; changing namespace breaks external ID integrity.

---

## Abstraction Gap Mapping

### GAP-001: No Shared HTTP Client Factory (Reconciliation Services)

**Services**: account-status-recon (`fetcher.py`), reconcile-ads (`fetcher.py`)

The three-line `Config → Client → ResilientCoreClient` construction pattern appears in 3+ fetcher functions across multiple services. A shared factory function would eliminate duplication.

### GAP-002: No Unified Alert Module — reconcile-spend

**Location**: `reconcile-spend/src/reconcile_spend/orchestrator.py` lines 685-773

`_build_stale_data_alert()` and `_build_circuit_open_alert()` are private functions in `orchestrator.py`. The canonical home for Block Kit construction is `report.py`.

### GAP-003: No Typed Campaign Tree Schema — reconcile-ads

Three modules re-implement `.get("campaign", {})` access patterns. No typed `AdsTreeResponse` Pydantic model enforced at the boundary.

### GAP-004: Revocation Check Not in Primary Auth Dependency — auth

`get_current_user` verifies JWT but does NOT call `RevocationClient.is_token_revoked`. Revoked tokens continue authenticating until natural expiry (up to 15 minutes).

### GAP-005: No Concrete `AdPlatform` Implementation — ads

Only `protocol.py` exists. `MetaPlatformAdapter` is absent. No production ad launch is possible without external injection.

### GAP-006: Batch Flush Logic Duplicated 4x — pull-payments

The "accumulate, flush at batch_size, flush remaining" pattern appears 4 times in orchestrator. No `BatchFlusher` abstraction exists.

### GAP-007: No Direct GUID Lookup Query — auth-mysql-sync

`MySQLReader` has no `get_chiropractor_by_guid()` method. The portover handler compensates with a full-scan workaround.

### Premature Abstractions

- **PREMATURE-001**: `AdFactory` with single strategy — `ads/src/autom8_ads/lifecycle/factory.py`
- **PREMATURE-002**: `TargetingSpec` typed fields — `ads/src/autom8_ads/models/targeting.py` — always uses `TargetingSpec(raw=payload.targeting)`
- **PREMATURE-003**: `DataServiceProtocol` with four methods (ads) — only `record_campaign` has a call site
- **PREMATURE-004**: `stale_days_threshold` and `three_way_severe_threshold_pct` — account-status-recon — threaded through call stacks but never consumed

### Zombie Abstractions

- **ZOMBIE-001**: Legacy `/tmp` staging functions — `pull-payments/src/pull_payments/staging.py:297-331`
- **ZOMBIE-002**: `FetchError` class — `account-status-recon/src/account_status_recon/errors.py:22-42` — never raised in `fetcher.py`
- **ZOMBIE-003**: `JoinError` class — `account-status-recon/src/account_status_recon/errors.py:45-49` — never raised in `joiner.py`
- **ZOMBIE-004**: `FetchError(ReconcileAdsError)` — `reconcile-ads/src/reconcile_ads/errors.py` — fetcher uses its own exception hierarchy
- **ZOMBIE-005**: `build_all_clear_report()` — `reconcile-spend/src/reconcile_spend/report.py:126-141` — not imported by any module

---

## Load-Bearing Code Identification

### LB-001: `parse_client_records()` — reconcile-spend

**Location**: `reconcile-spend/src/reconcile_spend/orchestrator.py` lines 55-152

Single gateway from `ReconciliationRow` to `ClientRecord`. The `math.isfinite()` guard (E10) MUST precede the negative-collected adjustment (E9) — reversing this order allows NaN to propagate and silently suppress anomaly detection.

### LB-002: `GUID_NAMESPACE` — auth-mysql-sync

**Location**: `auth-mysql-sync/src/sync/guid_converter.py:20`

Immutable post-deployment. Changing it generates different UUIDs for all numeric chiropractor GUIDs, breaking all existing Auth external ID cross-references. **Rating: FROZEN.**

### LB-003: `_hasher` Argon2id Instance — auth

**Location**: `auth/src/auth/password.py` lines 18-21

Canonical hasher for all Argon2id operations: password hashing, API key hashing, and refresh token hashing. Changing parameters invalidates ALL existing stored hashes. No migration path exists. **Rating: FROZEN.**

### LB-004: `JWT_ALGORITHM = "RS256"` ClassVar — auth

**Location**: `auth/src/config.py` lines 37-39

`ClassVar` prevents env var override. Switching to HS256 would invalidate all issued tokens and break every service using `autom8y_auth_client`. **Rating: FROZEN.**

### LB-005: `billing_key_fn` / `campaign_key_fn` / `contract_key_fn` — account-status-recon

**Location**: `account-status-recon/src/account_status_recon/joiner.py` lines 25-65

Define the `(office_phone, vertical)` composite key. The bullet separator `\u2022` at index positions `[0]` and `[2]` must match `autom8y-ads` campaign naming encoding.

### LB-006: `_decode_campaign_name` / `_decode_ad_group_name` — reconcile-ads

**Location**: `reconcile-ads/src/reconcile_ads/joiner.py` lines 70-91

Bridge between Meta campaign name strings and join keys. The `\u2022` separator is hardcoded in 3 places.

### LB-007: `MetaUrlBuilder` Filter Encoding — ads

**Location**: `ads/src/autom8_ads/urls/meta.py:89`, `133`

`%1E` / `%1D` separators replicate Meta's proprietary ASCII control character encoding.

### LB-008: `_is_full_success()` / `_extract_failed_records()` — pull-payments

**Location**: `pull-payments/src/pull_payments/replay.py` lines 356-377

`_is_full_success()` controls whether a staged batch is deleted post-replay. Partial failures without error detail are treated as full failures, risking data duplication.

### LB-009: `RuntimeError` Raise After Batch — slack-alert

**Location**: `slack-alert/src/slack_alert/handler.py:83-85`

This raise fires the Lambda Errors metric that drives the 99.9% SLO burn-rate alert. Removing it silently breaks SLO monitoring.

### LB-010: `get_settings()` with `@lru_cache` (Fleet-Wide Pattern)

**Locations**: Every service config module

Settings are resolved once per Lambda container lifecycle. A warm Lambda reused after SSM secret rotation serves stale settings until container kill. `clear_settings_cache()` exists but is test-only.

### LB-011: `ReconciliationResult.to_dict()` — Manual Lambda Response Serialization

**Locations**: `reconcile-spend/src/reconcile_spend/models.py:159-175`; `reconcile-ads/src/reconcile_ads/models.py:293-309`

Manual `to_dict()` on Pydantic `BaseModel` instead of `model_dump()`. Fields not wired into `to_dict()` are silently absent from Lambda response.

### LB-012: `models/__init__.py` Mothballed Model Imports — auth

**Location**: `auth/src/models/__init__.py:13-16`

Imports all mothballed Credential Vault models. Alembic `env.py` requires these for `autogenerate`. Removing them breaks migration generation. **Rating: COORDINATED.**

---

## Evolution Constraint Documentation

| Area | Service | Rating | Evidence |
|------|---------|--------|----------|
| JWT token algorithm (RS256) | auth | **FROZEN** | `ClassVar` prevents override; all tokens + consuming services depend on it |
| Argon2id hash parameters | auth | **MIGRATION** | Changing invalidates all stored hashes with no migration tooling |
| Mothballed Charter module | auth | **FROZEN** | Cannot reactivate without `is_admin` fix, RBAC migration |
| Mothballed Credential Vault | auth | **FROZEN** | 15 mothballed files; Alembic dependency blocks removal |
| Redis connection architecture | auth | **COORDINATED** | 3 pools + lifecycle management spread across files |
| Sync→Async DB migration | auth | **COORDINATED** | ADR-ASYNC-DB-001 in-progress; SQLModel async incomplete |
| `GUID_NAMESPACE` constant | auth-mysql-sync | **FROZEN** | Post-deployment change breaks all numeric GUID→UUID mappings |
| Campaign name encoding format | reconcile-ads, account-status-recon | **MIGRATION** | Coordinated change with `autom8y-ads` required |
| `OfferPayload` field set | ads | **FROZEN** | PRD-locked; requires PRD update + `autom8y_asana` coordination |
| Idempotency cache key format | ads | **FROZEN** | Changes break in-flight request deduplication |
| `MetaUrlBuilder` encoding | ads | **FROZEN** | `%1E`/`%1D` separators match Meta legacy format |
| `StubDataServiceClient` | ads | **MIGRATION** | Swap for real client at Move 4 |
| `algo_version` V2 validator | ads | **FROZEN** | ADR-ADS-002 enforcement |
| Service API key alias chain | all Lambda services | **MIGRATION** | Dual-alias; removing either side requires IaC coordination |
| Lambda handler signatures | all Lambda services | **FROZEN** | Lambda runtime contract |
| S3 staging infrastructure | pull-payments | **COORDINATED** | Requires IaC + code + migration of legacy `/tmp` files |
| Replay algorithm | pull-payments | **COORDINATED** | External algorithm specification |
| `ReconciliationRow` all-optional fields | reconcile-spend | **COORDINATED** | Tightening any field requires `autom8y-data` coordination |
| `stubs.py` zombie classes | reconcile-spend | **FROZEN** (as documentation) | Reference for ADR |
| Composite key schema | account-status-recon | **MIGRATION** | Tied to Meta campaign naming convention |
| EventBridge event schemas | account-status-recon | **FROZEN** | External EventBridge rule consumers |
| `SLACK_` prefix on bot token | slack-alert | **COORDINATED** | Locked to third-party SDK until modified |
| SNS+CloudWatch event shape | slack-alert | **FROZEN** | AWS-defined schema |
| Rule logic (reconciliation) | reconcile-ads, reconcile-spend, account-status-recon | **COORDINATED** | Must be propagated across all three services manually |
| `src.` package import prefix | auth-mysql-sync | **SAFE** | Non-standard; rename to `auth_mysql_sync` would fix |
| devconsole service topology | devconsole | **SAFE** | Developer-only; no deployment constraints |

---

## Risk Zone Mapping

### RISK-001: `assert isinstance` Stripped by Python `-O` — reconcile-spend

**Location**: `reconcile-spend/src/reconcile_spend/clients/data_service.py:29`
**Missing defense**: Replace with `if not isinstance(...): raise TypeError(...)`.

### RISK-002: `ReconciliationRow` All-Optional — Silent Total-Failure Mode — reconcile-spend

**Location**: `reconcile-spend/src/reconcile_spend/clients/models.py:47-74`
If upstream stops sending `office_phone`, every row has `office_phone=None`. Lambda returns HTTP 200, `accounts_analyzed=0`, `all_clear=True`.
**Missing defense**: Alert on `accounts_analyzed=0` when previous runs had non-zero counts.

### RISK-003: Revocation Check Absent from `get_current_user` — auth

**Location**: `auth/src/auth/dependencies.py`
Explicitly revoked tokens continue authenticating for up to 15 minutes.

### RISK-004: `business_scope_validation` Not Registered — auth

**Location**: `auth/src/middleware/business_scope.py`
Cross-tenant validation middleware is implemented but absent from ASGI stack.

### RISK-005: In-Memory Brute Force State Not Shared — auth

**Location**: `auth/src/utils/brute_force.py`
Module-level globals. In horizontally scaled deployments, attackers distribute attempts across instances.

### RISK-006: Warm Lambda Stale Secret Risk (Fleet-Wide)

**Locations**: Every `config.py` with `@lru_cache` on `get_settings()`
AWS secret rotation while Lambda container is warm serves old credentials until container kill.

### RISK-007: S3 Load Error Sends Transient Errors to `corrupt/` — pull-payments

**Location**: `pull-payments/src/pull_payments/staging.py:185-196`
Any S3 API error (throttling, access denied, network timeout) triggers `move_to_corrupt(key)`.
**Missing defense**: Distinguish S3 API errors (transient) from JSON parse errors (permanent).

### RISK-008: Silent Data Loss When Both S3 and /tmp Fail — pull-payments

**Location**: `pull-payments/src/pull_payments/staging.py:136-148`
Lambda disk-full causes uncaught `OSError` propagating through `batch_create_payments`.

### RISK-009: Unguarded Name Decode Drift (Three Services)

**Locations**: `reconcile-ads/src/reconcile_ads/joiner.py`; `account-status-recon/src/account_status_recon/joiner.py`
No contract tests between reconciliation services and `autom8y-ads` asserting campaign name format stability.
**Missing defense**: Contract test asserting separator and field positions.

### RISK-010: `_publish_complete_event()` Failure at DEBUG Level (Fleet-Wide)

**Locations**: reconcile-spend, reconcile-ads, account-status-recon orchestrators
EventBridge publish failures logged at `DEBUG` level (below default `INFO`). Production-invisible.
**Missing defense**: Elevate to `WARNING` level.

### RISK-011: Production Service Cannot Start Without External Platform Adapter — ads

**Location**: `ads/src/autom8_ads/app.py:76-89`
`LaunchService` only constructed when `platform_adapter` is in `app.state` at startup. No startup validation.

### RISK-012: Secrets Manager Path Ambiguity — slack-alert

Three SM paths exist; slack-alert defaults to `autom8y/slack/alerts-bot-token`, but shared Terraform creates `autom8y/slack/bot-token`. Status: `PENDING_HUMAN_ACTION`.

### RISK-013: `_unpack_aws_secrets_bundle` Silently Swallows JSON Parse Errors — auth-mysql-sync

**Location**: `auth-mysql-sync/src/config.py:101-135`
The `except (ValueError, TypeError): pass` block means a malformed `AWS_SECRETS` bundle is silently ignored.
**Missing defense**: Log a warning when `AWS_SECRETS` is present but fails to parse.

---

## Knowledge Gaps

1. **ADR files not found**: 18+ ADRs referenced in source code but no `.ledge/decisions/` directories exist in any service tree. Trade-off documentation relies on code comments and inferred rationale.
2. **`autom8y-ads` `ActiveCampaignTreeResponse` schema**: The exact field names and nesting structure are not visible in this monorepo. Load-bearing external dependency for three reconciliation services.
3. **`autom8y-interop` `DataInsightProtocol` interface**: Whether `circuit_breaker_state` could be lifted to the protocol to resolve TENSION-004 is not determinable.
4. **`autom8y-reconciliation` SDK internals**: `Correlator`, `ReadinessGate`, `UnifiedVerdict` are used across five services but their source is not visible here.
5. **devconsole constraints**: No `.know/` files exist for devconsole. NiceGUI, OTLP receiver, Tempo integration design constraints are not formally documented.
6. **sms-performance-report constraints**: No `.know/` files exist. Follows reconciliation service pattern but specific tensions are not documented.
7. **auth-mysql-sync constraints**: Key tensions (TENSION-015, TENSION-016, TENSION-024, RISK-013) documented here for the first time.
