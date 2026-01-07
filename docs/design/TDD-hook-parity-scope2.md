# TDD: Settings.json Hook Management (Scope 2)

## Overview

This Technical Design Document specifies the implementation of automated settings.json hook registration management for the roster ecosystem. The design enables rites to declaratively define hook configurations (event type, path, matcher, timeout) in YAML, which `swap-rite.sh` merges and generates into `settings.local.json` during team swaps.

## Context

| Reference | Location |
|-----------|----------|
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` |
| Sprint | `sprint-hook-parity-20251231` |
| Task | `task-003` |
| Current settings.local.json | `/Users/tomtenuta/Code/roster/.claude/settings.local.json` |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh` |
| Related Scope | Scope 1 (Team Hooks Parity - file sync) |

### Problem Statement

Currently, `swap-rite.sh` syncs hook FILES to `.claude/hooks/` but does not update hook REGISTRATIONS in `settings.local.json`. This creates several issues:

1. **Manual Registration**: After swapping teams, users must manually edit `settings.local.json` to register new hooks
2. **No Team Customization**: Teams cannot specify custom matchers, timeouts, or event types for their hooks
3. **Inconsistent State**: Hook files exist but aren't registered, causing confusion
4. **No Merge Strategy**: No way to combine base hooks with team-specific hooks

### Design Goals

1. Declarative YAML schema for hook registration configuration
2. Base hooks configuration at `roster/user-hooks/base_hooks.yaml`
3. Team-specific hooks configuration at `rites/<team>/hooks.yaml`
4. Automated `settings.local.json` generation during team swap
5. Merge strategy: base hooks first, team hooks append per event type
6. Preservation of non-roster hooks in `settings.local.json`

### Requirements Coverage

| Requirement | Description | Addressed In |
|-------------|-------------|--------------|
| FR-2.1 | Hook registration schema | YAML Schema Definition |
| FR-2.2 | `base_hooks.yaml` in `roster/user-hooks/` | Base Hooks Configuration |
| FR-2.3 | Team `hooks.yaml` for team-specific registrations | Team Hooks Configuration |
| FR-2.4 | `swap-rite.sh` generates settings.local.json | JSON Generation Algorithm |
| FR-2.5 | Merge strategy: base first, team append | Merge Strategy Design |

---

## System Design

### Architecture Diagram

```
                                    +-----------------------+
                                    |   swap-rite.sh        |
                                    +-----------+-----------+
                                                |
                        +-----------------------+-----------------------+
                        |                       |                       |
                        v                       v                       v
              +------------------+    +------------------+    +------------------+
              | roster/          |    | rites/<team>/    |    | .claude/         |
              | user-hooks/      |    | hooks.yaml       |    | settings.local   |
              | base_hooks.yaml  |    | (optional)       |    | .json            |
              +--------+---------+    +--------+---------+    +--------+---------+
                       |                       |                       |
                       v                       v                       |
              +------------------------------------------------+      |
              |          Hook Registration Merger              |      |
              |  +------------------------------------------+  |      |
              |  | 1. Parse base_hooks.yaml                 |  |      |
              |  | 2. Parse team hooks.yaml (if exists)     |  |      |
              |  | 3. Validate all registrations            |  |      |
              |  | 4. Merge by event type (append)          |  |      |
              |  +------------------------------------------+  |      |
              +------------------------+-----------------------+      |
                                       |                              |
                                       v                              v
                              +------------------+           +------------------+
                              | Generated hooks  |           | Preserved user   |
                              | configuration    |<--------->| hooks (non-roster)|
                              +------------------+           +------------------+
                                       |
                                       v
                              +------------------+
                              | settings.local   |
                              | .json (output)   |
                              +------------------+
```

### Components

| Component | Responsibility | Location |
|-----------|---------------|----------|
| **Base Hooks Config** | Default hook registrations for all teams | `roster/user-hooks/base_hooks.yaml` |
| **Team Hooks Config** | Team-specific hook registrations | `rites/<team>/hooks.yaml` |
| **Hook Registration Merger** | Parses, validates, merges YAML configs | `swap-rite.sh` (new functions) |
| **Settings Generator** | Writes merged config to settings.local.json | `swap-rite.sh` (new functions) |
| **User Hook Preserver** | Identifies and preserves non-roster hooks | `swap-rite.sh` (new functions) |

