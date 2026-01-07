# TDD: /consult Meta-Advisor Enhancement

| Field | Value |
|-------|-------|
| **Initiative** | /consult Meta-Advisor Enhancement |
| **PRD** | PRD-consult-meta-advisor.md |
| **Author** | Context Architect |
| **Status** | Draft |
| **Date** | 2026-01-02 |
| **Complexity** | MODULE |

## 1. Overview

This document specifies the technical design for enhancing `/consult` to serve as the "cognitive load absorber" for the roster ecosystem. The design introduces a `team-discovery` skill for dynamic team metadata access, enhances `consult-ref` with explicit skill references, updates `team-ref` to dynamically discover all 11 teams, and adds `/consult` to CLAUDE.md's Getting Help table.

### 1.1 Problem Statement

Current state:
- `consult-ref` provides ecosystem guidance but lacks explicit references to `prompting` skill for invocation patterns
- `consult-ref` does not reference `10x-workflow` for workflow phase transitions and quality gates
- The role of `/consult` as "cognitive load absorber" is undocumented
- `team-ref` hardcodes 4 teams (`10x-dev`, `docs`, `hygiene`, `debt-triage`) when 11 teams exist
- Team profile information is scattered across `rites/*/orchestrator.yaml` and `rites/*/README.md` without structured access
- CLAUDE.md Getting Help table does not include `/consult` for routing guidance

### 1.2 Design Principles

- **Dynamic discovery over static lists**: Team counts come from filesystem, not hardcoded values
- **Skill reference over duplication**: `/consult` references `prompting` and `10x-workflow` rather than duplicating content
- **Single source of truth**: `team-discovery` reads from `rites/*/orchestrator.yaml` as canonical team profiles
- **Cognitive load absorption**: Users describe intent; ecosystem handles routing complexity
- **Stateless operation**: `/consult` advises but does not execute

### 1.3 Key Constraints

| Constraint | Source | Impact |
|------------|--------|--------|
| `/consult` advises only, does not invoke skills | PRD out-of-scope | Response includes patterns but user must invoke |
| Team profiles use existing `orchestrator.yaml` | Existing infrastructure | No new TEAM.yaml files; use existing schema |
| CLAUDE.md changes use table format | PRD-claude-md-descriptive-architecture | Adding to Getting Help table, not prose |
| Backward compatible | Existing skill structure | Enhance existing skills, do not replace |

### 1.4 Scope

**In Scope:**
- New `team-discovery` skill for structured team metadata (FR-2.1, FR-2.4)
- `consult-ref` enhancement with skill references (FR-1.1, FR-1.2, FR-3.2, FR-4.1)
- `team-ref` enhancement for dynamic team discovery (US-8)
- CLAUDE.md Getting Help table update (FR-3.1)

**Out of Scope:**
- Active delegation from `/consult` to skills
- Changes to `orchestrator.yaml` schema
- Session history integration (FR-S.6)
- Interactive wizard mode (FR-C.1)

---

## 2. Component Specifications

### 2.1 team-discovery Skill (NEW)

Create a new skill that provides structured team metadata for routing decisions.

#### 2.1.1 Location

```
.claude/skills/team-discovery/
  SKILL.md              # Entry point with discovery behavior
  schemas/
    rite-profile.yaml   # Canonical team profile schema
```

#### 2.1.2 team-profile Schema

The `team-discovery` skill reads from existing `rites/*/orchestrator.yaml` files and structures the output for routing:

