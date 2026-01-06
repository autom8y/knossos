#!/usr/bin/env bash
#
# sync-checksum.sh - Cross-Platform SHA-256 Checksum Utilities
#
# Provides consistent checksum computation across macOS (shasum) and
# Linux (sha256sum) with optional caching for performance.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$KNOSSOS_HOME/lib/sync/sync-checksum.sh"
#   init_checksum_cache
#   compute_checksum "/path/to/file"
#   save_checksum_cache
#
# Functions:
#   detect_checksum_tool    - Detect available checksum command
#   compute_checksum        - Compute SHA-256 for a file
#   init_checksum_cache     - Initialize in-memory cache
#   get_cached_checksum     - Get checksum from cache
#   set_cached_checksum     - Store checksum in cache
#   save_checksum_cache     - Persist cache to disk
#   load_checksum_cache     - Load cache from disk

# Guard against re-sourcing
[[ -n "${_SYNC_CHECKSUM_LOADED:-}" ]] && return 0
readonly _SYNC_CHECKSUM_LOADED=1

# ============================================================================
# Tool Detection (per TDD 10.3)
# ============================================================================

# Detect the available checksum command
# Sets CHECKSUM_CMD global variable
detect_checksum_tool() {
    if command -v shasum &>/dev/null; then
        # macOS and some Linux
        CHECKSUM_CMD="shasum -a 256"
    elif command -v sha256sum &>/dev/null; then
        # Most Linux distributions
        CHECKSUM_CMD="sha256sum"
    else
        sync_log_error "No SHA-256 tool found. Install shasum or sha256sum."
        return 1
    fi

    sync_log_debug "Using checksum tool: $CHECKSUM_CMD"
    return 0
}

# Global checksum command (set by detect_checksum_tool)
CHECKSUM_CMD=""

# ============================================================================
# Checksum Computation
# ============================================================================

# Compute SHA-256 checksum for a file
# Usage: compute_checksum "/path/to/file"
# Returns: 64-character hex checksum or empty on error
compute_checksum() {
    local file="$1"

    # Ensure tool is detected
    if [[ -z "$CHECKSUM_CMD" ]]; then
        detect_checksum_tool || return 1
    fi

    # Check file exists
    if [[ ! -f "$file" ]]; then
        sync_log_debug "Cannot compute checksum: file not found: $file"
        echo ""
        return 1
    fi

    # Check cache first
    local cached
    cached=$(get_cached_checksum "$file")
    if [[ -n "$cached" ]]; then
        sync_log_debug "Cache hit for: $file"
        echo "$cached"
        return 0
    fi

    # Compute checksum
    local checksum
    checksum=$($CHECKSUM_CMD "$file" 2>/dev/null | awk '{print $1}')

    if [[ -z "$checksum" ]]; then
        sync_log_error "Failed to compute checksum for: $file"
        return 1
    fi

    # Cache the result
    set_cached_checksum "$file" "$checksum"

    echo "$checksum"
    return 0
}

# Compute checksum for a string content (no file needed)
# Usage: compute_content_checksum "string content"
# Returns: 64-character hex checksum
compute_content_checksum() {
    local content="$1"

    # Ensure tool is detected
    if [[ -z "$CHECKSUM_CMD" ]]; then
        detect_checksum_tool || return 1
    fi

    local checksum
    checksum=$(echo -n "$content" | $CHECKSUM_CMD | awk '{print $1}')

    echo "$checksum"
}

# ============================================================================
# Checksum Cache (in-memory with file persistence)
# ============================================================================

# Associative array for in-memory cache
# Key: absolute file path
# Value: checksum:mtime (checksum with modification time for validation)
declare -gA _CHECKSUM_CACHE

# Initialize the checksum cache
# Loads from disk if cache file exists
init_checksum_cache() {
    # Clear in-memory cache
    _CHECKSUM_CACHE=()

    # Load from disk if available
    load_checksum_cache

    sync_log_debug "Checksum cache initialized (${#_CHECKSUM_CACHE[@]} entries)"
}