---

## YAML Schema Definition

### Hook Registration Schema

```yaml
# Schema: hook-registration.schema.yaml
# Version: 1.0.0

schema_version: "1.0"

# Hook registrations organized by event type
hooks:
  # Event type: SessionStart | Stop | PreToolUse | PostToolUse | UserPromptSubmit
  - event: <event_type>
    # Matcher pattern (regex, optional for some events)
    matcher: <matcher_pattern>
    # Relative path to hook script (from .claude/hooks/)
    path: <hook_script.sh>
    # Timeout in seconds (1-60, default 5)
    timeout: <seconds>
    # Optional: description for documentation
    description: <string>
```

### Field Definitions

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `schema_version` | string | Yes | - | Must be "1.0" |
| `hooks` | array | Yes | - | At least one entry |
| `hooks[].event` | enum | Yes | - | One of: SessionStart, Stop, PreToolUse, PostToolUse, UserPromptSubmit |
| `hooks[].matcher` | string | Conditional | - | Required for: PreToolUse, PostToolUse, SessionStart. Regex pattern. |
| `hooks[].path` | string | Yes | - | Relative path, must exist in `.claude/hooks/` |
| `hooks[].timeout` | integer | No | 5 | Range: 1-60 seconds |
| `hooks[].description` | string | No | - | Human-readable description |

### Event Type Specifications

| Event | Matcher Required | Matcher Examples | Notes |
|-------|-----------------|------------------|-------|
| `SessionStart` | Optional | `startup\|resume`, `startup` | Matches session start reason |
| `Stop` | No | - | Fires when Claude finishes responding |
| `PreToolUse` | Yes | `Bash`, `Edit\|Write`, `*` | Tool name pattern |
| `PostToolUse` | Yes | `Write`, `Bash`, `Edit\|Write` | Tool name pattern |
| `UserPromptSubmit` | Optional | `^/`, `^/start` | Matches user prompt content |

### Validation Rules

1. **Event Validity**: `event` must be one of the allowed values
2. **Matcher Requirement**: PreToolUse and PostToolUse require a non-empty matcher
3. **Path Existence**: Hook path must resolve to existing file after sync
4. **Timeout Range**: If provided, timeout must be 1-60 (clamped with warning if exceeded)
5. **Matcher Syntax**: Matcher must be valid regex (warn and skip on invalid)

---

## Base Hooks Configuration

### File Location

`/Users/tomtenuta/Code/roster/user-hooks/base_hooks.yaml`

### Example Configuration

```yaml
# Base Hooks Registration
# These hooks are applied to all rites
# Team-specific hooks.yaml will append to these registrations

schema_version: "1.0"

hooks:
  # === SessionStart Hooks ===
  - event: SessionStart
    matcher: "startup|resume"
    path: session-context.sh
    timeout: 10
    description: "Injects session context on startup and resume"

  - event: SessionStart
    matcher: "startup|resume"
    path: coach-mode.sh
    timeout: 5
    description: "Activates coach mode guidance"

  # === Stop Hooks ===
  - event: Stop
    path: auto-park.sh
    timeout: 5
    description: "Auto-parks session on stop"

  # === PostToolUse Hooks ===
  - event: PostToolUse
    matcher: "Write"
    path: artifact-tracker.sh
    timeout: 5
    description: "Tracks written artifacts"

  - event: PostToolUse
    matcher: "Write"
    path: session-audit.sh
    timeout: 5
    description: "Audits session file writes"

  - event: PostToolUse
    matcher: "Bash"
    path: commit-tracker.sh
    timeout: 5
    description: "Tracks git commits"

  # === PreToolUse Hooks ===
  - event: PreToolUse
    matcher: "Bash"
    path: command-validator.sh
    timeout: 5
    description: "Validates bash commands"

  - event: PreToolUse
    matcher: "Edit|Write"
    path: session-write-guard.sh
    timeout: 3
    description: "Guards session context writes"

  - event: PreToolUse
    matcher: "Edit|Write"
    path: delegation-check.sh
    timeout: 3
    description: "Checks for delegation requirements"

  # === UserPromptSubmit Hooks ===
  - event: UserPromptSubmit
    matcher: "^/"
    path: start-preflight.sh
    timeout: 5
    description: "Preflight check for slash commands only"
```

