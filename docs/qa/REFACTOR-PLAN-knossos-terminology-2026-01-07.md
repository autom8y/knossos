# Refactoring Plan: Knossos Terminology Migration Completion

**Plan ID**: RP-KNOSSOS-TERM-2026-01-07
**Generated**: 2026-01-07
**Architect Enforcer**: Claude Code (hygiene-pack)
**Source Report**: `docs/qa/SMELL-REPORT-knossos-terminology-migration-2026-01-07.md`

---

## Executive Summary

This plan addresses 47 terminology migration smells identified by Code Smeller. The migration from "team" to "rite" terminology is incomplete across schemas, Go code, shell scripts, and documentation.

**Critical Constraint**: User specified **zero backward compatibility** - clean break allowed. This eliminates deprecation period requirements and simplifies execution.

**Key Findings**:
1. No existing SESSION_CONTEXT.md files use legacy `team:` key (verified via grep)
2. Two session-context schemas exist - one clean, one with legacy fields
3. Shell script backward-compat code can be removed entirely
4. 47 smells across 68+ files, groupable into 4 phases

---

## Architectural Assessment

### Boundary Health

| Boundary | Health | Issue |
|----------|--------|-------|
| Schema Layer | UNHEALTHY | Self-violating keys (`"team"` with `"description": "Rite name"`) |
| Go Struct Layer | DEGRADED | `MigrationInfo.FromTeam` propagates legacy into serialization |
| Shell Script Layer | DEGRADED | Backward-compat code for never-used fields |
| Documentation Layer | STALE | 150+ occurrences of legacy terminology |
| Semantic Layer | CONFUSED | "rite roster" vs "pantheon" terminology inconsistent |

### Root Cause Clusters

1. **Incomplete Schema Migration**: Descriptions updated but keys left unchanged
2. **Dead Backward Compat**: Shell scripts support `active_team` but no data uses it
3. **Semantic Drift**: "rite roster" used when "pantheon" is canonical term for agent collections
4. **Documentation Lag**: Docs not updated after code migrations

### Breaking Changes Requiring ADR

| Change | ADR Required | Rationale |
|--------|--------------|-----------|
| SM-015: `cross_team_protocol` -> `cross_rite_protocol` | **YES** | Breaking schema change affecting all orchestrator.yaml files |
| SM-001/002: Remove `team` from schemas | NO | Deprecated field never used in production |
| SM-003: Go struct tag change | NO | Internal migration metadata, no external consumers |
| SM-008: `team_mappings` -> `rite_mappings` | NO | Internal config file |

---

## Phase Overview

| Phase | Severity | Files | Risk | Rollback Cost |
|-------|----------|-------|------|---------------|
| 1 | BLOCKER | 2 | LOW | Single commit revert |
| 2 | HIGH | 12 | MEDIUM | Multiple commit reverts |
| 3 | MEDIUM | 18 | LOW | Single commit revert |
| 4 | LOW | 36+ | MINIMAL | Bulk revert |

---

## Phase 1: BLOCKER Resolution (Schema Self-Violations)

**Risk Level**: LOW
**Blast Radius**: Schema validation, Go code parsing schemas
**Rollback Point**: After Phase 1 commit

### RF-001: Remove duplicate `team` field from embedded session-context schema

**Smell ID**: SM-001, SM-016
**File**: `/Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/session-context.schema.json`

**Before State**:
```json
"active_rite": {
  "type": "string",
  "description": "Active rite name"
},
"team": {
  "oneOf": [
    {"type": "string"},
    {"type": "null"}
  ],
  "description": "Rite name (null for cross-cutting sessions)"
}
```

**After State**:
```json
"active_rite": {
  "type": "string",
  "description": "Active rite name"
}
```
(Field `team` removed entirely - lines 54-60)

**Invariants**:
- No SESSION_CONTEXT.md files use `team:` key (verified)
- `active_rite` remains required field
- All existing session contexts validate against updated schema

**Verification**:
1. Run: `cd ariadne && go test ./internal/validation/...`
2. Confirm schema loads without error
3. Validate sample SESSION_CONTEXT.md files

