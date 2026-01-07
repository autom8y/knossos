---
name: ecosystem-ref
description: "Quick reference for roster ecosystem patterns. Use when: working with roster-sync, swap-rite.sh, manifest schema, rite management. Triggers: roster-sync, swap-rite, roster, manifest, ecosystem patterns."
---

# ecosystem-ref

> Quick reference for roster ecosystem patterns.

## roster-sync (Ecosystem Manager)

Roster-sync manages synchronization between roster repository and user/project Claude configurations.

### Sync Scripts
| Script | Purpose | Target |
|--------|---------|--------|
| `sync-user-agents.sh` | Sync agents to user config | `~/.claude/agents/` |
| `sync-user-commands.sh` | Sync commands to user config | `~/.claude/commands/` |
| `sync-user-skills.sh` | Sync skills to user config | `~/.claude/skills/` |
| `swap-rite.sh` | Switch active rite | `.claude/agents/` |

### Key Paths
- Roster: `$ROSTER_HOME` or `~/Code/roster`
- User Agents: `~/.claude/agents/`
- User Commands: `~/.claude/commands/`
- User Skills: `~/.claude/skills/`
- Team Manifest: `.claude/TEAM_MANIFEST.json`

### Common Commands
```bash
./sync-user-agents.sh          # Sync user-agents to ~/.claude/agents/
./sync-user-commands.sh        # Sync user-commands to ~/.claude/commands/
./sync-user-skills.sh          # Sync user-skills to ~/.claude/skills/
./swap-rite.sh <pack>          # Switch active rite
./swap-rite.sh --list          # List available rites
./swap-rite.sh --refresh       # Refresh current team
```

## Roster (Team Pack Manager)

### Team Pack Structure
```
rites/{name}/
  agents/           # Agent definitions (*.md)
  commands/         # Team-specific slash commands
  skills/           # Team-specific skills (Phase 2)
  workflow.yaml     # Phase orchestration
  README.md         # Pack documentation
```

### swap-rite.sh
```bash
swap-rite.sh <pack>           # Switch to rite
swap-rite.sh --list           # List available packs
swap-rite.sh --refresh        # Refresh current team
swap-rite.sh <pack> --keep-all    # Preserve orphan agents
swap-rite.sh <pack> --remove-all  # Remove orphan agents
```

### Orphan Handling
Orphan = agent from previous team not in new team.
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

## Team Manifest Schema

```json
{
  "schema_version": 1,
  "team": { "name": "", "last_swap": "" },
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

- [doc-ecosystem skill](../doc-ecosystem/SKILL.md) - Templates for ecosystem documentation
- [claude-md-architecture skill](../claude-md-architecture/SKILL.md) - CLAUDE.md architecture patterns
