# Refactoring Plan: lib/team/team-transaction.sh Extraction

**Based on**: SPIKE-script-code-smell-refactoring.md
**Prepared**: 2026-01-03
**Scope**: Extract transaction infrastructure from swap-team.sh to lib/team/team-transaction.sh
**Priority**: 2 (following team-resource.sh extraction)

---

## Executive Summary

This plan extracts transaction infrastructure functions from `swap-team.sh` into a new `lib/team/team-transaction.sh` module. The extraction consolidates ~300 lines of atomic write, journal, staging, and backup operations into a reusable module while preserving transaction atomicity guarantees.

**Expected outcome**: Clean module boundary, testable transaction infrastructure, ~300 LOC extraction, no behavior change.

---

## Architectural Assessment

### Boundary Health

| Module | Assessment |
|--------|------------|
| `swap-team.sh` (main script) | Recovery orchestration and rollback logic must stay; infrastructure can extract |
| `lib/team/team-resource.sh` | Extracted in Priority 1; provides pattern to follow |
| `lib/team/team-transaction.sh` (new) | Will establish clean boundary for transaction infrastructure |

### Root Causes Identified

1. **Infrastructure mixed with orchestration**: Atomic write utilities, journal CRUD, staging helpers, and backup creation are infrastructure that doesn't need access to orchestration state.

2. **Circular dependency risk**: `create_swap_backup()` calls `update_journal_backups()` - this must be preserved in extraction design.

3. **Signal handler coupling**: `handle_interrupt()` calls journal/staging/backup functions. Signal handler registration must stay in swap-team.sh, but called functions can extract.

### Transaction Atomicity Preservation

The extraction must NOT break the transaction safety model:

| Phase | Functions Called | Must Preserve |
|-------|-----------------|---------------|
| PREPARING | `create_journal()` | Journal prevents concurrent swaps |
| BACKING | `create_swap_backup()`, `update_journal_backups()` | Backup enables rollback |
| STAGING | `create_staging()`, `stage_*()` | Staging enables verification |
| VERIFYING | `verify_staging()` | Verification before commit |
| COMMITTING | `update_journal_phase()`, `update_journal_error()` | Phase tracking for recovery |
| COMPLETED | `delete_journal()`, `cleanup_*()` | Cleanup after success |

**Design principle**: Extract infrastructure functions, keep orchestration decisions in swap-team.sh.

---

## Extraction Manifest

### Functions to Extract from swap-team.sh (~300 LOC)

| Current Function | Line | LOC | Category |
|-----------------|------|-----|----------|
| `write_atomic()` | 100-131 | 32 | Atomic I/O |
| `create_journal()` | 135-188 | 54 | Journal CRUD |
| `update_journal_phase()` | 192-213 | 22 | Journal CRUD |
| `update_journal_backups()` | 217-230 | 14 | Journal CRUD |
| `update_journal_error()` | 234-245 | 12 | Journal CRUD |
| `get_journal_field()` | 249-258 | 10 | Journal CRUD |
| `get_journal_phase()` | 261-263 | 3 | Journal CRUD |
| `delete_journal()` | 266-271 | 6 | Journal CRUD |
| `journal_exists()` | 274-276 | 3 | Journal CRUD |
| `create_staging()` | 283-295 | 13 | Staging |
| `cleanup_staging()` | 298-303 | 6 | Staging |
| `stage_agents()` | 307-326 | 20 | Staging |
| `stage_workflow()` | 330-343 | 14 | Staging |
| `stage_active_team()` | 347-357 | 11 | Staging |
| `verify_staging()` | 361-393 | 33 | Staging |
| `create_swap_backup()` | 401-490 | 90 | Backup |
| `cleanup_swap_backup()` | 493-498 | 6 | Backup |
| `verify_backup_integrity()` | 502-534 | 33 | Backup |
| **Total** | | **~302** | |

### Functions STAYING in swap-team.sh