---

## Team Hooks Configuration

### File Location

`/Users/tomtenuta/Code/roster/rites/<team-name>/hooks.yaml`

### Example: Security Pack

```yaml
# Security Pack Hooks
# Appended to base hooks during team swap

schema_version: "1.0"

hooks:
  # Credential scanning on file writes
  - event: PreToolUse
    matcher: "Edit|Write"
    path: credential-scanner.sh
    timeout: 10
    description: "Scans for hardcoded credentials"

  # Security audit on bash commands
  - event: PreToolUse
    matcher: "Bash"
    path: security-command-audit.sh
    timeout: 5
    description: "Audits bash commands for security risks"

  # Post-write security check
  - event: PostToolUse
    matcher: "Write"
    path: post-write-security-check.sh
    timeout: 5
    description: "Validates written files for security issues"
```

### Example: 10x-dev-pack

```yaml
# 10x Dev Pack Hooks
# Currently inherits all base hooks without additions

schema_version: "1.0"

hooks: []
# No team-specific hooks - uses base hooks only
```

---

## JSON Generation Algorithm

### Overview

The generation algorithm:
1. Reads existing `settings.local.json` and extracts non-roster hooks
2. Parses and validates `base_hooks.yaml`
3. Parses and validates team `hooks.yaml` (if exists)
4. Merges registrations by event type
5. Generates Claude Code hook JSON format
6. Preserves non-roster hooks in output
7. Writes result to `settings.local.json`

### Pseudocode

```bash
generate_settings_hooks() {
    local rite_name="$1"
    local settings_file=".claude/settings.local.json"
    local base_hooks="$ROSTER_HOME/user-hooks/base_hooks.yaml"
    local team_hooks="$ROSTER_HOME/rites/$rite_name/hooks.yaml"

    # Step 1: Backup existing settings
    local backup_file="${settings_file}.backup"
    [[ -f "$settings_file" ]] && cp "$settings_file" "$backup_file"

    # Step 2: Extract non-roster hooks from existing settings
    local preserved_hooks
    preserved_hooks=$(extract_non_roster_hooks "$settings_file")

    # Step 3: Parse and validate base hooks
    local base_registrations
    base_registrations=$(parse_hooks_yaml "$base_hooks")
    [[ $? -ne 0 ]] && { log_error "Failed to parse base_hooks.yaml"; return 1; }

    # Step 4: Parse team hooks (optional)
    local team_registrations=""
    if [[ -f "$team_hooks" ]]; then
        team_registrations=$(parse_hooks_yaml "$team_hooks")
        if [[ $? -ne 0 ]]; then
            log_warning "Failed to parse team hooks.yaml, using base only"
            team_registrations=""
        fi
    fi

    # Step 5: Merge registrations by event type
    local merged_registrations
    merged_registrations=$(merge_hook_registrations "$base_registrations" "$team_registrations")

    # Step 6: Generate Claude Code JSON format
    local generated_json
    generated_json=$(generate_hooks_json "$merged_registrations")

    # Step 7: Merge with preserved hooks
    local final_hooks
    final_hooks=$(merge_with_preserved "$generated_json" "$preserved_hooks")

    # Step 8: Update settings.local.json
    update_settings_hooks "$settings_file" "$final_hooks"

    # Step 9: Validate result
    if ! validate_settings_json "$settings_file"; then
        log_error "Generated settings.json is invalid, rolling back"
        [[ -f "$backup_file" ]] && mv "$backup_file" "$settings_file"
        return 1
    fi

    rm -f "$backup_file"
    log "Hook registrations updated in settings.local.json"
}
```

### YAML Parsing Function

