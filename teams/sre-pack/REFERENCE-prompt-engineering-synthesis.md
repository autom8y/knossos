# Prompt Engineering Best Practices Synthesis

> Reference document for sre-pack agent optimization sprint
> Sources: Anthropic Claude 4.x Best Practices, Claude Code Best Practices, Context Engineering Guide

---

## Core Principles

### 1. Be Explicit and Clear
Claude 4.x models respond to **precise instruction following**. Vague requests yield vague results.

**Anti-pattern**: "Handle incidents well"
**Best practice**: "Classify incident severity (SEV1-4), create incident channel, assign roles (IC/Technical Lead/Comms), set update cadence based on severity"

### 2. Provide Context/Motivation
Explain **why** instructions matter. Claude generalizes from understanding intent.

**Anti-pattern**: "Always verify file writes"
**Best practice**: "After every Write operation, verify the file exists via Read tool. This prevents hallucinated artifacts—agents sometimes report success without actual file creation."

### 3. Align Examples Carefully
Claude pays close attention to examples. Ensure they demonstrate desired behavior precisely.

**Anti-pattern**: Including outdated or inconsistent examples
**Best practice**: Curate diverse, canonical examples that show the expected behavior across different scenarios

### 4. Front-Load Critical Information
Signal density matters. Put the most important constraints first.

**Anti-pattern**: Burying critical constraints in verbose prose
**Best practice**: Start sections with the essential "You must" / "You must not" before elaboration

---

## Agent Structure Template

```markdown
---
name: agent-name
role: "Concise role statement (5-10 words)"
description: "When to invoke, what triggers this agent (2-3 sentences)"
tools: [Specific tools this agent uses]
model: claude-opus-4-5
color: color
---

# Agent Name

[2-3 sentence core purpose statement. What this agent does and why it exists.]

## Core Responsibilities

- **Responsibility 1**: Brief explanation
- **Responsibility 2**: Brief explanation
- [3-7 total, use active voice]

## Position in Workflow

[ASCII diagram showing upstream/downstream relationships]

**Upstream**: [What feeds into this agent]
**Downstream**: [What receives output from this agent]

## Domain Authority

**You decide:**
- [Decisions within this agent's authority]
- [Be specific: "Pipeline architecture" not "technical decisions"]

**You escalate to [Agent]:**
- [Decisions requiring higher authority]
- [Cross-cutting concerns]

**You route to [Agent]:**
- [Work to hand off]
- [Include clear handoff triggers]

## Approach

1. **Phase**: [Imperative action] - [What to produce/verify]
2. **Phase**: [Imperative action] - [What to produce/verify]
[Number phases sequentially, 4-6 typical]

## What You Produce

| Artifact | Description |
|----------|-------------|
| **Artifact Name** | What it contains, when produced |

## Handoff Criteria

Ready for [Next Phase/Agent] when:
- [ ] Criterion 1
- [ ] Criterion 2
[Use checkboxes for explicit verification]

## The Acid Test

*"[Question that validates this agent's work quality]"*

If uncertain: [Guidance for edge cases]

## [Domain] Patterns

[Domain-specific guidance, tables, code blocks as appropriate]

## Anti-Patterns to Avoid

- **Anti-pattern name**: Why it's problematic
- [3-5 critical anti-patterns]
```

---

## Quality Checklist

### Role Clarity (Score 1-5)
- [ ] Core purpose stated in 2-3 sentences
- [ ] Role distinguishable from other agents
- [ ] Clear trigger conditions for invocation
- [ ] Frontmatter description matches body content

### Instruction Precision (Score 1-5)
- [ ] Active voice throughout
- [ ] Imperative mood for instructions
- [ ] Specific actions, not vague guidance
- [ ] Quantified where possible (e.g., "3-5 bullet points" not "several")

### Constraint Completeness (Score 1-5)
- [ ] Domain authority explicitly defined (You decide / You escalate)
- [ ] Anti-patterns documented
- [ ] Handoff criteria are checkable
- [ ] Tool access matches actual needs

### Example Quality (Score 1-5)
- [ ] Examples demonstrate expected behavior
- [ ] Examples cover diverse scenarios
- [ ] No outdated or inconsistent examples
- [ ] Templates show concrete structure

### Structure Adherence (Score 1-5)
- [ ] All required sections present
- [ ] Consistent formatting with other agents
- [ ] Logical section ordering
- [ ] ASCII workflow diagrams accurate

### Token Efficiency (Score 1-5)
- [ ] No redundant content
- [ ] Concise without losing clarity
- [ ] High signal-to-noise ratio
- [ ] Under 300 lines

---

## Anti-Patterns to Avoid in Agent Prompts

### 1. Vague Role Boundaries
**Bad**: "Handles platform concerns"
**Good**: "Owns CI/CD pipelines, infrastructure as code, and developer environments"

### 2. Passive Voice
**Bad**: "Incidents should be classified by severity"
**Good**: "Classify incident severity (SEV1-4) immediately upon declaration"

### 3. Missing Escalation Paths
**Bad**: Lists responsibilities without decision authority
**Good**: Explicit "You decide" / "You escalate" sections

### 4. Redundant Content
**Bad**: Repeating the same constraint in multiple sections
**Good**: State once, reference if needed

### 5. Complexity Worship
**Bad**: Over-engineered workflows with 15+ steps
**Good**: 4-6 phase approach with clear deliverables

### 6. Missing "Acid Test"
**Bad**: No quality validation heuristic
**Good**: Single question that validates output quality

### 7. Tool Mismatch
**Bad**: Agent prompt references tools not in frontmatter
**Good**: Frontmatter tools match actual capability needs

### 8. Orphaned Handoffs
**Bad**: "Hand off to next phase"
**Good**: "Ready for Chaos Engineer when: [checklist]"

---

## Sources

- [Claude 4.x Best Practices](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-4-best-practices)
- [Claude Code Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices)
- [Context Engineering for AI Agents](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents)
