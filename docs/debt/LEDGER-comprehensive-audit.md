---
audit_id: session-20260302-232344-1b73b3a8
scan_date: "2026-03-02"
scope: "platform-wide"
status: "collection-complete"
surfaces_completed: [1, 2, 3, 4, 5, 6, 7, 8]
surfaces_remaining: []
total_items: 81
severity_breakdown:
  critical: 0
  high: 10
  medium: 43
  low: 28
---

# Comprehensive Debt Audit Ledger

## Summary

| Metric | Value |
|--------|-------|
| Total items | 81 |
| Surface 1 (Test Coverage) | 16 items |
| Surface 2 (Monolith Extraction) | 9 items |
| Surface 3 (Duplication) | 7 items |
| Surface 4 (Shell Scripts) | 5 items |
| Surface 5 (Naming/Schema) | 4 items |
| Surface 6 (Observability) | 8 items |
| Surface 7 (Architectural Boundaries) | 7 items |
| Surface 8 (Deferred Work) | 17 items |
| Surface 9 (Radar-Sourced) | 8 items |
| Severity: high | 10 |
| Severity: medium | 43 |
| Severity: low | 28 |
| Items carried from existing intelligence | 38 |
| Newly discovered items | 29 |
| Items corrected vs prior intelligence | 8 |
| Radar-sourced items (OPP cross-reference) | 8 |

### Coverage Profile Headline

**Total statement coverage**: 61.3% (1,754 functions; 502 at 0.0%)

The bimodal distribution hypothesis from the audit frame is confirmed:
- **Well-tested core** (>70%): materialize (80.2%), session (84.2%), inscription (83.2%), hook (69.5%), agent (87.6%)
- **Untested CLI surface** (<15%): output (11.7%), lint (12.9%), org (1.2%), explain (42.6%)
- **Previously-documented-as-zero now tested**: `internal/errors` at 100.0%, `internal/cmd/validate` at 59.2%, `internal/cmd/sync` at 47.2%, `internal/cmd/agent` at 51.5%, `internal/cmd/rite` at 20.5%, `internal/cmd/worktree` at 52.1%

Three stale items in `.know/test-coverage.md` require correction: `internal/errors` (documented as zero, actually 100%), `internal/cmd/validate` (documented as zero, actually 59.2%), `internal/cmd/sync` (documented as zero, actually 47.2%).

---

## Surface 1: Test Coverage Structural Gaps

### DEBT-100: output package has 11.7% statement coverage

- **Category**: Testing
- **Severity**: high
- **Title**: `internal/output` has 47 of 1,754 functions at 0% coverage
- **Location**: `internal/output/output.go`, `internal/output/rite.go`, `internal/output/manifest.go`
- **Impact**: Output formatting logic defines the CLI contract. JSON output is partially covered (PKG-013) but text formatting, rite output, and manifest output are untested. 47 functions at 0% -- the single largest uncovered package by function count.
- **Effort estimate**: 2-3 days (high surface area, many output types)
- **Cross-reference**: `.know/test-coverage.md` Priority 4, TENSION-004 (output is 781 lines + 696 lines)

### DEBT-101: lint package has 12.9% statement coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/lint` has 15 functions at 0% coverage despite having test files
- **Location**: `internal/cmd/lint/lint.go`
- **Impact**: Lint rules protect against SCAR-017 and SCAR-019 regressions. At 12.9% coverage, most lint rule implementations are untested. The lint test file exists but exercises a narrow subset.
- **Effort estimate**: 1-2 days
- **Cross-reference**: SCAR-017, SCAR-019 (lint rules without full coverage)

### DEBT-102: org command has 1.2% statement coverage

- **Category**: Testing
- **Severity**: low
- **Title**: `internal/cmd/org` has near-zero effective coverage
- **Location**: `internal/cmd/org/`
- **Impact**: Org commands are low-frequency administrative commands. Low blast radius.
- **Effort estimate**: 1 day

### DEBT-103: cmd/explain has 42.6% coverage with 18 functions at 0%

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/explain/context.go` has 18 zero-coverage functions
- **Location**: `internal/cmd/explain/context.go`
- **Impact**: `ari explain` context functions (contextRite, contextAgent, contextSession, etc.) provide the user-facing context dashboard. Untested formatting could produce confusing output.
- **Effort estimate**: 1 day

### DEBT-104: cmd/artifact has zero statement coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/artifact` CLI layer completely untested
- **Location**: `internal/cmd/artifact/artifact.go`, `list.go`, `query_cmd.go`, `rebuild.go`, `register.go`
- **Impact**: 13 functions at 0%. Artifact management commands have no CLI-level tests. The underlying `internal/artifact` library is at 88.0% but the command dispatch is untested.
- **Effort estimate**: 1-2 days
- **Cross-reference**: `.know/test-coverage.md` package listing

### DEBT-105: cmd/common has zero statement coverage

- **Category**: Testing
- **Severity**: low
- **Title**: `internal/cmd/common` shared context utilities untested
- **Location**: `internal/cmd/common/annotations.go`, `context.go`, `embedded.go`
- **Impact**: 16 functions at 0%. These are thin accessor functions (GetPrinter, GetResolver, etc.) -- low complexity but high coupling (used by every command).
- **Effort estimate**: 0.5 days

### DEBT-106: cmd/inscription has zero statement coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/inscription` CLI layer completely untested
- **Location**: `internal/cmd/inscription/`
- **Impact**: 25 functions at 0%. Inscription sync commands manage CLAUDE.md -- the core user-facing artifact. The underlying `internal/inscription` library is at 83.2%.
- **Effort estimate**: 1-2 days

### DEBT-107: cmd/manifest has zero statement coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/manifest` CLI layer completely untested
- **Location**: `internal/cmd/manifest/`
- **Impact**: 13 functions at 0%. Manifest commands (diff, merge, show, validate) have no CLI tests. `internal/manifest` library is at 44.3%.
- **Effort estimate**: 1 day

### DEBT-108: cmd/provenance has zero statement coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/provenance` CLI layer completely untested
- **Location**: `internal/cmd/provenance/`
- **Impact**: 10 functions at 0%. Provenance inspection commands untested. `internal/provenance` library is at 74.7%.
- **Effort estimate**: 1 day

### DEBT-109: cmd/tribute has zero statement coverage

- **Category**: Testing
- **Severity**: low
- **Title**: `internal/cmd/tribute` CLI layer completely untested
- **Location**: `internal/cmd/tribute/`
- **Impact**: Low-frequency command for tribute/acknowledgment generation. `internal/tribute` library is at 79.1%.
- **Effort estimate**: 0.5 days

### DEBT-110: cmd/naxos has zero statement coverage

- **Category**: Testing
- **Severity**: low
- **Title**: `internal/cmd/naxos` CLI layer completely untested
- **Location**: `internal/cmd/naxos/`
- **Impact**: Scanner commands. `internal/naxos` library is at 80.0%.
- **Effort estimate**: 0.5 days

### DEBT-111: cmd/root has zero statement coverage

- **Category**: Testing
- **Severity**: low
- **Title**: `internal/cmd/root` Cobra wiring layer untested
- **Location**: `internal/cmd/root/root.go`
- **Impact**: Low -- Cobra wiring only. Tested indirectly by all command tests.
- **Effort estimate**: 0.5 days

### DEBT-112: materialize/userscope has 23.7% coverage

- **Category**: Testing
- **Severity**: high
- **Title**: `internal/materialize/userscope` user-scope sync has very low coverage
- **Location**: `internal/materialize/userscope/sync.go` (1,530 lines)
- **Impact**: User-scope sync manages `~/.claude/` global state. At 23.7% on a 1,530-line file, the majority of sync logic is unexercised. 16 functions at 0%.
- **Effort estimate**: 3-4 days (large file, complex filesystem interactions)
- **Cross-reference**: MEMORY.md hotspot (secondary monolith), `.know/test-coverage.md`
- **FLAG FOR ASSESSOR**: This is the second-largest Go file and manages user-global state with low coverage.

### DEBT-113: config package has 34.6% coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/config` configuration management under-tested
- **Location**: `internal/config/`
- **Impact**: Config is foundational (KnossosHome, ActiveOrg, etc.). RISK-003 documents the `sync.Once` caching hazard. Low coverage means edge cases around XDG paths, org resolution, and env var fallbacks are unexercised.
- **Effort estimate**: 1 day
- **Cross-reference**: RISK-003, D3 (SourceResolver freezes config)

### DEBT-114: 502 functions at 0% coverage across all packages

- **Category**: Testing
- **Severity**: high
- **Title**: 28.6% of all functions (502/1,754) have zero statement coverage
- **Location**: Concentrated in `internal/output` (47), `internal/rite` (38), `internal/hook/clewcontract` (26), `internal/cmd/session` (25), `internal/cmd/inscription` (25), `internal/manifest` (24)
- **Impact**: Aggregate measure. Individual packages are itemized above. The 502 zero-coverage functions represent the total untested surface area.
- **Effort estimate**: N/A (aggregate; remediated via individual items)

### DEBT-115: .know/test-coverage.md contains stale data

- **Category**: Documentation
- **Severity**: medium
- **Title**: Three packages documented as zero-coverage are now tested
- **Location**: `.know/test-coverage.md` lines 45-69
- **Impact**: Stale intelligence causes incorrect prioritization. `internal/errors` is at 100%, `internal/cmd/validate` at 59.2%, `internal/cmd/sync` at 47.2% -- all documented as zero in the .know/ file.
- **Effort estimate**: 15 minutes (regenerate .know/test-coverage.md)
- **Cross-reference**: `.know/test-coverage.md` frontmatter `source_hash: 89b109c` vs current HEAD

---

## Surface 4: Shell Script Residue

### DEBT-116: context-injection.sh is a dead runtime dependency

- **Category**: Infrastructure > Shell residue
- **Severity**: medium
- **Title**: `rites/ecosystem/context-injection.sh` has zero runtime callers
- **Location**: `rites/ecosystem/context-injection.sh` (80 lines)
- **Impact**: The script defines `inject_rite_context()` which is documented as "called by session-context.sh via rite-context-loader.sh". However: (1) `rite-context-loader.sh` does not exist on disk (`user-hooks/` directory does not exist); (2) zero YAML or JSON configuration files reference `context-injection.sh`; (3) all hook invocations are now Go binaries per ADR-0011. The script is dead code masquerading as a live dependency. 37 documentation references create the illusion of activity.
- **Effort estimate**: 30 minutes to remove + documentation cleanup
- **Cross-reference**: ADR-0011 (hook binary decision), SCAR-015, SCAR-016

**Caller trace results**:
- `.yaml` references: 0
- `.json` references: 0
- Hook config references: 0
- Documentation references: 37 (across PRDs, context designs, shell rename reports, assessments)
- Source dependency: `rite-context-loader.sh` (does not exist on disk)
- Conclusion: **DEAD CODE** -- the entire call chain (`session-context.sh -> rite-context-loader.sh -> context-injection.sh`) was replaced by Go hooks.

### DEBT-117: validation.sh in cross-rite-handoff mena is shell in a Go codebase

- **Category**: Infrastructure > Shell residue
- **Severity**: low
- **Title**: `rites/shared/mena/cross-rite-handoff/validation.sh` is referenced from legomenon but not invoked by Go code
- **Location**: `rites/shared/mena/cross-rite-handoff/validation.sh` (144 lines)
- **Impact**: The script validates HANDOFF artifact frontmatter. It is referenced from `INDEX.lego.md` line 50 as a companion file within the cross-rite-handoff skill. It is materialized to `.claude/skills/cross-rite-handoff/validation.sh` where it is available as a reference/tool for CC agents to source. The `internal/validation/` Go package provides parallel validation capability. The script itself carries SCAR-016 risk (`set -euo pipefail` + bash arithmetic patterns).
- **Effort estimate**: 1-2 days to port to Go (`ari validate handoff`)
- **Cross-reference**: ADR-0011, SCAR-016, TDD-cross-rite-handoff.md

**Caller trace results**:
- Referenced from `INDEX.lego.md` as skill companion file
- Materialized to `.claude/skills/` (CC agent-accessible)
- Not imported or executed by any Go code
- Conclusion: **LIVE BUT SHELL** -- provides CC-accessible validation, should be ported per ADR-0011

### DEBT-118: e2e-validate.sh is CI-only with no Go equivalent

