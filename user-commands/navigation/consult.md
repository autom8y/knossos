---
description: Get ecosystem guidance, team recommendations, and command-flows
argument-hint: [query] [--playbook=NAME] [--team] [--commands]
allowed-tools: Bash, Read, Grep, Glob, Task
model: opus
---

## Cognitive Load Absorber Role

`/consult` is the designated entry point for users who are confused, overwhelmed, or unsure where to start. When users don't know which team, workflow, or command to use, `/consult` absorbs that cognitive load by:

1. **Parsing natural language intent** - Understanding what the user wants to accomplish
2. **Matching to ecosystem capabilities** - Identifying the right team, workflow, and complexity
3. **Providing actionable patterns** - Referencing `prompting` skill for exact invocation commands
4. **Explaining the journey** - Referencing `10x-workflow` for phase transitions and quality gates

**Key principle**: Users describe outcomes; `/consult` handles routing complexity.

## Context

Auto-injected by SessionStart hook (project, team, session, git).

**Active team**: !`cat .claude/ACTIVE_TEAM 2>/dev/null || echo "none"`
**Available teams**: !`ls ${ROSTER_HOME:-~/Code/roster}/teams/ 2>/dev/null | tr '\n' ' '`

## Pre-flight

1. **Knowledge base accessible**:
   - Verify `${ROSTER_HOME:-~/Code/roster}/teams/` exists
   - If missing: WARN "Roster not found at expected location."

## Your Task

Provide ecosystem guidance and recommendations. $ARGUMENTS

## Behavior

### No arguments (general help)
1. Summarize current ecosystem state (active team, session status)
2. Display teams dynamically (use `team-discovery` skill for current count)
3. List common starting points based on user goals
4. Point to playbook library for detailed workflows

### Query provided (e.g., `/consult "improve code quality"`)
1. Parse user intent using knowledge base
2. Use `team-discovery` skill to match intent to team routing conditions
3. Match to appropriate team and workflow
4. **Reference invocation patterns**: Use `prompting` skill to retrieve exact copy-paste patterns
5. **Reference workflow context**: Use `10x-workflow` skill for phase/gate information
6. Provide command-flow with phases and decision points
7. Offer alternatives if multiple valid approaches exist

### `--playbook=NAME` flag
1. Load playbook from `~/.claude/knowledge/consultant/playbooks/curated/{NAME}.md`
2. Present complete workflow with current context
3. If not found, list available playbooks

### `--team` flag
Display all teams dynamically (use `team-discovery` skill):
```
| Team              | Command       | Best For                           |
|-------------------|---------------|------------------------------------|
| 10x-dev-pack      | /10x          | Full feature development lifecycle |
| debt-triage-pack  | /debt         | Technical debt prioritization      |
| doc-team-pack     | /docs         | Documentation, technical writing   |
| ecosystem-pack    | /ecosystem    | CEM/skeleton/roster infrastructure |
| forge-pack        | /forge        | Team pack creation and validation  |
| hygiene-pack      | /hygiene      | Code quality, refactoring          |
| intelligence-pack | /intelligence | Analytics, A/B testing, research   |
| rnd-pack          | /rnd          | Exploration, prototyping           |
| security-pack     | /security     | Security assessment, compliance    |
| sre-pack          | /sre          | Operations, reliability            |
| strategy-pack     | /strategy     | Market research, business analysis |
```
**Note**: This list reflects current roster inventory. Use `team-discovery` skill for programmatic access.

### `--commands` flag
Display all commands by category:
```
Session (5): /start, /park, /continue, /handoff, /wrap
Team (10): /team, /10x, /docs, /hygiene, /debt, /sre, /security, /intelligence, /rnd, /strategy
Workflow (4): /task, /sprint, /hotfix, /spike
Operations (5): /architect, /build, /qa, /pr, /code-review
```

## Response Format

Always structure responses as:

1. **Assessment**: What you understand the user needs
2. **Recommendation**: Which team/workflow fits best
3. **Command-Flow**: Step-by-step commands to execute
4. **Alternatives**: Other valid approaches (if any)

## Example Interactions

```bash
# General guidance
/consult

# Specific goal
/consult "I need to add user authentication"

# Load specific playbook
/consult --playbook=new-feature

# List all teams
/consult --team

# List all commands
/consult --commands
```

## Skill Reference Patterns

`/consult` references other skills to provide accurate, current information:

### Referencing prompting Skill

When providing invocation patterns, retrieve patterns from:
- `prompting/SKILL.md` - Agent invocation quick reference
- `prompting/patterns/discovery.md` - PRD creation, session initialization
- `prompting/patterns/implementation.md` - TDD, coding, testing

**Pattern**: Instead of generating ad-hoc commands, extract from `prompting` skill.

### Referencing 10x-workflow Skill

When explaining workflow journeys, retrieve context from:
- `10x-workflow/SKILL.md` - Agent routing, complexity calibration
- `10x-workflow/lifecycle.md` - Phase protocol (PLAN -> CLARIFY -> EXECUTE -> VERIFY -> HANDOFF)
- `10x-workflow/quality-gates.md` - Gate criteria per phase

**Pattern**: Include quality gate expectations when describing phase transitions.

### Referencing team-discovery Skill

When recommending teams, retrieve current team inventory from:
- `team-discovery` skill for dynamic team list
- `team-discovery/schemas/team-profile.yaml` for profile structure

**Pattern**: Never hardcode team counts or capabilities; always read from `team-discovery`.

## Knowledge Sources

- `$ROSTER_HOME/teams/*/orchestrator.yaml` - Team profiles (via team-discovery)
- `prompting` skill - Invocation patterns
- `10x-workflow` skill - Phase transitions and quality gates

## Reference

Full documentation: `.claude/skills/consult-ref/skill.md`
