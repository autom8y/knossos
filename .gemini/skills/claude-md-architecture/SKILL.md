---
description: 'CLAUDE.md architecture first principles. Use when: modifying CLAUDE.md content, deciding where content belongs, making ari sync placement decisions. Triggers: CLAUDE.md architecture, CLAUDE.md sync, content placement, section ownership.'
name: claude-md-architecture
version: "1.0"
---
---
name: claude-md-architecture
description: "CLAUDE.md architecture first principles. Use when: modifying CLAUDE.md content, deciding where content belongs, making ari sync placement decisions. Triggers: CLAUDE.md architecture, CLAUDE.md sync, content placement, section ownership."
---

# CLAUDE.md Architecture

> First principles for what belongs in CLAUDE.md and why.

## Purpose of CLAUDE.md

CLAUDE.md is the **entry point** for Claude Code. It answers three questions:

1. **What is this project?** (Rite, agents, capabilities)
2. **What patterns are available?** (Skills, hooks, workflows)
3. **Where do I go for guidance?** (Routing, help resources)

CLAUDE.md is a **behavioral contract**, not a knowledge base, session log, or scratchpad.

## Quick Reference

### The Stability Rule

```
CLAUDE.md contains: STABLE content (changes weeks/months)
CLAUDE.md excludes: DYNAMIC + EPHEMERAL content (changes daily/hourly)
```

### Section Ownership

| Owner | Sync Behavior | Examples |
|-------|---------------|----------|
| Knossos | SYNC | Skills docs, hooks docs, workflow patterns |
| Satellite | PRESERVE | Project extensions, custom sections |
| Rite | REGENERATE | Quick Start, Agent Configurations |
| Session | NOT IN CLAUDE.md | Current task, git state, handoff context |

### The Decay Test

> "If I don't update this for a month, is CLAUDE.md incorrect?"

- **No** (still accurate) → Belongs in CLAUDE.md
- **Yes** (becomes stale) → Does not belong

## Companion Reference

| Topic | File | When to Load |
|-------|------|-------------|
| 6 foundational principles, layering model | [first-principles.md](first-principles.md) | Core architecture understanding |
| Section ownership, sync behaviors, marker syntax | [ownership-model.md](ownership-model.md) | Making sync/placement decisions |
| 5-question validation checklist | [boundary-test.md](boundary-test.md) | Validating proposed changes |
| 11 anti-patterns — what NOT to put in CLAUDE.md | [anti-patterns.md](anti-patterns.md) | Content exclusion decisions |
| Descriptive vs prescriptive tone examples | [content-tone-guide.md](content-tone-guide.md) | Writing section content |

## Decision Flowchart

```
New content to add to CLAUDE.md?
           |
           v
  Stable for 1 month? ----NO----> NOT in CLAUDE.md
           |                      (Use SESSION_CONTEXT or hooks)
          YES
           |
           v
  Project-wide scope? ----NO----> SESSION_CONTEXT
           |
          YES
           |
           v
  Who owns this content?
     /        |        \
  KNOSSOS   RITE    SATELLITE
    |         |          |
    v         v          v
  SYNC    REGENERATE  PRESERVE
 section   from state  section
```

## Related Skills

- [ecosystem-ref](../ecosystem-ref/SKILL.md) — Sync pipeline mechanics
- `execution-mode` skill — Enforcement rules (when delegation is required)
