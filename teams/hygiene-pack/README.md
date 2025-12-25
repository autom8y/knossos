# Hygiene Pack

Code quality lifecycle team for systematic cleanup and refactoring.

## When to Use This Team

**Triggers**:
- Codebase feels messy or inconsistent
- Dead code, unused imports, or complexity hotspots
- Need prioritized technical debt inventory
- Planning refactoring work with clear contracts
- Want atomic, reversible cleanup commits
- Code review before merging cleanup work

**Not for**: New features (behavior changes), ecosystem infrastructure (use ecosystem-pack), quick formatting fixes.

## Quick Start

```bash
/team hygiene-pack
```

## Agents

| Agent | Role | Model | Artifact |
|-------|------|-------|----------|
| code-smeller | Diagnose code quality issues | claude-sonnet-4-5 | Smell Report |
| architect-enforcer | Plan refactoring with contracts | claude-opus-4-5 | Refactoring Plan |
| janitor | Execute cleanup atomically | claude-sonnet-4-5 | Commit Stream |
| audit-lead | Verify behavior preservation | claude-sonnet-4-5 | Audit Report |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- SPOT: Single smell fix (execution, audit)
- MODULE: Module-level cleanup (full pipeline)
- CODEBASE: Full codebase hygiene (full pipeline)

## Related Teams

- **ecosystem-pack**: When smells reveal infrastructure issues
- **10x-dev-pack**: When cleanup unblocks feature work
