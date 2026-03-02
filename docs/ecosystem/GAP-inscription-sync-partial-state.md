# Gap Analysis: Inscription Sync -- Partial State, Error Masking, and IsKnossosProject Detection

**Date**: 2026-03-02
**Status**: Complete
**Predecessor**: [SPIKE-inscription-sync-stale-rite.md](../spikes/SPIKE-inscription-sync-stale-rite.md)
**Severity**: P1 (compound defects in core sync pipeline)

## Purpose

Deep trace of the three issues identified in the initial inscription sync spike. This document maps exact file:line references, full surface area of each bug, additional related issues discovered during analysis, and remediation recommendations with effort estimates.

---

## Bug 1: Partial State on Inscription Failure

### Root Cause

**File**: `internal/materialize/materialize.go`
**Function**: `MaterializeWithOptions()` (lines 292-468)

The pipeline executes 10+ steps sequentially. Each step writes to disk immediately. There is no transaction boundary, staging area, or rollback mechanism. When any step fails after prior steps have already written, the `.claude/` directory is left in a partial state.

### Exact Disk Write Sequence

| Step | Line | What it Writes | Can it Fail? |
|------|------|---------------|-------------|
| 2 | 349 | `.claude/` directory (mkdir) | Yes (permissions) |
| Provenance | 355 | Loads prev manifest (read-only) | No (degraded) |
| 2.5 | 370 | Removes `INVOCATION_STATE.yaml` | Unlikely |
| 3 | 375-396 | Orphan backup/removal/promotion | Yes |
| 4 | 407 | `agents/*.md` (all rite agents) | Yes |
| 5 | 413 | `commands/`, `skills/` (mena) | Yes |
| 6 | 420 | `rules/*.md` | Yes |
| **7** | **426** | **`CLAUDE.md` + `KNOSSOS_MANIFEST.yaml`** | **YES -- this is where it failed** |
| 8 | 434 | `settings.local.json` | Yes |
| 9 | 440 | `sync/state.json` | Yes |
| 9.5 | 446 | `ACTIVE_WORKFLOW.yaml` | Yes |
| 10 | 458 | `ACTIVE_RITE` | Yes |
| Provenance | 463 | `PROVENANCE_MANIFEST.yaml` | Yes |

### Full Surface Area of Partial State

When step 7 (`materializeCLAUDEmd`) fails, the following files are in an **inconsistent state**:

**Already written (stale new rite content):**
- `.claude/agents/*.md` -- agents from the NEW rite are on disk
- `.claude/commands/*` -- mena from the NEW rite (if not `--soft`)
- `.claude/skills/*` -- mena from the NEW rite (if not `--soft`)
- `.claude/rules/*.md` -- rules from the NEW rite (if not `--soft`)
- `INVOCATION_STATE.yaml` -- removed (step 2.5)
- Orphan agents -- may have been removed/promoted (step 3)

**Never updated (stale OLD rite content):**
- `.claude/CLAUDE.md` -- still references the OLD rite
- `.claude/KNOSSOS_MANIFEST.yaml` -- still says the OLD rite is active
- `.claude/sync/state.json` -- still records the OLD rite
- `.claude/ACTIVE_WORKFLOW.yaml` -- still has the OLD rite's workflow
- `.claude/ACTIVE_RITE` -- still says the OLD rite (or does not exist)
- `.claude/PROVENANCE_MANIFEST.yaml` -- never updated
- `.claude/settings.local.json` -- never updated with new MCP servers

**Destroyed and unrecoverable (without backup):**
- Orphan agents from the OLD rite that were removed at step 3 (if `--remove-all`)

### Additional Failure Points Not in Initial Spike

The initial spike focused on step 7 failure. But ANY step from 4 onwards creates the same partial state problem:

1. **Step 5 failure (mena)**: Agents are on disk but commands/skills are incomplete. The `SyncMena()` function at line 1116 uses `MenaProjectionDestructive` mode which removes stale files BEFORE writing new ones. If it fails mid-write, both old and new mena entries are partially present.

