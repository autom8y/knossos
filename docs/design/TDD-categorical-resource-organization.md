# TDD: Categorical Resource Organization

| Field | Value |
|-------|-------|
| **Initiative** | Resource Organization Standardization |
| **ADR Reference** | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0006-categorical-resource-organization.md` |
| **Author** | Architect |
| **Status** | Draft |
| **Date** | 2025-12-31 |

## 1. Overview

This document specifies the technical design for standardizing categorical organization across all roster user-level resources (skills, commands, hooks). The design implements a "categorical source to flat destination" pattern that improves developer navigation while maintaining Claude Code compatibility.

### 1.1 Design Principles

- **Organize for developers**: Source structure uses logical categories for human navigation
- **Flatten for runtime**: Destination structure remains flat for Claude Code consumption
- **Track provenance**: Manifests store category metadata for debugging and status reporting
- **Independent categories**: Each pillar defines categories appropriate to its domain
- **Minimal exceptions**: Root-level resources are rare and well-justified

### 1.2 Key Constraints

| Constraint | Source | Impact |
|------------|--------|--------|
| Big bang migration | User decision | No backward compatibility period |
| kebab-case naming | User decision | All category names use kebab-case |
| lib/ at root | User decision | Hook libraries remain uncategorized |
| SKILL.md only required | User decision | Minimal skill structure requirements |

### 1.3 Scope

**In Scope:**
- Skill category definitions and mapping (24 skills)
- Command category validation (31 commands, 7 existing categories)
- Hook category definitions and mapping (10 hooks)
- Root-level exception patterns
- Sync script modifications
- Manifest schema updates

**Out of Scope:**
- Agents (remain flat per existing pattern)
- Team-level resources (only user-level affected)
- Claude Code modifications (destination remains flat)

## 2. Category Definitions

### 2.1 Skill Categories (Domain-Based)

Skills are grouped by functional domain. Five categories capture all 24 skills with one root-level exception.

| Category | Description | Skills | Count |
|----------|-------------|--------|-------|
| `session-lifecycle` | Session state management from start to wrap | start-ref, park-ref, resume, handoff-ref, wrap-ref | 5 |
| `orchestration` | Multi-phase workflow coordination and execution | orchestration, orchestrator-templates, sprint-ref, task-ref, initiative-scoping | 5 |
| `operations` | Code shipping activities (commit, PR, review, hotfix) | commit-ref, pr-ref, qa-ref, review, hotfix-ref, spike-ref, worktree-ref | 7 |
| `documentation` | Templates, standards, and conventions | documentation, doc-artifacts, standards | 3 |
| `guidance` | Meta-skills for navigation and quality | prompting, cross-team, file-verification | 3 |
| **Root exception** | Shared reference module | session-common | 1 |

**Total: 24 skills**

#### 2.1.1 Category Rationale

**session-lifecycle**: Groups all skills that manage session state transitions. These skills share the `session-common` dependency and operate on SESSION_CONTEXT.md.

**orchestration**: Groups skills that coordinate multi-step or multi-agent work. Distinguishes from session-lifecycle (state) vs orchestration (workflow execution).

**operations**: Groups skills used during active development work. Includes code review, commits, PRs, QA, and the exploration/hotfix escape hatches.

**documentation**: Groups template and standards skills. These are reference materials consulted during work rather than invoked as workflows.

**guidance**: Groups meta-skills that help navigate the ecosystem rather than execute specific workflows.

#### 2.1.2 Root Exception: session-common

`session-common` is explicitly excluded from categorization because:

1. **Reference module, not invocable**: It contains shared schemas, not executable skill logic
2. **Cross-cutting dependency**: All session-lifecycle skills reference it
3. **Pattern precedent**: Similar to how `lib/` remains at root for hooks
4. **Discoverability**: Root placement signals "shared infrastructure" status

### 2.2 Command Categories (Workflow-Based)

Commands use existing 7 categories, validated and optimized.

| Category | Description | Commands | Count | Status |
|----------|-------------|----------|-------|--------|
| `session` | Session lifecycle commands | start, park, continue, handoff, wrap | 5 | Keep |
| `workflow` | Multi-step workflow initiators | task, sprint, hotfix | 3 | Keep |
| `operations` | Individual operation commands | architect, build, qa, code-review, commit | 5 | Keep |
| `navigation` | Ecosystem navigation and discovery | consult, team, worktree, sessions, ecosystem | 5 | Keep |
| `meta` | Initialization and scoping commands | minus-1, zero, one | 3 | Keep |
| `rite-switching` | Team pack activation shortcuts | 10x, docs, hygiene, debt, sre, security, intelligence, rnd, strategy, forge | 10 | Keep |
| `cem` | Claude Ecosystem Management tooling | sync | 1 | Keep |

**Total: 32 commands (31 + sync)**

#### 2.2.1 Validation Analysis

The existing command categories are well-designed:

- **Clear separation of concerns**: Each category serves a distinct purpose
- **Balanced distribution**: Categories range from 1-10 commands (acceptable variance)
- **Logical groupings**: Related commands are co-located
- **No orphans**: Every command has a natural category home

**Recommendation: No changes needed.** The current structure serves as the reference implementation.

### 2.3 Hook Categories (Event/Purpose-Based)

Hooks are grouped by their primary function in the Claude Code lifecycle.

| Category | Description | Hooks | Count |
|----------|-------------|-------|-------|
| `context-injection` | SessionStart hooks that inject context | session-context, coach-mode | 2 |
| `session-guards` | Session state protection and automation | auto-park, session-write-guard, start-preflight | 3 |
| `validation` | PreToolUse validators | command-validator, delegation-check | 2 |
| `tracking` | PostToolUse trackers and auditors | artifact-tracker, commit-tracker, session-audit | 3 |
| **Root exception** | Shared hook libraries | lib/ (10 files) | N/A |

**Total: 10 hooks + lib/**

#### 2.3.1 Category Rationale

**context-injection**: Hooks that run at SessionStart and inject context into Claude's understanding. They are "read-only" from the perspective of session state.

**session-guards**: Hooks that protect session integrity. `auto-park` ensures sessions don't orphan, `session-write-guard` prevents direct writes to context files, `start-preflight` validates before session start.

**validation**: PreToolUse hooks that validate commands before execution. Focus on authorization and correctness checks.

**tracking**: PostToolUse hooks that record what happened. Focus on logging, auditing, and artifact tracking.

#### 2.3.2 Root Exception: lib/

The `lib/` directory remains at root level because:

1. **ADR-0002 established pattern**: Hook library resolution depends on predictable path
2. **Shared infrastructure**: Used by all hooks regardless of category
3. **Not a hook itself**: Contains libraries, not hook scripts
4. **User decision**: Explicitly confirmed as root-level

### 2.4 Root-Level Exception Pattern

Resources may remain at root level when they meet ALL of these criteria:

1. **Infrastructure role**: Contains shared code/schemas, not executable resources
2. **Cross-cutting usage**: Referenced by multiple categories
3. **Not directly invoked**: Users do not invoke them as commands/skills
4. **Established precedent**: Pattern already exists in codebase

| Pillar | Root Exceptions | Justification |
|--------|-----------------|---------------|
| Skills | `session-common/` | Shared schemas for session-lifecycle |
| Commands | None | All commands are invocable |
| Hooks | `lib/` | Shared libraries per ADR-0002 |

## 3. Directory Structure

### 3.1 Skills (Target State)

```
user-skills/
  session-common/                    # ROOT EXCEPTION: shared schemas
    SKILL.md
    session-context-schema.md
    session-phases.md
    ...
  session-lifecycle/
    start-ref/
      SKILL.md
    park-ref/
      SKILL.md
    resume/
      SKILL.md
    handoff-ref/
      SKILL.md
    wrap-ref/
      SKILL.md
  orchestration/
    orchestration/
      SKILL.md
      main-thread-guide.md
      ...
    orchestrator-templates/
      SKILL.md
    sprint-ref/
      SKILL.md
    task-ref/
      SKILL.md
    initiative-scoping/
      SKILL.md
  operations/
    commit-ref/
      SKILL.md
    pr-ref/
      SKILL.md
    qa-ref/
      SKILL.md
    review/
      SKILL.md
    hotfix-ref/
      SKILL.md
    spike-ref/
      SKILL.md
    worktree-ref/
      SKILL.md
  documentation/
    documentation/
      SKILL.md
    doc-artifacts/
      SKILL.md
      schemas/
    standards/
      SKILL.md
      code-conventions.md
      ...
  guidance/
    prompting/
      SKILL.md
      patterns/
      workflows/
    cross-team/
      SKILL.md
    file-verification/
      SKILL.md
