---
domain: conventions
generated_at: "2026-03-03T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "1599813"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Conventions

## Error Handling Style

The project uses a **custom structured error type** in `internal/errors/errors.go` as the primary error infrastructure. Standard library `fmt.Errorf("%w", ...)` is used only in packages that do not import the internal errors package (mainly `internal/know/know.go` and similar peripheral packages). The custom error system is dominant across 70+ files.

### Error Creation

Three constructors at `internal/errors/errors.go`:

```go
errors.New(code string, message string) *Error
errors.NewWithDetails(code string, message string, details map[string]interface{}) *Error
errors.Wrap(code string, message string, cause error) *Error
```

`Wrap` stores the underlying error as `details["cause"] = cause.Error()` (string, not error chain). This means stdlib `errors.Is/As` does NOT work through a wrapped `*errors.Error`. The cause is for human-readable logging, not programmatic unwrapping.

### Error Structure

`internal/errors/errors.go`, line 66-71:

```go
type Error struct {
    Code     string                 `json:"code"`
    Message  string                 `json:"message"`
    Details  map[string]interface{} `json:"details,omitempty"`
    ExitCode int                    `json:"-"`
}
```

The `Code` is a SCREAMING_SNAKE_CASE string constant (e.g., `"GENERAL_ERROR"`, `"FILE_NOT_FOUND"`). `ExitCode` is derived automatically from `Code` via `exitCodeForCode()`. Codes and exit codes are co-located in `internal/errors/errors.go`.

### Domain-Grouped Error Constructors

Errors are grouped by domain with comment-delimited sections. Each domain has:
- String constants for error codes (e.g., `CodeRiteNotFound = "RITE_NOT_FOUND"`)
- Integer constants for exit codes (e.g., `ExitRiteNotFound = 19`)
- Constructor functions named `Err{Noun}{Condition}` (e.g., `ErrRiteNotFound`, `ErrSessionExists`, `ErrLifecycleViolation`)
- `Is{Condition}` predicate functions (e.g., `IsRiteNotFound(err error) bool`)

### Domains defined in errors.go
- Session domain (line 1-63): CodeGeneralError, CodeUsageError, etc.
- Rite domain (line 267-303): CodeOrphanConflict, CodeValidationFailed, etc.
- Manifest domain (line 305-353)
- Sync domain (line 363-452)
- Rite errors (line 454-549): CodeRiteNotFound, CodeBorrowConflict, CodeBudgetExceeded, etc.

### Error Propagation Pattern

The dominant propagation pattern in cmd layer (e.g., `internal/cmd/session/create.go`):

```go
if err != nil {
    err := errors.Wrap(errors.CodeGeneralError, "human readable message", err)
    printer.PrintError(err)
    return err
}
```

Note: **the same variable `err` is re-declared** with `:=` to replace the stdlib error with a `*errors.Error`. This is intentional—both the printer output AND the return value are the structured error.

### Boundary Handling (CLI output)

At CLI boundaries, the `output.Printer.PrintError(err error)` method is always called BEFORE returning. If the error implements `JSON() string` (as `*errors.Error` does), JSON format outputs via that method. Text format outputs `"Error: {message}\n"`.

The main entry point (`cmd/ari/main.go`) calls `errors.GetExitCode(err)` to extract the numeric exit code for `os.Exit`.

Hook commands always force JSON output (see `internal/cmd/hook/hook.go`, line 112: `return output.NewPrinter(output.FormatJSON, ...)`).

### Non-critical Error Logging

For non-fatal errors in pipeline stages (e.g., `internal/materialize/materialize.go`):

```go
log.Printf("Warning: failed to detect provenance divergence: %v", err)
```

Standard `log.Printf` is used for warnings that should not abort the operation.

### Mixed stdlib fmt.Errorf Usage

Packages in `internal/know/`, `internal/materialize/` (non-cmd parts), and `internal/inscription/` occasionally use `fmt.Errorf("%w", err)` for wrapping where they do not need structured exit codes. Example from `internal/know/know.go`:

```go
return nil, fmt.Errorf("read .know/ directory: %w", err)
```

This coexists with the custom error system but is more common in infrastructure-layer packages that are not at CLI boundaries.

## File Organization

### Top-Level Structure

```
cmd/ari/main.go          — entry point, minimal logic only
embed.go                 — module-root embed.FS declarations (package knossos)
internal/                — all implementation packages
```

`cmd/ari/main.go` is explicitly minimal: "It contains minimal logic - all command implementations are in `internal/cmd/`."

### The cmd/internal Split

- `cmd/ari/main.go` — only wires version, embedded assets, and calls `root.Execute()`
- `internal/cmd/root/root.go` — root cobra command, global flags, subcommand registration
- `internal/cmd/{domain}/` — one directory per command group (session, hook, rite, sync, etc.)

