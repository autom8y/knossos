# Gap Analysis: Ariadne PRD Review

> Analysis of PRD-ariadne.md against SPIKE-ariadne-go-cli-architecture.md

**Date**: 2026-01-04
**Analyst**: ecosystem-analyst
**Status**: READY FOR TDD (with minor revisions noted)

---

## Executive Summary

The PRD (`docs/requirements/PRD-ariadne.md`) is **substantially complete** and ready for TDD development. All major spike findings have been incorporated correctly. This analysis identifies **4 minor gaps** and **2 clarifications needed**, none of which block TDD progression.

**Verdict**: READY FOR TDD

**Confidence**: HIGH (all critical areas covered, gaps are clarifications not blockers)

---

## 1. PRD Completeness Assessment

### 1.1 Spike Library Recommendations vs PRD Dependencies

| Library | Spike Recommendation | PRD Section 3.2 | Status |
|---------|---------------------|-----------------|--------|
| spf13/cobra v1.8+ | CLI Framework | Listed | MATCH |
| spf13/viper v1.18+ | Config | Listed | MATCH |
| santhosh-tekuri/jsonschema/v6 | Schema validation | Listed | MATCH |
| adrg/xdg v0.5+ | XDG paths | Listed | MATCH |
| evanphx/json-patch/v5 | JSON merge | Listed | MATCH |
| gopkg.in/yaml.v3 | YAML parsing | Listed | MATCH |
| yuin/goldmark v1.6+ | Markdown parsing | Listed | MATCH |

**Result**: 100% alignment. All spike library recommendations are in PRD.

### 1.2 Architecture Pattern Alignment

| Spike Pattern | PRD Coverage | Location |
|--------------|--------------|----------|
| gh CLI factory pattern | Implicit in structure | Section 3.1 |
| Domain directories | Explicit structure | Section 3.1 |
| I/O abstraction for testing | Not explicit | GAP-1 |
| Piped output detection | Not explicit | GAP-2 |
| Embedded schema pattern | Covered | Section 6.1 |
| XDG directory layout | Covered | Section 8.2 |
| Strangler fig migration | Covered | Section 9.3 |

### 1.3 Spike Risks vs PRD Mitigations

| Spike Risk | PRD Mitigation | Status |
|------------|----------------|--------|
| Markdown merge conflicts | Section 7.2: Flag conflicts for manual resolution | COVERED |
| Schema validation performance | Section 6.1: Embedded schemas, cached compilation | COVERED |
| Cross-platform XDG differences | Section 3.2: adrg/xdg handles automatically | COVERED |
| Behavioral parity with bash | Section 11.2: Spec-based testing | COVERED |
| k8s strategic merge complexity | Section 7.1: Using json-patch/v5 instead | COVERED |

---

## 2. Interface Contract Gaps

### 2.1 Command Coverage Analysis

**PRD Section 4.1 specifies 26 commands:**

| Domain | Command | Bash Equivalent | Interface Complete |
|--------|---------|-----------------|-------------------|
| session | create | session-manager.sh create | YES |
| session | status | session-manager.sh status | YES |
| session | list | (new capability) | YES |
| session | park | session-manager.sh mutate park | YES |
| session | resume | session-manager.sh mutate resume | YES |
| session | wrap | session-manager.sh mutate wrap | YES |
| session | transition | session-manager.sh transition | YES |
| session | migrate | session-migrate.sh | YES |
| session | audit | (new capability) | PARTIAL - GAP-3 |
| session | lock | (implicit in bash) | YES |
| session | unlock | (implicit in bash) | YES |
| team | switch | swap-team.sh | YES |
| team | list | swap-team.sh --list | YES |
| team | status | swap-team.sh --status | YES |
| team | validate | swap-team.sh --verify | YES |
| manifest | show | (new capability) | YES |
| manifest | diff | (new capability) | YES |
| manifest | validate | (new capability) | YES |
| manifest | merge | (new capability) | YES |
| sync | init | roster-sync init | YES |
| sync | pull | roster-sync pull | YES |
| sync | push | roster-sync push | YES |
| sync | status | roster-sync status | YES |
| sync | diff | roster-sync diff | YES |
| sync | validate | roster-sync validate | YES |
| sync | repair | roster-sync repair | YES |

**All 26 commands have interfaces specified.**

### 2.2 GAP-3: Audit Command Incomplete

**Location**: PRD Section 4.1

**Current**:
```
ari session audit [--session-id=ID] [--limit=N]
```

**Missing**:
- Output format specification (what fields?)
- Filter options (by date range, by operation type?)
- Relationship to `session-mutations.log` format

**Recommendation**: Add audit output specification in TDD. Not blocking.

### 2.3 Flag Coverage Analysis

**Global flags (PRD 4.2)**: Complete
- `--output` / `-o`
- `--verbose` / `-v`
- `--config`
- `--project-dir` / `-p`
- `--session-id` / `-s`

**GAP-4: Missing --dry-run Global Flag**

