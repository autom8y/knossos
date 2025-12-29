#!/bin/bash
# Session utilities for multi-session support
# Backward compatibility shim - sources decomposed modules
#
# Part of Ecosystem v2 refactoring (RF-006)
#
# ARCHITECTURE NOTE:
# This file now sources session-state.sh which provides the full dependency chain:
#   session-utils.sh -> session-state.sh -> session-core.sh -> primitives.sh -> config.sh
#
# All functions previously in this file have been moved to:
#   - session-core.sh: Session identification, locks, atomic operations
#   - session-state.sh: Session state queries, validation, team sync, worktree utils
#
# This file is kept for backward compatibility - existing code that sources
# session-utils.sh will continue to work unchanged.

# Source session-state.sh which brings in the full dependency chain
# This provides all functions that were previously defined in this file
source "$(dirname "${BASH_SOURCE[0]}")/session-state.sh"

# =============================================================================
# Backward Compatibility Notes
# =============================================================================
#
# Functions now available via session-core.sh:
#   - get_session_id()
#   - get_session_dir()
#   - get_current_session_file()
#   - generate_session_id()
#   - set_current_session()
#   - get_current_session()
#   - clear_current_session()
#   - map_tty_to_session() / set_session_for_tty()
#   - unmap_tty() / clear_session_for_tty()
#   - is_session_active()
#   - acquire_session_lock()
#   - release_session_lock()
#
# Functions now available via session-state.sh:
#   - get_session_state()
#   - get_session_field()
#   - set_session_field()
#   - is_parked()
#   - get_initiative()
#   - get_complexity()
#   - validate_session_context()
#   - validate_session_id_format()
#   - touch_session()
#   - is_session_stale()
#   - list_sessions()
#   - list_parked_sessions()
#   - list_stale_sessions()
#   - cleanup_stale_mappings()
#   - atomic_team_update()
#   - is_worktree()
#   - get_worktree_meta()
#   - get_worktree_field()
#
# Functions available via primitives.sh (transitive):
#   - md5_portable()
#   - date_portable_7days_ago()
#   - get_yaml_field()
#   - get_shell_session_id()
#   - get_tty_hash()
#   - atomic_write()
#   - atomic_write_stdin()
#
# Configuration available via config.sh (transitive):
#   - CLAUDE_PROJECT_DIR
#   - ROSTER_HOME
#   - SKELETON_HOME
#   - SESSIONS_DIR
#   - CURRENT_SESSION_FILE
#   - LOCK_DIR
#   - LOCK_TIMEOUT
#   - HOOK_TIMEOUT
#   - HOOKS_LOG_DIR
#   - HOOKS_LOG_FILE
#   - HOOKS_LOG_MAX_AGE_DAYS
#   - HOOKS_LOG_MAX_SIZE_MB
#   - SAFE_READ_COMMANDS
#   - SAFE_GIT_COMMANDS
#   - SAFE_GH_COMMANDS
