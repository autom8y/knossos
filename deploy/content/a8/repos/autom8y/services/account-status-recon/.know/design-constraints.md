---
domain: design-constraints
generated_at: "2026-03-16T14:32:40Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog Completeness

### TENSION-001: Inline Rule Reimplementation vs. Shared Rule Library

**Type**: Duplication / Premature isolation

**Location**: `src/account_status_recon/rules.py` (entire file)

**Description**: All five verdict-axis rules (STATUS, BUDGET, DELIVERY, BILLING, THREE_WAY) are inline reimplementations of logic that lives in two sibling services (`reconcile-ads/rules.py` and `reconcile-spend/rules.py`). The file comments make this explicit:

- Line 36: `# STATUS axis rules (mirrors reconcile-ads/rules.py:rule_status_alignment)`
- Line 91: `# BUDGET axis rules (mirrors reconcile-ads/rules.py:rule_budget_alignment)`
- Line 133: `# DELIVERY axis rules (mirrors reconcile-ads/rules.py:rule_delivery_health)`
- Line 170: `# BILLING axis rules (mirrors reconcile-spend/rules.py:Rules 1-5)`
- Line 268: `Implements Rules 6-7 from reconcile-spend/stubs.py`

**Historical reason**: ADR-ASR-003 explicitly chose inline reimplementation using SDK verdict types rather than inheriting from sibling services. The rationale is that rule logic should be expressed through the SDK's `UnifiedVerdict.create()` / `VerdictAxis` types — not by importing from peer services (which are not packages).

**Ideal resolution**: Extract shared rule logic into `autom8y-reconciliation` SDK as reusable primitives.

**Resolution cost**: High — requires SDK changes, versioned release, coordinated migration across three services.

---

### TENSION-002: Fetch Client Construction Inside Fetcher Functions

**Type**: Over-engineering risk / missing abstraction

**Location**: `src/account_status_recon/fetcher.py` — all three `fetch_*` functions (lines 27-51, 54-94, 97-164)

**Description**: Each fetcher function (`fetch_billing`, `fetch_campaigns`, `fetch_offers`) constructs its own `Config`, `Client`, `ResilientCoreClient` stack inline. The imports for these are deferred inside the function body to isolate test mocking. The same three-line construction pattern repeats three times:

```python
config = Config(service_key=settings.service_api_key_value, auth_url=settings.auth_base_url)
core = Client.from_config(config)
http = ResilientCoreClient.wrap(core)
```

**Historical reason**: The comment on `fetch_billing` (line 26) says `Pattern: reconcile-spend/orchestrator.py:299-315` — this was copied verbatim from the sibling service. Deferred imports were retained to allow `respx` mock patching in tests.

**Ideal resolution**: Extract a factory function `_build_http_client(settings)` and share it across the three fetchers.

**Resolution cost**: Low — single-file refactor with no cross-service impact.

---

### TENSION-003: Financial Summary Fields Never Populated

**Type**: Dead abstraction / zombie fields

**Location**: `src/account_status_recon/models.py` lines 247-250, `src/account_status_recon/report.py` lines 112-115

**Description**: `ReconciliationResult` declares three financial summary fields:

```python
total_ghost_daily_budget: float = 0.0
total_underbilled_variance: float = 0.0
total_contract_drift: float = 0.0
```

`report.py` conditionally renders them if non-zero (lines 112-115). However, nowhere in the orchestrator or rules is any of these fields populated — they remain at `0.0` for every run. The Slack report code block for financial impact (`FR-27`) is structurally dead.

**Historical reason**: These were likely stubs introduced during sprint planning for FR-27 financial summary, with population logic deferred. The field declaration and report-side rendering were scaffolded but never wired.

**Ideal resolution**: Either populate these fields from findings in `apply_all_rules` / `_aggregate_result_counts`, or remove them and the corresponding report block.

**Resolution cost**: Medium — requires deciding which calculation logic to add, or confirming these fields are abandoned.

---

### TENSION-004: `stale_days_threshold` Parameter Never Used in Rule 5 Logic

**Type**: Interface / implementation mismatch

**Location**: `src/account_status_recon/rules.py` lines 179, 194, 248-249, 345, 364, 413

**Description**: `rule_billing` accepts a `stale_days_threshold: int = 14` parameter and documents it as "Rule 5 staleness threshold in days." But Rule 5's actual implementation (line 248) is:

```python
if billing.days_with_activity == 0:
```

