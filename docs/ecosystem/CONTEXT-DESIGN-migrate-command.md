# Context Design: `ari migrate roster-to-knossos` Command

**Date**: 2026-02-06
**Architect**: Context Architect
**Reference**: Knossos Migration Sprint - Satellite Manifest Migration
**Prerequisite**: Codebase renamed from "roster" to "knossos" (commit bbbc026)

---

## Executive Summary

Satellite projects still carry manifest files referencing the old "roster" branding in their `source` fields and CEM metadata keys. The `ari migrate roster-to-knossos` command provides a safe, idempotent CLI tool that satellite owners run once to rewrite these references. The command is dry-run by default, requires explicit `--apply` to mutate, and generates a shell script for environment variable migration that users review before executing.

**Scope**:
- User-level manifests at `~/.claude/USER_*_MANIFEST.json` (4 files: agents, skills, commands, hooks)
- Project-level CEM manifest at `.claude/.cem/manifest.json`
- Environment variable advisory for `ROSTER_*` variables
- Shell profile migration script generation

**Approach**: New `migrate` command group under root (`ari migrate`), with `roster-to-knossos` as first subcommand. Follows the established pattern from `internal/cmd/session/migrate.go` for backup/dry-run/apply flow. Does NOT require a project context (works at user-level); project-level CEM migration is opt-in via `--project` flag.

---

## Solution Architecture

### Options Considered

**Option A: Standalone script (rejected)**
A shell script shipped alongside `ari` that uses `sed` to rewrite JSON files. Rejected because: (1) fragile JSON manipulation via sed risks corruption, (2) no structured output for automation, (3) inconsistent with the Go CLI pattern established by all other `ari` commands, (4) no dry-run/backup semantics without reinventing what Cobra gives us.

**Option B: Subcommand under `ari session migrate` (rejected)**
Extending the existing session migration command to also handle manifest rewrites. Rejected because: (1) session migration operates on session context files with lock semantics -- manifests have no locks, (2) conflates two unrelated migration domains, (3) the session migrate command has a specific schema version progression (v1 -> v2.1) that does not apply here.

**Option C: New `ari migrate` command group with `roster-to-knossos` subcommand (selected)**
A dedicated `migrate` command group that can host future migration subcommands. The `roster-to-knossos` subcommand handles manifest rewriting, env var advisory, and script generation. Selected because: (1) clean separation of concerns from session migration, (2) extensible for future migrations (path renames, config migrations), (3) follows the established Cobra command group pattern (`internal/cmd/manifest/`, `internal/cmd/session/`), (4) core migration logic is testable in pure functions.

---

## Command Interface Specification

### Command Structure

```
ari migrate roster-to-knossos [flags]

Flags:
  --dry-run              Preview changes without applying (default: true)
  --apply                Execute the migration (sets dry-run to false)
  --backup               Create backup files before rewriting (default: true)
  --no-backup            Skip backup creation
  --generate-script      Output shell profile migration script to stdout
  --script-file PATH     Write shell profile migration script to file instead of stdout
  --project PATH         Target project directory for CEM manifest (default: cwd if .claude/ exists)
  --skip-project         Skip CEM manifest migration even if project detected
  --skip-user            Skip user-level manifest migration
```

### Help Text

```
Migrates satellite manifests from "roster" to "knossos" branding.

Rewrites source fields in USER_*_MANIFEST.json files and CEM manifest
metadata keys. Safe to run multiple times (idempotent).

By default, runs in dry-run mode. Use --apply to execute changes.

Targets:
  User manifests   ~/.claude/USER_{AGENT,SKILL,COMMAND,HOOKS}_MANIFEST.json
  CEM manifest     .claude/.cem/manifest.json (project-level)
  Env variables    Advisory for ROSTER_* environment variables

Examples:
  ari migrate roster-to-knossos                     # Preview all changes
  ari migrate roster-to-knossos --apply             # Execute migration
  ari migrate roster-to-knossos --generate-script   # Output env var migration script
  ari migrate roster-to-knossos --apply --no-backup # Migrate without backups
  ari migrate roster-to-knossos --skip-project      # Only migrate user manifests
```

### Flag Semantics

| Flag | Default | Behavior |
|------|---------|----------|
| `--dry-run` | `true` | Preview mode; prints what would change. Mutually exclusive with `--apply`. |
| `--apply` | `false` | Execute the migration. Explicitly sets dry-run to false. |
| `--backup` | `true` | Before rewriting any file, copy original to `<filename>.roster-backup`. |
| `--no-backup` | `false` | Skip backup creation. Overrides `--backup`. |
| `--generate-script` | `false` | After migration report, output shell script for env var updates. |
| `--script-file` | `""` | Write script to this path instead of stdout. Implies `--generate-script`. |
| `--project` | auto-detect | Directory containing `.claude/.cem/manifest.json`. |
| `--skip-project` | `false` | Do not attempt CEM manifest migration. |
| `--skip-user` | `false` | Do not attempt user-level manifest migration. |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (including dry-run with changes detected) |
| 1 | General error (file I/O, unexpected failure) |
| 2 | Usage error (invalid flag combination) |
| 11 | Migration failed (corrupt manifest that cannot be parsed) |
| 15 | Parse error (JSON parsing failed on a manifest) |

