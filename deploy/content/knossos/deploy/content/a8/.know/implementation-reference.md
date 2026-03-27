# a8 Implementation Reference

Compiled best practices for implementing the a8 ecosystem control plane recipes.
Sources: just manual, yq docs, Docker Compose docs, clig.dev, AWS CLI reference, and ecosystem CLI patterns.

---

## 1. Justfile Patterns

### Built-in Constants (v1.37+) — USE THESE, not raw ANSI codes in recipe definitions

```
BOLD, NORMAL, RED, GREEN, YELLOW, BLUE, CYAN, MAGENTA, DIM, UNDERLINE
BG_RED, BG_GREEN, BG_YELLOW, etc.
```

Note: These work in `@echo` lines but NOT inside `#!/usr/bin/env bash` shebang bodies.
In shebang recipes, use raw `\033[32m` etc. since the recipe body is a bash script.

### Key Attributes

- `[group('name')]` — organize in `--list` output
- `[private]` — hide from `--list`
- `[no-exit-message]` — suppress "recipe failed" noise
- `[confirm('prompt')]` — built-in confirmation gate
- `[doc('text')]` — set doc comment shown in `--list`

### Parameter Patterns

```just
# Zero-or-more variadic (empty if omitted)
recipe *args:

# One-or-more variadic (error if omitted)
recipe +args:

# Export to env var
recipe $name:

# Default value
recipe target='staging':
```

### Import/Module Rules

- Imported files share variables and recipes (flat namespace)
- `source_directory()` returns the module's own dir (vs `justfile_directory()` = root)
- Variables from `_globals.just` are available in all imported modules

### Shell Recipe Best Practices

- Always use `#!/usr/bin/env bash` + `set -euo pipefail` for shebang recipes
- `{{var}}` substitution happens BEFORE bash sees the script — always quote: `"{{var}}"`
- `cd` persists in shebang recipes (whole body is one script)
- Use `[no-exit-message]` for recipes that produce their own error output

---

## 2. yq v4 Patterns (mikefarah/yq)

### Critical Rules

- Expression comes BEFORE filename: `yq '.path' file.yaml`
- Use `strenv(VAR)` for string env vars (not `env()` which parses YAML types)
- Env vars must be **exported** or **inline**: `KEY="val" yq '.x = strenv(KEY)' f.yaml`
- Use `//` for defaults: `yq '.field // "fallback"' f.yaml`
- Wrap LHS in parens for in-place updates: `yq '(.arr[] | select(.n == "x") | .v) = 1' f.yaml`

### Performance — Minimize yq Invocations

```bash
# BAD: N yq calls in a loop
for svc in $(yq '.services | keys | .[]' manifest.yaml); do
    arch=$(yq ".services[\"$svc\"].archetype" manifest.yaml)  # Another yq call per iteration
done

# GOOD: Single yq call, iterate in bash
while IFS=$'\t' read -r name arch enabled; do
    printf "%-20s %-25s %s\n" "$name" "$arch" "$enabled"
done < <(yq '.services | to_entries[] | [.key, .value.archetype, (.value.control.enabled // true)] | @tsv' manifest.yaml)
```

### Common Operations

```bash
# Read key
yq '.services.auth.archetype' manifest.yaml

# In-place update
yq -i '.services.auth.control.enabled = true' manifest.yaml

# Delete key
yq -i 'del(.old_field)' manifest.yaml

# Append to array
yq -i '.release_trains += [{"version": "2026.03", "status": "dev"}]' manifest.yaml

# Multiple updates in one pass
yq -i '.a = "1" | .b = "2" | del(.c)' manifest.yaml

# Iterate keys
yq '.services | keys | .[]' manifest.yaml

# Filter with select
yq '.services | to_entries[] | select(.value.control.enabled == true) | .key' manifest.yaml

# Length
yq '.release_trains | length' manifest.yaml

# Check for null
yq '.field // "default_value"' manifest.yaml
```

### Bash Integration — The JSON-per-line Pattern

```bash
# Best pattern for iterating complex objects
while IFS= read -r entry; do
    name=$(echo "$entry" | yq '.key')
    arch=$(echo "$entry" | yq '.value.archetype')
    # ...
done < <(yq -o=json -I=0 '.services | to_entries[]' manifest.yaml)
```

---

## 3. Terminal Output Patterns

### Color Convention

| Color | Meaning | Use For |
|-------|---------|---------|
| Green `\033[32m` | Success | `[OK]`, `enabled`, healthy |
| Yellow `\033[33m` | Warning | `[WARN]`, degraded, stale |
| Red `\033[31m` | Error/failure | `[ERROR]`, disabled, failed |
| Bold `\033[1m` | Emphasis | Headers, important values |
| Dim `\033[2m` | Secondary | Metadata, paths, timestamps |
| Reset `\033[0m` | Reset | After every color use |

### Table Formatting — printf Pattern

```bash
# Header + separator + data rows
printf "\033[1m%-20s %-25s %-9s %-12s %s\033[0m\n" "NAME" "ARCHETYPE" "ENABLED" "SCHEDULES" "DEPLOY"
printf "%-20s %-25s %-9s %-12s %s\n" "────────────────────" "─────────────────────────" "─────────" "────────────" "──────"
printf "%-20s %-25s \033[32m%-9s\033[0m %-12s %s\n" "auth" "ecs-fargate-rds" "true" "-" "ecs-deploy"
```

