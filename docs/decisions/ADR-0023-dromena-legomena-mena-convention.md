# ADR-0023: Dromena/Legomena -- The Mena Convention

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-02-06 |
| **Deciders** | Architecture Team |
| **Supersedes** | ADR-0021 Section on invokable/non-invokable routing (refined, not replaced) |
| **Superseded by** | N/A |

## Context

The Knossos platform uses Greek mythology as its design language. Content routing between
invokable commands (projected to `.claude/commands/`) and reference knowledge (projected to
`.claude/skills/`) was previously controlled by YAML frontmatter fields (`invokable: true/false`
and `category: reference`). This created several problems:

1. **Metadata duplication**: The routing intent was declared in frontmatter AND had to be read
   at materialization time, creating coupling between content and infrastructure.
2. **Opaque convention**: Looking at a file's frontmatter required reading inside the file to
   determine its routing behavior.
3. **Naming inconsistency**: The terms "invokable" and "reference" didn't align with the
   Greek mythology design language used throughout Knossos.

Greek religious rites consist of two halves:
- **Dromena** (dromena, "things enacted"): ritual actions performed
- **Legomena** (legomena, "things spoken"): sacred words recited

These map perfectly to our two content types:
- Dromena = invokable commands (actions users perform)
- Legomena = reference knowledge (information agents consult)

## Decision

### File Extension Convention

Replace frontmatter-based routing with a filesystem convention using double extensions:

- `INDEX.dro.md` -- Dromena (invokable), projects to `.claude/commands/`
- `INDEX.lego.md` -- Legomena (reference), projects to `.claude/skills/`
- Standalone files: `commit.dro.md`, `standards.lego.md`
- Sub-files remain plain `.md` (only the entry point declares type)

### Source Directory

- `user-commands/` -> `mena/` (shared root of dromena + legomena, sounds like "mana")
- `rites/*/commands/` -> `rites/*/mena/`
- `teams/*/commands/` -> `teams/*/mena/`

### Manifest Schema

- `commands:` -> split into `dromena:` + `legomena:` in rite manifest.yaml files
- Backward compatibility: `commands:` field still parsed for older manifests

### Frontmatter Simplification

- `invokable` field: **Removed** (extension is the type declaration)
- `category` field: **Removed** (directory structure handles categorization)
- `CommandFrontmatter` Go struct -> `MenaFrontmatter`

### Migration Approach

Big-bang atomic migration in a single PR, consistent with ADR-0021 approach.

## Consequences

### Positive

1. **File type is visible from filename without reading content.** A `ls` or `find` reveals the routing intent immediately, eliminating the need to parse frontmatter for routing decisions.
2. **Aligns with Greek mythology design language (mythology gradient L3).** Dromena and legomena are historically accurate terms for the two halves of Greek rites, deepening the platform's mythological coherence.
3. **Simpler frontmatter (fewer fields to validate).** Removing `invokable` and `category` reduces the validation surface and eliminates a class of data integrity errors where frontmatter routing fields contradict directory placement.
4. **`DetectMenaType()` function replaces frontmatter parsing for routing.** File extension parsing is simpler, faster, and less error-prone than YAML frontmatter parsing for routing decisions.

### Negative

1. **Unfamiliar `.dro.md`/`.lego.md` extensions.** Mitigated: double extension preserves `.md` for editor syntax highlighting, and the convention is documented in this ADR and the doctrine.
2. **Learning curve for new contributors.** Mitigated: clear documentation, this ADR, and the doctrine concordance provide onboarding material.
3. **Large migration PR.** Mitigated: automated, well-tested, consistent with ADR-0021's atomic migration precedent.

### Neutral

1. **Projection targets unchanged.** `.claude/commands/` and `.claude/skills/` remain the projection destinations. Claude Code platform sees no difference (still reads markdown files).
2. **No behavioral change for users or agents.** Slash commands still work via `/name`. Reference content still loads via Skill tool.

## Implementation

### Components Modified

| File | Change |
|------|--------|
| `internal/materialize/frontmatter.go` | `CommandFrontmatter` -> `MenaFrontmatter`. Remove `invokable` and `category` fields. Add `DetectMenaType()` based on file extension. |
| `internal/materialize/materialize.go` | `materializeCommands()` -> `materializeMena()`. Route based on `DetectMenaType()` instead of `IsInvokable()`. |
| `rites/*/manifest.yaml` | `commands:` split into `dromena:` + `legomena:`. Backward compat: `commands:` still parsed. |
| `rites/*/commands/` | Renamed to `rites/*/mena/`. Entry points renamed with `.dro.md`/`.lego.md` extensions. |
| `user-commands/` | Renamed to `mena/`. Entry points renamed with `.dro.md`/`.lego.md` extensions. |
| `knossos/templates/sections/commands.md.tpl` | Updated to reference `mena/` and explain dromena/legomena. |

### Migration Execution

The migration follows the ADR-0021 precedent of atomic execution:

1. Rename `user-commands/` to `mena/`
2. Rename `rites/*/commands/` to `rites/*/mena/`
3. Rename entry point `INDEX.md` files to `INDEX.dro.md` or `INDEX.lego.md` based on current `invokable` frontmatter value
4. Remove `invokable` and `category` fields from all frontmatter
5. Update `manifest.yaml` files to use `dromena:`/`legomena:` fields
6. Update materialization engine to use `DetectMenaType()`
7. Update inscription templates

## Alternatives Considered

### Alternative 1: Keep Frontmatter-Based Routing with Mythology Naming

Rename `invokable: true` to `dromena: true` in frontmatter but keep the routing mechanism identical. This was rejected because it addresses only the naming inconsistency (problem 3) without solving the metadata duplication (problem 1) or opacity (problem 2).

### Alternative 2: Directory-Based Routing (dromena/ and legomena/ subdirectories)

Create separate `dromena/` and `legomena/` subdirectories instead of using file extensions. This was rejected because it recreates the original skills/commands split that ADR-0021 unified -- two directory trees with the same maintenance overhead.

### Alternative 3: Prefix Convention (dro-commit.md, lego-standards.md)

Use filename prefixes instead of double extensions. This was rejected because prefixes change the command name (the prefix would need to be stripped during projection), while double extensions naturally decompose (`commit.dro.md` -> command name is `commit`, type is `dro`).

## Related Decisions

- **ADR-0021**: Two-Axis Context Model (this ADR refines ADR-0021's routing mechanism from frontmatter to filesystem convention)
- **ADR-0014**: ari CLI Sync (materialization infrastructure this decision modifies)
- **ADR-0009**: Knossos-Roster Identity (SOURCE/PROJECTION model referenced here)

## References

| Reference | Location |
|-----------|----------|
| ADR-0021 | `docs/decisions/ADR-0021-two-axis-context-model.md` |
| Frontmatter Implementation | `internal/materialize/frontmatter.go` |
| Materialization Engine | `internal/materialize/materialize.go` |
| Rite Manifests | `rites/*/manifest.yaml` |
| Inscription Templates | `knossos/templates/sections/*.md.tpl` |
| Knossos Doctrine | `docs/philosophy/knossos-doctrine.md` |

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-06 | Claude Code (Context Architect) | Initial acceptance -- the Mena convention |
