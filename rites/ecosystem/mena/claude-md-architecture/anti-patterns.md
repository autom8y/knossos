# CLAUDE.md Anti-Patterns

What NOT to put in CLAUDE.md. Each anti-pattern includes the violation, why it's wrong, and the correct approach.

---

## Anti-Pattern Summary Table

| # | Anti-Pattern | Why Wrong | Correct Location |
|---|--------------|-----------|------------------|
| 1 | Session state in CLAUDE.md | Changes every session, creates maintenance burden | SESSION_CONTEXT |
| 2 | Dynamic git references | Stale after any git operation | Hook injection |
| 3 | Duplicating knossos state | Creates two sources of truth, can desync | ACTIVE_RITE + agents/ |
| 4 | Personal preferences in project | User-specific, causes team conflicts | ~/.claude/CLAUDE.md |
| 5 | Hardcoded dynamic values | Sprint/velocity/blockers become stale quickly | Project management tools |
| 6 | Session history | CLAUDE.md is not a changelog | .claude/sessions/ |
| 7 | Knossos rite in satellites | Satellite has its own rite | Regenerate from satellite's ACTIVE_RITE |
| 8 | "Last updated" timestamps | Immediately stale, git provides this | Git history |
| 9 | Task lists/checkboxes | Session-scoped, better tools exist | TodoWrite, issue tracker |
| 10 | Environment configuration | Security risk, varies by machine | .env files, secret managers |
| 11 | Misplaced ownership markers | Sync pipeline cannot parse correctly | Marker BEFORE section header |

---

## Detailed Anti-Patterns

### 1. Session State in CLAUDE.md

**Violation**: `## Current Work: Currently implementing auth...`

**Why wrong**: Changes every session, stale within hours, clutters behavioral contract.

**Correct**: Session state in `SESSION_CONTEXT` (`.claude/sessions/session-id/`), injected by hooks.

---

### 2. Dynamic Git References

**Violation**: `Branch: feature/auth` or `Uncommitted: 5 files`

**Why wrong**: Changes constantly, stale immediately after any git operation.

**Correct**: Hooks inject git state at session start (never written to file).

---

### 3. Duplicating Knossos State

**Violation**: `Active Rite: docs, Swapped: 2024-12-25`

**Why wrong**: Duplicates ACTIVE_RITE file, dates become stale, creates two sources of truth.

**Correct**: Rite sections regenerated from ACTIVE_RITE + agents/ directory.

---

### 4. Personal Preferences in Project File

**Violation**: `## My Preferences: I prefer TypeScript, 2-space indent`

**Why wrong**: User-specific, creates conflict in shared repos.

**Correct**: `~/.claude/CLAUDE.md` for personal preferences, `.editorconfig` for project conventions.

---

### 5. Hardcoded Dynamic Values

**Violation**: `Sprint: 23 (ends Jan 15), Velocity: 42, Blockers: waiting on API review`

**Why wrong**: All change frequently, all become stale quickly.

**Correct**: Project management tools (Jira, Linear), PRD, or session files.

---

### 6. Session History

**Violation**: `## Session History: 2024-12-24 worked on auth...`

**Why wrong**: CLAUDE.md is not a changelog, grows without bound.

**Correct**: Session summaries in `.claude/sessions/`, use `/sessions` command to list.

---

### 7. Copying Knossos Rite to Satellites

**Violation**: Satellite Quick Start shows `ecosystem` agents instead of satellite's own rite.

**Why wrong**: Satellite has its own rite, knossos rite is irrelevant, creates incorrect routing.

**Correct**: Rite sections are PRESERVE or REGENERATE from satellite's own ACTIVE_RITE, never SYNC.

---

### 8. "Last Updated" Timestamps

**Violation**: `Last updated: 2024-12-25 14:30, Version: 1.4.2`

**Why wrong**: Immediately stale, requires manual updates, git provides this automatically.

**Correct**: Use `git log --oneline -5 .claude/CLAUDE.md` for history.

---

### 9. Task Lists in CLAUDE.md

**Violation**: `## TODO: - [ ] Implement login, - [x] Create model`

**Why wrong**: Task state is session-scoped, completed tasks become noise, better tools exist.

**Correct**: TodoWrite tool for session tasks, issue tracker for project tasks.

---

### 10. Environment-Specific Configuration

**Violation**: `Database: postgresql://localhost:5432, API_KEY: sk-1234`

**Why wrong**: Environment varies by machine, secrets should never be in version control.

**Correct**: `.env` files (gitignored), environment variables, secret managers.

---

### 11. Misplaced or Missing Ownership Markers

**Violation**: Marker appears AFTER section header, or no marker at all.

**Why wrong**: Sync pipeline parses markers to determine ownership; misplaced markers cause sync errors.

**Correct**:
```markdown
<!-- PRESERVE: satellite-owned -->
## Quick Start
```

Rules:
1. Marker on line immediately BEFORE section header
2. No blank line between marker and `## Header`
3. Every section should have an ownership marker

---

## Quick Reference: Red Flags

| If You See... | It's Wrong Because... | Move To... |
|---------------|----------------------|------------|
| "Currently working on X" | Session state | SESSION_CONTEXT |
| "Last updated: DATE" | Stale metadata | Git history |
| "Git branch: X" | Changes constantly | Hook output |
| Knossos rite in satellite | Wrong rite source | Regenerate from satellite's ACTIVE_RITE |
| Personal preferences | Wrong scope | ~/.claude/CLAUDE.md |
| Sprint/initiative details | Too dynamic | PRD, session files |
| Task checkboxes | Session-scoped | TodoWrite, issue tracker |
| Hardcoded secrets | Security risk | .env, secret manager |

---

## The Quick Test

Before adding content to CLAUDE.md:

1. **Stale tomorrow?** -> Do not add
2. **Single session scope?** -> SESSION_CONTEXT
3. **Personal, not project-wide?** -> ~/.claude/CLAUDE.md
4. **Duplicates another source?** -> Reference, do not duplicate
5. **Claude can work without it?** -> Consider removing

---

## Related Files

- [first-principles.md](first-principles.md) - Core architectural principles
- [ownership-model.md](ownership-model.md) - Section ownership details
- [boundary-test.md](boundary-test.md) - Validation checklist
