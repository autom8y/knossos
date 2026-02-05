# Context Tier Model

## Status

Draft | 2026-02-05

## Problem Statement

The CLAUDE.md file is the single most expensive piece of context in the Knossos platform. It is loaded into every Claude Code agent's context window on every conversational turn. Today, the project-level CLAUDE.md (`.claude/CLAUDE.md`) contains 329 lines, of which 135 are satellite user-content. Two parent files -- `~/.claude/CLAUDE.md` (180 lines) and `~/Code/.claude/CLAUDE.md` (180 lines) -- duplicate nearly the entire knossos-managed portion of the project file. Claude Code loads all three into the system prompt on every turn.

**Measured cost** (using 4 chars/token heuristic):

| File | Lines | Chars | Est. Tokens |
|------|-------|-------|-------------|
| `~/.claude/CLAUDE.md` | 180 | ~6,800 | ~1,700 |
| `~/Code/.claude/CLAUDE.md` | 180 | ~6,800 | ~1,700 |
| `.claude/CLAUDE.md` (project) | 329 | ~12,400 | ~3,100 |
| `MEMORY.md` (project memory) | 46 | ~2,000 | ~500 |
| Global `~/CLAUDE.md` | ~100 | ~3,800 | ~950 |
| **Total per turn** | | **~31,800** | **~7,950** |

The parent CLAUDE.md files (`~/.claude/CLAUDE.md` and `~/Code/.claude/CLAUDE.md`) are byte-for-byte identical to the knossos-managed sections of the project CLAUDE.md, contributing ~3,400 tokens of pure duplication every turn. Meanwhile, the project file's `user-content` section (lines 194-329, 135 lines) contains material that duplicates MEMORY.md or belongs in on-demand skills.

**The core dysfunction**: Content at the wrong abstraction level. CLI reference (35 lines of `ari` commands), Moirai invocation patterns (40 lines), identity tables (15 lines), and full architecture descriptions (134 lines) all occupy L0 -- the always-loaded tier -- when most of this content is needed only occasionally and is available via cheaper mechanisms.

## Design Goals

1. **Reduce fixed token cost by 49%**: From ~7,950 tokens/turn to ~4,050 tokens/turn.
2. **Eliminate duplication across CLAUDE.md hierarchy**: Parent files should contain only content unique to their scope (cross-project guidance), not copies of project-level sections.
3. **Establish tier placement rules**: Every piece of context has exactly one canonical tier, with clear criteria for tier assignment.
4. **Preserve agent effectiveness**: No behavioral regression. Agents must still find the information they need, just via the appropriate tier mechanism rather than always-loaded.
5. **Backward compatibility**: The migration must work with existing satellites and existing inscription infrastructure. No schema version bump required if changes are additive.

## Tier Definitions

### L0: Always-Loaded (CLAUDE.md)

**Mechanism**: Loaded by Claude Code into every agent's system prompt on every conversational turn. No agent action required.

**Content criteria**: Information that MUST be immediately available without any tool invocation. Failure to have this information would cause the agent to make a wrong decision before it has a chance to look anything up.

**What belongs here**:
- Execution mode (determines how the agent behaves for the entire session)
- Active team/agent roster (who to delegate to)
- Navigation pointers (where to find everything else)
- Hard constraints (anti-patterns that prevent costly mistakes)
- Slash command behavior rules

**What does NOT belong here**:
- CLI reference (agent can run `ari --help`)
- Invocation examples (agent can load a skill)
- Architecture descriptions (agent can read docs)
- File location tables (agent has MEMORY.md and Grep)

**Target budget**: Less than 120 knossos-managed lines. Less than 50 satellite lines. Total less than 170 lines, estimated at ~1,200 tokens.

**Token economics**: ~1,200 tokens x 3 files (project + 2 parents) = ~3,600 tokens if parents duplicate. After deduplication (parents contain only unique content): ~1,200 project + ~200 parent overhead = ~1,400 tokens.

### L1: Tool-Invoked Operational (Bash tool / `ari --help`)

**Mechanism**: Agent runs a Bash command to retrieve information. Cost: one tool call (~200-500 tokens of output).

