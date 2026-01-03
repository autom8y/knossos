#!/usr/bin/env bash
#
# swap-team.sh - Agent Team Pack Management System
#
# Swaps Claude Code agent team packs with atomic-ish operations.
# See TDD-0003 for design details.

set -euo pipefail

# Constants
readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly ROSTER_DEBUG="${ROSTER_DEBUG:-0}"
readonly EXIT_SUCCESS=0
readonly EXIT_INVALID_ARGS=1
readonly EXIT_VALIDATION_FAILURE=2
readonly EXIT_BACKUP_FAILURE=3
readonly EXIT_SWAP_FAILURE=4
readonly EXIT_ORPHAN_CONFLICT=5
readonly EXIT_RECOVERY_REQUIRED=6

# Transaction safety constants
readonly JOURNAL_FILE=".claude/.swap-journal"
readonly JOURNAL_VERSION="1.0"
readonly STAGING_DIR=".claude/.swap-staging"
readonly SWAP_BACKUP_DIR=".claude/.swap-backup"

# Transaction phases
readonly PHASE_PREPARING="PREPARING"
readonly PHASE_BACKING="BACKING"
readonly PHASE_STAGING="STAGING"
readonly PHASE_VERIFYING="VERIFYING"
readonly PHASE_COMMITTING="COMMITTING"
readonly PHASE_COMPLETED="COMPLETED"

# Manifest file path
readonly MANIFEST_FILE=".claude/AGENT_MANIFEST.json"
readonly MANIFEST_VERSION="1.2"

# Orphan handling mode (set by flags)
ORPHAN_MODE=""  # "", "keep", "remove", "promote"

# Update mode: re-pull agents even if already on target team
UPDATE_MODE=0

# Dry-run mode: preview changes without applying
DRY_RUN_MODE=0
RESET_MODE=0

# Recovery modes
AUTO_RECOVER=0
RECOVER_MODE=0
VERIFY_MODE=0

# Colors for output (if terminal supports it)
if [[ -t 1 ]]; then
    readonly RED='\033[0;31m'
    readonly GREEN='\033[0;32m'
    readonly YELLOW='\033[1;33m'
    readonly NC='\033[0m' # No Color
else
    readonly RED=''
    readonly GREEN=''
    readonly YELLOW=''
    readonly NC=''
fi

# Logging functions
log() {
    echo "[Roster] $*"
}

log_error() {
    echo "[Roster] Error: $*" >&2
}

log_warning() {
    echo "[Roster] Warning: $*" >&2
}

log_debug() {
    if [[ "$ROSTER_DEBUG" == "1" ]]; then
        echo "[Roster DEBUG] $*" >&2
    fi
}

# ============================================================================
# Transaction Safety Functions
# ============================================================================

# Write content atomically using temp file + rename pattern
# Usage: write_atomic "target_path" "content"
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

# Create a new journal entry for swap operation
# Usage: create_journal "source_team" "target_team"
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
# Usage: update_journal_phase "PHASE_NAME"
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
# Usage: update_journal_backups "commands" "$path" | update_journal_backups "skills" "$path" ...
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
# Usage: update_journal_error "error message"
update_journal_error() {
    local error_msg="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        return 1
    fi

    local updated
    updated=$(jq --arg err "$error_msg" '.error = $err' "$JOURNAL_FILE") || return 1

    write_atomic "$JOURNAL_FILE" "$updated"
}

# Read journal field
# Usage: get_journal_field "field_name"
get_journal_field() {
    local field="$1"

    if [[ ! -f "$JOURNAL_FILE" ]]; then
        echo ""
        return 1
    fi

    jq -r ".$field // empty" "$JOURNAL_FILE" 2>/dev/null
}

# Get current journal phase
get_journal_phase() {
    get_journal_field "phase"
}

# Delete journal (on successful completion)
delete_journal() {
    if [[ -f "$JOURNAL_FILE" ]]; then
        rm -f "$JOURNAL_FILE"
        log_debug "Journal deleted"
    fi
}

# Check if journal exists
journal_exists() {
    [[ -f "$JOURNAL_FILE" ]]
}

# ============================================================================
# Staging Functions
# ============================================================================

# Create staging directory structure
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
cleanup_staging() {
    if [[ -d "$STAGING_DIR" ]]; then
        rm -rf "$STAGING_DIR"
        log_debug "Staging directory cleaned up"
    fi
}

# Stage agents from team pack
# Usage: stage_agents "team_name"
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
# Usage: stage_workflow "team_name"
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
# Usage: stage_active_team "team_name"
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
# Usage: verify_staging "expected_agent_count"
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
# Comprehensive Backup Functions
# ============================================================================

# Create comprehensive backup for transaction safety
# Usage: create_swap_backup
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
    if [[ -f ".claude/ACTIVE_TEAM" ]]; then
        cp .claude/ACTIVE_TEAM "$SWAP_BACKUP_DIR/ACTIVE_TEAM" || {
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

    # Backup commands if team commands exist
    if [[ -f ".claude/commands/.team-commands" ]]; then
        mkdir -p "$SWAP_BACKUP_DIR/commands"
        while IFS= read -r cmd_file; do
            [[ -z "$cmd_file" ]] && continue
            if [[ -f ".claude/commands/$cmd_file" ]]; then
                cp ".claude/commands/$cmd_file" "$SWAP_BACKUP_DIR/commands/$cmd_file"
            fi
        done < ".claude/commands/.team-commands"
        cp ".claude/commands/.team-commands" "$SWAP_BACKUP_DIR/commands/.team-commands"
        update_journal_backups "commands" "$SWAP_BACKUP_DIR/commands"
        log_debug "Backed up team commands"
    fi

    # Backup skills if team skills exist
    if [[ -f ".claude/skills/.team-skills" ]]; then
        mkdir -p "$SWAP_BACKUP_DIR/skills"
        while IFS= read -r skill_dir; do
            [[ -z "$skill_dir" ]] && continue
            if [[ -d ".claude/skills/$skill_dir" ]]; then
                cp -rp ".claude/skills/$skill_dir" "$SWAP_BACKUP_DIR/skills/$skill_dir"
            fi
        done < ".claude/skills/.team-skills"
        cp ".claude/skills/.team-skills" "$SWAP_BACKUP_DIR/skills/.team-skills"
        update_journal_backups "skills" "$SWAP_BACKUP_DIR/skills"
        log_debug "Backed up team skills"
    fi

    # Backup hooks if team hooks exist
    if [[ -f ".claude/hooks/.team-hooks" ]]; then
        mkdir -p "$SWAP_BACKUP_DIR/hooks"
        while IFS= read -r hook_file; do
            [[ -z "$hook_file" ]] && continue
            if [[ -f ".claude/hooks/$hook_file" ]]; then
                cp ".claude/hooks/$hook_file" "$SWAP_BACKUP_DIR/hooks/$hook_file"
            fi
        done < ".claude/hooks/.team-hooks"
        cp ".claude/hooks/.team-hooks" "$SWAP_BACKUP_DIR/hooks/.team-hooks"
        update_journal_backups "hooks" "$SWAP_BACKUP_DIR/hooks"
        log_debug "Backed up team hooks"
    fi

    log_debug "Comprehensive backup complete"
    return 0
}

# Clean up swap backup (after successful swap)
cleanup_swap_backup() {
    if [[ -d "$SWAP_BACKUP_DIR" ]]; then
        rm -rf "$SWAP_BACKUP_DIR"
        log_debug "Swap backup cleaned up"
    fi
}

# Verify backup integrity for recovery
# Usage: verify_backup_integrity
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

# ============================================================================
# Rollback Functions
# ============================================================================

# Rollback swap operation to previous state
rollback_swap() {
    log "Rolling back swap operation..."

    # Verify backup exists and is valid
    if [[ ! -d "$SWAP_BACKUP_DIR" ]]; then
        log_error "Cannot rollback: backup directory missing"
        log "Manual recovery required. Check $STAGING_DIR for new state."
        return 1
    fi

    # Restore agents
    if [[ -d "$SWAP_BACKUP_DIR/agents" ]]; then
        rm -rf .claude/agents 2>/dev/null
        cp -rp "$SWAP_BACKUP_DIR/agents" .claude/agents || {
            log_error "Failed to restore agents"
            return 1
        }
        log_debug "Restored agents"
    fi

    # Restore ACTIVE_WORKFLOW.yaml
    if [[ -f "$SWAP_BACKUP_DIR/ACTIVE_WORKFLOW.yaml" ]]; then
        cp "$SWAP_BACKUP_DIR/ACTIVE_WORKFLOW.yaml" .claude/ACTIVE_WORKFLOW.yaml
        log_debug "Restored workflow"
    else
        rm -f .claude/ACTIVE_WORKFLOW.yaml 2>/dev/null
    fi

    # Restore AGENT_MANIFEST.json
    if [[ -f "$SWAP_BACKUP_DIR/AGENT_MANIFEST.json" ]]; then
        cp "$SWAP_BACKUP_DIR/AGENT_MANIFEST.json" "$MANIFEST_FILE"
        log_debug "Restored manifest"
    else
        rm -f "$MANIFEST_FILE" 2>/dev/null
    fi

    # Restore ACTIVE_TEAM (LAST - this is the rollback commit point)
    if [[ -f "$SWAP_BACKUP_DIR/ACTIVE_TEAM" ]]; then
        cp "$SWAP_BACKUP_DIR/ACTIVE_TEAM" .claude/ACTIVE_TEAM
        log_debug "Restored ACTIVE_TEAM"
    else
        # Virgin swap had no ACTIVE_TEAM - remove it
        rm -f .claude/ACTIVE_TEAM
        log_debug "Removed ACTIVE_TEAM (virgin swap rollback)"
    fi

    # Restore commands if backed up
    if [[ -d "$SWAP_BACKUP_DIR/commands" ]]; then
        # Remove current team commands
        if [[ -f ".claude/commands/.team-commands" ]]; then
            while IFS= read -r cmd_file; do
                [[ -z "$cmd_file" ]] && continue
                rm -f ".claude/commands/$cmd_file"
            done < ".claude/commands/.team-commands"
        fi

        # Restore backed up commands
        for cmd_file in "$SWAP_BACKUP_DIR/commands"/*.md; do
            [[ -f "$cmd_file" ]] || continue
            cp "$cmd_file" ".claude/commands/$(basename "$cmd_file")"
        done
        if [[ -f "$SWAP_BACKUP_DIR/commands/.team-commands" ]]; then
            cp "$SWAP_BACKUP_DIR/commands/.team-commands" ".claude/commands/.team-commands"
        fi
        log_debug "Restored commands"
    fi

    # Restore skills if backed up
    if [[ -d "$SWAP_BACKUP_DIR/skills" ]]; then
        # Remove current team skills
        if [[ -f ".claude/skills/.team-skills" ]]; then
            while IFS= read -r skill_dir; do
                [[ -z "$skill_dir" ]] && continue
                rm -rf ".claude/skills/$skill_dir"
            done < ".claude/skills/.team-skills"
        fi

        # Restore backed up skills
        for skill_path in "$SWAP_BACKUP_DIR/skills"/*/; do
            [[ -d "$skill_path" ]] || continue
            local skill_name
            skill_name=$(basename "$skill_path")
            [[ "$skill_name" == "." ]] && continue
            cp -rp "$skill_path" ".claude/skills/$skill_name"
        done
        if [[ -f "$SWAP_BACKUP_DIR/skills/.team-skills" ]]; then
            cp "$SWAP_BACKUP_DIR/skills/.team-skills" ".claude/skills/.team-skills"
        fi
        log_debug "Restored skills"
    fi

    # Restore hooks if backed up
    if [[ -d "$SWAP_BACKUP_DIR/hooks" ]]; then
        # Remove current team hooks
        if [[ -f ".claude/hooks/.team-hooks" ]]; then
            while IFS= read -r hook_file; do
                [[ -z "$hook_file" ]] && continue
                rm -f ".claude/hooks/$hook_file"
            done < ".claude/hooks/.team-hooks"
        fi

        # Restore backed up hooks
        for hook_file in "$SWAP_BACKUP_DIR/hooks"/*; do
            [[ -f "$hook_file" ]] || continue
            local hook_name
            hook_name=$(basename "$hook_file")
            [[ "$hook_name" == ".team-hooks" ]] && continue
            cp "$hook_file" ".claude/hooks/$hook_name"
        done
        if [[ -f "$SWAP_BACKUP_DIR/hooks/.team-hooks" ]]; then
            cp "$SWAP_BACKUP_DIR/hooks/.team-hooks" ".claude/hooks/.team-hooks"
        fi
        log_debug "Restored hooks"
    fi

    # Cleanup staging and backup
    cleanup_staging
    cleanup_swap_backup
    delete_journal

    log "Rollback complete. Previous state restored."
    return 0
}

# ============================================================================
# Signal Handling
# ============================================================================

# Global flag to prevent re-entrant signal handling
SIGNAL_HANDLING=0

# Handle interrupt signals (SIGTERM, SIGINT, SIGHUP)
handle_interrupt() {
    if [[ "$SIGNAL_HANDLING" -eq 1 ]]; then
        return
    fi
    SIGNAL_HANDLING=1

    log_warning "Swap interrupted by signal"

    # Check current phase from journal
    local phase
    phase=$(get_journal_phase)

    case "$phase" in
        "$PHASE_PREPARING"|"$PHASE_BACKING"|"")
            log "No changes made. Exiting."
            cleanup_staging
            delete_journal
            ;;
        "$PHASE_STAGING"|"$PHASE_VERIFYING")
            log "Rolling back partial changes..."
            rollback_swap
            ;;
        "$PHASE_COMMITTING")
            log_warning "Interrupted during commit - state may be inconsistent"
            log "Run with --recover to check and restore state"
            # Don't delete journal - needed for recovery
            ;;
        "$PHASE_COMPLETED")
            log "Swap was completed. Cleaning up."
            cleanup_staging
            cleanup_swap_backup
            delete_journal
            ;;
    esac

    exit "$EXIT_SWAP_FAILURE"
}

# Handle normal exit
handle_exit() {
    # Cleanup temp files on normal exit
    rm -f ".claude/.swap-journal.tmp" 2>/dev/null
    rm -f "$STAGING_DIR"/*.tmp.$$ 2>/dev/null
}

# Set up signal handlers
setup_signal_handlers() {
    trap 'handle_interrupt' SIGINT SIGTERM SIGHUP
    trap 'handle_exit' EXIT
}

# ============================================================================
# Recovery Functions
# ============================================================================

# Check for journal on startup and handle recovery
# Returns: 0 if ok to proceed, 1 if recovery needed
check_journal_recovery() {
    if [[ ! -f "$JOURNAL_FILE" ]]; then
        return 0  # No recovery needed
    fi

    log_warning "Incomplete swap detected from previous run"

    # Read journal details
    local phase source target
    phase=$(jq -r '.phase // "unknown"' "$JOURNAL_FILE" 2>/dev/null)
    source=$(jq -r '.source_team // "none"' "$JOURNAL_FILE" 2>/dev/null)
    target=$(jq -r '.target_team // "unknown"' "$JOURNAL_FILE" 2>/dev/null)

    log "  Previous: $source -> $target"
    log "  Phase: $phase"

    # Check if swap completed but cleanup didn't finish
    if [[ "$phase" == "$PHASE_COMPLETED" ]]; then
        log "Previous swap completed. Cleaning up..."
        cleanup_staging
        cleanup_swap_backup
        delete_journal
        return 0
    fi

    # Check backup validity
    local backup_valid=false
    if verify_backup_integrity; then
        backup_valid=true
    fi

    if [[ "$backup_valid" == "false" ]]; then
        log_error "Backup is missing or corrupted"
        log "Manual intervention required:"
        log "  1. Check $STAGING_DIR for new state"
        log "  2. Check $SWAP_BACKUP_DIR for old state"
        log "  3. Manually restore desired state"
        log "  4. Delete $JOURNAL_FILE when done"
        exit "$EXIT_RECOVERY_REQUIRED"
    fi

    # Interactive vs non-interactive recovery
    if [[ -t 0 ]]; then
        prompt_recovery_action "$source" "$target" "$phase"
    elif [[ "$AUTO_RECOVER" -eq 1 ]]; then
        log "Auto-recovery enabled. Rolling back..."
        rollback_swap
        return 0
    else
        log_error "Non-interactive mode. Use --auto-recover to enable automatic rollback"
        log "Or manually resolve with: swap-team.sh --recover"
        exit "$EXIT_RECOVERY_REQUIRED"
    fi

    return 0
}

# Interactive recovery prompt
prompt_recovery_action() {
    local source="$1" target="$2" phase="$3"

    echo ""
    echo "Recovery Options:"
    echo "  [r] Rollback to previous state ($source)"
    echo "  [c] Continue swap to $target (may fail)"
    echo "  [a] Abort (leave as-is for manual recovery)"
    echo ""

    local choice
    while true; do
        read -r -p "Choice [r/c/a]: " choice < /dev/tty
        case "$choice" in
            r|R)
                rollback_swap
                return 0
                ;;
            c|C)
                continue_interrupted_swap "$phase" "$target"
                return $?
                ;;
            a|A)
                log "Aborted. Journal preserved for manual recovery."
                exit "$EXIT_SUCCESS"
                ;;
            *)
                echo "Invalid choice. Enter r, c, or a."
                ;;
        esac
    done
}

# Attempt to continue an interrupted swap
continue_interrupted_swap() {
    local phase="$1"
    local target_team="$2"

    log "Attempting to continue swap from phase: $phase"

    case "$phase" in
        "$PHASE_STAGING")
            # Need to re-stage and continue
            log_warning "Re-staging is not fully implemented. Recommend rollback."
            return 1
            ;;
        "$PHASE_VERIFYING"|"$PHASE_COMMITTING")
            # Try to complete the commit
            if [[ -d "$STAGING_DIR" ]]; then
                log "Attempting to complete commit from staging..."
                commit_staged_resources "$target_team"
                return $?
            else
                log_error "Staging directory missing. Cannot continue."
                return 1
            fi
            ;;
        *)
            log_error "Cannot continue from phase: $phase"
            return 1
            ;;
    esac
}

# Verify current state consistency
verify_state_consistency() {
    local errors=0

    log "Verifying state consistency..."

    # Check ACTIVE_TEAM exists
    if [[ ! -f ".claude/ACTIVE_TEAM" ]]; then
        log_warning "No ACTIVE_TEAM file (virgin state or corrupted)"
        ((errors++)) || true
    else
        local active_team
        active_team=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')

        # Check agents directory exists
        if [[ ! -d ".claude/agents" ]]; then
            log_error "ACTIVE_TEAM is $active_team but no agents directory exists"
            ((errors++)) || true
        else
            local agent_count
            agent_count=$(find .claude/agents -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
            if [[ "$agent_count" -eq 0 ]]; then
                log_error "ACTIVE_TEAM is $active_team but no agent files found"
                ((errors++)) || true
            fi
        fi

        # Check manifest matches ACTIVE_TEAM
        if [[ -f "$MANIFEST_FILE" ]]; then
            local manifest_team
            manifest_team=$(jq -r '.active_team // "unknown"' "$MANIFEST_FILE" 2>/dev/null)
            if [[ "$manifest_team" != "$active_team" ]]; then
                log_error "ACTIVE_TEAM ($active_team) does not match manifest ($manifest_team)"
                ((errors++)) || true
            fi
        fi

        # Check team pack exists in roster
        if [[ ! -d "$ROSTER_HOME/teams/$active_team" ]]; then
            log_warning "Active team $active_team not found in roster (may be orphaned)"
        fi
    fi

    # Check for orphaned journal
    if [[ -f "$JOURNAL_FILE" ]]; then
        log_warning "Swap journal exists - incomplete swap detected"
        local phase
        phase=$(get_journal_phase)
        log "  Phase: $phase"
        ((errors++)) || true
    fi

    # Check for orphaned staging
    if [[ -d "$STAGING_DIR" ]]; then
        log_warning "Staging directory exists - incomplete swap detected"
        ((errors++)) || true
    fi

    if [[ "$errors" -eq 0 ]]; then
        log "State is consistent"
        return 0
    else
        log_error "Found $errors consistency issue(s)"
        return 1
    fi
}

# ============================================================================
# Commit Functions
# ============================================================================

# Commit staged resources to live directories
# This is the atomic commit phase
# Usage: commit_staged_resources "team_name"
commit_staged_resources() {
    local team_name="$1"

    log_debug "Committing staged resources..."

    update_journal_phase "$PHASE_COMMITTING"

    # 1. Remove live directories (fast operations)
    rm -rf .claude/agents 2>/dev/null

    # 2. Move staged agents to live (atomic rename on same filesystem)
    if [[ -d "$STAGING_DIR/agents" ]]; then
        mv "$STAGING_DIR/agents" .claude/agents || {
            log_error "Failed to commit agents"
            update_journal_error "Failed to commit agents"
            return 1
        }
    fi

    # 3. Move workflow atomically
    if [[ -f "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" ]]; then
        mv "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" .claude/ACTIVE_WORKFLOW.yaml || {
            log_warning "Failed to commit workflow"
        }
    fi

    # Note: ACTIVE_TEAM is committed LAST (after manifest is written)
    # This happens in the main perform_swap flow

    log_debug "Core resources committed"
    return 0
}

# ============================================================================
# Library Imports
# ============================================================================

# Source roster utilities for dynamic roster generation
source "$ROSTER_HOME/lib/roster-utils.sh"

# ============================================================================
# Manifest Functions
# ============================================================================

# Read manifest and return JSON or empty if not exists
read_manifest() {
    if [[ -f "$MANIFEST_FILE" ]]; then
        cat "$MANIFEST_FILE"
    else
        echo ""
    fi
}

# Get agent info from manifest
# Usage: get_agent_from_manifest "agent-name.md"
# Returns: "source:origin" or empty if not found
get_agent_from_manifest() {
    local agent_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        echo ""
        return
    fi

    # Extract the line containing this agent (handles single-line JSON format)
    local agent_line
    agent_line=$(echo "$manifest" | grep "\"$agent_name\":")

    if [[ -z "$agent_line" ]]; then
        echo ""
        return
    fi

    # Extract source and origin from the single line
    local source origin
    source=$(echo "$agent_line" | sed 's/.*"source": *"\([^"]*\)".*/\1/')
    origin=$(echo "$agent_line" | sed 's/.*"origin": *"\([^"]*\)".*/\1/')

    echo "$source:$origin"
}

