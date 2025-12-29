# /wrap Behavior Specification

> Full step-by-step sequence for completing a work session.

## Behavior Sequence

### 1. Pre-flight Validation

- **Check for active session**: Verify session exists
  - If missing → Error: "No active session to wrap"
- **Check not parked**: Verify `parked_at` not set
  - If parked → Warning: "Session is parked. Resume before wrapping? [y/n]"
  - If yes → Auto-invoke `/resume`, then continue wrap

See [session-validation](../session-common/session-validation.md) for patterns.

### 2. Run Quality Gates (unless --skip-checks)

Validate artifacts based on session complexity. See [quality-gates.md](quality-gates.md).

Quality gates run in order:
1. PRD Quality Gate (all complexity levels)
2. TDD Quality Gate (MODULE+)
3. Code Quality Gate (if engineer was last_agent)
4. Validation Quality Gate (if qa was last_agent)

If any gate fails: Surface issues, offer resolution options.

### 3. Optional: Invoke QA for Final Review

If `last_agent` is not `qa-adversary`:
- Offer final QA review before wrapping
- If yes: Invoke QA Adversary via Task tool
- Wait for response; abort if issues found

### 4. Generate Session Summary

Create comprehensive summary. See [session-summary.md](session-summary.md).

Summary includes:
- Session metadata (ID, duration, complexity)
- Accomplishments and artifacts
- Key decisions and ADRs
- Quality metrics and test results
- Workflow history (handoffs, parks)
- Next session starting point

### 5. Archive or Delete SESSION_CONTEXT

**If `--archive` (default)**:
- Move to `.claude/.archive/sessions/{session_id}/`
- Preserves full session history

**If not `--archive`**:
- Delete session directory
- Session state removed (summary still available)

### 6. Save Session Summary

Write to `/docs/sessions/SUMMARY-{session_id}.md`

### 7. Update Session Index

Append entry to `/docs/sessions/INDEX.md`:
```markdown
| Date | Session ID | Initiative | Complexity | Duration | Artifacts |
```

Create index if it doesn't exist.

### 8. Confirmation

Display:
- Session name and duration
- Quality gate status
- Artifacts delivered
- Summary location
- Next session starting point

---

## State Changes

### Files Created

- `/docs/sessions/SUMMARY-{session_id}.md` - Summary report
- `/docs/sessions/INDEX.md` - Session index (if first wrap)
- `.claude/.archive/sessions/{session_id}/` - Archived directory

### Files Modified

- `/docs/sessions/INDEX.md` - New entry appended

### Files Deleted/Moved

- `.claude/sessions/{session_id}/` - Removed or moved to archive

---

## Error Cases

| Error | Condition | Resolution |
|-------|-----------|------------|
| No active session | Session missing | Already completed or never started |
| Quality gates failing | Artifacts incomplete | Fix issues or use --skip-checks |
| Session parked | `parked_at` set | Use `/resume` first |
| Git dirty | Uncommitted changes | Commit or use --skip-checks |
| Tests failing | Test suite errors | Fix tests or use --skip-checks |

---

## Design Notes

### Why Archive by Default?

Preserves audit trail, debugging context, and compliance requirements. Deletion available but discouraged.

### Why Session Index?

Provides discoverability, metrics tracking, and project history - "git log for work sessions."
