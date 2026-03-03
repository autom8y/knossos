# Artifact Glossary

Naming patterns and conventions for workflow artifacts.

---

## Artifact Types

### By Domain

| Domain | Artifact Types |
|--------|---------------|
| **Development** | prd, tdd, adr, code, test-plan |
| **Documentation** | audit-report, doc-structure, documentation, review-signoff |
| **Hygiene** | smell-report, refactor-plan, commits, audit-signoff |
| **Debt** | debt-ledger, risk-report, sprint-plan |
| **SRE** | observability-report, reliability-plan, infrastructure-changes, resilience-report |

### Type Naming Convention
- All lowercase
- Hyphenated for multi-word
- Describes what is produced, not the agent

```
prd              # Product Requirements Document
tdd              # Technical Design Document
adr              # Architecture Decision Record
smell-report     # Code smell analysis
resilience-report # Chaos engineering results
```

---

## Path Templates

### Pattern
```
.ledge/{category}/{PREFIX}-{slug}.md
```

### Components

| Component | Description | Examples |
|-----------|-------------|----------|
| `.ledge/` | Knossos work artifact root | Always `.ledge/` |
| `{category}` | Domain-specific subdirectory | specs, decisions, reviews, spikes |
| `{PREFIX}` | Artifact type abbreviation (uppercase) | PRD, TDD, ADR, OBS |
| `{slug}` | Dynamic name placeholder | user-auth, api-gateway |

### Examples by Rite

| Rite | Path Template | Example |
|------|--------------|---------|
| 10x-dev | `.ledge/specs/PRD-{slug}.md` | `.ledge/specs/PRD-user-auth.md` |
| 10x-dev | `.ledge/specs/TDD-{slug}.md` | `.ledge/specs/TDD-payment-api.md` |
| doc-rite | `.ledge/reviews/AUDIT-{slug}.md` | `.ledge/reviews/AUDIT-api-docs.md` |
| hygiene | `.ledge/reviews/SMELL-{slug}.md` | `.ledge/reviews/SMELL-legacy-utils.md` |
| debt | `.ledge/reviews/LEDGER-{slug}.md` | `.ledge/reviews/LEDGER-2024-q4.md` |
| sre | `.ledge/reviews/OBS-{slug}.md` | `.ledge/reviews/OBS-payment-svc.md` |

---

## Prefix Conventions

### Standard Prefixes

| Prefix | Full Name | Rite |
|--------|-----------|------|
| PRD | Product Requirements Document | 10x-dev |
| TDD | Technical Design Document | 10x-dev |
| ADR | Architecture Decision Record | 10x-dev |
| AUDIT | Audit Report | doc-rite |
| SMELL | Smell Report | hygiene |
| REFACTOR | Refactor Plan | hygiene |
| LEDGER | Debt Ledger | debt-triage |
| OBS | Observability Report | sre |
| REL | Reliability Plan | sre |
| CHAOS | Chaos Experiment | sre |

### Creating New Prefixes
- 3-6 uppercase letters
- Abbreviation of artifact name
- Unique within rite
- Clear meaning when seen in file listings

---

## Category Directories

### Standard Categories

| Category | Purpose | Rites |
|----------|---------|-------|
| `specs/` | PRDs, TDDs, and test plans | 10x-dev |
| `decisions/` | ADRs and design decisions | 10x-dev |
| `reviews/` | Audit reports, smell reports, debt ledgers | doc-rite, hygiene, debt-triage, sre |
| `spikes/` | Exploration and research artifacts | rnd, intelligence |

### Creating New Categories
- Lowercase, singular or plural (be consistent)
- Describes the artifact type, not the rite
- Create under `.ledge/`

---

## Artifact Quality Patterns

### Handoff Artifacts
Artifacts passed between phases should be:
- Self-contained (reader needs no external context)
- Actionable (next agent can work from it alone)
- Complete (all required sections filled)
- Validated (meets handoff criteria)

### Template Sections
Most artifacts include:

1. **Metadata** - Title, date, author, status
2. **Summary** - Executive overview
3. **Body** - Main content sections
4. **Decisions** - Key choices made
5. **Next Steps** - Handoff instructions
6. **References** - Links to related artifacts

---

## Artifact Lifecycle

```
Created → In Progress → Ready for Review → Approved → Archived
```

### Status Tracking
In artifact frontmatter:

```yaml
---
title: PRD-user-auth
status: approved  # draft | in-progress | ready-for-review | approved
author: requirements-analyst
created: 2024-01-15
approved: 2024-01-16
---
```

---

## Cross-References

### Linking Artifacts
Use relative paths:

```markdown
See [PRD-user-auth](../requirements/PRD-user-auth.md) for requirements.
```

### Artifact Chains
Track upstream/downstream:

```markdown
## Related Artifacts
- **Upstream**: [PRD-user-auth](../requirements/PRD-user-auth.md)
- **Downstream**: [Code Implementation](../../src/auth/)
```
