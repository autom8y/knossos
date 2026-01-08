# Information Architecture Assessment - Knossos Doctrine
# Sprint: Knossos Doctrine Documentation v2 - Task 002
# Date: 2026-01-08
# Architect: Information Architect Agent
# Perspective: Future Self returning after 6+ months

---

## Executive Summary

**Assessment Scope**: Structure, navigation, and mental model alignment of `docs/doctrine/` directory
**Structure Score**: 6/10 (Architecturally sound foundations, but hollow promises undermine trust)
**Findability Score**: 4.5/10 (Philosophy discoverable; operations fragmented or absent)
**Overall Grade**: **C+ (Promising architecture with critical execution gaps)**

**Critical Finding**: The structure promises a complete reference architecture but delivers empty scaffolding. Future Self would experience **expectation whiplash**—seduced by elegant taxonomy, then frustrated by empty directories and missing operational content.

**Primary Recommendation**: Adopt a **Two-Phase Consolidation Strategy**:
1. **Phase 1 (Immediate)**: Collapse empty structure, fix critical paths, consolidate what exists
2. **Phase 2 (Future)**: Expand structure organically as content emerges (not before)

---

## 1. Structure Evaluation

### 1.1 The Taxonomy (Directory Organization)

```
docs/doctrine/
├── DOCTRINE.md              ✓ EFFECTIVE (clear entry point)
├── philosophy/              ✓ EXCELLENT (complete, coherent)
│   ├── knossos-doctrine.md
│   ├── design-principles.md
│   └── mythology-concordance.md
├── foundations/             ✓ GOOD (symlinks work well)
├── compliance/              ⚠ PARTIALLY EFFECTIVE (status exists, subdirs empty)
│   ├── COMPLIANCE-STATUS.md ✓
│   ├── status/              ✗ EMPTY
│   ├── audits/              ✗ EMPTY (audit is in parent dir)
│   ├── quality-gates/       ✗ EMPTY
│   └── certifications/      ✗ EMPTY
├── architecture/            ✗ COMPLETELY HOLLOW (3 empty subdirs)
│   ├── core-components/     ✗ EMPTY
│   ├── subsystems/          ✗ EMPTY
│   └── patterns/            ✗ EMPTY
├── operations/              ⚠ FRAGMENTED (guides work, cli-ref/workflows empty)
│   ├── guides/              ✓ GOOD (5 symlinks to ../../guides/)
│   ├── cli-reference/       ✗ EMPTY (68 commands undocumented)
│   └── workflows/           ✗ EMPTY
├── rites/                   ✗ EMPTY (12 rites in /roster/rites/, none documented here)
├── evolution/               ✗ COMPLETELY HOLLOW (3 empty subdirs)
│   ├── roadmap/             ✗ EMPTY
│   ├── experiments/         ✗ EMPTY
│   └── retrospectives/      ✗ EMPTY
└── reference/               ✓ EXCELLENT (INDEX + GLOSSARY both strong)
```

**Structural Diagnosis**:
- **Philosophy**: World-class. Complete, coherent, deeply integrated.
- **Reference**: Excellent. INDEX.md and GLOSSARY.md both deliver.
- **Foundations + Operations/guides**: Good symlink strategy creates unified view without duplication.
- **Everything else**: Architectural fiction. Empty directories promise capabilities not yet delivered.

**Score: 6/10**
- +4 for philosophy, reference, and symlink strategy (strong core)
- -2 for empty compliance subdirectories (organizational clutter)
- -2 for completely hollow architecture/, evolution/, rites/ (expectation mismatch)

### 1.2 The Directory Depth Analysis

| Level | Path | Files | Assessment |
|-------|------|-------|------------|
| 1 | `docs/doctrine/` | 1 | ✓ Appropriate entry point |
| 2 | `philosophy/`, `reference/` | 5 | ✓ Flat, navigable |
| 2 | `compliance/` | 1 | ⚠ Parent has content, subdirs empty |
| 2 | `foundations/`, `operations/guides/` | 8 symlinks | ✓ Symlinks effective |
| 3 | `compliance/status/`, `audits/`, etc. | 0 | ✗ Premature depth, no content |
| 3 | `architecture/*`, `evolution/*` | 0 | ✗ Premature depth, no content |
| 2 | `operations/cli-reference/`, `workflows/` | 0 | ✗ Missing high-value content |
| 2 | `rites/` | 0 | ✗ Empty despite 12 rites in `/roster/rites/` |

**Depth Philosophy Assessment**:
- **Good**: Philosophy and reference dirs are appropriately flat (2 levels max)
- **Problematic**: Empty 3-level hierarchies (compliance/audits/, architecture/patterns/) add navigation cost with no content benefit
- **Anti-Pattern**: Creating directory structure before content exists

**Recommendation**: Collapse all empty level-3 directories. Reinstate only when content exists to populate them.

---

## 2. Mental Model Alignment

### 2.1 Future Self's Journey Scenarios

#### Scenario A: "What is Knossos and why does it exist?"

**Expected Path**: `DOCTRINE.md` → `philosophy/knossos-doctrine.md`

**Actual Experience**: ✓ **EXCELLENT**
- DOCTRINE.md provides clear entry point with navigation table
- philosophy/knossos-doctrine.md delivers complete philosophical foundation
- Voice is consistent, depth is appropriate
- **Time to Answer**: <2 minutes

**Mental Model Score**: 10/10

---

#### Scenario B: "Where is the session lifecycle implemented?"

**Expected Path**: `mythology-concordance.md` → `/roster/internal/session/`

**Actual Experience**: ⚠ **CONFUSING**
- mythology-concordance.md says SOURCE = `/roster/`
- But contradicts ADR-0009 which says "roster/.claude/ IS Knossos"
- Correct path exists but trust is undermined by doctrinal conflict
- **Time to Answer**: 3-5 minutes (with doubt)

**Mental Model Score**: 5/10 (correct answer, damaged confidence)

---

#### Scenario C: "What architectural decisions have been made?"

**Expected Path**: `DOCTRINE.md` directory diagram shows `architecture/` → expect TDDs, design docs