# Write manifest with current agent state
# Usage: write_manifest "team-name"
write_manifest() {
    local team_name="$1"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local manifest_dir
    manifest_dir=$(dirname "$MANIFEST_FILE")
    mkdir -p "$manifest_dir"

    # Start JSON
    {
        echo "{"
        echo "  \"manifest_version\": \"$MANIFEST_VERSION\","
        echo "  \"active_team\": \"$team_name\","
        echo "  \"last_swap\": \"$timestamp\","
        echo "  \"agents\": {"
    } > "$MANIFEST_FILE"

    # Add each agent from .claude/agents/
    local first=true
    if [[ -d ".claude/agents" ]]; then
        for agent_file in .claude/agents/*.md; do
            [[ ! -f "$agent_file" ]] && continue

            local agent_name
            agent_name=$(basename "$agent_file")

            # Determine source: check if it came from the team or is user-added
            local source="team"
            local origin="$team_name"

            # Check if this was a kept user agent (marked by stash)
            if [[ -f ".claude/.agent_stash/$agent_name.meta" ]]; then
                source=$(cat ".claude/.agent_stash/$agent_name.meta" | grep "source=" | cut -d= -f2)
                origin=$(cat ".claude/.agent_stash/$agent_name.meta" | grep "origin=" | cut -d= -f2)
                rm -f ".claude/.agent_stash/$agent_name.meta"
            fi

            # Add comma separator
            if [[ "$first" == true ]]; then
                first=false
            else
                echo "," >> "$MANIFEST_FILE"
            fi

            # Write agent entry
            {
                echo -n "    \"$agent_name\": {"
                echo -n "\"source\": \"$source\", "
                echo -n "\"origin\": \"$origin\", "
                echo -n "\"installed_at\": \"$timestamp\""
                echo -n "}"
            } >> "$MANIFEST_FILE"
        done
    fi

    # Close agents section, add comma for commands
    echo "" >> "$MANIFEST_FILE"
    echo "  }," >> "$MANIFEST_FILE"

    # Add commands section
    echo "  \"commands\": {" >> "$MANIFEST_FILE"

    # Read team commands from marker file
    local first_cmd=true
    if [[ -f ".claude/commands/.team-commands" ]]; then
        while IFS= read -r cmd_file; do
            [[ -z "$cmd_file" ]] && continue

            if [[ "$first_cmd" == true ]]; then
                first_cmd=false
            else
                echo "," >> "$MANIFEST_FILE"
            fi

            {
                echo -n "    \"$cmd_file\": {"
                echo -n "\"source\": \"team\", "
                echo -n "\"origin\": \"$team_name\", "
                echo -n "\"installed_at\": \"$timestamp\""
                echo -n "}"
            } >> "$MANIFEST_FILE"
        done < ".claude/commands/.team-commands"
    fi

    # Close commands section, add comma for hooks
    echo "" >> "$MANIFEST_FILE"
    echo "  }," >> "$MANIFEST_FILE"

    # Add hooks section
    echo "  \"hooks\": {" >> "$MANIFEST_FILE"

    local first_hook=true

    # Track hooks (base from user-hooks, team from .team-hooks marker)
    if [[ -d ".claude/hooks" ]]; then
        for hook_file in .claude/hooks/*.sh; do
            [[ ! -f "$hook_file" ]] && continue

            local hook_name
            hook_name=$(basename "$hook_file")

            # Determine source: team if in marker, base otherwise
            local source="base"
            local origin="user-hooks"

            if [[ -f ".claude/hooks/.team-hooks" ]] && grep -q "^$hook_name$" ".claude/hooks/.team-hooks" 2>/dev/null; then
                source="team"
                origin="$team_name"
            fi

            if [[ "$first_hook" == true ]]; then
                first_hook=false
            else
                echo "," >> "$MANIFEST_FILE"
            fi

            {
                echo -n "    \"$hook_name\": {"
                echo -n "\"source\": \"$source\", "
                echo -n "\"origin\": \"$origin\", "
                echo -n "\"installed_at\": \"$timestamp\""
                echo -n "}"
            } >> "$MANIFEST_FILE"
        done
    fi

    # Close hooks and JSON
    {
        echo ""
        echo "  }"
        echo "}"
    } >> "$MANIFEST_FILE"

    log_debug "Manifest written: $MANIFEST_FILE"
}

# Initialize manifest for first-time use (treats existing agents as unknown/user)
init_manifest_from_existing() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local manifest_dir
    manifest_dir=$(dirname "$MANIFEST_FILE")
    mkdir -p "$manifest_dir"

    # Start JSON
    {
        echo "{"
        echo "  \"manifest_version\": \"$MANIFEST_VERSION\","
        echo "  \"active_team\": \"unknown\","
        echo "  \"last_swap\": \"$timestamp\","
        echo "  \"agents\": {"
    } > "$MANIFEST_FILE"

    # Add each existing agent as "unknown" source
    local first=true
    if [[ -d ".claude/agents" ]]; then
        for agent_file in .claude/agents/*.md; do
            [[ ! -f "$agent_file" ]] && continue

            local agent_name
            agent_name=$(basename "$agent_file")

            # Add comma separator
            if [[ "$first" == true ]]; then
                first=false
            else
                echo "," >> "$MANIFEST_FILE"
            fi

            # Write agent entry as unknown source
            {
                echo -n "    \"$agent_name\": {"
                echo -n "\"source\": \"unknown\", "
                echo -n "\"origin\": \"unknown\", "
                echo -n "\"installed_at\": \"$timestamp\""
                echo -n "}"
            } >> "$MANIFEST_FILE"
        done
    fi

    # Close agents section, add comma for commands
    echo "" >> "$MANIFEST_FILE"
    echo "  }," >> "$MANIFEST_FILE"

    # Add empty commands section (no team commands during init)
    {
        echo "  \"commands\": {"
        echo "  }"
        echo "}"
    } >> "$MANIFEST_FILE"

    log_debug "Initialized manifest from existing agents"
}

# ============================================================================
# Orphan Detection Functions
# ============================================================================

# List agents in incoming team pack
# Usage: list_incoming_agents "team-name"
# Output: One agent filename per line
list_incoming_agents() {
    local team_name="$1"
    local pack_dir="$ROSTER_HOME/teams/$team_name/agents"

    if [[ -d "$pack_dir" ]]; then
        for agent_file in "$pack_dir"/*.md; do
            [[ -f "$agent_file" ]] && basename "$agent_file"
        done
    fi
}

# List current agents in project
# Usage: list_current_agents
# Output: One agent filename per line
list_current_agents() {
    if [[ -d ".claude/agents" ]]; then
        for agent_file in .claude/agents/*.md; do
            [[ -f "$agent_file" ]] && basename "$agent_file"
        done
    fi
}

# Detect orphan agents (current agents not in incoming team)
# Usage: detect_orphans "team-name"
# Sets ORPHAN_AGENTS array with orphan info: "agent.md:source:origin"
detect_orphans() {
    local team_name="$1"
    ORPHAN_AGENTS=()

    # Get list of incoming agents
    local -a incoming_agents=()
    while IFS= read -r agent; do
        [[ -n "$agent" ]] && incoming_agents+=("$agent")
    done < <(list_incoming_agents "$team_name")

    log_debug "Incoming agents from $team_name: ${incoming_agents[*]:-none}"

    # If no manifest exists and agents exist, initialize it
    if [[ ! -f "$MANIFEST_FILE" ]] && [[ -d ".claude/agents" ]]; then
        local agent_count
        agent_count=$(find .claude/agents -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        if [[ "$agent_count" -gt 0 ]]; then
            log_debug "No manifest found, initializing from existing agents"
            init_manifest_from_existing
        fi
    fi

    # Check each current agent
    while IFS= read -r agent; do
        [[ -z "$agent" ]] && continue

        # Is this agent in the incoming team?
        local is_incoming=false
        for inc_agent in "${incoming_agents[@]:-}"; do
            if [[ "$agent" == "$inc_agent" ]]; then
                is_incoming=true
                break
            fi
        done

        if [[ "$is_incoming" == false ]]; then
            # This is an orphan - get its provenance from manifest
            local info
            info=$(get_agent_from_manifest "$agent")

            if [[ -z "$info" ]]; then
                # Not in manifest - treat as unknown/user-added
                info="unknown:unknown"
            fi

            ORPHAN_AGENTS+=("$agent:$info")
            log_debug "Orphan detected: $agent ($info)"
        fi
    done < <(list_current_agents)

    log_debug "Total orphans: ${#ORPHAN_AGENTS[@]}"
}

# Format orphan for display
# Usage: format_orphan "agent.md:source:origin"
format_orphan() {
    local orphan="$1"
    local agent source origin

    agent=$(echo "$orphan" | cut -d: -f1)
    source=$(echo "$orphan" | cut -d: -f2)
    origin=$(echo "$orphan" | cut -d: -f3)

    case "$source" in
        "user")
            echo "$agent (user-added)"
            ;;
        "team")
            echo "$agent (from $origin)"
            ;;
        "unknown")
            echo "$agent (unknown origin)"
            ;;
        *)
            echo "$agent"
            ;;
    esac
}

# ============================================================================
# Interactive Disposition Functions
# ============================================================================

# Global arrays for disposition decisions
declare -a AGENTS_TO_KEEP=()
declare -a AGENTS_TO_PROMOTE=()
declare -a AGENTS_TO_REMOVE=()
declare -a ORPHAN_SKILLS=()
declare -a ORPHAN_COMMANDS=()
declare -a ORPHAN_HOOKS=()

# Stash agents to keep (before swap clears .claude/agents/)
stash_kept_agents() {
    local stash_dir=".claude/.agent_stash"

    if [[ ${#AGENTS_TO_KEEP[@]} -eq 0 ]]; then
        return 0
    fi

    mkdir -p "$stash_dir"
    log_debug "Stashing ${#AGENTS_TO_KEEP[@]} agent(s) for preservation"

    for entry in "${AGENTS_TO_KEEP[@]}"; do
        local agent source origin
        agent=$(echo "$entry" | cut -d: -f1)
        source=$(echo "$entry" | cut -d: -f2)
        origin=$(echo "$entry" | cut -d: -f3)

        if [[ -f ".claude/agents/$agent" ]]; then
            cp ".claude/agents/$agent" "$stash_dir/$agent"
            # Save metadata for manifest reconstruction
            echo "source=$source" > "$stash_dir/$agent.meta"
            echo "origin=$origin" >> "$stash_dir/$agent.meta"
            log_debug "Stashed: $agent"
        fi
    done
}

# Restore stashed agents after swap
restore_kept_agents() {
    local stash_dir=".claude/.agent_stash"

    if [[ ! -d "$stash_dir" ]]; then
        return 0
    fi

    local restored=0
    for agent_file in "$stash_dir"/*.md; do
        [[ ! -f "$agent_file" ]] && continue

        local agent
        agent=$(basename "$agent_file")

        cp "$agent_file" ".claude/agents/$agent"
        log_debug "Restored: $agent"
        ((restored++)) || true
    done

    if [[ "$restored" -gt 0 ]]; then
        log "Kept: $restored agent(s) preserved"
    fi
}

# Promote agents to user-level (~/.claude/agents/)
promote_agents() {
    if [[ ${#AGENTS_TO_PROMOTE[@]} -eq 0 ]]; then
        return 0
    fi

    local user_agents_dir="$HOME/.claude/agents"
    mkdir -p "$user_agents_dir"

    local promoted=0
    for entry in "${AGENTS_TO_PROMOTE[@]}"; do
        local agent
        agent=$(echo "$entry" | cut -d: -f1)

        if [[ -f ".claude/agents/$agent" ]]; then
            if [[ -f "$user_agents_dir/$agent" ]]; then
                log_warning "Skipped promote: $agent already exists in ~/.claude/agents/"
            else
                cp ".claude/agents/$agent" "$user_agents_dir/$agent"
                log_debug "Promoted: $agent → ~/.claude/agents/"
                ((promoted++)) || true
            fi
        fi
    done

    if [[ "$promoted" -gt 0 ]]; then
        log "Promoted: $promoted agent(s) → ~/.claude/agents/"
    fi
}

# Clean up stash directory
cleanup_stash() {
    rm -rf ".claude/.agent_stash"
}

# Interactive prompt for orphan disposition
# Returns: 0 if user made choices, 1 if cancelled
prompt_disposition() {
    local team_name="$1"

    # Check if we're in an interactive terminal
    if [[ ! -t 0 ]]; then
        # Non-interactive - error if no flag set
        if [[ -z "$ORPHAN_MODE" ]]; then
            log_error "Orphan agents detected in non-interactive mode."
            log "Found ${#ORPHAN_AGENTS[@]} agent(s) not in $team_name:"
            for orphan in "${ORPHAN_AGENTS[@]}"; do
                echo "  - $(format_orphan "$orphan")"
            done
            log ""
            log "Use one of these flags:"
            log "  --keep-all     Preserve all orphans in project"
            log "  --remove-all   Remove all orphans (backup available)"
            log "  --promote-all  Move all orphans to ~/.claude/agents/"
            exit "$EXIT_ORPHAN_CONFLICT"
        fi

        # Apply the flag to all orphans
        for orphan in "${ORPHAN_AGENTS[@]}"; do
            case "$ORPHAN_MODE" in
                "keep")
                    AGENTS_TO_KEEP+=("$orphan")
                    ;;
                "remove")
                    AGENTS_TO_REMOVE+=("$orphan")
                    ;;
                "promote")
                    AGENTS_TO_PROMOTE+=("$orphan")
                    ;;
            esac
        done
        return 0
    fi

    # Interactive mode
    local current_team="unknown"
    if [[ -f ".claude/ACTIVE_TEAM" ]]; then
        current_team=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')
    fi

    echo ""
    echo -e "${YELLOW}[Roster]${NC} Switching from $current_team to $team_name..."
    echo -e "${YELLOW}[Roster]${NC} Found ${#ORPHAN_AGENTS[@]} agent(s) not in $team_name:"
    echo ""

    local idx=1
    for orphan in "${ORPHAN_AGENTS[@]}"; do
        echo "  $idx. $(format_orphan "$orphan")"
        ((idx++))
    done

    echo ""
    echo "For each agent, choose:"
    echo "  [k] Keep in project    [p] Promote to ~/.claude/agents/"
    echo "  [r] Remove             [a] Apply to all remaining"
    echo ""

    local apply_all=""
    for orphan in "${ORPHAN_AGENTS[@]}"; do
        local agent
        agent=$(echo "$orphan" | cut -d: -f1)

        if [[ -n "$apply_all" ]]; then
            # Apply previous "all" choice
            case "$apply_all" in
                "k") AGENTS_TO_KEEP+=("$orphan") ;;
                "p") AGENTS_TO_PROMOTE+=("$orphan") ;;
                "r") AGENTS_TO_REMOVE+=("$orphan") ;;
            esac
            continue
        fi

        # Prompt for this agent
        local choice=""
        while true; do
            echo -n "$agent [k/p/r/a]: "
            read -r choice < /dev/tty

            case "$choice" in
                k|K)
                    AGENTS_TO_KEEP+=("$orphan")
                    break
                    ;;
                p|P)
                    AGENTS_TO_PROMOTE+=("$orphan")
                    break
                    ;;
                r|R)
                    AGENTS_TO_REMOVE+=("$orphan")
                    break
                    ;;
                a|A)
                    # Ask what action to apply to all
                    echo -n "Apply which action to all remaining? [k/p/r]: "
                    read -r apply_choice < /dev/tty
                    case "$apply_choice" in
                        k|K)
                            apply_all="k"
                            AGENTS_TO_KEEP+=("$orphan")
                            ;;
                        p|P)
                            apply_all="p"
                            AGENTS_TO_PROMOTE+=("$orphan")
                            ;;
                        r|R)
                            apply_all="r"
                            AGENTS_TO_REMOVE+=("$orphan")
                            ;;
                        *)
                            echo "Invalid choice. Please enter k, p, or r."
                            continue
                            ;;
                    esac
                    break
                    ;;
                "")
                    # Ctrl+C or empty - abort
                    echo ""
                    log "Swap cancelled by user"
                    exit "$EXIT_SUCCESS"
                    ;;
                *)
                    echo "Invalid choice. Please enter k, p, r, or a."
                    ;;
            esac
        done
    done

    echo ""
    return 0
}

# Usage information
usage() {
    cat <<EOF
Usage: swap-team.sh [OPTIONS] [COMMAND]

Commands:
  <pack-name>    Switch to specified team pack
  --list         List all available team packs
  --reset        Reset to skeleton baseline (remove all team resources)
  --verify       Verify current state consistency
  --recover      Interactive recovery from interrupted swap
  (no args)      Show current active team

Options:
  --update, -u   Update agents from roster (even if already on team)
  --refresh, -r  [DEPRECATED] Alias for --update
  --dry-run      Preview changes without applying
  --keep-all     Preserve orphan agents in project
  --remove-all   Remove orphan agents/commands/skills/hooks (backup available)
  --promote-all  Move orphan agents to ~/.claude/agents/
  --auto-recover Automatically rollback if interrupted swap detected (for CI/CD)

When switching teams interactively, you'll be prompted for each orphan agent
(agents in current team but not in target team). In non-interactive mode
(scripts, CI), you must specify one of the orphan handling flags.

Environment Variables:
  ROSTER_HOME         Roster repository location (default: ~/Code/roster)
  ROSTER_DEBUG        Enable debug logging (set to 1)
  ROSTER_AUTO_RECOVER Enable auto-recovery in non-interactive mode (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Validation failure (pack doesn't exist or is invalid)
  3  Backup failure
  4  Swap failure
  5  Orphan conflict (non-interactive without flag)
  6  Recovery required (interrupted swap, manual intervention needed)

Examples:
  ./swap-team.sh dev-pack               # Switch to dev-pack (interactive prompts)
  ./swap-team.sh                        # Show current team
  ./swap-team.sh --list                 # List available teams
  ./swap-team.sh dev-pack --keep-all    # Keep all orphans during swap
  ./swap-team.sh dev-pack --remove-all  # Remove all orphans during swap
  ./swap-team.sh --update               # Update current team from roster
  ./swap-team.sh dev-pack --update      # Update even if already on dev-pack
  ./swap-team.sh --update --dry-run     # Preview what update would change
  ./swap-team.sh --reset                # Reset to skeleton baseline
  ./swap-team.sh --reset --dry-run      # Preview what reset would remove
  ./swap-team.sh --verify               # Check state consistency
  ./swap-team.sh --recover              # Recover from interrupted swap
  ./swap-team.sh --auto-recover dev-pack # CI/CD mode with auto-rollback

EOF
}

# Known valid Claude Code tools
# REQ-3.3: Tool validation for agent declarations
VALID_TOOLS="Bash Glob Grep Read Write Edit WebFetch WebSearch TodoWrite Task Skill NotebookEdit AskUserQuestion"

# Validate tools field in agent frontmatter
# Returns 0 if valid or no tools field, 1 if invalid tools found
validate_agent_tools() {
    local agent_file="$1"
    local agent_name
    agent_name=$(basename "$agent_file" .md)

    # Extract tools field from YAML frontmatter
    local tools_line
    tools_line=$(sed -n '/^---$/,/^---$/p' "$agent_file" | grep "^tools:" | head -1)

    if [[ -z "$tools_line" ]]; then
        # No tools field - acceptable (agent won't have tool restrictions)
        return 0
    fi

    # Parse comma-separated tools (POSIX-compatible approach)
    local tools
    tools=$(echo "$tools_line" | sed 's/^tools:[[:space:]]*//')

    local invalid_tools=""

    # Split on comma and space, process each tool
    echo "$tools" | tr ',' '\n' | while read -r tool; do
        # Trim whitespace
        tool=$(echo "$tool" | tr -d '[:space:]')

        # Skip empty entries
        [[ -z "$tool" ]] && continue

        if [[ ! " $VALID_TOOLS " =~ " $tool " ]]; then
            # Use a temp file to communicate from subshell
            echo "$tool" >> /tmp/.swap-team-invalid-tools-$$
        fi
    done

    # Check if any invalid tools were found
    if [[ -f /tmp/.swap-team-invalid-tools-$$ ]]; then
        invalid_tools=$(cat /tmp/.swap-team-invalid-tools-$$)
        rm -f /tmp/.swap-team-invalid-tools-$$
        log_warning "Invalid tools in ${agent_name}.md: $invalid_tools"
        return 1
    fi

    return 0
}

