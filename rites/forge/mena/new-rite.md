---
description: Create a new rite through The Forge workflow
argument-hint: "<rite-name> [--complexity=PATCH|RITE|ECOSYSTEM] [--deep] [--interview]"
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
  - RITE: Full rite with 3-5 agents (all 7 phases)
  - ECOSYSTEM: Multi-rite initiative (all 7 phases)
- `--deep`: Optional. Enables the archaeology phase after design. The domain-forensics agent runs 6-pass codebase analysis (scar tissue, defensive patterns, design tensions, golden paths, synthesis) to produce HANDOFF-PROMPT-FUEL for expert-level agent prompts. Required for ECOSYSTEM complexity.
- `--interview`: Optional. Requires `--deep`. Adds Pass 5 (tribal knowledge interview) where the domain-forensics agent conducts a structured interview with the user to extract jurisdiction boundaries, priorities, and unwritten rules.

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
Deep archaeology: {--deep flag present? yes/no}

When the RITE-SPEC is complete, hand off to {domain-forensics if --deep, else Prompt Architect}."
```

### 2b. Archaeology Phase (--deep only)

When `--deep` is present, Pythia routes to the domain-forensics agent after design:

```
Use the Task tool to invoke the domain-forensics agent with:

"Run codebase archaeology against the target codebase for the {rite-name} rite.

RITE-SPEC location: .claude/wip/RITE-SPEC-{rite-name}.md
Target codebase: {the codebase the rite will operate on}
Interview mode: {--interview flag present? yes/no}

Execute passes per the codebase-archaeology skill. Produce HANDOFF-PROMPT-FUEL.md
for the Prompt Architect."
```

### 3. Workflow Continues

The Forge workflow proceeds through all phases:

```
Agent Designer → [Domain Forensics] → Prompt Architect → Workflow Engineer → Platform Engineer → Eval Specialist → Agent Curator
```

The `[Domain Forensics]` phase runs only when `--deep` is present. Without it, the workflow skips directly from design to prompts (current behavior).

Each agent hands off to the next when their handoff criteria are met.

### 4. Completion

When Agent Curator finishes, the rite is:
- Deployed to `$KNOSSOS_HOME/rites/{rite-name}/`
- Discoverable via `/consult`
- Ready for use via `/{rite-name}` or `/rite {rite-name}`

## Example Usage

```bash
# Create a new rite (quick, no archaeology)
/new-rite api-dev

# Create with codebase archaeology for expert-level agents
/new-rite data-analyst --deep

# Create with archaeology + domain expert interview
/new-rite data-analyst --deep --interview

# Create a minimal agent modification
/new-rite security-auditor --complexity=PATCH

# Create a multi-rite ecosystem initiative (--deep is required)
/new-rite observability-platform --complexity=ECOSYSTEM --deep
```

## Reference

Full documentation: `rites/forge/mena/forge-ref/INDEX.lego.md`
Forge overview: `/forge`
