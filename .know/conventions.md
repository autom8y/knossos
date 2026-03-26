---
domain: conventions
generated_at: "2026-03-26T17:14:25Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "a73d68a6"
confidence: 0.92
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

**Scope:** `cmd/`, `internal/` — 704 Go source files, 46 top-level packages in `internal/`

---

## Error Handling Style

### Philosophy

The project uses a **dual-layer error system**. Domain errors (structured, typed, JSON-serializable) are used at boundaries and propagated upward. Standard library errors (`fmt.Errorf`, `errors.New`) are used for low-level internal operations where structure is not needed.

### Layer 1: Domain Error Package (`internal/errors`)

All domain-meaningful errors are created via `internal/errors` (imported as `errors` in most packages). The package provides:

**`errors.Error` struct** — the canonical error type:
```go
type Error struct {
    Code     string         `json:"code"`
    Message  string         `json:"message"`
    Details  map[string]any `json:"details,omitempty"`
    ExitCode int            `json:"-"`
    cause    error          // unexported, enables errors.As chain traversal
}
```

**Three constructors:**
- `errors.New(code, message)` — basic structured error
- `errors.NewWithDetails(code, message, details)` — with structured context map
- `errors.Wrap(code, message, cause)` — wraps another error; stores cause in `Details["cause"]` (for JSON) and unexported `cause` field (for Go chain traversal via `errors.As`)

**Named error constructors** — every domain has its own `Err*` functions:
- `errors.ErrProjectNotFound()`, `errors.ErrSessionNotFound(id)`, `errors.ErrLockTimeout(path, meta)`
- `errors.ErrLifecycleViolation(from, to, reason)`, `errors.ErrValidationFailed(rite, count, issues)`
- `errors.ErrParseError(path, format, cause)`, `errors.ErrNetworkError(url, cause)`
- Source: `internal/errors/errors.go`

**`Is*` predicate functions** — for error type checking without type assertions:
- `errors.IsNotFound(err)`, `errors.IsLifecycleError(err)`, `errors.IsMergeConflict(err)`
- `errors.IsNetworkError(err)`, `errors.IsRiteNotFound(err)`, `errors.IsBudgetExceeded(err)`
- All implemented via `errors.As` chain traversal — safe through `fmt.Errorf("%w", ...)` wrappers

### Error Code System

Codes are `SCREAMING_SNAKE_CASE` string constants, organized by domain:
- **Session domain:** `GENERAL_ERROR`, `LIFECYCLE_VIOLATION`, `SESSION_EXISTS`, `SESSION_NOT_FOUND`
- **File/schema:** `FILE_NOT_FOUND`, `SCHEMA_INVALID`, `PARSE_ERROR`, `SCHEMA_NOT_FOUND`
- **Sync domain:** `SYNC_STATE_CORRUPT`, `REMOTE_REJECTED`, `NETWORK_ERROR`, `SYNC_NOT_CONFIGURED`
- **Rite domain:** `RITE_NOT_FOUND`, `BORROW_CONFLICT`, `BUDGET_EXCEEDED`, `QUALITY_GATE_FAILED`
- **Serve domain:** `SIGNATURE_INVALID`, `TIMESTAMP_EXPIRED`, `SERVER_START_FAILED`

Exit codes map to codes via `exitCodeForCode()`. Exit codes 0–21 are defined.

### Error Propagation Convention

**Immediate return pattern** — the dominant style:
```go
if err != nil {
    return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid YAML frontmatter", err)
}
```
Evidence: 538 uses of `errors.New`/`errors.Wrap` across 123 files; 157 uses of `fmt.Errorf %w` across 48 files.

**When `fmt.Errorf` is acceptable:** Internal packages that are not on API boundaries (e.g., `internal/ledge/promote.go`, `internal/cmd/land/synthesize.go`) use `fmt.Errorf("%w", err)` for simple wrapping without domain structure.

**When sentinel errors are used:** Low-level leaf packages like `internal/frontmatter` use `errors.New("missing frontmatter opening delimiter")` (stdlib `errors.New`, not domain `errors.New`).

### Boundary Handling

