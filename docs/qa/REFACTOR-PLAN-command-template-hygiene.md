# Refactoring Plan: Command Template Standardization

**Date**: 2025-12-29
**Architect**: Architect Enforcer (hygiene-pack)
**Input**: SMELL-REPORT-command-template-hygiene.md
**Standard**: CONTEXT-DESIGN-command-argument-standardization.md

---

## Architectural Assessment

### Boundary Analysis

The command ecosystem has three distinct architectural boundaries:

1. **Standard Commands**: Pass-through or internal-parsed commands following the canonical template (32 files)
2. **Meta Commands**: Session initialization templates using `{TAG}` pattern (3 files: minus-1.md, zero.md, one.md)
3. **Team Commands**: Team-specific commands under `teams/*/commands/` (variable)

**Root Cause Analysis**: The smells cluster into two root causes:

1. **Organic Growth Pattern**: Early commands were written before the canonical template was defined. The meta commands especially predate the frontmatter convention.
2. **Template Drift**: team-switching commands were copied from a template that used inline bullet flags instead of table format.

### Smell Classification

| Smell | Classification | Root Cause |
|-------|---------------|------------|
| S1: Missing argument-hint | BOUNDARY | Meta commands predate frontmatter convention |
| S2: Non-standard $ARGUMENTS | LOCAL | Single file formatting issue |
| S3: Section order violations | LOCAL | Style drift, low impact |
| S4: Non-standard flag format | MODULE | Template drift across team-switching category |
| S6: Missing frontmatter fields | BOUNDARY | Mixed: meta commands (boundary), others (local) |

---

## Decision Records

### D1: Meta Commands Pattern

**Decision**: B - Add minimal frontmatter but preserve `{TAG}` pattern

**Rationale**:
1. The meta commands (`minus-1.md`, `zero.md`, `one.md`) serve a fundamentally different purpose: they are session orchestration templates, not action commands
2. The `{TAG}` pattern is intentional design - it signals "this is a placeholder for the initiative description"
3. Converting to `$ARGUMENTS` would obscure the semantic intent and break established muscle memory
4. Adding frontmatter provides discoverability (description in `/help`) without changing behavior

**Invariants**:
- `{TAG}` pattern preserved exactly
- No behavioral change to session orchestration
- Commands appear correctly in help listings

**Trade-off Accepted**: Minor template inconsistency for semantic clarity

### D2: Flag Documentation Format

**Decision**: A - Convert all to tables with "Handled By" column (strict compliance)

**Rationale**:
1. Consistency enables tooling (automatic flag extraction, validation)
2. "Handled By" column clarifies responsibility boundaries (swap-team.sh vs internal vs Claude)
3. The 11 affected files all follow the same pattern - batch conversion is low-risk
4. Table format is more scannable than bullet lists for complex flag sets

**Invariants**:
- Same flags documented (no additions/removals)
- Same descriptions preserved
- Flags section moved from inline in Behavior to dedicated `## Flags` section

**Trade-off Accepted**: Slightly more verbose format for consistency

### D3: Missing allowed-tools

**Decision**: A - Add `allowed-tools: Bash, Read` to all commands that call external scripts

**Rationale**:
1. Explicit tool declarations improve predictability
2. Prevents accidental privilege escalation if Claude model behavior changes
3. Currently only `team.md` is missing this among commands that use Bash
4. Low cost, high clarity

**Invariants**:
- Only add tools that are actually used
- Do not add unnecessary tool permissions

---

## Refactoring Contracts

### Phase 1: Critical Frontmatter (HIGH priority, LOW risk)

#### RF-001: Add frontmatter to minus-1.md

**File**: `/roster/user-commands/meta/minus-1.md`

**Before State** (lines 1-2):
```markdown
# Session -1: Initiative Assessment

You are a **prompter**...
```

**After State**:
```markdown
---
description: Assess initiative readiness before Session 0 planning
argument-hint: <initiative>
model: opus
---

# Session -1: Initiative Assessment

You are a **prompter**...
```

**Invariants**:
- All content after frontmatter unchanged
- `{TAG}` pattern preserved in body
- No behavioral change

**Verification**:
1. Run: `head -10 user-commands/meta/minus-1.md`
2. Confirm frontmatter block present with three required fields
3. Confirm `# Session -1:` header follows frontmatter

**Rollback**: `git checkout user-commands/meta/minus-1.md`

---

#### RF-002: Add frontmatter to zero.md

**File**: `/roster/user-commands/meta/zero.md`

**Before State** (lines 1-2):
```markdown
# Session 0: Orchestrator Initialization

You are a **prompter**...
```

