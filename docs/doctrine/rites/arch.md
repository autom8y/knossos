---
last_verified: 2026-02-26
---

# Rite: arch

> Architecture assessment and remediation lifecycle.

The arch rite provides workflows for evaluating codebase architecture — topology mapping, structural analysis, dependency auditing, and remediation planning.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | arch |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | pythia |

---

## When to Use

- Mapping codebase topology and module boundaries
- Evaluating structural health and coupling patterns
- Auditing dependency graphs for risk
- Planning architectural remediation work
- Assessing migration readiness

---

## Agents

| Agent | Role |
|-------|------|
| **pythia** | Coordinates architecture assessment phases |
| **topology-cartographer** | Maps codebase structure, module boundaries, and package relationships |
| **structure-evaluator** | Evaluates structural patterns, coupling, and cohesion |
| **dependency-analyst** | Audits dependency graphs and identifies risks |
| **remediation-planner** | Plans architecture improvements and migration paths |

See agent files: `rites/arch/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[topology] --> B[evaluation]
    B --> C[dependency-audit]
    C --> D[remediation]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| topology | topology-cartographer | Topology Map | Always |
| evaluation | structure-evaluator | Structure Report | Always |
| dependency-audit | dependency-analyst | Dependency Audit | complexity >= MODULE |
| remediation | remediation-planner | Remediation Plan | complexity >= MODULE |

---

## Invocation Patterns

```bash
# Quick switch to arch
/arch

# Map codebase topology
Task(topology-cartographer, "map module boundaries and package relationships")

# Evaluate structural health
Task(structure-evaluator, "assess coupling patterns in internal/ packages")

# Audit dependencies
Task(dependency-analyst, "audit dependency graph for circular imports")
```

---

## Source

**Manifest**: `rites/arch/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- [Knossos Doctrine - Rites](../philosophy/knossos-doctrine.md#iv-the-rites)
