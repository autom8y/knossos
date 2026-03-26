---
last_verified: 2026-03-26
---

# CLI Reference: agent

> Validate, inspect, scaffold, and manage agent specifications.

Agent commands work with agent files across rites and the shared agents directory. Use them to validate agent frontmatter, list agents, scaffold new agents from archetypes, inspect an agent's full experiential context, and manage the summon/dismiss lifecycle for on-demand agents.

**Family**: agent
**Commands**: 8
**Priority**: HIGH

---

## Commands

### ari agent list

List agents with their metadata.

**Synopsis**:
```bash
ari agent list [flags]
```

**Description**:
Lists agents with their metadata from frontmatter. Without flags, lists all agents across all rites and the shared `agents/` directory.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | true | List all agents |
| `-r, --rite` | string | - | List agents in a specific rite |

**Examples**:
```bash
# List all agents
ari agent list

# List agents in the ecosystem rite only
ari agent list --rite ecosystem

# JSON output for scripting
ari agent list -o json
```

---

### ari agent roster

Show the full agent roster: standing, summoned, and available agents.

**Synopsis**:
```bash
ari agent roster [flags]
```

**Description**:
Displays three sections of the agent roster:

- **Standing**: Core platform agents (pythia, moirai, metis) — always active
- **Summoned**: Agents you have summoned with `ari agent summon`
- **Available**: Agents available to summon (`tier: summonable` in source)

This is the right first step before using `ari agent summon`.

**Examples**:
```bash
# Show full roster
ari agent roster

# JSON output for scripting
ari agent roster -o json
```

