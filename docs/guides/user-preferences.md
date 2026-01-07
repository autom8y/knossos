# User Preferences Guide

> Configure how Claude Code behaves in your Roster projects

## Overview

User preferences allow you to customize Claude Code's behavior to match your workflow. These settings persist across sessions and control aspects like:

- **Autonomy level**: How much Claude asks before taking action
- **Failure handling**: What happens when operations fail
- **Output verbosity**: How detailed Claude's explanations are
- **Git behavior**: Auto-push, auto-PR, and commit requirements
- **Orchestration mode**: How multi-agent workflows coordinate

Preferences are stored in `.claude/user-preferences.json` and loaded automatically at session start.

## Quick Start

### First-Run Experience

When you start a new project (or clone an existing one) without preferences, Roster automatically:

1. Detects no `user-preferences.json` exists
2. Creates the file with sensible defaults
3. Shows a first-run message explaining key settings

You'll see this message in your session context:

```
## First-Run Setup

No user preferences found. Creating defaults at `.claude/user-preferences.json`.

**Key preferences set**:
- Autonomy: `interactive` (asks before major actions)
- Failure handling: `ask` (prompts on errors)
- Output: `verbose` (detailed explanations)
```

### Where the File Lives

Your preferences file is always at:

```
<project-root>/.claude/user-preferences.json
```

This file is project-local, so you can have different preferences per project. It's recommended to add this file to `.gitignore` if you don't want to share preferences with collaborators.

## Configuration Reference

### Core Preferences

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `version` | string | `"1.0.0"` | Schema version (required) |
| `autonomy_level` | enum | `"interactive"` | How much Claude asks before acting |
| `failure_handling` | enum | `"ask"` | What happens when operations fail |
| `output_format` | enum | `"verbose"` | Detail level in explanations |

#### autonomy_level

Controls how much Claude asks before taking action:

| Value | Behavior |
|-------|----------|
| `"auto"` | Full autonomy - Claude proceeds without asking |
| `"interactive"` | Asks before major actions (default) |
| `"manual"` | Explicit approval required for most actions |

#### failure_handling

Controls what happens when operations fail:

| Value | Behavior |
|-------|----------|
| `"ask"` | Prompts you to decide next steps (default) |
| `"rollback"` | Automatically reverts changes on failure |
| `"continue"` | Logs the error and proceeds with remaining tasks |

#### output_format

Controls verbosity of Claude's explanations:

| Value | Behavior |
|-------|----------|
| `"verbose"` | Detailed explanations and reasoning (default) |
| `"terse"` | Minimal output, just the essentials |

### Git Preferences

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `commit_auto_push` | boolean | `false` | Auto-push after commits |
| `pr_auto_create` | boolean | `false` | Auto-create PR after feature completion |
| `test_before_commit` | boolean | `true` | Require tests before commits |
| `default_branch` | string | `"main"` | Default branch name for new repos |

### Session Preferences

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `session_auto_park` | boolean | `false` | Auto-park session on context switch |
| `orchestration_mode` | enum | `"task_tool"` | Multi-agent coordination style |
| `artifact_verification` | enum | `"always"` | File verification frequency |
| `notification_level` | enum | `"errors"` | Background operation notifications |

#### orchestration_mode

Controls how multi-agent workflows coordinate:

| Value | Behavior |
|-------|----------|
| `"task_tool"` | Delegates via Task tool (default) |
| `"coach"` | Provides guidance only, you execute |
| `"direct"` | No orchestration, direct execution |

#### artifact_verification

Controls file verification after writes:

| Value | Behavior |
|-------|----------|
| `"always"` | Verify all file writes (default) |
| `"on_error"` | Verify only when errors occur |
| `"never"` | Skip verification (faster but riskier) |

### Editor Integration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `editor_integration.auto_open_files` | boolean | `false` | Auto-open modified files in editor |
| `editor_integration.preserve_cursor_position` | boolean | `true` | Remember cursor position across sessions |

### Team Preferences

Team-specific overrides for workflow customization:

```json
"team_preferences": {
  "ecosystem": {
    "preferred_agent": "integration-engineer",
    "auto_activate": true
  }
}
```

### Custom Workflows

User-defined workflow triggers and actions:

```json
"custom_workflows": {
  "pre_commit_lint": {
    "enabled": true,
    "trigger_pattern": "^(feat|fix|docs):",
    "action": "run_linter"
  }
}
```

## Customization

### Manual Editing

Edit `.claude/user-preferences.json` directly:

```bash
# Open in your editor
code .claude/user-preferences.json

# Or with vim
vim .claude/user-preferences.json
```

### Example Configurations

#### High Autonomy Developer

For experienced developers who want Claude to move fast:

```json
{
  "version": "1.0.0",
  "autonomy_level": "auto",
  "failure_handling": "continue",
  "output_format": "terse",
  "commit_auto_push": true,
  "pr_auto_create": true,
  "test_before_commit": true,
  "artifact_verification": "on_error"
}
```

#### Cautious Reviewer

For careful, review-focused work:

```json
{
  "version": "1.0.0",
  "autonomy_level": "manual",
  "failure_handling": "rollback",
  "output_format": "verbose",
  "commit_auto_push": false,
  "pr_auto_create": false,
  "test_before_commit": true,
  "artifact_verification": "always"
}
```

