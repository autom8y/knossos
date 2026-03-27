---
domain: scar-tissue
generated_at: "2026-03-18T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "51f5e8d"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

## Failure Catalog Completeness

The git log contains 358 commits matching fix/bug/regression/revert keywords. The following scars have been identified from code markers (SCAR-NNN, BUG-N, DEF-N, DEF-SN-NNN, HF-N) and key fix commits.

### SCAR-001 — Fact-to-Fact Cartesian Product (Metric Inflation)

**What failed:** Joining two fact tables (e.g., `leads` and `ads_insights`) in a single query produced an N*M row explosion, inflating every metric (spend, leads, scheds) by a multiplier equal to the number of rows in the secondary table. Silent — no error was raised.

**When fixed:** Promoted from warning to hard failure in commit `2ac4dbe` (Sprint 5 P1-5 production readiness). Earlier guard added in `6a88479` (single-query fact-to-fact guard, QS-P2-06).

**How fixed:** (1) `CartesianRiskError` exception raised when single-query path detects a foreign fact table in `required_tables`. (2) Multi-fact queries forced through `FactResolver` split-query path. (3) Hard guard also present in `QueryBuilder._validate_fact_tables()`.

---

### SCAR-002 — Non-Deterministic Raw Grain Query Ordering

**What failed:** Raw grain queries executed per fact table in set/dict iteration order. Python set iteration is non-deterministic across runs, causing different fact tables to be queried in different orders, producing inconsistent multi-fact rolling results (different metric values on retry).

**When fixed:** Commit `a27425a` (QS-P3-01) sorted ancestor candidates; broader sorted() guards applied throughout optimizer.py and engine.py.

**How fixed:** `sorted()` applied at all fact table iteration sites. Specifically at `engine.py:1346` (`for ft, ft_raw_metrics in sorted(raw_by_fact.items())`).

---

### SCAR-003 — Backtick Quoting Causes DuckDB Materializer Crash

**What failed:** `_quote_identifier()` in the materializer used MySQL-style backtick quoting. DuckDB rejects backtick-quoted identifiers, causing all 29 Parquet table syncs to fail. The circuit breaker went permanently OPEN, making analytics unavailable.

**When fixed:** Commit `bc1d327` (RC001 — backtick quoting incident).

**How fixed:** Switched `_quote_identifier()` to ANSI SQL double-quote style (`"name"`). Also adds the pattern to `insight_executor.py:31` `_quote_identifier()` with an explicit comment referencing SCAR-003 as prior art. An identical function exists in both `materializer.py:1793` and `insight_executor.py:31`.

---

### SCAR-008 — Window Aggregation Skip Set Logic Corruption

**What failed:** The window aggregation skip set logic incorrectly included raw grain metrics, display dimensions, and enrichment columns in the aggregation pass, producing wrong aggregated values for rolling window insights.

**When/how fixed:** Encapsulated in `window_aggregation.py:82`. The `apply_windowed_aggregation()` function documents the skip set logic explicitly: raw grain metrics, display dimensions, and enrichment columns bypass the standard aggregation path.

---

### SCAR-009 — Composite Recomputation Ordering in Rolling Aggregation

**What failed:** Composite metrics were being computed before rolling aggregation was applied to base metrics, producing wrong composite values (e.g., ROAS = composite(spend, revenue) computed on un-rolled values, then base metrics rolled independently).

**When/how fixed:** Composite recomputation now happens **inside** `aggregate_by_time_bucket()`, ensuring base metrics are rolled before composite metrics are derived. Documented at `window_aggregation.py:83`.

---

### SCAR-012 — `__future__ annotations` Breaks Pydantic Runtime Evaluation

**What failed:** `from __future__ import annotations` was added across 11 scheduling subsystem files. This defers annotation evaluation, but FastAPI/Pydantic evaluates annotations at runtime for dependency injection (function parameters) and model field resolution. This caused `NameError` / `TypeError` at startup for `AsyncSession` references.

**When fixed:** Commits `903c560` and `8dfbfe7` (identical fix, dual-branch).