- **Category**: Infrastructure > Shell residue
- **Severity**: low
- **Title**: `scripts/e2e-validate.sh` orchestrates distribution validation in bash
- **Location**: `scripts/e2e-validate.sh` (341 lines)
- **Impact**: Runs in CI (GitHub Actions workflow `e2e-distribution.yml`), Dockerfile, and Makefile. Orchestrates brew tap/install and ari validation. Not on the runtime hot path. The script has timeout handling (`--brew-timeout`, `--ari-timeout` flags).
- **Effort estimate**: 2-3 days if porting to Go; 0 if accepted as bash
- **Cross-reference**: ADR-0011 (scope is hooks, not CI scripts), SCOUT-e2e-distribution-harness.md

**Caller trace results**:
- `Dockerfile.e2e`: COPY + ENTRYPOINT (active)
- `Makefile`: `e2e-local` target (active)
- `.github/workflows/e2e-distribution.yml`: direct execution (active)
- Conclusion: **LIVE, CI-ONLY** -- outside ADR-0011 scope (which covers hooks, not CI scripts). Acceptable as bash.

### DEBT-119: Shell script total footprint is 565 lines across 3 files

- **Category**: Infrastructure > Shell residue
- **Severity**: low
- **Title**: Aggregate shell script residue: 80 + 144 + 341 = 565 lines
- **Location**: Three files (see DEBT-116, DEBT-117, DEBT-118)
- **Impact**: Down from original shell-heavy architecture. The 80-line dead script (DEBT-116) should be removed. The 144-line validation script should be ported. The 341-line CI script is acceptable as-is.
- **Effort estimate**: N/A (aggregate)

### DEBT-120: 37 documentation references to dead context-injection.sh call chain

- **Category**: Documentation
- **Severity**: medium
- **Title**: Extensive docs reference defunct shell hook pipeline (context-injection.sh + rite-context-loader.sh)
- **Location**: `docs/requirements/PRD-rite-hook-context.md`, `docs/ecosystem/CONTEXT-DESIGN-team-context-loader.md`, `docs/hygiene/REFACTOR-shell-cleanse.md`, and 14 other docs
- **Impact**: Creates confusion about which infrastructure is active. New contributors may believe the shell hook pipeline is live. Documentation cleanup should accompany DEBT-116 removal.
- **Effort estimate**: 1-2 hours to add deprecation notices or remove references

---

## Surface 8: Deferred Work Items

### DEBT-121: Resume cross-rite -- still deferred, no ecosystem evidence

- **Category**: Architecture > Deferred feature
- **Severity**: low
- **Title**: Cross-rite resume protocol remains ecosystem-only
- **Location**: MEMORY.md "Deferred" section
- **Impact**: Resume (throughline persistence across sessions) works only in the ecosystem rite. Cross-rite rollout was deferred pending empirical evidence. No evidence of ecosystem-rite resume usage has been documented since the deferral.
- **Effort estimate**: 2-3 days for rollout if triggered
- **Staleness**: **VALID** -- decision to defer still applies. Code has not drifted.
- **Cross-reference**: CC Agent Capability Uplift, Wave 3 (DEFERRED)

### DEBT-122: arch-ref skill creation -- arch rite still has no mena

- **Category**: Architecture > Missing capability
- **Severity**: low
- **Title**: The `arch` rite has no mena directory and no reference skill
- **Location**: `rites/arch/` (confirmed: no `mena/` directory exists)
- **Impact**: Arch rite agents cannot preload reference material via skills. This was noted during CC Agent Uplift as an exception rite. Low impact since arch work is infrequent.
- **Effort estimate**: 2-4 hours to create arch-ref skill
- **Staleness**: **VALID** -- `rites/arch/mena/` still does not exist.

### DEBT-123: 10x-dev ghost skills -- NOT ghost, files exist as dromena

- **Category**: Architecture > Deferred cleanup
- **Severity**: low
- **Title**: `10x-ref`, `architect-ref`, `build-ref` exist as functional dromena, not ghosts
- **Location**: `rites/10x-dev/mena/10x-ref/INDEX.dro.md`, `rites/10x-dev/mena/architect-ref/INDEX.dro.md`, `rites/10x-dev/mena/build-ref/INDEX.dro.md`
- **Impact**: The MEMORY.md note "ghost skills -- declare nonexistent files" is **INCORRECT for 10x-dev**. All three exist as dromena with INDEX.dro.md files and are listed in `manifest.yaml` under `dromena:`. They are rite-switching commands (`/10x`, `/architect`, `/build`). Not ghost, not broken, not causing errors.
- **Effort estimate**: 5 minutes to correct MEMORY.md entry
- **Staleness**: **RESOLVED** -- the "ghost" label was inaccurate. These are functional dromena.
- **FLAG FOR ASSESSOR**: MEMORY.md contains stale/incorrect intelligence about this item.

### DEBT-124: ADR-0028 -- still unwritten

- **Category**: Documentation > Missing ADR
- **Severity**: medium
- **Title**: ADR-0028 (CC Agent Capability Uplift) remains unwritten
- **Location**: `docs/decisions/` (confirmed: no ADR-0028 file exists)
- **Impact**: The CC Agent Capability Uplift was a significant initiative (pilot + cross-rite rollout, 11 commits). Design decisions are captured in MEMORY.md but not formalized as an ADR. Without the ADR, future maintainers lack the "why" behind skill preloading ceilings, memory tier decisions, and triple-layer hook enforcement.
- **Effort estimate**: 2-3 hours to write from existing MEMORY.md notes
- **Staleness**: **VALID** -- still deferred, waiting for empirical evidence per original plan.

### DEBT-125: state.json last_sync is written but never read

- **Category**: Code > Dead write
- **Severity**: medium
- **Title**: `state.json` `last_sync` field is written on every sync but read by zero runtime consumers
- **Location**: `internal/sync/state.go:19` (struct definition), `internal/materialize/materialize_settings.go:142` (write site)
- **Impact**: Every `ari sync` writes `last_sync` to state.json. No runtime code reads it back. Only consumers are test assertions (`rite_switch_integration_test.go:89,126,294`). The `cmd/status` command has its own `LastSync` field computed independently.
- **Effort estimate**: 1-2 hours to remove field + update tests
- **Staleness**: **VALID** -- `active_rite` was removed from state.json in PKG-008 but `last_sync` was left as follow-up. The follow-up has not been executed.
- **Cross-reference**: `.know/design-constraints.md` DEBT-039, `internal/sync/state.go` CurrentSchemaVersion "1.1"

### DEBT-126: ADR-0027 dual event schema -- 3-version bridge still active

- **Category**: Architecture > Migration debt
- **Severity**: medium
- **Title**: Event read bridge supports v1/v2/v3 formats simultaneously
- **Location**: `internal/session/events_read.go` (153 lines, 3 format detection branches)
- **Impact**: The `ReadEvents()` function maintains 3 parallel parsing paths: v1 legacy (SCREAMING_CASE), v2 flat (snake_case), v3 typed (with `data` field). The write path is fully unified on `clewcontract.Event` (v3). The read bridge exists for `ari session audit` to read historical event logs. Removal trigger documented: "once all sessions created before ADR-0027 sprint 3 have been wrapped and archived."
- **Effort estimate**: 2-3 hours to remove once trigger is met
- **Staleness**: **VALID** -- the removal trigger (all pre-ADR-0027 sessions archived) has likely been met but has not been verified. `clewcontract.Event` is the canonical write type with 30+ references; `session.Event` (v1) appears only in `events_read.go` and one test comment.
- **Cross-reference**: TENSION-008, ADR-0027
- **FLAG FOR ASSESSOR**: The trigger condition may already be satisfied -- worth verifying.

### DEBT-127: Shell script deep cleanse -- partially complete

- **Category**: Infrastructure > Initiative tracking
- **Severity**: medium
- **Title**: Shell script elimination initiative partially complete
- **Location**: MEMORY.md "Current Priorities" item 1
- **Impact**: Listed as Priority 1 in MEMORY.md. Surface 4 analysis shows: script count reduced to 3 (from original larger inventory), `context-injection.sh` is dead code (DEBT-116), `validation.sh` needs Go port (DEBT-117), `e2e-validate.sh` is acceptable as-is (DEBT-118). The initiative is ~60% complete.
- **Effort estimate**: 1-2 days remaining (remove dead script + port validation)
- **Staleness**: **PARTIALLY STALE** -- MEMORY.md lists this as Priority 1 but the remaining work is small. Could be downgraded from "initiative" to "cleanup task."

### DEBT-128: Hook architecture (eliminate bash) -- largely complete

- **Category**: Infrastructure > Initiative tracking
- **Severity**: low
- **Title**: Bash hook elimination is nearly complete
- **Location**: MEMORY.md "Current Priorities" item 2
- **Impact**: Finding: `user-hooks/` directory does not exist. No `.sh` files exist in any hook path (excluding docs and .claude/). All hooks are Go binaries via `ari hook *`. The only remaining shell item is `context-injection.sh` (DEBT-116, dead code). This initiative can be closed once DEBT-116 is resolved.
- **Effort estimate**: 30 minutes (remove dead script, update MEMORY.md)
- **Staleness**: **NEARLY RESOLVED** -- the work is effectively done. MEMORY.md has not been updated to reflect completion.

### DEBT-129: Single-binary goals -- rite embedding exists, "remaining ports" unclear

- **Category**: Architecture > Initiative tracking
- **Severity**: medium
- **Title**: Single-binary initiative needs scope clarification
- **Location**: MEMORY.md "Current Priorities" item 3
- **Impact**: `embed.FS` is used in production: `EmbeddedRites`, `EmbeddedTemplates`, `EmbeddedAgents`, `EmbeddedMena` (in `embed.go`). Rite embedding is functional. "ari init" exists as a command. The phrase "remaining ports" is undefined -- no concrete task list exists for what would complete this initiative.
- **Effort estimate**: Unknown (scope undefined)
- **Staleness**: **DRIFTED** -- the work described has largely been done but the initiative was never formally closed or its remaining scope defined. This item needs scoping before it can be assessed.
- **FLAG FOR ASSESSOR**: Needs user input to determine if this initiative is complete.

### DEBT-130: state.json full elimination deferred

- **Category**: Code > Dead infrastructure
- **Severity**: low
- **Title**: `state.json` may be entirely eliminable
- **Location**: `internal/sync/state.go`, `.claude/sync/state.json`
- **Impact**: After removing `active_rite` (PKG-008) and potentially `last_sync` (DEBT-125), `state.json` would contain only `schema_version`. The question is whether the file serves any purpose beyond versioning itself. `StateManager.IsInitialized()` checks for file existence -- this could be replaced by checking for `.claude/` existence.
- **Effort estimate**: 2-3 hours (audit all StateManager callers, replace with simpler check)
- **Staleness**: **VALID** -- follow-up to PKG-008/DEBT-039.
- **Cross-reference**: DEBT-125, `.know/design-constraints.md` Knowledge Gap 1

### DEBT-131: SCAR regression gaps -- 9 of 27 SCARs lack tests

- **Category**: Testing > Regression coverage
- **Severity**: high
- **Title**: 9 SCARs have no automated regression tests
- **Location**: See scar-tissue.md Defensive Patterns table
- **Impact**: SCARs without regression tests: SCAR-002 (CC freeze -- structural fix), SCAR-004 (silent error discard), SCAR-008 (async hook spam), SCAR-015 (stdout pollution), SCAR-016 (bash arithmetic), SCAR-018 (context:fork), SCAR-020 (session ID subprocess), SCAR-023 (template path), SCAR-027 (shared mena anti-pattern). Behavioral guards (SCAR-004, SCAR-023) are highest risk for regression.
- **Effort estimate**: 3-5 days for all 9; 1 day for the 2 highest risk
- **Cross-reference**: `.know/scar-tissue.md` Defensive Patterns table
- **FLAG FOR ASSESSOR**: SCAR-004 and SCAR-023 are behavioral (not structural) fixes without regression tests.

