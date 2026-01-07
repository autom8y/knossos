---
name: claude-md-architecture
description: "First principles for CLAUDE.md architecture. Use when: modifying CLAUDE.md content, deciding CEM sync behavior, determining content placement, validating CLAUDE.md changes. Triggers: CLAUDE.md architecture, CLAUDE.md sync, CLAUDE.md content placement, CLAUDE.md section ownership, CLAUDE.md behavioral contract, CLAUDE.md vs SESSION_CONTEXT, CLAUDE.md anti-patterns, CLAUDE.md validation."
---

# CLAUDE.md Architecture

> First principles for what belongs in CLAUDE.md and why.

## When to Use This Skill

Activate this skill when:

- Modifying any CLAUDE.md content (roster or satellite)
- Making CEM sync decisions (SYNC vs PRESERVE vs REGENERATE)
- Determining where content belongs (CLAUDE.md vs SESSION_CONTEXT vs hooks)
- Validating proposed CLAUDE.md changes
- Resolving content placement disputes

---

## Purpose of CLAUDE.md

CLAUDE.md serves as the **entry point** for Claude Code. It answers three questions:

1. **What is this project?** (Team, agents, capabilities)
2. **What patterns are available?** (Skills, hooks, workflows)
3. **Where do I go for guidance?** (Routing, help resources)

CLAUDE.md is NOT where enforcement happens. It describes; orchestration enforces.

### The Entry Point Principle

CLAUDE.md is the first file Claude reads in any project. Content should:

- **Orient** new users to available capabilities
- **Route** to appropriate resources for deeper context
- **Describe** patterns without mandating behavior

The guiding question:

> "Is this a behavioral contract (what Claude can do and how) or transient state (what's happening now)?"

CLAUDE.md is a **behavioral contract**, not a knowledge base, session log, or scratchpad.

---

## Content Architecture Principles

### Principle 1: Descriptive Over Prescriptive

CLAUDE.md describes what's available and when patterns apply. Enforcement belongs in the orchestration layer.

| Prescriptive (Avoid) | Descriptive (Preferred) |
|---------------------|------------------------|
| "MUST delegate via Task tool" | "Orchestrated workflows coordinate via Task tool delegation" |
| "NEVER use Edit/Write directly" | "Specialists handle implementation; main thread coordinates" |
| "You must include session context" | "state-mate requires session context for mutations" |

**Transformation patterns:**

- Replace "MUST do X" with "X applies when Y"
- Replace global mandates with conditional tables
- Replace commands with explanations of purpose

### Principle 2: Conditional Guidance with Clear Triggers

When behavior varies by context, use conditional tables that show which pattern applies when.

**Three-Mode Model** (from hybrid session model):

| Mode | Session | Team | Behavior |
|------|---------|------|----------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Coach pattern, delegate via Task tool |

Entry-point text should acknowledge all three modes. The user or orchestrator determines which applies based on current state.

### Principle 3: Route to /consult for Uncertainty

When the reader is unsure which pattern applies, route to `/consult` rather than prescribing behavior.

**Pattern:**
```
[Describe what's available]
**Unsure?** Use `/consult` for routing guidance.
```

This positions `/consult` as the decision helper, keeping entry-point text descriptive.

### Principle 4: Enforcement Lives in Orchestration

CLAUDE.md describes; orchestration enforces. The `execution-mode.md` file defines when delegation is required. CLAUDE.md should reference but not duplicate these rules.

**Correct pattern:**
```markdown
## Execution Mode

This project supports three operating modes...
[Mode table]

For enforcement rules: `orchestration/execution-mode.md`
```

**Incorrect pattern:**
```markdown
## Execution Mode

**Active workflow?** MUST delegate via Task tool.
```

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
| Roster | SYNC | Skills docs, hooks docs, workflow patterns |
| Satellite | PRESERVE | Project extensions, custom sections |
| Team | REGENERATE | Quick Start, Agent Configurations |
| Session | NOT IN CLAUDE.md | Current task, git state, handoff context |

---

## Section Content Guide

What belongs in each CLAUDE.md section:

### Roster-Owned Sections (SYNC)

These sections sync from roster to all satellites. Content describes ecosystem infrastructure.

