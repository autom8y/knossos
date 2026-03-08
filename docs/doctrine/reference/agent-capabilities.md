---
last_verified: 2026-02-26
---

# Agent Capabilities (CC-OPP)

> CC Operational Platform Properties — the capability uplift giving agents memory, skills, hooks, and resume.

---

## Overview

The CC-OPP uplift adds four capability dimensions to Knossos agents, declared via YAML frontmatter in agent `.md` files. The frontmatter schema is defined in `internal/agent/frontmatter.go` (AgentFrontmatter struct).

| Capability | Agents Enabled | Mechanism |
|------------|---------------|-----------|
| **Memory** | 17 | Persistent auto-memory directory per agent |
| **Skills** | 68 | Frontmatter `skills:` field, preloaded into agent context |
| **Hooks** | 10 | `ari hook agent-guard` enforcing tool restrictions |
| **Resume** | Ecosystem only | Throughline protocol via `resume: {agentId}` |

---

## Memory

Agents with memory get a persistent auto-memory directory at `~/.claude/projects/{project}/memory/`. They self-curate observations across sessions.

**Tiering:**
| Tier | Seeding | Agents | Example |
|------|---------|--------|---------|
| Tier 1 | Content-rich MEMORY.md | 4 | Potnia orchestrators (deep workflow knowledge) |
| Tier 2 | Structure-only MEMORY.md | 8 | Specialists with recurring patterns |
| Tier 3 | No seed (self-populate) | 5 | Lower-frequency agents |

**Constraints:**
- 150-line soft cap on MEMORY.md (self-curating)
- Topic files for detailed notes (linked from MEMORY.md)
- First invocation may have empty memory — acceptable degradation

---

## Skills

Agents declare skills they need preloaded via frontmatter:

```yaml
skills:
  - orchestrator-templates
  - session-common
  - cross-rite-handoff
```

**Design decisions:**
- ~3,500 token ceiling for preloaded skills per agent
- Aggressive preloading for agents without Skill tool access
- Minimal preloading for high-turn agents (Potnia) to preserve context
- Mixed dro/lego directories block skill resolution — 4 rites required directory splits as prerequisite
- `forge-ref` (346 lines) is Potnia-only — specialists use Skill tool on-demand

**Coverage:** 68 of 75 agents have `skills:` frontmatter. Exceptions: 3 agents in rites without mena, and agents where no skills apply.

---

## Hooks

Selective agents get tool restrictions enforced via `ari hook agent-guard`:

```yaml
hooks:
  - agent-guard
disallowedTools:
  - Write
```

**Triple-layer enforcement:**
1. `disallowedTools` in frontmatter — CC native restriction
2. `hooks: [agent-guard]` — runtime guard via `ari hook agent-guard`
3. Tool restriction text in agent prompt — behavioral instruction

**Coverage:** 10 agents across 5 rites. Applied to agents that should only Edit (not Write) or have other tool restrictions.

**Implementation:** `internal/cmd/hook/agentguard.go` (150 lines, 14 tests)

---

## Resume (Throughline Protocol)

Ecosystem Potnia supports conversation continuity across multiple Task invocations within a single CC session:

```
Task(potnia, "consultation prompt", resume: previousAgentId)
```

**How it works:**
- Main thread stores agent ID from Task tool result
- Subsequent calls pass `resume: {agentId}` to continue with full prior context
- If resume fails (session changed, invalid ID), falls back to fresh invocation

**Scope:** Ecosystem rite only. Cross-rite rollout deferred pending empirical evidence from ecosystem usage.

---

## Frontmatter Schema

Defined in `internal/agent/frontmatter.go`:

```go
type AgentFrontmatter struct {
    Skills          []string `yaml:"skills"`
    Hooks           []string `yaml:"hooks"`
    DisallowedTools []string `yaml:"disallowedTools"`
    Memory          *MemoryField `yaml:"memory"`
}
```

Frontmatter is parsed during `ari sync materialize` and used by:
- Agent transform (`internal/materialize/agent_transform.go`) — injects capabilities
- Agent-guard hook — enforces tool restrictions at runtime
- Skill resolution — preloads declared skills into agent context

---

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Hooks before memory | Safety enforcement precedes capability addition |
| Per-feature waves (not per-rite) | Cross-cutting rollouts scale better |
| Mixed dro/lego splits required | Skill resolution fails on mixed directories |
| Tier 1 seeds are manual | Not part of sync pipeline; created during enablement sprints |
| Resume is ecosystem-only | Needs empirical evidence before cross-rite rollout |

---

**See Also:**
- `internal/agent/frontmatter.go` — Frontmatter struct definition
- `internal/materialize/agent_transform.go` — Agent capability injection
- `internal/cmd/hook/agentguard.go` — Agent-guard hook implementation
- [architecture-map.md](architecture-map.md) — Subsystem overview
