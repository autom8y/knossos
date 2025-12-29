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

ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
TEAM="${1:-rnd-pack}"
DRY_RUN="${2:-}"

TEMPLATE="$ROSTER_HOME/templates/orchestrator-base.md.tpl"
CONFIG="$ROSTER_HOME/teams/$TEAM/orchestrator.yaml"
WORKFLOW="$ROSTER_HOME/teams/$TEAM/workflow.yaml"
OUTPUT="$ROSTER_HOME/teams/$TEAM/agents/orchestrator.md"

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

    # Build a simple linear diagram
    echo "                    +-----------------+"
    echo "                    |   ORCHESTRATOR  |"
    echo "                    +--------+--------+"
    echo "                             |"

    if [[ $num_agents -eq 4 ]]; then
        # 4-agent layout (like rnd-pack)
        echo "        +--------------------+--------------------+"
        echo "        v                    v                    v"
        printf "+---------------+   +---------------+   +---------------+\n"
        printf "|  %-11s |-->|  %-11s |-->|   %-10s |\n" "$(echo ${agents[0]} | cut -d'-' -f1)" "$(echo ${agents[1]} | cut -d'-' -f1)" "$(echo ${agents[2]} | cut -d'-' -f1)"
        printf "|  %-11s |   |  %-11s |   |   %-10s |\n" "$(echo ${agents[0]} | cut -d'-' -f2-)" "$(echo ${agents[1]} | cut -d'-' -f2-)" "$(echo ${agents[2]} | cut -d'-' -f2-)"
        printf "+---------------+   +---------------+   +---------------+\n"
        echo "                                              |"
        echo "                                              v"
        echo "                                       +---------------+"
        printf "                                       |   %-10s |\n" "$(echo ${agents[3]} | cut -d'-' -f1)"
        printf "                                       |   %-10s |\n" "$(echo ${agents[3]} | cut -d'-' -f2-)"
        echo "                                       +---------------+"
    else
        # Fallback: simple vertical list
        for agent in "${agents[@]}"; do
            echo "        +-> $agent"
        done
    fi
}

WORKFLOW_DIAGRAM=$(generate_workflow_diagram)

# --- Generate routing table ---

generate_routing_table() {
    local agents=($SPECIALISTS)
    for agent in "${agents[@]}"; do
        local condition=$(yq ".routing[\"$agent\"]" "$CONFIG")
        echo "| $agent | $condition |"
    done
}

ROUTING_TABLE=$(generate_routing_table)

# --- Generate skills reference ---

generate_skills_reference() {
    yq '.skills[]' "$CONFIG" | while read -r skill; do
        echo "- $skill"
    done
}

SKILLS_REFERENCE=$(generate_skills_reference)

# --- Perform substitutions ---

# Read template
CONTENT=$(cat "$TEMPLATE")

# Simple substitutions
CONTENT="${CONTENT//\{\{ROLE\}\}/$ROLE}"
CONTENT="${CONTENT//\{\{COLOR\}\}/$COLOR}"
CONTENT="${CONTENT//\{\{COMPLEXITY_ENUM\}\}/$COMPLEXITY_ENUM}"
CONTENT="${CONTENT//\{\{SPECIALIST_ENUM\}\}/$SPECIALIST_ENUM}"

# Multi-line substitutions need special handling
# POC: Use sed for these

# Create temp file for processing
TMPFILE=$(mktemp)
echo "$CONTENT" > "$TMPFILE"

# Replace DESCRIPTION (may have special chars)
DESCRIPTION_ESCAPED=$(echo "$DESCRIPTION" | sed 's/[&/\]/\\&/g')
sed -i '' "s|{{DESCRIPTION}}|$DESCRIPTION_ESCAPED|g" "$TMPFILE"

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
