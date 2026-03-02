---
name: forge-ref-agents
description: "Detailed agent profiles for The Forge rite. Use when: understanding what each Forge agent does, their domain boundaries, handoff criteria. Triggers: agent designer, prompt architect, workflow engineer, platform engineer, eval specialist, agent curator."
---

# The Forge: Agent Profiles

## Agent Designer (Entry Point)

**Purpose**: Creates rite specifications and role definitions.

**Domain**:
- Rite purpose and scope
- Agent role boundaries
- Input/output contracts
- Complexity level design

**Produces**: RITE-SPEC.md with all roles defined

**Handoff**: When RITE-SPEC is complete and approved

---

## Prompt Architect

**Purpose**: Crafts system prompts for agents.

**Domain**:
- Agent identity and personality
- Instruction clarity and constraints
- Token efficiency
- Example creation

**Produces**: Complete agent .md files with 11 sections

**Handoff**: When all agents have complete prompts

---

## Workflow Engineer

**Purpose**: Designs orchestration and commands.

**Domain**:
- Phase sequencing
- workflow.yaml configuration
- Slash command creation
- Complexity gating

**Produces**: workflow.yaml and command files

**Handoff**: When workflow is complete and validates

---

## Platform Engineer

**Purpose**: Implements knossos infrastructure.

**Domain**:
- Directory structure creation
- File deployment
- ari sync --rite integration
- Structure validation

**Produces**: Rite deployed to knossos

**Handoff**: When ari sync --rite loads successfully

---

## Eval Specialist

**Purpose**: Validates rites before shipment.

**Domain**:
- Completeness checks
- Schema validation
- Logic validation
- Adversarial testing

**Produces**: eval-report.md with pass/fail

**Handoff**: When all validations pass

---

## Agent Curator (Terminal)

**Purpose**: Finalizes integration and documentation.

**Domain**:
- Consultant knowledge sync
- Rite profile creation
- Version recording
- Documentation

**Produces**: Catalog entry + Consultant sync

**Terminal**: Workflow completes here
