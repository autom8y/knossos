# PRD: MenaScope Filtering

```yaml
status: draft
impact: low
impact_categories: []
author: requirements-analyst
date: 2026-02-07
initiative: mena-scope
pr_sequence: "PR2 of Mena Scope Initiative"
predecessor: "PR1 (1f9d677) - Unified mena distribution API"
```

## Problem Statement

Knossos distributes mena content (commands and skills) through two independent pipelines:

1. **Materialize** (`ari rite start`, `ari sync materialize`) -- projects rite-level and shared mena into `.claude/commands/` and `.claude/skills/` within a project directory.
2. **Usersync** (`ari sync user mena`) -- projects distribution-level mena into `~/.claude/commands/` and `~/.claude/skills/` in the user's home directory.

Currently, every distribution-level mena file (`mena/`) is eligible for both pipelines. There is no mechanism to restrict a mena entry to a single pipeline. This creates two problems:

- **Unwanted duplication**: Content that only makes sense at the user level (e.g., keybindings-help, cross-rite guidance) gets projected into project directories when materialize runs, polluting project-level `.claude/` with irrelevant commands/skills.
- **Missing exclusivity**: There is no way for a future mena entry to declare "I am only meaningful inside a project context" or "I am only meaningful at the user level" -- every entry goes everywhere.

PR1 (commit `1f9d677`) unified the mena distribution API with `ProjectMena()`, `MenaFrontmatter`, and extension stripping. PR2 adds a `scope` field to `MenaFrontmatter` so mena entries can declare their intended distribution target, and both pipelines filter accordingly.

## Scope Model

### MenaScope Type

A new string type `MenaScope` with three valid states:

| Value | Meaning | Materialize includes? | Usersync includes? |
|-------|---------|----------------------|-------------------|
| `"user"` | User-level only | No | Yes |
| `"project"` | Project-level only | Yes | No |
| `""` (absent) | Both pipelines | Yes | Yes |

The empty/absent case is the default. This preserves full backward compatibility -- every existing mena file that lacks a `scope` field continues to behave identically to today.

### Frontmatter Field

```yaml
scope: user    # or "project", or omitted for both
```

Added to `MenaFrontmatter` as:

```go
Scope string `yaml:"scope,omitempty"`
```

### Validation Rules

- **Known values only**: `"user"`, `"project"`, or empty string. Any other value (e.g., `"global"`, `"both"`, `"USER"`) must fail validation with a clear error message identifying the file and the invalid value.
- **Case-sensitive**: `"User"` is invalid; only lowercase `"user"` and `"project"` are accepted.
- **Validation location**: Inside `MenaFrontmatter.Validate()`, alongside existing `name` and `description` checks.

## User Stories

### US-1: Mena author restricts content to user pipeline

**As** a mena content author, **I want** to mark a mena entry with `scope: user` **so that** it is distributed only to `~/.claude/` via usersync and excluded from project-level materialization.

**Acceptance criteria**:
- A mena INDEX file with `scope: user` in its frontmatter is included when usersync runs.
- The same file is excluded when materialize runs (destructive mode).
- The file does not appear in `.claude/commands/` or `.claude/skills/` after `ari rite start`.

### US-2: Mena author restricts content to project pipeline

**As** a mena content author, **I want** to mark a mena entry with `scope: project` **so that** it is distributed only to `.claude/` via materialize and excluded from usersync.

**Acceptance criteria**:
- A mena INDEX file with `scope: project` in its frontmatter is included when materialize runs.
- The same file is excluded when usersync runs.
- The file does not appear in `~/.claude/commands/` or `~/.claude/skills/` after `ari sync user mena`.

### US-3: Unscoped mena distributed to both pipelines (backward compat)

**As** an existing mena content author, **I want** my mena entries that lack a `scope` field to continue being distributed to both pipelines **so that** nothing breaks when the scope feature is added.

**Acceptance criteria**:
- A mena INDEX file with no `scope` field is included by both materialize and usersync.
- Behavior is identical to pre-PR2 behavior for every existing mena file.

### US-4: Invalid scope value rejected at validation

**As** a developer building or validating mena content, **I want** an invalid `scope` value to produce a clear error **so that** typos and misunderstandings are caught early.

**Acceptance criteria**:
- `MenaFrontmatter.Validate()` returns an error when `scope` is set to an unrecognized value.
- Error message includes the invalid value and the list of valid options.
- Build/test commands surface this error clearly.

## Functional Requirements

### FR-1: MenaScope type and constants [MUST]

