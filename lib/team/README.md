# lib/team - Team Swap Module Library

This directory contains extracted modules from `swap-team.sh`, providing reusable infrastructure for team resource management, transaction safety, and hook registration.

## Module Overview

The `lib/team/` directory is the extraction target for three major refactorings (Priority 1-3) from the `SPIKE-script-code-smell-refactoring.md` analysis. These modules consolidate ~1400 lines of duplicated and complex logic into ~1432 lines of well-tested, self-contained modules.

**Purpose**: Separate infrastructure concerns from orchestration logic in `swap-team.sh`, enabling:
- Independent testing of resource operations, transaction safety, and hook management
- Reuse across multiple team swap scenarios
- Clearer separation between "what to do" (orchestration) and "how to do it" (infrastructure)

## Architecture

```
swap-team.sh (orchestration)
    |
    +-- sources --> lib/team/team-resource.sh (resource operations)
    |                  |-- backup_team_resource()
    |                  |-- remove_team_resource()
    |                  |-- detect_resource_orphans()
    |                  |-- remove_resource_orphans()
    |                  |-- is_resource_from_team()
    |                  +-- get_resource_team()
    |
    +-- sources --> lib/team/team-transaction.sh (transaction safety)
    |                  |-- write_atomic()
    |                  |-- create_journal() / update_journal_*()
    |                  |-- create_staging() / stage_*() / verify_staging()
    |                  +-- create_swap_backup() / verify_backup_integrity()
    |
    +-- sources --> lib/team/team-hooks-registration.sh (hook registration)
                       |-- require_yq()
                       |-- parse_hooks_yaml()
                       |-- extract_non_roster_hooks()
                       |-- generate_hooks_json()
                       +-- swap_hook_registrations()
```

### Source Order in swap-team.sh

```bash
source "$ROSTER_HOME/lib/roster-utils.sh"
source "$ROSTER_HOME/lib/team/team-transaction.sh"    # Transaction infrastructure first
source "$ROSTER_HOME/lib/team/team-resource.sh"       # Resource operations second
source "$ROSTER_HOME/lib/team/team-hooks-registration.sh"  # Hook registration third
```

## Modules

### team-resource.sh (347 LOC)

**Purpose**: Generic resource operations for commands, skills, and hooks.

**Extracted From**: Priority 1 refactoring (SM-001 through SM-005 DRY violations)

**Consolidation**: 6 resource-type-specific patterns → 5 generic functions (6x DRY improvement, ~250 net LOC reduction)

#### Public API

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `is_resource_from_team` | resource_name, resource_type, find_type | 0 if from team, 1 otherwise | Check if resource belongs to any team pack |
| `get_resource_team` | resource_name, resource_type, find_type | team name (stdout) | Get team name that owns a resource |
| `backup_team_resource` | resource_type, resource_dir, marker_file, find_type | 0 on success | Backup team-owned resources to .backup directory |
| `remove_team_resource` | resource_type, resource_dir, marker_file, find_type | 0 on success | Remove team-owned resources listed in marker file |
| `detect_resource_orphans` | resource_type, resource_dir, incoming_team, find_type, glob_pattern | orphan list (stdout) | Detect orphaned resources from other teams |
| `remove_resource_orphans` | resource_type, resource_dir, orphan_mode, find_type (+ stdin) | 0 on success | Remove orphaned resources with backup |

#### Usage Example

```bash
# Source the module
source "$ROSTER_HOME/lib/team/team-resource.sh"

# Backup team commands before swap
backup_team_resource "commands" ".claude/commands" ".team-commands" "f"

# Remove old team commands
remove_team_resource "commands" ".claude/commands" ".team-commands" "f"

# Detect and remove orphaned skills
detect_resource_orphans "skills" ".claude/skills" "hygiene-pack" "d" "*/" \
    | remove_resource_orphans "skills" ".claude/skills" "remove" "d"
```

#### Key Design Decisions

- **Parameterized by resource type**: Single implementation serves commands (files), skills (directories), and hooks (files)
- **Stdout-based data flow**: Uses stdout instead of global arrays for bash 3.2 portability
- **Marker file pattern**: Resources tracked via `.team-commands`, `.team-skills`, `.team-hooks` files
- **Logging stubs**: Provides basic logging when used standalone, overridden by swap-team.sh

---

### team-transaction.sh (548 LOC)

**Purpose**: Transaction infrastructure for atomic team swaps with rollback capability.

