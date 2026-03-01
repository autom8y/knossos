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
| **Orchestrated** | Yes | Yes (ACTIVE) | Pythia coordinates; delegate via Task tool |

Use `/go` to start any session. Use `/consult` for mode selection.
{{ else }}
Use the available agents and slash commands. Delegate complex work to specialists via Task tool.
{{ end -}}
<!-- KNOSSOS:END execution-mode -->
