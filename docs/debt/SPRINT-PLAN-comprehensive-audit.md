---
audit_id: session-20260302-232344-1b73b3a8
plan_date: "2026-03-03"
total_items: 81
workstreams: 5
sprints: 12
estimated_duration: "19-28 days"
source_matrix: RISK-REPORT-comprehensive-audit.md
source_ledger: LEDGER-comprehensive-audit.md
status: ready
schema_version: "1.0"
capacity:
  assumed_velocity: "6-8 hours/day effective"
  buffer_percent: 20
  sprint_length: "varies by workstream (1 day to 8 days)"
---

# Debt Remediation Sprint Plan

## Executive Summary

This plan transforms 81 scored debt items into 12 sprint-ready packages across 5 workstreams plus a Sprint 0 quick-wins pass. The risk assessment identified 4 Critical, 4 High, 30 Medium, and 40 Low items with 3 Resolved. Total remediation portfolio is 19-28 engineering days.

**Strategic approach**: Start with Sprint 0 (quick wins, 5 hours) and WS4 (knowledge refresh, 2-3 hours) to clear documentation debt and establish correct intelligence baseline. Then attack the convergent hotspot (WS1, materialize pipeline) before moving to SCAR regression safety (WS3), CLI conventions (WS2), and deep structural remediation (WS5).

**Key packaging decisions**:
- 11 quick wins extracted to Sprint 0 (all sub-1-hour, 5h total)
- WS1 split into two sprints: critical fixes first, then hook hardening
- WS2 split into three sprints: test harness, stdout routing, error formatting
- WS3 split into two sprints: systemic fixes, then individual SCAR tests
- WS5 split into two sprints: coverage first, then structural unification
- 3 Resolved items (DEBT-145, DEBT-148, DEBT-152) excluded from all packages
- 2 items requiring user input (DEBT-129, DEBT-126) documented in Deferred section
- Aggregate items (DEBT-114, DEBT-119) tracked as portfolio metrics, not sprint tasks

---

## Sprint 0: Quick Wins (Day 0)

**Goal**: Clear all sub-1-hour items in a single focused session. Resolves the entire Documentation Accuracy Cluster and provides immediate ROI.

**Total effort**: 5 hours
**Confidence**: High (all items are mechanical, well-scoped changes)
**Rite**: None required (knowledge maintenance + small code fixes)
**Dependencies**: None. Start here.

### PKG-000a: Documentation Accuracy Sweep

**Priority**: High (cluster resolution)
**Size**: XS (1-2h total for all 5 items)
**Confidence**: High

**Source Items**:
- DEBT-115: Regenerate `.know/test-coverage.md` (3 packages documented as zero that are now tested)
- DEBT-163: Fix `.know/architecture.md` 3 stale claims (leaf list, line counts, layer violations)
- DEBT-145: Update `.know/design-constraints.md` resolved RISKs (RISK-001, RISK-004, RISK-005)
- DEBT-137: Correct MEMORY.md ghost skills note (10x-dev dromena are functional, not ghosts)
- DEBT-157: Correct `.know/architecture.md` leaf list (overlaps with DEBT-163; registry is not a leaf)

**Task Breakdown**:
1. Run `/know --force test-coverage` to regenerate test-coverage.md with current coverage data
2. Edit `.know/architecture.md`: (a) remove `internal/registry` from leaf package list and add note about frontmatter+mena imports, (b) update TENSION-004 materialize.go line count from 1,562 to 732 (5 stage files extracted), (c) add naxos and tribute layer violations to layer diagram notes
3. Edit `.know/design-constraints.md`: mark RISK-001 as RESOLVED (transformAgentContent returns errors), mark RISK-005 as RESOLVED (LoadOrBootstrap propagates errors), update RISK-004 as PARTIALLY RESOLVED (CleanEmptyDirs returns errors but stale-entry removal in DEBT-143 remains)
4. Edit MEMORY.md: change "10x-dev ghost skills" deferred note to "10x-dev dromena confirmed functional (10x-ref, architect-ref, build-ref all have INDEX.dro.md files)"
5. DEBT-157 is subsumed by DEBT-163 step 2a -- no separate action needed

**Acceptance Criteria**:
- [ ] `.know/test-coverage.md` shows `internal/errors` at 100%, `internal/cmd/validate` at 59.2%, `internal/cmd/sync` at 47.2%
- [ ] `.know/architecture.md` leaf list does not include `internal/registry`
- [ ] `.know/architecture.md` TENSION-004 references 732-line materialize.go (not 1,562)
- [ ] `.know/design-constraints.md` RISK-001 and RISK-005 marked RESOLVED
- [ ] MEMORY.md "Deferred" section no longer says "ghost skills" for 10x-dev

---

### PKG-000b: Dead Shell Removal

**Priority**: Medium
**Size**: S (2h)
**Confidence**: High

**Source Items**:
- DEBT-116: Remove dead `rites/ecosystem/context-injection.sh` (80 lines, zero runtime callers)
- DEBT-120: Clean 37 documentation references to dead context-injection.sh call chain

**Task Breakdown**:
1. Delete `rites/ecosystem/context-injection.sh`
2. Search all docs for references to `context-injection.sh`, `rite-context-loader.sh`, and `session-context.sh` call chain
3. For each reference: add deprecation notice or remove the reference depending on context (PRDs get deprecation notice, implementation docs get removal)
4. Update MEMORY.md "Current Priorities" item 1 (shell script deep cleanse) and item 2 (hook bash elimination) to reflect that context-injection.sh is removed and bash hook elimination is effectively complete
5. Close DEBT-127 (shell cleanse partially complete) and DEBT-128 (hook bash elimination nearly done) as resolved by this action

**Acceptance Criteria**:
- [ ] `rites/ecosystem/context-injection.sh` does not exist on disk
- [ ] `grep -r "context-injection" docs/` returns zero non-deprecation-notice hits
- [ ] `grep -r "rite-context-loader" docs/` returns zero non-deprecation-notice hits
- [ ] MEMORY.md shell cleanse and hook bash priorities updated

---

### PKG-000c: Code Quick Fixes

**Priority**: Medium
**Size**: XS (1h)
**Confidence**: High

**Source Items**:
- DEBT-143: Collect engine.go stale-entry `os.RemoveAll`/`os.Remove` errors into `result.Warnings`
- DEBT-140: Add `log.Printf` on each of 6 silent error paths in `extractEmbeddedMenaToXDG`
- DEBT-176: Add `t.Cleanup` to 8 test functions that mutate KnossosHome cache (sync.Once poisoning)
- DEBT-180: Update `.know/conventions.md` testify count from 18 to 23 packages