```

### 3.2 Commands (Current State - Validated)

```
user-commands/
  session/
    start.md
    park.md
    continue.md
    handoff.md
    wrap.md
  workflow/
    task.md
    sprint.md
    hotfix.md
  operations/
    architect.md
    build.md
    qa.md
    code-review.md
    commit.md
  navigation/
    consult.md
    team.md
    worktree.md
    sessions.md
    ecosystem.md
  meta/
    minus-1.md
    zero.md
    one.md
  rite-switching/
    10x.md
    docs.md
    hygiene.md
    debt.md
    sre.md
    security.md
    intelligence.md
    rnd.md
    strategy.md
    forge.md
  cem/
    sync.md
```

### 3.3 Hooks (Target State)

```
user-hooks/
  lib/                              # ROOT EXCEPTION: shared libraries
    config.sh
    logging.sh
    primitives.sh
    session-core.sh
    session-state.sh
    session-fsm.sh
    session-manager.sh
    session-migrate.sh
    session-utils.sh
    worktree-manager.sh
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

## 4. Sync Script Modifications

### 4.1 sync-user-skills.sh Changes

The skills sync script needs to implement the flatten pattern already used by commands.

```bash
# Current behavior: direct rsync
rsync -av --delete "$SOURCE_DIR/" "$DEST_DIR/"

# Target behavior: categorical discovery with flatten
sync_skills() {
    local SOURCE_DIR="$ROSTER_HOME/user-skills"
    local DEST_DIR="$HOME/.claude/skills"

    # Process root-level exceptions first
    for skill_dir in "$SOURCE_DIR"/session-common/; do
        [[ -d "$skill_dir" ]] || continue
        local skill_name=$(basename "$skill_dir")
        sync_skill "$skill_dir" "$DEST_DIR/$skill_name" "root"
    done

    # Process categorized skills
    for category_dir in "$SOURCE_DIR"/*/; do
        [[ -d "$category_dir" ]] || continue
        local category=$(basename "$category_dir")

        # Skip root exceptions
        [[ "$category" == "session-common" ]] && continue

        for skill_dir in "$category_dir"/*/; do
            [[ -d "$skill_dir" ]] || continue
            local skill_name=$(basename "$skill_dir")
            sync_skill "$skill_dir" "$DEST_DIR/$skill_name" "$category"
        done
    done
}

sync_skill() {
    local source="$1"
    local dest="$2"
    local category="$3"

    # Validate SKILL.md exists
    if [[ ! -f "$source/SKILL.md" ]]; then
        log_warning "Skipping $source: missing SKILL.md"
        return
    fi

    # Copy entire skill directory (rsync with delete)
    rsync -av --delete "$source/" "$dest/"

    # Track in manifest with category
    add_to_manifest "skills" "$(basename "$dest")" "$category"
}
```

