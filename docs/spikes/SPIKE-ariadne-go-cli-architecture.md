# SPIKE: Ariadne Go CLI Architecture

> Research spike for Go binary replacement of bash script harness for Claude Code agentic workflows

**Time-boxed**: Research spike (findings, not production code)
**Date**: 2026-01-04
**Author**: Technology Scout (rnd-pack)

---

## Executive Summary

This spike researches best practices and patterns for building **ariadne**, a Go CLI binary to replace the current bash script harness. After analyzing reputable OSS projects (gh CLI, terraform, kubectl, chezmoi), we recommend **Cobra** as the CLI framework, **santhosh-tekuri/jsonschema** for schema validation with **embed.FS**, the **k8s strategic merge patch** library for three-way JSON merge, and **adrg/xdg** for state management. The strangler fig pattern provides a low-risk migration path.

**Verdict**: Proceed with implementation. All major technical questions have satisfactory answers from production-proven patterns.

---

## 1. Go CLI Framework Recommendation

### Comparison Matrix

| Criteria | Cobra | Kong | urfave/cli |
|----------|-------|------|------------|
| **Adoption** | 42K+ stars, 173K+ projects | Growing | 20K+ stars |
| **Notable Users** | gh, kubectl, terraform, docker, hugo | - | - |
| **Subcommand Support** | Excellent (unlimited nesting) | Good | Good |
| **JSON Output Pattern** | Well-documented | Manual | Manual |
| **Shell Completion** | Built-in (bash, zsh, fish, PowerShell) | Plugin | Manual |
| **Testing** | Excellent separation | Good | Fair |
| **Learning Curve** | Moderate | Low | Low |
| **Viper Integration** | Native | Manual | Manual |

### Recommendation: **Cobra**

**Rationale**:
1. **Industry standard**: Used by all reference projects (gh, terraform, kubectl)
2. **Ecosystem compatibility**: Already specified in roster tech stack (`user-skills/documentation/standards/tech-stack-infrastructure.md`)
3. **Battle-tested patterns**: Rich corpus of real-world examples to reference
4. **Built-in features**: Help generation, shell completion, flag inheritance reduce boilerplate

**Risk**: Cobra has been criticized for "clunky" API and edge cases. Mitigation: Follow gh CLI's factory pattern for testability.

