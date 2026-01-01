# TDD: Orchestrator Skill Extraction

| Field | Value |
|-------|-------|
| **Initiative** | Multi-Sprint Remediation (Sprint 5) |
| **Sprint Task** | task-001, task-002, task-003 |
| **Author** | Context Architect |
| **Status** | Draft |
| **Date** | 2026-01-01 |

## 1. Overview

This document specifies the technical design for extracting shared orchestrator content (~74% of current orchestrator.md files) into a referenceable skill, while maintaining team-specific customization and preserving the existing YAML-to-MD generation pipeline.

### 1.1 Problem Statement

Current state:
- 11 team packs each have their own `orchestrator.md` (~193 lines each)
- Analysis shows ~142 lines (74%) are identical across all teams
- Only ~51 lines (26%) are genuinely team-specific
- The `orchestrator-templates` skill documents generation but does not provide shared content for runtime reference
- Template updates require regenerating all 11 orchestrators

### 1.2 Design Principles

- **Single source of truth**: Shared protocol content lives in one location
- **Minimal orchestrator.md**: Team files contain only team-specific configuration
- **Generation compatibility**: Existing `orchestrator-generate.sh` continues to work
- **Progressive disclosure**: Core protocol loads on demand via skill reference
- **No runtime breakage**: Existing orchestrators remain functional during migration

### 1.3 Key Constraints

| Constraint | Source | Impact |
|------------|--------|--------|
| YAML-to-MD pipeline must work | Existing infrastructure | Skill must integrate with template substitution |
| @ reference syntax required | Claude Code skill loading | Skill must be named for @ reference |
| No duplicate content | DRY principle | Template should not embed what skill provides |
| Backward compatible | Migration path | Existing orchestrators must work during transition |

### 1.4 Scope

**In Scope:**
- Skill file structure for shared orchestrator content
- Template modification to reference skill instead of embedding
- Frontmatter standardization schema (task-002)
- Pattern propagation approach (task-003)
- Migration path for existing 11 orchestrators

**Out of Scope:**
- Changes to workflow.yaml structure
- Changes to orchestrator.yaml schema (beyond frontmatter)
- Modifications to consultation-protocol.md (already extracted)

## 2. Content Classification

### 2.1 Shared Content (Extract to Skill)

Lines that are identical across all orchestrators and represent core protocol:

| Section | Lines | Content |
|---------|-------|---------|
| Consultation Role (CRITICAL) | ~30 | What You DO/DO NOT, Litmus Test |
| Tool Access | ~15 | Read-only access, tool restrictions |
| Consultation Protocol | ~25 | Input/Output schemas (references consultation-protocol.md) |
| Core Responsibilities | ~6 | Phase Decomposition, Specialist Routing, etc. |
| Domain Authority (generic) | ~12 | Phase sequencing, escalation rules |
| Behavioral Constraints (DO NOT) | ~18 | Six constraint patterns |
| Handling Failures | ~10 | Recovery protocol |
| The Acid Test | ~6 | State clarity check |
| Anti-Patterns (generic) | ~8 | Doing work, prose responses, etc. |

**Total shared: ~130 lines** (67% of orchestrator.md)

### 2.2 Team-Specific Content (Remain in orchestrator.md)

Lines that vary per team and must remain in the generated file:

| Section | Lines | Content |
|---------|-------|---------|
| Frontmatter | ~8 | name, description, tools, model, color |
| Introduction paragraph | ~3 | Team name substitution |
| Position in Workflow | ~15 | ASCII diagram, upstream/downstream |
| Phase Routing Table | ~6 | Specialist-to-condition mapping |
| Handoff Criteria Table | ~12 | Phase-specific criteria |
| Cross-Team Protocol | ~6 | Team-specific escalation paths |
| Skills Reference | ~4 | Team-specific skill list |
| Team-Specific Anti-Patterns | ~4 | Specialty warnings |

**Total team-specific: ~58 lines** (30% of orchestrator.md)

### 2.3 Hybrid Content (Template with Placeholders)

Some sections are mostly shared but contain team-specific placeholders:

| Section | Shared | Team-Specific |
|---------|--------|---------------|
| Domain Authority | Generic rules | (none, fully shared) |
| Consultation Protocol | Schema structure | Complexity enum, Specialist enum |

## 3. Solution Architecture

### 3.1 Design Decision: Inline Reference vs. Link Reference

**Option A: Inline Reference (Rejected)**
```markdown
# Orchestrator
@orchestrator-core  <!-- Skill content inlined here -->

## Position in Workflow
```

