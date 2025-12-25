#!/bin/bash
# Generate team routing context for session injection
# Usage: ./generate-team-context.sh [team-name]
# Output: Markdown routing table for specified team
# Exit 0 with no output if team/workflow not found (graceful degradation)

set -e

ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
TEAM="${1:-$(cat .claude/ACTIVE_TEAM 2>/dev/null || echo "")}"

# Exit gracefully if no team
if [[ -z "$TEAM" ]]; then
    exit 0
fi

WORKFLOW_FILE="$ROSTER_HOME/teams/$TEAM/workflow.yaml"

# Exit gracefully if workflow doesn't exist
if [[ ! -f "$WORKFLOW_FILE" ]]; then
    exit 0
fi

# Extract entry point agent
ENTRY_AGENT=$(grep -A2 "^entry_point:" "$WORKFLOW_FILE" | grep "agent:" | head -1 | awk '{print $2}' || echo "unknown")

# Extract complexity levels (just the names, comma-separated)
# Need to parse the full complexity_levels section
COMPLEXITY_LEVELS=""
IN_COMPLEXITY=0
while IFS= read -r line; do
    if [[ "$line" =~ ^complexity_levels: ]]; then
        IN_COMPLEXITY=1
        continue
    fi
    if [[ $IN_COMPLEXITY -eq 1 ]] && [[ "$line" =~ ^[a-z_]+: ]] && [[ ! "$line" =~ ^[[:space:]]+ ]]; then
        break
    fi
    if [[ $IN_COMPLEXITY -eq 1 ]] && [[ "$line" =~ ^[[:space:]]*-[[:space:]]+name:[[:space:]]+(.+)$ ]]; then
        if [[ -n "$COMPLEXITY_LEVELS" ]]; then
            COMPLEXITY_LEVELS="$COMPLEXITY_LEVELS, ${BASH_REMATCH[1]}"
        else
            COMPLEXITY_LEVELS="${BASH_REMATCH[1]}"
        fi
    fi
done < "$WORKFLOW_FILE"

# Start output
echo "## Team Context: $TEAM"
echo ""
echo "| Phase | Agent | Artifact |"
echo "|-------|-------|----------|"

# Parse phases section
# We need to extract: name, agent, produces (artifact type)
IN_PHASES=0
while IFS= read -r line; do
    # Detect start of phases section
    if [[ "$line" =~ ^phases: ]]; then
        IN_PHASES=1
        CURRENT_PHASE=""
        CURRENT_AGENT=""
        CURRENT_ARTIFACT=""
        continue
    fi

    # Detect end of phases section (next top-level key or complexity_levels)
    if [[ $IN_PHASES -eq 1 ]] && [[ "$line" =~ ^[a-z_]+: ]] && [[ ! "$line" =~ ^[[:space:]]+ ]]; then
        IN_PHASES=0
        break
    fi

    # Process phase entries
    if [[ $IN_PHASES -eq 1 ]]; then
        # Phase name (starts with "  - name:")
        if [[ "$line" =~ ^[[:space:]]*-[[:space:]]+name:[[:space:]]+(.+)$ ]]; then
            # Output previous phase if we have complete data
            if [[ -n "$CURRENT_PHASE" && -n "$CURRENT_AGENT" && -n "$CURRENT_ARTIFACT" ]]; then
                echo "| $CURRENT_PHASE | $CURRENT_AGENT | $CURRENT_ARTIFACT |"
            fi
            CURRENT_PHASE="${BASH_REMATCH[1]}"
            CURRENT_AGENT=""
            CURRENT_ARTIFACT=""
        fi

        # Agent name
        if [[ "$line" =~ ^[[:space:]]+agent:[[:space:]]+(.+)$ ]]; then
            CURRENT_AGENT="${BASH_REMATCH[1]}"
        fi

        # Artifact type (produces field)
        if [[ "$line" =~ ^[[:space:]]+produces:[[:space:]]+(.+)$ ]]; then
            CURRENT_ARTIFACT="${BASH_REMATCH[1]}"
        fi
    fi
done < "$WORKFLOW_FILE"

# Output final phase
if [[ -n "$CURRENT_PHASE" && -n "$CURRENT_AGENT" && -n "$CURRENT_ARTIFACT" ]]; then
    echo "| $CURRENT_PHASE | $CURRENT_AGENT | $CURRENT_ARTIFACT |"
fi

# Output footer with entry and complexity
echo ""
echo "**Entry**: $ENTRY_AGENT | **Complexity**: $COMPLEXITY_LEVELS"
echo "**Routing**: Match task to phase, invoke that agent via Task tool."
