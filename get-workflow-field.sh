#!/bin/bash
# Extract specific field from workflow config
# Usage: ./get-workflow-field.sh <field> [team-name]
# Examples:
#   ./get-workflow-field.sh entry_point.agent
#   ./get-workflow-field.sh name
#   ./get-workflow-field.sh description

ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
FIELD="$1"
TEAM="${2:-$(cat .claude/ACTIVE_RITE 2>/dev/null)}"

if [ -z "$FIELD" ]; then
  echo "Error: No field specified" >&2
  echo "Usage: $0 <field> [team-name]" >&2
  echo "Common fields: name, description, entry_point.agent, workflow_type" >&2
  exit 1
fi

if [ -z "$TEAM" ]; then
  echo "Error: No team specified and no ACTIVE_TEAM found" >&2
  exit 1
fi

WORKFLOW_FILE="$ROSTER_HOME/rites/$TEAM/workflow.yaml"

if [ ! -f "$WORKFLOW_FILE" ]; then
  exit 1
fi

# Use yq if available for complex queries
if command -v yq &>/dev/null; then
  yq ".$FIELD" "$WORKFLOW_FILE" 2>/dev/null
else
  # Fallback: grep/awk for common fields
  case "$FIELD" in
    "name")
      grep "^name:" "$WORKFLOW_FILE" | head -1 | awk '{print $2}'
      ;;
    "workflow_type")
      grep "^workflow_type:" "$WORKFLOW_FILE" | head -1 | awk '{print $2}'
      ;;
    "description")
      grep "^description:" "$WORKFLOW_FILE" | head -1 | cut -d: -f2- | sed 's/^ *//'
      ;;
    "entry_point.agent")
      grep -A2 "^entry_point:" "$WORKFLOW_FILE" | grep "agent:" | head -1 | awk '{print $2}'
      ;;
    "entry_point.artifact.type")
      grep -A5 "^entry_point:" "$WORKFLOW_FILE" | grep "type:" | head -1 | awk '{print $2}'
      ;;
    "entry_point.artifact.path_template")
      grep -A5 "^entry_point:" "$WORKFLOW_FILE" | grep "path_template:" | head -1 | awk '{print $2}'
      ;;
    *)
      echo "Error: Complex field '$FIELD' requires yq" >&2
      echo "Install with: brew install yq" >&2
      exit 1
      ;;
  esac
fi
