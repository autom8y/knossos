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

> Entry point for Claude Code. Skills-based progressive disclosure architecture.

{{/* Core navigation (read first) */}}
{{include "sections/execution-mode.md.tpl"}}

{{/* Team context (who is available) */}}
{{include "sections/quick-start.md.tpl"}}

{{include "sections/agent-routing.md.tpl"}}

{{include "sections/skills.md.tpl"}}

{{include "sections/agent-configurations.md.tpl"}}

{{/* Infrastructure (how things work) */}}
{{include "sections/hooks.md.tpl"}}

{{include "sections/dynamic-context.md.tpl"}}

{{include "sections/ariadne-cli.md.tpl"}}

{{/* Reference (consult as needed) */}}
{{include "sections/getting-help.md.tpl"}}

{{include "sections/state-management.md.tpl"}}

{{include "sections/slash-commands.md.tpl"}}

{{/* User customization (edit freely) */}}
{{include "sections/user-content.md.tpl"}}