### Mutual Exclusion Rules

- `--apply` and `--dry-run` cannot both be explicitly set. `--apply` wins if both specified (with warning).
- `--backup` and `--no-backup` cannot both be explicitly set. `--no-backup` wins if both specified.
- `--skip-project` and `--project PATH` cannot both be specified. Error if both provided.
- `--skip-user` and `--skip-project` cannot both be true. Error: nothing to migrate.

---

## Data Flow

```
                      ari migrate roster-to-knossos --apply
                                     |
                    +----------------+----------------+
                    |                                 |
            User Manifests                    CEM Manifest
            (unless --skip-user)              (unless --skip-project)
                    |                                 |
    +-------+-------+-------+-------+       .claude/.cem/manifest.json
    |       |       |       |                         |
  AGENT   SKILL   CMD    HOOKS               Read JSON (map[string]any)
    |       |       |       |                         |
    +---+---+---+---+                    Rewrite keys:
        |                                  .roster.* -> .knossos.*
  For each manifest:                       .team.roster_path -> .team.knossos_path
    Read JSON (manifestJSON)               managed_files[].source "roster" -> "knossos"
    For each entry:                                   |
      "roster" -> "knossos"              Write JSON (backup first if --backup)
      "roster-diverged" -> "knossos-diverged"
    Write JSON (backup first if --backup)
                    |
                    +------> Env Var Scan (os.Environ)
                    |           Detect ROSTER_* variables
                    |
                    +------> Script Generation (if --generate-script)
                    |           Generate sed commands for shell profiles
                    |
                    v
              RosterMigrateOutput (JSON or text)
```

---

## File Structure

### New Files

| File | Purpose |
|------|---------|
| `internal/cmd/migrate/migrate.go` | `NewMigrateCmd()` parent command group |
| `internal/cmd/migrate/roster_to_knossos.go` | Core migration logic: manifest rewriting, env scan, script generation |
| `internal/cmd/migrate/roster_to_knossos_test.go` | Unit tests for pure migration functions |

### Modified Files

| File | Change |
|------|--------|
| `internal/cmd/root/root.go` | Register `migrate.NewMigrateCmd()` in `init()` |
| `internal/output/output.go` | Add `RosterMigrateOutput` struct with `Text()` method |

---

## File-by-File Implementation Specification

### 1. `internal/cmd/migrate/migrate.go`

**Package**: `migrate`

**Purpose**: Parent command group. Follows the pattern established by `internal/cmd/manifest/manifest.go` (non-session command group using `common.BaseContext`).

**Structs**:

```go
// cmdContext holds shared state for migrate commands.
type cmdContext struct {
    common.BaseContext
}
```

**Functions**:

```go
// NewMigrateCmd creates the migrate command group.
// Signature matches the manifest pattern: accepts outputFlag, verboseFlag, projectDir pointers.
func NewMigrateCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command
```

**Registration**: The parent command does NOT require project context. It uses `common.SetNeedsProject(cmd, false, true)` because migrate operates at user-level by default.

**Subcommand registration**:
```go
cmd.AddCommand(newRosterToKnossosCmd(ctx))
```

**Helper**:
```go
func (c *cmdContext) getPrinter() *output.Printer {
    return c.GetPrinter(output.FormatText)
}
```

### 2. `internal/cmd/migrate/roster_to_knossos.go`

**Package**: `migrate`

**Purpose**: Core migration command and logic.

**Options struct**:

```go
type rosterToKnossosOptions struct {
    apply          bool
    backup         bool
    noBackup       bool
    generateScript bool
    scriptFile     string
    projectDir     string
    skipProject    bool
    skipUser       bool
}
```

**Cobra command constructor**:

```go
func newRosterToKnossosCmd(ctx *cmdContext) *cobra.Command
```

Binds all flags as specified in the Command Interface section. The `RunE` function calls `runRosterToKnossos(ctx, opts)`.

**Main execution function**:

```go
func runRosterToKnossos(ctx *cmdContext, opts rosterToKnossosOptions) error
```

