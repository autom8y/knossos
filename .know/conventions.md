---
domain: conventions
generated_at: "2026-03-13T10:04:06Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "59a0de2"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "9b8f28035904e1dcd19d584717ac3629753bc808309497ce41f27180caabe7ac"
---

# Codebase Conventions

## Error Handling Style

### Error Creation

The project uses a **custom domain error type** (`internal/errors.Error`) as the single authoritative error mechanism. Standard library `errors.New` and `fmt.Errorf` are used only in two special circumstances: leaf packages that explicitly carry no internal imports (e.g., `internal/provenance`, `internal/resolution`), and in `internal/frontmatter/errors.go` for sentinel errors. Everywhere else in the codebase (122 files import `internal/errors`), errors are created through the project's own constructors.

The three constructors:
- `errors.New(code string, message string) *Error` — for fresh errors
- `errors.NewWithDetails(code string, message string, details map[string]any) *Error` — when structured context is needed
- `errors.Wrap(code string, message string, cause error) *Error` — when wrapping a lower-level error

Every error carries a **string code** (e.g., `CodeFileNotFound = "FILE_NOT_FOUND"`) and an **integer exit code** derived via `exitCodeForCode()`. The mapping is centralized — callers never set exit codes directly.

Named constructors (`ErrSessionNotFound`, `ErrLifecycleViolation`, `ErrLockTimeout`, etc.) exist for every common domain error. Callers use these rather than writing `errors.NewWithDetails(...)` inline.

### Error Codes and Exit Codes

All error codes are SCREAMING_SNAKE_CASE string constants in `internal/errors/errors.go`. Exit codes 0-21 are also declared there. The `exitCodeForCode()` function maps Code -> ExitCode at creation time — callers never set `ExitCode` manually.

### Error Wrapping

`errors.Wrap(code, message, cause)` is the primary pattern. The `Wrap` constructor stores the cause string in `Details["cause"]` for JSON serialization **and** sets the unexported `cause` field for `errors.As`/`errors.Is` chain traversal.

When standard library `errors` is also needed (for `errors.As`, `errors.Is`), it is aliased as `stderrors "errors"`. This alias appears in `internal/errors/errors.go`, `internal/manifest/manifest.go`, `internal/manifest/schema.go`, and `internal/cmd/root/root.go`.

### Error Propagation

The universal propagation style in cmd-layer functions is immediate return with `common.PrintAndReturn(printer, err)`:

```go
sessCtx, err := session.LoadContext(ctxPath)
if err != nil {
    if errors.IsNotFound(err) {
        err = errors.ErrSessionNotFound(sessionID)
    }
    return common.PrintAndReturn(printer, err)
}
```

`common.PrintAndReturn` calls `printer.PrintError(err)` then wraps the error with `errors.Handled(err)`. This marks the error as already printed, so `main.go`'s error handler only extracts the exit code but does not re-print.

Domain (non-cmd) packages return `*Error` directly without printing. The printing contract belongs to the cmd layer.

### Error Classification Helpers

`Is*(err error) bool` predicates are provided for every error category: `IsNotFound`, `IsLifecycleError`, `IsOrphanConflict`, `IsMergeConflict`, `IsRiteNotFound`, `IsBorrowConflict`, etc. All use the private `isCode(err, codes...)` function which calls `errors.As` for chain traversal.

### Error Handling at Boundaries

- **CLI output**: `output.Printer.PrintError(err)` — routes to stderr; for JSON format, calls `err.(interface{ JSON() string })` if available; for text format, writes `Error: <message>`.
- **Exit codes**: `os.Exit(errors.GetExitCode(err))` in `main()` — uses `errors.As` to find the `*Error` in the chain.
- **Lock/cleanup defers**: errors from defer-cleanup are silently discarded with `_ = ...` (e.g., `defer func() { _ = sessionLock.Release() }`).

---

## File Organization

### cmd/ vs internal/ Separation

`cmd/ari/main.go` is minimal: it sets version info, wires embedded assets via package-level vars, calls `root.Execute()`, and handles the final error. No business logic lives in `cmd/`.

All business logic lives in `internal/`. The split within `internal/`:
- `internal/cmd/<subcommand>/` — Cobra command definitions and RunE implementations
- `internal/<domain>/` — Pure domain logic, no Cobra dependency

### Per-Package File Organization Patterns