```yaml
# Schema: rite-profile.yaml
# Describes the structure that team-discovery extracts from orchestrator.yaml

team_profile:
  name: string           # Required: rite name (e.g., "10x-dev")
  domain: string         # Required: team specialization (e.g., "software development")
  description: string    # Required: from orchestrator.yaml frontmatter.description
  role: string           # Required: from orchestrator.yaml frontmatter.role
  quick_switch: string   # Derived: "/" + first word of name (e.g., "/10x")
  status: enum           # Derived: "active" | "dormant" based on presence

  routing:               # From orchestrator.yaml routing section
    conditions: map[string]string  # specialist -> condition

  workflow_position:     # From orchestrator.yaml workflow_position
    upstream: string
    downstream: string

  agents: list[string]   # Derived: filenames from rites/{name}/agents/*.md
  agent_count: integer   # Derived: count of agents

  skills: list[string]   # From orchestrator.yaml skills array

  complexity_levels: list[string]  # From workflow.yaml complexities (if available)

  use_cases:             # From orchestrator.yaml routing (rephrased for user intent)
    - pattern: string    # Natural language query pattern
      matches: boolean   # Whether this team handles it

# Validation rules:
# - name must match directory name in rites/
# - quick_switch derived from name pattern (debt-triage -> /debt)
# - agent_count must match actual agent files
# - All required fields must be present
```

#### 2.1.3 SKILL.md Content Structure

