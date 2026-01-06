# Refactoring Plan: Doctrine Purity Sprint Phase 2

**Based on**: Doctrine Purity Sprint Phase 2 Breaking Changes
**Prepared**: 2026-01-06
**Scope**: CLI command rename, JSON schema migration, directory rename

## Executive Summary

This plan covers three breaking changes required to complete the doctrine purity migration from "team" terminology to "rite" terminology. All three tracks are independent and can be executed in parallel, but the recommended sequence minimizes risk.

## Architectural Assessment

### Boundary Health

- **CLI Layer** (`ariadne/internal/cmd/`): Clean separation. Both `team/` and `rite/` command groups exist. The `team` command is already marked as legacy in its help text.
- **Output Layer** (`ariadne/internal/output/`): Mixed state. `team.go` uses `active_team` JSON tags, `rite.go` correctly uses `active_rite`.
- **Domain Layer** (`ariadne/internal/team/`): Uses Go struct field names with `Rite` suffix but preserves `team` JSON tags for backward compatibility.
- **Schema Layer** (`ariadne/internal/validation/schemas/`): Still requires `active_team` field.
- **User Commands** (`user-commands/team-switching/`): Directory name contains legacy terminology.

### Root Causes Identified

1. **Phased migration**: The codebase underwent partial migration (Go field names changed, JSON tags preserved for compatibility). This is intentional technical debt.
2. **Schema coupling**: JSON schemas and YAML fixtures depend on `active_team` field names.
3. **Path coupling**: Documentation and scripts reference `user-commands/team-switching/` path.

### Current State Summary

| Component | Go Field Name | JSON Tag | Status |
|-----------|---------------|----------|--------|
| `team/manifest.go:Manifest.ActiveRite` | `ActiveRite` | `active_team` | Needs migration |
| `team/switch.go:SwitchResult.Rite` | `Rite` | `team` | Needs migration |
| `team/switch.go:SwitchResult.PreviousRite` | `PreviousRite` | `previous_team` | Needs migration |
| `team/switch.go:DryRunResult.CurrentTeam` | `CurrentTeam` | `current_team` | Needs migration |
| `output/team.go:TeamListOutput.ActiveTeam` | `ActiveTeam` | `active_team` | Needs migration |
| `output/team.go:TeamSwitchDryRunOutput.CurrentTeam` | `CurrentTeam` | `current_team` | Needs migration |
| `output/team.go:TeamStatusOutput.Team` | `Team` | `team` | Keep (describes team entity) |
| `output/rite.go:RiteSwapOutput.Team` | `Team` | `team` | Needs migration to `rite` |
| `output/rite.go:RiteSwapOutput.PreviousTeam` | `PreviousTeam` | `previous_team` | Needs migration |
| `output/output.go:StatusOutput.ActiveTeam` | `ActiveTeam` | `active_team` | Needs migration |
| `cmd/handoff/status.go:HandoffStatusOutput.ActiveTeam` | `ActiveTeam` | `active_team` | Needs migration |

## Refactoring Sequence

### Phase 1: Directory Rename [Low Risk]

**Goal**: Rename `user-commands/team-switching/` to `user-commands/rite-switching/`

#### RF-P2-001: Rename user-commands directory

- **Category**: Local (filesystem only)
- **Before State**:
  - Directory: `user-commands/team-switching/` (10 files)
  - Referenced in: `sync-user-commands.sh`, `README.md`, various docs
- **After State**:
  - Directory: `user-commands/rite-switching/` (10 files)
  - All references updated
- **Invariants**:
  - All 10 command files present after rename
  - No broken symlinks
  - File contents unchanged
- **Verification**:
  ```bash
  ls user-commands/rite-switching/*.md | wc -l  # Should be 10
  grep -r "team-switching" user-commands/ docs/ sync-user-commands.sh README.md  # Should be 0
  ```
- **Commit scope**: Single commit with git mv + reference updates

**Files to modify**:
1. `git mv user-commands/team-switching/ user-commands/rite-switching/`
2. `sync-user-commands.sh` (line 862)
3. `README.md` (line 102)
4. `docs/briefs/categorical-reorg-architect-brief.md` (lines 64, 94)
5. `docs/reports/VERIFICATION-remediation-sprint-2026-01-03.md` (line 39)
6. `docs/plans/REFACTOR-PLAN-doctrine-purity.md` (lines 295, 300, 474)
7. `docs/qa/AUDIT-REPORT-command-template-hygiene.md` (multiple lines)
8. `docs/qa/SMELL-REPORT-command-template-hygiene.md` (multiple lines)
9. `docs/qa/REFACTOR-PLAN-command-template-hygiene.md` (multiple lines)
10. `docs/ecosystem/CONTEXT-DESIGN-command-argument-standardization.md` (multiple lines)
11. `docs/ecosystem/GAP-ANALYSIS-command-argument-standardization.md` (multiple lines)
12. `docs/decisions/ADR-0006-categorical-resource-organization.md` (lines 105, 304)
13. `docs/design/TDD-categorical-resource-organization.md` (lines 96, 261, 464)

