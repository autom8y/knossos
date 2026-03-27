---
domain: design-constraints
generated_at: "2026-03-26T17:14:25Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "a73d68a6"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "77ed0121d9982cb4e1bf8e7c09b414494a561c0ccd3b9ea546d210acd7354553"
---

# Codebase Design Constraints

## Tension Catalog Completeness

17 TENSION-NNN entries cataloged. 16 accurate, 1 stale-resolved (TENSION-015).

**TENSION-001** (`internal/provenance/provenance.go:91`, `internal/inscription/types.go:14`): Dual `OwnerType` types in separate packages — distinct types with same name, cross-referenced by comments.

**TENSION-002** (`internal/materialize/materialize.go:118,284-289,390-391`): `channelDirOverride` mutation-with-defer pattern — saves and restores channel dir override to support multi-channel materialization.

**TENSION-003** (`internal/perspective/context.go:60`, `internal/perspective/resolvers.go:274`): Hardcoded `.claude/` paths marked with `// HA-FS:` annotations — blocks Gemini feature parity.

**TENSION-004** (`internal/agent/frontmatter.go:38-43`): `maxTurns`, `disallowedTools`, `permissionMode`, `mcpServers` as direct struct fields alongside `Contract *BehavioralContract` which also contains `MaxTurns`.

**TENSION-005** (`internal/resolution/chain.go:5-6`): Zero-import invariant — "This package has ZERO internal imports. All tier paths are injected via constructor to avoid import cycles (TENSION-005)."

**TENSION-006** (`internal/provenance/provenance.go:69-70`): String values must stay in sync with `internal/materialize/source/types.go` — manual synchronization without shared constants.

**TENSION-007** (`internal/hook/env.go:86-93`): `GetAdapter()` defaults to ClaudeAdapter when `KNOSSOS_CHANNEL` is unset — implicit default.

**TENSION-008** (`internal/materialize/sync_types.go:51`): Sync state types coupled to materialization pipeline.

**TENSION-009** (`internal/materialize/userscope/sync.go`): Partially stale — mechanism changed to `paths.UserChannelDir(params.Opts.Channel)` dynamic resolution, but fallback behavior when channel is unconfigured remains a risk.

**TENSION-010** (`internal/agent/frontmatter.go:55-70`): Agent archetype fields duplicated across frontmatter and behavioral contract.

**TENSION-011** (`internal/materialize/materialize.go:69-70`): `Skills []string` (deprecated) and `Commands []string` (backward compat) both present in `RiteManifest`.

**TENSION-012** (`internal/hook/output.go:28`): `HookEventName: "PreToolUse"` hardcoded wire value — CC reads this exact string for security decisions.

**TENSION-013** (`internal/inscription/pipeline.go:156-163`): Harness-agnosticism gap in inscription pipeline.

**TENSION-014** (`internal/tribute/types.go:58,158`, `internal/tribute/renderer.go:64,250`): "Phase 2 - placeholder" comments — tribute feature incomplete.

**TENSION-015** (`internal/search/collectors.go:12`): **RESOLVED** — `internal/search` no longer imports `internal/cmd/explain`. Resolved via `internal/concept` extraction. `concept/concept.go` documents: "This package was extracted from internal/cmd/explain to resolve TENSION-015."

**TENSION-016** (`internal/config/home.go:11-23`): `sync.Once` singleton for KnossosHome — `ResetKnossosHome()` exists with test-only warning: "RISK-003: KnossosHome is cached via sync.Once."

**TENSION-017** (`internal/materialize/materialize.go:579-584`): Settings to channel dir, MCP to project root — split output targets.

---

## Trade-off Documentation

**Trade-off 1 (WriteIfChanged):** CC file watcher behavior requires avoiding unnecessary file writes. `writeIfChanged()` compares content before writing. Current state persists because CC file watcher crashes are worse than the overhead.

**Trade-off 2 (Dual read path, v1/v2/v3):** `internal/session/events_read.go:82-117` — dual-format event reader. Persists until "all pre-ADR-0027 sessions are archived."