```markdown
---
name: team-discovery
description: "Dynamic team metadata discovery from roster. Use when: routing decisions need team profiles, displaying team inventory, matching intent to teams. Triggers: list teams, team capabilities, what teams exist, team metadata."
---

# Team Discovery

> Provides structured team metadata by reading from `rites/*/orchestrator.yaml`.

## Purpose

Enables dynamic team discovery without hardcoding team counts or capabilities. Other skills (`consult-ref`, `team-ref`) reference this skill for current team inventory.

## Usage

This skill provides read-only team metadata. It does not switch teams.

### List All Teams

Read all orchestrator.yaml files from `$ROSTER_HOME/rites/*/orchestrator.yaml` and extract:
- Team name, domain, description
- Quick-switch command
- Agent roster (from rites/{name}/agents/*.md)
- Routing conditions (what triggers each specialist)

### Match Intent to Team

Given a user query, compare against team routing conditions:
1. Parse query for key verbs (build, fix, deploy, document, etc.)
2. Match against team domains and routing conditions
3. Return ranked list with confidence scores

## Data Source

**Primary**: `$ROSTER_HOME/rites/*/orchestrator.yaml`
**Supplementary**: `$ROSTER_HOME/rites/*/README.md` for use cases

### Current Team Inventory

[This section is dynamically generated by reading the filesystem]

To get current team count:
```bash
ls -d $ROSTER_HOME/rites/*-pack | wc -l
```

To list rite names:
```bash
ls -d $ROSTER_HOME/rites/*-pack | xargs -n1 basename
```

## Schema Reference

See [schemas/rite-profile.yaml](schemas/rite-profile.yaml) for the structured output format.

## Integration Points

- **consult-ref**: Calls team-discovery for routing recommendations
- **team-ref**: Calls team-discovery for --list output
- **SessionStart hook**: May call team-discovery for context injection
```

#### 2.1.4 Design Decision: Read orchestrator.yaml vs. New TEAM.yaml

**Option A: Create new TEAM.yaml files (Rejected)**

Would require:
- 11 new TEAM.yaml files
- Potential schema drift between TEAM.yaml and orchestrator.yaml
- Additional maintenance burden

**Option B: Read existing orchestrator.yaml (Selected)**

Rationale:
- `orchestrator.yaml` already contains team metadata (name, domain, description, routing)
- Agent list derivable from `rites/{name}/agents/*.md`
- Single source of truth for team configuration
- No new files to maintain

**Trade-off**: Some fields (like `quick_switch`) must be derived rather than explicit. This is acceptable because the derivation logic is deterministic.

---

### 2.2 consult-ref Skill Enhancement

Modify existing `.claude/skills/consult-ref/SKILL.md` to add:
1. Explicit "cognitive load absorber" positioning
2. References to `prompting` skill for invocation patterns
3. References to `10x-workflow` skill for phase transitions

#### 2.2.1 Section Additions

**New section after "Purpose" (approximately line 20):**

```markdown
## Cognitive Load Absorber Role

`/consult` is the designated entry point for users who are confused, overwhelmed, or unsure where to start. When users don't know which team, workflow, or command to use, `/consult` absorbs that cognitive load by:

1. **Parsing natural language intent** - Understanding what the user wants to accomplish
2. **Matching to ecosystem capabilities** - Identifying the right team, workflow, and complexity
3. **Providing actionable patterns** - Referencing `prompting` skill for exact invocation commands
4. **Explaining the journey** - Referencing `10x-workflow` for phase transitions and quality gates

**Key principle**: Users describe outcomes; `/consult` handles routing complexity.

**Positioning in CLAUDE.md**: "Unsure? Route to `/consult` for guidance."
```

**Enhancement to "Mode 2: Query Provided" section (approximately line 109):**

Add after "Generate Command-Flow" step:

```markdown
3b. **Reference Invocation Patterns**
   - Invoke `prompting` skill mentally to retrieve exact copy-paste patterns
   - Include patterns from `prompting/patterns/*.md` appropriate to workflow type
   - For orchestrator-coordinated workflows, reference `prompting/patterns/discovery.md`
   - For implementation phases, reference `prompting/patterns/implementation.md`

3c. **Reference Workflow Context**
   - Invoke `10x-workflow` skill mentally for phase/gate information
   - Include quality gate expectations from `10x-workflow/quality-gates.md`
   - Reference lifecycle phases from `10x-workflow/lifecycle.md` when explaining journey
```

**Enhancement to "Response Format" section (approximately line 388):**

Add to "Command-Flow" subsection:

```markdown
### 3. Command-Flow
Step-by-step commands to execute:
- Numbered sequence of commands
- Expected artifacts at each step
- Decision points and checkpoints
- Agent handoff patterns
- **Invocation patterns** from `prompting` skill (copy-paste ready)
- **Quality gates** from `10x-workflow` at phase transitions
```

**New section after "Response Format" (approximately line 420):**

```markdown
---

## Skill Reference Patterns

`/consult` references other skills to provide accurate, current information:

### Referencing prompting Skill

When providing invocation patterns, `/consult` retrieves patterns from:
- `prompting/SKILL.md` - Agent invocation quick reference
- `prompting/patterns/discovery.md` - PRD creation, session initialization
- `prompting/patterns/implementation.md` - TDD, coding, testing
- `prompting/patterns/validation.md` - QA, review gates

**Pattern**: Instead of generating ad-hoc commands, extract from `prompting` skill.

Example transformation:
- WITHOUT skill reference: "Start a session with `/start 'Add auth'`"
- WITH skill reference: "Per `prompting` skill, invoke: `Act as Requirements Analyst. Create PRD for: Add auth`"

### Referencing 10x-workflow Skill

When explaining workflow journeys, `/consult` retrieves context from:
- `10x-workflow/SKILL.md` - Agent routing, complexity calibration
- `10x-workflow/lifecycle.md` - Phase protocol (PLAN -> CLARIFY -> EXECUTE -> VERIFY -> HANDOFF)
- `10x-workflow/quality-gates.md` - Gate criteria per phase

**Pattern**: Include quality gate expectations when describing phase transitions.

Example enhancement:
- WITHOUT: "After PRD, handoff to Architect"
- WITH: "After PRD, quality gate requires: problem clear, scope defined, requirements testable. Then handoff to Architect per `10x-workflow` lifecycle."

### Referencing team-discovery Skill

When recommending teams, `/consult` retrieves current team inventory from:
- `team-discovery` skill for dynamic team list
- `team-discovery/schemas/rite-profile.yaml` for profile structure

**Pattern**: Never hardcode team counts or capabilities; always read from `team-discovery`.
```

#### 2.2.2 File Changes Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `.claude/skills/consult-ref/SKILL.md` | Modify | Add cognitive load absorber section (~15 lines) |
| `.claude/skills/consult-ref/SKILL.md` | Modify | Add skill reference patterns section (~60 lines) |
| `.claude/skills/consult-ref/SKILL.md` | Modify | Enhance Mode 2 with skill reference steps (~15 lines) |
| `.claude/skills/consult-ref/SKILL.md` | Modify | Enhance Response Format with skill references (~5 lines) |

---

### 2.3 team-ref Skill Enhancement

Modify existing `.claude/skills/team-ref/skill.md` to dynamically discover teams rather than hardcoding 4.

#### 2.3.1 Current vs. Target State

**Current (lines 84-103):**
```markdown
## Available Team Packs

### 10x-dev
...
### docs
...
### hygiene
...
### debt-triage
...
```

**Target:**
```markdown
## Available Team Packs

Team packs are discovered dynamically from `$ROSTER_HOME/rites/`. Use `team-discovery` skill for structured metadata.

### Discovering Teams

The current team inventory is read at runtime:
```bash
ls -d $ROSTER_HOME/rites/*-pack | xargs -n1 basename
```

### Team Quick Reference

| Team | Command | Domain | Agents |
|------|---------|--------|--------|
| *Dynamically populated from team-discovery* |

For full team profiles, run `/consult --team` or reference `team-discovery` skill.

**Note**: This skill previously listed 4 teams. The ecosystem now has 11 teams. Always use dynamic discovery.
```

#### 2.3.2 Section Replacements

**Replace lines 84-103** (Available Team Packs section) with:

```markdown
## Available Team Packs

Team packs are discovered dynamically from `$ROSTER_HOME/rites/`. Reference the `team-discovery` skill for structured metadata access.

### Current Inventory

To list all teams at runtime:
```bash
ls -d $ROSTER_HOME/rites/*-pack 2>/dev/null | xargs -n1 basename
```

As of this writing, the roster contains 11 teams:
- 10x-dev (software development)
- debt-triage (technical debt)
- docs (documentation)
- ecosystem (CEM/skeleton/roster infrastructure)
- forge (rite creation)
- hygiene (code quality)
- intelligence (analytics/research)
- rnd (exploration/prototyping)
- security (security assessment)
- sre (operations/reliability)
- strategy (business analysis)

**Important**: This list is informational. For current, accurate team data, use `team-discovery` skill or read directly from `$ROSTER_HOME/rites/*/orchestrator.yaml`.

### Team Details

For detailed team profiles including agents, routing conditions, and use cases:
- Run `/consult --team` for formatted display
- Reference `team-discovery` skill for structured data
- Read `rites/{name}/README.md` for extended documentation
```