Rejected because: Claude Code skill loading is not inline expansion; @ references are for skill invocation, not content inclusion.

**Option B: Link Reference with Minimal Duplication (Selected)**
```markdown
# Orchestrator
For core protocol, see: @orchestrator-core

## Position in Workflow
```

Selected because: Matches how skills work. Team orchestrator.md becomes a "routing configuration" that references shared protocol.

**Option C: Template Inheritance (Alternative Considered)**
```
orchestrator-base.md.tpl (shared)
    └── orchestrator-{team}.md.tpl (team overrides)
```

Rejected because: Adds template complexity; current single-template approach is sufficient.

### 3.2 Skill Structure: orchestrator-core

Create new skill at `user-skills/orchestration/orchestrator-core/`:

```
user-skills/orchestration/orchestrator-core/
  SKILL.md              # Entry point with full shared protocol
  schemas/
    consultation-request.md   # (moved from orchestrator-templates/references/)
    consultation-response.md  # (moved from orchestrator-templates/references/)
```

**Rationale**: The `orchestrator-templates` skill documents generation mechanics. The new `orchestrator-core` skill documents runtime protocol. These are distinct concerns.

### 3.3 SKILL.md Content Structure

```markdown
# Orchestrator Core Protocol

> Canonical protocol for all orchestrator agents. Team orchestrators reference this skill.

## Consultation Role

You are a **stateless advisor** that receives context and returns structured directives.

### What You DO
- Analyze initiative context and session state
- Decide which specialist should act next
- Craft focused prompts for specialists
- Define handoff criteria for phase transitions
- Surface blockers and recommend resolutions
- Maintain decision consistency across phases

### What You DO NOT DO
- Invoke the Task tool (you have no delegation authority)
- Read large files to analyze content (request summaries)
- Write code, PRDs, TDDs, or any artifacts
- Execute any phase yourself
- Make implementation decisions (that's specialist authority)
- Run commands or modify files

### The Litmus Test

Before responding, ask: *"Am I generating a prompt for someone else, or doing work myself?"*

If doing work yourself -> STOP. Reframe as guidance.

## Tool Access

**You have:** `Read` only

Use Read for:
- SESSION_CONTEXT.md (current session state)
- Approved artifacts (PRD, TDD) when summaries are insufficient
- Agent handoff notes

**You do NOT have and MUST NOT attempt:**
- Task (no subagent spawning)
- Edit/Write (no artifact creation)
- Bash (no command execution)
- Glob/Grep (no codebase exploration)

If you need information not in the consultation request, include it in your `information_needed` response field.

## Consultation Protocol

See schemas:
- @orchestrator-core/schemas/consultation-request.md
- @orchestrator-core/schemas/consultation-response.md

**Response Size Target**: Keep responses compact (~400-500 tokens).

## Core Responsibilities

- **Phase Decomposition**: Break complex work into ordered phases with clear boundaries
- **Specialist Routing**: Direct work to the right agent based on current phase and artifact readiness
- **Dependency Management**: Track what blocks what via state_update
- **Throughline Consistency**: Maintain decision rationale across consultations

## Domain Authority

**You decide:**
- Phase sequencing (what happens in what order)
- Which specialist handles which aspect
- When to parallelize vs. serialize phases
- When handoff criteria are sufficiently met
- Whether to pause pending clarification
- How to restructure when reality diverges from plan

**You escalate to User** (via `await_user` action):
- Scope changes affecting resources
- Unresolvable conflicts between specialist recommendations
- External dependencies outside team's control
- Decisions requiring product or business judgment

## Behavioral Constraints

**DO NOT** say: "Let me check the codebase to understand..."
**INSTEAD**: Request information in `information_needed` field.

**DO NOT** say: "I'll create the PRD now..."
**INSTEAD**: Return specialist prompt for appropriate specialist.

**DO NOT** say: "Let me verify the tests pass..."
**INSTEAD**: Define verification criteria for main agent to check.

**DO NOT** provide implementation guidance in your response text.
**INSTEAD**: Include implementation context in the specialist prompt.

**DO NOT** use tools beyond Read.
**INSTEAD**: Include what you need in `information_needed`.

**DO NOT** respond with prose explanations.
**INSTEAD**: Always use CONSULTATION_RESPONSE format.

## Handling Failures

When main agent reports specialist failure (type: "failure"):

1. **Understand**: Read the failure_reason carefully
2. **Diagnose**: Was it insufficient context? Scope too large? Missing prerequisite?
3. **Recover**: Generate new specialist prompt addressing the issue, OR recommend phase rollback
4. **Document**: Include diagnosis in throughline.rationale

You do NOT attempt to fix issues yourself.

## The Acid Test

*"Can I look at any piece of work in progress and immediately tell: who owns it, what phase it's in, what's blocking it, and what happens next?"*

Your CONSULTATION_RESPONSE should answer all of these.

## Anti-Patterns

- **Doing work**: Reading files to analyze, writing artifacts, running commands
- **Direct delegation**: Using Task tool (you don't have it)
- **Prose responses**: Answering conversationally instead of structured format
- **Scope creep tolerance**: New scope is new work; update state_update.next_phases
- **Vague handoffs**: "It's ready" is not valid—criteria must be explicit in specialist prompt
- **Micromanaging**: Let specialists own their domains; you provide prompts, not implementation guidance
```

