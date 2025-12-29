---
name: initiative-scoping
description: "Session -1 and Session 0 protocols for initiative kickoff. Use when: starting new projects, scoping major initiatives, initializing the Orchestrator. Triggers: session -1, session 0, initiative scoping, project kickoff, new project, major initiative, initialization."
status: complete
---

# Initiative Scoping (Session -1/0)

> Protocols for initialization sessions that prepare the Orchestrator for execution.

## Decision Tree

```
User describes initiative
├─ Already scoped and validated?
│   ├─ YES, simple → Use /start directly
│   └─ YES, complex → Skip to Session 0
├─ Needs assessment?
│   └─ YES → Session -1 first
└─ After Session -1 GO → Session 0
```

**vs /start**: This skill is deliberative ("Should we? How?"). `/start` is direct execution ("Let's do it"). Use `/start` when scope is already validated.

## Session Quick Reference

| Session | Question | Orchestrator Output |
|---------|----------|---------------------|
| **-1** | "Should we do this?" | Go/No-Go + conditions |
| **0** | "How will we do this?" | Delegation map + plan |
| **1+** | Execution | Specialists produce artifacts |

Sessions -1 and 0 are **pre-work** - Orchestrator ingestion only. Real work begins in Session 1.

## Session -1: Initiative Assessment

**Input**: Initiative description from user

**Output**: North Star, Go/No-Go, workflow sizing, blocking questions, risks

**Protocol**: [session-minus-1-protocol.md](session-minus-1-protocol.md)

## Session 0: Orchestrator Initialization

**Input**: Initiative context + Session -1 output (if available)

**Output**: North Star, 10x Plan, Delegation Map, blocking questions, risks

**Protocol**: [session-0-protocol.md](session-0-protocol.md)

## When to Use Each Session

| Scenario | Session -1? | Session 0? |
|----------|-------------|------------|
| New feature (complex) | Yes | Yes |
| New feature (simple) | No | Yes |
| Major refactoring | Yes | Yes |
| Bug fix (isolated) | No | No |
| Bug fix (cross-cutting) | Yes | Yes |
| Exploration/spike | No | No |

## Progressive Disclosure

| Content | Location |
|---------|----------|
| Agent hierarchy & behavior rules | [shared-principles.md](shared-principles.md) |
| Session -1 execution details | [session-minus-1-protocol.md](session-minus-1-protocol.md) |
| Session 0 execution details | [session-0-protocol.md](session-0-protocol.md) |
| Workflow definition | [10x-workflow](../10x-workflow/SKILL.md) |
| Agent invocation patterns | [prompting](../prompting/SKILL.md) |

## Related Skills

| Skill | Purpose |
|-------|---------|
| [start-ref](../start-ref/skill.md) | Direct session start (scope already validated) |
| [10x-workflow](../10x-workflow/SKILL.md) | Defines the workflow (do not repeat) |
| [documentation](../documentation/SKILL.md) | Templates for specialists |
| [prompting](../prompting/SKILL.md) | Agent invocation patterns |

## Archived Templates

For historical reference, legacy user-input templates are archived:
- [prompt-minus-1.md](.archive/prompt-minus-1.md)
- [prompt-0.md](.archive/prompt-0.md)

These are **not required** for current workflow.
