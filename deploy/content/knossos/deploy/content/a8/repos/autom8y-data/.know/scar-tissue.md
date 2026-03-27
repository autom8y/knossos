---
domain: scar-tissue
generated_at: "2026-03-25T02:05:48Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "b8da042"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog Completeness

The git log contains 528 commits matching fix/bug/regression/revert keywords across 846 total commits. The following scars are cataloged from code markers (SCAR-NNN, BUG-N, DEF-N, HF-N, PSEC-N, P0-N, WS4-N, FSH) and key fix commits. The prior knowledge file (generated 2026-03-18 at commit `51f5e8d`) is extended here with post-date discoveries: appointment status conditions mis-mapping, fut rolling temporal condition, contacts table migration, DuckDB access_mode incompatibility, enrichment-aware dimension resolution, PII HTTP/2 stream truncation, DuckDB segfault tolerance, and multi-fact COUNT_DISTINCT raw grain isolation.

---

### SCAR-001 — Fact-to-Fact Cartesian Product (Metric Inflation)

**What failed:** Joining two fact tables (e.g., `leads` and `ads_insights`) in a single query produced an N*M row explosion, inflating every metric (spend, leads, scheds) by a multiplier equal to the number of rows in the secondary table. Silent — no error was raised, results appeared plausible.

**When fixed:** Guard added in commit `6a88479` (single-query fact-to-fact guard, QS-P2-06). Promoted to hard `CartesianRiskError` in commit `2ac4dbe` (Sprint 5 P1-5). Enrichment-aware bypass added in commit `ba52856` (2026-03-24) to avoid false-positive CartesianRiskError when enrichment views pre-materialize dimension columns that previously required cross-table JOINs.

**How fixed:** (1) `CartesianRiskError` raised when single-query path detects foreign fact table in `required_tables`. (2) Multi-fact queries forced through `FactResolver` split-query path. (3) `adapt_dimensions_for_fact_table()` now checks `ENRICHMENT_ADDED_COLUMNS` first (Priority 0) so enrichment-provided dimensions never trigger the CartesianRiskError guard.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/query/builder.py` (line 580)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/exceptions.py` (line 343)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/query/planner.py` (line 302)

**Regression tests:** `tests/analytics/test_multi_fact_rolling.py:454,493`, `tests/golden_master/test_analytics_golden.py:456`, `tests/analytics/test_enrichment_aware_planner.py:8`

---

### SCAR-002 — Non-Deterministic Raw Grain Query Ordering

**What failed:** Raw grain queries executed per fact table in set/dict iteration order. Python set iteration is non-deterministic across runs, causing different fact tables to be queried in different orders, producing inconsistent multi-fact rolling results — different metric values on retry.

**When fixed:** Commit `a27425a` (QS-P3-01) sorted ancestor candidates. Broader `sorted()` guards applied throughout `optimizer.py`, `engine.py`, `fact_resolver.py`.

**How fixed:** `sorted()` applied at all fact table iteration sites: `engine.py:1501`, `optimizer.py:158,586`, `fact_resolver.py:468,389`, `canonical_paths.py`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/engine.py` (line 1501)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/joins/optimizer.py` (lines 158, 586)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/query/fact_resolver.py` (lines 389, 468)

**Regression tests:** Implicit — deterministic golden master snapshot tests would catch regression.

---

### SCAR-003 — Backtick Quoting Causes DuckDB Materializer Crash

**What failed:** `_quote_identifier()` in the materializer used MySQL-style backtick quoting. DuckDB rejects backtick-quoted identifiers, causing all 29 Parquet table syncs to fail with circuit breaker permanently OPEN, making analytics unavailable.

**When fixed:** Commit `bc1d327` (RC001 — backtick quoting incident). Comment referencing this as prior art in `insight_executor.py:38`.

**How fixed:** Switched `_quote_identifier()` to ANSI SQL double-quote style (`"name"`). Two independent implementations exist — both must use double-quote style.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/materializer.py` (function `_quote_identifier`)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/insight_executor.py` (line 38)

**Regression tests:** `tests/analytics/test_reconciliation_gauges.py`

---

### SCAR-004 — COUNT_DISTINCT Overcounting in Rolling Windows

**What failed:** COUNT_DISTINCT metrics queried in rolling window contexts double-counted entities that appeared in multiple daily buckets within the window. Daily grain has one row per entity per day; rolling aggregation on SUM of daily distinctness incorrectly re-counted the same entity multiple times.

**How fixed:** Three-declaration pattern: `requires_distinct=True`, `requires_raw_grain_for_rolling=True`, `raw_grain_column` on every COUNT_DISTINCT metric. Rolling execution fetches raw grain and applies `n_unique()` across the full window, not by summing daily `COUNT(DISTINCT)` values. Documented at `library.py:435` as the canonical pattern.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/metrics/library.py` (line 435 — canonical pattern comment, applied to every COUNT_DISTINCT metric)

**Regression tests:** `tests/analytics/test_rolling_legacy_parity.py:1052`

---

### SCAR-008 — Window Aggregation Skip Set Logic Corruption

**What failed:** The window aggregation skip set logic incorrectly included raw grain metrics, display dimensions, and enrichment columns in the standard aggregation pass, producing wrong aggregated values for rolling window insights.

**How fixed:** `apply_windowed_aggregation()` at `window_aggregation.py:82` documents the skip set logic explicitly: raw grain metrics, display dimensions, and enrichment columns bypass the standard aggregation path.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/window_aggregation.py` (line 82)

**Regression tests:** `tests/analytics/test_window_aggregation.py`

---

### SCAR-009 — Composite Recomputation Ordering in Rolling Aggregation

**What failed:** Composite metrics were computed before rolling aggregation was applied to base metrics, producing wrong composite values (e.g., ROAS computed on un-rolled base values, then base metrics rolled independently).

**How fixed:** Composite recomputation now happens inside `aggregate_by_time_bucket()`, ensuring base metrics are rolled before composites are derived. Documented at `window_aggregation.py:83`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/window_aggregation.py` (line 83)

**Regression tests:** `tests/analytics/test_window_aggregation.py`

---

### SCAR-012 — `__future__ annotations` Breaks Pydantic/FastAPI Runtime Evaluation

**What failed:** `from __future__ import annotations` was added across 11 scheduling subsystem files. This defers annotation evaluation, but FastAPI/Pydantic evaluates annotations at runtime for dependency injection (function parameters) and model field resolution. Caused `NameError` / `TypeError` at startup for `AsyncSession` references.

**When fixed:** Commits `903c560` and `8dfbfe7` (dual-branch).

**How fixed:** Removed `from __future__ import annotations` from all 11 scheduling files. Where `AsyncSession` was guarded under `TYPE_CHECKING`, moved to regular import block. Rule: Never use `__future__ annotations` in files with FastAPI route handlers or Pydantic models.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/routes/scheduling.py`
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/scheduling/booking.py`
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/scheduling/engine.py` (+ 8 more scheduling files)

**Regression tests:** None automated — enforced by test-file comment rule (`SCAR-012` note in 5 test files)

---

### SCAR-013 — Window Metric Vertical Join Uses Wrong Path (DEFERRED)

**What failed / is failing:** Window metric SQL generator joins the `vertical` dimension through `chiropractors.default_vertical_id` (business default path), not through campaign hierarchy. For multi-vertical businesses, this produces incorrect vertical attribution.

