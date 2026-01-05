#!/usr/bin/env bash
# fail-open.sh - Fail-open audit logging for hooks
#
# Per ADR-0010: When hooks cannot invoke the `ari` binary (e.g., binary not built,
# PATH issues), they must fail-open to avoid blocking Claude Code's workflow while
# maintaining auditability.
#
# Usage:
#   source "$(dirname "$0")/../lib/fail-open.sh"
#
#   if ! command -v "$ARI" &>/dev/null; then
#       log_fail_open "hook-name.sh" "Edit" "ari binary not found" \
#           '{"file_path": "/path/to/file.md"}'
#       exit 0  # Allow operation to proceed
#   fi

# Determine project root - prefer CLAUDE_PROJECT_DIR, fallback to script location detection
_FAIL_OPEN_PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-}"
if [[ -z "$_FAIL_OPEN_PROJECT_ROOT" ]]; then
    # Try to find project root from script location
    _FAIL_OPEN_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    _FAIL_OPEN_PROJECT_ROOT="${_FAIL_OPEN_SCRIPT_DIR%/.claude/hooks/lib}"
fi

AUDIT_DIR="${_FAIL_OPEN_PROJECT_ROOT}/.claude/audit"
FAIL_OPEN_LOG="$AUDIT_DIR/fail-open.jsonl"

# Ensure audit directory exists
# Creates .claude/audit/ if missing
ensure_audit_dir() {
    mkdir -p "$AUDIT_DIR" 2>/dev/null || true
}

# Log a fail-open event to the audit trail
#
# Arguments:
#   $1 - hook: Hook script name (e.g., "context.sh", "writeguard.sh")
#   $2 - operation: Tool operation that triggered hook (e.g., "Edit", "Write", "SessionStart")
#   $3 - error: Error message explaining why fail-open occurred
#   $4 - context: (optional) JSON object with additional context
#
# Schema (per ADR-0010):
# {
#   "timestamp": "2026-01-05T10:00:00Z",
#   "hook": "session-write-guard.sh",
#   "operation": "Edit",
#   "error": "ari binary not found",
#   "context": {"file_path": "...", "tool_name": "Edit"}
# }
#
# Returns: 0 always (fail-open logging must never crash)
log_fail_open() {
    local hook="${1:-unknown}"
    local operation="${2:-unknown}"
    local error="${3:-unknown error}"
    local context="$4"
    [[ -z "$context" ]] && context='{}'

    # Ensure audit directory exists
    ensure_audit_dir

    # Generate ISO 8601 timestamp
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Validate context is valid JSON, fallback to empty object
    # Note: Using printf to avoid echo interpretation issues
    if printf '%s' "$context" | jq -e . >/dev/null 2>&1; then
        : # Valid JSON, keep as-is
    else
        context='{}'
    fi

    # Create JSON line using jq for proper escaping
    local json_line
    json_line=$(jq -c -n \
        --arg ts "$timestamp" \
        --arg hook "$hook" \
        --arg op "$operation" \
        --arg err "$error" \
        --argjson ctx "$context" \
        '{timestamp: $ts, hook: $hook, operation: $op, error: $err, context: $ctx}' 2>/dev/null)

    # Fallback if jq fails (shouldn't happen, but fail-open on our own logging too)
    if [[ -z "$json_line" ]]; then
        # Manual JSON construction with basic escaping
        local escaped_error
        escaped_error=$(printf '%s' "$error" | sed 's/"/\\"/g; s/\\/\\\\/g')
        json_line="{\"timestamp\":\"$timestamp\",\"hook\":\"$hook\",\"operation\":\"$operation\",\"error\":\"$escaped_error\",\"context\":$context}"
    fi

    # Append to audit log
    echo "$json_line" >> "$FAIL_OPEN_LOG" 2>/dev/null || true

    return 0
}

# Check if ari binary is available
# Usage: if ! ari_available; then log_fail_open ...; exit 0; fi
#
# Checks the ARIADNE_BIN path or default location
# Returns: 0 if available, 1 if not
ari_available() {
    local ari="${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}"
    [[ -x "$ari" ]] && command -v "$ari" &>/dev/null
}

# Get the ari binary path
# Usage: ARI=$(get_ari_path)
get_ari_path() {
    echo "${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}"
}

# Helper: Build context JSON for common hook scenarios
# Usage: context=$(build_fail_open_context "file_path" "/path/to/file" "tool_name" "Edit")
#
# Arguments: key1 value1 key2 value2 ...
# Returns: JSON object string
build_fail_open_context() {
    local json="{}"
    while [[ $# -ge 2 ]]; do
        local key="$1"
        local value="$2"
        shift 2
        json=$(echo "$json" | jq -c --arg k "$key" --arg v "$value" '. + {($k): $v}' 2>/dev/null || echo "$json")
    done
    echo "$json"
}

# Log state-mate bypass events to audit trail
#
# Per ADR-0010 Section 5: When state-mate agent is allowed to bypass
# the session-write-guard, we log the bypass for audit purposes.
#
# Arguments:
#   $1 - hook: Hook script name (e.g., "session-write-guard.sh")
#   $2 - bypass_mechanism: How bypass was detected ("agent_name" or "env_var")
#   $3 - file_path: Path to the file being written
#   $4 - context: (optional) JSON object with additional context
#
# Schema:
# {
#   "timestamp": "2026-01-05T10:00:00Z",
#   "event": "state-mate-bypass",
#   "hook": "session-write-guard.sh",
#   "bypass_mechanism": "agent_name",
#   "file_path": ".claude/sessions/session-xxx/SESSION_CONTEXT.md",
#   "context": {"agent_name": "state-mate", "tool_name": "Edit"}
# }
#
# Returns: 0 always (logging must never crash)
log_bypass() {
    local hook="${1:-unknown}"
    local bypass_mechanism="${2:-unknown}"
    local file_path="${3:-unknown}"
    local context="$4"
    [[ -z "$context" ]] && context='{}'

    # Ensure audit directory exists
    ensure_audit_dir

    # Generate ISO 8601 timestamp
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Validate context is valid JSON, fallback to empty object
    if printf '%s' "$context" | jq -e . >/dev/null 2>&1; then
        : # Valid JSON, keep as-is
    else
        context='{}'
    fi

    # Create JSON line using jq for proper escaping
    local json_line
    json_line=$(jq -c -n \
        --arg ts "$timestamp" \
        --arg hook "$hook" \
        --arg mechanism "$bypass_mechanism" \
        --arg path "$file_path" \
        --argjson ctx "$context" \
        '{timestamp: $ts, event: "state-mate-bypass", hook: $hook, bypass_mechanism: $mechanism, file_path: $path, context: $ctx}' 2>/dev/null)

    # Fallback if jq fails
    if [[ -z "$json_line" ]]; then
        local escaped_path
        escaped_path=$(printf '%s' "$file_path" | sed 's/"/\\"/g; s/\\/\\\\/g')
        json_line="{\"timestamp\":\"$timestamp\",\"event\":\"state-mate-bypass\",\"hook\":\"$hook\",\"bypass_mechanism\":\"$bypass_mechanism\",\"file_path\":\"$escaped_path\",\"context\":$context}"
    fi

    # Append to state-mate bypass audit log
    echo "$json_line" >> "$AUDIT_DIR/state-mate-bypass.jsonl" 2>/dev/null || true

    return 0
}
