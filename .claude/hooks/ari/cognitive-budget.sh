#!/bin/bash
# cognitive-budget.sh - Track tool use count and warn on cognitive budget thresholds
# Event: PostToolUse (all tools)
# Category: DEFENSIVE - must never block tool execution
#
# Maintains per-CLI-invocation message count and warns when thresholds are breached.
# Count resets automatically when Claude Code session ends (temp file based).
#
# Environment Variables:
#   ARIADNE_MSG_WARN - Warning threshold (default: 250)
#   ARIADNE_MSG_PARK - Park suggestion threshold (default: none/disabled)
#   ARIADNE_BUDGET_DISABLE - Set to 1 to disable budget tracking
#
# Output:
#   Warnings to stderr when thresholds are breached (does not block tool execution)
set -euo pipefail

# =============================================================================
# Fast Path: Early Exit Checks
# =============================================================================

# Feature flag: disable budget tracking entirely
[[ "${ARIADNE_BUDGET_DISABLE:-0}" == "1" ]] && exit 0

# =============================================================================
# Configuration
# =============================================================================

# Default thresholds
# 250 is calibrated for tool-heavy work (~4-5 hours of active development)
DEFAULT_WARN_THRESHOLD=250

# Thresholds from environment (or defaults)
WARN_THRESHOLD="${ARIADNE_MSG_WARN:-$DEFAULT_WARN_THRESHOLD}"
PARK_THRESHOLD="${ARIADNE_MSG_PARK:-}"

# =============================================================================
# State Management
# =============================================================================

# Use a temp file keyed by Claude Code process tree for per-invocation scope
# This ensures count resets when a new Claude Code session starts
#
# Key resolution order:
# 1. ARIADNE_SESSION_KEY - explicit key (for testing)
# 2. CLAUDE_SESSION_ID - Claude-provided session ID
# 3. PPID - parent process ID (Claude Code process)
#
# The temp file lives in /tmp and is automatically cleaned by OS on reboot
get_state_file() {
    local session_key

    if [[ -n "${ARIADNE_SESSION_KEY:-}" ]]; then
        # Explicit key (testing/debugging)
        session_key="$ARIADNE_SESSION_KEY"
    elif [[ -n "${CLAUDE_SESSION_ID:-}" ]]; then
        # Claude-provided session ID (most reliable in production)
        session_key="$CLAUDE_SESSION_ID"
    else
        # Fall back to parent PID (Claude Code process)
        session_key="ppid-${PPID:-unknown}"
    fi

    echo "/tmp/ariadne-msg-count-${session_key}"
}

STATE_FILE="$(get_state_file)"

# =============================================================================
# Count Management
# =============================================================================

# Read current count (0 if file doesn't exist)
read_count() {
    if [[ -f "$STATE_FILE" ]]; then
        cat "$STATE_FILE" 2>/dev/null || echo "0"
    else
        echo "0"
    fi
}

# Increment and write count atomically
increment_count() {
    local current
    current=$(read_count)
    local new_count=$((current + 1))

    # Atomic write using temp file + mv
    local temp_file
    temp_file=$(mktemp)
    echo "$new_count" > "$temp_file"
    mv "$temp_file" "$STATE_FILE" 2>/dev/null || {
        rm -f "$temp_file" 2>/dev/null
        echo "$new_count" > "$STATE_FILE" 2>/dev/null || true
    }

    echo "$new_count"
}

# =============================================================================
# Warning Output
# =============================================================================

# Emit warning to stderr (non-blocking, informational)
emit_warning() {
    local count="$1"
    local threshold="$2"
    local severity="$3"

    case "$severity" in
        warn)
            echo "[cognitive-budget] Warning: Tool use count ($count) reached warning threshold ($threshold). Consider using /park to preserve session state." >&2
            ;;
        park)
            echo "[cognitive-budget] Alert: Tool use count ($count) reached park threshold ($threshold). Recommend /park now to preserve session state and avoid context degradation." >&2
            ;;
    esac
}

# =============================================================================
# Main Logic
# =============================================================================

# Increment the count
CURRENT_COUNT=$(increment_count)

# Check thresholds and emit warnings
# Only warn once per threshold crossing (use marker files)

# Warning threshold check
if [[ -n "$WARN_THRESHOLD" && "$CURRENT_COUNT" -ge "$WARN_THRESHOLD" ]]; then
    WARN_MARKER="${STATE_FILE}.warned"
    if [[ ! -f "$WARN_MARKER" ]]; then
        emit_warning "$CURRENT_COUNT" "$WARN_THRESHOLD" "warn"
        touch "$WARN_MARKER" 2>/dev/null || true
    fi
fi

# Park threshold check (if configured)
if [[ -n "$PARK_THRESHOLD" && "$CURRENT_COUNT" -ge "$PARK_THRESHOLD" ]]; then
    PARK_MARKER="${STATE_FILE}.park-warned"
    if [[ ! -f "$PARK_MARKER" ]]; then
        emit_warning "$CURRENT_COUNT" "$PARK_THRESHOLD" "park"
        touch "$PARK_MARKER" 2>/dev/null || true
    fi
fi

# Success - exit cleanly
exit 0
