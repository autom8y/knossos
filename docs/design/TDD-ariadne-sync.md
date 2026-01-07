# TDD: Ariadne Sync Domain

> Technical Design Document for the sync domain of the Ariadne Go CLI

**Status**: Draft
**Author**: Architect Agent
**Date**: 2026-01-04
**PRD**: docs/requirements/PRD-ariadne.md
**Reference**: docs/design/TDD-ariadne-session.md (Phase 1), docs/design/TDD-ariadne-rite.md (Phase 2), docs/design/TDD-ariadne-manifest.md (Phase 3)

---

## 1. Overview

This Technical Design Document specifies the implementation of the **sync domain** for Ariadne (`ari`), the Go binary replacement for the roster bash script harness. The sync domain encompasses 7 commands that manage synchronization of Claude Code resources -- rites, skills, and hooks -- between local projects and remote sources. This is Phase 4 and the final domain for Ariadne v1.0.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-ariadne.md` (Sections 2.1, 4.1) |
| Spike | `docs/spikes/SPIKE-ariadne-go-cli-architecture.md` |
| Session TDD | `docs/design/TDD-ariadne-session.md` |
| Team TDD | `docs/design/TDD-ariadne-rite.md` |
| Manifest TDD | `docs/design/TDD-ariadne-manifest.md` |
| Error Taxonomy | `ariadne/internal/errors/errors.go` |
| Merge Logic | `ariadne/internal/manifest/merge.go` |
| Path Resolution | `ariadne/internal/paths/paths.go` |
| Current Implementation | `roster-sync` (bash script) |

### 1.2 Scope

**In Scope**:
- 7 sync commands: status, pull, push, diff, resolve, history, reset
- Internal packages: `cmd/sync/`, `sync/`
- Error handling with exit codes per PRD Section 5.1
- Tracking state in `.claude/sync/state.json`
- Audit trail in `.claude/sync/history.json`
- Three-way merge for conflict resolution (reuse manifest/merge.go)
- Remote source resolution (GitHub raw URLs, local paths, git refs)

**Out of Scope**:
- Team pack authoring (forge responsibility)
- Agent manifest operations (team domain handles AGENT_MANIFEST.json)
- Schema validation (manifest domain handles via `ari manifest validate`)
- Remote registry/discovery service (future enhancement)

### 1.3 Design Goals

1. **Deterministic Sync**: Same inputs produce same outputs; no hidden state
2. **Explicit Conflicts**: Three-way merge with clear conflict markers
3. **Audit Trail**: All sync operations logged for traceability
4. **Offline Resilience**: Track state locally; handle network failures gracefully
5. **Reuse Infrastructure**: Leverage manifest domain's merge logic
6. **Checksum Integrity**: SHA256 checksums for change detection

---

## 2. Architecture

### 2.1 Package Structure

```
ariadne/
├── internal/
│   ├── cmd/
│   │   └── sync/
│   │       ├── sync.go               # Parent command registration
│   │       ├── status.go             # ari sync status
│   │       ├── pull.go               # ari sync pull
│   │       ├── push.go               # ari sync push
│   │       ├── diff.go               # ari sync diff
│   │       ├── resolve.go            # ari sync resolve
│   │       ├── history.go            # ari sync history
│   │       └── reset.go              # ari sync reset
│   ├── sync/
│   │   ├── state.go                  # Sync state management
│   │   ├── tracker.go                # Resource tracking and checksums
│   │   ├── remote.go                 # Remote source resolution
│   │   ├── pull.go                   # Pull logic
│   │   ├── push.go                   # Push logic
│   │   ├── diff.go                   # Diff computation
│   │   ├── conflict.go               # Conflict detection and resolution
│   │   └── history.go                # Audit trail operations
│   ├── paths/
│   │   └── sync.go                   # Sync-specific path resolution (extends existing)
│   └── output/
│       └── sync.go                   # Sync-specific output structures (extends existing)
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────────┐
                    │  internal/cmd/sync/sync.go      │
                    │  (7 commands)                   │
                    └─────────────┬───────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         v                        v                        v
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ internal/sync/  │     │ internal/paths/ │     │ internal/output/│
│ (business logic)│     │ (extended)      │     │ (extended)      │
└────────┬────────┘     └─────────────────┘     └─────────────────┘
         │
         ├────────────────────────┐
         │                        │
         v                        v
┌─────────────────┐     ┌─────────────────────────────┐
│ internal/       │     │ net/http, os/exec (git)     │
│ manifest/merge  │     │ (remote fetching)           │
│ (reused)        │     └─────────────────────────────┘
└─────────────────┘
         │
         v
┌─────────────────────────────────────────────────────────────────┐
│  Filesystem: .claude/sync/state.json, .claude/sync/history.json,│
│  rites/, ~/.claude/skills/, .claude/hooks/                       │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 Key Concepts

#### Sync Targets

Resources that can be synced:

| Target | Local Path | Remote Source |
|--------|------------|---------------|
| Team Packs | `rites/{name}/` | GitHub repo, local path |
| Skills | `~/.claude/skills/` | GitHub repo, local path |
| Hooks | `.claude/hooks/` | GitHub repo, local path |

#### Remote Sources

```yaml
# .claude/sync/config.yaml (optional per-project config)
remotes:
  roster:
    url: "https://github.com/autom8y/roster"
    branch: main
    paths:
      teams: rites/
      skills: .claude/skills/
      hooks: .claude/hooks/
  custom-skills:
    url: "https://github.com/org/claude-skills"
    branch: main
    paths:
      skills: skills/
```

#### Tracking State

```
.claude/sync/
├── state.json                 # Current sync state (checksums, timestamps)
├── history.json               # Audit log of sync operations
└── conflicts/                 # Pending conflict files for resolution
    └── {resource}-{timestamp}.conflict
```

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Modifies State |
|---------|-------------|----------------|
| `status` | Show sync status for all tracked paths | No |
| `pull` | Pull remote changes with conflict detection | Yes |
| `push` | Push local changes to remote | Yes (remote) |
| `diff` | Show local vs remote differences | No |
| `resolve` | Resolve sync conflicts | Yes |
| `history` | Show sync history/audit log | No |
| `reset` | Reset sync state (dangerous) | Yes |

### 3.2 Command: `ari sync status`

