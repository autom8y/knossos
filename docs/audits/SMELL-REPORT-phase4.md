# Smell Report: Phase 4 Assessment

**Codebase**: roster (knossos platform)
**Analyzed**: 2026-01-07
**Scope**: Phase 4 HYGIENE assessment for knossos-finalization initiative
**Upstream**: Phase 3 ECOSYSTEM implementation complete
**Phase 1 Reference**: docs/spikes/SPIKE-knossos-consolidation-architecture.md

---

## Executive Summary

- **Phase 1 findings status**: 4 RESOLVED, 2 PARTIALLY_RESOLVED, 2 OUTSTANDING
- **New findings**: 3 identified
- **Total cleanup items**: 7 actionable items
- **Estimated effort**: 2-4 days

Phase 3 (ECOSYSTEM) successfully implemented the core architectural changes:
- Go module migrated from `ariadne/` to root
- `ari sync materialize` command implemented
- `.claude/` removed from git tracking (gitignored)
- 12 rite manifests created

Remaining work is cleanup of transitional artifacts and legacy shell scripts.

---

## Phase 1 Findings Status

### SM-001: Triple Hook Duplication - RESOLVED

**Original Finding**: Hooks duplicated in `.claude/hooks/`, `ariadne/.claude/hooks/`, and `user-hooks/`.

**Current Status**: RESOLVED

**Evidence**:
- `ariadne/` directory deleted (visible in git status: 200+ deleted files)
- `.claude/` now gitignored and generated via `ari sync materialize`
- Single source of truth established in `templates/` directory

**Verification**:
```
$ ls -la ariadne/
ariadne/ directory does not exist or is empty
```

---

### SM-002: Go Module Path Mismatch - RESOLVED

**Original Finding**: Module path `github.com/autom8y/knossos` in `ariadne/` subdirectory.

**Current Status**: RESOLVED

**Evidence**:
- Go module now at root: `/Users/tomtenuta/Code/roster/go.mod`
- Module path: `module github.com/autom8y/knossos` (kept as decided)
- Standard Go layout: `cmd/ari/` + `internal/`

**Verification**:
```
$ cat go.mod | head -1
module github.com/autom8y/knossos
```

---

### SM-003: Shell Library Divergence - RESOLVED

**Original Finding**: Shell libraries diverged between `.claude/hooks/lib/`, `ariadne/.claude/hooks/lib/`, and `user-hooks/lib/`.

**Current Status**: RESOLVED

**Evidence**:
- `ariadne/.claude/hooks/lib/` deleted with parent directory
- `.claude/` now generated from templates
- Single canonical source remains

---

### SM-004: Monolithic swap-rite.sh - PARTIALLY_RESOLVED

**Original Finding**: 3,773-line monolithic script with 68 functions.

**Current Status**: PARTIALLY_RESOLVED

**Evidence**:
- `lib/rite/` directory created with extracted modules:
  - `rite-resource.sh` (347 LOC) - resource operations
  - `rite-transaction.sh` (548 LOC) - transaction safety
  - `rite-hooks-registration.sh` (537 LOC) - hook registration
- `swap-rite.sh` remains at **3,773 lines** (unchanged)

**Gap Analysis**:
- Modules extracted but not yet integrated into main script
- Main script still contains all original functions
- `ari sync materialize` Go command exists but doesn't replace swap-rite.sh

**Remaining Work**:
| Task | Complexity | Priority |
|------|------------|----------|
| Integrate lib/rite modules into swap-rite.sh | MEDIUM | P2 |
| OR: Replace swap-rite.sh with Go implementation | HIGH | P3 |

---

### SM-007: Orphaned Backup Directories - OUTSTANDING

**Original Finding**: ~1MB of orphaned backup directories.

**Current Status**: OUTSTANDING

**Evidence**:
```
.claude/agents.backup           52KB
.claude/commands.backup         0KB
.claude/skills.backup           0KB
.claude/commands.orphan-backup  12KB
.claude/skills.orphan-backup    1.0MB
```

**Location**: `/Users/tomtenuta/Code/roster/.claude/*.backup*`

