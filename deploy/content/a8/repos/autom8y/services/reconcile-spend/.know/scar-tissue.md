---
domain: scar-tissue
generated_at: "2026-03-09T00:04:44Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "8e41207"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---
# Codebase Scar Tissue

## Failure Catalog

This service has a rich failure history documented across git commit messages, code comments, and a dedicated test file (`tests/test_defect_remediation.py`). The failure numbering uses two schemes: a `DEF-N` / `DEF-E0N` / `DEF-002` scheme for QA-assigned defects and a `BUG-N` scheme for upstream-sourced bugs.

---

### DEF-1 — Malformed Response Handling (Data Service + Asana Resolve)

**What failed**: When the data service or Asana resolve API returned non-200 status codes or structurally malformed JSON, the service either raised unhandled exceptions or silently returned incorrect results. Two subsystems were affected:
1. `DataInsightClient.get_insight()` did not translate HTTP errors into domain exceptions.
2. `batch_resolve_units()` in `asana_resolve.py` called `response.json()` and then `model_validate()` without any try/except, so a malformed `AsanaResolveResponse` raised a raw `pydantic.ValidationError` that propagated uncaught.

**Commit**: `c893811` (2026-02-22) — "feat(reconcile-spend): add typed service boundary contracts (WS-1/REC-3)"

**Marker**: `DEF-1` — referenced in `tests/test_defect_remediation.py` lines 1-4, 16, 89.

**Fix**: `ValidationError` now caught in `batch_resolve_units()` with structured logging and graceful empty-dict return. `DataInsightClient` raises `DataServiceUnavailableError` on non-200.

---

### DEF-4 — Non-Finite Guard Gap for `variance_pct`

**What failed**: The existing non-finite guard (which skipped rows with NaN/Inf `spend` or `collected`) did not extend to `variance` or `variance_pct`. A row with finite `spend` and `collected` but a NaN `variance_pct` would pass the guard and silently suppress anomaly detection — the anomaly score computation would produce NaN, which compares as false against all thresholds.

**Commit**: `c893811` (2026-02-22)

**Marker**: `DEF-4` — referenced in `tests/test_defect_remediation.py` lines 4, 142, 147; code comment in `orchestrator.py` line 75.

**Fix**: `parse_client_records()` in `orchestrator.py` (lines 76-90) now checks `math.isfinite()` on all four numeric fields: `spend`, `collected`, `variance`, and `variance_pct` (when not None).

---

### DEF-E01 — Pipe Character Not Escaped in Slack mrkdwn Link Labels

**What failed**: When a client's office name contained a pipe character (`|`), the Slack mrkdwn link syntax `<url|label>` broke because `|` is the Slack link delimiter. A name like `Smith | Jones Dental` would split the label field and corrupt the link rendering.

**Commit**: `5be1729` (2026-02-20) commit series and `19881f2` (2026-02-19)

**Marker**: `DEF-E01` — `tests/test_enrichment.py` line 665.

**Fix**: Link label construction in `report.py` escapes `|` as `&#124;` before embedding in `<url|label>` mrkdwn syntax.

---

### DEF-E02 — Angle Brackets Not Escaped in Slack mrkdwn Link Labels

**What failed**: Client office names containing `<` or `>` broke Slack mrkdwn link syntax since those characters delimit links in mrkdwn.

**Commit**: `5be1729` (2026-02-20)

**Marker**: `DEF-E02` — `tests/test_enrichment.py` line 688.

**Fix**: Link label construction in `report.py` escapes `<` as `&lt;` and `>` as `&gt;`.

---

### DEF-E03 — Non-HTTPS Invoice URLs Rendered as Slack Links

**What failed**: `hosted_invoice_url` values from the upstream API were passed through to Slack mrkdwn links without protocol validation. An `http://` (non-HTTPS) URL would be rendered as a clickable link, exposing users to potential plaintext-protocol content. Code comment in `orchestrator.py` line 247: `# Invoice URL (passthrough, HTTPS only per DEF-E03)`.

**Commit**: `5be1729` (2026-02-20)

**Marker**: `DEF-E03` — `tests/test_enrichment.py` lines 502, 526; `orchestrator.py` line 247.