### 3.4 Modified Template Structure

The `orchestrator-base.md.tpl` template changes from embedding shared content to referencing the skill:

**Current Template** (141 lines):
```markdown
---
name: orchestrator
...
---

# Orchestrator

[~90 lines of embedded shared protocol]

## Position in Workflow
{{WORKFLOW_DIAGRAM}}

[~20 lines of team-specific sections]
```

**New Template** (~60 lines):
```markdown
---
name: orchestrator
description: |
  {{DESCRIPTION}}
tools: Read
model: opus
color: {{COLOR}}
---

# Orchestrator

> {{ROLE}} for {{TEAM_NAME}}. For core protocol, see @orchestrator-core.

## Position in Workflow

```
{{WORKFLOW_DIAGRAM}}
```

**Upstream**: {{UPSTREAM}}
**Downstream**: {{DOWNSTREAM}}

## Phase Routing

| Specialist | Route When |
|------------|-----------|
{{ROUTING_TABLE}}

## Handoff Criteria

| Phase | Criteria |
|-------|----------|
{{HANDOFF_CRITERIA}}

{{#if CROSS_TEAM_PROTOCOL}}
## Cross-Team Protocol

{{CROSS_TEAM_PROTOCOL}}
{{/if}}

## Skills Reference

Reference these skills as appropriate:
{{SKILLS_REFERENCE}}

## Team-Specific Anti-Patterns

{{TEAM_ANTIPATTERNS}}
```

**Reduction**: 141 lines -> ~60 lines (57% reduction in template size)

## 4. Frontmatter Standardization Schema (Task-002)

### 4.1 Current Frontmatter Variations

Analysis of 11 orchestrator.yaml files shows inconsistent frontmatter:

| Field | Variations |
|-------|------------|
| description | 1-3 sentences, varied structure |
| tools | Some list "Read, Skill", others just "Read" |
| triggers | Inconsistent trigger word patterns |

### 4.2 Standardized Frontmatter Schema

```yaml
# orchestrator.yaml frontmatter section
frontmatter:
  role: string          # Required. One-line role description (< 80 chars)
  description: string   # Required. Multi-line description with:
                        #   - Primary function (sentence 1)
                        #   - Use conditions (sentence 2, "Use when: ...")
                        #   - Triggers (sentence 3, "Triggers: ...")
  tools: enum           # Required. Always "Read" for orchestrators
  model: enum           # Required. "opus" | "sonnet" (default: opus)
  color: string         # Required. Team color for UI
```

### 4.3 Description Template

```
{{PRIMARY_FUNCTION}}. Use when: {{USE_CONDITIONS}}. Triggers: {{TRIGGER_WORDS}}.
```

**Example**:
```yaml
description: |
  Coordinates ecosystem-pack phases for CEM/skeleton/roster infrastructure work. Use when: work spans multiple phases or requires cross-component coordination. Triggers: coordinate, orchestrate, multi-phase, ecosystem workflow.
```

### 4.4 Validation Rules

| Rule | Validation |
|------|------------|
| role length | <= 80 characters |
| description structure | Contains "Use when:" and "Triggers:" |
| tools value | Must be "Read" |
| model value | Must be "opus" or "sonnet" |
| color value | Must be valid CSS color name |

