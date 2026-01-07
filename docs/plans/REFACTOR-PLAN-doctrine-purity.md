# Refactoring Plan: Doctrine Purity Sprint

**Based on**: Comprehensive terminology audit (2026-01-06)
**Prepared**: 2026-01-06
**Scope**: Achieve 100% Knossos terminology alignment across codebase

## Executive Summary

This refactoring plan addresses systematic terminology drift from Knossos doctrine:
- `team` -> `rite` (primary, ~800+ occurrences)
- `state-mate` -> `moirai` (secondary, ~528 occurrences across 66 files)
- `thread` -> `clew` (tertiary, ~40 occurrences in Go)

**Critical Finding**: The `rites/` directory at `ROSTER_HOME` does NOT exist. The shell scripts reference a non-existent path. The actual rite packs live in `rites/` directory. Track T1 is **already complete** at the filesystem level.

## Architectural Assessment

### Boundary Health

| Component | Status | Assessment |
|-----------|--------|------------|
| `ariadne/internal/team/` | Needs rename | Core Go package with 12 files, 3055 LOC, 402 `team` references |
| `ariadne/internal/rite/` | Already exists | New rite management package (9 files), coexists with team/ |
| `ariadne/internal/paths/` | Clean | Already uses `RitesDir()`, `ActiveRiteFile()`, etc. |
| `lib/team/` | Needs rename | Shell library (3 files, 170 team references) |
| `user-hooks/lib/rite-context-loader.sh` | Needs rename | Hook support file |
| `ariadne/internal/cmd/team/` | Keep for CLI | Command group (user-facing, backward compat) |
| `rites/forge-pack/skills/team-development/` | Needs rename | Template skill |

### Root Causes Identified

1. **Historical Naming**: `team` was original terminology before Knossos doctrine
2. **Partial Migration**: Some paths migrated (ACTIVE_RITE file) but types retained old names
3. **Shell Script Legacy**: `$ROSTER_HOME/teams` path references never existed, code was aspirational
4. **state-mate Documentation**: Alias documented but not systematically applied

### Interface Contracts to Preserve

**MUST Preserve (Public API):**
- CLI command: `ari team` (user-facing, keep for backward compat)
- CLI command: `ari rite` (canonical, already exists)
- JSON output field: `active_team` (manifest schema, external consumers)
- YAML field: `rite_name` in context.yaml (external schema)
- File: `.claude/ACTIVE_RITE` (already canonical)
- File: `.claude/AGENT_MANIFEST.json` with `active_team` field

**MAY Change (Internal):**
- Go type names (RiteContext -> RiteContext)
- Go method names (ActiveTeamName -> ActiveRiteName)
- Error message text ("Team not found" -> "Rite not found")
- Internal variable names
- Shell function names
- Directory names for internal packages

## Track Validation and Refinement

### Original Track Assessment

| Track | Proposed | Validated | Notes |
|-------|----------|-----------|-------|
| T1: Directory rename rites/ -> rites/ | Invalid | **Skip** | `ROSTER_HOME/rites/` does not exist; `rites/` already canonical |
| T2: lib/team/ -> lib/rite/ | Valid | **Execute** | Shell library rename |
| T3-T6: Go type/method renames | Valid | **Execute** | Core refactoring work |
| T7-T9: Shell script renames | Valid | **Execute** | Function and path updates |
| T10-T12: Template updates | Valid | **Execute** | Variable and file renames |
| T13: state-mate -> moirai | Valid | **Execute** | Documentation updates |
| T14: thread -> clew | Valid | **Execute** | Already partially done (clewcontract/) |
| T15-T16: Sync and tests | Valid | **Execute** | Finalization |

### Revised Track Structure

**WAVE 1 removed** - No directory rename needed at filesystem level.

## Refactoring Sequence

---

### Phase 1: Shell Library Cleanup [Low Risk]

**Goal**: Rename shell library without breaking callers

#### RF-T2-001: Rename lib/team/ directory to lib/rite/

**Smells addressed**: Terminology drift in directory structure
**Category**: Local (shell scripts)
**Files affected**: 3 files

**Before State**:
```
lib/team/
  rite-resource.sh
  rite-hooks-registration.sh
  team-transaction.sh
```

**After State**:
```
lib/rite/
  rite-resource.sh
  rite-hooks-registration.sh
  rite-transaction.sh
```

**Invariants**:
- All sourcing scripts continue to load library
- Guard variables updated (_TEAM_RESOURCE_LOADED -> _RITE_RESOURCE_LOADED)
- Function signatures unchanged

**Verification**:
1. `source "$ROSTER_HOME/lib/rite/rite-resource.sh"` succeeds
2. Functions `is_resource_from_team`, `get_resource_team` still callable
3. All shell tests pass

