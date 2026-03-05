# TDD: Provenance Manifest

> Technical Design Document for the unified file-level provenance system per ADR-0026.

**Status**: Draft
**Author**: Context Architect
**Date**: 2026-02-09
**ADR**: docs/decisions/ADR-0026-unified-provenance.md
**Scope**: Phase 2 (manifest + pipeline integration) + Phase 3 (orphan detection + CLI + divergence warnings)

---

## Decision Register

Decisions locked from stakeholder interview on 2026-02-09.

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Package location | `internal/provenance/` (new) | Dedicated package avoids circular imports with materialize; both materialize and future CLI commands import it |
| 2 | On-disk format | YAML | Matches KNOSSOS_MANIFEST.yaml convention; human-readable for debugging |
| 3 | Mena granularity | Directory-level keys for mena, file-level for everything else | Mena entries are atomic directories (INDEX + support files); tracking individual mena files creates false divergence on re-projection |
| 4 | Divergence detection | Always-on (every sync) | No config toggle; divergence check runs before every pipeline write pass |
| 5 | Sprint scope | Phase 2 + Phase 3 | Implement manifest and pipeline integration, then orphan detection and CLI |
| 6 | CLI shape | `ari provenance show` (top-level command group) | Matches `ari session show`, `ari rite show` pattern |
| 7 | Rite switch behavior | Full rebuild | Manifest is rebuilt from scratch on each materialization; carried-forward entries are merged from previous manifest |
| 8 | Bootstrap unknown files | owner:user | Pre-existing files not written by the pipeline default to user-owned for safety |
| 9 | Git tracking | Committed | PROVENANCE_MANIFEST.yaml is committed alongside .claude/ contents |
| 10 | KNOSSOS_MANIFEST migration | Keep separate | Two manifests serve different concerns (file-level vs region-level); no merge |

---

## Section 1: Go Structs

Package: `internal/provenance/`

### ProvenanceManifest

```go
package provenance

import "time"

// ManifestFileName is the provenance manifest filename within .claude/.
const ManifestFileName = "PROVENANCE_MANIFEST.yaml"

// SchemaVersion is the current manifest schema version.
// Starts at "1.0" for the provenance manifest (independent of inscription's "1.0").
const SchemaVersion = "1.0"

// ProvenanceManifest is the unified file-level provenance tracker for .claude/.
// Stored at .knossos/PROVENANCE_MANIFEST.yaml.
type ProvenanceManifest struct {
    // SchemaVersion is the manifest format version. Currently "1.0".
    SchemaVersion string `yaml:"schema_version"`

    // LastSync is the UTC timestamp of the most recent materialization.
    LastSync time.Time `yaml:"last_sync"`

    // ActiveRite is the rite name that produced this manifest.
    // Empty string for minimal (cross-cutting) materializations.
    ActiveRite string `yaml:"active_rite,omitempty"`

    // Entries maps relative paths within .claude/ to their provenance records.
    // Keys use forward slashes. Directory entries end with "/" (mena only).
    // Examples: "agents/orchestrator.md", "commands/commit/", "CLAUDE.md"
    Entries map[string]*ProvenanceEntry `yaml:"entries"`
}
```

### ProvenanceEntry

```go
// ProvenanceEntry tracks the origin and state of a single file or directory in .claude/.
type ProvenanceEntry struct {
    // Owner determines sync behavior for this entry.
    Owner OwnerType `yaml:"owner"`

    // SourcePipeline identifies which pipeline placed this file.
    // "materialize" for project-level, empty string for user-created files.
    SourcePipeline string `yaml:"source_pipeline,omitempty"`

    // SourcePath is the relative path (from project root) to the source file.
    // Empty for user-created files.
    // Examples: "rites/ecosystem/agents/orchestrator.md",
    //           "mena/operations/commit/INDEX.dro.md",
    //           "knossos/templates/rules/internal-hook.md"
    SourcePath string `yaml:"source_path,omitempty"`

    // SourceType records which tier of the source resolution chain provided the file.
    // Values match materialize.SourceType: "project", "user", "knossos", "explicit", "embedded".
    // Additional values for mena provenance: "template", "shared", "dependency".
    SourceType string `yaml:"source_type,omitempty"`

    // Checksum is the SHA256 hash of the file (or directory for mena) at write time.
    // Uses the "sha256:" prefix per ADR-0026 and internal/checksum convention.
    Checksum string `yaml:"checksum"`

    // LastSynced is the UTC timestamp when this entry was last written by the pipeline.
    LastSynced time.Time `yaml:"last_synced"`
}
```

### Validation Rules

| Field | Required | Validation |
|-------|----------|------------|
| `SchemaVersion` | Yes | Must match `^[0-9]+\.[0-9]+$` |
| `LastSync` | Yes | Must be non-zero time |
| `ActiveRite` | No | Empty string permitted (minimal mode) |
| `Entries` | Yes | Must be non-nil map (may be empty) |
| `Entry.Owner` | Yes | Must be one of: `knossos`, `user`, `unknown` |
| `Entry.SourcePipeline` | No | If present, must be `"materialize"` |
| `Entry.SourcePath` | No | If Owner is `knossos`, must be non-empty |
| `Entry.SourceType` | No | If Owner is `knossos`, must be non-empty |
| `Entry.Checksum` | Yes | Must match `^sha256:[0-9a-f]{64}$` |
| `Entry.LastSynced` | Yes | Must be non-zero time |

---

## Section 2: OwnerType Enum

