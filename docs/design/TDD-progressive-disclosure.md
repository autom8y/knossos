# TDD: Progressive Disclosure Pattern for Monolithic Skills

> Technical Design Document for splitting monolithic skill files into progressive disclosure structures.

---

## Overview

This document defines the canonical pattern for splitting large, monolithic skill files (500+ lines) into a progressive disclosure architecture. The pattern prioritizes quick reference at the entry point with detailed specifications accessible via links.

**Reference Implementation**: `/user-skills/session-lifecycle/start-ref/` demonstrates the target pattern.

---

## Problem Statement

### Current State

Operations skills have grown into monolithic files that mix:
- Quick reference (usage, parameters)
- Full behavioral specification (step-by-step sequences)
- Examples with sample outputs
- Notes, edge cases, troubleshooting

**File sizes (lines)**:
| Skill | Current Lines | Target Entry Point |
|-------|---------------|-------------------|
| spike-ref | 768 | ~100 |
| hotfix-ref | 648 | ~100 |
| commit-ref | 730 | ~100 |
| pr-ref | 680 | ~100 |

### Target State

Each skill becomes a directory with progressive disclosure:
```
skill-name/
  SKILL.md           # Entry point: ~80-120 lines
  behavior.md        # Full specification: 100-200 lines
  examples.md        # Usage scenarios: 100-300 lines
  [topic].md         # Optional additional references
```

---

## Canonical File Structure

### SKILL.md (Entry Point)

**Purpose**: Quick reference for immediate use. Agent reads this first and often needs nothing more.

**Line Budget**: 80-120 lines

**Required Sections** (in order):

```markdown
---
name: skill-name
description: "One-sentence skill description with triggers."
---

# /command - Short Title

> One-line purpose statement.

## Decision Tree

[When to use this vs alternatives - ASCII diagram]

## Usage

[Command syntax with parameters table]

## Quick Reference

[Pre-flight, Actions list, Creates/Produces]

## Anti-Patterns

[Table: Do NOT | Why | Instead]

## Prerequisites

[Bullet list of requirements]

## Success Criteria

[Bullet list of outcomes]

## Related Commands

[Table: Command | When to Use]

## Progressive Disclosure

[Links to behavior.md, examples.md, and other references]
```

**Content Rules**:
1. NO step-by-step behavioral sequences (belongs in behavior.md)
2. NO example outputs with multiple lines (belongs in examples.md)
3. NO troubleshooting guides (belongs in separate file)
4. NO implementation details (belongs in behavior.md)
5. Tables and ASCII diagrams preferred over prose
6. Every section fits on one screen (~25 lines max)

### behavior.md (Full Specification)

**Purpose**: Complete step-by-step behavioral specification for edge cases and debugging.

**Line Budget**: 100-200 lines

**Required Sections**:

```markdown
# /command Behavior Specification

> Full step-by-step sequence for [action].

## Behavior Sequence

### 1. Phase Name

[Detailed steps with code blocks, error handling]

Apply [Pattern Name](../shared-sections/pattern.md):
- Requirement: {requirement}
- Verb: "{verb}"

### 2. Phase Name

[Continue sequence...]

---

## State Changes

### Files Created
[Table: File | Location | Condition]

### Files Modified
[Table: File | Modification]

---

## Error Cases

[Table: Error | Condition | Resolution]

---

## Design Notes (optional)

[Rationale for non-obvious decisions]
```

**Content Rules**:
1. Reference shared patterns via links (do not duplicate)
2. Include code blocks for commands, templates, schemas
3. Document all error conditions and recovery paths
4. Explain state transitions explicitly
5. Cross-reference SKILL.md for quick reference

### examples.md (Usage Scenarios)

**Purpose**: Concrete examples with sample inputs and outputs.

**Line Budget**: 100-300 lines (varies by complexity)

**Structure**:

```markdown
# /command Examples

> Usage scenarios with sample outputs.

## Example 1: [Scenario Name]

**Command**:
\`\`\`bash
/command "argument" --flag=value
\`\`\`

**Output**:
\`\`\`
[Full sample output]
\`\`\`

**Notes**: [Optional explanation]

## Example 2: [Scenario Name]

[Continue pattern...]

---

## Edge Case Examples

### Edge Case: [Name]

**Scenario**: [Description]

**Behavior**: [What happens]

**Example**:
\`\`\`
[Sample]
\`\`\`
```

**Content Rules**:
1. Each example is standalone and complete
2. Show realistic, non-trivial scenarios
3. Include edge cases that clarify behavior
4. Group related examples together
5. Annotate with notes when behavior is non-obvious

### Shared Sections Directory

**Purpose**: Extract patterns duplicated across multiple skills.

**Location**: `{category}/shared-sections/`

**Structure**:
```
shared-sections/
  INDEX.md           # List of partials with purpose
  pattern-1.md       # Reusable pattern
  pattern-2.md       # Reusable pattern
```

**Pattern File Structure**:

```markdown
# Pattern Name

> One-line purpose.

## When to Apply

[List of skills/commands that use this pattern]

## Validation/Implementation

[Steps or checks to perform]

## Error Messages

[Table: Condition | Message Template]

## Customization Points

[Table: Parameter | Description | Commands Using]

## Cross-Reference

[Links to related schemas or patterns]
```

---

## Content Allocation Rules

### What Goes Where

| Content Type | Location | Rationale |
|--------------|----------|-----------|
| Command syntax | SKILL.md | Immediate reference |
| Parameter table | SKILL.md | Immediate reference |
| Decision tree | SKILL.md | When-to-use guidance |
| Anti-patterns table | SKILL.md | Quick guardrails |
| Prerequisites list | SKILL.md | Pre-flight checks |
| Success criteria | SKILL.md | Definition of done |
| Related commands | SKILL.md | Navigation |
| Step-by-step sequence | behavior.md | Detailed specification |
| State changes | behavior.md | Implementation detail |
| Error handling | behavior.md | Edge case coverage |
| Code templates | behavior.md | Implementation reference |
| Full example outputs | examples.md | Concrete demonstration |
| Edge case examples | examples.md | Clarify boundaries |
| Scenario walkthroughs | examples.md | Learning by example |
| Duplicated patterns | shared-sections/ | DRY principle |
| Validation rules | shared-sections/ | Reusable logic |
| Error message templates | shared-sections/ | Consistency |

### Extraction Triggers

Extract to shared-sections when:
1. Pattern appears in 2+ behavior.md files
2. Error messages should be consistent across commands
3. Validation logic is identical across commands
4. Schema reference is needed by multiple skills

### Cross-Reference Conventions

**From SKILL.md to details**:
```markdown
## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [examples.md](examples.md) - Usage scenarios
- [topic.md](topic.md) - Specific topic reference
```

**From behavior.md to shared patterns**:
```markdown
Apply [Pattern Name](../shared-sections/pattern.md):
- Requirement: {specific requirement}
- Verb: "{command verb}"
```

**From any file to session-common**:
```markdown
See [session-context-schema](../../session-common/session-context-schema.md).
```

---

## Anti-Patterns

### During Split

| Anti-Pattern | Problem | Instead |
|--------------|---------|---------|
| Duplicating content | Maintenance burden | Extract to shared-sections or link |
| Leaving SKILL.md too long | Defeats purpose | Enforce line budget |
| Moving everything to behavior.md | Inverts disclosure | Keep quick reference in SKILL.md |
| Breaking links | Skill tool fails | Verify all cross-references |
| Changing behavior | Scope creep | Pure structural refactor only |
| Orphaning examples | Incomplete disclosure | Always link from SKILL.md |

### File Organization

| Anti-Pattern | Problem | Instead |
|--------------|---------|---------|
| Flat directory with many files | Hard to navigate | Use subdirectories for topics |
| Deep nesting | Hard to find files | Maximum 2 levels from SKILL.md |
| Inconsistent naming | Confusion | Follow conventions exactly |
| Missing INDEX.md in shared-sections | Undiscoverable patterns | Always maintain index |

