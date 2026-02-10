# Intelligence-Pack Agent Audit

> Scored against canonical prompt engineering best practices. Scale: 1-5 (5 = exemplary).

## Summary Score Matrix

| Agent | Role Clarity | Instruction Precision | Constraint Completeness | Example Quality | Structure Adherence | Token Efficiency | **Total** |
|-------|-------------|----------------------|------------------------|----------------|--------------------|-----------------:|-------:|
| pythia | 4 | 4 | 5 | 4 | 5 | 3 | **25** |
| analytics-engineer | 4 | 4 | 4 | 2 | 4 | 4 | **22** |
| user-researcher | 3 | 3 | 4 | 2 | 4 | 3 | **19** |
| experimentation-lead | 3 | 3 | 4 | 2 | 4 | 4 | **20** |
| insights-analyst | 3 | 3 | 4 | 2 | 4 | 3 | **19** |

## Priority Order (lowest score first)

1. **user-researcher** (19/30) - Needs most improvement
2. **insights-analyst** (19/30) - Tied for lowest
3. **experimentation-lead** (20/30) - Close third
4. **analytics-engineer** (22/30) - Moderate improvements needed
5. **pythia** (25/30) - Best baseline, minor polish

---

## Detailed Audit: pythia.md

**Lines**: 293 | **Current Quality**: Good baseline

### Role Clarity: 4/5
- Clear "consultative throughline" purpose
- Good distinction between what it does vs. doesn't do
- **Issue**: Opening paragraph is slightly abstract ("consultative throughline" needs unpacking)

### Instruction Precision: 4/5
- Good CONSULTATION_REQUEST/RESPONSE format
- Clear behavioral constraints with DO/INSTEAD pattern
- **Issue**: Some prose sections could be more imperative

### Constraint Completeness: 5/5
- Excellent domain authority section
- Clear escalation triggers
- Explicit tool access limitations
- Thorough handoff criteria

### Example Quality: 4/5
- Good YAML schema examples
- **Issue**: Missing concrete example of a complete consultation exchange

### Structure Adherence: 5/5
- All required sections present
- Consistent heading hierarchy
- Good use of tables and code blocks

### Token Efficiency: 3/5
- **Issue**: 293 lines is near limit
- Some redundancy between "The Litmus Test" and "The Acid Test"
- Position in Workflow diagram takes significant tokens
- Consultation Protocol is comprehensive but verbose

### Specific Issues
1. Line 12-13: "consultative throughline" is jargon-heavy
2. Lines 55-137: Consultation protocol is thorough but could be more compact
3. Lines 270-293: Anti-patterns section is good but duplicates some behavioral constraints
4. Missing "When Invoked (First Actions)" section

---

## Detailed Audit: analytics-engineer.md

**Lines**: 141 | **Current Quality**: Solid

### Role Clarity: 4/5
- Clear "data foundation" purpose
- Good opening persona ("I build the data foundation")
- **Issue**: Could be more specific about what differentiates from generic data engineer

### Instruction Precision: 4/5
- Good approach sequence (Understand, Design, Validate, Guide)
- **Issue**: Some vague language ("as appropriate", "relevant to")

### Constraint Completeness: 4/5
- Domain authority present but brief
- **Issue**: Missing specific escalation scenarios
- File verification section references skill but doesn't summarize

### Example Quality: 2/5
- **Major Issue**: No concrete examples of event naming, tracking plan format, or validation rules
- Anti-patterns list is abstract

### Structure Adherence: 4/5
- All required sections present
- **Issue**: Missing "When Invoked (First Actions)" section
- Position in Workflow diagram is minimal

### Token Efficiency: 4/5
- Good length at 141 lines
- Compact but complete

### Specific Issues
1. Lines 10-12: Opening paragraph uses first person ("I build...") inconsistently with other content
2. Lines 55-58: "Understand" phase is vague ("map user journeys to instrument")
3. Lines 68-77: Artifact production section references external template but gives minimal customization
4. No concrete examples of event taxonomy or tracking plan structure
5. Missing "When Invoked (First Actions)" section

---

## Detailed Audit: user-researcher.md

**Lines**: 178 | **Current Quality**: Needs work

### Role Clarity: 3/5
- Purpose is clear but opening is too casual/conversational
- "I talk to humans" doesn't match professional tone of other agents
- **Issue**: Blends persona with instruction in confusing way

### Instruction Precision: 3/5
- Approach section is reasonable but vague
- **Issue**: "Design: Clarify research questions, select methodology" lacks specificity
- Session checkpoints section is procedural but copied from template

### Constraint Completeness: 4/5
- Domain authority present
- Handoff criteria checklist is good
- File verification section present

### Example Quality: 2/5
- **Major Issue**: No examples of interview questions, synthesis approach, or research findings format
- Anti-patterns are useful but abstract