**Status:** DEFERRED. No current insight combines window metrics with the `vertical` dimension, so there is no production impact. Fix required if such an insight is added.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/query/window_metric_sql.py` (lines 273-279)

**Regression tests:** None — explicitly deferred with no production exercise path.

---

### SCAR-023 — NULL-Unsafe Join Keys Drop Rows in Multi-Metric Merges

**What failed:** Polars FULL JOIN for merging window metric result DataFrames defaulted to NULL-unsafe key matching. Rows where a join key was NULL (e.g., `business_phone = NULL` for a partial result) were not unified — they produced duplicate rows or were dropped depending on Polars version.

**How fixed:** `nulls_equal=True` added to all `pl.DataFrame.join()` calls on dimension columns in multi-result merges. Pattern documented at `engine.py:1305` and `engine.py:1438`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/engine.py` (lines 1305, 1438, 1652)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/query/fact_resolver.py` (line 740)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/execution/rolling.py` (lines 784, 851)

**Regression tests:** Implicit in multi-metric golden master tests.

---

### BUG-1 — Empty Coverage Result Silent Failure (Reconciliation 500)

**What failed:** `ReconciliationCoverageProcessor` returned an empty result when no coverage data was available. Downstream code expected a non-None `data_quality` dict; receiving None caused a 500 error on the reconciliation insight endpoint.

**When fixed:** Commit `bb70e55` ("resolve reconciliation insight 500 via payments budget resolution").

**How fixed:** Empty coverage result now populates a zeroed `data_quality` dict rather than returning None.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/insights/processors/reconciliation.py`

**Regression tests:** `tests/analytics/test_reconciliation.py:665`

---

### BUG-4 — Appointments Table Contains Message Rows (Inflated Contacts Count)

**What failed:** The `appointments` MySQL table contains rows of both `type='appt'` and `type='message'`. Without filtering, DuckDB queries against the Parquet file saw message rows, inflating the `contacts` metric count.

**How fixed:** `DEFAULT_TABLE_FILTERS["appointments"] = "type = 'appt'"` applied during all sync and row-count validation operations. Filter lives in the materializer (not metric definitions) so Parquet files only contain appt rows.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/materializer.py` (line 745, `DEFAULT_TABLE_FILTERS`)

**Regression tests:** Implicit in materialization tests.

---

### BUG-6 — Activity Filter Suppresses Zero-Spend/Zero-Leads Asset Rows

**What failed:** The activity filter in `question_level_stats` suppressed asset rows with zero spend and zero leads, hiding assets that had impressions or other activity.

**How fixed:** Filter applies to `asset.frame_type` correctly, not to spend/leads absence.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/services/` (question_level_stats service)

**Regression tests:** `tests/api/services/test_question_level_stats.py:421`, `tests/api/services/test_question_level_stats_adversarial.py:613`

---

### DEF-001 — Composite Metric Multi-Level Dependency (Single-Pass Calculation Failure)

**What failed:** `CompositeCalculator.calculate()` applied all expressions in a single `with_columns()` call. In Polars, expressions in `with_columns()` are evaluated against the original DataFrame. Composite metrics depending on other composites (e.g., `pacing_ratio` depends on `expected_spend`) failed because dependencies did not exist in the original frame.

**Also covers:** API response column alias remapping — when user requested aliases (e.g., `ad_group_id` instead of canonical `adset_id`), the canonical column names were returned, not the requested aliases.

**When fixed:** Commits `aca37ca` and `94831f3` (DEF-001 followup).

**How fixed:** Multi-pass loop: iterate until no more metrics can be added per pass. Column alias remapping added at API serialization layer.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/execution/composite.py` (line 84)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/engine.py` (line 2662)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/models.py` (line 314)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/output/result.py` (line 347)

**Regression tests:** `tests/analytics/test_dimension_aliases.py:193`, `tests/api/routes/test_analytics_health.py:537`

---

### DEF-1 — Window Aggregation Auto-Discovery Overrides Cause Wrong Columns

**What failed:** `aggregate_by_time_bucket()` had auto-discovery of windowed overrides. When not explicitly provided, auto-discovery could pick up wrong columns, applying windowed aggregation to display dimensions or enrichment columns.

**How fixed:** `apply_windowed_aggregation()` always passes an explicit `windowed_overrides` dict, disabling auto-discovery.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/window_aggregation.py` (line 84)

**Regression tests:** `tests/analytics/test_window_aggregation.py:294`

---

### DEF-002 — NaN Metric Value Propagates as Zero in Health Score

**What failed:** NaN metric values propagated through normalization and became `0.0` via `max(0.0, NaN)`. NaN data points were treated as low-performing rather than missing, skewing health scores.

**How fixed:** NaN detected early, treated as `null_value`, excluded via re-weighting.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/primitives/health/score_engine.py`

**Regression tests:** `tests/analytics/primitives/health/test_score_engine_golden.py:1034`

---

### DEF-003 — Negative Metric Value Produces Negative Health Score Component

**What failed:** A negative metric value (e.g., `-10.0`) produced a negative normalized component score visible in the API response.

**How fixed:** `max(0, normalized_score)` clamp applied.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/primitives/health/score_engine.py`

**Regression tests:** `tests/analytics/primitives/health/test_score_engine_golden.py:1102`

---

### DEF-004 — Inverted Health Score Thresholds Accepted Without Validation

**What failed:** `BucketConfig` accepted `performant_threshold=25, at_risk_threshold=75` — a meaningless inverted ordering that produces incorrect health band assignments.

**How fixed:** Pydantic validator enforces `performant < underperforming < at_risk` strict ascending order.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/primitives/config/loader.py` (BucketConfig validator)

**Regression tests:** `tests/analytics/primitives/config/test_health_scoring_config.py:183`

---

### DEF-S1-002 / DEF-S1-003 — Multi-Fact Rolling Column Drop and Duplicate Columns

**What failed:** Multi-fact rolling queries dropped columns during the merge step (DEF-S1-002). Duplicate raw columns passed through when multiple metrics shared the same `raw_grain_column` on one fact table (DEF-S1-003).

**When fixed:** Commit `ae03efb`.

**How fixed:** DEF-S1-003: `dict.fromkeys()` deduplication of raw grain columns in `engine._execute_and_aggregate()`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/engine.py`

**Regression tests:** `tests/analytics/test_multi_fact_rolling.py:378,422`

---

### DEF-P2-002 — Schema Dict Mutation Across Batch Executor Partitions

**What failed:** `pre_schema` dict was passed by reference to `QueryResult.from_partition_slice()` across multiple partition slices. Downstream mutation in one partition contaminated other partitions' schema views.

**How fixed:** `schema=dict(pre_schema)` shallow copy at `unified_batch_executor.py:535`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/services/unified_batch_executor.py` (line 535)

**Regression tests:** None documented.

---

### HF-13 / HF-14 — ECS Deployment Stabilization Failures

**HF-13:** RF-C1 migrated `RedisSettings` from `REDIS_URL` to component fields (`REDIS_HOST`, etc.). ECS task definition still set `REDIS_URL`, causing `RedisSettings` to fail validation on ECS. Fixed by `model_validator` parsing `REDIS_URL` into components. Re-broken by commit `3617cc8`, restored in `ee8fdb5`.

**HF-14:** `SlowAPIMiddleware` reads `app.state.limiter` but code only set `app.state.infra.limiter`, causing `AttributeError` on ALB health checks and preventing ECS deployment stabilization. Fixed in `07a7524`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/core/config.py` (HF-13: `_parse_redis_url_fallback` model_validator)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/main.py` (HF-14: both `app.state.limiter` and `app.state.infra.limiter`)

**Regression tests:** `tests/core/test_config_env_hygiene.py` (HF-13); none for HF-14.

---

### PSEC-01 — Scheduling Overlap Check Uses String Comparison Instead of UTC

**What failed:** `_check_overlap_nonlocking()` compared appointment datetimes as strings. Two datetimes representing the same UTC instant with different timezone offsets were not detected as overlapping.

**When fixed:** Commit `e9db28b`.

**How fixed:** `_parse_appointment_datetime()` normalizes all datetimes to UTC before comparison.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/core/repositories/appointments.py` (lines 35, 128, 156)