# Validate all agent tools in a team pack
# Returns count of agents with invalid tools (0 = all valid)
validate_pack_tools() {
    local pack_dir="$1"
    local invalid_count=0

    for agent_file in "$pack_dir/agents"/*.md; do
        [[ -f "$agent_file" ]] || continue
        if ! validate_agent_tools "$agent_file"; then
            ((invalid_count++)) || true
        fi
    done

    echo "$invalid_count"
}

# Validate team pack exists and has required structure
validate_pack() {
    local team_name="$1"
    local pack_dir="$ROSTER_HOME/teams/$team_name"

    log_debug "Validating pack: $team_name"

    # Check pack directory exists
    if [[ ! -d "$pack_dir" ]]; then
        log_error "Team pack '$team_name' not found in $ROSTER_HOME/teams/"
        log "Use './swap-team.sh --list' to see available packs"
        exit "$EXIT_VALIDATION_FAILURE"
    fi

    # Check agents/ subdirectory exists
    if [[ ! -d "$pack_dir/agents" ]]; then
        log_error "Team pack '$team_name' missing agents/ directory"
        exit "$EXIT_VALIDATION_FAILURE"
    fi

    # Check at least one .md file exists
    local agent_count
    agent_count=$(find "$pack_dir/agents" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$agent_count" -eq 0 ]]; then
        log_error "Team pack '$team_name' has no agent files (.md)"
        exit "$EXIT_VALIDATION_FAILURE"
    fi

    # Check workflow.yaml exists
    if [[ ! -f "$pack_dir/workflow.yaml" ]]; then
        log_warning "Team pack '$team_name' missing workflow.yaml (commands may fail)"
    fi

    # Check for missing directories (REQ-3.4)
    if [[ ! -d "$pack_dir/commands" ]]; then
        log_warning "Team pack '$team_name' missing commands/ (run normalize-team-structure.sh)"
    fi
    if [[ ! -d "$pack_dir/skills" ]]; then
        log_warning "Team pack '$team_name' missing skills/ (run normalize-team-structure.sh)"
    fi

    # Validate agent tools (REQ-3.3)
    local invalid_tool_count
    invalid_tool_count=$(validate_pack_tools "$pack_dir")
    if [[ "$invalid_tool_count" -gt 0 ]]; then
        log_warning "$invalid_tool_count agent(s) have invalid tools declarations"
    fi

    log_debug "Pack validation passed: $agent_count agents found"
    echo "$agent_count"
}

# Validate project has .claude/ directory and is writable
validate_project() {
    log_debug "Validating project environment"

    # Create .claude/ if it doesn't exist
    if [[ ! -d ".claude" ]]; then
        log_debug "Creating .claude/ directory"
        mkdir -p .claude || {
            log_error "Cannot create .claude/ directory (permissions?)"
            exit "$EXIT_BACKUP_FAILURE"
        }
    fi

    # Check .claude/ is writable
    if [[ ! -w ".claude" ]]; then
        log_error ".claude/ is not writable (permissions?)"
        exit "$EXIT_BACKUP_FAILURE"
    fi

    # Check available disk space (at least 1MB)
    local available
    available=$(df -k .claude 2>/dev/null | tail -1 | awk '{print $4}')

    if [[ -n "$available" ]] && [[ "$available" -lt 1024 ]]; then
        log_error "Insufficient disk space (< 1MB free)"
        exit "$EXIT_SWAP_FAILURE"
    fi

    log_debug "Project validation passed"
}

# Query current active team
query_current_team() {
    log_debug "Querying current team"

    if [[ ! -f ".claude/ACTIVE_TEAM" ]]; then
        log "No team active (virgin project)"
        exit "$EXIT_SUCCESS"
    fi

    local current
    current=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')

    if [[ -z "$current" ]]; then
        log_error "ACTIVE_TEAM file is empty (undefined state)"
        exit "$EXIT_INVALID_ARGS"
    fi

    # Check if team still exists in roster
    if [[ ! -d "$ROSTER_HOME/teams/$current" ]]; then
        log_warning "Active team '$current' not found in roster (orphaned state)"
        log "Consider swapping to a valid team"
    else
        log "Active team: $current"
    fi

    exit "$EXIT_SUCCESS"
}

# List all available team packs
list_teams() {
    log_debug "Listing available teams"

    local teams_dir="$ROSTER_HOME/teams"

    if [[ ! -d "$teams_dir" ]]; then
        log_error "Roster teams directory not found: $teams_dir"
        exit "$EXIT_VALIDATION_FAILURE"
    fi

    local teams=()
    local has_teams=false

    # Find all directories with agents/ subdirectory
    while IFS= read -r -d '' pack_dir; do
        local pack_name
        pack_name=$(basename "$pack_dir")

        # Validate pack has agents/
        if [[ -d "$pack_dir/agents" ]]; then
            local agent_count
            agent_count=$(find "$pack_dir/agents" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

            if [[ "$agent_count" -gt 0 ]]; then
                teams+=("$pack_name")
                has_teams=true
            fi
        fi
    done < <(find "$teams_dir" -mindepth 1 -maxdepth 1 -type d -print0 2>/dev/null)

    if [[ "$has_teams" == false ]]; then
        log "No teams available"
        exit "$EXIT_SUCCESS"
    fi

    log "Available teams:"
    for team in "${teams[@]}"; do
        echo "  - $team"
    done

    exit "$EXIT_SUCCESS"
}

# Backup current agents
backup_current_agents() {
    log_debug "Starting backup phase"

    local backup_dir=".claude/agents.backup"

    # If no agents exist yet, skip backup (virgin swap)
    if [[ ! -d ".claude/agents" ]] || [[ -z "$(ls -A .claude/agents 2>/dev/null)" ]]; then
        log "No agents to back up (virgin swap)"
        return 0
    fi

    # Remove old backup if exists
    if [[ -d "$backup_dir" ]]; then
        log_debug "Removing old backup"
        rm -rf "$backup_dir" || {
            log_error "Failed to remove old backup"
            exit "$EXIT_BACKUP_FAILURE"
        }
    fi

    # Copy current agents to backup
    log_debug "Copying agents to backup"
    cp -rp .claude/agents "$backup_dir" || {
        log_error "Backup failed (disk full? permissions?)"
        exit "$EXIT_BACKUP_FAILURE"
    }

    log "Backed up current agents to $backup_dir/"
}

# Perform the agent swap
swap_agents() {
    local team_name="$1"
    local agent_count="$2"
    local source_dir="$ROSTER_HOME/teams/$team_name/agents"

    log_debug "Starting swap phase"

    # Clear old agents
    if [[ -d ".claude/agents" ]]; then
        log_debug "Removing old agents"
        rm -rf .claude/agents || {
            log_error "Failed to remove old agents"
            exit "$EXIT_SWAP_FAILURE"
        }
    fi

    # Copy new agents
    log_debug "Copying new agents from $source_dir"
    cp -rp "$source_dir" .claude/agents || {
        log_error "Failed to copy new agents (disk full? permissions?)"
        log "Restore from backup: cp -r .claude/agents.backup/* .claude/agents/"
        exit "$EXIT_SWAP_FAILURE"
    }

    # Check for same-name conflicts with user-level agents
    local user_agents_dir="$HOME/.claude/agents"
    if [[ -d "$user_agents_dir" ]]; then
        for agent_file in .claude/agents/*.md; do
            [[ -f "$agent_file" ]] || continue
            local agent_name
            agent_name=$(basename "$agent_file")
            if [[ -f "$user_agents_dir/$agent_name" ]]; then
                log_warning "Team agent '$agent_name' shadows user-level agent in ~/.claude/agents/"
            fi
        done
    fi

    # Verify copy completed
    local dest_count
    dest_count=$(find .claude/agents -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$dest_count" -ne "$agent_count" ]]; then
        log_error "File count mismatch (expected $agent_count, got $dest_count)"
        log "Restore from backup: cp -r .claude/agents.backup/* .claude/agents/"
        exit "$EXIT_SWAP_FAILURE"
    fi

    # Copy workflow.yaml if exists
    local workflow_file="$ROSTER_HOME/teams/$team_name/workflow.yaml"
    if [[ -f "$workflow_file" ]]; then
        log_debug "Copying workflow.yaml"
        cp "$workflow_file" .claude/ACTIVE_WORKFLOW.yaml || {
            log_warning "Failed to copy workflow.yaml (agents swapped successfully)"
        }
    fi

    log_debug "Swap phase completed successfully"
}

# DEPRECATED: preserve_global_agents() removed
# Global agents now live at ~/.claude/agents/ (user-level) and are loaded
# automatically by Claude Code. No need to copy them to project agents.
# See: cem install-user for user-level agent installation.

# ============================================================================
# Team Commands Functions
# ============================================================================

# Backup current team commands (if any exist)
backup_team_commands() {
    log_debug "Checking for team commands to backup"

    local backup_dir=".claude/commands.backup"

    # Check if any team commands exist (marked by .team-command file)
    if [[ ! -d ".claude/commands" ]] || [[ ! -f ".claude/commands/.team-commands" ]]; then
        log_debug "No team commands to backup"
        return 0
    fi

    # Remove old backup if exists
    if [[ -d "$backup_dir" ]]; then
        log_debug "Removing old commands backup"
        rm -rf "$backup_dir" || {
            log_warning "Failed to remove old commands backup"
        }
    fi

    # Read list of team commands and backup
    mkdir -p "$backup_dir"
    while IFS= read -r cmd_file; do
        [[ -z "$cmd_file" ]] && continue
        if [[ -f ".claude/commands/$cmd_file" ]]; then
            cp ".claude/commands/$cmd_file" "$backup_dir/$cmd_file"
            log_debug "Backed up command: $cmd_file"
        fi
    done < ".claude/commands/.team-commands"

    log_debug "Team commands backed up"
}

# Remove team commands from previous team
remove_team_commands() {
    log_debug "Removing team commands from previous team"

    if [[ ! -f ".claude/commands/.team-commands" ]]; then
        log_debug "No team commands marker found"
        return 0
    fi

    # Read list and remove each command
    while IFS= read -r cmd_file; do
        [[ -z "$cmd_file" ]] && continue
        if [[ -f ".claude/commands/$cmd_file" ]]; then
            rm -f ".claude/commands/$cmd_file"
            log_debug "Removed team command: $cmd_file"
        fi
    done < ".claude/commands/.team-commands"

    # Remove the marker file
    rm -f ".claude/commands/.team-commands"

    log_debug "Team commands removed"
}

# Check if a command belongs to ANY team pack
is_team_command() {
    local cmd_name="$1"
    find "$ROSTER_HOME/teams" -path "*/commands/$cmd_name" -type f 2>/dev/null | grep -q .
}

