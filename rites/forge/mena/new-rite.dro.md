---
name: new-rite
description: Create a new rite through The Forge workflow (direct creation)
argument-hint: "<rite-name> [--complexity=PATCH|RITE|ECOSYSTEM]"
allowed-tools: Bash, Glob, Grep, Read, Write, Edit, Task, TodoWrite
model: opus
---

## Context

Auto-injected by SessionStart hook (project, rite, session context).

## Your Task

Create a new rite: $ARGUMENTS

## Behavior

### 1. Parse Arguments

- `rite-name`: Required. Name for the new rite
- `--complexity`: Optional. Default is RITE.
  - PATCH: Single agent modification (phases: design, prompting, validation)
  - RITE: Full rite with 3-5 agents (all phases)
  - ECOSYSTEM: Multi-rite initiative (all phases)

If the user passes `--deep` or `--interview`, respond with:
> For archaeology-first rite creation with domain expertise extraction, use `/forge-rite <name>` instead.

### 2. Invoke Agent Designer

Start the Forge workflow by invoking the Agent Designer:

```
Use the Task tool to invoke the agent-designer agent with:

"Create a RITE-SPEC for {rite-name}.

The user wants to create a new rite. Your job is to:
1. Clarify the rite's purpose with the user
2. Design 3-5 agent roles with clear boundaries
3. Define input/output contracts
4. Specify complexity levels
5. Document handoff criteria

Complexity level: {complexity}

When the RITE-SPEC is complete, hand off to Prompt Architect."
```

### 3. Workflow Continues

The Forge workflow proceeds through all phases:

```
Agent Designer → Prompt Architect → Workflow Engineer → Platform Engineer → Eval Specialist → Agent Curator
```

Each agent hands off to the next when their handoff criteria are met.

### 4. Completion

When Agent Curator finishes, the rite is:
- Deployed to `$KNOSSOS_HOME/rites/{rite-name}/`
- Discoverable via `/consult`
- Ready for use via `/{rite-name}` or `/rite {rite-name}`

## Example Usage

```bash
# Create a new rite with default complexity
/new-rite api-dev

# Create a minimal agent modification
/new-rite security-auditor --complexity=PATCH

# Create a multi-rite ecosystem initiative
/new-rite observability-platform --complexity=ECOSYSTEM
```

## See Also

- `/forge-rite <name>` — Archaeology-first rite creation with domain expertise extraction
- Full documentation: `rites/forge/mena/forge-ref/INDEX.lego.md`
- Forge overview: `/forge`
