# Refactoring Plan: Doctrine Purity Cleanup

**Based on**: Greenfield migration audit (2026-01-06/07)
**Prepared**: 2026-01-07
**Scope**: Eliminate legacy terminology (roster->knossos, team->rite) with no backward compatibility

## Executive Summary

This refactoring plan addresses the final greenfield migration to achieve 100% Knossos doctrine terminology alignment. Unlike previous plans, this assumes **no backward compatibility requirements**--breaking changes are approved.

### Audit Summary (From Context)

| Category | Critical | High | Medium | Notes |
|----------|----------|------|--------|-------|
| roster->knossos | 47 | 4 | ~200 | Env vars, log prefixes, comments |
| team in Go code | - | - | ~180 | Output types, JSON tags |
| team in docs/comments | - | - | 600+ | Documentation debt |
| File/dir naming | 4 | - | 32 | roster-utils.sh, output/team.go |
| Shell/config | 8 | 15 | 45+ | ROSTER_DEBUG, [Roster] prefixes |

### Key Findings

1. **output/team.go**: Legacy output types (TeamListOutput, TeamStatusOutput, etc.) with `team` JSON tags
2. **roster-utils.sh**: File name and internal terminology
3. **ROSTER_DEBUG**: Environment variable used in 27+ shell scripts
4. **active_team JSON fields**: ~68 files still use this field name
5. **swap-rite.sh**: Uses `[Roster]` log prefix, `ROSTER_DEBUG` env var

---

## Architectural Assessment

### Boundary Health

| Component | Status | Assessment |
|-----------|--------|------------|
| `ariadne/internal/output/team.go` | **CRITICAL** | Output structs export `team` JSON tags; needs rename to rite.go with rite terminology |
| `ariadne/internal/output/rite.go` | Clean | Already uses correct terminology |
| `ariadne/internal/cmd/rite/` | Clean | Uses output.TeamStatusOutput--needs update after output migration |
| `lib/roster-utils.sh` | **HIGH** | File name contains roster; sourced by swap-rite.sh |
| `swap-rite.sh` | **HIGH** | Uses ROSTER_DEBUG, [Roster] log prefix |
| `lib/sync/*.sh` | **MEDIUM** | Uses roster terminology in comments and variable names |
| `user-hooks/lib/session-manager.sh` | **MEDIUM** | Outputs active_team in JSON |

### Root Causes Identified

1. **Partial Migration (Phase 1)**: Previous refactoring renamed file paths and Go struct field names but preserved JSON tags for backward compatibility
2. **Shell Script Legacy**: ROSTER_DEBUG and [Roster] prefixes predate Knossos doctrine
3. **Documentation Debt**: ~600+ references to team in docs/comments accumulated over time
4. **Dual Output Types**: Both TeamListOutput and RiteListOutput exist--need consolidation

### Dependency Graph

```
Phase 1: Critical Infrastructure (no dependencies)
  +--> lib/roster-utils.sh rename
  +--> ROSTER_DEBUG -> KNOSSOS_DEBUG
  +--> [Roster] -> [Knossos] log prefix

Phase 2: Go Output Types (depends on Phase 1 for consistency)
  +--> output/team.go rename to output/rite.go
  |    +--> Type renames: Team* -> Rite*
  |    +--> JSON tags: active_team -> active_rite
  +--> cmd/rite/ imports update

Phase 3: Shell Function Updates (depends on Phase 1)
  +--> swap-rite.sh internal variables
  +--> session-manager.sh JSON output

Phase 4: JSON Migration (depends on Phase 2)
  +--> active_team -> active_rite in JSON output
  +--> Schema updates for validation

Phase 5: Finalization (depends on all)
  +--> ari inscription sync
  +--> Documentation sweep
```

---

## Phase Structure with Architectural Contracts

### Phase 1: Critical Infrastructure [Low Risk]

**Goal**: Rename shell library and update environment variables without breaking callers

#### RF-001: Rename lib/roster-utils.sh

