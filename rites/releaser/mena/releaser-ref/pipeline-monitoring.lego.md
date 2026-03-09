---
name: pipeline-monitoring
description: "Pipeline-monitor agent reference data: chain monitoring protocol, binary release verification stages, and failure diagnosis classification. Use when: pipeline-monitor needs monitoring procedures, retry backoff tables, or failure classification during CI verification. Triggers: chain monitoring, binary release verification, failure diagnosis, dispatch retry, e2e validation, homebrew tap."
scope: releaser
---

# Pipeline Monitoring Reference Data

## Chain Monitoring Protocol

When `pipeline_expectations` in `execution-ledger.yaml` contains chains for a repo, monitoring extends beyond the initial CI run to cover the full chain depth.

### Phased Monitoring Procedure

**Phase 1 -- Initial CI**: Monitor the primary CI workflow run as normal (existing protocol).

**Phase 2 -- Downstream Triggers**: After initial CI completes green, check for downstream workflow runs triggered by the CI completion:
```
gh run list --repo {owner/repo} --limit 10 --json status,conclusion,databaseId,name,updatedAt,event
```
Match runs against expected chain stages by workflow name and trigger event type.

**Phase 3 -- Cross-Repo Dispatch**: For dispatch_chain and deployment_chain types, monitor the target repo for runs triggered by the dispatch event:
```
gh run list --repo {target-owner/target-repo} --limit 10 --json status,conclusion,databaseId,name,updatedAt,event
```
Filter by workflow name matching the expected stage and by timing (started after the dispatch was sent).

**Phase 4 -- Deployment Verification**: For deployment_chain types, monitor until the terminal stage resolves. If the terminal stage has `has_health_check: true`, wait for the health check stage to complete before declaring the chain resolved.

### Cross-Repo Run Discovery

To find dispatch-triggered runs in a target repo:
```
gh run list --repo {target-owner/target-repo} --workflow {workflow-filename} --limit 5 --json status,conclusion,databaseId,name,updatedAt,event,createdAt
```

Filter results by:
- `event` matches the expected trigger type (e.g., `repository_dispatch`, `workflow_dispatch`, or `release` for binary release E2E chains)
- `createdAt` is after the source workflow's completion time
- `name` or workflow file matches the expected stage

### Retry Backoff for Dispatch Discovery

Cross-repo dispatches may have propagation delay. When an expected dispatch run is not found:
- Attempt 1: check immediately after source workflow completes
- Attempt 2: wait 30 seconds, check again
- Attempt 3: wait 60 seconds, check again
- Attempt 4 (final): wait 120 seconds, check again
- After 4 attempts (3 retries): classify the chain link as `dispatch_not_received`

Total maximum wait for dispatch discovery: 210 seconds (3.5 minutes).

### Extended Timeouts

| Scenario | Timeout |
|----------|---------|
| Flat CI monitoring (no chains) | 15 minutes (unchanged) |
| Chain with deployment stages | 30 minutes |
| Chain depth > 3 | 30 minutes |

### Chain Depth Tracking

For each chain being monitored, track progress through its stages:
- Record which stage is currently active (highest stage number with a running or completed workflow)
- When a stage completes, immediately check for the next stage's trigger
- A chain is resolved when its terminal stage reaches a conclusive state

### Failure Propagation

When any stage in a chain fails:
- Mark the entire chain as `failed`
- Do not wait for downstream stages (they will not trigger)
- Record the failing stage and its error details
- The chain failure contributes to the overall verdict via `all_chains_resolved`

## Binary Release Verification

For repos with `distribution_type: binary` in the execution ledger, verification covers two stages: GoReleaser CI completion and the downstream E2E validation chain. Both must pass for the repo to be marked `green`.

### Stage 1: GoReleaser CI Completion

Monitor `release.yml` (or the detected release workflow) triggered by the tag push. Standard CI polling applies (30-60 second intervals, 15-minute timeout for this stage). On green, immediately proceed to Stage 2 verification.

Once `release.yml` is green, verify the GitHub Release artifact:

```bash
# Confirm release object exists and list assets
gh release view v{version} --repo {goreleaser_release_repo}
gh release view v{version} --repo {goreleaser_release_repo} --json assets --jq '.assets[].name'
gh release view v{version} --repo {goreleaser_release_repo} --json isDraft --jq '.isDraft'
```

Required artifact checks (use `goreleaser_expected_assets` from the state map as the authoritative asset list):

