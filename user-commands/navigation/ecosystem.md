---
description: Quick switch to ecosystem-pack (CEM/skeleton/roster infrastructure workflow)
allowed-tools: Bash, Read
model: claude-sonnet-4-5
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Switch to the ecosystem infrastructure team pack and display the team roster.

## Behavior

1. Execute: `${ROSTER_HOME:-~/Code/roster}/swap-team.sh ecosystem-pack`
2. Display the roster output from swap-team.sh (agents and their roles)
3. If SESSION_CONTEXT exists, update `active_team` to `ecosystem-pack`

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

- **PATCH**: Single file change, no schema impact (Analysis → Implementation → Validation)
- **MODULE**: Single system change (CEM or skeleton or roster) - all phases
- **SYSTEM**: Multi-system change - all phases + ADRs
- **MIGRATION**: Cross-satellite rollout - all phases + extended validation

## Reference

Full documentation: `.claude/skills/ecosystem-ref/skill.md`
