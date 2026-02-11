# QA Report: ari Binary Full Surface Area Review

**Date**: 2026-02-11
**Scope**: All 21 top-level commands, ~75 subcommands
**Method**: Smoke test (100 commands), functional testing, adversarial testing, output format testing
**Binary**: `ari dev (none, unknown)` — go1.24.13 darwin/arm64

## Verdict: CONDITIONAL SHIP

The binary is **stable** (zero panics, zero crashes) and **functionally correct** for all core workflows. However, there are **output formatting bugs** that affect user experience and tool integration, and one **inscription template bug** that could corrupt CLAUDE.md on sync.

---

## Smoke Test Results

| Metric | Count |
|--------|-------|
| PASS | 100 |
| WARN | 0 |
| FAIL | 0 |
| PANIC | 0 |
| CRASH | 0 |

All `--help` flags work. All commands handle missing prerequisites gracefully.

---

## Findings

### CRITICAL — Must Fix Before Distribution

#### C1: `ari sync --dry-run` Outputs Raw Go Map Literals with Memory Pointers

**Observed**: `map[dry_run:true rite:map[...] user:map[resources:map[agents:0x140000ba2c0 hooks:0x140000ba420 ...]]]`

**Root Cause**: `formatSyncResult()` in `internal/cmd/sync/sync.go:193-244` returns `map[string]any` which has no `Text()` method. The printer falls through to `fmt.Fprintln()` at `internal/output/output.go:167`.

**Impact**: Primary user-facing command outputs garbage. Memory pointers leak internal Go runtime addresses. The `--budget` flag produces good output but is preceded by the raw map dump.

**Fix**: Create a `SyncResultOutput` struct with a `Text()` method that formats as a summary table.

**Files**: `internal/cmd/sync/sync.go` (lines 127, 164, 177, 193-244), `internal/output/output.go` (line 167)

---

#### C2: `ari rite pantheon` Outputs Raw Go Map Literal

**Observed**: `map[agents:[map[description:Architectural refactoring specialist... file:architect-enforcer.md model:opus name:architect-enforcer] ...] count:8 rite:hygiene]`

**Root Cause**: `runPantheon()` in `internal/cmd/rite/pantheon.go:94-100` builds a `map[string]interface{}` with no `Text()` method. Same printer fallback.

**Impact**: Pantheon command is unusable in text mode. Works fine in JSON mode.

**Fix**: Create a `PantheonOutput` struct with `Text()` that renders an agent table (name | model | role).

**Files**: `internal/cmd/rite/pantheon.go` (lines 56-101)

---

#### C3: Inscription Template Generates Bloated CLAUDE.md

**Observed**: `ari inscription diff` shows the generated `quick-start` and `agent-configurations` regions would contain **full multi-paragraph agent descriptions** (including all CC Task tool description text) instead of the current compressed one-line summaries.

**Current CLAUDE.md** (correct):
```
This project uses a 5-agent workflow (hygiene):
- `pythia.md` - Coordinates code hygiene initiative phases
```

**Generated** (wrong):
```
This project uses a 8-agent workflow (hygiene):
- `architect-enforcer.md` - Architectural refactoring specialist who evaluates smells through a boundary
  lens and produces refactoring plans with before/after contracts.
  When to use this agent: [... 20+ lines per agent ...]
```

**Two sub-issues**:
1. Agent descriptions not compressed to one-line summaries
2. Agent count is 8 (includes shared agents: consultant, context-engineer, moirai) instead of 5 (rite-native only)

**Impact**: Running `ari sync` or `ari inscription sync` would **replace compressed CLAUDE.md with massively bloated content**, destroying the carefully optimized token budget.

**Fix**: Template must (a) read only rite-native agents from workflow.yaml, not all files in .claude/agents/, and (b) use frontmatter `description` field truncated to first sentence.

**Files**: `knossos/templates/sections/quick-start.md.tpl`, `knossos/templates/sections/agent-configurations.md.tpl`, `internal/inscription/generator.go`

---

### HIGH — Should Fix Before Dogfooding

#### H1: Hook `allow` Responses Output Raw Go Struct

**Observed**: `{{PreToolUse allow  [] }}`

**Expected**: `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}`

**Root Cause**: `PreToolUseOutput` struct in `internal/hook/output.go` has no `Text()` method. Hook commands use `FormatJSON` by default but `--output=text` overrides, causing the fallback.

**Impact**: If CC ever passes `--output=text` to hooks, they'd return unparseable output. Currently hooks default to JSON so production is unaffected. The `deny` path works because CC reads the stderr message, not stdout format.

**Fix**: Add `Text()` method to `PreToolUseOutput`, or ensure hooks always force JSON regardless of `--output` flag.

**Files**: `internal/hook/output.go`, `internal/cmd/hook/writeguard.go`

---

#### H2: Multiple Commands Print Error Messages Twice

**Pattern**:
```
Error: No active session. Use 'ari session create' first.
Error: No active session. Use 'ari session create' first.
```

**Affected commands**: `ari handoff status`, `ari handoff history`, `ari tribute generate`, `ari rite info <nonexistent>`