Shows the synchronization status for all tracked resources.

**Signature**:
```
ari sync status [--resource=TYPE] [--verbose]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resource` | `-r` | string | (all) | Filter by resource type: teams, skills, hooks |
| `--verbose` | `-v` | bool | false | Show detailed per-file status |

**Output (JSON)**:
```json
{
  "synced_at": "2026-01-04T18:00:00Z",
  "status": "changes_detected",
  "resources": {
    "teams": {
      "status": "synced",
      "tracked": 4,
      "local_path": "rites/",
      "remote": "https://github.com/autom8y/roster",
      "last_sync": "2026-01-04T18:00:00Z",
      "items": [
        {"name": "10x-dev", "status": "synced", "checksum": "sha256:abc123..."},
        {"name": "rnd", "status": "synced", "checksum": "sha256:def456..."},
        {"name": "security", "status": "local_modified", "checksum": "sha256:789abc..."},
        {"name": "sre", "status": "synced", "checksum": "sha256:012def..."}
      ]
    },
    "skills": {
      "status": "remote_ahead",
      "tracked": 12,
      "local_path": "~/.claude/skills/",
      "remote": "https://github.com/autom8y/roster",
      "last_sync": "2026-01-04T17:00:00Z",
      "pending_updates": 3
    },
    "hooks": {
      "status": "synced",
      "tracked": 5,
      "local_path": ".claude/hooks/",
      "remote": "https://github.com/autom8y/roster",
      "last_sync": "2026-01-04T18:00:00Z"
    }
  },
  "conflicts": [],
  "has_conflicts": false
}
```

**Status Values**:

| Status | Description |
|--------|-------------|
| `synced` | Local matches remote |
| `local_modified` | Local has uncommitted changes |
| `remote_ahead` | Remote has new changes available |
| `conflict` | Both local and remote modified (needs resolution) |
| `untracked` | Resource exists locally but not tracked |
| `missing` | Tracked resource missing locally |

**Output (text)**:
```
Sync Status (as of 2026-01-04T18:00:00Z)

Teams (rites/)
  Status: SYNCED
  Tracked: 4 teams
  Remote: https://github.com/autom8y/roster

Skills (~/.claude/skills/)
  Status: REMOTE_AHEAD (3 updates available)
  Tracked: 12 skills
  Remote: https://github.com/autom8y/roster
  Run 'ari sync pull --resource=skills' to update

Hooks (.claude/hooks/)
  Status: SYNCED
  Tracked: 5 hooks
  Remote: https://github.com/autom8y/roster

No conflicts.
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Status retrieved successfully |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Reads from `.claude/sync/state.json`
- Computes checksums of local files for comparison
- Does not fetch from remote (use `--fetch` for that behavior)
- Groups resources by type (teams, skills, hooks)

### 3.3 Command: `ari sync pull`

Pulls changes from remote sources with conflict detection.

**Signature**:
```
ari sync pull [--resource=TYPE] [--remote=NAME] [--dry-run] [--force]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resource` | `-r` | string | (all) | Resource type to pull: teams, skills, hooks |
| `--remote` | | string | (default) | Remote source name from config |
| `--dry-run` | | bool | false | Preview changes without applying |
| `--force` | `-f` | bool | false | Overwrite local changes (destructive) |

**Output (JSON)**:
```json
{
  "pulled_at": "2026-01-04T19:00:00Z",
  "dry_run": false,
  "remote": "roster",
  "changes": {
    "updated": [
      {
        "resource": "skills",
        "name": "commit-ref",
        "path": "~/.claude/skills/commit-ref.md",
        "previous_checksum": "sha256:old123...",
        "new_checksum": "sha256:new456..."
      },
      {
        "resource": "skills",
        "name": "pr-ref",
        "path": "~/.claude/skills/pr-ref.md",
        "previous_checksum": "sha256:old789...",
        "new_checksum": "sha256:newabc..."
      }
    ],
    "added": [
      {
        "resource": "skills",
        "name": "hotfix-ref",
        "path": "~/.claude/skills/hotfix-ref.md",
        "new_checksum": "sha256:def789..."
      }
    ],
    "deleted": [],
    "conflicts": []
  },
  "summary": {
    "updated": 2,
    "added": 1,
    "deleted": 0,
    "conflicts": 0
  }
}
```

**Output (with conflicts, JSON)**:
```json
{
  "pulled_at": "2026-01-04T19:00:00Z",
  "dry_run": false,
  "remote": "roster",
  "changes": {
    "updated": [],
    "added": [],
    "deleted": [],
    "conflicts": [
      {
        "resource": "teams",
        "name": "10x-dev",
        "path": "rites/10x-dev/workflow.yaml",
        "conflict_file": ".claude/sync/conflicts/workflow.yaml-20260104-190000.conflict",
        "base_checksum": "sha256:base123...",
        "local_checksum": "sha256:local456...",
        "remote_checksum": "sha256:remote789..."
      }
    ]
  },
  "summary": {
    "updated": 0,
    "added": 0,
    "deleted": 0,
    "conflicts": 1
  },
  "has_conflicts": true,
  "resolution_hint": "Run 'ari sync resolve' to resolve conflicts"
}
```

**Output (text)**:
```
Pulling from roster...

Updated:
  [skills] commit-ref.md
  [skills] pr-ref.md

Added:
  [skills] hotfix-ref.md

Summary: 2 updated, 1 added, 0 deleted, 0 conflicts

Sync state updated.
```

**Output (text, with conflicts)**:
```
Pulling from roster...

Conflicts:
  [teams] 10x-dev/workflow.yaml
    Local and remote both modified since last sync.
    Conflict file: .claude/sync/conflicts/workflow.yaml-20260104-190000.conflict

Summary: 0 updated, 0 added, 0 deleted, 1 conflict

Run 'ari sync resolve' to resolve conflicts before continuing.
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Pull completed successfully (no conflicts) |
| 8 | Pull completed with conflicts (MERGE_CONFLICT) |
| 6 | Remote not found or unreachable |
| 7 | Permission denied (cannot write local files) |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Fetches checksums from remote first
- Compares local, remote, and base (last synced) checksums
- Uses three-way merge from manifest domain for conflict detection
- Creates conflict files in `.claude/sync/conflicts/`
- Updates `.claude/sync/state.json` after successful pull
- Logs operation to `.claude/sync/history.json`
- `--force` skips conflict detection and overwrites local

