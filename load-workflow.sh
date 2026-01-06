#!/bin/bash
# Load workflow config for current or specified team
# Usage: ./load-workflow.sh [team-name]
# Output: Workflow YAML to stdout, or error message to stderr

set -e

# Source Knossos home resolution (handles ROSTER_HOME deprecation)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/knossos-home.sh"
TEAM="${1:-$(cat .claude/ACTIVE_RITE 2>/dev/null)}"

if [ -z "$TEAM" ]; then
  echo "Error: No team specified and no ACTIVE_RITE found" >&2
  echo "Usage: $0 [team-name]" >&2
  echo "Available teams:" >&2
  ls -1 "$KNOSSOS_HOME/rites" 2>/dev/null | sed 's/^/  /' >&2
  exit 1
fi

WORKFLOW_FILE="$KNOSSOS_HOME/rites/$TEAM/workflow.yaml"

if [ ! -f "$WORKFLOW_FILE" ]; then
  echo "Error: No workflow.yaml found for team '$TEAM'" >&2
  echo "Expected: $WORKFLOW_FILE" >&2
  echo "" >&2
  echo "Create a workflow.yaml following the schema at:" >&2
  echo "  $KNOSSOS_HOME/workflow-schema.yaml" >&2
  exit 1
fi

cat "$WORKFLOW_FILE"
