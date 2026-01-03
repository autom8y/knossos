#!/bin/bash
# hooks-init.sh - Unified initialization script for all hooks
# Provides standardized error handling, library sourcing, and logging setup
#
# Part of Sprint 002: Hooks Standardization
# Design: docs/design/TDD-hooks-init.md

# =============================================================================
# Core Initialization Function
# =============================================================================

# Initialize hook with appropriate error handling mode
# Usage: hooks_init <hook_name> <category>
# Categories: DEFENSIVE | RECOVERABLE
#
# Effects:
#   - Sets HOOK_NAME and HOOK_CATEGORY globals
#   - Configures shell options based on category
#   - Initializes logging
#   - Sources required libraries
#   - Sets up error trap (RECOVERABLE only)
#
# Returns: 0 always (errors are logged, not propagated)

hooks_init() {
    local hook_name="${1:-unknown}"
    local category="${2:-RECOVERABLE}"

    # Export for use in error messages
    export HOOK_NAME="$hook_name"
    export HOOK_CATEGORY="$category"

    # Determine library directory
    local lib_dir="$(dirname "${BASH_SOURCE[0]}")"

    # Source dependencies (order matters)
    # Use 2>/dev/null || true pattern to never crash during init
    source "$lib_dir/config.sh" 2>/dev/null || true
    source "$lib_dir/logging.sh" 2>/dev/null || true
    source "$lib_dir/primitives.sh" 2>/dev/null || true

    # ==========================================================================
    # OPTIMIZATION: Cache session directory on first resolution
    # Subsequent hooks in same process/subshell reuse cached value
    # Saves ~10-15ms per get_session_dir() call
    # ==========================================================================
    if [[ -z "${CACHED_SESSION_DIR:-}" ]]; then
        # Only attempt if session-core functions are available
        if type get_session_dir &>/dev/null; then
            export CACHED_SESSION_DIR
            CACHED_SESSION_DIR=$(get_session_dir 2>/dev/null || echo "")
        fi
    fi

    # Capture start time for duration tracking (always-on timing)
    export HOOK_START_TIME_MS
    HOOK_START_TIME_MS=$(get_time_ms 2>/dev/null || echo 0)

    # Initialize logging
    log_init "$hook_name" 2>/dev/null || true
    log_start 2>/dev/null || true

    # Category-specific setup
    case "$category" in
        DEFENSIVE)
            # Explicitly disable strict modes
            # DEFENSIVE hooks must never crash Claude's tool flow
            set +e +u +o pipefail 2>/dev/null || true
            ;;
        RECOVERABLE)
            # Enable strict modes with recovery trap
            # RECOVERABLE hooks can detect errors but must degrade gracefully
            _hooks_setup_recovery_trap
            set -euo pipefail
            ;;
        *)
            # Unknown category defaults to DEFENSIVE for safety
            log_warn "Unknown hook category: $category, defaulting to DEFENSIVE" 2>/dev/null || true
            set +e +u +o pipefail 2>/dev/null || true
            ;;
    esac

    return 0
}

# =============================================================================
# Error Trap for RECOVERABLE Hooks
# =============================================================================

# Internal: Set up error trap for RECOVERABLE hooks
# Catches any error, logs it, and exits 0 to prevent hook from crashing Claude
_hooks_setup_recovery_trap() {
    trap '_hooks_handle_error $? $LINENO "$BASH_COMMAND"' ERR
}

_hooks_handle_error() {
    local exit_code="$1"
    local line_number="$2"
    local command="$3"
    local duration_ms=""

    # Calculate duration if start time was captured
    if [[ -n "${HOOK_START_TIME_MS:-}" && "$HOOK_START_TIME_MS" != "0" ]]; then
        duration_ms=$(calc_duration_ms "$HOOK_START_TIME_MS" 2>/dev/null || echo "")
    fi

    # Log the error
    log_error "Hook failed at line $line_number: $command (exit $exit_code)" 2>/dev/null || true

    # Log completion with error and duration
    log_end "$exit_code" "$duration_ms" 2>/dev/null || true

    # Log to JSONL timing file
    _timing_log_jsonl "$exit_code" "$duration_ms" 2>/dev/null || true

    # Exit 0 to prevent hook from blocking Claude
    exit 0
}

