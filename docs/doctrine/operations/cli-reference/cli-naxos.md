# CLI Reference: naxos

> Cleanup tooling for abandoned sessions.

Named after the island where Theseus abandoned Ariadne, [Naxos](../../reference/GLOSSARY.md#naxos) identifies sessions that may need cleanup attention.

**Family**: naxos
**Commands**: 1
**Priority**: LOW

---

## Commands

### ari naxos scan

Scan for orphaned sessions.

**Synopsis**:
```bash
ari naxos scan [flags]
```

**Description**:
Scans for sessions that may be abandoned or need attention:
- **Inactive sessions**: No activity for extended period
- **Stale sails**: Gray sails that haven't been upgraded
- **Incomplete wraps**: Sessions marked for wrap but never completed

**⚠️ Important**: Naxos is report-only. It suggests actions but does not automatically clean up sessions.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--inactive-threshold` | string | `48h` | Inactivity threshold (e.g., 48h, 7d) |
| `--json` | bool | false | JSON output for scripting |

**Examples**:
```bash
# Scan for orphans
ari naxos scan

# Custom inactivity threshold
ari naxos scan --inactive-threshold=72h

# JSON output
ari naxos scan --json
```

**Output Fields**:
- `session_id`: Session identifier
- `status`: Current session status
- `last_activity`: Timestamp of last event
- `age`: Time since creation
- `reason`: Why it's flagged (inactive, stale_sails, incomplete_wrap)
- `suggested_action`: Recommended action (resume, wrap, archive)

**Related Commands**:
- [`ari session list`](cli-session.md#ari-session-list) — List all sessions
- [`ari session wrap`](cli-session.md#ari-session-wrap) — Complete orphaned session
- [`ari worktree cleanup`](cli-worktree.md#ari-worktree-cleanup) — Clean stale worktrees

---

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `$XDG_CONFIG_HOME/ariadne/config.yaml` | Config file path |
| `-o, --output` | string | `text` | Output format: text, json, yaml |
| `-p, --project-dir` | string | auto-discovered | Project root directory |
| `-s, --session-id` | string | current session | Override session ID |
| `-v, --verbose` | bool | false | Enable verbose output |

---

## See Also

- [Naxos Glossary Entry](../../reference/GLOSSARY.md#naxos)
- [Mythology Concordance - Naxos](../../philosophy/mythology-concordance.md#naxos-shore-of-abandonment)
