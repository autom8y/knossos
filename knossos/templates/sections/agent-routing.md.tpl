{{/* agent-routing section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START agent-routing -->
## Agent Routing

Delegate to specialists via Task tool.{{ if .IsKnossosProject }} Potnia coordinates phases and handoffs.{{ end }}
{{- if .IsKnossosProject }}
Without a session, execute directly or use `/task`. Routing guidance: `/consult`.
{{ end -}}
<!-- KNOSSOS:END agent-routing -->
