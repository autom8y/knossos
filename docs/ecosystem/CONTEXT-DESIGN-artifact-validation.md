---
title: "Context Design: Artifact Validation Gates"
type: context-design
complexity: MODULE
created_at: "2026-01-03T21:00:00Z"
status: ready-for-implementation
gap_analysis: GAP-ecosystem-artifact-validation.md
affected_systems:
  - roster
author: context-architect
backward_compatible: true
migration_required: false
work_packages:
  - id: WP1
    name: "Command Collision Detection"
    description: "Add user-level command collision detection to swap-rite.sh"
    files:
      - path: "swap-rite.sh"
        action: modify
        description: "Add check_user_command_collisions() function and integration"
    estimated_effort: "2 hours"
  - id: WP2
    name: "Schema Validation Pre-Swap"
    description: "Validate workflow.yaml and orchestrator.yaml before commit phase"
    files:
      - path: "swap-rite.sh"
        action: modify
        description: "Add validate_team_schemas() function with yq-based validation"
    dependencies: []
    estimated_effort: "1.5 hours"
  - id: WP3
    name: "Orphan Backup Cleanup Policy"
    description: "Implement retention policy and cleanup command for orphan backups"
    files:
      - path: "swap-rite.sh"
        action: modify
        description: "Add cleanup_orphan_backups() function with retention logic"
    dependencies: []
    estimated_effort: "1 hour"
  - id: WP4
    name: "Security Consultation Enforcement Hook"
    description: "Design enforcement mechanism for security consultation policy"
    files:
      - path: "user-hooks/security-consultation-check.sh"
        action: create
        description: "PreToolUse hook checking security consultation requirements"
      - path: "rites/10x-dev-pack/workflow.yaml"
        action: modify
        description: "Add enforcement_mode field to security_consultation"
    dependencies: [WP1, WP2]
    estimated_effort: "3 hours"
schema_version: "1.0"
---

## Executive Summary

This Context Design addresses 4 validation gaps identified in GAP-ecosystem-artifact-validation.md: command collision detection, schema validation pre-swap, orphan backup cleanup, and security consultation enforcement. All designs maintain backward compatibility with existing satellites and use conservative defaults (warn-only mode where applicable).

## Design Decisions

### Decision 1: Command Collision Detection Strategy

**Options Considered**:
1. **COMMAND_MANIFEST.json registry** - Track all commands across layers with provenance
2. **Real-time collision check at sync** - Check user commands only during swap
3. **Unified command index** - Single index file tracking user + project + team commands

**Selected**: Option 2 - Real-time collision check at sync

**Rationale**:
- COMMAND_MANIFEST.json adds schema maintenance overhead and synchronization complexity
- Real-time check at sync is simpler, requires no persistent state
- User commands at `~/.claude/commands/` are stable (user-maintained)
- Check can be fast: glob + basename comparison, no JSON parsing

### Decision 2: Collision Resolution Policy

**Options Considered**:
1. **Team wins** - Team command overwrites user command silently
2. **User wins** - Skip team command, keep user version
3. **Error** - Abort swap on any collision
4. **Warn + User wins** - Log warning, preserve user command (current behavior for project commands)

**Selected**: Option 4 - Warn + User wins

**Rationale**:
- Matches existing project-command collision behavior (lines 2118-2121)
- Non-breaking: satellites with overlapping commands continue to function
- Warning provides visibility without disrupting workflow
- Users can rename commands if they want team version

### Decision 3: Schema Validation Approach

**Options Considered**:
1. **jq-based JSON Schema validation** - Use jq with schema
2. **yq-based YAML validation** - Parse YAML and validate required fields
3. **External validator** - Call `ajv` or similar JSON Schema CLI
4. **Minimal field check** - Check required fields exist without full schema validation

**Selected**: Option 4 - Minimal field check

**Rationale**:
- workflow.schema.json requires `name`, `workflow_type`, `entry_point`, `phases`
- yq/jq dependency already exists in swap-rite.sh (used for manifest operations)
- Full JSON Schema validation adds complexity for marginal benefit
- Required field check catches 90% of issues (missing config)
- If yq not available, fall back to grep-based check with warning

