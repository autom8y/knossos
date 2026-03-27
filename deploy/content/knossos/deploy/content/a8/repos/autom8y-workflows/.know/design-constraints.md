---
domain: design-constraints
generated_at: "2026-03-27T09:41:06Z"
expires_after: "7d"
source_scope:
  - "./src/**/*"
generator: theoros
source_hash: "094fa67"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

> Knowledge reference for `/Users/tomtenuta/Code/a8/repos/autom8y-workflows`. Primary language: YAML (GitHub Actions workflow). Single-file codebase with a 10-commit history. All design constraints observed in `.github/workflows/satellite-ci-reusable.yml` and `README.md`.

## Tension Catalog Completeness

This codebase is a single-file, YAML-based GitHub Actions workflow repository. Structural tensions are fewer than in a large multi-package source tree, but several load-bearing tensions exist.

### TENSION-001
- **Type**: Duplication / Layering (step-level duplication across jobs)
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 125-144 (lint job), 191-209 (test job), 286-305 (integration job), 344-363 (convention-check job)
- **Description**: The four-step block for AWS/CodeArtifact setup (Configure AWS credentials -> Get CodeArtifact auth token -> Configure uv for CodeArtifact) is repeated verbatim in every job. There is no mechanism in GitHub Actions reusable workflows to define shared step sequences -- composite actions are the only factoring mechanism, but none are defined here.
- **Historical reason**: GitHub Actions does not support shared step sequences within a workflow file. The only factoring options are composite actions (a separate repository/file) or reusable workflows (a separate workflow). The current design chose inline duplication over adding another layer of indirection.
- **Ideal resolution**: Extract the AWS+CodeArtifact setup block into a composite action (e.g., `autom8y/autom8y-workflows/.github/actions/setup-codeartifact/action.yml`). This would reduce each job's boilerplate by ~10 lines.
- **Resolution cost**: Medium. Requires creating a new action file, updating all four jobs, and testing across satellite repos. The public-repo constraint (see README) means any composite action must also be in this repo.

### TENSION-002
- **Type**: Naming mismatch / Leaky abstraction
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 109, 174, 259, 326 (`UV_NO_SOURCES: "1"`) and all `uv run --no-sources` / `uv sync --no-sources` calls
- **Description**: `UV_NO_SOURCES=1` as an env var does not work -- uv does not recognize it (per commit `8b5ba88`). The env var is still set at job level on every job as dead config, while the actual enforcement is via `--no-sources` flags on every individual command. This creates a false signal: the env var looks authoritative but is inoperative. Any agent or developer reading the env block would believe `UV_NO_SOURCES` controls behavior; it does not.
- **Historical reason**: The env var was added first (`a719b0f`) as a "clean" single-source approach. When it turned out uv doesn't recognize it, explicit `--no-sources` flags were added to each command (`8b5ba88`) rather than removing the now-dead env var. The env var was retained because it serves as documentation (signals intent) even though it is inoperative.
- **Ideal resolution**: Either (a) remove `UV_NO_SOURCES: "1"` from all job env blocks and add a comment near the first `--no-sources` flag explaining why it must be explicit, or (b) file a uv upstream request to support the env var and restore the env-only approach when support lands.
- **Resolution cost**: Low. Removing dead env vars is a cosmetic change with no behavioral risk. Risk is only confusion during the removal PR review.

### TENSION-003
- **Type**: Dual-system / Incomplete abstraction (convention-check filtering)
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 442-456 (inline Python in `run:`)
- **Description**: The convention-check step uses inline Python within a `run:` block to filter violations by span name dot-notation. This filtering logic went through two iterations: first filtering by `type != UNKNOWN_ATTRIBUTE` (commit `9ab5eb7`), then superseded by filtering by `'.' in span_name` (commit `094fa67`). The previous filter was semantically wrong (type-based rather than name-based) and was replaced, but the underlying problem -- that `convention-check` emits violations for framework spans that the CI job cannot suppress via CLI flag -- remains. The workaround is a post-processing Python snippet inline in YAML, which is fragile (YAML indentation, shell escaping, no testing).
- **Historical reason**: The `convention-check` CLI lacks a `--filter-framework-spans` or `--domain-only` flag. Two attempts have been made to implement the filter (UNKNOWN_ATTRIBUTE type, then dot-notation on span name) without modifying the upstream CLI.
- **Ideal resolution**: Add a `--filter` or `--domain-spans-only` flag to the `convention-check` CLI (owned by the `autom8y_telemetry` package in a satellite repo). The inline Python block could then be replaced by a single CLI flag.
- **Resolution cost**: High (cross-repo). Requires modifying `autom8y_telemetry`, releasing a new version, and updating all satellites that pin to the old version. The inline workaround is currently load-bearing.

