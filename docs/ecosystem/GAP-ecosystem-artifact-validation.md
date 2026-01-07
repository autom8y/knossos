---
title: "Gap Analysis: Ecosystem Artifact Validation"
type: gap-analysis
complexity: MODULE
created_at: "2026-01-03T20:00:00Z"
status: ready-for-design
affected_systems:
  - roster
author: ecosystem-analyst
issue_count: 4
critical_count: 0
high_count: 2
medium_count: 2
root_cause: "Additive-only sync design lacks validation gates and collision detection across artifact pillars"
success_criteria:
  - id: GAP-SC-001
    description: "Command collision detection prevents user/team command shadowing"
    artifact_type: implementation
  - id: GAP-SC-002
    description: "workflow.yaml validated against schema before team swap completes"
    artifact_type: implementation
  - id: GAP-SC-003
    description: "Orphan backup cleanup policy prevents accumulation"
    artifact_type: context-design
  - id: GAP-SC-004
    description: "Security routing triggers enforced, not advisory-only"
    artifact_type: context-design
schema_version: "1.0"
---

## Executive Summary

Deep audit of roster's five artifact pillars (agents, skills, hooks, commands, workflows/schemas) reveals 4 validation gaps that enable silent failures, resource accumulation, and bypassed security routing. All issues trace to historical additive-only sync design that deferred validation to execution time.

## Root Cause Chain

1. sync-user-*.sh scripts designed for additive-only operation (never remove)
2. swap-rite.sh mirrors this pattern for team-level artifacts
3. No manifest or registry exists for command collision detection
4. Schema validation exists but is not invoked pre-swap
5. Orphan backups created but no cleanup policy defined
6. Security consultation triggers are advisory (workflow.yaml policy.recommended) not enforced

## Issue Inventory

### Category 1: Collision Detection (1 Issue)

#### COL-001: Command Collision Detection Missing (HIGH)

**Description**: When swapping teams, team commands can silently shadow user-level commands of the same name. No COMMAND_MANIFEST.json or collision check exists in swap-rite.sh.

**Evidence**:
- File: `/Users/tomtenuta/Code/roster/swap-rite.sh`
- Lines 2117-2122: Collision check only protects project commands, not user commands

```bash
# Check for collision with existing project command
if [[ -f ".claude/commands/$cmd_name" ]] && ! grep -q "^$cmd_name$" "$marker_file" 2>/dev/null; then
    # Not a team command, this is a project command - skip with warning
    log_warning "Skipped: $cmd_name (project command exists)"
    continue
fi
```

- Overlapping command names exist:
  - Team commands: pr, spike (from 10x-dev)
  - User commands: pr, spike (from user-commands/operations/)

**Impact**: Team command could silently replace user command behavior during swap. User expects `/pr` to invoke their customized version but gets team version instead.

**Root Cause**: Historical additive-only sync design did not anticipate user/team command layering.

---

### Category 2: Schema Validation (1 Issue)

#### SCH-001: Schema Validation Not Enforced Pre-Swap (HIGH)

**Description**: swap-rite.sh copies workflow.yaml without validating it against the schema at `.claude/schemas/workflow.schema.json`. Invalid workflow files are loaded and fail at runtime.

**Evidence**:
- File: `/Users/tomtenuta/Code/roster/swap-rite.sh`
- Lines 1915-1919: Copy without validation

```bash
# Copy workflow.yaml if exists
local workflow_file="$ROSTER_HOME/rites/$rite_name/workflow.yaml"
    log_debug "Copying workflow.yaml"
    cp "$workflow_file" .claude/ACTIVE_WORKFLOW.yaml || {
        log_warning "Failed to copy workflow.yaml (agents swapped successfully)"
```

- Schema exists at: `/Users/tomtenuta/Code/roster/.claude/schemas/workflow.schema.json`
- Line 2779 validates hook schema version but not workflow.yaml structure

**Impact**: Invalid workflow.yaml (missing required fields, invalid phase names) loads successfully but fails when agents attempt phase transitions at runtime.

**Root Cause**: Validation deferred to execution time. Schema exists but no pre-swap validation gate.

---

### Category 3: Resource Management (1 Issue)

#### ORP-001: Orphan Skill Backups Accumulating (MEDIUM)

**Description**: When skills are orphaned during team swaps, they are backed up to `.claude/skills.orphan-backup/` but never cleaned up. Backups accumulate indefinitely.

**Evidence**:
- File: `/Users/tomtenuta/Code/roster/swap-rite.sh`
- Lines 2231, 2258-2259:

```bash
local backup_dir=".claude/skills.orphan-backup"
...
if [[ "$ORPHAN_MODE" == "remove" ]] && [[ -d "$backup_dir" ]]; then
    log "Orphan skill backups saved to: $backup_dir"
```

