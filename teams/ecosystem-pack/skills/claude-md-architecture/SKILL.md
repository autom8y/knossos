---
name: claude-md-architecture
description: "First principles for CLAUDE.md architecture. Use when: modifying CLAUDE.md content, deciding CEM sync behavior, determining content placement, validating CLAUDE.md changes. Triggers: CLAUDE.md architecture, CLAUDE.md sync, CLAUDE.md content placement, CLAUDE.md section ownership, CLAUDE.md behavioral contract, CLAUDE.md vs SESSION_CONTEXT, CLAUDE.md anti-patterns, CLAUDE.md validation."
---

# CLAUDE.md Architecture

> First principles for what belongs in CLAUDE.md and why.

## When to Use This Skill

Activate this skill when:

- Modifying any CLAUDE.md content (skeleton or satellite)
- Making CEM sync decisions (SYNC vs PRESERVE vs REGENERATE)
- Determining where content belongs (CLAUDE.md vs SESSION_CONTEXT vs hooks)
- Validating proposed CLAUDE.md changes
- Resolving content placement disputes

---

## The Core Question

> "Is this a behavioral contract (what Claude can do and how) or transient state (what's happening now)?"

CLAUDE.md is a **behavioral contract**, not a knowledge base, session log, or scratchpad.

---

## Quick Reference

### The Stability Rule

```
CLAUDE.md contains: STABLE content (changes weeks/months)
CLAUDE.md excludes: DYNAMIC + EPHEMERAL content (changes daily/hourly)
```

### The Decay Test

> "If I don't update this for a month, is CLAUDE.md incorrect?"

- **No** (still accurate) -> Belongs in CLAUDE.md
- **Yes** (becomes stale) -> Does not belong

### Section Ownership Quick Reference

| Owner | Sync Behavior | Examples |
|-------|---------------|----------|
| Skeleton | SYNC | Skills docs, hooks docs, workflow patterns |
| Satellite | PRESERVE | Project extensions, custom sections |
| Roster | REGENERATE | Quick Start, Agent Configurations |
| Session | NOT IN CLAUDE.md | Current task, git state, handoff context |

---

## Progressive Disclosure

**Core Concepts**:

- [first-principles.md](first-principles.md) - The 6 foundational principles, decision record, layering model
- [ownership-model.md](ownership-model.md) - Section ownership, sync behaviors, marker syntax
- [boundary-test.md](boundary-test.md) - 5-question validation checklist
- [anti-patterns.md](anti-patterns.md) - What NOT to put in CLAUDE.md (11 anti-patterns)

**Related Skills**:

- [ecosystem-ref](../ecosystem-ref/SKILL.md) - CEM implementation and sync mechanics
- [documentation](../documentation/SKILL.md) - General documentation standards
- [standards](../standards/SKILL.md) - Repository conventions

---

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
 SKELETON   ROSTER   SATELLITE
    |         |          |
    v         v          v
  SYNC    REGENERATE  PRESERVE
 section   from state  section
```

---

## Validation Checklist (Quick)

Before modifying CLAUDE.md:

- [ ] Content passes Stability Test (accurate in 1 month)
- [ ] Content passes Source of Truth Test (CLAUDE.md is authoritative)
- [ ] Content passes Scope Test (project-wide, not session-specific)
- [ ] No dates, timestamps, or "currently" language
- [ ] No git state or file status references
- [ ] Correct owner identified (skeleton/satellite/roster)
- [ ] Correct sync behavior specified (SYNC/PRESERVE/REGENERATE)

See [boundary-test.md](boundary-test.md) for the complete 5-question validation.

---

## Agent Invocation Guidance

When modifying CLAUDE.md content, follow these pre-flight and post-flight patterns:

### Pre-Flight Checklist

Before making changes:

1. **Read current state**: Always use Read tool on CLAUDE.md before editing
2. **Identify section owner**: Check for ownership markers (`<!-- SYNC: -->`, `<!-- PRESERVE: -->`)
3. **Validate with boundary test**: Run 5-question test (see [boundary-test.md](boundary-test.md))
4. **Check anti-patterns**: Ensure content does not match any of the 11 anti-patterns

### Post-Flight Verification

After making changes:

1. **Verify marker placement**: Ownership marker immediately precedes section header
2. **Validate stability**: Content passes "accurate in 1 month" test
3. **Check propagation intent**: SYNC content should be skeleton-owned infrastructure only
4. **No session state leaked**: No dates, "currently", git state, or task references

### Common Modification Scenarios

| Scenario | Pre-Flight | Post-Flight |
|----------|------------|-------------|
| Add project extension | Check `## Project:*` namespace available | Verify PRESERVE marker added |
| Update workflow routing | Confirm this is skeleton content | Ensure SYNC marker preserved |
| Fix team configuration | Read ACTIVE_TEAM + agents/ | Verify regeneration source noted |
| Remove stale content | Confirm it fails decay test | Verify no orphaned references |
