# SPIKE: Materialization Model for Knossos Configuration

**Date**: 2026-01-07
**Initiative**: knossos-finalization
**Phase**: R&D (research only, NO working code)
**Session**: session-20260107-164631-8dd6f03a
**Upstream**: SPIKE-knossos-consolidation-architecture.md

---

## Executive Summary

This SPIKE evaluates configuration materialization and templating patterns for Knossos, where `.claude/` directories are fully generated from templates rather than checked into repositories. The core finding is that a **chezmoi-inspired generation model** with **Go's text/template + Sprig functions** provides the optimal balance of power, simplicity, and ecosystem alignment.

**Key Recommendations**:
1. **Templating Engine**: Go `text/template` with Sprig function library (already partially implemented)
2. **UX Pattern**: Single `ari sync` command with automatic initialization (idempotent)
3. **Conflict Model**: Three-way merge with explicit user resolution for customizations
4. **User Journey**: Install binary -> Clone project -> `ari sync` materializes `.claude/`

---

## 1. Research Areas

### 1.1 Go Templating Engines

#### Options Evaluated

| Engine | Type | Functions | Ecosystem |
|--------|------|-----------|-----------|
| **text/template** | Standard library | ~15 built-in | Universal Go support |
| **Sprig** | Function library | 100+ functions | Helm, gomplate, many CLIs |
| **gomplate** | CLI + library | Sprig + datasources | DevOps/config generation |
| **Sprout** | Sprig fork | Same as Sprig, faster | Newer, less adoption |

#### Analysis

**Go text/template (Standard Library)**
- Pros: No dependencies, stable API, excellent tooling support
- Cons: Limited built-in functions (no string manipulation, no date formatting)
- Current usage: `ariadne/internal/inscription/generator.go` already uses this

