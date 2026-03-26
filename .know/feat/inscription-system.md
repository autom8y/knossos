---
domain: feat/inscription-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/inscription/**/*.go"
  - "./internal/cmd/inscription/**/*.go"
  - "./knossos/templates/sections/**/*.tpl"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# CLAUDE.md / GEMINI.md Inscription System

## Purpose and Design Rationale

Generates and maintains AI harness context files with region-based ownership: knossos regions (always overwritten from templates), satellite regions (never overwritten -- user content survives), regenerate regions (rebuilt from live project state). Markers use HTML comments (invisible in rendered markdown). SHA-256 hashes detect user edits to knossos regions. writeIfChanged prevents CC file watcher triggers (LB-001). Legacy migration backs up unmarked files. Multi-channel via RenderContext.Channel and template branching.

## Conceptual Model

**Three ownership types:** knossos (platform), satellite (user), regenerate (dynamic). **9 default sections:** execution-mode, model-override, quick-start, agent-routing, commands, agent-configurations, platform-infrastructure, know, user-content. **KNOSSOS_MANIFEST.yaml** tracks regions, hashes, ordering, inscription_version. **Three components:** MarkerParser (line-by-line scanner with code-block escape), Generator (Go templates + Sprig + 4 custom functions), Merger (ownership-respecting merge with double-wrap prevention). **Two deprecated regions** (slash-commands v18, navigation v20).

## Implementation Map

`internal/inscription/` (8 files): types.go (OwnerType, Region, Manifest), marker.go (MarkerParser), manifest.go (ManifestLoader, DefaultSectionOrder, AdoptNewDefaults), generator.go (Generator, RenderContext), merger.go (Merger, conflict detection), sync.go (SyncInscription -- canonical 7-step entry point), pipeline.go (standalone CLI path), backup.go (5-backup retention). `internal/cmd/inscription/` (6 files): sync, diff, validate, rollback, backups. `knossos/templates/sections/` (8 .md.tpl files).

## Boundaries and Failure Modes

Malformed marker pairs: ParseError, region not extracted. Hash mismatch without hash: first sync silently overwrites. Double-wrap prevention via isWrapped() check. Deprecated region zombie adoption prevented by static list + dynamic check. TENSION-013: Pipeline hardcodes .claude/CLAUDE.md (harness-agnosticism gap). Template rendering failure: non-satellite regions silently skipped if CanGenerateRegion returns false. No concurrent sync locking. Legacy backup uses fixed filename (not timestamped).

## Knowledge Gaps

1. ADR-0021 not found on disk
2. Materialization pipeline integration depth inferred from architecture doc
3. Conditional region evaluation may be defined but not implemented
