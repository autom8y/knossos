#!/bin/bash
# context.sh - Smart dispatch wrapper for session context injection
# Thin wrapper for ari hook context
# Event: SessionStart
set -euo pipefail

# FAST PATH: No early exit checks needed - always run on session start
# Context injection is unconditional for startup/resume events

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

# Source fail-open logging (ADR-0010)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/fail-open.sh" 2>/dev/null || true

# DISPATCH: Call ari (<100ms total)
ARI=$(get_ari_path 2>/dev/null || echo "${ARIADNE_BIN:-/Users/tomtenuta/Code/roster/ariadne/ari}")

# Fail-open: If ari unavailable, log and allow operation to proceed
if ! [[ -x "$ARI" ]]; then
    CONTEXT=$(build_fail_open_context "event" "SessionStart" "ari_path" "$ARI")
    log_fail_open "context.sh" "SessionStart" "ari binary not found or not executable" "$CONTEXT"
    exit 0
fi

exec "$ARI" hook context --output json