**Before State**:
- File: `lib/roster-utils.sh`
- Guard: `_ROSTER_UTILS_LOADED`
- Functions: `generate_roster()`, `get_produces()`, `truncate()`, `get_frontmatter()`
- Sourced by: `swap-rite.sh` (line 793)

**After State**:
- File: `lib/knossos-utils.sh`
- Guard: `_KNOSSOS_UTILS_LOADED`
- Functions unchanged (semantic names, roster means "list")

**Invariants**:
- `source "$KNOSSOS_HOME/lib/knossos-utils.sh"` succeeds
- All functions callable with identical behavior
- `swap-rite.sh` sources new path

**Verification**:
```bash
# Check file exists
test -f "$KNOSSOS_HOME/lib/knossos-utils.sh"
# Check functions exist
source "$KNOSSOS_HOME/lib/knossos-utils.sh"
type generate_roster >/dev/null 2>&1 && echo "OK"
# Run swap-rite.sh smoke test
./swap-rite.sh list
```

**Commit**: `refactor(shell): rename lib/roster-utils.sh to lib/knossos-utils.sh [DP-P1-001]`

---

#### RF-002: Update ROSTER_DEBUG to KNOSSOS_DEBUG

**Before State**:
- 27 files use `ROSTER_DEBUG="${ROSTER_DEBUG:-0}"`
- Debug function: `[[ "$ROSTER_DEBUG" == "1" ]]`

**After State**:
- All files use `KNOSSOS_DEBUG="${KNOSSOS_DEBUG:-0}"`
- Debug function: `[[ "$KNOSSOS_DEBUG" == "1" ]]`

**Files Affected**:
1. `swap-rite.sh` (lines 16, 98)
2. `sync-user-skills.sh` (lines 31, 90)
3. `sync-user-hooks.sh` (lines 30, 89)
4. `sync-user-agents.sh` (lines 29, 81)
5. `sync-user-commands.sh` (lines 30, 83)
6. `lib/rite/rite-transaction.sh`
7. `lib/rite/rite-hooks-registration.sh`
8. `tests/lib/rite/test-rite-transaction.sh`
9. `templates/generate-orchestrator.sh`
10. `templates/validate-orchestrator.sh`
11. `install-hooks.sh`
12. `load-workflow.sh`
13. `get-workflow-field.sh`
14. `bin/fix-hardcoded-paths.sh`
15. And 12 more test files

**Invariants**:
- `KNOSSOS_DEBUG=1 ./swap-rite.sh --help 2>&1 | grep DEBUG` returns debug output
- No references to ROSTER_DEBUG in shell files (except docs)

**Verification**:
```bash
# No ROSTER_DEBUG in active code
grep -r "ROSTER_DEBUG" --include="*.sh" | grep -v "^docs/" | wc -l  # Should be 0
# Debug mode works
KNOSSOS_DEBUG=1 ./swap-rite.sh list 2>&1 | head -5
```

**Commit**: `refactor(shell): rename ROSTER_DEBUG to KNOSSOS_DEBUG [DP-P1-002]`

---

#### RF-003: Update [Roster] log prefix to [Knossos]

**Before State**:
- `swap-rite.sh`: `log() { echo "[Roster] $*" }` (lines 85-100)
- Similar patterns in sync-user-*.sh scripts

**After State**:
- All scripts: `log() { echo "[Knossos] $*" }`

**Files Affected**:
1. `swap-rite.sh` (lines 85-100)
2. `sync-user-skills.sh`
3. `sync-user-hooks.sh`
4. `sync-user-agents.sh`
5. `sync-user-commands.sh`

**Invariants**:
- All log output shows `[Knossos]` prefix
- Error/warning/debug functions updated consistently

**Verification**:
```bash
grep -r "\[Roster\]" --include="*.sh" | wc -l  # Should be 0
./swap-rite.sh list | head -3  # Should show [Knossos]
```

**Commit**: `refactor(shell): update [Roster] log prefix to [Knossos] [DP-P1-003]`

---

**[ROLLBACK POINT 1]**: Infrastructure renamed. Scripts functional. Can stop here safely.

