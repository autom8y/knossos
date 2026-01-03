# Worktree Integration

> How worktrees integrate with the broader ecosystem.

## Architecture

### Directory Structure

```
~/Code/project/                     # Main working tree
  .git/                             # Shared git database
  .claude/                          # Main ecosystem
  worktrees/                        # Worktree container
    .gitignore                      # Ignore all worktrees from git
    wt-20251224-143052-abc/         # Per-session worktree
      .claude/                      # Independent ecosystem
        .worktree-meta.json         # Worktree metadata
        agents/                     # Team agents (isolated)
        sessions/                   # Single session
        ACTIVE_TEAM                 # Team state (isolated)
        SPRINT_CONTEXT              # Sprint state (isolated)
```

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Detached HEAD** | No branch pollution, ephemeral by design |
| **Fresh CEM init** | Each worktree syncs independently |
| **Auto-session creation** | Worktree comes ready with session |
| **Subdirectory structure** | Clean organization under `worktrees/` |
| **7-day cleanup threshold** | Stale worktrees auto-cleaned |

## Session Lifecycle Integration

### /start Interaction

When `/start` detects an existing session, it offers worktree as an option:

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first
4. /worktree create "<name>" - Start in ISOLATED worktree (parallel work)
```

### /wrap Interaction

When completing a session in a worktree, `/wrap` offers cleanup:

```
Session Complete: billing-sprint
...
This session ran in an isolated worktree: wt-20251224-150000-xyz
Remove worktree? (y/n)
```

### /sessions --all

Shows sessions across all worktrees:

```bash
=== Main Project ===
session-20251224-143052-a1b2 | ACTIVE | feature-auth | 2025-12-24T14:30:52Z

=== Worktrees ===
[wt-20251224-150000-xyz] billing-sprint (team: 10x-dev-pack)
  session-20251224-150000-c3d4 | ACTIVE | billing-sprint
```

## Roster Integration

Worktree status can be checked via metadata:

```
Worktree Status
------------------------

[WORKTREE]    wt-20251224-150000-xyz
Name:         billing-sprint
Team:         10x-dev-pack

Roster:       /Users/user/Code/roster
...
```

### Worktree Functions

Worktree functions in `worktree-manager.sh`:
- `is_worktree()` - Detect worktree context
- `get_worktree_info()` - Read metadata
- `get_worktree_id()` - Extract ID from metadata

## Team Pack Integration

### swap-team.sh

Works identically in worktrees:
```bash
cd worktrees/wt-{id}
./swap-team.sh 10x-dev-pack
```

### ACTIVE_TEAM

Each worktree has its own ACTIVE_TEAM file, enabling different teams in parallel terminals.

## Git Integration

### Shared Database

All worktrees share the same `.git` database:
- Commits are visible across all worktrees
- Branches are shared
- Refs are shared

### Detached HEAD

Worktrees use detached HEAD to avoid branch pollution:
- No accidental commits to shared branches
- Clean separation of work
- Easy cleanup without branch cleanup
