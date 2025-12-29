# /resume Behavior Specification

> Full step-by-step sequence for resuming a parked work session.

## Behavior Sequence

### 1. Pre-flight Validation

- **Check for parked session**: Verify session exists (uses `get_session_dir()`)
  - If missing → Error: "No parked session found. Use `/start` to begin"
- **Check park status**: Verify `parked_at` field is set
  - If not set → Error: "Session is already active (not parked). Continue working"

See [session-validation](../session-common/session-validation.md) for validation patterns.

### 2. Load and Display Session Summary

Read SESSION_CONTEXT and display:
- Session details (started, parked, reason, complexity, team, phase, agent)
- Artifacts produced
- Blockers and open questions
- Next steps from park time

### 3. Validate Context

Perform context validation checks. See [validation-checks.md](validation-checks.md) for details.

**Team Consistency Check**: Compare `ACTIVE_TEAM` to `session.active_team`
- If mismatch: Offer to switch back or continue

**Git Status Check**: Compare current status to `parked_git_status`
- If new changes: Surface and offer to review diff

### 4. Agent Selection

Confirm which agent to continue with:
- Default: `last_agent` from SESSION_CONTEXT
- Override: `--agent=NAME` parameter
- Validate agent exists in current team

### 5. Remove Park Metadata

Update SESSION_CONTEXT by removing:
```yaml
# Remove these fields:
parked_at
parked_reason
parked_phase
parked_git_status
parked_uncommitted_files
```

See [session-context-schema](../session-common/session-context-schema.md) for field definitions.

### 6. Invoke Selected Agent

Use Task tool to invoke agent with full session context:
- Initiative, complexity, phase
- Park duration and reason
- Artifacts, blockers, questions
- Next steps

### 7. Update SESSION_CONTEXT

Add resume metadata:
```yaml
resumed_at: "ISO timestamp"
resume_count: {increment or set to 1}
last_agent: "{selected-agent}"
```

### 8. Confirmation

Display:
- Session name and park duration
- Selected agent and phase
- First item from next steps
- Available commands

---

## State Changes

### Fields Modified

| Field | Action | Description |
|-------|--------|-------------|
| `parked_at` | REMOVED | Park timestamp deleted |
| `parked_reason` | REMOVED | Park reason deleted |
| `parked_phase` | REMOVED | Phase at park time deleted |
| `parked_git_status` | REMOVED | Git status deleted |
| `parked_uncommitted_files` | REMOVED | File count deleted |
| `resumed_at` | ADDED | Resume timestamp |
| `resume_count` | ADDED | Park/resume cycle count |
| `last_agent` | UPDATED | Agent selected for resume |

---

## Error Cases

| Error | Condition | Resolution |
|-------|-----------|------------|
| No parked session | No session for project | Use `/start` to begin new session |
| Session not parked | `parked_at` not set | Session is active, continue working |
| Invalid agent | Agent not in team | Specify valid agent or switch teams |
| Team unavailable | Session team not in roster | Install team pack or choose different |
| Merge conflicts | Git detects conflicts | Resolve conflicts before resuming |

---

## Design Notes

### Why Allow Agent Override?

Session phases evolve. A session parked during design (last_agent: architect) may be ready for implementation when resumed. `--agent` override supports natural phase transitions without separate `/handoff`.

### Resume Count Tracking

`resume_count` helps identify:
1. Frequently interrupted sessions (potential issues)
2. Sessions needing smaller scope
3. External dependencies causing delays
