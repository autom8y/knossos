---
last_verified: 2026-02-26
---

# CLI Reference: hook

> Knossos hook infrastructure.

Hooks process harness tool events and can modify, validate, or enrich tool operations. The hook system provides Go-based implementations with consistent behavior.

**Family**: hook
**Commands**: 6
**Priority**: MEDIUM

---

## Commands

### ari hook clew

Record tool events on PostToolUse.

**Synopsis**:
```bash
ari hook clew [flags]
```

**Description**:
Records tool events to `events.jsonl` as part of the [clew](../../reference/GLOSSARY.md#clew) contract. This hook is called by the harness on PostToolUse events.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Examples**:
```bash
# Called by the harness hooks system
ari hook clew
```

**Related Commands**:
- [`ari session audit`](cli-session.md#ari-session-audit) — View recorded events

---

### ari hook context

Inject session context on SessionStart.

**Synopsis**:
```bash
ari hook context [flags]
```

**Description**:
Injects session context into Claude sessions at startup. Provides initiative, complexity, phase, and other session metadata.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Examples**:
```bash
# Called by the harness hooks system
ari hook context
```

**Related Commands**:
- [`ari rite context`](cli-rite.md#ari-rite-context) — Rite context injection

---

### ari hook autopark

Auto-park session on Stop event.

**Synopsis**:
```bash
ari hook autopark [flags]
```

**Description**:
Automatically parks the active session when the harness stops. Preserves session state for later resumption.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Examples**:
```bash
# Called by the harness hooks system
ari hook autopark
```

**Related Commands**:
- [`ari session park`](cli-session.md#ari-session-park) — Manual park

---

### ari hook route

Route slash commands on UserPromptSubmit.

**Synopsis**:
```bash
ari hook route [flags]
```

**Description**:
Routes slash commands (e.g., `/commit`, `/wrap`) to appropriate skill handlers on UserPromptSubmit events.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Examples**:
```bash
# Called by the harness hooks system
ari hook route
```

---

### ari hook validate

Validate bash commands against security rules.

**Synopsis**:
```bash
ari hook validate [flags]
```

**Description**:
Validates bash commands on PreToolUse to enforce security rules and prevent dangerous operations.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Examples**:
```bash
# Called by the harness hooks system
ari hook validate
```

---

### ari hook writeguard

Block direct writes to context files.

**Synopsis**:
```bash
ari hook writeguard [flags]
```

**Description**:
Intercepts Write/Edit operations targeting `*_CONTEXT.md` files and instructs use of [Moirai](../../reference/GLOSSARY.md#moirai) instead. Enforces centralized state authority.

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Examples**:
```bash
# Called by the harness hooks system
ari hook writeguard
```

**Related Commands**:
- [Moirai agent](../../reference/GLOSSARY.md#moirai) — Authorized context mutator

---

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--timeout` | int | 100 | Hook operation timeout in milliseconds (max 500) |

**Performance Targets**:
- Early exit: <5ms (when hooks disabled or no session)
- Full execution: <100ms (with all processing)

**Environment Variables**:
- `USE_ARI_HOOKS=0` — Emergency kill switch to disable ari hook implementations (default: enabled)
<!-- HA-CC: CLAUDE_HOOK_* are CC-specific environment variable names injected by the CC hooks system -->
- `CLAUDE_HOOK_*` — Standard harness hook variables (CC-specific prefix)

---

## See Also

- [Knossos Doctrine - Hooks](../../philosophy/knossos-doctrine.md)
- [Hook Glossary Entry](../../reference/GLOSSARY.md#hook)
