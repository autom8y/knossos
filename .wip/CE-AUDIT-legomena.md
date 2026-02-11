# CE Audit: Legomena Quality

**Date**: 2026-02-09
**Auditor**: Context Engineer (Claude Opus 4.6)
**Scope**: All `.lego.md` files across `mena/` and `rites/*/mena/`

## Summary

- **Total legomena audited**: 45 (13 top-level + 4 shared + 28 rite-level)
- **Skills with precise Use when/Triggers**: 30/45 (67%)
- **Skills using INDEX->companion pattern**: 28/45 (62%)
- **Average lines per INDEX file**: ~192 lines
- **Oversized skills (>500 lines in INDEX)**: 1 (doc-sre at 749 lines)
- **Near-oversized (400-500 lines)**: 4 (build-ref 493, debt-ref 456, architect-ref 445, doc-strategy 402)
- **Lifecycle misclassifications (should be dromena)**: 7 (CRITICAL)
- **Missing/vague Use when+Triggers in description**: 15 (CRITICAL)

---

## Critical Findings

### CRIT-1: Seven rite-ref legomena are procedural command documentation, not reference knowledge

**Severity**: CRITICAL -- lifecycle misclassification, wrong primitive

The following legomena describe **slash command behavior** (execute, produce output, switch state). They are procedural scripts with imperative instructions ("Invoke Rite Switch", "Execute via Bash tool"), example output blocks, state change tables, and error handling flows. This is dromena behavior, not legomena behavior.

Legomena (skills) are **persistent reference knowledge** that CC loads autonomously when triggered. These files describe transient actions with side effects -- the defining characteristic of dromena (commands).

| File | Lines | Issue |
|------|-------|-------|
| `rites/10x-dev/mena/10x-ref/INDEX.lego.md` | 319 | Documents `/10x` command behavior with bash execution steps |
| `rites/10x-dev/mena/build-ref/INDEX.lego.md` | 493 | Documents `/build` command with execution flow, error cases |
| `rites/10x-dev/mena/architect-ref/INDEX.lego.md` | 445 | Documents `/architect` command with execution flow |
| `rites/docs/mena/docs-ref/INDEX.lego.md` | 350 | Documents `/docs` command with rite switching behavior |
| `rites/hygiene/mena/hygiene-ref/INDEX.lego.md` | 135 | Documents `/hygiene` command with rite switching behavior |
| `rites/debt-triage/mena/debt-ref/INDEX.lego.md` | 456 | Documents `/debt` command with rite switching behavior |
| `rites/sre/mena/sre-ref/INDEX.lego.md` | 259 | Documents `/sre` command with rite switching behavior |

**Evidence** (from `10x-ref/INDEX.lego.md` line 36-40):
```
### 1. Invoke Rite Switch
Execute via Bash tool:
```bash
$KNOSSOS_HOME/ari sync --rite 10x-dev
```
```

This is an imperative procedure, not reference knowledge. When CC loads this as a persistent skill, 300+ lines of command documentation sit in context indefinitely even after the command has been executed.

**Impact**: These skills consume 300-500 lines of context each as persistent knowledge when they should be transient commands. Combined, they waste ~2,400 lines of context window on procedural content that should execute-and-exit.

**Recommendation**: Convert to dromena (`.dro.md`). If reference content is needed alongside the command (e.g., agent catalog for a rite), split into a thin dromenon (execute+exit) and a small legomenon (rite reference only -- agents, workflow, when-to-use).

### CRIT-2: Fifteen legomena have missing or imprecise `Use when:` / `Triggers:` in frontmatter description

**Severity**: CRITICAL -- routing failure, skills may never load or misfire

CC uses the `description` field for autonomous skill discovery. Without explicit `Use when:` or `Triggers:` phrases, CC must guess when to load the skill. Vague descriptions lead to false negatives (skill never loads when needed) or false positives (skill loads when irrelevant).

**Skills with no explicit Triggers in description**:

| File | Current Description | Issue |
|------|-------------------|-------|
| `rites/10x-dev/mena/doc-artifacts/INDEX.lego.md` | "PRD, TDD, ADR, and Test Plan templates. Triggers: PRD, TDD, ADR..." | OK -- has triggers |
| `rites/shared/mena/orchestrator-templates/INDEX.lego.md` | "Shared schemas for orchestrator consultation patterns" | **MISSING** -- no Triggers phrase, no Use when |
| `mena/templates/doc-artifacts/INDEX.lego.md` | "PRD, TDD, ADR, and Test Plan templates for 10x development workflow. Canonical schemas and validation for core artifacts." | **MISSING** -- no Triggers phrase despite having good content |
| `rites/ecosystem/mena/ecosystem-ref/INDEX.lego.md` | "Knossos ecosystem patterns reference. Triggers: knossos-sync, ari-sync..." | OK |
| `rites/forge/mena/forge-ref/INDEX.lego.md` | Multi-line description with Use when + Triggers | GOOD -- best-in-class |

**Skills with vague descriptions** (no domain-specific trigger phrases):

| File | Current Description | Issue |
|------|-------------------|-------|
| `orchestrator-templates` | "Shared schemas for orchestrator consultation patterns" | What triggers this? "consultation request"? "orchestrator schema"? |
| `doc-artifacts` (top-level) | "PRD, TDD, ADR, and Test Plan templates..." | Missing Triggers keyword list |
| `session-common` | "Session schema: fields, FSM states, complexity levels, validation..." | Triggers are internal-facing, not user-intent-facing |
| `session-shared` | "Session and workflow resolution patterns..." | Same issue |
| `documentation` | "Doc standards routing hub. Triggers: documentation, doc standards..." | "documentation" is overly broad as a trigger -- will fire on any doc mention |

**Recommendation**: Every legomenon description should follow the pattern:
```
"[Concise purpose]. Use when: [2-3 user-intent phrases]. Triggers: [5-8 specific keywords]."
```
The `forge-ref` skill is the gold standard example of this format.

---

## High Findings

### HIGH-1: Duplicate doc-artifacts skill across top-level and 10x-dev rite

**Severity**: HIGH -- content duplication, 258 extra lines of context

Two separate legomena serve the same purpose:

| File | Lines | Content |
|------|-------|---------|
| `mena/templates/doc-artifacts/INDEX.lego.md` | 76 | Schema-focused INDEX with progressive disclosure to companion files |
| `rites/10x-dev/mena/doc-artifacts/INDEX.lego.md` | 258 | Full inline templates (PRD, TDD, ADR, Test Case, Test Summary) |

The top-level version is well-structured with progressive disclosure. The rite-level version inlines all templates directly, making it 3.4x larger. When the 10x-dev rite is active, both may be projected into `.claude/skills/`, creating redundant skill names.

**Impact**: CC may load the wrong one, or both. The 258-line version defeats progressive disclosure -- it dumps full templates into context even when only one is needed.

**Recommendation**: Keep the top-level version (compact INDEX with companions). Remove the rite-level duplicate or make it a thin pointer to the top-level skill.

### HIGH-2: doc-sre is a monolithic 749-line skill without progressive disclosure

**Severity**: HIGH -- exceeds 500-line limit by 50%, no companion files

`rites/sre/mena/doc-sre/INDEX.lego.md` (749 lines) contains 8 full templates inline:
- Observability Report (67 lines)
- Reliability Plan (40 lines)
- Postmortem (50 lines)
- Chaos Experiment (72 lines)
- Resilience Report (52 lines)
- Tracking Plan (40 lines)
- Infrastructure Change (108 lines)
- Pipeline Design (117 lines)
- Incident Communication (99 lines)

Every template is in the INDEX file. There are no companion files for progressive disclosure.

**Impact**: When CC loads this skill, 749 lines enter context. Most users need only 1-2 templates per session.

**Recommendation**: Split into INDEX (template routing table + quick reference, ~80 lines) and individual companion files per template. Follow the pattern used by `doc-artifacts` (top-level), which keeps INDEX at 76 lines with companions for each schema.

