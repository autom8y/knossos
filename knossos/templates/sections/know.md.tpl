{{/* know section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START know -->
## Codebase Knowledge

Persistent knowledge in `.know/`. Generate with `/know --all` if not present.

- `Read(".know/architecture.md")` — package structure, layers, data flow (read before code changes)
- `Read(".know/scar-tissue.md")` — past bugs, defensive patterns
- `Read(".know/design-constraints.md")` — frozen areas, structural tensions
- `Read(".know/conventions.md")` — error handling, file organization, domain idioms
- `Read(".know/test-coverage.md")` — test gaps, coverage patterns
{{- if .IsKnossosProject }}
- `Read(".know/literature-{domain}.md")` — external scholarship (generate with `/research`)
{{ end -}}
<!-- KNOSSOS:END know -->