**Commit message**: `refactor(commands): rename team-switching/ to rite-switching/ [RF-P2-001]`

[Rollback point: `git checkout HEAD~1 -- user-commands/ && git mv user-commands/rite-switching user-commands/team-switching`]

---

### Phase 2: JSON Schema Migration [Medium Risk]

**Goal**: Update all JSON output fields from `active_team`/`team`/`previous_team`/`current_team` to `active_rite`/`rite`/`previous_rite`/`current_rite`

This phase has multiple sub-tasks that must be done atomically to prevent inconsistent JSON output.

#### RF-P2-002: Update JSON schema files

- **Category**: Schema
- **Before State**:
  - `session-context.schema.json`: requires `active_team`
  - `agent-manifest.schema.json`: requires `active_team`
- **After State**:
  - Both schemas require `active_rite`
  - Both schemas accept `active_team` as optional (backward read compatibility)
- **Invariants**:
  - Schema validation passes for existing sessions with `active_team`
  - New sessions use `active_rite`
- **Files**:
  - `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/session-context.schema.json`
  - `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/schemas/agent-manifest.schema.json`

#### RF-P2-003: Update domain layer JSON tags

- **Category**: Domain
- **Before State**:
  ```go
  // team/manifest.go:19
  ActiveRite  string  `json:"active_team"`

  // team/switch.go:46-47
  Rite           string  `json:"team"`
  PreviousRite   string  `json:"previous_team"`

  // team/switch.go:76
  CurrentTeam    string  `json:"current_team"`
  ```
- **After State**:
  ```go
  ActiveRite  string  `json:"active_rite"`
  Rite           string  `json:"rite"`
  PreviousRite   string  `json:"previous_rite"`
  CurrentRite    string  `json:"current_rite"`
  ```
- **Files**:
  - `/Users/tomtenuta/Code/roster/ariadne/internal/team/manifest.go` (line 19)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/team/switch.go` (lines 46, 47, 76)

#### RF-P2-004: Update output layer JSON tags

- **Category**: Output
- **Before State**: See table above
- **After State**: All `_team` suffixes become `_rite`
- **Files**:
  - `/Users/tomtenuta/Code/roster/ariadne/internal/output/team.go` (lines 16, 178)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/output/rite.go` (lines 314, 315)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/output/output.go` (line 221)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/handoff/status.go` (line 315)

#### RF-P2-005: Update test fixtures and schema validation code

- **Category**: Test
- **Files**:
  - `/Users/tomtenuta/Code/roster/ariadne/test/hooks/fixtures/session_context.yaml` (line 8)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/schema_test.go` (lines 71, 79)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/schema.go` (lines 187-190)
  - `/Users/tomtenuta/Code/roster/ariadne/internal/worktree/session_integration.go` (line 401 - already handles both)

**Verification**:
```bash
cd ariadne && go test ./...
ari session status -o json | jq '.active_rite'  # Should work
ari rite list -o json | jq '.active_rite'  # Should work
```

**Commit message**: `refactor(json): migrate active_team to active_rite in JSON output [RF-P2-002-005]`

[Rollback point: Revert single commit]

---

### Phase 3: CLI Command Deprecation [Low Risk]

**Goal**: Add deprecation warning when `ari team` is used

**Note**: Analysis shows `ari team` already has deprecation messaging in its help text (lines 31-36 of team.go). This phase only needs to add a runtime deprecation warning.

#### RF-P2-006: Add deprecation warning to team command

- **Category**: CLI
- **Before State**:
  - `ari team` works silently
  - Help text mentions it's legacy
- **After State**:
  - `ari team` prints deprecation warning to stderr
  - Format: `Warning: 'ari team' is deprecated. Use 'ari rite' instead.`
- **Invariants**:
  - Command still functions identically
  - Warning goes to stderr, not stdout (so JSON output is clean)
  - Exit codes unchanged
- **Files**:
  - `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/team/team.go`
