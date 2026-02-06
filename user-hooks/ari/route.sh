#!/bin/bash
# route.sh - Smart dispatch wrapper for slash command routing
# Thin wrapper for ari hook route
# Event: UserPromptSubmit
# Category: RECOVERABLE - graceful degradation if ari binary unavailable
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
exec "$ARI" hook route --output json
