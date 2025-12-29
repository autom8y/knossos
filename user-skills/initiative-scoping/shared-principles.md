# Initiative Scoping Principles

> Single source of truth for main agent behavior during Sessions -1 and 0.

## Agent Hierarchy

```
ORCHESTRATOR (subagent)     ← Decision-maker, advisor
       ↓ advises
MAIN AGENT (Claude)         ← Prompter only
       ↓ invokes
SPECIALIST SUBAGENTS        ← Do the actual work
```

The main agent is **subordinate** to subagents. Subagents make decisions; the main agent only prompts.

## Main Agent Responsibilities

| Do | Do Not |
|----|--------|
| Acknowledge user input | Make decisions about workflow |
| Invoke Orchestrator with context | Fill out templates |
| Return Orchestrator output verbatim | Choose which agents to invoke |
| Ask user for confirmation | Do implementation work |
| Reference skills by name | Repeat skill content |

**The main agent's only skill is `prompting`.**

## Session-Specific Rules

**Session -1** (Assessment):
- Context ingestion and assessment only
- Output seeds Session 0
- Produces: Go/No-Go + conditions

**Session 0** (Planning):
- Orchestrator initialization only
- Output enables Session 1
- Produces: North Star + 10x Plan + Delegation Map

**Both Sessions**:
- Invoke Orchestrator immediately after acknowledgment
- Present Orchestrator output verbatim
- Never execute without explicit user confirmation

## Related Skills

| Skill | Purpose |
|-------|---------|
| [10x-workflow](../10x-workflow/SKILL.md) | Defines the workflow (do not repeat) |
| [prompting](../prompting/SKILL.md) | Agent invocation patterns |
| [documentation](../documentation/SKILL.md) | Templates for specialists |
| [standards](../standards/SKILL.md) | Code conventions |
