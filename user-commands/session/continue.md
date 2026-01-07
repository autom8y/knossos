---
description: Resume a parked work session with full context
argument-hint:
allowed-tools: Bash, Read, Write, Task
model: sonnet
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

1. **Acquire exclusive lock** on the session to prevent race conditions

2. **Load and validate session context**:
   - Read SESSION_CONTEXT.md
   - Verify FSM allows PARKED -> ACTIVE transition

3. **Update session state**:
   - Set status to ACTIVE
   - Set `resumed_at` timestamp
   - Clear `parked_at` timestamp
   - Clear `parked_reason` field

4. **Save updated context** to SESSION_CONTEXT.md

5. **Emit resume event** to session event log

6. **Display resumption summary**:
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
