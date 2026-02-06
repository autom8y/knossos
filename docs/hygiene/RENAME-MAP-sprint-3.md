# Sprint 3 Rename Map

**Initiative**: Knossos Code Hygiene - Terminology Sweep
**Phase**: Sprint 3 - Naming Consistency
**Date**: 2026-02-06
**Architect**: architect-enforcer
**Upstream**: Sprint 1 deleted 1,272 LOC dead code. Sprint 2 consolidated 345 LOC duplication.

## Architectural Assessment

The codebase has been through two major naming transitions:
1. **roster -> knossos** (commit bbbc026): Completed for production code, but stale references remain in test fixtures and comments.
2. **skills -> commands/dromena/legomena** (ADR-0021 + ADR-0023): The unification is structurally complete. However, the original "skill" terminology persists pervasively in internal identifiers, types, and comments. This is the largest category below.

### Key Decision: "skill" in Path/Filesystem Context

The term "skill" in paths like `.claude/skills/`, `RiteSkillsDir()`, `UserSkillsDir()`, and the `usersync.ResourceSkills` constant refers to **Claude Code's filesystem convention**, not Knossos's domain model. These are correct and MUST NOT be renamed -- Claude Code owns that directory name. The rename targets below are limited to places where "skill" is used in the Knossos domain model (rite manifests, budget calculations, form descriptions, invoker component labels) where the correct term is now "legomena" or "mena" per ADR-0023.

**However**, the rite manifest YAML schema itself uses `skills:` as the field name, and this is a serialization format used by rite authors. Renaming YAML fields is a **breaking schema change** that would affect all 12 rites. This is explicitly out of scope for a terminology sweep and would require its own ADR.

### Verdict

After thorough analysis, the majority of "skill" usage in Go code falls into two categories:
1. **Filesystem path references** (`.claude/skills/`): Correct, must not change.
2. **Rite manifest schema** (`skills:` YAML field, `SkillRef` type, `SkillsCost` budget field): Breaking schema change, out of scope.

This leaves a smaller but meaningful set of renames concentrated in comments, test fixtures, a stale output type, and one stale comment referencing the old `/stamp skill` terminology.

---

## Target 1: Stale "roster" in Test Fixtures

**Category**: Stale terminology from pre-bbbc026 era
**Find**: `CEM/roster` in test fixture strings
**Replace**: `CEM/knossos`
**Risk**: TRIVIAL
**Batch**: 15

| File | Line(s) | Current | Proposed |
|------|---------|---------|----------|
| `internal/agent/frontmatter_test.go` | 45, 88, 89 | `"Designs CEM/roster schemas"` | `"Designs CEM/knossos schemas"` |
| `internal/agent/validate_test.go` | 52 | `"Designs CEM/roster schemas"` | `"Designs CEM/knossos schemas"` |
| `internal/agent/validate_test.go` | 306, 339 | `"Coordinates ecosystem phases for CEM/roster infrastructure work"` | `"Coordinates ecosystem phases for CEM/knossos infrastructure work"` |

**Invariants**: These are test fixture strings (fake agent descriptions). No production behavior affected.
**Verification**: `CGO_ENABLED=0 go test ./internal/agent/...` -- all existing tests pass with updated fixture strings.

---

## Target 2: Stale "roster" in usersync Test Data

**Category**: Stale terminology
**Find**: `[]byte("roster")` as file content in test
**Replace**: `[]byte("source")` (generic, matches test intent)
**Risk**: TRIVIAL
**Batch**: 15

| File | Line(s) | Current | Proposed |
|------|---------|---------|----------|
| `internal/usersync/usersync_test.go` | 416 | `[]byte("roster")` | `[]byte("source")` |

**Invariants**: Test writes arbitrary content to a temp file. The string "roster" is meaningless here -- it's just file content for a collision detection test. The test asserts the file exists, not what it contains.
**Verification**: `CGO_ENABLED=0 go test ./internal/usersync/...`

---

## Target 3: Stale "roster-owned" in Inscription Marker Test

**Category**: Stale terminology in legacy marker test
**Find**: `<!-- SYNC: roster-owned -->` in test fixture
**Risk**: SKIP

This is a test for `ParseLegacyMarkers()` -- the function's purpose is to detect OLD marker formats. The test fixture intentionally uses the legacy format. Renaming it would defeat the test's purpose.

