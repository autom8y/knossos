---
name: forge
description: Quick switch to forge (meta-rite for building and maintaining rites)
argument-hint: [--force] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: haiku
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Switch to the forge rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite forge $ARGUMENTS`
2. Display the pantheon output from ari (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `forge`

After the switch, display this condensed overview:

```
THE FORGE - Agent Factory Rite
==============================

The rite that builds rites. Global singleton (always available).

AGENTS (6):
  Agent Designer    - Role specs and contracts
  Prompt Architect  - System prompts (11 sections)
  Workflow Engineer - Orchestration and commands
  Platform Engineer - Roster infrastructure
  Eval Specialist   - Testing and validation
  Agent Curator     - Versioning and integration

COMMANDS:
  /new-rite <name>      - Full rite creation workflow
  /validate-rite <name> - Run validation suite on rite
  /eval-agent <name>    - Test single agent in isolation

Full docs: rites/forge/mena/forge-ref/INDEX.lego.md
```

## Flags

| Flag | Description |
|------|-------------|
| `--force` | Force regeneration even if already on this rite |
| `--dry-run` | Preview changes without applying |
| `--keep-all` | Preserve all orphan agents in project (default) |
| `--remove-all` | Remove orphans (backup in `.claude/.orphan-backup/`) |
| `--promote-all` | Move orphans to user-level (`~/.claude/agents/`) |

## When to Use

- Creating or modifying rites
- Designing new agents or workflows
- Testing and validating agent behavior
- Any meta-work on the Knossos platform itself

## Reference

Full documentation: `rites/forge/mena/forge-ref/INDEX.lego.md`
