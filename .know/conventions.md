---
domain: conventions
generated_at: "2026-03-23T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "78abb186"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "9b8f28035904e1dcd19d584717ac3629753bc808309497ce41f27180caabe7ac"
---

# Codebase Conventions

> Module: `github.com/autom8y/knossos` | Go 1.23+ | CLI binary: `ari` (Ariadne)

## Error Handling Style

### Error Creation

The codebase has a dedicated custom error package at `internal/errors/errors.go`. **All domain errors are created through this package**, not through `fmt.Errorf` or stdlib `errors.New` with raw strings.

The domain error package exposes three creation primitives:

- `errors.New(code, message)` — creates an `*errors.Error` with a symbolic code
- `errors.NewWithDetails(code, message, details)` — same, with a structured `map[string]any` for context
- `errors.Wrap(code, message, cause)` — wraps a lower-level error; stores cause in `Details["cause"]` (for JSON serialization) AND in an unexported `cause` field (for `errors.As` chain traversal)

All three automatically map `code` to an exit code via an internal `exitCodeForCode` switch.

### Error Code System

Errors carry symbolic string codes (uppercase with underscores) such as `"FILE_NOT_FOUND"`, `"LIFECYCLE_VIOLATION"`, `"MERGE_CONFLICT"`. These codes serve two purposes:

1. **Exit code routing** — `exitCodeForCode()` maps each code to a numeric exit code (1–21). Exit codes are defined as `ExitXxx` constants in the same package.
2. **JSON serialization** — the `*Error.JSON()` method wraps the struct as `{"error": {"code": "...", "message": "...", "details": {...}}}`.

Domain-specific code groups are co-located in the errors package: session-domain codes, rite-domain codes, sync-domain codes, and manifest-domain codes all live in `internal/errors/errors.go`.

### Domain Error Constructors

The errors package provides ~25 named constructor functions (`ErrProjectNotFound()`, `ErrSessionNotFound(sessionID)`, `ErrLifecycleViolation(from, to, reason)`, etc.). These are the idiomatic way to produce domain errors in `cmd/` and `internal/` code. They are used in 116 files with 584 total `errors.New/Wrap/NewWithDetails` call sites.

### Stdlib `fmt.Errorf` Usage

`fmt.Errorf` with `%w` is used in 43 files (148 occurrences), primarily for:
- Low-level wrapping of stdlib errors without needing a domain code (e.g., `fmt.Errorf("failed to marshal frontmatter: %w", err)`)
- Cobra argument-validation helper strings
- Leaf packages that predate or are outside the domain error model (e.g., compiler sub-packages)

This is acceptable usage — the domain error system does not replace all `fmt.Errorf` wrapping of stdlib errors.

### Sentinel Errors

Sentinel errors using stdlib `errors.New` appear only in leaf packages that cannot import the domain errors package (to avoid import cycles). Example: `internal/frontmatter/errors.go` defines `ErrMissingOpenDelimiter` and `ErrMissingCloseDelimiter` as package-level vars.

### `stderrors` Import Alias

When a file needs both the domain `errors` package and the stdlib `errors` package, it imports stdlib with the alias `stderrors`. This is the consistent convention — 13 files use it. Example:

```go
import (
    stderrors "errors"
    "github.com/autom8y/knossos/internal/errors"
)
```

### Error Propagation

The standard propagation style is immediate `return` on error — no aggregation, no retry loops. `if err != nil { return ... }` appears 2,443 times across 374 files.

Defer-cleanup is used for resource release (file handles, writers): `defer func() { _ = f.Close() }()`. The ignored-error pattern `_ = f.Close()` is standard for cleanup paths where a prior error is already in flight.

### Error Handling at CLI Boundary

The CLI boundary (`cmd/ari/main.go`) uses a handled-error sentinel to prevent double-printing:

1. Commands that print their own errors call `common.PrintAndReturn(printer, err)`, which prints via `output.Printer` and wraps the error in `errors.Handled(err)`.
2. `main.go` checks `errors.IsHandled(err)` — if handled, skips printing and only calls `os.Exit(errors.GetExitCode(err))`.
3. Unhandled errors are printed by `main.go` via `printer.PrintError(err)`, which renders as JSON or plain text based on `--output` format.

`common.PrintAndReturn` is called at 248 call sites across the cmd layer.

### `slog` for Internal Warnings

Non-user-visible diagnostics (e.g., non-fatal degraded paths, file watcher skip decisions) use `log/slog` with key-value pairs:

```go
slog.Warn("failed to detect provenance divergence", "error", err)
```

`slog` is used 83 times, concentrated in `internal/materialize/` (the hotspot package).

### No Must-Style Panics

