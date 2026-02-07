---
name: sprint
description: Multi-task sprint orchestration
argument-hint: <sprint-name> [--tasks="task1,task2,task3"]
allowed-tools: Bash, Read, Write, Task, Glob, Grep
model: opus
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

2. **Create SPRINT_CONTEXT** using the Write tool:

   Path: `$SESSION_DIR/SPRINT_CONTEXT.md`

   IMPORTANT: Use Write tool with YAML frontmatter format. The file MUST start with `---` on line 1.

   See `session-common` skill (mena/session/common/) for required fields and validation rules.

   Required fields:
   - Sprint metadata (links to session)
   - Task breakdown with status
   - Dependencies between tasks
   - Note: Uses session-scoped path to allow parallel sprints without collision

3. **For each task**, invoke `/task` workflow:
   - PRD for each task
   - TDD if MODULE+ complexity
   - Implementation
   - QA validation

4. **Track progress**:
   - Update task status (pending → in_progress → complete)
   - Handle blockers
   - Generate sprint burndown

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

## Reference

Full documentation: `mena/workflow/sprint.dro.md` (self-contained)
