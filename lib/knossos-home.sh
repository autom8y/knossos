#!/usr/bin/env bash
# knossos-home.sh - Centralized KNOSSOS_HOME resolution with deprecation handling
#
# Usage:
#   source "$SCRIPT_DIR/../lib/knossos-home.sh"
#   resolve_knossos_home
#   # KNOSSOS_HOME is now set and exported
#
# Environment Variables:
#   KNOSSOS_HOME - Primary platform home (preferred)
#   ROSTER_HOME  - Deprecated fallback (with warning)
#   Default: $HOME/Code/roster
#
# Part of the Knossos Platform (roster → knossos migration)

# Version for migration tracking
readonly KNOSSOS_HOME_RESOLVER_VERSION="1.0.0"

# Deprecation warning control
# Set KNOSSOS_SUPPRESS_DEPRECATION=1 to silence warnings (for tests)
KNOSSOS_SUPPRESS_DEPRECATION="${KNOSSOS_SUPPRESS_DEPRECATION:-0}"

# resolve_knossos_home - Resolve and export KNOSSOS_HOME
# Idempotent: safe to call multiple times
resolve_knossos_home() {
    # Already resolved
    if [[ -n "${_KNOSSOS_HOME_RESOLVED:-}" ]]; then
        return 0
    fi

    if [[ -n "${KNOSSOS_HOME:-}" ]]; then
        # Primary: KNOSSOS_HOME is set
        export KNOSSOS_HOME
    elif [[ -n "${ROSTER_HOME:-}" ]]; then
        # Fallback: ROSTER_HOME is set (deprecated)
        export KNOSSOS_HOME="$ROSTER_HOME"
        if [[ "$KNOSSOS_SUPPRESS_DEPRECATION" != "1" ]]; then
            {
                echo "[DEPRECATED] Environment variable ROSTER_HOME is deprecated."
                echo "  Update your shell profile (~/.bashrc, ~/.zshrc, etc.):"
                echo "  - Remove: export ROSTER_HOME=\"$ROSTER_HOME\""
                echo "  + Add:    export KNOSSOS_HOME=\"$ROSTER_HOME\""
                echo "  ROSTER_HOME support will be removed in version 3.0"
            } >&2
        fi
    else
        # Default: Neither set
        export KNOSSOS_HOME="$HOME/Code/roster"
    fi

    # Mark as resolved
    export _KNOSSOS_HOME_RESOLVED=1
}

# Auto-resolve on source (can be disabled with KNOSSOS_HOME_NO_AUTO_RESOLVE=1)
if [[ "${KNOSSOS_HOME_NO_AUTO_RESOLVE:-0}" != "1" ]]; then
    resolve_knossos_home
fi
