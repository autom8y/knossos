# Doctrine Documentation Refinements
# Sprint: Knossos Doctrine Documentation v2 - Task 003
# Date: 2026-01-08
# Perspective: Future Self returning after 6+ months

---

## Executive Summary

**Assessment Context**: Based on comprehensive audit (Task 001) and information architecture analysis (Task 002)

**Current State**: Philosophically world-class, operationally incomplete. The doctrine demonstrates architectural vision but suffers from **expectation whiplash**—elegant structure promises capabilities not yet documented.

**Critical Finding**: The platform's foundational identity claim (SOURCE vs PROJECTION) is contradictory, and 88% of the CLI surface area is undocumented.

**Refinement Strategy**: Two-phase approach
1. **Phase 1 (Immediate)**: Fix critical issues, collapse empty structure, restore trust
2. **Phase 2 (Future)**: Expand organically as content emerges

**Effort Estimate**:
- Phase 1 (Critical): 4-10 hours → Restores trust, fixes identity crisis
- Phase 2 (High-Value): 8-12 hours → Documents operational reality (CLI, rites, worktrees)
- Total Immediate Impact: 12-22 hours to achieve 100% findability on core scenarios

---

## Gap Analysis

### Critical Gaps (Actively Misleading)

#### G1: SOURCE/PROJECTION Identity Contradiction
**Severity**: CRITICAL - Foundation of platform identity
**Impact**: Engineers cannot determine "What is Knossos?"

**Evidence**:
- **ADR-0009 line 49**: "roster/.claude/ IS Knossos" (claims PROJECTION is Knossos)
- **mythology-concordance.md lines 6-9**: "SOURCE = `/roster/` repository (what Knossos IS)"
- **mythology-concordance.md line 30**: Claims ADR-0009 was "INCORRECT" without formal supersession

**Reality Check**: Codebase behavior confirms SOURCE = `/roster/`, PROJECTION = `.claude/`
- `.claude/` is gitignored
- `ari sync materialize` generates `.claude/` from `/roster/`
- All implementation exists in `/roster/internal/`, `/roster/rites/`, `/roster/cmd/`

**Why This Matters**: Future Self needs authoritative answer to foundational identity question. Two contradictory sources undermine all downstream documentation.

---

#### G2: Invalid Path References
**Severity**: CRITICAL - Breaks "find source code" use case
**Impact**: Engineers following documentation encounter 404

**Evidence**:
- **design-principles.md line 16**: References `/roster/hooks/` (does not exist)
- **mythology-concordance.md line 298**: References `/roster/hooks/` in SOURCE mapping table

**Actual Locations**:
- **Go implementation**: `/roster/internal/hook/` (verified exists)
- **Rite-specific scripts**: `/roster/rites/[rite-name]/hooks/` (verified: ecosystem, intelligence, security, etc.)

**Why This Matters**: Developers attempting to locate hook source code will fail, damaging trust in all path references.

---

### High-Priority Gaps (Impedes Understanding)

#### G3: CLI Command Coverage Gap
**Severity**: HIGH - 88% of surface area undocumented
**Impact**: Future Self unaware of platform capabilities

**Evidence**:
- **Implemented**: 68 commands across 15 families (per COMPLIANCE-STATUS.md)
- **Documented**: ~8 commands mentioned in passing
- **Gap**: 60 commands (88% of CLI) with zero documentation

**Command Families Missing Documentation**:
- worktree (11 commands) - completely absent from doctrine
- session (11 commands) - only examples, no reference
- rite (10 commands) - only conceptual coverage
- hook (6 commands) - mechanism explained, not catalog
- handoff (4 commands) - validation unclear
- inscription (5 commands) - brief mention only
- artifact (3 commands) - not mentioned
- sync (7 commands) - materialize mentioned, others absent
- validate (3 commands) - not mentioned
- manifest (4 commands) - not mentioned
- sails (1 command) - explained conceptually
- naxos (1 command) - detection explained, command not
- tribute (1 command) - vague operational coverage

**Why This Matters**: Operators must explore `ari --help` manually. Major capabilities (worktree parallel sessions, artifact querying) are invisible.

---

#### G4: Empty Directory Structure
**Severity**: HIGH - False expectations damage trust
**Impact**: Navigation promises unfulfilled

**Empty Directories**:
```
docs/doctrine/
├── compliance/
│   ├── status/              ✗ EMPTY
│   ├── quality-gates/       ✗ EMPTY
│   └── certifications/      ✗ EMPTY
├── architecture/
│   ├── core-components/     ✗ EMPTY
│   ├── subsystems/          ✗ EMPTY
│   └── patterns/            ✗ EMPTY
├── evolution/
│   ├── roadmap/             ✗ EMPTY
│   ├── experiments/         ✗ EMPTY
│   └── retrospectives/      ✗ EMPTY
└── operations/
    └── workflows/           ✗ EMPTY
```

**Referenced But Unfulfilled**:
- DOCTRINE.md lines 47-59: Structure diagram shows these directories
- INDEX.md: Points to architecture/, evolution/ content
- Navigation creates false expectations

**Why This Matters**: Future Self explores promising directory, finds nothing, loses confidence in documentation completeness.

---

