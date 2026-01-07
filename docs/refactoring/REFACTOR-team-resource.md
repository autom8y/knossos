# Refactoring Plan: lib/team/team-resource.sh Extraction

**Based on**: SPIKE-script-code-smell-refactoring.md
**Prepared**: 2026-01-03
**Scope**: Extract team resource operations from swap-team.sh to lib/team/team-resource.sh

---

## Executive Summary

This plan extracts 5 DRY-violation patterns (SM-001 through SM-005) from `swap-team.sh` into a new `lib/team/team-resource.sh` module. The extraction consolidates ~400 lines of duplicated code into ~150 lines of generic functions while preserving all existing behavior.

**Expected outcome**: 6x DRY improvement, ~250 net LOC reduction, testable module isolation.

---

## Architectural Assessment

### Boundary Health

| Module | Assessment |
|--------|------------|
| `swap-team.sh` (main script) | Orchestration logic is sound; resource operations are inline duplicates that should be extracted |
| `lib/sync/` | Clean module pattern to follow; demonstrates proper sourcing guards and function organization |
| `lib/team/` (new) | Will establish clean boundary for team resource infrastructure |

### Root Causes Identified

1. **Resource-type-specific duplication**: Each resource type (commands, skills, hooks) has near-identical backup, remove, detect, and cleanup functions. The only differences are:
   - Directory paths (`.claude/commands/`, `.claude/skills/`, `.claude/hooks/`)
   - Marker files (`.rite-commands`, `.rite-skills`, `.rite-hooks`)
   - Find type (`-type f` for files, `-type d` for directories)

2. **Global array coupling**: Orphan detection writes to global arrays (`ORPHAN_COMMANDS`, `ORPHAN_SKILLS`, `ORPHAN_HOOKS`) rather than returning data via stdout.

### Smells Addressed

| ID | Pattern | Occurrences | Lines | Root Cause |
|----|---------|-------------|-------|------------|
| SM-001 | `backup_team_*` | 3 | ~90 | Resource-type-specific duplication |
| SM-002 | `remove_team_*` | 3 | ~66 | Resource-type-specific duplication |
| SM-003 | `detect_*_orphans` | 3 | ~102 | Resource-type-specific duplication |
| SM-004 | `remove_orphan_*` | 3 | ~99 | Resource-type-specific duplication |
| SM-005 | `is_team_*/get_*_team` | 6 | ~42 | Resource-type-specific duplication |

---

## Extraction Manifest

### Functions to Extract from swap-team.sh

| Current Function | Location (Line) | Extract To | New Generic Name |
|-----------------|-----------------|------------|------------------|
| `backup_team_commands()` | 2050 | team-resource.sh | `backup_team_resource()` |
| `backup_team_skills()` | 2306 | team-resource.sh | (consolidated) |
| `backup_team_hooks()` | 2620 | team-resource.sh | (consolidated) |
| `remove_team_commands()` | 2083 | team-resource.sh | `remove_team_resource()` |
| `remove_team_skills()` | 2427 | team-resource.sh | (consolidated) |
| `remove_team_hooks()` | 2653 | team-resource.sh | (consolidated) |
| `is_team_command()` | 2107 | team-resource.sh | `is_resource_from_team()` |
| `is_team_skill()` | 2339 | team-resource.sh | (consolidated) |
| `is_team_hook()` | 2677 | team-resource.sh | (consolidated) |
| `get_command_team()` | 2113 | team-resource.sh | `get_resource_team()` |
| `get_skill_team()` | 2345 | team-resource.sh | (consolidated) |
| `get_hook_team()` | 2683 | team-resource.sh | (consolidated) |
| `detect_command_orphans()` | 2123 | team-resource.sh | `detect_resource_orphans()` |
| `detect_skill_orphans()` | 2356 | team-resource.sh | (consolidated) |
| `detect_hook_orphans()` | 2693 | team-resource.sh | (consolidated) |
| `remove_orphan_commands()` | 2154 | team-resource.sh | `remove_resource_orphans()` |
| `remove_orphan_skills()` | 2389 | team-resource.sh | (consolidated) |
| `remove_orphan_hooks()` | 2728 | team-resource.sh | (consolidated) |

### Functions Staying in swap-team.sh

