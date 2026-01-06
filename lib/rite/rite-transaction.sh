#!/usr/bin/env bash
#
# rite-transaction.sh - Transaction Infrastructure for Rite Swaps
#
# Provides atomic write, journal management, staging, and backup
# operations for swap-rite.sh transaction safety.
#
# Part of: roster rite-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/rite/rite-transaction.sh"
#   create_journal "$source_rite" "$target_rite"
#   create_staging && stage_agents "$rite_name" && verify_staging "$count"
#
# Dependencies:
#   - jq (for JSON manipulation)
#   - Logging functions (log, log_debug, log_warning, log_error)
#   - Constants: JOURNAL_FILE, STAGING_DIR, SWAP_BACKUP_DIR, etc.

# Guard against re-sourcing
[[ -n "${_RITE_TRANSACTION_LOADED:-}" ]] && return 0
readonly _RITE_TRANSACTION_LOADED=1

# ============================================================================
# Module Constants (if not defined by caller)
# ============================================================================

# Default paths (can be overridden before sourcing)
: "${JOURNAL_FILE:=.claude/.swap-journal}"
: "${JOURNAL_VERSION:=1.0}"
: "${STAGING_DIR:=.claude/.swap-staging}"
: "${SWAP_BACKUP_DIR:=.claude/.swap-backup}"

# Transaction phases
: "${PHASE_PREPARING:=PREPARING}"
: "${PHASE_BACKING:=BACKING}"
: "${PHASE_STAGING:=STAGING}"
: "${PHASE_VERIFYING:=VERIFYING}"
: "${PHASE_COMMITTING:=COMMITTING}"
: "${PHASE_COMPLETED:=COMPLETED}"

# ============================================================================
# Logging Stubs (overridden when sourced from swap-rite.sh)
# ============================================================================

if ! type log >/dev/null 2>&1; then
    log() { echo "[Transaction] $*"; }
fi

if ! type log_debug >/dev/null 2>&1; then
    log_debug() { echo "[DEBUG] $*" >&2; }
fi

if ! type log_warning >/dev/null 2>&1; then
    log_warning() { echo "[WARNING] $*" >&2; }
fi

if ! type log_error >/dev/null 2>&1; then
    log_error() { echo "[ERROR] $*" >&2; }
fi

# ============================================================================
# Atomic I/O
# ============================================================================

# Write content atomically using temp file + rename pattern
# Parameters:
#   $1 - target: Target file path
#   $2 - content: Content to write
# Returns: 0 on success, 1 on failure
write_atomic() {
    local target="$1"
    local content="$2"
    local temp="${target}.tmp.$$"

    # Ensure parent directory exists
    local parent_dir
    parent_dir=$(dirname "$target")
    mkdir -p "$parent_dir" || {
        log_error "Cannot create directory: $parent_dir"
        return 1
    }

    # Write to temp file
    printf '%s' "$content" > "$temp" || {
        rm -f "$temp" 2>/dev/null
        log_error "Failed to write temp file: $temp"
        return 1
    }

    # Sync to disk (best effort)
    sync "$temp" 2>/dev/null || true

    # Atomic rename
    mv "$temp" "$target" || {
        rm -f "$temp" 2>/dev/null
        log_error "Failed to rename temp file to: $target"
        return 1
    }

    return 0
}

# ============================================================================
# Journal Operations
# ============================================================================

# Create a new journal entry for swap operation
# Parameters:
#   $1 - source_team: Current rite (empty string for virgin swap)
#   $2 - target_team: Rite being swapped to
# Returns: 0 on success, 1 if journal already exists (concurrent swap)
# Requires: JOURNAL_FILE, JOURNAL_VERSION, SWAP_BACKUP_DIR, STAGING_DIR
create_journal() {
    local source_team="$1"
    local target_team="$2"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Check if journal already exists (concurrent swap protection)
    if [[ -f "$JOURNAL_FILE" ]]; then
        local existing_pid
        existing_pid=$(jq -r '.pid // "unknown"' "$JOURNAL_FILE" 2>/dev/null)
        log_error "Swap already in progress (journal exists, PID: $existing_pid)"
        log "Use --recover to handle the interrupted swap"
        return 1
    fi

    # Format source_team for JSON (null if empty, quoted string otherwise)
    local source_team_json="null"
    if [[ -n "$source_team" ]]; then
        source_team_json="\"$source_team\""
    fi

    local journal_content
    journal_content=$(cat <<EOF
{
  "version": "$JOURNAL_VERSION",
  "started_at": "$timestamp",
  "phase": "$PHASE_PREPARING",
  "source_team": $source_team_json,
  "target_team": "$target_team",
  "backup_location": {
    "agents": "$SWAP_BACKUP_DIR/agents",
    "manifest": "$SWAP_BACKUP_DIR/AGENT_MANIFEST.json",
    "active_rite": "$SWAP_BACKUP_DIR/ACTIVE_RITE",
    "workflow": "$SWAP_BACKUP_DIR/ACTIVE_WORKFLOW.yaml",
    "commands": null,
    "skills": null,
    "hooks": null
  },
  "staging_location": "$STAGING_DIR",
  "checksums": {},
  "pid": $$,
  "error": null
}
EOF
)

    write_atomic "$JOURNAL_FILE" "$journal_content" || {
        log_error "Failed to create swap journal"
        return 1
    }

    log_debug "Journal created: $source_team -> $target_team"
    return 0
}

