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
# Collision Handling (Intentional Design):
#   When a command exists in both user-commands/ (global) and rites/<name>/commands/
#   (rite-specific), collisions are logged as warnings but are expected behavior.
#   Rite-specific commands override global commands when that rite is active,
#   allowing rites to customize command behavior while preserving global defaults.
#
# Usage:
#   ./sync-user-commands.sh              # Sync user-commands to ~/.claude/commands/
#   ./sync-user-commands.sh --dry-run    # Preview changes without applying
#   ./sync-user-commands.sh --status     # Show sync status
#   ./sync-user-commands.sh --help       # Show usage
#
# Environment Variables:
#   KNOSSOS_HOME   Knossos platform location (default: ~/Code/roster)
#   ROSTER_HOME    Deprecated - use KNOSSOS_HOME instead
#   KNOSSOS_DEBUG   Enable debug logging (set to 1)

set -euo pipefail

# Source Knossos home resolution (handles ROSTER_HOME deprecation)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/knossos-home.sh"

# Constants
readonly KNOSSOS_DEBUG="${KNOSSOS_DEBUG:-0}"
readonly USER_COMMANDS_DIR="$HOME/.claude/commands"
readonly USER_MANIFEST_FILE="$HOME/.claude/USER_COMMAND_MANIFEST.json"
readonly SOURCE_DIR="$KNOSSOS_HOME/user-commands"
readonly MANIFEST_VERSION="1.0"

readonly EXIT_SUCCESS=0
readonly EXIT_INVALID_ARGS=1
readonly EXIT_SOURCE_MISSING=2
readonly EXIT_SYNC_FAILURE=3

# Mode flags
DRY_RUN_MODE=0
ADOPT_MODE=0
CLEANUP_MODE=0

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
    echo "[User-Commands] $*" >&2
}

log_success() {
    echo -e "[User-Commands] ${GREEN}$*${NC}" >&2
}

log_info() {
    echo -e "[User-Commands] ${BLUE}$*${NC}" >&2
}

log_warning() {
    echo -e "[User-Commands] ${YELLOW}Warning:${NC} $*" >&2
}

log_error() {
    echo "[User-Commands] Error: $*" >&2
}

