# Worktree Troubleshooting

> Common issues and solutions for git worktree management.

## Common Issues

### "Failed to create git worktree"

**Cause:** Git worktree add failed.

**Check:**
- Is this a git repository?
- Does the target path already exist?
- Is the from_ref valid?

**Fix:**
```bash
# Verify git repository
git status

# Check if path exists
ls worktrees/

# Verify ref exists
git rev-parse HEAD
```

### "Failed to initialize ecosystem in worktree"

**Cause:** Roster sync scripts not found or failed.

**Check:**
- Is `$ROSTER_HOME` set correctly?
- Do the sync scripts exist and are executable?

**Fix:**
```bash
# Verify roster exists
ls -la $ROSTER_HOME/sync-user-agents.sh

# Make executable if needed
chmod +x $ROSTER_HOME/sync-*.sh
```

### "Worktree has uncommitted changes"

**Cause:** Trying to remove worktree with uncommitted changes.

**Fix:** Either commit/stash changes or use `--force`:
```bash
# Option 1: Commit changes
cd worktrees/wt-{id}
git add . && git commit -m "WIP: worktree work"

# Option 2: Force remove
/worktree remove wt-{id} --force
```

### Cannot find worktree after creation

**Cause:** Worktree created but user didn't `cd` into it.

**Fix:** Follow the instructions in the output:
```bash
cd worktrees/wt-{id} && claude
```

### Worktree shows stale session

**Cause:** Session was not properly closed before leaving worktree.

**Fix:**
```bash
# Navigate to worktree
cd worktrees/wt-{id}

# Check session status
.claude/hooks/lib/session-manager.sh status

# Wrap or park the session
/wrap
# or
/park
```

### Git worktree prune warnings

**Cause:** Orphaned worktree references in git.

**Fix:**
```bash
/worktree gc
# or manually
git worktree prune
```

### Team not active in worktree

**Cause:** Team swap failed during creation or was never set.

**Fix:**
```bash
cd worktrees/wt-{id}
./swap-team.sh <pack-name>
```

## Diagnostic Commands

```bash
# List all git worktrees
git worktree list

# Check worktree metadata
cat worktrees/wt-{id}/.claude/.worktree-meta.json

# Verify CEM state
cd worktrees/wt-{id} && cem status

# Check session state
cd worktrees/wt-{id} && .claude/hooks/lib/session-manager.sh status
```