### TENSION-004
- **Type**: Missing abstraction (secrets coupling)
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 265-269 (integration job, `actions/create-github-app-token`)
- **Description**: The integration job requires `secrets.APP_ID` and `secrets.APP_PRIVATE_KEY` to be available in every satellite that calls this workflow with `run_integration: true`. These secrets are not declared in the `workflow_call` inputs/secrets block -- they are accessed as bare `secrets.APP_ID` references, relying on `secrets: inherit` in the caller. If a satellite does not pass `secrets: inherit`, or if the org-level secrets are not named exactly `APP_ID` and `APP_PRIVATE_KEY`, the job fails silently at startup (the token generation step fails with a 404-equivalent, and the job reports `startup_failure`).
- **Historical reason**: GitHub Actions reusable workflows can declare explicit secrets in `workflow_call.secrets` or rely on `secrets: inherit`. The explicit approach requires callers to pass each secret individually. The implicit `inherit` approach is simpler for callers but makes the contract invisible. The current design chose implicit inheritance for minimal caller YAML.
- **Ideal resolution**: Add explicit `secrets:` declarations under `workflow_call` for `APP_ID` and `APP_PRIVATE_KEY` (both `required: false`, only needed when `run_integration: true`). This surfaces the contract in the workflow definition.
- **Resolution cost**: Low-medium. Requires updating callers to pass secrets explicitly, but the secrets themselves already exist at org level.

### TENSION-005
- **Type**: Layering violation / Region coupling
- **Location**: `.github/workflows/satellite-ci-reusable.yml` line 129 (`vars.AWS_ACCOUNT_ID`), line 130 (`aws-region: us-east-1`)
- **Description**: The AWS region `us-east-1` is hardcoded in all four jobs. `AWS_ACCOUNT_ID` was externalized to an org variable (commit `224e30d`), but the region was not. If the org ever migrates to a different AWS region, four places in the workflow must be updated. The region is not an input parameter, not a variable, and not documented as a conscious choice.
- **Historical reason**: `us-east-1` is a widely-assumed default. The AWS account ID was hardcoded first and was only externalized after it appeared in a diff review comment. The region was never scrutinized similarly.
- **Ideal resolution**: Add `aws_region` as an optional input (default: `us-east-1`) or externalize to an org variable like `AWS_REGION`. Low priority unless a region migration is planned.
- **Resolution cost**: Low.

### TENSION-006
- **Type**: Asymmetric behavior / Incomplete guarding
- **Location**: `.github/workflows/satellite-ci-reusable.yml` line 317 (`continue-on-error: true` on integration tests)
- **Description**: The integration test job uses `continue-on-error: true`, meaning integration test failures do not block merge. This is an explicit opt-in but creates a behavioral asymmetry: unit test failures block CI (no `continue-on-error`), integration failures do not. The README and workflow header do not document this asymmetry. Any developer checking CI green to merge will not see integration failures as blockers.
- **Historical reason**: Integration tests hit live external APIs and are expected to be flaky in CI. The `continue-on-error` pattern is the standard GitHub Actions mechanism for advisory-only checks. The choice was made to prevent live API flakiness from blocking feature work.
- **Ideal resolution**: Document the asymmetry in the workflow header comment. Optionally expose `integration_continue_on_error` as an input (default: `true`) so satellites that want hard-fail integration tests can opt in.
- **Resolution cost**: Low (documentation only), or Low-medium (adding input).

### TENSION-007
- **Type**: Missing abstraction (dependency version pinning)
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 114, 121, 126, 246, 265 (pinned action SHA hashes)
- **Description**: All GitHub Actions are pinned to commit SHAs (e.g., `actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5`). The SHAs appear in four jobs for the same three actions (checkout, setup-uv, setup-python, configure-aws-credentials). Each SHA is duplicated 4 times in the file. There is no single source of truth: if `setup-uv` needs to be updated to a new version, the SHA must be changed in 4 places. The comment `# v4` is the only human-readable version marker.
- **Historical reason**: SHA pinning is a security best practice for Actions supply chain security. The duplication is a side effect of the same job-level duplication in TENSION-001. If the AWS setup block were extracted to a composite action, SHA duplication would be partially resolved.
- **Ideal resolution**: Dependabot (already implicit in the org) can automate SHA updates. The deeper fix is TENSION-001 resolution (composite action). Alternatively, document the "update all 4 occurrences" procedure in a CONTRIBUTING note.
- **Resolution cost**: Low (process), Medium (structural fix via TENSION-001).

