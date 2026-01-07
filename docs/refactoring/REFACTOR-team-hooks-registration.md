# Refactoring Plan: lib/team/rite-hooks-registration.sh Extraction

**Based on**: SPIKE-script-code-smell-refactoring.md
**Prepared**: 2026-01-03
**Scope**: Extract hook registration operations from swap-rite.sh to lib/team/rite-hooks-registration.sh
**Priority**: 3 (following rite-resource.sh and team-transaction.sh extraction)

---

## Executive Summary

This plan extracts hook registration infrastructure from `swap-rite.sh` into a new `lib/team/rite-hooks-registration.sh` module. The extraction consolidates ~430 lines (lines 2136-2565) of YAML parsing, JSON generation, and settings.local.json manipulation into a self-contained, testable module.

**Expected outcome**: Clean module boundary, testable hook registration logic, ~430 LOC extraction, no behavior change.

---

## Architectural Assessment

### Boundary Health

| Module | Assessment |
|--------|------------|
| `swap-rite.sh` (main script) | swap_hooks() orchestration stays; registration logic extracts |
| `lib/team/rite-resource.sh` | Extracted in Priority 1; provides pattern to follow |
| `lib/team/team-transaction.sh` | Extracted in Priority 2; provides sourcing guard pattern |
| `lib/team/rite-hooks-registration.sh` (new) | Will establish clean boundary for hook registration infrastructure |

### Root Causes Identified

1. **YAML/JSON manipulation mixed with orchestration**: The functions are pure data transformation with no side effects beyond the final file write. Perfect for module extraction.

2. **External tool dependency**: `require_yq()` validates yq v4+ presence. This dependency check can live in the module.

3. **User hook preservation**: `extract_non_roster_hooks()` implements important behavior to preserve user-added hooks. Must be tested thoroughly.

### Functions Scope (lines 2136-2565, ~430 LOC)

| Function | Line | LOC | Category |
|----------|------|-----|----------|
| `require_yq()` | 2136-2156 | 21 | Validation |
| `parse_hooks_yaml()` | 2161-2247 | 87 | YAML Parsing |
| `extract_non_roster_hooks()` | 2253-2312 | 60 | JSON Extraction |
| `merge_hook_registrations()` | 2318-2324 | 7 | Data Merge |
| `generate_hooks_json()` | 2330-2418 | 89 | JSON Generation |
| `merge_with_preserved()` | 2423-2456 | 34 | Data Merge |
| `swap_hook_registrations()` | 2461-2565 | 105 | Orchestration |
| **Total** | | **~403** | |

---

## Dependency Analysis

### Call Graph

```
swap_hook_registrations()
  |-- require_yq()                      [Validation]
  |-- extract_non_roster_hooks()        [JSON Extraction]
  |-- parse_hooks_yaml()                [YAML Parsing - base]
  |-- parse_hooks_yaml()                [YAML Parsing - team]
  |-- merge_hook_registrations()        [Data Merge]
  |-- generate_hooks_json()             [JSON Generation]
  |-- merge_with_preserved()            [Data Merge]
  |-- [jq write to settings.local.json] [Side Effect]

parse_hooks_yaml()
  |-- [yq for YAML parsing]
  |-- [jq for JSON-lines output]

extract_non_roster_hooks()
  |-- [jq for JSON manipulation]

generate_hooks_json()
  |-- [jq for JSON construction]
  |-- [awk for unique matchers]

merge_with_preserved()
  |-- [jq for JSON merging]
```

### Extraction Order

Based on dependencies, the extraction order is:

1. **require_yq()** - No dependencies, standalone validation
2. **parse_hooks_yaml()** - Depends only on yq and jq
3. **extract_non_roster_hooks()** - Depends only on jq
4. **merge_hook_registrations()** - Pure function, no dependencies
5. **generate_hooks_json()** - Depends only on jq
6. **merge_with_preserved()** - Depends only on jq
7. **swap_hook_registrations()** - Orchestrates all above

### External Dependencies