# =============================================================================
# Safe Sourcing for Optional Dependencies
# =============================================================================

# Safely source optional dependency with fallback
# Usage: safe_source <file_path> [fallback_action]
#
# Returns: 0 if sourced successfully, 1 if not (never crashes)

safe_source() {
    local file_path="$1"
    local fallback="${2:-}"

    if [[ -f "$file_path" ]]; then
        source "$file_path" 2>/dev/null
        return $?
    else
        if [[ -n "$fallback" ]]; then
            log_debug "Optional dependency not found: $file_path, using fallback" 2>/dev/null || true
            eval "$fallback" 2>/dev/null || true
        fi
        return 1
    fi
}

# =============================================================================
# Cached Session Directory Access
# =============================================================================

# Get session directory with caching
# Usage: get_cached_session_dir
# Returns: Session directory path or empty string
# Note: Falls back to get_session_dir if cache miss
get_cached_session_dir() {
    if [[ -n "${CACHED_SESSION_DIR:-}" ]]; then
        echo "$CACHED_SESSION_DIR"
    elif type get_session_dir &>/dev/null; then
        # Cache miss but function available - resolve and cache
        CACHED_SESSION_DIR=$(get_session_dir 2>/dev/null || echo "")
        export CACHED_SESSION_DIR
        echo "$CACHED_SESSION_DIR"
    else
        echo ""
    fi
}

# =============================================================================
# Explicit Finalization for DEFENSIVE Hooks
# =============================================================================

# Call at end of hook to log completion
# Usage: hooks_finalize [exit_code]
#
# Note: For DEFENSIVE hooks that explicitly manage their exit
# RECOVERABLE hooks auto-finalize via trap

hooks_finalize() {
    local exit_code="${1:-0}"
    local duration_ms=""

    # Calculate duration if start time was captured
    if [[ -n "${HOOK_START_TIME_MS:-}" && "$HOOK_START_TIME_MS" != "0" ]]; then
        duration_ms=$(calc_duration_ms "$HOOK_START_TIME_MS" 2>/dev/null || echo "")
    fi

    # Log to standard hooks.log with duration
    log_end "$exit_code" "$duration_ms" 2>/dev/null || true

    # Log to JSONL timing file for analysis
    _timing_log_jsonl "$exit_code" "$duration_ms" 2>/dev/null || true
}

# Internal: Log timing data to JSONL file for analysis
# Format: {"ts":"ISO8601","hook":"name","duration_ms":N,"exit_code":N}
#
# OPTIMIZATION: Opt-in timing via HOOK_TIMING_ENABLE environment variable
# Set HOOK_TIMING_ENABLE=1 to collect timing data for analysis
# Default (unset or not "1"): No timing overhead, immediate return
_timing_log_jsonl() {
    # Opt-in guard: skip all timing work unless explicitly enabled
    # This eliminates I/O overhead in production (default = disabled)
    [[ "${HOOK_TIMING_ENABLE:-}" != "1" ]] && return 0

    local exit_code="${1:-0}"
    local duration_ms="${2:-}"

    # Skip if no duration captured
    [[ -z "$duration_ms" ]] && return 0

    local timing_file="${HOOK_TIMING_FILE:-$HOME/.claude/hook-timing.jsonl}"
    local timing_dir
    timing_dir=$(dirname "$timing_file")

    # Ensure directory exists
    mkdir -p "$timing_dir" 2>/dev/null || return 0

    # Build JSON line
    local ts
    ts=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local json_line="{\"ts\":\"$ts\",\"hook\":\"${HOOK_NAME:-unknown}\",\"duration_ms\":$duration_ms,\"exit_code\":$exit_code}"

    # Append to file
    echo "$json_line" >> "$timing_file" 2>/dev/null || true

    # Rolling retention: keep last 1000 entries
    if [[ -f "$timing_file" ]]; then
        local line_count
        line_count=$(wc -l < "$timing_file" 2>/dev/null | tr -d ' ')
        if [[ "$line_count" -gt 1100 ]]; then
            local temp_file
            temp_file=$(mktemp)
            tail -1000 "$timing_file" > "$temp_file" 2>/dev/null && mv "$temp_file" "$timing_file" 2>/dev/null || rm -f "$temp_file"
        fi
    fi
}
