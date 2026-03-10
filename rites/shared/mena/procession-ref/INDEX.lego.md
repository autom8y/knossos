---
name: procession-ref
description: "Cross-rite procession workflow reference. Use when: understanding procession concepts, creating handoff artifacts between stations, managing station transitions. Triggers: procession reference, station workflow, cross-rite coordination, handoff schema, transition protocol."
---

# Procession Reference

> A procession is a coordinated cross-rite workflow -- a predetermined sequence of stations, each mapped to a rite, with formal handoff artifacts at each boundary.

## Core Concepts

- **Procession**: A named workflow template (e.g., `security-remediation`) instantiated for a session
- **Station**: A step in the procession, bound to a specific rite and goal
- **Handoff artifact**: A `HANDOFF-{source}-to-{target}.md` file with schema-valid frontmatter that transfers context between stations
- **Artifact directory**: `.sos/wip/{procession-name}/` -- all handoff artifacts live here

## Lifecycle

```
create -> station 1 (work) -> proceed -> station 2 (work) -> ... -> final station -> complete
                                  ^                                        |
                                  |-------- recede (on failure) -----------|
                                  |
                              abandon (terminate)
```

## Companion Files

- [handoff-schema.md](handoff-schema.md) -- Required YAML frontmatter fields for handoff artifacts
- [transition-protocol.md](transition-protocol.md) -- Proceed, recede, and abandon operation semantics

## CLI Commands

State-mutating operations are invoked by Moirai or by the named procession command (e.g., `/security-remediation`).

| Command | Description | Mutates State |
|---------|-------------|---------------|
| `ari procession create --template=<name>` | Initialize from template | Yes |
| `ari procession status` | Show current state | No |
| `ari procession proceed [--artifacts=path]` | Advance to next station | Yes |
| `ari procession recede --to=<station>` | Roll back to previous station | Yes |
| `ari procession abandon` | Terminate procession | Yes |

## Named Commands

Each procession template generates a named command during `ari sync`:
- `security-remediation.yaml` → `/security-remediation`
- The command handles full lifecycle: startup, station orchestration, transitions

## Related

- `cross-rite-handoff` skill -- General cross-rite handoff schema (non-procession)
- Per-procession skills (e.g., `security-remediation-ref`) -- Template-specific station data