**Regression tests:** `tests/scheduling/test_adversarial.py`

---

### P0-1 — SQL Injection via table_prefix in Reconciliation Gauges

**What failed:** `refresh_reconciliation_gauges()` constructed SQL with unquoted `table_prefix`, enabling SQL injection if prefix was attacker-controlled (sourced from config).

**When fixed:** Commit `c44eff6`.

**How fixed:** `table_prefix` identifier quoted via ANSI double-quote escaping in `insight_executor.py:31` `_quote_identifier()`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/insight_executor.py` (line 31)

**Regression tests:** `tests/analytics/test_reconciliation_gauges.py`

---

### P0-2 — Cache Key Does Not Include Parquet Data Version

**What failed:** `ConnectionRouter` cache key did not include the Parquet data version. On materialization swap, cache served stale results.

**When fixed:** Commit `c44eff6`.

**How fixed:** `ConnectionRouter.get_current_version()` reads the `current` symlink target; version included in cache key at check and store sites.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/` (connection router)

**Regression tests:** Implicit in cache tests.

---

### FSH — Filter Semantic Schism (Silent Wrong SQL)

**What failed:** Three overlapping scope-dimension translation mechanisms existed independently. Scope dimensions (`vertical`, `business`) in the `filters` dict were only translated by some code paths — others produced invalid SQL silently (e.g., `WHERE vertical = 'chiro'` instead of `WHERE key = 'chiro'`).

**When fixed:** Commit `126ae2e`.

**How fixed:** Single canonical `normalize_scope_filter_keys()` in `constants.py`. `merge_filters()` extended with `scope_dimension_map` param. Defense-in-depth normalization in `_build_query_filters()`. 68 new tests.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/constants.py`
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/engine.py`

**Regression tests:** `tests/analytics/test_filter_merge.py` (26 tests), `tests/analytics/test_fsh_adversarial.py` (42 tests)

---

### WS4-C3 — Empty-String Phone Match in Materializer JOINs

**What failed:** `messages` and `calls` JOIN queries matched on Twilio phone (`sent_from`/`sent_to`). Empty-string phone values (`twil_phone = ''`) generated incorrect `business_phone` matches against any message with an empty sender/recipient.

**When fixed:** Commit `159bd80`.

**How fixed:** `AND sent_from != ''` / `AND sent_to != ''` guards added to all 4 JOIN clauses in messages and calls materialization queries.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/materializer.py`

**Regression tests:** Implicit in materializer tests.

---

### WS4-C5 — Corrupted Parquet File Served After Copy-Forward Without Verification

**What failed:** `_copy_previous_version()` used `shutil.copy2()` without verifying copy integrity. Corrupted copies were served to readers without detection.

**When fixed:** Commit `b06ab36`.

**How fixed:** SHA256 checksum comparison after `shutil.copy2()`. Corrupted copies discarded.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/materializer.py`

**Regression tests:** None — no test exercises the corrupted-copy scenario.

---

### WS1-C6 — Contacts Metric COUNT Instead of COUNT_DISTINCT

**What failed:** The `contacts` metric used `COUNT(id)` instead of `COUNT_DISTINCT(id)`. In rolling window queries, the same appointment appeared across multiple daily buckets, causing COUNT to overcount.

**When fixed:** Commit `4e18b6f`.

**How fixed:** Applied three-declaration pattern (`requires_distinct`, `requires_raw_grain_for_rolling`, `raw_grain_column`) in `library.py`.

**Note:** Subsequently superseded by a more fundamental fix — contacts metric migrated from `appointments` table to `messages` table (commit `3bcb221`, 2026-03-24). See CONTACTS-MIGRATION below.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/metrics/library.py`

**Regression tests:** `tests/analytics/test_rolling_legacy_parity.py:1052`

---

### CONTACTS-MIGRATION — Contacts Metric Table Changed (Appointments → Messages)

**What failed (semantic):** The `contacts` metric was counting distinct appointment IDs from the `appointments` table. This was semantically wrong — `contacts` should count unique inbound text message conversations (distinct `lead_phone`/`business_phone` pairs from the `messages` table with direction `LIKE 'inbound%'`).

**When fixed:** Commit `3bcb221` (2026-03-24, feat: migrate contacts from appointments to messages table).

**How fixed:** Metric redefined: `table=messages`, `column=contact_key`, `fact_table=messages`, `condition="messages.direction LIKE 'inbound%'"`. Enrichment view for messages now computes `contact_key = lead_phone || '::' || business_phone`. LIKE pattern support added to `_parse_sql_condition_to_polars()` for Polars rolling path.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/metrics/library.py` (line 428-447)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/execution/rolling.py` (LIKE pattern support)

**Regression tests:** Implicit in rolling parity tests.

---

### APPOINTMENT-STATUS — Five Status Conditions Semantically Wrong

**What failed:** Five appointment metric conditions were semantically wrong, discovered via stakeholder interview against actual Parquet data (5,503 appointments, 10 distinct status values):
- `fut` (future appointments): only counted `scheduled` status; missed `confirmed`, `requested`, `rescheduled`
- `ns` (no-show): was counting cancellations as no-shows (inverted — used `status = 'cancelled'`)
- Cancellations: were using `status = 'reschedule'` instead of including both `cancelled` and `reschedule`

**When fixed:** Commit `822983e` (2026-03-24).

**How fixed:** Status conditions corrected:
- `fut`: `status IN ('scheduled', 'confirmed', 'requested', 'rescheduled') AND start_datetime > NOW()`
- `ns`: `status IN ('no-show', 'no_show', 'rescheduling')`
- cancellations: `status IN ('cancelled', 'reschedule')`

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/metrics/library.py` (lines 548, 606, 637)

**Regression tests:** None explicit — golden master snapshots updated.

---

### FUT-ROLLING-TEMPORAL — Temporal Condition Fails in Polars Rolling Path

**What failed:** The `fut` metric condition `CAST(appointments.start_datetime AS TIMESTAMP) > NOW()` was evaluated in the Polars rolling path, but the rolling path had no `NOW()` reference. `_parse_sql_condition_to_polars()` returned `None` for temporal conditions, causing `fut` to be excluded from rolling window results.

**When fixed:** Commit `d6a95a2` (2026-03-24, feat: optimize LOCAL mode init and fix fut rolling temporal condition).