#### G5: Rite Catalog Missing
**Severity**: HIGH - Major capability invisible
**Impact**: Future Self unaware what rites exist or how to use them

**Evidence**:
- **Implemented**: 12 rites in `/roster/rites/` with manifests
- **Documented**: INDEX.md lists 11 rites with one-line descriptions
- **Reference location**: `docs/doctrine/rites/` directory exists but is completely empty

**Missing Content**:
- Purpose per rite (what problem does it solve?)
- Agent configurations per rite
- When to use each rite
- Invocation patterns
- Workflow phases
- Comparison table for rite selection

**Why This Matters**: Rites are the primary invocation model. Without catalog, users cannot select appropriate rite for their task.

---

#### G6: Worktree System Undocumented
**Severity**: HIGH - Complete capability invisible
**Impact**: Future Self unaware of parallel session capability

**Evidence**:
- **Implemented**: 11 commands (worktree create, switch, list, cleanup, etc.)
- **Documented**: Zero mentions in doctrine (COMPLIANCE-STATUS.md confirms "not mentioned")
- **Capability**: Parallel Claude sessions with filesystem isolation

**Missing Content**:
- Use case explanation (why worktrees exist)
- Command reference (11 commands undocumented)
- Workflow (create → switch → cleanup lifecycle)
- Integration (how worktrees interact with rite invocation)
- Troubleshooting (stale worktrees, lock conflicts)

**Why This Matters**: Parallel sessions are a major platform feature. Unknown unknowns prevent adoption.

---

### Medium-Priority Gaps (Suboptimal but Functional)

#### G7: TLA+ Formal Verification Unmentioned
**Severity**: MEDIUM - Correctness claims lack evidence
**Impact**: Future Self unaware platform has formal verification

**Evidence**:
- **Exists**: `docs/specs/session-fsm.tla` (complete formal specification)
- **Documented**: Zero mentions in doctrine
- **Correctness**: Session FSM formally verified for safety properties

**Missing Content**: Why TLA+ matters, what's verified, how to read spec

---

#### G8: Artifact Registry Undocumented
**Severity**: MEDIUM - Workflow tracking invisible
**Impact**: Future Self unaware of artifact tracking beyond event log

**Evidence**:
- **Implemented**: `/roster/internal/artifact/` with query API
- **Documented**: Zero operational coverage
- **Capability**: Artifact registration, querying, session association

---

#### G9: Tribute System Vaguely Documented
**Severity**: MEDIUM - Session summary unclear
**Impact**: Future Self uncertain what tribute contains or how to generate

**Evidence**:
- **Implemented**: `ari tribute generate` command exists
- **Mythologically mentioned**: Minos tribute in concordance
- **Operationally vague**: Format, content, and generation unclear

---

#### G10: Hook Catalog Missing
**Severity**: MEDIUM - Operational reference absent
**Impact**: Future Self doesn't know what hooks exist or what they do

**Evidence**:
- **Conceptually covered**: knossos-doctrine.md Section IX describes hook mechanism
- **Operationally incomplete**: Lists 5 "Key Hooks" but not exhaustive catalog
- **Reality**: Multiple hook event types (SessionStart, PreToolUse, PostToolUse variants)

---

#### G11: Handoff Validation Gates Unclear
**Severity**: MEDIUM - Quality assurance uncertain
**Impact**: Future Self unsure if handoffs validated before execution

**Evidence**: COMPLIANCE-STATUS.md says "Events recorded, validation gates partial"

---

### Low-Priority Gaps (Polish Items)

#### G12: Implementation Drift Registry Outdated
**Severity**: LOW - Minor contradiction
**Impact**: Drift registry contradicts compliance status

**Evidence**:
- knossos-doctrine.md Section XIV says Naxos "not implemented"
- COMPLIANCE-STATUS.md says Naxos "100% complete"

---

#### G13: Cognitive Budget Tracking Status Unclear
**Severity**: LOW - Ambiguous implementation status
**Impact**: Future Self uncertain if budget warnings work

**Evidence**: Mentioned as "partial" without defining what's partial

---

#### G14: Getting Started Guide Missing
**Severity**: LOW - Learning curve steeper than necessary
**Impact**: New users lack tutorial walkthrough

---

#### G15: Troubleshooting Guide Missing
**Severity**: LOW - Error resolution requires exploration
**Impact**: Common errors lack documented solutions

---

### Gap Summary Table

| Gap | Severity | Type | Documented? | Implemented? | Impact |
|-----|----------|------|-------------|--------------|--------|
| G1: SOURCE/PROJECTION contradiction | CRITICAL | Identity | Contradictory | N/A | Cannot determine what Knossos is |
| G2: Invalid path references | CRITICAL | Accuracy | Incorrect | N/A | Breaks source code findability |
| G3: CLI coverage (88% undocumented) | HIGH | Completeness | Minimal | Complete | Capabilities invisible |
| G4: Empty directory structure | HIGH | Structure | Promised | None | False expectations |
| G5: Rite catalog | HIGH | Completeness | Partial | Complete | Cannot select rites |
| G6: Worktree system | HIGH | Completeness | Absent | Complete | Unknown unknown |
| G7: TLA+ verification | MEDIUM | Completeness | Absent | Complete | Correctness claims unsubstantiated |
| G8: Artifact registry | MEDIUM | Completeness | Absent | Complete | Workflow tracking invisible |
| G9: Tribute system | MEDIUM | Clarity | Vague | Complete | Session summary unclear |
| G10: Hook catalog | MEDIUM | Completeness | Partial | Complete | Hook enumeration missing |
| G11: Handoff validation | MEDIUM | Clarity | Unclear | Partial | Quality gates uncertain |
| G12: Drift registry stale | LOW | Accuracy | Contradictory | N/A | Minor confusion |
| G13: Budget tracking status | LOW | Clarity | Vague | Partial | Implementation status unclear |
| G14: Getting started guide | LOW | Completeness | Absent | N/A | Steeper learning curve |
| G15: Troubleshooting guide | LOW | Completeness | Absent | N/A | Error resolution harder |