**After State**:
```markdown
---
description: Initialize Orchestrator with 4-agent workflow plan
argument-hint: <initiative>
model: opus
---

# Session 0: Orchestrator Initialization

You are a **prompter**...
```

**Invariants**:
- All content after frontmatter unchanged
- `{TAG}` pattern preserved in body
- No behavioral change

**Verification**:
1. Run: `head -10 user-commands/meta/zero.md`
2. Confirm frontmatter block present with three required fields
3. Confirm `# Session 0:` header follows frontmatter

**Rollback**: `git checkout user-commands/meta/zero.md`

---

#### RF-003: Add frontmatter to one.md

**File**: `/roster/user-commands/meta/one.md`

**Before State** (lines 1-2):
```markdown
# Session 1: Autonomous Execution

You are a **prompter**...
```

**After State**:
```markdown
---
description: Execute workflow phases autonomously via daisy-chain loop
argument-hint: (uses session context from Session 0)
model: opus
---

# Session 1: Autonomous Execution

You are a **prompter**...
```

**Note**: `one.md` uses implicit session context rather than explicit argument, so argument-hint describes this.

**Invariants**:
- All content after frontmatter unchanged
- No behavioral change

**Verification**:
1. Run: `head -10 user-commands/meta/one.md`
2. Confirm frontmatter block present with three required fields
3. Confirm `# Session 1:` header follows frontmatter

**Rollback**: `git checkout user-commands/meta/one.md`

---

#### RF-004: Add model to sync.md

**File**: `/roster/user-commands/cem/sync.md`

**Before State** (lines 1-5):
```yaml
---
description: Sync project with skeleton_claude ecosystem
argument-hint: [init|sync|status|diff|install-user]
allowed-tools: Bash, Read
---
```

**After State**:
```yaml
---
description: Sync project with skeleton_claude ecosystem
argument-hint: [init|sync|status|diff|install-user] [--refresh] [--force] [--dry-run]
allowed-tools: Bash, Read
model: sonnet
---
```

**Note**: Also expanding argument-hint to document the flags that sync.md accepts.

**Invariants**:
- No behavioral change
- CEM's `--refresh` flag preserved (distinct from swap-team.sh's deprecated flag)

**Verification**:
1. Run: `head -6 user-commands/cem/sync.md`
2. Confirm `model: sonnet` present

**Rollback**: `git checkout user-commands/cem/sync.md`

---

#### RF-005: Add allowed-tools to team.md

**File**: `/roster/user-commands/navigation/team.md`

**Before State** (lines 1-5):
```yaml
---
description: Switch agent team packs or list available teams
argument-hint: [pack-name] [--list] [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
model: sonnet
---
```

**After State**:
```yaml
---
description: Switch agent team packs or list available teams
argument-hint: [pack-name] [--list] [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: sonnet
---
```

**Invariants**:
- No behavioral change
- Field order matches canonical template (description, argument-hint, allowed-tools, model)

**Verification**:
1. Run: `head -6 user-commands/navigation/team.md`
2. Confirm `allowed-tools: Bash, Read` present

**Rollback**: `git checkout user-commands/navigation/team.md`

---

#### RF-006: Add argument-hint to cem-debug.md

**File**: `/roster/teams/ecosystem-pack/commands/cem-debug.md`

**Before State** (lines 1-5):
```yaml
---
description: Diagnose CEM sync issues and conflicts (Ecosystem Analyst with CEM focus)
allowed-tools: Bash, Read, Grep, Glob
model: opus
---
```

**After State**:
```yaml
---
description: Diagnose CEM sync issues and conflicts (Ecosystem Analyst with CEM focus)
argument-hint: [issue-description]
allowed-tools: Bash, Read, Grep, Glob
model: opus
---
```

**Invariants**:
- No behavioral change
- Implicit argument now documented

**Verification**:
1. Run: `head -6 teams/ecosystem-pack/commands/cem-debug.md`
2. Confirm `argument-hint: [issue-description]` present

**Rollback**: `git checkout teams/ecosystem-pack/commands/cem-debug.md`

---

### Phase 2: Flag Table Standardization (MEDIUM priority, LOW risk)

All 11 files in this phase follow the same transformation pattern.

#### RF-007 through RF-016: Team-switching flag tables

**Affected Files**:
- RF-007: `user-commands/team-switching/10x.md`
- RF-008: `user-commands/team-switching/hygiene.md`
- RF-009: `user-commands/team-switching/rnd.md`
- RF-010: `user-commands/team-switching/debt.md`
- RF-011: `user-commands/team-switching/intelligence.md`
- RF-012: `user-commands/team-switching/strategy.md`
- RF-013: `user-commands/team-switching/docs.md`
- RF-014: `user-commands/team-switching/sre.md`
- RF-015: `user-commands/team-switching/security.md`
- RF-016: `user-commands/navigation/ecosystem.md`