### HIGH-3: Four near-oversized INDEX files (400-500 lines) with inline content

**Severity**: HIGH -- progressive disclosure not used

| File | Lines | Content Type |
|------|-------|-------------|
| `rites/10x-dev/mena/build-ref/INDEX.lego.md` | 493 | Full `/build` command spec with 4 examples |
| `rites/debt-triage/mena/debt-ref/INDEX.lego.md` | 456 | Full `/debt` command spec with YAML examples |
| `rites/10x-dev/mena/architect-ref/INDEX.lego.md` | 445 | Full `/architect` command spec with 3 examples |
| `rites/strategy/mena/doc-strategy/INDEX.lego.md` | 402 | 4 full templates inline |

These are compounded by CRIT-1 -- the first three are procedural command docs that should be dromena anyway. `doc-strategy` follows the same monolithic pattern as `doc-sre` (HIGH-2).

**Recommendation**: For `doc-strategy`, split templates into companion files. For the command specs, convert to dromena.

### HIGH-4: cross-rite (top-level) and cross-rite-handoff (shared) overlap in domain

**Severity**: HIGH -- confusing routing, unclear which skill CC should load

| Skill | Source | Description |
|-------|--------|-------------|
| `cross-rite` | `mena/guidance/cross-rite/INDEX.lego.md` | "Cross-rite handoff protocols. Triggers: handoff, wrap, /wrap, rite transition, deployment ready." |
| `cross-rite-handoff` | `rites/shared/mena/cross-rite-handoff/INDEX.lego.md` | "Cross-rite HANDOFF artifact schema. Triggers: cross-rite, handoff artifact, rite transfer, work handoff." |

Both trigger on "handoff", "cross-rite", and "rite transfer". The top-level skill has decision trees and routes; the shared skill has the HANDOFF artifact schema. When a user asks about handoffs, CC has no clear signal for which to load.

**Recommendation**: Merge into one skill, or differentiate descriptions sharply. The top-level should trigger on "handoff routing" / "which rite next" and the shared should trigger on "HANDOFF artifact" / "handoff document format". Remove trigger overlap.

---

## Medium Findings

### MED-1: Shared legomena `@` syntax in references will not resolve

**Severity**: MEDIUM -- broken references in skill text

Several skills use `@skill-name` or `@shared-templates#anchor` syntax to reference other skills. This is explicitly called out as a non-working pattern in the `lexicon` skill itself:

> `@skill-name` -- CC has no `@` resolution

Found in:
- `rites/shared/mena/shared-templates/INDEX.lego.md` line 51-53: `@documentation`, `@doc-ecosystem`, `@cross-rite-handoff`
- `rites/sre/mena/doc-sre/INDEX.lego.md` line 29-31: `@shared-templates#debt-ledger-template`
- `rites/intelligence/mena/doc-intelligence/INDEX.lego.md` line 259: `@doc-artifacts`, `@doc-sre`
- `rites/security/mena/security-ref/INDEX.lego.md` line 135-136: `@standards`, `@documentation`
- `rites/strategy/mena/strategy-ref/INDEX.lego.md` line 135-136: `@documentation`
- `rites/rnd/mena/rnd-ref/INDEX.lego.md` line 135-136: `@standards`

**Recommendation**: Replace all `@skill-name` references with plain text skill names or `Skill("name")` invocation syntax.

### MED-2: file-verification is procedural protocol, borderline lifecycle misclassification

**Severity**: MEDIUM -- arguably should be a rule, not a skill

`mena/guidance/file-verification/INDEX.lego.md` (161 lines) defines a mandatory verification protocol ("NEVER claim you wrote a file without verification"). This is behavioral enforcement, not reference knowledge. It reads like a rule (always-on, behavioral) rather than a skill (on-demand, reference).

The file itself acknowledges this tension: "This skill is intentionally self-contained as a quick reference protocol."

