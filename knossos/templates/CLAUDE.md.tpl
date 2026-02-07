{{/* CLAUDE.md Master Template - Knossos Inscription System */}}
{{/* This template assembles all sections into the final CLAUDE.md */}}
{{/*
  Template variables available:
  - .ActiveRite   - Current rite name (e.g., "10x-dev-pack")
  - .AgentCount   - Number of agents in active rite
  - .Agents       - List of AgentInfo (Name, Role, Produces, FilePath)
  - .KnossosVars  - Custom variables map for project-specific values

  Custom functions available:
  - include "path" - Include another template file
  - ifdef "var"    - Conditional include if variable is set
  - agents         - Load agent table data
  - term "key"     - Lookup terminology
*/}}
# CLAUDE.md

> Entry point for Claude Code. Navigation pointers to on-demand context.

{{/* Core behavior (determines agent mode) */}}
{{include "sections/execution-mode.md.tpl"}}

{{/* Team context (who is available) */}}
{{include "sections/quick-start.md.tpl"}}

{{include "sections/agent-routing.md.tpl"}}

{{include "sections/commands.md.tpl"}}

{{include "sections/agent-configurations.md.tpl"}}

{{/* Infrastructure pointer (how to access platform tools) */}}
{{include "sections/platform-infrastructure.md.tpl"}}

{{/* User customization (edit freely) */}}
{{include "sections/user-content.md.tpl"}}
