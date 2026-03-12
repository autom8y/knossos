# Context Loading Order

Per-sprint file list for Read() at session start. Ensures Potnia and agents begin each sprint with the right context loaded in the right order.

## Schema

```yaml
context_loading:
  sprint-1:
    - path: ".know/architecture.md"
      reason: "<why this file matters for this sprint>"
    - path: ".sos/wip/frames/<slug>.md"
      reason: "Initiative framing document"
  sprint-2:
    - path: ".ledge/spikes/<prior-artifact>.md"
      reason: "Sprint-1 exit artifact"
    - path: "<config-or-reference-file>"
      reason: "<why needed>"
```

## Guidance

- **Order matters.** List files in the order they should be read. Architecture and framing documents first, then prior sprint artifacts, then configuration files. Early reads establish mental models; later reads add specifics.
- **Name specific files, not directories.** `Read(".know/architecture.md")` is actionable. `Read(".know/")` is not -- agents cannot read directories.
- **Include the reason.** Context without purpose is noise. The `reason` field tells Potnia why this file matters for this specific sprint, enabling her to prioritize if context budget is tight.
- **Reference prior sprint exit artifacts.** Each sprint's context loading should include relevant exit artifacts from completed sprints. This is how continuity flows forward.
- **Be context-budget aware.** Each Read() call consumes context window. For high-turn sprints, prioritize essential files and note secondary references that agents can load on-demand via Skill tool.

## Anti-Patterns

- **Loading everything.** Listing every `.know/` file for every sprint wastes context. Load what is relevant to the sprint's mission.
- **Missing prior sprint artifacts.** If sprint-3 depends on sprint-2's exit artifact but the context loading does not include it, agents start without critical information.
