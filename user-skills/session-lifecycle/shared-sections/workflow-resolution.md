# Workflow Resolution Pattern

> Validate team context and agent availability.

## When to Apply

Commands that invoke agents or switch teams:
- /start - validates target team, may switch
- /resume - validates session team matches active team
- /handoff - validates target agent exists in team

## Validation Checks

| Check | Method | Pass | Fail |
|-------|--------|------|------|
| Team exists | `$KNOSSOS_HOME/teams/{team}` exists | Directory exists | Error: Team not found |
| Team matches session | Compare ACTIVE_RITE to session.active_team | Match | Warning + prompt |
| Agent exists | `.claude/agents/{agent}.md` exists | File exists | Error: Agent not found |

## Implementation

```
1. Read ACTIVE_RITE file
   - Path: .claude/ACTIVE_RITE
   - Returns: Current team pack name

2. If command specifies team change:
   a. Verify team exists in roster
   b. Invoke swap-team.sh
   c. Confirm ACTIVE_RITE updated

3. For session operations, check consistency:
   a. Read session.active_team from SESSION_CONTEXT
   b. Compare to ACTIVE_RITE
   c. If mismatch: Surface warning, offer switch or override

4. For agent invocation:
   a. Verify .claude/agents/{agent}.md exists
   b. If missing: Error with available agents list
```

## Error Messages

| Condition | Message Template |
|-----------|------------------|
| Team not found | "Team '{name}' not found. Use `/roster` to list available teams." |
| Team mismatch | "Session team ({session_team}) differs from active team ({active_team})." |
| Agent not found | "Agent '{agent}' not found in team '{team}'." |
| Roster unavailable | "Roster system unavailable. Set KNOSSOS_HOME or check installation." |

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `target_team` | Team to validate/switch to | start |
| `target_agent` | Agent to validate | handoff |
| `allow_override` | Allow continuing despite mismatch | resume |
