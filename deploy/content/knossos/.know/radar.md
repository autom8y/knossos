---
domain: radar
generator: radar
generated_at: "2026-03-03T00:00:00Z"
expires_after: "7d"
signals_evaluated:
  - radar-confidence-gaps
  - radar-staleness
  - radar-unguarded-scars
  - radar-constraint-violations
  - radar-convention-drift
  - radar-architecture-decay
  - radar-recurring-scars
know_files_read:
  - architecture
  - conventions
  - scar-tissue
  - design-constraints
  - test-coverage
opportunity_count: 18
high_count: 7
medium_count: 8
low_count: 3
---

# Knowledge Radar — 2026-03-03

## Summary

18 opportunities across 7 signals. The strongest convergent signal is in the **materialize pipeline + untested CLI entry points** — three signals (unguarded scars, convention drift, architecture decay) independently flag this area. Two systemic scar categories (integration failure and silent failure) at 4 entries each indicate structural tendencies toward invisible failures at CC integration boundaries. Four design constraints documented as risks have been silently resolved but not marked in `.know/design-constraints.md`, creating stale knowledge.

## Opportunities

### OPP-001: os.Stdout Bypass in Agent Commands

- **Signal**: radar-convention-drift
- **Severity**: HIGH
- **Confidence**: 0.84 (0.88 × 0.95)
- **Evidence**:
  - `internal/cmd/agent/validate.go` — Printer created at line 54, then 18 `fmt.Fprintf(os.Stdout, ...)` calls bypass it entirely
  - `internal/cmd/agent/update.go` — 8 direct os.Stdout writes
  - `internal/cmd/agent/list.go` — table output via direct os.Stdout
  - `internal/cmd/session/gc.go` — 10 direct os.Stdout writes, no Printer used at all
  - 41 total bypass sites in cmd/ non-test code
- **Suggested Action**: The agent validate command is the most impactful fix — it creates a Printer but ignores it, meaning `--output=json` silently omits all validation results. Refactor `printValidationResults()` to use a structured output type routed through `printer.Print()`. Then tackle `session/gc.go` which never creates a Printer at all. This is a good hygiene session target — mechanical refactoring with high user-facing payoff.

---

### OPP-002: fmt.Errorf at CLI Boundaries (39 violations)

- **Signal**: radar-convention-drift
- **Severity**: HIGH
- **Confidence**: 0.77 (0.88 × 0.88)
- **Evidence**:
  - `internal/cmd/org/init.go` — 6 `fmt.Errorf` returns in RunE
  - `internal/cmd/org/set.go` — 3 `fmt.Errorf` at CLI boundary without PrintError
  - `internal/cmd/knows/knows.go` — 5 `fmt.Errorf` at CLI boundary
  - `internal/cmd/sync/sync.go` — 3 `fmt.Errorf` in RunE validation
  - `internal/cmd/worktree/sync.go:111` — Printer in scope, still uses fmt.Errorf
  - 39 total violations in `internal/cmd/` production code
- **Suggested Action**: These errors lose exit code control and JSON-friendly formatting. The `cmd/org/` package (8 violations) and `cmd/knows/` (5 violations) are the highest-count cleanup targets. A grep-based CI gate on `fmt.Errorf` in `internal/cmd/` would prevent further drift. Consider a focused debt-triage or hygiene session for the org and knows packages first.

---

### OPP-003: Layer 3 → Layer 2 Import Violations (tribute, naxos)

- **Signal**: radar-architecture-decay
- **Severity**: HIGH
- **Confidence**: 0.76 (0.78 × 0.98)
- **Evidence**:
  - `internal/tribute/extractor.go:11,13` — imports `internal/artifact` (L2) and `internal/session` (L2)
  - `internal/naxos/scanner.go:9,10` — imports `internal/sails` (L2) and `internal/session` (L2)
  - Both packages are classified Layer 3 (Support) but perform domain-level orchestration
- **Suggested Action**: Both packages perform domain-level work: tribute orchestrates session+artifact data for report generation, naxos implements orphan session detection requiring domain knowledge. The cleanest fix is promoting both to Layer 2 in the documented architecture — they ARE domain packages by behavior, just misclassified. Update `.know/architecture.md` layer model during the next `/know architecture` refresh. An arch review would confirm whether promotion is sufficient or if interface extraction is needed.

---

### OPP-004: Systemic "Silent Failure / Observability" Pattern (4 SCARs)

