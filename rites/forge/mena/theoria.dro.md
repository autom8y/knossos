---
name: theoria
description: "Structured domain audit producing State of the {target} report with letter grades"
argument-hint: "<domain|all> [--target=NAME] [--domains=DOMAIN1,DOMAIN2]"
allowed-tools: Bash, Read, Glob, Grep, Task, Skill, Write
model: opus
context: fork
---

# /theoria - Domain Audit System

Orchestrates multi-domain audits using parallel theoros agents, produces "State of the X" synthesis reports with cross-domain grading.

## Context

This command operates in forked context (transient session). Session state and active rite tracking are not available during execution.

## Pre-flight

Parse `$ARGUMENTS` to determine audit scope:

1. **Single domain**: `/theoria dromena` — audit one domain
2. **All domains**: `/theoria all` or `/theoria` — audit all registered domains
3. **Target name**: `--target=NAME` — sets report title (default: "Ecosystem")
4. **Domain override**: `--domains=D1,D2` — audit specific domains (comma-separated)

**Validation:**
- If no domains specified and none registered: ERROR "No domains registered. See pinakes for domain registry."
- If specific domain requested, verify it exists in registry before proceeding
- `--domains` takes precedence over positional domain argument

**Argument parsing examples:**
- `/theoria dromena` → domain="dromena", target="Ecosystem"
- `/theoria all --target=Infrastructure` → all domains, target="Infrastructure"
- `/theoria --domains=dromena,agents` → domains=["dromena","agents"], target="Ecosystem"

## Phase 1: Domain Discovery

### Load Domain Registry

1. Invoke the pinakes skill to load domain registry and grading schemas:
   ```
   Skill("pinakes")
   ```

2. Read the domain registry table from pinakes INDEX:
   ```
   Read(".claude/skills/pinakes/INDEX.md")
   ```

3. Parse the Domain Registry table to extract:
   - Domain name
   - Criteria file path (relative to repo root)
   - Scope category (framework, codebase, process, culture)
   - Description

4. If specific domain requested, verify it appears in registry table
   - If not found: ERROR "Domain '{domain}' not found in registry. Available domains: {list}"

5. Build audit list:
   - If `--domains=X,Y`: use X and Y only
   - If `all` or no args: use all domains from registry
   - If specific domain: use that domain only

## Phase 2: Criteria Loading

For each domain in the audit list:

1. Read the domain criteria file from the path extracted in Phase 1
   - Example: `rites/shared/mena/pinakes/domains/dromena.lego.md`

2. Extract from criteria file:
   - **Scope**: Target files glob pattern (e.g., `.claude/commands/**/*.md`)
   - **Criteria definitions**: Each criterion with name, weight, and grading thresholds
   - **Evaluation guidance**: Evidence collection instructions

3. Store criteria content for injection into theoros prompt

## Phase 3: Parallel Dispatch (Argus Pattern)

Dispatch one theoros agent per domain in parallel. Each theoros receives:

- **Domain name**: The domain being audited
- **Domain criteria**: Full text of the criteria file
- **Scope pattern**: The file glob pattern to audit

### Theoros Invocation

For each domain, construct Task prompt:

```
Task(subagent_type="theoros", prompt="
Audit domain: {domain_name}

## Domain Criteria

{full_criteria_file_content}

## Audit Instructions

1. Use Glob to discover all files matching the scope pattern
2. Evaluate each criterion according to the grading thresholds
3. Provide evidence: file paths, line numbers, counts
4. Compute overall domain grade using weighted average
5. Produce structured Domain Assessment following the output schema in your prompt

Use the grading scale from your agent prompt (A-F, no modifiers). Show all percentage calculations.
")
```

**Launch all theoros agents in parallel** (multiple Task calls in one response block).

**Wait for all theoros to complete** before proceeding to synkrisis phase.

## Phase 4: Synkrisis (Inline Synthesis)

After all theoros agents return their Domain Assessments, synthesize cross-domain findings.

### Collect Domain Results

For each theoros output:
1. Extract domain name
2. Extract overall letter grade
3. Extract overall percentage
4. Extract key findings (strengths, weaknesses, recommendations)
5. Store complete Domain Assessment for embedding in final report

### Compute Aggregate Grade

Use the midpoint conversion method from grading schema:

| Letter | Midpoint Percentage |
|--------|---------------------|
| A      | 95%                 |
| B      | 85%                 |
| C      | 75%                 |
| D      | 65%                 |
| F      | 40%                 |

**Process:**
1. Convert each domain letter grade to midpoint percentage
2. Compute simple average across all domains (equal weighting)
3. Map average back to letter grade using thresholds:
   - 90-100% → A
   - 80-89% → B
   - 70-79% → C
   - 60-69% → D
   - Below 60% → F

**Example:**
- dromena: B (85%) → 85%
- legomena: A (95%) → 95%
- agents: C (75%) → 75%
- Average: (85 + 95 + 75) / 3 = 85.0% → **B**

### Identify Cross-Domain Patterns

**Systemic Strengths:**
- Capabilities appearing in 2+ domain assessments with grade A or B
- Example: "Consistent frontmatter schema across all primitives"

**Systemic Weaknesses:**
- Gaps appearing in 2+ domain assessments with grade D or F
- Example: "Missing examples across dromena, legomena, and agents"

**Priority Recommendations:**
- Actions with cross-domain impact
- Ranked by number of domains affected and severity
- Example: "Add argument-hint field to all 8 dromena lacking it (affects 67% of slash commands)"

## Phase 5: Report Generation

Write the synkrisis report to `.wip/STATE-OF-{TARGET}-{YYYY-MM-DD}.md` using the format below.

### Report Format

