---
name: agent-prompt-engineering
description: "Agent prompt engineering standards. Use when: writing new agent prompts, auditing existing prompt quality, debugging agent behavior, optimizing prompt token cost. Triggers: agent prompt, system prompt engineering, prompt rubric, prompt optimization, agent design."
---

# Agent Prompt Engineering

> Standards for writing agent prompts that work the first time

## Overview

**Target 150-200 lines per agent.** Agents exceeding 250 lines likely contain redundancy.

Effective agent prompts share three qualities: **clarity** (the agent knows exactly what it does), **boundaries** (the agent owns vs. escalates), and **testability** (handoff criteria are objectively verifiable).

## When NOT to Use This Skill

Skip full template compliance for:
- Quick experimental agents under 50 lines (prototypes, one-off tests)
- Single-purpose scripts without workflow integration
- Agents discarded after one session (throwaway debugging aids)

For these cases, include only: role identity (2 sentences) and basic responsibilities.

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

Full details with fixes: [principles.md](principles.md). For skill user mistakes, see [skill-anti-patterns.md](skill-anti-patterns.md).

## Companion Reference

| Topic | File | When to Load |
|-------|------|--------------|
| 9 core principles + anti-patterns | [principles.md](principles.md) | Writing prompts or debugging behavior |
| 11-section agent template | [template.md](template.md) | Creating a new agent |
| 6-dimension quality rubric | [scoring/rubric.md](scoring/rubric.md) | Auditing existing agents |
| Pre-deployment checklist | [validation/checklist.md](validation/checklist.md) | Before shipping any agent |
| Before/after examples | [examples/before-after.md](examples/before-after.md) | Learning transformation patterns |
| Skill user anti-patterns | [skill-anti-patterns.md](skill-anti-patterns.md) | Diagnosing skill misuse |

## Escalation

Route to human when:
- Designing entirely new agent patterns (not variations of existing)
- Audit reveals systemic issues across 3+ agents
- Rite structure changes require workflow redesign
