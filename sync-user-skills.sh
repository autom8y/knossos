#!/usr/bin/env bash
#
# sync-user-skills.sh - Sync roster user-skills to ~/.claude/skills/
#
# Syncs skills from roster/user-skills/ to the user-level skills directory.
# Skills are directories containing SKILL.md and supporting files.
#
# Behavior:
#   - Additive: Never removes existing skills from ~/.claude/skills/
#   - Overwrites: Only skills previously installed from roster (tracked in manifest)
#   - Preserves: User-created skills not from roster
#
# Usage:
#   ./sync-user-skills.sh              # Sync user-skills to ~/.claude/skills/
#   ./sync-user-skills.sh --dry-run    # Preview changes without applying
#   ./sync-user-skills.sh --status     # Show sync status
#   ./sync-user-skills.sh --help       # Show usage
#
# Environment Variables:
#   ROSTER_HOME    Roster repository location (default: ~/Code/roster)
#   ROSTER_DEBUG   Enable debug logging (set to 1)

set -euo pipefail

# Constants
readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly ROSTER_DEBUG="${ROSTER_DEBUG:-0}"
readonly USER_SKILLS_DIR="$HOME/.claude/skills"
readonly USER_MANIFEST_FILE="$HOME/.claude/USER_SKILL_MANIFEST.json"
readonly SOURCE_DIR="$ROSTER_HOME/user-skills"
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
    echo "[User-Skills] $*" >&2
}

log_success() {
    echo -e "[User-Skills] ${GREEN}$*${NC}" >&2
}

log_info() {
    echo -e "[User-Skills] ${BLUE}$*${NC}" >&2
}

log_warning() {
    echo -e "[User-Skills] ${YELLOW}Warning:${NC} $*" >&2
}

log_error() {
    echo "[User-Skills] Error: $*" >&2
}

log_debug() {
    if [[ "$ROSTER_DEBUG" == "1" ]]; then
        echo "[User-Skills DEBUG] $*" >&2
    fi
}

# ============================================================================
# Checksum Functions
# ============================================================================

# Calculate checksum of a skill directory (hash of all file contents, sorted by path)
# This ensures consistent checksums across systems
calculate_skill_checksum() {
    local skill_dir="$1"
    if command -v shasum >/dev/null 2>&1; then
        find "$skill_dir" -type f -print0 | sort -z | \
            xargs -0 cat 2>/dev/null | shasum -a 256 | cut -d' ' -f1
    elif command -v sha256sum >/dev/null 2>&1; then
        find "$skill_dir" -type f -print0 | sort -z | \
            xargs -0 cat 2>/dev/null | sha256sum | cut -d' ' -f1
    else
        # Fallback to md5 if sha256 unavailable
        if command -v md5 >/dev/null 2>&1; then
            find "$skill_dir" -type f -print0 | sort -z | \
                xargs -0 cat 2>/dev/null | md5 -q
        else
            find "$skill_dir" -type f -print0 | sort -z | \
                xargs -0 cat 2>/dev/null | md5sum | cut -d' ' -f1
        fi
    fi
}