#### 2.3.3 Quick Switch Commands Update

**Replace lines 336-344** (Quick Switch Commands section) with:

```markdown
## Quick Switch Commands

Quick-switch commands are derived from rite names:

| Team | Quick Switch | Derivation |
|------|--------------|------------|
| 10x-dev | `/10x` | First token before hyphen |
| debt-triage | `/debt` | First token before hyphen |
| docs | `/docs` | First token before hyphen |
| ecosystem | `/ecosystem` | First token before hyphen |
| forge | `/forge` | First token before hyphen |
| hygiene | `/hygiene` | First token before hyphen |
| intelligence | `/intelligence` | First token before hyphen |
| rnd | `/rnd` | First token before hyphen |
| security | `/security` | First token before hyphen |
| sre | `/sre` | First token before hyphen |
| strategy | `/strategy` | First token before hyphen |

These commands invoke `/team {pack-name}` internally and display team roster after switch.
```

---

### 2.4 CLAUDE.md Update

Add `/consult` to the Getting Help table per FR-3.1.

#### 2.4.1 Current State

```markdown
## Getting Help

| Question | Skill |
|----------|-------|
| Invoke agents | `prompting` |
| Templates | `documentation` |
| Conventions | `standards` |
| Coordination | `10x-workflow` |
```

