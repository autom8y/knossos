---
name: worktree-ref
description: "Git worktree management for parallel Claude sessions with filesystem isolation. Use when: running parallel sessions, needing different teams per terminal, isolating sprint context, creating ephemeral workspaces. Triggers: /worktree, git worktree, parallel sessions, session isolation, worktree create, worktree cleanup."
---

# Worktree Management Reference

> Per-session git worktree isolation for parallel Claude sessions

## Overview

Git worktrees provide true filesystem isolation for parallel Claude sessions. Each worktree has its own `.claude/` directory with independent agents, sessions, sprints, and team configuration.

**The Problem**: Claude Code requires a local `.claude/` directory structure that cannot be configured per-terminal. When multiple terminals work on the same project with different teams/sprints, they collide on shared files.

**The Solution**: Git worktrees create separate working directories that share the same git database but have independent file systems.

## Quick Reference

| Command | Description |
|---------|-------------|
| `/worktree create [name]` | Create isolated worktree with CEM init |
| `/worktree list` | Show all worktrees with status |
| `/worktree status [id]` | Detailed worktree info |
| `/worktree remove <id>` | Remove worktree (warns on changes) |
| `/worktree cleanup` | Remove stale worktrees (7+ days) |
| `/worktree gc` | Garbage collect orphaned refs |

## Common Flags

| Flag | Description |
|------|-------------|
| `--team=PACK` | Team pack to use (default: current) |
| `--from=REF` | Git ref to base on (default: HEAD) |
| `--force` | Override safety checks |

## Typical Workflow

```bash
# 1. Create worktree for parallel work
/worktree create "feature-billing" --team=10x-dev-pack

# 2. Navigate in new terminal
cd worktrees/wt-{id} && claude

# 3. Work independently (different team, session, sprint)

# 4. Complete and cleanup
/wrap
# Prompted: "Remove worktree? (y/n)"
```

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Detached HEAD** | No branch pollution, ephemeral by design |
| **Fresh CEM init** | Each worktree syncs independently |
| **Auto-session creation** | Worktree comes ready with session |
| **7-day cleanup threshold** | Stale worktrees auto-cleaned |

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Commit worktree changes to main | Worktrees are ephemeral | Cherry-pick specific commits |
| Share session across worktrees | Breaks isolation | One session per worktree |
| Leave stale worktrees | Accumulates disk usage | Run `/worktree cleanup` regularly |
| Use worktree for long-term branches | Not designed for persistence | Use standard git branches |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full command reference and implementation details
- [examples.md](examples.md) - Workflow scenarios and integration examples
- [troubleshooting.md](troubleshooting.md) - Common issues and solutions
- [integration.md](integration.md) - CEM, session, and team pack integration

## Related Commands

| Command | When to Use |
|---------|-------------|
| `/start` | Begin session (offers worktree option if session exists) |
| `/wrap` | Complete session (offers worktree cleanup) |
| `/sessions --all` | View sessions across all worktrees |
