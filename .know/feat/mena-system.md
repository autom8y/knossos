---
domain: feat/mena-system
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/mena/**/*.go"
  - "./internal/materialize/mena/**/*.go"
  - "./mena/**/*.md"
  - "./docs/decisions/ADR-0023*.md"
  - "./docs/decisions/ADR-0025*.md"
  - "./docs/decisions/ADR-0021*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.82
format_version: "1.0"
---

# Mena (Dromena + Legomena) System

## Purpose and Design Rationale

Unified convention for CC context artifacts as first-class source files. **ADR-0021**: Unified skills/commands into single tree. **ADR-0023**: Replaced frontmatter routing with filesystem extension convention (`.dro.md`/`.lego.md`). **ADR-0025**: Added `MenaScope` for pipeline targeting.

### Two-Type Taxonomy

| Type | Extension | CC Target | Context Lifecycle |
|------|-----------|-----------|------------------|
| Dromena | `.dro.md` | `.claude/commands/` | Transient (execute and exit) |
| Legomena | `.lego.md` | `.claude/skills/` | Persistent (stays in context) |

## Conceptual Model

### Source Priority (lowest → highest)

`mena/` (platform) → `rites/shared/mena/` → `rites/{dependency}/mena/` → `rites/{active}/mena/`

### INDEX Files and Namespace Flattening

- `INDEX.dro.md` → promoted to flat file at target (e.g., `.claude/commands/bar.md`)
- `INDEX.lego.md` → renamed to `SKILL.md` at target (e.g., `.claude/skills/bar/SKILL.md`)
- Dromena are namespace-flattened; legomena retain path structure

### Five-Pass Sync Pipeline

1. Collect entries from all sources → 2. Namespace resolution → 3. Filter + flat name assignment → 4. Write directories → 5. Write standalones → 6. Clean stale knossos-owned entries

### Two Projection Modes

- **Additive**: user-scope sync (preserves user-created skills)
- **Destructive**: rite-scope sync (removes stale knossos-owned files via provenance)

## Implementation Map

Leaf package: `/Users/tomtenuta/Code/knossos/internal/mena/` (4 files, zero imports). Materialize sub-package: `/Users/tomtenuta/Code/knossos/internal/materialize/mena/` (6 files).

### Key Boundary Rules

- SCAR-005: Never `os.RemoveAll` on managed directories
- SCAR-006: Satellite rites must use `KnossosHome()` for shared mena base
- SCAR-007: Mixed dro/lego directories → entire directory routes as dromena
- SCAR-018: `context: fork` blocks Task tool access
- SCAR-027: Session artifacts in `rites/shared/mena/` become permanent

## Boundaries and Failure Modes

- `internal/mena` is a leaf package (zero internal imports — enforced)
- Embedded FS standalone files silently ignored
- `os.DirFS` does not follow symlinks
- Frontmatter parse failures return zero-value struct with `log.Printf` warning
- `MenaScope` in ADR-0025 not present in current `MenaFrontmatter` struct (implementation gap)

## Knowledge Gaps

1. `materialize_mena.go` bridge file not read.
2. `MenaScope` implementation gap between ADR-0025 and current code.
3. Usersync pipeline not fully traced.
