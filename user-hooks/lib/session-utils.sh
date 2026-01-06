#!/bin/bash
# Session utilities for multi-session support
#
# Part of Ecosystem v2 (RF-006)
#
# ARCHITECTURE:
# This file sources session-state.sh which provides the full dependency chain:
#   session-utils.sh -> session-state.sh -> session-core.sh -> primitives.sh -> config.sh
#
# Function locations:
#   - session-core.sh: Session identification, locks, atomic operations
#   - session-state.sh: Session state queries, validation, rite sync, worktree utils

# Source session-state.sh which brings in the full dependency chain
# shellcheck source=session-state.sh
source "$(dirname "${BASH_SOURCE[0]}")/session-state.sh"