**How fixed:** `_condition_has_temporal()` added to detect NOW()/CURRENT_TIMESTAMP/CURRENT_DATE patterns. Temporal conditions now use a reference date passed through `InsightExecutionContext`. `_parse_sql_condition_to_polars()` interprets `> NOW()` as `> today`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/execution/rolling.py`

**Regression tests:** `tests/analytics/test_rolling_legacy_parity.py` (fut-specific cases)

---

### DUCKDB-CONCURRENCY — Concurrent Queries Crash Single-Connection DuckDB

**What failed:** 12 concurrent analytics queries crashed the single-worker uvicorn via NULL pointer dereference in DuckDB's `FetchArrowTable`. Root cause: `QueryOrchestrator` received both a pooled connection AND the `connection_provider`, enabling re-acquisition during multi-table parallel queries that violated DuckDB's single-connection threading guarantee.

**When fixed:** Commit `a36c72a` (2026-03-19).

**How fixed:** `asyncio.Lock` added to both connection adapters (`execute_to_polars_safe`) in `backend.py`. `ConnectionPool` returns independent connections per request for concurrent isolation.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/backend.py` (lines 248, 392, `_query_lock`)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/connection_pool.py`

**Regression tests:** None — performance/concurrency tests in `tests/performance/`

---

### DUCKDB-CTE-CRASH — DuckDB 1.4.4 ARM64 Crashes on Complex Multi-CTE VIEW Queries

**What failed:** DuckDB 1.4.4 on ARM64 crashed with NULL unique_ptr dereference when complex multi-CTE queries referenced VIEW definitions in the query planner. The crash was deterministic on 4 aggregated table specs.

**When fixed:** Commit `cf9bee4` (2026-03-19).

**How fixed:** 4 aggregated table specs (`business_offers_budget`, `payments_budget`, `ad_optimizations_budget`, `platform_assets_agg`) materialized as physical `TABLE` (CREATE TABLE AS) instead of VIEWs. DuckDB resolves physical tables without re-planning the VIEW SQL.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/materializer.py`

**Regression tests:** None — DuckDB 1.4.4-specific workaround.

---

### DUCKDB-ACCESS-MODE — READ_ONLY Incompatible with ATTACH Mode

**What failed:** `duckdb.connect(access_mode='read_only')` is incompatible with DuckDB's ATTACH mode — DuckDB cannot attach MySQL databases in read-only mode. This caused connection failures when using ATTACH-mode queries.

**When fixed:** Commit `d191afc` (2026-03-24).

**How fixed:** `access_mode='read_only'` parameter removed from `ConnectionPool` connection creation in `connection_pool.py`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/connection_pool.py` (lines 264-272)

**Regression tests:** None.

---

### DUCKDB-SEGFAULT — MySQL Scanner SIGSEGV During Connection Teardown

**What failed:** DuckDB's MySQL scanner SIGSEGV/SIGABRT during `connection.__aexit__()` after materialization completes (exit codes 139/134). The Parquet data is already written; the crash is cosmetic and non-fatal, but it masked success output.

**When fixed:** Commit `59eb016` (2026-03-23).

**How fixed:** Two-layer fix: `materialize_local.py` flushes stdout before `__aexit__` and wraps cleanup in `try/except`. `justfile` setup-local and sync-offline tolerate exit codes 139/134 with user-facing note.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/` (materialize_local.py)
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/justfile`

**Regression tests:** None — infra-level, not testable in CI.

---

### DUCKDB-DATE-TRUNC — DuckDB 1.5.0 DATE_TRUNC Returns TIMESTAMP Instead of DATE

**What failed:** DuckDB 1.5.0 changed `DATE_TRUNC('week', ...)` to return `TIMESTAMP` instead of `DATE`, causing golden master snapshot mismatches for week-grain time dimensions.

**When fixed:** Commit `3610263` (2026-03-23).

**How fixed:** Explicit `CAST(... AS DATE)` wrapper added to `DATE_TRUNC('week', ...)` in time dimension SQL, matching the pattern already used in `datetime_utils.py`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/dimensions/time_dimensions.py`

**Regression tests:** Golden master snapshots in `tests/golden_master/__snapshots__/`

---

### PII-HTTP2-TRUNCATION — Content-Length Mismatch Truncates PII-Masked Responses

**What failed:** The PII masking middleware modified response body length (by masking phone numbers and emails) without updating the `Content-Length` header. HTTP/2 clients received truncated responses because the declared length was shorter than the actual masked body.

**When fixed:** Commit `6da0d57` (2026-03-24).

**How fixed:** `Content-Length` header stripped from PII-masked responses (HTTP/2 uses chunked transfer encoding and does not require it). Stripping is the correct fix rather than recalculating, as the body length change from masking is unpredictable.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/pii.py` (line 151, `CRITICAL` comment + header strip logic)

**Regression tests:** None explicit.

---

### SQL-INJECTION-FILTER — Numeric Bypass in Filter Pipeline Allows Unquoted Values

**What failed:** `_needs_quoting()` returned `False` for int/float-parseable filter values (e.g., `"12345"`), allowing those values to enter DuckDB SQL as unquoted literals. This was a SQL injection vector for numeric-looking attacker-controlled filter values.

**When fixed:** Commit `bf110c3` (2026-03-23).

**How fixed:** `_needs_quoting()` now always returns `True`. DuckDB handles implicit type casting. `_parse_in_clause()` and `_parse_single_filter()` updated to always single-quote with `''` escaping.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/entity_source/filter.py`

**Regression tests:** `tests/security/` (26-payload adversarial test suite)

---

### ENRICHMENT-SCHEMA-DRIFT — ATTACH Mode Enrichment View Column Mismatches

**What failed:** Three enrichment view bugs caused cascading failures in production MySQL ATTACH mode:
1. `messages` enrichment view: `PARTITION BY m.id` referenced nonexistent column (messages has composite PK `lead_phone, created, uri`)
2. `leads` enrichment: `EXCLUDE(ltv)` failed because MySQL column is already named `client_ltv`
3. `calls` enrichment: similar column name drift

**When fixed:** Commit `b255d26` (2026-03-23).