**Issue**: `--dry-run` is specified for `sync pull`, `sync push`, `sync repair` but not as a global flag. The bash scripts (swap-team.sh, roster-sync) support `--dry-run` at multiple levels.

**Recommendation**: Add `--dry-run` as global flag or explicitly document which commands support it.

### 2.4 Error Code Coverage

**PRD Section 5.1 error codes vs spike risks:**

| Error Code | Exit | Spike Risk Covered |
|------------|------|-------------------|
| SUCCESS | 0 | - |
| GENERAL_ERROR | 1 | - |
| USAGE_ERROR | 2 | - |
| LOCK_TIMEOUT | 3 | Behavioral parity (flock timeout) |
| LOCK_STALE | 3 | Behavioral parity (stale detection) |
| SCHEMA_INVALID | 4 | Schema validation performance |
| LIFECYCLE_VIOLATION | 5 | Session FSM |
| FILE_NOT_FOUND | 6 | - |
| PERMISSION_DENIED | 7 | - |
| MERGE_CONFLICT | 8 | Markdown merge conflicts |
| PROJECT_NOT_FOUND | 9 | - |

**GAP-5: Missing Error Codes for Team/Sync Operations**

| Missing Code | Context |
|--------------|---------|
| TEAM_NOT_FOUND | `ari team switch <invalid>` |
| SYNC_CONFLICT | Three-way sync classification conflict |
| MANIFEST_MISMATCH | Manifest checksum validation failure |
| ORPHAN_CONFLICT | swap-team.sh EXIT_ORPHAN_CONFLICT equivalent |
| RECOVERY_REQUIRED | swap-team.sh EXIT_RECOVERY_REQUIRED equivalent |

**Recommendation**: Add domain-specific error codes in TDD. Exit codes can share (e.g., TEAM_NOT_FOUND = 6 like FILE_NOT_FOUND).

---

## 3. Integration Gaps

### 3.1 State-Mate Integration

**PRD Section 9.1**: State-mate invokes ari for state mutations.

**Analysis**:
- Current state-mate uses shell commands for locking, schema validation
- PRD correctly identifies capability discovery as "hardcoded in state-mate.md"
- Migration requires state-mate.md update to use `ari` commands

**Gap Assessment**: NO GAP. Integration path is clear.

**Migration Sequence**:
1. ari session commands implemented
2. Update state-mate.md to invoke `ari session park/resume/wrap`
3. State-mate continues using Read/Write/Edit for context file mutations (ari does infrastructure)

### 3.2 Hook Integration

**PRD Section 9.2**: Hooks remain bash scripts that call ari.

**Current hooks using session-manager.sh**:
- `.claude/hooks/lib/session-manager.sh` (sourced by hooks)
- Auto-park on SessionStop
- Session status injection on SessionStart

**Gap Assessment**: NO GAP. Hooks call `ari` binary, same interface as calling bash scripts.

### 3.3 Migration Bridge

**PRD Section 9.3**: Bash scripts call ari during migration.