There are zero `Must`-style panic wrappers in `internal/`. The codebase does not use panic for error escalation.

---

## File Organization

### cmd/ vs internal/ Separation Philosophy

- `cmd/ari/main.go` — single entry point; only wires version variables, embeds assets, and runs the root command. No business logic.
- `internal/cmd/` — all Cobra command implementations, organized as one sub-package per CLI command group (e.g., `internal/cmd/session/`, `internal/cmd/rite/`, `internal/cmd/hook/`)
- `internal/` (non-cmd packages) — pure domain logic, no Cobra dependencies

Each `internal/cmd/{group}/` package contains:
- `{group}.go` — the `New{Group}Cmd(outputFlag, verboseFlag, projectDir...)` root factory; registers sub-commands
- One file per sub-command verb: `create.go`, `list.go`, `status.go`, `park.go`, etc.
- Test files co-located: `create_test.go`, `session_test.go`, etc.

### Command Implementation Pattern

Every command file follows a three-part structure:
1. **Options struct** (`type {verb}Options struct`) — holds flag values, unexported, local to the file
2. **Constructor** (`func new{Verb}Cmd(ctx *cmdContext) *cobra.Command`) — creates the Cobra command, declares flags, wires `RunE` to call `run{Verb}`
3. **Runner** (`func run{Verb}(ctx *cmdContext, args..., opts {verb}Options) error`) — implements the command logic; returns domain errors

There are 124 `run[A-Z]` functions and 50 `*Options` types across `internal/cmd/`.

### Shared Context via cmdContext

Each cmd group defines a package-private `cmdContext` struct embedding `common.BaseContext` or `common.SessionContext`. These embed flag pointers (`*string`, `*bool`) for output format, verbosity, and project directory. Helper methods (`GetPrinter`, `GetResolver`, `GetSessionID`, `GetLockManager`) live on `common.BaseContext`/`common.SessionContext` in `internal/cmd/common/context.go`.

### File Naming Within Domain Packages

Domain packages use descriptive single-noun or compound-noun file names:
- Verb-named operation files: `promote.go`, `rotate.go`, `resolve.go`, `discover.go`
- Noun-named type files: `status.go`, `timeline.go`, `snapshot.go`, `context.go`
- Specialization files: `materialize_agents.go`, `materialize_mena.go`, `materialize_settings.go` (all in `internal/materialize/` — the hotspot package with 40+ files)
- `types.go` — pure type/const declarations, present in 12 packages
- `models.go` — present in 3 packages (`ask`, `explain`, `tour`) for command-local response models

### Constants and var Declarations

Constants live in the package where they are first used. The `internal/errors/` package centralizes all exit codes and error codes as `const` blocks. Other packages define local type-based constants (e.g., `session.StatusActive`, `naxos.ReasonInactive`). There are 115 `const` blocks across `internal/`.

No package uses a dedicated `constants.go` file — constants live in the most relevant file (e.g., status constants in `status.go`, error codes in `errors.go`).

### init() Usage

`init()` is used sparingly — exactly 4 files:
- `internal/hook/events.go` — computes `wireToCanonical` reverse map from `canonicalToWire` at startup
- `internal/channel/tools.go` — registers channel tool sets
- `internal/cmd/root/root.go` — viper config binding
- `internal/cmd/explain/concepts.go` — registers concept index

### Internal Sub-packages

Sub-packages within a domain package are used to enforce package boundaries on large packages:
- `internal/materialize/compiler/` — `ChannelCompiler` interface + `ClaudeCompiler`/`GeminiCompiler` implementations
- `internal/materialize/hooks/` — MCP/hook config types
- `internal/materialize/mena/` — mena directory walking/rendering
- `internal/materialize/orgscope/`, `userscope/`, `procession/`, `source/` — scope-gated pipeline stages

### Output Types Organization

All CLI output structs (`*Output`, `*Result`) for JSON/YAML/text rendering live in `internal/output/output.go` for session-domain commands, or in `internal/output/rite.go` for rite-domain commands. There are 115 `*Output` types total. Each implements either `Textable` (for single-string text rendering) or `Tabular` (for table rendering). Commands in `internal/cmd/` import `output.*` structs and pass them to `printer.Print(outputStruct)`.

### Generated Code

There is no `//go:generate` machinery. The `.gitignore` file at `.claude/.gitignore` is marked `# DO NOT EDIT — regenerated on every sync.` but this is content generation by the `ari sync` pipeline, not Go code generation.

---

## Domain-Specific Idioms

### The Materializer Pattern

The `internal/materialize/` package is the single most-changed hotspot (87 changes per `.sos/land/` experiential knowledge). Its core idiom:

- `materialize.Options` struct — all boolean/string configuration flags for a sync pipeline run
- `materialize.Materializer` struct — stateful executor; constructed via `NewWiredMaterializer(resolver)` in the cmd layer
- `MaterializeWithOptions(resolver, opts)` — the rite-scope pipeline entry point
- `writeIfChanged(path, content, perm)` — idempotency invariant: files are only written when content actually changes (prevents unnecessary file watcher triggers)
- `fileutil.AtomicWriteFile` — temp-file-then-rename pattern for all file writes to prevent partial writes

### Polymorphic YAML Fields

YAML fields that accept multiple representations implement custom `UnmarshalYAML`. Two concrete examples:
- `frontmatter.FlexibleStringSlice` — accepts both a CSV string (`"Bash, Read, Glob"`) and a YAML sequence. Used for `tools:` and `disallowedTools:` in agent frontmatter.
- `agent.MemoryField` — accepts both boolean (`true`/`false`) and string paths. The `UnmarshalYAML` method detects YAML node type and converts accordingly.

This pattern is used whenever the upstream YAML sources (agent files, rite manifests) use informal representations.

### ChannelCompiler Interface

Channel-agnostic rendering is implemented via `compiler.ChannelCompiler`:

```go
type ChannelCompiler interface {
    CompileCommand(name, description, argHint, body string) (string, []byte, error)
    CompileSkill(name, description, body string) (string, string, []byte, error)
    CompileAgent(name string, frontmatter map[string]any, body string) ([]byte, error)
    ContextFilename() string
}
```

`ClaudeCompiler` and `GeminiCompiler` implement this. The materialize pipeline selects the correct compiler based on the `channel` option and calls the same interface methods regardless of target.

### Domain-Typed String Enums

The codebase consistently uses `type X string` with typed constants for domain states rather than bare strings:
- `session.Status` — `StatusNone`, `StatusActive`, `StatusParked`, `StatusArchived`
- `naxos.OrphanReason` — `ReasonInactive`, `ReasonStaleSails`, `ReasonIncompleteWrap`
- `rite.CheckStatus` — defined in `internal/rite/validate.go`
- `output.Format` — `FormatText`, `FormatJSON`, `FormatYAML`

Each typed string has a `.String()` method and usually an `.IsValid()` or `.IsTerminal()` method.

### handledError Sentinel

The `errors.Handled(err)` / `errors.IsHandled(err)` pattern is a codebase-specific idiom to prevent double-printing of errors at the CLI boundary. This is not a standard Go pattern — it is specific to this project. See `internal/errors/errors.go` lines 537–558.

### Scope-Gated Pipeline Stages (rite scope / user scope)

The sync pipeline has two scopes: rite scope and user scope. Both use `materialize.Options` but different stages execute for each. User-scope file writes skip `writeIfChanged` in some cases (noted inline in `internal/materialize/userscope/sync.go` line 50) because user-scope files are not covered by the file-watcher optimization.

### Registry Pattern (internal/registry)

`internal/registry` is a **leaf package** (imports only stdlib) providing a typed key-value map of platform references. `RefKey` is a typed string constant. This is used for denial-recovery when agents need to look up stable CLI commands or agent names. It is NOT a singleton — it is used via the typed constants directly.

### Procession/Provenance Pattern

`internal/provenance` and `internal/procession` are support packages for the pipeline. Provenance tracks file ownership (which files were written by which rite sync), preventing user-owned files from being overwritten. A `Collector` interface is threaded through pipeline stages.

### `paths.Resolver` Everywhere

Almost all domain packages receive a `*paths.Resolver` rather than raw directory strings. The resolver centralizes path construction for sessions, locks, channel dirs, and XDG base dirs. Constructed via `paths.NewResolver(projectDir)` in the cmd common context.

---

## Naming Patterns

### Package Naming

- Packages use single lowercase nouns: `session`, `errors`, `paths`, `manifest`, `materialize`, `inscription`, `procession`, `tribute`, `registry`, `validation`, `worktree`
- Sub-packages use nouns: `compiler`, `mena`, `hooks`, `orgscope`, `userscope`, `source`
- The `cmd` sub-tree mirrors the CLI verb hierarchy: `internal/cmd/session/`, `internal/cmd/rite/`
- One exception: `internal/cmd/initialize/` (package name is `initialize`, not `init`, to avoid collision with the Go builtin)

### Type Naming