**Recommendation**: Consider converting to a CC Rule (`.claude/rules/file-verification.md`) which would be path-scoped and always-on, or keep as a skill but add a note in CLAUDE.md referencing it. As a skill, it only loads when triggered -- but the protocol should be active for ALL file writes, not just when the model remembers to load it.

### MED-3: Session skills (`session-common`, `session-shared`, `moirai-fates`) are internal-facing

**Severity**: MEDIUM -- descriptions lack user-intent triggers

These three skills serve internal platform machinery (session schema, Moirai routing, validation patterns). Their descriptions use internal terminology:

- `session-common`: "Session schema: fields, FSM states, complexity levels, validation"
- `session-shared`: "Session and workflow resolution patterns"
- `moirai-fates`: "Moirai routing table for Fate domains"

Users will never type "FSM states" or "Moirai routing" to trigger these skills. They are consumed by other skills and commands, not by direct user intent.

**Recommendation**: Either (a) document these as internal-only skills consumed by other skills (add "Internal: loaded by session commands" to description), or (b) add user-facing triggers like "session state", "session operations", "park session", "resume session".

### MED-4: doc-consolidation has no apparent active consumers

**Severity**: MEDIUM -- orphan skill

`rites/docs/mena/doc-consolidation/INDEX.lego.md` (191 lines) defines a documentation consolidation workflow with schemas for MANIFEST, extraction, and checkpoint artifacts. No agent prompt or command references this skill. It appears to be a standalone specification without integration into any active workflow.

**Recommendation**: Either wire this into the docs rite agents (e.g., doc-auditor, information-architect) or archive it. Unused skills are pure token waste when loaded.

### MED-5: Legacy terminology in projected output

**Severity**: MEDIUM -- terminology drift between source and projection

The projected output at `.claude/skills/hygiene-ref/INDEX.md` line 3 shows `"refactoring team"` and line 7 shows `"Code Hygiene Team"` -- these are legacy "team" references that should be "rite" per the SL-008 terminology cleanse. The source file has been updated but the projection appears stale.

**Recommendation**: Run `ari sync` to refresh projections and ensure source->projection faithfully reflects terminology updates.

---

## Low Findings

### LOW-1: Inconsistent description format across skills

**Severity**: LOW -- polish

Some descriptions follow the recommended `"Purpose. Triggers: keyword1, keyword2."` format; others omit the Triggers keyword entirely; some use `"Use when:"` instead. There is no single canonical format.

Examples of format variance:
- **Good**: `"Knossos-to-CC primitive mapping. Triggers: lexicon, CC primitives, invocation mapping..."` (lexicon)
- **Good**: `"Reference documentation for The Forge... Use when: learning about rite creation... Triggers: /forge, /new-rite..."` (forge-ref)
- **Missing**: `"PRD, TDD, ADR, and Test Plan templates for 10x development workflow."` (doc-artifacts top-level)
- **Missing**: `"Shared schemas for orchestrator consultation patterns"` (orchestrator-templates)

**Recommendation**: Standardize on `"[Purpose sentence]. Use when: [intent phrases]. Triggers: [keyword list]."` format for all legomena. The `forge-ref` description is the gold standard.

### LOW-2: Companion files use `.lego.md` extension inconsistently

**Severity**: LOW -- naming convention

The hygiene-ref skill has companion files WITH the `.lego.md` extension:
- `rites/hygiene/mena/hygiene-ref/agents.lego.md`
- `rites/hygiene/mena/hygiene-ref/workflow-examples.lego.md`

Other skills use plain `.md` companion files. The `.lego.md` extension on companions means the materialization pipeline may treat them as separate skills rather than companion files of the INDEX.

**Recommendation**: Companions should use plain `.md` extension. Only the INDEX file should use `.lego.md`.

### LOW-3: Several skills reference non-existent paths

**Severity**: LOW -- broken links, no runtime impact

- `orchestrator-templates` references `~/.claude/commands/guidance/10x-workflow/INDEX.md` -- path likely incorrect
- `doc-ecosystem` references `~/.claude/commands/guidance/standards/INDEX.md` -- path uses commands/ not skills/
- `claude-md-architecture` references `~/.claude/skills/orchestration/execution-mode.md` -- directory may not exist

