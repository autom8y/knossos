# SPIKE: Distribution Initiative -- Comprehensive Audit & Gap Report

> First-principles audit of the Knossos/Ariadne CLI distribution initiative, measured against literature best practices and current codebase state.

**Date**: 2026-03-01
**Author**: Spike (distribution audit)
**Literature Reference**: `.know/literature-developer-cli-distribution-dogfooding-best-practices.md`
**Prior Art**: `docs/spikes/SPIKE-cli-distribution.md`, `docs/decisions/ADR-cli-distribution.md`, `docs/decisions/TDD-single-binary-completion.md`

---

## Executive Summary

The Knossos distribution initiative is **architecturally sound but operationally stalled**. The foundational decisions (GoReleaser, Homebrew tap, single-binary embedding, `ari init`) are implemented and align with literature consensus. However, the v0.1.0 release (2026-01-07) is nearly two months old with 305 unreleased commits, the Homebrew formula was never generated (empty Formula/ directory), no install script exists, the dogfooding infrastructure has zero formal measurement, and several planned Phase 2/3 features remain unstarted. The binary is functionally distributed to exactly one user (the author) via manual `go build`.

**Severity**: The distribution infrastructure is an untested pipeline. The release mechanism has fired once and the primary distribution channel (Homebrew) never delivered a formula. This represents a **release confidence gap** -- the team cannot currently demonstrate that the end-to-end distribution pipeline works.

---

## 1. Audit Framework

### Questions This Audit Answers

1. What does the literature say about distribution best practices, and how does Knossos align?
2. What was planned (spike + ADR + TDD), what was implemented, and what remains?
3. Where are the gaps between current state and first-principles best practice?
4. What is the dogfooding posture, and how does it compare to literature recommendations?
5. What should be done next, in priority order?

### Evidence Sources

| Source | Type | Files Examined |
|--------|------|----------------|
| Literature review | External scholarship (13 sources) | `.know/literature-developer-cli-distribution-dogfooding-best-practices.md` |
| Prior spike | Internal research | `docs/spikes/SPIKE-cli-distribution.md` |
| ADR | Architecture decision | `docs/decisions/ADR-cli-distribution.md` |
| TDD | Technical design | `docs/decisions/TDD-single-binary-completion.md` |
| GoReleaser config | Implementation | `.goreleaser.yaml` |
| Release workflow | Implementation | `.github/workflows/release.yml` |
| Test workflow | Implementation | `.github/workflows/ariadne-tests.yml` |
| GitHub release | Artifact | `gh release view v0.1.0` |
| Homebrew tap | External repo | `~/Code/homebrew-tap/` |
| Binary | Artifact | 34MB dev build, `ari version` |
| Init command | Implementation | `internal/cmd/initialize/init.go` |
| Embed system | Implementation | `embed.go` |
| Codebase knowledge | Internal docs | `.know/architecture.md`, `.know/design-constraints.md`, `.know/conventions.md`, `.know/scar-tissue.md` |

---

## 2. Literature Alignment Matrix

Cross-referencing the 5 thematic findings from the literature review against current Knossos state.

### Theme 1: Single-Binary Distribution [STRONG consensus]

| Best Practice | Knossos Status | Gap |
|---------------|---------------|-----|
| Self-contained executable, no runtime dependency | IMPLEMENTED | None -- `CGO_ENABLED=0` produces static binary |
| Embedded assets for standalone operation | IMPLEMENTED | `embed.go` embeds rites, templates, agents, mena, hooks |
| Cross-platform builds (darwin/linux, amd64/arm64) | CONFIGURED | GoReleaser config covers darwin + linux; Windows builds configured in archive overrides but NOT in `goos:` list |
| Binary size manageable (<10MB for CLI tools) | OBSERVED | 34MB dev build. Release binary ~5MB (v0.1.0 arm64 was 4.5MB). Dev build inflated by debug symbols; `ldflags: -s -w` strips them at release time. Acceptable. |
| `go install` as documented fallback | DOCUMENTED | In release notes. Caveat: shows "dev" version. Documented in spike as "not promoted". |

