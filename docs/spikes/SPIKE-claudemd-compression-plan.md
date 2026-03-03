# SPIKE: CLAUDE.md Compression Initiative -- Implementation Plan

> Planning spike synthesizing the TDD-context-tier-model and C1-content-redesign design documents into an executable implementation plan with 8 prioritized recommendations.

**Date**: 2026-03-01
**Author**: Spike (CLAUDE.md compression)
**Prior Art**: `docs/design/TDD-context-tier-model.md`, `docs/design/C1-content-redesign.md`, `docs/guides/claude-md-hierarchy.md`

---

## Executive Summary

The CLAUDE.md compression initiative aims to reduce per-turn context overhead from the always-loaded CLAUDE.md hierarchy. Two detailed design documents exist (TDD-context-tier-model and C1-content-redesign) but implementation has been partial. Structural work is complete (section consolidation from 12 to 8 regions, legacy section removal, `know` section addition), but content-level compression is incomplete and the parent file duplication problem persists at the `~/Code/.claude/CLAUDE.md` level.

This spike distills the prior design work into 8 concrete, ordered recommendations and maps them to an executable implementation plan.

---

## 1. Question and Context

### What are we trying to learn?

Given the existing design documents (TDD-context-tier-model, C1-content-redesign), what is the exact current gap, and what is the optimal execution order for the remaining compression work?

### What decision will this inform?

- Prioritization of remaining compression tasks
- Whether to execute as a single sprint or phased rollout
- Which recommendations can be done independently vs. sequentially

---

## 2. Current State Audit

### Token Budget (measured 2026-03-01)

| File | Lines | Chars | Est. Tokens | Status |
|------|-------|-------|-------------|--------|
| `~/CLAUDE.md` (global root) | 143 | 5,888 | ~1,472 | Unchanged -- out of scope |
| `~/.claude/CLAUDE.md` (global knossos) | 6 | 258 | ~65 | OPTIMAL -- already trimmed |
| `~/Code/.claude/CLAUDE.md` (directory) | 64 | 2,823 | ~706 | PROBLEM -- full knossos sections from forge rite |
| `~/Code/knossos/.claude/CLAUDE.md` (project) | 128 | 6,003 | ~1,501 | PARTIAL -- structure done, content verbose |
| MEMORY.md (auto-memory) | 105 | 6,354 | ~1,589 | Separate concern |
| **Total per turn** | **446** | **21,326** | **~5,333** | |

### Structural Progress (what is DONE)

The TDD-context-tier-model's Phase 1 structural work is largely complete:

1. Section consolidation: 12 sections reduced to 8 (`execution-mode`, `quick-start`, `agent-routing`, `commands`, `agent-configurations`, `platform-infrastructure`, `know`, `user-content`)
2. Legacy sections removed: `knossos-identity`, `hooks`, `dynamic-context`, `ariadne-cli`, `getting-help`, `state-management` all gone
3. `platform-infrastructure` section created (replaces 3 collapsed sections)
4. `know` section added (new, post-TDD)
5. `navigation` and `slash-commands` sections added then deprecated (v18/v20 -- iterative refinement)
6. `DeprecatedRegions()` mechanism in manifest.go handles cleanup

### Content-Level Gaps (what REMAINS)

| Section | Current Lines | C1 Target Lines | Gap |
|---------|-------------|----------------|-----|
| `execution-mode` | 12 | 10 | Minor -- column header verbose |
| `quick-start` | 15 | 12-15 | Minor -- "This project uses a" prefix |
| `agent-routing` | 20 | 4 | MAJOR -- Exousia contract + Resume Protocol (~16 lines) |
| `commands` | 15 | 8 | MAJOR -- full Rosetta Stone table vs. compact 2-type list |
| `agent-configurations` | 11 | 9 | Minor |
| `platform-infrastructure` | 8 | 3-4 | MODERATE -- `/go` description + session detail |
| `know` | 20 | N/A | New section, not in C1 spec. Evaluate independently. |
| `user-content` | 12 | 23 (for roster) | ALREADY OPTIMAL for knossos project |

### Parent File Duplication

The critical remaining problem: `~/Code/.claude/CLAUDE.md` still contains 64 lines of knossos-managed sections from the **forge** rite (execution-mode, quick-start, agent-routing, agent-configurations, user-content). This means every turn in any project under `~/Code/` loads these sections twice -- once from the directory file and once from the project file. At ~706 tokens, this is pure waste.

Meanwhile, `~/.claude/CLAUDE.md` was already trimmed to 6 lines (~65 tokens). No action needed there.

---

## 3. The 8 Recommendations

Synthesized from TDD-context-tier-model and C1-content-redesign, adapted to current state:

### R1: Trim `~/Code/.claude/CLAUDE.md` to minimal cross-project defaults

