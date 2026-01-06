---
title: "Migration Runbook: Artifact Validation Features (swap-team.sh v1.3)"
type: migration-runbook
created_at: "2026-01-03T20:00:00Z"
author: documentation-engineer
source_version: "1.2.0"
target_version: "1.3.0"
breaking_change: false
estimated_effort: "15 minutes"
risk_level: low
prerequisites:
  - description: "swap-team.sh v1.3.0 with validation features available"
    verification: "Run: grep -q 'validate_workflow_yaml' swap-team.sh (returns 0)"
  - description: "Git working directory is clean"
    verification: "Run: git status --porcelain (should return empty)"
  - description: "No swap operation in progress"
    verification: "Run: ls .claude/.swap-journal 2>/dev/null (should not exist)"
steps:
  - number: 1
    action: "Pull latest roster changes"
    verification: "swap-team.sh contains validation functions"
    command: "cd ~/Code/roster && git pull"
    expected_output: "Already up to date or merge successful"
  - number: 2
    action: "Validate team pack schemas before using --update"
    verification: "No validation errors in output"
    command: "./swap-team.sh <team-name> --dry-run"
    expected_output: "Dry-run preview shows no validation errors"
  - number: 3
    action: "Run swap with new validation features active"
    verification: "Swap completes successfully with validation messages"
    command: "./swap-team.sh <team-name> --update"
    expected_output: "Switched to <team-name> (N agents loaded)"
  - number: 4
    action: "Clean up orphan backups if desired"
    verification: "Old backups removed, last 3 retained"
    command: "./swap-team.sh --cleanup-orphans"
    expected_output: "Cleaned up N old orphan backup(s)"
rollback_steps:
  - number: 1
    action: "Restore from previous roster commit"
    verification: "swap-team.sh no longer has validation functions"
    command: "cd ~/Code/roster && git checkout HEAD~1 -- swap-team.sh"
  - number: 2
    action: "Re-run swap without validation"
    verification: "Swap completes (may encounter schema issues)"
verification:
  - description: "Schema validation runs during swap"
    expected: "Logs show 'Validating team schemas' or validation passes silently"
    command: "ROSTER_DEBUG=1 ./swap-team.sh <team-name> --dry-run 2>&1 | grep -i validat"
  - description: "Command collision detection active"
    expected: "Warnings appear if user commands override team commands"
    command: "Create test command in ~/.claude/commands/, then run swap"
  - description: "Orphan backup cleanup available"
    expected: "--cleanup-orphans and --auto-cleanup flags accepted"
    command: "./swap-team.sh --help | grep cleanup"
context_design: N/A
schema_version: "1.0"
---

## Overview

This runbook documents the new artifact validation features added to `swap-team.sh` in v1.3.0. These features improve swap reliability by:

1. **WP1: Command Collision Detection** - Warns when user commands would override team commands
2. **WP2: Schema Validation Pre-Swap** - Validates workflow.yaml and orchestrator.yaml before applying
3. **WP3: Orphan Backup Cleanup** - Automatic management of old orphan backup directories

These are **non-breaking changes** - existing swap operations continue to work, with added validation that catches schema issues early.

---

## What Changed

### WP1: Command Collision Detection

**Location**: `swap-team.sh` lines 2189-2237, integrated at line 4107

**Behavior**:
- Checks for naming conflicts between team commands and user commands (in `~/.claude/commands/`)
- Warns about collisions but does not block the swap
- User commands always take precedence (team commands with same name are skipped)

**New Output Example**:
```
[Roster] Warning: Command collision(s) detected: 2 command(s)
[Roster] Warning: User commands (in ~/.claude/commands/) will take precedence:
[Roster] Warning:   - review.md
[Roster] Warning:   - deploy.md
[Roster] Warning: Team commands with same names will be skipped during sync
```

### WP2: Schema Validation Pre-Swap

**Location**: `swap-team.sh` lines 1763-1875, integrated at lines 3983-3987

**Behavior**:
- Validates `workflow.yaml` schema (if present): requires `name`, `workflow_type`, `phases` fields
- Validates `orchestrator.yaml` schema (if present): requires `team`, `team.name`, `routing` fields
- **Hard fail**: If validation fails, swap is aborted with exit code 2 (`EXIT_VALIDATION_FAILURE`)

**Validated Fields**:

| File | Required Fields | Validation |
|------|-----------------|------------|
| `workflow.yaml` | `name`, `workflow_type`, `phases` | `phases` must be non-empty list |
| `orchestrator.yaml` | `team`, `team.name`, `routing` | Nested structure validated |

### WP3: Orphan Backup Cleanup

