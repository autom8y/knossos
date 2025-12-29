---
description: Sync project with skeleton_claude ecosystem
argument-hint: [init|sync|status|diff|install-user]
allowed-tools: Bash, Read
---

## Context

Auto-injected by SessionStart hook.

## Your Task

$ARGUMENTS

Manage ecosystem synchronization with skeleton_claude using the CEM (Claude Ecosystem Manager) tool.

## Skeleton Detection

First, check if we're in the skeleton project itself:
```bash
[[ -d ".claude/user-agents" ]] && echo "IN_SKELETON" || echo "IN_SATELLITE"
```

**If IN_SKELETON**: Use `install-user` behavior (push to user-level)
**If IN_SATELLITE**: Use normal sync behavior (pull from skeleton)

## Behavior

### For Skeleton Project (has .claude/user-agents/)

**If no arguments or `sync`:**
Run `~/Code/skeleton_claude/cem install-user` to push updates to ~/.claude/

**If `--force`:**
Add `--force` flag to overwrite existing user resources.

### For Satellite Projects

**If no arguments or `sync`:**
Run `~/Code/skeleton_claude/cem sync` to pull updates from skeleton.

**If `--refresh` (waterfall sync):**
Run `~/Code/skeleton_claude/cem sync --refresh` to:
1. Sync CEM infrastructure from skeleton
2. If ACTIVE_TEAM exists, refresh team agents from roster

This is the recommended sync command when you want to pull all updates.

**If `init`:**
Run `~/Code/skeleton_claude/cem init` to initialize this project with the ecosystem.

**If `status`:**
Run `~/Code/skeleton_claude/cem status` to show current sync state.

**If `diff`:**
Run `~/Code/skeleton_claude/cem diff` to show differences with skeleton.

**If `--force`:**
Add `--force` flag to overwrite local modifications.

**If `--dry-run`:**
Add `--dry-run` flag to preview changes without applying.

### Explicit Commands (Either Context)

**If `install-user`:**
Run `~/Code/skeleton_claude/cem install-user` to install user-level resources.

## Examples

```bash
# In satellite projects:
/sync              # Pull latest from skeleton (infrastructure only)
/sync --refresh    # Sync infrastructure AND refresh team agents (recommended)
/sync init         # Initialize project
/sync status       # Show sync state
/sync --force      # Force overwrite local changes
/sync --dry-run    # Preview changes without applying

# In skeleton project:
/sync              # Push user resources to ~/.claude/
/sync --force      # Force overwrite user resources

# Anywhere:
/sync install-user # Explicitly install user resources
```

## After Running

Report the output to the user. If there are conflicts, explain what happened and how to resolve them.

## Reference

Full CEM documentation: Run `~/Code/skeleton_claude/cem --help`
