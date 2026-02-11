# CE Audit: CLAUDE.md + Rules

**Date**: 2026-02-09
**Auditor**: Context Engineer (Opus 4.6)
**Governing Principles**: `rites/ecosystem/mena/claude-md-architecture/first-principles.md`

---

## Summary

| Metric | Value |
|--------|-------|
| **CLAUDE.md lines** | 94 (including HTML markers) |
| **CLAUDE.md words** | 496 |
| **CLAUDE.md estimated tokens** | ~700 |
| **CLAUDE.md bytes** | 3,722 |
| **Knossos-owned sections** | 6 (execution-mode, quick-start, agent-routing, commands, agent-configurations, platform-infrastructure) |
| **Satellite sections** | 1 (user-content) |
| **Rules (template source)** | 8 (7 internal/ + 1 mena/) |
| **Rules (projected)** | 9 (8 templates + 1 orphan: internal-usersync) |
| **Rules path coverage** | `internal/agent`, `internal/hook`, `internal/inscription`, `internal/materialize`, `internal/provenance`, `internal/sails`, `internal/session`, `internal/usersync` (DELETED), `mena/` |

---

## CLAUDE.md Line-by-Line Value Assessment

### Section 1: execution-mode (lines 1-13, ~86 words, ~120 tokens)

**Verdict: KEEP -- this is the core behavioral contract.**

This is the most important section. It tells Claude which of three operating modes to use based on session state. Every line earns its per-turn cost:
- The table is a decision matrix Claude references on every routing decision
- "/consult for mode selection" is the correct escape hatch

No changes needed. This is what CLAUDE.md is for.

### Section 2: quick-start (lines 15-29, ~77 words rendered, ~110 tokens)

**Verdict: COMPRESS -- agent table duplicates Section 5.**

This section renders the active rite name plus an agent table. The problem: `agent-configurations` (Section 5) renders the *exact same agent list*. The quick-start adds the rite name and agent count -- useful -- but the 5-row table is pure duplication.

Current rendered form shows the same 5 agents (orchestrator, code-smeller, architect-enforcer, janitor, audit-lead) in both Section 2 and Section 5.

**Recommendation**: Keep the one-liner ("This project uses a 5-agent workflow (hygiene)") and the invocation hints. Remove the agent table from quick-start since Section 5 already lists them. Saves ~5 lines / ~50 tokens.

### Section 3: agent-routing (lines 31-37, ~46 words, ~65 tokens)

**Verdict: KEEP -- tight and correct.**

Two sentences of routing guidance plus an escape hatch. This is pure behavioral contract. Passes all six first-principles tests.

### Section 4: commands (lines 39-54, ~135 words, ~190 tokens)

**Verdict: COMPRESS -- partial knowledge dump.**

This section has two layers:
1. **The mapping table** (5 rows, CC Primitive to Knossos Name): This IS behavioral contract -- it tells Claude what tools exist and how they map.
2. **The explanation text** (lines 50-53): Three lines of clarification about dromena vs legomena lifecycle and agent spawning constraints.

The table earns its cost. The explanation lines are borderline:
- "Dromena have side effects..." -- useful behavioral rule
- "Agents cannot spawn other agents..." -- critical constraint, earns its cost
- "Full mapping: `lexicon` skill..." -- pure pointer, earns its cost at low cost

**Recommendation**: The 5-row table could be compressed to 3 rows. "Hook" and "CLAUDE.md" rows tell Claude about concepts it can observe directly (hooks fire automatically, CLAUDE.md is self-evident). Only Slash command, Skill tool, and Task tool require routing guidance.

Remove the Hook and CLAUDE.md rows. Saves ~2 lines / ~30 tokens. Minor optimization.

### Section 5: agent-configurations (lines 56-66, ~63 words, ~90 tokens)

**Verdict: KEEP -- but see Section 2 overlap.**

This is a generated manifest of available agents. It is behavioral contract (what Claude can delegate to). The one-line descriptions are appropriately terse.

If the Section 2 table is removed per the recommendation above, this becomes the single source for the agent list. Correct behavior.

### Section 6: platform-infrastructure (lines 68-73, ~39 words, ~55 tokens)

**Verdict: KEEP -- extremely tight.**

