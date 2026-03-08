# Exousia Calibration from Tribal Knowledge

## What EX-NN Entries Are

The HANDOFF's `## Exousia Overrides from Tribal Knowledge` section contains EX-NN entries -- jurisdiction boundaries derived from real operational failures. Each entry specifies:

- **Applies to**: Which agent(s) the override targets
- **Boundary**: What the agent MUST or MUST NOT do
- **Rationale**: The scar or tribal knowledge that justifies the constraint

## Translation Rules

### "You Do NOT Decide" Constraints

Map EX-NN boundaries phrased as "MUST NOT" to the agent's Exousia `### You Do NOT Decide` section.

**Format**:
```markdown
### You Do NOT Decide
- {Boundary description} (per {EX-NN source})
```

**Example**:
```markdown
### You Do NOT Decide
- Destructive file operations on user-containing directories (per EX-03)
- Shared template modifications for rite-specific needs (per EX-04)
```

### "You Escalate" Triggers

Map EX-NN boundaries phrased as conditional actions ("MUST delegate", "MUST check first") to `### You Escalate`.

**Format**:
```markdown
### You Escalate
- {Condition} -> {escalation target} (per {EX-NN source})
```

**Example**:
```markdown
### You Escalate
- Protected file mutations -> Moirai agent (per EX-06)
- Architectural changes -> check ADRs first, escalate to Potnia if no guidance (per CK-06)
```

### All-Agent Overrides

When an EX-NN entry has `Applies to: ALL agents`:
- Add the constraint to EVERY agent's Exousia section
- Use identical wording across agents for consistency
- If more than 3 all-agent overrides exist, consider a shared `## Platform Constraints` section referenced by all agents instead of duplicating

## Calibration Checklist

For each agent, verify:

- [ ] Every EX-NN targeting this agent appears in Exousia
- [ ] Every EX-NN targeting ALL agents appears in Exousia
- [ ] "MUST NOT" boundaries are in "You Do NOT Decide"
- [ ] "MUST delegate/check" boundaries are in "You Escalate"
- [ ] Source IDs (EX-NN, TRIBAL-NNN) are preserved for traceability
- [ ] No EX-NN entries were silently dropped

## Conflict Resolution

When an EX-NN override conflicts with the RITE-SPEC's original Exousia definition:
- **EX-NN wins** -- tribal knowledge overrides spec-level assumptions
- Document the override with rationale: "Overridden by {EX-NN}: {reason}"
- Escalate to user if the conflict changes the agent's fundamental role
