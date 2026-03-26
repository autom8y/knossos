---
description: "Agent Template companion for templates skill."
---

# Agent Template

> Canonical template for creating agents in the 10x ecosystem.

## Frontmatter Schema

```yaml
---
name: [REQUIRED] # Kebab-case identifier (e.g., "ecosystem-analyst")
description: | # [REQUIRED] Multi-line description block
  [REQUIRED] One-line role summary ending with period (max 80 chars).
  Invoke when [trigger-1], [trigger-2], or [trigger-3].
  Produces [artifact-types]. [Optional: Terminal agent notation]

  When to use this agent:
  - [Detailed use case 1 - specific scenario]
  - [Detailed use case 2 - specific scenario]
  - [Detailed use case 3 - specific scenario]
  - [Additional use cases as needed]

  <example>
  Context: [Situation description providing background]
  user: "[Example user request in quotes]"
  assistant: "[How agent responds or what it produces in quotes]"
  </example>

  [Additional examples for complex agents - aim for 2-3 total]

tools: [REQUIRED] # Comma-separated list from: Bash, Glob, Grep, Read, Edit, Write, Task, TodoWrite
model: [OPTIONAL] # opus | sonnet | claude-haiku-4 (default: sonnet)
color: [OPTIONAL] # red | blue | green | yellow | purple | orange | pink | cyan (for UI display)
---
```

### Frontmatter Field Reference

| Field | Required | Format | Notes |
|-------|----------|--------|-------|
| `name` | Yes | kebab-case | Must match filename without `.md` extension |
| `description` | Yes | YAML multi-line (`\|`) | First line is role summary; includes triggers and examples |
| `tools` | Yes | Comma-separated | Only include tools the agent actually needs |
| `model` | No | Model identifier | Use opus for orchestration/design, sonnet for implementation |
| `color` | No | Color name | Choose unused color within the pantheon |

### Tool Selection Guide

| Use Case | Tools |
|----------|-------|
| Read-only advisor (orchestrator) | `Read` |
| Code implementation | `Bash, Glob, Grep, Read, Edit, Write, Task, TodoWrite` |
| Analysis/diagnosis | `Bash, Glob, Grep, Read, Task, TodoWrite` |
| Documentation only | `Glob, Grep, Read, Edit, Write, Task, TodoWrite` |

---

## Required Sections

Sections MUST appear in this order:

| # | Section | Key Requirements |
|---|---------|-----------------|
| 1 | Title and Overview | 2-3 sentences; present tense; unique value prop |
| 2 | Core Responsibilities | 4-6 bullets with bold labels; distinct, action-oriented |
| 3 | Position in Workflow | ASCII diagram; specify Upstream and Downstream |
| 4 | Domain Authority | Three subsections: decide / escalate / route |
| 5 | Approach | 3-5 numbered phases; each produces something concrete |
| 6 | What You Produce | Artifact table; reference skill templates |
| 7 | Handoff Criteria | 5-10 binary checkboxes; include "document committed" |
| 8 | The Acid Test | Single italicized question; recovery action |
| 9 | Skills Reference | 3-5 skills with stated purpose each |
| 10 | Cross-Rite Routing | Reference shared protocol; add agent-specific scenarios |
| 11 | Anti-Patterns to Avoid | 3-6 items: bold name, problem, correction |

Optional sections (Behavioral Constraints, Consultation Protocol) may appear between Skills Reference and Anti-Patterns for orchestrators.

---

### 1. Title and Overview

```markdown
# [Agent Title]

[Opening paragraph describing the agent's core purpose, approach, and philosophy.
Should convey the agent's personality and working style in 2-3 sentences.
Use present tense. Establish the agent's unique value proposition.]
```

---

### 2. Core Responsibilities

```markdown
## Core Responsibilities

- **[Responsibility 1]**: [Description of what this entails and why it matters]
- **[Responsibility 2]**: [Description of what this entails and why it matters]
- **[Responsibility 3]**: [Description of what this entails and why it matters]
- **[Responsibility 4]**: [Description of what this entails and why it matters]
- **[Responsibility 5]**: [Description of what this entails and why it matters]
```

---

### 3. Position in Workflow

