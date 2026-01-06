#!/bin/bash
# generate-orchestrator.sh - Generate orchestrator.md from template + config
# Usage: ./generate-orchestrator.sh <team-name> [--dry-run]
#
# POC SHORTCUTS (see end of file for full list):
# - Minimal error handling
# - Hardcoded paths
# - No input validation
# - Single-team focus (rnd-pack)

set -e

# Source Knossos home resolution (handles ROSTER_HOME deprecation)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/knossos-home.sh"

TEAM="${1:-rnd-pack}"
DRY_RUN="${2:-}"

TEMPLATE="$KNOSSOS_HOME/templates/base-orchestrator.md"
CONFIG="$KNOSSOS_HOME/rites/$TEAM/orchestrator.yaml"
WORKFLOW="$KNOSSOS_HOME/rites/$TEAM/workflow.yaml"
OUTPUT="$KNOSSOS_HOME/rites/$TEAM/agents/orchestrator.md"

# Check dependencies
if ! command -v yq &> /dev/null; then
    echo "ERROR: yq is required. Install via: brew install yq"
    exit 1
fi

# Verify files exist
[[ -f "$TEMPLATE" ]] || { echo "ERROR: Template not found: $TEMPLATE"; exit 1; }
[[ -f "$CONFIG" ]] || { echo "ERROR: Config not found: $CONFIG"; exit 1; }
[[ -f "$WORKFLOW" ]] || { echo "ERROR: Workflow not found: $WORKFLOW"; exit 1; }

echo "Generating orchestrator for: $TEAM"
echo "  Template: $TEMPLATE"
echo "  Config: $CONFIG"
echo "  Workflow: $WORKFLOW"
echo "  Output: $OUTPUT"

# --- Extract values from orchestrator.yaml ---

ROLE=$(yq '.frontmatter.role' "$CONFIG")
DESCRIPTION=$(yq '.frontmatter.description' "$CONFIG" | tr '\n' ' ' | sed 's/  */ /g' | sed 's/ *$//')
COLOR=$(yq '.team.color' "$CONFIG")
TEAM_NAME=$(yq '.team.name' "$CONFIG")

# Workflow position (upstream/downstream)
UPSTREAM_SOURCES=$(yq '.workflow_position.upstream' "$CONFIG" | tr -d '"')
DOWNSTREAM_AGENTS=$(yq '.workflow_position.downstream' "$CONFIG" | tr -d '"')

# --- Extract values from workflow.yaml ---

# Get specialist names from phases (in workflow order, not sorted)
SPECIALISTS=$(yq '.phases[].agent' "$WORKFLOW")

# Build complexity enum from complexity_levels
COMPLEXITY_ENUM=$(yq '.complexity_levels[].name' "$WORKFLOW" | tr '\n' ' ' | sed 's/ *$//' | sed 's/ / | /g')
COMPLEXITY_ENUM="\"$COMPLEXITY_ENUM\""

# Build specialist enum
SPECIALIST_ENUM=$(echo "$SPECIALISTS" | tr '\n' ' ' | sed 's/ *$//' | sed 's/ /" | "/g')
SPECIALIST_ENUM="\"$SPECIALIST_ENUM\""

# --- Generate workflow diagram ---
# POC: Hardcoded structure for rnd-pack linear workflow
# Production would need to parse workflow.yaml graph structure

generate_workflow_diagram() {
    local agents=($SPECIALISTS)
    local num_agents=${#agents[@]}

    # Header (common to all)
    echo "                    +-----------------+"
    echo "                    |   ORCHESTRATOR  |"
    echo "                    +--------+--------+"
    echo "                             |"

    case $num_agents in
        1|2|3)
            # Simple horizontal layout
            local sep="        "
            for agent in "${agents[@]}"; do
                echo "${sep}+-> $agent"
            done
            ;;
        4)
            # 2x2 grid layout
            echo "        +----------+----------+"
            echo "        v          v          v"
            printf "   %-14s %-14s %-14s\n" "${agents[0]}" "${agents[1]}" "${agents[2]}"
            echo "        |          |          |"
            echo "        +----------+----------+"
            echo "                   |"
            echo "                   v"
            printf "              %-14s\n" "${agents[3]}"
            ;;
        5)
            # Top row: 3, Bottom row: 2
            echo "   +-----------+-----------+-----------+"
            echo "   v           v           v           "
            printf "%-12s %-12s %-12s\n" "${agents[0]}" "${agents[1]}" "${agents[2]}"
            echo "   |           |           |           "
            echo "   +-----------+-----------+           "
            echo "               |                       "
            echo "       +-------+-------+               "
            echo "       v               v               "
            printf "   %-12s     %-12s\n" "${agents[3]}" "${agents[4]}"
            ;;
        6)
            # 2 rows of 3
            echo "   +-----------+-----------+-----------+"
            echo "   v           v           v           "
            printf "%-12s %-12s %-12s\n" "${agents[0]}" "${agents[1]}" "${agents[2]}"
            echo "   |           |           |           "
            echo "   +-----------+-----------+-----------+"
            echo "   v           v           v           "
            printf "%-12s %-12s %-12s\n" "${agents[3]}" "${agents[4]}" "${agents[5]}"
            ;;
        *)
            # 7+ agents: compact list with count
            echo "        +--- ${num_agents} specialists ---+"
            for agent in "${agents[@]}"; do
                echo "        | $agent"
            done
            echo "        +------------------------+"
            ;;
    esac
}