**How fixed:** Removed `from __future__ import annotations` from all 11 scheduling files. For files where `AsyncSession` was guarded under `TYPE_CHECKING`, moved the import to the regular block so it is available at runtime.

---

### SCAR-013 — Window Metric Vertical Join Uses Wrong Path (DEFERRED)

**What failed / is failing:** Window metric SQL generator joins the `vertical` dimension through `chiropractors.default_vertical_id` (the business default path), not through the campaign hierarchy. For multi-vertical businesses, this produces incorrect vertical attribution in window metric queries.

**Status:** Documented as DEFERRED. No current insight combines window metrics with the `vertical` dimension, so there is no production impact. Fix is required if such an insight is added.

**Location:** `src/autom8_data/analytics/core/query/window_metric_sql.py:273-279`.

---

### BUG-1 — Empty Coverage Result Silent Failure

**What failed:** `ReconciliationCoverageProcessor` returned an empty result when no coverage data was available. Downstream code expected a non-None `data_quality` dict; receiving None caused a `500` error on the reconciliation insight endpoint.

**When fixed:** Git evidence: commit `bb70e55` ("resolve reconciliation insight 500 via payments budget resolution").

**How fixed:** Empty coverage result now populates a zeroed `data_quality` dict rather than returning None. Test class `TestReconciliationCoverage` at `tests/analytics/test_reconciliation.py:664`.

---

### BUG-4 — Appointments Table Contains Message Rows (Inflated Contacts Count)

**What failed:** The `appointments` table in MySQL contains rows of both `type='appt'` and `type='message'`. Without filtering, DuckDB queries against the Parquet file saw message rows, inflating the `contacts` metric count.

**When fixed:** Applied as part of materializer hardening (see `materializer.py:741-748`).

**How fixed:** `DEFAULT_TABLE_FILTERS["appointments"] = "type = 'appt'"` applied during all sync and row-count validation operations. The filter lives in the materializer, not in metric definitions, so Parquet files themselves only contain `appt` rows.

---

### BUG-6 — Activity Filter Suppresses Zero-Spend/Zero-Leads Asset Rows

**What failed:** The activity filter applied to `question_level_stats` suppressed asset rows with zero spend and zero leads, incorrectly hiding assets that had received impressions or other activity.

**How fixed:** Filter applies to `asset.frame_type` correctly. Tests at `tests/api/services/test_question_level_stats.py:421` and `tests/api/services/test_question_level_stats_adversarial.py:613`.

---

### DEF-001 — Composite Metric Multi-Level Dependency (Single-Pass Calculation Failure)

**What failed:** `CompositeCalculator.calculate()` applied all expressions in a single `with_columns()` call. In Polars, expressions in `with_columns()` are evaluated against the **original** DataFrame. Composite metrics depending on other composite metrics (e.g., `pacing_ratio` depends on `expected_spend`) would fail because their dependencies did not exist in the original frame.

**When fixed:** Commits `aca37ca` and `94831f3` (DEF-001 followup).

**How fixed:** Multi-pass loop: iterate until no more metrics can be added per pass, ensuring dependency order is respected. Also: response column alias remapping added at API serialization layer (`api/models.py:294`, `engine.py:2211`).

---

### DEF-1 — Window Aggregation Auto-Discovery Overrides Cause Wrong Columns

**What failed:** `aggregate_by_time_bucket()` had auto-discovery of windowed overrides. When not explicitly provided, auto-discovery could pick up wrong columns, applying windowed aggregation to display dimensions or enrichment columns.

**How fixed:** `apply_windowed_aggregation()` always passes an explicit `windowed_overrides` dict to disable auto-discovery. Documented at `window_aggregation.py:84`. Test at `tests/analytics/test_window_aggregation.py:294`.

---

### DEF-002 — NaN Metric Value Propagates as Zero in Health Score (Incorrect Weighting)

**What failed:** NaN metric values propagated through normalization and became `0.0` via `max(0.0, NaN)`. This caused NaN data points to be treated as low-performing rather than missing, skewing health scores.

