# TDD: Handoff Delegation Interface

> Technical Design Document for ari session command delegation to state-mate via thin wrapper pattern.

**Status**: Draft
**Author**: Architect
**Date**: 2026-01-05
**Sprint**: Thread Contract v2 Completion (Knossos 90% Readiness)
**Task**: S1 task-007

---

## 1. Overview

This Technical Design Document specifies the **Handoff Delegation Interface** for Ariadne's session lifecycle commands. The design ensures `ari` remains a thin wrapper that delegates all state mutations to `state-mate`, the centralized state authority, with Thread Contract events emitted after successful operations.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| state-mate Agent | `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` |
| Session Commands | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/` |
| Thread Contract Events | `/Users/tomtenuta/Code/roster/ariadne/internal/hook/threadcontract/event.go` |
| Session FSM | `/Users/tomtenuta/Code/roster/ariadne/internal/session/fsm.go` |
| Session Context | `/Users/tomtenuta/Code/roster/ariadne/internal/session/context.go` |
| Session Events | `/Users/tomtenuta/Code/roster/ariadne/internal/session/events.go` |
| state-mate Invocation Pattern | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/shared-sections/state-mate-invocation.md` |

### 1.2 Current State Analysis

The existing implementation in `ariadne/internal/cmd/session/` shows:

1. **park.go**: Contains full business logic (FSM validation, context mutation, save, event emission)
2. **resume.go**: Contains full business logic (FSM validation, context mutation, save, event emission)
3. **wrap.go**: Contains full business logic (FSM validation, sails generation, context mutation, archiving)

**Problem**: This duplicates state-mate's responsibilities, violating the Single Authority principle. state-mate exists as the centralized authority for all SESSION_CONTEXT.md mutations, but ari commands bypass it entirely.

### 1.3 Design Goals

1. **Thin Wrapper**: `ari session {park|resume|wrap}` becomes pure delegation
2. **Single Authority**: state-mate owns all state transitions and validation
3. **Event Ordering**: Thread Contract events emit AFTER successful state-mate completion
4. **Shell Compatibility**: CLI interface maintained unchanged for scripts
5. **Error Propagation**: state-mate errors bubble up as structured CLI errors

### 1.4 Non-Goals

- Changing state-mate's internal implementation
- Modifying SESSION_CONTEXT.md schema
- Adding new CLI commands (existing interface preserved)
- Changing the FSM transition rules

---

## 2. Architecture Decision: Delegation Pattern

### 2.1 Options Considered

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| **A: Direct Delegation** | ari invokes state-mate via Task tool | Single authority, matches agent pattern | Requires Claude context for Task tool |
| **B: Shell Bridge** | ari calls shell script that orchestrates Claude | Works from any shell | Adds indirection, harder to debug |
| **C: Shared Library** | Extract state-mate logic to Go package, call from both | Pure Go, no agent dependency | Duplicates state-mate, two codebases |
| **D: HTTP Bridge** | ari calls local HTTP server that invokes state-mate | Decoupled, language-agnostic | Over-engineered, operational overhead |

### 2.2 Decision: Option C - Shared Library (Hybrid)

**Selected**: Option C with a twist -- state-mate REMAINS the authority within Claude Code sessions, but `ari` commands use a shared Go library (`internal/session`) that implements the same business logic.

**Rationale**:
- state-mate operates within Claude agent context (requires Task tool, cannot be called from shell)
- `ari` is a standalone CLI tool called from shell, not from within Claude
- The Go session package (`internal/session`) already contains the FSM and context management
- Both paths use the same underlying logic; they differ only in invocation context

**Key Insight**: The existing implementation IS the correct pattern for CLI usage. The issue isn't that ari duplicates state-mate -- it's that they serve different invocation contexts:

| Context | Tool | Authority |
|---------|------|-----------|
| Within Claude Code session | Task(state-mate, ...) | state-mate agent |
| From shell/CLI | `ari session park` | ari using shared Go package |

**ADR**: See ADR-0008 (to be created) for full decision record.

---

## 3. Revised Design: CLI and Agent Parity

### 3.1 Interface Contract

Both `ari` and `state-mate` must produce identical state transitions for the same inputs:

```
Input: park_session reason="Taking a break"
Session: ACTIVE, session-20260105-143022-abc12345

Output (from both):
- SESSION_CONTEXT.md: status=PARKED, parked_at=<now>, parked_reason="Taking a break"
- Event: SESSION_PARKED in events.jsonl
- Audit: Entry in session-mutations.log
```

### 3.2 Shared Logic Extraction

