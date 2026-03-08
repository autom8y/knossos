---
domain: conventions
generated_at: "2026-03-08T21:08:37Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "dbf81b8"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "602e86d06024078a2882f921cbc613ed4bb78196e981232db79d664f342225c8"
---

# Codebase Conventions

> Language: Go 1.23+ (module: `github.com/autom8y/knossos`). Single binary (`ari`) built with `CGO_ENABLED=0`. 299 source files, 197 test files across 50+ packages in `internal/` and one entry-point in `cmd/ari/main.go`.

## Error Handling Style

### Error Creation Philosophy

The project uses a **custom domain error type** (`internal/errors.Error`) as the authoritative error vehicle — not `errors.New` from stdlib directly. The custom type carries a string `Code`, human-readable `Message`, structured `Details map[string]any`, `ExitCode int`, and an unexported `cause error` for Go error chain traversal.

**Three creation helpers** cover all use cases:

```go
errors.New(code string, message string) *Error
errors.NewWithDetails(code string, message string, details map[string]any) *Error
errors.Wrap(code string, message string, cause error) *Error
```

`Wrap` stores the cause both in `Details["cause"]` (for JSON serialization) and as the unexported `cause` field (for `errors.As`/`errors.Is` traversal). This is the canonical pattern when wrapping an underlying stdlib or third-party error.

**Named constructors** for every domain error are defined in `internal/errors/errors.go`:

```go
errors.ErrProjectNotFound() *Error
errors.ErrSessionNotFound(sessionID string) *Error
errors.ErrLifecycleViolation(from, to, reason string) *Error
errors.ErrLockTimeout(lockPath string, lockMeta any) *Error
errors.ErrRiteNotFound(riteName string) *Error
errors.ErrBudgetExceeded(current, requested, limit int) *Error
// ... 25+ domain constructors
```

### Error Codes and Exit Codes

All error codes are SCREAMING_SNAKE_CASE string constants in `internal/errors/errors.go`. Exit codes 0-21 are also declared there. The `exitCodeForCode()` function maps Code → ExitCode at creation time — callers never set `ExitCode` manually.

```go
const (
    CodeGeneralError       = "GENERAL_ERROR"
    CodeSessionNotFound    = "SESSION_NOT_FOUND"
    CodeLifecycleViolation = "LIFECYCLE_VIOLATION"
    // ...
)
```

Domain-specific code groups are declared in grouped `const` blocks with inline comments, e.g., `// Rite-domain error codes`, `// Sync-domain error codes`.

### Error Wrapping

Two wrapping styles coexist:

1. **Domain wrapping** (preferred): `errors.Wrap(code, message, cause)` — used when the error originates from a lower layer and needs to be re-coded for the domain. Example (`internal/paths/paths.go`):

```go
return "", errors.Wrap(errors.CodeGeneralError, "failed to get working directory", err)
```

2. **Stdlib wrapping** (`fmt.Errorf("%w", ...)`) — used inside packages that don't cross a CLI boundary (e.g., `internal/materialize/mena/engine.go`, `internal/know/`). The domain error type's `Unwrap()` is compatible with this: callers use `errors.As` to traverse mixed chains. Found in 23 files (70 occurrences of `%w`).

### Error Propagation Pattern

Dominant pattern is **immediate return on error** — `if err != nil { return ..., err }`. Found in 2,220 occurrences across 328 files. No error aggregation library is used; domain functions accumulate issues in `[]string` slices (e.g., `(c *Context) Validate() []string`) rather than multi-error types.

### Error Handling at Boundaries

**CLI boundary** (`cmd/ari/main.go`): Cobra commands propagate errors back to `root.Execute()`. `main()` then:
1. Checks `errors.IsHandled(err)` — if already printed, skips re-printing.
2. Prints via `output.NewPrinter().PrintError(err)` — format-aware (JSON vs text).
3. Exits with `errors.GetExitCode(err)`.

