{{/* agent-configurations section template */}}
{{/* Owner: regenerate - Generated from agents/*.md */}}
<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agents

{{- if .Agents }}
Prompts in `.claude/agents/`:

{{- range .Agents }}
- `{{ .File }}` - {{ .Role }}
{{- end }}
{{- else }}
No agents installed. Run `ari rite switch <name>` to install.
{{- end }}
<!-- KNOSSOS:END agent-configurations -->