- Current orphan backups in roster:
  - `.claude/skills.orphan-backup/` contains 27+ skill directories
  - Also: `.claude/commands.orphan-backup/`, `.claude/hooks.orphan-backup/`
  - No cleanup policy or retention limit defined

**Impact**: Storage waste, audit confusion when determining which skills are active, stale backups may be mistaken for current configuration.

**Root Cause**: Backup-only design without corresponding cleanup policy. Backups preserve safety but no lifecycle management.

---

### Category 4: Security Routing (1 Issue)

#### SEC-001: Security Routing Not Automated (MEDIUM)

**Description**: 10x-dev defines 9 security patterns in workflow.yaml security_consultation section, but routing is advisory-only. Agents are not required to invoke security consultation.

**Evidence**:
- File: `/Users/tomtenuta/Code/roster/rites/10x-dev/workflow.yaml`
- Lines 75-127: security_consultation section defines triggers and policy

```yaml
security_consultation:
  triggers:
    - pattern: "auth|authentication|authorization"
      domain: authentication
      severity: high
    - pattern: "crypto|encryption|hashing|bcrypt|argon"
      domain: cryptography
      severity: critical
    # ... 7 more patterns
  policy:
    required:
      - complexity: SYSTEM
        domains: [authentication, cryptography, data_privacy, financial]
    recommended:
      - complexity: MODULE
        domains: [authentication, data_privacy]
```

- No enforcement mechanism exists - policy is declarative only
- Architect agent receives invocation pattern but no hook verifies consultation occurred

**Impact**: Security-sensitive changes (authentication, cryptography, financial) may bypass security review. "Required" policy is not enforced, only documented.

**Root Cause**: Advisory-only consultation triggers. Workflow.yaml declares policy but no pre-commit or phase-transition hook enforces it.

---

## Prioritized Fix List

| Priority | ID | Issue | Effort | Impact |
|----------|-----|-------|--------|--------|
| 1 | COL-001 | Command collision detection | Medium | Prevents silent shadowing |
| 2 | SCH-001 | Schema validation pre-swap | Low | Catches invalid workflows early |
| 3 | SEC-001 | Security routing enforcement | High | Ensures security review compliance |
| 4 | ORP-001 | Orphan backup cleanup | Low | Reduces storage/confusion |

## Success Criteria

- [ ] GAP-SC-001: Command collision detection prevents user/team command shadowing
  - swap-rite.sh checks user-level commands at ~/.claude/commands/ before copying team commands
  - Warning or error raised when collision detected
  - Manifest tracks command provenance (user vs. team vs. project)

- [ ] GAP-SC-002: workflow.yaml validated against schema before team swap completes
  - swap-rite.sh invokes schema validation after staging, before commit
  - Invalid workflow.yaml causes swap to abort with clear error
  - Validation uses existing `.claude/schemas/workflow.schema.json`

- [ ] GAP-SC-003: Orphan backup cleanup policy prevents accumulation
  - Configurable retention (e.g., keep last 3 backups per type)
  - `--cleanup-orphans` flag to manually prune old backups
  - Optional auto-cleanup after successful swap

- [ ] GAP-SC-004: Security routing triggers enforced, not advisory-only
  - PreToolUse or phase transition hook checks security_consultation policy
  - SYSTEM+cryptography requires proof of security consultation
  - Enforcement configurable (strict vs. warn-only)

## Affected Components

| Component | Files | Change Type |
|-----------|-------|-------------|
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh:2100-2136` | Add collision detection |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh:1915-1919` | Add schema validation |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh:2231-2260` | Add cleanup policy |
| workflow.yaml | All team workflow files | Document enforcement gates |
| hooks | New security-enforcement hook | Enforce consultation policy |

## Test Satellites

For verification, test fixes against:

1. **roster** (baseline) - Main repository with full ecosystem
2. **test-satellite-minimal** - No custom commands/skills, clean swap
3. **test-satellite-complex** - Custom user commands overlapping team commands

## Recommendations

### Immediate (This Sprint)

1. **COL-001**: Add user-command collision detection to swap-rite.sh sync_team_commands()
2. **SCH-001**: Add workflow.yaml schema validation call before commit phase

### Short-term (Next Sprint)

3. **ORP-001**: Implement backup retention policy (e.g., 3 most recent)
4. Document cleanup command in /team help

### Medium-term (Design Required)

5. **SEC-001**: Design security enforcement hook with context-architect
   - Needs: Definition of "consultation proof" (artifact path? session log?)
   - Needs: Graceful degradation for teams without security

## Handoff

This gap analysis is ready for context-architect to produce:
1. Context design for command collision registry
2. Context design for orphan cleanup policy
3. Context design for security consultation enforcement gates

Route to: `context-architect` with reference to this document.