# Count files in a skill directory
count_skill_files() {
    local skill_dir="$1"
    find "$skill_dir" -type f | wc -l | tr -d ' '
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

# Check if skill is managed by roster (exists in manifest with source=roster)
is_roster_managed() {
    local skill_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        return 1
    fi

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        local source
        source=$(echo "$manifest" | jq -r --arg name "$skill_name" '.skills[$name].source // empty' 2>/dev/null)
        if [[ "$source" == "roster" || "$source" == "roster-diverged" ]]; then
            return 0
        fi
        return 1
    fi

    # Fallback to grep-based parsing
    local skill_block
    skill_block=$(echo "$manifest" | grep -A4 "\"$skill_name\":" 2>/dev/null | head -5)

    if [[ -z "$skill_block" ]]; then
        return 1
    fi

    local source
    source=$(echo "$skill_block" | grep '"source"' | sed 's/.*"source":[[:space:]]*"\([^"]*\)".*/\1/')

    if [[ "$source" == "roster" || "$source" == "roster-diverged" ]]; then
        return 0
    fi

    return 1
}

# Add or update a single manifest entry for a skill
# Usage: add_to_manifest skill_name source_type checksum file_count
add_to_manifest() {
    local skill_name="$1"
    local source_type="$2"
    local checksum="$3"
    local file_count="$4"
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
        updated=$(echo "$manifest" | jq --arg name "$skill_name" \
            --arg src "$source_type" \
            --arg ts "$timestamp" \
            --arg cs "$checksum" \
            --argjson fc "$file_count" \
            '.skills[$name] = {"source": $src, "installed_at": $ts, "checksum": $cs, "file_count": $fc}')
        echo "$updated" > "$USER_MANIFEST_FILE"
    else
        log_warning "jq not available, cannot update manifest entry for $skill_name"
    fi
}

# Recover manifest entries from existing skill directories that match roster sources
recover_manifest() {
    log_info "Recovering manifest from existing skills..."

    local target_dir="$USER_SKILLS_DIR"
    local recovered=0
    local diverged=0

    # Ensure source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        return 1
    fi

    # Process each skill directory in target
    for target_skill in "$target_dir"/*/; do
        [[ -d "$target_skill" ]] || continue

        # Remove trailing slash
        target_skill="${target_skill%/}"
        local skill_name
        skill_name=$(basename "$target_skill")

        # Skip if not a valid skill (no SKILL.md)
        [[ -f "$target_skill/SKILL.md" ]] || continue

        # Skip if already in manifest as roster-managed
        if is_roster_managed "$skill_name"; then
            log_debug "Already managed: $skill_name"
            continue
        fi

        # Check if this skill exists in roster source
        local source_skill="$SOURCE_DIR/$skill_name"
        if [[ -d "$source_skill" && -f "$source_skill/SKILL.md" ]]; then
            local source_checksum target_checksum source_file_count target_file_count
            source_checksum=$(calculate_skill_checksum "$source_skill")
            target_checksum=$(calculate_skill_checksum "$target_skill")
            source_file_count=$(count_skill_files "$source_skill")
            target_file_count=$(count_skill_files "$target_skill")

            if [[ "$source_checksum" == "$target_checksum" ]]; then
                # Exact match - adopt as roster-managed
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt: $skill_name (exact match, $target_file_count files)"
                else
                    add_to_manifest "$skill_name" "roster" "$target_checksum" "$target_file_count"
                    log_success "Adopted: $skill_name (exact match, $target_file_count files)"
                fi
                ((recovered++)) || true
            else
                # Diverged - mark as roster-diverged to preserve user changes
                if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                    log_info "Would adopt (diverged): $skill_name (local modifications preserved)"
                else
                    add_to_manifest "$skill_name" "roster-diverged" "$target_checksum" "$target_file_count"
                    log_warning "Adopted (diverged): $skill_name (local modifications preserved)"
                fi
                ((diverged++)) || true
            fi
        else
            log_debug "Not in roster: $skill_name (user-created)"
        fi
    done

    log_info "Recovery complete: $recovered adopted, $diverged diverged"
}

# Get checksum from manifest for a skill
get_manifest_checksum() {
    local skill_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        echo ""
        return
    fi

    # Use jq if available for reliable JSON parsing
    if command -v jq >/dev/null 2>&1; then
        echo "$manifest" | jq -r --arg name "$skill_name" '.skills[$name].checksum // empty' 2>/dev/null
        return
    fi

    # Fallback to grep-based parsing
    local skill_block
    skill_block=$(echo "$manifest" | grep -A5 "\"$skill_name\":" 2>/dev/null | head -6)

    if [[ -z "$skill_block" ]]; then
        echo ""
        return
    fi

    echo "$skill_block" | grep '"checksum"' | sed 's/.*"checksum":[[:space:]]*"\([^"]*\)".*/\1/'
}

# Get file_count from manifest for a skill
get_manifest_file_count() {
    local skill_name="$1"
    local manifest
    manifest=$(read_manifest)

    if [[ -z "$manifest" ]]; then
        echo "0"
        return
    fi

    local skill_block
    skill_block=$(echo "$manifest" | grep -A5 "\"$skill_name\":" 2>/dev/null | head -6)

    if [[ -z "$skill_block" ]]; then
        echo "0"
        return
    fi

    echo "$skill_block" | grep '"file_count"' | sed 's/.*"file_count":[[:space:]]*\([0-9]*\).*/\1/'
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
  "skills": {}
}
EOF

    log_debug "Initialized empty manifest at $USER_MANIFEST_FILE"
}