| Dependency | Type | Source |
|------------|------|--------|
| `ROSTER_HOME` | Global variable | Environment / swap-rite.sh |
| `DRY_RUN_MODE` | Global variable | swap-rite.sh (passed as behavior) |
| `yq` | External command | System (v4+ required) |
| `jq` | External command | System |
| `log()` | Function | swap-rite.sh |
| `log_debug()` | Function | swap-rite.sh |
| `log_warning()` | Function | swap-rite.sh |
| `log_error()` | Function | swap-rite.sh |

---

## Target API

### lib/team/rite-hooks-registration.sh Public Interface

```bash
#!/usr/bin/env bash
#
# rite-hooks-registration.sh - Hook Registration for settings.local.json
#
# Parses hooks.yaml files and generates Claude Code hook registrations
# in settings.local.json while preserving user-defined hooks.
#
# Part of: roster team-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/team/rite-hooks-registration.sh"
#   swap_hook_registrations "team-name"
#
# Dependencies:
#   - yq v4+ (for YAML parsing)
#   - jq (for JSON manipulation)
#   - Logging functions (log, log_debug, log_warning, log_error)
#
# Environment:
#   ROSTER_HOME - Path to roster installation
#   DRY_RUN_MODE - If set to 1, preview changes without writing

# Guard against re-sourcing
[[ -n "${_TEAM_HOOKS_REGISTRATION_LOADED:-}" ]] && return 0
readonly _TEAM_HOOKS_REGISTRATION_LOADED=1

# ============================================================================
# Validation
# ============================================================================

# Check if yq v4+ is available
# Returns: 0 if yq v4+ available, 1 otherwise
# Side effects: Logs error if not available
require_yq() { ... }

# ============================================================================
# YAML Parsing
# ============================================================================

# Parse hooks.yaml file and emit JSON-lines format
# Parameters:
#   $1 - yaml_file: Path to hooks.yaml file
# Output: One JSON object per line to stdout
#   Format: {"event":"...","matcher":"...","path":"...","timeout":N}
# Returns: 0 always (empty output for missing/invalid file)
# Side effects: Logs warnings for invalid entries
parse_hooks_yaml() { ... }

# ============================================================================
# JSON Extraction
# ============================================================================

# Extract non-roster hooks from existing settings.local.json
# These are hooks whose command does NOT contain ".claude/hooks/"
# Parameters:
#   $1 - settings_file: Path to settings.local.json
# Output: JSON object with preserved hooks by event type to stdout
# Returns: 0 always (empty {} for missing file)
extract_non_roster_hooks() { ... }

# ============================================================================
# Data Merge
# ============================================================================

# Merge hook registrations (base first, team appended)
# Parameters:
#   $1 - base_registrations: JSON-lines format (from base hooks)
#   $2 - team_registrations: JSON-lines format (from team hooks)
# Output: Combined JSON-lines to stdout (base first, then team)
# Returns: 0 always
merge_hook_registrations() { ... }

# Merge generated hooks with preserved user hooks
# Parameters:
#   $1 - generated_json: Generated hooks JSON object
#   $2 - preserved_json: Preserved user hooks JSON object
# Output: Combined hooks JSON object to stdout
# Returns: 0 always
merge_with_preserved() { ... }

# ============================================================================
# JSON Generation
# ============================================================================

# Generate Claude Code hooks JSON format from registrations
# Parameters:
#   $1 - registrations: JSON-lines format
# Output: Claude Code settings.local.json hooks object to stdout
# Returns: 0 always (empty {} for no registrations)
generate_hooks_json() { ... }

# ============================================================================
# Main Orchestrator
# ============================================================================

# Sync hook registrations to settings.local.json
# Called after swap_hooks() syncs the actual hook files
# Parameters:
#   $1 - rite_name: Name of team being activated
# Returns: 0 on success, 1 on error
# Side effects:
#   - Updates .claude/settings.local.json hooks section
#   - Preserves non-roster hooks in settings
#   - Creates settings.local.json if missing
#   - Backs up corrupted settings.local.json
# Environment:
#   ROSTER_HOME - Must be set
#   DRY_RUN_MODE - If 1, prints preview without writing
swap_hook_registrations() { ... }
```

