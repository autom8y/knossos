---
domain: conventions
generated_at: "2026-03-06T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "3847e28"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "f6d34ca03d89c8f6e949f8b89f44463c9ce80aa2cb884aa1d31c8abbd3f054ba"
---

# Codebase Conventions

## Error Handling Style

### Error Creation

The project uses a **custom domain error type** defined in `internal/errors/errors.go`, not raw `errors.New` or bare `fmt.Errorf`. The `errors.Error` struct carries:

- `Code string` — a SCREAMING_SNAKE_CASE string constant (e.g., `"SESSION_NOT_FOUND"`, `"LIFECYCLE_VIOLATION"`)
- `Message string` — human-readable text
- `Details map[string]any` — structured context (JSON-serializable)
- `ExitCode int` — process exit code (automatically derived from Code via `exitCodeForCode`)
- `cause error` (unexported) — supports Go error chain traversal via `Unwrap`

Three constructors exist:

```go
errors.New(code, message)                      // plain error
errors.NewWithDetails(code, message, details)   // with structured context
errors.Wrap(code, message, cause)               // wraps an underlying error
```

Domain-specific convenience constructors are the normal usage, not the raw constructors:

```go
errors.ErrSessionNotFound(sessionID)
errors.ErrLifecycleViolation(from, to, reason)
errors.ErrLockTimeout(lockPath, lockMeta)
errors.ErrRiteNotFound(riteName)
// ... ~20 total domain constructors
```

`fmt.Errorf` with `%w` is used (70 occurrences across 23 files) for lightweight wrapping within internal package logic, especially when adding context while threading errors back to a boundary. The custom `errors.Wrap` is preferred at any boundary that needs a structured code.

### Error Checking and Predicate Functions

The `errors` package provides typed predicate functions for all domain error categories:

```go
errors.IsNotFound(err)        // CodeFileNotFound, CodeSessionNotFound, CodeProjectNotFound
errors.IsLifecycleError(err)  // CodeLifecycleViolation
errors.IsMergeConflict(err)   // CodeMergeConflict
errors.IsRiteNotFound(err)    // CodeRiteNotFound
errors.IsBudgetExceeded(err)  // CodeBudgetExceeded
// ... ~15 total predicate functions
```

These use `errors.As` internally and work through `fmt.Errorf("%w", ...)` chains.

When stdlib `errors.As/Is` is needed in non-`errors` package code, the import alias is always `stderrors "errors"` (8 files use this pattern) to avoid shadowing the internal `errors` package.

### Error Propagation

The dominant pattern is **immediate return** — 1,023 occurrences of `return err` or `return nil, err` across 188 files. No error aggregation patterns (multi-error) are used — each function returns after the first failure.

### Error Handling at CLI Boundaries

The `output.Printer` handles all user-facing error display:

```go
printer.PrintError(err)  // to stderr, format-aware
```

- **Text mode**: `"Error: <message>\n"` to stderr
- **JSON mode**: Calls `err.(interface{JSON() string}).JSON()` if available (the `errors.Error` type implements this), producing `{"error": {"code": "...", "message": "...", "details": {...}}}`

The main entry point (`cmd/ari/main.go`) extracts the exit code:

```go
os.Exit(errors.GetExitCode(err))  // walks chain via errors.As
```

Exit codes are defined as numeric constants in the 0-21 range (defined in `errors.go`).

### Logging

`log/slog` is used for structured diagnostic logging within materialization logic (16 files, 45 calls concentrated in `internal/materialize/`). It is NOT used for user-visible output — that goes through `output.Printer`. The rest of the codebase uses `printer.VerboseLog(level, msg, fields)`. Do not add `slog` to other packages.

## File Organization

### cmd/ vs internal/ Separation

- `cmd/ari/main.go` — minimal entry point: sets version, calls `root.Execute()`, converts exit code. Zero domain logic.
- `internal/cmd/` — Cobra command implementations, one sub-package per command group
- `internal/` (non-cmd) — domain logic packages, no Cobra dependency

**Rule**: Business logic never lives in `internal/cmd/`. Commands call into `internal/` domain packages.

### internal/cmd/ Package Pattern

