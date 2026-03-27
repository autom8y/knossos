---
domain: scar-tissue
generated_at: "2026-03-16T20:00:00Z"
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

# Codebase Scar Tissue

## Failure Catalog

### SCAR-SA-001: Lambda error metric blind spot (CHAOS-001 P0)

**What failed**: Slack delivery failures (invalid_auth, channel_not_found, network errors) were silently swallowed. The Lambda Errors CloudWatch metric never fired on Slack-side failures because the handler returned HTTP 200 even when all records errored. The 99.9% SLO burn-rate alert had no signal.

**When**: Discovered during CHAOS-001 chaos engineering experiment. Fixed in commit `0e9840c` (2026-03-04).

**How fixed**: After all per-record attempts complete, the handler now raises `RuntimeError` if `result["errors"] > 0`. This surfaces delivery failures to Lambda Errors metrics. Per-record isolation is preserved — all records are attempted before the invocation fails.

**Current state**: Lines 84-85 of `src/slack_alert/handler.py`:
```python
if result["errors"] > 0:
    raise RuntimeError(f"Failed to deliver {result['errors']} of {result['processed']} alerts")
```

---

### SCAR-SA-002: Hand-rolled secret bridge antipattern

**What failed**: The handler contained a `_resolve_secret()` function that manually called the AWS Parameters and Secrets Lambda Extension HTTP endpoint (`localhost:2773`). This was brittle, service-specific, and duplicated across multiple services. The function was also flagged by mypy (`no-any-return`) because `json.loads` returns `Any`.

**When**: Initial implementation. Mypy error surfaced in commit `f9c07c1` (2026-02-17). Root antipattern removed in commit `15c26bf` (2026-03-03).

**How fixed**: Removed `_resolve_secret()` entirely. `SlackConfig()` now self-resolves via `Autom8yBaseSettings` ARN auto-resolution — `SLACK_BOT_TOKEN_ARN` env var is resolved transparently by the SDK settings infrastructure.

**Residual marker**: `# type: ignore[call-arg]` on line 31 of `src/slack_alert/handler.py` — mypy cannot see the model_validator that injects `bot_token` at runtime (commit `717322a`).

---

### SCAR-SA-003: Missing CI dev dependencies

**What failed**: `pyproject.toml` was missing `ruff`, `mypy`, `pytest`, and `pytest-cov` in dev dependencies. `service-ci.yml` runs `uv run ruff` which requires ruff to be resolvable from the project. The CI pipeline would fail at the lint step immediately.

**When**: CI pipeline bootstrap. Fixed in commit `590f6d8` (2026-02-17).

**How fixed**: Added `[dependency-groups] dev = [pytest, pytest-cov, mypy, ruff]` plus `[tool.pytest.ini_options]` and `[tool.mypy]` config sections to match the canonical service pattern. Also added `respx` dev dependency later (commit `291d80b`) when transitive import of `autom8y_http.testing` during pytest collection required it.

---

### SCAR-SA-004: Dockerfile build arg name mismatch

**What failed**: The Dockerfile used `SECRETS_LAYER_URL` as the build arg name, but `service-build.yml` and all other Lambda service Dockerfiles used `SECRETS_EXT_LAYER_URL`. The image could not be built via CI.

**When**: Fixed in commit `0ae009f` (2026-02-17).

**How fixed**: Renamed `SECRETS_LAYER_URL` -> `SECRETS_EXT_LAYER_URL` in the Dockerfile. Three occurrences updated.

---

### SCAR-SA-005: COPY --link breaking /bin/sh in BuildKit

**What failed**: `COPY --link --from=uv /uv /uvx /bin/` creates an independent layer that shadows `/bin/` from the base image, making `/bin/sh` inaccessible. BuildKit /bin/sh overlay failure. Manifested first in auth-mysql-sync (commit `5f52b71`) then propagated as a pattern fix to all services.

**When**: Fixed across all services in commit `eb77ac4` (2026-02-26).

**How fixed**: Removed `--link` flag specifically from the `COPY --from=uv` line. `--link` is still used for other `COPY` operations (source files, secrets extension) where it is safe. Current Dockerfile line 41: `COPY --from=uv /uv /uvx /bin/` (no `--link`).

---

### SCAR-SA-006: Missing CodeArtifact index URL in pip install step

**What failed**: `uv pip install --require-hashes` could not resolve private `autom8y-*` packages without the CodeArtifact index URL. The build step would fail to find internal SDK packages. Fix pattern originated in auth-mysql-sync (commit `13ce631`) and was applied to all services.

**When**: Fixed in commit `48e37d4` (2026-02-26).

**How fixed**: Added conditional `EXTRA_INDEX_URL` build arg to the pip install RUN steps. When set (CI path), passes `--index-url "$EXTRA_INDEX_URL" --extra-index-url https://pypi.org/simple/`. Local builds use `UV_INDEX_AUTOM8Y_PASSWORD_COMMAND` fallback.