**Task Breakdown**:
1. `internal/materialize/mena/engine.go:263-272`: Wrap `os.RemoveAll` and `os.Remove` in error checks; append non-nil errors to `result.Warnings` with context (file path, operation)
2. `internal/materialize/mena/engine.go:88-96`: Same treatment for dromena INDEX cleanup block
3. `internal/cmd/initialize/init.go:275-321`: Add `log.Printf("Warning: extractEmbeddedMena: %s failed: %v", operation, err)` for each of: RemoveAll, MkdirAll, ReadFile, WriteFile, WalkDir skip, sentinel-write
4. Locate the 8 test functions with KnossosHome cache poisoning (likely in `internal/config/` tests) and add `t.Cleanup(func() { /* reset sync.Once */ })` or use `t.Setenv` pattern
5. Edit `.know/conventions.md`: update testify package count from 18 to 23

**Acceptance Criteria**:
- [ ] `engine.go` stale-entry removal block appends errors to `result.Warnings` instead of discarding
- [ ] `extractEmbeddedMenaToXDG` has a `log.Printf` on every error path
- [ ] Test functions with KnossosHome mutations have `t.Cleanup` guards
- [ ] `.know/conventions.md` testify count matches reality (23)
- [ ] `CGO_ENABLED=0 go test ./...` passes

---

### PKG-000d: SCAR Boundary Lint Rule

**Priority**: Medium
**Size**: S (2-3h)
**Confidence**: Medium (depends on lint.go structure; add 25% buffer)

**Source Items**:
- DEBT-179: Add `ari lint` rule that flags session-specific artifacts in shared mena directories

**Task Breakdown**:
1. Read `internal/cmd/lint/lint.go` to understand rule registration pattern
2. Add a new lint function `lintSessionArtifactsInSharedMena()` that walks shared mena directories and flags files with session-specific frontmatter (session_id, throughline references)
3. Register the new rule in the existing lint pipeline
4. Add test case in `internal/cmd/lint/lint_test.go`

**Acceptance Criteria**:
- [ ] `ari lint` reports findings when session artifacts appear in shared mena
- [ ] New lint rule has at least one test case
- [ ] Existing lint tests pass

---

**Sprint 0 Summary**:

| Package | Items | Effort | Confidence |
|---------|-------|--------|------------|
| PKG-000a | DEBT-115, 163, 145, 137, 157 | 1-2h | High |
| PKG-000b | DEBT-116, 120 | 2h | High |
| PKG-000c | DEBT-143, 140, 176, 180 | 1h | High |
| PKG-000d | DEBT-179 | 2-3h | Medium |
| **Total** | **11 items** | **~5h** | **High** |

Also closes transitively: DEBT-127 (shell cleanse), DEBT-128 (hook bash elimination).

---

## WS1: Materialize Pipeline Hardening

**Theme**: Address the convergent hotspot. Make the materialize pipeline reliable through atomic writes, error propagation, and timeout safety.

**Total effort**: 3-5 days
**Dependencies**: None. Can start immediately (parallel with Sprint 0).
**Rite**: `hygiene`
**Cluster**: Convergent Hotspot (17 items, 21% of portfolio)

### Sprint 1.1: Atomic Writes and Error Propagation (Critical Fixes)

**Priority**: Critical
**Size**: L (10-16h)
**Confidence**: Medium (sync path changes need careful testing; add 30% buffer)
**Sprint**: Week 1

**Source Items**:
- DEBT-175 (Composite 23, Critical): Replace `os.WriteFile` with `fileutil.AtomicWriteFile` on 4 critical state files
- DEBT-138 (Composite 25, Critical): Fix 16 checksum error-discard sites in `userscope/`

**Task Breakdown**:

*DEBT-175: Atomic writes (1 day)*
1. `internal/rite/state.go:93` -- Replace `os.WriteFile` with `fileutil.AtomicWriteFile` for rite invocation state. This is the highest-risk site (read on every `ari` invocation)
2. `internal/worktree/metadata.go:78,90,256` -- Replace all 3 `os.WriteFile` calls with `fileutil.AtomicWriteFile` for worktree metadata
3. `internal/manifest/manifest.go:191` -- Atomic write for manifest serialization
4. `internal/artifact/registry.go:173,247` -- Atomic writes for artifact registry
5. Verify `fileutil.AtomicWriteFile` exists and uses temp-file-then-rename pattern; if not, create it based on `fileutil.WriteIfChanged` patterns
6. Run full test suite; watch for failures on filesystem edge cases

*DEBT-138: Checksum error propagation (1-2 days)*
7. `internal/materialize/userscope/sync.go` (6 sites): Find all `_ := checksum.File(...)` and `_ = checksum.File(...)` patterns. Replace with error handling that returns the error up the call chain. When a checksum cannot be computed, the sync should report a warning and re-sync the file (safe default) rather than comparing against zero
8. `internal/materialize/userscope/sync_cleanup.go` (5 sites): Same pattern
9. `internal/materialize/userscope/sync_mena.go` (3 sites): Same pattern
10. `internal/materialize/userscope/sync_agents.go` (2 sites): Same pattern
11. For each site: determine if the caller should abort, warn, or skip. The safe default is warn + re-sync (conservative: treat unreadable files as changed)
12. Add test cases for at least 2 representative checksum failure scenarios (permission error, missing file)

**Risk Mitigation**:
- Atomic writes use temp-file-then-rename. On Docker volumes and network mounts, rename semantics may differ. Test on macOS local filesystem (primary target)
- Checksum error handling changes control flow in sync paths. Regression risk is real. Run `ari sync` in a test project after changes
- Back up any state files before first test run

**Acceptance Criteria**:
- [ ] `os.WriteFile` replaced with `fileutil.AtomicWriteFile` in `rite/state.go`, `worktree/metadata.go` (3 sites), `manifest/manifest.go`, `artifact/registry.go` (2 sites)
- [ ] Zero `_ := checksum.File(` or `_ = checksum.File(` patterns remain in `userscope/`
- [ ] Each former checksum-discard site either returns an error or logs a warning and forces re-sync
- [ ] At least 2 new test cases covering checksum failure scenarios
- [ ] `CGO_ENABLED=0 go test ./internal/materialize/...` passes
- [ ] `CGO_ENABLED=0 go test ./internal/rite/...` passes
- [ ] `CGO_ENABLED=0 go test ./internal/worktree/...` passes
- [ ] Manual `ari sync` produces correct output in a test project

---

### Sprint 1.2: Hook Timeout and Consistency

**Priority**: High
**Size**: M (4-8h)
**Confidence**: Medium (timeout interaction with CC requires careful handling)
**Sprint**: Week 1 (follows Sprint 1.1 or parallel)

**Source Items**:
- DEBT-149 (Composite 21, High): Add `withTimeout` to `cheapo_revert` and `worktreeremove`
- DEBT-142 (Composite 12, Medium): Fix `worktreeseed` missing embedded Agents and Mena FS sources

**Task Breakdown**:

*DEBT-149: Hook timeout safety (2-4h)*
1. `internal/cmd/hook/cheapo_revert.go`: Refactor to use `ctx.withTimeout(func() { ... })` wrapper. Currently inlines RunE body without timeout, meaning `m.Sync()` can run unbounded
2. `internal/cmd/hook/worktreeremove.go`: Refactor to use `ctx.withTimeout(func() { ... })` wrapper. Currently shells out to `git worktree remove` without timeout
3. For both hooks: convert stdin parsing from raw `io.ReadAll` + `json.Unmarshal` to `ctx.getHookEnv()` for consistency
4. Add event type guards to both hooks (currently missing)
5. Test that timeout produces a clean no-op JSON response rather than partial state

