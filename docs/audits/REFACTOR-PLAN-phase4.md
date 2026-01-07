# Refactor Plan: Phase 4 Cleanup

**Initiative**: knossos-finalization
**Phase**: 4 (HYGIENE)
**Agent**: architect-enforcer
**Upstream**: Code Smeller (SMELL-REPORT-phase4.md)
**Downstream**: Janitor (execution)
**Created**: 2026-01-07

---

## Executive Summary

This plan addresses 7 cleanup items identified in the Phase 4 smell report. All items are classified as **local cleanup** (no architectural boundary changes). The plan is structured in 3 phases with explicit rollback points between each.

**Scope**: Deletion of orphaned artifacts only. No behavior changes. No code modifications.

**Risk Level**: LOW - All items are orphaned files/directories with no runtime dependencies.

---

## Architectural Assessment

### Boundary Health

The smell report confirmed the major architectural improvements from Phase 3:
- Go module successfully migrated from `ariadne/` to root
- `.claude/` now generated via `ari sync materialize`
- Single source of truth established

### Root Cause Analysis

All remaining items fall into two categories:

1. **Transitional Artifacts** (60%): Documentation and backups created during Phase 3 migration that were never cleaned up
2. **Scope Clarification Needed** (40%): Scripts whose relationship to `ari sync materialize` needs documentation

### Classification

| Item | Category | Blast Radius | Action |
|------|----------|--------------|--------|
| .claude.backup.pre-materialize/ | Orphaned Backup | 1 directory | DELETE |
| Root IMPLEMENTATION*.md | Transitional Doc | 7 files | DELETE |
| rites/*.md artifacts | Misplaced Doc | 2 files | DELETE |
| sync-user-*.sh scripts | Scope Clarification | 4 files | DEFER (document) |
| roster-sync | Scope Clarification | 1 file | DEFER (document) |

---

## Pre-Cleanup Verification Checklist

Before executing ANY changes, the Janitor MUST verify:

```bash
# 1. Tests pass (baseline)
just test
# Expected: All packages OK

# 2. Build succeeds
just build
# Expected: ari binary created

# 3. swap-rite.sh works
./swap-rite.sh --list
# Expected: Lists available rites

# 4. ari sync materialize works
./ari sync materialize --rite hygiene
# Expected: .claude/ directory generated

# 5. Record current state
git status > /tmp/pre-cleanup-status.txt
ls -la *.md > /tmp/pre-cleanup-root-md.txt
```

**Gate**: Do NOT proceed if any verification fails.

---

## Phase 1: Orphaned Backup Removal (Lowest Risk)

### RF-001: Remove .claude.backup.pre-materialize Directory

**Risk**: NONE - Orphaned backup created during Phase 3 migration

**Before State**:
```
.claude.backup.pre-materialize/      6.8MB orphaned backup
  .archive/
  .cem/
  .shared-skills
  ACTIVE_RITE
  ACTIVE_WORKFLOW.yaml
  AGENT_MANIFEST.json
  agents/
  agents.backup/
  audit/
  backups/
  ... (runtime artifacts)
```

**After State**:
```
.claude.backup.pre-materialize/      DOES NOT EXIST
```

**Invariants**:
- No code references this directory
- Not in .gitignore (untracked, never committed)
- All tests continue to pass
- `ari sync materialize` still works

**Command**:
```bash
rm -rf .claude.backup.pre-materialize/
```

**Verification**:
```bash
# 1. Directory gone
[ ! -d ".claude.backup.pre-materialize" ] && echo "PASS: Directory removed"

# 2. Tests still pass
just test

# 3. Build still works
just build
```

**Rollback**: Not needed (untracked directory). If somehow required:
```bash
# Recreate by materializing an old state - but this is never needed
# The directory was a one-time backup
```

**Commit Message**:
```
chore(cleanup): remove orphaned .claude.backup.pre-materialize directory

Phase 4 cleanup item RF-001. This 6.8MB backup directory was created
during Phase 3 migration and is no longer needed. The canonical source
is now templates/ with generation via `ari sync materialize`.
```

---

## Phase 2: Root-Level Implementation Documentation (Low Risk)

### RF-002: Remove Transitional Implementation Documentation

**Risk**: LOW - Historical documentation from prior phases, no runtime impact

**Before State**:
```
/CONTEXT_SEED.md                      5.0KB
/DEFECT-D002-RESOLUTION.md            6.1KB
/IMPLEMENTATION-SUMMARY-sails-status.md  6.2KB
/IMPLEMENTATION_SUMMARY.md            8.2KB
/IMPLEMENTATION_VERIFICATION.md       10KB
/PHASE-5-HANDOFF.md                   9.9KB
/SAILS-STATUS-REFERENCE.md            2.9KB

Total: ~48KB in 7 files
```

**After State**:
```
(All 7 files removed)
```

**Files Explicitly PRESERVED**:
- `/README.md` - Project README (intentional)
- `/RITE_SKILL_MATRIX.md` - Active reference matrix (intentional)

**Invariants**:
- No code imports or references these files
- No scripts depend on these files
- All tests continue to pass
- README.md and RITE_SKILL_MATRIX.md remain

**Commands** (execute in order, one commit per file or grouped):
```bash
# Remove transitional docs
rm CONTEXT_SEED.md
rm DEFECT-D002-RESOLUTION.md
rm IMPLEMENTATION-SUMMARY-sails-status.md
rm IMPLEMENTATION_SUMMARY.md
rm IMPLEMENTATION_VERIFICATION.md
rm PHASE-5-HANDOFF.md
rm SAILS-STATUS-REFERENCE.md
```

**Verification**:
```bash
# 1. Files gone
ls *.md
# Expected: Only README.md and RITE_SKILL_MATRIX.md remain

# 2. Tests still pass
just test

# 3. No broken references
grep -r "IMPLEMENTATION_SUMMARY\|PHASE-5-HANDOFF\|CONTEXT_SEED" . --include="*.sh" --include="*.go" --include="*.yaml" 2>/dev/null || echo "PASS: No references found"
```

**Rollback**:
```bash
git checkout HEAD -- CONTEXT_SEED.md DEFECT-D002-RESOLUTION.md \
  IMPLEMENTATION-SUMMARY-sails-status.md IMPLEMENTATION_SUMMARY.md \
  IMPLEMENTATION_VERIFICATION.md PHASE-5-HANDOFF.md SAILS-STATUS-REFERENCE.md
```

**Commit Message**:
```
chore(cleanup): remove transitional implementation docs from root

Phase 4 cleanup item RF-002. These files were created during Phase 1-3
implementation and are now obsolete:
- CONTEXT_SEED.md
- DEFECT-D002-RESOLUTION.md
- IMPLEMENTATION-SUMMARY-sails-status.md
- IMPLEMENTATION_SUMMARY.md
- IMPLEMENTATION_VERIFICATION.md
- PHASE-5-HANDOFF.md
- SAILS-STATUS-REFERENCE.md

README.md and RITE_SKILL_MATRIX.md intentionally preserved.
```

---

## Phase 3: rites/ Directory Cleanup (Low Risk)

### RF-003: Remove Misplaced Implementation Artifacts from rites/

**Risk**: LOW - Planning documents in wrong location, no runtime impact

**Before State**:
```
/rites/AUDIT_SUMMARY.md               8.2KB
/rites/IMPLEMENTATION_BLUEPRINT.md    14KB

Total: ~22KB in 2 files
```

**After State**:
```
(Both files removed)
```

**Invariants**:
- No code references these files
- rites/ directory structure unchanged (only removes .md files at root level)
- All rite manifests (rites/*/manifest.yaml) remain
- All tests continue to pass