---

## Target 4: `RosterMigrateOutput` Type Name

**Category**: Stale type name in output package
**Find**: `RosterMigrateOutput`
**Replace**: Not renamed -- SKIP
**Risk**: N/A

This type is part of the `ari migrate roster-to-knossos` command, which is an **active migration tool**. The name accurately describes what it does: it outputs the results of a roster-to-knossos migration. Renaming it would be confusing since the command itself references "roster". The entire `internal/cmd/migrate/roster_to_knossos.go` file is migration-specific code that will eventually be removed when the migration window closes.

---

## Target 5: "/stamp skill" Comment

**Category**: Stale terminology in comment (ADR-0023)
**Find**: `the /stamp skill` in comment
**Replace**: `the /stamp command`
**Risk**: TRIVIAL
**Batch**: 15

| File | Line(s) | Current | Proposed |
|------|---------|---------|----------|
| `internal/hook/clewcontract/record.go` | 122 | `// This is the primary integration point for the /stamp skill.` | `// This is the primary integration point for the /stamp command.` |

**Invariants**: Comment-only change. Zero behavior impact.
**Verification**: `CGO_ENABLED=0 go test ./internal/hook/clewcontract/...`

---

## Target 6: Rite Form Comments Referencing "skills"

**Category**: Comment terminology inconsistent with ADR-0023
**Find**: Comments describing rite forms using "skills"
**Replace**: Update to use "mena" or "legomena" where appropriate
**Risk**: TRIVIAL
**Batch**: 15

| File | Line(s) | Current | Proposed |
|------|---------|---------|----------|
| `internal/rite/manifest.go` | 20 | `// FormSimple represents a rite with skills only, no agents.` | `// FormSimple represents a rite with mena only, no agents.` |
| `internal/rite/manifest.go` | 22 | `// FormPractitioner represents a rite with agents + skills.` | `// FormPractitioner represents a rite with agents + mena.` |

**Invariants**: Comment-only changes. The `FormSimple` and `FormPractitioner` constant values and usage are unchanged.
**Verification**: `CGO_ENABLED=0 go test ./internal/rite/...`

---

## Target 7: `Skills` Field Deprecation Comment in Materialize

**Category**: Stale field that should have a stronger deprecation signal
**Find**: `Skills []string` field in materialize `RiteManifest`
**Risk**: SKIP (already marked deprecated)

The field at `internal/materialize/materialize.go:59` is already annotated `// Deprecated: use Legomena instead`. This is correct. No rename needed. The field exists for backward compatibility with existing rite YAML files.

---

## Target 8: Invoker Component String Literals

**Category**: Domain model uses "skills" as component identifier
**Find**: `"skills"` as component value in invoker and state
**Replace**: Not renamed -- SKIP
**Risk**: N/A

The strings `"skills"`, `"agents"` at `internal/rite/invoker.go:13,237,282` and `internal/rite/state.go:28` are component identifiers used in the rite invocation protocol. These map directly to filesystem directory names within a rite (`rites/{name}/skills/`). Renaming these would be a **breaking protocol change** requiring updates to all rite manifests, state files, and the invocation command interface. Out of scope.

---

## Target 9: Package Doc Comments Referencing "Ariadne"

**Category**: Naming audit -- "Ariadne" is the CLI binary name, appropriate in many contexts
**Find**: `for Ariadne` in package doc comments
**Risk**: SKIP

The binary is named `ari` (short for Ariadne). Package doc comments like `// Package errors provides domain-specific error types for Ariadne.` are technically correct -- Ariadne is the product name. These do NOT need renaming to "Knossos" because Knossos is the framework/platform, while Ariadne is the CLI tool these packages support. The naming is intentional and correct.

---

## Target 10: `ARIADNE_` Environment Variable Prefix

**Category**: Naming audit
**Find**: `ARIADNE_STALE_SESSION_DAYS`, `ARIADNE_BUDGET_DISABLE`, `ARIADNE_MSG_WARN`, `ARIADNE_MSG_PARK`, `ARIADNE_SESSION_KEY`
**Risk**: SKIP (breaking change)

These are **published environment variables** documented in CLI help text. Users may have them in shell profiles. Renaming would be a breaking change requiring a deprecation period and migration tooling. Out of scope for a hygiene sprint.

---