# Get which team a command belongs to
get_command_team() {
    local cmd_name="$1"
    local match
    match=$(find "$ROSTER_HOME/teams" -path "*/commands/$cmd_name" -type f 2>/dev/null | head -1)
    if [[ -n "$match" ]]; then
        echo "$match" | sed 's|.*/teams/\([^/]*\)/commands/.*|\1|'
    fi
}

# Detect orphan commands - commands from OTHER teams that shouldn't be here
detect_command_orphans() {
    local incoming_team="$1"
    local incoming_cmds_dir="$ROSTER_HOME/teams/$incoming_team/commands"

    ORPHAN_COMMANDS=()

    [[ -d ".claude/commands" ]] || return 0

    for cmd_path in .claude/commands/*.md; do
        [[ -f "$cmd_path" ]] || continue
        local cmd_name
        cmd_name=$(basename "$cmd_path")

        # Is this command in the incoming team?
        if [[ -f "$incoming_cmds_dir/$cmd_name" ]]; then
            continue
        fi

        # Is this command from ANY team pack?
        if is_team_command "$cmd_name"; then
            local origin_team
            origin_team=$(get_command_team "$cmd_name")
            ORPHAN_COMMANDS+=("$cmd_name:$origin_team")
            log_debug "Orphan command detected: $cmd_name (from $origin_team)"
        fi
    done

    log_debug "Total orphan commands: ${#ORPHAN_COMMANDS[@]}"
}

# Remove orphan commands based on ORPHAN_MODE
remove_orphan_commands() {
    if [[ ${#ORPHAN_COMMANDS[@]} -eq 0 ]]; then
        return 0
    fi

    local backup_dir=".claude/commands.orphan-backup"

    for orphan in "${ORPHAN_COMMANDS[@]}"; do
        local cmd_name origin_team
        cmd_name=$(echo "$orphan" | cut -d: -f1)
        origin_team=$(echo "$orphan" | cut -d: -f2)

        case "$ORPHAN_MODE" in
            "remove")
                mkdir -p "$backup_dir"
                if [[ -f ".claude/commands/$cmd_name" ]]; then
                    cp ".claude/commands/$cmd_name" "$backup_dir/$cmd_name"
                    rm ".claude/commands/$cmd_name"
                    log "Removed orphan command: $cmd_name (was from $origin_team)"
                fi
                ;;
            "keep")
                log "Keeping orphan command: $cmd_name (from $origin_team)"
                ;;
            *)
                log_debug "Keeping orphan command: $cmd_name (no disposition)"
                ;;
        esac
    done

    if [[ "$ORPHAN_MODE" == "remove" ]] && [[ -d "$backup_dir" ]]; then
        log "Orphan command backups saved to: $backup_dir"
    fi
}

# Sync team-specific commands to project
# Team commands are copied to .claude/commands/ with a marker file
swap_commands() {
    local team_name="$1"
    local source_dir="$ROSTER_HOME/teams/$team_name/commands"

    log_debug "Checking for team commands in $source_dir"

    # Ensure commands directory exists
    mkdir -p ".claude/commands"

    # Backup and remove previous team commands
    backup_team_commands
    remove_team_commands

    # Check if team has commands
    if [[ ! -d "$source_dir" ]]; then
        log_debug "Team $team_name has no commands/ directory"
        return 0
    fi

    local cmd_count
    cmd_count=$(find "$source_dir" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$cmd_count" -eq 0 ]]; then
        log_debug "Team $team_name has no command files"
        return 0
    fi

    log_debug "Syncing $cmd_count command(s) from $team_name"

    # Create marker file to track which commands belong to this team
    local marker_file=".claude/commands/.team-commands"
    : > "$marker_file"

    # Copy each command and record in marker
    for cmd_file in "$source_dir"/*.md; do
        [[ -f "$cmd_file" ]] || continue

        local cmd_name
        cmd_name=$(basename "$cmd_file")

        # Check for collision with existing project command
        if [[ -f ".claude/commands/$cmd_name" ]] && ! grep -q "^$cmd_name$" "$marker_file" 2>/dev/null; then
            # Not a team command, this is a project command - skip with warning
            log_warning "Skipped: $cmd_name (project command exists)"
            continue
        fi

        cp "$cmd_file" ".claude/commands/$cmd_name"
        echo "$cmd_name" >> "$marker_file"
        log_debug "Synced command: $cmd_name"
    done

    # Count successfully synced commands
    local synced_count
    synced_count=$(wc -l < "$marker_file" | tr -d ' ')

    if [[ "$synced_count" -gt 0 ]]; then
        log "Synced: $synced_count team command(s)"
    fi
}

# ============================================================================
# Team Skills Functions (Phase 2: Unified Sync)
# ============================================================================

# Backup current team skills (if any exist)
backup_team_skills() {
    log_debug "Checking for team skills to backup"

    local backup_dir=".claude/skills.backup"

    # Check if any team skills exist (marked by .team-skills file)
    if [[ ! -d ".claude/skills" ]] || [[ ! -f ".claude/skills/.team-skills" ]]; then
        log_debug "No team skills to backup"
        return 0
    fi

    # Remove old backup if exists
    if [[ -d "$backup_dir" ]]; then
        log_debug "Removing old skills backup"
        rm -rf "$backup_dir" || {
            log_warning "Failed to remove old skills backup"
        }
    fi

    # Read list of team skills and backup
    mkdir -p "$backup_dir"
    while IFS= read -r skill_dir; do
        [[ -z "$skill_dir" ]] && continue
        if [[ -d ".claude/skills/$skill_dir" ]]; then
            cp -rp ".claude/skills/$skill_dir" "$backup_dir/$skill_dir"
            log_debug "Backed up skill: $skill_dir"
        fi
    done < ".claude/skills/.team-skills"

    log_debug "Team skills backed up"
}

# Check if a skill belongs to ANY team pack (not skeleton)
is_team_skill() {
    local skill_name="$1"
    find "$ROSTER_HOME/teams" -path "*/skills/$skill_name" -type d 2>/dev/null | grep -q .
}

# Get which team a skill belongs to
get_skill_team() {
    local skill_name="$1"
    local match
    match=$(find "$ROSTER_HOME/teams" -path "*/skills/$skill_name" -type d 2>/dev/null | head -1)
    if [[ -n "$match" ]]; then
        echo "$match" | sed 's|.*/teams/\([^/]*\)/skills/.*|\1|'
    fi
}

# Detect orphan skills - skills from OTHER teams that shouldn't be here
# Usage: detect_skill_orphans "incoming-team-name"
detect_skill_orphans() {
    local incoming_team="$1"
    local incoming_skills_dir="$ROSTER_HOME/teams/$incoming_team/skills"

    ORPHAN_SKILLS=()

    # Check each skill in .claude/skills/
    for skill_path in .claude/skills/*/; do
        [[ -d "$skill_path" ]] || continue
        local skill_name
        skill_name=$(basename "$skill_path")

        # Skip hidden files/dirs
        [[ "$skill_name" == .* ]] && continue

        # Is this skill in the incoming team?
        if [[ -d "$incoming_skills_dir/$skill_name" ]]; then
            continue  # Will be updated by swap_skills
        fi

        # Is this skill from ANY team pack?
        if is_team_skill "$skill_name"; then
            local origin_team
            origin_team=$(get_skill_team "$skill_name")
            ORPHAN_SKILLS+=("$skill_name:$origin_team")
            log_debug "Orphan skill detected: $skill_name (from $origin_team)"
        fi
    done

    log_debug "Total orphan skills: ${#ORPHAN_SKILLS[@]}"
}

# Remove orphan skills based on ORPHAN_MODE
remove_orphan_skills() {
    if [[ ${#ORPHAN_SKILLS[@]} -eq 0 ]]; then
        return 0
    fi

    local backup_dir=".claude/skills.orphan-backup"

    for orphan in "${ORPHAN_SKILLS[@]}"; do
        local skill_name origin_team
        skill_name=$(echo "$orphan" | cut -d: -f1)
        origin_team=$(echo "$orphan" | cut -d: -f2)

        case "$ORPHAN_MODE" in
            "remove")
                # Backup before removing
                mkdir -p "$backup_dir"
                if [[ -d ".claude/skills/$skill_name" ]]; then
                    cp -rp ".claude/skills/$skill_name" "$backup_dir/$skill_name"
                    rm -rf ".claude/skills/$skill_name"
                    log "Removed orphan skill: $skill_name (was from $origin_team)"
                fi
                ;;
            "keep")
                log "Keeping orphan skill: $skill_name (from $origin_team)"
                ;;
            *)
                # Default: keep (safe)
                log_debug "Keeping orphan skill: $skill_name (no disposition)"
                ;;
        esac
    done

    if [[ "$ORPHAN_MODE" == "remove" ]] && [[ -d "$backup_dir" ]]; then
        log "Orphan skill backups saved to: $backup_dir"
    fi
}

# Remove team skills from previous team
remove_team_skills() {
    log_debug "Removing team skills from previous team"

    if [[ ! -f ".claude/skills/.team-skills" ]]; then
        log_debug "No team skills marker found"
        return 0
    fi

    # Read list and remove each skill directory
    while IFS= read -r skill_dir; do
        [[ -z "$skill_dir" ]] && continue
        if [[ -d ".claude/skills/$skill_dir" ]]; then
            rm -rf ".claude/skills/$skill_dir"
            log_debug "Removed team skill: $skill_dir"
        fi
    done < ".claude/skills/.team-skills"

    # Remove the marker file
    rm -f ".claude/skills/.team-skills"

    log_debug "Team skills removed"
}

# Sync team-specific skills to project
# Team skills are copied to .claude/skills/ with a marker file
# Skills from team layer overlay skeleton skills (team wins on collision)
swap_skills() {
    local team_name="$1"
    local source_dir="$ROSTER_HOME/teams/$team_name/skills"

    log_debug "Checking for team skills in $source_dir"

    # Ensure skills directory exists
    mkdir -p ".claude/skills"

    # Backup and remove previous team skills
    backup_team_skills
    remove_team_skills

    # Check if team has skills
    if [[ ! -d "$source_dir" ]]; then
        log_debug "Team $team_name has no skills/ directory"
        return 0
    fi

    # Count skill directories (each skill is a directory)
    local skill_count
    skill_count=$(find "$source_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$skill_count" -eq 0 ]]; then
        log_debug "Team $team_name has no skill directories"
        return 0
    fi

    log_debug "Syncing $skill_count skill(s) from $team_name"

    # Create marker file to track which skills belong to this team
    local marker_file=".claude/skills/.team-skills"
    : > "$marker_file"

    # Copy each skill directory and record in marker
    for skill_path in "$source_dir"/*/; do
        [[ -d "$skill_path" ]] || continue

        local skill_name
        skill_name=$(basename "$skill_path")

        # Check for collision with existing skeleton skill
        # Team wins: overwrite with warning
        if [[ -d ".claude/skills/$skill_name" ]] && ! grep -q "^$skill_name$" "$marker_file" 2>/dev/null; then
            # Exists but not from team - this is a skeleton skill
            log_warning "Team skill $skill_name overrides skeleton skill"
            rm -rf ".claude/skills/$skill_name"
        fi

        # Copy skill directory
        cp -rp "$skill_path" ".claude/skills/$skill_name"
        echo "$skill_name" >> "$marker_file"
        log_debug "Synced skill: $skill_name"
    done

    # Count successfully synced skills
    local synced_count
    synced_count=$(wc -l < "$marker_file" | tr -d ' ')

    if [[ "$synced_count" -gt 0 ]]; then
        log "Synced: $synced_count team skill(s)"
    fi
}

