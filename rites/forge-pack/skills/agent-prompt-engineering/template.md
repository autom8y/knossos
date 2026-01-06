# Agent Prompt Template

> 11-section structure with token budgets. Target 150-200 lines total.

Copy this template. Agents exceeding 250 lines likely contain redundancy.

---

## Section 1: YAML Frontmatter (15-25 lines)

```yaml
---
name: {agent-name}
description: |
  {One-line role description ending with period}.
  Invoke when {trigger-1}, {trigger-2}, or {trigger-3}.
  Produces {artifact-types}.

  When to use this agent:
  - {Detailed use case 1}
  - {Detailed use case 2}
  - {Detailed use case 3}

  <example>
  Context: {Situation description}
  user: "{Example user request}"
  assistant: "{How agent responds or what it produces}"
  </example>
tools: Bash, Glob, Grep, Read, Edit, Write, Task, TodoWrite
model: {opus|sonnet|haiku}  # opus=coordination/architecture, sonnet=implementation, haiku=boilerplate
color: {purple|pink|cyan|green|red|orange|blue}
---
```

<!-- Avoid vague triggers: "Use for documentation" matches too broadly. "Use when creating API reference docs from code comments" is specific. -->

### Model Selection

| Agent Type | Model | Rationale |
|------------|-------|-----------|
| Orchestration/Architecture | opus | Complex coordination, multi-phase planning, trade-off analysis |
| Implementation/Documentation | sonnet | Balanced speed/quality for coding and prose |
| Validation/Boilerplate | haiku | Fast iteration, pattern matching, simple edits |

**Default**: Start with sonnet. Upgrade to opus if agent shows reasoning limitations. Downgrade to haiku if speed becomes bottleneck.

### XML Tags vs Markdown

Use XML for structured examples with multi-turn conversations or embedded code. Markdown headers suffice for typical sections.

---

## Section 2: Title and Overview (3-5 lines)

```markdown
# {Agent Title}

{Opening paragraph: core purpose, approach, and philosophy in 2-3 sentences.
Establishes identity. Answers: Who is this agent? What problem does it solve?}
```

<!-- Avoid generic openings like "This agent helps with X". Use specific phrasing: "Turns ambiguity into specification" conveys purpose and approach. -->

---

## Section 3: Core Responsibilities (6-10 lines)

```markdown
## Core Responsibilities

- **{Verb} {Object}**: {Description with success criterion}
- **{Verb} {Object}**: {Description with success criterion}
- **{Verb} {Object}**: {Description with success criterion}
- **{Verb} {Object}**: {Description with success criterion}
```

<!-- Use action verbs (Analyze, Produce, Validate, Transform). "Code quality" is not actionable; "Identify code smells using static analysis" is. -->

---

## Section 4: Position in Workflow (8-12 lines)

```markdown
## Position in Workflow

```
+-------------------+      +-------------------+      +-------------------+
|  {Upstream Agent} |----->|   {THIS AGENT}    |----->| {Downstream Agent}|
+-------------------+      +-------------------+      +-------------------+
                                    |
                                    v
                            {artifact-type}
```

**Upstream**: {What this agent receives and from whom}
**Downstream**: {What this agent produces and for whom}
```

<!-- Every agent receives from someone and produces for someone. Even entry-point agents receive from "user". -->

---

## Section 5: Domain Authority (12-18 lines)

```markdown
## Domain Authority

**You decide:**
- {Decision fully within your expertise}
- {Decision fully within your expertise}
- {Decision fully within your expertise}

**You escalate to {Role/User}:**
- {Condition requiring escalation}
- {Condition requiring escalation}

**You route to {Next Agent}:**
- {Handoff trigger condition}
- {Handoff trigger condition}
```

<!-- Missing "You escalate" leads to silent failures. Agent assumes ownership of decisions that should go to user or orchestrator. -->

---

## Section 6: How You Work (20-30 lines)

```markdown
## How You Work

### Phase 1: {Phase Name}
{Description of what happens}
1. {Specific step}
2. {Specific step}
3. {Specific step}

### Phase 2: {Phase Name}
{Description of what happens}
1. {Specific step}
2. {Specific step}
3. {Specific step}

### Phase 3: {Phase Name}
...
```

<!-- "Analyze the situation" is not helpful. "Read all source files matching *.ts in src/, identify unused exports" is reproducible. -->

---

## Section 7: What You Produce (15-25 lines)

```markdown
## What You Produce

| Artifact | Description |
|----------|-------------|
| **{Primary Artifact}** | {When/why produced} |
| **{Secondary Artifact}** | {When/why produced} |

### {Primary Artifact} Template

```markdown
# {Title}

## {Section 1}
{Content guidance}

## {Section 2}
{Content guidance}
```
```

<!-- Missing templates cause inconsistent outputs. Downstream agents cannot parse reliably without structure definition. -->

---

## Section 8: Handoff Criteria (8-12 lines)

```markdown
## Handoff Criteria

Ready for {Next Phase/Agent} when:
- [ ] {Criterion 1 - specific, verifiable}
- [ ] {Criterion 2 - specific, verifiable}
- [ ] {Criterion 3 - specific, verifiable}
- [ ] {Criterion 4 - specific, verifiable}
- [ ] {Criterion 5 - specific, verifiable}
```

<!-- "When quality is good" is not testable. "All functions have docstrings with param/return types" is testable. -->

---

## Section 9: The Acid Test (3-5 lines)

```markdown
## The Acid Test

*"{Single yes/no question that determines completion. Forces deep reflection.}"*

If uncertain: {What to do when the answer is not clearly yes}
```

<!-- Pick ONE question. If you need three questions, the agent's scope is too broad. -->

---

## Section 10: Skills Reference (4-6 lines)

```markdown
## Skills Reference

- `documentation` for artifact templates (PRD, TDD, ADR)
- `standards` for code conventions
- `file-verification` for artifact verification protocol
- `{domain-skill}` for {specific guidance}
```

<!-- If three agents share the same 20-line protocol, extract it to a skill. One line replaces 25+. -->

---

## Section 11: Anti-Patterns (8-12 lines)

```markdown
## Anti-Patterns

- **{Pattern Name}**: {Why problematic and what to do instead}
- **{Pattern Name}**: {Why problematic and what to do instead}
- **{Pattern Name}**: {Why problematic and what to do instead}
```

<!-- "Avoid errors" helps no one. "Do not generate code during requirements phase; route design questions to Architect" is actionable. -->

---

## Pre-Deployment Verification

For pre-deployment verification, see [validation/checklist.md](validation/checklist.md).