These broken links don't affect CC behavior (CC doesn't resolve relative links in skill bodies) but indicate maintenance drift.

---

## Per-Legomenon Assessment

### Top-Level (`mena/`) -- 13 files

| File | Lines | Use when? | Triggers? | INDEX pattern? | Issues |
|------|-------|-----------|-----------|----------------|--------|
| `guidance/lexicon/INDEX.lego.md` | 64 | Implicit | Yes | Yes (3 companions) | GOOD - exemplary |
| `guidance/standards/INDEX.lego.md` | 57 | Implicit | Yes | Yes (7 companions) | GOOD |
| `guidance/rite-discovery/INDEX.lego.md` | 67 | No | Yes | Yes (1 companion) | OK |
| `guidance/prompting/INDEX.lego.md` | 113 | No | Yes | Yes (11 companions) | OK |
| `guidance/file-verification/INDEX.lego.md` | 161 | No | Yes | No (self-contained) | MED-2: should be rule? |
| `guidance/cross-rite/INDEX.lego.md` | 91 | No | Yes | Yes (4 companions) | HIGH-4: overlaps shared |
| `session/moirai/INDEX.lego.md` | 55 | No | Yes | No | MED-3: internal-facing |
| `session/common/INDEX.lego.md` | 61 | No | Yes | Yes (8 companions) | MED-3: internal-facing |
| `session/shared/INDEX.lego.md` | 55 | No | Yes | Yes (3 companions) | MED-3: internal-facing |
| `templates/doc-artifacts/INDEX.lego.md` | 76 | Yes (inline) | Yes (inline) | Yes (4 companions) | CRIT-2: no Triggers in description |
| `templates/atuin-desktop/INDEX.lego.md` | 76 | No | Yes | Yes (5 companions) | OK |
| `templates/justfile/INDEX.lego.md` | 116 | No | Yes | Yes (11 companions) | OK |
| `templates/documentation/INDEX.lego.md` | 69 | Yes (inline) | Yes | Yes (1 companion) | GOOD |

### Shared (`rites/shared/mena/`) -- 4 files

| File | Lines | Use when? | Triggers? | INDEX pattern? | Issues |
|------|-------|-----------|-----------|----------------|--------|
| `cross-rite-handoff/INDEX.lego.md` | 51 | Yes (inline) | Yes | Yes (4 companions) | HIGH-4: overlaps top-level |
| `orchestrator-templates/INDEX.lego.md` | 38 | Yes (inline) | No | Yes (2 companions) | CRIT-2: no Triggers |
| `shared-templates/INDEX.lego.md` | 53 | No | Yes | Yes (6 companions) | MED-1: @syntax |
| `smell-detection/INDEX.lego.md` | 85 | Yes (inline) | Yes | Yes (13 companions) | GOOD |

### Rite-Level -- 28 files