**How fixed:** NaN detected early and treated as `null_value`, excluded via re-weighting. Regression tests at `tests/analytics/primitives/health/test_score_engine_golden.py:1034`.

---

### DEF-003 — Negative Metric Value Produces Negative Health Score Component

**What failed:** A negative metric value (e.g., `-10.0`) produced a negative normalized component score (e.g., `-25.0`), which was visible in the API response.

**How fixed:** `max(0, normalized_score)` clamp applied. Regression tests at `tests/analytics/primitives/health/test_score_engine_golden.py:1102`.

---

### DEF-004 — Inverted Health Score Thresholds Not Rejected

**What failed:** `BucketConfig` accepted `performant_threshold=25, at_risk_threshold=75`, a meaningless inverted ordering that would produce incorrect health band assignments.

**How fixed:** Pydantic validator enforces `performant < underperforming < at_risk` strict ascending order. Regression test at `tests/analytics/primitives/config/test_health_scoring_config.py:183`.

---

### DEF-S1-002 / DEF-S1-003 — Multi-Fact Rolling Column Drop and Duplicate Columns

**What failed:** Multi-fact rolling queries dropped columns during the merge step (DEF-S1-002). Also, duplicate raw columns passed through when multiple metrics shared the same `raw_grain_column` on one fact table (DEF-S1-003).

**When fixed:** Commit `ae03efb` (regression guard + QA defect fixes for multi-fact rolling).

**How fixed:** DEF-S1-003: deduplicate raw grain columns via `dict.fromkeys()` in `engine._execute_and_aggregate()`. Tests at `tests/analytics/test_multi_fact_rolling.py:12,378`.

---

### DEF-P2-002 — Schema Dict Mutation Across Batch Executor Partitions

**What failed:** The `pre_schema` dict was passed by reference to `QueryResult.from_partition_slice()` across multiple partition slices. Downstream mutation of the schema dict in one partition contaminated other partitions' schema views.

**How fixed:** `schema=dict(pre_schema)` (shallow copy) at `unified_batch_executor.py:535`.

---

### HF-13 / HF-14 — ECS Deployment Stabilization Failures

**HF-13:** RF-C1 migrated `RedisSettings` from `REDIS_URL` to component fields (`REDIS_HOST`, etc.). ECS task definition still set `REDIS_URL`, causing `RedisSettings` to fail validation on ECS. Fixed by `model_validator` that parses `REDIS_URL` into components when `REDIS_HOST` is absent. Then **re-broken** by an intermediate commit (`3617cc8`) that removed the fallback again, requiring restore in `ee8fdb5`.

**HF-14:** `SlowAPIMiddleware` reads `app.state.limiter`, but code only set `app.state.infra.limiter`. This caused `AttributeError` on ALB health checks, preventing ECS deployment stabilization. Fixed in `07a7524` by setting both `app.state.limiter` and `app.state.infra.limiter`.

---

### PSEC-01 — Scheduling Overlap Check Uses String Comparison Instead of UTC

**What failed:** `_check_overlap_nonlocking()` compared appointment `start_datetime`/`end_datetime` as strings. Two datetimes representing the same UTC instant but expressed with different timezone offsets (e.g., `2024-01-01T12:00:00Z` vs `2024-01-01T07:00:00-05:00`) were not detected as overlapping.

**When fixed:** Commit `e9db28b` (promote PSEC-01 xfail to passing regression test). Test in `tests/scheduling/test_adversarial.py`.

**How fixed:** `_parse_appointment_datetime()` in `src/autom8_data/core/repositories/appointments.py:35` normalizes all datetimes to UTC before comparison.

---

### P0-1 — SQL Injection via table_prefix in Reconciliation Gauges

**What failed:** `refresh_reconciliation_gauges()` constructed SQL with unquoted `table_prefix`, enabling SQL injection if the prefix was attacker-controlled (sourced from config). Also a direct `500` risk if the prefix contained special characters.

**When fixed:** Commit `c44eff6` (P0-1 of 3 production readiness blockers).