---

## Migration Checklist

For each monolithic skill file:

### Phase 1: Analysis
- [ ] Count total lines
- [ ] Identify distinct sections
- [ ] Mark content for SKILL.md (quick reference)
- [ ] Mark content for behavior.md (step-by-step)
- [ ] Mark content for examples.md (sample outputs)
- [ ] Identify candidates for shared-sections

### Phase 2: Create Structure
- [ ] Create skill directory: `{skill-name}/`
- [ ] Create SKILL.md with frontmatter
- [ ] Create behavior.md with header
- [ ] Create examples.md with header

### Phase 3: Populate Files
- [ ] Move quick reference to SKILL.md
- [ ] Move behavior sequence to behavior.md
- [ ] Move examples to examples.md
- [ ] Add Progressive Disclosure section to SKILL.md
- [ ] Add cross-references between files

### Phase 4: Extract Shared Patterns
- [ ] Identify patterns duplicated across skills
- [ ] Create shared-sections/ if needed
- [ ] Create INDEX.md for shared-sections
- [ ] Extract patterns with proper structure
- [ ] Update behavior.md files to reference patterns

### Phase 5: Verify
- [ ] SKILL.md is within line budget (80-120)
- [ ] All links resolve correctly
- [ ] No content lost during migration
- [ ] No behavioral changes introduced
- [ ] Skill tool can load SKILL.md

---

## Integration Tests

### Skill Loading
```bash
# Verify skill tool finds entry point
grep -l "^name: spike-ref" user-skills/operations/spike-ref/SKILL.md
```

### Link Resolution
```bash
# Verify all markdown links resolve
for file in user-skills/operations/spike-ref/*.md; do
  grep -oE '\[.*\]\(([^)]+)\)' "$file" | while read link; do
    target=$(echo "$link" | sed 's/.*(\([^)]*\)).*/\1/')
    if [[ ! -f "$(dirname $file)/$target" ]]; then
      echo "BROKEN: $file -> $target"
    fi
  done
done
```

### Content Completeness
```bash
# Verify no orphaned content
# Compare line counts before/after
wc -l user-skills/operations/spike-ref/{SKILL,behavior,examples}.md
```

---

## Backward Compatibility

This refactor is **COMPATIBLE**:
- Skill tool continues to load SKILL.md as entry point
- No changes to skill behavior
- No changes to command syntax
- No changes to outputs
- Links are relative (portable)

**No migration path needed**: This is a pure structural reorganization.

---

## File-Level Changes

### For spike-ref Split

**Current**: `/user-skills/operations/spike-ref/SKILL.md` (768 lines)

**Target Structure**:
```
/user-skills/operations/spike-ref/
  SKILL.md      # ~100 lines (entry point)
  behavior.md   # ~150 lines (step-by-step sequence)
  examples.md   # ~200 lines (4 examples)
  templates.md  # ~100 lines (spike report template, agent prompts)
```

### For operations/shared-sections/

**New Directory**: `/user-skills/operations/shared-sections/`

**Files to Create**:
```
/user-skills/operations/shared-sections/
  INDEX.md              # Partial index
  time-boxing.md        # Time budget enforcement pattern
  agent-invocation.md   # Task tool delegation templates
  workflow-diagrams.md  # Mermaid diagram conventions
```

---

## Appendix: Reference Implementation Analysis

### session-lifecycle/start-ref Structure

| File | Lines | Content |
|------|-------|---------|
| SKILL.md | 96 | Frontmatter, usage, quick reference, anti-patterns, progressive disclosure |
| behavior.md | 147 | 8 behavior phases, state changes, error cases |
| examples.md | ~100 | 3 usage scenarios (estimated) |
| integration.md | ~50 | Agent delegation templates (estimated) |

### Key Patterns Observed

