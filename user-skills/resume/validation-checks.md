# Resume Validation Checks

> Context validation for resuming parked sessions.

## Team Consistency Check

### Purpose

Team packs contain different agents. If session started with `10x-dev-pack` but current team is `doc-team-pack`, expected agents may not be available.

### Check Logic

```
Read current .claude/ACTIVE_TEAM
Compare to SESSION_CONTEXT.active_team
If different → Surface mismatch warning
```

### Mismatch Handling

```
⚠ Team Mismatch Detected

Session started with: {session.active_team}
Current active team: {current ACTIVE_TEAM}

This session's agents may not be available in the current team.

Options:
1. Switch back to {session.active_team} (recommended)
2. Continue with {current ACTIVE_TEAM} (may cause issues)
3. Cancel resume

Choice [1/2/3]:
```

**Option 1**: Invoke `~/Code/roster/swap-team.sh {session.active_team}`
**Option 2**: Continue with potential agent mismatch
**Option 3**: Abort resume

---

## Git Status Check

### Purpose

Git changes since park indicate:
1. **Concurrent work**: Files modified outside the session
2. **Merge issues**: Branch diverged, conflicts possible
3. **Stale session**: Session may no longer be relevant

### Check Logic

```
Run git status
Compare to parked_git_status from SESSION_CONTEXT
If mismatch (was clean, now dirty) → Surface warning
If new uncommitted files → List changes
```

### Change Handling

```
⚠ Git Changes Detected

Git status at park time: {parked_git_status}
Current git status: {current status}

New/modified files since park:
- {file1}
- {file2}

This may indicate:
1. Work done outside this session
2. Merge conflicts from branch updates
3. Unrelated changes

Review changes before continuing? [y/n]:
```

**If yes**: Display `git diff --stat` output
**Either way**: User decides whether to continue

---

## Agent Availability Check

### Purpose

Validate selected agent exists in current team.

### Check Logic

```
Determine target agent:
  - --agent parameter if provided
  - Otherwise: SESSION_CONTEXT.last_agent

Check .claude/agents/{agent}.md exists
If not found → Error with agent list
```

### Error Message

```
Agent '{agent}' not found in team '{active_team}'.

Available agents:
- requirements-analyst
- architect
- principal-engineer
- qa-adversary

Use --agent=NAME to specify a valid agent.
```

---

## Validation Flow Summary

```
1. Pre-flight (session exists, is parked)
      ↓
2. Team consistency check
   ├─ Match → Continue
   └─ Mismatch → User chooses action
      ↓
3. Git status check
   ├─ Clean/unchanged → Continue
   └─ Dirty/new files → User reviews
      ↓
4. Agent availability check
   ├─ Found → Continue to invoke
   └─ Not found → Error with guidance
```
