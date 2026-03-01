# SCOUT-e2e-distribution-harness

## Executive Summary

This assessment evaluates four approaches for building a clean-room E2E validation harness that proves the `ari` CLI install-to-use pipeline works from a pristine environment. The recommended approach is a **layered strategy**: Docker (Linuxbrew) as the primary fast-feedback harness with a GitHub Actions macOS runner as the high-fidelity validation gate. Docker provides sub-minute reproducible validation for development iteration at zero cost. The GitHub Actions macOS runner provides real-environment fidelity on every release at zero cost (public repo). Lima and Codespaces are both inferior tradeoffs for this use case. Verdict: **Adopt** the layered Docker + GitHub Actions approach.

## Technology Overview

- **Category**: Test Infrastructure / Distribution Validation
- **Maturity**: Mainstream (all four options use production-grade tooling)
- **License**: All options use MIT/Apache-licensed tooling
- **Backing**: Docker Inc, GitHub/Microsoft, Lima (CNCF), Homebrew (community)

## Context

The v0.2.0 release exposed a gap predicted by the distribution audit (GAP-D01): the Homebrew formula delivery failed due to an empty `HOMEBREW_TAP_TOKEN`. The release was recovered, but the incident demonstrated that no automated validation exists for the distribution pipeline. Before inviting external dogfooders, the team needs confidence that `brew tap autom8y/tap && brew install ari` through `ari sync --rite 10x-dev` works end-to-end.

The harness must validate five steps from a pristine environment (no Go toolchain, no `KNOSSOS_HOME`, no prior `ari` install):

1. `brew tap autom8y/tap && brew install ari` succeeds
2. `ari version` shows expected version
3. `ari init` works in a fresh directory
4. `ari sync --rite 10x-dev` materializes `.claude/` correctly
5. `.claude/` structure contains expected files (agents/, commands/, skills/, settings.json, CLAUDE.md)

## Option Analysis

### Option 1: Docker (Linux Container with Linuxbrew)

**How it works**: Use the official `homebrew/brew` Docker image (Ubuntu-based, 1.4 GB, includes Linuxbrew at `/home/linuxbrew/.linuxbrew/`) to run the full install-to-use pipeline in an isolated container.

**Capabilities**:
- Sub-minute startup, seconds for cached runs
- Deterministic -- same image, same result every time
- No local state contamination between runs
- CI-native -- runs on any Docker-capable runner (ubuntu-latest)
- Single command: `docker run --rm homebrew/brew bash -c "brew tap autom8y/tap && brew install ari && ari version && ..."`
- Free in CI (ubuntu-latest runners, public repo)

**Limitations**:
- Tests Linuxbrew, not macOS Homebrew -- different binary (linux_amd64/arm64), different paths
- Cannot validate macOS-specific behavior (Gatekeeper, notarization, macOS PATH conventions)
- homebrew/brew image is 1.4 GB (one-time pull)
- Homebrew inside Docker runs as non-root user (requires `USER linuxbrew` or custom setup)

### Option 2: Lima/Colima (macOS VM on macOS)

**How it works**: Lima launches a Linux VM on macOS with file sharing and port forwarding. Cannot run macOS guests (Apple licensing restriction on non-Apple hardware in CI). Would only provide Linux validation, same as Docker but heavier.

**Capabilities**:
- Lima v2.0 (Dec 2025) supports krunkit VM driver with GPU acceleration
- Actively maintained (CNCF project, commits as of Feb 2026)
- Good for local Linux development environments

**Limitations**:
- Cannot run macOS guests -- this is the critical disqualifier for macOS Homebrew testing
- Significantly heavier than Docker (full VM vs container)
- Not CI-friendly without self-hosted macOS runners
- Provides no fidelity advantage over Docker for Linux testing
- Requires `brew install lima` on the host (chicken-and-egg for clean-room testing)
- Setup complexity: VM definition YAML, boot time (30-60s), teardown

### Option 3: GitHub Codespaces

