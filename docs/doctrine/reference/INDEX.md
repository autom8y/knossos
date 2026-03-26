---
last_verified: 2026-03-26
---

# Doctrine Reference Index

> Navigation hub for the Knossos philosophical and architectural foundation.

This index organizes doctrine content by audience and purpose. Use this as your entry point when exploring the philosophy, architecture, and evolution of the Knossos platform.

---

## Quick Navigation

| I want to... | Go to |
|--------------|-------|
| Understand the core philosophy | [Philosophy](#philosophy) |
| Learn why Knossos is designed this way | [Design Principles](#design-principles) |
| Map mythology to implementation | [Mythology Concordance](#mythology-concordance) |
| Use the Ariadne CLI | [CLI Reference](#cli-reference) |
| Choose a rite for my work | [Rite Catalog](#rite-catalog) |
| Run parallel AI coding sessions | [Worktree Guide](#worktree-guide) |
| See architectural decisions | [Architecture](#architecture) |
| Check compliance status | [Compliance](#compliance) |

---

## Philosophy

### The Knossos Doctrine

**Path:** [`../philosophy/knossos-doctrine.md`](../philosophy/knossos-doctrine.md)

The complete philosophical foundation—The Coda that gives the platform its meaning.

**Contents:**
- I. Cosmogony (why Knossos exists)
- II. The Naming of Things (mythological mapping)
- III. Mortal Limits (design constraints)
- IV. The Rites (practice bundles)
- V. The Journey (session lifecycle)
- VI. The Clew Contract (event recording)
- VII. The Confidence Signal (White Sails)
- VIII. The Handoff (context transfer)
- IX. The Hooks (trap mechanisms)
- X. The Complete Service Map
- XI. Design Principles (the eight revelations)
- XII. Terminology Concordance
- XIII. The Coda (philosophical summary)
**Audience:** All contributors, architects, anyone seeking to understand the "why" behind Knossos.

---

### Design Principles

**Path:** [`../philosophy/design-principles.md`](../philosophy/design-principles.md)

The eight foundational principles extracted from Section XI of the Knossos Doctrine, with implementation details and application guidance.

**Principles:**
1. The Clew Is Sacred
2. Honest Signals Over Comfortable Lies
3. Mutation Through the Fates
4. Rites Over Teams
5. Heroes Are Mortal
6. The Labyrinth Grows
7. Return Is the Victory
8. The Inscription Prepares

**Each principle includes:**
- Statement (doctrinal truth)
- Implementation (Go packages and source locations)
- Application guidance (how to apply in practice)

**Audience:** Architects, engineers implementing platform features, contributors needing design guidance.

---

### Mythology Concordance

**Path:** [`../philosophy/mythology-concordance.md`](../philosophy/mythology-concordance.md)

Authoritative mapping from Greek mythology to SOURCE implementation (not projections).

**Covers:**
- Core components (Knossos, Ariadne, The Clew, Theseus, Heroes)
- The Moirai (Clotho, Lachesis, Atropos)
- Supporting elements (Inscription, White Sails, Rites, Daedalus, Potnia)
- Destinations and states (Athens, Naxos, Dionysus, Minos, Minotaur)
- SOURCE vs PROJECTION distinction (critical!)
- Materialization flow

**For each mythological element:**
- Mythological origin (brief context)
- Knossos implementation (source locations)
- Key files (paths relative to repo root)
- Design rationale (why this mapping)

**Audience:** New contributors learning the mythology, engineers needing to locate source code, anyone confused by SOURCE vs PROJECTION.

---

## Architecture

### Decisions (ADRs)

**Path:** [`../../decisions/`](../../decisions/)

Architecture Decision Records documenting significant platform decisions.

**Key ADRs:**
- [`ADR-0009-knossos-roster-identity.md`](../../decisions/ADR-0009-knossos-roster-identity.md) — SOURCE vs PROJECTION clarification
- [`ADR-0005-moirai-centralized-state-authority.md`](../../decisions/ADR-0005-moirai-centralized-state-authority.md) — Moirai as mutation authority
- [`ADR-0001-session-state-machine-redesign.md`](../../decisions/ADR-0001-session-state-machine-redesign.md) — Session lifecycle states
- [`ADR-0031-multi-channel-architecture.md`](../../decisions/ADR-0031-multi-channel-architecture.md) — Multi-channel projection architecture
- [`ADR-0032-harness-agnostic-event-vocabulary.md`](../../decisions/ADR-0032-harness-agnostic-event-vocabulary.md) — Harness-agnostic event vocabulary

**Audience:** Architects, engineers implementing features, anyone needing decision context.

---

### CLI Reference

**Path:** [`../operations/cli-reference/`](../operations/cli-reference/)

Complete reference for all Ariadne CLI commands (32 command families; see `ari --help` for current list).

**Command Families (all 32 documented):**
- [session](../operations/cli-reference/cli-session.md) — Session lifecycle (15 commands)
- [rite](../operations/cli-reference/cli-rite.md) — Rite management (10 commands)
- [worktree](../operations/cli-reference/cli-worktree.md) — Parallel sessions (11 commands)
- [sync](../operations/cli-reference/cli-sync.md) — Materialization (8 commands)
- [hook](../operations/cli-reference/cli-hook.md) — Hook operations (11 commands)
- [handoff](../operations/cli-reference/cli-handoff.md) — Agent handoffs (4 commands)
- [agent](../operations/cli-reference/cli-agent.md) — Agent management (summon, dismiss, roster)
- [serve](../operations/cli-reference/cli-serve.md) — Clew HTTP server
- [procession](../operations/cli-reference/cli-procession.md) — Cross-rite workflows
- [land](../operations/cli-reference/cli-land.md) — Session synthesis (Dionysus)
- [status](../operations/cli-reference/cli-status.md) — Platform health dashboard
- [org](../operations/cli-reference/cli-org.md) — Organization management
- [registry](../operations/cli-reference/cli-registry.md) — Registry sync and status
- [ask](../operations/cli-reference/cli-ask.md) — Domain-filtered queries
- [knows](../operations/cli-reference/cli-knows.md) — Knowledge base checks
- [ledge](../operations/cli-reference/cli-ledge.md) — Work product artifacts
- [lint](../operations/cli-reference/cli-lint.md) — Codebase linting
- [provenance](../operations/cli-reference/cli-provenance.md) — Content provenance
- [explain](../operations/cli-reference/cli-explain.md) — Term definitions
- [complaint](../operations/cli-reference/cli-complaint.md) — Complaint tracking
- [init](../operations/cli-reference/cli-init.md) — Project initialization
- [tour](../operations/cli-reference/cli-tour.md) — Interactive tour
- [version](../operations/cli-reference/cli-version.md) — Version info
- [help](../operations/cli-reference/cli-help.md) — CLI help

**Entry Point:** [CLI Reference Index](../operations/cli-reference/index.md)

**Audience:** Operators, developers, anyone using the `ari` CLI.

---

### Rite Catalog

**Path:** [`../rites/`](../rites/)

Documentation for each canonical rite (practice bundle).

**Available Rites (19; all documented):**
- [10x-dev](../rites/10x-dev.md) — Full development lifecycle
- [arch](../rites/arch.md) — Architecture assessment
- [clinic](../rites/clinic.md) — Clinical debugging workflow
- [docs](../rites/docs.md) — Documentation workflow
- [forge](../rites/forge.md) — Agent and tool creation (meta-rite)
- [hygiene](../rites/hygiene.md) — Code quality maintenance
- [debt-triage](../rites/debt-triage.md) — Technical debt remediation
- [releaser](../rites/releaser.md) — Release management
- [review](../rites/review.md) — Code review workflow
- [security](../rites/security.md) — Threat modeling and compliance
- [sre](../rites/sre.md) — Operations and reliability
- [intelligence](../rites/intelligence.md) — Research and synthesis
- [rnd](../rites/rnd.md) — Exploration and prototypes
- [strategy](../rites/strategy.md) — Business analysis
- [thermia](../rites/thermia.md) — Thermal/performance analysis
- [ecosystem](../rites/ecosystem.md) — Platform infrastructure
- [slop-chop](../rites/slop-chop.md) — AI code quality gate
- [ui](../rites/ui.md) — UI development workflow
- [shared](../rites/shared.md) — Cross-rite resources

**Entry Point:** [Rite Catalog Index](../rites/index.md)

**Audience:** Practitioners selecting or invoking rites, rite authors, workflow designers.

---

### Worktree Guide

**Path:** [`../guides/worktree-guide.md`](../guides/worktree-guide.md)

Comprehensive guide to running parallel AI coding sessions with filesystem isolation.

**Topics:**
- Worktree creation and lifecycle
- Merge, diff, and cherry-pick operations
- Production patterns (parallel features, hotfixes, CI/CD)
- Troubleshooting common issues
- Architecture and best practices

**Audience:** Developers running parallel sessions, CI/CD engineers, anyone needing worktree patterns.

---

### Agent Capabilities (CC-OPP)

**Path:** [`agent-capabilities.md`](agent-capabilities.md)

Reference for the CC Operational Platform Properties uplift — memory, skills, hooks, and resume capabilities for agents.

**Audience:** Agent authors, platform engineers, anyone configuring agent frontmatter.

---

### Architecture Map

**Path:** [`architecture-map.md`](architecture-map.md)

Subsystem table mapping Go packages to entry points and purpose. CLI-to-package mapping and key flow descriptions.

**Audience:** Engineers navigating the codebase, contributors seeking implementation entry points.

---

## Compliance

### Compliance Status

**Path:** [`../compliance/COMPLIANCE-STATUS.md`](../compliance/COMPLIANCE-STATUS.md)

Current compliance status against doctrine principles and architectural decisions.

**Tracks:**
- Implementation gaps
- Drift from doctrine
- Remediation priorities
- Milestone progress

**Audience:** Architects tracking alignment, project managers, contributors checking platform health.

---

## Operations

### Foundations

**Path:** [`../foundations/`](../foundations/)

ADR symlinks providing doctrine-accessible access to architecture decisions. Contains symlinks to `../../decisions/` for ADR-0001, ADR-0005, and ADR-0009.

**Audience:** Engineers reading foundational architecture decisions in context.

---

## Summary Document

### DOCTRINE.md

**Path:** [`../DOCTRINE.md`](../DOCTRINE.md)

Executive summary of the doctrine—concise overview for stakeholders and new contributors.

**Contents:**
- Platform vision (one-paragraph summary)
- Core principles (brief)
- Mythology overview (elevator pitch)
- Pointer to full doctrine

**Audience:** Stakeholders, new contributors, anyone needing a quick introduction before diving into full doctrine.

---

## Reading Paths

### New Contributor

1. Start: [`../DOCTRINE.md`](../DOCTRINE.md) (executive summary)
2. Context: [`../philosophy/knossos-doctrine.md`](../philosophy/knossos-doctrine.md) (full philosophy)
3. Practical: [`../philosophy/design-principles.md`](../philosophy/design-principles.md) (how to apply)
4. Reference: [`../philosophy/mythology-concordance.md`](../philosophy/mythology-concordance.md) (find source code)

### Architect / Designer

1. Foundation: [`../philosophy/knossos-doctrine.md`](../philosophy/knossos-doctrine.md) (complete doctrine)
2. Principles: [`../philosophy/design-principles.md`](../philosophy/design-principles.md) (design DNA)
3. Decisions: [`../../decisions/ADR-*.md`](../../decisions/) (historical context)
4. Compliance: [`../compliance/COMPLIANCE-STATUS.md`](../compliance/COMPLIANCE-STATUS.md) (current state)

### Engineer / Implementer

1. Mapping: [`../philosophy/mythology-concordance.md`](../philosophy/mythology-concordance.md) (find the code)
2. Principles: [`../philosophy/design-principles.md`](../philosophy/design-principles.md) (understand constraints)
3. Decisions: Specific ADRs for feature area
4. Rites: [`../rites/[relevant-rite]/`](../rites/) (practice documentation)

### Stakeholder / Manager

1. Summary: [`../DOCTRINE.md`](../DOCTRINE.md) (vision and principles)
2. Compliance: [`../compliance/COMPLIANCE-STATUS.md`](../compliance/COMPLIANCE-STATUS.md) (progress tracking)
3. ADRs: [`../../decisions/`](../../decisions/) (architectural decisions)

---

## Doctrine Maintenance

### Who Updates What

| Document | Updated By | Trigger |
|----------|-----------|---------|
| `knossos-doctrine.md` | Architects | Philosophical shifts, major revelations |
| `design-principles.md` | Architects | New principles emerge, implementation changes |
| `mythology-concordance.md` | Engineers + Architects | Source code moves, new components added |
| `COMPLIANCE-STATUS.md` | Automated + Manual review | Continuous (CI checks + periodic audits) |
| ADRs | Decision authors | Significant architectural decisions |
| Rite docs | Rite owners | Rite composition or manifest changes |

### Review Cycles

- **Doctrine (philosophy)**: Quarterly review, updated when foundational insights emerge
- **Design Principles**: Updated when implementations change or new patterns solidify
- **Mythology Concordance**: Updated with source code refactoring or component additions
- **Compliance Status**: Continuous automated checks, monthly manual review
- **ADRs**: Immutable once accepted (new ADRs supersede old ones)

---

## Related Resources

### External Documentation

- Worktree Guide: [`../guides/worktree-guide.md`](../guides/worktree-guide.md) — Parallel session patterns

### Skills

- `documentation` skill: Template and standards reference
- `standards` skill: Code conventions and tech stack
- `orchestration` skill: Consultation protocol routing

### Source Code

- Repository root (the Knossos SOURCE)
- CLI: `cmd/ari/`
- Internal packages: `internal/`
- Rites: `rites/`
- Cross-cutting agents: `agents/`
- Rite-specific agents: `rites/*/agents/`

---

## Contributing to Doctrine

### Adding New Doctrine Content

1. **Propose** via issue or discussion (for philosophical additions)
2. **Draft** in appropriate directory (`philosophy/`, `architecture/`, etc.)
3. **Review** with architects (doctrine changes require consensus)
4. **Update** this INDEX.md to include new content
5. **Cross-reference** from related documents

### Fixing Drift

If you notice implementation drift from doctrine:

1. **Document** in [`../compliance/COMPLIANCE-STATUS.md`](../compliance/COMPLIANCE-STATUS.md)
2. **Decide**: Update implementation to match doctrine OR update doctrine to match reality
3. **Create ADR** if philosophy changes
4. **Update** affected doctrine documents
5. **Track** remediation in compliance status

---

## Glossary

| Term | Definition | Reference |
|------|------------|-----------|
| **Doctrine** | Philosophical foundation of Knossos | `philosophy/knossos-doctrine.md` |
| **Principle** | Architectural revelation (one of eight) | `philosophy/design-principles.md` |
| **Myth** | Mythological element mapped to implementation | `philosophy/mythology-concordance.md` |
| **SOURCE** | Canonical code in Knossos repository | `philosophy/mythology-concordance.md` |
| **PROJECTION** | Materialized channel directories | `philosophy/mythology-concordance.md` |
| **ADR** | Architecture Decision Record | `../../decisions/` |
| **Rite** | Practice bundle (agents, skills, hooks, workflows) | `../rites/` |
| **Clew** | Session state + event log (the thread) | `philosophy/knossos-doctrine.md` Section VI |
| **White Sails** | Confidence signal (WHITE/GRAY/BLACK) | `philosophy/knossos-doctrine.md` Section VII |
| **Moirai** | Session lifecycle authority (Clotho, Lachesis, Atropos) | `philosophy/knossos-doctrine.md` Section II |
| **CC-OPP** | CC Operational Platform Properties — agent capability uplift | `reference/agent-capabilities.md` |
| **Frontmatter** | Agent YAML frontmatter declaring capabilities (memory, skills, hooks) | `reference/agent-capabilities.md` |
| **Mena** | Dromena (transient commands) + legomena (persistent knowledge) lifecycle | `philosophy/knossos-doctrine.md` Section XII |

---

**Welcome to the labyrinth. May the clew guide your return.**
