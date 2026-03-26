---
last_verified: 2026-02-26
---

# Rite: 10x-dev

> Full development lifecycle from requirements through validation.

The 10x-dev rite is the production feature delivery engine — requirements to running code, with every specialist playing a distinct role in sequence. It does not assume the request is well-understood: requirements-analyst begins by treating every input as ambiguous, turning stakeholder language into a PRD with measurable acceptance criteria and explicit scope boundaries before any design or code begins. This separates 10x-dev from a simple coding workflow: the rite enforces that ambiguity is resolved on paper, not in code. For PATCH-level changes the design phase is skipped; for everything MODULE and above, architect produces a TDD that qa-adversary will later test against — not as a formality, but adversarially, probing for gaps between spec and implementation.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | 10x-dev |
| **Form** | Full (multi-agent workflow) |
| **Agents** | 5 |
| **Entry Agent** | potnia |

---

## When to Use

- Implementing a new feature where requirements are vague or disputed — requirements-analyst elicits the real spec before code starts
- Building anything MODULE or larger that needs an explicit technical design before implementation
- Work where you need a QA adversary to test against the spec, not just verify the code compiles
- Multi-task sprints where several features share a sprint boundary — use `/sprint` to parallelize
- **Not for**: bug fixes, minor changes, or one-off patches — use clinic or a direct `/task` for those

---

## Agents

| Agent | Role |
|-------|------|
| **potnia** | Coordinates development lifecycle phases; gates each transition on prior-phase output quality |
| **requirements-analyst** | Transforms vague stakeholder requests into PRDs with measurable acceptance criteria, MoSCoW priorities, and explicit out-of-scope statements |
| **architect** | Evaluates competing design tradeoffs and produces TDDs and ADRs — decides component boundaries, not just documents them |
| **principal-engineer** | Implements code against the approved TDD; flag deviations rather than silently resolve them |
| **qa-adversary** | Tests against the PRD acceptance criteria adversarially — probes edge cases and boundary conditions, not just happy paths |

See agent files: `rites/10x-dev/agents/`

---

## Workflow Phases

```mermaid
flowchart LR
    A[requirements] --> B[design]
    B --> C[implementation]
    C --> D[validation]
    D --> E[complete]
```

| Phase | Agent | Produces | Condition |
|-------|-------|----------|-----------|
| requirements | requirements-analyst | PRD | Always |
| design | architect | TDD | complexity >= MODULE |
| implementation | principal-engineer | Code | Always |
| validation | qa-adversary | Test Plan | Always |

---

## Invocation Patterns

```bash
# Full task lifecycle — potnia drives all four phases
/task "implement feature X"

# Multi-task sprint — parallelizes features within a sprint boundary
/sprint "Authentication Sprint" --tasks="Login,Logout,Session"

# Start from orchestrator with explicit context
Task(potnia, "implement user authentication — PRD scope: email+password only, OAuth deferred to next sprint")

# Skip to implementation when PRD and TDD already exist
Task(principal-engineer, "implement the notification service per TDD at .ledge/specs/notification-tdd.md")
```

---

## Complexity Levels

| Level | Scope | Design Phase |
|-------|-------|--------------|
| PATCH | Bug fix, minor change | Skipped |
| MODULE | Single component | Required |
| SYSTEM | Multi-component | Required |
| INITIATIVE | Major feature | Required |
| MIGRATION | Breaking change | Required |

---

## Skills

- `10x-ref` — Workflow reference
- `10x-workflow` — Phase orchestration

---

## Source

**Manifest**: `rites/10x-dev/manifest.yaml`

---

## See Also

- [CLI: session](../operations/cli-reference/cli-session.md) — Session management
- [CLI: rite](../operations/cli-reference/cli-rite.md) — Rite operations
- [Rite Catalog](index.md)
