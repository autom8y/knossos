{{/* Inscription Master Template - Knossos Inscription System */}}
{{/* This template assembles all sections into the final context file (CLAUDE.md or GEMINI.md) */}}
{{/*
  Template variables available:
  - .ActiveRite   - Current rite name (e.g., "10x-dev")
  - .AgentCount   - Number of agents in active rite
  - .Agents       - List of AgentInfo (Name, Role, Produces, FilePath)
  - .KnossosVars  - Custom variables map for project-specific values
  - .Channel      - Target channel ("claude" or "gemini")

  Custom functions available:
  - include "path" - Include another template file
  - ifdef "var"    - Conditional include if variable is set
  - agents         - Load agent table data
  - term "key"     - Lookup terminology
  - channelDir     - Returns ".claude" or ".gemini"
  - toolName "cc"  - Returns channel-appropriate tool name
*/}}
{{ if eq .Channel "gemini" -}}
# GEMINI.md

> Entry point for Gemini CLI. Navigation pointers to on-demand context.
{{- else -}}
# CLAUDE.md

> Entry point for Claude Code. Navigation pointers to on-demand context.
{{- end }}

{{/* Core behavior (determines agent mode) */}}
{{include "sections/execution-mode.md.tpl"}}

{{/* Rite context (who is available) */}}
{{include "sections/quick-start.md.tpl"}}

{{include "sections/agent-routing.md.tpl"}}

{{include "sections/commands.md.tpl"}}

{{include "sections/agent-configurations.md.tpl"}}

{{/* Infrastructure pointer (how to access platform tools) */}}
{{include "sections/platform-infrastructure.md.tpl"}}

{{/* User customization (edit freely) */}}
{{include "sections/user-content.md.tpl"}}
