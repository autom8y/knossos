# TDD: Single-Binary Completion

**Status**: Implemented
**Author**: Architect
**Date**: 2026-02-06
**PRD**: N/A (Sprint-level initiative from MEMORY.md priorities)

## 1. Context

The `ari` binary requires the Knossos source repo on disk to resolve rites and templates. The lowest-priority tier in `SourceResolver` (`SourceKnossos`) resolves to `$KNOSSOS_HOME/rites/`, and the inscription `Generator` reads section templates from `knossos/templates/` on the filesystem. This means:

1. Installing `ari` from a Go binary (e.g. `go install`) does not work -- the binary cannot find rites
2. There is no `ari init` command to bootstrap a project without the source repo
3. Shell hook scripts still reference `USE_ARI_HOOKS` feature flags from the deprecation timeline (ADR-0011)

This sprint resolves all three issues. The binary becomes self-contained: embedded rites, an init command, and cleanup of the shell-to-Go migration residue.

## 2. Dependency Graph

```
Task 1: Rite Embedding -----> Task 2: ari init
                                (init uses embedded rites)

Task 3: Remaining Ports       (independent, no dependencies)
```

Task 2 depends on Task 1 because `ari init` must materialize from embedded rites when `KNOSSOS_HOME` is unset. Task 3 is independent cleanup.

## 3. Task 1: Rite Embedding

### 3.1 Problem

`//go:embed` directives require paths relative to the Go source file, and the paths must be within the module boundary. The content to embed lives at project root:

| Content | Path | Size |
|---------|------|------|
| Rites | `rites/` | ~2.0 MB |
| Templates | `knossos/templates/` | ~48 KB |
| Hooks config | `user-hooks/ari/hooks.yaml` | <1 KB |

### 3.2 Embed Strategy

**Decision**: Create a new package `embedded/` at the project root.

**Why not `cmd/ari/`?** The `cmd/ari/` directory is two levels deep. `//go:embed` paths cannot use `../` -- they must resolve within or below the file's directory. Placing the embed file at the module root in a dedicated package is the only viable option.

**Why a package, not just a file in the root?** Go convention: the root `main` package is `cmd/ari/main.go`. A root-level Go file would need to be in the `knossos` module package, which is not imported anywhere. A dedicated `embedded` package is clean and importable.

**File**: `/embedded/embedded.go`

```go
// Package embedded provides compiled-in rite definitions and templates.
package embedded

import "embed"

//go:embed rites
var Rites embed.FS

//go:embed knossos/templates
var Templates embed.FS

//go:embed user-hooks/ari/hooks.yaml
var HooksYAML []byte
```

**Wait -- `//go:embed` paths are relative to the source file.** If `embedded.go` is in `/embedded/`, then `rites` would look for `/embedded/rites/`, which does not exist. The content is at the module root.

**Revised strategy**: Place `embedded.go` at the **module root** in a package named after the module (`knossos`). But the module root already has no Go files and the module name is `github.com/autom8y/knossos`.

**Correct approach**: Create `embed.go` in the **module root directory** as part of a new package. Go allows this -- the file would declare `package embedded` but the directory path doesn't need to match if it is the module root. Actually, that is incorrect -- the directory name determines the package name convention but Go allows any package name.

**Final decision**: Symlinks and path tricks are fragile. The cleanest approach:

1. Create `/embedded/` directory
2. Create symlinks from `/embedded/rites` -> `../rites` and `/embedded/knossos` -> `../knossos`
3. Place embed directives in `/embedded/embedded.go`

**No -- symlinks are not followed by `//go:embed`.**

**Actual final decision**: Place `embed.go` directly in the module root directory. Go supports having Go files at the module root. The package name will be `knossos` (matching the last segment of the module path `github.com/autom8y/knossos`). This is an idiomatic pattern used by many projects (e.g., `github.com/user/project` with a root `project.go` file).

**File**: `/embed.go` (at module root)

```go
// Package knossos provides embedded rite definitions and templates
// for single-binary distribution.
package knossos

import "embed"

// EmbeddedRites contains all rite definitions from rites/.
// Access individual rites via fs.Sub(EmbeddedRites, "rites/<name>").
//
//go:embed rites
var EmbeddedRites embed.FS

// EmbeddedTemplates contains inscription templates from knossos/templates/.
// Access section templates via fs.Sub(EmbeddedTemplates, "knossos/templates/sections").
//
//go:embed knossos/templates
var EmbeddedTemplates embed.FS

// EmbeddedHooksYAML contains the hooks.yaml configuration.
//
//go:embed user-hooks/ari/hooks.yaml
var EmbeddedHooksYAML []byte
```