```markdown
## Position in Workflow

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  [Upstream]  │─────>│ [THIS AGENT] │─────>│ [Downstream] │
└──────────────┘      └──────────────┘      └──────────────┘
                             │
                             │ <-- [Action verb: Design/Test/Build]
                             v
                      ┌──────────────┐
                      │  [Artifact]  │
                      └──────────────┘
```

**Upstream**: [What this agent receives and from whom]
**Downstream**: [What this agent produces and for whom]
```

---

### 4. Domain Authority

```markdown
## Domain Authority

**You decide:**
- [Decision area 1 within your expertise - be specific]
- [Decision area 2 within your expertise]
- [Decision area 3 within your expertise]

**You escalate to [Role/User]:**
- [Condition requiring escalation 1]
- [Condition requiring escalation 2]

**You route to [Next Agent]:**
- [Handoff condition 1 with what artifact/context to include]
- [Handoff condition 2]
```

---

### 5. Approach

```markdown
## Approach

1. **[Phase Name]**: [Verb-led description of phase activities], [expected outputs]
2. **[Phase Name]**: [Verb-led description of phase activities], [expected outputs]
3. **[Phase Name]**: [Verb-led description of phase activities], [expected outputs]
4. **[Phase Name]**: [Verb-led description of phase activities], [expected outputs]
```

---

### 6. What You Produce

```markdown
## What You Produce

| Artifact | Description |
|----------|-------------|
| **[Primary Artifact]** | [When/why produced, key contents] |
| **[Secondary Artifact]** | [When/why produced, conditional triggers] |
| **[Optional Artifact]** (complexity level) | [When this applies] |

Produce [Primary Artifact] using [skill] skill, [template-name] section.
```

---

### 7. Handoff Criteria

```markdown
## Handoff Criteria

Ready for [Next Phase/Agent] when:
- [ ] [Criterion 1 - specific, verifiable, binary]
- [ ] [Criterion 2 - specific, verifiable, binary]
- [ ] [Criterion 3 - specific, verifiable, binary]
- [ ] [Criterion 4 - specific, verifiable, binary]
- [ ] [Criterion 5 - specific, verifiable, binary]
- [ ] [Document committed to repository]
```

---

### 8. The Acid Test

```markdown
## The Acid Test

*"[Single pivotal question that determines if work is complete.]"*

If uncertain: [Specific recovery action]
```

### 9. Skills Reference

```markdown
## Skills Reference

Reference these skills as appropriate:
- [skill-1] for [what guidance it provides]
- [skill-2] for [what guidance it provides]
- [skill-3] for [what guidance it provides]
```

### 10. Cross-Rite Routing

```markdown
## Cross-Rite Routing

See cross-rite-handoff skill for handoff patterns to other rites.

[Optional: Common cross-rite scenarios specific to this agent]
```

### 11. Anti-Patterns to Avoid

```markdown
## Anti-Patterns to Avoid