2. **Step 8 failure (settings)**: Everything through CLAUDE.md is written, but `settings.local.json` still has old MCP servers and hooks. This means the CLAUDE.md references agents that may need MCP servers that are not configured.

3. **Step 9 failure (state tracking)**: All content is correct on disk but `sync/state.json` is stale. The next `ari sync` without `--rite` will read the stale state and may attempt the wrong rite, or report incorrect status.

4. **Provenance failure (last step)**: Everything is correct but provenance is stale. The divergence detection on the NEXT sync will incorrectly flag all newly-written files as "diverged from last known state."

### Soft Mode Interaction

When `opts.Soft` is true (lines 412-416, 419-423, 433-437, 445-449), steps 5, 6, 8, and 9.5 are skipped. This means a soft-mode sync only writes agents (step 4) and CLAUDE.md (step 7). The partial state window is smaller but still exists: if step 7 fails, agents are on disk but CLAUDE.md is not updated.

### Severity Assessment

The partial state is **not self-healing**. A subsequent `ari sync` with the same rite will succeed and complete all steps, BUT:
- The user may not know a partial state exists (see Bug 2)
- Orphan agents removed at step 3 cannot be recovered by a retry
- The window between failure and retry leaves an inconsistent `.claude/` directory that Claude Code is actively reading

---

## Bug 2: Silent Error Swallowing in Sync()

### Root Cause

**File**: `internal/materialize/materialize.go`
**Function**: `Sync()` (lines 486-538)

### Rite Scope Error Handling (lines 506-518)

```go
// Phase 1: Rite scope
if opts.Scope == ScopeAll || opts.Scope == ScopeRite {
    riteResult, err := m.syncRiteScope(opts)
    if err != nil {
        if opts.Scope == ScopeRite {
            return nil, err  // <-- ONLY surfaces error when scope=rite
        }
        // scope=all: skip rite, continue to user
        result.RiteResult = &RiteScopeResult{Status: "skipped"}  // <-- ERROR LOST
    } else {
        result.RiteResult = riteResult
    }
}
```

**Line 514**: The error is discarded entirely. The `RiteScopeResult` has no `Error` or `Errors` field. The caller (CLI command) sees `Status: "skipped"` with no indication of what went wrong.

### User Scope Error Handling (lines 520-535)

```go
// Phase 2: User scope
if opts.Scope == ScopeAll || opts.Scope == ScopeUser {
    userResult, err := m.syncUserScope(opts)
    if err != nil {
        if opts.Scope == ScopeUser {
            return nil, err // hard fail only if explicitly user-only
        }
        // scope=all: log and skip, don't block rite results
        result.UserResult = &UserScopeResult{
            Status: "skipped",
            Errors: []UserResourceError{{Resource: ResourceAll, Err: err.Error()}},
        }
    } else {
        result.UserResult = userResult
    }
}
```

**Asymmetry**: The user scope error handling at lines 528-532 is BETTER than the rite scope handling. It preserves the error in `UserScopeResult.Errors`. But the rite scope has no `Errors` field at all.

### Impact Chain

1. `ari sync --rite=ecosystem` (default scope=all)
2. Rite scope fails (e.g., CLAUDE.md generation error)
3. `Sync()` catches error, sets `RiteResult = {Status: "skipped"}`
4. User scope runs successfully
5. CLI reports: `Rite: skipped, User: success`
6. User sees "skipped" -- assumes no rite was active, not that it errored
7. `.claude/` is in partial state (Bug 1) and nobody knows

### RiteScopeResult Missing Error Field

The `RiteScopeResult` struct lacks an `Errors` field. Compare with `UserScopeResult`:

```go
type UserScopeResult struct {
    Status string
    Errors []UserResourceError  // <-- has error tracking
    // ...
}
```

vs.

