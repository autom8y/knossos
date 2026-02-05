# SPIKE: Claude Code Skills/Commands Unification

**Date**: 2026-01-10
**Author**: Claude Code (Spike Session)
**Status**: Complete (Final)
**Decision Informs**: Knossos architecture evolution, deprecation strategy

---

## Executive Summary

**Decision**: Deprecate `user-skills/` entirely. Everything becomes a command.

- **Invokable commands** (`invokable: true`): User-callable via `/name`
- **Reference commands** (`invokable: false`): Library content, not user-callable
- **Single source**: `user-commands/` only
- **Progressive disclosure**: Preserved within command directories
- **`.claude/skills/`**: Deprecated, replaced by `.claude/commands/`

---

## Question

How should Knossos adapt to Claude Code's merging of slash commands and skills?

**Answer**: Complete unification. Skills are deprecated. Everything is a command.

## Context

Claude Code released a change described as "Merged slash commands and skills, simplifying the mental model with no change in behavior." While the immediate impact appears minimal, the long-term architectural direction suggests skills may eventually be deprecated in favor of a unified command model.

### Current Knossos Architecture

Knossos maintains a clear separation:

| Concept | Source Location | Projection Location | Purpose |
|---------|-----------------|---------------------|---------|
| **Commands** | `user-commands/` | `.claude/commands/` | User-invokable actions (`/name`) |
| **Skills** | `user-skills/` | `.claude/skills/` | Reference documentation + auto-invocation triggers |

**Current counts:**
- 35+ commands across 7 categories (navigation, meta, operations, workflow, session, rite-switching, cem)
- 90+ skill files across 6 categories (documentation, guidance, orchestration, session-lifecycle, operations)

### Frontmatter Differences

**Commands** (`user-commands/*.md`):
```yaml
---
description: Initialize a new work session
argument-hint: <initiative> [--complexity=LEVEL]
allowed-tools: Bash, Read, Task
model: opus
---
```

**Skills** (`user-skills/*/SKILL.md`):
```yaml
---
name: start-ref
description: "Begin a new work session... Triggers: /start, new session, begin work"
---
```

## Findings

### 1. Claude Code's Unified Model

The Skill tool's "Available skills" section now displays **both** commands and skills as a single list. Claude Code treats them identically at the invocation level:

```
Available skills:
- session-common: Shared session lifecycle schemas...
- 10x-ref: Quick switch to 10x-dev...
- sprint: Multi-task sprint orchestration...  # Was a command
- pr: Create pull request...                   # Was a command
```

**Key insight**: The distinction between "commands" and "skills" is now Knossos-specific, not Claude Code-mandated.

### 2. Invocation Mechanism Comparison

| Aspect | Commands (legacy) | Skills (legacy) | Claude Code (unified) |
|--------|-------------------|-----------------|----------------------|
| Manual trigger | `/name` | Skill tool with name | Both work via Skill tool |
| Auto-trigger | Never | Based on "Triggers:" in description | Based on "Triggers:" |
| Rich docs | References skill | Self-contained | Self-contained preferred |
| Arguments | `$ARGUMENTS` | N/A | `$ARGUMENTS` supported |
| Tool restrictions | `allowed-tools` | N/A | No restriction mechanism |
| Model selection | `model:` | N/A | No selection mechanism |

### 3. The Duplicated Work Problem

For `/start`, Knossos currently maintains:
1. `user-commands/session/start.md` - 119 lines, actionable prompt
2. `user-skills/session-lifecycle/start-ref/SKILL.md` - 131 lines, documentation
3. Supporting files in `start-ref/` - behavior.md, examples.md, integration.md

**Total: ~500 lines across 5+ files** for one command/skill pair.

This pattern repeats for: `/park`, `/wrap`, `/resume`, `/handoff`, `/sprint`, `/task`, `/pr`, `/spike`, `/hotfix`, `/commit`, `/architect`, `/build`, `/qa`, and rite-switching commands.

### 4. What Claude Code's Merger Actually Means

The "no change in behavior" is accurate for **invocation** but masks architectural direction:

1. **Discovery unified**: Both live in "Available skills" list
2. **Invocation unified**: Both use Skill tool mechanism
3. **Format NOT unified**: Frontmatter still differs (this is the gap)
4. **Auto-trigger exists**: "Triggers:" keyword in description enables auto-invocation

