# Pre-Deployment Validation Checklist

> Verify agent prompts before adding to production rite catalog

## Quick Validation

Use this abbreviated checklist for rapid verification:

- [ ] Frontmatter parses as valid YAML
- [ ] Description contains at least 3 trigger phrases
- [ ] First 2 sentences establish role identity
- [ ] All responsibilities start with action verbs
- [ ] Domain Authority has decide/escalate/route sections
- [ ] Handoff criteria are objectively testable
- [ ] Total lines under 200

---

## Detailed Validation

### Frontmatter Validation

| Check | Criteria | Pass/Fail |
|-------|----------|-----------|
| YAML syntax | Parses without errors | |
| name field | Lowercase, hyphenated, matches filename | |
| description | 3+ lines with triggers and example | |
| tools | Lists only tools agent actually uses | |
| model | Valid model identifier (opus/sonnet/haiku) | |
| color | One of: purple, pink, cyan, green, red, orange, blue | |

**Validation command**: Run YAML linter on frontmatter block.

---

### Section-by-Section Checks

#### Title and Overview
- [ ] Title matches frontmatter name (human-readable form)
- [ ] Opening establishes identity in first 2 sentences
- [ ] No generic language ("helps with", "works on")
- [ ] Clear problem statement (what this agent solves)

#### Core Responsibilities
- [ ] 4-6 responsibilities listed
- [ ] Each starts with action verb (Analyze, Produce, Validate, etc.)
- [ ] Each has measurable outcome
- [ ] No overlap with other agents in same team

#### Position in Workflow
- [ ] ASCII diagram present and renders correctly
- [ ] Upstream agent/source identified
- [ ] Downstream agent/destination identified
- [ ] Artifact type labeled

#### Domain Authority
- [ ] "You decide" section present with 3+ items
- [ ] "You escalate" section present with 2+ conditions
- [ ] "You route to" section present with specific triggers
- [ ] No implicit assumptions about ownership

#### How You Work
- [ ] 3-4 phases defined
- [ ] Each phase has numbered steps
- [ ] Steps are specific enough to reproduce
- [ ] No gaps between phases

#### What You Produce
- [ ] Artifact table present
- [ ] Primary artifact template included
- [ ] Template structure matches other agents' expectations
- [ ] Output format parseable by downstream agents

#### Handoff Criteria
- [ ] 5-7 checklist items
- [ ] All items use objective criteria (not "good quality")
- [ ] Items are independently verifiable
- [ ] Completing all items means work is truly done

#### The Acid Test
- [ ] Single yes/no question
- [ ] Question is specific to this agent's domain
- [ ] "If uncertain" guidance provided

#### Skills Reference
- [ ] References existing skills (not placeholder names)
- [ ] Skills are relevant to this agent's work
- [ ] No embedded content that belongs in skills

#### Anti-Patterns
- [ ] 3-5 anti-patterns listed
- [ ] Each is specific to this agent (not generic advice)
- [ ] Each explains why it's problematic
- [ ] Each provides alternative behavior

---

### Integration Verification

| Check | Criteria | Pass/Fail |
|-------|----------|-----------|
| Upstream compatibility | Agent can receive expected input format | |
| Downstream compatibility | Agent produces expected output format | |
| Skill references valid | All referenced skills exist in skills directory | |
| No duplicate content | Content unique (not copied from other agents) | |
| Cross-references valid | All file paths/links point to existing files | |

---

### Token Efficiency Check

| Check | Criteria | Pass/Fail |
|-------|----------|-----------|
| Line count | Under 200 lines total | |
| No repeated protocols | File verification etc. use skill references | |
| Active voice | Grep for "should be", "might need", "could" | |
| No over-explanation | No "As you know", "It's important to note" | |
| Concept density | Each paragraph adds new information | |

**Note**: Detection methods for these efficiency issues are documented in [principles.md](../principles.md) Detection Checklist.

---

### Final Sign-Off

Complete these checks before adding to rite catalog:

- [ ] Scored 4+ on all rubric dimensions (see [rubric.md](../scoring/rubric.md))
- [ ] Zero anti-patterns (see [principles.md](../principles.md) Detection Checklist)
- [ ] Reviewed by second person for clarity
- [ ] Tested with sample input to verify expected behavior
- [ ] Added to rite catalog manifest
- [ ] CLAUDE.md updated with agent reference

---

## Validation Failure Remediation

### Common Failures and Fixes

| Failure | Typical Cause | Fix |
|---------|---------------|-----|
| YAML parse error | Unquoted special characters | Quote description string |
| Missing trigger phrases | Vague description | Add 3+ specific trigger phrases |
| Subjective handoff criteria | "Quality" language | Replace with measurable condition |
| Over line limit | Embedded content | Extract to skill reference |
| Failed integration | Format mismatch | Align with upstream/downstream templates |

### Re-validation Process

After making fixes:
1. Re-run failed section checks only
2. Verify fix didn't break passing checks
3. Re-score affected rubric dimensions
4. Document changes in PR description
