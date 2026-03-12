---
name: ecosystem-ref
description: "Knossos ecosystem patterns reference. Use when: running ari sync commands, switching rites, understanding manifest lifecycle, debugging sync issues. Triggers: knossos-sync, ari-sync, rite-switch, manifest, ecosystem patterns."
---

# ecosystem-ref

> Quick reference for knossos ecosystem patterns.

## ari sync (Ecosystem Manager)

Ari sync manages synchronization between knossos repository and user/project Claude configurations.

### Sync Commands
| Command | Purpose | Target |
|---------|---------|--------|
| `ari sync --scope=user --resource=agents` | Sync agents to user config | user channel agents directory |
| `ari sync --scope=user --resource=mena` | Sync commands + skills to user config | user channel commands + skills directories |
| `ari sync --scope=user --resource=hooks` | Sync hooks to user config | user channel hooks directory |
| `ari sync --scope=user` | Sync all user resources | All of the above |

### Key Paths
- Knossos: `$KNOSSOS_HOME` or `~/Code/knossos`
- User Agents: user channel agents directory
- User Commands: user channel commands directory
- User Skills: user channel skills directory

### Common Commands
```bash
ari sync --scope=user --resource=agents  # Sync agents/ to user channel agents directory
ari sync --scope=user --resource=mena   # Sync mena/ to user channel commands + skills directories
ari sync --scope=user --resource=hooks  # Sync hooks/ to user channel hooks directory
ari sync --scope=user                   # Sync all user resources
ari sync --rite=<name>         # Switch/activate a rite
ari rite list                  # List available rites
```

## Rite Manager

### Rite Structure
```
rites/{name}/
  agents/           # Agent definitions (*.md)
  commands/         # Rite-specific slash commands
  skills/           # Rite-specific skills
  workflow.yaml     # Phase orchestration
  README.md         # Rite documentation
```

### Rite Switching
```bash
ari sync --rite <rite>                    # Switch to rite
ari rite list                             # List available rites
ari sync --scope=rite                     # Refresh current rite
ari sync --rite <rite> --keep-orphans     # Preserve orphan agents
```

### Orphan Handling
Orphan = agent from previous rite not in new rite.
- Default: auto-remove knossos-owned orphans
- `--keep-orphans` preserves orphaned files

## Two-Tier Layering

```
knossos (base) -> project (local overlay)
```

| Layer | Source | Precedence |
|-------|--------|------------|
| Knossos | `$KNOSSOS_HOME/rites/{name}/` | Base agents and skills |
| Project | channel agents and skills directories | Local overrides |

## Rite Manifest Schema

```yaml
# rites/{name}/workflow.yaml
name: ""
workflow_type: sequential
description: ""
entry_point:
  agent: ""
phases:
  - name: ""
    agent: ""
    produces: ""
    next: ""
```

## Debugging

```bash
ari sync --dry-run             # Preview sync changes
ari sync --help                # Full flag reference
ari rite --help                # Rite management subcommands
ari session --help             # Session management subcommands
```

## Progressive Disclosure

- [doc-ecosystem skill](../doc-ecosystem/INDEX.md) - Templates for ecosystem documentation
- [claude-md-architecture skill](../claude-md-architecture/INDEX.md) - CLAUDE.md architecture patterns