### 3.4 Command: `ari sync push`

Pushes local changes to remote (for writable remotes).

**Signature**:
```
ari sync push [--resource=TYPE] [--remote=NAME] [--dry-run] [--message=TEXT]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resource` | `-r` | string | (all) | Resource type to push: teams, skills, hooks |
| `--remote` | | string | (default) | Remote source name from config |
| `--dry-run` | | bool | false | Preview changes without applying |
| `--message` | `-m` | string | "ari sync push" | Commit message for push |

**Output (JSON)**:
```json
{
  "pushed_at": "2026-01-04T19:30:00Z",
  "dry_run": false,
  "remote": "roster",
  "changes": {
    "pushed": [
      {
        "resource": "teams",
        "name": "10x-dev",
        "path": "rites/10x-dev/workflow.yaml",
        "checksum": "sha256:new123..."
      }
    ]
  },
  "summary": {
    "pushed": 1
  },
  "commit": "abc123def456",
  "message": "Update 10x-dev workflow"
}
```

**Output (text)**:
```
Pushing to roster...

Pushed:
  [teams] 10x-dev/workflow.yaml

Summary: 1 file pushed

Commit: abc123def456
Message: Update 10x-dev workflow
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Push completed successfully |
| 6 | Remote not found or not writable |
| 7 | Permission denied (no push access) |
| 8 | Push rejected (remote ahead, pull first) |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Only works with writable remotes (git repos with push access)
- Requires local git configuration for authentication
- Creates commit with specified message
- Updates state after successful push
- Logs operation to history

### 3.5 Command: `ari sync diff`

Shows differences between local and remote versions.

**Signature**:
```
ari sync diff [--resource=TYPE] [--name=NAME] [--format=FORMAT]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resource` | `-r` | string | (all) | Resource type: teams, skills, hooks |
| `--name` | `-n` | string | (all) | Specific resource name |
| `--format` | `-f` | string | `unified` | Output format: unified, json, side-by-side |

**Output (JSON)**:
```json
{
  "compared_at": "2026-01-04T19:00:00Z",
  "diffs": [
    {
      "resource": "teams",
      "name": "10x-dev",
      "file": "workflow.yaml",
      "local_path": "rites/10x-dev/workflow.yaml",
      "status": "modified",
      "changes": [
        {
          "path": "$.entry_point.agent",
          "type": "modified",
          "local_value": "requirements-analyst",
          "remote_value": "architect"
        }
      ]
    }
  ],
  "summary": {
    "modified": 1,
    "added": 0,
    "deleted": 0,
    "unchanged": 15
  }
}
```

**Output (unified, text)**:
```
--- local: rites/10x-dev/workflow.yaml
+++ remote: roster:rites/10x-dev/workflow.yaml

@@ entry_point @@
  entry_point:
-   agent: requirements-analyst
+   agent: architect

1 file modified, 0 added, 0 deleted
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Diff completed successfully (no differences) |
| 1 | Diff completed, differences detected |
| 6 | Resource not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Uses manifest domain's diff logic for structured comparison
- Supports git ref syntax for remote (`roster:path/to/file`)
- Exit code 1 indicates differences (useful for scripting)

### 3.6 Command: `ari sync resolve`

Resolves pending sync conflicts.

**Signature**:
```
ari sync resolve [--strategy=STRATEGY] [--conflict=FILE] [--all]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--strategy` | `-s` | string | (interactive) | Resolution strategy: ours, theirs, merge |
| `--conflict` | `-c` | string | | Specific conflict file to resolve |
| `--all` | `-a` | bool | false | Resolve all conflicts with strategy |

**Resolution Strategies**:

| Strategy | Behavior |
|----------|----------|
| `ours` | Keep local version, discard remote |
| `theirs` | Accept remote version, discard local |
| `merge` | Attempt three-way merge (may still conflict) |
| (interactive) | Prompt for each conflict |

**Output (JSON)**:
```json
{
  "resolved_at": "2026-01-04T20:00:00Z",
  "conflicts_resolved": [
    {
      "resource": "teams",
      "name": "10x-dev",
      "file": "workflow.yaml",
      "strategy": "ours",
      "result_checksum": "sha256:resolved123..."
    }
  ],
  "summary": {
    "resolved": 1,
    "remaining": 0
  }
}
```

**Output (text)**:
```
Resolving conflicts...

Resolved:
  [teams] 10x-dev/workflow.yaml (strategy: ours)

Summary: 1 resolved, 0 remaining

Sync state updated.
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | All conflicts resolved |
| 1 | Some conflicts remain |
| 6 | Conflict file not found |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Reads conflict files from `.claude/sync/conflicts/`
- Uses manifest domain's merge logic for `merge` strategy
- Removes conflict files after resolution
- Updates state and history after resolution

### 3.7 Command: `ari sync history`

Shows sync operation history.

**Signature**:
```
ari sync history [--limit=N] [--resource=TYPE] [--since=TIMESTAMP]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--limit` | `-n` | int | 20 | Maximum entries to show |
| `--resource` | `-r` | string | (all) | Filter by resource type |
| `--since` | | string | | Only entries after timestamp |

**Output (JSON)**:
```json
{
  "entries": [
    {
      "timestamp": "2026-01-04T19:30:00Z",
      "operation": "push",
      "remote": "roster",
      "resource": "teams",
      "details": {
        "files_pushed": 1,
        "commit": "abc123def456"
      }
    },
    {
      "timestamp": "2026-01-04T19:00:00Z",
      "operation": "pull",
      "remote": "roster",
      "resource": "skills",
      "details": {
        "files_updated": 2,
        "files_added": 1,
        "conflicts": 0
      }
    },
    {
      "timestamp": "2026-01-04T18:00:00Z",
      "operation": "pull",
      "remote": "roster",
      "resource": "all",
      "details": {
        "files_updated": 0,
        "files_added": 0,
        "conflicts": 0
      }
    }
  ],
  "total": 3,
  "filtered": false
}
```

**Output (text)**:
```
TIMESTAMP                OPERATION  REMOTE   RESOURCE  DETAILS
2026-01-04T19:30:00Z     push       roster   teams     1 pushed, commit: abc123d
2026-01-04T19:00:00Z     pull       roster   skills    2 updated, 1 added
2026-01-04T18:00:00Z     pull       roster   all       up to date