### Structure Adherence: 4/5
- All sections present
- **Issue**: Session Checkpoints section feels copy-pasted (same as principal-engineer)
- Missing "When Invoked (First Actions)" section

### Token Efficiency: 3/5
- Session checkpoints section adds 40+ lines that could reference skill
- File Operation Discipline section is nearly identical to other agents (25+ lines)

### Specific Issues
1. Lines 10-12: Casual tone ("I talk to humans") doesn't match professional agent style
2. Lines 55-58: Approach is high-level without concrete guidance
3. Lines 79-143: File verification + session checkpoints = 64 lines of copy-pasted boilerplate
4. No example interview guide, research template, or synthesis approach
5. Missing "When Invoked (First Actions)" section

---

## Detailed Audit: experimentation-lead.md

**Lines**: 142 | **Current Quality**: Moderate

### Role Clarity: 3/5
- Clear experimentation focus
- Opening is conversational but less casual than user-researcher
- **Issue**: "I run the scientific method on product" is catchy but vague

### Instruction Precision: 3/5
- Approach section is reasonable
- **Issue**: Sample size calculation guidance is generic ("80-90% power")
- No specific statistical test guidance

### Constraint Completeness: 4/5
- Domain authority present
- Handoff criteria checklist is good
- File verification section present

### Example Quality: 2/5
- **Major Issue**: No example experiment design, sample size calculation, or pre-registration
- Anti-patterns are useful but could have examples

### Structure Adherence: 4/5
- All sections present
- **Issue**: Missing "When Invoked (First Actions)" section
- Could benefit from example statistical formulas

### Token Efficiency: 4/5
- Reasonable length at 142 lines
- Less boilerplate than user-researcher

### Specific Issues
1. Lines 10-12: "I run the scientific method on product" is vague
2. Lines 55-59: Approach lacks concrete statistical guidance
3. Lines 69-78: Artifact production references template but sample size guidance is generic
4. No example of MDE calculation, power analysis, or experiment design document
5. Missing "When Invoked (First Actions)" section

---

## Detailed Audit: insights-analyst.md

**Lines**: 159 | **Current Quality**: Needs work

### Role Clarity: 3/5
- Purpose is clear (data → decisions)
- Opening is conversational
- **Issue**: "I turn data into decisions" is generic—could apply to many roles

### Instruction Precision: 3/5
- Approach section is reasonable
- **Issue**: "Validate significance, analyze segments" lacks statistical specificity
- Communication guidance is vague

### Constraint Completeness: 4/5
- Domain authority present
- Handoff criteria checklist is good
- **Issue**: "Route to" section is minimal compared to other agents

### Example Quality: 2/5
- **Major Issue**: No example insights report, executive summary, or data narrative
- Anti-patterns are useful but abstract

### Structure Adherence: 4/5
- All sections present
- Session Boundaries section is different from checkpoints (good customization)
- **Issue**: Missing "When Invoked (First Actions)" section

### Token Efficiency: 3/5
- Reasonable length at 159 lines
- Could be more compact

### Specific Issues
1. Lines 10-12: "I turn data into decisions" is too generic
2. Lines 55-58: Approach lacks concrete guidance on segment analysis, confidence intervals
3. Lines 69-79: Artifact production is good but examples would help
4. No example of impact/confidence rating, alternative explanation analysis
5. Missing "When Invoked (First Actions)" section

---

## Cross-Cutting Issues (Apply to All)

### 1. Missing "When Invoked (First Actions)" Section
All agents lack explicit numbered first actions. Add:
```markdown
## When Invoked (First Actions)

1. Read upstream artifacts completely (PRD, TDD, previous phase output)
2. Verify context is sufficient to proceed
3. Confirm session directory and artifact paths
4. Begin primary phase work
```

### 2. Boilerplate Duplication
File Operation Discipline and Session Checkpoints sections are nearly identical across agents. Solution: Compress to single reference:
```markdown
## File Verification

See `file-verification` skill for verification protocol (absolute paths, Read confirmation, attestation tables, session checkpoints).
```

### 3. Inconsistent Opening Tone
Some use casual first-person ("I talk to humans"), others are more professional. Standardize to brief, professional tone:
- Keep personality but front-load professional purpose
- Move colorful description to second sentence

### 4. Missing Concrete Examples
Most agents reference templates but provide no concrete examples. Add at least one example per artifact type.

### 5. Generic Anti-Patterns
Anti-patterns are correct but abstract. Make them actionable with specific examples of what to do instead.

---

## Rewrite Priority Order

Based on audit scores and improvement potential:

1. **user-researcher** (19/30) - Most issues, most boilerplate, casualist tone
2. **insights-analyst** (19/30) - Generic role definition, missing examples
3. **experimentation-lead** (20/30) - Needs statistical specificity
4. **analytics-engineer** (22/30) - Good structure, needs examples
5. **pythia** (25/30) - Polish only, compress token usage
