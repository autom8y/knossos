---
domain: feat/mena-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/mena/**/*.go"
  - "./internal/materialize/mena/**/*.go"
  - "./mena/**/*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Mena (Dromena + Legomena) Distribution System

## Purpose and Design Rationale

Different content types have different context lifecycles: dromena (commands, transient) vs legomena (skills, persistent). Extension infix (.dro.md/.lego.md) is the routing signal -- eliminates frontmatter inspection during routing. Two-level package: internal/mena (leaf, zero imports) for primitives, internal/materialize/mena (hub) for full projection. Priority bottom-up: platform < shared < dependency < active rite. INDEX-file pattern with companion files for progressive disclosure. Namespace flattening for user-friendly /name invocation.

## Conceptual Model

**Two types:** Dromenon (.dro.md -> commands/, transient) vs Legomenon (.lego.md -> skills/, persistent). **Four-level source priority:** active rite > dependency rite > shared rite > procession > platform. **Three entry structures:** directory (INDEX file + companions), standalone file, grouping directory. **6-pass projection pipeline:** collect -> namespace resolve -> apply flat names -> write (with transforms) -> clean stale -> reconcile untracked. **Two modes:** destructive (rite scope, wipes stale) and additive (user scope, preserves). **Content rewriting:** extension refs + channel path substitution (4-pass, code-block safe).

## Implementation Map

Leaf: `internal/mena/` (types.go, source.go, exists.go, walk.go). Hub: `internal/materialize/mena/` (types.go, frontmatter.go, collect.go, namespace.go, engine.go, walker.go, transform.go, content_rewrite.go). Integration: `internal/materialize/materialize_mena.go` (materializeMena, materializeMinimalMena, renderProcessionMena). MenaFrontmatter schema: name (required), description (required), argument-hint, triggers, allowed-tools, model, etc. ChannelCompiler interface for harness-specific output (TOML for Gemini commands).

## Boundaries and Failure Modes

Mixed dro/lego directory blocks skill resolution (documented in MEMORY.md). Walk does NOT support embedded FS (lint operates on filesystem only). Silent skips for nonexistent sources. Namespace collision falls back to directory-path routing with warning. renderProcessionMena is fail-open. Stale cleanup requires provenance (first sync after rite switch can't clean). Standalone files default to "dro" type. Compiler applied only to primary files (companions pass through as-is).

## Knowledge Gaps

1. renderProcessionMena rendering details not traced
2. ADRs 0021, 0023, 0025 not found on disk
3. Gemini channel compiler TOML format not read
