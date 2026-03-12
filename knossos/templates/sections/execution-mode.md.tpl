{{/* execution-mode section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START execution-mode -->
## Execution Mode
{{ if .IsKnossosProject }}
Three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | {{ if eq .Channel "gemini" }}Potnia consulted via description matching{{ else }}Potnia coordinates; delegate via Task tool{{ end }} |

Use `/go` to start any session. Use `/consult` for mode selection.
{{ else }}
{{ if eq .Channel "gemini" -}}
Use the available agents and slash commands. Agents activate automatically when your prompt matches their description.
{{- else -}}
Use the available agents and slash commands. Delegate complex work to specialists via Task tool.
{{- end }}
{{ end -}}
<!-- KNOSSOS:END execution-mode -->
