# Mena Scope & Projection Overhaul — Decision Record

> Produced from structured interview, 2026-02-06. Informed by a 5-agent swarm analysis of the materialization and usersync pipelines.

## Problem Statement

Claude Code strips only `.md` when deriving command names. A source file `sync.dro.md` becomes `/cem:sync.dro` — a different command than `/cem:sync` from user-level `sync.md`. Shadowing never activates because names don't match. Result: 22+ commands appear twice in autocomplete, one tagged `(user)`, one tagged `(project)`.

Investigation revealed six anti-patterns: name bifurcation, legomena-as-commands leak, context pollution in non-knossos projects, governance bypass via user-level selection, structural degradation (flat vs. progressive disclosure), and shadow commands persisting across projects.

## Foundational Insight: The Distribution Model

The `user-` prefix on `user-agents/`, `user-hooks/`, `user-skills/` was misleading. These are not "user-authored" resources — they are **first-party distribution content** that ships with the knossos binary. In the distribution model:

- `mena/` = commands and skills distributed to every user
- `user-agents/` = agents distributed to every user
- `user-hooks/` = hooks distributed to every user

**User-level `~/.claude/` IS the distribution target.** `ari sync user` is the distribution mechanism. The sync of `mena/` to `~/.claude/commands/` is intentional and correct — not a legacy leak.

The duplication is caused by the `.dro` extension leaking into projected filenames and user sync not implementing dromena/legomena routing, not by the sync itself being wrong.

---

## Decisions

### D1: Strip .dro/.lego from Projected Filenames

**Decision**: Both materialization and user sync strip the `.dro` and `.lego` infixes when writing to target directories. `INDEX.dro.md` becomes `INDEX.md` in `.claude/commands/`. `INDEX.lego.md` becomes `INDEX.md` in `.claude/skills/`.

**Rationale**: `.dro`/`.lego` is a source-level routing signal (ADR-0023). It tells the materialization pipeline where to route — commands/ vs skills/. It should not leak into the consumer-facing namespace where Claude Code parses filenames.

**Aggressiveness**: No transition period. `.claude/` is generated output — next materialize fixes everything.

### D2: Shadowing as Default Behavior

**Decision**: Most commands exist at both user-level and project-level. Claude Code's native precedence (project shadows user) handles resolution. In a knossos-managed project, the project version shows. In a plain repo, the user version shows.

**Rationale**: This is the CSS cascade working correctly. The distributed version is the baseline; project-level materialization provides the override when a rite is active.

### D3: Two Scope Values — `user` and `project`

**Decision**: Add an optional `scope` field to `MenaFrontmatter`. Two values:

| Value | Meaning |
|-------|---------|
| `scope: user` | Only project to user-level `~/.claude/`. Skip during project materialization. |
| `scope: project` | Only project to project-level `.claude/`. Skip during user sync. |
| *(no field)* | Both pipelines include. Shadowing handles precedence. Default. |

**Rationale**: No `both` value — absence of the field IS the default (both). No `project-exclusive` — advisory enforcement makes it functionally identical to `project`. Two values, not four. Explicit scope means restriction; no scope means open.

### D4: Advisory Enforcement Only

**Decision**: When a scope violation is detected (e.g., user-level copy exists for a `scope: project` command), emit a warning during sync and validation. Never delete user-level files. Never block at runtime.

**Rationale**: The user owns `~/.claude/`. Knossos is a developer tool, not a security boundary. If someone deliberately bypasses project policy, that's a people problem. Tooling warns; trust is the enforcement mechanism.

### D5: Rename Source Directories

**Decision**: Rename distribution-level source directories in the knossos source tree:

| Before | After | Rationale |
|--------|-------|-----------|
| `user-agents/` | `agents/` | Remove misleading `user-` prefix. These are distribution agents. |
| `user-hooks/` | `hooks/` | Same. |
| `user-skills/` | *(merged into `mena/`)* | Skills are legomena. They belong in `mena/` with `.lego.md` convention. |