---

## Refactoring Sequence

### Phase 1: Foundation [Low Risk]

**Goal**: Create module structure with sourcing guard and logging stubs.

#### RF-017: Create lib/team/rite-hooks-registration.sh skeleton

- **Smells addressed**: CH-004 (swap_hook_registrations complexity)
- **Category**: Local (infrastructure)
- **Before**: No `lib/team/rite-hooks-registration.sh` exists
- **After**: Module exists with:
  - Sourcing guard (`_TEAM_HOOKS_REGISTRATION_LOADED`)
  - Module header documentation
  - Logging stubs (same pattern as rite-resource.sh)
  - Empty function stubs with documentation
- **Invariants**: swap-rite.sh unchanged, still works
- **Verification**:
  1. Run: `bash -n lib/team/rite-hooks-registration.sh` (syntax check)
  2. Run: `./swap-rite.sh --help` (existing behavior unchanged)
- **Commit scope**: Create module skeleton with guards and stubs

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 2: Validation and Parsing [Low Risk]

**Goal**: Extract validation and YAML parsing functions.

#### RF-018: Extract require_yq()

- **Smells addressed**: External tool validation
- **Category**: Local
- **Before**:
  ```bash
  # swap-rite.sh:2136-2156
  require_yq() {
      if ! command -v yq &>/dev/null; then
          log_error "yq is required but not installed"
          ...
      }
  }
  ```
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Same version check logic (v4+ required)
  - Same error messages
  - Returns 0 on success, 1 on failure
- **Verification**:
  1. Unit test: Mock yq not installed, verify returns 1
  2. Unit test: Mock yq v3, verify returns 1
  3. Unit test: Mock yq v4+, verify returns 0
- **Commit scope**: Add `require_yq()`, add unit tests

#### RF-019: Extract parse_hooks_yaml()

- **Smells addressed**: YAML parsing logic
- **Category**: Local
- **Before**: Lines 2161-2247 in swap-rite.sh
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Same schema version validation (warns on non-1.0)
  - Same event type validation (SessionStart, Stop, PreToolUse, PostToolUse, UserPromptSubmit)
  - Same matcher requirement for PreToolUse/PostToolUse
  - Same regex validation for matchers
  - Same timeout clamping (1-60, default 5)
  - Same JSON-lines output format
- **Verification**:
  1. Unit test: Parse valid hooks.yaml, verify JSON-lines output
  2. Unit test: Parse file with invalid event, verify skipped with warning
  3. Unit test: Parse PreToolUse without matcher, verify skipped
  4. Unit test: Parse with timeout > 60, verify clamped
  5. Unit test: Parse with invalid regex matcher, verify skipped
  6. Unit test: Missing file returns empty
- **Commit scope**: Add `parse_hooks_yaml()`, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 3: JSON Operations [Low Risk]

**Goal**: Extract JSON extraction and generation functions.

#### RF-020: Extract extract_non_roster_hooks()

- **Smells addressed**: User hook preservation logic
- **Category**: Local
- **Before**: Lines 2253-2312 in swap-rite.sh
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Same detection logic (commands NOT containing ".claude/hooks/")
  - Same event type iteration
  - Returns empty {} for missing file
  - Same JSON structure in output
- **Verification**:
  1. Unit test: Extract from file with only roster hooks, verify {}
  2. Unit test: Extract from file with mixed hooks, verify user hooks preserved
  3. Unit test: Extract from file with only user hooks, verify all preserved
  4. Unit test: Missing file returns {}
  5. Unit test: Invalid JSON file returns {}
- **Commit scope**: Add `extract_non_roster_hooks()`, add unit tests

#### RF-021: Extract merge_hook_registrations()

- **Smells addressed**: Simple consolidation function
- **Category**: Local
- **Before**: Lines 2318-2324 in swap-rite.sh
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Base registrations first in output
  - Team registrations appended
  - Empty lines filtered
