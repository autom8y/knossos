{{/* agent-routing section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START agent-routing -->
## Agent Routing

{{ if eq .Channel "gemini" -}}
Agents activate automatically based on description matching. Write prompts that align with specialist descriptions for effective routing.
{{- if .IsKnossosProject }}
Without a session, execute directly. Routing guidance: `/consult`.
{{- end }}
{{- else -}}
Delegate to specialists via Task tool.{{ if .IsKnossosProject }} Potnia coordinates phases and handoffs.{{ end }}
{{- if .IsKnossosProject }}
Without a session, execute directly or use `/task`. Routing guidance: `/consult`.
{{ end -}}
{{- end }}
<!-- KNOSSOS:END agent-routing -->
