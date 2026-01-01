# Agent Invocation Pattern

> Delegate to specialized agents via Task tool.

## When to Apply

- `/spike`: Research delegation
- `/hotfix`: Diagnose/fix delegation

## Template Structure

```markdown
Act as **{Agent Name}**.

{MODE} ({context})
{Key parameter}: {value}
Time budget: {duration}

{Instructions}:
1. Step 1
2. Step 2
...

Deliverable: {artifact type}
Save to: {path}
```

## Agent Selection

| Research Type | Agent | Mode |
|---------------|-------|------|
| Architecture/Design | Architect | SPIKE MODE |
| Feasibility | Principal Engineer | SPIKE MODE |
| Technology comparison | Architect | SPIKE MODE |
| Hotfix diagnose/fix | Principal Engineer | HOTFIX MODE |
| Hotfix validation | QA Adversary | HOTFIX VALIDATION |

## Customization Points

| Parameter | Description | Commands |
|-----------|-------------|----------|
| Agent Name | Target specialist | spike, hotfix |
| MODE | Context label | spike, hotfix |
| Time budget | Time limit | spike, hotfix |
| Deliverable | Output type | spike, hotfix |
| Save path | Artifact location | spike, hotfix |

## Example Invocations

### Spike Research
```markdown
Act as **Architect**.

SPIKE MODE (Time-boxed research)
Question: Can we use GraphQL instead of REST?
Time budget: 4h

Research and document findings:
1. Understand the question/problem
2. Research options (libraries, patterns, approaches)
3. Build proof of concept if needed (throwaway code)
4. Document findings with pros/cons
5. Provide recommendation

Deliverable: Spike report
Save to: /docs/research/SPIKE-graphql-vs-rest.md
```

### Hotfix Delegation
```markdown
Act as **Principal Engineer**.

HOTFIX MODE (Rapid resolution)
Issue: Authentication failing for OAuth users
Time budget: 60 min
Severity: CRITICAL

Diagnose and fix:
1. Reproduce issue
2. Identify root cause
3. Implement minimal fix
4. Verify fix resolves issue

Deliverable: Fix implementation + verification
```
