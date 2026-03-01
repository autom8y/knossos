# SPIKE: Knossos Ecosystem Artifact Bleed to Satellites

> Comprehensive audit of low-level platform artifacts that could unintentionally leak knossos-internal concerns into satellite projects.

**Date**: 2026-03-01
**Session**: session-20260301-225723-b3144a8b
**Prior Art**: `SPIKE-golden-rules-placement.md`, `SPIKE-sync-safety-audit.md`

---

## 1. Question

When a satellite project (e.g., a web app) runs `ari init --rite hygiene` or `ari sync`, what knossos-internal concepts bleed into its `.claude/` directory? Which artifacts assume the user is working on the knossos platform itself rather than using it?

---

## 2. Findings: Six Leak Vectors

### Vector 1: CLAUDE.md Template Sections (HIGH)

Every satellite receives these sections with knossos-internal references:

| Section | Leaked Concepts |
|---------|----------------|
| `execution-mode` | "Pythia coordinates", three operating modes table |
| `agent-routing` | "Pythia coordinates phases and handoffs", Task tool delegation |
| `commands` | "Dromena", "Legomena", "Knossos Name", full Rosetta Stone mythology |
| `platform-infrastructure` | "Mutate `*_CONTEXT.md` only via `Task(moirai, \"...\")`", `/go`, `/start`, `/park`, `/continue`, `/wrap` |
| `quick-start` | "`prompting` skill. Routing guidance: `/consult`" |
| `know` | "`.know/` persistent knowledge. Generate with `/know --all`" |

**Impact**: A satellite user sees instructions referencing Moirai, Pythia, dromena/legomena, and session commands that may not exist in their project. ~400 tokens of noise per turn.

### Vector 2: Rules Files (MEDIUM-HIGH)

All `knossos/templates/rules/*.md` sync to `.claude/rules/`:

| Rule File | Knossos-Internal Content |
|-----------|--------------------------|
| `mena.md` | `.dro.md` and `.lego.md` as "dromena" and "legomena"; `mena/` directory structure |
| `rites.md` | rite.yaml, orchestrator.yaml, workflow.yaml; multi-rite structure |
| `internal-agent.md` | Agent archetypes, Task tool patterns, orchestrator model |
| `internal-session.md` | SESSION_CONTEXT.md mutations via Task(moirai), session state machine |
| `knossos-templates.md` | Materialization, satellite regions, KNOSSOS_MANIFEST.yaml |
| `internal-materialize.md` | Materialization pipeline, collision detection |
| `internal-inscription.md` | KNOSSOS_MANIFEST.yaml, region ownership |

**Impact**: Rules are path-activated (loaded when editing matching files), so impact is conditional. But a satellite editing `.claude/` files gets rules teaching knossos internals.

### Vector 3: Shared Mena Skills (MEDIUM)

Via `rites/shared/manifest.yaml`, all satellites receive:

| Skill | Problem |
|-------|---------|
| `cross-rite-handoff` | Completely inapplicable (assumes multiple rites) |
| `orchestrator-templates` | Assumes Pythia orchestration model |
| `ephemeral-artifacts` | `.claude/wip/` pattern (knossos-specific) |
| `pinakes` | Domain registry assumes framework domains (dromena, legomena, agents, hooks, mena-structure) |
| `codebase-archaeology` | Teaches "Prompt Fuel" compression, rite creation as endpoint |

Generic/reusable: `smell-detection`, `interview`, `literature-review`, `shared-templates`.

### Vector 4: Agent Prompts (MEDIUM)

Agents synced to satellites reference knossos coordination infrastructure:

- **pythia**: CONSULTATION_REQUEST/RESPONSE protocol, resume parameter, SESSION_CONTEXT.md
- **theoros**: Pinakes, domain criteria, theoria audit model
- Agent `skills:` frontmatter pre-loads knossos-specific knowledge (orchestrator-templates, hygiene-catalog)

### Vector 5: Hook Configuration (MEDIUM)

`ari init` injects into `.claude/settings.json`:
```json
{ "hooks": { "PreToolUse": [{ "command": "ari hook agent-guard" }] } }
```
Assumes `ari` is installed and available in PATH. If satellite user doesn't have ari, every tool use triggers a hook failure.

