---
name: agent-prompt-engineering
description: "Agent prompt engineering standards. Use when: writing new agent prompts, auditing existing prompt quality, debugging agent behavior, optimizing prompt token cost. Triggers: agent prompt, system prompt engineering, prompt rubric, prompt optimization, agent design."
---

# Agent Prompt Engineering

> Standards for writing agent prompts that work the first time

## Overview

**Target 150-200 lines per agent.** Agents exceeding 250 lines likely contain redundancy.

This skill codifies learnings from 10 deep optimization sprints across production agent pantheons. Use it to create new agents, audit existing prompts, or debug agent behavior problems.

Effective agent prompts share three qualities: **clarity** (the agent knows exactly what it does), **boundaries** (the agent knows what it owns vs. escalates), and **testability** (handoff criteria are objectively verifiable).

## The Acid Test

*"Could someone unfamiliar with agent development create a production-ready agent using only this skill?"*

If uncertain: The skill needs more concrete examples or clearer step-by-step guidance.

## When NOT to Use This Skill

Skip full template compliance for:
- Quick experimental agents under 50 lines (prototypes, one-off tests)
- Single-purpose scripts without workflow integration
- Agents discarded after one session (throwaway debugging aids)

For these cases, include only: role identity (2 sentences) and basic responsibilities.

## Invocation

Via Skill tool: `skill: "agent-prompt-engineering"`

Example prompts:
- "Create a new agent following agent-prompt-engineering template for {purpose}"
- "Audit this agent prompt against the rubric"

## Decision Tree

**Creating a new agent?**
1. Read [template.md](template.md) - 11-section structure
2. Review [principles.md](principles.md) - 9 core principles with integrated anti-patterns
3. Validate against [validation/checklist.md](validation/checklist.md)

**Auditing an existing agent?**
1. Apply [scoring/rubric.md](scoring/rubric.md) - 6-dimension assessment
2. Run [validation/checklist.md](validation/checklist.md) - pre-deployment checks
3. If score < 4.0, see [examples/](examples/) for transformation patterns

**Debugging agent behavior?**
1. Check [principles.md](principles.md) - match symptoms to anti-patterns in each principle
2. Review Anti-Pattern Summary table (below) for detection patterns
3. Apply targeted fix from [examples/](examples/)

## Quick Reference

| Section | Purpose | Target Lines |
|---------|---------|--------------|
| **Frontmatter** | YAML metadata for agent discovery | 15-25 |
| **Title + Overview** | Role identity in 2-3 sentences | 3-5 |
| **Core Responsibilities** | 4-6 action-verb bullet points | 6-10 |
| **Position in Workflow** | ASCII diagram + upstream/downstream | 8-12 |
| **Domain Authority** | Decide / Escalate / Route breakdown | 12-18 |
| **How You Work** | 3-4 phase methodology | 20-30 |
| **What You Produce** | Artifact table + primary template | 15-25 |
| **Handoff Criteria** | 5-7 verifiable checklist items | 8-12 |
| **The Acid Test** | Single pivotal yes/no question | 3-5 |
| **Skills Reference** | Cross-references to related skills | 4-6 |
| **Anti-Patterns** | 3-5 specific failure modes | 8-12 |

## Quick Example: Role Identity Transformation

**Before** (vague):
```markdown
# Documentation Agent
This agent helps with documentation-related activities and supports the rite.
```

**After** (crystal clear):
```markdown
# Documentation Engineer
Transforms technical specifications into human-readable guides. The translator between developer knowledge and user understanding.
```

**Key improvements**: Active voice, specific purpose, no "helps with" language, establishes metaphor (translator) for role clarity.

## Model and Color Assignment

### Model Selection

| Role Type | Model | Rationale |
|-----------|-------|-----------|
| Orchestration/Senior | opus | Complex coordination, multi-phase planning |
| Analyst/Architect | opus | Deep analysis, design decisions |
| Implementation | sonnet | Balanced speed/quality for coding |
| Documentation | sonnet | Content creation, moderate complexity |
| Assessment/Triage | haiku | Fast discovery, high volume |

### Color Assignment