log_debug() {
    if [[ "$KNOSSOS_DEBUG" == "1" ]]; then
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

# Check if a command name exists in any rite
# Returns 0 if found in a rite (collision), 1 if not found
is_rite_command() {
    local cmd_name="$1"
    local roster_rites="${KNOSSOS_HOME}/rites"

    if [[ ! -d "$roster_rites" ]]; then
        return 1
    fi

    # Search for command in any rite's commands/ directory
    if find "$roster_rites" -path "*/commands/$cmd_name" -type f 2>/dev/null | grep -q .; then
        return 0
    fi

    return 1
}

# Get which rite(s) contain a command
get_rite_for_command() {
    local cmd_name="$1"
    local roster_rites="${KNOSSOS_HOME}/rites"

    if [[ ! -d "$roster_rites" ]]; then
        echo ""
        return
    fi

    # Find rite directories containing this command
    find "$roster_rites" -path "*/commands/$cmd_name" -type f 2>/dev/null | while read -r path; do
        # Extract rite name from path: .../rites/RITE_NAME/commands/command.md
        echo "$path" | sed 's|.*/rites/\([^/]*\)/commands/.*|\1|'
    done | tr '\n' ',' | sed 's/,$//'
}

# ============================================================================
# Rite-Level Orphan Cleanup
# ============================================================================

# Clean up commands at user-level that should only exist at rite-level
# These are commands that:
#   1. Exist in ~/.claude/commands/
#   2. Do NOT exist in roster/user-commands/
#   3. DO exist in roster/rites/*/commands/ (rite-level resources)
cleanup_rite_orphans() {
    log_info "Scanning for rite-level commands that leaked to user-level..."

    local target_dir="$USER_COMMANDS_DIR"
    local backup_dir="$HOME/.claude/.backup/commands"
    local cleaned=0
    local skipped=0

    # Process each command in user-level directory
    for target_file in "$target_dir"/*.md; do
        [[ -f "$target_file" ]] || continue
        local cmd_name
        cmd_name=$(basename "$target_file")

        # Check if this command exists in roster/user-commands/ (legitimate user-level)
        local found_in_user_commands=false
        for category_dir in "$SOURCE_DIR"/*/; do
            [[ -d "$category_dir" ]] || continue
            if [[ -f "$category_dir/$cmd_name" ]]; then
                found_in_user_commands=true
                break
            fi
        done

        if [[ "$found_in_user_commands" == true ]]; then
            log_debug "Keeping: $cmd_name (in user-commands)"
            continue
        fi

        # Not in user-commands - check if it's a rite-level command
        if is_rite_command "$cmd_name"; then
            local rites
            rites=$(get_rite_for_command "$cmd_name")

            if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                log_info "Would remove: $cmd_name (rite-level from: $rites)"
            else
                # Backup before removing
                mkdir -p "$backup_dir"
                cp "$target_file" "$backup_dir/$cmd_name.$(date +%Y%m%d%H%M%S).bak"

                # Remove the file
                rm "$target_file"
                log_success "Removed: $cmd_name (rite-level from: $rites)"
                log_debug "  Backup: $backup_dir/$cmd_name.*.bak"

                # Remove from manifest if present
                if is_roster_managed "$cmd_name"; then
                    remove_from_manifest "$cmd_name"
                fi
            fi
            ((cleaned++)) || true
        else
            # Not in user-commands AND not a rite command = truly user-created
            log_debug "Keeping: $cmd_name (user-created, not from any roster source)"
            ((skipped++)) || true
        fi
    done

    echo ""
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        log "Cleanup preview:"
    else
        log "Cleanup complete:"
    fi
    echo "  Removed:   $cleaned (rite-level commands)"
    echo "  Preserved: $skipped (user-created commands)"

    if [[ "$cleaned" -gt 0 ]] && [[ "$DRY_RUN_MODE" -eq 0 ]]; then
        log_info "Backups saved to: $backup_dir"
    fi
}

# Remove an entry from the manifest
remove_from_manifest() {
    local cmd_name="$1"

    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        return
    fi

    if command -v jq >/dev/null 2>&1; then
        local updated
        updated=$(cat "$USER_MANIFEST_FILE" | jq --arg name "$cmd_name" 'del(.commands[$name])')
        echo "$updated" > "$USER_MANIFEST_FILE"
        log_debug "Removed from manifest: $cmd_name"
    else
        log_warning "jq not available, cannot remove $cmd_name from manifest"
    fi
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

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        local source
        source=$(echo "$manifest" | jq -r --arg name "$cmd_name" '.commands[$name].source // empty' 2>/dev/null)
        if [[ "$source" == "roster" || "$source" == "roster-diverged" ]]; then
            return 0
        fi
        return 1
    fi

    # Fallback to grep-based parsing
    local cmd_block
    cmd_block=$(echo "$manifest" | grep -A4 "\"$cmd_name\":" 2>/dev/null | head -5)

    if [[ -z "$cmd_block" ]]; then
        return 1
    fi

    local source
    source=$(echo "$cmd_block" | grep '"source"' | sed 's/.*"source":[[:space:]]*"\([^"]*\)".*/\1/')

    if [[ "$source" == "roster" || "$source" == "roster-diverged" ]]; then
        return 0
    fi

    return 1
}

# Add or update a single manifest entry
# Usage: add_to_manifest cmd_name source category checksum
add_to_manifest() {
    local cmd_name="$1"
    local source_type="$2"
    local category="$3"
    local checksum="$4"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Ensure manifest exists
    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        init_manifest
    fi

    # Read current manifest
    local manifest
    manifest=$(read_manifest)

    # Check if command already exists in manifest
    if echo "$manifest" | grep -q "\"$cmd_name\":"; then
        # Update existing entry using jq if available, else use sed
        if command -v jq >/dev/null 2>&1; then
            local updated
            updated=$(echo "$manifest" | jq --arg name "$cmd_name" \
                --arg src "$source_type" \
                --arg cat "$category" \
                --arg ts "$timestamp" \
                --arg cs "$checksum" \
                '.commands[$name] = {"source": $src, "category": $cat, "installed_at": $ts, "checksum": $cs}')
            echo "$updated" > "$USER_MANIFEST_FILE"
        else
            log_warning "jq not available, cannot update manifest entry for $cmd_name"
        fi
    else
        # Add new entry - need to insert before closing brace
        if command -v jq >/dev/null 2>&1; then
            local updated
            updated=$(echo "$manifest" | jq --arg name "$cmd_name" \
                --arg src "$source_type" \
                --arg cat "$category" \
                --arg ts "$timestamp" \
                --arg cs "$checksum" \
                '.commands[$name] = {"source": $src, "category": $cat, "installed_at": $ts, "checksum": $cs}')
            echo "$updated" > "$USER_MANIFEST_FILE"
        else
            log_warning "jq not available, cannot add manifest entry for $cmd_name"
        fi
    fi
}

# Recover manifest entries from existing files that match roster sources
recover_manifest() {
    log_info "Recovering manifest from existing commands..."

    local target_dir="$USER_COMMANDS_DIR"
    local recovered=0
    local diverged=0

    # Ensure source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        return 1
    fi

    # Process each command in target directory
    for target_file in "$target_dir"/*.md; do
        [[ -f "$target_file" ]] || continue
        local cmd_name
        cmd_name=$(basename "$target_file")

        # Skip if already in manifest as roster-managed
        if is_roster_managed "$cmd_name"; then
            log_debug "Already managed: $cmd_name"
            continue
        fi

        # Search for matching source file in roster (check all category subdirectories)
        local source_file=""
        local category=""
        for category_dir in "$SOURCE_DIR"/*/; do
            [[ -d "$category_dir" ]] || continue
            if [[ -f "$category_dir/$cmd_name" ]]; then
                source_file="$category_dir/$cmd_name"
                category=$(basename "$category_dir")
                break
            fi
        done

        if [[ -n "$source_file" && -f "$source_file" ]]; then
            local source_checksum target_checksum
            source_checksum=$(calculate_checksum "$source_file")
            target_checksum=$(calculate_checksum "$target_file")

            if [[ "$source_checksum" == "$target_checksum" ]]; then
                # Exact match - adopt as roster-managed
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt: $cmd_name (exact match)"
                else
                    add_to_manifest "$cmd_name" "roster" "$category" "$target_checksum"
                    log_success "Adopted: $cmd_name (exact match)"
                fi
                ((recovered++)) || true
            else
                # Diverged - mark as roster-diverged to preserve user changes
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt (diverged): $cmd_name (local modifications preserved)"
                else
                    add_to_manifest "$cmd_name" "roster-diverged" "$category" "$target_checksum"
                    log_warning "Adopted (diverged): $cmd_name (local modifications preserved)"
                fi
                ((diverged++)) || true
            fi
        else
            log_debug "Not in roster: $cmd_name (user-created)"
        fi
    done

    log_info "Recovery complete: $recovered adopted, $diverged diverged"
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

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        echo "$manifest" | jq -r --arg name "$cmd_name" '.commands[$name].checksum // empty' 2>/dev/null
        return
    fi

    # Fallback to grep-based parsing
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

            # Check for rite collision (user command with same name as rite command)
            if is_rite_command "$cmd_name"; then
                local rites
                rites=$(get_rite_for_command "$cmd_name")
                log_warning "Collision: $cmd_name exists in rite(s): $rites"
                log_warning "  Rite command will override when that rite is active"
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
        # Extract existing commands from manifest using jq if available
        local existing_commands
        if command -v jq >/dev/null 2>&1; then
            existing_commands=$(echo "$manifest" | jq -r '.commands | keys[]' 2>/dev/null || true)
        else
            existing_commands=$(echo "$manifest" | grep -o '"[^"]*\.md":' | tr -d '":' || true)
        fi

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
                # Get existing category from manifest using jq if available
                local existing_category
                if command -v jq >/dev/null 2>&1; then
                    existing_category=$(echo "$manifest" | jq -r --arg name "$existing" '.commands[$name].category // "unknown"' 2>/dev/null)
                else
                    existing_category=$(echo "$manifest" | grep -A3 "\"$existing\":" | grep '"category"' | sed 's/.*"category":[[:space:]]*"\([^"]*\)".*/\1/' || echo "unknown")
                fi
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
  --adopt        Recover manifest from existing commands (bootstrap/repair)
  --cleanup      Remove team-level commands that leaked to user-level
  --help, -h     Show this help message

Behavior:
  - Additive:   Never removes existing commands from ~/.claude/commands/
  - Overwrites: Only commands previously installed from roster
  - Preserves:  User-created commands not from roster
  - Flattens:   Subdirectories (session/, workflow/, etc.) become flat

The manifest at ~/.claude/USER_COMMAND_MANIFEST.json tracks which commands
were installed from roster, allowing safe updates while preserving
user-created commands.

Adopt Mode (--adopt):
  Scans existing commands in ~/.claude/commands/ and matches them against
  roster sources. Commands that match are adopted into the manifest:
  - Exact matches: marked as "roster" (fully managed)
  - Diverged files: marked as "roster-diverged" (preserves local changes)
  - User-created: not added to manifest (remain user-owned)

  Use --adopt when:
  - First-time setup with existing commands
  - Manifest was deleted or corrupted
  - Commands were installed before manifest tracking existed

Cleanup Mode (--cleanup):
  Scans commands in ~/.claude/commands/ and removes any that:
  - Do NOT exist in roster/user-commands/ (not user-level resources)
  - DO exist in roster/rites/*/commands/ (rite-level resources)

  Rite-level commands should only exist in .claude/commands/ (per-project)
  when that rite is active, not in ~/.claude/commands/ (user-level global).

  Use --cleanup when:
  - Rite commands leaked to user-level directory
  - Cleaning up after switching rites
  - Resetting to clean user-level state

  Backups are saved to ~/.claude/.backup/commands/ before removal.

Source Structure:
  roster/user-commands/
    session/       # start, park, continue, handoff, wrap
    workflow/      # task, sprint, hotfix
    operations/    # architect, build, qa, code-review, commit
    navigation/    # consult, team, worktree, sessions, ecosystem
    meta/          # minus-1, zero, one
    rite-switching/ # 10x, docs, hygiene, debt, sre, security, etc.

Environment Variables:
  ROSTER_HOME    Roster repository location (default: ~/Code/roster)
  KNOSSOS_DEBUG   Enable debug logging (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Source directory missing
  3  Sync failure

Examples:
  ./sync-user-commands.sh              # Sync user-commands
  ./sync-user-commands.sh --dry-run    # Preview what would change
  ./sync-user-commands.sh --status     # Show current sync status
  ./sync-user-commands.sh --adopt      # Recover manifest from existing files
  ./sync-user-commands.sh --adopt --dry-run  # Preview adopt results
  ./sync-user-commands.sh --cleanup    # Remove rite-level leaks
  ./sync-user-commands.sh --cleanup --dry-run  # Preview cleanup

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
            --cleanup)
                CLEANUP_MODE=1
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

    # Ensure target directory exists
    mkdir -p "$USER_COMMANDS_DIR"

    # Run cleanup if enabled (removes rite-level commands from user-level)
    if [[ "$CLEANUP_MODE" -eq 1 ]]; then
        cleanup_rite_orphans
    fi

    # Run manifest recovery if adopt mode is enabled
    if [[ "$ADOPT_MODE" -eq 1 ]]; then
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