| Function | Lines | Reason |
|----------|-------|--------|
| `rollback_swap()` | 541-668 | Recovery orchestration - makes decisions about what to restore |
| `handle_interrupt()` | 673-709 | Signal handler - decides recovery action by phase |
| `handle_exit()` | 712-716 | Exit handler |
| `setup_signal_handlers()` | 719-722 | trap registration |
| `check_journal_recovery()` | 730-785 | Recovery orchestration - interactive/auto-recover decision |
| `prompt_recovery_action()` | 788+ | User interaction |

**Rationale**: Infrastructure functions are side-effect-free operations on files. Recovery functions make policy decisions and interact with users.

---

## Dependency Analysis

### Call Graph for Extracted Functions

```
write_atomic()
  |-- [standalone, no dependencies]

create_journal()
  |-- write_atomic()

update_journal_phase()
  |-- write_atomic()

update_journal_backups()
  |-- write_atomic()

update_journal_error()
  |-- write_atomic()

get_journal_field()
  |-- [standalone, reads JOURNAL_FILE]

get_journal_phase()
  |-- get_journal_field()

delete_journal()
  |-- [standalone]

journal_exists()
  |-- [standalone]

create_staging()
  |-- [standalone]

cleanup_staging()
  |-- [standalone]

stage_agents()
  |-- [standalone, uses ROSTER_HOME]

stage_workflow()
  |-- [standalone, uses ROSTER_HOME]

stage_active_team()
  |-- [standalone]

verify_staging()
  |-- [standalone]

create_swap_backup()
  |-- update_journal_backups()   <-- INTERNAL CALL

cleanup_swap_backup()
  |-- [standalone]

verify_backup_integrity()
  |-- [standalone, reads JOURNAL_FILE for virgin swap check]
```

### Circular Dependency Analysis

**Identified**: `create_swap_backup()` calls `update_journal_backups()` (lines 456, 470, 484)

**Resolution**: Both functions extract to same module. Internal call preserved.

### External Dependencies

| Dependency | Type | Source |
|------------|------|--------|
| `ROSTER_HOME` | Global variable | Environment / swap-team.sh |
| `JOURNAL_FILE` | Constant | swap-team.sh (must pass or expose) |
| `STAGING_DIR` | Constant | swap-team.sh (must pass or expose) |
| `SWAP_BACKUP_DIR` | Constant | swap-team.sh (must pass or expose) |
| `JOURNAL_VERSION` | Constant | swap-team.sh (must pass or expose) |
| `PHASE_*` | Constants | swap-team.sh (must pass or expose) |
| `MANIFEST_FILE` | Constant | swap-team.sh (for journal backup_location) |
| `log()` | Function | swap-team.sh |
| `log_debug()` | Function | swap-team.sh |
| `log_warning()` | Function | swap-team.sh |
| `log_error()` | Function | swap-team.sh |
| `jq` | External command | System |

---

## Target API

### lib/team/team-transaction.sh Public Interface