Total: 3 entries
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | History retrieved successfully |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Reads from `.claude/sync/history.json`
- JSONL format for efficient append-only writes
- Supports filtering by resource and timestamp

### 3.8 Command: `ari sync reset`

Resets sync state (dangerous operation).

**Signature**:
```
ari sync reset [--resource=TYPE] [--hard] [--confirm]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--resource` | `-r` | string | (all) | Resource type to reset |
| `--hard` | | bool | false | Also delete local files (very dangerous) |
| `--confirm` | | bool | false | Confirm destructive operation |

**Output (JSON)**:
```json
{
  "reset_at": "2026-01-04T21:00:00Z",
  "resource": "all",
  "hard": false,
  "cleared": {
    "state_entries": 21,
    "history_entries": 15,
    "conflict_files": 2
  }
}
```

**Output (text)**:
```
Resetting sync state...

Cleared:
  State entries: 21
  History entries: 15
  Conflict files: 2

Sync state has been reset.
Next pull will treat all resources as new.
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Reset completed successfully |
| 2 | Missing --confirm flag for destructive operation |
| 9 | No .claude/ directory found |

**Implementation Notes**:
- Requires `--confirm` flag to prevent accidental data loss
- `--hard` also deletes local files (requires double confirmation)
- Backs up state before reset (recoverable for 7 days)
- Logs reset operation to history before clearing

---

## 4. Error Handling

### 4.1 Error Code Taxonomy

Extending PRD Section 5.1 with sync-domain-specific codes:

| Code | Exit | Name | Description |
|------|------|------|-------------|
| `SUCCESS` | 0 | Success | Operation completed successfully |
| `GENERAL_ERROR` | 1 | General Error | Unspecified error |
| `USAGE_ERROR` | 2 | Usage Error | Invalid arguments or flags |
| `REMOTE_NOT_FOUND` | 6 | Remote Not Found | Remote source not configured or unreachable |
| `RESOURCE_NOT_FOUND` | 6 | Resource Not Found | Tracked resource missing |
| `PERMISSION_DENIED` | 7 | Permission Denied | Cannot write local or push to remote |
| `MERGE_CONFLICT` | 8 | Merge Conflict | Sync has unresolved conflicts |
| `PROJECT_NOT_FOUND` | 9 | Project Not Found | No .claude/ directory found |
| `SYNC_STATE_CORRUPT` | 16 | Sync State Corrupt | state.json is invalid or corrupt |
| `REMOTE_REJECTED` | 17 | Remote Rejected | Push rejected by remote |
| `NETWORK_ERROR` | 18 | Network Error | Failed to fetch from remote |

### 4.2 Error Response Structure

All errors follow the PRD Section 4.4 contract:

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

Example:
```json
{
  "error": {
    "code": "MERGE_CONFLICT",
    "message": "Sync pull has unresolved conflicts",
    "details": {
      "conflict_count": 2,
      "conflicts": [
        "rites/10x-dev/workflow.yaml",
        "skills/commit-ref.md"
      ],
      "resolution_hint": "Run 'ari sync resolve' to resolve conflicts"
    }
  }
}
```

### 4.3 Error Constructors

New error constructors for sync domain:

```go
// ErrRemoteNotFound returns an error for missing remote.
func ErrRemoteNotFound(remoteName string) *Error {
    return NewWithDetails(CodeRemoteNotFound,
        fmt.Sprintf("Remote not found: %s", remoteName),
        map[string]interface{}{"remote": remoteName})
}

// ErrSyncConflict returns an error for sync conflicts.
func ErrSyncConflict(conflicts []string) *Error {
    return NewWithDetails(CodeMergeConflict,
        "Sync pull has unresolved conflicts",
        map[string]interface{}{
            "conflict_count":   len(conflicts),
            "conflicts":        conflicts,
            "resolution_hint": "Run 'ari sync resolve' to resolve conflicts",
        })
}

// ErrSyncStateCorrupt returns an error for corrupt sync state.
func ErrSyncStateCorrupt(path string, reason string) *Error {
    return NewWithDetails(CodeSyncStateCorrupt,
        fmt.Sprintf("Sync state corrupt: %s", reason),
        map[string]interface{}{
            "path":   path,
            "reason": reason,
        })
}

// ErrNetworkError returns an error for network failures.
func ErrNetworkError(url string, cause error) *Error {
    details := map[string]interface{}{"url": url}
    if cause != nil {
        details["cause"] = cause.Error()
    }
    return NewWithDetails(CodeNetworkError,
        fmt.Sprintf("Network error fetching %s", url),
        details)
}