**Command layer** (`internal/cmd/common/errors.go`): Commands that handle their own error display use `common.PrintAndReturn(printer, err)` which calls `printer.PrintError(err)` then wraps in `errors.Handled(err)` to suppress re-printing.

**JSON output**: `errors.Error` implements `JSON() string` returning `{"error": {"code":..., "message":..., "details":...}}`. The `PrintError` method in `internal/output/output.go` checks for the `JSON() string` interface before falling back to generic JSON wrapping.

**`Is*` predicates**: Every code group has typed predicate functions (`IsNotFound`, `IsLifecycleError`, `IsBudgetExceeded`, etc.) that use `errors.As` for chain-safe traversal — never string comparison.

**`handledError` sentinel** (`internal/errors/errors.go:537`): An unexported wrapper type that marks an error as already printed. `errors.Handled(err)` wraps, `errors.IsHandled(err)` unwraps. This prevents double-printing at the CLI boundary.

### Packages Using `fmt.Errorf` Directly

Some internal packages (not crossing CLI boundary) use stdlib patterns. `internal/materialize/mena/`, `internal/know/`, `internal/inscription/generator.go`, and a few others use `fmt.Errorf("...: %w", err)` for lightweight wrapping when the error will be re-wrapped at a higher level. This is a pragmatic pattern, not a violation.

## File Organization

### Directory Layout Philosophy

```
cmd/ari/main.go          — entry point, minimal logic
internal/
  cmd/                   — cobra command implementations (one subdirectory per command group)
    session/             — 20+ files, one file per subcommand
    hook/                — 15 files, one file per hook type
    agent/, rite/, ...
  errors/errors.go       — single-file package (small domain)
  paths/paths.go         — single-file package
  fileutil/fileutil.go   — single-file package
  session/               — domain logic, one file per concern
  materialize/           — hotspot, prefixed file names for sub-concerns
  inscription/           — pipeline pattern, types.go + named-concern files
  output/output.go       — one large file (all output types co-located)
```

### File-per-Concern in Domain Packages

Domain packages (e.g., `internal/session/`, `internal/inscription/`) use one file per logical concern:

| File | Contents |
|---|---|
| `types.go` / `status.go` | Type definitions, constants, enum predicates |
| `fsm.go` | State machine |
| `context.go` | Primary domain entity |
| `discovery.go` | Filesystem scanning |
| `resolve.go` | Entity lookup |
| `validate.go` | Validation logic |
| `generator.go` | Generation logic |
| `merger.go` / `syncer.go` | Merge / sync logic |
| `backup.go` | Backup management |
| `pipeline.go` | Orchestration of sub-steps |
| `marker.go` | Specific sub-feature |

### Naming Within `internal/cmd/`

The `internal/cmd/` subtree mirrors Cobra command groups. Each directory contains:

- `{group}.go` — root command (subcommand registration)
- `{subcommand}.go` — one file per leaf subcommand
- Test files named `{subcommand}_test.go`

Example (`internal/cmd/session/`): `session.go`, `create.go`, `park.go`, `resume.go`, `wrap.go`, `list.go`, `status.go`, `lock.go`, `unlock.go`, `gc.go`, etc.

### Naming Within `internal/materialize/`

The hottest package uses **prefixed file names** to group by processing stage:

- `materialize.go` — orchestrator, `Options`, `Result`, `RiteManifest` types, `Materializer` struct
- `materialize_agents.go` — agent materialization stage
- `materialize_claudemd.go` — CLAUDE.md inscription stage
- `materialize_mena.go` — mena projection stage
- `materialize_settings.go` — settings.json stage
- `materialize_rules.go` — rules/ directory stage
- `sync_types.go` — unified sync types
- `syncer.go` — interface implementation shim
- Sub-packages: `hooks/`, `mena/`, `source/`, `userscope/`, `orgscope/`

### Test File Placement