The threshold is never consulted. The parameter is threaded through `apply_all_rules` → `rule_billing` but silently dropped. The only staleness check performed is whether `days_with_activity == 0` (zero activity, not "days_with_activity < stale_days_threshold").

**Historical reason**: Likely copied from a reconcile-spend pattern where the threshold was meaningful. In this codebase, the simpler zero-check was implemented but the parameter signature was kept.

**Ideal resolution**: Either implement `billing.days_with_activity < stale_days_threshold` as intended, or remove the parameter and document that Rule 5 is a binary zero-check.

**Resolution cost**: Low — single-function change with test update.

---

### TENSION-005: `three_way_severe_threshold_pct` Accepted but Not Applied

**Type**: Interface / implementation mismatch

**Location**: `src/account_status_recon/config.py` line 95, `src/account_status_recon/rules.py` line 342, `src/account_status_recon/orchestrator.py` line 202

**Description**: `Settings` exposes `three_way_severe_threshold_pct` (default 25.0). The orchestrator passes it into `apply_all_rules`. `apply_all_rules` accepts it in its signature. But `apply_all_rules` does not pass it to `rule_three_way`, and `rule_three_way` only uses a single `drift_threshold_pct`. The severe threshold is wired through config and the function signature but never consulted by any rule function.

The intention from `config.py` (line 97): "Three-way severe divergence threshold (%). Default 25.0." suggests a two-tier three-way verdict (DRIFT vs. SEVERE) was planned but never implemented.

**Historical reason**: Config and orchestrator wiring were done in anticipation of a SEVERE verdict variant in the THREE_WAY axis, but the rule logic was implemented with only a single threshold.

**Ideal resolution**: Either implement a SEVERE verdict tier in `rule_three_way` that uses this threshold, or remove the parameter from config and the function signature.

**Resolution cost**: Low — rule change with test coverage update.

---

### TENSION-006: Campaign Data Access Pattern Mixes Dict and SDK Model

**Type**: Abstraction inconsistency / defensive casting

**Location**: `src/account_status_recon/fetcher.py` lines 75-79

**Description**: The campaign fetch coerces the response to a dict:

```python
response_dict: dict[str, Any] = (
    response.model_dump()
    if hasattr(response, "model_dump")
    else cast("dict[str, Any]", response)
)
```

Then the joiner accesses this as a raw dict (`campaign_items` parameter), while all other sources use typed parsed rows. The campaign path uses a `hasattr` guard rather than a type check, indicating uncertainty about whether the SDK returns a Pydantic model or a plain dict.

**Historical reason**: Comment references `Pattern: reconcile-ads/fetcher.py:54-141` — this suggests the pattern was copied from a service that had already adopted this defensive coercion.

**Ideal resolution**: Pin the SDK type for `AdsCampaignTreeClient.get_active_tree()` response and remove the hasattr guard.

**Resolution cost**: Low — requires checking SDK contract.

---

## Trade-off Documentation

### TENSION-001 Trade-off: Rule Portability vs. Service Isolation

**Current state**: STATUS, BUDGET, DELIVERY, BILLING, and THREE_WAY rules live entirely in `src/account_status_recon/rules.py` as inline functions.

**Ideal state**: These functions would live in `autom8y-reconciliation` SDK as reusable rule primitives shared across all reconciliation services.

**Why current state persists**: ADR-ASR-003 explicitly documents this choice. The rationale was that the SDK already provides verdict types (`UnifiedVerdict`, `VerdictAxis`) as the integration contract, and rule logic is service-specific enough that sharing it through the SDK would couple sibling services via a library dependency. Importing rule functions from `reconcile-ads` or `reconcile-spend` directly was rejected because those services are not Python packages.

**Attempt history**: The comment on `rule_billing` documents that it "mirrors reconcile-spend/rules.py:Rules 1-5" — indicating the original implementation was a straight port, not independent derivation.

---

### TENSION-002 Trade-off: Client Construction Duplication vs. Testability

**Current state**: All imports for `Config`, `Client`, `ResilientCoreClient` are deferred inside each `fetch_*` function body.

**Why current state persists**: Deferred imports allow test mocking via `respx` without patching at module-import time. Moving client construction to a shared factory would require the factory to be patchable, complicating test setup.

---

### TENSION-003 Trade-off: FR-27 Financial Summary as Zombie Fields

**Current state**: `total_ghost_daily_budget`, `total_underbilled_variance`, `total_contract_drift` initialized to `0.0` and never written.

