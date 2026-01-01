# Operations Shared Patterns

> Reusable patterns extracted from operations skills.

## Purpose

This directory contains patterns duplicated across multiple operations skills. Reference these partials from behavior.md files to maintain DRY principles.

## Available Patterns

| Pattern | Purpose | Used By |
|---------|---------|---------|
| [time-boxing.md](time-boxing.md) | Enforce time limits with progress checkpoints | spike-ref, hotfix-ref |
| [agent-invocation.md](agent-invocation.md) | Delegate to specialized agents via Task tool | spike-ref, hotfix-ref |
| [git-validation.md](git-validation.md) | Validate git repository state before operations | commit-ref, pr-ref |

## Usage Convention

Reference patterns from behavior.md using this format:

```markdown
Apply [Pattern Name](../shared-sections/pattern.md):
- Requirement: {specific requirement}
- Verb: "{command verb}"
```

## Adding New Patterns

Extract to shared-sections when:
1. Pattern appears in 2+ behavior.md files
2. Error messages should be consistent across commands
3. Validation logic is identical across commands
4. Schema reference is needed by multiple skills
