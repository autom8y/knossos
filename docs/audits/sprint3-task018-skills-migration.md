# Sprint 3 Task 018: Skills Migration Audit

> Migration of 11 skills from skeleton_claude to roster
> Date: 2026-01-03
> Sprint: Skeleton Deprecation Sprint 3

## Overview

This document records the migration of 11 skills from `~/Code/skeleton_claude/.claude/skills/` to `/Users/tomtenuta/Code/roster/.claude/skills/` as part of the Skeleton Deprecation initiative.

## Migration Summary

| Status | Count |
|--------|-------|
| Successfully Migrated | 11 |
| Path Replacements Made | 3 |
| Issues Encountered | 0 |

## Skills Migrated

### Batch 1: Core Workflow Skills

#### 1. state-mate
- **Source**: `~/Code/skeleton_claude/.claude/skills/state-mate/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/state-mate/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: Centralized state mutation skill, critical for session/sprint management

#### 2. task-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/task-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/task-ref/`
- **Files Copied**:
  - `skill.md`
  - `examples/scenarios.md`
- **Path Replacements**: None required
- **Notes**: Full lifecycle task execution workflow

#### 3. sprint-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/sprint-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/sprint-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: Multi-task sprint orchestration

### Batch 2: Development Workflow Skills

#### 4. commit-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/commit-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/commit-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: AI-assisted commits with session tracking

#### 5. pr-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/pr-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/pr-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: Pull request workflow with test plan generation

#### 6. qa-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/qa-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/qa-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: Validation-only session with adversarial testing

#### 7. review
- **Source**: `~/Code/skeleton_claude/.claude/skills/review/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/review/`
- **Files Copied**:
  - `skill.md`
  - `references/output-format.md`
  - `references/review-prompt-template.md`
  - `references/examples/full-review.md`
- **Path Replacements**: None required
- **Notes**: Code review workflow with structured feedback

#### 8. hotfix-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/hotfix-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/hotfix-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: Rapid fix workflow for urgent production issues

### Batch 3: Utility Skills

#### 9. spike-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/spike-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/spike-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**: None required
- **Notes**: Time-boxed research producing no production code

#### 10. worktree-ref
- **Source**: `~/Code/skeleton_claude/.claude/skills/worktree-ref/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/worktree-ref/`
- **Files Copied**:
  - `skill.md`
- **Path Replacements**:
  1. Line 64: `skeleton_claude` -> `roster-sync` (CEM init reference)
  2. Line 260: `/Users/user/Code/skeleton_claude` -> `/Users/user/Code/roster` (status output example)
  3. Lines 325-327: Updated troubleshooting section to reference `roster-sync` instead of `skeleton_claude/cem`
- **Notes**: Git worktree management for parallel sessions

#### 11. documentation
- **Source**: `~/Code/skeleton_claude/.claude/skills/documentation/`
- **Target**: `/Users/tomtenuta/Code/roster/.claude/skills/documentation/`
- **Files Copied**:
  - `SKILL.md`
  - `workflow.md`
- **Path Replacements**: None required
- **Notes**: Documentation standards routing hub

## Path Replacement Details

The following patterns were searched for and replaced where necessary:

| Old Pattern | New Pattern | Occurrences Found |
|-------------|-------------|-------------------|
| `~/Code/skeleton_claude/` | (roster-native paths) | 3 |
| `skeleton_claude` | (roster-native references) | 3 |
| `cem sync` | `roster-sync sync` | 0 |
| `cem init` | `roster-sync init` | 0 |
| `SKELETON_HOME` | (removed) | 0 |

All replacements were made in `worktree-ref/skill.md`.

## Skills NOT Migrated (Already Present)

These skills already existed in roster and were intentionally kept as-is:

| Skill | Reason |
|-------|--------|
| 10x-ref | Roster-native version exists |
| 10x-workflow | Roster-native version exists |
| architect-ref | Roster-native version exists |
| atuin-desktop | Roster-native version exists |
| build-ref | Roster-native version exists |
| doc-artifacts | Roster-native version exists |
| justfile | Roster-native version exists |

## Verification

### Post-Migration Checks

1. **All skill directories exist**:
   ```
   state-mate/      - Present
   task-ref/        - Present
   sprint-ref/      - Present
   commit-ref/      - Present
   pr-ref/          - Present
   qa-ref/          - Present
   review/          - Present
   hotfix-ref/      - Present
   spike-ref/       - Present
   worktree-ref/    - Present
   documentation/   - Present
   ```

2. **No skeleton-specific paths remain**:
   - Searched all migrated skills for: `skeleton_claude`, `cem sync`, `cem init`, `SKELETON_HOME`
   - Result: No matches found in migrated skills

3. **Skills accessible via Skill tool**:
   - Skills are registered in `.claude/skills/` directory
   - Available for invocation

## Final Roster Skills Directory

After migration, roster now has 34 skill directories including the 11 newly migrated skills.

## Recommendations

1. **Test worktree-ref skill** - The path replacements should be validated in a real worktree creation scenario
2. **Update .rite-skills if needed** - If these skills should be auto-included in team packs, update team skill configurations
3. **Consider deprecation notice in skeleton** - Add notes in skeleton indicating these skills have been migrated to roster

## Attestation

| File | Verified | Path |
|------|----------|------|
| state-mate/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/state-mate/skill.md |
| task-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/task-ref/skill.md |
| task-ref/examples/scenarios.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/task-ref/examples/scenarios.md |
| sprint-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/sprint-ref/skill.md |
| commit-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/commit-ref/skill.md |
| pr-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/pr-ref/skill.md |
| qa-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/qa-ref/skill.md |
| review/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/review/skill.md |
| review/references/output-format.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/review/references/output-format.md |
| review/references/review-prompt-template.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/review/references/review-prompt-template.md |
| review/references/examples/full-review.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/review/references/examples/full-review.md |
| hotfix-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/hotfix-ref/skill.md |
| spike-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/spike-ref/skill.md |
| worktree-ref/skill.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/worktree-ref/skill.md |
| documentation/SKILL.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/documentation/SKILL.md |
| documentation/workflow.md | YES | /Users/tomtenuta/Code/roster/.claude/skills/documentation/workflow.md |
