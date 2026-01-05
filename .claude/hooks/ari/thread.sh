#!/bin/bash
# thread.sh - Smart dispatch wrapper for thread/artifact tracking
# Thin wrapper for ari hook thread
# Event: PostToolUse (Edit|Write|Bash)
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)
# Check 1: Only process Edit, Write, or Bash tools
case "$CLAUDE_HOOK_TOOL_NAME" in Edit|Write|Bash) ;; *) exit 0 ;; esac

# Check 2: Skip if no active session (fast directory check)
SESSION_DIR="${CLAUDE_SESSION_DIR:-.claude/sessions}"
[[ ! -d "$SESSION_DIR" ]] && exit 0

# Check 3: Any active session context file? (fast glob)
shopt -s nullglob; SESSION_FILES=("$SESSION_DIR"/*/SESSION_CONTEXT.md); shopt -u nullglob
[[ ${#SESSION_FILES[@]} -eq 0 ]] && exit 0

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

# Source fail-open logging (ADR-0010)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/fail-open.sh" 2>/dev/null || true

# DISPATCH: Call ari (<100ms total)
ARI=$(get_ari_path 2>/dev/null || echo "${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}")

# Fail-open: If ari unavailable, log and allow operation to proceed
if ! [[ -x "$ARI" ]]; then
    CONTEXT=$(build_fail_open_context "event" "PostToolUse" "tool_name" "$CLAUDE_HOOK_TOOL_NAME" "ari_path" "$ARI")
    log_fail_open "thread.sh" "$CLAUDE_HOOK_TOOL_NAME" "ari binary not found or not executable" "$CONTEXT"
    exit 0
fi

exec "$ARI" hook thread --output json