#### 2.4.2 Target State

```markdown
## Getting Help

| Question | Skill |
|----------|-------|
| Invoke agents | `prompting` |
| Templates | `documentation` |
| Conventions | `standards` |
| Coordination | `10x-workflow` |
| Unsure where to start | `/consult` |
```

#### 2.4.3 File Location and Line Change

**File**: `/Users/tomtenuta/Code/roster/.claude/CLAUDE.md`

**Change**: Add one row to Getting Help table after "Coordination" row.

**Exact edit**: After line containing `| Coordination | \`10x-workflow\` |`, insert:
```markdown
| Unsure where to start | `/consult` |
```

---

## 3. Data Flow

### 3.1 Team Discovery Flow

```
User: /consult "I need to build a payment system"
         │
         ▼
    ┌─────────────────┐
    │   consult-ref   │
    │      skill      │
    └────────┬────────┘
             │ 1. Parse intent (build, payment, system)
             ▼
    ┌─────────────────┐
    │ team-discovery  │◄── Reads: $ROSTER_HOME/rites/*/orchestrator.yaml
    │      skill      │
    └────────┬────────┘
             │ 2. Match intent to team routing conditions
             │    Returns: [{team: 10x-dev, confidence: 0.95}, ...]
             ▼
    ┌─────────────────┐
    │    prompting    │◄── Reads: prompting/patterns/discovery.md
    │      skill      │
    └────────┬────────┘
             │ 3. Retrieve invocation pattern for recommended workflow
             │    Returns: "Act as Requirements Analyst. Create PRD for: {feature}"
             ▼
    ┌─────────────────┐
    │   10x-workflow  │◄── Reads: 10x-workflow/quality-gates.md
    │      skill      │
    └────────┬────────┘
             │ 4. Retrieve quality gate criteria for journey explanation
             │    Returns: Phase transitions with gate criteria
             ▼
    ┌─────────────────────────────────────────┐
    │           /consult Response             │
    │  Assessment: Build payment system       │
    │  Recommendation: 10x-dev, /task    │
    │  Command-Flow: /10x && /start...        │
    │  (with prompting patterns)              │
    │  (with 10x-workflow gates)              │
    │  Alternatives: /security if PCI scope   │
    └─────────────────────────────────────────┘
```

### 3.2 Team Listing Flow

```
User: /team --list
         │
         ▼
    ┌─────────────────┐
    │    team-ref     │
    │      skill      │
    └────────┬────────┘
             │ 1. Invoke team-discovery for current inventory
             ▼
    ┌─────────────────┐
    │ team-discovery  │◄── Reads: $ROSTER_HOME/rites/*/orchestrator.yaml
    │      skill      │    Reads: $ROSTER_HOME/rites/*/agents/*.md (for counts)
    └────────┬────────┘
             │ 2. Returns structured team list with:
             │    - name, domain, quick_switch
             │    - agent_count
             │    - status (active/dormant)
             ▼
    ┌─────────────────────────────────────────┐
    │         /team --list Response           │
    │  [Roster] Available teams:              │
    │    - 10x-dev (5 agents) - /10x     │
    │    - debt-triage (3 agents) - /debt│
    │    - ...11 teams total                  │
    └─────────────────────────────────────────┘
```

---

## 4. File Changes Summary

### 4.1 New Files

| File | Purpose | Lines (est.) |
|------|---------|--------------|
| `.claude/skills/team-discovery/SKILL.md` | Team discovery skill entry point | ~80 |
| `.claude/skills/team-discovery/schemas/rite-profile.yaml` | Team profile schema | ~40 |

### 4.2 Modified Files

