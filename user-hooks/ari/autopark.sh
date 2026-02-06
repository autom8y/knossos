#!/bin/bash
# autopark.sh - Smart dispatch wrapper for auto-park on session stop
# Thin wrapper for ari hook autopark
# Event: Stop
# Category: RECOVERABLE - graceful degradation if ari binary unavailable
set -euo pipefail

# FAST PATH: No early exit checks needed - always run on stop
# Auto-park should attempt to preserve session state on any stop event

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
exec "$ARI" hook autopark --output json