| Check | Criterion |
|-------|-----------|
| Release object exists | `gh release view` exits 0 |
| Platform archives | All entries in `goreleaser_expected_assets` appear in the asset list (e.g., `ari_{version}_darwin_amd64.tar.gz`, `ari_{version}_darwin_arm64.tar.gz`, `ari_{version}_linux_amd64.tar.gz`, `ari_{version}_linux_arm64.tar.gz`) |
| Checksums file | `checksums.txt` in asset list |
| Not a draft | `isDraft` is `false` |

**Homebrew tap update** (if `goreleaser_brew_tap` non-null in the state map):
```bash
# Check for GoReleaser formula commit in tap repo after the tag push timestamp
gh api repos/{goreleaser_brew_tap}/commits --jq '.[0] | {sha: .sha, message: .commit.message, date: .commit.author.date}'
```
The most recent commit message should match `Brew formula update for {goreleaser_project_name} version {tag}` (where `{tag}` is the full tag string like `v0.3.0` — GoReleaser's `{{ .Tag }}` includes the `v` prefix). If the commit predates the tag push, apply one retry cycle (wait 60 seconds, check again) before marking `homebrew_tap_updated: false`. GoReleaser pushes the tap commit as the final step of CI — it may trail the Release creation by 30-90 seconds.

Record Stage 1 results in the `binary_release` sub-object:
```yaml
binary_release:
  github_release_exists: true|false
  assets_present: true|false        # true only if ALL goreleaser_expected_assets are found
  checksums_present: true|false
  release_is_draft: false           # false = correct; true = GoReleaser misconfiguration
  homebrew_tap_updated: true|false|null  # null when goreleaser_brew_tap is null
```

### Stage 2: E2E Validation Chain (release.published Trigger)

The GitHub Release creation event (`release: types: [published]`) automatically triggers `e2e-distribution.yml`. This is NOT a `workflow_run` trigger — it fires when GoReleaser creates (publishes) the GitHub Release object. The E2E workflow runs TWO parallel jobs:

- **macOS E2E** (`macos-e2e`): Apple Silicon runner (arm64) — validates Homebrew install path
- **Linux E2E** (`linux-e2e`): Ubuntu runner (amd64) with Docker + Linuxbrew — validates Linux distribution (amd64 only; arm64 Linux is not validated by CI)

Discover the E2E workflow run:
```bash
gh run list --repo {goreleaser_release_repo} --workflow e2e-distribution.yml --limit 5 \
  --json status,conclusion,databaseId,name,updatedAt,event,createdAt
```

Filter for runs with `event: release` started after the GoReleaser CI completion time. Apply the standard retry backoff for discovery (30s, 60s, 120s intervals) — the `release: published` event may lag the Release creation by 30-120 seconds.

**E2E timeout**: 30 minutes (extended timeout — chain with deployment stage).

**E2E assertions** (derived from `scripts/e2e-validate.sh`):

| Assertion | What passes |
|-----------|-------------|
| `brew tap {goreleaser_brew_tap}` | Tap is reachable; formula exists |
| `brew install {goreleaser_brew_tap}/{goreleaser_project_name}` | Formula installs without error on macOS arm64 and Linuxbrew (amd64) |
| `{binary_name} version` | Binary output matches the release tag version |
| Smoke test commands | Project-specific functional validation (e.g., init, basic CLI operation) |

Monitor both macOS and Linux E2E jobs. Both must be green. If either fails, classify with the standard failure table (flaky_test / regression / infra_issue / timeout). E2E macOS failure on brew operations is often `infra_issue` (slow CI, brew rate limits); E2E Linux failure on Docker image build is often `infra_issue`.

**Full binary release PASS verdict** requires all of:
1. `release.yml` CI: green
2. GitHub Release: exists, not draft, all expected assets present, `checksums.txt` present
3. Homebrew tap: updated (or null when not configured)
4. `e2e-distribution.yml`: both macOS and Linux jobs green

Do not declare a binary release `green` until Stage 2 is terminal. Record Stage 2 in `chain_results` using the standard chain schema (chain_type: `trigger_chain`, depth 2, stage 1 = release.yml, stage 2 = e2e-distribution.yml).

When `distribution_type: binary` is absent from the ledger entry (registry repos), skip binary verification entirely — no behavioral change.

## Failure Diagnosis

Pull failed logs via `gh run view {id} --repo {repo} --log-failed`, then classify:

| Classification | Indicators | Recommendation |
|---------------|------------|----------------|
| flaky_test | Known flaky test name, intermittent pattern | retry |
| regression | New test failure on code that was changed | fix_and_retry |
| infra_issue | Docker pull failure, network timeout, runner unavailable | retry |
| timeout | Run exceeded time limit | escalate |

> See `releaser-ref/cross-rite-routing.md` for the full routing table.