- **Signal**: radar-recurring-scars
- **Severity**: HIGH
- **Confidence**: 0.78 (0.82 × 0.95)
- **Evidence**:
  - SCAR-004: Silent error discard at provenance load
  - SCAR-006: Shared mena drop for satellite-local rites (exit code 0)
  - SCAR-007: Mixed dro/lego directories block skill resolution silently
  - SCAR-017: @skill-name anti-pattern — 195+ files silently ignored
  - SCAR-025: Deleted files in user scope sync kept manifest entries forever
  - All scars share: system completed successfully while silently discarding data
- **Suggested Action**: This is the most dangerous scar category because failures are invisible. The pattern reveals a structural tendency to not surface errors at component boundaries in the materialize pipeline. Consider instituting an "exit-0 must mean complete output" invariant: add a post-materialization verification step that counts expected vs. actually written artifacts and emits WARN when counts diverge. Review remaining blank-identifier usages in `internal/materialize/` for SCAR-004 siblings. Debt-triage is the right rite for this systemic investigation.

---

### OPP-005: Systemic "Integration Failure" Pattern (4 SCARs)

- **Signal**: radar-recurring-scars
- **Severity**: HIGH
- **Confidence**: 0.75 (0.82 × 0.92)
- **Evidence**:
  - SCAR-002: CC file watcher freeze on `.claude/` rename
  - SCAR-009: Wrong hook format (flat vs nested) — CC rejected silently
  - SCAR-018: `context: fork` blocks Task tool access — silent downgrade
  - SCAR-020: Session ID not passed to CLI subprocesses
  - 3 of 4 have no automated regression tests
- **Suggested Action**: Every CC API surface or protocol change carries high silent-failure risk. The pattern: a Knossos-side assumption about CC behavior is wrong, CC produces no diagnostic, failure is only observable behaviorally. Recommend a CC integration smoke-test harness that validates hook format roundtrip, fork-context Tool availability, and session ID propagation. This prevents future CC protocol mismatches. Debt-triage for the systemic fix, hygiene for the individual test gaps.

---

### OPP-006: Untested CLI Entry Points (inscription, manifest, artifact)

- **Signal**: radar-unguarded-scars
- **Severity**: HIGH
- **Confidence**: 0.70 (0.85 × 0.82)
- **Evidence**:
  - `internal/cmd/inscription/` — 0 test files, ~600 lines (inscription sync, rollback, validation)
  - `internal/cmd/manifest/` — 0 test files, ~500 lines (diff, merge, validate)
  - `internal/cmd/artifact/` — 0 test files, ~500 lines (register, list, rebuild)
  - These packages handle operations with idempotency and destructive-write concerns (SCAR-003, SCAR-005 patterns)
- **Suggested Action**: inscription rollback is the highest-consequence untested operation — a regression there could destroy CLAUDE.md content. Manifest validate is consumed by the materialize pipeline — a regression produces silent downstream failures matching the SCAR-004/SCAR-006 pattern. Start with smoke tests for inscription rollback and manifest validate. Hygiene session, targeted scope.

---

### OPP-007: Non-Atomic Writes on Critical State Files

- **Signal**: radar-convention-drift
- **Severity**: HIGH
- **Confidence**: 0.75 (0.88 × 0.85)
- **Evidence**:
  - `internal/rite/state.go:93` — `os.WriteFile` for invocation state (read every ari invocation)
  - `internal/worktree/metadata.go:78,90,256` — 3 `os.WriteFile` for worktree registry
  - `internal/manifest/manifest.go:191` — `os.WriteFile` for manifest save
  - `internal/artifact/registry.go:173,247` — 2 `os.WriteFile` for artifact registry
  - 59 total non-atomic writes vs 13 files using `fileutil.AtomicWriteFile` correctly
- **Suggested Action**: `rite/state.go:93` and `worktree/metadata.go` are the priority — both write YAML state files read on every ari invocation. A crash during write corrupts state with no recovery. The fix is mechanical: replace `os.WriteFile` with `fileutil.AtomicWriteFile`. Hygiene session, small scope.

---

### OPP-008: RISK-004 Active — engine.go os.Remove Without Error Propagation

- **Signal**: radar-constraint-violations
- **Severity**: MEDIUM
- **Confidence**: 0.74 (0.78 × 0.95)
- **Evidence**:
  - `internal/materialize/mena/engine.go` lines 88, 95, 263, 265, 271, 355 — six os.Remove/os.RemoveAll calls with errors discarded
  - Stale cleanup loop (lines 258-274) is highest-priority site
  - Pipeline reports "success" when file removal silently fails
- **Suggested Action**: Append removal errors to `result.Warnings` in MenaProjectionResult. The stale cleanup loop is the primary target. This aligns with the OPP-004 "exit-0 must mean complete output" theme.

---

### OPP-009: Stale Knowledge — design-constraints.md (ADDRESSED)

