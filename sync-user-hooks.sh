#!/usr/bin/env bash
#
# sync-user-hooks.sh - Sync knossos hooks to ~/.claude/hooks/
#
# Syncs hooks from knossos/hooks/ (canonical source) to user-level ~/.claude/hooks/.
# Behavior:
#   - Additive: Never removes existing hooks from ~/.claude/hooks/
#   - Overwrites: Only hooks previously installed from knossos (tracked in manifest)
#   - Preserves: User-created hooks not from knossos
#   - Nested: Handles lib/ subdirectory preserving structure
#
# Usage:
#   ./sync-user-hooks.sh              # Sync hooks to ~/.claude/hooks/
#   ./sync-user-hooks.sh --dry-run    # Preview changes without applying
#   ./sync-user-hooks.sh --status     # Show sync status
#   ./sync-user-hooks.sh --help       # Show usage
#
# Environment Variables:
#   KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)
#   KNOSSOS_DEBUG   Enable debug logging (set to 1)

set -euo pipefail

# Source Knossos home resolution (resolves KNOSSOS_HOME)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/knossos-home.sh"

# Constants
readonly KNOSSOS_DEBUG="${KNOSSOS_DEBUG:-0}"
readonly USER_HOOKS_DIR="$HOME/.claude/hooks"
readonly USER_MANIFEST_FILE="$HOME/.claude/USER_HOOKS_MANIFEST.json"
readonly SOURCE_DIR="$KNOSSOS_HOME/user-hooks"
readonly MANIFEST_VERSION="1.1"

# Root exceptions (items that stay at root level, not in categories)
# lib: hook library files, ari: Ariadne CLI hooks (separate sync)
readonly ROOT_EXCEPTIONS="lib ari"

# Valid categories for hooks
readonly HOOK_CATEGORIES="context-injection session-guards validation tracking"

readonly EXIT_SUCCESS=0
readonly EXIT_INVALID_ARGS=1
readonly EXIT_SOURCE_MISSING=2
readonly EXIT_SYNC_FAILURE=3

# Mode flags
DRY_RUN_MODE=0
ADOPT_MODE=0
FORCE_MODE=0

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
# captured stdout in functions like sync_file() that return data via echo
log() {
    echo "[User-Hooks] $*" >&2
}

log_success() {
    echo -e "[User-Hooks] ${GREEN}$*${NC}" >&2
}

log_info() {
    echo -e "[User-Hooks] ${BLUE}$*${NC}" >&2
}

log_warning() {
    echo -e "[User-Hooks] ${YELLOW}Warning:${NC} $*" >&2
}

log_error() {
    echo "[User-Hooks] Error: $*" >&2
}