### Fail-Forward Pattern

```bash
ERRORS=0
WARNINGS=0

# Run all checks, don't abort early
check_something || ERRORS=$((ERRORS + 1))
check_another || WARNINGS=$((WARNINGS + 1))

# Summary at end
echo ""
echo "Summary: $((TOTAL - ERRORS - WARNINGS)) passed, $WARNINGS warnings, $ERRORS failed"
exit $((ERRORS > 0 ? 1 : 0))
```

### Doctor Pattern — Pass/Warn/Fail with Fix Suggestions

```bash
check_tool() {
    local name="$1" cmd="$2"
    if command -v "$cmd" >/dev/null 2>&1; then
        local ver
        ver=$("$cmd" --version 2>/dev/null | head -1 || echo "unknown")
        printf "  \033[32m[OK]\033[0m   %-12s %s\n" "$name" "$ver"
        PASSED=$((PASSED + 1))
    else
        printf "  \033[31m[FAIL]\033[0m %-12s not installed\n" "$name"
        printf "         \033[2m→ Install: %s\033[0m\n" "$hint"
        ERRORS=$((ERRORS + 1))
    fi
}
```

---

## 4. Docker Compose Patterns

### Multi-file Composition

- Files merge in order specified with `-f` flags
- Scalars: last file wins; Arrays: concatenated; Maps: merged by key
- All relative paths resolve from the FIRST compose file's directory
- Verify merged config: `docker compose -f a.yml -f b.yml config`

### Dynamic Override Discovery (a8 pattern)

```bash
FLAGS="-f compose/docker-compose.yml -f compose/docker-compose.platform.yml"
for svc in repos/autom8y/services/*/; do
    [[ -f "$svc/docker-compose.override.yml" ]] && FLAGS="$FLAGS -f $svc/docker-compose.override.yml"
done
for sat in repos/autom8y-*/; do
    if [[ -f "$sat/docker-compose.override.yml" ]]; then
        FLAGS="$FLAGS -f $sat/docker-compose.override.yml"
        SUFFIX=$(basename "$sat" | sed 's/^autom8y-//')
        UPPER=$(echo "$SUFFIX" | tr '[:lower:]-' '[:upper:]_')
        export "AUTOM8Y_${UPPER}_DIR=$sat"
    fi
done
```

### Health Check depends_on Conditions

- `service_started` — container running (default)
- `service_healthy` — healthcheck passes
- `service_completed_successfully` — container exits 0

---

## 5. AWS CLI Patterns

### Credential Check

```bash
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "AWS credentials not configured"
    return 1
fi
```

### ECS Service Status

```bash
aws ecs describe-services --cluster "$CLUSTER" --services "$SERVICE" \
    | jq -r '.services[0] | [.serviceName, .status, .runningCount, .desiredCount, (.deployments[0].updatedAt | todate)] | @tsv'
```

### Lambda Function Status

```bash
aws lambda get-function-configuration --function-name "$FN" \
    | jq -r '[.FunctionName, .State, .Runtime, (.LastModified | split("T")[0])] | @tsv'
```

### ECS Scale (enable/disable)

```bash
aws ecs update-service --cluster "$CLUSTER" --service "$SERVICE" --desired-count 0  # disable
aws ecs update-service --cluster "$CLUSTER" --service "$SERVICE" --desired-count 2  # enable
```

### Port Check (macOS)

```bash
if lsof -iTCP:"$PORT" -sTCP:LISTEN -n -P >/dev/null 2>&1; then
    echo "Port $PORT in use"
fi
```

---

## 6. Ecosystem Patterns

### Topological Sort (for SDK DAG)

Use `tsort` for topological ordering:
```bash
# Build edge pairs: dependency dependent
echo "autom8y-log autom8y-http"
echo "autom8y-config autom8y-core"
# ... pipe to tsort
```

### DOT Graph Generation

```bash
echo "digraph sdks {"
echo "  rankdir=BT;"
echo "  node [shape=box, style=rounded];"
yq '.sdks | to_entries[] | .key' manifest.yaml | while read -r sdk; do
    for dep in $(yq ".sdks[\"$sdk\"].depends_on[]" manifest.yaml 2>/dev/null); do
        echo "  \"$sdk\" -> \"$dep\";"
    done
done
echo "}"
```

### Version Comparison (for sat-audit)

```bash
# Compare semver using sort -V
version_lt() {
    local sorted_first
    sorted_first=$(printf '%s\n%s' "$1" "$2" | sort -V | head -1)
    [[ "$sorted_first" == "$1" && "$1" != "$2" ]]
}
```

### Scaffolding — Port Allocation

Scan existing compose overrides for used ports, find next available in range:
```bash
USED_PORTS=$(grep -rh 'ports:' repos/autom8y-*/docker-compose.override.yml 2>/dev/null \
    | grep -oE '[0-9]+:8000' | cut -d: -f1 | sort -n)
```

### Portable sed -i (macOS/Linux)

```bash
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|old|new|g" file
else
    sed -i "s|old|new|g" file
fi
```