**How fixed:** `table_prefix` identifier quoted via ANSI double-quote escaping in `insight_executor.py:31` `_quote_identifier()`.

---

### P0-2 — Cache Key Does Not Include Parquet Data Version

**What failed:** `ConnectionRouter` cache key did not include the Parquet data version. On materialization swap (new Parquet files written), the cache still served stale results from the previous materialization.

**When fixed:** Commit `c44eff6` (P0-2 of 3).

**How fixed:** `ConnectionRouter.get_current_version()` reads the `current` symlink target. Engine passes version at cache check and store sites.

---

### FSH — Filter Semantic Schism (Silent Wrong SQL)

**What failed:** Three overlapping scope-dimension translation mechanisms existed independently: one in the engine, one in the translator, one in the insights service. When scope dimensions (e.g., `vertical`, `business`) appeared in the `filters` dict, only some paths translated them to technical column names (`key`, `office_phone`). Other paths produced invalid SQL (e.g., `WHERE vertical = 'chiro'` instead of `WHERE key = 'chiro'`), silently returning wrong results or empty result sets.

**When fixed:** Commit `126ae2e` (harden filter pipeline — eliminate semantic schism).

**How fixed:** Single canonical `normalize_scope_filter_keys()` utility in `constants.py`. `merge_filters()` extended with `scope_dimension_map` param. Symmetric `required_filters` validation. Defense-in-depth normalization in `_build_query_filters()`. 68 new tests.

---

### WS4-C3 — Empty-String Phone Match in Materializer JOINs

**What failed:** `messages` and `calls` JOIN queries matched on Twilio phone (`sent_from`/`sent_to`). Empty-string phone values (`twil_phone = ''`) in the `chiropractors` table generated incorrect `business_phone` matches against any message with an empty sender/recipient phone.

**When fixed:** Commit `159bd80` (WS4-C3).

**How fixed:** `AND sent_from != ''` / `AND sent_to != ''` guards added to all 4 JOIN clauses in the messages and calls materialization queries (`materializer.py`).

---

### WS4-C5 — Corrupted Parquet File Served After Copy-Forward

**What failed:** `_copy_previous_version()` used `shutil.copy2()` without verifying the copy succeeded byte-for-byte. A corrupted copy would be served to readers without detection.

**When fixed:** Commit `b06ab36` (WS4-C5).

**How fixed:** SHA256 checksum comparison after `shutil.copy2()`. Corrupted copies are discarded rather than served.

---

### Contacts COUNT_DISTINCT (WS1-C6)

**What failed:** The `contacts` metric used `COUNT(id)` instead of `COUNT_DISTINCT(id)`. In rolling window queries, the same appointment can appear across multiple daily buckets, causing `COUNT` to overcount. Standard date-range queries were unaffected (dedup at table level eliminates duplicates).

**When fixed:** Commit `4e18b6f` (WS1-C6).

**How fixed:** Applied the three-declaration pattern (`requires_distinct`, `requires_raw_grain_for_rolling`, `raw_grain_column`) identical to `scheds` and `sms_conversation_count` in `src/autom8_data/analytics/core/metrics/library.py`.

---

## Category Coverage

| ID | Scar | Category |
|----|------|----------|
| SCAR-001 | Fact-to-fact Cartesian product | Data Corruption / Query Correctness |
| SCAR-002 | Non-deterministic set iteration | Race Condition / Non-determinism |
| SCAR-003 | Backtick quoting crashes DuckDB | Integration Failure (DuckDB dialect) |
| SCAR-008 | Window skip set logic | Query Correctness |
| SCAR-009 | Composite recomputation ordering | Query Correctness |
| SCAR-012 | `__future__ annotations` / Pydantic runtime | Integration Failure (Python runtime) |
| SCAR-013 | Window vertical join wrong path (deferred) | Query Correctness (latent) |
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
| HF-13 | REDIS_URL removed from ECS env | Config Drift |
| HF-14 | SlowAPI state key mismatch | Integration Failure (middleware) |
| PSEC-01 | Timezone string comparison overlap | Security / Correctness |
| P0-1 | SQL injection via table_prefix | Security |
| P0-2 | Cache stale after materialization swap | Performance Cliff / Stale Cache |
| FSH | Filter semantic schism | Integration Failure (silent wrong SQL) |
| WS4-C3 | Empty-string phone JOIN | Data Corruption (false matches) |
| WS4-C5 | Corrupted Parquet served without detection | Data Corruption |
| WS1-C6 | contacts COUNT overcounting | Data Corruption (inflated counts) |

