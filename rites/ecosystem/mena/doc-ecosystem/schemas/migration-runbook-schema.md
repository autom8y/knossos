---
schema_name: migration-runbook
schema_version: "1.0"
file_pattern: ".ledge/reviews/MIGRATION-RUNBOOK-*.md"
artifact_type: migration-runbook
---

# Migration Runbook Schema

> Canonical schema for Migration Runbooks at `.ledge/reviews/MIGRATION-RUNBOOK-{slug}.md`

## YAML Frontmatter

```yaml
---
# Required fields
title: string              # Human-readable title
type: string               # Must be "migration-runbook"
created_at: string         # ISO 8601 timestamp
author: string             # Agent or user who created

# Version information
source_version: string     # Version migrating from (e.g., "1.0.0")
target_version: string     # Version migrating to (e.g., "2.0.0")
breaking_change: boolean   # Whether this is a breaking change

# Effort estimation
estimated_effort: string   # Time estimate (e.g., "30 minutes", "2 hours")
risk_level: enum           # low | medium | high | critical

# Prerequisites
prerequisites:             # What must be true before starting
  - description: string    # Prerequisite description
    verification: string   # How to verify it's met

# Migration steps
steps:                     # Ordered migration steps (at least one)
  - number: integer        # Step number
    action: string         # What to do
    verification: string   # How to verify success

# Rollback plan
rollback_steps:            # How to revert (at least one)
  - number: integer        # Step number
    action: string         # What to do

# Success verification
verification:              # How to confirm migration worked
  - description: string    # What to check
    expected: string       # Expected outcome
    command: string        # (optional) Command to run

# Traceability
context_design: string     # Reference to source Context Design

# Schema versioning
schema_version: "1.0"      # Must be "1.0" for this version
---
```

## Required Fields

| Field | Type | Description | Set By |
|-------|------|-------------|--------|
| `title` | string | Human-readable title | documentation-engineer |
| `type` | string | Must be "migration-runbook" | documentation-engineer |
| `created_at` | string | ISO 8601 creation timestamp | documentation-engineer |
| `author` | string | Creating agent or user | documentation-engineer |
| `source_version` | string | Version migrating from | documentation-engineer |
| `target_version` | string | Version migrating to | documentation-engineer |
| `breaking_change` | boolean | Is this breaking? | documentation-engineer |
| `estimated_effort` | string | Time estimate | documentation-engineer |
| `risk_level` | enum | Migration risk level | documentation-engineer |
| `prerequisites` | array | Pre-migration requirements | documentation-engineer |
| `steps` | array | Migration steps (min 1) | documentation-engineer |
| `rollback_steps` | array | Rollback steps (min 1) | documentation-engineer |
| `verification` | array | Success checks (min 1) | documentation-engineer |
| `context_design` | string | Source Context Design | documentation-engineer |
| `schema_version` | string | Schema version | documentation-engineer |

## Prerequisite Object Schema

```yaml
prerequisites:
  - description: string    # "Git working directory is clean"
    verification: string   # "Run: git status --porcelain returns empty"
    blocking: boolean      # (optional) If false, can proceed with warning
```

## Step Object Schema

```yaml
steps:
  - number: integer        # 1, 2, 3, etc.
    action: string         # Clear instruction
    verification: string   # How to verify step succeeded
    rollback_note: string  # (optional) How this step affects rollback
    warning: string        # (optional) Risk or consideration
    command: string        # (optional) Command to run
    expected_output: string  # (optional) What command should output
```

## Rollback Step Object Schema

```yaml
rollback_steps:
  - number: integer        # 1, 2, 3, etc.
    action: string         # Clear instruction
    verification: string   # (optional) How to verify rollback worked
    warning: string        # (optional) Risk or consideration
```

## Verification Object Schema

```yaml
verification:
  - description: string    # What to check
    expected: string       # Expected outcome
    command: string        # (optional) Command to verify
    manual: boolean        # (optional) Requires manual verification
```

## Validation Rules

### Structure Validation
1. File MUST start with `---` on line 1
2. File MUST have closing `---` within first 100 lines
3. Content between delimiters MUST be valid YAML

