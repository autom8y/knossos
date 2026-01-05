#!/bin/bash
# validate.sh - Smart dispatch wrapper for command validation
# Thin wrapper for ari hook validate
# Event: PreToolUse (Bash)
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)

# Check 1: Only process Bash tool
[[ "$CLAUDE_HOOK_TOOL_NAME" != "Bash" ]] && exit 0

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

# Source fail-open logging (ADR-0010)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/fail-open.sh" 2>/dev/null || true

# DISPATCH: Call ari (<100ms total)
ARI=$(get_ari_path 2>/dev/null || echo "${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}")

# Fail-open: If ari unavailable, log and allow operation to proceed
if ! [[ -x "$ARI" ]]; then
    CONTEXT=$(build_fail_open_context "event" "PreToolUse" "tool_name" "Bash" "ari_path" "$ARI")
    log_fail_open "validate.sh" "Bash" "ari binary not found or not executable" "$CONTEXT"
    exit 0
fi

exec "$ARI" hook validate --output json