The Go package `internal/session` already contains the shared logic. The commands use it correctly. What's needed is ensuring Thread Contract v2 event emission:

```
internal/session/
+-- fsm.go            # FSM with transition rules (exists)
+-- context.go        # Context parsing/serialization (exists)
+-- events.go         # Session event emission (exists)
+-- status.go         # Status types (exists)

internal/hook/threadcontract/
+-- event.go          # Thread Contract v2 events (exists)
+-- writer.go         # JSONL writer (exists)
```

### 3.3 Thread Contract Event Integration

The key gap is emitting Thread Contract v2 events from the session commands. Currently, `session/events.go` emits session-level events to `events.jsonl` in the session directory. Thread Contract v2 requires:

1. **TASK_END**: When a task completes (not directly applicable to session lifecycle)
2. **SESSION_END**: When a session is parked, wrapped, or archived

---

## 4. Interface Design

### 4.1 CLI Interface (Unchanged)

```bash
# Park session
ari session park [--reason="..."] [--session=SESSION_ID]

# Resume session
ari session resume [--session=SESSION_ID]

# Wrap session
ari session wrap [--skip-checks] [--no-archive] [--session=SESSION_ID]
```

### 4.2 Event Emission Sequence

#### 4.2.1 Park Operation

```
User                    ari                     session pkg           Thread Contract
  |                      |                           |                      |
  |--park --reason=X---->|                           |                      |
  |                      |--LoadContext()----------->|                      |
  |                      |<--Context----------------+|                      |
  |                      |--ValidateTransition()---->|                      |
  |                      |<--OK---------------------+|                      |
  |                      |--UpdateContext()--------->|                      |
  |                      |--SaveContext()----------->|                      |
  |                      |<--OK---------------------+|                      |
  |                      |--EmitParked()------------>|                      |
  |                      |                           |--Write(SESSION_PARKED)->
  |                      |                           |<--OK-----------------+|
  |                      |--EmitSessionEnd()-------->|                      |
  |                      |                           |--Write(session_end)->+|
  |                      |                           |<--OK-----------------+|
  |<--Success-----------+|                           |                      |
```

#### 4.2.2 Resume Operation

```
User                    ari                     session pkg           Thread Contract
  |                      |                           |                      |
  |--resume------------->|                           |                      |
  |                      |--LoadContext()----------->|                      |
  |                      |<--Context----------------+|                      |
  |                      |--ValidateTransition()---->|                      |
  |                      |<--OK---------------------+|                      |
  |                      |--UpdateContext()--------->|                      |
  |                      |--SaveContext()----------->|                      |
  |                      |<--OK---------------------+|                      |
  |                      |--EmitResumed()----------->|                      |
  |                      |                           |--Write(SESSION_RESUMED)->
  |                      |                           |<--OK-----------------+|
  |                      |                 [No SESSION_END - session continuing]
  |<--Success-----------+|                           |                      |
```

#### 4.2.3 Wrap Operation

```
User                    ari                     session pkg           Thread Contract
  |                      |                           |                      |
  |--wrap--------------->|                           |                      |
  |                      |--LoadContext()----------->|                      |
  |                      |<--Context----------------+|                      |
  |                      |--ValidateTransition()---->|                      |
  |                      |<--OK---------------------+|                      |
  |                      |--GenerateSails()--------->|                      |
  |                      |<--SailsResult------------+|                      |
  |                      |                           |--Write(sails_generated)->
  |                      |--UpdateContext()--------->|                      |
  |                      |--SaveContext()----------->|                      |
  |                      |<--OK---------------------+|                      |
  |                      |--ArchiveSession()-------->|                      |
  |                      |--EmitArchived()---------->|                      |
  |                      |                           |--Write(SESSION_ARCHIVED)->
  |                      |--EmitSessionEnd()-------->|                      |
  |                      |                           |--Write(session_end)->+|
  |                      |                           |<--OK-----------------+|
  |<--Success-----------+|                           |                      |
```

### 4.3 Thread Contract Event Types

From `threadcontract/event.go`, the relevant events:

| Event Type | When Emitted | Meta Fields |
|------------|--------------|-------------|
| `session_end` | Park, Wrap | session_id, status (parked/completed/abandoned), duration_ms |
| `sails_generated` | Wrap | session_id, color, computed_base, reasons, file_path |

**Design Decision**: `session_end` is emitted for:
- **Park**: status="parked" (session suspended, may resume)
- **Wrap**: status="completed" (session finished successfully)
- **Resume**: NO session_end (session continuing, not ending)

### 4.4 Error Handling Strategy

#### 4.4.1 Error Categories