log_debug() {
    if [[ "$KNOSSOS_DEBUG" == "1" ]]; then
        echo "[User-Hooks DEBUG] $*" >&2
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

# Check if a hook name exists in any rite
# Returns 0 if found in a rite (collision), 1 if not found
is_rite_hook() {
    local hook_name="$1"
    local knossos_rites="${KNOSSOS_HOME}/rites"

    if [[ ! -d "$knossos_rites" ]]; then
        return 1
    fi

    # Search for hook in any rite's hooks/ directory
    if find "$knossos_rites" -path "*/hooks/$hook_name" -type f 2>/dev/null | grep -q .; then
        return 0
    fi

    return 1
}

# Get which rite(s) contain a hook
get_rite_for_hook() {
    local hook_name="$1"
    local knossos_rites="${KNOSSOS_HOME}/rites"

    if [[ ! -d "$knossos_rites" ]]; then
        echo ""
        return
    fi

    # Find rite directories containing this hook
    find "$knossos_rites" -path "*/hooks/$hook_name" -type f 2>/dev/null | while read -r path; do
        # Extract rite name from path: .../rites/RITE_NAME/hooks/hook.sh
        echo "$path" | sed 's|.*/rites/\([^/]*\)/hooks/.*|\1|'
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

# Check if hook is managed by knossos (exists in manifest with source=knossos)
is_knossos_managed() {
    local hook_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        return 1
    fi

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        local source
        source=$(echo "$manifest" | jq -r --arg name "$hook_name" '.hooks[$name].source // empty' 2>/dev/null)
        if [[ "$source" == "knossos" || "$source" == "knossos-diverged" ]]; then
            return 0
        fi
        return 1
    fi

    # Fallback to grep-based parsing
    local hook_block
    hook_block=$(echo "$manifest" | grep -A4 "\"$hook_name\":" 2>/dev/null | head -5)

    if [[ -z "$hook_block" ]]; then
        return 1
    fi

    local source
    source=$(echo "$hook_block" | grep '"source"' | sed 's/.*"source":[[:space:]]*"\([^"]*\)".*/\1/')

    if [[ "$source" == "knossos" || "$source" == "knossos-diverged" ]]; then
        return 0
    fi

    return 1
}

# Add or update a single manifest entry
# Usage: add_to_manifest hook_name source location checksum category
add_to_manifest() {
    local hook_name="$1"
    local source_type="$2"
    local location="$3"
    local checksum="$4"
    local category="${5:-root}"
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
        updated=$(echo "$manifest" | jq --arg name "$hook_name" \
            --arg src "$source_type" \
            --arg loc "$location" \
            --arg ts "$timestamp" \
            --arg cs "$checksum" \
            --arg cat "$category" \
            '.hooks[$name] = {"source": $src, "location": $loc, "installed_at": $ts, "checksum": $cs, "category": $cat}')
        echo "$updated" > "$USER_MANIFEST_FILE"
    else
        log_warning "jq not available, cannot update manifest entry for $hook_name"
    fi
}

# Remove an entry from the manifest
remove_from_manifest() {
    local hook_name="$1"

    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        return
    fi

    if command -v jq >/dev/null 2>&1; then
        local updated
        updated=$(cat "$USER_MANIFEST_FILE" | jq --arg name "$hook_name" 'del(.hooks[$name])')
        echo "$updated" > "$USER_MANIFEST_FILE"
        log_debug "Removed from manifest: $hook_name"
    else
        log_warning "jq not available, cannot remove $hook_name from manifest"
    fi
}

# Get checksum from manifest for a hook
get_manifest_checksum() {
    local hook_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        echo ""
        return
    fi

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        echo "$manifest" | jq -r --arg name "$hook_name" '.hooks[$name].checksum // empty' 2>/dev/null
        return
    fi

    # Fallback to grep-based parsing
    local hook_block
    hook_block=$(echo "$manifest" | grep -A5 "\"$hook_name\":" 2>/dev/null | head -6)

    if [[ -z "$hook_block" ]]; then
        echo ""
        return
    fi

    echo "$hook_block" | grep '"checksum"' | sed 's/.*"checksum":[[:space:]]*"\([^"]*\)".*/\1/'
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
  "hooks": {}
}
EOF

    log_debug "Initialized empty manifest at $USER_MANIFEST_FILE"
}

# Find source hook path and category for a hook name
# Returns: "source_path:category" or empty if not found
find_source_hook() {
    local hook_name="$1"

    # Check lib/ first (root exception) - hook_name includes lib/ prefix
    if [[ "$hook_name" == lib/* ]]; then
        local file_name="${hook_name#lib/}"
        local source_file="$SOURCE_DIR/lib/$file_name"
        if [[ -f "$source_file" ]]; then
            echo "$source_file:root"
            return 0
        fi
    else
        # Check category directories
        for category in $HOOK_CATEGORIES; do
            local source_file="$SOURCE_DIR/$category/$hook_name"
            if [[ -f "$source_file" ]]; then
                echo "$source_file:$category"
                return 0
            fi
        done
    fi

    return 1
}

# Recover manifest entries from existing files that match knossos sources
recover_manifest() {
    log_info "Recovering manifest from existing hooks..."

    local target_dir="$USER_HOOKS_DIR"
    local recovered=0
    local diverged=0

    # Ensure source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        return 1
    fi

    # Process root-level hooks (in flat destination)
    for target_file in "$target_dir"/*.sh; do
        [[ -f "$target_file" ]] || continue
        local hook_name
        hook_name=$(basename "$target_file")

        # Skip if already in manifest as knossos-managed
        if is_knossos_managed "$hook_name"; then
            log_debug "Already managed: $hook_name"
            continue
        fi

        # Check if source exists in knossos (now categorical)
        local source_info
        if source_info=$(find_source_hook "$hook_name"); then
            local source_file category
            source_file=$(echo "$source_info" | cut -d: -f1)
            category=$(echo "$source_info" | cut -d: -f2)

            local source_checksum target_checksum
            source_checksum=$(calculate_checksum "$source_file")
            target_checksum=$(calculate_checksum "$target_file")

            if [[ "$source_checksum" == "$target_checksum" ]]; then
                # Exact match - adopt as knossos-managed
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt: $hook_name (exact match, category: $category)"
                else
                    add_to_manifest "$hook_name" "knossos" "root" "$target_checksum" "$category"
                    log_success "Adopted: $hook_name (exact match, category: $category)"
                fi
                ((recovered++)) || true
            else
                # Diverged - mark as knossos-diverged to preserve user changes
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt (diverged): $hook_name (local modifications preserved)"
                else
                    add_to_manifest "$hook_name" "knossos-diverged" "root" "$target_checksum" "$category"
                    log_warning "Adopted (diverged): $hook_name (local modifications preserved)"
                fi
                ((diverged++)) || true
            fi
        else
            log_debug "Not in knossos: $hook_name (user-created)"
        fi
    done

    # Process lib/ subdirectory hooks
    if [[ -d "$target_dir/lib" ]]; then
        for target_file in "$target_dir/lib"/*.sh; do
            [[ -f "$target_file" ]] || continue
            local hook_name
            hook_name="lib/$(basename "$target_file")"

            # Skip if already in manifest as knossos-managed
            if is_knossos_managed "$hook_name"; then
                log_debug "Already managed: $hook_name"
                continue
            fi

            # Check if source exists in knossos
            local source_info
            if source_info=$(find_source_hook "$hook_name"); then
                local source_file category
                source_file=$(echo "$source_info" | cut -d: -f1)
                category=$(echo "$source_info" | cut -d: -f2)

                local source_checksum target_checksum
                source_checksum=$(calculate_checksum "$source_file")
                target_checksum=$(calculate_checksum "$target_file")

                if [[ "$source_checksum" == "$target_checksum" ]]; then
                    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                        log_info "Would adopt: $hook_name (exact match)"
                    else
                        add_to_manifest "$hook_name" "knossos" "lib" "$target_checksum" "$category"
                        log_success "Adopted: $hook_name (exact match)"
                    fi
                    ((recovered++)) || true
                else
                    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                        log_info "Would adopt (diverged): $hook_name (local modifications preserved)"
                    else
                        add_to_manifest "$hook_name" "knossos-diverged" "lib" "$target_checksum" "$category"
                        log_warning "Adopted (diverged): $hook_name (local modifications preserved)"
                    fi
                    ((diverged++)) || true
                fi
            else
                log_debug "Not in knossos: $hook_name (user-created)"
            fi
        done
    fi

    log_info "Recovery complete: $recovered adopted, $diverged diverged"
}

# Write manifest with current knossos-managed hooks
# Usage: write_manifest "hook1:checksum1:location1:category1" "hook2:checksum2:location2:category2" ...
write_manifest() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    mkdir -p "$(dirname "$USER_MANIFEST_FILE")"

    # Start JSON
    {
        echo "{"
        echo "  \"manifest_version\": \"$MANIFEST_VERSION\","
        echo "  \"last_sync\": \"$timestamp\","
        echo "  \"hooks\": {"
    } > "$USER_MANIFEST_FILE"

    # Add each hook entry
    local first=true
    for entry in "$@"; do
        # Skip empty entries
        [[ -z "$entry" ]] && continue

        local hook_name checksum location category
        hook_name=$(echo "$entry" | cut -d: -f1)
        checksum=$(echo "$entry" | cut -d: -f2)
        location=$(echo "$entry" | cut -d: -f3)
        category=$(echo "$entry" | cut -d: -f4)
        category="${category:-root}"

        # Skip entries with empty hook name
        [[ -z "$hook_name" ]] && continue

        # Add comma separator
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$USER_MANIFEST_FILE"
        fi

        # Write hook entry
        {
            echo -n "    \"$hook_name\": {"
            echo -n "\"source\": \"knossos\", "
            echo -n "\"location\": \"$location\", "
            echo -n "\"installed_at\": \"$timestamp\", "
            echo -n "\"checksum\": \"$checksum\", "
            echo -n "\"category\": \"$category\""
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

# Check if a directory name is a root exception
is_root_exception() {
    local name="$1"
    for exception in $ROOT_EXCEPTIONS; do
        [[ "$name" == "$exception" ]] && return 0
    done
    return 1
}

# Check if a directory name is a valid category
is_valid_category() {
    local name="$1"
    local cat
    for cat in $HOOK_CATEGORIES; do
        [[ "$name" == "$cat" ]] && return 0
    done
    return 1
}

# Sync a single file from source to target
# Returns: 0=added, 1=updated, 2=unchanged, 3=skipped
# Outputs: manifest entry string (hook_name:checksum:location:category)
sync_file() {
    local source_file="$1"
    local target_file="$2"
    local hook_name="$3"
    local location="$4"
    local category="$5"

    local source_checksum
    source_checksum=$(calculate_checksum "$source_file")

    log_debug "Processing: $hook_name (location: $location, category: $category, checksum: ${source_checksum:0:8}...)"

    # Check for rite collision
    if is_rite_hook "$(basename "$hook_name")"; then
        local rites
        rites=$(get_rite_for_hook "$(basename "$hook_name")")
        log_warning "Collision: $hook_name exists in rite(s): $rites"
        log_warning "  Rite hook will override when that rite is active"
    fi

    if [[ -f "$target_file" ]]; then
        # Target exists - check if we can overwrite
        if is_knossos_managed "$hook_name" || [[ "$FORCE_MODE" -eq 1 ]]; then
            # Knossos-managed or force mode: check if update needed
            local manifest_checksum
            manifest_checksum=$(get_manifest_checksum "$hook_name")

            if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                # No change needed
                log_debug "Unchanged: $hook_name"
                echo "$hook_name:$source_checksum:$location:$category"
                return 2  # unchanged
            else
                # Update needed
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would update: $hook_name"
                else
                    cp "$source_file" "$target_file"
                    chmod +x "$target_file"
                    log_success "Updated: $hook_name"
                fi
                echo "$hook_name:$source_checksum:$location:$category"
                return 1  # updated
            fi
        else
            # User-created: skip with warning
            log_warning "Skipped: $hook_name (user-created, not overwriting)"
            return 3  # skipped
        fi
    else
        # Target doesn't exist - add new hook
        if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
            log_info "Would add: $hook_name"
        else
            # Ensure parent directory exists
            mkdir -p "$(dirname "$target_file")"
            cp "$source_file" "$target_file"
            chmod +x "$target_file"
            log_success "Added: $hook_name"
        fi
        echo "$hook_name:$source_checksum:$location:$category"
        return 0  # added
    fi
}

# Perform the sync operation
perform_sync() {
    log_debug "Starting sync from $SOURCE_DIR to $USER_HOOKS_DIR"

    # Check if source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        log "Create the directory and add hook files to sync"
        exit "$EXIT_SOURCE_MISSING"
    fi

    # Ensure target directory exists
    mkdir -p "$USER_HOOKS_DIR"

    # Initialize manifest if it doesn't exist
    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        log_debug "No manifest found, initializing"
        init_manifest
    fi

    # Track hooks to write to manifest
    local manifest_entries=()
    local added=0
    local updated=0
    local skipped=0
    local unchanged=0

    # Phase 1: Process lib/ subdirectory (root exception, preserving structure)
    if [[ -d "$SOURCE_DIR/lib" ]]; then
        for source_file in "$SOURCE_DIR/lib"/*.sh; do
            [[ -f "$source_file" ]] || continue

            local file_name
            file_name=$(basename "$source_file")
            local hook_name="lib/$file_name"
            local target_file="$USER_HOOKS_DIR/lib/$file_name"

            local entry_result sync_status
            entry_result=$(sync_file "$source_file" "$target_file" "$hook_name" "lib" "root") || sync_status=$?
            sync_status=${sync_status:-0}

            if [[ -n "$entry_result" ]]; then
                manifest_entries+=("$entry_result")
            fi

            case $sync_status in
                0) ((added++)) || true ;;
                1) ((updated++)) || true ;;
                2) ((unchanged++)) || true ;;
                3) ((skipped++)) || true ;;
            esac
        done
    fi

    # Phase 2: Process categorized hooks (hooks inside category directories)
    for category_dir in "$SOURCE_DIR"/*/; do
        [[ -d "$category_dir" ]] || continue
        local category
        category=$(basename "${category_dir%/}")

        # Skip root exceptions (already processed)
        if is_root_exception "$category"; then
            continue
        fi

        # Skip if not a valid category
        if ! is_valid_category "$category"; then
            log_warning "Skipping unknown directory: $category (not a valid category)"
            continue
        fi

        # Process each hook in this category
        for source_file in "$category_dir"/*.sh; do
            [[ -f "$source_file" ]] || continue

            local hook_name
            hook_name=$(basename "$source_file")
            local target_file="$USER_HOOKS_DIR/$hook_name"

            local entry_result sync_status
            entry_result=$(sync_file "$source_file" "$target_file" "$hook_name" "root" "$category") || sync_status=$?
            sync_status=${sync_status:-0}

            if [[ -n "$entry_result" ]]; then
                manifest_entries+=("$entry_result")
            fi

            case $sync_status in
                0) ((added++)) || true ;;
                1) ((updated++)) || true ;;
                2) ((unchanged++)) || true ;;
                3) ((skipped++)) || true ;;
            esac
        done
    done

    # Preserve existing knossos-managed hooks that are no longer in source
    local manifest
    manifest=$(read_manifest)
    if [[ -n "$manifest" ]]; then
        local existing_hooks
        if command -v jq >/dev/null 2>&1; then
            existing_hooks=$(echo "$manifest" | jq -r '.hooks | keys[]' 2>/dev/null || true)
        else
            existing_hooks=$(echo "$manifest" | grep -o '"[^"]*\.sh":' | tr -d '":' || true)
        fi

        for existing in $existing_hooks; do
            # Check if this hook is still in source
            local still_in_source=false
            for entry in "${manifest_entries[@]:-}"; do
                if [[ "$entry" == "$existing:"* ]]; then
                    still_in_source=true
                    break
                fi
            done

            if [[ "$still_in_source" == false ]]; then
                local target_path="$USER_HOOKS_DIR/$existing"
                if [[ -f "$target_path" ]]; then
                    local checksum
                    checksum=$(calculate_checksum "$target_path")
                    local existing_location
                    if command -v jq >/dev/null 2>&1; then
                        existing_location=$(echo "$manifest" | jq -r --arg name "$existing" '.hooks[$name].location // "root"' 2>/dev/null)
                    else
                        existing_location="root"
                    fi
                    manifest_entries+=("$existing:$checksum:$existing_location")
                    log_debug "Preserved manifest entry: $existing (no longer in knossos)"
                fi
            fi
        done
    fi

    # Write updated manifest
    if [[ "$DRY_RUN_MODE" -eq 0 ]]; then
        write_manifest "${manifest_entries[@]:-}"
    fi

    # Summary (all output to stderr to keep stdout clean for data)
    echo "" >&2
    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
        log "Dry-run complete:"
    else
        log "Sync complete:"
    fi

    local total=$((added + updated + unchanged + skipped))
    echo "  Added:     $added" >&2
    echo "  Updated:   $updated" >&2
    echo "  Unchanged: $unchanged" >&2
    echo "  Skipped:   $skipped (user-created)" >&2
    echo "  Total:     $total hook(s) processed" >&2
}

# Count total hooks in categorical source structure
count_source_hooks() {
    local count=0

    # Count lib/ hooks (root exception)
    if [[ -d "$SOURCE_DIR/lib" ]]; then
        for f in "$SOURCE_DIR/lib"/*.sh; do
            [[ -f "$f" ]] && ((count++)) || true
        done
    fi

    # Count categorized hooks
    for category in $HOOK_CATEGORIES; do
        local category_dir="$SOURCE_DIR/$category"
        [[ -d "$category_dir" ]] || continue
        for f in "$category_dir"/*.sh; do
            [[ -f "$f" ]] && ((count++)) || true
        done
    done

    echo "$count"
}

# Show sync status
show_status() {
    echo "User-Hooks Sync Status"
    echo "======================"
    echo ""
    echo "Source:  $SOURCE_DIR (categorical)"
    echo "Target:  $USER_HOOKS_DIR (flat)"
    echo ""

    # Check source - count by category
    if [[ -d "$SOURCE_DIR" ]]; then
        local source_count
        source_count=$(count_source_hooks)
        echo "Knossos hooks:    $source_count"

        # Show breakdown by category
        if [[ -d "$SOURCE_DIR/lib" ]]; then
            local lib_count=0
            for f in "$SOURCE_DIR/lib"/*.sh; do
                [[ -f "$f" ]] && ((lib_count++)) || true
            done
            [[ "$lib_count" -gt 0 ]] && echo "  lib (root): $lib_count"
        fi

        for category in $HOOK_CATEGORIES; do
            local category_dir="$SOURCE_DIR/$category"
            [[ -d "$category_dir" ]] || continue
            local cat_count=0
            for f in "$category_dir"/*.sh; do
                [[ -f "$f" ]] && ((cat_count++)) || true
            done
            [[ "$cat_count" -gt 0 ]] && echo "  $category: $cat_count"
        done
    else
        echo "Knossos hooks:    (directory not found)"
    fi

    echo ""

    # Check target (flat)
    if [[ -d "$USER_HOOKS_DIR" ]]; then
        local target_root=0
        local target_lib=0

        target_root=$(find "$USER_HOOKS_DIR" -maxdepth 1 -name "*.sh" -type f 2>/dev/null | wc -l | tr -d ' ')
        if [[ -d "$USER_HOOKS_DIR/lib" ]]; then
            target_lib=$(find "$USER_HOOKS_DIR/lib" -maxdepth 1 -name "*.sh" -type f 2>/dev/null | wc -l | tr -d ' ')
        fi

        local target_count=$((target_root + target_lib))
        echo "User hooks:      $target_count"
        echo "  root:          $target_root"
        echo "  lib/:          $target_lib"
    else
        echo "User hooks:      (directory not found)"
    fi

    # Check manifest
    if [[ -f "$USER_MANIFEST_FILE" ]]; then
        local manifest_count last_sync
        manifest_count=$(grep -c '"source": "knossos"' "$USER_MANIFEST_FILE" 2>/dev/null || echo "0")
        last_sync=$(grep '"last_sync"' "$USER_MANIFEST_FILE" | sed 's/.*"last_sync":[[:space:]]*"\([^"]*\)".*/\1/' || echo "unknown")
        echo "Knossos-managed:  $manifest_count"
        echo "Last sync:       $last_sync"
    else
        echo "Manifest:        (not initialized)"
    fi

    echo ""

    # Show detailed comparison
    if [[ -d "$SOURCE_DIR" ]] && [[ -d "$USER_HOOKS_DIR" ]]; then
        echo "Hook Status:"
        echo "------------"

        # Track hooks we've seen from source
        local source_hooks=()

        # Check lib/ hooks (root exception)
        if [[ -d "$SOURCE_DIR/lib" ]]; then
            for source_file in "$SOURCE_DIR/lib"/*.sh; do
                [[ -f "$source_file" ]] || continue

                local file_name hook_name
                file_name=$(basename "$source_file")
                hook_name="lib/$file_name"
                source_hooks+=("$hook_name")
                local target_file="$USER_HOOKS_DIR/lib/$file_name"

                if [[ -f "$target_file" ]]; then
                    if is_knossos_managed "$hook_name"; then
                        local source_checksum manifest_checksum
                        source_checksum=$(calculate_checksum "$source_file")
                        manifest_checksum=$(get_manifest_checksum "$hook_name")

                        if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                            echo "  [=] $hook_name (root, up to date)"
                        else
                            echo "  [~] $hook_name (root, update available)"
                        fi
                    else
                        echo "  [!] $hook_name (root, user-created, would skip)"
                    fi
                else
                    echo "  [+] $hook_name (root, would add)"
                fi
            done
        fi

        # Check categorized hooks
        for category in $HOOK_CATEGORIES; do
            local category_dir="$SOURCE_DIR/$category"
            [[ -d "$category_dir" ]] || continue

            for source_file in "$category_dir"/*.sh; do
                [[ -f "$source_file" ]] || continue

                local hook_name
                hook_name=$(basename "$source_file")
                source_hooks+=("$hook_name")
                local target_file="$USER_HOOKS_DIR/$hook_name"

                if [[ -f "$target_file" ]]; then
                    if is_knossos_managed "$hook_name"; then
                        local source_checksum manifest_checksum
                        source_checksum=$(calculate_checksum "$source_file")
                        manifest_checksum=$(get_manifest_checksum "$hook_name")

                        if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                            echo "  [=] $hook_name ($category, up to date)"
                        else
                            echo "  [~] $hook_name ($category, update available)"
                        fi
                    else
                        echo "  [!] $hook_name ($category, user-created, would skip)"
                    fi
                else
                    echo "  [+] $hook_name ($category, would add)"
                fi
            done
        done

        # Check for user hooks not in knossos
        for target_file in "$USER_HOOKS_DIR"/*.sh; do
            [[ -f "$target_file" ]] || continue

            local hook_name
            hook_name=$(basename "$target_file")

            # Check if this hook was in source
            local in_source=false
            for src in "${source_hooks[@]:-}"; do
                [[ "$src" == "$hook_name" ]] && in_source=true && break
            done

            if [[ "$in_source" == false ]]; then
                if is_knossos_managed "$hook_name"; then
                    echo "  [-] $hook_name (was from knossos, now removed from source)"
                else
                    echo "  [*] $hook_name (user-created)"
                fi
            fi
        done

        # Check lib/ user hooks
        if [[ -d "$USER_HOOKS_DIR/lib" ]]; then
            for target_file in "$USER_HOOKS_DIR/lib"/*.sh; do
                [[ -f "$target_file" ]] || continue

                local file_name hook_name
                file_name=$(basename "$target_file")
                hook_name="lib/$file_name"

                # Check if this hook was in source
                local in_source=false
                for src in "${source_hooks[@]:-}"; do
                    [[ "$src" == "$hook_name" ]] && in_source=true && break
                done

                if [[ "$in_source" == false ]]; then
                    if is_knossos_managed "$hook_name"; then
                        echo "  [-] $hook_name (was from knossos, now removed from source)"
                    else
                        echo "  [*] $hook_name (user-created)"
                    fi
                fi
            done
        fi
    fi
}

# Usage information
usage() {
    cat <<EOF
Usage: sync-user-hooks.sh [OPTIONS]

Syncs knossos hooks to ~/.claude/hooks/

Options:
  --dry-run      Preview changes without applying
  --status       Show sync status without making changes
  --adopt        Recover manifest from existing hooks (bootstrap/repair)
  --force        Overwrite user-created hooks (use with caution)
  --help, -h     Show this help message

Behavior:
  - Additive:    Never removes existing hooks from ~/.claude/hooks/
  - Overwrites:  Only hooks previously installed from knossos
  - Preserves:   User-created hooks not from knossos
  - Nested:      lib/ subdirectory is preserved with structure

The manifest at ~/.claude/USER_HOOKS_MANIFEST.json tracks which hooks
were installed from knossos, allowing safe updates while preserving
user-created hooks.

Adopt Mode (--adopt):
  Scans existing hooks in ~/.claude/hooks/ and matches them against
  knossos sources. Hooks that match are adopted into the manifest:
  - Exact matches: marked as "knossos" (fully managed)
  - Diverged files: marked as "knossos-diverged" (preserves local changes)
  - User-created: not added to manifest (remain user-owned)

  Use --adopt when:
  - First-time setup with existing hooks
  - Manifest was deleted or corrupted
  - Hooks were installed before manifest tracking existed

Force Mode (--force):
  Overwrites user-created hooks with knossos versions. Use with caution
  as this will replace any local modifications.

Source Structure:
  knossos/.claude/hooks/
    *.sh             # Root-level hooks (artifact-tracker, session-context, etc.)
    lib/             # Shared hook utilities
      config.sh
      logging.sh
      primitives.sh
      session-*.sh
      worktree-manager.sh

Environment Variables:
  KNOSSOS_HOME   Knossos platform location (default: ~/Code/knossos)
  KNOSSOS_DEBUG   Enable debug logging (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Source directory missing
  3  Sync failure

Examples:
  ./sync-user-hooks.sh              # Sync hooks
  ./sync-user-hooks.sh --dry-run    # Preview what would change
  ./sync-user-hooks.sh --status     # Show current sync status
  ./sync-user-hooks.sh --adopt      # Recover manifest from existing files
  ./sync-user-hooks.sh --adopt --dry-run  # Preview adopt results
  ./sync-user-hooks.sh --force      # Overwrite all hooks (use with caution)

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
            --force)
                FORCE_MODE=1
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
    mkdir -p "$USER_HOOKS_DIR"

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