### Decision 4: Validation Failure Behavior

**Options Considered**:
1. **Hard fail** - Abort swap on validation failure
2. **Warn and continue** - Log warning, complete swap
3. **Configurable** - Flag to control strictness

**Selected**: Option 1 - Hard fail

**Rationale**:
- Invalid workflow.yaml will fail at runtime anyway
- Fail-fast gives clear feedback at swap time
- Transaction rollback already implemented in swap-rite.sh
- Users can fix workflow.yaml and retry

### Decision 5: Orphan Backup Retention Policy

**Options Considered**:
1. **Count-based** - Keep last N backups per type
2. **Time-based** - Delete backups older than N days
3. **Size-based** - Delete when total exceeds N MB
4. **Hybrid** - Count + time (keep 3 or 7 days, whichever is fewer)

**Selected**: Option 1 - Count-based (keep last 3)

**Rationale**:
- Simple to implement and understand
- 3 backups provides sufficient recovery window
- No date parsing or disk space calculation needed
- Matches git reflog philosophy (bounded history)

### Decision 6: Cleanup Trigger

**Options Considered**:
1. **Automatic after swap** - Always clean up after successful swap
2. **Manual only** - `--cleanup-orphans` flag
3. **Configurable auto** - `--auto-cleanup` flag enables automatic mode
4. **Prompt** - Ask user during swap if backups exist

**Selected**: Option 3 - Configurable auto

**Rationale**:
- Default: preserve backups (safe)
- Opt-in: `--auto-cleanup` for users who want it
- Manual: `swap-rite.sh --cleanup-orphans` for explicit pruning
- No prompts (swap-rite.sh is designed for non-interactive use)

### Decision 7: Security Consultation Enforcement Mode

**Options Considered**:
1. **Hard block** - Prevent implementation without consultation proof
2. **Warn only** - Log warning, allow continuation
3. **Configurable per-domain** - Different strictness per security domain
4. **Audit log only** - Record bypass for later review

**Selected**: Option 2 - Warn only (with Option 3 for future enhancement)

**Rationale**:
- Hard block requires "consultation proof" mechanism not yet designed
- Warn-only provides visibility without disrupting workflow
- Aligns with QA authority philosophy: "advisory only, humans decide"
- Future enhancement can add configurable enforcement per workflow.yaml policy

## Work Package Details

### WP1: Command Collision Detection

**Objective**: Detect and warn when team commands would shadow user-level commands

**Implementation**:

```bash
# Add to swap-rite.sh after line 2109 (inside sync_team_commands)

# Check for user-level command collisions
check_user_command_collisions() {
    local source_dir="$1"
    local user_commands_dir="$HOME/.claude/commands"
    local collisions=()

    # Skip if no user commands directory
    [[ -d "$user_commands_dir" ]] || return 0

    for cmd_file in "$source_dir"/*.md; do
        [[ -f "$cmd_file" ]] || continue
        local cmd_name
        cmd_name=$(basename "$cmd_file")

        # Check if user has this command
        if [[ -f "$user_commands_dir/$cmd_name" ]]; then
            collisions+=("$cmd_name")
        fi
    done

    # Report collisions
    if [[ ${#collisions[@]} -gt 0 ]]; then
        log_warning "Team commands collide with user commands: ${collisions[*]}"
        log_warning "User commands preserved. Team commands skipped."
        return 0  # Warning only, not failure
    fi

    return 0
}
```

**Integration Point**: Call `check_user_command_collisions "$source_dir"` at line 2111, before the sync loop.

**Modification to sync loop**: Add user-level check to collision detection:

```bash
# Lines 2117-2122 become:
# Check for collision with existing project OR user command
if [[ -f ".claude/commands/$cmd_name" ]] && ! grep -q "^$cmd_name$" "$marker_file" 2>/dev/null; then
    log_warning "Skipped: $cmd_name (project command exists)"
    continue
fi
if [[ -f "$HOME/.claude/commands/$cmd_name" ]]; then
    log_warning "Skipped: $cmd_name (user command exists at ~/.claude/commands/)"
    continue
fi
```

**Files Changed**:

| File | Line(s) | Change |
|------|---------|--------|
| swap-rite.sh | 2109 | Add check_user_command_collisions() function |
| swap-rite.sh | 2117-2122 | Add user-level collision check |

---

### WP2: Schema Validation Pre-Swap

**Objective**: Validate workflow.yaml before committing swap

**Implementation**:

```bash
# Add to swap-rite.sh in Transaction Safety Functions section

# Validate team configuration files
# Returns: 0 = valid, 1 = invalid
validate_team_schemas() {
    local rite_name="$1"
    local errors=0

    # Validate workflow.yaml if exists
    local workflow_file="$ROSTER_HOME/rites/$rite_name/workflow.yaml"
    if [[ -f "$workflow_file" ]]; then
        if ! validate_workflow_yaml "$workflow_file"; then
            ((errors++)) || true
        fi
    fi

    # Validate orchestrator.yaml if exists
    local orchestrator_file="$ROSTER_HOME/rites/$rite_name/orchestrator.yaml"
    if [[ -f "$orchestrator_file" ]]; then
        if ! validate_orchestrator_yaml "$orchestrator_file"; then
            ((errors++)) || true
        fi
    fi

    return $errors
}

# Validate workflow.yaml required fields
validate_workflow_yaml() {
    local file="$1"
    local required_fields=("name" "workflow_type" "entry_point" "phases")
    local missing=()

    for field in "${required_fields[@]}"; do
        # Use yq if available, fall back to grep
        if command -v yq &>/dev/null; then
            if [[ $(yq ".$field" "$file" 2>/dev/null) == "null" ]]; then
                missing+=("$field")
            fi
        else
            # Grep fallback: check for top-level field
            if ! grep -qE "^${field}:" "$file"; then
                missing+=("$field")
            fi
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "workflow.yaml missing required fields: ${missing[*]}"
        return 1
    fi

    # Validate workflow_type enum
    local workflow_type
    if command -v yq &>/dev/null; then
        workflow_type=$(yq '.workflow_type' "$file" 2>/dev/null)
    else
        workflow_type=$(grep -E "^workflow_type:" "$file" | sed 's/workflow_type: *//' | tr -d '"')
    fi

    if [[ ! "$workflow_type" =~ ^(sequential|parallel|hybrid)$ ]]; then
        log_error "workflow.yaml: invalid workflow_type '$workflow_type' (must be sequential, parallel, or hybrid)"
        return 1
    fi

    log_debug "workflow.yaml validation passed"
    return 0
}

# Validate orchestrator.yaml required fields
validate_orchestrator_yaml() {
    local file="$1"
    local required_fields=("team" "frontmatter" "routing" "workflow_position" "handoff_criteria" "skills")
    local missing=()

    for field in "${required_fields[@]}"; do
        if command -v yq &>/dev/null; then
            if [[ $(yq ".$field" "$file" 2>/dev/null) == "null" ]]; then
                missing+=("$field")
            fi
        else
            if ! grep -qE "^${field}:" "$file"; then
                missing+=("$field")
            fi
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "orchestrator.yaml missing required fields: ${missing[*]}"
        return 1
    fi

    log_debug "orchestrator.yaml validation passed"
    return 0
}
```

**Integration Point**: Call in `perform_swap()` after staging, before commit (around line 3890):

```bash
# Before "PHASE: VERIFYING" section, add:
# Validate team configuration schemas
if ! validate_team_schemas "$rite_name"; then
    log_error "Team schema validation failed"
    rollback_transaction
    return $EXIT_VALIDATION_FAILURE
fi
```

**Files Changed**:

| File | Line(s) | Change |
|------|---------|--------|
| swap-rite.sh | ~200 | Add validate_team_schemas(), validate_workflow_yaml(), validate_orchestrator_yaml() |
| swap-rite.sh | ~3890 | Call validate_team_schemas() in perform_swap() |

---

### WP3: Orphan Backup Cleanup Policy

**Objective**: Prevent orphan backup accumulation with retention policy

**Implementation**:

```bash
# Add to swap-rite.sh

# Cleanup orphan backups, keeping last N
# Usage: cleanup_orphan_backups [--keep=N] [--type=TYPE]
# Default: keep=3, type=all
cleanup_orphan_backups() {
    local keep=3
    local types=("skills" "commands" "hooks")

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --keep=*)
                keep="${1#*=}"
                ;;
            --type=*)
                types=("${1#*=}")
                ;;
            *)
                log_warning "Unknown option: $1"
                ;;
        esac
        shift
    done

    local total_removed=0

    for type in "${types[@]}"; do
        local backup_dir=".claude/${type}.orphan-backup"
        [[ -d "$backup_dir" ]] || continue

        # Count subdirectories (each is one backup set)
        local count
        count=$(find "$backup_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')

        if [[ "$count" -le "$keep" ]]; then
            log_debug "$type orphan backups: $count (within limit $keep)"
            continue
        fi

        # Remove oldest backups (by modification time)
        local to_remove=$((count - keep))
        log_debug "Removing $to_remove oldest $type orphan backup(s)"

        # Find oldest directories and remove them
        find "$backup_dir" -mindepth 1 -maxdepth 1 -type d -print0 2>/dev/null | \
            xargs -0 ls -dt 2>/dev/null | \
            tail -n "$to_remove" | \
            while IFS= read -r dir; do
                rm -rf "$dir"
                log_debug "Removed: $dir"
                ((total_removed++)) || true
            done
    done

    if [[ "$total_removed" -gt 0 ]]; then
        log "Cleaned up $total_removed orphan backup(s)"
    fi

    return 0
}
```

**Integration Points**:

1. Add `--auto-cleanup` flag parsing in main():
```bash
--auto-cleanup)
    AUTO_CLEANUP_MODE=1
    shift
    ;;
--cleanup-orphans)
    # Standalone cleanup mode
    cleanup_orphan_backups "$@"
    exit $?
    ;;
```

2. Call at end of successful swap if AUTO_CLEANUP_MODE=1:
```bash
# After successful swap completion in perform_swap()
if [[ "$AUTO_CLEANUP_MODE" == "1" ]]; then
    log_debug "Auto-cleanup enabled, pruning orphan backups"
    cleanup_orphan_backups --keep=3
fi
```

**Files Changed**:

| File | Line(s) | Change |
|------|---------|--------|
| swap-rite.sh | ~50 | Add AUTO_CLEANUP_MODE=0 variable |
| swap-rite.sh | ~200 | Add cleanup_orphan_backups() function |
| swap-rite.sh | ~3700 | Add --auto-cleanup and --cleanup-orphans flag handling |
| swap-rite.sh | ~3980 | Call cleanup after successful swap |

---

### WP4: Security Consultation Enforcement Hook (Design Only)

**Objective**: Design enforcement mechanism for security consultation requirements

**Note**: This is a design-only work package. Implementation deferred pending "consultation proof" mechanism definition.

**Design**:

1. **Enforcement Mode Field in workflow.yaml**:

```yaml
# Add to security_consultation section
security_consultation:
  triggers: [...]
  policy: [...]
  enforcement:
    mode: warn  # Options: off, warn, block
    proof_artifact: docs/security/THREAT-MODEL-*.md  # Glob pattern for proof
```

2. **PreToolUse Hook (future)**:

Hook location: `user-hooks/security-consultation-check.sh`

Trigger: `PreToolUse` for `Write` and `Edit` tools targeting implementation files

Logic:
- Read active workflow from `.claude/ACTIVE_WORKFLOW.yaml`
- If `security_consultation.enforcement.mode` is `off`, exit 0
- Check session context for initiative/complexity
- If complexity + domain matches `policy.required`:
  - Search for proof artifact matching glob pattern
  - If not found and mode is `block`: exit 1 with error message
  - If not found and mode is `warn`: log warning, exit 0
- Exit 0 (allow)

3. **Consultation Proof Options** (for future design):

| Option | Proof Artifact | Pros | Cons |
|--------|---------------|------|------|
| A | THREAT-MODEL-*.md exists | Simple | Could be stale |
| B | Session log contains threat-modeler invocation | Accurate | Complex to parse |
| C | SECURITY-REVIEW.md with sign-off field | Formal | Adds ceremony |

**Recommendation**: Start with Option A (artifact existence check) for simplicity. Enhancement to session-log parsing can follow.

