# SRE-Pack Agents Audit

> Audit against canonical prompt engineering best practices
> Date: 2024-12-28

---

## Score Matrix

| Agent | Role Clarity | Instruction Precision | Constraint Completeness | Example Quality | Structure Adherence | Token Efficiency | Total | Priority |
|-------|-------------|----------------------|------------------------|-----------------|--------------------|--------------------|-------|----------|
| orchestrator | 4 | 4 | 5 | 4 | 4 | 3 | 24/30 | 5 |
| incident-commander | 4 | 4 | 4 | 4 | 4 | 3 | 23/30 | 4 |
| observability-engineer | 4 | 4 | 4 | 4 | 4 | 3 | 23/30 | 3 |
| chaos-engineer | 4 | 4 | 4 | 4 | 4 | 2 | 22/30 | 2 |
| platform-engineer | 4 | 4 | 4 | 3 | 4 | 2 | 21/30 | 1 |

**Priority Order** (lowest score first): platform-engineer, chaos-engineer, observability-engineer, incident-commander, orchestrator

---

## Detailed Agent Assessments

### 1. platform-engineer.md (Priority 1 - Rewrite First)

**Current Line Count**: 340 lines (exceeds 300 target)

| Criterion | Score | Issues |
|-----------|-------|--------|
| Role Clarity | 4 | Good "roads developers drive on" metaphor, clear triggers |
| Instruction Precision | 4 | Active voice, clear phases |
| Constraint Completeness | 4 | Domain authority defined, handoff criteria present |
| Example Quality | 3 | Templates are verbose, could be more canonical |
| Structure Adherence | 4 | All sections present, consistent format |
| Token Efficiency | 2 | **Major issue**: Templates bloat (Infrastructure Change, Pipeline Design) consume ~100 lines. Session Checkpoints section duplicates file-verification skill. |

**Specific Issues**:
1. Templates (lines 87-175) are overly detailed—should reference skill instead
2. Session Checkpoints (lines 207-240) duplicates content from file-verification skill
3. Deployment Strategies table and IaC Best Practices could be condensed
4. Developer Experience Checklist is generic—not SRE-specific

**Recommendations**:
- Remove inline templates, reference `@doc-sre#infrastructure-change-template` and `@doc-sre#pipeline-design-template`
- Remove Session Checkpoints section, keep reference to skill
- Condense patterns section to key heuristics
- Focus on platform engineering for **reliability** (SRE context)

---

### 2. chaos-engineer.md (Priority 2)

**Current Line Count**: 246 lines (within target)

| Criterion | Score | Issues |
|-----------|-------|--------|
| Role Clarity | 4 | "Breaks production on purpose" is memorable |
| Instruction Precision | 4 | Hypothesize/Design/Execute/Analyze/Report is clear |
| Constraint Completeness | 4 | Domain authority well-defined |
| Example Quality | 4 | Common Experiments and Gameday Protocol are useful |
| Structure Adherence | 4 | All sections present |
| Token Efficiency | 2 | **Issue**: Failure Types table (lines 158-166) and Common Experiments (lines 168-188) are reference material, not behavioral guidance |

**Specific Issues**:
1. Failure Types table is reference data, not agent instruction
2. Common Experiments could be more concise
3. Gameday Protocol is detailed but rarely invoked
4. Safety Principles (lines 213-223) overlap with anti-patterns

**Recommendations**:
- Move Failure Types to a reference skill or doc
- Condense Common Experiments to 3 key examples
- Merge Safety Principles into Approach or Anti-Patterns
- Strengthen "You escalate to Incident Commander" with specific triggers

---

### 3. observability-engineer.md (Priority 3)

**Current Line Count**: 211 lines (within target)

| Criterion | Score | Issues |
|-----------|-------|--------|
| Role Clarity | 4 | "Makes the invisible visible" is clear |
| Instruction Precision | 4 | Inventory/Analyze/Design/Define/Recommend is logical |
| Constraint Completeness | 4 | Domain authority defined, escalation paths clear |
| Example Quality | 4 | Three Pillars explanation is good |
| Structure Adherence | 4 | All sections present |
| Token Efficiency | 3 | Some redundancy in pillars explanation and SLI categories |

