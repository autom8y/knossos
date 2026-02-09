# Hygiene Pack

Code quality lifecycle rite for systematic cleanup and refactoring.

## When to Use This Rite

**Triggers**:
- Codebase feels messy or inconsistent
- Dead code, unused imports, or complexity hotspots
- Need prioritized technical debt inventory
- Planning refactoring work with clear contracts
- Want atomic, reversible cleanup commits
- Code review before merging cleanup work

**Not for**: New features (behavior changes), ecosystem infrastructure (use ecosystem), quick formatting fixes.

## Quick Start

```bash
/rite hygiene
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| code-smeller | Diagnose code quality issues | Smell Report |
| architect-enforcer | Plan refactoring with contracts | Refactoring Plan |
| janitor | Execute cleanup atomically | Commit Stream |
| audit-lead | Verify behavior preservation | Audit Report |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity Levels**:
- SPOT: Single smell fix (execution, audit)
- MODULE: Module-level cleanup (full pipeline)
- CODEBASE: Full codebase hygiene (full pipeline)

## Related Rites

- **ecosystem**: When smells reveal infrastructure issues
- **10x-dev**: When cleanup unblocks feature work