```markdown
# State of the {target}: {date}

**Overall Grade: {letter}** (across {N} domains)

## Executive Summary

{2-3_sentence_overview_of_overall_ecosystem_state}

{highlight_most_critical_cross_domain_finding_or_pattern}

## Domain Grades

| Domain | Grade | Key Finding |
|--------|-------|-------------|
| {domain_name} | {letter} ({pct}%) | {one_sentence_summary} |
| {domain_name} | {letter} ({pct}%) | {one_sentence_summary} |

## Cross-Domain Findings

### Systemic Strengths

- **{strength_theme}**: {description_with_examples_across_domains}
- **{strength_theme}**: {description_with_examples_across_domains}

### Systemic Weaknesses

- **{weakness_theme}**: {description_with_impact_across_domains}
- **{weakness_theme}**: {description_with_impact_across_domains}

### Priority Recommendations

1. **{recommendation}**: {specific_action_with_cross_domain_scope}
2. **{recommendation}**: {specific_action_with_cross_domain_scope}
3. **{recommendation}**: {specific_action_with_cross_domain_scope}

## Domain Details

{embedded_domain_assessment_1}

---

{embedded_domain_assessment_2}

---

{embedded_domain_assessment_3}

## Methodology

- **Domains evaluated**: {comma_separated_domain_list}
- **Grading scale**: A-F (90-100% = A, 80-89% = B, 70-79% = C, 60-69% = D, <60% = F)
- **Evaluation date**: {YYYY-MM-DD}
- **Aggregation method**: Equal weighting across domains, midpoint conversion
- **Report generated by**: /theoria v1 (Argus Pattern with parallel theoros dispatch)
```

### Report Rules

**Executive Summary:**
- Maximum 3 sentences
- Focus on overall state and highest-impact finding
- Plain language, no jargon

**Cross-Domain Findings:**
- Systemic patterns only (appear in 2+ domains)
- Evidence-based with domain references
- Recommendations must be actionable and specific

**Domain Details:**
- Embed complete Domain Assessment from each theoros
- Use `---` separator between assessments
- Preserve all criterion detail and evidence

**File Naming:**
- Pattern: `STATE-OF-{TARGET}-{YYYY-MM-DD}.md`
- TARGET in uppercase
- Date in ISO format (YYYY-MM-DD)
- Examples: `STATE-OF-ECOSYSTEM-2026-02-10.md`, `STATE-OF-FRAMEWORK-2026-02-10.md`

## Output

After writing the report, display a summary to the user:

```
Theoria Audit Complete

Report: .wip/STATE-OF-{TARGET}-{YYYY-MM-DD}.md
Overall Grade: {letter} (across {N} domains)

Domain Grades:
- {domain}: {letter} ({pct}%)
- {domain}: {letter} ({pct}%)

Key Finding: {most_critical_weakness_or_strength}

Run `cat .wip/STATE-OF-{TARGET}-{YYYY-MM-DD}.md` to view full report.
```

## Examples

### Example 1: Single Domain Audit
```
/theoria dromena
```
Audits slash commands only, produces `STATE-OF-ECOSYSTEM-2026-02-10.md` with one domain.

### Example 2: Full Ecosystem Audit
```
/theoria all
```
Audits all registered domains (dromena, legomena, agents), produces comprehensive ecosystem report.

### Example 3: Custom Target Audit
```
/theoria --domains=dromena,agents --target=Infrastructure
```
Audits dromena and agents domains only, report titled "State of the Infrastructure".

### Example 4: Default Behavior
```
/theoria
```
Same as `/theoria all` — audits all registered domains.

## Error Handling

**No domains registered:**
- Check pinakes INDEX for empty Domain Registry table
- Error message: "No domains registered in pinakes. Cannot run audit."

**Requested domain not found:**
- Verify domain exists in registry table
- Error message: "Domain '{domain}' not found. Available: {list}"

**Theoros dispatch fails:**
- If any theoros returns error, include error in report as domain grade "ERROR"
- Note in synkrisis which domains failed to evaluate

**Criteria file missing:**
- If domain listed in registry but criteria file doesn't exist
- Error message: "Criteria file missing for domain '{domain}' at {path}"

## Integration

**Consumes:**
- `pinakes` legomena (Skill tool) — domain registry and grading schemas
- `rites/shared/mena/pinakes/domains/*.lego.md` — domain criteria files
- `theoros` agent (Task tool) — domain evaluator

**Produces:**
- `.wip/STATE-OF-{TARGET}-{YYYY-MM-DD}.md` — synkrisis synthesis report

**Dependencies:**
- Requires `pinakes` legomena to be materialized in `.claude/skills/`
- Requires `theoros` agent to be materialized in `.claude/agents/`
- Both available after `ari sync` in forge rite

## Design Decisions

**Why inline synkrisis?**
- Synkrisis is orchestration logic, not domain evaluation
- /theoria already has opus model for complex synthesis
- Reduces agent indirection, keeps synthesis code visible

**Why Argus Pattern (parallel dispatch)?**
- Domains are independent, can be audited in parallel
- Reduces total audit time from O(N) to O(1) for N domains
- Named for Argus Panoptes — the all-seeing observer with many eyes

**Why simple A-F grades?**
- No grade inflation via +/- modifiers
- Forces clear threshold decisions
- Easier to aggregate and compare across domains

**Why equal weighting across domains?**
- Default policy prevents gaming the system
- Can be overridden with `--domains` for focused audits
- Simpler aggregation logic

## Related

- `pinakes` legomena — Domain registry and grading schemas
- `theoros.md` agent — Domain evaluator
- `rites/shared/mena/pinakes/schemas/report-format.lego.md` — Report format specification
- `rites/shared/mena/pinakes/schemas/grading.lego.md` — Grading calculation rules