Tests live in the same directory as the source, named `{source}_test.go`. Integration tests are named `integration_test.go`. Package declarations match the package under test (white-box testing, same package name), except for external integration tests that use `package {name}_test` in some packages. Fuzz tests live in `fuzz_test.go`.

### Types Files

`types.go` is the conventional home for domain entity structs (e.g., `internal/naxos/types.go`, `internal/tribute/types.go`, `internal/mena/types.go`, `internal/perspective/types.go`). When a package is small (one primary concern), all types live in the single package file (e.g., `internal/paths/paths.go`, `internal/fileutil/fileutil.go`).

### Generated/Special Files

- `embed.go` (root) — Go `//go:embed` directives for embedded FS
- `internal/assets/assets.go` — embedded asset access
- `internal/cmd/common/embedded.go` — passes embedded FS to commands

### `internal/` Boundary

`internal/errors`, `internal/paths`, `internal/fileutil`, `internal/frontmatter`, `internal/output` are foundational packages imported by many others. `internal/cmd/` imports domain packages; domain packages do NOT import `internal/cmd/`. The `internal/errors` package is imported by ~115 domain and cmd files.

## Domain-Specific Idioms

### Typed String Enums

The project uses a consistent enum pattern: declare a named `type Foo string`, then `const` blocks with `StatusFoo`, `FooBar` values, then `IsValid() bool`, `String() string` methods, and typed predicate helpers. Used throughout:

- `session.Status` — NONE / ACTIVE / PARKED / ARCHIVED
- `session.Phase` — requirements / design / implementation / validation / complete
- `session.Complexity` — in `complexity.go`
- `sails.Color` — WHITE / GRAY / BLACK
- `inscription.OwnerType` — knossos / satellite / regenerate
- `inscription.MarkerDirective` — START / END / ANCHOR
- `materialize.SyncScope` — all / rite / org / user
- `hook.HookEvent` — PreToolUse / PostToolUse / Stop / etc.

Each type implements `IsValid() bool` and `String() string`. This avoids stringly-typed comparisons everywhere the enum is used.

### `New*` Constructor Pattern

Every non-trivial struct has a `New*` constructor:

- `NewFSM()`, `NewGenerator(sessionPath)`, `NewManager(dir)`, `NewResolver(root)`, `NewMaterializer(resolver)`, `NewBackupManager(projectRoot)`, `NewPrinter(format, out, errOut, verbose)`, etc.

Constructor variants for different initialization paths use `New*With*` naming:

- `NewMaterializerWithSource(resolver, source)`
- `NewContextLoaderWithPaths(ritesDir, userDir)`
- `NewGeneratorWithFS(templateFS, manifest, ctx)`
- `NewGeneratorWithValidator(sessionPath, validator)`
- `NewPipelineWithPaths(claudeMDPath, manifestPath, templateDir, backupDir)`

Method chaining on builders uses `With*` methods returning the receiver:

- `materializer.WithEmbeddedFS(fsys)`, `.WithEmbeddedTemplates(fsys)`, `.WithEmbeddedAgents(fsys)`, `.WithClaudeDirOverride(dir)`

### `*Options` / `*Config` / `*Params` Structs

Configuration is always passed as a named struct, never as a long argument list:

- `materialize.Options` (8 boolean fields for sync behavior)
- `materialize.SyncOptions` (10 fields, scope + flags)
- `inscription.InscriptionSyncOptions`, `inscription.InscriptionMergeOptions`
- `worktree.CreateOptions`, `WorktreeSwitchOptions`, `CloneOptions`
- `mena.MenaProjectionOptions`

The pattern is: small `Options` struct with `bool` fields (DryRun, Force, KeepAll, etc.) and specific scalar fields.

### `*Result` Structs

Functions that produce output return a named `*Result` struct instead of multiple return values where feasible:

- `materialize.Result`, `materialize.SyncResult`, `RiteScopeResult`, `OrgScopeResult`
- `inscription.SyncResult`, `MergeResult`, `ValidationResult`
- `sails.GenerateResult`, `GateResult`
- `rite.InvokeResult`, `ReleaseResult`, `ValidationResult`
- `worktree.CleanupResult`
- `ledge.PromoteResult`, `AutoPromoteResult`

### Interface Design

Interfaces are small (1-3 methods) and defined in the consuming package:

- `internal/provenance/collector.go`: `Collector` interface (2 methods: `Record`, `Entries`)
- `internal/rite/syncer.go`: `Syncer` interface (1 method: `SyncRite`)
- `internal/output/output.go`: `Tabular` (2 methods), `Textable` (1 method)
- `internal/search/synonyms.go`: `SynonymSource` interface

### `Tabular` and `Textable` Output Pattern

All CLI output types implement one of two interfaces from `internal/output/output.go`:
- `Textable` — single `Text() string` method for custom text rendering
- `Tabular` — `Headers() []string` + `Rows() [][]string` for table rendering

This allows `output.Printer.Print(data)` to dispatch without type switches in callers. All `*Output` structs in `internal/output/output.go` implement one of these.

### Atomic File Writes

All file writes outside of tests use `fileutil.AtomicWriteFile(path, content, perm)` (temp-file-then-rename) or `fileutil.WriteIfChanged(path, content, perm)` (skip write if content identical, preventing file-watcher thrash). Defined in `internal/fileutil/fileutil.go`, used by 31 files.

### `paths.Resolver` Dependency Injection

The `paths.Resolver` struct centralizes all path resolution and is passed as a dependency to most domain packages. It wraps a `projectRoot string` and exposes typed path accessors (e.g., `resolver.SessionsDir()`, `resolver.AgentsDir()`, `resolver.KnossosManifestFile()`). Functions that accept `*paths.Resolver` never construct their own paths directly.

### Mena File Extension Idiom

Files with `.dro.md` extension are dromena (commands); `.lego.md` are legomena (skills). This double-extension convention is unique to this project and enforced by `internal/mena/types.go`. Extension stripping is done by `StripMenaExtension(filename string) string`.

### `FlexibleStringSlice`

`internal/frontmatter.FlexibleStringSlice` is a custom YAML type that accepts both comma-separated strings and YAML list syntax. It is aliased (not re-implemented) in `internal/agent/types.go` as `type FlexibleStringSlice = frontmatter.FlexibleStringSlice`. Used for `tools:` and `allowedTools:` frontmatter fields.

### `slog` for Structured Logging

Where logging occurs in production code (not tests), `log/slog` from the stdlib is used. Found in 10 files (`internal/materialize/`, primarily). No third-party logging library. CLI commands do not log; they use `output.Printer` for user-facing output.

### Provenance Collector Thread

The `provenance.Collector` interface is threaded through every stage of the materialization pipeline as an explicit parameter. Stages call `collector.Record(relativePath, entry)` to build a provenance manifest as a side effect of writing files. The collector is not global state.

## Naming Patterns

### Package Names

All package names are **single lowercase words** (Go convention). Multi-word domains use compound names: `fileutil`, `clewcontract`, `materialize`, `worktree`, `frontmatter`, `tokenizer`, `checksum`. No underscores in package names. Sub-packages are nested directories: `materialize/mena`, `materialize/hooks`, `materialize/userscope`, `materialize/source`, `hook/clewcontract`.

### Type Names

