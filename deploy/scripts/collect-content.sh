#!/usr/bin/env bash
# collect-content.sh - Collect .know/ files from org repos for Docker pre-baking.
#
# Usage: deploy/scripts/collect-content.sh [--sync] [--catalog path]
#
# Options:
#   --sync     Run 'ari registry sync --org autom8y' before collecting to ensure
#              the domains.yaml catalog is up to date.
#   --catalog  Path to domains.yaml (default: deploy/registry/domains.yaml).
#
# This script reads the domains.yaml catalog, determines which repos have
# .know/ domains, and copies their .know/ files into deploy/content/{repo}/.know/.
# The Docker build then COPYs this directory into the container image at /data/content/.
#
# Prerequisites:
#   - All repos must be cloned locally (siblings of the knossos repo, or specified via
#     REPO_BASE_DIR env var)
#   - domains.yaml must be up to date (run 'ari registry sync --org autom8y' first,
#     or use the --sync flag)
#
# The script is idempotent: running it twice produces identical output.
#
# Exit codes:
#   0  Success (content collected, all catalog domains have content)
#   1  Fatal error (missing catalog, bad args)
#   0  Partial success (some domains missing content -- logged as warnings)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$DEPLOY_DIR")"

# Parse arguments.
SYNC=false
CATALOG=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --sync)
            SYNC=true
            shift
            ;;
        --catalog)
            CATALOG="$2"
            shift 2
            ;;
        --catalog=*)
            CATALOG="${1#*=}"
            shift
            ;;
        *)
            # Positional fallback for backward compat.
            CATALOG="$1"
            shift
            ;;
    esac
done

CATALOG="${CATALOG:-$DEPLOY_DIR/registry/domains.yaml}"
OUTPUT_DIR="$DEPLOY_DIR/content"
REPO_BASE_DIR="${REPO_BASE_DIR:-$(dirname "$PROJECT_ROOT")}"

# Step 0 (optional): Sync the catalog.
if [ "$SYNC" = true ]; then
    echo "Syncing catalog via 'ari registry sync --org autom8y'..."
    if command -v ari &>/dev/null; then
        ari registry sync --org autom8y
        echo "  Catalog synced."
    else
        echo "  WARNING: 'ari' not found in PATH, skipping sync."
        echo "  Build ari first: CGO_ENABLED=0 go build -o ari ./cmd/ari"
    fi
fi

if [ ! -f "$CATALOG" ]; then
    echo "ERROR: catalog not found at $CATALOG"
    echo "Run 'ari registry sync --org autom8y' first, or use --sync flag."
    exit 1
fi

# Check catalog staleness (warn if > 7 days old).
if command -v python3 &>/dev/null; then
    SYNCED_AT=$(grep -E '^synced_at:' "$CATALOG" | head -1 | sed 's/synced_at:\s*//' | tr -d '"' || echo "")
    if [ -n "$SYNCED_AT" ]; then
        DAYS_OLD=$(python3 -c "
from datetime import datetime, timezone
import sys
try:
    synced = datetime.fromisoformat('$SYNCED_AT'.replace('Z', '+00:00'))
    age = (datetime.now(timezone.utc) - synced).days
    print(age)
except:
    print(-1)
" 2>/dev/null || echo "-1")
        if [ "$DAYS_OLD" -gt 7 ] 2>/dev/null; then
            echo "WARNING: catalog is ${DAYS_OLD} days old (synced_at: $SYNCED_AT)"
            echo "  Recommended: run with --sync or 'ari registry sync --org autom8y'"
        elif [ "$DAYS_OLD" -ge 0 ] 2>/dev/null; then
            echo "Catalog age: ${DAYS_OLD} days (synced_at: $SYNCED_AT)"
        fi
    fi
fi

echo ""
echo "Collecting .know/ content from repos..."
echo "  Catalog: $CATALOG"
echo "  Output:  $OUTPUT_DIR"
echo "  Repos:   $REPO_BASE_DIR"

# Clean output directory for idempotency.
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Extract unique repo names from the catalog.
# Uses grep to find "name:" entries under repos, then extracts the value.
REPOS=$(grep -E '^\s+- name:' "$CATALOG" | sed 's/.*name:\s*//' | sort -u)

total=0
skipped=0
for repo in $REPOS; do
    repo_path="$REPO_BASE_DIR/$repo"
    know_dir="$repo_path/.know"

    if [ ! -d "$know_dir" ]; then
        echo "  SKIP $repo (no .know/ directory at $repo_path)"
        skipped=$((skipped + 1))
        continue
    fi

    # Count .know/ files.
    count=$(find "$know_dir" -name '*.md' -type f | wc -l | tr -d ' ')
    if [ "$count" -eq 0 ]; then
        echo "  SKIP $repo (no .md files in .know/)"
        skipped=$((skipped + 1))
        continue
    fi

    # Copy .know/ directory.
    target="$OUTPUT_DIR/$repo/.know"
    mkdir -p "$target"

    # Copy preserving subdirectory structure (e.g., .know/feat/).
    (cd "$know_dir" && find . -name '*.md' -type f -exec sh -c '
        for f; do
            dir=$(dirname "$f")
            mkdir -p "'"$target"'/$dir"
            cp "$f" "'"$target"'/$f"
        done
    ' _ {} +)

    echo "  OK   $repo ($count files)"
    total=$((total + count))
done

# Validate: check for catalog domains that have no collected content.
echo ""
echo "--- Validation ---"
# Extract all qualified_name entries from the catalog.
DOMAIN_NAMES=$(grep -E '^\s+qualified_name:' "$CATALOG" 2>/dev/null | sed 's/.*qualified_name:\s*//' | tr -d '"' || echo "")
missing=0
if [ -n "$DOMAIN_NAMES" ]; then
    while IFS= read -r qn; do
        # Extract repo from qualified name (org::repo::domain -> repo).
        repo=$(echo "$qn" | cut -d: -f3)  # Using : delimiter; qualified names use ::
        repo=$(echo "$qn" | sed 's/.*::\(.*\)::.*/\1/')
        domain=$(echo "$qn" | sed 's/.*:://')

        content_file="$OUTPUT_DIR/$repo/.know/$domain.md"
        if [ ! -f "$content_file" ]; then
            echo "  MISSING: $qn (expected at $content_file)"
            missing=$((missing + 1))
        fi
    done <<< "$DOMAIN_NAMES"
fi

echo ""
echo "=== Summary ==="
echo "  Repos processed:   $(echo "$REPOS" | wc -w | tr -d ' ')"
echo "  Repos skipped:     $skipped"
echo "  Files collected:   $total"
echo "  Domains missing:   $missing"
echo "  Output directory:  $OUTPUT_DIR"

if [ "$missing" -gt 0 ]; then
    echo ""
    echo "WARNING: $missing catalog domains have no content."
    echo "  These domains will appear in startup coherence warnings."
    echo "  To fix: ensure repos are cloned at $REPO_BASE_DIR and re-run."
fi

echo ""
echo "Next: 'docker build -f deploy/Dockerfile -t clew:latest .' to bake into the container image."
