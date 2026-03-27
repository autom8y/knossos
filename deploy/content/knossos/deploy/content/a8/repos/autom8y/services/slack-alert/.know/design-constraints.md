---
domain: design-constraints
generated_at: "2026-03-16T20:00:00Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "faee0db"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog

### TENSION-001: Third-Party Prefix Lock on Slack SDK vs. Ecosystem Naming Convention

**Location**: `src/slack_alert/handler.py:31`, `secretspec.toml:27-30`

The ecosystem-wide convention is `AUTOM8Y_*` for platform variables (Tier 2), but `SlackConfig` in `autom8y-slack` SDK uses `env_prefix="SLACK_"`, making the runtime env var `SLACK_BOT_TOKEN`. This is a naming mismatch between ecosystem convention and SDK constraint.

The SDK exposes a `validation_alias=AliasChoices("AUTOM8Y_SLACK_BOT_TOKEN")` as a workaround for non-Lambda environments, but the Lambda production path uses `SLACK_BOT_TOKEN_ARN` (computed from the `SLACK_` prefix), not the canonical `AUTOM8Y_SLACK_BOT_TOKEN` name. The `secretspec.toml` documents this explicitly: "The SLACK_ prefix is a third-party SDK limitation documented in ADR Decision 2."

The docker-compose override underwent a rename (`SLACK_BOT_TOKEN` -> `AUTOM8Y_SLACK_BOT_TOKEN`) in commit `5f94481`, reflecting the canonical org-level name — yet the Lambda production resolution path still uses `SLACK_BOT_TOKEN_ARN`. The two names coexist with different scopes and neither is wrong, but any engineer reading the code must understand this split.

**Classification**: Naming mismatch / structural constraint from third-party SDK

---

### TENSION-002: mypy Opacity on Runtime Secret Injection

**Location**: `src/slack_alert/handler.py:31`

```python
config = SlackConfig()  # type: ignore[call-arg]  # bot_token resolved via _ARN env var
```

`SlackConfig.bot_token` is a required `SecretStr` field. At construction time, no argument is passed. The `Autom8yBaseSettings._resolve_lambda_arns` model validator populates it at runtime by inspecting environment variables for `SLACK_BOT_TOKEN_ARN`. mypy cannot see this — it sees a required field with no value and flags an error.

The `# type: ignore[call-arg]` suppression is load-bearing: removing it breaks `mypy --strict` CI. This is a permanent mypy opacity introduced by the ARN-resolution pattern in `autom8y-config`. Any future refactor of `SlackConfig` must keep this suppression in place or the base class must gain a mypy plugin/protocol that satisfies strict type checking.

**Classification**: Tooling tension (mypy vs. runtime injection) / load-bearing suppression

---

### TENSION-003: Sync Lambda Entry Point Wrapping Async Logic

**Location**: `src/slack_alert/handler.py:64-90`

The Lambda entry point `lambda_handler` is synchronous (required by AWS Lambda runtime), but all business logic lives in the async `_process_event` coroutine. The bridge is `asyncio.run(_process_event(event))` at line 77.

This `asyncio.run()` call creates a new event loop on every Lambda invocation. Lambda cold-start initializes the interpreter once, but Lambda warm invocations reuse the same process. `asyncio.run()` is safe for warm invocations because it creates a fresh loop each time and closes it after completion — but it means no event loop reuse between invocations. If a future engineer attempts to hoist `client` or `loop` to module-level for performance, they will encounter Lambda warm-reuse semantics that conflict with `asyncio.run()`.

**Classification**: Structural layering (sync adapter over async core) / evolution constraint

---

### TENSION-004: Error-Reporting Behavior Diverges from Per-Record Isolation

**Location**: `src/slack_alert/handler.py:83-85`

```python
if result["errors"] > 0:
    raise RuntimeError(f"Failed to deliver {result['errors']} of {result['processed']} alerts")
```

The design deliberately allows all SNS records to be attempted (per-record isolation, loop continues on exception at line 57), then raises at the end if any failed. This is intentional: it surfaces failures to the CloudWatch Lambda Errors metric for SLO burn-rate alerting (documented inline comment, introduced in commit `0e9840c`, CHAOS-001 P0).

