---
name: clinic
description: Quick switch to clinic (investigation and debugging workflow)
argument-hint: "[--overwrite-diverged] [--dry-run] [--keep-orphans]"
allowed-tools: Bash, Read
model: haiku
disable-model-invocation: true
context: fork
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the Clinic rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync --rite clinic $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `clinic`

## Flags

| Flag | Description |
|------|-------------|
| `--overwrite-diverged` | Force regeneration of diverged files |
| `--dry-run` | Preview changes without applying |
| `--keep-orphans` | Preserve orphaned knossos files (default: auto-remove) |

## When to Use

- Debugging production errors or failing tests
- Investigating intermittent failures or performance regressions
- Processing SRE escalations that need root cause analysis
- Any situation requiring structured evidence collection and diagnosis

## Rite Capabilities

This rite specializes in:
- Structured investigation with evidence collection and cataloging
- Hypothesis-driven root cause analysis (including compound failures)
- Cross-rite handoff to 10x-dev, SRE, and debt-triage

## Workflow Phases

```
intake -> examination -> diagnosis -> treatment
```

Back-routes: diagnosis->examination (evidence gap), treatment->diagnosis (insufficient), diagnosis->intake (scope expansion, requires user confirmation)

## Reference

Full documentation: `.claude/skills/clinic-ref/INDEX.md`
