# PRD: Ariadne (ari)

> The thread that makes the maze survivable—forged in Go.

**Status**: Draft
**Author**: Tom Tenuta + Claude
**Date**: 2026-01-04
**Initiative**: Ariadne Go CLI Architecture

---

## Executive Summary

**Ariadne** (`ari`) is a Go binary that replaces the current bash script harness for Claude Code agentic workflows. It provides the "thread" that enables deterministic return paths through complex multi-agent sessions.

### The Knossos Metaphor

| Myth | Ariadne Equivalent |
|------|-------------------|
| The Thread | Session state + provenance + audit trail |
| The Labyrinth | Codebase complexity |
| Navigation | `ari session`, `ari team` commands |
| Survival | Lock management, atomic operations, recovery |
| Return to Athens | Successful session wrap with quality gates |

---

## 1. Problem Statement

The current bash script harness (`session-manager.sh`, `swap-rite.sh`, `roster-sync`, etc.) has limitations:

1. **Distribution**: Requires consistent shell environments, jq, yq dependencies
2. **Performance**: File hashing, manifest diffing slow in bash
3. **Type Safety**: Edge cases accumulate; string parsing errors
4. **Maintainability**: Complex bash logic is hard to test and debug

### Success Criteria

- All bash script functionality replaced by Go binary
- Superset capabilities: structured validation, three-way merge
- Single binary distribution via `go install`, mise, or brew
- State-mate agent can invoke ari for all state mutations

---

## 2. Scope

### 2.1 In Scope (v1.0)

**Four domains, fully implemented:**

| Domain | Commands | Replaces |
|--------|----------|----------|
| **session** | create, status, park, resume, wrap, list, transition, migrate, audit, lock, unlock | session-manager.sh, session-fsm.sh |
| **team** | switch, list, status, validate | swap-rite.sh |
| **manifest** | show, diff, validate, merge | CEM manifest operations |
| **sync** | init, pull, push, status, diff, validate, repair | roster-sync |

**Superset capabilities:**
- Structured JSON schema validation (embedded schemas)
- Three-way merge for JSON (smart merge) and Markdown (anchor-based)

### 2.2 Out of Scope (Non-Goals)

| Exclusion | Rationale |
|-----------|-----------|
| **No AI/LLM logic** | state-mate remains the "brain"; ari is the "hands" |
| **No TUI** | CLI only. No curses, no bubble tea, no interactive UI |
| **No daemon mode** | Invoked, does work, exits. No long-running process |
| **Shell completion** | Deferred to v1.1+ |
| **Self-update** | Defer to package manager (mise/brew) |

---

## 3. Architecture

### 3.1 Repository Location

```
roster/              # Future: knossos/
├── ariadne/         # Go module root
│   ├── cmd/
│   │   └── ari/
│   │       └── main.go
│   ├── internal/
│   │   ├── cmd/           # Command implementations
│   │   │   ├── root.go
│   │   │   ├── session/
│   │   │   ├── team/
│   │   │   ├── manifest/
│   │   │   └── sync/
│   │   ├── validation/    # Schema validation
│   │   ├── merge/         # Three-way merge
│   │   ├── paths/         # XDG + project discovery
│   │   ├── lock/          # File locking
│   │   └── output/        # JSON/text formatting
│   ├── schemas/           # Embedded JSON schemas
│   ├── go.mod
│   └── go.sum
└── ...
```

**Module path**: `github.com/autom8y/ariadne`

### 3.2 Dependencies

| Purpose | Library | Version |
|---------|---------|---------|
| CLI Framework | `github.com/spf13/cobra` | v1.8+ |
| Config | `github.com/spf13/viper` | v1.18+ |
| JSON Schema | `github.com/santhosh-tekuri/jsonschema/v6` | v6+ |
| XDG Paths | `github.com/adrg/xdg` | v0.5+ |
| JSON Merge | `github.com/evanphx/json-patch/v5` | v5+ |
| YAML | `gopkg.in/yaml.v3` | v3 |
| Markdown | `github.com/yuin/goldmark` | v1.6+ |

### 3.3 Knossos Future

When roster renames to knossos, `ari` remains an **independent sibling binary**. They are peers, not parent-child:

```
knossos/        # The world (platform)
ariadne/        # The thread (survival mechanism)
```

---

## 4. Interface Specification

### 4.1 Command Structure

