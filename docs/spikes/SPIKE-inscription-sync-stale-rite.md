# SPIKE: Inscription Sync Produces Stale Rite Content After Rite Switch

**Date**: 2026-03-02
**Status**: Complete
**Severity**: P1 (core framework function broken)

## Question

Why do hygiene-rite references persist in `.claude/CLAUDE.md` after syncing the ecosystem rite? Is the materialization pipeline broken?

**Decision this informs**: Whether a code fix is needed in the inscription/materialization pipeline or if this is an operational issue.

---

## Approach

1. Examine current `.claude/` state to catalog discrepancies between ACTIVE_RITE and actual CLAUDE.md content
2. Trace the materialization pipeline end-to-end (`ari sync` -> `MaterializeWithOptions` -> `materializeCLAUDEmd` -> `SyncCLAUDEmd`)
3. Attempt to reproduce the failure with the installed binary vs freshly built binary
4. Identify root cause and document

---

## Findings

### Observed State (Pre-Fix)

| File | Expected (ecosystem) | Actual |
|------|---------------------|--------|
| `.claude/ACTIVE_RITE` | `ecosystem` | `ecosystem` (correct) |
| `.claude/agents/` | ecosystem agents | ecosystem agents (correct) |
| `.claude/CLAUDE.md` | ecosystem content | **hygiene content** (stale) |
| `.knossos/KNOSSOS_MANIFEST.yaml` active_rite | `ecosystem` | **`hygiene`** (stale) |
| `.knossos/sync/state.json` active_rite | `ecosystem` | **`hygiene`** (stale) |
| `.claude/ACTIVE_WORKFLOW.yaml` | ecosystem workflow | **hygiene workflow** (stale) |
| `.claude/PROVENANCE_MANIFEST.yaml` active_rite | `ecosystem` | **`hygiene`** (stale) |

### Timestamp Analysis