**Why current state persists**: No ADR documents this explicitly. The fields were scaffolded for FR-27 but the sprint likely ended before computation logic was wired. Report rendering is guarded by `> 0` so no incorrect output occurs.

**External constraint**: The `to_dict()` method on `ReconciliationResult` intentionally omits these fields from the Lambda response body, suggesting they are not yet part of the external contract.

---

### TENSION-004 Trade-off: `stale_days_threshold` Signal Retention

**Why current state persists**: The parameter appears in function signatures used by both the orchestrator and tests. Removing it would be a breaking change to the calling interface. The zero-check was likely "good enough" for MVP and the threshold parameter was retained to avoid interface churn.

---

### TENSION-005 Trade-off: `three_way_severe_threshold_pct` Signal Retention

**Why current state persists**: Config and orchestrator wiring were likely added in anticipation of SEVERE verdict implementation. Removing the config field would require coordinated IaC changes (SSM parameter removal). Keeping it is lower friction than removing it.

---

## Abstraction Gap Mapping

### Missing Abstraction: HTTP Client Factory

**Scope**: `src/account_status_recon/fetcher.py`

The three-line client construction pattern appears in all three fetcher functions (lines 32-34, 64-66, 120-122). This is a missing extraction — a factory function `_build_http_client(settings: Settings)` returning a `ResilientCoreClient` would eliminate the duplication.

**Maintenance burden**: Any change to auth configuration (e.g., adding mTLS, changing `auth_base_url` resolution) must be applied in three places.

---

### Missing Abstraction: Financial Summary Aggregation

**Scope**: `src/account_status_recon/orchestrator.py` / `src/account_status_recon/models.py`

`_aggregate_result_counts` in `orchestrator.py` (lines 284-307) aggregates severity counts and three-way verdict counts. The financial summary fields (`total_ghost_daily_budget`, `total_underbilled_variance`, `total_contract_drift`) represent a parallel aggregation that was never implemented alongside the existing count aggregation.

**Maintenance burden**: When FR-27 is eventually implemented, the developer must locate the aggregation logic in `_aggregate_result_counts`, understand the existing pattern, and add financial aggregation — with no existing test to anchor against.

---

### Premature Abstraction: `stale_days_threshold` and `three_way_severe_threshold_pct` Parameters

**Scope**: `src/account_status_recon/rules.py` (function signatures), `src/account_status_recon/config.py`

These parameters generalize rule behavior that is not yet implemented. They serve zero use cases currently. They occupy function signatures, config fields, and orchestrator pass-through code. Each must be considered when reading/modifying these functions, even though they have no effect.

---

### Zombie Abstraction: `FetchError` Class

**Scope**: `src/account_status_recon/errors.py` lines 22-42

`FetchError` is defined with `source_name`, `method`, and `time_remaining` fields. It is never raised in `fetcher.py` — the fetchers catch all exceptions and return `FetchResult(error=exc)` instead. `FetchError` exists in the hierarchy but is not on any code path.

**Evidence**: Grepping `FetchError` across `src/` yields only the definition in `errors.py`. No code raises it.

---

### Zombie Abstraction: `JoinError` Class

**Scope**: `src/account_status_recon/errors.py` lines 45-49

`JoinError` is defined but never raised in `joiner.py`. The joiner returns empty lists and logs skipped records rather than raising errors. The error class documents an error path that doesn't exist.

---

## Load-Bearing Code Identification

### Load-Bearing: `billing_key_fn`, `campaign_key_fn`, `contract_key_fn`

**Location**: `src/account_status_recon/joiner.py` lines 25-65

**What it does**: Extracts the `(office_phone, vertical)` composite key from each data source's raw row format. These three functions define the entire correlation identity across all three sources.

**What depends on it**: Every `AccountRecord` produced by `three_way_join` depends on these functions. The entire reconciliation output changes if any key function is modified. Tests in `tests/test_joiner.py` exercise these paths.

**What a naive fix would break**:

1. `campaign_key_fn` decodes the Meta campaign name using the bullet separator `\u2022` (line 48): `parts = raw_name.split("\u2022")`. Phone is `parts[0]`, vertical is `parts[2]`. If an agent "fixes" this to use a different delimiter or index, all campaign records will fail to correlate with billing and contract data — producing a mass of `SourcePresence(billing=True, campaign=False, contract=True)` records.

