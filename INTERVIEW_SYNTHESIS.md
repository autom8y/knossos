# Knossos/Ari — Interview Synthesis

Context download from sole developer. Conducted 2026-02-06 across 8 rounds. This document is a reference for all future Claude sessions on this project.

---

## The Golden Rule

**Never edit `.claude/` directly.** Always edit source files (`rites/`, `mena/`, `knossos/templates/`, `user-agents/`, `user-hooks/`) and rematerialize. The `.claude/` directory is output, not input. This is the #1 mistake Claude makes on this project, and the #1 instruction for every session.

The one exception: satellite regions in `.claude/CLAUDE.md` are user-owned and preserved across materializations. But even those should be authored with awareness that they coexist with platform-generated regions.

---

## What Knossos Is

Knossos is a **framework bet** — the "Rails for Claude Code." It's a context-engineering meta-framework that transforms Claude Code from a generic AI coding assistant into a structured, multi-agent development environment.

The core insight: Claude Code becomes exponentially more powerful when its context window is precisely engineered. Knossos is the system that does that engineering.

**It is NOT:**
- A personal power tool (though the developer uses it daily)
- An abstraction over multiple AI coding tools
- A proof of concept

**It IS:**
- Tightly coupled to Claude Code's `.claude/` convention, with no abstraction layer
- A freemium framework with a firm business model
- Built by a real company (autom8y) with stakeholders who will dogfood and extend it

---

## The Self-Referential Nature

The hardest thing to internalize: **Knossos is a framework FOR Claude that is BUILT BY Claude.** The `.claude/` directory it generates is also the `.claude/` directory Claude uses to build Knossos itself. This creates a 4-hop indirection chain:

```
Source files → Materialization → .claude/ projection → Claude's context window
```

When working on Knossos, you are modifying the system that controls how you operate. Every change to a rite, agent, or template potentially changes Claude's own behavior in the next session. Treat this with appropriate care.

---

## Architecture: Source → Projection Model

```
SOURCE (authored)                    PROJECTION (consumed by Claude Code)
─────────────────                    ─────────────────────────────────────
rites/*/agents/*.md        ──┐
rites/*/mena/*.dro.md      ──┤      .claude/
rites/*/mena/*.lego.md     ──┤      ├── agents/          (rite agents)
rites/*/manifest.yaml      ──┤      ├── commands/         (dromena)
user-agents/*.md           ──┼──▶   ├── skills/           (legomena)
user-hooks/                ──┤      ├── hooks/            (hook scripts + ari)
mena/**/*.{dro,lego}.md    ──┤      ├── CLAUDE.md         (inscription)
knossos/templates/         ──┤      ├── settings.local.json (MCP merge)
                           ──┘      ├── KNOSSOS_MANIFEST.yaml
                                    └── ACTIVE_RITE
```

**Materialization invariants (sacred, in priority order):**
1. **User content is NEVER destroyed** — satellite regions, user-agents, user-hooks survive materialization
2. **Materialization is idempotent** — running it twice produces identical output
3. `.claude/` is NOT a pure cache — it contains user state (satellite regions). "Delete and regenerate" is an anti-pattern because it would violate invariant #1.

**4-tier source resolution** (highest priority first): explicit path → project `rites/` → user `~/.local/share/knossos/rites/` → knossos platform rites

---

## The Mena Model

Mena ("sacred commands") has two fundamentally different types, distinguished by **context lifecycle**:

| Aspect | Dromena (`.dro.md`) | Legomena (`.lego.md`) |
|--------|--------------------|-----------------------|
| Greek meaning | "things done" — enacted rituals | "things said" — spoken words |
| Projects to | `.claude/commands/` | `.claude/skills/` |
| Invocation | User-invoked (`/start`, `/commit`) | Agent-loaded (Skill tool, on demand) |
| Context lifecycle | **Transient** — execute and exit | **Persistent** — stays in context as reference |
| Nature | Imperative actions | Declarative knowledge |

This is NOT just a routing difference. The distinction is architectural: dromena change state, legomena inform decisions. The file extension (`.dro.md` vs `.lego.md`) declares the type. Nested files inherit from their INDEX file.

---

## The Mythology

The Greek naming is **non-negotiable core identity** — not decoration, not branding. The terms are load-bearing architecture:

| Term | Greek | Architectural meaning |
|------|-------|-----------------------|
| **Knossos** | The labyrinth palace | The framework/platform — complex, structured, navigable with the right thread |
| **Ariadne (ari)** | Princess who gave the thread | The CLI — your guide through the labyrinth |
| **Mena** | Sacred commands | Source content that becomes Claude's capabilities |
| **Dromena** | Things enacted | Transient commands (actions) |
| **Legomena** | Things spoken | Persistent knowledge (reference) |
| **Clew** | Ball of thread | Append-only event log — memory through the maze of decisions |
| **Naxos** | Island where Ariadne was abandoned | Orphan session scanner — finding abandoned things |
| **Moirai** | The Fates | Session lifecycle agent — spins, measures, and cuts sessions |
| **White Sails** | Theseus's signal to Athens | Confidence/quality signal (WHITE/GRAY/BLACK) |

Understanding WHY something is named tells you HOW it should behave. If you don't know the myth, you'll misuse the component.

---

## Business Model

**Freemium, with a firm boundary:**

| Tier | Includes |
|------|----------|
| **Free** | Rites, materialization engine, mena (dromena + legomena), agent factory, /consult |
| **Enterprise** | Sessions, Clew events, White Sails proof system, fray (session forking) |

The gating mechanism is **intentionally deferred**. It's not just a CLI feature flag problem — the `.md` source files that flow into Claude's context would also need gating, which requires deep review of the context-engineering layer. For now, everything ships open for internal dogfooding.

---

## Working With Claude on This Project

### Autonomy Model
**High autonomy within the active rite.** When a session is active with a rite, follow the orchestrator pattern. Delegate to specialist agents via Task tool. Don't ask for every decision.

### Session Scope
The **rite + workflow/task type** together determine session shape. 10x-dev initiative sessions are long and deep. Hygiene tasks are quick. Security audits are structured. The workflow (from dromena) defines the session's character as much as the rite does.

### Ambiguity Protocol
1. **Check the ADRs first** — 22 ADRs in `docs/decisions/`. They are the source of truth for architectural decisions.
2. If no ADR covers it, follow existing patterns in the codebase.
3. If still ambiguous, ask.

### The Three Failure Modes (avoid these)
1. **Editing generated files** — modifying `.claude/` instead of source and rematerializing
2. **Ignoring the orchestrator** — doing everything directly instead of delegating to specialist agents
3. **Misunderstanding the mythology** — using wrong terminology, creating files in wrong places, or breaking the conceptual model

### Testing Philosophy
**Pragmatic coverage.** Write tests when you write code. Don't backfill tests for working code. Don't skip tests for new code. Let coverage emerge. 92 test files across the codebase reflects this — thorough where it matters, absent where code is stable.

---

## Current State and Priorities

### Priority Order (next 2-4 weeks)
1. **Shell script deep cleanse** — Claude audits the call graph across 124 scripts (39K LOC), produces triage report, developer reviews. Many are dead/redundant.
2. **Hook architecture overhaul** — eliminate bash wrappers. End state: `ari` IS the hook binary, `.claude/settings.json` points directly at it.
3. **Single-binary completion** — `ari init`, rite embedding (2MB), remaining shell→Go ports (~900 LOC)
4. **Dogfooding** — autom8y team runs real projects through Knossos

### Extension Path State
- **user-agents/**: Solid, usersync with collision detection works
- **user-hooks/**: Functional but not optimally architected (bash wrapper layers)
- **Custom mena**: Has routing gaps in materialization

### Open Architectural Questions
- **ari init bootstrapping without a repo** — the biggest unsolved design problem. Everything currently assumes a knossos repo checkout with filesystem sources. Embedded rites fundamentally change the source model.
- **Hook event completeness** — Claude Code's event surface keeps growing. Principled subset vs. full coverage.
- **Rite composition** — rites have dependencies but no formal inheritance/extension model

### V1 Definition
`brew install ari && ari init` works for a developer with Claude Code, AND the autom8y team is using it daily on real projects.

---

## Quick Reference

| Question | Answer |
|----------|--------|
| Build | `CGO_ENABLED=0 go build ./cmd/ari` |
| Test | `CGO_ENABLED=0 go test ./...` |
| Key entry point | `cmd/ari/main.go` → `internal/cmd/root.go` |
| Materialize | `ari sync materialize --rite <name>` |
| Rite swap | `ari rite swap <name>` |
| Active rite | `.claude/ACTIVE_RITE` |
| ADRs | `docs/decisions/` (22 documents) |
| Templates | `knossos/templates/sections/*.md.tpl` |
| Agent schema | `internal/validation/schemas/agent.schema.json` |
| Go version | 1.22+ |
| Dependencies | 6 direct (Cobra, Viper, yaml.v3, jsonschema, sprig, xdg) |
| Distribution | GoReleaser → GitHub releases → `autom8y/tap` Homebrew |

---

*Produced from 8-round structured interview, 2026-02-06. Update when architectural decisions change.*
