---
last_verified: 2026-03-26
---

# CLI Reference: lint

> Lint source files to catch errors before projection.

`ari lint` validates agents, dromena, and legomena before they are materialized. Run it before committing to catch frontmatter errors, archetype mismatches, and harness-specific language early.

**Family**: lint
**Commands**: 1 (with scope flags)
**Priority**: MEDIUM

---

## Synopsis

```bash
ari lint [flags]
```

## Description

Validates source files in `rites/`, `agents/`, and mena directories. Checks for:

- Missing or malformed frontmatter
- Required fields (name, description, etc.)
- Agent archetype mismatches (maxTurns, type, color)
- Dromena `context:fork` allowlist mismatches
- Dromena `context:fork` + Agent tool conflicts (SCAR-018)
- Workflow commands must be model-invocable
- Legomena missing Triggers keyword in description
- Preferential harness-specific language in Go source and mena content

## Subcommands

None. `ari lint` uses flags to control scope.

## Key Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--scope` | string | all | Limit to: `agents`, `dromena`, or `legomena` |
| `--check` | string | - | Run only a specific check: `preferential-language` |

## Examples

```bash
# Lint all sources
ari lint

# Lint agents only
ari lint --scope=agents

# Lint dromena only
ari lint --scope=dromena

# Lint legomena only
ari lint --scope=legomena

# Run only the preferential-language check
ari lint --check=preferential-language
```

## See Also

- [`ari agent validate`](cli-agent.md#ari-agent-validate) — Agent schema validation
- [`ari sync materialize`](cli-sync.md) — Materialization (runs after lint passes)