Define `MenaScope` as a string type in `internal/materialize/frontmatter.go` with constants:
- `MenaScopeUser MenaScope = "user"`
- `MenaScopeProject MenaScope = "project"`
- `MenaScopeBoth MenaScope = ""` (the zero value)

### FR-2: Scope field on MenaFrontmatter [MUST]

Add `Scope MenaScope` to `MenaFrontmatter` with YAML tag `yaml:"scope,omitempty"`.

### FR-3: Scope validation in Validate() [MUST]

Extend `MenaFrontmatter.Validate()` to reject unknown scope values. Valid: `""`, `"user"`, `"project"`. All others produce an error of the form: `frontmatter: invalid scope "VALUE" (must be "user", "project", or omitted)`.

### FR-4: Scope filtering in ProjectMena() [MUST]

`ProjectMena()` must accept a scope context that indicates which pipeline is calling. Two integration points:

- **Materialize caller** passes `MenaScopeProject` (or equivalent indicator). Entries with `scope: user` are skipped.
- **Usersync caller** passes `MenaScopeUser` (or equivalent indicator). Entries with `scope: project` are skipped.
- **Unscoped entries** (`scope: ""`) are always included.

Implementation approach: add a `PipelineScope MenaScope` field to `MenaProjectionOptions`. When set, `ProjectMena()` reads the INDEX file's frontmatter for each collected entry and skips entries whose scope excludes the current pipeline.

### FR-5: Scope filtering in usersync syncFiles() [MUST]

`syncFiles()` in `internal/usersync/usersync.go` walks source files directly (does not call `ProjectMena()`). It must also read frontmatter from INDEX files and skip entries whose scope is `"project"`.

Two approaches (architect decides):
- **(A)** Refactor usersync to call `ProjectMena()` with additive mode and `PipelineScope: "user"`. This centralizes filtering in one place.
- **(B)** Add frontmatter parsing in `syncFiles()` directly. When encountering an INDEX file, parse its frontmatter and check scope before syncing.

Recommendation: Approach (A) is cleaner but may require more refactoring. Approach (B) is lower-risk for PR2. Architect decides.

### FR-6: Standalone file scope [SHOULD]

Standalone mena files (files directly in grouping directories, not in leaf INDEX directories) also need scope filtering. For standalone files, scope is read from the file's own frontmatter, since there is no separate INDEX file.

### FR-7: Scope on MenaProjectionOptions [MUST]

Add `PipelineScope MenaScope` to `MenaProjectionOptions`. Callers set this to indicate which pipeline is running:
- `materializeMena()` sets `PipelineScope: MenaScopeProject`
- Usersync sets `PipelineScope: MenaScopeUser` (if approach A from FR-5)
- Empty/zero value means "no filtering" (backward compat for any callers that do not set it)

### FR-8: ADR-0025 [SHOULD]

Document the scope decision in `docs/decisions/ADR-0025-mena-scope.md`. Key points:
- Why scope is mena-only (agents/hooks deferred)
- Why the default is "both" (backward compat)
- Why rite-level mena does not need scope (implicitly project-scoped by pipeline)
- Rejected alternative: bitmask scope (over-engineering for two pipelines)

## Non-Functional Requirements

### NFR-1: Backward compatibility [MUST]

Every existing mena file that lacks a `scope` field must behave identically to the current behavior. Zero behavior changes for unscoped content. Verified by running existing tests without modification.

### NFR-2: Performance [SHOULD]

Frontmatter parsing adds I/O (reading INDEX files during collection). This must not materially slow down `ari rite start` or `ari sync user mena`. Target: less than 50ms additional latency for the full mena set (25 distribution-level + 30 rite-level entries).

### NFR-3: Error clarity [MUST]

Invalid scope values produce errors that identify: (a) the file path, (b) the invalid value, (c) the valid options. No silent failures -- an unrecognized scope value must not be treated as "both".

### NFR-4: Testability [MUST]

All scope filtering logic must be unit-testable via `ProjectMena()` with in-memory or temp-dir sources. No dependency on live knossos home or active rites.

## Edge Cases

### EC-1: Rite-level mena with scope field

Rite-level mena files (`rites/*/mena/`) are only ever processed by materialize. They are inherently project-scoped. If a rite-level mena file includes `scope: user`, the behavior should be:

**Decision needed from architect**: Two options:
- **(A) Ignore scope on rite-level mena** -- rite-level sources are always included in materialize regardless of scope. Simpler, but `scope: user` on rite-level mena is a silent no-op.
- **(B) Honor scope on rite-level mena** -- a rite-level file with `scope: user` would be excluded from materialize, which means it goes nowhere (usersync never sees rite-level mena). This is probably an authoring mistake, but honoring the field is consistent.

