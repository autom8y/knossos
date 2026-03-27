#!/usr/bin/env bash
# build-knowledge-index.sh - Generate knowledge-index.json for Docker pre-baking.
#
# Usage: deploy/scripts/build-knowledge-index.sh [--catalog path] [--content path] [--output path]
#
# This script builds and runs cmd/build-knowledge-index to produce a knowledge-index.json
# file that gets COPY'd into the Docker image. This eliminates cold-start LLM calls
# for summary generation in the container.
#
# Prerequisites:
#   - Go 1.22+ installed
#   - ANTHROPIC_API_KEY set in environment (for Haiku summary generation)
#   - deploy/registry/domains.yaml up to date (run collect-content.sh first)
#   - deploy/content/ populated (run collect-content.sh first)
#
# The script is idempotent: running it twice produces a valid index (the builder
# uses source_hash comparison to skip unchanged domains).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$DEPLOY_DIR")"

# Parse arguments.
CATALOG=""
CONTENT=""
OUTPUT=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --catalog)
            CATALOG="$2"
            shift 2
            ;;
        --catalog=*)
            CATALOG="${1#*=}"
            shift
            ;;
        --content)
            CONTENT="$2"
            shift 2
            ;;
        --content=*)
            CONTENT="${1#*=}"
            shift
            ;;
        --output)
            OUTPUT="$2"
            shift 2
            ;;
        --output=*)
            OUTPUT="${1#*=}"
            shift
            ;;
        *)
            echo "Unknown argument: $1"
            echo "Usage: $0 [--catalog path] [--content path] [--output path]"
            exit 1
            ;;
    esac
done

CATALOG="${CATALOG:-$DEPLOY_DIR/registry/domains.yaml}"
CONTENT="${CONTENT:-$DEPLOY_DIR/content}"
OUTPUT="${OUTPUT:-$DEPLOY_DIR/knowledge-index.json}"

# Validate prerequisites.
if [ ! -f "$CATALOG" ]; then
    echo "ERROR: catalog not found at $CATALOG"
    echo "Run 'deploy/scripts/collect-content.sh' first."
    exit 1
fi

if [ ! -d "$CONTENT" ]; then
    echo "ERROR: content directory not found at $CONTENT"
    echo "Run 'deploy/scripts/collect-content.sh' first."
    exit 1
fi

if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
    echo "ERROR: ANTHROPIC_API_KEY is not set."
    echo "Export your API key: export ANTHROPIC_API_KEY=sk-ant-..."
    exit 1
fi

echo "Building knowledge index..."
echo "  Catalog: $CATALOG"
echo "  Content: $CONTENT"
echo "  Output:  $OUTPUT"
echo ""

# Build and run the tool.
cd "$PROJECT_ROOT"
CGO_ENABLED=0 go run ./cmd/build-knowledge-index \
    --catalog "$CATALOG" \
    --content "$CONTENT" \
    --output "$OUTPUT"

echo ""
echo "Done. knowledge-index.json is ready for Docker build."
echo "Next: docker build -f deploy/Dockerfile -t clew:latest ."
