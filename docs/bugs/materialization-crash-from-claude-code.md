# Bug: `ari sync materialize` Crashes Claude Code When Run From Within Session

## Severity: P1 (blocks developer workflow)

## Reproduction

1. Open Claude Code in a knossos project (`/Users/tomtenuta/Code/roster`)
2. Run `ari sync materialize --force` via Bash tool (or let an agent run it)
3. Claude Code crashes or hangs at "Running PreToolUse hook... Running..."

Also triggers when deleting `.claude/CLAUDE.md` directly:
1. Run `rm .claude/CLAUDE.md` from within Claude Code
2. Claude Code crashes shortly after (detects file removal)

## Root Cause

`ari sync materialize` modifies files that Claude Code actively depends on mid-session:
- `.claude/CLAUDE.md` — Claude Code's project instructions (loaded into system prompt)
- `.claude/commands/` — Skill registry (slash commands)
- `.claude/skills/` — Skill registry (reference skills)
- `.claude/KNOSSOS_MANIFEST.yaml` — Inscription state

When Claude Code detects changes to these files (via filesystem watching or post-tool-use reload), the mid-session mutation causes a crash or hang. The same issue occurs when `.claude/CLAUDE.md` is deleted — Claude Code attempts to reload and fails.

## Impact

- Cannot run `ari sync materialize` from within Claude Code
- Cannot have agents (C3, compatibility-tester) verify materialization output in-session
- Any agent task that modifies `.claude/` directory structure risks crashing the parent session

## Workaround

Run materialization from a **separate terminal**, not from within Claude Code:

```bash
# In a separate terminal (not Claude Code)
cd /path/to/project
ari sync materialize --force
```

Then restart Claude Code to pick up the changes.

## Recommended Fixes

### Short-term (P0): Guard in CLI
Add a check in `ari sync materialize` that detects when running inside Claude Code (via `CLAUDE_SESSION_ID` or similar env var) and either:
- Warns the user to run from a separate terminal
- Defers the actual file writes until after the Claude Code session ends

### Medium-term: Atomic materialization
Write all output to a temporary directory, then atomically swap:
1. Write `.claude.new/` with all generated content
2. Rename `.claude/` → `.claude.bak/`
3. Rename `.claude.new/` → `.claude/`
4. Delete `.claude.bak/`

This minimizes the window where files are in an inconsistent state.

### Long-term: Claude Code API for config reload
Request Claude Code support a "reload configuration" signal so materialization can notify it to reload cleanly rather than crashing on filesystem changes.

## Related

- D2 lock audit: Lock file in `.claude/sessions/.locks/` also modified during materialization
- C3 verification: Could not complete gate verification due to this bug
- ADR-0022: Session model should account for "no self-modification during active session" constraint
