# ADR-0025: MenaScope Filtering -- Pipeline-Targeted Mena Distribution

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-02-07 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A (new capability) |
| **Superseded by** | N/A |

## Context

Knossos distributes mena content (commands and skills) through two independent pipelines:

1. **Materialize** (`ari rite start`, `ari sync materialize`) -- projects rite-level and shared mena into `.claude/commands/` and `.claude/skills/` within a project directory.
2. **Usersync** (`ari sync user mena`) -- projects distribution-level mena into `~/.claude/commands/` and `~/.claude/skills/` in the user's home directory.

PR1 (commit `1f9d677`) unified the mena distribution API with `ProjectMena()`, `MenaFrontmatter`, extension stripping, and dual-target routing. After PR1, every distribution-level mena entry is eligible for both pipelines. There is no mechanism to restrict an entry to a single pipeline.

This creates two problems:

- **Unwanted duplication**: Content that only makes sense at the user level (cross-rite guidance, rite discovery instructions) gets projected into project directories when materialize runs, polluting project-level `.claude/` with irrelevant content.
- **Missing exclusivity**: Future mena entries have no way to declare "I am only meaningful inside a project context" or "I am only meaningful at the user level."

The scope model must be fully backward compatible: every existing mena file that lacks the new field must behave identically to today.

### Why Mena-Only

Scope is a mena-specific concern because mena is the only content type distributed through two independent pipelines. Agents are materialized project-level only (from rite definitions). Hooks are materialized project-level only (from templates). Neither agents nor hooks participate in usersync distribution today. Adding scope to agents or hooks would be premature abstraction with no current use case.

If agent or hook distribution becomes multi-pipeline in the future, the same pattern can be extended to their respective frontmatter schemas. The `MenaScope` type is defined in the `materialize` package and can be imported by other packages if needed.

### Why the Default is "Both"

The zero value of the `MenaScope` type (empty string) means "distribute to both pipelines." This is the only backward-compatible default. Today, every mena entry goes everywhere. Changing the default to anything else would require annotating all 46 existing mena files, and missing any annotation would silently drop content from a pipeline.

## Decision

### MenaScope Type

Define `MenaScope` as a Go string type with three valid values:

| Value | Constant | Meaning |
|-------|----------|---------|
| `""` | `MenaScopeBoth` | Included in both pipelines (zero value, backward compat) |
| `"user"` | `MenaScopeUser` | Included in usersync only |
| `"project"` | `MenaScopeProject` | Included in materialize only |

String type was chosen over integer enum because: (a) YAML serialization is human-readable (`scope: user` vs. `scope: 1`), (b) the type has exactly three values with no bitwise combination semantics, and (c) the zero value maps naturally to "both."

### Frontmatter Integration

Add `Scope MenaScope` to `MenaFrontmatter` with `yaml:"scope,omitempty"`. Extend `MenaFrontmatter.Validate()` to reject unrecognized scope values with an error message that identifies the invalid value and the valid options.

### Filtering Mechanism

Add `PipelineScope MenaScope` to `MenaProjectionOptions`. Callers set this field to indicate which pipeline is running:

- `materializeMena()` sets `PipelineScope: MenaScopeProject`
- Usersync sets filtering inline via frontmatter parsing in `syncFiles()`
- The zero value (empty) means "no scope filtering" for backward compatibility

In `ProjectMena()`, scope filtering is applied in the Pass 2 routing loop after the existing type filter (dro/lego). An entry's INDEX file frontmatter is parsed to extract scope, and entries whose scope does not match the pipeline are skipped.

For usersync, scope filtering is applied inline in `syncFiles()` before manifest-key computation. The usersync pipeline is not refactored to call `ProjectMena()` because the two pipelines have fundamentally different semantics (usersync tracks per-file manifests with checksums, divergence detection, and collision checking; `ProjectMena()` is a stateless copy operation).

### Rite-Level Mena with scope: user

Rite-level mena is processed only by the materialize pipeline. If a rite-level mena entry has `scope: user`, it is excluded from materialize, which means it reaches no pipeline. This is honored (not silently ignored) with a warning to stderr: "mena X has scope: user but is only reachable by materialize (will not be distributed)."

Honoring the field uniformly avoids an inconsistency where the same frontmatter field has different semantics depending on the file's location. The warning catches what is almost certainly an authoring mistake.

## Consequences

### Positive

