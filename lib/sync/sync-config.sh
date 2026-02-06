#!/usr/bin/env bash
#
# sync-config.sh - Sync Configuration and Constants
#
# Defines file classification lists, exit codes, and configuration
# for the roster-sync ecosystem synchronization system.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$KNOSSOS_HOME/lib/sync/sync-config.sh"
#
# Functions:
#   get_copy_replace_items  - Files completely replaced from roster
#   get_merge_items         - Files using intelligent merge strategies
#   get_merge_dir_items     - Directories that sync contents
#   get_ignore_items        - Files/directories never touched by sync

# Guard against re-sourcing
[[ -n "${_SYNC_CONFIG_LOADED:-}" ]] && return 0
readonly _SYNC_CONFIG_LOADED=1

# ============================================================================
# Exit Codes (per TDD 3.1)
# ============================================================================

readonly EXIT_SYNC_SUCCESS=0
readonly EXIT_SYNC_ERROR=1
readonly EXIT_SYNC_VALIDATION_WARNINGS=1
readonly EXIT_SYNC_VALIDATION_FAILURE=2
readonly EXIT_SYNC_INIT_FAILED=3
readonly EXIT_SYNC_INVALID_MANIFEST=4
readonly EXIT_SYNC_CONFLICTS=5
readonly EXIT_SYNC_ORPHAN_CONFLICTS=6

# ============================================================================
# Schema Constants
# ============================================================================

readonly SYNC_SCHEMA_VERSION=3
# These can be overridden for testing, so not readonly
: "${SYNC_MANIFEST_FILE:=.claude/.cem/manifest.json}"
: "${SYNC_CHECKSUM_CACHE:=.claude/.cem/checksum-cache.json}"
: "${SYNC_ORPHAN_BACKUP_DIR:=.claude/.cem/orphan-backup}"

# ============================================================================
# File Classification Functions (per TDD 5.2)
# ============================================================================

# Files that are completely replaced from roster
# These files are owned by roster - local changes are overwritten on sync
get_copy_replace_items() {
    cat <<'EOF'
COMMAND_REGISTRY.md
forge-workflow.yaml
EOF
}

# Files that use intelligent merging
# Format: filename:strategy
get_merge_items() {
    cat <<'EOF'
settings.local.json:merge-settings
CLAUDE.md:merge-docs
EOF
}

# Directories that sync contents (preserving satellite-specific files)
# Reserved for future use
get_merge_dir_items() {
    cat <<'EOF'
EOF
}

# Files/directories never touched by sync
# These are satellite-owned or session-specific
get_ignore_items() {
    cat <<'EOF'
ACTIVE_RITE
ACTIVE_WORKFLOW.yaml
sessions
agents
agents.backup
.cem
.archive
user-agents
mena
user-skills
user-hooks
commands
skills
hooks
PROJECT.md
EOF
}

# ============================================================================
# Utility Functions
# ============================================================================

# Check if a file is in the ignore list
# Usage: is_ignored "filename"
# Returns: 0 if ignored, 1 if not
is_ignored() {
    local filename="$1"
    local ignored_item

    while IFS= read -r ignored_item; do
        [[ -z "$ignored_item" ]] && continue
        if [[ "$filename" == "$ignored_item" ]]; then
            return 0
        fi
    done < <(get_ignore_items)

    return 1
}

# Get the merge strategy for a file
# Usage: get_merge_strategy "filename"
# Returns: strategy name or empty if not a merge file
get_merge_strategy() {
    local filename="$1"
    local line strategy

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        local file="${line%%:*}"
        if [[ "$filename" == "$file" ]]; then
            strategy="${line#*:}"
            echo "$strategy"
            return 0
        fi
    done < <(get_merge_items)

    echo ""
    return 1
}

# Check if a file is a copy-replace item
# Usage: is_copy_replace "filename"
# Returns: 0 if copy-replace, 1 if not
is_copy_replace() {
    local filename="$1"
    local item

    while IFS= read -r item; do
        [[ -z "$item" ]] && continue
        if [[ "$filename" == "$item" ]]; then
            return 0
        fi
    done < <(get_copy_replace_items)

    return 1
}

# Get all managed file patterns (copy-replace + merge items)
# Returns: list of filenames (without strategy suffix for merge items)
get_all_managed_files() {
    local item line

    # Copy-replace items
    while IFS= read -r item; do
        [[ -n "$item" ]] && echo "$item"
    done < <(get_copy_replace_items)

    # Merge items (extract filename only)
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        echo "${line%%:*}"
    done < <(get_merge_items)
}

# ============================================================================
# Logging (sync-specific prefixes)
# ============================================================================

# Note: These can be overridden by sourcing script
# Defaults to simple echo if not defined

sync_log() {
    if declare -F log >/dev/null 2>&1; then
        log "$@"
    else
        echo "[roster-sync] $*" >&2
    fi
}

sync_log_error() {
    if declare -F log_error >/dev/null 2>&1; then
        log_error "$@"
    else
        echo "[roster-sync] Error: $*" >&2
    fi
}

sync_log_warning() {
    if declare -F log_warning >/dev/null 2>&1; then
        log_warning "$@"
    else
        echo "[roster-sync] Warning: $*" >&2
    fi
}

sync_log_debug() {
    if declare -F log_debug >/dev/null 2>&1; then
        log_debug "$@"
    elif [[ "${ROSTER_SYNC_DEBUG:-0}" == "1" ]]; then
        echo "[roster-sync DEBUG] $*" >&2
    fi
}