**Recommendation**: Option (B) with a warning. If a rite-level mena entry has `scope: user`, log a warning ("rite-level mena with scope: user will not be distributed by any pipeline") and skip it. This catches authoring mistakes.

### EC-2: Distribution-level mena with scope: project

A distribution-level mena file (`mena/`) with `scope: project` is included by materialize but excluded by usersync. If a rite-level mena file exists with the same name, the rite-level file takes priority (existing override behavior). No special handling needed -- this just means the distribution-level entry only contributes when no rite-level override exists.

### EC-3: Standalone file with no frontmatter

Standalone mena files (e.g., `mena/rite-switching/10x.dro.md`) may or may not have frontmatter. If a standalone file has no frontmatter at all (no `---` delimiters), its scope is treated as empty (both pipelines). Parsing failure for missing frontmatter is not an error -- it simply means "unscoped."

### EC-4: INDEX file with scope, companion files without

A mena leaf directory has an INDEX file and possibly companion files (sub-pages referenced from INDEX). Scope applies at the directory level, determined by the INDEX file's frontmatter. Companion files inherit the INDEX's scope -- they are not individually checked.

### EC-5: Embedded FS mena sources

`ProjectMena()` supports both filesystem and embedded FS (`fs.FS`) sources. Frontmatter parsing must work for both. For embedded FS sources, the INDEX file is read via `fs.ReadFile()` and parsed identically.

### EC-6: Scope collision with MenaFilter

`MenaFilter` (ProjectDro/ProjectLego/ProjectAll) controls whether dromena or legomena are projected. `MenaScope` is orthogonal -- it controls which pipeline includes the entry. Both filters are applied: an entry must pass both the type filter AND the scope filter to be projected.

### EC-7: Frontmatter parse failure on existing file

If an INDEX file exists but its YAML frontmatter is malformed (broken YAML between `---` delimiters), the behavior should be:
- **Log a warning** with the file path and parse error.
- **Treat as unscoped** (include in both pipelines). This prevents a syntax error in one file from silently dropping content.

### EC-8: MenaProjectionOptions.PipelineScope not set

If `PipelineScope` is empty (zero value), no scope filtering is applied. All entries pass regardless of their scope field. This is the backward-compatible default and ensures callers that do not know about scope are unaffected.

## INDEX Annotation Strategy

Based on stakeholder guidance, the 25 distribution-level INDEX entries and 21 standalone files should be annotated as follows:

### Remain unscoped (both pipelines) -- default

Most distribution-level mena should remain unscoped. Session commands (/start, /park, /wrap, /continue, /handoff), operations (/commit, /pr, /code-review, /qa, /spike, /build, /architect), workflow (/hotfix, /task, /sprint), navigation (/worktree, /consult, /rite, /ecosystem, /sessions), rite-switching shortcuts (/10x, /forge, /rnd, etc.), CEM (/sync), templates, and most guidance -- all should continue going to both pipelines. This is the correct default because these commands are useful both at the user level (available in any project) and at the project level (available when a rite is materialized).

### Candidates for scope: user

Content that only makes sense at the user-global level. Initial candidates (to be confirmed during implementation):
- `mena/guidance/cross-rite/` -- Cross-rite handoff protocols. Only relevant as user-level reference since rite-switching is a user-level operation.
- `mena/guidance/rite-discovery/` -- How to discover available rites. User-level concern.

**Note**: The stakeholder indicated that very few entries need `scope: user` today. The primary value of the scope field is the mechanism itself -- enabling future mena entries to be targeted. Annotating existing entries is a COULD, not a MUST.

### Candidates for scope: project

No distribution-level mena entries need `scope: project` today. Rite-level mena is implicitly project-scoped by pipeline mechanics. The `scope: project` value exists for future distribution-level mena that should only appear in project contexts.

## Work Items

### WI-1: MenaScope type + frontmatter field + validation

**File**: `internal/materialize/frontmatter.go`

- Define `MenaScope` type and constants
- Add `Scope MenaScope` field to `MenaFrontmatter`
- Extend `Validate()` to reject unknown scope values

**Tests**: `internal/materialize/frontmatter_test.go`
- Test valid scope values parse correctly
- Test invalid scope value rejected with clear error
- Test missing scope defaults to empty (both)

**Acceptance**: `CGO_ENABLED=0 go test ./internal/materialize/ -run TestMenaFrontmatter`

### WI-2: PipelineScope on MenaProjectionOptions