```go
// OwnerType represents who owns a file in .claude/.
type OwnerType string

const (
    // OwnerKnossos indicates files managed by Knossos.
    // These are safe to overwrite on sync.
    OwnerKnossos OwnerType = "knossos"

    // OwnerUser indicates files created or modified by the user.
    // These are NEVER overwritten by the pipeline.
    OwnerUser OwnerType = "user"

    // OwnerUnknown indicates pre-existing files discovered during bootstrap.
    // Treated as user-owned for safety. Promoted to OwnerUser or OwnerKnossos
    // on the next sync that interacts with the file.
    OwnerUnknown OwnerType = "unknown"
)

// IsValid returns true if the owner type is a recognized value.
func (o OwnerType) IsValid() bool {
    switch o {
    case OwnerKnossos, OwnerUser, OwnerUnknown:
        return true
    default:
        return false
    }
}

// String returns the string representation.
func (o OwnerType) String() string {
    return string(o)
}
```

### Promotion Rules

State transitions for the Owner field across sync operations:

| Current Owner | Condition | New Owner | Rationale |
|---------------|-----------|-----------|-----------|
| `knossos` | Checksum matches on-disk | `knossos` | File unchanged; retain ownership |
| `knossos` | Checksum mismatch (user edited) | `user` | User modified a knossos file; promote to protect edits |
| `user` | Any condition | `user` | User files are never demoted; ownership is permanent |
| `unknown` | Pipeline writes the file this sync | `knossos` | File is now managed by knossos |
| `unknown` | Pipeline does NOT write the file | `user` | Default to user for safety |
| `unknown` | Any user interaction | `user` | Any interaction confirms user ownership |

Key invariant: once a file reaches `owner: user`, it stays `owner: user` forever. The only way to reclaim a user-promoted file for knossos management is manual deletion of the file (knossos recreates it as owner:knossos on next sync) or manual editing of the manifest.

---

## Section 3: SourceType Values

The `SourceType` field in ProvenanceEntry records where the file content originated. Values are a superset of `materialize.SourceType` from `internal/materialize/source.go`.

### Base Types (from source.go)

| Value | Meaning | Example |
|-------|---------|---------|
| `"project"` | From the project's `rites/` directory | `rites/ecosystem/agents/orchestrator.md` |
| `"user"` | From user-level `~/.local/share/knossos/rites/` | User-installed rite agents |
| `"knossos"` | From `$KNOSSOS_HOME/rites/` | Platform-distributed rite agents |
| `"explicit"` | From `--source` flag explicit path | Explicitly specified rite agents |
| `"embedded"` | Compiled into the ari binary | Embedded rite agents |

### Extended Types (provenance-specific)

| Value | Meaning | Example |
|-------|---------|---------|
| `"template"` | From `knossos/templates/` directory | `knossos/templates/rules/internal-hook.md` |
| `"shared"` | From `rites/shared/mena/` | Shared rite mena entries |
| `"dependency"` | From a dependency rite's `mena/` | `rites/base/mena/commit/` |

### Assignment Logic

The SourceType is determined by the pipeline stage that writes the file:

- `materializeAgents()`: Uses `resolved.Source.Type` (project/user/knossos/explicit/embedded)
- `materializeMena()`: Uses "shared" for `rites/shared/mena/`, "dependency" for dependency rites, the resolved source type for the current rite, or "project" for top-level `mena/` directory
- `materializeHooks()`: Uses "template" (hooks come from `knossos/templates/hooks/`)
- `materializeRules()`: Uses "template" (rules come from `knossos/templates/rules/`)
- `mergeCLAUDEmd()`: Uses "template" (CLAUDE.md is generated from templates)
- `materializeSettingsWithManifest()`: Uses "template" (settings are generated)
- `materializeWorkflow()`: Uses `resolved.Source.Type` (workflow.yaml comes from the rite)
- `writeActiveRite()`: Uses "template" (ACTIVE_RITE is a pipeline-generated marker)
- `trackState()`: Uses "template" (sync/state.json is a pipeline-generated state file)

---

## Section 4: On-Disk YAML Schema

File path: `.knossos/PROVENANCE_MANIFEST.yaml`

