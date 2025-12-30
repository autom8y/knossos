#!/usr/bin/env bash
#
# sync-user-agents.sh - Sync roster user-agents to ~/.claude/agents/
#
# Syncs agents from roster/user-agents/ to the user-level agents directory.
# Behavior:
#   - Additive: Never removes existing agents from ~/.claude/agents/
#   - Overwrites: Only agents previously installed from roster (tracked in manifest)
#   - Preserves: User-created agents not from roster
#
# Usage:
#   ./sync-user-agents.sh              # Sync user-agents to ~/.claude/agents/
#   ./sync-user-agents.sh --dry-run    # Preview changes without applying
#   ./sync-user-agents.sh --status     # Show sync status
#   ./sync-user-agents.sh --help       # Show usage
#
# Environment Variables:
#   ROSTER_HOME    Roster repository location (default: ~/Code/roster)
#   ROSTER_DEBUG   Enable debug logging (set to 1)

set -euo pipefail

# Constants
readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly ROSTER_DEBUG="${ROSTER_DEBUG:-0}"
readonly USER_AGENTS_DIR="$HOME/.claude/agents"
readonly USER_MANIFEST_FILE="$HOME/.claude/USER_AGENT_MANIFEST.json"
readonly SOURCE_DIR="$ROSTER_HOME/user-agents"
readonly MANIFEST_VERSION="1.0"

readonly EXIT_SUCCESS=0
readonly EXIT_INVALID_ARGS=1
readonly EXIT_SOURCE_MISSING=2
readonly EXIT_SYNC_FAILURE=3

# Mode flags
DRY_RUN_MODE=0
ADOPT_MODE=0

# Colors for output (if terminal supports it)
if [[ -t 1 ]]; then
    readonly GREEN='\033[0;32m'
    readonly YELLOW='\033[1;33m'
    readonly BLUE='\033[0;34m'
    readonly NC='\033[0m' # No Color
else
    readonly GREEN=''
    readonly YELLOW=''
    readonly BLUE=''
    readonly NC=''
fi

# Logging functions
# IMPORTANT: All log functions MUST output to stderr to avoid polluting
# captured stdout in functions that return data via echo
log() {
    echo "[User-Agents] $*" >&2
}

log_success() {
    echo -e "[User-Agents] ${GREEN}$*${NC}" >&2
}

log_info() {
    echo -e "[User-Agents] ${BLUE}$*${NC}" >&2
}

log_warning() {
    echo -e "[User-Agents] ${YELLOW}Warning:${NC} $*" >&2
}

log_error() {
    echo "[User-Agents] Error: $*" >&2
}

log_debug() {
    if [[ "$ROSTER_DEBUG" == "1" ]]; then
        echo "[User-Agents DEBUG] $*" >&2
    fi
}

# ============================================================================
# Checksum Functions
# ============================================================================

# Calculate checksum of a file (portable across macOS/Linux)
calculate_checksum() {
    local file="$1"
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file" | cut -d' ' -f1
    elif command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file" | cut -d' ' -f1
    else
        # Fallback to md5 if sha256 unavailable
        if command -v md5 >/dev/null 2>&1; then
            md5 -q "$file"
        else
            md5sum "$file" | cut -d' ' -f1
        fi
    fi
}

# ============================================================================
# Team Collision Detection
# ============================================================================

# Check if an agent name exists in any team pack
# Returns 0 if found in a team (collision), 1 if not found
is_team_agent() {
    local agent_name="$1"
    local roster_teams="${ROSTER_HOME}/teams"

    if [[ ! -d "$roster_teams" ]]; then
        return 1
    fi

    # Search for agent in any team's agents/ directory
    if find "$roster_teams" -path "*/agents/$agent_name" -type f 2>/dev/null | grep -q .; then
        return 0
    fi

    return 1
}

# Get which team(s) contain an agent
get_team_for_agent() {
    local agent_name="$1"
    local roster_teams="${ROSTER_HOME}/teams"

    if [[ ! -d "$roster_teams" ]]; then
        echo ""
        return
    fi

    # Find team directories containing this agent
    find "$roster_teams" -path "*/agents/$agent_name" -type f 2>/dev/null | while read -r path; do
        # Extract team name from path: .../teams/TEAM_NAME/agents/agent.md
        echo "$path" | sed 's|.*/teams/\([^/]*\)/agents/.*|\1|'
    done | tr '\n' ',' | sed 's/,$//'
}

