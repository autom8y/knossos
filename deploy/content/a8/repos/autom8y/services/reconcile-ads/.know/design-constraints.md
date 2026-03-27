---
domain: design-constraints
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

# Codebase Design Constraints

## Tension Catalog

**TENSION-001: Duplicate Name Encoding — SDK Boundary Without Import**

The canonical `NameEncoding` (bullet-separated campaign/ad-group name format) lives in `autom8y-ads` (a separate service repository, not a pip dependency). `reconcile-ads` cannot import from `autom8y-ads` so it replicates the decode logic inline in `src/reconcile_ads/joiner.py` (lines 36-91). This is explicitly acknowledged in a comment at line 34-38:

> "The canonical NameEncoding lives in autom8y-ads (separate repo, not an importable dependency). We replicate the decode logic here using the same bullet-separator pattern and field positions."

If `autom8y-ads` changes the separator character or field order without a coordinated update to `joiner.py`, the reconciliation engine silently misdirects or drops all data. There is no contract test or schema version pinning between the two.

---

**TENSION-002: Dual Verdict System — Legacy Fields + Unified SDK Verdicts**

`Finding` (in `src/reconcile_ads/models.py`, lines 198-250) carries two overlapping verdict representations simultaneously:

- `status_verdict: StatusVerdict | None` (legacy scalar)
- `budget_verdict: BudgetVerdict | None` (legacy scalar)
- `delivery_verdict: DeliveryVerdict | None` (legacy scalar)
- `verdicts: dict[VerdictAxis, UnifiedVerdict]` (unified SDK dict)

The `severity` property (lines 232-250) uses the SDK path when `verdicts` is populated but falls back to a hardcoded `severity_map` dict against legacy `status_verdict` values. The comment reads "Legacy fallback for backward compatibility during migration." Both code paths are active simultaneously. The report builder in `src/reconcile_ads/report.py` reads the legacy scalar fields (lines 256-260), not the unified dict — meaning the SDK migration is half-complete. Changing either path without updating the other produces divergent behavior.

---

**TENSION-003: Single-Account MVP Hardcode**

`meta_account_id` is a required single-string field in Settings (`src/reconcile_ads/config.py`, line 64): "Meta ad account ID (single account MVP)." The entire join engine, ghost detection, and report builder are designed around one account. Multi-account support would require index keying changes, report grouping changes, and metric dimension changes. The label "MVP" signals intent to expand, but the code treats single-account as structural.

---

**TENSION-004: Naming Mismatch — Settings Field vs. Env Var**

`config.py` field `ads_service_url` (line 46) maps to the env var `ADS_SERVICE_URL` via default Pydantic field naming, but `secretspec.toml` documents the canonical env var as `AUTOM8Y_ADS_URL` (Tier 3 naming per `ADR-ENV-NAMING-CONVENTION`). Similarly, `asana_service_url` vs. `AUTOM8Y_ASANA_URL`. The Pydantic `model_config` uses `env_prefix=""` and `case_sensitive=False`, so `ADS_SERVICE_URL` and `AUTOM8Y_ADS_URL` are both accepted as long as one is present — but this creates an implicit two-name mapping that is not reflected in the field definitions via `AliasChoices`. If the Terraform injection uses `AUTOM8Y_ADS_URL` but a developer's `.env` uses `ADS_SERVICE_URL`, both work, masking the drift.

---

**TENSION-005: Fetcher Error Hierarchy vs. Service Error Hierarchy**

`fetcher.py` defines `AdsServiceUnavailableError` and `AsanaServiceUnavailableError` (lines 34-51) as plain `Exception` subclasses, not as subclasses of `ReconcileAdsError` from `errors.py`. Meanwhile, `errors.py` defines `FetchError(ReconcileAdsError)` (line 22) which is never used anywhere in the codebase. Two parallel error hierarchies exist: one in `errors.py` (unused for its intended purpose) and one in `fetcher.py` (actively used). An agent looking at `errors.py` would expect `FetchError` to be the operational error but it is inert.

---

**TENSION-006: Weekly Ad Spend Dedup — First-Wins Semantic with Silent Override**

`dedup_weekly_ad_spend` in `joiner.py` (lines 524-566) takes the first offer's `weekly_ad_spend` value per `(office_phone, vertical)` key and discards subsequent ones (E17). The function is documented as "first offer's value per key" but the ordering of `offers` list is not guaranteed — it depends on the concatenation order of active + activating results from `fetcher.py` (line 222: `combined_data = active_result.data + activating_result.data`). If the Asana query returns results in a different order across runs, the budget comparison baseline can shift without any configuration change.

