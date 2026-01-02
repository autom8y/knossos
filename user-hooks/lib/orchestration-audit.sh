#!/bin/bash
# orchestration-audit.sh - Centralized orchestration event logging
#
# Usage: log_orchestration_event <event_type> <details_json> [outcome] [hook_name]
# Events are logged to session's orchestration-audit.jsonl
#
# Part of orchestration mode consolidation (WP4)
# See: docs/ecosystem/CONTEXT-DESIGN-orchestration-mode-consolidation.md

set -euo pipefail

# Get absolute path to this script's directory (required for reliable sourcing)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source session utilities for get_session_dir()
source "$SCRIPT_DIR/session-utils.sh" 2>/dev/null || return 0

# log_orchestration_event - Write event to session audit log
#
# Args:
#   $1 - event_type: DELEGATION_WARNING, BYPASS_WARNING, ORCHESTRATOR_CONSULTED, MODE_TRANSITION
#   $2 - details_json: JSON object with event-specific details
#   $3 - outcome: CONTINUED, SUCCESS, BLOCKED (default: CONTINUED)
#   $4 - hook_name: Hook that generated event (default: unknown)
#
# Returns: 0 (silent failure if no session)
log_orchestration_event() {
    local event_type="$1"
    local details_json="$2"
    local outcome="${3:-CONTINUED}"
    local hook_name="${4:-unknown}"

    # Get session directory - fail silently if no session
    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 0

    local audit_file="$session_dir/orchestration-audit.jsonl"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Append event to audit log (create if needed)
    # Each line is valid JSON (JSONL format)
    cat >> "$audit_file" <<EOF
{"timestamp":"$timestamp","event":"$event_type","hook":"$hook_name","details":$details_json,"outcome":"$outcome"}
EOF
}

# log_delegation_warning - Log Edit/Write in orchestrated mode
#
# Args:
#   $1 - tool: Edit or Write
#   $2 - file_path: Path to file being modified
#   $3 - mode: Execution mode (orchestrated)
#   $4 - complexity: Session complexity level (optional, default: unknown)
#   $5 - enforcement_tier: Enforcement tier (optional, default: warn)
#   $6 - override_active: Override status (optional, default: false)
#   $7 - override_reason: Override reason (optional)
#   $8 - outcome: Operation outcome (optional, default: CONTINUED)
log_delegation_warning() {
    local tool="$1"
    local file_path="$2"
    local mode="$3"
    local complexity="${4:-unknown}"
    local enforcement_tier="${5:-warn}"
    local override_active="${6:-false}"
    local override_reason="${7:-}"
    local outcome="${8:-CONTINUED}"

    local details="{\"tool\":\"$tool\",\"file_path\":\"$file_path\",\"mode\":\"$mode\",\"complexity\":\"$complexity\",\"enforcement_tier\":\"$enforcement_tier\",\"override_active\":$override_active"

    if [[ -n "$override_reason" ]]; then
        details="${details},\"override_reason\":\"$override_reason\""
    fi
    details="${details}}"

    log_orchestration_event "DELEGATION_WARNING" "$details" "$outcome" "delegation-check.sh"
}

# log_bypass_warning - Log specialist invocation without orchestrator
#
# Args:
#   $1 - specialist: Agent name being invoked
#   $2 - complexity: Session complexity level (optional, default: unknown)
#   $3 - enforcement_tier: Enforcement tier (optional, default: warn)
#   $4 - override_active: Override status (optional, default: false)
#   $5 - override_reason: Override reason (optional)
#   $6 - outcome: Operation outcome (optional, default: CONTINUED)
log_bypass_warning() {
    local specialist="$1"
    local complexity="${2:-unknown}"
    local enforcement_tier="${3:-warn}"
    local override_active="${4:-false}"
    local override_reason="${5:-}"
    local outcome="${6:-CONTINUED}"

    local details="{\"specialist\":\"$specialist\",\"complexity\":\"$complexity\",\"enforcement_tier\":\"$enforcement_tier\",\"override_active\":$override_active"

    if [[ -n "$override_reason" ]]; then
        details="${details},\"override_reason\":\"$override_reason\""
    fi
    details="${details}}"

    log_orchestration_event "BYPASS_WARNING" "$details" "$outcome" "orchestrator-bypass-check.sh"
}

# log_orchestrator_consulted - Log successful orchestrator consultation
#
# Args:
#   $1 - request_type: Type of request (CONSULTATION_REQUEST, TASK_DELEGATION, etc)
log_orchestrator_consulted() {
    local request_type="$1"
    log_orchestration_event "ORCHESTRATOR_CONSULTED" \
        "{\"request_type\":\"$request_type\"}" \
        "SUCCESS" "main-thread"
}
