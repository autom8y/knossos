---
domain: feat/materialization-pipeline
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/materialize/**/*.go"
  - "./internal/cmd/sync/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.82
format_version: "1.0"
---

# Rite Materialization Pipeline

## Purpose and Design Rationale

Converts rite source definitions into channel-ready configuration directories (.claude/ or .gemini/). Three non-negotiable invariants: idempotency (writeIfChanged guards every write, preventing CC file watcher crashes), user content preservation (satellite regions, user agents never destroyed), harness agnosticism (ChannelCompiler interface projects same source to any channel). Provenance lives in .knossos/ not .claude/ (keeps it out of CC context window). Embedded assets enable single-binary distribution.

## Conceptual Model

**Three-scope architecture:** Rite (project .claude/), Org (user ~/.claude/), User (user ~/.claude/). **6-tier source resolution:** explicit > project > user > org > platform > embedded. **Three agent tiers:** standing (always), rite (per-sync), summonable (on-demand). **Mena priority:** platform < shared < dependency < active rite. **Provenance:** two YAML manifests (rite + user scope), SHA-256 checksums, orphan detection. **Collision checker** prevents user-scope shadowing rite-owned resources.

## Implementation Map

CLI: `internal/cmd/sync/sync.go` + `budget.go`. Core: `internal/materialize/materialize.go` -- MaterializeWithOptions() 20+ step pipeline (resolve rite, validate, load manifest, ensure dirs, load provenance, detect divergence, handle orphans, resolve hooks/skills/MCP, materializeAgents, materializeMena, materializeRules, materializeInscription, materializeSettings, materializeMcp, track state, materialize workflow, write ACTIVE_RITE, save provenance, generate .gitignore, untrack). Agent transform: strip knossosOnlyFields, inject defaults, apply skill policies, inject MCP servers. Mena engine: 4-pass algorithm (collect, namespace resolve, apply flat names, write). Compiler: ClaudeCompiler (pass-through) vs GeminiCompiler (TOML commands, key-stripped agents).

## Boundaries and Failure Modes

Fatal: source resolution failure, manifest parse failure, CLAUDE.md pre-validation failure. Non-fatal: rite reference warnings, divergence detection, org/user scope failures in scope=all. Idempotency: writeIfChanged (LB-001), no pre-delete (SCAR-005/LB-007), selective write from managed-set. Non-transactional pipeline (RZ-001). channelDirOverride save-and-restore (TENSION-002). Soft mode skips mena/rules/settings. El-cheapo injects haiku model override.

## Knowledge Gaps

1. procession sub-package not read
2. User-scope mena sync details not fully verified
3. Inscription full merge pipeline not traced
4. ADR-0016 not found on disk