### 4.5 Schema JSON

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "frontmatter": {
      "type": "object",
      "required": ["role", "description", "tools", "model", "color"],
      "properties": {
        "role": {
          "type": "string",
          "maxLength": 80,
          "description": "One-line role description"
        },
        "description": {
          "type": "string",
          "pattern": ".*Use when:.*Triggers:.*",
          "description": "Multi-line description with use conditions and triggers"
        },
        "tools": {
          "type": "string",
          "enum": ["Read"],
          "description": "Orchestrators always have Read-only access"
        },
        "model": {
          "type": "string",
          "enum": ["opus", "sonnet"],
          "default": "opus"
        },
        "color": {
          "type": "string",
          "description": "CSS color name for UI"
        }
      }
    }
  }
}
```

## 5. Pattern Propagation Approach (Task-003)

### 5.1 Propagation Strategy

When `orchestrator-core` skill is updated, team orchestrators automatically benefit because they reference the skill rather than embed the content.

**No regeneration needed** for protocol changes in:
- Consultation Role section
- Tool Access section
- Behavioral Constraints section
- Handling Failures section
- Anti-Patterns (generic)

**Regeneration required** only for:
- Template structure changes
- New placeholder additions
- Frontmatter schema changes

### 5.2 Propagation Workflow

```
1. Update orchestrator-core/SKILL.md
   └── All orchestrators immediately get updated protocol
       (via @ reference at runtime)

2. Update orchestrator-base.md.tpl (if needed)
   └── Run: orchestrator-generate.sh --all --force
       └── Regenerate all 11 orchestrators
           └── Commit regenerated files
```

### 5.3 Versioning Strategy

The `orchestrator-core` skill should include a version marker:

```markdown
---
version: "1.0.0"
---
# Orchestrator Core Protocol
```

This enables:
- Change tracking for protocol evolution
- Backward compatibility checks
- Migration documentation when breaking changes occur

### 5.4 CI/CD Integration

```yaml
# .github/workflows/orchestrator-validation.yml
on:
  push:
    paths:
      - 'user-skills/orchestration/orchestrator-core/**'
      - 'templates/orchestrator-base.md.tpl'