**Rollback**: `git revert <commit-hash>`

---

### RF-002: Remove deprecated `team` field from sprint-context schema

**Smell ID**: SM-002
**File**: `/Users/tomtenuta/Code/roster/schemas/artifacts/sprint-context.schema.json`

**Before State** (lines 47-50):
```json
"team": {
  "$ref": "common.schema.json#/$defs/rite_name",
  "description": "Alias for active_rite (deprecated, use active_rite)"
}
```

**After State**:
(Field removed entirely)

**Invariants**:
- `active_rite` remains in required array
- Sprint context examples use only `active_rite`
- No SPRINT_CONTEXT.md files use `team:` key

**Verification**:
1. Run schema validation on existing SPRINT_CONTEXT.md files
2. Confirm no validation errors

**Rollback**: `git revert <commit-hash>`

---

### Phase 1 Commit Contract

```
Commit message: fix(schemas): remove deprecated team field from session/sprint schemas

Files changed:
- ariadne/internal/validation/schemas/session-context.schema.json (remove lines 54-60)
- schemas/artifacts/sprint-context.schema.json (remove lines 47-50)

Tests to run before commit:
- cd ariadne && go test ./internal/validation/...
```

---

## Phase 2: HIGH Severity (API/Struct Changes)

**Risk Level**: MEDIUM
**Blast Radius**: Go serialization, documentation examples, skill descriptions
**Rollback Point**: After Phase 2 complete (multi-commit)

### RF-003: Rename Go struct field `FromTeam` to `FromRite`

**Smell ID**: SM-003
**File**: `/Users/tomtenuta/Code/roster/ariadne/internal/rite/manifest.go`

**Before State** (lines 121-125):
```go
// MigrationInfo contains migration metadata from legacy naming.
type MigrationInfo struct {
    FromTeam   string `yaml:"from_team,omitempty" json:"from_team,omitempty"`
    MigratedAt string `yaml:"migrated_at,omitempty" json:"migrated_at,omitempty"`
}
```

**After State**:
```go
// MigrationInfo contains migration metadata from legacy naming.
type MigrationInfo struct {
    FromRite   string `yaml:"from_rite,omitempty" json:"from_rite,omitempty"`
    MigratedAt string `yaml:"migrated_at,omitempty" json:"migrated_at,omitempty"`
}
```

**Invariants**:
- No rite.yaml files currently use `migration:` block with `from_team`
- Struct is internal, not exposed via public API
- YAML/JSON output format changes (acceptable per clean break)

**Verification**:
1. Run: `cd ariadne && go build ./...`
2. Run: `cd ariadne && go test ./...`
3. Grep for `from_team` references in test fixtures

**Rollback**: `git revert <commit-hash>`

---

### RF-004: Update documentation examples from `source: "team"` to `source: "rite"`

**Smell ID**: SM-004
**Files**:
- `/Users/tomtenuta/Code/roster/schemas/INTEGRATION-POINTS.md` (line 146)
- `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-rite.md` (lines 627, 633)
- `/Users/tomtenuta/Code/roster/docs/design/TDD-hook-parity-scope1.md` (line 344)
- `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/architecture-overview.md` (line 327)

**Before State**:
```json
"architect.md": {"source": "team", "origin": "10x-dev", ...}
```

**After State**:
```json
"architect.md": {"source": "rite", "origin": "10x-dev", ...}
```

**Invariants**:
- Schema enum already uses `["rite", "project"]`
- Only documentation examples affected
- No runtime behavior change

**Verification**:
1. Grep: `rg '"source".*"team"'` returns 0 matches

**Rollback**: `git revert <commit-hash>`

---

### RF-005: Replace "rite roster" with "pantheon" in rite-switching commands

**Smell ID**: SM-005
**Files** (22 occurrences):
- `/Users/tomtenuta/Code/roster/user-commands/rite-switching/*.md` (6 files, line 13 each)
- `/Users/tomtenuta/Code/roster/user-commands/navigation/ecosystem.md` (line 13)

