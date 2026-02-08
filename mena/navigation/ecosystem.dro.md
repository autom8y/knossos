---
name: ecosystem
description: Quick switch to ecosystem (CEM/skeleton/roster infrastructure workflow)
argument-hint: [--update] [--dry-run] [--keep-all|--remove-all|--promote-all]
allowed-tools: Bash, Read
model: sonnet
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the ecosystem infrastructure rite and display the pantheon. $ARGUMENTS

## Behavior

1. Execute: `ari sync materialize --rite ecosystem $ARGUMENTS`
2. Display the ari output (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_rite` to `ecosystem`

**Note**: If `ari` is not in PATH, use `~/bin/ari` or build with:
`CGO_ENABLED=0 go build -o ~/bin/ari ${KNOSSOS_HOME:-~/Code/knossos}/cmd/ari`

## Flags

| Flag | Short | Description | Handled By |
|------|-------|-------------|------------|
| `--update` | `-u` | Pull latest agent definitions from roster even if already on rite | ari sync materialize |
| `--dry-run` | - | Preview changes without applying | ari sync materialize |
| `--keep-all` | - | Preserve all orphan agents in project | ari sync materialize |
| `--remove-all` | - | Remove all orphans (backup available) | ari sync materialize |
| `--promote-all` | - | Move all orphans to user-level | ari sync materialize |

## When to Use

- CEM sync failures or conflicts
- Hook/skill registration issues
- Designing new hook or skill patterns
- Migrating satellites to new ecosystem versions
- Cross-satellite compatibility work
- Infrastructure bug fixes in CEM/skeleton/roster

## Workflow Phases

1. **Analysis** - Diagnose issues, trace root causes (Gap Analysis)
2. **Design** - Design hooks/skills/schemas (Context Design) [MODULE+]
3. **Implementation** - Code CEM/skeleton/roster changes
4. **Documentation** - Write migration runbooks [MODULE+]
5. **Validation** - Test across satellite matrix (Compatibility Report)

## Complexity Levels

- **PATCH**: Single file change, no schema impact (Analysis -> Implementation -> Validation)
- **MODULE**: Single system change (CEM or skeleton or roster) - all phases
- **SYSTEM**: Multi-system change - all phases + ADRs
- **MIGRATION**: Cross-satellite rollout - all phases + extended validation

## Reference

Full documentation: `.claude/skills/ecosystem-ref/INDEX.md`
