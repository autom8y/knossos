#!/bin/bash
# validate.sh - Smart dispatch wrapper for command validation
# Thin wrapper for ari hook validate
# Event: PreToolUse (Bash)
# Category: RECOVERABLE - graceful degradation if ari binary unavailable
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)

# Check 1: Only process Bash tool
[[ "$CLAUDE_HOOK_TOOL_NAME" != "Bash" ]] && exit 0

# Binary resolution with PATH fallback (per ADR-0002 style)
ARI="${ARIADNE_BIN:-}"
if [[ -z "$ARI" ]]; then
    # Priority 1: PATH lookup (for installed binary)
    if command -v ari &>/dev/null; then
        ARI="$(command -v ari)"
    # Priority 2: Project-relative location (for development)
    elif [[ -x "${CLAUDE_PROJECT_DIR:-$PWD}/ariadne/ari" ]]; then
        ARI="${CLAUDE_PROJECT_DIR:-$PWD}/ariadne/ari"
    fi
fi

# Guard: binary must exist and be executable (graceful degradation)
[[ -x "$ARI" ]] || exit 0

# DISPATCH: Call ari (<100ms total)
exec "$ARI" hook validate --output json