**Before State**:
```markdown
Switch to the 10x development rite and display the rite roster.
```

**After State**:
```markdown
Switch to the 10x development rite and display the pantheon.
```

**Invariants**:
- Semantic accuracy: "pantheon" = agent collection within rite
- Consistent with knossos doctrine terminology table

**Verification**:
1. Grep: `rg 'rite roster' user-commands/` returns 0 matches

**Rollback**: `git revert <commit-hash>`

---

### RF-006: Replace "Display Team Roster" headers with "Display Pantheon"

**Smell ID**: SM-006, SM-007
**Files**:
- `/Users/tomtenuta/Code/roster/rites/*/skills/*-ref/skill.md` (8 files, line 40 pattern)
- `/Users/tomtenuta/Code/roster/rites/strategy/skills/strategy-ref/skill.md` (line 18)
- `/Users/tomtenuta/Code/roster/rites/rnd/skills/rnd-ref/skill.md` (line 18)
- `/Users/tomtenuta/Code/roster/rites/security/skills/security-ref/skill.md` (line 18)
- `/Users/tomtenuta/Code/roster/rites/intelligence/skills/intelligence-ref/skill.md` (line 18)

**Before State**:
```markdown
### 2. Display Team Roster
```
or
```markdown
## Team Roster
```

**After State**:
```markdown
### 2. Display Pantheon
```
or
```markdown
## Pantheon
```

**Invariants**:
- Section content unchanged
- Only header text updated

**Verification**:
1. Grep: `rg 'Team Roster' rites/` returns 0 matches

**Rollback**: `git revert <commit-hash>`

---

### Phase 2 Commit Contract

Phase 2 should be split into 4 atomic commits:

```
Commit 1: refactor(ariadne): rename FromTeam to FromRite in MigrationInfo struct
Commit 2: docs: update source enum examples from "team" to "rite"
Commit 3: docs(commands): replace "rite roster" with "pantheon" terminology
Commit 4: docs(skills): replace "Team Roster" headers with "Pantheon"
```

---

## Phase 3: MEDIUM Severity (Config/Scripts)

**Risk Level**: LOW
**Blast Radius**: YAML configs, shell scripts, orchestrator schema
**Rollback Point**: After Phase 3 complete

### RF-007: Rename `team_mappings` to `rite_mappings` in complexity scale

**Smell ID**: SM-008
**File**: `/Users/tomtenuta/Code/roster/schemas/complexity-scale-mapping.yaml`

**Before State**:
```yaml
# Maps team-specific complexity levels to meta-scale for cross-team coordination
team_mappings:
  # Development teams
  10x-dev:
    ...
```

**After State**:
```yaml
# Maps rite-specific complexity levels to meta-scale for cross-rite coordination
rite_mappings:
  # Development rites
  10x-dev:
    ...
```

**Invariants**:
- All comments updated: "team" -> "rite"
- Key names updated: `team_mappings` -> `rite_mappings`
- Value structure unchanged

**Changes Required**:
- Line 2: `team-specific` -> `rite-specific`, `cross-team` -> `cross-rite`
- Line 29: `team_mappings:` -> `rite_mappings:`
- Lines 30, 37, 44, 52, 59, 65, 73, 79, 86, 93, 100: `# Development teams` -> `# Development rites` etc.
- Line 107: `Cross-team routing rules` -> `Cross-rite routing rules`
- Line 113-114: `Cross-team`, `affects_multiple_teams` -> `Cross-rite`, `affects_multiple_rites`

**Verification**:
1. Grep: `rg 'team' schemas/complexity-scale-mapping.yaml` returns 0 matches
2. YAML parses without error

**Rollback**: `git revert <commit-hash>`

---

### RF-008: Remove dead `active_team` backward-compat code from shell scripts

**Smell ID**: SM-011
**Files**:
- `/Users/tomtenuta/Code/roster/user-hooks/validation/command-validator.sh`
- `/Users/tomtenuta/Code/roster/user-hooks/lib/session-fsm.sh`
- `/Users/tomtenuta/Code/roster/user-hooks/lib/session-state.sh`
- `/Users/tomtenuta/Code/roster/user-hooks/lib/session-manager.sh`

