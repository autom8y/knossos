---
domain: design-constraints
generated_at: "2026-03-09T00:04:44Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "8e41207"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---
# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Dual Staleness Threshold System (minutes vs seconds)

**Location**: `src/reconcile_spend/config.py` lines 85-105

`Settings` exposes two parallel freshness-gate configurations: `max_staleness_minutes` (default 45, int) and `max_staleness_seconds` (default 1800, int). The comment in `config.py` acknowledges the design tension explicitly:

> "The unit inconsistency reflects their separate origins. Unifying would break the anomaly detection threshold (45 min = 2700 sec != 1800 sec)."

The two fields serve different purposes by design: `max_staleness_minutes` was intended for `analyzer.py` (which no longer exists), and `max_staleness_seconds` feeds `readiness.py`. However, `max_staleness_minutes` is never read by any active source module — the only reference to it in the codebase is its declaration in `config.py`.

The Slack alert in `orchestrator.py` line 390 converts `max_staleness_seconds` to minutes via integer division (`int(settings.max_staleness_seconds) // 60 = 30`) but uses a parameter named `max_staleness_minutes`, meaning the alert threshold displayed to operators (30 min) differs from the declared `max_staleness_minutes` setting (45 min).

**Nature**: Naming mismatch + dead config field + divergent threshold values for what reads as equivalent concerns.

---

### TENSION-002: Protocol Downcast via `assert isinstance` in `data_service.py`

**Location**: `src/reconcile_spend/clients/data_service.py` lines 27-29

`resolve_insight_client()` returns `DataInsightProtocol`, but `create_data_client()` casts the return to the concrete `DataInsightClient` type via `assert isinstance(client, DataInsightClient)`. The comment reads: "the concrete type for circuit_breaker_state and http access".

This breaks the protocol abstraction that `resolve_insight_client` was designed to provide. The orchestrator accesses `data_client.circuit_breaker_state` (line 529) and `data_client.http` (line 488), neither of which is part of `DataInsightProtocol`. The assertion gates these accesses against a runtime failure rather than a type-safe API.

**Nature**: Layering violation. The interop protocol boundary was designed to abstract the concrete client — this code intentionally punches through it.

---

### TENSION-003: Slack Block Builder Inline in Orchestrator

**Location**: `src/reconcile_spend/orchestrator.py` lines 560-647

`_build_stale_data_alert()` and `_build_circuit_open_alert()` are private functions in `orchestrator.py` that construct Slack Block Kit payloads. The normal-path report builder lives in `report.py`. This is a layering inconsistency: degraded-path alert formatting lives in the orchestration layer, while normal-path formatting is a separate module.

