<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Use the available agents and slash commands. Agents activate automatically when your prompt matches their description.
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE+agents -->
## Quick Start

7-agent workflow (ui):

| Agent | Role |
| ----- | ---- |
| **potnia** | Coordinates UI development phases and routes work to specialists |
| **design-system-architect** | Defines token taxonomy, component architecture, and governance pipeline |
| **rendering-architect** | Determines per-route rendering strategy, hydration, and performance budgets |
| **stylist** | Translates design tokens into CSS architecture, layout patterns, and visual implementation |
| **component-engineer** | Implements components with state management, testing, and structured output |
| **a11y-engineer** | Validates WCAG 2.2 AA compliance and gates accessibility quality |
| **frontend-fanatic** | Browser-first visual auditing and UX evaluation through a designer's lens |

Agents activate when your prompt matches their description.
<!-- KNOSSOS:END quick-start -->

<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Agents activate automatically based on description matching. Write prompts that align with specialist descriptions for effective routing.
<!-- KNOSSOS:END agent-routing -->

<!-- KNOSSOS:START commands -->
## Gemini Primitives

| Primitive | Invocation | Source |
|---|---|---|
| Slash command | User types `/name` | `.gemini/commands/` |
| Skill | Loaded into context | `.gemini/skills/` |
| Agent | Activates on description match | `.gemini/agents/` |
| Hook | Auto-fires on lifecycle events | `.gemini/settings.local.json` |

<!-- KNOSSOS:END commands -->

<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agents

Prompts in `.gemini/agents/`:

- `potnia.md` - Coordinates UI development phases and routes work to specialists
- `design-system-architect.md` - Defines token taxonomy, component architecture, and governance pipeline
- `rendering-architect.md` - Determines per-route rendering strategy, hydration, and performance budgets
- `stylist.md` - Translates design tokens into CSS architecture, layout patterns, and visual implementation
- `component-engineer.md` - Implements components with state management, testing, and structured output
- `a11y-engineer.md` - Validates WCAG 2.2 AA compliance and gates accessibility quality
- `frontend-fanatic.md` - Browser-first visual auditing and UX evaluation through a designer's lens
<!-- KNOSSOS:END agent-configurations -->

<!-- KNOSSOS:START platform-infrastructure -->
## Platform

CLI reference: `ari --help`.
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