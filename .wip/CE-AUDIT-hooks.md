# CE Audit: Hooks + Event System

**Date**: 2026-02-09
**Auditor**: Context Engineer (Claude Opus 4.6)
**Scope**: `internal/hook/`, `internal/hook/clewcontract/`, `internal/cmd/hook/`, `.claude/settings.local.json`, `hooks/hooks.yaml`

---

## Summary

- **CC hook events handled**: 8 of 14
- **CC hook events NOT handled**: SessionEnd, PostToolUseFailure, PermissionRequest, Notification, TeammateIdle, TaskCompleted
- **Event types in clewcontract**: 20
- **Async hooks**: 4 (clew, subagent-start, subagent-stop, route)
- **Sync hooks**: 4 (context, autopark, writeguard, validate, precompact, budget)
- **Total handler implementations**: 10 (`internal/cmd/hook/`)

---

## Hook Configuration Map

Source of truth: `.claude/settings.local.json` (materialized from `hooks/hooks.yaml`)

| CC Event | Handler | Async? | Matcher | Timeout | What It Does |
|----------|---------|--------|---------|---------|--------------|
| SessionStart | `ari hook context` | No (sync) | none | 10s | Injects session context (session ID, status, initiative, rite, phase, mode, git branch, available rites/agents). Rehydrates COMPACT_STATE.md. Emits session.started event. |
| Stop | `ari hook autopark` | No (sync) | none | 5s | Transitions ACTIVE session to PARKED. Saves parked_at timestamp. Logs git status. |
| PreToolUse | `ari hook writeguard` | No (sync) | Edit\|Write | 3s | Blocks writes to *_CONTEXT.md unless Moirai lock is held. Returns deny with Moirai delegation instructions. |
| PreToolUse | `ari hook validate` | No (sync) | Bash | 5s | Blocks rm -rf on .claude/.git/.github/node_modules, force push to main/master, --no-verify, reset --hard, clean -fd. |
| PostToolUse | `ari hook clew` | **Yes (async)** | Edit\|Write\|Bash | 5s | Records tool events to events.jsonl. Emits supplemental file_change/artifact_created events. Extracts orchestrator throughline stamps from Task results. Runs trigger detection (sacred path, file count, failure repeat). |
| PostToolUse | `ari hook budget` | No (sync) | none (all tools) | 3s | Per-session tool-use counter. Warns at 250 calls (configurable). One-shot warn/park alerts via marker files. |
| PreCompact | `ari hook precompact` | No (sync) | none | 5s | Rotates SESSION_CONTEXT.md (archive old content, keep recent 80 lines). Writes COMPACT_STATE.md checkpoint for SessionStart rehydration. |
| SubagentStart | `ari hook subagent-start` | **Yes (async)** | none | 5s | Logs agent.task_start event to clew with agent name/type/task_id. |
| SubagentStop | `ari hook subagent-stop` | **Yes (async)** | none | 5s | Logs agent.task_end event to clew with agent name/type/task_id. |
| UserPromptSubmit | `ari hook route` | **Yes (async)** | ^/ | 5s | Detects slash commands (/start, /park, /resume, /wrap, /consult, /task, /sprint, /commit, /pr, /stamp). Returns routing metadata but does NOT execute. |

### CC Events NOT Handled

| CC Event | Assessment | Priority |
|----------|-----------|----------|
| **SessionEnd** | CRITICAL gap. SessionEnd fires when CC session terminates. Currently no hook handles it. The autopark hook runs on Stop, which is different -- Stop fires when Claude finishes a turn, SessionEnd fires when the conversation window closes. Session cleanup, final event emission, and context archival should happen here. | CRITICAL |
| **PostToolUseFailure** | MEDIUM gap. Tool failures are not tracked in the clew. When Bash commands fail, Edit fails to find old_string, etc., there is no event recorded. This would be valuable for failure pattern detection and debugging. | MEDIUM |
| **PermissionRequest** | LOW. CC fires this when permission is requested from the user. Could be used to track permission patterns, but low value for current use cases. | LOW |
| **Notification** | LOW. Fires when notifications are sent. Not relevant to session tracking. | LOW |
| **TeammateIdle** | LOW. Relevant for multi-instance collaboration. Currently Knossos does not use teammate mode. | LOW |
| **TaskCompleted** | MEDIUM. Fires when a Task tool completes. Currently the clew hook only captures orchestrator throughline from PostToolUse on Task. A dedicated TaskCompleted handler could capture completion status, duration, and output summary more reliably. | MEDIUM |