# ============================================================================
# Shared Skills Functions
# ============================================================================

# Remove shared skills from previous sync
remove_shared_skills() {
    log_debug "Removing shared skills from previous sync"

    if [[ ! -f ".claude/.shared-skills" ]]; then
        log_debug "No shared skills marker found"
        return 0
    fi

    # Read list and remove each skill directory
    while IFS= read -r skill_dir; do
        [[ -z "$skill_dir" ]] && continue
        if [[ -d ".claude/skills/$skill_dir" ]]; then
            rm -rf ".claude/skills/$skill_dir"
            log_debug "Removed shared skill: $skill_dir"
        fi
    done < ".claude/.shared-skills"

    # Remove the marker file
    rm -f ".claude/.shared-skills"

    log_debug "Shared skills removed"
}

# Sync shared skills to project
# Shared skills are copied to .claude/skills/ with a marker file
# Team skills win over shared skills (team-privileged override)
sync_shared_skills() {
    local source_dir="$ROSTER_HOME/teams/shared/skills"

    log_debug "Checking for shared skills in $source_dir"

    # Ensure skills directory exists
    mkdir -p ".claude/skills"

    # Remove previous shared skills
    remove_shared_skills

    # Check if shared skills exist
    if [[ ! -d "$source_dir" ]]; then
        log_debug "No shared skills directory found"
        return 0
    fi

    # Count skill directories (each skill is a directory)
    local skill_count
    skill_count=$(find "$source_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$skill_count" -eq 0 ]]; then
        log_debug "No shared skill directories found"
        return 0
    fi

    log_debug "Syncing $skill_count shared skill(s)"

    # Create marker file to track which skills are from shared
    local marker_file=".claude/.shared-skills"
    : > "$marker_file"

    # Copy each skill directory and record in marker
    for skill_path in "$source_dir"/*/; do
        [[ -d "$skill_path" ]] || continue

        local skill_name
        skill_name=$(basename "$skill_path")

        # Check if team has same skill (team-privileged override)
        if [[ -f ".claude/skills/.team-skills" ]] && grep -q "^$skill_name$" ".claude/skills/.team-skills" 2>/dev/null; then
            log_debug "Team skill $skill_name overrides shared skill"
            continue  # Skip this shared skill
        fi

        # Check for collision with existing skeleton skill
        # Shared wins over skeleton (but loses to team, checked above)
        if [[ -d ".claude/skills/$skill_name" ]]; then
            log_debug "Shared skill $skill_name overrides skeleton skill"
            rm -rf ".claude/skills/$skill_name"
        fi

        # Copy skill directory (flattened - no subdirectory)
        cp -rp "$skill_path" ".claude/skills/$skill_name"
        echo "$skill_name" >> "$marker_file"
        log_debug "Synced shared skill: $skill_name"
    done

    # Count successfully synced skills
    local synced_count
    synced_count=$(wc -l < "$marker_file" | tr -d ' ')

    if [[ "$synced_count" -gt 0 ]]; then
        log "Synced: $synced_count shared skill(s)"
    fi
}

# ============================================================================
# Team Hooks Functions (Phase 2: Unified Sync)
# ============================================================================

# Backup current team hooks (if any exist)
backup_team_hooks() {
    log_debug "Checking for team hooks to backup"

    local backup_dir=".claude/hooks.backup"

    # Check if any team hooks exist (marked by .team-hooks file)
    if [[ ! -d ".claude/hooks" ]] || [[ ! -f ".claude/hooks/.team-hooks" ]]; then
        log_debug "No team hooks to backup"
        return 0
    fi

    # Remove old backup if exists
    if [[ -d "$backup_dir" ]]; then
        log_debug "Removing old hooks backup"
        rm -rf "$backup_dir" || {
            log_warning "Failed to remove old hooks backup"
        }
    fi

    # Read list of team hooks and backup
    mkdir -p "$backup_dir"
    while IFS= read -r hook_file; do
        [[ -z "$hook_file" ]] && continue
        if [[ -f ".claude/hooks/$hook_file" ]]; then
            cp ".claude/hooks/$hook_file" "$backup_dir/$hook_file"
            log_debug "Backed up hook: $hook_file"
        fi
    done < ".claude/hooks/.team-hooks"

    log_debug "Team hooks backed up"
}

# Remove team hooks from previous team
remove_team_hooks() {
    log_debug "Removing team hooks from previous team"

    if [[ ! -f ".claude/hooks/.team-hooks" ]]; then
        log_debug "No team hooks marker found"
        return 0
    fi

    # Read list and remove each hook file
    while IFS= read -r hook_file; do
        [[ -z "$hook_file" ]] && continue
        if [[ -f ".claude/hooks/$hook_file" ]]; then
            rm -f ".claude/hooks/$hook_file"
            log_debug "Removed team hook: $hook_file"
        fi
    done < ".claude/hooks/.team-hooks"

    # Remove the marker file
    rm -f ".claude/hooks/.team-hooks"

    log_debug "Team hooks removed"
}

# Check if a hook belongs to ANY team pack
is_team_hook() {
    local hook_name="$1"
    find "$ROSTER_HOME/teams" -path "*/hooks/$hook_name" -type f 2>/dev/null | grep -q .
}

# Get which team a hook belongs to
get_hook_team() {
    local hook_name="$1"
    local match
    match=$(find "$ROSTER_HOME/teams" -path "*/hooks/$hook_name" -type f 2>/dev/null | head -1)
    if [[ -n "$match" ]]; then
        echo "$match" | sed 's|.*/teams/\([^/]*\)/hooks/.*|\1|'
    fi
}