---

### Phase 2: Go Core Types [Medium Risk]

**Goal**: Consolidate output/team.go into output/rite.go with correct terminology

#### RF-004: Rename output/team.go to output/rite-legacy.go (preparation)

**Rationale**: output/rite.go already exists with RiteListOutput, RiteInfoOutput, etc. The team.go types (TeamListOutput, TeamStatusOutput, TeamSwitchOutput, etc.) are used by cmd/rite/ commands. Strategy:
1. First deprecate team.go types by renaming file
2. Update cmd/rite/ to use equivalent rite.go types
3. Remove team.go entirely

**Before State**:
- `ariadne/internal/output/team.go`: TeamListOutput, TeamSummary, TeamStatusOutput, TeamSwitchOutput, TeamSwitchDryRunOutput, TeamValidateOutput (243 lines)
- `ariadne/internal/output/rite.go`: RiteListOutput, RiteSummary, RiteInfoOutput, etc. (correct terminology)
- `ariadne/internal/cmd/rite/status.go`: Uses `output.TeamStatusOutput`
- `ariadne/internal/cmd/rite/validate.go`: Uses `output.TeamValidateOutput`

**After State**:
- `ariadne/internal/output/team.go`: DELETED
- `ariadne/internal/output/rite.go`: Contains all output types with rite terminology
- `ariadne/internal/cmd/rite/status.go`: Uses `output.RiteStatusOutput`
- `ariadne/internal/cmd/rite/validate.go`: Uses `output.RiteValidateOutput`

**Sub-Tasks**:

##### RF-004a: Add missing Rite output types to rite.go

Add to `output/rite.go`:
```go
// RiteStatusOutput (equivalent to TeamStatusOutput)
type RiteStatusOutput struct {
    Rite           string        `json:"rite"`
    IsActive       bool          `json:"is_active"`
    Path           string        `json:"path"`
    Description    string        `json:"description"`
    WorkflowType   string        `json:"workflow_type"`
    Agents         []AgentStatus `json:"agents"`
    Phases         []string      `json:"phases,omitempty"`
    EntryPoint     string        `json:"entry_point"`
    Orphans        []string      `json:"orphans,omitempty"`
    ManifestValid  bool          `json:"manifest_valid"`
    ClaudeMDSynced bool          `json:"claude_md_synced"`
}

// RiteSwitchOutput (equivalent to TeamSwitchOutput)
type RiteSwitchOutput struct {
    Rite               string                    `json:"rite"`
    PreviousRite       string                    `json:"previous_rite"`
    SwitchedAt         string                    `json:"switched_at"`
    AgentsInstalled    []string                  `json:"agents_installed"`
    OrphansHandled     *OrphanHandleResult       `json:"orphans_handled,omitempty"`
    ClaudeMDUpdated    bool                      `json:"claude_md_updated"`
    ManifestPath       string                    `json:"manifest_path"`
    InscriptionSynced  bool                      `json:"inscription_synced,omitempty"`
    InscriptionVersion string                    `json:"inscription_version,omitempty"`
    SyncConflicts      []InscriptionConflictOut  `json:"sync_conflicts,omitempty"`
}

// RiteSwitchDryRunOutput (equivalent to TeamSwitchDryRunOutput)
type RiteSwitchDryRunOutput struct {
    DryRun                 bool     `json:"dry_run"`
    WouldSwitchTo          string   `json:"would_switch_to"`
    CurrentRite            string   `json:"current_rite"`
    WouldInstall           []string `json:"would_install"`
    OrphansDetected        []string `json:"orphans_detected"`
    OrphanStrategyRequired bool     `json:"orphan_strategy_required"`
    SuggestedFlags         []string `json:"suggested_flags,omitempty"`
}

// RiteValidateOutput (equivalent to TeamValidateOutput)
type RiteValidateOutput struct {
    Rite     string               `json:"rite"`
    Valid    bool                 `json:"valid"`
    Checks   []ValidationCheckOut `json:"checks"`
    Errors   int                  `json:"errors"`
    Warnings int                  `json:"warnings"`
    Fixable  []string             `json:"fixable,omitempty"`
}
```