### TENSION-008
- **Type**: Naming mismatch (test job vs. convention-check job design)
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 166-249 (test job) vs. lines 319-456 (convention-check job)
- **Description**: The test job and convention-check job both run pytest but for different purposes. The test job runs with `test_markers_exclude` (default: `not integration`), standard coverage flags, and produces `coverage.xml`. The convention-check job runs pytest with a different filter (`convention_check_test_filter`) but does NOT measure coverage. These two jobs can execute overlapping tests with different configurations on the same commit. There is no de-duplication mechanism, no shared artifact passing, and no guarantee of consistency (e.g., if `convention_check_test_filter` is a subset of what the test job already runs, spans are collected twice).
- **Historical reason**: The convention-check job was added later (`7958509`) as a separate, independent job rather than as a step extension of the test job. This was likely to avoid coupling the span collection logic into the test job's step sequence, and to keep it opt-in via `convention_check: false` default.
- **Ideal resolution**: Consider combining span collection into the test job when `convention_check: true`, writing spans as a job artifact, and running convention-check as a dependent job that consumes the artifact. This avoids double test execution.
- **Resolution cost**: High (significant job restructuring, artifact upload/download, conditional logic).

---

## Trade-off Documentation

Each tension above includes trade-off rationale. This section synthesizes the persistent trade-offs and documents external constraints.

### No ADRs on file

`.ledge/decisions/` does not exist. No formal ADRs are present. All trade-off evidence is reconstructed from git history and inline comments.

### Trade-off: Public repository constraint (load-bearing architectural constraint)
- **Current state**: Repository is public. Satellites can call its workflows.
- **Ideal state**: No change. This is the correct state.
- **Why current state persists**: GitHub enforces that public repos cannot call reusable workflows from private repos. This is an external platform constraint, documented in `README.md` lines 28-36.
- **Impact**: Any maintainer who makes this repo private will break all satellite CI silently. This constraint is explicitly documented in the README but is not enforced by any technical control.

### Trade-off: OIDC for CodeArtifact (security vs. complexity)
- **Current state**: Every job authenticates via OIDC to AWS to get a CodeArtifact token. This requires `id-token: write` permission in every job, and the calling workflow must also declare `id-token: write`.
- **Ideal state**: No change. OIDC is the correct approach over stored credentials.
- **Why current state persists**: CodeArtifact requires an AWS auth token. OIDC is the only secret-free mechanism. The `id-token: write` + role-to-assume + CodeArtifact token 3-step is mandated by the AWS/GitHub integration.
- **Prior failure**: The calling workflow must declare permissions explicitly. The failure mode when it does not is opaque -- GitHub silently creates 0 jobs, reported as `startup_failure`. This is documented in the workflow header (lines 19-21).

### Trade-off: Implicit secrets contract for GitHub App (visibility vs. simplicity)
- **Current state**: `APP_ID` and `APP_PRIVATE_KEY` are accessed via `secrets: inherit`. No explicit declaration in `workflow_call`.
- **Ideal state**: Explicit `secrets:` block in `workflow_call`.
- **Why current state persists**: Simpler for callers. Explicit declaration requires callers to name-map each secret. The `secrets: inherit` pattern was established org-wide before this workflow.

### Trade-off: Inline Python for convention-check filtering (correctness vs. upstreamability)
- **Current state**: Post-processing Python block inline in YAML filters violations before failing CI.
- **Ideal state**: `convention-check` CLI supports `--domain-only` or equivalent flag.
- **Why current state persists**: CLI modification requires cross-repo coordination. The inline workaround was faster to ship. Two iterations of the filter (type-based, then name-based) indicate the correct filtering heuristic is still evolving.
- **External constraint**: `autom8y_telemetry` is maintained in a separate satellite repo. This workflow cannot unilaterally modify the CLI.

### Trade-off: `continue-on-error: true` on integration tests
- **Current state**: Integration failures are advisory. CI is green even if integration tests fail.
- **Ideal state**: Documented asymmetry. Possibly: satellites opt in to hard-fail.
- **Why current state persists**: Integration tests hit live APIs. Flakiness from external services should not block feature work.
- **No prior refactoring attempts** on this trade-off. It is a deliberate initial choice.