---

**TENSION-007: Raw Tree Dict Threading vs. Typed Models**

The `ads_tree` response is passed as `dict[str, Any]` throughout the pipeline (orchestrator, joiner, rules, readiness). The same dict is re-walked in three separate places: `build_campaign_index`, `build_ad_set_index`, and `_evaluate_delivery_health` (rules.py line 333), each with their own `item.get("campaign", {})` access patterns. There is no typed schema enforced at the boundary — the raw dict shape is load-bearing but undeclared. A change to the `autom8y-ads` tree response format (e.g., renaming `items` to `campaigns`) would produce silent empty results rather than a type error.

---

## Trade-off Documentation

**TRADE-OFF-001: Fail-Open Slack Posting**

The `_safe_slack_post` wrapper in `orchestrator.py` (lines 419-458) catches all exceptions and logs them without re-raising. The design decision (labeled E15 in the comment at line 458) is: Lambda should return 200 even when Slack fails. The trade-off is operational visibility vs. reliability signal — a Slack failure is silent from Lambda's perspective (CloudWatch logs capture it but no alarm fires on it unless explicitly configured). The reconciliation result is computed correctly and the Lambda reports success regardless of whether the report reached the channel.

**TRADE-OFF-002: Event Publish Best-Effort**

`_publish_complete_event` in `orchestrator.py` (lines 382-416) catches all exceptions with only a `log.debug("event_publish_failed")`. EventBridge events are pure best-effort — no retry, no dead letter, no alarm. Consumers of `CampaignAlignmentComplete` events have no guarantee of delivery on any given run.

**TRADE-OFF-003: Truncation Causes Abort, Not Partial Report**

When `has_more_campaigns=True` (ads data truncated) or Asana returns fewer rows than `total_count`, the pipeline aborts entirely via the readiness gate (orchestrator lines 176-192) rather than running with partial data. The trade-off is correctness vs. availability — a run with 100/120 campaigns would produce false ghost detections and false missing-campaign detections, so the service chooses to post an alert and skip rather than produce misleading verdicts.

**TRADE-OFF-004: Inline Name Decode vs. Shared Library**

Rather than extracting name encoding/decoding into a shared package (e.g., `autom8y-naming`), the decode logic is replicated inline (TENSION-001). The trade-off is: zero cross-service import friction vs. schema drift risk. The chosen path is operationally simpler but creates a maintenance coupling that is invisible to standard dependency tooling.

**TRADE-OFF-005: Pydantic `extra="ignore"` on All Data Models**

`OfferRecord`, `AdRecord`, `AdGroupRecord`, `CampaignRecord` all use `model_config = ConfigDict(extra="ignore")`. This means any new fields returned by upstream services are silently dropped at the parse boundary. The trade-off is: clean forward-compatibility vs. missed data. An upstream service adding a field that this service needs (e.g., a new `pause_reason` on campaigns) would be ignored without error.

---

## Abstraction Gap Mapping

**MISSING-001: No Typed Tree Schema**

The `ActiveCampaignTreeResponse` is consumed as `dict[str, Any]` everywhere. There is no `AdsTreeResponse` Pydantic model in `models.py` that reflects the full tree structure (`items`, `campaign`, `children`, `ad_group`, `ads` nesting). The untyped boundary is the largest structural gap — three separate modules re-implement the same `.get("campaign", {})` access pattern.

**MISSING-002: No Join Strategy Protocol**

The two-level join is implemented as a concrete function `execute_join` that bundles Join A (campaign-level) and Join B (ad-set-level) into one linear loop. There is no abstraction that allows new join levels or alternate join strategies (e.g., a future TikTok join on different keys). Adding a Join C would require modifying `execute_join`, `JoinResult`, and all downstream consumers.

**MISSING-003: No Report Renderer Interface**

`_render_finding_block` in `report.py` is a module-level private function passed as a callback to `ReconciliationReportBuilder`. This callback pattern exists but is not typed as a `Protocol` — the renderer contract is implicit. Any change to `Finding` structure requires manually tracing which renderer fields are accessed.

**PREMATURE-001: `ReconciliationResult.to_dict()` as Manual Serialization**

