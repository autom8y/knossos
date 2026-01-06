# ADR-0006: Categorical Resource Organization Pattern

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2025-12-31 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A (foundational) |
| **Superseded by** | N/A |
| **TDD Reference** | `/Users/tomtenuta/Code/roster/docs/design/TDD-categorical-resource-organization.md` |

## Context

The roster ecosystem manages four types of user-level resources that sync to Claude Code:
- **Skills** (`user-skills/` -> `~/.claude/skills/`)
- **Commands** (`user-commands/` -> `~/.claude/commands/`)
- **Hooks** (`user-hooks/` -> `~/.claude/hooks/`)
- **Agents** (`user-agents/` -> `~/.claude/agents/`)

### Current State

| Resource | Source Structure | Sync Behavior | Category Tracking |
|----------|------------------|---------------|-------------------|
| Commands | Categorical (`session/`, `workflow/`, etc.) | Flattens to destination | Yes (in manifest) |
| Hooks | Flat + `lib/` subdirectory | Preserves `lib/` structure | Partial |
| Skills | Flat directories | Direct copy (rsync) | No |
| Agents | Flat files | Direct copy | No |

### Problem Statement

1. **Inconsistent organization**: Commands use categorical organization; skills/hooks/agents are flat
2. **Developer experience**: Flat structures make navigation harder as resource counts grow
3. **Discoverability**: Related resources not grouped together
4. **Manifest gaps**: Skills and agents don't track category metadata

### Discovery: Commands Already Implement the Pattern

The `sync-user-commands.sh` script (lines 552-619) implements a "categorical source -> flat destination" pattern:

```bash
# Process each command in source subdirectories (flatten structure)
for category_dir in "$SOURCE_DIR"/*/; do
    [[ -d "$category_dir" ]] || continue
    local category
    category=$(basename "$category_dir")
    for source_file in "$category_dir"/*.md; do
        # Copies to flat ~/.claude/commands/
    done
done
```

The manifest preserves category as metadata:
```json
{
  "commands": {
    "start.md": {
      "source": "roster",
      "category": "session",
      "checksum": "..."
    }
  }
}
```

## Decision

**Standardize categorical organization across all resource pillars while maintaining flat sync output for Claude Code compatibility.**

### Core Principles

1. **Organize for developers**: Source structure uses logical categories for human navigation
2. **Flatten for runtime**: Destination structure remains flat for Claude Code consumption
3. **Track provenance**: Manifests store category metadata for debugging/status
4. **Independent categories**: Each pillar defines categories appropriate to its domain
5. **Minimal root exceptions**: Only shared infrastructure remains at root level

### Migration Approach

- **Big bang swap**: No backward compatibility period
- **Naming convention**: kebab-case for all category names
- **Skill internals**: Minimal requirements (SKILL.md only)

## Category Definitions

### Skill Categories (Domain-Based)

| Category | Description | Skills |
|----------|-------------|--------|
| `session-lifecycle` | Session state management from start to wrap | start-ref, park-ref, resume, handoff-ref, wrap-ref |
| `orchestration` | Multi-phase workflow coordination and execution | orchestration, orchestrator-templates, sprint-ref, task-ref, initiative-scoping |
| `operations` | Code shipping activities | commit-ref, pr-ref, qa-ref, review, hotfix-ref, spike-ref, worktree-ref |
| `documentation` | Templates, standards, and conventions | documentation, doc-artifacts, standards |
| `guidance` | Meta-skills for navigation and quality | prompting, cross-team, file-verification |
| **Root exception** | Shared reference module | session-common |

### Command Categories (Workflow-Based) - Validated

| Category | Commands | Status |
|----------|----------|--------|
| `session` | start, park, continue, handoff, wrap | Keep unchanged |
| `workflow` | task, sprint, hotfix | Keep unchanged |
| `operations` | architect, build, qa, code-review, commit | Keep unchanged |
| `navigation` | consult, team, worktree, sessions, ecosystem | Keep unchanged |
| `meta` | minus-1, zero, one | Keep unchanged |
| `rite-switching` | 10x, docs, hygiene, debt, sre, security, intelligence, rnd, strategy, forge | Keep unchanged |
| `cem` | sync | Keep unchanged |

### Hook Categories (Event/Purpose-Based)

| Category | Description | Hooks |
|----------|-------------|-------|
| `context-injection` | SessionStart hooks that inject context | session-context, coach-mode |
| `session-guards` | Session state protection and automation | auto-park, session-write-guard, start-preflight |
| `validation` | PreToolUse validators | command-validator, delegation-check |
| `tracking` | PostToolUse trackers and auditors | artifact-tracker, commit-tracker, session-audit |
| **Root exception** | Shared hook libraries | lib/ |

### Root-Level Exception Pattern

Resources may remain at root when they meet ALL criteria:

1. **Infrastructure role**: Contains shared code/schemas, not executable resources
2. **Cross-cutting usage**: Referenced by multiple categories
3. **Not directly invoked**: Users do not invoke them as commands/skills
4. **Established precedent**: Pattern already exists in codebase

| Pillar | Root Exception | Justification |
|--------|----------------|---------------|
| Skills | `session-common/` | Shared schemas for session-lifecycle skills |
| Commands | None | All commands are invocable |
| Hooks | `lib/` | Shared libraries per ADR-0002 |

## Consequences

### Positive