# ============================================================================
# Manifest Functions
# ============================================================================

# Read manifest and return JSON or empty if not exists
read_manifest() {
    if [[ -f "$USER_MANIFEST_FILE" ]]; then
        cat "$USER_MANIFEST_FILE"
    else
        echo ""
    fi
}

# Check if agent is managed by roster (exists in manifest with source=roster)
is_roster_managed() {
    local agent_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        return 1
    fi

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        local source
        source=$(echo "$manifest" | jq -r --arg name "$agent_name" '.agents[$name].source // empty' 2>/dev/null)
        if [[ "$source" == "roster" || "$source" == "roster-diverged" ]]; then
            return 0
        fi
        return 1
    fi

    # Fallback to grep-based parsing
    local agent_block
    agent_block=$(echo "$manifest" | grep -A3 "\"$agent_name\":" 2>/dev/null | head -4)

    if [[ -z "$agent_block" ]]; then
        return 1
    fi

    local source
    source=$(echo "$agent_block" | grep '"source"' | sed 's/.*"source":[[:space:]]*"\([^"]*\)".*/\1/')

    if [[ "$source" == "roster" || "$source" == "roster-diverged" ]]; then
        return 0
    fi

    return 1
}

# Add or update a single manifest entry for an agent
# Usage: add_to_manifest agent_name source_type checksum
add_to_manifest() {
    local agent_name="$1"
    local source_type="$2"
    local checksum="$3"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Ensure manifest exists
    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        init_manifest
    fi

    # Read current manifest
    local manifest
    manifest=$(read_manifest)

    # Use jq to add/update entry
    if command -v jq >/dev/null 2>&1; then
        local updated
        updated=$(echo "$manifest" | jq --arg name "$agent_name" \
            --arg src "$source_type" \
            --arg ts "$timestamp" \
            --arg cs "$checksum" \
            '.agents[$name] = {"source": $src, "installed_at": $ts, "checksum": $cs}')
        echo "$updated" > "$USER_MANIFEST_FILE"
    else
        log_warning "jq not available, cannot update manifest entry for $agent_name"
    fi
}

# Recover manifest entries from existing agent files that match roster sources
recover_manifest() {
    log_info "Recovering manifest from existing agents..."

    local target_dir="$USER_AGENTS_DIR"
    local recovered=0
    local diverged=0

    # Ensure source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        return 1
    fi

    # Process each agent in target directory
    for target_file in "$target_dir"/*.md; do
        [[ -f "$target_file" ]] || continue
        local agent_name
        agent_name=$(basename "$target_file")

        # Skip if already in manifest as roster-managed
        if is_roster_managed "$agent_name"; then
            log_debug "Already managed: $agent_name"
            continue
        fi

        # Check if this agent exists in roster source
        local source_file="$SOURCE_DIR/$agent_name"
        if [[ -f "$source_file" ]]; then
            local source_checksum target_checksum
            source_checksum=$(calculate_checksum "$source_file")
            target_checksum=$(calculate_checksum "$target_file")

            if [[ "$source_checksum" == "$target_checksum" ]]; then
                # Exact match - adopt as roster-managed
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt: $agent_name (exact match)"
                else
                    add_to_manifest "$agent_name" "roster" "$target_checksum"
                    log_success "Adopted: $agent_name (exact match)"
                fi
                ((recovered++)) || true
            else
                # Diverged - mark as roster-diverged to preserve user changes
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt (diverged): $agent_name (local modifications preserved)"
                else
                    add_to_manifest "$agent_name" "roster-diverged" "$target_checksum"
                    log_warning "Adopted (diverged): $agent_name (local modifications preserved)"
                fi
                ((diverged++)) || true
            fi
        else
            log_debug "Not in roster: $agent_name (user-created)"
        fi
    done

    log_info "Recovery complete: $recovered adopted, $diverged diverged"
}

# Get checksum from manifest for an agent
get_manifest_checksum() {
    local agent_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        echo ""
        return
    fi

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        echo "$manifest" | jq -r --arg name "$agent_name" '.agents[$name].checksum // empty' 2>/dev/null
        return
    fi

    # Fallback to grep-based parsing
    local agent_block
    agent_block=$(echo "$manifest" | grep -A4 "\"$agent_name\":" 2>/dev/null | head -5)

    if [[ -z "$agent_block" ]]; then
        echo ""
        return
    fi

    echo "$agent_block" | grep '"checksum"' | sed 's/.*"checksum":[[:space:]]*"\([^"]*\)".*/\1/'
}

