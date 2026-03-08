---
name: forge-ref
description: |
  Reference documentation for The Forge - the meta-rite for creating and maintaining
  agent rites. Use when: learning about rite creation, understanding the Forge workflow,
  invoking Forge commands. Triggers: /forge, /new-rite, rite creation, agent factory,
  build rite, create agents.
---

# The Forge Reference

> The rite that builds rites. Meta-level agent factory for the Claude Code ecosystem.

## Commands

| Command | Purpose | Entry Agent |
|---------|---------|-------------|
| `/forge` | Display Forge overview and help | (info only) |
| `/new-rite <name>` | Create a new rite | Agent Designer |
| `/validate-rite <name>` | Run validation on rite | Eval Specialist |
| `/eval-agent <name>` | Test single agent | Eval Specialist |

## Agent Pantheon

| Agent | Model | Produces |
|-------|-------|----------|
| **Agent Designer** | opus | RITE-SPEC, role definitions |
| **Prompt Architect** | opus | Agent .md files (11 sections) |
| **Workflow Engineer** | opus | workflow.yaml, commands |
| **Platform Engineer** | sonnet | Rite catalog files, directory structure |
| **Eval Specialist** | opus | eval-report.md, test results |
| **Agent Curator** | sonnet | Catalog entry, Pythia sync |

## Workflow

```
Agent Designer → Prompt Architect → Workflow Engineer → Platform Engineer → Eval Specialist → Agent Curator
```

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| **AGENT** | Single agent modification | design, prompts, validation |
| **MODULE** | New rite with 3-5 agents | All 6 phases |
| **SYSTEM** | Multi-rite initiative | All 6 phases |

## Knowledge Base

| Type | Location |
|------|----------|
| Patterns | `patterns/` — role-definition, domain-authority, handoff-criteria |
| Evals | `evals/` — agent-completeness, workflow-validity, integration-tests |

## Companion Reference

| Topic | File | When to Load |
|-------|------|-------------|
| Agent profiles (full detail) | `agents.lego.md` | Understanding agent domains/handoffs |
| Best practices | `best-practices.lego.md` | Designing rites, writing prompts |
| Troubleshooting | `troubleshooting.lego.md` | Diagnosing sync/validation failures |
| Architecture | `architecture.lego.md` | Understanding Forge structure and availability |

## Related Resources

- `rite-development` skill — Manual rite creation guidance
- `10x-workflow` skill — Workflow patterns
- `/consult` command — Ecosystem navigation