#### CI/CD Pipeline

For automated environments:

```json
{
  "version": "1.0.0",
  "autonomy_level": "auto",
  "failure_handling": "rollback",
  "output_format": "terse",
  "notification_level": "errors",
  "artifact_verification": "always"
}
```

## For Hook Authors

### Using preferences-loader.sh

The `preferences-loader.sh` library provides functions for accessing preferences in custom hooks.

#### Sourcing the Library

```bash
#!/bin/bash
HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/preferences-loader.sh"

# Load preferences (idempotent, caches results)
load_user_preferences
```

#### Getting Preferences

```bash
# Get a preference value (returns default if not set)
autonomy=$(get_preference "autonomy_level")

# Get nested preferences with dotted path
auto_open=$(get_preference "editor_integration.auto_open_files")

# Check boolean preferences
if is_preference_enabled "test_before_commit"; then
    run_tests
fi

# Get boolean with explicit true/false string
commit_push=$(get_preference_bool "commit_auto_push")
```

#### Exporting to Environment

```bash
# Export all preferences as ROSTER_PREF_* variables
export_preferences_env

# Now available as:
echo "$ROSTER_PREF_AUTONOMY_LEVEL"
echo "$ROSTER_PREF_FAILURE_HANDLING"
echo "$ROSTER_PREF_EDITOR_INTEGRATION_AUTO_OPEN_FILES"
```

#### Available Functions

| Function | Description |
|----------|-------------|
| `load_user_preferences` | Load and cache preferences from file |
| `get_preference <key>` | Get preference value (supports dotted paths) |
| `get_preference_bool <key>` | Get preference as "true" or "false" string |
| `is_preference_enabled <key>` | Returns 0 if true, 1 if false |
| `export_preferences_env` | Export as ROSTER_PREF_* environment variables |
| `validate_preferences` | Validate current preferences against schema |
| `is_first_run` | Returns 0 if no preferences file exists |
| `create_default_preferences` | Create preferences file with defaults |
| `reset_preferences_cache` | Clear cached preferences (for reload) |

#### Example: Custom Hook with Preferences

```bash
#!/bin/bash
# my-custom-hook.sh - Example hook using preferences

HOOKS_LIB="${CLAUDE_PROJECT_DIR:-.}/.claude/hooks/lib"
source "$HOOKS_LIB/hooks-init.sh"
hooks_init "my-custom-hook" "RECOVERABLE"

# Source preferences (graceful if missing)
source "$HOOKS_LIB/preferences-loader.sh" 2>/dev/null || true

# Load preferences
load_user_preferences 2>/dev/null || true

# Use preferences in hook logic
output_format=$(get_preference "output_format")

if [[ "$output_format" == "verbose" ]]; then
    echo "## Detailed Hook Output"
    echo "Running with full verbosity..."
else
    echo "Hook: OK"
fi

# Check boolean preference
if is_preference_enabled "test_before_commit"; then
    echo "Tests required before commit"
fi
```

## Troubleshooting

### Invalid JSON

**Symptom**: Preferences not loading, defaults used instead.

**Diagnosis**:
```bash
# Validate JSON syntax
jq . .claude/user-preferences.json
```

**Fix**: Correct the JSON syntax error. Common issues:
- Missing commas between properties
- Trailing commas after last property
- Unquoted string values

### Missing File

**Symptom**: First-run message appears every session.

**Diagnosis**:
```bash
ls -la .claude/user-preferences.json
```

**Fix**: The file should auto-create on first run. If not:
```bash
# Manual creation with defaults
cat > .claude/user-preferences.json << 'EOF'
{
  "version": "1.0.0",
  "autonomy_level": "interactive",
  "failure_handling": "ask",
  "output_format": "verbose"
}
EOF
```

### Preferences Not Taking Effect

**Symptom**: Changed preferences but behavior unchanged.

**Diagnosis**:
1. Check session context output shows new values
2. Verify JSON is valid
3. Confirm correct key names (case-sensitive)

**Fix**: Start a new session. Preferences load at session start, not mid-session.

### Permission Errors

**Symptom**: "Failed to write default preferences" errors.

**Diagnosis**:
```bash
ls -la .claude/
```

**Fix**:
```bash
# Ensure .claude directory is writable
chmod 755 .claude
```

### Validation Warnings

**Symptom**: "Preferences validation: invalid X" warnings in logs.

**Diagnosis**: A preference has an invalid value for its enum type.

**Fix**: Check the Configuration Reference above for valid values. Example:
```json
// Wrong
"autonomy_level": "automatic"

// Correct
"autonomy_level": "auto"
```

## Schema Reference

The full JSON Schema is at `.claude/user-preferences.schema.json`. Key constraints:

- `version` is required and must match semver pattern (`X.Y.Z`)
- Enum fields only accept listed values
- Boolean fields must be `true` or `false` (not strings)
- `additionalProperties: false` means unknown keys are rejected

## Related Documentation

- [Session Lifecycle](./session-lifecycle.md) - How sessions work with preferences
- [Hook Ecosystem](../reference/hook-exit-codes.md) - Hook development patterns
- [Orchestration](../design/TDD-hybrid-session-model.md) - Multi-agent coordination modes
