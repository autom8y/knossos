{{/* quick-start section template */}}
{{/* Owner: regenerate - Generated from ACTIVE_RITE + agents/ */}}
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

{{- if .ActiveRite }}
This project uses a {{ .AgentCount }}-agent workflow ({{ .ActiveRite }}):

{{include "partials/agent-table.md.tpl"}}

Use `prompting` for agent invocation patterns. Use `/consult` for routing guidance.
{{- else }}
No active rite. Run `ari rite switch <name>` to activate.

Use `/consult` to get started.
{{- end }}
<!-- KNOSSOS:END quick-start -->
