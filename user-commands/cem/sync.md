---
description: "[DEPRECATED] Use roster-sync instead. Legacy CEM sync command."
argument-hint: [init|sync|status|diff] [--refresh] [--force] [--dry-run]
allowed-tools: Bash, Read
model: sonnet
---

## DEPRECATION NOTICE

This command is deprecated. Use `roster-sync` directly instead:

```bash
# From command line (preferred)
$ROSTER_HOME/roster-sync sync
$ROSTER_HOME/roster-sync init
$ROSTER_HOME/roster-sync status
$ROSTER_HOME/roster-sync diff

# With options
$ROSTER_HOME/roster-sync sync --refresh   # Sync and refresh team
$ROSTER_HOME/roster-sync sync --force     # Force overwrite local changes
$ROSTER_HOME/roster-sync sync --dry-run   # Preview changes
```

## Migration Guide

| Old (CEM) | New (roster-sync) |
|-----------|-------------------|
| `cem sync` | `$ROSTER_HOME/roster-sync sync` |
| `cem init` | `$ROSTER_HOME/roster-sync init` |
| `cem status` | `$ROSTER_HOME/roster-sync status` |
| `cem diff` | `$ROSTER_HOME/roster-sync diff` |

## Legacy Behavior (Forwarding)

If invoked, this command forwards to roster-sync:

```bash
$ROSTER_HOME/roster-sync $ARGUMENTS
```

## Reference

Full roster-sync documentation: Run `$ROSTER_HOME/roster-sync --help`
