#!/usr/bin/env bash
# collect-content.sh - Collect .know/ files from org repos for Docker pre-baking.
#
# Usage: deploy/scripts/collect-content.sh [--sync] [--check-freshness] [--catalog path]
#
# Options:
#   --sync                Run 'ari registry sync --org autom8y' before collecting to ensure
#                         the domains.yaml catalog is up to date.
#   --check-freshness     Validate per-domain freshness against expires_after thresholds.
#                         Exits 1 if stale domains exceed threshold (default: 10).
#   --freshness-threshold N  Set the stale domain threshold for --check-freshness (default: 10).
#   --catalog  Path to domains.yaml (default: deploy/registry/domains.yaml).
#
# This script reads the domains.yaml catalog, determines which repos have
# .know/ domains, and copies their .know/ files into deploy/content/{repo}/
# preserving directory structure including nested scopes.
#
# For a repo with nested .know/ directories:
#   {repo}/.know/architecture.md           -> deploy/content/{repo}/.know/architecture.md
#   {repo}/services/ads/.know/arch.md      -> deploy/content/{repo}/services/ads/.know/arch.md
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
CHECK_FRESHNESS=false
FRESHNESS_THRESHOLD=10
CATALOG=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --sync)
            SYNC=true
            shift
            ;;
        --check-freshness)
            CHECK_FRESHNESS=true
            shift
            ;;
        --freshness-threshold)
            FRESHNESS_THRESHOLD="$2"
            shift 2
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
REPOS=$(grep -E '^\s+- name:' "$CATALOG" | sed 's/.*name:\s*//' | sort -u)