### Sources
- [Cobra GitHub](https://github.com/spf13/cobra)
- [Cobra Documentation](https://cobra.dev/)
- [Go CLI Comparison](https://github.com/gschauer/go-cli-comparison)
- [Kong Analysis](https://danielms.site/zet/2023/kong-is-an-amazing-cli-for-go-apps/)

---

## 2. Architecture Patterns from Reference Projects

### 2.1 gh CLI (GitHub CLI)

**Structure Pattern**:
```
cmd/
  gh/
    main.go           # Entry point
pkg/
  cmd/
    factory/          # Dependency injection
    root/             # Root command
    pr/               # Domain: pull requests
      create/
      list/
      view/
    issue/            # Domain: issues
    repo/             # Domain: repositories
  cmdutil/            # Shared utilities
  iostreams/          # I/O abstraction for testing
  jsoncolor/          # JSON output formatting
```

**Key Patterns**:
- **Factory Pattern**: Centralized command construction with dependency injection
- **Domain Directories**: Each top-level command (pr, issue, repo) is a directory
- **I/O Abstraction**: `iostreams` package enables testing without actual terminal
- **Piped Output Detection**: Automatic machine-readable format when piped

**JSON Output Handling**:
```go
// Pattern: --json flag with field selection
cmd.Flags().StringSliceVar(&opts.JSON, "json", nil, "Output JSON with specified fields")
cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Format using Go template")
cmd.Flags().StringVarP(&opts.JQ, "jq", "q", "", "Filter JSON with jq expression")
```

### 2.2 Terraform

**Structure Pattern**:
```
main.go              # Entry point
commands.go          # Command registry mapping
internal/
  command/           # All commands here
    apply.go
    plan.go
    init.go
  backend/           # State backend abstraction
  configs/           # Configuration loading
    configload/      # Module loading
```

**Key Patterns**:
- **Command Registry**: Single `commands.go` maps user-facing names to implementations
- **Backend Abstraction**: State management separated from command logic
- **Graph Engine**: Operations build a dependency graph for execution
- **terraform-exec**: Separate Go library for programmatic access

### 2.3 kubectl

**Structure Pattern**:
```
cmd/kubectl/
  kubectl.go         # Minimal entry point
pkg/kubectl/         # Main logic (enables unit testing)
staging/
  src/k8s.io/
    cli-runtime/     # Reusable CLI helpers
    apimachinery/    # Strategic merge patch here
```

**Key Patterns**:
- **Minimal main.go**: Entry point contains no logic
- **cli-runtime**: Generic CLI options (kubeconfig, namespace, output format)
- **RESTMapper**: Maps resource names to API endpoints dynamically

### 2.4 chezmoi

**Key Patterns**:
- **XDG Compliance**: Fully respects XDG Base Directory Specification
- **Multi-format Config**: Supports JSON, JSONC, TOML, YAML for config files
- **Cross-platform**: Single codebase handles Unix, Windows, macOS

### Sources
- [gh CLI Repository](https://github.com/cli/cli)
- [Terraform Architecture](https://github.com/hashicorp/terraform/blob/main/docs/architecture.md)
- [kubectl Source](https://github.com/kubernetes/kubernetes/blob/master/cmd/kubectl/kubectl.go)
- [chezmoi Configuration](https://www.chezmoi.io/reference/configuration-file/)

---

## 3. Schema Embedding and Validation

### Embedding with embed.FS

```go
package schemas

import "embed"

//go:embed *.json
var SchemaFS embed.FS

// Access schema at runtime
func GetSchema(name string) ([]byte, error) {
    return SchemaFS.ReadFile(name + ".json")
}
```

**Key Points**:
- Introduced in Go 1.16
- Compile-time inclusion (files must exist at build)
- Read-only, goroutine-safe
- Implements `fs.FS` interface (works with `net/http`, templates)

### Validation Library Comparison

| Criteria | santhosh-tekuri/jsonschema | xeipuuv/gojsonschema | CUE |
|----------|---------------------------|---------------------|-----|
| **Draft Support** | 2020-12, 2019-09, 7, 6, 4 | 4, 6, 7 | Custom |
| **Correctness** | Excellent | Excellent | Excellent |
| **Performance** | ~2x faster | Baseline | Slower |
| **Active Maintenance** | Yes (v5, v6 releases) | Less active | Yes |
| **Custom Formats** | Yes | Yes | Yes |
| **Error Detail** | Rich context | Good | Excellent |

### Recommendation: **santhosh-tekuri/jsonschema**

**Rationale**:
1. Best-in-class performance for complex schemas
2. Supports latest JSON Schema drafts (2020-12)
3. Active maintenance
4. Rich validation error context

**Pattern for ariadne**:
```go
package validation

import (
    "embed"
    "github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schemas/*.json
var schemaFS embed.FS

var (
    sessionSchema  *jsonschema.Schema
    manifestSchema *jsonschema.Schema
)

func init() {
    compiler := jsonschema.NewCompiler()
    compiler.UseLoader(jsonschema.SchemeURLLoader{
        "embed": &embedLoader{fs: schemaFS},
    })

    sessionSchema = compiler.MustCompile("embed:///schemas/session.json")
    manifestSchema = compiler.MustCompile("embed:///schemas/manifest.json")
}

func ValidateSession(data []byte) error {
    return sessionSchema.Validate(bytes.NewReader(data))
}
```

### CUE Consideration

CUE offers powerful configuration validation but adds complexity:
- **Pro**: Can generate Go types from CUE schemas
- **Pro**: Unification-based validation
- **Con**: Learning curve for team
- **Con**: Requires Go 1.24+

**Verdict**: Start with JSON Schema. Consider CUE for v2 if configuration complexity warrants it.

### Sources
- [Go embed Package](https://pkg.go.dev/embed)
- [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)
- [JSON Schema Validator Benchmarks](https://dev.to/vearutop/benchmarking-correctness-and-performance-of-go-json-schema-validators-3247)
- [CUE Go Integration](https://cuelang.org/docs/concept/how-cue-works-with-go/)

---

## 4. Three-Way Merge Implementation

### JSON Three-Way Merge: Kubernetes Strategic Merge Patch

The `k8s.io/apimachinery/pkg/util/strategicpatch` package provides production-grade three-way merge:

```go
import (
    "k8s.io/apimachinery/pkg/util/strategicpatch"
    "k8s.io/apimachinery/pkg/util/mergepatch"
)

// Three documents required:
// - original: baseline (last-applied configuration)
// - modified: desired state (new configuration)
// - current: live state (what's actually there)

func ThreeWayMergeJSON(original, modified, current []byte) ([]byte, error) {
    // For non-Kubernetes use, use JSON merge patch variant
    patch, err := strategicpatch.CreateThreeWayMergePatch(
        original,
        modified,
        current,
        lookupPatchMeta,  // Schema for merge strategy
        false,            // Don't overwrite conflicts
    )
    if err != nil {
        return nil, err
    }

    return strategicpatch.StrategicMergePatchUsingLookupPatchMeta(
        current,
        patch,
        lookupPatchMeta,
    )
}
```

**For simpler JSON without strategic merge keys**, use JSON Merge Patch (RFC 7396):

```go
import "github.com/evanphx/json-patch/v5"

func SimpleThreeWayMerge(original, modified, current []byte) ([]byte, error) {
    // Create patch from original -> modified
    patch, err := jsonpatch.CreateMergePatch(original, modified)
    if err != nil {
        return nil, err
    }

    // Apply patch to current
    return jsonpatch.MergePatch(current, patch)
}
```

### Markdown Section-Level Merge

No off-the-shelf library exists for markdown section merge. **Custom implementation required**.

**Proposed Algorithm**:
```go
type Section struct {
    Header   string   // e.g., "## Installation"
    Level    int      // Header level (1-6)
    Content  string   // Content until next header
    Children []Section
}

func ParseMarkdownSections(md string) []Section {
    // Parse markdown into tree by headers
    // Treat each header as section boundary
}

func ThreeWayMergeMarkdown(original, modified, current string) (string, error) {
    origSections := ParseMarkdownSections(original)
    modSections := ParseMarkdownSections(modified)
    curSections := ParseMarkdownSections(current)

    // Match sections by header text
    // For each section:
    //   - If only modified changed it: use modified
    //   - If only current changed it: use current
    //   - If both changed same section: CONFLICT (flag for manual resolution)
    //   - If neither changed: use original
}
```

**Implementation Notes**:
- Use `github.com/yuin/goldmark` for markdown parsing
- Section matching by header text (not position)
- Flag conflicts rather than auto-merging conflicting sections
- Consider YAML frontmatter as special "header-less" section

### Sources
- [Kubernetes Strategic Merge Patch](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/strategicpatch)
- [Strategic Merge Patch Documentation](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md)
- [json-patch library](https://github.com/evanphx/json-patch)
- [Three-Way Merge Theory](https://blog.jcoglan.com/2017/05/08/merging-with-diff3/)

---

## 5. XDG State Management

### Library: adrg/xdg

**Recommendation**: Use `github.com/adrg/xdg` - the most comprehensive XDG implementation for Go.

```go
import "github.com/adrg/xdg"

// Directory purposes for ariadne:
var (
    // Config: read-only user preferences (e.g., default team, aliases)
    ConfigDir = filepath.Join(xdg.ConfigHome, "ariadne")

    // Data: persistent data that should survive cache clear (e.g., team packs)
    DataDir = filepath.Join(xdg.DataHome, "ariadne")

    // State: mutable state that's not config (e.g., sessions, locks)
    StateDir = filepath.Join(xdg.StateHome, "ariadne")

    // Cache: disposable cached data (e.g., schema cache, parsed manifests)
    CacheDir = filepath.Join(xdg.CacheHome, "ariadne")
)
```

### Directory Layout for ariadne

```
$XDG_CONFIG_HOME/ariadne/      # ~/.config/ariadne/
  config.yaml                   # User preferences
  aliases.yaml                  # Command aliases

$XDG_DATA_HOME/ariadne/        # ~/.local/share/ariadne/
  teams/                        # Downloaded team packs
    rnd-pack/
    security-pack/
  manifests/                    # Team manifests

$XDG_STATE_HOME/ariadne/       # ~/.local/state/ariadne/
  sessions/                     # Active session state
    current -> abc123/
    abc123/
      SESSION_CONTEXT.md
  locks/                        # Lock files

$XDG_CACHE_HOME/ariadne/       # ~/.cache/ariadne/
  schemas/                      # Cached compiled schemas
  parsed/                       # Parsed markdown cache
```

### Helper Functions

```go
package paths

import (
    "os"
    "path/filepath"
    "github.com/adrg/xdg"
)

// EnsureConfigDir creates config directory if needed
func EnsureConfigDir() (string, error) {
    dir := filepath.Join(xdg.ConfigHome, "ariadne")
    return dir, os.MkdirAll(dir, 0755)
}

// ConfigFile returns path to config file, creating dir if needed
func ConfigFile(name string) (string, error) {
    dir, err := EnsureConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, name), nil
}

// StateFile returns path to state file, creating dir if needed
func StateFile(relPath string) (string, error) {
    dir := filepath.Join(xdg.StateHome, "ariadne")
    if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, relPath)), 0755); err != nil {
        return "", err
    }
    return filepath.Join(dir, relPath), nil
}

// CacheFile returns path to cache file (may not exist)
func CacheFile(relPath string) string {
    return filepath.Join(xdg.CacheHome, "ariadne", relPath)
}
```

### Cross-Platform Considerations

| Platform | ConfigHome | DataHome | StateHome | CacheHome |
|----------|------------|----------|-----------|-----------|
| Linux | ~/.config | ~/.local/share | ~/.local/state | ~/.cache |
| macOS | ~/Library/Application Support | ~/Library/Application Support | ~/Library/Application Support | ~/Library/Caches |
| Windows | %LOCALAPPDATA% | %LOCALAPPDATA% | %LOCALAPPDATA% | %LOCALAPPDATA%\cache |

The `adrg/xdg` library handles these differences automatically.

### Sources
- [adrg/xdg GitHub](https://github.com/adrg/xdg)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
- [chezmoi Configuration](https://www.chezmoi.io/reference/configuration-file/)

---

## 6. Bash-to-Go Migration Strategy

### Strangler Fig Pattern

The strangler fig pattern enables gradual migration without a risky big-bang rewrite:

```
Phase 1: Wrapper
┌─────────────────────────────────────────┐
│  ariadne (Go)                           │
│    ├── validate   → Go implementation   │
│    ├── sync       → calls bash script   │
│    └── session    → calls bash script   │
└─────────────────────────────────────────┘

Phase 2: Hybrid
┌─────────────────────────────────────────┐
│  ariadne (Go)                           │
│    ├── validate   → Go implementation   │
│    ├── sync       → Go implementation   │
│    └── session    → calls bash script   │
└─────────────────────────────────────────┘

Phase 3: Complete
┌─────────────────────────────────────────┐
│  ariadne (Go)                           │
│    ├── validate   → Go implementation   │
│    ├── sync       → Go implementation   │
│    └── session    → Go implementation   │
└─────────────────────────────────────────┘
```

### Implementation Pattern

```go
package legacy

import (
    "os"
    "os/exec"
)

// CallLegacyScript executes a bash script during migration
func CallLegacyScript(script string, args ...string) error {
    cmd := exec.Command("bash", append([]string{script}, args...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// In command implementation:
func (c *SessionCommand) Run() error {
    if c.useLegacy {
        return legacy.CallLegacyScript("session-manager.sh", c.args...)
    }
    return c.runNative()
}
```

### Testing Strategy for Behavioral Parity

```go
// Golden file testing pattern
func TestSessionList_Parity(t *testing.T) {
    // Run legacy bash implementation
    legacyOut, _ := exec.Command("bash", "session-manager.sh", "list").Output()

    // Run new Go implementation
    var buf bytes.Buffer
    cmd := NewSessionListCmd()
    cmd.SetOut(&buf)
    cmd.Execute()

    // Compare outputs (normalize whitespace, timestamps)
    if !equalNormalized(legacyOut, buf.Bytes()) {
        t.Errorf("Output mismatch:\nLegacy: %s\nNew: %s", legacyOut, buf.Bytes())
    }
}
```

### Migration Order Recommendation

| Order | Domain | Rationale |
|-------|--------|-----------|
| 1 | **validate** | Stateless, low risk, immediate value |
| 2 | **manifest** | Read-heavy, builds on validation |
| 3 | **sync** | Can leverage validation/manifest |
| 4 | **team** | Depends on manifest |
| 5 | **session** | Most stateful, highest risk |

### Sources
- [Strangler Fig Pattern](https://dev.to/aman_kumar_bdd40f1b711c15/from-monolithic-clis-to-modular-plugins-applying-the-strangler-fig-pattern-3gok)
- [AWS Strangler Fig Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/strangler-fig.html)

---

## 7. Code Sketches

### 7.1 Project Structure

```
ariadne/
├── cmd/
│   └── ariadne/
│       └── main.go              # Entry point only
├── internal/
│   ├── cmd/                     # Command implementations
│   │   ├── root.go
│   │   ├── sync/
│   │   │   ├── sync.go
│   │   │   └── sync_test.go
│   │   ├── session/
│   │   ├── team/
│   │   └── manifest/
│   ├── config/                  # Configuration loading
│   ├── validation/              # Schema validation
│   ├── merge/                   # Three-way merge
│   ├── paths/                   # XDG path helpers
│   └── legacy/                  # Bash script bridge
├── schemas/                     # JSON schemas (embedded)
│   ├── session.json
│   ├── manifest.json
│   └── team.json
├── go.mod
├── go.sum
└── Makefile
```

### 7.2 Root Command with JSON Output

```go
// internal/cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

type GlobalOptions struct {
    Output   string // "text", "json", "yaml"
    Verbose  bool
    Config   string
}

var globalOpts GlobalOptions

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ariadne",
        Short: "Claude Code workflow harness",
        Long:  `Ariadne manages sessions, teams, and manifests for Claude Code agentic workflows.`,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            return initConfig()
        },
    }

    // Global flags
    cmd.PersistentFlags().StringVarP(&globalOpts.Output, "output", "o", "text",
        "Output format: text, json, yaml")
    cmd.PersistentFlags().BoolVarP(&globalOpts.Verbose, "verbose", "v", false,
        "Enable verbose output")
    cmd.PersistentFlags().StringVar(&globalOpts.Config, "config", "",
        "Config file (default: $XDG_CONFIG_HOME/ariadne/config.yaml)")

    // Register subcommands
    cmd.AddCommand(
        newSyncCmd(),
        newSessionCmd(),
        newTeamCmd(),
        newManifestCmd(),
        newValidateCmd(),
    )

    return cmd
}

func initConfig() error {
    if globalOpts.Config != "" {
        viper.SetConfigFile(globalOpts.Config)
    } else {
        viper.AddConfigPath(filepath.Join(xdg.ConfigHome, "ariadne"))
        viper.SetConfigName("config")
    }
    viper.AutomaticEnv()
    return viper.ReadInConfig()
}
```

### 7.3 Subcommand with Nested Commands

```go
// internal/cmd/session/session.go
package session

import "github.com/spf13/cobra"

func NewSessionCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "session",
        Short: "Manage workflow sessions",
        Long:  `Create, list, and manage Claude Code workflow sessions.`,
    }

    cmd.AddCommand(
        newListCmd(),
        newCreateCmd(),
        newParkCmd(),
        newResumeCmd(),
    )

    return cmd
}

// internal/cmd/session/list.go
func newListCmd() *cobra.Command {
    var opts listOptions

    cmd := &cobra.Command{
        Use:   "list",
        Short: "List sessions",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runList(cmd.Context(), opts)
        },
    }

    cmd.Flags().BoolVar(&opts.All, "all", false, "Include completed sessions")
    cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status")

    return cmd
}

type listOptions struct {
    All    bool
    Status string
}

func runList(ctx context.Context, opts listOptions) error {
    sessions, err := loadSessions(opts)
    if err != nil {
        return err
    }

    // Output based on global format flag
    switch globalOpts.Output {
    case "json":
        return outputJSON(sessions)
    case "yaml":
        return outputYAML(sessions)
    default:
        return outputTable(sessions)
    }
}
```

### 7.4 Schema Validation Module

```go
// internal/validation/validator.go
package validation

import (
    "bytes"
    "embed"
    "fmt"
    "io/fs"
    "sync"

    "github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schemas/*.json
var schemaFS embed.FS

var (
    compiler *jsonschema.Compiler
    schemas  = make(map[string]*jsonschema.Schema)
    mu       sync.RWMutex
)

func init() {
    compiler = jsonschema.NewCompiler()

    // Register embedded filesystem as URL scheme
    compiler.UseLoader(&embedLoader{})
}

type embedLoader struct{}

func (l *embedLoader) Load(uri string) (any, error) {
    // uri format: "embed:///schemas/session.json"
    path := strings.TrimPrefix(uri, "embed:///")
    data, err := schemaFS.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var v any
    return v, json.Unmarshal(data, &v)
}

// GetSchema returns a compiled schema, caching the result
func GetSchema(name string) (*jsonschema.Schema, error) {
    mu.RLock()
    if s, ok := schemas[name]; ok {
        mu.RUnlock()
        return s, nil
    }
    mu.RUnlock()

    mu.Lock()
    defer mu.Unlock()

    // Double-check after acquiring write lock
    if s, ok := schemas[name]; ok {
        return s, nil
    }

    uri := fmt.Sprintf("embed:///schemas/%s.json", name)
    s, err := compiler.Compile(uri)
    if err != nil {
        return nil, fmt.Errorf("compiling schema %s: %w", name, err)
    }

    schemas[name] = s
    return s, nil
}

// ValidateSession validates session JSON against schema
func ValidateSession(data []byte) error {
    schema, err := GetSchema("session")
    if err != nil {
        return err
    }

    var v any
    if err := json.Unmarshal(data, &v); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    return schema.Validate(v)
}
```

### 7.5 Output Formatting Pattern

```go
// internal/output/output.go
package output

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "text/tabwriter"

    "gopkg.in/yaml.v3"
)

type Format string

const (
    FormatText Format = "text"
    FormatJSON Format = "json"
    FormatYAML Format = "yaml"
)

// Printer handles formatted output
type Printer struct {
    format Format
    out    io.Writer
}

func NewPrinter(format Format, out io.Writer) *Printer {
    if out == nil {
        out = os.Stdout
    }
    return &Printer{format: format, out: out}
}

// Print outputs data in the configured format
func (p *Printer) Print(data any) error {
    switch p.format {
    case FormatJSON:
        enc := json.NewEncoder(p.out)
        enc.SetIndent("", "  ")
        return enc.Encode(data)
    case FormatYAML:
        enc := yaml.NewEncoder(p.out)
        enc.SetIndent(2)
        return enc.Encode(data)
    default:
        return p.printText(data)
    }
}

func (p *Printer) printText(data any) error {
    // Handle Tabular interface for table output
    if t, ok := data.(Tabular); ok {
        return p.printTable(t)
    }
    // Fallback to fmt
    _, err := fmt.Fprintln(p.out, data)
    return err
}

// Tabular interface for types that can be rendered as tables
type Tabular interface {
    Headers() []string
    Rows() [][]string
}

func (p *Printer) printTable(t Tabular) error {
    w := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
    defer w.Flush()

    // Headers
    for i, h := range t.Headers() {
        if i > 0 {
            fmt.Fprint(w, "\t")
        }
        fmt.Fprint(w, h)
    }
    fmt.Fprintln(w)

    // Rows
    for _, row := range t.Rows() {
        for i, cell := range row {
            if i > 0 {
                fmt.Fprint(w, "\t")
            }
            fmt.Fprint(w, cell)
        }
        fmt.Fprintln(w)
    }

    return nil
}
```

---

## 8. Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Markdown merge conflicts unresolvable | Medium | Medium | Flag conflicts for manual resolution; don't auto-merge |
| Schema validation performance | Low | Low | santhosh-tekuri is 2x faster; cache compiled schemas |
| Cross-platform XDG differences | Medium | Low | adrg/xdg handles automatically; test on CI |
| Behavioral parity with bash scripts | Medium | High | Golden file testing; gradual rollout |
| k8s strategic merge patch complexity | Medium | Medium | Start with simple JSON Merge Patch (RFC 7396) |

### Unknowns

| Unknown | Impact | Resolution |
|---------|--------|------------|
| Exact bash script behaviors | Could affect migration | Audit scripts, write parity tests |
| Schema complexity (nested refs, conditionals) | Could affect validation approach | Prototype with real schemas |
| Concurrent session access patterns | Could affect locking strategy | Design spike for locking |
| Windows compatibility requirements | Could affect path handling | Clarify with stakeholders |

---

## 9. Recommendations Summary

### Libraries to Adopt

| Purpose | Library | Version |
|---------|---------|---------|
| CLI Framework | `github.com/spf13/cobra` | v1.8+ |
| Config | `github.com/spf13/viper` | v1.18+ |
| JSON Schema | `github.com/santhosh-tekuri/jsonschema/v6` | v6+ |
| XDG Paths | `github.com/adrg/xdg` | v0.5+ |
| JSON Merge Patch | `github.com/evanphx/json-patch/v5` | v5+ |
| YAML | `gopkg.in/yaml.v3` | v3 |
| Markdown Parsing | `github.com/yuin/goldmark` | v1.6+ |

### Implementation Order

1. **Phase 0**: Skeleton with root command, XDG paths, embedded schemas
2. **Phase 1**: `validate` command (stateless, proves schema embedding)
3. **Phase 2**: `manifest` command (read operations, JSON output)
4. **Phase 3**: `sync` command (first write operations)
5. **Phase 4**: `team` command (depends on manifest)
6. **Phase 5**: `session` command (most stateful, replaces bash)

### Next Steps

1. **Create ariadne repo** with Phase 0 skeleton
2. **Port real schemas** from roster to test embedding
3. **Write parity tests** against existing bash scripts
4. **Prototype markdown merge** with goldmark to assess complexity
5. **Design session locking** strategy (file locks vs atomic renames)

---

## Appendix: Reference Links

### Official Documentation
- [Go embed Package](https://pkg.go.dev/embed)
- [Cobra Documentation](https://cobra.dev/)
- [Viper Documentation](https://github.com/spf13/viper)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

### Reference Implementations
- [GitHub CLI](https://github.com/cli/cli)
- [Terraform](https://github.com/hashicorp/terraform)
- [kubectl](https://github.com/kubernetes/kubernetes/tree/master/cmd/kubectl)
- [chezmoi](https://github.com/twpayne/chezmoi)

### Libraries
- [spf13/cobra](https://github.com/spf13/cobra)
- [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)
- [adrg/xdg](https://github.com/adrg/xdg)
- [k8s strategic merge patch](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/strategicpatch)
- [evanphx/json-patch](https://github.com/evanphx/json-patch)

### Design Patterns
- [Strangler Fig Pattern](https://martinfowler.com/bliki/StranglerFigApplication.html)
- [Three-Way Merge](https://blog.jcoglan.com/2017/05/08/merging-with-diff3/)
- [Strategic Merge Patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md)
