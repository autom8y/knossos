# Refactoring Plan: Legacy Migration Residue + CLI Discoverability

**Author**: architect-enforcer
**Date**: 2026-02-11
**Status**: READY FOR JANITOR
**Scope**: Subtractive (ghost CLI commands, dead code) + Additive (CLI discoverability hints)

---

## Architectural Assessment

### Root Cause

The ADR-0026 Phase 4b "Pipeline Absorption" unified the CLI surface under `ari sync` but did not propagate the change to all reference sites. The old CLI topology included `ari sync materialize`, `ari sync pull/push/resolve/diff/history/reset`, `ari rite switch`, and `ari rite start`. These commands were deleted from the command tree but their string references survive in:

1. **Generated CLAUDE.md content** (read by Claude agents at runtime)
2. **Go error messages and output types** (user-facing strings)
3. **Mena files** (projected into `.claude/commands/` and `.claude/skills/`)
4. **Documentation** (docs/, README.md, rites/*/TODO.md)

### Boundary Health

- **Critical boundary violation**: Templates and generator fallbacks contain ghost CLI commands (`ari rite switch`) that are injected into CLAUDE.md on every `ari sync`. This means every project using knossos is currently receiving instructions referencing commands that do not exist.
- **Dead code boundary**: `internal/output/sync.go` contains 6 output types (SyncPullOutput, SyncPushOutput, SyncDiffOutput, SyncResolveOutput, SyncHistoryOutput, SyncResetOutput) plus helper functions that have zero callers anywhere in the codebase. These correspond to the 6 remote sync subcommands deleted in Phase 4b.
- **Dead state boundary**: `internal/sync/state.go` contains State, TrackedFile, Conflict, StateManager and methods that ARE still used by `internal/materialize/materialize.go:1437` and `internal/materialize/rite_switch_integration_test.go:85`. This file is NOT dead -- it provides state tracking for materialization. Only `ComputeFileHash` and `ComputeContentHash` are thin wrappers around `internal/checksum/` but they have callers in state_test.go. The entire file must be KEPT.

### Classification Summary

| Category | Count | Risk |
|----------|-------|------|
| Live agent behavior (ghost commands in generated content) | 8 items | CRITICAL |
| Go source dead code / stale messages | 8 items | HIGH |
| Documentation / TODO legacy | 7 categories (~200+ refs) | MEDIUM |
| CLI discoverability (additive) | 5 items | LOW |

---

## Phase A: Critical Fixes -- Live Agent Behavior

**Goal**: Eliminate ghost CLI commands from all content that Claude agents read at runtime.
**Blast radius**: Templates and generator affect every project's CLAUDE.md on next sync.
**Rollback**: Single `git revert` restores previous template/generator content.
**Verification**: `CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./...`

### RF-A01: Fix generator.go fallback (C1)

**File**: `/Users/tomtenuta/Code/knossos/internal/inscription/generator.go`

**Before State** (line 492):
```go
No active rite. Use ` + "`/go`" + ` to get started, or ` + "`ari rite switch <name>`" + ` to activate directly.`
```

**After State**:
```go
No active rite. Use ` + "`/go`" + ` to get started, or ` + "`ari sync --rite=<name>`" + ` to activate directly.`
```

**Invariants**: Function signature unchanged. Return type unchanged. Only the string constant changes.

**Verification**:
1. `CGO_ENABLED=0 go build ./cmd/ari`
2. `CGO_ENABLED=0 go test ./internal/inscription/...`

---

### RF-A02: Fix provenance.go error message (C2)

**File**: `/Users/tomtenuta/Code/knossos/internal/cmd/provenance/provenance.go`

**Before State** (line 132):
```go
printer.PrintLine("No provenance manifest found. Run 'ari sync materialize' or 'ari sync user all' first.")
```

**After State**:
```go
printer.PrintLine("No provenance manifest found. Run 'ari sync' first.")
```

**Invariants**: Function signature unchanged. Error semantics unchanged (informational hint only).

**Verification**:
1. `CGO_ENABLED=0 go build ./cmd/ari`
2. `CGO_ENABLED=0 go test ./internal/cmd/provenance/...`

---

### RF-A03: Fix pantheon.go error message (discovered during verification)

**File**: `/Users/tomtenuta/Code/knossos/internal/cmd/rite/pantheon.go`

**Before State** (line 46):
```go
return printer.PrintError(fmt.Errorf("no active rite (use 'ari sync materialize --rite <name>' to activate)"))
```

**After State**:
```go
return printer.PrintError(fmt.Errorf("no active rite (use 'ari sync --rite=<name>' to activate)"))
```

**Invariants**: Error semantics unchanged. Only hint text changes.

**Verification**:
1. `CGO_ENABLED=0 go build ./cmd/ari`
2. `CGO_ENABLED=0 go test ./internal/cmd/rite/...`

---

### RF-A04: Fix sessions.dro.md ghost command (C3)

**File**: `/Users/tomtenuta/Code/knossos/mena/navigation/sessions.dro.md`

**Before State** (lines 93-100):
```markdown
### --switch {id}

Switch this terminal to a different session:

```bash
# Session resolution is handled internally by ari
ari session switch "$SESSION_ID"
```
```

**After State**:
```markdown
### --switch {id}

Switch this terminal to a different session:

```bash
# Session resolution is handled internally by ari
ari session resume "$SESSION_ID"
```

> For full subcommand list: `ari session --help`
```

**Rationale**: `ari session switch` does not exist. The closest equivalent is `ari session resume`. The ground truth CLI tree confirms `resume` exists as a subcommand of `ari session`.

**Invariants**: This is a mena source file projected to `.claude/commands/`. No Go compilation involved.

**Verification**:
1. Confirm `ari session resume` appears in `internal/cmd/session/resume.go`
2. Content review only (no compilation)

---

### RF-A05: Fix README.md ghost commands (C4, C5)

**File**: `/Users/tomtenuta/Code/knossos/README.md`

**Before State** (lines 17-18):
```markdown
| `ari sync materialize --rite <name>` | Switch active rite (syncs to `.claude/`) |
| `ari rite switch <name>` | Alias for rite switching |
```

**After State**:
```markdown
| `ari sync --rite=<name>` | Switch active rite (syncs to `.claude/`) |
```

**Also** (line 23):
**Before**:
```markdown
Rite-level content (`rites/{rite}/`) syncs to `.claude/` (project-specific via ari sync materialize).
```
**After**:
```markdown
Rite-level content (`rites/{rite}/`) syncs to `.claude/` (project-specific via `ari sync --rite=<name>`).
```

**Also** (line 37, 40, 43 -- `ari sync user` subcommands):
These lines reference `ari sync user agents`, `ari sync user mena`, etc. These are valid aliases currently -- the README describes `ari sync user <resource>` which maps to `ari sync --scope=user --resource=<resource>`. Per the ground truth CLI tree, `ari sync` has NO subcommands; it uses flags only. However, this is a broader README rewrite beyond the scope of this plan. The Janitor should flag these for a separate pass but NOT change them in this phase.

**Invariants**: Documentation only. No compilation.

**Verification**: Content review.

---

### RF-A06: Fix quick-start.md.tpl template (H7)

**File**: `/Users/tomtenuta/Code/knossos/knossos/templates/sections/quick-start.md.tpl`

**Before State** (line 13):
```
No active rite. Use `/go` to get started, or `ari rite switch <name>` to activate directly.
```

**After State**:
```
No active rite. Use `/go` to get started, or `ari sync --rite=<name>` to activate directly.
```

**Invariants**: Template output structure unchanged. Only the command string changes.

**Verification**:
1. `CGO_ENABLED=0 go test ./internal/inscription/...` (templates are tested via inscription)
2. `CGO_ENABLED=0 go test ./knossos/...` (if template tests exist)

---

### RF-A07: Fix agent-configurations.md.tpl template (H8)

**File**: `/Users/tomtenuta/Code/knossos/knossos/templates/sections/agent-configurations.md.tpl`

**Before State** (line 13):
```
No agents installed. Run `ari rite switch <name>` to install.
```

**After State**:
```
No agents installed. Run `ari sync --rite=<name>` to install.
```

**Invariants**: Template output structure unchanged.

**Verification**: Same as RF-A06.

---

### RF-A08: Fix ecosystem-ref INDEX.lego.md ghost commands (M2)

**File**: `/Users/tomtenuta/Code/knossos/rites/ecosystem/mena/ecosystem-ref/INDEX.lego.md`

**Before State** (lines 34-35):
```
ari rite start <rite>          # Start a rite (includes materialize)
ari rite list                  # List available rites
```

**After State**:
```
ari sync --rite=<name>         # Switch/activate a rite
ari rite list                  # List available rites
```

**Rationale**: `ari rite start` does not exist in the ground truth CLI tree. `ari rite list` DOES exist and is correct. `ari sync --rite=<name>` is the replacement.

**Also** (line 53):
**Before**:
```
ari rite list                             # List available rites
```
This line is correct -- `ari rite list` exists. No change needed.

**Invariants**: This is a legomena source file. No compilation.

**Verification**: Content review. Confirm `ari rite start` is not a valid command (grep Go source for command registration).

---

### Phase A Commit Convention

**Single commit**: `fix(templates): replace ghost CLI commands in generated content and mena`

All 8 items in one commit because they share the same root cause (stale CLI references in agent-visible content) and the same rollback semantics.

---

## Phase B: Go Source Cleanup -- Dead Code Removal

**Goal**: Remove dead output types, stale error hints, and stale comments from Go source.
**Blast radius**: Go source only. Compilation verifies correctness.
**Rollback**: Single `git revert`.
**Verification**: `CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./...`

### RF-B01: Remove ErrSyncConflict resolution hint (H1)

**File**: `/Users/tomtenuta/Code/knossos/internal/errors/errors.go`

**Before State** (line 379):
```go
"resolution_hint": "Run 'ari sync resolve' to resolve conflicts",
```

**After State**:
```go
"resolution_hint": "Resolve conflicts manually or re-run sync with --overwrite-diverged",
```

**Rationale**: `ari sync resolve` does not exist. The `--overwrite-diverged` flag is the current mechanism.

**Caller analysis**: `ErrSyncConflict` is defined in errors.go but has ZERO callers outside the definition file itself (grep confirmed). However, it is a public function that could be called by future code, so the fix should update the hint rather than delete the function.

**Invariants**: Function signature unchanged. Return type unchanged. Only hint string changes.

---

### RF-B02: Remove dead output types from sync.go (H2, H3, H4)

**File**: `/Users/tomtenuta/Code/knossos/internal/output/sync.go`

**Deletion scope**: Lines 107-349 contain 6 dead types + their methods + 2 helper functions:
- `SyncPullOutput` (lines 107-151) -- includes ghost hint on line 145
- `actionSymbol` helper (lines 153-164) -- only used by SyncPullOutput and SyncPushOutput
- `SyncPushOutput` (lines 166-199)
- `SyncDiffOutput` (lines 201-228)
- `SyncResolveOutput` (lines 230-259)
- `SyncHistoryOutput` (lines 261-315)
- `SyncResetOutput` (lines 317-349) (note: `StateCleared` on line 322 has no alignment with field tag -- cosmetic, irrelevant since we are deleting)

**Also dead but less clear**:
- `SyncStatusOutput` (lines 12-86) -- ZERO external callers. However, this type represents local sync status which conceptually still applies. The Janitor should DELETE it because it has zero callers and the current `ari sync` does not produce this output type.
- `SyncTrackedPath` (lines 22-30) -- only used by SyncStatusOutput
- `SyncConflictEntry` (lines 32-39) -- only used by SyncStatusOutput and SyncPullOutput
- `SyncFileChange` (lines 119-126) -- only used by SyncPullOutput and SyncPushOutput
- `statusIndicator` helper (lines 88-105) -- only used by SyncStatusOutput

**Caller verification** (grep confirmed): NO external callers of ANY type in this file. Every `output.Sync*` reference lives within `internal/output/sync.go` itself.

**Before State**: File is 349 lines containing 7 output types, 4 helper types, 2 helper functions, and their methods.

**After State**: File is deleted entirely. The `fmt` and `strings` imports and all contents are removed.

**Alternative**: If the Janitor prefers to keep the file skeleton for future use, they may keep only the package declaration and a comment explaining the deletion. But deletion is preferred -- the types can be reconstructed from git history if needed.

**Invariants**: No callers means no behavior change. Build must pass.

**Verification**:
1. `CGO_ENABLED=0 go build ./cmd/ari` -- confirms no compilation errors
2. `CGO_ENABLED=0 go test ./internal/output/...` -- confirms no test breakage

---

### RF-B03: Update stale comment in project_mena.go (H5)

**File**: `/Users/tomtenuta/Code/knossos/internal/materialize/project_mena.go`

**Before State** (lines 64-69):
```go
	// MenaProjectionAdditive adds/updates files without removing unmanaged content.
	// Used by usersync (ari sync user mena).
	MenaProjectionAdditive MenaProjectionMode = iota

	// MenaProjectionDestructive wipes target commands/ and skills/ directories
	// before projecting. Used by materialize (ari rite start).
	MenaProjectionDestructive
```

**After State**:
```go
	// MenaProjectionAdditive adds/updates files without removing unmanaged content.
	// Used by user scope sync (ari sync --scope=user).
	MenaProjectionAdditive MenaProjectionMode = iota

	// MenaProjectionDestructive wipes target commands/ and skills/ directories
	// before projecting. Used by rite scope sync (ari sync --scope=rite).
	MenaProjectionDestructive
```

**Invariants**: Comments only. No code changes.

---

### CB1 Analysis: internal/sync/state.go

**Verdict: DO NOT DELETE. File is actively used.**

Callers found:
- `internal/materialize/materialize.go:1437` -- `sync.NewStateManager(m.resolver)` in `trackState()`
- `internal/materialize/rite_switch_integration_test.go:85` -- `sync.NewStateManager(resolver)` in test
- `internal/sync/state_test.go` -- 7 call sites testing StateManager, ComputeFileHash, ComputeContentHash

The State, TrackedFile, Conflict, StateManager types and all methods are used by the materialization pipeline to track `sync/state.json`. The `ComputeFileHash` and `ComputeContentHash` wrappers are thin delegates to `internal/checksum/` but are called by the state_test.go tests.

**Recommendation**: Leave this file entirely alone. It is live infrastructure.

---

### Phase B Commit Convention

**Single commit**: `refactor(output): delete dead remote sync output types and update stale hints`

---

## Phase C: Content Cleanup + Additive Hints

**Goal**: Update documentation references and add CLI discoverability patterns.
**Blast radius**: No Go compilation involved. Content-only changes.
**Rollback**: Single `git revert`.
**Verification**: Content review only.

### Phase C1: Documentation Bulk Updates

#### RF-C01: Fix migration-runbook-schema.md (H6)

**File**: `/Users/tomtenuta/Code/knossos/rites/ecosystem/mena/doc-ecosystem/schemas/migration-runbook-schema.md`

**Before State** (line 168):
```yaml
    verification: "Run: ls .claude/sessions/.current-session (should not exist)"
```

**After State**:
```yaml
    verification: "Run: ari session status (should show no active session)"
```

**Rationale**: `.current-session` was deprecated and removed per Pillar 2 of the CC Session Map initiative. `ari session status` is the current way to check session state.

---

#### RF-C02: Fix TODO.md source_team/target_team refs (M1)

**Files** (4 files, 12 occurrences):
- `/Users/tomtenuta/Code/knossos/rites/sre/TODO.md` (lines 74-75)
- `/Users/tomtenuta/Code/knossos/rites/security/TODO.md` (lines 103-104)
- `/Users/tomtenuta/Code/knossos/rites/rnd/TODO.md` (lines 82-83, 104-105)
- `/Users/tomtenuta/Code/knossos/rites/intelligence/TODO.md` (lines 130-131, 147-148)

**Before State** (pattern):
```yaml
  source_team: <name>
  target_team: <name>
```

**After State**:
```yaml
  source_rite: <name>
  target_rite: <name>
```

**Invariants**: TODO.md files are planning documents, not executed code. The rite names themselves stay the same.

---

#### RF-C03: Delete knossos-sync.backup (M7)

**File**: `/Users/tomtenuta/Code/knossos/knossos-sync.backup`

**Action**: Delete the file.

**Rationale**: Stale backup from the shell script era. The `knossos-sync` shell script was replaced by `ari sync`. This backup has no callers and no references.

---

#### RF-C04: Bulk update docs/ references (M3, M4, M5 -- DEFERRED)

**Scope**: 185+ references to `ari sync materialize` and ~15 references to `ari rite switch` across `docs/` directory (50+ files).

**Decision: DEFER to a separate initiative.**

**Rationale**:
1. `docs/` files are design documents, ADRs, TDDs, PRDs, guides, and spikes. They describe the system AT THE TIME THEY WERE WRITTEN. Bulk-updating historical documents changes their meaning.
2. The blast radius is 50+ files with 200+ changes. This is a separate initiative, not a cleanup task.
3. The files in `docs/` are NOT read by Claude agents at runtime (they are not in mena/ or templates/).
4. The risk of introducing errors in bulk changes across 50+ files outweighs the benefit.

**Exception**: `docs/briefs/BRIEF-doc-alignment-cleanup.md` contains `ari rite switch` references but this is itself a cleanup brief describing the migration. Changing it would be revisionist.

**Janitor note**: Record this deferral. If the user later requests a docs/ sweep, produce a separate plan.

---

#### RF-C05: Update INTERVIEW_SYNTHESIS.md (M6 -- DEFERRED)

**Decision: DEFER.** Same rationale as RF-C04. This is a historical synthesis document.

---

### Phase C2: CLI Discoverability Hints (Additive)

#### RF-C06: Add session --help hint to sessions.dro.md

**File**: `/Users/tomtenuta/Code/knossos/mena/navigation/sessions.dro.md`

This is already addressed in RF-A04 where we add the `--help` hint alongside the command fix. No separate task needed.

---

#### RF-C07: Add --help hint to ecosystem-ref INDEX.lego.md

**File**: `/Users/tomtenuta/Code/knossos/rites/ecosystem/mena/ecosystem-ref/INDEX.lego.md`

**Before State** (end of "Debugging" section, line 93-94):
```markdown
## Debugging

```bash
ari sync --dry-run             # Preview sync changes
```
```

**After State**:
```markdown
## Debugging

```bash
ari sync --dry-run             # Preview sync changes
ari sync --help                # Full flag reference
ari rite --help                # Rite management subcommands
ari session --help             # Session management subcommands
```
```

**Invariants**: Additive only. No existing content modified.

---

#### RF-C08: Add error recovery hints to sync.dro.md

**File**: `/Users/tomtenuta/Code/knossos/mena/cem/sync.dro.md`

The sync dromena already has `For all flags: ari sync --help` on line 44 and a Legacy Compatibility section. It also has good error interpretation on lines 29-31. The discoverability is already adequate here.

**Decision: NO CHANGE NEEDED.** The sync dromena is well-structured. The `--help` hint exists. Recovery flags (`--recover`, `--overwrite-diverged`) are listed in the Command Flags table.

---

### Phase C Commit Convention

**Single commit**: `chore(content): fix legacy refs in rite TODOs, schema, and add CLI hints`

---

## Risk Matrix

| Phase | Blast Radius | Failure Detection | Rollback Cost | Risk Level |
|-------|-------------|-------------------|---------------|------------|
| A: Critical Fixes | Every project CLAUDE.md on next sync | Build + test | `git revert` (1 commit) | MEDIUM (high impact, low complexity) |
| B: Go Source Cleanup | Compilation only | Build + test | `git revert` (1 commit) | LOW (zero callers confirmed) |
| C: Content Cleanup | Documentation only | Content review | `git revert` (1 commit) | MINIMAL |

### Phase Ordering Rationale

Phase A must go first because it fixes live agent behavior. A Claude agent reading `ari rite switch <name>` will attempt to run a nonexistent command and fail. This is the highest-value fix.

Phase B follows because Go source changes require compilation verification and should not be mixed with content changes in the same commit.

Phase C is lowest risk and can be done independently.

---

## Janitor DO NOT List

- DO NOT modify any file in `.claude/` directly (run `ari sync` to project changes)
- DO NOT change any CLI command implementations in `internal/cmd/`
- DO NOT add new commands or flags
- DO NOT modify test fixtures unless tests reference ghost commands
- DO NOT expand scope beyond listed findings
- DO NOT touch `context_switch` event (documented deferral)
- DO NOT delete `internal/sync/state.go` (CB1 -- actively used)
- DO NOT bulk-update `docs/` directory (deferred -- RF-C04)
- DO NOT modify `INTERVIEW_SYNTHESIS.md` (deferred -- RF-C05)
- DO NOT change `README.md` lines 37-43 (`ari sync user` subcommand syntax) -- separate scope
- DO NOT modify `.wip/` files

## Janitor Execution Notes

1. **Commit granularity**: 3 commits total, one per phase (A, B, C). Each is independently revertible.

2. **Build verification**: After Phase A and Phase B, run:
   ```bash
   CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./...
   ```

3. **Template regeneration**: After Phase A commit, the changes to templates and generator will take effect on the next `ari sync` run in any project. The Janitor does NOT need to run `ari sync` -- that is the user's responsibility.

4. **File deletion in Phase B**: When deleting `internal/output/sync.go`, verify that the `output` package has other files (it should -- this is one of several files in the package). If sync.go is the only file, keep the package declaration.

5. **Replace-all patterns**: For RF-C02, use `replace_all` with `source_team` -> `source_rite` and `target_team` -> `target_rite` within each TODO.md file.

6. **Post-execution**: After all 3 commits, report the deferred items (RF-C04, RF-C05) for the audit-lead to record.

---

## Handoff Checklist

- [x] Every smell classified (addressed, deferred with reason, or dismissed)
- [x] Each refactoring has before/after contract documented
- [x] Invariants and verification criteria specified
- [x] Refactorings sequenced with explicit dependencies (A -> B -> C)
- [x] Rollback points identified between phases (each phase = 1 commit)
- [x] Risk assessment complete for each phase
- [x] CB1 caller analysis complete (KEEP, not delete)
- [x] DO NOT list specified
- [x] Deferred items documented with rationale (M3-M5, M6)

---

## Attestation

| File | Read | Verified |
|------|------|----------|
| `internal/inscription/generator.go:492` | Yes | Ghost command confirmed |
| `internal/cmd/provenance/provenance.go:132` | Yes | Stale hint confirmed |
| `internal/cmd/rite/pantheon.go:46` | Yes | Stale hint confirmed (discovered during verification) |
| `mena/navigation/sessions.dro.md:99` | Yes | Ghost command confirmed |
| `README.md:17-23` | Yes | Ghost commands confirmed |
| `internal/errors/errors.go:379` | Yes | Stale hint confirmed, zero callers |
| `internal/output/sync.go:107-349` | Yes | 6 dead types confirmed, zero external callers |
| `internal/output/sync.go:12-105` | Yes | Additional dead types confirmed, zero external callers |
| `internal/materialize/project_mena.go:65-69` | Yes | Stale comments confirmed |
| `rites/ecosystem/mena/doc-ecosystem/schemas/migration-runbook-schema.md:168` | Yes | `.current-session` reference confirmed |
| `internal/sync/state.go` | Yes | ALIVE -- 2 production callers + tests. DO NOT DELETE. |
| `knossos/templates/sections/quick-start.md.tpl:13` | Yes | Ghost command confirmed |
| `knossos/templates/sections/agent-configurations.md.tpl:13` | Yes | Ghost command confirmed |
| `rites/ecosystem/mena/ecosystem-ref/INDEX.lego.md:34` | Yes | Ghost command confirmed |
| `rites/*/TODO.md` (4 files) | Yes | 12 source_team/target_team refs confirmed |
| `knossos-sync.backup` | Yes | File exists at repo root |
| `mena/cem/sync.dro.md` | Yes | Already has --help hint, no change needed |