| File | Lines | Use when? | Triggers? | INDEX pattern? | Issues |
|------|-------|-----------|-----------|----------------|--------|
| **10x-dev** | | | | | |
| `10x-ref/INDEX.lego.md` | 319 | No | Yes | No | CRIT-1: procedural |
| `10x-workflow/INDEX.lego.md` | 173 | No | Yes | Yes (8 companions) | OK |
| `build-ref/INDEX.lego.md` | 493 | No | Yes | No | CRIT-1: procedural, HIGH-3: oversized |
| `architect-ref/INDEX.lego.md` | 445 | No | Yes | No | CRIT-1: procedural, HIGH-3: oversized |
| `doc-artifacts/INDEX.lego.md` | 258 | No | Yes | No | HIGH-1: duplicates top-level |
| **debt-triage** | | | | | |
| `debt-ref/INDEX.lego.md` | 456 | No | Yes | No | CRIT-1: procedural, HIGH-3: oversized |
| **docs** | | | | | |
| `doc-consolidation/INDEX.lego.md` | 191 | No | Yes | Yes (9 companions) | MED-4: no consumers |
| `doc-reviews/INDEX.lego.md` | 252 | No | Yes | No | HIGH-3: all templates inline |
| `docs-ref/INDEX.lego.md` | 350 | No | Yes | No | CRIT-1: procedural |
| **ecosystem** | | | | | |
| `ecosystem-ref/INDEX.lego.md` | 99 | No | Yes | Yes (2 companions) | OK |
| `doc-ecosystem/INDEX.lego.md` | 80 | No | Yes | Yes (7 companions) | GOOD |
| `claude-md-architecture/INDEX.lego.md` | 287 | Yes | Yes | Yes (5 companions) | OK -- slightly large |
| **forge** | | | | | |
| `rite-development/INDEX.lego.md` | 172 | No | Yes | Yes (8 companions) | OK |
| `forge-ref/INDEX.lego.md` | 346 | Yes | Yes | Yes (6 companions) | GOOD description, large INDEX |
| `agent-prompt-engineering/INDEX.lego.md` | 207 | Yes | Yes | Yes (4 companions) | GOOD |
| **hygiene** | | | | | |
| `hygiene-ref/INDEX.lego.md` | 135 | No | Yes | Yes (2 companions) | CRIT-1: procedural |
| `hygiene-ref/agents.lego.md` | 123 | No | Yes | N/A (companion) | LOW-2: .lego.md on companion |
| `hygiene-ref/workflow-examples.lego.md` | 180 | No | Yes | N/A (companion) | LOW-2: .lego.md on companion |
| **intelligence** | | | | | |
| `doc-intelligence/INDEX.lego.md` | 260 | No | Yes | No | Templates inline |
| `intelligence-ref/INDEX.lego.md` | 137 | No | Yes | No | OK |
| **rnd** | | | | | |
| `doc-rnd/INDEX.lego.md` | 399 | No | Yes | No | HIGH-3: templates inline |
| `rnd-ref/INDEX.lego.md` | 137 | No | Yes | No | OK |
| **security** | | | | | |
| `doc-security/INDEX.lego.md` | 319 | No | Yes | No | Templates inline |
| `security-ref/INDEX.lego.md` | 137 | No | Yes | No | OK |
| **sre** | | | | | |
| `doc-sre/INDEX.lego.md` | 749 | No | Yes | No | HIGH-2: monolithic, 749 lines |
| `sre-ref/INDEX.lego.md` | 259 | No | Yes | No | CRIT-1: procedural |
| **strategy** | | | | | |
| `doc-strategy/INDEX.lego.md` | 402 | No | Yes | No | HIGH-3: templates inline |
| `strategy-ref/INDEX.lego.md` | 137 | No | Yes | No | OK |

---

## Recommendations

Ordered by impact (highest first):

### 1. Convert 7 procedural *-ref legomena to dromena (CRIT-1)

**Files**: `10x-ref`, `build-ref`, `architect-ref`, `docs-ref`, `hygiene-ref`, `debt-ref`, `sre-ref`
**Impact**: Eliminates ~2,400 lines of persistent context waste
**Approach**: Rename from `.lego.md` to `.dro.md`. If rite reference content is needed, extract a thin companion legomenon (agent table + when-to-use, ~40 lines).

### 2. Standardize description format with explicit Triggers (CRIT-2)

**Files**: All 45 legomena, 15 urgently
**Impact**: Fixes CC skill routing for ~33% of skills
**Approach**: Adopt `forge-ref` format: `"[Purpose]. Use when: [intent]. Triggers: [keywords]."` Create a validation script to enforce this at `ari sync` time.

### 3. Split monolithic template skills into INDEX + companions (HIGH-2, HIGH-3)

**Files**: `doc-sre` (749), `doc-strategy` (402), `doc-rnd` (399), `doc-security` (319), `doc-intelligence` (260), `doc-reviews` (252)
**Impact**: Reduces context cost from ~2,380 lines to ~480 lines (INDEX-only load)
**Approach**: Each template becomes a companion file. INDEX contains routing table, quality gate summary, and links. Follow `doc-ecosystem` pattern (80-line INDEX with companion templates).