```bash
#!/usr/bin/env bash
#
# team-transaction.sh - Transaction Infrastructure for Team Swaps
#
# Provides atomic write, journal management, staging, and backup
# operations for swap-team.sh transaction safety.
#
# Part of: roster team-swap infrastructure
#
# Usage:
#   source "$ROSTER_HOME/lib/team/team-transaction.sh"
#   create_journal "$source_team" "$target_team"
#   create_staging && stage_agents "$team_name" && verify_staging "$count"
#
# Dependencies:
#   - jq (for JSON manipulation)
#   - Logging functions (log, log_debug, log_warning, log_error)
#   - Constants: JOURNAL_FILE, STAGING_DIR, SWAP_BACKUP_DIR, etc.

# Guard against re-sourcing
[[ -n "${_TEAM_TRANSACTION_LOADED:-}" ]] && return 0
readonly _TEAM_TRANSACTION_LOADED=1

# ============================================================================
# Module Constants (if not defined by caller)
# ============================================================================

# Default paths (can be overridden before sourcing)
: "${JOURNAL_FILE:=.claude/.swap-journal}"
: "${JOURNAL_VERSION:=1.0}"
: "${STAGING_DIR:=.claude/.swap-staging}"
: "${SWAP_BACKUP_DIR:=.claude/.swap-backup}"

# Transaction phases
: "${PHASE_PREPARING:=PREPARING}"
: "${PHASE_BACKING:=BACKING}"
: "${PHASE_STAGING:=STAGING}"
: "${PHASE_VERIFYING:=VERIFYING}"
: "${PHASE_COMMITTING:=COMMITTING}"
: "${PHASE_COMPLETED:=COMPLETED}"

# ============================================================================
# Logging Stubs (overridden when sourced from swap-team.sh)
# ============================================================================

if ! type log >/dev/null 2>&1; then
    log() { echo "[Transaction] $*"; }
fi

if ! type log_debug >/dev/null 2>&1; then
    log_debug() { echo "[DEBUG] $*" >&2; }
fi

if ! type log_warning >/dev/null 2>&1; then
    log_warning() { echo "[WARNING] $*" >&2; }
fi

if ! type log_error >/dev/null 2>&1; then
    log_error() { echo "[ERROR] $*" >&2; }
fi

# ============================================================================
# Atomic I/O
# ============================================================================

# Write content atomically using temp file + rename pattern
# Parameters:
#   $1 - target: Target file path
#   $2 - content: Content to write
# Returns: 0 on success, 1 on failure
write_atomic() { ... }

# ============================================================================
# Journal Operations
# ============================================================================

# Create a new journal entry for swap operation
# Parameters:
#   $1 - source_team: Current team (empty string for virgin swap)
#   $2 - target_team: Team being swapped to
# Returns: 0 on success, 1 if journal already exists (concurrent swap)
# Requires: JOURNAL_FILE, JOURNAL_VERSION, SWAP_BACKUP_DIR, STAGING_DIR
create_journal() { ... }

# Update journal phase
# Parameters:
#   $1 - new_phase: Phase name (PHASE_* constant)
# Returns: 0 on success, 1 if journal missing
update_journal_phase() { ... }

# Update journal backup locations for resources
# Parameters:
#   $1 - resource_type: "commands" | "skills" | "hooks"
#   $2 - backup_path: Path to backup directory
# Returns: 0 on success, 1 if journal missing
update_journal_backups() { ... }

# Update journal with error message
# Parameters:
#   $1 - error_msg: Error message to record
# Returns: 0 on success, 1 if journal missing
update_journal_error() { ... }

# Read arbitrary journal field
# Parameters:
#   $1 - field: Field name in journal JSON
# Outputs: Field value to stdout, empty if not found
# Returns: 0 always (empty output for missing field)
get_journal_field() { ... }

# Get current journal phase
# Outputs: Phase name to stdout
# Returns: 0 on success, 1 if journal missing
get_journal_phase() { ... }

# Delete journal (on successful completion)
# Returns: 0 always
delete_journal() { ... }

# Check if journal exists
# Returns: 0 if exists, 1 otherwise
journal_exists() { ... }

# ============================================================================
# Staging Operations
# ============================================================================

# Create staging directory structure
# Returns: 0 on success, 1 on failure
# Side effects: Creates STAGING_DIR, removes any existing staging
create_staging() { ... }

# Clean up staging directory
# Returns: 0 always
# Side effects: Removes STAGING_DIR if exists
cleanup_staging() { ... }

# Stage agents from team pack
# Parameters:
#   $1 - team_name: Team to stage agents from
# Returns: 0 on success, 1 on failure
# Requires: ROSTER_HOME, STAGING_DIR
stage_agents() { ... }

# Stage workflow file
# Parameters:
#   $1 - team_name: Team to stage workflow from
# Returns: 0 on success, 1 on failure (warning if no workflow.yaml)
# Requires: ROSTER_HOME, STAGING_DIR
stage_workflow() { ... }

# Stage ACTIVE_RITE file
# Parameters:
#   $1 - team_name: Team name to write
# Returns: 0 on success, 1 on failure
# Requires: STAGING_DIR
stage_active_team() { ... }

# Verify staging directory integrity
# Parameters:
#   $1 - expected_count: Expected number of agent .md files
# Returns: 0 on success, 1 on verification failure
# Requires: STAGING_DIR
verify_staging() { ... }

# ============================================================================
# Swap Backup Operations
# ============================================================================

# Create comprehensive backup for transaction safety
# Returns: 0 on success, 1 on failure
# Requires: SWAP_BACKUP_DIR, MANIFEST_FILE (for backup path in journal)
# Side effects: Updates journal with backup locations
create_swap_backup() { ... }

# Clean up swap backup (after successful swap)
# Returns: 0 always
# Side effects: Removes SWAP_BACKUP_DIR if exists
cleanup_swap_backup() { ... }

# Verify backup integrity for recovery
# Returns: 0 if backup valid, 1 if missing/corrupted
# Requires: SWAP_BACKUP_DIR, JOURNAL_FILE (for virgin swap check)
verify_backup_integrity() { ... }
```

