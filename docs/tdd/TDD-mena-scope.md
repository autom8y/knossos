# TDD: MenaScope Filtering

```yaml
status: accepted
date: 2026-02-07
author: architect
prd: docs/prd/PRD-mena-scope.md
adr: docs/decisions/ADR-0025-mena-scope.md
complexity: low
```

## Overview

Add a `scope` field to `MenaFrontmatter` so mena entries can declare which distribution pipeline(s) they target. Both the materialize pipeline (project-level) and usersync pipeline (user-level) filter entries based on scope at collection time. The zero value means "both pipelines" for full backward compatibility.

## 1. MenaScope Type

**File**: `internal/materialize/frontmatter.go`

Define a string type and three constants immediately after the existing `MenaFrontmatter` struct definition:

```go
// MenaScope controls which distribution pipeline(s) include a mena entry.
// The zero value (empty string) means "both pipelines" for backward compatibility.
type MenaScope string

const (
	// MenaScopeBoth is the zero value -- entry is included in both pipelines.
	MenaScopeBoth MenaScope = ""

	// MenaScopeUser restricts the entry to the usersync pipeline (~/.claude/).
	MenaScopeUser MenaScope = "user"

	// MenaScopeProject restricts the entry to the materialize pipeline (.claude/).
	MenaScopeProject MenaScope = "project"
)
```

Add a validation helper:

```go
// ValidScope returns true if s is a recognized MenaScope value.
func (s MenaScope) ValidScope() bool {
	switch s {
	case MenaScopeBoth, MenaScopeUser, MenaScopeProject:
		return true
	default:
		return false
	}
}

// String returns the string representation of the scope.
// Returns "both" for the zero value to aid logging/debugging.
func (s MenaScope) String() string {
	if s == MenaScopeBoth {
		return "both"
	}
	return string(s)
}
```

**Rationale**: String type (not int enum) per stakeholder decision. The zero value `""` maps to "both" so every existing file without a `scope` field is automatically backward-compatible. `ValidScope()` is a method on the type rather than a standalone function so it reads naturally: `scope.ValidScope()`.

## 2. Frontmatter Changes

**File**: `internal/materialize/frontmatter.go`

Add the `Scope` field to `MenaFrontmatter`:

```go
type MenaFrontmatter struct {
	// Identity (required for all)
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Distribution Control
	Scope MenaScope `yaml:"scope,omitempty"`

	// Invocation Control
	ArgumentHint               string              `yaml:"argument-hint,omitempty"`
	Triggers                   FlexibleStringSlice `yaml:"triggers,omitempty"`
	AllowedTools               FlexibleStringSlice `yaml:"allowed-tools,omitempty"`
	Model                      string              `yaml:"model,omitempty"`
	DisableModelInvocation     bool                `yaml:"disable-model-invocation,omitempty"`

	// Optional Metadata
	Version      string `yaml:"version,omitempty"`
	Deprecated   bool   `yaml:"deprecated,omitempty"`
	DeprecatedBy string `yaml:"deprecated-by,omitempty"`
}
```

Extend `Validate()` to check scope:

```go
func (f *MenaFrontmatter) Validate() error {
	if f.Name == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: name is required")
	}
	if f.Description == "" {
		return errors.New(errors.CodeValidationFailed, "frontmatter: description is required")
	}
	if !f.Scope.ValidScope() {
		return errors.New(errors.CodeValidationFailed,
			fmt.Sprintf("frontmatter: invalid scope %q (must be \"user\", \"project\", or omitted)", string(f.Scope)))
	}
	return nil
}
```

Note: This requires adding `"fmt"` to the imports in `frontmatter.go`.

## 3. PipelineScope on MenaProjectionOptions

**File**: `internal/materialize/project_mena.go`

Add the field to `MenaProjectionOptions`:

```go
type MenaProjectionOptions struct {
	Mode   MenaProjectionMode
	Filter MenaFilter

	// PipelineScope indicates which pipeline is calling ProjectMena().
	// When set, entries whose scope excludes this pipeline are skipped.
	// When empty (zero value), no scope filtering is applied (backward compat).
	PipelineScope MenaScope

	TargetCommandsDir string
	TargetSkillsDir   string
}
```

Add a scope-match helper function (unexported, package-level):