**Scope**: Source tree only. `$KNOSSOS_HOME` layout follows from whatever the build produces. Rite-level disambiguation is structural: `rites/*/agents/` vs top-level `agents/`.

### D6: Merge user-skills/ into mena/

**Decision**: Eliminate `user-skills/` as a separate directory. Migrate its content into `mena/` using the `.lego.md` extension convention established by ADR-0023. Legomena have always conceptually belonged to the mena system.

**Rationale**: `user-skills/` exists because it predates the mena convention. Now that legomena are first-class mena citizens (`.lego.md`), maintaining a separate directory creates confusion. One source directory for all distribution-level commands and skills.

### D7: Full Structure Parity Across Scopes

**Decision**: User sync produces the exact same output structure as project materialization. Both strip extensions. Both preserve directory trees. Single-file dromena stay as single files. Directory dromena stay as directories. No structural degradation at user level.

**Applies to all resource types**: agents, hooks, and mena.

**Rationale**: Previously, user sync produced flat files while materialize produced progressive-disclosure directories. This meant the user-level version was a degraded copy. Both scopes should present identical quality.

### D8: Materialize Owns Projection Logic

**Decision**: The mena projection engine (extension stripping, dromena/legomena routing, directory copying, scope filtering) lives in `internal/materialize/`. User sync calls into it with a different target path. No new `internal/mena/` package.

**Rationale**: One projection engine, two callers. Prevents structural drift between scopes. The materialize package already has the correct routing logic — user sync just needs to use it instead of its own copy pipeline.

### D9: One Atomic PR

**Decision**: Ship all changes in a single PR: extension stripping, scope field, directory renames, user-skills merge, structure parity, usersync upgrade. No stacked PRs, no phased delivery.

**Rationale**: The pieces are interdependent. Partial delivery creates inconsistent state (e.g., renamed source dirs but old sync paths). One coherent change, one review.

### D10: Wipe and Resync Manifests

**Decision**: When the new usersync runs for the first time after this change, existing manifests (`USER_COMMAND_MANIFEST.json`, `USER_AGENT_MANIFEST.json`, etc.) are wiped and rebuilt from scratch.

**Rationale**: Source directory renames and structural changes invalidate every checksum and path entry. Migration logic adds complexity for a one-time event. Wiping is clean and self-healing. Users lose divergence tracking history — acceptable for a major pipeline overhaul.

### D11: New ADR Required

**Decision**: Write a new ADR (next available number) documenting the scope model, distribution model, rename rationale, projection parity requirement, and the relationship to ADR-0023.

**Rationale**: This is an architectural decision with the same gravity as ADR-0023 (which established the mena convention). The scope model, distribution model insight, and pipeline unification are durable decisions that future contributors need to understand.

### D12: Full Parity for Agents and Hooks

**Decision**: Agents and hooks get the same treatment as mena: rename, structure parity, and scope field support. Not just mena.

**Rationale**: Consistent treatment across all resource types. The distribution pipeline is one system, not three.

### D13: No Embedding Constraints

**Decision**: Rite/mena embedding into the single binary is not yet implemented. This rename makes the eventual embedding cleaner — simpler paths, no `user-` prefix ambiguity.

---

## Implementation Sequence

All changes ship in one atomic PR. Internal ordering:

### Step 1: Source Tree Renames
- `user-agents/` → `agents/`
- `user-hooks/` → `hooks/`
- Merge `user-skills/` content into `mena/` (with `.lego.md` convention)
- Update all path references in Go source (`internal/usersync/`, `internal/config/`, `internal/paths/`, `internal/cmd/`)

### Step 2: Extension Stripping in Materialize
- Modify `materializeMena()` to strip `.dro`/`.lego` from filenames when copying to target
- `INDEX.dro.md` → `INDEX.md` in `.claude/commands/`
- `INDEX.lego.md` → `INDEX.md` in `.claude/skills/`
- Single-file dromena: `architect.dro.md` → `architect.md`
- Update routing tests

