---
name: sprint
description: Multi-task sprint orchestration
argument-hint: <sprint-name> [--tasks="task1,task2,task3"]
allowed-tools: Bash, Read, Task, Glob, Grep
model: opus
disable-model-invocation: true
---

## Context
Auto-injected by SessionStart hook (project, team, session, git).

## Pre-flight

1. **Workflow required**:
   - Verify `.claude/ACTIVE_WORKFLOW.yaml` exists
   - If missing: ERROR "No active workflow. Use /rite to select a rite first."

2. **Rite context**:
   - Verify `.claude/ACTIVE_RITE` exists
   - If missing: ERROR "No rite active. Use /rite <pack-name> to select a rite."

## Your Task

Plan and execute a sprint with multiple coordinated tasks. $ARGUMENTS

## Behavior

1. **Gather sprint parameters**:
   - Sprint name/goal
   - Task list (prompt if not provided)
   - Duration/timebox

2. **Create SPRINT_CONTEXT via Moirai**:

   Delegate to Moirai for sprint state creation:
   ```
   Task(moirai, "create_sprint name='<sprint-name>' goal='<sprint-goal>' tasks='<task1,task2,task3>'")
   ```

   Moirai will:
   - Create `SPRINT_CONTEXT.md` in the session directory with proper schema
   - Initialize task breakdown with status tracking
   - Set up sprint metadata linked to current session
   - Return confirmation with sprint path

3. **For each task**, invoke `/task` workflow:
   - PRD for each task
   - TDD if MODULE+ complexity
   - Implementation
   - QA validation

4. **Track progress** via Moirai:
   - Mark tasks complete: `Task(moirai, "mark_complete task_id='{id}'")`
   - Handle blockers and update task notes
   - Generate sprint burndown summary

5. **Complete sprint**:
   - Generate retrospective
   - Archive SPRINT_CONTEXT

## Example

```
/sprint "Authentication Sprint" --tasks="Login API,Session management,Password reset"
```

## When to Use

- 3+ related tasks for coordinated delivery
- Time-boxed work periods
- Feature epics spanning multiple components

## Parallel Sprint Pattern

For truly parallel sprints across multiple rites/focuses, use **worktrees**:

```bash
# Create isolated worktrees per sprint
/worktree create "sprint-backend" --rite=10x-dev
/worktree create "sprint-frontend" --rite=10x-dev
/worktree create "sprint-docs" --rite=docs

# In each terminal, navigate and start sprint independently:
# Terminal 1:
cd worktrees/wt-xxx && claude
/sprint "Backend Sprint" --tasks="API,Database,Auth"

# Terminal 2:
cd worktrees/wt-yyy && claude
/sprint "Frontend Sprint" --tasks="Components,State,Tests"

# Terminal 3:
cd worktrees/wt-zzz && claude
/sprint "Docs Sprint" --tasks="API Docs,User Guide,Examples"
```

**Why worktrees for parallel sprints?**
- Each sprint gets isolated SPRINT_CONTEXT (no collision)
- Different rites can work simultaneously
- Changes don't affect each other
- Use `/sessions --all` to monitor all sprints

**Single sprint, multiple tasks** → use this command directly
**Multiple parallel sprints** → use `/worktree` per sprint

