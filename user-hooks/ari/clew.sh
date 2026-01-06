#!/bin/bash
# clew.sh - Smart dispatch wrapper for clew/artifact tracking
# Thin wrapper for ari hook clew
# Event: PostToolUse (Edit|Write|Bash)
# Category: RECOVERABLE - graceful degradation if ari binary unavailable
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
exec "$ARI" hook clew --output json