### Step 3: Scope Field
- Add `Scope` field to `MenaFrontmatter` struct in `internal/materialize/frontmatter.go`
- Add `MenaScope` type with `ScopeUser` and `ScopeProject` constants
- Add `EffectiveScope()` method (returns empty string for "both" default)
- Add validation: scope must be `"user"`, `"project"`, or empty
- Add `parseMenaScope()` helper for reading scope from INDEX files
- Integrate into `materializeMena()`: skip if `scope: user`
- Wire scope support for agents and hooks frontmatter

### Step 4: Usersync Calls Materialize
- Expose mena projection API from `internal/materialize/`
- Refactor usersync to call materialize's projection for mena (commands + skills)
- Apply same pattern for agents and hooks where applicable
- User sync skips commands with `scope: project`
- Implement manifest wipe-and-resync on schema version mismatch

### Step 5: Scope Annotations
- Audit all mena INDEX files and assign scope where needed
- Default (no field) for most commands
- `scope: project` for rite-specific commands (e.g., rite-switching, rite-aware navigation)
- `scope: user` for commands that should never appear at project level (if any identified)

### Step 6: ADR
- Write ADR documenting: distribution model, scope model, rename rationale, projection parity, relationship to ADR-0023

### Step 7: Validation & Testing
- Extend routing tests for scope filtering
- Extend usersync tests for structure parity
- Add scope validation to `ari agent validate` (or new `ari mena validate`)
- Verify wipe-and-resync behavior

---

## Invariants

These must hold after implementation:

1. **No `.dro` or `.lego` in projected filenames.** The `.claude/` directory never contains `.dro.md` or `.lego.md` files.
2. **Structure parity.** A command at `~/.claude/commands/X/` and `.claude/commands/X/` have identical directory structure (modulo content differences from rite overrides).
3. **Scope filtering is additive.** No `scope` field = included by both pipelines. Only explicit `scope: user` or `scope: project` restricts.
4. **One projection engine.** Both usersync and materialize produce output via the same code path. No structural drift.
5. **User files are sacred.** Knossos never deletes or overwrites user-authored content in `~/.claude/`.
6. **Advisory enforcement.** Scope violations produce warnings, never errors or blocks.

---

## What This Does NOT Cover

- **Specific scope annotations per command**: Will be proposed during implementation (Step 5). The defaults from the swarm analysis are a starting point.
- **Runtime enforcement**: Explicitly rejected. Scope is a sync/materialize-time concern.
- **SCOPE_EXCLUSIONS.json**: Removed from design. Advisory model doesn't need a machine-readable exclusion list.
- **Hook-based enforcement**: Explicitly rejected. Wrong architectural layer.
- **CLAUDE.md inscription for scope**: Explicitly rejected. Claude Code uses filesystem discovery, not instruction-based filtering.
- **Manifest version migration logic**: Rejected in favor of wipe-and-resync.

---

## Key Files Affected

| File | Change |
|------|--------|
| `internal/materialize/frontmatter.go` | Add `Scope` field, `MenaScope` type, validation |
| `internal/materialize/materialize.go` | Extension stripping in `materializeMena()`, scope filtering |
| `internal/usersync/usersync.go` | Call materialize projection, scope filtering, manifest wipe logic |
| `internal/usersync/collision.go` | Update for renamed source directories |
| `internal/config/` or `internal/paths/` | Update path constants for renamed directories |
| `internal/cmd/sync/user.go` | Update resource type source paths |
| `agents/` (renamed from `user-agents/`) | Directory move |
| `hooks/` (renamed from `user-hooks/`) | Directory move |
| `mena/` | Absorb content from `user-skills/` as `.lego.md` |
| `internal/materialize/routing_test.go` | Scope-aware routing tests |
| `internal/usersync/*_test.go` | Structure parity tests, manifest wipe tests |
| `docs/decisions/ADR-00XX-*.md` | New ADR |
