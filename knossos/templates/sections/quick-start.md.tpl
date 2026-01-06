{{/* quick-start section template */}}
{{/* Owner: regenerate - Generated from ACTIVE_RITE + agents/ */}}
<!-- KNOSSOS:START quick-start regenerate=true source=ACTIVE_RITE -->
## Quick Start

{{- if .ActiveRite }}
This project uses a {{ .AgentCount }}-agent rite ({{ .ActiveRite }}):

{{include "partials/agent-table.md.tpl"}}

**New here?** Use the `prompting` skill for copy-paste patterns, or `initiative-scoping` to start a new project.
{{- else }}
No active rite configured. Use `ari team switch <rite-name>` to activate a rite.

**New here?** Use `/consult` to get started.
{{- end }}
<!-- KNOSSOS:END quick-start -->