```bash
parse_hooks_yaml() {
    local yaml_file="$1"
    local output=""

    # Validate file exists
    [[ -f "$yaml_file" ]] || { echo ""; return 0; }

    # Validate schema version
    local schema_version
    schema_version=$(yq -r '.schema_version // ""' "$yaml_file" 2>/dev/null)
    if [[ "$schema_version" != "1.0" ]]; then
        log_warning "Unknown schema version: $schema_version (expected 1.0)"
    fi

    # Parse hooks array
    local hook_count
    hook_count=$(yq -r '.hooks | length' "$yaml_file" 2>/dev/null)
    [[ "$hook_count" -eq 0 ]] && { echo ""; return 0; }

    # Process each hook entry
    for ((i=0; i<hook_count; i++)); do
        local event matcher path timeout description

        event=$(yq -r ".hooks[$i].event // \"\"" "$yaml_file")
        matcher=$(yq -r ".hooks[$i].matcher // \"\"" "$yaml_file")
        path=$(yq -r ".hooks[$i].path // \"\"" "$yaml_file")
        timeout=$(yq -r ".hooks[$i].timeout // 5" "$yaml_file")
        description=$(yq -r ".hooks[$i].description // \"\"" "$yaml_file")

        # Validate event type
        case "$event" in
            SessionStart|Stop|PreToolUse|PostToolUse|UserPromptSubmit) ;;
            *)
                log_warning "Invalid event type: $event (skipping)"
                continue
                ;;
        esac

        # Validate matcher requirement
        if [[ "$event" == "PreToolUse" || "$event" == "PostToolUse" ]] && [[ -z "$matcher" ]]; then
            log_warning "Event $event requires matcher (skipping: $path)"
            continue
        fi

        # Validate matcher syntax (basic regex check)
        if [[ -n "$matcher" ]]; then
            if ! echo "" | grep -E "$matcher" >/dev/null 2>&1; then
                log_warning "Invalid matcher regex: $matcher (skipping: $path)"
                continue
            fi
        fi

        # Clamp timeout to 60s max
        if [[ "$timeout" -gt 60 ]]; then
            log_warning "Timeout $timeout exceeds 60s limit, clamping to 60 (hook: $path)"
            timeout=60
        fi
        if [[ "$timeout" -lt 1 ]]; then
            timeout=5
        fi

        # Emit registration record (JSON-lines format for internal processing)
        echo "{\"event\":\"$event\",\"matcher\":\"$matcher\",\"path\":\"$path\",\"timeout\":$timeout}"
    done
}
```

### Merge Strategy Implementation

```bash
merge_hook_registrations() {
    local base_registrations="$1"
    local team_registrations="$2"

    # Combine all registrations (base first, team second)
    local all_registrations
    all_registrations=$(printf '%s\n%s' "$base_registrations" "$team_registrations" | grep -v '^$')

    # Group by event type and output merged list
    echo "$all_registrations"
}

generate_hooks_json() {
    local registrations="$1"

    # Group registrations by event type
    local events=("SessionStart" "Stop" "PreToolUse" "PostToolUse" "UserPromptSubmit")

    echo "{"

    local first_event=true
    for event in "${events[@]}"; do
        # Filter registrations for this event
        local event_hooks
        event_hooks=$(echo "$registrations" | jq -sc "[.[] | select(.event == \"$event\")]" 2>/dev/null)

        # Skip empty events
        local count
        count=$(echo "$event_hooks" | jq 'length')
        [[ "$count" -eq 0 ]] && continue

        # Add comma separator
        if [[ "$first_event" == "true" ]]; then
            first_event=false
        else
            echo ","
        fi

        # Generate Claude Code format for this event
        echo "    \"$event\": ["
        generate_event_hooks_json "$event" "$event_hooks"
        echo "    ]"
    done

    echo "}"
}

generate_event_hooks_json() {
    local event="$1"
    local hooks="$2"

    # Group by matcher for this event
    local matchers
    matchers=$(echo "$hooks" | jq -r '.[].matcher' | sort -u)

    local first_matcher=true
    while IFS= read -r matcher; do
        [[ -z "$matcher" && "$event" != "Stop" && "$event" != "UserPromptSubmit" ]] && continue

        if [[ "$first_matcher" == "true" ]]; then
            first_matcher=false
        else
            echo ","
        fi

        # Get all hooks for this matcher
        local matcher_hooks
        if [[ -n "$matcher" ]]; then
            matcher_hooks=$(echo "$hooks" | jq -c "[.[] | select(.matcher == \"$matcher\")]")
        else
            matcher_hooks=$(echo "$hooks" | jq -c "[.[] | select(.matcher == \"\")]")
        fi

        echo "      {"
        if [[ -n "$matcher" ]]; then
            echo "        \"matcher\": \"$matcher\","
        fi
        echo "        \"hooks\": ["

        # Generate individual hook entries
        local hook_count
        hook_count=$(echo "$matcher_hooks" | jq 'length')
        for ((i=0; i<hook_count; i++)); do
            local path timeout
            path=$(echo "$matcher_hooks" | jq -r ".[$i].path")
            timeout=$(echo "$matcher_hooks" | jq -r ".[$i].timeout")

            [[ $i -gt 0 ]] && echo ","
            cat <<HOOK
          {
            "type": "command",
            "command": "\$CLAUDE_PROJECT_DIR/.claude/hooks/$path",
            "timeout": $timeout
          }
HOOK
        done

        echo "        ]"
        echo "      }"
    done <<< "$matchers"
}
```

