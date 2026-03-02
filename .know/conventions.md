---
domain: conventions
generated_at: "2026-03-01T16:08:41Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "89b109c"
confidence: 0.88
format_version: "1.0"
---

# Codebase Conventions

## Error Handling Style

### Primary Pattern: Custom Error Type with Code + Exit Code

The codebase uses a single, centralized custom error system in `internal/errors/errors.go`. Standard library `fmt.Errorf` and `errors.New` are **not** the primary patterns — they appear in only ~65 places compared to ~475 uses of the custom package.

**The `*errors.Error` type:**
```go
// internal/errors/errors.go
type Error struct {
    Code     string                 `json:"code"`
    Message  string                 `json:"message"`
    Details  map[string]interface{} `json:"details,omitempty"`
    ExitCode int                    `json:"-"`
}
```

### Error Creation Constructors

Three creation levels, ordered by context richness:

```go
// 1. Message only
errors.New(errors.CodeFileNotFound, "session context file not found")

// 2. Message + structured details
errors.NewWithDetails(errors.CodeFileNotFound, "manifest file not found",
    map[string]interface{}{"path": path})

// 3. Wrap an existing error (stores cause in details["cause"])
errors.Wrap(errors.CodeGeneralError, "failed to read session context", err)
```

**`Wrap` does NOT use `%w` from `fmt.Errorf`.** It stores the cause as a string in the `details` map. This means `errors.Is` / `errors.As` on wrapped errors will not traverse the cause chain — the custom type is the terminal wrapper.

### Convenience Constructors

The `errors` package provides named constructors for every domain:

```go
errors.ErrProjectNotFound()
errors.ErrSessionNotFound(sessionID)
errors.ErrSessionExists(existingID, status)
errors.ErrLifecycleViolation(from, to, reason)
errors.ErrLockTimeout(lockPath, lockMeta)
errors.ErrRiteNotFound(riteName)
errors.ErrBudgetExceeded(current, requested, limit)
errors.ErrParseError(path, format, cause)
errors.ErrMergeConflict(conflictPaths, outputPath)
```

Named constructors follow the pattern `Err{DomainName}(...)`.

### Error Codes and Exit Codes

Error codes are `SCREAMING_SNAKE_CASE` string constants. Each code maps to an exit code (1-21) via `exitCodeForCode()`. This mapping is the authoritative CLI exit code table.

```go
// codes: CODE_NAME = "CODE_NAME"  (string equals constant name)
CodeGeneralError    = "GENERAL_ERROR"     // exit 1
CodeUsageError      = "USAGE_ERROR"       // exit 2
CodeFileNotFound    = "FILE_NOT_FOUND"    // exit 6
CodeValidationFailed = "VALIDATION_FAILED" // exit 12
```

Domain namespacing convention for codes:
- Core: `GENERAL_ERROR`, `USAGE_ERROR`, `FILE_NOT_FOUND`, etc.
- Rite domain: `RITE_NOT_FOUND`, `BORROW_CONFLICT`, `BUDGET_EXCEEDED`, `ORPHAN_CONFLICT`
- Manifest domain: `SCHEMA_NOT_FOUND`, `PARSE_ERROR`, `MERGE_CONFLICT`
- Sync domain: `SYNC_STATE_CORRUPT`, `REMOTE_REJECTED`, `NETWORK_ERROR`, `SYNC_NOT_CONFIGURED`

### Type Predicate Functions

Every error code has a paired `Is{Name}(err error) bool` function:

```go
errors.IsNotFound(err)
errors.IsLifecycleError(err)
errors.IsMergeConflict(err)
errors.IsRiteNotFound(err)
errors.IsSyncStateCorrupt(err)
```

All use the same pattern — type-assert to `*Error`, then compare `.Code`:
```go
func IsRiteNotFound(err error) bool {
    if e, ok := err.(*Error); ok {
        return e.Code == CodeRiteNotFound
    }
    return false
}
```

### Error Propagation at Command Boundaries

