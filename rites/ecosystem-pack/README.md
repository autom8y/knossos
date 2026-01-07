# Ecosystem Pack

Infrastructure lifecycle team for CEM and roster ecosystem maintenance.

## When to Use This Rite

**Triggers**:
- Satellite sync failures or integration errors
- Hook/skill/agent registration not working
- Need to design new infrastructure patterns
- CEM/roster bugs need diagnosis and fixes
- Breaking changes require migration planning
- Cross-satellite compatibility issues

**Not for**: Application code in satellites (use 10x-dev-pack), rite-specific workflows (use rite-pack).

## Quick Start

```bash
/rite ecosystem-pack
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| ecosystem-analyst | Diagnose ecosystem issues | Gap Analysis |
| context-architect | Design infrastructure solutions | Context Design |
| integration-engineer | Implement CEM/roster changes | Working Implementation |
| documentation-engineer | Write migration runbooks | Migration Runbook |
| compatibility-tester | Validate across satellite matrix | Compatibility Report |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- PATCH: Single file/config change (analysis, implementation, validation)
- MODULE: Single system change (full pipeline)
- SYSTEM: Multi-system change (full pipeline)
- MIGRATION: Cross-satellite rollout (full pipeline with coordination)

## Related Rites

- **hygiene-pack**: When cleanup reveals ecosystem bugs
- **rite-pack**: When rite-specific infrastructure needs changes
