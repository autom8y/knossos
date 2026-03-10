---
name: procession
description: "Manage cross-rite procession workflows. Use when: checking procession status, advancing to next station, rolling back to a previous station, abandoning a procession. Triggers: procession, station, cross-rite workflow, proceed, recede, abandon."
argument-hint: "[status|proceed|recede|abandon]"
allowed-tools: Bash, Read, Task
disallowed-tools: Write, Edit, NotebookEdit
model: opus
---

# /procession -- Cross-Rite Workflow Management

Manage a coordinated cross-rite procession: check status, advance stations, roll back, or abandon.

## Pre-computed Context

The SessionStart hook has already injected session state as YAML frontmatter above. The `procession:` block (if present) contains: `id`, `template`, `current_station`, `completed_stations`, `artifact_dir`.

## Your Task

$ARGUMENTS

## Routing

Parse $ARGUMENTS to determine the operation. If empty, default to **status** mode.

### Subcommand Table

| Input Pattern | Operation | Mutates State |
|---------------|-----------|---------------|
| (empty) | Status overview | No |
| `status` | Status overview | No |
| `proceed` | Advance to next station | Yes |
| `recede` | Roll back to previous station | Yes |
| `abandon` | Terminate procession | Yes |

### Decision Flow

1. **Extract first word** from $ARGUMENTS
2. **If it matches a subcommand** (status, proceed, recede, abandon): route directly
3. **If no match**: treat entire input as natural language, infer closest subcommand
4. **If empty**: run status mode

### Session ID Protocol

**CRITICAL**: Extract `session_id` and `procession.id` from the hook-injected YAML frontmatter above. You MUST pass these to Moirai for state-mutating operations.

If no `procession:` block exists in the frontmatter, report "No active procession" and suggest `ari procession create --template=<name>`.

## Operations

### Status (read-only)

Display procession overview via CLI:

```bash
ari procession status
```

Present to the user:
- **Procession**: template name and ID
- **Current station**: name, rite, goal
- **Completed stations**: list with checkmarks
- **Next station**: name, rite (if not at final station)
- **Artifact directory**: path

### Proceed (state-mutating)

Before proceeding, validate:

1. Read `procession.current_station` from session context
2. Read `procession.artifact_dir` to find the handoff artifact
3. **Verify handoff artifact exists**: `{artifact_dir}/HANDOFF-{current}-to-{next}.md`
4. If artifact is missing, tell the user what file is expected and STOP

If validation passes, dispatch via Moirai:
```
Task(moirai, "procession_proceed session_id=\"{session_id}\"")
```

After success, tell the user:
```
Station advanced: {current} -> {next}
Next rite: {next_rite}
Run: ari sync --rite {next_rite}
```

### Recede (state-mutating)

Ask the user which station to roll back to (from completed_stations list). Then dispatch:
```
Task(moirai, "procession_recede session_id=\"{session_id}\" to=\"{target_station}\"")
```

### Abandon (state-mutating)

Confirm with the user before abandoning. Then dispatch:
```
Task(moirai, "procession_abandon session_id=\"{session_id}\"")
```

## Dispatch Modes

**Via Task(moirai)** -- state-mutating operations (proceed, recede, abandon):
- These touch SESSION_CONTEXT.md and require Moirai's lifecycle management

**Via Bash** -- read-only operations (status):
- Direct CLI: `ari procession status`

## Sigils

### On Success

| Operation | Sigil |
|-----------|-------|
| status | (no sigil -- informational) |
| proceed | `(proceeded) next: ari sync --rite {next_rite}` |
| recede | `(receded) station: {target_station} -- resume work` |
| abandon | `(abandoned) procession terminated` |

### On Failure

`procession failed: {brief reason} -- fix: {recovery}`

## Anti-Patterns

- **Proceeding without handoff artifact**: ALWAYS verify the HANDOFF file exists before dispatching proceed
- **Direct state mutation**: NEVER write SESSION_CONTEXT.md directly -- use Task(moirai)
- **Skipping confirmation on abandon**: ALWAYS confirm with user before terminating a procession
- **Guessing station names**: Read them from session context, never fabricate