### DEBT-132: rite package has 38 functions at 0% coverage (47.6% overall)

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/rite` has 38 uncovered functions despite 47.6% package coverage
- **Location**: `internal/rite/`
- **Impact**: Rite resolution, loading, and management. The covered functions exercise core resolution logic but listing, validation, and utility functions are untested.
- **Effort estimate**: 1-2 days

### DEBT-133: hook/clewcontract has 26 functions at 0% (81.2% overall)

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/hook/clewcontract` has 26 uncovered functions despite 81.2% package coverage
- **Location**: `internal/hook/clewcontract/`
- **Impact**: Clew contract types are the canonical event format. The high overall coverage masks that many type constructors and helper functions are untested. These are mostly simple constructors (low regression risk) but high coupling.
- **Effort estimate**: 1 day

### DEBT-134: cmd/session has 25 functions at 0% (59.0% overall)

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/session` has 25 uncovered functions despite 158 test functions
- **Location**: `internal/cmd/session/`
- **Impact**: Session commands are well-tested overall (158 test functions) but 25 functions have zero coverage. These are likely edge case handlers, output formatters, or less-used subcommands.
- **Effort estimate**: 1-2 days

### DEBT-135: manifest package has 24 functions at 0% (44.3% overall)

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/manifest` has 24 uncovered functions at 44.3% coverage
- **Location**: `internal/manifest/`
- **Impact**: Manifest parsing, diffing, and merging. At 44.3%, roughly half the package is untested. The black-box test style (4 external test files) provides good contract testing but internal helpers are unexercised.
- **Effort estimate**: 1-2 days

### DEBT-136: cmd/sails has 39.5% coverage

- **Category**: Testing
- **Severity**: medium
- **Title**: `internal/cmd/sails` quality gate CLI layer at 39.5% coverage
- **Location**: `internal/cmd/sails/`
- **Impact**: Sails gate commands manage quality gate lifecycle. The underlying `internal/sails` library is at 79.3% but CLI dispatch coverage is low.
- **Effort estimate**: 1 day

### DEBT-137: MEMORY.md contains inaccurate ghost skills note

- **Category**: Documentation
- **Severity**: low
- **Title**: MEMORY.md "10x-dev ghost skills" note is factually incorrect
- **Location**: MEMORY.md "Deferred" section, "Ghost skill detection" heuristic
- **Impact**: All three dromena (`10x-ref`, `architect-ref`, `build-ref`) exist as functional INDEX.dro.md files. The "ghost skill" label was accurate for a prior state but code has since been corrected or the original assessment was wrong.
- **Effort estimate**: 5 minutes (update MEMORY.md)
- **Cross-reference**: DEBT-123

---

## Surface 6: Observability and Silent Failure Debt

### DEBT-138: 30+ `_ =` and `_ :=` error-discarding patterns in non-test code

- **Category**: Code > Silent failure
- **Severity**: high
- **Title**: Systematic error discarding via blank identifier across 25+ files
- **Location**: `internal/materialize/userscope/sync.go` (11 sites), `internal/manifest/` (8 sites), `internal/worktree/` (5 sites), `internal/lock/lock.go` (3 sites), `internal/cmd/hook/` (5 sites), and 15+ other files
- **Impact**: 67 total `_ :=` sites and 30 `_ =` sites in non-test Go code. Risk-categorized:
  - **Correctness risk** (11 sites): `checksum.File()` errors discarded in userscope sync (6 sites in `sync.go`, 5 in `sync_cleanup.go`, 3 in `sync_mena.go`, 2 in `sync_agents.go`). A filesystem permission error or I/O failure would cause the checksum comparison to use a zero value, potentially triggering a no-op "unchanged" decision on a corrupted file.
  - **Data integrity risk** (8 sites): `json.Marshal` errors discarded in `internal/manifest/diff.go` (3 sites), `merge.go` (5 sites). These operate on already-parsed data so failure is unlikely, but would produce malformed output silently.
  - **Debugging hindrance** (15 sites): `os.UserHomeDir()` errors in `config/home.go` (3 sites), `filepath.Rel()` errors in agent/update and materialize files (10+ sites), `strconv.Atoi()` in worktree (2 sites).
  - **Intentional best-effort** (33+ sites): `os.Remove` cleanup, `cmd.MarkFlagRequired` (Cobra panics on actual failure), `json.Unmarshal` on untrusted data with fallback paths.
- **Effort estimate**: 2-3 days to add error handling to the 19 correctness+integrity sites; debugging sites are lower priority
- **Cross-reference**: RISK-001 (RESOLVED -- transform failures now return errors), RISK-004 (PARTIALLY RESOLVED -- CleanEmptyDirs now returns errors), RISK-005 (RESOLVED -- LoadOrBootstrap now propagates errors), SCAR-004
- **FLAG FOR ASSESSOR**: The 16 checksum-related discard sites in userscope/ are the highest-risk subset.

### DEBT-139: Zero `log.Debug` infrastructure in production code

- **Category**: Infrastructure > Observability
- **Severity**: medium
- **Title**: No debug-level logging exists anywhere in the codebase
- **Location**: All `internal/` packages
- **Impact**: Zero occurrences of `log.Debug` in the entire codebase. The only logging available is `log.Printf("Warning: ...")` (25 occurrences) and `printer.VerboseLog()` (80 occurrences across 29 files). `VerboseLog` requires `--verbose` flag AND is only available in command contexts (not library code). Library packages (`materialize/`, `session/`, `mena/`, `provenance/`) have no debug output mechanism. When pipeline steps silently succeed but produce wrong results, there is no way to trace execution without adding temporary print statements.
- **Effort estimate**: 1 day to add `log.Debug` wrapper + environment variable toggle; 2-3 days to instrument key paths
- **Cross-reference**: D4 from SPIKE (extractEmbeddedMena swallows 6 error paths with zero logging)

### DEBT-140: `extractEmbeddedMenaToXDG` swallows 6 error paths silently

- **Category**: Code > Silent failure
- **Severity**: medium
- **Title**: XDG mena extraction has 6 error paths that silently return/continue
- **Location**: `internal/cmd/initialize/init.go:275-321`
- **Impact**: `RemoveAll`, `MkdirAll`, `ReadFile`, `WriteFile`, `WalkDir` skip, and sentinel-write errors are all silently discarded. Intentionally best-effort (documented), but debugging "why is my mena wrong?" is painful with zero logging on any failure path.
- **Effort estimate**: 15 minutes (add `log.Printf` on each failure path)
- **Cross-reference**: D4 from SPIKE-path-resolution-hierarchy-debt.md

### DEBT-141: Hook handlers have no silent-failure patterns -- graceful degradation is intentional

- **Category**: Architecture > Observation (not debt)
- **Severity**: low
- **Title**: Hook handler error patterns are consistent and intentional
- **Location**: `internal/cmd/hook/*.go` (15 handlers)
- **Impact**: All 15 hook handlers follow a consistent pattern: (1) `ctx.withTimeout()` wrapping, (2) `ctx.getHookEnv()` stdin parsing, (3) event type guard (`hookEnv.Event != "" && hookEnv.Event != hook.EventXxx`), (4) `ctx.resolveSession()` for session-aware hooks. Error handling is uniform: session resolution failures log via `VerboseLog("warn")` and return a no-op JSON response. This is intentional graceful degradation -- hook failures should never block CC.
  - **Exception**: `cheapo_revert.go` (line 88) and `worktreeremove.go` bypass the shared `cmdContext` helpers (no `getHookEnv`, no event guard). These are newer hooks with a different stdin-parsing pattern (`io.ReadAll` + manual `json.Unmarshal` instead of `getHookEnv()`).
  - **Exception**: `sessionend.go:128` discards `writer.Write(event)` error. The session-ended event being dropped means audit trail gap, but the session state change (park) has already been saved.
- **Effort estimate**: N/A (observation, not actionable debt)
- **Cross-reference**: SCAR-010 (withTimeout requirement), hook handler count now 15 (up from 11 documented in task)

### DEBT-142: `cheapo_revert.go` and `worktreeseed.go` duplicate embedded FS wiring boilerplate

- **Category**: Code > Duplication
- **Severity**: low
- **Title**: Two hook handlers duplicate the Materializer embedded-FS setup pattern
- **Location**: `internal/cmd/hook/cheapo_revert.go:62-72`, `internal/cmd/hook/worktreeseed.go:159-163`
- **Impact**: Both hooks manually construct a `Materializer` and wire 2-4 embedded FS sources using identical `if embXxx := common.EmbeddedXxx(); embXxx != nil { m.WithEmbeddedXxx(embXxx) }` boilerplate. `cheapo_revert` wires 4 sources (Rites, Templates, Agents, Mena); `worktreeseed` wires only 2 (Rites, Templates), creating an inconsistency -- `worktreeseed` is missing embedded Agents and Mena. The same pattern exists in `internal/cmd/sync/sync.go` as the canonical version.
- **Effort estimate**: 1-2 hours to extract a shared `common.NewWiredMaterializer(resolver)` helper
- **Cross-reference**: DEBT-141 (hook handler consistency observations)
- **FLAG FOR ASSESSOR**: `worktreeseed` missing 2 of 4 embedded FS sources may cause incomplete materialization in worktrees.

### DEBT-143: Stale engine.go error patterns partially remediated

- **Category**: Code > Silent failure
- **Severity**: medium
- **Title**: `mena/engine.go` stale-entry removal discards `os.RemoveAll` and `os.Remove` errors
- **Location**: `internal/materialize/mena/engine.go:263-272` (stale entry removal), `engine.go:88-96` (dromena INDEX cleanup)
- **Impact**: `CleanEmptyDirs` was fixed (now returns errors as warnings per RISK-004). However, the stale-entry removal block at lines 263-272 still silently discards `os.RemoveAll` and `os.Remove` errors. If a stale mena entry cannot be removed (permission error, file lock), the pipeline reports success while orphan files persist. The dromena cleanup block (lines 88-96) also discards `os.Remove` for old INDEX.md files and empty directories.
- **Effort estimate**: 30 minutes to collect these errors into `result.Warnings`
- **Cross-reference**: RISK-004 (partially resolved), SCAR-006

### DEBT-144: `log.Printf("Warning: ...")` output has no user-visible channel

- **Category**: Infrastructure > Observability
- **Severity**: medium
- **Title**: 25 `log.Printf("Warning: ...")` calls in materialize pipeline go to stderr with no user feedback
- **Location**: `internal/materialize/materialize.go` (2 sites), `agent_transform.go` (1), `skill_policies.go` (1), `materialize_settings.go` (2), `mena/namespace.go` (3), `mena/engine.go` (3), `mena/collect.go` (1), `mena/frontmatter.go` (1), `userscope/sync.go` (2), `orgscope/sync.go` (4)
- **Impact**: These warnings go to Go's default `log` package stderr output. In CC hook context, stderr is captured but not shown to the user by default. During `ari sync` direct invocation, warnings appear in terminal stderr but are easily missed. There is no mechanism to promote warnings to a `SyncResult.Warnings` field that would appear in structured output.
- **Effort estimate**: 1-2 days to add a `SyncResult.Warnings []string` field and plumb through pipeline
- **Cross-reference**: DEBT-139 (no debug infrastructure)

### DEBT-145: 3 RISK items from design-constraints.md are now resolved

- **Category**: Documentation > Stale intelligence
- **Severity**: low
- **Title**: RISK-001, RISK-004, RISK-005 have been partially or fully addressed since documentation
- **Location**: `.know/design-constraints.md` Risk Zone Mapping section
- **Impact**: Current state vs documented state:
  - **RISK-001** (agent transform silent fallback): **RESOLVED**. `transformAgentContent` failures now return `fmt.Errorf` at all 3 call sites in `materialize_agents.go:60,113,167`. The original description of "writes untransformed content" is no longer accurate.
  - **RISK-004** (CleanEmptyDirs swallows errors): **PARTIALLY RESOLVED**. `CleanEmptyDirs` now returns `[]error`. Callers surface these as `result.Warnings`. However, the stale-entry removal block (DEBT-143) still discards errors.
  - **RISK-005** (provenance.Load swallowed on warm path): **RESOLVED**. `LoadOrBootstrap` errors now propagate and abort the pipeline at both `MaterializeMinimal:244` and `MaterializeWithOptions:374`.
- **Effort estimate**: 15 minutes to update `.know/design-constraints.md`

---

## Surface 3: Duplicated Logic and Missing Abstractions

### DEBT-146: Three independent `copyDir` implementations