2. `billing_key_fn` and `contract_key_fn` strip and normalize `vertical` to empty string if absent (lines 34-35, 64-65). This empty-string normalization ensures that accounts with no vertical still correlate across sources. Changing the fallback to `None` would produce `(phone, None)` keys that never match `(phone, "")` keys in sibling indexes.

**Hot path**: Yes — called for every row across all three data sources on every Lambda invocation.

---

### Load-Bearing: `three_way_join` Key Union Logic

**Location**: `src/account_status_recon/joiner.py` lines 174-178

**What it does**: Computes the union of all keys across three indexes: `all_keys.update(billing_idx.keys())`, etc.

**What depends on it**: The union semantics determine that every account present in ANY source appears in the output. If this were changed to an intersection, all partial-source accounts (present in only 1 or 2 sources) would be silently dropped.

**What a naive fix would break**: Changing from union to intersection would eliminate all GHOST_CAMPAIGN and MISSING_CAMPAIGN findings because those require accounts that exist in only one source.

---

### Load-Bearing: `rule_billing` Multi-Verdict Return

**Location**: `src/account_status_recon/rules.py` lines 174-251

**What it does**: Returns a `list[UnifiedVerdict]` rather than a single verdict. `apply_all_rules` takes the worst severity from this list (line 418): `worst_billing = min(billing_verdicts, key=lambda v: v.severity.value)`.

**What depends on it**: The report renders only one BILLING axis verdict per account (the worst). If the multi-verdict list were collapsed at the rule level instead of in `apply_all_rules`, the calling code in `apply_all_rules` would need to handle a single verdict. This is a contract difference between BILLING and all other axes.

**What a naive fix would break**: Changing `rule_billing` to return `UnifiedVerdict | None` (like the other rules) without updating `apply_all_rules` would cause a type error and break multi-anomaly detection.

---

### Load-Bearing: `get_settings` with `lru_cache`

**Location**: `src/account_status_recon/config.py` lines 136-138

**What it does**: Returns a cached `Settings` singleton. `clear_settings_cache()` (line 142-145) is the only safe way to reset it in tests.

**What depends on it**: All tests that call `get_settings()` must call `clear_settings_cache()` between test cases or risk cross-contamination. The `conftest.py` likely handles this, but any test that calls `get_settings()` outside of the fixture setup path will get a stale cached instance.

**What a naive fix would break**: Removing `lru_cache` would cause `Settings()` to be reconstructed on every call — including re-reading environment variables and re-initializing the secret resolver. This would break Lambda cold-start performance and cause intermittent failures if environment variables are set after module initialization.

---

### Load-Bearing: Contract Dedup in `three_way_join`

**Location**: `src/account_status_recon/joiner.py` lines 168-171

**What it does**: Applies `Correlator.dedup(contract_rows, contract_key_fn)` to the offer data, enforcing first-wins semantics on `(office_phone, vertical)` keys. The comment says "EC-12."

**What depends on it**: All three-way join downstream logic assumes exactly one `ContractData` per key. The `contract_data_list[0]` access in line 211 would incorrectly discard later contract rows if dedup were removed.

**What a naive fix would break**: Removing the dedup would produce multi-row contract lists and the first-element access at line 211 would silently use only the first row for parsing, creating inconsistency with the index shape.

---

## Evolution Constraint Documentation

### Area: Rule Thresholds — SAFE to change

**Files**: `src/account_status_recon/config.py`

All numeric thresholds (`budget_drift_threshold_pct`, `budget_mismatch_threshold_pct`, `three_way_drift_threshold_pct`, `overbilled_threshold_pct`, `underbilled_threshold_pct`, `billing_staleness_threshold_seconds`, etc.) are externalized to `Settings` and read through `get_settings()`. They can be changed via environment variable or SSM without code modification.

**Changeability**: Safe (ENV-only change).

---

### Area: Verdict Rules — COORDINATED change required

**Files**: `src/account_status_recon/rules.py`

Any change to the verdict logic for STATUS, BUDGET, DELIVERY, BILLING, or THREE_WAY axes must be coordinated with:
1. Test updates in `tests/test_rules.py` and `tests/qa/test_edge_cases_adversarial.py`
2. Report rendering in `src/account_status_recon/report.py` (if new axes or verdicts are added)
3. Metric emission in `src/account_status_recon/metrics.py` (if verdict dimension values change)
4. The sibling services (`reconcile-ads`, `reconcile-spend`) that share the same rule logic via copy