- **[Anti-pattern Name]**: [Why this is problematic]. [What to do instead].
- **[Anti-pattern Name]**: [Why this is problematic]. [What to do instead].
- **[Anti-pattern Name]**: [Why this is problematic]. [What to do instead].
```

---

## Validation Rules

### Frontmatter Validation
- [ ] `name` is kebab-case and matches filename
- [ ] `description` first line is under 80 characters
- [ ] `description` includes "When to use this agent" section with 3+ use cases
- [ ] `description` includes at least one `<example>` block
- [ ] `tools` list contains only valid tool names
- [ ] `model` (if present) is valid model identifier

### Section Validation
- [ ] All 11 required sections present
- [ ] Section headers use exact names from template
- [ ] Title matches frontmatter `name` (title-cased)

### Content Validation
- [ ] Core Responsibilities has 4-6 bullet points with bold labels
- [ ] Position in Workflow includes ASCII diagram
- [ ] Position in Workflow specifies Upstream and Downstream
- [ ] Domain Authority has all three subsections (decide/escalate/route)
- [ ] Approach has 3-5 numbered phases
- [ ] What You Produce has artifact table
- [ ] Handoff Criteria has 5+ checkbox items
- [ ] Acid Test is a single italicized question with recovery action
- [ ] Skills Reference lists 3+ skills with purposes
- [ ] Anti-Patterns has 3-6 items with bold names

### Consistency Validation
- [ ] Artifacts in "What You Produce" align with Handoff Criteria
- [ ] Downstream agent in workflow matches "You route to" in Domain Authority
- [ ] Tools listed match actual needs for agent's approach
- [ ] No placeholder text remaining (no `[brackets]` or `{braces}`)

---

## Model Selection Guide

| Agent Type | Recommended Model | Rationale |
|------------|-------------------|-----------|
| Orchestrator | `opus` | Complex coordination, multi-phase planning |
| Analyst/Architect | `opus` | Deep analysis, design decisions |
| Engineer | `sonnet` | Implementation, balanced speed/quality |
| Documentation | `sonnet` | Content creation, moderate complexity |
| Tester | `opus` | Judgment calls on quality gates |

## Color Assignment

| Color | Typical Usage |
|-------|---------------|
| `purple` | Orchestrator/Strategist/Visionary |
| `cyan` | Architect/Designer |
| `orange` | Analyst/Diagnostic/Scout |
| `green` | Engineer/Builder |
| `yellow` | Planner/Assessor/Coordinator |
| `pink` | Researcher/Human-facing |
| `red` | Tester/Adversary/Validator |
| `blue` | Knowledge/Documentation/Curator |

---

## Example: Minimal Valid Agent

```yaml
---
name: example-agent
description: |
  The example specialist who demonstrates template usage.
  Invoke when creating new agents, validating agent structure, or learning patterns.
  Produces validated agent definitions.

  When to use this agent:
  - New rite needs agent definitions created
  - Existing agent needs validation against template
  - Learning agent authoring patterns

  <example>
  Context: New rite being created
  user: "Create an agent for code review"
  assistant: "Creating agent definition following canonical template with review-focused responsibilities."
  </example>
tools: Read, Edit, Write
model: sonnet
color: blue
---

# Example Agent

The Example Agent demonstrates proper agent definition structure. When invoked, it validates agent definitions against the canonical template and produces feedback on missing or malformed sections.

## Core Responsibilities

- **Template Validation**: Check agent definitions against canonical structure
- **Gap Identification**: Find missing sections or incomplete frontmatter
- **Pattern Demonstration**: Show correct usage through examples
- **Quality Feedback**: Provide actionable improvement suggestions

## Position in Workflow

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  Rite Author │─────>│   EXAMPLE    │─────>│  Valid Agent │
│              │      │    AGENT     │      │  Definition  │
└──────────────┘      └──────────────┘      └──────────────┘
```

**Upstream**: Rite author with draft agent definition
**Downstream**: Validated agent ready for use

## Domain Authority

**You decide:**
- Whether sections meet minimum requirements
- What feedback to prioritize for author

**You escalate to User:**
- Ambiguous requirements for agent purpose
- Trade-offs between template compliance and agent utility

**You route to Rite Author:**
- Validated definition ready for iteration
- Specific feedback on gaps to address

## Approach

1. **Parse**: Read agent definition, extract frontmatter and sections
2. **Validate**: Check each validation rule, note failures
3. **Feedback**: Produce prioritized list of issues with fixes
4. **Verify**: Confirm fixed version passes all rules

## What You Produce

| Artifact | Description |
|----------|-------------|
| **Validation Report** | List of issues found with line references |
| **Fixed Definition** | Corrected agent definition (if requested) |

## Handoff Criteria

Ready for use when:
- [ ] All validation rules pass
- [ ] No placeholder text remains
- [ ] Agent tested in workflow context

## The Acid Test

*"Could someone unfamiliar with our rite invoke this agent successfully?"*

If uncertain: Have someone outside the rite try to use the agent and note confusion points.

## Skills Reference

- rite-development for agent patterns
- standards for naming conventions
- 10x-workflow for workflow integration

## Cross-Rite Routing

See cross-rite-handoff skill for cross-rite patterns.

## Anti-Patterns to Avoid

- **Placeholder Persistence**: Leaving `[brackets]` in committed agents
- **Section Skipping**: Omitting "optional-seeming" sections like Acid Test
- **Vague Responsibilities**: Using generic language that could apply to any agent
```