**Content criteria**: Operational procedures the agent needs to execute but does not need to memorize. The agent knows the tool exists (from L0 pointers) and invokes it when needed.

**What belongs here**:
- CLI command reference (`ari --help`, `ari session --help`)
- Build and test commands (`just build`, `go test ./...`)
- Hook configuration details
- Dynamic context syntax

**Cost model**: Zero tokens when not needed. ~200-500 tokens per invocation. An agent needing CLI reference pays 300 tokens once vs. 1,700 tokens every turn.

### L2: Skill-Loaded Domain Knowledge (Skill tool)

**Mechanism**: Agent invokes a skill via the Skill tool. Claude Code loads the skill's markdown files into context. Cost: ~500-2,000 tokens per skill load.

**Content criteria**: Domain knowledge, constraints, patterns, and templates that specialist agents need during specific phases of work. General-purpose knowledge that multiple agents might need but not on every turn.

**What belongs here**:
- Moirai invocation patterns and state management rules
- Knossos identity and mythology mapping
- Architecture reference (session FSM, White Sails rules, rites catalog)
- Code conventions and standards
- Orchestration loop patterns
- Getting-help routing table (expanded)

**Cost model**: Zero tokens when not needed. ~500-2,000 tokens when loaded. A skill is loaded once per conversation when the agent enters the relevant domain.

### L3: Reference Documents (Read tool)

**Mechanism**: Agent reads a specific file via the Read tool. Cost: variable, typically 500-5,000 tokens.

**Content criteria**: Full reference material, decision records, guides, and detailed specifications. Content that is useful as background but too large or specialized for L2 skills.

**What belongs here**:
- ADRs (`docs/decisions/`)
- Philosophy and doctrine (`docs/doctrine/`)
- User guides (`docs/guides/`)
- Full PRDs and TDDs (`docs/design/`, `docs/requirements/`)
- Source code reference (reading actual implementation files)

**Cost model**: Zero tokens when not needed. Variable cost on read. Agents read specific documents when they need authoritative detail on a specific topic.

## Current State Analysis

### Project CLAUDE.md Section-by-Section Audit

| Section | Owner | Lines | Est. Tokens | Current Tier | Appropriate Tier | Action |
|---------|-------|-------|-------------|-------------|-----------------|--------|
| `execution-mode` | knossos | 14 | ~110 | L0 | L0 | KEEP -- determines agent behavior |
| `knossos-identity` | knossos | 17 | ~200 | L0 | L2 | REMOVE -- mythology reference, not decision-critical |
| `quick-start` | regenerate | 15 | ~170 | L0 | L0 | KEEP -- agent roster is essential |
| `agent-routing` | knossos | 6 | ~60 | L0 | L0 | KEEP -- delegation rules |
| `skills` / `commands` | knossos | 4 | ~80 | L0 | L0 | KEEP (as `commands`) -- navigation pointer |
| `agent-configurations` | regenerate | 8 | ~100 | L0 | L0 | KEEP -- agent file locations |
| `hooks` | knossos | 3 | ~40 | L0 | L1 | COLLAPSE -- merge into `platform-infrastructure` |
| `dynamic-context` | knossos | 3 | ~40 | L0 | L1 | COLLAPSE -- merge into `platform-infrastructure` |
| `ariadne-cli` | knossos | 35 | ~380 | L0 | L1 | REMOVE -- replace with `ari --help` pointer |
| `getting-help` | knossos | 14 | ~140 | L0 | L0 (trimmed) / L2 | TRIM -- reduce to `/consult` pointer |
| `state-management` | knossos | 41 | ~470 | L0 | L2 | REMOVE -- move to `moirai-ref` skill |
| `slash-commands` | knossos | 3 | ~30 | L0 | L0 | KEEP |
| `user-content` | satellite | 135 | ~1,280 | L0 | L0 (trimmed) | TRIM -- move bulk to L2/L3 |
| **Total** | | **298** | **~3,100** | | | |

### Parent File Duplication

Both `~/.claude/CLAUDE.md` and `~/Code/.claude/CLAUDE.md` contain identical copies of the knossos-managed sections (execution-mode through slash-commands). Neither file contains any unique content. They exist solely because Claude Code's CLAUDE.md hierarchy loads parent directories.