---

### SCAR-SA-007: Silent OTLP export failure (spans created, never shipped)

**What failed**: `opentelemetry-exporter-otlp-proto-http` was missing from service dependencies. `_create_exporter()` in autom8y-telemetry silently swallowed an `ImportError` when no logger was passed, so `instrument_lambda` decorated the handler, spans were created, but nothing was ever exported to Tempo.

**When**: Discovered during SRE Sprint 2 observability work. Fixed in commit `f3bc481` (2026-03-06). Recorded as `SCAR-SRE-011` in `.know/scar-tissue.md`.

**How fixed**: Added `[project.optional-dependencies] otlp = ["autom8y-telemetry[otlp]>=0.5.0"]` pin to `pyproject.toml`. SDK fix: `_create_exporter()` now always logs via stdlib logging when OTLP package is missing.

---

### SCAR-SA-008: AUTOM8Y_ENV env prefix masking

**What failed**: If a pydantic-settings child class sets a custom `env_prefix` (e.g., `SLACK_`), it would look for `SLACK_AUTOM8Y_ENV` instead of the canonical `AUTOM8Y_ENV`. In ECS/Lambda where only `AUTOM8Y_ENV` is set, `autom8y_env` would default to `LOCAL`, triggering production URL guards and incorrect behavior.

**When**: Ecosystem-wide fix in commit `1367461` (2026-03-08). slack-alert received the version bump as a consumer.

**How fixed**: Added `validation_alias=AliasChoices("autom8y_env", "AUTOM8Y_ENV")` to the base field in `autom8y-config` SDK (v1.2.1). Pydantic-settings now always reads the canonical env var regardless of child class prefix.

---

### SCAR-SA-009: SLACK_BOT_TOKEN env var name drift

**What failed**: The service originally used `SLACK_BOT_TOKEN` as the env var name (holding a secret ARN). The hotfix (CHAOS-001) switched to `SLACK_BOT_TOKEN_ARN` to match the ARN auto-resolution convention of `Autom8yBaseSettings`. docker-compose.override.yml retained a comment explaining the old naming rationale ("Tier 2 third-party SDK name — cannot be renamed to AUTOM8Y_*") which was then removed as stale in commit `35f29a1` during the clean-break env var standardization (2026-03-14).

**When**: Naming transition across commits `0e9840c` -> `15c26bf` -> `35f29a1`.

**How fixed**: `SLACK_BOT_TOKEN_ARN` is the canonical env var name. Terraform wires `OBSERVABILITY_` qualifier prefix for the secrets manager ARN reference.

## Category Coverage

| Category | Scars | IDs |
|---|---|---|
| Observability / silent failure | 2 | SCAR-SA-001, SCAR-SA-007 |
| Secret management antipattern | 2 | SCAR-SA-002, SCAR-SA-009 |
| Build / Docker toolchain | 2 | SCAR-SA-004, SCAR-SA-005 |
| Dependency / CI bootstrap | 2 | SCAR-SA-003, SCAR-SA-006 |
| Configuration / env var | 1 | SCAR-SA-008 |

**Coverage breadth**: 5 distinct failure mode categories across 9 scars. All failures are in the infrastructure/deployment surface area. No runtime logic bugs in the business logic path (SNS parsing, alarm formatting, Slack posting) have been recorded in git history.

## Fix-Location Mapping

| Scar | Fix Location | Current State |
|---|---|---|
| SCAR-SA-001 | `src/slack_alert/handler.py` lines 79-85 | `RuntimeError` raise on errors > 0; explanatory comment block preserved |
| SCAR-SA-002 | `src/slack_alert/handler.py` line 31 | `SlackConfig()  # type: ignore[call-arg]` — hand-rolled `_resolve_secret()` fully removed |
| SCAR-SA-003 | `pyproject.toml` `[dependency-groups]` and `[tool.pytest.ini_options]` / `[tool.mypy]` | Present and complete |
| SCAR-SA-004 | `Dockerfile` ARG line 18: `ARG SECRETS_EXT_LAYER_URL` | Consistent with CI workflow |
| SCAR-SA-005 | `Dockerfile` line 41: `COPY --from=uv /uv /uvx /bin/` | No `--link` on the uv copy line |
| SCAR-SA-006 | `Dockerfile` lines 56-78 | Conditional `EXTRA_INDEX_URL` in both compile and install RUN steps |
| SCAR-SA-007 | `pyproject.toml` dependency on `autom8y-telemetry[otlp]>=0.5.0` | Present; SDK fix in `autom8y_telemetry/init.py` |
| SCAR-SA-008 | `pyproject.toml` `autom8y-config>=1.2.1` version pin | Ecosystem fix in SDK, consumed via version bump |
| SCAR-SA-009 | `src/slack_alert/handler.py` module docstring; Terraform `SLACK_BOT_TOKEN_ARN` | Docstring documents the ARN convention |