```go
// scopeIncludesPipeline returns true if the entry's scope allows inclusion
// in the given pipeline. Returns true when either value is the zero value
// (MenaScopeBoth), providing backward compatibility for callers that do not
// set PipelineScope and for entries that do not set scope.
func scopeIncludesPipeline(entryScope, pipelineScope MenaScope) bool {
	if pipelineScope == MenaScopeBoth {
		return true // Caller did not set pipeline scope -- no filtering
	}
	if entryScope == MenaScopeBoth {
		return true // Entry has no scope restriction
	}
	return entryScope == pipelineScope
}
```

**Truth table**:

| entryScope | pipelineScope | Result | Reason |
|------------|---------------|--------|--------|
| `""` | `""` | true | No filtering on either side |
| `""` | `"project"` | true | Entry goes to both, pipeline is project |
| `""` | `"user"` | true | Entry goes to both, pipeline is user |
| `"user"` | `""` | true | No pipeline filtering requested |
| `"user"` | `"user"` | true | Match |
| `"user"` | `"project"` | false | Entry is user-only, pipeline is project |
| `"project"` | `""` | true | No pipeline filtering requested |
| `"project"` | `"user"` | false | Entry is project-only, pipeline is user |
| `"project"` | `"project"` | true | Match |

## 4. Frontmatter Parsing Helper

**File**: `internal/materialize/project_mena.go`

Four helpers that read INDEX files and extract `MenaFrontmatter`. Two are exported (for cross-package use by usersync), two are unexported (package-internal). These are needed by `ProjectMena()` to determine scope at collection/routing time without requiring the full file content.

```go
// ReadMenaFrontmatterFromDir reads the INDEX file from a filesystem directory,
// parses its YAML frontmatter, and returns the result.
// Returns a zero-value MenaFrontmatter (scope="") if the INDEX file has no
// frontmatter or if parsing fails (with a logged warning for parse failures).
func ReadMenaFrontmatterFromDir(dirPath string) MenaFrontmatter {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return MenaFrontmatter{}
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			data, err := os.ReadFile(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				return MenaFrontmatter{}
			}
			return parseMenaFrontmatterBytes(data)
		}
	}
	return MenaFrontmatter{}
}

// readMenaFrontmatterFromFS reads the INDEX file from an fs.FS path,
// parses its YAML frontmatter, and returns the result. Unexported: only
// used within the materialize package by ProjectMena() for embedded sources.
// Returns a zero-value MenaFrontmatter (scope="") if the INDEX file has no
// frontmatter or if parsing fails.
func readMenaFrontmatterFromFS(fsys fs.FS, dirPath string) MenaFrontmatter {
	entries, err := fs.ReadDir(fsys, dirPath)
	if err != nil {
		return MenaFrontmatter{}
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "INDEX") {
			data, err := fs.ReadFile(fsys, dirPath+"/"+entry.Name())
			if err != nil {
				return MenaFrontmatter{}
			}
			return parseMenaFrontmatterBytes(data)
		}
	}
	return MenaFrontmatter{}
}

// ReadMenaFrontmatterFromFile reads a standalone mena file and parses its
// YAML frontmatter. Returns a zero-value MenaFrontmatter if no frontmatter
// is present or parsing fails.
func ReadMenaFrontmatterFromFile(filePath string) MenaFrontmatter {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return MenaFrontmatter{}
	}
	return parseMenaFrontmatterBytes(data)
}

// parseMenaFrontmatterBytes extracts YAML frontmatter from raw file bytes.
// Returns a zero-value MenaFrontmatter if no frontmatter delimiters are found
// or if YAML parsing fails. Parse failures are silent (the entry is treated
// as unscoped per EC-7 in the PRD).
func parseMenaFrontmatterBytes(data []byte) MenaFrontmatter {
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		return MenaFrontmatter{}
	}

	// Find closing delimiter
	var endIndex int
	searchStart := 4
	if idx := bytes.Index(data[searchStart:], []byte("\n---\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\r\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(data[searchStart:], []byte("\r\n---\n")); idx != -1 {
		endIndex = idx
	} else {
		return MenaFrontmatter{}
	}

	var fm MenaFrontmatter
	if err := yaml.Unmarshal(data[searchStart:searchStart+endIndex], &fm); err != nil {
		// EC-7: malformed YAML -- treat as unscoped (include in both pipelines)
		return MenaFrontmatter{}
	}
	return fm
}
```