`ReconciliationResult` is a Pydantic `BaseModel` but implements a manual `to_dict()` method (lines 293-309) that explicitly enumerates fields rather than using `model_dump()`. This was presumably done to exclude `findings` from the Lambda response body (findings are logged separately). The manual dict omits `financial_summary` and `findings`, which are present in the model but not in the Lambda response. This is not documented as intentional — an agent adding a new field to `ReconciliationResult` would not know whether to add it to `to_dict()`.

---

## Load-Bearing Code Identification

**LOAD-BEARING-001: `_decode_campaign_name` / `_decode_ad_group_name` — joiner.py lines 70-91**

These two functions are the only bridge between raw Meta campaign names (bullet-separated strings) and the join keys `(office_phone, vertical)` and `(office_phone, offer_id)`. Every campaign in the Meta tree becomes either indexed or dropped depending on these decoders. Any change to the field positions in `CampaignNameFields` or `AdGroupNameFields` NamedTuples will silently remap or drop all campaign data. The bullet separator `\u2022` (U+2022) is hardcoded in three places: `joiner.py` line 40, `rules.py` line 423, and `conftest.py` line 28. These must remain synchronized.

**LOAD-BEARING-002: `dedup_weekly_ad_spend` — joiner.py lines 524-566**

This function produces the `deduped_budgets` dict that is the Asana side of every budget comparison. If it returns wrong values (wrong key, wrong first-wins selection), every budget verdict is wrong. The budget comparison formula in `rule_budget_alignment` (`rules.py` lines 147-154) computes `variance_pct = abs(meta_weekly - asana_weekly) / asana_weekly * 100` — division by `asana_weekly` which comes directly from this dedup function. A zero value from this function produces `BUDGET_UNAVAILABLE`, not a crash, but a wrong non-zero value silently produces a miscalibrated verdict.

**LOAD-BEARING-003: `check_pipeline_readiness` — readiness.py**

This function is the only gate preventing stale or truncated data from flowing into the join engine. If it returns `PASS` or `WARN` on data that should be `FAIL`, the reconciliation runs on bad data and produces wrong verdicts without any alert. The function delegates entirely to the SDK `ReadinessGate` — meaning a bug in the SDK `ReadinessGate` directly affects production output without any local defensive layer.

**LOAD-BEARING-004: `get_settings` lru_cache — config.py lines 123-126**

The `@lru_cache` on `get_settings()` means settings are resolved once per Lambda container lifecycle. If a Lambda container is warm-reused after an SSM secret rotation, the old secret persists in memory until the container is killed. The `clear_settings_cache()` function exists but is only called in tests (via the autouse fixture in `conftest.py` line 37). There is no runtime mechanism to force cache invalidation on container reuse.

**LOAD-BEARING-005: `_safe_slack_post` in orchestrator.py (lines 419-458)**

Every Slack post in the service (degraded alert, stale alert, truncation alert, all-clear, findings report) routes through this single wrapper. Changing this function's exception handling changes the behavior for all alert types simultaneously. The "Lambda returns 200 regardless" contract (E15) is enforced here and nowhere else.

---

## Evolution Constraint Documentation

**EVOLVE-001: Multi-Account Support Requires Structural Refactor**

Adding a second Meta account ID requires: (1) changing `meta_account_id: str` to a list type, (2) changing the orchestrator to iterate accounts, (3) keying the `CampaignIndex` with a third dimension (account), (4) separating ghost detection per account, (5) changing metrics to emit per-account dimensions. This is not safely addable without coordinated changes across `config.py`, `orchestrator.py`, `joiner.py`, `models.py`, and `metrics.py`.

**EVOLVE-002: Adding a New Verdict Axis Requires Touching 5 Files**

To add a new verdict axis (e.g., `QUALITY` or `AUDIENCE`): (1) add the verdict enum to `autom8y-reconciliation` SDK, (2) add it to `VerdictAxis` in the SDK, (3) add a rule function in `rules.py`, (4) add the finding field to `Finding` in `models.py`, (5) add a renderer branch in `_render_finding_block` in `report.py`, (6) add a section in `_build_sections`, (7) add a metric in `metrics.py`. The dual verdict system (TENSION-002) means step 4 must add both a legacy scalar field and a `VerdictAxis` entry.

**EVOLVE-003: Name Encoding Version Bumps Must Be Coordinated**

If `autom8y-ads` increments the campaign name format (e.g., adds an 8th field, changes field positions), `joiner.py` must be updated in the same deployment. There is no version negotiation, no schema version field in the tree response, and no test that validates against a live `autom8y-ads` response. The only safety net is that `_decode_campaign_name` pads short names with empty strings (E16), so a shorter format fails silently rather than crashing — which means wrong-key joins rather than exceptions.