The tension: SNS re-delivery semantics. When Lambda raises, SNS retries the entire batch. If only 1 of N records failed, the N-1 successful records will be re-delivered. For a notification service, this causes duplicate Slack alerts for the N-1 records that already succeeded. The comment acknowledges the SLO motivation but does not document this re-delivery trade-off.

**Classification**: Behavioral trade-off (observability SLO vs. duplicate delivery) / undocumented consequence

---

### TENSION-005: No Dedicated Config Class — os.environ Direct Read

**Location**: `src/slack_alert/handler.py:29`, `secretspec.toml:11`

```python
channel = os.environ.get("SLACK_CHANNEL", "#platform-alerts")
```

Unlike other autom8y services (which have a dedicated `Config(Autom8yBaseSettings)` class), `slack-alert` has no Settings class. `SLACK_CHANNEL` is read directly via `os.environ.get()` in the handler body. The `secretspec.toml` documents this explicitly: "Note: slack-alert does not have a Settings class."

This is a divergence from fleet convention. The justification (not explicitly documented) is that the service is minimal enough that a Settings class would be over-engineering — the only non-SDK config value is `SLACK_CHANNEL`. However, this means `SLACK_CHANNEL` is not validated, has no type annotation, and cannot be mypy-checked. A future engineer adding more config might add it to `os.environ` rather than creating a Settings class, compounding the deviation.

**Classification**: Under-engineering / fleet convention deviation

---

### TENSION-006: Orphaned Shared Secrets Manager Path

**Location**: `terraform/shared/main.tf` (creates `autom8y/slack/bot-token`), `secretspec.toml:29`

Three Secrets Manager paths exist: `autom8y/slack/bot-token` (created by shared Terraform module), `autom8y/slack/alerts-bot-token` (default for `slack-alert`), and `autom8y/slack/reporting-bot-token` (default for reconcile-spend). The shared module path `autom8y/slack/bot-token` is not referenced by default in any service's Terraform defaults. This is documented as pending human resolution in `.ledge/reviews/WS3-slack-token-identity-investigation.md` (Gap G-4, status: `PENDING_HUMAN_ACTION`).

**Classification**: Infrastructure naming tension / unresolved human-gated ambiguity

## Trade-off Documentation

### Trade-off 1: Raise on Partial Failure (SLO Observability vs. Duplicate Delivery)

**Chosen**: Raise `RuntimeError` after all records attempted, even if only some failed.
**Rejected**: Return success with error count in response body (or silently swallow).
**Why chosen**: The 99.9% SLO burn-rate alert is driven by the Lambda Errors metric in CloudWatch. Without the raise, Slack delivery failures (e.g., `invalid_auth`, `channel_not_found`) would be invisible to the SLO dashboard.
**Cost**: SNS retries the full batch on any partial failure, causing duplicate delivery for already-successful records.
**Evidence**: Commit `0e9840c` (CHAOS-001 P0), inline comment at `src/slack_alert/handler.py:79-82`.

### Trade-off 2: SDK ARN Resolution vs. Per-Service Secret Bridge

**Chosen**: Delegate to `SlackConfig()` with runtime ARN resolution via `Autom8yBaseSettings._resolve_lambda_arns`.
**Rejected**: Per-service hand-rolled `_resolve_secret()` function (removed in commit `15c26bf`).
**Why chosen**: Removes duplicated secret-resolution logic across all Lambda services; the base class handles `*_ARN` env var pattern uniformly.
**Cost**: Introduces a mypy blind spot (`# type: ignore[call-arg]`); makes the secret injection invisible to static analysis.
**Evidence**: Commits `15c26bf`, `717322a`; comment at `src/slack_alert/handler.py:8`.

### Trade-off 3: No Dedicated Config Class

**Chosen**: Read `SLACK_CHANNEL` via `os.environ.get()` in handler body.
**Rejected**: `class SlackAlertConfig(Autom8yBaseSettings)` wrapper.
**Why chosen**: Service is minimal (one handler file, one env var beyond SDK config); a Settings class would be over-engineering for a single optional string.
**Cost**: `SLACK_CHANNEL` is unvalidated, untyped, and not visible to mypy. Fleet convention deviation that future engineers may extend ad hoc.
**Evidence**: `secretspec.toml:11-13`, `src/slack_alert/handler.py:29`.