---

## Refactoring Sequence

### Phase 1: Foundation [Low Risk]

**Goal**: Create module structure and implement atomic write utility.

#### RF-010: Create lib/team/team-transaction.sh with sourcing guard

- **Smells addressed**: Preparation for transaction function extraction
- **Category**: Local (infrastructure)
- **Before**: No `lib/team/team-transaction.sh` exists
- **After**: Module exists with:
  - Sourcing guard (`_TEAM_TRANSACTION_LOADED`)
  - Module header documentation
  - Default constant definitions with `: "${VAR:=default}"` pattern
  - Logging stubs (same pattern as team-resource.sh)
- **Invariants**: swap-team.sh unchanged, still works
- **Verification**:
  1. Run: `bash -n lib/team/team-transaction.sh` (syntax check)
  2. Run: `./swap-team.sh --help` (existing behavior unchanged)
- **Commit scope**: Create module skeleton with guards and stubs

#### RF-011: Extract write_atomic()

- **Smells addressed**: Prepare for journal function extraction
- **Category**: Local
- **Before**:
  ```bash
  # swap-team.sh:100-131
  write_atomic() {
      local target="$1"
      local content="$2"
      local temp="${target}.tmp.$$"
      # ... 28 lines of atomic write logic
  }
  ```
- **After**:
  ```bash
  # lib/team/team-transaction.sh
  write_atomic() {
      # Same implementation, identical behavior
  }
  ```
- **Invariants**:
  - Same temp file naming pattern (`.tmp.$$`)
  - Same error handling (returns 1, logs error)
  - Parent directory creation preserved
  - Sync to disk best-effort preserved
- **Verification**:
  1. Unit test: Write to new file, verify content
  2. Unit test: Write to existing file, verify atomic replacement
  3. Unit test: Write to non-existent parent, verify directory created
  4. Unit test: Verify temp file cleaned up on failure
- **Commit scope**: Add `write_atomic()`, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 2: Journal Operations [Low Risk]

**Goal**: Extract all journal CRUD operations.

#### RF-012: Extract journal CRUD functions

- **Smells addressed**: Clean module boundary for journal operations
- **Category**: Local
- **Before**: 9 journal functions inline in swap-team.sh (lines 135-276)
- **After**: Same 9 functions in team-transaction.sh
- **Functions extracted**:
  1. `create_journal()` - Creates journal entry
  2. `update_journal_phase()` - Updates phase field
  3. `update_journal_backups()` - Updates backup_location
  4. `update_journal_error()` - Records error
  5. `get_journal_field()` - Reads arbitrary field
  6. `get_journal_phase()` - Gets phase (wrapper around get_journal_field)
  7. `delete_journal()` - Removes journal file
  8. `journal_exists()` - Checks if journal exists
