# ecosystem-ref

> Quick reference for CEM/skeleton/roster ecosystem patterns.

## CEM (Claude Ecosystem Manager)

### File Strategies
| Strategy | Behavior | Used For |
|----------|----------|----------|
| `copy-replace` | Skeleton overwrites satellite | commands/, hooks/, knowledge/ |
| `merge-dir` | Union content, preserve satellite-specific | skills/ |
| `merge-settings` | Deep merge arrays (permissions, MCP servers) | settings.local.json |
| `merge-docs` | Section-aware merge with markers | CLAUDE.md |

### Key Paths
- Skeleton: `$SKELETON_HOME` or `~/Code/skeleton_claude`
- Roster: `$ROSTER_HOME` or `~/Code/roster`
- Manifest: `.claude/.cem/manifest.json`
- State: `.claude/.cem/`

### Common Commands
```bash
cem init              # Initialize satellite from skeleton
cem sync              # Pull skeleton updates
cem sync --refresh    # Sync + refresh active team
cem validate          # Check manifest integrity
cem repair            # Rebuild manifest from .claude/
cem status            # Show sync status
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

## Three-Tier Layering

```
skeleton (base) -> team (overlay) -> satellite (local)
```

| Layer | Source | Precedence |
|-------|--------|------------|
| Skeleton | `$SKELETON_HOME/.claude/` | Base |
| Team | `$ROSTER_HOME/teams/{name}/` | Overlay (wins collisions) |
| Satellite | `.claude/user-*/`, `.claude/PROJECT.md` | Preserved (never touched) |

## Manifest Schema v2

```json
{
  "schema_version": 2,
  "skeleton": { "path": "", "commit": "", "ref": "", "last_sync": "" },
  "team": { "name": "", "last_swap": "" },
  "managed": {
    "skills": [],
    "commands": [],
    "agents": [],
    "hooks": []
  },
  "preserved": { ... }
}
```

## Debugging

```bash
CEM_DEBUG=1 cem sync          # Verbose CEM output
ROSTER_DEBUG=1 swap-team.sh   # Verbose roster output
```
