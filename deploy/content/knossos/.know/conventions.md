---
domain: conventions
generated_at: "2026-03-27T19:57:42Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "5501b0aa"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "1f2c9d187ac50eb67dffec49dddc3dd9217e4be0cb56e62cda1bd1d52dd7c00f"
---

# Codebase Conventions

> Reference document for any CC agent contributing Go code to the knossos/ari project. Read this before writing or modifying code in `cmd/` or `internal/`.

**Scope:** `cmd/`, `internal/` -- 718 Go source files, 45 top-level packages in `internal/`

---

## Error Handling Style

### Philosophy

The project maintains a dedicated `internal/errors` package that is the **canonical error infrastructure**. All domain-significant errors flow through this package. Standard library `fmt.Errorf` with `%w` wrapping is used for infrastructure/utility errors in lower-level packages.

### Error Creation

Three creation functions in `internal/errors/errors.go`:

```go
errors.New(code string, message string) *Error
errors.NewWithDetails(code string, message string, details map[string]any) *Error
errors.Wrap(code string, message string, cause error) *Error
```

The `*Error` type carries:
- `Code` -- a `SCREAMING_SNAKE_CASE` string constant from the same file
- `Message` -- human-readable text
- `Details` -- `map[string]any` (serialized to JSON; key `"cause"` holds the wrapped error string)
- `ExitCode` -- integer derived automatically from `Code` via `exitCodeForCode()`
- `cause` -- unexported `error` field enabling Go chain traversal via `errors.As`/`errors.Is`

### Error Code System

Exit codes (0-21) are defined as integer constants (`ExitSuccess`, `ExitGeneralError`, ..., `ExitSyncNotConfigured`). Error codes are SCREAMING_SNAKE_CASE string constants (`CodeGeneralError`, `CodeUsageError`, ..., `CodeQualityGateFailed`). Each code maps to exactly one exit code through `exitCodeForCode()`.

**Domain grouping**: codes are declared in blocks per domain -- session-domain, rite-domain, manifest-domain, sync-domain, serve-domain. Each domain block has its own constructors (`ErrRiteNotFound`, `ErrBorrowConflict`, etc.) and type-testing predicates (`IsRiteNotFound`, `IsBorrowConflict`, etc.).

### Error Wrapping Convention

Two wrapping patterns coexist:

1. **Domain errors** (`errors.Wrap`): used when the error has a meaningful code and the cause is attached both as a string in `Details["cause"]` and as an unexported field for chain traversal. This is the preferred pattern for errors that surface to CLI output.

2. **`fmt.Errorf("%w")`**: used in lower-level packages (session, slack, trust, registry, paths) for infrastructure errors that will be caught and rewrapped or logged upstream.

Both wrapping styles support `errors.As` traversal.

### Error Propagation Style

Immediate return (`return nil, err` / `return err`) is the universal pattern. No error aggregation at call sites. No `defer`-cleanup with error assignment. Errors are returned up the call stack to the command layer.

### The `handledError` Sentinel

`internal/errors/errors.go` defines a private `handledError` wrapper. Command code calls `errors.Handled(err)` after printing an error to signal that the error was already displayed to the user. The root command uses `errors.IsHandled(err)` to skip re-printing.

### CLI Error Boundary

All command `RunE` functions return errors. The root command catches returned errors and calls `printer.PrintError(err)`. `PrintError` checks whether the error implements `interface{ JSON() string }` -- if so, it serializes as structured JSON; otherwise it formats as `"Error: {message}\n"` to stderr.

The helper `common.PrintAndReturn(printer, err)` at `internal/cmd/common/group.go` prints then wraps with `errors.Handled(err)`, preventing duplicate output.

### Sentinel Errors (Frontmatter Package)

`internal/frontmatter/errors.go` uses plain `var ErrMissing* = errors.New(...)` sentinel pattern (stdlib `errors.New`, not domain `errors.New`) for parse errors. This is the only package using sentinel variables.

### Logging

Structured logging uses `log/slog` (stdlib). Errors are not logged at call sites -- they are returned. Logging happens at pipeline entry points (e.g., `internal/slack/handler.go`). `slog.Warn(msg, "key", val)` style -- key-value pairs, not printf.

---

## File Organization

### One Concern Per File

