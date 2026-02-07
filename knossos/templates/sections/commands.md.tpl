{{/* commands section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START commands -->
## CC Primitives

| CC Primitive | Knossos Name | Invocation | Source |
|---|---|---|---|
| Slash command | **Dromena** | User types `/name` | `.claude/commands/` |
| Skill tool | **Legomena** | Model calls `Skill("name")` | `.claude/skills/` |
| Task tool | **Agent** | Model calls `Task(subagent_type)` | `.claude/agents/` |
| Hook | **Hook** | Auto-fires on lifecycle events | `.claude/settings.json` |
| CLAUDE.md | **Inscription** | Always in context | `knossos/templates/` |

Dromena have side effects and are user-controlled. Legomena are reference knowledge Claude loads autonomously.
Agents cannot spawn other agents — only the main thread has Task tool access.

Full mapping: `lexicon` skill. Dromena list: `.claude/commands/`. Legomena list: `.claude/skills/`.
<!-- KNOSSOS:END commands -->