*DEBT-142: Embedded FS wiring (1-2h)*
6. `internal/cmd/hook/worktreeseed.go:159-163`: Add missing `m.WithEmbeddedAgents(embAgents)` and `m.WithEmbeddedMena(embMena)` to match the 4-source pattern used by `cheapo_revert.go:62-72` and `cmd/sync/sync.go`
7. Extract a shared `common.NewWiredMaterializer(resolver)` helper that wires all 4 embedded FS sources consistently
8. Replace the inline wiring in `cheapo_revert.go`, `worktreeseed.go`, and optionally `cmd/sync/sync.go` with the shared helper

**Risk Mitigation**:
- Adding timeout to `cheapo_revert` means `m.Sync()` can be interrupted mid-operation. Verify that partial sync does not corrupt state (the atomic write changes from Sprint 1.1 help here)
- The missing embedded FS sources in `worktreeseed` means worktrees may currently be missing mena content. Adding them could change behavior for existing worktrees. Test in a fresh worktree

**Acceptance Criteria**:
- [ ] `cheapo_revert.go` uses `ctx.withTimeout()` wrapper
- [ ] `worktreeremove.go` uses `ctx.withTimeout()` wrapper
- [ ] Both hooks use `ctx.getHookEnv()` for stdin parsing
- [ ] Both hooks have event type guards
- [ ] `worktreeseed.go` wires all 4 embedded FS sources (Rites, Templates, Agents, Mena)
- [ ] A shared `NewWiredMaterializer` helper exists and is used by at least 2 hooks
- [ ] `CGO_ENABLED=0 go test ./internal/cmd/hook/...` passes

---

**WS1 Summary**:

| Sprint | Items | Effort | Confidence | Dependencies |
|--------|-------|--------|------------|--------------|
| 1.1 | DEBT-175, DEBT-138 | 10-16h (2-3d) | Medium | None |
| 1.2 | DEBT-149, DEBT-142 | 4-8h (1d) | Medium | Benefits from 1.1 (atomic writes) |
| **Total** | **4 items** | **3-5 days** | **Medium** | **None** |

---

## WS2: CLI Convention Alignment

**Theme**: Fix 80 convention violations (stdout bypass + fmt.Errorf) and establish guards against recurrence.

**Total effort**: 5-7 days
**Dependencies**: DEBT-100 (output tests) should precede or parallel DEBT-173/174 refactoring
**Rite**: `hygiene`
**Cluster**: CLI Convention Drift (Cluster 6)

### Sprint 2.1: Output Test Harness

**Priority**: High
**Size**: M (6-8h)
**Confidence**: Medium (47 functions across 3 files; scoping JSON vs text paths)
**Sprint**: Week 3

**Source Items**:
- DEBT-100 (Composite 21, High): Write JSON output contract tests for `internal/output` package (partial -- JSON paths only)

**Task Breakdown**:
1. Inventory the 47 zero-coverage functions in `internal/output/`. Categorize by: (a) JSON output (machine-readable contract -- highest priority), (b) text formatting (human-readable -- lower priority), (c) table formatting (Rows/Headers methods -- medium priority)
2. Write test file `internal/output/output_test.go` (or extend existing) targeting JSON output paths first. For each output struct: verify `MarshalJSON()` produces valid JSON with expected fields
3. Write tests for at least the 5 highest-traffic output types: SyncOutput, RiteListOutput, SessionListOutput, AuditOutput, and ManifestOutput
4. Verify Text() methods produce non-empty, non-panicking output for representative inputs
5. Do NOT attempt full coverage of all 47 functions -- this sprint establishes the test harness and JSON contract tests. Text formatting coverage can follow in later sprints

**Acceptance Criteria**:
- [ ] JSON output tests exist for at least 5 core output struct types
- [ ] Each JSON test verifies valid JSON and expected field presence
- [ ] Text() methods have at least smoke tests (non-nil, non-panic) for 5 types
- [ ] `internal/output` coverage above 25% (up from 11.7%)
- [ ] `CGO_ENABLED=0 go test ./internal/output/...` passes

---

### Sprint 2.2: Stdout Routing Fix

**Priority**: High
**Size**: L (12-20h)
**Confidence**: Medium (41 sites; each needs structured output type routed through Printer)
**Sprint**: Week 3-4

**Source Items**:
- DEBT-173 (Composite 21, High): Refactor 41 `os.Stdout` bypass sites to use Printer

**Task Breakdown**:
1. Start with `internal/cmd/agent/validate.go` -- the most egregious case (creates Printer then ignores it). Replace all `fmt.Fprintf(os.Stdout, ...)` with `printer.Print()` calls using appropriate output types
2. `internal/cmd/session/gc.go` -- never creates Printer at all. Add Printer creation from command context and route all output through it
3. Group remaining 39 sites by package. For each `cmd/` package:
   a. Identify all `fmt.Fprintf(os.Stdout, ...)`, `fmt.Println(...)`, and `fmt.Printf(...)` patterns
   b. Determine if an appropriate output type exists in `internal/output/`; if not, create one
   c. Replace direct stdout writes with `printer.Print(outputType)` calls
4. Run the Sprint 2.1 output tests after each package to verify JSON output is not regressed
5. For sites where creating a full output type is disproportionate (e.g., simple progress messages), use `printer.VerboseLog()` or `printer.Text()` depending on message importance

**Risk Mitigation**:
- Changing output routing can break text formatting for users who rely on current stdout patterns
- Run `ari session list`, `ari rite list`, `ari agent validate` and compare output before/after
- If a site is ambiguous (intentional direct write vs convention violation), leave it and add a `// NOTE: intentional stdout write, not a Printer candidate` comment

**Acceptance Criteria**:
- [ ] `agent/validate.go` routes all output through Printer
- [ ] `session/gc.go` creates and uses Printer
- [ ] Total `fmt.Fprintf(os.Stdout` count in `internal/cmd/` reduced by at least 30 (from 41)
- [ ] `ari agent validate --output=json` produces complete, valid JSON
- [ ] `ari session list --output=json` produces complete, valid JSON
- [ ] `CGO_ENABLED=0 go test ./internal/cmd/...` passes

---

### Sprint 2.3: Structured Error Formatting

**Priority**: Medium
**Size**: M (6-12h)
**Confidence**: Medium (39 sites; mechanical replacement but exit code semantics matter)
**Sprint**: Week 4

**Source Items**:
- DEBT-174 (Composite 16, Medium): Replace 39 `fmt.Errorf` in RunE handlers with `PrintError` + proper exit codes

