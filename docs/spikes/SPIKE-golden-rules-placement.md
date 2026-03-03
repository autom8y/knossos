# SPIKE: Golden Rules Placement -- MEMORY.md vs. CLAUDE.md user-content

> Research spike analyzing where Knossos project-level operational invariants ("golden rules") should live in the CC instruction hierarchy.

**Date**: 2026-03-01
**Author**: Spike (golden rules placement)
**Prior Art**: `docs/guides/claude-md-hierarchy.md`, `docs/spikes/SPIKE-claudemd-compression-plan.md`, `INTERVIEW_SYNTHESIS.md`

---

## 1. Question and Context

### What are we trying to learn?

The golden rules ("NEVER edit .claude/ directly", "NEVER add session artifacts to mena", ".claude/ is NOT a pure cache") currently live in MEMORY.md (CC auto-memory). The argument for promoting them to the project-level CLAUDE.md `user-content` section is that MEMORY.md is evictable under context pressure, whereas CLAUDE.md is structurally permanent (loaded every turn, unconditionally).

But are these rules truly project-specific (relevant only to knossos), or are they cross-project guidance that belongs at a higher level?

### What decision will this inform?

1. Whether to move golden rules from MEMORY.md into the project CLAUDE.md `user-content` satellite region
2. Whether any of these rules are cross-project (belong in `~/.claude/CLAUDE.md` instead)
3. Whether the current placement creates real risk of rule loss under context pressure

---

## 2. Anatomy of the Golden Rules

| Rule | Content | Scope |
|------|---------|-------|
| **R1** | Never edit `.claude/` directly -- edit source and rematerialize | Knossos project only |
| **R2** | Never add session artifacts to `rites/shared/mena/` | Knossos project only |
| **R3** | `.claude/` is NOT a pure cache -- contains user state | Knossos project only |

### Scope Analysis

All three rules are **exclusively relevant to the knossos project itself**. They reference knossos-specific concepts:
- `.claude/` as a materialization target (the source-to-projection model)
- `rites/shared/mena/` as a legomena directory
- Satellite regions as user state within `.claude/`

A non-knossos satellite project (e.g., a web app using knossos rites) does NOT need these rules -- it has no `rites/` directory, no materialization pipeline running in-repo, and no mena source files. Its `.claude/` IS a pure output directory.

**Verdict**: These rules are project-local to the knossos repo. They should NOT go in `~/.claude/CLAUDE.md` (global user scope).

---

## 3. The Three Candidate Locations

### Option A: MEMORY.md (current state)

**File**: `~/.claude/projects/-Users-tomtenuta-Code-knossos/memory/MEMORY.md`

| Property | Value |
|----------|-------|
| **Loaded** | Every turn (when present), same as CLAUDE.md |
| **Persistence** | Auto-memory -- CC manages lifecycle, evictable under context pressure |
| **Edit control** | Claude can self-modify (append, reorganize, evict) |
| **Git-tracked** | No -- lives in `~/.claude/projects/` outside the repo |
| **Survives materialization** | Yes -- not in `.claude/`, not affected by `ari sync` |
| **Token cost** | ~1,589 tokens (entire MEMORY.md, not just golden rules) |

**Risk**: CC can evict MEMORY.md content when context window fills up. The golden rules could be dropped if a long orchestrated session pushes against the window limit. However, in practice, MEMORY.md eviction is rare -- CC tends to evict older, less-referenced entries first, and the "Golden Rules" section is at the top of the file.

### Option B: Project CLAUDE.md `user-content` region (proposed)

**File**: `~/Code/knossos/.claude/CLAUDE.md`, inside `<!-- KNOSSOS:START user-content -->...<!-- KNOSSOS:END user-content -->`

| Property | Value |
|----------|-------|
| **Loaded** | Every turn, unconditionally |
| **Persistence** | Structurally permanent -- satellite region, never overwritten by `ari sync` |
| **Edit control** | Manual only (human or explicit Claude edit) |
| **Git-tracked** | Yes -- committed to the knossos repo |
| **Survives materialization** | Yes -- `user-content` has `owner: satellite` in manifest |
| **Token cost** | ~30-50 additional tokens for the 3 rules |

**Risk**: None structurally. The `user-content` region is `OwnerSatellite` -- the merger code explicitly preserves it (`mergeRegion()` returns `oldContent` for satellite regions, line 254 of `merger.go`). This is the one region designed for exactly this purpose.

### Option C: `~/.claude/CLAUDE.md` (global user scope)

**File**: `~/.claude/CLAUDE.md`

| Property | Value |
|----------|-------|
| **Loaded** | Every turn in every project |
| **Persistence** | Structurally permanent |
| **Edit control** | Manual only |
| **Git-tracked** | No |

**Why NOT**: As analyzed in Section 2, these rules are knossos-specific. Placing them at global scope violates the hierarchy principle from `docs/guides/claude-md-hierarchy.md` Rule 2: "Parent files contain only cross-project guidance." Loading knossos materialization rules in every project (including non-knossos ones) wastes tokens and creates confusion.

---

## 4. Detailed Comparison