Files are named by their single responsibility. Examples from `internal/session/`:

| File | Contents |
|------|----------|
| `status.go` | `Status` typed string + `IsValid`, `IsTerminal`, `NormalizeStatus` |
| `fsm.go` | `FSM` struct + `Phase` type |
| `context.go` | `Context` struct (the core session data model) |
| `complexity.go` | `Complexity` type |
| `id.go` | Session ID generation |
| `timeline.go` | Timeline entry management |
| `snapshot.go` | Snapshot rendering (Markdown and JSON) |
| `events_read.go` | Reading events from events.jsonl |

### Test Files Colocated

Every `foo.go` has a corresponding `foo_test.go` in the same directory and same package (or `package foo_test` for black-box tests). Integration tests use `*_integration_test.go` suffix. Scar regression tests live in `scar_regression_test.go`.

### Subdirectory-as-Subpackage Pattern

Large packages split concern-groups into subdirectories:

```
internal/materialize/
  compiler/    (package compiler)
  hooks/       (package hooks)
  mena/        (package mena)
  orgscope/    (package orgscope)
  procession/  (package procession)
  source/      (package source)
  userscope/   (package userscope)
```

### cmd/ vs internal/ Separation

`cmd/ari/main.go` is the sole entry point -- imports `internal/cmd/root` and calls `root.Execute()`. All business logic lives in `internal/`. The `internal/cmd/` tree contains one file per CLI subcommand. No business logic in `cmd/`.

### Special Files by Name Convention

| Filename | Contents |
|----------|----------|
| `types.go` | Type declarations for the package |
| `errors.go` | Sentinel errors or error types local to the package |
| `doc.go` | Package-level documentation |
| `*_test.go` | Tests (same-package or black-box) |
| `*_integration_test.go` | Integration tests |
| `scar_regression_test.go` | Scar-labeled regression tests |

### Package Doc Comments

Every package file opens with `// Package {name} provides ...`. Consistently applied across `internal/`.

---

## Domain-Specific Idioms

### 1. Typed String Enums (Dominant Pattern)

Almost all enumerated values use `type Foo string` with `const` blocks. Each type gets `String()`, `IsValid()`, and sometimes `IsTerminal()` methods.

```go
type Status string
const (
    StatusNone     Status = "NONE"
    StatusActive   Status = "ACTIVE"
    StatusParked   Status = "PARKED"
    StatusArchived Status = "ARCHIVED"
)
func (s Status) IsValid() bool { ... }
func (s Status) IsTerminal() bool { return s == StatusArchived }
```

Seen in: `session.Status`, `sails.Color`, `sails.ModifierType`, `sails.ProofStatus`, `artifact.ArtifactType`, `artifact.Phase`, `output.Format`. Never use raw `string` for domain states.

Iota-based int enums appear only for non-serialized internal state.

### 2. Options Struct Pattern (Not Functional Options)

Configuration for major operations is passed as a value struct (`Options`), not functional options:

```go
type Options struct {
    Force   bool
    DryRun  bool
    ...
}
func (m *Materializer) MaterializeWithOptions(riteName string, opts Options) (*Result, error)
```

Functional options (`Option func(*Server)`) exist in `internal/serve/server.go` but are the exception.

### 3. `New*` / `New*With*` Constructor Families

Constructors are named `New{Type}` (basic) and `New{Type}With{Context}` (injectable/testable). The `With` variant takes explicit paths or dependencies.

```go
func NewPipeline(projectRoot string) *Pipeline           // production
func NewPipelineWithPaths(inscriptionPath, ...) *Pipeline // testable
```

### 4. SCAR Regression Test Naming

Regression tests for documented scars are named `TestSCAR{NNN}_{Description}` in `scar_regression_test.go` files. Each test has a comment citing the SCAR number.

### 5. HA / GAP / DEBT / TDD Code Annotations

Inline comments use a structured annotation system:

| Annotation | Meaning |
|-----------|---------|
| `// HA-{TAG}:` | Harness-agnostic exception |
| `// HA-TEST:` | Test fixture harness-specific content |
| `// HA-SELF:` | Lint package exempt from its own rule |
| `// GAP-{N}:` | Known gap or temporary workaround |
| `// DEBT-{NNN}:` | Tracked technical debt item |
| `// TDD Section {X.Y}:` | References a Technical Design Document section |
| `// SCAR-{NNN}:` | References a known regression scar |
| `// TENSION-{NNN}:` | Documents a design tension |