## Target 11: `ariadne-msg-count-` Temp File Prefix

**Category**: Internal implementation detail
**Find**: `"ariadne-msg-count-"` prefix for temp state files
**Risk**: SKIP (low value, nonzero risk)

This prefix is used in `internal/cmd/hook/budget.go:163` for per-session temp files. Renaming it would invalidate any in-flight session state. The files are ephemeral (in `os.TempDir()`) and users never see this string. Risk/reward does not justify the change.

---

## Target 12: TODO Comments Inventory

**Category**: TODO marker triage
**Risk**: Informational only -- no renames

| File | Line | Comment | Status |
|------|------|---------|--------|
| `internal/agent/regenerate.go` | 73-80 | `// Preserve existing content or add TODO marker` | LIVE CODE -- generates TODO markers in scaffolded agents. Correct behavior. |
| `internal/agent/archetype.go` | 19 | `// Scaffold generates TODO markers for these sections.` | LIVE CODE -- documents scaffold behavior. Correct. |
| `internal/agent/archetype.go` | 38 | `// TodoHint provides guidance for author-owned sections (used in TODO markers).` | LIVE CODE -- documents struct field. Correct. |
| `internal/cmd/agent/new.go` | 32, 118 | `Author sections are marked with TODO comments for you to fill in.` | LIVE CODE -- user-facing help text. Correct. |
| `internal/cmd/agent/update.go` | 40 | `Author sections are never modified (or added with TODO markers if missing).` | LIVE CODE -- user-facing help text. Correct. |
| `internal/cmd/session/status_test.go` | 93 | `// TODO: Capture and verify JSON output contains sails_color: WHITE` | DEFERRED -- test coverage gap. Track for Sprint 4. |

**Verdict**: Only one actual TODO marker found that represents incomplete work (`status_test.go:93`). All others are live code that generates or references TODO markers as a feature. The `status_test.go` TODO should be tracked but is not a rename target.

---

## Summary

### Actionable Renames (Batch 15 only)

| # | Target | Files | Lines Changed | Risk |
|---|--------|-------|---------------|------|
| 1 | `CEM/roster` -> `CEM/knossos` in test fixtures | 2 files | 6 lines | TRIVIAL |
| 2 | `"roster"` -> `"source"` in usersync test data | 1 file | 1 line | TRIVIAL |
| 5 | `/stamp skill` -> `/stamp command` in comment | 1 file | 1 line | TRIVIAL |
| 6 | "skills" -> "mena" in rite form comments | 1 file | 2 lines | TRIVIAL |
| **Total** | | **4 files** | **10 lines** | **TRIVIAL** |

### Skipped with Justification

| # | Target | Reason |
|---|--------|--------|
| 3 | `roster-owned` in marker test | Intentionally tests legacy format |
| 4 | `RosterMigrateOutput` type | Active migration command, name is accurate |
| 7 | `Skills` field deprecation | Already properly annotated |
| 8 | `"skills"` component strings | Breaking protocol change, maps to filesystem |
| 9 | `for Ariadne` package docs | Ariadne is correct product name |
| 10 | `ARIADNE_*` env vars | Published interface, breaking change |
| 11 | `ariadne-msg-count-` prefix | Low value, nonzero risk |

### Deferred to Sprint 4

| Item | Reason |
|------|--------|
| `status_test.go:93` TODO | Test coverage gap for sails_color in JSON output |
| Rite manifest `skills:` YAML field rename | Breaking schema change, needs ADR |

---

## Batch 15: Comment and Test Fixture Cleanup

**Risk**: TRIVIAL
**Dependencies**: None
**Rollback**: Single commit revert

### RF-015-A: Update "CEM/roster" to "CEM/knossos" in Agent Test Fixtures

**Before State:**
- `internal/agent/frontmatter_test.go:45`: `role: "Designs CEM/roster schemas"`
- `internal/agent/frontmatter_test.go:88-89`: assertion string `"Designs CEM/roster schemas"`
- `internal/agent/validate_test.go:52`: `role: "Designs CEM/roster schemas"`
- `internal/agent/validate_test.go:306`: `description: "Coordinates ecosystem phases for CEM/roster infrastructure work"`
- `internal/agent/validate_test.go:339`: `description: "Coordinates ecosystem phases for CEM/roster infrastructure work"`