**Actual Experience**: ✗ **BROKEN PROMISE**
- Directory structure in DOCTRINE.md lines 47-59 shows:
  - `architecture/core-components/`
  - `architecture/subsystems/`
  - `architecture/patterns/`
- All three subdirectories: **completely empty**
- Symlinks in `foundations/` point to ADRs in `../../decisions/` (which work)
- But no TDDs, no component design docs, no pattern catalog
- **Time to Answer**: FAILED (content doesn't exist)

**Mental Model Score**: 2/10 (false advertising damages trust)

---

#### Scenario D: "How do I use the `ari` CLI?"

**Expected Path**: `operations/cli-reference/` (promised in DOCTRINE.md structure)

**Actual Experience**: ✗ **ABSENT**
- `operations/cli-reference/` directory exists but is empty
- 68 commands across 15 families implemented (per COMPLIANCE-STATUS.md)
- ~8 commands mentioned in passing across doctrine docs
- No systematic CLI reference documentation
- **Time to Answer**: FAILED (must use `ari --help` or read source)

**Mental Model Score**: 1/10 (critical operational gap)

---

#### Scenario E: "What rites are available and how do they work?"

**Expected Path**: `rites/` directory (promised in DOCTRINE.md + INDEX.md)

**Actual Experience**: ✗ **COMPLETELY MISSING**
- `rites/` directory exists but is empty
- INDEX.md lines 118-137 lists 11 rites with descriptions
- But points to `../rites/` which has zero files
- Actual rite SOURCE exists at `/roster/rites/` (12 rites with manifests)
- No consolidated rite catalog in doctrine
- **Time to Answer**: FAILED (must explore `/roster/rites/` directly)

**Mental Model Score**: 2/10 (promised but undelivered)

---

#### Scenario F: "Can I run parallel Claude sessions?"

**Expected Path**: ??? (should be in `operations/` or mentioned in doctrine)

**Actual Experience**: ✗ **UNDOCUMENTED**
- Worktree system exists (11 commands, complete implementation)
- Compliance report says "not mentioned" in doctrine
- No worktree guide in `operations/`
- Feature is invisible to Future Self
- **Time to Answer**: FAILED (unknown unknown)

**Mental Model Score**: 0/10 (capability exists but architecturally invisible)

---

#### Scenario G: "What is the current compliance status?"

**Expected Path**: `compliance/COMPLIANCE-STATUS.md`

**Actual Experience**: ✓ **EXCELLENT**
- File exists, comprehensive, well-structured
- Subdirectories (`status/`, `audits/`, `quality-gates/`, `certifications/`) are empty but don't impede findability
- Could imagine future expansion into subdirs, but flat file works for now
- **Time to Answer**: <1 minute

**Mental Model Score**: 9/10

---

### 2.2 Mental Model Summary

**Future Self asks**:
1. "Why?" → **EXCELLENT** (philosophy complete)
2. "What decisions?" → **GOOD** (ADRs symlinked, though architecture/ hollow)
3. "How?" → **FRAGMENTED** (operations guides exist, CLI ref missing)
4. "Where in source?" → **CONFUSING** (paths correct but contradictory doctrine)
5. "What's available?" → **BROKEN** (rites/, cli-reference/ empty despite implementation)

**Average Score**: 5.2/10

**Key Insight**: The mental model succeeds at the **conceptual level** (philosophy, principles) but fails at the **operational level** (CLI, rites, workflows). Future Self can understand "why Knossos exists" but struggles with "how do I actually use it."

---

## 3. Navigation Effectiveness

### 3.1 Entry Points Analysis

| Entry Point | Effectiveness | Evidence |
|-------------|--------------|----------|
| `DOCTRINE.md` | ✓ STRONG | Clear overview, navigation table, structure diagram |
| `reference/INDEX.md` | ✓ EXCELLENT | Reading paths, audience-based routing, comprehensive cross-refs |
| `philosophy/knossos-doctrine.md` | ✓ CANONICAL | 3,777 words, complete doctrine, well-sectioned |
| Directory browsing | ⚠ MISLEADING | Empty dirs create false expectations |
| Search for "CLI" | ✗ FRAGMENTED | No consolidated reference, scattered examples |
| Search for "rites" | ✗ BROKEN | INDEX lists rites, but `../rites/` empty |

**Entry Point Score**: 7/10
- +5 for DOCTRINE.md and INDEX.md (excellent intentional entry points)
- +2 for philosophy breadth
- -2 for directory browsing misleading experience
- -1 for search ineffectiveness on operational topics

### 3.2 Cross-Reference Integrity

**Working Cross-References** (✓):
- DOCTRINE.md → philosophy files (all valid)
- INDEX.md → philosophy files (all valid)
- INDEX.md → ADRs via ../../decisions/ (all valid)
- foundations/ symlinks → ../../decisions/ (all 3 resolve)
- operations/guides/ symlinks → ../../../guides/ (all 5 resolve)

**Broken Cross-References** (✗):
- INDEX.md line 122 → `../rites/` (empty directory)
- DOCTRINE.md lines 47-59 → architecture/, evolution/ subdirs (all empty)
- DOCTRINE.md lines 54-56 → operations/cli-reference/, workflows/ (both empty)
- mythology-concordance.md, design-principles.md → `/roster/hooks/` (does not exist—should be `/roster/internal/hook/`)

**Cross-Reference Integrity Score**: 6/10
- Symlinks: 100% success rate (8/8 valid)
- Internal markdown links: ~60% valid (philosophy interconnected, but empty dir refs broken)
- External path references: ~85% valid (most `/roster/*` paths correct, but hooks/ wrong)

### 3.3 Symlink Strategy Assessment

**Symlink Locations**:
1. `foundations/` → `../../decisions/` (ADRs)
2. `operations/guides/` → `../../../guides/` (operational guides)

**Effectiveness**: ✓ **EXCELLENT**

**Rationale**:
- **Avoids duplication**: Single source of truth in canonical locations
- **Creates unified view**: Doctrine can reference foundational ADRs without moving them
- **Maintains discoverability**: Both `docs/doctrine/foundations/` and `docs/decisions/` work
- **No confusion**: Symlinks clearly marked in `ls -la` output

**Anti-Pattern NOT Present**: No symlink cycles, no broken symlinks, no symlinks to gitignored paths.

**Recommendation**: **Expand symlink strategy** to:
- Symlink relevant specs from `docs/specs/` into `architecture/` (e.g., TLA+ FSM spec)
- Consider symlinking rite manifests from `/roster/rites/` into `rites/` (or create summaries)

**Symlink Strategy Score**: 9/10 (best practice, should be expanded)

---

## 4. Gap Identification (IA Perspective)

### 4.1 Content Type Gaps

| Content Type | Expected Location | Current Status | Impact |
|--------------|------------------|----------------|--------|
| **CLI Reference** | `operations/cli-reference/` | ✗ Missing (68 commands, 0 documented) | **CRITICAL** |
| **Rite Catalog** | `rites/` | ✗ Missing (12 rites, 0 documented) | **HIGH** |
| **Worktree Guide** | `operations/guides/` | ✗ Missing (11 commands, 0 documented) | **HIGH** |
| **Workflow Guides** | `operations/workflows/` | ✗ Missing (session lifecycle, handoffs) | **MEDIUM** |
| **TDD/Design Docs** | `architecture/core-components/` | ✗ Missing (Moirai, Ariadne, Orchestrator) | **MEDIUM** |
| **Pattern Catalog** | `architecture/patterns/` | ✗ Missing (execution modes, handoffs, rite invocation) | **MEDIUM** |
| **Hook Catalog** | `architecture/subsystems/` or `operations/` | ✗ Missing (5+ hooks, only conceptual coverage) | **MEDIUM** |
| **Artifact Registry Docs** | `architecture/subsystems/` | ✗ Missing (query API undocumented) | **LOW** |
| **TLA+ Verification Explanation** | `architecture/` or `compliance/` | ✗ Missing (spec exists, no explanation) | **LOW** |
| **Roadmap** | `evolution/roadmap/` | ✗ Missing (future plans unknown) | **LOW** |
| **Retrospectives** | `evolution/retrospectives/` | ✗ Missing (sprint reviews not documented) | **LOW** |

**Gap Severity Distribution**:
- **CRITICAL**: 1 gap (CLI reference—88% of surface area undocumented)
- **HIGH**: 2 gaps (rites, worktree—major capabilities invisible)
- **MEDIUM**: 5 gaps (workflows, architecture docs, patterns, hooks, subsystems)
- **LOW**: 4 gaps (roadmap, retrospectives, artifact registry, TLA+)

### 4.2 Structural Gaps

**Missing Navigational Aids**:
1. **No "Getting Started" path**: DOCTRINE.md explains philosophy but not "Your First Session"
2. **No troubleshooting section**: Common errors, failure modes, debugging workflows
3. **No FAQ**: Frequently asked questions about platform usage
4. **No migration guides**: How to upgrade, terminology changes, breaking changes (evolution/ empty)
5. **No examples section**: Concrete session examples, rite invocation patterns

**Missing Organizational Elements**:
1. **No content ownership**: Files lack CODEOWNERS or "maintained by" metadata
2. **No last-reviewed dates**: Unknown staleness (all dated 2026-01-08, but future reviews unclear)
3. **No status indicators**: Which docs are draft vs authoritative vs deprecated

### 4.3 IA-Level Recommendations

**R1: Collapse Empty Directory Structure** (Priority: **CRITICAL**)

**Issue**: Empty directories create false expectations and navigation dead ends.

**Action**:
```
DELETE (or mark as future expansion):
- docs/doctrine/compliance/status/
- docs/doctrine/compliance/audits/        # Move doctrine-audit-20260108.md here first
- docs/doctrine/compliance/quality-gates/
- docs/doctrine/compliance/certifications/
- docs/doctrine/architecture/core-components/
- docs/doctrine/architecture/subsystems/
- docs/doctrine/architecture/patterns/
- docs/doctrine/evolution/roadmap/
- docs/doctrine/evolution/experiments/
- docs/doctrine/evolution/retrospectives/
- docs/doctrine/operations/workflows/
```

**Rationale**: Structure should reflect what exists, not what we hope to build. Empty directories are navigation friction.

**Alternative**: Keep directories but add README.md in each:
```markdown
# [Directory Name] - Planned Expansion

**Status**: Not yet populated
**Planned Content**: [Description]
**Expected Timeline**: [TBD or date]

For now, see [alternative location] for related content.
```

**Recommendation**: **DELETE** empty level-3 directories, keep level-2 as expansion targets with README placeholders.

---

**R2: Reorganize for Two-Tier Navigation** (Priority: **HIGH**)

**Current Problem**: 11 top-level categories dilute focus; operational content fragmented.

**Proposed Structure**:
```
docs/doctrine/
├── DOCTRINE.md                    # Entry point
│
├── understanding/                 # The "Why" (conceptual)
│   ├── knossos-doctrine.md        # Moved from philosophy/
│   ├── design-principles.md       # Moved from philosophy/
│   ├── mythology-concordance.md   # Moved from philosophy/
│   ├── INDEX.md                   # Moved from reference/
│   └── GLOSSARY.md                # Moved from reference/
│
├── using/                         # The "How" (operational)
│   ├── getting-started.md         # NEW - first session walkthrough
│   ├── cli-reference/             # EXPAND - document 68 commands
│   │   ├── session-commands.md
│   │   ├── rite-commands.md
│   │   ├── worktree-commands.md
│   │   └── [...]
│   ├── rites/                     # POPULATE - catalog of 12 rites
│   │   ├── catalog.md
│   │   └── [rite-name].md (or symlinks to /roster/rites/)
│   ├── workflows/                 # CREATE - session lifecycle, handoffs
│   └── troubleshooting.md         # NEW - common errors
│
├── building/                      # The "What" (architecture for implementers)
│   ├── decisions/                 # Symlink to ../../decisions/ (ADRs)
│   ├── architecture/              # NEW - component TDDs, patterns
│   ├── formal-verification/       # NEW - TLA+ specs with explanations
│   └── implementation-guides/     # NEW - how to extend platform
│
├── compliance/                    # The "Status" (health and audits)
│   ├── COMPLIANCE-STATUS.md
│   ├── doctrine-audit-20260108.md # Moved from audits/
│   └── [future audits]
│
└── evolution/                     # The "Future" (roadmap and history)
    ├── roadmap.md                 # Single file for now
    └── retrospectives.md          # Single file for now
```

**Rationale**:
- **Audience-based navigation**: "Understanding" (all users), "Using" (operators), "Building" (implementers)
- **Reduced categories**: 11 → 5 top-level dirs
- **Flat where possible**: Single files beat empty subdirectories
- **Symlinks preserve paths**: `foundations/` becomes `building/decisions/` symlink

**Migration Impact**: Moderate (requires updating cross-references, but symlinks can ease transition).

---

**R3: Create CLI Reference via Auto-Generation** (Priority: **CRITICAL**)

**Issue**: 68 commands, 0 systematic documentation. Highest ROI gap.

**Action**:
1. Write script to extract `ari [command] --help` output
2. Generate markdown per command family:
   - `cli-reference/session-commands.md`
   - `cli-reference/rite-commands.md`
   - `cli-reference/worktree-commands.md`
   - etc.
3. Include examples from actual usage (extract from compliance doc)

**Format Template**:
```markdown
# `ari session create`

**Purpose**: Initialize a new work session

**Usage**:
ari session create <initiative> <complexity>

**Arguments**:
- `initiative`: Session title/goal (string)
- `complexity`: SIMPLE, MODERATE, COMPLEX, EPIC

**Examples**:
ari session create "Add dark mode toggle" MODERATE

**See Also**: `ari session status`, `ari session park`, `ari session wrap`
```

**Effort**: 4-8 hours (manual) OR 2 hours (automated extraction + template)

---

**R4: Populate Rite Catalog** (Priority: **HIGH**)

**Issue**: INDEX.md lists 11 rites, but `rites/` directory empty.

**Options**:
1. **Symlink approach**: Symlink `/roster/rites/[rite]/manifest.yaml` into `rites/`
2. **Summary approach**: Create `rites/catalog.md` with one-paragraph summaries per rite
3. **Hybrid approach**: catalog.md for overview + individual `rites/[rite-name].md` files

**Recommendation**: **Summary approach** (catalog.md)

**Rationale**:
- Rite manifests (YAML) are implementation details, not user-facing docs
- Future Self needs "What does this rite do?" not "What's in the manifest?"
- Single catalog.md easier to maintain than 12 separate docs

**Template**:
```markdown
# Rite Catalog

## 10x-dev
**Purpose**: Full development lifecycle (PRD → TDD → Code → QA → PR)
**Agents**: Orchestrator, Planner, Architect, Builder, QA Adversary
**When to use**: Implementing features from requirements to production
**Invocation**: `/task`, `/sprint`, or direct orchestrator invocation

## docs
**Purpose**: Documentation workflow (audit → architecture → writing → review)
**Agents**: Orchestrator, Doc Auditor, Information Architect, Tech Writer, Doc Reviewer
**When to use**: Creating or improving technical documentation
**Invocation**: `/docs` or Task(orchestrator, "document X")

[... 9 more rites ...]
```

**Effort**: 2-3 hours

---

**R5: Document Worktree System** (Priority: **HIGH**)

**Issue**: 11-command subsystem completely undocumented.

**Action**: Create `operations/guides/worktree-guide.md` (or `using/worktree.md` in reorganized structure)

**Content**:
- **Use case**: Parallel Claude sessions with filesystem isolation
- **Command reference**: 11 commands with examples
- **Workflow**: create → switch → list → cleanup
- **Integration**: How worktrees interact with rite invocation
- **Troubleshooting**: Common issues (stale worktrees, lock conflicts)

**Effort**: 2-3 hours (with SRE/engineer consult for verification)

---

**R6: Fix Invalid Path References** (Priority: **CRITICAL** for trust)

**Issue**: `/roster/hooks/` documented but doesn't exist.

**Action**:
Replace all references to `/roster/hooks/` with correct paths:
- **Go implementation**: `/roster/internal/hook/`
- **Rite-specific scripts**: `/roster/rites/[rite-name]/hooks/`

**Files to update**:
- `philosophy/design-principles.md` line 16
- `philosophy/mythology-concordance.md` line 298

**Effort**: 15 minutes

---

**R7: Resolve SOURCE/PROJECTION Contradiction** (Priority: **CRITICAL** for doctrine integrity)

**Issue**: ADR-0009 says "roster/.claude/ IS Knossos"; mythology-concordance.md says the opposite.

**Action** (choose one):
1. **Supersede ADR-0009**: Create ADR-0009-superseded.md with corrected identity
2. **Update mythology-concordance.md**: Align with ADR-0009 (if ADR is correct)
3. **Amend ADR-0009**: Add amendment section clarifying SOURCE = `/roster/`

**Recommendation**: **Supersede ADR-0009** (evidence supports mythology-concordance.md's interpretation)

**Rationale**: Codebase behavior confirms SOURCE = `/roster/`, PROJECTION = `.claude/`. ADR-0009's statement was incorrect and should be formally superseded, not quietly contradicted.

**Effort**: 1 hour (decision + ADR creation)

---

## 5. Priority Ranking

### Tier 1: Critical (Do First)

| Rec | Action | Impact | Effort | ROI |
|-----|--------|--------|--------|-----|
| R6 | Fix invalid path references | Restores trust in docs | 15 min | **EXTREME** |
| R7 | Resolve SOURCE/PROJECTION contradiction | Clarifies platform identity | 1 hour | **EXTREME** |
| R1 | Collapse empty directory structure | Eliminates false expectations | 1 hour | **HIGH** |
| R3 | Create CLI reference | Documents 88% of surface area | 2-8 hours | **HIGH** |

**Rationale**: Trust repair (R6, R7) must precede content expansion. Structural cleanup (R1) prevents compounding navigation debt. CLI reference (R3) is highest-value missing content.

### Tier 2: High Value (Do Soon)

| Rec | Action | Impact | Effort | ROI |
|-----|--------|--------|--------|-----|
| R4 | Populate rite catalog | Makes 12 rites discoverable | 2-3 hours | **MEDIUM** |
| R5 | Document worktree system | Reveals major hidden capability | 2-3 hours | **MEDIUM** |
| R2 | Reorganize for two-tier navigation | Improves overall findability | 4-6 hours | **MEDIUM** |

**Rationale**: Rite catalog and worktree docs fill critical content gaps. Navigation reorganization (R2) is lower priority—current structure workable after empty dir cleanup.

### Tier 3: Quality Improvements (Do Later)

- Add troubleshooting section
- Create getting-started guide
- Document workflows (session lifecycle, handoffs)
- Populate architecture/ with TDDs
- Explain TLA+ formal verification
- Add FAQ section
- Create migration guides (evolution/)

**Rationale**: These improve UX but don't block current usage. Defer until Tier 1-2 complete.

---

## 6. Navigation Design Specification

### 6.1 Reading Paths (Audience-Based)

**Path 1: New Contributor** (First-Time Visitor)
```
Entry → DOCTRINE.md (2 min read)
  ↓
Deep Dive → understanding/knossos-doctrine.md (15 min read)
  ↓
Practical → using/getting-started.md (10 min tutorial)
  ↓
Reference → using/cli-reference/ (as needed)
```

**Path 2: Operator** (Using Knossos Daily)
```
Entry → using/ directory (browse by task)
  ↓
Need CLI → using/cli-reference/[command-family].md
  ↓
Need Rite → using/rites/catalog.md
  ↓
Stuck → using/troubleshooting.md
```

**Path 3: Implementer** (Extending Platform)
```
Entry → understanding/design-principles.md (design DNA)
  ↓
Decisions → building/decisions/ (ADR history)
  ↓
Architecture → building/architecture/ (component TDDs)
  ↓
Source → mythology-concordance.md (SOURCE locations)
```

**Path 4: Auditor** (Checking Compliance)
```
Entry → compliance/COMPLIANCE-STATUS.md (current state)
  ↓
Evidence → compliance/audits/ (validation reports)
  ↓
Doctrine → understanding/knossos-doctrine.md (expected state)
  ↓
Gap Analysis → compliance/COMPLIANCE-STATUS.md Section II (drift)
```

### 6.2 Cross-Reference Strategy

**Hub Documents** (every doc should link to at least one):
1. `DOCTRINE.md` - primary entry point
2. `understanding/INDEX.md` - navigation hub
3. `using/cli-reference/index.md` - operational hub (NEW)
4. `compliance/COMPLIANCE-STATUS.md` - health dashboard

**Cross-Reference Principles**:
1. **Bidirectional**: If A links to B, B should acknowledge A (via "See Also" section)
2. **Contextual**: Link where reader might ask "what's that?" not exhaustively
3. **Shallow linking**: Link to stable hub docs, not deep-linked sections (sections may move)
4. **External clarity**: Links to `/roster/` paths should be absolute, not relative

**See Also Template** (add to end of each doc):
```markdown
---

## See Also

- [Understanding/Knossos Doctrine](../understanding/knossos-doctrine.md) - Philosophical foundation
- [Using/CLI Reference](../using/cli-reference/session-commands.md) - Session management commands
- [Compliance Status](../compliance/COMPLIANCE-STATUS.md) - Current implementation status
```

### 6.3 Metadata Schema (Frontmatter)

Recommend adding YAML frontmatter to all docs:

```yaml
---
title: Knossos Doctrine
audience: [all, contributors, architects]
status: authoritative  # draft | authoritative | deprecated
last_reviewed: 2026-01-08
maintained_by: Architect Team
related:
  - philosophy/design-principles.md
  - reference/GLOSSARY.md
  - compliance/COMPLIANCE-STATUS.md
---
```

**Benefits**:
- Programmatic staleness detection (`last_reviewed` > 90 days → warning)
- Audience filtering (generate role-specific views)
- Status clarity (distinguish draft from canonical)
- Automated cross-reference validation

---

## 7. Structural Change Recommendations

### 7.1 What to Add

**Immediate (Tier 1)**:
1. `using/cli-reference/` - populate with 68 commands
2. `compliance/audits/` - move existing audit, prepare for future
3. Corrected path references (fix `/roster/hooks/`)
4. ADR-0009 supersession (resolve SOURCE/PROJECTION)

**Soon (Tier 2)**:
5. `using/rites/catalog.md` - document 12 rites
6. `using/worktree-guide.md` - document 11 worktree commands
7. `using/getting-started.md` - first session tutorial
8. `using/troubleshooting.md` - common errors and fixes

**Later (Tier 3)**:
9. `building/architecture/` - component TDDs, patterns
10. `building/formal-verification/` - TLA+ explanation
11. `using/workflows/` - session lifecycle, handoff diagrams
12. `evolution/roadmap.md` - future plans
13. `evolution/retrospectives.md` - sprint reviews

### 7.2 What to Remove

**Delete Immediately**:
- `compliance/status/` (empty)
- `compliance/quality-gates/` (empty)
- `compliance/certifications/` (empty)
- `architecture/core-components/` (empty)
- `architecture/subsystems/` (empty)
- `architecture/patterns/` (empty)
- `evolution/roadmap/` (empty)
- `evolution/experiments/` (empty)
- `evolution/retrospectives/` (empty)
- `operations/workflows/` (empty)

**Rationale**: Empty directories create navigation dead ends. Reinstate when content exists.

**Preserve**:
- `compliance/audits/` - move existing audit here, keep dir
- `operations/cli-reference/` - keep dir, populate immediately
- `rites/` - keep dir, populate soon

### 7.3 What to Reorganize

**Option A: Minimal Reorganization** (Lower Risk)
1. Collapse empty level-3 directories
2. Move audit into `compliance/audits/`
3. Add content to `cli-reference/`, `rites/`
4. Fix path references
5. Resolve SOURCE/PROJECTION

**Option B: Two-Tier Reorganization** (Higher Value, Higher Risk)
1. Rename `philosophy/` → `understanding/`
2. Move `reference/INDEX.md` and `GLOSSARY.md` into `understanding/`
3. Rename `operations/` → `using/`
4. Create `building/` for implementation docs
5. Collapse `compliance/`, `evolution/` to single-file or minimal structure
6. Update all cross-references

**Recommendation**: **Start with Option A**, consider Option B after Tier 1-2 content complete.

**Rationale**: Content gaps are more urgent than structural elegance. Fix findability first, optimize navigation second.

---

## 8. Findability Validation Criteria

### 8.1 The 30-Second Test

For each question below, Future Self should find an answer in under 30 seconds:

| Question | Target Location | Currently Passes? | After Recommendations? |
|----------|----------------|-------------------|----------------------|
| What is Knossos? | `DOCTRINE.md` | ✓ YES | ✓ YES |
| Why this architecture? | `philosophy/knossos-doctrine.md` | ✓ YES | ✓ YES |
| How do I create a session? | `using/cli-reference/session.md` | ✗ NO (missing) | ✓ YES (after R3) |
| What rites exist? | `using/rites/catalog.md` | ✗ NO (empty dir) | ✓ YES (after R4) |
| Where is Moirai implemented? | `understanding/mythology-concordance.md` | ⚠ CONFUSING (contradiction) | ✓ YES (after R7) |
| What commands are available? | `using/cli-reference/` | ✗ NO (missing) | ✓ YES (after R3) |
| Current compliance status? | `compliance/COMPLIANCE-STATUS.md` | ✓ YES | ✓ YES |
| Can I run parallel sessions? | `using/worktree-guide.md` | ✗ NO (missing) | ✓ YES (after R5) |

**Current Pass Rate**: 3/8 (37.5%)
**After Tier 1-2**: 8/8 (100%)

### 8.2 Breadcrumb Clarity

Every document should answer:
1. **Where am I?** (breadcrumb path at top)
2. **What is this?** (clear title + one-sentence purpose)
3. **Who is this for?** (audience statement)
4. **What's related?** (See Also section)

**Current Status**: Only `INDEX.md` consistently provides this context.

**Recommendation**: Add standard header template:
```markdown
# [Document Title]

> [One-sentence purpose]

**Audience**: [Primary audience] | **Related**: [Key related doc 1], [Key related doc 2]

---
```

### 8.3 Search Effectiveness

Test queries Future Self would use:

| Query | Expected Result | Currently Finds? | Recommendation |
|-------|----------------|------------------|----------------|
| "session lifecycle" | `philosophy/knossos-doctrine.md` Section V | ✓ YES | No change |
| "CLI commands" | `using/cli-reference/` | ✗ NO (dir empty) | Populate (R3) |
| "parallel sessions" | `using/worktree-guide.md` | ✗ NO (doesn't exist) | Create (R5) |
| "rite catalog" | `using/rites/catalog.md` | ✗ NO (dir empty) | Create (R4) |
| "SOURCE vs PROJECTION" | `understanding/mythology-concordance.md` | ⚠ CONTRADICTORY | Fix (R7) |

**Search relies on**:
1. **Content existence** (can't find what's not documented)
2. **Consistent terminology** (GLOSSARY helps)
3. **Heading structure** (well-structured docs surface in search)

**Recommendation**: After content gaps filled, add search optimization:
- Keywords in first paragraph of each doc
- Heading structure follows question patterns ("What is X?", "How do I Y?")
- Glossary terms appear in context

---

## 9. Content Briefs (Gap Filling)

### Brief 1: CLI Reference Documentation

**Location**: `docs/doctrine/operations/cli-reference/` (or `using/cli-reference/` post-reorg)

**Audience**: Operators, practitioners, anyone using `ari` CLI

**Purpose**: Comprehensive reference for all 68 `ari` commands

**Scope**:
- Cover all 15 command families (session, rite, worktree, hook, handoff, inscription, manifest, artifact, sync, validate, sails, naxos, tribute, etc.)
- Each command: syntax, arguments, examples, related commands
- Organize by family (one file per family or index + individual pages)

**Format**:
```
cli-reference/
├── index.md              # Overview, command family list
├── session.md            # session {create|status|park|resume|wrap|...}
├── rite.md               # rite {invoke|swap|release|pantheon|...}
├── worktree.md           # worktree {create|switch|list|cleanup|...}
├── hook.md               # hook {clew|context|...}
├── handoff.md            # handoff {prepare|execute|status|history}
├── inscription.md        # inscription {generate|validate|...}
├── artifact.md           # artifact {query|...}
├── sync.md               # sync {materialize|...}
└── [...]
```

**Priority**: CRITICAL
**Effort**: 4-8 hours (manual) OR 2 hours (scripted extraction from --help)
**Blockers**: None (all commands implemented and operational)

---

### Brief 2: Rite Catalog

**Location**: `docs/doctrine/rites/catalog.md` (or `using/rites/catalog.md`)

**Audience**: Practitioners selecting rites, architects understanding workflows

**Purpose**: Comprehensive catalog of all available rites with purpose, agents, and invocation patterns

**Scope**:
- All 12 rites from `/roster/rites/` (10x-dev, docs, forge, hygiene, debt-triage, security, sre, intelligence, rnd, strategy, ecosystem, [any others])
- Per rite: purpose, agents, when to use, invocation patterns, key workflows
- Comparison table (quick reference for rite selection)

**Format**:
```markdown
# Rite Catalog

## Overview
[One paragraph on what rites are and how they work]

## Quick Reference

| Rite | Purpose | Primary Use Case |
|------|---------|------------------|
| 10x-dev | Full dev lifecycle | Feature implementation |
| docs | Documentation workflow | Technical writing |
| [...]

## Rites

### 10x-dev
**Purpose**: [One-sentence description]
**Agents**: [List]
**When to use**: [Scenarios]
**Invocation**: [Command patterns]
**Key Workflows**: [Phases]

[Repeat for all 12 rites]
```

**Priority**: HIGH
**Effort**: 2-3 hours
**Blockers**: None (all rites exist with manifests in `/roster/rites/`)

---

### Brief 3: Worktree Guide

**Location**: `docs/doctrine/operations/guides/worktree-guide.md` (or `using/worktree.md`)

**Audience**: Operators running parallel Claude sessions, SREs managing platform

**Purpose**: Complete guide to worktree system for parallel session isolation

**Scope**:
- **Concept**: What worktrees are, why they exist, filesystem isolation model
- **Use Cases**: Parallel sessions, different rites per terminal, sprint isolation
- **Commands**: All 11 worktree commands with examples
- **Workflows**: Create → switch → work → cleanup lifecycle
- **Integration**: How worktrees interact with rite invocation, session management
- **Troubleshooting**: Stale worktrees, lock conflicts, cleanup procedures

**Format**:
```markdown
# Worktree Guide

## Overview
[What worktrees are, why they exist]

## Use Cases
- Parallel Claude sessions
- Different rites per terminal
- Sprint isolation

## Workflow

### Creating a Worktree
[Step-by-step with examples]

### Switching Between Worktrees
[Commands and state management]

### Cleanup
[Removing stale worktrees]

## Command Reference
[All 11 commands with syntax and examples]

## Troubleshooting
[Common issues and solutions]
```

**Priority**: HIGH
**Effort**: 2-3 hours (with engineer consult for verification)
**Blockers**: None (worktree system complete and operational)

---

### Brief 4: Getting Started Guide

**Location**: `docs/doctrine/operations/getting-started.md` (or `using/getting-started.md`)

**Audience**: New users, first-time Knossos operators

**Purpose**: Tutorial walkthrough of creating and completing a first session

**Scope**:
- Prerequisites (installation, environment setup)
- "Your First Session" tutorial (end-to-end example)
- Key concepts (clew, rites, agents, White Sails)
- Common patterns (task invocation, handoffs)
- Next steps (where to go from here)

**Format**:
```markdown
# Getting Started with Knossos

## Prerequisites
[What you need before starting]

## Your First Session

### Step 1: Create a Session
ari session create "My first task" SIMPLE
[Explanation of what happened]

### Step 2: Invoke a Rite
[Example rite invocation]

### Step 3: Complete the Work
[Task execution example]

### Step 4: Wrap the Session
ari session wrap
[What gets generated: tribute, White Sails]

## Key Concepts
[Brief explanations with links to deeper docs]

## Common Patterns
[Typical workflows]

## Next Steps
[Where to learn more]
```

**Priority**: MEDIUM
**Effort**: 3-4 hours
**Blockers**: None

---

### Brief 5: Troubleshooting Guide

**Location**: `docs/doctrine/operations/troubleshooting.md` (or `using/troubleshooting.md`)

**Audience**: All users encountering errors or unexpected behavior

**Purpose**: Diagnostic guide for common issues and failure modes

**Scope**:
- Session state issues (orphaned sessions, lock conflicts)
- Hook failures (event recording errors, context degradation)
- Rite invocation failures (missing manifests, agent errors)
- White Sails degradation (what causes GRAY/BLACK signals)
- Worktree issues (stale worktrees, switch failures)
- Recovery procedures (how to repair broken state)

**Format**:
```markdown
# Troubleshooting Guide

## Session Issues

### Orphaned Session
**Symptom**: [Description]
**Cause**: [Why it happens]
**Solution**: [How to fix]

### Lock Conflict
[Same structure]

## Hook Issues

### Event Recording Failure
[Same structure]

## Rite Invocation Issues

### Missing Manifest
[Same structure]

## Recovery Procedures

### Manual Session Cleanup
[Step-by-step]
```

**Priority**: MEDIUM
**Effort**: 2-3 hours (requires cataloging known failure modes)
**Blockers**: Input from engineers on common error patterns

---

## 10. Migration Plan

### 10.1 Immediate Actions (Week 1)

**Phase 1A: Trust Repair** (2 hours)
1. Fix invalid path references (R6: 15 min)
   - Edit `philosophy/design-principles.md` line 16
   - Edit `philosophy/mythology-concordance.md` line 298
   - Replace `/roster/hooks/` with correct paths
2. Resolve SOURCE/PROJECTION contradiction (R7: 1 hour)
   - Create `docs/decisions/ADR-0009-superseded.md`
   - Update references to clarify SOURCE = `/roster/`
3. Move audit to proper location (15 min)
   - `mv docs/doctrine/audits/doctrine-audit-20260108.md docs/doctrine/compliance/audits/`
   - Update cross-references

**Phase 1B: Structural Cleanup** (1 hour)
4. Delete empty level-3 directories (R1: 30 min)
   - Remove compliance/status/, quality-gates/, certifications/
   - Remove architecture/core-components/, subsystems/, patterns/
   - Remove evolution/roadmap/, experiments/, retrospectives/
   - Remove operations/workflows/
5. Add README placeholders to preserved empty dirs (30 min)
   - `operations/cli-reference/README.md` - "In progress, see R3"
   - `rites/README.md` - "In progress, see R4"

**Deliverables**:
- ✓ Path references correct
- ✓ SOURCE/PROJECTION doctrine unified
- ✓ Empty directories collapsed
- ✓ Navigation expectations aligned with reality

---

### 10.2 High-Priority Content (Weeks 2-3)

**Phase 2A: CLI Reference** (R3: 4-8 hours)
1. Write extraction script (`scripts/generate-cli-docs.sh`)
2. Generate markdown from `ari [cmd] --help` output
3. Organize by command family (15 files)
4. Add examples from compliance doc and actual usage
5. Create `cli-reference/index.md` overview

**Phase 2B: Rite Catalog** (R4: 2-3 hours)
1. Extract rite metadata from `/roster/rites/*/manifest.yaml`
2. Create `rites/catalog.md` with:
   - Quick reference table
   - Per-rite sections (purpose, agents, invocation)
3. Link from INDEX.md and DOCTRINE.md

**Phase 2C: Worktree Guide** (R5: 2-3 hours)
1. Interview engineer/SRE on worktree design
2. Write guide covering concept, commands, workflows
3. Add to `operations/guides/` (symlink or direct)
4. Link from CLI reference and INDEX

**Deliverables**:
- ✓ 68 commands documented
- ✓ 12 rites cataloged
- ✓ Worktree system documented
- ✓ Findability increased from 37.5% → ~80%

---

### 10.3 Quality Improvements (Weeks 4-6)

**Phase 3A: Operational Guides** (6-8 hours)
1. Create getting-started.md (3-4 hours)
2. Create troubleshooting.md (2-3 hours)
3. Document workflows (session lifecycle, handoffs) (2-3 hours)

**Phase 3B: Architecture Docs** (8-10 hours)
1. Populate `architecture/` with component TDDs
2. Document patterns (execution modes, handoffs)
3. Explain TLA+ formal verification
4. Document artifact registry

**Phase 3C: Evolution Content** (2-3 hours)
1. Create `evolution/roadmap.md` (future plans)
2. Create `evolution/retrospectives.md` (sprint reviews)

**Deliverables**:
- ✓ Complete operational documentation
- ✓ Architecture documentation for implementers
- ✓ Evolution tracking for stakeholders

---

### 10.4 Optional: Navigation Reorganization (Week 7)

**Phase 4: Two-Tier Reorganization** (R2: 4-6 hours)

Only proceed if:
- [ ] Phases 1-3 complete
- [ ] User feedback indicates navigation still confusing
- [ ] Team consensus on new structure

**Actions**:
1. Rename directories (philosophy → understanding, operations → using)
2. Create building/ for implementation docs
3. Consolidate reference/ into understanding/
4. Update all cross-references
5. Test all symlinks
6. Update DOCTRINE.md structure diagram

**Deliverables**:
- ✓ Audience-based navigation (understanding, using, building)
- ✓ Reduced top-level categories (11 → 5)

---

## 11. Conclusion

### 11.1 The Diagnosis

**Current State**: Philosophically world-class, operationally incomplete.

The `docs/doctrine/` structure demonstrates **architectural vision** (excellent taxonomy, thoughtful categorization, mythological coherence) but suffers from **premature scaffolding**—directories created before content, promises made before delivery.

**The Core Problem**: Future Self experiences **expectation whiplash**:
1. Seduced by elegant philosophy (knossos-doctrine.md, design-principles.md)
2. Impressed by comprehensive structure (DOCTRINE.md directory diagram)
3. Frustrated by hollow promises (empty architecture/, rites/, cli-reference/)
4. Confused by contradictions (SOURCE/PROJECTION identity crisis)
5. Forced to manual exploration (68 undocumented commands, 12 invisible rites)

**Result**: Trust degrades. The doctrine is philosophically sound but operationally unreliable.

### 11.2 The Prescription

**Two-Phase Strategy**:

**Phase 1: Consolidation** (Immediate)
- Collapse empty scaffolding (remove false promises)
- Fix path inaccuracies (restore technical trust)
- Resolve doctrinal contradictions (unify identity claim)
- Document operational reality (CLI, rites, worktrees)

**Phase 2: Expansion** (Future)
- Grow structure organically as content emerges
- Add architecture docs when TDDs written
- Add evolution content when roadmap clarified
- Reorganize navigation only if findability still suffers

**Principle**: **Structure follows content, not the reverse.**

### 11.3 Success Criteria

**6 months from now**, Future Self should be able to:

✓ Find the answer to "What is Knossos?" in under 1 minute (philosophy)
✓ Find the answer to "How do I [use CLI command]?" in under 30 seconds (operations)
✓ Find the answer to "Where is [component] implemented?" in under 30 seconds (architecture)
✓ Trust that documented paths are accurate (no broken references)
✓ Discover capabilities without needing source code exploration (CLI, rites, worktrees)
✓ Navigate from concept → implementation → usage without dead ends

**Quantitative Target**: 30-second findability test pass rate of **100%** (currently 37.5%).

### 11.4 The Acid Test

*Can Future Self, returning after 6 months with no context, navigate this documentation to accomplish a task without frustration?*

**Current Answer**: **No**—philosophy is findable, operations are fragmented, promises are broken.

**After Tier 1-2 Recommendations**: **Yes**—philosophy intact, operations documented, expectations aligned.

---

## 12. Final Recommendations Summary

### Critical (Do Immediately)

1. **R6**: Fix invalid path references (`/roster/hooks/` → correct paths) - 15 min
2. **R7**: Resolve SOURCE/PROJECTION contradiction (supersede ADR-0009) - 1 hour
3. **R1**: Collapse empty directory structure - 1 hour
4. **R3**: Create CLI reference documentation - 2-8 hours

**Total Effort**: 4-10 hours
**Impact**: Restores trust, documents 88% of surface area

### High Value (Do Soon)

5. **R4**: Populate rite catalog - 2-3 hours
6. **R5**: Document worktree system - 2-3 hours
7. **R2** (Optional): Reorganize for two-tier navigation - 4-6 hours

**Total Effort**: 8-12 hours
**Impact**: Completes operational documentation, improves findability

### Quality (Do Later)

8. Add getting-started guide
9. Add troubleshooting guide
10. Document workflows
11. Populate architecture/ with TDDs
12. Explain TLA+ verification
13. Add FAQ section
14. Create migration guides

**Total Effort**: 16-24 hours
**Impact**: Polishes UX, aids implementers

---

## Assessment Metadata

**Architect**: Information Architect Agent
**Date**: 2026-01-08
**Sprint**: Knossos Doctrine Documentation v2
**Task**: 002 - Information Architecture Assessment
**Methodology**: Structure analysis + mental model simulation + gap identification + navigation design
**Prerequisites**: Doctrine Audit (Task 001) - doctrine-audit-20260108.md
**Confidence Level**: HIGH
**Evidence Quality**: Direct directory traversal + file reading + cross-reference verification

---

*The labyrinth's architecture is sound. Now fill the rooms with navigable truth.*

*May the clew guide not just Theseus, but Future Self returning home.*