### 4.2 sync-user-hooks.sh Changes

The hooks sync script needs to handle categorization while preserving lib/ at root.

```bash
sync_hooks() {
    local SOURCE_DIR="$ROSTER_HOME/user-hooks"
    local DEST_DIR="$HOME/.claude/hooks"

    # Sync lib/ directly (root exception)
    if [[ -d "$SOURCE_DIR/lib" ]]; then
        rsync -av --delete "$SOURCE_DIR/lib/" "$DEST_DIR/lib/"
    fi

    # Process categorized hooks
    for category_dir in "$SOURCE_DIR"/*/; do
        [[ -d "$category_dir" ]] || continue
        local category=$(basename "$category_dir")

        # Skip lib/ (already handled)
        [[ "$category" == "lib" ]] && continue

        for hook_file in "$category_dir"/*.sh; do
            [[ -f "$hook_file" ]] || continue
            local hook_name=$(basename "$hook_file")

            # Flatten to destination root
            cp "$hook_file" "$DEST_DIR/$hook_name"
            add_to_manifest "hooks" "$hook_name" "$category"
        done
    done
}
```

### 4.3 sync-user-commands.sh (Reference)

No changes needed. Current implementation is the reference pattern.

## 5. Manifest Schema Updates

### 5.1 Current Manifest Structure

