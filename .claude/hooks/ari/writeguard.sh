#!/bin/bash
# writeguard.sh - Smart dispatch wrapper for session write guard
# Thin wrapper for ari hook writeguard
# Event: PreToolUse (Edit|Write)
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)

# Check 1: Only process Write or Edit tools
[[ "$CLAUDE_HOOK_TOOL_NAME" != "Write" && "$CLAUDE_HOOK_TOOL_NAME" != "Edit" ]] && exit 0

# Check 2: Read tool input to check file path
TOOL_INPUT="${CLAUDE_HOOK_TOOL_INPUT:-}"

# Check 3: Skip if file path doesn't contain _CONTEXT.md
# Extract file_path from JSON input (fast grep, no jq dependency)
if [[ "$TOOL_INPUT" != *"_CONTEXT.md"* ]]; then
    exit 0
fi

# Source fail-open logging (ADR-0010)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/fail-open.sh" 2>/dev/null || true

# =============================================================================
# STATE-MATE BYPASS DETECTION (per ADR-0010 Section 5)
# =============================================================================
# Priority 1: Agent name detection (future Claude Code enhancement)
# Priority 2: Environment marker (current mechanism)

# Extract file_path for logging (basic extraction from JSON)
_WRITEGUARD_FILE_PATH=""
if command -v jq &>/dev/null; then
    _WRITEGUARD_FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.file_path // empty' 2>/dev/null || echo "")
fi

# Priority 1: Check if invoked by state-mate agent via CLAUDE_TASK_AGENT_NAME
AGENT_NAME="${CLAUDE_TASK_AGENT_NAME:-}"
if [[ "$AGENT_NAME" == "state-mate" ]]; then
    BYPASS_CONTEXT=$(build_fail_open_context "agent_name" "$AGENT_NAME" "tool_name" "$CLAUDE_HOOK_TOOL_NAME" 2>/dev/null || echo '{}')
    log_bypass "writeguard.sh" "agent_name" "$_WRITEGUARD_FILE_PATH" "$BYPASS_CONTEXT" 2>/dev/null || true
    exit 0  # Allow write - state-mate is authorized
fi

# Priority 2: Environment marker (current/fallback mechanism)
if [[ "${STATE_MATE_BYPASS:-}" == "true" ]]; then
    BYPASS_CONTEXT=$(build_fail_open_context "env_var" "STATE_MATE_BYPASS" "tool_name" "$CLAUDE_HOOK_TOOL_NAME" 2>/dev/null || echo '{}')
    log_bypass "writeguard.sh" "env_var" "$_WRITEGUARD_FILE_PATH" "$BYPASS_CONTEXT" 2>/dev/null || true
    exit 0  # Allow write
fi

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

# DISPATCH: Call ari (<100ms total)
ARI=$(get_ari_path 2>/dev/null || echo "${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}")

# Fail-open: If ari unavailable, log and allow operation to proceed
if ! [[ -x "$ARI" ]]; then
    CONTEXT=$(build_fail_open_context "tool_name" "$CLAUDE_HOOK_TOOL_NAME" "ari_path" "$ARI")
    log_fail_open "writeguard.sh" "$CLAUDE_HOOK_TOOL_NAME" "ari binary not found or not executable" "$CONTEXT"
    exit 0
fi

exec "$ARI" hook writeguard --output json
