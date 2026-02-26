---
type: spike
status: complete
date: 2026-02-26
---

# SPIKE: gopls LSP Plugin Stale Diagnostics in Claude Code

## Question

Why does the gopls-lsp plugin for Claude Code consistently report false-positive compilation errors -- diagnostics that show "compilation errors" (unused imports, undefined references) which turn out to be stale, not reflecting the actual current state of the code on disk?

## Context

The observed pattern:
1. Subagents (e.g., integration-engineer) edit multiple Go files across packages in a sprint
2. After subagent completion, Claude Code surfaces: "Found 21 new diagnostic issues in 6 files"
3. The main thread investigates and discovers the diagnostics are stale -- the code compiles and tests pass
4. This wastes main-thread turns: reading files, reasoning about non-existent errors, running `go build` to verify

The specific incident in question: subagents migrated hook files to use a new `internal/registry` package. The diagnostics claimed unused imports and undefined references in 6 files, but `go build` and `go test` both passed cleanly.

## Approach

Traced the full diagnostic pipeline from first principles:

1. **Plugin architecture**: Examined `~/.claude/settings.json` (enabledPlugins), `~/.claude/plugins/installed_plugins.json`, plugin cache structure
2. **LSP protocol lifecycle**: Read CC debug logs to trace `initialize`, `didOpen`, `didChange`, `didSave`, `publishDiagnostics` message flow
3. **Diagnostic delivery mechanism**: Traced the PASSIVE DIAGNOSTICS handler, async registration, and `getLSPDiagnosticAttachments` consumption pattern
4. **Subagent interaction**: Analyzed how subagent file edits interact with the single gopls instance

## Findings

### Architecture

The diagnostic pipeline has these components:

```
gopls-lsp plugin (declarative, no code -- just README.md)
    |
    v
Claude Code LSP Manager (built into CC binary)
    |-- Starts gopls as subprocess
    |-- Sends textDocument/didOpen, didChange, didSave
    |-- Receives textDocument/publishDiagnostics
    |
    v
PASSIVE DIAGNOSTICS Handler
    |-- Registers diagnostics for "async delivery"
    |-- getLSPDiagnosticAttachments polls pending queue
    |-- Diagnostics injected into next API request as context
```

### Root Cause Analysis

The false-positive diagnostics stem from a **fundamental timing mismatch** between gopls's analysis model and Claude Code's multi-file editing pattern. There are three contributing factors:

#### 1. Incremental Analysis with Cross-Package Dependencies

gopls uses **incremental analysis** -- when it receives a `didSave` for file A, it re-analyzes file A and its immediate dependents. But in a multi-file sprint, the subagent edits files in sequence:

```
T=0:  Edit budget.go       (adds `import "registry"`)
T=1:  gopls receives didSave for budget.go
T=2:  gopls analyzes budget.go -- registry package doesn't exist yet
T=3:  gopls publishes diagnostics: "could not import registry" (VALID at this instant)
T=4:  Edit registry.go     (creates the new package)
T=5:  gopls receives didSave for registry.go
T=6:  gopls re-analyzes registry.go and budget.go -- NOW valid
T=7:  gopls publishes empty diagnostics for budget.go (clearing the error)
```

The problem: **step 7 may not complete before diagnostics from step 3 are consumed**. Claude Code's diagnostic attachment system batches and queues diagnostics asynchronously. If the attachment is consumed between steps 3 and 7, the stale diagnostic is injected into the model context.

#### 2. Per-File didSave Notifications Without Workspace-Level Batching

From the debug logs, Claude Code sends LSP notifications **per-file, immediately after each Write/Edit tool completes**:

```
[DEBUG] File write_test.go written atomically
[DEBUG] Sending notification 'textDocument/didSave'
[DEBUG] Sending notification 'textDocument/didOpen'    -- for new files
[DEBUG] Hook PostToolUse:Write success
```

There is no mechanism to batch multiple file edits and send a single workspace-level "these N files all changed" notification. This means gopls sees edits arrive one at a time and publishes intermediate diagnostics after each one.

In contrast, an IDE like VS Code typically batches workspace changes and has a concept of "wait for gopls to quiesce" before displaying diagnostics.

#### 3. Subagent Boundary Creates a Diagnostic Collection Window

The most insidious factor: diagnostics are registered during the subagent's execution but **consumed by the main thread after the subagent returns**. The timeline:

```
Main thread spawns subagent
    Subagent edits file A  --> gopls publishes diags (intermediate state)
    Subagent edits file B  --> gopls publishes diags (intermediate state)
    Subagent edits file C  --> gopls publishes diags (intermediate state)
    Subagent edits file D  --> gopls publishes diags (final state, maybe clean)
    ...gopls is still re-analyzing...
Subagent returns
    <-- gopls may still be publishing updated diagnostics
Main thread calls getLSPDiagnosticAttachments
    Consumes whatever diagnostics were registered so far
    May include stale intermediate diagnostics not yet superseded
```

The critical race: gopls's re-analysis of dependent files (especially across package boundaries) takes 200ms-2s depending on package graph size. If the main thread polls diagnostics before gopls has finished propagating the final state, it gets stale results.

### Evidence from Debug Logs

The debug logs confirm:

1. **didSave triggers immediate publishDiagnostics** (typically ~12ms for same-file, ~1.3s for cross-package):
   - `18:26:25.345` didSave for materialize.go
   - `18:26:25.385` publishDiagnostics: 0 diagnostics (same file, fast)
   - `18:26:26.674` publishDiagnostics: 8 diagnostics for staging_test.go (cross-package, 1.3s later)

2. **Diagnostics are registered for "async delivery"** and consumed by `getLSPDiagnosticAttachments` before the next API request. This is a polling model, not a quiescence-aware model.

3. **No "wait for gopls to settle" logic** exists in the diagnostic consumption path. The handler registers diagnostics immediately as they arrive from gopls.

### Why This Disproportionately Affects Go

Go has a structural property that makes this worse than other languages:

- **Package-level compilation**: Go compiles entire packages, not individual files. Adding a new package (like `internal/registry`) requires gopls to discover, parse, and type-check the entire package before other files can import it.
- **Strict import checking**: Unlike TypeScript (which can partially type-check with missing imports), Go treats an unresolvable import as a hard error. Every file importing the new package gets an error until gopls fully processes the new package.
- **No forward declarations**: Go has no header files or forward declarations. The type information is only available after the full package is analyzed.

## Recommendation

### Short-term Workaround (User-side)

1. **Ignore subagent-boundary diagnostics**: When diagnostics appear immediately after subagent completion, assume they may be stale. Verify with `go build` before acting on them.
2. **Single-file sprints for new packages**: When creating new packages, create the package first in one sprint, then add imports in a subsequent sprint. This gives gopls time to index the new package.

### Medium-term Fix (Claude Code feature request)

1. **Quiescence wait**: After a batch of file edits (especially after subagent completion), wait for gopls to "settle" before consuming diagnostics. gopls supports the `$/progress` notification which signals when analysis is complete. A 2-3 second debounce after the last `publishDiagnostics` would eliminate most false positives.
2. **Diagnostic deduplication with supersession**: When gopls publishes new diagnostics for a file, supersede (not accumulate) any pending diagnostics for that same file. The debug logs show diagnostics being "registered" but the supersession logic may not be working correctly across the subagent boundary.
3. **Workspace-level didChange batching**: Instead of sending `didSave` after every Write tool, batch edits within a subagent's execution and send them together when the subagent completes.

### Long-term Fix (Architecture)

The fundamental issue is that the LSP diagnostic model assumes an interactive editor where a human is typing slowly, not an AI agent making rapid multi-file changes. The diagnostic pipeline needs a mode for "batch edit" scenarios:

1. **Suppress diagnostics during active editing**: Don't register diagnostics while a subagent is actively editing files. Only start collecting after the subagent completes and gopls has quiesced.
2. **Cross-validate with go build**: For Go specifically, consider running `go build ./...` after subagent completion and using its output as the authoritative diagnostic source, ignoring gopls diagnostics that don't match.

## Follow-up Actions

- [ ] File a Claude Code feature request for gopls quiescence wait (CC GitHub or Anthropic feedback)
- [ ] Test whether disabling the gopls-lsp plugin reduces false noise without losing value (diagnostic value vs. noise ratio)
- [ ] Monitor whether gopls v0.21+ improves incremental analysis speed for new package creation
- [ ] Consider adding a PostToolUse hook that runs `go build` after Go file edits and injects real diagnostics, as an alternative to relying on gopls