**Assessment**: STRONG alignment. The single-binary story is the most complete part of the distribution initiative.

### Theme 2: Structured Dogfooding [MODERATE consensus]

| Best Practice | Knossos Status | Gap |
|---------------|---------------|-----|
| Dedicated metrics infrastructure for dogfooding signal | NOT IMPLEMENTED | No CLI analytics, no usage telemetry, no dashboards |
| Diverse participant population | NOT APPLICABLE | Single developer; no external dogfood testers |
| Dogfooding feedback treated with production priority | INFORMAL | Scar tissue (`SCAR-001` through `SCAR-016`) demonstrates dogfooding catches real bugs, but captured via ad hoc git commits, not structured feedback |
| Formal feedback channels | NOT IMPLEMENTED | No issue templates, no structured feedback mechanism |
| Dogfooding as complement to formal testing | PARTIAL | Test suite exists (121+ test files), but no integration testing of the distribution pipeline itself |

**Assessment**: WEAK alignment. Dogfooding happens organically but without any of the structure the literature recommends. This is expected for a single-developer project (literature gap acknowledged in SRC knowledge), but represents a scaling risk.

### Theme 3: Feedback Loop Optimization [MODERATE consensus]

| Best Practice | Knossos Status | Gap |
|---------------|---------------|-----|
| Installation takes <1 command, <2 minutes | NOT ACHIEVED | Homebrew formula never delivered; current install: `go build ./cmd/ari` requires Go toolchain |
| Update mechanism (auto-update or version check) | NOT IMPLEMENTED | No `ari update` command, no version check warning |
| Measure actual install/update times | NOT MEASURED | No data collected |
| Auto-update or minimum-version enforcement | NOT IMPLEMENTED | Mentioned in spike as future enhancement |

**Assessment**: GAP. The installed binary (`which ari`) shows `ari dev (none, unknown)` -- even the author is running an unversioned dev build rather than a released version. The feedback loop between "code change" and "binary in PATH" relies on manual `go build && cp`.

### Theme 4: Internal Tools as Products [MODERATE consensus]

| Best Practice | Knossos Status | Gap |
|---------------|---------------|-----|
| Dedicated ownership | YES | Single developer, dedicated focus |
| Product management rigor (roadmap, user research) | PARTIAL | MEMORY.md tracks priorities; no formal roadmap or user research |
| Built-in health checks (`doctor` command) | NOT IMPLEMENTED | No `ari doctor` or health check command |
| Helpful error messages | IMPLEMENTED | Structured error system with 20+ codes, human-readable messages |
| Onboarding documentation | STALE | `docs/guides/ariadne-cli.md` references `just build` (no Justfile exists), outdated command list |
| Usage analytics with opt-out | NOT IMPLEMENTED | No telemetry |

**Assessment**: PARTIAL alignment. The error system and code quality are production-grade, but the product surface (documentation, health checks, analytics) is underdeveloped.

### Theme 5: Distribution Strategy (Fragmented consensus)

| Best Practice | Knossos Status | Gap |
|---------------|---------------|-----|
| Follow clig.dev and 12 Factor CLI for interaction design | MOSTLY ALIGNED | `--help`, `--version`, `-o json/yaml`, `-v` verbose mode all present. Missing: shell completions (Cobra provides it but not documented), `--no-color`/color handling |
| Multi-channel distribution (Homebrew primary, direct download fallback) | CONFIGURED BUT BROKEN | GoReleaser + Homebrew configured; Formula/ directory empty; Homebrew install fails |
| GoReleaser automation | IMPLEMENTED | Working -- produced v0.1.0 release with 4 platform binaries + checksums |
| `install.sh` curl-pipe-sh fallback | NOT IMPLEMENTED | Recommended in spike, never created |
| Scoop/Windows support | NOT IMPLEMENTED | Planned as Phase 2; `.goreleaser.yaml` has `windows` in format_overrides but not in `goos:` build matrix |
| Code signing / notarization | NOT IMPLEMENTED | Planned as Phase 3 |
| GPG signing of releases | NOT IMPLEMENTED | GoReleaser supports it; not configured |

