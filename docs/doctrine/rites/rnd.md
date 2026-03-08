---
last_verified: 2026-02-26
---

# Rite: rnd

> Technology exploration lifecycle for scouting, prototyping, and future architecture.

The R&D rite provides workflows for exploring emerging technologies and building proof-of-concept prototypes.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | rnd |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 6 |
| **Entry Agent** | potnia |

---

## When to Use

- Scouting emerging technologies
- Evaluating build vs buy decisions
- Building proof-of-concept prototypes
- Designing long-term architecture
- Technology transfer to production

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates technology exploration phases |
| **technology-scout** | Scouts emerging technologies and provides build vs buy analysis |
| **integration-researcher** | Maps integration dependencies and assesses compatibility |
| **prototype-engineer** | Builds proof-of-concept prototypes with deliberate shortcuts |
| **moonshot-architect** | Designs visionary long-term architecture plans |
| **tech-transfer** | Facilitates technology transfer from R&D to production teams |

See agent files: `rites/rnd/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[scouting] --> B[integration-analysis]
    B --> C[prototyping]
    C --> D[future-architecture]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| scouting | technology-scout | Tech Assessment | Always |
| integration-analysis | integration-researcher | Integration Map | complexity >= EVALUATION |
| prototyping | prototype-engineer | Prototype | Always |
| future-architecture | moonshot-architect | Moonshot Plan | Always |

---

## Invocation Patterns

```bash
# Quick switch to R&D
/rnd

# Technology scouting
Task(technology-scout, "evaluate vector databases for semantic search")

# Build prototype
Task(prototype-engineer, "prototype AI-powered code review")

# Moonshot architecture
Task(moonshot-architect, "design architecture for multi-agent orchestration")
```

---

## Source

**Manifest**: `rites/rnd/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- [Spike Workflow](/spike)