- **Verification**:
  1. Unit test: Merge base + team, verify order preserved
  2. Unit test: Merge with empty base, verify team only
  3. Unit test: Merge with empty team, verify base only
- **Commit scope**: Add `merge_hook_registrations()`, add unit tests

#### RF-022: Extract generate_hooks_json()

- **Smells addressed**: JSON generation logic
- **Category**: Local
- **Before**: Lines 2330-2418 in swap-rite.sh
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Same grouping by event type
  - Same grouping by matcher within event
  - Same hook object structure (type, command, timeout)
  - Same path prefix ($CLAUDE_PROJECT_DIR/.claude/hooks/)
  - Empty input returns {}
- **Verification**:
  1. Unit test: Generate from single hook, verify structure
  2. Unit test: Generate from multiple hooks same event, verify grouped
  3. Unit test: Generate from multiple matchers, verify separate entries
  4. Unit test: Generate with no matcher, verify entry without matcher field
  5. Unit test: Empty input returns {}
- **Commit scope**: Add `generate_hooks_json()`, add unit tests

#### RF-023: Extract merge_with_preserved()

- **Smells addressed**: Hook merging logic
- **Category**: Local
- **Before**: Lines 2423-2456 in swap-rite.sh
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Preserved hooks appended to generated for each event
  - Empty preserved returns generated unchanged
  - Same JSON structure
- **Verification**:
  1. Unit test: Merge with empty preserved, verify generated unchanged
  2. Unit test: Merge with preserved, verify both present
  3. Unit test: Merge with same event type, verify appended not replaced
- **Commit scope**: Add `merge_with_preserved()`, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 4: Orchestration [Medium Risk]

**Goal**: Extract the main orchestration function.

#### RF-024: Extract swap_hook_registrations()

- **Smells addressed**: CH-004 (complexity), module boundary
- **Category**: Module boundary
- **Before**: Lines 2461-2565 in swap-rite.sh
- **After**: Same function in rite-hooks-registration.sh
- **Invariants**:
  - Same 7-step process (require_yq, extract preserved, parse base, parse team, merge registrations, generate JSON, merge with preserved, write)
  - Same dry-run behavior (preview without write)
  - Same error handling for invalid JSON
  - Same backup for corrupted settings.local.json
  - Same atomic write pattern (temp file + mv)
  - Same logging messages
- **Verification**:
  1. Unit test: Full flow with mock files
  2. Unit test: Dry-run mode shows preview
  3. Unit test: Creates settings.local.json if missing
  4. Unit test: Backs up corrupted file
  5. Integration: Full swap with hook registration
- **Commit scope**: Add `swap_hook_registrations()`, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 5: Integration [Medium Risk]

**Goal**: Wire new module into swap-rite.sh, remove inline implementations.

#### RF-025: Add source statement in swap-rite.sh

- **Smells addressed**: Final integration
- **Category**: Boundary integration
- **Before**: All hook registration functions inline
- **After**:
  ```bash
  # swap-rite.sh (after lib/team/team-transaction.sh)
  source "$ROSTER_HOME/lib/team/rite-hooks-registration.sh"
  ```
- **Source order in swap-rite.sh**:
  ```bash
  source "$ROSTER_HOME/lib/roster-utils.sh"
  source "$ROSTER_HOME/lib/team/team-transaction.sh"
  source "$ROSTER_HOME/lib/team/rite-resource.sh"
  source "$ROSTER_HOME/lib/team/rite-hooks-registration.sh"  # NEW
  ```
- **Invariants**:
  - All existing function calls work unchanged
  - Logging functions available before sourcing
  - ROSTER_HOME set before sourcing
- **Verification**:
  1. Run: `./swap-rite.sh --help` (smoke test)
  2. Run: `./swap-rite.sh --dry-run hygiene` (full flow)
  3. Verify: settings.local.json updated correctly
  4. Verify: User hooks preserved
- **Commit scope**: Add source statement only

#### RF-026: Remove inline hook registration functions