**Domain packages** (e.g., `internal/session/`, `internal/rite/`, `internal/manifest/`) follow a consistent pattern:
- One file per major concept: `status.go` (type declarations), `fsm.go` (state machine), `context.go` (data + persistence), `discovery.go` (filesystem scanning), `id.go` (ID generation)
- Test files colocated: `status_test.go` alongside `status.go`
- No separate `types.go` in most domain packages; types live with their logic

**cmd packages** (`internal/cmd/<name>/`) follow a consistent pattern:
- `<name>.go` — cobra `Command` constructor + subcommand registration + shared `cmdContext` type + `getPrinter()` helper
- `<verb>.go` — one file per subcommand verb (e.g., `create.go`, `park.go`, `resume.go`)
- Each `<verb>.go` contains: `<verb>Options` struct, `new<Verb>Cmd()` constructor, `run<Verb>()` implementation

**Large materialize package** (`internal/materialize/`) uses feature-based file splitting:
- `materialize.go` — types, Options, Result, entry point
- `materialize_agents.go`, `materialize_settings.go`, `materialize_gitignore.go`, etc. — one file per materialization stage
- Subpackages for specialized concerns: `compiler/`, `mena/`, `source/`, `hooks/`, `userscope/`, `orgscope/`, `procession/`

### Where Constants, Variables, and Init Live

- **Constants** are colocated with the type they describe. Error codes in `internal/errors/errors.go` alongside the `Error` type.
- **Package-level vars** (regexp patterns, map tables) are declared at the top of the file using them, not in a separate `vars.go`.
- **`init()` functions** exist in `internal/cmd/explain/concepts.go`, `internal/cmd/root/root.go`, and `internal/channel/tools.go`. Used sparingly.

### internal/ Boundary

Several packages are explicitly annotated as **leaf packages** (zero internal imports):
- `internal/mena/` — "LEAF package — imports only stdlib"
- `internal/registry/` — "LEAF package — imports only stdlib. No internal/ imports."
- `internal/resolution/` — "ZERO internal imports. All tier paths are injected via constructor"
- `internal/provenance/` — "leaf package (no internal imports per ADR-0026)"

### Generated/Special Files

- `embed.go` at the module root holds all `//go:embed` directives for single-binary distribution
- `internal/assets/assets.go` — embedded asset access
- `internal/cmd/common/embedded.go` — passes embedded FS to commands

---

## Domain-Specific Idioms

### Typed String Enumerations

The dominant pattern for enums is `type Foo string` with exported `const` blocks:

```go
type Status string
const (
    StatusNone     Status = "NONE"
    StatusActive   Status = "ACTIVE"
    StatusParked   Status = "PARKED"
    StatusArchived Status = "ARCHIVED"
)
```

Each type implements `IsValid() bool` and `String() string`. Used throughout: `session.Status`, `session.Phase`, `sails.Color`, `inscription.OwnerType`, `hook.HookEvent`, `materialize.SyncScope`.

Integer enums use `iota` with bitflag pattern for filters (e.g., `mena.MenaFilter`).

### Polymorphic YAML Fields

`FlexibleStringSlice` in `internal/frontmatter/frontmatter.go` accepts both comma-separated strings and YAML list sequences. `MemoryField` in `internal/agent/types.go` accepts both boolean and string values. `session.strandList` accepts both old `[]string` and new `[]Strand` formats.

### Options + Result Struct Pattern

Every significant operation uses plain structs for options and results:
- `materialize.Options`, `materialize.SyncOptions`, `inscription.InscriptionSyncOptions`
- `materialize.Result`, `materialize.SyncResult`, `inscription.SyncResult`, `sails.GateResult`

No functional-options (`WithFoo()`) pattern. Options structs are passed by value.

### WriteIfChanged Idempotency Guard

`fileutil.WriteIfChanged(path, content, perm)` is the canonical write function for pipeline stages. Returns `(bool, error)` — true if a write occurred. Prevents unnecessary file watcher triggers. `fileutil.AtomicWriteFile` is the underlying primitive (temp file -> sync -> rename).

### cmdContext Pattern (cmd layer)

Each cmd subpackage defines a local `cmdContext` struct that embeds `common.SessionContext` (or `common.BaseContext`) and adds a package-local `getPrinter()` method. Commands are wired: `new<Verb>Cmd(ctx *cmdContext) *cobra.Command`.