```json
{
  "version": "1.0.0",
  "last_sync": "2025-12-31T12:00:00Z",
  "commands": {
    "start.md": {
      "source": "roster",
      "category": "session",
      "checksum": "abc123..."
    }
  }
}
```

### 5.2 Extended Manifest Structure

```json
{
  "version": "1.1.0",
  "last_sync": "2025-12-31T12:00:00Z",
  "commands": {
    "start.md": {
      "source": "roster",
      "category": "session",
      "checksum": "abc123..."
    }
  },
  "skills": {
    "start-ref": {
      "source": "roster",
      "category": "session-lifecycle",
      "checksum": "def456..."
    },
    "session-common": {
      "source": "roster",
      "category": "root",
      "checksum": "ghi789..."
    }
  },
  "hooks": {
    "session-context.sh": {
      "source": "roster",
      "category": "context-injection",
      "checksum": "jkl012..."
    },
    "lib/": {
      "source": "roster",
      "category": "root",
      "checksum": "mno345..."
    }
  }
}
```

### 5.3 Category Values

| Pillar | Valid Categories |
|--------|------------------|
| Commands | session, workflow, operations, navigation, meta, rite-switching, cem |
| Skills | session-lifecycle, orchestration, operations, documentation, guidance, root |
| Hooks | context-injection, session-guards, validation, tracking, root |

The special `root` category indicates a root-level exception.

## 6. Migration Steps

### 6.1 Phase 1: Create Category Directories

```bash
# Skills
mkdir -p user-skills/{session-lifecycle,orchestration,operations,documentation,guidance}

# Hooks
mkdir -p user-hooks/{context-injection,session-guards,validation,tracking}

# Commands - already categorized
```

### 6.2 Phase 2: Move Skills

```bash
# session-lifecycle (5 skills)
mv user-skills/start-ref user-skills/session-lifecycle/
mv user-skills/park-ref user-skills/session-lifecycle/
mv user-skills/resume user-skills/session-lifecycle/
mv user-skills/handoff-ref user-skills/session-lifecycle/
mv user-skills/wrap-ref user-skills/session-lifecycle/

# orchestration (5 skills)
mv user-skills/orchestration user-skills/orchestration/orchestration
mv user-skills/orchestrator-templates user-skills/orchestration/
mv user-skills/sprint-ref user-skills/orchestration/
mv user-skills/task-ref user-skills/orchestration/
mv user-skills/initiative-scoping user-skills/orchestration/

# operations (7 skills)
mv user-skills/commit-ref user-skills/operations/
mv user-skills/pr-ref user-skills/operations/
mv user-skills/qa-ref user-skills/operations/
mv user-skills/review user-skills/operations/
mv user-skills/hotfix-ref user-skills/operations/
mv user-skills/spike-ref user-skills/operations/
mv user-skills/worktree-ref user-skills/operations/

# documentation (3 skills)
mv user-skills/documentation user-skills/documentation/documentation
mv user-skills/doc-artifacts user-skills/documentation/
mv user-skills/standards user-skills/documentation/

# guidance (3 skills)
mv user-skills/prompting user-skills/guidance/
mv user-skills/cross-team user-skills/guidance/
mv user-skills/file-verification user-skills/guidance/

# session-common stays at root (1 skill)
```

