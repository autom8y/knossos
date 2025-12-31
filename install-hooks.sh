#!/usr/bin/env bash
#
# install-hooks.sh - Install hooks from roster/hooks/ to a target .claude/hooks/
#
# Installs hooks from roster/hooks/ (canonical source) to a target location.
# Works for both user-level (~/.claude/hooks/) and project-level (.claude/hooks/).
#
# Usage:
#   ./install-hooks.sh                         # Install to current project
#   ./install-hooks.sh /path/to/project        # Install to specified project
#   ./install-hooks.sh ~/.claude               # Install to user-level
#   ./install-hooks.sh --dry-run               # Preview changes
#
# Environment Variables:
#   ROSTER_HOME    Roster repository location (default: ~/Code/roster)

set -euo pipefail

readonly ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
readonly SOURCE_DIR="$ROSTER_HOME/user-hooks"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

DRY_RUN=0
TARGET_PROJECT=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--dry-run] [project-path]"
            echo ""
            echo "Syncs hooks from roster/hooks/ to project .claude/hooks/"
            echo ""
            echo "Options:"
            echo "  --dry-run    Preview changes without applying"
            echo "  --help       Show this help message"
            echo ""
            echo "Arguments:"
            echo "  project-path  Path to project (default: current directory)"
            exit 0
            ;;
        *)
            TARGET_PROJECT="$1"
            shift
            ;;
    esac
done

# Determine target project
if [[ -z "$TARGET_PROJECT" ]]; then
    TARGET_PROJECT="$(pwd)"
fi

TARGET_HOOKS="$TARGET_PROJECT/.claude/hooks"

# Validate source exists
if [[ ! -d "$SOURCE_DIR" ]]; then
    log_error "Source directory not found: $SOURCE_DIR"
    exit 1
fi

# Validate target project has .claude directory
if [[ ! -d "$TARGET_PROJECT/.claude" ]]; then
    log_error "Not a Claude project: $TARGET_PROJECT (no .claude directory)"
    exit 1
fi

log_info "Syncing hooks from: $SOURCE_DIR"
log_info "Syncing hooks to:   $TARGET_HOOKS"

if [[ $DRY_RUN -eq 1 ]]; then
    log_warn "DRY RUN - no changes will be made"
fi

# Create target directory if needed
if [[ ! -d "$TARGET_HOOKS" ]]; then
    if [[ $DRY_RUN -eq 0 ]]; then
        mkdir -p "$TARGET_HOOKS"
        log_info "Created: $TARGET_HOOKS"
    else
        log_info "Would create: $TARGET_HOOKS"
    fi
fi

# Create lib directory if needed
if [[ ! -d "$TARGET_HOOKS/lib" ]]; then
    if [[ $DRY_RUN -eq 0 ]]; then
        mkdir -p "$TARGET_HOOKS/lib"
        log_info "Created: $TARGET_HOOKS/lib"
    else
        log_info "Would create: $TARGET_HOOKS/lib"
    fi
fi

# Sync hook files
sync_count=0
for hook in "$SOURCE_DIR"/*.sh; do
    if [[ -f "$hook" ]]; then
        hook_name=$(basename "$hook")
        target_file="$TARGET_HOOKS/$hook_name"

        if [[ $DRY_RUN -eq 0 ]]; then
            cp "$hook" "$target_file"
            chmod +x "$target_file"
            log_info "Synced: $hook_name"
        else
            log_info "Would sync: $hook_name"
        fi
        sync_count=$((sync_count + 1))
    fi
done

# Sync library files
lib_count=0
for lib in "$SOURCE_DIR/lib"/*.sh; do
    if [[ -f "$lib" ]]; then
        lib_name=$(basename "$lib")
        target_file="$TARGET_HOOKS/lib/$lib_name"

        if [[ $DRY_RUN -eq 0 ]]; then
            cp "$lib" "$target_file"
            chmod +x "$target_file" 2>/dev/null || true
            log_info "Synced lib: $lib_name"
        else
            log_info "Would sync lib: $lib_name"
        fi
        lib_count=$((lib_count + 1))
    fi
done

echo ""
log_info "Summary: $sync_count hooks, $lib_count libraries"

if [[ $DRY_RUN -eq 1 ]]; then
    log_warn "DRY RUN complete - run without --dry-run to apply changes"
fi
