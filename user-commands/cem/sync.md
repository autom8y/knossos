---
description: Sync project with roster ecosystem using ari CLI
argument-hint: [status|pull|push|diff|materialize|resolve|history|reset] [--rite=NAME] [--force]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Execute ari sync to synchronize project with roster ecosystem.

## Behavior

**CRITICAL**: Execute EXACTLY this command based on arguments:

| If $ARGUMENTS is... | Run this command |
|---------------------|------------------|
| Empty (no args) | `ari sync status` |
| `--refresh` | `ari sync materialize` |
| Anything else | `ari sync $ARGUMENTS` |

**IMPORTANT - Interpret status output**:
- If status shows **empty table** (just headers, no rows): Project is NOT configured. Tell user:
  > "Project not yet configured for Knossos sync. To set up, run one of:"
  > - `ari sync materialize --rite=10x-dev` (within roster repo)
  > - `ari sync materialize --rite=10x-dev --source=knossos` (consumer project)
  > - `ari sync materialize --minimal --source=knossos` (cross-cutting mode, no agents)
- If status shows tracked paths: Report the actual status

**Handle errors**:
- If "no ACTIVE_RITE found": Suggest `ari sync materialize --rite=<name> --source=knossos` or `--minimal`
- If ari not found: `cd ~/Code/roster && CGO_ENABLED=0 go install ./cmd/ari`

## Command Mapping

| /sync command | ari sync command | Description |
|---------------|------------------|-------------|
| `/sync status` | `ari sync status` | Show sync status |
| `/sync pull` | `ari sync pull` | Pull remote changes |
| `/sync push` | `ari sync push` | Push local changes |
| `/sync diff` | `ari sync diff` | Show differences |
| `/sync materialize` | `ari sync materialize` | Generate .claude/ from templates |
| `/sync materialize --rite=X` | `ari sync materialize --rite X` | Generate for specific rite |
| `/sync --refresh` | `ari sync materialize` | Refresh/regenerate .claude/ |
| `/sync resolve` | `ari sync resolve` | Resolve conflicts |
| `/sync history` | `ari sync history` | Show audit log |
| `/sync reset` | `ari sync reset` | Reset sync state (dangerous) |

## Legacy Compatibility

The following legacy commands are deprecated:
- `roster-sync` shell script → Use `ari sync` instead
- `/sync init` → Use `ari sync materialize` for new projects
- `/sync validate` → Use `ari manifest validate` instead
- `/sync repair` → Use `ari sync reset` followed by `ari sync materialize`

## Common Commands

```bash
# Within roster repo (has local rites)
/sync                             # Show sync status (default)
/sync status                      # Show sync status
/sync materialize                 # Generate .claude/ from templates
/sync materialize --rite=hygiene  # Generate for specific rite

# Bootstrap NEW project (creates .claude/ if missing)
/sync materialize --rite=10x-dev --source=knossos

# Cross-cutting mode (no rite, just base infrastructure)
/sync materialize --minimal --source=knossos
```

## Bootstrapping New Projects

`materialize` can bootstrap a new project from scratch - it will create the `.claude/` directory if it doesn't exist:

```bash
cd ~/Code/my-new-project
ari sync materialize --rite=10x-dev --source=knossos  # Full orchestrated workflow
# OR
ari sync materialize --minimal --source=knossos       # Just base infrastructure
```

## Reference

Full documentation: `.claude/skills/ecosystem-ref/skill.md`
