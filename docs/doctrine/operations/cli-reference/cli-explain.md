---
last_verified: 2026-03-26
---

# CLI Reference: explain

> Look up definitions for knossos domain concepts.

`ari explain` provides project-aware definitions for knossos terminology. Use it when you encounter an unfamiliar term or want to understand a concept in the context of your current project.

**Family**: explain
**Commands**: 1
**Priority**: MEDIUM

---

## Synopsis

```bash
ari explain [concept] [flags]
```

## Description

Without a concept argument, lists all known concepts with one-line summaries. With a concept name, shows the full definition with project-aware context.

Covers terms like: rite, session, agent, mena, dromena, legomena, and more.

## Subcommands

None. `ari explain` is a single command.

## Key Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-h, --help` | bool | false | Show command help |

## Examples

```bash
# List all concepts
ari explain

# Get the full definition of "rite"
ari explain rite

# JSON output — all concepts
ari explain -o json

# JSON output — single concept
ari explain rite -o json
```

## See Also

- [GLOSSARY.md](../../reference/GLOSSARY.md) — Complete terminology reference
- [`ari ask`](cli-ask.md) — Natural language query for commands and workflows
