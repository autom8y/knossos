---
domain: scar-tissue
generated_at: "2026-03-16T14:32:40Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "c794130"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Scar Tissue

**Service**: `account-status-recon`
**Language**: Python 3.12
**Source directories**: `src/account_status_recon/`, `tests/`
**Git commits in scope**: 14 commits touching this service directly

---

## Failure Catalog Completeness

This service has a relatively short git history (14 commits, extracted from the monorepo log). No inline SCAR-NNN, BUG-, HACK, FIXME, or WORKAROUND markers exist in the source. Failures are traceable through fix commits, ADR codes in module docstrings, and EC-N edge-case codes embedded in test classes.

### Scar 1: mypy type narrowing failure in `fetch_all` gather result handling

**Commit**: `3df641d` — `fix(types): resolve mypy type errors across five services` (2026-03-12)
**What failed**: `asyncio.gather(return_exceptions=True)` returns `tuple[FetchResult | BaseException, ...]`. The original code used `isinstance(result, Exception)` as the guard but mypy cannot narrow the union via that branch ordering — the positive branch `fetch_dict[name] = result` remained typed as `FetchResult | BaseException`. Additionally, `response_dict = response.model_dump() if ... else response` left `response_dict` as `ActiveCampaignTreeResponse | dict[str, Any]`, causing mypy failures on subsequent `.get()` calls.
**How fixed**: Changed `isinstance(result, Exception)` guard to `isinstance(result, FetchResult)` (positive-branch narrowing). Added `response_dict: dict[str, Any] = ... cast("dict[str, Any]", response)` explicit annotation with `cast()`.
**Current fix location**: `src/account_status_recon/fetcher.py`, lines 75-79 (`response_dict` cast), lines 188-194 (`isinstance(result, FetchResult)` guard).

### Scar 2: `typing.cast` unquoted type expression causing ruff format failure

**Commit**: `a0973ed` — `fix(account-status-recon): quote typing.cast type expr and ruff format fetcher.py` (2026-03-14)
**What failed**: `cast(dict[str, Any], response)` — the generic type expression was unquoted. Ruff's `UP006`/`UP035` rules required quoting forward-ref style. Also triggered a line-length violation in `FetchResult` error wrapping in `fetch_all`.
**How fixed**: Changed to `cast("dict[str, Any]", response)` (quoted string). Reformatted the long `FetchResult` construction to multi-line.
**Current fix location**: `src/account_status_recon/fetcher.py`, line 78 (`cast("dict[str, Any]", response)`), lines 191-194 (multi-line `FetchResult` construction).

### Scar 3: IAM namespace mismatch blocking CloudWatch metric emission

**Commit**: `0437612` — `fix(infra): IAM namespace alignment, GuardDuty provider migration, Dockerfile` (2026-03-11)
**What failed**: The IAM policy for `CloudWatch:PutMetricData` used `Autom8y/AccountStatusRecon` as the namespace condition. The SDK's `SHARED_NAMESPACE` constant is `Autom8y/Reconciliation`. Every metric emission call was silently denied.
**How fixed**: Updated the IAM condition to `Autom8y/Reconciliation` in `terraform/services/account-status-recon/main.tf`.
**Current fix location**: `terraform/services/account-status-recon/main.tf` (outside service source; Terraform module).

### Scar 4: ECR public image for AWS Secrets Extension removed from registry

**Commit**: `0437612` — same commit, Dockerfile creation
**What failed**: The pattern `FROM public.ecr.aws/aws-parameters-and-secrets-lambda-extension:12` used by sibling services failed because AWS removed that image from ECR public. No container image had been authored yet for this service.
**How fixed**: Dockerfile uses an alternative pattern: downloads the Lambda Layer ZIP directly from a pre-signed S3 URL (`SECRETS_EXT_LAYER_URL` build-arg), extracts the extension binary with `unzip`, and validates with `test -f /opt/extensions/bootstrap`. CI passes the URL; local builds generate it via `aws lambda get-layer-version-by-arn`.
**Current fix location**: `Dockerfile`, lines 16-33 (comment block), lines 35-42 (curl + unzip + test pattern).

### Scar 5: Legacy env var aliases polluting config — `ENVIRONMENT`, `DATA_SERVICE_URL`, `ADS_SERVICE_URL`