### Resolution Chain

`internal/resolution.Chain` is the multi-tier resolution primitive. Constructed with ordered `Tier` structs (highest priority first). Used for resolving rites, processions, and contexts. Zero internal imports — all tier paths injected.

### Atomic File Writes

All file writes outside of tests use `fileutil.AtomicWriteFile(path, content, perm)` (temp-file-then-rename) or `fileutil.WriteIfChanged(path, content, perm)`. Used by 31 files.

### Provenance Collector Thread

The `provenance.Collector` interface is threaded through every stage of the materialization pipeline as an explicit parameter. `NullCollector` for dry-run.

### Mena File Extension Idiom

Files with `.dro.md` extension are dromena (commands); `.lego.md` are legomena (skills). Extension stripping done by `StripMenaExtension(filename string) string`.

### Canonical Vocabulary Convention

Internal event names use snake_case (`pre_tool`, `session_start`). Go constants use PascalCase (`EventPreTool`, `EventSessionStart`). Wire names per channel are translations — CC receives `PreToolUse`, Gemini receives `BeforeTool`. Core code always uses canonical form. Same principle for tools: knossos canonical names are snake_case (`run_shell`, `read_file`), adapters translate per channel.

---

## Naming Patterns

### Package Names

All package names are **single lowercase words** (Go convention). Multi-word domains use compound names: `fileutil`, `clewcontract`, `materialize`, `worktree`, `frontmatter`, `tokenizer`, `checksum`. No underscores.

### Type Names

Exported types follow `<Concept>` when the package provides domain context. Consistent suffixes:
- `*Manager` — manages a resource with lifecycle (e.g., `lock.Manager`, `inscription.BackupManager`)
- `*Resolver` — resolves paths or entities (e.g., `paths.Resolver`, `session.Resolver`)
- `*Generator` — produces artifacts (e.g., `inscription.Generator`, `sails.Generator`)
- `*Loader` — loads from storage (e.g., `inscription.ManifestLoader`)
- `*Collector` — accumulates items (e.g., `sails.ProofCollector`)
- `*Validator` — validates (e.g., `rite.Validator`)
- `*Scanner` — filesystem scanning (e.g., `naxos.Scanner`)
- `*Pipeline` — orchestrates stages (e.g., `inscription.Pipeline`)
- `*Merger` — merges content (e.g., `inscription.Merger`)

### Function Names

- Domain constructors: `New<Type>()` — e.g., `NewFSM()`, `NewPrinter()`, `NewResolver()`
- Named error constructors: `Err<Condition>()` — e.g., `ErrSessionNotFound()`, `ErrLifecycleViolation()`
- Error predicates: `Is<Condition>(err error) bool` — e.g., `IsNotFound()`, `IsMergeConflict()`
- cmd constructors: `new<Verb>Cmd(ctx) *cobra.Command` — lowercase `new`, capitalized verb
- cmd runners: `run<Verb>(ctx, ...) error` — lowercase `run`, unexported

### Variable Names

- `err` everywhere for errors
- `sessionID` (two words, `ID` all-caps following Go convention)
- `ctx` for command context structs
- `opts` for options structs
- `resolver` for `*paths.Resolver`
- `printer` for `*output.Printer`

### Acronym Handling

Go-standard all-caps: `SessionID`, `URL`, `HTTP`, `MCP` (e.g., `MCPServer`, `MCPServerConfig`). `CC` is project-specific abbreviation for "Claude Code" (e.g., `CCSessionID`, `CCMapDir()`).

### Existing Naming Inconsistency (Do Not Spread)

`internal/agent/types.go` declares `type McpServerConfig struct` (Mcp not MCP), while `internal/materialize/materialize.go` declares `MCPServer` and `MCPServerConfig`. New types should use `MCP` (uppercase).

---

## Knowledge Gaps

- `internal/config/` package conventions not inspected (config loading patterns, viper integration)
- `internal/lock/` locking protocol not fully documented (Manager API, lock types)
- `internal/channel/` tools.go `init()` pattern not fully explored
- Full `internal/materialize/hooks/` subpackage not inspected
- `internal/naxos/` type patterns not fully documented
- `internal/tribute/` and `internal/ledge/` file organization patterns not inspected
- Test conventions excluded per criteria: documented in `.know/test-coverage.md`