**How fixed:** `messages` partition key corrected. `leads` enrichment removed entirely (column doesn't need enrichment). `calls` enrichment corrected.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/enrichment_views.py` (or equivalent enrichment view definition file)

**Regression tests:** None — ATTACH mode tests require production credentials.

---

### DUCKDB-MEMORY-LIMIT — DuckDB Crashes Instead of Spilling to Disk When Memory Full

**What failed:** Without a configured `temp_directory`, DuckDB would crash instead of spilling when memory was exhausted during large query execution.

**When fixed:** Commit `0f03add` (DuckDB disk spilling via temp_directory fallback).

**How fixed:** `SET temp_directory = '{temp_dir}'` applied to every new DuckDB connection at pool creation time. Default to `/tmp` if `settings.duckdb.temp_directory` is not configured. Documented as `CRITICAL` at `connection_pool.py:270`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/analytics/core/infra/connection_pool.py` (line 270)

**Regression tests:** None — infrastructure constraint.

---

### STARTUP-POOL-RACE — ConnectionPool Blocks ECS Startup Healthcheck

**What failed:** `ConnectionPool` initialization was synchronous in the lifespan handler. ECS ALB healthcheck fired before pool initialization completed, causing health failures and preventing deployment stabilization.

**When fixed:** Commit `94b71fe` (fix: defer ConnectionPool initialization to background task).

**How fixed:** Pool initialization deferred to `asyncio.create_task()` in lifespan. Health endpoint returns degraded (not failing) state while pool is initializing.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/api/main.py`

**Regression tests:** None — deployment-environment specific.

---

### MESSAGES-COMPOSITE-PK — Wrong PK on Messages Table Causes 500 on GET /messages

**What failed:** The `messages` table was given a single `id` field as primary key, breaking `GET /messages`. The table has a composite primary key `(lead_phone, created, uri)` reflecting its MySQL schema. Setting `id` as PK caused SQL queries to use `id` as the entity identifier, producing 500 errors.

**When fixed:** Commit `dd3e979` (fix: restore Messages composite primary key).

**How fixed:** Reverted `id` field to non-PK. Composite PK `(lead_phone, created, uri)` restored.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/core/models.py`

**Regression tests:** None explicit — ATTACH mode integration test would catch.

---

### CRITICAL-ORM-001 — Offer.offer_id Self-Referential Foreign Key (RESOLVED)

**What failed:** `Offer.offer_id` had a self-referential foreign key pointing to `offers.offer_id`. This is a meaningless circular reference that prevents DB schema creation and ORM introspection.

**Status:** RESOLVED. Regression guard at `tests/core/orm/test_orm_integrity_awareness.py:88`.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/core/models.py`

---

### CRITICAL-ORM-002 through CRITICAL-ORM-004 — ORM Type/FK Issues (OPEN xfail)

Three ORM issues identified in `SCAN-orm-dead-code-phase3.md` remain unresolved as `xfail(strict=True)` tests:

- **CRITICAL-002:** `BusinessOffer.offer_id` is `str` but references `Offer.offer_id` which is `Optional[int]` (type mismatch)
- **CRITICAL-003:** `AdAccount.guid` is `Optional` primary key (nullable PK)
- **CRITICAL-004:** `PlatformAsset.ad_account_id` FK references `ad_accounts.ad_account_id` but AdAccount's PK is `guid`

**Fix prerequisite for all three:** Verify production schema column types before fixing.

**Regression tests:** `tests/core/orm/test_orm_integrity_awareness.py:95,116,132` — `xfail(strict=True)` tripwires.

---

### DEF-FACTORY-001 — polyfactory Random Int PKs Cause Birthday-Paradox Collisions

**What failed:** `polyfactory` generates random ints in ~0-10000 range for int PKs. At `rows_per_table >= 50`, PK collision probability became non-negligible, causing FK integrity failures in test fixtures.

**How fixed:** `_next_pk()` sequential counter overrides random int PKs at fixture construction time. FK-coordinated PKs (composite PKs that are also FK columns) are preserved.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/core/factories.py` (line 1043)

**Regression tests:** `tests/test_factory_system.py:272` (T-8: PK Scaling)

---

### DEF-FACTORY-002 — polyfactory Generates None for Optional FK Fields (Orphaned FKs)

**What failed:** `Offer.offer_id` is `int | None`. `polyfactory` legitimately generated `None`. Child tables (`adsets`, `chiropractors`, `assets`, `business_offers`) then received `None` as the FK value, creating orphaned references.

**How fixed:** `_OfferOverrideFactory` overrides `offer_id` with a sequential counter ensuring it is always non-None.

**Fix files:**
- `/Users/tomtenuta/Code/a8/repos/autom8y-data/src/autom8_data/core/factories.py` (line 652, 871, `_OfferOverrideFactory`)

**Regression tests:** `tests/test_factory_system.py:302` (T-9: FK Orphan Check)

---

## Category Coverage

| ID | Scar | Category |
|----|------|----------|
| SCAR-001 | Fact-to-fact Cartesian product | Data Corruption / Query Correctness |
| SCAR-002 | Non-deterministic set iteration | Race Condition / Non-determinism |
| SCAR-003 | Backtick quoting crashes DuckDB | Integration Failure (DuckDB dialect) |
| SCAR-004 | COUNT_DISTINCT overcounting in rolling | Data Corruption (inflated counts) |
| SCAR-008 | Window skip set logic | Query Correctness |
| SCAR-009 | Composite recomputation ordering | Query Correctness |
| SCAR-012 | `__future__ annotations` / Pydantic runtime | Integration Failure (Python runtime) |
| SCAR-013 | Window vertical join wrong path (deferred) | Query Correctness (latent) |
| SCAR-023 | NULL-unsafe join key drops rows | Query Correctness |
| BUG-1 | Empty coverage result 500 | Integration Failure |
| BUG-4 | Appointments table type filter | Data Corruption (inflated counts) |
| BUG-6 | Activity filter suppresses zero rows | Query Correctness |
| DEF-001 | Composite metric single-pass | Query Correctness |
| DEF-1 | Window auto-discovery overrides | Query Correctness |
| DEF-002 | NaN propagates as zero in health | Data Corruption (health scores) |
| DEF-003 | Negative health score not clamped | Data Corruption (health scores) |
| DEF-004 | Inverted health thresholds accepted | Schema / Validation |
| DEF-S1-002/003 | Multi-fact rolling column drop/dupe | Query Correctness |
| DEF-P2-002 | Schema dict mutation across partitions | Race Condition / State Mutation |
| DEF-FAC-001 | polyfactory int PK birthday collisions | Test Infrastructure |
| DEF-FAC-002 | polyfactory None FK orphans | Test Infrastructure |
| HF-13 | REDIS_URL removed from ECS env | Config Drift |
| HF-14 | SlowAPI state key mismatch | Integration Failure (middleware) |
| PSEC-01 | Timezone string comparison overlap | Security / Correctness |
| P0-1 | SQL injection via table_prefix | Security |
| P0-2 | Cache stale after materialization swap | Performance Cliff / Stale Cache |
| FSH | Filter semantic schism | Integration Failure (silent wrong SQL) |
| WS4-C3 | Empty-string phone JOIN | Data Corruption (false matches) |
| WS4-C5 | Corrupted Parquet served without detection | Data Corruption |
| WS1-C6 | contacts COUNT overcounting | Data Corruption (inflated counts) |
| CONTACTS-MIGRATION | contacts metric table changed (semantic) | Data Correctness (semantic) |
| APPOINTMENT-STATUS | 5 appointment status conditions wrong | Data Corruption (business logic) |
| FUT-ROLLING-TEMPORAL | Temporal condition fails in Polars rolling | Query Correctness |
| DUCKDB-CONCURRENCY | Concurrent queries crash single connection | Stability / Crash |
| DUCKDB-CTE-CRASH | DuckDB 1.4.4 ARM64 CTE VIEW crash | Stability / Crash (version-specific) |
| DUCKDB-ACCESS-MODE | READ_ONLY incompatible with ATTACH | Integration Failure (DuckDB API) |
| DUCKDB-SEGFAULT | MySQL scanner crash on teardown | Stability / Crash (non-fatal) |
| DUCKDB-DATE-TRUNC | DuckDB 1.5.0 DATE_TRUNC type change | Integration Failure (version-specific) |
| PII-HTTP2-TRUNCATION | Content-Length mismatch truncates PII response | Integration Failure (HTTP/2) |
| SQL-INJECTION-FILTER | Numeric bypass allows unquoted filter values | Security |
| ENRICHMENT-SCHEMA-DRIFT | ATTACH mode enrichment view column mismatches | Integration Failure (schema) |
| DUCKDB-MEMORY-LIMIT | No temp_directory causes crash vs. spill | Stability / Crash |
| STARTUP-POOL-RACE | ConnectionPool blocks ECS healthcheck | Availability / Deployment |
| MESSAGES-COMPOSITE-PK | Wrong PK on messages causes 500 | Integration Failure (ORM/schema) |
| CRITICAL-ORM-001..004 | ORM type/FK integrity issues | Schema / Validation |
| SQL-INJECTION-FILTER | Numeric bypass in filter pipeline | Security |

**Categories present:**
1. Data Corruption — 10 scars
2. Query Correctness — 10 scars
3. Integration Failure — 10 scars (DuckDB dialect, Python runtime, middleware, HTTP/2, schema)
4. Security — 3 scars
5. Stability / Crash — 5 scars
6. Race Condition / Non-determinism — 2 scars
7. Config Drift — 1 scar
8. Performance Cliff / Stale Cache — 1 scar
9. Schema / Validation — 2 scars
10. Test Infrastructure — 2 scars
11. Data Correctness (semantic) — 2 scars
12. Availability / Deployment — 1 scar

12 distinct categories present.

---

## Fix-Location Mapping

| Scar | Fix File(s) | Key Function(s) |
|------|-------------|-----------------|
| SCAR-001 | `src/autom8_data/analytics/core/query/builder.py:580`, `src/autom8_data/analytics/core/infra/exceptions.py:343`, `src/autom8_data/analytics/core/query/planner.py:302` | `_validate_fact_tables()`, `CartesianRiskError`, `adapt_dimensions_for_fact_table()` |
| SCAR-002 | `src/autom8_data/analytics/engine.py:1501`, `src/autom8_data/analytics/core/joins/optimizer.py:158,586`, `src/autom8_data/analytics/core/query/fact_resolver.py:389,468` | `sorted()` at all fact iteration sites |
| SCAR-003 | `src/autom8_data/analytics/core/infra/materializer.py`, `src/autom8_data/analytics/insight_executor.py:38` | `_quote_identifier()` (two independent copies) |
| SCAR-004 | `src/autom8_data/analytics/core/metrics/library.py:435` | `requires_raw_grain_for_rolling` three-declaration pattern |
| SCAR-008 | `src/autom8_data/analytics/window_aggregation.py:82` | `apply_windowed_aggregation()` |
| SCAR-009 | `src/autom8_data/analytics/window_aggregation.py:83` | `aggregate_by_time_bucket()` |
| SCAR-012 | `src/autom8_data/api/routes/scheduling.py` + 10 scheduling files | Module-level import removal |
| SCAR-013 (deferred) | `src/autom8_data/analytics/core/query/window_metric_sql.py:273` | `_build_vertical_join()` — deferred comment |
| SCAR-023 | `src/autom8_data/analytics/engine.py:1305,1438,1652`, `src/autom8_data/analytics/core/query/fact_resolver.py:740`, `src/autom8_data/analytics/core/execution/rolling.py:784,851` | `nulls_equal=True` on all merge joins |
| BUG-1 | `src/autom8_data/analytics/insights/processors/reconciliation.py` | `ReconciliationCoverageProcessor.process()` |
| BUG-4 | `src/autom8_data/analytics/core/infra/materializer.py:745` | `DEFAULT_TABLE_FILTERS` |
| BUG-6 | `src/autom8_data/api/services/` (question_level_stats) | activity filter logic |
| DEF-001 | `src/autom8_data/analytics/core/execution/composite.py:84`, `src/autom8_data/analytics/engine.py:2662`, `src/autom8_data/api/models.py:314`, `src/autom8_data/analytics/core/output/result.py:347` | Multi-pass composite + alias remapping |
| DEF-1 | `src/autom8_data/analytics/window_aggregation.py:84` | `apply_windowed_aggregation()` explicit overrides |
| DEF-002 | `src/autom8_data/analytics/primitives/health/score_engine.py` | `HealthScoreEngine.compute_score()` |
| DEF-003 | `src/autom8_data/analytics/primitives/health/score_engine.py` | `max(0, normalized_score)` clamp |
| DEF-004 | `src/autom8_data/analytics/primitives/config/loader.py` | `BucketConfig` validator |
| DEF-S1-002/003 | `src/autom8_data/analytics/engine.py` | `_execute_and_aggregate()` deduplication |
| DEF-P2-002 | `src/autom8_data/analytics/services/unified_batch_executor.py:535` | `dict(pre_schema)` shallow copy |
| DEF-FAC-001/002 | `src/autom8_data/core/factories.py:1043,652,871` | `_next_pk()`, `_OfferOverrideFactory` |
| HF-13 | `src/autom8_data/core/config.py` | `RedisSettings._parse_redis_url_fallback()` |
| HF-14 | `src/autom8_data/api/main.py` | `create_app()` — both `app.state.limiter` paths |
| PSEC-01 | `src/autom8_data/core/repositories/appointments.py:35,128,156` | `_parse_appointment_datetime()` |
| P0-1 | `src/autom8_data/analytics/insight_executor.py:31` | `_quote_identifier()` |
| P0-2 | `src/autom8_data/analytics/core/infra/` (connection_router.py) | `get_current_version()` in cache key |
| FSH | `src/autom8_data/analytics/core/infra/constants.py`, `src/autom8_data/analytics/engine.py` | `normalize_scope_filter_keys()` |
| WS4-C3 | `src/autom8_data/analytics/core/infra/materializer.py` | Messages/calls JOIN guards |
| WS4-C5 | `src/autom8_data/analytics/core/infra/materializer.py` | `_copy_previous_version()` + SHA256 |
| WS1-C6 | `src/autom8_data/analytics/core/metrics/library.py` | contacts metric three-declaration pattern |
| CONTACTS-MIGRATION | `src/autom8_data/analytics/core/metrics/library.py:428`, `src/autom8_data/analytics/core/execution/rolling.py` | Contacts redefined to messages table + LIKE support |
| APPOINTMENT-STATUS | `src/autom8_data/analytics/core/metrics/library.py:548,606,637` | fut/ns/cancellation status conditions |
| FUT-ROLLING-TEMPORAL | `src/autom8_data/analytics/core/execution/rolling.py` | `_condition_has_temporal()`, temporal reference date |
| DUCKDB-CONCURRENCY | `src/autom8_data/analytics/core/infra/backend.py:248,392`, `src/autom8_data/analytics/core/infra/connection_pool.py` | `asyncio.Lock` on execute methods |
| DUCKDB-CTE-CRASH | `src/autom8_data/analytics/core/infra/materializer.py` | CREATE TABLE AS instead of VIEW |
| DUCKDB-ACCESS-MODE | `src/autom8_data/analytics/core/infra/connection_pool.py` | Removed `access_mode='read_only'` |
| DUCKDB-SEGFAULT | `src/autom8_data/` (materialize_local.py), `justfile` | try/except cleanup + exit code tolerance |
| DUCKDB-DATE-TRUNC | `src/autom8_data/analytics/core/dimensions/time_dimensions.py` | `CAST(DATE_TRUNC(...) AS DATE)` |
| PII-HTTP2-TRUNCATION | `src/autom8_data/api/pii.py:151` | Strip `Content-Length` after body masking |
| SQL-INJECTION-FILTER | `src/autom8_data/analytics/entity_source/filter.py` | `_needs_quoting()` always True |
| ENRICHMENT-SCHEMA-DRIFT | `src/autom8_data/analytics/core/infra/enrichment_views.py` | Partition key and column corrections |
| DUCKDB-MEMORY-LIMIT | `src/autom8_data/analytics/core/infra/connection_pool.py:270` | `SET temp_directory` on every connection |
| STARTUP-POOL-RACE | `src/autom8_data/api/main.py` | `asyncio.create_task()` for pool init |
| MESSAGES-COMPOSITE-PK | `src/autom8_data/core/models.py` | Messages model PK restoration |
| CRITICAL-ORM-001 | `src/autom8_data/core/models.py` | Offer self-referential FK removal |
| CRITICAL-ORM-002..004 | `src/autom8_data/core/models.py:220,695,818` | Open — production schema verification needed |

---

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Location | Regression Test |
|------|-------------------|----------|-----------------|
| SCAR-001 | `CartesianRiskError` + enrichment-aware Priority-0 bypass | `builder.py:580`, `exceptions.py:343`, `planner.py:302` | `test_analytics_golden.py:456`, `test_enrichment_aware_planner.py:8` |
| SCAR-002 | `sorted()` at all fact-table/set iteration sites | `engine.py:1501`, `optimizer.py:158,586`, `fact_resolver.py:468,389` | Implicit — golden master snapshots |
| SCAR-003 | ANSI double-quote `_quote_identifier()` (two independent copies) | `materializer.py`, `insight_executor.py:38` | `test_reconciliation_gauges.py` |
| SCAR-004 | Three-declaration pattern on every COUNT_DISTINCT metric | `metrics/library.py:435` | `test_rolling_legacy_parity.py:1052` |
| SCAR-008 | Explicit skip set for raw grain metrics, display dims, enrichment cols | `window_aggregation.py:82` | `test_window_aggregation.py` |
| SCAR-009 | Composite recomputation inside `aggregate_by_time_bucket()` | `window_aggregation.py:83` | `test_window_aggregation.py` |
| SCAR-012 | No `__future__ annotations` in scheduling/FastAPI files (5 test-file comments) | All scheduling files | None automated — comment-enforced |
| SCAR-013 | Deferred comment at join site | `window_metric_sql.py:279` | None |
| SCAR-023 | `nulls_equal=True` on all dimension-column merges | `engine.py:1305,1438,1652`, `fact_resolver.py:740` | Implicit in multi-metric tests |
| BUG-1 | Zeroed `data_quality` dict on empty coverage | `reconciliation.py` | `test_reconciliation.py:665` |
| BUG-4 | `DEFAULT_TABLE_FILTERS["appointments"] = "type = 'appt'"` | `materializer.py:745` | Implicit in materializer tests |
| BUG-6 | Activity filter on `frame_type`, not spend/leads | question_level_stats service | `test_question_level_stats.py:421` |
| DEF-001 | Multi-pass composite calculator + column alias remapping | `composite.py:84`, `result.py:347`, `models.py:314` | `test_dimension_aliases.py:193` |
| DEF-1 | Explicit `windowed_overrides` dict, no auto-discovery | `window_aggregation.py:84` | `test_window_aggregation.py:294` |
| DEF-002 | NaN detected early, excluded via re-weighting | health/score_engine.py | `test_score_engine_golden.py:1034` |
| DEF-003 | `max(0, normalized_score)` clamp | health/score_engine.py | `test_score_engine_golden.py:1102` |
| DEF-004 | Pydantic validator: `performant < underperforming < at_risk` | `primitives/config/loader.py` | `test_health_scoring_config.py:183` |
| DEF-S1-002/003 | `dict.fromkeys()` deduplication of raw grain columns | `engine.py` | `test_multi_fact_rolling.py:378,422` |
| DEF-P2-002 | `schema=dict(pre_schema)` shallow copy | `unified_batch_executor.py:535` | None |
| DEF-FAC-001 | `_next_pk()` sequential counter overrides random int PKs | `factories.py:1043` | `test_factory_system.py:272` |
| DEF-FAC-002 | `_OfferOverrideFactory` ensures non-None offer_id | `factories.py:652,871` | `test_factory_system.py:302` |
| HF-13 | `_parse_redis_url_fallback` model_validator on RedisSettings | `config.py` | `test_config_env_hygiene.py` |
| HF-14 | Both `app.state.limiter` and `app.state.infra.limiter` set | `main.py` | None |
| PSEC-01 | `_parse_appointment_datetime()` UTC normalization | `appointments.py:35` | `tests/scheduling/test_adversarial.py` |
| P0-1 | `_quote_identifier()` on all config-sourced SQL identifiers | `insight_executor.py:31` | `test_reconciliation_gauges.py` |
| P0-2 | Parquet version in cache key | `connection_router.py` | Implicit in cache tests |
| FSH | `normalize_scope_filter_keys()` + defense-in-depth in `_build_query_filters()` | `constants.py`, `engine.py` | `test_filter_merge.py` (26), `test_fsh_adversarial.py` (42) |
| WS4-C3 | `AND sent_from != ''` guards in JOIN clauses | `materializer.py` | Implicit |
| WS4-C5 | SHA256 checksum after copy | `materializer.py` | None |
| WS1-C6 | Three-declaration pattern on contacts | `metrics/library.py` | `test_rolling_legacy_parity.py:1052` |
| DUCKDB-CONCURRENCY | `asyncio.Lock` on every DuckDB execute call | `backend.py:248,392` | None automated |
| DUCKDB-CTE-CRASH | Physical TABLE not VIEW for aggregated specs | `materializer.py` | None — ARM64 specific |
| DUCKDB-ACCESS-MODE | No `access_mode='read_only'` in ATTACH mode | `connection_pool.py` | None |
| DUCKDB-SEGFAULT | `try/except` on `__aexit__` + exit code tolerance | `materialize_local.py`, `justfile` | None |
| DUCKDB-DATE-TRUNC | `CAST(DATE_TRUNC('week',...) AS DATE)` | `time_dimensions.py` | Golden master snapshots |
| DUCKDB-MEMORY-LIMIT | `SET temp_directory` on every connection (`CRITICAL` comment) | `connection_pool.py:270` | None |
| PII-HTTP2-TRUNCATION | Strip `Content-Length` after PII body modification | `pii.py` | None |
| SQL-INJECTION-FILTER | `_needs_quoting()` always True — adversarial test suite | `entity_source/filter.py` | `tests/security/` (26 payloads) |
| APPOINTMENT-STATUS | Status conditions validated against Parquet data — golden master | `metrics/library.py:548,606,637` | Golden master snapshots |
| FUT-ROLLING-TEMPORAL | `_condition_has_temporal()` guard + reference date | `execution/rolling.py` | `test_rolling_legacy_parity.py` |
| STARTUP-POOL-RACE | `asyncio.create_task()` for deferred pool init | `main.py` | None |
| CRITICAL-ORM-001 | Regression guard test (xfail removed post-fix) | `test_orm_integrity_awareness.py:88` | `test_orm_integrity_awareness.py:85` |
| CRITICAL-ORM-002..004 | `xfail(strict=True)` tripwires for open issues | `test_orm_integrity_awareness.py:94,115,131` | Tripwire (will fail when unexpectedly fixed) |

**Scars with no regression test:** SCAR-012, SCAR-013, DEF-P2-002, HF-14, WS4-C5, DUCKDB-CONCURRENCY, DUCKDB-CTE-CRASH, DUCKDB-ACCESS-MODE, DUCKDB-SEGFAULT, DUCKDB-MEMORY-LIMIT, PII-HTTP2-TRUNCATION, STARTUP-POOL-RACE, ENRICHMENT-SCHEMA-DRIFT.

---

## Agent-Relevance Tagging

| Scar | Relevant Agents | Why |
|------|-----------------|-----|
| SCAR-001 | **query-strategist**, **analytics-qa** | Multi-fact query modification risks re-enabling Cartesian path. Enrichment columns bypass the guard via Priority-0 check — understand the bypass before adding new enrichment dimensions. |
| SCAR-002 | **query-strategist** | All new dict/set iteration over fact tables must use `sorted()`. |
| SCAR-003 | **query-strategist**, **data-quality-sentinel** | All SQL identifier construction from config must use ANSI double-quote `_quote_identifier()`. Backtick causes materializer circuit-breaker OPEN. |
| SCAR-004 | **metric-architect**, **analytics-qa** | Every new COUNT_DISTINCT metric that queries in rolling window contexts needs three-declaration pattern. |
| SCAR-008, SCAR-009 | **insight-engineer**, **analytics-qa** | New rolling window insights must ensure composite recomputation inside `aggregate_by_time_bucket()` and explicit skip sets. |
| SCAR-012 | **insight-engineer**, **metric-architect** | Never add `from __future__ import annotations` to files with FastAPI route handlers or Pydantic model fields. |
| SCAR-013 | **insight-engineer** | Any insight combining window metrics with `vertical` dimension will hit wrong join path. Re-implement using campaign hierarchy before adding such an insight. |
| SCAR-023 | **query-strategist** | All Polars DataFrame merges on dimension columns must use `nulls_equal=True` to prevent NULL-key row loss. |
| BUG-1 | **data-quality-sentinel**, **insight-engineer** | Post-processors must always return structured (possibly zeroed) response, never None. |
| BUG-4 | **data-quality-sentinel** | `DEFAULT_TABLE_FILTERS` is canonical for table-level row filters. Never rely on metric definitions to filter appointment types. |
| BUG-6 | **analytics-qa** | Activity filter behavior in question_level_stats is a regression-tested invariant. Zero-spend/zero-leads rows must survive. |
| DEF-001 | **insight-engineer**, **analytics-qa** | Composite metrics depending on other composites need multi-pass calculation. API response must use alias remapping. |
| DEF-1 | **insight-engineer** | Always pass explicit `windowed_overrides` to `aggregate_by_time_bucket()`. |
| DEF-002, DEF-003, DEF-004 | **data-quality-sentinel** | Health score engine has three documented edge case fixes. Run golden master suite after any health metric changes. |
| DEF-S1-002/003 | **analytics-qa**, **query-strategist** | Multi-fact rolling queries require raw grain column deduplication via `dict.fromkeys()`. |
| DEF-P2-002 | **insight-engineer** | Always shallow-copy `pre_schema` before passing to `QueryResult.from_partition_slice()`. |
| DEF-FAC-001/002 | **analytics-qa** | Factory system uses sequential PKs and override factories. Do not bypass these with direct `polyfactory` calls on models with int/Optional PKs. |
| HF-13 | **data-quality-sentinel** | `RedisSettings` backward-compat `REDIS_URL` fallback must not be removed without ECS task definition coordination. |
| HF-14 | **data-quality-sentinel** | Rate limiter must be set on both `app.state.limiter` and `app.state.infra.limiter`. |
| PSEC-01 | **analytics-qa** (scheduling adversarial) | Appointment datetime comparison must always go through `_parse_appointment_datetime()` UTC normalization. |
| P0-1 | **query-strategist** | All config-sourced SQL identifiers (table_prefix, column names) must use `_quote_identifier()`. |
| P0-2 | **data-quality-sentinel** | Cache invalidation must include Parquet data version. |
| FSH | **query-strategist**, **insight-engineer** | All filter dicts must go through `normalize_scope_filter_keys()`. Never hardcode scope names as SQL column names. |
| WS4-C3 | **data-quality-sentinel** | New materializer JOINs on phone columns must guard against empty-string values. |
| WS4-C5 | **data-quality-sentinel** | File copies in materializer should verify checksum post-copy. |
| WS1-C6, CONTACTS-MIGRATION | **metric-architect** | `contacts` metric is now on messages table, not appointments. Don't revert or duplicate on appointments. |
| APPOINTMENT-STATUS | **metric-architect**, **data-quality-sentinel** | Status conditions were validated against real Parquet data. Future status taxonomy changes require re-validation against production data. |
| FUT-ROLLING-TEMPORAL | **query-strategist**, **metric-architect** | Temporal conditions in metric definitions require `_condition_has_temporal()` handling in the Polars rolling path. |
| DUCKDB-CONCURRENCY | **data-quality-sentinel** | DuckDB does not support concurrent queries on a single connection. All DuckDB execution paths must hold `asyncio.Lock`. |
| DUCKDB-CTE-CRASH | **data-quality-sentinel** | Aggregated table specs must be `CREATE TABLE AS` (not VIEWs) to avoid DuckDB 1.4.4 ARM64 query planner crash. |
| DUCKDB-ACCESS-MODE | **data-quality-sentinel** | Never pass `access_mode='read_only'` to DuckDB when ATTACH mode is required. |
| DUCKDB-DATE-TRUNC | **query-strategist** | Always `CAST(DATE_TRUNC('week',...) AS DATE)` — DuckDB 1.5.0 changed return type. |
| PII-HTTP2-TRUNCATION | **data-quality-sentinel** | After any body modification in middleware, strip or recalculate `Content-Length` header. |
| SQL-INJECTION-FILTER | **query-strategist** | Never bypass `_needs_quoting()`. All filter values are quoted unconditionally. |
| ENRICHMENT-SCHEMA-DRIFT | **data-quality-sentinel** | Enrichment view column names must exactly match MySQL column names (including ORM alias renames). Validate in ATTACH mode before deploying. |
| DUCKDB-MEMORY-LIMIT | **data-quality-sentinel** | Every new DuckDB connection must set `temp_directory` before executing large queries. |
| STARTUP-POOL-RACE | **data-quality-sentinel** | Pool/resource initialization that involves DB connections must be deferred via `asyncio.create_task()` in FastAPI lifespan. |
| CRITICAL-ORM-002..004 | **metric-architect**, **query-strategist** | Three open ORM integrity issues (type mismatch, nullable PK, broken FK) require production schema verification before resolving. Do not use `BusinessOffer.offer_id`, `AdAccount.guid`, or `PlatformAsset.ad_account_id` as reliable join keys. |

---

## Knowledge Gaps

1. **SCAR-004 through SCAR-007, SCAR-010, SCAR-011:** SCAR-004 has a code marker (`library.py:435`) but no corresponding named git commit. SCAR-005, SCAR-006, SCAR-007, SCAR-010, SCAR-011 have no code markers or commits found — may have been assigned and never committed, exist in a different repo, or exist in squashed/cleaned history.

2. **BUG-2, BUG-3, BUG-5:** These BUG numbers are missing from source and test markers. BUG-1, BUG-4, BUG-6 are documented; intermediate numbers are unaccounted for.

3. **DEF-S1-004, DEF-S1-005:** Referenced in commit `ae03efb` body but no corresponding code markers found in source files.

4. **DEF-S2-001 (grouping_dimensions case normalization):** Referenced only in a test docstring. No source code marker found; fix location not confirmed.

5. **HF-1 and HF-2:** Absent from git log — may predate visible history.

6. **SCAR-013 defensive test:** No test covers the wrong-vertical-attribution-in-window-metric-with-vertical-dimension scenario. This gap is intentional but unguarded.

7. **DEF-P2-002 regression test:** No test exercises schema mutation across batch executor partitions.

8. **ENRICHMENT-SCHEMA-DRIFT regression test:** ATTACH mode requires production credentials — no CI-executable test covers enrichment view column correctness.

9. **DUCKDB-SEGFAULT toleration:** The segfault is tolerated, not fixed. A future DuckDB version may resolve the MySQL scanner teardown crash, at which point the `try/except` cleanup and `justfile` exit code tolerance can be removed.

10. **CRITICAL-ORM-002..004 open issues:** These ORM integrity issues are intentionally deferred pending production schema verification. The `xfail(strict=True)` tripwires will alert the team when unexpectedly fixed.

---
