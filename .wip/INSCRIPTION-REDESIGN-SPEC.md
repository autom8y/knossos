# Inscription Redesign Spec

**Sprint**: The Front Door
**Wave**: W3-1
**Status**: Design Complete

## Design Principles (from Sprint Decisions)

- P4: Comprehensive and correct over token minimization
- Must reflect Pythia as coordination identity
- Must reference Exousia as authority contract standard
- Must be correct about Fates-as-primitives architecture
- Progressive disclosure: inscription teaches current state, skills provide depth

## Current Structure (7 sections)

| # | Section | Owner | Lines | Assessment |
|---|---------|-------|-------|------------|
| 1 | execution-mode | knossos | 15 | Good. Add Pythia to Orchestrated mode. |
| 2 | quick-start | regenerate | 17 | Good. Add /go entry point. |
| 3 | agent-routing | knossos | 9 | Too terse. Needs Pythia role + Exousia explanation. |
| 4 | commands | knossos | 18 | Good as-is. Comprehensive primitive mapping. |
| 5 | agent-configurations | regenerate | 15 | Good as-is. Dynamic agent list. |
| 6 | platform-infrastructure | knossos | 8 | Too terse. Needs session lifecycle + /go reference. |
| 7 | user-content | satellite | 6 | User-safe zone. Never modify. |

## Redesigned Structure

### Section 1: execution-mode (MINOR UPDATE)

**Change**: Add Pythia identity to Orchestrated mode description.

```markdown
## Execution Mode

Three operating modes:

| Mode | Session | Rite | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Pythia coordinates; delegate via Task tool |

Use `/go` to start any session. Use `/consult` for mode selection.
```

**Delta**: "Coach pattern, delegate via Task tool" -> "Pythia coordinates; delegate via Task tool". Add /go reference.

### Section 2: quick-start (MINOR UPDATE)

**Change**: Add /go as cold-start entry point. Keep dynamic agent table.

```
## Quick Start

{{- if .ActiveRite }}
This project uses a {{ .AgentCount }}-agent workflow ({{ .ActiveRite }}):

{agent table}

Entry point: `/go`. Agent invocation patterns: `prompting` skill. Routing guidance: `/consult`.
{{- else }}
No active rite. Use `/go` to get started, or `ari rite switch <name>` to activate directly.
{{- end }}
```

**Delta**: Replace "Use `prompting` for agent invocation patterns. Use `/consult` for routing guidance." with tighter single-line format including /go.

### Section 3: agent-routing (SIGNIFICANT EXPANSION)

**Change**: Describe Pythia's role and Exousia contract. This is the most important section change.

```markdown
## Agent Routing

**Pythia** coordinates each rite's workflow — routing tasks to specialists, verifying phase gates, and managing handoffs. In orchestrated sessions, the main thread delegates to specialists via Task tool.

Every agent defines its authority via **Exousia** (jurisdiction contract):
- **You Decide**: Actions within the agent's autonomous authority
- **You Escalate**: Situations requiring Pythia or user input
- **You Do NOT Decide**: Boundaries the agent must never cross

Without a session, execute directly or use `/task`. Routing guidance: `/consult`.
```

**Delta**: From 3 lines to ~8 lines. Explains Pythia's role and Exousia in the always-on context where every conversation starts.

### Section 4: commands (NO CHANGE)

CC Primitives table stays as-is. Already comprehensive and correct.

### Section 5: agent-configurations (NO CHANGE)

Dynamic agent list stays as-is. Already reflects pythia.md via regeneration.

### Section 6: platform-infrastructure (MODERATE EXPANSION)

**Change**: Add session lifecycle with Fates reference and /go entry point.

```markdown
## Platform

**Entry**: `/go` — cold-start dispatcher. Detects session state, resumes parked work, or routes new tasks.

**Sessions**: Managed by Moirai agent via `/start`, `/park`, `/continue`, `/wrap`. Moirai loads Fate skills (Clotho/Lachesis/Atropos) for progressive context. Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: `ari --help`.
```

**Delta**: From 2 lines to ~5 lines. Adds /go, session lifecycle with Fates explanation, and autopark.

### Section 7: user-content (NO CHANGE)

User-safe zone. Preserved during sync.

## Token Impact

| Section | Before | After | Delta |
|---------|--------|-------|-------|
| execution-mode | ~80 tokens | ~90 tokens | +10 |
| quick-start | ~100 tokens | ~105 tokens | +5 |
| agent-routing | ~45 tokens | ~120 tokens | +75 |
| commands | ~150 tokens | ~150 tokens | 0 |
| agent-configurations | ~80 tokens | ~80 tokens | 0 |
| platform-infrastructure | ~40 tokens | ~85 tokens | +45 |
| user-content | ~20 tokens | ~20 tokens | 0 |
| **Total** | **~515 tokens** | **~650 tokens** | **+135** |

Net cost: ~135 additional tokens in every conversation's always-on context. This is well within P4's "comprehensive over minimized" guidance.

## Implementation Plan (W3-2)

1. Update `knossos/templates/sections/execution-mode.md.tpl`
2. Update `knossos/templates/sections/quick-start.md.tpl`
3. Update `knossos/templates/sections/agent-routing.md.tpl`
4. Update `knossos/templates/sections/platform-infrastructure.md.tpl`
5. Run `ari sync` to materialize
6. Verify CLAUDE.md output matches spec