### 6. Interface Compliance Guards

Packages use `var _ Interface = (*Impl)(nil)` at package level to assert interface compliance at compile time.

### 7. Atomic File Writes

All persistent file writes use `fileutil.AtomicWriteFile(path, content, perm)`. Writes to a temp file, syncs, then renames. Direct `os.WriteFile` is used only in tests.

### 8. Output Interfaces (Tabular / Textable)

Output types implement either `Tabular` (`Headers()`, `Rows()`) or `Textable` (`Text()`) from `internal/output/output.go`. Never `fmt.Print` directly in commands.

### 9. Dual YAML+JSON Struct Tags

All persisted types use both `yaml:` and `json:` tags:
```go
SessionID string `yaml:"session_id" json:"session_id"`
```

### 10. Knossos Marker Regions in CLAUDE.md

The `inscription` package manages CLAUDE.md files via structured markers (`<!-- KNOSSOS:START:{name} -->` / `<!-- KNOSSOS:END:{name} -->`). Region ownership: `"knossos"` (always overwritten), `"satellite"` (never overwritten), `"regenerate"` (regenerated from state).

### 11. SchemaVersion Field in Persisted Types

All persisted structs include `SchemaVersion string` for migration detection. Versions are semver strings like `"2.0"`, `"2.1"`, `"2.3"`.

---

## Naming Patterns

### Package Names

Singular nouns, lowercase, matching the directory name: `session`, `agent`, `artifact`, `mena`, `manifest`, `procession`, `tribute`, `inscription`. Sub-packages: `userscope`, `orgscope`, `clewcontract`. No plural package names.

### Type Names

- `*Options` -- configuration struct for an operation
- `*Result` -- outcome struct
- `*Resolver` -- contextual lookup service
- `*Validator` -- validation service
- `*Materializer` -- build/generation service
- `*Config` -- `{Domain}Config` for subsystem configuration
- `*Output` -- `{Command}Output` for CLI result types
- Domain state types: named after the concept (`Status`, `Color`, `Phase`)

### Constructor Names

Standard Go `New*` constructors. Named domain error constructors use `Err*` prefix. `New{Type}WithPaths(...)` is the dominant testable variant pattern.

### Constant Names

- Domain state values: typed + prefixed with type name (`StatusActive`, `ColorWhite`)
- Error codes: `SCREAMING_SNAKE_CASE` string constants
- Exit codes: `Exit{Name}` integer constants
- Default values: `Default{Thing}` -- `DefaultTimeout`, `DefaultTriageModel`

### Acronym Conventions

All-caps in exported names: `SessionID`, `JSON`, `YAML`, `XDG`, `URL`, `CLI`. No instances of `Id` in exported type definitions.

### Receiver Names

Single-letter receivers matching the type's first letter: `(s *Session)`, `(m *Materializer)`, `(p *Printer)`, `(r *Resolver)`.

### File Names

Concern-per-file: `types.go`, `validate.go`, `manifest.go`, `errors.go`. Integration tests: `{concern}_integration_test.go`.

### Naming Anti-Patterns to Avoid

- Do not use `Id` (lowercase d) in exported identifiers -- use `ID`.
- Do not name option structs `FooOptions` inside the package -- use unqualified `Options`.
- Do not use `log` package -- `log/slog` is standard.
- Do not use `os.WriteFile` directly for persistent writes -- use `fileutil.AtomicWriteFile`.
- Do not use `helper.go`, `util.go`, `common.go` inside domain packages.

---

## Knowledge Gaps

1. **`internal/perspective` package** -- defines audit-related types whose full role was not explored.
2. **`internal/resolution` package** -- full semantics of the `resolution.Chain` abstraction not captured here.
3. **Config loading conventions** -- how `viper` config values flow from root flags into command implementations not traced.
4. **`internal/reason` pipeline** -- LLM reasoning pipeline package structure observed but not traced in depth.
5. **Functional options in `serve`** -- `internal/serve/server.go` uses `type Option func(*Server)` but was not fully explored.
