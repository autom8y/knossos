# Path Reference Audit Triage Report

**Date**: 2026-01-02
**Triggered By**: UserPromptSubmit hook failure on `/spike` command
**Scope**: Full path reference audit across hooks, libraries, skills, agents, and commands

---

## Executive Summary

The harness has **7 critical blocking issues** and **12 warning-level issues** stemming from an incomplete categorical migration of hooks and skills. The immediate blocker is that 2 hook scripts referenced in settings.local.json do not exist at the expected paths.

| Category | Critical | Warning | Compliant |
|----------|----------|---------|-----------|
| Hook Paths | 2 | 10 | 10 |
| Library Paths | 2 | 0 | 33 |
| Skill/Agent/Command Paths | 3 | 2 | 20+ |
| **TOTAL** | **7** | **12** | **63+** |

---

## Critical Issues (BLOCKING)

### Issue 1: Missing Hook Scripts at Root Level
**Severity**: CRITICAL - Blocks all /start, /sprint, /task commands
**Root Cause**: Hooks moved to subdirectories but settings.local.json still references root paths

| File | Expected Path | Actual Location |
|------|---------------|-----------------|
| orchestrator-router.sh | `.claude/hooks/orchestrator-router.sh` | `.claude/hooks/validation/orchestrator-router.sh` |
| orchestrator-bypass-check.sh | `.claude/hooks/orchestrator-bypass-check.sh` | `.claude/hooks/validation/orchestrator-bypass-check.sh` |

**Evidence**:
- `settings.local.json` line 188: `$CLAUDE_PROJECT_DIR/.claude/hooks/orchestrator-router.sh`
- `settings.local.json` line 149: `$CLAUDE_PROJECT_DIR/.claude/hooks/orchestrator-bypass-check.sh`
- Files exist only in `validation/` subdirectory

**Impact**: UserPromptSubmit hooks fail silently, breaking orchestrator routing for workflow commands.

---

### Issue 2: Installation Scripts Reference Wrong Source Directory
**Severity**: CRITICAL - Blocks hook installation to new projects
**Root Cause**: ADR-0002 specifies `roster/hooks/` but scripts reference `roster/user-hooks/`

| Script | Line | Current | Expected (ADR-0002) |
|--------|------|---------|---------------------|
| install-hooks.sh | 20 | `$ROSTER_HOME/user-hooks` | `$ROSTER_HOME/hooks` |
| sync-user-hooks.sh | 29 | `$ROSTER_HOME/user-hooks` | `$ROSTER_HOME/hooks` |

**Evidence**:
- ADR-0002 Section 1: "Canonical template source must be at `roster/hooks/`"
- Directory `roster/hooks/` does not exist
- Directory `roster/user-hooks/` exists and is used instead

**Impact**: Cannot install hooks to new projects or sync to user ~/.claude/hooks/

---

### Issue 3: Missing ADR-0005
**Severity**: CRITICAL - Documentation reference broken
**Root Cause**: ADR-0005 referenced in CLAUDE.md but file was never created

| Reference Location | Referenced Path |
|-------------------|-----------------|
| .claude/CLAUDE.md line 119 | `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md` |

**Evidence**:
- Only ADR-0001, ADR-0002, ADR-0006 exist in `docs/decisions/`
- ADR-0003, ADR-0004, ADR-0005 are missing (sequence gap)

**Impact**: Users cannot find state-mate architecture documentation.

---

### Issue 4: Missing @documentation Skill in .claude/skills/
**Severity**: HIGH - Skill invocation broken
**Root Cause**: Skill exists in user-skills/ but referenced as if in .claude/skills/

| Reference Location | Expected | Actual |
|-------------------|----------|--------|
| .claude/CLAUDE.md line 75 | `.claude/skills/documentation/` | `user-skills/documentation/` |
| Orchestrator references | `@documentation` | Not found in active skills |

**Impact**: Documentation skill cannot be invoked via standard pattern.

---

### Issue 5: Missing @10x-workflow Skill
**Severity**: HIGH - Orchestrator references broken
**Root Cause**: Skill referenced by orchestrators but not present

| Reference Location | Referenced Skill |
|-------------------|------------------|
| ecosystem-pack orchestrator.yaml | `@10x-workflow` |
| security-pack orchestrator.yaml | `@10x-workflow` |

**Evidence**:
- Not in `.claude/skills/`
- May exist in `.claude/skills.orphan-backup/10x-workflow/`

**Impact**: Orchestrators cannot route to 10x-workflow coordination patterns.

---

## Warning Issues (Non-Blocking)