### Vector 6: Generator Defaults (LOW)

When templates fail, `generator.go` hardcodes fallback content with:
- Mythology terminology (lookupTerminology: dromena, legomena, moirai, etc.)
- Session commands (`/go`, `/start`, `/park`, etc.)
- Agent delegation patterns

Unlikely to trigger (templates are embedded), but the fallback path exposes internal concepts.

---

## 3. Concrete Example: What a Satellite Actually Gets

A satellite web app runs `ari init --rite hygiene`:

1. **CLAUDE.md** tells them to use `/go` to start sessions and `Task(moirai, ...)` to mutate context — neither exists in their project
2. **Agent pythia** expects structured consultation requests — the satellite has no consultation protocol
3. **Rules** (internal-session.md) teach about writeguard hooks and SESSION_CONTEXT.md — the satellite has no session files
4. **Skills** (pinakes) reference domain audits for dromena/legomena — the satellite has no mena directories
5. **Hooks** call `ari hook agent-guard` — if ari isn't installed, every edit triggers an error

---

## 4. Root Cause

The materialization pipeline treats the knossos repo and satellite projects identically. There is no concept of **"am I materializing for knossos itself, or for a consumer project?"** Templates, rules, skills, and agents all assume the user is a knossos developer.

The shared rite (`rites/shared/`) was designed as a commons for cross-rite reuse within knossos, not as an SDK for external consumers. When it syncs to satellites, it brings internal infrastructure assumptions with it.

---

## 5. Severity Assessment

| Vector | Severity | Frequency | Token Waste/Turn |
|--------|----------|-----------|------------------|
| CLAUDE.md sections | HIGH | Every turn | ~400 tokens |
| Rules files | MEDIUM-HIGH | When editing .claude/ | ~200 tokens (conditional) |
| Shared mena skills | MEDIUM | On-demand | ~0 (loaded via Skill tool) |
| Agent prompts | MEDIUM | Agent invocation | ~0 (loaded per-agent) |
| Hook configuration | MEDIUM | Every tool use | Hook failure noise |
| Generator defaults | LOW | Template failure only | ~0 |

**Total unconditional waste**: ~400 tokens/turn from CLAUDE.md sections alone.

---

## 6. Remediation Options (Not Scoped for Implementation)

### Option A: Satellite-Aware Templates
Add a `satellite: true/false` flag to rite manifests. Templates conditionally render knossos-internal content only when `satellite: false`. Satellites get clean, framework-agnostic sections.

### Option B: Tiered Shared Rite
Split `rites/shared/` into `shared-core/` (generic, satellite-safe) and `shared-platform/` (knossos-internal). Manifests reference one or both.

### Option C: Content Filtering at Sync
`ari sync` strips knossos-internal references when materializing to a non-knossos project (detected by presence of `KNOSSOS_MANIFEST.yaml` at root or a flag in rite config).

### Option D: Template Variants
Create `*.satellite.md.tpl` variants alongside existing templates. The materializer selects the satellite variant when the project doesn't have a `KNOSSOS_MANIFEST.yaml`.

---

## 7. Immediate Low-Risk Improvements

These require no architectural changes:

1. **Rules scoping**: Move `internal-*.md` rules behind a path guard that only activates in the knossos repo (check for `KNOSSOS_MANIFEST.yaml` or `knossos/` directory)
2. **Hook safety**: `ari hook agent-guard` should fail silently if ari is not installed (exit 0 with no output instead of error)
3. **Template conditionals**: Add `{{ if .IsKnossosRepo }}` guards around Moirai/session/mythology references in existing templates
4. **Skill pruning**: Remove `cross-rite-handoff` and `orchestrator-templates` from shared manifest (only useful within knossos)

---

## 8. Relationship to Golden Rules Spike

The golden rules ("never edit .claude/ directly", "never add session artifacts to mena") are knossos-repo-specific — they make no sense in satellites. The golden rules spike correctly identified these as project-local. This audit confirms the broader pattern: **most of what materializes assumes knossos, not satellites.**

The golden rules should go in the `user-content` satellite region of the knossos project CLAUDE.md (as recommended). They should NOT be baked into templates, which would leak them to every satellite.