Note: This requires adding `"bytes"` to the imports in `project_mena.go`. The `yaml.v3` import is already available via the parent package; ensure it is imported in `project_mena.go` as well (`"gopkg.in/yaml.v3"`).

**Design note**: The `parseMenaFrontmatterBytes` function reuses the same frontmatter extraction pattern found in `frontmatter_test.go` line 108-126. This is intentional -- it matches the established pattern rather than introducing a new parsing library. The function returns a zero-value struct on any failure, which means `Scope` defaults to `MenaScopeBoth` (empty string), satisfying EC-3 and EC-7.

## 5. Scope Filtering in ProjectMena() -- Leaf Directories

**File**: `internal/materialize/project_mena.go`

In the Pass 2 loop (line 173), after the existing type filter check (line 198-203) and before the destination directory assignment (line 205), insert the scope filter:

```go
// Pass 2: Route each collected leaf directory by filename convention.
for name, ce := range collected {
	menaType := "dro" // default: route to commands/

	if ce.source.IsEmbedded {
		// ... existing INDEX file detection for menaType ...
	} else {
		// ... existing INDEX file detection for menaType ...
	}

	// Apply type filter (existing)
	if menaType == "dro" && opts.Filter&ProjectDro == 0 {
		continue
	}
	if menaType == "lego" && opts.Filter&ProjectLego == 0 {
		continue
	}

	// --- NEW: Apply scope filter ---
	if opts.PipelineScope != MenaScopeBoth {
		var fm MenaFrontmatter
		if ce.source.IsEmbedded {
			fm = readMenaFrontmatterFromFS(ce.source.Fsys, ce.source.FsysPath)
		} else {
			fm = ReadMenaFrontmatterFromDir(ce.source.Path)
		}
		if !scopeIncludesPipeline(fm.Scope, opts.PipelineScope) {
			continue
		}
	}
	// --- END scope filter ---

	var destDir string
	// ... rest of existing routing logic ...
}
```

**Key point**: The scope filter is guarded by `opts.PipelineScope != MenaScopeBoth`. When `PipelineScope` is the zero value, the entire scope check is skipped -- no frontmatter parsing, no I/O. This preserves performance for callers that do not opt into scope filtering and satisfies EC-8.

## 6. Scope Filtering for Standalone Files

**File**: `internal/materialize/project_mena.go`

In the standalone files loop (line 236), after the existing type filter and before destination path computation, insert the scope filter:

```go
// Copy standalone files
for _, sf := range standalones {
	menaType := DetectMenaType(filepath.Base(sf.srcPath))

	// Apply type filter (existing)
	if menaType == "dro" && opts.Filter&ProjectDro == 0 {
		continue
	}
	if menaType == "lego" && opts.Filter&ProjectLego == 0 {
		continue
	}

	// --- NEW: Apply scope filter ---
	if opts.PipelineScope != MenaScopeBoth {
		fm := ReadMenaFrontmatterFromFile(sf.srcPath)
		if !scopeIncludesPipeline(fm.Scope, opts.PipelineScope) {
			continue
		}
	}
	// --- END scope filter ---

	var baseDir string
	// ... rest of existing standalone routing logic ...
}
```

## 7. materializeMena() Wire-up

**File**: `internal/materialize/materialize.go`

In `materializeMena()` (line 610-618), set `PipelineScope` on the options struct:

```go
// Delegate to ProjectMena with destructive mode
opts := MenaProjectionOptions{
	Mode:              MenaProjectionDestructive,
	Filter:            ProjectAll,
	PipelineScope:     MenaScopeProject, // NEW: filter out scope:user entries
	TargetCommandsDir: commandsDir,
	TargetSkillsDir:   skillsDir,
}

_, err := ProjectMena(sources, opts)
return err
```

This is a one-line change. After this change, any mena entry with `scope: user` in its INDEX frontmatter will be excluded from project-level materialization.

## 8. Usersync Wire-up

### Decision: Approach B (Inline Frontmatter Parsing)

**Chosen**: Approach B -- add frontmatter parsing directly in `syncFiles()`.

**Rejected**: Approach A (refactor usersync to call `ProjectMena()`).

**Rationale**:

Approach A would require significant refactoring of `syncFiles()` to replace its `filepath.WalkDir` with `ProjectMena()` calls. The usersync pipeline has fundamentally different semantics from materialize:

1. **Manifest tracking**: Usersync maintains a per-file manifest with checksums, source types (knossos/diverged/user), and collision detection. `ProjectMena()` has no manifest concept -- it simply copies files.

2. **Additive mode with per-file decisions**: Usersync makes add/update/skip decisions per file based on manifest state. `ProjectMena()` in additive mode copies everything, overwriting existing files without checking divergence.

3. **Different responsibility boundary**: Usersync owns the file lifecycle (install, track, detect divergence, recover). `ProjectMena()` owns the collection-and-routing concern. Scope filtering is a collection concern, but usersync needs to apply it at the walk level, before its per-file lifecycle logic runs.

Approach B adds approximately 15 lines of code to `syncFiles()` and requires no structural changes. The frontmatter parsing helpers from Section 4 are reused, so there is no code duplication of the parsing logic itself.

### Implementation

**File**: `internal/usersync/usersync.go`

In `syncFiles()` (line 287), add scope checking for mena resources. The check goes after the `relPath` computation (line 299) and before the manifest key computation (line 305):

```go
func (s *Syncer) syncFiles(manifest *Manifest, result *Result, opts Options) error {
	return filepath.WalkDir(s.sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.sourceDir, path)
		if err != nil {
			return err
		}

		// --- NEW: Scope filtering for mena resources ---
		if s.resourceType == ResourceMena {
			if scope := s.readMenaScope(path, relPath); scope == materialize.MenaScopeProject {
				// Entry is project-only -- skip in usersync pipeline
				return nil
			}
		}
		// --- END scope filtering ---

		// For flat resources, use just the filename
		manifestKey := relPath
		if !s.nested {
			manifestKey = filepath.Base(relPath)
		}

		// ... rest of existing syncFiles logic ...
	})
}
```

Add a helper method on `Syncer`:

```go
// readMenaScope determines the scope of a mena source file.
// For INDEX files, reads frontmatter from the file itself.
// For companion files (non-INDEX files in a leaf directory), reads
// frontmatter from the sibling INDEX file (scope is directory-level per EC-4).
// Returns MenaScopeBoth if no scope is set or on any parse failure.
func (s *Syncer) readMenaScope(absPath, relPath string) materialize.MenaScope {
	baseName := filepath.Base(relPath)

	if strings.HasPrefix(baseName, "INDEX") {
		// This IS the INDEX file -- parse its own frontmatter
		fm := materialize.ReadMenaFrontmatterFromFile(absPath)
		return fm.Scope
	}

	// Companion file -- find sibling INDEX file in the same directory
	dir := filepath.Dir(absPath)
	fm := materialize.ReadMenaFrontmatterFromDir(dir)
	return fm.Scope
}
```

**Export requirement**: For usersync to call the frontmatter parsing helpers, `readMenaFrontmatterFromDir` and `readMenaFrontmatterFromFile` must be exported. Rename them to `ReadMenaFrontmatterFromDir` and `ReadMenaFrontmatterFromFile` in `project_mena.go`. The `readMenaFrontmatterFromFS` helper remains unexported since it is only used within the materialize package by `ProjectMena()`. Similarly, `parseMenaFrontmatterBytes` remains unexported.

Updated signatures in `project_mena.go`:

```go
func ReadMenaFrontmatterFromDir(dirPath string) MenaFrontmatter { ... }
func ReadMenaFrontmatterFromFile(filePath string) MenaFrontmatter { ... }
func readMenaFrontmatterFromFS(fsys fs.FS, dirPath string) MenaFrontmatter { ... }
func parseMenaFrontmatterBytes(data []byte) MenaFrontmatter { ... }
```

Within `ProjectMena()`, the leaf directory scope check (Section 5) calls the unexported variant for filesystem sources and the FS variant for embedded sources, so no cross-package export is needed for those code paths.

**Standalone files in usersync**: Standalone mena files (not in a leaf INDEX directory) are walked as individual files. The `readMenaScope` helper handles this case: since the file basename does not start with "INDEX", it looks for a sibling INDEX file. If no sibling INDEX exists (true standalone), `ReadMenaFrontmatterFromDir` returns a zero-value struct with `Scope: ""`, meaning "both pipelines." This is correct -- standalone files without frontmatter are unscoped. If the standalone file itself has frontmatter, the caller would need to parse the file directly. However, since standalone files ARE walked individually by usersync, we should also check the file's own frontmatter. Update the helper:

```go
func (s *Syncer) readMenaScope(absPath, relPath string) materialize.MenaScope {
	baseName := filepath.Base(relPath)

	if strings.HasPrefix(baseName, "INDEX") {
		fm := materialize.ReadMenaFrontmatterFromFile(absPath)
		return fm.Scope
	}

	// Companion or standalone file -- check sibling INDEX first (EC-4: directory-level scope)
	dir := filepath.Dir(absPath)
	fm := materialize.ReadMenaFrontmatterFromDir(dir)
	if fm.Scope != materialize.MenaScopeBoth {
		return fm.Scope // Directory-level scope from INDEX takes precedence
	}

	// No INDEX-level scope -- check the file's own frontmatter (standalone files)
	fm = materialize.ReadMenaFrontmatterFromFile(absPath)
	return fm.Scope
}
```

**Import addition**: Add `"github.com/autom8y/knossos/internal/materialize"` to usersync.go imports. This import already exists (line 10 of current file).

## 9. EC-1 Resolution: Rite-Level Mena with scope: user

### Decision: Honor scope with warning (Option B)

**Chosen**: If a rite-level mena entry has `scope: user`, it is excluded from the materialize pipeline (the only pipeline that processes rite-level mena). This means the entry is effectively distributed to no pipeline.

**Rationale**:

Option A (ignore scope on rite-level mena) creates an inconsistency: the `scope` field means different things depending on where the file lives. For distribution-level mena, `scope: user` excludes it from materialize. For rite-level mena, `scope: user` would be silently ignored. This inconsistency would confuse authors and create a maintenance hazard -- moving a mena entry from rite-level to distribution-level would silently change its filtering behavior.

Option B (honor scope) is consistent: `scope: user` always means "exclude from the materialize pipeline." If a rite-level entry has `scope: user`, it goes nowhere, which is almost certainly an authoring mistake. The warning makes this visible.

**Implementation**: No special code needed in `ProjectMena()`. The scope filter in Section 5 applies uniformly to all sources, regardless of whether they are distribution-level or rite-level. The rite-level source is just another `MenaSource` in the priority-ordered list.

The warning is added in the scope filter block:

```go
// --- NEW: Apply scope filter ---
if opts.PipelineScope != MenaScopeBoth {
	var fm MenaFrontmatter
	if ce.source.IsEmbedded {
		fm = readMenaFrontmatterFromFS(ce.source.Fsys, ce.source.FsysPath)
	} else {
		fm = ReadMenaFrontmatterFromDir(ce.source.Path)
	}
	if !scopeIncludesPipeline(fm.Scope, opts.PipelineScope) {
		// EC-1: warn if rite-level mena has scope:user (goes nowhere)
		if fm.Scope == MenaScopeUser && opts.PipelineScope == MenaScopeProject {
			// Log warning: this is likely an authoring mistake
			// Use fmt.Fprintf(os.Stderr, ...) for now; a logger can be added later
			fmt.Fprintf(os.Stderr, "warning: mena %q has scope: user but is only reachable by materialize (will not be distributed)\n", name)
		}
		continue
	}
}
// --- END scope filter ---
```

**Why stderr, not a logger**: The materialize package currently uses no structured logger. Adding one for a single warning would over-engineer a low-complexity change. The principal-engineer may substitute `slog` or another logger if one is introduced in the package before implementation.

## 10. Test Plan

### 10.1 Unit Tests: Frontmatter Validation

**File**: `internal/materialize/frontmatter_test.go`

```
TestMenaScope_ValidScope
  - MenaScopeBoth ("") returns true
  - MenaScopeUser ("user") returns true
  - MenaScopeProject ("project") returns true
  - "global" returns false
  - "both" returns false
  - "User" returns false (case-sensitive)
  - "PROJECT" returns false

TestMenaScope_String
  - MenaScopeBoth returns "both"
  - MenaScopeUser returns "user"
  - MenaScopeProject returns "project"

TestMenaFrontmatter_Validate_Scope
  - Valid: scope omitted (empty) -- passes
  - Valid: scope "user" -- passes
  - Valid: scope "project" -- passes
  - Invalid: scope "global" -- error contains "invalid scope"
  - Invalid: scope "both" -- error (explicit "both" is not valid, omit instead)
  - Invalid: scope "User" -- error (case-sensitive)
  - Error message includes the invalid value and valid options
```

