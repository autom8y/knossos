---
name: cartographer-reference
description: "Cartographer agent reference data: ecosystem detection matrices, GoReleaser config parsing tables, and pipeline chain discovery heuristics. Use when: cartographer needs lookup tables during reconnaissance. Triggers: ecosystem detection, goreleaser config, pipeline chain, asset naming."
scope: releaser
---

# Cartographer Reference Data

## Distribution Type Detection

After ecosystem detection, check for distribution type indicators:

| File / Pattern | Result |
|---------------|--------|
| `.goreleaser.yaml` or `.goreleaser.yml` | `distribution_type: binary`, `goreleaser_config: {path}` |
| Neither goreleaser file | `distribution_type: registry`, `goreleaser_config: null` |
| Dockerfile + GHCR/DockerHub in workflow | `distribution_type: container` (record only — escalate; not yet supported) |

When goreleaser is detected, record the config path relative to the repo root. The release-planner uses this to generate binary-appropriate publish commands.

### GoReleaser Config Parsing (binary repos)

When `.goreleaser.yaml` is present, read and extract the following fields for the state map:

| Field | GoReleaser YAML Path | State Map Key |
|-------|---------------------|---------------|
| Project name | `project_name` | `goreleaser_project_name` |
| Target OS list | `builds[].goos` | `goreleaser_goos` |
| Target arch list | `builds[].goarch` | `goreleaser_goarch` |
| Homebrew tap repo | `brews[].repository.{owner,name}` | `goreleaser_brew_tap` (formatted as `owner/name`) |
| GitHub release target | `release.github.{owner,name}` | `goreleaser_release_repo` (formatted as `owner/name`) |
| Homebrew token env var | `brews[].repository.token` | `goreleaser_brew_token_env` (extract env var name, e.g., `HOMEBREW_TAP_TOKEN`) |

Expected asset names follow the pattern `{project_name}_{version}_{os}_{arch}.tar.gz`. Note: GoReleaser's `{{ .Version }}` strips the `v` prefix from tags — use the bare version (e.g., `0.3.0`, not `v0.3.0`) when constructing expected asset names. Record these as `goreleaser_expected_assets` using the cross-product of `goos` x `goarch` (e.g., for darwin+linux x amd64+arm64: four archives plus `checksums.txt`).

Also check `version:` at the top of the goreleaser config — `version: 2` indicates GoReleaser v2 configuration syntax. Record as `goreleaser_config_version`.

If any of these fields are absent, record `null` for the specific key and note the absence — do not fail detection.

### Pipeline Chain: release.yml → e2e-distribution.yml

When scanning a binary repo's CI workflows, detect the two-workflow release chain:

1. **Release trigger workflow** (`release.yml` or similar): triggered by `push: tags: ["v*"]`, runs GoReleaser action (`goreleaser/goreleaser-action`). This is stage 1.
2. **E2E validation workflow** (`e2e-distribution.yml` or similar): triggered by `release: types: [published]` (NOT workflow_run — triggered by the GitHub Release creation event that GoReleaser produces). This is stage 2.

Record this as a `trigger_chain` in `pipeline_chains` with depth 2:
- Stage 1: release.yml, trigger: `push` (tag pattern), classification: `build`
- Stage 2: e2e-distribution.yml, trigger: `release.published`, classification: `deploy`

The e2e workflow may also support `workflow_dispatch` for manual re-runs — this is auxiliary and does not affect chain classification.

### Makefile e2e Target Detection

When `makefile_exists: true`, read the Makefile and scan for e2e-related targets alongside build targets:

| Target Pattern | Semantic |
|---------------|----------|
| `e2e-linux`, `e2e-local`, `e2e-*` | e2e_validation |
| `build`, `compile` | build |

Record e2e targets in `makefile_e2e_targets: [{name: "e2e-linux", semantic: "e2e_validation"}, ...]`. A Makefile with e2e targets is a strong signal that an e2e validation workflow exists — cross-reference with `.github/workflows/` to confirm.

## Pipeline Chain Discovery

After ecosystem detection, scan each release-candidate repo's CI workflow definitions to discover pipeline chains that extend beyond the initial CI run.

### Scan Procedure

1. List workflow files in the repo's CI configuration directory (e.g., `.github/workflows/`)
2. For each workflow file, read its contents and scan for chain indicators
3. For cross-repo dispatches, use `gh api` to read the receiver repo's workflow files:
   ```
   gh api -H "Accept: application/vnd.github.raw+json" repos/{owner}/{repo}/contents/{path-to-workflow-file}
   ```
4. Classify each discovered chain link using the heuristic table below
5. Build the chain graph from trigger source to terminal stage

### Chain Indicator Heuristics

| Pattern Category | Indicators (file content patterns) | Classification |
|-----------------|-----------------------------------|----------------|
| Downstream trigger | `workflow_run`, `workflow_call`, triggered-by references | trigger_chain |
| Cross-repo dispatch | `repository_dispatch`, `workflow_dispatch` with external trigger, dispatch event names | dispatch_chain |
| Deployment stage | deploy, release, publish to infrastructure, task/service update, health check, smoke test, rollout | deployment_chain |
| Attestation/signing | attest, sign, sbom, provenance | deployment_chain (intermediate stage) |
| Auxiliary | scheduled, cron, manual-only triggers with no chain relationship | auxiliary (exclude from chain) |

When uncertain whether a workflow is part of the release chain or auxiliary, include it as a chain link. False positives are preferable to missed deployment stages.

### Cross-Repo Scanning

When a workflow dispatches to another repository:
1. Extract the target repository identifier from the dispatch configuration
2. Use `gh api` to list and read workflow files in the target repo
3. Identify which workflow in the target repo receives the dispatch event
4. Continue scanning the receiver's workflows for further chain links (up to depth 5)
5. Record each cross-repo link with source repo, target repo, and dispatch event name

### Retry Protocol for Cross-Repo Discovery

Cross-repo API calls may fail due to permissions or rate limits:
- Attempt 1: immediate
- Attempt 2: immediate retry
- Attempt 3: immediate retry
- After 3 failures: log warning, set `chain_discovery_status: failed`, continue with remaining repos

### Graceful Degradation

If chain discovery fails for a repo (API errors, permission denied, unparseable workflow files):
- Set `chain_discovery_status: failed` for that repo
- Log the failure reason
- The repo proceeds through the pipeline with flat (chain-unaware) monitoring
- Pipeline-monitor treats repos with `chain_discovery_status: failed` as flat CI monitoring targets
- This is a warning, not a blocking error