| Function | Reason |
|----------|--------|
| `swap_commands()` | Orchestration - calls backup/remove/copy sequence |
| `swap_skills()` | Orchestration - calls backup/remove/copy sequence |
| `swap_hooks()` | Orchestration - calls backup/remove/copy sequence |
| `check_user_command_collisions()` | Collision detection with user-specific logging |
| `remove_team_agents()` | Different pattern (manifest-based, not marker-file) |

---

## Target API

### lib/team/team-resource.sh Public Interface

```bash
#!/usr/bin/env bash
# lib/team/team-resource.sh - Generic team resource operations
#
# Consolidates backup, removal, orphan detection, and team membership
# checks for commands, skills, and hooks into parameterized functions.
#
# Usage:
#   source "$ROSTER_HOME/lib/team/team-resource.sh"
#   backup_team_resource "commands" ".claude/commands" ".rite-commands" "f"
#   detect_resource_orphans "commands" ".claude/commands" "my-team" "f"

# =============================================================================
# Function: backup_team_resource
# =============================================================================
# Backs up team-owned resources to a .backup directory before swap.
#
# Parameters:
#   $1 - resource_type: "commands" | "skills" | "hooks"
#   $2 - resource_dir:  ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - marker_file:   ".rite-commands" | ".rite-skills" | ".rite-hooks"
#   $4 - find_type:     "f" (file) | "d" (directory)
#
# Returns: 0 on success, 0 if nothing to backup
#
# Side effects:
#   - Creates ${resource_dir}.backup/ directory
#   - Copies team resources to backup
#   - Logs via log_debug()
backup_team_resource() { ... }

# =============================================================================
# Function: remove_team_resource
# =============================================================================
# Removes team-owned resources listed in marker file.
#
# Parameters:
#   $1 - resource_type: "commands" | "skills" | "hooks"
#   $2 - resource_dir:  ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - marker_file:   ".rite-commands" | ".rite-skills" | ".rite-hooks"
#   $4 - find_type:     "f" (file) | "d" (directory)
#
# Returns: 0 on success
#
# Side effects:
#   - Removes resources listed in marker file
#   - Removes marker file itself
#   - Logs via log_debug()
remove_team_resource() { ... }

# =============================================================================
# Function: is_resource_from_team
# =============================================================================
# Checks if a resource belongs to ANY team pack in ROSTER_HOME/teams/.
#
# Parameters:
#   $1 - resource_name: basename of resource (e.g., "commit.md", "qa-ref")
#   $2 - resource_type: "commands" | "skills" | "hooks"
#   $3 - find_type:     "f" (file) | "d" (directory)
#
# Returns: 0 if resource is from a team, 1 otherwise
#
# Requires: ROSTER_HOME environment variable
is_resource_from_team() { ... }

# =============================================================================
# Function: get_resource_team
# =============================================================================
# Gets the team name that owns a specific resource.
#
# Parameters:
#   $1 - resource_name: basename of resource
#   $2 - resource_type: "commands" | "skills" | "hooks"
#   $3 - find_type:     "f" (file) | "d" (directory)
#
# Outputs: team name to stdout, empty if not found
#
# Requires: ROSTER_HOME environment variable
get_resource_team() { ... }

# =============================================================================
# Function: detect_resource_orphans
# =============================================================================
# Detects orphaned resources from other teams that shouldn't be present.
#
# Parameters:
#   $1 - resource_type:     "commands" | "skills" | "hooks"
#   $2 - resource_dir:      ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - incoming_team:     name of team being swapped in
#   $4 - find_type:         "f" (file) | "d" (directory)
#   $5 - glob_pattern:      "*.md" | "*/" | "*" (for find pattern)
#
# Outputs: One "resource_name:origin_team" per line to stdout
#
# Returns: 0 always (empty output means no orphans)
#
# Note: Uses stdout instead of global arrays for bash 3.2 portability
detect_resource_orphans() { ... }

# =============================================================================
# Function: remove_resource_orphans
# =============================================================================
# Removes orphaned resources based on orphan mode.
#
# Parameters:
#   $1 - resource_type:  "commands" | "skills" | "hooks"
#   $2 - resource_dir:   ".claude/commands" | ".claude/skills" | ".claude/hooks"
#   $3 - orphan_mode:    "remove" | "keep" | ""
#   $4 - find_type:      "f" (file) | "d" (directory)
#   stdin:               orphan list (one "name:team" per line)
#
# Returns: 0 on success
#
# Side effects:
#   - Creates ${resource_dir}.orphan-backup/ if removing
#   - Backs up and removes orphaned resources
#   - Logs removals via log()
remove_resource_orphans() { ... }
```