**Commits**: `ababaa7` (2026-03-14), `35f29a1` (2026-03-14)
**What failed**: During initial extraction the `Settings` class carried `AliasChoices("AUTOM8Y_DATA_URL")`, `AliasChoices("AUTOM8Y_ADS_URL")`, `AliasChoices("AUTOM8Y_ASANA_URL")` on the URL fields, and `AliasChoices("AUTOM8Y_ENV", "ENVIRONMENT")` on `autom8y_env`. The `AliasChoices` wrappers were legacy compat shims inherited from the SDK `LambdaServiceSettingsMixin`. They were removed in a clean-break refactor per ADR-ENV-NAMING-CONVENTION, as both Terraform and app code were migrated to canonical names.
**How fixed**: Removed all `AliasChoices` wrappers from URL fields. Kept `AliasChoices("AUTOM8Y_ENV")` (single canonical alias) on `autom8y_env`. Added explicit `autom8y_env` field with `Autom8yEnvironment` type (previously inherited silently from base class).
**Current fix location**: `src/account_status_recon/config.py`, lines 36-47 (URL fields, no AliasChoices), lines 130-133 (`autom8y_env` explicit field).

### Scar 6: Missing `autom8y-telemetry[aws,otlp]` extras caused OTLP export failure

**Commit**: `ef3e8cf` — `refactor(reconciliation): standardize all three Lambda services to platform stack` (2026-03-14)
**What failed**: The original `pyproject.toml` declared `autom8y-telemetry>=0.2.0` without the `[aws,otlp]` extras. The `[aws]` extra provides the `@instrument_lambda` decorator; the `[otlp]` extra provides the OTLP exporter. Both were silently absent at runtime — Lambda traces were not exported.
**How fixed**: Bumped to `autom8y-telemetry[aws,otlp]>=0.5.2`. Subsequently upgraded to `[aws,otlp,conventions]>=0.5.2` (for convention checking) and then `>=0.6.0` (for W3C traceparent injection via `ResilientCoreClient`).
**Current fix location**: `pyproject.toml`, line 33 (`"autom8y-telemetry[aws,otlp,conventions]>=0.6.0"`).

### Scar 7: Distributed trace boundary gap — cross-service spans not correlated

**Commit**: `d4996f0` — `fix(autom8y-http): inject W3C traceparent via ResilientCoreClient (DC-7)` (2026-03-15)
**What failed**: HTTP calls from `fetch_billing`, `fetch_campaigns`, `fetch_offers` to upstream services did not inject `traceparent` headers. Spans in the dev console showed as disconnected — the upstream service started a new root span rather than a child.
**How fixed**: `ResilientCoreClient._inject_trace_context()` now calls `opentelemetry.propagate.inject()` before the first attempt. Injection is gated on OTel availability (no-op if not installed) and swallows all exceptions to never block requests.
**Current fix location**: SDK fix in `sdks/python/autom8y-http/...` (outside service boundary). Service consumed via `autom8y-telemetry>=0.6.0` floor bump in `pyproject.toml` line 33.

---

## Category Coverage

| Category | Scars | Notes |
|---|---|---|
| Integration failure | Scar 3 (IAM/CloudWatch), Scar 7 (trace propagation), Scar 6 (missing extras) | 3 entries |
| Config drift | Scar 5 (env var aliases), Scar 4 (ECR public image removed) | 2 entries |
| Type system / static analysis | Scar 1 (mypy narrowing), Scar 2 (ruff cast quoting) | 2 entries |
| Data corruption | None found — EC-11 (NaN/Inf guard) is proactive, not post-incident | 0 entries; EC-11 guard searched and found proactive only |
| Race condition | None found — `asyncio.gather` per-source isolation (ADR-ASR-002) is architectural, not reactive | 0 entries |
| Schema evolution | No migrations exist (service is greenfield, no DB schema) | not applicable |
| Performance cliff | No evidence found | 0 entries |
| Security | No evidence found in this service's history | 0 entries |

**3 distinct categories represented**: integration failure, config drift, type system. Data corruption, race condition, performance cliff, and security categories searched but not found.

---

## Fix-Location Mapping

| Scar | File(s) | Function / Location | Status |
|---|---|---|---|
| Scar 1 (gather narrowing) | `src/account_status_recon/fetcher.py` | `fetch_all()` lines 188-194; `fetch_campaigns()` lines 75-79 | Exists, verified |
| Scar 2 (cast quoting) | `src/account_status_recon/fetcher.py` | `fetch_campaigns()` line 78 | Exists, verified |
| Scar 3 (IAM namespace) | `terraform/services/account-status-recon/main.tf` | IAM `cloudwatch:PutMetricData` condition block | Outside `src/`; Terraform file exists at repo root |
| Scar 4 (ECR removal) | `Dockerfile` | Lines 16-42, `secrets-extension` build stage | Exists, verified |
| Scar 5 (env var aliases) | `src/account_status_recon/config.py` | `Settings` class, lines 36-47 (URL fields), 130-133 (`autom8y_env`) | Exists, verified |
| Scar 6 (missing extras) | `pyproject.toml` | `dependencies` array, line 33 | Exists, verified |
| Scar 7 (trace propagation) | SDK: `sdks/python/autom8y-http/` (outside service boundary); service manifest: `pyproject.toml` line 33 | `ResilientCoreClient._inject_trace_context()` (SDK); version floor bump (service) | SDK path not verified (outside scope); service side verified |