**Invariants**:
- All new types have `json:"rite"` tags, not `json:"team"`
- Text() methods return user-friendly output
- Headers()/Rows() methods implement Tabular interface

##### RF-004b: Update cmd/rite/ to use new output types

**Files**:
- `ariadne/internal/cmd/rite/status.go` (line 89): Change `output.TeamStatusOutput` to `output.RiteStatusOutput`
- `ariadne/internal/cmd/rite/validate.go`: Change `output.TeamValidateOutput` to `output.RiteValidateOutput`

##### RF-004c: Delete output/team.go

Once all consumers migrated, delete the file.

**Verification**:
```bash
cd ariadne
go build ./...
go test ./...
# Test JSON output
go run ./cmd/ari/main.go rite status -o json | jq '.rite'  # Should work
```

**Commit**: `refactor(go): consolidate output/team.go into output/rite.go [DP-P2-004]`

---

#### RF-005: Update validate.go error messages

**Before State**:
- `ariadne/internal/cmd/validate/validate.go`: "Team not found", "Team validation failed" messages

**After State**:
- All messages use "Rite" terminology

**Verification**:
```bash
grep -r "Team not found\|Team validation" ariadne/ | wc -l  # Should be 0
```

**Commit**: `refactor(go): update error messages to rite terminology [DP-P2-005]`

---

**[ROLLBACK POINT 2]**: Go types migrated. All tests pass. JSON output uses rite fields.

---

### Phase 3: Shell Function Updates [Medium Risk]

**Goal**: Update internal shell functions and JSON output

#### RF-006: Update swap-rite.sh internal terminology

**Before State**:
- Variables: `active_team`, `current_team`, `team_name` (~50 occurrences)
- JSON output: `"active_team": "$team_name"`
- Log messages: References to "team"

**After State**:
- Variables: `active_rite`, `current_rite`, `rite_name`
- JSON output: `"active_rite": "$rite_name"`
- Log messages: References to "rite"

**Key Lines**:
- Line 860: `"active_team": "$team_name"` -> `"active_rite": "$rite_name"`
- Line 1004: `"active_team": "unknown"` -> `"active_rite": "unknown"`

**Invariants**:
- `swap-rite.sh list -o json | jq '.active_rite'` returns value
- `swap-rite.sh hygiene-pack -o json | jq '.rite'` returns value

**Verification**:
```bash
./swap-rite.sh list -o json | jq '.active_rite'
grep -c "active_team" swap-rite.sh  # Should be 0 (excluding comments)
```

**Commit**: `refactor(shell): update swap-rite.sh to rite terminology [DP-P3-006]`

---

#### RF-007: Update session-manager.sh JSON output

**Before State**:
- Line 272: `"active_team": "$active_rite"`

**After State**:
- `"active_rite": "$active_rite"`

**Verification**:
```bash
grep "active_team" user-hooks/lib/session-manager.sh  # Should be 0
```

**Commit**: `refactor(shell): update session-manager.sh to rite terminology [DP-P3-007]`

---

#### RF-008: Update lib/sync/*.sh terminology

**Before State**:
- `lib/sync/sync-core.sh`: Uses `roster_file`, `roster_checksum`, `roster_updated` variable names
- Comments reference "roster" extensively

**After State**:
- Variable names: `knossos_file`, `knossos_checksum`, `knossos_updated`
- Comments updated to reference "knossos" where appropriate
- Keep `roster` where semantically correct (e.g., "generate_roster" generates a list)

**Files**:
1. `lib/sync/sync-core.sh` (~100 occurrences)
2. `lib/sync/sync-manifest.sh` (~30 occurrences)
3. `lib/sync/merge/merge-docs.sh` (~20 occurrences)
4. `lib/sync/merge/merge-settings.sh` (~15 occurrences)
5. `lib/sync/merge/dispatcher.sh` (~10 occurrences)