# Detect orphan hooks - hooks from OTHER teams
detect_hook_orphans() {
    local incoming_team="$1"
    local incoming_hooks_dir="$ROSTER_HOME/teams/$incoming_team/hooks"

    ORPHAN_HOOKS=()

    [[ -d ".claude/hooks" ]] || return 0

    for hook_path in .claude/hooks/*; do
        [[ -f "$hook_path" ]] || continue
        local hook_name
        hook_name=$(basename "$hook_path")

        # Skip marker files and lib directory
        [[ "$hook_name" == .* ]] && continue
        [[ "$hook_name" == "lib" ]] && continue

        # Is this hook in the incoming team?
        if [[ -f "$incoming_hooks_dir/$hook_name" ]]; then
            continue
        fi

        # Is this hook from ANY team pack?
        if is_team_hook "$hook_name"; then
            local origin_team
            origin_team=$(get_hook_team "$hook_name")
            ORPHAN_HOOKS+=("$hook_name:$origin_team")
            log_debug "Orphan hook detected: $hook_name (from $origin_team)"
        fi
    done

    log_debug "Total orphan hooks: ${#ORPHAN_HOOKS[@]}"
}

# Remove orphan hooks based on ORPHAN_MODE
remove_orphan_hooks() {
    if [[ ${#ORPHAN_HOOKS[@]} -eq 0 ]]; then
        return 0
    fi

    local backup_dir=".claude/hooks.orphan-backup"

    for orphan in "${ORPHAN_HOOKS[@]}"; do
        local hook_name origin_team
        hook_name=$(echo "$orphan" | cut -d: -f1)
        origin_team=$(echo "$orphan" | cut -d: -f2)

        case "$ORPHAN_MODE" in
            "remove")
                mkdir -p "$backup_dir"
                if [[ -f ".claude/hooks/$hook_name" ]]; then
                    cp ".claude/hooks/$hook_name" "$backup_dir/$hook_name"
                    rm ".claude/hooks/$hook_name"
                    log "Removed orphan hook: $hook_name (was from $origin_team)"
                fi
                ;;
            "keep")
                log "Keeping orphan hook: $hook_name (from $origin_team)"
                ;;
            *)
                log_debug "Keeping orphan hook: $hook_name (no disposition)"
                ;;
        esac
    done

    if [[ "$ORPHAN_MODE" == "remove" ]] && [[ -d "$backup_dir" ]]; then
        log "Orphan hook backups saved to: $backup_dir"
    fi
}

# Sync base hooks AND team-specific hooks to project
# Base hooks provide foundation, team hooks can override
swap_hooks() {
    local team_name="$1"
    local base_hooks_dir="$ROSTER_HOME/user-hooks"
    local team_hooks_dir="$ROSTER_HOME/teams/$team_name/hooks"

    log_debug "Syncing hooks: base=$base_hooks_dir, team=$team_hooks_dir"

    # Ensure hooks directory exists
    mkdir -p ".claude/hooks"
    mkdir -p ".claude/hooks/lib"

    # Backup and remove previous team hooks
    backup_team_hooks
    remove_team_hooks

    # =========================================================================
    # PHASE 1: Install base hooks from roster/user-hooks/
    # =========================================================================
    if [[ ! -d "$base_hooks_dir" ]]; then
        log_warning "Base hooks directory not found: $base_hooks_dir"
        # Continue anyway - team hooks may still work
    else
        log_debug "Installing base hooks from $base_hooks_dir"

        # Copy root-level hooks (if any exist)
        for hook_file in "$base_hooks_dir"/*.sh; do
            [[ -f "$hook_file" ]] || continue
            local hook_name
            hook_name=$(basename "$hook_file")

            # Skip hidden files
            [[ "$hook_name" == .* ]] && continue

            cp "$hook_file" ".claude/hooks/$hook_name"
            chmod +x ".claude/hooks/$hook_name"
            log_debug "Installed base hook: $hook_name"
        done

        # Copy categorical subdirectories (context-injection, session-guards, tracking, validation)
        for category_dir in "$base_hooks_dir"/*/; do
            [[ -d "$category_dir" ]] || continue
            local category_name
            category_name=$(basename "$category_dir")

            # Skip lib directory (handled separately below)
            [[ "$category_name" == "lib" ]] && continue

            # Create category directory in destination
            mkdir -p ".claude/hooks/$category_name"

            # Copy all .sh files from this category
            for hook_file in "$category_dir"/*.sh; do
                [[ -f "$hook_file" ]] || continue
                local hook_name
                hook_name=$(basename "$hook_file")

                # Skip hidden files
                [[ "$hook_name" == .* ]] && continue

                cp "$hook_file" ".claude/hooks/$category_name/$hook_name"
                chmod +x ".claude/hooks/$category_name/$hook_name"
                log_debug "Installed $category_name hook: $hook_name"
            done
        done

        # Copy lib/ directory contents
        if [[ -d "$base_hooks_dir/lib" ]]; then
            mkdir -p ".claude/hooks/lib"
            for lib_file in "$base_hooks_dir/lib"/*.sh; do
                [[ -f "$lib_file" ]] || continue
                local lib_name
                lib_name=$(basename "$lib_file")

                cp "$lib_file" ".claude/hooks/lib/$lib_name"
                chmod +x ".claude/hooks/lib/$lib_name" 2>/dev/null || true
                log_debug "Installed lib: $lib_name"
            done
        fi

        # Copy base_hooks.yaml if it exists
        if [[ -f "$base_hooks_dir/base_hooks.yaml" ]]; then
            cp "$base_hooks_dir/base_hooks.yaml" ".claude/hooks/base_hooks.yaml"
            log_debug "Installed base_hooks.yaml"
        fi
    fi

    # =========================================================================
    # PHASE 2: Overlay team hooks (if team has hooks directory)
    # =========================================================================
    if [[ ! -d "$team_hooks_dir" ]]; then
        log_debug "Team $team_name has no hooks/ directory"
        return 0
    fi

    local hook_count
    hook_count=$(find "$team_hooks_dir" -maxdepth 1 -type f -name "*.sh" 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$hook_count" -eq 0 ]]; then
        log_debug "Team $team_name has no hook files"
        return 0
    fi

    log_debug "Overlaying $hook_count hook(s) from team $team_name"

    # Create marker file to track team hooks
    local marker_file=".claude/hooks/.team-hooks"
    : > "$marker_file"

    # Copy each team hook (may override base hooks)
    for hook_file in "$team_hooks_dir"/*.sh; do
        [[ -f "$hook_file" ]] || continue

        local hook_name
        hook_name=$(basename "$hook_file")

        # Skip hidden files
        [[ "$hook_name" == .* ]] && continue

        # Check for collision with base hook
        if [[ -f ".claude/hooks/$hook_name" ]]; then
            log_warning "Team hook overrides base: $hook_name"
        fi

        cp "$hook_file" ".claude/hooks/$hook_name"
        chmod +x ".claude/hooks/$hook_name"
        echo "$hook_name" >> "$marker_file"
        log_debug "Installed team hook: $hook_name"
    done

    # Count successfully synced team hooks
    local synced_count
    synced_count=$(wc -l < "$marker_file" | tr -d ' ')

    if [[ "$synced_count" -gt 0 ]]; then
        log "Synced: $synced_count team hook(s)"
    fi
}

# ============================================================================
# Hook Registration Functions (Scope 2: YAML to settings.local.json)
# ============================================================================

# Check if yq v4+ is available
# Usage: require_yq
require_yq() {
    if ! command -v yq &>/dev/null; then
        log_error "yq is required but not installed"
        log_error "Install with: brew install yq (macOS) or pip install yq"
        return 1
    fi

    # Check for yq v4+ (mikefarah/yq)
    local yq_version
    yq_version=$(yq --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
    local major_version
    major_version=$(echo "$yq_version" | cut -d. -f1)

    if [[ -z "$major_version" ]] || [[ "$major_version" -lt 4 ]]; then
        log_error "yq v4+ is required (found: $yq_version)"
        log_error "Install with: brew install yq"
        return 1
    fi

    return 0
}

# Parse hooks.yaml file and emit JSON-lines format
# Usage: parse_hooks_yaml "path/to/hooks.yaml"
# Output: One JSON object per line: {"event":"...","matcher":"...","path":"...","timeout":N}
parse_hooks_yaml() {
    local yaml_file="$1"

    # File doesn't exist - return empty
    if [[ ! -f "$yaml_file" ]]; then
        return 0
    fi

    # Validate schema version
    local schema_version
    schema_version=$(yq -r '.schema_version // ""' "$yaml_file" 2>/dev/null)
    if [[ -n "$schema_version" ]] && [[ "$schema_version" != "1.0" ]]; then
        log_warning "Unknown schema version: $schema_version (expected 1.0)"
    fi

    # Get hook count
    local hook_count
    hook_count=$(yq -r '.hooks | length' "$yaml_file" 2>/dev/null)
    if [[ -z "$hook_count" ]] || [[ "$hook_count" -eq 0 ]]; then
        return 0
    fi

    # Process each hook entry
    local i
    for ((i=0; i<hook_count; i++)); do
        local event matcher path timeout

        event=$(yq -r ".hooks[$i].event // \"\"" "$yaml_file")
        matcher=$(yq -r ".hooks[$i].matcher // \"\"" "$yaml_file")
        path=$(yq -r ".hooks[$i].path // \"\"" "$yaml_file")
        timeout=$(yq -r ".hooks[$i].timeout // 5" "$yaml_file")

        # Validate event type
        case "$event" in
            SessionStart|Stop|PreToolUse|PostToolUse|UserPromptSubmit)
                ;;
            *)
                log_warning "Invalid event type: $event (skipping)"
                continue
                ;;
        esac

        # Validate matcher requirement for PreToolUse and PostToolUse
        if [[ "$event" == "PreToolUse" || "$event" == "PostToolUse" ]]; then
            if [[ -z "$matcher" ]]; then
                log_warning "Event $event requires matcher (skipping: $path)"
                continue
            fi
        fi

        # Validate path is provided
        if [[ -z "$path" ]]; then
            log_warning "Hook entry $i missing path (skipping)"
            continue
        fi

        # Validate matcher syntax (check regex compiles without error)
        if [[ -n "$matcher" ]]; then
            # Use grep -E with a test string to validate regex syntax
            # We check exit code 0 or 1 (valid regex), 2 means syntax error
            echo "test" | grep -E "$matcher" >/dev/null 2>&1
            local grep_exit=$?
            if [[ $grep_exit -eq 2 ]]; then
                log_warning "Invalid matcher regex: $matcher (skipping: $path)"
                continue
            fi
        fi

        # Clamp timeout to valid range
        if [[ "$timeout" -gt 60 ]]; then
            log_warning "Timeout $timeout exceeds 60s limit, clamping to 60 (hook: $path)"
            timeout=60
        fi
        if [[ "$timeout" -lt 1 ]]; then
            timeout=5
        fi

        # Emit registration record (JSON-lines format)
        # Use jq to properly escape strings
        jq -n -c \
            --arg event "$event" \
            --arg matcher "$matcher" \
            --arg path "$path" \
            --argjson timeout "$timeout" \
            '{event: $event, matcher: $matcher, path: $path, timeout: $timeout}'
    done
}

# Extract non-roster hooks from existing settings.local.json
# These are hooks whose command does NOT contain ".claude/hooks/"
# Usage: extract_non_roster_hooks "settings_file"
# Output: JSON object with preserved hooks by event type
extract_non_roster_hooks() {
    local settings_file="$1"

    # File doesn't exist - return empty object
    if [[ ! -f "$settings_file" ]]; then
        echo "{}"
        return 0
    fi

    # Read current hooks section
    local current_hooks
    current_hooks=$(jq '.hooks // {}' "$settings_file" 2>/dev/null)
    if [[ -z "$current_hooks" ]] || [[ "$current_hooks" == "null" ]]; then
        echo "{}"
        return 0
    fi

    # For each event type, filter out roster-managed hooks
    # Roster hooks contain ".claude/hooks/" in the command path
    local preserved="{}"
    local events=("SessionStart" "Stop" "PreToolUse" "PostToolUse" "UserPromptSubmit")

    for event in "${events[@]}"; do
        local event_entries
        event_entries=$(echo "$current_hooks" | jq -c ".\"$event\" // []")

        local entry_count
        entry_count=$(echo "$event_entries" | jq 'length')
        [[ "$entry_count" -eq 0 ]] && continue

        local filtered_entries="[]"
        local i
        for ((i=0; i<entry_count; i++)); do
            local entry
            entry=$(echo "$event_entries" | jq -c ".[$i]")

            # Filter hooks array within entry to exclude roster-managed ones
            local filtered_hooks
            filtered_hooks=$(echo "$entry" | jq -c '[.hooks // [] | .[] | select(.command | contains(".claude/hooks/") | not)]')

            local filtered_count
            filtered_count=$(echo "$filtered_hooks" | jq 'length')

            if [[ "$filtered_count" -gt 0 ]]; then
                # Update entry with filtered hooks
                local new_entry
                new_entry=$(echo "$entry" | jq -c ".hooks = $filtered_hooks")
                filtered_entries=$(echo "$filtered_entries" | jq -c ". + [$new_entry]")
            fi
        done

        local filtered_len
        filtered_len=$(echo "$filtered_entries" | jq 'length')
        if [[ "$filtered_len" -gt 0 ]]; then
            preserved=$(echo "$preserved" | jq -c ".\"$event\" = $filtered_entries")
        fi
    done

    echo "$preserved"
}

# Merge hook registrations (base first, team appended)
# Usage: merge_hook_registrations "base_registrations" "team_registrations"
# Input: JSON-lines format (one JSON object per line)
# Output: Combined JSON-lines (base first, then team)
merge_hook_registrations() {
    local base_registrations="$1"
    local team_registrations="$2"

    # Combine all registrations (base first, team second)
    printf '%s\n%s' "$base_registrations" "$team_registrations" | grep -v '^$' || true
}

# Generate Claude Code hooks JSON format from registrations
# Usage: generate_hooks_json "registrations"
# Input: JSON-lines format
# Output: Claude Code settings.local.json hooks object
generate_hooks_json() {
    local registrations="$1"

    # If no registrations, return empty object
    if [[ -z "$registrations" ]]; then
        echo "{}"
        return 0
    fi

    # Convert JSON-lines to JSON array
    local all_hooks
    all_hooks=$(echo "$registrations" | jq -s '.' 2>/dev/null)
    if [[ -z "$all_hooks" ]] || [[ "$all_hooks" == "null" ]]; then
        echo "{}"
        return 0
    fi

    # Group by event type and build Claude Code format
    local events=("SessionStart" "Stop" "PreToolUse" "PostToolUse" "UserPromptSubmit")
    local result="{}"

    for event in "${events[@]}"; do
        # Filter hooks for this event
        local event_hooks
        event_hooks=$(echo "$all_hooks" | jq -c "[.[] | select(.event == \"$event\")]")

        local count
        count=$(echo "$event_hooks" | jq 'length')
        [[ "$count" -eq 0 ]] && continue

        # Get unique matchers for this event (preserve order)
        local matchers
        matchers=$(echo "$event_hooks" | jq -r '.[].matcher' | awk '!seen[$0]++')

        # Build entries for this event
        local event_entries="[]"

        while IFS= read -r matcher; do
            # Get all hooks for this matcher
            local matcher_hooks
            if [[ -z "$matcher" ]]; then
                matcher_hooks=$(echo "$event_hooks" | jq -c "[.[] | select(.matcher == \"\")]")
            else
                matcher_hooks=$(echo "$event_hooks" | jq -c --arg m "$matcher" '[.[] | select(.matcher == $m)]')
            fi

            local hook_count
            hook_count=$(echo "$matcher_hooks" | jq 'length')
            [[ "$hook_count" -eq 0 ]] && continue

            # Build hooks array for this matcher
            local hooks_array="[]"
            local j
            for ((j=0; j<hook_count; j++)); do
                local path timeout
                path=$(echo "$matcher_hooks" | jq -r ".[$j].path")
                timeout=$(echo "$matcher_hooks" | jq -r ".[$j].timeout")

                local hook_obj
                hook_obj=$(jq -n -c \
                    --arg path "\$CLAUDE_PROJECT_DIR/.claude/hooks/$path" \
                    --argjson timeout "$timeout" \
                    '{type: "command", command: $path, timeout: $timeout}')

                hooks_array=$(echo "$hooks_array" | jq -c ". + [$hook_obj]")
            done

            # Build entry object
            local entry
            if [[ -n "$matcher" ]]; then
                entry=$(jq -n -c \
                    --arg matcher "$matcher" \
                    --argjson hooks "$hooks_array" \
                    '{matcher: $matcher, hooks: $hooks}')
            else
                entry=$(jq -n -c \
                    --argjson hooks "$hooks_array" \
                    '{hooks: $hooks}')
            fi

            event_entries=$(echo "$event_entries" | jq -c ". + [$entry]")
        done <<< "$matchers"

        # Add event entries to result
        result=$(echo "$result" | jq -c --argjson entries "$event_entries" ".\"$event\" = \$entries")
    done

    echo "$result"
}

# Merge generated hooks with preserved user hooks
# Usage: merge_with_preserved "generated_json" "preserved_json"
# Output: Combined hooks JSON object
merge_with_preserved() {
    local generated="$1"
    local preserved="$2"

    # If no preserved hooks, return generated
    if [[ -z "$preserved" ]] || [[ "$preserved" == "{}" ]]; then
        echo "$generated"
        return 0
    fi

    # For each event type, append preserved entries to generated
    local merged="$generated"
    local events=("SessionStart" "Stop" "PreToolUse" "PostToolUse" "UserPromptSubmit")

    for event in "${events[@]}"; do
        local preserved_entries
        preserved_entries=$(echo "$preserved" | jq -c ".\"$event\" // []")

        local preserved_count
        preserved_count=$(echo "$preserved_entries" | jq 'length')
        [[ "$preserved_count" -eq 0 ]] && continue

        # Append preserved entries to generated event
        local generated_entries
        generated_entries=$(echo "$merged" | jq -c ".\"$event\" // []")

        local combined
        combined=$(jq -n -c --argjson gen "$generated_entries" --argjson pres "$preserved_entries" '$gen + $pres')

        merged=$(echo "$merged" | jq -c --argjson entries "$combined" ".\"$event\" = \$entries")
    done

    echo "$merged"
}

# Sync hook registrations to settings.local.json
# Called after swap_hooks() syncs the actual hook files
# Usage: swap_hook_registrations "team_name"
swap_hook_registrations() {
    local team_name="$1"
    local settings_file=".claude/settings.local.json"
    local base_hooks_yaml="$ROSTER_HOME/user-hooks/base_hooks.yaml"
    local team_hooks_yaml="$ROSTER_HOME/teams/$team_name/hooks.yaml"

    log_debug "Updating hook registrations for team: $team_name"

    # Require yq for YAML parsing
    if ! require_yq; then
        log_error "Cannot update hook registrations without yq"
        return 1
    fi

    # Ensure settings file exists with valid JSON
    if [[ ! -f "$settings_file" ]]; then
        echo '{}' > "$settings_file"
    fi

    # Validate JSON before proceeding
    if ! jq empty "$settings_file" 2>/dev/null; then
        log_error "Invalid JSON in $settings_file, backing up and creating fresh"
        mv "$settings_file" "${settings_file}.corrupt.$(date +%s)"
        echo '{}' > "$settings_file"
    fi

    # Dry-run mode: preview changes
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        log "Hook registrations preview (dry-run):"
    fi

    # Step 1: Extract non-roster hooks for preservation
    local preserved_hooks
    preserved_hooks=$(extract_non_roster_hooks "$settings_file")
    local preserved_count
    preserved_count=$(echo "$preserved_hooks" | jq '[.[] | length] | add // 0')
    if [[ "$preserved_count" -gt 0 ]]; then
        log_debug "Preserved $preserved_count non-roster hook entries"
    fi

    # Step 2: Parse base hooks
    local base_registrations=""
    if [[ -f "$base_hooks_yaml" ]]; then
        base_registrations=$(parse_hooks_yaml "$base_hooks_yaml")
        local base_count
        base_count=$(echo "$base_registrations" | grep -c '^{' 2>/dev/null || echo 0)
        log_debug "Parsed $base_count base hook registrations"
    else
        log_warning "Base hooks file not found: $base_hooks_yaml"
    fi

    # Step 3: Parse team hooks (optional)
    local team_registrations=""
    if [[ -f "$team_hooks_yaml" ]]; then
        team_registrations=$(parse_hooks_yaml "$team_hooks_yaml")
        local team_count
        team_count=$(echo "$team_registrations" | grep -c '^{' 2>/dev/null || echo 0)
        log_debug "Parsed $team_count team hook registrations"
    else
        log_debug "No team hooks.yaml for $team_name"
    fi

    # Step 4: Merge registrations (base first, team second)
    local merged_registrations
    merged_registrations=$(merge_hook_registrations "$base_registrations" "$team_registrations")

    # Step 5: Generate hooks JSON
    local generated_hooks
    generated_hooks=$(generate_hooks_json "$merged_registrations")

    # Step 6: Merge with preserved hooks
    local final_hooks
    final_hooks=$(merge_with_preserved "$generated_hooks" "$preserved_hooks")

    # Dry-run mode: show what would be written
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        echo "$final_hooks" | jq '.'
        return 0
    fi

    # Step 7: Update settings.local.json
    local temp_file="${settings_file}.tmp.$$"
    if ! jq --argjson hooks "$final_hooks" '.hooks = $hooks' "$settings_file" > "$temp_file" 2>/dev/null; then
        rm -f "$temp_file"
        log_error "Failed to generate updated settings.local.json"
        return 1
    fi

    # Validate generated JSON
    if ! jq empty "$temp_file" 2>/dev/null; then
        rm -f "$temp_file"
        log_error "Generated invalid JSON, hook registrations not updated"
        return 1
    fi

    # Atomic rename
    mv "$temp_file" "$settings_file" || {
        rm -f "$temp_file"
        log_error "Failed to update settings.local.json"
        return 1
    }

    log "Updated hook registrations in settings.local.json"
    return 0
}

# ============================================================================
# CLAUDE.md Update Functions
# ============================================================================

# REQ-3.1: Get produces field from workflow.yaml for an agent
# Reads directly from roster source instead of hardcoded mapping
get_produces_from_workflow() {
    local team_name="$1"
    local agent_name="$2"
    local workflow_file="$ROSTER_HOME/teams/$team_name/workflow.yaml"

    if [[ ! -f "$workflow_file" ]]; then
        echo "Artifacts"  # Fallback
        return
    fi

    # Extract produces for agent from workflow.yaml phases section
    # Uses awk for reliable parsing of YAML structure
    local produces
    produces=$(awk -v agent="$agent_name" '
        /^phases:/ { in_phases = 1; next }
        /^[a-z_]+:/ && !/^[[:space:]]/ { in_phases = 0 }
        in_phases && /agent:/ && $0 ~ agent {
            found_agent = 1
            next
        }
        in_phases && found_agent && /produces:/ {
            gsub(/.*produces:[[:space:]]*/, "")
            gsub(/[[:space:]]*$/, "")
            print
            exit
        }
        in_phases && found_agent && /^[[:space:]]*-[[:space:]]/ {
            # New phase started without finding produces
            found_agent = 0
        }
    ' "$workflow_file")

    if [[ -n "$produces" ]]; then
        # Capitalize first letter and format (e.g., "prd" -> "PRD", "tdd" -> "TDD", "code" -> "Code")
        case "$produces" in
            prd|PRD) echo "PRD" ;;
            tdd|TDD) echo "TDD" ;;
            adr|ADR) echo "ADR" ;;
            code) echo "Code" ;;
            test-plan|test_plan) echo "Test reports" ;;
            *)
                # Capitalize first letter
                echo "$(echo "${produces:0:1}" | tr '[:lower:]' '[:upper:]')${produces:1}"
                ;;
        esac
    else
        echo "Artifacts"  # Fallback
    fi
}

# REQ-3.1: Get all phases from workflow.yaml in order
# Returns list of "agent:produces" pairs in workflow order
get_workflow_phases() {
    local team_name="$1"
    local workflow_file="$ROSTER_HOME/teams/$team_name/workflow.yaml"

    if [[ ! -f "$workflow_file" ]]; then
        return
    fi

    # Parse phases section - extract agent and produces for each phase
    # This is a simple parser that works with the standard workflow.yaml format
    awk '
        /^phases:/ { in_phases = 1; next }
        /^[a-z_]+:/ && !/^[[:space:]]/ { in_phases = 0 }
        in_phases && /agent:/ {
            gsub(/.*agent:[[:space:]]*/, "")
            gsub(/[[:space:]]*$/, "")
            agent = $0
        }
        in_phases && /produces:/ {
            gsub(/.*produces:[[:space:]]*/, "")
            gsub(/[[:space:]]*$/, "")
            produces = $0
            if (agent != "") {
                print agent ":" produces
                agent = ""
            }
        }
    ' "$workflow_file"
}