# Initialize empty manifest
init_manifest() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    mkdir -p "$(dirname "$USER_MANIFEST_FILE")"

    cat > "$USER_MANIFEST_FILE" <<EOF
{
  "manifest_version": "$MANIFEST_VERSION",
  "last_sync": "$timestamp",
  "agents": {}
}
EOF

    log_debug "Initialized empty manifest at $USER_MANIFEST_FILE"
}

# Write manifest with current roster-managed agents
# Usage: write_manifest agent1:checksum1 agent2:checksum2 ...
write_manifest() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    mkdir -p "$(dirname "$USER_MANIFEST_FILE")"

    # Start JSON
    {
        echo "{"
        echo "  \"manifest_version\": \"$MANIFEST_VERSION\","
        echo "  \"last_sync\": \"$timestamp\","
        echo "  \"agents\": {"
    } > "$USER_MANIFEST_FILE"

    # Add each agent entry
    local first=true
    for entry in "$@"; do
        # Skip empty entries
        [[ -z "$entry" ]] && continue

        local agent_name checksum
        agent_name=$(echo "$entry" | cut -d: -f1)
        checksum=$(echo "$entry" | cut -d: -f2)

        # Skip entries with empty agent name
        [[ -z "$agent_name" ]] && continue

        # Add comma separator
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$USER_MANIFEST_FILE"
        fi

        # Write agent entry
        {
            echo -n "    \"$agent_name\": {"
            echo -n "\"source\": \"roster\", "
            echo -n "\"installed_at\": \"$timestamp\", "
            echo -n "\"checksum\": \"$checksum\""
            echo -n "}"
        } >> "$USER_MANIFEST_FILE"
    done

    # Close JSON
    {
        echo ""
        echo "  }"
        echo "}"
    } >> "$USER_MANIFEST_FILE"

    log_debug "Manifest written: $USER_MANIFEST_FILE"
}

# ============================================================================
# Sync Functions
# ============================================================================