| File | Change | Lines Changed |
|------|--------|---------------|
| `.claude/skills/consult-ref/SKILL.md` | Add cognitive load absorber section | +15 |
| `.claude/skills/consult-ref/SKILL.md` | Add skill reference patterns section | +60 |
| `.claude/skills/consult-ref/SKILL.md` | Enhance Mode 2 with skill references | +15 |
| `.claude/skills/consult-ref/SKILL.md` | Enhance Response Format | +5 |
| `.claude/skills/team-ref/skill.md` | Replace hardcoded teams with dynamic discovery | ~-20/+40 |
| `.claude/skills/team-ref/skill.md` | Update quick switch commands table | ~-9/+15 |
| `.claude/CLAUDE.md` | Add /consult to Getting Help table | +1 |

---

## 5. Backward Compatibility

### 5.1 Classification: COMPATIBLE

All changes are additive or enhance existing behavior:

| Component | Change Type | Backward Compatible |
|-----------|-------------|---------------------|
| team-discovery | New skill | Yes - additive |
| consult-ref | Enhanced behavior | Yes - adds capabilities |
| team-ref | Documentation update | Yes - reflects reality |
| CLAUDE.md | Table row addition | Yes - additive |

### 5.2 Migration Path

None required. All changes enhance existing functionality without breaking existing invocations.

### 5.3 Deprecation Notes

The following hardcoded references are deprecated in favor of dynamic discovery:

| Deprecated | Replacement |
|------------|-------------|
| "9 teams" in consult-ref SKILL.md line 61 | Dynamic count from team-discovery |
| "All 9 Teams" in consult-ref SKILL.md line 295 | Dynamic list from team-discovery |
| "4 rites" implicit in team-ref | Dynamic discovery showing 11 teams |

---

## 6. Integration Test Matrix

### 6.1 team-discovery Validation