```
ari
├── session
│   ├── create <initiative> [--complexity=MODULE] [--team=NAME]
│   ├── status [--session-id=ID]
│   ├── list [--all] [--status=STATUS]
│   ├── park [--reason=TEXT]
│   ├── resume [--session-id=ID]
│   ├── wrap [--skip-checks]
│   ├── transition <phase> [--force]
│   ├── migrate [--session-id=ID]
│   ├── audit [--session-id=ID] [--limit=N]
│   ├── lock [--session-id=ID]
│   └── unlock [--session-id=ID] [--force]
│
├── team
│   ├── switch <team-name> [--remove-all|--keep-all|--promote-all]
│   ├── list
│   ├── status
│   └── validate [--team=NAME]
│
├── manifest
│   ├── show [--path=PATH]
│   ├── diff <path1> <path2>
│   ├── validate <path>
│   └── merge <base> <ours> <theirs> [--output=PATH]
│
├── sync
│   ├── init [--team=NAME]
│   ├── pull [--dry-run]
│   ├── push [--dry-run]
│   ├── status
│   ├── diff
│   ├── validate
│   └── repair [--dry-run]
│
└── version
```

### 4.2 Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output format: text, json, yaml | text |
| `--verbose` | `-v` | Enable verbose output (JSON lines) | false |
| `--config` | | Config file path | $XDG_CONFIG_HOME/ariadne/config.yaml |
| `--project-dir` | `-p` | Project root (overrides discovery) | (walk up for .claude/) |
| `--session-id` | `-s` | Session ID (overrides current) | (from .current-session) |

### 4.3 Project Discovery

1. Start from current working directory
2. Walk up directory tree looking for `.claude/`
3. If found, that's the project root
4. If not found (hit filesystem root), exit with error:
   ```
   Error: No .claude/ directory found.
   Run from within a project or use --project-dir.
   ```

### 4.4 Output Contract

**Success (--output=json)**:
```json
{
  "session_id": "session-20260104-160414-563c681e",
  "status": "PARKED",
  "parked_at": "2026-01-04T16:30:00Z",
  "parked_reason": "waiting for user input"
}
```

**Error (--output=json)**:
```json
{
  "error": {
    "code": "LOCK_TIMEOUT",
    "message": "Could not acquire lock within 10s",
    "details": {
      "lock_path": ".claude/sessions/.locks/session-abc.lock",
      "holder_pid": 12345
    }
  }
}
```

**Success (--output=text, default)**:
- Silent (exit 0, no output) for mutations
- Data output for queries (status, list, etc.)

**Verbose (--verbose)**:
- JSON lines to stderr for debugging
- `{"level":"info","msg":"Acquiring lock...","ts":"2026-01-04T16:30:00Z"}`

---

## 5. Error Handling

### 5.1 Error Codes (Flat Taxonomy)

| Code | Exit | Description |
|------|------|-------------|
| `SUCCESS` | 0 | Operation completed successfully |
| `GENERAL_ERROR` | 1 | Unspecified error |
| `USAGE_ERROR` | 2 | Invalid arguments or flags |
| `LOCK_TIMEOUT` | 3 | Could not acquire lock |
| `LOCK_STALE` | 3 | Lock holder process dead (auto-recovered) |
| `SCHEMA_INVALID` | 4 | Data failed schema validation |
| `LIFECYCLE_VIOLATION` | 5 | Invalid state transition |
| `FILE_NOT_FOUND` | 6 | Required file missing |
| `PERMISSION_DENIED` | 7 | Cannot read/write file |
| `MERGE_CONFLICT` | 8 | Three-way merge has conflicts |
| `PROJECT_NOT_FOUND` | 9 | No .claude/ directory found |

### 5.2 Corrupt State Recovery

When encountering corrupt SESSION_CONTEXT.md:

1. **Backup** original file to `.corrupt/SESSION_CONTEXT.md.{timestamp}`
2. **Attempt repair**:
   - Fix missing required fields with defaults
   - Correct type mismatches where possible
   - Remove unknown fields
3. **Validate** repaired state against schema
4. **Write** repaired state if valid
5. **Report** what was repaired in output

### 5.3 Concurrency & Locking

**Strategy**: File locking via `flock()` with stale detection

1. Attempt to acquire advisory lock on `.claude/sessions/.locks/{session-id}.lock`
2. If lock held, check if holder PID is alive:
   - **Alive**: Wait up to 10s, then fail with `LOCK_TIMEOUT`
   - **Dead**: Steal lock (stale detection), log warning
3. Perform operation
4. Release lock

---

## 6. Schema Management

### 6.1 Embedded Schemas

Compiled into binary via `//go:embed`:

| Schema | Purpose |
|--------|---------|
| `session-context.schema.json` | SESSION_CONTEXT.md validation |
| `sprint-context.schema.json` | SPRINT_CONTEXT.md validation |
| `manifest.schema.json` | CEM manifest validation |
| `team-manifest.schema.json` | Team pack manifest validation |

### 6.2 External Schema Loading

Artifact schemas (PRD, TDD, ADR, etc.) loaded from filesystem:

```yaml
# In embedded config
artifact_schemas:
  prd:
    path: "schemas/artifacts/prd.schema.json"
    hash: "sha256:abc123..."  # Optional integrity check
  tdd:
    path: "schemas/artifacts/tdd.schema.json"
```

---

## 7. Three-Way Merge

### 7.1 JSON Merge (Smart Merge)

Field-level analysis:

| Scenario | Resolution |
|----------|------------|
| Field in theirs only (new) | Accept theirs |
| Field in ours only (new) | Accept ours |
| Both modified same field | CONFLICT - flag for manual resolution |
| Only ours modified | Accept ours |
| Only theirs modified | Accept theirs |
| Neither modified | Keep original |

Implementation: `github.com/evanphx/json-patch/v5` with custom conflict detection.

### 7.2 Markdown Merge (Anchor-Based)

Use existing anchor system from `claude-md-architecture` skill:

1. Parse markdown into sections by `<!-- ANCHOR: name -->` comments
2. Match sections by anchor name
3. Apply same conflict resolution as JSON
4. Sections without anchors: position-based fallback with warning

---

## 8. Configuration

### 8.1 Minimal Config Scope

Only two settings are user-configurable:

```yaml
# $XDG_CONFIG_HOME/ariadne/config.yaml
default_output: json    # text | json | yaml
default_team: 10x-dev-pack
```

Everything else is hardcoded for consistency.

### 8.2 XDG Directory Layout

```
$XDG_CONFIG_HOME/ariadne/     # ~/.config/ariadne/
├── config.yaml                # User preferences

$XDG_STATE_HOME/ariadne/      # ~/.local/state/ariadne/
├── audit.log                  # Global audit trail

$XDG_CACHE_HOME/ariadne/      # ~/.cache/ariadne/
├── schemas/                   # Cached external schemas
```

---

## 9. Integration

### 9.1 State-Mate Interface

State-mate (LLM agent) invokes ari for all state mutations:

```bash
# State-mate calls:
ari session park --reason="waiting for user input" --output=json

# Ari returns:
{"session_id": "...", "status": "PARKED", ...}
```

Capability discovery: Hardcoded in `state-mate.md` agent prompt. Future: MCP resource.

### 9.2 Hook Integration

Hooks remain bash scripts that call ari:

```bash
#!/bin/bash
# .claude/hooks/session-guards/auto-park.sh
ari session park --reason="auto-park on stop" --output=json
```

### 9.3 Migration Bridge

During migration, existing bash scripts call ari:

```bash
# session-manager.sh (updated)
case "$1" in
  create) ari session create "${@:2}" ;;
  park)   ari session park "${@:2}" ;;
  *)      echo "Unknown command: $1" >&2; exit 1 ;;
esac
```

**Post v1.0**: Delete all bash scripts. Clean break.

---

## 10. Distribution

### 10.1 Installation Methods

**Primary**: Bundled with roster
```bash
# roster-sync installs ari automatically
./roster-sync init
```

**Secondary**: Package managers
```bash
# mise
mise use -g ariadne

# brew
brew install autom8y/tap/ariadne
```

**Developer**: go install
```bash
go install github.com/autom8y/ariadne/cmd/ari@latest
```

### 10.2 Versioning

**Scheme**: SemVer with v0 freedom

- `v0.x.x`: Breaking changes allowed during development
- `v1.0.0`: Stable release, SemVer strictly enforced thereafter

### 10.3 Version Command

```bash
$ ari version
ari v0.1.0 (abc1234, 2026-01-04)
go1.23.0 darwin/arm64
```

---

## 11. Testing Strategy

### 11.1 Test Types

| Type | Purpose | Tool |
|------|---------|------|
| Unit | Core logic validation | `go test` |
| Integration | CLI command testing | `go test` + fixtures |
| Parity | Behavior specification | Custom spec tests |
| Concurrency | Race + chaos | `go test -race` + parallel CLI |

### 11.2 Behavior Specification

Tests validate against **specification**, not bash script behavior:

```go
func TestSessionPark_Specification(t *testing.T) {
    // Specification: Park transitions ACTIVE -> PARKED
    // Specification: Park sets parked_at timestamp
    // Specification: Park requires reason
    // ...
}
```

