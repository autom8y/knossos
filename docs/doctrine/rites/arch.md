---
last_verified: 2026-02-26
---

# Rite: arch

> Architecture assessment and remediation lifecycle.

The arch rite is a structured architectural assessment that reads the codebase as it actually exists, not as the team imagines it. It starts by mapping topology — service boundaries, tech stacks, API surfaces — before passing that inventory to structure-evaluator, who looks for the pathologies that accumulate invisibly: distributed monoliths, god services, circular dependencies, boundary drift. What distinguishes arch from a code review or debt-triage pass is the progression: topology first, structural patterns second, dependency risk third. Each phase produces an artifact the next phase reads, so findings have evidence chains rather than opinions. The remediation-planner does not output a wish list — it sequences improvements by blast radius, producing a plan the next sprint can act on.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | arch |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | potnia |

---

## When to Use

- Mapping service boundaries and package relationships before a large refactor or migration
- Identifying structural anti-patterns: distributed monoliths, circular imports, god services, boundary drift
- Auditing the dependency graph for transitive risk before a major dependency upgrade
- Producing a prioritized remediation plan when the codebase "feels unhealthy" but the problems aren't named
- **Not for**: active bug investigation — use clinic. Not for code quality cleanup — use hygiene. Arch targets structural patterns, not individual smell instances.

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates architecture assessment phases; gates structural evaluation on completed topology inventory |
| **topology-cartographer** | Discovers and catalogs service types, tech stacks, and API surfaces across repos or modules; produces topology-inventory before any analysis begins |
| **structure-evaluator** | Identifies structural anti-patterns (distributed monoliths, circular deps, god services) and scores boundary alignment against actual coupling data |
| **dependency-analyst** | Builds the dependency graph, detects version mismatches and circular imports, calculates transitive blast radius for each high-risk node |
| **remediation-planner** | Sequences improvements by risk, producing a phased remediation plan with before/after contracts and rollback boundaries |

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

# Map topology across multiple repos — give cartographer explicit paths
Task(topology-cartographer, "map the topology of ~/code/acme-* — catalog service types, tech stacks, and API surfaces per repo")

# Evaluate structural health after topology is complete
Task(structure-evaluator, "assess coupling patterns and boundary alignment using topology-inventory at .ledge/")

# Audit dependency graph for transitive risk
Task(dependency-analyst, "audit dependency graph for circular imports and version mismatch blast radius")

# Plan remediation sequenced by risk
Task(remediation-planner, "produce phased remediation plan from structure-assessment and dependency-audit")
```

---

## Source

**Manifest**: `rites/arch/manifest.yaml`

---

## See Also

- [CLI: rite](../operations/cli-reference/cli-rite.md)
- [Knossos Doctrine - Rites](../philosophy/knossos-doctrine.md#iv-the-rites)