| Test Case | Input | Expected Outcome | Validates |
|-----------|-------|------------------|-----------|
| TD-01: List all teams | `ls $ROSTER_HOME/rites/` | Returns 11 rite names | Team enumeration |
| TD-02: Parse orchestrator.yaml | Read 10x-dev/orchestrator.yaml | Extracts name, domain, routing | Schema parsing |
| TD-03: Count agents | Read 10x-dev/agents/*.md | Returns 4-5 agents | Agent enumeration |
| TD-04: Derive quick_switch | Parse "debt-triage" | Returns "/debt" | Derivation logic |
| TD-05: Handle missing orchestrator.yaml | rites/bad-pack/ (no yaml) | Skips gracefully | Error handling |

### 6.2 consult-ref Validation

| Test Case | Input | Expected Outcome | Validates |
|-----------|-------|------------------|-----------|
| CR-01: Intent to 10x | "I want to build a feature" | Recommends 10x-dev | Intent matching |
| CR-02: Intent to security | "audit my code for vulnerabilities" | Recommends security | Domain matching |
| CR-03: Prompting reference | Any recommendation | Includes prompting pattern | FR-1.1 |
| CR-04: Workflow reference | Any recommendation | Includes quality gates | FR-1.2 |
| CR-05: No arguments | `/consult` | Shows scannable overview | FR-3.3 |
| CR-06: Ambiguous intent | "help me code" | Asks clarifying questions | FR-4.3 |

### 6.3 team-ref Validation

| Test Case | Input | Expected Outcome | Validates |
|-----------|-------|------------------|-----------|
| TR-01: List teams | `/team --list` | Shows 11 teams | US-8 |
| TR-02: Quick switch mapping | All quick switches | Maps correctly to rite names | Derivation |
| TR-03: Dynamic count | Add/remove team | Count updates | Dynamic discovery |

### 6.4 CLAUDE.md Validation

| Test Case | Input | Expected Outcome | Validates |
|-----------|-------|------------------|-----------|
| CM-01: Getting Help table | Read CLAUDE.md | Contains `/consult` row | FR-3.1 |
| CM-02: Table format | Parse Getting Help | Valid markdown table | Format compliance |

---

## 7. Implementation Notes for Integration Engineer

### 7.1 team-discovery Implementation

1. **Read orchestrator.yaml files**: Use glob pattern `rites/*/orchestrator.yaml`
2. **Parse YAML frontmatter**: Extract `rite.name`, `team.domain`, `frontmatter.description`, `frontmatter.role`
3. **Derive quick_switch**: Split rite name on `-`, take first token, prefix with `/`
4. **Count agents**: Glob `rites/{name}/agents/*.md`, count results
5. **Handle errors**: Skip teams missing orchestrator.yaml, log warning

### 7.2 consult-ref Implementation

1. **Add sections at appropriate locations** per line numbers in 2.2.1
2. **Preserve existing content**: Enhancements are additive
3. **Update hardcoded "9 teams"**: Replace with reference to team-discovery
4. **Test with representative queries**: Validate skill references appear in output

### 7.3 team-ref Implementation

1. **Replace Available Team Packs section**: Lines 84-103 with new content
2. **Replace Quick Switch Commands section**: Lines 336-344 with full 11-team table
3. **Verify backward compatibility**: Existing `/team` commands work unchanged

### 7.4 CLAUDE.md Implementation

1. **Locate Getting Help table**: Search for `## Getting Help`
2. **Add row after Coordination**: Single line addition
3. **Verify table renders**: Check markdown preview

---

## 8. Quality Gates

### 8.1 Pre-Implementation Gate

- [ ] PRD requirements understood (FR-1.1, FR-1.2, FR-2.1, FR-2.4, FR-3.1, FR-3.2, FR-4.1)
- [ ] All 11 teams accessible in `$ROSTER_HOME/rites/`
- [ ] Existing `consult-ref`, `team-ref`, and CLAUDE.md files readable

### 8.2 Implementation Gate

- [ ] team-discovery skill created with SKILL.md and schema
- [ ] consult-ref SKILL.md enhanced with all sections
- [ ] team-ref skill.md updated with dynamic discovery
- [ ] CLAUDE.md Getting Help table updated

### 8.3 Validation Gate

- [ ] All test cases in section 6 pass
- [ ] `/consult` responses include prompting skill patterns
- [ ] `/consult` responses include 10x-workflow quality gates
- [ ] `/team --list` shows 11 teams
- [ ] CLAUDE.md Getting Help table includes `/consult`

---

## 9. Open Design Decisions

### 9.1 Resolved

| Decision | Resolution | Rationale |
|----------|------------|-----------|
| How does consult-ref reference prompting? | Mental invocation + pattern extraction | Matches skill design; avoids recursive invocation |
| What's the schema for team-discovery output? | rite-profile.yaml (section 2.1.2) | Structured for routing decisions |
| Skill or script for team-discovery? | Skill | Aligns with ecosystem architecture; reusable |
| How to keep team counts synchronized? | Dynamic filesystem read | Single source of truth; no maintenance burden |

### 9.2 Implementation Deferred

| Decision | Recommendation | Notes |
|----------|----------------|-------|
| Confidence scoring algorithm | Simple keyword matching | Can be enhanced in future iteration |
| Session context awareness (FR-S.1) | Defer to hooks | SessionStart hook already provides context |
| Usage logging (FR-C.2) | Defer | Could add metrics hook in future |

---

## 10. Success Criteria Verification

From PRD:

| Criterion | TDD Section | Implementation |
|-----------|-------------|----------------|
| `/consult` includes skill refs for invocation | 2.2.1 | consult-ref enhancement |
| `/consult` includes workflow skill ref | 2.2.1 | consult-ref enhancement |
| CLAUDE.md Getting Help includes `/consult` | 2.4 | Table row addition |
| `consult-ref` includes cognitive load absorber role | 2.2.1 | New section |
| All teams discoverable via `/teams` | 2.3 | team-ref dynamic discovery |
| `team-discovery` skill created | 2.1 | New skill |
| `/consult` no args displays scannable overview | 2.2.1 | Existing behavior preserved |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-PRD-2-consult-meta-advisor.md` | Created |