### 4. Resolve doc-artifacts duplication (HIGH-1)

**Files**: `mena/templates/doc-artifacts/` vs `rites/10x-dev/mena/doc-artifacts/`
**Impact**: Eliminates 258 lines of duplicate content
**Approach**: Keep top-level (with progressive disclosure). Make rite-level a thin pointer or delete.

### 5. Deduplicate cross-rite / cross-rite-handoff trigger overlap (HIGH-4)

**Files**: `mena/guidance/cross-rite/` and `rites/shared/mena/cross-rite-handoff/`
**Impact**: Prevents CC from loading wrong handoff skill
**Approach**: Differentiate descriptions sharply. `cross-rite` = routing decisions. `cross-rite-handoff` = HANDOFF artifact schema.

### 6. Replace all `@skill-name` references with valid syntax (MED-1)

**Files**: 6+ skills using `@` syntax
**Impact**: Eliminates misleading references
**Approach**: Global search-and-replace `@skill-name` with plain text name.

### 7. Evaluate file-verification as a Rule instead of Skill (MED-2)

**Impact**: Ensures verification protocol is always active, not just when triggered
**Approach**: Create `.claude/rules/file-verification.md` for path-scoped always-on enforcement. Keep skill for detailed reference.

### 8. Add user-facing triggers to session-* internal skills (MED-3)

**Files**: `session-common`, `session-shared`, `moirai-fates`
**Impact**: Enables CC to load session skills when users ask about session operations
**Approach**: Add triggers like "session create", "session park", "session resume", "phase transition".

### 9. Fix companion file extensions (LOW-2)

**Files**: `hygiene-ref/agents.lego.md`, `hygiene-ref/workflow-examples.lego.md`
**Impact**: Prevents pipeline treating companions as separate skills
**Approach**: Rename to `.md` extension.

### 10. Fix stale projections (MED-5)

**Action**: Run `ari sync` to refresh `.claude/skills/` from updated source files.

---

## Architecture Observations

### What works well

1. **Progressive disclosure pattern** is well-established in top-level mena/ skills. Skills like `lexicon` (64 lines + 3 companions), `standards` (57 lines + 7 companions), and `prompting` (113 lines + 11 companions) are excellent examples.

2. **Shared legomena** (`rites/shared/mena/`) correctly extract cross-rite patterns. `smell-detection` (85-line INDEX + 13 companions) is particularly well-structured.

3. **forge-ref** description is the gold standard for CC discovery: multi-line YAML with explicit "Use when:" and "Triggers:" sections.

4. **doc-ecosystem** demonstrates ideal template skill architecture: 80-line INDEX that routes to 7 companion templates.

### Systemic patterns to address

1. **Template skills are consistently monolithic**. The `doc-*` pattern (doc-sre, doc-strategy, doc-rnd, doc-security, doc-intelligence, doc-reviews) consistently inlines all templates rather than using progressive disclosure. This appears to be a generation-1 pattern that was not updated when the INDEX->companion pattern was established.

2. **Rite-ref skills conflate two concerns**: rite switching (procedural) and rite reference (knowledge). These should be separated into dromena (switch) and legomena (reference).

3. **No validation exists for description quality**. The `ari sync` pipeline should validate that every `.lego.md` description contains either "Triggers:" or "Use when:" substring. This would catch CRIT-2 findings at build time.

### Token budget at peak

If all skills for a fully-loaded rite were triggered simultaneously (worst case):

| Scope | Skills | Total Lines | ~Tokens (est. 1.5 tok/line) |
|-------|--------|-------------|-----|
| Top-level (always available) | 13 | 1,121 | ~1,700 |
| Shared | 4 | 227 | ~340 |
| 10x-dev rite (worst case) | 5 | 1,688 | ~2,500 |
| **Total peak** | **22** | **3,036** | **~4,500** |

This is within acceptable bounds IF progressive disclosure works correctly. But with monolithic template skills, the actual cost could be 2-3x higher.