In `internal/cmd/` packages, the standard propagation pattern is:
1. Call business logic
2. If error: call `printer.PrintError(err)` to format to stderr
3. Return the error to propagate exit code upward

```go
// internal/cmd/session/create.go (repeated pattern)
if err := paths.EnsureDir(sessionDir); err != nil {
    err := errors.Wrap(errors.CodeGeneralError, "failed to create session directory", err)
    printer.PrintError(err)
    return err
}
```

Note: the local `err` is **re-assigned** (shadowed) when wrapping, so the wrapped version is both printed and returned.

### Exception: `fmt.Errorf` Usage

`fmt.Errorf` with `%w` appears in ~65 non-test locations, mostly in:
- Internal hook command helpers (`internal/cmd/hook/`) where errors are implementation-internal and not exposed at CLI boundaries
- `internal/sails/contract.go` for JSONL parsing errors
- A small number of `internal/cmd/worktree/`, `internal/cmd/sync/`, `internal/cmd/knows/` for transient errors not requiring exit codes

**Rule:** Use `errors.Wrap`/`errors.New` for anything that exits the `RunE` func. `fmt.Errorf` is acceptable for internal helper chains within a single file that ultimately wrap into the custom type before returning.

### Error Handling at the CLI Root

`cmd/ari/main.go` extracts exit code via `errors.GetExitCode(err)` which type-asserts to `*errors.Error`:
```go
if err := root.Execute(); err != nil {
    fmt.Fprintln(os.Stderr, "Error:", err)
    os.Exit(errors.GetExitCode(err))
}
```

`SilenceErrors: true` on the root cobra command prevents cobra from printing errors itself — the command is responsible for printing via `printer.PrintError`.

### JSON Error Format

The `*Error` type serializes to a wrapped JSON envelope:
```json
{
  "error": {
    "code": "FILE_NOT_FOUND",
    "message": "session context file not found",
    "details": { "path": "..." }
  }
}
```

---

## File Organization

### Two-Layer cmd Architecture

All CLI commands follow a strict two-layer split:
- `cmd/ari/main.go` — entry point only; wires embedded assets + version info + calls `root.Execute()`
- `internal/cmd/{domain}/` — all implementation lives here

**Nothing substantive lives in `cmd/`** beyond main's 31 lines.

### cmd Package Internal Structure

Each `internal/cmd/{domain}/` package follows a predictable layout:

| File | Contents |
|------|----------|
| `{domain}.go` | `New{Domain}Cmd(...)` constructor, `cmdContext` type, shared helpers |
| `{operation}.go` | One subcommand per file: `new{Op}Cmd()` + `run{Op}()` |
| `{operation}_test.go` | Tests for the corresponding subcommand |

Examples:
- `internal/cmd/session/session.go` — `NewSessionCmd`, `cmdContext`, helpers
- `internal/cmd/session/create.go` — `newCreateCmd`, `runCreate`, `runCreateSeeded`
- `internal/cmd/session/park.go` — `newParkCmd`, `runPark`

**Rule:** The group file (`{domain}.go`) owns the `cmdContext` struct and the `New{Domain}Cmd` constructor. Each operation file owns exactly one `cobra.Command`.

### internal/ Package Organization

Packages in `internal/` are organized by **domain concern**, not layer:

```
internal/
├── agent/          # Agent frontmatter parsing + validation
├── artifact/       # Federated artifact registry
├── checksum/       # SHA256 utilities (LEAF — no internal imports)
├── cmd/            # CLI command implementations (thin shell over domain packages)
├── config/         # XDG config helpers
├── errors/         # Custom error types and exit codes
├── fileutil/       # Atomic file write primitives
├── frontmatter/    # YAML frontmatter parser (shared primitive)
├── hook/           # Hook infrastructure + CC integration
│   └── clewcontract/  # Append-only JSONL event recording
├── inscription/    # CLAUDE.md region management
├── know/           # .know/ knowledge file parsing
├── lock/           # Advisory flock-based locking
├── manifest/       # Manifest load/validate/diff/merge
│   └── schemas/    # Embedded JSON schemas
├── materialize/    # .claude/ generation pipeline
│   ├── hooks/      # Hook config generation sub-package
│   ├── mena/       # Mena projection sub-package
│   ├── source/     # Rite source resolution sub-package
│   └── userscope/  # User-scope sync sub-package
├── mena/           # Mena walk + type detection (LEAF — no internal imports)
├── naxos/          # Session orphan scanning
├── output/         # Format-aware output printing
├── paths/          # Path resolution + XDG directories
├── provenance/     # Provenance manifest tracking
├── registry/       # Platform reference registry (LEAF — no internal imports)
├── rite/           # Rite context, invoker, workflow
├── sails/          # Sails confidence signaling
├── session/        # Session FSM, context, lifecycle
├── sync/           # Remote sync operations
├── tokenizer/      # Token counting
├── tribute/        # Session summary generation
├── validation/     # Cross-domain field validators
└── worktree/       # Git worktree management
```

### LEAF Package Convention

Two packages declare themselves as LEAF (no internal imports):
- `internal/mena/` — "LEAF package — it imports only stdlib (os, path/filepath, io/fs, strings)" — **verified true**
- `internal/registry/` — "LEAF package — it imports only stdlib. No internal/ imports." — **stale claim**: `registry/validate.go` imports `internal/frontmatter` and `internal/mena`

This indicates an intentional dependency-isolation pattern for frequently-imported primitive packages. When a package is a hotspot dependency, consider declaring it LEAF. Note: LEAF claims should be verified against actual imports.

### Types vs. Logic Separation

Several packages split into explicit files:
- `types.go` — data structures, constants, type definitions
- `{domain}.go` — methods and logic on those types

Examples:
- `internal/naxos/types.go` vs `internal/naxos/report.go` and `scanner.go`
- `internal/tribute/types.go` vs `tribute/renderer.go`
- `internal/materialize/sync_types.go` (types only) vs `materialize.go` (pipeline)

### Constants and Enums

Constants live in the file most relevant to their usage:
- Exit codes and error codes: `internal/errors/errors.go` (grouped by domain at the bottom)
- Status types: `internal/session/status.go` (dedicated status file)
- Phase types: `internal/session/fsm.go` (co-located with FSM logic that uses them)
- Format types: `internal/output/output.go` (co-located with printer that uses them)

There is no single "constants.go" convention. Constants are co-located with their consumer.

### Generated Code and Embedded Files

Embedded assets declared via `//go:embed` in `internal/manifest/schema.go`:
```go
//go:embed schemas/*.json
var schemaFS embed.FS
```

The root `knossos` package embeds rites, templates, hooks YAML, agents, and mena as `embed.FS` fields accessed via `knossos.EmbeddedRites`, `knossos.EmbeddedTemplates`, etc.

### Output Structures

All output types live in `internal/output/`. Three files organize them by domain:
- `output.go` — core `Printer` type + session-domain output structs
- `manifest.go` — manifest-domain output structs
- `rite.go` — rite-domain output structs

Every output struct implements either `Textable` (`Text() string`) or `Tabular` (`Headers() []string` + `Rows() [][]string`). JSON output uses struct JSON tags.

---

## Domain-Specific Idioms

### Polymorphic YAML Fields: `FlexibleStringSlice`

`internal/frontmatter/FlexibleStringSlice` accepts both comma-separated strings and YAML list syntax for the same field. Used pervasively for agent `tools` and `disallowedTools`:

```go
// Accepts: "Bash, Read, Glob" OR [Bash, Read, Glob]
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalYAML(value *yaml.Node) error {
    if value.Kind == yaml.SequenceNode { /* parse list */ }
    // fall back to comma-split string
}
```

The type is aliased into `internal/agent/types.go` as `FlexibleStringSlice = frontmatter.FlexibleStringSlice` and again into `internal/materialize/frontmatter.go` as `FlexibleStringSlice = mena.FlexibleStringSlice`.

### Polymorphic YAML Fields: `MemoryField`

`internal/agent/types.go` `MemoryField` accepts `bool` (true -> "project") or string enum ("user", "project", "local"):