---

## Abstraction Gap Mapping

### Missing Abstractions

**MA-001: AWS/CodeArtifact setup block**
- Duplicated in 4 locations: lint job (lines 125-144), test job (lines 191-209), integration job (lines 286-305), convention-check job (lines 344-363)
- Each instance is identical: Configure AWS credentials -> Get CodeArtifact auth token -> Configure uv credentials
- Recommended abstraction: composite action at `.github/actions/setup-codeartifact/action.yml`
- Maintenance burden: Any change to the IAM role name, AWS region, CodeArtifact domain (`autom8y`), or credential env var names must be made in 4 places. The region hardcode (`us-east-1`) in particular appears 4 times (TENSION-005).

**MA-002: `uv` setup steps**
- The three-step uv/Python setup sequence (`actions/checkout`, `astral-sh/setup-uv`, `actions/setup-python`) is repeated in all 4 jobs.
- Recommended abstraction: could be part of the same composite action as MA-001, or a separate composite action.
- Maintenance burden: Pinned SHA updates for `setup-uv` and `setup-python` require 4-location edits.

### Premature Abstractions

None identified. The codebase has one file and no abstraction layers. The parametric inputs (`mypy_targets`, `coverage_threshold`, etc.) are appropriately calibrated to actual caller variation observed in the git history. The `semgrep_config` and `convention_check` inputs are opt-in and default-off, so they do not represent over-abstraction.

### Zombie Abstractions

**ZA-001: `UV_NO_SOURCES: "1"` env var in every job**
- This env var was the intended abstraction point for controlling uv source resolution. It was rendered inoperative when uv was found not to honor it (commit `8b5ba88`).
- The env var is present in 4 jobs but controls nothing. The actual enforcement is via `--no-sources` CLI flags.
- File: `.github/workflows/satellite-ci-reusable.yml` lines 109, 174, 259, 326
- Documented further as TENSION-002.

---

## Load-Bearing Code Identification

### LB-001: OIDC permission requirement pattern
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 19-21 (header comment) and lines 105-107, 173-175, 257-259, 325-327 (job-level permissions blocks)
- **What it does**: Every job declares `id-token: write` and `contents: read`. This is required for OIDC to function. The header comment explains why the calling workflow must also declare permissions (without it, GitHub creates 0 jobs with `startup_failure`).
- **Dependents**: All satellite repositories that call this workflow. If any satellite omits `id-token: write` from its calling job, all 4 jobs fail at startup.
- **Naive-fix failure mode**: Removing `id-token: write` from even one job (to "clean up" permissions) would silently break CodeArtifact auth for that job. The failure is `startup_failure`, not a test failure -- it is reported as no jobs created, which is easy to misread as a workflow dispatch issue.
- **Safe-refactor requirement**: Any change to the permissions model requires testing across at least one satellite repo before merging to main, as this is a cross-repo contract.
- **Hot path**: Yes -- this is the first gate for every CI run across all satellite repos.

### LB-002: `secrets: inherit` pattern and implicit APP_ID/APP_PRIVATE_KEY contract
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 265-269
- **What it does**: The integration job uses `actions/create-github-app-token` with `secrets.APP_ID` and `secrets.APP_PRIVATE_KEY`. These are not declared in the workflow's `workflow_call` block.
- **Dependents**: All satellites with `run_integration: true`. If the org renames `APP_ID` to `GITHUB_APP_ID`, the integration job breaks for all satellites.
- **Naive-fix failure mode**: Renaming the secrets without updating all callers causes `startup_failure` on the integration job only. Because `run_integration` defaults to `false`, the failure would only surface on push-to-main runs, making it easy to miss in PR CI.
- **Safe-refactor requirement**: Any secret rename requires a coordinated update: (1) rename org secret, (2) update this workflow, (3) test with one satellite before rolling out.
- **Security boundary**: Yes -- this step mints GitHub App tokens with `contents: read` scope across the org.

### LB-003: Inline Python convention-check filter
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 442-456
- **What it does**: Filters violation list from `convention-check --format json` to exclude framework spans (those without dots in span name). Determines whether CI fails or passes.
- **Dependents**: All satellites with `convention_check: true`. The filter is the sole arbiter of pass/fail for the convention-check gate.
- **Naive-fix failure mode**: Changing the filtering heuristic (e.g., from dot-notation to a different rule) without coordinating with `autom8y_telemetry` maintainers could silently suppress legitimate domain violations or falsely block CI on framework spans.
- **Prior refactoring**: Two filter iterations already (commit `9ab5eb7` UNKNOWN_ATTRIBUTE-based, then `094fa67` dot-notation-based). The filter is not stable.
- **Safe-refactor requirement**: Any change to the filter heuristic requires (a) understanding the span naming contract from `autom8y_telemetry`, (b) testing against a satellite that emits both framework and domain spans, (c) verifying no domain violations are suppressed.

