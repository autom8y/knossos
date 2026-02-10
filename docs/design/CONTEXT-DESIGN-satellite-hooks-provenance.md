# Context Design: Satellite Distribution Blockers

**Author**: context-architect
**Date**: 2026-02-09
**Status**: Ready for Implementation
**Gap Analysis**: `docs/ecosystem/GAP-satellite-distribution-blockers.md`
**Backward Compatibility**: COMPATIBLE (all 4 areas are additive or corrective)

---

## Table of Contents

1. [Area 1: Provenance-Aware Hook Merge (B1)](#area-1-provenance-aware-hook-merge-b1)
2. [Area 2: Stale Reference Cleanup (B2)](#area-2-stale-reference-cleanup-b2)
3. [Area 3: Soft Rite Switch (B3)](#area-3-soft-rite-switch-b3)
4. [Area 4: Deployment Sequence](#area-4-deployment-sequence)
5. [Integration Test Matrix](#integration-test-matrix)
6. [Implementation Order](#implementation-order)

---

## Area 1: Provenance-Aware Hook Merge (B1)

### Problem

`isAriManagedGroup()` in `internal/materialize/hooks.go:210-247` only recognizes hooks whose command starts with `"ari hook"`. Legacy bash hooks (commands like `$CLAUDE_PROJECT_DIR/.claude/hooks/context-injection/session-context.sh`) are classified as "user" hooks and preserved indefinitely by `mergeHooksSettings()`. These legacy hooks reference `.sh` files that no longer exist, causing SessionStart errors in CC.

### Solution: Invert the Merge Model

**Decision**: Replace the current "preserve everything except ari hooks" model with "preserve nothing except genuine user hooks." The merge function will recognize a new category -- **legacy platform hooks** -- and strip them during merge, leaving only ari-managed hooks from `hooks.yaml` plus any genuinely user-authored hooks.

**Rationale**: The stakeholder preference is "nuke all non-ari hooks, with responsible validation." A pure nuke (remove all non-`ari hook` entries) risks destroying genuine user hooks that a satellite developer added manually. The responsible approach is to identify known legacy patterns and strip those specifically, while still preserving truly unknown user entries. In practice, across all 4 satellites, every non-ari hook is a legacy bash hook -- but the design must be correct for future cases where a user adds a custom hook.

**Rejected alternative 1 -- Pure nuke (remove all non-ari)**: Simpler, but violates the materialization invariant "user content is NEVER destroyed." If a satellite developer manually added a custom hook (not ari, not legacy bash), it would be silently removed. The cost of pattern-matching is low; the cost of data loss is high.

**Rejected alternative 2 -- Per-entry provenance tracking in manifest**: Track each hook entry (not just the settings.local.json file) in PROVENANCE_MANIFEST.yaml. This is architecturally clean but requires schema changes to the provenance manifest (entries currently map to files/directories, not JSON sub-entries). The complexity is disproportionate to the problem: we need to identify ~10-12 known-stale patterns, not build a general-purpose sub-file provenance system. Provenance at the file level (settings.local.json) remains sufficient; the hook merge logic itself handles entry-level ownership.

### Design

#### 1.1 New Function: `isLegacyPlatformHook()`

Add a new function in `internal/materialize/hooks.go` after `isAriManagedGroup()`:

```
isLegacyPlatformHook(group map[string]any) bool
```

This function returns true if a matcher group contains hooks that match known legacy platform patterns. It examines the `command` field of each hook in the group using these detection rules:

| Pattern | Example | Rationale |
|---------|---------|-----------|
| Contains `$CLAUDE_PROJECT_DIR` | `$CLAUDE_PROJECT_DIR/.claude/hooks/...` | All legacy bash hooks use this env var expansion |
| Contains `.claude/hooks/` as a path segment | `/path/to/.claude/hooks/session-context.sh` | Legacy hooks lived in `.claude/hooks/` directory |
| Ends with `.sh` AND does not start with `ari` | `./hooks/delegation-check.sh` | Any shell script not from ari is legacy |

The function checks both nested format (`group["hooks"][]`) and old flat format (`group["command"]`), mirroring the structure of `isAriManagedGroup()`.

**Detection logic**: A group is legacy if ANY hook in the group matches ANY legacy pattern. This is intentionally aggressive -- a matcher group mixing legacy and non-legacy hooks should not exist in practice, and if it does, the legacy hook within it is the problem.

#### 1.2 Modified Function: `mergeHooksSettings()`

Modify the existing merge loop in `internal/materialize/hooks.go:169-183`. Currently:

```go
if !isAriManagedGroup(group) {
    userEntries = append(userEntries, group)
}
```

Change to three-way classification:

```
if isAriManagedGroup(group):
    // Skip -- will be replaced by new ari hooks from hooks.yaml
elif isLegacyPlatformHook(group):
    // Skip -- strip legacy bash hooks (the core fix)
else:
    // Genuine user hook -- preserve
    userEntries = append(userEntries, group)
```

The new merge order for each event is: ari hooks (from hooks.yaml, sorted by priority) followed by genuine user hooks. Legacy hooks are dropped entirely.

#### 1.3 Responsible Validation: Diff Output

Add a return value to `mergeHooksSettings()` that reports what was stripped. Change the signature:

```
Current:  mergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig) map[string]any
New:      mergeHooksSettings(existingSettings map[string]any, hooksConfig *HooksConfig) (map[string]any, []string)
```

The second return value is a slice of human-readable strings describing each stripped legacy hook, for example: `"SessionStart: stripped legacy hook: $CLAUDE_PROJECT_DIR/.claude/hooks/context-injection/session-context.sh"`.

The caller (`materializeSettingsWithManifest()` in `materialize.go:1386-1431`) can log these strings to stdout in non-JSON mode, giving the user visibility into what was removed. This satisfies "responsible validation" -- the user sees exactly what was stripped, without requiring a separate dry-run step.

#### 1.4 File-Level Changes

| File | Function | Change |
|------|----------|--------|
| `internal/materialize/hooks.go` | `isLegacyPlatformHook()` | NEW -- legacy pattern detection (after line 247) |
| `internal/materialize/hooks.go` | `mergeHooksSettings()` | MODIFY lines 169-183 -- add three-way classification |
| `internal/materialize/hooks.go` | `mergeHooksSettings()` | MODIFY signature -- add `[]string` return for stripped-hooks report |
| `internal/materialize/materialize.go` | `materializeSettingsWithManifest()` | MODIFY line 1397 -- handle new return value, log stripped hooks |
| `internal/materialize/hooks_test.go` | (new test cases) | ADD -- test legacy detection, merge with mixed hooks, merge with pure legacy |

#### 1.5 Backward Compatibility: COMPATIBLE

- Satellites with zero legacy hooks: no behavior change (three-way classification yields same result as two-way when no legacy hooks exist).
- Satellites with legacy hooks: legacy hooks are stripped on next `ari sync`. This is a corrective behavior change, not a breaking change -- the legacy hooks were already broken (referencing missing `.sh` files).
- Genuine user hooks (if any exist): preserved by the else branch. No user content destroyed.

---

## Area 2: Stale Reference Cleanup (B2)

### Problem

177 references to `ari sync` remain post-ADR-0026 Phase 4b. The `materialize` subcommand no longer exists; the correct command is `ari sync [--rite=NAME]`. The Go source references are actively executed and cause haiku-class models to hallucinate stale commands.

### Solution: Tiered Cleanup

**Decision**: Three-tier approach -- fix Go source (critical), fix user-facing docs (high), bulk-replace internal docs (medium).

**Rationale**: Go source references appear in CLI error output that models read and follow. User-facing docs (README.md, INTERVIEW_SYNTHESIS.md) are loaded into CC project context. Internal docs (docs/) are rarely loaded into model context but represent search noise and confusion vectors.

**Rejected alternative -- Annotate as historical**: Adding "(deprecated, use `ari sync`)" to 170+ references is more work than replacing them, and leaves the stale string in the search surface. The old command genuinely does not exist; there is no historical value in preserving the exact string.

### Design

#### 2.1 Critical: Go Source (2 files)

| File | Line | Current String | Replacement String |
|------|------|----------------|-------------------|
| `internal/cmd/rite/pantheon.go` | 46 | `"no active rite (use 'ari sync --rite <name>' to activate)"` | `"no active rite (use 'ari sync --rite <name>' to activate)"` |
| `internal/cmd/provenance/provenance.go` | 132 | `"No provenance manifest found. Run 'ari sync' or 'ari sync user all' first."` | `"No provenance manifest found. Run 'ari sync' first."` |

Note for provenance.go: The `ari sync user all` reference is also stale -- `ari sync` with default `--scope=all` covers both rite and user scopes. The replacement simplifies to just `ari sync`.

#### 2.2 High: User-Facing Docs (2 files)

**README.md**:

| Line | Current | Replacement |
|------|---------|-------------|
| 17 | `ari sync --rite <name>` | `ari sync --rite <name>` |
| 23 | `ari sync` | `ari sync` |
| 93 | `ari sync` | `ari sync` |

**INTERVIEW_SYNTHESIS.md**:

| Line | Current | Replacement |
|------|---------|-------------|
| 173 | `ari sync --rite <name>` | `ari sync --rite <name>` |

#### 2.3 Medium: Internal Docs (170+ references across 20+ files in docs/)

**Policy**: Bulk find-and-replace across all files in `docs/`. Two replacement patterns:

| Pattern | Replacement |
|---------|-------------|
| `ari sync --rite` | `ari sync --rite` |
| `ari sync` (standalone, not followed by ` --rite`) | `ari sync` |

Order matters: replace the longer pattern first to avoid double-replacement.

**Exception**: `mena/cem/sync.dro.md:48` already correctly documents `ari sync` as legacy with "Use `ari sync` instead." This file should NOT be modified -- the legacy note is intentional documentation.

#### 2.4 File-Level Changes

| File | Change |
|------|--------|
| `internal/cmd/rite/pantheon.go:46` | MODIFY -- replace error string |
| `internal/cmd/provenance/provenance.go:132` | MODIFY -- replace help string |
| `README.md:17,23,93` | MODIFY -- 3 string replacements |
| `INTERVIEW_SYNTHESIS.md:173` | MODIFY -- 1 string replacement |
| `docs/**/*.md` (20+ files) | MODIFY -- bulk replacement of 170+ references |

#### 2.5 Backward Compatibility: COMPATIBLE

These are string-only changes in error messages, help text, and documentation. No API, schema, or behavioral changes. The old command does not exist, so no backward compatibility concern.

---

## Area 3: Soft Rite Switch (B3)

### Problem

`MaterializeWithOptions()` in `internal/materialize/materialize.go:264-397` writes to agents/, commands/, skills/, CLAUDE.md, settings.local.json, and several other files in a single pass. When called from within a CC session (via Bash tool), the rapid multi-file changes trigger CC's file watcher, causing a hang or deadlock between the Bash tool and the watcher.

CC constraints:
- **Agents**: Read on-demand -- safe for mid-session modification.
- **CLAUDE.md**: Re-read mid-session -- safe for mid-session modification.
- **Commands/Skills**: Cached at startup -- changes require CC restart.
- **Settings (hooks)**: Snapshotted at startup -- changes are no-ops until restart.
- **Rapid multi-file writes**: Trigger watcher congestion leading to hang.

### Solution: `--soft` Flag with CC-Safe Pipeline Subset

**Decision**: Add a `--soft` flag to `ari sync` that limits the materialization pipeline to only CC-safe stages: `materializeAgents()` and `materializeCLAUDEmd()`. All other stages (mena, rules, settings, workflow) are skipped in soft mode. The flag is explicit and user-controlled; there is no automatic CC detection.

**Rationale**: The soft switch must update only what CC can consume mid-session (agents and CLAUDE.md). Writing agents alone does not trigger the watcher hang because agent files are read on-demand and the watcher handles them gracefully. Writing CLAUDE.md alone is also safe -- the documented hang occurs when many files change in rapid succession. Two sequential writes (agents dir + one CLAUDE.md file) are within CC's tolerance. Skipping mena, settings, and rules avoids writing to the cached/snapshotted files entirely, eliminating both the hang and the "silent no-op" confusion where changes appear to succeed but have no effect until restart.

**Rejected alternative 1 -- Auto-detect CC via `CLAUDE_PROJECT_DIR` env var**: CC sets this env var when running hooks, but `ari sync` is called via Bash tool, not as a hook. The env var may or may not be present in the Bash tool environment depending on CC's shell inheritance. Relying on it would create a fragile heuristic that breaks silently when CC changes its environment propagation. An explicit flag is deterministic.

**Rejected alternative 2 -- Separate `SoftSwitch()` method**: Adding a parallel method duplicates the rite resolution, provenance, and orchestration logic. The cost of maintaining two methods that must stay in sync exceeds the cost of adding a conditional within the existing method. A single method with a mode flag is simpler and ensures all rite resolution logic remains in one place.

**Rejected alternative 3 -- Document-only approach ("use external terminal")**: This is the current workaround and it works, but it breaks the developer flow. The user must context-switch from CC to an external terminal, run the command, then return to CC. The soft switch eliminates this friction for the 90% case (agents + CLAUDE.md updates) while preserving external terminal for the 10% case (full sync with hooks/mena changes).

### Design

#### 3.1 New Field in SyncOptions

Add to `internal/materialize/sync_types.go`:

```
SyncOptions struct {
    ...existing fields...
    Soft  bool  // CC-safe mode: only update agents + CLAUDE.md
}
```

#### 3.2 New Field in Options (Legacy)

Add to `internal/materialize/materialize.go` `Options` struct:

```
Options struct {
    ...existing fields...
    Soft  bool  // CC-safe mode: only update agents + CLAUDE.md
}
```

#### 3.3 Modified Function: `MaterializeWithOptions()`

In `internal/materialize/materialize.go:264-397`, wrap stages 5-9.5 in a soft-mode conditional. The pipeline becomes:

| Step | Stage | Soft Mode | Full Mode |
|------|-------|-----------|-----------|
| 1 | Resolve rite source | RUN | RUN |
| 2 | Ensure .claude/ directory | RUN | RUN |
| 2.5 | Clear invocation state | RUN | RUN |
| 3 | Handle orphans | RUN | RUN |
| 4 | `materializeAgents()` (line 350) | RUN | RUN |
| 5 | `materializeMena()` (line 355) | SKIP | RUN |
| 6 | `materializeRules()` (line 360) | SKIP | RUN |
| 7 | `materializeCLAUDEmd()` (line 365) | RUN | RUN |
| 8 | `materializeSettingsWithManifest()` (line 372) | SKIP | RUN |
| 9 | `trackState()` (line 377) | RUN | RUN |
| 9.5 | `materializeWorkflow()` (line 382) | SKIP | RUN |
| 10 | `writeActiveRite()` (line 387) | RUN | RUN |
| 11 | Save provenance manifest | RUN | RUN |

The conditional is a simple `if !opts.Soft { ... }` wrapping each skipped stage. Steps 1-4, 7, 9, 10, 11 always run regardless of mode.

**Provenance in soft mode**: The provenance collector still runs for agents and CLAUDE.md. Skipped stages simply do not record entries. The manifest merge (step 11) carries forward previous entries for skipped stages via the `prevManifest` carry-forward logic in `saveProvenanceManifest()` (lines 1570-1586). This is correct -- the provenance manifest reflects what is actually on disk, and skipped stages leave their files unchanged.

#### 3.4 Result Reporting for Soft Mode

Modify `Result` struct to indicate soft mode:

```
Result struct {
    ...existing fields...
    SoftMode        bool     // true if soft mode was used
    DeferredStages  []string // stages skipped in soft mode
}
```

When `opts.Soft` is true, set `result.SoftMode = true` and `result.DeferredStages = []string{"mena", "rules", "settings", "workflow"}`.

The sync command output handler (`formatSyncResult()` in `internal/cmd/sync/sync.go`) should display a notice when soft mode was used:

```
Soft sync complete (CLAUDE.md + agents updated).
Deferred: commands, skills, hooks, rules (restart CC for full sync, or run 'ari sync' from external terminal).
```

#### 3.5 CLI Wiring

Add flag in `internal/cmd/sync/sync.go`:

```go
var soft bool
cmd.Flags().BoolVar(&soft, "soft", false, "CC-safe mode: update only agents and CLAUDE.md (skip hooks/mena/rules)")
```

Wire to `SyncOptions.Soft` in the `RunE` handler. Also wire through `syncRiteScope()` to the `legacyOpts` construction:

```go
legacyOpts := Options{
    ...existing fields...
    Soft: opts.Soft,
}
```

#### 3.6 File-Level Changes

| File | Function/Struct | Change |
|------|-----------------|--------|
| `internal/materialize/sync_types.go` | `SyncOptions` struct | ADD `Soft bool` field |
| `internal/materialize/materialize.go` | `Options` struct | ADD `Soft bool` field |
| `internal/materialize/materialize.go` | `Result` struct | ADD `SoftMode bool`, `DeferredStages []string` fields |
| `internal/materialize/materialize.go` | `MaterializeWithOptions()` | MODIFY -- wrap steps 5, 6, 8, 9.5 in `if !opts.Soft` |
| `internal/materialize/materialize.go` | `syncRiteScope()` | MODIFY -- pass `opts.Soft` to `legacyOpts` |
| `internal/cmd/sync/sync.go` | `NewSyncCmd()` | ADD `--soft` flag |
| `internal/cmd/sync/sync.go` | `runSync()` | MODIFY -- wire `soft` to `SyncOptions` |
| `internal/cmd/sync/sync.go` | `formatSyncResult()` | MODIFY -- display soft mode notice and deferred stages |

#### 3.7 Backward Compatibility: COMPATIBLE

- Default behavior unchanged: `ari sync` without `--soft` runs the full pipeline exactly as before.
- New `--soft` flag is opt-in. Existing scripts, dromena, and automation are unaffected.
- The flag is a strict subset of the full pipeline -- it can never produce a state that the full pipeline cannot.
- No schema changes, no new file formats, no API changes.

#### 3.8 Interaction with B1 (Legacy Hook Cleanup)

When `--soft` is used, `materializeSettingsWithManifest()` is skipped, which means legacy hooks are NOT cleaned up in soft mode. This is correct -- hooks are snapshotted at CC startup, so cleaning them mid-session has no effect anyway. The user must either:

1. Run `ari sync` from an external terminal (full sync, cleans hooks, requires CC restart), OR
2. Run `ari sync --soft` mid-session (updates agents + CLAUDE.md), then restart CC (which triggers a full hook reload from the already-cleaned settings.local.json from a previous full sync).

The expected workflow for initial satellite cleanup is always a full sync from an external terminal, not a soft switch. Soft switch is for subsequent rite changes during an active CC session.

---

## Area 4: Deployment Sequence

### Overview

All fixes ship as a single `ari` binary rebuild. Deployment to satellites follows a fixed sequence: build, then sync each satellite, then verify.

### Step 1: Build

1. Apply all code changes (B1, B2, B3) to knossos
2. Run tests: `CGO_ENABLED=0 go test ./...`
3. Build: `CGO_ENABLED=0 go build ./cmd/ari`
4. Install: `cp ./ari $(which ari)`

### Step 2: Sync Satellites

**Order**: autom8y_platform first (most complex: mixed hooks + active rite), then autom8_asana (confirms consistency), then autom8_data and autom8 (first-time ari hook sync).

For each satellite:

```bash
cd /path/to/satellite
ari sync  # Full sync from external terminal (NOT from CC session)
```

### Step 3: Per-Satellite Verification

#### autom8y_platform (12 legacy + 10 ari + 1 missing .sh)

| Check | Command/Method | Expected |
|-------|---------------|----------|
| Exit code | `ari sync` exit code | 0 |
| Legacy hooks gone | `grep -c "CLAUDE_PROJECT_DIR" .claude/settings.local.json` | 0 matches |
| Legacy hooks gone | `grep -c "\.sh" .claude/settings.local.json` | 0 matches |
| Ari hooks present | `grep -c "ari hook" .claude/settings.local.json` | 10 (one per hooks.yaml entry) |
| Stripped report | `ari sync` stdout during sync | Lists 12 stripped legacy entries |
| CC session start | Launch `claude`, observe no hook errors | Clean startup |
| Stale refs | `grep -r "ari sync" .claude/` | 0 matches |
| Soft switch | From CC: `ari sync --rite ecosystem --soft` | Exits 0, prints soft mode notice |

#### autom8_asana (12 legacy + 10 ari + 1 missing .sh)

Same checks as autom8y_platform. Confirms that the fix is consistent across satellites with identical hook states.

#### autom8_data (10 legacy + 0 ari + 1 missing .sh)

| Check | Command/Method | Expected |
|-------|---------------|----------|
| Exit code | `ari sync` exit code | 0 |
| Legacy hooks gone | `grep -c "CLAUDE_PROJECT_DIR" .claude/settings.local.json` | 0 matches |
| Ari hooks present | `grep -c "ari hook" .claude/settings.local.json` | 10 (first-time ari hook install) |
| Stripped report | `ari sync` stdout | Lists 10 stripped legacy entries |
| CC session start | Launch `claude` | Clean startup, no missing .sh errors |

#### autom8 (10 legacy + 0 ari + 0 missing .sh)

| Check | Command/Method | Expected |
|-------|---------------|----------|
| Exit code | `ari sync` exit code | 0 |
| Legacy hooks gone | `grep -c "CLAUDE_PROJECT_DIR" .claude/settings.local.json` | 0 matches |
| Ari hooks present | `grep -c "ari hook" .claude/settings.local.json` | 10 (first-time ari hook install) |
| CC session start | Launch `claude` | Clean startup |

---

## Integration Test Matrix

### B1: Legacy Hook Merge Tests

Tests go in `internal/materialize/hooks_test.go`.

| Test Name | Satellite Type | Input | Expected Outcome |
|-----------|---------------|-------|------------------|
| `TestMergeHooks_MixedLegacyAndAri` | platform/asana | Existing settings with 12 legacy + 10 ari hooks per event | Output has 10 ari hooks (from hooks.yaml), 0 legacy, 0 user. Stripped report lists 12 entries. |
| `TestMergeHooks_LegacyOnly` | data/autom8 | Existing settings with 10 legacy hooks, 0 ari hooks | Output has 10 ari hooks (from hooks.yaml), 0 legacy. Stripped report lists 10 entries. |
| `TestMergeHooks_CleanSlate` | new install | Empty settings (no hooks key) | Output has 10 ari hooks. Stripped report is empty. |
| `TestMergeHooks_UserHookPreserved` | hypothetical | Existing settings with 1 user hook (command: "my-custom-tool check") + 5 legacy | Output has 10 ari hooks + 1 user hook. Stripped report lists 5 legacy entries. |
| `TestMergeHooks_FlatFormatLegacy` | backward compat | Existing settings with old flat format legacy hooks (command at top level) | Legacy hooks detected and stripped via flat format check. |
| `TestIsLegacyPlatformHook_Patterns` | unit | Individual matcher groups with each legacy pattern | Each pattern correctly identified. |
| `TestIsLegacyPlatformHook_NonLegacy` | unit | Matcher groups with non-legacy, non-ari commands | Returns false (not identified as legacy). |

### B2: Stale Reference Tests

No automated integration tests needed -- this is a string replacement verified by grep.

**Verification command** (run as part of CI or manually):

```bash
# Must return 0 matches
grep -r "ari sync" internal/ README.md INTERVIEW_SYNTHESIS.md

# May return matches only in mena/cem/sync.dro.md (intentional legacy note)
grep -r "ari sync" mena/
```

### B3: Soft Switch Tests

Tests go in `internal/materialize/materialize_test.go` or a new `materialize_soft_test.go`.

| Test Name | Input | Expected Outcome |
|-----------|-------|------------------|
| `TestMaterializeWithOptions_SoftMode_AgentsUpdated` | Soft=true, valid rite | agents/ directory updated with rite agents |
| `TestMaterializeWithOptions_SoftMode_CLAUDEmdUpdated` | Soft=true, valid rite | CLAUDE.md updated with new rite sections |
| `TestMaterializeWithOptions_SoftMode_MenaSkipped` | Soft=true, valid rite | commands/ and skills/ directories unchanged from pre-sync state |
| `TestMaterializeWithOptions_SoftMode_SettingsSkipped` | Soft=true, valid rite | settings.local.json unchanged from pre-sync state |
| `TestMaterializeWithOptions_SoftMode_ActiveRiteUpdated` | Soft=true, valid rite | ACTIVE_RITE file contains new rite name |
| `TestMaterializeWithOptions_SoftMode_ProvenanceCarryForward` | Soft=true, pre-existing provenance manifest | Provenance manifest retains entries for skipped stages from previous manifest |
| `TestMaterializeWithOptions_SoftMode_ResultReporting` | Soft=true | result.SoftMode=true, result.DeferredStages has 4 entries |
| `TestMaterializeWithOptions_FullMode_Unchanged` | Soft=false | All stages run (regression guard) |

---

## Implementation Order

**Order: B2 -> B1 -> B3**

This matches the gap analysis recommendation.

### Sprint 1: B2 -- Stale Reference Cleanup

**Scope**: String replacements only. No logic changes. Independently deployable.

**Files to touch**:
1. `internal/cmd/rite/pantheon.go:46` -- 1 string
2. `internal/cmd/provenance/provenance.go:132` -- 1 string
3. `README.md:17,23,93` -- 3 strings
4. `INTERVIEW_SYNTHESIS.md:173` -- 1 string
5. `docs/**/*.md` -- bulk replacement (~170 references)

**Verification**: `grep -r "ari sync" internal/ README.md INTERVIEW_SYNTHESIS.md` returns 0 results.

**DO NOT touch**: `mena/cem/sync.dro.md` (intentional legacy documentation).

### Sprint 2: B1 -- Provenance-Aware Hook Merge

**Scope**: New `isLegacyPlatformHook()` function, modified `mergeHooksSettings()`, test coverage.

**Files to touch**:
1. `internal/materialize/hooks.go` -- new function + modified merge logic + modified signature
2. `internal/materialize/materialize.go:1397` -- handle new return value
3. `internal/materialize/hooks_test.go` -- 7 new test cases

**Verification**: All tests pass. Manual test on one satellite (autom8y_platform) confirms legacy hooks stripped.

### Sprint 3: B3 -- Soft Rite Switch

**Scope**: New `--soft` flag, conditional pipeline stages, result reporting.

**Files to touch**:
1. `internal/materialize/sync_types.go` -- add `Soft` field to `SyncOptions`
2. `internal/materialize/materialize.go` -- add `Soft` field to `Options` and `Result`, conditional in `MaterializeWithOptions()`, wire in `syncRiteScope()`
3. `internal/cmd/sync/sync.go` -- add `--soft` flag, wire to options, update result formatting
4. `internal/materialize/materialize_test.go` or new `materialize_soft_test.go` -- 8 test cases

**Verification**: `ari sync --rite ecosystem --soft` from CC Bash tool exits 0 within 10 seconds. CLAUDE.md and agents/ updated. Commands/skills/settings unchanged.

### Sprint 4: Deployment

**Scope**: Build, distribute, verify across 4 satellites.

1. Rebuild ari: `CGO_ENABLED=0 go build ./cmd/ari && cp ./ari $(which ari)`
2. Sync each satellite per Area 4 verification checklist
3. Confirm CC session startup in each satellite

---

## Handoff Notes for Integration Engineer

### Explicit Implementation Constraints

**DO NOT**:
- Modify provenance manifest schema (PROVENANCE_MANIFEST.yaml format unchanged)
- Add automatic CC detection (no `CLAUDE_PROJECT_DIR` env var sniffing in sync path)
- Change `writeIfChanged()` or `AtomicWriteFile()` behavior
- Touch any file in `mena/cem/sync.dro.md` (intentional legacy reference)
- Add `--soft` to user scope (soft mode is rite-scope only)

**DO**:
- Follow existing patterns: `isLegacyPlatformHook()` should mirror `isAriManagedGroup()` structure
- Handle both `[]any` and `[]map[string]any` type assertions (JSON unmarshal vs. in-memory)
- Thread the stripped-hooks report through to stdout only in text output mode (not JSON/YAML)
- Ensure `--soft` flag is ignored for user scope and minimal mode (only applies to `MaterializeWithOptions()`)
- Write tests first for B1 (the legacy pattern detection is safety-critical)

### Key Files Reference

| File | Purpose |
|------|---------|
| `internal/materialize/hooks.go` | Hook merge logic (B1 primary) |
| `internal/materialize/materialize.go` | Pipeline orchestration (B1 caller, B3 primary) |
| `internal/materialize/sync_types.go` | Unified sync types (B3 types) |
| `internal/cmd/sync/sync.go` | CLI wiring (B3 flag) |
| `internal/cmd/rite/pantheon.go` | Stale string (B2) |
| `internal/cmd/provenance/provenance.go` | Stale string (B2) |
| `hooks/hooks.yaml` | Canonical hook definitions (10 entries) |

### Provenance Impact

- B1: No provenance schema changes. `settings.local.json` continues to be tracked as a single file-level entry. The hook merge logic is below provenance granularity.
- B2: No provenance impact (string changes only).
- B3: Provenance manifest in soft mode carries forward entries for skipped stages via existing `saveProvenanceManifest()` carry-forward logic. No changes needed to provenance code.

---

## Attestation

| Artifact | Absolute Path | Status |
|----------|--------------|--------|
| Gap Analysis (input) | `/Users/tomtenuta/Code/knossos/docs/ecosystem/GAP-satellite-distribution-blockers.md` | Read |
| hooks.go (explored) | `/Users/tomtenuta/Code/knossos/internal/materialize/hooks.go` | Read |
| materialize.go (explored) | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | Read |
| sync_types.go (explored) | `/Users/tomtenuta/Code/knossos/internal/materialize/sync_types.go` | Read |
| sync.go CLI (explored) | `/Users/tomtenuta/Code/knossos/internal/cmd/sync/sync.go` | Read |
| pantheon.go (explored) | `/Users/tomtenuta/Code/knossos/internal/cmd/rite/pantheon.go` | Read |
| provenance.go CLI (explored) | `/Users/tomtenuta/Code/knossos/internal/cmd/provenance/provenance.go` | Read |
| fileutil.go (explored) | `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go` | Read |
| hooks.yaml (explored) | `/Users/tomtenuta/Code/knossos/hooks/hooks.yaml` | Read |
| provenance types (explored) | `/Users/tomtenuta/Code/knossos/internal/provenance/provenance.go` | Read |
| README.md (explored) | `/Users/tomtenuta/Code/knossos/README.md` | Read |
| INTERVIEW_SYNTHESIS.md (explored) | `/Users/tomtenuta/Code/knossos/INTERVIEW_SYNTHESIS.md` | Read |
| Context Design (output) | `/Users/tomtenuta/Code/knossos/docs/design/CONTEXT-DESIGN-satellite-hooks-provenance.md` | Written |