**Fix**: `enrich_anomalies()` in `orchestrator.py` (lines 249-253) only passes `hosted_invoice_url` through when it `.startswith("https://")`.

---

### DEF-E05 — Coroutine Never Awaited in Orchestrator Integration Tests

**What failed**: Test setup for orchestrator integration tests left `_ensure_http` as an unbound coroutine on `AsyncMock` data client instances (7 test sites). This generated `RuntimeWarning: coroutine was never awaited` noise during test runs, indicating that the mock did not faithfully replicate the actual client interface.

**Commit**: `5be1729` (2026-02-20)

**Marker**: `DEF-E05` — commit message.

**Fix**: All 7 `AsyncMock` data client instances in orchestrator tests now have `_ensure_http` pinned to a `MagicMock`.

---

### DEF-002 — Variance Not Recalculated After E9 Negative-Collected Adjustment

**What failed**: Edge case E9 treats negative `collected` as zero. However, the original `variance` and `variance_pct` from the API (computed with the negative `collected` value) were preserved on the `ClientRecord` after the adjustment. This meant the record's internal numeric invariant (`variance == collected - spend`) was violated for E9 rows, producing incorrect variance display in the Slack report.

**Commit**: `5be1729` (2026-02-20)

**Marker**: `DEF-002` — `tests/test_orchestrator.py` line 708.

**Fix**: In `parse_client_records()` (`orchestrator.py` lines 93-101), after `collected` is floored to 0: `variance = collected - spend` (recomputed), `variance_pct = None` (invalidated).

---

### BUG-1 Companion — Misleading "100% Coverage" on Zero-Payment Periods

**What failed**: When the upstream `autom8y-data` API returned `data_quality.total_ad_spend_payments = 0` (no ad spend payments in the lookback period), `_format_data_quality()` in `report.py` would compute and display "100% coverage | 0 unreconciled ($0)" — technically accurate but operationally misleading. Operators reading the Slack report might conclude data collection was perfect when in fact there was nothing to reconcile.

**Commit**: `dfc1625` (2026-02-19) — "fix(reconcile-spend): handle zero-payment period in data quality display"

**Marker**: `BUG-1 companion` — `report.py` line 155 comment; `tests/test_report.py` line 418 class docstring.

**Fix**: `_format_data_quality()` (`report.py` lines 158-160) checks `total_ad_spend_payments == 0` first, and if so returns `"*Data Quality*: No ad spend payments in period"` before computing coverage.

---

### RS-M02 — Missing Threshold Keys in Failure-Path `near_miss` Dict

**What failed**: The failure-path `near_miss` dict hardcoded in `handler.py` (returned when an unhandled exception occurs before orchestrator completes) was created before REC-9 added asymmetric threshold fields (`overbilled_threshold_pct`, `underbilled_threshold_pct`). The `metrics.py` `.get()` defaults masked this at runtime, but failure-path metrics were silently reporting hardcoded defaults (10.0/25.0) instead of actual configured thresholds.

**Commit**: `ac3bd14` (2026-02-19) — "fix(reconcile-spend): add missing threshold keys to failure-path near_miss dict (RS-M02)"

**Marker**: `RS-M02` — commit message; `handler.py` failure-path dict at lines 128-136.

**Fix**: Failure-path dict in `handler.py` (lines 134-135) now reads threshold values from `get_settings()` dynamically.

---

### Asana Positional Parsing Bug (Unnamed)

**What failed**: The Asana resolve API returns result items positionally correlated with the input `criteria` array — it does NOT echo back `phone` or `vertical` fields in result items. The original parser in `batch_resolve_units()` searched for `.phone` and `.vertical` attributes on result items that did not exist, silently resolving 0 GIDs for every request. The Asana enrichment path appeared to work (no errors) but produced no Asana links in the Slack report.

**Commit**: `8cb8f8a` (2026-02-20) — "fix(reconcile-spend): parse Asana resolve response positionally, not by echoed fields"

**Marker**: No inline code marker; documented in commit message.

**Fix**: `batch_resolve_units()` (`asana_resolve.py` line 69) uses `zip(criteria, parsed.results, strict=False)` for positional correlation.

