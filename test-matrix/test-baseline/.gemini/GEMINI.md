<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Potnia consulted via description matching |

Use `/go` to start any session. Use `/consult` for mode selection.
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START quick-start source=ACTIVE_RITE+agents regenerate=true -->
## Quick Start

Multi-agent workflow:

Entry point: `/go`. Agent invocation patterns: `prompting` skill. Routing guidance: `/consult`.
<!-- KNOSSOS:END quick-start -->

<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Agents activate automatically based on description matching. Write prompts that align with specialist descriptions for effective routing.
Without a session, execute directly. Routing guidance: `/consult`.
<!-- KNOSSOS:END agent-routing -->

<!-- KNOSSOS:START commands -->
## Gemini Primitives

| Primitive | Knossos Name | Invocation | Source |
|---|---|---|---|
| Slash command | **Dromena** | User types `/name` | `.gemini/commands/` |
| Skill | **Legomena** | Loaded into context | `.gemini/skills/` |
| Agent | **Agent** | Activates on description match | `.gemini/agents/` |
| Hook | **Hook** | Auto-fires on lifecycle events | `.gemini/settings.local.json` |
| GEMINI.md | **Inscription** | Always in context | `knossos/templates/` |
<!-- KNOSSOS:END commands -->

<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agents

Prompts in `.gemini/agents/`.
<!-- KNOSSOS:END agent-configurations -->

<!-- KNOSSOS:START platform-infrastructure -->
## Platform

**Entry**: `/go` — detects session state, resumes parked work, or routes new tasks.

**Sessions**: `/sos` (start, park, resume, wrap), `/handoff`, `/fray`. Mutate `*_CONTEXT.md` only via the moirai agent.

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: `ari --help`.
<!-- KNOSSOS:END platform-infrastructure -->