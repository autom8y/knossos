# SPIKE: .knossos/.know/.ledge/.sos/ Directory Architecture

**Date**: 2026-03-01
**Session**: session-20260301-230114-075ad051
**Rite**: hygiene (scoping audit)
**Status**: COMPLETE

## Executive Summary

This spike audits the surface area required to integrate four semantically-derived directories into the Knossos framework, decomposing the project name into architecture:

```
KNOSSOS = .knossos/ + .know/ + .ledge/ + .sos/
          config     observe   produce   survive
```

**Finding: The codebase is remarkably well-prepared for this change.** Path resolution is centralized through `paths.Resolver`, materialization uses relative paths after initial resolution, and `.know/` already exists with zero changes needed. The total code change is estimated at ~15 production code lines + ~80 test fixture lines.

---

## The Architecture

| Directory | Role | Managed by | Lifecycle | Git |
|-----------|------|------------|-----------|-----|
| `.knossos/` | Framework configuration source | Human + `ari` | Edit here | Committed |
| `.claude/` | CC runtime output | Materialization pipeline | Never edit directly | Mixed |
| `.know/` | Codebase observations | Theoros | Regenerable | Gitignored |
| `.ledge/` | Work product artifacts | Rites | Append-only | Committed |
| `.sos/` | Session or State | Moirai | Mutable | Committed + pruned |

### Name Decomposition

```
K  +  KNOW  +  LEDGE  +  SOS
      observe   produce   survive

.knossos/  →  the labyrinth itself (configuration)
.know/     →  what we KNOW (observations, regenerable)
.ledge/    →  the LEDGE we shelve to (decisions, specs, reviews)
.sos/      →  Session Or State / Save Our Souls (the clew)
```

### Materialization Flow

```
ari init         →  .knossos/  (scaffolded)
ari materialize  →  .claude/   (generated from .knossos/)
/know            →  .know/     (observed from source)
/start, /wrap    →  .sos/      (session lifecycle)
rite work        →  .ledge/    (graduated artifacts)
```

---

## Surface Area Audit Results

### 1. `.knossos/` — Framework Configuration Consolidation

**What moves**: Root-level `rites/`, `mena/`, `agents/` → `.knossos/rites/`, `.knossos/mena/`, `.knossos/agents/`

**Production code changes (5 files, ~12 lines)**:

| File | Line(s) | Change | Impact |
|------|---------|--------|--------|
| `embed.go` | 15, 21, 33, 39 | Update 4 `go:embed` directives | Binary recompilation |
| `internal/paths/paths.go` | 150-152 | `RitesDir()`: `"rites"` → `".knossos/rites"` | All rite discovery |
| `internal/materialize/source/resolver.go` | 32 | `projectRitesDir`: `"rites"` → `".knossos/rites"` | Rite resolution |
| `internal/materialize/materialize.go` | 1104 | `projectMena`: `"mena"` → `".knossos/mena"` | Mena discovery |
| `internal/materialize/materialize.go` | 117, 236 | `templatesDir` references | Template resolution |

**Key finding**: The 4-tier source resolution chain (project → user → knossos → embedded) continues to work — only tier 1 (project) changes its root prefix.

**Template changes**: `knossos/templates/sections/commands.md.tpl` — update source column in CC Primitives table.

**Test fixtures**: ~10 test files create mock `rites/` and `mena/` directories that need path updates (~50 lines).

**Risk**: LOW. All paths flow through `Resolver` or `SourceResolver`. No hardcoded paths in business logic.

---

### 2. `.know/` — Codebase Observations (NO CHANGES)

**Current state**: Fully functional, independent subsystem.

**Infrastructure**:
- `internal/know/` — 4 Go files (parsing, validation, freshness)
- `internal/cmd/knows/` — CLI command (`ari knows`)
- `internal/cmd/hook/context.go` — Session hook freshness injection
- 2 dromena (`/know`, `/research`) for generation

**Path resolution**: Always `filepath.Join(projectDir, ".know")` — hardcoded at 2 call sites, NOT through `paths.Resolver`. This is intentional.

**Finding**: `.know/` is a mature, isolated subsystem. Zero structural changes needed. Already gitignored, already at project root, already schema-validated with frontmatter.

---

### 3. `.ledge/` — Work Product Artifacts (NEW DIRECTORY)

**Current problem**: Artifacts scatter across:
- `.sos/wip/` (gitignored — good artifacts disappear)
- `docs/` (10x-dev dumps here with no structure)
- `SESSION_CONTEXT.md` body (decisions buried in conversation history)

**Proposed structure**:
```
.ledge/
  decisions/     # ADRs (currently in docs/decisions/, 23 total)
  specs/         # Initiative specifications
  reviews/       # Audit findings, code review artifacts
  handoffs/      # Cross-rite transfer documents
  spikes/        # Spike conclusions (graduated from .wip/)
```

**Code changes needed**: MINIMAL
- New `internal/paths` methods: `LedgeDir()`, `LedgeDecisionsDir()`, etc.
- New `ari init` scaffold step to create `.ledge/` on project bootstrap
- Template update: `know.md.tpl` to reference `.ledge/` for agents
- Optional: `internal/cmd/ledge/` CLI command for artifact management

**Migration**: `docs/decisions/` → `.ledge/decisions/` (23 ADR files, one-time move)

**Risk**: LOW. New directory, no existing code to break. Migration is optional.

---

### 4. `.sos/` — Session Or State (SESSION MIGRATION)