| Parent File | Unique Content | Duplicated Content | Waste |
|-------------|---------------|-------------------|-------|
| `~/.claude/CLAUDE.md` | 0 lines | 180 lines (~1,700 tok) | 100% waste |
| `~/Code/.claude/CLAUDE.md` | 0 lines | 180 lines (~1,700 tok) | 100% waste |
| **Subtotal** | | | ~3,400 tokens/turn |

### Satellite Content (user-content) Audit

The `user-content` section in the project CLAUDE.md contains 135 lines. Here is the item-by-item assessment:

| Subsection | Lines | Duplicates | Target Tier | Rationale |
|-----------|-------|------------|-------------|-----------|
| Session FSM diagram | 12 | MEMORY.md line 11 | L2 (skill) | Available via `session/common` skill and source code |
| White Sails rules | 10 | MEMORY.md line 13 | L2 (skill) | Available in `internal/sails/color.go` |
| Rites catalog (12 items) | 16 | None at L2 | L2 (skill) | Create `rites-catalog` skill or add to `ecosystem-ref` |
| Orchestration loop | 12 | Orchestration skill | L2 (skill) | Already exists in orchestration skill |
| Inscription system | 7 | MEMORY.md line 10 | L2 (skill) | Already in `ecosystem-ref` |
| Decision records (`/stamp`) | 6 | None | L0 (keep) | Behavioral constraint, prevents audit gaps |
| Code conventions | 12 | MEMORY.md lines 32-33 | L2 (skill) | Move to `standards` skill |
| File structure tree | 16 | MEMORY.md lines 16-29 | L3 (MEMORY) | Exact duplicate |
| Anti-patterns list | 6 | Partial in MEMORY.md | L0 (keep) | High-value guardrails |
| Progressive disclosure | 7 | Meta-description | L0 (trim to 2) | Keep as a 2-line pointer |
| Key file locations table | 9 | MEMORY.md lines 16-29 | L3 (MEMORY) | Exact duplicate |
| Build/test commands | 2 | MEMORY.md lines 32-33 | L1 | Available via `just --list` |

**Summary**: Of 135 satellite lines, 96 lines (~71%) should move to L2/L3. Approximately 14 lines of high-value content should remain at L0.

### Total Token Economics

| Component | Current (tok/turn) | Target (tok/turn) | Savings |
|-----------|-------------------|-------------------|---------|
| Project CLAUDE.md (knossos) | ~1,820 | ~590 | -1,230 |
| Project CLAUDE.md (satellite) | ~1,280 | ~350 | -930 |
| `~/.claude/CLAUDE.md` | ~1,700 | ~100 | -1,600 |
| `~/Code/.claude/CLAUDE.md` | ~1,700 | ~100 | -1,600 |
| MEMORY.md | ~500 | ~500 | 0 |
| Global `~/CLAUDE.md` | ~950 | ~950 | 0 |
| **Total** | **~7,950** | **~2,590** | **-5,360 (67%)** |

## Target State

### CLAUDE.md Target Structure

The target project CLAUDE.md contains 9 regions totaling approximately 90 knossos-managed lines and 40 satellite lines (130 total, down from 298).

#### Region 1: `execution-mode` (14 lines) -- KEEP UNCHANGED

**Owner**: knossos | **Tier**: L0

No changes. This section determines how the agent behaves (native/cross-cutting/orchestrated) and must be immediately available.

```
<!-- KNOSSOS:START execution-mode -->
## Execution Mode
[... existing 14 lines unchanged ...]
<!-- KNOSSOS:END execution-mode -->
```

#### Region 2: `quick-start` (variable, ~15 lines) -- KEEP UNCHANGED

**Owner**: regenerate (source: `ACTIVE_RITE+agents`) | **Tier**: L0

No changes. Regenerated from active rite. Contains agent table essential for delegation decisions.

#### Region 3: `agent-routing` (6 lines) -- KEEP UNCHANGED

**Owner**: knossos | **Tier**: L0

No changes. Delegation rules are always needed.