total=0
skipped=0
scoped=0
for repo in $REPOS; do
    repo_path="$REPO_BASE_DIR/$repo"

    if [ ! -d "$repo_path" ]; then
        echo "  SKIP $repo (directory not found at $repo_path)"
        skipped=$((skipped + 1))
        continue
    fi

    # Find ALL .know/ directories within the repo, excluding vendor/node_modules/etc.
    know_dirs=$(find "$repo_path" -type d -name '.know' \
        -not -path '*/vendor/*' \
        -not -path '*/node_modules/*' \
        -not -path '*/.git/*' \
        -not -path '*/.terraform/*' \
        -not -path '*/.knossos/worktrees/*' \
        2>/dev/null || true)

    if [ -z "$know_dirs" ]; then
        echo "  SKIP $repo (no .know/ directories found)"
        skipped=$((skipped + 1))
        continue
    fi

    repo_count=0
    while IFS= read -r know_dir; do
        # Compute the relative path from repo root to this .know/ directory.
        rel_know="${know_dir#$repo_path/}"

        # Count .md files in this .know/ directory (non-recursive, but include feat/release).
        count=$(find "$know_dir" -name '*.md' -type f | wc -l | tr -d ' ')
        if [ "$count" -eq 0 ]; then
            continue
        fi

        # Determine scope from the relative path.
        # ".know" -> root scope, "services/ads/.know" -> scoped
        scope_label="root"
        if [ "$rel_know" != ".know" ]; then
            scope_label="${rel_know%/.know}"
            scoped=$((scoped + count))
        fi

        # Copy all .md files preserving subdirectory structure.
        target="$OUTPUT_DIR/$repo/$rel_know"
        mkdir -p "$target"
        (cd "$know_dir" && find . -name '*.md' -type f -exec sh -c '
            for f; do
                dir=$(dirname "$f")
                mkdir -p "'"$target"'/$dir"
                cp "$f" "'"$target"'/$f"
            done
        ' _ {} +)

        repo_count=$((repo_count + count))
    done <<< "$know_dirs"

    if [ "$repo_count" -gt 0 ]; then
        echo "  OK   $repo ($repo_count files)"
        total=$((total + repo_count))
    else
        echo "  SKIP $repo (no .md files in any .know/ directory)"
        skipped=$((skipped + 1))
    fi
done

# Validate: check catalog domains against collected content using the path field.
echo ""
echo "--- Validation ---"
# Extract all path entries paired with their repo from the catalog.
# Uses the "path:" field which already contains the correct repo-relative path.
DOMAIN_ENTRIES=$(grep -E '^\s+(qualified_name|path):' "$CATALOG" 2>/dev/null | tr -d '"' || echo "")
missing=0
if [ -n "$DOMAIN_ENTRIES" ]; then
    current_qn=""
    while IFS= read -r line; do
        if echo "$line" | grep -q 'qualified_name:'; then
            current_qn=$(echo "$line" | sed 's/.*qualified_name:\s*//')
        elif echo "$line" | grep -q 'path:'; then
            domain_path=$(echo "$line" | sed 's/.*path:\s*//')
            # Extract repo name from qualified name: org::repo[/scope]::domain -> repo
            # The repo is everything between first :: and first / or second ::
            repo=$(echo "$current_qn" | sed 's/[^:]*::\([^:/]*\).*/\1/')
            content_file="$OUTPUT_DIR/$repo/$domain_path"
            if [ ! -f "$content_file" ]; then
                echo "  MISSING: $current_qn (expected at $content_file)"
                missing=$((missing + 1))
            fi
        fi
    done <<< "$DOMAIN_ENTRIES"
fi

echo ""
echo "=== Summary ==="
echo "  Repos processed:   $(echo "$REPOS" | wc -w | tr -d ' ')"
echo "  Repos skipped:     $skipped"
echo "  Files collected:   $total (root) + $scoped (scoped)"
echo "  Domains missing:   $missing"
echo "  Output directory:  $OUTPUT_DIR"

if [ "$missing" -gt 0 ]; then
    echo ""
    echo "WARNING: $missing catalog domains have no content."
    echo "  These domains will appear in startup coherence warnings."
    echo "  To fix: ensure repos are cloned at $REPO_BASE_DIR and re-run."
fi

# --- Freshness Check ---
# When --check-freshness is set, validate per-domain freshness against expires_after thresholds.
if [ "$CHECK_FRESHNESS" = true ] && command -v python3 &>/dev/null; then
    echo ""
    echo "--- Freshness Check ---"
    STALE_COUNT=$(python3 -c "
import yaml
from datetime import datetime, timezone

with open('$CATALOG') as f:
    data = yaml.safe_load(f)

now = datetime.now(timezone.utc)
stale = 0
total = 0
by_repo = {}

for repo in data.get('repos', []):
    rn = repo['name']
    repo_stale = 0
    repo_total = 0
    for domain in repo.get('domains', []):
        total += 1
        repo_total += 1
        gen_at = domain.get('generated_at', '')
        exp = domain.get('expires_after', '14d')
        exp_days = int(exp.replace('d','')) if exp.endswith('d') else 14
        if gen_at:
            try:
                dt = datetime.fromisoformat(gen_at.replace('Z', '+00:00'))
                age = (now - dt).days
                if age > exp_days:
                    stale += 1
                    repo_stale += 1
            except:
                stale += 1
                repo_stale += 1
    if repo_total > 0:
        by_repo[rn] = (repo_total, repo_stale)

for rn in sorted(by_repo):
    t, s = by_repo[rn]
    status = 'STALE' if s > 0 else 'FRESH'
    print(f'  {rn:<25} {t:>3} domains, {s:>3} stale  [{status}]')

print(f'')
print(f'  Total: {total} domains, {stale} stale')
print(stale)  # Last line is the count for shell to capture
" 2>/dev/null)

    # Extract the last line as the stale count.
    STALE_NUM=$(echo "$STALE_COUNT" | tail -1)
    # Print all but the last line as the report.
    echo "$STALE_COUNT" | sed '$ d'

    if [ "$STALE_NUM" -gt "$FRESHNESS_THRESHOLD" ] 2>/dev/null; then
        echo ""
        echo "FRESHNESS GATE FAILED: $STALE_NUM stale domains (threshold: $FRESHNESS_THRESHOLD)"
        echo "  Run 'ari know --all' in repos with stale domains, then re-collect."
        exit 1
    else
        echo ""
        echo "Freshness gate passed: $STALE_NUM stale domains (<= $FRESHNESS_THRESHOLD threshold)"
    fi
fi

echo ""
echo "Next: 'docker build -f deploy/Dockerfile -t clew:latest .' to bake into the container image."