Each command group has:
- `{group}.go` — `New{Group}Cmd` (exported) that builds the parent Cobra command and adds subcommands
- One file per subcommand — `{verb}.go`, containing an unexported `new{Verb}Cmd` constructor
- Command file naming follows the verb: `create.go`, `park.go`, `resume.go`, `wrap.go`, `status.go`

The `{group}.go` file defines the shared `cmdContext` struct that embeds `common.BaseContext` or `common.SessionContext`, plus a package-level `getPrinter()` helper.

### internal/cmd/common/ Pattern

Shared context inheritance via struct embedding:

```
common.BaseContext          — Output, Verbose, ProjectDir pointers
  └── common.SessionContext — + SessionID pointer
        └── cmdContext      — package-local, may add more fields
```

Methods on `BaseContext` (`GetPrinter`, `GetResolver`, `GetActiveRite`) and `SessionContext` (`GetSessionID`, `GetLockManager`) are inherited by all command packages via embedding.

### Domain Package File Organization

Each domain package uses a consistent layout:

| File name | Contents |
|---|---|
| `types.go` | Type definitions, constants, interfaces for the domain |
| `{noun}.go` | Primary package file, often houses the main type and its constructor |
| `{verb}.go` or `{noun}_{verb}.go` | Operations on the primary type |

The `materialize` package uses `materialize_{noun}.go` naming for each concern:
- `materialize.go` — core `Materializer` type, `Options`, `Result`
- `materialize_agents.go` — agent file generation
- `materialize_mena.go` — mena (skills/commands) file generation
- `materialize_claudemd.go` — CLAUDE.md section generation
- `materialize_settings.go` — settings.json generation
- `materialize_rules.go` — rules processing

Sub-domains with meaningful complexity get their own sub-package: `materialize/mena/`, `materialize/hooks/`, `materialize/userscope/`, `materialize/orgscope/`, `materialize/source/`.

### Constants and Variables

Constants live in the same file as the type they describe, not in a separate `constants.go`. Enum-like constants always use a typed string (`type Status string`) with a `const (...)` block.

### `init()` Functions

`init()` functions appear only in `cmd/root/root.go` to register Cobra flags and subcommands. No `init()` functions in domain packages.

### Internal Package Boundary

`internal/` is used at the project level (the entire `internal/` tree). No `internal/` sub-packages within packages. The LEAF package comment convention is: `// This is a LEAF package — it imports only stdlib...`

## Domain-Specific Idioms

### The `Resolver` Pattern

`internal/paths.Resolver` is a value type that holds `projectRoot string` and provides all path computations as methods. It is created early in command execution via `common.BaseContext.GetResolver()` and threaded explicitly to functions.

### Typed String Enums with IsValid/IsTerminal Methods

All domain states are typed strings with const blocks and behavior methods:

```go
type Status string
const (StatusNone Status = "NONE"; StatusActive Status = "ACTIVE"; ...)

func (s Status) IsValid() bool { switch s { ... } }
func (s Status) IsTerminal() bool { return s == StatusArchived }
func (s Status) String() string { return string(s) }
```

This pattern appears across: `session.Status`, `sails.Color`, `sails.ModifierType`, `sails.ProofStatus`, `inscription.OwnerType`, `inscription.MarkerDirective`, `session.Phase`.

### The `Options` + `Result` Pair

Functions with multiple tunable behaviors use an `Options` struct as input and a `Result` struct for output. These are NOT functional options — they are plain structs.

### The `output.Tabular` / `output.Textable` Interfaces

All CLI output types implement either:
- `Tabular`: `Headers() []string` + `Rows() [][]string` — for table output
- `Textable`: `Text() string` — for custom text formatting

### KNOSSOS Marker System (Inscription Domain)

CLAUDE.md files use a structured marker syntax: `<!-- KNOSSOS:START {region-name} [options] -->`. Ownership types (`OwnerKnossos`, `OwnerSatellite`, `OwnerRegenerate`) determine sync behavior.

### FSM Pattern for Lifecycle State

The session lifecycle is implemented as an explicit finite state machine in `internal/session/fsm.go` with `CanTransition`, `ValidateTransition`, and `ValidTransitions` methods.

### Event/Hook Contract (`clewcontract` Sub-Package)

The `internal/hook/clewcontract/` sub-package defines typed event records for the session event log with typed constructors and a `BufferedEventWriter`.

### Polymorphic YAML via `json.RawMessage`

