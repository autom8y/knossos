---
domain: feat/materialization-pipeline
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/materialize/**/*.go"
  - "./internal/cmd/sync/**/*.go"
  - "./docs/decisions/ADR-sync*.md"
  - "./docs/decisions/TDD-single-binary-completion.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Rite Materialization Pipeline

## Purpose and Design Rationale

Core engine converting declarative rite definitions into `.claude/` directory structure. Makes `.claude/` a fully derived artifact regenerated from canonical sources via `ari sync`.

**ADR-0016**: Chezmoi-inspired generation model. Generation over symlinks. Single idempotent command. Crash-only design. **ADR-0026**: Unified provenance model. **TDD-single-binary**: Binary embeds all assets.

## Conceptual Model

### 11-Stage Materialization Lifecycle

0. Pre-validate CLAUDE.md → 1. Resolve rite (6-tier) → 2. Load provenance → 2.5. Clear stale invocations → 3. Handle orphan agents → 3.5. Resolve hook/skill defaults → 4. Materialize agents → 5. Materialize mena → 6. Materialize rules → 7. Materialize CLAUDE.md → 8. Materialize settings → 9. Track state → 9.5. Materialize workflow → 10. Write ACTIVE_RITE → 11. Save provenance

### Three Sync Phases

Phase 1 (rite scope) → Phase 1.5 (org scope) → Phase 2 (user scope)

### 6-Tier Source Resolution

1. Explicit → 2. Project satellite → 3. User → 4. Org → 5. Knossos platform → 6. Embedded

### Key Modes

- **Soft mode**: Skips mena/rules/settings/workflow (CC-safe live update)
- **El-cheapo mode**: Forces haiku model on all agents (ephemeral)
- **Minimal mode**: Cross-cutting, no rite (CLAUDE.md + settings only)

## Implementation Map

Hub package: `/Users/tomtenuta/Code/knossos/internal/materialize/` (48 files). Sub-packages: `source/`, `mena/`, `hooks/`, `userscope/`, `orgscope/`. CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/sync/`.

### Key Types

`Materializer`, `RiteManifest`, `Options` (Soft, ElCheapo, DryRun, OverwriteDiverged), `SyncOptions` (Scope, RiteName), `SourceResolver`, `ResolvedRite`.

### 26 Test Files

Most heavily tested package. SCAR regression tests cover SCAR-002, 003, 004, 005, 006, 008, 015, 016, 018, 020, 021, 023, 024, 027.

### Load-Bearing Code

- **LOAD-001**: `provenance.Save()` structural equality guard
- **LOAD-002**: `fileutil.WriteIfChanged()` for all writes
- **LOAD-003**: `inscription.MergeRegions()` satellite preservation
- **LOAD-004**: `namespace.resolveNamespace()` provenance-based collision detection

## Boundaries and Failure Modes

- Does NOT manage session state, rite workflows, or remote sources
- SCAR-002: Never rename `.claude/`
- SCAR-004: Provenance parse errors abort, never bootstrap silently
- SCAR-005: Never `os.RemoveAll` on agents/commands/skills
- SCAR-006: Satellite rites use `KnossosHome()` for shared mena
- SCAR-024: Rite switch cleans stale throughline IDs
- RISK-002: `materializeMena()` discards collision warnings

## Knowledge Gaps

1. `userscope/sync.go` and `orgscope/sync.go` not read in full.
2. `materialize_agents.go` selective write logic not read.
3. `ACTIVE_WORKFLOW.yaml` format not examined.
