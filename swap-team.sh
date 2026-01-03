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

# roster-sync integration modes
SYNC_FIRST_MODE=0     # --sync-first: run roster-sync before team apply
AUTO_SYNC_MODE=0      # --auto-sync: conditionally sync if roster has updates

# Orphan backup cleanup modes
CLEANUP_ORPHANS_MODE=0    # --cleanup-orphans: manual cleanup of old orphan backups
AUTO_CLEANUP_MODE=0       # --auto-cleanup: automatic cleanup during swap

# Interactive mode: "auto" (default), "yes" (force prompts), "no" (force auto-select)
# --interactive forces prompts even without TTY
# --no-interactive forces auto-selection even with TTY
INTERACTIVE_MODE="auto"

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

# Check if running in interactive mode
# Returns: 0 if interactive (prompts allowed), 1 if non-interactive
# Logic:
#   --interactive forces interactive (returns 0) even without TTY
#   --no-interactive forces non-interactive (returns 1) even with TTY
#   "auto" (default) uses TTY detection: [[ -t 0 ]]
is_interactive() {
    case "$INTERACTIVE_MODE" in
        "yes")
            return 0
            ;;
        "no")
            return 1
            ;;
        "auto"|*)
            [[ -t 0 ]]
            return $?
            ;;
    esac
}

# ============================================================================
# Transaction Safety Functions
# ============================================================================

# Transaction infrastructure functions (write_atomic, journal CRUD, staging,
# and backup operations) are now provided by lib/team/team-transaction.sh
# See that module for implementation details.

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

    # Delete journal first to mark transaction complete, then best-effort cleanup
    # This prevents orphaned journal if process dies during cleanup
    delete_journal
    cleanup_staging
    cleanup_swap_backup

    # Clean up agent stash to prevent orphan accumulation on failed swaps
    # Without this, restore_kept_agents runs on every subsequent swap attempt
    cleanup_stash

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
            delete_journal
            cleanup_staging
            cleanup_stash  # Clean up any stashed agents from orphan handling
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
            delete_journal
            cleanup_staging
            cleanup_swap_backup
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
        delete_journal
        cleanup_staging
        cleanup_swap_backup
        return 0
    fi

    # Check if we're past point-of-no-return in COMMITTING phase
    local past_ponr=false
    if [[ "$phase" == "$PHASE_COMMITTING" ]] && is_past_point_of_no_return; then
        past_ponr=true
        log_warning "Past point-of-no-return. Must complete forward (cannot rollback)."
    fi

    # Check backup validity (only needed if rollback is an option)
    local backup_valid=false
    if [[ "$past_ponr" != "true" ]]; then
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
    fi

    # Interactive vs non-interactive recovery
    if is_interactive; then
        prompt_recovery_action "$source" "$target" "$phase" "$past_ponr"
    elif [[ "$AUTO_RECOVER" -eq 1 ]]; then
        if [[ "$past_ponr" == "true" ]]; then
            log "Auto-recovery: completing forward (past point-of-no-return)..."
            complete_partial_commit "$target"
        else
            log "Auto-recovery enabled. Rolling back..."
            rollback_swap
        fi
        return 0
    else
        log_error "Non-interactive mode. Use --auto-recover to enable automatic rollback"
        log "Or manually resolve with: swap-team.sh --recover"
        exit "$EXIT_RECOVERY_REQUIRED"
    fi

    return 0
}

# Interactive recovery prompt
# Parameters:
#   $1 - source: Source team
#   $2 - target: Target team
#   $3 - phase: Current phase
#   $4 - past_ponr: "true" if past point-of-no-return
prompt_recovery_action() {
    local source="$1" target="$2" phase="$3" past_ponr="${4:-false}"

    echo ""
    if [[ "$past_ponr" == "true" ]]; then
        echo "Recovery Options (past point-of-no-return):"
        echo "  [c] Complete swap to $target (recommended)"
        echo "  [a] Abort (leave as-is for manual recovery)"
        echo ""
        echo "Note: Rollback is not available because ACTIVE_TEAM was already written."
        echo ""

        local choice
        while true; do
            read -r -p "Choice [c/a]: " choice < /dev/tty
            case "$choice" in
                c|C)
                    complete_partial_commit "$target"
                    return $?
                    ;;
                a|A)
                    log "Aborted. Journal preserved for manual recovery."
                    log_warning "System may be in inconsistent state."
                    exit "$EXIT_SUCCESS"
                    ;;
                *)
                    echo "Invalid choice. Enter c or a."
                    ;;
            esac
        done
    else
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
    fi
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
        "$PHASE_VERIFYING")
            # Verification phase - can restart from staging
            if [[ -d "$STAGING_DIR" ]]; then
                log "Attempting to complete commit from staging..."
                commit_staged_resources "$target_team"
                return $?
            else
                log_error "Staging directory missing. Cannot continue."
                return 1
            fi
            ;;
        "$PHASE_COMMITTING")
            # Partial commit - check point-of-no-return
            recover_partial_commit "$target_team"
            return $?
            ;;
        *)
            log_error "Cannot continue from phase: $phase"
            return 1
            ;;
    esac
}

