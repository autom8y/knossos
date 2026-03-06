---
domain: feat/embedded-assets
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./embed.go"
  - "./internal/assets/**/*.go"
  - "./internal/materialize/source/**/*.go"
  - "./docs/decisions/TDD-single-binary-completion.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.86
format_version: "1.0"
---

# Single-Binary Embedded Asset Distribution

## Purpose and Design Rationale

Makes `ari` self-contained: the binary carries all rites, templates, hooks config, cross-rite agents, and platform mena compiled in at build time. Eliminates dependency on `$KNOSSOS_HOME` source repository.

**TDD**: `/Users/tomtenuta/Code/knossos/docs/decisions/TDD-single-binary-completion.md`

**Key decisions**: `embed.go` at module root as `package knossos`. Five asset classes. ~2MB binary size increase (LOW risk). Hybrid distribution: XDG extraction on `ari init` for filesystem performance on subsequent syncs.

## Conceptual Model

### Five Asset Classes

| Variable | Source | Type | Role |
|---|---|---|---|
| `EmbeddedRites` | `rites/` | `embed.FS` | All rite definitions |
| `EmbeddedTemplates` | `knossos/templates/` | `embed.FS` | CLAUDE.md section templates |
| `EmbeddedHooksYAML` | `config/hooks.yaml` | `[]byte` | Hook configuration |
| `EmbeddedAgents` | `agents/` | `embed.FS` | Cross-rite agents |
| `EmbeddedMena` | `mena/` | `embed.FS` | Platform mena |

### Six-Tier Resolution (Embedded = Tier 6)

1. Explicit → 2. Project satellite → 3. User → 4. Org → 5. Knossos platform → 6. **Binary embedded**

### Wiring Chain

`embed.go` → `main.go` → `common.SetEmbeddedAssets()` → `assets.SetEmbedded()` → `Materializer.WithEmbeddedFS()`

## Implementation Map

Core files: `embed.go`, `internal/assets/assets.go`, `internal/cmd/common/embedded.go`, `cmd/ari/main.go`. Resolution: `internal/materialize/source/resolver.go` (tier 6 `checkEmbeddedSource()`). Pipeline: `materialize.go` (`riteFS()`, `templatesFS()`, `copyDirFromFS()`).

## Boundaries and Failure Modes

- Embedded content is a compile-time snapshot (stale across binary versions)
- `riteFS()` fallback silently degrades to nonexistent OS path on `fs.Sub` failure
- XDG `RemoveAll` on version upgrade can destroy user-created files without provenance guard
- `KnossosHome()` default `~/Code/knossos` can shadow embedded tier unexpectedly

## Knowledge Gaps

1. XDG `RemoveAll` user-file destruction risk undocumented.
2. `riteFS` error path untested.
3. TDD Task 3 (shell cleanup) completion status unverified.
