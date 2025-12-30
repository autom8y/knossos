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

### 5. Invoke state-mate for Resume Mutation

Use Task tool to invoke state-mate agent:

```
Task(state-mate, "resume_session

Session Context:
- Session ID: {session_id}
- Session Path: .claude/sessions/{session_id}/SESSION_CONTEXT.md")
```

**Expected Response** (JSON):
```json
{
  "success": true,
  "operation": "resume_session",
  "message": "Session resumed successfully",
  "state_before": { "session_state": "PARKED" },
  "state_after": { "session_state": "ACTIVE", "resumed_at": "..." }
}
```

**Error Handling**:
- If state-mate returns `success: false` (e.g., session not parked), surface error
- If LIFECYCLE_VIOLATION, show allowed transitions from state-mate response

### 6. Update Agent Selection (Post-Resume)

After successful resume, if agent override was specified, invoke:

```
Task(state-mate, "update_field last_agent='{selected_agent}'

Session Context:
- Session ID: {session_id}
- Session Path: .claude/sessions/{session_id}/SESSION_CONTEXT.md")
```

### 7. Invoke Selected Agent

Use Task tool to invoke agent with full session context:
- Initiative, complexity, phase
- Park duration and reason
- Artifacts, blockers, questions
- Next steps

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
