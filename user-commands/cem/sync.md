---
description: Sync project with roster ecosystem
argument-hint: [sync|init|status|diff|validate|repair] [--refresh] [--force] [--dry-run]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Execute roster-sync to synchronize project with roster ecosystem. $ARGUMENTS

## Behavior

1. **Execute roster-sync** using standard path resolution:
   ```bash
   ${KNOSSOS_HOME:-~/Code/roster}/roster-sync [command] $ARGUMENTS
   ```
   This expands to `$KNOSSOS_HOME/roster-sync` if set, otherwise `~/Code/roster/roster-sync`

2. **Pass through all arguments**:
   - Command: sync, init, status, diff, validate, repair
   - Flags: --refresh, --force, --dry-run, --prune, --auto-refresh, etc.
   - Display output directly to user

3. **Handle errors**:
   - If roster-sync not found at `~/Code/roster/roster-sync`:
     - ERROR: "roster-sync not found. Expected location: ~/Code/roster/roster-sync"
     - Suggest: "Clone roster repository to ~/Code/roster or set KNOSSOS_HOME"
   - If execution fails: Display stderr for debugging

## Common Commands

```bash
/sync sync              # Pull updates from roster
/sync sync --refresh   # Sync and refresh active rite
/sync status           # Show sync status and version
/sync diff             # Show pending changes
/sync init             # Initialize new project
/sync validate         # Check manifest integrity
```

## Reference

Full documentation: `.claude/skills/ecosystem-ref/skill.md`
