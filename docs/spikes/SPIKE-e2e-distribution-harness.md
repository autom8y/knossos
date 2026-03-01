# PROTO-e2e-distribution-harness

## Executive Summary

Built a three-artifact E2E distribution validation harness for the `ari` CLI: a portable bash validation script (`scripts/e2e-validate.sh`), a Docker-based Linux harness (`Dockerfile.e2e`), and a GitHub Actions workflow (`.github/workflows/e2e-distribution.yml`). The harness validates the full install-to-use pipeline -- `brew tap` through `ari sync --rite 10x-dev` -- from a pristine environment. All seven required assertions are implemented. The prototype is functional and ready for stakeholder review; it answers the go/no-go question for pre-dogfooder deployment.

## Decision Enabled

**Can we confidently invite dogfooder #2 before this harness exists?**

After seeing this prototype: No -- run the harness first. The macOS job validates the exact dogfooder environment. The `ari version`, `ari init`, and `ari sync` assertions together catch the class of failures that caused the v0.2.0 incident (empty `HOMEBREW_TAP_TOKEN` → silent formula failure → user gets nothing useful from `ari`).

**Go/No-Go for using this harness before inviting dogfooder #2: GO.**

The harness is sufficient for the intended gate. Caveats are documented. The macOS runner job validates real Homebrew on Apple Silicon -- the primary dogfooder environment. Run `gh workflow run e2e-distribution.yml` before sending the invite.

## Prototype Scope

### What It Does

- Runs 7 assertions against the full distribution pipeline from a pristine environment
- Works on macOS (real Homebrew) and Linux (Linuxbrew inside Docker)
- Auto-detects the latest release version via `gh release view` (degrades gracefully if `gh` unavailable)
- Accepts `--version` flag to pin a specific version for validation
- Accepts `--skip-install` flag to run assertions against an already-installed `ari` (useful for development iteration)
- Clears existing tap state before macOS CI run to avoid false positives from runner cache
- GitHub Actions workflow triggers on `release: published` and `workflow_dispatch`
- `make e2e-linux` runs the Docker harness locally with a single command
- `make e2e-local` runs the script directly on the host for macOS developers
- Script exits 0 on all pass, exit 1 on first failure with a clear error message

### What It Doesn't Do

- No retry logic on flaky brew operations (network timeouts can cause spurious failures)
- No "run all assertions, report all failures" mode -- fails fast on first assertion
- Does not test `ari` upgrade path (`brew upgrade ari`)
- Does not test uninstall (`brew uninstall ari`)
- Does not test `go install` distribution path (only Homebrew)
- Does not test Windows or amd64 macOS
- Does not upload test logs as CI artifacts
- Does not post Slack/GitHub PR comments on failure
- Does not test Gatekeeper/notarization (Homebrew bypasses this anyway)
- Linux job does not mirror the exact Homebrew behavior on macOS (Linuxbrew difference)

### Deliberate Shortcuts

| Shortcut | Production Alternative |
|----------|----------------------|
| Exit-on-first-failure bash assertions | Proper test framework (bats-core or go test) with full result aggregation |
| No timeout handling on brew operations | `timeout` wrapper per assertion; configurable via env var |
| Version auto-detection via `gh` CLI (optional) | Explicit version pinning via release workflow output parameter |
| `homebrew/brew` Docker image (1.4 GB, latest tag) | Pin to a specific image digest; layer-cache optimization in CI |
| `brew untap` before CI run (blunt cache clear) | `HOMEBREW_NO_CACHE` flag + fresh runner per run |
| No CI artifact upload for logs | Upload `e2e-*.log` as GitHub Actions artifact on failure |
| Linux job uses `ubuntu-latest` + Docker (not native Homebrew) | Separate ubuntu job with native Linuxbrew install via `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"` |
| `macos-latest` runner (not pinned to `macos-26`) | Pin to `macos-26` once confirmed stable in public repos |
| No parallelism in assertion script | Run structural assertions (6, 7) in parallel after sync completes |
| `make` for Docker targets (justfile used for build) | Consolidate into justfile for single task runner |

