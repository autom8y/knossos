---
domain: design-constraints
generated_at: "2026-03-27T19:57:42Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "5501b0aa"
confidence: 0.87
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

20 TENSION-NNN entries cataloged. 16 active, 1 resolved, 3 new.

**TENSION-001** (`internal/provenance/provenance.go:91`, `internal/inscription/types.go:14`): Dual `OwnerType` types in separate packages -- identical names, distinct types, cross-referenced by comments only.

**TENSION-002** (`internal/materialize/materialize.go:118,284-289,390-391`): `channelDirOverride` mutation-with-defer pattern for multi-channel materialization. Save-and-restore at both rite scope and user scope.

**TENSION-003** (`internal/perspective/context.go:60`, `internal/perspective/resolvers.go:274`): Two `// HA-FS:` markers hardcoding `.claude/` paths. Blocks full harness agnosticism.

**TENSION-004** (`internal/agent/frontmatter.go:38-43`): `maxTurns`, `disallowedTools`, `permissionMode`, `mcpServers` as direct struct fields alongside `Contract *BehavioralContract` which also contains `MaxTurns`.

**TENSION-005** (`internal/resolution/chain.go:5-6`): Zero-import invariant -- "This package has ZERO internal imports. All tier paths are injected via constructor to avoid import cycles (TENSION-005)."

**TENSION-006** (`internal/provenance/provenance.go:69-70`): String values must stay in sync with `internal/materialize/source/types.go` -- manual synchronization without shared constants.

**TENSION-007** (`internal/hook/env.go`): `GetAdapter()` defaults to ClaudeAdapter when `KNOSSOS_CHANNEL` is unset -- implicit default.

**TENSION-008** (`internal/materialize/sync_types.go:51`): Sync state types coupled to materialization pipeline.

**TENSION-009** (`internal/materialize/userscope/sync.go`): Partially stale -- mechanism changed to `paths.UserChannelDir(params.Opts.Channel)` dynamic resolution, but fallback when channel is unconfigured remains a risk.

**TENSION-010** (`internal/agent/frontmatter.go:55-70`): Agent archetype fields duplicated across frontmatter and behavioral contract.

**TENSION-011** (`internal/materialize/materialize.go:69-70`): `Skills []string` (deprecated) and `Commands []string` (backward compat) both present in `RiteManifest`.

**TENSION-012** (`internal/hook/output.go:28`): `HookEventName: "PreToolUse"` hardcoded wire value -- CC reads this exact string for security decisions.

**TENSION-013** (`internal/inscription/pipeline.go:155-156`): Harness-agnosticism gap in inscription pipeline -- `InscriptionPath` targets CC channel context file.

**TENSION-014** (`internal/tribute/types.go:58,158`, `internal/tribute/renderer.go:64,250`): "Phase 2 - placeholder" comments -- tribute feature incomplete.

**TENSION-015** (`internal/search/collectors.go:12`): **RESOLVED** -- `internal/search` no longer imports `internal/cmd/explain`. Resolved via `internal/concept` extraction.

**TENSION-016** (`internal/config/home.go:11-23`): `sync.Once` singleton for KnossosHome -- `ResetKnossosHome()` exists with test-only warning.

**TENSION-017** (`internal/materialize/materialize.go:579-584`): Split output targets -- settings to channel dir, MCP to project root.

**TENSION-018** (NEW -- `internal/search/bm25/build.go:32,191`): Partial consolidation of `repoFromQualifiedName` duplication. ADR-hierarchical-qualified-names mandated consolidation but `bm25/build.go` still has 2 private inline `SplitN` calls using `parts[1]` instead of `know.RepoFromQualifiedName`. Scope-unsafe for hierarchical qualified names.

**TENSION-019** (NEW -- `internal/reason/context/assembler.go`): CE diversity floor configuration defaults to empty. If `Pipeline` wires empty config, floor enforcement silently does nothing. Silent configuration degradation risk.

**TENSION-020** (NEW -- `internal/slack/handler.go:554-556`): Emoji ACK is fire-and-forget with no observability. Silent fail-open in user-facing signal.

---

## Trade-off Documentation

**Trade-off 1 (WriteIfChanged):** `writeIfChanged()` compares content before writing. Persists because CC file watcher crashes are worse than the overhead.

**Trade-off 2 (Dual read path, v1/v2/v3):** `internal/session/events_read.go:13-27` -- ADR-0027 exception: write path unified, read path bridges pre- and post-ADR-0027 formats. Removal trigger: "once all pre-ADR-0027 sessions are archived."

**Trade-off 3 (Provenance in `.knossos/`):** Provenance manifests stored in `.knossos/` not `.claude/` -- CC context window exposure risk.

**Trade-off 4 (CC event names as lingua franca):** ADR-0032 -- rejected a third vocabulary. CC wire names are canonical.

**Trade-off 5 (XDG logic duplicated):** `config.ActiveOrg()` duplicates path resolution -- import cycle constraint prevents sharing with paths package.

**Trade-off 6 (Two-manifest architecture):** Rite-scope and user-scope provenance manifests are separate to avoid cross-contamination.

