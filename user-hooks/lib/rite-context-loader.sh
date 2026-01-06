#!/bin/bash
# Rite Context Loader - discovers and executes rite-specific context injection
# Part of Per-Rite Hook Context Injection feature
#
# Usage:
#   source "$HOOKS_LIB/rite-context-loader.sh"
#   output=$(load_rite_context)
#   [[ -n "$output" ]] && echo "$output"

# =============================================================================
# Configuration
# =============================================================================

# Rite context script name (convention)
if [[ -z "${RITE_CONTEXT_SCRIPT_NAME:-}" ]]; then
    readonly RITE_CONTEXT_SCRIPT_NAME="context-injection.sh"
fi

# Function name rites must export
if [[ -z "${RITE_CONTEXT_FUNCTION_NAME:-}" ]]; then
    readonly RITE_CONTEXT_FUNCTION_NAME="inject_team_context"
fi

# =============================================================================
# Main Function
# =============================================================================

# Load rite-specific context if available
# Arguments: None (uses ACTIVE_RITE file and ROSTER_HOME)
# Output: Markdown content to stdout (may be empty)
# Returns: 0 always (errors logged, not propagated)
#
# Contract:
#   - Reads ACTIVE_RITE from .claude/ACTIVE_RITE
#   - Looks for $ROSTER_HOME/rites/$ACTIVE_RITE/context-injection.sh
#   - Sources script and calls inject_team_context()
#   - Returns function output on stdout
#   - Never fails (RECOVERABLE pattern)

load_rite_context() {
    local active_rite
    local rite_script
    local output=""

    # Read active rite (with backward compatibility fallback to ACTIVE_TEAM)
    active_rite=$(cat ".claude/ACTIVE_RITE" 2>/dev/null || cat ".claude/ACTIVE_TEAM" 2>/dev/null || echo "")
    if [[ -z "$active_rite" || "$active_rite" == "none" ]]; then
        # No rite active - nothing to inject
        return 0
    fi

    # Resolve rite context script path
    local roster_home="${ROSTER_HOME:-$HOME/Code/roster}"
    rite_script="$roster_home/rites/$active_rite/$RITE_CONTEXT_SCRIPT_NAME"

    # Check if rite has context script
    if [[ ! -f "$rite_script" ]]; then
        # Rite has no context script - normal, not an error
        log_debug "Rite $active_rite has no context script at $rite_script" 2>/dev/null || true
        return 0
    fi

    # Check if script is executable (warning if not)
    if [[ ! -x "$rite_script" ]]; then
        log_warning "Rite context script exists but not executable: $rite_script" 2>/dev/null || true
        # Try sourcing anyway - bash doesn't require +x for sourcing
    fi

    # Source the rite script
    # Use subshell to isolate any side effects
    output=$(
        # Source rite script
        source "$rite_script" 2>/dev/null || {
            log_warning "Failed to source rite context script: $rite_script" 2>/dev/null || true
            exit 0
        }

        # Check if function exists
        if ! declare -f "$RITE_CONTEXT_FUNCTION_NAME" >/dev/null 2>&1; then
            log_warning "Rite context script missing function: $RITE_CONTEXT_FUNCTION_NAME" 2>/dev/null || true
            exit 0
        fi

        # Call the function
        "$RITE_CONTEXT_FUNCTION_NAME" 2>/dev/null || {
            log_warning "$RITE_CONTEXT_FUNCTION_NAME returned non-zero" 2>/dev/null || true
        }
    )

    # Output result (may be empty)
    echo "$output"
    return 0
}

# =============================================================================
# Utility Functions for Rite Scripts
# =============================================================================

# Rites can use these helpers in their context-injection.sh

# Format a key-value pair for rite context table
# Usage: rite_context_row "Key" "Value"
rite_context_row() {
    local key="$1"
    local value="$2"
    echo "| **$key** | $value |"
}

# Check if a file is newer than N minutes
# Usage: is_file_stale "/path/to/file" 60  # true if older than 60 minutes
is_file_stale() {
    local file="$1"
    local max_age_minutes="${2:-60}"

    if [[ ! -f "$file" ]]; then
        return 0  # Non-existent = stale
    fi

    local now file_time age_seconds max_age_seconds
    now=$(date +%s)
    file_time=$(stat -f %m "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || echo 0)
    age_seconds=$((now - file_time))
    max_age_seconds=$((max_age_minutes * 60))

    [[ $age_seconds -gt $max_age_seconds ]]
}

# =============================================================================
# Backward Compatibility Aliases
# =============================================================================

# Backward compatibility alias
load_team_context() { load_rite_context "$@"; }
team_context_row() { rite_context_row "$@"; }