Bash scripts may have bugs; spec is authoritative.

### 11.3 Satellite Matrix Testing

CI matrix tests against fixture projects:

```yaml
# .github/workflows/test.yaml
strategy:
  matrix:
    fixture: [minimal, standard, complex]
```

### 11.4 Coverage Bar

- Critical paths: 100% tested
- Integration scenarios: All documented use cases
- Line coverage: Not a target (scenario coverage > line %)

---

## 12. Documentation Requirements

### 12.1 Required for v1.0

| Document | Location | Content |
|----------|----------|---------|
| README.md | ariadne/README.md | Installation, quick start, overview |
| Command Reference | ariadne/docs/commands.md | All commands with examples |

### 12.2 Deferred

- Architecture documentation
- Contributor guide
- Migration guide (users moving from bash)

---

## 13. Implementation Plan

### 13.1 Phases

| Phase | Deliverable | Domain |
|-------|-------------|--------|
| **0** | Skeleton | Root cmd, XDG paths, embedded schemas, CI |
| **1** | Session domain | `ari session *` commands |
| **2** | Team domain | `ari team *` commands |
| **3** | Manifest domain | `ari manifest *` commands |
| **4** | Sync domain | `ari sync *` commands |
| **5** | Polish | Docs, dogfooding, final validation |

### 13.2 Sprint Structure

- **One sprint per domain** (Phases 1-4)
- **Single long session** for entire initiative
- **Milestone-based timeline** (quality over calendar)

### 13.3 First Domain: Session

Session domain first because:
1. Most critical (the actual "thread")
2. Most complex (front-load risk)
3. Enables dogfooding early

### 13.4 Definition of Done (v1.0)

- [ ] All four domains implemented
- [ ] All tests passing (unit, integration, concurrency)
- [ ] CI pipeline green
- [ ] Dogfooding period complete (internal use)
- [ ] README.md complete
- [ ] Command reference complete
- [ ] Compatibility-tester agent validates
- [ ] You (Tom) approve final release

---

## 14. Success Metrics

| Metric | Target |
|--------|--------|
| Bash scripts replaced | 100% (delete on ship) |
| Test scenarios covered | All documented use cases |
| CI matrix passing | All fixture types |
| Startup time | < 50ms |
| Lock acquisition | < 100ms (p99) |

---

## 15. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Markdown merge complexity | Medium | Medium | Flag conflicts for manual resolution |
| Behavioral parity gaps | Medium | High | Spec-based testing, not bash comparison |
| Cross-platform issues | Low | Medium | CI matrix includes Linux + macOS |
| Lock contention | Low | Medium | Stale detection + configurable timeout |

---

## Appendix A: Decision Log

All decisions from Q&A session (2026-01-04):

### Scope & Repository
- Repo location: Inside roster (future knossos)
- V1 scope: All four domains complete
- Non-goals: No AI, no TUI, no daemon
- Module path: github.com/autom8y/ariadne

### Interface
- Session commands: Full lifecycle (11 commands)
- JSON contract: gh CLI pattern
- Global flags: --output, --verbose, --config, --project-dir, --session-id
- Project discovery: Walk up to find .claude/

### Integration
- Migration bridge: Bash calls ari
- Bash fate: Delete immediately on v1.0
- Hook integration: Hooks call ari binary
- Capability discovery: Hardcoded in state-mate.md

### Testing
- Parity testing: Behavior specification
- Coverage bar: Critical path + integration
- Matrix testing: CI matrix job
- Concurrency: Race detector + parallel CLI + chaos

### Error Handling
- Error taxonomy: Flat codes
- Corrupt state: Attempt repair
- Write conflict: flock()
- Lock timeout: Stale detection + steal

### Configuration
- Config scope: Minimal (2 settings)
- Versioning: SemVer v0.x freedom
- Installation: Bundled with roster + mise/brew
- Self-update: Defer to package manager

### Implementation
- Done criteria: Domains + tests + dogfooding + docs
- Sprint structure: Domain per sprint
- Timeline: Milestone-based
- Final validator: You + CI + agent
- First domain: session

### Schema & Merge
- Schema scope: Embed session/sprint/manifest/team
- JSON merge: Smart merge
- MD merge: Anchor-based
- Required docs: README + command ref

### Final Details
- No project behavior: Error and exit
- Shell completion: Not for v1
- Default verbosity: Silent success
- Log format: JSON lines
- Binary name: ari
- Knossos future: Independent sibling
- Exit codes: Categorized
- Version flag: Full build info