// ErrRemoteRejected returns an error when push is rejected.
func ErrRemoteRejected(remote string, reason string) *Error {
    return NewWithDetails(CodeRemoteRejected,
        fmt.Sprintf("Push rejected by %s: %s", remote, reason),
        map[string]interface{}{
            "remote": remote,
            "reason": reason,
        })
}
```

---

## 5. Data Model

### 5.1 Sync State (`state.json`)

Tracks the synchronization state of all resources:

```json
{
  "schema_version": "1.0",
  "last_sync": "2026-01-04T18:00:00Z",
  "remotes": {
    "roster": {
      "url": "https://github.com/autom8y/roster",
      "branch": "main",
      "last_fetched": "2026-01-04T18:00:00Z",
      "commit": "abc123def456..."
    }
  },
  "resources": {
    "teams": {
      "10x-dev": {
        "local_checksum": "sha256:abc123...",
        "remote_checksum": "sha256:abc123...",
        "base_checksum": "sha256:abc123...",
        "last_synced": "2026-01-04T18:00:00Z",
        "status": "synced",
        "files": {
          "workflow.yaml": {
            "local_checksum": "sha256:def456...",
            "remote_checksum": "sha256:def456...",
            "base_checksum": "sha256:def456..."
          },
          "agents/architect.md": {
            "local_checksum": "sha256:789abc...",
            "remote_checksum": "sha256:789abc...",
            "base_checksum": "sha256:789abc..."
          }
        }
      }
    },
    "skills": {
      "commit-ref": {
        "local_checksum": "sha256:skill123...",
        "remote_checksum": "sha256:skill456...",
        "base_checksum": "sha256:skill123...",
        "last_synced": "2026-01-04T17:00:00Z",
        "status": "remote_ahead"
      }
    },
    "hooks": {}
  },
  "conflicts": []
}
```

### 5.2 Sync History (`history.json`)

JSONL format for append-only audit trail:

```jsonl
{"timestamp":"2026-01-04T18:00:00Z","operation":"pull","remote":"roster","resource":"all","result":"success","details":{"updated":0,"added":0,"conflicts":0}}
{"timestamp":"2026-01-04T19:00:00Z","operation":"pull","remote":"roster","resource":"skills","result":"success","details":{"updated":2,"added":1,"conflicts":0}}
{"timestamp":"2026-01-04T19:30:00Z","operation":"push","remote":"roster","resource":"teams","result":"success","details":{"pushed":1,"commit":"abc123"}}
```

### 5.3 Conflict File Format

Conflict files stored in `.claude/sync/conflicts/`:

```json
{
  "created_at": "2026-01-04T19:00:00Z",
  "resource": "teams",
  "name": "10x-dev",
  "file": "workflow.yaml",
  "local_path": "rites/10x-dev/workflow.yaml",
  "remote_url": "https://github.com/autom8y/roster/blob/main/rites/10x-dev/workflow.yaml",
  "base": {
    "checksum": "sha256:base123...",
    "content": "... base content ..."
  },
  "ours": {
    "checksum": "sha256:local456...",
    "content": "... local content ..."
  },
  "theirs": {
    "checksum": "sha256:remote789...",
    "content": "... remote content ..."
  }
}
```

### 5.4 Resource Entry Structure

```go
// ResourceEntry represents a tracked resource.
type ResourceEntry struct {
    LocalChecksum  string            `json:"local_checksum"`
    RemoteChecksum string            `json:"remote_checksum"`
    BaseChecksum   string            `json:"base_checksum"`
    LastSynced     time.Time         `json:"last_synced"`
    Status         ResourceStatus    `json:"status"`
    Files          map[string]FileEntry `json:"files,omitempty"`
}

// FileEntry represents a tracked file within a resource.
type FileEntry struct {
    LocalChecksum  string `json:"local_checksum"`
    RemoteChecksum string `json:"remote_checksum"`
    BaseChecksum   string `json:"base_checksum"`
}

// ResourceStatus represents the sync status.
type ResourceStatus string

const (
    StatusSynced        ResourceStatus = "synced"
    StatusLocalModified ResourceStatus = "local_modified"
    StatusRemoteAhead   ResourceStatus = "remote_ahead"
    StatusConflict      ResourceStatus = "conflict"
    StatusUntracked     ResourceStatus = "untracked"
    StatusMissing       ResourceStatus = "missing"
)
```

---

## 6. Internal Package Design

### 6.1 Package: `internal/sync`

Core sync operations, independent of CLI.

```go
package sync

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
    "time"
)

// State represents the sync state.
type State struct {
    SchemaVersion string                          `json:"schema_version"`
    LastSync      time.Time                       `json:"last_sync"`
    Remotes       map[string]RemoteState          `json:"remotes"`
    Resources     map[string]map[string]ResourceEntry `json:"resources"`
    Conflicts     []ConflictEntry                 `json:"conflicts"`
}

// RemoteState tracks remote source state.
type RemoteState struct {
    URL         string    `json:"url"`
    Branch      string    `json:"branch"`
    LastFetched time.Time `json:"last_fetched"`
    Commit      string    `json:"commit"`
}

// ConflictEntry represents a pending conflict.
type ConflictEntry struct {
    Resource    string    `json:"resource"`
    Name        string    `json:"name"`
    File        string    `json:"file"`
    ConflictFile string   `json:"conflict_file"`
    CreatedAt   time.Time `json:"created_at"`
}

// Load reads sync state from path.
func LoadState(path string) (*State, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return NewEmptyState(), nil
        }
        return nil, err
    }

    var state State
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, errors.ErrSyncStateCorrupt(path, err.Error())
    }

    return &state, nil
}

// Save writes sync state to path.
func (s *State) Save(path string) error {
    data, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

// NewEmptyState creates a new empty sync state.
func NewEmptyState() *State {
    return &State{
        SchemaVersion: "1.0",
        Remotes:       make(map[string]RemoteState),
        Resources:     make(map[string]map[string]ResourceEntry),
        Conflicts:     []ConflictEntry{},
    }
}
```

### 6.2 Package: `internal/sync/tracker`

Resource tracking and checksum computation.

```go
package sync

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
    "path/filepath"
)

// Tracker handles resource tracking and checksum computation.
type Tracker struct {
    state *State
}

// NewTracker creates a new tracker with the given state.
func NewTracker(state *State) *Tracker {
    return &Tracker{state: state}
}

// ComputeChecksum calculates SHA256 checksum of a file.
func ComputeChecksum(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }

    return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeDirectoryChecksum calculates checksum of all files in directory.
func ComputeDirectoryChecksum(dir string) (string, map[string]string, error) {
    files := make(map[string]string)
    h := sha256.New()

    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }

        relPath, _ := filepath.Rel(dir, path)
        checksum, err := ComputeChecksum(path)
        if err != nil {
            return err
        }

        files[relPath] = checksum
        h.Write([]byte(relPath + ":" + checksum + "\n"))
        return nil
    })

    if err != nil {
        return "", nil, err
    }

    return "sha256:" + hex.EncodeToString(h.Sum(nil)), files, nil
}

// DetermineStatus determines the sync status of a resource.
func (t *Tracker) DetermineStatus(entry ResourceEntry) ResourceStatus {
    switch {
    case entry.LocalChecksum == "" && entry.RemoteChecksum != "":
        return StatusMissing
    case entry.LocalChecksum != "" && entry.RemoteChecksum == "":
        return StatusUntracked
    case entry.LocalChecksum == entry.RemoteChecksum:
        return StatusSynced
    case entry.LocalChecksum != entry.BaseChecksum && entry.RemoteChecksum == entry.BaseChecksum:
        return StatusLocalModified
    case entry.LocalChecksum == entry.BaseChecksum && entry.RemoteChecksum != entry.BaseChecksum:
        return StatusRemoteAhead
    default:
        return StatusConflict
    }
}
```

### 6.3 Package: `internal/sync/remote`

Remote source resolution and fetching.

```go
package sync

