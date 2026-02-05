# /start Agent Integration

> Task tool delegation templates for /start.

## Requirements Analyst Delegation

Always invoked for all complexity levels:

```markdown
Act as **Requirements Analyst**.

Initiative: {initiative-name}
Complexity: {complexity}

Create a PRD following the template at `.claude/skills/documentation/templates/prd.md`.

Clarify any ambiguities with the user before drafting. When complete, save to:
`/docs/requirements/PRD-{initiative-slug}.md`
```

## Architect Delegation

Invoked when complexity > SCRIPT (MODULE, SERVICE, PLATFORM):

```markdown
Act as **Architect**.

Initiative: {initiative-name}
PRD Location: /docs/requirements/PRD-{slug}.md

Create TDD following template at `.claude/skills/documentation/templates/tdd.md`.

Identify architecture decisions and create ADRs using template at `.claude/skills/documentation/templates/adr.md`.

When complete, save:
- TDD to: /docs/design/TDD-{slug}.md
- ADRs to: /docs/decisions/ADR-{NNNN}-{decision-slug}.md
```

## Complexity → Agent Matrix

| Complexity | Agents Invoked | Artifacts Produced |
|------------|----------------|-------------------|
| SCRIPT | Requirements Analyst | PRD only |
| MODULE | Requirements Analyst → Architect | PRD, TDD, ADRs |
| SERVICE | Requirements Analyst → Architect | PRD, TDD, ADRs |
| PLATFORM | Requirements Analyst → Architect | PRD, TDD, multiple ADRs |

## Notes

- All agent invocation happens via Claude Code's native Task tool
- No direct shell execution of agent files
- Agents are defined in `.claude/agents/{agent-name}.md`
- Templates are in `.claude/skills/documentation/templates/`