---

## User Hook Preservation Strategy

### Problem

Users may have manually added hooks to `settings.local.json` that are not managed by roster. These must be preserved during team swaps.

### Detection Algorithm

```bash
extract_non_roster_hooks() {
    local settings_file="$1"

    [[ -f "$settings_file" ]] || { echo "{}"; return 0; }

    # Read current hooks
    local current_hooks
    current_hooks=$(jq '.hooks // {}' "$settings_file" 2>/dev/null)

    # Get list of roster-managed hook paths
    local roster_hooks_pattern='\$CLAUDE_PROJECT_DIR/.claude/hooks/'

    # For each event type, filter out roster hooks
    local preserved="{}"

    for event in SessionStart Stop PreToolUse PostToolUse UserPromptSubmit; do
        local event_entries
        event_entries=$(echo "$current_hooks" | jq -c ".\"$event\" // []")

        # Filter each entry's hooks to exclude roster-managed ones
        local filtered_entries="[]"
        local entry_count
        entry_count=$(echo "$event_entries" | jq 'length')

        for ((i=0; i<entry_count; i++)); do
            local entry
            entry=$(echo "$event_entries" | jq -c ".[$i]")

            # Filter hooks array within entry
            local filtered_hooks
            filtered_hooks=$(echo "$entry" | jq -c '[.hooks[] | select(.command | contains(".claude/hooks/") | not)]')

            local filtered_count
            filtered_count=$(echo "$filtered_hooks" | jq 'length')

            if [[ "$filtered_count" -gt 0 ]]; then
                # Update entry with filtered hooks
                local new_entry
                new_entry=$(echo "$entry" | jq -c ".hooks = $filtered_hooks")
                filtered_entries=$(echo "$filtered_entries" | jq -c ". + [$new_entry]")
            fi
        done

        local filtered_len
        filtered_len=$(echo "$filtered_entries" | jq 'length')
        if [[ "$filtered_len" -gt 0 ]]; then
            preserved=$(echo "$preserved" | jq -c ".\"$event\" = $filtered_entries")
        fi
    done

    echo "$preserved"
}
```

### Merge with Generated Hooks

```bash
merge_with_preserved() {
    local generated="$1"
    local preserved="$2"

    # For each event type, append preserved hooks to generated
    local merged="$generated"

    for event in SessionStart Stop PreToolUse PostToolUse UserPromptSubmit; do
        local preserved_entries
        preserved_entries=$(echo "$preserved" | jq -c ".\"$event\" // []")

        local preserved_count
        preserved_count=$(echo "$preserved_entries" | jq 'length')
        [[ "$preserved_count" -eq 0 ]] && continue

        # Append preserved entries to generated event
        local generated_entries
        generated_entries=$(echo "$merged" | jq -c ".\"$event\" // []")

        local combined
        combined=$(echo "$generated_entries $preserved_entries" | jq -sc 'add')

        merged=$(echo "$merged" | jq -c ".\"$event\" = $combined")
    done

    echo "$merged"
}
```

### Roster Hook Marker

To reliably identify roster-managed hooks, all generated hooks use a consistent path pattern:

```
$CLAUDE_PROJECT_DIR/.claude/hooks/<script-name>.sh
```

Hooks with commands NOT matching this pattern are considered user hooks and are preserved.

---

## Integration with swap-rite.sh

### New Function: `swap_hook_registrations()`

```bash
# Sync hook registrations to settings.local.json
# Called after swap_hooks() syncs the actual hook files
swap_hook_registrations() {
    local rite_name="$1"
    local settings_file=".claude/settings.local.json"
    local base_hooks_yaml="$ROSTER_HOME/user-hooks/base_hooks.yaml"
    local team_hooks_yaml="$ROSTER_HOME/rites/$rite_name/hooks.yaml"

    log_debug "Updating hook registrations for team: $rite_name"

    # Ensure settings file exists with valid JSON
    if [[ ! -f "$settings_file" ]]; then
        echo '{}' > "$settings_file"
    fi

    # Validate JSON before proceeding
    if ! jq empty "$settings_file" 2>/dev/null; then
        log_error "Invalid JSON in $settings_file"
        return 1
    fi

    # Step 1: Extract non-roster hooks for preservation
    local preserved_hooks
    preserved_hooks=$(extract_non_roster_hooks "$settings_file")
    log_debug "Preserved $(echo "$preserved_hooks" | jq '[.[] | length] | add // 0') non-roster hook entries"

    # Step 2: Parse base hooks
    local base_registrations=""
    if [[ -f "$base_hooks_yaml" ]]; then
        base_registrations=$(parse_hooks_yaml "$base_hooks_yaml")
        local base_count
        base_count=$(echo "$base_registrations" | grep -c '^{' || echo 0)
        log_debug "Parsed $base_count base hook registrations"
    else
        log_warning "Base hooks file not found: $base_hooks_yaml"
    fi

    # Step 3: Parse team hooks (optional)
    local team_registrations=""
    if [[ -f "$team_hooks_yaml" ]]; then
        team_registrations=$(parse_hooks_yaml "$team_hooks_yaml")
        local team_count
        team_count=$(echo "$team_registrations" | grep -c '^{' || echo 0)
        log_debug "Parsed $team_count team hook registrations"
    else
        log_debug "No team hooks.yaml for $rite_name"
    fi

    # Step 4: Merge registrations
    local merged_registrations
    merged_registrations=$(merge_hook_registrations "$base_registrations" "$team_registrations")

    # Step 5: Generate hooks JSON
    local generated_hooks
    generated_hooks=$(generate_hooks_json "$merged_registrations")

    # Step 6: Merge with preserved hooks
    local final_hooks
    final_hooks=$(merge_with_preserved "$generated_hooks" "$preserved_hooks")

    # Step 7: Update settings.local.json
    local temp_file="${settings_file}.tmp"
    jq --argjson hooks "$final_hooks" '.hooks = $hooks' "$settings_file" > "$temp_file"

    if jq empty "$temp_file" 2>/dev/null; then
        mv "$temp_file" "$settings_file"
        log "Updated hook registrations in settings.local.json"
    else
        rm -f "$temp_file"
        log_error "Generated invalid JSON, hook registrations not updated"
        return 1
    fi
}
```

### Integration Point in `do_swap()`

Location: Line ~3070 in `swap-rite.sh`, after `swap_hooks "$rite_name"`

```bash
    # Sync team hooks
    swap_hooks "$rite_name"

    # NEW: Update hook registrations in settings.local.json
    swap_hook_registrations "$rite_name"
```

---

## Edge Cases

### Edge Case: Timeout > 60s

**Behavior**: Clamp to 60 seconds with warning

```bash
# In parse_hooks_yaml()
if [[ "$timeout" -gt 60 ]]; then
    log_warning "Timeout $timeout exceeds Claude Code limit of 60s, clamping to 60 (hook: $path)"
    timeout=60
fi
```

### Edge Case: Invalid Matcher

**Behavior**: Skip hook with warning, continue processing

```bash
# In parse_hooks_yaml()
if [[ -n "$matcher" ]]; then
    if ! echo "" | grep -E "$matcher" >/dev/null 2>&1; then
        log_warning "Invalid matcher regex: $matcher (skipping hook: $path)"
        continue
    fi
fi
```

