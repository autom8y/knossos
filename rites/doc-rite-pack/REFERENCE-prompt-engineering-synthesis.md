# Prompt Engineering Synthesis for Agent Optimization

> Distilled best practices from Anthropic Claude 4.x documentation, Claude Code best practices, and context engineering principles.

## Core Principles for Claude 4.x Agents

### 1. Be Explicit and Specific

Claude 4.x models take instructions literally. Previous models inferred intent and expanded on vague requests; Claude 4.x does exactly what you ask.

**Implications for agents:**
- Write precise instructions, not general guidance
- Explicitly request behaviors you want (don't assume Claude will go "above and beyond")
- Specify edge cases, constraints, and expected outputs directly

### 2. Provide Context and Motivation

Explaining *why* helps Claude understand goals and generalize appropriately.

**Implications for agents:**
- Include the reasoning behind rules and constraints
- Explain the consequences of not following instructions
- Describe the downstream impact of agent outputs

### 3. Use Structured Prompts

Claude was trained on structured prompts and parses them well.

**Recommended structure:**
- Use XML tags (`<section>`, `<constraints>`, `<examples>`) or Markdown headers
- Organize into: Background → Instructions → Constraints → Output Spec
- Front-load critical information

### 4. Be Vigilant with Examples

Claude 4.x pays close attention to details and examples. Poorly chosen examples create unwanted behaviors.

**Best practices:**
- Curate diverse, canonical examples (quality over quantity)
- Avoid exhaustive edge case lists—select representative cases
- Ensure examples match desired behaviors exactly

## Agent Structure Template

Based on analysis of well-optimized agents (10x-dev-pack):

```markdown
---
YAML frontmatter: name, role, description, tools, model, color
---

# Agent Name

Core purpose paragraph (2-3 sentences): What this agent does and why it matters.

## Core Responsibilities
- 3-7 bullets of specific responsibilities
- Active voice, action-oriented
- Start with verbs

## Position in Workflow
- ASCII diagram showing upstream/downstream
- Clear handoff relationships

## Domain Authority

**You decide:**
- Specific decisions within agent's authority
- Concrete, not abstract

**You escalate to [upstream]:**
- Situations requiring human/orchestrator judgment

**You route to [downstream]:**
- Completed work and its destination

## Approach (numbered steps)
1. Step with specific actions
2. Include what to do AND what to check
3. End states should be clear

## What You Produce
- Artifact types with brief descriptions
- Reference to templates where applicable

## Handoff Criteria
- [ ] Checklist of completion conditions
- [ ] Measurable, verifiable items
- [ ] All artifacts verified

## The Acid Test
*Single question that tests if the work meets quality bar*

If uncertain: Default behavior guidance.

## Anti-Patterns (optional)
- 3-5 specific behaviors to avoid
- Concrete examples of failure modes

## Skills Reference
- Links to related skills
```

## Quality Checklist

### Role Clarity
- [ ] Core purpose stated in first 2-3 sentences
- [ ] Role is distinct from other agents
- [ ] Responsibilities don't overlap with adjacent agents
- [ ] Clear boundaries on what this agent does/doesn't do

### Instruction Precision
- [ ] Active voice throughout ("Produce X" not "X should be produced")
- [ ] Specific actions, not vague guidance
- [ ] Imperative instructions where action is expected
- [ ] Numbers and quantities specified where relevant

### Constraint Completeness
- [ ] "You decide" vs "You escalate" boundaries clear
- [ ] Failure modes addressed
- [ ] Edge cases handled or explicitly deferred
- [ ] Time/scope limits specified if applicable

### Example Quality
- [ ] Examples are canonical (representative of common cases)
- [ ] Examples show both input AND output format
- [ ] Bad examples marked as anti-patterns (if included)
- [ ] Examples match the exact format expected

### Structure Adherence
- [ ] Uses standard sections (Core Responsibilities, Domain Authority, Handoff Criteria)
- [ ] Workflow diagram present
- [ ] Skills references at bottom
- [ ] YAML frontmatter complete

### Token Efficiency
- [ ] No redundant text
- [ ] No filler phrases ("It is important to...", "Please ensure...")
- [ ] No duplicate information across sections
- [ ] Under 300 lines total

## Anti-Patterns to Avoid

### Vague Generalities
**Bad:** "Handle documentation appropriately"
**Good:** "Produce audit reports using `@doc-reviews#documentation-audit-report` template"

### Passive Voice
**Bad:** "Documentation should be reviewed"
**Good:** "Review documentation for technical accuracy against codebase"

### Missing Boundaries
**Bad:** "Work with other agents as needed"
**Good:** "Route to Doc Reviewer when accuracy validation required"

### Exhaustive Edge Cases
**Bad:** Long lists of every possible scenario
**Good:** Representative examples + "If uncertain: [default behavior]"

### Implicit Assumptions
**Bad:** Assuming reader knows context
**Good:** Explicit statements of context, audience, and constraints

### Bloated Tool Descriptions
**Bad:** Explaining what each tool does
**Good:** List tools in frontmatter, use them in workflow without explanation

## Optimization Techniques

### 1. Front-Load Critical Information
Put the most important constraints and behaviors at the top. Claude 4.x is more responsive to early content.

### 2. Use Active Voice
"Analyze the audit findings" not "The audit findings should be analyzed"

### 3. Explain the Why
"Verify artifacts via Read tool—hallucinated file paths cause downstream failures"

### 4. Be Prescriptive About Format
If you want specific output structure, show it explicitly.

### 5. Add Constraints as Constraints
Put behavioral limits in a `## Constraints` or `## Domain Authority / You escalate` section—not scattered throughout.

### 6. Include Acid Test
A single question that validates quality. Forces concise quality definition.

### 7. Remove Redundancy
If something is said once clearly, don't repeat it in different words.

## Sources

- [Claude 4.x Best Practices](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-4-best-practices)
- [Claude Code Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices)
- [Effective Context Engineering for AI Agents](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents)
