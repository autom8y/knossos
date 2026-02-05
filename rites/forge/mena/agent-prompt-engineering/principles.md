# Prompt Engineering Principles

> 8 principles that separate working prompts from frustrating ones

## 1. Crystal Clear Role Identity

Establish the agent's identity in the first two sentences. The reader should know exactly who this agent is and what problem it solves.

**Why**: Prevents scope drift, clarifies intent, establishes authority boundaries.

**Pattern**:
```markdown
# Requirements Analyst

Turns ambiguity into specification. Extracts what users actually need from what they say they want.
```

**Anti-pattern (Vague Role)**:
```markdown
# Documentation Agent

This agent helps with documentation-related activities and supports the rite with various writing tasks.
```
Detection: First paragraph contains "helps with", "works on", "general-purpose", or "various tasks". Agent scope unclear, overlaps with other agents, users unsure when to invoke.

## 2. Actionable Responsibilities

Convert vague descriptions into specific, verifiable actions. Each responsibility starts with an action verb and describes a measurable outcome.

**Why**: Reduces hallucination, enables measurement, clarifies success criteria.

**Pattern**: Start with action verbs: Analyze, Produce, Validate, Orchestrate, Transform, Extract.

| Vague | Actionable |
|-------|------------|
| "Helps with code quality" | "Identify code smells using static analysis" |
| "Works on documentation" | "Transform technical specs into user guides" |
| "Reviews things" | "Validate API contracts against implementation" |

## 3. Explicit Boundaries

Define what this agent DOES and DOES NOT own. Every agent needs three boundary types: decisions it owns, conditions requiring escalation, and handoff triggers.

**Why**: Prevents scope creep, clarifies handoff points, reduces inter-agent confusion.

**Pattern**:
```markdown
**You decide:** Implementation details, refactoring approach, test structure
**You escalate:** API changes, new dependencies, architectural shifts
**You route to QA:** When implementation matches TDD and tests pass
```

**Anti-pattern (Implicit Escalation)**:
```markdown
## Domain Authority

**You decide:**
- Implementation approach
- Code structure
- Testing strategy
```
Detection: Missing "You escalate" section, or section lists only obvious cases. Agent makes decisions it should escalate; silent failures; wrong architectural choices propagate.

## 4. Front-Load Critical Instructions

Place the most important constraints and behaviors first. The agent reads top-to-bottom with attention decay.

**Why**: Better token economy, reduces exploration of irrelevant paths, improves first-try accuracy.

**Order**: Constraints first, then context, then examples. Not the reverse.

**Pattern**: Put "never do X" before "consider doing Y". Put critical requirements in the overview, not buried in subsections.

### 4.1 Extended Thinking Triggers

For complex reasoning tasks, add explicit triggers that engage extended thinking:
- **Orchestrators**: Multi-phase routing decisions
- **Architects**: Trade-off analysis and design choices
- **Adversarial agents**: Edge case discovery and attack surface analysis

**Trigger intensity by complexity**:

| Complexity | Trigger Phrase | Use Case |
|------------|----------------|----------|
| Standard | "think about" | Single-dimension analysis |
| Complex | "think hard about" | Multi-factor trade-offs |
| Critical | "ultrathink" | Architectural decisions, security analysis |

**Placement**: Add to "How You Work" section for analysis phases. Not for execution phases.

## 5. Active Voice, Specific Language

Use direct imperatives. Avoid passive constructions and hedge words.

**Why**: Clarity, reduces interpretation variance, shorter token count.

| Passive/Vague | Active/Specific |
|---------------|-----------------|
| "The code should be reviewed for potential issues" | "Review code for security vulnerabilities, performance bottlenecks, and API mismatches" |
| "Documentation might need updating" | "Update README when adding public functions" |
| "Consider checking for errors" | "Validate all inputs before processing" |

**Anti-pattern (Passive Instructions)**: Grep for "should be", "might need", "could", "consider". Agent interprets instructions differently across invocations, producing inconsistent outputs.

## 6. Measurable Handoff Criteria

