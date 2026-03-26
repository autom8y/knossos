{{/* agent-configurations section template */}}
{{/* Owner: regenerate - Generated from agents/*.md */}}
<!-- KNOSSOS:START agent-configurations regenerate=true source=agents/*.md -->
## Agents
{{- if .Agents }}
Prompts in `{{ channelDir }}/agents/`:
{{- range .Agents }}
- `{{ .File }}` - {{ .Role }}
{{- end }}
{{- end }}
{{- if .SummonableAgents }}

### Summonable Heroes
Operational agents available on demand. Their commands handle the lifecycle:
{{- range .SummonableAgents }}
- **{{ .Name }}** - {{ .Role }} -> `{{ .Command }}`
{{- end }}

Summon: `ari agent summon {name}` then {{ if eq .Channel "gemini" }}restart Gemini Code Assist{{ else }}restart CC{{ end }}.
Dismiss: `ari agent dismiss {name}` then {{ if eq .Channel "gemini" }}restart Gemini Code Assist{{ else }}restart CC{{ end }}.
{{- end }}
{{- if not .Agents }}{{- if not .SummonableAgents }}
No agents installed. Run `ari sync --rite=<name>` to install.
{{- end }}{{- end }}
<!-- KNOSSOS:END agent-configurations -->
