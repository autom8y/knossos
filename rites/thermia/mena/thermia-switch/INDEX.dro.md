---
name: thermia
description: Start a cache architecture consultation
argument-hint: "[--overwrite-diverged] [--dry-run] [--keep-orphans]"
allowed-tools: Bash, Read
model: haiku
disable-model-invocation: true
context: fork
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Thermia rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync --rite thermia $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `thermia`

## Flags

| Flag | Description |
|------|-------------|
| `--overwrite-diverged` | Force regeneration of diverged files |
| `--dry-run` | Preview changes without applying |
| `--keep-orphans` | Preserve orphaned knossos files (default: auto-remove) |

## When to Use

- Evaluating whether a system needs caching (or something else entirely)
- Designing cache layers for a new or existing service
- Post-mortem analysis when caching is causing production problems
- Auditing existing cache implementations for anti-patterns
- Capacity planning and eviction policy selection for cache infrastructure
- Any question that starts with "should I cache..." or "our cache is..."

## Rite Capabilities

This rite specializes in:
- Consultative cache architecture with alternatives-first analysis
- 6-gate decision framework for validating cache necessity
- Pattern selection grounded in distributed systems theory (CAP, consistency models)
- Working set analysis with derived capacity sizing (not gut-feel numbers)
- Observability design with miss-rate-first alerting and failure mode runbooks

## Complexity Levels

| Level | Trigger | Phases |
|-------|---------|--------|
| **QUICK** | "Should I cache this?", single yes/no question | assessment -> validation (lite) |
| **STANDARD** | New cache design, existing cache review | assessment -> architecture -> specification -> validation |
| **DEEP** | Post-mortem, production crisis, full redesign | Same as STANDARD at extended depth |

Pythia determines complexity from the user's problem statement.

## Workflow Phases

```
assessment -> architecture -> specification -> validation
(heat-mapper)  (systems-thermodynamicist)  (capacity-engineer)  (thermal-monitor)
```

QUICK mode: assessment -> validation (lite)

Back-routes: architecture->assessment (assessment gap, max 1), validation->specification (design inconsistency, max 1)

## Quick-Switch Commands

- `/heat-map` -- invoke heat-mapper directly for assessment-only work
- `/cache-design` -- invoke systems-thermodynamicist for architecture work