# Update journal phase
# Parameters:
#   $1 - new_phase: Phase name (PHASE_* constant)
# Returns: 0 on success, 1 if journal missing
update_journal_phase() {
    local new_phase="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        log_error "Cannot update phase: journal does not exist"
        return 1
    fi

    local updated
    updated=$(jq --arg phase "$new_phase" '.phase = $phase' "$JOURNAL_FILE") || {
        log_error "Failed to parse journal for phase update"
        return 1
    }

    write_atomic "$JOURNAL_FILE" "$updated" || {
        log_error "Failed to update journal phase to: $new_phase"
        return 1
    }

    log_debug "Journal phase updated: $new_phase"
    return 0
}

# Update journal backup locations for resources
# Parameters:
#   $1 - resource_type: "commands" | "skills" | "hooks"
#   $2 - backup_path: Path to backup directory
# Returns: 0 on success, 1 if journal missing
update_journal_backups() {
    local resource_type="$1"
    local backup_path="$2"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        return 1
    fi

    local updated
    updated=$(jq --arg type "$resource_type" --arg path "$backup_path" \
        '.backup_location[$type] = $path' "$JOURNAL_FILE") || return 1

    write_atomic "$JOURNAL_FILE" "$updated"
}

# Update journal with error message
# Parameters:
#   $1 - error_msg: Error message to record
# Returns: 0 on success, 1 if journal missing
update_journal_error() {
    local error_msg="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        return 1
    fi

    local updated
    updated=$(jq --arg err "$error_msg" '.error = $err' "$JOURNAL_FILE") || return 1

    write_atomic "$JOURNAL_FILE" "$updated"
}

# Record manifest write timestamp in journal
# Parameters:
#   $1 - timestamp: ISO 8601 timestamp when manifest was written
# Returns: 0 on success, 1 if journal missing
# Note: Used for staleness detection during recovery
update_journal_manifest_timestamp() {
    local timestamp="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        return 1
    fi

    local updated
    updated=$(jq --arg ts "$timestamp" '.manifest_written_at = $ts' "$JOURNAL_FILE") || return 1

    write_atomic "$JOURNAL_FILE" "$updated"
}

# Get manifest write timestamp from journal
# Outputs: Timestamp to stdout, empty if not found
# Returns: 0 always (empty output for missing field)
get_journal_manifest_timestamp() {
    if [[ ! -f "$JOURNAL_FILE" ]]; then
        echo ""
        return 1
    fi

    jq -r '.manifest_written_at // empty' "$JOURNAL_FILE" 2>/dev/null
}

# Read arbitrary journal field
# Parameters:
#   $1 - field: Field name in journal JSON
# Outputs: Field value to stdout, empty if not found
# Returns: 0 always (empty output for missing field)
get_journal_field() {
    local field="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        echo ""
        return 1
    fi

    jq -r ".$field // empty" "$JOURNAL_FILE" 2>/dev/null
}

# Get current journal phase
# Outputs: Phase name to stdout
# Returns: 0 on success, 1 if journal missing
get_journal_phase() {
    get_journal_field "phase"
}

# Delete journal (on successful completion)
# Returns: 0 always
delete_journal() {
    if [[ -f "$JOURNAL_FILE" ]]; then
        rm -f "$JOURNAL_FILE"
        log_debug "Journal deleted"
    fi
}

# Check if journal exists
# Returns: 0 if exists, 1 otherwise
journal_exists() {
    [[ -f "$JOURNAL_FILE" ]]
}

# ============================================================================
# Commit Step Tracking
# ============================================================================
# Commit steps track sub-operations within the COMMITTING phase.
# This enables recovery to determine what completed and what remains.
#
# Steps are stored in journal.commit_steps as an object:
#   { "agents": true, "workflow": true, "commands": false, ... }
#
# The point-of-no-return is when "active_rite" step is marked complete.
# Before that point: rollback to backup
# After that point: complete the remaining steps