```go
type RiteScopeResult struct {
    Status string  // "success", "skipped", "minimal"
    // ... NO Errors field
}
```

This is a structural gap: even if `Sync()` wanted to pass through the error, the result type cannot carry it.

### Additional Error Masking Locations

Beyond the `Sync()` function, there are multiple warning-only error paths inside `MaterializeWithOptions()`:

| Line | Pattern | What is Lost |
|------|---------|-------------|
| 250 | `log.Printf("Warning: failed to load provenance manifest...")` | Provenance load failure degraded to log warning |
| 258 | `log.Printf("Warning: failed to detect provenance divergence...")` | Divergence detection failure degraded to log warning |
| 323 | `log.Printf("Warning: %s: %s (%s)")` | Rite reference validation warnings are log-only |
| 860 | `log.Printf("Warning: agent transform failed...")` | Agent frontmatter transform failure is non-fatal |
| 911 | `log.Printf("Warning: agent transform failed...")` | Same, for embedded agents |
| 963 | `log.Printf("Warning: agent transform failed...")` | Same, for filesystem agents |
| 1002 | `log.Printf("Warning: agent '%s' declared in manifest but no .md file found...")` | Phantom agent is warning-only |

These `log.Printf` calls go to stderr, which the CLI may or may not surface depending on invocation context (e.g., suppressed when called from a hook).

---

## Bug 3: IsKnossosProject Detection Fragility

### Root Cause

**File**: `internal/materialize/materialize.go`
**Lines**: 1307, 1360

```go
IsKnossosProject: m.templatesDir != "" && strings.HasPrefix(m.templatesDir, projectRoot),
```

### Full Resolution Trace for Knossos Repo

When `ari sync --rite=ecosystem` runs from `/Users/tomtenuta/Code/knossos`:

1. **NewMaterializer** (line 117): `m.templatesDir = "/Users/tomtenuta/Code/knossos/templates"` -- this directory does NOT exist on disk
2. **SourceResolver.ResolveRite** (resolver.go line 61): walks priority chain:
   - SourceProject: `.knossos/rites/ecosystem` -- does NOT exist, skip
   - SourceUser: `~/.local/share/knossos/rites/ecosystem` -- does NOT exist, skip
   - SourceKnossos: resolves `KNOSSOS_HOME` = `~/Code/knossos`, checks `~/Code/knossos/rites/ecosystem/manifest.yaml` -- EXISTS
3. **checkSource for SourceKnossos** (resolver.go lines 228-231):
   ```go
   case SourceKnossos:
       templatesDir = filepath.Join(filepath.Dir(source.Path), "knossos", "templates")
   ```
   `source.Path` = `~/Code/knossos/rites`, so `templatesDir = ~/Code/knossos/knossos/templates`
4. **Line 328-330** of MaterializeWithOptions:
   ```go
   if resolved.TemplatesDir != "" {
       m.templatesDir = resolved.TemplatesDir
   }
   ```
   `m.templatesDir` is now `"/Users/tomtenuta/Code/knossos/knossos/templates"`
5. **Line 1307**: `strings.HasPrefix("/Users/tomtenuta/Code/knossos/knossos/templates", "/Users/tomtenuta/Code/knossos")` = **TRUE**

**Current behavior**: For the standard developer setup where KNOSSOS_HOME = projectRoot, the detection works correctly.

### When Detection Fails

The detection breaks in these scenarios:

1. **KNOSSOS_HOME set to a different directory**: If `KNOSSOS_HOME=/opt/knossos` and projectRoot is `/Users/dev/knossos`, the rite resolves from `/opt/knossos/rites/` and `templatesDir` becomes `/opt/knossos/knossos/templates`. The `HasPrefix` check against `/Users/dev/knossos` fails. Result: **knossos renders satellite-style CLAUDE.md**.

2. **Symlinked paths**: If KNOSSOS_HOME or projectRoot uses a symlink, the string prefix check may fail even when they point to the same physical directory.