| Role Type | Color | Examples |
|-----------|-------|----------|
| Coordination | purple | orchestrator, incident-commander |
| Requirements/Entry | pink/orange | requirements-analyst, observability-engineer |
| Design/Architecture | cyan | architect, platform-engineer |
| Execution/Implementation | green | principal-engineer, janitor, tech-writer |
| Validation/Testing | red | qa-adversary, chaos-engineer |

**Rule**: Avoid duplicate colors within the same pantheon for visual distinction.

## Task Tool Patterns

Create or audit agents via Task tool delegation:

**Create new agent:**
```
/task Create a new agent for [domain]. Use agent-prompt-engineering skill: start with template.md, apply principles.md, validate against checklist.md, score with rubric.md. Target 150-200 lines.
```

**Audit existing agent:**
```
/task Audit [agent-name].md using agent-prompt-engineering skill: score against rubric.md, check principles.md anti-patterns, validate with checklist.md. Report scores and recommended fixes.
```

## Skill Contents

| File | Purpose |
|------|---------|
| [principles.md](principles.md) | 9 core principles with integrated anti-patterns and detection checklist |
| [template.md](template.md) | 11-section agent template with inline guidance |
| [scoring/rubric.md](scoring/rubric.md) | 6-dimension quality rubric (1-5 scale) |
| [validation/checklist.md](validation/checklist.md) | Pre-deployment verification checklist |
| [examples/before-after.md](examples/before-after.md) | Real transformation examples |

## When to Use This Skill

**Creating a new agent:**
1. Start with [template.md](template.md)
2. Apply [principles.md](principles.md) while writing
3. Validate against [validation/checklist.md](validation/checklist.md)
4. Score using [scoring/rubric.md](scoring/rubric.md)

**Auditing existing agents:**
1. Score each agent using [scoring/rubric.md](scoring/rubric.md)
2. Check Anti-Pattern Summary table above for detection patterns
3. Compare against [examples/before-after.md](examples/before-after.md)
4. Apply fixes following [principles.md](principles.md)

**Debugging agent behavior:**
1. Check if frontmatter description matches actual behavior
2. Verify Domain Authority section defines boundaries
3. Review Handoff Criteria for measurability
4. Look for anti-pattern violations in principles.md

## Escalation

Route to human when:
- Designing entirely new agent patterns (not variations of existing)
- Audit reveals systemic issues across 3+ agents
- Rite structure changes require workflow redesign
- Conflicting requirements between skill principles

**Non-prompt issues**: Agent performance -> infrastructure team. Tool limitations -> platform team. Model behavior -> `/consult`.

## Cross-Skill Integration

| Skill | Relationship |
|-------|--------------|
| `rite-development` skill | Uses this skill's template for new agents |
| `documentation` skill | Artifact templates referenced by agents |
| `standards` skill | Code conventions agents should enforce |
| `file-verification` skill | Verification protocol agents should reference |

## Quality Targets

A well-engineered agent prompt:
- Scores 4+ on all 6 rubric dimensions
- Passes all validation checklist items
- Contains zero anti-patterns (see Anti-Pattern Summary above)
- Stays under 200 lines
- Uses active voice throughout
- Has objectively testable handoff criteria

## Anti-Pattern Summary

Quick reference for the 7 most common prompt failures:

| Anti-Pattern | Core Issue | Detection |
|--------------|------------|-----------|
| Vague Role | Scope unclear | "helps with", "works on" |
| Passive Instructions | Inconsistent interpretation | "should be", "might need" |
| Implicit Escalation | Wrong decisions made | Missing "You escalate" |
| Fuzzy Handoff | Indefinite cycling | "quality", "complete", "ready" |
| Over-Explanation | Token waste | "As you know", 250+ lines |
| Generic Examples | No learning signal | Could apply to any agent |
| Missing Anti-Patterns | Repeated failures | No domain-specific guardrails |

Full details with fixes: [principles.md](principles.md) (integrated into each principle). For skill user mistakes (vs agent author mistakes), see [skill-anti-patterns.md](skill-anti-patterns.md).

## Related Documentation

- Agent template: `.claude/skills/rite-development/templates/agent-template.md`
- Existing agents: `.claude/agents/*.md`
- Rite catalog: Check `rites/` directory for production examples
