# Execution Protocol

> Generic station orchestration protocol for procession dromena. Values in `{braces}` refer to the invoking procession dromena's Station Map and frontmatter. Read them from the dromena that loaded this skill.

## Phase 1 -- Station Orchestration

Load the station context:

1. **Read current station** from `procession.current_station` in hook-injected frontmatter
2. **Read station goal** from the Station Map table in the invoking dromena (match by station name)
3. **Read inbound handoff** (if not the first station):
   ```bash
   # Check for handoff from previous station
   ls {artifact_dir}/HANDOFF-*-to-{current_station}.md 2>/dev/null
   ```
   If found, read the handoff artifact for context. Extract `acceptance_criteria` from its frontmatter.
4. **Load procession skill** for additional context:
   ```
   Skill("{skill_name}")
   ```

Construct a startup CONSULTATION_REQUEST and invoke Potnia:

```
Task(subagent_type="potnia", prompt="
## CONSULTATION_REQUEST

consultation:
  type: startup

  initiative:
    title: '{procession_name}: {current_station} station'
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
    Active procession: {procession_name} (station {n}/{station_count})
    Current station: {current_station}
    Station goal: {station_goal}
    Artifact directory: {artifact_dir}
    {If handoff artifact exists: include acceptance_criteria and key context from handoff body}
    {If no handoff: 'This is the first station — no inbound handoff.'}
")
```

Store Potnia's agent ID for throughline resumption.

## Phase 2 -- Orchestration Loop

Loop until Potnia returns `directive.action: complete`:

### On `invoke_specialist`:

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
       title: '{procession_name}: {current_station} station'
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

### On `await_user`:

1. Read `user_question.prompt` and `user_question.context`
2. Present to the user via AskUserQuestion
3. Feed the user's answer back to Potnia in the next consultation

### On specialist failure:

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

### On `complete`:

Potnia signals station work is done. Proceed to Phase 3.

## Phase 3 -- Station Completion

1. **Read procession state** to determine the next station:
   ```bash
   ari procession status
   ```

2. **Write handoff artifact** (if there is a next station):
   Create `{artifact_dir}/HANDOFF-{current_station}-to-{next_station}.md` with:

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
   ari procession proceed --artifacts={artifact_dir}/HANDOFF-{current}-to-{next}.md
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
   Procession {procession_name} COMPLETE

   Completed stations:
   {list all completed stations with timestamps}

   Artifacts: {artifact_dir}
   ```

## Loop Handling

Some stations have a `loop_to` hint (see Station Map in the invoking dromena). This means the station may need to iterate:

- When Potnia signals station work is done, but validation or review reveals issues:
  1. Consult Potnia: "Validation found issues. The station template has loop_to={loop_target}. Should we loop back?"
  2. If Potnia recommends looping:
     ```bash
     ari procession recede --to={loop_target}
     ```
  3. If `loop_target` is on a different rite, sync and tell user to restart CC
  4. If same rite, loop back to Phase 1

- The loop decision is ALWAYS Potnia-driven. `loop_to` is a hint, not automatic.

## Throughline Resumption

- Store Potnia's agent ID after the first consultation
- On subsequent consultations within the same station, use `resume="{potnia_agent_id}"`
- On rite switch (new CC session), start a fresh Potnia -- the handoff artifact carries the context
- If resume fails (agent expired), the consultation still works -- throughline is advisory

## Error Recovery

| Scenario | Action |
|----------|--------|
| No session | Create one (Case A in dromena Pre-Flight) |
| Wrong rite for station | Show status, tell user to sync (Case B) |
| Handoff artifact missing | Write it before proceeding |
| Specialist failure | Report to Potnia, follow recovery guidance |
| Potnia resume fails | Start fresh consultation (throughline is advisory) |
| Station goal unclear | Load skill via `Skill("{skill_name}")` for station details |