**Note**: These are in `.claude/` which is now gitignored, so they are runtime artifacts not in repo. However, the `.claude.backup.pre-materialize/` directory (6.8MB) IS untracked and should be addressed.

**Remaining Work**:
| Task | Complexity | Priority |
|------|------------|----------|
| Remove .claude.backup.pre-materialize/ directory | LOW | P1 |
| Document backup cleanup in swap-rite.sh | LOW | P3 |

---

### SM-008: Implementation Docs at Repo Root - OUTSTANDING

**Original Finding**: ~48KB of implementation documentation at repo root.

**Current Status**: OUTSTANDING

**Evidence**:
```
CONTEXT_SEED.md                    5KB
DEFECT-D002-RESOLUTION.md          6KB
IMPLEMENTATION-SUMMARY-sails-status.md  6KB
IMPLEMENTATION_SUMMARY.md          8KB
IMPLEMENTATION_VERIFICATION.md    10KB
PHASE-5-HANDOFF.md                10KB
RITE_SKILL_MATRIX.md              12KB
SAILS-STATUS-REFERENCE.md          3KB
```

**Total**: ~60KB of transitional documentation

**Recommendation**: Move to `docs/archive/` or delete if no longer needed.

**Remaining Work**:
| Task | Complexity | Priority |
|------|------------|----------|
| Archive or remove root implementation docs | LOW | P1 |

---

### SM-012: roster-sync Shell vs Go Duplication - PARTIALLY_RESOLVED

**Original Finding**: Parallel implementations in shell (1,413 lines) and Go.

**Current Status**: PARTIALLY_RESOLVED

**Evidence**:
- Shell `roster-sync` still exists: **1,413 lines**
- Go `ari sync materialize` implemented (see `/Users/tomtenuta/Code/roster/internal/cmd/sync/materialize.go`)
- Go materialize uses `internal/materialize/` package (372 lines)

**Gap Analysis**:
- `ari sync materialize` handles core materialization
- `roster-sync` still used for satellite project sync (not yet migrated)
- Both coexist in current state

**Remaining Work**:
| Task | Complexity | Priority |
|------|------------|----------|
| Migrate roster-sync to Go OR deprecate | MEDIUM | P2 |
| Document which sync tool to use where | LOW | P1 |

---

## New Findings (Phase 4)

### SM-P4-001: sync-user-*.sh Script Duplication (MEDIUM)

**Category**: DRY Violation
**Locations**:
- `/Users/tomtenuta/Code/roster/sync-user-agents.sh` (734 lines)
- `/Users/tomtenuta/Code/roster/sync-user-commands.sh` (959 lines)
- `/Users/tomtenuta/Code/roster/sync-user-hooks.sh` (1,096 lines)
- `/Users/tomtenuta/Code/roster/sync-user-skills.sh` (997 lines)

**Evidence**: Four parallel scripts with similar structure for syncing different resource types to user directories. Total: **3,786 lines**.

**Blast Radius**: 4 files, ~3,800 lines
**Fix Complexity**: MEDIUM (extract common patterns, or replace with Go)
**ROI Score**: 6/10

**Note**: These may be deprecated by `ari sync materialize`. Need to verify if still in use.

---

### SM-P4-002: .claude.backup.pre-materialize Directory (LOW)

**Category**: Orphaned Artifact
**Location**: `/Users/tomtenuta/Code/roster/.claude.backup.pre-materialize/`

**Evidence**: 6.8MB backup directory created during Phase 3 migration, now orphaned.

**Blast Radius**: 1 directory
**Fix Complexity**: LOW (safe to delete)
**ROI Score**: 9/10

**Verification**: Untracked in git (visible in `git status`).

---

### SM-P4-003: rites/ Implementation Artifacts (LOW)

**Category**: Documentation Housekeeping
**Locations**:
- `/Users/tomtenuta/Code/roster/rites/AUDIT_SUMMARY.md` (8KB)
- `/Users/tomtenuta/Code/roster/rites/IMPLEMENTATION_BLUEPRINT.md` (14KB)