**Assessment**: PARTIALLY IMPLEMENTED. The automation layer works but the delivery channels are incomplete. The Homebrew tap -- the primary recommended channel -- is non-functional.

---

## 3. Implementation Status: Planned vs. Actual

### From SPIKE-cli-distribution.md (2026-01-07)

| Item | Planned | Actual | Status |
|------|---------|--------|--------|
| GoReleaser configuration | Phase 1 | `.goreleaser.yaml` exists, v2 format | DONE |
| GitHub Actions release workflow | Phase 1 | `.github/workflows/release.yml` exists | DONE |
| Homebrew tap repository | Phase 1 | `autom8y/homebrew-tap` exists with README | PARTIAL -- Formula/ empty |
| PAT for cross-repo publishing | Phase 1 | `HOMEBREW_TAP_TOKEN` referenced in workflow | UNCLEAR -- formula was never pushed |
| Tag v1.0.0 | Phase 1 | Tagged v0.1.0 | DONE (different version) |
| Scoop bucket | Phase 2 | Not started | NOT DONE |
| deb/rpm packages | Phase 2 | Not started | NOT DONE |
| install.sh script | Phase 2 | Not started | NOT DONE |
| Nix NUR | Phase 3 | Not started | NOT DONE |
| Code signing | Phase 3 | Not started | NOT DONE |
| APT/YUM repository | Phase 3 | Not started | NOT DONE |

### From TDD-single-binary-completion.md (2026-02-06)

| Task | Planned | Actual | Status |
|------|---------|--------|--------|
| Task 1: Rite Embedding (`embed.go`) | Create root package with `//go:embed` | `embed.go` exists with 5 embedded assets | DONE |
| Task 1: SourceResolver embedded tier | Add SourceEmbedded tier 5 | Implemented (embedded FS wired through materializer) | DONE |
| Task 1: Materializer fs.FS support | Refactor to support `fs.FS` | Implemented (WithEmbeddedFS, WithEmbeddedTemplates, etc.) | DONE |
| Task 1: Inscription Generator dual-path | Add TemplateFS field | Implemented | DONE |
| Task 1: Hooks config embedded fallback | Embedded hooks.yaml as lowest-priority fallback | Implemented | DONE |
| Task 1: Wiring in main.go | Import root knossos package, wire embedded assets | Implemented (`common.SetEmbeddedAssets`, `SetEmbeddedUserAssets`) | DONE |
| Task 2: `ari init` command | Create `/internal/cmd/initialize/init.go` | Implemented with tests (5 passing) | DONE |
| Task 2: XDG mena extraction | Extract embedded mena to XDG data dir | Implemented (`extractEmbeddedMenaToXDG`) | DONE (beyond original spec) |
| Task 3: USE_ARI_HOOKS removal | Remove from 4 shell scripts | Not verified in this audit | UNCLEAR |

### From ADR-cli-distribution.md

| Checklist Item | Status |
|----------------|--------|
| `.goreleaser.yaml` created and validated | DONE (not validated -- `goreleaser` not installed locally) |
| `homebrew-tap` repository created with README | DONE (README exists, Formula/ empty) |
| GitHub Actions workflow created | DONE |
| PAT generated and stored as secret | PARTIALLY DONE (release fired, but Homebrew formula missing suggests token issue) |
| Test release with `--snapshot` flag | UNKNOWN (no evidence of snapshot testing) |
| Documentation updated with installation instructions | NOT DONE (docs/guides/ariadne-cli.md still says `just build`) |
| First tagged release published | DONE (v0.1.0) |

---

## 4. Gap Inventory

### GAP-D01: Homebrew Formula Never Generated [CRITICAL]

**Evidence**: `~/Code/homebrew-tap/Formula/` directory is empty. `brew info autom8y/tap/ari` returns "No available formula." The v0.1.0 release has 0 downloads for darwin_amd64 and 1 download for darwin_arm64 (likely the author).

**Root Cause Hypothesis**: The `HOMEBREW_TAP_TOKEN` secret may not have sufficient permissions, or the GoReleaser Homebrew configuration may have a subtle error. The release ran successfully (produced binaries) but the Homebrew formula push silently failed or was never triggered.