# Commit step constants
readonly COMMIT_STEP_AGENTS="agents"
readonly COMMIT_STEP_WORKFLOW="workflow"
readonly COMMIT_STEP_COMMANDS="commands"
readonly COMMIT_STEP_SKILLS="skills"
readonly COMMIT_STEP_SHARED_SKILLS="shared_skills"
readonly COMMIT_STEP_HOOKS="hooks"
readonly COMMIT_STEP_HOOK_REGISTRATIONS="hook_registrations"
readonly COMMIT_STEP_MANIFEST="manifest"
readonly COMMIT_STEP_CEM_MANIFEST="cem_manifest"
readonly COMMIT_STEP_ACTIVE_RITE="active_rite"  # Point-of-no-return

# Initialize commit steps tracking in journal
# Called when entering COMMITTING phase
# Returns: 0 on success, 1 if journal missing
init_commit_steps() {
    if [[ ! -f "$JOURNAL_FILE" ]]; then
        log_error "Cannot init commit steps: journal does not exist"
        return 1
    fi

    local updated
    updated=$(jq '.commit_steps = {
        "agents": false,
        "workflow": false,
        "commands": false,
        "skills": false,
        "shared_skills": false,
        "hooks": false,
        "hook_registrations": false,
        "manifest": false,
        "cem_manifest": false,
        "active_rite": false
    }' "$JOURNAL_FILE") || {
        log_error "Failed to parse journal for commit steps init"
        return 1
    }

    write_atomic "$JOURNAL_FILE" "$updated" || {
        log_error "Failed to init commit steps in journal"
        return 1
    }

    log_debug "Commit steps initialized"
    return 0
}

# Mark a commit step as completed
# Parameters:
#   $1 - step: Step name (COMMIT_STEP_* constant)
# Returns: 0 on success, 1 if journal missing or update fails
mark_commit_step() {
    local step="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        log_error "Cannot mark commit step: journal does not exist"
        return 1
    fi

    local updated
    updated=$(jq --arg step "$step" '.commit_steps[$step] = true' "$JOURNAL_FILE") || {
        log_error "Failed to parse journal for commit step update"
        return 1
    }

    write_atomic "$JOURNAL_FILE" "$updated" || {
        log_error "Failed to mark commit step: $step"
        return 1
    }

    log_debug "Commit step completed: $step"
    return 0
}

# Check if a commit step is completed
# Parameters:
#   $1 - step: Step name to check
# Returns: 0 if completed, 1 if not completed or journal missing
is_commit_step_done() {
    local step="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        return 1
    fi

    local result
    result=$(jq -r --arg step "$step" '.commit_steps[$step] // false' "$JOURNAL_FILE" 2>/dev/null)
    [[ "$result" == "true" ]]
}

# Check if we're past the point-of-no-return (ACTIVE_RITE written)
# Returns: 0 if past point-of-no-return, 1 if before
is_past_point_of_no_return() {
    is_commit_step_done "$COMMIT_STEP_ACTIVE_RITE"
}

# Get list of incomplete commit steps
# Outputs: Space-separated list of incomplete steps
get_incomplete_commit_steps() {
    if [[ ! -f "$JOURNAL_FILE" ]]; then
        echo ""
        return 1
    fi

    jq -r '.commit_steps | to_entries | map(select(.value == false)) | .[].key' "$JOURNAL_FILE" 2>/dev/null | tr '\n' ' '
}

# ============================================================================
# Staging Operations
# ============================================================================

# Create staging directory structure
# Returns: 0 on success, 1 on failure
# Side effects: Creates STAGING_DIR, removes any existing staging
create_staging() {
    log_debug "Creating staging directory: $STAGING_DIR"

    # Clean any existing staging
    rm -rf "$STAGING_DIR" 2>/dev/null

    mkdir -p "$STAGING_DIR" || {
        log_error "Failed to create staging directory"
        return 1
    }

    return 0
}

# Clean up staging directory
# Returns: 0 always
# Side effects: Removes STAGING_DIR if exists
cleanup_staging() {
    if [[ -d "$STAGING_DIR" ]]; then
        rm -rf "$STAGING_DIR"
        log_debug "Staging directory cleaned up"
    fi
}

