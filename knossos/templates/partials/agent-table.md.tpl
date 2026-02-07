{{/* agent-table partial template */}}
{{/* Renders an agent table from .Agents context */}}
{{- if .Agents }}
| Agent | Role |
| ----- | ---- |
{{- range .Agents }}
| **{{ .Name }}** | {{ .Role }} |
{{- end }}
{{- else }}
_No agents configured._
{{- end }}