# Write manifest with current roster-managed skills
# Usage: write_manifest "skill1:checksum1:count1" "skill2:checksum2:count2" ...
write_manifest() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    mkdir -p "$(dirname "$USER_MANIFEST_FILE")"

    # Start JSON
    {
        echo "{"
        echo "  \"manifest_version\": \"$MANIFEST_VERSION\","
        echo "  \"last_sync\": \"$timestamp\","
        echo "  \"skills\": {"
    } > "$USER_MANIFEST_FILE"

    # Add each skill entry
    local first=true
    for entry in "$@"; do
        # Skip empty entries
        [[ -z "$entry" ]] && continue

        local skill_name checksum file_count
        skill_name=$(echo "$entry" | cut -d: -f1)
        checksum=$(echo "$entry" | cut -d: -f2)
        file_count=$(echo "$entry" | cut -d: -f3)

        # Skip entries with empty skill name
        [[ -z "$skill_name" ]] && continue

        # Add comma separator
        if [[ "$first" == true ]]; then
            first=false
        else
            echo "," >> "$USER_MANIFEST_FILE"
        fi

        # Write skill entry
        {
            echo -n "    \"$skill_name\": {"
            echo -n "\"source\": \"roster\", "
            echo -n "\"installed_at\": \"$timestamp\", "
            echo -n "\"checksum\": \"$checksum\", "
            echo -n "\"file_count\": $file_count"
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
    log_debug "Starting sync from $SOURCE_DIR to $USER_SKILLS_DIR"

    # Check if source directory exists
    if [[ ! -d "$SOURCE_DIR" ]]; then
        log_error "Source directory not found: $SOURCE_DIR"
        log "Create the directory and add skill directories to sync"
        exit "$EXIT_SOURCE_MISSING"
    fi

    # Ensure target directory exists
    mkdir -p "$USER_SKILLS_DIR"

    # Initialize manifest if it doesn't exist
    if [[ ! -f "$USER_MANIFEST_FILE" ]]; then
        log_debug "No manifest found, initializing"
        init_manifest
    fi

    # Track skills to write to manifest
    local manifest_entries=()
    local added=0
    local updated=0
    local skipped=0
    local unchanged=0

    # Process each skill directory in source
    for source_skill_path in "$SOURCE_DIR"/*/; do
        # Skip if not a directory
        [[ -d "$source_skill_path" ]] || continue

        # Remove trailing slash for consistent handling
        local source_skill="${source_skill_path%/}"

        # Skip if no SKILL.md file (not a valid skill)
        [[ -f "${source_skill}/SKILL.md" ]] || continue

        local skill_name
        skill_name=$(basename "$source_skill")
        local target_skill="$USER_SKILLS_DIR/$skill_name"
        local source_checksum
        source_checksum=$(calculate_skill_checksum "$source_skill")
        local source_file_count
        source_file_count=$(count_skill_files "$source_skill")

        log_debug "Processing: $skill_name (checksum: ${source_checksum:0:8}..., files: $source_file_count)"

        if [[ -d "$target_skill" ]]; then
            # Target exists - check if we can overwrite
            if is_roster_managed "$skill_name"; then
                # Roster-managed: check if update needed
                local manifest_checksum
                manifest_checksum=$(get_manifest_checksum "$skill_name")

                if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                    # No change needed
                    log_debug "Unchanged: $skill_name"
                    ((unchanged++)) || true
                    manifest_entries+=("$skill_name:$source_checksum:$source_file_count")
                else
                    # Update needed - use rsync with delete for clean update
                    if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                        local manifest_file_count
                        manifest_file_count=$(get_manifest_file_count "$skill_name")
                        log_info "Would update: $skill_name ($manifest_file_count -> $source_file_count files)"
                    else
                        rsync -a --delete "$source_skill" "$USER_SKILLS_DIR/"
                        log_success "Updated: $skill_name ($source_file_count files)"
                    fi
                    ((updated++)) || true
                    manifest_entries+=("$skill_name:$source_checksum:$source_file_count")
                fi
            else
                # User-created: skip with warning
                log_warning "Skipped: $skill_name (user-created, not overwriting)"
                ((skipped++)) || true
                # Do NOT add to manifest - preserve user ownership
            fi
        else
            # Target doesn't exist - add new skill
            if [[ "$DRY_RUN_MODE" -eq 1 ]]; then
                log_info "Would add: $skill_name ($source_file_count files)"
            else
                rsync -a "$source_skill" "$USER_SKILLS_DIR/"
                log_success "Added: $skill_name ($source_file_count files)"
            fi
            ((added++)) || true
            manifest_entries+=("$skill_name:$source_checksum:$source_file_count")
        fi
    done

    # Preserve existing roster-managed skills that are no longer in source
    # (This handles the case where roster removes a skill - we still track it
    # but don't remove it, honoring the additive-only requirement)
    local manifest
    manifest=$(read_manifest)
    if [[ -n "$manifest" ]]; then
        # Extract existing skills from manifest
        local existing_skills
        existing_skills=$(echo "$manifest" | grep -o '"[^"]*":' | grep -v 'manifest_version\|last_sync\|skills\|source\|installed_at\|checksum\|file_count' | tr -d '":' || true)

        for existing in $existing_skills; do
            # Check if this skill is still in source
            local still_in_source=false
            for entry in "${manifest_entries[@]:-}"; do
                if [[ "$entry" == "$existing:"* ]]; then
                    still_in_source=true
                    break
                fi
            done

            if [[ "$still_in_source" == false ]] && [[ -d "$USER_SKILLS_DIR/$existing" ]]; then
                # Skill removed from roster but still exists - keep in manifest
                # so we know it came from roster originally
                local checksum file_count
                checksum=$(calculate_skill_checksum "$USER_SKILLS_DIR/$existing")
                file_count=$(count_skill_files "$USER_SKILLS_DIR/$existing")
                manifest_entries+=("$existing:$checksum:$file_count")
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
    echo "  Total:     $total skill(s) processed"
}

# Show sync status
show_status() {
    echo "User-Skills Sync Status"
    echo "======================="
    echo ""
    echo "Source:  $SOURCE_DIR"
    echo "Target:  $USER_SKILLS_DIR"
    echo ""

    # Check source
    if [[ -d "$SOURCE_DIR" ]]; then
        local source_count
        source_count=$(find "$SOURCE_DIR" -maxdepth 1 -type d -name "*" ! -path "$SOURCE_DIR" 2>/dev/null | wc -l | tr -d ' ')
        echo "Roster skills:  $source_count"
    else
        echo "Roster skills:  (directory not found)"
    fi

    # Check target
    if [[ -d "$USER_SKILLS_DIR" ]]; then
        local target_count
        target_count=$(find "$USER_SKILLS_DIR" -maxdepth 1 -type d -name "*" ! -path "$USER_SKILLS_DIR" 2>/dev/null | wc -l | tr -d ' ')
        echo "User skills:    $target_count"
    else
        echo "User skills:    (directory not found)"
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
    if [[ -d "$SOURCE_DIR" ]] && [[ -d "$USER_SKILLS_DIR" ]]; then
        echo "Skill Status:"
        echo "-------------"

        # Check each source skill
        for source_skill in "$SOURCE_DIR"/*/; do
            [[ -d "$source_skill" ]] || continue
            [[ -f "${source_skill}SKILL.md" ]] || continue

            local skill_name
            skill_name=$(basename "$source_skill")
            local target_skill="$USER_SKILLS_DIR/$skill_name"

            if [[ -d "$target_skill" ]]; then
                if is_roster_managed "$skill_name"; then
                    local source_checksum manifest_checksum source_file_count
                    source_checksum=$(calculate_skill_checksum "$source_skill")
                    manifest_checksum=$(get_manifest_checksum "$skill_name")
                    source_file_count=$(count_skill_files "$source_skill")

                    if [[ "$source_checksum" == "$manifest_checksum" ]]; then
                        echo "  [=] $skill_name (up to date, $source_file_count files)"
                    else
                        local manifest_file_count
                        manifest_file_count=$(get_manifest_file_count "$skill_name")
                        echo "  [~] $skill_name (update available, $manifest_file_count -> $source_file_count files)"
                    fi
                else
                    echo "  [!] $skill_name (user-created, would skip)"
                fi
            else
                local source_file_count
                source_file_count=$(count_skill_files "$source_skill")
                echo "  [+] $skill_name (would add, $source_file_count files)"
            fi
        done

        # Check for user skills not in roster
        for target_skill in "$USER_SKILLS_DIR"/*/; do
            [[ -d "$target_skill" ]] || continue

            local skill_name
            skill_name=$(basename "$target_skill")

            if [[ ! -d "$SOURCE_DIR/$skill_name" ]]; then
                if is_roster_managed "$skill_name"; then
                    echo "  [-] $skill_name (was from roster, now removed from source)"
                else
                    echo "  [*] $skill_name (user-created)"
                fi
            fi
        done
    fi
}

# Usage information
usage() {
    cat <<EOF
Usage: sync-user-skills.sh [OPTIONS]

Syncs roster user-skills to ~/.claude/skills/

Options:
  --dry-run      Preview changes without applying
  --status       Show sync status without making changes
  --adopt        Recover manifest from existing skills (bootstrap/repair)
  --help, -h     Show this help message

Behavior:
  - Additive:   Never removes existing skills from ~/.claude/skills/
  - Overwrites: Only skills previously installed from roster
  - Preserves:  User-created skills not from roster

Skills are directories containing SKILL.md and supporting files.
Updates use rsync --delete to ensure clean sync within each skill directory.

The manifest at ~/.claude/USER_SKILL_MANIFEST.json tracks which skills
were installed from roster, allowing safe updates while preserving
user-created skills.

Adopt Mode (--adopt):
  Scans existing skills in ~/.claude/skills/ and matches them against
  roster sources. Skills that match are adopted into the manifest:
  - Exact matches: marked as "roster" (fully managed)
  - Diverged skills: marked as "roster-diverged" (preserves local changes)
  - User-created: not added to manifest (remain user-owned)

  Use --adopt when:
  - First-time setup with existing skills
  - Manifest was deleted or corrupted
  - Skills were installed before manifest tracking existed

Environment Variables:
  ROSTER_HOME    Roster repository location (default: ~/Code/roster)
  ROSTER_DEBUG   Enable debug logging (set to 1)

Exit Codes:
  0  Success
  1  Invalid arguments
  2  Source directory missing
  3  Sync failure

Examples:
  ./sync-user-skills.sh              # Sync user-skills
  ./sync-user-skills.sh --dry-run    # Preview what would change
  ./sync-user-skills.sh --status     # Show current sync status
  ./sync-user-skills.sh --adopt      # Recover manifest from existing skills
  ./sync-user-skills.sh --adopt --dry-run  # Preview adopt results

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
        mkdir -p "$USER_SKILLS_DIR"
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
