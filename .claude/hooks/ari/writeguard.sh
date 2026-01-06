#!/bin/bash
# writeguard.sh - Smart dispatch wrapper for session write guard
# Thin wrapper for ari hook writeguard
# Event: PreToolUse (Edit|Write)
# Category: RECOVERABLE - graceful degradation if ari binary unavailable
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

# Check 4: Allow bypass for Moirai operations (backward compatible with STATE_MATE_BYPASS)
[[ "${MOIRAI_BYPASS:-}" == "true" ]] || [[ "${STATE_MATE_BYPASS:-}" == "true" ]] && exit 0

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

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
exec "$ARI" hook writeguard --output json