```yaml
schema_version: "1.0"
last_sync: "2026-02-09T14:30:00Z"
active_rite: ecosystem
entries:
  # --- Agents (file-level) ---
  agents/orchestrator.md:
    owner: knossos
    source_pipeline: materialize
    source_path: rites/ecosystem/agents/orchestrator.md
    source_type: project
    checksum: "sha256:a1b2c3d4e5f6..."
    last_synced: "2026-02-09T14:30:00Z"
  agents/ecosystem-analyst.md:
    owner: knossos
    source_pipeline: materialize
    source_path: rites/ecosystem/agents/ecosystem-analyst.md
    source_type: project
    checksum: "sha256:b2c3d4e5f6a1..."
    last_synced: "2026-02-09T14:30:00Z"
  agents/my-custom-agent.md:
    owner: user
    checksum: "sha256:c3d4e5f6a1b2..."
    last_synced: "2026-02-08T10:00:00Z"

  # --- Mena commands (directory-level, key ends with /) ---
  commands/commit/:
    owner: knossos
    source_pipeline: materialize
    source_path: mena/operations/commit/
    source_type: project
    checksum: "sha256:d4e5f6a1b2c3..."
    last_synced: "2026-02-09T14:30:00Z"
  commands/consult/:
    owner: knossos
    source_pipeline: materialize
    source_path: rites/shared/mena/consult/
    source_type: shared
    checksum: "sha256:e5f6a1b2c3d4..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- Mena skills (directory-level, key ends with /) ---
  skills/prompting/:
    owner: knossos
    source_pipeline: materialize
    source_path: mena/reference/prompting/
    source_type: project
    checksum: "sha256:f6a1b2c3d4e5..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- Hooks (file-level) ---
  hooks/hooks.yaml:
    owner: knossos
    source_pipeline: materialize
    source_path: knossos/templates/hooks/hooks.yaml
    source_type: template
    checksum: "sha256:1a2b3c4d5e6f..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- Rules (file-level) ---
  rules/internal-hook.md:
    owner: knossos
    source_pipeline: materialize
    source_path: knossos/templates/rules/internal-hook.md
    source_type: template
    checksum: "sha256:2b3c4d5e6f1a..."
    last_synced: "2026-02-09T14:30:00Z"
  rules/internal-materialize.md:
    owner: knossos
    source_pipeline: materialize
    source_path: knossos/templates/rules/internal-materialize.md
    source_type: template
    checksum: "sha256:3c4d5e6f1a2b..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- CLAUDE.md (file-level) ---
  CLAUDE.md:
    owner: knossos
    source_pipeline: materialize
    source_path: knossos/templates/CLAUDE.md.tpl
    source_type: template
    checksum: "sha256:4d5e6f1a2b3c..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- Settings (file-level) ---
  settings.local.json:
    owner: knossos
    source_pipeline: materialize
    source_path: (generated)
    source_type: template
    checksum: "sha256:5e6f1a2b3c4d..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- Workflow (file-level) ---
  ACTIVE_WORKFLOW.yaml:
    owner: knossos
    source_pipeline: materialize
    source_path: rites/ecosystem/workflow.yaml
    source_type: project
    checksum: "sha256:6f1a2b3c4d5e..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- ACTIVE_RITE marker (file-level) ---
  ACTIVE_RITE:
    owner: knossos
    source_pipeline: materialize
    source_path: (generated)
    source_type: template
    checksum: "sha256:7a1b2c3d4e5f..."
    last_synced: "2026-02-09T14:30:00Z"

  # --- Sync state (file-level) ---
  sync/state.json:
    owner: knossos
    source_pipeline: materialize
    source_path: (generated)
    source_type: template
    checksum: "sha256:8b2c3d4e5f6a..."
    last_synced: "2026-02-09T14:30:00Z"
```

### Key Naming Conventions

| Resource Type | Key Pattern | Examples |
|---------------|-------------|---------|
| Agents | `agents/{name}.md` | `agents/orchestrator.md` |
| Mena commands | `commands/{name}/` | `commands/commit/`, `commands/code-review/` |
| Mena skills | `skills/{name}/` | `skills/prompting/`, `skills/lexicon/` |
| Hooks | `hooks/{filename}` | `hooks/hooks.yaml` |
| Rules | `rules/{filename}` | `rules/internal-hook.md` |
| CLAUDE.md | `CLAUDE.md` | `CLAUDE.md` |
| Settings | `settings.local.json` | `settings.local.json` |
| Workflow | `ACTIVE_WORKFLOW.yaml` | `ACTIVE_WORKFLOW.yaml` |
| Rite marker | `ACTIVE_RITE` | `ACTIVE_RITE` |
| Sync state | `sync/state.json` | `sync/state.json` |

Directory-level keys (mena) always end with `/`. File-level keys never end with `/`. This convention is enforced during recording and enables reliable discrimination.

---

## Section 5: Collector Interface

The Collector is the pipeline's interface for recording provenance entries during materialization. Each pipeline stage calls `Record()` after successfully writing a file.

### Interface Definition

```go
package provenance

// Collector accumulates provenance entries during a materialization run.
// Pipeline stages call Record() after each successful file write.
// At the end of materialization, the orchestrating function calls Entries()
// to retrieve all recorded entries for manifest construction.
type Collector interface {
    // Record adds or updates a provenance entry for the given relative path.
    // relativePath is relative to .claude/ (e.g., "agents/orchestrator.md").
    // Duplicate paths overwrite previous entries (last-writer-wins, matching
    // the materialize pipeline's priority semantics).
    Record(relativePath string, entry *ProvenanceEntry)

    // Entries returns all recorded entries. The returned map must not be
    // modified by the caller.
    Entries() map[string]*ProvenanceEntry
}
```

### Default Implementation

```go
// defaultCollector is the in-memory implementation of Collector.
type defaultCollector struct {
    entries map[string]*ProvenanceEntry
}

// NewCollector creates a new in-memory Collector.
func NewCollector() Collector {
    return &defaultCollector{
        entries: make(map[string]*ProvenanceEntry),
    }
}

// Record adds or overwrites a provenance entry for the given path.
func (c *defaultCollector) Record(relativePath string, entry *ProvenanceEntry) {
    c.entries[relativePath] = entry
}

// Entries returns the accumulated entries map.
func (c *defaultCollector) Entries() map[string]*ProvenanceEntry {
    return c.entries
}
```

### Integration Pattern

The Collector is created at the start of `MaterializeWithOptions()` and threaded through each pipeline stage. Each stage records its writes:

```
MaterializeWithOptions()
    collector := provenance.NewCollector()
    |
    +-- materializeAgents(... collector)
    |     Record("agents/orchestrator.md", ...)
    |     Record("agents/ecosystem-analyst.md", ...)
    |
    +-- materializeMena(... collector)
    |     Record("commands/commit/", ...)
    |     Record("skills/prompting/", ...)
    |
    +-- materializeHooks(... collector)
    |     Record("hooks/hooks.yaml", ...)
    |
    +-- materializeRules(... collector)
    |     Record("rules/internal-hook.md", ...)
    |
    +-- mergeCLAUDEmd(... collector)
    |     Record("CLAUDE.md", ...)
    |
    +-- materializeSettingsWithManifest(... collector)
    |     Record("settings.local.json", ...)
    |
    +-- materializeWorkflow(... collector)
    |     Record("ACTIVE_WORKFLOW.yaml", ...)
    |
    +-- writeActiveRite(... collector)
    |     Record("ACTIVE_RITE", ...)
    |
    +-- trackState(... collector)
    |     Record("sync/state.json", ...)
    |
    +-- buildAndSaveManifest(collector, previousManifest)
          merge collector.Entries() with previous manifest
          write PROVENANCE_MANIFEST.yaml
```

### NullCollector for Minimal/DryRun Mode

```go
// NullCollector is a no-op Collector for dry-run and minimal modes
// where provenance tracking is not needed.
type NullCollector struct{}

func (NullCollector) Record(string, *ProvenanceEntry) {}
func (NullCollector) Entries() map[string]*ProvenanceEntry {
    return nil
}
```

---

## Section 6: Divergence Detection Algorithm

Divergence detection runs at the start of every materialization, before any pipeline writes. It detects user modifications to knossos-owned files and promotes them to user-owned.

### Pseudocode

```
function detectDivergence(claudeDir string, previous *ProvenanceManifest) *ProvenanceManifest:
    if previous == nil:
        return nil  // No previous manifest; first sync, no divergence possible

    promoted := new ProvenanceManifest (clone of previous)

    for path, entry in previous.Entries:
        if entry.Owner != OwnerKnossos:
            continue  // Only check knossos-owned files for divergence

        currentChecksum := computeCurrentChecksum(claudeDir, path)

        if currentChecksum == "":
            // File was deleted by user. Promote to user-owned so pipeline
            // does not recreate it (respects user intent to remove).
            promoted.Entries[path].Owner = OwnerUser
            promoted.Entries[path].Checksum = ""
            continue

        if currentChecksum != entry.Checksum:
            // User modified a knossos file. Promote to user-owned.
            promoted.Entries[path].Owner = OwnerUser
            promoted.Entries[path].Checksum = currentChecksum
            // SourcePipeline, SourcePath, SourceType retained for provenance history

    return promoted


function computeCurrentChecksum(claudeDir string, relativePath string) string:
    fullPath := filepath.Join(claudeDir, relativePath)

    if strings.HasSuffix(relativePath, "/"):
        // Directory-level entry (mena). Use checksum.Dir().
        dirPath := strings.TrimSuffix(fullPath, "/")
        hash, err := checksum.Dir(dirPath)
        if err != nil:
            return ""  // Directory missing or unreadable
        return hash
    else:
        // File-level entry. Use checksum.File().
        hash, err := checksum.File(fullPath)
        if err != nil:
            return ""  // File missing or unreadable
        return hash
```

### Merge Algorithm (End of Materialization)

After all pipeline stages have recorded entries via the Collector, the merge step combines:
1. Entries from the current sync (collector)
2. Promoted entries from divergence detection
3. Carried-forward user entries from the previous manifest

```
function buildFinalManifest(
    collector Collector,
    promoted *ProvenanceManifest,     // from divergence detection
    activeRite string,
) *ProvenanceManifest:

    final := new ProvenanceManifest
    final.SchemaVersion = SchemaVersion
    final.LastSync = time.Now().UTC()
    final.ActiveRite = activeRite
    final.Entries = make(map[string]*ProvenanceEntry)

    // Step 1: Carry forward all promoted entries (user-owned + unknown)
    // These are files from the previous manifest that divergence detection
    // identified as user-modified or that were already user/unknown-owned.
    if promoted != nil:
        for path, entry in promoted.Entries:
            if entry.Owner == OwnerUser || entry.Owner == OwnerUnknown:
                final.Entries[path] = entry

    // Step 2: Layer current sync entries on top.
    // Pipeline-written files take precedence, but ONLY if the path was not
    // promoted to user-owned in Step 1.
    for path, entry in collector.Entries():
        if existing, ok := final.Entries[path]; ok:
            if existing.Owner == OwnerUser:
                // User promoted this file via divergence detection.
                // Do NOT overwrite with the pipeline entry.
                // The pipeline wrote the file but the manifest records user ownership.
                continue
        final.Entries[path] = entry

    // Step 3: Resolve unknown entries from previous manifest.
    // Files in previous manifest with owner:unknown that the pipeline did NOT
    // write this sync are promoted to owner:user (Decision #8: bootstrap unknown = user).
    if promoted != nil:
        for path, entry in promoted.Entries:
            if entry.Owner == OwnerUnknown:
                if _, writtenThisSync := collector.Entries()[path]; !writtenThisSync:
                    entry.Owner = OwnerUser
                    final.Entries[path] = entry

    return final
```

### Important Edge Case: User-Promoted Files Still Get Written

When divergence detection promotes a knossos file to user-owned, the pipeline still writes the file (it does not know about the promotion at write time). The merge algorithm resolves this: the collector records the pipeline's entry, but the merge gives precedence to user-owned entries from the promoted manifest. This means:

- The file on disk is overwritten by the pipeline (same as today)
- The manifest records it as user-owned
- On the NEXT sync, divergence detection sees owner:user and skips it entirely

This is intentional for the Phase 2 implementation. Phase 3 adds divergence warnings to inform the user before the pipeline overwrites.

---

## Section 7: `ari provenance show` Output Format

### Table Format (default)

```
$ ari provenance show

Provenance Manifest (ecosystem rite, synced 2026-02-09T14:30:00Z)

PATH                            OWNER     SOURCE                                    STATUS
agents/orchestrator.md          knossos   rites/ecosystem/agents/orchestrator.md    match
agents/ecosystem-analyst.md     knossos   rites/ecosystem/agents/ecosystem-anal...  match
agents/my-custom-agent.md       user      (user-created)                            -
commands/commit/                knossos   mena/operations/commit/                   match
commands/consult/               knossos   rites/shared/mena/consult/                match
skills/prompting/               knossos   mena/reference/prompting/                 match
hooks/hooks.yaml                knossos   knossos/templates/hooks/hooks.yaml        match
rules/internal-hook.md          knossos   knossos/templates/rules/internal-hook.md  diverged
rules/internal-materialize.md   knossos   knossos/templates/rules/internal-mater... match
CLAUDE.md                       knossos   knossos/templates/CLAUDE.md.tpl           match
settings.local.json             knossos   (generated)                               match
ACTIVE_WORKFLOW.yaml            knossos   rites/ecosystem/workflow.yaml             match
ACTIVE_RITE                     knossos   (generated)                               match
sync/state.json                 knossos   (generated)                               match

14 entries (12 knossos, 1 user, 0 unknown)
1 diverged file (use 'ari provenance show --diverged' to list)
```

### Column Definitions

| Column | Source | Notes |
|--------|--------|-------|
| PATH | Entry map key | Sorted alphabetically; directories end with `/` |
| OWNER | `entry.Owner` | Colorized: knossos=cyan, user=green, unknown=yellow |
| SOURCE | `entry.SourcePath` | Truncated to 42 chars with `...`; user files show `(user-created)` |
| STATUS | Computed at display time | `match` if on-disk checksum equals manifest checksum; `diverged` if mismatch; `-` for user files; `missing` if file deleted |

### Filter Flags

| Flag | Effect |
|------|--------|
| `--owner knossos` | Show only knossos-owned entries |
| `--owner user` | Show only user-owned entries |
| `--diverged` | Show only entries where STATUS is `diverged` |
| `--json` | Output full manifest as JSON |
| `--yaml` | Output full manifest as YAML (raw file dump) |

### JSON Format

```json
{
  "schema_version": "1.0",
  "last_sync": "2026-02-09T14:30:00Z",
  "active_rite": "ecosystem",
  "entries": {
    "agents/orchestrator.md": {
      "owner": "knossos",
      "source_pipeline": "materialize",
      "source_path": "rites/ecosystem/agents/orchestrator.md",
      "source_type": "project",
      "checksum": "sha256:a1b2c3d4e5f6...",
      "last_synced": "2026-02-09T14:30:00Z",
      "status": "match"
    }
  },
  "summary": {
    "total": 14,
    "knossos": 12,
    "user": 1,
    "unknown": 0,
    "diverged": 1
  }
}
```

The `status` field in JSON output is computed at display time (not stored in the manifest) by comparing on-disk checksums.

### CLI Package Location

`cmd/ari/cmd/provenance.go` -- follows the existing Cobra command structure alongside `cmd/ari/cmd/session.go`, `cmd/ari/cmd/rite.go`.

---

## Section 8: Orphan Detection Migration Plan

### Current Implementation

`detectOrphans()` at `internal/materialize/materialize.go:391` identifies agent files not in the incoming rite manifest's agent list:

```go
func (m *Materializer) detectOrphans(manifest *RiteManifest, claudeDir string) ([]string, error) {
    // Build set of expected agents from manifest
    expectedAgents := make(map[string]bool)
    for _, agent := range manifest.Agents {
        expectedAgents[agent.Name+".md"] = true
    }
    // Find files that aren't expected
    // ...
}
```

This has two weaknesses:
1. Only detects agent orphans (not mena, rules, hooks)
2. User-created agents with names not in the manifest are flagged as orphans

### New Implementation (Phase 3)

The provenance manifest enables ownership-aware orphan detection:

```
function detectOrphansV2(manifest *RiteManifest, claudeDir string, provManifest *ProvenanceManifest) []string:
    if provManifest == nil:
        // No provenance manifest exists. Fall back to existing rite-manifest-membership
        // check for backward compatibility.
        return detectOrphansLegacy(manifest, claudeDir)

    orphans := []string{}

    // Build set of expected agent names from rite manifest
    expectedAgents := set(agent.Name+".md" for agent in manifest.Agents)

    for path, entry in provManifest.Entries:
        if !strings.HasPrefix(path, "agents/"):
            continue  // Orphan detection applies only to agents

        agentFilename := strings.TrimPrefix(path, "agents/")

        if entry.Owner == OwnerUser || entry.Owner == OwnerUnknown:
            continue  // User-owned agents are never orphans

        if entry.Owner == OwnerKnossos && !expectedAgents[agentFilename]:
            orphans = append(orphans, agentFilename)

    return orphans
```

### Migration Strategy

| Condition | Behavior |
|-----------|----------|
| `PROVENANCE_MANIFEST.yaml` exists | Use provenance-based orphan detection |
| `PROVENANCE_MANIFEST.yaml` absent | Fall back to existing `detectOrphans()` (rite manifest membership) |