| Criterion | MEMORY.md (A) | user-content (B) | Global (C) |
|-----------|:---:|:---:|:---:|
| Correct scope (knossos-only) | Yes | Yes | No |
| Survives context pressure | No | Yes | Yes |
| Survives `ari sync` | Yes | Yes | Yes |
| Git-tracked | No | Yes | No |
| Shared with collaborators | No | Yes | No |
| Token cost impact | 0 (already loaded) | +30-50 | +30-50 (every project) |
| Self-modifiable by Claude | Yes | No | No |
| Duplicated across levels | No | No | Yes (loads in non-knossos) |

---

## 5. Current user-content Section

The existing `user-content` satellite region in the knossos project CLAUDE.md already contains closely related content:

```markdown
## Project Context

### Anti-Patterns
- Never directly write `*_CONTEXT.md` -- use Moirai agent
- Never edit knossos-owned CLAUDE.md sections -- lost on sync
- Satellite regions are user-safe, knossos regions are platform-owned

### Context Loading
Architecture: `MEMORY.md`
Build: `ari --help` or `CGO_ENABLED=0 go build ./cmd/ari`
Test: `CGO_ENABLED=0 go test ./...`
Templates: `knossos/templates/sections/*.md.tpl`
```

The anti-patterns section already contains a partial version of the golden rules ("Never edit knossos-owned CLAUDE.md sections", "Satellite regions are user-safe"). Promoting the full golden rules here would consolidate related guidance rather than split it across two locations.

---

## 6. Risk Assessment: MEMORY.md Eviction

How real is the eviction risk?

**Low in practice.** CC's auto-memory eviction prioritizes:
1. Older entries (golden rules are at line 1-5, frequently referenced)
2. Completed work logs (the "Completed Initiatives" section is the real eviction target)
3. Large low-signal blocks

The golden rules at 3 lines / ~60 tokens are unlikely eviction candidates. But "unlikely" is not "guaranteed." The entire point of structurally permanent context (CLAUDE.md) vs. managed context (MEMORY.md) is that structural guarantees eliminate probabilistic reasoning about what's in context.

**The question is not "will eviction happen?" but "should safety-critical rules depend on probabilistic persistence?"**

---

## 7. Recommendation

**Promote the golden rules to the `user-content` satellite region of the project CLAUDE.md.**

Rationale:
1. **Correct scope**: Rules are knossos-project-specific, and `user-content` is the designated place for project-specific user guidance
2. **Structural permanence**: Satellite regions are guaranteed loaded every turn, not subject to eviction
3. **Git-tracked**: Rules become part of the repo -- shared with any future collaborators, versioned, reviewable
4. **Consolidation**: The `user-content` section already contains partial versions of these rules; promotion eliminates duplication between MEMORY.md and CLAUDE.md
5. **Minimal cost**: ~30-50 additional tokens per turn for safety-critical rules is a worthwhile tradeoff
6. **MEMORY.md cleanup**: After promotion, the "Golden Rules" and "Materialization Invariants" sections in MEMORY.md can be trimmed or replaced with a pointer (`See: project CLAUDE.md user-content section`), recovering ~100 tokens of MEMORY.md budget for operational content

### Suggested Edit

Replace the current `user-content` section with:

```markdown
<!-- KNOSSOS:START user-content -->
## Project Context

### Golden Rules
- **NEVER edit `.claude/` directly** -- edit source (rites/, mena/, knossos/templates/, user-*/) and rematerialize
- **NEVER add session artifacts to mena** -- legomena are permanent platform knowledge; ephemeral content goes in `.sos/wip/`
- **`.claude/` is NOT a pure cache** -- it contains user state (satellite regions). "Delete and regen" destroys user content

### Anti-Patterns
- Never directly write `*_CONTEXT.md` -- use Moirai agent
- Never edit knossos-owned CLAUDE.md sections -- lost on sync

### Context Loading
Build: `CGO_ENABLED=0 go build ./cmd/ari`
Test: `CGO_ENABLED=0 go test ./...`
Templates: `knossos/templates/sections/*.md.tpl`
<!-- KNOSSOS:END user-content -->
```

Changes:
- Added "Golden Rules" subsection with the 3 critical invariants
- Removed "Satellite regions are user-safe, knossos regions are platform-owned" (redundant with golden rule R3)
- Removed `Architecture: MEMORY.md` pointer (MEMORY.md is auto-loaded; pointing to it adds no value)
- Net addition: ~3 lines / ~50 tokens

---

## 8. Follow-Up Actions

1. **Edit the `user-content` region** in `/Users/tomtenuta/Code/knossos/.claude/CLAUDE.md` with the suggested content above
2. **Trim MEMORY.md**: Replace the "Golden Rules" and "Materialization Invariants" sections with a brief pointer: `See project CLAUDE.md user-content for golden rules`
3. **Verify `ari sync`**: Run `ari sync` and confirm the `user-content` region is preserved (it should be -- `owner: satellite` in manifest)
4. **Do NOT** add these rules to `~/.claude/CLAUDE.md` -- they are not cross-project guidance

---

## 9. Summary

The user's intuition is correct: the golden rules are project-local to knossos and belong in the project-level CLAUDE.md `user-content` satellite region, not in MEMORY.md (evictable) and not in global `~/.claude/CLAUDE.md` (wrong scope). The `user-content` section is explicitly designed for this use case -- it is satellite-owned (never overwritten by `ari sync`), git-tracked, and structurally permanent in every conversational turn.
