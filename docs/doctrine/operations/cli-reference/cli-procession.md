---
last_verified: 2026-03-26
---

# CLI Reference: procession

> Manage template-defined, station-based workflows that coordinate work across multiple rites.

A procession is a structured cross-rite workflow. Each procession lives within a session and tracks progress through an ordered sequence of rite-scoped stations. When a station completes, artifacts produced are recorded before advancing.

**Family**: procession
**Commands**: 6
**Priority**: HIGH

---

## Commands

### ari procession list

List available procession templates.

**Synopsis**:
```bash
ari procession list [flags]
```

**Description**:
Lists all available procession templates resolved through the 5-tier resolution chain:

```
project > user > org > platform > embedded
```

Higher-priority tiers shadow lower-priority ones by template name. Use this to discover which templates are available before creating a procession.

**Examples**:
```bash
# List available templates
ari procession list

# JSON output
ari procession list -o json
```

**Related Commands**:
- [`ari procession create`](#ari-procession-create) — Start a procession from a template

---

### ari procession create

Start a new procession from a named template.

**Synopsis**:
```bash
ari procession create [flags]
```

**Description**:
Starts a new cross-rite coordinated workflow from a named template. Resolves the template through the 5-tier resolution chain, creates the artifact directory, and stores the procession state in the active session context.

The procession ID is generated as `{template-name}-{YYYY-MM-DD}`.

Requires an active session.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--template` | string | required | Procession template name |

**Examples**:
```bash
# Start a security remediation procession
ari procession create --template=security-remediation

# List templates first to find the right one
ari procession list
ari procession create --template=my-workflow
```

**Related Commands**:
- [`ari procession list`](#ari-procession-list) — Discover available templates
- [`ari procession status`](#ari-procession-status) — Check current state

---

### ari procession status

Show the current procession state.

**Synopsis**:
```bash
ari procession status [flags]
```

**Description**:
Shows the current procession state for the active session, including:
- Current station
- Completed stations (with timestamps and artifact paths)
- Next station
- Artifact directory path

**Examples**:
```bash
# Check current procession state
ari procession status

# JSON output for scripting
ari procession status -o json
```

**Related Commands**:
- [`ari procession proceed`](#ari-procession-proceed) — Advance to the next station
- [`ari procession recede`](#ari-procession-recede) — Move back to an earlier station

---

### ari procession proceed

Advance to the next station.

**Synopsis**:
```bash
ari procession proceed [flags]
```

**Description**:
Advances the procession to the next station, recording the current station as completed. The current station is appended to `completed_stations` with a timestamp and optional artifact paths.

If there is no next station, the procession is complete. The procession block remains in the session context with an empty `next_station` field.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--artifacts` | string | - | Comma-separated artifact paths produced at this station |
| `--skip-validation` | bool | false | Skip handoff artifact frontmatter validation |

**Examples**:
```bash
# Advance without artifacts
ari procession proceed

# Advance with a handoff artifact
ari procession proceed --artifacts=.sos/wip/sr/HANDOFF-audit-to-assess.md

# Advance with multiple artifacts
ari procession proceed --artifacts=path1.md,path2.md

# Advance without frontmatter validation
ari procession proceed --skip-validation --artifacts=draft.md
```

**Related Commands**:
- [`ari procession status`](#ari-procession-status) — Check state before advancing
- [`ari procession recede`](#ari-procession-recede) — Move back if validation fails

---

### ari procession recede

Move back to an earlier station.

**Synopsis**:
```bash
ari procession recede [flags]
```

**Description**:
Moves the procession back to a named earlier station. The `completed_stations` log is **append-only** and is NOT modified — recede is a position change, not a state rollback.

Use this when a station needs to be re-executed (for example, when validation fails and remediation must be retried).

The `--to` station must:
- Exist in the template
- Appear before the current station in template order

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--to` | string | required | Station name to recede to |

**Examples**:
```bash
# Recede to the remediate station
ari procession recede --to=remediate

# Check status after receding
ari procession status
```

**Related Commands**:
- [`ari procession proceed`](#ari-procession-proceed) — Advance after re-executing
- [`ari procession status`](#ari-procession-status) — Check current position

---

### ari procession abandon

Terminate the active procession.

**Synopsis**:
```bash
ari procession abandon [flags]
```

**Description**:
Terminates the active procession. The session continues and can be used normally, but the cross-rite workflow coordination is removed. Artifact files in the artifact directory are NOT deleted.

Use this when the cross-rite workflow is no longer needed or was started in error.

**Examples**:
```bash
# Abandon the active procession
ari procession abandon

# Verify session continues normally
ari session status
```

**Related Commands**:
- [`ari session status`](cli-session.md#ari-session-status) — Confirm session still active

---

## Procession Flow

```
list templates → create → status → proceed → proceed → ... → complete
                                         ↕
                                      recede (if retry needed)
```

The `completed_stations` log is append-only. Receding does not erase history — it repositions the cursor so a station can be re-run.

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

- [Glossary: Procession](../../reference/GLOSSARY.md#procession) — Concept definition
- [`ari session status`](cli-session.md#ari-session-status) — Session state