This works because:
- `rites/` is a directory at the module root, adjacent to `embed.go`
- `knossos/templates/` is a directory at the module root, adjacent to `embed.go`
- `user-hooks/ari/hooks.yaml` is a file at the module root, adjacent to `embed.go`
- The package `knossos` can be imported as `github.com/autom8y/knossos`

### 3.3 SourceResolver Changes

Add a new `SourceEmbedded` tier as the lowest-priority fallback (below `SourceKnossos`):

**File**: `/internal/materialize/source.go`

```go
// New source type
const (
    SourceEmbedded SourceType = "embedded"
)
```

Add an `embeddedFS` field to `SourceResolver`:

```go
type SourceResolver struct {
    projectRoot     string
    projectRitesDir string
    userRitesDir    string
    knossosHome     string
    embeddedFS      fs.FS  // NEW: embedded rites filesystem

    mu       sync.RWMutex
    resolved map[string]*ResolvedRite
}
```

**Constructor change** -- add `WithEmbeddedFS` option:

```go
// NewSourceResolver creates a new source resolver for the given project root.
func NewSourceResolver(projectRoot string) *SourceResolver {
    return &SourceResolver{
        projectRoot:     projectRoot,
        projectRitesDir: filepath.Join(projectRoot, "rites"),
        userRitesDir:    paths.UserRitesDir(),
        knossosHome:     config.KnossosHome(),
        resolved:        make(map[string]*ResolvedRite),
    }
}

// WithEmbeddedFS sets the embedded filesystem for fallback resolution.
func (r *SourceResolver) WithEmbeddedFS(fsys fs.FS) *SourceResolver {
    r.embeddedFS = fsys
    return r
}
```

**Resolution chain change** -- add tier 5 to `ResolveRite()`:

After the existing tier 4 (Knossos platform) check, add:

```go
// 5. Embedded rites (compiled-in fallback)
if result == nil && r.embeddedFS != nil {
    if res, err := r.checkEmbeddedSource(riteName); err == nil {
        result = res
    } else {
        checkedPaths = append(checkedPaths, "embedded://rites/"+riteName)
    }
}
```

**New method** -- `checkEmbeddedSource()`:

```go
// checkEmbeddedSource checks if a rite exists in the embedded filesystem.
func (r *SourceResolver) checkEmbeddedSource(riteName string) (*ResolvedRite, error) {
    manifestPath := "rites/" + riteName + "/manifest.yaml"
    if _, err := fs.Stat(r.embeddedFS, manifestPath); err != nil {
        return nil, err
    }

    return &ResolvedRite{
        Name: riteName,
        Source: RiteSource{
            Type:        SourceEmbedded,
            Path:        "embedded://rites/" + riteName,
            Description: "compiled-in rite definition",
        },
        RitePath:     "rites/" + riteName,
        ManifestPath: manifestPath,
        TemplatesDir: "knossos/templates",
    }, nil
}
```

**`ListAvailableRites` change**: Add embedded rites as the lowest-priority source:

```go
// After iterating filesystem sources, add embedded rites
if r.embeddedFS != nil {
    entries, err := fs.ReadDir(r.embeddedFS, "rites")
    if err == nil {
        for _, entry := range entries {
            if !entry.IsDir() || seen[entry.Name()] {
                continue
            }
            if resolved, err := r.checkEmbeddedSource(entry.Name()); err == nil {
                result = append(result, *resolved)
                seen[entry.Name()] = true
            }
        }
    }
}
```

### 3.4 Materializer Changes

The critical challenge: `Materializer` methods use `os.ReadFile`, `os.ReadDir`, and `filepath.WalkDir` throughout. When the source is embedded, these must use `fs.FS` equivalents instead.

**Strategy: Abstract file access behind the resolved source.**

Add a new method to `Materializer` that returns an `fs.FS` rooted at the rite path:

```go
// riteFS returns a filesystem rooted at the rite's directory.
// For embedded sources, returns the embedded FS.
// For filesystem sources, returns os.DirFS rooted at the rite path.
func (m *Materializer) riteFS(resolved *ResolvedRite) fs.FS {
    if resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil {
        sub, err := fs.Sub(m.sourceResolver.embeddedFS, resolved.RitePath)
        if err != nil {
            return os.DirFS(resolved.RitePath)
        }
        return sub
    }
    return os.DirFS(resolved.RitePath)
}

// templatesFS returns a filesystem for templates.
// For embedded sources, returns the embedded templates FS.
// For filesystem sources, returns os.DirFS rooted at the templates dir.
func (m *Materializer) templatesFS(resolved *ResolvedRite) fs.FS {
    if resolved.Source.Type == SourceEmbedded && m.sourceResolver.embeddedFS != nil {
        sub, err := fs.Sub(m.sourceResolver.embeddedFS, resolved.TemplatesDir)
        if err != nil {
            return os.DirFS(m.templatesDir)
        }
        return sub
    }
    return os.DirFS(m.templatesDir)
}
```