### Field Validation
1. `type` MUST be exactly "migration-runbook"
2. `created_at` MUST be valid ISO 8601 timestamp
3. `breaking_change` MUST be boolean
4. `risk_level` MUST be one of: low, medium, high, critical
5. `prerequisites` MUST be array (may be empty for simple migrations)
6. `steps` MUST be array with at least one item
7. `rollback_steps` MUST be array with at least one item
8. `verification` MUST be array with at least one item
9. `schema_version` MUST be "1.0"

### Step Validation
1. Each step MUST have `number`, `action`, `verification`
2. Steps MUST be numbered sequentially starting from 1
3. `action` MUST be clear, actionable instruction

### Rollback Validation
1. Each rollback step MUST have `number`, `action`
2. Rollback steps MUST be numbered sequentially
3. Rollback MUST be complete (can restore to source_version)

## Example: Valid Migration Runbook

```yaml
---
title: "Migration Runbook: Session Schema v1.0 to v2.0"
type: migration-runbook
created_at: "2025-12-29T14:00:00Z"
author: documentation-engineer
source_version: "1.0.0"
target_version: "2.0.0"
breaking_change: true
estimated_effort: "30 minutes"
risk_level: medium
prerequisites:
  - description: "Git working directory is clean"
    verification: "Run: git status --porcelain (should return empty)"
  - description: "No active sessions"
    verification: "Run: ari session status (should show no active session)"
  - description: "Backup of .claude/ directory"
    verification: "Confirm backup exists at known location"
steps:
  - number: 1
    action: "Pull latest knossos changes"
    verification: "git log shows latest knossos commit"
    command: "git pull origin main"
  - number: 2
    action: "Run ari sync to update hooks"
    verification: "No sync errors, hooks updated"
    command: "just ari-sync"
    expected_output: "Sync complete. 3 files updated."
  - number: 3
    action: "Migrate existing SESSION_CONTEXT.md files"
    verification: "All sessions have schema_version: 2.0"
    command: "find .sos/sessions -name 'SESSION_CONTEXT.md' -exec head -20 {} \\;"
    warning: "This modifies existing session files"
  - number: 4
    action: "Validate migrated sessions"
    verification: "All sessions pass validation"
    command: "just validate-sessions"
    expected_output: "All sessions valid."
rollback_steps:
  - number: 1
    action: "Restore .claude/ from backup"
    verification: "git status shows only expected changes"
  - number: 2
    action: "Reset to previous commit"
    verification: "git log shows previous commit as HEAD"
    command: "git reset --hard HEAD~1"
verification:
  - description: "Session validation passes"
    expected: "Exit code 0"
    command: "just validate-sessions"
  - description: "New sessions use v2.0 schema"
    expected: "schema_version: 2.0 in new sessions"
    command: "/start test-session && head -20 .sos/sessions/*/SESSION_CONTEXT.md"
  - description: "Hooks load without error"
    expected: "No bash errors on session start"
    manual: true
context_design: CONTEXT-DESIGN-session-schema-v2.md
schema_version: "1.0"
---

## Overview

This runbook guides satellite owners through migrating from session schema v1.0 to v2.0. The migration adds `schema_version` field and restructures artifact tracking.

## Pre-Migration Checklist

- [ ] Git working directory is clean
- [ ] No active sessions
- [ ] Backup of .claude/ directory created
- [ ] Stakeholders notified of brief downtime

## Migration Steps

### Step 1: Pull Latest Knossos Changes

```bash
git pull origin main
```

**Verification**: Check git log shows the session schema v2.0 commit.

### Step 2: Run Ari Sync

```bash
just ari-sync
```

**Verification**: Output shows hooks updated without errors.

### Step 3: Migrate Existing Sessions

The migration script will:
1. Add `schema_version: "2.0"` to all SESSION_CONTEXT.md files
2. Restructure `artifacts` array to new format
3. Preserve all existing data

```bash
just migrate-sessions
```

**Warning**: This modifies existing session files. Ensure backup exists.

### Step 4: Validate Migrated Sessions

```bash
just validate-sessions
```

**Expected**: "All sessions valid. 5 sessions checked."

## Rollback Procedure

If issues occur, rollback is safe:

### Step 1: Restore from Backup

```bash
cp -r /path/to/backup/.claude .claude
```

### Step 2: Reset Git State

```bash
git reset --hard HEAD~1
```

## Post-Migration Verification

| Check | Command | Expected |
|-------|---------|----------|
| Sessions valid | `just validate-sessions` | Exit 0 |
| New sessions | `/start test` | schema_version: 2.0 |
| Hooks load | (start session) | No errors |

## Troubleshooting

### "Missing required field" Error

**Cause**: Session file missing new required field.

**Fix**: Run migration script again, or manually add field.

### Validation Failure After Migration

**Cause**: Corrupt session file or incomplete migration.

**Fix**: Restore from backup and re-run migration.

## Support

If issues persist after troubleshooting:
1. Open issue in knossos repository
2. Include: error message, session file content, git log
```

