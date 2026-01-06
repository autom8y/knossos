---
name: ecosystem-ref
description: "Quick reference for roster ecosystem patterns. Use when: working with roster-sync, swap-team.sh, manifest schema, team management. Triggers: roster-sync, swap-team, roster, manifest, ecosystem patterns."
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
| `swap-team.sh` | Switch active team pack | `.claude/agents/` |

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
./swap-team.sh <pack>          # Switch active team pack
./swap-team.sh --list          # List available team packs
./swap-team.sh --refresh       # Refresh current team
```

## Roster (Team Pack Manager)

### Team Pack Structure
```
teams/{name}/
  agents/           # Agent definitions (*.md)
  commands/         # Team-specific slash commands
  skills/           # Team-specific skills (Phase 2)
  workflow.yaml     # Phase orchestration
  README.md         # Pack documentation
```

### swap-team.sh
```bash
swap-team.sh <pack>           # Switch to team pack
swap-team.sh --list           # List available packs
swap-team.sh --refresh        # Refresh current team
swap-team.sh <pack> --keep-all    # Preserve orphan agents
swap-team.sh <pack> --remove-all  # Remove orphan agents
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
| Roster | `$ROSTER_HOME/teams/{name}/` | Base agents and skills |
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
ROSTER_DEBUG=1 swap-team.sh   # Verbose roster output
```

## Progressive Disclosure

- [doc-ecosystem skill](../doc-ecosystem/SKILL.md) - Templates for ecosystem documentation
- [claude-md-architecture skill](../claude-md-architecture/SKILL.md) - CLAUDE.md architecture patterns