**Invariants**:
- All sync operations still work
- `roster-sync` command functions (this is the command name, may stay)

**Verification**:
```bash
# Run sync tests
./tests/sync/test-sync-config.sh
./tests/sync/test-sync-conflict.sh
```

**Commit**: `refactor(shell): update lib/sync/ terminology [DP-P3-008]`

---

**[ROLLBACK POINT 3]**: Shell scripts fully updated. Integration tests must pass.

---

### Phase 4: File/Directory Renames [Low Risk]

**Goal**: Rename remaining files with legacy terminology

#### RF-009: Rename user-commands/team-switching/

**Before State**:
- Directory: `user-commands/team-switching/` (10 files)
- Referenced in: `sync-user-commands.sh`, docs

**After State**:
- Directory: `user-commands/rite-switching/`
- All references updated

**Verification**:
```bash
test -d user-commands/rite-switching && ls user-commands/rite-switching/*.md | wc -l  # Should be 10
grep -r "team-switching" user-commands/ docs/ sync-user-commands.sh | wc -l  # Should be 0 (or only historical docs)
```

**Commit**: `refactor(commands): rename team-switching/ to rite-switching/ [DP-P4-009]`

---

#### RF-010: Update forge-pack skill directory

**Before State** (if not already done):
- `rites/forge-pack/skills/team-development/`

**After State**:
- `rites/forge-pack/skills/rite-development/`

**Note**: Verify current state first--may already be migrated.

**Commit**: `refactor(forge): rename team-development to rite-development [DP-P4-010]`

---

#### RF-011: Update schema files

**Before State**:
- `user-skills/guidance/rite-discovery/schemas/team-profile.yaml`

**After State**:
- `user-skills/guidance/rite-discovery/schemas/rite-profile.yaml`
- Internal references updated

**Commit**: `refactor(schemas): rename team-profile.yaml to rite-profile.yaml [DP-P4-011]`

---

**[ROLLBACK POINT 4]**: Directory renames complete.

---

### Phase 5: Finalization [Low Risk]

**Goal**: Regenerate files and validate complete migration

#### RF-012: Run inscription sync

**Command**: `ari inscription sync`

**Purpose**: Regenerate .claude/ files with updated terminology from templates

**Verification**:
1. CLAUDE.md updated correctly
2. No sync conflicts
3. Agent tables reflect rite terminology

**Commit**: `chore: regenerate .claude/ via inscription sync [DP-P5-012]`

---

#### RF-013: Documentation sweep

**Scope**: Update remaining documentation references

**Files** (sampling):
- `docs/design/TDD-ariadne-team.md`: Update examples
- `docs/design/TDD-session-manager-ecosystem-audit.md`: Update field references
- `README.md`: Update terminology

**Note**: Historical documents (ADRs, dated reports) should retain original terminology for accuracy.

**Commit**: `docs: update team->rite terminology in documentation [DP-P5-013]`

---

#### RF-014: Integration test validation

**Test Suite**:
```bash
# Go tests
cd ariadne && go test ./...

# Shell script tests
./tests/sync/test-sync-config.sh
./tests/sync/test-sync-conflict.sh
./tests/sync/test-swap-rite-integration.sh

# Manual validation
ari rite list
ari rite status
ari rite list -o json | jq '.active_rite'
./swap-rite.sh hygiene-pack --dry-run
```

**Commit**: `test: validate doctrine purity refactoring [DP-P5-014]`

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-001 | Low | 2 files | 1 commit | None |
| RF-002 | Medium | 27 files | 1 commit | None |
| RF-003 | Low | 5 files | 1 commit | None |
| RF-004 | Medium | 4 Go files | 1 commit | None |
| RF-005 | Low | 1 file | 1 commit | RF-004 |
| RF-006 | Medium | 1 file | 1 commit | RF-001, RF-002, RF-003 |
| RF-007 | Low | 1 file | 1 commit | None |
| RF-008 | Medium | 5 files | 1 commit | RF-001 |
| RF-009 | Low | 10 files | 1 commit (git mv) | None |
| RF-010 | Low | 3 dirs | 1 commit | None |
| RF-011 | Low | 1 file | 1 commit | None |
| RF-012 | Low | Generated | N/A | All above |
| RF-013 | Low | Many | 1 commit | All above |
| RF-014 | N/A | Validation | N/A | All above |

