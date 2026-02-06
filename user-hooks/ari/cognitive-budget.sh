#!/bin/bash
# cognitive-budget.sh - Track tool use count and warn on cognitive budget thresholds
# Thin wrapper for ari hook budget
# Event: PostToolUse (all tools)
# Category: DEFENSIVE - must never block tool execution
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)
# Check 1: Feature flag (default: enabled)
[[ "${ARIADNE_BUDGET_DISABLE:-0}" == "1" ]] && exit 0

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
exec "$ARI" hook budget --output json
