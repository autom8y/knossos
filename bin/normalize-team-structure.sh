#!/usr/bin/env bash
# normalize-team-structure.sh
# Creates missing required directories in team packs with .gitkeep files
#
# Usage: normalize-team-structure.sh [--dry-run]
#
# Part of REQ-3.4: Missing Commands Directory Structure

set -euo pipefail

ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
TEAMS_DIR="$ROSTER_HOME/rites"

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

normalize_team() {
    local team_dir="$1"
    local team_name
    team_name=$(basename "$team_dir")

    for dir in $REQUIRED_DIRS; do
        if [[ ! -d "$team_dir/$dir" ]]; then
            if $DRY_RUN; then
                echo "[DRY-RUN] Would create: $team_name/$dir/"
            else
                mkdir -p "$team_dir/$dir"
                touch "$team_dir/$dir/.gitkeep"
                echo "Created: $team_name/$dir/"
            fi
            ((CHANGES_MADE++)) || true
        fi
    done
}

main() {
    if $DRY_RUN; then
        echo "Dry-run: Checking team pack structure..."
    else
        echo "Normalizing team pack structure..."
    fi
    echo ""

    if [[ ! -d "$TEAMS_DIR" ]]; then
        echo "Error: Teams directory not found: $TEAMS_DIR"
        echo "Set ROSTER_HOME environment variable to your roster repository"
        exit 1
    fi

    for team_dir in "$TEAMS_DIR"/*/; do
        [[ -d "$team_dir" ]] || continue
        normalize_team "$team_dir"
    done

    echo ""
    if $DRY_RUN; then
        echo "Dry-run complete. $CHANGES_MADE directories would be created."
    else
        if [[ $CHANGES_MADE -gt 0 ]]; then
            echo "Complete. $CHANGES_MADE directories created."
        else
            echo "Complete. All teams already have required directories."
        fi
    fi
}

main "$@"