---

## Specific Refinements

### 1. ADR-0009 Amendment

**Issue**: Line 49 states "roster/.claude/ IS Knossos" which contradicts actual implementation and mythology-concordance.md

**Required Action**: Create superseding amendment clarifying SOURCE vs PROJECTION

**Exact Amendment Text**:

```markdown
## Amendment: SOURCE/PROJECTION Clarification (2026-01-08)

**Context**: The original identity statement "roster/.claude/ IS Knossos" (line 49) was imprecise and led to downstream confusion in documentation.

**Clarification**:

The relationship between SOURCE and PROJECTION is:

| Term | Definition | Example Paths |
|------|------------|---------------|
| **SOURCE** | The `/roster/` repository (versioned, canonical, what Knossos IS) | `/roster/internal/`, `/roster/rites/`, `/roster/cmd/ari/` |
| **PROJECTION** | The `.claude/` directories (gitignored, materialized by `ari sync materialize`) | `.claude/hooks/`, `.claude/agents/`, `.claude/sessions/` |

**Corrected Identity Statement**:

**The `/roster/` repository IS Knossos.** The `.claude/` directory structure is a PROJECTION of Knossos—generated artifacts materialized into consuming projects.

**Rationale**:
- Knossos is the SOURCE (the palace/labyrinth itself)
- `.claude/` is the PROJECTION (the rooms within, generated for each project)
- The labyrinth is the palace, not the rooms within it
- Implementation exists in `/roster/`, not `.claude/`

**Impact on Documentation**:
- mythology-concordance.md correctly states this relationship (lines 6-9)
- This amendment supersedes the original line 49 identity statement
- All subsequent documentation should reference SOURCE = `/roster/` when locating implementation

**See Also**: `docs/doctrine/philosophy/mythology-concordance.md` for comprehensive SOURCE/PROJECTION mapping
```

**Files to Update After Amendment**:
1. **ADR-0009**: Add amendment section (above)
2. **mythology-concordance.md line 30**: Remove claim that ADR-0009 was "INCORRECT"—replace with reference to amendment
3. **DOCTRINE.md**: Ensure identity statement aligns with corrected version

---

### 2. Path Reference Corrections

**Issue**: Multiple files reference `/roster/hooks/` which does not exist

**Files Requiring Correction**:

#### File 1: `docs/doctrine/philosophy/design-principles.md`

**Line 16 - Current**:
```markdown
Hook-based event capture (source: `/roster/hooks/`, materialized: `.claude/hooks/`)
```

**Line 16 - Corrected**:
```markdown
Hook-based event capture (Go implementation: `/roster/internal/hook/`, rite-specific scripts: `/roster/rites/[rite]/hooks/`, materialized: `.claude/hooks/`)
```

---

#### File 2: `docs/doctrine/philosophy/mythology-concordance.md`

**Line 298 - Current** (in SOURCE/PROJECTION mapping table):
```markdown
`/roster/hooks/` → `.claude/hooks/`
```

**Line 298 - Corrected**:
```markdown
`/roster/internal/hook/` (Go) + `/roster/rites/[rite]/hooks/` (scripts) → `.claude/hooks/`
```

---

**Additional Context to Add** (mythology-concordance.md, after corrected line):
```markdown
> **Note**: Hook implementation is two-tiered:
> - **Platform hooks** (Go): `/roster/internal/hook/` - Core event handling
> - **Rite-specific hooks** (shell): `/roster/rites/[rite-name]/hooks/` - Per-rite customization
> - **Materialized output**: `.claude/hooks/` - Generated runtime hooks
```

---

### 3. Structural Changes

#### Remove Empty Level-3 Directories

**Directories to Delete** (create false navigation expectations):
```bash
rm -rf docs/doctrine/compliance/status/
rm -rf docs/doctrine/compliance/quality-gates/
rm -rf docs/doctrine/compliance/certifications/
rm -rf docs/doctrine/architecture/core-components/
rm -rf docs/doctrine/architecture/subsystems/
rm -rf docs/doctrine/architecture/patterns/
rm -rf docs/doctrine/evolution/roadmap/
rm -rf docs/doctrine/evolution/experiments/
rm -rf docs/doctrine/evolution/retrospectives/
rm -rf docs/doctrine/operations/workflows/
```

**Rationale**: Structure should reflect what exists, not aspirations. Empty directories are navigation friction.

---

#### Preserve with README Placeholders

**Directories to Keep** (immediate population planned):

