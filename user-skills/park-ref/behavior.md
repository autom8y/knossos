# /park Behavior Specification

> Full step-by-step sequence for pausing a work session.

## Behavior Sequence

### 1. Pre-flight Validation

- **Check for active session**: Verify session exists (uses `get_session_dir()` from session-utils.sh)
  - If missing → Error: "No active session to park. Use `/start` to begin a session"
- **Check if already parked**: Verify `parked_at` field not already set
  - If set → Error: "Session already parked at {timestamp}. Use `/resume` to continue"

See [session-validation](../session-common/session-validation.md) for validation patterns.

### 2. Capture Current Work State

Gather current session state:

- **Git status**: Run `git status` to detect uncommitted changes
  - If uncommitted work → Add warning to park notes
- **Current phase**: Record from SESSION_CONTEXT `current_phase` field
- **Last agent**: Record from SESSION_CONTEXT `last_agent` field
- **Artifacts produced**: List from SESSION_CONTEXT `artifacts` array
- **Open questions**: Extract from SESSION_CONTEXT content
- **Blockers**: Extract from SESSION_CONTEXT `blockers` array

### 3. Generate Parking Summary

Create a human-readable summary of session state at park time.
See [parking-summary.md](parking-summary.md) for template.

### 4. Update SESSION_CONTEXT

Add park metadata to SESSION_CONTEXT YAML frontmatter:

```yaml
---
# ... existing fields ...
parked_at: "2025-12-24T15:30:00Z"
parked_reason: "{user-provided or 'Manual park'}"
parked_phase: "{current_phase}"
parked_git_status: "{clean|dirty}"
parked_uncommitted_files: 3  # if applicable
---
```

Append parking summary to SESSION_CONTEXT body (preserves existing content).

See [session-context-schema](../session-common/session-context-schema.md) for field definitions.

### 5. Save SESSION_CONTEXT

Write updated SESSION_CONTEXT to `.claude/sessions/{session_id}/SESSION_CONTEXT.md` file.

### 6. Confirmation

Display confirmation message with:
- Session name and park timestamp
- Reason for parking
- State summary (phase, agent, artifacts)
- Git status warning if dirty
- Resume instructions

---

## State Changes

### Fields Modified

| Field | Value | Description |
|-------|-------|-------------|
| `parked_at` | ISO 8601 timestamp | When session was parked |
| `parked_reason` | User-provided string or "Manual park" | Why work was paused |
| `parked_phase` | Current phase value | Phase at park time |
| `parked_git_status` | "clean" or "dirty" | Git working directory state |
| `parked_uncommitted_files` | Integer (if dirty) | Count of uncommitted files |

### Content Additions

- Parking summary appended to SESSION_CONTEXT body
- Preserves all existing content and metadata

---

## Error Cases

| Error | Condition | Resolution |
|-------|-----------|------------|
| No active session | No session for current project | Use `/start` to begin a new session |
| Already parked | `parked_at` field already set | Use `/resume` to continue, or check session status |
| File write error | Permission denied on SESSION_CONTEXT | Check file permissions, ensure not read-only |

---

## Design Notes

### Why Preserve Git Status?

Git status at park time helps detect:
1. **Stale work**: If files changed outside the session
2. **Incomplete work**: If work was paused mid-implementation
3. **Merge conflicts**: If branch diverged during park period

This enables `/resume` to warn about potential issues before continuing.

### Why Append vs Overwrite?

Parking summaries are appended (not overwriting) to create an audit trail. Multiple park/resume cycles preserve the full session history, making it easier to understand context when resuming days or weeks later.

### Idempotency

Parking an already-parked session is an error (not idempotent) because:
1. It indicates user confusion about session state
2. Multiple park timestamps would create ambiguity
3. `/status` should be used to check state first