#### Region 4: `commands` (8 lines) -- KEEP (replaces `skills`)

**Owner**: knossos | **Tier**: L0

Existing `commands` section template is already correct. This replaces the legacy `skills` section name. The template at `knossos/templates/sections/commands.md.tpl` already has the right content.

#### Region 5: `agent-configurations` (variable, ~8 lines) -- KEEP UNCHANGED

**Owner**: regenerate (source: `agents/*.md`) | **Tier**: L0

No changes. Agent file locations are essential for Task tool delegation.

#### Region 6: `platform-infrastructure` (4 lines) -- NEW

**Owner**: knossos | **Tier**: L0

Collapses `hooks`, `dynamic-context`, and `ariadne-cli` into a single navigation pointer. Replaces 42 lines with 4.

Target content:
```markdown
## Platform Infrastructure

Hooks auto-inject session context (no manual loading). CLI operations: run `ari --help`.
State management: use `Task(moirai, "...")` for all `*_CONTEXT.md` changes.
Build: `cd ariadne && just build` | Test: `go test ./...`
```

**Rationale**: The 35-line `ariadne-cli` section contains CLI reference that `ari --help` provides on-demand. The 41-line `state-management` section contains Moirai invocation patterns that belong in a skill (the agent prompt for Moirai already contains this). The `hooks` and `dynamic-context` sections (3 lines each) convey so little that they are better as part of a consolidated pointer.

#### Region 7: `navigation` (2 lines) -- NEW (replaces `getting-help`)

**Owner**: knossos | **Tier**: L0

Replaces the 14-line getting-help table with a minimal pointer.

Target content:
```markdown
## Navigation

For workflow routing use `/consult`. For domain knowledge use the Skill tool.
```

**Rationale**: The original `getting-help` table maps question types to skills. This mapping is itself available via `/consult` and the skill directory. The table duplicates the skill index without adding value at L0.

#### Region 8: `slash-commands` (3 lines) -- KEEP UNCHANGED

**Owner**: knossos | **Tier**: L0

No changes. Behavioral constraint for all agents.

#### Region 9: `user-content` (satellite, ~40 lines) -- TRIM

**Owner**: satellite | **Tier**: L0

After removing content that belongs at L2/L3, the satellite section retains only high-value operational content not available elsewhere:

Target content (approximately):
```markdown
## Project-Specific Instructions

### Decision Records
When making significant workflow decisions, use `/stamp` to record rationale in the clew. Triggers:
- Sacred path edits (`.claude/`, `*_CONTEXT.md`, `docs/decisions/`)
- Repeated failures (same command failed 2+ times)
- Multi-file changes (5+ files modified)

### Anti-Patterns to Avoid
1. **Direct writes to `*_CONTEXT.md`** -- Use Moirai agent instead
2. **Modifying knossos-owned sections** -- Edits are lost on sync
3. **Skipping `/stamp` on significant decisions** -- Audit trail gaps
4. **Shipping GRAY without QA acknowledgment** -- False confidence risk
5. **Swap-rite for temporary skill needs** -- Use invoke instead (cheaper)

### Active Refactoring (2026-02)
- `skills/` -> `commands/` unification in progress
- User sync system (`internal/usersync/`) is new

### Context Loading
Load on demand: Skills via Skill tool, agent prompts on Task invocation, docs via Read tool.
```

**Items removed** (with destination):

| Removed Item | Destination | Justification |
|-------------|-------------|---------------|
| Session FSM diagram | `session/common` skill (already exists) | Duplicate of MEMORY.md and source code |
| White Sails rules | `sails-ref` skill (create) or `ecosystem-ref` | Duplicate of MEMORY.md |
| Rites catalog | `ecosystem-ref` skill (already covers this) | Low L0 value, rarely needed |
| Orchestration loop | `orchestration` skill (already exists) | Exact duplicate |
| Inscription system | `ecosystem-ref` skill (already exists) | Duplicate of MEMORY.md |
| Code conventions | `standards` skill (already exists) | Duplicate of MEMORY.md |
| File structure tree | MEMORY.md (already there) | Exact duplicate |
| Key file locations | MEMORY.md (already there) | Exact duplicate |
| Build/test commands | `platform-infrastructure` section (2-line pointer) | Available via `just --list` |

