<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Use the available agents and slash commands. Delegate complex work to specialists via Task tool.
<!-- KNOSSOS:END execution-mode -->

<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE+agents -->
## Quick Start

6-agent workflow (ecosystem):

| Agent | Role |
| ----- | ---- |
| **potnia** | Coordinates ecosystem infrastructure phases |
| **ecosystem-analyst** | Traces sync pipeline/knossos problems to root causes and produces gap analysis |
| **context-architect** | Designs context solutions, schemas, and ecosystem patterns |
| **integration-engineer** | Implements sync pipeline and knossos changes with integration tests |
| **documentation-engineer** | Creates migration runbooks and compatibility documentation |
| **compatibility-tester** | Validates ecosystem changes across satellite configurations |

Delegate to specialists via Task tool.
<!-- KNOSSOS:END quick-start -->

<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Delegate to specialists via Task tool.<!-- KNOSSOS:END agent-routing -->

<!-- KNOSSOS:START commands -->
## CC Primitives

| CC Primitive | Invocation | Source |
|---|---|---|
| Slash command | User types `/name` | `.claude/commands/` |
| Skill tool | Model calls `Skill("name")` | `.claude/skills/` |
| Task tool | Model calls `Task(subagent_type)` | `.claude/agents/` |
| Hook | Auto-fires on lifecycle events | `.claude/settings.json` |
Agents cannot spawn other agents — only the main thread has Task tool access.
<!-- KNOSSOS:END commands -->

<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agents

Prompts in `.claude/agents/`:

- `potnia.md` - Coordinates ecosystem infrastructure phases
- `ecosystem-analyst.md` - Traces sync pipeline/knossos problems to root causes and produces gap analysis
- `context-architect.md` - Designs context solutions, schemas, and ecosystem patterns
- `integration-engineer.md` - Implements sync pipeline and knossos changes with integration tests
- `documentation-engineer.md` - Creates migration runbooks and compatibility documentation
- `compatibility-tester.md` - Validates ecosystem changes across satellite configurations
<!-- KNOSSOS:END agent-configurations -->

<!-- KNOSSOS:START platform-infrastructure -->
## Platform

CLI reference: `ari --help`.
<!-- KNOSSOS:END platform-infrastructure -->

<!-- KNOSSOS:START know -->
## Codebase Knowledge

Persistent knowledge in `.know/`. Generate with `/know --all` if not present.

- `Read(".know/architecture.md")` — package structure, layers, data flow (read before code changes)
- `Read(".know/scar-tissue.md")` — past bugs, defensive patterns
- `Read(".know/design-constraints.md")` — frozen areas, structural tensions
- `Read(".know/conventions.md")` — error handling, file organization, domain idioms
- `Read(".know/test-coverage.md")` — test gaps, coverage patterns
- `Read(".know/feat/INDEX.md")` — feature catalog and taxonomy (generate with `/know --scope=feature`)
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