**Impact**: The primary distribution channel is non-functional. Anyone following the documented `brew install autom8y/tap/ari` instructions will fail.

**Recommendation**: Debug the v0.1.0 release workflow logs to identify why the Homebrew formula was not pushed. Fix and validate with a v0.1.1 or v0.2.0 release.

### GAP-D02: 305 Unreleased Commits [HIGH]

**Evidence**: `git log --oneline v0.1.0..HEAD | wc -l` returns 305. The v0.1.0 release was 2026-01-07; current date is 2026-03-01. Nearly 2 months of development has not been released.

**Impact**: Any user who installed v0.1.0 has a binary that is 305 commits behind. The installed binary shows `ari dev (none, unknown)` even for the author, meaning the release pipeline is not part of the development workflow.

**Recommendation**: Cut a v0.2.0 release incorporating the single-binary completion, `ari init`, and all subsequent work. Establish a release cadence (monthly or milestone-based).

### GAP-D03: No Version Check or Update Mechanism [MEDIUM]

**Evidence**: No `ari update` or `ari self-update` command exists. No version check warning when running an old binary. The `version` command outputs local version but does not compare against latest release.

**Literature Reference**: SRC-003 recommends "reject unacceptably old versions" and SRC-004 frames micro-feedback loop friction around installation/update speed.

**Recommendation**: Phase 1: Add version-check warning on `ari sync` (compare embedded version against latest GitHub release tag via API). Phase 2: Consider `ari update` command for self-update.

### GAP-D04: No `install.sh` Script [MEDIUM]

**Evidence**: No `install.sh` file exists at repository root. The spike recommended creating one for curl-pipe-sh installation.

**Impact**: Users without Homebrew (or on systems where Homebrew is not appropriate) have no single-command installation path.

**Recommendation**: Create `install.sh` following the spike's template. Host at a stable URL. Test on macOS and Linux.

### GAP-D05: Documentation Stale and Incomplete [MEDIUM]

**Evidence**: `docs/guides/ariadne-cli.md` line 13 says `just build` but no Justfile exists. Line 22 says `just install`. No mention of Homebrew, GoReleaser, or binary download. Release notes reference Homebrew installation that does not work.

**Impact**: Any new user following the documentation will encounter broken instructions immediately.

**Recommendation**: Rewrite the Installation section of `docs/guides/ariadne-cli.md` with working instructions for: (1) Homebrew (once fixed), (2) binary download from GitHub Releases, (3) `go install`, (4) build from source via `CGO_ENABLED=0 go build ./cmd/ari`.

### GAP-D06: Windows Not in Build Matrix [LOW]

**Evidence**: `.goreleaser.yaml` `goos:` list contains `darwin` and `linux` only. The original spike recommended `windows` in the build matrix. The `format_overrides` section has a `windows: zip` entry but no Windows binaries are built.

**Impact**: Windows users cannot install ari at all. Low priority because the Claude Code ecosystem is primarily macOS/Linux.

**Recommendation**: Add `windows` to `goos:` when Windows users appear. The format_override is already configured correctly.

### GAP-D07: No `ari doctor` or Health Check Command [LOW]

**Evidence**: No health check command exists. SRC-003 recommends built-in health checks that validate prerequisites.

**Recommendation**: Create `ari doctor` that checks: (1) binary version vs latest, (2) KNOSSOS_HOME or embedded assets available, (3) `.claude/` directory health, (4) active session state consistency, (5) hook configuration validity. This reduces support burden as the user base grows.

### GAP-D08: No CLI Analytics or Usage Telemetry [LOW]

**Evidence**: No telemetry code exists. SRC-003 recommends "Google Analytics for your CLI" with `--incognito` opt-out.

**Impact**: No data on which commands are used, error rates, or usage patterns.

**Recommendation**: DEFER. For a single-developer project, this is premature. Revisit when external users exist. If implemented, follow SRC-003's opt-out pattern (`--incognito` flag, respect `DO_NOT_TRACK` env var).

### GAP-D09: No GoReleaser Snapshot Testing in CI [LOW]