### Parent CLAUDE.md Target Structure

#### `~/.claude/CLAUDE.md` (~8 lines)

This file should contain only cross-project guidance that applies regardless of which project directory Claude Code is operating in. For the Knossos user, this means:

```markdown
## Global Preferences

- Use Go conventions (gofmt, golint) for Go projects
- Prefer editing existing files over creating new files
- No unnecessary documentation files unless requested
```

All knossos-managed sections (execution-mode, quick-start, agent-routing, etc.) are REMOVED. They are project-specific and belong only in the project CLAUDE.md.

#### `~/Code/.claude/CLAUDE.md` (~8 lines)

Same treatment. This file currently exists only because the inscription system was applied at the wrong directory level. Target: minimal cross-project defaults or empty file.

### New/Modified L2 Skills

The content removed from L0 needs canonical L2 homes. Most already exist:

| Content | Target Skill | Status |
|---------|-------------|--------|
| Moirai invocation patterns | `session/common` or `moirai-ref` | Moirai agent prompt already contains this |
| Knossos identity table | `ecosystem-ref` | Add a `## Terminology` section |
| Session FSM | `session/common` | Already exists |
| White Sails rules | `session/common` or `sails-ref` | Add to existing skill |
| Rites catalog | `ecosystem-ref` | Already partially covered |
| Code conventions | `standards` | Already exists |
| Orchestration loop | `orchestration` | Already exists |
| Getting-help routing table | `/consult` command | Already exists |

**New skills required**: None. All displaced content has an existing home. Two skills (`ecosystem-ref`, `session/common`) may need minor additions.

## CLAUDE.md Target Structure (Section-by-Section)

### New Section Order

```go
// Target section order for DefaultSectionOrder()
func DefaultSectionOrder() []string {
    return []string{
        // Core behavior (determines agent mode)
        "execution-mode",

        // Team context (who is available)
        "quick-start",
        "agent-routing",
        "commands",
        "agent-configurations",

        // Infrastructure pointer (how to access platform tools)
        "platform-infrastructure",

        // Navigation pointer (where to find everything else)
        "navigation",

        // Behavioral rules
        "slash-commands",

        // User customization
        "user-content",
    }
}
```

### Removed Sections

The following sections are removed from the manifest's default regions and section order:

| Section | Current Lines | Replacement |
|---------|-------------|-------------|
| `knossos-identity` | 17 | Content moves to `ecosystem-ref` skill |
| `hooks` | 3 | Absorbed into `platform-infrastructure` |
| `dynamic-context` | 3 | Absorbed into `platform-infrastructure` |
| `ariadne-cli` | 35 | Absorbed into `platform-infrastructure` (as `ari --help` pointer) |
| `getting-help` | 14 | Replaced by `navigation` (2-line pointer) |
| `state-management` | 41 | Absorbed into `platform-infrastructure` (as Moirai pointer) |

### Section Template Changes

| Template File | Action |
|--------------|--------|
| `knossos/templates/sections/execution-mode.md.tpl` | No change |
| `knossos/templates/sections/quick-start.md.tpl` | No change |
| `knossos/templates/sections/agent-routing.md.tpl` | No change |
| `knossos/templates/sections/commands.md.tpl` | No change |
| `knossos/templates/sections/agent-configurations.md.tpl` | No change |
| `knossos/templates/sections/platform-infrastructure.md.tpl` | CREATE -- new 4-line section |
| `knossos/templates/sections/navigation.md.tpl` | CREATE -- new 2-line section |
| `knossos/templates/sections/slash-commands.md.tpl` | No change |
| `knossos/templates/sections/user-content.md.tpl` | No change (satellite preserved) |
| `knossos/templates/sections/hooks.md.tpl` | DELETE (already deleted per git status) |
| `knossos/templates/sections/knossos-identity.md.tpl` | DELETE (already deleted per git status) |
| `knossos/templates/sections/skills.md.tpl` | DELETE (already deleted per git status) |
| `knossos/templates/sections/ariadne-cli.md.tpl` | DELETE (does not exist as template file) |
| `knossos/templates/sections/state-management.md.tpl` | DELETE |
| `knossos/templates/sections/getting-help.md.tpl` | DELETE |
| `knossos/templates/sections/dynamic-context.md.tpl` | DELETE |