# Recover from a partial COMMITTING phase
# Checks point-of-no-return and either rolls back or completes forward
# Parameters:
#   $1 - target_team: Team being swapped to
# Returns: 0 on success, 1 on failure
recover_partial_commit() {
    local target_team="$1"

    log "Recovering from partial commit..."

    # Check if we're past the point-of-no-return (ACTIVE_TEAM written)
    if is_past_point_of_no_return; then
        log "Past point-of-no-return. Completing swap forward..."
        complete_partial_commit "$target_team"
        return $?
    else
        log "Before point-of-no-return. Safe to rollback."
        # Show which steps completed for diagnostics
        local incomplete
        incomplete=$(get_incomplete_commit_steps)
        if [[ -n "$incomplete" ]]; then
            log_debug "Incomplete steps: $incomplete"
        fi
        return 1  # Signal that rollback is the recommended action
    fi
}

# Complete a partial commit after point-of-no-return
# Runs only the steps that haven't completed yet
# Parameters:
#   $1 - target_team: Team being swapped to
# Returns: 0 on success, 1 on failure
complete_partial_commit() {
    local target_team="$1"

    log "Completing partial commit..."

    # Run incomplete steps in order
    # Note: agents and workflow are handled by commit_staged_resources
    # We only need to handle steps after that point

    if ! is_commit_step_done "$COMMIT_STEP_COMMANDS"; then
        log "Completing: commands sync"
        swap_commands "$target_team"
        mark_commit_step "$COMMIT_STEP_COMMANDS"
    fi

    if ! is_commit_step_done "$COMMIT_STEP_SKILLS"; then
        log "Completing: skills sync"
        swap_skills "$target_team"
        mark_commit_step "$COMMIT_STEP_SKILLS"
    fi

    if ! is_commit_step_done "$COMMIT_STEP_SHARED_SKILLS"; then
        log "Completing: shared skills sync"
        sync_shared_skills
        mark_commit_step "$COMMIT_STEP_SHARED_SKILLS"
    fi

    if ! is_commit_step_done "$COMMIT_STEP_HOOKS"; then
        log "Completing: hooks sync"
        swap_hooks "$target_team"
        mark_commit_step "$COMMIT_STEP_HOOKS"
    fi

    if ! is_commit_step_done "$COMMIT_STEP_HOOK_REGISTRATIONS"; then
        log "Completing: hook registrations"
        swap_hook_registrations "$target_team"
        mark_commit_step "$COMMIT_STEP_HOOK_REGISTRATIONS"
    fi

    if ! is_commit_step_done "$COMMIT_STEP_MANIFEST"; then
        log "Completing: manifest write"
        write_manifest "$target_team"
        mark_commit_step "$COMMIT_STEP_MANIFEST"
    fi

    # ACTIVE_TEAM should already be written (that's what got us past point-of-no-return)
    # But verify and fix if somehow missing
    if ! is_commit_step_done "$COMMIT_STEP_ACTIVE_TEAM"; then
        log_warning "ACTIVE_TEAM step not marked but we're past point-of-no-return"
        if [[ -f ".claude/ACTIVE_TEAM" ]]; then
            local current_team
            current_team=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')
            if [[ "$current_team" == "$target_team" ]]; then
                mark_commit_step "$COMMIT_STEP_ACTIVE_TEAM"
            fi
        fi
    fi

    # Mark as completed
    update_journal_phase "$PHASE_COMPLETED"

    # Delete journal first to mark transaction complete, then best-effort cleanup
    # This prevents orphaned journal if process dies during cleanup
    delete_journal
    cleanup_staging
    cleanup_swap_backup

    log "Partial commit recovery complete"
    return 0
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

    # Initialize commit step tracking for recovery
    init_commit_steps || {
        log_error "Failed to initialize commit step tracking"
        return 1
    }

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
    mark_commit_step "$COMMIT_STEP_AGENTS"

    # 3. Move workflow atomically
    if [[ -f "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" ]]; then
        mv "$STAGING_DIR/ACTIVE_WORKFLOW.yaml" .claude/ACTIVE_WORKFLOW.yaml || {
            log_warning "Failed to commit workflow"
        }
    fi
    mark_commit_step "$COMMIT_STEP_WORKFLOW"

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

# Source transaction infrastructure for team swaps
source "$ROSTER_HOME/lib/team/team-transaction.sh"

# Source team resource operations
source "$ROSTER_HOME/lib/team/team-resource.sh"

# Source hook registration infrastructure
source "$ROSTER_HOME/lib/team/team-hooks-registration.sh"

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

    # Check if we're in an interactive mode
    if ! is_interactive; then
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
  --reset        Reset to baseline (remove all team resources)
  --verify       Verify current state consistency
  --recover      Interactive recovery from interrupted swap
  (no args)      Show current active team

Options:
  --update, -u       Update agents from roster (even if already on team)
  --refresh, -r      [DEPRECATED] Alias for --update
  --dry-run          Preview changes without applying
  --keep-all         Preserve orphan agents in project
  --remove-all       Remove orphan agents/commands/skills/hooks (backup available)
  --promote-all      Move orphan agents to ~/.claude/agents/
  --auto-recover     Automatically rollback if interrupted swap detected (for CI/CD)
  --sync-first       Run roster-sync before applying team (waterfall pattern)
  --auto-sync        Conditionally sync if roster has updates available
  --cleanup-orphans  Clean up old orphan backup directories (keep last 3)
  --auto-cleanup     Automatically clean orphan backups during swap
  --interactive      Force interactive prompts even without TTY (for containers)
  --no-interactive   Force non-interactive mode even with TTY (alias: --batch)

When switching teams interactively, you'll be prompted for each orphan agent
(agents in current team but not in target team). In non-interactive mode
(scripts, CI), you must specify one of the orphan handling flags.

Interactive Mode:
  By default, TTY auto-detection determines prompting behavior. This can be
  unreliable in containers, CI, or when piping. Use --interactive to force
  prompts (requires /dev/tty access), or --no-interactive to skip prompts
  and require explicit flags like --keep-all for orphan handling.

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
  ./swap-team.sh --reset                # Reset to baseline (no team)
  ./swap-team.sh --reset --dry-run      # Preview what reset would remove
  ./swap-team.sh --verify               # Check state consistency
  ./swap-team.sh --recover              # Recover from interrupted swap
  ./swap-team.sh --auto-recover dev-pack # CI/CD mode with auto-rollback
  ./swap-team.sh dev-pack --sync-first  # Sync infrastructure then apply team
  ./swap-team.sh dev-pack --auto-sync   # Sync only if roster has updates
  ./swap-team.sh --cleanup-orphans      # Clean up old orphan backups
  ./swap-team.sh dev-pack --auto-cleanup # Swap team and auto-clean orphan backups
  echo "team" | ./swap-team.sh --interactive --keep-all # Force prompts in pipe
  ./swap-team.sh dev-pack --no-interactive --keep-all   # Skip TTY check in CI

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

# Validate workflow.yaml schema for a team pack
# Checks required fields: name, workflow_type, phases
# Returns 0 if valid or file doesn't exist, 1 if validation fails
validate_workflow_yaml() {
    local team_name="$1"
    local workflow_file="$ROSTER_HOME/teams/$team_name/workflow.yaml"

    # Skip if workflow.yaml doesn't exist (optional file)
    if [[ ! -f "$workflow_file" ]]; then
        log_debug "No workflow.yaml for $team_name (optional)"
        return 0
    fi

    log_debug "Validating workflow.yaml schema for $team_name"

    # Check required field: name
    if ! grep -q "^name:" "$workflow_file"; then
        log_error "workflow.yaml missing required field: name"
        return 1
    fi

    # Check required field: workflow_type
    if ! grep -q "^workflow_type:" "$workflow_file"; then
        log_error "workflow.yaml missing required field: workflow_type"
        return 1
    fi

    # Check required field: phases
    if ! grep -q "^phases:" "$workflow_file"; then
        log_error "workflow.yaml missing required field: phases"
        return 1
    fi

    # Validate that phases is a list (has at least one item with "- name:")
    if ! grep -A 1 "^phases:" "$workflow_file" | grep -q "  - name:"; then
        log_error "workflow.yaml phases must be a non-empty list"
        return 1
    fi

    log_debug "workflow.yaml schema validation passed"
    return 0
}

# Validate orchestrator.yaml schema for a team pack
# Checks required fields: team, routing
# Returns 0 if valid or file doesn't exist, 1 if validation fails
validate_orchestrator_yaml() {
    local team_name="$1"
    local orchestrator_file="$ROSTER_HOME/teams/$team_name/orchestrator.yaml"

    # Skip if orchestrator.yaml doesn't exist (optional file)
    if [[ ! -f "$orchestrator_file" ]]; then
        log_debug "No orchestrator.yaml for $team_name (optional)"
        return 0
    fi

    log_debug "Validating orchestrator.yaml schema for $team_name"

    # Check required field: team
    if ! grep -q "^team:" "$orchestrator_file"; then
        log_error "orchestrator.yaml missing required field: team"
        return 1
    fi

    # Check required nested field: team.name
    if ! grep -A 3 "^team:" "$orchestrator_file" | grep -q "  name:"; then
        log_error "orchestrator.yaml missing required field: team.name"
        return 1
    fi

    # Check required field: routing
    if ! grep -q "^routing:" "$orchestrator_file"; then
        log_error "orchestrator.yaml missing required field: routing"
        return 1
    fi

    log_debug "orchestrator.yaml schema validation passed"
    return 0
}

# Validate team pack schemas before swap
# Called during PHASE_PREPARING to catch schema issues early
# Returns 0 if all schemas valid, 1 if any validation fails
validate_team_schemas() {
    local team_name="$1"

    log_debug "Validating team schemas for $team_name"

    # Validate workflow.yaml if present
    if ! validate_workflow_yaml "$team_name"; then
        log_error "Schema validation failed for workflow.yaml"
        log_error "Team pack $team_name has invalid configuration"
        return 1
    fi

    # Validate orchestrator.yaml if present
    if ! validate_orchestrator_yaml "$team_name"; then
        log_error "Schema validation failed for orchestrator.yaml"
        log_error "Team pack $team_name has invalid configuration"
        return 1
    fi

    log_debug "Team schema validation passed"
    return 0
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
# Team Resource Wrapper Functions (for backward compatibility)
# ============================================================================
# These wrappers call the generic functions from lib/team/team-resource.sh

# Backup wrappers
backup_team_commands() { backup_team_resource "commands" ".claude/commands" ".team-commands" "f"; }
backup_team_skills()   { backup_team_resource "skills" ".claude/skills" ".team-skills" "d"; }
backup_team_hooks()    { backup_team_resource "hooks" ".claude/hooks" ".team-hooks" "f"; }

# Remove wrappers
remove_team_commands() { remove_team_resource "commands" ".claude/commands" ".team-commands" "f"; }
remove_team_skills()   { remove_team_resource "skills" ".claude/skills" ".team-skills" "d"; }
remove_team_hooks()    { remove_team_resource "hooks" ".claude/hooks" ".team-hooks" "f"; }

# Team membership check wrappers
is_team_command() { is_resource_from_team "$1" "commands" "f"; }
is_team_skill()   { is_resource_from_team "$1" "skills" "d"; }
is_team_hook()    { is_resource_from_team "$1" "hooks" "f"; }

# Team origin lookup wrappers
get_command_team() { get_resource_team "$1" "commands" "f"; }
get_skill_team()   { get_resource_team "$1" "skills" "d"; }
get_hook_team()    { get_resource_team "$1" "hooks" "f"; }

# ============================================================================
# Team Commands Functions
# ============================================================================

# Check for collisions between user commands and team commands
# Warns about conflicts but allows user commands to win (non-blocking)
# Called before team commands are staged to inform user of potential conflicts
check_user_command_collisions() {
    local team_name="$1"
    local source_dir="$ROSTER_HOME/teams/$team_name/commands"

    # Skip if team has no commands directory
    if [[ ! -d "$source_dir" ]]; then
        log_debug "No team commands to check for collisions"
        return 0
    fi

    # Skip if no user commands exist
    if [[ ! -d "$HOME/.claude/commands" ]]; then
        log_debug "No user commands directory, skipping collision check"
        return 0
    fi

    local collision_count=0
    local collisions=()

    # Check each team command against user commands
    for team_cmd in "$source_dir"/*.md; do
        [[ -f "$team_cmd" ]] || continue

        local cmd_name
        cmd_name=$(basename "$team_cmd")

        # Check if user command with same name exists
        if [[ -f "$HOME/.claude/commands/$cmd_name" ]]; then
            collisions+=("$cmd_name")
            ((collision_count++))
        fi
    done

    # Warn about collisions if any found
    if [[ $collision_count -gt 0 ]]; then
        log_warning "Command collision(s) detected: $collision_count command(s)"
        log_warning "User commands (in ~/.claude/commands/) will take precedence:"
        for cmd in "${collisions[@]}"; do
            log_warning "  - $cmd"
        done
        log_warning "Team commands with same names will be skipped during sync"
    fi

    return 0
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

# Sync team-specific skills to project
# Team skills are copied to .claude/skills/ with a marker file
# Skills from team layer overlay baseline skills (team wins on collision)
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

        # Check for collision with existing baseline skill
        # Team wins: overwrite with warning
        if [[ -d ".claude/skills/$skill_name" ]] && ! grep -q "^$skill_name$" "$marker_file" 2>/dev/null; then
            # Exists but not from team - this is a baseline skill
            log_warning "Team skill $skill_name overrides baseline skill"
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

        # Check for collision with existing baseline skill
        # Shared wins over baseline (but loses to team, checked above)
        if [[ -d ".claude/skills/$skill_name" ]]; then
            log_debug "Shared skill $skill_name overrides baseline skill"
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
# Orphan Backup Cleanup Functions
# ============================================================================

# Clean up old orphan backup directories, keeping only the last N backups per type
# Usage: cleanup_orphan_backups [keep_count]
cleanup_orphan_backups() {
    local keep_count="${1:-3}"  # Default: keep last 3 backups per type
    local backup_types=("agents" "commands" "skills" "hooks")
    local cleaned_count=0

    log_debug "Cleaning orphan backups (keeping last $keep_count per type)"

    for backup_type in "${backup_types[@]}"; do
        local backup_base_dir=".claude/${backup_type}.orphan-backup"

        # Skip if backup directory doesn't exist
        if [[ ! -d "$backup_base_dir" ]]; then
            log_debug "No orphan backups for $backup_type"
            continue
        fi

        # Find all timestamped backup subdirectories, sorted by modification time (newest first)
        # Format: .claude/{type}.orphan-backup/{timestamp}-{team}/
        # Use portable approach compatible with both GNU and BSD find
        local backup_dirs=()
        while IFS= read -r backup_dir; do
            backup_dirs+=("$backup_dir")
        done < <(find "$backup_base_dir" -mindepth 1 -maxdepth 1 -type d -exec stat -f '%m %N' {} \; 2>/dev/null | sort -rn | cut -d' ' -f2- || \
                 find "$backup_base_dir" -mindepth 1 -maxdepth 1 -type d -exec stat -c '%Y %n' {} \; 2>/dev/null | sort -rn | cut -d' ' -f2-)

        local total_backups=${#backup_dirs[@]}

        # Skip if we have fewer backups than the keep count
        if [[ $total_backups -le $keep_count ]]; then
            log_debug "Only $total_backups ${backup_type} backup(s), keeping all"
            continue
        fi

        # Remove backups beyond the keep count
        local to_remove=$((total_backups - keep_count))
        log_debug "Found $total_backups ${backup_type} backup(s), removing $to_remove oldest"

        for ((i = keep_count; i < total_backups; i++)); do
            local old_backup="${backup_dirs[$i]}"
            if [[ -d "$old_backup" ]]; then
                rm -rf "$old_backup"
                ((cleaned_count++))
                log_debug "Removed old backup: $(basename "$old_backup")"
            fi
        done
    done

    if [[ $cleaned_count -gt 0 ]]; then
        log "Cleaned up $cleaned_count old orphan backup(s)"
    else
        log_debug "No orphan backups to clean up"
    fi

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
    # Handle both team mode ("This project uses...") and baseline mode ("No team currently active")
    local before_table_line
    before_table_line=$(grep -n "^This project uses a [0-9]*-agent" "$claude_md" 2>/dev/null | head -1 | cut -d: -f1 || true)
    if [[ -z "$before_table_line" ]]; then
        # Check for baseline mode pattern
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
    # Handle both team mode ("**New here") and baseline mode ("**Get started")
    local table_end_line
    table_end_line=$(grep -n "^\*\*New here" "$claude_md" 2>/dev/null | head -1 | cut -d: -f1 || true)
    if [[ -z "$table_end_line" ]]; then
        # Check for baseline mode pattern
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
    local orphan_commands
    orphan_commands=$(detect_resource_orphans "commands" ".claude/commands" "$team_name" "f" "*.md")
    if [[ -n "$orphan_commands" ]]; then
        while IFS=: read -r cmd_name origin_team; do
            echo "  ? $cmd_name (from $origin_team)"
        done <<< "$orphan_commands"
    else
        echo "  (none)"
    fi

    # Check for orphan skills (skills from other teams)
    echo ""
    echo "Skill orphans (from other teams):"
    local orphan_skills
    orphan_skills=$(detect_resource_orphans "skills" ".claude/skills" "$team_name" "d" "*/")
    if [[ -n "$orphan_skills" ]]; then
        while IFS=: read -r skill_name origin_team; do
            echo "  ? $skill_name (from $origin_team)"
        done <<< "$orphan_skills"
    else
        echo "  (none)"
    fi

    # Check for orphan hooks (hooks from other teams)
    echo ""
    echo "Hook orphans (from other teams):"
    local orphan_hooks
    orphan_hooks=$(detect_resource_orphans "hooks" ".claude/hooks" "$team_name" "f" "*")
    if [[ -n "$orphan_hooks" ]]; then
        while IFS=: read -r hook_name origin_team; do
            echo "  ? $hook_name (from $origin_team)"
        done <<< "$orphan_hooks"
    else
        echo "  (none)"
    fi

    # Summary of orphans
    local cmd_count skill_count hook_count
    cmd_count=$(echo "$orphan_commands" | grep -c . || true)
    skill_count=$(echo "$orphan_skills" | grep -c . || true)
    hook_count=$(echo "$orphan_hooks" | grep -c . || true)
    # Default to 0 if grep returned nothing
    cmd_count=${cmd_count:-0}
    skill_count=${skill_count:-0}
    hook_count=${hook_count:-0}
    local total_orphans=$((cmd_count + skill_count + hook_count))
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

# ============================================================================
# roster-sync Integration (Waterfall Pattern)
# ============================================================================

# Check if roster-sync is available
# Returns: 0 if available, 1 if not
roster_sync_available() {
    local roster_sync="$ROSTER_HOME/roster-sync"
    [[ -x "$roster_sync" ]]
}

# Check if roster has updates compared to manifest
# Uses roster-sync's logic via sync-core.sh if available
# Returns: 0 if updates available, 1 if up to date
roster_has_updates() {
    # Source sync-core if not already loaded (for roster_has_updates function)
    local sync_lib="$ROSTER_HOME/lib/sync"
    if [[ -d "$sync_lib" ]]; then
        # Need to source dependencies first
        if [[ ! -f "$sync_lib/sync-config.sh" ]]; then
            log_debug "sync-config.sh not found, assuming updates available"
            return 0
        fi

        # Source in order
        source "$sync_lib/sync-config.sh" 2>/dev/null || true
        source "$sync_lib/sync-checksum.sh" 2>/dev/null || true
        source "$sync_lib/sync-manifest.sh" 2>/dev/null || true
        source "$sync_lib/sync-core.sh" 2>/dev/null || true

        # Check if roster_has_updates function exists now
        if type roster_has_updates_internal &>/dev/null 2>&1; then
            roster_has_updates_internal
            return $?
        fi
    fi

    # Fallback: compare roster git commit with manifest
    local manifest_file=".claude/.cem/manifest.json"
    if [[ ! -f "$manifest_file" ]]; then
        log_debug "No manifest found, assuming updates available"
        return 0  # No manifest means first sync needed
    fi

    local current_commit manifest_commit
    current_commit=$(git -C "$ROSTER_HOME" rev-parse HEAD 2>/dev/null)
    manifest_commit=$(jq -r '.roster.commit // empty' "$manifest_file" 2>/dev/null)

    if [[ -z "$current_commit" ]]; then
        log_debug "Cannot determine roster commit"
        return 0  # Assume updates available
    fi

    if [[ "$current_commit" != "$manifest_commit" ]]; then
        log_debug "Roster has updates: $manifest_commit -> $current_commit"
        return 0  # Updates available
    fi

    log_debug "Roster is up to date"
    return 1  # Up to date
}

# Run roster-sync before team apply (waterfall pattern)
# Usage: run_roster_sync_waterfall [--force]
# Returns: 0 on success, non-zero on failure
run_roster_sync_waterfall() {
    local force_flag=""
    [[ "${1:-}" == "--force" ]] && force_flag="--force"

    local roster_sync="$ROSTER_HOME/roster-sync"

    if [[ ! -x "$roster_sync" ]]; then
        log_warning "roster-sync not found, skipping infrastructure sync"
        return 0
    fi

    log "Syncing infrastructure before team apply..."

    # Run roster-sync sync (without --refresh to avoid recursion)
    local sync_output
    local sync_exit
    if [[ -n "$force_flag" ]]; then
        sync_output=$("$roster_sync" sync $force_flag 2>&1) || sync_exit=$?
    else
        sync_output=$("$roster_sync" sync 2>&1) || sync_exit=$?
    fi
    sync_exit=${sync_exit:-0}

    if [[ $sync_exit -ne 0 ]]; then
        log_error "roster-sync failed (exit $sync_exit)"
        echo "$sync_output" >&2
        return $sync_exit
    fi

    log_debug "roster-sync completed successfully"
    # Show summary if there were updates
    if [[ "$sync_output" == *"Updated:"* || "$sync_output" == *"Merging:"* ]]; then
        log "Infrastructure sync completed"
    else
        log "Infrastructure already up to date"
    fi

    return 0
}

# Update CEM manifest team section after swap
# Usage: update_cem_manifest_team "team_name"
update_cem_manifest_team() {
    local team_name="$1"
    local manifest_file=".claude/.cem/manifest.json"

    # Skip if no manifest exists (roster-sync not initialized)
    if [[ ! -f "$manifest_file" ]]; then
        log_debug "No CEM manifest found, skipping team section update"
        return 0
    fi

    # Compute team directory checksum for staleness detection
    local team_dir="$ROSTER_HOME/teams/$team_name"
    local team_checksum=""
    if [[ -d "$team_dir" ]]; then
        # Use tar to create stable checksum of team directory contents
        team_checksum=$(find "$team_dir" -type f -exec cat {} \; 2>/dev/null | shasum -a 256 | awk '{print $1}')
    fi

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Update manifest team section using jq
    local updated
    updated=$(jq \
        --arg name "$team_name" \
        --arg checksum "$team_checksum" \
        --arg roster_path "$team_dir" \
        --arg timestamp "$timestamp" '
        .team = {
            name: $name,
            checksum: (if $checksum != "" then $checksum else null end),
            last_refresh: $timestamp,
            roster_path: $roster_path
        }
    ' "$manifest_file") || {
        log_warning "Failed to update CEM manifest team section"
        return 1
    }

    # Write atomically
    local temp_file="${manifest_file}.tmp.$$"
    echo "$updated" > "$temp_file" || {
        rm -f "$temp_file"
        log_warning "Failed to write updated manifest"
        return 1
    }
    mv "$temp_file" "$manifest_file" || {
        rm -f "$temp_file"
        log_warning "Failed to rename manifest"
        return 1
    }

    log_debug "Updated CEM manifest team section: $team_name"
    return 0
}

# Main swap orchestration
perform_swap() {
    local team_name="$1"

    log_debug "Starting swap to $team_name"

    # Set up signal handlers for graceful interruption
    setup_signal_handlers

    # Check for recovery from interrupted swap
    check_journal_recovery

    # =========================================================================
    # WATERFALL SYNC: roster-sync integration (--sync-first or --auto-sync)
    # =========================================================================
    if [[ "$SYNC_FIRST_MODE" -eq 1 ]]; then
        # --sync-first: Always run roster-sync before team apply
        run_roster_sync_waterfall || {
            log_error "Infrastructure sync failed, aborting team swap"
            exit "$EXIT_SWAP_FAILURE"
        }
    elif [[ "$AUTO_SYNC_MODE" -eq 1 ]]; then
        # --auto-sync: Only sync if roster has updates
        if roster_has_updates; then
            log "Roster updates detected, syncing infrastructure..."
            run_roster_sync_waterfall || {
                log_error "Infrastructure sync failed, aborting team swap"
                exit "$EXIT_SWAP_FAILURE"
            }
        else
            log_debug "No roster updates, skipping infrastructure sync"
        fi
    fi

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

    # Validate team schemas (workflow.yaml, orchestrator.yaml) before swap
    validate_team_schemas "$team_name" || {
        log_error "Team schema validation failed, aborting swap"
        exit "$EXIT_VALIDATION_FAILURE"
    }

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
    local orphan_commands
    orphan_commands=$(detect_resource_orphans "commands" ".claude/commands" "$team_name" "f" "*.md")
    if [[ -n "$orphan_commands" ]]; then
        if [[ -z "$ORPHAN_MODE" ]]; then
            local orphan_count
            orphan_count=$(echo "$orphan_commands" | wc -l | tr -d ' ')
            log_warning "Found ${orphan_count} orphan command(s) from other teams:"
            while IFS=: read -r cmd_name origin_team; do
                echo "  - $cmd_name (from $origin_team)"
            done <<< "$orphan_commands"
            log "Use --remove-all to clean up orphan commands"
        else
            echo "$orphan_commands" | remove_resource_orphans "commands" ".claude/commands" "$ORPHAN_MODE" "f"
        fi
    fi

    # Check for user command collisions before syncing team commands
    check_user_command_collisions "$team_name"

    # Sync team-specific commands
    swap_commands "$team_name"
    mark_commit_step "$COMMIT_STEP_COMMANDS"

    # Detect and handle orphan skills (skills from other teams)
    local orphan_skills
    orphan_skills=$(detect_resource_orphans "skills" ".claude/skills" "$team_name" "d" "*/")
    if [[ -n "$orphan_skills" ]]; then
        if [[ -z "$ORPHAN_MODE" ]]; then
            # Non-interactive mode without flags - warn but don't block
            local orphan_count
            orphan_count=$(echo "$orphan_skills" | wc -l | tr -d ' ')
            log_warning "Found ${orphan_count} orphan skill(s) from other teams:"
            while IFS=: read -r skill_name origin_team; do
                echo "  - $skill_name (from $origin_team)"
            done <<< "$orphan_skills"
            log "Use --remove-all to clean up orphan skills"
        else
            echo "$orphan_skills" | remove_resource_orphans "skills" ".claude/skills" "$ORPHAN_MODE" "d"
        fi
    fi

    # Sync team-specific skills (Phase 2: Unified Sync)
    swap_skills "$team_name"
    mark_commit_step "$COMMIT_STEP_SKILLS"

    # Sync shared skills (always active, team-privileged override)
    sync_shared_skills
    mark_commit_step "$COMMIT_STEP_SHARED_SKILLS"

    # Detect and handle orphan hooks (hooks from other teams)
    local orphan_hooks
    orphan_hooks=$(detect_resource_orphans "hooks" ".claude/hooks" "$team_name" "f" "*")
    if [[ -n "$orphan_hooks" ]]; then
        if [[ -z "$ORPHAN_MODE" ]]; then
            local orphan_count
            orphan_count=$(echo "$orphan_hooks" | wc -l | tr -d ' ')
            log_warning "Found ${orphan_count} orphan hook(s) from other teams:"
            while IFS=: read -r hook_name origin_team; do
                echo "  - $hook_name (from $origin_team)"
            done <<< "$orphan_hooks"
            log "Use --remove-all to clean up orphan hooks"
        else
            echo "$orphan_hooks" | remove_resource_orphans "hooks" ".claude/hooks" "$ORPHAN_MODE" "f"
        fi
    fi

    # Sync team hooks
    swap_hooks "$team_name"
    mark_commit_step "$COMMIT_STEP_HOOKS"

    # Update hook registrations in settings.local.json (Scope 2)
    swap_hook_registrations "$team_name"
    mark_commit_step "$COMMIT_STEP_HOOK_REGISTRATIONS"

    # =========================================================================
    # PART 2 OF COMMIT: Manifest and ACTIVE_TEAM (the actual commit)
    # =========================================================================
    # Write manifest with current state (after commands synced so we capture them)
    # IMPORTANT: Manifest must be written BEFORE ACTIVE_TEAM
    write_manifest "$team_name"
    mark_commit_step "$COMMIT_STEP_MANIFEST"

    # =========================================================================
    # FINAL COMMIT: Write ACTIVE_TEAM (LAST - this is the commit point)
    # =========================================================================
    # ACTIVE_TEAM is the commit indicator - if it contains the new team name,
    # the swap is considered complete. Writing it LAST ensures all resources
    # are in place first. This is the POINT-OF-NO-RETURN.
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
    # Mark point-of-no-return - after this, recovery must complete forward
    mark_commit_step "$COMMIT_STEP_ACTIVE_TEAM"

    # =========================================================================
    # PHASE: COMPLETED - Transaction is committed
    # =========================================================================
    update_journal_phase "$PHASE_COMPLETED"

    # Delete journal first to mark transaction complete, then best-effort cleanup
    # This ordering prevents orphaned journal if process dies during cleanup
    # Staging and backup cleanup are best-effort (orphaned dirs don't block future swaps)
    delete_journal
    cleanup_staging
    cleanup_swap_backup

    # =========================================================================
    # POST-COMMIT OPERATIONS (Non-critical, swap is already complete)
    # =========================================================================
    # These operations are best-effort. If they fail, the swap is still valid.

    # Restore kept agents after swap
    restore_kept_agents
    cleanup_stash

    # Update CEM manifest team section (for roster-sync staleness tracking)
    update_cem_manifest_team "$team_name"

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

    # Auto-cleanup orphan backups if flag enabled (non-critical)
    if [[ "$AUTO_CLEANUP_MODE" -eq 1 ]]; then
        cleanup_orphan_backups || log_warning "Orphan backup cleanup failed (non-critical)"
    fi

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
# Reset to Baseline (No Team)
# ============================================================================

# Preview what reset would remove (for --dry-run with --reset)
preview_reset() {
    log "Dry-run: Would reset to baseline (no team)"
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
    echo "Would regenerate: CLAUDE.md (baseline)"

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

# Regenerate CLAUDE.md for baseline (no active team)
regenerate_baseline_claude_md() {
    local claude_md=".claude/CLAUDE.md"

    [[ -f "$claude_md" ]] || return 0

    log_debug "Regenerating CLAUDE.md for baseline (no team)"

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

    log_debug "CLAUDE.md regenerated for baseline"
}

# Perform reset to baseline (no team)
perform_reset() {
    log_debug "Starting reset to baseline"

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
        log "No team active. Already at baseline."
        return 0
    fi

    log "Resetting from $current_team to baseline..."

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
    regenerate_baseline_claude_md
    log "Regenerated: CLAUDE.md (baseline)"

    echo ""
    log "Reset complete. Baseline active (no team)."
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
            --reset)
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
            --sync-first)
                SYNC_FIRST_MODE=1
                shift
                ;;
            --auto-sync)
                AUTO_SYNC_MODE=1
                shift
                ;;
            --cleanup-orphans)
                CLEANUP_ORPHANS_MODE=1
                shift
                ;;
            --auto-cleanup)
                AUTO_CLEANUP_MODE=1
                shift
                ;;
            --interactive)
                INTERACTIVE_MODE="yes"
                shift
                ;;
            --no-interactive|--batch)
                INTERACTIVE_MODE="no"
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
        if is_interactive; then
            check_journal_recovery
        else
            log_error "Recovery mode requires interactive terminal (or use --interactive)"
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

    # Handle cleanup-orphans mode (takes precedence)
    if [[ "$CLEANUP_ORPHANS_MODE" -eq 1 ]]; then
        cleanup_orphan_backups
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