**Before State**:
```bash
# Backward compat: check both active_rite and active_team
SESSION_TEAM=$(grep -m1 "^active_team:" "$SESSION_FILE" 2>/dev/null | cut -d: -f2- | tr -d ' "')
```

**After State**:
```bash
# Only active_rite is supported
```
(Remove backward-compat branches entirely)

**Invariants**:
- No SESSION_CONTEXT.md files use `active_team:` key (verified)
- Only `active_rite` parsing retained
- Backward-compat comments removed

**Verification**:
1. Grep: `rg 'active_team' user-hooks/` returns 0 matches
2. Run hook tests (if they exist)
3. Test session creation/parsing manually

**Rollback**: `git revert <commit-hash>`

---

### RF-009: Update legacy path references `/roster/teams/` to `/roster/rites/`

**Smell ID**: SM-010
**Files**:
- `/Users/tomtenuta/Code/roster/.gitignore`
- `/Users/tomtenuta/Code/roster/schemas/INTEGRATION-POINTS.md`
- `/Users/tomtenuta/Code/roster/.github/workflows/validate-orchestrators.yml`

**Before State**:
```
/roster/teams/{team-name}/agents/orchestrator.md
```

**After State**:
```
/roster/rites/{rite-name}/agents/orchestrator.md
```

**Invariants**:
- Paths match actual directory structure
- CI workflow uses correct paths

**Verification**:
1. Grep: `rg 'roster/teams' .` returns 0 matches
2. CI workflow runs successfully

**Rollback**: `git revert <commit-hash>`

---

### RF-010: Rename Go test fixtures from `test-team` to `test-rite`

**Smell ID**: SM-012, SM-013
**Files**:
- `/Users/tomtenuta/Code/roster/ariadne/internal/rite/context_loader_test.go`
- `/Users/tomtenuta/Code/roster/ariadne/internal/rite/discovery_test.go`

**Before State**:
```go
ctx := NewRiteContext("test-team")
teamsDir := filepath.Join(tempDir, "teams")
```

**After State**:
```go
ctx := NewRiteContext("test-rite")
ritesDir := filepath.Join(tempDir, "rites")
```

**Invariants**:
- Test behavior unchanged
- Only naming/terminology updated

**Verification**:
1. Run: `cd ariadne && go test ./internal/rite/...`
2. All tests pass

**Rollback**: `git revert <commit-hash>`

---

### RF-011: Update Go code comments from "team" to "rite"

**Smell ID**: SM-014
**Files**:
- `/Users/tomtenuta/Code/roster/ariadne/internal/rite/context_loader.go`
- `/Users/tomtenuta/Code/roster/ariadne/internal/rite/switch.go`

**Before State**:
```go
// Switcher handles team switching operations.
type Switcher struct { ... }

// NewSwitcher creates a new team switcher.
func NewSwitcher() *Switcher { ... }
```

**After State**:
```go
// Switcher handles rite switching operations.
type Switcher struct { ... }

// NewSwitcher creates a new rite switcher.
func NewSwitcher() *Switcher { ... }
```

**Invariants**:
- Code logic unchanged
- Only comments updated

**Verification**:
1. Run: `cd ariadne && go build ./...`

**Rollback**: `git revert <commit-hash>`

---

### RF-012: **ADR REQUIRED** - Rename `cross_team_protocol` to `cross_rite_protocol`

**Smell ID**: SM-015
**File**: `/Users/tomtenuta/Code/roster/schemas/orchestrator.yaml.schema.json`

**ADR Required**: YES - This is a breaking schema change affecting:
- All orchestrator.yaml files
- Schema consumers
- Documentation

**ADR Outline**:
```
ADR-XXXX: Rename cross_team_protocol to cross_rite_protocol

Status: Proposed
Context: Knossos terminology migration requires consistent "rite" usage
Decision: Rename field, update all orchestrator.yaml files
Consequences: Breaking change, all rites must update orchestrator.yaml
```