**Extracted From**: Priority 2 refactoring (lines 100-534 in swap-team.sh, ~300 LOC)

**Consolidation**: Atomic writes, journal CRUD, staging, and backup operations into single module

#### Public API

##### Atomic I/O

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `write_atomic` | target, content | 0 on success | Write content atomically using temp file + rename pattern |

##### Journal Operations

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `create_journal` | source_team, target_team | 0 on success, 1 if exists | Create new journal entry for swap operation |
| `update_journal_phase` | new_phase | 0 on success | Update journal phase (PREPARING, BACKING, etc.) |
| `update_journal_backups` | resource_type, backup_path | 0 on success | Update journal backup locations for resources |
| `update_journal_error` | error_msg | 0 on success | Record error message in journal |
| `get_journal_field` | field | field value (stdout) | Read arbitrary journal field |
| `get_journal_phase` | - | phase name (stdout) | Get current journal phase |
| `delete_journal` | - | 0 always | Delete journal (on successful completion) |
| `journal_exists` | - | 0 if exists | Check if journal exists |

##### Staging Operations

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `create_staging` | - | 0 on success | Create staging directory structure |
| `cleanup_staging` | - | 0 always | Clean up staging directory |
| `stage_agents` | team_name | 0 on success | Stage agents from team pack |
| `stage_workflow` | team_name | 0 on success | Stage workflow.yaml file |
| `stage_active_team` | team_name | 0 on success | Stage ACTIVE_TEAM file |
| `verify_staging` | expected_count | 0 on success | Verify staging directory integrity |

##### Backup Operations

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `create_swap_backup` | - | 0 on success | Create comprehensive backup for transaction safety |
| `cleanup_swap_backup` | - | 0 always | Clean up swap backup (after successful swap) |
| `verify_backup_integrity` | - | 0 if valid | Verify backup integrity for recovery |

#### Usage Example

```bash
# Source the module
source "$ROSTER_HOME/lib/team/team-transaction.sh"

# Transaction flow example
create_journal "$current_team" "$target_team"
update_journal_phase "$PHASE_BACKING"

create_swap_backup
update_journal_phase "$PHASE_STAGING"

create_staging
stage_agents "$target_team"
stage_active_team "$target_team"
verify_staging "$expected_agent_count"
update_journal_phase "$PHASE_COMMITTING"

# ... commit changes ...

delete_journal
cleanup_swap_backup
```

#### Journal Schema

```json
{
  "version": "1.0",
  "started_at": "ISO8601 timestamp",
  "phase": "PREPARING|BACKING|STAGING|VERIFYING|COMMITTING|COMPLETED",
  "source_team": "string|null",
  "target_team": "string",
  "backup_location": {
    "agents": "path|null",
    "manifest": "path|null",
    "active_team": "path|null",
    "workflow": "path|null",
    "commands": "path|null",
    "skills": "path|null",
    "hooks": "path|null"
  },
  "staging_location": "path",
  "checksums": {},
  "pid": "integer",
  "error": "string|null"
}
```

#### Key Design Decisions

- **Atomic writes**: All file modifications use temp file + rename pattern
- **Journal-based recovery**: Tracks swap progress for interrupt/failure recovery
- **Phase tracking**: Six phases enable granular rollback decisions
- **Default constants**: Module provides defaults, caller can override before sourcing
- **Concurrent swap prevention**: Journal existence blocks concurrent swap attempts

---

### team-hooks-registration.sh (537 LOC)

**Purpose**: Hook registration infrastructure for settings.local.json management.

**Extracted From**: Priority 3 refactoring (lines 2136-2565 in swap-team.sh, ~430 LOC)

**Consolidation**: YAML parsing, JSON generation, and user hook preservation into single module

#### Public API

##### Validation

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `require_yq` | - | 0 if yq v4+ available | Check if yq v4+ is installed |

##### YAML Parsing

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `parse_hooks_yaml` | yaml_file | JSON-lines (stdout) | Parse hooks.yaml and emit JSON-lines format |

##### JSON Operations

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `extract_non_roster_hooks` | settings_file | JSON object (stdout) | Extract non-roster hooks from settings.local.json |
| `generate_hooks_json` | registrations (JSON-lines) | Claude hooks JSON (stdout) | Generate Claude Code hooks JSON format |
| `merge_hook_registrations` | base_registrations, team_registrations | JSON-lines (stdout) | Merge hook registrations (base first, team appended) |
| `merge_with_preserved` | generated_json, preserved_json | JSON object (stdout) | Merge generated hooks with preserved user hooks |

