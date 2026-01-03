# Worktree Behavior Reference

> Full command reference and implementation details for git worktree management.

## Commands

### `/worktree create [name] [--team=PACK] [--from=REF]`

Create a new isolated worktree with full ecosystem initialization.

**Parameters:**
- `name` - Descriptive name for the worktree (default: "unnamed")
- `--team=PACK` - Team pack to use (default: current team)
- `--from=REF` - Git ref to base worktree on (default: HEAD)

**What Happens:**
1. Creates git worktree with detached HEAD (no branch pollution)
2. Initializes ecosystem (fresh sync from roster)
3. Sets team if specified via swap-team.sh
4. Creates initial session via session-manager.sh
5. Returns path for user to `cd` into

**Example:**
```bash
/worktree create "auth-sprint" --team=10x-dev-pack
# Output:
# {
#   "success": true,
#   "worktree_id": "wt-20251224-143052-abc",
#   "path": "worktrees/wt-20251224-143052-abc",
#   "instructions": "cd worktrees/wt-20251224-143052-abc && claude"
# }
```

### `/worktree list`

Show all worktrees with status information.

**Output:**
```json
{
  "worktrees": [
    {
      "id": "wt-20251224-143052-abc",
      "path": "worktrees/wt-20251224-143052-abc",
      "name": "auth-sprint",
      "team": "10x-dev-pack",
      "created_at": "2025-12-24T14:30:52Z",
      "has_changes": false,
      "session_status": "active"
    }
  ],
  "count": 1
}
```

### `/worktree status [id]`

Detailed status of a specific worktree or all worktrees.

**With ID:**
```json
{
  "worktree_id": "wt-20251224-143052-abc",
  "created_at": "2025-12-24T14:30:52Z",
  "name": "auth-sprint",
  "from_ref": "HEAD",
  "team": "10x-dev-pack",
  "complexity": "MODULE",
  "parent_project": "/Users/user/Code/project"
}
```

### `/worktree remove <id> [--force]`

Remove a specific worktree.

**Pre-checks:**
- Warns if uncommitted changes exist
- Requires `--force` to override

**Example:**
```bash
/worktree remove wt-20251224-143052-abc
# Error: Worktree has uncommitted changes. Use --force to override.

/worktree remove wt-20251224-143052-abc --force
# { "success": true, "removed": "wt-20251224-143052-abc" }
```

### `/worktree cleanup [--force]`

Remove stale worktrees (7+ days old, no uncommitted changes).

**Example:**
```bash
/worktree cleanup
# { "removed": 2, "skipped": 1, "cutoff_days": 7 }
```

### `/worktree gc`

Garbage collect orphaned worktree refs.

**Example:**
```bash
/worktree gc
# { "pruned": true, "remaining_worktrees": 3 }
```

## Worktree Metadata

Each worktree stores metadata in `.claude/.worktree-meta.json`:

```json
{
  "worktree_id": "wt-20251224-143052-abc",
  "created_at": "2025-12-24T14:30:52Z",
  "name": "auth-sprint",
  "from_ref": "HEAD",
  "team": "10x-dev-pack",
  "complexity": "MODULE",
  "parent_project": "/Users/user/Code/project"
}
```

## Implementation Details

### worktree-manager.sh

Core script at `.claude/hooks/lib/worktree-manager.sh` (~250 lines).

**Functions:**
- `generate_worktree_id()` - Creates unique ID: `wt-YYYYMMDD-HHMMSS-xxxx`
- `ensure_worktrees_dir()` - Creates `worktrees/` with `.gitignore`
- `is_worktree()` - Detects if in a worktree
- `cmd_create()` - Full creation flow
- `cmd_list()` - JSON listing with status
- `cmd_status()` - Detailed info
- `cmd_remove()` - Safe removal with change check
- `cmd_cleanup()` - Stale worktree cleanup
- `cmd_gc()` - Git worktree prune

### CEM Integration

CEM functions in `cem` script:
- `is_worktree()` - Detect worktree context
- `get_worktree_info()` - Read metadata
- `get_worktree_id()` - Extract ID from metadata