- **Signal**: radar-constraint-violations
- **Severity**: LOW (was MEDIUM)
- **Confidence**: 0.74 (0.78 × 0.95)
- **Status**: ADDRESSED in PKG-000a (debt remediation Wave 0). TENSION-004 line count corrected to 732, TENSION-006 marked RESOLVED, RISK-001/RISK-005 marked RESOLVED, RISK-004 marked PARTIALLY RESOLVED.
- **Remaining**: RISK-002 (namespace collision silent yield) still open.

---

### OPP-010: Low Confidence — architecture + design-constraints at 0.78

- **Signal**: radar-confidence-gaps
- **Severity**: MEDIUM
- **Confidence**: 0.62 (0.78 × 0.80)
- **Evidence**:
  - `.know/architecture.md`: confidence 0.78 — gaps in data flow documentation, layer model has misclassifications (OPP-003)
  - `.know/design-constraints.md`: confidence 0.78 — stale entries (OPP-009), tension catalog incomplete
  - Both generated same day (2026-03-01), same generator
- **Suggested Action**: Refresh both with `/know --force architecture` and `/know --force design-constraints`. Architecture should incorporate the layer corrections from OPP-003 (promote tribute and naxos). Design-constraints should mark the 4 resolved items from OPP-009. Target confidence >= 0.85 after refresh.

---

### OPP-011: RISK-003 Partial — KnossosHome Cache Poisoning in Tests

- **Signal**: radar-constraint-violations
- **Severity**: MEDIUM
- **Confidence**: 0.70 (0.78 × 0.90)
- **Evidence**:
  - `internal/materialize/unified_sync_test.go` — 8 test functions set `KNOSSOS_HOME` via `os.Setenv` + `config.ResetKnossosHome()` but defer only restores env var, not cache
  - Missing `t.Cleanup(config.ResetKnossosHome)` in all 8 functions
  - `internal/cmd/hook/context_test.go` uses the correct pattern (`t.Cleanup(config.ResetKnossosHome)`)
- **Suggested Action**: Convert the 8 manual defer blocks in `unified_sync_test.go` to use `t.Setenv` + `t.Cleanup(config.ResetKnossosHome)`. Mechanical fix, prevents cache poisoning between tests.

---

### OPP-012: Systemic "Schema Evolution" Pattern (3 SCARs)

- **Signal**: radar-recurring-scars
- **Severity**: MEDIUM
- **Confidence**: 0.72 (0.82 × 0.88)
- **Evidence**:
  - SCAR-011: Writeguard used deprecated `.current-session` after CC Session Map migration
  - SCAR-014: Phantom status values `COMPLETE`/`COMPLETED` not in FSM
  - SCAR-016: Bash arithmetic `((var++))` returns exit code 1 when zero under `set -euo pipefail`
  - Pattern: schema/contract changed, consumers not updated atomically
- **Suggested Action**: Session state management is the common thread. The `NormalizeStatus()` alias map is the right defensive pattern. Add a canonical schema registry test that enumerates all valid session status values and fails on undeclared aliases. Remaining shell scripts should be audited for `set -e` traps.

---

### OPP-013: Systemic "Data Corruption" Pattern (3 SCARs)

- **Signal**: radar-recurring-scars
- **Severity**: MEDIUM
- **Confidence**: 0.70 (0.82 × 0.85)
- **Evidence**:
  - SCAR-004: Silent error discard masked filesystem permission errors and corrupted manifests
  - SCAR-015: Shell log functions wrote to stdout, corrupting manifest JSON keys
  - SCAR-022: Abbreviated SHA256 test fixtures rejected by schema validation
  - Common root: manifest serialization/deserialization boundary failures
- **Suggested Action**: Add manifest schema validation at load time (not just test time). If shell sync scripts remain in active use, add shellcheck CI step or port to `ari`. Provenance checksum format should be validated by a constructor function.

---

### OPP-014: Systemic "Historical Boundary" Pattern (3 SCARs)

- **Signal**: radar-recurring-scars
- **Severity**: MEDIUM
- **Confidence**: 0.66 (0.82 × 0.80)
- **Evidence**:
  - SCAR-013: Ghost dirs and already-archived session wrap edge cases
  - SCAR-026: Revert — Moirai delegation coupled to writeguard output
  - SCAR-027: Revert — ephemeral skill added to permanent shared mena
  - Pattern: architectural boundary violations caught by revert, not by enforcement
- **Suggested Action**: Add an `ari lint` rule detecting session artifact filenames in `rites/shared/mena/`. Document the writeguard coupling constraint in `.know/design-constraints.md`. The 3 wrap edge case tests exist and pass — verify they run in CI.

---