**Evidence**: The `ariadne-tests.yml` workflow builds a binary with `go build` but does not run `goreleaser build --snapshot` to validate the GoReleaser configuration on every PR.

**Recommendation**: Add a GoReleaser snapshot build step to the CI pipeline. This catches configuration drift before release time.

### GAP-D10: Shell Completions Not Documented [LOW]

**Evidence**: Cobra generates `ari completion` command (visible in `ari --help` output), but no documentation mentions it. The existing CLI guide does not cover shell completions.

**Recommendation**: Add shell completion instructions to installation documentation: `ari completion bash > /etc/bash_completion.d/ari`, `ari completion zsh > "${fpath[1]}/_ari"`.

### GAP-D11: No GPG or Code Signing [LOW]

**Evidence**: No signing configured in `.goreleaser.yaml`. macOS Gatekeeper blocks unsigned downloaded binaries (SRC-009). Homebrew tap distribution bypasses Gatekeeper, but direct binary download does not.

**Recommendation**: Phase 3 as originally planned. Consider when direct binary download becomes a primary channel.

### GAP-D12: Release Notes Reference Non-Existent Files [LOW]

**Evidence**: `.goreleaser.yaml` archives include `docs/guides/ariadne-cli.md` in the release tarball `files:` list. This is fine if the file exists, but also includes `README.md` and `LICENSE*` which should be verified.

**Recommendation**: Verify that all referenced files exist at release time. GoReleaser will fail if they don't, but explicit verification in CI is safer.

---

## 5. Dogfooding Assessment

### Current Dogfooding Posture

The author is the sole user. The installed binary at `/Users/tomtenuta/.local/share/mise/installs/go/1.22.3/bin/ari` shows `ari dev (none, unknown)`, meaning:

1. The binary was installed via `go install` (which does not inject version info)
2. It is NOT the v0.1.0 release binary
3. It is running an unknown commit of the codebase

This violates the MEMORY.md documented trap: "After rebuilding, MUST also `cp ./ari $(which ari)` to update the installed binary." The distribution pipeline is completely disconnected from the development workflow.

### Dogfooding Against Literature Standards

| Literature Recommendation | Current State | Grade |
|---------------------------|---------------|-------|
| Use your own distribution channel (SRC-006, SRC-010) | Author uses `go install`, not Homebrew | F |
| Metrics infrastructure for dogfooding signal (SRC-010) | None | F |
| Diverse participant population (SRC-005) | 1 participant | N/A |
| Structured feedback channels (SRC-005) | Git commits with fix() prefix | D |
| Dedicated measurement (SRC-010) | None | F |

### The Fundamental Dogfooding Gap

The author is not eating the distributed product. The author is eating the source code. This means:

- **Installation friction is unknown**: The actual Homebrew installation experience has never been tested by a human
- **Update friction is unknown**: No one has ever upgraded from one version to another via the distribution channel
- **Binary staleness is invisible**: No mechanism detects that the running binary is outdated
- **Release quality is unvalidated**: The v0.1.0 binary was produced by GoReleaser but never functionally tested after release

---

## 6. Comparison Matrix: Current vs. Best Practice

| Dimension | Literature Best Practice | Knossos Current | Gap Severity |
|-----------|------------------------|-----------------|--------------|
| **Build** | Automated cross-platform | GoReleaser configured | NONE |
| **Package** | Multi-channel (Homebrew + direct + installer) | Homebrew broken, direct works | HIGH |
| **Release** | Automated on tag push | GitHub Actions configured | LOW (works but untested recently) |
| **Install** | One command, <2 min | `go build` requires toolchain | HIGH |
| **Update** | Version check + easy upgrade | Manual rebuild | HIGH |
| **Version** | `--version` with build metadata | Implemented but shows "dev" locally | MEDIUM |
| **Health** | `doctor` command validates setup | Not implemented | LOW |
| **Docs** | Current, accurate installation guide | Stale, references nonexistent Justfile | MEDIUM |
| **Analytics** | Opt-out telemetry | None | LOW (appropriate for scale) |
| **Signing** | GPG or code signing | None | LOW |
| **Testing** | CI validates release config | CI tests binary but not GoReleaser | LOW |
| **Dogfooding** | Structured program with metrics | Ad hoc, author-only | MEDIUM |