WORKFLOW_DIAGRAM=$(generate_workflow_diagram)

# --- Generate phase routing section ---

generate_phase_routing() {
    local agents=($SPECIALISTS)

    echo "## Phase Routing"
    echo ""
    echo "| Specialist | Route When |"
    echo "|------------|------------|"

    for agent in "${agents[@]}"; do
        local condition=$(yq ".routing[\"$agent\"]" "$CONFIG" | tr -d '"')
        echo "| $agent | $condition |"
    done
}

PHASE_ROUTING=$(generate_phase_routing)

# --- Generate skills reference ---

generate_skills_reference() {
    yq '.skills[]' "$CONFIG" | while read -r skill; do
        echo "- $skill"
    done
}

SKILLS_REFERENCE=$(generate_skills_reference)

# --- Generate handoff criteria ---

generate_handoff_criteria() {
    local phases
    phases=$(yq '.phases[].name' "$WORKFLOW")

    echo "| Phase | Criteria |"
    echo "|-------|----------|"

    for phase in $phases; do
        local criteria
        criteria=$(yq ".handoff_criteria.$phase[]" "$CONFIG" 2>/dev/null | \
            sed 's/^/- /' | tr '\n' '<br>')
        if [[ -n "$criteria" ]]; then
            echo "| $phase | ${criteria%<br>} |"
        fi
    done
}

HANDOFF_CRITERIA=$(generate_handoff_criteria)

# --- Generate cross-team protocol (conditional) ---

generate_cross_team_protocol() {
    local protocol
    protocol=$(yq '.cross_team_protocol' "$CONFIG" | tr -d '"')

    if [[ -n "$protocol" && "$protocol" != "null" && "$protocol" != "" ]]; then
        cat <<EOF

## Cross-Team Protocol

$protocol

When routing cross-team concerns:
1. Identify the affected team(s)
2. Include current session context in handoff
3. Notify user of cross-team escalation
4. Track resolution in throughline
EOF
    fi
}

CROSS_TEAM_PROTOCOL=$(generate_cross_team_protocol)

# --- Generate team-specific antipatterns (conditional) ---

generate_team_antipatterns() {
    local antipatterns
    antipatterns=$(yq '.antipatterns[]' "$CONFIG" 2>/dev/null)

    if [[ -n "$antipatterns" ]]; then
        echo ""
        echo "### Team-Specific Anti-Patterns"
        echo ""
        yq '.antipatterns[]' "$CONFIG" | while read -r pattern; do
            echo "- **$pattern**"
        done
    fi
}

TEAM_ANTIPATTERNS=$(generate_team_antipatterns)

# --- Perform substitutions ---

# Read template
CONTENT=$(cat "$TEMPLATE")

# Simple substitutions
CONTENT="${CONTENT//\{\{ROLE\}\}/$ROLE}"
CONTENT="${CONTENT//\{\{TEAM_COLOR\}\}/$COLOR}"
CONTENT="${CONTENT//\{\{COMPLEXITY_ENUM\}\}/$COMPLEXITY_ENUM}"
CONTENT="${CONTENT//\{\{SPECIALIST_ENUM\}\}/$SPECIALIST_ENUM}"
CONTENT="${CONTENT//\{\{TEAM_NAME\}\}/$TEAM_NAME}"
CONTENT="${CONTENT//\{\{UPSTREAM_SOURCES\}\}/$UPSTREAM_SOURCES}"
CONTENT="${CONTENT//\{\{DOWNSTREAM_AGENTS\}\}/$DOWNSTREAM_AGENTS}"

# Multi-line substitutions need special handling
# POC: Use sed for these

# Create temp file for processing
TMPFILE=$(mktemp)
echo "$CONTENT" > "$TMPFILE"

# Replace TEAM_DESCRIPTION (may have special chars)
DESCRIPTION_ESCAPED=$(echo "$DESCRIPTION" | sed 's/[&/\]/\\&/g')
sed -i '' "s|{{TEAM_DESCRIPTION}}|$DESCRIPTION_ESCAPED|g" "$TMPFILE"