1. **SKILL.md Structure**: Follows exact template above
2. **Progressive Disclosure Section**: Always last section in SKILL.md
3. **Shared Pattern References**: Use `Apply [Pattern](path):` format
4. **Error Tables**: Consistent 3-column format (Error | Condition | Resolution)
5. **State Changes Section**: Separate from behavior sequence
6. **Design Notes**: Optional, explains non-obvious decisions

---

## Appendix: spike-ref Split Design

### Current Content Map (768 lines)

| Lines | Section | Target File | Notes |
|-------|---------|-------------|-------|
| 1-4 | Frontmatter | SKILL.md | Keep as-is |
| 6-16 | Title, Purpose | SKILL.md | Condense to 1-line purpose |
| 18-31 | Usage, Parameters | SKILL.md | Keep as-is |
| 34-156 | Behavior (5 phases) | behavior.md | Full sequence |
| 157-223 | Spike Report Template | templates.md | Agent prompt templates |
| 225-252 | Completion Summary | behavior.md | Part of behavior |
| 254-273 | Workflow Diagram | behavior.md | Mermaid diagram |
| 276-296 | Deliverables | SKILL.md (condensed) | Quick list only |
| 299-493 | Examples (4 full) | examples.md | All examples |
| 496-522 | When to Use | SKILL.md | Decision tree |
| 524-543 | Complexity Level | SKILL.md | Quick reference |
| 546-576 | State Changes | behavior.md | Files created/modified |
| 578-586 | Prerequisites | SKILL.md | Keep as-is |
| 588-598 | Error Cases | behavior.md | Error table |
| 599-612 | Related Commands/Skills | SKILL.md | Keep as-is |
| 614-733 | Notes (8 subsections) | notes.md | Design notes |
| 735-768 | Spike Templates | templates.md | Quick-start templates |

### Target File Structure

```
/user-skills/operations/spike-ref/
  SKILL.md           # 105 lines
  behavior.md        # 175 lines
  examples.md        # 200 lines
  templates.md       # 140 lines
  notes.md           # 130 lines
```

**Total**: ~750 lines (content preserved, slight reduction from consolidation)

### SKILL.md Content Outline (~105 lines)

```markdown
---
name: spike-ref
description: "Time-boxed research..."
---

# /spike - Time-Boxed Research

> Execute time-boxed research WITHOUT producing production code.

## Decision Tree

Starting research?
├─ Feasibility question → /spike
├─ Known implementation → /task
├─ Multiple approaches to evaluate → /spike --deliverable=comparison
├─ Quick library check → /spike --timebox=30m
└─ Complex research (8h+) → Break into phases

## Usage

[Command + parameters table - 15 lines]

## Quick Reference

**Pre-flight**: Clear question, defined success criteria, realistic timebox

**Actions**:
1. Define question and success criteria
2. Set timebox (default: 2h)
3. Invoke appropriate agent (Architect or Engineer)
4. Research with progress checkpoints
5. Generate spike report

**Produces**:
- `/docs/research/SPIKE-{slug}.md` (always)
- `/tmp/spike-{slug}/` POC code (optional, throwaway)

**Never Produces**:
- Production code
- PRD/TDD
- Production tests

## Anti-Patterns

| Do NOT | Why | Instead |
|--------|-----|---------|
| Spike without timebox | Becomes endless research | Set explicit limit |
| Ship spike code | Quality relaxed | Create /task for production |
| Spike known solutions | Wastes time | Use /task directly |
| Over-scope research | Loses focus | One question per spike |

## Prerequisites

- Clear research question
- Defined success criteria
- Realistic timebox (30m-8h)

## Success Criteria

- Question answered (or documented as unanswerable)
- Findings in spike report
- Time budget respected
- Next steps clear

## Related Commands

| Command | When to Use |
|---------|-------------|
| /task | After spike approves approach |
| /start | Begin full session (not for spikes) |
| /sprint | Multiple related spikes |

## Progressive Disclosure

- [behavior.md](behavior.md) - Full step-by-step sequence
- [examples.md](examples.md) - 4 usage scenarios
- [templates.md](templates.md) - Agent prompts, report template
- [notes.md](notes.md) - Time-boxing philosophy, POC guidelines
```