# Get checksum from cache
# Returns: checksum if valid (mtime matches), empty otherwise
get_cached_checksum() {
    local file="$1"
    local abs_path

    # Normalize to absolute path
    abs_path=$(cd "$(dirname "$file")" 2>/dev/null && pwd)/$(basename "$file")

    if [[ -z "${_CHECKSUM_CACHE[$abs_path]:-}" ]]; then
        return 1
    fi

    local cached="${_CHECKSUM_CACHE[$abs_path]}"
    local cached_checksum="${cached%%:*}"
    local cached_mtime="${cached#*:}"

    # Validate mtime
    local current_mtime
    current_mtime=$(stat -f "%m" "$file" 2>/dev/null || stat -c "%Y" "$file" 2>/dev/null)

    if [[ "$cached_mtime" == "$current_mtime" ]]; then
        echo "$cached_checksum"
        return 0
    fi

    # Cache invalid (file modified)
    unset "_CHECKSUM_CACHE[$abs_path]"
    return 1
}

# Store checksum in cache
set_cached_checksum() {
    local file="$1"
    local checksum="$2"
    local abs_path

    # Normalize to absolute path
    abs_path=$(cd "$(dirname "$file")" 2>/dev/null && pwd)/$(basename "$file")

    # Get file mtime
    local mtime
    mtime=$(stat -f "%m" "$file" 2>/dev/null || stat -c "%Y" "$file" 2>/dev/null)

    _CHECKSUM_CACHE[$abs_path]="${checksum}:${mtime}"
}

# Clear cache entry for a file
clear_cached_checksum() {
    local file="$1"
    local abs_path

    abs_path=$(cd "$(dirname "$file")" 2>/dev/null && pwd)/$(basename "$file")
    unset "_CHECKSUM_CACHE[$abs_path]"
}

# ============================================================================
# Cache Persistence
# ============================================================================

# Load checksum cache from disk
load_checksum_cache() {
    local cache_file="${SYNC_CHECKSUM_CACHE:-.claude/.cem/checksum-cache.json}"

    if [[ ! -f "$cache_file" ]]; then
        return 0
    fi

    # Parse JSON cache file
    local path checksum mtime
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        path=$(echo "$line" | jq -r '.path // empty')
        checksum=$(echo "$line" | jq -r '.checksum // empty')
        mtime=$(echo "$line" | jq -r '.mtime // empty')

        if [[ -n "$path" && -n "$checksum" && -n "$mtime" ]]; then
            _CHECKSUM_CACHE[$path]="${checksum}:${mtime}"
        fi
    done < <(jq -c '.entries[]?' "$cache_file" 2>/dev/null)

    sync_log_debug "Loaded ${#_CHECKSUM_CACHE[@]} cached checksums"
}

# Save checksum cache to disk
save_checksum_cache() {
    local cache_file="${SYNC_CHECKSUM_CACHE:-.claude/.cem/checksum-cache.json}"
    local cache_dir

    cache_dir=$(dirname "$cache_file")
    if [[ ! -d "$cache_dir" ]]; then
        mkdir -p "$cache_dir" || return 1
    fi

    # Build JSON entries
    local entries="[]"
    local path cached checksum mtime

    for path in "${!_CHECKSUM_CACHE[@]}"; do
        cached="${_CHECKSUM_CACHE[$path]}"
        checksum="${cached%%:*}"
        mtime="${cached#*:}"

        entries=$(echo "$entries" | jq --arg p "$path" --arg c "$checksum" --arg m "$mtime" \
            '. + [{"path": $p, "checksum": $c, "mtime": $m}]')
    done

    # Write cache file
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n --argjson entries "$entries" --arg ts "$timestamp" '{
        "version": 1,
        "updated_at": $ts,
        "entries": $entries
    }' > "$cache_file"

    sync_log_debug "Saved ${#_CHECKSUM_CACHE[@]} cached checksums"
}

# Clear the entire cache
clear_checksum_cache() {
    _CHECKSUM_CACHE=()
    rm -f "${SYNC_CHECKSUM_CACHE:-.claude/.cem/checksum-cache.json}"
    sync_log_debug "Checksum cache cleared"
}

# ============================================================================
# Comparison Utilities
# ============================================================================

# Compare two checksums
# Usage: checksums_match "checksum1" "checksum2"
# Returns: 0 if match, 1 if different
checksums_match() {
    local cs1="$1"
    local cs2="$2"

    [[ "$cs1" == "$cs2" ]]
}

# Check if file has changed from expected checksum
# Usage: file_changed "file" "expected_checksum"
# Returns: 0 if changed, 1 if unchanged
file_changed() {
    local file="$1"
    local expected="$2"

    local current
    current=$(compute_checksum "$file")

    [[ "$current" != "$expected" ]]
}