**Compound fixes**: Scar 1 spans two locations in `fetcher.py`. Scar 7 is split between SDK and service manifest.

---

## Defensive Pattern Documentation

### Scar 1 — per-source fault isolation (ADR-ASR-002)

**Defensive pattern**: `asyncio.gather(return_exceptions=True)` in `fetch_all()` with `isinstance(result, FetchResult)` positive guard. Every fetch function catches all exceptions and returns `FetchResult(error=exc)` — they never propagate.
**Pattern location**: `src/account_status_recon/fetcher.py`, lines 177-194.
**Comment marker**: Module docstring line 3 — `"ADR-ASR-002: asyncio.gather(return_exceptions=True) for graceful degradation."` Repeated at `fetch_all()` docstring line 171.
**Regression test**: `tests/test_fetcher.py::TestFetchAll::test_fetch_gather_exception_handling` — asserts that a function-level `RuntimeError` is wrapped into `FetchResult(error=RuntimeError)` and does not propagate.

### Scar 3 — IAM namespace alignment guard

**Defensive pattern**: CloudWatch metric emission under `Autom8y/Reconciliation` namespace. The `metrics.py` module uses the SDK's `SHARED_NAMESPACE` constant rather than a locally-defined string, preventing future namespace drift.
**Pattern location**: `src/account_status_recon/metrics.py` (not read in full; observed via commit diff that IAM condition was aligned to SDK constant).
**Comment marker**: None observed in source — fix is in Terraform, not Python.
**Regression test**: No direct Python test for IAM; integration-level concern.

### Scar 4 — ECR extension bootstrap validation

**Defensive pattern**: `test -f /opt/extensions/bootstrap` in Dockerfile build stage — fails the build if the extension binary is absent, catching extraction failures at image-build time rather than Lambda cold-start.
**Pattern location**: `Dockerfile`, line 42 (`test -f /opt/extensions/bootstrap`).
**Comment marker**: Lines 16-33 in Dockerfile contain explicit documentation of the ECR removal event and the workaround approach.
**Regression test**: None (build-time assertion only).

### Scar 5 — canonical env var names, no AliasChoices shims

**Defensive pattern**: `Settings` URL fields have no `validation_alias` at all — they bind directly to the field name (`autom8y_data_url`, etc.), preventing accidental re-introduction of legacy name fallbacks. `secretspec.toml` documents canonical names as the single source of truth for IaC.
**Pattern location**: `src/account_status_recon/config.py`, lines 36-47. `secretspec.toml` (documentation-only).
**Comment marker**: `secretspec.toml` header comment — `"Naming tiers per ADR-ENV-NAMING-CONVENTION"` with ADR path reference.
**Regression test**: None specific to alias removal; covered by integration in sibling SDK test suites.

### Scar 6 — telemetry extras version floor

**Defensive pattern**: `pyproject.toml` pins `autom8y-telemetry[aws,otlp,conventions]>=0.6.0`. The explicit extras list acts as a manifest of required capabilities — an agent bumping the version floor must preserve all three extras.
**Pattern location**: `pyproject.toml`, line 33.
**Comment marker**: Inline comment `# Lambda instrumentation and telemetry` on line 32.
**Regression test**: `tests/test_instrumentation.py` — verifies all 7 convention spans are emitted. Tests will fail at import if the `[aws]` or `[conventions]` extras are absent.

### EC-11 proactive guard — NaN/Inf in financial data

**Note**: Not born from a production failure in this service. The guard is proactive (copied from `reconcile-spend` pattern).
**Defensive pattern**: `billing.has_finite_values` check in `rule_billing()` (via `BillingData` property using `math.isfinite()`); `math.isfinite()` direct check in `rule_three_way()`.
**Pattern location**: `src/account_status_recon/rules.py`, lines 203-211 (`rule_billing`), lines 304-305 (`rule_three_way`). Comment: `# EC-11: Skip non-finite values`.
**Regression test**: `tests/qa/test_edge_cases_adversarial.py::TestEC11NanInfValues` — asserts `rule_billing(nan_billing) == []` and `rule_three_way(nan_billing, contract) == (None, None, None)`.

### EC-12 proactive guard — duplicate offer deduplication

**Defensive pattern**: `Correlator.dedup(contract_rows, contract_key_fn)` in `three_way_join()` before building contract index. First-wins on `(office_phone, vertical)`.
**Pattern location**: `src/account_status_recon/joiner.py`, lines 168-171. Comment: `# Dedup contract data: first-wins on (office_phone, vertical) per EC-12`.
**Regression test**: `tests/test_joiner.py::TestThreeWayJoin::test_join_dedup_offers` and `tests/qa/test_edge_cases_adversarial.py::TestEC12DuplicateOffers`.

