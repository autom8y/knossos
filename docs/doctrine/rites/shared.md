---
last_verified: 2026-02-26
---

# Rite: shared

> Cross-rite resources inherited by all rites.

The shared rite is not a workflow — it provides agents and mena (skills/commands) that are available across all rites via the overlay mechanism in materialization.

---

## Overview

| Property | Value |
|----------|-------|
| **Name** | shared |
| **Form** | Cross-rite overlay |
| **Agents** | 1 |
| **Entry Agent** | — (not directly invokable) |

---

## Contents

### Agents

| Agent | Role |
|-------|------|
| **theoros** | Domain evaluator agent for theoria audit operations |

### Mena

The shared mena directory (`rites/shared/mena/`) provides skills and commands inherited by all rites, including:
- Session lifecycle schemas and patterns
- Cross-rite handoff templates
- Orchestrator consultation templates
- Pinakes domain registry (theoria audit criteria)

---

## Source

**Manifest**: `rites/shared/manifest.yaml`

---

## See Also

- [Theoria](../philosophy/knossos-doctrine.md#the-delegation-theoria) — Audit operation using shared agents
- [Knossos Doctrine - Rites](../philosophy/knossos-doctrine.md#iv-the-rites)