### Edge Case: Missing Hook File

**Behavior**: Warn but still register (file may be synced separately)

```bash
# In parse_hooks_yaml() - soft validation
local hook_path=".claude/hooks/$path"
if [[ ! -f "$hook_path" ]]; then
    log_debug "Hook file not found yet: $hook_path (may be synced later)"
fi
```

### Edge Case: Preserve Non-Roster Hooks

**Behavior**: Identify by command path pattern, preserve in output

Detection criteria:
- Command does NOT contain `.claude/hooks/`
- OR command is an absolute path not under `.claude/hooks/`

### Edge Case: Team Hook Same Name as Base

**Behavior**: Both run (append, not override). Order: base hooks first, team hooks second.

```yaml
# If both base_hooks.yaml and team hooks.yaml have:
# - event: PreToolUse
#   matcher: "Bash"
#   path: command-validator.sh

# Result: command-validator.sh runs twice with same matcher
# This is intentional - allows team to add additional validation
```

### Edge Case: Missing base_hooks.yaml

**Behavior**: Warn and continue with team hooks only (or empty)

```bash
if [[ ! -f "$base_hooks_yaml" ]]; then
    log_warning "Base hooks file not found: $base_hooks_yaml"
    base_registrations=""
fi
```

### Edge Case: Corrupted settings.local.json

**Behavior**: Backup, attempt repair, or fail gracefully

```bash
if ! jq empty "$settings_file" 2>/dev/null; then
    log_error "Invalid JSON in $settings_file, backing up and creating fresh"
    mv "$settings_file" "${settings_file}.corrupt.$(date +%s)"
    echo '{}' > "$settings_file"
fi
```

---

## Dry Run Support

### Implementation

```bash
swap_hook_registrations() {
    local rite_name="$1"

    # ... parsing and merging ...

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "Hook registrations that would be written:"
        echo "$final_hooks" | jq '.'
        return 0
    fi

    # ... actual file write ...
}
```

### Example Output

```
$ ./swap-rite.sh --dry-run security-pack

Hook registrations that would be written:
{
  "SessionStart": [
    {
      "matcher": "startup|resume",
      "hooks": [
        { "type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/session-context.sh", "timeout": 10 },
        { "type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/coach-mode.sh", "timeout": 5 }
      ]
    }
  ],
  "PreToolUse": [
    {
      "matcher": "Edit|Write",
      "hooks": [
        { "type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/session-write-guard.sh", "timeout": 3 },
        { "type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/credential-scanner.sh", "timeout": 10 }
      ]
    }
  ]
}
```

---

## Test Strategy

### Unit Tests

Location: `tests/unit/hook-registration.bats`

| Test ID | Description | Expected Result |
|---------|-------------|-----------------|
| `hr_001` | Parse valid base_hooks.yaml | All hooks extracted correctly |
| `hr_002` | Parse hooks.yaml with invalid event | Invalid event skipped with warning |
| `hr_003` | Parse hooks.yaml with missing matcher for PreToolUse | Hook skipped with warning |
| `hr_004` | Parse hooks.yaml with timeout > 60 | Timeout clamped to 60 |
| `hr_005` | Parse hooks.yaml with invalid regex matcher | Hook skipped with warning |
| `hr_006` | Merge base + team hooks | Team hooks appended after base |
| `hr_007` | Generate JSON from registrations | Valid Claude Code format |
| `hr_008` | Preserve non-roster hooks | User hooks retained in output |
| `hr_009` | Handle missing base_hooks.yaml | Empty base, team-only output |
| `hr_010` | Handle missing team hooks.yaml | Base-only output |

### Integration Tests

Location: `tests/integration/hook-registration.bats`

| Test ID | Description | Expected Result |
|---------|-------------|-----------------|
| `int_hr_001` | Full swap with hook registration | settings.local.json updated |
| `int_hr_002` | Swap preserves user hooks | Non-roster hooks retained |
| `int_hr_003` | Dry run shows hook preview | Preview output, no file change |
| `int_hr_004` | Swap to team without hooks.yaml | Uses base hooks only |
| `int_hr_005` | Swap to team with hooks.yaml | Merges base + team |