---

## Event Type Inventory

20 event types defined in `internal/hook/clewcontract/event.go`:

| Event Type | Namespace | Emitted By | Consumed By | Assessment |
|-----------|-----------|-----------|-------------|------------|
| `tool.call` | tool | clew hook (PostToolUse) | audit cmd, tribute extractor, sails contract | ACTIVE, USEFUL |
| `tool.file_change` | tool | clew hook (supplemental) | tribute extractor | ACTIVE, USEFUL |
| `tool.artifact_created` | tool | clew hook (Write to PRD/TDD/ADR) | tribute extractor | ACTIVE, USEFUL |
| `tool.error` | tool | clew hook (error paths) | audit cmd | ACTIVE, USEFUL |
| `agent.decision` | agent | /stamp command, orchestrator throughline | audit cmd, tribute extractor | ACTIVE, USEFUL |
| `agent.task_start` | agent | subagent-start hook, CLI NewTaskStartEvent() | audit cmd | ACTIVE, USEFUL |
| `agent.task_end` | agent | subagent-stop hook, CLI NewTaskEndEvent() | audit cmd | ACTIVE, USEFUL |
| `agent.handoff_prepared` | agent | `ari handoff prepare` | handoff history, validation | ACTIVE, USEFUL |
| `agent.handoff_executed` | agent | `ari handoff execute` | handoff history, validation | ACTIVE, USEFUL |
| `session.started` | session | context hook (SessionStart) | audit cmd | ACTIVE, USEFUL |
| `session.ended` | session | `ari session park`, `ari session wrap` | audit cmd | ACTIVE, USEFUL |
| `session.created` | session | `ari session create` | audit cmd | ACTIVE, USEFUL |
| `session.parked` | session | `ari session park` | audit cmd | ACTIVE, USEFUL |
| `session.resumed` | session | `ari session resume` | audit cmd | ACTIVE, USEFUL |
| `session.archived` | session | `ari session wrap` | audit cmd | ACTIVE, USEFUL |
| `session.frayed` | session | `ari session fray` | audit cmd | ACTIVE, LOW-TRAFFIC |
| `session.strand_resolved` | session | `ari session wrap` (strand) | audit cmd | ACTIVE, LOW-TRAFFIC |
| `phase.transitioned` | phase | `ari session transition` | audit cmd | ACTIVE, USEFUL |
| `quality.sails_generated` | quality | `ari session wrap` (sails check) | sails gate, audit cmd | ACTIVE, USEFUL |
| `context_switch` | (deferred) | Constructor exists, NEVER emitted | triggers.go (checked but never fires) | DEAD CODE |

---

## Critical Findings

### C-1: SessionEnd Hook Missing -- Sessions Never Formally Close

**Severity**: CRITICAL
**Location**: `.claude/settings.local.json` (missing entry), `internal/cmd/hook/` (no handler)

The SessionEnd CC event is defined in `env.go` (line 38: `EventSessionEnd HookEvent = "SessionEnd"`) but has NO handler registered and NO implementation. The autopark hook handles Stop, but Stop and SessionEnd are different CC lifecycle events:

- **Stop**: Fires when Claude finishes a response turn. Session may continue.
- **SessionEnd**: Fires when the CC conversation window is closing/terminated. Session is ending.

Currently, if a user closes the CC window without explicitly running `/park` or `/wrap`, the session remains in ACTIVE status indefinitely. The autopark on Stop partially mitigates this, but Stop does not always fire before SessionEnd.

**Impact**: Sessions can remain ACTIVE forever. No final `session.ended` event is emitted on natural conversation termination. Cognitive budget counters are never cleaned up. The clew has no terminal event for sessions that end without explicit park/wrap.

### C-2: context_switch Event Type Is Dead Code

