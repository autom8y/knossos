---
name: continue
description: Resume a parked work session with full context
argument-hint:
allowed-tools: Bash, Read, Task
model: sonnet
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Resume a parked work session with full context restoration. $ARGUMENTS

## Session Resolution

The command uses the current session marker at `.claude/sessions/.current-session`:

1. **Read current session**: Reads session ID from `.current-session` file
2. **Fail if no session**: Returns error if no session is currently set
3. **Validate status**: Session must be in PARKED status to resume

No flags are accepted - the command operates on the current session only.

## Pre-flight

1. Verify a current session exists (`.current-session` file)
2. Load session context from `SESSION_CONTEXT.md`
3. Verify session is in PARKED status (FSM validates PARKED -> ACTIVE transition)

## Behavior

1. **Delegate to Moirai** for session state mutation:
   ```
   Task(moirai, "resume_session")
   ```
   Moirai will:
   - Acquire lock to prevent race conditions
   - Validate FSM allows PARKED -> ACTIVE transition
   - Execute `ari session resume`
   - Update session state (status, resumed_at, clear parked_at/reason)
   - Validate the result and log to audit trail
   - Return structured response

2. **Display resumption summary** based on Moirai's response:
   ```json
   {
     "session_id": "session-20251224-143052-a1b2c3d4",
     "status": "ACTIVE",
     "previous_status": "PARKED",
     "resumed_at": "2025-01-07T10:30:00Z"
   }
   ```

## Error Conditions

- **No current session**: Fails if `.current-session` file doesn't exist
- **Session not found**: Fails if session directory doesn't exist
- **Invalid transition**: Fails if session is not in PARKED status (e.g., already ACTIVE or WRAPPED)
- **Lock contention**: Fails if another process holds the session lock

## Example

```
/continue
```

## CLI Equivalent

```bash
ari session resume
```