---

## Evolution Constraint Documentation

### Area: `.github/workflows/satellite-ci-reusable.yml` -- Caller-facing input contract
- **Changeability rating**: Migration (breaking change to callers)
- **Evidence**: This is a `workflow_call` workflow. All `inputs:` are part of a public API consumed by satellite repositories outside this repo. Removing or renaming any `required: true` input (`mypy_targets`, `coverage_package`) is a breaking change that immediately breaks all callers.
- **Optional inputs** have softer changeability: adding a new optional input with a default is a non-breaking additive change. Changing a default value is a soft-breaking change (may silently alter behavior in callers that rely on the default).
- **Frozen inputs** (do not remove or rename without cross-repo migration):
  - `mypy_targets` (required, used in all callers)
  - `coverage_package` (required, used in all callers)
- **In-progress evolution**: `convention_check` and related inputs were added recently (commit `7958509`). They are opt-in with `default: false`. Satellites are gradually adopting.

### Area: `.github/workflows/satellite-ci-reusable.yml` -- Job structure
- **Changeability rating**: Coordinated (multi-file in this repo, no external break if jobs are added/restructured)
- **Evidence**: Adding new jobs or restructuring existing jobs is safe as long as input defaults are preserved. Callers do not name-reference specific jobs within the reusable workflow. However, removing a job would silently drop that check from all callers.
- **Caution**: The integration job structure is more constrained (LB-001, LB-002).

### Area: `.github/workflows/satellite-ci-reusable.yml` -- Action SHA pins
- **Changeability rating**: Safe (local change, no external break)
- **Evidence**: SHA pins are implementation details. Updating a SHA to a newer version of the same action is safe for all callers as long as the action's interface is unchanged. The `# v4` comments tag the semantic version for human review.
- **Note**: Dependabot, if enabled on this repo, would automate SHA updates.

### Area: `README.md` -- Public repository warning
- **Changeability rating**: Frozen (do not make this repo private)
- **Evidence**: The README explicitly states this constraint at lines 28-36. The constraint is enforced by GitHub's platform rules, not by code. Making the repo private is irreversible without breaking all satellite CI until the repo is made public again.

### Area: IAM role `github-actions-deploy` and org variable `AWS_ACCOUNT_ID`
- **Changeability rating**: Migration (cross-system)
- **Evidence**: The IAM role name is embedded in every job's `role-to-assume` ARN (lines 129, 192, 288, 346). The account ID is sourced from `vars.AWS_ACCOUNT_ID` (externalized in commit `224e30d`). Renaming the IAM role requires updating this workflow and potentially IAM policy documents. Changing `AWS_ACCOUNT_ID` is a one-place change (org variable).

### Deprecated / In-progress migrations
- **`UV_NO_SOURCES: "1"` env vars**: De facto deprecated (TENSION-002, ZA-001). Still present in 4 jobs. No formal deprecation marker. A future agent removing these will find the behavior unchanged.
- **UNKNOWN_ATTRIBUTE filter** (commit `9ab5eb7` -> `094fa67`): The first convention-check filter implementation was replaced in the same day. The current dot-notation filter is second-generation and should be considered provisional until the upstream CLI provides a proper flag.

---

## Risk Zone Mapping

### RISK-001
- **Location**: `.github/workflows/satellite-ci-reusable.yml` line 214 (`echo '${{ inputs.test_env }}' >> $GITHUB_ENV`)
- **Type**: Shell injection via unquoted/untrusted input interpolation
- **Evidence of missing protection**: `inputs.test_env` is interpolated directly into a shell `echo` command. If a caller passes a value containing shell metacharacters or newline-injected env vars (e.g., `KEY=VALUE\nGITHUB_TOKEN=evil`), those would be written directly into `$GITHUB_ENV`. The GitHub Actions documentation warns against direct interpolation of user inputs in `run:` commands.
- **Recommended guard**: Use a sanitized write approach. Either (a) validate `test_env` format before writing (regex check that all lines match `^[A-Z_][A-Z0-9_]*=`), or (b) pass env vars as a structured input (JSON object or `key=value` with strict format validation), or (c) use the `env:` block with explicit key-value declarations rather than bulk-writing to `$GITHUB_ENV`.
- **Related tension**: TENSION-001 (step duplication means this pattern appears in both test and convention-check jobs). The convention-check job has a parallel risk at line 368 (`echo '${{ inputs.convention_check_test_env }}' >> $GITHUB_ENV`).
- **Cross-reference**: TENSION-001, TENSION-003