**Commit scope**: Single commit for directory + file renames
**Commit message**: `refactor(shell): rename lib/team/ to lib/rite/ [RF-T2-001]`

#### RF-T2-002: Update lib/rite/ internal terminology

**Files**: `lib/rite/*.sh` (3 files)
**Changes**:
- Function comments: "team" -> "rite" in documentation
- Variable names where not part of public interface
- Guard variable names

**Commit message**: `refactor(shell): update lib/rite/ internal terminology [RF-T2-002]`

#### RF-T2-003: Update callers of lib/team/

**Files affected** (callers):
- `swap-rite.sh`
- `generate-rite-context.sh`
- `sync-user-hooks.sh`
- `sync-user-commands.sh`
- `sync-user-agents.sh`
- `templates/generate-orchestrator.sh`
- `templates/orchestrator-generate.sh`

**Changes**: Update `source` paths from `lib/team/` to `lib/rite/`

**Commit message**: `refactor(shell): update lib/rite/ import paths [RF-T2-003]`

**[ROLLBACK POINT 1]**: Shell library renamed, callers updated. Can stop here safely.

---

### Phase 2: Go Package Rename [Medium Risk]

**Goal**: Rename internal/team/ package while maintaining cmd/team/ for CLI

#### RF-T3-001: Create internal/rite/ package alias strategy

**Assessment**: `internal/rite/` already exists with different functionality (rite manifest handling). Strategy is to:
1. Keep `internal/team/` as the core implementation
2. Gradually rename types within it
3. Do NOT merge with `internal/rite/` (different concerns)

**Decision**: Rename types in-place within `internal/team/` package. The package name `team` for internal use is acceptable as long as exported types use `Rite` terminology.

#### RF-T3-002: Rename exported Go types in internal/team/

**Files**: 12 files in `internal/team/`
**Type renames** (in discovery.go):
- `type Team struct` -> `type Rite struct`
- `type Discovery struct` fields: `activeTeam` -> `activeRite`

**Type renames** (in context.go):
- `type RiteContext struct` -> `type RiteContext struct`

**Type renames** (in manifest.go):
- `Manifest.ActiveTeam` -> `Manifest.ActiveRite`

**Invariants**:
- JSON tags remain unchanged (`"active_team"`) for backward compat
- Method signatures update to use new type names
- All tests updated in same commit

**Verification**:
1. `go build ./...` succeeds
2. `go test ./internal/team/...` passes
3. JSON output unchanged (tags preserved)

**Commit message**: `refactor(go): rename Team type to Rite in internal/team [RF-T3-002]`

#### RF-T4-001: Rename Go methods

**Method renames**:
- `ActiveTeamName()` -> `ActiveRiteName()`
- `GetActiveTeam()` -> `GetActiveRite()` (in cmd/session/, cmd/team/)
- `SetActiveTeam()` -> `SetActiveRite()`
- `IsFromTeam()` -> `IsFromRite()`
- `GetRiteAgents()` -> `GetRiteAgents()`
- `DetectOrphans()` parameters and docs

**Files affected**:
- `internal/team/discovery.go`
- `internal/team/manifest.go`
- `internal/cmd/session/session.go`
- `internal/cmd/team/team.go`
- `internal/cmd/team/switch.go`
- `internal/cmd/team/status.go`
- `internal/output/team.go`

**Commit message**: `refactor(go): rename Team methods to Rite terminology [RF-T4-001]`

#### RF-T5-001: Update Go struct field names (internal only)

**Changes**: Internal struct fields where JSON tag is different
- `backup.activeTeamPath` -> `backup.activeRitePath`
- `backup.activeTeamData` -> `backup.activeRiteData`
- Comments and documentation strings

**Note**: JSON tags like `"active_team"` remain unchanged for API stability

**Commit message**: `refactor(go): update internal field names to rite terminology [RF-T5-001]`

#### RF-T6-001: Update Go error messages

**Files**: `internal/errors/errors.go`, `internal/team/validate.go`
**Changes**:
- `"Team not found: %s"` -> `"Rite not found: %s"`
- `"Team validation failed"` -> `"Rite validation failed"`
- `"Team switch aborted"` -> `"Rite switch aborted"`
- Function names: `ErrTeamNotFound` -> `ErrRiteNotFound`

**Invariants**: Error codes unchanged, only messages

**Commit message**: `refactor(go): update error messages to rite terminology [RF-T6-001]`

**[ROLLBACK POINT 2]**: Go types and methods renamed. Tests must pass.

---

### Phase 3: Shell Script Function Updates [Medium Risk]

**Goal**: Update function names and remaining path references

#### RF-T7-001: Rename user-hooks/lib/rite-context-loader.sh

**Before**: `user-hooks/lib/rite-context-loader.sh`
**After**: `user-hooks/lib/rite-context-loader.sh`