**After State:**
- All 5 occurrences changed from `CEM/roster` to `CEM/knossos`

**Invariants:**
- Test fixture strings only -- no production behavior
- All existing assertions pass (the assertion strings are updated to match)

**Verification:**
1. `CGO_ENABLED=0 go test ./internal/agent/...`
2. Confirm all tests pass

### RF-015-B: Update Stale "roster" Content in UserSync Test

**Before State:**
- `internal/usersync/usersync_test.go:416`: `[]byte("roster")`

**After State:**
- `internal/usersync/usersync_test.go:416`: `[]byte("source")`

**Invariants:**
- File content is arbitrary -- test verifies file existence and collision behavior, not content string

**Verification:**
1. `CGO_ENABLED=0 go test ./internal/usersync/...`
2. Confirm all tests pass

### RF-015-C: Update "/stamp skill" to "/stamp command" in Comment

**Before State:**
- `internal/hook/clewcontract/record.go:122`: `// This is the primary integration point for the /stamp skill.`

**After State:**
- `internal/hook/clewcontract/record.go:122`: `// This is the primary integration point for the /stamp command.`

**Invariants:**
- Comment-only change

**Verification:**
1. `CGO_ENABLED=0 go test ./internal/hook/clewcontract/...`
2. Confirm all tests pass

### RF-015-D: Update Rite Form Comments from "skills" to "mena"

**Before State:**
- `internal/rite/manifest.go:20`: `// FormSimple represents a rite with skills only, no agents.`
- `internal/rite/manifest.go:22`: `// FormPractitioner represents a rite with agents + skills.`

**After State:**
- `internal/rite/manifest.go:20`: `// FormSimple represents a rite with mena only, no agents.`
- `internal/rite/manifest.go:22`: `// FormPractitioner represents a rite with agents + mena.`

**Invariants:**
- Comment-only changes
- Constant values `FormSimple` and `FormPractitioner` unchanged

**Verification:**
1. `CGO_ENABLED=0 go test ./internal/rite/...`
2. Confirm all tests pass

---

## Recommendation

This sprint is intentionally small. The 10 lines of changes across 4 files reflect the reality that the major naming transitions (roster->knossos, skills->commands) were executed well during their respective initiatives. The remaining inconsistencies are either:

1. **Trivial stragglers** (test fixtures, comments) -- addressed in Batch 15
2. **Correct usage** (filesystem paths, product names) -- SKIP
3. **Breaking changes** (env vars, YAML schema, protocol strings) -- out of scope

**Batch 15 is the only batch needed.** Batches 16-18 are empty because there are no string literal renames, variable renames, or type renames that are both safe and worthwhile.

The Janitor should execute Batch 15 as a single commit.

---

## Handoff Checklist

- [x] Every smell classified (addressed, deferred with reason, or dismissed)
- [x] Each refactoring has before/after contract documented
- [x] Invariants and verification criteria specified
- [x] Refactorings sequenced with explicit dependencies (single batch, no dependencies)
- [x] Rollback points identified (single commit revert)
- [x] Risk assessment complete (TRIVIAL across all targets)

## Artifact Attestation

| Artifact | Path | Verified Via |
|----------|------|-------------|
| This document | `/Users/tomtenuta/Code/knossos/docs/hygiene/RENAME-MAP-sprint-3.md` | Written by architect-enforcer |
| Sprint 1 contract | `/Users/tomtenuta/Code/knossos/docs/hygiene/DELETION-CONTRACT-sprint-1.md` | Read tool, lines 1-40 |
| Sprint 2 contract | `/Users/tomtenuta/Code/knossos/docs/hygiene/CONSOLIDATION-CONTRACT-sprint-2.md` | Read tool, lines 1-30 |
| frontmatter_test.go | `internal/agent/frontmatter_test.go` | Read tool, lines 40-94 |
| validate_test.go | `internal/agent/validate_test.go` | Read tool, lines 45-64, 300-344 |
| usersync_test.go | `internal/usersync/usersync_test.go` | Read tool, lines 410-424 |
| record.go | `internal/hook/clewcontract/record.go` | Read tool, full file |
| manifest.go | `internal/rite/manifest.go` | Read tool, lines 18-25 |
| output.go | `internal/output/output.go` | Read tool, lines 555-614 |
| paths.go | `internal/paths/paths.go` | Read tool, lines 185-319 |