### behavior.md Content Outline (~175 lines)

```markdown
# /spike Behavior Specification

> Full step-by-step sequence for time-boxed research.

## Behavior Sequence

### 1. Spike Planning (lines 36-62 from current)

[Prompt for question, success criteria, timebox, deliverable type]
[Spike context YAML schema]

### 2. Research Execution (lines 64-143 from current)

Apply [Agent Invocation Pattern](../shared-sections/agent-invocation.md):
- Agent: Architect (design) or Principal Engineer (feasibility)
- Mode: "SPIKE MODE (Time-boxed research)"

[Three agent invocation templates - condensed references to templates.md]

### 3. Time-Boxing Enforcement (lines 144-156 from current)

Apply [Time-Boxing Pattern](../shared-sections/time-boxing.md):
- Checkpoints: 25%, 50%, 75%, 100%
- Hard stop at limit
- Incomplete is acceptable

### 4. Spike Report Generation (reference templates.md)

### 5. Completion Summary (lines 225-252 from current)

---

## Workflow Diagram (lines 254-273 from current)

[Mermaid diagram]

---

## State Changes

### Files Created (lines 548-556 from current)

### Files NOT Created (lines 557-565 from current)

---

## Error Cases (lines 588-598 from current)

| Error | Condition | Resolution |
|-------|-----------|------------|
| Question too vague | Can't define success | Clarify with user |
| Timebox too short | Can't complete | Document partial, recommend follow-up |
| Timebox too long | > 8h | Break into phases |
| Scope creep | Becoming implementation | Stop, remind research only |
```

### examples.md Content Outline (~200 lines)

```markdown
# /spike Examples

> Usage scenarios with sample outputs.

## Example 1: Architecture Spike (lines 301-346)
## Example 2: Feasibility Spike (lines 347-398)
## Example 3: Technology Comparison (lines 399-448)
## Example 4: Quick 30-Minute Spike (lines 449-493)
```

### templates.md Content Outline (~140 lines)

```markdown
# /spike Templates

> Agent prompts and report templates.

## Agent Invocation Templates

### Architecture/Design Questions (lines 70-92)
### Implementation Feasibility (lines 96-114)
### Technology Comparison (lines 118-142)

## Spike Report Template (lines 163-223)

## Quick-Start Templates (lines 735-768)

### Technology Selection Spike
### Performance Spike
### Integration Spike
### Risk Assessment Spike
```

### notes.md Content Outline (~130 lines)

```markdown
# /spike Design Notes

> Philosophy and guidelines for effective spikes.

## Time-Boxing Philosophy (lines 616-628)
## POC Code Guidelines (lines 629-642)
## Spike to Task Handoff (lines 643-657)
## Multi-Phase Spikes (lines 658-683)
## Spike Report Retention (lines 684-697)
## Collaborative Spikes (lines 698-714)
## Quality vs Speed (lines 718-733)
```

---

## Appendix: Shared Patterns Analysis

### Patterns Common to Operations Skills

Analyzing spike-ref, hotfix-ref, commit-ref, pr-ref reveals shared patterns:

| Pattern | spike-ref | hotfix-ref | commit-ref | pr-ref | Extract? |
|---------|-----------|------------|------------|--------|----------|
| Time budget enforcement | Yes (checkpoints) | Yes (30-60 min) | No | No | Yes |
| Agent invocation template | Yes (3 agents) | Yes (2 agents) | No | No | Yes |
| Deliverable checklist | Yes | Yes | No | No | Partial |
| Git state validation | No | No | Yes | Yes | Yes |
| When to Use table | Yes | Yes | Yes | Yes | No (skill-specific) |
| Error Cases table | Yes | Yes | Yes | Yes | No (skill-specific) |
| Related Commands table | Yes | Yes | Yes | Yes | No (skill-specific) |

### Recommended shared-sections/

```
/user-skills/operations/shared-sections/
  INDEX.md
  time-boxing.md        # Used by: spike-ref, hotfix-ref
  agent-invocation.md   # Used by: spike-ref, hotfix-ref
  git-validation.md     # Used by: commit-ref, pr-ref
```