**Sprig Library** ([GitHub](https://github.com/Masterminds/sprig))
- 100+ template functions: string manipulation, math, date, lists, dictionaries
- De facto standard for Go templating (Helm, Kubernetes ecosystem)
- Active maintenance (v3 is current stable)
- Provides: `trim`, `upper`, `lower`, `default`, `coalesce`, `toJson`, `fromJson`, etc.

**gomplate** ([Documentation](https://docs.gomplate.ca/))
- CLI tool wrapping text/template + Sprig
- Adds datasource support (JSON, YAML, environment, vault)
- Useful for: multi-file template processing, CI pipelines
- Overkill for embedded use; better as CLI companion

**Sprout** ([GitHub](https://github.com/go-sprout/sprout))
- ~45% faster than Sprig, 16.5% less memory
- Better error handling (safe defaults over panics)
- Newer, less battle-tested

#### Recommendation: text/template + Sprig

**Rationale**:
1. Already using text/template in `generator.go`
2. Sprig is the de facto standard (Helm compatibility)
3. Current implementation adds custom functions (`include`, `ifdef`, `agents`, `term`)
4. Sprout is promising but premature to adopt

**Current Implementation Gap**:
```go
// ariadne/internal/inscription/generator.go lines 201-212
func (g *Generator) templateFuncs() template.FuncMap {
    return template.FuncMap{
        "include": g.includePartial,
        "ifdef":   g.conditionalInclude,
        "agents":  g.loadAgentTable,
        "term":    g.lookupTerminology,
        "join":    strings.Join,
        "lower":   strings.ToLower,
        "upper":   strings.ToUpper,
        "title":   strings.Title,
    }
}
```

**Enhancement**: Add Sprig functions alongside custom functions:
```go
import "github.com/Masterminds/sprig/v3"

func (g *Generator) templateFuncs() template.FuncMap {
    funcs := sprig.TxtFuncMap() // Base Sprig functions
    // Add Knossos-specific functions
    funcs["include"] = g.includePartial
    funcs["ifdef"] = g.conditionalInclude
    funcs["agents"] = g.loadAgentTable
    funcs["term"] = g.lookupTerminology
    return funcs
}
```

---

### 1.2 Configuration Generation Patterns

#### Tools Evaluated

| Tool | Approach | Trigger | Conflict Handling |
|------|----------|---------|-------------------|
| **direnv** | Auto-load .envrc | Directory entry | None (declarative) |
| **mise-en-place** | Runtime manager | Shell hook | Override hierarchy |
| **Nix/devenv** | Declarative env | Flake activation | Deterministic |
| **chezmoi** | File generation | Explicit command | Three-way merge |
| **GNU Stow** | Symlink farm | Explicit command | None |

#### Analysis

**direnv + mise Pattern**
- Environment variables loaded automatically on directory entry
- Good for: secrets, tool versions, project-specific config
- Not applicable: File generation (only env vars)

**Nix Flakes + devenv**
- Fully declarative, reproducible environments
- Excellent for: dependency management, toolchain
- Overkill for: file generation from templates
- Learning curve: Steep Nix language

**chezmoi** ([Design FAQ](https://www.chezmoi.io/user-guide/frequently-asked-questions/design/))
- File generation from templates (not symlinks)
- Key design decisions:
  - **Copies over symlinks**: Enables encryption, permissions, templates
  - **Single source of truth**: One repo branch, no ambiguity
  - **Metadata in filenames**: `dot_`, `executable_`, `private_` prefixes
  - **Explicit apply**: `chezmoi apply` materializes files

**GNU Stow**
- Symlink management only
- Limitation: "not all applications behave well when their configuration directories are symlinked"
- Not recommended: Claude Code expects real files in `.claude/`

#### Recommendation: chezmoi-Inspired Generation Model

**Rationale**:
1. Claude Code hooks require real files (not symlinks)
2. Templates enable rite-specific customization
3. Explicit command prevents accidental state changes
4. Three-way merge supports user customizations

**Key Design Principles from chezmoi**:
- Generate files in final location (not symlinks)
- Single canonical template source (`templates/` or `rites/{name}/`)
- Idempotent operations (safe to re-run)
- Explicit conflict resolution

---

### 1.3 File Synchronization Approaches

#### Sync Models Evaluated

| Model | Direction | Conflict | Idempotent |
|-------|-----------|----------|------------|
| **rsync** | One-way | Overwrite | Yes |
| **Unison** | Two-way | Interactive | Yes |
| **Mutagen** | Two-way | Three-way merge | Yes |
| **Git** | Distributed | Manual resolution | Yes |

#### Analysis

**rsync Pattern**
- Idempotent: Same result on multiple runs
- One-way only: Source always wins
- Limitation: No conflict detection for local changes

**Mutagen Three-Way Merge**
- Tracks "most-recently agreed-upon content" as base
- Compares both endpoints against base
- Detects conflicts without data loss
- Mode: `two-way-safe` (default) preserves both sides on conflict

**Git Merge Pattern**
- Three-way merge with common ancestor
- Conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`)
- Requires manual resolution

#### Recommendation: Three-Way Merge with Base State

**Model**:
```
                    ┌─────────────────┐
                    │   BASE STATE    │
                    │ (last sync)     │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
              v                             v
     ┌─────────────────┐           ┌─────────────────┐
     │  LOCAL STATE    │           │  REMOTE STATE   │
     │ (user changes)  │           │ (template)      │
     └────────┬────────┘           └────────┬────────┘
              │                             │
              └──────────────┬──────────────┘
                             │
                             v
                    ┌─────────────────┐
                    │  MERGE RESULT   │
                    └─────────────────┘
```

**Implementation**:
- Store checksums of generated files in `.knossos/sync/state.json`
- On sync: Compare LOCAL vs BASE and REMOTE vs BASE
- No changes to LOCAL: Accept REMOTE (simple update)
- Changes to LOCAL, no changes to REMOTE: Keep LOCAL (user customization)
- Both changed: CONFLICT (requires resolution)

---

### 1.4 Idempotent Sync vs Init+Update Patterns

#### Patterns Evaluated

| Pattern | Commands | First Run | Subsequent |
|---------|----------|-----------|------------|
| **Init + Update** | `init`, `update` | `init` required | `update` only |
| **Single Sync** | `sync` | Auto-init | Same command |
| **Terraform** | `init`, `apply` | `init` required | `apply` only |
| **chezmoi** | `init`, `apply` | `init` required | `apply` only |

#### Analysis

**Separate Init + Update**
- Pro: Explicit lifecycle (users know when they're initializing)
- Con: Two commands to learn; error if running wrong one
- Example: Terraform (`init` required before `apply`)

**Single Idempotent Command**
- Pro: One command for all states ("just run sync")
- Pro: Aligns with CLI UX best practice ("crash-only design")
- Con: Less explicit about initialization

**Hybrid (Init with Automatic Sync)**
- `ari init` creates project structure, then syncs
- `ari sync` works after initialization
- Clear separation but seamless usage

#### Recommendation: Single Idempotent `ari sync`

**Rationale**:
1. **Simplicity**: "When in doubt, run `ari sync`"
2. **Crash-only design**: Re-running after failure recovers state
3. **No "wrong command" errors**: Works regardless of current state
4. **CLI UX best practice**: Per [clig.dev](https://clig.dev/), prefer idempotent operations

**Behavior Matrix**:

| State | `ari sync` Behavior |
|-------|---------------------|
| No `.claude/` | Initialize + materialize |
| `.claude/` exists, no state.json | Initialize state, detect changes |
| `.claude/` exists, state.json valid | Normal sync (three-way merge) |
| Conflicts exist | Report conflicts, prompt resolution |

**Optional Explicit Commands**:
- `ari sync --init`: Force re-initialization
- `ari sync --force`: Overwrite local changes (dangerous)
- `ari sync status`: Check state without modifying

---

## 2. User Journey Mapping

### 2.1 User Personas

| Persona | Description | Primary Need |
|---------|-------------|--------------|
| **New Adopter** | First time using Knossos | Quick setup, working defaults |
| **Existing User** | Already using Knossos | Update templates, preserve customizations |
| **Rite Switcher** | Switching between rites | Fast context switch, clean state |
| **Project Maintainer** | Manages Knossos templates | Author and distribute rites |

### 2.2 User Journey: New User Onboarding

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        NEW USER ONBOARDING FLOW                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 1: Install Binary                                                     │
│  ──────────────────────                                                     │
│  User runs one of:                                                          │
│    brew install knossos/tap/ari                                             │
│    curl -fsSL https://knossos.dev/install | sh                              │
│    go install github.com/autom8y/knossos/cmd/ari@latest                     │
│                                                                             │
│  ✓ `ari` available in PATH                                                  │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 2: Clone/Enter Project                                                │
│  ───────────────────────────                                                │
│  User enters project directory:                                             │
│    cd ~/projects/my-app                                                     │
│                                                                             │
│  Project may or may not have Knossos configured                             │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 3: Initialize Knossos                                                 │
│  ─────────────────────────                                                  │
│  User runs:                                                                 │
│    ari sync                                                                 │
│                                                                             │
│  First-run detection (no .claude/ or state.json):                           │
│    → Interactive prompt:                                                    │
│      "No Knossos configuration found. Initialize? [Y/n]"                    │
│    → Select rite (or use default):                                          │
│      "Select rite: [1] 10x-dev (default) [2] hygiene [3] rnd [4] custom"    │
│                                                                             │
│  Actions:                                                                   │
│    1. Create .claude/ directory                                             │
│    2. Materialize hooks from templates/hooks/                               │
│    3. Materialize agents from rites/{selected}/agents/                      │
│    4. Materialize skills from rites/{selected}/skills/ + rites/shared/      │
│    5. Generate CLAUDE.md from templates/CLAUDE.md.tpl                       │
│    6. Write .knossos/sync/state.json with checksums                          │
│    7. Write .claude/ACTIVE_RITE file                                        │
│                                                                             │
│  Output:                                                                    │
│    "Knossos initialized with '10x-dev' rite"                                │
│    "Generated: 6 agents, 15 skills, 12 hooks"                               │
│    "Run 'ari session create' to start working"                              │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 4: Start Working                                                      │
│  ────────────────────                                                       │
│  User starts Claude Code session:                                           │
│    claude                                                                   │
│                                                                             │
│  CLAUDE.md is read, hooks activate, agents available                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 User Journey: Existing User Update

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        EXISTING USER UPDATE FLOW                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 1: Check for Updates                                                  │
│  ────────────────────────                                                   │
│  User runs:                                                                 │
│    ari sync status                                                          │
│                                                                             │
│  Output:                                                                    │
│    "Remote: github.com/autom8y/knossos (main)"                              │
│    "Local rite: 10x-dev"                                                    │
│    "Status: 3 files have upstream changes"                                  │
│    "  agents/architect.md (MODIFIED upstream)"                              │
│    "  skills/prompting/prompting.md (MODIFIED upstream)"                    │
│    "  hooks/lib/session-manager.sh (MODIFIED upstream)"                     │
│    ""                                                                       │
│    "Run 'ari sync' to apply updates"                                        │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 2: Apply Updates                                                      │
│  ────────────────────                                                       │
│  User runs:                                                                 │
│    ari sync                                                                 │
│                                                                             │
│  Three-way merge process:                                                   │
│    For each changed file:                                                   │
│      1. Compare local vs base (last sync)                                   │
│      2. Compare remote vs base                                              │
│      3. Determine action:                                                   │
│         - Local unchanged: Accept remote                                    │
│         - Local changed, remote unchanged: Keep local                       │
│         - Both changed: Mark conflict                                       │
│                                                                             │
│  Output (no conflicts):                                                     │
│    "Sync complete"                                                          │
│    "  Updated: 3 files"                                                     │
│    "  Preserved: 2 local customizations"                                    │
│    "  Conflicts: 0"                                                         │
│                                                                             │
│  Output (with conflicts):                                                   │
│    "Sync complete with conflicts"                                           │
│    "  Updated: 2 files"                                                     │
│    "  Conflicts: 1 file"                                                    │
│    "    hooks/lib/session-manager.sh"                                       │
│    ""                                                                       │
│    "Run 'ari sync resolve' to resolve conflicts"                            │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 3: Resolve Conflicts (if any)                                         │
│  ─────────────────────────────────                                          │
│  User runs:                                                                 │
│    ari sync resolve                                                         │
│                                                                             │
│  Interactive resolution:                                                    │
│    "Conflict: hooks/lib/session-manager.sh"                                 │
│    "  [1] Keep local (your changes)"                                        │
│    "  [2] Accept remote (upstream changes)"                                 │
│    "  [3] Merge manually (opens diff)"                                      │
│    "  [4] Skip (resolve later)"                                             │
│                                                                             │
│  Alternative (batch):                                                       │
│    ari sync resolve --strategy=ours   # Keep all local                      │
│    ari sync resolve --strategy=theirs # Accept all remote                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.4 User Journey: Rite Switching

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          RITE SWITCHING FLOW                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 1: List Available Rites                                               │
│  ───────────────────────────                                                │
│  User runs:                                                                 │
│    ari rite list                                                            │
│                                                                             │
│  Output:                                                                    │
│    "Available rites:"                                                       │
│    "  * 10x-dev (active) - Full development workflow with 6 agents"         │
│    "    hygiene         - Code quality and cleanup (4 agents)"              │
│    "    rnd             - Research and innovation (6 agents)"               │
│    "    ecosystem       - Knossos platform development (8 agents)"          │
│    "    docs            - Documentation specialist (3 agents)"              │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 2: Switch Rite                                                        │
│  ─────────────────                                                          │
│  User runs:                                                                 │
│    ari rite switch rnd                                                      │
│                                                                             │
│  Actions:                                                                   │
│    1. Backup current customizations (if any)                                │
│    2. Clear .claude/agents/                                                 │
│    3. Clear rite-specific .claude/skills/                                   │
│    4. Materialize rites/rnd/agents/ -> .claude/agents/                      │
│    5. Materialize rites/rnd/skills/ -> .claude/skills/                      │
│    6. Sync shared skills from rites/shared/skills/                          │
│    7. Regenerate CLAUDE.md with new rite context                            │
│    8. Update .claude/ACTIVE_RITE                                            │
│    9. Update .knossos/sync/state.json                                        │
│                                                                             │
│  Output:                                                                    │
│    "Switched to 'rnd' rite"                                                 │
│    "  Agents: technology-scout, integration-researcher, prototype-engineer" │
│    "          moonshot-architect, tech-transfer, orchestrator"              │
│    "  Skills: 18 loaded (12 rite-specific, 6 shared)"                       │
│    ""                                                                       │
│    "Previous customizations backed up to .claude/backup/10x-dev/"           │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  STEP 3: Continue Working                                                   │
│  ──────────────────────                                                     │
│  User starts Claude Code:                                                   │
│    claude                                                                   │
│                                                                             │
│  New rite context active immediately                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Conflict Resolution Strategy

### 3.1 Conflict Types

| Type | Detection | Resolution |
|------|-----------|------------|
| **Clean Update** | Local unchanged, remote changed | Auto-accept remote |
| **Local Override** | Local changed, remote unchanged | Auto-preserve local |
| **Merge Conflict** | Both local and remote changed | User resolution required |
| **New File** | Remote has file, local doesn't | Auto-add |
| **Deleted File** | Remote deleted file | Auto-remove (or preserve with flag) |

### 3.2 Conflict File Format

When conflicts occur, store marker file:

```
.knossos/sync/conflicts/
  session-manager.sh.conflict.json
```

```json
{
  "file": "hooks/lib/session-manager.sh",
  "detected_at": "2026-01-07T16:00:00Z",
  "base_checksum": "sha256:abc123...",
  "local_checksum": "sha256:def456...",
  "remote_checksum": "sha256:ghi789...",
  "local_path": ".knossos/sync/conflicts/session-manager.sh.local",
  "remote_path": ".knossos/sync/conflicts/session-manager.sh.remote",
  "base_path": ".knossos/sync/conflicts/session-manager.sh.base"
}
```

### 3.3 Resolution Strategies

| Strategy | Flag | Behavior |
|----------|------|----------|
| **Interactive** | (default) | Prompt for each conflict |
| **Ours** | `--strategy=ours` | Keep all local changes |
| **Theirs** | `--strategy=theirs` | Accept all remote changes |
| **Manual** | `--strategy=manual` | Open diff tool (EDITOR) |

---

## 4. State Management

### 4.1 State File Schema

`.knossos/sync/state.json`:

```json
{
  "version": "1.0",
  "initialized_at": "2026-01-07T15:00:00Z",
  "last_sync": "2026-01-07T16:00:00Z",
  "active_rite": "10x-dev",
  "remote": {
    "url": "github.com/autom8y/knossos",
    "ref": "main",
    "commit": "abc123def456..."
  },
  "files": {
    "agents/architect.md": {
      "checksum": "sha256:...",
      "source": "rites/10x-dev/agents/architect.md",
      "generated_at": "2026-01-07T15:00:00Z",
      "modified_locally": false
    },
    "hooks/lib/session-manager.sh": {
      "checksum": "sha256:...",
      "source": "templates/hooks/lib/session-manager.sh",
      "generated_at": "2026-01-07T15:00:00Z",
      "modified_locally": true,
      "local_checksum": "sha256:..."
    }
  },
  "conflicts": []
}
```

### 4.2 Checksum Strategy

- Use SHA256 for file integrity
- Store both generated checksum and current checksum
- Detect local modifications by comparing stored vs current
- Track source file for regeneration

---

## 5. Templating Architecture

### 5.1 Template Directory Structure

```
templates/
├── CLAUDE.md.tpl              # Main CLAUDE.md template
├── hooks/
│   ├── lib/                   # Shell libraries
│   │   ├── session-manager.sh
│   │   └── ...
│   └── *.sh                   # Hook entry points
├── sections/
│   ├── quick-start.md.tpl
│   ├── execution-mode.md.tpl
│   └── ...
└── partials/
    ├── agent-table.md.tpl
    └── terminology-table.md.tpl

rites/
├── 10x-dev/
│   ├── manifest.yaml          # Rite metadata
│   ├── agents/
│   │   ├── architect.md
│   │   └── ...
│   └── skills/
│       └── ...
├── shared/
│   └── skills/                # Cross-rite skills
└── ...
```

### 5.2 Template Context

```go
type MaterializationContext struct {
    // Rite context
    ActiveRite   string
    RiteManifest RiteManifest
    Agents       []AgentInfo
    Skills       []SkillInfo

    // Project context
    ProjectRoot  string
    ProjectName  string

    // User context
    UserName     string
    UserEmail    string
    Preferences  UserPreferences

    // Knossos context
    KnossosVersion string
    TemplateVersion string
}
```

### 5.3 Template Functions

Standard Sprig + Knossos-specific:

| Function | Purpose | Example |
|----------|---------|---------|
| `include` | Include partial template | `{{include "partials/agent-table.md.tpl"}}` |
| `agents` | Generate agent table | `{{agents}}` |
| `term` | Lookup terminology | `{{term "knossos"}}` |
| `ifdef` | Conditional include | `{{ifdef .ActiveRite "content"}}` |
| `riteSkills` | List skills for rite | `{{riteSkills .ActiveRite}}` |

---

## 6. Recommendations Summary

### 6.1 Templating Engine

**Decision**: Go `text/template` + Sprig function library

**Rationale**:
- Already implemented base in `generator.go`
- Sprig provides 100+ utility functions
- Helm ecosystem compatibility
- No additional binary dependencies

**Action**: Add `github.com/Masterminds/sprig/v3` to go.mod

### 6.2 UX Pattern

**Decision**: Single idempotent `ari sync` command

**Rationale**:
- Simplest mental model ("when in doubt, sync")
- Aligns with CLI best practices (crash-only design)
- Handles all states (init, update, conflict)
- Optional flags for explicit control

**Commands**:
| Command | Purpose |
|---------|---------|
| `ari sync` | Initialize or update (idempotent) |
| `ari sync status` | Check state without modifying |
| `ari sync resolve` | Resolve conflicts |
| `ari sync --force` | Force overwrite local changes |

### 6.3 Conflict Model

**Decision**: Three-way merge with explicit resolution

**Rationale**:
- Preserves user customizations by default
- Clear conflict detection (not silent overwrites)
- Multiple resolution strategies (interactive, ours, theirs)
- Aligns with git mental model

### 6.4 User Journey

**Decision**: Install -> Clone -> `ari sync` -> Work

**Rationale**:
- Minimal steps to productivity
- Single command after project clone
- Interactive first-run guidance
- Non-interactive mode for CI/automation

---

## 7. Exit Criteria Verification

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Clear recommendation on materialization approach | DONE | Section 6: chezmoi-inspired generation with text/template + Sprig |
| User journey map with recommended flow | DONE | Section 2: Three user journeys mapped |
| Templating engine recommendation | DONE | Section 1.1 + 6.1: text/template + Sprig |
| Sync vs init UX decision | DONE | Section 1.4 + 6.2: Single idempotent `ari sync` |

---

## 8. Sources

### Web Research
- [Go Template Libraries Performance Comparison - LogRocket](https://blog.logrocket.com/golang-template-libraries-performance-comparison/)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
- [gomplate Documentation](https://docs.gomplate.ca/)
- [chezmoi Design FAQ](https://www.chezmoi.io/user-guide/frequently-asked-questions/design/)
- [Command Line Interface Guidelines](https://clig.dev/)
- [Nix and direnv - Determinate Systems](https://determinate.systems/blog/nix-direnv/)
- [Mutagen File Synchronization](https://mutagen.io/documentation/synchronization/)
- [Dotfiles Management Tools Comparison](https://gbergatto.github.io/posts/tools-managing-dotfiles/)

### Codebase References
- `/Users/tomtenuta/Code/roster/ariadne/internal/inscription/generator.go`
- `/Users/tomtenuta/Code/roster/knossos/templates/CLAUDE.md.tpl`
- `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-sync.md`
- `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-knossos-consolidation-architecture.md`

---

## Appendix A: Verification Attestation

| Artifact | Verified Via | Attestation |
|----------|--------------|-------------|
| Current templating implementation | Read tool | `generator.go` uses text/template with custom functions |
| Existing sync TDD | Read tool | Sync domain design exists in `TDD-ariadne-sync.md` |
| Template directory structure | Glob tool | 16 `.tpl` files in `knossos/templates/` |
| ADR format | Read tool | ADR-0012 provides consistent format reference |

---

**Document Status**: SPIKE COMPLETE
**Next Step**: Create ADR-sync-materialization.md (draft)
**Handoff**: Ready for ecosystem rite implementation