---

## Dependency Map

### Required by lib/team/team-resource.sh

| Dependency | Type | Source |
|------------|------|--------|
| `ROSTER_HOME` | Global variable | Environment / swap-team.sh |
| `log()` | Function | swap-team.sh |
| `log_debug()` | Function | swap-team.sh |
| `log_warning()` | Function | swap-team.sh |

### Not Required (Design Decision)

| Avoided | Reason |
|---------|--------|
| `ORPHAN_MODE` | Passed as parameter instead of global |
| `ORPHAN_*` arrays | Return via stdout for bash 3.2 portability |
| Direct file operations | Kept generic; caller provides paths |

---

## Refactoring Sequence

### Phase 1: Foundation [Low Risk]

**Goal**: Create module structure and implement non-breaking additions.

#### RF-001: Create lib/team/ directory structure

- **Smells addressed**: Preparation for SM-001 through SM-005
- **Category**: Local (infrastructure)
- **Before**: No `lib/team/` directory exists
- **After**: `lib/team/team-resource.sh` exists with sourcing guard
- **Invariants**: swap-team.sh still works unchanged (not sourcing new module yet)
- **Verification**:
  1. Run: `bash -n lib/team/team-resource.sh` (syntax check)
  2. Run: `./swap-team.sh --help` (existing behavior unchanged)
- **Commit scope**:
  - Create `lib/team/team-resource.sh` with header, sourcing guard, and empty function stubs

#### RF-002: Implement is_resource_from_team() and get_resource_team()

- **Smells addressed**: SM-005 (is_team_*/get_*_team 6x)
- **Category**: Local
- **Before**: Six separate functions with identical logic except path component
  ```bash
  # swap-team.sh:2107
  is_team_command() {
      local cmd_name="$1"
      find "$ROSTER_HOME/teams" -path "*/commands/$cmd_name" -type f 2>/dev/null | grep -q .
  }
  # ... repeated for skill (-type d) and hook (-type f)
  ```
- **After**: Two generic functions with type parameter
  ```bash
  # lib/team/team-resource.sh
  is_resource_from_team() {
      local resource_name="$1"
      local resource_type="$2"  # commands, skills, hooks
      local find_type="$3"      # f or d
      find "$ROSTER_HOME/teams" -path "*/${resource_type}/$resource_name" -type "$find_type" 2>/dev/null | grep -q .
  }
  ```
- **Invariants**:
  - Same return codes (0 = found, 1 = not found)
  - Same find patterns after parameter substitution
  - ROSTER_HOME must be set
- **Verification**:
  1. Unit test: `is_resource_from_team "commit.md" "commands" "f"` returns same as `is_team_command "commit.md"`
  2. Unit test: `get_resource_team "qa-ref" "skills" "d"` returns same as `get_skill_team "qa-ref"`
- **Commit scope**: Add two functions to team-resource.sh, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 2: Core Operations [Low Risk]

**Goal**: Implement backup and remove operations.

#### RF-003: Implement backup_team_resource()

- **Smells addressed**: SM-001 (backup_team_* 3x)
- **Category**: Local
- **Before**: Three functions with identical structure
  ```bash
  # swap-team.sh:2050 - backup_team_commands()
  local backup_dir=".claude/commands.backup"
  if [[ ! -d ".claude/commands" ]] || [[ ! -f ".claude/commands/.rite-commands" ]]; then
      return 0
  fi
  # ... identical pattern for skills and hooks
  ```
- **After**: Single parameterized function
  ```bash
  backup_team_resource() {
      local resource_type="$1"
      local resource_dir="$2"
      local marker_file="$3"
      local find_type="$4"
      local backup_dir="${resource_dir}.backup"

      if [[ ! -d "$resource_dir" ]] || [[ ! -f "$resource_dir/$marker_file" ]]; then
          log_debug "No team ${resource_type} to backup"
          return 0
      fi
      # ... generic implementation
  }
  ```