---

## Rollback Strategy

### Phase 1 Rollback
```bash
git revert HEAD~3..HEAD  # Reverts RF-001, RF-002, RF-003
# Or selective:
git checkout HEAD~1 -- lib/knossos-utils.sh && mv lib/knossos-utils.sh lib/roster-utils.sh
```

### Phase 2 Rollback
```bash
git revert HEAD~2..HEAD  # Reverts RF-004, RF-005
# Restore team.go from git history:
git checkout HEAD~2 -- ariadne/internal/output/team.go
```

### Phase 3 Rollback
```bash
git revert HEAD~3..HEAD  # Reverts RF-006, RF-007, RF-008
```

### Phase 4 Rollback
```bash
# Directory renames are trivial git mv operations
git revert HEAD~3..HEAD  # RF-009, RF-010, RF-011
```

### Phase 5 Rollback
```bash
ari inscription sync  # Regenerate from source
git revert HEAD  # RF-013 documentation
```

---

## Architectural Contracts

### Contract 1: JSON Output Field Names

**Invariant**: All CLI commands outputting JSON use `rite` terminology

| Before | After | Commands Affected |
|--------|-------|-------------------|
| `active_team` | `active_rite` | session status, rite list, rite status |
| `team` | `rite` | rite switch, rite status |
| `previous_team` | `previous_rite` | rite switch |
| `current_team` | `current_rite` | rite switch --dry-run |

**Verification**: `ari <cmd> -o json | jq 'keys' | grep -v team`

### Contract 2: Environment Variables

**Invariant**: All shell scripts use KNOSSOS_* environment variables

| Before | After |
|--------|-------|
| ROSTER_DEBUG | KNOSSOS_DEBUG |
| ROSTER_HOME | KNOSSOS_HOME (already done) |

**Verification**: `grep -r "ROSTER_" --include="*.sh" | grep -v "^docs/"`

### Contract 3: Log Prefixes

**Invariant**: All log output uses [Knossos] prefix

**Verification**: `./swap-rite.sh list 2>&1 | grep -c "\[Roster\]"` returns 0

### Contract 4: File Path Consistency

**Invariant**: No file paths contain "team" where "rite" is correct

| Before | After |
|--------|-------|
| `lib/roster-utils.sh` | `lib/knossos-utils.sh` |
| `user-commands/team-switching/` | `user-commands/rite-switching/` |
| `output/team.go` | `output/rite.go` (consolidated) |

---

## Test Strategy

### Unit Tests (per commit)

```bash
# After Go changes
cd ariadne && go test ./internal/output/... ./internal/cmd/rite/...

# After shell changes
./tests/lib/rite/test-rite-transaction.sh
./tests/lib/rite/test-rite-hooks-registration.sh
```

### Integration Tests (per phase)

```bash
# Phase 1: Infrastructure
KNOSSOS_DEBUG=1 ./swap-rite.sh --help  # Check debug works
./swap-rite.sh list  # Check log prefix

# Phase 2: Go types
ari rite list -o json | jq .  # Valid JSON
ari rite status -o json | jq '.rite'  # Field exists

# Phase 3: Shell
./swap-rite.sh list -o json | jq '.active_rite'  # Field exists

# Phase 4/5: Full
./tests/sync/test-swap-rite-integration.sh
```

### Regression Detection

Key behaviors that must not change:
1. `swap-rite.sh hygiene-pack` successfully swaps rite
2. `ari rite list` shows available rites
3. Session creation includes active rite in context
4. Inscription sync regenerates CLAUDE.md correctly

---

## Notes for Janitor

### Commit Message Conventions

Format: `{type}({scope}): {description} [DP-P{phase}-{seq}]`