- **Smells addressed**: Final cleanup, eliminate duplication
- **Category**: Local
- **Before**: 7 inline functions in swap-rite.sh (lines 2136-2565, ~430 LOC)
- **After**: Only source statement remains
- **Functions removed**:
  - `require_yq()`
  - `parse_hooks_yaml()`
  - `extract_non_roster_hooks()`
  - `merge_hook_registrations()`
  - `generate_hooks_json()`
  - `merge_with_preserved()`
  - `swap_hook_registrations()`
- **Invariants**: All tests pass, behavior unchanged
- **Verification**:
  1. Run full test suite
  2. Run: `./swap-rite.sh hygiene` (full integration)
  3. Verify settings.local.json hooks are correct
  4. Verify user hooks preserved after swap
- **Commit scope**: Remove inline functions (~430 LOC)

**[Final state: Module fully integrated, ~430 LOC extracted]**

---

## Interface Contracts

### parse_hooks_yaml Contract

**Input**: Path to hooks.yaml file

**hooks.yaml Schema**:
```yaml
schema_version: "1.0"  # Optional, warns if not 1.0
hooks:
  - event: SessionStart|Stop|PreToolUse|PostToolUse|UserPromptSubmit
    matcher: "regex"  # Required for PreToolUse/PostToolUse
    path: "relative/path/to/hook.sh"  # Required
    timeout: 5  # Optional, 1-60, default 5
```

**Output**: JSON-lines to stdout
```
{"event":"SessionStart","matcher":"","path":"session-start.sh","timeout":5}
{"event":"PostToolUse","matcher":"Write|Edit","path":"post-write.sh","timeout":10}
```

**Error handling**:
- Missing file: Return empty (0 exit)
- Invalid event: Log warning, skip entry
- Missing matcher for Pre/PostToolUse: Log warning, skip entry
- Invalid regex: Log warning, skip entry
- Timeout > 60: Clamp to 60, log warning
- Timeout < 1: Default to 5

### extract_non_roster_hooks Contract

**Input**: Path to settings.local.json

**Roster hook detection**: Command path contains ".claude/hooks/"

**Output**: JSON object with preserved hooks
```json
{
  "SessionStart": [
    {"hooks": [{"type": "command", "command": "/usr/local/bin/my-hook.sh", "timeout": 5}]}
  ]
}
```

**Invariant**: Only hooks with commands NOT containing ".claude/hooks/" are preserved.

### generate_hooks_json Contract

**Input**: JSON-lines format (from parse/merge)

**Output**: Claude Code hooks object
```json
{
  "SessionStart": [
    {"hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/session-start.sh", "timeout": 5}]}
  ],
  "PostToolUse": [
    {"matcher": "Write|Edit", "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/post-write.sh", "timeout": 10}]}
  ]
}
```

**Grouping**:
1. Group by event type
2. Within event, group by matcher (including empty matcher)
3. Multiple hooks with same event+matcher go in same entry's hooks array

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-017 | Low | 0 files modified | Trivial (delete new file) | None |
| RF-018 | Low | 1 new file | Trivial | RF-017 |
| RF-019 | Low | 1 new file | Trivial | RF-017, RF-018 |
| RF-020 | Low | 1 new file | Trivial | RF-017 |
| RF-021 | Low | 1 new file | Trivial | RF-017 |
| RF-022 | Low | 1 new file | Trivial | RF-017 |
| RF-023 | Low | 1 new file | Trivial | RF-017 |
| RF-024 | Medium | 1 new file | Trivial | RF-017 through RF-023 |
| RF-025 | Medium | 1 file (swap-rite.sh) | 1 commit | RF-017 through RF-024 |
| RF-026 | Medium | 1 file (swap-rite.sh) | 1 commit | RF-025 |

### Risk Details

**RF-019 (parse_hooks_yaml)**: Complex YAML parsing with yq. Must handle edge cases:
- Unicode in paths
- Special characters in matchers
- Malformed YAML