**Exported types** follow consistent suffixes:
- `*Options` — command configuration structs (unexported in cmd layer; exported in domain packages like `materialize.Options`)
- `*Result` — operation outcomes (`materialize.Result`, `session.RotationResult`, `artifact.GenerateResult`)
- `*Output` — CLI output structs implementing `Textable` or `Tabular`
- `*Frontmatter` — parsed YAML frontmatter types (`agent.AgentFrontmatter`)
- `*Context` — shared state containers (`session.Context`, `common.BaseContext`, `common.SessionContext`)
- `*Manifest` — parsed manifest file structs (`manifest.Manifest`, `materialize.RiteManifest`)
- `*Manager` — stateful service types (`lock.Manager`)
- `*Resolver` — path/dependency resolution types (`paths.Resolver`)
- `*Ref` — reference types in workflow declarations (`agent.UpstreamRef`, `agent.DownstreamRef`)
- `*Compiler` — channel-specific rendering implementations (`compiler.ClaudeCompiler`)
- `*Registry` — typed lookup tables (`artifact.Registry`, `registry` package)

**Unexported types** in cmd layer: `cmdContext`, `createOptions`, `listOptions` — always lowercase first letter, no suffix variation.

### Function Naming

- Constructor pattern: `New{Type}(...)` — used for all service/domain objects: `NewResolver`, `NewPrinter`, `NewManager`, `NewMaterializer`, `NewSessionCmd`
- Sub-command constructors: `new{Verb}Cmd(ctx)` (lowercase `new`, uppercase verb) — used exclusively within cmd packages
- Runner functions: `run{Verb}(ctx, args, opts)` — pairs with `new{Verb}Cmd`, lowercase `run` + title-case verb
- Error constructors: `Err{DomainConcept}(args...)` — e.g., `ErrProjectNotFound()`, `ErrSessionNotFound(id)`, `ErrLifecycleViolation(from, to, reason)`
- Predicate helpers: `Is{Condition}(err error) bool` — e.g., `IsNotFound`, `IsHandled`, `IsBudgetExceeded`

### Variable Naming

- Short loop variables: `i`, `v`, `f`, `s`, `e` — standard Go style
- Cobra flag binding: `opts.{fieldName}` into `cmd.Flags().{Type}VarP(&opts.{fieldName}, ...)` — flags always bound into an options struct, never into bare vars
- Error variable: always `err` for the first error, `{noun}Err` for secondary errors in the same scope (e.g., `writeErr`, `readErr`, `checksumErr`, `removeErr`)
- Result variable: typically named after the operation noun: `result`, `registry`, `ctx`, `manifest`

### Acronym Conventions

- `URL` stays uppercase: `url` (lowercase in local vars), `URL` in exported names
- `ID` stays uppercase: `sessionID`, `SessionID`
- `JSON` stays uppercase: `FormatJSON`, `printJSON`
- `YAML` stays uppercase: `FormatYAML`, `printYAML`
- `MCP`, `CLI`, `FSM`, `CC` stay uppercase in comments and identifiers where they appear in exported names
- Exception: `McpServers` in `AgentFrontmatter` (camelCase to match the Claude Code harness YAML schema — documented as an intentional deviation)

### File Naming

- Go files: `{noun}.go` or `{verb}.go`, matching the primary type or operation in the file
- Test files: `{name}_test.go` co-located with source
- Integration tests: `{package}_integration_test.go` or `{domain}_integration_test.go` (e.g., `moirai_integration_test.go`)
- Regression tests: `{name}_regression_test.go` (e.g., `scar_regression_test.go`)
- No `_gen.go` generated files exist
- Materialize package uses prefixed files: `materialize_{topic}.go` (e.g., `materialize_agents.go`, `materialize_settings.go`) to keep the large package scannable

### Package-Doc Comment Pattern

Every non-cmd package has a `// Package {name} provides/implements ...` comment as the first line of one file (usually the primary file). Cmd packages repeat the package doc comment redundantly across their multiple files — this is an observed inconsistency but not a violation.

---

## Knowledge Gaps

1. **`internal/channel/` package** — not examined in detail; `tools.go` was only partially read. Channel-specific tool allowlisting behavior is not fully documented here.
2. **`internal/hook/clewcontract/` package** — the clew-contract protocol (structured event writing between hook commands and session context) was not examined. This may have additional idioms.
3. **`internal/lock/` package** — lock protocol specifics (JSON LockMetadata v2, stale threshold) were noted but not read from source.
4. **`internal/perspective/` and `internal/resolution/` packages** — not examined; likely provide rite-switching and dependency-chain resolution idioms.
5. **Error wrapping in `fmt.Errorf`** — 148 occurrences in 43 files; not all call sites were examined to determine whether `%w` is universally used or sometimes omitted.
6. **Test helper patterns** — test-specific context constructors (`newTestContext`, `newQueryTestContext`) were observed but not fully documented.