**`docs/doctrine/compliance/audits/README.md`**:
```markdown
# Compliance Audits

**Status**: Active - audits in progress
**Current**: doctrine-audit-20260108.md (moved from parent directory)

Future audits will be added here as documentation evolves.
```

**`docs/doctrine/operations/cli-reference/README.md`**:
```markdown
# CLI Reference - In Progress

**Status**: Under development (Refinement R3)
**Target**: Comprehensive reference for all 68 `ari` commands

See [Refinement Recommendations](../../audits/refinement-recommendations-20260108.md) for implementation plan.

Temporary: Use `ari --help` and `ari [command] --help` for command documentation.
```

**`docs/doctrine/rites/README.md`**:
```markdown
# Rite Catalog - In Progress

**Status**: Under development (Refinement R4)
**Target**: Catalog of all 12 rites with purpose, agents, and invocation patterns

See [Refinement Recommendations](../audits/refinement-recommendations-20260108.md) for implementation plan.

Temporary: See INDEX.md for rite summaries, or explore `/roster/rites/` for manifests.
```

---

#### Move Misplaced Content

**Action**: Move existing audit to proper location
```bash
mv docs/doctrine/audits/doctrine-audit-20260108.md docs/doctrine/compliance/audits/
rmdir docs/doctrine/audits/
```

**Update Cross-References**:
- Update any links pointing to old path
- Verify symlinks (none currently point to audits/)

---

### 4. Content Additions (Prioritized)

#### Tier 1: Critical (Do First - 4-10 hours total)

**C1: Fix Path References** (15 minutes)
- Edit design-principles.md line 16
- Edit mythology-concordance.md line 298
- Add clarifying note about two-tiered hook implementation

**C2: Create ADR-0009 Amendment** (1 hour)
- Add amendment section to ADR-0009
- Update mythology-concordance.md to reference amendment
- Verify DOCTRINE.md aligns with corrected identity

**C3: Collapse Empty Directories** (1 hour)
- Delete 10 empty level-3 directories
- Add README placeholders to 3 preserved directories
- Move audit to compliance/audits/
- Update DOCTRINE.md structure diagram

**C4: CLI Reference Foundation** (2-8 hours)
- **Option A (Automated)**: Write extraction script from `ari --help` output (2 hours)
- **Option B (Manual)**: Document command families manually (4-8 hours)

**Recommended Format** (per command):
```markdown
# `ari session create`

**Purpose**: Initialize a new work session with tracking and state management

**Usage**:
```bash
ari session create <initiative> <complexity>
```

**Arguments**:
- `initiative` (string): Session title/goal
- `complexity` (enum): SIMPLE | MODERATE | COMPLEX | EPIC

**Examples**:
```bash
ari session create "Add dark mode toggle" MODERATE
ari session create "Refactor authentication system" COMPLEX
```

**Behavior**:
- Creates session directory in `.claude/sessions/`
- Initializes SESSION_CONTEXT.md with FSM state
- Records SessionCreated event in clew
- Generates session ID (UUID)

**Output**:
- Session ID
- Session path
- Initial state (ACTIVE)

**Related Commands**:
- `ari session status` - View current session state
- `ari session park` - Pause session
- `ari session wrap` - Complete session with quality gates

**See Also**:
- [Session Lifecycle](../../philosophy/knossos-doctrine.md#section-v-session-lifecycle)
- [Moirai Authority](../../philosophy/mythology-concordance.md#moirai)
```

**Command Family Organization**:
```
operations/cli-reference/
├── index.md              # Overview, all families
├── session.md            # 11 session commands
├── rite.md               # 10 rite commands
├── worktree.md           # 11 worktree commands
├── hook.md               # 6 hook commands
├── handoff.md            # 4 handoff commands
├── inscription.md        # 5 inscription commands
├── artifact.md           # 3 artifact commands
├── sync.md               # 7 sync commands
├── validate.md           # 3 validate commands
├── manifest.md           # 4 manifest commands
└── utilities.md          # sails, naxos, tribute (3 commands)
```

---

#### Tier 2: High-Value (Do Soon - 8-12 hours total)

**C5: Rite Catalog** (2-3 hours)

**Location**: `docs/doctrine/rites/catalog.md`