| Section | Purpose | Content Pattern |
|---------|---------|-----------------|
| **Execution Mode** | Explain operating modes | Three-mode table, reference to execution-mode.md |
| **Agent Routing** | Describe routing pattern | Conditional explanation, /consult as escape hatch |
| **Skills** | List available skills | Skill invocation pattern, key skill examples |
| **Hooks** | Document auto-injection | Hook trigger types, manual context note |
| **Dynamic Context** | Explain ! syntax | Usage pattern, when to prefer hooks |
| **Getting Help** | Route to resources | Question-to-skill mapping table |
| **State Management** | Describe state-mate | Usage pattern, control flags |
| **Slash Commands** | Set response expectation | Outcome requirement note |

**Tone for SYNC sections:** Explanatory. Describes what exists and how it works. References enforcement rules rather than stating them directly.

### Satellite-Owned Sections (PRESERVE)

These sections are never overwritten by CEM sync. Content is project-specific.

| Section | Purpose | Content Pattern |
|---------|---------|-----------------|
| **Quick Start** | Orient to current team | Rite name, agent table (regenerated from roster) |
| **Agent Configurations** | List available agents | Agent-to-file mapping (regenerated from roster) |
| **Project:\*** namespace | Project extensions | Any project-specific patterns |
| Custom sections | Satellite additions | Not matching roster section names |

**Tone for PRESERVE sections:** Contextual. Describes this specific project's configuration.

### Content That Does NOT Belong

See [anti-patterns.md](anti-patterns.md) for the full list. Key exclusions:

- Session state (current task, work in progress)
- Dynamic git references (branch, uncommitted files)
- Timestamps or "last updated" markers
- Personal preferences (belong in ~/.claude/CLAUDE.md)

---

## Progressive Disclosure

**Core Concepts**:

- [first-principles.md](first-principles.md) - The 6 foundational principles, decision record, layering model
- [ownership-model.md](ownership-model.md) - Section ownership, sync behaviors, marker syntax
- [boundary-test.md](boundary-test.md) - 5-question validation checklist
- [anti-patterns.md](anti-patterns.md) - What NOT to put in CLAUDE.md (11 anti-patterns)
- [content-tone-guide.md](content-tone-guide.md) - Examples of descriptive vs prescriptive patterns

**Related Skills**:

- [ecosystem-ref](../ecosystem-ref/SKILL.md) - CEM implementation and sync mechanics
- [documentation](~/.claude/skills/documentation/SKILL.md) - General documentation standards
- [standards](~/.claude/skills/standards/SKILL.md) - Repository conventions

**Enforcement Reference**:

- [execution-mode.md](~/.claude/skills/orchestration/execution-mode.md) - When delegation is required (source of enforcement rules)

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
  ROSTER    TEAM    SATELLITE
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
- [ ] Correct owner identified (roster/team/satellite)
- [ ] Correct sync behavior specified (SYNC/PRESERVE/REGENERATE)
- [ ] **Descriptive tone** (no global MUST mandates in entry sections)
- [ ] **Conditional guidance** where behavior varies by context
- [ ] **/consult** referenced as routing escape hatch

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
5. **Verify tone**: Content describes rather than prescribes

### Post-Flight Verification

After making changes:

1. **Verify marker placement**: Ownership marker immediately precedes section header
2. **Validate stability**: Content passes "accurate in 1 month" test
3. **Check propagation intent**: SYNC content should be roster-owned infrastructure only
4. **No session state leaked**: No dates, "currently", git state, or task references
5. **Descriptive tone verified**: No global MUST mandates in entry sections

### Common Modification Scenarios

| Scenario | Pre-Flight | Post-Flight |
|----------|------------|-------------|
| Add project extension | Check `## Project:*` namespace available | Verify PRESERVE marker added |
| Update workflow routing | Confirm this is roster content | Ensure SYNC marker preserved, descriptive tone |
| Fix team configuration | Read ACTIVE_RITE + agents/ | Verify regeneration source noted |
| Remove stale content | Confirm it fails decay test | Verify no orphaned references |
| Convert prescriptive to descriptive | Identify MUST mandates | Verify conditional tables used |
