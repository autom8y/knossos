---
name: park
description: Pause work session and preserve state for later
argument-hint: "[reason]"
allowed-tools: Bash, Read, Task
model: sonnet
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Pause the current work session and save state for later resumption. $ARGUMENTS

## Session Resolution

The current session is resolved automatically by `ari session` via the session marker:
```bash
# ari handles resolution internally
ari session status -o json
```

## Pre-flight

1. Verify an active session exists (`ari session status` succeeds)
2. Check session is not already parked

## Behavior

1. **Capture state**:
   - Current phase and last agent
   - Artifacts produced so far
   - Git status (warn about uncommitted changes)
   - Open questions and blockers

2. **Delegate to Moirai** for session state mutation:
   ```
   Task(moirai, "park_session reason=\"<REASON>\"")
   ```
   Moirai will:
   - Acquire lock to prevent race conditions
   - Execute `ari session park --reason="<REASON>"`
   - Validate the result and log to audit trail
   - Return structured response

3. **Display parking summary** to user:
   - Duration so far
   - Progress (completed/in-progress artifacts)
   - Next steps when resuming
   - Park reason and timestamp

## Example

```
/park "Waiting for stakeholder feedback on PRD"
```

Output:
```
Session parked at 2025-12-24 15:30

Progress: PRD complete, TDD in progress
Duration: 2h 15m
Reason: Waiting for stakeholder feedback on PRD

Resume with: /continue
```

## Reference

Full documentation: `.claude/commands/session/park/INDEX.md`