**Task Breakdown**:
1. Inventory all 39 `fmt.Errorf` sites in RunE handlers. The risk report identifies `cmd/org/` (8 violations) and `cmd/knows/` (5 violations) as the densest
2. For each site: replace `return fmt.Errorf(...)` with the established error pattern: `printer.PrintError(err)` + `return nil` (or the Cobra error-handling pattern used by the rest of the codebase)
3. Ensure structured output mode (`--output=json`) receives error information in the JSON response rather than as a raw error string
4. Start with `cmd/org/` (8 sites, highest density) and `cmd/knows/` (5 sites) for quick coverage
5. Verify exit codes are correct: command failures should exit non-zero

**Acceptance Criteria**:
- [ ] `cmd/org/` has zero `fmt.Errorf` returns in RunE handlers
- [ ] `cmd/knows/` has zero `fmt.Errorf` returns in RunE handlers
- [ ] Total `fmt.Errorf` in RunE handlers reduced by at least 25 (from 39)
- [ ] Error output in `--output=json` mode is structured JSON, not raw text
- [ ] `CGO_ENABLED=0 go test ./internal/cmd/...` passes

---

**WS2 Summary**:

| Sprint | Items | Effort | Confidence | Dependencies |
|--------|-------|--------|------------|--------------|
| 2.1 | DEBT-100 (partial) | 6-8h (1d) | Medium | None |
| 2.2 | DEBT-173 | 12-20h (2-3d) | Medium | Benefits from 2.1 |
| 2.3 | DEBT-174 | 6-12h (1-2d) | Medium | Independent of 2.2 |
| **Total** | **3 items** | **5-7 days** | **Medium** | **2.1 before 2.2** |

---

## WS3: SCAR Regression Safety Net

**Theme**: Build automated protection against systemic SCAR patterns and write regression tests for the highest-risk unguarded SCARs.

**Total effort**: 4-5 days
**Dependencies**: None (DEBT-179 already handled in Sprint 0)
**Rite**: `hygiene` (test writing) with `debt-triage` support (systemic analysis)
**Cluster**: Systemic SCAR Patterns (Cluster 5)

### Sprint 3.1: Systemic SCAR Fixes

**Priority**: Critical (DEBT-131) / Medium (DEBT-177, DEBT-178)
**Size**: L (10-16h)
**Confidence**: Medium (schema validation at load time is new infrastructure)
**Sprint**: Week 2-3

**Source Items**:
- DEBT-131 (Composite 22, Critical): Write regression tests for SCAR-004 (silent provenance error) and SCAR-023 (template path)
- DEBT-178 (Composite 17, Medium): Add manifest schema validation at load time
- DEBT-177 (Composite 15, Medium): Add schema registry test for session status values

**Task Breakdown**:

*DEBT-131: SCAR regression tests (1 day)*
1. Read `.know/scar-tissue.md` for SCAR-004 and SCAR-023 details
2. SCAR-004 (silent provenance error discard): Write test that simulates provenance load failure and verifies the pipeline aborts rather than silently continuing. This was structurally fixed (RISK-005 resolved) but has no regression test guarding the fix
3. SCAR-023 (template path): Write test that verifies template paths resolve correctly when KnossosHome differs from project root. The original bug was a template path that worked in dev but broke in production
4. Each test should be named `TestRegression_SCAR_NNN` for discoverability

*DEBT-178: Manifest schema validation (1-2 days)*
5. Create a `manifest.Validate()` function (or extend existing validation) that runs at manifest load time
6. Validate: required fields present, YAML structure correct, dromena/legomena entries reference files that exist on disk (at least in the loaded FS)
7. Add validation call to `manifest.Load()` so corrupt manifests are caught at parse time rather than silently producing wrong results
8. Write test cases: valid manifest, manifest with missing required field, manifest with extra unknown field, manifest referencing nonexistent mena directory

*DEBT-177: Schema registry test (1 day)*
9. Create a test that enumerates all valid session status values from `internal/session/` and verifies they are all handled by `NormalizeStatus()` alias map
10. The test should fail if a new status value is added without a corresponding alias entry
11. Add similar enumeration test for other schema-evolution-sensitive enums (rite states, event types)

**Risk Mitigation**:
- Manifest validation at load time could reject manifests that currently load successfully but have minor schema issues. Start with warnings, not hard failures
- SCAR regression tests are additive (safe). They may expose latent regressions already present -- that is the point

**Acceptance Criteria**:
- [ ] `TestRegression_SCAR_004` exists and verifies provenance error propagation
- [ ] `TestRegression_SCAR_023` exists and verifies template path resolution
- [ ] `manifest.Validate()` or equivalent runs at load time
- [ ] Corrupt manifest produces an error at parse time (not a downstream silent failure)
- [ ] Schema registry test covers all session status values in NormalizeStatus
- [ ] `CGO_ENABLED=0 go test ./...` passes

---

### Sprint 3.2: Individual SCAR Regression Tests

**Priority**: Medium
**Size**: M (6-10h)
**Confidence**: Medium (7 remaining SCARs vary in testability)
**Sprint**: Week 3 (follows Sprint 3.1)

**Source Items**:
- DEBT-131 (continued): Write regression tests for remaining 7 SCARs without automated tests

**Task Breakdown**:

The 7 remaining untested SCARs (after SCAR-004 and SCAR-023 from Sprint 3.1):
1. SCAR-002 (CC freeze -- structural fix): Write test verifying the structural fix is in place. This may be a simple assertion on code structure rather than a behavioral test
2. SCAR-008 (async hook spam): Write test that verifies hook invocation is rate-limited or deduplicated
3. SCAR-015 (stdout pollution): Write test verifying hooks produce only JSON on stdout (no log.Printf leaking to stdout)
4. SCAR-016 (bash arithmetic): Verify no bash arithmetic patterns remain in any `.sh` files. This is a static analysis test
5. SCAR-018 (context:fork): Write test verifying context fork handling
6. SCAR-020 (session ID subprocess): Write test that session ID is correctly passed through subprocess invocations
7. SCAR-027 (shared mena anti-pattern): The lint rule from Sprint 0 PKG-000d addresses this. Write an integration test that runs `ari lint` and verifies the rule catches the anti-pattern

**Risk Mitigation**:
- Some SCARs may have structural fixes that make regression impossible (SCAR-002). For these, write a code-structure assertion rather than a behavioral test
- Prioritize by risk: SCAR-015 (stdout pollution, silent failure) and SCAR-008 (async hook spam) first

**Acceptance Criteria**:
- [ ] At least 5 of 7 remaining SCARs have regression tests (some may be structural assertions)
- [ ] All regression tests follow `TestRegression_SCAR_NNN` naming convention
- [ ] `CGO_ENABLED=0 go test ./...` passes

---

**WS3 Summary**:

| Sprint | Items | Effort | Confidence | Dependencies |
|--------|-------|--------|------------|--------------|
| 3.1 | DEBT-131 (top 2), DEBT-178, DEBT-177 | 10-16h (2-3d) | Medium | None |
| 3.2 | DEBT-131 (remaining 7) | 6-10h (1-2d) | Medium | After 3.1 |
| **Total** | **3 items** | **4-5 days** | **Medium** | **3.1 before 3.2** |