**Severity**: CRITICAL (as a context engineering concern -- wasted token budget)
**Location**: `internal/hook/clewcontract/event.go:18`, `triggers.go:185`

`EventTypeContextSwitch` is defined, has a constructor (`NewContextSwitchEvent`), has trigger detection code (`checkContextSwitch`), and is tested -- but **nothing ever emits it**. Zero callers of `NewContextSwitchEvent` outside the test file.

This was documented as "deferred" in the ADR-0027 work, but the code remains in the codebase consuming ~60 lines across event.go, triggers.go, and test files. The trigger code checks for this event type on every PostToolUse invocation (wasted CPU in hot path).

---

## High Findings

### H-1: Autopark Has No Timeout on git status Subprocess

**Severity**: HIGH
**Location**: `internal/cmd/hook/autopark.go:152-162`

```go
func getGitStatusQuick() string {
    cmd := exec.Command("git", "status", "--short")
    out, err := cmd.Output()
    // ...
}
```

Unlike `getGitBranch` and `getBaseBranch` in `context.go` which use `context.WithTimeout(context.Background(), gitCommandTimeout)` (50ms), `getGitStatusQuick` in autopark.go runs `git status --short` with NO timeout. On large repos or NFS-mounted directories, `git status` can hang for seconds or minutes. The parent `withTimeout` (100ms default) protects against infinite hangs, but the hook will hit the 500ms MaxTimeout ceiling rather than the 50ms git timeout used elsewhere.

### H-2: BufferedEventWriter Used in Short-Lived Hook Processes

**Severity**: HIGH
**Location**: `internal/hook/clewcontract/record.go:20`, `internal/cmd/hook/clew.go:208`, `internal/cmd/hook/context.go:287`, `internal/cmd/hook/subagent.go:118`

Multiple locations create a `BufferedEventWriter` (5-second flush interval, background goroutine) for a hook process that lives <100ms:

```go
// record.go
writer := NewBufferedEventWriter(sessionDir, DefaultFlushInterval)
defer writer.Close()
writer.Write(event)
if err := writer.Flush(); err != nil { ... }
```

The pattern is: create buffered writer, write one event, immediately flush, close. The buffered writer starts a background goroutine with a 5-second ticker that never fires because the process exits in <100ms. This creates unnecessary overhead:
- Goroutine creation and channel allocation
- Ticker creation (5s interval, never fires)
- `close(w.done)` + `<-w.flushed` synchronization on Close()

For short-lived hook processes, `EventWriter.Write()` (synchronous, no goroutine) would be more appropriate. The `BufferedEventWriter` was designed for long-lived processes, not hook handlers.

### H-3: Route Hook Output Not Consumed by CC

**Severity**: HIGH
**Location**: `internal/cmd/hook/route.go`, `.claude/settings.local.json:108-116`

The route hook runs on UserPromptSubmit and outputs JSON with routing information:
```json
{"routed": true, "command": "/park", "args": "my-session", "category": "session"}
```

But this is configured as `async: true`. Async hooks fire-and-forget -- CC does NOT read their stdout. The routing information is computed and immediately discarded. This hook has two problems:
1. **Async = output discarded**: If routing info should influence CC behavior, it must be sync.
2. **No side effect**: Unlike the async clew hook (writes to events.jsonl), the route hook performs no side effects. It just computes and returns JSON that nobody reads.

The route hook is currently a no-op in production. It consumes CPU on every slash command but produces no observable effect.

### H-4: PreToolUse validate Does Not Guard PROVENANCE_MANIFEST.yaml or .claude/settings.local.json

**Severity**: HIGH
**Location**: `internal/cmd/hook/writeguard.go:24-27`

The writeguard only protects two patterns:
```go
var protectedPatterns = []string{
    "SESSION_CONTEXT.md",
    "SPRINT_CONTEXT.md",
}
```

Critical platform files that should NOT be directly written by Claude:
- `PROVENANCE_MANIFEST.yaml` -- ownership/checksum tracking, corruption breaks sync
- `.claude/settings.local.json` -- hook configuration, modification could disable guards
- `.claude/CLAUDE.md` -- platform-owned sections would be overwritten on next sync
- `KNOSSOS_MANIFEST.yaml` -- rite configuration

