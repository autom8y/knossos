# ADR-0021: Two-Axis Context Model (Skills and Commands Unification)

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-02-05 |
| **Deciders** | Architecture Team |
| **Supersedes** | Implicit skills/commands separation (pre-ADR) |
| **Superseded by** | N/A |

## Context

Claude Code released a change merging slash commands and skills into a unified invocation model: both are discovered via the same "Available skills" list and both are invoked via the Skill tool. While Claude Code described this as "no change in behavior," it eliminated the platform-level distinction between commands and skills. Knossos had maintained a structural separation:

- `user-commands/` (projected to `.claude/commands/`) -- user-callable actions
- `user-skills/` (projected to `.claude/skills/`) -- reference documentation and agent knowledge

This separation created concrete problems:

1. **Duplication**: The `/start` command required ~500 lines across 5+ files (the command itself plus `start-ref/` skill with behavior, examples, and integration docs). This pattern repeated for 14 command/skill pairs.
2. **Two mental models**: Contributors had to decide whether new content was a "command" or a "skill," a distinction Knossos imposed but Claude Code no longer enforced.
3. **Maintenance burden**: 35+ commands and 90+ skill files across separate directory trees, with drift between paired content.
4. **Token waste**: The `skills` CLAUDE.md section and `commands` CLAUDE.md section consumed context budget directing the agent to two different directories for the same conceptual purpose.

### Forces

- **Simplification**: Fewer concepts reduces cognitive load for both the agent and human contributors.
- **Token economics**: CLAUDE.md (L0 context) is loaded every turn; minimizing its surface area directly reduces cost.
- **Progressive disclosure**: Large reference modules (e.g., `session-common` at 3,189 lines) must remain loadable piecewise, not flattened into a single file.
- **Backward compatibility**: 12 rite manifests, the materialization engine, and inscription templates all reference the old structure.
- **Claude Code trajectory**: The upstream platform is converging on a single namespace; Knossos should align rather than fight.

### Prior Art

- **SPIKE**: `docs/spikes/SPIKE-skills-commands-unification.md` -- explored options, recommended full unification
- **TDD**: `docs/design/TDD-skills-commands-unification.md` -- specified frontmatter schema, materialization changes, migration plan

## Decision

Adopt a **Two-Axis Context Model** that replaces the skills/commands split with two orthogonal classification axes.

### Axis 1: Invocation Model

Every piece of content in the system is a "command" stored under a unified source directory. The `invokable` field in YAML frontmatter determines how it is projected and discovered:

| Value | Behavior | Projection Target | Discovery |
|-------|----------|--------------------|-----------|
| `invokable: true` (default) | User-callable via `/name` syntax | `.claude/commands/{name}/` | Appears as slash command in Claude Code |
| `invokable: false` | Agent-initiated via Skill tool | `.claude/skills/{name}/INDEX.md` | Loaded on-demand when agents need domain knowledge |

The discriminator is the `invokable` field parsed by `CommandFrontmatter.IsInvokable()` in `internal/materialize/frontmatter.go`. When `invokable` is absent from frontmatter, it defaults to `true`, preserving backward compatibility with legacy command files that predate this schema.

**Unified frontmatter schema** (`CommandFrontmatter` struct):

```yaml
---
name: string              # Required. Identifier matching filename.
description: string        # Required. Human-readable description.
invokable: boolean         # Optional. Default: true. User-callable via /name.
argument-hint: string      # Optional. Usage pattern (invokable=true only).
triggers: string[]         # Optional. Auto-invocation keywords.
allowed-tools: string[]    # Optional. Tool restrictions (invokable=true only).
model: string              # Optional. Model selection (invokable=true only).
category: enum             # Conditional. Required when invokable=false.
                           # Values: reference | template | schema
version: string            # Optional. Semantic version.
deprecated: boolean        # Optional. Default: false.
deprecated-by: string      # Optional. Replacement command reference.
---
```

Validation rules (enforced by `CommandFrontmatter.Validate()`):
- `name` and `description` are required for all commands.
- `category` is required when `invokable: false`; valid values are `reference`, `template`, `schema`.
- `argument-hint`, `allowed-tools`, and `model` are semantically meaningful only for invokable commands but are not rejected on non-invokable commands to avoid brittle validation.

### Axis 2: Context Tier

Content is organized into four tiers by loading frequency and token cost:

| Tier | Name | Location | Loading Pattern | Budget |
|------|------|----------|-----------------|--------|
| L0 | Always | `CLAUDE.md` | Every turn (auto-injected) | <250 lines |
| L1 | Operational | `ari --help` | Via Bash tool when procedures needed | Unbounded |
| L2 | Domain | `.claude/skills/` and `.claude/commands/` | Via Skill tool on-demand | Per-command |
| L3 | Reference | `docs/` | Via Read tool, loaded rarely | Unbounded |