**Categories present:**
1. Data Corruption — 8 scars
2. Query Correctness — 9 scars
3. Integration Failure — 5 scars (DuckDB dialect, Python runtime, middleware, endpoint 500s)
4. Config Drift — 1 scar
5. Race Condition / Non-determinism — 2 scars
6. Security — 2 scars
7. Performance Cliff / Stale Cache — 1 scar
8. Schema / Validation — 1 scar

9 distinct categories present.

---

## Fix-Location Mapping

| Scar | Fix File(s) | Function(s) |
|------|-------------|-------------|
| SCAR-001 | `src/autom8_data/analytics/core/query/builder.py:548`, `src/autom8_data/analytics/core/infra/exceptions.py:343` | `QueryBuilder._validate_fact_tables()`, `CartesianRiskError.__init__()` |
| SCAR-002 | `src/autom8_data/analytics/engine.py:1346`, `src/autom8_data/analytics/core/joins/optimizer.py:158,576`, `src/autom8_data/analytics/core/query/fact_resolver.py:468` | `_execute_and_aggregate()`, `_build_spanning_tree()`, `_sort_fact_tables()` |
| SCAR-003 | `src/autom8_data/analytics/core/infra/materializer.py:1793`, `src/autom8_data/analytics/insight_executor.py:31` | `_quote_identifier()` (two independent copies) |
| SCAR-008 | `src/autom8_data/analytics/window_aggregation.py:82` | `apply_windowed_aggregation()` |
| SCAR-009 | `src/autom8_data/analytics/window_aggregation.py:83` | `apply_windowed_aggregation()` / `aggregate_by_time_bucket()` |
| SCAR-012 | `src/autom8_data/api/routes/scheduling.py`, `src/autom8_data/api/scheduling/booking.py`, `src/autom8_data/api/scheduling/engine.py` (+ 8 more) | Module-level import removal |
| SCAR-013 (deferred) | `src/autom8_data/analytics/core/query/window_metric_sql.py:273` | `WindowMetricSQLGenerator._build_vertical_join()` |
| BUG-1 | `src/autom8_data/analytics/insights/processors/reconciliation.py` | `ReconciliationCoverageProcessor.process()` |
| BUG-4 | `src/autom8_data/analytics/core/infra/materializer.py:745` | `Materializer.DEFAULT_TABLE_FILTERS` |
| BUG-6 | `src/autom8_data/api/services/` (question_level_stats service) | activity filter logic |
| DEF-001 | `src/autom8_data/analytics/core/execution/composite.py:84`, `src/autom8_data/analytics/engine.py:2211`, `src/autom8_data/api/models.py:294`, `src/autom8_data/analytics/core/output/result.py:347` | `CompositeCalculator.calculate()`, `_compute_column_aliases()`, `query_result_to_response()`, `QueryResult` |
| DEF-1 | `src/autom8_data/analytics/window_aggregation.py:84` | `apply_windowed_aggregation()` |
| DEF-002 | `src/autom8_data/analytics/primitives/health/` (score engine) | `HealthScoreEngine.compute_score()` |
| DEF-003 | Same as DEF-002 | `HealthScoreEngine.compute_score()` |
| DEF-004 | `src/autom8_data/analytics/primitives/config/` (bucket config) | `BucketConfig` validator |
| DEF-S1-002/003 | `src/autom8_data/analytics/engine.py` | `_execute_and_aggregate()` |
| DEF-P2-002 | `src/autom8_data/analytics/services/unified_batch_executor.py:535` | `UnifiedBatchExecutor._build_partition_result()` |
| HF-13 | `src/autom8_data/core/config.py` | `RedisSettings._parse_redis_url_fallback()` |
| HF-14 | `src/autom8_data/api/main.py` | `create_app()` rate limiter setup |
| PSEC-01 | `src/autom8_data/core/repositories/appointments.py:35,128,156` | `_parse_appointment_datetime()`, `_check_overlap_nonlocking()` |
| P0-1 | `src/autom8_data/analytics/insight_executor.py:31` | `_quote_identifier()` |
| P0-2 | `src/autom8_data/analytics/core/infra/connection_router.py` | `ConnectionRouter.get_current_version()` |
| FSH | `src/autom8_data/analytics/core/infra/constants.py`, `src/autom8_data/analytics/engine.py` | `normalize_scope_filter_keys()`, `merge_filters()`, `_build_query_filters()` |
| WS4-C3 | `src/autom8_data/analytics/core/infra/materializer.py` | `MaterializationJob._build_messages_query()`, `_build_calls_query()` |
| WS4-C5 | `src/autom8_data/analytics/core/infra/materializer.py` | `MaterializationJob._copy_previous_version()` |
| WS1-C6 | `src/autom8_data/analytics/core/metrics/library.py` | `register_metrics()` contacts metric definition |

