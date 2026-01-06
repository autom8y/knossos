# Prompt Engineering Best Practices Synthesis

> Distilled from Anthropic's official documentation, Claude 4.x best practices, and context engineering research.

## Core Principles

### 1. Be Explicit and Specific

Claude 4.x models respond to clear, explicit instructions. Vague prompts yield vague results.

**Less effective**: "Create an analytics dashboard"
**More effective**: "Create an analytics dashboard. Include user metrics, engagement rates, and retention charts. Add filters for date range and user segment."

### 2. Context is Precious

Treat context as a finite resource. The smallest possible set of high-signal tokens maximizes desired outcomes. As token count increases, models experience degraded recall—"context rot."

- Start minimal, add instructions only for identified failure modes
- Use structured formats (YAML, JSON) for state data
- Use unstructured text for progress notes
- Front-load important information

### 3. Find the Goldilocks Zone

Balance between two failure modes:
- **Too rigid**: Hardcoded complex logic creates fragility and maintenance burden
- **Too vague**: High-level guidance lacks concrete behavioral signals

Aim for: Strong behavioral heuristics without over-specification.

### 4. Provide Motivation

Explain WHY instructions matter. Claude generalizes from explanations.

**Less effective**: "NEVER use ellipses"
**More effective**: "Your response will be read aloud by text-to-speech, so never use ellipses since the engine won't know how to pronounce them."

### 5. Active Voice, Imperative Mood

Tell Claude what to DO, not what NOT to do.

**Less effective**: "Do not use markdown in your response"
**More effective**: "Write your response in smoothly flowing prose paragraphs."

## Agent Prompt Structure Template

```markdown
---
name: agent-name
role: "Concise role statement (1 phrase)"
description: "2-3 sentence description for routing. Include triggers and use cases."
tools: [List, Of, Tools]
model: claude-model-id
color: theme-color
---

# Agent Name

Opening paragraph: WHO this agent is and WHAT unique value it provides. Write in active voice. Front-load the most important information. 2-3 sentences maximum.

## Core Responsibilities

- **Responsibility 1**: Active verb + specific outcome
- **Responsibility 2**: Active verb + specific outcome
- **Responsibility 3**: Active verb + specific outcome
(3-7 responsibilities, no more)

## Position in Workflow

```
[ASCII diagram showing upstream/downstream relationships]
```

**Upstream**: What inputs this agent receives and from whom
**Downstream**: What outputs this agent produces and for whom

## Domain Authority

**You decide:**
- Specific decision 1 (within your authority)
- Specific decision 2

**You escalate:**
- When to escalate and to whom
- Specific escalation triggers

**You route to [Agent Name]:**
- Specific handoff conditions

## When Invoked (First Actions)

1. First thing to do upon invocation
2. Second priority action
3. Third priority action

## Approach

1. **Phase 1**: What to do first
2. **Phase 2**: What to do next
3. **Phase 3**: Continue sequence

## What You Produce

| Artifact | Description |
|----------|-------------|
| **Artifact 1** | What it is and its purpose |
| **Artifact 2** | What it is and its purpose |

### Artifact Production

Reference templates: `@template-skill#template-name`

**Context customization**:
- Specific customization guidance
- Quality requirements

## Handoff Criteria

Ready for [Next Phase] when:
- [ ] Criterion 1 is complete
- [ ] Criterion 2 is complete
- [ ] All artifacts verified via Read tool
- [ ] Attestation table included with absolute paths

## The Acid Test

*"Single question that validates quality of work?"*

If uncertain: Guidance on what to do.

## Anti-Patterns (DO NOT)

- **Anti-pattern 1**: Why it's problematic
- **Anti-pattern 2**: Why it's problematic
- **Anti-pattern 3**: Why it's problematic
(3-5 anti-patterns)
```

## Quality Checklist

### Role Clarity (1-5)
- [ ] Single, unambiguous purpose stated in first paragraph
- [ ] Role distinct from other agents in the team
- [ ] "WHO this is" is immediately clear
- [ ] No overlapping responsibilities with other agents

### Instruction Precision (1-5)
- [ ] Uses active voice and imperative mood
- [ ] Specific, concrete actions (not vague guidance)
- [ ] Numbered sequences for multi-step procedures
- [ ] No ambiguous phrases like "as appropriate" or "when needed"

### Constraint Completeness (1-5)
- [ ] Domain authority is explicit (You decide / You escalate)
- [ ] Handoff criteria are checklists, not prose
- [ ] Anti-patterns prevent common failure modes
- [ ] Tool access is explicitly stated

### Example Quality (1-5)
- [ ] Examples show good vs. bad approaches
- [ ] Concrete, not abstract
- [ ] Relevant to the agent's domain
- [ ] Demonstrate correct format/structure

### Structure Adherence (1-5)
- [ ] YAML frontmatter complete and accurate
- [ ] All required sections present
- [ ] Consistent heading hierarchy
- [ ] Tables used for structured data

### Token Efficiency (1-5)
- [ ] No redundant phrasing
- [ ] Compact but complete
- [ ] Under 300 lines total
- [ ] No filler content

## Anti-Patterns to Avoid

### Prompt Anti-Patterns

1. **Vague Role Definitions**: "Helps with analysis" vs "Synthesizes experiment results into actionable recommendations with confidence levels"

2. **Conflicting Instructions**: "Be concise" but also "Include comprehensive details"

3. **Missing Error Handling**: No guidance on what to do when things go wrong

4. **Implicit Assumptions**: Assuming the agent knows context it doesn't have

5. **Over-reliance on Examples**: Lists of every possible case instead of canonical examples

6. **Static Context Overload**: Pre-loading all data instead of just-in-time retrieval

7. **Passive Voice**: "Experiments should be analyzed" vs "Analyze experiments"

### Agent Design Anti-Patterns

1. **Overlapping Responsibilities**: Multiple agents that could handle the same task

2. **Missing Handoff Criteria**: "When ready" instead of specific checklists

3. **Vague Escalation Triggers**: "When appropriate" instead of specific conditions

4. **Tool Access Ambiguity**: Not stating what tools the agent has/doesn't have

5. **Bloated Tool Sets**: Giving agents tools they don't need

## Claude 4.x Specific Guidance

### Instruction Following
Claude 4.x models follow instructions more precisely. Be explicit about desired behaviors—the model won't add "above and beyond" behaviors unless asked.

### Parallel Tool Calls
Claude 4.x aggressively parallelizes tool calls. If dependencies exist, explicitly state them. If parallelization is desired, encourage it explicitly.

### Output Formatting
- Tell Claude what to do instead of what not to do
- Use XML format indicators when needed
- Match prompt style to desired output style

### Thinking Keywords
When extended thinking is disabled, avoid "think" and variants. Use "consider," "evaluate," "assess" instead.

### Proactive Action
By default, Claude 4.x may suggest rather than act. For proactive behavior:
```
By default, implement changes rather than only suggesting them. If the user's intent is unclear, infer the most useful likely action and proceed.
```

### State Management
- Use structured formats (JSON) for state data
- Use git for tracking state across sessions
- Emphasize incremental progress
- Include checkpoints for long-running work

## Sources

- [Anthropic Context Engineering](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents)
- [Claude 4 Best Practices](https://docs.claude.com/en/docs/build-with-claude/prompt-engineering/claude-4-best-practices)
- [Claude Code Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices)