**Trade-off 3 (Provenance in `.knossos/`):** Provenance manifests stored in `.knossos/` not `.claude/` — CC context window exposure risk.

**Trade-off 4 (CC event names as lingua franca):** ADR-0032 confirmed — rejected a third vocabulary. CC wire names are the canonical representation.

**Trade-off 5 (XDG logic duplicated):** `config.ActiveOrg()` at `home.go:84` duplicates path resolution — import cycle constraint prevents sharing with paths package.

**Trade-off 6 (Two-manifest architecture):** Rite-scope and user-scope provenance manifests are separate to avoid cross-contamination during sync.

**Trade-off 7 (channelDirOverride save-and-restore):** ADR-0031 acknowledges as "pragmatic hack." Constructor refactor would affect 15+ call sites.

**Trade-off 8 (ClaudeCompiler as pass-through):** Symmetric pipeline rejected — Gemini needs active compilation, Claude is mostly pass-through.

---

## Abstraction Gap Mapping

**AGM-001** (`internal/perspective/`): Hardcoded `.claude/` paths with `// HA-FS:` markers. Blocks Gemini feature parity.

**AGM-002** (`internal/agent/frontmatter.go`): `MaxTurns int` at direct field level coexists with `Contract *BehavioralContract` containing `MaxTurns`.

**AGM-003** (`AGENTS.md` compilation): No AGENTS.md (OpenAI/ChatGPT format) compilation exists in materialize pipeline.

**AGM-004** (`BehavioralContract.MaxTurns` not wired): Contract MaxTurns not plumbed through to agent output.

**AGM-005** (`RiteManifest.Commands` and `.Skills`): Both deprecated fields still present at `materialize.go:69-70`.

**AGM-006** (Duplicate resolution logic): Both `materialize/source` and `resolution` packages implement tier traversal. Moderate duplication.

**AGM-007** (`tribute/Commits` placeholder): Type surface with "Phase 2 - placeholder" — feature incomplete.

**AGM-008** (`fileutil.AtomicWriteFile` vs. `os.WriteFile`): `userscope/sync.go` uses `os.WriteFile` while most of the pipeline uses `AtomicWriteFile`. Intentional but asymmetric.

---

## Load-Bearing Code Identification

**LB-001** (`internal/fileutil/fileutil.go:66-72`): `WriteIfChanged()` — naive replacement with `os.WriteFile` causes CC file watcher to crash.

**LB-002** (`internal/provenance`): `structurallyEqual()` — prevents unnecessary writes that would trigger CC file watcher.

**LB-003** (`internal/materialize/mena/content_rewrite.go:128-138`): Three-pass rewrite ordering (INDEX.lego.md -> SKILL.md precedes general `.lego.md -> .md`). Wrong order causes double-rewrite corruption.

**LB-004** (`internal/session/events_read.go:82-117`): Dual-format event reader. Wrong format order causes silent event misread.

**LB-005** (`internal/hook/clewcontract/type_rename.go:14-22`): Append-only event type rename map. Removing entries breaks backward compatibility.

**LB-006** (`internal/config/home.go:11-23`): `sync.Once` singleton for KnossosHome. `ResetKnossosHome()` only for tests.

**LB-007** (`internal/materialize/materialize_agents.go:42-45`): No pre-delete before agent overwrite — CC watcher fires DELETE events that cause agent disappearance.

**LB-008** (`internal/inscription`): Satellite region preservation in `Merger` — core invariant protecting user content from being destroyed during CLAUDE.md regeneration.

**LB-009** (`internal/provenance`): `LoadOrBootstrap()` aborts on corrupt manifests — not fail-open.

**LB-010** (`internal/hook/output.go:28`): `HookEventName: "PreToolUse"` hardcoded wire format string — security boundary. CC reads this exact string for permission decisions.

**LB-011** (`internal/materialize/materialize.go:457-627`): Pipeline step ordering — non-transactional, partial failures possible.

---

## Evolution Constraint Documentation