**Location**: `swap-team.sh` lines 2895-2962

**New Flags**:
- `--cleanup-orphans`: Manual cleanup of old orphan backup directories
- `--auto-cleanup`: Automatic cleanup during swap operations

**Behavior**:
- Scans `.claude/{agents,commands,skills,hooks}.orphan-backup/` directories
- Keeps the last 3 backups per type (sorted by modification time)
- Removes older backups to prevent disk space accumulation

---

## Operator Guide

### Running Swap with New Validation

Standard swap operation now includes schema validation automatically:

```bash
# Swap to a team (validation runs automatically)
./swap-team.sh dev-pack

# Preview swap without applying (includes validation check)
./swap-team.sh dev-pack --dry-run

# Force re-apply current team with validation
./swap-team.sh dev-pack --update
```

### Interpreting Collision Warnings

When you see collision warnings:

```
[Roster] Warning: Command collision(s) detected: 1 command(s)
[Roster] Warning: User commands (in ~/.claude/commands/) will take precedence:
[Roster] Warning:   - commit.md
```

**What this means**:
- You have a file `~/.claude/commands/commit.md`
- The team pack also has a command `commit.md`
- Your user command will be used; the team command will be skipped

**What to do**:
- If intentional: No action needed, your customization is preserved
- If unintentional: Remove or rename your user command to use team version

### Handling Validation Failures

If swap fails with validation error:

```
[Roster] Error: workflow.yaml missing required field: phases
[Roster] Error: Schema validation failed for workflow.yaml
[Roster] Error: Team pack dev-pack has invalid configuration
[Roster] Error: Team schema validation failed, aborting swap
```

**What this means**:
- The team pack has an invalid `workflow.yaml` or `orchestrator.yaml`
- Swap was aborted to prevent loading broken configuration

**Resolution**: See Troubleshooting section below.

### Managing Orphan Backups

**Manual cleanup** (run independently):
```bash
# Clean up old orphan backups (keeps last 3 per type)
./swap-team.sh --cleanup-orphans
```

**Automatic cleanup during swap**:
```bash
# Swap and clean orphan backups in one operation
./swap-team.sh dev-pack --auto-cleanup
```

**Check current backup usage**:
```bash
# List orphan backup directories and sizes
du -sh .claude/*.orphan-backup/* 2>/dev/null | sort -h
```

---

## Troubleshooting

### Decision Tree: Validation Failure

```
Validation error during swap?
|
+-> workflow.yaml error?
|   |
|   +-> "missing required field: name"
|   |   -> Add "name: <workflow-name>" at top level
|   |
|   +-> "missing required field: workflow_type"
|   |   -> Add "workflow_type: phased" or "workflow_type: linear"
|   |
|   +-> "missing required field: phases"
|   |   -> Add "phases:" section with at least one phase
|   |
|   +-> "phases must be a non-empty list"
|       -> Ensure phases section has "- name:" entries
|
+-> orchestrator.yaml error?
    |
    +-> "missing required field: team"
    |   -> Add "team:" section at top level
    |
    +-> "missing required field: team.name"
    |   -> Add "name:" under "team:" section
    |
    +-> "missing required field: routing"
        -> Add "routing:" section at top level
```

### Common Validation Failure Messages

#### workflow.yaml Errors

**Error**: `workflow.yaml missing required field: name`
```yaml
# Fix: Add name field
name: my-workflow
workflow_type: phased
phases:
  - name: analyze
    agent: analyst
```

**Error**: `workflow.yaml phases must be a non-empty list`
```yaml
# Wrong:
phases:

# Correct:
phases:
  - name: phase-one
    agent: my-agent
```

#### orchestrator.yaml Errors

**Error**: `orchestrator.yaml missing required field: team.name`
```yaml
# Wrong:
team:
  description: "My team"

# Correct:
team:
  name: my-team
  description: "My team"
routing:
  default: orchestrator
```

### Resolving Command Collisions

**Scenario**: Team command overridden by user command

1. **Keep user command** (no action needed - this is the default)

2. **Use team command instead**:
   ```bash
   # Remove or rename user command
   mv ~/.claude/commands/review.md ~/.claude/commands/my-review.md

   # Re-run swap
   ./swap-team.sh dev-pack --update
   ```

3. **Merge commands**: Copy relevant parts from team command into user command

### Orphan Backup Issues

**Error**: Disk space filling up with orphan backups
```bash
# Check backup sizes
du -sh .claude/*.orphan-backup/*

# Clean up old backups
./swap-team.sh --cleanup-orphans
```