**Commands**:
```bash
rm rites/AUDIT_SUMMARY.md
rm rites/IMPLEMENTATION_BLUEPRINT.md
```

**Verification**:
```bash
# 1. Files gone, manifests remain
ls rites/*.md 2>/dev/null && echo "FAIL: Files still exist" || echo "PASS: Files removed"
ls rites/*/manifest.yaml | wc -l
# Expected: 12 (one per rite)

# 2. Tests still pass
just test

# 3. swap-rite still works
./swap-rite.sh --list
```

**Rollback**:
```bash
git checkout HEAD -- rites/AUDIT_SUMMARY.md rites/IMPLEMENTATION_BLUEPRINT.md
```

**Commit Message**:
```
chore(cleanup): remove implementation artifacts from rites/

Phase 4 cleanup item RF-003. Planning documents that should not be in
rites/ directory:
- AUDIT_SUMMARY.md
- IMPLEMENTATION_BLUEPRINT.md

All rite manifests and team content preserved.
```

---

## Deferred Items (Priority 2-3)

The following items require additional verification or are intentionally deferred:

### DEFERRED: sync-user-*.sh Scripts (P2)

**Reason for Deferral**: These scripts sync to `~/.claude/` (user-level), not `.claude/` (project-level). They serve a DIFFERENT purpose than `ari sync materialize`.

**Scope Clarification**:
| Tool | Target | Purpose |
|------|--------|---------|
| `ari sync materialize` | `.claude/` (project) | Generate project config from templates |
| `sync-user-*.sh` | `~/.claude/` (user) | Sync roster resources to user directory |

**Recommendation**: Document this distinction. Consider future consolidation into `ari sync user` subcommand.

**No Action Required This Phase**.

### DEFERRED: roster-sync Script (P2)