**Before State** (lines 181-188):
```json
"cross_team_protocol": {
  "type": "string",
  "description": "OPTIONAL: Cross-team handoff protocol..."
}
```

**After State**:
```json
"cross_rite_protocol": {
  "type": "string",
  "description": "OPTIONAL: Cross-rite handoff protocol..."
}
```

**Also update**:
- Lines 255-258: Example using `cross_team_protocol`
- Lines 317, 320, 374, 423: Additional references
- All `rites/*/orchestrator.yaml` files that use this field

**Invariants**:
- All orchestrator.yaml files updated simultaneously
- Schema validation passes for all rites

**Verification**:
1. Run: orchestrator validation workflow
2. All rites pass schema validation

**Rollback**: `git revert <commit-hash>` (includes schema + all orchestrator.yaml changes)

---

### Phase 3 Commit Contract

```
Commit 1: refactor(schemas): rename team_mappings to rite_mappings in complexity scale
Commit 2: refactor(hooks): remove dead active_team backward-compat code
Commit 3: fix(paths): update roster/teams references to roster/rites
Commit 4: refactor(ariadne): update test fixtures to use rite terminology
Commit 5: docs(ariadne): update code comments from team to rite
Commit 6: feat(schemas)!: rename cross_team_protocol to cross_rite_protocol [ADR-XXXX]
```

**Note**: Commit 6 is a BREAKING change and requires ADR creation first.

---

## Phase 4: LOW Severity (Bulk Documentation)

**Risk Level**: MINIMAL
**Blast Radius**: Documentation only
**Rollback Point**: Single bulk commit

### RF-013: Bulk documentation terminology updates

**Smell ID**: SM-017 through SM-044
**Files**: 36+ documentation files

**Patterns to Replace**:

| Pattern | Replacement | Files Affected |
|---------|-------------|----------------|
| `team roster` | `pantheon` | 45+ occurrences |
| `active_team` in docs | `active_rite` | 150+ occurrences |
| `swap-team` in active refs | `swap-rite` | 30+ occurrences |
| `--team` flag | `--rite` flag | 10 occurrences |

**Special Handling**:

1. **Historical Documents**: Add `[HISTORICAL]` marker instead of updating
   - Migration plans that document the team->rite transition
   - Gap analyses that reference old terminology

2. **Active Documents**: Full terminology update
   - Skill files
   - Command references
   - User guides

**Invariants**:
- Historical context preserved where appropriate
- Active documentation uses current terminology
- Cross-references updated

**Verification**:
1. Grep: `rg 'team roster' docs/` returns only `[HISTORICAL]` marked files
2. Grep: `rg 'active_team' docs/` returns only `[HISTORICAL]` marked files

**Rollback**: `git revert <commit-hash>`

---

### RF-014: Update hygiene-ref skill with terminology fixes

**Smell ID**: Related to SM-005
**File**: `/Users/tomtenuta/Code/roster/rites/hygiene/skills/hygiene-ref/skill.md`

This file has extensive "team" terminology that should be "rite" or "pantheon":

**Changes Required**:
- Line 8: `Team Management` -> `Rite Management`
- Line 10: `hygiene, a specialized team` -> `hygiene, a specialized rite`
- Line 27: `Displays team roster` -> `Displays pantheon`
- Line 40: `Display Team Roster` -> `Display Pantheon`
- Line 42: `active rite roster` -> `active pantheon`
- Line 47: `Team Roster:` -> `Pantheon:`
- Line 63: `active_team` -> `active_rite`
- Line 69-70: `Team Name` -> `Rite Name`
- And many more throughout the file

**Invariants**:
- Functional behavior unchanged
- Only terminology updated

**Verification**:
1. Grep: `rg 'team' rites/hygiene/skills/hygiene-ref/skill.md` returns 0 inappropriate matches

---

### Phase 4 Commit Contract

```
Commit 1: docs: bulk terminology update - team to rite/pantheon
Commit 2: docs: mark historical migration documents with [HISTORICAL] tag
```

---

## Risk Matrix

