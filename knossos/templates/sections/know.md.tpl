{{/* know section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START know -->
## Codebase Knowledge

Persistent codebase knowledge is available in `.know/`. Read domain files for pre-computed context:
- `Read(".know/architecture.md")` -- package structure, layers, entry points, abstractions, data flow
- `Read(".know/conventions.md")` -- naming patterns, error handling, test patterns, file organization
- `Read(".know/dependencies.md")` -- dependency graph, version currency, health signals
- `Read(".know/test-coverage.md")` -- test structure, coverage patterns, fixture patterns
- `Read(".know/api-surface.md")` -- CLI contracts, exported interfaces, public types

List available domains with `ari knows`. Refresh with `/know [domain]`. Check freshness with `ari knows --check`.
<!-- KNOSSOS:END know -->