# Update CLAUDE.md to reflect current team's agents
# REQ-3.1: Reads from roster source instead of disk state after copy
# This ensures Claude Code's context matches the swapped agents
update_claude_md() {
    local team_name="$1"
    local claude_md=".claude/CLAUDE.md"

    if [[ ! -f "$claude_md" ]]; then
        log_debug "No CLAUDE.md found, skipping update"
        return 0
    fi

    log_debug "Updating CLAUDE.md for team $team_name"

    # REQ-3.1: Read from roster source directly, not disk state after copy
    local source_agents="$ROSTER_HOME/teams/$team_name/agents"

    # Create temp files for agent data
    local agent_list_file agent_table_file temp_file
    agent_list_file=$(mktemp)
    agent_table_file=$(mktemp)
    temp_file=$(mktemp)

    # REQ-3.1: Build agent list from roster source, not .claude/agents/
    for agent_file in "$source_agents"/*.md; do
        [[ -f "$agent_file" ]] || continue

        local basename name desc role produces
        basename=$(basename "$agent_file" .md)

        # Extract name from YAML frontmatter
        name=$(sed -n '/^---$/,/^---$/p' "$agent_file" | grep "^name:" | head -1 | sed 's/^name:[[:space:]]*//')

        # Extract description - handle both single-line and multiline YAML (for Agent Configurations list)
        local raw_desc
        local desc_line
        desc_line=$(sed -n '/^---$/,/^---$/p' "$agent_file" | grep "^description:")
        # Check if value is on same line (single-line) or next line (multiline with |)
        if echo "$desc_line" | grep -q 'description:[[:space:]]*["|'"'"']'; then
            # Single-line: description: "value" or description: 'value'
            raw_desc=$(echo "$desc_line" | sed 's/^description:[[:space:]]*//' | sed 's/^["'"'"']//' | sed 's/["'"'"']$//')
        elif echo "$desc_line" | grep -q 'description:[[:space:]]*|'; then
            # Multiline: description: | followed by indented text
            raw_desc=$(sed -n '/^---$/,/^---$/p' "$agent_file" | grep -A1 "^description:" | tail -1 | sed 's/^[[:space:]]*//')
        else
            # Fallback: try same line without quotes
            raw_desc=$(echo "$desc_line" | sed 's/^description:[[:space:]]*//')
        fi

        # Find first sentence (up to first period) or take full line
        if [[ "$raw_desc" == *"."* ]]; then
            desc=$(echo "$raw_desc" | sed 's/\([^.]*\.\).*/\1/')
        else
            desc="$raw_desc"
        fi

        # Truncate to 80 chars at word boundary for Agent Configurations
        if [[ ${#desc} -gt 80 ]]; then
            desc=$(echo "$desc" | cut -c1-80 | sed 's/[[:space:]][^[:space:]]*$//')
        fi

        # Extract role field for Quick Start table
        local role_field
        role_field=$(sed -n '/^---$/,/^---$/p' "$agent_file" | grep "^role:" | head -1 | sed 's/^role:[[:space:]]*//' | sed 's/^["'"'"']//' | sed 's/["'"'"']$//')

        # Build agent list for Agent Configurations section
        echo "- \`${basename}.md\` - ${desc}" >> "$agent_list_file"

        # Use role field if available, otherwise fallback to first 50 chars of desc
        if [[ -n "$role_field" ]]; then
            role="$role_field"
        else
            if [[ ${#desc} -gt 50 ]]; then
                role=$(echo "$desc" | cut -c1-50 | sed 's/[[:space:]][^[:space:]]*$//')
            else
                role="$desc"
            fi
        fi

        # REQ-3.1: Get produces from workflow.yaml instead of hardcoded case statement
        produces=$(get_produces_from_workflow "$team_name" "$basename")

        # Special case for orchestrator (not in phases but always present)
        if [[ "$basename" == "orchestrator" && "$produces" == "Artifacts" ]]; then
            produces="Work breakdown"
        fi

        echo "| **${name:-$basename}** | ${role} | ${produces} |" >> "$agent_table_file"
    done

    # REQ-3.1: Count agents from roster source, not disk
    local agent_count
    agent_count=$(find "$source_agents" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    # Update Quick Start section using sed
    # First, copy everything before the table (BSD-compatible: use awk instead of head -n -1)
    # Handle both team mode ("This project uses...") and skeleton mode ("No team currently active")
    local before_table_line
    before_table_line=$(grep -n "^This project uses a [0-9]*-agent" "$claude_md" 2>/dev/null | head -1 | cut -d: -f1 || true)
    if [[ -z "$before_table_line" ]]; then
        # Check for skeleton mode pattern
        before_table_line=$(grep -n "^No team currently active" "$claude_md" 2>/dev/null | head -1 | cut -d: -f1 || true)
    fi
    if [[ -n "$before_table_line" ]] && [[ "$before_table_line" -gt 1 ]]; then
        head -n $((before_table_line - 1)) "$claude_md" > "$temp_file"
    else
        : > "$temp_file"  # Empty file
    fi

    # Add new header and table
    echo "This project uses a ${agent_count}-agent workflow (${team_name}):" >> "$temp_file"
    echo "" >> "$temp_file"
    echo "| Agent | Role | Produces |" >> "$temp_file"
    echo "| ----- | ---- | -------- |" >> "$temp_file"
    cat "$agent_table_file" >> "$temp_file"
    echo "" >> "$temp_file"

    # Find where the old table ends and copy from there
    # Handle both team mode ("**New here") and skeleton mode ("**Get started")
    local table_end_line
    table_end_line=$(grep -n "^\*\*New here" "$claude_md" 2>/dev/null | head -1 | cut -d: -f1 || true)
    if [[ -z "$table_end_line" ]]; then
        # Check for skeleton mode pattern
        table_end_line=$(grep -n "^\*\*Get started" "$claude_md" 2>/dev/null | head -1 | cut -d: -f1 || true)
    fi
    if [[ -n "$table_end_line" ]]; then
        sed -n "${table_end_line},\$p" "$claude_md" >> "$temp_file"
    fi

    # Now update the Agent Configurations section
    local config_start config_end
    config_start=$(grep -n "^## Agent Configurations" "$temp_file" | head -1 | cut -d: -f1)
    config_end=$(sed -n "${config_start},\$p" "$temp_file" | grep -n "^## " | sed -n '2p' | cut -d: -f1)

    if [[ -n "$config_start" ]]; then
        # Copy everything before Agent Configurations
        head -n "$config_start" "$temp_file" > "$claude_md"
        echo "" >> "$claude_md"
        echo "Full agent prompts live in \`.claude/agents/\`:" >> "$claude_md"
        echo "" >> "$claude_md"
        cat "$agent_list_file" >> "$claude_md"
        echo "" >> "$claude_md"

        # Copy everything after Agent Configurations section
        if [[ -n "$config_end" ]]; then
            local skip_lines=$((config_start + config_end - 1))
            tail -n +"$skip_lines" "$temp_file" >> "$claude_md"
        fi
    else
        # No Agent Configurations section, just use temp_file as-is
        cp "$temp_file" "$claude_md"
    fi

    # Cleanup
    rm -f "$agent_list_file" "$agent_table_file" "$temp_file"

    log_debug "CLAUDE.md updated with $agent_count agents"
}

# Update active team state
update_active_team() {
    local team_name="$1"

    log_debug "Updating ACTIVE_TEAM state"

    echo -n "$team_name" > .claude/ACTIVE_TEAM || {
        log_warning "Failed to update ACTIVE_TEAM (agents swapped successfully)"
        log "Manually fix: echo '$team_name' > .claude/ACTIVE_TEAM"
        exit "$EXIT_SWAP_FAILURE"
    }

    log_debug "State updated to $team_name"
}

# Preview what refresh would change (for --dry-run)
preview_refresh() {
    local team_name="$1"
    local source_dir="$ROSTER_HOME/teams/$team_name/agents"

    log "Dry-run: Would refresh $team_name"
    echo ""
    echo "Agent changes:"

    for agent_file in "$source_dir"/*.md; do
        [[ -f "$agent_file" ]] || continue
        local agent_name
        agent_name=$(basename "$agent_file")

        if [[ -f ".claude/agents/$agent_name" ]]; then
            if diff -q ".claude/agents/$agent_name" "$agent_file" >/dev/null 2>&1; then
                echo "  = $agent_name (unchanged)"
            else
                echo "  ~ $agent_name (modified in roster)"
            fi
        else
            echo "  + $agent_name (new)"
        fi
    done

    # Check for agents that would become orphans
    if [[ -d ".claude/agents" ]]; then
        for local_agent in .claude/agents/*.md; do
            [[ -f "$local_agent" ]] || continue
            local agent_name
            agent_name=$(basename "$local_agent")
            if [[ ! -f "$source_dir/$agent_name" ]]; then
                echo "  ? $agent_name (orphan - not in roster)"
            fi
        done
    fi

    # Check for orphan commands (commands from other teams)
    echo ""
    echo "Command orphans (from other teams):"
    detect_command_orphans "$team_name"
    if [[ ${#ORPHAN_COMMANDS[@]} -gt 0 ]]; then
        for orphan in "${ORPHAN_COMMANDS[@]}"; do
            local cmd_name origin_team
            cmd_name=$(echo "$orphan" | cut -d: -f1)
            origin_team=$(echo "$orphan" | cut -d: -f2)
            echo "  ? $cmd_name (from $origin_team)"
        done
    else
        echo "  (none)"
    fi

    # Check for orphan skills (skills from other teams)
    echo ""
    echo "Skill orphans (from other teams):"
    detect_skill_orphans "$team_name"
    if [[ ${#ORPHAN_SKILLS[@]} -gt 0 ]]; then
        for orphan in "${ORPHAN_SKILLS[@]}"; do
            local skill_name origin_team
            skill_name=$(echo "$orphan" | cut -d: -f1)
            origin_team=$(echo "$orphan" | cut -d: -f2)
            echo "  ? $skill_name (from $origin_team)"
        done
    else
        echo "  (none)"
    fi

    # Check for orphan hooks (hooks from other teams)
    echo ""
    echo "Hook orphans (from other teams):"
    detect_hook_orphans "$team_name"
    if [[ ${#ORPHAN_HOOKS[@]} -gt 0 ]]; then
        for orphan in "${ORPHAN_HOOKS[@]}"; do
            local hook_name origin_team
            hook_name=$(echo "$orphan" | cut -d: -f1)
            origin_team=$(echo "$orphan" | cut -d: -f2)
            echo "  ? $hook_name (from $origin_team)"
        done
    else
        echo "  (none)"
    fi

    # Summary of orphans
    local total_orphans=$((${#ORPHAN_COMMANDS[@]} + ${#ORPHAN_SKILLS[@]} + ${#ORPHAN_HOOKS[@]}))
    if [[ $total_orphans -gt 0 ]]; then
        echo ""
        echo "Use --remove-all to clean up $total_orphans orphan(s)"
    fi

    # Preview hook registrations (Scope 2)
    echo ""
    echo "Hook registrations (settings.local.json):"
    if require_yq 2>/dev/null; then
        swap_hook_registrations "$team_name"
    else
        echo "  (skipped - yq not available)"
    fi

    echo ""
    echo "No changes made (--dry-run mode)"
}

# Main swap orchestration
perform_swap() {
    local team_name="$1"

    log_debug "Starting swap to $team_name"

    # Set up signal handlers for graceful interruption
    setup_signal_handlers

    # Check for recovery from interrupted swap
    check_journal_recovery

    # Get current team (before swap) for journal
    local source_team=""
    if [[ -f ".claude/ACTIVE_TEAM" ]]; then
        source_team=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')
    fi

    # Check if already active (idempotency, unless --update)
    if [[ -n "$source_team" ]] && [[ "$UPDATE_MODE" -eq 0 ]]; then
        if [[ "$source_team" == "$team_name" ]]; then
            log "Already using $team_name (no changes needed)"
            log "Use --update to pull latest from roster"
            exit "$EXIT_SUCCESS"
        fi
    fi

    # Validate pack and project
    local agent_count
    agent_count=$(validate_pack "$team_name")
    validate_project

    # Dry-run mode: preview changes and exit
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        preview_refresh "$team_name"
        exit "$EXIT_SUCCESS"
    fi

    # =========================================================================
    # PHASE: PREPARING - Create journal and validate
    # =========================================================================
    create_journal "$source_team" "$team_name" || {
        exit "$EXIT_SWAP_FAILURE"
    }

    # Detect orphan agents (current agents not in target team)
    detect_orphans "$team_name"

    # Handle orphans if any exist
    if [[ ${#ORPHAN_AGENTS[@]} -gt 0 ]]; then
        # Get user disposition for orphans (interactive or via flags)
        prompt_disposition "$team_name"

        # Promote agents before swap (while they still exist)
        promote_agents

        # Stash agents to keep (will be restored after swap)
        stash_kept_agents

        # Log removed agents
        if [[ ${#AGENTS_TO_REMOVE[@]} -gt 0 ]]; then
            log_debug "Will remove ${#AGENTS_TO_REMOVE[@]} agent(s) (available in backup)"
        fi
    fi

    # =========================================================================
    # PHASE: BACKING - Create comprehensive backup
    # =========================================================================
    update_journal_phase "$PHASE_BACKING"

    # Create comprehensive backup for rollback (new transaction-safe backup)
    create_swap_backup || {
        log_error "Failed to create swap backup"
        delete_journal
        exit "$EXIT_BACKUP_FAILURE"
    }

    # Also create legacy backup for backward compatibility
    backup_current_agents

    # =========================================================================
    # PHASE: STAGING - Prepare all resources in staging directory
    # =========================================================================
    update_journal_phase "$PHASE_STAGING"

    create_staging || {
        log_error "Failed to create staging directory"
        rollback_swap
        exit "$EXIT_SWAP_FAILURE"
    }

    # Stage agents
    stage_agents "$team_name" || {
        log_error "Failed to stage agents"
        rollback_swap
        exit "$EXIT_SWAP_FAILURE"
    }

    # Stage workflow
    stage_workflow "$team_name"

    # Stage ACTIVE_TEAM (prepared but committed last)
    stage_active_team "$team_name" || {
        log_error "Failed to stage ACTIVE_TEAM"
        rollback_swap
        exit "$EXIT_SWAP_FAILURE"
    }

    # =========================================================================
    # PHASE: VERIFYING - Verify staging integrity
    # =========================================================================
    update_journal_phase "$PHASE_VERIFYING"

    verify_staging "$agent_count" || {
        log_error "Staging verification failed"
        rollback_swap
        exit "$EXIT_SWAP_FAILURE"
    }

    # =========================================================================
    # PHASE: COMMITTING - Atomic commit of staged resources
    # =========================================================================
    commit_staged_resources "$team_name" || {
        log_error "Commit failed - attempting rollback"
        rollback_swap
        exit "$EXIT_SWAP_FAILURE"
    }

    # =========================================================================
    # PART 1 OF COMMIT: Commands, Skills, Hooks (still rollback-able)
    # =========================================================================

    # Detect and handle orphan commands (commands from other teams)
    detect_command_orphans "$team_name"
    if [[ ${#ORPHAN_COMMANDS[@]} -gt 0 ]]; then
        if [[ -z "$ORPHAN_MODE" ]]; then
            log_warning "Found ${#ORPHAN_COMMANDS[@]} orphan command(s) from other teams:"
            for orphan in "${ORPHAN_COMMANDS[@]}"; do
                local cmd_name origin_team
                cmd_name=$(echo "$orphan" | cut -d: -f1)
                origin_team=$(echo "$orphan" | cut -d: -f2)
                echo "  - $cmd_name (from $origin_team)"
            done
            log "Use --remove-all to clean up orphan commands"
        else
            remove_orphan_commands
        fi
    fi

    # Sync team-specific commands
    swap_commands "$team_name"

    # Detect and handle orphan skills (skills from other teams)
    detect_skill_orphans "$team_name"
    if [[ ${#ORPHAN_SKILLS[@]} -gt 0 ]]; then
        if [[ -z "$ORPHAN_MODE" ]]; then
            # Non-interactive mode without flags - warn but don't block
            log_warning "Found ${#ORPHAN_SKILLS[@]} orphan skill(s) from other teams:"
            for orphan in "${ORPHAN_SKILLS[@]}"; do
                local skill_name origin_team
                skill_name=$(echo "$orphan" | cut -d: -f1)
                origin_team=$(echo "$orphan" | cut -d: -f2)
                echo "  - $skill_name (from $origin_team)"
            done
            log "Use --remove-all to clean up orphan skills"
        else
            remove_orphan_skills
        fi
    fi

    # Sync team-specific skills (Phase 2: Unified Sync)
    swap_skills "$team_name"

    # Sync shared skills (always active, team-privileged override)
    sync_shared_skills

    # Detect and handle orphan hooks (hooks from other teams)
    detect_hook_orphans "$team_name"
    if [[ ${#ORPHAN_HOOKS[@]} -gt 0 ]]; then
        if [[ -z "$ORPHAN_MODE" ]]; then
            log_warning "Found ${#ORPHAN_HOOKS[@]} orphan hook(s) from other teams:"
            for orphan in "${ORPHAN_HOOKS[@]}"; do
                local hook_name origin_team
                hook_name=$(echo "$orphan" | cut -d: -f1)
                origin_team=$(echo "$orphan" | cut -d: -f2)
                echo "  - $hook_name (from $origin_team)"
            done
            log "Use --remove-all to clean up orphan hooks"
        else
            remove_orphan_hooks
        fi
    fi

    # Sync team hooks
    swap_hooks "$team_name"

    # Update hook registrations in settings.local.json (Scope 2)
    swap_hook_registrations "$team_name"

    # =========================================================================
    # PART 2 OF COMMIT: Manifest and ACTIVE_TEAM (the actual commit)
    # =========================================================================
    # Write manifest with current state (after commands synced so we capture them)
    # IMPORTANT: Manifest must be written BEFORE ACTIVE_TEAM
    write_manifest "$team_name"

    # =========================================================================
    # FINAL COMMIT: Write ACTIVE_TEAM (LAST - this is the commit point)
    # =========================================================================
    # ACTIVE_TEAM is the commit indicator - if it contains the new team name,
    # the swap is considered complete. Writing it LAST ensures all resources
    # are in place first.
    if [[ -f "$STAGING_DIR/ACTIVE_TEAM" ]]; then
        mv "$STAGING_DIR/ACTIVE_TEAM" .claude/ACTIVE_TEAM || {
            log_error "Failed to commit ACTIVE_TEAM"
            update_journal_error "Failed to commit ACTIVE_TEAM"
            rollback_swap
            exit "$EXIT_SWAP_FAILURE"
        }
    else
        # Fallback if staging was already cleaned up
        echo -n "$team_name" > .claude/ACTIVE_TEAM || {
            log_error "Failed to write ACTIVE_TEAM"
            rollback_swap
            exit "$EXIT_SWAP_FAILURE"
        }
    fi

    # =========================================================================
    # PHASE: COMPLETED - Transaction is committed
    # =========================================================================
    update_journal_phase "$PHASE_COMPLETED"

    # Clean up staging and backup (swap successful)
    cleanup_staging
    cleanup_swap_backup
    delete_journal

    # =========================================================================
    # POST-COMMIT OPERATIONS (Non-critical, swap is already complete)
    # =========================================================================
    # These operations are best-effort. If they fail, the swap is still valid.

    # Restore kept agents after swap
    restore_kept_agents
    cleanup_stash

    # Update session team if active session exists (non-critical)
    # Check both user-level and project-level for session-manager.sh
    local session_mgr=""
    [[ -x "$HOME/.claude/hooks/lib/session-manager.sh" ]] && session_mgr="$HOME/.claude/hooks/lib/session-manager.sh"
    [[ -x ".claude/hooks/lib/session-manager.sh" ]] && session_mgr=".claude/hooks/lib/session-manager.sh"
    if [[ -f ".claude/sessions/.current-session" && -n "$session_mgr" ]]; then
        local current_session
        current_session=$(cat ".claude/sessions/.current-session" 2>/dev/null)
        if [[ -n "$current_session" && -f ".claude/sessions/$current_session/SESSION_CONTEXT.md" ]]; then
            # Warn user about team change
            log_warning "Active session detected: $current_session"
            log_warning "Session team will be updated to: $team_name"

            local session_file=".claude/sessions/$current_session/SESSION_CONTEXT.md"

            # Validate SESSION_CONTEXT format before mutation
            # Check for YAML frontmatter structure (opening --- on line 1)
            local first_line
            first_line=$(head -n 1 "$session_file")
            if [[ "$first_line" != "---" ]]; then
                log_warning "Cannot update session - SESSION_CONTEXT missing YAML frontmatter"
                log_warning "ACTIVE_TEAM updated but session state may be inconsistent"
            else
                # Check for active_team field exists
                if ! grep -q "^active_team:" "$session_file" 2>/dev/null; then
                    log_warning "Cannot update session - active_team field not found"
                    log_warning "ACTIVE_TEAM updated but session state may be inconsistent"
                else
                    # Safe to mutate
                    if sed -i '' "s/^active_team: .*/active_team: \"$team_name\"/" "$session_file"; then
                        log "Session team updated to: $team_name"
                    else
                        log_warning "Failed to update active_team in SESSION_CONTEXT"
                    fi
                fi
            fi
        fi
    fi

    # Update CLAUDE.md to reflect new team's agents (non-critical)
    update_claude_md "$team_name" || log_warning "CLAUDE.md update failed (non-critical)"

    # Display team roster (dynamic generation from agent frontmatter)
    generate_roster "$team_name"

    # Success - show workflow info if available
    local workflow_file="$ROSTER_HOME/teams/$team_name/workflow.yaml"
    if [[ -f "$workflow_file" ]]; then
        local entry_agent
        local phase_count
        entry_agent=$(grep -A2 "^entry_point:" "$workflow_file" | grep "agent:" | head -1 | awk '{print $2}')
        # Count phases only (lines with "agent:" under phases: section, before complexity_levels:)
        phase_count=$(sed -n '/^phases:/,/^complexity_levels:/p' "$workflow_file" | grep -c "agent:" 2>/dev/null || echo "?")
        log "Switched to $team_name ($agent_count agents, $phase_count phases, entry: $entry_agent)"
    else
        log "Switched to $team_name ($agent_count agents loaded)"
    fi

    # Restart warning - Claude Code scans agents at session startup only
    log ""
    log "NOTE: Restart Claude Code session (/exit then claude) for agent changes to take effect."
    log "      The /agents command will show stale agents until session restart."

    exit "$EXIT_SUCCESS"
}

# ============================================================================
# Reset to Skeleton Baseline
# ============================================================================

# Preview what reset would remove (for --dry-run with --reset)
preview_reset() {
    log "Dry-run: Would reset to skeleton baseline"
    echo ""

    # Check for team agents (those marked as "team" in manifest)
    echo "Team agents to remove:"
    local team_agent_count=0
    if [[ -f "$MANIFEST_FILE" ]] && [[ -d ".claude/agents" ]]; then
        local manifest
        manifest=$(read_manifest)
        for agent_file in .claude/agents/*.md; do
            [[ -f "$agent_file" ]] || continue
            local agent_name
            agent_name=$(basename "$agent_file")
            local info
            info=$(get_agent_from_manifest "$agent_name")
            local source
            source=$(echo "$info" | cut -d: -f1)
            if [[ "$source" == "team" ]]; then
                local origin
                origin=$(echo "$info" | cut -d: -f2)
                echo "  - $agent_name (from $origin)"
                ((team_agent_count++)) || true
            fi
        done
    fi
    if [[ "$team_agent_count" -eq 0 ]]; then
        echo "  (none)"
    fi

    # Check for team commands
    echo ""
    echo "Team commands to remove:"
    if [[ -f ".claude/commands/.team-commands" ]]; then
        while IFS= read -r cmd; do
            [[ -z "$cmd" ]] && continue
            echo "  - $cmd"
        done < ".claude/commands/.team-commands"
    else
        echo "  (none)"
    fi

    # Check for team skills
    echo ""
    echo "Team skills to remove:"
    if [[ -f ".claude/skills/.team-skills" ]]; then
        while IFS= read -r skill; do
            [[ -z "$skill" ]] && continue
            echo "  - $skill"
        done < ".claude/skills/.team-skills"
    else
        echo "  (none)"
    fi

    # Check for team hooks
    echo ""
    echo "Team hooks to remove:"
    if [[ -f ".claude/hooks/.team-hooks" ]]; then
        while IFS= read -r hook; do
            [[ -z "$hook" ]] && continue
            echo "  - $hook"
        done < ".claude/hooks/.team-hooks"
    else
        echo "  (none)"
    fi

    # Show what will be cleared
    echo ""
    if [[ -f ".claude/ACTIVE_TEAM" ]]; then
        local current
        current=$(cat .claude/ACTIVE_TEAM 2>/dev/null | tr -d '[:space:]')
        echo "Would clear: ACTIVE_TEAM (currently: $current)"
    fi
    echo "Would regenerate: CLAUDE.md (skeleton baseline)"

    echo ""
    echo "No changes made (--dry-run mode)"
}

# Remove team agents only (preserve user-added agents)
remove_team_agents() {
    log_debug "Removing team agents (preserving user-added)"

    if [[ ! -f "$MANIFEST_FILE" ]] || [[ ! -d ".claude/agents" ]]; then
        log_debug "No manifest or agents directory"
        return 0
    fi

    local removed=0
    local manifest
    manifest=$(read_manifest)

    for agent_file in .claude/agents/*.md; do
        [[ -f "$agent_file" ]] || continue
        local agent_name
        agent_name=$(basename "$agent_file")

        local info
        info=$(get_agent_from_manifest "$agent_name")
        local source
        source=$(echo "$info" | cut -d: -f1)

        if [[ "$source" == "team" ]]; then
            rm -f "$agent_file"
            log_debug "Removed team agent: $agent_name"
            ((removed++)) || true
        else
            log_debug "Preserved: $agent_name (source: $source)"
        fi
    done

    log "Removed: $removed team agent(s)"

    # Clear manifest (will be regenerated if user agents remain)
    rm -f "$MANIFEST_FILE"

    # Regenerate manifest for remaining user agents
    if [[ -d ".claude/agents" ]]; then
        local remaining
        remaining=$(find .claude/agents -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        if [[ "$remaining" -gt 0 ]]; then
            log_debug "Regenerating manifest for $remaining remaining agent(s)"
            init_manifest_from_existing
        fi
    fi
}

# Regenerate CLAUDE.md for skeleton baseline (no active team)
regenerate_skeleton_claude_md() {
    local claude_md=".claude/CLAUDE.md"

    [[ -f "$claude_md" ]] || return 0

    log_debug "Regenerating CLAUDE.md for skeleton baseline"

    # Write replacement content to temp files (avoids awk multiline string issues)
    local qs_file ac_file
    qs_file=$(mktemp)
    ac_file=$(mktemp)

    cat > "$qs_file" << 'QSEOF'
## Quick Start

No team currently active. Available commands:

| Command | Description |
|---------|-------------|
| `/team <pack-name>` | Switch to a team pack |
| `/team --list` | List available teams |
| `/consult` | Get guidance on which team to use |

**Get started**: Run `/consult` to find the right team for your task.
QSEOF

    cat > "$ac_file" << 'ACEOF'
## Agent Configurations

No team agents loaded. Switch to a team pack to load agents.
ACEOF

    # Update sections (handles PRESERVE comment and ## heading on separate lines)
    if grep -q "<!-- PRESERVE: satellite-owned" "$claude_md" 2>/dev/null; then
        awk -v qs_file="$qs_file" -v ac_file="$ac_file" '
        # Track when we see a PRESERVE comment
        /<!-- PRESERVE: satellite-owned/ {
            preserve_line = 1
            print $0
            next
        }
        # If previous line was PRESERVE and this is Quick Start, replace section
        preserve_line && /^## Quick Start/ {
            in_qs_section = 1
            preserve_line = 0
            while ((getline line < qs_file) > 0) print line
            close(qs_file)
            next
        }
        # If previous line was PRESERVE and this is Agent Configurations, replace section
        preserve_line && /^## Agent Configurations/ {
            in_ac_section = 1
            preserve_line = 0
            while ((getline line < ac_file) > 0) print line
            close(ac_file)
            next
        }
        # Reset preserve flag if next line is not a known section
        preserve_line {
            preserve_line = 0
        }
        # End Quick Start section at next section marker or ## heading
        in_qs_section && /^<!-- (SYNC|PRESERVE):/ {
            in_qs_section = 0
        }
        in_qs_section && /^##[^#]/ {
            in_qs_section = 0
        }
        # End Agent Configurations section at next section marker or ## heading
        in_ac_section && /^<!-- (SYNC|PRESERVE):/ {
            in_ac_section = 0
        }
        in_ac_section && /^##[^#]/ {
            in_ac_section = 0
        }
        # Print lines not in replaced sections
        !in_qs_section && !in_ac_section { print }
        ' "$claude_md" > "$claude_md.tmp" && mv "$claude_md.tmp" "$claude_md"
    fi

    # Cleanup temp files
    rm -f "$qs_file" "$ac_file"

    log_debug "CLAUDE.md regenerated for skeleton baseline"
}

# Perform reset to skeleton baseline
perform_reset() {
    log_debug "Starting reset to skeleton baseline"

    # Validate we're in a project
    validate_project

    # Check if dry-run mode
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        preview_reset
        return 0
    fi

    # Get current team for reporting
    local current_team=""
    if [[ -f ".claude/ACTIVE_TEAM" ]]; then
        current_team=$(cat .claude/ACTIVE_TEAM 2>/dev/null | tr -d '[:space:]')
    fi

    if [[ -z "$current_team" ]]; then
        log "No team active. Already at skeleton baseline."
        return 0
    fi

    log "Resetting from $current_team to skeleton baseline..."

    # Backup current state
    backup_current_agents

    # Remove team resources using existing functions
    remove_team_agents
    remove_team_commands
    remove_team_skills
    remove_team_hooks

    # Clear ACTIVE_TEAM
    rm -f ".claude/ACTIVE_TEAM"
    rm -f ".claude/ACTIVE_WORKFLOW.yaml"
    log "Cleared: ACTIVE_TEAM"

    # Regenerate CLAUDE.md
    regenerate_skeleton_claude_md
    log "Regenerated: CLAUDE.md (skeleton baseline)"

    echo ""
    log "Reset complete. Skeleton baseline active."
    log ""
    log "To switch to a team: $ROSTER_HOME/swap-team.sh <team-name>"
    log "To list teams:       $ROSTER_HOME/swap-team.sh --list"
}

# Main entry point
main() {
    local team_name=""

    # Check environment variable for auto-recover
    if [[ "${ROSTER_AUTO_RECOVER:-0}" == "1" ]]; then
        AUTO_RECOVER=1
    fi

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            "")
                shift
                ;;
            --list|-l)
                list_teams
                ;;
            --help|-h)
                usage
                exit "$EXIT_SUCCESS"
                ;;
            --keep-all)
                ORPHAN_MODE="keep"
                shift
                ;;
            --remove-all)
                ORPHAN_MODE="remove"
                shift
                ;;
            --promote-all)
                ORPHAN_MODE="promote"
                shift
                ;;
            --update|-u|--refresh|-r)
                UPDATE_MODE=1
                # Deprecation warning for --refresh
                if [[ "$1" == "--refresh" || "$1" == "-r" ]]; then
                    log_warning "Flag --refresh/-r is deprecated. Use --update/-u instead."
                fi
                shift
                ;;
            --dry-run)
                DRY_RUN_MODE=1
                UPDATE_MODE=1  # dry-run implies update
                shift
                ;;
            --reset|--skeleton)
                RESET_MODE=1
                shift
                ;;
            --auto-recover)
                AUTO_RECOVER=1
                shift
                ;;
            --recover)
                RECOVER_MODE=1
                shift
                ;;
            --verify)
                VERIFY_MODE=1
                shift
                ;;
            -*)
                log_error "Unknown option: $1"
                usage
                exit "$EXIT_INVALID_ARGS"
                ;;
            *)
                if [[ -z "$team_name" ]]; then
                    team_name="$1"
                else
                    log_error "Multiple team names specified"
                    usage
                    exit "$EXIT_INVALID_ARGS"
                fi
                shift
                ;;
        esac
    done

    # Handle --verify mode (takes precedence)
    if [[ "$VERIFY_MODE" -eq 1 ]]; then
        verify_state_consistency
        exit $?
    fi

    # Handle --recover mode (takes precedence)
    if [[ "$RECOVER_MODE" -eq 1 ]]; then
        if [[ ! -f "$JOURNAL_FILE" ]]; then
            log "No interrupted swap detected. State is clean."
            exit "$EXIT_SUCCESS"
        fi
        # Force interactive recovery
        if [[ -t 0 ]]; then
            check_journal_recovery
        else
            log_error "Recovery mode requires interactive terminal"
            log "Use --auto-recover for non-interactive recovery"
            exit "$EXIT_INVALID_ARGS"
        fi
        exit "$EXIT_SUCCESS"
    fi

    # Handle reset mode (takes precedence)
    if [[ "$RESET_MODE" -eq 1 ]]; then
        perform_reset
        exit "$EXIT_SUCCESS"
    fi

    # Handle the command
    if [[ -z "$team_name" ]]; then
        if [[ "$UPDATE_MODE" -eq 1 ]]; then
            # Update mode without team name - update current team
            if [[ -f ".claude/ACTIVE_TEAM" ]]; then
                team_name=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')
                log "Updating current team: $team_name"
                perform_swap "$team_name"
            else
                log_error "No team active. Specify a team name to update."
                exit "$EXIT_INVALID_ARGS"
            fi
        else
            # No team specified and not update mode - query current team
            query_current_team
        fi
    else
        # Team pack name - perform swap
        perform_swap "$team_name"
    fi
}

# Run main if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
