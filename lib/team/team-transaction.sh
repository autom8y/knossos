#!/usr/bin/env bash
#
# team-transaction.sh - Transaction Infrastructure for Team Swaps
#
# Provides atomic write, journal management, staging, and backup
# operations for swap-team.sh transaction safety.
#
# Part of: roster team-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/team/team-transaction.sh"
#   create_journal "$source_team" "$target_team"
#   create_staging && stage_agents "$team_name" && verify_staging "$count"
#
# Dependencies:
#   - jq (for JSON manipulation)
#   - Logging functions (log, log_debug, log_warning, log_error)
#   - Constants: JOURNAL_FILE, STAGING_DIR, SWAP_BACKUP_DIR, etc.

# Guard against re-sourcing
[[ -n "${_TEAM_TRANSACTION_LOADED:-}" ]] && return 0
readonly _TEAM_TRANSACTION_LOADED=1

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
# Logging Stubs (overridden when sourced from swap-team.sh)
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
#   $1 - source_team: Current team (empty string for virgin swap)
#   $2 - target_team: Team being swapped to
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
    "active_team": "$SWAP_BACKUP_DIR/ACTIVE_TEAM",
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

# Stage agents from team pack
# Parameters:
#   $1 - team_name: Team to stage agents from
# Returns: 0 on success, 1 on failure
# Requires: ROSTER_HOME, STAGING_DIR
stage_agents() {
    local team_name="$1"
    local source_dir="$ROSTER_HOME/teams/$team_name/agents"
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
#   $1 - team_name: Team to stage workflow from
# Returns: 0 on success, 1 on failure (warning if no workflow.yaml)
# Requires: ROSTER_HOME, STAGING_DIR
stage_workflow() {
    local team_name="$1"
    local source_file="$ROSTER_HOME/teams/$team_name/workflow.yaml"

    if [[ -f "$source_file" ]]; then
        cp "$source_file" "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" || {
            log_warning "Failed to stage workflow.yaml"
            return 1
        }
        log_debug "Staged workflow.yaml"
    fi

    return 0
}

# Stage ACTIVE_TEAM file
# Parameters:
#   $1 - team_name: Team name to write
# Returns: 0 on success, 1 on failure
# Requires: STAGING_DIR
stage_active_team() {
    local team_name="$1"

    echo -n "$team_name" > "$STAGING_DIR/ACTIVE_TEAM" || {
        log_error "Failed to stage ACTIVE_TEAM"
        return 1
    }

    log_debug "Staged ACTIVE_TEAM: $team_name"
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

    # Verify ACTIVE_TEAM staged
    if [[ ! -f "$STAGING_DIR/ACTIVE_TEAM" ]]; then
        log_error "Staged ACTIVE_TEAM missing"
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
    # Function stub - to be implemented
    return 1
}

# Clean up swap backup (after successful swap)
# Returns: 0 always
# Side effects: Removes SWAP_BACKUP_DIR if exists
cleanup_swap_backup() {
    # Function stub - to be implemented
    return 0
}

# Verify backup integrity for recovery
# Returns: 0 if backup valid, 1 if missing/corrupted
# Requires: SWAP_BACKUP_DIR, JOURNAL_FILE (for virgin swap check)
verify_backup_integrity() {
    # Function stub - to be implemented
    return 1
}
