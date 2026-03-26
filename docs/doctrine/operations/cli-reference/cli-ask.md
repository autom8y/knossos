---
last_verified: 2026-03-26
---

# CLI Reference: ask

> Natural language query interface to the knossos CLI surface.

`ari ask` translates plain English questions into ranked suggestions for commands, rites, agents, and workflows. Use it when you know what you want to accomplish but aren't sure which command to use.

**Family**: ask
**Commands**: 1
**Priority**: MEDIUM

---

## Synopsis

```bash
ari ask [query] [flags]
```

## Description

Ask a question in plain English and get ranked suggestions. Without a project context, searches CLI commands and concepts. With a project, also searches rites, agents, dromena, and routing.

## Subcommands

None. `ari ask` is a single command.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--domain` | string | all | Filter by domain (comma-separated): command, concept, rite, agent, dromena, routing, session, knowledge |
| `--limit` | int | 5 | Maximum results to return |
| `--session` | string | auto-detect | Session ID override |

## Examples

```bash
# Discover release workflow
ari ask "how do I release my project?"

# Search by concept
ari ask "code quality"

# Start a session
ari ask "start a session"

# JSON output for scripting
ari ask -o json "release"

# Filter to rite domain only
ari ask --domain=rite "ecosystem"

# Return up to 10 results
ari ask --limit 10 "session"

# Ask in context of a specific session
ari ask --session=session-20260308-143022-a1b2c3d4 "what next?"
```

## See Also

- [`ari explain`](cli-explain.md) — Look up definitions for specific concepts
- `ari --help` — Root command listing all families