Note: DEBT-176 and DEBT-179 handled in Sprint 0.

---

## WS4: Knowledge Refresh

**Theme**: Fix all stale `.know/` files in a single batch. This is entirely subsumed by Sprint 0 PKG-000a. Listed here for workstream tracking completeness.

**Total effort**: 2-3 hours (included in Sprint 0)
**Dependencies**: None
**Rite**: None required

### Sprint 4.1: Batch Knowledge Update

**Priority**: High (cluster resolution ROI)
**Size**: S (2-3h)
**Confidence**: High

This sprint is identical to Sprint 0 PKG-000a plus the DEBT-180 correction from PKG-000c. All WS4 items are handled in Sprint 0.

**Source Items** (all in Sprint 0):
- DEBT-115: test-coverage.md regeneration (PKG-000a)
- DEBT-163: architecture.md corrections (PKG-000a)
- DEBT-145: design-constraints.md RISK updates (PKG-000a)
- DEBT-137: MEMORY.md ghost skills correction (PKG-000a)
- DEBT-157: architecture.md leaf list (PKG-000a, subsumed by DEBT-163)
- DEBT-180: conventions.md testify count (PKG-000c)
- DEBT-116: Dead shell removal (PKG-000b)
- DEBT-120: Dead doc references (PKG-000b)

**No separate sprint required. WS4 is fully absorbed into Sprint 0.**

---

## WS5: Userscope Structural Remediation

**Theme**: Address the userscope cluster structural debt -- coverage, parallel path unification, and coupling reduction. This is the deep remediation after WS1 handles immediate error-propagation fixes.

**Total effort**: 5-8 days
**Dependencies**: WS1 Sprint 1.1 should complete first (error propagation fixes make test writing more effective)
**Rite**: `hygiene`
**Cluster**: Userscope Cluster (Cluster 1)

### Sprint 5.1: Userscope Test Coverage

**Priority**: Critical (DEBT-112)
**Size**: XL (20-28h)
**Confidence**: Low (large file, complex filesystem interactions; add 50% buffer)
**Sprint**: Week 4-5

**Source Items**:
- DEBT-112 (Composite 23, Critical): Write tests for userscope sync paths (sync_mena.go first)

**Task Breakdown**:
1. Read the userscope file inventory: `sync.go`, `sync_mena.go` (654 lines), `sync_cleanup.go`, `sync_agents.go`, `sync_hooks.go`, `sync_settings.go`, `sync_mcp.go` -- 7 files, 2,716 lines total
2. Create test infrastructure: `userscope/sync_test.go` with a `setupTestUserScope(t)` helper that creates a temp `~/.claude/`-like directory structure with known state
3. **Priority 1**: Test `sync_mena.go` paths (largest file, most error discards from DEBT-138). Test both filesystem and embedded FS sync paths. Cover: new mena added, mena updated, mena removed, mena with namespace collision, mena with companion files
4. **Priority 2**: Test `sync_agents.go` paths. Cover: agent added, agent updated, agent removed, agent collision detection, skills frontmatter injection
5. **Priority 3**: Test `sync_cleanup.go` paths. Cover: stale file removal, orphan directory cleanup, error handling on permission failures
6. **Priority 4**: Test `sync_hooks.go` and `sync_settings.go` (smaller files, simpler logic)
7. Target 50%+ coverage (up from 23.7%) by end of sprint. Full coverage is a follow-up effort

**Risk Mitigation**:
- Filesystem-dependent tests need temp directories and cleanup. Use `t.TempDir()` for automatic cleanup
- Tests may discover bugs in the untested 76% of code. Budget time for investigation and filing new DEBT items
- Do NOT fix bugs discovered during test writing in this sprint (unless trivial). File them and keep moving

**Acceptance Criteria**:
- [ ] `internal/materialize/userscope/` test coverage above 50% (up from 23.7%)
- [ ] `sync_mena.go` has tests for at least 5 sync scenarios (add, update, remove, collision, companion)
- [ ] `sync_agents.go` has tests for at least 3 sync scenarios (add, update, remove)
- [ ] Test infrastructure (`setupTestUserScope`) is reusable for future test additions
- [ ] `CGO_ENABLED=0 go test ./internal/materialize/userscope/...` passes

---

### Sprint 5.2: Path Unification and Decoupling

**Priority**: Medium
**Size**: L (10-16h)
**Confidence**: Medium (follows proven copyDirFS pattern from DEBT-152)
**Sprint**: Week 5-6

**Source Items**:
- DEBT-171 (Composite 18, Medium): Unify `syncUserMena` and `syncUserMenaFromEmbedded` parallel paths
- DEBT-158 (Composite 19, Medium): Extract shared mena sync interface to decouple userscope from mena
- DEBT-147 (Composite 16, Medium): Unify dual source chain construction

**Task Breakdown**:

*DEBT-171: Parallel path unification (1-2 days)*
1. Read `sync_mena.go` and identify the two parallel sync paths (filesystem vs embedded)
2. Apply the same `fs.FS` adapter pattern used by `copyDirFS` (DEBT-152, resolved) to create a unified sync path
3. The unified path should accept an `fs.FS` interface and handle both `os.DirFS` (filesystem) and `fs.Sub` (embedded) transparently
4. Remove the duplicated embedded-specific logic
5. Run Sprint 5.1 tests to verify the unification does not regress

*DEBT-158: Mena sync interface extraction (1-2 days)*
6. Identify the 18 call sites where `userscope` directly calls `mena.*` functions
7. Extract a `MenaSyncer` interface (or similar) that encapsulates the mena operations needed by userscope: Collect, Project, StripExtension, InjectCompanion, DetectType, CleanEmpty
8. Have `mena` package implement the interface
9. Change `userscope` to depend on the interface rather than the concrete package
10. This decouples userscope from mena's internal API changes

*DEBT-147: Source chain unification (2-3h)*
11. Compare the inline source chain construction in `materialize_mena.go:24-100` with `mena.BuildSourceChain()` in `source.go:46-71`
12. Unify into a single `BuildSourceChain()` that handles both embedded FS interleaving and filesystem-only cases
13. Add alignment comment or test that verifies both paths produce identical chain ordering

**Risk Mitigation**:
- Unifying parallel paths is the highest-risk change in this sprint. The DEBT-152 copyDirFS unification provides the pattern template and reduces uncertainty
- Run full `ari sync` integration test after each unification step
- The decoupling (DEBT-158) is a refactoring with no behavior change. Tests from Sprint 5.1 provide the safety net

**Acceptance Criteria**:
- [ ] `sync_mena.go` has a single unified sync path using `fs.FS` interface
- [ ] `syncUserMenaFromEmbedded` is removed or merged into the unified path
- [ ] `userscope` imports `mena` through an interface rather than direct package dependency (or call sites reduced by at least 50%)
- [ ] Source chain construction exists in one place only
- [ ] `CGO_ENABLED=0 go test ./internal/materialize/...` passes
- [ ] `ari sync` produces correct output in a test project

---

**WS5 Summary**:

| Sprint | Items | Effort | Confidence | Dependencies |
|--------|-------|--------|------------|--------------|
| 5.1 | DEBT-112 | 20-28h (3-4d) | Low | WS1 Sprint 1.1 |
| 5.2 | DEBT-171, DEBT-158, DEBT-147 | 10-16h (2-3d) | Medium | Sprint 5.1 |
| **Total** | **4 items** | **5-8 days** | **Low-Medium** | **WS1 before WS5** |

---

## Dependency Map

```
Sprint 0 (Quick Wins)
  |
  +---> no dependencies, start immediately
  |
  v
WS4 (Knowledge Refresh) -- fully absorbed into Sprint 0

WS1 Sprint 1.1 (Atomic Writes + Error Propagation)
  |
  +---> no dependencies, can parallel Sprint 0
  |
  +---> WS1 Sprint 1.2 (Hook Timeout) -- benefits from 1.1 atomic writes
  |
  +---> WS5 Sprint 5.1 (Userscope Coverage) -- depends on 1.1 error propagation
  |         |
  |         +---> WS5 Sprint 5.2 (Path Unification) -- depends on 5.1 tests
  |
  +---> WS3 Sprint 3.1 (Systemic SCAR) -- independent but benefits from 1.1
           |
           +---> WS3 Sprint 3.2 (Individual SCAR) -- depends on 3.1

WS2 Sprint 2.1 (Output Test Harness)
  |
  +---> no dependencies, can parallel any workstream
  |
  +---> WS2 Sprint 2.2 (Stdout Routing) -- depends on 2.1 tests
  |
  +---> WS2 Sprint 2.3 (Error Formatting) -- independent of 2.2
```

**Critical path**: Sprint 0 -> WS1 1.1 -> WS5 5.1 -> WS5 5.2 (longest chain: ~12-15 days)

**Parallelism opportunities**:
- Sprint 0 + WS1 1.1 can run in parallel (different files)
- WS1 1.2 + WS3 3.1 can run in parallel (different domains)
- WS2 (all sprints) is independent and can interleave with any workstream
- WS3 3.1 can start after Sprint 0 (Sprint 0 PKG-000d adds the lint rule that WS3 Sprint 3.2 tests)

---

## Sequencing Timeline

```
Week 1:   [Sprint 0 (5h)]  +  [WS1 1.1 (2-3d)]  +  [WS1 1.2 (1d)]
          ^^^^^^^^^^^^^^       ^^^^^^^^^^^^^^^^       ^^^^^^^^^^^^^
          Quick wins           Critical fixes         Hook hardening
          Day 0                Days 1-3               Day 3-4

Week 2:   [WS3 3.1 (2-3d)]  +  [WS2 2.1 (1d)]
          ^^^^^^^^^^^^^^^^      ^^^^^^^^^^^^^^
          Systemic SCARs        Output test harness
          Days 5-7              Day 8

Week 3:   [WS3 3.2 (1-2d)]  +  [WS2 2.2 (2-3d)]
          ^^^^^^^^^^^^^^^^      ^^^^^^^^^^^^^^^^
          Individual SCARs      Stdout routing fix
          Days 9-10             Days 9-12

Week 4:   [WS2 2.3 (1-2d)]  +  [WS5 5.1 starts (3-4d)]
          ^^^^^^^^^^^^^^^^      ^^^^^^^^^^^^^^^^^^^^^^^^^
          Error formatting      Userscope test coverage
          Days 13-14            Days 13-17

Week 5:   [WS5 5.1 finishes]  +  [WS5 5.2 starts (2-3d)]
                                   ^^^^^^^^^^^^^^^^^^^^^^^^
                                   Path unification
                                   Days 18-20

Week 6:   [WS5 5.2 finishes]  +  [Buffer / discovered work]
          ^^^^^^^^^^^^^^^^^^^
          Decoupling
          Days 20-22
```

**Best case**: 19 working days (minimal unknowns, no discovered bugs)
**Expected case**: 22-24 working days (medium buffer, some test-discovered issues)
**Worst case**: 28+ working days (significant unknowns in WS5, discovered bugs during test writing)

---

## Deferred Items

**Count**: 37 items not included in sprint packages (40 Low tier - 3 Resolved + aggregate items + user-input items)

### Items Requiring User Input

| ID | Description | Priority | Reason for Deferral |
|----|-------------|----------|---------------------|
| DEBT-129 | Single-binary scope unclear | Medium | Requires user to define "remaining ports" or close initiative |
| DEBT-126 | 3-version event bridge | Medium | Requires verification that all pre-ADR-0027 sessions are archived; if verified, becomes a 2-3h quick win |

### Aggregate / Non-Actionable Items

| ID | Description | Reason |
|----|-------------|--------|
| DEBT-114 | 502 functions at 0% (aggregate) | Remediated through individual package items |
| DEBT-119 | Shell 565-line footprint (aggregate) | Remediated through DEBT-116/117/118 |
| DEBT-141 | Hook handler graceful degradation | Observation, not debt. Intentional design |
| DEBT-148 | TENSION-006 shared manifest | Resolved |
| DEBT-152 | copyDirFS unification | Resolved |
| DEBT-145 | 3 RISK items documentation | Resolved (handled in Sprint 0) |
| DEBT-156 | SourceType convention stable | Intentional design, not debt |
| DEBT-161 | Output zero imports | Positive observation, not debt |

### Low-Priority Test Coverage (Deferred)

| ID | Description | Composite | Reason for Deferral |
|----|-------------|-----------|---------------------|
| DEBT-102 | org cmd 1.2% coverage | 7 | Low-frequency admin commands |
| DEBT-103 | explain cmd 42.6% coverage | 9 | Formatting-only functions |
| DEBT-104 | artifact cmd 0% CLI coverage | 10 | Library at 88% provides safety |
| DEBT-105 | common cmd 0% coverage | 8 | Thin accessors, tested indirectly |
| DEBT-108 | provenance cmd 0% CLI coverage | 11 | Library at 74.7% |
| DEBT-109 | tribute cmd 0% CLI coverage | 6 | Low-frequency |
| DEBT-110 | naxos cmd 0% CLI coverage | 6 | Library at 80.0% |
| DEBT-111 | root cmd 0% coverage | 6 | Cobra wiring only |
| DEBT-133 | clewcontract 26 funcs at 0% | 9 | Simple constructors, 81.2% overall |
| DEBT-136 | cmd/sails 39.5% coverage | 11 | Library at 79.3% |
| DEBT-132 | rite pkg 38 funcs at 0% | 14 | Higher priority items first |
| DEBT-134 | cmd/session 25 funcs at 0% | 12 | 158 test functions exist already |
| DEBT-135 | manifest pkg 24 funcs at 0% | 14 | Higher priority items first |
| DEBT-101 | lint pkg 12.9% coverage | 14 | Blocked by DEBT-164 (lint.go split) |
| DEBT-113 | config pkg 34.6% coverage | 14 | Foundational but not critical path |
| DEBT-106 | inscription cmd 0% CLI coverage | 16 | Library at 83.2%, rescored but not critical tier |