Types:
- `refactor` - Code structure changes
- `fix` - Bug fixes
- `docs` - Documentation only
- `chore` - Maintenance tasks
- `test` - Test updates

Examples:
- `refactor(shell): rename lib/roster-utils.sh to lib/knossos-utils.sh [DP-P1-001]`
- `refactor(go): consolidate output/team.go into output/rite.go [DP-P2-004]`

### Test Run Requirements

After each commit:
1. `go build ./...` must succeed (for Go changes)
2. `go test ./...` must pass (for Go changes)
3. Source file and verify functions exist (for shell changes)

After each phase:
1. Full test suite for affected components
2. Manual smoke test of affected commands

### Critical Ordering

1. RF-001 must complete before RF-006 (swap-rite.sh sources new path)
2. RF-002 must complete before RF-006 (KNOSSOS_DEBUG used)
3. RF-004 must complete before RF-005 (types before error messages)
4. RF-012 must be LAST (regenerates from updated sources)

### Files to Avoid Touching

- Historical ADR documents (preserve original terminology for accuracy)
- Test fixture files with documented `active_team` fields (update in separate PR after schema migration)
- Generated files in `.claude/` (will be regenerated in RF-012)

---

## Out of Scope

Findings deferred for future work:

1. **roster-sync command name**: The command `roster-sync` may warrant renaming to `knossos-sync`, but this affects external tooling and documentation. Deferred.

2. **Go package rename internal/team/**: Package exists but may be intentionally named for the `ari team` backward compat command. Low priority.

3. **Thread contract package name**: `internal/hook/threadcontract/` vs `clewcontract/` - doctrine allows drift here.

4. **Cross-team handoff skills**: These use "team" semantically (cross-team coordination). Keep.

---

## Success Criteria

The refactoring is complete when:

- [ ] No `ROSTER_DEBUG` references in shell files (except docs)
- [ ] No `[Roster]` log prefixes
- [ ] `lib/knossos-utils.sh` exists, `lib/roster-utils.sh` does not
- [ ] No `output/team.go` (consolidated into rite.go)
- [ ] All JSON output uses `active_rite`, `rite`, `previous_rite`, `current_rite`
- [ ] `user-commands/rite-switching/` exists with 10 files
- [ ] All Go tests pass
- [ ] All shell tests pass
- [ ] `ari inscription sync` completes without conflict
- [ ] `grep -r "active_team" --include="*.go" ariadne/` returns 0 (excluding test fixtures)

---

## Artifact Attestation

| Source | Path | Verified |
|--------|------|----------|
| output/team.go | `/Users/tomtenuta/Code/roster/ariadne/internal/output/team.go` | Read |
| output/rite.go | `/Users/tomtenuta/Code/roster/ariadne/internal/output/rite.go` | Read |
| swap-rite.sh | `/Users/tomtenuta/Code/roster/swap-rite.sh` | Read (partial) |
| roster-utils.sh | `/Users/tomtenuta/Code/roster/lib/roster-utils.sh` | Read |
| knossos-home.sh | `/Users/tomtenuta/Code/roster/lib/knossos-home.sh` | Read |
| session-manager.sh | Referenced in GAP-ANALYSIS |
| cmd/rite/status.go | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/rite/status.go` | Read |
| GAP-ANALYSIS-team-to-rite-migration.md | `/Users/tomtenuta/Code/roster/docs/ecosystem/GAP-ANALYSIS-team-to-rite-migration.md` | Read |
| doctrine-compliance-checklist.yaml | `/Users/tomtenuta/Code/roster/knossos/doctrine-compliance-checklist.yaml` | Read |
| REFACTOR-PLAN-doctrine-purity.md | `/Users/tomtenuta/Code/roster/docs/plans/REFACTOR-PLAN-doctrine-purity.md` | Read |
| REFACTOR-PLAN-doctrine-purity-phase2.md | `/Users/tomtenuta/Code/roster/docs/plans/REFACTOR-PLAN-doctrine-purity-phase2.md` | Read |