**What moves**: `.sos/sessions/` → `.sos/sessions/`, `.sos/archive/` → `.sos/archive/`

**Production code changes (1 file, ~4 methods)**:

| File | Method | Current | Proposed |
|------|--------|---------|----------|
| `internal/paths/paths.go` | `SessionsDir()` | `.sos/sessions` | `.sos/sessions` |
| `internal/paths/paths.go` | `LocksDir()` | `.sos/sessions/.locks` | `.sos/sessions/.locks` |
| `internal/paths/paths.go` | `CCMapDir()` | `.sos/sessions/.cc-map` | `.sos/sessions/.cc-map` |
| `internal/paths/paths.go` | `ArchiveDir()` | `.claude/.archive/sessions` | `.sos/archive` |

**Key finding**: ALL 175+ session path references across 52 files flow through these 4 `Resolver` methods. Changing them propagates everywhere automatically.

**Dependent systems (no code changes, automatic propagation)**:
- Session lifecycle: create, park, resume, wrap, gc, fray (~12 cmd files)
- Hook infrastructure: writeguard, autopark, sessionend, clew (~6 hook files)
- Lock management: advisory locks, Moirai locks (~2 lock files)
- Event logging: clewcontract BufferedEventWriter (~2 files)
- Scanner: naxos orphan scanning (~1 file)
- Discovery: FindActiveSession filesystem scan (~1 file)

**CC session map**: The `.cc-map/` subdirectory maps CC conversation IDs to Knossos session IDs. This moves with the sessions directory — no special handling needed.

**Test fixtures**: ~30 lines of test path updates across session/hook test files.

**Risk**: LOW-MEDIUM. Single point of change, but session state is critical path. Requires migration command for existing sessions.

---

## Consolidated Change Matrix

### By Risk Level

| Risk | Component | Lines Changed | Files Touched |
|------|-----------|---------------|---------------|
| NONE | `.know/` (no changes) | 0 | 0 |
| LOW | `.ledge/` (new directory) | ~20 new | 3-4 new |
| LOW | `.knossos/` (source consolidation) | ~12 changed | 5 production + 10 test |
| LOW-MED | `.sos/` (session migration) | ~4 changed | 1 production + 8 test |

### By Implementation Phase

**Phase 1: `.ledge/` (Zero Risk)**
- New directory, no existing code affected
- Add `paths.Resolver` methods
- Update `ari init` scaffold
- Update CLAUDE.md template
- Migrate `docs/decisions/` → `.ledge/decisions/`

**Phase 2: `.sos/` (Low Risk, High Value)**
- Update 4 `Resolver` methods in `paths.go`
- 175+ references auto-propagate
- Write `ari session migrate-storage` command
- Update tests

**Phase 3: `.knossos/` (Moderate Complexity)**
- Update `embed.go` directives (requires binary rebuild)
- Update `SourceResolver` and `Materializer` paths
- Move `rites/`, `mena/`, `agents/` to `.knossos/`
- Update all test fixtures
- Add backward-compat fallback: check `.knossos/rites` first, fall back to `./rites`

**Phase 4: Templates & Documentation**
- Update `know.md.tpl`, `commands.md.tpl`, `user-content.md.tpl`
- Update MEMORY.md golden rules
- Write ADR documenting the directory architecture decision
- Update `ari init` to create all four directories

---

## Architectural Considerations

### What MUST NOT change
- `.claude/` as CC runtime target (protocol invariant — `FindProjectRoot()` looks for `.claude/`)
- Materialization pipeline stages and ordering
- Region ownership semantics (knossos/satellite/regenerate)
- Session FSM states and transitions
- Event schema (clewcontract)

### What the codebase already gets right
- **Path centralization**: `paths.Resolver` is the single authority — no scattered hardcoded paths
- **Relative resolution**: After initial rite discovery, everything resolves relative to the rite path
- **Atomic writes**: All state mutations use `fileutil.AtomicWriteFile`
- **Scan-based discovery**: Session finding scans filesystem, not in-memory cache — directory location is transparent

### Open Questions

1. **`.knossos/` for the knossos repo itself**: In the framework's own repo, both `knossos/` (Go source) and `.knossos/` (project config) would exist. Is this confusing or actually clarifying?

2. **`.sos/` git status**: Should `.sos/` be fully committed, partially committed, or have its own gitignore rules? Active sessions are mutable, archived sessions are historical.

3. **Backward compatibility period**: How long do we support `./rites/` alongside `.knossos/rites/`? Suggestion: one minor version with fallback + deprecation warning.

4. **`.ledge/` schema**: Should `.ledge/` artifacts have validated frontmatter (like `.know/`) for provenance tracking? Lean yes — rite, date, status, author.

5. **`ari init` evolution**: Should `ari init` create all four directories upfront, or lazily create them as features are used?

---

## Conclusion

The Knossos codebase is **architecturally prepared** for this directory restructuring. The framework's own design principles — centralized path resolution, atomic writes, scan-based discovery, and clean layer boundaries — make the migration surface area remarkably small.

The name decomposition (`KNOSSOS = .knossos + .know + .ledge + .sos`) transforms from a clever pun into **load-bearing architecture** where every directory is self-documenting, every concept has a canonical home, and the project name IS the directory listing.

**Recommended next step**: Write ADR-0029 codifying this architecture, then implement Phase 1 (`.ledge/`) as the zero-risk proof of concept.