### Trade-off 4: IAM via Standalone Policy (Not stack-managed secrets_arn_patterns)

**Chosen**: Empty `secrets_arn_patterns = []` in Terraform, with a standalone IAM policy passed via `additional_iam_policies`.
**Rejected**: Fleet-standard `secrets_arn_patterns` for IAM scoping.
**Why chosen**: Documented as "predates the stack... avoid destroy/create cycle" (from IAC-SLACK-ALERT-001 in `.ledge/reviews/iac-contract-triage-slack-alert.md`).
**Cost**: Fleet inconsistency; future maintainers may not understand why this service diverges from the pattern.
**Evidence**: `.ledge/reviews/iac-contract-triage-slack-alert.md`, finding IAC-SLACK-ALERT-001.

### Trade-off 5: X-Ray Disabled (Not fleet default)

**Chosen**: `enable_xray = false`, `enable_insights = false` in Terraform.
**Rejected**: Fleet standard (`enable_xray = true`).
**Why chosen**: "Match current service defaults (not stack defaults)" — intentional divergence documented in IAC-SLACK-ALERT-002.
**Cost**: X-Ray traces unavailable for this service; fleet inconsistency in observability coverage.
**Evidence**: `.ledge/reviews/iac-contract-triage-slack-alert.md`, finding IAC-SLACK-ALERT-002.

## Abstraction Gap Mapping

### Missing: Typed Configuration Class

The service lacks a `SlackAlertConfig(Autom8yBaseSettings)` class. `SLACK_CHANNEL` is read via raw `os.environ.get()`. If the service gains more configuration (e.g., retry count, timeout overrides, channel routing by severity), there is no typed home for it. The pattern will likely be extended ad hoc.

**Impact**: Low (current scope), Medium (if service grows).
**Files**: `src/slack_alert/handler.py:29`

### Missing: Batch Retry or Dead Letter Queue Design

The handler catches all exceptions per-record and counts errors. There is no retry, no dead-letter queue write, and no per-record error context beyond logging. If `format_cloudwatch_alarm` raises on a malformed alarm payload, the alarm is silently dropped (error count incremented, no payload preserved).

**Impact**: Medium (silent alarm loss without forensic context).
**Files**: `src/slack_alert/handler.py:38-58`

### Missing: Integration Tests

The test suite (`tests/test_handler.py`) contains a single smoke test verifying `lambda_handler` is importable. There are no tests that exercise the full event processing path — SNS envelope parsing, CloudWatch alarm JSON parsing, Slack API interaction, or the error-raise behavior on partial failures. The `respx` mock library is listed as a dev dependency, but no HTTP mocking tests exist.

**Impact**: High (entire handler behavior is untested).
**Files**: `tests/test_handler.py`

### Premature Abstraction: None identified

The service is lean (1 handler file, 91 lines). No premature abstractions were found.

## Load-Bearing Code Identification

### LB-001: `# type: ignore[call-arg]` at handler.py:31

**File**: `src/slack_alert/handler.py:31`
**Why load-bearing**: This suppression is the only thing preventing `mypy --strict` CI failure. `SlackConfig()` has a required `bot_token` field with no default; mypy correctly flags this as a missing argument. The runtime ARN resolution (`_resolve_lambda_arns` model validator) populates the field after construction. Removing this comment without simultaneously giving mypy a way to see the injection (e.g., a stub, a protocol, a factory function) will break CI.

### LB-002: The RuntimeError raise at handler.py:83-85

**File**: `src/slack_alert/handler.py:83-85`
**Why load-bearing**: This raise is what causes the Lambda Errors metric to fire in CloudWatch, which drives the 99.9% SLO burn-rate alert. Removing it (e.g., in an attempt to "fix" the batch re-delivery issue) would silently break SLO monitoring without any immediate observable failure. The monitoring depends on this behavior.

### LB-003: Per-record exception isolation loop at handler.py:37-58