**Common Before State** (in Behavior section):
```markdown
## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh {pack-name} $ARGUMENTS`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `{pack-name}`

**Flags:**
- `--update`, `-u`: Pull latest agent definitions from roster even if already on team
- `--dry-run`: Preview changes without applying
- `--keep-all`: Preserve all orphan agents in project
- `--remove-all`: Remove all orphans (backup available)
- `--promote-all`: Move all orphans to user-level
```

**Common After State** (separate Flags section after Behavior):
```markdown
## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh {pack-name} $ARGUMENTS`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `{pack-name}`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on team | swap-team.sh |
| `--dry-run` | - | Preview changes without applying | swap-team.sh |
| `--keep-all` | - | Preserve all orphan agents in project | swap-team.sh |
| `--remove-all` | - | Remove all orphans (backup available) | swap-team.sh |
| `--promote-all` | - | Move all orphans to user-level | swap-team.sh |
```

**Invariants**:
- Same five flags documented
- Same descriptions (may normalize minor wording)
- No behavioral change
- Section appears between Behavior and When to Use

**Verification** (for each file):
1. Run: `grep -A 10 "## Flags" user-commands/team-switching/{file}.md`
2. Confirm table format with 5 rows
3. Confirm "Handled By" column shows `swap-team.sh`

**Rollback**: `git checkout user-commands/team-switching/`

---

### Phase 3: Minor Fixes (LOW priority, LOW risk)

#### RF-017: Fix $ARGUMENTS placement in forge.md

**File**: `/roster/user-commands/team-switching/forge.md`

**Before State** (lines 8-12):
```markdown
## Your Task

Display information about The Forge - the meta-team for creating and maintaining agent teams.

**Arguments**: $ARGUMENTS
```

**After State**:
```markdown
## Your Task

Display information about The Forge - the meta-team for creating and maintaining agent teams. $ARGUMENTS
```

**Invariants**:
- Same content, different formatting
- $ARGUMENTS follows task description inline per standard

**Verification**:
1. Run: `grep -A 3 "## Your Task" user-commands/team-switching/forge.md`
2. Confirm `$ARGUMENTS` is inline, not separate labeled field

**Rollback**: `git checkout user-commands/team-switching/forge.md`

---

### Phase 4: Section Order (DEFERRED)

**Files**: 7 commands with non-standard section order

**Decision**: DEFER to future maintenance sprint

**Rationale**:
1. Section order violations are LOW severity (no functional impact)
2. Extra sections like "When to Use", "Complexity Levels", "Parallel Sprint Pattern" provide value
3. Reordering would require careful review of content dependencies
4. Risk of introducing errors exceeds benefit of consistency

**Future Action**: When these files are next modified for other reasons, normalize section order opportunistically.

**Deferred Files**:
- `user-commands/navigation/team.md`
- `user-commands/navigation/sessions.md`
- `user-commands/navigation/worktree.md`
- `user-commands/session/start.md`
- `user-commands/workflow/sprint.md`
- `teams/doc-team-pack/commands/consolidate.md`
- `teams/ecosystem-pack/commands/cem-debug.md`

---

## Risk Assessment

### Phase 1: Critical Frontmatter

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Frontmatter breaks command parsing | LOW | HIGH | Test each command after change |
| Wrong model assignment | LOW | MEDIUM | Use existing similar commands as reference |
| Missing required field | LOW | MEDIUM | Checklist verification |

**Blast Radius**: Single file per change
**Rollback Cost**: Single `git checkout`
**Overall Risk**: LOW

### Phase 2: Flag Table Standardization

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Table formatting error | MEDIUM | LOW | Template-based transformation |
| Flag description mismatch | LOW | LOW | Preserve exact wording |
| Missing flag | LOW | MEDIUM | Count flags before/after |

**Blast Radius**: Single file per change, but 10 similar changes
**Rollback Cost**: Single `git checkout` per file or `git checkout user-commands/team-switching/`
**Overall Risk**: LOW (mechanical transformation)

### Phase 3: Minor Fixes

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| $ARGUMENTS placement breaks parsing | VERY LOW | LOW | Simple text movement |

**Blast Radius**: Single file
**Rollback Cost**: Trivial
**Overall Risk**: VERY LOW

---

## Execution Sequence

```
Phase 1 ─────────────────────────────────────────────────────────────►
  RF-001 ──► RF-002 ──► RF-003 ──► RF-004 ──► RF-005 ──► RF-006
  (meta)     (meta)     (meta)     (sync)     (team)     (cem-debug)
                                      │
                                      ▼
                               [CHECKPOINT 1]
                               Commit: "fix(commands): add required frontmatter fields"
                                      │
                                      ▼