**How it works**: Spin up a cloud-hosted Linux development environment with Linuxbrew available. Run the E2E harness interactively or via devcontainer automation.

**Capabilities**:
- Zero local infrastructure -- runs entirely in the cloud
- Pre-configured dev environments via devcontainer.json
- Good for ad-hoc developer validation ("does this work on a clean machine?")
- Can install Linuxbrew in the devcontainer

**Limitations**:
- Linux only -- same fidelity as Docker, higher cost
- Billed per hour ($0.18/hr for 2-core, $0.36/hr for 4-core)
- Cannot trigger automatically on release (no native CI integration)
- Requires manual invocation or a secondary workflow to create/destroy codespaces
- Startup time 30-120 seconds (vs seconds for cached Docker)
- Overkill for a scripted validation -- Codespaces provides a full IDE when we need a shell script

### Option 4: GitHub Actions macOS Runner

**How it works**: GitHub-hosted macOS runner (Apple Silicon, arm64) with Homebrew pre-installed. Runs the full E2E pipeline in the exact target environment.

**Capabilities**:
- Real macOS Homebrew -- the actual environment dogfooders will use
- Homebrew pre-installed on runner (no setup required)
- Free for public repositories with standard runners (macos-26, arm64)
- Can trigger on release events (`on: release: types: [published]`)
- Apple Silicon runners available (macos-26 label, arm64)
- Validates darwin_arm64 binary -- the primary distribution target
- Full GitHub Actions ecosystem (caching, artifacts, notifications)

**Limitations**:
- Slower startup than Docker (VM provisioning: 1-3 minutes)
- macOS runners are slower overall than Linux runners
- Cannot test Linux installation path (need separate Linux job)
- Rate limited -- cannot run hundreds of times per day during development
- Not ideal for rapid iteration during harness development

## Comparison Matrix

| Criteria | Docker (Linuxbrew) | Lima/Colima | GitHub Codespaces | GH Actions macOS |
|----------|-------------------|-------------|-------------------|------------------|
| **Fidelity** | MEDIUM -- Linux binary, not macOS | LOW -- Linux VM, no advantage over Docker | MEDIUM -- Linux, same as Docker | HIGH -- real macOS + Homebrew |
| **Reproducibility** | HIGH -- deterministic container | MEDIUM -- VM state can drift | MEDIUM -- devcontainer helps | HIGH -- fresh runner every time |
| **CI Integration** | HIGH -- any ubuntu runner | LOW -- needs macOS host | LOW -- manual or secondary workflow | HIGH -- native workflow trigger |
| **Setup Cost** | LOW -- Dockerfile + script | HIGH -- VM config + Lima install | MEDIUM -- devcontainer.json | LOW -- workflow YAML |
| **Runtime Cost** | FREE (public repo, ubuntu runner) | FREE (local) but not CI-usable | $0.18-0.36/hr per run | FREE (public repo, standard runner) |
| **Single Command** | YES -- `docker run ...` or `make e2e` | NO -- multi-step VM lifecycle | NO -- web UI or gh codespace create | YES -- `gh workflow run` |
| **Iteration Speed** | FAST -- seconds for cached runs | SLOW -- 30-60s boot | SLOW -- 30-120s startup | SLOW -- 1-3 min provisioning |
| **macOS Validation** | NO | NO (cannot run macOS guests) | NO | YES |

**Scoring** (1-5, 5 is best):

| Criteria (Weight) | Docker | Lima | Codespaces | GH Actions macOS |
|-------------------|--------|------|------------|------------------|
| Fidelity (30%) | 3 | 2 | 3 | 5 |
| Reproducibility (20%) | 5 | 3 | 3 | 5 |
| CI Integration (20%) | 5 | 1 | 2 | 5 |
| Setup Cost (15%) | 5 | 2 | 3 | 5 |
| Runtime Cost (10%) | 5 | 5 | 2 | 5 |
| Single Command (5%) | 5 | 1 | 2 | 4 |
| **Weighted Total** | **4.30** | **2.20** | **2.70** | **5.00** |