### 6.3 Phase 3: Move Hooks

```bash
# context-injection (2 hooks)
mv user-hooks/session-context.sh user-hooks/context-injection/
mv user-hooks/coach-mode.sh user-hooks/context-injection/

# session-guards (3 hooks)
mv user-hooks/auto-park.sh user-hooks/session-guards/
mv user-hooks/session-write-guard.sh user-hooks/session-guards/
mv user-hooks/start-preflight.sh user-hooks/session-guards/

# validation (2 hooks)
mv user-hooks/command-validator.sh user-hooks/validation/
mv user-hooks/delegation-check.sh user-hooks/validation/

# tracking (3 hooks)
mv user-hooks/artifact-tracker.sh user-hooks/tracking/
mv user-hooks/commit-tracker.sh user-hooks/tracking/
mv user-hooks/session-audit.sh user-hooks/tracking/

# lib/ stays at root
```

### 6.4 Phase 4: Update Sync Scripts

1. Modify `sync-user-skills.sh` per Section 4.1
2. Modify `sync-user-hooks.sh` per Section 4.2
3. Update manifest version to 1.1.0
4. Add category field to skill/hook entries

### 6.5 Phase 5: Validation

```bash
# Run sync
./sync-user-skills.sh
./sync-user-hooks.sh
./sync-user-commands.sh

# Verify flat destination
ls ~/.claude/skills/   # Should be flat
ls ~/.claude/hooks/    # Should be flat + lib/
ls ~/.claude/commands/ # Should be flat

# Verify manifest
cat ~/.claude/.cem/manifest.json | jq '.skills, .hooks'

# Test Claude Code
# - Skills should activate correctly
# - Hooks should fire correctly
# - Commands should execute correctly
```

## 7. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Skill path breaks in SKILL.md references | Medium | High | Update relative paths to peer skills |
| Hook library resolution fails | Low | High | lib/ stays at root, paths unchanged |
| Sync script bugs | Medium | Medium | Thorough testing before migration |
| Manifest version incompatibility | Low | Medium | Version bump with clear migration |
| Category naming drift | Low | Low | Document categories in ADR |

### 7.1 Path Reference Updates

Skills with internal relative paths need updates:

| Skill | Reference | Change Needed |
|-------|-----------|---------------|
| prompting | `../orchestration/main-thread-guide.md` | `../../orchestration/orchestration/main-thread-guide.md` |
| prompting | `../standards/SKILL.md` | `../../documentation/standards/SKILL.md` |
| standards | `../justfile/SKILL.md` | Update based on justfile location |

These must be updated as part of the migration.

## 8. ADRs

### 8.1 Related ADRs

- **ADR-0006**: Categorical Resource Organization Pattern (this design implements)
- **ADR-0002**: Hook Library Resolution Architecture (establishes lib/ pattern)

### 8.2 Decisions Made in This Design

| Decision | Choice | Rationale |
|----------|--------|-----------|
| 5 skill categories | session-lifecycle, orchestration, operations, documentation, guidance | Domain-based grouping matches mental model |
| session-common at root | Root exception | Reference module, not invocable skill |
| 4 hook categories | context-injection, session-guards, validation, tracking | Purpose-based grouping matches hook lifecycle |
| lib/ at root | Root exception | Per ADR-0002 and user decision |
| Keep command categories | No changes | Existing structure is optimal |
| Manifest category field | "root" for exceptions | Explicit tracking of exception status |

## 9. Open Items

| Item | Owner | Status |
|------|-------|--------|
| Validate prompting path references | Principal Engineer | Pending |
| justfile skill location decision | Architect | Pending (if exists) |
| Sync script implementation | Principal Engineer | Pending |
| Migration script automation | Principal Engineer | Pending |

## 10. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-categorical-resource-organization.md` | Pending |
| ADR-0006 Update | `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0006-categorical-resource-organization.md` | Pending |

All artifacts to be verified via Read tool after completion.
