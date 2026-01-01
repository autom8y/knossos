# state-mate Invocation Pattern

> Delegate session state mutations to state-mate agent via Task tool.

## When to Apply

Commands that mutate SESSION_CONTEXT:
- /park - sets parked_at, parked_reason, etc.
- /resume - clears park fields, sets resumed_at
- /wrap - transitions to ARCHIVED state

## Task Invocation Template

```
Task(state-mate, "{operation}

Session Context:
- Session ID: {session_id}
- Session Path: .claude/sessions/{session_id}/SESSION_CONTEXT.md")
```

### Operations

| Operation | Command | Mutations |
|-----------|---------|-----------|
| `park_session reason='{reason}'` | /park | Set parked_at, parked_reason |
| `resume_session` | /resume | Clear parked_*, set resumed_at |
| `wrap_session` | /wrap | Set completed_at, archive |

## Response Handling

### Success Response

```json
{
  "success": true,
  "operation": "{operation_name}",
  "message": "Session {operation} successfully",
  "state_before": { "session_state": "..." },
  "state_after": { "session_state": "...", ... }
}
```

**Action**: Parse response, display confirmation to user.

### Failure Response

```json
{
  "success": false,
  "error_type": "LIFECYCLE_VIOLATION",
  "message": "Cannot {operation}: {reason}",
  "hint": "Use /{suggested_command} first"
}
```

**Action**: Surface error message and hint to user.

## Error Types

| Error Type | Cause | Recovery |
|------------|-------|----------|
| `LIFECYCLE_VIOLATION` | Invalid state transition | Follow hint (e.g., resume before wrap) |
| `VALIDATION_ERROR` | Missing required field | Provide missing data |
| `UNAVAILABLE` | state-mate not responding | Retry or check agent configuration |

## Implementation

```
1. Get session context:
   session_id=$(session-manager.sh status | jq -r '.session_id')
   session_path=".claude/sessions/${session_id}/SESSION_CONTEXT.md"

2. Invoke state-mate via Task tool:
   Task(state-mate, "{operation}

   Session Context:
   - Session ID: {session_id}
   - Session Path: {session_path}")

3. Parse JSON response:
   - If success: true → Extract state_after, continue
   - If success: false → Extract message and hint, surface to user

4. Post-operation (command-specific):
   - /park: Display parking summary
   - /resume: Invoke selected agent
   - /wrap: Archive session directory
```

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `operation` | state-mate operation name | All |
| `reason` | User-provided reason | park |
| `post_action` | Action after successful mutation | All |

## Cross-Reference

- state-mate agent: `.claude/agents/state-mate.md`
- ADR: `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md`