- **Invariants**:
  - Backup directory naming convention preserved (`.backup` suffix)
  - Old backup removed before new backup created
  - Resources listed in marker file are copied
  - `log_debug` messages reference resource type
- **Verification**:
  1. Unit test: Mock team commands, verify backup directory created
  2. Unit test: Verify old backup removed before new created
  3. Integration: Swap between two teams, verify backup behavior unchanged
- **Commit scope**: Add `backup_team_resource()`, add unit tests

#### RF-004: Implement remove_team_resource()

- **Smells addressed**: SM-002 (remove_team_* 3x)
- **Category**: Local
- **Before**: Three functions with identical structure (lines 2083, 2427, 2653)
- **After**: Single parameterized function
- **Invariants**:
  - Resources listed in marker file are removed
  - Marker file itself is removed after resources
  - Directories (skills) use `rm -rf`, files use `rm -f`
  - `log_debug` messages reference resource type
- **Verification**:
  1. Unit test: Create mock marker file, verify listed resources removed
  2. Unit test: Verify marker file removed
  3. Unit test: Verify non-listed resources preserved
- **Commit scope**: Add `remove_team_resource()`, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 3: Orphan Management [Medium Risk]

**Goal**: Implement orphan detection and removal with stdout-based data flow.

#### RF-005: Implement detect_resource_orphans()

- **Smells addressed**: SM-003 (detect_*_orphans 3x)
- **Category**: Module boundary (changes data flow from global arrays to stdout)
- **Before**: Writes to global arrays (`ORPHAN_COMMANDS`, `ORPHAN_SKILLS`, `ORPHAN_HOOKS`)
  ```bash
  # swap-team.sh:2127
  ORPHAN_COMMANDS=()
  # ... detection logic ...
  ORPHAN_COMMANDS+=("$cmd_name:$origin_team")
  ```
- **After**: Returns data via stdout (bash 3.2 portable)
  ```bash
  detect_resource_orphans() {
      # ... detection logic ...
      echo "$resource_name:$origin_team"  # One per line
  }
  ```
- **Invariants**:
  - Same orphan detection logic (resource not in incoming team, but is from some team)
  - Same output format: "name:origin_team"
  - Skills use directory glob (`*/`), commands/hooks use file glob
- **Verification**:
  1. Unit test: Set up mock teams, verify correct orphans detected
  2. Unit test: Verify incoming team resources NOT flagged as orphans
  3. Unit test: Verify non-team resources NOT flagged as orphans
- **Commit scope**: Add `detect_resource_orphans()`, add unit tests

#### RF-006: Implement remove_resource_orphans()

- **Smells addressed**: SM-004 (remove_orphan_* 3x)
- **Category**: Module boundary (reads from stdin instead of global arrays)
- **Before**: Reads from global arrays, uses global `ORPHAN_MODE`
  ```bash
  # swap-team.sh:2154
  for orphan in "${ORPHAN_COMMANDS[@]}"; do
      case "$ORPHAN_MODE" in
  ```
- **After**: Reads from stdin, receives mode as parameter
  ```bash
  remove_resource_orphans() {
      local orphan_mode="$3"
      while IFS=: read -r name team; do
          case "$orphan_mode" in
  ```
- **Invariants**:
  - Same backup-before-remove behavior in "remove" mode
  - Same log messages (referencing origin team)
  - Same backup directory naming (`.orphan-backup` suffix)
  - "keep" mode logs keeping message
  - Empty/unknown mode keeps silently
- **Verification**:
  1. Unit test: Pipe orphan list, verify backup created in remove mode
  2. Unit test: Verify keep mode preserves resources
  3. Unit test: Verify backup directory summary logged
- **Commit scope**: Add `remove_resource_orphans()`, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 4: Integration [Medium Risk]

**Goal**: Wire new module into swap-team.sh, replace inline implementations.

#### RF-007: Add source statement and wrapper functions

- **Smells addressed**: Final consolidation
- **Category**: Boundary integration
- **Before**: All functions inline in swap-team.sh
- **After**:
  ```bash
  # swap-team.sh (near top, after other lib sources)
  source "$ROSTER_HOME/lib/team/team-resource.sh"

  # Wrapper functions for backward compatibility
  backup_team_commands() { backup_team_resource "commands" ".claude/commands" ".rite-commands" "f"; }
  backup_team_skills()   { backup_team_resource "skills" ".claude/skills" ".rite-skills" "d"; }
  backup_team_hooks()    { backup_team_resource "hooks" ".claude/hooks" ".rite-hooks" "f"; }
  # ... similar for remove, is_team, get_team
  ```