- **Verification**:
  ```bash
  ari team list 2>&1 | grep -i deprecated  # Should find warning
  ari team list -o json 2>/dev/null | jq .  # JSON should be valid
  ```

**Commit message**: `refactor(cli): add deprecation warning to ari team command [RF-P2-006]`

[Rollback point: Revert single commit]

---

## Risk Matrix

| Refactor | Risk | Blast Radius | Rollback Cost | Dependencies |
|----------|------|--------------|---------------|--------------|
| RF-P2-001 | Low | Docs only | Trivial (git mv) | None |
| RF-P2-002-005 | Medium | JSON consumers | 1 commit | None |
| RF-P2-006 | Low | CLI only | Trivial | None |

## Deprecation Message Format

Standard format for all deprecation warnings:

```
Warning: 'ari team' is deprecated and will be removed in v2.0. Use 'ari rite' instead.
```

Properties:
- Printed to stderr
- Single line
- Includes version target
- Provides migration path

## Success Criteria

### Track 1: Directory Rename
- [ ] `user-commands/rite-switching/` exists with 10 files
- [ ] `user-commands/team-switching/` does not exist
- [ ] `grep -r "team-switching" .` returns only historical docs (if any)
- [ ] `sync-user-commands.sh` references `rite-switching`

### Track 2: JSON Schema Migration
- [ ] `go test ./...` passes in ariadne directory
- [ ] `ari session status -o json | jq '.active_rite'` returns value
- [ ] `ari rite list -o json | jq '.active_rite'` returns value
- [ ] `ari team switch hygiene -o json | jq '.rite'` returns value
- [ ] No JSON output contains `active_team` (except for backward-compat reads)

### Track 3: CLI Deprecation
- [ ] `ari team list 2>&1 | grep -i deprecated` finds warning
- [ ] `ari team list -o json 2>/dev/null | jq .` produces valid JSON
- [ ] `ari rite list` produces no deprecation warning

## Notes for Janitor

### Commit Conventions
- Prefix: `refactor(scope):` where scope is `commands`, `json`, or `cli`
- Include task ID in brackets: `[RF-P2-XXX]`
- Keep commits atomic per phase

### Test Run Requirements
- After RF-P2-001: `sync-user-commands.sh --help` should work
- After RF-P2-002-005: `cd ariadne && go test ./...`
- After RF-P2-006: Manual verification of deprecation warning

### Files to Avoid Touching
- `ariadne/internal/rite/` - Already uses correct terminology
- `ariadne/internal/cmd/rite/` - Already uses correct terminology
- Files with `// Keep JSON tag for backward compatibility` comments need tag updates, not comment removal

### Order Dependencies
- RF-P2-002 through RF-P2-005 MUST be in same commit (atomic JSON change)
- RF-P2-001 and RF-P2-006 can be done in any order
- Recommended sequence: RF-P2-001 -> RF-P2-002-005 -> RF-P2-006

## Out of Scope

Findings deferred for future work:

1. **`json:"team"` tags in worktree and tribute commands**: These describe the entity type, not the field name. The field semantically means "which team/rite". Future work could evaluate whether these should also migrate, but they are not part of the `active_team` pattern.

2. **Full removal of `ari team` command**: This plan adds deprecation warning. Actual removal requires a major version bump and migration period.

3. **Schema validation backward compatibility**: The schema changes in RF-P2-002 may need `oneOf` patterns to accept both `active_team` and `active_rite` for reading existing sessions. This is a design decision for the Janitor.

## Artifact Attestation

| Artifact | Path | Verified |
|----------|------|----------|
| Team command | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/team/team.go` | Read |
| Rite command | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/rite/rite.go` | Read |
| Root command | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/root/root.go` | Read |
| Team output | `/Users/tomtenuta/Code/roster/ariadne/internal/output/team.go` | Read |
| Rite output | `/Users/tomtenuta/Code/roster/ariadne/internal/output/rite.go` | Read |
| Common output | `/Users/tomtenuta/Code/roster/ariadne/internal/output/output.go` | Read |
| Manifest | `/Users/tomtenuta/Code/roster/ariadne/internal/team/manifest.go` | Read |
| Switch | `/Users/tomtenuta/Code/roster/ariadne/internal/team/switch.go` | Read |
| Handoff status | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/handoff/status.go` | Read |
| Session schema | `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/session-context.schema.json` | Read |
| Agent manifest schema | `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/schemas/agent-manifest.schema.json` | Read |
| team-switching dir | `/Users/tomtenuta/Code/roster/user-commands/team-switching/` | Listed |
