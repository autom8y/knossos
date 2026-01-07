# Doc Team Pack

Documentation lifecycle management from content audit through quality review.

## When to Use This Rite

**Triggers**:
- "Our documentation is scattered and inconsistent"
- "We need to document this feature/system/API"
- "Audit what documentation exists before we write more"
- "Reorganize our docs so engineers can actually find things"

**Not for**: Code implementation, infrastructure work, writing code comments

## Quick Start

```bash
/task audit and improve documentation for authentication system
# or invoke directly
/rite docs
```

## Agents

| Agent | Role | Artifact |
|-------|------|----------|
| doc-auditor | Inventory existing docs, identify gaps | Audit report |
| information-architect | Design doc structure and taxonomy | Doc structure |
| tech-writer | Write clear, accessible documentation | Documentation |
| doc-reviewer | Verify technical accuracy against code | Review signoff |

## Workflow

See `workflow.md` for phase flow and complexity levels.

**Complexity levels**:
- PAGE: Single document (skips audit and architecture)
- SECTION: Multiple related documents (skips audit)
- SITE: Full documentation site

## Related Rites

- **10x-dev**: When code implementation produces documentation needs
- **api-rite-pack**: When API documentation needs specialized treatment