Hook stdin payloads use `json.RawMessage` for `ToolInput` and `ToolResponse` fields in `hook.StdinPayload`.

### `slog` Scope

`log/slog` is used ONLY in the `materialize` package tree and `cmd/initialize/init.go`. Do not add `slog` to other packages; use `printer.VerboseLog` instead.

## Naming Patterns

### Package Names

Packages use single-word, lowercase names matching their directory: `session`, `inscription`, `materialize`, `provenance`, `perspective`. The collision between the project's custom `errors` package and stdlib `errors` is resolved by importing stdlib with the alias `stderrors`.

### Type Naming

Exported types follow these suffixes:
- `*Manager` — stateful coordinator: `StateManager`, `BackupManager`, `MetadataManager`
- `*Resolver` — path or source resolution: `Resolver`, `SourceResolver`
- `*Validator` — schema/contract validation: `AgentValidator`, `HandoffValidator`
- `*Writer` — output writing: `EventWriter`, `BufferedEventWriter`
- `*Options` — configuration input struct
- `*Result` — operation outcome struct
- `*Output` — CLI display struct (lives in `output` package)

### Function Naming

- Constructors: `New{TypeName}` (exported) for types used across packages; `new{TypeName}` (unexported) for cobra subcommand constructors within `cmd` packages
- Predicate helpers: `Is{Condition}` for error predicates; `Is{State}()` method for type predicates
- Normalize helpers: `Normalize{Field}` — e.g., `NormalizeStatus`

### Cobra Command Constructor Convention

`New{Group}Cmd(...)` is exported (called from `root.go`). `new{Verb}Cmd(ctx)` is unexported (called from `New{Group}Cmd`). Only the parent group constructor is the public API.

### Constants: SCREAMING_SNAKE_CASE

All constants using string values use SCREAMING_SNAKE_CASE for the constant name AND for the string value itself:

```go
const StatusActive Status = "ACTIVE"
const CodeSessionNotFound = "SESSION_NOT_FOUND"
const ColorBlack Color = "BLACK"
```

### Acronyms in Identifiers

Acronyms follow Go conventions (capitalize entirely or lowercase entirely):
- `MCP` → `MCPServer`, `mcpServerNames`
- `FSM` → `FSM`, `NewFSM`
- `JSON` → `JSON()` method, `printJSON`
- `CC` → prefix for Claude Code context (e.g., `CCSessionID`, `CCMapOrphans`)
- `ID` → always `ID` not `Id` (`SessionID`, `ToolUseID`)

### File Naming

- Primary type file: matches type name lowercased (`fsm.go`, `color.go`, `lock.go`)
- Operations on a type: `{noun}_{verb}.go` in large packages (`materialize_agents.go`)
- Sub-command files: `{verb}.go` in `cmd/` packages (`create.go`, `park.go`)
- Test files: `{source_file}_test.go` always co-located

## Mena Materialization Path Conventions

### Materialized Filename Rules

| Source Pattern | Materialized Output | Rule |
|---|---|---|
| `INDEX.lego.md` | `SKILL.md` | Legomena entry points rename to `SKILL.md` |
| `INDEX.dro.md` | `{name}.md` (promoted to parent) | Dromena entry points promote one level |
| `{name}.lego.md` | `{name}.md` | Extension stripping only |
| `{name}.dro.md` | `{name}.md` | Extension stripping only |

### Namespace Flattening

Source path `mena/{category}/{name}/` flattens to `.claude/{commands|skills}/{name}/` — the `{category}` level is stripped during materialization.

### Content Reference Rules

When referencing materialized paths in source files:
- **Legomena entry points**: Always use `SKILL.md`, never `INDEX.md`
- **Dromena entry points**: Always use `{name}.md` at parent level, never `INDEX.md` in a subdirectory
- **No namespace prefix**: Never include the `{category}/` level in `.claude/` paths

## Knowledge Gaps

1. **`internal/perspective/` package** — reviewed type definitions but not the full assembly logic.
2. **`internal/provenance/` divergence logic** — `divergence.go` and `merge.go` not inspected.
3. **`internal/mena/` walk semantics** — the `walk.go` file and mena resolution rules for mixed dro/lego directories were not read directly.
4. **Goroutine usage** — no concurrency patterns were checked beyond `sync.Once`.