**Prediction**: Future Claude Code versions will:
- Deprecate `.claude/skills/` in favor of unified `.claude/commands/`
- Add `triggers:` frontmatter to commands
- Remove Skill tool in favor of native command invocation
- Surface auto-invocation via description parsing regardless of location

### 5. Progressive Disclosure Analysis

After examining actual skill content, progressive disclosure remains essential:

| Content | Total Lines | Files | Why Progressive Disclosure |
|---------|-------------|-------|---------------------------|
| `session-common` | 3,189 | 8 | Load one schema, not all 8 |
| `smell-detection` | 2,340 | 13 | Load one taxonomy, not all 7 |
| `agent-prompt-engineering` | 2,200 | 10+ | Load rubric OR template, not both |
| `/commit` ecosystem | 1,104 | 4 | Load command (120) OR examples (478) |

**Token economics**: Loading 3,189 lines when you need one 200-line schema is wasteful.

**Conclusion**: Progressive disclosure stays. But it moves INTO `user-commands/`.

---

## Recommendation: Complete Unification

### The Decision

**Deprecate `user-skills/` entirely. Everything becomes a command.**

A command with `invokable: false` is functionally what a skill was. But now:
- Single source of truth: `user-commands/`
- Single projection: `.claude/commands/`
- Single mental model: "Commands. Some you call, some you read."

### Unified Frontmatter Schema

```yaml
---
name: start                              # Required for all
description: Initialize a new work session

# For user-callable commands
invokable: true                          # Default: true
argument-hint: <initiative> [--complexity=LEVEL]
triggers: [new session, begin work]      # Auto-invocation keywords
allowed-tools: [Bash, Read, Task]
model: opus

# For reference/library commands
invokable: false                         # Not user-callable
category: reference                      # reference | template | schema
---
```

### Directory Structure After Migration

```
user-commands/                           # SINGLE SOURCE OF TRUTH
├── session/                             # Domain: session lifecycle
│   ├── start.md                        # invokable: true
│   ├── start/                          # Progressive disclosure
│   │   ├── behavior.md
│   │   └── examples.md
│   ├── park.md                         # invokable: true
│   ├── park/
│   ├── common/                         # invokable: false (was session-common)
│   │   ├── INDEX.md                   # Entry point
│   │   ├── session-context-schema.md
│   │   ├── session-state-machine.md
│   │   └── ...
│   └── ...
│
├── operations/                          # Domain: git/build operations
│   ├── commit.md                       # invokable: true
│   ├── commit/                         # Progressive disclosure
│   │   ├── behavior.md
│   │   └── examples.md
│   ├── pr.md
│   └── ...
│
├── guidance/                            # Domain: patterns & reference
│   ├── prompting/                      # invokable: false
│   │   ├── INDEX.md
│   │   ├── patterns/
│   │   └── workflows/
│   ├── cross-rite/                     # invokable: false
│   └── ...
│
├── templates/                           # Domain: document templates
│   ├── doc-artifacts/                  # invokable: false
│   │   ├── INDEX.md
│   │   ├── schemas/
│   │   └── ...
│   └── shared-templates/               # invokable: false
│
├── navigation/                          # Existing domain
├── workflow/                            # Existing domain
├── meta/                                # Existing domain
├── cem/                                 # Existing domain
└── rite-switching/                      # Existing domain

rites/                                   # Rite-specific content
├── shared/commands/                     # Was: shared/skills/
│   ├── smell-detection/                # invokable: false
│   └── cross-rite-handoff/
├── forge/commands/                      # Was: forge/skills/
│   └── agent-prompt-engineering/       # invokable: false
└── ...

.claude/commands/                        # SINGLE PROJECTION
                                        # .claude/skills/ DEPRECATED
```

### Migration Mapping