### Function Signature Compatibility

The new implementation preserves the existing function signature:

```go
func (m *Materializer) detectOrphans(manifest *RiteManifest, claudeDir string) ([]string, error)
```

Internally, it loads the provenance manifest if available and delegates to the provenance-aware path. If no manifest exists, it delegates to the legacy path. Callers are unaffected.

### Extended Orphan Detection (Future)

Phase 3 can extend orphan detection beyond agents to all knossos-owned resources:

- Mena entries in the manifest with owner:knossos that are not in the current rite's projected mena set
- Rule files in the manifest with owner:knossos that no longer have a template source
- Hook files in the manifest with owner:knossos that no longer have a template source

This is not in Phase 2 scope but the manifest structure supports it without changes.

---

## Section 9: Mena Directory-Level Granularity

### Rationale

Mena entries are atomic directories containing an INDEX file and optional support files. Individual file tracking would produce false divergence signals because:
- `StripMenaExtension()` renames files during projection (e.g., `INDEX.dro.md` becomes `INDEX.md`)
- The projection process rewrites all files in the directory even if only one changed
- The logical unit of a mena entry is the directory, not individual files

### Key Format

Mena entries use directory-level keys ending with `/`:

```yaml
commands/commit/:
  owner: knossos
  source_pipeline: materialize
  source_path: mena/operations/commit/
  source_type: project
  checksum: "sha256:..."
  last_synced: "2026-02-09T14:30:00Z"
```

### Checksum Computation

Directory-level checksums use `checksum.Dir()` from `internal/checksum/checksum.go`:

1. Walk the directory recursively, collecting all file relative paths
2. Sort paths lexicographically
3. For each path in order: hash the relative path bytes + file content bytes
4. Return `sha256:` + hex digest of the combined hash

This produces a deterministic checksum for the entire directory regardless of filesystem traversal order.

### Recording Pattern in materializeMena()

After `ProjectMena()` completes, the collector records one entry per projected mena directory:

```
for each projected mena entry directory:
    relativePath := "commands/{name}/" or "skills/{name}/"
    hash, _ := checksum.Dir(filepath.Join(claudeDir, relativePath))
    collector.Record(relativePath, &ProvenanceEntry{
        Owner:          OwnerKnossos,
        SourcePipeline: "materialize",
        SourcePath:     sourceDir + "/",      // e.g., "mena/operations/commit/"
        SourceType:     determinedSourceType,  // "project", "shared", "dependency", etc.
        Checksum:       hash,
        LastSynced:     now,
    })
```

### Divergence Detection for Directories

When checking mena entries during divergence detection:

```
if strings.HasSuffix(path, "/"):
    dirPath := filepath.Join(claudeDir, strings.TrimSuffix(path, "/"))
    currentChecksum, err := checksum.Dir(dirPath)
    // compare currentChecksum to entry.Checksum
```

If the user modifies any file within a mena directory (including adding or removing files), the directory checksum changes and the entry is promoted to owner:user.

### ProjectMena() Integration

`ProjectMena()` in `internal/materialize/project_mena.go` returns a `MenaProjectionResult` that includes the list of projected entries. The integration point captures this return value and records each entry:

```go
type MenaProjectionResult struct {
    Projected []MenaProjectedEntry
    Errors    []error
}

type MenaProjectedEntry struct {
    Name       string // e.g., "commit"
    Target     string // "commands" or "skills"
    SourcePath string // Path to source mena directory
    SourceType string // "project", "shared", "dependency"
}
```

The `materializeMena()` function iterates over `result.Projected` and calls `collector.Record()` for each.

---

## Section 10: Bootstrap Behavior

### First Sync (No Previous Manifest)

When `PROVENANCE_MANIFEST.yaml` does not exist:

1. Divergence detection is skipped (no previous manifest to compare against)
2. Pipeline runs normally, writing all files
3. Collector records all pipeline-written files with `owner: knossos`
4. Pre-existing files NOT written by the pipeline are NOT discovered
5. The manifest is written containing only pipeline-written entries
6. Files in `.claude/` that predate the manifest and are not written by the pipeline remain invisible to the provenance system

This is intentional. The manifest is a record of what the pipeline placed, not a census of all files in `.claude/`. User files that were never managed by knossos do not need provenance tracking.

### Second Sync (Previous Manifest Exists)

1. Load previous manifest
2. Run divergence detection on all knossos-owned entries
3. Promote diverged entries to owner:user
4. Pipeline runs normally
5. Collector records all pipeline-written files
6. Merge: pipeline entries + promoted entries + carried-forward user entries
7. Write updated manifest

### Rite Switch

