---
domain: feat/project-initialization
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/initialize/**/*.go"
  - "./docs/decisions/TDD-single-binary-completion.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.91
format_version: "1.0"
---

# Project Initialization (ari init)

## Purpose and Design Rationale

Before `ari init`, the binary required the Knossos source repository on disk. `go install` produced a non-functional binary. The feature was one of three goals in the "single-binary completion" sprint (TDD at `/Users/tomtenuta/Code/knossos/docs/decisions/TDD-single-binary-completion.md`).

**Key decisions**: Embed rites/templates in binary via `//go:embed` at module root. `needsProject=false` (init creates project context). Idempotency sentinel is `KNOSSOS_MANIFEST.yaml` presence. XDG mena extraction with version sentinel for hybrid distribution.

## Conceptual Model

### Three Modes

| Mode | Trigger | What happens |
|------|---------|--------------|
| `already_initialized` | `KNOSSOS_MANIFEST.yaml` exists | Exit 0, no writes |
| `minimal` | No `--rite` | `MaterializeMinimal()` — CLAUDE.md, settings, manifest only |
| `rite` | `--rite <name>` | Full `Sync()` — agents, mena, hooks, inscription |

### Directory Scaffold

Creates `.knossos/`, `.sos/`, `.ledge/{decisions,specs,reviews,spikes}` with `.gitkeep` files.

## Implementation Map

2 files: `/Users/tomtenuta/Code/knossos/internal/cmd/initialize/init.go`, `init_test.go` (10 test functions). Package named `initialize` because `init` is a Go keyword.

### Key Entry Points

- `NewInitCmd()` — Cobra constructor
- `runInit()` — execution function
- `extractEmbeddedMenaToXDG()` — version-gated XDG extraction
- `scaffoldProjectDirs()` — creates directory tree

## Boundaries and Failure Modes

- Does NOT manage session lifecycle
- Does NOT validate `--rite` name before materialization attempt
- `scaffoldProjectDirs` errors silently swallowed (best-effort)
- Non-Knossos `.claude/` protected: error unless `--force`

## Knowledge Gaps

1. `MaterializeMinimal` behavior not fully traced.
2. `--force` behavior on existing satellite CLAUDE.md regions not verified through merger source.
