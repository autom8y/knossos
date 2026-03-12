---
user-invocable: false
---

# /sos Subcommand Specifications

## park

**Preconditions**: Session exists, status ACTIVE
**Parameters**: reason (required — prompt if not provided)
**Delegation**: `Task(moirai, "park_session reason=\"{reason}\" session_id=\"{id}\"")`
**Output**: Duration so far, progress summary, park reason, resume hint

## resume

**Preconditions**: Session exists, status PARKED
**Parameters**: none
**Delegation**: `Task(moirai, "resume_session session_id=\"{id}\"")`
**Output**: Resumed status, current phase, next action hint

## start

**Preconditions**: No active session (or offer to park existing)
**Parameters**: initiative (required), --complexity (optional, default PATCH), --rite (optional)
**Delegation**: `Task(moirai, "create_session initiative='{initiative}' complexity={complexity}")`

### PATCH Fast Path (default for /sos start)

When complexity is PATCH:
1. Create session via Moirai (Clotho)
2. Do NOT invoke entry agent
3. Do NOT generate PRD
4. Output: "Session {id} started (PATCH). Work directly — no PRD needed."

When complexity is MODULE or higher:
1. Create session via Moirai (Clotho)
2. Optionally switch rite if --rite specified
3. Invoke entry agent via workflow entry point
4. Output: Standard session-start output

### Existing-Session Handling

If a session exists when `/sos start` is invoked:
- ACTIVE: "Session already active. `/sos park` first, or `/fray` for parallel work."
- PARKED: "Session parked ({reason}). Options: `/sos resume`, `/sos wrap`, or park + start."

## wrap

**Preconditions**: Session exists, status ACTIVE or PARKED
**Parameters**: --force (optional — bypass quality gates)
**Delegation**: `Task(moirai, "wrap_session session_id=\"{id}\" [--force]")`
**Output**: Session summary, archive location. Suggest `/land` for knowledge synthesis.

**Important**: /sos wrap is the minimal wrap. For wrap-with-synthesis, use `/land` instead.
The wrap subcommand does NOT invoke Dionysus. This separation is intentional:
wrap is a state mutation (Atropos); synthesis is knowledge extraction (Dionysus).
