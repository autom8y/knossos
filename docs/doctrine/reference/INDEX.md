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
| See architectural decisions | [Architecture](#architecture) |
| Check compliance status | [Compliance](#compliance) |
| Track doctrine evolution | [Evolution](#evolution) |

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
- XIV. Implementation Drift Registry

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
- Supporting elements (Inscription, White Sails, Rites, Daedalus, Pythia)
- Destinations and states (Athens, Naxos, Dionysus, Minos, Minotaur)
- SOURCE vs PROJECTION distinction (critical!)
- Materialization flow

**For each mythological element:**
- Mythological origin (brief context)
- Knossos implementation (CORRECT source locations)
- Key files (actual paths in `/roster/`)
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
- [`ADR-0013-moirai-consolidation.md`](../../decisions/ADR-0013-moirai-consolidation.md) — Unified Moirai architecture
- [`ADR-0001-session-state-machine-redesign.md`](../../decisions/ADR-0001-session-state-machine-redesign.md) — Session lifecycle states
- [`ADR-sync-materialization.md`](../../decisions/ADR-sync-materialization.md) — Materialization system design

**Audience:** Architects, engineers implementing features, anyone needing decision context.

---

### Rite Catalog

**Path:** [`../rites/`](../rites/)

Documentation for each canonical rite (practice bundle).

**Available Rites:**
- 10x-dev (full development lifecycle)
- docs (documentation workflow)
- forge (agent and tool creation)
- hygiene (code quality maintenance)
- debt-triage (technical debt remediation)
- security (threat modeling and compliance)
- sre (operations and reliability)
- intelligence (research and synthesis)
- rnd (exploration and prototypes)
- strategy (business analysis)
- ecosystem (platform infrastructure)

**Audience:** Practitioners selecting or invoking rites, rite authors, workflow designers.

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

Operational foundations—how the platform runs, not why it exists.

**Topics:**
- Session management
- Hook architecture
- Materialization workflows
- CLI usage patterns

**Audience:** Operators, SREs, engineers maintaining platform infrastructure.

---

### Evolution

**Path:** [`../evolution/`](../evolution/)

Migration guides, deprecation timelines, evolution roadmaps.

**Topics:**
- Terminology migrations (thread → clew, state-mate → Moirai)
- Rename criteria (roster → knossos)
- Upgrade paths
- Breaking changes

**Audience:** Contributors updating code for new conventions, maintainers planning migrations.

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
3. Evolution: [`../evolution/`](../evolution/) (roadmap and migrations)

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

- User Preferences: [`../../guides/user-preferences.md`](../../guides/user-preferences.md)
- Knossos Integration Guide: [`../../guides/knossos-integration.md`](../../guides/knossos-integration.md)
- Knossos Migration Path: [`../../guides/knossos-migration.md`](../../guides/knossos-migration.md)

### Skills

- `documentation` skill: Template and standards reference
- `standards` skill: Code conventions and tech stack
- `orchestration` skill: Consultation protocol routing

### Source Code

- Repository: `/roster/` (the Knossos SOURCE)
- CLI: `/roster/cmd/ari/`
- Internal packages: `/roster/internal/`
- Rites: `/roster/rites/`
- Agents: `/roster/user-agents/`
- Skills: `/roster/user-skills/`

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
| **SOURCE** | Canonical code in `/roster/` | `philosophy/mythology-concordance.md` |
| **PROJECTION** | Materialized `.claude/` directories | `philosophy/mythology-concordance.md` |
| **ADR** | Architecture Decision Record | `../../decisions/` |
| **Rite** | Practice bundle (agents, skills, hooks, workflows) | `../rites/` |
| **Clew** | Session state + event log (the thread) | `philosophy/knossos-doctrine.md` Section VI |
| **White Sails** | Confidence signal (WHITE/GRAY/BLACK) | `philosophy/knossos-doctrine.md` Section VII |
| **Moirai** | Session lifecycle authority (Clotho, Lachesis, Atropos) | `philosophy/knossos-doctrine.md` Section II |

---

**Welcome to the labyrinth. May the clew guide your return.**