## Implementation Plan

### Phase 1: Template and Generator Changes (inscription system)

**Files modified**:

1. **`/Users/tomtenuta/Code/roster/internal/inscription/manifest.go`**
   - `DefaultSectionOrder()` (lines 241-266): Replace with new 9-section order
   - `CreateDefault()` (lines 189-238): Update `defaultKnossosRegions` list -- remove `knossos-identity`, `hooks`, `dynamic-context`, `ariadne-cli`, `getting-help`, `state-management`; add `platform-infrastructure`, `navigation`

2. **`/Users/tomtenuta/Code/roster/internal/inscription/generator.go`**
   - `getDefaultSectionContent()` (lines 366-390): Remove entries for `knossos-identity`, `hooks`, `dynamic-context`, `ariadne-cli`, `getting-help`, `state-management`; add entries for `platform-infrastructure`, `navigation`
   - Remove functions: `getDefaultKnossosIdentityContent()`, `getDefaultHooksContent()`, `getDefaultDynamicContextContent()`, `getDefaultAriadneCliContent()`, `getDefaultGettingHelpContent()`, `getDefaultStateManagementContent()`
   - Add functions: `getDefaultPlatformInfrastructureContent()`, `getDefaultNavigationContent()`

3. **`/Users/tomtenuta/Code/roster/knossos/templates/CLAUDE.md.tpl`**
   - Replace section includes to match new section order
   - Remove: `knossos-identity`, `hooks`, `dynamic-context`, `ariadne-cli`, `getting-help`, `state-management`
   - Add: `platform-infrastructure`, `navigation`

4. **`/Users/tomtenuta/Code/roster/knossos/templates/sections/platform-infrastructure.md.tpl`** -- CREATE
5. **`/Users/tomtenuta/Code/roster/knossos/templates/sections/navigation.md.tpl`** -- CREATE
6. **`/Users/tomtenuta/Code/roster/knossos/templates/sections/state-management.md.tpl`** -- DELETE
7. **`/Users/tomtenuta/Code/roster/knossos/templates/sections/getting-help.md.tpl`** -- DELETE
8. **`/Users/tomtenuta/Code/roster/knossos/templates/sections/dynamic-context.md.tpl`** -- DELETE

### Phase 2: Parent CLAUDE.md Deduplication

**Files modified**:

1. **`/Users/tomtenuta/.claude/CLAUDE.md`** -- Replace all knossos-managed sections with minimal global preferences (~8 lines)
2. **`/Users/tomtenuta/Code/.claude/CLAUDE.md`** -- Replace all knossos-managed sections with minimal cross-project defaults (~8 lines)

These are manual edits outside the inscription system. The inscription system only manages the project-level CLAUDE.md. Parent files are maintained by the user.

### Phase 3: Satellite Content Trim

**Files modified**:

1. **`/Users/tomtenuta/Code/roster/.claude/CLAUDE.md`** -- Edit the `user-content` satellite region to remove 96 lines of content that belongs at L2/L3. Retain the ~40 lines specified in the Target State section above.

This is a satellite-owned region, so the edit is preserved across syncs. The trim must happen after Phase 1 (so the knossos-managed sections are already updated) but can be done in the same inscription sync.

### Phase 4: L2 Skill Enrichment

**Files modified** (additive, no breakage):

1. **Ecosystem-ref skill**: Add `## Terminology` section with knossos identity table and rites catalog
2. **Session/common skill**: Verify Session FSM and White Sails rules are present (they already are)
3. **Standards skill**: Verify code conventions are present (they already are)

### Phase 5: Verification

Run `ari sync inscription` and verify:
- New CLAUDE.md matches target structure
- No existing satellite content is lost (only deliberately removed items)
- Agents can still find all displaced content via the appropriate tier mechanism
- Token count of generated CLAUDE.md is under 170 lines

## Migration Strategy

### Backward Compatibility: COMPATIBLE (with caveats)

