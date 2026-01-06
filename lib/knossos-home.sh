#!/usr/bin/env bash
# knossos-home.sh - Centralized KNOSSOS_HOME resolution
#
# Usage:
#   source "$SCRIPT_DIR/../lib/knossos-home.sh"
#   # KNOSSOS_HOME is now set and exported
#
# Environment Variables:
#   KNOSSOS_HOME - Platform home directory
#   Default: $HOME/Code/roster
#
# Part of the Knossos Platform

# Version for tracking
readonly KNOSSOS_HOME_RESOLVER_VERSION="2.0.0"

# resolve_knossos_home - Resolve and export KNOSSOS_HOME
# Idempotent: safe to call multiple times
resolve_knossos_home() {
    # Already resolved
    if [[ -n "${_KNOSSOS_HOME_RESOLVED:-}" ]]; then
        return 0
    fi

    if [[ -n "${KNOSSOS_HOME:-}" ]]; then
        # KNOSSOS_HOME is set
        export KNOSSOS_HOME
    else
        # Default
        export KNOSSOS_HOME="$HOME/Code/roster"
    fi

    # Mark as resolved
    export _KNOSSOS_HOME_RESOLVED=1
}

# Auto-resolve on source (can be disabled with KNOSSOS_HOME_NO_AUTO_RESOLVE=1)
if [[ "${KNOSSOS_HOME_NO_AUTO_RESOLVE:-0}" != "1" ]]; then
    resolve_knossos_home
fi
