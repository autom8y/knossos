#!/bin/bash
# Centralized logging library for hooks
# Provides standard logging functions with timestamp, hook name, and exit code
# Logs to .claude/logs/hooks.log with automatic rotation
#
# Part of Ecosystem v2 refactoring

# NOTE: Libraries should not impose shell options on sourcing scripts
# Hooks that need fail-fast can set their own 'set -euo pipefail'

# =============================================================================
# Configuration
# =============================================================================

# Source configuration (provides HOOKS_LOG_DIR, HOOKS_LOG_FILE, etc.)
source "$(dirname "${BASH_SOURCE[0]}")/config.sh"

# Get calling hook name from first argument or BASH_SOURCE
HOOK_NAME="${1:-${BASH_SOURCE[1]:-unknown}}"
HOOK_NAME="${HOOK_NAME##*/}"
HOOK_NAME="${HOOK_NAME%.sh}"

# =============================================================================
# Logging Functions
# =============================================================================

# Initialize logging (call at start of hook)
# Usage: source logging.sh && log_init "hook-name"
log_init() {
    local hook_name="${1:-$HOOK_NAME}"
    HOOK_NAME="$hook_name"

    # Ensure log directory exists
    mkdir -p "$HOOKS_LOG_DIR" 2>/dev/null || true

    # Rotate logs if needed (on init only, not every log call)
    _log_rotate
}

# Log a message with timestamp and hook name
# Usage: log_hook "message" [level]
# Levels: INFO (default), WARN, ERROR, DEBUG
log_hook() {
    local message="$1"
    local level="${2:-INFO}"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Format: timestamp | level | hook_name | message
    local log_line="$timestamp | $level | $HOOK_NAME | $message"

    # Append to log file (create if needed)
    echo "$log_line" >> "$HOOKS_LOG_FILE" 2>/dev/null || true
}

# Log hook start
log_start() {
    log_hook "Hook started" "INFO"
}

# Log hook completion with exit code
log_end() {
    local exit_code="${1:-0}"
    local duration="${2:-}"

    if [[ -n "$duration" ]]; then
        log_hook "Hook completed (exit=$exit_code, duration=${duration}ms)" "INFO"
    else
        log_hook "Hook completed (exit=$exit_code)" "INFO"
    fi
}

# Log error
log_error() {
    local message="$1"
    log_hook "$message" "ERROR"
}

# Log warning
log_warn() {
    local message="$1"
    log_hook "$message" "WARN"
}

# Log debug (only if CLAUDE_HOOK_DEBUG is set)
log_debug() {
    if [[ "${CLAUDE_HOOK_DEBUG:-0}" == "1" ]]; then
        local message="$1"
        log_hook "$message" "DEBUG"
    fi
}

# =============================================================================
# Log Rotation
# =============================================================================

# Rotate logs: remove entries older than HOOKS_LOG_MAX_AGE_DAYS
# Also truncate if file exceeds HOOKS_LOG_MAX_SIZE_MB
_log_rotate() {
    # Skip if log file doesn't exist
    [[ -f "$HOOKS_LOG_FILE" ]] || return 0

    # Check file size (in MB)
    local size_mb
    if [[ "$(uname)" == "Darwin" ]]; then
        size_mb=$(stat -f%z "$HOOKS_LOG_FILE" 2>/dev/null | awk '{print int($1/1024/1024)}')
    else
        size_mb=$(stat -c%s "$HOOKS_LOG_FILE" 2>/dev/null | awk '{print int($1/1024/1024)}')
    fi

    # If file is too large, keep only last 1000 lines
    if [[ "$size_mb" -gt "$HOOKS_LOG_MAX_SIZE_MB" ]]; then
        local temp_file
        temp_file=$(mktemp)
        tail -1000 "$HOOKS_LOG_FILE" > "$temp_file" 2>/dev/null && mv "$temp_file" "$HOOKS_LOG_FILE"
        log_hook "Log rotated (size exceeded ${HOOKS_LOG_MAX_SIZE_MB}MB)" "INFO"
    fi

    # Remove entries older than HOOKS_LOG_MAX_AGE_DAYS
    # This is expensive, so only do it if file is reasonably large (>100KB)
    local size_kb
    if [[ "$(uname)" == "Darwin" ]]; then
        size_kb=$(stat -f%z "$HOOKS_LOG_FILE" 2>/dev/null | awk '{print int($1/1024)}')
    else
        size_kb=$(stat -c%s "$HOOKS_LOG_FILE" 2>/dev/null | awk '{print int($1/1024)}')
    fi

    if [[ "$size_kb" -gt 100 ]]; then
        local cutoff_date
        if [[ "$(uname)" == "Darwin" ]]; then
            cutoff_date=$(date -v-${HOOKS_LOG_MAX_AGE_DAYS}d -u +"%Y-%m-%dT%H:%M:%SZ")
        else
            cutoff_date=$(date -u -d "${HOOKS_LOG_MAX_AGE_DAYS} days ago" +"%Y-%m-%dT%H:%M:%SZ")
        fi

        # Filter to keep only recent entries (crude but effective)
        local temp_file
        temp_file=$(mktemp)
        awk -v cutoff="$cutoff_date" '$1 >= cutoff' "$HOOKS_LOG_FILE" > "$temp_file" 2>/dev/null && \
            mv "$temp_file" "$HOOKS_LOG_FILE" || rm -f "$temp_file"
    fi
}