The changes are backward-compatible at the inscription system level:

1. **Manifest schema**: No schema version bump needed. The `schema_version` remains `"1.0"`. Section order changes and region additions/removals are normal manifest operations.
2. **Region ownership model**: No changes. `knossos`, `regenerate`, and `satellite` ownership types are unchanged.
3. **Existing satellites**: The satellite `user-content` region is preserved. Knossos-managed sections are regenerated on sync, so the content change is applied automatically.

**Caveats requiring coordination**:

1. **Parent CLAUDE.md files**: These are outside the inscription system. The user must manually edit `~/.claude/CLAUDE.md` and `~/Code/.claude/CLAUDE.md`. If not edited, the old content still loads (no breakage, just wasted tokens).
2. **Satellite content trim**: The user must manually trim the `user-content` section. The inscription system preserves satellite content -- it will not remove it automatically.
3. **Agents referencing removed sections**: If any agent prompt or skill references "see the `ariadne-cli` section of CLAUDE.md" or "see the `state-management` section", those references must be updated. This is a documentation update, not a functional breakage.

### Migration Sequence

1. Merge Phase 1 changes (templates + generator)
2. Run `ari sync inscription` -- this regenerates all knossos-managed sections
3. Manually edit parent CLAUDE.md files (Phase 2)
4. Manually trim satellite content (Phase 3)
5. Verify L2 skill coverage (Phase 4)
6. Run verification (Phase 5)

Steps 3-4 can be done in any order and are independent of each other.

### Rollback

If the migration causes agent effectiveness problems:

1. Revert the generator and manifest changes (Phase 1)
2. Run `ari sync inscription` to regenerate old sections
3. Parent files and satellite content are manually managed and unaffected by rollback

## Verification Criteria

### Quantitative

| Metric | Current | Target | Method |
|--------|---------|--------|--------|
| Project CLAUDE.md lines (knossos) | 163 | <90 | `wc -l` on knossos regions |
| Project CLAUDE.md lines (satellite) | 135 | <50 | `wc -l` on satellite region |
| Project CLAUDE.md total lines | 298 | <140 | `wc -l` |
| Parent `~/.claude/CLAUDE.md` lines | 180 | <15 | `wc -l` |
| Parent `~/Code/.claude/CLAUDE.md` lines | 180 | <15 | `wc -l` |
| Total tokens/turn (all CLAUDE.md + MEMORY) | ~7,950 | <4,100 | Character count / 4 |
| Number of knossos-managed sections | 12 | 9 | Count regions in manifest |

### Qualitative

| Criterion | Verification Method |
|-----------|-------------------|
| Agent can determine execution mode | Read CLAUDE.md -- `execution-mode` section present |
| Agent can identify team members | Read CLAUDE.md -- `quick-start` section present with agent table |
| Agent can find CLI reference | Run `ari --help` -- output contains all commands |
| Agent can find Moirai patterns | Load `session/common` or moirai agent -- patterns present |
| Agent can find identity table | Load `ecosystem-ref` skill -- terminology section present |
| Agent can find rites catalog | Load `ecosystem-ref` skill -- rites listed |
| Agent can find code conventions | Load `standards` skill -- conventions present |
| No duplicate content across tiers | Grep for key phrases -- each appears in exactly one tier |

### Integration Test Matrix

| Satellite Type | Test | Expected Outcome |
|---------------|------|------------------|
| Fresh project (no satellite) | `ari sync inscription` on new project | CLAUDE.md generated with 9 sections, user-content has template placeholder |
| Existing project (current satellite) | `ari sync inscription` | Knossos sections regenerated to new structure; satellite content preserved as-is |
| Minimal satellite (empty user-content) | `ari sync inscription` | Clean output, no errors, all 9 sections present |
| Complex satellite (custom regions added) | `ari sync inscription` | Custom regions preserved, knossos sections updated, no region conflicts |
| Parent file migration | Manual edit of parent CLAUDE.md | Claude Code loads trimmed parent + full project CLAUDE.md without duplication |

## Appendix A: Content Migration Map

Complete mapping of every piece of current L0 content to its target tier:

```
CURRENT CLAUDE.md (329 lines)
├── execution-mode (14 lines) ──────────── STAYS at L0
├── knossos-identity (17 lines) ────────── MOVES to L2 (ecosystem-ref skill)
├── quick-start (15 lines) ─────────────── STAYS at L0
├── agent-routing (6 lines) ────────────── STAYS at L0
├── skills/commands (4 lines) ──────────── STAYS at L0 (renamed to commands)
├── agent-configurations (8 lines) ─────── STAYS at L0
├── hooks (3 lines) ────────────────────── ABSORBED into platform-infrastructure (L0, 1 line)
├── dynamic-context (3 lines) ──────────── ABSORBED into platform-infrastructure (L0, 0 lines)
├── ariadne-cli (35 lines) ─────────────── ABSORBED into platform-infrastructure (L0, 1 line)
│                                           Full content available at L1 (ari --help)
├── getting-help (14 lines) ────────────── REPLACED by navigation (L0, 2 lines)
│                                           Full routing at L2 (/consult)
├── state-management (41 lines) ────────── ABSORBED into platform-infrastructure (L0, 1 line)
│                                           Full content at L2 (moirai agent prompt)
├── slash-commands (3 lines) ───────────── STAYS at L0
└── user-content (135 lines) ───────────── TRIMMED to ~40 lines at L0
    ├── Session FSM (12) ───────────────── L2 (session/common skill) + L3 (MEMORY.md)
    ├── White Sails (10) ───────────────── L2 (session/common skill) + L3 (MEMORY.md)
    ├── Rites catalog (16) ─────────────── L2 (ecosystem-ref skill)
    ├── Orchestration loop (12) ────────── L2 (orchestration skill)
    ├── Inscription system (7) ─────────── L2 (ecosystem-ref skill) + L3 (MEMORY.md)
    ├── Decision records (6) ───────────── STAYS at L0 (behavioral constraint)
    ├── Code conventions (12) ──────────── L2 (standards skill) + L3 (MEMORY.md)
    ├── File structure (16) ────────────── L3 (MEMORY.md) -- exact duplicate removed
    ├── Anti-patterns (6) ──────────────── STAYS at L0 (behavioral constraint)
    ├── Progressive disclosure (7) ─────── TRIMMED to 2 lines at L0
    ├── Key file locations (9) ─────────── L3 (MEMORY.md) -- exact duplicate removed
    └── Build/test (2) ────────────────── L0 (platform-infrastructure, 1 line)
```

## Appendix B: Tier Assignment Decision Tree

Use this tree when deciding where new content should live:

```
Is this content needed to determine how the agent behaves
before it makes any tool calls?
├── YES → L0 (CLAUDE.md)
│   Examples: execution mode, team roster, hard constraints
└── NO
    Is this content an operational procedure the agent
    can retrieve via a CLI command?
    ├── YES → L1 (Bash tool / ari --help)
    │   Examples: CLI reference, build commands, hook details
    └── NO
        Is this content domain knowledge needed by specific
        agents during specific phases of work?
        ├── YES → L2 (Skill tool)
        │   Examples: Moirai patterns, architecture reference,
        │   code conventions, identity mapping
        └── NO → L3 (Read tool / docs/)
            Examples: ADRs, guides, doctrine, PRDs, TDDs
```

## Appendix C: Token Counting Methodology

Token estimates in this document use the following methodology:

1. **Character count**: Measured via `wc -c` on the rendered markdown (after template execution, including KNOSSOS markers).
2. **Token estimate**: Characters divided by 4. This is the standard approximation for English text with Claude's tokenizer. Markdown syntax slightly inflates this (more special characters), but the effect is small (less than 10%).
3. **Per-turn cost**: Claude Code loads CLAUDE.md files into the system prompt. Every user message and every agent response includes the full system prompt. The per-turn cost is the cost of the CLAUDE.md content multiplied by the number of turns in the conversation.
4. **Conversation cost**: A typical Knossos session involves 30-100 conversational turns. At ~7,950 tokens/turn, the CLAUDE.md hierarchy costs 238,500-795,000 tokens per session. At the target ~2,590 tokens/turn, this drops to 77,700-259,000 tokens per session.