---

### CloudWatch Client Eager Init — CI Failure (Unnamed)

**What failed**: `boto3.client("cloudwatch")` was called at module import time in `metrics.py`. This required `AWS_DEFAULT_REGION` to be set in the environment. In CI (and any environment without AWS credentials), importing the `metrics` module caused an exception, blocking all tests that touched the metrics path.

**Commit**: `f76f33d` (2026-02-18) — "fix(reconcile-spend): defer CloudWatch client init for CI compatibility"

**Marker**: No inline code marker.

**Fix**: CloudWatch client is now lazy-initialized via `_get_cw_client()` pattern. The module-level binding is removed; the client is created on first use and cached at module scope.

---

### InsightMetadata Fields Optional — Silent Drop (Unnamed)

**What failed**: The local `InsightMetadata` consumer model declared only 2 of 9 fields present in the upstream `autom8y-data ResponseMetadata`. When the API returned the full payload, `model_dump()` silently dropped 7 fields: `query_id`, `duration_ms`, `cache_hit`, `cache_layer`, `tables_joined`, `contract_version`, `metric_versions`. Additionally, `staleness_seconds` was typed as `int` in the local model but `float` in the upstream schema.

**Commit**: `6310a2a` (2026-02-18) — "fix(reconcile-spend): align InsightMetadata with autom8y-data ResponseMetadata", `0ba200c` (2026-02-18) — "fix(reconcile-spend): make InsightMetadata fields optional"

**Marker**: No inline code marker.

**Fix**: All 9 fields now declared in the model (in `clients/models.py`); all marked `Optional` with `None` default; `staleness_seconds` typed as `float | None`.

---

### SCAR-001 — Docker `--link` on Non-Base-Image COPY

**What failed**: Using `COPY --link --from=<stage>` for non-base-image stages (secrets extension, uv) in Dockerfiles caused BuildKit to create independent overlay layers that shadowed `/bin/` from the base image, making `/bin/sh` inaccessible during the build. This manifested as cryptic `exec /bin/sh: no such file or directory` build failures.

**Commit**: `eb77ac4` (2026-02-26) — "fix(docker): remove --link from COPY --from=uv to fix BuildKit /bin/sh overlay"; codified as `SCAR-001` in `251a5f9` (2026-02-27) — "fix(templates): harden scaffold Dockerfiles against SCAR-001 and supply-chain risks"

**Marker**: `SCAR-001` — referenced in commit `251a5f9` message.

**Fix**: `--link` removed from all `COPY --from=<non-base-stage>` instructions in the `Dockerfile`.

---

### Test Environment Leak — `AUTOM8Y_ENV` from direnv Shell