```go
type MemoryField string

func (m *MemoryField) UnmarshalYAML(value *yaml.Node) error {
    if value.Tag == "!!bool" { /* true -> "project", false -> "" */ }
    // fall back to string
}
```

### Dual YAML/JSON Mirror Structs

When a struct needs custom timestamp serialization, the codebase defines a parallel `{Type}YAML` internal struct:

```go
// internal/session/context.go
type Context struct {
    CreatedAt time.Time `yaml:"created_at" json:"created_at"`
}

type contextYAML struct {
    CreatedAt string `yaml:"created_at"` // RFC3339 string for YAML round-trip
}
```

The exported struct uses `time.Time`; the internal YAML struct uses `string`. `ParseContext` unmarshals into `contextYAML`, then converts. `Serialize` converts `Context` -> `contextYAML` -> yaml.Marshal.

### Resolver Pattern

Path operations use `*paths.Resolver`, a struct that encodes a project root and provides named path methods:

```go
resolver := paths.NewResolver(projectDir)
resolver.SessionDir(sessionID)        // .sos/sessions/{id}
resolver.SessionContextFile(sessionID) // .sos/sessions/{id}/SESSION_CONTEXT.md
resolver.AgentsDir()                  // .claude/agents/
resolver.LocksDir()                   // .sos/sessions/.locks/
```

This prevents hardcoded path construction scattered across packages. All path construction must go through `Resolver` or a `paths.*` package function.

### Atomic File Writes via `fileutil.AtomicWriteFile`

All file mutations that must be durable use `fileutil.AtomicWriteFile(path, content, perm)`:
```
write -> temp file -> fsync -> close -> chmod -> rename
```

Calling `os.WriteFile` directly is a pattern deviation — use `fileutil.AtomicWriteFile` for session context, manifests, and any critical state file.

A `WriteIfChanged` variant skips the write if content matches:
```go
changed, err := fileutil.WriteIfChanged(path, content, 0644)
```

### Registry Reference Pattern

The `internal/registry` package provides stable string keys for platform references (agents, CLI commands, skills). Use `registry.Ref(key)` not hardcoded strings when generating denial messages or delegation guidance:

```go
registry.Ref(registry.AgentMoirai)           // "moirai"
registry.Ref(registry.CLISessionFieldSet)    // "ari session field-set"
registry.TaskDelegation(registry.AgentMoirai, "transition_phase", "update_field")
```

### cmdContext Embedding Pattern

Every `internal/cmd/{domain}/` package defines a `cmdContext` that embeds `common.BaseContext` or `common.SessionContext`:

```go
type cmdContext struct {
    common.SessionContext  // provides GetPrinter, GetResolver, GetLockManager, GetSessionID
}
```

`BaseContext` holds `*string` pointers to global flags (`Output`, `Verbose`, `ProjectDir`). This pointer-to-pointer pattern is intentional — cobra flag binding requires pointers that persist.

### Printer as the Single Output Gate

Commands never write to `os.Stdout` or `os.Stderr` directly. All output goes through `output.Printer`:

```go
printer := ctx.getPrinter()
printer.Print(result)           // JSON/YAML/text dispatch
printer.PrintError(err)         // always to stderr
printer.VerboseLog("info", msg, fields)  // JSON lines to stderr when --verbose
```

Hook commands always use `FormatJSON` regardless of `-o` flag.

### Frontmatter-in-Markdown Pattern

The codebase parses multiple file types with YAML frontmatter delimited by `---\n...\n---\n`. The shared parser is `internal/frontmatter.Parse(content []byte) (yamlBytes, body []byte, err)`.

Consumer pattern:
```go
yamlBytes, body, err := frontmatter.Parse(content)
if err != nil { /* ErrMissingOpenDelimiter or ErrMissingCloseDelimiter */ }
var fm MyFrontmatter
yaml.Unmarshal(yamlBytes, &fm)
```

This pattern appears in: agent parsing, session context parsing, wip file validation, mena frontmatter parsing.

### Source Chain / Priority Ordering