| Phase | Blast Radius | Failure Detection | Recovery Path | Estimated Time |
|-------|--------------|-------------------|---------------|----------------|
| 1 | Schema validation | Go tests fail | Single revert | 15 min |
| 2 | Serialization | Go build/test | Per-commit revert | 45 min |
| 3 | Scripts, CI | Manual test, CI | Per-commit revert | 60 min |
| 4 | Documentation | Visual review | Single bulk revert | 90 min |

**Total Estimated Time**: 3.5 hours

---

## Janitor Notes

### Commit Conventions

All commits should follow:
```
<type>(<scope>): <description>

Types: fix, refactor, docs, feat (for breaking changes)
Scope: schemas, ariadne, hooks, skills, docs

For breaking changes: feat(schemas)!: description [ADR-XXXX]
```

### Test Requirements Per Phase

| Phase | Required Tests |
|-------|----------------|
| 1 | `go test ./internal/validation/...` |
| 2 | `go build ./...`, `go test ./...` |
| 3 | `go test ./internal/rite/...`, manual hook test |
| 4 | Grep verification only |

### Critical Ordering

1. **Phase 1 MUST complete before Phase 2** - Schema cleanup enables struct changes
2. **ADR for RF-012 MUST be created before executing RF-012** - Breaking change requires documentation
3. **Phase 3 RF-012 MUST update schema AND orchestrator.yaml files atomically** - Partial update breaks validation

### Files NOT to Modify

- `/Users/tomtenuta/Code/roster/schemas/artifacts/session-context.schema.json` - Already clean
- `/Users/tomtenuta/Code/roster/schemas/artifacts/common.schema.json` - Already uses `rite_name`

---

## Verification Attestation

| File | Path | Status |
|------|------|--------|
| SMELL-REPORT | /Users/tomtenuta/Code/roster/docs/qa/SMELL-REPORT-knossos-terminology-migration-2026-01-07.md | READ |
| session-context.schema.json (ariadne) | /Users/tomtenuta/Code/roster/ariadne/internal/validation/schemas/session-context.schema.json | READ |
| session-context.schema.json (artifacts) | /Users/tomtenuta/Code/roster/schemas/artifacts/session-context.schema.json | READ |
| sprint-context.schema.json | /Users/tomtenuta/Code/roster/schemas/artifacts/sprint-context.schema.json | READ |
| manifest.go | /Users/tomtenuta/Code/roster/ariadne/internal/rite/manifest.go | READ |
| complexity-scale-mapping.yaml | /Users/tomtenuta/Code/roster/schemas/complexity-scale-mapping.yaml | READ |
| orchestrator.yaml.schema.json | /Users/tomtenuta/Code/roster/schemas/orchestrator.yaml.schema.json | READ |
| 10x.md (rite-switching) | /Users/tomtenuta/Code/roster/user-commands/rite-switching/10x.md | READ |
| hygiene-ref skill.md | /Users/tomtenuta/Code/roster/rites/hygiene/skills/hygiene-ref/skill.md | READ |
| common.schema.json | /Users/tomtenuta/Code/roster/schemas/artifacts/common.schema.json | READ |
| Existing SESSION_CONTEXT.md files | .sos/sessions/**/*CONTEXT*.md | GREP (no legacy keys found) |

---

## Handoff Criteria

Ready for Janitor execution when:
- [x] Every smell classified (addressed, deferred, or dismissed)
- [x] Each refactoring has before/after contract documented
- [x] Invariants and verification criteria specified
- [x] Refactorings sequenced with explicit dependencies
- [x] Rollback points identified between phases
- [x] Risk assessment complete for each phase
- [x] ADR requirement identified for RF-012

---

## Summary

| Metric | Value |
|--------|-------|
| Total Smells | 47 |
| Smells Addressed | 47 |
| Smells Deferred | 0 |
| Phases | 4 |
| Commits Planned | 12 |
| ADRs Required | 1 (RF-012) |
| Breaking Changes | 1 (`cross_team_protocol` -> `cross_rite_protocol`) |
| Estimated Time | 3.5 hours |

---

*Plan generated by Architect Enforcer agent following `@refactoring-plan-template` protocols.*
