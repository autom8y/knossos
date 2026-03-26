---
domain: feat/embedded-assets
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./embed.go"
  - "./internal/assets/**/*.go"
  - "./internal/cmd/common/embedded.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.82
format_version: "1.0"
---

# Single-Binary Embedded Asset Distribution

## Purpose and Design Rationale

Solves bootstrapping: users run ari init immediately after download with no KNOSSOS_HOME. Zero external runtime dependencies. Version coherence (embedded assets match binary version). Hybrid distribution: embedded FS provides canonical source, ari init extracts mena to XDG with version sentinel for upgrade detection. Six //go:embed directives in module-root embed.go.

## Conceptual Model

**Role 1 (fallback source):** EmbeddedRites as lowest-priority tier in resolution chain. fs.Sub creates sub-FS rooted at rite directory. **Role 2 (bootstrap to XDG):** EmbeddedMena and EmbeddedAgents extracted during ari init with .ari-version sentinel. Wipe-and-reextract on version mismatch. **Role 3 (hooks bootstrap):** EmbeddedHooksYAML ([]byte, single file) bootstrapped if missing. **Role 4 (processions):** EmbeddedProcessions for template resolution. **Global state flow:** embed.go -> main.go SetEmbedded* -> internal/assets store -> internal/cmd/common accessors -> Materializer With* methods.

## Implementation Map

`embed.go` (6 exports: EmbeddedRites, EmbeddedTemplates, EmbeddedHooksYAML, EmbeddedAgents, EmbeddedMena, EmbeddedProcessions). `internal/assets/assets.go` (package-level storage + getters/setters + BuildVersion). `internal/cmd/common/embedded.go` (thin pass-through). `internal/materialize/source/resolver.go` (WithEmbeddedFS, riteChain adds embedded tier). `internal/materialize/materialize.go` (riteFS, templatesFS using fs.Sub). `internal/cmd/initialize/init.go` (extractEmbeddedMenaToXDG).

## Boundaries and Failure Modes

embed.go MUST live at module root (//go:embed paths are relative). SourceEmbedded skips platform mena injection (prevents developer-local leaks). EmbeddedHooksYAML is write-once (no version-aware upgrade). No embedded FS in rite.Discovery (only sync/init use embedded). assets.SetEmbedded not called -> embedded tier silently absent. fs.Sub failure -> falls back to os.DirFS (likely fails on FS-relative path). XDG RemoveAll failure -> warning, stale mena retained. Operational trap: rebuilt binary vs PATH binary divergence.

## Knowledge Gaps

1. TDD-single-binary-completion.md not found on disk
2. EmbeddedAgents user-scope consumption path not traced
3. resolution.RiteChain constructor signature inferred from call sites