**Root Cause**: Commands call both `printer.PrintError(err)` (writes to stderr) AND `return err` (returns to Cobra). Despite `SilenceErrors: true` on root, Cobra still prints the error for some subcommands.

**Fix**: Either remove `printer.PrintError()` calls and let Cobra handle errors, or change to `printer.PrintError(err); return nil` to suppress Cobra's error output.

**Files**: `internal/cmd/handoff/status.go:47-55`, `internal/cmd/handoff/history.go`, `internal/cmd/tribute/generate.go`, `internal/cmd/rite/pantheon.go`

---

#### H3: `ari rite info` Has Blank Fields

**Observed**: `Form: ` (blank), agent `file: ""`, `schema_version: ""` in JSON output. Meanwhile `ari rite list` correctly shows `form: practitioner`.

**Root Cause**: `rite info` reads from a different code path than `rite list`. The info command likely doesn't populate the form field from the rite manifest.

**Fix**: Ensure `rite info` populates all fields that `rite list` has.

---

#### H4: `ari rite status` Shows Wrong Agent Count

**Observed**: Shows "Agents 4" but `ari rite info` and `ari rite current` show 5 agents.

**Root Cause**: Different counting logic — `rite status` may be counting workflow phases (4: assessment, planning, execution, audit) instead of agents (5: pythia + 4 specialists).

**Fix**: Align agent counting across all rite commands.

---

### MEDIUM — Quality Improvements

#### M1: `-o xml` and Other Invalid Output Formats Silently Ignored

**Observed**: `ari version -o xml` outputs text without warning.

**Fix**: Validate output format and warn/error on unrecognized values.

---

#### M2: `-s <invalid-session-id>` Silently Ignored

**Observed**: `ari session status -s "not-a-real-session"` returns "No active session" with exit 0 instead of erroring about the nonexistent session.

**Fix**: When `-s` is explicitly provided, validate the session exists and error if not found.

---

#### M3: `ari sync --rite=nonexistent --dry-run` Returns Exit 0

**Observed**: Silently skips rite scope and only syncs user scope. Should error when the user explicitly specifies a nonexistent rite.

**Fix**: Validate rite name against available rites before syncing.

---

#### M4: Legacy Reference in `handoff status` Text Output

**Observed**: `handoff/status.go:329` outputs `Team: %s` instead of `Rite: %s`. Legacy "team" terminology.

**Fix**: Change `Team:` to `Rite:` in HandoffStatusOutput.Text().

---

### LOW — Nice to Have

#### L1: `ari version` Shows `dev (none, unknown)` for Local Builds

No build info injected. Expected for dev builds but should have proper values for distribution.

**Fix**: Add `-ldflags` to build process for version, commit, date.

---

#### L2: Lint Reports 14 HIGH Name Collisions

Legitimate content issue: shared dromena names collide with rite-specific dromena (build, hygiene, 10x, sre, debt, docs, architect). These are real naming conflicts to address.

---

#### L3: All 11 Pythia Orchestrators Have maxTurns=40

Agent validate warns all pythia agents should have maxTurns <= 5 for consultation pattern. This is a consistent deviation from the validation rule — either fix all pythias or adjust the rule.

---

## Commands That Work Well

These commands are polished, functional, and produce good output:

| Command | Notes |
|---------|-------|
| `ari session list` | Clean table with STALE hints, good JSON |
| `ari session status` | Clean text and JSON |
| `ari rite list` | Clean table with all fields |
| `ari rite current` | Good context budget info |
| `ari rite context` | Clean markdown table for Claude injection |
| `ari agent list` | Excellent cross-rite discovery table |
| `ari agent validate` | Actionable warnings with evidence |
| `ari lint` | Good severity-ranked output |
| `ari naxos scan` | Actionable stale session recommendations |
| `ari provenance show` | Clean table with status tracking |
| `ari inscription validate` | Clean pass/fail |
| `ari sync --budget` | Excellent token budget breakdown (after raw map) |
| `ari worktree list/status` | Good empty-state UX with suggestions |
| `ari hook context` | Proper session detection |
| `ari hook writeguard` (deny) | Correct blocking with clear guidance |

---

## Test Coverage Summary

| Category | Commands Tested | Pass | Issues Found |
|----------|----------------|------|--------------|
| Smoke (--help) | 75 | 75 | 0 |
| Smoke (functional) | 25 | 25 | 0 |
| Core workflow | 15 | 15 | 4 (output formatting) |
| Hook infrastructure | 6 | 6 | 1 (allow format) |
| Secondary commands | 12 | 12 | 2 (duplicate errors) |
| Output formats | 8 | 8 | 2 (format validation) |
| Adversarial | 6 | 6 | 3 (silent failures) |
| **Total** | **147** | **147** | **12** |

---

## Priority Action Items

1. **C3 first** — Inscription template fix is blocking (running sync would corrupt CLAUDE.md)
2. **C1 + C2** — Sync and pantheon Text() methods (most visible user-facing commands)
3. **H2** — Duplicate error messages (quick pattern fix across 4 files)
4. **H1** — Hook output format (production safety)
5. **H3 + H4** — Rite info/status consistency
6. **L1** — Build version injection (distribution requirement)