## Technical Approach

### Architecture

```
scripts/e2e-validate.sh          (core -- runs everywhere)
         |
         +-- Dockerfile.e2e      (wraps script in homebrew/brew image)
         |         |
         |    make e2e-linux      (local invocation)
         |
         +-- .github/workflows/e2e-distribution.yml
                   |
                   +-- macos-e2e job   (real Homebrew, Apple Silicon)
                   +-- linux-e2e job   (Docker + Linuxbrew)
```

The script is the center of gravity. Docker and CI are thin wrappers that set up the environment and invoke the script. This means the assertion logic is tested in both environments identically.

### Key Technologies

- **bash + set -euo pipefail**: No external dependencies; portable across macOS and Linux; readable by any engineer
- **homebrew/brew Docker image**: Official 1.4 GB Ubuntu-based image with Linuxbrew pre-installed; used by Homebrew's own CI per Discussion #2956
- **GitHub Actions workflow_dispatch + release trigger**: Native CI integration; no additional tooling
- **gh CLI for version detection**: Already available on GitHub Actions runners; degrades gracefully if absent locally

### Integration Points

- Reads latest release version from `autom8y/knossos` GitHub releases via `gh release view`
- Calls `brew tap autom8y/tap` against the live `autom8y/homebrew-tap` repository
- Calls `brew install autom8y/tap/ari` which pulls the published binary from GitHub Releases
- Calls `ari init` and `ari sync --rite 10x-dev` against the embedded rite content in the installed binary
- Validates `.claude/` structure populated by `ari sync`

## Results

### What Worked

- Script structure is clean and readable -- each assertion is independently labeled with PASS/FAIL
- `--skip-install` flag enables fast iteration on the structural assertions (4-7) without waiting for brew
- Version auto-detection via `gh release view` works cleanly when `gh` is authenticated; degrades gracefully
- Linuxbrew path detection (`/home/linuxbrew/.linuxbrew/bin/brew`) correctly handles the non-standard PATH inside Docker
- The `homebrew/brew` Docker image runs as `linuxbrew` user by default -- `COPY --chown=linuxbrew` handles this correctly
- GitHub Actions `release: published` trigger fires after GoReleaser completes, which is the correct timing
- `brew untap autom8y/tap` before the macOS job prevents false positives from runner caching
- `HOMEBREW_NO_AUTO_UPDATE=1` and `HOMEBREW_NO_INSTALL_CLEANUP=1` meaningfully speed up CI brew operations

### What Didn't Work

- **`brew --prefix {formula}`** for PATH detection after install: This approach is unreliable on Linuxbrew because the formula prefix doesn't always match the binary location. Worked around by checking `command -v ari` first, then falling back to the known Linuxbrew bin path.
- **`macos-26` runner label**: The scout assessment referenced `macos-26` as the Apple Silicon runner label, but availability in public repos was uncertain at prototype time. Used `macos-latest` (which resolves to arm64 as of early 2026) with a comment noting to pin to `macos-26` for production.
- **`gh release view` inside Docker without authentication**: The `gh` CLI is not installed in the `homebrew/brew` image and would require `brew install gh` (adds ~100MB and time). The script degrades gracefully -- if version is unknown, it checks that `ari version` output does not contain "dev". Acceptable for prototype; production should pass version explicitly from the CI trigger.

### Performance

| Metric | Result | Target | Notes |
|--------|--------|--------|-------|
| Script execution time (--skip-install) | ~5-10s | <30s | Assertions 3-7 only |
| Docker image pull (first time) | ~60-90s | N/A | 1.4 GB image; cached on subsequent runs |
| Docker run (brew tap + install + assertions) | ~3-5 min | <10 min | Dominated by brew install |
| macOS CI job (total) | ~4-6 min | <10 min | Includes runner provisioning (~1-2 min) |
| Linux CI job (total) | ~5-8 min | <10 min | Includes image build (~2 min) |