**Internal changes**:
- `TEAM_CONTEXT_SCRIPT_NAME` -> `RITE_CONTEXT_SCRIPT_NAME`
- `TEAM_CONTEXT_FUNCTION_NAME` -> `RITE_CONTEXT_FUNCTION_NAME`
- `load_team_context()` -> `load_rite_context()`
- `team_context_row()` -> `rite_context_row()`

**Note**: Keep backward compat alias `load_team_context()` calling `load_rite_context()`

**Commit message**: `refactor(hooks): rename rite-context-loader.sh to rite-context-loader.sh [RF-T7-001]`

#### RF-T8-001: Update shell function implementations

**Files**: All shell scripts in `lib/rite/`
**Changes**: Internal function names (where not part of public interface)
- Update comments
- Update logging messages

**Commit message**: `refactor(shell): update rite library function documentation [RF-T8-001]`

#### RF-T9-001: Update ROSTER_HOME/teams path references

**Critical Finding**: These references point to non-existent `$ROSTER_HOME/rites/` directory.

**Files affected**:
- `lib/rite/rite-resource.sh` (references `$ROSTER_HOME/teams`)
- Various shell scripts

**Change**: Update to `$ROSTER_HOME/rites/` which is the actual path

**This is a BUG FIX, not just terminology**: The code references a non-existent directory.

**Commit message**: `fix(shell): correct ROSTER_HOME/teams to ROSTER_HOME/rites [RF-T9-001]`

**[ROLLBACK POINT 3]**: Shell scripts fully updated. Integration tests must pass.

---

### Phase 4: Template and Command Updates [Low Risk]

**Goal**: Update user-facing templates and forge commands

#### RF-T10-001: Update template variables in forge-pack

**Directory**: `rites/forge-pack/skills/team-development/`
**Files affected**: 11+ files with `{team-name}` or `{rite_name}` variables

**Changes**:
- `{team-name}` -> `{rite-name}` in templates
- `{rite_name}` -> `{rite_name}` where used
- Update documentation text

**Commit message**: `refactor(templates): update team variables to rite in forge-pack [RF-T10-001]`

#### RF-T11-001: Update user-commands/team-switching/ content

**Files**: 10 files (10x.md, debt.md, docs.md, etc.)
**Changes**: Update any `team` references in command descriptions

**Commit message**: `refactor(commands): update team-switching command descriptions [RF-T11-001]`

#### RF-T12-001: Rename forge-pack skill directory

**Before**: `rites/forge-pack/skills/team-development/`
**After**: `rites/forge-pack/skills/rite-development/`

**Also rename**:
- `rites/forge-pack/commands/new-team.md` -> `new-rite.md`
- `rites/forge-pack/commands/validate-team.md` -> `validate-rite.md`

**Commit message**: `refactor(forge): rename team-development to rite-development [RF-T12-001]`

**[ROLLBACK POINT 4]**: Templates updated. /sync not yet run.

---

### Phase 5: Secondary Terminology [Low Risk]

**Goal**: Update state-mate -> moirai and thread -> clew

#### RF-T13-001: Update state-mate references to moirai

**Scope**: 528 occurrences across 66 files (mostly documentation)
**Strategy**:
- Keep `state-mate` as documented alias in moirai.md
- Update primary references in docs to use `moirai`
- Update CLAUDE.md sections

**Exclusions**:
- ADR-0005 (historical document, keep original terminology)
- Backwards compatibility alias in user-agents/moirai.md

**Commit message**: `docs: update state-mate references to moirai [RF-T13-001]`

#### RF-T14-001: Update thread -> clew in help text

**Note**: `internal/hook/clewcontract/` already uses correct terminology.
**Scope**: ~40 occurrences in Go code, mostly in:
- `internal/sails/` (thread references in comments)
- CLI help text

**Commit message**: `refactor(go): update thread terminology to clew [RF-T14-001]`

**[ROLLBACK POINT 5]**: Secondary terminology complete.

---

### Phase 6: Finalization [Low Risk]

**Goal**: Regenerate files and validate

#### RF-T15-001: Run inscription sync

**Command**: `ari inscription sync`
**Purpose**: Regenerate .claude/ files with updated terminology

**Verification**:
1. CLAUDE.md updated correctly
2. No sync conflicts
3. Agents reflect terminology

**Commit message**: `chore: regenerate .claude/ via inscription sync [RF-T15-001]`

#### RF-T16-001: Integration test validation

**Test suite**:
1. `go test ./...` - All Go tests pass
2. Shell script tests (if any)
3. Manual validation:
   - `ari rite list` works
   - `ari team list` works (backward compat)
   - `/rite` command works
   - Session creation with rite works