### SAFE (local change only)
- `internal/resolution/chain.go` — zero-import policy, self-contained
- `internal/tribute/types.go` — placeholder types, no external callers
- `internal/concept/concept.go` — isolated concept registry
- `internal/tokenizer` — stdlib wrapper
- `internal/fileutil` — utility functions
- `internal/checksum` — hash utilities

### COORDINATED (multi-file, no external break)
- Adding a new `TargetChannel`: minimum 5 files across 4 packages
- New hook handler: handler file + hooks.yaml + test + CLAUDE.md regeneration
- New agent frontmatter field: `agent/frontmatter.go` + transform + test + lint rule
- New session state: `status.go` FSM + `NormalizeStatus()` + migrate
- New mena type: `mena/source.go` routing + materialize pipeline + lint

### MIGRATION (breaking change to callers)
- Provenance schema version bump (`CurrentSchemaVersion = "2.0"`) — requires migration function
- Session context schema change — requires `internal/cmd/session/migrate.go` entry
- Inscription schema version — `DefaultSchemaVersion = "1.0"`
- Deprecated field removal (`RiteManifest.Skills`, `.Commands`) — multi-rite impact

### FROZEN (do not touch without explicit decision)
- `internal/hook/output.go` `HookEventName: "PreToolUse"` — CC security boundary
- `.claude/` directory name — CC hardcoded expectation
- `PROVENANCE_MANIFEST.yaml` filename — referenced in multiple packages
- `KNOSSOS_CHANNEL` env var name — harness detection mechanism
- `.knossos/` directory structure — project detection marker

### Deprecated Markers
| Location | Deprecated | Replacement |
|----------|-----------|-------------|
| `materialize.go:70` | `RiteManifest.Skills` | `Legomena` |
| `materialize.go:69` | `RiteManifest.Commands` | `Dromena` |
| `session/status.go:44-51` | `COMPLETE`/`COMPLETED` aliases | `ARCHIVED` |
| `hook/env.go` | `.current-session` file | Priority chain resolution |

---

## Risk Zone Mapping

**RZ-001** (non-transactional materialize pipeline): `internal/materialize/materialize.go:454-627` — partial failures accumulate; no rollback. "Partial failures are non-fatal."

**RZ-002** (`config.KnossosHome()` test cache): `sync.Once` singleton. Cross-referenced to TENSION-016. `ResetKnossosHome()` mitigation for tests only.

**RZ-003** (search circular layer risk): **PARTIALLY STALE** — original import resolved via `internal/concept` extraction. Latent risk mitigated but hypothetically reintroducible.

**RZ-004** (Mena namespace collision): Mixed dro/lego warnings but no hard failure on collision.

**RZ-005** (time-based lock stale threshold): `internal/lock/lock.go:32` — 5-minute timeout. No heartbeat mechanism.

**RZ-006** (`org_scope.go` silently non-fatal): `org_scope.go:54,69` — "Non-fatal: accumulate error and continue."

**RZ-007** (`userscope/sync.go` fallback channel): Mechanism changed to `paths.UserChannelDir(params.Opts.Channel)` but fallback when channel unconfigured remains a risk.

**RZ-008** (`perspective/context.go` channel dir hardcode): Both `// HA-FS:` markers confirmed — blocks multi-harness.

**RZ-009** (`procession/resolver.go` global config calls): `resolver.go:38-43` calls `config.KnossosHome()` and `config.ActiveOrg()` directly without injection.

**RZ-010** (`isGitWorktree()` 10s timeout): `context.WithTimeout(context.Background(), 10*time.Second)` — no cancellation propagation from caller.

---

## Knowledge Gaps

1. `internal/naxos/triage.go` — debt triage scoring heuristics
2. `internal/sails/` — white sails signaling constraints (`gate.go`, `thresholds.go`, `contract.go`, `proofs.go` uncharted)
3. `internal/session/fsm.go` — state machine transition invariants
4. `internal/procession/` — template schema validation constraints
5. `internal/validation/schemas/` — JSON schema versioning constraints
6. `hooks.yaml` — canonical structure end-to-end format constraints
7. `internal/registry/` — rite registration constraints
8. `internal/materialize/mcp_ownership.go` — MCP ownership model
