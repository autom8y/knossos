# Hygiene Pack Workflow

## Phase Flow

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│     Code      │─────▶│   Architect   │─────▶│    Janitor    │─────▶│  Audit Lead   │
│   Smeller     │      │   Enforcer    │      │               │      │               │
└───────────────┘      └───────────────┘      └───────────────┘      └───────────────┘
 Smell Report        Refactor Plan            Commits             Audit Signoff
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| assessment | code-smeller | Smell Report | User request |
| planning | architect-enforcer | Refactor Plan | Smell report complete |
| execution | janitor | Commits | Refactor plan approved |
| audit | audit-lead | Audit Signoff | Commits complete |

## Complexity Levels

- **SPOT**: Single smell fix
  - Phases: execution, audit
- **MODULE**: Module-level cleanup
  - Phases: assessment, planning, execution, audit
- **CODEBASE**: Full codebase hygiene
  - Phases: assessment, planning, execution, audit

## Phase Skipping

At SPOT complexity, assessment and planning phases are skipped for immediate fixes.