- **Invariants**:
  - Journal JSON structure unchanged
  - Same concurrent swap detection (journal exists check)
  - Same PID tracking in journal
  - `jq` usage patterns preserved
- **Verification**:
  1. Unit test: `create_journal` creates valid JSON
  2. Unit test: `create_journal` fails if journal exists
  3. Unit test: `update_journal_phase` updates correctly
  4. Unit test: `get_journal_field` returns correct values
  5. Unit test: `journal_exists` returns correct status
- **Commit scope**: Add all 9 journal functions, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 3: Staging Operations [Low Risk]

**Goal**: Extract all staging operations.

#### RF-013: Extract staging functions

- **Smells addressed**: Clean module boundary for staging operations
- **Category**: Local
- **Before**: 6 staging functions inline in swap-team.sh (lines 283-393)
- **After**: Same 6 functions in team-transaction.sh
- **Functions extracted**:
  1. `create_staging()` - Creates staging directory
  2. `cleanup_staging()` - Removes staging directory
  3. `stage_agents()` - Stages agents from team pack
  4. `stage_workflow()` - Stages workflow.yaml
  5. `stage_active_team()` - Stages ACTIVE_RITE file
  6. `verify_staging()` - Verifies staging integrity
- **Invariants**:
  - Same staging directory structure
  - Same agent file counting method
  - Same verification logic (agent count, ACTIVE_RITE presence)
  - `cp -rp` for agents (preserve timestamps)
- **Verification**:
  1. Unit test: `create_staging` creates directory
  2. Unit test: `create_staging` cleans existing staging
  3. Unit test: `stage_agents` copies from team pack
  4. Unit test: `verify_staging` fails on wrong count
  5. Unit test: `verify_staging` fails on missing ACTIVE_RITE
- **Commit scope**: Add all 6 staging functions, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 4: Backup Operations [Medium Risk]

**Goal**: Extract backup operations while preserving internal dependency.

**Risk factor**: `create_swap_backup()` internally calls `update_journal_backups()`. Both must be in module for call to work.

#### RF-014: Extract backup functions

- **Smells addressed**: Complete transaction infrastructure extraction
- **Category**: Module boundary (internal function call)
- **Before**: 3 backup functions inline in swap-team.sh (lines 401-534)
- **After**: Same 3 functions in team-transaction.sh
- **Functions extracted**:
  1. `create_swap_backup()` - Creates comprehensive backup
  2. `cleanup_swap_backup()` - Removes backup directory
  3. `verify_backup_integrity()` - Verifies backup for recovery
- **Internal dependency preserved**:
  ```bash
  # create_swap_backup() calls update_journal_backups() at lines 456, 470, 484
  # Both functions will be in same module - call works unchanged
  ```
- **Invariants**:
  - Same backup directory structure
  - Same backup order (agents, ACTIVE_RITE, manifest, workflow, commands, skills, hooks)
  - Virgin swap detection preserved in verify_backup_integrity()
  - Marker file backup preserved (.rite-commands, etc.)
- **Verification**:
  1. Unit test: `create_swap_backup` creates complete backup
  2. Unit test: `create_swap_backup` updates journal backup locations
  3. Unit test: `verify_backup_integrity` detects missing backup
  4. Unit test: `verify_backup_integrity` handles virgin swap
  5. Integration: Full backup/verify cycle
- **Commit scope**: Add all 3 backup functions, add unit tests

**[Rollback point: can stop here safely - module not integrated yet]**

---

### Phase 5: Integration [Medium Risk]

**Goal**: Wire new module into swap-team.sh, remove inline implementations.

#### RF-015: Add source statement in swap-team.sh