## Ecosystem Assessment

- **Community**: All four technologies have large, active communities. Docker and GitHub Actions are industry standards. Lima is a CNCF project with active maintenance. Homebrew has 40k+ GitHub stars.
- **Documentation**: Docker and GitHub Actions have extensive official documentation. The `homebrew/brew` Docker image has a dedicated Homebrew Discussion (#2956) documenting its intended use for CI/CD testing.
- **Tooling**: GitHub Actions provides native workflow authoring, run visualization, artifact storage, and caching. Docker provides BuildKit caching, multi-stage builds, and layer optimization.
- **Adoption**: Homebrew formula testing in Docker is used by Homebrew's own CI. GitHub Actions macOS runners are used by major projects (Homebrew itself, Swift, Xcode-dependent projects).

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Linuxbrew passes but macOS Homebrew fails (binary arch, path differences) | MEDIUM | HIGH | Layered approach: Docker catches formula/tap issues, macOS runner catches platform-specific issues |
| macOS runner Homebrew cache causes false positives (tap already cached) | LOW | MEDIUM | Use `brew untap autom8y/tap` before test, or use `--no-cache` flag |
| homebrew/brew Docker image version drift (Linuxbrew version != dogfooder Homebrew version) | LOW | LOW | Pin Docker image tag; formula syntax is cross-platform |
| GitHub Actions macOS runner availability/flakiness | LOW | LOW | macOS-26 is GA as of Feb 2026; fallback to Docker-only validation |
| Harness passes but `ari init` fails due to missing embedded assets in release binary | MEDIUM | HIGH | Harness must test `ari init` and `ari sync`, not just `ari version` |
| Rate limiting on `brew tap` from CI (GitHub API limits for tap clone) | LOW | LOW | Standard GitHub rate limits apply; one run per release is well within limits |

## Fit Assessment

- **Philosophy Alignment**: Strong. The layered approach follows the Knossos principle of pragmatic testing -- fast feedback loop for development, high-fidelity gate for releases. No over-engineering.
- **Stack Compatibility**: Excellent. The existing release workflow (`.github/workflows/release.yml`) already uses GitHub Actions. Adding a post-release validation workflow is a natural extension. Docker requires no new infrastructure.
- **Team Readiness**: High. The codebase already has GitHub Actions workflows (`ariadne-tests.yml`, `release.yml`, `verify-formal-specs.yml`). Docker is a standard tool. No new technology to learn.

## Recommendation

**Verdict**: Adopt (layered approach)

**Primary**: GitHub Actions macOS runner -- post-release E2E workflow triggered on `release: published`. This is the highest-fidelity option, runs in the exact dogfooder environment, is free for public repos, and integrates natively with the existing release pipeline.

**Secondary**: Docker (Linuxbrew) -- developer-invocable local harness (`make e2e-linux` or `./scripts/e2e-validate.sh`). This provides fast iteration during harness development and catches formula/tap issues without waiting for a release cycle.

**Eliminated**:
- **Lima/Colima**: Provides no advantage over Docker for Linux testing and cannot test macOS. Strictly dominated.
- **GitHub Codespaces**: Higher cost, lower automation, same fidelity as Docker. Overkill for a scripted validation.

**Rationale**:

1. The primary target is macOS Homebrew. Only the GitHub Actions macOS runner tests this directly.
2. Public repos get free standard macOS runner minutes -- no cost concern.
3. Docker provides the fast inner loop for harness development and catches 80% of issues (tap configuration, formula syntax, binary download, `ari version/init/sync` on Linux).
4. The layered approach means Docker catches issues fast during development; the macOS runner catches platform-specific issues on release.
5. Both options have trivial setup cost (Dockerfile + workflow YAML) and high reproducibility.

**Confidence**: High (85%). The only uncertainty is whether macOS-specific issues exist that neither option would catch during development (e.g., Gatekeeper quarantine on direct download -- but Homebrew bypasses this).

**Next Steps**:
1. Prototype Engineer builds the Docker harness first (fast iteration, validates harness logic)
2. Prototype Engineer adds the GitHub Actions macOS workflow (validates real environment)
3. Both are wired: Docker runs locally via `make e2e-linux`, macOS workflow triggers on release
4. Optionally: macOS workflow can also be triggered via `workflow_dispatch` for ad-hoc validation

## Complexity Estimate for Prototype Phase

| Component | Effort | Deliverables |
|-----------|--------|-------------|
| Docker harness | 1-2 hours | Dockerfile, `scripts/e2e-validate.sh`, Makefile target |
| GitHub Actions macOS workflow | 1-2 hours | `.github/workflows/e2e-distribution.yml` |
| Integration testing | 30 min | Run both, verify they catch real issues |
| Documentation | 30 min | Update spike with results, add `make e2e-linux` to dev guide |
| **Total** | **3-5 hours** | Time-boxed spike, not production CI infrastructure |

### Implementation Notes for Prototype Engineer

**Docker harness structure**:
```
scripts/e2e-validate.sh          # Standalone script, runs inside container
Dockerfile.e2e                   # Based on homebrew/brew, adds test script
Makefile (target: e2e-linux)     # One-command invocation
```

**GitHub Actions workflow structure**:
```yaml
name: E2E Distribution Validation
on:
  release:
    types: [published]
  workflow_dispatch:             # Manual trigger for ad-hoc validation
jobs:
  e2e-macos:
    runs-on: macos-26           # Apple Silicon, arm64
    steps:
      - brew tap autom8y/tap
      - brew install ari
      - ari version              # Assert contains expected version
      - mkdir /tmp/e2e-test && cd /tmp/e2e-test
      - ari init
      - ari sync --rite 10x-dev
      - ls .claude/agents/ .claude/commands/ .claude/skills/ .claude/settings.json .claude/CLAUDE.md
  e2e-linux:
    runs-on: ubuntu-latest
    steps:
      # Install Linuxbrew, then same validation steps
```

**Key assertions the harness must make**:
1. `brew install` exit code 0
2. `ari version` output contains the release version string (not "dev")
3. `ari init` exit code 0, creates expected directory structure
4. `ari sync --rite 10x-dev` exit code 0
5. `.claude/` contains: `agents/`, `commands/`, `skills/`, `settings.json`, `CLAUDE.md`
6. At least one agent file exists in `.claude/agents/`
7. `CLAUDE.md` contains expected section markers (e.g., `KNOSSOS:START`)

---

## References

- [Homebrew/brew Docker image](https://hub.docker.com/r/homebrew/brew/) -- Official 1.4 GB Ubuntu-based image with Linuxbrew
- [Homebrew Docker Discussion #2956](https://github.com/orgs/Homebrew/discussions/2956) -- Intended use for CI/development
- [macOS-26 GA for GitHub Actions](https://github.blog/changelog/2026-02-26-macos-26-is-now-generally-available-for-github-hosted-runners/) -- Apple Silicon runners available
- [GitHub Actions Runner Pricing](https://docs.github.com/en/billing/reference/actions-runner-pricing) -- Standard macOS runners free for public repos
- [GitHub Actions 2026 Pricing Changes](https://resources.github.com/actions/2026-pricing-changes-for-github-actions/) -- Up to 39% price reduction, free tier unchanged
- [Lima v2.0](https://www.cncf.io/blog/2025/12/11/lima-v2-0-new-features-for-secure-ai-workflows/) -- CNCF project, Linux VMs on macOS
- [Homebrew Releaser Action](https://github.com/marketplace/actions/homebrew-releaser) -- GitHub Action for Homebrew formula automation
- `docs/spikes/SPIKE-distribution-audit-gap-report.md` -- Prior audit predicting GAP-D01 (Homebrew failure)
- `docs/spikes/SPIKE-v020-release-recovery.md` -- v0.2.0 recovery analysis
- `.goreleaser.yaml` -- Release configuration (darwin + linux, amd64 + arm64)
- `.github/workflows/release.yml` -- Current release workflow
