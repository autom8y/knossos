---
name: forge-rite
description: Create a new rite with codebase archaeology (domain expertise extraction first)
argument-hint: "<rite-name> [--complexity=MODULE|SYSTEM] [--interview]"
allowed-tools: Bash, Glob, Grep, Read, Write, Edit, Task, TodoWrite
model: opus
---

## Context

Auto-injected by SessionStart hook (project, rite, session context).

## Your Task

Create a new rite with domain expertise extraction: $ARGUMENTS

## Behavior

### 1. Parse Arguments

- `rite-name`: Required. Name for the new rite
- `--complexity`: Optional. Default is MODULE.
  - MODULE: New rite with 3-5 agents (all phases including archaeology)
  - SYSTEM: Major rite redesign or cross-rite refactor (all phases)
  - PATCH is not available for `/forge-rite` — archaeology requires MODULE or higher. If the user passes `--complexity=PATCH`, respond with:
    > PATCH complexity is for single-agent modifications. Use `/new-rite <name> --complexity=PATCH` instead.
- `--interview`: Optional. Adds Pass 5 (tribal knowledge interview) where the domain-forensics agent conducts a structured interview with the user to extract jurisdiction boundaries, priorities, and unwritten rules.

### 2. Check Codebase Knowledge Freshness

Before archaeology, check whether the target codebase has fresh .know/ domains that can accelerate forensic analysis:

1. Look for `.know/` directory in the target codebase root
2. If it exists, read frontmatter of any `.know/*.md` files and check freshness:
   - Parse `generated_at` + `expires_after` for time-freshness
   - Compare `source_hash` to current `git rev-parse --short HEAD` for code-freshness
3. Report status:
   - If ALL relevant domains (scar-tissue, defensive-patterns, design-constraints) are fresh: "Codebase knowledge is current. Domain forensics will use .know/ as seed context."
   - If SOME are stale or missing: "Stale or missing codebase knowledge detected: {list}. Consider running `/know --all` to refresh before archaeology."
   - If `.know/` does not exist: "No codebase knowledge found. Consider running `/know --all` first for faster archaeology."
4. This check is advisory only -- the forge continues regardless of freshness status.

### 3. Invoke Agent Designer

Start the Forge workflow by invoking the Agent Designer:

```
Use the Task tool to invoke the agent-designer agent with:

"Create a RITE-SPEC for {rite-name}.

The user wants to create a new rite with domain expertise extraction.
Your job is to:
1. Clarify the rite's purpose with the user
2. Design 3-5 agent roles with clear boundaries
3. Define input/output contracts
4. Specify complexity levels
5. Document handoff criteria

Complexity level: {complexity}

When the RITE-SPEC is complete, hand off to Domain Forensics for codebase archaeology."
```

### 4. Archaeology Phase

After design, the domain-forensics agent runs 6-pass codebase archaeology:

```
Use the Task tool to invoke the domain-forensics agent with:

"Run codebase archaeology against the target codebase for the {rite-name} rite.

RITE-SPEC location: .claude/wip/RITE-SPEC-{rite-name}.md
Target codebase: {the codebase the rite will operate on}
Interview mode: {--interview flag present? yes/no}

Execute passes per the codebase-archaeology skill. Produce HANDOFF-PROMPT-FUEL.md
for the Prompt Architect."
```

This produces HANDOFF-PROMPT-FUEL.md containing per-agent domain knowledge tiered as CRITICAL / IMPORTANT / CONTEXTUAL. The Prompt Architect consumes this to write expert-level agent prompts.

### 5. Workflow Continues

The Forge workflow proceeds through all phases:

```
Agent Designer → Domain Forensics → Prompt Architect → Workflow Engineer → Platform Engineer → Eval Specialist → Agent Curator
```

Each agent hands off to the next when their handoff criteria are met.

### 6. Completion

When Agent Curator finishes, the rite is:
- Deployed to `$KNOSSOS_HOME/rites/{rite-name}/`
- Discoverable via `/consult`
- Ready for use via `/{rite-name}` or `/rite {rite-name}`

## Example Usage

```bash
# Create a rite with domain expertise extraction
/forge-rite data-analyst

# Create with tribal knowledge interview
/forge-rite data-analyst --interview

# Create a major ecosystem initiative with interview
/forge-rite observability-platform --complexity=SYSTEM --interview
```

## See Also

- `/new-rite <name>` — Direct rite creation without archaeology
- Full documentation: `.claude/skills/forge-ref/INDEX.md`
- Forge overview: `/forge`