Phase 2 ─────────────────────────────────────────────────────────────►
  RF-007 through RF-016 (can be batched or individual)
                                      │
                                      ▼
                               [CHECKPOINT 2]
                               Commit: "refactor(commands): standardize flag documentation tables"
                                      │
                                      ▼
Phase 3 ─────────────────────────────────────────────────────────────►
  RF-017
                                      │
                                      ▼
                               [CHECKPOINT 3]
                               Commit: "fix(forge): inline $ARGUMENTS per standard"
```

**Phase Dependencies**: None - phases can be executed in parallel, but sequential execution allows cleaner rollback points.

**Commit Strategy**:
- One commit per phase for atomic rollback
- If Phase 2 is large, may split into two commits (team-switching + ecosystem.md)

---

## Janitor Notes

### Commit Conventions

Follow repository's existing commit style:
- `fix(scope): message` for fixes
- `refactor(scope): message` for structural changes
- Include co-authorship attribution

### Test Requirements

After each phase:
1. Verify frontmatter parses: `grep -l "^---$" user-commands/**/*.md | head -5`
2. Spot-check help display: Run `/help` in Claude session (if available)
3. No regression: Commands should behave identically

### Critical Ordering

1. **RF-001, RF-002, RF-003 first**: Meta commands are foundational to session workflow
2. **RF-004, RF-005, RF-006 next**: Complete frontmatter before structural changes
3. **RF-007-RF-016 batch**: All use same template, can parallelize
4. **RF-017 last**: Lowest priority, smallest change

### Edge Cases

1. **sync.md `--refresh`**: This is CEM's flag, NOT swap-team.sh's deprecated flag. Do not change.
2. **team.md mixed flags**: Has inline bullets AND orphan handling table. Only add `allowed-tools`, do not restructure.
3. **forge.md Context section**: Missing Context section. Do not add - it uses Read/Glob, not session context.

---

## Verification Checklist

Before marking Phase complete:

### Phase 1 Verification
- [ ] RF-001: `grep -c "^---$" user-commands/meta/minus-1.md` returns 2
- [ ] RF-002: `grep -c "^---$" user-commands/meta/zero.md` returns 2
- [ ] RF-003: `grep -c "^---$" user-commands/meta/one.md` returns 2
- [ ] RF-004: `grep "model:" user-commands/cem/sync.md` returns line
- [ ] RF-005: `grep "allowed-tools:" user-commands/navigation/team.md` returns line
- [ ] RF-006: `grep "argument-hint:" teams/ecosystem-pack/commands/cem-debug.md` returns line

### Phase 2 Verification
- [ ] RF-007-016: `grep -l "| Flag | Short |" user-commands/team-switching/*.md` returns 9 files
- [ ] RF-007-016: `grep -l "| Flag | Short |" user-commands/navigation/ecosystem.md` returns 1 file
- [ ] No inline `**Flags:**` remaining: `grep -l "\*\*Flags:\*\*" user-commands/team-switching/*.md` returns 0

### Phase 3 Verification
- [ ] RF-017: `grep "Arguments.*ARGUMENTS" user-commands/team-switching/forge.md` returns 0
- [ ] RF-017: `grep "\$ARGUMENTS$" user-commands/team-switching/forge.md` returns 1

---

## Summary

| Metric | Value |
|--------|-------|
| Total refactorings | 17 |
| Phases | 3 (+ 1 deferred) |
| Files modified | 17 |
| Risk level | LOW |
| Estimated effort | 30-45 minutes |
| Deferred items | 7 (section order) |

### Change Distribution

| Category | Count | Priority |
|----------|-------|----------|
| REQUIRED (frontmatter compliance) | 6 | HIGH |
| RECOMMENDED (flag table standard) | 10 | MEDIUM |
| OPTIONAL ($ARGUMENTS style) | 1 | LOW |
| DEFERRED (section order) | 7 | FUTURE |

---

## Handoff Criteria

- [x] All decisions documented with rationale (D1, D2, D3)
- [x] Before/after contracts defined for each refactoring (RF-001 through RF-017)
- [x] Invariants specified per refactoring
- [x] Verification criteria provided (grep commands)
- [x] Rollback points identified (per-phase commits)
- [x] Risk assessment complete (phase-level analysis)
- [x] Deferred items documented with rationale (section order)
- [x] Janitor notes for edge cases and commit conventions

**Ready for Janitor execution.**

---

*Plan generated by Architect Enforcer. Verified against smell report and canonical template.*
