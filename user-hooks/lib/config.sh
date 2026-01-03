#!/bin/bash
# Configuration for hooks - sourced first by all hooks
# NO LOGIC - only variable definitions
#
# Addresses: CEM-001 (scattered configuration)
# Part of Ecosystem v2 refactoring (RF-001)

# =============================================================================
# Project Paths
# =============================================================================

# Project root directory
export CLAUDE_PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# External repository paths
export ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
# Note: SKELETON_HOME deprecated - roster is now standalone (Sprint 4 migration)

# Session paths
export SESSIONS_DIR="$CLAUDE_PROJECT_DIR/.claude/sessions"
export CURRENT_SESSION_FILE="$SESSIONS_DIR/.current-session"

# Lock management
export LOCK_DIR="$CLAUDE_PROJECT_DIR/.claude/sessions/.locks"

# =============================================================================
# Timeouts
# =============================================================================

# Lock acquisition timeout (seconds)
export LOCK_TIMEOUT="${LOCK_TIMEOUT:-5}"

# Default hook timeout (seconds)
export HOOK_TIMEOUT="${HOOK_TIMEOUT:-5}"

# =============================================================================
# Logging Configuration
# =============================================================================

export HOOKS_LOG_DIR="$CLAUDE_PROJECT_DIR/.claude/logs"
export HOOKS_LOG_FILE="$HOOKS_LOG_DIR/hooks.log"
export HOOKS_LOG_MAX_AGE_DAYS="${HOOKS_LOG_MAX_AGE_DAYS:-7}"
export HOOKS_LOG_MAX_SIZE_MB="${HOOKS_LOG_MAX_SIZE_MB:-10}"

# =============================================================================
# Safe Command Patterns (data-driven allowlist for validators)
# =============================================================================

# These patterns are used by PreToolUse validators for auto-approval
# Extend by editing this section - no code changes required

export SAFE_READ_COMMANDS="cat|head|tail|less|wc"
export SAFE_GIT_COMMANDS="status|branch|log|diff|symbolic-ref|rev-list|rev-parse|remote|config|show"
export SAFE_GH_COMMANDS="pr|issue"