All file paths verified to exist in the repository.

---

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Location | Regression Test |
|------|-------------------|----------|-----------------|
| SCAR-001 | `CartesianRiskError` hard exception raised when fact-to-fact join detected in single-query path | `builder.py:548`, `exceptions.py:343` | `tests/analytics/test_multi_fact_rolling.py:454,493`, `tests/golden_master/test_analytics_golden.py:456`, `tests/analytics/test_count_distinct_rolling.py:236,344` |
| SCAR-002 | `sorted()` wrappers at all fact-table and set-iteration sites | `engine.py:1346`, `optimizer.py:158,576`, `fact_resolver.py:468,389` | Implicit — deterministic test assertions |
| SCAR-003 | `_quote_identifier()` ANSI double-quote escaping (two independent implementations) | `materializer.py:1793`, `insight_executor.py:31` | `tests/analytics/test_reconciliation_gauges.py` |
| SCAR-008 | Skip set built explicitly for raw grain metrics, display dims, enrichment cols | `window_aggregation.py:82` | `tests/analytics/test_window_aggregation.py` |
| SCAR-009 | Composite recomputation after rolling aggregation (inside `aggregate_by_time_bucket()`) | `window_aggregation.py:83` | `tests/analytics/test_window_aggregation.py` |
| SCAR-012 | No `__future__ annotations` in scheduling subsystem (enforced by code review) | All 11 scheduling files | None (no automated enforcement) |
| SCAR-013 | Deferred — comment at join site warns future developers | `window_metric_sql.py:279` | None (no production exercise path) |
| BUG-1 | Zeroed `data_quality` dict returned on empty coverage result | `reconciliation.py` | `tests/analytics/test_reconciliation.py:665` |
| BUG-4 | `DEFAULT_TABLE_FILTERS["appointments"] = "type = 'appt'"` in materializer | `materializer.py:745` | Implicit in materialization tests |
| BUG-6 | Activity filter applies to `frame_type` (not suppressing zero rows) | question_level_stats service | `tests/api/services/test_question_level_stats.py:421` |
| DEF-001 | Multi-pass composite calculator + column alias remapping at API layer | `composite.py:84`, `result.py:347`, `models.py:294`, `engine.py:2211` | `tests/analytics/test_execute_insight.py:702-824`, `tests/analytics/test_dimension_aliases.py:193`, `tests/api/routes/test_analytics_health.py:537` |
| DEF-1 | Always passes explicit `windowed_overrides` dict | `window_aggregation.py:84` | `tests/analytics/test_window_aggregation.py:294` |
| DEF-002 | NaN detected early, excluded via re-weighting | health score engine | `tests/analytics/primitives/health/test_score_engine_golden.py:1034` |
| DEF-003 | `max(0, normalized_score)` clamp | health score engine | `tests/analytics/primitives/health/test_score_engine_golden.py:1102` |
| DEF-004 | Pydantic validator enforces ascending threshold ordering | `BucketConfig` | `tests/analytics/primitives/config/test_health_scoring_config.py:183` |
| DEF-S1-002/003 | `dict.fromkeys()` deduplication of raw grain columns | `engine.py` | `tests/analytics/test_multi_fact_rolling.py:378,422` |
| DEF-P2-002 | `schema=dict(pre_schema)` shallow copy before partition slice | `unified_batch_executor.py:535` | None documented |
| HF-13 | `_parse_redis_url_fallback` model_validator on `RedisSettings` | `config.py` | `tests/config/test_config.py` (env hygiene regression tests) |
| HF-14 | Both `app.state.limiter` and `app.state.infra.limiter` set | `main.py` | None documented |
| PSEC-01 | `_parse_appointment_datetime()` UTC normalization before comparison | `appointments.py:35` | `tests/scheduling/test_adversarial.py` |
| P0-1 | `_quote_identifier()` in `insight_executor.py` | `insight_executor.py:31` | `tests/analytics/test_reconciliation_gauges.py` |
| P0-2 | `get_current_version()` included in cache key | `connection_router.py` | Implicit in cache tests |
| FSH | `normalize_scope_filter_keys()` + `merge_filters(scope_dimension_map)` + `_build_query_filters()` defense-in-depth | `constants.py`, `engine.py` | `tests/analytics/test_filter_merge.py` (26), `tests/analytics/test_fsh_adversarial.py` (42) |
| WS4-C3 | `AND sent_from != ''` guards in messages/calls JOIN clauses | `materializer.py` | Implicit in materializer tests |
| WS4-C5 | SHA256 checksum after `shutil.copy2()` | `materializer.py` | None documented |
| WS1-C6 | `requires_distinct=True`, `requires_raw_grain_for_rolling=True` on contacts metric | `metrics/library.py` | No rolling-window contacts test at fix time |

