#!/bin/bash
# hooks-init.sh - Unified initialization script for all hooks
# Provides standardized error handling, library sourcing, and logging setup
#
# Part of Sprint 002: Hooks Standardization
# Design: docs/design/TDD-hooks-init.md

# =============================================================================
# Core Initialization Function
# =============================================================================

# Initialize hook with appropriate error handling mode
# Usage: hooks_init <hook_name> <category>
# Categories: DEFENSIVE | RECOVERABLE
#
# Effects:
#   - Sets HOOK_NAME and HOOK_CATEGORY globals
#   - Configures shell options based on category
#   - Initializes logging
#   - Sources required libraries
#   - Sets up error trap (RECOVERABLE only)
#
# Returns: 0 always (errors are logged, not propagated)

hooks_init() {
    local hook_name="${1:-unknown}"
    local category="${2:-RECOVERABLE}"

    # Export for use in error messages
    export HOOK_NAME="$hook_name"
    export HOOK_CATEGORY="$category"

    # Determine library directory
    local lib_dir="$(dirname "${BASH_SOURCE[0]}")"

    # Source dependencies (order matters)
    # Use 2>/dev/null || true pattern to never crash during init
    source "$lib_dir/config.sh" 2>/dev/null || true
    source "$lib_dir/logging.sh" 2>/dev/null || true
    source "$lib_dir/primitives.sh" 2>/dev/null || true

    # Initialize logging
    log_init "$hook_name" 2>/dev/null || true
    log_start 2>/dev/null || true

    # Category-specific setup
    case "$category" in
        DEFENSIVE)
            # Explicitly disable strict modes
            # DEFENSIVE hooks must never crash Claude's tool flow
            set +e +u +o pipefail 2>/dev/null || true
            ;;
        RECOVERABLE)
            # Enable strict modes with recovery trap
            # RECOVERABLE hooks can detect errors but must degrade gracefully
            _hooks_setup_recovery_trap
            set -euo pipefail
            ;;
        *)
            # Unknown category defaults to DEFENSIVE for safety
            log_warn "Unknown hook category: $category, defaulting to DEFENSIVE" 2>/dev/null || true
            set +e +u +o pipefail 2>/dev/null || true
            ;;
    esac

    return 0
}

# =============================================================================
# Error Trap for RECOVERABLE Hooks
# =============================================================================

# Internal: Set up error trap for RECOVERABLE hooks
# Catches any error, logs it, and exits 0 to prevent hook from crashing Claude
_hooks_setup_recovery_trap() {
    trap '_hooks_handle_error $? $LINENO "$BASH_COMMAND"' ERR
}

_hooks_handle_error() {
    local exit_code="$1"
    local line_number="$2"
    local command="$3"

    # Log the error
    log_error "Hook failed at line $line_number: $command (exit $exit_code)" 2>/dev/null || true

    # Log completion with error
    log_end "$exit_code" 2>/dev/null || true

    # Exit 0 to prevent hook from blocking Claude
    exit 0
}

# =============================================================================
# Safe Sourcing for Optional Dependencies
# =============================================================================

# Safely source optional dependency with fallback
# Usage: safe_source <file_path> [fallback_action]
#
# Returns: 0 if sourced successfully, 1 if not (never crashes)

safe_source() {
    local file_path="$1"
    local fallback="${2:-}"

    if [[ -f "$file_path" ]]; then
        source "$file_path" 2>/dev/null
        return $?
    else
        if [[ -n "$fallback" ]]; then
            log_debug "Optional dependency not found: $file_path, using fallback" 2>/dev/null || true
            eval "$fallback" 2>/dev/null || true
        fi
        return 1
    fi
}

# =============================================================================
# Explicit Finalization for DEFENSIVE Hooks
# =============================================================================

# Call at end of hook to log completion
# Usage: hooks_finalize [exit_code]
#
# Note: For DEFENSIVE hooks that explicitly manage their exit
# RECOVERABLE hooks auto-finalize via trap

hooks_finalize() {
    local exit_code="${1:-0}"
    log_end "$exit_code" 2>/dev/null || true
}
