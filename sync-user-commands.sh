#!/usr/bin/env bash
#
# sync-user-commands.sh - Sync roster user-commands to ~/.claude/commands/
#
# Syncs commands from roster/user-commands/ to the user-level commands directory.
# Behavior:
#   - Additive: Never removes existing commands from ~/.claude/commands/
#   - Overwrites: Only commands previously installed from roster (tracked in manifest)
#   - Preserves: User-created commands not from roster
#   - Flattens: Subdirectories in source become flat list in target
#
# Usage:
#   ./sync-user-commands.sh              # Sync user-commands to ~/.claude/commands/
#   ./sync-user-commands.sh --dry-run    # Preview changes without applying
#   ./sync-user-commands.sh --status     # Show sync status
#   ./sync-user-commands.sh --help       # Show usage
#
# Environment Variables:
#   ROSTER_HOME    Roster repository location (default: ~/Code/roster)
#   ROSTER_DEBUG   Enable debug logging (set to 1)

set -euo pipefail

# Constants
readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly ROSTER_DEBUG="${ROSTER_DEBUG:-0}"
readonly USER_COMMANDS_DIR="$HOME/.claude/commands"
readonly USER_MANIFEST_FILE="$HOME/.claude/USER_COMMAND_MANIFEST.json"
readonly SOURCE_DIR="$ROSTER_HOME/user-commands"
readonly MANIFEST_VERSION="1.0"

readonly EXIT_SUCCESS=0
readonly EXIT_INVALID_ARGS=1
readonly EXIT_SOURCE_MISSING=2
readonly EXIT_SYNC_FAILURE=3

# Mode flags
DRY_RUN_MODE=0

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
log() {
    echo "[User-Commands] $*"
}

log_success() {
    echo -e "[User-Commands] ${GREEN}$*${NC}"
}

log_info() {
    echo -e "[User-Commands] ${BLUE}$*${NC}"
}

log_warning() {
    echo -e "[User-Commands] ${YELLOW}Warning:${NC} $*" >&2
}

log_error() {
    echo "[User-Commands] Error: $*" >&2
}

