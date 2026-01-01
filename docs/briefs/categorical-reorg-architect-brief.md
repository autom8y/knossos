# Architect Brief: Categorical Resource Organization

| Field | Value |
|-------|-------|
| **Session Type** | /architect |
| **Requested By** | User |
| **Priority** | High |
| **Related ADR** | ADR-0006 (Draft) |

## Objective

Design optimal categorical organization for all roster user-level resources (skills, hooks, commands) and produce a TDD for implementation.

## Scope

### In Scope
- Define domain-based categories for **skills** (24 resources)
- Confirm/optimize workflow-based categories for **commands** (31 resources, 7 existing categories)
- Define categories for **hooks** (10 root hooks + 10 lib utilities)
- Identify resources that should remain at root level (uncategorized)
- Design manifest schema updates
- Produce migration plan

### Out of Scope
- Team-level resources (agents managed by swap-team.sh)
- CEM infrastructure changes (install-user delegates correctly)
- Claude Code destination changes (remains flat)

## Constraints

### Confirmed Decisions (From User)
| Decision | Value |
|----------|-------|
| Migration approach | Big bang swap |
| Backward compatibility | None required |
| Naming convention | kebab-case |
| Skill internals | Minimal (SKILL.md required only) |
| Category independence | Each pillar defines its own categories |
| Manifest update | Add `category` field |
| Hooks lib/ | Preserve at root level |

### Technical Constraints
- Sync scripts flatten to `~/.claude/{type}/` (no nested categories at destination)
- Skills are directories with SKILL.md; commands/hooks are individual files
- Manifests use JSON format with checksum tracking
- rsync with --delete used for skill directory sync

## Current State

### Skills (24 total, currently flat)
```
commit-ref       documentation      initiative-scoping  park-ref       qa-ref      session-common  start-ref      wrap-ref
cross-team       file-verification  orchestration       pr-ref         resume      spike-ref       task-ref       worktree-ref
doc-artifacts    handoff-ref        orchestrator-templates  prompting  review      sprint-ref      standards
```

### Commands (31 total, 7 categories)
```
session/:        start, park, continue, handoff, wrap
workflow/:       task, sprint, hotfix
operations/:     architect, build, qa, code-review, commit
navigation/:     consult, team, worktree, sessions, ecosystem
meta/:           minus-1, zero, one
team-switching/: 10x, docs, hygiene, debt, sre, security, intelligence, rnd, strategy, forge
cem/:            sync
```

### Hooks (20 total, flat + lib/)
```
Root (10): session-context, coach-mode, auto-park, artifact-tracker,
           session-audit, commit-tracker, command-validator, delegation-check,
           start-preflight, session-write-guard

Lib (10):  config, logging, primitives, session-core, session-state,
           session-fsm, session-manager, session-migrate, session-utils,
           worktree-manager
```

## Key Questions to Answer

### 1. Skill Categories (Domain-Based)
- What domains best organize the 24 skills?
- Proposed groupings (validate/revise):
  - `session-lifecycle/`: start-ref, park-ref, resume, handoff-ref, wrap-ref, session-common
  - `documentation/`: doc-artifacts, documentation, standards
  - `orchestration/`: orchestration, orchestrator-templates, cross-team
  - `workflow-patterns/`: sprint-ref, task-ref, hotfix-ref, spike-ref, worktree-ref
  - `operations/`: commit-ref, pr-ref, qa-ref, review, file-verification
  - `meta-guidance/`: prompting, initiative-scoping, consult-ref (if consult-ref exists)

### 2. Command Categories (Workflow-Based)
- Are current 7 categories optimal?
- Should any be merged? (e.g., `meta/` into `navigation/`)
- Should any be split? (e.g., `team-switching/` is large at 10 commands)

### 3. Hook Categories
- Event-based: `session-events/`, `tool-events/`, `lifecycle-events/`
- Purpose-based: `tracking/`, `validation/`, `context-injection/`
- Hybrid approach?
- What about `lib/`? Keep at root or distribute?

### 4. Root-Level Exceptions
- Which resources should NOT be categorized?
- Candidates: `lib/` (hooks), `session-common` (skills), `standards` (skills)
- Pattern: `*-common` stays at root?

### 5. Cross-Pillar Alignment
- Should `session/` category exist in all three pillars?
- Should `operations/` category exist in skills AND commands?
- Or is independence cleaner?

## Expected Deliverables

1. **Category Design**
   - Final category list for each pillar
   - Resource-to-category mapping
   - Root-level exception list

2. **TDD (Technical Design Document)**
   - Sync script modifications
   - Manifest schema changes
   - Migration steps (directory structure)
   - Rollback plan (if needed)

3. **Success Criteria**
   - All resources properly categorized
   - Sync scripts work with new structure
   - Manifests track category metadata
   - Claude Code compatibility maintained

## Reference Materials

| Document | Path |
|----------|------|
| ADR-0006 (Draft) | `docs/decisions/ADR-0006-categorical-resource-organization.md` |
| ADR-0002 (Hooks lib/) | `docs/decisions/ADR-0002-hook-library-resolution-architecture.md` |
| sync-user-skills.sh | `sync-user-skills.sh` (lines 429-491 for discovery) |
| sync-user-commands.sh | `sync-user-commands.sh` (lines 552-619 for flatten pattern) |
| sync-user-hooks.sh | `sync-user-hooks.sh` |

## Exploration Summary

A 3-agent deep exploration (2025-12-31) analyzed:
1. **Skills structure**: Current flat, proposed categorical
2. **CEM infrastructure**: Correctly delegates to sync scripts
3. **Commands/hooks patterns**: Commands already implement flatten pattern

Key finding: **Commands already implement the pattern we want to standardize.**

## Session Invocation

```bash
/architect "Design categorical organization for roster user-level resources per ADR-0006"
```

Read this brief and ADR-0006 before beginning design work.