Flow:
1. Validate flag combinations (mutual exclusions). Return `errors.New(errors.CodeUsageError, ...)` on conflict.
2. Determine effective dry-run state: `dryRun := !opts.apply`.
3. Determine effective backup state: `backup := opts.backup && !opts.noBackup`.
4. Initialize `output.RosterMigrateOutput{DryRun: dryRun}`.
5. If `!opts.skipUser`: call `migrateUserManifests(dryRun, backup)`, accumulate results.
6. If `!opts.skipProject`: resolve project dir (from `--project`, or `ctx.ProjectDir`, or cwd auto-detect). Call `migrateCEMManifest(projectDir, dryRun, backup)`, accumulate results.
7. Scan environment variables: call `scanRosterEnvVars()`, populate `EnvVarsDetected`.
8. If `opts.generateScript || opts.scriptFile != ""`: call `generateMigrationScript(envVars, opts.scriptFile)`.
9. Print result via `printer.Print(result)`.

**Pure functions (testable without filesystem)**:

```go
// rewriteUserManifestBytes rewrites source fields in a user manifest JSON blob.
// Returns the rewritten JSON bytes and the count of entries changed.
// Returns the input unchanged if no roster references found (idempotent).
func rewriteUserManifestBytes(data []byte) ([]byte, int, error)
```

Algorithm:
1. Unmarshal into `map[string]interface{}` to preserve all fields.
2. For each resource type key (`"agents"`, `"skills"`, `"commands"`, `"hooks"`):
   a. If value is `map[string]interface{}`, iterate entries.
   b. For each entry, if `"source"` field equals `"roster"`, set to `"knossos"`, increment counter.
   c. If `"source"` field equals `"roster-diverged"`, set to `"knossos-diverged"`, increment counter.
3. Marshal back with `json.MarshalIndent(data, "", "  ")` and append newline.
4. Return bytes, count, nil.

```go
// rewriteCEMManifestBytes rewrites roster references in a CEM manifest JSON blob.
// Returns the rewritten JSON bytes and the count of fields changed.
func rewriteCEMManifestBytes(data []byte) ([]byte, int, error)
```

Algorithm:
1. Unmarshal into `map[string]interface{}`.
2. If key `"roster"` exists at top level:
   a. Move its value to key `"knossos"`.
   b. Delete `"roster"` key.
   c. Increment counter by 1 (counts as one logical change: the metadata block rename).
3. If key `"team"` exists and contains `"roster_path"`:
   a. Move value to `"knossos_path"`.
   b. Delete `"roster_path"`.
   c. Increment counter.
4. If `"managed_files"` exists as `[]interface{}`:
   a. For each entry with `"source": "roster"`, set to `"knossos"`, increment counter.
5. If `"migration"` exists and contains `"skeleton_path"`, leave unchanged (historical reference).
6. Marshal back, return.

```go
// scanRosterEnvVars returns a list of ROSTER_* environment variables currently set.
func scanRosterEnvVars() []EnvVarMapping
```

Where:
```go
type EnvVarMapping struct {
    Old   string // e.g., "ROSTER_HOME"
    New   string // e.g., "KNOSSOS_HOME"
    Value string // current value
}
```

Algorithm:
1. Call `os.Environ()`.
2. For each entry starting with `"ROSTER_"`:
   a. Split on `=`.
   b. Compute new name by replacing `ROSTER_` prefix with `KNOSSOS_`.
   c. Append to result.
3. Return result.

```go
// generateMigrationScript produces a shell script that updates shell profile env vars.
// If outputPath is empty, returns script as string (caller writes to stdout).
// If outputPath is non-empty, writes script to that file with 0755 permissions.
func generateMigrationScript(envVars []EnvVarMapping) string
```

Algorithm:
1. Build script header with `#!/bin/bash` and generation comment.
2. Detect shell profile files: check existence of `~/.zshrc`, `~/.bashrc`, `~/.bash_profile`, `~/.profile`.
3. For each detected env var:
   a. For each existing profile file:
      - Add `sed -i.bak 's/ROSTER_OLD_NAME/KNOSSOS_NEW_NAME/g' <profile>` line.
4. Add verification step at end: `echo "Migration complete. Restart your shell or run: source <profile>"`.
5. Return script string.

**Filesystem functions** (non-pure, but thin wrappers):

```go
// migrateUserManifests discovers and rewrites all USER_*_MANIFEST.json files.
func migrateUserManifests(dryRun, backup bool) ([]ManifestMigration, error)
```

Algorithm:
1. Get `paths.UserClaudeDir()` -> `~/.claude`.
2. Glob for `USER_*_MANIFEST.json` files using `filepath.Glob`.
3. For each file:
   a. Read file contents.
   b. Call `rewriteUserManifestBytes(data)`.
   c. If count == 0, record as skipped (already migrated).
   d. If `dryRun`, record changes without writing.
   e. If `!dryRun && backup`, copy original to `<path>.roster-backup`.
   f. If `!dryRun`, write rewritten bytes atomically (temp file + rename).
4. Return results.

```go
// migrateCEMManifest discovers and rewrites the CEM manifest in a project.
func migrateCEMManifest(projectDir string, dryRun, backup bool) (*ManifestMigration, error)
```

