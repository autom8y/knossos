---
name: orchestration
description: "Consultation protocol routing hub for multi-phase workflows. Use when: running /task, /sprint, or /consolidate commands, implementing consultation loop, coordinating specialist handoffs. Triggers: orchestration, consultation loop, workflow coordination, multi-phase, orchestrator, specialist handoff, directive parsing."
---

# Orchestration Skill

Quick routing hub for consultation protocol documentation. Domain-specific files contain the details.

## When to Use This Skill

1. **Understanding the consultation loop**: Start with `main-thread-guide.md`
2. **Building requests/responses**: See Quick Reference table
3. **Integrating commands**: Read `command-integration.md`
4. **Debugging workflows**: Check invariants in `consultation-loop.md`

Do NOT use this skill to get loop diagrams or schemas directly--it is a routing hub.

## Quick Reference

| Need | File | Description |
|------|------|-------------|
| Primary entry point | `main-thread-guide.md` | How main thread executes the loop |
| Loop pattern + invariants | `consultation-loop.md` | ASCII diagram, cycle counts, token economics |
| Request schema | `request-format.md` | CONSULTATION_REQUEST structure and examples |
| Response schema | `response-format.md` | CONSULTATION_RESPONSE actions and prompts |
| Command implementation | `command-integration.md` | How /task, /sprint, /consolidate use the loop |
| Specialist output | `specialist-returns.md` | What specialists return (P3 format) |
| Execution mode | `execution-mode.md` | When to delegate vs. execute directly |
| Entry pattern | `entry-pattern.md` | How /start, /sprint, /task route through hooks |

## File Ownership

This skill **routes** to protocol files--it does not contain protocol details itself.

For actual schemas and patterns, use the appropriate file:
- Loop mechanics and invariants --> `consultation-loop.md`
- Request/response structure --> `request-format.md`, `response-format.md`
- Main thread behavior --> `main-thread-guide.md`

## Core Concept

The orchestrator is CONSULTED, not EXECUTED. Main thread owns the Task tool and all specialist invocations. Orchestrator is stateless--receives summaries, returns directives.

```
Main Thread --[consult]--> Orchestrator
            <--[directive]--
Main Thread --[Task tool]--> Specialist
            <--[artifact]--
Main Thread --[checkpoint]--> Orchestrator
```

See `consultation-loop.md` for the full loop diagram and `main-thread-guide.md` for execution details.

## Loop Invariants (Summary)

1. **Main agent owns Task tool** - only main agent invokes specialists
2. **Orchestrator is stateless** - all state comes from request
3. **Summaries not files** - main agent summarizes artifacts for orchestrator
4. **Structured formats only** - no prose back-and-forth
5. **Throughline tracking** - every response includes decision/rationale

Canonical source: `consultation-loop.md#loop-invariants`

## Context Overflow Guidance

If context is constrained (P2 situation):

1. **Minimal path**: Read only `main-thread-guide.md` for execution pattern
2. **Request building**: Add `request-format.md` for schema
3. **Response parsing**: Add `response-format.md` for directive handling
4. **Full understanding**: Add `consultation-loop.md` for invariants and token economics

Avoid loading all files simultaneously. Route to specific files based on immediate need.

## Token Economics (Summary)

| Component | Target |
|-----------|--------|
| CONSULTATION_REQUEST | 200-400 tokens |
| CONSULTATION_RESPONSE | 400-500 tokens |
| Specialist prompt (embedded) | 200-300 tokens |

Details and rationale: `consultation-loop.md#token-economics`

## Getting Started

New to orchestration? Read in this order:

1. `main-thread-guide.md` - Understand the coach pattern
2. `execution-mode.md` - Know when to delegate
3. `consultation-loop.md` - See the full loop mechanics

Building a command? Add:

4. `command-integration.md` - Integration template
5. `request-format.md` and `response-format.md` - Schemas

## Anti-Patterns

- Asking orchestrator to "execute the sprint" (it cannot execute)
- Passing full file contents in requests (use summaries)
- Prose conversation instead of structured YAML
- Main thread using Edit/Write during active workflow

## Related Resources

- Orchestrator agent: `.claude/agents/orchestrator.md`
- Base template: `roster/shared/base-orchestrator.md`
- Workflow detection: Check `SESSION_CONTEXT.md` for `workflow.active`