### RISK-002
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 428-428 (`|| true` on pytest span collection)
- **Type**: Silent failure swallowing
- **Evidence of missing protection**: The `Collect spans from instrumentation tests` step uses `|| true` on the `uv run pytest` command. A pytest crash (not a test failure, but a collection error, import error, or plugin crash) would silently succeed. The only guard is a subsequent check for `spans.json` existence (line 429-432). If the plugin writes an empty/corrupt `spans.json` before crashing, the existence check passes but span data is invalid. The `SPAN_COUNT` check (lines 433-437) only warns (not fails) on zero spans.
- **Recommended guard**: Distinguish between "pytest ran and some tests passed" and "pytest crashed during collection." Consider removing `|| true` and instead using `continue-on-error: true` at the step level, which preserves the exit code for inspection in subsequent steps.
- **Cross-reference**: TENSION-003 (the overall convention-check filtering fragility)

### RISK-003
- **Location**: `.github/workflows/satellite-ci-reusable.yml` lines 375-416 (inline `_conftest_convention_ci.py` written via heredoc)
- **Type**: Unvalidated dynamic code injection
- **Evidence of missing protection**: The convention-check job writes a Python plugin file to disk using a `cat > _conftest_convention_ci.py << 'CONFTEST_EOF'` heredoc. This file monkeypatches `InMemorySpanExporter.__init__` at import time. The plugin is not checked in, not versioned, and not tested independently. If `autom8y_telemetry`'s `InMemorySpanExporter` changes its constructor signature, the monkeypatch silently fails or raises at collection time, and the `|| true` on the pytest command (RISK-002) would swallow the error.
- **Recommended guard**: Extract the conftest plugin to a checked-in file in this repo (e.g., `.github/ci/conftest_convention_ci.py`) and copy it to the workspace. This allows version control, PR review, and independent testing.
- **Cross-reference**: TENSION-003, RISK-002

### RISK-004
- **Location**: `.github/workflows/satellite-ci-reusable.yml` line 317 (`continue-on-error: true` on integration tests)
- **Type**: Advisory-only gate with no visibility mechanism
- **Evidence of missing protection**: Integration test failures produce a yellow warning in GitHub Actions UI but no blocking signal. There is no notification mechanism, no required status check, and no integration failure trend tracking. A regression in live API integration could persist unnoticed across many merges before being detected.
- **Recommended guard**: Add a job summary step that writes integration test results to the GitHub Actions summary (`$GITHUB_STEP_SUMMARY`). Consider adding an optional input `integration_failure_notify` to post to a Slack channel or create a GitHub issue on failure.
- **Cross-reference**: TENSION-006

---

## Knowledge Gaps

1. **Satellite caller inventory**: No satellite repositories were examined. The caller contract is understood from this workflow's inputs, but the actual set of satellites, their specific input configurations, and which are using `convention_check: true` or `run_integration: true` are unknown.

2. **`autom8y_telemetry` span naming contract**: The dot-notation filter in RISK-001 and TENSION-003 assumes that domain spans always have dots in their names and framework spans do not. This assumption is unverified from this repo alone -- the actual span naming convention is defined in `autom8y_telemetry` (a separate satellite repo not in scope).

3. **IAM role structure**: The `github-actions-deploy` role ARN is referenced but its policy document, trust policy, and least-privilege status are unknown. The OIDC trust relationship (which GitHub org/repo patterns are trusted) cannot be verified from this repo.

4. **Org-level variable `AWS_ACCOUNT_ID`**: Existence confirmed from commit `224e30d`, but current value and whether it is set correctly across all environments (prod/staging) is unknowable from this repo.

5. **Dependabot configuration**: No `.github/dependabot.yml` was found. Whether action SHA pins are automatically updated is unknown.

6. **`convention-check` CLI version**: The `convention-check` command is invoked via `uv run`, but no version pin appears in this workflow. The CLI version used depends on what `uv sync --all-extras` resolves from `uv.lock` in each satellite's repo. Version drift across satellites is a possible risk not visible from this file.
