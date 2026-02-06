#!/bin/bash
# Load workflow config for current or specified rite
# Usage: ./load-workflow.sh [rite-name]
# Output: Workflow YAML to stdout, or error message to stderr

set -e

# Source Knossos home resolution (resolves KNOSSOS_HOME)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/knossos-home.sh"
RITE="${1:-$(cat .claude/ACTIVE_RITE 2>/dev/null)}"

if [ -z "$RITE" ]; then
  echo "Error: No rite specified and no ACTIVE_RITE found" >&2
  echo "Usage: $0 [rite-name]" >&2
  echo "Available rites:" >&2
  ls -1 "$KNOSSOS_HOME/rites" 2>/dev/null | sed 's/^/  /' >&2
  exit 1
fi

WORKFLOW_FILE="$KNOSSOS_HOME/rites/$RITE/workflow.yaml"

if [ ! -f "$WORKFLOW_FILE" ]; then
  echo "Error: No workflow.yaml found for rite '$RITE'" >&2
  echo "Expected: $WORKFLOW_FILE" >&2
  echo "" >&2
  echo "Create a workflow.yaml following the schema at:" >&2
  echo "  $KNOSSOS_HOME/workflow-schema.yaml" >&2
  exit 1
fi

cat "$WORKFLOW_FILE"
