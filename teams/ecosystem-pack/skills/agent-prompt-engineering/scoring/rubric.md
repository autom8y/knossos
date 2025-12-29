# Agent Quality Rubric

> 6-dimension scoring system for evaluating agent prompts

## Scoring Matrix Template

| Agent | Role Clarity | Instruction Precision | Constraint Completeness | Example Quality | Structure Adherence | Token Efficiency | **Avg** |
|-------|-------------|----------------------|------------------------|-----------------|--------------------|--------------------|---------|
| {agent-1} | /5 | /5 | /5 | /5 | /5 | /5 | /5 |
| {agent-2} | /5 | /5 | /5 | /5 | /5 | /5 | /5 |

**Interpretation**: Agents scoring below 4.0 average require optimization. Prioritize lowest-scoring dimension.

---

## Dimension 1: Role Clarity (1-5)

Does the agent's identity come through clearly in the first two sentences?

| Score | Criteria |
|-------|----------|
| **5** | Role unmistakable. Reader knows exactly what agent does and why it exists. |
| **4** | Role clear. Minor ambiguity about scope boundaries. |
| **3** | Role recognizable. Some confusion about what distinguishes from similar agents. |
| **2** | Role vague. Generic description that could apply to multiple agents. |
| **1** | Role unclear. Cannot determine agent's purpose from opening. |

**Test**: Show first paragraph to someone unfamiliar with the system. Can they explain what this agent does in one sentence?

---

## Dimension 2: Instruction Precision (1-5)

Are responsibilities and tasks unambiguous and actionable?

| Score | Criteria |
|-------|----------|
| **5** | All instructions use action verbs with specific targets and success criteria. |
| **4** | Most instructions actionable. 1-2 vague items that could be interpreted multiple ways. |
| **3** | Mix of precise and vague. Several responsibilities lack clear success criteria. |
| **2** | Mostly vague. Instructions describe activities rather than outcomes. |
| **1** | Entirely vague. "Helps with X" language throughout. |

**Test**: For each responsibility, ask "How would I verify this was done correctly?" If no clear answer, score lower.

---

## Dimension 3: Constraint Completeness (1-5)

Are boundaries well-defined? Does agent know what it owns vs. escalates?

| Score | Criteria |
|-------|----------|
| **5** | Complete Domain Authority: decide/escalate/route all explicit with examples. |
| **4** | Good boundaries. Missing 1-2 edge cases for escalation or routing. |
| **3** | Partial boundaries. Decide section present but escalate/route incomplete. |
| **2** | Minimal boundaries. Only mentions what agent does, not what it avoids. |
| **1** | No boundaries. Agent scope undefined. Could attempt anything. |

**Test**: Present an edge case scenario. Can you determine from the prompt whether agent should handle it, escalate, or route?

---

## Dimension 4: Example Quality (1-5)

Are examples concrete, representative, and helpful for understanding agent behavior?

| Score | Criteria |
|-------|----------|
| **5** | Examples show real input/output. Cover typical case and edge case. |
| **4** | Good examples. Representative but missing edge cases or error handling. |
| **3** | Basic examples. Show happy path but don't illuminate nuance. |
| **2** | Generic examples. Could apply to many agents. Don't show this agent's specifics. |
| **1** | No examples or placeholder examples only. |

**Test**: Could someone replicate the agent's expected behavior from the examples alone?

---

## Dimension 5: Structure Adherence (1-5)

Does the prompt follow the 11-section template structure?

| Score | Criteria |
|-------|----------|
| **5** | All 11 sections present in correct order. Each section fulfills its purpose. |
| **4** | 10-11 sections present. Minor ordering deviation or 1 underdeveloped section. |
| **3** | 8-9 sections present. Some sections combined or missing. |
| **2** | 5-7 sections. Custom structure that omits key sections. |
| **1** | Freeform structure. Does not follow template. |

**Required sections**: Frontmatter, Title/Overview, Responsibilities, Workflow Position, Domain Authority, How You Work, Outputs, Handoff Criteria, Acid Test, Skills Reference, Anti-Patterns.

---

## Dimension 6: Token Efficiency (1-5)

Is the prompt concise without losing clarity?

| Score | Criteria |
|-------|----------|
| **5** | Under 180 lines. No redundancy. Active voice throughout. Skills referenced, not embedded. |
| **4** | 180-220 lines. Minimal redundancy. Mostly active voice. |
| **3** | 220-260 lines. Some redundancy. Passive constructions present. |
| **2** | 260-300 lines. Significant redundancy. Content that could be shared skills. |
| **1** | Over 300 lines. Extensive redundancy. Repeated content across sections. |

**Test**: Can you remove 20 lines without losing information? If yes, score lower.

**Common efficiency issues**:
- File verification protocol repeated (25+ lines) instead of skill reference (1 line)
- Explanations of concepts Claude already understands
- Passive voice requiring more words than active

---

## Scoring Process

### Step 1: Initial Read
Read the prompt once without scoring. Note first impressions.

### Step 2: Dimension Scoring
Score each dimension independently. Reference criteria tables above.

### Step 3: Evidence Collection
For scores below 4, document specific issues:
- Line numbers with problems
- Quotes of vague/passive language
- Missing sections or content

### Step 4: Priority Calculation
Calculate average. Identify lowest-scoring dimension for priority optimization.

### Step 5: Recommendations
For each dimension below 4, provide specific rewrite guidance.

---

## Example Scored Audit

```markdown
## Orchestrator Audit

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Role Clarity | 4/5 | Clear role but "stateless advisor" concept buried in middle |
| Instruction Precision | 3/5 | Consultation protocol complex; could simplify |
| Constraint Completeness | 4/5 | Good DO/DO NOT but some implicit assumptions |
| Example Quality | 3/5 | YAML structure shown but no complete request/response example |
| Structure Adherence | 5/5 | All sections present and ordered correctly |
| Token Efficiency | 3/5 | 291 lines with 25-line file verification repeated |
| **Average** | **3.7/5** | **Priority: Token Efficiency, Instruction Precision** |

### Recommendations
1. Front-load "stateless advisor" in first paragraph
2. Add complete CONSULTATION_REQUEST/RESPONSE example
3. Replace file verification with skill reference: -25 lines
4. Simplify consultation protocol: -20 lines
5. Target: 180 lines (current: 291)
```

---

## Interpretation Guide

| Average Score | Assessment | Action |
|---------------|------------|--------|
| 4.5-5.0 | Excellent | Minor polish only |
| 4.0-4.4 | Good | Address lowest dimension |
| 3.5-3.9 | Adequate | Prioritize 2 lowest dimensions |
| 3.0-3.4 | Needs work | Significant rewrite required |
| Below 3.0 | Poor | Consider complete rewrite |
