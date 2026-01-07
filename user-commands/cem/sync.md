---
description: Sync project with roster ecosystem using ari CLI
argument-hint: [status|pull|push|diff|materialize|resolve|history|reset] [--rite=NAME] [--force]
allowed-tools: Bash, Read
model: haiku
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Your Task

Execute ari sync to synchronize project with roster ecosystem. $ARGUMENTS

## Behavior

1. **Execute ari sync** using the installed binary:
   ```bash
   ari sync [command] $ARGUMENTS
   ```
   If `ari` is not in PATH, fall back to: `~/bin/ari sync [command] $ARGUMENTS`

2. **Pass through all arguments**:
   - Commands: status, pull, push, diff, materialize, resolve, history, reset
   - Flags: --rite=NAME (for materialize), --force, --verbose
   - Display output directly to user

3. **Handle errors**:
   - If ari not found:
     - ERROR: "ari not found. Install via: brew install autom8y/tap/ari"
     - Or build locally: "cd ~/Code/roster && just build && cp ari ~/bin/"
   - If execution fails: Display stderr for debugging

4. **Special handling for --refresh flag**:
   - If `--refresh` is passed, translate to: `ari sync materialize`
   - This regenerates .claude/ from templates and active rite

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
/sync status                      # Show sync status
/sync materialize                 # Generate .claude/ from templates
/sync materialize --rite=hygiene  # Generate for specific rite
/sync pull                        # Pull remote changes
/sync diff                        # Show pending changes
```

## Reference

Full documentation: `.claude/skills/ecosystem-ref/skill.md`
