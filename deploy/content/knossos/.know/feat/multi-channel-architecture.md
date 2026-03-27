---
domain: feat/multi-channel-architecture
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/paths/channel.go"
  - "./internal/materialize/compiler/**/*.go"
  - "./internal/hook/adapter*.go"
  - "./internal/hook/events.go"
  - "./internal/channel/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Multi-Channel Harness Architecture (Claude + Gemini)

## Purpose and Design Rationale

Projects the same rite definitions into any supported AI assistant's config directory. ADR-0031 established the interface-based architecture. ADR-0032 replaced CC-canonical internal naming with knossos-owned snake_case vocabulary (pre_tool, run_shell) -- CC and Gemini are now translation peers. Adding a third channel requires implementing TargetChannel (5 methods) + ChannelCompiler (3 methods) + translation tables. Zero core pipeline changes.

## Conceptual Model

**Three interface layers:** TargetChannel (identity: DirName, ContextFile), ChannelCompiler (format: commands to markdown/TOML, agent key stripping), LifecycleAdapter (hook payload: wire-to-canonical translation). **Canonical vocabulary:** 18 lifecycle events, 11 tool concepts. **Dispatch:** `--channel=all` iterates AllChannels() with channelDirOverride save-and-restore pattern. **Per-channel provenance:** PROVENANCE_MANIFEST.yaml (claude) vs PROVENANCE_MANIFEST_GEMINI.yaml.

## Implementation Map

`internal/paths/channel.go` (TargetChannel, ClaudeChannel, GeminiChannel), `internal/materialize/compiler/` (ChannelCompiler, ClaudeCompiler, GeminiCompiler), `internal/hook/adapter*.go` (LifecycleAdapter, ClaudeAdapter, GeminiAdapter), `internal/hook/events.go` (canonical<->wire maps), `internal/channel/tools.go` (tool translation).

## Boundaries and Failure Modes

channelDirOverride save-and-restore is pragmatic hack (not goroutine-safe). GeminiCompiler strips unknown keys silently. Partial tool translation for hook matchers (CC-only tools pass through). ReadFiles wire alias not in CanonicalTool. Budget report hardcodes .claude. AGENTS.md interop target documented but not implemented.

## Knowledge Gaps

1. compilerForChannel() not read in detail
2. BuildHooksSettings() outbound translation not traced
3. AGENTS.md implementation status unconfirmed