**Savings**: ~650 tokens/turn
**Effort**: 5 minutes (manual edit)
**Risk**: None -- project-level files take precedence
**Dependencies**: None

Replace the 64-line file with 5 lines:

```markdown
# Code Directory Preferences

- Project-specific CLAUDE.md files always take precedence
- Use language-appropriate conventions for each project
- Prefer standard library solutions when possible
```

This was already specified in C1 Part 4 and the hierarchy guide. The file was trimmed in the roster project ecosystem but not for the knossos project at `~/Code/.claude/`.

### R2: Compress `agent-routing` section (20 lines to 4)

**Savings**: ~120 tokens/turn
**Effort**: 1 hour (template + generator default + test)
**Risk**: Low -- Exousia and Resume Protocol details available via agent prompts
**Dependencies**: None

Current content includes:
- Exousia jurisdiction contract explanation (6 lines) -- belongs in agent prompts, not L0
- Throughline Resume Protocol (8 lines) -- belongs in Pythia agent prompt, not L0

C1 target:
```markdown
## Agent Routing

Orchestrated sessions: delegate to specialists via Task tool. No session: execute directly or use `/task`.

Routing guidance: `/consult`
```

Files to modify:
- `knossos/templates/sections/agent-routing.md.tpl`
- `internal/inscription/generator.go` (`getDefaultAgentRoutingContent()`)

### R3: Compress `commands` section (15 lines to 8)

**Savings**: ~80 tokens/turn
**Effort**: 1 hour (template + generator default + test)
**Risk**: Low -- lexicon skill provides full mapping
**Dependencies**: None

The current "CC Primitives" Rosetta Stone table is 5-column, 5-row (CC Primitive, Knossos Name, Invocation, Source). The C1 design recommends a compact 2-type list format. However, the current format has earned its place -- the Rosetta Stone is the single most-referenced section for agents learning Knossos vocabulary.

**Revised recommendation**: Keep the Rosetta Stone table but trim the prose. Remove the 2-line explanation below the table ("Dromena have side effects..." + "Agents cannot spawn...") and the footer. The table itself is 8 lines; the prose adds 4 lines of content that the `lexicon` skill covers in depth.

Compressed target (10 lines, down from 15):
```markdown
## CC Primitives

| CC Primitive | Knossos Name | Invocation | Source |
|---|---|---|---|
| Slash command | **Dromena** | User types `/name` | `.claude/commands/` |
| Skill tool | **Legomena** | Model calls `Skill("name")` | `.claude/skills/` |
| Task tool | **Agent** | Model calls `Task(subagent_type)` | `.claude/agents/` |
| Hook | **Hook** | Auto-fires on lifecycle events | `.claude/settings.json` |
| CLAUDE.md | **Inscription** | Always in context | `knossos/templates/` |

Full mapping: `lexicon` skill.
```

Files to modify:
- `knossos/templates/sections/commands.md.tpl`
- `internal/inscription/generator.go` (`getDefaultCommandsContent()`)

### R4: Compress `platform-infrastructure` section (8 lines to 4)

**Savings**: ~50 tokens/turn
**Effort**: 30 minutes (template + generator default)
**Risk**: Low -- session details available via `ari --help`
**Dependencies**: None

Current content describes `/go`, session lifecycle commands, Fate skills, and hooks in full prose. The C1 target was:

```markdown
## Platform

**Entry**: `/go` -- cold-start dispatcher.

**Sessions**: Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.

**Hooks**: Auto-inject session context. CLI reference: `ari --help`.
```

The `/start`, `/park`, `/continue`, `/wrap` enumeration and Fate skill names (Clotho/Lachesis/Atropos) are mythology details that agents discover from agent prompts and skill files.

Files to modify:
- `knossos/templates/sections/platform-infrastructure.md.tpl`
- `internal/inscription/generator.go` (`getDefaultPlatformInfrastructureContent()`)

### R5: Evaluate and compress `know` section (20 lines)

**Savings**: ~100 tokens/turn (if compressed to ~8 lines)
**Effort**: 1 hour (template + generator default + evaluation)
**Risk**: Medium -- this section was added post-C1, need empirical data on usage
**Dependencies**: None

The `know` section (20 lines) was added after the C1 design was written and is not covered by it. It has 4 subsections (auto-load, load-on-demand, task-specific, literature) with explicit `Read()` call patterns.

**Assessment**: The auto-load instruction ("read `.know/architecture.md` before any code modification") has high L0 value. The other 3 subsections are on-demand guidance that agents can discover from the `.know/` directory listing or from MEMORY.md.

