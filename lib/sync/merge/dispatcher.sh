#!/usr/bin/env bash
#
# dispatcher.sh - Merge Strategy Dispatcher
#
# Routes merge operations to appropriate strategy implementations.
# Provides a unified interface for all merge strategies.
#
# Part of: roster-sync (TDD-cem-replacement)
#
# Usage:
#   source "$ROSTER_HOME/lib/sync/merge/dispatcher.sh"
#   dispatch_merge_strategy "merge-settings" "$knossos_file" "$local_file" "$output_file"
#
# Strategies:
#   copy-replace    - Complete file replacement
#   merge-settings  - JSON union merge for settings
#   merge-docs      - Section-based merge with markers
#   merge-dir       - Directory content sync
#   merge-init      - Copy if missing

# Guard against re-sourcing
[[ -n "${_MERGE_DISPATCHER_LOADED:-}" ]] && return 0
readonly _MERGE_DISPATCHER_LOADED=1

# ============================================================================
# Strategy Dispatcher (per TDD 5.6)
# ============================================================================

# Dispatch to appropriate merge strategy
# Usage: dispatch_merge_strategy "strategy" "knossos_file" "local_file" "output_file"
# Returns: 0 on success, 1 on failure
dispatch_merge_strategy() {
    local strategy="$1"
    local knossos_file="$2"
    local local_file="$3"
    local output_file="$4"

    sync_log_debug "Dispatching merge: strategy=$strategy"
    sync_log_debug "  roster: $knossos_file"
    sync_log_debug "  local:  $local_file"
    sync_log_debug "  output: $output_file"

    # Verify roster file exists
    if [[ ! -f "$knossos_file" ]]; then
        sync_log_error "Roster file not found: $knossos_file"
        return 1
    fi

    case "$strategy" in
        copy-replace)
            # Complete replacement - just copy
            cp "$knossos_file" "$output_file" || {
                sync_log_error "copy-replace failed: $knossos_file -> $output_file"
                return 1
            }
            sync_log_debug "copy-replace complete: $output_file"
            ;;

        merge-settings)
            # JSON union merge for settings.local.json
            merge_settings_json "$knossos_file" "$local_file" "$output_file" || {
                sync_log_error "merge-settings failed"
                return 1
            }
            sync_log_debug "merge-settings complete: $output_file"
            ;;

        merge-docs)
            # Section-based merge for CLAUDE.md
            merge_documentation "$knossos_file" "$local_file" "$output_file" || {
                sync_log_error "merge-docs failed"
                return 1
            }
            sync_log_debug "merge-docs complete: $output_file"
            ;;

        merge-dir)
            # Directory content sync
            merge_directory "$knossos_file" "$local_file" 0 || {
                sync_log_error "merge-dir failed"
                return 1
            }
            sync_log_debug "merge-dir complete: $output_file"
            ;;

        merge-init)
            # Copy only if local doesn't exist
            if [[ ! -f "$local_file" ]]; then
                cp "$knossos_file" "$output_file" || {
                    sync_log_error "merge-init failed: $knossos_file -> $output_file"
                    return 1
                }
                sync_log_debug "merge-init complete (copied): $output_file"
            else
                sync_log_debug "merge-init skipped (exists): $local_file"
            fi
            ;;

        *)
            sync_log_error "Unknown merge strategy: $strategy"
            return 1
            ;;
    esac

    return 0
}

# ============================================================================
# Strategy Validation
# ============================================================================

# Check if a strategy is valid
# Usage: is_valid_strategy "strategy_name"
# Returns: 0 if valid, 1 if not
is_valid_strategy() {
    local strategy="$1"

    case "$strategy" in
        copy-replace|merge-settings|merge-docs|merge-dir|merge-init)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Get all valid strategy names
list_strategies() {
    echo "copy-replace"
    echo "merge-settings"
    echo "merge-docs"
    echo "merge-dir"
    echo "merge-init"
}