Algorithm:
1. Construct path: `filepath.Join(projectDir, ".claude", ".cem", "manifest.json")`.
2. If file does not exist, return nil (no CEM manifest to migrate).
3. Read file, call `rewriteCEMManifestBytes(data)`.
4. Follow same backup/write pattern as user manifests.

**Internal types** (not exported, used within the command):

```go
// ManifestMigration records the result of migrating a single manifest file.
type ManifestMigration struct {
    Path            string `json:"path"`
    EntriesRewritten int   `json:"entries_rewritten"`
    Skipped         bool   `json:"skipped"`
    SkipReason      string `json:"skip_reason,omitempty"`
    BackupPath      string `json:"backup_path,omitempty"`
    Error           string `json:"error,omitempty"`
}
```

### 3. `internal/output/output.go` -- New Output Struct

Add `RosterMigrateOutput` after the existing `FrayOutput` struct (line 569).

```go
// RosterMigrateOutput represents the result of roster-to-knossos migration.
type RosterMigrateOutput struct {
    DryRun             bool                `json:"dry_run"`
    ManifestsFound     int                 `json:"manifests_found"`
    ManifestsChanged   int                 `json:"manifests_changed"`
    ManifestsSkipped   int                 `json:"manifests_skipped"`
    EntriesRewritten   int                 `json:"entries_rewritten"`
    UserManifests      []ManifestMigResult `json:"user_manifests,omitempty"`
    CEMManifest        *ManifestMigResult  `json:"cem_manifest,omitempty"`
    EnvVarsDetected    []EnvVarDetected    `json:"env_vars_detected,omitempty"`
    BackupsCreated     []string            `json:"backups_created,omitempty"`
    ScriptGenerated    bool                `json:"script_generated,omitempty"`
    ScriptPath         string              `json:"script_path,omitempty"`
    Errors             []string            `json:"errors,omitempty"`
}

// ManifestMigResult records migration outcome for a single manifest.
type ManifestMigResult struct {
    Path             string `json:"path"`
    EntriesRewritten int    `json:"entries_rewritten"`
    Skipped          bool   `json:"skipped"`
    SkipReason       string `json:"skip_reason,omitempty"`
    BackupPath       string `json:"backup_path,omitempty"`
}

// EnvVarDetected records a detected ROSTER_* environment variable.
type EnvVarDetected struct {
    Current string `json:"current"`
    Replace string `json:"replace"`
    Value   string `json:"value"`
}
```

**Text() method**:

```go
func (r RosterMigrateOutput) Text() string
```

Output format for text mode:

```
Roster-to-Knossos Migration (dry-run)

User Manifests:
  ~/.claude/USER_AGENT_MANIFEST.json    2 entries rewritten
  ~/.claude/USER_SKILL_MANIFEST.json    skipped (already migrated)
  ~/.claude/USER_COMMAND_MANIFEST.json  1 entry rewritten
  ~/.claude/USER_HOOKS_MANIFEST.json    0 entries (no roster references)

CEM Manifest:
  .claude/.cem/manifest.json            4 fields rewritten

Summary: 3 manifests changed, 1 skipped, 7 entries rewritten

Environment variables to update:
  ROSTER_HOME -> KNOSSOS_HOME (current: /Users/tom/Code/roster)

Use --apply to execute this migration.
```

When `DryRun` is false:

```
Roster-to-Knossos Migration

User Manifests:
  ~/.claude/USER_AGENT_MANIFEST.json    2 entries rewritten (backup: .roster-backup)
  ...

Migration complete. 3 manifests updated.
```

### 4. `internal/cmd/root/root.go` -- Registration

**Change**: Add import for `migrate` package and register the command.

At line 15 (imports), add:
```go
"github.com/autom8y/knossos/internal/cmd/migrate"
```

At line 124 (after `rootCmd.AddCommand(tribute.NewTributeCmd(...))`), add:
```go
rootCmd.AddCommand(migrate.NewMigrateCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
```

Rationale for placement: `migrate` is a top-level command group like `session`, `manifest`, `sync`. It takes the same `(outputFlag, verboseFlag, projectDir)` signature as `manifest.NewManifestCmd`.

---

## Backward Compatibility

### Classification: COMPATIBLE

This is a new command. No existing commands, schemas, or behaviors change. The manifests being rewritten are the same JSON format before and after; only string values within entries change.

### Idempotency Guarantee

Running `ari migrate roster-to-knossos --apply` twice:
1. First run: rewrites `"roster"` -> `"knossos"`, creates backups.
2. Second run: `rewriteUserManifestBytes()` finds zero entries with `"roster"` source. Reports all manifests as skipped. No files written. No new backups created.