**Structure**:
```markdown
# Rite Catalog

> Comprehensive catalog of all available rites with purpose, agents, and invocation patterns

**Audience**: Practitioners selecting rites, architects understanding workflows

---

## Quick Reference

| Rite | Purpose | Primary Use Case | Agents | Invocation |
|------|---------|------------------|--------|------------|
| 10x-dev | Full dev lifecycle | Feature implementation | 5 | `/task`, `/sprint` |
| docs | Documentation workflow | Technical writing | 5 | `/docs` |
| hygiene | Code quality | Debt reduction | 3 | `/hygiene` |
| debt-triage | Technical debt | Debt assessment | 4 | `/debt` |
| security | Security assessment | Vulnerability analysis | 4 | `/security` |
| sre | Reliability workflow | Production operations | 4 | `/sre` |
| intelligence | Product analytics | Data-driven insights | 3 | `/intelligence` |
| strategy | Business strategy | Strategic planning | 3 | `/strategy` |
| rnd | Innovation lab | Research & exploration | 3 | `/rnd` |
| ecosystem | CEM infrastructure | Roster ecosystem work | 4 | `/ecosystem` |
| forge | Meta-rite overview | Rite management | - | `/forge` |
| [discovery others via pantheon] | - | - | - | `ari rite pantheon` |

---

## Rites

### 10x-dev (Full Development Lifecycle)

**Purpose**: Feature implementation from requirements to production-ready code

**Agents**:
1. **Orchestrator** - Coordinates workflow phases
2. **Planner** - Creates PRD from user requirements
3. **Architect** - Designs technical approach (TDD)
4. **Builder** - Implements code from design
5. **QA Adversary** - Validates implementation

**When to Use**:
- Implementing new features
- Building from requirements to deployment
- Need integrated design-to-test workflow

**Invocation Patterns**:
```bash
/task "implement feature X"
/sprint "goals for multi-task sprint"
Task(orchestrator, "implement X")
```

**Key Workflows**:
- Requirements → PRD → TDD → Implementation → QA → PR
- Phases can be invoked individually or as full pipeline

**Source**: `/roster/rites/10x-dev/manifest.yaml`

---

### docs (Documentation Workflow)

**Purpose**: Documentation creation and improvement through structured workflow

**Agents**:
1. **Orchestrator** - Coordinates documentation phases
2. **Doc Auditor** - Audits existing docs, identifies gaps
3. **Information Architect** - Designs doc structure
4. **Tech Writer** - Writes clear documentation
5. **Doc Reviewer** - Reviews for accuracy and quality

**When to Use**:
- Creating API documentation
- Writing technical guides
- Auditing doc quality
- Restructuring documentation

**Invocation Patterns**:
```bash
/docs "document X"
Task(orchestrator, "audit and improve documentation for Y")
```

**Key Workflows**:
- Audit → Architecture → Writing → Review
- Can start at any phase depending on needs

**Source**: `/roster/rites/docs/manifest.yaml`

---

[... Continue for all 12 rites ...]

---

## Rite Selection Guide

### By Intent

**I want to build a feature**: → 10x-dev
**I want to document something**: → docs
**I want to reduce technical debt**: → hygiene or debt-triage
**I want to assess security**: → security
**I want to improve reliability**: → sre
**I want data insights**: → intelligence
**I want strategic planning**: → strategy
**I want to explore/research**: → rnd
**I want to work on roster ecosystem**: → ecosystem

### By Team Size

- **1-3 agents**: strategy, intelligence, rnd, hygiene
- **4-5 agents**: 10x-dev, docs, debt-triage, security, sre, ecosystem

### By Workflow Complexity

- **Simple**: Single-phase invocation
- **Complex**: Multi-phase orchestration with handoffs

---

## See Also

- [Rite System Overview](../philosophy/knossos-doctrine.md#section-iv-the-rite-system)
- [Rite Commands](../operations/cli-reference/rite.md)
- [Mythology: Rites as Ceremonies](../philosophy/mythology-concordance.md)
```

**Data Source**: Extract from `/roster/rites/*/manifest.yaml` files

---

**C6: Worktree Guide** (2-3 hours)

**Location**: `docs/doctrine/operations/guides/worktree-guide.md`

**Structure**:
```markdown
# Worktree Guide

> Complete guide to worktree system for parallel session isolation

**Audience**: Operators running parallel Claude sessions, SREs managing platform

---

## Overview

### What Are Worktrees?

Worktrees enable **parallel Claude Code sessions** with **filesystem isolation**. Each worktree is a separate working directory sharing the same Git repository but with independent:
- `.claude/` directories
- Active sessions
- Rite invocations
- File system state

### Why Worktrees Exist

**Use Cases**:
1. **Parallel sessions**: Run multiple Claude terminals simultaneously
2. **Rite isolation**: Different rites per terminal (10x-dev in one, docs in another)
3. **Sprint isolation**: Separate worktrees per sprint or initiative
4. **Context switching**: Switch between tasks without session conflicts

**Without Worktrees**: Multiple Claude sessions conflict over `.claude/sessions/current-session`
**With Worktrees**: Each worktree maintains independent session state

---

## Workflow

### Creating a Worktree

```bash
ari worktree create <name> [branch]
```

**Example**:
```bash
ari worktree create docs-sprint main
# Creates worktree at ../roster-docs-sprint/
# Branches from main
# Independent .claude/ directory
```

**What Happens**:
- Git worktree created at `../<repo>-<name>/`
- `.claude/` materialized in new worktree
- Worktree registered in roster's worktree list
- Ready for independent session

---

### Switching Between Worktrees

**List Available**:
```bash
ari worktree list
# Output: main, docs-sprint, feature-x
```

**Switch Terminal to Worktree**:
```bash
cd ../<repo>-<name>/
# Now in isolated worktree
# ari session create will use this worktree's .claude/
```

**State Management**:
- Each worktree maintains independent session state
- Sessions don't conflict across worktrees
- Can have ACTIVE session in multiple worktrees simultaneously

---

### Cleanup

**Remove Worktree**:
```bash
ari worktree remove <name>
```

**Prune Stale Worktrees**:
```bash
ari worktree prune
# Removes worktrees with deleted directories
```

**Best Practice**: Always wrap or park sessions before removing worktree

---

## Command Reference

| Command | Purpose |
|---------|---------|
| `ari worktree create <name> [branch]` | Create new worktree |
| `ari worktree list` | List all worktrees |
| `ari worktree remove <name>` | Remove worktree |
| `ari worktree prune` | Clean up stale worktrees |
| `ari worktree switch <name>` | Switch to worktree (if supported) |
| `ari worktree status` | Show current worktree info |
| [... 5 more commands ...] | [Enumerate from CLI] |

---

## Integration with Rites

**Scenario**: Run 10x-dev in one terminal, docs in another

**Terminal 1** (main worktree):
```bash
cd /path/to/roster
ari session create "Feature implementation" MODERATE
/task "implement feature X"
# 10x-dev rite active
```

**Terminal 2** (docs worktree):
```bash
ari worktree create docs main
cd ../roster-docs
ari session create "Document feature X" SIMPLE
/docs "document feature X"
# docs rite active
```

**Result**: Both sessions run independently, no state conflicts

---

## Troubleshooting

### Stale Worktree Error

**Symptom**: `ari worktree list` shows worktree but directory doesn't exist

**Cause**: Worktree directory deleted outside of `ari worktree remove`

**Solution**:
```bash
ari worktree prune
# Cleans up stale references
```

---

### Lock Conflict Across Worktrees

**Symptom**: Lock acquisition fails despite being in different worktree

**Cause**: Locks are repository-scoped, not worktree-scoped

**Solution**: Ensure operations don't conflict at repository level (e.g., git operations)

---

### Session Confusion

**Symptom**: `ari session status` shows different session than expected

**Cause**: Ran command in wrong worktree

**Solution**: Check current directory—ensure you're in intended worktree

---

## See Also

- [Session Management](./knossos-integration.md#session-management)
- [Parallel Sessions Guide](./parallel-sessions.md)
- [Worktree CLI Reference](../cli-reference/worktree.md)
```

