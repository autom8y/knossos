# Content Tone Guide

Examples and patterns for writing descriptive CLAUDE.md content.

---

## The Core Distinction

**Prescriptive content** tells Claude what to do. It removes agency and assumes a single mode of operation.

**Descriptive content** explains what's available and when patterns apply. It preserves agency and acknowledges multiple valid modes.

---

## Transformation Examples

### Execution Mode Section

**Prescriptive (Before):**
```markdown
## Execution Mode

**Active workflow?** MUST delegate via Task tool. See `orchestration/execution-mode.md`.
**No workflow?** May execute directly for single-phase work.
```

**Descriptive (After):**
```markdown
## Execution Mode

This project supports three operating modes:

| Mode | Session | Rite | Behavior |
|------|---------|------|----------|
| **Native** | No | - | Direct execution, no tracking |
| **Cross-Cutting** | Yes | No | Direct execution + session tracking |
| **Orchestrated** | Yes | Yes | Coach pattern, delegate via Task tool |

**Unsure?** Use `/consult` for routing guidance.

For enforcement rules: `orchestration/execution-mode.md`
```

**Why better:**
- Describes all three modes, not just two
- Uses conditional table showing when each applies
- Routes to `/consult` for uncertainty
- References enforcement rules instead of stating them

---

### Agent Routing Section

**Prescriptive (Before):**
```markdown
## Agent Routing

**Active workflow?** Delegate via Task tool. **No workflow?** Execute directly or use `/task`.
**Unsure?** Route to `/consult` for guidance.
```

**Descriptive (After):**
```markdown
## Agent Routing

When working within an orchestrated session, the main thread coordinates via Task tool delegation to specialist agents. Without an active session, direct execution or `/task` initialization are both valid approaches.

For routing guidance: `/consult`
```

**Why better:**
- Explains the pattern conceptually first
- Uses "when working within" (conditional) instead of imperatives
- Acknowledges both paths as valid
- Keeps `/consult` as guidance rather than command

---

### State Management Section

**Prescriptive (Before):**
```markdown
**Invocation Pattern** (MUST include session context):
```

**Descriptive (After):**
```markdown
**Invocation Pattern** (requires session context):
```

**Why better:**
- "Requires" describes a technical constraint
- "MUST" prescribes behavior
- Same information, different framing

---

## Pattern Library

### Conditional Tables

Use tables when behavior varies by state:

```markdown
| Condition | Pattern |
|-----------|---------|
| Session active + rite configured | Knossos workflow applies |
| Session active + no rite | Partial workflow, native patterns may apply |
| No session | Native Claude Code is fully valid |
```

### Explanatory Paragraphs

Use paragraphs to explain purpose before showing how:

```markdown
Skills provide domain knowledge on-demand. The main thread invokes skills
via the Skill tool when specialized context is needed. Skills never
execute directly; they inform the invoking agent.

Key skills: `10x-workflow` (coordination), `documentation` (templates),
`prompting` (agent invocation), `standards` (conventions).
```

### Escape Hatches

Always provide an escape hatch for uncertainty:

```markdown
**Unsure?** Use `/consult` for routing guidance.
```

---

## Words to Avoid vs Prefer

| Avoid | Prefer | Reason |
|-------|--------|--------|
| MUST | requires, expects | Describes constraint vs mandates behavior |
| NEVER | avoid, prefer not to | Acknowledges exceptions may exist |
| Always | typically, generally | Allows for context-dependent variation |
| You will | X applies when Y | Conditional rather than directive |
| Do X | X is used for Y | Describes purpose rather than commanding |

---

## Tone Checklist

Before writing CLAUDE.md content, verify:

- [ ] No global MUST/NEVER/ALWAYS in entry sections
- [ ] Conditional language ("when X, then Y") for variable behavior
- [ ] `/consult` referenced as routing escape hatch
- [ ] Enforcement referenced, not duplicated
- [ ] All three modes acknowledged where relevant
- [ ] Purpose explained before pattern shown

---

## The Litmus Test

Read your content and ask:

> "Does this describe what's available, or does it tell Claude what to do?"

If the latter, revise to describe when each pattern applies rather than prescribing which pattern to use.

---

## Related Files

- [INDEX.lego.md](INDEX.lego.md) - Parent skill with content architecture principles
- [anti-patterns.md](anti-patterns.md) - Content that should not appear
- [execution-mode.md](~/.claude/skills/orchestration/execution-mode.md) - Where enforcement rules live
