{{/* agent-table partial template */}}
{{/* Renders an agent table from .Agents context */}}
{{- if .Agents }}
| Agent | Role | Produces |
| ----- | ---- | -------- |
{{- range .Agents }}
| **{{ .Name }}** | {{ .Role }} | {{ .Produces }} |
{{- end }}
{{- else }}
_No agents configured._
{{- end }}