---

**C7: Update DOCTRINE.md Structure** (30 minutes)

After collapsing empty directories and adding content, update structure diagram in DOCTRINE.md to reflect reality:

**Current** (lines 47-59):
```markdown
docs/doctrine/
├── philosophy/         # The "Why" - foundational principles
├── foundations/        # Architectural decisions (ADRs)
├── reference/          # Navigation and terminology
├── compliance/         # Achievement tracking
├── architecture/       # System design
│   ├── core-components/
│   ├── subsystems/
│   └── patterns/
├── operations/         # Practical guides
│   ├── guides/
│   ├── cli-reference/
│   └── workflows/
├── rites/              # Rite catalog
└── evolution/          # Roadmap and retrospectives
```

**Corrected**:
```markdown
docs/doctrine/
├── DOCTRINE.md             # This file - entry point
├── philosophy/             # The "Why" - foundational principles
│   ├── knossos-doctrine.md
│   ├── design-principles.md
│   └── mythology-concordance.md
├── foundations/            # Architectural decisions (symlinks to ../../decisions/)
│   ├── ADR-0001-session-state-machine-redesign.md
│   ├── ADR-0005-moirai-centralized-state-authority.md
│   └── ADR-0009-knossos-roster-identity.md
├── reference/              # Navigation and terminology
│   ├── INDEX.md
│   └── GLOSSARY.md
├── compliance/             # Achievement tracking and audits
│   ├── COMPLIANCE-STATUS.md
│   └── audits/
│       ├── doctrine-audit-20260108.md
│       ├── ia-assessment-20260108.md
│       └── refinement-recommendations-20260108.md
├── operations/             # Practical guides and CLI reference
│   ├── guides/             # Symlinks to ../../../guides/
│   │   ├── ariadne-cli.md
│   │   ├── knossos-integration.md
│   │   ├── parallel-sessions.md
│   │   ├── user-preferences.md
│   │   └── white-sails.md
│   └── cli-reference/      # [IN PROGRESS] 68 commands across 15 families
│       └── README.md
└── rites/                  # [IN PROGRESS] Catalog of 12 rites
    └── README.md
```

**Note Section to Add**:
```markdown
### Structure Notes

**In Progress**:
- `operations/cli-reference/` - CLI documentation under development (see compliance/audits/)
- `rites/` - Rite catalog under development

**Removed** (2026-01-08):
- Empty level-3 directories collapsed (architecture/, evolution/ subdirs)
- Structure now reflects actual content, not aspirations
- Will expand organically as content emerges

**Symlinks**:
- `foundations/` → `../../decisions/` (ADRs)
- `operations/guides/` → `../../../guides/` (operational guides)
```

---

#### Tier 3: Quality Improvements (Defer Until Tier 1-2 Complete)

**C8: Getting Started Guide** (3-4 hours)
- Tutorial walkthrough of first session
- Prerequisites, installation, environment
- Key concepts with examples
- Common patterns

**C9: Troubleshooting Guide** (2-3 hours)
- Session state issues
- Hook failures
- Rite invocation errors
- Recovery procedures

**C10: Workflow Guides** (2-3 hours)
- Session lifecycle diagrams
- Handoff protocols
- Quality gates

**C11: TLA+ Verification Explanation** (1-2 hours)
- What's formally verified
- Why it matters
- How to read spec

**C12: Artifact Registry Documentation** (1-2 hours)
- Registry purpose
- Query API
- Session association