### Duplicate Hook Content
**10 hooks exist in both root and subdirectories with DIFFERENT content**

| Hook | Root Location | Subdirectory Location |
|------|---------------|----------------------|
| session-context.sh | `.claude/hooks/` | `.claude/hooks/context-injection/` |
| coach-mode.sh | `.claude/hooks/` | `.claude/hooks/context-injection/` |
| auto-park.sh | `.claude/hooks/` | `.claude/hooks/session-guards/` |
| artifact-tracker.sh | `.claude/hooks/` | `.claude/hooks/tracking/` |
| session-audit.sh | `.claude/hooks/` | `.claude/hooks/tracking/` |
| commit-tracker.sh | `.claude/hooks/` | `.claude/hooks/tracking/` |
| delegation-check.sh | `.claude/hooks/` | `.claude/hooks/validation/` |
| command-validator.sh | `.claude/hooks/` | `.claude/hooks/validation/` |
| session-write-guard.sh | `.claude/hooks/` | `.claude/hooks/session-guards/` |
| start-preflight.sh | `.claude/hooks/` | `.claude/hooks/session-guards/` |

**Issue**: MD5 checksums differ between root and subdirectory versions.

---

### Orphaned Skills (22+ in backup)
**Skills moved to .claude/skills.orphan-backup/ but some are still referenced**

Key orphaned skills:
- 10x-ref, 10x-workflow
- architect-ref, build-ref
- debt-ref, hygiene-ref, intelligence-ref, security-ref, sre-ref, strategy-ref
- doc-* (multiple consolidation skills)

---

### ADR Sequence Gap
**ADRs 0003, 0004, 0005 missing from sequence**

| ADR | Status |
|-----|--------|
| ADR-0001 | ✓ Exists (session-state-machine-redesign) |
| ADR-0002 | ✓ Exists (hook-library-resolution-architecture) |
| ADR-0003 | ✗ Missing |
| ADR-0004 | ✗ Missing |
| ADR-0005 | ✗ Missing (state-mate - referenced in CLAUDE.md) |
| ADR-0006 | ✓ Exists (categorical-resource-organization) |

---

## Root Cause Analysis

### Primary Root Cause: Incomplete Categorical Migration

The harness is mid-transition from:
- **Old Model**: Flat hooks in `.claude/hooks/*.sh`
- **New Model**: Categorical hooks in `.claude/hooks/{category}/*.sh`

**Evidence of incomplete migration**:
1. ADR-0006 specifies categorical organization (validation/, session-guards/, context-injection/, tracking/)
2. Subdirectory structure was created and hooks were copied
3. Root-level copies were NOT removed
4. Two new hooks (orchestrator-router, orchestrator-bypass-check) were added ONLY to subdirectories
5. settings.local.json generation still expects root-level paths
6. base_hooks.yaml still uses flat paths (e.g., `path: orchestrator-router.sh`)

### Secondary Root Cause: ADR-0002 Implementation Gap

ADR-0002 defines canonical source at `roster/hooks/` but:
1. Directory `roster/hooks/` was never created
2. Installation scripts reference `roster/user-hooks/` (legacy location)
3. This creates confusion about source of truth

---

## Compliance Summary

### Hook Path Resolution (ADR-0002 Compliance: 75%)

| Requirement | Status |
|-------------|--------|
| Canonical source at roster/hooks/ | ✗ FAIL |
| Runtime CLAUDE_PROJECT_DIR usage | ✓ PASS |
| Library relative paths | ✓ PASS |
| Graceful fallback mechanisms | ✓ PASS |
| Installation script correctness | ✗ FAIL |

### Categorical Organization (ADR-0006 Compliance: 60%)

| Requirement | Status |
|-------------|--------|
| Subdirectory structure created | ✓ PASS |
| Hooks categorized correctly | ✓ PASS |
| Root-level copies removed | ✗ FAIL |
| Path references updated | ✗ FAIL |
| New hooks documented in ADR | ✗ FAIL |

---

## Recommended Fix Sequence

### Phase 1: Immediate Unblock (15 min)

**Option A: Copy missing hooks to root (Quick fix)**
```bash
cp .claude/hooks/validation/orchestrator-router.sh .claude/hooks/
cp .claude/hooks/validation/orchestrator-bypass-check.sh .claude/hooks/
```

**Option B: Update base_hooks.yaml paths (Proper fix)**
```yaml
# Change from:
path: orchestrator-router.sh
# To:
path: validation/orchestrator-router.sh
```

Then regenerate settings.local.json via `swap-team.sh ecosystem-pack`

### Phase 2: Resolve Duplicates (30 min)

