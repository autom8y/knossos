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
3. If the current rite IS `{{.FirstRite}}`: proceed to Phase 1.

**Case B: Active procession of type `{{.Name}}`** (check `procession.type` in frontmatter)
- Read `procession.current_station` and `active_rite` from frontmatter
- If `active_rite` matches the current station's rite: proceed to Phase 1
- If `active_rite` does NOT match: display procession status (see Status Display below) and tell user:
  ```
  You are stationed at {current_station} which runs on the {station_rite} rite.
  Run: ari sync --rite {station_rite}
  Then restart CC.
  ```

**Case C: Active procession of DIFFERENT type**
- Display: "Session has an active `{other_type}` procession. Abandon it first (`ari procession abandon`) or start a new session."

**Case D: Session active but no procession**
- Create the procession:
  ```bash
  ari procession create --template={{.Name}}
  ```
- If current rite is NOT `{{.FirstRite}}`, sync and tell user to restart CC.
- Otherwise proceed to Phase 1.

**Case E: Procession complete** (no `next_station` in frontmatter)
- Display completion summary with all completed stations and artifacts.

### Status Display

When showing procession status (used in Cases B, E, and on request):

```
Procession: {{.Name}} ({procession_id})
Station:    {current_station} ({n}/{{.StationCount}})
Rite:       {station_rite}
Goal:       {station_goal}
Completed:  {list of completed stations with checkmarks}
Next:       {next_station} ({next_rite}) or "final station"
Artifacts:  {{.ArtifactDir}}
```

### Phase 1 -- Station Orchestration

Load the station context:

1. **Read current station** from `procession.current_station` in frontmatter
2. **Read station goal** from the Station Map table above (match by station name)
3. **Read inbound handoff** (if not the first station):
   ```bash
   # Check for handoff from previous station
   ls {{.ArtifactDir}}/HANDOFF-*-to-{current_station}.md 2>/dev/null
   ```
   If found, read the handoff artifact for context. Extract `acceptance_criteria` from its frontmatter.
4. **Load procession skill** for additional context:
   ```
   Skill("{{.SkillName}}")
   ```

**Rite-context awareness**: The Potnia you are consulting is scoped to the current rite ({{.FirstRite}} for the first station). When the procession transitions to a different rite, a new Potnia with different specialists will be available. The handoff artifact is the continuity mechanism -- it must be self-contained because the next station's Potnia has no shared context with this one.

Construct a startup CONSULTATION_REQUEST and invoke Potnia:

```
Task(subagent_type="potnia", prompt="
## CONSULTATION_REQUEST

consultation:
  type: startup

  initiative:
    title: '{{.Name}}: {current_station} station'
    goal: '{station_goal from Station Map}'

  state:
    current_phase: none
    completed_phases: []
    blocked_on: none

  results:
    last_specialist: none
    last_outcome: none
    artifacts_ready: []

  context_summary: |
    Active procession: {{.Name}} (station {n}/{{.StationCount}})
    Current station: {current_station}
    Station goal: {station_goal}
    Artifact directory: {{.ArtifactDir}}
    {If handoff artifact exists: include acceptance_criteria and key context from handoff body}
    {If no handoff: 'This is the first station — no inbound handoff.'}
")
```

Store Potnia's agent ID for throughline resumption.

### Phase 2 -- Orchestration Loop

Loop until Potnia returns `directive.action: complete`:

#### On `invoke_specialist`:

1. Read `specialist.agent` and `specialist.prompt` from Potnia's response
2. Invoke the specialist via Task tool:
   ```
   Task(subagent_type="{specialist.agent}", prompt="{specialist.prompt}")
   ```
3. When the specialist completes, summarize the result
4. Construct a continuation CONSULTATION_REQUEST and re-consult Potnia:
   ```
   Task(subagent_type="potnia", resume="{potnia_agent_id}", prompt="
   ## CONSULTATION_REQUEST

   consultation:
     type: continuation
     initiative:
       title: '{{.Name}}: {current_station} station'
       goal: '{station_goal}'
     state:
       current_phase: '{phase from Potnia}'
       completed_phases: [{phases completed so far}]
       blocked_on: none
     results:
       last_specialist: '{agent name}'
       last_outcome: success
       artifacts_ready:
         - '{artifact description}'
     context_summary: |
       {Brief summary of specialist results}
   ")
   ```

#### On `await_user`:

1. Read `user_question.prompt` and `user_question.context`
2. Present to the user via AskUserQuestion
3. Feed the user's answer back to Potnia in the next consultation

#### On specialist failure:

1. Construct a failure CONSULTATION_REQUEST:
   ```
   Task(subagent_type="potnia", resume="{potnia_agent_id}", prompt="
   ## CONSULTATION_REQUEST
   consultation:
     type: failure
     ...
     results:
       last_specialist: '{agent}'
       last_outcome: failure
       error_summary: '{what went wrong}'
   ")
   ```
2. Follow Potnia's recovery guidance (retry, skip, or escalate)

#### On `complete`:

Potnia signals station work is done. Proceed to Phase 3.

### Phase 3 -- Station Completion

1. **Read procession state** to determine the next station:
   ```bash
   ari procession status
   ```

2. **Write handoff artifact** (if there is a next station):
   Create `{{.ArtifactDir}}/HANDOFF-{current_station}-to-{next_station}.md` with:

   ```yaml
   ---
   type: handoff
   procession_id: {procession_id}
   source_station: {current_station}
   source_rite: {current_rite}
   target_station: {next_station}
   target_rite: {next_rite}
   produced_at: {ISO 8601 UTC}
   artifacts:
     - type: {artifact_type}
       path: {artifact_path}
   acceptance_criteria:
     - {criteria derived from station goal and Potnia's assessment}
   ---

   ## Summary

   {What was accomplished at this station}

   ## Key Findings

   {Important findings or decisions}

   ## Guidance for Next Station

   {Recommendations for the next station's work}

   ## Open Questions

   {Unresolved issues to carry forward}
   ```

3. **Advance the procession**:
   ```bash
   ari procession proceed --artifacts={{.ArtifactDir}}/HANDOFF-{current}-to-{next}.md
   ```

4. **Handle rite transition**:
   - Read the updated procession state to get `next_rite`
   - If `next_rite` differs from `active_rite`:
     ```bash
     ari sync --rite {next_rite}
     ```
     Tell user: "Station **{current}** complete. Restart CC to continue at **{next}** ({next_rite} rite)."
     STOP.
   - If same rite: display "Station **{current}** complete. Continuing to **{next}**."
     Loop back to Phase 1 for the next station.

5. **If final station** (no next station):
   Display procession completion summary:
   ```
   Procession {{.Name}} COMPLETE

   Completed stations:
   {list all completed stations with timestamps}

   Artifacts: {{.ArtifactDir}}
   ```

### Loop Handling

Some stations have a `loop_to` hint (see Station Map). This means the station may need to iterate:

- When Potnia signals station work is done, but validation or review reveals issues:
  1. Consult Potnia: "Validation found issues. The station template has loop_to={loop_target}. Should we loop back?"
  2. If Potnia recommends looping:
     ```bash
     ari procession recede --to={loop_target}
     ```
  3. If `loop_target` is on a different rite, sync and tell user to restart CC
  4. If same rite, loop back to Phase 1

- The loop decision is ALWAYS Potnia-driven. `loop_to` is a hint, not automatic.

### Throughline Resumption

- Store Potnia's agent ID after the first consultation
- On subsequent consultations within the same station, use `resume="{potnia_agent_id}"`
- On rite switch (new CC session), start a fresh Potnia — the handoff artifact carries the context
- If resume fails (agent expired), the consultation still works — throughline is advisory

## Anti-Patterns

- **Skipping Potnia**: ALWAYS consult Potnia for station orchestration. Never invoke specialists directly.
- **Hardcoding specialists**: The template is generic. Specialists come from the active rite, not this command.
- **Direct state mutation**: Use `ari procession proceed/recede` — never write SESSION_CONTEXT.md directly.
- **Proceeding without handoff**: ALWAYS write the handoff artifact before calling `ari procession proceed`.
- **Ignoring rite mismatch**: If the current rite doesn't match the station's rite, show status and stop.

## Error Recovery

| Scenario | Action |
|----------|--------|
| No session | Create one (Case A) |
| Wrong rite for station | Show status, tell user to sync (Case B) |
| Handoff artifact missing | Write it before proceeding |
| Specialist failure | Report to Potnia, follow recovery guidance |
| Potnia resume fails | Start fresh consultation (throughline is advisory) |
| Station goal unclear | Load skill via `Skill("{{.SkillName}}")` for station details |