##### Main Orchestrator

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `swap_hook_registrations` | team_name | 0 on success | Sync hook registrations to settings.local.json |

#### Usage Example

```bash
# Source the module
source "$ROSTER_HOME/lib/team/team-hooks-registration.sh"

# Update hook registrations for active team
swap_hook_registrations "hygiene-pack"
```

#### hooks.yaml Schema

```yaml
schema_version: "1.0"  # Optional, warns if not 1.0
hooks:
  - event: SessionStart|Stop|PreToolUse|PostToolUse|UserPromptSubmit
    matcher: "regex"  # Required for PreToolUse/PostToolUse
    path: "relative/path/to/hook.sh"  # Required
    timeout: 5  # Optional, 1-60, default 5
```

#### Key Design Decisions

- **YAML source of truth**: Hook definitions in human-readable hooks.yaml files
- **User hook preservation**: Detects and preserves hooks NOT from roster (commands not containing `.claude/hooks/`)
- **Base + team layering**: Base hooks apply to all teams, team hooks override/extend
- **Validation and sanitization**: Event type validation, regex syntax checking, timeout clamping
- **Dry-run support**: Preview changes without writing via `DRY_RUN_MODE=1`
- **Atomic updates**: Uses temp file + rename pattern for settings.local.json updates
- **Corrupted JSON recovery**: Backs up invalid settings.local.json before recreating

---

## Testing

### Test Suites

| Module | Test File | LOC | Test Count |
|--------|-----------|-----|------------|
| team-resource.sh | `tests/lib/team/test-team-resource.sh` | - | 14 unit tests |
| team-transaction.sh | `tests/lib/team/test-team-transaction.sh` | - | 28 unit tests |
| team-hooks-registration.sh | `tests/lib/team/test-team-hooks-registration.sh` | - | 22 unit tests |

### Running Tests

```bash
# Run all lib/team tests
./tests/lib/team/test-team-resource.sh
./tests/lib/team/test-team-transaction.sh
./tests/lib/team/test-team-hooks-registration.sh

# Run specific test function
./tests/lib/team/test-team-resource.sh test_backup_team_resource_commands
```

### Test Fixtures

```
tests/fixtures/
  team-resource/
    mock-teams/           # Mock team structures
    mock-project/         # Mock .claude/ directories
  team-transaction/
    mock-teams/
    mock-project/
  team-hooks-registration/
    valid-hooks.yaml
    invalid-event.yaml
    settings-roster-only.json
    settings-mixed.json
```

## Dependencies

### External Tools

- **jq**: JSON manipulation (required by all modules)
- **yq v4+**: YAML parsing (required by team-hooks-registration.sh only)
- **find, grep, sed, awk**: Standard POSIX utilities

### Internal Dependencies

All modules require logging functions from swap-team.sh:
- `log()` - Standard logging
- `log_debug()` - Debug logging
- `log_warning()` - Warning logging
- `log_error()` - Error logging

Modules provide stub implementations for standalone use (e.g., in unit tests).

### Environment Variables

- `ROSTER_HOME`: Path to roster installation (required by all modules)
- `DRY_RUN_MODE`: If set to 1, preview changes without writing (used by team-hooks-registration.sh)

## Compatibility

- **Bash version**: 3.2+ (macOS default bash compatibility)
- **No bashisms**: Portable across Linux/macOS
- **No associative arrays**: Uses stdout-based data flow instead
- **No nameref**: Uses function return values and stdout

## Refactoring Documentation

For detailed refactoring plans and architectural decisions, see:

- `docs/refactoring/REFACTOR-team-resource.md` - Priority 1 extraction plan
- `docs/refactoring/REFACTOR-team-transaction.md` - Priority 2 extraction plan
- `docs/refactoring/REFACTOR-team-hooks-registration.md` - Priority 3 extraction plan
- `docs/spikes/SPIKE-script-code-smell-refactoring.md` - Original code smell analysis

## Future Work

Potential future extractions from swap-team.sh:

- **rollback and recovery operations** - Currently orchestration logic mixed with recovery policy
- **CLAUDE.md update logic** - update_claude_md() complexity (CH-002)
- **Hook file sync** - swap_hooks() file copying logic (CH-003, separate from registration)
- **Resource swap orchestration** - swap_commands/skills/hooks consolidation (SM-006)

See `docs/spikes/SPIKE-script-code-smell-refactoring.md` for full analysis.
