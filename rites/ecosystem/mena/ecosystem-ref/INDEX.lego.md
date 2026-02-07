---
name: ecosystem-ref
description: "Quick reference for roster ecosystem patterns. Use when: working with roster-sync, swap-rite.sh, manifest schema, rite management. Triggers: roster-sync, swap-rite, roster, manifest, ecosystem patterns."
---

# ecosystem-ref

> Quick reference for roster ecosystem patterns.

## roster-sync (Ecosystem Manager)

Roster-sync manages synchronization between roster repository and user/project Claude configurations.

### Sync Commands
| Command | Purpose | Target |
|---------|---------|--------|
| `ari sync user agents` | Sync agents to user config | `~/.claude/agents/` |
| `ari sync user mena` | Sync commands + skills to user config | `~/.claude/commands/` + `~/.claude/skills/` |
| `ari sync user hooks` | Sync hooks to user config | `~/.claude/hooks/` |
| `ari sync user` | Sync all user resources | All of the above |

### Key Paths
- Knossos: `$KNOSSOS_HOME` or `~/Code/knossos`
- User Agents: `~/.claude/agents/`
- User Commands: `~/.claude/commands/`
- User Skills: `~/.claude/skills/`

### Common Commands
```bash
ari sync user agents           # Sync agents/ to ~/.claude/agents/
ari sync user mena             # Sync mena/ to ~/.claude/commands/ + skills/
ari sync user hooks            # Sync hooks/ to ~/.claude/hooks/
ari sync user                  # Sync all user resources
ari rite start <rite>          # Start a rite (includes materialize)
ari rite list                  # List available rites
```

## Roster (Rite Manager)

### Rite Structure
```
rites/{name}/
  agents/           # Agent definitions (*.md)
  commands/         # Rite-specific slash commands
  skills/           # Rite-specific skills
  workflow.yaml     # Phase orchestration
  README.md         # Rite documentation
```

### swap-rite.sh
```bash
swap-rite.sh <rite>           # Switch to rite
swap-rite.sh --list           # List available rites
swap-rite.sh --refresh        # Refresh current rite
swap-rite.sh <rite> --keep-all    # Preserve orphan agents
swap-rite.sh <rite> --remove-all  # Remove orphan agents
```

### Orphan Handling
Orphan = agent from previous rite not in new rite.
- Interactive: k/p/r per agent (keep/promote/remove)
- Non-interactive: `--keep-all`, `--remove-all`, `--promote-all`

## Two-Tier Layering

```
roster (base) -> project (local overlay)
```

| Layer | Source | Precedence |
|-------|--------|------------|
| Roster | `$ROSTER_HOME/rites/{name}/` | Base agents and skills |
| Project | `.claude/agents/`, `.claude/skills/` | Local overrides |

## Rite Manifest Schema

```json
{
  "schema_version": 1,
  "rite": { "name": "", "last_swap": "" },
  "managed": {
    "agents": [],
    "commands": [],
    "skills": []
  }
}
```

## Debugging

```bash
ROSTER_DEBUG=1 swap-rite.sh   # Verbose roster output
```

## Progressive Disclosure

- [doc-ecosystem skill](../doc-ecosystem/INDEX.lego.md) - Templates for ecosystem documentation
- [claude-md-architecture skill](../claude-md-architecture/INDEX.lego.md) - CLAUDE.md architecture patterns
