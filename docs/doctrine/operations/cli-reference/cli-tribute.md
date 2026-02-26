---
last_verified: 2026-02-26
---

# CLI Reference: tribute

> Generate and manage session summaries.

In Knossos mythology, King [Minos](../../reference/GLOSSARY.md#minos) demanded tribute from Athens. TRIBUTE.md is the "payment" for navigating the labyrinth—a comprehensive record of session accomplishments.

**Family**: tribute
**Commands**: 1
**Priority**: LOW

---

## Commands

### ari tribute generate

Generate TRIBUTE.md for a session.

**Synopsis**:
```bash
ari tribute generate [flags]
```

**Description**:
Generates a TRIBUTE.md summary document for a completed session. The tribute serves as both human-readable documentation and machine-parseable metadata.

**Content**:
- Session metadata (initiative, complexity, duration)
- Decisions made during session
- Artifacts produced
- Key events from clew
- Quality gate outcome

**Examples**:
```bash
# Generate tribute for current session
ari tribute generate

# Generate for specific session
ari tribute generate -s session-20260108-123456

# JSON output
ari tribute generate -o json
```

**Output Location**:
`[session-dir]/TRIBUTE.md`

**Related Commands**:
- [`ari session wrap`](cli-session.md#ari-session-wrap) — Generates tribute on wrap
- [`ari session audit`](cli-session.md#ari-session-audit) — View underlying events

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

- [Tribute Glossary Entry](../../reference/GLOSSARY.md#tribute)
- [Minos Glossary Entry](../../reference/GLOSSARY.md#minos)
