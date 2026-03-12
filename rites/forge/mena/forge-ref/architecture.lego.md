---
name: forge-ref-architecture
description: "Architecture for The Forge rite. Use when: understanding how Forge is structured, how ari sync handles rite activation, deploying Forge. Triggers: forge architecture, forge availability, ari sync rite."
---

# The Forge: Architecture

## Rite Structure

The Forge is a standard knossos rite. Activate it with `ari sync --rite forge`.

### How It Works

1. Forge agents live in `$KNOSSOS_HOME/rites/forge/agents/`
2. `ari sync --rite forge` projects agents to the channel agents directory
3. Invoke Forge agents via the Task tool from the main thread

### Why Forge?

- Rite creation is meta-level work (about rites, not within rites)
- Should be activated when building or extending the agent ecosystem
- Switch back to your working rite when Forge work is complete

## File Locations

| Type | Location |
|------|----------|
| Forge agents (source) | `$KNOSSOS_HOME/rites/forge/agents/` |
| Forge agents (projected) | channel agents directory (when forge rite is active) |
| Forge commands | channel commands directory |
| Patterns | `patterns/` |
| Evals | `evals/` |
