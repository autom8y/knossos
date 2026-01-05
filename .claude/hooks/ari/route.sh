#!/bin/bash
# route.sh - Smart dispatch wrapper for slash command routing
# Thin wrapper for ari hook route
# Event: UserPromptSubmit
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)

# Check 1: Get user prompt from environment or stdin
USER_PROMPT="${CLAUDE_HOOK_USER_PROMPT:-}"
if [[ -z "$USER_PROMPT" ]]; then
    # Read from stdin if not in environment
    read -r USER_PROMPT 2>/dev/null || true
fi

# Check 2: Skip if prompt doesn't start with "/"
# Fast string prefix check without spawning subprocess
[[ "${USER_PROMPT:0:1}" != "/" ]] && exit 0

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

# Source fail-open logging (ADR-0010)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/fail-open.sh" 2>/dev/null || true

# DISPATCH: Call ari (<100ms total)
ARI=$(get_ari_path 2>/dev/null || echo "${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}")

# Fail-open: If ari unavailable, log and allow operation to proceed
if ! [[ -x "$ARI" ]]; then
    CONTEXT=$(build_fail_open_context "event" "UserPromptSubmit" "prompt_prefix" "${USER_PROMPT:0:50}" "ari_path" "$ARI")
    log_fail_open "route.sh" "UserPromptSubmit" "ari binary not found or not executable" "$CONTEXT"
    exit 0
fi

exec "$ARI" hook route --output json
