---
last_verified: 2026-02-26
---

# CLI Reference: sails

> White Sails quality gate operations.

[White Sails](../../reference/GLOSSARY.md#white-sails) provides typed confidence signals for session completion: WHITE (high confidence), GRAY (unknown), BLACK (known failure).

**Family**: sails
**Commands**: 1
**Priority**: LOW

---

## Commands

### ari sails check

Check quality gate for a session.

**Synopsis**:
```bash
ari sails check [flags]
```

**Description**:
Evaluates the session against quality criteria and returns a confidence signal. Used before session wrap to determine if work is ready to ship.

**Exit Codes**:
- `0`: WHITE sails (high confidence, ship without QA)
- `1`: GRAY sails (unknown confidence, needs QA review)
- `2`: BLACK sails (known failure, do not ship)

**Examples**:
```bash
# Check current session
ari sails check

# Check specific session
ari sails check -s session-20260108-123456

# JSON output for scripting
ari sails check -o json
```

**Output Fields**:
- `signal`: WHITE, GRAY, or BLACK
- `confidence`: Confidence score (0-100)
- `criteria`: List of evaluated criteria
- `failures`: Any failing criteria
- `recommendations`: Suggested actions

**Related Commands**:
- [`ari session wrap`](cli-session.md#ari-session-wrap) — Wrap with quality gate
- [`ari validate artifact`](cli-validate.md#ari-validate-artifact) — Artifact validation

---

## Signal Meanings

| Signal | Meaning | Action |
|--------|---------|--------|
| **WHITE** | High confidence, verified | Ship without additional QA |
| **GRAY** | Uncertain confidence | Needs independent QA review |
| **BLACK** | Known failure | Do not ship, fix issues first |

---

## Anti-Gaming Rules

1. **No self-upgrade**: Agent cannot elevate its own GRAY to WHITE
2. **Proofs required**: WHITE requires verification evidence
3. **Independent review**: QA review can elevate GRAY → WHITE
4. **BLACK is final**: Only fixing issues can change BLACK

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

- [White Sails Glossary Entry](../../reference/GLOSSARY.md#white-sails)
- [Knossos Doctrine - Confidence Signal](../../philosophy/knossos-doctrine.md)
- [CLI: session wrap](cli-session.md#wrap) — Session completion with sails