**Trade-off 7 (channelDirOverride save-and-restore):** ADR-0031 acknowledges as "pragmatic hack." Constructor refactor would affect 15+ call sites.

**Trade-off 8 (ClaudeCompiler as pass-through):** Symmetric pipeline rejected -- Gemini needs active compilation, Claude is mostly pass-through.

**Trade-off 9** (NEW -- CE parameters as configurable, not code invariants): ADR-contextual-equilibrium-parameters documents 5 CE mechanisms with calibrated defaults. Configured via `AssemblerConfig` fields. Trade-off: allows tuning without code changes, but defaults can be silently bypassed.

**Trade-off 10** (NEW -- Hierarchical qualified names use GitHub `/` restriction): ADR-hierarchical-qualified-names uses GitHub's repo-name-no-slash constraint as parsing invariant. Near-zero likelihood of breaking, critical impact if it does.

**Trade-off 11** (NEW -- Mutex-free summary generation in knowledge builder): `internal/search/knowledge/builder.go:82` -- `GenerateSummary()` runs without mutex, `Set()` acquires lock. Explicit fix for cascade timeout (commit `58a5def9`). Two-phase pattern must be followed by all callers.

---

## Abstraction Gap Mapping

**AGM-001** (`internal/perspective/`): Hardcoded `.claude/` paths with `// HA-FS:` markers. Blocks Gemini feature parity.

**AGM-002** (`internal/agent/frontmatter.go`): `MaxTurns int` at direct field level coexists with `Contract *BehavioralContract` containing `MaxTurns`. Duplicate semantics.

**AGM-003** (`AGENTS.md` compilation): No AGENTS.md (OpenAI/ChatGPT format) compilation exists in materialize pipeline.

**AGM-004** (`BehavioralContract.MaxTurns` not wired): Contract MaxTurns not plumbed through to agent output.

**AGM-005** (`RiteManifest.Commands` and `.Skills`): Both deprecated fields still present at `internal/materialize/materialize.go:69-70`.

**AGM-006** (Duplicate resolution logic): Both `materialize/source` and `resolution` packages implement tier traversal. Moderate duplication.

**AGM-007** (`tribute/Commits` placeholder): Type surface with "Phase 2 - placeholder" -- feature incomplete.

**AGM-008** (`fileutil.AtomicWriteFile` vs. `os.WriteFile`): `userscope/sync.go` uses `os.WriteFile` while most pipeline uses `AtomicWriteFile`. Intentional but asymmetric.

**AGM-009** (NEW -- `repoFromQualifiedName` stale copies in `bm25/build.go`): ADR mandated consolidation but `internal/search/bm25/build.go:36,194` still has inline `parts[1]` extraction. For scoped qualified names, returns `"repo/scope"` instead of `"repo"` -- silent correctness bug.

**AGM-010** (NEW -- CE configuration coupling): `AssemblerConfig` fields have no defaults enforced at construction time. Caller can pass empty config and silently disable all CE diversity mechanisms.

---

## Load-Bearing Code Identification

**LB-001** (`internal/fileutil/fileutil.go:66-72`): `WriteIfChanged()` -- naive replacement with `os.WriteFile` causes CC file watcher to crash.

**LB-002** (`internal/provenance`): `structurallyEqual()` -- prevents unnecessary writes that would trigger CC file watcher.

**LB-003** (`internal/materialize/mena/engine.go`): Three-pass rewrite ordering. Wrong order causes double-rewrite corruption.

**LB-004** (`internal/session/events_read.go:82-117`): Dual-format event reader. Wrong format order causes silent event misread.

**LB-005** (`internal/hook/clewcontract/type_rename.go:14-22`): Append-only event type rename map. Removing entries breaks backward compatibility.

**LB-006** (`internal/config/home.go:11-23`): `sync.Once` singleton for KnossosHome. `ResetKnossosHome()` only for tests.

**LB-007** (`internal/materialize/materialize_agents.go:42-45`): No pre-delete before agent overwrite -- CC watcher fires DELETE events that cause agent disappearance.

**LB-008** (`internal/inscription`): Satellite region preservation in `Merger` -- core invariant protecting user content.

**LB-009** (`internal/provenance`): `LoadOrBootstrap()` aborts on corrupt manifests -- not fail-open.

**LB-010** (`internal/hook/output.go:28`): `HookEventName: "PreToolUse"` hardcoded wire format string -- security boundary.

**LB-011** (`internal/materialize/materialize.go:459-632`): Pipeline step ordering -- non-transactional, partial failures possible.

**LB-012** (NEW -- `internal/search/knowledge/builder.go:80-82`): Two-phase summary generation (`GenerateSummary()` outside lock, `Set()` under lock). Callers using `Generate()` instead reintroduce serialization bug.

**LB-013** (NEW -- `internal/triage/orchestrator.go:792-795`): Graph injection constants (`maxGraphInjectPerCandidate=2`, `maxGraphInjectTotal=4`, `baselineScore=0.15`). Hardcoded, not configurable. Changing without recalibration risks flooding or suppressing graph candidates.

---

## Evolution Constraint Documentation