Two sentences. Both are critical behavioral constraints:
- Hooks auto-inject context (don't duplicate in CLAUDE.md)
- Context mutations only via Moirai (prevents incorrect direct writes)

This section exemplifies what CLAUDE.md should be.

### Section 7: user-content (lines 75-94, ~50 words rendered, ~150 tokens)

**Verdict: MIXED -- contains stale knowledge and anti-patterns.**

This is the satellite-owned section where the user has placed project-specific instructions. Three subsections:

#### Anti-Patterns (lines 78-81): KEEP
Three critical don'ts. Pure behavioral contract. Every line earns its cost.

#### Active Refactoring (lines 83-86): CRITICAL FINDING -- STALE CONTENT
Three bullets that are factually incorrect or stale:
- Line 84: "`skills/` -> `commands/` unification complete (ADR-0021)" -- This is *historical* information. It describes something that already happened. Claude does not need to know this. It violates Principle 2 (Stable Content Only) and Principle 6 (Decay Test). **Already stale.**
- Line 85: "User sync system (`internal/usersync/`) is new" -- **`internal/usersync/` was deleted in Phase 4b (ADR-0026).** This line is actively misleading. It points Claude at a directory that does not exist. **Critically stale.**
- Line 86: "`frontmatter.go` adds unified command schema with `invokable` routing" -- This describes a completed feature, not a behavioral contract. **Already stale.**

#### Context Loading (lines 88-93): COMPRESS
Five pointer lines. Some earn their cost, some don't:
- "Architecture: `MEMORY.md`" -- useful pointer
- "Build: `ari --help` or `CGO_ENABLED=0 go build ./cmd/ari`" -- useful
- "Test: `CGO_ENABLED=0 go test ./...`" -- useful
- "Templates: `knossos/templates/sections/*.md.tpl`" -- useful only when editing templates. Could be a rule.
- "Commands: `mena/` (source) -> `.claude/commands/` + `.claude/skills/` (projection)" -- duplicates information already in the commands table (Section 4). The mena rule already covers this.

---

## Critical Findings

### CRIT-1: Stale reference to deleted directory (user-content, line 85)

**Line**: `User sync system (internal/usersync/) is new`
**Reality**: `internal/usersync/` was deleted in Phase 4b (ADR-0026). The sync pipeline was absorbed into `internal/materialize/`.
**Impact**: Claude may attempt to explore or reference a nonexistent directory. If asked about user sync, it will look in the wrong place.
**Fix**: Delete this line entirely. If replacement is wanted, it should say: "Unified sync pipeline: `internal/materialize/` (rite + user scopes via `ari sync`)"
**Principle violated**: P2 (Stable Content Only), P6 (Decay Test)

### CRIT-2: Orphan rule for deleted package (`.claude/rules/internal-usersync.md`)

**File**: `/Users/tomtenuta/Code/knossos/.claude/rules/internal-usersync.md`
**Path scope**: `internal/usersync/**`
**Reality**: `internal/usersync/` does not exist. This rule can never activate.
**Source discrepancy**: No corresponding template in `knossos/templates/rules/`. This rule exists only in `.claude/rules/` -- it is an orphan from the pre-Phase 4b era.
**Impact**: Zero runtime cost (never matches), but it is provenance debt. A future `ari sync` may or may not clean it up depending on orphan handling behavior.
**Fix**: Delete `.claude/rules/internal-usersync.md`. If `ari sync --scope=user` handles orphan removal, it should catch this. If not, manual deletion is needed.

---

## High Findings

### HIGH-1: "Active Refactoring" subsection is entirely stale knowledge

**Location**: user-content section, lines 83-86
**Content**:
```
### Active Refactoring
- `skills/` -> `commands/` unification complete (ADR-0021)
- User sync system (`internal/usersync/`) is new
- `frontmatter.go` adds unified command schema with `invokable` routing
```
**Assessment**: All three bullets describe completed past work, not active refactoring. They are historical artifacts that violate Principle 2 and Principle 6. The subsection header "Active Refactoring" compounds the problem -- it implies currency when the content is stale.
**Fix**: Delete the entire "Active Refactoring" subsection. If there IS active refactoring happening, replace with current information. If not, the section should not exist.

### HIGH-2: Missing rules for high-edit-frequency paths

The following paths contain source files that are frequently edited but have no path-scoped rules:

| Path | File types | Why it needs a rule |
|------|-----------|---------------------|
| `rites/` | YAML manifests, agent configs | Rite manifest schema, agent slot constraints, dependency declarations. 12 rites with different structures. |
| `hooks/` | Shell scripts (.sh) | Hook output format (PreToolUseOutput), decision mapping, graceful degradation. These are the bash scripts being targeted for elimination. |
| `knossos/templates/` | Go templates (.tpl), rules (.md) | Template syntax constraints, section ownership model, region marker format. |
| `docs/decisions/` | ADR markdown | ADR numbering convention (ADR-0NNN), required sections, status values. |

**Assessment**: `rites/` is the most critical gap. It contains 12 rite directories with manifests, orchestrators, and agent configurations. Editing these without knowing the schema constraints is error-prone. The `mena/` rule exists but `rites/` does not -- yet they are equally important source directories.

**Fix**: Create rules templates for at least `rites/` and `knossos/templates/`. The `hooks/` rule is less urgent since the hook scripts are being replaced by Go code (which will be covered by an `internal/hook` rule).

### HIGH-3: Agent table duplication between quick-start and agent-configurations

**Sections**: quick-start (Section 2) and agent-configurations (Section 5)
**Overlap**: Both render the full agent list with role descriptions.
**Token cost**: ~50 tokens of pure duplication per turn, every turn.
**Fix**: Remove the agent table from quick-start. Keep the rite name + agent count one-liner. The template already has conditional logic -- adjust the template to emit just the summary line, not the table partial.

---

## Medium Findings

### MED-1: Commands table includes self-referential rows

**Location**: commands section (Section 4), rows for "Hook" and "CLAUDE.md"
**Assessment**: The Hook and CLAUDE.md rows describe concepts that are self-evident in context. Claude can observe CLAUDE.md (it is reading it). Hooks fire automatically (Claude does not invoke them). These rows describe infrastructure, not actionable routing. The three actionable rows (Slash command, Skill tool, Task tool) are the ones Claude actually uses for decision-making.
**Fix**: Consider removing the Hook and CLAUDE.md rows. Saves ~30 tokens. Low priority.

### MED-2: Context Loading pointers partially duplicate existing coverage

**Location**: user-content section, lines 88-93
**Assessment**:
- "Templates: `knossos/templates/sections/*.md.tpl`" -- This is useful only when editing templates. If a `knossos/templates/` rule existed (see HIGH-2), this pointer would be unnecessary in CLAUDE.md.
- "Commands: `mena/` (source) -> `.claude/commands/` + `.claude/skills/` (projection)" -- This duplicates the commands table (Section 4) and the mena rule. Three places state the same mapping.
**Fix**: Once rules exist for `knossos/templates/`, remove the Templates pointer from CLAUDE.md. The Commands pointer can be removed now since it duplicates.

### MED-3: `internal-materialize.md` rule references deleted package

**Location**: `/Users/tomtenuta/Code/knossos/.claude/rules/internal-materialize.md`, line 9
**Content**: `SyncOptions replaces both materialize.Options and usersync.Options (ADR-0026 Phase 4b)`
**Assessment**: The reference to "usersync.Options" is historical migration context. The usersync package no longer exists. Claude does not need to know what SyncOptions replaced -- it needs to know what SyncOptions IS.
**Fix**: Rewrite as: `SyncOptions controls both rite and user scope sync (scope, dry-run, recover, overwrite flags)`

### MED-4: `internal-provenance.md` rule references deleted package

**Location**: `/Users/tomtenuta/Code/knossos/.claude/rules/internal-provenance.md`, line 16
**Content**: `One-way dependency: materialize and usersync import provenance, never the reverse`
**Assessment**: `usersync` no longer exists. The dependency is now only `materialize -> provenance`.
**Fix**: Rewrite as: `One-way dependency: materialize imports provenance, never the reverse`

---

## Low Findings

### LOW-1: HTML comment markers consume ~14 lines

**Assessment**: The `<!-- KNOSSOS:START -->` / `<!-- KNOSSOS:END -->` markers consume 14 lines of the 94-line file. These are infrastructure for the sync pipeline and cannot be removed. However, they add ~50 tokens of non-informational content per turn.
**Impact**: Negligible. CC may or may not tokenize HTML comments efficiently. The structural value (enabling idempotent sync) far outweighs the token cost.

### LOW-2: Rule file sizes are well-controlled

All rule files are 12-20 lines, 71-173 words. This is excellent. Rules are tightly scoped, single-purpose, and load only when their path is touched. No rule exceeds 200 tokens. No compression needed.

### LOW-3: Template source files include Go template comments

The `.md.tpl` files contain `{{/* ... */}}` comments that don't render into the projected CLAUDE.md. These are appropriate development documentation and have zero runtime cost.

---

## Rules Gap Analysis

| Path | Has Rule? | Needs Rule? | Priority | What It Should Say |
|------|-----------|-------------|----------|-------------------|
| `internal/agent/` | Yes | Yes | -- | Archetypes, validation tiers, Task tool stripping |
| `internal/hook/` | Yes | Yes | -- | Output formats, decision mapping, graceful degradation |
| `internal/inscription/` | Yes | Yes | -- | Owner types, region markers, merge pipeline |
| `internal/materialize/` | Yes | Yes | -- | Unified pipeline, scope gating, idempotency |
| `internal/provenance/` | Yes | Yes | -- | Schema v2, collectors, manifest format |
| `internal/sails/` | Yes | Yes | -- | Ship colors, proofs, gate criteria |
| `internal/session/` | Yes | Yes | -- | FSM states, lock protocol, scan discovery |
| `internal/usersync/` | Yes (orphan) | **No** (deleted) | CRIT | **Delete this rule** |
| `mena/` | Yes | Yes | -- | Dromena vs legomena, frontmatter, projection |
| `rites/` | **No** | **Yes** | HIGH | Rite manifest schema (rite.yaml), orchestrator contract, agent slot declarations, dependency resolution order, shared/ overlay semantics |
| `knossos/templates/` | **No** | **Yes** | HIGH | Section ownership model (knossos/satellite/regenerate), region marker syntax, template Go syntax, rules path-scope frontmatter |
| `hooks/` | **No** | **Conditional** | MED | Hook output format, decision mapping, error-defaults-to-allow. Lower priority if bash hooks are being eliminated. |
| `docs/decisions/` | **No** | **Maybe** | LOW | ADR numbering (ADR-0NNN), required sections, status enum. Low priority -- ADRs are human-written docs. |
| `schemas/` | **No** | **Maybe** | LOW | JSON Schema conventions, schema versioning. Low edit frequency. |
| `cmd/ari/` | **No** | **Maybe** | LOW | Cobra command patterns, flag conventions, CGO_ENABLED=0 build requirement. |
| `internal/config/` | **No** | **Maybe** | LOW | Config loading, env var precedence. Low complexity. |
| `internal/frontmatter/` | **No** | **Maybe** | LOW | Frontmatter parsing, schema enforcement. |
| `internal/lock/` | **No** | **Maybe** | LOW | Flock protocol, stale threshold, TOCTOU handling. |
| `internal/rite/` | **No** | **Maybe** | LOW | Rite loading, ACTIVE_RITE management, dependency resolution. |
| `internal/validation/` | **No** | **Maybe** | LOW | Validation patterns, error accumulation. |
| `internal/worktree/` | **No** | **Maybe** | LOW | Git worktree management. |
| `agents/` | **No** | **Maybe** | LOW | Cross-cutting agent prompts (consultant, context-engineer, moirai). |
| `test/` | **No** | No | -- | Test fixtures; no special conventions needed. |
| `docs/` (other) | **No** | No | -- | Freeform documentation; no special conventions needed. |

---

## Token Budget Assessment

### Current Per-Turn Cost

| Component | Tokens (est.) |
|-----------|--------------|
| CLAUDE.md (all sections) | ~700 |
| HTML markers (14 lines) | ~50 |
| **Total inscription cost** | **~750 tokens/turn** |

### After Recommended Changes

| Component | Tokens (est.) | Delta |
|-----------|--------------|-------|
| Remove agent table from quick-start | -50 | |
| Remove stale "Active Refactoring" | -45 | |
| Remove duplicate Commands pointer | -15 | |
| **Net savings** | **~110 tokens/turn** | -15% |
| **New total** | **~640 tokens/turn** | |

This is a healthy budget. 640-750 tokens for a behavioral contract governing a 5-agent orchestrated workflow is within the optimal range (500-1000 tokens). The inscription is NOT bloated -- it is on the lean side of acceptable. The findings above are about correctness and staleness, not about size.

---

## Recommendations

### Immediate (CRIT/HIGH fixes)

1. **Delete the "Active Refactoring" subsection** from user-content. All three bullets are stale. If there is active refactoring, write new bullets. If not, delete the header too.

2. **Delete `.claude/rules/internal-usersync.md`**. The package it targets no longer exists. This is provenance debt. Also remove any template source if one was manually added outside the template pipeline.

3. **Remove the agent table from the quick-start template**. Change `quick-start.md.tpl` to emit only the summary line ("This project uses a N-agent workflow (rite-name)") and the invocation hints, not the full agent table partial. The agent-configurations section already lists all agents.

4. **Create rules for `rites/` and `knossos/templates/`**:

   **`knossos/templates/rules/rites.md`** (proposed):
   ```markdown
   ---
   paths:
     - "rites/**"
   ---

   When modifying files in rites/:
   - rite.yaml is the manifest: name, description, agents, dependencies, mena entries
   - Orchestrator prompts define the coaching contract (read-only, structured YAML output)
   - Agent prompts are specialist definitions with tool access and behavioral constraints
   - shared/ directory contains cross-rite overlay resources (inherited by all rites)
   - Dependencies are resolved in order: rite > dependency > shared > user
   - Never edit .claude/ directly from rite definitions -- run `ari sync` to project
   ```

   **`knossos/templates/rules/knossos-templates.md`** (proposed):
   ```markdown
   ---
   paths:
     - "knossos/templates/**"
   ---

   When modifying files in knossos/templates/:
   - sections/*.md.tpl: Go templates that render CLAUDE.md regions
   - rules/*.md: Path-scoped instructions projected to .claude/rules/
   - 3 owner types in sections: knossos (SYNC), satellite (PRESERVE), regenerate (from source)
   - Region markers: <!-- KNOSSOS:START name --> ... <!-- KNOSSOS:END name -->
   - Templates must be idempotent: rendering twice produces identical output
   - Changes here require `ari sync` to project into .claude/
   ```

5. **Fix stale references in existing rules**:
   - `internal-materialize.md` line 9: replace usersync.Options reference
   - `internal-provenance.md` line 16: remove usersync from dependency statement

### Near-term (MED fixes)

6. **Remove the Commands pointer** from user-content Context Loading. It duplicates the commands table and the mena rule.

7. **Consider removing Hook and CLAUDE.md rows** from the commands table. Minor token savings, lower priority.

### Structural observations

8. **The section decomposition is correct**. The 7-section architecture maps well to the first principles: execution-mode is identity, quick-start is capability summary, agent-routing is workflow, commands is tool mapping, agent-configurations is capability manifest, platform-infrastructure is constraint contract, user-content is satellite extension point. No sections need to merge. No new knossos-owned sections are needed.

9. **Rules are the right mechanism for path-specific knowledge.** The existing 8 rules are well-written, appropriately scoped, and tightly compressed. The pattern should extend to cover `rites/` and `knossos/templates/` as the highest-priority gaps.

10. **The inscription correctly leverages CC's precedence model.** CLAUDE.md provides the stable behavioral contract. Rules provide path-triggered implementation detail. Skills provide on-demand reference knowledge. MEMORY.md provides persistent project memory. Hooks inject ephemeral session context. Each tier is used appropriately.

---

## Compliance with First Principles

| Principle | Status | Notes |
|-----------|--------|-------|
| P1: Behavioral Contract | PASS | Inscription describes capabilities and workflow, not knowledge |
| P2: Stable Content Only | **FAIL** | "Active Refactoring" subsection contains stale/completed work |
| P3: Separation by Source | PASS | Clear knossos/satellite/regenerate ownership per section |
| P4: Injection for Transient | PASS | No session state in CLAUDE.md; hooks handle transient context |
| P5: Single Purpose per Content | **WARN** | Agent table appears in both quick-start and agent-configurations |
| P6: Decay Test | **FAIL** | 3 bullets in user-content would be stale within a month (and are stale now) |

---

## Files Referenced in This Audit

- `/Users/tomtenuta/Code/knossos/.claude/CLAUDE.md` -- the projected inscription (94 lines)
- `/Users/tomtenuta/Code/knossos/knossos/templates/sections/*.md.tpl` -- 7 section templates
- `/Users/tomtenuta/Code/knossos/knossos/templates/rules/*.md` -- 8 rule templates
- `/Users/tomtenuta/Code/knossos/.claude/rules/*.md` -- 9 projected rules (1 orphan)
- `/Users/tomtenuta/Code/knossos/.claude/rules/internal-usersync.md` -- orphan rule for deleted package
- `/Users/tomtenuta/Code/knossos/rites/ecosystem/mena/claude-md-architecture/first-principles.md` -- governing principles