### 10.2 Unit Tests: Frontmatter Parsing Helpers

**File**: `internal/materialize/project_mena_test.go`

```
TestParseMenaFrontmatterBytes
  - File with scope:user returns fm.Scope == MenaScopeUser
  - File with scope:project returns fm.Scope == MenaScopeProject
  - File with no scope field returns fm.Scope == MenaScopeBoth
  - File with no frontmatter (no --- delimiters) returns zero-value
  - File with malformed YAML returns zero-value (EC-7)
  - File with empty content returns zero-value

TestReadMenaFrontmatterFromDir (exported)
  - Directory with INDEX.dro.md containing scope:user returns correct scope
  - Directory with INDEX.lego.md containing scope:project returns correct scope
  - Directory with INDEX file missing frontmatter returns MenaScopeBoth
  - Nonexistent directory returns MenaScopeBoth

TestReadMenaFrontmatterFromFile (exported)
  - Standalone file with scope:user returns correct scope
  - Standalone file with no frontmatter returns MenaScopeBoth
```

### 10.3 Unit Tests: Scope Filtering in ProjectMena()

**File**: `internal/materialize/project_mena_test.go`

```
TestScopeIncludesPipeline
  - All 9 combinations from the truth table in Section 3

TestProjectMena_ScopeUser_ExcludedFromProject
  - Create mena source with INDEX.dro.md containing scope:user
  - Call ProjectMena with PipelineScope: MenaScopeProject
  - Verify entry is NOT in output commands/

TestProjectMena_ScopeProject_ExcludedFromUser
  - Create mena source with INDEX.dro.md containing scope:project
  - Call ProjectMena with PipelineScope: MenaScopeUser
  - Verify entry is NOT in output commands/

TestProjectMena_ScopeUser_IncludedInUser
  - Create mena source with INDEX.dro.md containing scope:user
  - Call ProjectMena with PipelineScope: MenaScopeUser
  - Verify entry IS in output commands/

TestProjectMena_ScopeProject_IncludedInProject
  - Create mena source with INDEX.dro.md containing scope:project
  - Call ProjectMena with PipelineScope: MenaScopeProject
  - Verify entry IS in output commands/

TestProjectMena_NoScope_IncludedInBoth
  - Create mena source with INDEX.dro.md with no scope field
  - Call ProjectMena with PipelineScope: MenaScopeProject -- included
  - Call ProjectMena with PipelineScope: MenaScopeUser -- included

TestProjectMena_NoPipelineScope_NoFiltering
  - Create mena source with scope:user entry
  - Call ProjectMena with PipelineScope: MenaScopeBoth (zero value)
  - Verify entry IS included (EC-8: no filtering when pipeline scope not set)

TestProjectMena_StandaloneFile_ScopeFiltered
  - Create standalone mena file with scope:user
  - Call ProjectMena with PipelineScope: MenaScopeProject
  - Verify standalone file is NOT in output

TestProjectMena_ScopeWithEmbeddedFS
  - Create fstest.MapFS with INDEX containing scope:user
  - Call ProjectMena with embedded source and PipelineScope: MenaScopeProject
  - Verify entry is excluded

TestProjectMena_MixedScopes
  - Create source with 3 entries: scope:user, scope:project, no scope
  - Call ProjectMena with PipelineScope: MenaScopeProject
  - Verify: scope:user excluded, scope:project included, no-scope included

TestProjectMena_BackwardCompat_ExistingTests
  - All existing TestProjectMena_* tests continue to pass without modification
  - (No changes needed to existing tests -- they use default PipelineScope)
```

### 10.4 Unit Tests: Usersync Scope Filtering

**File**: `internal/usersync/usersync_test.go`

```
TestMenaSyncer_ScopeProject_ExcludedFromUsersync
  - Create mena source dir with INDEX.dro.md containing scope:project
  - Create MenaSyncerWithPaths and Sync
  - Verify: file is NOT synced to target commands dir
  - Verify: manifest does NOT contain the entry

TestMenaSyncer_ScopeUser_IncludedInUsersync
  - Create mena source dir with INDEX.dro.md containing scope:user
  - Create MenaSyncerWithPaths and Sync
  - Verify: file IS synced to target commands dir

TestMenaSyncer_NoScope_IncludedInUsersync
  - Create mena source dir with INDEX.dro.md with no scope field
  - Sync and verify: file IS synced (backward compat)

TestMenaSyncer_CompanionFile_InheritsIndexScope
  - Create leaf dir with INDEX.dro.md (scope:project) and helper.md
  - Sync with usersync
  - Verify: both INDEX and helper are excluded (EC-4)

TestMenaSyncer_StandaloneFile_OwnScope
  - Create standalone mena file (not in INDEX dir) with scope:project
  - Sync with usersync
  - Verify: file is excluded
```

