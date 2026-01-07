# Roster - Agent Rite Management

## Scripts

### User-Level Sync (roster -> ~/.claude/)

| Script | Source | Target |
|--------|--------|--------|
| `sync-user-agents.sh` | `user-agents/` | `~/.claude/agents/` |
| `sync-user-commands.sh` | `user-commands/` | `~/.claude/commands/` |
| `sync-user-skills.sh` | `user-skills/` | `~/.claude/skills/` |
| `sync-user-hooks.sh` | `user-hooks/` | `~/.claude/hooks/` |

### Rite/Project Management

| Script | Purpose |
|--------|---------|
| `swap-team.sh` | Switch active rite (syncs to `.claude/`) |
| `generate-team-context.sh` | Output team routing table (used by session hooks) |
| `load-workflow.sh` | Load workflow.yaml for a team |
| `get-workflow-field.sh` | Extract specific workflow fields |

### Architecture Note

User-level content (`user-*/`) syncs to `~/.claude/` (global, available in all projects).
Rite-level content (`rites/{rite}/`) syncs to `.claude/` (project-specific via swap-team).

**Important**: NO `.claude/user-*` directories should exist in satellite projects. These were stale migration artifacts.

See [docs/INTEGRATION.md](docs/INTEGRATION.md) for full artifact architecture details.

## Usage

### Generate Rite Context

```bash
# For active rite
./generate-team-context.sh

# For specific rite
./generate-team-context.sh 10x-dev
```

Output: Markdown table of phase→agent mappings for session hook injection.

### Sync User Agents

Syncs agents from `roster/user-agents/` to `~/.claude/agents/`.

```bash
# Sync user-agents
./sync-user-agents.sh

# Preview changes
./sync-user-agents.sh --dry-run

# Show sync status
./sync-user-agents.sh --status
```

**Behavior:**
- Additive: Never removes existing agents from `~/.claude/agents/`
- Overwrites: Only agents previously installed from roster (tracked in manifest)
- Preserves: User-created agents not from roster

**Integration Points:**
- Run manually after pulling roster updates: `git pull && ./sync-user-agents.sh`
- Add to shell profile for automatic sync on terminal open (optional)
- Hook into roster post-merge git hook (optional)

**Manifest:** `~/.claude/USER_AGENT_MANIFEST.json` tracks roster-managed agents.

### Sync User Commands

Syncs slash commands from `roster/user-commands/` to `~/.claude/commands/`.

```bash
# Sync user-commands
./sync-user-commands.sh

# Preview changes
./sync-user-commands.sh --dry-run

# Show sync status
./sync-user-commands.sh --status
```

**Behavior:**
- Additive: Never removes existing commands from `~/.claude/commands/`
- Overwrites: Only commands previously installed from roster (tracked in manifest)
- Preserves: User-created commands not from roster
- Flattens: Source subdirectories (session/, workflow/, etc.) become flat in target

**Source Structure:**
```
user-commands/
  session/       # start, park, continue, handoff, wrap (5)
  workflow/      # task, sprint, hotfix (3)
  operations/    # architect, build, qa, code-review, commit (5)
  navigation/    # consult, team, worktree, sessions, ecosystem (5)
  meta/          # minus-1, zero, one (3)
  rite-switching/ # 10x, docs, hygiene, debt, sre, security, intelligence, rnd, strategy, forge (10)
```

**Rite Commands:**
Rite-specific commands live in `rites/<rite>/commands/` and are synced to `.claude/commands/` by `swap-team.sh`. Rite commands take precedence over user commands of the same name (project > user).

**Manifest:** `~/.claude/USER_COMMAND_MANIFEST.json` tracks roster-managed commands.

### Sync User Skills

Syncs skill directories from `roster/user-skills/` to `~/.claude/skills/`.

```bash
# Sync user-skills
./sync-user-skills.sh

# Preview changes
./sync-user-skills.sh --dry-run

# Show sync status
./sync-user-skills.sh --status
```

**Behavior:**
- Additive: Never removes existing skills from `~/.claude/skills/`
- Overwrites: Only skills previously installed from roster (tracked in manifest)
- Preserves: User-created skills not from roster
- Uses `rsync --delete` for clean updates within roster-managed skills

**Key Differences from User Agents:**
- Skills are directories (containing SKILL.md + supporting files)
- Checksum computed over all files in skill directory
- Manifest tracks `file_count` in addition to checksum

**Integration Points:**
- Run manually after pulling roster updates: `git pull && ./sync-user-skills.sh`
- Combine with agent sync: `./sync-user-agents.sh && ./sync-user-skills.sh`
- Add to shell profile for automatic sync on terminal open (optional)

**Manifest:** `~/.claude/USER_SKILL_MANIFEST.json` tracks roster-managed skills.

**Included Skills:**
- `consult-ref/` - Ecosystem navigation reference (command reference, playbooks, team profiles)
- `forge-ref/` - Team creation patterns and evaluation harnesses
