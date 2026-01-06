{{/* agent-configurations section template */}}
{{/* Owner: regenerate - Generated from agents/*.md */}}
<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agent Configurations

Full agent prompts live in `.claude/agents/`:

{{- if .Agents }}
{{- range .Agents }}
- `{{ .FilePath }}` - {{ .Role }}
{{- end }}
{{- else }}
_No agents installed. Use `ari team switch <rite-name>` to install agents._
{{- end }}
<!-- KNOSSOS:END agent-configurations -->