When switching from rite A to rite B (Decision #7: full rebuild):

1. Load previous manifest (from rite A)
2. Run divergence detection
3. Pipeline materializes rite B (full rebuild of agents, mena, hooks, rules, etc.)
4. Collector records all rite B entries
5. Merge step:
   - Rite A's knossos-owned entries that were NOT promoted to user are dropped (they are no longer relevant; the new rite provides different files)
   - Rite A's user-promoted entries are carried forward (user modifications are preserved regardless of rite)
   - Rite B's pipeline entries are recorded
6. Result: manifest reflects rite B's managed files + user's custom files

### Minimal (Cross-Cutting) Mode

`MaterializeMinimal()` produces a reduced manifest:
- Only hooks, rules, CLAUDE.md, and settings are recorded
- No agents, mena, workflow, or ACTIVE_RITE entries
- `active_rite` field is empty string

---

## Integration Points in Materialize Pipeline

### Overview

Each pipeline stage in `MaterializeWithOptions()` (materialize.go:258) needs one modification: accept a `provenance.Collector` parameter and call `Record()` after each successful write.

### Detailed Integration Points

#### 1. materializeAgents() -- line 522

**Current**: Copies agent files from rite to `.claude/agents/` via `writeIfChanged()` or `copyDirFromFS()`.

**Change**: After each agent file is written, record the entry.

```
After writeIfChanged(destPath, content, 0644) succeeds:
    relPath := "agents/" + relPath  // relative to .claude/
    hash := checksum.Bytes(content)
    collector.Record(relPath, &ProvenanceEntry{
        Owner:          OwnerKnossos,
        SourcePipeline: "materialize",
        SourcePath:     relativize(sourceAgentsDir, ritePath, path),
        SourceType:     string(resolved.Source.Type),
        Checksum:       hash,
        LastSynced:     now,
    })
```

**Function signature change**: `func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error`

#### 2. materializeMena() -- line 604

**Current**: Delegates to `ProjectMena()` which handles collection, routing, and file copying.

**Change**: After `ProjectMena()` completes, iterate over projected entries and record directory-level provenance. Requires `ProjectMena()` to return projected entry metadata (name, target, source path, source type).

**Function signature change**: `func (m *Materializer) materializeMena(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error`

**ProjectMena() change**: Return a `MenaProjectionResult` struct containing the list of projected entries with their source metadata.

#### 3. materializeHooks() -- line 879

**Current**: Copies hook files from `knossos/templates/hooks/` to `.claude/hooks/`.

**Change**: After each hook file is written, record the entry.

**Function signature change**: `func (m *Materializer) materializeHooks(claudeDir string, resolved *ResolvedRite, collector provenance.Collector) (bool, error)`

#### 4. materializeRules() -- line 786

**Current**: Copies rule files from `knossos/templates/rules/` to `.claude/rules/`.

**Change**: After `writeIfChanged(dstPath, content, 0644)` at line 867, record each rule.

**Function signature change**: `func (m *Materializer) materializeRules(claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error`

#### 5. mergeCLAUDEmd() -- line 1020

**Current**: Generates CLAUDE.md from templates via inscription system, writes via `writeIfChanged()`.

**Change**: After `writeIfChanged(claudeMdPath, ...)` at line 1088, record the CLAUDE.md entry.

**Function signature change**: `func (m *Materializer) mergeCLAUDEmd(claudeDir string, renderCtx *inscription.RenderContext, activeRite string, resolved *ResolvedRite, updateManifest bool, collector provenance.Collector) (string, error)`

#### 6. materializeSettingsWithManifest() -- line 938

**Current**: Generates settings.local.json, writes via `saveSettings()`.

**Change**: After `saveSettings(settingsPath, existingSettings)` at line 1133, record the entry.

**Function signature change**: `func (m *Materializer) materializeSettingsWithManifest(claudeDir string, manifest *RiteManifest, collector provenance.Collector) error`

#### 7. materializeWorkflow() -- line 1178

**Current**: Copies workflow.yaml to ACTIVE_WORKFLOW.yaml, writes via `writeIfChanged()`.

**Change**: After `writeIfChanged(dstPath, content, 0644)` at line 1189, record the entry. If workflow was removed (no workflow.yaml in rite), do not record.

**Function signature change**: `func (m *Materializer) materializeWorkflow(claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error`

#### 8. writeActiveRite() -- line 1194

**Current**: Writes ACTIVE_RITE marker file via `writeIfChanged()`.

**Change**: After the write, record the entry.

**Function signature change**: `func (m *Materializer) writeActiveRite(riteName, claudeDir string, collector provenance.Collector) error`

#### 9. trackState() -- line 1137

**Current**: Writes sync/state.json via `stateManager.Save()`.

**Change**: After `stateManager.Save(state)` at line 1162, record the entry.

**Function signature change**: `func (m *Materializer) trackState(manifest *RiteManifest, activeRiteName string, collector provenance.Collector) error`

### Pipeline Orchestration Change

`MaterializeWithOptions()` is modified to:

1. Create a `provenance.Collector` at the start
2. Load the previous `ProvenanceManifest` (if exists)
3. Run divergence detection against the previous manifest
4. Thread the collector through all pipeline stages
5. After all stages complete, call `buildAndSaveManifest()` to merge and write

```go
func (m *Materializer) MaterializeWithOptions(activeRiteName string, opts Options) (*Result, error) {
    // ... existing setup ...

    // Provenance: load previous manifest and detect divergence
    claudeDir := m.getClaudeDir()
    prevManifest, _ := provenance.Load(filepath.Join(claudeDir, provenance.ManifestFileName))
    promoted := provenance.DetectDivergence(claudeDir, prevManifest)
    collector := provenance.NewCollector()

    // ... existing pipeline stages with collector parameter ...

    // Provenance: build final manifest and save
    finalManifest := provenance.BuildManifest(collector, promoted, activeRiteName)
    if err := provenance.Save(filepath.Join(claudeDir, provenance.ManifestFileName), finalManifest); err != nil {
        return nil, errors.Wrap(errors.CodeGeneralError, "failed to save provenance manifest", err)
    }

    return result, nil
}
```

---

## Package Structure

```
internal/provenance/
    provenance.go       // ProvenanceManifest, ProvenanceEntry, OwnerType types
    collector.go        // Collector interface, defaultCollector, NullCollector
    divergence.go       // DetectDivergence(), computeCurrentChecksum()
    manifest.go         // Load(), Save(), BuildManifest(), Validate()
    provenance_test.go  // Unit tests for all functions
```

### Dependency Graph

```
internal/provenance/
    imports: internal/checksum, internal/errors, gopkg.in/yaml.v3

internal/materialize/
    imports: internal/provenance (NEW), internal/checksum, ...existing...

cmd/ari/cmd/provenance.go
    imports: internal/provenance
```

No circular dependencies. `provenance` is a leaf package that depends only on `checksum`, `errors`, and `yaml.v3`.

---

## Backward Compatibility

### Classification: COMPATIBLE

The provenance manifest is purely additive. No existing behavior changes.

| Aspect | Impact |
|--------|--------|
| Existing `.claude/` directories | Continue to work. First sync writes the manifest alongside existing files. No existing files are modified or removed. |
| KNOSSOS_MANIFEST.yaml | Untouched. Two manifests coexist. |
| Usersync manifests | Untouched. Usersync operates on `~/.claude/`, provenance operates on `{project}/.claude/`. |
| Pipeline function signatures | Internal change (added `collector` parameter). No public API impact. |
| detectOrphans() | Backward compatible: falls back to legacy behavior when no provenance manifest exists. |
| Minimal mode | Gets provenance tracking (reduced entry set). No behavior change for the user. |
| CLI | New `ari provenance show` command. Does not conflict with existing commands. |

### Migration Path

No migration required. The system bootstraps automatically:
1. First `ari rite start` after implementation writes `PROVENANCE_MANIFEST.yaml`
2. Subsequent syncs use the manifest for divergence detection
3. Pre-existing installations get the manifest on their next sync

---

## Integration Test Matrix

### Test Satellites

| Type | Description | Configuration |
|------|-------------|---------------|
| **baseline** | Standard knossos project with ecosystem rite | 6 agents, 10+ mena, hooks, rules |
| **minimal** | Cross-cutting mode, no rite | No agents, no mena, hooks and rules only |
| **complex** | Multiple dependency rites, shared mena, user customizations | 4 agents from rite, 2 user agents, shared+dependency mena |
| **fresh** | New project, first-ever sync | Empty `.claude/` directory |

### Test Cases

| ID | Satellite | Test | Expected Outcome |
|----|-----------|------|------------------|
| T01 | baseline | Sync, verify manifest created | `PROVENANCE_MANIFEST.yaml` exists with all pipeline entries as owner:knossos |
| T02 | baseline | Sync twice (idempotency) | Second sync produces identical manifest content |
| T03 | baseline | Modify a knossos agent, resync | Modified agent promoted to owner:user; not overwritten on next sync |
| T04 | baseline | Delete a knossos agent, resync | Deleted agent entry promoted to owner:user with empty checksum |
| T05 | minimal | Minimal sync | Manifest created with hooks, rules, CLAUDE.md, settings only. No agent/mena entries. |
| T06 | complex | Sync with shared+dependency mena | Mena entries have correct source_type (shared, dependency, project) |
| T07 | complex | User agent coexists with rite agents | User agent has owner:user; rite agents have owner:knossos |
| T08 | fresh | First sync on empty .claude/ | Manifest bootstraps with all pipeline entries as knossos. No unknown entries. |
| T09 | baseline | Switch rite A to rite B | Rite A entries removed; rite B entries added; user modifications carried forward |
| T10 | baseline | `ari provenance show` table output | Correct PATH, OWNER, SOURCE, STATUS columns for all entries |
| T11 | baseline | `ari provenance show --json` | Valid JSON matching manifest schema with computed status field |
| T12 | baseline | `ari provenance show --diverged` | Shows only entries where on-disk checksum differs from manifest |
| T13 | complex | Orphan detection with manifest | User agents NOT flagged as orphans; removed rite agents flagged |
| T14 | baseline | Mena directory modified by user | Entire mena directory promoted to owner:user (directory-level checksum change) |
| T15 | baseline | Manifest schema validation | Invalid manifest (missing required fields) produces clear error on Load() |

---

## Phase 2 vs Phase 3 Scope Boundary

### Phase 2: Manifest + Pipeline Integration

| Work Item | Description |
|-----------|-------------|
| Create `internal/provenance/` package | Types, Collector, manifest Load/Save, validation |
| Wire Collector into materialize pipeline | All 9 pipeline stages record entries |
| Divergence detection | Run before pipeline; promote diverged entries |
| Merge algorithm | Combine pipeline entries + promoted entries + user entries |
| Write manifest at end of sync | `PROVENANCE_MANIFEST.yaml` written after all stages complete |
| Unit tests | Full coverage of provenance package |
| Integration test T01-T09 | Core pipeline scenarios |

### Phase 3: Orphan Detection + CLI + Warnings

| Work Item | Description |
|-----------|-------------|
| `ari provenance show` command | Table, JSON, YAML output with filter flags |
| Provenance-aware orphan detection | Replace rite-manifest-membership with manifest-based detection |
| Divergence warnings | Print warnings when knossos files have been modified locally |
| `--force` override for diverged files | Allow pipeline to overwrite user-promoted files when explicitly requested |
| Integration test T10-T15 | CLI and advanced scenarios |

---

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-09 | Context Architect | Initial TDD from stakeholder interview decisions |