1. Audit which version is canonical (root vs subdirectory) via git history
2. Remove duplicate hooks (keep either root OR subdirectory)
3. Update base_hooks.yaml to use correct paths
4. Regenerate settings.local.json

### Phase 3: Fix Installation Scripts (15 min)

1. Create `roster/hooks/` directory (or rename `roster/user-hooks/`)
2. Update install-hooks.sh line 20
3. Update sync-user-hooks.sh line 29
4. Test installation to new project

### Phase 4: Documentation Cleanup (30 min)

1. Create ADR-0005 for state-mate
2. Move @documentation skill to .claude/skills/ OR update CLAUDE.md reference
3. Restore @10x-workflow from orphan-backup OR update orchestrator references
4. Update ADR-0006 with orchestrator-router and orchestrator-bypass-check

### Phase 5: Validation (20 min)

1. Run `/spike test` to verify UserPromptSubmit hooks work
2. Run `/start test` to verify orchestrator routing works
3. Run hook installation to fresh project
4. Verify all CLAUDE.md references resolve

---

## Files Requiring Changes

### Critical (Must Fix)

| File | Change Required |
|------|-----------------|
| `.claude/hooks/base_hooks.yaml` | Update paths to include subdirectory OR ensure root copies exist |
| `.claude/settings.local.json` | Regenerate after base_hooks.yaml fix |
| `install-hooks.sh` line 20 | Change `user-hooks` to `hooks` |
| `sync-user-hooks.sh` line 29 | Change `user-hooks` to `hooks` |
| `docs/decisions/ADR-0005-*.md` | Create missing file |

### Warning (Should Fix)

| File | Change Required |
|------|-----------------|
| `.claude/hooks/*.sh` (10 files) | Remove root copies OR subdirectory copies |
| `.claude/CLAUDE.md` line 75 | Update documentation skill reference |
| `docs/decisions/ADR-0006-*.md` | Add orchestrator hooks to spec |
| Orchestrator.yaml files | Update skill references |

---

## Verification Checklist

After fixes are applied:

- [ ] `/spike "test"` executes without hook errors
- [ ] `/start "test"` triggers orchestrator routing
- [ ] `/task "test"` triggers orchestrator routing
- [ ] `install-hooks.sh` successfully installs to new project
- [ ] `sync-user-hooks.sh` successfully syncs to ~/.claude/hooks/
- [ ] All hooks in settings.local.json resolve to existing files
- [ ] No duplicate hooks (root OR subdirectory, not both)
- [ ] ADR-0005 exists and is readable
- [ ] @documentation skill is invocable
- [ ] @10x-workflow skill is invocable

---

## Appendix: Full File Inventory

### Hooks at Root Level (10 files)
```
.claude/hooks/artifact-tracker.sh
.claude/hooks/auto-park.sh
.claude/hooks/coach-mode.sh
.claude/hooks/command-validator.sh
.claude/hooks/commit-tracker.sh
.claude/hooks/delegation-check.sh
.claude/hooks/session-audit.sh
.claude/hooks/session-context.sh
.claude/hooks/session-write-guard.sh
.claude/hooks/start-preflight.sh
```

### Hooks in Subdirectories (12 files)
```
.claude/hooks/context-injection/coach-mode.sh
.claude/hooks/context-injection/session-context.sh
.claude/hooks/session-guards/auto-park.sh
.claude/hooks/session-guards/session-write-guard.sh
.claude/hooks/session-guards/start-preflight.sh
.claude/hooks/tracking/artifact-tracker.sh
.claude/hooks/tracking/commit-tracker.sh
.claude/hooks/tracking/session-audit.sh
.claude/hooks/validation/command-validator.sh
.claude/hooks/validation/delegation-check.sh
.claude/hooks/validation/orchestrator-bypass-check.sh  ← ONLY HERE
.claude/hooks/validation/orchestrator-router.sh        ← ONLY HERE
```

### Hook Libraries (11 files)
```
.claude/hooks/lib/config.sh
.claude/hooks/lib/hooks-init.sh
.claude/hooks/lib/logging.sh
.claude/hooks/lib/primitives.sh
.claude/hooks/lib/session-core.sh
.claude/hooks/lib/session-fsm.sh
.claude/hooks/lib/session-manager.sh
.claude/hooks/lib/session-migrate.sh
.claude/hooks/lib/session-state.sh
.claude/hooks/lib/session-utils.sh
.claude/hooks/lib/worktree-manager.sh
```

---

**Report Generated By**: 3 parallel Explore agents
**Agent IDs**: a8f5e08 (hooks), a1776f6 (libraries), aa13903 (skills/agents/commands)