The validate hook guards Bash commands (rm -rf, force push), but the writeguard does not protect platform infrastructure files from Edit/Write tools.

---

## Medium Findings

### M-1: PostToolUseFailure Not Tracked -- Failure Pattern Blindspot

**Severity**: MEDIUM
**Location**: Missing handler

When tools fail (Bash exit code != 0, Edit fails to match old_string, etc.), CC fires `PostToolUseFailure`. This is not handled. The clew only tracks successful PostToolUse events. Failed operations are invisible in the event stream.

The trigger system in `triggers.go` has `checkFailureRepeat` which looks for `exit_code` in event metadata, but this metadata is only populated if the hook fires -- which it does not for PostToolUseFailure.

### M-2: Budget Hook Writes to stderr in Production

**Severity**: MEDIUM
**Location**: `internal/cmd/hook/budget.go:118,132`

```go
fmt.Fprintf(os.Stderr, "[cognitive-budget] Warning: %s\n", out.Message)
fmt.Fprintf(os.Stderr, "[cognitive-budget] Alert: %s\n", out.Message)
```

Hook stderr output is passed through to Claude's context by CC. This means budget warnings inject unstructured text into the conversation. While the intent is good (warn about session length), the mechanism is ad-hoc. The budget hook returns structured JSON on stdout AND writes to stderr -- dual-channel output with no coordination.

### M-3: Clew Hook Creates Duplicate BufferedEventWriter for Supplemental Events

**Severity**: MEDIUM
**Location**: `internal/cmd/hook/clew.go:134-144`

The clew hook creates up to 3 separate `BufferedEventWriter` instances in a single invocation:
1. `RecordToolEvent` in record.go creates one (line 20)
2. `emitSupplementalEvents` creates another (line 208)
3. `emitErrorEvent` may create a third (line 273)

Each creates a goroutine, ticker, and opens the file separately. These could be unified into a single writer passed through the call chain.

### M-4: Subagent Hooks Assume tool_input Contains Agent Info

**Severity**: MEDIUM
**Location**: `internal/cmd/hook/subagent.go:102-103`

```go
agentInfo := parseSubagentInfo(hookEnv.ToolInput)
```

The SubagentStart/SubagentStop hooks parse `hookEnv.ToolInput` for agent name/type/task_id. But CC's stdin payload for SubagentStart may not include these fields in `tool_input` -- the agent info might be in different fields. The fallback to "unknown" means all subagent events could be logged with `agent_name: "unknown"` if CC's payload format differs from what is expected.

### M-5: validate Bypass via Environment Variable

**Severity**: MEDIUM
**Location**: `internal/cmd/hook/validate.go:96-98`

```go
if os.Getenv(ValidateBypassEnvVar) == "1" {
    return outputValidateAllow(printer)
}
```

Setting `ARI_VALIDATE_BYPASS=1` disables ALL validation guards. While this is documented and intentional for testing, if an agent sets this env var via Bash tool before executing a destructive command, the validate hook will not fire. The Bash tool could `export ARI_VALIDATE_BYPASS=1 && rm -rf .claude/` in a single command. The regex patterns would still miss compound commands that set the var inline.

---

## Low Findings

### L-1: Env Var Fallback Path Is Tested but Deprecated

**Severity**: LOW
**Location**: `internal/hook/env.go:113-121`

ParseEnv still reads from environment variables as a fallback when stdin is empty. The env var names (CLAUDE_HOOK_EVENT, CLAUDE_TOOL_NAME, etc.) are marked deprecated in comments, but the fallback code remains and is actively tested. This is defensive (graceful degradation) and correct, but the env var constants and fallback code add ~30 lines of dead-in-production code.

### L-2: Sacred Path Detection Has False Positive on "docs/decisions/" Substring

**Severity**: LOW
**Location**: `internal/hook/clewcontract/triggers.go:111`

```go
return strings.Contains(path, pattern) || strings.Contains(path, strings.TrimSuffix(pattern, "/"))
```