- **Invariants**:
  - All existing function names remain callable
  - All existing callers work without modification
  - No behavior change observable from outside
- **Verification**:
  1. Run: `./swap-team.sh hygiene-pack` (full integration)
  2. Run: `./swap-team.sh different-team` (team swap)
  3. Verify: Backups created, orphans handled, team activated
- **Commit scope**: Add source statement, add wrapper functions

#### RF-008: Migrate orphan detection to stdout pattern

- **Smells addressed**: Global array elimination
- **Category**: Boundary integration
- **Before**: Global arrays populated, then read by remove functions
  ```bash
  detect_command_orphans "$incoming_team"  # Sets ORPHAN_COMMANDS
  remove_orphan_commands                   # Reads ORPHAN_COMMANDS
  ```
- **After**: Pipe-based data flow
  ```bash
  detect_resource_orphans "commands" ".claude/commands" "$incoming_team" "f" "*.md" \
      | remove_resource_orphans "commands" ".claude/commands" "$ORPHAN_MODE" "f"
  ```
- **Invariants**:
  - Same orphans detected and processed
  - Same logging output
  - Same backup behavior
- **Verification**:
  1. Test: Swap from team A to team B with orphan commands
  2. Verify: Orphans backed up and removed (or kept per mode)
  3. Verify: Log messages unchanged
- **Commit scope**: Update orphan handling callsites in swap-team.sh

#### RF-009: Remove inline duplicate functions

- **Smells addressed**: Final cleanup
- **Category**: Local
- **Before**: 18 inline functions in swap-team.sh (3 types x 6 operations)
- **After**: 6 wrapper functions + sourced module
- **Invariants**: All tests pass, behavior unchanged
- **Verification**:
  1. Run full test suite
  2. Run: `./swap-team.sh --dry-run` for each team
  3. Manual verification: LOC reduced by ~250
- **Commit scope**: Remove inline functions, keep only wrappers

**[Final state: Module fully integrated, duplicates eliminated]**

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-001 | Low | 0 files modified | Trivial (delete new dir) | None |
| RF-002 | Low | 1 new file | Trivial | RF-001 |
| RF-003 | Low | 1 new file | Trivial | RF-001, RF-002 |
| RF-004 | Low | 1 new file | Trivial | RF-001 |
| RF-005 | Low | 1 new file | Trivial | RF-001, RF-002 |
| RF-006 | Low | 1 new file | Trivial | RF-001, RF-005 |
| RF-007 | Medium | 1 file (swap-team.sh) | 1 commit | RF-001 through RF-006 |
| RF-008 | Medium | 1 file (swap-team.sh) | 1 commit | RF-007 |
| RF-009 | Medium | 1 file (swap-team.sh) | 1 commit | RF-008 |

---

## Test Requirements

### Unit Tests (tests/lib/team/test-team-resource.sh)

| Test | Function | Scenario |
|------|----------|----------|
| test_is_resource_from_team_command | `is_resource_from_team` | Finds command in teams/ |
| test_is_resource_from_team_skill | `is_resource_from_team` | Finds skill directory in teams/ |
| test_is_resource_from_team_not_found | `is_resource_from_team` | Returns 1 for non-team resource |
| test_get_resource_team | `get_resource_team` | Returns correct team name |
| test_get_resource_team_not_found | `get_resource_team` | Returns empty for non-team resource |
| test_backup_team_resource_commands | `backup_team_resource` | Creates .backup directory |
| test_backup_team_resource_no_marker | `backup_team_resource` | Returns 0 when no marker file |
| test_backup_team_resource_replaces_old | `backup_team_resource` | Removes old backup first |
| test_remove_team_resource | `remove_team_resource` | Removes listed resources |
| test_remove_team_resource_removes_marker | `remove_team_resource` | Removes marker file |
| test_detect_resource_orphans | `detect_resource_orphans` | Outputs orphan:team format |
| test_detect_resource_orphans_skips_incoming | `detect_resource_orphans` | Incoming team not flagged |
| test_remove_resource_orphans_remove_mode | `remove_resource_orphans` | Backs up and removes |
| test_remove_resource_orphans_keep_mode | `remove_resource_orphans` | Preserves resources |

