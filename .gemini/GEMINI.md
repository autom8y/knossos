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

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE+agents -->
## Quick Start

5-agent workflow (security):

| Agent | Role |
| ----- | ---- |
| **potnia** | Coordinates security initiative phases |
| **threat-modeler** | Models threats and identifies security risks and attack vectors |
| **compliance-architect** | Maps compliance requirements and designs control frameworks |
| **penetration-tester** | Executes penetration tests and documents vulnerabilities |
| **security-reviewer** | Performs final security review and grants deployment approval |

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

<!-- KNOSSOS:START agent-configurations source=agents/*.md regenerate=true -->
## Agents

Prompts in `.gemini/agents/`:

- `potnia.md` - Coordinates security initiative phases
- `threat-modeler.md` - Models threats and identifies security risks and attack vectors
- `compliance-architect.md` - Maps compliance requirements and designs control frameworks
- `penetration-tester.md` - Executes penetration tests and documents vulnerabilities
- `security-reviewer.md` - Performs final security review and grants deployment approval
<!-- KNOSSOS:END agent-configurations -->

<!-- KNOSSOS:START platform-infrastructure -->
## Platform

**Entry**: `/go` — detects session state, resumes parked work, or routes new tasks.

**Sessions**: `/sos` (start, park, resume, wrap), `/handoff`, `/fray`. Mutate `*_CONTEXT.md` only via the moirai agent.

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: `ari --help`.
<!-- KNOSSOS:END platform-infrastructure -->

<!-- KNOSSOS:START know -->
## Codebase Knowledge

Persistent knowledge in `.know/`. Generate with `/know --all` if not present.

- `read_file(".know/architecture.md")` — package structure, layers, data flow (read before code changes)
- `read_file(".know/scar-tissue.md")` — past bugs, defensive patterns
- `read_file(".know/design-constraints.md")` — frozen areas, structural tensions
- `read_file(".know/conventions.md")` — error handling, file organization, domain idioms
- `read_file(".know/test-coverage.md")` — test gaps, coverage patterns
- `read_file(".know/feat/INDEX.md")` — feature catalog and taxonomy (generate with `/know --scope=feature`)
- `read_file(".know/literature-{domain}.md")` — external scholarship (generate with `/research`)

Work product artifacts in `.ledge/`:

- `.ledge/decisions/` — ADRs and design decisions
- `.ledge/specs/` — PRDs and technical specs
- `.ledge/reviews/` — audit reports and code reviews
- `.ledge/spikes/` — exploration and research artifacts
<!-- KNOSSOS:END know -->

<!-- KNOSSOS:START user-content -->
## Project-Specific Instructions

<!-- Add project conventions, anti-patterns, and active work here.
     This section is preserved during sync. -->
<!-- KNOSSOS:END user-content -->