**File**: `src/slack_alert/handler.py:37-58`
**Why load-bearing**: The loop must not be converted to `asyncio.gather(*[...])` with default `return_exceptions=False`, because that would short-circuit on the first failure and skip remaining records entirely. The current design ensures all records are attempted before the error count is checked. This is paired with LB-002.

### LB-004: `asyncio.run()` at handler.py:77

**File**: `src/slack_alert/handler.py:77`
**Why load-bearing**: Lambda does not provide a persistent event loop. `asyncio.run()` creates a fresh loop and closes it. Moving to module-level `loop = asyncio.get_event_loop()` or `loop = asyncio.new_event_loop()` would be incorrect for Lambda because warm invocations reuse the process but the loop state is not guaranteed to be clean. Do not replace `asyncio.run()` with a cached loop without understanding Lambda warm-start lifecycle.

### LB-005: `@instrument_lambda` decorator at handler.py:64

**File**: `src/slack_alert/handler.py:64`
**Why load-bearing**: This decorator wraps the handler to inject OpenTelemetry trace context. Removing it would sever the Lambda span from the distributed trace. The telemetry pipeline (ADOT -> OTLP gateway -> Tempo) depends on this decorator being present on the entry point.

## Evolution Constraint Documentation

### EC-001: SLACK_ Prefix Cannot Be Changed Without SDK Change

The `SLACK_BOT_TOKEN_ARN` env var name is derived by the `Autom8yBaseSettings._resolve_lambda_arns` method from the SDK's `env_prefix="SLACK_"` on `SlackConfig`. Changing the Terraform injection name from `SLACK_BOT_TOKEN` to `AUTOM8Y_SLACK_BOT_TOKEN` (for ecosystem alignment) would require updating both the SDK field prefix and the base class ARN resolution logic — not just this service.

**Constraint**: Env var prefix of Slack token is locked to third-party SDK convention until `autom8y-slack` SDK is modified.
**Files**: `src/slack_alert/handler.py:8,31`, `secretspec.toml:27-30`

### EC-002: SNS Event Schema is Load-Bearing

The handler assumes SNS event structure at `record["Sns"]["Message"]` and CloudWatch alarm JSON structure at `alarm["AlarmName"]`, `alarm["NewStateValue"]`. These are AWS-defined schemas. Any change to the event source (e.g., EventBridge instead of SNS, or a custom alarm format) would require handler changes.

**Constraint**: Handler is tightly coupled to SNS+CloudWatch Alarms event shape.
**Files**: `src/slack_alert/handler.py:39-54`

### EC-003: Slack Block Kit Format Is SDK-Owned

Message formatting is fully delegated to `format_cloudwatch_alarm()` from `autom8y-slack`. The handler has no formatting logic. Adding service-specific formatting (e.g., enriching with additional context, severity routing) would require either SDK extension or post-processing of the `message` object before posting.

**Constraint**: Cannot customize alarm message format without touching `autom8y-slack` SDK or adding a post-processing step.
**Files**: `src/slack_alert/handler.py:42-48`

### EC-004: Single Channel for All Alarms

`SLACK_CHANNEL` is read once per invocation and applied to all records in the batch. There is no per-alarm channel routing. Routing different alarm types to different channels would require either a per-alarm routing table (new logic in handler) or multiple Lambda functions with different `SLACK_CHANNEL` values.

**Constraint**: No per-alarm channel routing capability in current design.
**Files**: `src/slack_alert/handler.py:29`

### EC-005: Two Slack Bot Tokens, Not One

The service uses `autom8y/slack/alerts-bot-token` (the "Autom8 Alerts" app), not `autom8y/slack/reporting-bot-token` (the "automation_assistant" bot used by reconcile-spend/ads). Adding functionality that posts to the same channels as other services may require token consolidation or explicit routing — the identity split is load-bearing for token/app isolation.

**Constraint**: slack-alert must use the Autom8 Alerts Slack app token, not the general automation_assistant token.
**Evidence**: `.ledge/reviews/WS3-slack-token-identity-investigation.md`

## Risk Zone Mapping

### RISK-001: Unguarded SNS Batch Re-Delivery on Partial Failure