**RF-020 (extract_non_roster_hooks)**: User hook preservation is critical. Incorrect implementation would delete user hooks.

**RF-024 (swap_hook_registrations)**: Orchestrates the full flow. Must maintain atomic write semantics.

---

## Test Requirements

### Unit Tests (tests/lib/team/test-rite-hooks-registration.sh)

| Test | Function | Scenario |
|------|----------|----------|
| test_require_yq_installed | `require_yq` | Returns 0 when yq v4+ installed |
| test_require_yq_missing | `require_yq` | Returns 1 when yq not installed |
| test_require_yq_old_version | `require_yq` | Returns 1 when yq < v4 |
| test_parse_hooks_yaml_valid | `parse_hooks_yaml` | Parses valid hooks.yaml |
| test_parse_hooks_yaml_missing | `parse_hooks_yaml` | Returns empty for missing file |
| test_parse_hooks_yaml_invalid_event | `parse_hooks_yaml` | Skips invalid event types |
| test_parse_hooks_yaml_no_matcher | `parse_hooks_yaml` | Skips PostToolUse without matcher |
| test_parse_hooks_yaml_invalid_regex | `parse_hooks_yaml` | Skips invalid matcher regex |
| test_parse_hooks_yaml_timeout_clamp | `parse_hooks_yaml` | Clamps timeout > 60 |
| test_extract_non_roster_roster_only | `extract_non_roster_hooks` | Returns {} for roster-only hooks |
| test_extract_non_roster_mixed | `extract_non_roster_hooks` | Preserves user hooks only |
| test_extract_non_roster_missing | `extract_non_roster_hooks` | Returns {} for missing file |
| test_merge_hook_registrations | `merge_hook_registrations` | Base first, team second |
| test_merge_hook_registrations_empty | `merge_hook_registrations` | Handles empty inputs |
| test_generate_hooks_json_single | `generate_hooks_json` | Single hook generates correctly |
| test_generate_hooks_json_grouped | `generate_hooks_json` | Same event grouped |
| test_generate_hooks_json_matchers | `generate_hooks_json` | Different matchers separate |
| test_generate_hooks_json_empty | `generate_hooks_json` | Empty input returns {} |
| test_merge_with_preserved_empty | `merge_with_preserved` | Empty preserved unchanged |
| test_merge_with_preserved_append | `merge_with_preserved` | Preserved appended |
| test_swap_hook_registrations_full | `swap_hook_registrations` | Full flow succeeds |
| test_swap_hook_registrations_dry | `swap_hook_registrations` | Dry-run shows preview |
| test_swap_hook_registrations_creates | `swap_hook_registrations` | Creates missing settings |
| test_swap_hook_registrations_corrupted | `swap_hook_registrations` | Backs up corrupted file |

### Test Fixtures Required

```
tests/fixtures/rite-hooks-registration/
  valid-hooks.yaml             # Valid hooks.yaml with multiple hooks
  invalid-event.yaml           # hooks.yaml with invalid event type
  no-matcher.yaml              # PostToolUse without matcher
  bad-regex.yaml               # Invalid regex in matcher
  timeout-exceed.yaml          # timeout > 60
  settings-roster-only.json    # settings.local.json with only roster hooks
  settings-mixed.json          # settings.local.json with mixed hooks
  settings-user-only.json      # settings.local.json with only user hooks
  settings-corrupted.json      # Invalid JSON
```

### Integration Tests (tests/integration/test-swap-hooks-registration.sh)

| Test | Scenario |
|------|----------|
| test_hooks_registration_end_to_end | Full swap with hook registration |
| test_user_hooks_preserved | User hooks survive team swap |
| test_team_switch_hooks | Switch team, verify new hooks registered |

---

## Rollback Strategy

### Per-Phase Rollback

| After Phase | Rollback Action | Data Loss |
|-------------|-----------------|-----------|
| Phase 1-4 | Delete `lib/team/rite-hooks-registration.sh` | None (not integrated) |
| Phase 5 (RF-025) | Remove source statement | None |
| Phase 5 (RF-026) | Revert commit, restore inline functions | None |

