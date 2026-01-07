# Integration Guide

> Roster is the unified source for the Claude Code ecosystem.

## Architecture

The Claude Code ecosystem uses a two-tier architecture:

```
+-------------------+
|      ROSTER       |   <-- Single source of truth
|  (this repository)|
+-------------------+
        |
        | roster-sync
        v
+-------------------+
| SATELLITE PROJECT |   <-- Your project
|  .claude/agents/  |
|  .claude/skills/  |
|  .claude/hooks/   |
+-------------------+
```

### Two-Tier Model

| Layer | Source | Purpose |
|-------|--------|---------|
| **Roster** | `$ROSTER_HOME` | Master repository for teams, agents, skills, commands |
| **Satellite** | Project's `.claude/` | Local project configuration, team agents, hooks |

This model replaces the previous three-tier architecture (skeleton -> roster -> satellite). Roster is now standalone and does not depend on any upstream repository.

## Synchronization

Use `roster-sync` to manage project ecosystem files:

```bash
# Initialize a new project
$ROSTER_HOME/roster-sync init

# Sync updates from roster
$ROSTER_HOME/roster-sync sync

# Sync and refresh active team
$ROSTER_HOME/roster-sync sync --refresh

# Check sync status
$ROSTER_HOME/roster-sync status

# Preview changes
$ROSTER_HOME/roster-sync diff
```

### roster-sync Commands

| Command | Purpose |
|---------|---------|
| `init [path]` | Initialize a new satellite project |
| `sync` | Pull updates from roster |
| `status` | Show sync status and active team |
| `validate` | Check manifest integrity |
| `diff [file]` | Show pending changes |
| `repair` | Rebuild manifest from current state |

### Sync Flags

| Flag | Purpose |
|------|---------|
| `--force`, `-f` | Override conflict detection |
| `--dry-run`, `-n` | Preview changes without applying |
| `--refresh`, `-r` | Also refresh active team |
| `--prune`, `-p` | Remove orphaned files |

## Team Management

```bash
# List available teams
./swap-team.sh --list

# Switch to a team
./swap-team.sh <team-name>

# Refresh current team from roster
./swap-team.sh --refresh

# Show current active team
./swap-team.sh
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ROSTER_HOME` | `$HOME/Code/roster` | Path to roster repository |

**Setup (add to ~/.zshrc or ~/.bashrc):**
```bash
export ROSTER_HOME="$HOME/Code/roster"
```

## Team Pack Structure

```
teams/<team-name>/
  +-- agents/           # Agent prompt files (*.md)
  |   +-- orchestrator.md
  |   +-- specialist.md
  +-- skills/           # Team-specific skills (optional)
  +-- commands/         # Team-specific commands (optional)
  +-- workflow.yaml     # Phase definitions
  +-- README.md         # Team documentation
```

## Artifact Architecture

Roster uses a **two-tier destination model**: user-level artifacts go to `~/.claude/` (global), while team-level artifacts go to `.claude/` (project-level).

### Source to Target Mapping

| Source (roster/) | Materialized Target | Sync Mechanism | Scope |
|------------------|---------------------|----------------|-------|
| `user-agents/` | `~/.claude/agents/` | `sync-user-agents.sh` | User (global) |
| `user-commands/` | `~/.claude/commands/` | `sync-user-commands.sh` | User (global) |
| `user-skills/` | `~/.claude/skills/` | `sync-user-skills.sh` | User (global) |
| `user-hooks/` | `~/.claude/hooks/` | `sync-user-hooks.sh` | User (global) |
| `teams/{pack}/agents/` | `.claude/agents/` | `swap-team.sh` | Project |
| `teams/{pack}/skills/` | `.claude/skills/` | `swap-team.sh` | Project |
| `teams/{pack}/commands/` | `.claude/commands/` | `swap-team.sh` | Project |
| `.claude/` (roster) | `.claude/` (satellite) | `roster-sync` | Project |

### Key Points

1. **User-level content** (`user-*/`) syncs to `~/.claude/` (available in all projects)
2. **Team-level content** (`teams/{pack}/`) syncs to `.claude/` (project-specific)
3. **NO `.claude/user-*` directories should exist** - these were stale migration artifacts from the skeleton deprecation
4. **Precedence**: Project-level (`.claude/`) takes precedence over user-level (`~/.claude/`)

### Data Flow Diagram

```
Roster Repository
    |
    +-- user-agents/    -> sync-user-agents.sh    -> ~/.claude/agents/   (global)
    +-- user-commands/  -> sync-user-commands.sh  -> ~/.claude/commands/ (global)
    +-- user-skills/    -> sync-user-skills.sh    -> ~/.claude/skills/   (global)
    +-- user-hooks/     -> sync-user-hooks.sh     -> ~/.claude/hooks/    (global)
    |
    +-- teams/{pack}/   -> swap-team.sh           -> .claude/agents/     (project)
    |                                             -> .claude/skills/     (project)
    |                                             -> .claude/commands/   (project)
    |
    +-- .claude/        -> roster-sync            -> satellite/.claude/  (project)
```

### Sync Scripts

Run these after pulling roster updates:

```bash
# Sync all user-level content
./sync-user-agents.sh
./sync-user-commands.sh
./sync-user-skills.sh
./sync-user-hooks.sh

# Or preview changes first
./sync-user-agents.sh --dry-run
./sync-user-commands.sh --dry-run
./sync-user-skills.sh --dry-run
./sync-user-hooks.sh --dry-run

# Check status
./sync-user-agents.sh --status
```

### Common Mistakes to Avoid

| Mistake | Correct Approach |
|---------|------------------|
| Creating `.claude/user-agents/` in satellite | User agents go to `~/.claude/agents/` |
| Copying team agents to `~/.claude/` | Team agents go to `.claude/agents/` via swap-team |
| Manually editing synced files | Edit in roster, then sync |

## Manifest Schema (v3)

```json
{
  "schema_version": 3,
  "roster": {
    "path": "/path/to/roster",
    "commit": "abc123",
    "ref": "main",
    "last_sync": "2026-01-03T00:00:00Z"
  },
  "team": {
    "name": "10x-dev-pack",
    "last_swap": "2026-01-03T00:00:00Z"
  },
  "managed_files": [
    {
      "path": ".claude/CLAUDE.md",
      "roster_checksum": "abc...",
      "local_checksum": "abc..."
    }
  ]
}
```

## Related Files in This Repository

- `swap-team.sh` - Team switching script
- `roster-sync/` - Ecosystem synchronization tool
- `workflow-schema.yaml` - Team pack schema reference
- `RITE_SKILL_MATRIX.md` - Agent skill assignments
- `generate-team-context.sh` - Team routing table generator

## Migration from CEM

If migrating from the previous CEM (Claude Ecosystem Manager) system:

1. Ensure `ROSTER_HOME` is set (previously `SKELETON_HOME` was also required)
2. Run `roster-sync sync --force` to migrate manifest from v1/v2 to v3
3. Remove `SKELETON_HOME` from shell configuration (no longer needed)

See `/Users/tomtenuta/Code/roster/docs/migration/cem-to-roster-migration.md` for detailed migration instructions.
