---
description: "Transition Protocol companion for procession-ref skill."
---

# Transition Protocol

Three operations govern procession state transitions: proceed, recede, and abandon.

## Proceed

Advances the procession to the next station.

**Semantics**:
- Current station is appended to `completed_stations`
- `current_station` is set to the next station in the template sequence
- The handoff artifact becomes the incoming context for the new station

**Preconditions**:
- A handoff artifact MUST exist at `{artifact_dir}/HANDOFF-{current}-to-{next}.md`
- The handoff artifact frontmatter MUST pass schema validation
- The current station's goal should be satisfied (enforced by convention, not automation)

**CLI**: `ari procession proceed [--artifacts=path1,path2]`

**Post-transition**:
- User switches to the target rite: `ari sync --rite {next_rite}`
- The new rite's orchestrator reads the handoff artifact for station context
- The orchestrator's Procession Context section governs behavior within the new station

## Recede

Rolls back to a previous station for rework.

**Semantics**:
- `current_station` is set to the specified target station
- `completed_stations` is preserved as an append-only log (stations are not removed)
- Existing handoff artifacts are NOT deleted (preserved for reference)

**When to use**:
- Station work fails and cannot be completed with current context
- Validation reveals the prior station's output was insufficient
- A `loop_to` directive in the template signals iterative rework (e.g., validate -> remediate)

**CLI**: `ari procession recede --to=<station>`

**Post-transition**:
- If the target station uses a different rite, user switches: `ari sync --rite {target_rite}`
- The rework station should read the original handoff artifact plus any new context from the failing station

## Abandon

Terminates the procession entirely.

**Semantics**:
- Procession state is removed from session context
- Artifact directory is NOT deleted (preserved for forensics)
- Session continues without procession coordination

**When to use**:
- The initiative is cancelled or deprioritized
- The procession template no longer fits the actual workflow
- Unrecoverable failure that cannot be addressed by recede

**CLI**: `ari procession abandon`

**Post-transition**:
- Session returns to normal (non-procession) mode
- User may continue work in the current rite without station constraints

## Orchestrator Behavior During a Procession

When a procession is active, the rite orchestrator (Potnia, invoked via the procession dromena) reads the `procession:` block from session context and adjusts behavior:

- **Station awareness**: Reads `current_station` to understand scope
- **Incoming context**: Reads the handoff artifact from the previous station at `{artifact_dir}/HANDOFF-{previous}-to-{current}.md`
- **Acceptance criteria**: The handoff artifact's frontmatter `acceptance_criteria` define directional goals for this station
- **Station goal**: Comes from the procession template (provided in the dromena's Station Map)
- **Completion protocol**:
  1. Write `HANDOFF-{current}-to-{next}.md` with schema-valid frontmatter and self-contained body
  2. Signal Moirai to run `ari procession proceed`
  3. Tell user to switch rites: `ari sync --rite {next_rite}`
- **Cross-rite boundary**: Do NOT invoke agents from other rites -- they are not loaded in the current CC invocation
- **Failure path**: If work cannot be completed, signal Moirai to run `ari procession recede --to={previous_station}`

When no procession is active, the orchestrator ignores procession context entirely.

## Handoff Artifact Creation Protocol

When completing a station and preparing to proceed:

1. **Gather artifacts**: Collect all work products produced during the station
2. **Write handoff file**: Create `{artifact_dir}/HANDOFF-{current}-to-{next}.md`
   - Frontmatter: all required fields per the handoff schema
   - Body: self-contained summary of findings, guidance for next station, open questions
3. **Validate**: Ensure frontmatter passes schema validation (the proceed operation will check)
4. **Signal transition**: Use the named procession command (e.g., `/security-remediation`) or delegate to Moirai
5. **Inform user**: Display the next rite and the sync command

The handoff body must be self-contained -- the target station's agents may have no prior context about this procession's history.