The function `rewriteUserManifestBytes` returns count=0 when no `"roster"` or `"roster-diverged"` source values exist, causing the caller to skip the file.

### Backup Safety

- Backup files use `.roster-backup` extension (distinct from `.v1.backup` used by session migration and `.corrupt` used by usersync).
- Backups are NOT overwritten on subsequent runs. If `.roster-backup` already exists, the migration skips backup creation for that file (backup already exists from previous run).
- Atomic writes via temp file + rename prevent partial writes on crash.

---

## Schema Definitions

### User Manifest Source Field

**Before migration**:
```json
{
  "source": "roster"       // or "roster-diverged"
}
```

**After migration**:
```json
{
  "source": "knossos"      // or "knossos-diverged"
}
```

**Validation**: The `SourceType` constants in `internal/usersync/usersync.go` already define `SourceKnossos = "knossos"` and `SourceDiverged = "knossos-diverged"` (lines 59-61). The migration command aligns manifests with these canonical values. No schema version bump needed -- the manifest format is unchanged; only the string values are updated.

### CEM Manifest Key Migration

**Before**:
```json
{
  "roster": {
    "path": "/path/to/knossos",
    "commit": "abc123",
    "ref": "main",
    "last_sync": "2026-01-07T17:57:29Z"
  },
  "team": {
    "roster_path": "/path/to/rites/10x-dev"
  },
  "managed_files": [
    {"source": "roster", ...}
  ]
}
```

**After**:
```json
{
  "knossos": {
    "path": "/path/to/knossos",
    "commit": "abc123",
    "ref": "main",
    "last_sync": "2026-01-07T17:57:29Z"
  },
  "team": {
    "knossos_path": "/path/to/rites/10x-dev"
  },
  "managed_files": [
    {"source": "knossos", ...}
  ]
}
```

**Validation**: The CEM manifest uses `schema_version: 3`. The migration does not change the schema version because the structure is unchanged -- only key names and string values update. Consumers of the CEM manifest (the sync infrastructure) already reference `knossos` as the platform name post-rename (commit bbbc026).

---

## Edge Cases and Error Handling

### Edge Case Matrix

| Scenario | Behavior |
|----------|----------|
| Manifest file does not exist | Skip silently; record in output as "not found". Not an error. |
| Manifest is empty file (0 bytes) | Skip with reason "empty file". Not an error. |
| Manifest has invalid JSON | Report as error in `Errors` list. Use `errors.CodeParseError`. Continue processing other manifests (non-fatal). |
| Manifest already migrated (no roster refs) | Skip with reason "already migrated". Count as skipped, not changed. |
| Manifest has mixed sources (some roster, some knossos) | Rewrite only the roster entries. This handles partial previous migrations. |
| Manifest has `"source": "user"` entries | Leave unchanged. Only `"roster"` and `"roster-diverged"` are rewritten. |
| Backup file already exists | Do not overwrite existing backup. Log as verbose info. |
| `~/.claude` directory does not exist | Skip user manifest migration with advisory message. Not an error. |
| `--project` points to dir without `.claude/.cem/manifest.json` | Skip CEM migration with advisory. Not an error. |
| No project detected and `--skip-project` not set | Auto-detect: try cwd for `.claude/.cem/manifest.json`. If not found, skip CEM silently. |
| File permission denied on read | Report as error for that file. Continue with other files. |
| File permission denied on write | Report as error for that file. Continue with other files. |
| `--apply` with `--skip-user` and `--skip-project` | Error: nothing to migrate. Exit code 2 (usage error). |
| JSON has unexpected structure (missing fields) | `rewriteUserManifestBytes` uses `map[string]interface{}` traversal with type assertions. Missing or wrong-typed fields are silently skipped (defensive). |
| CEM manifest has no `"roster"` key | No changes to CEM manifest. Counted as skipped. |
| No ROSTER_* env vars detected | `EnvVarsDetected` is empty. No script generated unless `--generate-script` specified (in which case, script is a no-op with comment). |

### Error Handling Strategy

1. **Non-fatal per-file errors**: Each manifest is processed independently. A parse error on one manifest does not abort processing of others. Errors accumulate in `RosterMigrateOutput.Errors`.
2. **Fatal flag errors**: Invalid flag combinations produce immediate exit with `CodeUsageError`.
3. **Atomic writes**: Write to temp file, then `os.Rename`. If rename fails, remove temp file and report error.
4. **Backup before write**: Backup creation failure aborts the write for that specific file (data safety). Reported as error, processing continues.

---

## Environment Variable Migration

### Known Mappings

| Old Variable | New Variable | Notes |
|-------------|-------------|-------|
| `ROSTER_HOME` | `KNOSSOS_HOME` | Platform home directory |
| `ROSTER_SYNC_DEBUG` | `KNOSSOS_SYNC_DEBUG` | Sync debug logging |
| `ROSTER_VERBOSE` | `KNOSSOS_VERBOSE` | Verbose output |
| `ROSTER_PREF_*` | `KNOSSOS_PREF_*` | User preferences (wildcard) |