# Replace multi-line blocks
# Workflow diagram
DIAGRAM_FILE=$(mktemp)
echo "$WORKFLOW_DIAGRAM" > "$DIAGRAM_FILE"
# Use awk for multi-line replacement
awk -v file="$DIAGRAM_FILE" '
/\{\{WORKFLOW_DIAGRAM\}\}/ {
    while ((getline line < file) > 0) print line
    close(file)
    next
}
{ print }
' "$TMPFILE" > "${TMPFILE}.new"
mv "${TMPFILE}.new" "$TMPFILE"
rm -f "$DIAGRAM_FILE"

# Routing table
ROUTING_FILE=$(mktemp)
echo "$ROUTING_TABLE" > "$ROUTING_FILE"
awk -v file="$ROUTING_FILE" '
/\{\{ROUTING_TABLE\}\}/ {
    while ((getline line < file) > 0) print line
    close(file)
    next
}
{ print }
' "$TMPFILE" > "${TMPFILE}.new"
mv "${TMPFILE}.new" "$TMPFILE"
rm -f "$ROUTING_FILE"

# Skills reference
SKILLS_FILE=$(mktemp)
echo "$SKILLS_REFERENCE" > "$SKILLS_FILE"
awk -v file="$SKILLS_FILE" '
/\{\{SKILLS_REFERENCE\}\}/ {
    while ((getline line < file) > 0) print line
    close(file)
    next
}
{ print }
' "$TMPFILE" > "${TMPFILE}.new"
mv "${TMPFILE}.new" "$TMPFILE"
rm -f "$SKILLS_FILE"

# Handoff criteria
HANDOFF_FILE=$(mktemp)
echo "$HANDOFF_CRITERIA" > "$HANDOFF_FILE"
awk -v file="$HANDOFF_FILE" '
/\{\{HANDOFF_CRITERIA\}\}/ {
    while ((getline line < file) > 0) print line
    close(file)
    next
}
{ print }
' "$TMPFILE" > "${TMPFILE}.new"
mv "${TMPFILE}.new" "$TMPFILE"
rm -f "$HANDOFF_FILE"

# Phase routing
ROUTING_FILE=$(mktemp)
echo "$PHASE_ROUTING" > "$ROUTING_FILE"
awk -v file="$ROUTING_FILE" '
/\{\{PHASE_ROUTING\}\}/ {
    while ((getline line < file) > 0) print line
    close(file)
    next
}
{ print }
' "$TMPFILE" > "${TMPFILE}.new"
mv "${TMPFILE}.new" "$TMPFILE"
rm -f "$ROUTING_FILE"

# Cross-team protocol (conditional)
if [[ -n "$CROSS_TEAM_PROTOCOL" ]]; then
    PROTOCOL_FILE=$(mktemp)
    echo "$CROSS_TEAM_PROTOCOL" > "$PROTOCOL_FILE"
    awk -v file="$PROTOCOL_FILE" '
    /\{\{CROSS_TEAM_PROTOCOL\}\}/ {
        while ((getline line < file) > 0) print line
        close(file)
        next
    }
    { print }
    ' "$TMPFILE" > "${TMPFILE}.new"
    mv "${TMPFILE}.new" "$TMPFILE"
    rm -f "$PROTOCOL_FILE"
else
    # Remove placeholder line entirely
    sed -i '' '/{{CROSS_TEAM_PROTOCOL}}/d' "$TMPFILE"
fi

# Team-specific antipatterns (conditional)
if [[ -n "$TEAM_ANTIPATTERNS" ]]; then
    ANTIPATTERNS_FILE=$(mktemp)
    echo "$TEAM_ANTIPATTERNS" > "$ANTIPATTERNS_FILE"
    awk -v file="$ANTIPATTERNS_FILE" '
    /\{\{TEAM_SPECIFIC_ANTIPATTERNS\}\}/ {
        while ((getline line < file) > 0) print line
        close(file)
        next
    }
    { print }
    ' "$TMPFILE" > "${TMPFILE}.new"
    mv "${TMPFILE}.new" "$TMPFILE"
    rm -f "$ANTIPATTERNS_FILE"
else
    # Remove placeholder line entirely
    sed -i '' '/{{TEAM_SPECIFIC_ANTIPATTERNS}}/d' "$TMPFILE"
fi

# --- Output ---

if [[ "$DRY_RUN" == "--dry-run" ]]; then
    echo ""
    echo "=== Generated Content (dry-run) ==="
    cat "$TMPFILE"
    echo ""
    echo "=== End Generated Content ==="
else
    cp "$TMPFILE" "$OUTPUT"
    echo ""
    echo "Generated: $OUTPUT"
fi

rm -f "$TMPFILE"

echo ""
echo "Done."
