---
last_verified: 2026-03-26
---

# CLI Reference: status

> Display a unified health overview of all Knossos directory trees.

`ari status` gives you a single-command dashboard of your project's Knossos state. It reads active rite, agent count, sync recency, knowledge freshness, artifact counts, and session state from the live filesystem.

**Family**: status
**Commands**: 1
**Priority**: HIGH

---

## Commands

### ari status

Display the Knossos project health dashboard.

**Synopsis**:
```bash
ari status [flags]
```

**Description**:
Displays a unified health overview of all Knossos directory trees:
- Channel directory (`.claude/` or `.gemini/`)
- `.knossos/`
- `.know/`
- `.ledge/`
- `.sos/`

Reports:
- Active rite
- Agent count
- Sync recency
- Knowledge freshness
- Artifact counts
- Session state

This is a **read-only** command. It does not modify any state.

**Examples**:
```bash
# Human-readable dashboard
ari status

# Machine-readable JSON output
ari status -o json
```

**Typical output includes**:
```
Active rite:    docs
Channel:        .claude/ (claude)
Agents:         12 standing, 3 summoned
Last sync:      2 hours ago
Knowledge:      5 domains (2 stale)
Artifacts:      3 in ledge, 0 on shelf
Session:        ACTIVE — "Document API endpoints"
```

---

## When to Use

Run `ari status` when you:
- Want a quick project health check before starting work
- Need to confirm the active rite and sync state
- Are debugging why an agent doesn't see expected context

For detailed directory contents, use [`ari tour`](cli-tour.md). For knowledge domain freshness, use [`ari knows`](cli-knows.md).

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

- [`ari tour`](cli-tour.md) — Directory tree with file counts
- [`ari knows`](cli-knows.md) — Knowledge domain freshness details
- [`ari session status`](cli-session.md#ari-session-status) — Session-specific state