**Scars with no regression test:** SCAR-012 (no automated enforcement of `__future__` annotation absence), SCAR-013 (deferred, no production path), DEF-P2-002 (no test for schema mutation), HF-14 (no test for state.limiter attribute), WS4-C5 (no test for copy checksum), WS1-C6 (no rolling-window contacts test).

---

## Agent-Relevance Tagging

| Scar | Relevant Agents | Why |
|------|-----------------|-----|
| SCAR-001 | **query-strategist**, **analytics-qa** | Any multi-fact query modification risks re-enabling the Cartesian path. The `CartesianRiskError` guard is a hard constraint around the query compilation pipeline. |
| SCAR-002 | **query-strategist** | Any new dict/set iteration over fact tables must use `sorted()`. Non-deterministic ordering produces flaky tests and inconsistent production results. |
| SCAR-003 | **query-strategist**, **data-quality-sentinel** | Any new SQL construction using identifiers from config or metrics must use `_quote_identifier()` with ANSI double-quote style. Backtick quoting causes materializer circuit-breaker OPEN. |
| SCAR-008, SCAR-009 | **insight-engineer**, **analytics-qa** | New rolling window insights must ensure composite recomputation happens after base metric rolling, and skip sets are correctly built. |
| SCAR-012 | **insight-engineer** (scheduling), **principal-engineer** | Never add `from __future__ import annotations` to files with FastAPI route handlers or Pydantic models using runtime annotation evaluation. |
| SCAR-013 | **insight-engineer** | Any insight that combines window metrics with the `vertical` dimension will hit the wrong join path. Must be re-implemented using campaign hierarchy path before such an insight is added. |
| BUG-1 | **data-quality-sentinel**, **insight-engineer** | Post-processors must always return a structured (possibly zeroed) response, never `None`. |
| BUG-4 | **data-quality-sentinel**, **query-strategist** | Materializer `DEFAULT_TABLE_FILTERS` is the canonical place for table-level row filters. Do not rely on metric definitions to filter appointment types. |
| BUG-6 | **analytics-qa** | Activity filter behavior in question_level_stats is a regression-tested invariant. Zero-spend/zero-leads rows must survive the filter. |
| DEF-001 | **insight-engineer**, **analytics-qa** | Composite metrics depending on other composites require multi-pass calculation. API response must use alias remapping for canonical column names. |
| DEF-1 | **insight-engineer** | Always pass explicit `windowed_overrides` to `aggregate_by_time_bucket()`. Never rely on auto-discovery. |
| DEF-002, DEF-003, DEF-004 | **data-quality-sentinel** | Health score engine has three documented edge case fixes. Any health metric changes must re-run the golden master test suite. |
| DEF-S1-002/003 | **analytics-qa**, **query-strategist** | Multi-fact rolling queries require raw grain column deduplication. Do not pass duplicate columns to rolling aggregation. |
| DEF-P2-002 | **insight-engineer** | When passing `pre_schema` to `QueryResult.from_partition_slice()`, always shallow-copy to prevent cross-partition mutation. |
| HF-13 | **data-quality-sentinel** | `RedisSettings` has a backward-compat REDIS_URL fallback. Do not remove it without coordinating ECS task definition update. |
| HF-14 | **data-quality-sentinel** | Rate limiter must be set on both `app.state.limiter` (SlowAPI) and `app.state.infra.limiter` (internal). |
| PSEC-01 | **analytics-qa** (scheduling adversarial) | Appointment datetime comparison must always use `_parse_appointment_datetime()` for UTC normalization. Raw string comparison is wrong. |
| P0-1 | **query-strategist** | `table_prefix` in SQL must be quoted. Any new SQL construction using config-sourced identifiers must use `_quote_identifier()`. |
| P0-2 | **data-quality-sentinel** | Cache invalidation must include Parquet data version. Any new cache key construction must include `get_current_version()`. |
| FSH | **query-strategist**, **insight-engineer** | All filter dicts must go through `normalize_scope_filter_keys()` before use. Do not hardcode scope dimension names (`vertical`, `business`) as SQL column names. Use `SCOPE_DIMENSION_MAP`. |
| WS4-C3 | **data-quality-sentinel** | Any new materializer JOIN on phone columns must guard against empty-string values. |
| WS4-C5 | **data-quality-sentinel** | Any file copy in materializer should verify checksum post-copy. |
| WS1-C6 | **analytics-qa**, **metric-architect** | Metrics queried in rolling window contexts that can have duplicate rows in daily grain must use `COUNT_DISTINCT` with the three-declaration pattern. |

