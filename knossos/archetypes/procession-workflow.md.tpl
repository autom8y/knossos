---
name: {{.Name}}
description: "{{.Description}} Use when: starting or continuing a {{.Name}} procession, checking station progress. Triggers: /{{.Name}}, {{.Name}} workflow, {{.Name}} procession."
allowed-tools: Bash, Read, Write, Glob, Grep, Task, Skill, AskUserQuestion
model: opus
---

# /{{.Name}} -- Cross-Rite Procession Workflow

One command to coordinate a multi-rite workflow. `/{{.Name}}` detects your session and procession state, starts or resumes the workflow, consults Potnia for station orchestration, and handles transitions between stations and rites automatically.

## Station Map

{{.StationTable}}

Artifact directory: `{{.ArtifactDir}}`

## Pre-computed Context

The SessionStart hook has already injected session state as YAML frontmatter above. The `procession:` block (if present) contains: `id`, `type`, `current_station`, `completed_stations`, `next_station`, `next_rite`, `artifact_dir`.

## Your Task

$ARGUMENTS

## Behavior

### Phase 0 -- Pre-Flight

**CRITICAL**: All state detection uses the hook-injected YAML frontmatter above. Do NOT run `ari session status`, `ari rite status`, or other CLI commands to detect state — the frontmatter is the authoritative source and is always present.

Classify the current state by reading the hook-injected YAML frontmatter above:

**Case A: No session active** (`has_session: false`)
1. Create a session and procession in one shot:
   ```bash
   ari session create --initiative "{{.Name}}" --complexity INITIATIVE
   ari procession create --template={{.Name}}
   ```
2. If the current rite is NOT `{{.FirstRite}}`:
   ```bash
   ari sync --rite {{.FirstRite}}
   ```
   Tell the user: "Restart CC to begin at station **{{.FirstStation}}** ({{.FirstRite}} rite)."
   STOP here -- the user must restart CC to load the correct rite agents.
3. If the current rite IS `{{.FirstRite}}`: proceed to Execution.

**Case B: Active procession of type `{{.Name}}`** (check `procession.type` in frontmatter)
- Read `procession.current_station` and `active_rite` from frontmatter
- If `active_rite` matches the current station's rite: proceed to Execution
- If `active_rite` does NOT match: display procession status (see Status Display below) and tell user:
  ```
  You are stationed at {current_station} which runs on the {station_rite} rite.
  Run: ari sync --rite {station_rite}
  Then restart CC.
  ```

**Case C: Active procession of DIFFERENT type**
- Display: "Session has an active `{other_type}` procession. Abandon it first (`ari procession abandon`) or start a new session."

**Case D: Session active but no procession** (includes post-completion — completed processions are auto-removed)
- Create the procession:
  ```bash
  ari procession create --template={{.Name}}
  ```
- If current rite is NOT `{{.FirstRite}}`, sync and tell user to restart CC.
- Otherwise proceed to Execution.

### Status Display

When showing procession status (used in Case B and on request):

```
Procession: {{.Name}} ({procession_id})
Station:    {current_station} ({n}/{{.StationCount}})
Rite:       {station_rite}
Goal:       {station_goal}
Completed:  {list of completed stations with checkmarks}
Next:       {next_station} ({next_rite}) or "final station"
Artifacts:  {{.ArtifactDir}}
```

### Execution

Load station context and follow the execution protocol:

1. Load `Skill("{{.SkillName}}")` for station-specific context
2. Load `Skill("procession-ref")` then read the `execution-protocol.md` companion for the full orchestration protocol

Follow the protocol's three phases:
- **Phase 1 — Station Orchestration**: Read station context, load inbound handoff, consult Potnia
- **Phase 2 — Orchestration Loop**: Invoke specialists per Potnia's directives until station work is complete
- **Phase 3 — Station Completion**: Write handoff artifact, advance procession, handle rite transition

The protocol also covers: **Loop Handling** (Potnia-driven, `loop_to` is a hint), **Throughline Resumption** (store Potnia agent ID), and **Error Recovery**.

**Rite-context awareness**: The Potnia you are consulting is scoped to the current rite. When the procession transitions to a different rite, a new Potnia with different specialists will be available. The handoff artifact is the continuity mechanism -- it must be self-contained because the next station's Potnia has no shared context with this one.

## Anti-Patterns

- **Skipping Potnia**: ALWAYS consult Potnia for station orchestration. Never invoke specialists directly.
- **Hardcoding specialists**: The template is generic. Specialists come from the active rite, not this command.
- **Direct state mutation**: Use `ari procession proceed/recede` -- never write SESSION_CONTEXT.md directly.
- **Proceeding without handoff**: ALWAYS write the handoff artifact before calling `ari procession proceed`.
- **Ignoring rite mismatch**: If the current rite doesn't match the station's rite, show status and stop.