The materialize pipeline uses a priority-ordered `[]MenaSource` list where **higher index = higher priority**:
```
index 0: platform mena (lowest)
index 1: shared rite mena
index N: dependency rite mena
index N+1: rite-local mena (highest)
```

Built via `mena.BuildSourceChain(opts)`. Consumers iterate the slice respecting index order for override resolution.

### Typed String Enums with `.IsValid()` / `.String()`

Domain-specific string types follow a consistent interface:

```go
type Status string

const (
    StatusActive   Status = "ACTIVE"
    StatusArchived Status = "ARCHIVED"
)

func (s Status) String() string   { return string(s) }
func (s Status) IsValid() bool    { switch s { case ...: return true; default: return false } }
func (s Status) IsTerminal() bool { return s == StatusArchived }
```

This pattern appears in: `session.Status`, `session.Phase`, `inscription.OwnerType`, `inscription.MarkerDirective`, `materialize.SyncScope`, `naxos.OrphanReason`, `naxos.SuggestedAction`.

### Event System: Append-Only JSONL

`internal/hook/clewcontract` provides a `BufferedEventWriter` that appends structured events to `events.jsonl` in session directories. Never overwrite event files — the log is append-only:

```go
w := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
defer w.Close()
w.Write(clewcontract.NewSessionCreatedEvent(sessionID, initiative, complexity, rite))
w.Flush()
```

Events are non-fatal — failures are logged at verbose level and ignored.

---

## Naming Patterns

### Package Naming

- Domain packages: single-word lowercase matching the directory (`session`, `manifest`, `sails`, `inscription`, `materialize`, `provenance`)
- Sub-packages: descriptive single-word (`clewcontract`, `userscope`, `source`)
- cmd packages: match the CLI command name (`hook`, `session`, `rite`, `inscription`)
- Exception: `internal/cmd/initialize/` uses package name `initialize` (not `init`, which is reserved)
- Exception: `initcmd "github.com/.../initialize"` import alias avoids conflict with `init()`

### Type Naming

**Struct types**: PascalCase matching their role, not the file name.
- `paths.Resolver` (not `PathResolver`)
- `session.Context` (not `SessionContext` — already namespaced)
- `output.Printer` (not `OutputPrinter`)
- `lock.Manager` (not `LockManager`)

**String enum types**: PascalCase noun.
- `session.Status`, `session.Phase`
- `inscription.OwnerType`, `inscription.MarkerDirective`
- `materialize.SyncScope`, `materialize.SyncResource`

**Const values for string enums**: PrefixedPascalCase.
- `StatusActive`, `StatusParked` (not `ACTIVE`, `PARKED`)
- `PhaseRequirements`, `PhaseDesign`
- `OwnerKnossos`, `OwnerSatellite`
- `ScopeAll`, `ScopeRite`, `ScopeUser`
- Exception: error codes use `SCREAMING_SNAKE_CASE` strings to match CLI output

### Function Naming

**Constructors**: `New{Type}(...)` for non-error-returning, functional pattern for error-returning:
- `paths.NewResolver(projectDir)` — no error
- `artifact.NewRegistry(projectRoot)` — no error
- `manifest.Load(path)` — returns `(*Manifest, error)` (verb, not `New`)

**Error constructors**: `Err{DomainConcept}(args)` returning `*Error`:
- `errors.ErrProjectNotFound()`
- `errors.ErrSessionExists(existingID, status)`
- `errors.ErrRiteNotFound(riteName)`

**Predicate functions**: `Is{Condition}(err)` for error type checks, `Can{Action}()` for capability checks:
- `errors.IsNotFound(err)`, `errors.IsMergeConflict(err)`
- `session.FSM.CanTransition(from, to)`

**cmd constructors**: `New{Domain}Cmd(flags...)` for group commands, `new{Op}Cmd(ctx)` (lowercase) for leaf subcommands:
- `session.NewSessionCmd(...)` — exported, registered in root
- `newCreateCmd(ctx)` — unexported, registered in group

**cmd run functions**: `run{Op}(ctx, ...)` — always unexported, wraps `run{Op}Core` when testing injection is needed.

