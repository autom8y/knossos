---
name: feature-knowledge-criteria
description: "Criteria for per-feature knowledge capture via theoros observation. Use when: theoros is producing a feature knowledge reference for .know/feat/, documenting purpose, conceptual model, implementation, and boundaries of a specific feature. Triggers: feature knowledge criteria, per-feature documentation, feature observation, feature knowledge capture."
---

# Feature Knowledge Observation Criteria

> The theoros observes and documents a single feature -- producing a knowledge reference that enables any CC agent to understand the feature well enough to modify it safely. This is a template: the `/know` dispatch parameterizes it with a specific feature slug and census context.

## Scope

**Target sources** (theoros checks all applicable to the feature being documented):
- Go source in relevant packages (`internal/{feature-related}/`)
- ADRs in `docs/decisions/` that reference the feature
- Agent definitions in `rites/*/agents/*.md` that describe feature behaviors and responsibilities
- Dromena (commands) in `rites/*/mena/**/*.dro.md` that define user-facing feature surface
- Existing `.know/` files (`architecture.md`, `scar-tissue.md`, `conventions.md`) for structural context
- Rite manifests (`rites/*/manifest.yaml`) if the feature maps to a rite capability

**NOTE**: Scan rite SOURCE artifacts (`rites/`), NOT materialized outputs (`.claude/`). Knossos materializes `rites/` → `.claude/` via `ari sync`. The `.claude/` directory is a projection, not a source of truth.

**Observation focus**: Produce a comprehensive knowledge reference for a single feature. The reference must answer four questions: Why does this feature exist? How should agents think about it? Where is it implemented? What are its boundaries and failure modes?

**NOTE**: This domain uses knowledge-capture grading. Grade the COMPLETENESS of the knowledge reference, NOT compliance of the feature implementation. A = "an agent reading only this file could understand the feature well enough to modify it safely." F = "the document is incomplete -- an agent would make mistakes."

## Criteria

### Criterion 1: Purpose and Design Rationale (weight: 30%)

**What to observe**: Why this feature exists, what problem it solves, what design decisions shaped it, what alternatives were rejected, and what tradeoffs were accepted. The knowledge reference must give an agent the "why" before the "how."

**Evidence to collect**:
- ADRs in `docs/decisions/` that reference this feature (titles, decision records, rejected alternatives)
- Spike artifacts in `.ledge/spikes/` if they exist for this feature
- Commit history for initial feature introduction (the "why" commit messages)
- Any INTERVIEW_SYNTHESIS.md sections describing the feature's purpose
- Existing `.know/` entries that mention the feature

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Problem statement clearly articulated. Design decisions documented with rationale and evidence (ADR references, commit hashes). Rejected alternatives listed with rejection reasons. Tradeoffs explicitly identified with impact assessment. Cross-references to ADRs provided. |
| B | 80-89% completeness | Purpose clearly stated with problem context. Most design decisions documented. Rejected alternatives mentioned but not all with rationale. Tradeoffs identified but impact assessment incomplete. |
| C | 70-79% completeness | Purpose stated but problem context thin. Some design decisions documented. Alternatives and tradeoffs partially covered. ADR references incomplete. |
| D | 60-69% completeness | Purpose mentioned but vague. Design decisions listed without rationale. No alternatives documented. Tradeoffs not identified. |
| F | < 60% completeness | Purpose unclear or missing. Design rationale undocumented. An agent cannot understand why this feature exists. |

---

### Criterion 2: Conceptual Model (weight: 25%)

**What to observe**: How users and agents think about this feature. Key abstractions, terminology, mental model, state machines or lifecycles if applicable, and relationship to other features. The knowledge reference must give an agent the vocabulary and mental framework for reasoning about the feature.

**Evidence to collect**:
- Core terminology used in source code (type names, function prefixes, package names)
- State transitions or lifecycle stages if the feature has them (e.g., session states, materialization phases)
- Diagrams or descriptions of workflows the feature participates in
- Inter-feature dependencies (what this feature consumes from others, what it provides to others)
- User-facing concepts from commands and agent descriptions

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Clear mental model with all key terminology defined and contextualized. State machine or lifecycle documented if applicable (states, transitions, triggers). Inter-feature relationships mapped with direction (provides/consumes). An agent could reason about the feature using only this section. |
| B | 80-89% completeness | Mental model clear with most terminology defined. Lifecycle documented but minor gaps in transitions. Inter-feature relationships listed but direction not always specified. |
| C | 70-79% completeness | Some terminology defined but mental model incomplete. Lifecycle mentioned but not fully documented. Inter-feature relationships partially listed. |
| D | 60-69% completeness | Terminology used without definition. No lifecycle documentation. Relationships mentioned in passing without structure. |
| F | < 60% completeness | No conceptual framework. Terminology undefined. An agent would have to reverse-engineer the mental model from source code. |

