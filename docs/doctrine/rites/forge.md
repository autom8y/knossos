---
last_verified: 2026-03-26
---

# Rite: forge

> Rite creation lifecycle for platform development.

The forge rite is the meta-rite for building agents, tools, and platform infrastructure. Named after [Daedalus](../reference/GLOSSARY.md#daedalus), the builder of the labyrinth.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | forge |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 8 |
| **Entry Agent** | potnia |

---

## When to Use

- Creating new rites
- Building platform tools
- Designing new agent pantheons
- Extending knossos infrastructure
- Evaluating agent performance

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates rite creation phases |
| **agent-designer** | Designs rite concepts and agent role specifications |
| **prompt-architect** | Creates agent prompt files and system instructions |
| **workflow-engineer** | Configures workflow phases and transitions |
| **platform-engineer** | Integrates agents into knossos ecosystem |
| **agent-curator** | Updates knowledge base and documentation |
| **domain-forensics** | Forensic analysis of domain patterns and constraints |
| **eval-specialist** | Evaluates and validates agent pantheon readiness |

See agent files: `rites/forge/agents/`

---

## Workflow Phases

```mermaid
flowchart TD
    A[design] --> B[prompts]
    B --> C[workflow]
    C --> D[platform]
    D --> E[catalog]
    E --> F[validation]
    F --> G[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| design | agent-designer | Rite Spec | Always |
| prompts | prompt-architect | Agent Files | Always |
| workflow | workflow-engineer | Workflow Config | Always |
| platform | platform-engineer | Knossos Integration | complexity >= MODULE |
| catalog | agent-curator | Knowledge Update | complexity >= MODULE |
| validation | eval-specialist | Eval Report | Always |

---

## Architecture

The forge creates new rites following this structure:

```
rites/{new-rite}/
├── manifest.yaml          # Rite composition
├── agents/                # Agent prompt files
│   ├── potnia.md
│   └── {specialists}.md
└── mena/                  # Rite-specific commands and knowledge
    ├── {command}.dro.md   # Dromena (transient commands)
    └── {knowledge}.lego.md # Legomena (persistent knowledge)
```

```mermaid
flowchart TD
    subgraph Design Phase
        A[Concept] --> B[Role Definitions]
        B --> C[Workflow Design]
    end

    subgraph Build Phase
        C --> D[Agent Prompts]
        D --> E[Manifest YAML]
        E --> F[Integration]
    end

    subgraph Validate Phase
        F --> G[Eval Suite]
        G --> H{Pass?}
        H -->|Yes| I[Catalog]
        H -->|No| D
    end
```

---

## Invocation Patterns

```bash
# Quick switch to forge
/forge

# Create new rite
Task(potnia, "create new rite for code review workflow")

# Design phase only
Task(agent-designer, "design rite for security auditing")

# Evaluate existing rite
Task(eval-specialist, "evaluate 10x-dev rite performance")
```

---

## Complexity Levels

| Level | Scope | Platform Phase |
|-------|-------|---------------|
| SIMPLE | Skills-only rite | Skipped |
| STANDARD | Single-agent rite | Skipped |
| FULL | Multi-agent rite | Required |
| ECOSYSTEM | Cross-rite integration | Required |

---

## Source

**Manifest**: `rites/forge/manifest.yaml`

---

## See Also

- [Daedalus Glossary Entry](../reference/GLOSSARY.md#daedalus)
- [Rite System Overview](../philosophy/knossos-doctrine.md)
- `/orchestrator-templates` skill — Orchestrator prompt templates