Work is complete when specific, observable conditions are met. No subjective judgments.

**Why**: Prevents indefinite iteration, enables workflow coordination, provides clear completion signal.

**Pattern**: Checklist with objective conditions:
```markdown
Ready for implementation when:
- [ ] PRD contains at least 3 acceptance criteria
- [ ] All user-facing text reviewed for clarity
- [ ] Dependencies listed with version constraints
- [ ] Rollback procedure documented
```

**Anti-pattern (Fuzzy Handoff)**:
```markdown
## Handoff Criteria

Ready for next phase when:
- [ ] Documentation is complete
- [ ] Quality is acceptable
- [ ] Work meets standards
```
Detection: Criteria use "quality", "complete", "ready", "sufficient". Agent cycles indefinitely or signals completion prematurely.

## 7. Anti-Pattern Explicitness

State what NOT to do with specific examples. Agents learn boundaries through constraints.

**Why**: Prevents common failure modes, guides behavior through explicit limits.

**Pattern**:
```markdown
- **Do not** generate code during requirements phase. Route to Architect if design questions arise.
- **Do not** approve PRDs lacking rollback procedures. Flag and request revision.
```

**Anti-pattern (Missing Anti-Patterns)**:
```markdown
## Anti-Patterns

- Avoid making mistakes
- Don't produce low-quality work
- Be careful with edge cases
```
Detection: No Anti-Patterns section, or section lists generic advice. Agent repeats predictable mistakes, falls into domain-specific traps.

## 8. Minimal Redundancy

Remove explanations of concepts Claude already understands. Move repeated content to shared skills.

**Why**: Token efficiency without clarity loss. Shorter prompts are easier to maintain.

**Remove**:
- "As you know..."
- "It's important to note that..."
- Explanations of common programming concepts
- Identical content across multiple agents

**Pattern**: If content appears in 3+ agents, extract to a skill and reference it:
```markdown
## File Verification
See `file-verification` skill for artifact verification protocol.
```
One-line reference replaces 25+ lines of repeated verification instructions.

**Anti-pattern (Over-Explanation)**:
```markdown
## File Operations

When working with files, it's important to note that you should always verify files exist before reading them. As you know, file systems can have permission issues...

[25 more lines of basic file handling advice]
```
Detection: Count lines (250+ is a warning sign). Search for "As you know", "It's important to note". Token budget wasted, important instructions buried in noise.

## 9. Concrete Examples

Include examples that illuminate agent-specific behavior with realistic inputs and outputs.

**Why**: Examples teach expected format, depth, and agent-specific handling. They demonstrate edge cases.

**Pattern**:
```markdown
<example>
Context: User needs API reference for new payment endpoint
user: "Document the /payments/process endpoint for the developer portal"
assistant: "I'll create API reference documentation.

## POST /payments/process

Initiates a payment transaction.

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| amount | integer | Yes | Amount in cents |
| currency | string | Yes | ISO 4217 code |
..."
</example>
```

**Anti-pattern (Generic Examples)**:
```markdown
<example>
user: "Create documentation"
assistant: "I'll create documentation for you."
</example>
```
Detection: Example input/output could come from any agent. No workflow-specific context, no signal about expected format or depth.

---

## Applying the Principles

When writing or reviewing an agent prompt:

1. Read first two sentences. Is role identity crystal clear?
2. Scan responsibilities. Does each start with an action verb?
3. Check Domain Authority. Are decide/escalate/route explicit?
4. Review section order. Are constraints front-loaded?
5. Grep for passive voice. Convert to active.
6. Examine handoff criteria. Are all items objectively testable?
7. Look for anti-patterns section. Are they domain-specific?
8. Count repeated content. Can it reference a shared skill?
9. Check examples. Do they show this agent's actual work?

---

## Detection Reference

For quick anti-pattern detection patterns, see the **Anti-Pattern Summary** table in [INDEX.lego.md](INDEX.lego.md#anti-pattern-summary) (lines 171-184), which includes the "Core Issue" column explaining why each pattern matters.
