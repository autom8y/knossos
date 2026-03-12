# Session Resolution Pattern

> Validate session existence and state before command execution.

## When to Apply

All session-lifecycle commands that require an existing session:
- /sos park - requires active session
- /sos resume - requires parked session
- /sos wrap - requires active session
- /handoff - requires active session

/sos start is the exception: it requires NO existing session.

## Validation Checks

| Check | Function | Pass | Fail |
|-------|----------|------|------|
| Session exists | `get_session_dir()` | Directory exists | Error: No active session |
| Session not parked | `parked_at` field absent | Field not set | Error: Session parked |
| Session is parked | `parked_at` field present | Field set | Error: Session not parked |

## Implementation

```
1. Call get_session_dir() from session-utils.sh
   - Returns: Session directory path or empty
   - If empty: Error "No active session to {verb}. Use /start to begin."

2. Read SESSION_CONTEXT.md frontmatter
   - Extract parked_at field

3. Validate state against command requirements:
   - /sos park requires: parked_at NOT set
   - /sos resume requires: parked_at IS set
   - /sos wrap requires: parked_at NOT set (or offer auto-resume)
   - /handoff requires: parked_at NOT set
```

## Error Messages

| Condition | Message Template |
|-----------|------------------|
| No session | "No active session to {verb}. Use `/sos start` to begin." |
| Already parked | "Session parked at {timestamp}. Use `/sos resume` first." |
| Not parked | "Session is already active (not parked). Continue working." |

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `verb` | Action verb for error message | All |
| `require_parked` | Whether session must be parked | resume only |
| `auto_resume_offer` | Offer to auto-resume if parked | wrap only |

## Cross-Reference

- Schema: [session-context-schema](../common/session-context-schema.md)
- State machine: [session-phases](../common/session-phases.md)