**Structs**: PascalCase, descriptive nouns. Consistent suffixes:
- `*Manager` — manages a resource with lifecycle (e.g., `lock.Manager`, `worktree.MetadataManager`, `inscription.BackupManager`)
- `*Resolver` — resolves paths or entities (e.g., `paths.Resolver`, `session.Resolver`)
- `*Generator` — produces artifacts (e.g., `inscription.Generator`, `sails.Generator`, `tribute.Generator`)
- `*Loader` — loads from storage (e.g., `inscription.ManifestLoader`, `rite.ContextLoader`)
- `*Collector` — accumulates items (e.g., `sails.ProofCollector`)
- `*Validator` — validates (e.g., `rite.Validator`, `validation.HandoffValidator`)
- `*Scanner` — filesystem scanning (e.g., `naxos.Scanner`)
- `*Aggregator` — aggregates (e.g., `artifact.Aggregator`)
- `*Extractor` — extracts (e.g., `tribute.Extractor`)
- `*Materializer` — the single materializer in `internal/materialize/`
- `*Printer` — the single printer in `internal/output/`
- `*Pipeline` — orchestrates stages (e.g., `inscription.Pipeline`)
- `*Merger` — merges content (e.g., `inscription.Merger`)

**Options/Config/Result**: `*Options`, `*Config`, `*Params`, `*Result` suffixes (documented above).

### Acronym Conventions

Go standard (uppercase acronyms):
- `SessionID` not `SessionId`
- `YAML` not `Yaml` (in constants/comments, but Go struct field names use `yaml.v3` tags, e.g., `yaml:"session_id"`)
- `JSON` not `Json`
- `URL` not `Url`
- `MCP` not `Mcp` (e.g., `MCPServer`, `MCPServerConfig`)
- `CWD` not `Cwd`

`CC` is used throughout as an abbreviation for "Claude Code" (project-specific, not a standard acronym). Example: `CCSessionID`, `CCMapDir()`, `CCMapOrphans`.

### Variable Names

Receiver names are single letters or short abbreviations consistent within a type:
- `(f *FSM)` — `f`
- `(m *Manager)` — `m`
- `(m *Materializer)` — `m`
- `(r *Resolver)` — `r`
- `(g *Generator)` — `g`
- `(p *Printer)` — `p`
- `(c *Context)` — `c`
- `(s *Snapshot)` — `s`

Local variable names: `err` for errors (universal), `opts` for options structs, `res` or `result` for results, `resolver` for `paths.Resolver`, `manifest` for manifest types.

### `Err*` Constructor Naming

Domain error constructors follow `Err{Domain}{Concept}` pattern:
- `ErrProjectNotFound()`, `ErrSessionNotFound(id)`, `ErrSessionExists(id, status)`
- `ErrLifecycleViolation(from, to, reason)`, `ErrLockTimeout(path, meta)`
- `ErrRiteNotFound(name)`, `ErrOrphanConflict(orphans, cur, tgt)`
- `ErrBudgetExceeded(current, requested, limit)`

`Is*` predicates follow `Is{Category}` pattern: `IsNotFound`, `IsLifecycleError`, `IsMergeConflict`, `IsRiteNotFound`, `IsBudgetExceeded`, etc.

### Existing Naming Inconsistency (Do Not Spread)

One notable inconsistency: `internal/agent/types.go` declares `type McpServerConfig struct` (Mcp not MCP), while `internal/materialize/materialize.go` declares `MCPServer` and `MCPServerConfig`. New types should use `MCP` (uppercase). The `Mcp` variant in `agent/types.go` is a known inconsistency and should not be replicated.

## Knowledge Gaps

1. **`internal/cmd/ask/`** (search/NLP commands introduced recently): Only briefly observed. The naming and error patterns appear consistent but were not deeply read.
2. **`internal/search/` package**: Observed at surface level (type declarations), not traced through all search algorithm patterns.
3. **`internal/suggest/` and `internal/tribute/`**: Types and `New*` constructors confirmed, but internal pipeline idioms not deeply documented.
4. **Test conventions** excluded per criteria: documented in `.know/test-coverage.md`.
5. **`internal/provenance/` package**: Collector interface documented; the full divergence/merge logic not traced in detail.
6. **config package** (`internal/config/home.go`): Single small file, not read.
