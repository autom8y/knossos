{{/* quick-start section template */}}
{{/* Owner: regenerate - Generated from ACTIVE_RITE + agents/ */}}
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

{{- if .ActiveRite }}
This project uses a {{ .AgentCount }}-agent workflow ({{ .ActiveRite }}):

{{include "partials/agent-table.md.tpl"}}

Entry point: `/go`. Agent invocation patterns: `prompting` skill. Routing guidance: `/consult`.
{{- else }}
No active rite. Use `/go` to get started, or `ari rite switch <name>` to activate directly.
{{- end }}
<!-- KNOSSOS:END quick-start -->