- **Category**: Code > Duplication (DRY violation)
- **Severity**: medium
- **Title**: Three `copyDir` functions with overlapping but divergent implementations
- **Location**:
  - `internal/materialize/mena/walker.go:40` -- `copyDirFS(fsys fs.FS, root, dst string, hideCompanions bool)` -- unified fs.FS walker with mena-specific stripping and companion hiding
  - `internal/materialize/materialize.go:705` -- `(m *Materializer) copyDir(src, dst string)` -- simple recursive copy using `filepath.WalkDir` + `fileutil.WriteIfChanged`
  - `internal/cmd/session/create.go:353` -- `copyDir(src, dst string)` -- standalone recursive copy using `os.ReadDir` + `copyFile` helper
- **Impact**: The mena walker is specialized (extension stripping, companion hide, INDEX promotion). The other two are general-purpose recursive directory copies with different implementations: the materialize version uses `filepath.WalkDir` and `WriteIfChanged` (atomic writes); the session version uses `os.ReadDir` recursion and direct `os.WriteFile`. The session version does NOT use `WriteIfChanged`, meaning non-atomic writes in session directory copies. A common `fileutil.CopyDir` would eliminate the duplication and ensure consistent write semantics.
- **Effort estimate**: 1-2 hours to extract a shared `fileutil.CopyDir` and replace the 2 general-purpose copies
- **Cross-reference**: TENSION-006 (shared manifest loaders -- now RESOLVED, see DEBT-148)

### DEBT-147: Dual source chain construction (inline vs `BuildSourceChain`)

- **Category**: Code > Duplication (DRY violation)
- **Severity**: medium
- **Title**: Mena source chain built inline in materializer and separately in `mena.BuildSourceChain()`
- **Location**:
  - `internal/materialize/materialize_mena.go:24-100` -- inline chain construction with embedded FS interleaving
  - `internal/mena/source.go:46-71` -- extracted `BuildSourceChain()` for filesystem-only
- **Impact**: The materializer duplicates chain logic because it must handle embedded FS and the satellite-vs-knossos `sharedRitesBase` distinction. The validator uses `BuildSourceChain()`. If someone adds a new source tier to `BuildSourceChain()`, they must remember to update the materializer too. "Works in validation, broken in sync" bugs would be hard to catch. No alignment comment exists.
- **Effort estimate**: 2-3 hours for full unification; 5 minutes for alignment comment
- **Cross-reference**: D1 from SPIKE-path-resolution-hierarchy-debt.md

### DEBT-148: TENSION-006 (shared manifest loaders) is resolved

- **Category**: Code > Duplication (RESOLVED)
- **Severity**: low
- **Title**: `loadSharedHookDefaults` and `loadSharedSkillPolicies` now share `loadSharedManifest`
- **Location**: `internal/materialize/agent_transform.go:143-183` (shared loader), `:188` (hook defaults caller), `skill_policies.go:266` (skill policies caller)
- **Impact**: The original TENSION-006 documented two structurally identical load paths. These have been unified into a single `(m *Materializer) loadSharedManifest(resolved *ResolvedRite) (*RiteManifest, error)` method. Both callers now delegate to it and extract their specific field. The three-tier resolution (embedded FS, KnossosHome, project root fallback) is implemented once.
- **Effort estimate**: N/A (resolved)
- **Cross-reference**: `.know/design-constraints.md` TENSION-006

### DEBT-149: Hook handler boilerplate is well-factored; 2 outlier hooks diverge

- **Category**: Code > Duplication (partial)
- **Severity**: low
- **Title**: 13 of 15 hook handlers use shared `cmdContext` helpers; 2 diverge
- **Location**:
  - Standard pattern (13 hooks): `RunE -> ctx.withTimeout(func() { runXxx(ctx) })`, `runXxx` calls `ctx.getPrinter()`, `ctx.getHookEnv()`, event guard, `ctx.resolveSession()`
  - Divergent: `cheapo_revert.go` -- inlines RunE body (no `runXxx`), creates its own printer, does not use `getHookEnv()` or event guard, does not use `withTimeout()`
  - Divergent: `worktreeremove.go` -- reads stdin directly via `io.ReadAll`, no `getHookEnv()`, no event guard, no `withTimeout()`
  - Partially divergent: `worktreeseed.go` -- reads stdin directly but structurally similar to worktreeremove
- **Impact**: The 2-3 divergent hooks miss the timeout safety net (`withTimeout`). `cheapo_revert` runs a full `m.Sync()` without timeout, meaning a slow sync could block CC indefinitely. `worktreeremove` shells out to `git worktree remove` without timeout. The event guard absence means these hooks would execute even on wrong event types (though CC routing makes this unlikely).
- **Effort estimate**: 2-4 hours to refactor the 3 outlier hooks to use shared helpers
- **Cross-reference**: DEBT-142 (embedded FS wiring duplication in same hooks), SCAR-010 (timeout requirement)
- **FLAG FOR ASSESSOR**: `cheapo_revert` running `m.Sync()` without timeout is a latent reliability risk.

### DEBT-150: Session test setup duplication across 6 test files

- **Category**: Testing > Duplication
- **Severity**: medium
- **Title**: Six near-identical `setupXxxTestSession()` functions in `internal/cmd/session/`
- **Location**:
  - `field_test.go:17` -- `setupFieldTestSession(t, complexity, initiative)`
  - `log_test.go:18` -- `setupLogTestSession(t)`
  - `snapshot_test.go:19` -- `setupSnapshotTestSession(t, contextBody)`
  - `timeline_cmd_test.go:19` -- `setupTimelineTestSession(t)`
  - `moirai_integration_test.go:37` -- `newTestContext(projectDir, sessionID...)` + `setupProjectDir(t)`
  - `lock_test.go:222` -- `setupTestEnv(t)`
- **Impact**: All 6 functions create: tmpDir, sessions directory structure, SESSION_CONTEXT.md with YAML frontmatter, `.current-session` marker, `cmdContext` with output/verbose/projectDir. The context content varies slightly (different session IDs, different body content). Any change to the session directory structure or context schema requires updating all 6 setup functions. `moirai_integration_test.go` has a slightly cleaner abstraction (`newTestContext` + `setupProjectDir`).
- **Effort estimate**: 2-3 hours to extract a shared `testutil.SetupSessionTestEnv(t, opts)` helper
- **Cross-reference**: DEBT-134 (25 uncovered functions in cmd/session)

### DEBT-151: Platform mena resolution computed differently by 2 callers

- **Category**: Code > Duplication
- **Severity**: low
- **Title**: `getMenaDir()` and validator receive platform mena path via different mechanisms
- **Location**:
  - `internal/materialize/materialize_mena.go:121-144` -- `getMenaDir()` (unexported Materializer method)
  - `internal/registry/validate.go:76` -- receives `platformMenaDir` as parameter (dependency injection)
- **Impact**: Both consumers need the same value but compute it differently. The validator's DI approach is cleaner. A third consumer would need to reimplement. Not broken today but structurally fragile.
- **Effort estimate**: 30 minutes to extract shared function to `internal/mena/`
- **Cross-reference**: D2 from SPIKE-path-resolution-hierarchy-debt.md

### DEBT-152: `copyDirWithStripping` / `copyDirFromFSWithStripping` duplication is RESOLVED

- **Category**: Code > Duplication (RESOLVED)
- **Severity**: low
- **Title**: The two mena walker functions have been unified into `copyDirFS`
- **Location**: `internal/materialize/mena/walker.go:40` -- `copyDirFS` with unified `fs.FS` interface
- **Impact**: The walker comment at line 37-39 explicitly documents that `copyDirFS` is "the unified replacement for the two previously separate functions: copyDirWithStripping (filesystem) and copyDirFromFSWithStripping (embed.FS)." Both embedded and filesystem sources now go through `openMenaFS()` which returns an `fs.FS` adapter (`os.DirFS` for filesystem, `fs.Sub` for embedded). This was a known debt item from `.know/design-constraints.md` Abstraction Gap Mapping.
- **Effort estimate**: N/A (resolved)

---

## Surface 5: Naming and Schema Debt

### DEBT-153: Dual `OwnerType` definitions remain (TENSION-001 confirmed active)

- **Category**: Design > Naming collision
- **Severity**: medium
- **Title**: `inscription.OwnerType` and `provenance.OwnerType` remain as separate incompatible types
- **Location**:
  - `internal/inscription/types.go:19` -- values: `knossos`, `satellite`, `regenerate` (region ownership)
  - `internal/provenance/provenance.go:83` -- values: `knossos`, `user`, `untracked` (file ownership)
- **Impact**: Both types share the name `OwnerType` and both have a `knossos` constant, but they are semantically distinct. Cross-references via `NOTE:` comments exist in both files (inscription.go line 16, provenance.go line 80), which is the minimum viable guard. The collision cannot cause compile-time errors (different packages) but can cause confusion during code review and grep-based audit. The value sets are genuinely different -- merging them would create a 5-value enum with context-dependent validity.
- **Effort estimate**: Medium (rename one, e.g., `inscription.RegionOwner`, plus all consumers in `materialize/` pipeline)
- **Cross-reference**: `.know/design-constraints.md` TENSION-001

### DEBT-154: `McpServerConfig` vs `MCPServer` naming inconsistency spans 3 types

- **Category**: Design > Naming inconsistency
- **Severity**: low
- **Title**: MCP-related types use inconsistent Go naming conventions across packages
- **Location**:
  - `internal/agent/types.go:129` -- `McpServerConfig` (Go camelCase, not idiomatic for acronyms)
  - `internal/materialize/materialize.go:49` -- `MCPServer` (correct Go acronym convention)
  - `internal/materialize/hooks/mcp.go:13` -- `MCPServerConfig` (correct Go acronym convention)
- **Impact**: `McpServerConfig` in `agent/types.go` violates Go naming conventions (acronyms should be all-caps: `MCPServerConfig`). The `materialize/mcp.go:10-13` bridge function `mergeMCPServers` explicitly converts between `materialize.MCPServer` and `hooks.MCPServerConfig`. The agent type is the odd one out. This causes no runtime issues but creates inconsistency for contributors.
- **Effort estimate**: 30 minutes to rename `McpServerConfig` to `MCPServerConfig` in `agent/types.go` and update the 3 callers
- **Cross-reference**: `.know/design-constraints.md` (not previously documented)

### DEBT-155: Zero satellite manifests use deprecated `Commands`/`Skills` fields

- **Category**: Design > Deprecated compat shim
- **Severity**: medium
- **Title**: `RiteManifest.Commands` and `RiteManifest.Skills` backward-compat fields have zero active consumers
- **Location**: `internal/materialize/materialize.go:68-69` (field definitions)
- **Impact**: Grep across all `manifest.yaml` files in `rites/` and the broader codebase returns zero matches for `commands:` or `skills:` as top-level YAML keys. All rite manifests use `dromena:` and `legomena:` exclusively. The `Commands` and `Skills` fields in `RiteManifest` exist solely for backward compatibility with satellite manifests that may still use old terminology. However, zero satellite manifests were found using these fields. The `Skills` field has a `// Deprecated` comment; `Commands` has a `// Backward compat` comment.
  - **Important caveat**: Satellite manifests outside this repository were not audited. The deprecated fields may still have external consumers.
- **Effort estimate**: 1 hour to remove if satellite audit confirms zero usage; add deprecation warning log if keeping
- **Cross-reference**: `.know/design-constraints.md` TENSION-002, ADR-0023

### DEBT-156: `SourceType` string convention between `provenance` and `source` packages is stable

- **Category**: Design > Naming (observation, low risk)
- **Severity**: low
- **Title**: `SourceType` alignment between provenance and materialize/source maintained by convention
- **Location**:
  - `internal/materialize/source/types.go:7-21` -- typed `SourceType` with 6 constants
  - `internal/provenance/provenance.go:59-68` -- plain `string` field with comment documenting alignment
- **Impact**: The provenance package intentionally uses plain strings (documented at line 65: "It intentionally uses plain strings rather than importing source.SourceType"). The comment at line 60 explicitly lists the matching values. This is a deliberate architectural decision (ADR-0026 keeps provenance as a leaf package with no internal imports), not accidental drift. Cross-package `provenance.SourceType string` values are validated at write time by the materializer which uses the typed constants. 6 values in `source/types.go`, same 6 referenced in provenance comment. The `session.Event` (System A) type has been fully eliminated from non-test code (only 1 test comment references it). `clewcontract.Event` (System B) is the sole production type with 30+ references. The events_read.go bridge (DEBT-126) is the only file maintaining dual-format awareness.
- **Effort estimate**: N/A (intentional design)
- **Cross-reference**: `.know/design-constraints.md` TENSION-007, TENSION-008