The pattern `docs/decisions/` will match any path containing this substring, including paths like `old-docs/decisions-archive/foo.txt` or `mydocs/decisions/draft.md`. For a trigger system (advisory, not blocking), this is acceptable but imprecise.

### L-3: Duplicate Event Type Constants Between env.go and clewcontract

**Severity**: LOW
**Location**: `internal/hook/env.go:31-46` vs `internal/hook/clewcontract/event.go:14-35`

Two separate type systems exist:
- `hook.HookEvent` -- CC lifecycle events (PreToolUse, PostToolUse, SessionStart, etc.)
- `clewcontract.EventType` -- Clew semantic events (tool.call, session.started, etc.)

These are architecturally distinct (CC events vs Knossos domain events), so this is correct by design. But the naming overlap (SessionStart in both) could cause confusion. Not a bug.

---

## Async/Sync Configuration Analysis

### Current Configuration (from settings.local.json)

| Hook | Config | Correct? | Analysis |
|------|--------|----------|----------|
| SessionStart/context | sync | YES | Must inject context into CC's next turn |
| Stop/autopark | sync | YES | Must complete state transition before process exits |
| PreToolUse/writeguard | sync | YES | Must return deny/allow before tool executes |
| PreToolUse/validate | sync | YES | Must return deny/allow before tool executes |
| PostToolUse/clew | async | YES | Side-effect only (writes to events.jsonl), should not delay tool use |
| PostToolUse/budget | sync | QUESTIONABLE | Budget warnings work via stderr. As sync, this adds latency to every tool call. Could be async if stderr is not needed for the current turn. |
| PreCompact | sync | YES | Must complete rotation before compaction |
| SubagentStart | async | YES | Side-effect only (clew logging) |
| SubagentStop | async | YES | Side-effect only (clew logging) |
| UserPromptSubmit/route | async | WRONG | Async means output is discarded. Either make sync to use output, or remove the hook entirely since it has no side effects. |

---

## Data Flow Diagram

```
CC Lifecycle Event
       |
       v
  [stdin JSON payload]
       |
       v
  ari hook <subcommand>
       |
       +---> ParseEnv() reads stdin JSON + env var fallback
       |         |
       |         v
       |    Env struct (event, tool, session, project)
       |         |
       |    +----+----+
       |    |         |
       |    v         v
       | PreToolUse  PostToolUse/SessionStart/Stop/etc.
       |    |         |
       |    v         v
       | writeguard   resolveSession()
       | validate        |
       |    |             v
       |    v         Session directory
       | stdout JSON     |
       | (deny/allow)    +---> events.jsonl (clewcontract)
       |                 |         |
       |                 |         v
       |                 |     audit cmd / tribute / sails
       |                 |
       |                 +---> SESSION_CONTEXT.md (context hook)
       |                 +---> COMPACT_STATE.md (precompact)
       |                 +---> temp counter file (budget)
       |
       v
  CC reads stdout JSON
  (sync hooks only)
```

---

## Recommendations

### Priority 1: Add SessionEnd Handler (CRITICAL)

Create `internal/cmd/hook/sessionend.go` that:
1. Emits `session.ended` event to clew
2. Optionally auto-parks if session is still ACTIVE (belt-and-suspenders with Stop/autopark)
3. Cleans up temp budget counter file
4. Register in `hooks/hooks.yaml` as sync with 5s timeout

This is the single highest-impact gap. Without it, sessions that end by the user closing CC never formally close.

### Priority 2: Fix Route Hook (HIGH)

Either:
- **Option A**: Make route hook sync so CC can read routing info. This would allow the route hook to inject additionalContext that helps CC execute the slash command.
- **Option B**: Remove route hook entirely. Slash commands already work via CC's native command system (`.claude/commands/`). The route hook adds no value if its output is discarded.

Recommendation: Option B (remove). The route hook was likely designed before CC's native command system matured. It is now redundant.

### Priority 3: Protect Platform Files in Writeguard (HIGH)

Add to `protectedPatterns`:
```go
var protectedPatterns = []string{
    "SESSION_CONTEXT.md",
    "SPRINT_CONTEXT.md",
    "PROVENANCE_MANIFEST.yaml",
    "settings.local.json",
    "KNOSSOS_MANIFEST.yaml",
}
```