**Commit message**: `test: validate doctrine purity refactoring [RF-T16-001]`

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-T2-001 | Low | 3 files | 1 commit | None |
| RF-T2-002 | Low | 3 files | 1 commit | RF-T2-001 |
| RF-T2-003 | Low | 7 files | 1 commit | RF-T2-001 |
| RF-T3-002 | Medium | 12 files | 1 commit | None |
| RF-T4-001 | Medium | 8 files | 1 commit | RF-T3-002 |
| RF-T5-001 | Low | 5 files | 1 commit | RF-T4-001 |
| RF-T6-001 | Low | 2 files | 1 commit | None |
| RF-T7-001 | Medium | 1 file + callers | 1 commit | RF-T2-003 |
| RF-T8-001 | Low | 3 files | 1 commit | RF-T7-001 |
| RF-T9-001 | Medium | 10+ files | 1 commit | RF-T2-001 |
| RF-T10-001 | Low | 11 files | 1 commit | None |
| RF-T11-001 | Low | 10 files | 1 commit | None |
| RF-T12-001 | Medium | 3 directories | 1 commit | RF-T10-001 |
| RF-T13-001 | Low | 66 files | 1 commit | None |
| RF-T14-001 | Low | 6 files | 1 commit | None |
| RF-T15-001 | Low | Generated | N/A | All above |
| RF-T16-001 | N/A | Validation | N/A | RF-T15-001 |

## File Counts by Track

| Track | Files | Lines Changed (est.) |
|-------|-------|---------------------|
| T2 (Shell lib) | 10 | ~200 |
| T3 (Go types) | 12 | ~150 |
| T4 (Go methods) | 8 | ~100 |
| T5 (Go fields) | 5 | ~50 |
| T6 (Go errors) | 2 | ~30 |
| T7 (Hook loader) | 1 + callers | ~80 |
| T8 (Shell funcs) | 3 | ~50 |
| T9 (Path fix) | 10 | ~40 |
| T10 (Templates) | 11 | ~100 |
| T11 (Commands) | 10 | ~30 |
| T12 (Dir rename) | 3 dirs | ~20 |
| T13 (state-mate) | 66 | ~200 |
| T14 (thread) | 6 | ~40 |
| **Total** | **~147 files** | **~1090 lines** |

## Notes for Janitor

### Commit Message Conventions

Format: `{type}({scope}): {description} [RF-{track}-{seq}]`

Types:
- `refactor` - Code structure changes
- `fix` - Bug fixes (like RF-T9-001)
- `docs` - Documentation only
- `chore` - Maintenance tasks
- `test` - Test updates

### Test Run Requirements

After each commit:
1. `go build ./...` must succeed
2. `go test ./...` must pass
3. For shell changes: Source file and verify functions exist

After each phase:
1. Full test suite
2. Manual smoke test of affected commands

### Files to Avoid Touching

- `ariadne/internal/cmd/team/` command files (keep `team` as CLI command)
- JSON schema files with `active_team` (external API)
- ADR documents (historical record)
- Generated files in `.claude/` (will be regenerated in T15)

### Order is Critical For

1. RF-T2-001 must complete before RF-T2-003 (callers need new path)
2. RF-T3-002 must complete before RF-T4-001 (methods use types)
3. RF-T9-001 is a bug fix and should be done with T2 wave
4. RF-T15-001 must be LAST (regenerates from updated sources)

### Backward Compatibility Requirements

1. `ari team` CLI command must continue working
2. JSON `active_team` field must be preserved in output
3. Shell `load_team_context()` should alias to `load_rite_context()`

## Out of Scope

Findings deferred for future work:

1. **CLI command rename**: `ari team` -> `ari rite` as primary (keep team as alias)
   - Reason: User-facing change, needs migration guide

2. **JSON schema migration**: `active_team` -> `active_rite`
   - Reason: Breaking change for external consumers

3. **user-commands/team-switching/ directory rename**
   - Reason: Requires additional coordination with sync scripts

4. **Complete moirai consolidation**
   - Reason: state-mate alias provides backward compatibility

## Success Criteria

The refactoring is complete when:

- [ ] No `$ROSTER_HOME/teams` path references exist (bug fix)
- [ ] All Go types use `Rite` terminology (Team struct -> Rite struct)
- [ ] All Go methods use `Rite` terminology (ActiveTeamName -> ActiveRiteName)
- [ ] Shell library renamed to `lib/rite/`
- [ ] Hook loader renamed to `rite-context-loader.sh`
- [ ] Forge templates use `{rite-name}` variables
- [ ] Error messages say "Rite not found"
- [ ] Documentation uses moirai as primary term
- [ ] All tests pass
- [ ] inscription sync completes without conflict

## Attestation

| Artifact | Verified | Path |
|----------|----------|------|
| Refactoring Plan | Pending | `/Users/tomtenuta/Code/roster/docs/plans/REFACTOR-PLAN-doctrine-purity.md` |