### Detection Algorithm

The command does NOT have a hardcoded list. It scans `os.Environ()` for any variable starting with `ROSTER_` and proposes the `KNOSSOS_` equivalent. This future-proofs against new variables.

### Generated Script Format

```bash
#!/bin/bash
# Generated by: ari migrate roster-to-knossos
# Date: 2026-02-06T15:30:00Z
#
# This script updates your shell profile to replace ROSTER_* env vars
# with their KNOSSOS_* equivalents. Review before running.
#
# Detected profile files:
#   ~/.zshrc
#
# Detected variables:
#   ROSTER_HOME -> KNOSSOS_HOME

set -euo pipefail

# Back up profile before modification
cp ~/.zshrc ~/.zshrc.pre-knossos-migrate

# ROSTER_HOME -> KNOSSOS_HOME
sed -i '' 's/ROSTER_HOME/KNOSSOS_HOME/g' ~/.zshrc

echo "Profile updated. Run: source ~/.zshrc"
echo "Backup saved to: ~/.zshrc.pre-knossos-migrate"
```

Notes:
- Uses `sed -i ''` (macOS) vs `sed -i` (Linux). The script detects platform via `uname -s` and adjusts.
- The script is advisory-only: output to stdout or file, never auto-executed.
- If no ROSTER_* vars detected, script contains a comment: `# No ROSTER_* environment variables detected. Nothing to do.`

---

## Test Plan

### Unit Tests (`internal/cmd/migrate/roster_to_knossos_test.go`)

**Test: rewriteUserManifestBytes**

| Test Case | Input | Expected Output | Validates |
|-----------|-------|-----------------|-----------|
| `TestRewriteUserManifest_RosterToKnossos` | Manifest with `"source": "roster"` entries | All entries changed to `"knossos"`, count > 0 | Core rewrite |
| `TestRewriteUserManifest_DivergedToKnossosDiverged` | Manifest with `"source": "roster-diverged"` | Changed to `"knossos-diverged"` | Diverged handling |
| `TestRewriteUserManifest_AlreadyMigrated` | Manifest with `"source": "knossos"` | Unchanged bytes, count == 0 | Idempotency |
| `TestRewriteUserManifest_MixedSources` | Mix of roster, knossos, user entries | Only roster entries changed | Selective rewrite |
| `TestRewriteUserManifest_EmptyManifest` | `{}` | Unchanged, count == 0 | Empty handling |
| `TestRewriteUserManifest_InvalidJSON` | `not json` | Error returned | Error handling |
| `TestRewriteUserManifest_PreservesOtherFields` | Manifest with checksum, installed_at | Non-source fields unchanged | Field preservation |

**Test: rewriteCEMManifestBytes**

| Test Case | Input | Expected Output | Validates |
|-----------|-------|-----------------|-----------|
| `TestRewriteCEM_RosterKeyRename` | CEM with `"roster": {...}` | Key renamed to `"knossos"`, values preserved | Key migration |
| `TestRewriteCEM_TeamRosterPath` | CEM with `"team.roster_path"` | Renamed to `"team.knossos_path"` | Nested key rename |
| `TestRewriteCEM_ManagedFilesSource` | CEM with managed_files `"source": "roster"` | Changed to `"knossos"` | Array entry rewrite |
| `TestRewriteCEM_AlreadyMigrated` | CEM with `"knossos": {...}` | Unchanged, count == 0 | Idempotency |
| `TestRewriteCEM_FullManifest` | Complete CEM manifest (real-world structure) | All roster references migrated | Integration |
| `TestRewriteCEM_NoRosterKey` | CEM without roster metadata | Unchanged, count == 0 | Absent key handling |

**Test: scanRosterEnvVars**

| Test Case | Setup | Expected | Validates |
|-----------|-------|----------|-----------|
| `TestScanEnvVars_DetectsRosterHome` | Set `ROSTER_HOME` | Returns mapping | Detection |
| `TestScanEnvVars_NoRosterVars` | Clean env | Empty result | Clean env |
| `TestScanEnvVars_MultipleVars` | Set `ROSTER_HOME`, `ROSTER_VERBOSE` | Both detected | Multiple |
| `TestScanEnvVars_PrefixVars` | Set `ROSTER_PREF_FOO` | Detected with `KNOSSOS_PREF_FOO` mapping | Wildcard |

**Test: generateMigrationScript**

| Test Case | Input | Expected | Validates |
|-----------|-------|----------|-----------|
| `TestGenerateScript_WithVars` | 2 env var mappings | Script with sed commands | Script content |
| `TestGenerateScript_NoVars` | Empty mappings | Script with "nothing to do" comment | Empty case |
| `TestGenerateScript_HasShebang` | Any input | Starts with `#!/bin/bash` | Format |

