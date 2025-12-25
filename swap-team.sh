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

# Manifest file path
readonly MANIFEST_FILE=".claude/AGENT_MANIFEST.json"
readonly MANIFEST_VERSION="1.0"

# Orphan handling mode (set by flags)
ORPHAN_MODE=""  # "", "keep", "remove", "promote"

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

    # Extract source and origin for this agent using grep/sed (no jq dependency)
    local agent_block
    agent_block=$(echo "$manifest" | grep -A3 "\"$agent_name\":" | head -4)

    if [[ -z "$agent_block" ]]; then
        echo ""
        return
    fi

    local source origin
    source=$(echo "$agent_block" | grep '"source"' | sed 's/.*"source": *"\([^"]*\)".*/\1/')
    origin=$(echo "$agent_block" | grep '"origin"' | sed 's/.*"origin": *"\([^"]*\)".*/\1/')

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

    # Close JSON
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

    # Close JSON
    {
        echo ""
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
  (no args)      Show current active team

Orphan Handling Options (for non-interactive use):
  --keep-all     Preserve orphan agents in project
  --remove-all   Remove orphan agents (backup available)
  --promote-all  Move orphan agents to ~/.claude/agents/

When switching teams interactively, you'll be prompted for each orphan agent
(agents in current team but not in target team). In non-interactive mode
(scripts, CI), you must specify one of the orphan handling flags.

Environment Variables:
  ROSTER_HOME    Roster repository location (default: ~/Code/roster)
  ROSTER_DEBUG   Enable debug logging (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Validation failure (pack doesn't exist or is invalid)
  3  Backup failure
  4  Swap failure
  5  Orphan conflict (non-interactive without flag)

Examples:
  ./swap-team.sh dev-pack           # Switch to dev-pack (interactive prompts)
  ./swap-team.sh                    # Show current team
  ./swap-team.sh --list             # List available teams
  ./swap-team.sh dev-pack --keep-all    # Keep all orphans during swap
  ./swap-team.sh dev-pack --remove-all  # Remove all orphans during swap

EOF
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

# Preserve global agents (always available regardless of team)
preserve_global_agents() {
    local global_agents_dir=".claude/global-agents"

    if [[ -d "$global_agents_dir" ]]; then
        local global_count
        global_count=$(find "$global_agents_dir" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

        if [[ "$global_count" -gt 0 ]]; then
            log_debug "Copying $global_count global agent(s) to .claude/agents/"
            cp "$global_agents_dir"/*.md .claude/agents/ 2>/dev/null || {
                log_warning "Failed to copy global agents (team agents still swapped successfully)"
            }
        fi
    fi
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

# Main swap orchestration
perform_swap() {
    local team_name="$1"

    log_debug "Starting swap to $team_name"

    # Check if already active (idempotency)
    if [[ -f ".claude/ACTIVE_TEAM" ]]; then
        local current
        current=$(cat .claude/ACTIVE_TEAM | tr -d '[:space:]')

        if [[ "$current" == "$team_name" ]]; then
            log "Already using $team_name (no changes needed)"
            exit "$EXIT_SUCCESS"
        fi
    fi

    # Validate pack and project
    local agent_count
    agent_count=$(validate_pack "$team_name")
    validate_project

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

    # Perform swap with backup
    backup_current_agents
    swap_agents "$team_name" "$agent_count"

    # Restore kept agents after swap
    restore_kept_agents
    cleanup_stash

    preserve_global_agents
    update_active_team "$team_name"

    # Write manifest with current state
    write_manifest "$team_name"

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
    exit "$EXIT_SUCCESS"
}

# Main entry point
main() {
    local team_name=""

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

    # Handle the command
    if [[ -z "$team_name" ]]; then
        # No team specified - query current team
        query_current_team
    else
        # Team pack name - perform swap
        perform_swap "$team_name"
    fi
}

# Run main if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
