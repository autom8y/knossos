{{/* commands section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START commands -->
## CC Primitives
{{ if .IsKnossosProject }}
| CC Primitive | Knossos Name | Invocation | Source |
|---|---|---|---|
| Slash command | **Dromena** | User types `/name` | `.claude/commands/` |
| Skill tool | **Legomena** | Model calls `Skill("name")` | `.claude/skills/` |
| Task tool | **Agent** | Model calls `Task(subagent_type)` | `.claude/agents/` |
| Hook | **Hook** | Auto-fires on lifecycle events | `.claude/settings.json` |
| CLAUDE.md | **Inscription** | Always in context | `knossos/templates/` |
{{ else }}
| CC Primitive | Invocation | Source |
|---|---|---|
| Slash command | User types `/name` | `.claude/commands/` |
| Skill tool | Model calls `Skill("name")` | `.claude/skills/` |
| Task tool | Model calls `Task(subagent_type)` | `.claude/agents/` |
| Hook | Auto-fires on lifecycle events | `.claude/settings.json` |
{{ end -}}
Agents cannot spawn other agents — only the main thread has Task tool access.
<!-- KNOSSOS:END commands -->