### 10.5 Integration Test

```
TestMenaScopeEndToEnd
  - Set up temp dirs with distribution-level mena (3 entries: user, project, both)
  - Run ProjectMena with PipelineScope: MenaScopeProject
  - Verify: project + both entries present, user entry absent
  - Run ProjectMena with PipelineScope: MenaScopeUser
  - Verify: user + both entries present, project entry absent
  - Verify all existing entries without scope remain in both
```

### 10.6 Test Execution

```bash
CGO_ENABLED=0 go test ./internal/materialize/ -run TestMenaScope
CGO_ENABLED=0 go test ./internal/materialize/ -run TestProjectMena
CGO_ENABLED=0 go test ./internal/materialize/ -run TestParseMenaFrontmatter
CGO_ENABLED=0 go test ./internal/materialize/ -run TestReadMenaFrontmatter
CGO_ENABLED=0 go test ./internal/usersync/ -run TestMenaSyncer_Scope
CGO_ENABLED=0 go test ./internal/materialize/ ./internal/usersync/
CGO_ENABLED=0 go test ./...  # Full suite -- verify no regressions
```

## File Change Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/materialize/frontmatter.go` | Modify | Add `MenaScope` type, constants, `ValidScope()`, `String()`. Add `Scope` field to `MenaFrontmatter`. Extend `Validate()`. Add `"fmt"` import. |
| `internal/materialize/project_mena.go` | Modify | Add `PipelineScope` to `MenaProjectionOptions`. Add `scopeIncludesPipeline()` (unexported). Add `ReadMenaFrontmatterFromDir()` and `ReadMenaFrontmatterFromFile()` (exported, used by usersync). Add `readMenaFrontmatterFromFS()` and `parseMenaFrontmatterBytes()` (unexported). Add scope filter in Pass 2 loop and standalone files loop. Add `"bytes"` and `"gopkg.in/yaml.v3"` imports. |
| `internal/materialize/materialize.go` | Modify | Set `PipelineScope: MenaScopeProject` in `materializeMena()`. One-line change. |
| `internal/usersync/usersync.go` | Modify | Add `readMenaScope()` method on `Syncer`. Add scope check in `syncFiles()`. Add `"strings"` import (already present). |
| `internal/materialize/frontmatter_test.go` | Modify | Add `TestMenaScope_ValidScope`, `TestMenaScope_String`, `TestMenaFrontmatter_Validate_Scope`. |
| `internal/materialize/project_mena_test.go` | Modify | Add scope filtering tests per Section 10.3. |
| `internal/usersync/usersync_test.go` | Modify | Add scope filtering tests per Section 10.4. |
| `docs/decisions/ADR-0025-mena-scope.md` | Create | ADR documenting the scope decision. |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Frontmatter parsing adds latency | Low | Low | Guarded by `PipelineScope != MenaScopeBoth` -- no parsing when scope not requested. 55 files with YAML frontmatter parsing takes <10ms. |
| Existing tests break | Very Low | Medium | `PipelineScope` zero value means no filtering. All existing callers omit the field, so behavior is unchanged. |
| Invalid scope silently accepted | Very Low | Low | `Validate()` rejects unknown values. `parseMenaFrontmatterBytes` returns zero-value on failure, defaulting to "both" (safe fallback). |
| Usersync manifest keys disrupted | Very Low | High | Scope filtering happens before manifest key computation. Filtered entries never reach the manifest. Existing manifest entries are unaffected. |

## Performance Analysis

Frontmatter parsing reads the first ~200 bytes of each INDEX file to extract YAML between `---` delimiters. For 55 mena entries (25 distribution + 30 rite-level), this is approximately 55 small file reads. On modern SSDs with OS page cache (these files are always recently read during collection), this adds <5ms of wall-clock time. Well within the 50ms NFR-2 budget.

The `PipelineScope != MenaScopeBoth` guard ensures zero additional I/O when scope filtering is not requested.