## Defensive Pattern Documentation

### DEF-SA-001: Per-record exception isolation before invocation failure

Spawned by SCAR-SA-001. The `_process_event` coroutine wraps each record in a `try/except Exception` block, increments `errors`, and continues. Only after all records are attempted does `lambda_handler` evaluate `result["errors"] > 0` and raise. This ensures partial delivery (some records succeed) rather than all-or-nothing on a multi-record SNS batch.

Location: `src/slack_alert/handler.py` lines 37-59 (per-record try/except), lines 84-85 (post-loop raise).

### DEF-SA-002: type: ignore suppression on ARN-resolved config

Spawned by SCAR-SA-002. The `# type: ignore[call-arg]` comment on line 31 is load-bearing documentation — it signals that `bot_token` is absent from the constructor call intentionally (resolved via model_validator at runtime). Removing this suppression without understanding the ARN resolution chain would cause a false mypy failure.

Location: `src/slack_alert/handler.py` line 31.

### DEF-SA-003: Conditional EXTRA_INDEX_URL in Dockerfile

Spawned by SCAR-SA-006. All three pip-related RUN steps in the Dockerfile use `if [ -n "$EXTRA_INDEX_URL" ]; then ... else ...; fi` branching. This makes the Dockerfile usable both in CI (where CodeArtifact URL is injected) and locally (where `UV_INDEX_AUTOM8Y_PASSWORD_COMMAND` handles auth). The branching must be preserved in both the compile and install steps.

Location: `Dockerfile` lines 56-78, 70-78, 84-89.

### DEF-SA-004: Explanatory comment block on RuntimeError raise

Spawned by SCAR-SA-001. A multi-line comment at lines 79-83 of `handler.py` explains why the handler raises on delivery failure ("Without this raise, the 99.9% SLO burn-rate alert is blind to Slack delivery failures"). This comment guards against future "cleanup" that removes the raise as seemingly unnecessary given per-record isolation above it.

Location: `src/slack_alert/handler.py` lines 79-83.

## Agent-Relevance Tagging

| Scar | Relevant Agents | Reason |
|---|---|---|
| SCAR-SA-001 | principal-engineer, qa-adversary | Any change to error handling in handler.py must preserve the RuntimeError raise; qa-adversary should test partial-failure scenarios |
| SCAR-SA-002 | principal-engineer, hallucination-hunter | ARN auto-resolution is non-obvious; hallucination-hunter should verify SlackConfig constructor behavior in any refactor |
| SCAR-SA-003 | principal-engineer | Any new service using the same pattern must include dev dependency group at bootstrap, not as a follow-up fix |
| SCAR-SA-004 | principal-engineer | Dockerfile build arg name must match `service-build.yml`; cross-check when adding Lambda services |
| SCAR-SA-005 | principal-engineer | `COPY --link` is safe for src/config files, unsafe for binary overlays (`/bin/`); pattern check for new Dockerfile stages |
| SCAR-SA-006 | principal-engineer | Private package installs require CodeArtifact URL; apply the conditional branching pattern for all new Lambda services |
| SCAR-SA-007 | hallucination-hunter, principal-engineer | Telemetry extras must be explicit; `instrument_lambda` is silent on missing exporter — no runtime error signals the gap |
| SCAR-SA-008 | principal-engineer, architect | Any new pydantic-settings subclass with custom env_prefix must inherit from `autom8y-config>=1.2.1`; do not add raw pydantic-settings BaseSettings subclasses |
| SCAR-SA-009 | principal-engineer | `_ARN` suffix convention for all secret env vars is load-bearing; do not rename to bare `SLACK_BOT_TOKEN` |

## Knowledge Gaps

1. **No regression tests for delivery failure path**: The test suite contains only a single smoke test (`test_handler_module_exists`). The SCAR-SA-001 defensive pattern (RuntimeError raise) has no test coverage. Future refactors could silently remove it.

2. **CHAOS-001 chaos test results not captured here**: The full CHAOS-001 findings are in `.sos/sessions/PAO-20260304-1349/SRE-SPRINT1-chaos-001-results.md`. This scar tissue document captures the outcome fix but not the chaos methodology that discovered it.

3. **Terraform-side scar history not observed**: This audit scoped to `services/slack-alert/` Python source. Terraform changes to `terraform/services/slack-alert/` (SLACK_BOT_TOKEN_ARN wiring, separate alerts bot token) are referenced in commit messages but not audited here.

4. **No SCAR- code markers in source**: The codebase uses `SCAR-SRE-*` tags in `.know/scar-tissue.md` (ecosystem-level) but does not embed `// SCAR-SA-*` markers inline in source code. Scars are traceable only via git log and the separate `.know/` knowledge file.