import (
    "io"
    "net/http"
    "os"
    "os/exec"
    "strings"
)

// RemoteSource represents a configured remote.
type RemoteSource struct {
    Name   string
    URL    string
    Branch string
    Paths  map[string]string // resource type -> path
}

// RemoteFetcher handles fetching from remote sources.
type RemoteFetcher struct {
    client *http.Client
}

// NewRemoteFetcher creates a new remote fetcher.
func NewRemoteFetcher() *RemoteFetcher {
    return &RemoteFetcher{
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// FetchFile fetches a single file from remote.
func (f *RemoteFetcher) FetchFile(remote RemoteSource, path string) ([]byte, error) {
    if strings.HasPrefix(remote.URL, "https://github.com") {
        return f.fetchGitHubRaw(remote, path)
    }

    if isLocalPath(remote.URL) {
        return os.ReadFile(filepath.Join(remote.URL, path))
    }

    return nil, errors.ErrRemoteNotFound(remote.Name)
}

// fetchGitHubRaw fetches a file from GitHub raw URL.
func (f *RemoteFetcher) fetchGitHubRaw(remote RemoteSource, path string) ([]byte, error) {
    // Convert github.com URL to raw.githubusercontent.com
    // https://github.com/owner/repo -> https://raw.githubusercontent.com/owner/repo/branch/path
    url := convertToRawURL(remote.URL, remote.Branch, path)

    resp, err := f.client.Get(url)
    if err != nil {
        return nil, errors.ErrNetworkError(url, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, errors.ErrNetworkError(url, fmt.Errorf("status: %d", resp.StatusCode))
    }

    return io.ReadAll(resp.Body)
}

// FetchGitRef fetches using git commands.
func (f *RemoteFetcher) FetchGitRef(remote RemoteSource, ref, path string) ([]byte, error) {
    // Clone or fetch if needed, then use git show
    cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, path))
    output, err := cmd.Output()
    if err != nil {
        return nil, errors.ErrNetworkError(remote.URL, err)
    }
    return output, nil
}

// convertToRawURL converts GitHub URL to raw content URL.
func convertToRawURL(repoURL, branch, path string) string {
    // https://github.com/owner/repo -> https://raw.githubusercontent.com/owner/repo/branch/path
    repoURL = strings.TrimSuffix(repoURL, ".git")
    repoURL = strings.Replace(repoURL, "github.com", "raw.githubusercontent.com", 1)
    return fmt.Sprintf("%s/%s/%s", repoURL, branch, path)
}

// isLocalPath checks if URL is a local filesystem path.
func isLocalPath(url string) bool {
    return strings.HasPrefix(url, "/") || strings.HasPrefix(url, "./") || strings.HasPrefix(url, "~")
}
```

### 6.4 Package: `internal/sync/pull`

Pull operation logic.

```go
package sync

import (
    "github.com/autom8y/ariadne/internal/manifest"
)

// PullOptions configures pull behavior.
type PullOptions struct {
    Resource string
    Remote   string
    DryRun   bool
    Force    bool
}

// PullResult holds the result of a pull operation.
type PullResult struct {
    PulledAt    time.Time      `json:"pulled_at"`
    DryRun      bool           `json:"dry_run"`
    Remote      string         `json:"remote"`
    Changes     PullChanges    `json:"changes"`
    Summary     PullSummary    `json:"summary"`
    HasConflicts bool          `json:"has_conflicts"`
}

// PullChanges categorizes the changes from a pull.
type PullChanges struct {
    Updated   []ResourceChange `json:"updated"`
    Added     []ResourceChange `json:"added"`
    Deleted   []ResourceChange `json:"deleted"`
    Conflicts []ConflictInfo   `json:"conflicts"`
}

// ResourceChange represents a single resource change.
type ResourceChange struct {
    Resource         string `json:"resource"`
    Name             string `json:"name"`
    Path             string `json:"path"`
    PreviousChecksum string `json:"previous_checksum,omitempty"`
    NewChecksum      string `json:"new_checksum"`
}

// ConflictInfo provides details about a conflict.
type ConflictInfo struct {
    Resource       string `json:"resource"`
    Name           string `json:"name"`
    Path           string `json:"path"`
    ConflictFile   string `json:"conflict_file"`
    BaseChecksum   string `json:"base_checksum"`
    LocalChecksum  string `json:"local_checksum"`
    RemoteChecksum string `json:"remote_checksum"`
}

// Puller handles pull operations.
type Puller struct {
    state    *State
    fetcher  *RemoteFetcher
    tracker  *Tracker
    merger   *manifest.Merger
    paths    *paths.Resolver
}

// NewPuller creates a new puller.
func NewPuller(state *State, paths *paths.Resolver) *Puller {
    return &Puller{
        state:   state,
        fetcher: NewRemoteFetcher(),
        tracker: NewTracker(state),
        paths:   paths,
    }
}

// Pull performs the pull operation.
func (p *Puller) Pull(ctx context.Context, opts PullOptions) (*PullResult, error) {
    result := &PullResult{
        PulledAt: time.Now().UTC(),
        DryRun:   opts.DryRun,
        Remote:   opts.Remote,
        Changes: PullChanges{
            Updated:   []ResourceChange{},
            Added:     []ResourceChange{},
            Deleted:   []ResourceChange{},
            Conflicts: []ConflictInfo{},
        },
    }

    // Get remote configuration
    remote, err := p.getRemote(opts.Remote)
    if err != nil {
        return nil, err
    }

    // Determine which resources to pull
    resources := p.getResourcesToSync(opts.Resource)

    // Process each resource
    for _, resType := range resources {
        if err := p.pullResource(ctx, remote, resType, opts, result); err != nil {
            return nil, err
        }
    }

    // Update summary
    result.Summary = PullSummary{
        Updated:   len(result.Changes.Updated),
        Added:     len(result.Changes.Added),
        Deleted:   len(result.Changes.Deleted),
        Conflicts: len(result.Changes.Conflicts),
    }
    result.HasConflicts = result.Summary.Conflicts > 0

    // Update state if not dry run
    if !opts.DryRun {
        p.state.LastSync = result.PulledAt
        if err := p.state.Save(p.paths.SyncStateFile()); err != nil {
            return nil, err
        }

        // Log to history
        p.logHistory(result)
    }

    return result, nil
}

// pullResource pulls a single resource type.
func (p *Puller) pullResource(ctx context.Context, remote RemoteSource,
    resType string, opts PullOptions, result *PullResult) error {

    // Fetch remote checksums
    remoteEntries, err := p.fetchRemoteChecksums(remote, resType)
    if err != nil {
        return err
    }

    // Compare with local
    for name, remoteEntry := range remoteEntries {
        localEntry := p.getLocalEntry(resType, name)
        status := p.determineChangeType(localEntry, remoteEntry)

        switch status {
        case StatusRemoteAhead:
            if opts.DryRun {
                result.Changes.Updated = append(result.Changes.Updated,
                    ResourceChange{Resource: resType, Name: name})
            } else {
                if err := p.applyUpdate(remote, resType, name); err != nil {
                    return err
                }
                result.Changes.Updated = append(result.Changes.Updated,
                    ResourceChange{
                        Resource:         resType,
                        Name:             name,
                        Path:             p.getResourcePath(resType, name),
                        PreviousChecksum: localEntry.LocalChecksum,
                        NewChecksum:      remoteEntry.RemoteChecksum,
                    })
            }

        case StatusConflict:
            if opts.Force {
                // Force overwrite
                if !opts.DryRun {
                    if err := p.applyUpdate(remote, resType, name); err != nil {
                        return err
                    }
                }
                result.Changes.Updated = append(result.Changes.Updated,
                    ResourceChange{Resource: resType, Name: name})
            } else {
                // Create conflict file
                conflictFile, err := p.createConflictFile(remote, resType, name, localEntry, remoteEntry)
                if err != nil {
                    return err
                }
                result.Changes.Conflicts = append(result.Changes.Conflicts,
                    ConflictInfo{
                        Resource:       resType,
                        Name:           name,
                        Path:           p.getResourcePath(resType, name),
                        ConflictFile:   conflictFile,
                        BaseChecksum:   localEntry.BaseChecksum,
                        LocalChecksum:  localEntry.LocalChecksum,
                        RemoteChecksum: remoteEntry.RemoteChecksum,
                    })
            }

        // Handle added, deleted cases...
        }
    }

    return nil
}
```

### 6.5 Package: `internal/paths` (Extension)

Sync-specific path helpers.

```go
// Add to internal/paths/paths.go

// SyncDir returns the path to the sync directory.
func (r *Resolver) SyncDir() string {
    return filepath.Join(r.ClaudeDir(), "sync")
}

// SyncStateFile returns the path to the sync state file.
func (r *Resolver) SyncStateFile() string {
    return filepath.Join(r.SyncDir(), "state.json")
}

// SyncHistoryFile returns the path to the sync history file.
func (r *Resolver) SyncHistoryFile() string {
    return filepath.Join(r.SyncDir(), "history.json")
}

// SyncConflictsDir returns the path to the conflicts directory.
func (r *Resolver) SyncConflictsDir() string {
    return filepath.Join(r.SyncDir(), "conflicts")
}

// SyncConfigFile returns the path to the sync config file.
func (r *Resolver) SyncConfigFile() string {
    return filepath.Join(r.SyncDir(), "config.yaml")
}

// SkillsDir returns the user-level skills directory.
func (r *Resolver) SkillsDir() string {
    return filepath.Join(r.projectRoot, ".claude", "skills")
}

// UserSkillsDir returns the global user skills directory.
func UserSkillsDir() string {
    return filepath.Join(xdg.ConfigHome, "claude", "skills")
}

// HooksDir returns the path to the hooks directory.
func (r *Resolver) HooksDir() string {
    return filepath.Join(r.ClaudeDir(), "hooks")
}
```

---

## 7. Conflict Resolution

### 7.1 Three-Way Merge Strategy

Sync domain reuses the manifest domain's merge logic:

```go
import "github.com/autom8y/ariadne/internal/manifest"

// ResolveConflict resolves a sync conflict using three-way merge.
func (r *Resolver) ResolveConflict(conflict ConflictEntry, strategy string) (*ResolveResult, error) {
    // Load conflict file
    conflictData, err := r.loadConflictFile(conflict.ConflictFile)
    if err != nil {
        return nil, err
    }

    // Create manifest wrappers for merge
    baseManifest := manifest.FromContent(conflictData.Base.Content)
    oursManifest := manifest.FromContent(conflictData.Ours.Content)
    theirsManifest := manifest.FromContent(conflictData.Theirs.Content)

    // Apply merge strategy
    opts := manifest.MergeOptions{Strategy: manifest.MergeStrategy(strategy)}
    result, err := manifest.Merge(baseManifest, oursManifest, theirsManifest, opts)
    if err != nil {
        return nil, err
    }

    if result.HasConflicts {
        return nil, errors.ErrMergeConflict([]string{conflict.File})
    }

    return &ResolveResult{
        Resource:       conflict.Resource,
        Name:           conflict.Name,
        File:           conflict.File,
        Strategy:       strategy,
        ResultChecksum: ComputeContentChecksum(result.Merged),
    }, nil
}
```

### 7.2 Merge Strategies

| Strategy | Behavior | Use Case |
|----------|----------|----------|
| `ours` | Keep local version | Preserve local customizations |
| `theirs` | Accept remote version | Adopt upstream changes |
| `merge` | Three-way merge | Combine both changes |

### 7.3 Conflict File Lifecycle

1. **Created**: During `ari sync pull` when conflict detected
2. **Stored**: In `.claude/sync/conflicts/{file}-{timestamp}.conflict`
3. **Resolved**: Via `ari sync resolve`
4. **Removed**: After successful resolution

---

## 8. Integration Points

### 8.1 Team Domain

Sync domain integrates with team domain for rite synchronization:
- `ari sync pull --resource=teams` updates `rites/` directory
- Team domain's `AGENT_MANIFEST.json` remains separate (team-local tracking)
- `ari team validate` can be called post-sync for validation

### 8.2 Manifest Domain

Sync domain heavily reuses manifest domain:
- `manifest.Merge()` for three-way conflict resolution
- `manifest.Diff()` for computing differences
- Schema validation via `ari manifest validate`

### 8.3 Session Domain

Sync operations are session-independent:
- Can sync without active session
- Session state is never synced (session-local)
- Sync history is separate from session audit trail

### 8.4 State-Mate Integration

State-mate can invoke sync commands:

```bash
# Check for available updates
ari sync status --output=json

# Pull latest rites
ari sync pull --resource=teams --output=json

# After resolving conflicts
ari sync resolve --strategy=ours --all --output=json
```

---

## 9. Test Strategy

### 9.1 Unit Tests

Location: `internal/sync/*_test.go`

| Package | Test Focus | Coverage Target |
|---------|-----------|-----------------|
| `sync` | State load/save, checksum computation | 100% |
| `sync/tracker` | Status determination, file tracking | 100% |
| `sync/remote` | URL conversion, fetch simulation | 90% |
| `sync/pull` | Pull logic, conflict detection | 100% |
| `sync/resolve` | Conflict resolution | 100% |

### 9.2 Integration Tests

Location: `tests/integration/sync_test.go`

| Test ID | Description |
|---------|-------------|
| `sync_001` | Status shows synced state |
| `sync_002` | Status detects local modifications |
| `sync_003` | Status detects remote ahead |
| `sync_004` | Pull fetches new files |
| `sync_005` | Pull updates existing files |
| `sync_006` | Pull detects conflicts |
| `sync_007` | Pull with --force overwrites |
| `sync_008` | Pull --dry-run previews changes |
| `sync_009` | Diff shows differences |
| `sync_010` | Resolve with strategy=ours keeps local |
| `sync_011` | Resolve with strategy=theirs accepts remote |
| `sync_012` | Resolve with strategy=merge combines |
| `sync_013` | History shows operations |
| `sync_014` | Reset clears state |
| `sync_015` | Reset --hard deletes files (with confirm) |

### 9.3 Network Tests

Mock HTTP server for testing remote fetching:

```go
func TestFetchGitHubRaw(t *testing.T) {
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("mock content"))
    }))
    defer server.Close()

    fetcher := NewRemoteFetcher()
    content, err := fetcher.FetchFile(RemoteSource{URL: server.URL}, "test.md")

    assert.NoError(t, err)
    assert.Equal(t, "mock content", string(content))
}
```

### 9.4 Test Fixtures

```
ariadne/
└── testdata/
    └── sync/
        ├── state/
        │   ├── empty-state.json
        │   ├── synced-state.json
        │   └── conflict-state.json
        ├── remotes/
        │   └── mock-roster/
        │       ├── rites/
        │       │   └── 10x-dev/
        │       │       └── workflow.yaml
        │       └── skills/
        │           └── commit-ref.md
        └── conflicts/
            ├── base-version.yaml
            ├── local-version.yaml
            └── remote-version.yaml
```

---

## 10. Implementation Guidance

### 10.1 Recommended Order

1. **Foundation** (Day 1-2)
   - `internal/paths/sync.go` - Path helpers
   - `internal/sync/state.go` - State management
   - `internal/sync/tracker.go` - Checksum computation

2. **Read Operations** (Day 3-4)
   - `cmd/sync/status.go` - Status command
   - `internal/sync/remote.go` - Remote fetching (mock first)
   - `cmd/sync/diff.go` - Diff command

3. **Pull Operations** (Day 5-7)
   - `internal/sync/pull.go` - Pull logic
   - `internal/sync/conflict.go` - Conflict detection
   - `cmd/sync/pull.go` - Pull command

4. **Resolution** (Day 8-9)
   - `internal/sync/resolve.go` - Resolution logic
   - `cmd/sync/resolve.go` - Resolve command

5. **Utilities** (Day 10-12)
   - `cmd/sync/push.go` - Push command
   - `cmd/sync/history.go` - History command
   - `cmd/sync/reset.go` - Reset command

6. **Integration** (Day 13-14)
   - Integration tests
   - Network tests with mock server
   - Documentation

### 10.2 Dependency on Existing Packages

Sync domain reuses from previous domains:
- `internal/paths` - Path resolution (extend)
- `internal/output` - Printer pattern (extend)
- `internal/errors` - Error types (extend)
- `internal/manifest/merge` - Three-way merge (reuse directly)

### 10.3 External Dependencies

No new dependencies required. Existing dependencies cover needs:
- `net/http` - HTTP fetching (stdlib)
- `os/exec` - Git commands (stdlib)
- `crypto/sha256` - Checksums (stdlib)
- `gopkg.in/yaml.v3` - Config parsing (existing)

---

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Network failures during pull | Medium | Medium | Retry logic, partial state recovery |
| Large file downloads | Low | Medium | Streaming, progress indicators |
| Git authentication issues | Medium | Low | Clear error messages, doc auth setup |
| State file corruption | Low | High | Backup before write, schema validation |
| Conflict resolution errors | Medium | Medium | Preview mode, backup original files |
| Rate limiting (GitHub API) | Low | Low | Use raw URLs, cache fetched content |

---

## 12. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| ADR-ariadne-010 | Proposed | Sync state schema design |
| ADR-ariadne-011 | Proposed | Remote source configuration |
| ADR-ariadne-012 | Proposed | Conflict file format |

---

## 13. Handoff Criteria

Ready for Implementation when:

- [x] All 7 sync commands have interface contracts
- [x] Sync state schema defined
- [x] Conflict resolution strategy documented
- [x] Error codes mapped to exit codes
- [x] Integration with manifest merge specified
- [x] Test scenarios cover critical paths
- [ ] Principal Engineer can implement without architectural questions
- [ ] All artifacts verified via Read tool

---

## 14. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-sync.md` | Write |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-ariadne.md` | Read |
| Session TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |
| Team TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-rite.md` | Read |
| Manifest TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-manifest.md` | Read |
| Errors Package | `/Users/tomtenuta/Code/roster/ariadne/internal/errors/errors.go` | Read |
| Merge Logic | `/Users/tomtenuta/Code/roster/ariadne/internal/manifest/merge.go` | Read |
| Paths Package | `/Users/tomtenuta/Code/roster/ariadne/internal/paths/paths.go` | Read |