Compressed target (~8 lines):
```markdown
## Codebase Knowledge

Persistent knowledge in `.know/`. Auto-load before code changes:
- `Read(".know/architecture.md")` -- package structure, layers, data flow

On demand: `Read(".know/scar-tissue.md")`, `Read(".know/design-constraints.md")`, `Read(".know/conventions.md")`, `Read(".know/test-coverage.md")`. Literature: `Read(".know/literature-{domain}.md")`.

Refresh: `ari knows`. Regenerate: `/know [domain]`.
```

Files to modify:
- `knossos/templates/sections/know.md.tpl`
- `internal/inscription/generator.go` (add `getDefaultKnowContent()` if not already present)

### R6: Minor compression of `execution-mode` section

**Savings**: ~15 tokens/turn
**Effort**: 15 minutes
**Risk**: None
**Dependencies**: None

Per C1: rename column "Main Agent Behavior" to "Behavior" (saves 2 words per row x 3 rows). Change "Pythia coordinates; delegate via Task tool" to "Delegate via Task tool to specialist agents" (removes mythology term from L0).

Files to modify:
- `knossos/templates/sections/execution-mode.md.tpl`
- `internal/inscription/generator.go` (`getDefaultExecutionModeContent()`)

### R7: Minor compression of `quick-start` section

**Savings**: ~10 tokens/turn
**Effort**: 15 minutes
**Risk**: None
**Dependencies**: None

Per C1: remove "This project uses a" prefix (self-referential). Change to `6-agent workflow (rnd):` directly.

Files to modify:
- `knossos/templates/sections/quick-start.md.tpl`
- `internal/inscription/generator.go` (`getDefaultQuickStartContent()`)

### R8: Add KnossosVars for conditional build/test in `platform-infrastructure`

**Savings**: Prevents future token leakage (no direct savings)
**Effort**: 30 minutes
**Risk**: Low -- additive, existing mechanism
**Dependencies**: R4 (platform-infrastructure compression)

Per C1 Part 5: Add `build_command` and `test_command` to KnossosVars. Render conditionally in the platform-infrastructure template. This ensures satellite projects get their own build commands (or none), and the knossos source repo renders its specific commands.

Files to modify:
- `knossos/templates/sections/platform-infrastructure.md.tpl` (conditional rendering)
- `.knossos/KNOSSOS_MANIFEST.yaml` (add `knossos_vars` block if not present)

---

## 4. Comparison Matrix

| # | Recommendation | Tokens Saved | Effort | Risk | Independence |
|---|---------------|-------------|--------|------|-------------|
| R1 | Trim `~/Code/.claude/CLAUDE.md` | ~650/turn | 5 min | None | Fully independent |
| R2 | Compress `agent-routing` | ~120/turn | 1 hr | Low | Fully independent |
| R3 | Compress `commands` | ~80/turn | 1 hr | Low | Fully independent |
| R4 | Compress `platform-infrastructure` | ~50/turn | 30 min | Low | Fully independent |
| R5 | Compress `know` | ~100/turn | 1 hr | Medium | Fully independent |
| R6 | Minor `execution-mode` | ~15/turn | 15 min | None | Fully independent |
| R7 | Minor `quick-start` | ~10/turn | 15 min | None | Fully independent |
| R8 | KnossosVars build/test | 0 (preventive) | 30 min | Low | Depends on R4 |
| **Total** | | **~1,025/turn** | **~5 hrs** | | |

### Projected Token Budget After All 8 Recommendations

| File | Current (tok) | After (tok) | Change |
|------|-------------|-----------|--------|
| `~/CLAUDE.md` | ~1,472 | ~1,472 | 0 (out of scope) |
| `~/.claude/CLAUDE.md` | ~65 | ~65 | 0 (already optimal) |
| `~/Code/.claude/CLAUDE.md` | ~706 | ~56 | -650 (R1) |
| `~/Code/knossos/.claude/CLAUDE.md` | ~1,501 | ~1,126 | -375 (R2-R7) |
| MEMORY.md | ~1,589 | ~1,589 | 0 (separate concern) |
| **Total per turn** | **~5,333** | **~4,308** | **-1,025 (19%)** |

Note: The original TDD projected ~7,950 tokens/turn. The current baseline of ~5,333 reflects that significant structural compression has already been done (section consolidation, legacy removal, parent file trimming at `~/.claude/`). The remaining 1,025-token savings target is the realistic ceiling for content-level compression.

---

## 5. Implementation Plan

### Phase 1: Quick Wins (30 minutes, ~665 tokens saved)

Execute R1 and R6+R7 together. No code changes, no tests needed for R1.

1. **R1**: Manually edit `~/Code/.claude/CLAUDE.md` to 5-line minimal content
2. **R6**: Update `execution-mode.md.tpl` and generator default (column rename)
3. **R7**: Update `quick-start.md.tpl` and generator default (remove "This project uses a")
4. Run `ari sync` to verify regeneration
5. Verify no agent behavioral regression in one session