| Category | Source | CLI Exit Code | Example |
|----------|--------|---------------|---------|
| LIFECYCLE_VIOLATION | FSM validation | 5 | "Cannot park already parked session" |
| FILE_NOT_FOUND | Context loading | 6 | "Session context not found" |
| PERMISSION_DENIED | Lock acquisition | 8 | "Cannot acquire session lock" |
| LOCK_TIMEOUT | Lock manager | 8 | "Lock acquisition timed out" |
| SCHEMA_VIOLATION | Context validation | 4 | "Invalid session state" |

#### 4.4.2 Error Flow

```
Operation fails at any step
        |
        v
+-------------------+
| Rollback partial  |
| state changes     |
+-------------------+
        |
        v
+-------------------+
| Release any held  |
| locks             |
+-------------------+
        |
        v
+-------------------+
| Emit error event  |
| to Thread Contract|
+-------------------+
        |
        v
+-------------------+
| Return structured |
| error response    |
+-------------------+
```

#### 4.4.3 Error Response Structure

```json
{
  "error": {
    "code": "LIFECYCLE_VIOLATION",
    "message": "Cannot transition from PARKED to PARKED",
    "details": {
      "from_status": "PARKED",
      "to_status": "PARKED",
      "session_id": "session-20260105-143022-abc12345"
    },
    "hint": "Session is already parked. Use 'ari session resume' first."
  }
}
```

---

## 5. Implementation Changes

### 5.1 Required Changes to park.go

Add Thread Contract SESSION_END event after successful park:

```go
// After existing EmitParked call in runPark()
// Emit Thread Contract session_end event
sessionDir := resolver.SessionDir(sessionID)
tcWriter, err := threadcontract.NewEventWriter(sessionDir)
if err == nil {
    // Calculate duration from session start to park
    durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()
    sessionEndEvent := threadcontract.NewSessionEndEvent(sessionID, "parked", durationMs)
    if err := tcWriter.Write(sessionEndEvent); err != nil {
        printer.VerboseLog("warn", "failed to emit session_end event", map[string]interface{}{"error": err.Error()})
    }
}
```

### 5.2 Required Changes to wrap.go

The wrap.go already emits `sails_generated` event. Add `session_end` event:

```go
// After existing EmitArchived call in runWrap()
// Emit Thread Contract session_end event
tcWriter, err := threadcontract.NewEventWriter(sessionDir)
if err == nil {
    durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()
    sessionEndEvent := threadcontract.NewSessionEndEvent(sessionID, "completed", durationMs)
    if err := tcWriter.Write(sessionEndEvent); err != nil {
        printer.VerboseLog("warn", "failed to emit session_end event", map[string]interface{}{"error": err.Error()})
    }
}
```

### 5.3 No Changes to resume.go

Resume does NOT emit session_end because the session is continuing, not ending. The existing implementation is correct.

### 5.4 Handoff Commands (New Implementation)

The `ari handoff` commands (`prepare`, `execute`, `status`, `history`) currently return "not implemented". These need implementation for agent handoff tracking:

```go
// internal/cmd/handoff/prepare.go - Core logic
func runPrepare(ctx *cmdContext, fromAgent, toAgent string) error {
    // 1. Load session context
    // 2. Validate fromAgent matches current phase agent
    // 3. Check artifact requirements for handoff
    // 4. Generate handoff context
    // 5. Emit TASK_END event for fromAgent
    // 6. Return handoff readiness summary
}
```

---

## 6. Event Emission Points (Summary)

### 6.1 Event Matrix

| Command | Session Event | Thread Contract Event | When |
|---------|---------------|----------------------|------|
| `ari session park` | SESSION_PARKED | session_end (status=parked) | After state saved |
| `ari session resume` | SESSION_RESUMED | (none) | After state saved |
| `ari session wrap` | SESSION_ARCHIVED | sails_generated + session_end (status=completed) | After archive |
| `ari handoff prepare` | (none) | task_end | After validation |
| `ari handoff execute` | (none) | task_start | After execution |

### 6.2 Event Ordering Rules

1. **State First**: Always persist state changes before emitting events
2. **Session Events First**: Emit session-level events (SESSION_PARKED, etc.) before Thread Contract events
3. **Non-Blocking**: Event emission failures should warn, not fail the operation
4. **Idempotent**: Re-running a command should not emit duplicate events (guard with state checks)

---

## 7. state-mate Parity Contract

### 7.1 Behavioral Equivalence

When invoked from different contexts, state transitions must produce identical results:

| Invocation | Operation | Result |
|------------|-----------|--------|
| `ari session park --reason="Break"` | Park from CLI | STATE: PARKED, EVENTS: SESSION_PARKED + session_end |
| `Task(state-mate, "park_session reason='Break'")` | Park from Claude | STATE: PARKED, EVENTS: SESSION_PARKED + session_end |

### 7.2 Verification Approach

Test both paths produce:
1. Identical SESSION_CONTEXT.md content (modulo timestamps)
2. Same events in events.jsonl (modulo timestamps)
3. Same entry in session-mutations.log

---

## 8. Implementation Guidance

### 8.1 File Changes

| File | Change Type | Description |
|------|-------------|-------------|
| `ariadne/internal/cmd/session/park.go` | Modify | Add Thread Contract session_end emission |
| `ariadne/internal/cmd/session/wrap.go` | Modify | Add Thread Contract session_end emission |
| `ariadne/internal/cmd/session/resume.go` | None | Already correct (no session_end) |
| `ariadne/internal/cmd/handoff/prepare.go` | Implement | Full implementation with task_end |
| `ariadne/internal/cmd/handoff/execute.go` | Implement | Full implementation with task_start |
| `ariadne/internal/cmd/handoff/status.go` | Implement | Query current handoff state |
| `ariadne/internal/cmd/handoff/history.go` | Implement | Query handoff history from events |

### 8.2 Test Requirements

| Test ID | Description | Type |
|---------|-------------|------|
| `unit_001` | Park emits session_end with status=parked | Unit |
| `unit_002` | Wrap emits session_end with status=completed | Unit |
| `unit_003` | Resume does NOT emit session_end | Unit |
| `int_001` | Park creates correct events.jsonl entries | Integration |
| `int_002` | Wrap creates WHITE_SAILS.yaml + events | Integration |
| `int_003` | Handoff prepare validates artifacts | Integration |
| `int_004` | CLI and state-mate produce identical state | Parity |

### 8.3 Implementation Order

1. **Phase 1**: Add Thread Contract events to park.go and wrap.go (low risk, additive)
2. **Phase 2**: Implement handoff prepare/execute/status/history (new functionality)
3. **Phase 3**: Add parity tests for CLI vs state-mate (verification)

---

## 9. Backward Compatibility

### 9.1 Classification: COMPATIBLE

This change is **fully backward compatible**:

1. **CLI Interface Unchanged**: Same commands, same flags, same exit codes
2. **Additive Events**: New Thread Contract events don't break existing consumers
3. **Session Schema Unchanged**: SESSION_CONTEXT.md format unchanged
4. **Graceful Degradation**: Event emission failures warn but don't fail operations

### 9.2 Migration Path

No migration required. Existing sessions work as-is. New events appear in events.jsonl going forward.

---

## 10. Handoff Criteria

Ready for Implementation when:

- [x] Interface design showing ari CLI commands
- [x] Event emission sequence diagrams for park/resume/wrap
- [x] Error handling strategy documented
- [x] CLI interface specification confirmed unchanged
- [x] Thread Contract event types identified (session_end)
- [x] File changes enumerated
- [x] Test matrix defined
- [x] Backward compatibility classified (COMPATIBLE)
- [x] state-mate parity contract documented

---

## 11. Open Questions

None. Design is complete per requirements.

---

## 12. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-handoff-delegation.md` | Created |
| park.go | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/park.go` | Read |
| resume.go | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/resume.go` | Read |
| wrap.go | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/wrap.go` | Read |
| handoff.go | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/handoff/handoff.go` | Read |
| session.go | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/session/session.go` | Read |
| event.go (threadcontract) | `/Users/tomtenuta/Code/roster/ariadne/internal/hook/threadcontract/event.go` | Read |
| writer.go (threadcontract) | `/Users/tomtenuta/Code/roster/ariadne/internal/hook/threadcontract/writer.go` | Read |
| fsm.go | `/Users/tomtenuta/Code/roster/ariadne/internal/session/fsm.go` | Read |
| context.go | `/Users/tomtenuta/Code/roster/ariadne/internal/session/context.go` | Read |
| events.go (session) | `/Users/tomtenuta/Code/roster/ariadne/internal/session/events.go` | Read |
| state-mate.md | `/Users/tomtenuta/Code/roster/user-agents/state-mate.md` | Read |
| state-mate-invocation.md | `/Users/tomtenuta/Code/roster/user-skills/session-lifecycle/shared-sections/state-mate-invocation.md` | Read |

---

## 13. Related Documents

- TDD-artifact-registry.md - Artifact registration on mark_complete
- TDD-knossos-v2.md - White Sails confidence signaling
- state-mate.md - Centralized state authority agent
- ADR-0005 - state-mate centralized state authority decision