**Specific Issues**:
1. Three Pillars section (lines 151-172) is educational, not behavioral
2. SLI Categories table duplicates knowledge Claude has
3. Alert Anti-Patterns table could be in anti-patterns section
4. File Operation Discipline duplicates file-verification skill

**Recommendations**:
- Condense Three Pillars to 1-2 lines referencing Claude's knowledge
- Move Alert Anti-Patterns table to Anti-Patterns section
- Remove File Operation Discipline, reference skill
- Strengthen SLO definition guidance (when to propose SLOs vs. accept existing)

---

### 4. incident-commander.md (Priority 4)

**Current Line Count**: 216 lines (within target)

| Criterion | Score | Issues |
|-----------|-------|--------|
| Role Clarity | 4 | "Runs the war room" is clear |
| Instruction Precision | 4 | Declare/Coordinate/Resolve/Facilitate/Plan is actionable |
| Constraint Completeness | 4 | Decision authority well-defined |
| Example Quality | 4 | War Room Protocol and Escalation Triggers are useful |
| Structure Adherence | 4 | All sections present |
| Token Efficiency | 3 | Communication Templates section is verbose |

**Specific Issues**:
1. Communication Templates (lines 186-194) are inline—should reference skill
2. War Room Protocol could be more concise
3. Escalation Triggers table is good but could have clearer thresholds
4. File Operation Discipline duplicates file-verification skill

**Recommendations**:
- Reference `@doc-sre#incident-communication-template` instead of inline
- Remove File Operation Discipline, keep skill reference
- Add explicit "When to NOT declare incident" guidance
- Strengthen blameless postmortem emphasis in Acid Test

---

### 5. orchestrator.md (Priority 5 - Least Critical)

**Current Line Count**: 295 lines (within target)

| Criterion | Score | Issues |
|-----------|-------|--------|
| Role Clarity | 4 | "Consultative throughline" is clear |
| Instruction Precision | 4 | CONSULTATION_REQUEST/RESPONSE format is precise |
| Constraint Completeness | 5 | Strongest domain authority definition of all agents |
| Example Quality | 4 | YAML examples are helpful |
| Structure Adherence | 4 | All sections present |
| Token Efficiency | 3 | Some redundancy in behavioral constraints |

**Specific Issues**:
1. "What You DO NOT DO" section (lines 27-33) overlaps with Behavioral Constraints (lines 209-227)
2. CONSULTATION_RESPONSE example could be condensed
3. Tool Access section repeats constraint information
4. Handling Failures section is somewhat redundant with throughline guidance

**Recommendations**:
- Consolidate "DO NOT" constraints into single section
- Reduce YAML example verbosity (show structure, not full content)
- Remove redundant Tool Access elaboration
- Strengthen "Handling Failures" with specific recovery patterns

---

## Common Issues Across All Agents

1. **File Operation Discipline duplication**: All 5 agents have identical 20-line sections. Should reference skill once.

2. **Session Checkpoints duplication**: 2 agents have identical checkpoint sections. Should reference skill.

3. **Skills Reference sections**: All agents have similar "Reference these skills" sections. Could be standardized.

4. **Cross-Team Routing**: All agents reference cross-team skill identically. One line sufficient.

5. **Template bloat**: Platform Engineer and Chaos Engineer have inline templates that should be skills.

---

## Rewrite Priority Order

1. **platform-engineer.md** - Highest priority (340 lines, most bloat, template redundancy)
2. **chaos-engineer.md** - High priority (reference material bloat)
3. **observability-engineer.md** - Medium priority (educational content redundancy)
4. **incident-commander.md** - Medium priority (inline templates)
5. **orchestrator.md** - Lowest priority (already well-structured, minor redundancy)

---

## Target Outcomes

After rewrite, each agent should:
- Be under 200 lines (stretch goal) or 250 lines (acceptable)
- Have no duplicated content with skills
- Use active voice and imperative mood throughout
- Have explicit handoff criteria as checkboxes
- Reference skills instead of inline templates
- Focus on behavioral guidance, not reference material