### Filesystem Integration Tests

| Test Case | Setup | Action | Validation |
|-----------|-------|--------|------------|
| `TestMigrateUserManifests_DryRun` | Create temp dir with roster manifests | Run with dryRun=true | Files unchanged, output reports changes |
| `TestMigrateUserManifests_Apply` | Create temp dir with roster manifests | Run with dryRun=false | Files rewritten, backups created |
| `TestMigrateUserManifests_ApplyNoBackup` | Create temp dir with roster manifests | Run with dryRun=false, backup=false | Files rewritten, no backups |
| `TestMigrateUserManifests_Idempotent` | Run migration, then run again | Second run | All manifests skipped |
| `TestMigrateUserManifests_MissingDir` | No ~/.claude dir | Run | No error, empty results |
| `TestMigrateCEMManifest_Apply` | Create temp dir with CEM manifest | Run with dryRun=false | CEM manifest rewritten |
| `TestMigrateCEMManifest_NotFound` | No CEM manifest | Run | No error, nil result |
| `TestBackupNotOverwritten` | Create manifest + backup, then migrate again | Run --apply | Original backup preserved |

### Test Fixtures

Create test fixtures as Go string constants in the test file:

```go
const testUserManifestRoster = `{
  "manifest_version": "1.0",
  "last_sync": "2026-01-15T10:00:00Z",
  "agents": {
    "moirai.md": {
      "source": "roster",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "abc123"
    },
    "custom.md": {
      "source": "user",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "def456"
    }
  }
}`

const testUserManifestKnossos = `{
  "manifest_version": "1.0",
  "last_sync": "2026-01-15T10:00:00Z",
  "agents": {
    "moirai.md": {
      "source": "knossos",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "abc123"
    },
    "custom.md": {
      "source": "user",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "def456"
    }
  }
}`

const testCEMManifestRoster = `{
  "schema_version": 3,
  "roster": {
    "path": "/Users/test/Code/knossos",
    "commit": "abc123",
    "ref": "main",
    "last_sync": "2026-01-07T17:57:29Z"
  },
  "team": {
    "name": "10x-dev",
    "roster_path": "/Users/test/Code/roster/rites/10x-dev"
  },
  "managed_files": [
    {"path": ".claude/commands", "source": "roster"},
    {"path": ".claude/hooks", "source": "roster"}
  ]
}`
```

---

## Implementation Sequence

### Phase 1: Output Struct (15 min)

1. Add `RosterMigrateOutput`, `ManifestMigResult`, `EnvVarDetected` to `internal/output/output.go`.
2. Implement `Text()` method.
3. Verify compilation.

### Phase 2: Pure Functions (45 min)

1. Create `internal/cmd/migrate/roster_to_knossos.go` with the pure functions:
   - `rewriteUserManifestBytes`
   - `rewriteCEMManifestBytes`
   - `scanRosterEnvVars`
   - `generateMigrationScript`
2. Create `internal/cmd/migrate/roster_to_knossos_test.go` with unit tests for pure functions.
3. Run `go test ./internal/cmd/migrate/...`.

### Phase 3: Filesystem Functions (30 min)

1. Add `migrateUserManifests` and `migrateCEMManifest` to `roster_to_knossos.go`.
2. Add filesystem integration tests using `t.TempDir()`.
3. Run tests.

### Phase 4: Cobra Wiring (20 min)

1. Create `internal/cmd/migrate/migrate.go` with `NewMigrateCmd`.
2. Add `newRosterToKnossosCmd` to register the subcommand.
3. Register in `internal/cmd/root/root.go`.
4. Build `ari` binary and verify `ari migrate roster-to-knossos --help`.

### Phase 5: End-to-End Verification (15 min)

1. Run `ari migrate roster-to-knossos` in the knossos repo itself (dry-run).
2. Verify output matches expected format.
3. Run `ari migrate roster-to-knossos --apply` and verify manifests updated.
4. Run again to verify idempotency (all skipped).
5. Run full test suite: `go test ./...`.

---

## Quality Gate Criteria

- [ ] `ari migrate roster-to-knossos --help` displays correct usage
- [ ] Dry-run mode shows changes without mutating files
- [ ] `--apply` rewrites user manifests correctly
- [ ] `--apply` rewrites CEM manifest correctly
- [ ] Backups created with `.roster-backup` extension
- [ ] `--no-backup` skips backup creation
- [ ] Running twice produces idempotent result (second run: all skipped)
- [ ] Invalid JSON manifests produce error without crashing
- [ ] Missing manifest files are skipped gracefully
- [ ] `--generate-script` outputs valid shell script
- [ ] `--skip-user` skips user manifests
- [ ] `--skip-project` skips CEM manifest
- [ ] JSON output (`-o json`) is valid JSON matching `RosterMigrateOutput` schema
- [ ] `go test ./internal/cmd/migrate/...` passes all tests
- [ ] `go test ./...` passes (no regressions)
- [ ] Binary builds cleanly: `CGO_ENABLED=0 go build ./cmd/ari`