### time-boxing.md Content

```markdown
# Time-Boxing Pattern

> Enforce time limits with progress checkpoints.

## When to Apply

- /spike: Research time limits (30m-8h)
- /hotfix: Fix time limits (30-90 min by severity)

## Implementation

### Checkpoint Schedule

| Checkpoint | Action |
|------------|--------|
| 25% | Report initial findings |
| 50% | Preliminary conclusions |
| 75% | Start wrapping up |
| 100% | STOP and document |

### Severity-Based Limits (hotfix only)

| Severity | Target | Max |
|----------|--------|-----|
| CRITICAL | 30 min | 60 min |
| HIGH | 45 min | 90 min |
| MEDIUM | 30 min | 60 min |

### Time Exceeded Handling

If exceeding limit:
1. Stop current work
2. Document partial findings
3. For spikes: Recommend follow-up spike
4. For hotfixes: Escalate to full /task

## Customization Points

| Parameter | Description | Commands |
|-----------|-------------|----------|
| timebox | Total duration | spike, hotfix |
| checkpoints | Percentage markers | spike |
| severity | Hotfix priority | hotfix |
```

### agent-invocation.md Content

```markdown
# Agent Invocation Pattern

> Delegate to specialized agents via Task tool.

## When to Apply

- /spike: Research delegation
- /hotfix: Diagnose/fix delegation

## Template Structure

\`\`\`markdown
Act as **{Agent Name}**.

{MODE} ({context})
{Key parameter}: {value}
Time budget: {duration}

{Instructions}:
1. Step 1
2. Step 2
...

Deliverable: {artifact type}
Save to: {path}
\`\`\`

## Agent Selection

| Research Type | Agent | Mode |
|---------------|-------|------|
| Architecture/Design | Architect | SPIKE MODE |
| Feasibility | Principal Engineer | SPIKE MODE |
| Technology comparison | Architect | SPIKE MODE |
| Hotfix diagnose/fix | Principal Engineer | HOTFIX MODE |
| Hotfix validation | QA Adversary | HOTFIX VALIDATION |

## Customization Points

| Parameter | Description | Commands |
|-----------|-------------|----------|
| Agent Name | Target specialist | spike, hotfix |
| MODE | Context label | spike, hotfix |
| Time budget | Time limit | spike, hotfix |
| Deliverable | Output type | spike, hotfix |
| Save path | Artifact location | spike, hotfix |
```

### git-validation.md Content

```markdown
# Git State Validation Pattern

> Validate git repository state before operations.

## When to Apply

- /commit: Pre-commit checks
- /pr: Pre-PR checks

## Validation Checks

| Check | Command | Pass | Fail |
|-------|---------|------|------|
| In repo | `git rev-parse --git-dir` | Exists | Error: Not a repo |
| Not main | `git rev-parse --abbrev-ref HEAD` | != main/master | Error: Switch branch |
| No conflict | `git status` | No "Unmerged" | Error: Resolve conflicts |
| Clean state | `git status --porcelain` | Empty or staged | Prompt or error |
| Has remote | `git remote -v` | Origin exists | Error: Add remote |

## Error Messages

| Condition | Message |
|-----------|---------|
| Not a repo | "Not a git repository. Run `git init` or navigate to repo." |
| On main | "Cannot {verb} from main branch. Switch to feature branch." |
| Merge conflict | "Unresolved merge conflict. Resolve conflicts first." |
| No remote | "No remote configured. Run `git remote add origin URL`." |

## Customization Points

| Parameter | Description | Commands |
|-----------|-------------|----------|
| verb | Action description | commit, pr |
| require_clean | Strict cleanliness | pr (yes), commit (no) |
| require_remote | Remote must exist | pr (yes), commit (no) |
```

---

## Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-01 | Context Architect | Initial design |
| 1.1 | 2026-01-01 | Context Architect | Added spike-ref split design, shared patterns analysis |