**Current bash scripts to migrate**:
1. `session-manager.sh` -> `ari session *`
2. `swap-team.sh` -> `ari team *`
3. `roster-sync` (lib/sync/*) -> `ari sync *`

**GAP-6: Migration Order Conflict**

**Spike Recommendation (Section 6)**: validate -> manifest -> sync -> team -> session

**PRD Recommendation (Section 13.1)**: session -> team -> manifest -> sync

**Rationale Analysis**:
- Spike: Risk-ordered (stateless first, stateful last)
- PRD: Value-ordered (session = "actual thread", enables dogfooding)

**Resolution**: PRD ordering is acceptable. The "front-load risk" rationale for session-first is valid for this project where session management is critical path. TDD should note this divergence and ensure session tests are comprehensive.

---

## 4. Cross-Domain Dependencies

### 4.1 Dependency Matrix

```
session ────────────────┐
   │                    │
   │ depends on         │
   v                    v
 lock ◄─────────────── paths
   │                    │
   │                    │
   v                    v
validation ◄────────── schemas (embedded)
   │
   │
   v
  team ────────────────┐
   │                    │
   │ depends on         │
   v                    v
manifest ◄─────────── validation
   │
   │
   v
 sync ◄────────────── merge
                        │
                        v
                      manifest
```

### 4.2 Internal Package Dependencies

| Package | Depends On | Used By |
|---------|-----------|---------|
| `internal/paths` | `adrg/xdg` | All domains |
| `internal/lock` | `internal/paths` | session |
| `internal/validation` | `santhosh-tekuri/jsonschema`, `schemas/` | All domains |
| `internal/merge` | `evanphx/json-patch`, `yuin/goldmark` | manifest, sync |
| `internal/output` | - | All domains |
| `internal/cmd/session` | lock, paths, validation, output | - |
| `internal/cmd/team` | paths, validation, output | - |
| `internal/cmd/manifest` | validation, merge, output | - |
| `internal/cmd/sync` | validation, merge, manifest ops, output | - |

### 4.3 Build Order (No Circular Dependencies)

1. `internal/paths` (foundational)
2. `internal/output` (foundational)
3. `internal/lock` (depends on paths)
4. `internal/validation` (depends on schemas/)
5. `internal/merge` (standalone)
6. `internal/cmd/session` (depends on 1-4)
7. `internal/cmd/team` (depends on 1, 2, 4)
8. `internal/cmd/manifest` (depends on 1, 2, 4, 5)
9. `internal/cmd/sync` (depends on all)

---

## 5. Testing Strategy Assessment

### 5.1 Spike Concerns vs PRD Coverage

| Spike Concern | PRD Section 11 | Status |
|--------------|----------------|--------|
| Golden file testing for parity | 11.2: Spec-based testing | DIFFERENT APPROACH (acceptable) |
| Concurrency testing | 11.4: `go test -race` | COVERED |
| Satellite matrix testing | 11.3: CI matrix fixtures | COVERED |
| Schema validation testing | 11.1: Unit tests | IMPLICIT |

### 5.2 PRD Testing Strategy Adequacy

**Strengths**:
- Clear test type taxonomy (unit, integration, parity, concurrency)
- Spec-based approach avoids bash bug replication
- Satellite matrix ensures real-world coverage

**Potential Gap**: No mention of testing state-mate integration.

**Recommendation**: Add integration test category for "state-mate invokes ari" scenario in TDD.

---

## 6. Gaps Summary

### 6.1 Blocking Gaps

**NONE** - PRD is ready for TDD.

### 6.2 Non-Blocking Gaps (Address in TDD)

| ID | Gap | Severity | Resolution |
|----|-----|----------|------------|
| GAP-1 | I/O abstraction for testing not explicit | LOW | Document in TDD architecture |
| GAP-2 | Piped output detection not explicit | LOW | Implement in output package |
| GAP-3 | Audit command output format incomplete | LOW | Specify in TDD |
| GAP-4 | --dry-run flag scope unclear | MEDIUM | Clarify in TDD (global vs command-specific) |
| GAP-5 | Missing domain-specific error codes | MEDIUM | Add in TDD error handling section |
| GAP-6 | Migration order diverges from spike | LOW | Document rationale, proceed with PRD order |

### 6.3 Clarifications Needed

1. **Audit output format**: What fields does `ari session audit` output?
2. **Dry-run scope**: Global flag or per-command?

---

## 7. Recommendation

### 7.1 Verdict

**READY FOR TDD**

The PRD is comprehensive and correctly incorporates spike findings. All 26 commands have interfaces specified. Error handling covers spike-identified risks. Testing strategy addresses spike concerns.

### 7.2 TDD Prerequisites

Before starting TDD:
1. Clarify audit command output format (can be documented in TDD)
2. Decide dry-run flag scope (can be decided in TDD)

### 7.3 Complexity Classification

**Recommended**: MODULE (per existing classification)

**Rationale**:
- Four well-defined domains
- Clear internal package boundaries
- No external service dependencies
- Defined integration points with existing bash infrastructure

### 7.4 Test Satellite Recommendations

| Satellite Type | Purpose |
|---------------|---------|
| test-satellite-baseline | Minimal .claude/ structure |
| test-satellite-minimal | No custom settings |
| test-satellite-complex | Nested arrays, custom hooks, multiple teams |
| test-satellite-virgin | No .claude/ directory (tests PROJECT_NOT_FOUND) |
| test-satellite-migrating | Has both bash scripts and ari (migration testing) |

---

## Appendix A: File Inventory

### A.1 Documents Analyzed

| Document | Path | Purpose |
|----------|------|---------|
| PRD | docs/requirements/PRD-ariadne.md | Requirements specification |
| Spike | docs/spikes/SPIKE-ariadne-go-cli-architecture.md | Research findings |
| state-mate | user-agents/state-mate.md | Integration target |
| session-manager.sh | .claude/hooks/lib/session-manager.sh | Replacement target |
| swap-team.sh | swap-team.sh | Replacement target |
| sync-core.sh | lib/sync/sync-core.sh | Replacement target |
| session-context.schema.json | schemas/artifacts/session-context.schema.json | Validation schema |

### A.2 Command Count Verification

- **Session**: 11 commands (create, status, list, park, resume, wrap, transition, migrate, audit, lock, unlock)
- **Team**: 4 commands (switch, list, status, validate)
- **Manifest**: 4 commands (show, diff, validate, merge)
- **Sync**: 7 commands (init, pull, push, status, diff, validate, repair)
- **Total**: 26 commands + version = 27

---

## Appendix B: Handoff Checklist

- [x] Root cause traced to specific component (PRD vs spike alignment)
- [x] Success criteria defined (all spike findings incorporated)
- [x] Affected systems enumerated (session-manager.sh, swap-team.sh, roster-sync, state-mate)
- [x] Complexity level recommended (MODULE)
- [x] Test satellite matrix specified
- [x] Gap analysis committed to session artifacts
- [x] Artifacts verified via Read tool after writing

---

*Generated by ecosystem-analyst*
*Handoff to: Context Architect (for TDD development)*