**Location**: `src/slack_alert/handler.py:83-85`
**Risk**: When any record fails, Lambda raises and SNS retries the full batch. Successfully delivered records produce duplicate Slack alerts. There is no deduplication guard (e.g., alarm name + state + timestamp as idempotency key).
**Current defense**: None.
**Missing defense**: Idempotency key check before posting; or DLQ for failed records only.
**Severity**: Medium (operational noise, not data loss).

### RISK-002: Silent Alarm Drop on Malformed Payload

**Location**: `src/slack_alert/handler.py:57-58`

```python
except Exception:
    log.exception("failed_to_post_alert")
    errors += 1
```

If `json.loads(sns_message)` raises on a malformed SNS message, or `format_cloudwatch_alarm(alarm)` raises on an unexpected alarm schema, the record is counted as an error but the payload is not preserved. Log output contains the exception traceback but no payload dump. Debugging a silent alarm drop requires CloudWatch Logs forensics.
**Current defense**: `log.exception()` writes traceback. RuntimeError raised at end surfaces in Lambda Errors metric.
**Missing defense**: Structured log field with sanitized payload excerpt for triage; DLQ write.
**Severity**: Medium.

### RISK-003: SLACK_CHANNEL with No Validation

**Location**: `src/slack_alert/handler.py:29`
**Risk**: `os.environ.get("SLACK_CHANNEL", "#platform-alerts")` returns whatever string is in the environment. An invalid channel name (e.g., a typo, a private channel the bot is not in) will not raise at startup — it will fail at `client.post_message()` and increment the error counter. The failure mode is identical to a Slack API error.
**Current defense**: Default of `#platform-alerts` prevents None.
**Missing defense**: No startup validation that the configured channel is reachable.
**Severity**: Low (caught at runtime, surfaced via Lambda Errors metric).

### RISK-004: Secrets Manager Path Ambiguity (Unresolved Human Gate)

**Location**: `terraform/services/slack-alert/variables.tf` (not in `services/slack-alert/` scope, but affects this service)
**Risk**: Three Secrets Manager paths exist (`autom8y/slack/bot-token`, `autom8y/slack/alerts-bot-token`, `autom8y/slack/reporting-bot-token`). The shared Terraform module creates `autom8y/slack/bot-token`, but slack-alert defaults to `autom8y/slack/alerts-bot-token`. If the former is populated but the latter is not, the service will fail to authenticate.
**Current defense**: Documented in `.ledge/reviews/WS3-slack-token-identity-investigation.md` as PENDING_HUMAN_ACTION.
**Missing defense**: Terraform validation or smoke test confirming the configured SM path exists and is populated.
**Severity**: High (would cause total service failure at startup).

### RISK-005: No Integration Tests — Full Handler Path Untested

**Location**: `tests/test_handler.py`
**Risk**: The only test is a smoke import check. The SNS parsing path, CloudWatch alarm JSON parsing, Slack API interaction, error counting, and the RuntimeError raise on partial failure are all exercised only in production. `respx` is a dev dependency but is unused.
**Current defense**: mypy + ruff catch type and style errors.
**Missing defense**: Integration tests mocking Slack HTTP responses (`respx` is already available).
**Severity**: High (any behavioral regression is invisible until production).

## Knowledge Gaps

1. **ADR-ENV-NAMING-CONVENTION** referenced in `secretspec.toml:5` — this document does not appear to exist in `.ledge/decisions/`. The closest document is `ADR-auth-key-naming-convention.md`. The referenced ADR and its "Decision 2" about the SLACK_ prefix constraint could not be verified.

2. **Terraform module details** for `service-lambda-event-driven` — the IaC triage document references `main.tf:135-136` and `secrets_arn_patterns`, but the Terraform directory is outside the audit scope (`services/slack-alert/`). The full IAM policy structure for secret access cannot be verified from this scope.

3. **`autom8y/slack/bot-token` Secrets Manager path resolution** — documented as PENDING_HUMAN_ACTION in `.ledge/reviews/WS3-slack-token-identity-investigation.md`. The actual token identity and whether this path is live or orphaned is unresolved.

4. **`autom8y-slack` SDK internals** — `SlackConfig` field definition, `validation_alias` behavior, and `_resolve_lambda_arns` implementation were referenced from comments and ledge documents rather than read directly (SDK source outside `services/slack-alert/` scope).
