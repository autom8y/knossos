---
name: hygiene-ref:workflow-examples
description: "Hygiene workflow phases, usage examples, and operational notes. Triggers: hygiene workflow, quality audit session, hygiene examples."
---

# Hygiene Workflow and Examples

Usage examples, typical workflow phases, and operational notes for the hygiene rite.

---

## Examples

### Example 1: Basic Switch

```bash
/hygiene
```

Output:
```
[Knossos] Switched to hygiene (4 agents loaded)

Knossos:
  - code-smeller: Detects code smells and anti-patterns
  - architect-enforcer: Validates architectural compliance
  - janitor: Cleans up code, refactors for quality
  - audit-lead: Conducts comprehensive quality audits

Ready for code quality workflow.
```

### Example 2: Quality Audit Session

```bash
/hygiene
/start "Q4 Codebase Quality Audit" --complexity=PLATFORM
```

Output:
```
[Knossos] Switched to hygiene (4 agents loaded)
Session started: Q4 Codebase Quality Audit
Complexity: PLATFORM

Next: Audit Lead will coordinate comprehensive quality review.
```

### Example 3: Refactoring After Implementation

After completing feature with `/10x`:

```bash
/hygiene
/handoff smeller
```

Output:
```
[Knossos] Switched to hygiene (4 agents loaded)
Handing off to: code-smeller

Code Smeller analyzing recent implementation...
Detecting code smells and refactoring opportunities...
```

---

## Typical Workflow

### Phase 1: Detection
```bash
/hygiene
/start "Refactor authentication module" --complexity=MODULE
# Code Smeller identifies issues in auth module
# Produces: List of code smells with severity
```

### Phase 2: Audit
```bash
/handoff audit-lead
# Audit Lead reviews smells, prioritizes by impact
# Produces: Refactoring roadmap with effort estimates
```

### Phase 3: Enforcement Check
```bash
/handoff architect-enforcer
# Architect Enforcer validates current state vs ADRs
# Identifies: Architecture violations needing correction
```

### Phase 4: Cleanup
```bash
/handoff janitor
# Janitor executes refactoring plan
# Performs: Safe refactorings with tests passing
```

### Phase 5: Validation
```bash
/handoff audit-lead
# Audit Lead validates improvements
# Produces: Before/after metrics, completion report
```

### Phase 6: Completion
```bash
/wrap
```

---

## State Changes

### Files Modified

| File | Change | Description |
|------|--------|-------------|
| `.claude/ACTIVE_RITE` | Set to `hygiene` | Active rite state |
| `.claude/agents/` | Populated | 4 agent files loaded |
| `.claude/sessions/{session_id}/SESSION_CONTEXT.md` | `active_rite` updated | If session active |

### Success Criteria

- Rite switched to hygiene
- 4 agent files present in `.claude/agents/`
- Rite catalog displayed to user
- If session active, SESSION_CONTEXT updated

---

## Error Handling

If swap fails:

```
[Knossos] Error: Rite 'hygiene' not found
[Knossos] Use '/rite --list' to see available packs
```

**Resolution**: Verify knossos installation at `$KNOSSOS_HOME/`

---

## Notes

### Continuous vs Project-Based Hygiene

**Continuous hygiene** (recommended):
- Run Code Smeller on each PR
- Monthly Audit Lead reviews
- Janitor cleanups after feature completion

**Project-based hygiene**:
- Quarterly quality initiatives
- Pre-release cleanup sprints
- Major refactoring projects

Both valid, continuous prevents accumulation.

### Difference from /rite

| Command | Behavior |
|---------|----------|
| `/rite hygiene` | Switches rite, shows swap confirmation |
| `/hygiene` | Switches rite, shows rite catalog with agent descriptions |

Use `/hygiene` when you want to see available agents after switching.

### Quality Metrics

Hygiene rite can track:
- Cyclomatic complexity trends
- Code duplication percentage
- Test coverage
- Linter violation counts
- Architecture compliance score

Store metrics in session artifacts for historical comparison.