These are platform-owned files that should only be modified by the ari binary, not by Claude directly.

### Priority 4: Add PostToolUseFailure Handler (MEDIUM)

Register a handler for PostToolUseFailure that:
1. Records `tool.error` events with failure details (exit code, error message)
2. Enables the existing `checkFailureRepeat` trigger to actually function
3. Should be async (same as PostToolUse/clew -- side-effect only)

### Priority 5: Use Synchronous EventWriter in Hook Processes (MEDIUM)

Replace `BufferedEventWriter` with `EventWriter` in all hook handlers. Hook processes live <100ms -- the buffered writer's 5-second flush interval and background goroutine are pure overhead. The code already does immediate `Flush()` after every `Write()`, making the buffering layer pointless.

### Priority 6: Remove context_switch Dead Code (LOW)

Remove `EventTypeContextSwitch`, `NewContextSwitchEvent`, and `checkContextSwitch` from clewcontract. This was deferred in ADR-0027 Sprint 1 and has remained dead code since. If context_switch tracking is needed in the future, it can be re-added with an actual emission path.

### Priority 7: Add git status Timeout in Autopark (LOW)

Apply the same `context.WithTimeout(context.Background(), gitCommandTimeout)` pattern used in `context.go` to `getGitStatusQuick()` in `autopark.go`.

---

## Token Budget Impact

The hook infrastructure is well-designed from a context engineering perspective:

**What enters CC's context window**:
- SessionStart/context: session metadata table (~200 tokens). Appropriate.
- writeguard deny: Moirai delegation instructions (~80 tokens). Appropriate.
- validate deny: safety explanation (~30 tokens). Appropriate.
- budget stderr: unstructured warning text (~40 tokens). Should be structured or removed.
- route output: **never enters context** (async). Wasted computation.

**What stays outside CC's context window**:
- events.jsonl: all clew events. Correct -- consumed by audit cmd, tribute, sails.
- temp counter files: budget state. Correct.
- COMPACT_STATE.md: consumed once on next SessionStart. Correct.

The system correctly uses hooks for ephemeral context injection rather than hardcoding session state into CLAUDE.md. This is the primary strength of the architecture.

---

## Files Audited

| File | Lines | Role |
|------|-------|------|
| `/Users/tomtenuta/Code/knossos/internal/hook/env.go` | 232 | Stdin JSON parsing, env var fallback, HookEvent types |
| `/Users/tomtenuta/Code/knossos/internal/hook/input.go` | 150 | Tool input JSON parsing |
| `/Users/tomtenuta/Code/knossos/internal/hook/output.go` | 22 | CC-native PreToolUse output format |
| `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/event.go` | 601 | 20 event types + constructors |
| `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/writer.go` | 283 | EventWriter + BufferedEventWriter |
| `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/record.go` | 149 | RecordToolEvent, BuildEventFromToolInput |
| `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/triggers.go` | 337 | Auto-trigger detection (sacred path, file count, failure, context switch) |
| `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/orchestrator.go` | 87 | Throughline extraction from orchestrator responses |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/hook.go` | 162 | Hook command group, shared context, timeout wrapper |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/context.go` | 314 | SessionStart handler |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go` | 190 | PreToolUse Edit/Write guard |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/validate.go` | 213 | PreToolUse Bash guard |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/clew.go` | 284 | PostToolUse event recording |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/autopark.go` | 163 | Stop auto-park |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/budget.go` | 201 | PostToolUse cognitive budget counter |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/precompact.go` | 182 | PreCompact SESSION_CONTEXT rotation |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/route.go` | 173 | UserPromptSubmit slash command detection |
| `/Users/tomtenuta/Code/knossos/internal/cmd/hook/subagent.go` | 216 | SubagentStart/SubagentStop logging |
| `/Users/tomtenuta/Code/knossos/.claude/settings.local.json` | 140 | Hook registration (materialized) |
| `/Users/tomtenuta/Code/knossos/hooks/hooks.yaml` | 104 | Hook registration (source of truth) |
| `/Users/tomtenuta/Code/knossos/internal/session/events_read.go` | 112 | Event reader (audit cmd consumer) |
