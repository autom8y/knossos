---
name: naming-decision-record
description: "Naming Decision Record template for documenting mythological naming choices with provenance. Use when: proposing a new mythological name, evaluating naming alternatives, recording naming rationale for platform components. Triggers: naming decision, name proposal, mythology naming, naming record, NDR."
---

# Naming Decision Record

> Template for documenting mythological naming choices with full provenance.

## Purpose

When a new Greek-derived name is proposed for a platform concept, this template captures the decision with enough context to evaluate, challenge, or revisit it later. Every name in the platform carries architectural meaning -- this record ensures that meaning is explicit and traceable.

## Template

Copy the template below and fill in all six fields when proposing a new name.

```markdown
# NDR: {Name}

## Name
{The proposed name, in its canonical form as it will appear in code and documentation.}

## Tier
{One of the four provenance tiers from the mythology concordance:}
{- **Tier 1**: Bronze Age Attestation — attested on Linear B tablets or Bronze Age archaeological evidence from Knossos}
{- **Tier 2**: Classical Source — drawn from Homer, Hesiod, Herodotus, Pausanias, Plutarch, or other classical texts}
{- **Tier 3**: Hellenistic/Scholarly — drawn from post-classical scholarship, Hellenistic institutions, or Panhellenic practice}
{- **Tier 4**: Functional Analogy — chosen for functional resonance with the architectural role}

## Source
{Primary source citation. Format depends on tier:}
{- Tier 1: Linear B tablet ID (e.g., "KN Gg 702")}
{- Tier 2: Classical text reference (e.g., "Plutarch, Theseus 15-20")}
{- Tier 3: Scholarly reference (e.g., "Callimachus, Pinakes — Pfeiffer fr. 429-453")}
{- Tier 4: Functional derivation (e.g., "Greek dromenon (thing done) — Clement of Alexandria, Protrepticus")}

## Architectural Role
{What platform concept this name maps to. Be specific about the component, subsystem, or pattern.}
{Example: "Cross-session knowledge synthesizer that transforms raw session data into refined persistent knowledge."}

## Rationale
{Why this name over alternatives. Address:}
{- How the mythological meaning maps to the architectural function}
{- Whether the name occupies the correct mythological position (e.g., Pythia is consulted before the journey, not during)}
{- Any tension with existing names and how it is resolved}

## Alternatives Considered
{Other names evaluated and why they were rejected. At minimum one alternative.}
{Format: Name — reason rejected.}
{Example: "Hermes — rejected because Hermes implies message-passing between peers, not hierarchical routing."}
```

## Usage

- **When to create**: Before introducing any new Greek-derived name into the platform
- **Where to store**: As a file in `.sos/wip/` during proposal, moved to `.ledge/decisions/` upon acceptance
- **Who reviews**: The forge rite's prompt-architect or agent-curator, with final approval from the user
- **Cross-reference**: Check `docs/doctrine/philosophy/mythology-concordance.md` for existing names and tier definitions

## Field Validation

| Field | Required | Constraint |
|-------|----------|------------|
| Name | Yes | Must not conflict with existing concordance entries |
| Tier | Yes | Must be 1, 2, 3, or 4 |
| Source | Yes | Must cite a specific text, tablet, or derivation |
| Architectural Role | Yes | Must name a concrete platform concept |
| Rationale | Yes | Must address mythological-architectural mapping |
| Alternatives Considered | Yes | At least one alternative with rejection reason |