L0 contains navigation pointers only -- it tells the agent where to find knowledge, not the knowledge itself. L2 is where the invocation model axis intersects: invokable commands and reference skills both live at L2 but are discovered differently (slash syntax vs. Skill tool).

### Three-Tier Ownership

Source content lives in three ownership tiers with different lifecycle rules:

| Tier | Source Location | Lifecycle | Availability |
|------|----------------|-----------|--------------|
| Rite-scoped | `rites/{rite-name}/commands/` | Active only when rite is materialized | Rite-dependent |
| Shared | `rites/shared/commands/` | Always materialized via `shared` dependency | All rites |
| User-level | `user-commands/{domain}/{name}/` | Always available, lowest priority in override chain | All rites |

**Override precedence** (highest to lowest): current rite > dependency rites > shared > user-level. This is implemented in `materializeCommands()` at `internal/materialize/materialize.go:446`.

### Projection Routing

The materializer reads from unified source directories and routes to projection targets based on the frontmatter discriminator:

```
SOURCE                                    PROJECTION
rites/*/commands/{name}/INDEX.md    -->   .claude/commands/{name}/ (if invokable)
                                    -->   .claude/skills/{name}/  (if not invokable)
user-commands/{domain}/{name}/      -->   .claude/commands/{name}/ (if invokable)
                                    -->   .claude/skills/{name}/  (if not invokable)
```

The routing logic resides in `materializeCommands()` (`internal/materialize/materialize.go:446`). The deprecated `materializeSkills()` function is retained at line 495 for backward compatibility with legacy manifests that still use the `skills:` field.

### Rite Manifest Schema

Rite manifests use `commands:` instead of `skills:`:

```yaml
# rites/10x-dev/manifest.yaml
name: 10x-dev
version: "1.0.0"
commands:
  - 10x-ref           # invokable: false (reference)
  - 10x-workflow       # invokable: false (reference)
  - architect-ref      # invokable: false (reference)
  - build-ref          # invokable: false (reference)
  - doc-artifacts      # invokable: false (template)
```

The `RiteManifest` struct in `materialize.go` retains a deprecated `Skills []string` field for backward compatibility. When a manifest has `skills:` but no `commands:`, the materializer falls back to legacy behavior.

### Directory Structure (Post-Migration)

```
user-commands/                          # User-level commands (always available)
  session/start.md                      # invokable: true
  session/start/                        # Progressive disclosure for /start
    behavior.md
    examples.md
  session/common/INDEX.md               # invokable: false, category: reference
  operations/commit/INDEX.md            # invokable: true
  guidance/prompting/INDEX.md           # invokable: false, category: reference
  templates/doc-artifacts/INDEX.md      # invokable: false, category: template

rites/10x-dev/commands/                 # Rite-scoped commands
  10x-ref/INDEX.md                      # invokable: false
  10x-workflow/INDEX.md                 # invokable: false

rites/shared/commands/                  # Shared commands (all rites)
  cross-rite-handoff/INDEX.md           # invokable: false
  smell-detection/INDEX.md              # invokable: false
```

### Inscription Template

The `skills` CLAUDE.md section is replaced by `commands` (`knossos/templates/sections/commands.md.tpl`):

```
## Commands

Commands are invoked via the **Skill tool**. Two types exist:

- **Invokable** (`/name`): User-callable actions like `/start`, `/commit`, `/pr`
- **Reference** (auto-loaded): Patterns and templates like `prompting`, `doc-artifacts`

See `.claude/commands/` for the full list.
```

## Consequences

### Positive

1. **Single source of truth**: All content lives in `commands/` directories. No more deciding whether something is a "skill" or a "command."
2. **Reduced token cost**: One CLAUDE.md section (`commands`) instead of two (`skills` + `commands`), saving ~5-10 lines of L0 context per turn.
3. **Aligned with Claude Code direction**: The upstream platform is converging on a single namespace. Knossos now mirrors that trajectory rather than maintaining a divergent abstraction.
4. **Progressive disclosure preserved**: Reference modules retain their subdirectory structure (INDEX.md + detail files). The `session-common` module's 3,189 lines across 8 files loads piecewise, not monolithically.
5. **Clean migration path**: The `invokable` field defaults to `true`, so legacy command files without the field continue to work without modification.
6. **Simplified rite manifests**: One `commands:` field replaces `skills:`. The names of the entries are unchanged; only the field name and source directory changed.
7. **383 files migrated atomically**: The refactoring was executed as a clean break (not phased), eliminating any dual-state transition period.

### Negative