**Methods that need refactoring to use `fs.FS`:**

| Method | Current | Change |
|--------|---------|--------|
| `loadRiteManifest` | `os.ReadFile(manifestPath)` | Accept `fs.FS`, use `fs.ReadFile(fsys, "manifest.yaml")` |
| `materializeAgents` | `os.ReadFile`, `filepath.WalkDir` | Accept `fs.FS`, use `fs.WalkDir`, `fs.ReadFile` |
| `materializeMena` | `os.ReadDir`, `os.Stat`, `os.ReadFile` | Accept `fs.FS` for reading sources; writes still go to `os` |
| `copyDir` | `filepath.WalkDir`, `os.ReadFile` | New `copyDirFromFS(fsys fs.FS, srcRoot, dst)` variant |
| `materializeHooks` | `os.Stat(sourceHooksDir)`, `m.copyDir` | Use `templatesFS` |
| `loadHooksConfig` | `os.ReadFile` candidate paths | Add embedded bytes as highest-priority candidate |

**Key refactoring pattern**: Each method that reads source content gains a `fromFS` variant or an `fs.FS` parameter. Write operations (to `.claude/`) remain `os.*` calls.

```go
// copyDirFromFS copies all files from an fs.FS to a destination directory on disk.
func copyDirFromFS(fsys fs.FS, dst string) error {
    return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        destPath := filepath.Join(dst, path)
        if d.IsDir() {
            return os.MkdirAll(destPath, 0755)
        }
        content, err := fs.ReadFile(fsys, path)
        if err != nil {
            return err
        }
        return os.WriteFile(destPath, content, 0644)
    })
}
```

### 3.5 Inscription Generator Changes

The `Generator` reads templates via `os.ReadFile` and `os.Stat`. Add an optional `fs.FS` field:

**File**: `/internal/inscription/generator.go`

```go
type Generator struct {
    TemplateDir      string
    TemplateFS       fs.FS  // NEW: optional fs.FS for template access
    Manifest         *Manifest
    Context          *RenderContext
    templates        *template.Template
    sectionTemplates map[string]string
}

// NewGeneratorWithFS creates a generator that reads templates from an fs.FS.
func NewGeneratorWithFS(templateFS fs.FS, manifest *Manifest, ctx *RenderContext) *Generator {
    return &Generator{
        TemplateFS:       templateFS,
        Manifest:         manifest,
        Context:          ctx,
        sectionTemplates: make(map[string]string),
    }
}
```

Modify `getSectionTemplatePath` and `renderTemplateFile` to check `TemplateFS` first:

```go
func (g *Generator) getSectionTemplatePath(regionName string) string {
    path := "sections/" + regionName + ".md.tpl"

    // Try embedded FS first
    if g.TemplateFS != nil {
        if _, err := fs.Stat(g.TemplateFS, path); err == nil {
            return path  // Return relative path for FS lookup
        }
    }

    // Fall back to filesystem
    if g.TemplateDir != "" {
        absPath := filepath.Join(g.TemplateDir, "sections", regionName+".md.tpl")
        if _, err := os.Stat(absPath); err == nil {
            return absPath
        }
    }

    return ""
}

func (g *Generator) renderTemplateFile(path string) (string, error) {
    var data []byte
    var err error

    if g.TemplateFS != nil {
        data, err = fs.ReadFile(g.TemplateFS, path)
    } else {
        data, err = os.ReadFile(path)
    }
    if err != nil {
        return "", errors.Wrap(errors.CodeFileNotFound, "failed to read template file", err)
    }

    return g.renderTemplateString(filepath.Base(path), string(data))
}
```

Similarly, the `includePartial` function needs the same dual-path logic.

### 3.6 Hooks Config Changes

**File**: `/internal/materialize/hooks.go`

`loadHooksConfig` currently reads from filesystem candidates. Add embedded hooks.yaml as the lowest-priority fallback:

```go
func (m *Materializer) loadHooksConfig() *HooksConfig {
    // ... existing filesystem candidates ...

    // Fallback: embedded hooks.yaml
    if m.embeddedHooksYAML != nil {
        var cfg HooksConfig
        if err := yaml.Unmarshal(m.embeddedHooksYAML, &cfg); err == nil {
            if cfg.SchemaVersion == "2.0" {
                return &cfg
            }
        }
    }

    return nil
}
```

Add `embeddedHooksYAML []byte` field to `Materializer` and a `WithEmbeddedHooks` setter.

### 3.7 Wiring: cmd/ari/main.go

Import the root `knossos` package and pass embedded assets down:

```go
import (
    "github.com/autom8y/knossos"
    "github.com/autom8y/knossos/internal/cmd/root"
)

func main() {
    root.SetVersion(version, commit, date)
    common.SetEmbeddedAssets(knossos.EmbeddedRites, knossos.EmbeddedTemplates, knossos.EmbeddedHooksYAML)
    if err := root.Execute(); err != nil {
        // ...
    }
}
```

`common.SetEmbeddedAssets` stores these in package-level vars accessible to command constructors. The `Materializer` constructors (`NewMaterializer`, `NewMaterializerWithSource`) gain access via a new global accessor:

```go
// In internal/cmd/common/embedded.go
var (
    embeddedRites     fs.FS
    embeddedTemplates fs.FS
    embeddedHooksYAML []byte
)

func SetEmbeddedAssets(rites, templates fs.FS, hooks []byte) {
    embeddedRites = rites
    embeddedTemplates = templates
    embeddedHooksYAML = hooks
}

func EmbeddedRites() fs.FS     { return embeddedRites }
func EmbeddedTemplates() fs.FS { return embeddedTemplates }
func EmbeddedHooksYAML() []byte { return embeddedHooksYAML }
```

### 3.8 Files Changed (Task 1)

| File | Action | Description |
|------|--------|-------------|
| `/embed.go` | **CREATE** | Root package with `//go:embed` directives |
| `/internal/materialize/source.go` | MODIFY | Add `SourceEmbedded`, `embeddedFS` field, `checkEmbeddedSource`, update `ListAvailableRites` |
| `/internal/materialize/materialize.go` | MODIFY | Add `riteFS`/`templatesFS` helpers, refactor `loadRiteManifest`/`materializeAgents`/`materializeMena`/`materializeHooks`/`copyDir` to support `fs.FS` |
| `/internal/materialize/hooks.go` | MODIFY | Add embedded hooks.yaml fallback, `embeddedHooksYAML` field |
| `/internal/inscription/generator.go` | MODIFY | Add `TemplateFS` field, `NewGeneratorWithFS`, dual-path template loading |
| `/internal/cmd/root/root.go` | MODIFY | Add `SetEmbeddedAssets`, embedded asset accessors |
| `/cmd/ari/main.go` | MODIFY | Import root knossos package, wire embedded assets |

## 4. Task 2: ari init

### 4.1 Command Design

```
ari init [--rite <name>] [--source <path>] [--force]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--rite` | (none) | Activate a rite after scaffolding |
| `--source` | (none) | Explicit source path (bypass resolution) |
| `--force` | false | Overwrite existing `.claude/` if present |

**Behavior**:
1. If `.claude/` exists and `--force` is not set, exit with informative message
2. Create `.claude/` directory
3. Generate `KNOSSOS_MANIFEST.yaml` (inscription system default)
4. Generate `settings.local.json` (with hooks from embedded or filesystem)
5. Generate `CLAUDE.md` (minimal, or rite-specific if `--rite` given)
6. If `--rite` specified, run full materialization (agents, mena, hooks, inscription)
7. Print success summary

**`needsProject` annotation**: `false`. This command must work without an existing project context -- it creates the project context. However, it does need a directory to work in. The command will use the current working directory (or `--project-dir`) as the target.

### 4.2 Integration with Materializer

When `--rite` is specified:
- If `KNOSSOS_HOME` is set, use normal 4-tier resolution
- If `KNOSSOS_HOME` is unset, embedded rites provide the fallback
- This "just works" because Task 1 adds embedded as tier 5

When no `--rite`:
- Use `MaterializeMinimal()` to scaffold base infrastructure

### 4.3 Idempotency

On an existing Knossos project (`.knossos/KNOSSOS_MANIFEST.yaml` exists):
- Without `--force`: Print "already initialized" and exit 0 (not an error)
- With `--force`: Re-run full scaffolding, preserving satellite regions in CLAUDE.md

On a project with `.claude/` but no `KNOSSOS_MANIFEST.yaml` (non-Knossos project):
- Without `--force`: Print warning that `.claude/` exists but is not Knossos-managed, suggest `--force`
- With `--force`: Initialize, preserving existing `settings.local.json` MCP config

### 4.4 Command Registration

**File**: `/internal/cmd/init/init.go` (new file)

```go
package init

import (
    "github.com/spf13/cobra"
    "github.com/autom8y/knossos/internal/cmd/common"
)

func NewInitCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
    var opts initOptions

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize a Knossos project",
        Long:  `Scaffolds .claude/ directory with CLAUDE.md, settings.local.json,
and KNOSSOS_MANIFEST.yaml. Optionally activates a rite.

Works without KNOSSOS_HOME set -- uses embedded rite definitions.