Each command group directory has:
- `{domain}.go` — top-level `New{Domain}Cmd()` constructor, shared `cmdContext`, helper functions
- One `.go` file per subcommand (e.g., `create.go`, `park.go`, `resume.go`)
- `_test.go` files co-located

### internal Package Conventions

Packages in `internal/` are organized by **domain concern**, not by layer. Each package has a clear single responsibility stated in the package-level doc comment.

Common patterns:

| Filename convention | Contents |
|---|---|
| `{package}.go` or the primary domain file | Main types and constructors |
| `types.go` | Types-only files when type count is large (e.g., `internal/naxos/types.go`, `internal/sails/color.go`) |
| `errors.go` | Domain-local error constants when package has its own — though most errors live centrally in `internal/errors/` |
| `context.go` | Context/state types (session, materialize) |
| `status.go` | Status/enum types and normalization |
| `id.go` | ID generation and validation |
| `fsm.go` | Finite state machine logic |

Large packages are split by concern into named files. The `internal/materialize/` package demonstrates this:
- `materialize.go` — main orchestration
- `source.go` — source resolution re-export (aliases to sub-package)
- `sync_types.go` — sync pipeline types
- `mena.go`, `hooks.go` — re-export from sub-packages
- `collision.go`, `org_scope.go`, `user_scope.go` — scope-specific logic

### Sub-packages

When a package grows large or needs to be independently testable, it is factored into sub-packages under the main package directory. Examples:

- `internal/materialize/source/` — rite source resolution
- `internal/materialize/mena/` — mena projection types
- `internal/materialize/userscope/` — user scope sync
- `internal/materialize/orgscope/` — org scope sync
- `internal/hook/clewcontract/` — Clew Contract event types

The parent package re-exports types from sub-packages via type aliases and `var` re-exports for backward compatibility (`source.go`, `mena.go`, `hooks.go` files use `type X = subpkg.X` pattern).

### internal/ Boundary Philosophy

- `internal/cmd/` — CLI concerns only, imports all other internal packages
- `internal/errors/` — NO internal imports (base package imported by all others)
- `internal/fileutil/` — minimal stdlib-only utilities
- `internal/registry/` — explicitly declared LEAF: "This is a LEAF package — it imports only stdlib. No internal/ imports."
- `internal/mena/source.go` — also declared LEAF

The LEAF package comment convention is: `// This is a LEAF package — it imports only stdlib...`

### Generated Code

`embed.go` at the module root uses `//go:embed` directives. This file is part of package `knossos` (not `main`), enabling embedded assets to be imported by `cmd/ari/main.go`. The package name matches the Go module name `github.com/autom8y/knossos`.

## Domain-Specific Idioms

### Typed String Enums (Universal Pattern)

Every domain has typed string constants. The pattern is `type X string` + constant block + `IsValid()` + `String()` methods:

```go
type Status string
const (
    StatusNone     Status = "NONE"
    StatusActive   Status = "ACTIVE"
    StatusParked   Status = "PARKED"
    StatusArchived Status = "ARCHIVED"
)
func (s Status) String() string { return string(s) }
func (s Status) IsValid() bool { switch s { ... } }
```

This pattern appears across: `Status` (session), `Color` (sails), `Phase` (session/fsm), `OwnerType` (inscription/provenance), `Format` (output), `SyncScope` (materialize), `HookEvent` (hook), `EventType` (clewcontract), `RiteForm` (rite), and many others.

The ALL_CAPS string values are idiomatic across session, sails, and rite domains. Lowercase is used for output-facing strings (e.g., `"text"`, `"json"`, `"rite"`, `"user"`).

### Finite State Machine (FSM) Pattern

The session domain uses an explicit FSM with a transition map. The FSM validates transitions rather than allowing arbitrary state changes:

```go
type FSM struct {
    transitions map[Status][]Status
}
func NewFSM() *FSM { ... }
func (f *FSM) CanTransition(from, to Status) bool { ... }
func (f *FSM) ValidateTransition(from, to Status) error { ... }
```

Location: `internal/session/fsm.go`.

### Options Struct + Result Struct Pattern

Operations that have many configuration parameters use `Options` structs. Operations that return complex outcomes use `Result` structs. These are never merged into one type.

```go
type Options struct {
    Force             bool
    DryRun            bool
    RemoveAll         bool
    ...
}
type Result struct {
    Status          string
    OrphansDetected []string
    ...
}
func (m *Materializer) MaterializeWithOptions(riteName string, opts Options) (*Result, error)
```

`internal/materialize/materialize.go` lines 22-46.