| Current Location | New Location | Notes |
|-----------------|--------------|-------|
| `user-commands/session/start.md` | `user-commands/session/start.md` | Unchanged |
| `user-skills/session-lifecycle/start-ref/` | `user-commands/session/start/` | Merged as progressive disclosure |
| `user-skills/session-lifecycle/session-common/` | `user-commands/session/common/` | Moved, `invokable: false` |
| `user-skills/guidance/prompting/` | `user-commands/guidance/prompting/` | Moved, `invokable: false` |
| `user-skills/documentation/doc-artifacts/` | `user-commands/templates/doc-artifacts/` | Moved, `invokable: false` |
| `rites/shared/skills/smell-detection/` | `rites/shared/commands/smell-detection/` | Renamed, `invokable: false` |
| `rites/forge/skills/agent-prompt-engineering/` | `rites/forge/commands/agent-prompt-engineering/` | Renamed, `invokable: false` |

### Command Classification

| Command | invokable | Category | Notes |
|---------|-----------|----------|-------|
| `/start` | true | - | User-callable |
| `/park` | true | - | User-callable |
| `/commit` | true | - | User-callable |
| `session/common` | false | reference | Schema definitions |
| `guidance/prompting` | false | reference | Invocation patterns |
| `templates/doc-artifacts` | false | template | PRD/TDD/ADR schemas |
| `smell-detection` | false | reference | Detection taxonomy |

### What Changes in Materialization

1. **Source directories**:
   - Read from `user-commands/` only (not `user-skills/`)
   - Read from `rites/*/commands/` only (not `rites/*/skills/`)

2. **Projection directory**:
   - Write to `.claude/commands/` only
   - `.claude/skills/` no longer generated

3. **Manifest schema**:
   - Rite manifests: `commands:` replaces `skills:`
   - Or: single `commands:` array with `invokable` flag

4. **Frontmatter parsing**:
   - Read `invokable` field (default: true)
   - Read `category` field for non-invokable

### Migration Execution

**This is greenfield. Clean break, not phased.**

1. **Create new structure** (1 session)
   - Create new domains: `guidance/`, `templates/`
   - Move `-ref` content into command progressive disclosure directories
   - Move library skills into appropriate domains

2. **Update rite manifests** (scripted)
   - Rename `skills/` → `commands/` in all rites
   - Update manifest.yaml `skills:` → `commands:`

3. **Update materialization** (1 session)
   - Change source paths
   - Change projection paths
   - Parse new frontmatter fields

4. **Delete legacy** (immediate)
   - `rm -rf user-skills/`
   - `rm -rf rites/*/skills/`
   - Remove `.claude/skills/` generation

## Follow-up Actions

1. **Execute migration**: Single sprint, clean break
2. **Update manifest schema**: Document new `commands:` format
3. **Update materialization**: `internal/materialize/materialize.go`
4. **Update inscription**: Templates reference new paths
5. **Create ADR**: Document deprecation decision

## Appendix: Full Migration Inventory

### Commands to Keep (invokable: true)
- **session/**: start, park, wrap, resume, handoff, continue
- **operations/**: commit, pr, spike, architect, build, qa, code-review
- **workflow/**: task, sprint, hotfix
- **navigation/**: worktree, sessions, consult, rite, ecosystem
- **meta/**: minus-1, zero, one
- **rite-switching/**: 10x, hygiene, debt, forge, docs, intelligence, security, strategy, sre, rnd
- **cem/**: sync

### Skills to Convert (invokable: false)

**Merge into command progressive disclosure:**
- `start-ref/` → `session/start/`
- `park-ref/` → `session/park/`
- `wrap-ref/` → `session/wrap/`
- `commit-ref/` → `operations/commit/`
- `pr-ref/` → `operations/pr/`
- `spike-ref/` → `operations/spike/`
- `hotfix-ref/` → `workflow/hotfix/`
- `consult-ref/` → `navigation/consult/`

**Move to new domains:**
- `session-common/` → `session/common/`
- `shared-sections/` → `session/shared/`
- `prompting/` → `guidance/prompting/`
- `cross-rite/` → `guidance/cross-rite/`
- `file-verification/` → `guidance/file-verification/`
- `doc-artifacts/` → `templates/doc-artifacts/`
- `justfile/` → `templates/justfile/`
- `atuin-desktop/` → `templates/atuin-desktop/`
- `standards/` → `guidance/standards/`

**Rite-specific (rename directory):**
- `rites/shared/skills/` → `rites/shared/commands/`
- `rites/forge/skills/` → `rites/forge/commands/`
- `rites/docs/skills/` → `rites/docs/commands/`
- etc.
