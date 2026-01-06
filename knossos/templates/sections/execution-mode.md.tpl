{{/* execution-mode section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START execution-mode -->
## Execution Mode

This project supports three operating modes (see PRD-hybrid-session-model for details):

| Mode | Session | Rite | Main Agent Behavior |
|------|---------|------|---------------------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes (ACTIVE) | Coach pattern, delegate via Task tool |

**Unsure?** Use `/consult` for workflow routing.

For enforcement rules: `orchestration/execution-mode.md`
<!-- KNOSSOS:END execution-mode -->