---

## Surface 7: Architectural Boundary Debt

### Leaf Package Verification

Documented leaf packages from `.know/architecture.md`: `internal/mena/`, `internal/registry/`, `internal/errors/`, `internal/fileutil/`, `internal/checksum/`, `internal/tokenizer/`, `internal/assets/`.

**Results**:
- `internal/mena/` -- PASS. stdlib only (io/fs, os, path/filepath, strings).
- `internal/errors/` -- PASS. stdlib only.
- `internal/fileutil/` -- PASS. stdlib only.
- `internal/checksum/` -- PASS. stdlib only.
- `internal/tokenizer/` -- PASS. stdlib only.
- `internal/assets/` -- PASS. stdlib only.
- `internal/config/` -- PASS. stdlib only.
- `internal/registry/` -- **FAIL**. Imports `internal/frontmatter` and `internal/mena`. See DEBT-157.

### DEBT-157: `internal/registry` is NOT a leaf package (imports frontmatter, mena)

- **Category**: Architecture > Leaf violation
- **Severity**: medium
- **Title**: Registry documented as leaf but imports two internal packages
- **Location**: `internal/registry/registry.go` (180 lines)
- **Impact**: `.know/architecture.md` line 71 lists `internal/registry` as a leaf package ("no internal imports"). Actual imports: `internal/frontmatter` (YAML parsing) and `internal/mena` (type detection). Both are foundation-layer packages, so this is not an upward violation, but the documentation is incorrect. The registry uses `frontmatter.Parse` and `mena.DetectMenaType` for rite discovery -- these are structurally necessary.
- **Effort estimate**: 15 minutes to correct `.know/architecture.md` documentation; actual code refactoring not warranted (the imports are appropriate for the registry's role)
- **Cross-reference**: `.know/architecture.md` line 71

### Sub-Package Boundary Verification (materialize/ sub-packages)

| Sub-package | Sibling imports | Assessment |
|-------------|----------------|------------|
| `mena/` | None | Clean leaf within materialize |
| `source/` | None | Clean leaf within materialize |
| `hooks/` | None | Clean leaf within materialize |
| `orgscope/` | None | Clean leaf within materialize |
| `userscope/` | **imports `mena/`** | Sibling coupling |

### DEBT-158: `userscope/` imports sibling `mena/` sub-package (tight coupling)

- **Category**: Architecture > Sub-package coupling
- **Severity**: medium
- **Title**: userscope directly imports materialize/mena for 7 distinct operations
- **Location**: `internal/materialize/userscope/sync_mena.go:10`
- **Impact**: `userscope` uses `mena.CollectMena`, `mena.MenaSource`, `mena.MenaProjectionOptions`, `mena.ProjectAll`, `mena.StripMenaExtension`, `mena.InjectCompanionHideFrontmatter`, `mena.DetectMenaType`, and `mena.CleanEmptyDirs` (18 call sites across sync_mena.go). This means userscope cannot evolve independently of mena's internal API. Changes to mena's collection interface force userscope changes. The coupling is functional (userscope genuinely needs mena operations to sync user-scoped mena files), but the depth of usage suggests userscope is re-implementing rite-scope mena logic for user-scope context rather than sharing a common abstraction.
- **Effort estimate**: 1-2 days to extract shared mena sync interface; medium blast radius (sync_mena.go is 654 lines)
- **Cross-reference**: DEBT-146 (copyDir duplication), DEBT-151 (mena resolution duplication)

### Layer Boundary Verification

**Layer diagram from `.know/architecture.md`**:
- Layer 1 (CLI): `internal/cmd/*`
- Layer 2 (Domain): `materialize`, `inscription`, `session`, `rite`, `agent`, `provenance`, `sails`, `artifact`, `worktree`
- Layer 3 (Support): `mena`, `manifest`, `sync`, `lock`, `validation`, `hook`, `registry`, `know`, `naxos`, `tribute`, `tokenizer`, `output`
- Layer 4 (Foundation): `paths`, `frontmatter`, `fileutil`, `checksum`, `config`, `assets`
- Layer 5 (Cross-cut): `errors`

**Upward violation check (domain importing CLI)**: NONE found. All 9 domain packages verified -- zero imports of `internal/cmd/*`.

**Upward violation check (support importing domain)**: 2 violations found.

### DEBT-159: `internal/naxos` (Layer 3) imports `sails` and `session` (Layer 2)

- **Category**: Architecture > Layer violation
- **Severity**: high
- **Title**: Support-layer naxos scanner imports two domain-layer packages
- **Location**: `internal/naxos/` imports `internal/sails` and `internal/session`
- **Impact**: Naxos is the linter/scanner for agent and mena files. It is documented in `.know/architecture.md` as Layer 3 (Support). However, it imports `internal/sails` (Layer 2 Domain) and `internal/session` (Layer 2 Domain). This creates an upward dependency: a support package depends on domain packages, meaning the linter cannot be used without the full domain layer. This also means changes to session or sails types can break the linter. The layer assignment may be wrong -- naxos may belong in Layer 2 given its dependencies. Alternatively, the types it needs from sails/session could be extracted to a shared types package.
- **Effort estimate**: 1-2 hours if re-classifying naxos as Layer 2; 1-2 days if extracting shared types
- **Cross-reference**: `.know/architecture.md` layer diagram
- **FLAG FOR ASSESSOR**: Layer violations indicate architectural boundary erosion

### DEBT-160: `internal/tribute` (Layer 3) imports `artifact` and `session` (Layer 2)

- **Category**: Architecture > Layer violation
- **Severity**: high
- **Title**: Support-layer tribute imports two domain-layer packages
- **Location**: `internal/tribute/` imports `internal/artifact` and `internal/session`
- **Impact**: Tribute generates tribute documents and needs session state and artifact registry data. Same structural issue as DEBT-159: a Layer 3 package depends on Layer 2 packages. Tribute and naxos together suggest the Layer 3 classification is wrong for packages that need domain context to do their work. The `.know/architecture.md` layer diagram should either (a) reclassify naxos and tribute as Layer 2, or (b) acknowledge that these packages bridge layers.
- **Effort estimate**: 15 minutes if reclassifying; structural refactoring not warranted
- **Cross-reference**: DEBT-159, `.know/architecture.md` layer diagram
- **FLAG FOR ASSESSOR**: Paired with DEBT-159, suggests layer diagram needs update

### DEBT-161: `internal/output` (Layer 3) has zero internal imports -- correct leaf

- **Category**: Architecture > Observation (not debt)
- **Severity**: low
- **Title**: Output package has no internal imports despite being a hub
- **Location**: `internal/output/` -- 781 + 696 + 240 = 1,717 lines across 3 files
- **Impact**: `.know/architecture.md` lists `output` as a "hub package" imported by many. Despite this, output itself imports ZERO internal packages. It depends only on stdlib + gopkg.in/yaml.v3. This is architecturally clean -- output is a pure formatting layer. However, the 29 output struct types (16 in output.go, 13 in rite.go) defined here create tight coupling with domain types through their field structures, even without explicit imports. Any domain type change requires matching output type changes. This is a coupling-by-convention pattern, not a Go import violation.
- **Effort estimate**: N/A (observation, not actionable debt)

### DEBT-162: `worktree` (Layer 2) imports `materialize` (Layer 2) -- cross-domain coupling

- **Category**: Architecture > Cross-domain coupling
- **Severity**: medium
- **Title**: Worktree package directly invokes materialize.Sync for worktree ecosystem setup
- **Location**: `internal/worktree/operations.go:659-673`
- **Impact**: `worktree.Manager.setupWorktreeEcosystem()` creates a new `materialize.Materializer` and calls `mat.Sync()` directly. This is the only place outside the CLI layer that triggers a full materialization. The coupling means worktree operations depend on the full materialize package graph (15 internal imports). This is a domain-to-domain coupling that the architecture description does not document. It does not create a cycle (materialize does not import worktree), but it means worktree changes could be affected by materialize API changes.
- **Effort estimate**: Medium -- could be decoupled by having the CLI layer orchestrate the sync after worktree creation instead of embedding it
- **Cross-reference**: `.know/architecture.md` import patterns

### Knowledge Gap Verification

| Gap # | Description | Status |
|-------|-------------|--------|
| KG-1 | `internal/worktree` and `internal/tribute` internals | **PARTIALLY FILLED** -- worktree has 6,522 lines across files, tribute has 5 files. Both are now documented via import analysis and concern grouping in this audit. Detailed function-level documentation still missing from `.know/`. |
| KG-2 | `internal/naxos` scanner rules | **UNFILLED** -- 1,646 lines, 5 files. Layer violation documented (DEBT-159) but lint rules not cataloged. |
| KG-3 | `internal/materialize/hooks/` sub-directory | **UNFILLED** -- 2 files (config.go, mcp.go) + tests. Hook materialization specifics not read. |
| KG-4 | `internal/materialize/mena/` sub-directory | **PARTIALLY FILLED** -- import analysis complete, copyDirFS unification documented (DEBT-152). Internal function catalog missing. |
| KG-5 | `internal/materialize/userscope/` sub-directory | **PARTIALLY FILLED** -- split from 1,530-line monolith into 7 files totaling 2,716 lines (non-test). Sibling coupling documented (DEBT-158). |
| KG-6 | `internal/sails/generator.go`, `thresholds.go` | **PARTIALLY FILLED** -- concern analysis done (6 extractors + 9 constructors in 678 lines). Threshold logic not detailed. |
| KG-7 | `internal/lock/moirai.go` | **UNFILLED** -- 39 lines, likely trivial. |
| KG-8 | `internal/registry/registry.go` | **FILLED** -- leaf violation documented (DEBT-157), import analysis complete. 180 lines. |
| KG-9 | `config/hooks.yaml` | **UNFILLED** -- file exists (4.7K), content not examined. |
| KG-10 | `rites/` directory structure | **PARTIALLY FILLED** -- 18 rite directories discovered. Individual manifest schemas not cataloged. |

### LOAD-Bearing Code Verification

| LOAD ID | Status | Location verified |
|---------|--------|-------------------|
| LOAD-001 | **IN PLACE** | `internal/provenance/manifest.go:72` -- `structurallyEqual()` guard before write |
| LOAD-002 | **IN PLACE** | `internal/fileutil/fileutil.go:66` -- `WriteIfChanged()` atomic write with change detection |
| LOAD-003 | **IN PLACE** | `internal/inscription/merger.go:129` -- `MergeRegions()` satellite preservation |
| LOAD-004 | **IN PLACE** | `internal/materialize/mena/namespace.go:23` -- `resolveNamespace()` collision detection |
| LOAD-005 | **IN PLACE** | `internal/session/fsm.go:39` -- `ValidateTransition()` state machine enforcer |

All 5 LOAD-bearing code items confirmed present and structurally unchanged.

### DEBT-163: `.know/architecture.md` contains 3 stale claims

- **Category**: Documentation > Stale knowledge
- **Severity**: medium
- **Title**: Architecture knowledge file has incorrect leaf list, stale line count, and missing layer violations
- **Location**: `.know/architecture.md` lines 71, 67, 82-86
- **Impact**: Three specific inaccuracies: (1) `internal/registry` listed as leaf package but imports `frontmatter` and `mena` (DEBT-157). (2) TENSION-004 references 1,562-line `materialize.go` but file is now 732 lines after extraction of 5 stage files (`materialize_agents.go`, `materialize_claudemd.go`, `materialize_mena.go`, `materialize_rules.go`, `materialize_settings.go`). (3) Layer diagram does not document naxos and tribute upward violations (DEBT-159, DEBT-160). Additionally, `userscope/sync.go` was previously 1,530 lines but has been split into 7 files (2,716 lines total non-test).
- **Effort estimate**: 30 minutes to update `.know/architecture.md` with corrected data
- **Cross-reference**: DEBT-157, DEBT-159, DEBT-160, TENSION-004

---

## Surface 2: Monolith and Extraction Debt

### File Size Inventory (>500 lines, non-test)

| Rank | File | Lines | In original target list? |
|------|------|-------|--------------------------|
| 1 | `internal/cmd/lint/lint.go` | 784 | Yes |
| 2 | `internal/output/output.go` | 781 | Yes |
| 3 | `internal/inscription/pipeline.go` | 763 | Yes |
| 4 | `internal/materialize/materialize.go` | 732 | Yes |
| 5 | `internal/worktree/operations.go` | 707 | **No** (new finding) |
| 6 | `internal/output/rite.go` | 696 | Yes |
| 7 | `internal/sails/generator.go` | 678 | Yes |
| 8 | `internal/materialize/userscope/sync_mena.go` | 654 | **No** (new finding) |
| 9 | `internal/inscription/generator.go` | 647 | **No** (new finding) |
| 10 | `internal/hook/clewcontract/event.go` | 644 | **No** (new finding) |
| 11 | `internal/inscription/merger.go` | 638 | **No** (new finding) |
| 12 | `internal/cmd/hook/writeguard.go` | 588 | **No** (new finding) |
| 13 | `internal/cmd/validate/validate.go` | 577 | **No** (new finding) |
| 14 | `internal/sails/proofs.go` | 560 | **No** (new finding) |
| 15 | `internal/errors/errors.go` | 548 | **No** (new finding) |
| 16 | `internal/inscription/manifest.go` | 529 | **No** (new finding) |

11 files over 500 lines were NOT in the original target list. The original list missed `worktree/operations.go` entirely and underestimated the inscription package (3 files over 500 lines, not just `pipeline.go`).

### DEBT-164: `internal/cmd/lint/lint.go` -- 784 lines, 5 distinct lint domains in one file

- **Category**: Code > Monolith
- **Severity**: medium
- **Title**: Lint command mixes agent, dromena, legomena, namespace, and infrastructure linting in single file
- **Location**: `internal/cmd/lint/lint.go` (784 lines, 17 functions)
- **Impact**: Five distinct concern groups: (1) Agent linting (3 functions: `lintAgents`, `findAgentDirs`, `lintAgentFile`), (2) Dromena linting (2 functions: `lintDromena`, `lintDromenFile`), (3) Legomena linting (2 functions: `lintLegomena`, `lintLegomenFile`), (4) Namespace linting (1 function: `lintMenaNamespace`), (5) Infrastructure/utilities (7 functions: `checkSkillAtRefs`, `checkSourcePathLeaks`, `buildAllMenaSources`, `parseFrontmatterLenient`, etc.), (6) Command setup (2 functions: `NewLintCmd`, `runLint`). Internal coupling is moderate: all lint functions share the `LintReport` accumulator and `Finding` struct, but the domain-specific functions are otherwise independent. This file is also at 12.9% test coverage (DEBT-107), making the monolith structure a blocker for incremental test improvement.
- **Effort estimate**: 3-4 hours to split into `lint_agents.go`, `lint_dromena.go`, `lint_legomena.go`, `lint_namespace.go` plus shared types. Same package, no API change.
- **Split recommendation**: Clean split. Each lint domain is self-contained. Shared types (`LintReport`, `Finding`, `cmdContext`) stay in `lint.go`. Blast radius: low (same package, no external consumers).
- **Cross-reference**: DEBT-107 (lint coverage 12.9%), DEBT-146 (duplication patterns)

### DEBT-165: `internal/output/output.go` + `rite.go` -- 1,477 combined lines, 29 output structs

- **Category**: Code > Monolith (output type proliferation)
- **Severity**: medium
- **Title**: Output package accumulates all output types for every CLI command in 2 files
- **Location**: `internal/output/output.go` (781 lines, 16 output structs), `internal/output/rite.go` (696 lines, 13 output structs)
- **Impact**: Every new CLI command adds its output struct + `Text()` + `Headers()` + `Rows()` methods to one of these two files. The pattern is structurally sound (output has zero internal imports, acting as a pure formatting layer), but the growth is unbounded. The file already contains types for sessions, syncs, audits, frays, timelines, logs, snapshots, rite lists, rite switches, invocations, budgets, releases, swaps, and more. Adding a new command means editing a 700+ line file to add 30-50 lines, with merge conflict risk against parallel changes.
- **Effort estimate**: 4-6 hours to split into `output_session.go`, `output_sync.go`, `output_rite.go` (already partially done), `output_worktree.go`, etc.
- **Split recommendation**: Clean split by command domain. Each output struct is self-contained (no cross-type references). The `Printer` infrastructure (13 functions) stays in `output.go`. Blast radius: very low (same package, output types are consumed by `cmd/*` packages which import `output` regardless).
- **Cross-reference**: DEBT-161 (output has zero internal imports -- architecturally clean)

### DEBT-166: `internal/inscription/` -- 3 files over 500 lines (pipeline.go, generator.go, merger.go)

- **Category**: Code > Monolith (package-level)
- **Severity**: medium
- **Title**: Inscription package has 3 files over 500 lines totaling 2,048 lines of non-test code in those files alone
- **Location**:
  - `internal/inscription/pipeline.go` -- 763 lines (4 lifecycle ops, 3 render helpers, 6 utilities)
  - `internal/inscription/generator.go` -- 647 lines (section generation, template rendering, 13+ `getDefault*Content` methods)
  - `internal/inscription/merger.go` -- 638 lines (region merging, conflict detection, validation)
- **Impact**: The inscription package has 9,407 total lines (tests + source) across 15 files. The three largest non-test files each serve a distinct concern (orchestration, generation, merging), so the package is already partially split. However, `pipeline.go` contains both pipeline orchestration AND 6 utility functions (`simpleDiff`, `contains`, `firstSentence`, `truncate`, `parseVersion`, `extractFrontmatter`). `generator.go` has 13+ `getDefault*Content()` methods that could be a separate `defaults.go`. The coupling between these files is intentional: Pipeline calls Generator and Merger, Generator uses shared types from types.go, Merger operates on the Manifest. LOAD-003 (`MergeRegions`) is in merger.go.
- **Effort estimate**: 2-3 hours to extract `inscription/defaults.go` (default content methods) and `inscription/util.go` (utility functions from pipeline.go). Same package.
- **Split recommendation**: Moderate improvement. Extract defaults and utilities. The three core files (pipeline, generator, merger) represent genuine distinct concerns and should remain separate files. Blast radius: very low.
- **Cross-reference**: LOAD-003 (merger.go is load-bearing)

### DEBT-167: `internal/worktree/operations.go` -- 707 lines, mixes CRUD + sync + import/export + git ops

- **Category**: Code > Monolith
- **Severity**: medium
- **Title**: Worktree operations file combines 5 distinct concern groups
- **Location**: `internal/worktree/operations.go` (707 lines, 11 functions)
- **Impact**: Five concern groups: (1) CRUD operations (`Switch`, `Clone` -- 2 functions), (2) Sync (`Sync` -- 1 function), (3) Import/Export (`Export` at 125 lines, `Import` at 210 lines -- the two largest functions), (4) Git operations (`gitFetch`, `gitPull`, `detectConflicts` -- 3 functions), (5) Helpers (`resolveWorktree`, `copySessionContext`, `setupWorktreeEcosystem` -- 3 functions). The `Import` function (lines 358-567, ~210 lines) is the largest single function, handling tar.gz extraction, git worktree creation, session context copying, and ecosystem setup. The `Export` function (lines 232-357, ~125 lines) handles tar.gz creation with similar complexity. Both Import and Export are self-contained and could be extracted.
- **Effort estimate**: 2-3 hours to extract `worktree/import_export.go` (Import + Export functions). Same package.
- **Split recommendation**: Clean split for import/export. The CRUD and sync operations share `Manager` receiver and are tightly coupled. Blast radius: very low.
- **Cross-reference**: DEBT-162 (worktree imports materialize for ecosystem setup)

### DEBT-168: `internal/sails/generator.go` -- 678 lines, 6 text-parsing extractors embedded in generator

- **Category**: Code > Monolith
- **Severity**: low
- **Title**: Sails generator mixes YAML generation with 6 regex-based text parsing functions
- **Location**: `internal/sails/generator.go` (678 lines, 15 functions)
- **Impact**: Two distinct concern groups: (1) Generator infrastructure (9 functions: constructors, Generate, loadSessionContext, proofSetToColorProofs, generateYAML, etc.), (2) Text extractors (6 functions: `extractSessionType`, `extractOpenQuestions`, `extractBlockers`, `extractModifiers`, `extractQAUpgrade`, `extractSessionIDFromPath`). The extractors are stateless functions that parse SESSION_CONTEXT.md body text using regex patterns. They have no dependency on the Generator struct and could live in a separate `sails/extractors.go` file. The coupling between groups is low: Generate calls loadSessionContext, which calls the extractors.
- **Effort estimate**: 1 hour to extract `sails/extractors.go`. Same package.
- **Split recommendation**: Clean split. Extractors are pure functions. Blast radius: none.

### DEBT-169: `internal/hook/clewcontract/event.go` -- 644 lines, 25 event constructors

- **Category**: Code > Monolith (growth pattern)
- **Severity**: low
- **Title**: Clew contract event file grows linearly with each new event type
- **Location**: `internal/hook/clewcontract/event.go` (644 lines, 25 `New*Event` constructors, 4 struct types)
- **Impact**: Every new event type adds ~20 lines to this file (constructor + metadata setup). The file has grown from the original session lifecycle events to include tool calls, file changes, decisions, context switches, sails, tasks, artifacts, errors, handoffs, stamps, and more. All constructors follow the same pattern (create Event, set fields, return). The file is not complex -- it is repetitive. There is no coupling between constructors. The 4 struct types (`SailsGeneratedData`, `Stamp`, `QualityProof`, `ArtifactType`) are interleaved with the constructors.
- **Effort estimate**: 1-2 hours to split by event domain (e.g., `event_session.go`, `event_task.go`, `event_sails.go`). Same package.
- **Split recommendation**: Optional improvement. The file is not complex, just long. The linear growth pattern means it will reach 800+ lines with the next batch of event types. Blast radius: none.
- **Cross-reference**: DEBT-126 (event format bridge in events_read.go)

### DEBT-170: TENSION-004 partially resolved -- materialize.go down from 1,562 to 732 lines

- **Category**: Code > Monolith (progress observation)
- **Severity**: medium
- **Title**: Materialize monolith extraction 53% complete by line count
- **Location**: `internal/materialize/materialize.go` (732 lines)
- **Impact**: Five stage files have been extracted from the original monolith: `materialize_agents.go` (399 lines), `materialize_settings.go` (389 lines), `materialize_claudemd.go` (156 lines), `materialize_mena.go` (149 lines), `materialize_rules.go` (141 lines). The parent package (non-test, non-subpackage) is now 3,387 lines across 22 files, with the largest file being `materialize.go` at 732 lines. The remaining 732 lines contain: types (RiteManifest struct, MCP helpers), constructors (`NewMaterializer`, `WithEmbedded*`), core orchestration (`MaterializeWithOptions`, `MaterializeMinimal`), sync dispatch (`Sync`, `syncRiteScope`), and rite manifest loading. These are the orchestration glue that ties the stages together. Further extraction is possible (`materialize_sync.go` for sync dispatch, `materialize_types.go` for type definitions) but the orchestration functions reference each other heavily, limiting clean extraction.
- **Effort estimate**: 2-3 hours for further type/sync extraction; diminishing returns below 500 lines
- **Split recommendation**: The current 732 lines is within acceptable range. Further splitting should target types.go extraction (RiteManifest, ~80 lines) and sync.go extraction (Sync/syncRiteScope, ~150 lines). Blast radius: low (same package).
- **Cross-reference**: TENSION-004 in `.know/design-constraints.md`, DEBT-163 (stale line count documentation)

### DEBT-171: `internal/materialize/userscope/sync_mena.go` -- 654 lines after split (was part of 1,530-line sync.go)

- **Category**: Code > Monolith (post-split assessment)
- **Severity**: medium
- **Title**: Userscope mena sync is the largest file after successful split of the original monolith
- **Location**: `internal/materialize/userscope/sync_mena.go` (654 lines, 6 functions)
- **Impact**: The original `userscope/sync.go` (1,530 lines, documented in MEMORY.md) was split into 7 files totaling 2,716 non-test lines: `sync.go` (462), `sync_mena.go` (654), `sync_agents.go` (208), `sync_cleanup.go` (247), `collision.go` (69), `worktree.go` (37), `types.go` (113). The split was successful -- the MEMORY.md hotspot note is now stale. However, `sync_mena.go` at 654 lines contains 3 parallel sync paths: `syncUserMena` (filesystem sources, 182 lines), `syncUserMenaFromEmbedded` (embedded FS sources, 148 lines), and `wipeKnossosOwnedMenaEntries` (cleanup, 88 lines). The first two functions have significant structural overlap (DEBT-158 sibling coupling). Additionally, 16 `checksum.File()` error discards in sync.go (DEBT-138 highest-risk subset) likely persist in this split file.
- **Effort estimate**: 1-2 days to unify the two sync paths behind a common interface (mirrors the `copyDirFS` unification in mena/walker.go)
- **Split recommendation**: Unify `syncUserMena` and `syncUserMenaFromEmbedded` -- they are structural duplicates differing only in source type (filesystem vs embedded). The `copyDirFS` unification (DEBT-152, resolved) provides a proven pattern. Blast radius: medium (single file, but changes sync behavior).
- **Cross-reference**: DEBT-138 (error discards), DEBT-152 (copyDirFS unification pattern), DEBT-158 (sibling coupling)

### DEBT-172: `internal/cmd/hook/writeguard.go` -- 588 lines, 20+ output helper functions

- **Category**: Code > Monolith
- **Severity**: low
- **Title**: Writeguard hook mixes core decision logic with 8+ output formatting functions
- **Location**: `internal/cmd/hook/writeguard.go` (588 lines, 27 functions)
- **Impact**: The file contains: (1) Core writeguard logic (`runWriteguard`, `runWriteguardCore`, `classifyEditSection` -- ~200 lines of actual decision-making), (2) Path parsing helpers (`parseFilePath`, `parseOldString`, `parseContentField` -- 3 functions), (3) Section classification predicates (`isTimelineIndicator`, `isFrontmatterIndicator`, `isOtherSectionIndicator`, `isProtectedFile`, `isSessionContext`, `isWipPath` -- 6 functions), (4) Output formatting (`outputAllow`, `outputBlock`, `outputBlockArchived`, `outputBlockFrontmatter`, `outputBlockOtherSection`, `outputBlockMixed`, `outputBlockUnknown`, `outputAllowWithContext`, `outputAllowTimeline` -- 9 functions). The 9 output functions follow a repetitive pattern (construct JSON response, print). These could be collapsed into a generic `outputDecision(printer, decision, message)` function.
- **Effort estimate**: 1-2 hours. Extract a generic output helper, collapse 9 functions into 1.
- **Split recommendation**: Refactor, not split. The output functions should be collapsed into a single parameterized function. Blast radius: low (same file).
- **Cross-reference**: DEBT-149 (hook handler divergence)

---

## Cross-References

| DEBT ID | Source Reference | Category |
|---------|-----------------|----------|
| DEBT-100 | `.know/test-coverage.md` Priority 4 | Test Coverage |
| DEBT-101 | SCAR-017, SCAR-019 | Test Coverage |
| DEBT-112 | MEMORY.md hotspot (userscope/sync.go) | Test Coverage |
| DEBT-113 | RISK-003, D3 | Test Coverage |
| DEBT-114 | `.know/test-coverage.md` aggregate | Test Coverage |
| DEBT-115 | `.know/test-coverage.md` (stale) | Documentation |
| DEBT-116 | ADR-0011, SCAR-015, SCAR-016 | Shell Residue |
| DEBT-117 | ADR-0011, SCAR-016, TDD-cross-rite-handoff | Shell Residue |
| DEBT-118 | ADR-0011 (out of scope), SCOUT-e2e-distribution | Shell Residue |
| DEBT-120 | DEBT-116 (dead references) | Documentation |
| DEBT-121 | CC Agent Uplift Wave 3 | Deferred Work |
| DEBT-122 | CC Agent Uplift rollout exception | Deferred Work |
| DEBT-123 | MEMORY.md "ghost skills" (incorrect) | Deferred Work |
| DEBT-124 | CC Agent Uplift | Documentation |
| DEBT-125 | DEBT-039, `.know/design-constraints.md` KG-1 | Dead Code |
| DEBT-126 | TENSION-008, ADR-0027 | Migration Debt |
| DEBT-127 | DEBT-116, DEBT-117, DEBT-118, MEMORY.md P1 | Initiative Tracking |
| DEBT-128 | ADR-0011, MEMORY.md P2 | Initiative Tracking |
| DEBT-129 | MEMORY.md P3 | Initiative Tracking |
| DEBT-130 | DEBT-125, DEBT-039 | Dead Infrastructure |
| DEBT-131 | `.know/scar-tissue.md` Defensive Patterns | Regression Gaps |
| DEBT-132 | `.know/test-coverage.md` | Test Coverage |
| DEBT-133 | `.know/test-coverage.md` | Test Coverage |
| DEBT-134 | `.know/test-coverage.md` | Test Coverage |
| DEBT-135 | `.know/test-coverage.md` | Test Coverage |
| DEBT-136 | `.know/test-coverage.md` | Test Coverage |
| DEBT-137 | DEBT-123 | Documentation |
| DEBT-138 | RISK-001/004/005, SCAR-004 | Silent Failure |
| DEBT-139 | D4 from SPIKE | Observability |
| DEBT-140 | D4 from SPIKE | Silent Failure |
| DEBT-141 | SCAR-010, hook handler analysis | Observation |
| DEBT-142 | DEBT-141 | Duplication |
| DEBT-143 | RISK-004 (partially resolved) | Silent Failure |
| DEBT-144 | DEBT-139 | Observability |
| DEBT-145 | `.know/design-constraints.md` RISK-001/004/005 | Documentation |
| DEBT-146 | TENSION-006 (resolved) | Duplication |
| DEBT-147 | D1 from SPIKE | Duplication |
| DEBT-148 | `.know/design-constraints.md` TENSION-006 | Resolved |
| DEBT-149 | SCAR-010, DEBT-142 | Duplication |
| DEBT-150 | DEBT-134 | Test Duplication |
| DEBT-151 | D2 from SPIKE | Duplication |
| DEBT-152 | `.know/design-constraints.md` Abstraction Gap | Resolved |
| DEBT-153 | `.know/design-constraints.md` TENSION-001 | Naming |
| DEBT-154 | Go naming conventions | Naming |
| DEBT-155 | `.know/design-constraints.md` TENSION-002, ADR-0023 | Deprecated Compat |
| DEBT-156 | `.know/design-constraints.md` TENSION-007/008 | Naming (stable) |
| DEBT-157 | `.know/architecture.md` line 71, KG-8 | Leaf violation |
| DEBT-158 | DEBT-146, DEBT-151, DEBT-152 | Sub-package coupling |
| DEBT-159 | `.know/architecture.md` layer diagram | Layer violation |
| DEBT-160 | DEBT-159, `.know/architecture.md` layer diagram | Layer violation |
| DEBT-161 | `.know/architecture.md` hub packages | Observation |
| DEBT-162 | `.know/architecture.md` import patterns | Cross-domain coupling |
| DEBT-163 | DEBT-157, DEBT-159, DEBT-160, TENSION-004 | Stale knowledge |
| DEBT-164 | DEBT-107, DEBT-146 | Monolith |
| DEBT-165 | DEBT-161 | Monolith |
| DEBT-166 | LOAD-003 | Monolith |
| DEBT-167 | DEBT-162 | Monolith |
| DEBT-168 | N/A | Monolith |
| DEBT-169 | DEBT-126 | Monolith (growth) |
| DEBT-170 | TENSION-004, DEBT-163 | Monolith (progress) |
| DEBT-171 | DEBT-138, DEBT-152, DEBT-158 | Monolith (post-split) |
| DEBT-172 | DEBT-149 | Monolith (refactor) |

---

## Coverage Data

### Package-Level Statement Coverage (from `go test -coverprofile`)

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/errors` | 100.0% | CORRECTED (was documented as 0%) |
| `internal/mena` | 96.2% | Excellent |
| `internal/frontmatter` | 93.3% | Excellent |
| `internal/tokenizer` | 91.8% | Excellent |
| `internal/registry` | 92.6% | Excellent |
| `internal/hook` | 90.3% | Excellent |
| `internal/artifact` | 88.0% | Good |
| `internal/materialize/hooks` | 88.5% | Good |
| `internal/agent` | 87.6% | Good |
| `internal/paths` | 86.9% | Good |
| `internal/session` | 84.2% | Good |
| `internal/checksum` | 84.2% | Good |
| `internal/inscription` | 83.2% | Good |
| `internal/cmd/initialize` | 81.3% | Good |
| `internal/hook/clewcontract` | 81.2% | Good |
| `internal/materialize` | 80.2% | Good |
| `internal/materialize/orgscope` | 80.3% | Good |
| `internal/naxos` | 80.0% | Good |
| `internal/cmd/tour` | 79.7% | Good |
| `internal/sails` | 79.3% | Good |
| `internal/tribute` | 79.1% | Good |
| `internal/lock` | 77.8% | Good |
| `internal/cmd/status` | 76.8% | Good |
| `internal/provenance` | 74.7% | Adequate |
| `internal/validation` | 74.1% | Adequate |
| `internal/materialize/mena` | 72.7% | Adequate |
| `internal/materialize/source` | 71.8% | Adequate |
| `internal/know` | 70.6% | Adequate |
| `internal/sync` | 70.3% | Adequate |
| `internal/cmd/hook` | 69.5% | Adequate |
| `internal/worktree` | 67.2% | Adequate |
| `internal/fileutil` | 66.7% | Adequate |
| `internal/cmd/handoff` | 66.1% | Adequate |
| `internal/cmd/session` | 59.0% | Below threshold |
| `internal/cmd/validate` | 59.2% | CORRECTED (was documented as 0%) |
| `internal/cmd/worktree` | 52.1% | Below threshold |
| `internal/cmd/agent` | 51.5% | Below threshold |
| `internal/rite` | 47.6% | Below threshold |
| `internal/cmd/sync` | 47.2% | CORRECTED (was documented as 0%) |
| `internal/cmd/knows` | 44.2% | Below threshold |
| `internal/manifest` | 44.3% | Below threshold |
| `internal/cmd/explain` | 42.6% | Below threshold |
| `internal/cmd/sails` | 39.5% | Below threshold |
| `internal/config` | 34.6% | Low |
| `internal/materialize/userscope` | 23.7% | Low |
| `internal/cmd/rite` | 20.5% | Low |
| `internal/cmd/lint` | 12.9% | Very Low |
| `internal/output` | 11.7% | Very Low |
| `internal/cmd/org` | 1.2% | Negligible |
| `internal/assets` | 0.0% | Zero (embedded accessors) |
| `internal/cmd/artifact` | 0.0% | Zero |
| `internal/cmd/common` | 0.0% | Zero |
| `internal/cmd/inscription` | 0.0% | Zero |
| `internal/cmd/manifest` | 0.0% | Zero |
| `internal/cmd/naxos` | 0.0% | Zero |
| `internal/cmd/provenance` | 0.0% | Zero |
| `internal/cmd/root` | 0.0% | Zero |
| `internal/cmd/tribute` | 0.0% | Zero |
| `cmd/ari` | 0.0% | Zero (main entry) |
| `test/hooks/testutil` | 0.0% | Zero (test utility) |
| `test/worktree/testutil` | 0.0% | Zero (test utility) |

**Overall**: 61.3% statement coverage, 1,754 functions, 502 at 0.0%

### Function Coverage Distribution

| Coverage Band | Function Count | Percentage |
|---------------|---------------|------------|
| 0% | 502 | 28.6% |
| 1-49% | 30 | 1.7% |
| 50-79% | 192 | 10.9% |
| 80-99% | 327 | 18.6% |
| 100% | 703 | 40.1% |

The distribution is bimodal: functions are either fully covered (40.1%) or completely uncovered (28.6%). The "partially covered" middle ground is relatively small (31.2%), confirming that the test investment follows an all-or-nothing pattern per function.

---

## Audit Limitations

1. **Line-level coverage only measures statement execution**: A function at 100% statement coverage may still have untested logical branches (e.g., short-circuit evaluation, error paths within compound expressions).

2. **Indirect coverage not measured**: `internal/errors` at 100% likely gets much of its coverage indirectly from callers' tests. The coverage tool attributes this correctly but the quality of those indirect tests (assertion depth, edge case coverage) is not assessed.

3. **Shell script callers traced by text search only**: The conclusion that `context-injection.sh` is dead code is based on zero references in YAML/JSON config files and the non-existence of `user-hooks/`. A runtime trace (e.g., running `ari sync` with strace) would provide stronger evidence.

4. **Deferred item staleness is assessed against current code state**: Items marked "VALID" mean the code context has not changed. They do NOT mean the item is still worth doing -- that is the Risk Assessor's determination.

5. **Test-only code at 0%**: `test/hooks/testutil` and `test/worktree/testutil` appear at 0% because they are test helper packages not directly tested. This is expected and not debt.

6. **All 8 surfaces audited**: Collection phase complete. Surfaces 1-8 covered across 3 passes.

7. **Import graph is static analysis only**: Layer violations (DEBT-159, DEBT-160) are based on `go list -json` output. Runtime call patterns (e.g., whether naxos actually exercises sails code at runtime) are not assessed.

8. **Monolith split recommendations are estimates**: Blast radius and coupling assessments for file splits (Surface 2) are based on function signature analysis and grep-based cross-reference. Actual split difficulty depends on shared local variables and closures not visible at the function boundary level.

9. **Satellite manifest audit boundary**: The assessment of deprecated `Commands`/`Skills` fields (DEBT-155) covers only manifests within the knossos monorepo. Satellite project manifests (external repositories using knossos) were not scanned.

10. **Error-discard risk categorization is judgment-based**: The categorization of 97 `_ =`/`_ :=` sites into risk tiers (correctness, integrity, debugging, intentional) is based on static analysis of what the discarded function returns. Actual impact depends on runtime conditions.

11. **Hook handler timeout analysis**: The observation that `cheapo_revert` and `worktreeremove` lack `withTimeout()` is structural. Whether this causes actual timeouts depends on CC's hook timeout enforcement at the process level.

---

## Collection Phase Summary

### Totals

| Metric | Count |
|--------|-------|
| **Total debt items** | **73** |
| Passes | 3 (Pass 1: Surfaces 1,4,8; Pass 2: Surfaces 6,3,5; Pass 3: Surfaces 7,2) |
| Items flagged for assessor | 2 (DEBT-159, DEBT-160) |
| Resolved items documented | 2 (DEBT-148, DEBT-152) |

### Breakdown by Surface

| Surface | Items | Severity Distribution |
|---------|-------|----------------------|
| 1. Test Coverage | 16 | 3 high, 8 medium, 5 low |
| 2. Monolith Extraction | 9 | 0 high, 6 medium, 3 low |
| 3. Duplication | 7 | 0 high, 3 medium, 4 low |
| 4. Shell Scripts | 5 | 0 high, 2 medium, 3 low |
| 5. Naming/Schema | 4 | 0 high, 2 medium, 2 low |
| 6. Observability | 8 | 1 high, 4 medium, 3 low |
| 7. Architectural Boundaries | 7 | 2 high, 4 medium, 1 low |
| 8. Deferred Work | 17 | 1 high, 10 medium, 6 low |
| **Total** | **73** | **7 high, 39 medium, 27 low** |

### Breakdown by Category

| Category | Count |
|----------|-------|
| Architecture (layer violations, coupling, boundaries) | 7 |
| Code > Monolith (file splits, concern mixing) | 9 |
| Code > Duplication (DRY violations) | 7 |
| Testing (coverage, regression, duplication) | 18 |
| Design (naming, schema, deprecated compat) | 4 |
| Documentation (stale knowledge, missing docs) | 5 |
| Observability (silent failures, logging) | 8 |
| Deferred work (tracked initiatives) | 15 |
| Resolved (confirmed fixed) | 2 |

### Provenance

| Source | Count | Description |
|--------|-------|-------------|
| Carried from existing intelligence | 38 | Items originating from `.know/`, MEMORY.md, SCAR tissue, prior ledger |
| Newly discovered | 29 | Items found through fresh analysis (coverage profiling, import graph, file inventory) |
| Corrected vs prior intelligence | 8 | Items where prior documentation was inaccurate (stale coverage numbers, incorrect leaf packages, stale line counts) |

### Collection Phase Status

**COMPLETE**. All 8 audit surfaces systematically covered. The ledger is ready for risk assessment.

The Debt Collector handoff criteria are met:
- [x] All in-scope areas systematically audited (8/8 surfaces)
- [x] Each debt item has location, category, and description
- [x] Duplicates and overlapping items consolidated
- [x] Summary statistics accurate and complete
- [x] Obvious severity items flagged for priority attention (DEBT-159, DEBT-160)
- [x] Audit limitations and gaps documented (11 items)
- [x] All artifacts verified via Read tool

**Next step**: Route to Risk Assessor for scoring and prioritization.

---

## Radar-Sourced Items (OPP Cross-Reference)

Items discovered by the Knowledge Radar scan (`.know/radar.md`, 2026-03-03) that have no corresponding DEBT item in the original ledger. These extend the inventory with convention-drift and systemic-pattern findings that static coverage and structural analysis missed.

### DEBT-173: os.Stdout bypass in 41 cmd/ sites

- **Category**: Code > Convention drift
- **Severity**: high
- **Title**: 41 `fmt.Fprintf(os.Stdout, ...)` sites bypass the Printer abstraction
- **Location**: `internal/cmd/agent/validate.go` (18 sites), `internal/cmd/agent/update.go` (8), `internal/cmd/agent/list.go`, `internal/cmd/session/gc.go` (10, no Printer at all)
- **Impact**: `--output=json` silently omits results from any command using direct stdout writes. Agent validate creates a Printer then ignores it -- the most egregious case. Session gc never creates one. This means `ari agent validate --output=json` returns incomplete JSON, and `ari session gc --output=json` returns nothing.
- **Effort estimate**: 2-3 days (41 sites across 4+ files, each needs structured output type)
- **Cross-reference**: OPP-001, DEBT-100 (output coverage), `.know/conventions.md` Printer convention

### DEBT-174: fmt.Errorf at CLI boundaries (39 violations)

- **Category**: Code > Convention drift
- **Severity**: high
- **Title**: 39 `fmt.Errorf` returns in RunE handlers lose exit code control and JSON formatting
- **Location**: `internal/cmd/org/init.go` (6), `internal/cmd/org/set.go` (3), `internal/cmd/knows/knows.go` (5), `internal/cmd/sync/sync.go` (3), `internal/cmd/worktree/sync.go:111` (1), additional files
- **Impact**: RunE handlers returning `fmt.Errorf` bypass `PrintError` and structured error output. In `--output=json` mode, these errors appear as unstructured text to stderr instead of JSON error objects. Exit codes default to 1 instead of domain-specific codes. `cmd/org/` (8 violations) is the worst offender.
- **Effort estimate**: 1-2 days (mechanical refactoring per-package)
- **Cross-reference**: OPP-002, `.know/conventions.md` error handling section

### DEBT-175: Non-atomic writes on critical state files (59 sites)

- **Category**: Code > Data integrity
- **Severity**: high
- **Title**: 59 `os.WriteFile` sites on state files read every ari invocation; crash during write corrupts with no recovery
- **Location**: `internal/rite/state.go:93` (invocation state), `internal/worktree/metadata.go:78,90,256` (worktree registry), `internal/manifest/manifest.go:191` (manifest save), `internal/artifact/registry.go:173,247` (artifact registry)
- **Impact**: `rite/state.go:93` and `worktree/metadata.go` are highest risk -- both write YAML state read on every `ari` invocation. A crash, power loss, or signal during write leaves a truncated file that fails to parse on next invocation, breaking all ari commands until manual repair. Only 13 files currently use `fileutil.AtomicWriteFile` correctly. The fix is mechanical: replace `os.WriteFile` with `fileutil.AtomicWriteFile`.
- **Effort estimate**: 1 day for the 4 highest-risk files; 2-3 days for all 59 sites
- **Cross-reference**: OPP-007, SCAR-003 (idempotency), DEBT-146 (non-atomic session copyDir)

### DEBT-176: KnossosHome cache poisoning in 8 test functions

- **Category**: Testing > Test isolation
- **Severity**: medium
- **Title**: 8 test functions in unified_sync_test.go set KNOSSOS_HOME without proper cleanup
- **Location**: `internal/materialize/unified_sync_test.go` (8 functions)
- **Impact**: Missing `t.Cleanup(config.ResetKnossosHome)` means the `sync.Once`-cached KnossosHome value can leak between tests. If test execution order changes (Go 1.24 shuffled by default), tests that depend on a specific KnossosHome may see a stale value from a prior test. The correct pattern exists in `internal/cmd/hook/context_test.go`.
- **Effort estimate**: 30 minutes (mechanical: add `t.Setenv` + `t.Cleanup` per function)
- **Cross-reference**: OPP-011, RISK-003 (sync.Once caching hazard), DEBT-113 (config coverage)

### DEBT-177: Schema evolution systemic pattern (3 SCARs)

- **Category**: Architecture > Systemic pattern
- **Severity**: medium
- **Title**: SCAR-011/014/016 share a "schema changed, consumers not updated atomically" root cause
- **Location**: Session state management (SCAR-011 writeguard, SCAR-014 phantom status, SCAR-016 bash arithmetic)
- **Impact**: Schema/contract evolution without atomic consumer updates is a recurring failure mode. `NormalizeStatus()` alias map is the correct defensive pattern. Partially overlaps DEBT-126 (event bridge) but adds a systemic lens: the codebase lacks a canonical schema registry test that would catch undeclared status values or format changes.
- **Effort estimate**: 1 day (schema registry test + audit of remaining shell scripts for `set -e` traps)
- **Cross-reference**: OPP-012, DEBT-126, SCAR-011, SCAR-014, SCAR-016

### DEBT-178: Data corruption systemic pattern (3 SCARs)

- **Category**: Architecture > Systemic pattern
- **Severity**: medium
- **Title**: SCAR-004/015/022 share a manifest serialization boundary failure root cause
- **Location**: Manifest serialization/deserialization boundary (provenance, shell log stdout, checksum format)
- **Impact**: All three SCARs failed at the point where structured data crosses a serialization boundary. SCAR-004 silently discarded filesystem errors during manifest load. SCAR-015 mixed log output with manifest JSON on stdout. SCAR-022 used abbreviated SHA256 that failed schema validation. The codebase lacks load-time manifest schema validation -- manifests are parsed optimistically with no structural integrity check.
- **Effort estimate**: 1-2 days (add manifest schema validation at load time + provenance checksum constructor function)
- **Cross-reference**: OPP-013, DEBT-131, DEBT-138, SCAR-004, SCAR-015, SCAR-022

### DEBT-179: Historical boundary violations caught by revert not enforcement (3 SCARs)

- **Category**: Architecture > Systemic pattern
- **Severity**: medium
- **Title**: SCAR-013/026/027 architectural boundary violations discovered by revert, not by automated enforcement
- **Location**: Session artifacts in shared mena (SCAR-027), writeguard/Moirai coupling (SCAR-026), ghost dirs (SCAR-013)
- **Impact**: These three SCARs were caught because someone noticed the problem and reverted, not because an automated check flagged it. An `ari lint` rule detecting session artifact filenames in `rites/shared/mena/` would catch SCAR-027-class issues automatically. The writeguard coupling constraint (SCAR-026) should be documented in `.know/design-constraints.md`.
- **Effort estimate**: 2-3 hours (new lint rule + design-constraints update)
- **Cross-reference**: OPP-014, DEBT-164 (lint.go monolith), SCAR-013, SCAR-026, SCAR-027

### DEBT-180: Testify drift beyond documented scope

- **Category**: Documentation > Convention staleness
- **Severity**: low
- **Title**: 23 files use testify vs documented 18; freeze line is stale
- **Location**: `.know/conventions.md` testify section; additional files in tour, explain, tokenizer, materialize/hooks/mcp, agent/mcp_validate
- **Impact**: Convention doc says 18 files use testify with an implied freeze. Actual count is 23. 5 files adopted testify outside the documented scope. No runtime risk -- this is documentation staleness only. New test files should continue using stdlib testing.
- **Effort estimate**: 15 minutes (update conventions.md with correct count)
- **Cross-reference**: OPP-018, `.know/conventions.md`