---

## 7. Recommendations (Priority Order)

### Immediate (this week)

1. **Fix Homebrew formula delivery** (GAP-D01). Debug v0.1.0 release logs. Verify HOMEBREW_TAP_TOKEN permissions. If the token works, manually push a formula. If not, regenerate the PAT.

2. **Cut v0.2.0 release** (GAP-D02). Tag and release to validate the full pipeline end-to-end, including Homebrew formula generation. This is the single most important action -- it simultaneously tests the release pipeline and provides a current binary.

3. **Author installs via Homebrew** (Dogfooding). After v0.2.0, the author should `brew install autom8y/tap/ari` and use that binary exclusively. This is the minimum viable dogfooding action.

### Short-term (next 2 weeks)

4. **Update installation documentation** (GAP-D05). Rewrite `docs/guides/ariadne-cli.md` Installation section with working instructions.

5. **Create `install.sh`** (GAP-D04). Follow the spike template. Test on macOS arm64 and Linux amd64 at minimum.

6. **Add GoReleaser snapshot to CI** (GAP-D09). Prevents configuration drift between releases.

### Medium-term (next month)

7. **Add version check warning** (GAP-D03). On `ari sync`, check if current version is behind latest GitHub release. Warn but do not block.

8. **Document shell completions** (GAP-D10). Low effort, high value for daily users.

9. **Create `ari doctor`** (GAP-D07). Health check command validating binary version, project state, and hook configuration.

### Deferred (backlog)

10. **Windows build matrix** (GAP-D06). Add when Windows users appear.
11. **Code signing** (GAP-D11). Add when direct binary download becomes primary.
12. **CLI telemetry** (GAP-D08). Add when external users exist.
13. **Establish release cadence** (GAP-D02 follow-up). Monthly or milestone-based releases.

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Homebrew token expired/rotated | HIGH | Release pipeline silently fails | Test with `--snapshot` before real release |
| GoReleaser config drift (untested for 2 months) | MEDIUM | Release fails on next tag push | Add snapshot build to CI |
| Binary size growth from embedded assets | LOW | Currently 34MB dev / 5MB release; acceptable | Monitor; strip with `-s -w` (already configured) |
| Stale embedded rites in distributed binary | MEDIUM | Users get outdated rite definitions | Filesystem sources override embedded; document version freshness |
| No rollback mechanism for bad releases | LOW | Broken release stays as "latest" | GoReleaser `mode: replace` allows overwriting |

---

## 9. Conclusion

The Knossos distribution initiative made strong architectural decisions: GoReleaser, single-binary with embedded assets, Homebrew tap, and `ari init` for bootstrapping. The implementation of the single-binary completion (TDD tasks 1-3) was thorough and well-tested.

However, the initiative stalled at the operational level. The release pipeline has not been exercised since its inaugural run, the primary distribution channel (Homebrew) is broken, the author does not use the distributed binary, and 305 commits sit unreleased. The gap is not in engineering quality but in **release discipline** -- the habit of regularly shipping through the distribution channels and using those channels personally.

The highest-leverage action is not a new feature or a complex engineering task. It is: fix the Homebrew token, tag v0.2.0, push the release, install via Homebrew, and use that binary for daily development. This single action resolves GAP-D01, GAP-D02, and the fundamental dogfooding gap simultaneously.

---

## References

- `.know/literature-developer-cli-distribution-dogfooding-best-practices.md` -- 13-source literature review
- `docs/spikes/SPIKE-cli-distribution.md` -- Original distribution research (2026-01-07)
- `docs/decisions/ADR-cli-distribution.md` -- Architecture decision record
- `docs/decisions/TDD-single-binary-completion.md` -- Technical design for embedding
- `.goreleaser.yaml` -- Current release configuration
- `.github/workflows/release.yml` -- Release automation
- `embed.go` -- Binary embedding declarations
- `internal/cmd/initialize/init.go` -- ari init implementation