---

### Criterion 3: Implementation Map (weight: 25%)

**What to observe**: Which packages and files implement this feature, key types and entry points, data flow through the feature, public API surface, and test coverage. The knowledge reference must tell an agent exactly where to look and what to expect.

**Evidence to collect**:
- List all packages under `internal/` (and `cmd/` if applicable) that implement this feature
- For each package: primary purpose relative to the feature, key exported types, entry point functions
- Data flow: input sources -> processing stages -> output destinations for the feature's primary path
- Public API surface: exported functions/types that other packages depend on
- Test file locations and what aspects of the feature they cover

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Every implementing package documented with purpose, key types, and entry points. Data flow traced from input to output with processing stages named. Public API surface listed with consuming packages. Test locations identified with coverage description. |
| B | 80-89% completeness | All packages listed with purpose. Most have key types documented. Data flow documented but minor gaps in processing stages. Test locations listed. |
| C | 70-79% completeness | Key packages documented but some omitted. Data flow partially traced. API surface incomplete. Test coverage mentioned but locations not specific. |
| D | 60-69% completeness | Some packages listed without purpose detail. Data flow described vaguely. API surface not documented. Test coverage unknown. |
| F | < 60% completeness | Implementation locations unknown or incomplete. An agent cannot find where to make changes for this feature. |

---

### Criterion 4: Boundaries and Failure Modes (weight: 20%)

**What to observe**: What this feature does NOT do (explicit scope boundaries), known edge cases and limitations, error paths and recovery mechanisms, and interaction points with other features where boundaries blur. The knowledge reference must protect an agent from making changes that violate implicit assumptions.

**Evidence to collect**:
- Explicit scope limitations documented in source comments, ADRs, or commit messages
- Known edge cases from scar tissue (`.know/scar-tissue.md` entries referencing this feature)
- Error return paths in key functions (what errors are returned, how callers handle them)
- Panic or fatal conditions if any
- Interaction points: where this feature's code calls into other features' packages or vice versa
- Configuration boundaries: what settings affect this feature, what values are invalid

| Grade | Threshold | Evidence Required |
|-------|-----------|-------------------|
| A | 90-100% completeness | Clear scope boundaries documented with "this feature does NOT" statements. Failure modes cataloged with evidence (scar references, error return analysis). Error paths traced with recovery mechanisms. Interaction points identified with boundary clarity assessment. Configuration boundaries specified. |
| B | 80-89% completeness | Scope boundaries documented. Most failure modes cataloged. Error paths listed but recovery not fully traced. Interaction points identified. Minor gaps in configuration boundary documentation. |
| C | 70-79% completeness | Some scope boundaries stated. Failure modes partially cataloged. Error paths mentioned but not traced. Interaction points partially identified. |
| D | 60-69% completeness | Boundaries vague. Failure modes listed without evidence. Error paths not analyzed. Interaction points unknown. |
| F | < 60% completeness | Boundaries undefined. Failure modes undocumented. An agent modifying this feature would not know what to avoid. |

---

## Grading Calculation

Final grade is weighted average of all criteria midpoint scores (see `schemas/grading.lego.md`). Example:
- Purpose and Design Rationale: A (midpoint 95%) x 30% = 28.5
- Conceptual Model: B (midpoint 85%) x 25% = 21.25
- Implementation Map: B (midpoint 85%) x 25% = 21.25
- Boundaries and Failure Modes: A (midpoint 95%) x 20% = 19.0
- **Total: 90.0 -> A**

## Related

- [Pinakes INDEX](../INDEX.lego.md) -- Full audit system documentation
- [feature-census-criteria](feature-census.lego.md) -- Feature enumeration census (upstream producer of feature slugs)
- [architecture-criteria](architecture.lego.md) -- Codebase architecture knowledge capture (complementary domain)
- [scar-tissue-criteria](scar-tissue.lego.md) -- Scar tissue knowledge capture (failure mode evidence source)
- [design-constraints-criteria](design-constraints.lego.md) -- Design constraint capture (boundary and tension evidence source)
