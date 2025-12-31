# Integration Guide

> For the complete integration documentation, see the primary guide in skeleton_claude.

## Quick Reference

The Claude Code ecosystem uses two repositories:
- **skeleton_claude**: Infrastructure (commands, skills, hooks)
- **roster**: Team packs (agents, workflows)

## Primary Documentation

**Full integration guide:** `$SKELETON_HOME/docs/INTEGRATION.md`

Or if skeleton is in the default location:
```
~/Code/skeleton_claude/docs/INTEGRATION.md
```

## Roster-Specific Commands

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
| `ROSTER_HOME` | `$ROSTER_HOME` | Path to this repository |
| `SKELETON_HOME` | `~/Code/skeleton_claude` | Path to skeleton repository |

## Team Pack Structure

```
teams/<team-name>/
  ├── agents/           # Agent prompt files (*.md)
  │   ├── orchestrator.md
  │   └── specialist.md
  ├── workflow.yaml     # Phase definitions
  └── commands/         # Team-specific commands (optional)
```

## Related Files in This Repository

- `swap-team.sh` - Team switching script
- `workflow-schema.yaml` - Team pack schema reference
- `TEAM_SKILL_MATRIX.md` - Agent skill assignments
- `generate-team-context.sh` - Team routing table generator