# Perform the sync operation
perform_sync() {
    log_debug "Starting sync from $SOURCE_DIR to $USER_AGENTS_DIR"

    # Check if source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        log "Create the directory and add agent files to sync"
        exit "$EXIT_SOURCE_MISSING"
    fi

    # Ensure target directory exists
    mkdir -p "$USER_AGENTS_DIR"

    # Initialize manifest if it doesn't exist
    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        log_debug "No manifest found, initializing"
        init_manifest
    fi

    # Track agents to write to manifest
    local manifest_entries=()
    local added=0
    local updated=0
    local skipped=0
    local unchanged=0

    # Process each agent in source directory
    for source_file in "$SOURCE_DIR"/*.md; do
        [[ -f "$source_file" ]] || continue

        local agent_name
        agent_name=$(basename "$source_file")
        local target_file="$USER_AGENTS_DIR/$agent_name"
        local source_checksum
        source_checksum=$(calculate_checksum "$source_file")

        log_debug "Processing: $agent_name (checksum: ${source_checksum:0:8}...)"

        # Check for team collision (user agent with same name as team agent)
        if is_team_agent "$agent_name"; then
            local teams
            teams=$(get_team_for_agent "$agent_name")
            log_warning "Collision: $agent_name exists in team pack(s): $teams"
            log_warning "  User-level agent will be shadowed when that team is active"
        fi

        if [[ -f "$target_file" ]]; then
            # Target exists - check if we can overwrite
            if is_roster_managed "$agent_name"; then
                # Roster-managed: check if update needed
                local manifest_checksum
                manifest_checksum=$(get_manifest_checksum "$agent_name")

                if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                    # No change needed
                    log_debug "Unchanged: $agent_name"
                    ((unchanged++)) || true
                    manifest_entries+=("$agent_name:$source_checksum")
                else
                    # Update needed
                    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                        log_info "Would update: $agent_name"
                    else
                        cp "$source_file" "$target_file"
                        log_success "Updated: $agent_name"
                    fi
                    ((updated++)) || true
                    manifest_entries+=("$agent_name:$source_checksum")
                fi
            else
                # User-created: skip with warning
                log_warning "Skipped: $agent_name (user-created, not overwriting)"
                ((skipped++)) || true
                # Do NOT add to manifest - preserve user ownership
            fi
        else
            # Target doesn't exist - add new agent
            if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                log_info "Would add: $agent_name"
            else
                cp "$source_file" "$target_file"
                log_success "Added: $agent_name"
            fi
            ((added++)) || true
            manifest_entries+=("$agent_name:$source_checksum")
        fi
    done

    # Preserve existing roster-managed agents that are no longer in source
    # (This handles the case where roster removes an agent - we still track it
    # but don't remove it, honoring the additive-only requirement)
    local manifest
    manifest=$(read_manifest)
    if [[ -n "$manifest" ]]; then
        # Extract existing agents from manifest
        local existing_agents
        existing_agents=$(echo "$manifest" | grep -o '"[^"]*\.md":' | tr -d '":' || true)

        for existing in $existing_agents; do
            # Check if this agent is still in source
            local still_in_source=false
            for entry in "${manifest_entries[@]:-}"; do
                if [[ "$entry" == "$existing:"* ]]; then
                    still_in_source=true
                    break
                fi
            done

            if [[ "$still_in_source" == false ]] && [[ -f "$USER_AGENTS_DIR/$existing" ]]; then
                # Agent removed from roster but still exists - keep in manifest
                # so we know it came from roster originally
                local checksum
                checksum=$(calculate_checksum "$USER_AGENTS_DIR/$existing")
                manifest_entries+=("$existing:$checksum")
                log_debug "Preserved manifest entry: $existing (no longer in roster)"
            fi
        done
    fi

    # Write updated manifest
    if [[ "$DRY_RUN_MODE" -eq 0 ]]; then
        write_manifest "${manifest_entries[@]:-}"
    fi

    # Summary
    echo ""
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        log "Dry-run complete:"
    else
        log "Sync complete:"
    fi

    local total=$((added + updated + unchanged + skipped))
    echo "  Added:     $added"
    echo "  Updated:   $updated"
    echo "  Unchanged: $unchanged"
    echo "  Skipped:   $skipped (user-created)"
    echo "  Total:     $total agent(s) processed"
}

# Show sync status
show_status() {
    echo "User-Agents Sync Status"
    echo "======================="
    echo ""
    echo "Source:  $SOURCE_DIR"
    echo "Target:  $USER_AGENTS_DIR"
    echo ""

    # Check source
    if [[ -d "$SOURCE_DIR" ]]; then
        local source_count
        source_count=$(find "$SOURCE_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        echo "Roster agents:  $source_count"
    else
        echo "Roster agents:  (directory not found)"
    fi

    # Check target
    if [[ -d "$USER_AGENTS_DIR" ]]; then
        local target_count
        target_count=$(find "$USER_AGENTS_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        echo "User agents:    $target_count"
    else
        echo "User agents:    (directory not found)"
    fi

    # Check manifest
    if [[ -f "$USER_MANIFEST_FILE" ]]; then
        local manifest_count last_sync
        manifest_count=$(grep -c '"source": "roster"' "$USER_MANIFEST_FILE" 2>/dev/null || echo "0")
        last_sync=$(grep '"last_sync"' "$USER_MANIFEST_FILE" | sed 's/.*"last_sync":[[:space:]]*"\([^"]*\)".*/\1/' || echo "unknown")
        echo "Roster-managed: $manifest_count"
        echo "Last sync:      $last_sync"
    else
        echo "Manifest:       (not initialized)"
    fi

    echo ""

    # Show detailed comparison
    if [[ -d "$SOURCE_DIR" ]] && [[ -d "$USER_AGENTS_DIR" ]]; then
        echo "Agent Status:"
        echo "-------------"

        # Check each source agent
        for source_file in "$SOURCE_DIR"/*.md; do
            [[ -f "$source_file" ]] || continue

            local agent_name
            agent_name=$(basename "$source_file")
            local target_file="$USER_AGENTS_DIR/$agent_name"

            if [[ -f "$target_file" ]]; then
                if is_roster_managed "$agent_name"; then
                    local source_checksum manifest_checksum
                    source_checksum=$(calculate_checksum "$source_file")
                    manifest_checksum=$(get_manifest_checksum "$agent_name")

                    if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                        echo "  [=] $agent_name (up to date)"
                    else
                        echo "  [~] $agent_name (update available)"
                    fi
                else
                    echo "  [!] $agent_name (user-created, would skip)"
                fi
            else
                echo "  [+] $agent_name (would add)"
            fi
        done

        # Check for user agents not in roster
        for target_file in "$USER_AGENTS_DIR"/*.md; do
            [[ -f "$target_file" ]] || continue

            local agent_name
            agent_name=$(basename "$target_file")

            if [[ ! -f "$SOURCE_DIR/$agent_name" ]]; then
                if is_roster_managed "$agent_name"; then
                    echo "  [-] $agent_name (was from roster, now removed from source)"
                else
                    echo "  [*] $agent_name (user-created)"
                fi
            fi
        done
    fi
}

# Usage information
usage() {
    cat <<EOF
Usage: sync-user-agents.sh [OPTIONS]

Syncs roster user-agents to ~/.claude/agents/

Options:
  --dry-run      Preview changes without applying
  --status       Show sync status without making changes
  --adopt        Recover manifest from existing agents (bootstrap/repair)
  --help, -h     Show this help message

Behavior:
  - Additive:   Never removes existing agents from ~/.claude/agents/
  - Overwrites: Only agents previously installed from roster
  - Preserves:  User-created agents not from roster

The manifest at ~/.claude/USER_AGENT_MANIFEST.json tracks which agents
were installed from roster, allowing safe updates while preserving
user-created agents.

Adopt Mode (--adopt):
  Scans existing agents in ~/.claude/agents/ and matches them against
  roster sources. Agents that match are adopted into the manifest:
  - Exact matches: marked as "roster" (fully managed)
  - Diverged files: marked as "roster-diverged" (preserves local changes)
  - User-created: not added to manifest (remain user-owned)

  Use --adopt when:
  - First-time setup with existing agents
  - Manifest was deleted or corrupted
  - Agents were installed before manifest tracking existed

Environment Variables:
  ROSTER_HOME    Roster repository location (default: ~/Code/roster)
  ROSTER_DEBUG   Enable debug logging (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Source directory missing
  3  Sync failure

Examples:
  ./sync-user-agents.sh              # Sync user-agents
  ./sync-user-agents.sh --dry-run    # Preview what would change
  ./sync-user-agents.sh --status     # Show current sync status
  ./sync-user-agents.sh --adopt      # Recover manifest from existing agents
  ./sync-user-agents.sh --adopt --dry-run  # Preview adopt results

EOF
}

# Main entry point
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dry-run)
                DRY_RUN_MODE=1
                shift
                ;;
            --adopt|--recover-manifest)
                ADOPT_MODE=1
                shift
                ;;
            --status)
                show_status
                exit "$EXIT_SUCCESS"
                ;;
            --help|-h)
                usage
                exit "$EXIT_SUCCESS"
                ;;
            -*)
                log_error "Unknown option: $1"
                usage
                exit "$EXIT_INVALID_ARGS"
                ;;
            *)
                log_error "Unexpected argument: $1"
                usage
                exit "$EXIT_INVALID_ARGS"
                ;;
        esac
    done

    # Run manifest recovery if adopt mode is enabled
    if [[ "$ADOPT_MODE" -eq 1 ]]; then
        # Ensure target directory exists before recovery
        mkdir -p "$USER_AGENTS_DIR"
        # Initialize manifest if needed
        if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
            init_manifest
        fi
        recover_manifest
    fi

    # Perform sync
    perform_sync
    exit "$EXIT_SUCCESS"
}

# Run main if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
