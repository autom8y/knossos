---
last_verified: 2026-03-26
---

# CLI Reference: org

> Create, configure, and manage organization-level resources.

Organizations provide shared rites, agents, and mena (commands + skills) across multiple projects for a team. Set up an org once, then all projects in that org resolve shared resources automatically.

**Family**: org
**Commands**: 4
**Priority**: HIGH

---

## Commands

### ari org init

Bootstrap an organization directory.

**Synopsis**:
```bash
ari org init <org-name> [flags]
```

**Description**:
Creates the org directory structure at the XDG data path. Org names must be kebab-case (lowercase letters, digits, hyphens).

Creates the following structure:
```
$XDG_DATA_HOME/knossos/orgs/<org-name>/
  org.yaml        # Org metadata
  rites/          # Org-level rites
  agents/         # Org-level agents
  mena/           # Org-level mena (commands + skills)
```

Run this before `ari org set` when setting up a new org.

**Arguments**:
- `org-name` (string, required): Organization name (kebab-case)

**Examples**:
```bash
# Bootstrap the autom8y org
ari org init autom8y

# Bootstrap a team org
ari org init my-team
```

**Related Commands**:
- [`ari org set`](#ari-org-set) — Activate the org after initialization

---

### ari org set

Set the active organization.

**Synopsis**:
```bash
ari org set <org-name> [flags]
```

**Description**:
Sets the active org by writing to `$XDG_CONFIG_HOME/knossos/active-org`. The active org is used by `ari sync` to resolve org-level resources during materialization.

Can also be set via the `KNOSSOS_ORG` environment variable (takes precedence over the file).

**Arguments**:
- `org-name` (string, required): Organization name to activate

**Examples**:
```bash
# Set autom8y as the active org
ari org set autom8y

# Verify the active org
ari org current
```

**Related Commands**:
- [`ari org current`](#ari-org-current) — Confirm the active org

---

### ari org list

List available organizations.

**Synopsis**:
```bash
ari org list [flags]
```

**Description**:
Discovers all organizations at the XDG data path. Lists all directories under `$XDG_DATA_HOME/knossos/orgs/`. The active org is marked with an asterisk.

**Aliases**: `list`, `ls`

**Examples**:
```bash
# List all orgs
ari org list

# Same with alias
ari org ls
```

**Related Commands**:
- [`ari org set`](#ari-org-set) — Activate an org from the list

---

### ari org current

Show the active organization.

**Synopsis**:
```bash
ari org current [flags]
```

**Description**:
Displays the currently active organization. Resolution order:
1. `$KNOSSOS_ORG` environment variable
2. `$XDG_CONFIG_HOME/knossos/active-org` file

**Examples**:
```bash
# Show the active org
ari org current
```

---

## Typical Org Setup Workflow

```bash
# 1. Bootstrap the org directory
ari org init my-team

# 2. Set it as active
ari org set my-team

# 3. Verify
ari org current

# 4. Configure serve.env (for Clew / ari serve)
# Edit $XDG_DATA_HOME/knossos/orgs/my-team/serve.env

# 5. Sync to pick up org-level resources
ari sync materialize
```

---

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--channel` | string | `all` | Target channel: claude, gemini, or all |
| `--config` | string | `$XDG_CONFIG_HOME/knossos/config.yaml` | Config file path |
| `-o, --output` | string | `text` | Output format: text, json, yaml |
| `-p, --project-dir` | string | auto-discovered | Project root directory |
| `-s, --session-id` | string | current session | Override session ID |
| `-v, --verbose` | bool | false | Enable verbose output (JSON lines to stderr) |

---

## See Also

- [`ari serve`](cli-serve.md) — Clew webhook server (uses org env file)
- [`ari registry`](cli-registry.md) — Cross-repo knowledge catalog (org-scoped)
- [`ari sync`](cli-sync.md) — Materialize channel directory (resolves org resources)