**File**: `internal/materialize/project_mena.go`

- Add `PipelineScope MenaScope` to `MenaProjectionOptions`
- Add frontmatter parsing helper: read INDEX file from collected entry, parse YAML frontmatter, extract scope
- In `ProjectMena()` Pass 2 loop: after type filter, apply scope filter. If `opts.PipelineScope` is set and entry scope excludes the pipeline, skip.
- Same for standalone files loop.

**Tests**: `internal/materialize/project_mena_test.go`
- Test: entry with `scope: user` excluded when PipelineScope is `project`
- Test: entry with `scope: project` excluded when PipelineScope is `user`
- Test: entry with no scope included regardless of PipelineScope
- Test: PipelineScope empty means no filtering (backward compat)

**Acceptance**: `CGO_ENABLED=0 go test ./internal/materialize/ -run TestProjectMena`

### WI-3: Wire scope into materializeMena()

**File**: `internal/materialize/materialize.go`

- In `materializeMena()`, set `opts.PipelineScope = MenaScopeProject` before calling `ProjectMena()`.

**Tests**: Integration tests covering materialize with scope-annotated mena sources.

**Acceptance**: `CGO_ENABLED=0 go test ./internal/materialize/`

### WI-4: Wire scope into usersync

**File**: `internal/usersync/usersync.go` (and potentially `internal/usersync/usersync_test.go`)

- **If approach (A)**: Refactor `syncFiles()` for mena to use `ProjectMena()` with additive mode and `PipelineScope: MenaScopeUser`.
- **If approach (B)**: Add frontmatter parsing in `syncFiles()`. When processing a mena file, if it is an INDEX file, parse frontmatter and check scope. If `scope: project`, skip the entire leaf directory.

**Tests**:
- Test: mena entry with `scope: project` is excluded from usersync
- Test: mena entry with `scope: user` is included by usersync
- Test: mena entry with no scope is included by usersync (backward compat)

**Acceptance**: `CGO_ENABLED=0 go test ./internal/usersync/`

### WI-5: Annotate distribution-level mena (optional for PR2)

**Files**: Selected `mena/*/INDEX.*.md` files

- Add `scope: user` to confirmed candidates (cross-rite, rite-discovery if confirmed)
- All other files remain unscoped

**Acceptance**: Manual verification that annotated files appear/disappear from correct pipelines.

### WI-6: ADR-0025

**File**: `docs/decisions/ADR-0025-mena-scope.md`

- Document decision rationale per FR-8

**Acceptance**: File exists and follows ADR template format.

## Out of Scope

The following are explicitly excluded from this PR:

- **Agent scope**: Scope filtering for agents (`internal/agent/`, `internal/usersync/` agent sync). Deferred to a future initiative.
- **Hook scope**: Scope filtering for hooks. Deferred to a future initiative.
- **Scope inheritance/composition**: No support for `scope: [user, project]` array syntax or scope inheritance from parent directories. Two pipelines, two values, plus default.
- **Scope in rite manifests**: Rite manifests (`rite.yaml`) do not gain a scope field. Scope is purely a mena frontmatter concept.
- **UI/CLI for scope queries**: No `ari mena list --scope=user` or similar. Could be added later.
- **Scope for embedded FS rite mena**: Rite-level mena embedded in the binary is always project-scoped by pipeline. Adding scope parsing to embedded sources is a COULD for consistency but not required for PR2.

## Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Build passes | `CGO_ENABLED=0 go build ./cmd/ari` | Exit code 0 |
| All tests pass | `CGO_ENABLED=0 go test ./...` | Exit code 0, no failures |
| Scope filtering in materialize | Test with `scope: user` entry, verify excluded from project output | Automated test |
| Scope filtering in usersync | Test with `scope: project` entry, verify excluded from user output | Automated test |
| Backward compatibility | Run existing test suite without modification, all pass | Zero regressions |
| Invalid scope rejected | Test with `scope: invalid`, verify error returned | Automated test |
| ADR documented | `docs/decisions/ADR-0025-mena-scope.md` exists | File present |
| No scope = both | Test with no scope field, verify included in both pipelines | Automated test |

## Open Questions

None. All decisions were resolved during stakeholder interview:
- Default is "both" (confirmed)
- Mena-only scope (confirmed, agents/hooks deferred)
- String enum, not bitmask (confirmed)
- Validation rejects unknown values (confirmed)

## Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| PRD | `/Users/tomtenuta/Code/knossos/docs/prd/PRD-mena-scope.md` | Read-verified 2026-02-07 |