---

## Notes for Integration Engineer

1. **Start with pure functions**: `rewriteUserManifestBytes` and `rewriteCEMManifestBytes` are the core. Write and test these first. Everything else is wiring.

2. **Use `map[string]interface{}` for JSON round-tripping**: Do NOT define Go structs for manifest parsing in this command. The manifests have varying structures (user vs CEM) and the rewrite needs to preserve all fields including unknown ones. Generic map traversal is the correct approach.

3. **Atomic writes pattern**: Follow the pattern from `internal/usersync/manifest.go` lines 147-155: write to `.tmp`, then `os.Rename`. This prevents partial writes.

4. **`NeedsProject` must be false**: The migrate command works at user-level by default (`~/.claude`). It should NOT fail if there is no `.claude/` directory in the current working directory. Use `common.SetNeedsProject(cmd, false, true)`.

5. **The existing `MigrateOutput` in output.go is for session migration**: Name the new struct `RosterMigrateOutput` to avoid collision. The existing `MigrateOutput` has `Migrated`, `Skipped`, `Failed` slices of session-specific types.

6. **Script generation is stdout by default**: When `--generate-script` is used without `--script-file`, the script goes to stdout. The migration report goes to stderr in this case (so the script can be piped to a file: `ari migrate roster-to-knossos --generate-script > migrate.sh`). When `--script-file` is specified, both report and script destination are separate.

7. **macOS vs Linux sed**: The generated script must handle both platforms. Use `uname -s` detection in the generated script, not in the Go code.

8. **Existing SourceType constants are already correct**: `internal/usersync/usersync.go` lines 59-61 already define `SourceKnossos = "knossos"` and `SourceDiverged = "knossos-diverged"`. The migration aligns old manifests with these values. No changes to usersync needed.

9. **CEM manifest at `.claude/.cem/manifest.json`**: The actual file in this repo shows `"roster": {...}` as a top-level key and `"source": "roster"` in managed_files entries. Both need rewriting. The `"team.roster_path"` field also needs renaming to `"team.knossos_path"`.

---

## Artifact Attestation

| Source File | Operation |
|-------------|-----------|
| `/Users/tomtenuta/Code/roster/internal/cmd/session/migrate.go` | Read (migration pattern reference) |
| `/Users/tomtenuta/Code/roster/internal/cmd/session/session.go` | Read (command registration pattern) |
| `/Users/tomtenuta/Code/roster/internal/cmd/root/root.go` | Read (root command registration) |
| `/Users/tomtenuta/Code/roster/internal/usersync/manifest.go` | Read (manifest JSON structure) |
| `/Users/tomtenuta/Code/roster/internal/usersync/usersync.go` | Read (SourceType constants, Syncer pattern) |
| `/Users/tomtenuta/Code/roster/internal/paths/paths.go` | Read (user manifest paths, Resolver pattern) |
| `/Users/tomtenuta/Code/roster/internal/config/home.go` | Read (KNOSSOS_HOME resolution) |
| `/Users/tomtenuta/Code/roster/internal/output/output.go` | Read (output struct patterns, Textable interface) |
| `/Users/tomtenuta/Code/roster/internal/cmd/common/context.go` | Read (BaseContext, SessionContext) |
| `/Users/tomtenuta/Code/roster/internal/cmd/common/annotations.go` | Read (NeedsProject annotation) |
| `/Users/tomtenuta/Code/roster/internal/cmd/manifest/manifest.go` | Read (non-session command group pattern) |
| `/Users/tomtenuta/Code/roster/internal/errors/errors.go` | Read (error codes, constructors) |
| `/Users/tomtenuta/Code/roster/.claude/.cem/manifest.json` | Read (real CEM manifest structure) |
| `/Users/tomtenuta/Code/roster/docs/ecosystem/CONTEXT-DESIGN-knossos-home-migration.md` | Read (context design format reference) |

---

## Handoff to Integration Engineer

This Context Design is complete and ready for implementation. No unresolved design decisions remain.

Implementation order:
1. Add `RosterMigrateOutput` to `internal/output/output.go`
2. Create `internal/cmd/migrate/roster_to_knossos.go` with pure functions + tests
3. Create `internal/cmd/migrate/migrate.go` with Cobra wiring
4. Register in `internal/cmd/root/root.go`
5. Build, test, verify end-to-end

The core complexity is in `rewriteUserManifestBytes` and `rewriteCEMManifestBytes` -- both are pure functions operating on byte slices, fully testable without filesystem access.