### With* Fluent Builder Methods

The `Materializer` struct uses `With*` methods that return the receiver for chaining embedded filesystem injection:

```go
func (m *Materializer) WithEmbeddedFS(fsys fs.FS) *Materializer { ... }
func (m *Materializer) WithEmbeddedTemplates(fsys fs.FS) *Materializer { ... }
func (m *Materializer) WithEmbeddedAgents(fsys fs.FS) *Materializer { ... }
func (m *Materializer) WithEmbeddedMena(fsys fs.FS) *Materializer { ... }
```

This pattern is used only for the Materializer (not universal).

### Type Alias Re-export Pattern

When a sub-package is created from a larger package, the parent re-exports types for backward compatibility:

```go
// internal/materialize/source.go
type (
    SourceType     = source.SourceType
    RiteSource     = source.RiteSource
    ResolvedRite   = source.ResolvedRite
    SourceResolver = source.SourceResolver
)
const (
    SourceProject  = source.SourceProject
    ...
)
var NewSourceResolver = source.NewSourceResolver
```

This pattern keeps imports stable when internal refactoring occurs.

### Materializer Pipeline Steps

The `MaterializeWithOptions` function uses numbered comment steps (1-10) to document the pipeline stages inline. This is a code documentation idiom unique to the materialize package:

```go
// 1. Resolve rite source using 4-tier resolution
// 2. Ensure .claude/ directory exists
// 3. Handle orphans before materializing agents
// 4. Generate agents/ directory from rite
...
```

### Atomic File Writes

All file writes use `fileutil.AtomicWriteFile` (temp + rename) or `fileutil.WriteIfChanged` (idempotent). Direct `os.WriteFile` is never used for production state files. Location: `internal/fileutil/fileutil.go`.

### Receiver Naming

Single-letter receivers matching the first letter of the type are used consistently:
- `(p *Printer)` for Printer
- `(m *Manager)` for Manager types (lock, materialize, worktree)
- `(r *Resolver)` for Resolver
- `(c *Context)` for Context
- `(f *FSM)` for FSM

### Knossos Mythology Naming