- **Smells addressed**: Final integration
- **Category**: Boundary integration
- **Before**: All transaction functions inline
- **After**:
  ```bash
  # swap-team.sh (after lib/roster-utils.sh, before lib/team/team-resource.sh)
  source "$ROSTER_HOME/lib/team/team-transaction.sh"
  ```
- **Order matters**: team-transaction.sh must be sourced before team-resource.sh if any resource functions need transaction functions (currently they don't, but order is defensive)
- **Invariants**:
  - Constants defined before sourcing (JOURNAL_FILE, etc.) take precedence
  - Logging functions defined before sourcing take precedence
  - All existing function calls work unchanged
- **Verification**:
  1. Run: `./swap-team.sh --help` (basic smoke test)
  2. Run: `./swap-team.sh --dry-run hygiene-pack` (full flow)
  3. Verify: Journal created/deleted correctly
  4. Verify: Staging created/verified/cleaned correctly
- **Commit scope**: Add source statement only

#### RF-016: Remove inline transaction functions

- **Smells addressed**: Final cleanup, eliminate duplication
- **Category**: Local
- **Before**: 18 inline functions in swap-team.sh (lines 100-534)
- **After**: Only source statement remains
- **Functions removed**:
  - `write_atomic()`
  - `create_journal()`
  - `update_journal_phase()`
  - `update_journal_backups()`
  - `update_journal_error()`
  - `get_journal_field()`
  - `get_journal_phase()`
  - `delete_journal()`
  - `journal_exists()`
  - `create_staging()`
  - `cleanup_staging()`
  - `stage_agents()`
  - `stage_workflow()`
  - `stage_active_team()`
  - `verify_staging()`
  - `create_swap_backup()`
  - `cleanup_swap_backup()`
  - `verify_backup_integrity()`
- **Invariants**: All tests pass, behavior unchanged
- **Verification**:
  1. Run full test suite
  2. Run: `./swap-team.sh hygiene-pack` (full integration)
  3. Test signal handling: `Ctrl+C` during swap, verify recovery works
  4. Test recovery: `./swap-team.sh --recover` after interrupted swap
- **Commit scope**: Remove inline functions (~300 LOC)

**[Final state: Module fully integrated, ~300 LOC extracted]**

---

## Interface Contracts

### Journal Contract

**create_journal(source_team, target_team)**

| Property | Specification |
|----------|---------------|
| Input | source_team: string (empty for virgin), target_team: string |
| Output | None |
| Return | 0 success, 1 if journal exists |
| Side effects | Creates JOURNAL_FILE with JSON content |
| Error behavior | Logs error, returns 1 |
| Concurrent safety | Fails if JOURNAL_FILE exists (prevents concurrent swaps) |

**Journal JSON Schema**:
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

### Staging Contract

**verify_staging(expected_count)**

| Property | Specification |
|----------|---------------|
| Input | expected_count: integer (expected .md file count) |
| Output | Debug logs |
| Return | 0 if verified, 1 if verification fails |
| Checks | STAGING_DIR exists, agents dir exists, agent count matches, ACTIVE_RITE exists |

### Backup Contract

**create_swap_backup()**

| Property | Specification |
|----------|---------------|
| Input | None (uses globals) |
| Output | Debug logs |
| Return | 0 success, 1 on failure |
| Side effects | Creates SWAP_BACKUP_DIR, updates journal backup_location |
| Backup order | agents, ACTIVE_RITE, manifest, workflow, commands, skills, hooks |
| Marker files | .rite-commands, .rite-skills, .rite-hooks copied if exist |

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-010 | Low | 0 files modified | Trivial (delete new file) | None |
| RF-011 | Low | 1 new file | Trivial | RF-010 |
| RF-012 | Low | 1 new file | Trivial | RF-010, RF-011 |
| RF-013 | Low | 1 new file | Trivial | RF-010 |
| RF-014 | Medium | 1 new file | Trivial | RF-010, RF-012 |
| RF-015 | Medium | 1 file (swap-team.sh) | 1 commit | RF-010 through RF-014 |
| RF-016 | Medium | 1 file (swap-team.sh) | 1 commit | RF-015 |

### Risk Details

**RF-014 (Medium)**: Internal function call from `create_swap_backup()` to `update_journal_backups()`. Both must extract together.

**RF-015 (Medium)**: First integration point. If constants not defined before sourcing, module defaults will apply. Verify constant precedence.

**RF-016 (Medium)**: Removes ~300 LOC. If anything missed, swap will break. Integration tests critical.

---

## Test Requirements

### Unit Tests (tests/lib/team/test-team-transaction.sh)

| Test | Function | Scenario |
|------|----------|----------|
| test_write_atomic_new_file | `write_atomic` | Creates new file atomically |
| test_write_atomic_overwrite | `write_atomic` | Overwrites existing file |
| test_write_atomic_creates_parent | `write_atomic` | Creates parent directory |
| test_write_atomic_cleanup_on_fail | `write_atomic` | Cleans temp file on failure |
| test_create_journal | `create_journal` | Creates valid JSON journal |
| test_create_journal_concurrent | `create_journal` | Fails if journal exists |
| test_update_journal_phase | `update_journal_phase` | Updates phase correctly |
| test_update_journal_backups | `update_journal_backups` | Updates backup_location |
| test_get_journal_field | `get_journal_field` | Returns correct field value |
| test_get_journal_phase | `get_journal_phase` | Returns phase via get_journal_field |
| test_delete_journal | `delete_journal` | Removes journal file |
| test_journal_exists_true | `journal_exists` | Returns 0 when exists |
| test_journal_exists_false | `journal_exists` | Returns 1 when missing |
| test_create_staging | `create_staging` | Creates directory |
| test_create_staging_cleans | `create_staging` | Removes existing staging |
| test_cleanup_staging | `cleanup_staging` | Removes staging directory |
| test_stage_agents | `stage_agents` | Copies agents from team pack |
| test_stage_workflow | `stage_workflow` | Copies workflow.yaml |
| test_stage_active_team | `stage_active_team` | Creates ACTIVE_RITE file |
| test_verify_staging_success | `verify_staging` | Passes with correct count |
| test_verify_staging_wrong_count | `verify_staging` | Fails on count mismatch |
| test_verify_staging_missing_dir | `verify_staging` | Fails on missing directory |
| test_create_swap_backup | `create_swap_backup` | Creates complete backup |
| test_create_swap_backup_updates_journal | `create_swap_backup` | Updates journal backup_location |
| test_cleanup_swap_backup | `cleanup_swap_backup` | Removes backup directory |
| test_verify_backup_integrity | `verify_backup_integrity` | Returns 0 for valid backup |
| test_verify_backup_missing | `verify_backup_integrity` | Returns 1 for missing backup |
| test_verify_backup_virgin | `verify_backup_integrity` | Handles virgin swap case |

### Integration Tests (tests/integration/test-swap-team-transaction.sh)

| Test | Scenario |
|------|----------|
| test_transaction_full_cycle | Journal -> Staging -> Backup -> Cleanup cycle |
| test_transaction_recovery | Create journal, simulate interrupt, verify recovery state |
| test_transaction_concurrent | Attempt two swaps, verify second blocked |

### Test Fixtures Required

```
tests/fixtures/team-transaction/
  mock-teams/
    test-team/
      agents/
        orchestrator.md
        worker.md
      workflow.yaml
  mock-project/
    .claude/
      agents/
        old-agent.md
      ACTIVE_RITE        # Contains: old-team
      AGENT_MANIFEST.json
      commands/
        .rite-commands
        test-cmd.md
```

---

## Rollback Strategy

### Per-Phase Rollback

| After Phase | Rollback Action | Data Loss |
|-------------|-----------------|-----------|
| Phase 1-4 | Delete `lib/team/team-transaction.sh` | None (not integrated) |
| Phase 5 (RF-015) | Remove source statement | None |
| Phase 5 (RF-016) | Revert commit, restore inline functions | None |

### Emergency Rollback

If issues discovered post-integration:
1. Run: `git revert HEAD~2..HEAD` (reverts RF-015, RF-016)
2. Verify: `./swap-team.sh --help` works
3. Verify: Team swap with journal works correctly

### Rollback Indicators

Trigger rollback if:
- Swap fails with sourcing errors
- Journal not created during swap
- Staging verification fails when it shouldn't
- Signal handling doesn't trigger rollback
- Recovery mode doesn't work

---

## Notes for Janitor

### Commit Message Conventions

```
refactor(lib/team): [RF-0XX] description

Body explaining what changed and why.

Refs: REFACTOR-team-transaction.md
```

### Test Run Requirements

After each commit:
1. `bash -n lib/team/team-transaction.sh` - Syntax check
2. Run unit tests for completed functions
3. After RF-015: Full integration test with `./swap-team.sh --dry-run hygiene-pack`
4. After RF-016: Test signal handling and recovery

### Files to Avoid Touching

- `lib/team/team-resource.sh` - Separate module, already extracted
- `.claude/settings.json` - Hook registration, different concern
- `teams/*/` - Source of truth, read-only for this refactor
- Rollback functions in swap-team.sh - Stay in main script

### Order is Critical For

- **RF-014 requires RF-012**: Backup functions call journal functions
- **RF-015 requires RF-010 through RF-014**: Module must be complete before sourcing
- **RF-016 requires RF-015**: Integration must work before removing inlines

### Source Order in swap-team.sh

Current:
```bash
source "$ROSTER_HOME/lib/roster-utils.sh"
source "$ROSTER_HOME/lib/team/team-resource.sh"
```

After:
```bash
source "$ROSTER_HOME/lib/roster-utils.sh"
source "$ROSTER_HOME/lib/team/team-transaction.sh"  # NEW
source "$ROSTER_HOME/lib/team/team-resource.sh"
```

### Bash 3.2 Compatibility Notes

- **No nameref**: Cannot use `declare -n`
- **Use stdout**: Results returned via stdout
- **jq required**: All JSON manipulation uses jq (already a dependency)

### Signal Handler Interaction

Signal handlers in swap-team.sh call these functions:

```bash
handle_interrupt() {
    phase=$(get_journal_phase)          # Calls module
    case "$phase" in
        ...
        cleanup_staging                 # Calls module
        delete_journal                  # Calls module
        rollback_swap                   # STAYS in swap-team.sh
    esac
}
```

The signal handler registration (`trap`) and the handler function itself stay in swap-team.sh. Only the infrastructure functions they call extract to the module.

---

## Out of Scope

Findings deferred for future work:

| Finding | Reason |
|---------|--------|
| `rollback_swap()` (~130 LOC) | Recovery orchestration, makes policy decisions |
| `check_journal_recovery()` | Interactive/auto-recover logic, not pure infrastructure |
| Hook registration (settings.local.json) | Different concern, separate module candidate |
| CH-001 (perform_swap complexity) | Higher risk, evaluate after transaction module proven |

---

## Verification Attestation

| Artifact | Path | Verified |
|----------|------|----------|
| Spike analysis | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-script-code-smell-refactoring.md` | Read |
| Source file | `/Users/tomtenuta/Code/roster/swap-team.sh` (lines 100-534, 669-785) | Analyzed |
| Module pattern | `/Users/tomtenuta/Code/roster/lib/team/team-resource.sh` | Read |
| Priority 1 plan | `/Users/tomtenuta/Code/roster/docs/refactoring/REFACTOR-team-resource.md` | Read |
| Template | `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/skills/doc-ecosystem/templates/refactoring-plan.md` | Read |