**Evidence**: Implementation planning documents in rites/ directory, should be in docs/.

**Blast Radius**: 2 files
**Fix Complexity**: LOW (move or delete)
**ROI Score**: 7/10

---

## Prioritized Cleanup List

### Priority 1: Safe to Remove Now

| Item | Location | Size | Action |
|------|----------|------|--------|
| .claude.backup.pre-materialize/ | repo root | 6.8MB | Delete |
| Root implementation docs | /*.md | ~60KB | Move to docs/archive/ |
| rites/ implementation artifacts | rites/*.md | ~22KB | Move to docs/archive/ |

### Priority 2: Requires Verification

| Item | Location | Action |
|------|----------|--------|
| sync-user-*.sh scripts | repo root | Verify if deprecated by `ari sync materialize`, then remove or consolidate |
| roster-sync | repo root | Verify scope overlap with `ari sync`, document or deprecate |

### Priority 3: Deferred (Architectural)

| Item | Location | Action |
|------|----------|--------|
| swap-rite.sh integration | swap-rite.sh + lib/rite/ | Integrate extracted modules OR migrate to Go |

---

## Safe-to-Remove Items List

The following items can be safely removed without breaking functionality:

1. **`.claude.backup.pre-materialize/`**
   - Status: Orphaned backup from Phase 3 migration
   - Size: 6.8MB
   - Risk: None (backup of gitignored directory)

2. **Root-level implementation documentation**:
   - `CONTEXT_SEED.md`
   - `DEFECT-D002-RESOLUTION.md`
   - `IMPLEMENTATION-SUMMARY-sails-status.md`
   - `IMPLEMENTATION_SUMMARY.md`
   - `IMPLEMENTATION_VERIFICATION.md`
   - `PHASE-5-HANDOFF.md`
   - `SAILS-STATUS-REFERENCE.md`

   **Note**: `README.md` and `RITE_SKILL_MATRIX.md` may be intentionally at root; verify before action.

3. **rites/ implementation artifacts**:
   - `rites/AUDIT_SUMMARY.md`
   - `rites/IMPLEMENTATION_BLUEPRINT.md`

---

## Verification Attestation

| File/Directory | Verified Via | Attestation |
|----------------|--------------|-------------|
| ariadne/ | `ls -la` | Directory does not exist (deleted in Phase 3) |
| go.mod | `Read` tool | Module at root: `github.com/autom8y/knossos` |
| swap-rite.sh | `wc -l` | 3,773 lines confirmed |
| roster-sync | `wc -l` | 1,413 lines confirmed |
| lib/rite/ | `ls -la` | 3 modules extracted (resource, transaction, hooks-registration) |
| .claude.backup.pre-materialize/ | `du -sh` | 6.8MB orphaned backup confirmed |
| sync-user-*.sh | `wc -l` | 3,786 lines total (4 scripts) |
| Root *.md files | `ls -la` | 9 files, ~60KB total |
| rites/*.md | `ls -la` | 2 files (AUDIT_SUMMARY.md, IMPLEMENTATION_BLUEPRINT.md) |
| internal/materialize/ | `Read` tool | Go materialize package confirmed (372 lines) |

---

## Session Context

- **Initiative**: knossos-finalization
- **Phase**: 4 (HYGIENE)
- **Agent**: code-smeller
- **Upstream**: Phase 3 ECOSYSTEM complete
- **Downstream**: Architect Enforcer receives this report for cleanup planning

---

## Appendix: Git Status Summary

Staged deletions (Phase 3 cleanup in progress):
- `ariadne/` directory (~200 files)
- `.claude/hooks/ari/*.sh` (7 files)
- `.claude/hooks/lib/fail-open.sh`
- `.claude/hooks/session-guards/session-write-guard.sh`

Untracked additions:
- `.claude.backup.pre-materialize/` (should be removed)
- `cmd/`, `internal/`, `go.mod`, `go.sum` (new Go structure)
- `rites/*/manifest.yaml` (12 rite manifests)
- Various docs and spikes

Modified:
- `.github/workflows/ariadne-tests.yml`
- Several ADRs and design docs (terminology updates)