### OPP-015: Peer Coupling at Domain Layer (sails→session, worktree→materialize)

- **Signal**: radar-architecture-decay
- **Severity**: MEDIUM
- **Confidence**: 0.74 (0.78 × 0.95)
- **Evidence**:
  - `internal/sails/generator.go:14` imports `internal/session` (peer L2→L2)
  - `internal/sails/contract.go:12` imports `internal/hook/clewcontract` (L2→L3 sub-package, partially intentional)
  - `internal/worktree/operations.go:16` imports `internal/materialize` (peer L2→L2, undocumented)
- **Suggested Action**: The sails→clewcontract import is intentional per the architecture doc. The sails→session and worktree→materialize imports create peer coupling that should either be documented as accepted (update architecture.md) or resolved by pushing the import up to the cmd layer. Lower priority than the L3→L2 violations in OPP-003.

---

### OPP-016: Token-Only Test Coverage (sync, rite, lint)

- **Signal**: radar-unguarded-scars
- **Severity**: MEDIUM
- **Confidence**: 0.68 (0.85 × 0.80)
- **Evidence**:
  - `internal/cmd/sync/` — sync_test.go has 35 tests but only exercises command metadata (Use, Short, NeedsProject), not the 364-line sync logic
  - `internal/cmd/rite/` — rite_test.go has 100 tests, all command-metadata pattern
  - `internal/cmd/lint/` — lint_test.go exists but SCAR-017 and SCAR-019 patterns at tail of file may not be exercised
- **Suggested Action**: `internal/cmd/sync/` is the primary entry point for the materialize pipeline with zero behavioral tests at the CLI layer. Add at least one integration test exercising sync against a fixture rite directory. The underlying materialize pipeline is well-tested — this gap is at the wiring layer only.

---

### OPP-017: Leaf Classification Mismatch (internal/registry)

- **Signal**: radar-architecture-decay
- **Severity**: LOW
- **Confidence**: 0.70 (0.78 × 0.90)
- **Evidence**:
  - `internal/registry/validate.go:13,14` imports `internal/frontmatter` and `internal/mena`
  - Registry is documented as a leaf package (no internal imports)
  - These are Layer 4 imports (not upward violations) but contradict the leaf designation
- **Suggested Action**: Remove `internal/registry` from the leaf package list in the architecture doc. The validate.go file requires frontmatter and mena parsing — this is legitimate functionality, just mislabeled. Low priority, update during next `/know architecture` refresh.

---

### OPP-018: Testify Drift Beyond Documented Scope

- **Signal**: radar-convention-drift
- **Severity**: LOW
- **Confidence**: 0.69 (0.88 × 0.78)
- **Evidence**:
  - Documented: 18 files use testify (materialize, sails)
  - Observed: 23 files (5 outside documented scope: tour, explain, tokenizer, materialize/hooks/mcp, agent/mcp_validate)
  - Convention states "Do not migrate" implying freeze
- **Suggested Action**: No corrective action on existing files. Update `conventions.md` to note the 23-file actual count as the new freeze line. New test files must use stdlib testing. Low priority — this is documentation staleness, not active risk.

## Signals with No Findings

| Signal | Result |
|---|---|
| radar-staleness | All 5 domains fresh (expire 2026-03-08, 5 days remaining) |

## Suppressed Findings

None. All findings exceeded the 0.40 confidence floor.

## Advisory

The `internal/materialize/` pipeline area is the convergent hotspot. Three independent signals (unguarded scars, convention drift, architecture decay) all point to it from different angles — untested CLI entry points feeding into it, silent failure patterns within it, and layer misclassifications around it. If you're choosing where to invest next, a combined hygiene + arch review session targeting the materialize ecosystem would address OPP-001, OPP-004, OPP-006, OPP-007, and OPP-008 in one pass.

The `.know/` files themselves need a refresh: both architecture and design-constraints sit at 0.78 confidence, and design-constraints has 4 resolved items still marked as active risks. Running `/know --force architecture design-constraints` before acting on architecture-related opportunities (OPP-003, OPP-015, OPP-017) ensures agents work from current knowledge.

## Methodology

- **Signals evaluated**: radar-confidence-gaps, radar-staleness, radar-unguarded-scars, radar-constraint-violations, radar-convention-drift, radar-architecture-decay, radar-recurring-scars
- **Source files read**: architecture, conventions, scar-tissue, design-constraints, test-coverage
- **Theoros dispatched**: 7 (Argus Pattern — parallel)
- **Deduplication**: Grouped by package/area; multi-signal entries combined with severity=max, confidence=min
- **Priority ordering**: Severity (HIGH → LOW) then confidence (descending)
- **Run date**: 2026-03-03