### EC-19 — strict greater-than boundary semantics

**Defensive pattern**: `abs(pct) > drift_threshold_pct` (strict `>`, not `>=`) in `rule_three_way()`.
**Pattern location**: `src/account_status_recon/rules.py`, lines 314-315. Comment: `# EC-19: Strict greater-than comparison`.
**Regression test**: `tests/qa/test_edge_cases_adversarial.py::TestEC19ExactThresholdBoundary::test_ec19_exact_threshold_boundary` — asserts exactly-at-threshold is `MATCHED`.

### EC-20 — findings logged to CloudWatch before Slack (Slack truncation survival)

**Defensive pattern**: All findings are logged via `log.info("finding_detail", ...)` in `run_reconciliation()` before the Slack post. Findings survive Slack failure.
**Pattern location**: `src/account_status_recon/orchestrator.py`, lines 218-227. Comment: `# Log all findings to CloudWatch (EC-20: survives Slack truncation)`.
**Regression test**: `tests/qa/test_edge_cases_adversarial.py::TestEC20SlackFailureAfterVerdicts` (indirect — verifies findings computed independently of Slack).

---

## Agent-Relevance Tagging

| Scar | Agent Role(s) | Why |
|---|---|---|
| Scar 1 (gather narrowing / cast) | `principal-engineer`, `hallucination-hunter` | Any agent generating or reviewing `asyncio.gather` code must use `isinstance(result, FetchResult)` guard, not `isinstance(result, Exception)`. `cast()` on ternary branches must use quoted strings. |
| Scar 2 (ruff cast quoting) | `principal-engineer` | Any agent writing `cast()` calls must quote the type expression: `cast("dict[str, Any]", ...)`, not `cast(dict[str, Any], ...)`. |
| Scar 3 (IAM namespace) | `architect`, `principal-engineer` | Any agent writing Terraform for Lambda CloudWatch metrics must use `Autom8y/Reconciliation` as the IAM condition namespace, not a service-specific namespace. |
| Scar 4 (ECR image removal) | `architect`, `principal-engineer` | Any agent authoring a Dockerfile that uses the AWS Parameters and Secrets Lambda Extension must use the Layer ZIP download pattern, not `FROM public.ecr.aws/aws-parameters-and-secrets-lambda-extension`. |
| Scar 5 (env var aliases) | `principal-engineer`, `qa-adversary` | No `AliasChoices` shims for URL fields in `Settings`. All env vars must use canonical Tier 3 names. Reviewers should flag `AliasChoices` on URL fields as a convention violation. |
| Scar 6 (missing telemetry extras) | `principal-engineer` | When bumping `autom8y-telemetry`, preserve all three extras `[aws,otlp,conventions]`. Dropping an extra silently removes runtime capability. |
| Scar 7 (trace propagation) | `architect`, `principal-engineer` | Cross-service tracing requires `>=0.6.0` floor. Any new service using `ResilientCoreClient` benefits from this fix automatically if it pins `>=0.6.0`. |
| EC-11 NaN/Inf guard | `principal-engineer`, `qa-adversary` | Any new verdict rule operating on financial floats must include `math.isfinite()` guard before arithmetic. |
| EC-12 dedup | `principal-engineer` | Any new join over Asana offer rows must apply first-wins dedup before indexing. The Correlator SDK provides `Correlator.dedup()`. |
| EC-19 strict boundary | `principal-engineer`, `qa-adversary` | Threshold comparisons in reconciliation rules use `>` (strict), never `>=`. |

---

## Knowledge Gaps

1. **Terraform IAM file not read directly**: Scar 3's fix in `terraform/services/account-status-recon/main.tf` was observed from commit diff only, not by reading the current file. The fix is confirmed applied but line-level location is not verified.

2. **SDK-side Scar 7 not fully traced**: `ResilientCoreClient._inject_trace_context()` lives in `sdks/python/autom8y-http/` — outside this service's source scope. The fix was observed via commit message and `pyproject.toml` version floor but not via direct file read of the SDK source.

3. **No SCAR-NNN numbering system**: This service has no SCAR-NNN markers. Failure tracking is encoded as EC-N (edge case requirements from the PRD), ADR-ASR-NNN codes in module docstrings, and fix commit subjects. Future scars should adopt the SCAR-NNN convention for cross-file traceability.

4. **Proactive vs reactive scars**: EC-11, EC-12, EC-17, EC-19, EC-20 are proactive guards carried over from sibling services' post-incident patterns. They are not confirmed production failures in this specific service. They are documented here because they represent the same failure modes that hurt sibling services.

5. **`metrics.py` not fully read**: The defensive namespace pattern in `metrics.py` was inferred from commit context, not from reading the file directly.