log_debug() {
    if [[ "$ROSTER_DEBUG" == "1" ]]; then
        echo "[User-Commands DEBUG] $*" >&2
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

# Check if a command name exists in any team pack
# Returns 0 if found in a team (collision), 1 if not found
is_team_command() {
    local cmd_name="$1"
    local roster_teams="${ROSTER_HOME}/teams"

    if [[ ! -d "$roster_teams" ]]; then
        return 1
    fi

    # Search for command in any team's commands/ directory
    if find "$roster_teams" -path "*/commands/$cmd_name" -type f 2>/dev/null | grep -q .; then
        return 0
    fi

    return 1
}

# Get which team(s) contain a command
get_team_for_command() {
    local cmd_name="$1"
    local roster_teams="${ROSTER_HOME}/teams"

    if [[ ! -d "$roster_teams" ]]; then
        echo ""
        return
    fi

    # Find team directories containing this command
    find "$roster_teams" -path "*/commands/$cmd_name" -type f 2>/dev/null | while read -r path; do
        # Extract team name from path: .../teams/TEAM_NAME/commands/command.md
        echo "$path" | sed 's|.*/teams/\([^/]*\)/commands/.*|\1|'
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

# Check if command is managed by roster (exists in manifest with source=roster)
is_roster_managed() {
    local cmd_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        return 1
    fi

    # Check if command exists in manifest with source=roster
    local cmd_block
    cmd_block=$(echo "$manifest" | grep -A4 "\"$cmd_name\":" 2>/dev/null | head -5)

    if [[ -z "$cmd_block" ]]; then
        return 1
    fi

    local source
    source=$(echo "$cmd_block" | grep '"source"' | sed 's/.*"source":[[:space:]]*"\([^"]*\)".*/\1/')

    if [[ "$source" == "roster" ]]; then
        return 0
    fi

    return 1
}

# Get checksum from manifest for a command
get_manifest_checksum() {
    local cmd_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        echo ""
        return
    fi

    local cmd_block
    cmd_block=$(echo "$manifest" | grep -A5 "\"$cmd_name\":" 2>/dev/null | head -6)

    if [[ -z "$cmd_block" ]]; then
        echo ""
        return
    fi

    echo "$cmd_block" | grep '"checksum"' | sed 's/.*"checksum":[[:space:]]*"\([^"]*\)".*/\1/'
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
  "commands": {}
}
EOF

    log_debug "Initialized empty manifest at $USER_MANIFEST_FILE"
}

# Write manifest with current roster-managed commands
# Usage: write_manifest "cmd1:checksum1:category1" "cmd2:checksum2:category2" ...
write_manifest() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    mkdir -p "$(dirname "$USER_MANIFEST_FILE")"

    # Start JSON
    {
        echo "{"
        echo "  \"manifest_version\": \"$MANIFEST_VERSION\","
        echo "  \"last_sync\": \"$timestamp\","
        echo "  \"commands\": {"
    } > "$USER_MANIFEST_FILE"

    # Add each command entry
    local first=true
    for entry in "$@"; do
        # Skip empty entries
        [[ -z "$entry" ]] && continue

        local cmd_name checksum category
        cmd_name=$(echo "$entry" | cut -d: -f1)
        checksum=$(echo "$entry" | cut -d: -f2)
        category=$(echo "$entry" | cut -d: -f3)

        # Skip entries with empty command name
        [[ -z "$cmd_name" ]] && continue

        # Add comma separator
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$USER_MANIFEST_FILE"
        fi

        # Write command entry
        {
            echo -n "    \"$cmd_name\": {"
            echo -n "\"source\": \"roster\", "
            echo -n "\"category\": \"$category\", "
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
    log_debug "Starting sync from $SOURCE_DIR to $USER_COMMANDS_DIR"

    # Check if source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        log "Create the directory and add command files to sync"
        exit "$EXIT_SOURCE_MISSING"
    fi

    # Ensure target directory exists
    mkdir -p "$USER_COMMANDS_DIR"

    # Initialize manifest if it doesn't exist
    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        log_debug "No manifest found, initializing"
        init_manifest
    fi

    # Track commands to write to manifest
    local manifest_entries=()
    local added=0
    local updated=0
    local skipped=0
    local unchanged=0

    # Process each command in source subdirectories (flatten structure)
    for category_dir in "$SOURCE_DIR"/*/; do
        [[ -d "$category_dir" ]] || continue

        local category
        category=$(basename "$category_dir")

        for source_file in "$category_dir"/*.md; do
            [[ -f "$source_file" ]] || continue

            local cmd_name
            cmd_name=$(basename "$source_file")
            local target_file="$USER_COMMANDS_DIR/$cmd_name"
            local source_checksum
            source_checksum=$(calculate_checksum "$source_file")

            log_debug "Processing: $cmd_name (category: $category, checksum: ${source_checksum:0:8}...)"

            # Check for team collision (user command with same name as team command)
            if is_team_command "$cmd_name"; then
                local teams
                teams=$(get_team_for_command "$cmd_name")
                log_warning "Collision: $cmd_name exists in team pack(s): $teams"
                log_warning "  Team command will override when that team is active"
            fi

            if [[ -f "$target_file" ]]; then
                # Target exists - check if we can overwrite
                if is_roster_managed "$cmd_name"; then
                    # Roster-managed: check if update needed
                    local manifest_checksum
                    manifest_checksum=$(get_manifest_checksum "$cmd_name")

                    if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                        # No change needed
                        log_debug "Unchanged: $cmd_name"
                        ((unchanged++)) || true
                        manifest_entries+=("$cmd_name:$source_checksum:$category")
                    else
                        # Update needed
                        if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                            log_info "Would update: $cmd_name"
                        else
                            cp "$source_file" "$target_file"
                            log_success "Updated: $cmd_name"
                        fi
                        ((updated++)) || true
                        manifest_entries+=("$cmd_name:$source_checksum:$category")
                    fi
                else
                    # User-created: skip with warning
                    log_warning "Skipped: $cmd_name (user-created, not overwriting)"
                    ((skipped++)) || true
                    # Do NOT add to manifest - preserve user ownership
                fi
            else
                # Target doesn't exist - add new command
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would add: $cmd_name"
                else
                    cp "$source_file" "$target_file"
                    log_success "Added: $cmd_name"
                fi
                ((added++)) || true
                manifest_entries+=("$cmd_name:$source_checksum:$category")
            fi
        done
    done

    # Preserve existing roster-managed commands that are no longer in source
    # (This handles the case where roster removes a command - we still track it
    # but don't remove it, honoring the additive-only requirement)
    local manifest
    manifest=$(read_manifest)
    if [[ -n "$manifest" ]]; then
        # Extract existing commands from manifest
        local existing_commands
        existing_commands=$(echo "$manifest" | grep -o '"[^"]*\.md":' | tr -d '":' || true)

        for existing in $existing_commands; do
            # Check if this command is still in source
            local still_in_source=false
            for entry in "${manifest_entries[@]:-}"; do
                if [[ "$entry" == "$existing:"* ]]; then
                    still_in_source=true
                    break
                fi
            done

            if [[ "$still_in_source" == false ]] && [[ -f "$USER_COMMANDS_DIR/$existing" ]]; then
                # Command removed from roster but still exists - keep in manifest
                # so we know it came from roster originally
                local checksum
                checksum=$(calculate_checksum "$USER_COMMANDS_DIR/$existing")
                # Get existing category from manifest
                local existing_category
                existing_category=$(echo "$manifest" | grep -A3 "\"$existing\":" | grep '"category"' | sed 's/.*"category":[[:space:]]*"\([^"]*\)".*/\1/' || echo "unknown")
                manifest_entries+=("$existing:$checksum:$existing_category")
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
    echo "  Total:     $total command(s) processed"
}

# Show sync status
show_status() {
    echo "User-Commands Sync Status"
    echo "========================="
    echo ""
    echo "Source:  $SOURCE_DIR"
    echo "Target:  $USER_COMMANDS_DIR"
    echo ""

    # Check source - count across all subdirectories
    if [[ -d "$SOURCE_DIR" ]]; then
        local source_count=0
        for category_dir in "$SOURCE_DIR"/*/; do
            [[ -d "$category_dir" ]] || continue
            local cat_count
            cat_count=$(find "$category_dir" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
            source_count=$((source_count + cat_count))
        done
        echo "Roster commands: $source_count"

        # Show by category
        for category_dir in "$SOURCE_DIR"/*/; do
            [[ -d "$category_dir" ]] || continue
            local category
            category=$(basename "$category_dir")
            local cat_count
            cat_count=$(find "$category_dir" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
            echo "  $category: $cat_count"
        done
    else
        echo "Roster commands: (directory not found)"
    fi

    echo ""

    # Check target
    if [[ -d "$USER_COMMANDS_DIR" ]]; then
        local target_count
        target_count=$(find "$USER_COMMANDS_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        echo "User commands:   $target_count"
    else
        echo "User commands:   (directory not found)"
    fi

    # Check manifest
    if [[ -f "$USER_MANIFEST_FILE" ]]; then
        local manifest_count last_sync
        manifest_count=$(grep -c '"source": "roster"' "$USER_MANIFEST_FILE" 2>/dev/null || echo "0")
        last_sync=$(grep '"last_sync"' "$USER_MANIFEST_FILE" | sed 's/.*"last_sync":[[:space:]]*"\([^"]*\)".*/\1/' || echo "unknown")
        echo "Roster-managed:  $manifest_count"
        echo "Last sync:       $last_sync"
    else
        echo "Manifest:        (not initialized)"
    fi

    echo ""

    # Show detailed comparison
    if [[ -d "$SOURCE_DIR" ]] && [[ -d "$USER_COMMANDS_DIR" ]]; then
        echo "Command Status:"
        echo "---------------"

        # Check each source command
        for category_dir in "$SOURCE_DIR"/*/; do
            [[ -d "$category_dir" ]] || continue

            for source_file in "$category_dir"/*.md; do
                [[ -f "$source_file" ]] || continue

                local cmd_name
                cmd_name=$(basename "$source_file")
                local target_file="$USER_COMMANDS_DIR/$cmd_name"

                if [[ -f "$target_file" ]]; then
                    if is_roster_managed "$cmd_name"; then
                        local source_checksum manifest_checksum
                        source_checksum=$(calculate_checksum "$source_file")
                        manifest_checksum=$(get_manifest_checksum "$cmd_name")

                        if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                            echo "  [=] $cmd_name (up to date)"
                        else
                            echo "  [~] $cmd_name (update available)"
                        fi
                    else
                        echo "  [!] $cmd_name (user-created, would skip)"
                    fi
                else
                    echo "  [+] $cmd_name (would add)"
                fi
            done
        done

        # Check for user commands not in roster
        for target_file in "$USER_COMMANDS_DIR"/*.md; do
            [[ -f "$target_file" ]] || continue

            local cmd_name
            cmd_name=$(basename "$target_file")

            # Check if this command exists in any source category
            local found_in_source=false
            for category_dir in "$SOURCE_DIR"/*/; do
                if [[ -f "$category_dir/$cmd_name" ]]; then
                    found_in_source=true
                    break
                fi
            done

            if [[ "$found_in_source" == false ]]; then
                if is_roster_managed "$cmd_name"; then
                    echo "  [-] $cmd_name (was from roster, now removed from source)"
                else
                    echo "  [*] $cmd_name (user-created)"
                fi
            fi
        done
    fi
}

# Usage information
usage() {
    cat <<EOF
Usage: sync-user-commands.sh [OPTIONS]

Syncs roster user-commands to ~/.claude/commands/

Options:
  --dry-run      Preview changes without applying
  --status       Show sync status without making changes
  --help, -h     Show this help message

Behavior:
  - Additive:   Never removes existing commands from ~/.claude/commands/
  - Overwrites: Only commands previously installed from roster
  - Preserves:  User-created commands not from roster
  - Flattens:   Subdirectories (session/, workflow/, etc.) become flat

The manifest at ~/.claude/USER_COMMAND_MANIFEST.json tracks which commands
were installed from roster, allowing safe updates while preserving
user-created commands.

Source Structure:
  roster/user-commands/
    session/       # start, park, continue, handoff, wrap
    workflow/      # task, sprint, hotfix
    operations/    # architect, build, qa, code-review, commit
    navigation/    # consult, team, worktree, sessions, ecosystem
    meta/          # minus-1, zero, one
    team-switching/ # 10x, docs, hygiene, debt, sre, security, etc.

Environment Variables:
  ROSTER_HOME    Roster repository location (default: ~/Code/roster)
  ROSTER_DEBUG   Enable debug logging (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Source directory missing
  3  Sync failure

Examples:
  ./sync-user-commands.sh              # Sync user-commands
  ./sync-user-commands.sh --dry-run    # Preview what would change
  ./sync-user-commands.sh --status     # Show current sync status

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

    # Perform sync
    perform_sync
    exit "$EXIT_SUCCESS"
}

# Run main if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
