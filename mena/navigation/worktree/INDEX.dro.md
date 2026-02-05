---
name: worktree
description: Manage isolated worktrees for parallel Claude sessions
argument-hint: <create|list|remove|cleanup|status> [args]
allowed-tools: Bash, Read
model: sonnet
---

## Context

Git worktrees provide true filesystem isolation for parallel Claude sessions. Each worktree has its own `.claude/` directory with independent agents, sessions, sprints, and rite configuration.

## Pre-flight

1. **Git repository required**:
   - Verify in git repository: `git rev-parse --git-dir`
   - If not: ERROR "Not in a git repository. Worktrees require git."

2. **Git worktree support**:
   - Check git version supports worktrees (2.5+)

## Your Task

$ARGUMENTS

## Commands

### create [name] [--rite=PACK] [--from=REF]

Create a new isolated worktree:

```bash
hooks/lib/worktree-manager.sh create "$NAME" --rite="$RITE"
```

**What happens:**
1. Creates git worktree with detached HEAD (no branch pollution)
2. Initializes ecosystem (fresh sync from roster)
3. Sets rite if specified
4. Creates initial session
5. Returns path for user to `cd` into

**Output to user:**
- Worktree ID and path
- Instructions: `cd <path> && claude`

### list

Show all worktrees:

```bash
hooks/lib/worktree-manager.sh list
```

Display as table:
| ID | Name | Team | Created | Session | Changes |

### status [id]

Detailed worktree info:

```bash
hooks/lib/worktree-manager.sh status "$ID"
```

### remove <id> [--force]

Remove a worktree:

```bash
hooks/lib/worktree-manager.sh remove "$ID" [--force]
```

**Pre-checks:**
- Warn if uncommitted changes exist
- Require --force to override

### cleanup [--force]

Remove stale worktrees (7+ days old, no changes):

```bash
hooks/lib/worktree-manager.sh cleanup [--force]
```

### gc

Garbage collect orphaned refs:

```bash
hooks/lib/worktree-manager.sh gc
```

## Examples

```bash
# Create worktree for auth sprint with 10x rite
/worktree create "auth-sprint" --rite=10x-dev

# Create worktree from specific branch
/worktree create "hotfix" --from=release-1.2

# List all worktrees
/worktree list

# Remove specific worktree
/worktree remove wt-20251224-143052-abc

# Cleanup stale worktrees
/worktree cleanup
```

## Typical Workflow

1. In main project, want parallel work:
   ```
   /worktree create "feature-x" --rite=10x-dev
   ```

2. Open new terminal, navigate to worktree:
   ```bash
   cd ~/Code/project/worktrees/wt-20251224-150000-xyz
   claude
   ```

3. Work in complete isolation (different team, sprint, session)

4. When done:
   ```
   /wrap  # Will offer to remove worktree
   ```

## Reference

Full documentation: `.claude/commands/navigation/worktree/INDEX.md`
