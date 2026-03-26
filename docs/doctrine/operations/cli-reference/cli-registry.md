---
last_verified: 2026-03-26
---

# CLI Reference: registry

> Sync, list, and inspect the cross-repo knowledge domain catalog.

The registry catalogs `.know/` domains from GitHub repositories in your org, enabling cross-repo knowledge discovery for Clew queries. Sync it periodically to keep the catalog fresh.

**Family**: registry
**Commands**: 3
**Priority**: HIGH

---

## Commands

### ari registry sync

Sync the knowledge domain catalog from GitHub.

**Synopsis**:
```bash
ari registry sync [flags]
```

**Description**:
Discovers repos and catalogs `.know/` domains for the active org. Reads repo list from `org.yaml` if configured; otherwise discovers repos via the GitHub API. Persists the catalog to:

```
$XDG_DATA_HOME/knossos/registry/{org}/domains.yaml
```

The GitHub token is read from `--token` flag or `GITHUB_TOKEN` environment variable. Without a token, requests are rate-limited to 60/hour.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--org` | string | active org | Override active org |
| `--token` | string | `GITHUB_TOKEN` env | GitHub token for API access |

**Examples**:
```bash
# Sync for the active org
ari registry sync

# Sync for a specific org
ari registry sync --org autom8y

# Provide a GitHub token explicitly
ari registry sync --token ghp_xxx
```

**Related Commands**:
- [`ari registry list`](#ari-registry-list) — View the catalog after syncing
- [`ari registry status`](#ari-registry-status) — Check sync summary

---

### ari registry list

List cataloged knowledge domains.

**Synopsis**:
```bash
ari registry list [flags]
```

**Description**:
Displays the domain catalog for the active org. Reads from the persisted catalog at:

```
$XDG_DATA_HOME/knossos/registry/{org}/domains.yaml
```

Run `ari registry sync` first to populate the catalog.

**Output columns**: REPO, DOMAIN, QUALIFIED NAME, GENERATED, EXPIRES, HASH

**Aliases**: `list`, `ls`

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--org` | string | active org | Override active org |
| `--repo` | string | - | Filter by repository name |
| `--stale` | bool | false | Show only stale domains |

**Examples**:
```bash
# List all cataloged domains
ari registry list

# Filter to knossos repo domains
ari registry list --repo knossos

# Show only stale domains (need re-sync)
ari registry list --stale

# JSON output for scripting
ari registry list -o json
```

**Related Commands**:
- [`ari registry sync`](#ari-registry-sync) — Refresh the catalog
- [`ari registry status`](#ari-registry-status) — Summary view

---

### ari registry status

Show registry sync summary.

**Synopsis**:
```bash
ari registry status [flags]
```

**Description**:
Displays a summary of the knowledge domain catalog: last sync time, domain count, repo count, and stale domain count.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--org` | string | active org | Override active org |

**Examples**:
```bash
# Show sync summary for active org
ari registry status

# Show for a specific org
ari registry status --org autom8y

# JSON output
ari registry status -o json
```

---

## Typical Registry Workflow

```bash
# 1. Set your active org
ari org set autom8y

# 2. Sync the catalog (set GITHUB_TOKEN first)
export GITHUB_TOKEN=ghp_xxx
ari registry sync

# 3. Verify sync succeeded
ari registry status

# 4. Browse domains
ari registry list

# 5. Check for stale entries periodically
ari registry list --stale
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

- [`ari org`](cli-org.md) — Org management (prerequisite for registry)
- [`ari serve`](cli-serve.md) — Clew server (uses the registry for queries)
- [`ari knows`](cli-knows.md) — Local knowledge domains (`.know/`)
