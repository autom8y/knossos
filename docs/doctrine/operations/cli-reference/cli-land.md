---
last_verified: 2026-03-26
---

# CLI Reference: land

> Synthesize and manage persistent knowledge from session archives.

`ari land` prepares the infrastructure for cross-session knowledge synthesis via the [Dionysus](../../reference/GLOSSARY.md#dionysus) agent. It enumerates archived sessions and prints an inventory, which you then use to invoke Dionysus via the Agent tool.

**Family**: land
**Commands**: 1
**Priority**: HIGH

---

## Commands

### ari land synthesize

Enumerate session archives for Dionysus knowledge synthesis.

**Synopsis**:
```bash
ari land synthesize [flags]
```

**Description**:
Prepares infrastructure for cross-session knowledge synthesis. This command:
1. Validates prerequisites (archives exist, Dionysus agent is available)
2. Enumerates archived sessions in `.sos/archive/`
3. Prints an inventory of available session data

The inventory guides construction of an Agent tool invocation for the Dionysus agent, which performs the actual synthesis. Run this command first to confirm archives exist, then delegate to Dionysus:

```
Task("dionysus", { domain: "scar-tissue", sessions: [...] })
```

**Supported domains**:

| Domain | Description |
|--------|-------------|
| `initiative-history` | Cross-session initiative outcomes and blockers |
| `scar-tissue` | Recurring bugs, root causes, and defensive patterns |
| `workflow-patterns` | Phase transitions, complexity calibration, rite usage |
| `all` | Synthesize all domains (default) |

Synthesis results land in `.sos/land/` as persistent cross-session knowledge.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--domain` | string | `all` | Domain to synthesize (initiative-history, scar-tissue, workflow-patterns, all) |

**Examples**:
```bash
# Enumerate archives for all domains
ari land synthesize

# Focus on scar-tissue synthesis
ari land synthesize --domain=scar-tissue

# JSON output for scripting
ari land synthesize -o json
```

---

## How Land Works

The `land` command does not perform synthesis itself — it prepares the context for the Dionysus agent. The workflow:

```
ari land synthesize          # Validates + enumerates archives
                             # → prints inventory
Task("dionysus", {...})      # Dionysus reads sessions → writes .sos/land/
ari knows                    # The synthesized knowledge appears here
```

This separation keeps the CLI lightweight while the LLM-intensive synthesis stays in the agent layer.

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

- [Glossary: Dionysus](../../reference/GLOSSARY.md#dionysus) — Knowledge synthesizer agent
- [Glossary: Naxos](../../reference/GLOSSARY.md#naxos) — Shore of abandonment (orphaned sessions)
- [`ari knows`](cli-knows.md) — Inspect the synthesized knowledge
- [`ari session wrap`](cli-session.md#ari-session-wrap) — Archive sessions for later synthesis