**Related Commands**:
- [`ari agent summon`](#ari-agent-summon) — Summon an available agent
- [`ari agent dismiss`](#ari-agent-dismiss) — Dismiss a summoned agent

---

### ari agent summon

Summon an agent to your user-level harness configuration.

**Synopsis**:
```bash
ari agent summon <name> [flags]
```

**Description**:
Summons a named agent to your user-level harness configuration (`~/.claude/agents/`). Only agents published with `tier: summonable` in their source frontmatter can be summoned. Standing agents (pythia, moirai, metis) cannot be summoned — they are always active.

This is the [Klesis](../../reference/GLOSSARY.md) operation — agent summoning via the Summonable Heroes tier.

**Arguments**:
- `name` (string, required): Agent name to summon

**Examples**:
```bash
# See available agents first
ari agent roster

# Summon the theoros agent
ari agent summon theoros

# Summon the naxos agent
ari agent summon naxos
```

**Related Commands**:
- [`ari agent roster`](#ari-agent-roster) — See available summonables first
- [`ari agent dismiss`](#ari-agent-dismiss) — Remove a summoned agent

---

### ari agent dismiss

Dismiss a summoned agent from your user-level harness configuration.

**Synopsis**:
```bash
ari agent dismiss <name> [flags]
```

**Description**:
Removes a previously summoned agent from your user-level harness configuration. Only agents summoned via `ari agent summon` can be dismissed. Standing agents (pythia, moirai, metis) and manually created agents are not affected.

This is the [Apolysis](../../reference/GLOSSARY.md) operation — agent dismissal.

**Arguments**:
- `name` (string, required): Agent name to dismiss

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Remove even if provenance is inconsistent |

**Examples**:
```bash
# Dismiss the theoros agent
ari agent dismiss theoros

# Check current roster after dismissal
ari agent roster

# Force dismiss if provenance check fails
ari agent dismiss theoros --force
```

**Related Commands**:
- [`ari agent summon`](#ari-agent-summon) — Summon an agent
- [`ari agent roster`](#ari-agent-roster) — See which agents are summoned

---

### ari agent embody

Show an agent's full experiential context from source.

**Synopsis**:
```bash
ari agent embody <agent-name> [flags]
```

**Description**:
Reconstructs an agent's full context as a first-person perspective view. Resolves identity, perception, capability, constraint, memory, position, surface, horizon, and provenance layers from **source files** (not materialized output). This captures all metadata including knossos-only fields stripped during materialization.

Use this when you need to understand exactly what an agent sees, what tools it has access to, and what constraints apply.

**Arguments**:
- `agent-name` (string, required): Agent name to inspect

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-r, --rite` | string | active rite | Rite to resolve the agent from |
| `--audit` | bool | false | Enable audit overlay with consistency checks |
| `--simulate` | bool | false | Enable simulate mode with capability mapping |
| `--simulate-prompt` | string | - | Natural language prompt for simulate mode (requires `--simulate`) |

**Examples**:
```bash
# Default perspective for potnia
ari agent embody potnia

# Inspect agent from a specific rite
ari agent embody principal-engineer --rite 10x-dev

# With audit overlay (consistency checks)
ari agent embody qa-adversary --audit

# Simulate what the agent would do
ari agent embody potnia --simulate --simulate-prompt "read a file"

# JSON output for tooling
ari agent embody potnia -o json
```

---

### ari agent new

Scaffold a new agent from an archetype template.

**Synopsis**:
```bash
ari agent new [flags]
```

**Description**:
Creates a new agent file from an archetype template. Archetypes provide default structure, sections, and platform content. Author sections are marked with TODO comments for you to fill in.

Available archetypes: `orchestrator`, `reviewer`, `specialist`

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-a, --archetype` | string | - | Archetype to use (orchestrator, reviewer, specialist) |
| `-n, --name` | string | - | Agent name (kebab-case) |
| `-r, --rite` | string | - | Rite to create the agent in |
| `-d, --description` | string | - | Agent description (optional) |

**Examples**:
```bash
# Create a specialist agent in the rnd rite
ari agent new --archetype specialist --rite rnd --name technology-scout

# Create a reviewer with description
ari agent new --archetype reviewer --rite security --name code-reviewer \
  --description "Reviews code for security issues"

# Create an orchestrator
ari agent new --archetype orchestrator --rite ecosystem --name coordinator
```

---

### ari agent update

Update platform-owned sections in agent files.

**Synopsis**:
```bash
ari agent update [path...] [flags]
```

**Description**:
Regenerates platform-owned and derived sections while preserving author content. Reads existing agent files, looks up the archetype from the `type` frontmatter field, regenerates platform sections from templates, and preserves all author-owned sections exactly as-is.

Use this after platform upgrades to pull in new section templates without losing your customizations.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | false | Update all agents in all rites |
| `-r, --rite` | string | - | Update all agents in a specific rite |
| `--dry-run` | bool | false | Show what would change without writing files |

**Examples**:
```bash
# Update a specific agent file
ari agent update rites/ecosystem/agents/potnia.md

# Update all agents in the ecosystem rite
ari agent update --rite ecosystem

# Preview changes before applying
ari agent update --dry-run --rite ecosystem

# Update all agents everywhere
ari agent update --all
```

---

### ari agent validate

Validate agent specifications against the agent JSON schema.

**Synopsis**:
```bash
ari agent validate [path...] [flags]
```

**Description**:
Validates agent frontmatter against the agent JSON schema. Reports validation errors with file and field context. In strict mode, also requires enhanced fields like `skills:`, `hooks:`, and `memory:`.

**Exit Codes**:
- `0` — All agents valid
- `1` — Validation errors found

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | false | Validate all agents in all rites and agents |
| `-r, --rite` | string | - | Rite to validate (all agents in that rite) |
| `--strict` | bool | false | Strict validation (requires enhanced fields) |

**Examples**:
```bash
# Validate all agents
ari agent validate

# Validate agents in the ecosystem rite
ari agent validate --rite ecosystem

# Strict validation (requires CC-OPP fields)
ari agent validate --strict

# Validate a specific agent file
ari agent validate agents/moirai.md

# Validate all rite agents by glob
ari agent validate rites/*/agents/*.md
```

**Related Commands**:
- [`ari lint`](cli-lint.md) — Broader source linting (agents, dromena, legomena)

---

## Agent Tiers

Knossos agents fall into three tiers:

| Tier | Examples | Behavior |
|------|----------|----------|
| **Standing** | pythia, moirai, metis | Always active; cannot be summoned/dismissed |
| **Rite** | potnia, specialist agents | Active when rite is materialized |
| **Summonable Heroes** | theoros, naxos | On-demand; materialized via `ari agent summon` |

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

- [Glossary: Heroes](../../reference/GLOSSARY.md#heroes) — Agent tier definitions
- [Glossary: CC-OPP](../../reference/GLOSSARY.md#cc-opp) — Agent capability frontmatter
- [agent-capabilities.md](../../reference/agent-capabilities.md) — CC-OPP capability details