**Changeability**: Coordinated (multi-file, multi-service).

---

### Area: Composite Key Schema — MIGRATION required

**Files**: `src/account_status_recon/joiner.py` (key functions), all three upstream services

The `(office_phone, vertical)` composite key is the identity contract that ties all three data sources together. Changing the key schema (e.g., adding a third dimension, normalizing phone format) requires:
1. Matching changes in all three upstream data sources' field naming
2. Campaign name encoding change (the `\u2022`-delimited format in Meta campaign names is set in the ads service, not here)
3. Historical reconciliation records would not correlate with new key formats

**Changeability**: Migration (external dependency on Meta campaign naming convention + upstream API contracts).

---

### Area: `autom8y-reconciliation` SDK Primitives — COORDINATED change required

**Files**: `src/account_status_recon/readiness.py`, `src/account_status_recon/joiner.py`, `src/account_status_recon/rules.py`, `src/account_status_recon/report.py`, `src/account_status_recon/metrics.py`

This service depends on `autom8y-reconciliation>=0.1.0` for: `Correlator`, `ReadinessGate`, `UnifiedVerdict`, `VerdictAxis`, all verdict enum types, `ReconciliationReportBuilder`, `ReconciliationMetrics`. Any breaking SDK change requires coordinated update here.

The `pyproject.toml` uses `>=0.1.0` pinning with no upper bound — a major SDK version bump could silently break this service at the next `uv sync`.

**Changeability**: Coordinated.

---

### Area: Slack Block Kit Rendering — SAFE to change (additive)

**Files**: `src/account_status_recon/report.py`

The `render_account_finding` function builds Block Kit blocks. Adding new fields to the rendered output (new financial lines, new verdict axes) is safe — Block Kit is additive and Slack ignores unknown block properties. Removing fields is safe for display but may break downstream consumers that parse Slack message text.

**Changeability**: Safe (additive), Coordinated (removals).

---

### Area: Lambda Handler Interface — FROZEN

**Files**: `src/account_status_recon/handler.py`

The `lambda_handler(event, context)` signature and the 200/raise response shape are load-bearing for the Lambda runtime contract and for whatever IaC configures the EventBridge scheduler. The `@instrument_lambda` decorator instruments the function at the telemetry level. Changes to the return shape would require coordinated IaC and monitoring changes.

**Changeability**: Frozen (Lambda runtime contract).

---

### Area: EventBridge Event Schema — FROZEN

**Files**: `src/account_status_recon/orchestrator.py` lines 310-329

The `AccountStatusComplete` event detail schema (keys: `accounts_analyzed`, `total_findings`, `findings_by_severity`, `source_coverage`, `three_way_verdict_counts`, `run_duration_seconds`, `pipeline_readiness`) represents an external contract consumed by any downstream EventBridge rule. Renaming or removing keys is a breaking change for downstream consumers.

**Changeability**: Frozen (external EventBridge contract).

---

## Risk Zone Mapping

### RISK-001: Campaign Name Parsing Has No Validation Guard

**Location**: `src/account_status_recon/joiner.py` — `campaign_key_fn` (lines 44-53), `_parse_campaign_items` (lines 91-113)

**Type**: Missing input validation

**Evidence**: `campaign_key_fn` splits `raw_name` on `\u2022` and accesses `parts[0]` for phone and `parts[2]` for vertical, but only guards `len(parts) > 0` for phone and `len(parts) > 2` for vertical. If Meta changes the campaign naming convention to use fewer bullet separators, `parts[2]` access is guarded (returns `""`), but if `raw_name` itself is a non-string type (e.g., API returns `None` for the field), `raw_name.split()` will raise `AttributeError`. The guard at line 46 is `if not raw_name: return None` — this catches `None` and empty string, but not other non-string types.

**Recommended guard**: Add `isinstance(raw_name, str)` check or use `str(raw_name or "").split(...)`.

---

### RISK-002: `_parse_billing_row` Float Coercion Has No Validation Guard for Non-Numeric Strings

**Location**: `src/account_status_recon/joiner.py` lines 70-74

**Type**: Missing input validation

**Evidence**:

```python
spend = float(row.get("spend") or 0)
collected = float(row.get("collected") or 0)
variance = float(row.get("variance") or 0)
```

These coercions will raise `ValueError` if the upstream API returns a non-numeric string for these fields (e.g., `"N/A"`, `"--"`). There is no `try/except` around the float coercions. The `EC-11` guard in `BillingData.has_finite_values` only runs after construction, not during.

