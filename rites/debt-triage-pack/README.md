# Debt Triage Pack

Technical debt management lifecycle for cataloging, risk assessment, and sprint planning.

## When to Use This Rite

**Triggers**:
- "What technical debt do we have?"
- "How do we prioritize debt paydown?"
- "What's our biggest technical risk?"
- "Plan a debt cleanup sprint"
- "We inherited this codebase—what debt are we taking on?"
- "How much new debt did we accumulate this quarter?"

**Not for**: Feature development, active incidents, or ongoing reliability work

## Quick Start

```bash
/rite debt-triage-pack
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| debt-collector | Catalogs all forms of debt with precision | Debt Ledger |
| risk-assessor | Scores debt by blast radius, likelihood, effort | Risk Report |
| sprint-planner | Packages prioritized debt into actionable work units | Sprint Plan |

## Workflow

Three-phase sequential workflow: **Collection → Assessment → Planning**

**Complexity Levels**:
- **QUICK**: Known debt items (skip collection phase)
- **AUDIT**: Full debt discovery across codebase

See `workflow.md` for detailed phase flow and handoff criteria.

## Related Rites

- **sre-pack**: When debt manifests as reliability or infrastructure problems