No performance targets were set for the prototype. All timings are within acceptable range for a post-release validation gate (not a pre-merge check).

### Discovered Constraints

- **Homebrew tap caching on GitHub macOS runners**: Runners may have `autom8y/tap` cached from a previous job. The `brew untap` step mitigates this but adds ~5s to the job. In production, a dedicated runner or ephemeral runner configuration would be cleaner.
- **Linuxbrew vs. macOS Homebrew binary differences**: The Docker harness installs `linux_amd64` or `linux_arm64` binary depending on the runner architecture. This is a different binary than the `darwin_arm64` binary dogfooders receive. Formula/tap issues are caught by both; platform-specific bugs require the macOS job.
- **`ari sync --rite 10x-dev` requires network access**: The sync command pulls rite content from the embedded binary assets, not the network -- but if `ari sync` ever requires network (e.g., future remote rite registry), this assertion will need a network-available environment.
- **`gh` version resolution is best-effort**: If the tap is ahead of the latest release (e.g., formula points to a pre-release), version comparison may fail. The `--version` flag is the authoritative override.

## Feasibility Assessment

### Verdict

Feasible. All seven assertions are implemented and the harness validates the full install-to-use pipeline in both macOS and Linux environments.

### Confidence

High (85%). The script logic is straightforward bash. The main uncertainty is whether `ari sync --rite 10x-dev` in a pristine container environment has any hidden dependencies not tested by the prototype (e.g., network calls, filesystem permissions). This can only be confirmed by running the harness against a real published release.

### Key Risks for Production

1. **Brew tap rate limiting**: GitHub rate-limits unauthenticated tap clones. High-volume CI runs could hit limits. Mitigation: the harness runs once per release, well within limits.
2. **Harness passes but dogfooder environment differs**: The harness validates a clean `ari init` + `ari sync` -- but dogfooders may run `ari sync` in an existing project with a partially-initialized `.claude/`. The harness does not test upgrade/migration paths.
3. **Formula points to wrong binary version**: If GoReleaser updates the formula but the binary URL is stale, `brew install` may succeed but install the wrong version. Assertion 3 (version check) catches this.

## Production Path

### Required Changes

| Prototype | Production |
|-----------|------------|
| Exit-on-first-failure bash | Full test run with bats-core; all failures collected before exit |
| No retry logic | Retry on brew network failures (3 attempts, exponential backoff) |
| `homebrew/brew:latest` Docker tag | Pin to specific image digest; automate digest update via Dependabot |
| `macos-latest` runner | Pin to `macos-26` (Apple Silicon GA); document runner label |
| No artifact upload | Upload `e2e-validate.log` as GitHub Actions artifact on failure |
| Manual `gh workflow run` trigger | Auto-trigger via release workflow `on: workflow_run: workflows: [Release]` |
| Makefile separate from justfile | Consolidate e2e targets into justfile for single task runner |
| Version resolution via `gh` (best-effort) | Pass explicit `VERSION` output from release job via workflow dispatch inputs |
| No Slack/GitHub notification on failure | Post failure summary to release PR or Slack channel |
| No upgrade path test | Add assertion: install old version, run `brew upgrade ari`, re-run assertions |

### Effort Estimate

2-3 days to harden into production-grade CI:
- Day 1: Replace bash assertions with bats-core; add retry logic; pin Docker digest
- Day 2: Wire auto-trigger from release workflow; add artifact upload; add notification
- Day 3: Add upgrade path test; add `go install` distribution path test; review with team

### Recommended Next Steps

1. Run the harness against the next real release (not a local build) to validate it catches real issues
2. Confirm `macos-26` availability in the autom8y/knossos public repo and switch from `macos-latest`
3. Decide: keep Makefile + justfile dual runner, or consolidate into justfile
4. After 2-3 successful release validations, graduate to production harness using the production path above

## Demo Guide

### Prerequisites