The two degraded alert builders follow the same pattern (`report.py`'s `build_report` pattern) but are not colocated with it.

**Nature**: Under-abstraction. Alert builders belong in `report.py` alongside `build_report`.

---

### TENSION-004: `stubs.py` as an Architectural Documentation File

**Location**: `src/reconcile_spend/stubs.py`

`stubs.py` defines `ThreeWayComparison` and `AsanaReconciliation` — classes that are explicitly never imported and never called from active code (docstring: "They are NOT imported or called from the active reconciliation flow"). The file documents a future three-way reconciliation architecture (actual spend vs billed vs expected from Asana).

This is a zombie abstraction that has been intentionally preserved for documentation purposes. The pattern is unusual: production source code containing pure stub classes that exist only to describe future design intent.

**Nature**: Premature abstraction / architectural documentation embedded in production source. The risk is that future maintainers might not distinguish it from active code.

---

### TENSION-005: `build_all_clear_report()` is a Thin Wrapper Never Called

**Location**: `src/reconcile_spend/report.py` lines 126-141

`build_all_clear_report()` is a public function that simply calls `build_report([], ...)`. Callers in `orchestrator.py` call `build_report` directly and pass an empty anomalies list when all-clear. `build_all_clear_report` is not imported anywhere in the source tree.

**Nature**: Dead abstraction. Serves as documentation of intent but adds noise to the public API of `report.py`.

---

### TENSION-006: `ReconciliationRow` All-Optional Fields — Deferred Tightening

**Location**: `src/reconcile_spend/clients/models.py` lines 47-74

`ReconciliationRow` declares all 18 fields optional with None defaults. The docstring explicitly documents this as a deferral: "Tighten required fields after production verification confirms upstream always sends them." This means all validation of required input (e.g., `office_phone`, `spend`) happens downstream in `parse_client_records()` via explicit checks, rather than at the Pydantic model boundary.

**Nature**: Intentional design trade-off (graceful degradation vs early validation) with documented intent to tighten. The current state moves validation complexity out of the contract layer and into the orchestration layer.

---

### TENSION-007: `autom8y_events` Prototype Integration

**Location**: `src/reconcile_spend/handler.py` lines 12-46

The `autom8y_events` SDK is imported via `try/except ImportError`, making the domain event publishing path entirely optional. Comments flag this as "POC 2", "PROTOTYPE SHORTCUT", and "Production: autom8y-events will be in requirements.txt". The event period is hardcoded as `f"{period_days}d"` rather than derived from actual date range processed.

This is unfinished production integration: the code is live, fire-and-forget, but the SDK is not in `pyproject.toml` dependencies and the event payload has known inaccuracies.

**Nature**: Prototype code in the production hot path. Not gated by a feature flag — the behavior changes if `autom8y_events` happens to be installed.

---

## Trade-off Documentation

### Trade-off for TENSION-001 (Dual Staleness Thresholds)

The config comment (`config.py` lines 85-91) documents the trade-off explicitly: unifying the two thresholds would change the operational behavior of one subsystem because `45 min != 30 min`. The decision was to preserve both fields to avoid inadvertently changing the anomaly detection skip threshold. The cost is a dead config field (`max_staleness_minutes`) and operator confusion about which threshold governs which behavior.

No ADR linked. The trade-off is fully self-documented in the config comment.

---

### Trade-off for TENSION-002 (Protocol Downcast)

The `autom8y-interop` `DataInsightProtocol` abstraction does not expose `circuit_breaker_state` or `http` as protocol members. Lifting these attributes to the protocol would require changing the shared library — a cross-repo change. The `assert isinstance` is a local workaround: it preserves the protocol for dependency injection (test doubles) while allowing the concrete type to be accessed in the specific module that needs it.

No ADR linked. The comment in `data_service.py` documents the constraint.

---

### Trade-off for TENSION-006 (All-Optional `ReconciliationRow`)

Documented in `clients/models.py` docstring: the lenient validation (ADR-WS1-001) ensures the service degrades gracefully when the upstream `autom8y-data` schema evolves without a coordinated deploy. The cost is that required-field violations only surface as log warnings at parse time, not as Pydantic validation errors at the HTTP boundary. This is a conscious graceful-degradation posture.

ADR-WS1-001 is referenced in code but no file for it exists in `.ledge/decisions/` (directory is empty).

---

### Trade-off for TENSION-007 (Prototype Events)

The fire-and-forget pattern (`try/except` around `publisher.publish()`) ensures domain event publishing never blocks the primary Slack report. This is consistent with the broader pattern in the handler where metric emission is also fire-and-forget. The trade-off: the feature is deployed in a prototype state, and its correctness (event period accuracy) is documented as deferred. A failed event publish is silently swallowed.

---

## Abstraction Gap Mapping

### GAP-001: No Slack Alert Module — Degraded-Path Alerts Live in Orchestrator

As noted in TENSION-003, `_build_stale_data_alert` and `_build_circuit_open_alert` are private functions buried at lines 560-647 of `orchestrator.py`. The `report.py` module is the designated home for Slack block construction. Moving the two alert builders to `report.py` would unify all Block Kit construction in one module.

**Maintenance burden**: Any change to alert formatting requires knowing to look in `orchestrator.py` rather than `report.py`.

---

### GAP-002: `max_staleness_minutes` Setting is Orphaned

`max_staleness_minutes` (config.py line 92) is a Settings field with no consumers in active code. The field was documented as consumed by `analyzer.py`, which does not exist. The freshness gate was consolidated into `readiness.py`, which reads only `max_staleness_seconds`.

**Maintenance burden**: The orphaned field will be passed to new settings readers by maintainers who read the config comments, potentially causing confusion about which threshold to use.

---

### GAP-003: No Abstraction for Edge-Case Filtering

`parse_client_records()` in `orchestrator.py` applies four edge-case filters (E4, E5, E9, E10) as inline conditions within a single loop. There is no named type or abstraction for the filter set. Adding a new edge case (e.g., E11) requires reading the full function to understand the existing cases and find the insertion point.

**Maintenance burden**: Low — the function is small and the cases are commented. But the filter set is not independently testable as a named concept.

---

## Load-Bearing Code

### LB-001: `parse_client_records()` — Multi-Caller Edge Case Gateway

**Location**: `src/reconcile_spend/orchestrator.py` lines 40-137

`parse_client_records()` is the single gateway between raw API data (`ReconciliationRow`) and domain data (`ClientRecord`). It applies all data normalization: NaN/Inf guard (DEF-4), negative collected adjustment (E9), zero/zero skip (E5), and missing phone skip (E4).

**Callers**: `run_reconciliation()` in `orchestrator.py` (line 423). Tests in `tests/test_orchestrator.py` and `tests/test_defect_remediation.py` depend on its behavior.

**Risk of naive change**: Removing or reordering the `math.isfinite` check (E10) before the negative-collected adjustment would allow NaN to silently pass through to anomaly detection, suppressing detections without error (DEF-4 regression). The E10 guard must precede the E9 adjustment.

---

### LB-002: `detect_anomalies()` Rule Interaction Contract

**Location**: `src/reconcile_spend/rules.py` lines 210-271

`detect_anomalies()` applies all 5 rules to every record. The docstring documents the mutual exclusion logic: R1-R4 have overlapping preconditions that prevent most double-firing. Only R5 (stale account) is independent and can co-fire with any of R1-R4.

**Callers**: `run_reconciliation()` in `orchestrator.py` (line 434). Tests in `tests/test_rules.py`, `tests/test_rules_properties.py`, `tests/test_adversarial_data.py`.

**Risk of naive change**: Adding a new rule that fires on `collected > 0 and spend > 0` would co-fire with R3/R4 (both require these preconditions). The documented mutual exclusion matrix is not enforced by code — it is a semantic contract.

---

### LB-003: `get_settings()` — Module-Level Singleton

**Location**: `src/reconcile_spend/config.py` lines 137-146

`get_settings()` is decorated with `@lru_cache`, making it a process-level singleton. It is called at module level in `handler.py` (indirectly via `orchestrator.py`) and lazily in `data_service.py`. `clear_settings_cache()` exists for test isolation.

**Risk of naive change**: Any test that modifies environment variables and does not call `clear_settings_cache()` between runs will observe stale settings. The `lru_cache` survives across test boundaries unless explicitly cleared.

---

### LB-004: `enrich_anomalies()` — Lazy Import of `asana_resolve`

**Location**: `src/reconcile_spend/orchestrator.py` line 209

`batch_resolve_units` is imported inside the body of `enrich_anomalies()`:

```python
from reconcile_spend.clients.asana_resolve import batch_resolve_units
```

This lazy import is load-bearing for test isolation: replacing it with a top-level import would require all tests loading `orchestrator.py` to have `asana_resolve` importable. The current pattern allows `orchestrator.py` to be imported without the Asana client loading.

---

### LB-005: `ReconciliationResult.to_dict()` — Lambda Response Contract

**Location**: `src/reconcile_spend/models.py` lines 105-120

`to_dict()` on `ReconciliationResult` directly shapes the Lambda response body (`handler.py` line 118: `json.dumps({"success": True, **result.to_dict()})`). Any field added to `ReconciliationResult` but not to `to_dict()` will be silently absent from the Lambda response. Any field removed from `to_dict()` may break callers that parse the Lambda response.

**Callers**: `handler.py` line 118. Tests in `tests/test_handler.py` and `tests/test_contract_regression.py`.

---

## Evolution Constraints

### Area: `clients/models.py` — COORDINATED

**Evidence**: ADR-WS1-001 (referenced in code comment, no file exists). All `ReconciliationRow` fields are optional; tightening requires production verification that upstream always sends required fields. A change from `field: T | None = None` to `field: T` on any required field constitutes a breaking contract change if upstream ever omits it.

**Changeability rating**: COORDINATED — requires cross-service (autom8y-data) verification before any field is made required.

---

### Area: `rules.py` — SAFE

**Evidence**: Five pure functions with no external I/O. All inputs are `ClientRecord` (frozen dataclass) and threshold floats. Adding a new rule is additive. Changing thresholds is config-driven. No shared state.

**Changeability rating**: SAFE — new rules can be added without coordination; existing rule logic can be changed with tests.

---

### Area: `config.py` Settings fields — COORDINATED for secrets, SAFE for thresholds

**Evidence**: `service_api_key` uses `AliasChoices("RECONCILE_SPEND_SERVICE_KEY", "SERVICE_API_KEY")` for backward compatibility. Removing the `SERVICE_API_KEY` alias would break deployments that set only the old env var name.

**Changeability rating**: `service_api_key` alias chain — COORDINATED. Threshold defaults (`overbilled_threshold_pct`, `underbilled_threshold_pct`) — SAFE (env-var overridable).

---

### Area: `stubs.py` — MIGRATION

**Evidence**: `ThreeWayComparison` and `AsanaReconciliation` document a planned extension (3-way reconciliation via Asana). The stub classes define the expected interface. Activating requires adding `autom8y_events` (or Asana entity source) to `pyproject.toml`, adding rules R6/R7 to `rules.py`, and updating `parse_client_records()` to include expected-spend data.

**Changeability rating**: MIGRATION — the stub is a design document; actual implementation requires coordinated changes across multiple modules.

---

### Area: `handler.py` (autom8y_events POC) — MIGRATION

**Evidence**: Three `PROTOTYPE SHORTCUT` comments. The `autom8y_events` SDK is absent from `pyproject.toml`. The event period field (`f"{period_days}d"`) is documented as inaccurate.

**Changeability rating**: MIGRATION — promoting from POC to production requires adding the SDK as a dependency, removing the `try/except ImportError` fallback, and deriving the period from the actual date range.

---

### Area: `readiness.py` — SAFE

**Evidence**: Pure function `check_pipeline_readiness()` with no external dependencies. The 2x abort threshold is hardcoded (`abort_threshold = max_staleness_seconds * 2`) rather than configurable. This is intentional (no config field for it), but changing the multiplier requires a code change, not a config change.

**Changeability rating**: SAFE for logic changes; note that the 2x multiplier is not configurable.

---

### Area: `report.py` Slack block structure — COORDINATED

**Evidence**: `MAX_BLOCKS = 50` is hardcoded against the Slack Block Kit API limit. `RESERVED_BLOCKS = 10` is empirically derived. If Slack changes the block limit, these constants must be updated. The `_sanitize_mrkdwn_label()` function protects against Slack link label injection — removing it would be a security regression.

**Changeability rating**: COORDINATED with Slack API. Truncation logic and block counts must be verified against live Slack limits after any structural changes.

---

## Risk Zone Mapping

### RISK-001: `max_staleness_minutes` Orphaned Field — Silent Operational Misconfiguration

**Location**: `src/reconcile_spend/config.py` line 92

`max_staleness_minutes` is a Settings field that is never consumed in active code. Operators who read the config and set `MAX_STALENESS_MINUTES` in the Lambda environment will observe no effect on behavior, but will receive no warning that the setting is ignored. This is an unguarded silent no-op.

**Cross-reference**: TENSION-001

---

### RISK-002: `autom8y_events` Opt-In via Import Side Effect

**Location**: `src/reconcile_spend/handler.py` lines 15-20

If `autom8y_events` is installed in the Lambda environment (e.g., bundled transitively), event publishing activates silently with no deployment gating. There is no feature flag, no configuration toggle, and no startup validation of the bus ARN. The only guard is whether the package is importable.

The event payload has a known inaccuracy (period field is `f"{period_days}d"`, not the calendar period). Consumers of the `ReconciliationComplete` event may observe incorrect period data.

**Cross-reference**: TENSION-007

---

### RISK-003: `asana_resolve.py` Silent Empty Dict on All Errors

**Location**: `src/reconcile_spend/clients/asana_resolve.py` lines 38-77

`batch_resolve_units()` returns `{}` on request failure, non-200 response, validation error, and parse error — all four paths produce the same result: enriched anomalies without Asana URLs. The failures are logged at `WARNING` level, but the report proceeds silently with degraded enrichment.

This is intentional graceful degradation, but operators may not notice that Asana links are absent from reports unless they check CloudWatch logs. No metric is emitted when Asana resolve fails.

---

### RISK-004: `create_data_client()` `assert isinstance` — Production Crash Risk

**Location**: `src/reconcile_spend/clients/data_service.py` line 29

`assert isinstance(client, DataInsightClient)` will raise `AssertionError` in production if `resolve_insight_client` returns any non-`DataInsightClient` implementation (e.g., a stub or a future alternative implementation). Python's `assert` statements are compiled away with `-O` optimization flags. If Lambda is deployed with optimizations enabled, this guard is silently removed.

**Cross-reference**: TENSION-002

---

### RISK-005: `_build_stale_data_alert` Abort Threshold Display Bug

**Location**: `src/reconcile_spend/orchestrator.py` lines 394-396

The stale text message displayed to operators includes:

```python
f" (abort threshold: {int(settings.max_staleness_seconds * 2) // 60} min)"
```

This computes the abort threshold inline as `max_staleness_seconds * 2 // 60`. But the threshold check in `readiness.py` uses `abort_threshold = max_staleness_seconds * 2` (line 84). These are consistent. However, the Slack alert threshold display (`max_staleness_minutes=int(settings.max_staleness_seconds) // 60 = 30`) on line 390 is passed as the `max_staleness_minutes` parameter to `_build_stale_data_alert`, where it is displayed as the warn threshold. An operator reading both lines of the Slack message will see "30 min threshold, 60 min abort" — which is accurate. But if `max_staleness_seconds` is changed without also verifying the Slack message text, the displayed values will be correct via derivation.

**Risk level**: Low — the computation is derived, not duplicated. Noted as an inspection point.

---

### RISK-006: `ReconciliationRow` All-Optional Fields Mask Missing Required Data

**Location**: `src/reconcile_spend/clients/models.py` lines 57-74

If `autom8y-data` stops sending `office_phone` in its response, the Pydantic model silently accepts the row with `office_phone=None`. `parse_client_records()` will skip it with a `WARNING: missing_office_phone` log. The Lambda returns HTTP 200 with zero accounts analyzed and no anomalies. This is a total silent failure mode: billing discrepancies are present but undetected, with no error surfaced.

**Cross-reference**: TENSION-006

---

## Knowledge Gaps

1. **ADR-WS1-001**: Referenced in `clients/models.py` line 7 but no ADR file exists in `.ledge/decisions/`. The content of the ADR is inferrable from context (lenient validation posture) but the formal record is absent.

2. **`autom8y-interop` `DataInsightProtocol` interface**: Not visible in this repo. The protocol definition governs what can safely be accessed on the client without the `assert isinstance` workaround. If the protocol is ever extended to include `circuit_breaker_state`, TENSION-002 would be resolved.

3. **`PERIOD_MAP` contents**: Imported from `autom8y_interop.data`. The map determines which `period_days` values trigger the preset API path vs the date-range path. Not visible in this repo.

4. **`autom8y_events` SDK status**: The package is absent from `pyproject.toml` but may be available in the Lambda runtime. The production state of POC 2 is unknown from this codebase alone.

5. **`LambdaServiceSettingsMixin` inherited fields**: `auth_base_url`, `service_api_key_value` property, and other inherited fields from `autom8y_config.lambda_mixin` are not visible here. `config.py` references inherited fields but their defaults and constraints are in the shared library.