**What failed**: The `AUTOM8Y_ENV` environment variable (set in the developer's direnv shell to `production` or `staging`) leaked into the test process. The production URL guard in `Autom8yBaseSettings` then fired on tests that set production-like `DATA_SERVICE_URL` defaults, causing false failures. Separately, tests were using the legacy `ENVIRONMENT` env var name instead of the canonical `AUTOM8Y_ENV`.

**Commit**: `6b821b1` (2026-02-26) — "fix(ci): resolve reconcile-spend test env leak and auth-mysql-sync mypy strict errors"

**Marker**: No inline code marker; documented in commit message.

**Fix**: An `autouse` fixture in `conftest.py` clears and resets `AUTOM8Y_ENV` for each test; test env vars migrated from `ENVIRONMENT` to `AUTOM8Y_ENV`.

---

### HAZ-1 — Production URL Reachable From Dev/Test Environment

**What failed**: No guard prevented a developer or test runner from accidentally configuring production service URLs (`*.autom8y.io`) while `AUTOM8Y_ENV` was set to `local`, `development`, or `test`. This created a risk of test data being written to production APIs.

**Commit**: Introduced as part of SDK base settings; test coverage added in `test_config_url_guard.py` (no single fix commit in reconcile-spend scope; guard lives upstream in `autom8y_config.Autom8yBaseSettings`).

**Marker**: `HAZ-1` — `tests/test_config_url_guard.py` module docstring line 5.

**Fix**: `Autom8yBaseSettings` raises `ValueError` with a `FATAL: Production URL detected` message when a URL containing `autom8y.io` is configured in a non-production environment.

---

### `AUTOM8Y_ENV` Not Read With Custom `env_prefix`

**What failed**: Child `Settings` classes with a custom `env_prefix` (e.g., `AUTH__`, `STRIPE_`) caused pydantic-settings to look for `{PREFIX}AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`, so the field defaulted to `LOCAL` in ECS task definitions where only `AUTOM8Y_ENV` was set. This triggered the production URL guard in production.

**Commit**: `1367461` (2026-03-08) — "fix(config): AUTOM8Y_ENV now read regardless of child class env_prefix"

**Marker**: No inline code marker; documented in commit and config comment at `config.py` lines 127-132.

**Fix**: `AliasChoices("AUTOM8Y_ENV", "ENVIRONMENT")` applied to the `autom8y_env` field in `Settings`, ensuring canonical env var takes precedence over prefix-qualified name.

---

### `data_quality` Omitted from Lambda Response Body

**What failed**: `ReconciliationResult.to_dict()` did not include `data_quality`, `staleness_seconds`, or `data_source` fields in the Lambda response body. The orchestrator had the data but it was silently dropped before the JSON response was built, making the Lambda response useless for downstream consumers that relied on financial quality signals.

**Commit**: `51ea849` (2026-02-19) — "fix(reconcile-spend): include data_quality in Lambda response body"

**Marker**: No inline code marker; documented in commit message.

**Fix**: Three fields wired from insight response through `ReconciliationResult` dataclass into `to_dict()` serialization.

---

## Category Coverage

| Category | Scars | Examples |
|---|---|---|
| **Data integrity / silent corruption** | DEF-4, DEF-002, Asana positional parsing, InsightMetadata silent field drop | NaN propagation, variance invariant violation after adjustment, zero GID resolution, silent model truncation |
| **External contract / API shape mismatch** | DEF-1, Asana positional parsing, InsightMetadata, BUG-1 companion | Non-200 unhandled, positional vs. echo response shape, optional field count mismatch, misleading quality display |
| **Output encoding / injection** | DEF-E01, DEF-E02, DEF-E03 | Slack mrkdwn pipe delimiter, angle bracket link corruption, non-HTTPS URL passthrough |
| **Infrastructure / build** | SCAR-001, CloudWatch eager init | Docker `--link` overlay, import-time credential requirement |
| **Configuration / environment** | HAZ-1, test env leak, `AUTOM8Y_ENV` prefix bug | Production URL guard, direnv shell contamination, env_prefix field resolution |
| **Missing output wiring** | RS-M02, `data_quality` omitted from response | Hardcoded defaults in failure-path, fields not serialized to Lambda response |
| **Test infrastructure** | DEF-E05 | Unawaited coroutine warning on AsyncMock |

Seven distinct categories are represented across 17 documented scars.

---

## Fix-Location Mapping

| Scar | Primary Fix Location | Secondary Fix Location |
|---|---|---|
| DEF-1 (data service) | `src/reconcile_spend/clients/data_service.py` — `DataInsightClient` non-200 handling | `src/reconcile_spend/clients/asana_resolve.py` — `batch_resolve_units()` ValidationError catch |
| DEF-4 | `src/reconcile_spend/orchestrator.py:76-90` — `parse_client_records()` non-finite guard extended | — |
| DEF-E01 | `src/reconcile_spend/report.py` — link label escaping | — |
| DEF-E02 | `src/reconcile_spend/report.py` — link label escaping | — |
| DEF-E03 | `src/reconcile_spend/orchestrator.py:249-253` — `enrich_anomalies()` HTTPS check | — |
| DEF-E05 | `tests/test_orchestrator.py` — 7 AsyncMock sites | — |
| DEF-002 | `src/reconcile_spend/orchestrator.py:93-101` — `parse_client_records()` E9 branch | — |
| BUG-1 companion | `src/reconcile_spend/report.py:155-160` — `_format_data_quality()` | — |
| RS-M02 | `src/reconcile_spend/handler.py:128-136` — failure-path `near_miss` dict | — |
| Asana positional parsing | `src/reconcile_spend/clients/asana_resolve.py:69` — `zip(criteria, parsed.results, strict=False)` | — |
| CloudWatch eager init | `src/reconcile_spend/metrics.py` — lazy `_get_cw_client()` | — |
| InsightMetadata truncation | `src/reconcile_spend/clients/models.py` — `InsightMetadata` model fields | — |
| SCAR-001 | `Dockerfile` — `COPY --from` stages | — |
| Test env leak | `tests/conftest.py` — `autouse` fixture | — |
| HAZ-1 | `autom8y_config.Autom8yBaseSettings` (upstream SDK, not in this service) | `tests/test_config_url_guard.py` (validation tests here) |
| `AUTOM8Y_ENV` prefix bug | `src/reconcile_spend/config.py:129-132` — `AliasChoices` on `autom8y_env` field | Upstream `autom8y-config` v1.2.1 |
| `data_quality` omitted | `src/reconcile_spend/models.py` — `ReconciliationResult.to_dict()` | — |

All fix file paths exist in the repository.

---

## Defensive Patterns

| Scar | Defensive Pattern | Pattern Location | Regression Test |
|---|---|---|---|
| DEF-1 | `try/except ValidationError` wraps all `model_validate()` at HTTP boundaries; returns `{}` or raises domain error | `asana_resolve.py:52-65`, `data_service.py` | `tests/test_defect_remediation.py` — `TestDef1DataServiceMalformedResponse`, `TestDef1AsanaResolveMalformedResponse` |
| DEF-4 | `math.isfinite()` guard on all four numeric fields before appending to `records` | `orchestrator.py:76-90` | `tests/test_defect_remediation.py` — `TestDef4NonFiniteVariancePct` (6 parametric cases) |
| DEF-E01/E02 | HTML entity escaping on link labels before mrkdwn construction | `report.py` | `tests/test_enrichment.py:664-707` |
| DEF-E03 | `startswith("https://")` guard on all URL passthroughs | `orchestrator.py:249-253` | `tests/test_enrichment.py:501-543` |
| DEF-002 | Variance/variance_pct recomputed immediately after E9 adjustment | `orchestrator.py:93-101` | `tests/test_orchestrator.py:708-727` |
| BUG-1 companion | Zero-check on `total_ad_spend_payments` before coverage calculation | `report.py:158-160` | `tests/test_report.py:417-444` (`TestFormatDataQualityZeroPayments`) |
| RS-M02 | Failure-path dict reads thresholds from `get_settings()` dynamically | `handler.py:134-135` | `tests/test_handler.py` — `test_failure_metric_includes_threshold_keys` |
| Asana positional parsing | `zip(criteria, parsed.results, strict=False)` for positional correlation | `asana_resolve.py:69` | `tests/test_defect_remediation.py` (via contract regression tests) |
| CloudWatch eager init | Lazy `_get_cw_client()` with module-level cache var | `metrics.py` | `tests/test_metrics.py` |
| InsightMetadata | All 9 fields declared optional; `float | None` for `staleness_seconds` | `clients/models.py` — `InsightMetadata` | `tests/test_contract_regression.py` |
| SCAR-001 | `--link` removed from all `COPY --from=<non-base-stage>` | `Dockerfile` | No unit test (build-time only) |
| Test env leak | `autouse` fixture clears `AUTOM8Y_ENV` before/after each test | `tests/conftest.py` | Self-documenting (the fixture is the guard) |
| HAZ-1 | Validator in `Autom8yBaseSettings` raises `ValueError` on production URL in dev env | Upstream SDK | `tests/test_config_url_guard.py` — `TestProductionUrlGuard` (5 cases) |
| `AUTOM8Y_ENV` prefix | `AliasChoices("AUTOM8Y_ENV", "ENVIRONMENT")` on `autom8y_env` field | `config.py:129-132` | `tests/test_config.py` |
| `data_quality` omitted | All three fields added to `ReconciliationResult.to_dict()` | `models.py` | `tests/test_handler.py` |
| DEF-E05 | `_ensure_http` pinned to `MagicMock` on all AsyncMock clients | `tests/test_orchestrator.py` | Self-guarding via test warnings-as-errors |
| Metric emission failure isolation | Entire `emit_reconciliation_metrics()` body wrapped in try/except; never raises | `metrics.py:49` | `tests/test_metrics.py` |

One scar (SCAR-001) has no regression test because it is a build-time failure only verifiable by Docker build execution, not Python unit tests.

---

## Agent-Relevance Tags

| Scar | Relevant Agent | Why |
|---|---|---|
| DEF-1 (HTTP boundary) | **principal-engineer** | Any new cross-service call must follow the `try/except ValidationError -> domain error` pattern established here |
| DEF-4 (non-finite guard) | **principal-engineer** | Any new numeric field from the insight API requires explicit `math.isfinite()` guard in `parse_client_records()` |
| DEF-E01/E02/E03 (Slack encoding) | **principal-engineer** | All text inserted into Slack mrkdwn links must be HTML-entity-escaped; all URL passthroughs must be HTTPS-only |
| DEF-002 (E9 variance invariant) | **principal-engineer** | After any numeric field adjustment in `parse_client_records()`, derived fields must be recomputed |
| BUG-1 companion | **principal-engineer** | `_format_data_quality()` has a zero-payment special case; do not collapse it into coverage calculation |
| RS-M02 (failure-path dict) | **principal-engineer** | Whenever new fields are added to `near_miss_data`, both the success path and the failure-path hardcoded dict in `handler.py` must be updated |
| Asana positional parsing | **principal-engineer** | Asana resolve API is positional, not echo-field. Never add field lookups by name on result items |
| SCAR-001 | **principal-engineer**, **architect** | `COPY --link` must never be used for non-base-image COPY stages in Dockerfiles |
| CloudWatch eager init | **principal-engineer** | AWS SDK clients requiring region/credentials must always be lazy-initialized; never at module import scope |
| InsightMetadata | **principal-engineer**, **requirements-analyst** | Any addition of fields to the upstream API response requires a corresponding model update here; all fields should be `Optional` |
| HAZ-1 | **architect**, **principal-engineer** | The production URL guard is active; all new service URL fields in `Settings` will be checked by the base class guard |
| `AUTOM8Y_ENV` prefix bug | **architect** | Settings classes with custom `env_prefix` must include `AliasChoices` on any canonical global env vars |
| Test env leak | **qa-adversary** | `AUTOM8Y_ENV` direnv contamination was a real failure mode; new test files that instantiate `Settings` must use the `_mock_env` autouse fixture pattern |
| `data_quality` omitted | **principal-engineer** | Any new field added to `ReconciliationResult` must also be wired into `to_dict()` |
| DEF-E05 (unawaited coroutine) | **qa-adversary** | `AsyncMock` instances for `DataInsightClient` require `_ensure_http` pinned to `MagicMock` to prevent warning noise |
| Metric emission failure isolation | **principal-engineer** | Metric emission must remain best-effort; never add code that could raise inside `emit_reconciliation_metrics()` without the internal try/except pattern |

---

## Knowledge Gaps

1. **DEF-E04** is not evidenced in source or tests despite `DEF-E01`, `DEF-E02`, `DEF-E03`, and `DEF-E05` existing. Either DEF-E04 was fixed in the upstream SDK (not in this service), or it was assigned but not yet addressed, or the numbering was non-sequential. No test or code comment references it.

2. **The `data_quality` omission fix** (`51ea849`) wired fields into the response body, but no specific test name was directly observed in `test_handler.py` confirming the fix is regression-guarded. The handler test file was not fully read.

3. **`WS-4` and `WS-7`** are referenced in commit `5be1729` as work-stream items but do not appear as inline code markers in any source file. Their precise fix locations within `active_section_days` parsing are not marked.

4. **The isort config alignment** (`624d314`) is a tooling fix with no defensive code pattern — it cannot regress at the Python level. Not a traditional scar.

5. **The `AUTOM8Y_ENV` prefix bug** (`1367461`) is primarily fixed in the upstream `autom8y-config` v1.2.1 SDK and in the base class. The `AliasChoices` in this service's `config.py` (lines 129-132) may be a belt-and-suspenders addition for clarity, not the authoritative fix location.