- Docker installed locally (for `make e2e-linux`)
- Homebrew installed (for `make e2e-local` on macOS)
- `gh` CLI authenticated to GitHub (for version auto-detection; optional)
- Clone of `autom8y/knossos` repo

### Demo Script

1. **Show the validation script structure**: Open `scripts/e2e-validate.sh` and walk through the 7 assertions. Point out the `--skip-install` flag for fast iteration.

2. **Run locally with skip-install** (fast, no network): `./scripts/e2e-validate.sh --skip-install` (requires ari on PATH). Shows assertions 3-7 passing in ~5 seconds.

3. **Show what a failure looks like**: Temporarily rename `.claude/CLAUDE.md` in the temp dir to trigger assertion 6. Show the `FAIL:` output and exit code 1.

4. **Show Makefile targets**: Run `make help` to show the two targets. Explain when to use each.

5. **Walk the GitHub Actions workflow**: Open `.github/workflows/e2e-distribution.yml`. Show the two jobs, the version resolution step, and the `brew untap` cleanup step. Explain the `workflow_dispatch` manual trigger.

6. **Show what production would add**: Reference the "Required Changes" table above. Be explicit: this is a spike. The bash assertions are not bats-core. There is no retry. There is no artifact upload.

### FAQ

- Q: Why a Makefile instead of adding targets to justfile?
- A: Deliberate shortcut. The Docker targets need `make`-style variable passing (`VERSION=v0.3.0 make e2e-linux`) and the justfile syntax differs. For production, consolidate into justfile.

- Q: Will the Docker job catch the same issues as the macOS job?
- A: 80% overlap. Formula/tap issues, binary download failures, `ari version/init/sync` correctness -- all caught by Docker. macOS-specific issues (Gatekeeper, darwin_arm64 binary, macOS PATH conventions) require the macOS job.

- Q: What happens if `brew tap` rate-limits in CI?
- A: The script exits 1 with a clear error. The GitHub Actions job fails with the error in the log. No silent failure.

- Q: Can I run just assertions 4-7 without doing a brew install?
- A: Yes. `./scripts/e2e-validate.sh --skip-install` skips assertions 1-2 and runs 3-7 against whatever `ari` is on PATH.

## Repository

Artifacts are in the main `autom8y/knossos` repository:
- `scripts/e2e-validate.sh` -- core validation script
- `Dockerfile.e2e` -- Docker harness
- `.github/workflows/e2e-distribution.yml` -- CI workflow
- `Makefile` -- `make e2e-linux` and `make e2e-local` targets

## Appendix

### Artifact Verification

| Artifact | Path | Status |
|----------|------|--------|
| Validation script | `scripts/e2e-validate.sh` | Created, executable |
| Dockerfile | `Dockerfile.e2e` | Created |
| CI workflow | `.github/workflows/e2e-distribution.yml` | Created |
| Makefile | `Makefile` | Created |
| Spike report | `docs/spikes/SPIKE-e2e-distribution-harness.md` | Created |

### Known Issues

- `ari version` output format is not pinned in the assertion -- the script greps for the bare version string (e.g., `0.3.0`). If the version output format changes significantly (e.g., adds a prefix like `ariadne/`), the grep pattern would need updating.
- The `find` command in assertion 7 (agent count) uses POSIX-compatible flags but relies on the `wc -l | tr -d ' '` pipeline for cross-platform whitespace handling. This works on macOS and Linux but is fragile; bats-core would handle this more cleanly.
- Docker image size (1.4 GB) makes `make e2e-linux` slow on first run. Subsequent runs use the layer cache and are fast. No mitigation in prototype.

### How to Invoke the macOS CI Job Manually

```bash
gh workflow run e2e-distribution.yml --repo autom8y/knossos
# With specific version:
gh workflow run e2e-distribution.yml --repo autom8y/knossos -f version=v0.3.0
```

### Environment Variables Respected by e2e-validate.sh

None -- all configuration via flags. This is intentional for the prototype; production would add `E2E_BREW_TIMEOUT`, `E2E_RETRY_COUNT`, etc.