3. **Satellite project**: Working in `/Users/dev/my-satellite`, KNOSSOS_HOME = `~/Code/knossos`. Rite resolves from `~/Code/knossos/rites/`. `templatesDir` = `~/Code/knossos/knossos/templates`. `HasPrefix(templatesDir, "/Users/dev/my-satellite")` = **FALSE**. This is the CORRECT behavior for satellites -- they are NOT knossos projects.

### Template and Rule Filter Coupling

The `HasPrefix` pattern appears THREE times in materialize.go:

| Line | Usage | Effect when Wrong |
|------|-------|-------------------|
| 1187 | `materializeRules()` skip guard | Skips internal rules on satellites; copies them on knossos. If wrong: knossos gets satellite behavior (no dev rules) or satellite gets knossos internal rules |
| 1307 | `materializeCLAUDEmd()` render context | Controls 7 template conditionals across 6 section templates |
| 1360 | `materializeMinimalCLAUDEmd()` render context | Same as 1307 but for minimal/cross-cutting mode |

### IsKnossosProject Template Impact

When `IsKnossosProject` is false (satellite mode), these sections change:

| Template | Knossos Version | Satellite Version |
|----------|----------------|-------------------|
| `execution-mode.md.tpl` | 3-mode table (Native/Cross-Cutting/Orchestrated) | Single sentence: "Use the available agents..." |
| `quick-start.md.tpl` | Footer: `/go`, `prompting` skill, `/consult` | Footer: "Delegate to specialists via Task tool." |
| `quick-start.md.tpl` | No-rite: includes `/go` | No-rite: omits `/go` |
| `commands.md.tpl` | 5-column table with Knossos Name column | 4-column table without Knossos Name |
| `platform-infrastructure.md.tpl` | Entry/Sessions/Hooks details | Single line: `ari --help` |
| `agent-routing.md.tpl` | Includes Pythia coordination + `/task` + `/consult` | Omits Pythia and routing commands |
| `know.md.tpl` | Includes `.know/literature-{domain}.md` entry | Omits literature line |

**Generator.go defaults** (lines 500-589): The same branching exists in 6 `getDefault*Content()` methods, which are fallbacks when template files are not found. These duplicate the template logic in Go code.

---

## Additional Issues Discovered

### Issue 4: materializeMena Uses Destructive Mode Without Rollback

**File**: `internal/materialize/materialize.go`, line 1107

```go
opts := MenaProjectionOptions{
    Mode: MenaProjectionDestructive,
    // ...
}
```

The `MenaProjectionDestructive` mode removes stale files from `commands/` and `skills/` before writing new ones. If the write phase fails after the delete phase, both old and new entries are partially present. This is the same partial-state pattern as Bug 1 but at a finer granularity.

### Issue 5: Provenance Warning Suppression

**File**: `internal/materialize/materialize.go`, lines 250-259, 357-366

Provenance manifest load failures and divergence detection failures are degraded to `log.Printf` warnings. These warnings:
- Go to stderr only
- Are not captured in any result struct
- May be suppressed in hook or non-interactive contexts
- Can indicate data corruption that silently degrades future syncs

### Issue 6: materializeRules Template Dir Guard Uses Same Fragile Pattern

**File**: `internal/materialize/materialize.go`, line 1187

```go
if m.templatesDir != "" && !strings.HasPrefix(m.templatesDir, projectRoot) {
```