**Error**: Cannot find backed-up agents
```bash
# Backups are in timestamped directories
ls -la .claude/agents.orphan-backup/

# Format: {timestamp}-{team-name}/
# Example: 20260103-143022-old-pack/
```

---

## Validation Checklist

### Pre-Swap Checklist

- [ ] No `.claude/.swap-journal` file exists (no interrupted swap)
- [ ] Git working directory is clean (or changes are intentional)
- [ ] Team pack exists in roster: `ls ~/Code/roster/teams/<team-name>`
- [ ] If team has workflow.yaml: `name`, `workflow_type`, `phases` fields present
- [ ] If team has orchestrator.yaml: `team`, `team.name`, `routing` fields present

### Post-Swap Verification

```bash
# 1. Check swap completed
cat .claude/ACTIVE_RITE
# Expected: <team-name>

# 2. Verify agents loaded
ls .claude/agents/*.md | wc -l
# Expected: Number of agents in team pack

# 3. Check for collision warnings in output
# Review any "Command collision(s) detected" warnings

# 4. Validate manifest updated
jq '.team.name' .claude/AGENT_MANIFEST.json
# Expected: "<team-name>"

# 5. Check orphan backup status (if --auto-cleanup used)
ls .claude/*.orphan-backup/ 2>/dev/null | wc -l
# Expected: 0-3 directories per type
```

### Health Check Commands

```bash
# Full validation dry-run
ROSTER_DEBUG=1 ./swap-team.sh <team-name> --dry-run 2>&1

# Check for schema issues before swap
grep -l "^name:" ~/Code/roster/teams/<team-name>/workflow.yaml
grep -l "^team:" ~/Code/roster/teams/<team-name>/orchestrator.yaml

# Verify current state consistency
./swap-team.sh --verify

# Check for orphan backups
du -sh .claude/*.orphan-backup/* 2>/dev/null
```

---

## Rollback Procedures

### Rollback: Swap Completed but Issues Found

If swap completed but team configuration causes problems:

```bash
# Option 1: Swap to different team
./swap-team.sh <previous-team>

# Option 2: Reset to no team (skeleton baseline)
./swap-team.sh --reset

# Option 3: Recover from backup
ls .claude/.swap-backup/  # Check if backup exists
./swap-team.sh --recover
```

### Rollback: Validation Blocked Legitimate Swap

If validation is incorrectly blocking a valid team pack:

```bash
# 1. Verify the team pack schema is correct
cat ~/Code/roster/teams/<team-name>/workflow.yaml

# 2. If schema is actually valid, file a bug
# The validation may have a false positive

# 3. Temporary workaround: fix schema in team pack
# Edit workflow.yaml or orchestrator.yaml to add missing fields
```

### Rollback: Orphan Cleanup Removed Needed Files

Orphan backups retain the last 3 versions:

```bash
# 1. Check remaining backups
ls -la .claude/agents.orphan-backup/

# 2. Restore from backup
cp -r .claude/agents.orphan-backup/<timestamp>-<team>/* .claude/agents/
```

---

## Exit Codes Reference

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | `EXIT_SUCCESS` | Operation completed successfully |
| 1 | `EXIT_INVALID_ARGS` | Invalid command-line arguments |
| 2 | `EXIT_VALIDATION_FAILURE` | Schema validation failed (new in v1.3) |
| 3 | `EXIT_BACKUP_FAILURE` | Failed to create backup |
| 4 | `EXIT_SWAP_FAILURE` | Swap operation failed |
| 5 | `EXIT_ORPHAN_CONFLICT` | Orphan handling conflict |
| 6 | `EXIT_RECOVERY_REQUIRED` | Manual recovery needed |

---

## Compatibility Matrix

| swap-team.sh | workflow.yaml | orchestrator.yaml | Behavior |
|--------------|---------------|-------------------|----------|
| < 1.3.0 | Any | Any | No validation, may load invalid configs |
| >= 1.3.0 | Valid | Valid | Swap proceeds normally |
| >= 1.3.0 | Invalid | Any | Swap blocked (exit 2) |
| >= 1.3.0 | Missing | Any | Swap proceeds (optional file) |
| >= 1.3.0 | Any | Invalid | Swap blocked (exit 2) |
| >= 1.3.0 | Any | Missing | Swap proceeds (optional file) |

---

## Support

If issues persist after troubleshooting:

1. Enable debug mode: `ROSTER_DEBUG=1 ./swap-team.sh <team> 2>&1 | tee swap-debug.log`
2. Check journal for interrupted swap: `cat .claude/.swap-journal`
3. File issue in roster repository with:
   - Error message
   - Debug log output
   - Team pack name and workflow.yaml/orchestrator.yaml contents