### SAFE (local change only)
- `internal/resolution/chain.go` -- zero-import policy, self-contained
- `internal/tribute/types.go` -- placeholder types, no external callers
- `internal/concept/concept.go` -- isolated concept registry
- `internal/tokenizer` -- stdlib wrapper
- `internal/fileutil` -- utility functions (care with `WriteIfChanged` -- LB-001)
- `internal/checksum` -- hash utilities

### COORDINATED (multi-file, no external break)
- Adding a new `TargetChannel`: minimum 5 files across 4 packages
- New hook handler: handler file + hooks.yaml + test + CLAUDE.md regeneration
- New agent frontmatter field: `agent/frontmatter.go` + transform + test + lint rule
- New session state: `status.go` FSM + `NormalizeStatus()` + migrate
- New mena type: `mena/source.go` routing + materialize pipeline + lint
- Adding CE diversity floor type: update config + domain vocabulary map
- Adding scoped qualified name consumer: use `know.RepoFromQualifiedName()` not `SplitN()[1]`

### MIGRATION (breaking change to callers)
- Provenance schema version bump (`CurrentSchemaVersion = "2.0"`) -- requires migration function
- Session context schema change -- requires `internal/cmd/session/migrate.go` entry
- Inscription schema version -- `DefaultSchemaVersion = "1.0"`
- Deprecated field removal (`RiteManifest.Skills`, `.Commands`) -- multi-rite impact

### FROZEN (do not touch without explicit decision)
- `internal/hook/output.go` `HookEventName: "PreToolUse"` -- CC security boundary
- `.claude/` directory name -- CC hardcoded expectation
- `PROVENANCE_MANIFEST.yaml` filename -- referenced in multiple packages
- `KNOSSOS_CHANNEL` env var name -- harness detection mechanism
- `.knossos/` directory structure -- project detection marker
- `"org::repo::domain"` qualified name format -- existing serialized names must round-trip

### Deprecated Markers
| Location | Deprecated | Replacement |
|----------|-----------|-------------|
| `internal/materialize/materialize.go:70` | `RiteManifest.Skills` | `Legomena` |
| `internal/materialize/materialize.go:69` | `RiteManifest.Commands` | `Dromena` |
| `internal/session/events_read.go` | Legacy `Event` struct | `clewcontract.Event` |
| `internal/hook/env.go` | `.current-session` file | Priority chain resolution |
| `internal/search/bm25/build.go:36,194` | `parts[1]` inline repo extraction | `know.RepoFromQualifiedName()` |

---

## Risk Zone Mapping

**RZ-001** (non-transactional materialize pipeline): `internal/materialize/materialize.go:459-632` -- partial failures accumulate; no rollback.

**RZ-002** (`config.KnossosHome()` test cache): `sync.Once` singleton. Cross-referenced to TENSION-016.

**RZ-003** (search circular layer risk): **PARTIALLY STALE** -- resolved via `internal/concept` extraction but reintroducible.

**RZ-004** (Mena namespace collision): Mixed dro/lego warnings but no hard failure on collision.

**RZ-005** (time-based lock stale threshold): `internal/lock/lock.go:32` -- 5-minute timeout. No heartbeat mechanism.

**RZ-006** (`org_scope.go` silently non-fatal): Org-scope failures are swallowed.

**RZ-007** (`userscope/sync.go` fallback channel): Fallback when channel unconfigured remains a risk.

**RZ-008** (`perspective/context.go` channel dir hardcode): Both `// HA-FS:` markers -- blocks multi-harness.

**RZ-009** (`procession/resolver.go` global config calls): Calls `config.KnossosHome()` and `config.ActiveOrg()` directly without injection.

**RZ-010** (`isGitWorktree()` 10s timeout): No cancellation propagation from caller.

**RZ-011** (NEW -- Scope-unsafe BM25 repo extraction): `internal/search/bm25/build.go:36,194` -- `parts[1]` returns `"repo/scope"` for scoped qualified names. Silent correctness bug activating with hierarchical knowledge.

**RZ-012** (NEW -- CE diversity floor silent bypass): `internal/reason/context/assembler.go:37-50` -- `DiversityFloorTypes` defaults to empty. No warning when floor types empty.

**RZ-013** (NEW -- Emoji ACK failure invisible): `internal/slack/handler.go:554-556` -- fire-and-forget with no metric or log on failure.

---

## Knowledge Gaps

1. `internal/naxos/triage.go` -- debt triage scoring heuristics
2. `internal/sails/` -- white sails signaling constraints uncharted
3. `internal/session/fsm.go` -- state machine transition invariants
4. `internal/procession/` -- template schema validation constraints
5. `internal/validation/schemas/` -- JSON schema versioning constraints
6. `internal/materialize/hooks/` -- hooks.yaml canonical format constraints
7. `internal/registry/` -- rite registration constraints
8. `internal/materialize/mcp_ownership.go` -- MCP ownership model
9. `internal/cmd/serve/serve.go:1005-1009` -- inline `SplitN` for qualified name parsing not verified for scope safety
10. New ADRs (org-topology-design, cassandra-dedup-boundary) not read
