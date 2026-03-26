---
domain: feat/project-initialization
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/initialize/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Project Initialization (ari init)

## Purpose and Design Rationale

Bootstraps new project into knossos workflow. Only command with SetNeedsProject(false,false). Two modes: minimal scaffold (CLAUDE.md + settings + directories) and rite scaffold (full ari sync). Non-knossos .claude/ guard (rejects if .claude/ exists without knossos manifest, unless --force). XDG mena extraction as side effect (version-sentinel controlled). Single-binary distribution (works without KNOSSOS_HOME via embedded assets).

## Conceptual Model

**Three modes:** already initialized (exit 0 no-op), minimal scaffold (MaterializeMinimal), rite scaffold (Sync with ScopeAll + KeepOrphans). **scaffoldProjectDirs creates:** .knossos/, .sos/, .sos/land/, .ledge/{decisions,specs,reviews,spikes}, .ledge/shelf/, .gitignore with Knossos marker block. **extractEmbeddedMenaToXDG:** version-sentinel (.ari-version), wipe-and-reextract on version mismatch. **Gitignore management:** surgically replaces Knossos marker block, preserves user content.

## Implementation Map

`internal/cmd/initialize/init.go`: runInit (project dir resolution, idempotency check, non-knossos guard, materializer construction with embedded FS wiring, hooks.yaml bootstrap, XDG extraction, scaffoldProjectDirs, branch to minimal or rite sync). Tests cover fresh dir, with rite, already initialized, force, non-knossos guard, ledge idempotent, XDG extraction (4 paths), gitignore (4 scenarios).

## Boundaries and Failure Modes

Channel hardcoded to .claude (HA-CC annotation). hooks.yaml write-once (no version-aware upgrade). extractEmbeddedMenaToXDG is best-effort (slog.Warn on failure). scaffoldProjectDirs mkdir failures silent. Corrupted KNOSSOS_MANIFEST.yaml propagates error from LoadOrBootstrap. KeepOrphans=true means stale artifacts from prior rite not cleaned on --force re-init.

## Knowledge Gaps

1. --source flag non-existent path behavior not traced
2. config/hooks.yaml content not examined
3. No user-facing documentation for ari init beyond --help