**Cross-reference**: TENSION-001 (these coercions are copied from reconcile-spend).

---

### RISK-003: Contract Classification Case Normalization Applied Inconsistently

**Location**: `src/account_status_recon/rules.py` lines 68, 78; `src/account_status_recon/joiner.py` line 140-143

**Type**: Inconsistent normalization

**Evidence**: In `rule_status`, classification is normalized with `.strip().lower()` before comparison to `_ACTIVE_CLASSIFICATIONS` and `_TRANSITIONAL_CLASSIFICATIONS`. However, in `fetcher.py` `fetch_offers`, the classification is injected as a literal string:

```python
row["classification"] = "active"      # line 141
row["classification"] = "activating"  # line 143
```

This is safe for the injected rows (always lowercase). But `_parse_contract_row` in `joiner.py` stores `row.get("classification")` as-is (line 127). If the Asana API ever returns a classification with different casing in the raw row data, the rule would not normalize it correctly at the joiner level — only at the rule level. This creates implicit coupling between the normalization in `rule_status` and the raw string values.

---

### RISK-004: `_publish_complete_event` Silently Swallows All Exceptions

**Location**: `src/account_status_recon/orchestrator.py` lines 310-344

**Type**: Silent fallback

**Evidence**:

```python
except Exception:
    log.debug("event_publish_failed", exc_info=True)
```

EventBridge publish failures are logged at `DEBUG` level and never surfaced. An EventBridge misconfiguration or IAM permission error would produce zero observable signal in production (DEBUG-level logs are not shipped in most prod configurations). The best-effort intent is documented (FR-22), but the logging level means failures will be invisible.

**Recommended guard**: Raise to `WARNING` level.

---

### RISK-005: `emit_metrics` Silently Swallows All Exceptions at DEBUG

**Location**: `src/account_status_recon/metrics.py` line 77-78

**Type**: Silent fallback

**Evidence**: Same pattern as RISK-004:

```python
except Exception:
    log.debug("metric_emission_failed", exc_info=True)
```

CloudWatch metric emission failures are lost silently. If the metric emission code path breaks (e.g., IAM regression, boto3 API change), the service would continue to report "success" with no metrics being recorded, defeating the dead-man's-switch (FR-23).

---

### RISK-006: `fetch_all` Return Assumes Exact 3 Results from `asyncio.gather`

**Location**: `src/account_status_recon/fetcher.py` lines 184-196

**Type**: Defensive assumption

**Evidence**: `zip(source_names, results, strict=True)` at line 187 will raise `ValueError` if `asyncio.gather` returns a different number of results than `source_names`. Since `source_names` is hardcoded as `["billing", "campaigns", "offers"]` and `gather` always returns as many results as coroutines given, this is safe as long as exactly three coroutines are passed. If a fourth source is added without updating `source_names`, `strict=True` would catch it — which is actually a protective pattern, not a risk.

**Residual risk**: The mismatch between fetch result keys (`"billing"`, `"campaigns"`, `"offers"`) and readiness gate keys (same names, checked via `.get()`) is implicit. Renaming a source name requires touching at least: `fetcher.py`, `readiness.py`, `orchestrator.py` lines 158-172.

---

## Knowledge Gaps

1. **ADR files not located**: References to `ADR-ASR-001`, `ADR-ASR-002`, `ADR-ASR-003` appear in source code comments but no `.ledge/decisions/` directory exists. The trade-off rationale is inferred from code comments and structure, not from recovered ADR text.

2. **Sibling service rule comparison**: TENSION-001 documents that rules mirror sibling services, but the sibling service files (`reconcile-ads/rules.py`, `reconcile-spend/rules.py`) were not in scope for this audit. The degree of divergence since initial port is unknown.

3. **`autom8y-reconciliation` SDK internals**: `Correlator`, `ReadinessGate`, `UnifiedVerdict`, and other SDK types are used throughout but their source is not in this service's tree. SDK contract constraints are inferred from usage patterns only.

4. **IaC / Terraform**: No `.tf` files found in this repo. EventBridge schedule, Lambda IAM roles, SSM parameter paths, and CloudWatch alarm thresholds are external constraints that cannot be read here. The frozen Lambda handler interface and event schema constraints are documented from code inference only.

5. **`reconcile-spend/stubs.py`**: Mentioned in `rules.py` line 268 as the source for THREE_WAY rules 6-7, but this file is outside scope.