**CLI boundary (`cmd/` layer):**
1. Commands call `common.PrintAndReturn(printer, err)` to print and mark as handled
2. `main.go` checks `errors.IsHandled(err)` before printing — prevents double printing
3. `errors.GetExitCode(err)` traverses the chain via `errors.As` to extract the correct exit code

**`errors.Handled` sentinel** — wraps errors that were already printed to the user. Main uses `IsHandled` to decide whether to print.

**Output boundary (`internal/output`):**
- `Printer.PrintError(err)` checks if `err` implements `JSON() string` (domain errors do) and outputs structured JSON when `--output json`
- Text mode: `Error: {message}` to stderr
- Verbose logging: `slog` JSON to stderr via `Printer.VerboseLog`

### Logging

Structured logging uses `log/slog` (stdlib). The `observe` package configures JSON output to stderr:
- Source: `internal/observe/logging.go`
- `observe.ConfigureStructuredLogging(level)` — sets default slog handler to JSON, level from env
- Call site: `slog.Warn(msg, "key", val)` style — key-value pairs, not printf
- Used selectively (non-fatal conditions, warnings) — not for all operations

---

## File Organization

### Top-Level Split: `cmd/` vs `internal/`

- `cmd/ari/main.go` — single entry point, minimal logic (wires version/assets, calls `root.Execute()`)
- `internal/cmd/` — all CLI command implementations, organized by command group
- `internal/` — all domain logic, strictly no CLI code

### `internal/cmd/` Pattern

Each command group is a subdirectory with a coordinator file plus per-subcommand files:
```
internal/cmd/session/
    session.go         # cobra command registration and group setup
    create.go          # ari session create
    park.go            # ari session park
    query.go           # ari session query
    wrap.go            # ari session wrap
    ...
```
The coordinator file adds subcommands to the group's cobra `Command`. Individual files contain `runXxx()` functions that are the cobra `RunE` implementations.

### `internal/` Domain Package Pattern

Each domain package contains files grouped by concern:
- `types.go` — exported types (structs, type aliases, interfaces) — present in 16 packages
- `validate.go` — validation logic — present in 10 packages
- `manifest.go` — YAML/JSON schema handling — present in 7 packages
- `context.go` — context/state for a subsystem — present in 7 packages
- `generator.go` — content generation logic — present in 5 packages
- `frontmatter.go` — frontmatter parsing — present in 5 packages
- `errors.go` — package-specific sentinel errors

This is a **concern-per-file** approach, not alphabetical. Naming follows the domain noun.

### Sub-package Directories

Large packages use subdirectories for distinct sub-concerns:
- `internal/materialize/` — 60+ files, splits into `compiler/`, `hooks/`, `mena/`, `orgscope/`, `procession/`, `source/`, `userscope/`
- `internal/search/` — splits into `bm25/`, `content/`, `fusion/`, `knowledge/` (which further splits into `embedding/`, `graph/`, `summary/`)
- `internal/reason/` — splits into `context/`, `intent/`, `response/`

### `internal/` Boundaries

- **LEAF packages** (documented with comment `// This is a LEAF package — it imports only stdlib`):
  - `internal/registry` — denial-recovery platform references, no internal imports
  - `internal/mena/source.go` — mena discovery, stdlib only
- `internal/output` — all CLI output types and formatting. Commands return data structs; the output package renders them.
- `internal/paths` — all XDG and project-root resolution. Never resolve paths inline in commands.

### Test File Convention

Test files colocated with source. Integration tests use `_integration_test.go` suffix. Fuzz tests use `fuzz_test.go` filename pattern.

---

## Domain-Specific Idioms

### 1. Typed String Constants for Domain States

All domain states use typed string constants with `IsValid()` and `String()` methods:
```go
type Status string
const (
    StatusActive   Status = "ACTIVE"
    StatusParked   Status = "PARKED"
    StatusArchived Status = "ARCHIVED"
)
func (s Status) IsValid() bool { ... }
func (s Status) IsTerminal() bool { ... }
```
Seen in: `session.Status`, `sails.Color`, `sails.ModifierType`, `sails.ProofStatus`, `artifact.ArtifactType`, `artifact.Phase`, `output.Format`. Never use raw `string` for domain states.

### 2. `Options` Struct + `WithOptions` Method Pattern