**Reason for Deferral**: roster-sync handles satellite project synchronization, a use case not covered by `ari sync materialize`.

**Scope Clarification**:
| Tool | Use Case |
|------|----------|
| `ari sync materialize` | Generate .claude/ in roster repo |
| `roster-sync` | Sync roster config TO satellite projects |

**Recommendation**: Document scope boundaries. Consider future migration to `ari sync push/pull` for satellite projects.

**No Action Required This Phase**.

### DEFERRED: swap-rite.sh Modularization (P3)

**Reason for Deferral**: Script remains functional. lib/rite/ modules extracted but not yet integrated. Full Go migration would be a larger undertaking.

**Current State**: 3,773 lines, functional, tested via `--list` and `--verify`

**Recommendation**: Keep as-is for Phase 4. Consider Phase 5 for lib/rite/ integration or Go migration.

**No Action Required This Phase**.

---

## Risk Matrix

| Phase | Item | Blast Radius | Failure Detection | Recovery Cost |
|-------|------|--------------|-------------------|---------------|
| 1 | RF-001 Backup Dir | 1 dir, 6.8MB | Instant (ls) | None needed |
| 2 | RF-002 Root Docs | 7 files | git status | git checkout |
| 3 | RF-003 rites/ Docs | 2 files | git status | git checkout |

**Overall Risk**: LOW

All changes are file deletions with:
- Zero code impact
- Zero test impact
- Instant verification
- Trivial rollback

---

## Execution Sequence

```
Phase 1: RF-001 (backup dir)
    |
    +-- VERIFY: tests, build, swap-rite, materialize
    |
    +-- COMMIT
    |
    +-- ROLLBACK POINT 1
    |
    v
Phase 2: RF-002 (root docs)
    |
    +-- VERIFY: tests, no broken refs
    |
    +-- COMMIT
    |
    +-- ROLLBACK POINT 2
    |
    v
Phase 3: RF-003 (rites/ docs)
    |
    +-- VERIFY: tests, manifests intact
    |
    +-- COMMIT
    |
    +-- ROLLBACK POINT 3
    |
    v
COMPLETE
```

---

## Janitor Notes

### Commit Conventions
- Use `chore(cleanup):` prefix for all commits
- Reference "Phase 4 cleanup item RF-XXX" in body
- Keep commits atomic (one logical change per commit)

### Test Requirements
- Run `just test` after EACH commit
- Run `just build` after EACH commit
- If tests fail, STOP and report

### Critical Ordering
1. Phase 1 MUST complete before Phase 2
2. Phase 2 MUST complete before Phase 3
3. Do NOT combine phases into single commits

### What NOT To Do
- Do NOT delete README.md or RITE_SKILL_MATRIX.md
- Do NOT modify any code files
- Do NOT touch sync-user-*.sh or roster-sync (deferred)
- Do NOT modify swap-rite.sh (deferred)

---

## Post-Cleanup Verification Checklist

After ALL phases complete:

```bash
# 1. Tests pass
just test
# Expected: All OK

# 2. Build succeeds
just build
# Expected: ari binary created

# 3. swap-rite.sh functional
./swap-rite.sh --list
./swap-rite.sh --verify
# Expected: Lists rites, verify passes

# 4. ari sync materialize functional
./ari sync materialize --rite hygiene
# Expected: .claude/ generated

# 5. Git status clean (after commits)
git status
# Expected: No unexpected changes

# 6. Root directory clean
ls *.md
# Expected: Only README.md, RITE_SKILL_MATRIX.md

# 7. rites/ directory clean
ls rites/*.md 2>/dev/null
# Expected: No files (only directories)

# 8. Backup directory gone
ls -la .claude.backup.pre-materialize 2>/dev/null
# Expected: "No such file or directory"
```

---

## Verification Attestation

| Item | Verified Via | Attestation |
|------|--------------|-------------|
| .claude.backup.pre-materialize/ size | `du -sh` | 6.8MB confirmed |
| Root *.md file count | `ls -la` | 9 files, 7 transitional |
| rites/*.md file count | `ls -la` | 2 files (AUDIT_SUMMARY, IMPLEMENTATION_BLUEPRINT) |
| sync-user-*.sh purpose | Read header | User-level sync to ~/.claude/ |
| roster-sync purpose | Read header | Satellite project sync |
| Tests baseline | `just test` | All packages pass |
| Build baseline | `just build` | ari binary created |
| swap-rite.sh functional | `--help` | Working |

---

## Session Context

- **Initiative**: knossos-finalization
- **Phase**: 4 (HYGIENE)
- **Agent**: architect-enforcer
- **Upstream**: Code Smeller (SMELL-REPORT-phase4.md)
- **Downstream**: Janitor receives this plan for execution
- **Handoff**: Ready when Janitor confirms receipt
