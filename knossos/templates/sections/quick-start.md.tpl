{{/* quick-start section template */}}
{{/* Owner: regenerate - Generated from ACTIVE_RITE + agents/ */}}
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

{{- if .ActiveRite }}
{{ .AgentCount }}-agent workflow ({{ .ActiveRite }}):

{{include "partials/agent-table.md.tpl"}}
{{ if .IsKnossosProject -}}
Entry point: `/go`. Agent invocation patterns: `prompting` skill. Routing guidance: `/consult`.
{{- else if eq .Channel "gemini" -}}
Agents activate when your prompt matches their description.
{{- else -}}
Delegate to specialists via Task tool.
{{- end }}
{{- else }}
No active rite. Use {{ if .IsKnossosProject }}`/go` to get started, or {{ end }}`ari sync --rite=<name>` to activate{{ if .IsKnossosProject }} directly{{ end }}.
{{- end }}
<!-- KNOSSOS:END quick-start -->
