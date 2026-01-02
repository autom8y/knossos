#!/bin/bash
# Team Context Loader - discovers and executes team-specific context injection
# Part of Per-Team Hook Context Injection feature
#
# Usage:
#   source "$HOOKS_LIB/team-context-loader.sh"
#   output=$(load_team_context)
#   [[ -n "$output" ]] && echo "$output"

# =============================================================================
# Configuration
# =============================================================================

# Team context script name (convention)
if [[ -z "${TEAM_CONTEXT_SCRIPT_NAME:-}" ]]; then
    readonly TEAM_CONTEXT_SCRIPT_NAME="context-injection.sh"
fi

# Function name teams must export
if [[ -z "${TEAM_CONTEXT_FUNCTION_NAME:-}" ]]; then
    readonly TEAM_CONTEXT_FUNCTION_NAME="inject_team_context"
fi

# =============================================================================
# Main Function
# =============================================================================

# Load team-specific context if available
# Arguments: None (uses ACTIVE_TEAM file and ROSTER_HOME)
# Output: Markdown content to stdout (may be empty)
# Returns: 0 always (errors logged, not propagated)
#
# Contract:
#   - Reads ACTIVE_TEAM from .claude/ACTIVE_TEAM
#   - Looks for $ROSTER_HOME/teams/$ACTIVE_TEAM/context-injection.sh
#   - Sources script and calls inject_team_context()
#   - Returns function output on stdout
#   - Never fails (RECOVERABLE pattern)

load_team_context() {
    local active_team
    local team_script
    local output=""

    # Read active team
    active_team=$(cat ".claude/ACTIVE_TEAM" 2>/dev/null || echo "")
    if [[ -z "$active_team" || "$active_team" == "none" ]]; then
        # No team active - nothing to inject
        return 0
    fi

    # Resolve team context script path
    local roster_home="${ROSTER_HOME:-$HOME/Code/roster}"
    team_script="$roster_home/teams/$active_team/$TEAM_CONTEXT_SCRIPT_NAME"

    # Check if team has context script
    if [[ ! -f "$team_script" ]]; then
        # Team has no context script - normal, not an error
        log_debug "Team $active_team has no context script at $team_script" 2>/dev/null || true
        return 0
    fi

    # Check if script is executable (warning if not)
    if [[ ! -x "$team_script" ]]; then
        log_warning "Team context script exists but not executable: $team_script" 2>/dev/null || true
        # Try sourcing anyway - bash doesn't require +x for sourcing
    fi

    # Source the team script
    # Use subshell to isolate any side effects
    output=$(
        # Source team script
        source "$team_script" 2>/dev/null || {
            log_warning "Failed to source team context script: $team_script" 2>/dev/null || true
            exit 0
        }

        # Check if function exists
        if ! declare -f "$TEAM_CONTEXT_FUNCTION_NAME" >/dev/null 2>&1; then
            log_warning "Team context script missing function: $TEAM_CONTEXT_FUNCTION_NAME" 2>/dev/null || true
            exit 0
        fi

        # Call the function
        "$TEAM_CONTEXT_FUNCTION_NAME" 2>/dev/null || {
            log_warning "$TEAM_CONTEXT_FUNCTION_NAME returned non-zero" 2>/dev/null || true
        }
    )

    # Output result (may be empty)
    echo "$output"
    return 0
}

# =============================================================================
# Utility Functions for Team Scripts
# =============================================================================

# Teams can use these helpers in their context-injection.sh

# Format a key-value pair for team context table
# Usage: team_context_row "Key" "Value"
team_context_row() {
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