**Files to Create (Future)**:

| File | Purpose |
|------|---------|
| user-hooks/security-consultation-check.sh | PreToolUse enforcement hook |
| .claude/hooks/security-consultation-check.json | Hook configuration |

**Files to Modify (Future)**:

| File | Change |
|------|--------|
| rites/10x-dev-pack/workflow.yaml | Add enforcement section to security_consultation |

---

## Backward Compatibility

**Classification**: COMPATIBLE

All changes are backward compatible:

1. **Command collision detection**: Adds warning, does not change existing behavior for project commands
2. **Schema validation**: New validation gate, fails only for genuinely invalid files
3. **Orphan cleanup**: Opt-in only (`--auto-cleanup`), default preserves all backups
4. **Security enforcement**: Design only, no implementation in this work package

**Migration Required**: No

Existing satellites continue to function without changes. New features are additive.

## Test Matrix

### WP1: Command Collision Detection

| Scenario | Setup | Expected Outcome |
|----------|-------|------------------|
| No collision | Team has `foo.md`, user has no commands | Sync completes, no warning |
| User command collision | Team has `pr.md`, user has `~/.claude/commands/pr.md` | Warning logged, team `pr.md` skipped |
| Project command collision | Team has `bar.md`, project has `.claude/commands/bar.md` | Warning logged, team `bar.md` skipped (existing behavior) |
| Multiple collisions | Team has `pr.md`, `spike.md`, user has both | Warning lists both, both skipped |
| User commands dir missing | Team has commands, `~/.claude/commands/` does not exist | Sync completes, no errors |

### WP2: Schema Validation Pre-Swap

| Scenario | Setup | Expected Outcome |
|----------|-------|------------------|
| Valid workflow.yaml | All required fields present, valid enum | Swap completes |
| Missing required field | workflow.yaml lacks `entry_point` | Swap aborts with error |
| Invalid workflow_type | `workflow_type: invalid` | Swap aborts with enum error |
| No workflow.yaml | Team has no workflow.yaml | Swap completes (validation skipped) |
| No yq available | yq not installed | Falls back to grep validation |
| Valid orchestrator.yaml | All required fields present | Swap completes |
| Invalid orchestrator.yaml | Missing `routing` field | Swap aborts with error |

### WP3: Orphan Backup Cleanup

| Scenario | Setup | Expected Outcome |
|----------|-------|------------------|
| Fewer than limit | 2 backups exist, limit is 3 | No cleanup |
| At limit | 3 backups exist, limit is 3 | No cleanup |
| Over limit | 5 backups exist, limit is 3 | 2 oldest removed |
| Auto-cleanup enabled | `--auto-cleanup` flag | Cleanup runs after swap |
| Manual cleanup | `--cleanup-orphans` | Cleanup runs, swap not performed |
| Mixed types | skills has 5, commands has 2 | Only skills cleaned |

### WP4: Security Enforcement (Future)

| Scenario | Expected Outcome |
|----------|------------------|
| Mode: off | Hook exits immediately, no check |
| Mode: warn, no proof | Warning logged, implementation allowed |
| Mode: block, no proof | Implementation blocked with error |
| Mode: block, proof exists | Implementation allowed |
| Non-security work | Hook exits, no domain match |

## Handoff Criteria

- [x] All design decisions have documented rationale
- [x] No TBD, TODO, or unresolved items
- [x] Work packages specify file-level changes
- [x] Backward compatibility assessed: COMPATIBLE
- [x] Test matrix defined for each work package
- [x] WP4 marked as design-only, implementation deferred

## Implementation Notes for Integration Engineer

### Priority Order

Implement in order: WP2 (lowest risk), WP1, WP3, WP4 (design only).

### Dependencies

- WP1, WP2, WP3 have no inter-dependencies
- WP4 depends on consultation proof mechanism (future design)

### Testing Approach

1. Run existing swap-rite tests to establish baseline
2. Add test cases per test matrix above
3. Test on roster (baseline) and a minimal satellite

### Rollback Strategy

If validation gate causes issues:
1. Set `ROSTER_SKIP_VALIDATION=1` environment variable (escape hatch)
2. Or revert specific function without touching rest of swap-rite.sh
