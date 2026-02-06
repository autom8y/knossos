#!/usr/bin/env bash
# fix-hardcoded-paths.sh
# Replaces hardcoded ~/Code/knossos paths with $KNOSSOS_HOME
#
# Usage: fix-hardcoded-paths.sh [--dry-run] [--no-backup]
#
# Part of REQ-3.2: Path Portability via Environment Variables
#
# EXCEPTIONS (paths NOT replaced):
# - Default value documentation (e.g., "default: ~/Code/knossos")
# - Already using variable fallback pattern (e.g., "${KNOSSOS_HOME:-~/Code/knossos}")
# - Archive/backup directories (.archive, .backup)
# - Session artifacts (transient data)

set -euo pipefail

# Source Knossos home resolution (handles KNOSSOS_HOME deprecation)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/knossos-home.sh"
DRY_RUN=false
NO_BACKUP=false
CHANGES_MADE=0

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-backup)
            NO_BACKUP=true
            shift
            ;;
        *)
            echo "Usage: $0 [--dry-run] [--no-backup]"
            exit 1
            ;;
    esac
done

# Files to skip (patterns)
SKIP_PATTERNS=(
    ".archive"
    ".backup"
    "/sessions/"
    "SMELL-REPORT"
    "REFACTOR-PLAN"
    "PHASE-5-HANDOFF"
    "fix-hardcoded-paths.sh"
)

should_skip_file() {
    local file="$1"
    for pattern in "${SKIP_PATTERNS[@]}"; do
        if [[ "$file" == *"$pattern"* ]]; then
            return 0
        fi
    done
    return 1
}

# Lines to skip (patterns that should NOT be replaced)
# These include default value documentation and already-variable patterns
should_skip_line() {
    local line="$1"

    # Skip lines with fallback pattern (already portable)
    if [[ "$line" == *'${KNOSSOS_HOME:-~/Code/knossos}'* ]]; then
        return 0
    fi

    # Skip default value documentation
    if [[ "$line" == *'default:'*'~/Code/knossos'* ]] || \
       [[ "$line" == *'Default:'*'~/Code/knossos'* ]] || \
       [[ "$line" == *'(default:'*'~/Code/knossos'* ]]; then
        return 0
    fi

    return 1
}

process_file() {
    local file="$1"
    local rel_path="${file#$KNOSSOS_HOME/}"

    if should_skip_file "$file"; then
        if $DRY_RUN; then
            echo "[SKIP] $rel_path (archive/session)"
        fi
        return 0
    fi

    # Check if file has any occurrences
    if ! grep -q '~/Code/knossos' "$file" 2>/dev/null; then
        return 0
    fi

    # Count standalone occurrences (not in fallback pattern)
    local count
    count=$(grep -c '~/Code/knossos' "$file" 2>/dev/null | grep -v '${KNOSSOS_HOME:-' || echo "0")

    if $DRY_RUN; then
        echo "[WOULD UPDATE] $rel_path"
        grep -n '~/Code/knossos' "$file" 2>/dev/null | while read -r line; do
            line_content=$(echo "$line" | cut -d: -f2-)
            if should_skip_line "$line_content"; then
                echo "  [SKIP] $line"
            else
                echo "  [REPLACE] $line"
            fi
        done
    else
        # Create temp file for processing
        local temp_file
        temp_file=$(mktemp)

        # Process line by line
        while IFS= read -r line || [[ -n "$line" ]]; do
            if should_skip_line "$line"; then
                echo "$line"
            else
                # Replace ~/Code/knossos with $KNOSSOS_HOME
                echo "$line" | sed 's|~/Code/knossos|\$KNOSSOS_HOME|g'
            fi
        done < "$file" > "$temp_file"

        # Check if file changed
        if ! diff -q "$file" "$temp_file" >/dev/null 2>&1; then
            mv "$temp_file" "$file"
            echo "[UPDATED] $rel_path"
            ((CHANGES_MADE++)) || true
        else
            rm -f "$temp_file"
        fi
    fi
}

main() {
    if $DRY_RUN; then
        echo "Dry-run: Checking hardcoded paths in knossos..."
    else
        echo "Fixing hardcoded paths in knossos..."

        # Create backup unless disabled
        if ! $NO_BACKUP; then
            local backup_dir="$KNOSSOS_HOME/.path-fix-backup-$(date +%Y%m%d-%H%M%S)"
            echo "Creating backup at $backup_dir"
            mkdir -p "$backup_dir"

            # Backup files that will be modified
            for file in $(grep -rl '~/Code/knossos' "$KNOSSOS_HOME" --include='*.md' --include='*.sh' 2>/dev/null || true); do
                if ! should_skip_file "$file"; then
                    rel_path="${file#$KNOSSOS_HOME/}"
                    mkdir -p "$backup_dir/$(dirname "$rel_path")"
                    cp "$file" "$backup_dir/$rel_path"
                fi
            done
        fi
    fi
    echo ""

    # Find all .md and .sh files
    local files
    files=$(grep -rl '~/Code/knossos' "$KNOSSOS_HOME" --include='*.md' --include='*.sh' 2>/dev/null || true)

    for file in $files; do
        process_file "$file"
    done

    echo ""
    if $DRY_RUN; then
        echo "Dry-run complete. Run without --dry-run to apply changes."
    else
        echo "Complete. $CHANGES_MADE files updated."
    fi
}

main "$@"