This is the inverse of the `IsKnossosProject` check. When this guard is wrong (templatesDir appears external but is actually the project's own templates), knossos-internal rules get SKIPPED during sync on the knossos project itself. The rules (`.claude/rules/*.md`) provide trigger-based instructions for CC, and missing them degrades agent behavior.

### Issue 7: Agent Transform Failures Are Non-Fatal

**File**: `internal/materialize/materialize.go`, lines 857-861, 908-912, 960-964

When `transformAgentContent()` fails (e.g., frontmatter parse error, hook resolution failure, skill policy evaluation error), the untransformed content is written to disk instead. This means:
- Write guards may be missing from the agent
- Skill preloading may not be wired
- Agent defaults may not be merged
- The agent functions but with degraded capabilities, and the only indication is a stderr warning

---

## State File Inventory

Complete list of files that must be mutually consistent after a successful sync:

| File | Updated At Step | Content Dependency |
|------|----------------|--------------------|
| `.claude/agents/*.md` | 4 | Rite manifest agents list |
| `.claude/commands/*` | 5 | Mena sources (platform + shared + deps + rite) |
| `.claude/skills/*` | 5 | Mena sources (platform + shared + deps + rite) |
| `.claude/rules/*.md` | 6 | Template rules (knossos-internal or none) |
| `.claude/CLAUDE.md` | 7 | Template rendering with IsKnossosProject + agent data |
| `.claude/KNOSSOS_MANIFEST.yaml` | 7 (via SyncCLAUDEmd) | Active rite, section hashes, inscription version |
| `.claude/settings.local.json` | 8 | MCP servers from manifest + hooks.yaml |
| `.claude/sync/state.json` | 9 | Active rite name + last sync timestamp |
| `.claude/ACTIVE_WORKFLOW.yaml` | 9.5 | Workflow definition from rite |
| `.claude/ACTIVE_RITE` | 10 | Rite name marker |
| `.claude/PROVENANCE_MANIFEST.yaml` | last | All file checksums + ownership + source info |

**11 distinct state files** that must all reflect the same rite. Currently, the pipeline writes them one at a time with no atomicity guarantee.

---

## Remediation Recommendations

### R1: Add Error Field to RiteScopeResult and Surface It

**Priority**: P1 (blocking -- makes all other bugs invisible)
**Effort**: ~15 LOC
**Complexity**: PATCH

Add an `Errors` field to `RiteScopeResult` mirroring `UserScopeResult.Errors`. In `Sync()` line 514, populate it with the actual error instead of discarding it. Update CLI output formatting to display rite scope errors.

**Files to modify**:
- `internal/materialize/materialize.go`: Add `Errors` field to `RiteScopeResult`, populate at line 514
- `internal/materialize/types.go` (or wherever `RiteScopeResult` is defined): struct change
- `internal/cmd/sync/sync.go` (or CLI handler): Display errors in output

### R2: Two-Phase Write with Validation Gate

**Priority**: P2 (prevents the most common partial-state scenario)
**Effort**: ~30-120 LOC depending on approach
**Complexity**: MODULE (full two-phase) or PATCH (step reorder)

**Simpler approach (~30 LOC)**: Move agent writes (step 4) AFTER CLAUDE.md generation (step 7). CLAUDE.md generation is the most failure-prone step (template parsing, merge logic). If it is validated first, the subsequent disk writes are unlikely to fail. This reorders the function body without changing any interfaces.

**Full approach (~120 LOC)**: Insert a validation gate between "generate content" and "write to disk". Steps 4-7 generate content into memory, validation confirms all generation succeeded, then a single write phase commits everything. This requires refactoring each `materialize*` function to return content rather than write directly.

**Files to modify**:
- `internal/materialize/materialize.go`: Reorder steps in `MaterializeWithOptions()`

### R3: Structural IsKnossosProject Detection

**Priority**: P3 (cosmetic for current setup, correctness for edge cases)
**Effort**: ~15 LOC
**Complexity**: PATCH

Replace the `strings.HasPrefix(m.templatesDir, projectRoot)` heuristic with a structural check:

```go
IsKnossosProject: fileExists(filepath.Join(projectRoot, "knossos", "templates", "sections")) ||
                  fileExists(filepath.Join(projectRoot, "rites"))
```

This checks whether the project IS knossos by looking for knossos-specific directory structure, independent of how `templatesDir` was resolved.

**Files to modify**:
- `internal/materialize/materialize.go`: Lines 1307, 1360, 1187

### R4: Promote Agent Transform Warnings to Result

**Priority**: P3 (degraded agent behavior is hard to diagnose)
**Effort**: ~25 LOC
**Complexity**: PATCH

Add a `Warnings []string` field to `Result`. When `transformAgentContent()` fails, append to warnings instead of (or in addition to) `log.Printf`. The CLI can then display: `Sync: success (3 warnings)`.

**Files to modify**:
- `internal/materialize/materialize.go`: Add warnings collection, pass through to Result

### R5: Provenance Warning Promotion

**Priority**: P4 (degraded future syncs, but non-blocking)
**Effort**: ~10 LOC
**Complexity**: PATCH

Same as R4 but for provenance load/divergence warnings at lines 250, 258, 357, 365.

---

## Priority Ordering

| Priority | Recommendation | Rationale |
|----------|---------------|-----------|
| **P1** | R1: Error surfacing | Without this, ALL other bugs are invisible to the user. Must be fixed first. |
| **P2** | R2: Step reordering | Prevents the most common partial-state scenario (CLAUDE.md failure). Simpler version is ~30 LOC. |
| **P3** | R3: Structural detection | Correctness fix. Current setup works by coincidence; will break when deployment model changes. |
| **P3** | R4: Transform warnings | Visibility for degraded agent behavior. |
| **P4** | R5: Provenance warnings | Nice-to-have; provenance degradation is rare and self-healing on next full sync. |

---

## Test Satellites for Verification

| Satellite | Configuration | Verifies |
|-----------|--------------|----------|
| knossos repo itself | Self-hosting, KNOSSOS_HOME = projectRoot | R2, R3 (IsKnossosProject = true) |
| A satellite project | KNOSSOS_HOME != projectRoot | R3 (IsKnossosProject = false) |
| Simulated failure | Mock template that errors during rendering | R1, R2 (partial state + error surfacing) |
| Fresh clone / worktree | No `.claude/` pre-existing | R2 (clean slate, no orphan complications) |

---

## Key Files Reference

| File | Absolute Path | Role |
|------|--------------|------|
| materialize.go | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | `Sync()`, `MaterializeWithOptions()`, `materializeCLAUDEmd()` |
| resolver.go | `/Users/tomtenuta/Code/knossos/internal/materialize/source/resolver.go` | `ResolveRite()`, `checkSource()` template dir resolution |
| generator.go | `/Users/tomtenuta/Code/knossos/internal/inscription/generator.go` | `IsKnossosProject` conditionals in defaults + templates |
| sync.go | `/Users/tomtenuta/Code/knossos/internal/inscription/sync.go` | `SyncCLAUDEmd()` canonical CLAUDE.md generation |
| user_scope.go | `/Users/tomtenuta/Code/knossos/internal/materialize/user_scope.go` | Delegates to userscope sub-package |
| execution-mode.md.tpl | `/Users/tomtenuta/Code/knossos/knossos/templates/sections/execution-mode.md.tpl` | IsKnossosProject template conditional |
| quick-start.md.tpl | `/Users/tomtenuta/Code/knossos/knossos/templates/sections/quick-start.md.tpl` | IsKnossosProject template conditional (3 uses) |
| commands.md.tpl | `/Users/tomtenuta/Code/knossos/knossos/templates/sections/commands.md.tpl` | IsKnossosProject template conditional |
| platform-infrastructure.md.tpl | `/Users/tomtenuta/Code/knossos/knossos/templates/sections/platform-infrastructure.md.tpl` | IsKnossosProject template conditional |
| agent-routing.md.tpl | `/Users/tomtenuta/Code/knossos/knossos/templates/sections/agent-routing.md.tpl` | IsKnossosProject template conditional (2 uses) |
| know.md.tpl | `/Users/tomtenuta/Code/knossos/knossos/templates/sections/know.md.tpl` | IsKnossosProject template conditional |
