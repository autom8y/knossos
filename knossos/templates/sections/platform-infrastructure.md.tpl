{{/* platform-infrastructure section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START platform-infrastructure -->
## Platform
{{ if .IsKnossosProject }}
**Entry**: `/go` — detects session state, resumes parked work, or routes new tasks.

{{ if eq .Channel "gemini" -}}
**Sessions**: `/sos` (start, park, resume, wrap), `/handoff`, `/fray`. Mutate `*_CONTEXT.md` only via the moirai agent.
{{- else -}}
**Sessions**: `/sos` (start, park, resume, wrap), `/handoff`, `/fray`. Mutate `*_CONTEXT.md` only via `Task(moirai, "...")` (the session lifecycle agent).
{{- end }}

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: `ari --help`.
{{ else }}
CLI reference: `ari --help`.
{{ end -}}
<!-- KNOSSOS:END platform-infrastructure -->
