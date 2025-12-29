# Session Lifecycle Anti-Patterns

> Common mistakes across session commands and their corrections.

| Do NOT | Why | Instead |
|--------|-----|---------|
| Store session state in CLAUDE.md | Changes every session, creates stale content | Use SESSION_CONTEXT in `.claude/sessions/` |
| Hardcode timestamps in SESSION_CONTEXT body | Becomes stale, duplicates frontmatter | Use YAML frontmatter for all timestamps |
| Allow /handoff while parked | Target agent lacks context restoration | Require /resume first, then /handoff |
| Skip quality gates in /wrap without flag | Incomplete work ships silently | Require explicit --skip-checks |
| Delete SESSION_CONTEXT without archive | Loses audit trail and debugging context | Archive to .claude/.archive/sessions/ by default |
| Update parked session state directly | Creates inconsistent state | Use /resume to reactivate first |
| Allow multiple active sessions per project | Causes context confusion | Enforce single-session constraint |
| Store git status as string without enum | Ambiguous comparisons | Use enum: clean, dirty |
| Increment handoff_count for same-agent | Inflates metrics, hides actual flow | Warn and skip increment |
| Allow /wrap on parked session | Park state conflicts with completion | Auto-resume or reject with guidance |