## Validation Function

```bash
# In ecosystem-validator.sh
# Usage: validate_migration_runbook "/path/to/MIGRATION-RUNBOOK-example.md"
# Returns: 0=valid, 1=not found, 2=no opener, 3=no closer, 4=missing field, 5=invalid field

validate_migration_runbook() {
    local file="$1"
    local required_fields=("title" "type" "created_at" "author" "source_version" "target_version" "breaking_change" "estimated_effort" "risk_level" "prerequisites" "steps" "rollback_steps" "verification" "context_design" "schema_version")

    # Check file exists
    [ -f "$file" ] || { echo "File not found: $file" >&2; return 1; }

    # Check opening delimiter on line 1
    local first_line
    first_line=$(head -n 1 "$file")
    if [[ "$first_line" != "---" ]]; then
        echo "Invalid format: Missing opening '---' delimiter on line 1" >&2
        return 2
    fi

    # Check closing delimiter within first 100 lines
    local closing_line
    closing_line=$(head -n 100 "$file" | tail -n +2 | grep -n "^---$" | head -1 | cut -d: -f1)
    if [[ -z "$closing_line" ]]; then
        echo "Invalid format: Missing closing '---' delimiter within first 100 lines" >&2
        return 3
    fi

    # Extract frontmatter
    local frontmatter_end=$((closing_line + 1))
    local frontmatter
    frontmatter=$(sed -n "2,$((frontmatter_end))p" "$file" | sed '$d')

    # Check required fields
    local missing=()
    for field in "${required_fields[@]}"; do
        if ! echo "$frontmatter" | grep -q "^${field}:"; then
            missing+=("$field")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        echo "Missing required fields: ${missing[*]}" >&2
        return 4
    fi

    # Validate type is exactly "migration-runbook"
    local type
    type=$(echo "$frontmatter" | grep "^type:" | sed 's/type: *//' | tr -d '"')
    if [[ "$type" != "migration-runbook" ]]; then
        echo "Invalid type: Must be 'migration-runbook'" >&2
        return 5
    fi

    # Validate risk_level enum
    local risk_level
    risk_level=$(echo "$frontmatter" | grep "^risk_level:" | sed 's/risk_level: *//' | tr -d '"')
    if [[ ! "$risk_level" =~ ^(low|medium|high|critical)$ ]]; then
        echo "Invalid risk_level: Must be low, medium, high, or critical" >&2
        return 5
    fi

    return 0
}
```

## Handoff Criteria

When Migration Runbook phase completes, Pythia verifies:

- [ ] `type` is "migration-runbook"
- [ ] `context_design` references existing Context Design
- [ ] All prerequisites have verification steps
- [ ] Migration steps are numbered and actionable
- [ ] Rollback steps exist and are complete
- [ ] Verification checks confirm success
- [ ] Troubleshooting section addresses common issues

## Relationship to Other Artifacts

```
MIGRATION-RUNBOOK-{slug}.md
    |
    +-- References Context Design (context_design field)
    |
    +-- Consumed by satellite owners
    |
    +-- Verified by Compatibility Report (tested across satellites)
```