1. **Projection split remains**: Despite unified sources, the materializer still writes to two projection directories (`.claude/commands/` and `.claude/skills/`) because Claude Code's discovery mechanisms differ for slash commands vs. Skill tool content. This split is a consequence of Claude Code's current behavior, not an architectural choice.
2. **Backward compatibility overhead**: The `RiteManifest` struct carries a deprecated `Skills` field, and `materializeSkills()` is retained as dead code for legacy fallback. This should be removed once all known consumers have migrated.
3. **Frontmatter required for routing**: Every INDEX.md must have valid YAML frontmatter with the `invokable` field for correct projection routing. Files without frontmatter default to invokable, which is wrong for reference content. This creates a data integrity requirement that did not exist when directory location alone determined behavior.

### Neutral

1. **No user-facing behavior change**: Slash commands still work via `/name`. Reference content still loads via Skill tool. The change is entirely in source organization and internal routing.
2. **Rite manifest field rename**: `skills:` to `commands:` is a mechanical change. The listed names are identical.
3. **Template rename**: `skills.md.tpl` to `commands.md.tpl` in `knossos/templates/sections/`. The section ID in CLAUDE.md changes from `skills` to `commands`.

## Alternatives Considered

### Alternative 1: Keep Separate Directories, Unify Projection

Maintain `user-skills/` and `user-commands/` as separate source directories but project both to `.claude/commands/`. This was rejected because it preserved the cognitive overhead of two source trees while solving only the projection problem. The duplication between paired commands and skills (e.g., `/start` + `start-ref/`) would persist.

### Alternative 2: Flatten Everything into `.claude/commands/`

Eliminate `.claude/skills/` entirely and project all content to `.claude/commands/`. Reference content would use a naming convention (e.g., `_ref-` prefix) instead of a separate directory. This was rejected because Claude Code's Skill tool and slash command discovery behave differently. Placing non-invokable reference content in `.claude/commands/` would surface it as slash commands, polluting the user's command palette with entries like `/session-common` that have no meaningful user action.

### Alternative 3: Phased Migration with Symlinks

Maintain both `user-skills/` and `user-commands/` during a transition period, using symlinks from the old locations to the new. This was rejected because the dual-state period creates confusion about which location is canonical, symlinks interact poorly with git across platforms, and the migration scope (383 files) was small enough to execute atomically.

## Implementation

### Components Modified

| File | Change |
|------|--------|
| `internal/materialize/frontmatter.go` | New file. `CommandFrontmatter` struct, `IsInvokable()`, `Validate()`, `ParseCommandFrontmatter()` |
| `internal/materialize/materialize.go` | `materializeCommands()` at line 446. `RiteManifest.Commands` field. Deprecated `materializeSkills()` retained at line 495 |
| `internal/materialize/source.go` | `SourceResolver` unchanged (resolves by rite name, agnostic to skills/commands) |
| `rites/*/manifest.yaml` | All 12 manifests: `skills:` field renamed to `commands:` |
| `rites/*/skills/` | All rite directories renamed from `skills/` to `commands/` |
| `knossos/templates/sections/commands.md.tpl` | New template replacing `skills.md.tpl` |
| `knossos/templates/sections/skills.md.tpl` | Deleted |
| `user-skills/` | Deleted. Content migrated to `user-commands/` |
| `user-commands/guidance/` | New domain for reference content (prompting, standards, etc.) |
| `user-commands/templates/` | New domain for template content (doc-artifacts, justfile, etc.) |
| `user-commands/session/common/` | Migrated from `user-skills/session-lifecycle/session-common/` |

### Migration Execution

The migration was executed as a single atomic operation (383 files):

1. Created new domains (`guidance/`, `templates/`) in `user-commands/`
2. Merged `-ref` skills into command progressive disclosure directories (e.g., `start-ref/` into `session/start/`)
3. Moved library skills to appropriate domains with `invokable: false` frontmatter
4. Renamed `rites/*/skills/` to `rites/*/commands/` across all 12 rites
5. Renamed `SKILL.md` files to `INDEX.md`
6. Updated all `manifest.yaml` files: `skills:` to `commands:`
7. Updated inscription templates
8. Deleted `user-skills/` and `.claude/skills/` legacy directories

No phased rollout or backward-compatibility shim period was used.

## References

- **Spike**: `docs/spikes/SPIKE-skills-commands-unification.md`
- **TDD**: `docs/design/TDD-skills-commands-unification.md`
- **ADR-0009**: Knossos-Roster Identity (SOURCE/PROJECTION model referenced here)
- **ADR-0014**: ari CLI Sync (materialization infrastructure this decision extends)
- `internal/materialize/frontmatter.go`: `CommandFrontmatter` struct and `IsInvokable()` discriminator
- `internal/materialize/materialize.go`: `materializeCommands()` projection routing

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-05 | Claude Code (Context Architect) | Initial acceptance -- documenting completed migration |