### Variable Naming

**No Hungarian notation**. Use short, clear names:
- `ctx` for `*cmdContext` or similar context holders
- `printer` for `*output.Printer`
- `resolver` for `*paths.Resolver`
- `opts` for options structs
- `err` always for errors (never `e`, `er`, `ferr`, `apiErr`)

**Avoid stutter** (package name repeated in identifier):
- `manifest.Manifest` (acceptable — Go convention for primary type)
- `session.Context` (not `session.SessionContext`)
- `lock.Manager` (not `lock.LockManager`)

### File Naming

- One concern per file: `create.go`, `park.go`, `resume.go` not `session_ops.go`
- Test files: `{name}_test.go` always
- Integration tests: `integration_test.go` or `{name}_integration_test.go`
- Types-only files: `types.go`
- Constants/shared errors: `errors.go` within a package (e.g., `internal/frontmatter/errors.go`)

### Acronym Conventions

- `CC` = Claude Code (used in comments, not exported names)
- `CLI` = all caps in constants (`CLISessionFieldSet`), `cli` in package path (`internal/cmd/`)
- `MCP` = all caps (`MCPServer`, `McpServerConfig` — inconsistency exists here)
- `JSON` = all caps in comments, `JSON()` method name, `FormatJSON` constant
- `YAML` = all caps in comments, `FormatYAML` constant
- `FSM` = all caps in type name (`session.FSM`)
- `ID` = all caps when standalone (`SessionID`, `ArtifactID`), lowercase `id` in variable names

### Notable Naming Inconsistency

`McpServerConfig` (PascalCase `Mcp`) appears in `internal/agent/types.go` while `MCPServer` (all-caps) appears in `internal/materialize/materialize.go`. The `Mcp` casing in frontmatter structs follows CC's schema expectations (camelCase YAML keys like `mcpServers`).

---

## Test Assertion Style

### Preferred Style: stdlib `testing` package

The codebase does **not** enforce a single assertion library. The dominant pattern (approximately 133 of 151 test files) uses stdlib `testing` only:

```go
if err != nil {
    t.Fatalf("Failed to do X: %v", err)
}
if got != want {
    t.Errorf("subject = %v, want %v", got, want)
}
```

**Rule:** New tests should default to stdlib assertions. This avoids adding a testify dependency to packages that do not already use it.

### When Testify is Acceptable

Testify (`github.com/stretchr/testify`) is used in 23 of 151 test files, concentrated in `internal/materialize/` and `internal/sails/`. Adding testify assertions to a file that already imports it is fine:

```go
require.NoError(t, err)
assert.Equal(t, expected, actual)
```

**Rule:** Do NOT add `require`/`assert` imports to packages that currently use stdlib-only assertions. Consistent per-file, not per-project.

### Do Not Migrate

Existing stdlib tests must NOT be migrated to testify (or vice versa). Tests are correct as-is; style migration adds churn with zero functional benefit.

### The `wantErr` Table Pattern

For systematic error-path coverage, use the `wantErr bool` table-test pattern:

```go
type testCase struct {
    name    string
    input   string
    wantErr bool
}
// ...
if (err != nil) != tt.wantErr {
    t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
}
```

This pattern appears in `manifest`, `frontmatter`, `config`, `mena`, `rite`, `sails`, `hook`, and `agent` packages.

---

## Knowledge Gaps

1. **`internal/materialize/materialize.go` full pipeline** — only the first 80 lines were read. The complete rite materialization pipeline was not fully observed.
2. **`internal/inscription/` pipeline files** — only `types.go` was read. The generator, merger, pipeline, and sync files were not examined.
3. **`internal/session/` discovery and lifecycle files** — `context.go`, `fsm.go`, `status.go` were read, but `discovery.go`, `snapshot.go`, `timeline.go`, `rotation.go`, `resolve.go` were not.
4. **`internal/tribute/` and `internal/worktree/`** — types observed but implementation logic not read.
5. **Root `knossos` package embed declarations** — referenced in `main.go` but source not read.