Large operations take an `Options` struct (not functional options):
```go
type Options struct {
    Force   bool
    DryRun  bool
    ...
}
func (m *Materializer) MaterializeWithOptions(riteName string, opts Options) (*Result, error)
```
Source: `internal/materialize/materialize.go`. Functional options (`Option func(*Server)`) exist in `internal/serve/server.go` but are the exception.

### 3. `Result` Struct for Operation Outcomes

Operations that return more than a simple value use a `Result` struct. This pairs with `Options`.

### 4. `Resolver` Pattern for Path and Context Access

Service types named `Resolver` encapsulate contextual lookup:
- `paths.Resolver` — project-root-relative path resolution
- Constructed once with `paths.NewResolver(projectRoot)`, methods resolve specific paths.

### 5. `Handled` Error Sentinel

The `errors.Handled(err)` wrapper signals that an error was already printed. Always use `common.PrintAndReturn(printer, err)` in cmd handlers instead of printing manually.

### 6. Dual YAML+JSON Struct Tags

All persisted types use both `yaml:` and `json:` tags:
```go
SessionID string `yaml:"session_id" json:"session_id"`
```

### 7. `Tabular` / `Textable` Interface for Output

Output types implement `Tabular` (`Headers()`, `Rows()`) or `Textable` (`Text()`). Never `fmt.Print` directly in commands. Source: `internal/output/output.go`.

### 8. `SchemaVersion` Field in Persisted Types

All persisted structs include `SchemaVersion string` for migration detection. Versions are semver strings like `"2.0"`, `"2.1"`, `"2.3"`.

### 9. Registry Pattern (Typed Keys)

`internal/registry` uses typed `RefKey string` for denial-recovery lookups. Prevents stringly-typed lookups.

### 10. LEAF Package Convention

Packages documented as `// This is a LEAF package` must not be extended with internal imports. Prevents circular dependencies.

### 11. Slog for Non-Fatal Conditions

`slog.Warn("message", "key", val)` for non-fatal events. The CLI is silent on success by design.

---

## Naming Patterns

### Package Names

Singular nouns, lowercase, matching the directory name: `session`, `agent`, `artifact`, `mena`, `manifest`, `procession`, `tribute`, `inscription`. Sub-packages: `userscope`, `orgscope`, `clewcontract`.

### Type Names

- `*Options` — configuration struct for an operation
- `*Result` — outcome struct
- `*Resolver` — contextual lookup service
- `*Validator` — validation service
- `*Materializer` — build/generation service
- Domain state types: named after the concept (`Status`, `Color`, `Phase`)

### Constructor Names

Standard Go `New*` constructors: `NewResolver(root)`, `NewPrinter(format, out, errOut, verbose)`. Named domain error constructors use `Err*` prefix.

### Constant Names

- Domain state values: typed + prefixed with type name (`StatusActive`, `ColorWhite`)
- Error codes: `SCREAMING_SNAKE_CASE` string constants
- Exit codes: `Exit{Name}` integer constants

### Acronym Conventions

All-caps in exported names: `SessionID`, `JSON`, `YAML`, `XDG`, `URL`, `CLI`.

### Receiver Names

Single-letter or short receivers: `(s *Session)`, `(m *Materializer)`, `(p *Printer)`, `(r *Resolver)`.

### File Names

Concern-per-file: `types.go`, `validate.go`, `manifest.go`, `errors.go`. Integration tests: `{concern}_integration_test.go`.

### Naming Anti-Patterns to Avoid

- Do not use `helper.go`, `util.go`, `common.go` inside domain packages — use concern-specific names instead.

---

## Knowledge Gaps

1. **`internal/hook` package** — 34 files, complex hook wiring. Hook-specific error handling and routing patterns not fully sampled.
2. **`internal/reason` package** — LLM reasoning pipeline. Error propagation in streaming/async paths not observed.
3. **`internal/search` package** — BM25 + embedding search. Store-level error handling patterns not sampled.
4. **`internal/procession` package** — Cross-rite coordinated workflows. Template generation idioms observed only at the type level.
5. **Functional options in `serve`** — `internal/serve/server.go` uses `type Option func(*Server)` but this package was not fully explored.