### Phase 2: Section Compression (2 hours, ~350 tokens saved)

Execute R2, R3, R4 in sequence. Each changes one template + one generator default.

1. **R2**: Compress `agent-routing` -- remove Exousia + Resume Protocol from template and default
2. **R3**: Compress `commands` -- trim prose below Rosetta Stone table
3. **R4**: Compress `platform-infrastructure` -- trim session detail to pointer
4. Run `CGO_ENABLED=0 go test ./internal/inscription/...` after each change
5. Run `ari sync` to verify end-to-end

### Phase 3: Knowledge Section + Preventive (1.5 hours, ~100 tokens saved + preventive)

Execute R5 and R8.

1. **R5**: Compress `know` section -- evaluate empirically first (run 2-3 sessions with compressed version)
2. **R8**: Add KnossosVars conditional build/test to platform-infrastructure template
3. Test with both knossos (has build command) and a satellite (no build command)

### Phase 4: Verification (30 minutes)

1. Measure final token budget: `wc -c` on all CLAUDE.md files
2. Run integration test matrix from C1 design doc (CT-01 through CT-10)
3. Verify deprecated regions cleaned up on sync
4. Run 2-3 agent sessions to confirm no behavioral regression

---

## 6. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Agent loses context after R2 (Exousia/Resume removed from L0) | LOW | Agent asks clarifying questions it didn't before | Exousia is in every agent prompt; Resume is Pythia-only. Both available at L2. |
| Commands Rosetta Stone compression loses navigability | LOW | Agents fail to find invocation patterns | Keep table, only trim prose; `lexicon` skill provides full reference |
| `know` section compression causes agents to skip `.know/` reading | MEDIUM | Agent makes code changes without reading architecture.md | Keep auto-load instruction; only compress on-demand subsections |
| KnossosVars not present in satellite manifests | NONE | Satellite simply renders no build/test line | Conditional rendering handles absent vars gracefully |
| `~/Code/.claude/CLAUDE.md` edit overwritten by something | LOW | Parent file duplication returns | Nothing in Knossos writes parent files; purely manual |

---

## 7. Relationship to Prior Design Documents

| Document | Status | Relevance to This Plan |
|----------|--------|----------------------|
| TDD-context-tier-model | Phase 1 DONE, Phases 2-5 PARTIAL | Structural source. Sections 3-6 inform R1-R5. |
| C1-content-redesign | UNIMPLEMENTED | Content spec. Exact target text for R2-R8. |
| claude-md-hierarchy guide | CURRENT | Documents parent file hierarchy and rules |

### What Changed Since C1 Was Written

1. **`know` section added**: Not in C1 spec. New section needs independent evaluation (R5).
2. **`navigation` and `slash-commands` sections deprecated**: C1 proposed them; they were created then removed in v18/v20. Already handled.
3. **Parent files partially trimmed**: `~/.claude/CLAUDE.md` already at 6 lines. `~/Code/.claude/CLAUDE.md` still verbose.
4. **Section count**: C1 targeted 9 sections. Current: 8 (the 9th was `navigation`, since deprecated and replaced by `know`). Slightly different composition but same spirit.

---

## 8. Follow-Up Actions

| # | Action | Priority | Depends On | Effort |
|---|--------|----------|-----------|--------|
| 1 | Execute Phase 1 (R1 + R6 + R7) | P0 | None | 30 min |
| 2 | Execute Phase 2 (R2 + R3 + R4) | P0 | Phase 1 verified | 2 hrs |
| 3 | Execute Phase 3 (R5 + R8) | P1 | Phase 2 verified | 1.5 hrs |
| 4 | Execute Phase 4 (verification) | P0 | Phase 3 done | 30 min |
| 5 | Update C1 design doc status to reflect completion | P2 | Phase 4 done | 15 min |
| 6 | Evaluate `~/CLAUDE.md` (143 lines) for separate compression | P3 | None | Separate spike |
| 7 | Evaluate MEMORY.md (105 lines) for compression | P3 | None | Separate spike |
| 8 | Monitor agent effectiveness over 5+ sessions post-compression | P1 | Phase 4 done | Ongoing |

---

## References

- `docs/design/TDD-context-tier-model.md` -- L0/L1/L2/L3 tier definitions, original token economics
- `docs/design/C1-content-redesign.md` -- Exact target content for all sections
- `docs/guides/claude-md-hierarchy.md` -- Parent file hierarchy documentation
- `internal/inscription/manifest.go` -- Section order, region definitions, deprecated regions
- `internal/inscription/generator.go` -- Default content generators
- `knossos/templates/sections/*.md.tpl` -- Current section templates