### Low-Priority Architecture/Code Items (Deferred)

| ID | Description | Composite | Reason for Deferral |
|----|-------------|-----------|---------------------|
| DEBT-117 | validation.sh shell in Go | 11 | Live but not runtime path |
| DEBT-118 | e2e-validate.sh CI-only | 10 | Outside ADR-0011 scope |
| DEBT-121 | Resume cross-rite deferred | 11 | Valid deferral, no drift |
| DEBT-122 | arch-ref skill missing | 6 | Low-frequency rite |
| DEBT-123 | 10x-dev "ghost skills" not ghosts | 6 | Corrected by Sprint 0 MEMORY.md fix |
| DEBT-124 | ADR-0028 unwritten | 12 | Awaiting empirical evidence |
| DEBT-125 | state.json last_sync dead write | 10 | Harmless dead write |
| DEBT-130 | state.json full elimination | 10 | Diminishing returns |
| DEBT-150 | Session test setup duplication | 11 | Annoying but not dangerous |
| DEBT-151 | Platform mena resolution dual paths | 9 | Working correctly |
| DEBT-153 | Dual OwnerType definitions | 13 | Cross-reference guards in place |
| DEBT-154 | McpServerConfig naming | 6 | Cosmetic |
| DEBT-155 | Deprecated Commands/Skills fields | 12 | Satellite audit incomplete |
| DEBT-159 | naxos imports Layer 2 | 13 | May be documentation fix |
| DEBT-160 | tribute imports Layer 2 | 11 | Paired with DEBT-159 |
| DEBT-162 | worktree imports materialize | 14 | Needs CLI orchestration refactoring |

### Low-Priority Monolith Items (Deferred)

| ID | Description | Composite | Reason for Deferral |
|----|-------------|-----------|---------------------|
| DEBT-164 | lint.go 784 lines, 5 domains | 12 | Optional split; unblocks DEBT-101 |
| DEBT-165 | output.go + rite.go 1,477 lines | 12 | Optional split after WS2 |
| DEBT-166 | inscription 3 files over 500 lines | 10 | Partially split already |
| DEBT-167 | worktree operations.go 707 lines | 10 | Optional extraction |
| DEBT-168 | sails generator.go 678 lines | 7 | Clean but optional split |
| DEBT-169 | clewcontract event.go 644 lines | 9 | Linear growth, optional |
| DEBT-170 | materialize.go 53% extracted | 11 | Diminishing returns |
| DEBT-172 | writeguard.go 588 lines | 8 | Not urgent |
| DEBT-139 | Zero log.Debug infrastructure | 15 | Needs design decision on logging approach |
| DEBT-144 | log.Printf warnings go to void | 16 | SyncResult.Warnings needs design |
| DEBT-107 | manifest cmd 0% CLI coverage | 12 | Library coverage provides safety |

**When to Revisit**:
- DEBT-126: After verifying pre-ADR-0027 session archival status
- DEBT-129: After user provides scope definition or closes initiative
- DEBT-139/144: When observability becomes a priority (e.g., debugging production issues)
- DEBT-164/165: When lint or output coverage sprints are planned
- Low-priority test coverage: When package-specific work creates a natural testing window
- DEBT-124: When empirical evidence from agent uplift accumulates

---

## Capacity Scenarios

### Scenario A: Full Allocation (1 engineer, 6 weeks)

**Assumptions**: 1 engineer, 6-8h/day effective, 20% buffer, sequential execution.

**Adjusted Plan**: Follow timeline as written. 22-24 working days expected.

**Impact**: All 5 workstreams completed. Portfolio risk reduction: Critical items 4 -> 0, High items 4 -> 0.

### Scenario B: Compressed (1 engineer, 3 weeks)

**Assumptions**: 1 engineer, 6-8h/day, must drop lowest-ROI sprints.

**Adjusted Packages**:
- Sprint 0: Keep (5h, highest ROI)
- WS1: Keep both sprints (3-5d, Critical items)
- WS2: Keep 2.1 only (1d, establishes test harness). Defer 2.2 and 2.3
- WS3: Keep 3.1 only (2-3d, systemic fixes). Defer 3.2
- WS5: Defer entirely

**Impact**: ~12-14 working days. Critical items resolved. High items partially addressed. Structural debt deferred.

### Scenario C: Parallel (2 engineers, 3 weeks)

**Assumptions**: 2 engineers, 6-8h/day each, can work in parallel on independent workstreams.

**Adjusted Plan**:
- Engineer 1: Sprint 0 -> WS1 -> WS5
- Engineer 2: WS2 -> WS3

**Impact**: ~15 working days calendar time (30 engineer-days). All workstreams completed 2 weeks faster.

---

## Package Reference

| ID | Title | Size | Hours | Priority | Sprint | Dependencies |
|----|-------|------|-------|----------|--------|--------------|
| PKG-000a | Documentation Accuracy Sweep | XS | 1-2 | High | S0 | None |
| PKG-000b | Dead Shell Removal | S | 2 | Medium | S0 | None |
| PKG-000c | Code Quick Fixes | XS | 1 | Medium | S0 | None |
| PKG-000d | SCAR Boundary Lint Rule | S | 2-3 | Medium | S0 | None |
| Sprint 1.1 | Atomic Writes + Error Propagation | L | 10-16 | Critical | WS1 | None |
| Sprint 1.2 | Hook Timeout + Consistency | M | 4-8 | High | WS1 | Benefits from 1.1 |
| Sprint 2.1 | Output Test Harness | M | 6-8 | High | WS2 | None |
| Sprint 2.2 | Stdout Routing Fix | L | 12-20 | High | WS2 | 2.1 |
| Sprint 2.3 | Structured Error Formatting | M | 6-12 | Medium | WS2 | None |
| Sprint 3.1 | Systemic SCAR Fixes | L | 10-16 | Critical | WS3 | None |
| Sprint 3.2 | Individual SCAR Regression Tests | M | 6-10 | Medium | WS3 | 3.1 |
| Sprint 5.1 | Userscope Test Coverage | XL | 20-28 | Critical | WS5 | WS1 1.1 |
| Sprint 5.2 | Path Unification + Decoupling | L | 10-16 | Medium | WS5 | 5.1 |

## Effort Distribution

**By Size**:
- XS: 2 packages (2-3h)
- S: 2 packages (4-5h)
- M: 4 packages (28-46h)
- L: 4 packages (36-68h)
- XL: 1 package (20-28h)

**By Priority**:
- Critical: 3 packages (40-60h)
- High: 4 packages (24-39h)
- Medium: 6 packages (25-44h)

**By Confidence**:
- High: 4 packages (6-8h)
- Medium: 7 packages (54-102h)
- Low: 1 package (20-28h)

**Total**: 90-153h (12-20 packages across 12 sprints)

---

## Session Commands

Copy-paste these commands to start each sprint. Each uses the appropriate rite, complexity level, and initiative description.

### Sprint 0: Quick Wins

