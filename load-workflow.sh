#!/bin/bash
# Load workflow config for current or specified team
# Usage: ./load-workflow.sh [team-name]
# Output: Workflow YAML to stdout, or error message to stderr

set -e

ROSTER_HOME="${ROSTER_HOME:-$HOME/Code/roster}"
TEAM="${1:-$(cat .claude/ACTIVE_TEAM 2>/dev/null)}"

if [ -z "$TEAM" ]; then
  echo "Error: No team specified and no ACTIVE_TEAM found" >&2
  echo "Usage: $0 [team-name]" >&2
  echo "Available teams:" >&2
  ls -1 "$ROSTER_HOME/teams" 2>/dev/null | sed 's/^/  /' >&2
  exit 1
fi

WORKFLOW_FILE="$ROSTER_HOME/teams/$TEAM/workflow.yaml"

if [ ! -f "$WORKFLOW_FILE" ]; then
  echo "Error: No workflow.yaml found for team '$TEAM'" >&2
  echo "Expected: $WORKFLOW_FILE" >&2
  echo "" >&2
  echo "Create a workflow.yaml following the schema at:" >&2
  echo "  $ROSTER_HOME/workflow-schema.yaml" >&2
  exit 1
fi

cat "$WORKFLOW_FILE"