1. **Targeted distribution**: Mena authors can restrict content to the appropriate pipeline, eliminating unwanted duplication in project directories.
2. **Full backward compatibility**: Every existing mena file continues to behave identically. No migration required.
3. **Extensible pattern**: The same scope model can be applied to agents or hooks if they gain multi-pipeline distribution in the future.
4. **Fail-safe defaults**: Missing scope, missing frontmatter, and malformed frontmatter all default to "both" -- content is never silently dropped.
5. **Minimal I/O impact**: Frontmatter parsing is guarded by a zero-value check. When no pipeline scope is set, no additional file reads occur.

### Negative

1. **Two filtering locations**: Scope filtering is implemented in both `ProjectMena()` and `syncFiles()` rather than a single centralized location. This is a consequence of the two pipelines having different architectures (stateless copy vs. manifest-tracked sync). The frontmatter parsing helpers are shared, so the actual parsing logic is not duplicated.
2. **Frontmatter parsing adds I/O to ProjectMena()**: Each collected entry's INDEX file is read during Pass 2. For the current mena set (~55 entries), this adds <5ms. Guarded by the zero-value check so callers that do not set PipelineScope pay nothing.

### Neutral

1. **No changes to agent or hook systems**: Scope is a mena-only concept in this iteration.
2. **No annotation of existing files required**: The mechanism is available immediately; annotating specific files with `scope: user` or `scope: project` is optional and can happen incrementally.
3. **Existing tests pass without modification**: The zero-value default means all existing `ProjectMena()` and usersync test cases continue to work with no scope filtering applied.

## Alternatives Considered

### Alternative 1: Bitmask Scope

Define `MenaScope` as an integer with bit flags (`ScopeUser = 1`, `ScopeProject = 2`, `ScopeAll = 3`). Entries could be assigned to arbitrary pipeline combinations via bitwise OR.

**Rejected**: There are exactly two pipelines and three valid states (user, project, both). A bitmask adds type complexity (YAML serialization, validation of bit combinations) with no expressiveness benefit. If a third pipeline were added, the string type can be extended to a new value or the model can be reconsidered. This is a reversible decision.

### Alternative 2: Implicit Scope from Directory Structure

Infer scope from the mena file's location: files in `mena/user/` are user-scoped, files in `mena/project/` are project-scoped, files elsewhere are both.

**Rejected**: This requires restructuring the existing `mena/` directory layout, which would break all existing paths and references. It also conflates organizational grouping (the current `mena/` directory structure groups by functional category like `guidance/`, `operations/`, `navigation/`) with distribution semantics. Frontmatter-based scope keeps the two concerns orthogonal.

### Alternative 3: Scope Arrays

Support `scope: [user, project]` as an array of target pipelines. This would make "both" explicit rather than implicit.

**Rejected**: An array of two possible values has four states (`[]`, `[user]`, `[project]`, `[user, project]`), and three of them (`[]`, `[user, project]`, and missing) would all need to mean "both." This adds parsing complexity, validation rules for degenerate cases, and YAML verbosity with no practical benefit over the simple string model.

### Alternative 4: Refactor Usersync to Call ProjectMena()

Centralize scope filtering by having usersync call `ProjectMena()` with additive mode and `PipelineScope: MenaScopeUser`, rather than implementing inline filtering in `syncFiles()`.

**Rejected for PR2**: Usersync's `syncFiles()` manages a per-file manifest with checksums, source tracking (knossos/diverged/user), collision detection, and recovery mode. `ProjectMena()` is a stateless copy operation with no manifest concept. Refactoring usersync to delegate to `ProjectMena()` would require either: (a) adding manifest tracking to `ProjectMena()` (which conflates two concerns), or (b) running `ProjectMena()` first and then reconciling its output with the manifest (which adds complexity for no benefit). The inline approach adds ~15 lines of code and reuses the shared frontmatter parsing helpers.

## Related Decisions

- **ADR-0023**: Dromena/Legomena Mena Convention (established `MenaFrontmatter` schema and the dro/lego type system that scope filtering complements)
- **ADR-0021**: Two-Axis Context Model (unified commands/skills model that `MenaScope` extends with pipeline targeting)
- **ADR-0024**: Agent Factory (established the pattern of structured frontmatter with validation that this decision follows for `MenaFrontmatter`)
- **ADR-0014**: ari CLI Sync (materialization infrastructure that scope filtering integrates into)

## References

| Reference | Location |
|-----------|----------|
| PRD | `docs/prd/PRD-mena-scope.md` |
| TDD | `docs/tdd/TDD-mena-scope.md` |
| MenaFrontmatter | `internal/materialize/frontmatter.go` |
| ProjectMena | `internal/materialize/project_mena.go` |
| Usersync | `internal/usersync/usersync.go` |
| PR1 (unified mena distribution) | Commit `1f9d677` |

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-07 | Architect (Claude Code) | Initial acceptance -- MenaScope filtering design |