### Integration Tests (tests/integration/test-swap-team-resource.sh)

| Test | Scenario |
|------|----------|
| test_swap_preserves_behavior | Full swap with generic module produces same result as original |
| test_orphan_flow_end_to_end | Orphan detection through removal works via piped data |

### Test Fixtures Required

```
tests/fixtures/team-resource/
  mock-teams/
    team-a/
      commands/
        cmd-a.md
      skills/
        skill-a/
      hooks/
        hook-a.sh
    team-b/
      commands/
        cmd-b.md
      skills/
        skill-b/
  mock-project/
    .claude/
      commands/
        cmd-a.md
        .rite-commands  # Contains: cmd-a.md
      skills/
        skill-a/
        .rite-skills    # Contains: skill-a
```

---

## Rollback Strategy

### Per-Phase Rollback

| After Phase | Rollback Action | Data Loss |
|-------------|-----------------|-----------|
| Phase 1-3 | Delete `lib/team/team-resource.sh` | None (not integrated) |
| Phase 4 (RF-007) | Revert commit, restore inline functions | None |
| Phase 4 (RF-008) | Revert commit, restore global array flow | None |
| Phase 4 (RF-009) | Revert commit | None |

### Emergency Rollback

If issues discovered post-integration:
1. Run: `git revert HEAD~3..HEAD` (reverts RF-007 through RF-009)
2. Verify: `./swap-team.sh --help` works
3. Verify: Team swap functions correctly

### Rollback Indicators

Trigger rollback if:
- Team swap fails with sourcing errors
- Orphan detection misses resources it should detect
- Backup directories not created when expected
- Any behavior change observed during integration tests

---

## Notes for Janitor

### Commit Message Conventions

```
refactor(lib/team): [RF-XXX] description

Body explaining what changed and why.

Refs: REFACTOR-team-resource.md
```

### Test Run Requirements

After each commit:
1. `bash -n lib/team/team-resource.sh` - Syntax check
2. Run unit tests for completed functions
3. After RF-007: Full integration test with `./swap-team.sh hygiene-pack`

### Files to Avoid Touching

- `lib/sync/*` - Separate module, not part of this refactoring
- `.claude/settings.json` - Hook registration, different concern
- `teams/*/` - Source of truth, read-only for this refactor

### Order is Critical For

- **RF-007 requires RF-001 through RF-006**: Module must be complete before integration
- **RF-008 requires RF-007**: Wrappers must exist before changing data flow
- **RF-009 requires RF-008**: Remove inlines only after new flow proven

### Glob Patterns by Resource Type

| Resource | Find Type | Glob Pattern | Notes |
|----------|-----------|--------------|-------|
| commands | `-type f` | `*.md` | Files only |
| skills | `-type d` | `*/` | Directories only |
| hooks | `-type f` | `*` | All files (no extension filter) |

### Bash 3.2 Compatibility

- **No nameref**: Cannot use `declare -n` for array passing
- **Use stdout**: All array-like returns via stdout, one item per line
- **Colon-separated**: Orphan format `name:team` for simple parsing

---

## Out of Scope

Findings deferred for future work:

| Finding | Reason |
|---------|--------|
| SM-006 (swap_* resources) | Higher complexity, evaluate after team-resource.sh proven |
| CH-001 (perform_swap complexity) | Requires architectural decision about phase decomposition |
| `remove_team_agents()` | Different pattern (manifest-based), separate refactoring |
| Hook registration (`settings.local.json`) | Different concern, separate module candidate |

---

## Verification Attestation

| Artifact | Path | Verified |
|----------|------|----------|
| Spike analysis | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-script-code-smell-refactoring.md` | Read |
| Source file | `/Users/tomtenuta/Code/roster/swap-team.sh` | Analyzed (lines 2050-2760) |
| Module pattern | `/Users/tomtenuta/Code/roster/lib/sync/sync-core.sh` | Read |
| Module pattern | `/Users/tomtenuta/Code/roster/lib/sync/sync-manifest.sh` | Read |
| Template | `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/skills/doc-ecosystem/templates/refactoring-plan.md` | Read |
