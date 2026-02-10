---
name: go
description: "Cold-start dispatcher — detects session state and gets you to productive work in seconds"
argument-hint: "[task description]"
allowed-tools: Bash, Read, Glob, Grep, Skill
model: opus
context: fork
---

## Purpose

`/go` is the single entry point for every work session. It reads system state, decides what to do, and dispatches to existing commands. It never duplicates logic that lives in `/start`, `/continue`, or `/consult`.

## Context

Auto-injected by SessionStart hook (project, rite, session, git).

## Your Task

Get the user to productive work immediately. $ARGUMENTS

## Phase 1: Collect State (parallel)

Run these in parallel to gather all state in under 3 seconds:

```bash
# 1. Session status
ari session status --output json 2>/dev/null

# 2. Session list (recent, any status)
ari session list --output json 2>/dev/null

# 3. Git state
git rev-parse --abbrev-ref HEAD 2>/dev/null
git status --porcelain 2>/dev/null | head -20

# 4. Active rite
cat .claude/ACTIVE_RITE 2>/dev/null || echo "none"

# 5. WIP artifacts
ls .wip/ 2>/dev/null
```

Also read the session context table injected by the SessionStart hook (already in your context above).

## Phase 2: Classify Scenario

Evaluate the collected state against exactly these four scenarios, in priority order:

| Priority | Scenario | Condition | Action |
|----------|----------|-----------|--------|
| 1 | ALREADY_ACTIVE | Session exists with status=ACTIVE | Show status, suggest next action |
| 2 | RESUME_PARKED | One or more sessions with status=PARKED | Auto-resume (or choose if multiple) |
| 3 | NEW_WORK | `$ARGUMENTS` contains a task description | Dispatch to `/consult` with intent |
| 4 | ORIENTATION | Nothing active, no arguments | Show dashboard with options |

**Priority matters.** If a session is ACTIVE, that is always Scenario 1 regardless of arguments. If a session is PARKED with no active session, that is Scenario 2. Only if no session exists at all do Scenarios 3-4 apply.

Exception: If a session is ACTIVE and the user provided `$ARGUMENTS`, acknowledge the active session AND relay the user's intent as the suggested next action.

## Phase 3: Execute Scenario

### Scenario 1: ALREADY_ACTIVE

Display a compact status block. No preamble. Get to the point.

```
Session: {session_id}
Rite: {active_rite} | Phase: {current_phase} | Branch: {branch}
Initiative: {initiative}

Sprint Progress:
- [x] Completed task
- [ ] Current task  <-- you are here
- [ ] Next task

Suggested next action: {inferred from phase and sprint state}
```

Read the session's `SESSION_CONTEXT.md` and any `SPRINT_CONTEXT.md` to populate the sprint progress. If sprint data is not available, show the phase and initiative without a task list.

Time target: ~3 seconds.

### Scenario 2: RESUME_PARKED

**Single parked session**: Resume immediately, no questions asked.

1. Invoke `/continue` (which delegates to Moirai for the state transition)
2. After resume, show the Scenario 1 status block

**Multiple parked sessions**: Ask ONE question, then act.

```
Found {n} parked sessions:

1. {session_id_1} — {initiative_1} (parked {relative_time})
2. {session_id_2} — {initiative_2} (parked {relative_time})

Which session? (number, or "new" to start fresh)
```

After the user answers, resume the chosen session via `/continue` or dispatch to Scenario 3/4.

**One question maximum.** Do not ask follow-ups.

Time target: ~8 seconds.

### Scenario 3: NEW_WORK

The user typed `/go add dark mode support` or similar. Dispatch their intent to `/consult` for rite routing and session creation.

1. Pass the user's intent to `/consult`: the arguments describe what they want to build
2. `/consult` handles rite recommendation and workflow routing
3. After routing, the user will invoke `/start` to create the session

Do NOT create the session yourself. Do NOT ask what rite to use. Let `/consult` handle the cognitive load.

If no rite is currently active, this is the expected path -- `/consult` is the universal pre-rite entry point.

Time target: ~12 seconds.

### Scenario 4: ORIENTATION

Nothing is happening. Show a terse dashboard and offer clear options.

```
Worktree: {worktree_name or "main"}
Branch: {branch_name}
Rite: {active_rite or "none"}
Sessions: {count active} active, {count parked} parked

Options:
  /go <task>     — Start new work (routes through /consult)
  /consult       — Get guidance on what to do
  /sessions      — Browse all sessions
  /rite --list   — See available rites
```

If `.wip/` artifacts exist without a session, mention them:

```
Note: Found {n} artifacts in .wip/ with no active session.
These may be from a previous session. Review with: ls .wip/
```

Time target: ~5 seconds.

## Behavioral Rules

1. **One question maximum.** If state is unambiguous, act. The only scenario that asks a question is RESUME_PARKED with multiple sessions.

2. **Dispatch, don't execute.** `/go` routes to `/start`, `/continue`, `/consult`. It does not contain session creation logic, resume logic, or rite routing logic.

3. **Read everything, ask almost nothing.** Collect all state in parallel, then decide. Never ask "do you have an active session?" -- check it yourself.

4. **No mythology.** Use plain operational vocabulary. Session, rite, phase, sprint. Not Clotho, Lachesis, Moirai.

5. **No preamble.** Do not say "Let me check your session state..." Just collect, decide, and present the result. The user should see output, not process narration.

6. **Terse output.** Tables and short lines, not paragraphs. Every extra line costs time-to-productive.

## Examples

```bash
# Active session exists — shows status immediately
/go

# Parked session — auto-resumes
/go

# New work with intent — dispatches to /consult
/go "add user authentication to the API"

# Nothing happening — shows dashboard
/go
```