### Example Test

```bash
@test "hr_006: Merge base + team hooks" {
    # Setup
    local base_yaml="$BATS_TMPDIR/base_hooks.yaml"
    local team_yaml="$BATS_TMPDIR/team_hooks.yaml"

    cat > "$base_yaml" <<'YAML'
schema_version: "1.0"
hooks:
  - event: PreToolUse
    matcher: "Bash"
    path: base-validator.sh
    timeout: 5
YAML

    cat > "$team_yaml" <<'YAML'
schema_version: "1.0"
hooks:
  - event: PreToolUse
    matcher: "Bash"
    path: team-validator.sh
    timeout: 10
YAML

    # Act
    local base_regs team_regs merged
    base_regs=$(parse_hooks_yaml "$base_yaml")
    team_regs=$(parse_hooks_yaml "$team_yaml")
    merged=$(merge_hook_registrations "$base_regs" "$team_regs")

    # Assert: Both hooks present, base first
    local count
    count=$(echo "$merged" | grep -c '^{')
    [ "$count" -eq 2 ]

    # Assert: Order is base then team
    local first_path last_path
    first_path=$(echo "$merged" | head -1 | jq -r '.path')
    last_path=$(echo "$merged" | tail -1 | jq -r '.path')
    [ "$first_path" = "base-validator.sh" ]
    [ "$last_path" = "team-validator.sh" ]
}
```

---

## Implementation Guidance

### Recommended Implementation Order

1. **Phase 1: YAML Schema and Parser**
   - Create `base_hooks.yaml` with current hook registrations
   - Implement `parse_hooks_yaml()` function
   - Add validation for all edge cases

2. **Phase 2: JSON Generator**
   - Implement `generate_hooks_json()`
   - Implement `generate_event_hooks_json()`
   - Test output matches Claude Code format

3. **Phase 3: Preservation Logic**
   - Implement `extract_non_roster_hooks()`
   - Implement `merge_with_preserved()`
   - Test user hook preservation

4. **Phase 4: Integration**
   - Implement `swap_hook_registrations()`
   - Integrate into `swap-rite.sh` after `swap_hooks()`
   - Add dry-run support

5. **Phase 5: Testing**
   - Unit tests for all functions
   - Integration tests for full swap flow
   - Manual testing with real rites

### Dependencies

| Dependency | Purpose | Installation |
|------------|---------|--------------|
| `yq` | YAML parsing | `brew install yq` or `pip install yq` |
| `jq` | JSON manipulation | `brew install jq` |

### Fallback if yq unavailable

```bash
# Minimal YAML parsing without yq (for bootstrapping)
parse_yaml_field() {
    local file="$1"
    local field="$2"
    grep "^${field}:" "$file" | sed 's/^[^:]*: *//' | tr -d '"'
}
```

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| yq not installed | Medium | Medium | Document requirement, provide fallback |
| Corrupt settings.local.json | Low | High | Backup before modification, validate after write |
| User hooks incorrectly identified | Low | Medium | Conservative detection (only preserve non-.claude/hooks paths) |
| Race condition with concurrent swaps | Low | Medium | File locking around settings modification |
| Team hooks.yaml syntax errors | Medium | Low | Validate YAML, skip invalid entries with warning |

---

## ADRs

This design does not introduce new architectural decisions requiring ADRs. It implements FR-2.1-2.5 from the PRD using established patterns from `swap-rite.sh`.

---

## Open Items

| Item | Status | Owner | Notes |
|------|--------|-------|-------|
| yq version compatibility | Open | Principal Engineer | Test with yq v4 and v3 |
| Settings file locking | Deferred | Future | Add flock for concurrent access |
| Hook validation CLI | Optional | Future | `just hooks:validate` task |

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-hook-parity-scope2.md` | Created |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` | Read |
| Current settings.local.json | `/Users/tomtenuta/Code/roster/.claude/settings.local.json` | Read |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh` | Read |
| TDD-session-state-machine.md (template) | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-state-machine.md` | Read |
| workflow-schema.yaml | `/Users/tomtenuta/Code/roster/workflow-schema.yaml` | Read |