1. **Improved discoverability**: Related resources grouped together
2. **Scalable organization**: Categories prevent flat directory sprawl
3. **Consistent pattern**: Same approach across all pillars
4. **Metadata preservation**: Category tracked in manifest for debugging
5. **No Claude Code changes**: Flat destination maintains compatibility

### Negative

1. **Sync script updates**: All sync scripts need category-aware discovery
2. **Migration effort**: Existing resources need reorganization
3. **Path changes**: Internal skill references may need updates

### Neutral

1. **Manifest schema changes**: Adding `category` field to skills/hooks manifests
2. **Documentation updates**: Explaining new structure to users

## Implementation Plan

### Phase 1: Prerequisites - COMPLETE
- [x] Fix case sensitivity (skill.md -> SKILL.md)
- [x] Clean up prototype artifacts
- [x] Draft ADR outline
- [x] Create architect brief

### Phase 2: Category Design - COMPLETE
- [x] Define skill categories (5 domain-based + 1 root exception)
- [x] Validate command categories (7 categories, no changes needed)
- [x] Define hook categories (4 purpose-based + 1 root exception)
- [x] Identify root-level exceptions (session-common, lib/)
- [x] Produce migration TDD

### Phase 3: Implementation
- [ ] Create category directories in source
- [ ] Move skills to category directories
- [ ] Move hooks to category directories
- [ ] Update `sync-user-skills.sh` for categorical discovery
- [ ] Update `sync-user-hooks.sh` for categorical discovery
- [ ] Update manifest schema (version 1.1.0)
- [ ] Update internal skill path references

### Phase 4: Validation
- [ ] Test sync scripts
- [ ] Validate manifest generation
- [ ] Verify Claude Code compatibility
- [ ] Verify skill/hook activation

## Target Directory Structure

### Skills
```
user-skills/
  session-common/              # ROOT EXCEPTION
  session-lifecycle/
    start-ref/
    park-ref/
    resume/
    handoff-ref/
    wrap-ref/
  orchestration/
    orchestration/
    orchestrator-templates/
    sprint-ref/
    task-ref/
    initiative-scoping/
  operations/
    commit-ref/
    pr-ref/
    qa-ref/
    review/
    hotfix-ref/
    spike-ref/
    worktree-ref/
  documentation/
    documentation/
    doc-artifacts/
    standards/
  guidance/
    prompting/
    cross-team/
    file-verification/
```

### Hooks
```
user-hooks/
  lib/                        # ROOT EXCEPTION
  context-injection/
    session-context.sh
    coach-mode.sh
  session-guards/
    auto-park.sh
    session-write-guard.sh
    start-preflight.sh
  validation/
    command-validator.sh
    delegation-check.sh
  tracking/
    artifact-tracker.sh
    commit-tracker.sh
    session-audit.sh
```

## Manifest Schema

Version 1.1.0 adds skills and hooks sections with category tracking:

```json
{
  "version": "1.1.0",
  "last_sync": "2025-12-31T12:00:00Z",
  "commands": { ... },
  "skills": {
    "start-ref": {
      "source": "roster",
      "category": "session-lifecycle",
      "checksum": "..."
    },
    "session-common": {
      "source": "roster",
      "category": "root",
      "checksum": "..."
    }
  },
  "hooks": {
    "session-context.sh": {
      "source": "roster",
      "category": "context-injection",
      "checksum": "..."
    },
    "lib/": {
      "source": "roster",
      "category": "root",
      "checksum": "..."
    }
  }
}
```

## References

- **TDD**: `/Users/tomtenuta/Code/roster/docs/design/TDD-categorical-resource-organization.md`
- **ADR-0002**: Hook Library Resolution Architecture (establishes `lib/` pattern)
- **sync-user-commands.sh**: Reference implementation of flatten pattern

## Appendix: Complete Resource Inventory

### Skills (24 total, 5 categories + 1 root)

| Category | Skills | Count |
|----------|--------|-------|
| session-lifecycle | start-ref, park-ref, resume, handoff-ref, wrap-ref | 5 |
| orchestration | orchestration, orchestrator-templates, sprint-ref, task-ref, initiative-scoping | 5 |
| operations | commit-ref, pr-ref, qa-ref, review, hotfix-ref, spike-ref, worktree-ref | 7 |
| documentation | documentation, doc-artifacts, standards | 3 |
| guidance | prompting, cross-team, file-verification | 3 |
| root | session-common | 1 |

### Commands (32 total, 7 categories)

| Category | Commands | Count |
|----------|----------|-------|
| session | start, park, continue, handoff, wrap | 5 |
| workflow | task, sprint, hotfix | 3 |
| operations | architect, build, qa, code-review, commit | 5 |
| navigation | consult, team, worktree, sessions, ecosystem | 5 |
| meta | minus-1, zero, one | 3 |
| rite-switching | 10x, docs, hygiene, debt, sre, security, intelligence, rnd, strategy, forge | 10 |
| cem | sync | 1 |

### Hooks (10 hooks + lib/, 4 categories + 1 root)

| Category | Hooks | Count |
|----------|-------|-------|
| context-injection | session-context, coach-mode | 2 |
| session-guards | auto-park, session-write-guard, start-preflight | 3 |
| validation | command-validator, delegation-check | 2 |
| tracking | artifact-tracker, commit-tracker, session-audit | 3 |
| root | lib/ | N/A |
