# Session Validation

> Pre-flight validation patterns reused across session commands.

## Common Validation Checks

### Session Exists

```
Check: .claude/sessions/{session_id}/SESSION_CONTEXT.md exists
Pass: Continue command execution
Fail: "No active session. Use /start to begin."
```

### Session Not Parked

```
Check: parked_at field is NOT set in frontmatter
Pass: Continue command execution
Fail: "Session parked at {timestamp}. Use /resume first."
```

### Session Is Parked

```
Check: parked_at field IS set in frontmatter
Pass: Continue /resume execution
Fail: "Session not parked. Continue working or check /status."
```

### Agent Exists

```
Check: .claude/agents/{agent}.md exists
Pass: Continue with agent invocation
Fail: "Agent '{agent}' not found in team '{active_team}'."
```

### Team Consistency

```
Check: SESSION_CONTEXT.active_team matches ACTIVE_RITE file
Pass: Continue without warning
Fail: Prompt user for team switch or override
```

## Error Message Patterns

| Pattern | Format |
|---------|--------|
| Missing session | "No active session to {verb}. Use /start to begin." |
| Already parked | "Session parked at {timestamp}. Use /resume first." |
| Invalid agent | "Agent '{name}' not found in team '{team}'." |
| Team mismatch | "Session team ({A}) differs from active rite ({B})." |