jobs:
  validate:
    steps:
      - name: Validate skill structure
        run: |
          test -f user-skills/orchestration/orchestrator-core/SKILL.md

      - name: Check orchestrator references
        run: |
          for team in teams/*-pack; do
            grep -q "@orchestrator-core" "$team/agents/orchestrator.md" || \
              echo "WARNING: $team missing @orchestrator-core reference"
          done
```

## 6. Generation Script Modifications

### 6.1 orchestrator-generate.sh Changes

**Current behavior**: Substitutes all placeholders into embedded template content.

**New behavior**:
1. Substitutes team-specific placeholders only
2. Template contains skill reference (not embedded protocol)
3. Validation checks for @orchestrator-core reference

```bash
# Add to validate_substitution()
validate_orchestrator_reference() {
    local file="$1"

    if ! grep -q "@orchestrator-core" "$file"; then
        log_warn "Generated file missing @orchestrator-core reference: $file"
        return 1
    fi

    return 0
}
```

### 6.2 New Placeholders

| Placeholder | Source | Example |
|-------------|--------|---------|
| `{{TEAM_NAME}}` | orchestrator.yaml team.name | "ecosystem-pack" |
| `{{UPSTREAM}}` | orchestrator.yaml workflow_position.upstream | "User request via /ecosystem" |
| `{{DOWNSTREAM}}` | orchestrator.yaml workflow_position.downstream | "Implementation artifacts to session" |
| `{{HANDOFF_CRITERIA}}` | orchestrator.yaml handoff_criteria | Markdown table rows |
| `{{TEAM_ANTIPATTERNS}}` | orchestrator.yaml antipatterns | Bulleted list |

### 6.3 orchestrator.yaml Schema Extension

Add new required fields:

```yaml
# New fields in orchestrator.yaml
workflow_position:
  upstream: string    # What triggers this orchestrator
  downstream: string  # What outputs this orchestrator produces

antipatterns:         # Team-specific anti-patterns (already exists)
  - string
```

## 7. Migration Path

### 7.1 Phase 1: Create Skill (No Breaking Changes)

1. Create `user-skills/orchestration/orchestrator-core/SKILL.md`
2. Move consultation schemas from `orchestrator-templates/references/` to `orchestrator-core/schemas/`
3. Update `orchestrator-templates/references/` to reference new location
4. Sync skills to ~/.claude/skills/

**Validation**: Existing orchestrators continue to work (no changes to them yet)

### 7.2 Phase 2: Update Template

1. Modify `templates/orchestrator-base.md.tpl` to new minimal structure
2. Update `templates/orchestrator-generate.sh` for new placeholders
3. Test with single team: `orchestrator-generate.sh rnd-pack --dry-run`

**Validation**: Dry-run output is valid, contains @orchestrator-core reference

### 7.3 Phase 3: Regenerate All Teams

1. Run: `orchestrator-generate.sh --all --force`
2. Review diffs for all 11 teams
3. Commit regenerated orchestrator.md files

**Validation**: All orchestrators reference skill, team-specific content preserved

### 7.4 Phase 4: Update Frontmatter

1. Update each team's `orchestrator.yaml` to match standardized schema
2. Regenerate affected orchestrators
3. Update schema validation in generator

**Validation**: All frontmatter passes schema validation

### 7.5 Rollback Plan

If issues discovered:
1. Revert `orchestrator-base.md.tpl` to previous version
2. Run `orchestrator-generate.sh --all --force`
3. Commit reverted orchestrators

The `orchestrator-core` skill can remain (no harm in having unused skill).

## 8. File Changes Summary

### 8.1 New Files

| Path | Purpose |
|------|---------|
| `user-skills/orchestration/orchestrator-core/SKILL.md` | Shared orchestrator protocol |
| `user-skills/orchestration/orchestrator-core/schemas/consultation-request.md` | Request schema (moved) |
| `user-skills/orchestration/orchestrator-core/schemas/consultation-response.md` | Response schema (moved) |
| `schemas/orchestrator-frontmatter.schema.json` | Frontmatter validation schema |

### 8.2 Modified Files

| Path | Change |
|------|--------|
| `templates/orchestrator-base.md.tpl` | Reduce to minimal template with skill reference |
| `templates/orchestrator-generate.sh` | Add validation for @orchestrator-core, new placeholders |
| `teams/*/agents/orchestrator.md` | Regenerated with new structure (all 11) |
| `teams/*/orchestrator.yaml` | Frontmatter standardization (all 11) |
| `user-skills/orchestration/orchestrator-templates/references/consultation-protocol.md` | Update to reference new location |

### 8.3 Deleted Files

| Path | Reason |
|------|--------|
| (none) | Old schemas redirected, not deleted |

## 9. Integration Test Matrix

| Test Case | Input | Expected Output | Satellite Type |
|-----------|-------|-----------------|----------------|
| Skill load | `@orchestrator-core` | Full protocol content | All |
| Generate minimal team | New team YAML | Valid orchestrator.md with skill ref | Minimal |
| Generate complex team | Full team YAML | Valid orchestrator.md with all sections | Complex |
| Regenerate existing | 10x-dev-pack | No content loss, skill ref added | Standard |
| Schema validation | Invalid frontmatter | Validation error | N/A |
| Missing skill ref | Template without ref | Warning in validation | N/A |

## 10. Success Criteria

| Criterion | Verification |
|-----------|--------------|
| Skill created and synced | `ls ~/.claude/skills/orchestrator-core/SKILL.md` |
| Template reduced by >50% | `wc -l templates/orchestrator-base.md.tpl` < 70 |
| All orchestrators regenerated | 11 files in teams/*/agents/orchestrator.md |
| Skill reference in all orchestrators | `grep -l "@orchestrator-core" teams/*/agents/orchestrator.md \| wc -l` = 11 |
| Frontmatter validated | Generator reports no schema errors |
| No functionality regression | Manual test of /task workflow |

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Skill not loaded at runtime | Low | High | Test skill loading before migration |
| Team content lost in regeneration | Medium | High | Backup orchestrator.md files before regeneration |
| Generator compatibility broken | Low | Medium | Dry-run all teams before commit |
| Circular reference (skill -> template) | Low | Low | Clear separation of concerns |
| Frontmatter validation too strict | Medium | Low | Start permissive, tighten gradually |

## 12. Open Items

| Item | Owner | Status |
|------|-------|--------|
| Verify @ reference syntax works in agent prompts | Integration Engineer | Pending |
| Confirm skill sync includes nested schemas/ | Integration Engineer | Pending |
| Review consultation-protocol.md redirect | Integration Engineer | Pending |

## 13. Artifact Attestation

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-orchestrator-skill-extraction.md` | Created |
| Skill SKILL.md | `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-core/SKILL.md` | Pending |
| Template | `/Users/tomtenuta/Code/roster/templates/orchestrator-base.md.tpl` | Pending modification |
| Schema | `/Users/tomtenuta/Code/roster/schemas/orchestrator-frontmatter.schema.json` | Pending |

All artifacts to be verified via Read tool after implementation.
