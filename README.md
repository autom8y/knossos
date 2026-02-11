# Knossos - Context Engineering Meta-Framework

## Commands

### User-Level Sync (knossos -> ~/.claude/)

| Command | Source | Target |
|---------|--------|--------|
| `ari sync user agents` | `agents/` | `~/.claude/agents/` |
| `ari sync user mena` | `mena/` | `~/.claude/commands/` + `~/.claude/skills/` |
| `ari sync user hooks` | `hooks/` | `~/.claude/hooks/` |

### Rite/Project Management

| Command | Purpose |
|---------|---------|
| `ari sync --rite=<name>` | Switch active rite (syncs to `.claude/`) |

### Architecture Note

User-level content (`agents/`, `mena/`, `hooks/`) syncs to `~/.claude/` (global, available in all projects).
Rite-level content (`rites/{rite}/`) syncs to `.claude/` (project-specific via `ari sync --rite=<name>`).

**Important**: NO `.claude/user-*` directories should exist in satellite projects. These were stale migration artifacts.

See [docs/INTEGRATION.md](docs/INTEGRATION.md) for full artifact architecture details.

## Usage

### Sync User Agents

Syncs agents from `agents/` to `~/.claude/agents/`.

```bash
# Sync user-agents
ari sync user agents

# Preview changes
ari sync user agents --dry-run

# Show sync status
ari sync user agents --status
```

**Behavior:**
- Additive: Never removes existing agents from `~/.claude/agents/`
- Overwrites: Only agents previously installed from knossos (tracked in manifest)
- Preserves: User-created agents not from knossos

**Integration Points:**
- Run manually after pulling knossos updates: `git pull && ari sync user agents`
- Add to shell profile for automatic sync on terminal open (optional)

**Manifest:** `~/.claude/USER_AGENT_MANIFEST.json` tracks knossos-managed agents.

### Sync User Mena

Syncs mena (commands + skills) from `mena/` to `~/.claude/commands/` and `~/.claude/skills/`.

```bash
# Sync user mena
ari sync user mena

# Preview changes
ari sync user mena --dry-run

# Show sync status
ari sync user mena --status
```

**Behavior:**
- Additive: Never removes existing commands/skills from `~/.claude/`
- Overwrites: Only mena previously installed from knossos (tracked in manifest)
- Preserves: User-created commands/skills not from knossos
- Distribution: `.dro.md` files → `commands/` (transient), `.lego.md` files → `skills/` (persistent)
- Scope filtering: `scope: user` = user pipeline only, `scope: project` = project pipeline only, no scope = both

**Source Structure:**
```
mena/
  session/        # Session management dromena
  workflow/       # Workflow dromena
  operations/     # Operation dromena
  navigation/     # Navigation dromena
  meta/           # Meta dromena
  rite-switching/ # Rite-switching dromena
  guidance/       # Guidance legomena
  templates/      # Template legomena
```

**Rite Mena:**
Rite-specific mena live in `rites/<rite>/mena/` and are synced to `.claude/commands/` and `.claude/skills/` by `ari sync`. Rite mena take precedence over user mena of the same name (project > user).

**Manifests:**
- `~/.claude/USER_COMMAND_MANIFEST.json` tracks knossos-managed commands
- `~/.claude/USER_SKILL_MANIFEST.json` tracks knossos-managed skills