Package and type names are drawn from Greek mythology, specific to the Knossos platform:
- **rite** — a practice bundle (workflow configuration)
- **mena** — combined dromena + legomena (slash commands + skills)
- **dromena** — executable slash commands (`.dro.md`)
- **legomena** — reference skills (`.lego.md`)
- **sails** — confidence signal (WHITE/GRAY/BLACK)
- **clew** — event logging contract (from Ariadne's thread)
- **naxos** — orphan cleanup (from Theseus abandoning Ariadne at Naxos)
- **inscription** — CLAUDE.md region management system
- **theoros** — audit/observation agent
- **satellite** — user-owned project (as opposed to knossos-platform-owned)

### Status/Color ALLCAPS Convention

Session statuses (`ACTIVE`, `PARKED`, `ARCHIVED`), sails colors (`WHITE`, `GRAY`, `BLACK`), and complexity tiers (`PATCH`, `MODULE`, `SYSTEM`, `INITIATIVE`, `MIGRATION`) are ALL_CAPS strings. These are domain-specific semantic states, not Go identifiers.

### Provenance Tracking

The materialization pipeline uses a `Collector` interface for accumulating file provenance during pipeline execution, then saves it at pipeline completion. This is an observer pattern specific to the sync pipeline:

```go
// internal/provenance/collector.go
type Collector interface {
    Record(relativePath string, entry *ProvenanceEntry)
    Entries() map[string]*ProvenanceEntry
}
```

`NullCollector` is provided for dry-run and minimal modes. Location: `internal/provenance/collector.go`.

### Testable Function Variables

The `know` package uses a package-level `var` to allow dependency injection in tests:

```go
// internal/know/know.go
var gitDiffNameOnly = defaultGitDiffNameOnly
func defaultGitDiffNameOnly(fromHash, toHash string) ([]string, error) { ... }
```

Tests replace `gitDiffNameOnly` with a stub. This is a targeted pattern, not universal.

## Naming Patterns

### Constructors

All constructors follow `New{TypeName}(...)` convention:
- `NewFSM() *FSM`
- `NewContext(initiative, complexity, rite string) *Context`
- `NewManager(locksDir string) *Manager`
- `NewPrinter(format Format, out, errOut io.Writer, verbose bool) *Printer`
- `NewResolver(projectRoot string) *Resolver`
- `NewMaterializer(resolver *paths.Resolver) *Materializer`

When a constructor takes an explicit source variant, it is `New{TypeName}With{Qualifier}`:
- `NewMaterializerWithSource`
- `NewGeneratorWithValidator`
- `NewGeneratorWithFS`
- `NewPipelineWithPaths`
- `NewBackupManagerWithTarget`

### Default* Functions

Functions returning default configuration or state objects:
- `DefaultConfig() ScanConfig` (naxos)
- `DefaultTimeout = 10 * time.Second` (lock, as const)

### Err* Constructors

Error constructor functions are prefixed with `Err` + domain noun + condition:
- `ErrProjectNotFound() *Error`
- `ErrSessionNotFound(sessionID string) *Error`
- `ErrSessionExists(existingID string, status string) *Error`
- `ErrLifecycleViolation(from, to string, reason string) *Error`
- `ErrRiteNotFound(riteName string) *Error`

All live in `internal/errors/errors.go`.

### Is* Predicates

Error predicate functions are prefixed with `Is`:
- `IsNotFound(err error) bool`
- `IsLifecycleError(err error) bool`
- `IsRiteNotFound(err error) bool`
- `IsMergeConflict(err error) bool`

### Package Naming

All packages use singular nouns or short nouns (Go convention). No plural package names observed. Packages use lowercase, no underscores:
- `session`, `hook`, `lock`, `paths`, `output`, `errors`, `registry`
- `materialize` (verb as noun), `inscription`, `provenance`, `fileutil`

`fileutil` is the only exception using compound without separator. The `cmd` sub-directories also use short nouns: `session`, `hook`, `rite`, `sync`.

### File Naming

Files within a package are named by concern, not type:
- Core logic: `{packagename}.go` or descriptive noun (e.g., `context.go`, `status.go`, `fsm.go`)
- CLI wrapper: `{subcommand}.go` (e.g., `create.go`, `park.go`, `resume.go`)
- Tests: `{file}_test.go` always co-located with the file under test
- Integration tests: `{concept}_integration_test.go` or `integration_test.go`

### Type Naming

- Typed string enums: `{Noun}` (e.g., `Status`, `Phase`, `Color`, `Format`)
- Configuration/input structs: `{Noun}Options` or `{Verb}Options` (e.g., `MaterializeOptions`, `SyncOptions`, `CreateOptions`)
- Output/result structs: `{Noun}Result` or `{Noun}Output` (e.g., `SyncResult`, `CreateOutput`, `RiteScopeResult`)
- Interface types: plain nouns (e.g., `Collector`, `Tabular`, `Textable`)
- Constant groups: named by domain prefix (e.g., `StatusActive`, `ColorWhite`, `ScopeAll`, `ResourceAgents`)

### Acronym Conventions

- `ID` (not `Id`) — e.g., `SessionID`, `session_id`
- `URL` (not `Url`) — not observed in exported names but aligns with Go stdlib
- `CC` (not `Cc`) for Claude Code abbreviation — e.g., `CCMapDir()`, `CCMapOrphans`
- `YAML` (not `Yaml`) as `yaml` in package names (from dep `gopkg.in/yaml.v3`)
- `JSON` (not `Json`) — e.g., `FormatJSON`, `printJSON`

### Constants Naming

String constants for error codes: `Code{Domain}{Condition}` (e.g., `CodeRiteNotFound`, `CodeSessionExists`)
Integer constants for exit codes: `Exit{Condition}` (e.g., `ExitFileNotFound`, `ExitBudgetExceeded`)

Typed string constants use ALL_CAPS values for domain states and SCREAMING_SNAKE_CASE for error codes. Lowercase string values are used for non-critical config strings (e.g., `"success"`, `"skipped"`, `"kept"`).

### Anti-Patterns to Avoid Spreading

1. `ValidationMode` uses `iota` (`internal/agent/validate.go`) — only one instance; prefer typed string constants.
2. `RefCategory` uses `iota` (`internal/registry/registry.go`) — appropriate for int-keyed enum, but string-typed enums are more common.
3. Inline `map[string]any` is used in VerboseLog calls and some error details — not a struct pattern, intentional for flexibility.

## Knowledge Gaps

1. **`internal/mena/` package** — not deeply examined. The mena source resolution logic (SourceChain, MenaProjection) was seen in types but not traced in detail.
2. **`internal/cmd/hook/writeguard.go`** and `SectionClass` iota usage — not read; mentioned in search results only.
3. **`internal/tribute/`** (TRIBUTE.md generation) — types seen in `ls`, not read in detail.
4. **`internal/agent/`** package internals — types.go referenced but not deeply explored.
5. **Test fixture patterns** — test file naming is documented but specific fixture conventions (table-driven tests, testdata/ directories) not examined. Test conventions are scoped to the `test-coverage` domain.
6. **`internal/sync/state.go`** — the sync state file format not examined.
7. **`internal/config/home.go`** — config loading patterns not explored beyond viper usage in root.go.