# Stage agents from rite pack
# Parameters:
#   $1 - team_name: Rite to stage agents from
# Returns: 0 on success, 1 on failure
# Requires: ROSTER_HOME, STAGING_DIR
stage_agents() {
    local team_name="$1"
    local source_dir="$ROSTER_HOME/rites/$team_name/agents"
    local staging_agents="$STAGING_DIR/agents"

    if [[ ! -d "$source_dir" ]]; then
        log_error "Source agents directory not found: $source_dir"
        return 1
    fi

    mkdir -p "$staging_agents" || return 1

    cp -rp "$source_dir"/* "$staging_agents/" || {
        log_error "Failed to stage agents"
        return 1
    }

    log_debug "Staged agents to: $staging_agents"
    return 0
}

# Stage workflow file
# Parameters:
#   $1 - team_name: Rite to stage workflow from
# Returns: 0 on success, 1 on failure (warning if no workflow.yaml)
# Requires: ROSTER_HOME, STAGING_DIR
stage_workflow() {
    local team_name="$1"
    local source_file="$ROSTER_HOME/rites/$team_name/workflow.yaml"

    if [[ -f "$source_file" ]]; then
        cp "$source_file" "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" || {
            log_warning "Failed to stage workflow.yaml"
            return 1
        }
        log_debug "Staged workflow.yaml"
    fi

    return 0
}

# Stage ACTIVE_RITE file
# Parameters:
#   $1 - rite_name: Rite name to write
# Returns: 0 on success, 1 on failure
# Requires: STAGING_DIR
stage_active_rite() {
    local rite_name="$1"

    echo -n "$rite_name" > "$STAGING_DIR/ACTIVE_RITE" || {
        log_error "Failed to stage ACTIVE_RITE"
        return 1
    }

    log_debug "Staged ACTIVE_RITE: $rite_name"
    return 0
}

# Verify staging directory integrity
# Parameters:
#   $1 - expected_count: Expected number of agent .md files
# Returns: 0 on success, 1 on verification failure
# Requires: STAGING_DIR
verify_staging() {
    local expected_count="$1"

    # Verify staging directory exists
    if [[ ! -d "$STAGING_DIR" ]]; then
        log_error "Staging directory missing"
        return 1
    fi

    # Verify agents staged
    if [[ ! -d "$STAGING_DIR/agents" ]]; then
        log_error "Staged agents directory missing"
        return 1
    fi

    # Verify agent count
    local actual_count
    actual_count=$(find "$STAGING_DIR/agents" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$actual_count" -ne "$expected_count" ]]; then
        log_error "Staging verification failed: expected $expected_count agents, found $actual_count"
        return 1
    fi

    # Verify ACTIVE_RITE staged
    if [[ ! -f "$STAGING_DIR/ACTIVE_RITE" ]]; then
        log_error "Staged ACTIVE_RITE missing"
        return 1
    fi

    log_debug "Staging verified: $actual_count agents"
    return 0
}

# ============================================================================
# Swap Backup Operations
# ============================================================================

# Create comprehensive backup for transaction safety
# Returns: 0 on success, 1 on failure
# Requires: SWAP_BACKUP_DIR, MANIFEST_FILE (for backup path in journal)
# Side effects: Updates journal with backup locations
create_swap_backup() {
    log_debug "Creating comprehensive backup in $SWAP_BACKUP_DIR"

    # Clean any existing swap backup
    rm -rf "$SWAP_BACKUP_DIR" 2>/dev/null

    mkdir -p "$SWAP_BACKUP_DIR" || {
        log_error "Failed to create swap backup directory"
        return 1
    }

    # Backup agents if they exist
    if [[ -d ".claude/agents" ]] && [[ -n "$(ls -A .claude/agents 2>/dev/null)" ]]; then
        cp -rp .claude/agents "$SWAP_BACKUP_DIR/agents" || {
            log_error "Failed to backup agents"
            return 1
        }
        log_debug "Backed up agents"
    fi

    # Backup ACTIVE_TEAM if exists
    if [[ -f ".claude/ACTIVE_RITE" ]]; then
        cp .claude/ACTIVE_RITE "$SWAP_BACKUP_DIR/ACTIVE_TEAM" || {
            log_error "Failed to backup ACTIVE_TEAM"
            return 1
        }
        log_debug "Backed up ACTIVE_TEAM"
    fi

    # Backup AGENT_MANIFEST.json if exists
    if [[ -f "$MANIFEST_FILE" ]]; then
        cp "$MANIFEST_FILE" "$SWAP_BACKUP_DIR/AGENT_MANIFEST.json" || {
            log_warning "Failed to backup manifest"
        }
        log_debug "Backed up manifest"
    fi

    # Backup ACTIVE_WORKFLOW.yaml if exists
    if [[ -f ".claude/ACTIVE_WORKFLOW.yaml" ]]; then
        cp .claude/ACTIVE_WORKFLOW.yaml "$SWAP_BACKUP_DIR/ACTIVE_WORKFLOW.yaml" || {
            log_warning "Failed to backup workflow"
        }
        log_debug "Backed up workflow"
    fi

    # Backup commands if rite commands exist
    if [[ -f ".claude/commands/.rite-commands" ]]; then
        mkdir -p "$SWAP_BACKUP_DIR/commands"
        while IFS= read -r cmd_file; do
            [[ -z "$cmd_file" ]] && continue
            if [[ -f ".claude/commands/$cmd_file" ]]; then
                cp ".claude/commands/$cmd_file" "$SWAP_BACKUP_DIR/commands/$cmd_file"
            fi
        done < ".claude/commands/.rite-commands"
        cp ".claude/commands/.rite-commands" "$SWAP_BACKUP_DIR/commands/.rite-commands"
        update_journal_backups "commands" "$SWAP_BACKUP_DIR/commands"
        log_debug "Backed up rite commands"
    fi

    # Backup skills if rite skills exist
    if [[ -f ".claude/skills/.rite-skills" ]]; then
        mkdir -p "$SWAP_BACKUP_DIR/skills"
        while IFS= read -r skill_dir; do
            [[ -z "$skill_dir" ]] && continue
            if [[ -d ".claude/skills/$skill_dir" ]]; then
                cp -rp ".claude/skills/$skill_dir" "$SWAP_BACKUP_DIR/skills/$skill_dir"
            fi
        done < ".claude/skills/.rite-skills"
        cp ".claude/skills/.rite-skills" "$SWAP_BACKUP_DIR/skills/.rite-skills"
        update_journal_backups "skills" "$SWAP_BACKUP_DIR/skills"
        log_debug "Backed up rite skills"
    fi

    # Backup hooks if rite hooks exist
    if [[ -f ".claude/hooks/.rite-hooks" ]]; then
        mkdir -p "$SWAP_BACKUP_DIR/hooks"
        while IFS= read -r hook_file; do
            [[ -z "$hook_file" ]] && continue
            if [[ -f ".claude/hooks/$hook_file" ]]; then
                cp ".claude/hooks/$hook_file" "$SWAP_BACKUP_DIR/hooks/$hook_file"
            fi
        done < ".claude/hooks/.rite-hooks"
        cp ".claude/hooks/.rite-hooks" "$SWAP_BACKUP_DIR/hooks/.rite-hooks"
        update_journal_backups "hooks" "$SWAP_BACKUP_DIR/hooks"
        log_debug "Backed up rite hooks"
    fi

    log_debug "Comprehensive backup complete"
    return 0
}

# Clean up swap backup (after successful swap)
# Returns: 0 always
# Side effects: Removes SWAP_BACKUP_DIR if exists
cleanup_swap_backup() {
    if [[ -d "$SWAP_BACKUP_DIR" ]]; then
        rm -rf "$SWAP_BACKUP_DIR"
        log_debug "Swap backup cleaned up"
    fi
}

# Verify backup integrity for recovery
# Returns: 0 if backup valid, 1 if missing/corrupted
# Requires: SWAP_BACKUP_DIR, JOURNAL_FILE (for virgin swap check)
verify_backup_integrity() {
    if [[ ! -d "$SWAP_BACKUP_DIR" ]]; then
        log_debug "Backup directory missing"
        return 1
    fi

    # For virgin swap, backup may not have agents
    if [[ -d "$SWAP_BACKUP_DIR/agents" ]]; then
        local agent_count
        agent_count=$(find "$SWAP_BACKUP_DIR/agents" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        if [[ "$agent_count" -eq 0 ]]; then
            log_debug "Backup has no agent files (may be virgin swap)"
        fi
    fi

    # Check if this was a virgin swap (source_team is null in journal)
    local was_virgin="false"
    if [[ -f "$JOURNAL_FILE" ]]; then
        local source_team
        source_team=$(jq -r '.source_team // "null"' "$JOURNAL_FILE" 2>/dev/null)
        if [[ "$source_team" == "null" ]]; then
            was_virgin="true"
        fi
    fi

    # For non-virgin swap, we need ACTIVE_TEAM in backup
    if [[ "$was_virgin" != "true" ]] && [[ ! -f "$SWAP_BACKUP_DIR/ACTIVE_TEAM" ]]; then
        log_debug "Backup missing ACTIVE_TEAM (non-virgin swap)"
        return 1
    fi

    return 0
}
