#!/usr/bin/env bash
# normalize-rite-structure.sh
# Creates missing required directories in rites with .gitkeep files
#
# Usage: normalize-rite-structure.sh [--dry-run]
#
# Part of REQ-3.4: Missing Commands Directory Structure

set -euo pipefail

# Source Knossos home resolution
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/knossos-home.sh"
RITES_DIR="$KNOSSOS_HOME/rites"

# Required directories (empty with .gitkeep if no content)
REQUIRED_DIRS="commands skills"

# Track changes
CHANGES_MADE=0
DRY_RUN=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Usage: $0 [--dry-run]"
            exit 1
            ;;
    esac
done

normalize_rite() {
    local rite_dir="$1"
    local rite_name
    rite_name=$(basename "$rite_dir")

    for dir in $REQUIRED_DIRS; do
        if [[ ! -d "$rite_dir/$dir" ]]; then
            if $DRY_RUN; then
                echo "[DRY-RUN] Would create: $rite_name/$dir/"
            else
                mkdir -p "$rite_dir/$dir"
                touch "$rite_dir/$dir/.gitkeep"
                echo "Created: $rite_name/$dir/"
            fi
            ((CHANGES_MADE++)) || true
        fi
    done
}

main() {
    if $DRY_RUN; then
        echo "Dry-run: Checking rite structure..."
    else
        echo "Normalizing rite structure..."
    fi
    echo ""

    if [[ ! -d "$RITES_DIR" ]]; then
        echo "Error: Rites directory not found: $RITES_DIR"
        echo "Set KNOSSOS_HOME environment variable to your knossos repository"
        exit 1
    fi

    for rite_dir in "$RITES_DIR"/*/; do
        [[ -d "$rite_dir" ]] || continue
        normalize_rite "$rite_dir"
    done

    echo ""
    if $DRY_RUN; then
        echo "Dry-run complete. $CHANGES_MADE directories would be created."
    else
        if [[ $CHANGES_MADE -gt 0 ]]; then
            echo "Complete. $CHANGES_MADE directories created."
        else
            echo "Complete. All rites already have required directories."
        fi
    fi
}

main "$@"
