{{/* terminology-table partial template */}}
{{/* Renders the Knossos terminology mapping table with channel-appropriate primitive names */}}
{{ if eq .Channel "gemini" -}}
| Knossos | Gemini Primitive | Function |
|---------|-----------------|----------|
| **Dromena** | Slash command (`/name`) | {{ term "dromena" }} |
| **Legomena** | Skill (loaded into context) | {{ term "legomena" }} |
| **Agent** | Agent (description matching) | Heroes summoned for specific labors |
| **Rite** | — | {{ term "rites" }} |
| **Moirai** | — | {{ term "moirai" }} |
| **Inscription** | GEMINI.md | {{ term "inscription" }} |
| **Ariadne** | — | {{ term "ariadne" }} |
{{- else -}}
| Knossos | CC Primitive | Function |
|---------|-------------|----------|
| **Dromena** | Slash command (`/name`) | {{ term "dromena" }} |
| **Legomena** | Skill tool (`Skill("name")`) | {{ term "legomena" }} |
| **Agent** | Task tool (`Task(name)`) | Heroes summoned for specific labors |
| **Rite** | — | {{ term "rites" }} |
| **Moirai** | — | {{ term "moirai" }} |
| **Inscription** | CLAUDE.md | {{ term "inscription" }} |
| **Ariadne** | — | {{ term "ariadne" }} |
{{- end }}
