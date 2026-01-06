# Doc Team Pack Workflow

## Phase Flow

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│  Doc Auditor  │─────▶│  Information  │─────▶│     Tech      │─────▶│      Doc      │
│               │      │   Architect   │      │    Writer     │      │   Reviewer    │
└───────────────┘      └───────────────┘      └───────────────┘      └───────────────┘
  Audit Report          Doc Structure        Documentation        Review Signoff
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| audit | doc-auditor | Audit Report | User request |
| architecture | information-architect | Doc Structure | Audit complete, complexity >= SECTION |
| writing | tech-writer | Documentation | Structure approved (or direct if PAGE) |
| review | doc-reviewer | Review Signoff | Documentation complete |

## Complexity Levels

- **PAGE**: Single document
  - Phases: writing, review
- **SECTION**: Multiple related documents
  - Phases: architecture, writing, review
- **SITE**: Full documentation site
  - Phases: audit, architecture, writing, review

## Phase Skipping

At PAGE complexity, both audit and architecture phases are skipped. At SECTION complexity, only audit is skipped.