Examples:
  ari init                    # Minimal scaffold
  ari init --rite 10x-dev     # Scaffold with 10x-dev rite
  ari init --force            # Re-initialize existing project`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runInit(ctx, opts)
        },
    }

    // This command does NOT require an existing project
    common.SetNeedsProject(cmd, false, false)

    cmd.Flags().StringVar(&opts.rite, "rite", "", "Rite to activate after scaffolding")
    cmd.Flags().StringVar(&opts.source, "source", "", "Explicit rite source path")
    cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing .claude/ directory")

    return cmd
}
```

**Registration in root.go**:

```go
import initcmd "github.com/autom8y/knossos/internal/cmd/init"

// In init():
rootCmd.AddCommand(initcmd.NewInitCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
```

Note: The package is named `init` which conflicts with Go's `init()` function. **Revised**: Name the package `initialize` or use an import alias. The directory will be `/internal/cmd/initialize/` with `package initialize`.

### 4.5 Implementation Flow

```go
func runInit(ctx *cmdContext, opts initOptions) error {
    projectDir := ctx.resolveProjectDir() // cwd if not specified
    claudeDir := filepath.Join(projectDir, ".claude")

    // 1. Check existing state
    manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
    if _, err := os.Stat(manifestPath); err == nil && !opts.force {
        printer.PrintSuccess("Already initialized (use --force to reinitialize)")
        return nil
    }
    if _, err := os.Stat(claudeDir); err == nil && !opts.force {
        return errors.New(errors.CodeUsageError,
            ".claude/ exists but is not Knossos-managed. Use --force to initialize.")
    }

    // 2. Create resolver targeting projectDir
    resolver := paths.NewResolver(projectDir)

    // 3. Create materializer with embedded assets
    mat := materialize.NewMaterializer(resolver)
    mat.WithEmbeddedFS(root.EmbeddedRites())
    mat.WithEmbeddedTemplates(root.EmbeddedTemplates())
    mat.WithEmbeddedHooks(root.EmbeddedHooksYAML())

    // 4. Materialize
    if opts.rite != "" {
        result, err := mat.MaterializeWithOptions(opts.rite, materialize.Options{
            Force: opts.force,
        })
        // ... handle result
    } else {
        result, err := mat.MaterializeMinimal(materialize.Options{})
        // ... handle result
    }

    // 5. Print summary
    return printer.PrintSuccess(initOutput{...})
}
```

### 4.6 Files Changed (Task 2)

| File | Action | Description |
|------|--------|-------------|
| `/internal/cmd/initialize/init.go` | **CREATE** | `ari init` command implementation |
| `/internal/cmd/root/root.go` | MODIFY | Register `ari init` command |

## 5. Task 3: Remaining Ports (Cleanup)

### 5.1 USE_ARI_HOOKS Feature Flag Removal

The following shell scripts contain `USE_ARI_HOOKS` checks:

| File | Line | Pattern |
|------|------|---------|
| `user-hooks/ari/autopark.sh` | 12 | `[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0` |
| `user-hooks/ari/route.sh` | 22 | `[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0` |
| `user-hooks/ari/writeguard.sh` | 26 | `[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0` |
| `user-hooks/ari/cognitive-budget.sh` | 13 | `[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0` |

**Action**: Remove these feature flag lines from all four scripts. The hooks are now unconditionally Go-powered (ADR-0011 Phase 2+3 complete per commit `bb1e9b6`). The default was already flipped to enabled (`26d91a5`), making the flag a no-op for all standard users.

Also search for `USE_ARI_HOOKS` references in:
- Documentation (remove/update references)
- Go code (verify no `os.Getenv("USE_ARI_HOOKS")` calls exist -- confirmed: none found)

### 5.2 Stale Documentation References

Check for references to deleted shell scripts in:
- ADRs
- TDDs
- CLAUDE.md templates
- Code comments

Per commit `26d91a5` ("fix stale comments referencing deleted shell scripts"), some cleanup was already done. Verify completeness.

### 5.3 exec.Command to .sh Verification

Confirmed: `grep` for `exec\.Command.*\.sh` in Go files returns **zero matches**. No remaining `exec.Command` calls to shell scripts exist. This sub-task is already complete.

### 5.4 Files Changed (Task 3)

| File | Action | Description |
|------|--------|-------------|
| `user-hooks/ari/autopark.sh` | MODIFY | Remove USE_ARI_HOOKS check |
| `user-hooks/ari/route.sh` | MODIFY | Remove USE_ARI_HOOKS check |
| `user-hooks/ari/writeguard.sh` | MODIFY | Remove USE_ARI_HOOKS check |
| `user-hooks/ari/cognitive-budget.sh` | MODIFY | Remove USE_ARI_HOOKS check |
| Various docs | MODIFY | Remove stale references if found |

## 6. Interface Specifications

### 6.1 SourceResolver Extended Interface

```go
// SourceType additions
const SourceEmbedded SourceType = "embedded"

// SourceResolver -- new methods and fields
type SourceResolver struct {
    // ... existing fields ...
    embeddedFS fs.FS  // Embedded rites filesystem
}

func (r *SourceResolver) WithEmbeddedFS(fsys fs.FS) *SourceResolver
func (r *SourceResolver) checkEmbeddedSource(riteName string) (*ResolvedRite, error)
```

Resolution order becomes:
1. Explicit (--source flag)
2. Project (`./rites/`)
3. User (`~/.local/share/knossos/rites/`)
4. Knossos (`$KNOSSOS_HOME/rites/`)
5. **Embedded** (compiled-in) -- NEW

### 6.2 Materializer Extended Interface

```go
type Materializer struct {
    // ... existing fields ...
    embeddedHooksYAML []byte
}

func (m *Materializer) WithEmbeddedFS(fsys fs.FS)
func (m *Materializer) WithEmbeddedTemplates(fsys fs.FS)
func (m *Materializer) WithEmbeddedHooks(data []byte)

// New helper
func copyDirFromFS(fsys fs.FS, dst string) error
```

### 6.3 Generator Extended Interface

```go
type Generator struct {
    TemplateDir      string
    TemplateFS       fs.FS   // NEW
    // ... existing fields ...
}

func NewGeneratorWithFS(templateFS fs.FS, manifest *Manifest, ctx *RenderContext) *Generator
```

### 6.4 ari init Command Interface

```
Usage:
  ari init [flags]

Flags:
      --force           Overwrite existing .claude/ directory
      --rite string     Rite to activate after scaffolding
      --source string   Explicit rite source path
  -h, --help            help for init

Global Flags:
  -o, --output string       Output format: text, json, yaml (default "text")
  -p, --project-dir string  Project root directory (overrides discovery)
  -v, --verbose             Enable verbose output
```

**Exit codes**:
- 0: Success (including "already initialized" without --force)
- 1: General error
- 2: Usage error (invalid flags)

**JSON output** (`-o json`):
```json
{
  "initialized": true,
  "project_dir": "/path/to/project",
  "rite": "10x-dev",
  "source": "embedded",
  "artifacts": [
    ".claude/CLAUDE.md",
    ".knossos/KNOSSOS_MANIFEST.yaml",
    ".claude/settings.local.json"
  ]
}
```

## 7. Data Model

No new persistent data models. The existing `KNOSSOS_MANIFEST.yaml`, `settings.local.json`, and `ACTIVE_RITE` files are used unchanged.

The `embed.FS` is read-only and has no state. Embedded content is a compile-time snapshot of the `rites/`, `knossos/templates/`, and `user-hooks/ari/hooks.yaml` directories.

## 8. Key Sequence: ari init --rite 10x-dev (No KNOSSOS_HOME)

```
User runs: ari init --rite 10x-dev
    |
    v
PersistentPreRunE (root.go)
    - needsProject(cmd) returns false (annotation)
    - Skips project root discovery
    |
    v
runInit (initialize/init.go)
    - projectDir = cwd
    - Check .claude/ doesn't exist (or --force)
    - Create paths.Resolver for projectDir
    |
    v
NewMaterializer(resolver)
    .WithEmbeddedFS(embeddedRites)
    .WithEmbeddedTemplates(embeddedTemplates)
    .WithEmbeddedHooks(embeddedHooksYAML)
    |
    v
MaterializeWithOptions("10x-dev", opts)
    |
    v
SourceResolver.ResolveRite("10x-dev", "")
    - Tier 1 (explicit): skip (no --source)
    - Tier 2 (project ./rites/): not found
    - Tier 3 (user): not found
    - Tier 4 (knossos): KNOSSOS_HOME unset, skip
    - Tier 5 (embedded): fs.Stat("rites/10x-dev/manifest.yaml") -> FOUND
    - Returns ResolvedRite{Source: SourceEmbedded}
    |
    v
loadRiteManifest via fs.ReadFile(embeddedFS, "rites/10x-dev/manifest.yaml")
    |
    v
materializeAgents -- reads from embeddedFS, writes to .claude/agents/
materializeMena -- reads mena sources from embeddedFS, writes to .claude/commands/ + .claude/skills/
materializeHooks -- reads from embeddedTemplatesFS, writes to .claude/hooks/
materializeCLAUDEmd -- Generator uses embeddedTemplatesFS for section templates
materializeSettings -- uses embeddedHooksYAML for hooks.yaml config
    |
    v
Output: "Initialized with rite '10x-dev' (source: embedded)"
```

## 9. Test Strategy

### 9.1 Task 1 Tests

**Unit tests for `SourceResolver` with embedded FS**:

File: `/internal/materialize/source_test.go`

```go
func TestSourceResolver_EmbeddedFallback(t *testing.T) {
    // Create in-memory fs.FS with a test rite
    fsys := fstest.MapFS{
        "rites/test-rite/manifest.yaml": &fstest.MapFile{
            Data: []byte("name: test-rite\nversion: 1.0\n"),
        },
    }

    resolver := NewSourceResolver("/nonexistent")
    resolver.WithEmbeddedFS(fsys)

    // Should find rite in embedded FS when filesystem sources don't exist
    resolved, err := resolver.ResolveRite("test-rite", "")
    require.NoError(t, err)
    assert.Equal(t, SourceEmbedded, resolved.Source.Type)
}

func TestSourceResolver_FilesystemOverridesEmbedded(t *testing.T) {
    // Create temp dir with project rites
    // Create embedded FS with same rite
    // Verify filesystem version wins
}

func TestSourceResolver_ListIncludesEmbedded(t *testing.T) {
    // Verify ListAvailableRites includes embedded rites
    // Verify shadowing: filesystem rites hide embedded ones
}
```

**Unit tests for `copyDirFromFS`**:

```go
func TestCopyDirFromFS(t *testing.T) {
    fsys := fstest.MapFS{
        "agents/foo.md": &fstest.MapFile{Data: []byte("# Foo")},
        "agents/bar.md": &fstest.MapFile{Data: []byte("# Bar")},
    }
    sub, _ := fs.Sub(fsys, "agents")
    dst := t.TempDir()
    err := copyDirFromFS(sub, dst)
    require.NoError(t, err)
    // Verify files were written
}
```

**Integration test for embedded materialization**:

```go
func TestMaterialize_FromEmbedded(t *testing.T) {
    // Use testing/fstest.MapFS to build a synthetic rite
    // Create Materializer with embedded FS
    // Run MaterializeWithOptions
    // Verify .claude/ contents match expected
}
```

**Compile-time verification**:

```go
func TestEmbeddedRites_ContainsKnownRites(t *testing.T) {
    // Import the root knossos package
    // Verify fs.Stat(knossos.EmbeddedRites, "rites/10x-dev/manifest.yaml") succeeds
    // Verify at least N rites are present
}
```

### 9.2 Task 2 Tests

**Unit tests for init command**:

```go
func TestInit_FreshDirectory(t *testing.T) {
    dir := t.TempDir()
    // Run init command targeting dir
    // Verify .knossos/KNOSSOS_MANIFEST.yaml exists
    // Verify .claude/CLAUDE.md exists
    // Verify .claude/settings.local.json exists
}

func TestInit_WithRite(t *testing.T) {
    dir := t.TempDir()
    // Run init --rite 10x-dev targeting dir (with embedded FS)
    // Verify .claude/agents/ contains expected agents
    // Verify .knossos/ACTIVE_RITE contains "10x-dev"
}

func TestInit_Idempotent(t *testing.T) {
    dir := t.TempDir()
    // Init once
    // Init again without --force -> should succeed with "already initialized"
    // Init again with --force -> should reinitialize
}

func TestInit_NeedsProjectFalse(t *testing.T) {
    // Verify the command annotation is set correctly
    cmd := initialize.NewInitCmd(...)
    assert.False(t, common.NeedsProject(cmd))
}
```

### 9.3 Task 3 Tests

No new tests needed. Verify by:
1. `grep -r USE_ARI_HOOKS` returns only the hooks.yaml config (which defines when hooks fire, not a feature flag)
2. Existing test suite passes without the feature flag

## 10. Risk Assessment

### 10.1 Binary Size Increase

**Risk**: Embedding 2MB of rites + 48KB of templates + 1KB hooks.yaml increases binary size by ~2MB.

**Assessment**: LOW. The current binary is already substantial (Go binaries typically 10-30MB). 2MB is <10% increase. Rite content is text (markdown, YAML) which `embed.FS` stores efficiently.

**Mitigation**: Monitor binary size. If needed, embed only a subset of rites (e.g., just `10x-dev` and `shared`).

### 10.2 Stale Embedded Content

**Risk**: Embedded rites are a compile-time snapshot. Users running older binaries get outdated rites.

**Assessment**: MEDIUM. This is inherent to any compiled-in content approach.

**Mitigation**:
- Filesystem sources (tiers 1-4) always override embedded content
- `ari init --rite X` will use filesystem sources when available
- Binary version is already tracked (`ari version`)
- Future: `ari update` could refresh from upstream

### 10.3 Package Name Conflict at Module Root

**Risk**: Adding `embed.go` at the module root creates a `package knossos` file. If other root-level Go files exist or are added, they must use the same package name.

**Assessment**: LOW. The module root currently has no Go files. The `embed.go` file is the only one needed.

**Mitigation**: Document in the file header that this package exists solely for embedding. Lint rules can prevent accidental additions.

### 10.4 Circular Import Risk

**Risk**: `cmd/ari/main.go` imports `github.com/autom8y/knossos` (root package) and `internal/cmd/root`. If `internal/cmd/root` also imports the root package, circular dependency occurs.

**Assessment**: LOW. The wiring pattern passes embedded assets from `main.go` -> `common.SetEmbeddedAssets()` -> stored in common package vars. No circular import: `main` imports both `knossos` (root) and `common` (internal). `common` does NOT import `knossos` (root). The root package has no dependencies on internal packages.

**Mitigation**: Keep the root `knossos` package dependency-free (only `embed` stdlib import).

### 10.5 fs.FS Compatibility Surface

**Risk**: Refactoring `Materializer` and `Generator` to support `fs.FS` alongside `os.*` creates a dual-path maintenance burden. Bugs could appear in one path but not the other.

**Assessment**: MEDIUM. This is the largest risk in the design.

**Mitigation**:
- The `copyDirFromFS` helper abstracts the common pattern
- `os.DirFS()` bridges filesystem paths to `fs.FS`, allowing a single code path once refactored
- Long-term: refactor ALL source reading to use `fs.FS`, eliminating the dual path. Filesystem sources would use `os.DirFS()` wrapped as `fs.FS`.

### 10.6 ari init on Existing Non-Knossos Projects

**Risk**: User runs `ari init --force` on a project that already has `.claude/settings.local.json` with custom MCP server config. Materialization could clobber it.

**Assessment**: LOW. The `materializeSettingsWithManifest` function already merges into existing settings rather than replacing them. MCP servers are additive.

**Mitigation**: Existing merge logic in `materializeSettingsWithManifest` handles this. The init command should explicitly test this scenario.

## 11. Implementation Order

Within Task 1 (largest task), implement in this order:

1. **Create `/embed.go`** -- the embed directives. Verify with `go build`.
2. **Add `SourceEmbedded` to source.go** -- new type, `embeddedFS` field, `WithEmbeddedFS`, `checkEmbeddedSource`. Write tests.
3. **Add `copyDirFromFS` helper** -- the shared utility. Write tests.
4. **Refactor `loadRiteManifest`** -- accept `fs.FS`, simplest method to convert.
5. **Refactor `materializeAgents`** -- use `copyDirFromFS` when source is embedded.
6. **Refactor `materializeMena`** -- most complex; handle both filesystem and embedded mena sources.
7. **Refactor `materializeHooks`** -- use templates FS.
8. **Refactor inscription `Generator`** -- add `TemplateFS`, dual-path template loading.
9. **Add `loadHooksConfig` embedded fallback** -- in hooks.go.
10. **Wire in root.go and main.go** -- `SetEmbeddedAssets`, import root package.
11. **Integration tests** -- end-to-end materialization from embedded FS.

Then Task 2:

12. **Create `/internal/cmd/initialize/init.go`** -- the command.
13. **Register in root.go**.
14. **Write tests**.

Task 3 (can be done in parallel):

15. **Remove `USE_ARI_HOOKS` lines** from shell scripts.
16. **Verify no stale doc references**.

## 12. ADR References

| ADR | Relevance |
|-----|-----------|
| ADR-0008 (Handoff Schema Embedding) | Precedent for `//go:embed` pattern in this codebase |
| ADR-0011 (Hook Deprecation Timeline) | Justifies Task 3 cleanup |
| ADR-0024 (Agent Factory) | Shows `//go:embed templates/*.md.tpl` pattern already in use |
| ADR-sync-materialization | Documents the materialization pipeline being extended |

A new ADR (ADR-0025) should be created for the rite embedding decision, documenting:
- Why embedded over alternative distribution mechanisms (e.g., HTTP fetch, separate data package)
- The 5-tier resolution order
- Binary size tradeoffs
- Staleness management strategy

---

## Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/knossos/docs/decisions/TDD-single-binary-completion.md` | Written |
| Source (SourceResolver) | `/Users/tomtenuta/Code/knossos/internal/materialize/source.go` | Read |
| Source (Materializer) | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | Read |
| Source (Generator) | `/Users/tomtenuta/Code/knossos/internal/inscription/generator.go` | Read |
| Source (hooks.go) | `/Users/tomtenuta/Code/knossos/internal/materialize/hooks.go` | Read |
| Source (root.go) | `/Users/tomtenuta/Code/knossos/internal/cmd/root/root.go` | Read |
| Source (main.go) | `/Users/tomtenuta/Code/knossos/cmd/ari/main.go` | Read |
| Source (annotations.go) | `/Users/tomtenuta/Code/knossos/internal/cmd/common/annotations.go` | Read |
| Source (home.go) | `/Users/tomtenuta/Code/knossos/internal/config/home.go` | Read |
| Source (switch.go) | `/Users/tomtenuta/Code/knossos/internal/rite/switch.go` | Read |