```
# PKG-000a: Documentation Accuracy Sweep
/know --force test-coverage architecture design-constraints
# Then manually correct MEMORY.md and conventions.md per PKG-000a task breakdown

# PKG-000b: Dead Shell Removal
/start hygiene -c SMALL "Remove dead context-injection.sh and clean 37 doc references to defunct shell hook pipeline"

# PKG-000c: Code Quick Fixes
/start hygiene -c SMALL "Quick fixes: engine.go error collection, extractEmbeddedMena logging, KnossosHome test cleanup, conventions.md testify count"

# PKG-000d: SCAR Boundary Lint Rule
/start hygiene -c SMALL "Add ari lint rule for session artifacts in shared mena directories (SCAR-027 enforcement)"
```

### WS1: Materialize Pipeline Hardening

```
# Sprint 1.1: Atomic Writes and Error Propagation
/start hygiene -c SERVICE "Materialize pipeline hardening: replace os.WriteFile with AtomicWriteFile on 7 critical state files, fix 16 checksum error-discard sites in userscope sync"

# Sprint 1.2: Hook Timeout and Consistency
/start hygiene -c FEATURE "Hook hardening: add withTimeout to cheapo_revert and worktreeremove, fix worktreeseed missing embedded FS sources, extract shared NewWiredMaterializer helper"
```

### WS2: CLI Convention Alignment

```
# Sprint 2.1: Output Test Harness
/start hygiene -c FEATURE "Output test harness: write JSON contract tests for 5 core output types in internal/output, establish test infrastructure for convention enforcement"

# Sprint 2.2: Stdout Routing Fix
/start hygiene -c SERVICE "CLI stdout routing: refactor 41 os.Stdout bypass sites in cmd/ packages to use Printer, starting with agent/validate.go and session/gc.go"

# Sprint 2.3: Structured Error Formatting
/start hygiene -c FEATURE "CLI error formatting: replace 39 fmt.Errorf in RunE handlers with PrintError + structured JSON errors, starting with cmd/org/ (8 sites) and cmd/knows/ (5 sites)"
```

### WS3: SCAR Regression Safety Net

```
# Sprint 3.1: Systemic SCAR Fixes
/start hygiene -c SERVICE "SCAR systemic fixes: regression tests for SCAR-004 and SCAR-023, manifest schema validation at load time, session status schema registry test"

# Sprint 3.2: Individual SCAR Regression Tests
/start hygiene -c FEATURE "SCAR individual regression tests: write TestRegression_SCAR_NNN for remaining 7 untested SCARs (SCAR-002, 008, 015, 016, 018, 020, 027)"
```

### WS4: Knowledge Refresh

```
# Fully handled by Sprint 0 commands above. No separate session needed.
```

### WS5: Userscope Structural Remediation

```
# Sprint 5.1: Userscope Test Coverage
/start hygiene -c SERVICE "Userscope test coverage: write tests for sync_mena.go (5 scenarios), sync_agents.go (3 scenarios), sync_cleanup.go, target 50%+ coverage from 23.7%"

# Sprint 5.2: Path Unification and Decoupling
/start hygiene -c SERVICE "Userscope structural remediation: unify parallel sync paths via fs.FS adapter, extract MenaSyncer interface for decoupling, unify dual source chain construction"
```

---

## HANDOFF

---
source_rite: debt-triage
target_rite: hygiene
handoff_type: execution
created: 2026-03-03
initiative: Comprehensive Debt Audit Remediation
priority: critical
status: pending
blocking: false
---

### Context

Sprint planning is complete for the comprehensive debt audit (session-20260302-232344-1b73b3a8). 81 debt items scored across 6 dimensions, organized into 12 sprint-ready packages across 5 workstreams. The risk assessment identified a convergent hotspot in the materialize pipeline (21% of portfolio) and 4 Critical items that should be addressed first.

### Source Artifacts

- Risk Matrix: `docs/debt/RISK-REPORT-comprehensive-audit.md`
- Sprint Plan: `docs/debt/SPRINT-PLAN-comprehensive-audit.md` (this document)
- Debt Catalog: `docs/debt/LEDGER-comprehensive-audit.md`

### Items for Execution

**Sprint 0 (5h, start immediately)**:
- PKG-000a: Documentation Accuracy Sweep (XS, 1-2h) -- 5 stale .know/ files
- PKG-000b: Dead Shell Removal (S, 2h) -- remove context-injection.sh + 37 doc refs
- PKG-000c: Code Quick Fixes (XS, 1h) -- error collection, logging, test cleanup
- PKG-000d: SCAR Boundary Lint Rule (S, 2-3h) -- session artifact enforcement

**WS1 Sprints (3-5d, Critical priority)**:
- Sprint 1.1: Atomic Writes + Error Propagation (L, 10-16h) -- DEBT-175, DEBT-138
- Sprint 1.2: Hook Timeout + Consistency (M, 4-8h) -- DEBT-149, DEBT-142

**WS2 Sprints (5-7d, High priority)**:
- Sprint 2.1: Output Test Harness (M, 6-8h) -- DEBT-100 partial
- Sprint 2.2: Stdout Routing Fix (L, 12-20h) -- DEBT-173
- Sprint 2.3: Structured Error Formatting (M, 6-12h) -- DEBT-174

**WS3 Sprints (4-5d, Critical/Medium)**:
- Sprint 3.1: Systemic SCAR Fixes (L, 10-16h) -- DEBT-131, DEBT-178, DEBT-177
- Sprint 3.2: Individual SCAR Regression Tests (M, 6-10h) -- DEBT-131 continued

**WS5 Sprints (5-8d, Critical/Medium)**:
- Sprint 5.1: Userscope Test Coverage (XL, 20-28h) -- DEBT-112
- Sprint 5.2: Path Unification + Decoupling (L, 10-16h) -- DEBT-171, DEBT-158, DEBT-147

### Notes for Target Rite

- **Start with Sprint 0**: highest ROI, no dependencies, 5 hours total
- **WS1 is the critical path**: contains 2 of 4 Critical items and the convergent hotspot
- **WS5 depends on WS1**: error propagation fixes must land before userscope test writing
- **WS2 is independent**: can interleave with any workstream
- **Confidence levels matter**: Sprint 5.1 is Low confidence (add 50% buffer for unknowns)
- **Test-discovered bugs**: sprints may reveal latent issues. File new DEBT items rather than expanding sprint scope
- **Total estimated effort**: 19-28 days (90-153 hours) across all workstreams

### Success Criteria

- [ ] All 4 Critical items (DEBT-138, DEBT-175, DEBT-112, DEBT-131) have been addressed
- [ ] All 4 High items (DEBT-100, DEBT-149, DEBT-114, DEBT-173) have been addressed or their sprint packages completed
- [ ] Documentation Accuracy Cluster fully resolved (5 .know/ files current)
- [ ] Convergent Hotspot (materialize pipeline) has atomic writes and error propagation
- [ ] SCAR regression test count increases from 18 to at least 25 (of 27)
- [ ] Userscope test coverage above 50% (from 23.7%)