### Emergency Rollback

If issues discovered post-integration:
1. Run: `git revert HEAD~2..HEAD` (reverts RF-025, RF-026)
2. Verify: `./swap-rite.sh --help` works
3. Verify: Hook registration works correctly

### Rollback Indicators

Trigger rollback if:
- Hook registration fails with sourcing errors
- settings.local.json not updated after swap
- User hooks deleted after swap
- Invalid JSON in settings.local.json
- yq version check fails when yq v4+ is installed

---

## Notes for Janitor

### Commit Message Conventions

```
refactor(lib/team): [RF-0XX] description

Body explaining what changed and why.

Refs: REFACTOR-rite-hooks-registration.md
```

### Test Run Requirements

After each commit:
1. `bash -n lib/team/rite-hooks-registration.sh` - Syntax check
2. Run unit tests for completed functions
3. After RF-025: Full integration test with `./swap-rite.sh hygiene`
4. After RF-026: Verify settings.local.json hooks are correct

### Files to Avoid Touching

- `lib/team/rite-resource.sh` - Separate module, already extracted
- `lib/team/team-transaction.sh` - Separate module, already extracted
- `rites/*/hooks.yaml` - Source of truth, read-only for this refactor
- `user-hooks/base_hooks.yaml` - Source of truth, read-only for this refactor

### Order is Critical For

- **RF-024 requires RF-017 through RF-023**: Orchestration depends on all helpers
- **RF-025 requires RF-017 through RF-024**: Module must be complete before sourcing
- **RF-026 requires RF-025**: Integration must work before removing inlines

### Source Order in swap-rite.sh

After Priority 3 extraction:
```bash
source "$ROSTER_HOME/lib/roster-utils.sh"
source "$ROSTER_HOME/lib/team/team-transaction.sh"
source "$ROSTER_HOME/lib/team/rite-resource.sh"
source "$ROSTER_HOME/lib/team/rite-hooks-registration.sh"  # NEW
```

### Bash 3.2 Compatibility Notes

- **No associative arrays**: Use parallel arrays or jq for key-value
- **jq required**: All JSON manipulation uses jq (already a dependency)
- **yq v4 required**: YAML parsing requires mikefarah/yq

### Edge Cases to Watch

1. **Unicode in paths**: Ensure hook paths with unicode work
2. **Spaces in paths**: Quote all path variables
3. **Empty hooks.yaml**: Handle files with no hooks
4. **Empty settings.local.json**: Handle {} correctly
5. **Malformed JSON**: Backup and create fresh
6. **yq not installed**: Fail gracefully with clear message

---

## Out of Scope

Findings deferred for future work:

| Finding | Reason |
|---------|--------|
| CH-002 (update_claude_md complexity) | Different concern, separate module candidate |
| CH-003 (swap_hooks complexity) | File sync logic, different from registration |
| CH-001 (perform_swap complexity) | Higher risk, evaluate after all modules extracted |
| SM-006 (swap_* resources) | Higher complexity, evaluate after priority 3 proven |

---

## Verification Attestation

| Artifact | Path | Verified |
|----------|------|----------|
| Spike analysis | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-script-code-smell-refactoring.md` | Read |
| Source file | `/Users/tomtenuta/Code/roster/swap-rite.sh` (lines 2136-2565) | Analyzed |
| Priority 1 plan | `/Users/tomtenuta/Code/roster/docs/refactoring/REFACTOR-rite-resource.md` | Read |
| Priority 2 plan | `/Users/tomtenuta/Code/roster/docs/refactoring/REFACTOR-team-transaction.md` | Read |
| Module pattern | `/Users/tomtenuta/Code/roster/lib/team/rite-resource.sh` | Read |
| Test pattern | `/Users/tomtenuta/Code/roster/tests/lib/team/test-rite-resource.sh` | Read |
| Template | `/Users/tomtenuta/Code/roster/rites/ecosystem/skills/doc-ecosystem/templates/refactoring-plan.md` | Read |