# =============================================================================
# Timing Helpers
# =============================================================================

# Get current time in milliseconds (for duration tracking)
get_time_ms() {
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS: use perl for milliseconds
        perl -MTime::HiRes=time -e 'printf "%.0f\n", time * 1000' 2>/dev/null || date +%s000
    else
        # Linux: use date with nanoseconds
        date +%s%3N 2>/dev/null || date +%s000
    fi
}

# Calculate duration in milliseconds
calc_duration_ms() {
    local start_ms="$1"
    local end_ms
    end_ms=$(get_time_ms)
    echo $((end_ms - start_ms))
}

# =============================================================================
# Timeout Protection
# =============================================================================

# Note: HOOK_TIMEOUT is defined in config.sh

# Run a command with timeout protection
# Usage: run_with_timeout <command> [timeout_seconds]
# Returns: 0 on success, 124 on timeout, other on command failure
run_with_timeout() {
    local cmd="$1"
    local timeout="${2:-$HOOK_TIMEOUT}"

    # Use timeout command if available (GNU coreutils)
    if command -v timeout &>/dev/null; then
        timeout --signal=TERM "$timeout" bash -c "$cmd"
        local result=$?
        if [[ $result -eq 124 ]]; then
            log_error "Command timed out after ${timeout}s: $cmd"
        fi
        return $result
    fi

    # macOS fallback using perl (more reliable than background process)
    if command -v perl &>/dev/null; then
        perl -e '
            use strict;
            use warnings;
            my $timeout = $ARGV[0];
            my $cmd = $ARGV[1];

            eval {
                local $SIG{ALRM} = sub { die "timeout\n" };
                alarm($timeout);
                system($cmd);
                alarm(0);
            };
            if ($@ && $@ eq "timeout\n") {
                exit(124);
            }
            exit($? >> 8);
        ' "$timeout" "$cmd"
        local result=$?
        if [[ $result -eq 124 ]]; then
            log_error "Command timed out after ${timeout}s: $cmd"
        fi
        return $result
    fi

    # Last resort: just run without timeout
    log_warn "No timeout mechanism available, running without timeout"
    bash -c "$cmd"
}

# Wrapper to run a hook with timeout
# Usage: hook_with_timeout <hook_script> [timeout_seconds]
hook_with_timeout() {
    local hook_script="$1"
    local timeout="${2:-$HOOK_TIMEOUT}"
    local hook_name
    hook_name=$(basename "$hook_script" .sh)

    local start_time
    start_time=$(get_time_ms 2>/dev/null || echo 0)

    # Run hook with timeout
    local result=0
    if command -v timeout &>/dev/null; then
        timeout --signal=TERM "$timeout" "$hook_script"
        result=$?
    else
        # macOS: use background job with manual timeout
        "$hook_script" &
        local pid=$!

        # Wait with timeout
        local elapsed=0
        while kill -0 "$pid" 2>/dev/null && [[ $elapsed -lt $timeout ]]; do
            sleep 0.1
            elapsed=$(( elapsed + 1 ))  # Approximate, each iteration ~0.1s
            [[ $elapsed -ge $(( timeout * 10 )) ]] && break
        done

        if kill -0 "$pid" 2>/dev/null; then
            # Still running after timeout - kill it
            kill -TERM "$pid" 2>/dev/null
            sleep 0.1
            kill -KILL "$pid" 2>/dev/null
            wait "$pid" 2>/dev/null
            log_error "Hook '$hook_name' timed out after ${timeout}s"
            result=124
        else
            wait "$pid"
            result=$?
        fi
    fi

    # Log completion
    local duration=""
    if [[ "$start_time" != "0" ]]; then
        duration=$(calc_duration_ms "$start_time" 2>/dev/null || echo "")
    fi

    if [[ $result -eq 124 ]]; then
        log_hook "Hook timed out (${timeout}s)" "ERROR"
    elif [[ $result -ne 0 ]]; then
        log_hook "Hook failed (exit=$result)" "ERROR"
    else
        log_hook "Hook completed${duration:+ (${duration}ms)}" "INFO"
    fi

    return $result
}

# =============================================================================
# Export Functions
# =============================================================================

export -f log_init log_hook log_start log_end log_error log_warn log_debug
export -f get_time_ms calc_duration_ms
export -f run_with_timeout hook_with_timeout
export HOOK_NAME HOOKS_LOG_FILE HOOKS_LOG_DIR HOOK_TIMEOUT