**C13: Expand Hook Documentation** (2-3 hours)
- Complete hook catalog
- Event types exhaustive list
- Hook customization guide

**C14: Update Drift Registry** (15 minutes)
- Mark Naxos COMPLETE
- Clarify cognitive budget status
- Update tribute implementation status

---

## Implementation Order

### Phase 1: Trust Repair (Week 1 - 3-4 hours)

**Objective**: Fix critical issues that undermine trust in documentation

**Tasks**:
1. ✓ **Path corrections** (15 min)
   - Edit design-principles.md line 16
   - Edit mythology-concordance.md line 298
   - Add two-tiered hook implementation note

2. ✓ **ADR-0009 amendment** (1 hour)
   - Add amendment section clarifying SOURCE/PROJECTION
   - Update mythology-concordance.md reference
   - Verify DOCTRINE.md alignment

3. ✓ **Collapse empty directories** (1 hour)
   - Delete 10 empty level-3 directories
   - Add README placeholders to preserved dirs
   - Move audit to compliance/audits/

4. ✓ **Update DOCTRINE.md structure** (30 min)
   - Correct directory diagram
   - Add structure notes section
   - Update cross-references

**Deliverables**:
- ✓ Path references accurate
- ✓ SOURCE/PROJECTION doctrine unified
- ✓ Navigation expectations aligned with reality
- ✓ Trust restored in technical accuracy

**Validation**: Can Future Self locate session lifecycle implementation without encountering 404?

---

### Phase 2: Operational Documentation (Weeks 2-3 - 8-12 hours)

**Objective**: Document operational reality (CLI, rites, worktrees)

**Tasks**:
1. ✓ **CLI Reference** (4-8 hours)
   - Option A: Automated extraction from --help (2 hours)
   - Option B: Manual documentation (4-8 hours)
   - 15 command family files
   - Index with navigation

2. ✓ **Rite Catalog** (2-3 hours)
   - Extract from manifest.yaml files
   - Quick reference table
   - Per-rite sections
   - Selection guide

3. ✓ **Worktree Guide** (2-3 hours)
   - Concept explanation
   - Command reference (11 commands)
   - Workflow examples
   - Troubleshooting section

**Deliverables**:
- ✓ 68 commands documented
- ✓ 12 rites cataloged
- ✓ Worktree system explained
- ✓ Findability increased from 37.5% → ~80%

**Validation**: Can Future Self discover how to run parallel sessions without source code exploration?

---

### Phase 3: Quality & Architecture (Weeks 4-6 - 12-18 hours)

**Objective**: Polish UX and document architecture for implementers

**Tasks**:
1. ✓ **Getting Started Guide** (3-4 hours)
2. ✓ **Troubleshooting Guide** (2-3 hours)
3. ✓ **TLA+ Verification Explanation** (1-2 hours)
4. ✓ **Artifact Registry Documentation** (1-2 hours)
5. ✓ **Hook Catalog Expansion** (2-3 hours)
6. ✓ **Workflow Guides** (2-3 hours)
7. ✓ **Update Drift Registry** (15 min)

**Deliverables**:
- ✓ New user onboarding streamlined
- ✓ Error resolution documented
- ✓ Architecture docs for implementers
- ✓ Findability approaches 100%

**Validation**: Can Future Self successfully complete first session using only documentation?

---

## Success Metrics

### Findability Test (30-Second Rule)

Future Self should find answers in under 30 seconds:

| Question | Target Location | Before | After Phase 1 | After Phase 2 |
|----------|----------------|--------|---------------|---------------|
| What is Knossos? | `DOCTRINE.md` | ✓ YES | ✓ YES | ✓ YES |
| Why this architecture? | `philosophy/knossos-doctrine.md` | ✓ YES | ✓ YES | ✓ YES |
| How do I create a session? | `cli-reference/session.md` | ✗ NO | ✗ NO | ✓ YES |
| What rites exist? | `rites/catalog.md` | ✗ NO | ✗ NO | ✓ YES |
| Where is Moirai implemented? | `mythology-concordance.md` | ⚠ CONFUSING | ✓ YES | ✓ YES |
| What commands are available? | `cli-reference/` | ✗ NO | ✗ NO | ✓ YES |
| Current compliance status? | `compliance/COMPLIANCE-STATUS.md` | ✓ YES | ✓ YES | ✓ YES |
| Can I run parallel sessions? | `worktree-guide.md` | ✗ NO | ✗ NO | ✓ YES |

**Current Pass Rate**: 3/8 (37.5%)
**After Phase 1**: 5/8 (62.5%)
**After Phase 2**: 8/8 (100%)

---

### Accuracy Test

All documented paths must resolve:

| Path Reference | Before | After Phase 1 |
|----------------|--------|---------------|
| `/roster/hooks/` | ✗ BROKEN | ✓ FIXED |
| `/roster/internal/hook/` | ✓ VALID | ✓ VALID |
| `/roster/rites/[rite]/hooks/` | ✓ VALID | ✓ VALID |
| SOURCE/PROJECTION identity | ✗ CONTRADICTORY | ✓ UNIFIED |

**Before**: 50% path accuracy (1 broken, 1 contradictory)
**After Phase 1**: 100% path accuracy

---

### Completeness Test

Platform capabilities documented:

| Capability | Implemented | Documented Before | After Phase 2 |
|------------|-------------|-------------------|---------------|
| Session lifecycle | ✓ | ✓ | ✓ |
| White Sails | ✓ | ✓ | ✓ |
| Moirai authority | ✓ | ✓ | ✓ |
| Rite system | ✓ | Partial | ✓ |
| CLI commands | ✓ (68) | Minimal (8) | ✓ (68) |
| Worktree system | ✓ (11) | ✗ | ✓ (11) |
| Hooks | ✓ | Partial | Partial → ✓ (Phase 3) |
| Artifact registry | ✓ | ✗ | ✗ → ✓ (Phase 3) |
| TLA+ verification | ✓ | ✗ | ✗ → ✓ (Phase 3) |

**Before**: 40% capability coverage
**After Phase 2**: 85% capability coverage
**After Phase 3**: 100% capability coverage

---

### Trust Test

**Question**: "Can Future Self trust that documented information is accurate?"

**Before**:
- ✗ Identity claim contradictory
- ✗ Path references broken
- ✗ Structure promises unfulfilled

**After Phase 1**:
- ✓ Identity unified and authoritative
- ✓ All paths verified correct
- ✓ Structure aligned with reality

**After Phase 2**:
- ✓ Operational reality documented
- ✓ Major capabilities visible
- ✓ CLI surface area covered

---

## Conclusion

### The Diagnosis

The Knossos doctrine demonstrates **architectural vision** and **philosophical coherence** but suffers from **premature scaffolding** and **critical accuracy gaps**.

**What Works**:
- Mythological framing encodes design intent
- Philosophy section is world-class
- Voice consistency maintained throughout
- Symlink strategy creates unified view

**What's Broken**:
- Identity crisis (SOURCE vs PROJECTION contradiction)
- Invalid paths break source code findability
- Empty directories create false expectations
- 88% of CLI undocumented

### The Prescription

**Two-Phase Strategy**:

**Phase 1: Consolidation** → Fix what's broken, collapse false promises, restore trust
**Phase 2: Documentation** → Document operational reality (CLI, rites, worktrees)

**Principle**: **Structure follows content, not the reverse.**

### The Outcome

**After Implementation**:
- ✓ Future Self can locate source code reliably
- ✓ Platform identity is authoritative and unified
- ✓ Operational capabilities are discoverable
- ✓ Navigation expectations align with reality
- ✓ Findability: 37.5% → 100%
- ✓ Trust: restored through accuracy and completeness

### The Acid Test

*Can Future Self, returning after 6 months with no context, navigate this documentation to accomplish a task without frustration?*

**Current Answer**: No—philosophy findable, operations fragmented, trust damaged
**After Phase 1**: Mostly—trust restored, identity clear, structure honest
**After Phase 2**: **Yes**—philosophy intact, operations documented, capabilities visible

---

## Appendix: Content Brief Templates

### CLI Reference Entry Template

```markdown
# `ari [family] [command]`

**Purpose**: [One sentence - what this command does]

**Usage**:
```bash
ari [family] [command] [args]
```

**Arguments**:
- `arg1` (type): Description
- `arg2` (type): Description

**Options**:
- `--flag`: Description

**Examples**:
```bash
[Realistic example 1]
[Realistic example 2]
```

**Behavior**:
- [What happens when command runs]
- [Side effects, state changes]
- [File system operations]

**Output**:
- [What command prints]
- [Exit codes]

**Related Commands**:
- `ari [related]` - [Why related]

**See Also**:
- [Relevant doctrine section]
```

---

### Rite Entry Template

```markdown
### [Rite Name]

**Purpose**: [One sentence - problem this rite solves]

**Agents**:
1. **[Agent 1]** - [Role]
2. **[Agent 2]** - [Role]
[...]

**When to Use**:
- [Scenario 1]
- [Scenario 2]

**Invocation Patterns**:
```bash
/[shorthand] "[goal]"
Task([primary-agent], "[directive]")
```

**Key Workflows**:
- [Phase 1] → [Phase 2] → [Phase 3]
- [Alternative workflow if applicable]

**Source**: `/roster/rites/[rite-name]/manifest.yaml`

---
```

---

### Guide Section Template

```markdown
## [Section Title]

### [Subsection - What]

[1-2 paragraph explanation of concept]

### [Subsection - Why]

**Use Cases**:
1. [Use case 1]
2. [Use case 2]

### [Subsection - How]

**[Task]**:
```bash
[Command or procedure]
```

**What Happens**:
- [Step-by-step breakdown]

**Example**:
```bash
[Concrete example with output]
```

---
```

---

## Document Metadata

**Author**: Tech Writer Agent (with Doc Auditor and Information Architect consultation)
**Date**: 2026-01-08
**Sprint**: Knossos Doctrine Documentation v2
**Task**: 003 - Gap Analysis & Refinement Recommendations
**Prerequisites**:
- Doctrine Audit (Task 001)
- IA Assessment (Task 002)
**Confidence Level**: HIGH
**Evidence Quality**: Direct audit findings + IA analysis + codebase verification

---

*The labyrinth's architecture is sound. The thread is nearly complete. Now weave the final strands.*

*May the clew guide Future Self home, and may the White Sails fly true.*