| File | Modification Epoch | Interpretation |
|------|-------------------|----------------|
| KNOSSOS_MANIFEST.yaml | 1772403018 | Mar 1 22:10 UTC -- last full hygiene sync |
| CLAUDE.md | 1772403018 | Same as manifest -- coherent with hygiene sync |
| sync/state.json | 1772403018 | Same -- coherent |
| ACTIVE_RITE | 1772410900 | ~2.2 hours LATER -- updated independently |
| agents/*.md | Mar 2 01:21 local | Even later -- agents swapped to ecosystem |

The ACTIVE_RITE and agents were updated AFTER the last successful full sync (which was hygiene). Something partially updated the project state to ecosystem without completing the inscription sync.

### Root Cause: Stale Installed Binary

The installed `ari` binary at `/opt/homebrew/bin/ari` has a bug in the CLAUDE.md materialization path that causes `materializeCLAUDEmd` to fail:

```
$ ari sync --rite=ecosystem --scope=rite
Error: failed to materialize CLAUDE.md
```

A freshly built binary from current source succeeds:

```
$ CGO_ENABLED=0 go build -o /tmp/ari-debug ./cmd/ari
$ /tmp/ari-debug sync --rite=ecosystem --scope=rite
Sync: success
  Rite: success (ecosystem)
```

### The Partial Update Mechanism

When `ari sync --rite=ecosystem` is run with `--scope=all` (default), the pipeline:

1. **Phase 1 (Rite Scope)**: Calls `syncRiteScope()` -> `MaterializeWithOptions()`
   - Step 4 (agents): SUCCEEDS -- agents are written to `.claude/agents/`
   - Step 7 (CLAUDE.md): **FAILS** with `"failed to materialize CLAUDE.md"`
   - Steps 8-10: NOT executed (error propagates)

2. **Error swallowed**: In `Sync()` (line 507-517), when `scope=all`, a rite scope error is caught and the rite result is set to "skipped" -- **the error is not surfaced to the user**

3. **Phase 2 (User Scope)**: Runs successfully regardless of rite scope failure

The result: agents are updated (step 4 happened before the failure), but CLAUDE.md, KNOSSOS_MANIFEST, sync state, workflow, and provenance are NOT updated. The ACTIVE_RITE file is NOT written either (step 10 never runs).

The ACTIVE_RITE file was likely updated by a **separate subsequent operation** (e.g., `ari worktree switch --update-rite`, manual write, or a different sync attempt).

### Two Bugs Identified

**Bug 1: Silent error swallowing in `Sync()` scope=all mode**

File: `internal/materialize/materialize.go`, lines 507-517

```go
if opts.Scope == ScopeAll || opts.Scope == ScopeRite {
    riteResult, err := m.syncRiteScope(opts)
    if err != nil {
        if opts.Scope == ScopeRite {
            return nil, err  // Only surfaces error in rite-only mode
        }
        // scope=all: skip rite, continue to user -- ERROR LOST
        result.RiteResult = &RiteScopeResult{Status: "skipped"}
    }
}
```

The user sees `Rite: skipped` instead of the actual error. This makes diagnosis nearly impossible.

**Bug 2: Partial state in `MaterializeWithOptions()` on inscription failure**

File: `internal/materialize/materialize.go`, lines 292-468

The pipeline writes agents (step 4) before attempting CLAUDE.md (step 7). When step 7 fails, agents are already on disk but the error prevents ACTIVE_RITE, sync state, and provenance from being updated. The result is an inconsistent `.claude/` directory.

### Secondary Finding: IsKnossosProject Detection

After fixing with the fresh binary, CLAUDE.md renders with **satellite-style templates** instead of knossos-specific ones. The `IsKnossosProject` detection at line 1307:

```go
IsKnossosProject: m.templatesDir != "" && strings.HasPrefix(m.templatesDir, projectRoot),
```

This depends on `m.templatesDir` being set correctly. When the rite resolves templates from a different source (e.g., embedded or KNOSSOS_HOME), this detection fails. This is a separate issue from the inscription sync breakage but worth noting as it changes the CLAUDE.md output for the knossos project itself.

---

## Recommendation

### Immediate Fix (unblock)

Run sync with the freshly built binary:

```bash
CGO_ENABLED=0 go build ./cmd/ari && cp ./ari $(which ari)
ari sync --rite=ecosystem
```

**This has already been done during the spike** using `/tmp/ari-debug`. The CLAUDE.md now shows correct ecosystem content.

### Bug Fixes Needed

1. **Error surfacing in scope=all**: When rite scope fails in `scope=all` mode, propagate the error (or at minimum surface it in the output as a warning, not silently "skipped"). The current behavior masks real failures.

2. **Atomicity in MaterializeWithOptions()**: Consider a two-phase approach where destructive writes (agents, ACTIVE_RITE) only happen after all non-destructive steps (CLAUDE.md generation) have been validated. Alternatively, roll back agents if a subsequent step fails.

3. **IsKnossosProject detection**: This should be based on project structure (does `knossos/templates/` exist in the project root?) rather than runtime `templatesDir` path prefix matching.

### Operational Lesson

The MEMORY.md already warns about the stale binary trap:

> After rebuilding, MUST also `cp ./ari $(which ari)` to update the installed binary.

This spike confirms that a stale binary can produce silently corrupted `.claude/` state that is extremely difficult to diagnose. The `scope=all` error swallowing makes it worse.

---

## Follow-Up Actions

| Action | Priority | Effort |
|--------|----------|--------|
| Fix error surfacing in `Sync()` scope=all mode | P1 | ~10 LOC |
| Investigate the specific CLAUDE.md failure in the old binary | P2 | Investigation |
| Add atomicity/rollback to `MaterializeWithOptions()` | P2 | ~50 LOC |
| Fix `IsKnossosProject` detection to be structural | P3 | ~10 LOC |
| Add integration test: rite switch verifies all state files updated | P2 | ~40 LOC |

---

## Key Files

| File | Role |
|------|------|
| `internal/materialize/materialize.go` | `Sync()`, `syncRiteScope()`, `MaterializeWithOptions()` |
| `internal/inscription/sync.go` | `SyncCLAUDEmd()` -- canonical CLAUDE.md generation |
| `internal/inscription/generator.go` | Template rendering with `RenderContext` |
| `internal/inscription/pipeline.go` | `Pipeline.buildRenderContext()`, agent loading |
| `knossos/templates/sections/quick-start.md.tpl` | Quick-start template using `.ActiveRite` |
| `knossos/templates/sections/agent-configurations.md.tpl` | Agent listing template |