**EVOLVE-004: TikTok Platform Extension**

`OfferRecord.platforms: list[str] | None` already carries `["Meta", "TikTok"]` as possible values. The existing join engine only operates against the Meta tree (`fetch_ads_tree`). Adding TikTok would require: a new fetch function, a new index type, new join levels, new verdict rules, and new Slack sections. The data model anticipates it but the pipeline does not.

**EVOLVE-005: `max_campaigns` Hard Cap at 500**

`config.py` line 103: `le=500`. This Pydantic validator is the absolute ceiling on campaigns fetched per run. Above 500, the ads service would need to support cursor-based pagination and the orchestrator would need to loop. The current readiness gate aborts on `has_more_campaigns=True`, so scale past 500 active campaigns triggers a permanent abort loop until the cap is raised or pagination is added.

---

## Risk Zone Mapping

**RISK-001: Unguarded Name Decode Drift**

There are no contract tests between `reconcile-ads` and `autom8y-ads` asserting that the name format is stable. If the `autom8y-ads` team changes the name encoding, `reconcile-ads` will produce wrong join keys with no exception raised. The readiness gate does not catch this — data is complete and fresh but semantically wrong.

**RISK-002: Budget Thresholds in Config Without Business Sign-Off Gate**

`budget_drift_threshold_pct` and `budget_mismatch_threshold_pct` are plain env vars with no validation beyond `ge=0`. Setting `BUDGET_DRIFT_THRESHOLD_PCT=0` causes every offer to produce a DRIFT verdict (any nonzero variance). Setting `BUDGET_MISMATCH_THRESHOLD_PCT=1` floods the Slack channel. There is no guard against operationally dangerous threshold values.

**RISK-003: Asana Offer Ordering Nondeterminism**

As documented in TENSION-006, the first-wins dedup for `weekly_ad_spend` depends on the ordering of `active_result.data + activating_result.data`. If the Asana query response ordering changes (e.g., due to a database query optimization), budget verdict baselines shift across runs with identical underlying data.

**RISK-004: Delivery Health Evaluated Outside the Join**

`_evaluate_delivery_health` in `rules.py` (line 333) re-walks the full `ads_tree` dict independently of the join results. It creates new `CampaignRecord` and `AdGroupRecord` objects from the raw dict, bypassing the indexed decode path. This means delivery health findings can reference campaign IDs that failed campaign-level decode (and are excluded from join), producing findings for campaigns whose phone/vertical are `"unknown"`. These findings appear in the Slack report with `(unknown, None)` as the identifier — actionable only if the CloudWatch log is consulted.

**RISK-005: Single Lambda Invocation Owns Full State**

The entire pipeline (fetch, readiness gate, join, rules, report, metrics, EventBridge) runs inside one Lambda invocation with no checkpointing. A Lambda timeout (default 15 min max for Lambda) mid-pipeline would produce no Slack report, no metrics, and no EventBridge event. There is no partial-completion recovery mechanism.

**RISK-006: `lru_cache` Settings on Warm Lambda — Stale Secret Risk**

Documented in LOAD-BEARING-004. On a warm Lambda container, `get_settings()` returns the cached `Settings` object. If AWS rotates the `SERVICE_API_KEY` secret while the container is warm, every subsequent invocation uses the old key until the container is recycled. The auth failure would surface as `AdsServiceUnavailableError` or `AsanaServiceUnavailableError`, triggering a degraded alert — but the root cause (stale cached secret) is not logged.

---

## Knowledge Gaps

1. The `autom8y-ads` service's actual `ActiveCampaignTreeResponse` schema is not visible in this codebase. The exact field names, nested structure, and any versioning of the tree response format would require reading the `autom8y-ads` repository.

2. The `autom8y-reconciliation` SDK internals (`ReadinessGate`, `ReconciliationReportBuilder`, `ReconciliationMetrics`) are external to this service. The behavior of these SDK types is load-bearing but their implementations are not auditable here.

3. Terraform/Lambda configuration (timeout, memory, EventBridge schedule rate) is in the `devops/` and `services/reconcile-ads/just/` tree and was not examined. The actual `max_campaigns=100` cap interaction with the Lambda timeout is not observable from source alone.

4. No evidence of an ADR-RCA-001 document referenced in `joiner.py` line 8. Its location (`.ledge/decisions/` or elsewhere) is unknown — may document additional design rationale for the two-level join.