---

## Knowledge Gaps

1. **SCAR-004 through SCAR-007, SCAR-010, SCAR-011:** These SCAR numbers are referenced in the numbering sequence (SCAR-001 through SCAR-013 with gaps) but have no corresponding code markers or commits found in the search. These may have been assigned and never committed, exist in a different repo, or exist in squashed/cleaned history.

2. **BUG-2, BUG-3, BUG-5:** These BUG numbers are missing from source and test markers. BUG-1, BUG-4, BUG-6 are documented; intermediate numbers are unaccounted for.

3. **DEF-S1-004, DEF-S1-005:** Referenced in commit `ae03efb` body (`PercentageFormula bounds validation`, `MetricType enum count`) but no corresponding code markers found in source files.

4. **DEF-S2-001 (grouping_dimensions case normalization):** Referenced only in a test docstring (`test_request_validation.py:114`). No source code marker found; fix location not confirmed.

5. **HF-1 through HF-6:** `HF-3..6` and `HF-7..10` are batch-labeled commits. The underlying issues are lint/format violations and RF-C1 regressions, not distinct production failures. However HF-1 and HF-2 are absent from git log — may predate the visible history.

6. **SCAR-013 defensive test:** SCAR-013 is deferred with no test. If the vertical dimension is ever added to a window metric insight, the wrong-attribution bug will silently activate. This gap is intentional but unguarded.

7. **DEF-P2-002 regression test:** No test exercises the schema mutation scenario across batch executor partitions.
